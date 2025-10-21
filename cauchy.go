// Package mayfly - Cauchy Distribution Implementation
//
// Implements Cauchy distribution for heavy-tailed mutation in GSASMA variant.
//
// The Cauchy distribution has heavier tails than Gaussian, providing better
// exploration capability while being easier to sample than Lévy flights.
//
// Reference:
// Standard inverse CDF method for Cauchy distribution.
// Used in evolutionary computation for exploration (see GSASMA variant).
package mayfly

import (
	"math"
	"math/rand"
)

// cauchyRand generates a Cauchy-distributed random number.
// If U ~ Uniform(0,1), then X = x0 + gamma * tan(π*(U - 0.5)) ~ Cauchy(x0, gamma).
// rng must not be nil (ensured by caller).
func cauchyRand(x0, gamma float64, rng *rand.Rand) float64 {
	// Generate uniform random number in (0, 1)
	// Avoid exact 0 and 1 to prevent tan() overflow
	u := rng.Float64()
	for u == 0.0 || u == 1.0 {
		u = rng.Float64()
	}

	// Apply inverse CDF: F^(-1)(u) = x0 + gamma * tan(π*(u - 0.5))
	result := x0 + gamma*math.Tan(math.Pi*(u-0.5))

	// Sanitize extreme values from tan() function
	// Cauchy can produce very large values; cap at reasonable limits
	if math.IsNaN(result) || math.IsInf(result, 0) {
		// Retry with different random value
		u = rng.Float64()
		result = x0 + gamma*math.Tan(math.Pi*(u-0.5))
		// If still invalid, return center point
		if math.IsNaN(result) || math.IsInf(result, 0) {
			return x0
		}
	}

	return result
}

// cauchyRandVec generates a vector of Cauchy-distributed random numbers.
// Each element is independently sampled from Cauchy(x0, gamma).
func cauchyRandVec(size int, x0, gamma float64, rng *rand.Rand) []float64 {
	vec := make([]float64, size)
	for i := 0; i < size; i++ {
		vec[i] = cauchyRand(x0, gamma, rng)
	}

	return vec
}

// MutateCauchy applies Cauchy mutation to a solution.
// Used in GSASMA for heavy-tailed exploration.
// rng must not be nil (ensured by caller).
// Returns: mutated position vector.
func MutateCauchy(x []float64, mu, lowerBound, upperBound float64, rng *rand.Rand) []float64 {
	nVar := len(x)
	nMu := int(math.Ceil(mu * float64(nVar)))

	// Scale parameter: Use 10% of search space as in Gaussian mutation
	// This provides comparable exploration range while leveraging heavy tails
	gamma := 0.1 * (upperBound - lowerBound)

	y := make([]float64, nVar)
	copy(y, x)

	// Select random indices to mutate
	indices := rng.Perm(nVar)[:nMu]

	for _, j := range indices {
		// Apply Cauchy perturbation centered at current position
		perturbation := cauchyRand(0, gamma, rng)
		y[j] = x[j] + perturbation

		// Cauchy can generate very large values; clip extreme outliers
		// to prevent numerical issues while preserving exploration capability
		searchSpan := upperBound - lowerBound
		if math.Abs(y[j]-x[j]) > 3*searchSpan {
			// If perturbation is > 3x search space, clip it
			if perturbation > 0 {
				y[j] = x[j] + 3*searchSpan
			} else {
				y[j] = x[j] - 3*searchSpan
			}
		}
	}

	// Apply position limits
	maxVec(y, lowerBound)
	minVec(y, upperBound)

	return y
}

// HybridMutate applies either Cauchy or Gaussian mutation based on probability.
// Used in GSASMA to balance exploration (Cauchy) and exploitation (Gaussian).
// rng must not be nil (ensured by caller).
// Returns: mutated position vector.
func HybridMutate(x []float64, mu, lowerBound, upperBound, cauchyProb float64, rng *rand.Rand) []float64 {
	// Decide which mutation type to use
	if rng.Float64() < cauchyProb {
		return MutateCauchy(x, mu, lowerBound, upperBound, rng)
	}

	return MutateGaussian(x, mu, lowerBound, upperBound, rng)
}
