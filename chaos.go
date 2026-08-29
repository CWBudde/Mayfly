package mayfly

import (
	"errors"
	"fmt"
	"math"
)

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
//
// Deprecated: use NewLogisticMapChecked. This compatibility constructor
// normalizes invalid seeds.
func NewLogisticMap(seed float64) *LogisticMap {
	seed = normalizeLogisticSeed(seed)

	return &LogisticMap{
		x: seed,
		r: 4.0, // Standard value for fully chaotic behavior
	}
}

// NewLogisticMapChecked constructs a map only for a finite seed in (0,1).
func NewLogisticMapChecked(seed float64) (*LogisticMap, error) {
	err := validateLogisticSeed(seed)
	if err != nil {
		return nil, err
	}

	return &LogisticMap{x: seed, r: 4}, nil
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

// NextChecked validates receiver state before advancing the sequence.
func (lm *LogisticMap) NextChecked() (float64, error) {
	if lm == nil {
		return 0, errors.New("logistic map is nil")
	}

	err := validateLogisticSeed(lm.x)
	if err != nil {
		return 0, fmt.Errorf("invalid logistic-map state: %w", err)
	}

	if !isFinite(lm.r) || lm.r <= 0 || lm.r > 4 {
		return 0, fmt.Errorf("logistic-map control parameter must be in (0,4], got %v", lm.r)
	}

	return lm.Next(), nil
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

// CurrentChecked returns the current finite state or an error for a nil or
// invalid map.
func (lm *LogisticMap) CurrentChecked() (float64, error) {
	if lm == nil {
		return 0, errors.New("logistic map is nil")
	}

	err := validateLogisticSeed(lm.x)
	if err != nil {
		return 0, fmt.Errorf("invalid logistic-map state: %w", err)
	}

	return lm.x, nil
}

// Reset resets the map to a new seed value.
// This allows reusing the same LogisticMap instance with a different
// starting point.
//
// Deprecated: use ResetChecked. This compatibility method normalizes invalid
// seeds and ignores a nil receiver.
func (lm *LogisticMap) Reset(seed float64) {
	if lm != nil {
		lm.x = normalizeLogisticSeed(seed)
	}
}

// ResetChecked validates the receiver and seed before resetting the map.
func (lm *LogisticMap) ResetChecked(seed float64) error {
	if lm == nil {
		return errors.New("logistic map is nil")
	}

	err := validateLogisticSeed(seed)
	if err != nil {
		return err
	}

	lm.x = seed

	return nil
}

func validateLogisticSeed(seed float64) error {
	if !isFinite(seed) || seed <= 0 || seed >= 1 {
		return fmt.Errorf("logistic-map seed must be finite and in (0,1), got %v", seed)
	}

	return nil
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

// chaoticConstrictionFactor converts the zero-based implementation iteration
// to the one-based generation used by Mayfly's historical OLCE compatibility
// equation s=(G-g+1)/G. The factor is one in the first generation and 1/G in
// the final generation. ChaosFactor is a compatibility multiplier; one keeps
// the historical behavior and zero disables the optional stage.
func chaoticConstrictionFactor(config *Config, iteration int) float64 {
	if config == nil || config.MaxIterations <= 0 {
		return 0
	}

	generation := min(max(iteration+1, 1), config.MaxIterations)

	return config.ChaosFactor * float64(config.MaxIterations-generation+1) /
		float64(config.MaxIterations)
}

// chaoticExploitationCandidate implements Mayfly's historical OLCE
// compatibility equation. Publisher pseudocode discovered after this stage was
// written loops over all crossover offspring, while indexed full-text prose
// names the fittest offspring. Both describe Chebyshev mutation, but its exact
// component equation is not available in accessible primary evidence.
// For every component C'=LB+C(UB-LB) is first mapped from the logistic
// sequence, then a crossover offspring O is replaced by
// O'=(1-s)O+sC'. Both slices must have the same length.
func chaoticExploitationCandidate(
	destination, source []float64, config *Config, chaosMap *LogisticMap, constriction float64,
) {
	width := config.UpperBound - config.LowerBound

	for j := range source {
		chaoticPosition := config.LowerBound + chaosMap.Next()*width
		destination[j] = (1-constriction)*source[j] + constriction*chaoticPosition
		destination[j] = min(max(destination[j], config.LowerBound), config.UpperBound)
	}
}

// applyChaoticExploitation forms and evaluates Mayfly's historical
// compatibility position for one caller-selected crossover offspring. The
// optimizer calls it for every crossover offspring to follow the publisher
// pseudocode loop, while retaining this Logistic equation until the prose versus
// pseudocode cardinality conflict and exact Chebyshev equation are resolved.
//
// Exactly one objective evaluation is spent per call. It returns true when the
// finite candidate was installed.
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
		chaoticConstrictionFactor(config, iteration),
	)
	evaluator.evaluateMayfly(candidate, false)

	return commitChaoticOffspring(target, candidate)
}

// commitChaoticOffspring replaces the crossover offspring with an already
// evaluated chaotic offspring and initializes its personal-best state.
// Candidates containing non-finite evaluation metadata are rejected as a
// library safety guard.
func commitChaoticOffspring(target, candidate *Mayfly) bool {
	if candidate.Cost == math.MaxFloat64 || !isFinite(candidate.Cost) ||
		!isFinite(candidate.ConstraintViolation) {
		return false
	}

	copy(target.Position, candidate.Position)
	target.Cost = candidate.Cost
	target.ConstraintViolation = candidate.ConstraintViolation
	copy(target.Best.Position, target.Position)
	target.Best.Cost = target.Cost
	target.Best.ConstraintViolation = target.ConstraintViolation

	return true
}
