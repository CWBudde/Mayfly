package mayfly

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"reflect"
	"testing"
)

func minimalCoreConfig(objective ObjectiveFunction) *Config {
	config := NewDefaultConfig()
	config.ObjectiveFunc = objective
	config.ProblemSize = 1
	config.LowerBound = -2
	config.UpperBound = 2
	config.MaxIterations = 1
	config.NPop = 1
	config.NPopF = 1
	config.NC = 0
	config.NM = 0
	config.G = 0
	config.A1 = 0
	config.A2 = 0
	config.A3 = 0
	config.Dance = 0
	config.FL = 0
	config.Rand = rand.New(rand.NewSource(1))

	return config
}

func TestFemaleCandidateCanBecomeGlobalBest(t *testing.T) {
	config := minimalCoreConfig(sphere)

	result, err := OptimizeContext(
		context.Background(),
		config,
		WithInitialPopulation([][]float64{{1}}, [][]float64{{0}}),
	)
	if err != nil {
		t.Fatalf("OptimizeContext: %v", err)
	}

	if result.GlobalBest.Cost != 0 || !reflect.DeepEqual(result.GlobalBest.Position, []float64{0}) {
		t.Fatalf("GlobalBest = %+v, want evaluated female optimum at [0]", result.GlobalBest)
	}
}

func TestOptimizeDoesNotResolveDefaultsIntoCallerConfig(t *testing.T) {
	config := minimalCoreConfig(sphere)

	config.Rand = nil
	if config.VelMin != 0 || config.VelMax != 0 {
		t.Fatal("test requires automatic velocity bounds")
	}

	_, err := OptimizeContext(context.Background(), config)
	if err != nil {
		t.Fatalf("OptimizeContext: %v", err)
	}

	if config.Rand != nil || config.VelMin != 0 || config.VelMax != 0 {
		t.Fatalf("OptimizeContext mutated caller Config: Rand=%p VelMin=%v VelMax=%v",
			config.Rand, config.VelMin, config.VelMax)
	}
}

func TestExplicitSeedIsTruthfulAndReusable(t *testing.T) {
	seed := int64(20260824)
	config := minimalCoreConfig(sphere)
	config.Rand = nil
	config.Seed = &seed

	first, err := Optimize(config)
	if err != nil {
		t.Fatalf("first Optimize: %v", err)
	}

	second, err := Optimize(config)
	if err != nil {
		t.Fatalf("second Optimize: %v", err)
	}

	if first.Seed == nil || *first.Seed != seed || second.Seed == nil || *second.Seed != seed {
		t.Fatalf("reported seeds = %v and %v, want %d", first.Seed, second.Seed, seed)
	}

	if !reflect.DeepEqual(first.GlobalBest, second.GlobalBest) ||
		!reflect.DeepEqual(first.ConvergenceCurve, second.ConvergenceCurve) {
		t.Fatal("reusing an explicitly seeded Config changed the run")
	}
}

func TestOpaqueRandHasNoReportedSeed(t *testing.T) {
	config := minimalCoreConfig(sphere)

	result, err := Optimize(config)
	if err != nil {
		t.Fatalf("Optimize: %v", err)
	}

	if result.Seed != nil {
		t.Fatalf("Result.Seed = %v for caller-owned Rand, want nil", *result.Seed)
	}
}

func TestSeedAndRandAreMutuallyExclusive(t *testing.T) {
	config := minimalCoreConfig(sphere)
	seed := int64(1)

	config.Seed = &seed
	if _, err := Optimize(config); err == nil {
		t.Fatal("Optimize accepted both Config.Seed and Config.Rand")
	}
}

func TestNegativeInfinityObjectiveIsInvalidNotOptimal(t *testing.T) {
	objective := func(position []float64) float64 {
		if position[0] < 0 {
			return math.Inf(-1)
		}

		return position[0] * position[0]
	}
	config := minimalCoreConfig(objective)

	result, err := OptimizeContext(
		context.Background(),
		config,
		WithInitialPopulation([][]float64{{-1}}, [][]float64{{0.5}}),
	)
	if err != nil {
		t.Fatalf("OptimizeContext: %v", err)
	}

	if result.GlobalBest.Cost != 0.25 {
		t.Fatalf("GlobalBest.Cost = %v, want finite candidate cost 0.25", result.GlobalBest.Cost)
	}
}

func TestAllNonFiniteInitialObjectivesReturnError(t *testing.T) {
	config := minimalCoreConfig(func([]float64) float64 { return math.NaN() })

	result, err := OptimizeContext(context.Background(), config)
	if !errors.Is(err, ErrNoFiniteObjectiveValue) {
		t.Fatalf("OptimizeContext error = %v, want %v", err, ErrNoFiniteObjectiveValue)
	}

	if result != nil {
		t.Fatalf("OptimizeContext returned result with invalid initialization: %+v", result)
	}
}

func TestParallelGeneticMutationPreservesSex(t *testing.T) {
	config := minimalCoreConfig(sphere)
	config.NPop = 2
	config.NPopF = 2
	config.NC = 0
	config.NM = 2
	config.Mu = 0

	males := []*Mayfly{newMayfly(1), newMayfly(1)}
	females := []*Mayfly{newMayfly(1), newMayfly(1)}
	males[0].Position[0], males[1].Position[0] = -2, -1
	females[0].Position[0], females[1].Position[0] = 1, 2

	pool := newEvaluationPool(config.ObjectiveFunc, 2)
	defer pool.close()

	offspring, _, _, err := evaluateParallelGeneticOperators(
		context.Background(), males, females, config,
		rand.New(rand.NewSource(9)), pool, 0,
	)
	if err != nil {
		t.Fatalf("evaluateParallelGeneticOperators: %v", err)
	}

	if len(offspring) != 4 {
		t.Fatalf("offspring count = %d, want two mutants per sex", len(offspring))
	}

	for i, child := range offspring[:2] {
		if child.Position[0] >= 0 {
			t.Errorf("male mutant %d inherited female position %v", i, child.Position)
		}
	}

	for i, child := range offspring[2:] {
		if child.Position[0] <= 0 {
			t.Errorf("female mutant %d inherited male position %v", i, child.Position)
		}
	}
}

func TestStandardSequentialAndParallelRunsMatchExactly(t *testing.T) {
	run := func(parallel bool) *Result {
		t.Helper()

		config := NewDefaultConfig()
		config.ObjectiveFunc = sphere
		config.ProblemSize = 4
		config.LowerBound = -3
		config.UpperBound = 3
		config.MaxIterations = 4
		config.NPop = 6
		config.NPopF = 6
		config.NC = 4
		config.NM = 2
		config.Rand = rand.New(rand.NewSource(20260824))
		config.EnableParallel = parallel
		config.MaxWorkers = 3

		result, err := Optimize(config)
		if err != nil {
			t.Fatalf("Optimize(parallel=%v): %v", parallel, err)
		}

		return result
	}

	sequential := run(false)

	parallel := run(true)
	if !reflect.DeepEqual(sequential.GlobalBest, parallel.GlobalBest) ||
		!reflect.DeepEqual(sequential.ConvergenceCurve, parallel.ConvergenceCurve) ||
		sequential.FuncEvalCount != parallel.FuncEvalCount ||
		sequential.IterationCount != parallel.IterationCount {
		t.Fatalf("sequential and parallel runs differ:\nsequential=%+v\nparallel=%+v",
			sequential, parallel)
	}
}
