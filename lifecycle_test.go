package mayfly

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

func lifecycleConfig(objective ObjectiveFunction) *Config {
	config := NewDefaultConfig()
	config.ObjectiveFunc = objective
	config.ProblemSize = 2
	config.LowerBound = -1
	config.UpperBound = 1
	config.MaxIterations = 3
	config.NPop = 4
	config.NPopF = 4
	config.NC = 0
	config.NM = 0
	config.Rand = rand.New(rand.NewSource(7))

	return config
}

func TestOptimizeRemainsCompatibleWithOptimizeContext(t *testing.T) {
	legacyConfig := lifecycleConfig(sphere)
	contextConfig := lifecycleConfig(sphere)

	legacy, err := Optimize(legacyConfig)
	if err != nil {
		t.Fatalf("Optimize: %v", err)
	}

	withContext, err := OptimizeContext(context.Background(), contextConfig)
	if err != nil {
		t.Fatalf("OptimizeContext: %v", err)
	}

	if !reflect.DeepEqual(legacy.GlobalBest, withContext.GlobalBest) {
		t.Errorf("GlobalBest differs: Optimize=%+v OptimizeContext=%+v",
			legacy.GlobalBest, withContext.GlobalBest)
	}

	if !reflect.DeepEqual(legacy.ConvergenceCurve, withContext.ConvergenceCurve) {
		t.Errorf("ConvergenceCurve differs: Optimize=%v OptimizeContext=%v",
			legacy.ConvergenceCurve, withContext.ConvergenceCurve)
	}

	if legacy.FuncEvalCount != withContext.FuncEvalCount {
		t.Errorf("FuncEvalCount differs: Optimize=%d OptimizeContext=%d",
			legacy.FuncEvalCount, withContext.FuncEvalCount)
	}
}

func TestWithInitialPopulationSeedsAndRandomFillsPopulations(t *testing.T) {
	var evaluated [][]float64
	objective := func(position []float64) float64 {
		evaluated = append(evaluated, append([]float64(nil), position...))
		return sphere(position)
	}

	males := [][]float64{{0.25, -0.25}, {0, 0}}
	females := [][]float64{{-0.75, 0.5}}
	option := WithInitialPopulation(males, females)

	// WithInitialPopulation promises a snapshot at option construction time.
	males[0][0] = 99
	females[0][0] = 99

	config := lifecycleConfig(objective)
	config.MaxIterations = 1
	result, err := OptimizeContext(context.Background(), config, option)
	if err != nil {
		t.Fatalf("OptimizeContext: %v", err)
	}

	if len(evaluated) != 16 {
		t.Fatalf("objective called %d times, want 8 initialization and 8 iteration evaluations", len(evaluated))
	}

	for _, testCase := range []struct {
		index int
		want  []float64
	}{
		{0, []float64{0.25, -0.25}},
		{1, []float64{0, 0}},
		{4, []float64{-0.75, 0.5}},
	} {
		if !reflect.DeepEqual(evaluated[testCase.index], testCase.want) {
			t.Errorf("initial evaluation %d = %v, want %v",
				testCase.index, evaluated[testCase.index], testCase.want)
		}
	}

	for i, position := range evaluated[:8] {
		for j, value := range position {
			if value < config.LowerBound || value > config.UpperBound {
				t.Errorf("initial evaluation %d dimension %d out of bounds: %v", i, j, value)
			}
		}
	}

	if result.GlobalBest.Cost != 0 {
		t.Errorf("GlobalBest.Cost = %v, want seeded optimum 0", result.GlobalBest.Cost)
	}
}

func TestWithInitialPopulationValidation(t *testing.T) {
	tests := []struct {
		name    string
		males   [][]float64
		females [][]float64
		wantErr string
	}{
		{
			name:    "too many males",
			males:   make([][]float64, 5),
			wantErr: "exceeds NPop",
		},
		{
			name:    "too many females",
			females: make([][]float64, 5),
			wantErr: "exceeds NPopF",
		},
		{
			name:    "wrong male dimension",
			males:   [][]float64{{0}},
			wantErr: "dimension 1, want 2",
		},
		{
			name:    "wrong female dimension",
			females: [][]float64{{0, 0, 0}},
			wantErr: "dimension 3, want 2",
		},
		{
			name:    "male NaN",
			males:   [][]float64{{math.NaN(), 0}},
			wantErr: "must be finite",
		},
		{
			name:    "female infinity",
			females: [][]float64{{0, math.Inf(1)}},
			wantErr: "must be finite",
		},
		{
			name:    "below lower bound",
			males:   [][]float64{{-1.01, 0}},
			wantErr: "outside bounds",
		},
		{
			name:    "above upper bound",
			females: [][]float64{{0, 1.01}},
			wantErr: "outside bounds",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			config := lifecycleConfig(sphere)
			_, err := OptimizeContext(
				context.Background(),
				config,
				WithInitialPopulation(testCase.males, testCase.females),
			)
			if err == nil {
				t.Fatal("OptimizeContext accepted invalid initial population")
			}

			if !strings.Contains(err.Error(), testCase.wantErr) {
				t.Errorf("error %q does not contain %q", err, testCase.wantErr)
			}
		})
	}
}

func TestProgressObserverReceivesIndependentSnapshots(t *testing.T) {
	var progress []Progress
	observer := func(update Progress) {
		progress = append(progress, update)
		update.Best.Position[0] = 999
	}

	config := lifecycleConfig(sphere)
	result, err := OptimizeContext(
		context.Background(),
		config,
		WithProgressObserver(observer),
	)
	if err != nil {
		t.Fatalf("OptimizeContext: %v", err)
	}

	if len(progress) != config.MaxIterations {
		t.Fatalf("observer called %d times, want %d", len(progress), config.MaxIterations)
	}

	for i, update := range progress {
		wantIteration := i + 1
		wantEvaluations := 8 + wantIteration*8
		if update.Iteration != wantIteration {
			t.Errorf("progress[%d].Iteration = %d, want %d", i, update.Iteration, wantIteration)
		}

		if update.EvaluationCount != wantEvaluations {
			t.Errorf("progress[%d].EvaluationCount = %d, want %d",
				i, update.EvaluationCount, wantEvaluations)
		}
	}

	if result.GlobalBest.Position[0] == 999 {
		t.Fatal("observer mutation changed result GlobalBest.Position")
	}

	if progress[len(progress)-1].Best.Cost != result.GlobalBest.Cost {
		t.Errorf("last progress cost %v != result cost %v",
			progress[len(progress)-1].Best.Cost, result.GlobalBest.Cost)
	}
}

func TestOptimizeStopsAtTargetCost(t *testing.T) {
	target := 1.0
	config := lifecycleConfig(func([]float64) float64 { return target })
	config.MaxIterations = 20
	config.Convergence = &ConvergenceConfig{TargetCost: &target}

	var progress []Progress
	result, err := OptimizeContext(
		context.Background(),
		config,
		WithProgressObserver(func(update Progress) { progress = append(progress, update) }),
	)
	if err != nil {
		t.Fatalf("OptimizeContext: %v", err)
	}

	if result.TerminationReason != TerminationTargetCost {
		t.Errorf("TerminationReason = %q, want %q",
			result.TerminationReason, TerminationTargetCost)
	}

	if result.IterationCount != 1 {
		t.Errorf("IterationCount = %d, want 1", result.IterationCount)
	}

	if len(result.ConvergenceCurve) != result.IterationCount {
		t.Errorf("ConvergenceCurve length = %d, want IterationCount=%d",
			len(result.ConvergenceCurve), result.IterationCount)
	}

	if len(progress) != result.IterationCount {
		t.Errorf("observer calls = %d, want IterationCount=%d", len(progress), result.IterationCount)
	}
}

func TestOptimizeStopsAfterStagnationAndMinimumIterations(t *testing.T) {
	config := lifecycleConfig(func([]float64) float64 { return 1 })
	config.MaxIterations = 20
	config.Convergence = &ConvergenceConfig{
		StagnationIterations: 3,
		MinIterations:        5,
	}

	result, err := Optimize(config)
	if err != nil {
		t.Fatalf("Optimize: %v", err)
	}

	if result.TerminationReason != TerminationStagnation {
		t.Errorf("TerminationReason = %q, want %q",
			result.TerminationReason, TerminationStagnation)
	}

	if result.IterationCount != config.Convergence.MinIterations {
		t.Errorf("IterationCount = %d, want minimum %d",
			result.IterationCount, config.Convergence.MinIterations)
	}

	if len(result.ConvergenceCurve) != result.IterationCount {
		t.Errorf("ConvergenceCurve length = %d, want IterationCount=%d",
			len(result.ConvergenceCurve), result.IterationCount)
	}
}

func TestOptimizeWithoutConvergencePolicyUsesMaximumIterations(t *testing.T) {
	config := lifecycleConfig(func([]float64) float64 { return 1 })

	result, err := Optimize(config)
	if err != nil {
		t.Fatalf("Optimize: %v", err)
	}

	if result.TerminationReason != TerminationMaxIterations {
		t.Errorf("TerminationReason = %q, want %q",
			result.TerminationReason, TerminationMaxIterations)
	}

	if result.IterationCount != config.MaxIterations {
		t.Errorf("IterationCount = %d, want MaxIterations=%d",
			result.IterationCount, config.MaxIterations)
	}
}

func TestConvergenceDetectionAppliesToEveryVariant(t *testing.T) {
	target := 1.0
	factories := []struct {
		name string
		new  func() *Config
	}{
		{name: "standard", new: NewDefaultConfig},
		{name: "DESMA", new: NewDESMAConfig},
		{name: "OLCE", new: NewOLCEConfig},
		{name: "EOBBMA", new: NewEOBBMAConfig},
		{name: "GSASMA", new: NewGSASMAConfig},
		{name: "MPMA", new: NewMPMAConfig},
		{name: "AOBLMOA", new: NewAOBLMOAConfig},
	}

	for _, factory := range factories {
		t.Run(factory.name, func(t *testing.T) {
			config := factory.new()
			config.ObjectiveFunc = func([]float64) float64 { return target }
			config.ProblemSize = 2
			config.LowerBound = -1
			config.UpperBound = 1
			config.MaxIterations = 10
			config.NPop = 6
			config.NPopF = 6
			config.NC = 0
			config.NM = 0
			config.Rand = rand.New(rand.NewSource(17))
			config.Convergence = &ConvergenceConfig{TargetCost: &target}

			result, err := Optimize(config)
			if err != nil {
				t.Fatalf("Optimize: %v", err)
			}

			if result.TerminationReason != TerminationTargetCost || result.IterationCount != 1 {
				t.Errorf("termination = (%q, %d), want (%q, 1)",
					result.TerminationReason, result.IterationCount, TerminationTargetCost)
			}
		})
	}
}

func TestOptimizeContextCancellation(t *testing.T) {
	t.Run("already canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		calls := 0
		config := lifecycleConfig(func(position []float64) float64 {
			calls++
			return sphere(position)
		})

		result, err := OptimizeContext(ctx, config)
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("error = %v, want context.Canceled", err)
		}

		if result != nil {
			t.Errorf("result = %+v, want nil", result)
		}

		if calls != 0 {
			t.Errorf("objective called %d times after prior cancellation", calls)
		}
	})

	t.Run("during initialization", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		calls := 0
		config := lifecycleConfig(func(position []float64) float64 {
			calls++
			cancel()
			return sphere(position)
		})

		_, err := OptimizeContext(ctx, config)
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("error = %v, want context.Canceled", err)
		}

		if calls != 1 {
			t.Errorf("objective called %d times, want cancellation before second initialization", calls)
		}
	})

	t.Run("at iteration boundary", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		calls := 0
		progressCalls := 0
		config := lifecycleConfig(func(position []float64) float64 {
			calls++
			if calls == 9 {
				cancel()
			}

			return sphere(position)
		})

		_, err := OptimizeContext(ctx, config, WithProgressObserver(func(Progress) {
			progressCalls++
		}))
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("error = %v, want context.Canceled", err)
		}

		// Initialization uses 8 evaluations. Cancellation during the next
		// iteration is deliberately observed only once that iteration ends.
		if calls != 16 {
			t.Errorf("objective called %d times, want completed iteration's 16", calls)
		}

		if progressCalls != 1 {
			t.Errorf("observer called %d times, want the completed iteration reported", progressCalls)
		}
	})
}

func TestOptimizeContextRejectsInvalidLifecycleInputs(t *testing.T) {
	config := lifecycleConfig(sphere)
	if _, err := OptimizeContext(nil, config); !errors.Is(err, errNilContext) {
		t.Errorf("nil context error = %v, want %v", err, errNilContext)
	}

	config = lifecycleConfig(sphere)
	if _, err := OptimizeContext(context.Background(), config, RunOption{}); err == nil {
		t.Fatal("OptimizeContext accepted zero-value RunOption")
	}
}
