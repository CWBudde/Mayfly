package main

import (
	"context"
	"math"
	"testing"

	"github.com/cwbudde/mayfly"
)

func TestTuneAndValidateUsesPairedBudgetsAndHeldOutSeeds(t *testing.T) {
	selected, tuning, validation, err := tuneAndValidate(context.Background())
	if err != nil {
		t.Fatalf("tuneAndValidate() error = %v", err)
	}

	if selected != 12 && selected != 20 && selected != 32 {
		t.Fatalf("selected population = %d, want a candidate", selected)
	}

	if tuning.BaseSeed != tuningSeed || validation.BaseSeed != validationSeed {
		t.Fatalf("base seeds = (%d, %d), want (%d, %d)",
			tuning.BaseSeed,
			validation.BaseSeed,
			tuningSeed,
			validationSeed,
		)
	}

	for _, comparison := range []*struct {
		name       string
		runResults int
	}{
		{name: "tuning", runResults: len(tuning.RunResults)},
		{name: "validation", runResults: len(validation.RunResults)},
	} {
		if comparison.runResults == 0 {
			t.Errorf("%s comparison has no candidates", comparison.name)
		}
	}

	for _, result := range []*mayfly.ComparisonResult{tuning, validation} {
		for variantIndex, runs := range result.RunResults {
			if len(runs) != comparisonRuns {
				t.Errorf("%s run count = %d, want %d", result.AlgorithmNames[variantIndex], len(runs), comparisonRuns)
			}

			for runIndex, run := range runs {
				wantSeed := result.BaseSeed + int64(runIndex)
				if run.Seed != wantSeed {
					t.Errorf("%s run %d seed = %d, want paired seed %d",
						result.AlgorithmNames[variantIndex], runIndex, run.Seed, wantSeed)
				}

				if run.FuncEvals != evaluationBudget {
					t.Errorf("%s run %d evaluations = %d, want %d",
						result.AlgorithmNames[variantIndex], runIndex, run.FuncEvals, evaluationBudget)
				}

				if run.Error != "" || math.IsNaN(run.BestCost) || math.IsInf(run.BestCost, 0) {
					t.Errorf("%s run %d is invalid: %+v", result.AlgorithmNames[variantIndex], runIndex, run)
				}
			}
		}
	}
}
