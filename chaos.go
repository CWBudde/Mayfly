package mayfly

import "math"

// LogisticMap implements the logistic chaotic map.
// The logistic map is defined by: x_{n+1} = r * x_n * (1 - x_n)
// where r is the control parameter. When r = 4.0, the map exhibits
// fully chaotic behavior with good ergodic properties.
//
// The logistic map is widely used in chaos-based optimization algorithms
// due to its simplicity, well-understood dynamics, and ability to generate
// pseudo-random sequences with better uniformity than standard PRNGs.
type LogisticMap struct {
	x float64 // Current state value in [0, 1]
	r float64 // Control parameter (typically 4.0 for full chaos)
}

// NewLogisticMap creates a new logistic map with the given seed.
// The seed should be in the range (0, 1), exclusive of boundaries.
// If seed is outside this range, it will be normalized to (0, 1).
// The control parameter r is set to 4.0 for fully chaotic behavior.
func NewLogisticMap(seed float64) *LogisticMap {
	seed = normalizeLogisticSeed(seed)

	return &LogisticMap{
		x: seed,
		r: 4.0, // Standard value for fully chaotic behavior
	}
}

// Next generates and returns the next value in the chaotic sequence.
// The returned value is in the range [0, 1].
// This method updates the internal state and should be called
// sequentially to generate a chaotic sequence.
func (lm *LogisticMap) Next() float64 {
	if lm == nil {
		return math.NaN()
	}

	if math.IsNaN(lm.x) || math.IsInf(lm.x, 0) {
		lm.x = normalizeLogisticSeed(lm.x)
	}

	// Apply logistic map equation: x_{n+1} = r * x_n * (1 - x_n)
	lm.x = lm.r * lm.x * (1.0 - lm.x)

	// Safeguard against numerical drift to boundaries
	// The logistic map should naturally stay in (0,1) but floating point
	// errors can occasionally push values to exactly 0 or 1, which would
	// cause the sequence to collapse to a fixed point.
	if lm.x <= 0.0 {
		lm.x = 1e-10
	}

	if lm.x >= 1.0 {
		lm.x = 1.0 - 1e-10
	}

	return lm.x
}

// Current returns the current state value without advancing the sequence.
// This is useful for debugging or when you need to inspect the state
// without modifying it.
func (lm *LogisticMap) Current() float64 {
	if lm == nil {
		return math.NaN()
	}

	return lm.x
}

// Reset resets the map to a new seed value.
// This allows reusing the same LogisticMap instance with a different
// starting point.
func (lm *LogisticMap) Reset(seed float64) {
	if lm != nil {
		lm.x = normalizeLogisticSeed(seed)
	}
}

func normalizeLogisticSeed(seed float64) float64 {
	switch {
	case math.IsNaN(seed), math.IsInf(seed, -1):
		return 0.314159
	case math.IsInf(seed, 1):
		return 0.271828
	case seed > 0 && seed < 1:
		return seed
	}

	seed = 0.1 + 0.8*(seed-math.Trunc(seed))
	if seed <= 0 || math.IsNaN(seed) {
		return 0.314159
	}

	if seed >= 1 {
		return 0.271828
	}

	return seed
}

// olceElitePercent is the fraction of the sorted male population that the
// OLCE-MA refinement steps (orthogonal learning and chaotic exploitation)
// operate on.
const olceElitePercent = 0.2

// olceEliteCount returns how many of the sorted males the OLCE-MA refinement
// steps operate on. At least one male is always selected.
func olceEliteCount(population int) int {
	return min(max(int(float64(population)*olceElitePercent), 1), population)
}

// chaoticExploitationRadius returns the perturbation radius of the chaotic
// exploitation step for the given iteration.
//
// The radius decays linearly from ChaosFactor to zero over the run, which is
// the shrinking neighborhood the chaotic local search literature specifies:
// early iterations perturb widely to escape local optima, late iterations
// refine. A constant radius turns the step into a persistent random walk that
// prevents convergence.
//
// The optimization loop supplies the iteration indices 0 to MaxIterations-1,
// so progress is measured against the last applied iteration, MaxIterations-1.
// The radius therefore really reaches zero on the final iteration. A run of a
// single iteration has no decay to spread out and keeps the full ChaosFactor,
// because a zero radius would degrade its only exploitation step to a no-op.
func chaoticExploitationRadius(config *Config, iteration int) float64 {
	lastIteration := config.MaxIterations - 1
	if lastIteration <= 0 {
		return config.ChaosFactor
	}

	progress := min(max(float64(iteration)/float64(lastIteration), 0.0), 1.0)

	return config.ChaosFactor * (1.0 - progress)
}

// chaoticExploitationCandidate writes a chaotic neighbor of source into
// destination. Both slices must have the same length.
//
// The displacement of each dimension is drawn from the logistic map and scaled
// by radius and the width of the search space.
func chaoticExploitationCandidate(
	destination, source []float64, config *Config, chaosMap *LogisticMap, radius float64,
) {
	width := config.UpperBound - config.LowerBound

	for j := range source {
		perturbation := radius * (chaosMap.Next() - 0.5) * width
		destination[j] = min(max(source[j]+perturbation, config.LowerBound), config.UpperBound)
	}
}

// applyChaoticExploitation performs one chaotic local search step on target.
//
// A single chaotic neighbor is generated and evaluated, and target is only
// moved onto it when the neighbor is not worse (greedy acceptance). The cost
// of target therefore never increases, which is what separates chaotic
// exploitation from an unconditional random kick.
//
// Exactly one objective evaluation is spent per call. It returns true when the
// candidate was accepted.
func applyChaoticExploitation(
	target *Mayfly,
	config *Config,
	chaosMap *LogisticMap,
	iteration int,
	evaluator *constraintEvaluator,
) bool {
	candidate := newMayfly(len(target.Position))
	chaoticExploitationCandidate(
		candidate.Position, target.Position, config, chaosMap,
		chaoticExploitationRadius(config, iteration),
	)
	evaluator.evaluateMayfly(candidate, false)

	return acceptChaoticCandidate(target, candidate, evaluator)
}

// acceptChaoticCandidate applies greedy acceptance of an already evaluated
// chaotic candidate to target and keeps the personal best consistent.
//
// Candidates whose cost or constraint violation is NaN are rejected outright:
// every comparison against NaN is false, so the greedy guarantee would not
// hold for objectives that are undefined on part of the domain, and the NaN
// would spread through mating and sorting.
func acceptChaoticCandidate(
	target, candidate *Mayfly, evaluator *constraintEvaluator,
) bool {
	if math.IsNaN(candidate.Cost) || math.IsNaN(candidate.ConstraintViolation) {
		return false
	}

	// Greedy acceptance: reject only strictly worse candidates.
	if evaluator.betterMayfly(target, candidate) {
		return false
	}

	copy(target.Position, candidate.Position)
	target.Cost = candidate.Cost
	target.ConstraintViolation = candidate.ConstraintViolation

	if evaluator.better(evaluationFromMayfly(target), evaluationFromBest(target.Best)) {
		copy(target.Best.Position, target.Position)
		target.Best.Cost = target.Cost
		target.Best.ConstraintViolation = target.ConstraintViolation
	}

	return true
}
