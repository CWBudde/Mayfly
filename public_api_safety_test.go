package mayfly

import (
	"bytes"
	"math"
	"math/rand"
	"strings"
	"testing"
)

func TestCheckedStochasticOperatorsRejectInvalidInputs(t *testing.T) {
	rng := rand.New(rand.NewSource(1))

	tests := []struct {
		name string
		run  func() error
	}{
		{"crossover nil RNG", func() error {
			_, _, err := CrossoverChecked([]float64{0}, []float64{1}, -1, 1, nil)
			return err
		}},
		{"crossover dimension mismatch", func() error {
			_, _, err := CrossoverChecked([]float64{0}, []float64{0, 1}, -1, 1, rng)
			return err
		}},
		{"crossover invalid gamma", func() error {
			_, _, err := CrossoverBlendChecked([]float64{0}, []float64{1}, math.NaN(), -1, 1, rng)
			return err
		}},
		{"Gaussian mutation nil RNG", func() error {
			_, err := MutateGaussianChecked([]float64{0}, 0.5, -1, 1, nil)
			return err
		}},
		{"Cauchy mutation invalid rate", func() error {
			_, err := MutateCauchyChecked([]float64{0}, 2, -1, 1, rng)
			return err
		}},
		{"hybrid mutation invalid probability", func() error {
			_, err := HybridMutateChecked([]float64{0}, 0.5, -1, 1, math.Inf(1), rng)
			return err
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.run()
			if err == nil {
				t.Fatal("checked operator accepted invalid input")
			}
		})
	}
}

func TestCheckedConstructorsRejectInsteadOfCoercing(t *testing.T) {
	if _, err := NewAnnealingSchedulerChecked(math.NaN(), 0.95, CoolingExponential); err == nil {
		t.Fatal("checked annealing constructor accepted NaN")
	}

	if _, err := NewAnnealingSchedulerChecked(10, 0.95, "unknown"); err == nil {
		t.Fatal("checked annealing constructor accepted unknown schedule")
	}

	if _, err := NewLogisticMapChecked(math.Inf(1)); err == nil {
		t.Fatal("checked logistic-map constructor accepted infinity")
	}

	if _, err := NewParetoArchiveChecked(0); err == nil {
		t.Fatal("checked Pareto archive constructor accepted zero capacity")
	}

	if _, err := NewVariantChecked("missing"); err == nil {
		t.Fatal("checked variant lookup accepted unknown name")
	}

	if _, err := NewBuilderChecked("missing"); err == nil {
		t.Fatal("checked builder lookup accepted unknown name")
	}

	if _, err := NewBuilderFromVariantChecked(nil); err == nil {
		t.Fatal("checked builder accepted nil variant")
	}

	if _, err := NewComparisonRunner().WithVariantNamesChecked("ma", "missing"); err == nil {
		t.Fatal("checked comparison runner accepted unknown variant")
	}

	if _, err := NewComparisonRunner().WithVariantsChecked(nil); err == nil {
		t.Fatal("checked comparison runner accepted nil variant")
	}
}

func TestAnnealingAndLogisticMapValidateMutableState(t *testing.T) {
	scheduler, err := NewAnnealingSchedulerChecked(10, 0.95, CoolingExponential)
	if err != nil {
		t.Fatalf("NewAnnealingSchedulerChecked: %v", err)
	}

	scheduler.CurrentTemperature = math.NaN()
	if err := scheduler.Validate(); err == nil {
		t.Fatal("scheduler validation accepted mutated NaN state")
	}

	if err := scheduler.UpdateChecked(); err == nil {
		t.Fatal("checked update accepted mutated NaN state")
	}

	if err := scheduler.ResetChecked(); err != nil {
		t.Fatalf("checked reset could not repair current state: %v", err)
	}

	logisticMap, err := NewLogisticMapChecked(0.25)
	if err != nil {
		t.Fatalf("NewLogisticMapChecked: %v", err)
	}

	if err := logisticMap.ResetChecked(1); err == nil {
		t.Fatal("ResetChecked accepted boundary seed")
	}

	if got := logisticMap.Current(); got != 0.25 {
		t.Fatalf("failed reset changed state to %v", got)
	}

	if _, err := (*LogisticMap)(nil).NextChecked(); err == nil {
		t.Fatal("checked next accepted nil map")
	}
}

func TestCheckedOrthogonalHelpersRejectShapeAndObjectiveFailures(t *testing.T) {
	if _, err := OrthogonalArrayChecked(0); err == nil {
		t.Fatal("checked orthogonal array accepted zero dimensions")
	}

	if _, err := OrthogonalArrayChecked(MaxOrthogonalArrayDimensions + 1); err == nil {
		t.Fatal("checked orthogonal array accepted excessive dimensions")
	}

	male := newMayfly(2)
	male.Position = []float64{0, 0}
	male.Best.Position = []float64{0, 0}
	male.Cost = 0
	male.Best.Cost = 0

	_, err := ApplyOrthogonalLearningChecked(
		male, []float64{0}, []float64{0, 0}, 0.3,
		[]float64{-1, -1}, []float64{1, 1}, Sphere, nil,
	)
	if err == nil || !strings.Contains(err.Error(), "dimension") {
		t.Fatalf("dimension mismatch error = %v", err)
	}

	_, err = ApplyOrthogonalLearningChecked(
		male, []float64{0, 0}, []float64{0, 0}, 0.3,
		[]float64{-1, -1}, []float64{1, 1}, func([]float64) float64 { return math.NaN() }, nil,
	)
	if err == nil {
		t.Fatal("checked orthogonal learning accepted non-finite objective")
	}
}

func TestParetoArchiveSnapshotsAreDefensive(t *testing.T) {
	archive, err := NewParetoArchiveWithObjectives(2, 2)
	if err != nil {
		t.Fatalf("NewParetoArchiveWithObjectives: %v", err)
	}

	solution := &ParetoSolution{Position: []float64{1}, ObjectiveValues: []float64{1, 2}}

	added, err := archive.Add(solution)
	if err != nil || !added {
		t.Fatalf("Add = %v, %v", added, err)
	}

	solution.ObjectiveValues[0] = -100
	snapshot := archive.GetSolutions()

	snapshot[0].ObjectiveValues[0] = -200
	if got := archive.GetBestSolution().ObjectiveValues[0]; got != 1 {
		t.Fatalf("caller mutation leaked into archive: %v", got)
	}
}

func TestCheckedConstraintAndRecommendationHelpers(t *testing.T) {
	_, err := EvaluateConstraintsChecked([]float64{0}, &ConstraintConfig{
		Inequalities: []ConstraintFunction{func([]float64) float64 { panic("boom") }},
	})
	if err == nil || !strings.Contains(err.Error(), "panicked") {
		t.Fatalf("constraint panic error = %v", err)
	}

	if _, err := PenalizedCostChecked(1, -1, 10, PenaltyLinear); err == nil {
		t.Fatal("checked penalty accepted negative violation")
	}

	if _, err := BetterConstrainedCandidateChecked(
		CandidateEvaluation{Cost: math.NaN()}, CandidateEvaluation{Cost: 1}, nil,
	); err == nil {
		t.Fatal("checked comparison accepted NaN cost")
	}

	if err := AutoTuneConfigChecked(nil, ProblemCharacteristics{}); err == nil {
		t.Fatal("checked auto-tuner accepted nil config")
	}

	if _, err := NewAlgorithmSelector().RecommendBestChecked(ProblemCharacteristics{MultiObjective: true}); err == nil {
		t.Fatal("checked selector claimed multi-objective support")
	}

	if err := WriteRecommendations(&bytes.Buffer{}, []AlgorithmRecommendation{{}}); err == nil {
		t.Fatal("recommendation writer accepted nil variant")
	}

	if _, err := RecommendForBenchmarkChecked("unknown"); err == nil {
		t.Fatal("checked benchmark lookup accepted unknown name")
	}

	if err := SaveConfigToFile(nil, t.TempDir()+"/config.json"); err == nil {
		t.Fatal("config saver accepted nil config")
	}
}
