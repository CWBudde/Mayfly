package mayfly

import (
	"math"
	"math/rand"
)

// Crossover performs crossover between two parent positions.
func Crossover(x1, x2 []float64, lowerBound, upperBound float64, rng *rand.Rand) ([]float64, []float64) {
	size := len(x1)
	off1 := make([]float64, size)
	off2 := make([]float64, size)

	for i := 0; i < size; i++ {
		L := unifrnd(0, 1, rng)
		off1[i] = L*x1[i] + (1-L)*x2[i]
		off2[i] = L*x2[i] + (1-L)*x1[i]
	}

	// Apply position limits
	maxVec(off1, lowerBound)
	minVec(off1, upperBound)
	maxVec(off2, lowerBound)
	minVec(off2, upperBound)

	return off1, off2
}

// MutateGaussian applies Gaussian mutation to a position vector.
// This uses a normal (Gaussian) distribution for perturbations.
func MutateGaussian(x []float64, mu, lowerBound, upperBound float64, rng *rand.Rand) []float64 {
	nVar := len(x)
	nMu := int(math.Ceil(mu * float64(nVar)))
	sigma := 0.1 * (upperBound - lowerBound)

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
		y[j] = x[j] + sigma*randn(rng)
	}

	// Apply position limits
	maxVec(y, lowerBound)
	minVec(y, upperBound)

	return y
}

// Mutate applies mutation to a position vector using Gaussian distribution.
// This is an alias for MutateGaussian for backward compatibility.
func Mutate(x []float64, mu, lowerBound, upperBound float64, rng *rand.Rand) []float64 {
	return MutateGaussian(x, mu, lowerBound, upperBound, rng)
}
