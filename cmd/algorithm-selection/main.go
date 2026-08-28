package main

import (
	"fmt"
	"os"

	"github.com/cwbudde/mayfly"
)

func chooseAndOptimize() (mayfly.AlgorithmRecommendation, *mayfly.Result, error) {
	characteristics := mayfly.ProblemCharacteristics{
		Dimensionality:            10,
		Modality:                  mayfly.Multimodal,
		Landscape:                 mayfly.Rugged,
		ExpensiveEvaluations:      false,
		RequiresFastConvergence:   false,
		RequiresStableConvergence: false,
		MultiObjective:            false,
	}

	selector := mayfly.NewAlgorithmSelector()

	recommendations, err := selector.RecommendAlgorithmsChecked(characteristics)
	if err != nil {
		return mayfly.AlgorithmRecommendation{}, nil, err
	}

	selected := recommendations[0]

	builder, err := mayfly.NewBuilderFromVariantChecked(selected.Variant)
	if err != nil {
		return mayfly.AlgorithmRecommendation{}, nil, err
	}

	seed := int64(42)

	result, err := builder.
		ForProblem(mayfly.Ackley, 10, -32.768, 32.768).
		WithIterations(200).
		WithConfig(func(config *mayfly.Config) {
			config.Seed = &seed
		}).
		Optimize()
	if err != nil {
		return mayfly.AlgorithmRecommendation{}, nil, err
	}

	return selected, result, nil
}

func main() {
	selected, result, err := chooseAndOptimize()
	if err != nil {
		fmt.Fprintf(os.Stderr, "selection or optimization failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "selected: %s\n", selected.Variant.Name())
	fmt.Fprintf(os.Stdout, "heuristic score: %.2f; confidence: %.2f\n",
		selected.Score,
		selected.Confidence,
	)
	fmt.Fprintf(os.Stdout, "reason: %s\n", selected.Reasoning)
	fmt.Fprintf(os.Stdout, "best cost: %.8g; evaluations: %d\n",
		result.GlobalBest.Cost,
		result.FuncEvalCount,
	)
}
