package mayfly

import (
	"fmt"
	"math"
)

type convergenceTracker struct {
	config             *ConvergenceConfig
	evaluator          *constraintEvaluator
	referenceBest      CandidateEvaluation
	stagnantIterations int
}

func newConvergenceTracker(
	config *ConvergenceConfig,
	initialBest Best,
	evaluators ...*constraintEvaluator,
) *convergenceTracker {
	evaluator := newConstraintEvaluator(nil, nil)
	if len(evaluators) > 0 {
		evaluator = evaluators[0]
	}

	return &convergenceTracker{
		config:        config,
		referenceBest: evaluationFromBest(initialBest),
		evaluator:     evaluator,
	}
}

func (tracker *convergenceTracker) observe(iteration int, best Best) (TerminationReason, bool) {
	if tracker.config == nil {
		return "", false
	}

	bestEvaluation := evaluationFromBest(best)

	if tracker.significantlyImproved(bestEvaluation) {
		tracker.referenceBest = bestEvaluation
		tracker.stagnantIterations = 0
	} else {
		tracker.stagnantIterations++
	}

	minimumIterations := max(tracker.config.MinIterations, 1)
	if iteration < minimumIterations {
		return "", false
	}

	if tracker.config.TargetCost != nil && IsFeasible(bestEvaluation.ConstraintViolation) &&
		bestEvaluation.Cost <= *tracker.config.TargetCost {
		return TerminationTargetCost, true
	}

	if tracker.config.StagnationIterations > 0 &&
		tracker.stagnantIterations >= tracker.config.StagnationIterations {
		return TerminationStagnation, true
	}

	return "", false
}

func (tracker *convergenceTracker) significantlyImproved(candidate CandidateEvaluation) bool {
	if !tracker.evaluator.better(candidate, tracker.referenceBest) {
		return false
	}

	if tracker.evaluator.constraints != nil &&
		tracker.evaluator.constraints.Handling == ConstraintHandlingPenalty {
		config := tracker.evaluator.constraints
		referenceScore := PenalizedCost(
			tracker.referenceBest.Cost, tracker.referenceBest.ConstraintViolation,
			config.PenaltyFactor, config.PenaltyMethod,
		)
		candidateScore := PenalizedCost(
			candidate.Cost, candidate.ConstraintViolation,
			config.PenaltyFactor, config.PenaltyMethod,
		)

		return referenceScore-candidateScore > tracker.config.MinImprovement
	}

	referenceFeasible := IsFeasible(tracker.referenceBest.ConstraintViolation)

	candidateFeasible := IsFeasible(candidate.ConstraintViolation)
	if referenceFeasible != candidateFeasible {
		return candidateFeasible
	}

	if candidateFeasible {
		return tracker.referenceBest.Cost-candidate.Cost > tracker.config.MinImprovement
	}

	return tracker.referenceBest.ConstraintViolation-candidate.ConstraintViolation >
		tracker.config.MinImprovement
}

func validateConvergenceConfig(config *ConvergenceConfig, maxIterations int) error {
	if config == nil {
		return nil
	}

	if config.TargetCost != nil &&
		(math.IsNaN(*config.TargetCost) || math.IsInf(*config.TargetCost, 0)) {
		return fmt.Errorf("target cost must be finite, got %v", *config.TargetCost)
	}

	if math.IsNaN(config.MinImprovement) || math.IsInf(config.MinImprovement, 0) ||
		config.MinImprovement < 0 {
		return fmt.Errorf("minimum improvement must be finite and non-negative, got %v",
			config.MinImprovement)
	}

	if config.StagnationIterations < 0 {
		return fmt.Errorf("stagnation iterations must be non-negative, got %d",
			config.StagnationIterations)
	}

	if config.MinIterations < 0 || config.MinIterations > maxIterations {
		return fmt.Errorf("minimum iterations must be in [0, %d], got %d",
			maxIterations, config.MinIterations)
	}

	return nil
}
