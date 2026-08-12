package mayfly

import (
	"fmt"
	"math"
)

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

// IsFeasible reports whether an aggregate constraint violation is zero.
func IsFeasible(violation float64) bool {
	return violation == 0
}

// PenalizedCost applies a linear or quadratic penalty to a raw objective cost.
// An empty method defaults to quadratic.
func PenalizedCost(cost, violation, factor float64, method PenaltyMethod) float64 {
	if method == "" {
		method = PenaltyQuadratic
	}

	if method == PenaltyLinear {
		return cost + factor*violation
	}

	return cost + factor*violation*violation
}

// BetterConstrainedCandidate reports whether candidate is preferred over
// incumbent under config. A nil config uses ordinary objective minimization.
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

func (evaluator *constraintEvaluator) evaluate(position []float64, sanitize bool) CandidateEvaluation {
	constraint := EvaluateConstraints(position, evaluator.constraints)

	cost := evaluator.objective(position)
	if sanitize {
		cost = sanitizeCost(cost)
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
