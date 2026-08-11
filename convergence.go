package mayfly

import (
	"fmt"
	"math"
)

type convergenceTracker struct {
	config             *ConvergenceConfig
	referenceBest      float64
	stagnantIterations int
}

func newConvergenceTracker(config *ConvergenceConfig, initialBest float64) *convergenceTracker {
	return &convergenceTracker{
		config:        config,
		referenceBest: initialBest,
	}
}

func (tracker *convergenceTracker) observe(iteration int, bestCost float64) (TerminationReason, bool) {
	if tracker.config == nil {
		return "", false
	}

	improvement := tracker.referenceBest - bestCost
	if improvement > tracker.config.MinImprovement {
		tracker.referenceBest = bestCost
		tracker.stagnantIterations = 0
	} else {
		tracker.stagnantIterations++
	}

	minimumIterations := max(tracker.config.MinIterations, 1)
	if iteration < minimumIterations {
		return "", false
	}

	if tracker.config.TargetCost != nil && bestCost <= *tracker.config.TargetCost {
		return TerminationTargetCost, true
	}

	if tracker.config.StagnationIterations > 0 &&
		tracker.stagnantIterations >= tracker.config.StagnationIterations {
		return TerminationStagnation, true
	}

	return "", false
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
