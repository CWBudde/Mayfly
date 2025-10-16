package mayfly

import (
	"math"
	"math/rand"
)

// The Golden Sine Algorithm (GSA) is a nature-inspired optimization algorithm
// that combines the golden ratio and sine function to balance exploration
// and exploitation. It's particularly effective for convergence acceleration.
//
// Key features:
// - Uses golden ratio (φ ≈ 1.618) for adaptive step sizing
// - Sine function provides oscillatory behavior for escaping local optima
// - Balances between global exploration and local exploitation
//
// Reference: Tanyildizi, E., & Demir, G. (2017). Golden Sine Algorithm:
// A Novel Math-Inspired Algorithm. Advances in Electrical and Computer Engineering

const (
	// GoldenRatio is the mathematical constant φ = (1 + √5) / 2 ≈ 1.618034.
	GoldenRatio = 1.618033988749895
)

// Returns: updated position vector.
func goldenSineUpdate(position, best []float64, goldenFactor, lowerBound, upperBound float64, rng *rand.Rand) []float64 {
	if rng == nil {
		rng = rand.New(rand.NewSource(0))
	}

	size := len(position)
	newPos := make([]float64, size)

	for i := 0; i < size; i++ {
		// Generate random coefficients
		r1 := rng.Float64() * 2 * math.Pi // [0, 2π]
		r2 := rng.Float64() * 2 * math.Pi // [0, 2π]
		r3 := rng.Float64() * 2           // [0, 2]

		// Apply Golden Sine Algorithm formula
		// The sine function creates oscillatory movement
		// The absolute difference provides distance-based adaptation
		distance := math.Abs(r3*best[i] - position[i])
		update := goldenFactor * r1 * math.Sin(r2) * distance

		newPos[i] = position[i] + update
	}

	// Apply boundary constraints
	maxVec(newPos, lowerBound)
	minVec(newPos, upperBound)

	return newPos
}

// Returns: updated position vector.
func goldenSineUpdateAdaptive(position, best []float64, goldenFactor float64,
	currentIter, maxIter int, lowerBound, upperBound float64, rng *rand.Rand) []float64 {
	// Calculate adaptive factor: decreases from 2 to 1 over iterations
	iterRatio := float64(currentIter) / float64(maxIter)
	adaptiveFactor := goldenFactor * (2.0 - iterRatio)

	return goldenSineUpdate(position, best, adaptiveFactor, lowerBound, upperBound, rng)
}

// Returns: number of function evaluations performed.
func applyGoldenSineToElite(mayflies []*Mayfly, eliteRatio float64, globalBest []float64,
	goldenFactor float64, currentIter, maxIter int, lowerBound, upperBound float64,
	objectiveFunc ObjectiveFunction, rng *rand.Rand) int {
	numElite := int(float64(len(mayflies)) * eliteRatio)
	if numElite < 1 {
		numElite = 1
	}

	if numElite > len(mayflies) {
		numElite = len(mayflies)
	}

	funcEvals := 0

	// Apply Golden Sine Algorithm to elite mayflies
	for i := 0; i < numElite; i++ {
		// Generate candidate position using adaptive Golden Sine
		candidatePos := goldenSineUpdateAdaptive(
			mayflies[i].Position,
			globalBest,
			goldenFactor,
			currentIter,
			maxIter,
			lowerBound,
			upperBound,
			rng,
		)

		// Evaluate candidate
		candidateCost := objectiveFunc(candidatePos)
		funcEvals++

		// Accept if better (greedy selection)
		if candidateCost < mayflies[i].Cost {
			copy(mayflies[i].Position, candidatePos)
			mayflies[i].Cost = candidateCost

			// Update personal best if applicable
			if candidateCost < mayflies[i].Best.Cost {
				copy(mayflies[i].Best.Position, candidatePos)
				mayflies[i].Best.Cost = candidateCost
			}
		}
	}

	return funcEvals
}

// Returns: updated position vector.
func goldenSineConvergence(position, best []float64, goldenFactor, lowerBound, upperBound float64, rng *rand.Rand) []float64 {
	if rng == nil {
		rng = rand.New(rand.NewSource(0))
	}

	size := len(position)
	newPos := make([]float64, size)

	// Calculate average distance to best as convergence indicator
	avgDistance := 0.0
	for i := 0; i < size; i++ {
		avgDistance += math.Abs(position[i] - best[i])
	}

	avgDistance /= float64(size)

	// Normalize by search space
	searchSpace := upperBound - lowerBound
	convergenceFactor := avgDistance / searchSpace

	// Adjust golden factor based on convergence
	// Close to best: smaller factor (exploitation)
	// Far from best: larger factor (exploration)
	adaptedFactor := goldenFactor * (0.5 + convergenceFactor)

	for i := 0; i < size; i++ {
		r1 := rng.Float64() * 2 * math.Pi
		r2 := rng.Float64() * 2 * math.Pi
		r3 := rng.Float64() * 2

		distance := math.Abs(r3*best[i] - position[i])
		update := adaptedFactor * r1 * math.Sin(r2) * distance

		newPos[i] = position[i] + update
	}

	maxVec(newPos, lowerBound)
	minVec(newPos, upperBound)

	return newPos
}
