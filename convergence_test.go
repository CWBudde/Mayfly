package mayfly

import (
	"math"
	"strings"
	"testing"
)

func TestConvergenceTrackerRequiresSignificantImprovement(t *testing.T) {
	tracker := newConvergenceTracker(&ConvergenceConfig{
		MinImprovement:       0.5,
		StagnationIterations: 2,
	}, Best{Cost: 10})

	if reason, stop := tracker.observe(1, Best{Cost: 9.8}); stop {
		t.Fatalf("small improvement stopped run with reason %q", reason)
	}

	if reason, stop := tracker.observe(2, Best{Cost: 9.4}); stop {
		t.Fatalf("cumulative significant improvement stopped run with reason %q", reason)
	}

	if reason, stop := tracker.observe(3, Best{Cost: 9.1}); stop {
		t.Fatalf("first stagnant iteration stopped run with reason %q", reason)
	}

	reason, stop := tracker.observe(4, Best{Cost: 9.0})
	if !stop || reason != TerminationStagnation {
		t.Fatalf("second stagnant iteration = (%q, %t), want (%q, true)",
			reason, stop, TerminationStagnation)
	}
}

func TestValidateConvergenceConfig(t *testing.T) {
	nan := math.NaN()

	tests := []struct {
		name   string
		config *ConvergenceConfig
	}{
		{name: "non-finite target", config: &ConvergenceConfig{TargetCost: &nan}},
		{name: "negative improvement", config: &ConvergenceConfig{MinImprovement: -1}},
		{name: "non-finite improvement", config: &ConvergenceConfig{MinImprovement: math.Inf(1)}},
		{name: "negative stagnation", config: &ConvergenceConfig{StagnationIterations: -1}},
		{name: "negative minimum", config: &ConvergenceConfig{MinIterations: -1}},
		{name: "minimum above maximum", config: &ConvergenceConfig{MinIterations: 11}},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := validateConvergenceConfig(testCase.config, 10)
			if err == nil {
				t.Fatal("validateConvergenceConfig accepted invalid configuration")
			}
		})
	}

	err := validateConvergenceConfig(nil, 10)
	if err != nil {
		t.Fatalf("nil convergence config: %v", err)
	}

	zero := 0.0

	valid := &ConvergenceConfig{
		TargetCost:           &zero,
		MinImprovement:       1e-9,
		MinIterations:        2,
		StagnationIterations: 4,
	}

	err = validateConvergenceConfig(valid, 10)
	if err != nil {
		t.Fatalf("valid convergence config: %v", err)
	}
}

func TestPublicValidationRejectsInvalidConvergenceConfig(t *testing.T) {
	validators := []struct {
		name     string
		validate func(*Config) error
	}{
		{name: "Optimize", validate: func(config *Config) error {
			_, err := Optimize(config)

			return err
		}},
		{name: "ValidateConfig", validate: ValidateConfig},
	}

	for _, validator := range validators {
		t.Run(validator.name, func(t *testing.T) {
			config := lifecycleConfig(sphere)
			config.Convergence = &ConvergenceConfig{MinIterations: config.MaxIterations + 1}

			err := validator.validate(config)
			if err == nil {
				t.Fatal("validation accepted invalid convergence configuration")
			}

			if !strings.Contains(err.Error(), "invalid convergence config") {
				t.Fatalf("error %q does not identify convergence configuration", err)
			}
		})
	}
}
