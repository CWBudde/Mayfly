package mayfly

import (
	"math/rand"
	"reflect"
	"testing"
)

// TestOrthogonalLearningZeroFactorSpendsNoEvaluations pins that a disabled
// orthogonal learning stage is free. With a zero factor every L4 candidate
// collapses onto the parent male, so evaluating them burns budget on a
// guaranteed no-op.
func TestOrthogonalLearningZeroFactorSpendsNoEvaluations(t *testing.T) {
	testCases := []struct {
		name      string
		factor    float64
		wantCalls int
	}{
		{name: "disabled", factor: 0, wantCalls: 0},
		{name: "enabled", factor: 0.3, wantCalls: len(OrthogonalArray(3)) + 1},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			calls := 0
			objective := func(position []float64) float64 {
				calls++

				return Sphere(position)
			}

			male := newMayfly(3)
			male.Position = []float64{1, 2, 3}
			male.Cost = Sphere(male.Position)
			male.Best.Position = []float64{1, 2, 3}
			male.Best.Cost = male.Cost

			lb := []float64{-5, -5, -5}
			ub := []float64{5, 5, 5}

			result := ApplyOrthogonalLearning(
				male, male.Best.Position, []float64{0, 0, 0}, testCase.factor,
				lb, ub, objective, rand.New(rand.NewSource(7)),
			)

			if calls != testCase.wantCalls {
				t.Errorf("objective calls = %d, want %d", calls, testCase.wantCalls)
			}

			if testCase.factor == 0 && result != male {
				t.Error("disabled orthogonal learning replaced the male")
			}
		})
	}
}

func TestOrthogonalLearningToEliteZeroFactorSpendsNoEvaluations(t *testing.T) {
	calls := 0
	objective := func(position []float64) float64 {
		calls++

		return Sphere(position)
	}

	males := make([]*Mayfly, 4)
	for i := range males {
		male := newMayfly(2)
		male.Position = []float64{float64(i), float64(i) + 1}
		male.Cost = Sphere(male.Position)
		male.Best.Position = []float64{float64(i), float64(i) + 1}
		male.Best.Cost = male.Cost
		males[i] = male
	}

	ApplyOrthogonalLearningToElite(
		males, 0.5, []float64{0, 0}, 0,
		[]float64{-5, -5}, []float64{5, 5},
		objective, rand.New(rand.NewSource(11)),
	)

	if calls != 0 {
		t.Errorf("objective calls = %d, want 0", calls)
	}
}

// TestOptimizeOrthogonalFactorZeroDoesNotGrowBudget checks the same property
// end to end, on both the sequential and the pooled evaluation path.
func TestOptimizeOrthogonalFactorZeroDoesNotGrowBudget(t *testing.T) {
	newConfig := func(parallel bool, factor float64) *Config {
		config := NewOLCEConfig()
		config.ObjectiveFunc = Sphere
		config.ProblemSize = 4
		config.LowerBound = -5
		config.UpperBound = 5
		config.MaxIterations = 20
		config.NPop = 10
		config.NPopF = 10
		config.OrthogonalFactor = factor
		config.EnableParallel = parallel
		config.MaxWorkers = 2
		config.Rand = rand.New(rand.NewSource(5))

		return config
	}

	for _, parallel := range []bool{false, true} {
		name := "sequential"
		if parallel {
			name = "parallel"
		}

		t.Run(name, func(t *testing.T) {
			disabled, err := Optimize(newConfig(parallel, 0))
			if err != nil {
				t.Fatalf("Optimize with disabled orthogonal learning: %v", err)
			}

			enabled, err := Optimize(newConfig(parallel, 0.3))
			if err != nil {
				t.Fatalf("Optimize with enabled orthogonal learning: %v", err)
			}

			if disabled.FuncEvalCount >= enabled.FuncEvalCount {
				t.Errorf("FuncEvalCount with factor 0 = %d, want fewer than %d",
					disabled.FuncEvalCount, enabled.FuncEvalCount)
			}

			// OL belongs to the primary male movement operator and therefore
			// applies to every male, rather than a post-selection elite subset.
			wantSaved := 20 * 10 * (len(OrthogonalArray(4)) + 1)
			if saved := enabled.FuncEvalCount - disabled.FuncEvalCount; saved != wantSaved {
				t.Errorf("saved evaluations = %d, want %d", saved, wantSaved)
			}
		})
	}
}

func TestOrthogonalArrayHasDistinctBalancedColumns(t *testing.T) {
	for _, dimensions := range []int{1, 3, 4, 7, 12} {
		array := OrthogonalArray(dimensions)
		if len(array) == 0 {
			t.Fatalf("OrthogonalArray(%d) returned no rows", dimensions)
		}

		for left := range dimensions {
			ones := 0
			for row := range array {
				ones += array[row][left]
			}

			if ones*2 != len(array) {
				t.Errorf("D=%d column %d has %d ones in %d rows", dimensions, left, ones, len(array))
			}

			for right := left + 1; right < dimensions; right++ {
				counts := [2][2]int{}
				for row := range array {
					counts[array[row][left]][array[row][right]]++
				}

				want := len(array) / 4
				if counts != [2][2]int{{want, want}, {want, want}} {
					t.Errorf("D=%d columns %d/%d are not pairwise balanced: %v",
						dimensions, left, right, counts)
				}
			}
		}
	}
}

func TestOrthogonalArrayReturnsDefensiveCopy(t *testing.T) {
	first := OrthogonalArray(4)
	first[0][0] = 99

	second := OrthogonalArray(4)
	if reflect.DeepEqual(first, second) || second[0][0] != 0 {
		t.Fatalf("OrthogonalArray shares mutable state: first=%v second=%v", first, second)
	}
}

func TestOrthogonalLearningIsDimensionSafeBeyondL4(t *testing.T) {
	const dimensions = 12

	male := newMayfly(dimensions)
	pbest := make([]float64, dimensions)
	gbest := make([]float64, dimensions)
	lower := make([]float64, dimensions)

	upper := make([]float64, dimensions)
	for dimension := range dimensions {
		male.Position[dimension] = 1
		male.Best.Position[dimension] = 1
		pbest[dimension] = 0.5
		gbest[dimension] = -0.5
		lower[dimension] = -2
		upper[dimension] = 2
	}

	male.Cost = Sphere(male.Position)
	male.Best.Cost = male.Cost

	calls := 0

	result := ApplyOrthogonalLearning(
		male, pbest, gbest, 0.3, lower, upper,
		func(position []float64) float64 {
			calls++
			return Sphere(position)
		},
		rand.New(rand.NewSource(3)),
	)
	if result == nil {
		t.Fatal("orthogonal learning returned nil")
	}

	if calls != len(OrthogonalArray(dimensions))+1 {
		t.Fatalf("evaluations = %d, want %d", calls, len(OrthogonalArray(dimensions))+1)
	}
}

func TestOptimizeRejectsOLCEDimensionAboveArrayCap(t *testing.T) {
	config := NewOLCEConfig()
	config.ObjectiveFunc = Sphere
	config.ProblemSize = MaxOrthogonalArrayDimensions + 1
	config.LowerBound = -1
	config.UpperBound = 1
	config.MaxIterations = 1

	if _, err := Optimize(config); err == nil {
		t.Fatal("Optimize accepted an OLCE dimension above the orthogonal-array cap")
	}
}
