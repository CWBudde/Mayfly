package mayfly

import (
	"errors"
	"fmt"
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
//
// Deprecated: use CrossoverChecked to receive validation errors. This legacy
// wrapper retains global-RNG and silent fallback behavior for source
// compatibility.
func Crossover(x1, x2 []float64, lowerBound, upperBound float64, rng *rand.Rand) ([]float64, []float64) {
	return CrossoverBlend(x1, x2, DefaultCrossoverGamma, lowerBound, upperBound, rng)
}

// CrossoverChecked validates both parents, bounds, and RNG before applying
// the default blend crossover.
func CrossoverChecked(
	x1, x2 []float64,
	lowerBound, upperBound float64,
	rng *rand.Rand,
) ([]float64, []float64, error) {
	return CrossoverBlendChecked(x1, x2, DefaultCrossoverGamma, lowerBound, upperBound, rng)
}

// CrossoverBlend performs blend (BLX-style) crossover between two parent
// positions. The per-dimension coefficient is drawn from U(-gamma, 1+gamma),
// so offspring may fall outside the interval spanned by the two parents by up
// to gamma times its width on either side. Offspring are clamped to
// [lowerBound, upperBound] afterwards.
//
// A negative or non-finite gamma is treated as zero; drawing the coefficient
// with an infinite or NaN gamma would otherwise produce NaN offspring, which
// the boundary clamps cannot repair.
//
// Deprecated: use CrossoverBlendChecked. This legacy wrapper preserves the
// pre-v0.7 coercion of an invalid gamma to zero.
func CrossoverBlend(
	x1, x2 []float64,
	gamma, lowerBound, upperBound float64,
	rng *rand.Rand,
) ([]float64, []float64) {
	if math.IsNaN(gamma) || math.IsInf(gamma, 0) || gamma < 0 {
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

// CrossoverBlendChecked performs blend crossover after validating every
// value. It never uses the package-global random generator.
func CrossoverBlendChecked(
	x1, x2 []float64,
	gamma, lowerBound, upperBound float64,
	rng *rand.Rand,
) ([]float64, []float64, error) {
	err := validateOperatorInput(x1, lowerBound, upperBound, rng)
	if err != nil {
		return nil, nil, fmt.Errorf("first parent: %w", err)
	}

	err = validateOperatorInput(x2, lowerBound, upperBound, rng)
	if err != nil {
		return nil, nil, fmt.Errorf("second parent: %w", err)
	}

	if len(x1) != len(x2) {
		return nil, nil, fmt.Errorf("parent dimensions differ: %d and %d", len(x1), len(x2))
	}

	if !isFinite(gamma) || gamma < 0 {
		return nil, nil, fmt.Errorf("crossover gamma must be finite and non-negative, got %v", gamma)
	}

	off1, off2 := CrossoverBlend(x1, x2, gamma, lowerBound, upperBound, rng)

	return off1, off2, nil
}

// mutationCount converts a mutation rate into the number of dimensions to
// perturb, saturating instead of producing a slice bound the operators cannot
// use.
//
// Optimize rejects a mu outside [0,1] before a run starts, so this only has to
// hold for callers who reach the exported operators directly. Their previous
// answer to an out-of-range rate was a slice-bounds panic: mu > 1 asks for more
// dimensions than exist, mu < 0 asks for a negative count, and NaN converts to
// the most negative int. None of those is a more useful outcome than mutating
// every dimension or none.
func mutationCount(mu float64, nVar int) int {
	if math.IsNaN(mu) || mu <= 0 {
		return 0
	}

	if mu >= 1 {
		return nVar
	}

	return min(int(math.Ceil(mu*float64(nVar))), nVar)
}

// MutateGaussian applies Gaussian mutation to a position vector.
// This uses a normal (Gaussian) distribution for perturbations.
//
// Deprecated: use MutateGaussianChecked to reject invalid rates, bounds, and
// RNG state rather than saturating or falling back to the global RNG.
func MutateGaussian(x []float64, mu, lowerBound, upperBound float64, rng *rand.Rand) []float64 {
	nVar := len(x)
	nMu := mutationCount(mu, nVar)
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

// MutateGaussianChecked validates its input before applying Gaussian mutation.
func MutateGaussianChecked(
	x []float64,
	mu, lowerBound, upperBound float64,
	rng *rand.Rand,
) ([]float64, error) {
	err := validateMutationInput(x, mu, lowerBound, upperBound, rng)
	if err != nil {
		return nil, err
	}

	return MutateGaussian(x, mu, lowerBound, upperBound, rng), nil
}

// Mutate applies mutation to a position vector using Gaussian distribution.
// This is an alias for MutateGaussian for backward compatibility.
// Deprecated: use MutateGaussianChecked.
func Mutate(x []float64, mu, lowerBound, upperBound float64, rng *rand.Rand) []float64 {
	return MutateGaussian(x, mu, lowerBound, upperBound, rng)
}

func validateMutationInput(
	x []float64,
	mu, lowerBound, upperBound float64,
	rng *rand.Rand,
) error {
	if !isFinite(mu) || mu < 0 || mu > 1 {
		return fmt.Errorf("mutation rate must be in [0,1], got %v", mu)
	}

	return validateOperatorInput(x, lowerBound, upperBound, rng)
}

func validateOperatorInput(x []float64, lowerBound, upperBound float64, rng *rand.Rand) error {
	if len(x) == 0 {
		return errors.New("position vector is empty")
	}

	if rng == nil {
		return errors.New("random generator is nil")
	}

	if !isFinite(lowerBound) || !isFinite(upperBound) || lowerBound >= upperBound {
		return fmt.Errorf("bounds must be finite and increasing, got [%v,%v]", lowerBound, upperBound)
	}

	for i, coordinate := range x {
		if !isFinite(coordinate) {
			return fmt.Errorf("position %d is not finite", i)
		}

		if coordinate < lowerBound || coordinate > upperBound {
			return fmt.Errorf("position %d=%v is outside [%v,%v]", i, coordinate, lowerBound, upperBound)
		}
	}

	return nil
}
