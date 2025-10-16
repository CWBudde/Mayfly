package mayfly

import (
	"math/rand"
)

// L4Array is a standard L4(2^3) orthogonal array.
// This is a 4x3 matrix where each column contains two 0s and two 1s,
// and all pairs of columns are balanced (each (0,0), (0,1), (1,0), (1,1)
// combination appears exactly once).
//
// This orthogonal array is used to systematically explore the parameter
// space with minimal experiments while maintaining statistical balance.
var L4Array = [][]int{
	{0, 0, 0},
	{0, 1, 1},
	{1, 0, 1},
	{1, 1, 0},
}

// ApplyOrthogonalLearning applies orthogonal learning to a male mayfly.
// This creates 4 candidate solutions using the L4 orthogonal array to
// systematically explore the space between the mayfly's current position,
// personal best, and global best.
//
// The orthogonal learning strategy increases population diversity and
// reduces oscillatory movement by testing balanced combinations of
// position updates.
//
// Parameters:
//   - male: The mayfly to apply orthogonal learning to
//   - pbest: The personal best position of the male
//   - gbest: The global best position found so far
//   - factor: The orthogonal learning strength factor (typically 0.3)
//   - lb: Lower bounds of the search space
//   - ub: Upper bounds of the search space
//   - rng: Random number generator for tie-breaking
//
// Returns:
//   - A new Mayfly representing the best candidate from the orthogonal exploration
func ApplyOrthogonalLearning(male *Mayfly, pbest, gbest []float64, factor float64,
	lb, ub []float64, objFunc func([]float64) float64, rng *rand.Rand) *Mayfly {
	dim := len(male.Position)
	candidates := make([]*Mayfly, len(L4Array))

	// Generate candidates using orthogonal array
	for i := 0; i < len(L4Array); i++ {
		candidate := newMayfly(dim)

		// For each dimension
		for j := 0; j < dim; j++ {
			// Select dimension mapping using modulo for dimensions > 3
			arrayCol := j % 3

			// Based on orthogonal array entry, choose between three positions
			// Entry 0: Use current position with factor towards pbest
			// Entry 1: Use current position with factor towards gbest
			var pos float64
			if L4Array[i][arrayCol] == 0 {
				// Blend current position with personal best
				pos = male.Position[j] + factor*(pbest[j]-male.Position[j])
			} else {
				// Blend current position with global best
				pos = male.Position[j] + factor*(gbest[j]-male.Position[j])
			}

			// Add small random perturbation for diversity
			perturbation := (rng.Float64()*2.0 - 1.0) * factor * 0.1
			pos += perturbation * (ub[j] - lb[j])

			// Apply bounds
			if pos < lb[j] {
				pos = lb[j]
			}

			if pos > ub[j] {
				pos = ub[j]
			}

			candidate.Position[j] = pos
		}

		// Evaluate candidate
		candidate.Cost = objFunc(candidate.Position)
		candidates[i] = candidate
	}

	// Select best candidate
	best := candidates[0]
	for i := 1; i < len(candidates); i++ {
		if candidates[i].Cost < best.Cost {
			best = candidates[i]
		}
	}

	// Only return improved solution
	if best.Cost < male.Cost {
		// Copy velocity from original male (maintain momentum)
		copy(best.Velocity, male.Velocity)
		return best
	}

	// If no improvement, return original male
	return male
}

// ApplyOrthogonalLearningToElite applies orthogonal learning to the
// top-performing males in the population. This is more efficient than
// applying to all males while still improving the best solutions.
//
// Parameters:
//   - males: The male population (assumed to be sorted by fitness)
//   - topPercent: Percentage of top males to apply OL to (e.g., 0.2 for top 20%)
//   - gbest: Global best position
//   - factor: Orthogonal learning strength
//   - lb, ub: Search space bounds
//   - objFunc: Objective function
//   - rng: Random number generator
//
// Returns:
//   - The males slice with top performers improved via orthogonal learning
func ApplyOrthogonalLearningToElite(males []*Mayfly, topPercent float64,
	gbest []float64, factor float64, lb, ub []float64,
	objFunc func([]float64) float64, rng *rand.Rand) {
	// Calculate number of elite males to improve
	numElite := int(float64(len(males)) * topPercent)
	if numElite < 1 {
		numElite = 1
	}

	if numElite > len(males) {
		numElite = len(males)
	}

	// Apply orthogonal learning to elite males
	for i := 0; i < numElite; i++ {
		improved := ApplyOrthogonalLearning(
			males[i],
			males[i].Best.Position, // Use personal best position
			gbest,                  // Use global best
			factor,
			lb, ub,
			objFunc,
			rng,
		)

		// Update male if improved
		if improved.Cost < males[i].Cost {
			// Preserve the personal best history
			improved.Best.Position = make([]float64, len(males[i].Best.Position))
			copy(improved.Best.Position, males[i].Best.Position)
			improved.Best.Cost = males[i].Best.Cost

			// Update personal best if current is better
			if improved.Cost < improved.Best.Cost {
				copy(improved.Best.Position, improved.Position)
				improved.Best.Cost = improved.Cost
			}

			males[i] = improved
		}
	}
}
