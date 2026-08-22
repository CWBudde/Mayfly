package mayfly

import (
	"math"
	"math/rand"
)

// DefaultCrossoverGamma is the blend-crossover expansion factor used by the
// reference Mayfly implementation. The author's own Python port draws the
// crossover coefficient from a uniform distribution over [-gamma, 1+gamma]
// with gamma = 0.4 (KZervoudakis/Mayfly-Optimization-Algorithm-Python,
// operators.py: ContinousCrossover; ma.py: MA(..., gamma=0.4)).
//
// A gamma of zero confines the coefficient to [0, 1], which makes every
// offspring a convex combination of its parents. That is a contraction: the
// convex hull of the population can only shrink from one generation to the
// next, and the swarm loses spread it can never regain by mating. gamma > 0
// lets an offspring land outside the parental interval, which is what keeps
// the blend operator from collapsing diversity.
const DefaultCrossoverGamma = 0.4

// Crossover performs crossover between two parent positions using
// DefaultCrossoverGamma. See CrossoverBlend for the general form.
func Crossover(x1, x2 []float64, lowerBound, upperBound float64, rng *rand.Rand) ([]float64, []float64) {
	return CrossoverBlend(x1, x2, DefaultCrossoverGamma, lowerBound, upperBound, rng)
}

// CrossoverBlend performs blend (BLX-style) crossover between two parent
// positions. The per-dimension coefficient is drawn from U(-gamma, 1+gamma),
// so offspring may fall outside the interval spanned by the two parents by up
// to gamma times its width on either side. Offspring are clamped to
// [lowerBound, upperBound] afterwards.
//
// A negative gamma is treated as zero.
func CrossoverBlend(
	x1, x2 []float64,
	gamma, lowerBound, upperBound float64,
	rng *rand.Rand,
) ([]float64, []float64) {
	if math.IsNaN(gamma) || gamma < 0 {
		gamma = 0
	}

	size := len(x1)
	off1 := make([]float64, size)
	off2 := make([]float64, size)

	for i := range size {
		L := unifrnd(-gamma, 1+gamma, rng)
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
