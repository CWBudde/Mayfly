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
	"errors"
	"fmt"
	"math"
	"math/bits"
	"math/rand"
	"sort"
)

// MaxOrthogonalArrayDimensions bounds the allocation made by the checked
// public constructor to roughly eight million bytes of array payload.
const MaxOrthogonalArrayDimensions = 1023

// OrthogonalArray returns a fresh two-level orthogonal array with at least
// dimensions pairwise-balanced columns. The returned matrix may be modified
// by the caller without affecting subsequent calls.
//
// Deprecated: use OrthogonalArrayChecked to reject invalid or impractically
// large dimensions explicitly.
func OrthogonalArray(dimensions int) [][]int {
	return orthogonalArray(dimensions)
}

// OrthogonalArrayChecked constructs a bounded orthogonal array.
func OrthogonalArrayChecked(dimensions int) ([][]int, error) {
	if dimensions <= 0 {
		return nil, fmt.Errorf("orthogonal-array dimensions must be positive, got %d", dimensions)
	}
	if dimensions > MaxOrthogonalArrayDimensions {
		return nil, fmt.Errorf(
			"orthogonal-array dimensions %d exceed the supported maximum %d",
			dimensions, MaxOrthogonalArrayDimensions,
		)
	}
	return orthogonalArray(dimensions), nil
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
//
// Deprecated: use ApplyOrthogonalLearningChecked. This compatibility wrapper
// returns the input male unchanged when validation fails.
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

// ApplyOrthogonalLearningChecked validates all dimensions, finite values,
// bounds, and objective results before returning a candidate.
func ApplyOrthogonalLearningChecked(
	male *Mayfly,
	pbest, gbest []float64,
	factor float64,
	lb, ub []float64,
	objFunc ObjectiveFunction,
	rng *rand.Rand,
) (*Mayfly, error) {
	if err := validateOrthogonalInputs(male, pbest, gbest, factor, lb, ub); err != nil {
		return nil, err
	}
	if objFunc == nil {
		return nil, errors.New("orthogonal-learning objective function is nil")
	}
	var objectiveErr error
	checkedObjective := func(position []float64) (value float64) {
		if objectiveErr != nil {
			return math.NaN()
		}
		defer func() {
			if recovered := recover(); recovered != nil {
				objectiveErr = fmt.Errorf("orthogonal-learning objective panicked: %v", recovered)
				value = math.NaN()
			}
		}()
		value = objFunc(position)
		if !isFinite(value) {
			objectiveErr = errors.New("orthogonal-learning objective returned a non-finite value")
		}
		return value
	}
	result := applyOrthogonalLearning(
		male, pbest, gbest, factor, lb, ub, newConstraintEvaluator(checkedObjective, nil), rng,
	)
	if objectiveErr != nil {
		return nil, objectiveErr
	}
	return result, nil
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
	return validateOrthogonalInputs(male, pbest, gbest, factor, lb, ub) == nil
}

func validateOrthogonalInputs(male *Mayfly, pbest, gbest []float64, factor float64, lb, ub []float64) error {
	if male == nil {
		return errors.New("orthogonal-learning male is nil")
	}
	if !isFinite(factor) || factor < 0 || factor > 1 {
		return fmt.Errorf("orthogonal factor must be in [0,1], got %v", factor)
	}
	dim := len(male.Position)
	if dim == 0 {
		return errors.New("orthogonal-learning position is empty")
	}
	for name, vector := range map[string][]float64{
		"velocity": male.Velocity, "personal best": pbest, "global best": gbest,
		"lower bounds": lb, "upper bounds": ub,
	} {
		if len(vector) != dim {
			return fmt.Errorf("%s dimension is %d, want %d", name, len(vector), dim)
		}
	}
	for i := range dim {
		values := []float64{male.Position[i], pbest[i], gbest[i], lb[i], ub[i]}
		for _, value := range values {
			if !isFinite(value) {
				return fmt.Errorf("orthogonal-learning dimension %d contains a non-finite value", i)
			}
		}
		if lb[i] > ub[i] {
			return fmt.Errorf("lower bound %v exceeds upper bound %v at dimension %d", lb[i], ub[i], i)
		}
		if male.Position[i] < lb[i] || male.Position[i] > ub[i] ||
			pbest[i] < lb[i] || pbest[i] > ub[i] || gbest[i] < lb[i] || gbest[i] > ub[i] {
			return fmt.Errorf("orthogonal-learning position is outside bounds at dimension %d", i)
		}
	}
	return nil
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
//
// Deprecated: use ApplyOrthogonalLearningToEliteChecked. This compatibility
// wrapper silently ignores invalid input.
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

// ApplyOrthogonalLearningToEliteChecked validates every selected male and
// reports invalid population metadata instead of silently doing nothing.
func ApplyOrthogonalLearningToEliteChecked(
	males []*Mayfly,
	topPercent float64,
	gbest []float64,
	factor float64,
	lb, ub []float64,
	objFunc ObjectiveFunction,
	rng *rand.Rand,
) error {
	if len(males) == 0 {
		return errors.New("orthogonal-learning population is empty")
	}
	if !isFinite(topPercent) || topPercent <= 0 || topPercent > 1 {
		return fmt.Errorf("elite percentage must be in (0,1], got %v", topPercent)
	}
	if objFunc == nil {
		return errors.New("orthogonal-learning objective function is nil")
	}
	for i, male := range males {
		if male == nil {
			return fmt.Errorf("male %d is nil", i)
		}
		if err := validateOrthogonalInputs(male, male.Best.Position, gbest, factor, lb, ub); err != nil {
			return fmt.Errorf("male %d: %w", i, err)
		}
	}
	var objectiveErr error
	checkedObjective := func(position []float64) (value float64) {
		if objectiveErr != nil {
			return math.NaN()
		}
		defer func() {
			if recovered := recover(); recovered != nil {
				objectiveErr = fmt.Errorf("orthogonal-learning objective panicked: %v", recovered)
				value = math.NaN()
			}
		}()
		value = objFunc(position)
		if !isFinite(value) {
			objectiveErr = errors.New("orthogonal-learning objective returned a non-finite value")
		}
		return value
	}
	working := make([]*Mayfly, len(males))
	for i, male := range males {
		working[i] = male.clone()
	}
	applyOrthogonalLearningToElite(
		working, topPercent, gbest, factor, lb, ub, newConstraintEvaluator(checkedObjective, nil), rng,
	)
	if objectiveErr != nil {
		return objectiveErr
	}
	copy(males, working)
	return nil
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
