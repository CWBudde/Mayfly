package mayfly

import (
	"math"
	"math/rand"
	"reflect"
	"testing"
)

func TestHMMACompatibilityScheduleProbability(t *testing.T) {
	const theta = 0.99
	for _, testCase := range []struct {
		iteration int
		maximum   int
	}{
		{1, 1000},
		{500, 1000},
		{1000, 1000},
	} {
		want := -math.Exp(-float64(testCase.iteration)/float64(testCase.maximum)) + theta
		want = min(max(want, 0), 1)

		got := hmmaScheduleProbability(testCase.iteration, testCase.maximum, theta)
		if math.Abs(got-want) > 1e-15 {
			t.Errorf("Ps(%d,%d) = %v, want %v",
				testCase.iteration, testCase.maximum, got, want)
		}
	}
}

func TestHMMAOppositionTargetMatchesEquations6And7(t *testing.T) {
	const (
		seed = int64(17)
		lb   = -5.0
		ub   = 5.0
		a4   = 1.5
	)

	best := []float64{-2, 0.5, 4}
	got := hmmaOppositionTarget(best, lb, ub, a4, rand.New(rand.NewSource(seed)))

	reference := rand.New(rand.NewSource(seed))

	want := make([]float64, len(best))
	for i, coordinate := range best {
		r3 := reference.Float64()
		opposition := ub + r3*(lb-coordinate)
		want[i] = max(lb, min(ub, a4*(coordinate-opposition)))
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("opposition target = %v, want %v", got, want)
	}
}

func TestHMMACauchyTargetMatchesEquation8(t *testing.T) {
	const seed = int64(23)

	best := []float64{-2, 0.5, 4}
	got := hmmaCauchyTarget(best, -5, 5, rand.New(rand.NewSource(seed)))

	reference := rand.New(rand.NewSource(seed))

	want := make([]float64, len(best))
	for i, coordinate := range best {
		u := reference.Float64()
		cauchy := math.Tan(math.Pi * (u - 0.5))
		want[i] = max(-5.0, min(5.0, cauchy*coordinate))
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Cauchy target = %v, want %v", got, want)
	}
}

func TestHMMAArtificialMutationMatchesEquation12(t *testing.T) {
	male := []float64{-4, 2, 1}
	female := []float64{2, -2, 5}
	gotMale, gotFemale := hmmaArtificialMutation(male, female, 0.25)
	wantMale := []float64{-2.5, 1, 2}

	wantFemale := []float64{0.5, -1, 4}
	if !reflect.DeepEqual(gotMale, wantMale) || !reflect.DeepEqual(gotFemale, wantFemale) {
		t.Fatalf("artificial mutation = (%v,%v), want (%v,%v)",
			gotMale, gotFemale, wantMale, wantFemale)
	}

	if !reflect.DeepEqual(male, []float64{-4, 2, 1}) ||
		!reflect.DeepEqual(female, []float64{2, -2, 5}) {
		t.Fatal("artificial mutation modified an input before calculating its sibling")
	}
}

func TestHMMALifecycleUsesOneGlobalMutationAndNoGaussianMutants(t *testing.T) {
	for _, parallel := range []bool{false, true} {
		config := NewHMMAConfig()
		config.ObjectiveFunc = sphere
		config.ProblemSize = 2
		config.LowerBound = -2
		config.UpperBound = 2
		config.MaxIterations = 3
		config.NPop = 2
		config.NPopF = 2
		config.NC = 2
		config.NM = 100 // inert: HMMA uses Eq. (12), not Gaussian mutants
		config.Rand = rand.New(rand.NewSource(31))
		config.EnableParallel = parallel

		result, err := Optimize(config)
		if err != nil {
			t.Fatalf("parallel=%v: Optimize: %v", parallel, err)
		}

		// 4 initial + 3*(4 movement + 1 global mutation + 2 offspring).
		const wantEvaluations = 25
		if result.FuncEvalCount != wantEvaluations {
			t.Errorf("parallel=%v: FuncEvalCount = %d, want %d",
				parallel, result.FuncEvalCount, wantEvaluations)
		}
	}
}

func TestHMMAConfigValidation(t *testing.T) {
	for _, mutate := range []func(*Config){
		func(config *Config) { config.HMMAInformationExchange = 0 },
		func(config *Config) { config.HMMAScheduleOffset = math.NaN() },
		func(config *Config) { config.HMMAArtificialMutation = 1.1 },
	} {
		config := NewHMMAConfig()
		config.ProblemSize = 2
		config.LowerBound = -1
		config.UpperBound = 1
		mutate(config)

		err := ValidateConfig(config)
		if err == nil {
			t.Fatalf("ValidateConfig accepted invalid HMMA config: %+v", config)
		}
	}
}
