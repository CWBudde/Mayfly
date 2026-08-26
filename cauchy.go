// Cauchy Distribution Implementation.
//
// Implements Cauchy distribution sampling for heavy-tailed mutation.
//
// The Cauchy distribution has heavier tails than Gaussian, providing better
// exploration capability while being easier to sample than Lévy flights.
//
// Reference:
// Standard inverse CDF method for Cauchy distribution.
// Used by HMMA's global-best mutation and the exported HybridMutate helper.

package mayfly

import (
	"fmt"
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
//
//nolint:unused // reserved for vectorised Cauchy mutation; not wired into Optimize() yet.
func cauchyRandVec(size int, x0, gamma float64, rng *rand.Rand) []float64 {
	vec := make([]float64, size)
	for i := range size {
		vec[i] = cauchyRand(x0, gamma, rng)
	}

	return vec
}

// MutateCauchy applies Cauchy mutation to a solution.
// It is an exported generic helper and is not part of the GSASMA lifecycle.
// rng must not be nil (ensured by caller).
// Returns: mutated position vector.
//
// Deprecated: use MutateCauchyChecked. This legacy wrapper assumes a non-nil
// RNG and silently saturates an invalid mutation rate.
func MutateCauchy(x []float64, mu, lowerBound, upperBound float64, rng *rand.Rand) []float64 {
	nVar := len(x)
	nMu := mutationCount(mu, nVar)

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

// MutateCauchyChecked validates its input before applying Cauchy mutation.
func MutateCauchyChecked(
	x []float64,
	mu, lowerBound, upperBound float64,
	rng *rand.Rand,
) ([]float64, error) {
	err := validateMutationInput(x, mu, lowerBound, upperBound, rng)
	if err != nil {
		return nil, err
	}

	return MutateCauchy(x, mu, lowerBound, upperBound, rng), nil
}

// HybridMutate applies either Cauchy or Gaussian mutation based on probability.
// It is an exported generic helper and is not part of the GSASMA lifecycle.
// rng must not be nil (ensured by caller).
// Returns: mutated position vector.
//
// Deprecated: use HybridMutateChecked. This legacy wrapper assumes a non-nil
// RNG and accepts invalid probabilities.
func HybridMutate(x []float64, mu, lowerBound, upperBound, cauchyProb float64, rng *rand.Rand) []float64 {
	// Decide which mutation type to use
	if rng.Float64() < cauchyProb {
		return MutateCauchy(x, mu, lowerBound, upperBound, rng)
	}

	return MutateGaussian(x, mu, lowerBound, upperBound, rng)
}

// HybridMutateChecked validates mutation and branch probabilities before
// applying hybrid mutation.
func HybridMutateChecked(
	x []float64,
	mu, lowerBound, upperBound, cauchyProb float64,
	rng *rand.Rand,
) ([]float64, error) {
	err := validateMutationInput(x, mu, lowerBound, upperBound, rng)
	if err != nil {
		return nil, err
	}

	if !isFinite(cauchyProb) || cauchyProb < 0 || cauchyProb > 1 {
		return nil, fmt.Errorf("Cauchy probability must be in [0,1], got %v", cauchyProb)
	}

	return HybridMutate(x, mu, lowerBound, upperBound, cauchyProb, rng), nil
}
