//go:build js && wasm

package main

import (
	"fmt"
	"math/rand"

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
// Fresh is load-bearing, not defensive: Optimize writes back into the config it
// is given — it installs a Rand when the field is nil, and fills in NM, VelMax
// and VelMin when they are zero. A cached config would therefore carry the
// previous run's advanced RNG state into the next one, and "run it again with
// the same seed" would quietly stop reproducing.
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
	config.Rand = rand.New(rand.NewSource(seed)) //nolint:gosec // Reproducibility is the point; this is not a security context.

	// Goroutines under js/wasm are cooperatively scheduled onto the single
	// browser thread, so the worker pool buys nothing and costs coordination.
	config.EnableParallel = false

	return config, nil
}
