package mayfly

import (
	"errors"
	"fmt"
	"math"
)

// ErrNoFiniteObjectiveValue reports that initialization could not establish a
// meaningful incumbent because every objective evaluation was non-finite.
var ErrNoFiniteObjectiveValue = errors.New("initial population produced no finite objective value")

// ConstraintEvaluation describes the aggregate constraint state of a
// position. A zero violation is feasible.
type ConstraintEvaluation struct {
	Violation float64
	Feasible  bool
}

// CandidateEvaluation contains the values used to compare two constrained
// candidates while keeping the raw objective cost intact.
type CandidateEvaluation struct {
	Cost                float64
	ConstraintViolation float64
}

// EvaluateConstraints evaluates and aggregates all configured constraints.
// Non-finite constraint values produce an infinite violation.
//
// Deprecated: use EvaluateConstraintsChecked to distinguish invalid
// constraints from a genuinely infeasible position.
func EvaluateConstraints(position []float64, config *ConstraintConfig) ConstraintEvaluation {
	if config == nil {
		return ConstraintEvaluation{Feasible: true}
	}

	violation := 0.0

	for _, constraint := range config.Inequalities {
		if constraint == nil {
			return ConstraintEvaluation{Violation: math.Inf(1)}
		}

		value := constraint(position)
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return ConstraintEvaluation{Violation: math.Inf(1)}
		}

		violation += max(0, value)
	}

	for _, constraint := range config.Equalities {
		if constraint == nil {
			return ConstraintEvaluation{Violation: math.Inf(1)}
		}

		value := constraint(position)
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return ConstraintEvaluation{Violation: math.Inf(1)}
		}

		violation += max(0, math.Abs(value)-config.EqualityTolerance)
	}

	if math.IsInf(violation, 1) {
		return ConstraintEvaluation{Violation: math.Inf(1)}
	}

	return ConstraintEvaluation{Violation: violation, Feasible: violation == 0}
}

// EvaluateConstraintsChecked evaluates constraints without converting nil,
// panicking, or non-finite functions into an infinite-violation sentinel.
func EvaluateConstraintsChecked(
	position []float64,
	config *ConstraintConfig,
) (evaluation ConstraintEvaluation, returnErr error) {
	if len(position) == 0 {
		return evaluation, errors.New("constraint position is empty")
	}
	for i, coordinate := range position {
		if !isFinite(coordinate) {
			return evaluation, fmt.Errorf("constraint position %d is not finite", i)
		}
	}
	if err := validateConstraintConfig(config); err != nil {
		return evaluation, err
	}
	if config == nil {
		return ConstraintEvaluation{Feasible: true}, nil
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			returnErr = fmt.Errorf("constraint function panicked: %v", recovered)
			evaluation = ConstraintEvaluation{}
		}
	}()
	violation := 0.0
	for i, constraint := range config.Inequalities {
		value := constraint(position)
		if !isFinite(value) {
			return ConstraintEvaluation{}, fmt.Errorf("inequality constraint %d returned a non-finite value", i)
		}
		violation += max(0, value)
	}
	for i, constraint := range config.Equalities {
		value := constraint(position)
		if !isFinite(value) {
			return ConstraintEvaluation{}, fmt.Errorf("equality constraint %d returned a non-finite value", i)
		}
		violation += max(0, math.Abs(value)-config.EqualityTolerance)
	}
	if !isFinite(violation) {
		return ConstraintEvaluation{}, errors.New("aggregate constraint violation overflowed")
	}
	return ConstraintEvaluation{Violation: violation, Feasible: violation == 0}, nil
}

// IsFeasible reports whether an aggregate constraint violation is zero.
func IsFeasible(violation float64) bool {
	return violation == 0
}

// PenalizedCost applies a linear or quadratic penalty to a raw objective cost.
// An empty method defaults to quadratic.
// Deprecated: use PenalizedCostChecked for validation.
func PenalizedCost(cost, violation, factor float64, method PenaltyMethod) float64 {
	if method == "" {
		method = PenaltyQuadratic
	}

	if method == PenaltyLinear {
		return cost + factor*violation
	}

	return cost + factor*violation*violation
}

// PenalizedCostChecked validates inputs and rejects overflow.
func PenalizedCostChecked(cost, violation, factor float64, method PenaltyMethod) (float64, error) {
	if !isFinite(cost) {
		return 0, fmt.Errorf("cost must be finite, got %v", cost)
	}
	if !isFinite(violation) || violation < 0 {
		return 0, fmt.Errorf("constraint violation must be finite and non-negative, got %v", violation)
	}
	if !isFinite(factor) || factor < 0 {
		return 0, fmt.Errorf("penalty factor must be finite and non-negative, got %v", factor)
	}
	if method != "" && method != PenaltyLinear && method != PenaltyQuadratic {
		return 0, fmt.Errorf("unknown penalty method %q", method)
	}
	result := PenalizedCost(cost, violation, factor, method)
	if !isFinite(result) {
		return 0, errors.New("penalized cost overflowed")
	}
	return result, nil
}

// BetterConstrainedCandidate reports whether candidate is preferred over
// incumbent under config. A nil config uses ordinary objective minimization.
// Deprecated: use BetterConstrainedCandidateChecked for validation.
func BetterConstrainedCandidate(candidate, incumbent CandidateEvaluation, config *ConstraintConfig) bool {
	if config == nil {
		return candidate.Cost < incumbent.Cost
	}

	handling := ConstraintHandlingFeasibility
	if config.Handling != "" {
		handling = config.Handling
	}

	if handling == ConstraintHandlingPenalty {
		candidateScore := PenalizedCost(
			candidate.Cost,
			candidate.ConstraintViolation,
			config.PenaltyFactor,
			config.PenaltyMethod,
		)
		incumbentScore := PenalizedCost(
			incumbent.Cost,
			incumbent.ConstraintViolation,
			config.PenaltyFactor,
			config.PenaltyMethod,
		)

		if candidateScore != incumbentScore {
			return candidateScore < incumbentScore
		}
	}

	candidateFeasible := IsFeasible(candidate.ConstraintViolation)

	incumbentFeasible := IsFeasible(incumbent.ConstraintViolation)
	if candidateFeasible != incumbentFeasible {
		return candidateFeasible
	}

	if !candidateFeasible && candidate.ConstraintViolation != incumbent.ConstraintViolation {
		return candidate.ConstraintViolation < incumbent.ConstraintViolation
	}

	return candidate.Cost < incumbent.Cost
}

// BetterConstrainedCandidateChecked validates both candidates and the ranking
// policy before comparing them.
func BetterConstrainedCandidateChecked(
	candidate, incumbent CandidateEvaluation,
	config *ConstraintConfig,
) (bool, error) {
	if err := validateConstraintConfig(config); err != nil {
		return false, err
	}
	for name, evaluation := range map[string]CandidateEvaluation{
		"candidate": candidate, "incumbent": incumbent,
	} {
		if !isFinite(evaluation.Cost) {
			return false, fmt.Errorf("%s cost must be finite", name)
		}
		if !isFinite(evaluation.ConstraintViolation) || evaluation.ConstraintViolation < 0 {
			return false, fmt.Errorf("%s constraint violation must be finite and non-negative", name)
		}
	}
	return BetterConstrainedCandidate(candidate, incumbent, config), nil
}

func validateConstraintConfig(config *ConstraintConfig) error {
	if config == nil {
		return nil
	}

	if math.IsNaN(config.EqualityTolerance) || math.IsInf(config.EqualityTolerance, 0) ||
		config.EqualityTolerance < 0 {
		return fmt.Errorf("equality tolerance must be finite and non-negative, got %v", config.EqualityTolerance)
	}

	for i, constraint := range config.Inequalities {
		if constraint == nil {
			return fmt.Errorf("inequality constraint %d is nil", i)
		}
	}

	for i, constraint := range config.Equalities {
		if constraint == nil {
			return fmt.Errorf("equality constraint %d is nil", i)
		}
	}

	if math.IsNaN(config.PenaltyFactor) || math.IsInf(config.PenaltyFactor, 0) || config.PenaltyFactor < 0 {
		return fmt.Errorf("penalty factor must be finite and non-negative, got %v", config.PenaltyFactor)
	}

	switch config.PenaltyMethod {
	case "", PenaltyLinear, PenaltyQuadratic:
	default:
		return fmt.Errorf("unknown penalty method %q", config.PenaltyMethod)
	}

	switch config.Handling {
	case "", ConstraintHandlingFeasibility:
		return nil
	case ConstraintHandlingPenalty:
		if config.PenaltyFactor == 0 {
			return fmt.Errorf("penalty factor must be finite and positive, got %v", config.PenaltyFactor)
		}

		return nil
	default:
		return fmt.Errorf("unknown constraint handling method %q", config.Handling)
	}
}

type constraintEvaluator struct {
	objective   ObjectiveFunction
	constraints *ConstraintConfig
}

func newConstraintEvaluator(objective ObjectiveFunction, constraints *ConstraintConfig) *constraintEvaluator {
	return &constraintEvaluator{objective: objective, constraints: constraints}
}

func (evaluator *constraintEvaluator) evaluate(position []float64, _ bool) CandidateEvaluation {
	constraint := EvaluateConstraints(position, evaluator.constraints)

	cost := evaluator.objective(position)
	if !isFinite(cost) {
		// This package only minimizes. NaN and both infinities are invalid
		// evaluations, never evidence that a candidate is exceptionally good.
		cost = math.MaxFloat64
	}

	return CandidateEvaluation{
		Cost:                cost,
		ConstraintViolation: constraint.Violation,
	}
}

func (evaluator *constraintEvaluator) evaluateMayfly(mayfly *Mayfly, sanitize bool) {
	evaluation := evaluator.evaluate(mayfly.Position, sanitize)
	mayfly.Cost = evaluation.Cost
	mayfly.ConstraintViolation = evaluation.ConstraintViolation
}

func (evaluator *constraintEvaluator) better(candidate, incumbent CandidateEvaluation) bool {
	return BetterConstrainedCandidate(candidate, incumbent, evaluator.constraints)
}

func (evaluator *constraintEvaluator) betterMayfly(candidate, incumbent *Mayfly) bool {
	return evaluator.better(evaluationFromMayfly(candidate), evaluationFromMayfly(incumbent))
}

func (evaluator *constraintEvaluator) betterMayflyThanBest(candidate *Mayfly, incumbent Best) bool {
	return evaluator.better(evaluationFromMayfly(candidate), evaluationFromBest(incumbent))
}

func (evaluator *constraintEvaluator) betterBest(candidate, incumbent Best) bool {
	return evaluator.better(evaluationFromBest(candidate), evaluationFromBest(incumbent))
}

func (evaluator *constraintEvaluator) acceptanceProbability(
	current, candidate CandidateEvaluation,
	temperature float64,
) float64 {
	if evaluator.better(candidate, current) {
		return 1
	}

	handling := ConstraintHandlingFeasibility
	if evaluator.constraints != nil && evaluator.constraints.Handling != "" {
		handling = evaluator.constraints.Handling
	}

	var currentScore, candidateScore float64
	if handling == ConstraintHandlingPenalty {
		currentScore = PenalizedCost(
			current.Cost, current.ConstraintViolation,
			evaluator.constraints.PenaltyFactor, evaluator.constraints.PenaltyMethod,
		)
		candidateScore = PenalizedCost(
			candidate.Cost, candidate.ConstraintViolation,
			evaluator.constraints.PenaltyFactor, evaluator.constraints.PenaltyMethod,
		)
	} else {
		currentFeasible := IsFeasible(current.ConstraintViolation)

		candidateFeasible := IsFeasible(candidate.ConstraintViolation)
		if currentFeasible != candidateFeasible {
			return 0
		}

		if currentFeasible {
			currentScore = current.Cost
			candidateScore = candidate.Cost
		} else {
			currentScore = current.ConstraintViolation
			candidateScore = candidate.ConstraintViolation
		}
	}

	return math.Exp(-(candidateScore - currentScore) / temperature)
}

func evaluationFromMayfly(mayfly *Mayfly) CandidateEvaluation {
	return CandidateEvaluation{Cost: mayfly.Cost, ConstraintViolation: mayfly.ConstraintViolation}
}

func evaluationFromBest(best Best) CandidateEvaluation {
	return CandidateEvaluation{Cost: best.Cost, ConstraintViolation: best.ConstraintViolation}
}

func bestFromMayfly(mayfly *Mayfly) Best {
	return Best{
		Position:            append([]float64(nil), mayfly.Position...),
		Cost:                mayfly.Cost,
		ConstraintViolation: mayfly.ConstraintViolation,
	}
}

func copyMayflyToBest(destination *Best, source *Mayfly) {
	destination.Cost = source.Cost
	destination.ConstraintViolation = source.ConstraintViolation
	copy(destination.Position, source.Position)
}
