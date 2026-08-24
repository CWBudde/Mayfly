//go:build js && wasm

package main

import (
	"fmt"

	"github.com/cwbudde/mayfly"
)

// variantOrder is the library's own deterministic ordering (GetAllVariants),
// pinned here as the demo's canonical order. ListVariants ranges over a map and
// so returns its names shuffled, which would reorder the UI on every load.
func variantOrder() []mayfly.AlgorithmVariant {
	return mayfly.GetAllVariants()
}

// variantKey is the string the page sends back. NewVariant lowercases and
// trims, so Name() round-trips through it.
func variantKey(variant mayfly.AlgorithmVariant) string {
	return variant.Name()
}

// qmcInitOption describes one initial-population strategy for the UI. The keys
// are the library's own Config.QMCInit values, so the page never invents a
// string the library would then reject.
type qmcInitOption struct {
	key   string
	label string
	note  string
}

// qmcInitOptions is the demo's reading of Config.QMCInit. The library validates
// the value it is given; what it cannot supply is a one-line explanation per
// strategy, which is the only reason this table exists here.
func qmcInitOptions() []qmcInitOption {
	return []qmcInitOption{
		{
			key:   mayfly.QMCInitUniform,
			label: "uniform",
			note:  "Every coordinate an independent random draw — the historical behaviour, and the one that leaves gaps and clusters.",
		},
		{
			key:   mayfly.QMCInitSobol,
			label: "Sobol",
			note:  "A scrambled Sobol sequence fills the box more evenly for the same number of evaluations. Watch iteration 0 on the heatmap.",
		},
		{
			key:   mayfly.QMCInitHalton,
			label: "Halton",
			note:  "A scrambled Halton sequence, with the usual burn-in. No dimension ceiling, slightly worse coverage than Sobol in high dimensions.",
		},
	}
}

// resolveQMCInit maps a value from the page onto a strategy the library knows,
// treating the empty string as the default rather than an error. Optimize
// validates too; catching it here names the control the user actually touched.
func resolveQMCInit(name string) (string, error) {
	if name == "" {
		return mayfly.QMCInitUniform, nil
	}

	for _, option := range qmcInitOptions() {
		if option.key == name {
			return option.key, nil
		}
	}

	return "", fmt.Errorf("unknown initial-population strategy %q", name)
}

// runSettings is one page request, already clamped and validated.
type runSettings struct {
	variant    string
	benchmark  string
	qmcInit    string
	dimensions int
	iterations int
	npop       int
	npopf      int
	seed       int64
	lower      float64
	upper      float64
}

// configFor builds a fresh configuration for one run.
//
// A fresh config keeps each UI run independent. Optimize treats it as immutable,
// and Config.Seed makes the reproducibility metadata truthful.
func configFor(settings runSettings) (*mayfly.Config, error) {
	variant := mayfly.NewVariant(settings.variant)
	if variant == nil {
		return nil, fmt.Errorf("unknown variant %q", settings.variant)
	}

	spec, ok := lookupBenchmark(settings.benchmark)
	if !ok {
		return nil, fmt.Errorf("unknown benchmark %q", settings.benchmark)
	}

	qmcInit, err := resolveQMCInit(settings.qmcInit)
	if err != nil {
		return nil, err
	}

	seed := settings.seed

	config := variant.GetConfig()
	config.ObjectiveFunc = spec.fn
	config.ProblemSize = settings.dimensions
	config.LowerBound = settings.lower
	config.UpperBound = settings.upper
	config.MaxIterations = settings.iterations
	config.NPop = settings.npop
	config.NPopF = settings.npopf
	config.Seed = &seed
	config.QMCInit = qmcInit

	// Goroutines under js/wasm are cooperatively scheduled onto the single
	// browser thread, so the worker pool buys nothing and costs coordination.
	config.EnableParallel = false

	return config, nil
}

// qmcVariant is an AlgorithmVariant that hands out its delegate's configuration
// with one field changed.
//
// ComparisonRunner builds every job's config from variant.GetConfig() and
// exposes no hook to amend it, so overriding the accessor is the only way to
// put the sweep's initial-population strategy in front of the framework without
// changing the library. Everything else — name, description, overhead — is the
// delegate's, which is what keeps the results table labelled correctly.
type qmcVariant struct {
	mayfly.AlgorithmVariant

	qmcInit string
}

func (v qmcVariant) GetConfig() *mayfly.Config {
	config := v.AlgorithmVariant.GetConfig()
	if config != nil {
		config.QMCInit = v.qmcInit
	}

	return config
}

// selectedVariants resolves the page's variant keys, defaulting to all of them,
// and applies one initial-population strategy to the whole set. A sweep that
// mixed strategies across columns would not be comparing variants any more.
func selectedVariants(names []string, qmcInit string) ([]mayfly.AlgorithmVariant, error) {
	variants := variantOrder()

	if len(names) > 0 {
		variants = make([]mayfly.AlgorithmVariant, 0, len(names))

		for _, name := range names {
			variant := mayfly.NewVariant(name)
			if variant == nil {
				return nil, fmt.Errorf("unknown variant %q", name)
			}

			variants = append(variants, variant)
		}
	}

	strategy, err := resolveQMCInit(qmcInit)
	if err != nil {
		return nil, err
	}

	wrapped := make([]mayfly.AlgorithmVariant, len(variants))
	for i, variant := range variants {
		wrapped[i] = qmcVariant{AlgorithmVariant: variant, qmcInit: strategy}
	}

	return wrapped, nil
}
