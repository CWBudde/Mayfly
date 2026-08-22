// Golden Sine Algorithm Implementation.
//
// Implements the Golden Sine Algorithm for convergence acceleration in GSASMA.
//
// Reference:
// Tanyildizi, E., & Demir, G. (2017). Golden Sine Algorithm: A Novel
// Math-Inspired Algorithm. Advances in Electrical and Computer Engineering,
// 17(2), 71-78.
// DOI: 10.4316/AECE.2017.02010
//
// The algorithm scans the unit circle with the sine function while a golden
// section search narrows the interval [-π, π] around the incumbent, so the
// golden ratio controls how the search interval shrinks rather than merely
// scaling a step size.

package mayfly

import (
	"math"
	"math/rand"
)

const (
	// GoldenRatio is the mathematical constant φ = (1 + √5) / 2 ≈ 1.618034.
	GoldenRatio = 1.618033988749895

	// goldenRatioConjugate is τ = 1/φ = (√5 - 1) / 2 ≈ 0.618034, the section
	// ratio used to split the search interval.
	goldenRatioConjugate = 1.0 / GoldenRatio
)

// goldenSection holds the golden section search interval of the Golden Sine
// Algorithm. x1 and x2 are the two interior section points that weight the
// incumbent and the current position in the update rule; the interval shrinks
// by the golden ratio after every candidate and resets to [-π, π] once the two
// section points coincide.
type goldenSection struct {
	a, b   float64
	x1, x2 float64
}

// newGoldenSection returns a golden section search over the full [-π, π]
// interval prescribed by the Golden Sine Algorithm.
func newGoldenSection() *goldenSection {
	section := &goldenSection{a: -math.Pi, b: math.Pi}
	section.recompute()

	return section
}

// recompute derives the interior section points from the current interval.
func (s *goldenSection) recompute() {
	s.x1 = s.a*goldenRatioConjugate + s.b*(1-goldenRatioConjugate)
	s.x2 = s.a*(1-goldenRatioConjugate) + s.b*goldenRatioConjugate
}

// snapshot returns a copy of the section for use by concurrent candidate
// generation, so that a batch of candidates shares one set of section points.
func (s *goldenSection) snapshot() goldenSection {
	return *s
}

// update narrows the search interval after a candidate has been judged.
// improved reports whether the candidate beat the position it was generated
// from. The interval resets once it has collapsed.
func (s *goldenSection) update(improved bool) {
	if improved {
		s.b = s.x2
		s.x2 = s.x1
		s.x1 = s.a*goldenRatioConjugate + s.b*(1-goldenRatioConjugate)
	} else {
		s.a = s.x1
		s.x1 = s.x2
		s.x2 = s.a*(1-goldenRatioConjugate) + s.b*goldenRatioConjugate
	}

	if s.x1 == s.x2 {
		s.a, s.b = -math.Pi, math.Pi
		s.recompute()
	}
}

// goldenSineUpdate applies the Golden Sine Algorithm update rule:
//
//	X(t+1) = X(t)·|sin(r1)| - r2·sin(r1)·|x1·P - x2·X(t)|
//
// with r1 ∈ [0, 2π], r2 ∈ [0, π] drawn once per individual, P the incumbent
// (destination) position and x1, x2 the golden section points. goldenFactor
// scales the second term; the published rule is recovered at goldenFactor = 1.
//
// rng must not be nil (ensured by caller).
// Returns: updated position vector.
func goldenSineUpdate(position, best []float64, goldenFactor float64, section goldenSection,
	lowerBound, upperBound float64, rng *rand.Rand,
) []float64 {
	size := len(position)
	newPos := make([]float64, size)

	// r1 determines the distance traveled, r2 the direction; both are drawn
	// once per individual, not per dimension.
	r1 := rng.Float64() * 2 * math.Pi
	r2 := rng.Float64() * math.Pi

	sinR1 := math.Sin(r1)
	absSinR1 := math.Abs(sinR1)

	for i := range size {
		newPos[i] = position[i]*absSinR1 -
			goldenFactor*r2*sinR1*math.Abs(section.x1*best[i]-section.x2*position[i])
	}

	// Apply boundary constraints
	maxVec(newPos, lowerBound)
	minVec(newPos, upperBound)

	return newPos
}
