package mayfly

import (
	"math"
	"math/rand"
)

// If U ~ Uniform(0,1), then X = x0 + gamma * tan(π*(U - 0.5)) ~ Cauchy(x0, gamma).
func cauchyRand(x0, gamma float64, rng *rand.Rand) float64 {
	if rng == nil {
		rng = rand.New(rand.NewSource(0))
	}

	// Generate uniform random number in (0, 1)
	// Avoid exact 0 and 1 to prevent tan() overflow
	u := rng.Float64()
	for u == 0.0 || u == 1.0 {
		u = rng.Float64()
	}

	// Apply inverse CDF: F^(-1)(u) = x0 + gamma * tan(π*(u - 0.5))
	return x0 + gamma*math.Tan(math.Pi*(u-0.5))
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
	var indices []int
	if rng != nil {
		indices = rng.Perm(nVar)[:nMu]
	} else {
		indices = rand.Perm(nVar)[:nMu]
	}

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

// Returns: mutated position vector.
func HybridMutate(x []float64, mu, lowerBound, upperBound, cauchyProb float64, rng *rand.Rand) []float64 {
	if rng == nil {
		rng = rand.New(rand.NewSource(0))
	}

	// Decide which mutation type to use
	if rng.Float64() < cauchyProb {
		return MutateCauchy(x, mu, lowerBound, upperBound, rng)
	}

	return MutateGaussian(x, mu, lowerBound, upperBound, rng)
}
