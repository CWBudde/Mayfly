package mayfly

import (
	"fmt"
	"math"
	"math/rand"
)

// unifrnd generates a random float64 between min and max.
func unifrnd(min, max float64, rng *rand.Rand) float64 {
	if rng == nil {
		return min + rand.Float64()*(max-min)
	}

	return min + rng.Float64()*(max-min)
}

// unifrndVec generates a vector of random float64 values between min and max.
func unifrndVec(min, max float64, size int, rng *rand.Rand) []float64 {
	vec := make([]float64, size)
	for i := range vec {
		vec[i] = unifrnd(min, max, rng)
	}

	return vec
}

// randn generates a normally distributed random number.
func randn(rng *rand.Rand) float64 {
	if rng == nil {
		return rand.NormFloat64()
	}

	return rng.NormFloat64()
}

// maxVec returns element-wise maximum of vector and scalar.
func maxVec(vec []float64, bound float64) {
	for i := range vec {
		if vec[i] < bound {
			vec[i] = bound
		}
	}
}

// minVec returns element-wise minimum of vector and scalar.
func minVec(vec []float64, bound float64) {
	for i := range vec {
		if vec[i] > bound {
			vec[i] = bound
		}
	}
}

// sortMayflies sorts mayflies by cost (ascending).
func sortMayflies(mayflies []*Mayfly) {
	// Simple bubble sort for small populations
	n := len(mayflies)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if mayflies[j].Cost > mayflies[j+1].Cost {
				mayflies[j], mayflies[j+1] = mayflies[j+1], mayflies[j]
			}
		}
	}
}

// effectiveNM reports the mutant count Optimize will actually use, resolving
// the "0 means 5% of NPop" default the same way the main loop does.
func effectiveNM(config *Config) int {
	if config.NM != 0 {
		return config.NM
	}

	return int(math.Round(0.05 * float64(config.NPop)))
}

// validateOffspring checks NC and NM against the population sizes.
//
// NC drives three separate index expressions in the main loop, none of which
// bounds-check: the mating loop reads males[k] and females[k] for k < NC/2, and
// the mutation step draws a uniform parent from the offspring slice. A caller
// who shrinks the population without also shrinking NC — the default NC of 20
// with any NPop below 10, for instance — used to get an out-of-range panic from
// inside the library rather than an error out of Optimize.
func validateOffspring(config *Config) error {
	if config.NC < 0 {
		return fmt.Errorf("NC (offspring count) must be non-negative, got %d", config.NC)
	}

	if config.NM < 0 {
		return fmt.Errorf("NM (mutant count) must be non-negative, got %d", config.NM)
	}

	// Mating pairs the k-th best male with the k-th best female, so neither
	// population may be shorter than the number of pairs.
	if pairs := config.NC / 2; pairs > config.NPop || pairs > config.NPopF {
		return fmt.Errorf(
			"NC (offspring count) of %d needs %d parent pairs, "+
				"which exceeds NPop=%d or NPopF=%d; lower NC or raise the populations",
			config.NC, pairs, config.NPop, config.NPopF,
		)
	}

	// Mutants are drawn from the offspring, so there must be at least one.
	if config.NC < 2 && effectiveNM(config) > 0 {
		return fmt.Errorf(
			"NC (offspring count) of %d produces no offspring for %d mutants to be drawn from; "+
				"raise NC to at least 2 or set NM to 0",
			config.NC, effectiveNM(config),
		)
	}

	return nil
}
