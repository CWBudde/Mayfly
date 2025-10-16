package mayfly

import (
	"math/rand"
)

// oppositionPoint generates the opposition point of a given position.
// The opposition point is calculated as: x_opp = a + b - x
// where a is the lower bound and b is the upper bound.
func oppositionPoint(position []float64, lowerBound, upperBound float64) []float64 {
	result := make([]float64, len(position))
	for i := 0; i < len(position); i++ {
		result[i] = lowerBound + upperBound - position[i]
	}

	return result
}

// gaussianUpdate performs a Bare Bones update using Gaussian sampling.
// The new position is sampled from a Gaussian distribution with mean
// at the midpoint between current and best positions, and standard
// deviation based on the distance between them.
func gaussianUpdate(current, best []float64, lowerBound, upperBound float64, rng *rand.Rand) []float64 {
	result := make([]float64, len(current))

	for i := 0; i < len(current); i++ {
		// Mean is the midpoint between current and best
		mean := (current[i] + best[i]) / 2.0

		// Standard deviation is half the distance between current and best
		// If they're the same, use a small exploration factor
		stddev := (current[i] - best[i]) / 2.0
		if stddev < 0 {
			stddev = -stddev
		}

		if stddev < 1e-10 {
			// Small exploration when current and best are very close
			stddev = (upperBound - lowerBound) * 0.01
		}

		// Sample from Gaussian distribution
		result[i] = mean + randn(rng)*stddev

		// Apply bounds
		if result[i] < lowerBound {
			result[i] = lowerBound
		}

		if result[i] > upperBound {
			result[i] = upperBound
		}
	}

	return result
}
