// Orthogonal Learning Implementation.
//
// Implements Orthogonal Experimental Design for systematic parameter exploration.
//
// Reference:
// Zhan, Z. H., Zhang, J., Li, Y., & Shi, Y. H. (2010). Orthogonal Learning
// Particle Swarm Optimization. IEEE Transactions on Evolutionary Computation,
// 15(6), 832-847.
// DOI: 10.1109/TEVC.2010.2052054
//
// Leung, Y., & Wang, Y. (2001). An orthogonal genetic algorithm with
// quantization for global numerical optimization. IEEE Transactions on
// Evolutionary Computation, 5(1), 41-53.
//
// Orthogonal learning uses orthogonal arrays (e.g., L4) to systematically
// explore combinations of position, personal best, and global best.
// Increases diversity and reduces oscillatory movement in OLCE-MA variant.

package mayfly

import (
	"math"
	"math/bits"
	"math/rand"
	"sort"
)

// OrthogonalArray returns a fresh two-level orthogonal array with at least
// dimensions pairwise-balanced columns. The returned matrix may be modified
// by the caller without affecting subsequent calls.
func OrthogonalArray(dimensions int) [][]int {
	return orthogonalArray(dimensions)
}

// orthogonalArray constructs a regular two-level array from all non-zero
// binary interaction columns. With 2^k rows there are 2^k-1 balanced,
// pairwise-orthogonal columns, so no dimensions have to alias an L4 column.
func orthogonalArray(dimensions int) [][]int {
	if dimensions <= 0 {
		return nil
	}

	rows := 2
	for rows-1 < dimensions {
		rows *= 2
	}

	array := make([][]int, rows)
	for row := range rows {
		array[row] = make([]int, dimensions)
		for column := range dimensions {
			array[row][column] = bits.OnesCount(uint(row&(column+1))) & 1
		}
	}

	return array
}

// ApplyOrthogonalLearning applies orthogonal learning to a male mayfly.
// This creates a dimension-sized orthogonal design plus one factor-analysis
// candidate to explore the space between the mayfly's current position,
// personal best, and global best without aliased columns.
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
//
// A factor of zero disables the step: every candidate would collapse onto the
// male itself, so male is returned unchanged without spending any evaluation.
func ApplyOrthogonalLearning(male *Mayfly, pbest, gbest []float64, factor float64,
	lb, ub []float64, objFunc func([]float64) float64, rng *rand.Rand,
) *Mayfly {
	if !validOrthogonalInputs(male, pbest, gbest, factor, lb, ub) || objFunc == nil {
		return male
	}

	return applyOrthogonalLearning(
		male, pbest, gbest, factor, lb, ub, newConstraintEvaluator(objFunc, nil), rng,
	)
}

func applyOrthogonalLearning(male *Mayfly, pbest, gbest []float64, factor float64,
	lb, ub []float64, evaluator *constraintEvaluator, rng *rand.Rand,
) *Mayfly {
	// With a zero factor both the blend and the perturbation vanish, so every
	// candidate is bit-identical to the male. Evaluating them would burn the
	// full L4 budget on a guaranteed no-op.
	if factor <= 0 {
		return male
	}

	dim := len(male.Position)
	array := orthogonalArray(dim)
	candidates := make([]*Mayfly, len(array))
	_ = rng // The paper's orthogonal design is deterministic once its factors are fixed.

	// Generate candidates using orthogonal array
	for i := range array {
		candidate := newMayfly(dim)

		// For each dimension
		for j := range dim {
			// Based on orthogonal array entry, choose between three positions
			// Entry 0: Use current position with factor towards pbest
			// Entry 1: Use current position with factor towards gbest
			var pos float64
			if array[i][j] == 0 {
				// Blend current position with personal best
				pos = male.Position[j] + factor*(pbest[j]-male.Position[j])
			} else {
				// Blend current position with global best
				pos = male.Position[j] + factor*(gbest[j]-male.Position[j])
			}

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
		evaluator.evaluateMayfly(candidate, false)
		candidates[i] = candidate
	}

	// Factor analysis: rank the experiments, total each level's ranks for
	// every dimension, and assemble the preferred level of each factor into
	// one predicted best combination. Ranking retains the evaluator's
	// constraint semantics and avoids overflow from summing raw objective
	// values with unrelated magnitudes.
	ranked := append([]*Mayfly(nil), candidates...)
	sort.SliceStable(ranked, func(i, j int) bool {
		return evaluator.betterMayfly(ranked[i], ranked[j])
	})

	rankByCandidate := make(map[*Mayfly]float64, len(ranked))
	for rank, candidate := range ranked {
		rankByCandidate[candidate] = float64(rank + 1)
	}

	predicted := newMayfly(dim)
	for dimension := range dim {
		levelScores := [2]float64{}
		for experiment, candidate := range candidates {
			levelScores[array[experiment][dimension]] += rankByCandidate[candidate]
		}

		level := 0
		if levelScores[1] < levelScores[0] {
			level = 1
		}

		if level == 0 {
			predicted.Position[dimension] = male.Position[dimension] +
				factor*(pbest[dimension]-male.Position[dimension])
		} else {
			predicted.Position[dimension] = male.Position[dimension] +
				factor*(gbest[dimension]-male.Position[dimension])
		}

		predicted.Position[dimension] = min(max(predicted.Position[dimension], lb[dimension]), ub[dimension])
	}

	evaluator.evaluateMayfly(predicted, false)
	candidates = append(candidates, predicted)

	// Select best candidate
	best := candidates[0]
	for i := 1; i < len(candidates); i++ {
		if evaluator.betterMayfly(candidates[i], best) {
			best = candidates[i]
		}
	}

	// Only return improved solution
	if evaluator.betterMayfly(best, male) {
		// Copy velocity from original male (maintain momentum)
		copy(best.Velocity, male.Velocity)
		copy(best.Best.Position, male.Best.Position)
		best.Best.Cost = male.Best.Cost

		best.Best.ConstraintViolation = male.Best.ConstraintViolation
		if evaluator.better(evaluationFromMayfly(best), evaluationFromBest(best.Best)) {
			copy(best.Best.Position, best.Position)
			best.Best.Cost = best.Cost
			best.Best.ConstraintViolation = best.ConstraintViolation
		}

		return best
	}

	// If no improvement, return original male
	return male
}

func validOrthogonalInputs(male *Mayfly, pbest, gbest []float64, factor float64, lb, ub []float64) bool {
	if male == nil || math.IsNaN(factor) || math.IsInf(factor, 0) || factor < 0 {
		return false
	}

	dim := len(male.Position)
	if dim == 0 || len(pbest) != dim || len(gbest) != dim || len(lb) != dim || len(ub) != dim {
		return false
	}

	for i := range dim {
		if math.IsNaN(lb[i]) || math.IsInf(lb[i], 0) || math.IsNaN(ub[i]) || math.IsInf(ub[i], 0) || lb[i] > ub[i] {
			return false
		}
	}

	return true
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
//
// A factor of zero disables the step and spends no evaluations.
func ApplyOrthogonalLearningToElite(males []*Mayfly, topPercent float64,
	gbest []float64, factor float64, lb, ub []float64,
	objFunc func([]float64) float64, rng *rand.Rand,
) {
	if len(males) == 0 || objFunc == nil || math.IsNaN(topPercent) || math.IsInf(topPercent, 0) || topPercent <= 0 {
		return
	}

	applyOrthogonalLearningToElite(
		males, topPercent, gbest, factor, lb, ub, newConstraintEvaluator(objFunc, nil), rng,
	)
}

func applyOrthogonalLearningToElite(males []*Mayfly, topPercent float64,
	gbest []float64, factor float64, lb, ub []float64,
	evaluator *constraintEvaluator, rng *rand.Rand,
) {
	// A zero factor cannot move any male, so skip the stage entirely.
	if factor <= 0 || len(males) == 0 || topPercent <= 0 {
		return
	}

	// Calculate number of elite males to improve
	numElite := min(max(int(float64(len(males))*topPercent), 1), len(males))

	// Apply orthogonal learning to elite males
	for i := range numElite {
		improved := applyOrthogonalLearning(
			males[i],
			males[i].Best.Position, // Use personal best position
			gbest,                  // Use global best
			factor,
			lb, ub,
			evaluator,
			rng,
		)

		// Update male if improved
		if evaluator.betterMayfly(improved, males[i]) {
			// Preserve the personal best history
			improved.Best.Position = make([]float64, len(males[i].Best.Position))
			copy(improved.Best.Position, males[i].Best.Position)
			improved.Best.Cost = males[i].Best.Cost
			improved.Best.ConstraintViolation = males[i].Best.ConstraintViolation

			// Update personal best if current is better
			if evaluator.better(
				evaluationFromMayfly(improved), evaluationFromBest(improved.Best),
			) {
				copy(improved.Best.Position, improved.Position)
				improved.Best.Cost = improved.Cost
				improved.Best.ConstraintViolation = improved.ConstraintViolation
			}

			males[i] = improved
		}
	}
}
