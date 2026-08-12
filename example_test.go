package mayfly_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"math/rand"

	"github.com/cwbudde/mayfly"
)

func ExampleOptimize() {
	config := mayfly.NewDefaultConfig()
	config.ObjectiveFunc = mayfly.Sphere
	config.ProblemSize = 2
	config.LowerBound = -10
	config.UpperBound = 10
	config.MaxIterations = 1
	config.NPop = 1
	config.NPopF = 1
	config.NC = 0
	config.Rand = rand.New(rand.NewSource(42))

	result, err := mayfly.Optimize(config)
	if err != nil {
		panic(err)
	}

	fmt.Println(len(result.GlobalBest.Position), result.IterationCount)
	// Output: 2 1
}

func ExampleOptimizeContext() {
	config := mayfly.NewDefaultConfig()
	config.ObjectiveFunc = mayfly.Rastrigin
	config.ProblemSize = 2
	config.LowerBound = -5.12
	config.UpperBound = 5.12
	config.MaxIterations = 1
	config.NPop = 1
	config.NPopF = 1
	config.NC = 0

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	updates := 0

	result, err := mayfly.OptimizeContext(
		context.Background(),
		config,
		mayfly.WithLogger(logger),
		mayfly.WithProgressObserver(func(progress mayfly.Progress) {
			updates++
		}),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(result.IterationCount, updates, result.TerminationReason)
	// Output: 1 1 maximum_iterations
}

func ExampleNewBuilder() {
	result, err := mayfly.NewBuilder("olce").
		ForProblem(mayfly.Rastrigin, 2, -5.12, 5.12).
		WithIterations(1).
		WithPopulation(1, 1).
		WithConfig(func(config *mayfly.Config) {
			config.NC = 0
		}).
		Optimize()
	if err != nil {
		panic(err)
	}

	fmt.Println(result.IterationCount)
	// Output: 1
}

func ExampleComparisonRunner_CompareContext() {
	runner := mayfly.NewComparisonRunner().
		WithVariantNames("ma", "desma", "olce").
		WithRuns(1).
		WithIterations(1).
		WithSeed(42).
		WithParallel(true).
		WithMaxWorkers(4)

	result, err := runner.CompareContext(
		context.Background(),
		"Rastrigin-10D",
		mayfly.Rastrigin,
		10,
		-5.12,
		5.12,
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(len(result.AlgorithmNames), len(result.RunResults))
	// Output: 3 3
}
