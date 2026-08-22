package mayfly

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"reflect"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

func parallelTestConfig(objective ObjectiveFunction) *Config {
	config := NewDefaultConfig()
	config.ObjectiveFunc = objective
	config.ProblemSize = 3
	config.LowerBound = -1
	config.UpperBound = 1
	config.MaxIterations = 2
	config.NPop = 8
	config.NPopF = 4
	config.NC = 0
	config.NM = 0
	config.Rand = rand.New(rand.NewSource(42))
	config.EnableParallel = true
	config.MaxWorkers = 3

	return config
}

func TestParallelEvaluationHonorsWorkerLimit(t *testing.T) {
	for _, maxWorkers := range []int{1, 3, 100} {
		t.Run(strconv.Itoa(maxWorkers), func(t *testing.T) {
			var active atomic.Int64

			var maximum atomic.Int64

			var calls atomic.Int64

			objective := func(position []float64) float64 {
				calls.Add(1)
				current := active.Add(1)

				defer active.Add(-1)

				for {
					observed := maximum.Load()
					if current <= observed || maximum.CompareAndSwap(observed, current) {
						break
					}
				}

				time.Sleep(time.Millisecond)

				return sphere(position)
			}

			config := parallelTestConfig(objective)
			config.MaxWorkers = maxWorkers

			result, err := Optimize(config)
			if err != nil {
				t.Fatalf("Optimize: %v", err)
			}

			wantCalls := int64((config.NPop + config.NPopF) * (config.MaxIterations + 1))
			if calls.Load() != wantCalls {
				t.Errorf("objective calls = %d, want %d", calls.Load(), wantCalls)
			}

			if result.FuncEvalCount != int(wantCalls) {
				t.Errorf("FuncEvalCount = %d, want %d", result.FuncEvalCount, wantCalls)
			}

			workerLimit := min(maxWorkers, max(config.NPop, config.NPopF))
			if maximum.Load() > int64(workerLimit) {
				t.Errorf("maximum concurrent evaluations = %d, exceeds effective worker limit %d",
					maximum.Load(), workerLimit)
			}

			if maxWorkers == 1 && maximum.Load() != 1 {
				t.Errorf("maximum concurrent evaluations = %d, want 1", maximum.Load())
			}

			if maxWorkers > 1 && maximum.Load() <= 1 {
				t.Errorf("maximum concurrent evaluations = %d, want more than one", maximum.Load())
			}
		})
	}
}

func TestParallelPopulationEvaluationPaths(t *testing.T) {
	variants := []struct {
		configure func(*Config)
		name      string
	}{
		{name: "standard", configure: func(*Config) {}},
		{name: "EOBBMA", configure: func(config *Config) {
			variant := NewEOBBMAConfig()
			config.UseEOBBMA = true
			config.LevyAlpha = variant.LevyAlpha
			config.LevyBeta = variant.LevyBeta
			config.EliteOppositionCount = variant.EliteOppositionCount
			config.OppositionRate = 0
		}},
	}

	for _, variant := range variants {
		t.Run(variant.name, func(t *testing.T) {
			config := parallelTestConfig(sphere)
			variant.configure(config)

			initialMales := [][]float64{{0, 0, 0}}

			result, err := OptimizeContext(
				context.Background(),
				config,
				WithInitialPopulation(initialMales, nil),
			)
			if err != nil {
				t.Fatalf("OptimizeContext: %v", err)
			}

			if result.GlobalBest.Cost != 0 {
				t.Errorf("GlobalBest.Cost = %v, want 0", result.GlobalBest.Cost)
			}

			wantEvaluations := (config.NPop + config.NPopF) * (config.MaxIterations + 1)
			if result.FuncEvalCount != wantEvaluations {
				t.Errorf("FuncEvalCount = %d, want %d", result.FuncEvalCount, wantEvaluations)
			}
		})
	}
}

func TestParallelGeneticEvaluationUsesOffspringBatchCapacity(t *testing.T) {
	testCases := []struct {
		name               string
		mutants            int
		maxWorkers         int
		wantMaxConcurrency int64
		wantEvaluations    int
	}{
		{
			name:               "crossover",
			mutants:            0,
			maxWorkers:         2,
			wantMaxConcurrency: 2,
			wantEvaluations:    6,
		},
		{
			name:               "mutation",
			mutants:            4,
			maxWorkers:         4,
			wantMaxConcurrency: 4,
			wantEvaluations:    10,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var active atomic.Int64

			var maximum atomic.Int64

			var calls atomic.Int64

			objective := func(position []float64) float64 {
				calls.Add(1)
				current := active.Add(1)

				defer active.Add(-1)

				for {
					observed := maximum.Load()
					if current <= observed || maximum.CompareAndSwap(observed, current) {
						break
					}
				}

				time.Sleep(5 * time.Millisecond)

				return sphere(position)
			}

			config := NewDefaultConfig()
			config.ObjectiveFunc = objective
			config.ProblemSize = 2
			config.LowerBound = -1
			config.UpperBound = 1
			config.MaxIterations = 1
			config.NPop = 1
			config.NPopF = 1
			config.NC = 2
			config.NM = testCase.mutants
			config.Rand = rand.New(rand.NewSource(42))
			config.EnableParallel = true
			config.MaxWorkers = testCase.maxWorkers

			result, err := Optimize(config)
			if err != nil {
				t.Fatalf("Optimize: %v", err)
			}

			if got := maximum.Load(); got != testCase.wantMaxConcurrency {
				t.Errorf("maximum concurrent evaluations = %d, want %d", got, testCase.wantMaxConcurrency)
			}

			if got := int(calls.Load()); got != testCase.wantEvaluations {
				t.Errorf("objective calls = %d, want %d", got, testCase.wantEvaluations)
			}

			if result.FuncEvalCount != testCase.wantEvaluations {
				t.Errorf("FuncEvalCount = %d, want %d", result.FuncEvalCount, testCase.wantEvaluations)
			}
		})
	}
}

func TestParallelGeneticEvaluationIsSchedulingIndependent(t *testing.T) {
	testCases := []struct {
		newConfig func() *Config
		name      string
	}{
		{name: "standard", newConfig: NewDefaultConfig},
		{name: "OLCE", newConfig: NewOLCEConfig},
		{name: "GSASMA", newConfig: NewGSASMAConfig},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			run := func(maxWorkers int) *Result {
				config := testCase.newConfig()
				config.ObjectiveFunc = sphere
				config.ProblemSize = 3
				config.LowerBound = -2
				config.UpperBound = 2
				config.MaxIterations = 3
				config.NPop = 4
				config.NPopF = 4
				config.NC = 8
				config.NM = 4
				config.Rand = rand.New(rand.NewSource(1234))
				config.EnableParallel = true
				config.MaxWorkers = maxWorkers

				result, err := Optimize(config)
				if err != nil {
					t.Fatalf("Optimize with MaxWorkers=%d: %v", maxWorkers, err)
				}

				return result
			}

			oneWorker := run(1)
			fourWorkers := run(4)

			if !reflect.DeepEqual(oneWorker.GlobalBest, fourWorkers.GlobalBest) {
				t.Errorf("GlobalBest differs by worker count: one=%+v four=%+v",
					oneWorker.GlobalBest, fourWorkers.GlobalBest)
			}

			if !reflect.DeepEqual(oneWorker.ConvergenceCurve, fourWorkers.ConvergenceCurve) {
				t.Errorf("ConvergenceCurve differs by worker count: one=%v four=%v",
					oneWorker.ConvergenceCurve, fourWorkers.ConvergenceCurve)
			}

			if oneWorker.FuncEvalCount != fourWorkers.FuncEvalCount {
				t.Errorf("FuncEvalCount differs by worker count: one=%d four=%d",
					oneWorker.FuncEvalCount, fourWorkers.FuncEvalCount)
			}
		})
	}
}

func TestParallelGeneticOffspringAreFullyInitialized(t *testing.T) {
	config := NewDefaultConfig()
	config.ObjectiveFunc = sphere
	config.ProblemSize = 2
	config.LowerBound = -2
	config.UpperBound = 2
	config.MaxIterations = 5
	config.NC = 2
	config.NM = 3

	male := newMayfly(config.ProblemSize)
	male.Position = []float64{-1, 1}
	female := newMayfly(config.ProblemSize)
	female.Position = []float64{1, -1}

	pool := newEvaluationPool(config.ObjectiveFunc, 3)
	defer pool.close()

	offspring, best, err := evaluateParallelGeneticOperators(
		context.Background(),
		[]*Mayfly{male},
		[]*Mayfly{female},
		config,
		rand.New(rand.NewSource(99)),
		pool,
		0,
	)
	if err != nil {
		t.Fatalf("evaluateParallelGeneticOperators: %v", err)
	}

	if len(offspring) != config.NC+config.NM {
		t.Fatalf("offspring count = %d, want %d", len(offspring), config.NC+config.NM)
	}

	for i, mayfly := range offspring {
		if mayfly == nil {
			t.Fatalf("offspring[%d] is nil", i)
		}

		if got := sphere(mayfly.Position); mayfly.Cost != got {
			t.Errorf("offspring[%d].Cost = %v, want %v", i, mayfly.Cost, got)
		}

		if mayfly.Best.Cost != mayfly.Cost || !reflect.DeepEqual(mayfly.Best.Position, mayfly.Position) {
			t.Errorf("offspring[%d] personal best is not initialized: %+v", i, mayfly.Best)
		}

		for _, coordinate := range mayfly.Position {
			if coordinate < config.LowerBound || coordinate > config.UpperBound {
				t.Errorf("offspring[%d] coordinate %v is outside [%v,%v]",
					i, coordinate, config.LowerBound, config.UpperBound)
			}
		}
	}

	if best.Cost != sphere(best.Position) {
		t.Errorf("batch best cost = %v, want objective(position)=%v", best.Cost, sphere(best.Position))
	}

	savedBestCoordinate := offspring[0].Best.Position[0]

	offspring[0].Position[0]++
	if offspring[0].Best.Position[0] != savedBestCoordinate {
		t.Error("offspring personal best position aliases its current position")
	}
}

func TestParallelVariantEvaluationUsesVariantBatchCapacity(t *testing.T) {
	testCases := []struct {
		newConfig           func() *Config
		configure           func(*Config)
		name                string
		malePopulation      int
		femalePopulation    int
		wantConcurrency     int64
		wantEvaluationCount int
	}{
		{
			name:                "DESMA elites",
			newConfig:           NewDESMAConfig,
			malePopulation:      1,
			femalePopulation:    1,
			wantConcurrency:     4,
			wantEvaluationCount: 12,
			configure: func(config *Config) {
				config.EliteCount = 5
			},
		},
		{
			// 2 initial + 2 crossover + 1 mutation + 4 orthogonal candidates
			// + 1 chaotic exploitation candidate per elite male.
			name:                "OLCE candidates",
			newConfig:           NewOLCEConfig,
			malePopulation:      1,
			femalePopulation:    1,
			wantConcurrency:     4,
			wantEvaluationCount: 12,
			configure:           func(*Config) {},
		},
		{
			name:                "AOBLMOA opposition candidates",
			newConfig:           NewAOBLMOAConfig,
			malePopulation:      2,
			femalePopulation:    2,
			wantConcurrency:     4,
			wantEvaluationCount: 19,
			configure: func(config *Config) {
				config.AquilaWeight = 1
				config.OppositionProbability = 1
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var active atomic.Int64

			var maximum atomic.Int64

			var calls atomic.Int64

			objective := func(position []float64) float64 {
				calls.Add(1)
				current := active.Add(1)

				defer active.Add(-1)

				for {
					observed := maximum.Load()
					if current <= observed || maximum.CompareAndSwap(observed, current) {
						break
					}
				}

				time.Sleep(5 * time.Millisecond)

				return sphere(position)
			}

			config := testCase.newConfig()
			config.ObjectiveFunc = objective
			config.ProblemSize = 2
			config.LowerBound = -1
			config.UpperBound = 1
			config.MaxIterations = 1
			config.NPop = testCase.malePopulation
			config.NPopF = testCase.femalePopulation
			config.NC = 2
			config.NM = 1
			config.Rand = rand.New(rand.NewSource(42))
			config.EnableParallel = true
			config.MaxWorkers = 4
			testCase.configure(config)

			result, err := Optimize(config)
			if err != nil {
				t.Fatalf("Optimize: %v", err)
			}

			if got := maximum.Load(); got != testCase.wantConcurrency {
				t.Errorf("maximum concurrent evaluations = %d, want %d", got, testCase.wantConcurrency)
			}

			if got := int(calls.Load()); got != testCase.wantEvaluationCount {
				t.Errorf("objective calls = %d, want %d", got, testCase.wantEvaluationCount)
			}

			if result.FuncEvalCount != testCase.wantEvaluationCount {
				t.Errorf("FuncEvalCount = %d, want %d", result.FuncEvalCount, testCase.wantEvaluationCount)
			}
		})
	}
}

func TestParallelExecutionIsDeterministicForSeedAcrossSchedules(t *testing.T) {
	testCases := []struct {
		newConfig func() *Config
		configure func(*Config)
		name      string
	}{
		{name: "standard", newConfig: NewDefaultConfig, configure: func(*Config) {}},
		{name: "DESMA", newConfig: NewDESMAConfig, configure: func(*Config) {}},
		{name: "OLCE", newConfig: NewOLCEConfig, configure: func(*Config) {}},
		{name: "EOBBMA", newConfig: NewEOBBMAConfig, configure: func(config *Config) {
			config.OppositionRate = 1
		}},
		{name: "GSASMA", newConfig: NewGSASMAConfig, configure: func(config *Config) {
			config.ApplyOBLToGlobalBest = false
		}},
		{name: "MPMA", newConfig: NewMPMAConfig, configure: func(config *Config) {
			config.UseWeightedMedian = true
		}},
		{name: "AOBLMOA", newConfig: NewAOBLMOAConfig, configure: func(config *Config) {
			config.AquilaWeight = 1
			config.OppositionProbability = 1
		}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			run := func(name string, maxWorkers int) *Result {
				t.Helper()

				config := testCase.newConfig()
				config.ObjectiveFunc = sphere
				config.ProblemSize = 4
				config.LowerBound = -2
				config.UpperBound = 2
				config.MaxIterations = 3
				config.NPop = 5
				config.NPopF = 5
				config.NC = 4
				config.NM = 2
				config.Rand = rand.New(rand.NewSource(991))
				config.EnableParallel = true
				config.MaxWorkers = maxWorkers
				testCase.configure(config)

				result, err := Optimize(config)
				if err != nil {
					t.Fatalf("Optimize %s: %v", name, err)
				}

				return result
			}

			baseline := run("parallel/baseline", 1)
			for _, parallelRun := range []struct {
				name       string
				maxWorkers int
			}{
				{name: "parallel/1_worker_repeat", maxWorkers: 1},
				{name: "parallel/4_workers", maxWorkers: 4},
			} {
				result := run(parallelRun.name, parallelRun.maxWorkers)
				if !reflect.DeepEqual(result.GlobalBest, baseline.GlobalBest) ||
					!reflect.DeepEqual(result.ConvergenceCurve, baseline.ConvergenceCurve) ||
					result.FuncEvalCount != baseline.FuncEvalCount ||
					result.IterationCount != baseline.IterationCount {
					t.Errorf("%s result differs from seeded baseline:\nresult:   %+v\nbaseline: %+v",
						parallelRun.name, result, baseline)
				}
			}
		})
	}
}

func TestParallelVariantFunctionCountsAreExact(t *testing.T) {
	testCases := []struct {
		newConfig func() *Config
		configure func(*Config)
		name      string
	}{
		{name: "DESMA", newConfig: NewDESMAConfig, configure: func(*Config) {}},
		{name: "OLCE", newConfig: NewOLCEConfig, configure: func(*Config) {}},
		{name: "EOBBMA", newConfig: NewEOBBMAConfig, configure: func(config *Config) {
			config.OppositionRate = 1
		}},
		{name: "GSASMA", newConfig: NewGSASMAConfig, configure: func(*Config) {}},
		{name: "MPMA", newConfig: NewMPMAConfig, configure: func(*Config) {}},
		{name: "AOBLMOA", newConfig: NewAOBLMOAConfig, configure: func(config *Config) {
			config.AquilaWeight = 1
			config.OppositionProbability = 1
		}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var calls atomic.Int64

			config := testCase.newConfig()
			config.ObjectiveFunc = func(position []float64) float64 {
				calls.Add(1)

				return sphere(position)
			}
			config.ProblemSize = 3
			config.LowerBound = -2
			config.UpperBound = 2
			config.MaxIterations = 2
			config.NPop = 5
			config.NPopF = 5
			config.NC = 4
			config.NM = 2
			config.Rand = rand.New(rand.NewSource(199))
			config.EnableParallel = true
			config.MaxWorkers = 4
			testCase.configure(config)

			result, err := Optimize(config)
			if err != nil {
				t.Fatalf("Optimize: %v", err)
			}

			if got := int(calls.Load()); result.FuncEvalCount != got {
				t.Errorf("FuncEvalCount = %d, actual objective calls = %d", result.FuncEvalCount, got)
			}
		})
	}
}

func TestParallelMPMAMedianMatchesSequential(t *testing.T) {
	population := make([]*Mayfly, 6)
	for i := range population {
		population[i] = newMayfly(5)
		for dimension := range population[i].Position {
			population[i].Position[dimension] = float64((i+1)*(dimension+2)%7) - 3
		}
	}

	median, err := calculateMedianPositionParallel(context.Background(), population, 4)
	if err != nil {
		t.Fatalf("calculateMedianPositionParallel: %v", err)
	}

	if want := calculateMedianPosition(population); !reflect.DeepEqual(median, want) {
		t.Errorf("parallel median = %v, want %v", median, want)
	}

	weights := []float64{1, 0.8, 0.6, 0.4, 0.2, 0.1}

	weighted, err := calculateWeightedMedianPositionParallel(context.Background(), population, weights, 4)
	if err != nil {
		t.Fatalf("calculateWeightedMedianPositionParallel: %v", err)
	}

	if want := calculateWeightedMedianPosition(population, weights); !reflect.DeepEqual(weighted, want) {
		t.Errorf("parallel weighted median = %v, want %v", weighted, want)
	}
}

func TestParallelVariantCancellationDoesNotCommitPartialBatch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{}, 1)
	release := make(chan struct{})

	pool := newEvaluationPool(func(position []float64) float64 {
		select {
		case started <- struct{}{}:
		default:
		}

		<-release

		return sphere(position)
	}, 2)
	defer pool.close()

	males := []*Mayfly{newMayfly(2), newMayfly(2)}
	for i, male := range males {
		male.Position = []float64{float64(i + 1), float64(i + 2)}
		male.Cost = sphere(male.Position)
		copy(male.Best.Position, male.Position)
		male.Best.Cost = male.Cost
	}

	original := []*Mayfly{males[0].clone(), males[1].clone()}
	done := make(chan error, 1)

	go func() {
		_, err := evaluateParallelOrthogonalLearning(
			ctx,
			males,
			1,
			[]float64{0, 0},
			0.3,
			[]float64{-5, -5},
			[]float64{5, 5},
			rand.New(rand.NewSource(42)),
			pool,
		)
		done <- err
	}()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("variant objective evaluation did not start")
	}

	cancel()
	close(release)

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("variant evaluation did not return after cancellation")
	}

	for i := range males {
		if !reflect.DeepEqual(males[i], original[i]) {
			t.Errorf("males[%d] changed after canceled batch: got=%+v want=%+v", i, males[i], original[i])
		}
	}
}

func TestEvaluationPoolBestUsesStableIndexAndCopiesPosition(t *testing.T) {
	pool := newEvaluationPool(func([]float64) float64 { return 1 }, 3)
	defer pool.close()

	population := []*Mayfly{newMayfly(1), newMayfly(1), newMayfly(1)}
	for i, mayfly := range population {
		mayfly.Position[0] = float64(i + 1)
	}

	best, err := pool.evaluate(context.Background(), population, false, true)
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}

	if !reflect.DeepEqual(best.Position, []float64{1}) {
		t.Fatalf("best position = %v, want first population position", best.Position)
	}

	population[0].Position[0] = 99
	if !reflect.DeepEqual(best.Position, []float64{1}) {
		t.Errorf("best position changed after population mutation: %v", best.Position)
	}
}

func TestEvaluationPoolEvaluatesEveryCandidateExactlyOnce(t *testing.T) {
	const populationSize = 37

	callCounts := make([]atomic.Int64, populationSize)

	var totalCalls atomic.Int64

	pool := newEvaluationPool(func(position []float64) float64 {
		index := int(position[0])
		callCounts[index].Add(1)
		totalCalls.Add(1)

		return float64(index)
	}, 4)
	defer pool.close()

	population := make([]*Mayfly, populationSize)
	for i := range population {
		population[i] = newMayfly(1)
		population[i].Position[0] = float64(i)
	}

	_, err := pool.evaluate(context.Background(), population, false, false)
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}

	if got := totalCalls.Load(); got != populationSize {
		t.Errorf("objective calls = %d, want %d", got, populationSize)
	}

	for i := range population {
		if got := callCounts[i].Load(); got != 1 {
			t.Errorf("candidate %d objective calls = %d, want 1", i, got)
		}

		if population[i].Cost != float64(i) {
			t.Errorf("candidate %d cost = %v, want %d", i, population[i].Cost, i)
		}
	}
}

func TestParallelEvaluationCancellationWaitsForInflightCalls(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{}, 1)
	release := make(chan struct{})

	config := parallelTestConfig(func(position []float64) float64 {
		select {
		case started <- struct{}{}:
		default:
		}

		<-release

		return sphere(position)
	})

	done := make(chan error, 1)

	go func() {
		_, err := OptimizeContext(ctx, config)
		done <- err
	}()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("objective evaluation did not start")
	}

	cancel()
	close(release)

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("OptimizeContext did not return after cancellation")
	}
}

func TestParallelInitializationSanitizesInvalidCosts(t *testing.T) {
	config := parallelTestConfig(func([]float64) float64 { return math.Inf(1) })
	config.MaxIterations = 1

	result, err := Optimize(config)
	if err != nil {
		t.Fatalf("Optimize: %v", err)
	}

	if math.IsInf(result.GlobalBest.Cost, 0) || math.IsNaN(result.GlobalBest.Cost) {
		t.Errorf("GlobalBest.Cost = %v, want sanitized finite cost", result.GlobalBest.Cost)
	}
}

func TestEffectiveMaxWorkers(t *testing.T) {
	config := &Config{}
	if got := effectiveMaxWorkers(config); got != defaultMaxWorkers() {
		t.Errorf("effectiveMaxWorkers with zero = %d, want %d", got, defaultMaxWorkers())
	}

	config.MaxWorkers = 2
	if got := effectiveMaxWorkers(config); got != 2 {
		t.Errorf("effectiveMaxWorkers = %d, want 2", got)
	}
}
