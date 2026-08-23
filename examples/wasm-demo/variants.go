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

// configFor builds a fresh configuration for one run.
//
// A fresh config keeps each UI run independent. Optimize treats it as immutable,
// and Config.Seed makes the reproducibility metadata truthful.
func configFor(variantName, benchmarkName string, dimensions, iterations, npop, npopf int, seed int64,
	lower, upper float64,
) (*mayfly.Config, error) {
	variant := mayfly.NewVariant(variantName)
	if variant == nil {
		return nil, fmt.Errorf("unknown variant %q", variantName)
	}

	spec, ok := lookupBenchmark(benchmarkName)
	if !ok {
		return nil, fmt.Errorf("unknown benchmark %q", benchmarkName)
	}

	config := variant.GetConfig()
	config.ObjectiveFunc = spec.fn
	config.ProblemSize = dimensions
	config.LowerBound = lower
	config.UpperBound = upper
	config.MaxIterations = iterations
	config.NPop = npop
	config.NPopF = npopf
	config.Seed = &seed

	// Goroutines under js/wasm are cooperatively scheduled onto the single
	// browser thread, so the worker pool buys nothing and costs coordination.
	config.EnableParallel = false

	return config, nil
}
