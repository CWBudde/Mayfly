package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cwbudde/mayfly"
)

const (
	dimensions       = 10
	evaluationBudget = 3_000
	tuningSeed       = int64(42)
	validationSeed   = int64(10_042)
	comparisonRuns   = 5
)

type populationVariant struct {
	base       mayfly.AlgorithmVariant
	name       string
	population int
}

func (variant populationVariant) Name() string { return variant.name }

func (variant populationVariant) FullName() string {
	return fmt.Sprintf("%s with population %d", variant.base.FullName(), variant.population)
}

func (variant populationVariant) Description() string { return variant.base.Description() }

func (variant populationVariant) GetConfig() *mayfly.Config {
	config := variant.base.GetConfig()
	config.NPop = variant.population
	config.NPopF = variant.population

	return config
}

func (variant populationVariant) ApplicableTo(characteristics mayfly.ProblemCharacteristics) float64 {
	return variant.base.ApplicableTo(characteristics)
}

func (variant populationVariant) EstimatedOverhead() float64 {
	return variant.base.EstimatedOverhead()
}

func (variant populationVariant) RecommendedFor() []string {
	return variant.base.RecommendedFor()
}

func comparePopulations(
	ctx context.Context,
	populations []int,
	seed int64,
) (*mayfly.ComparisonResult, error) {
	base, err := mayfly.NewVariantChecked("ma")
	if err != nil {
		return nil, err
	}

	variants := make([]mayfly.AlgorithmVariant, 0, len(populations))
	for _, population := range populations {
		variants = append(variants, populationVariant{
			base:       base,
			name:       fmt.Sprintf("MA-pop-%d", population),
			population: population,
		})
	}

	runner, err := mayfly.NewComparisonRunner().WithVariantsChecked(variants...)
	if err != nil {
		return nil, err
	}

	return runner.
		WithRuns(comparisonRuns).
		WithIterations(1_000).
		WithMaxEvaluations(evaluationBudget).
		WithSeed(seed).
		CompareContext(ctx, "Rastrigin-10D", mayfly.Rastrigin, dimensions, -5.12, 5.12)
}

func tuneAndValidate(ctx context.Context) (int, *mayfly.ComparisonResult, *mayfly.ComparisonResult, error) {
	populations := []int{12, 20, 32}

	tuning, err := comparePopulations(ctx, populations, tuningSeed)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("tuning comparison: %w", err)
	}

	selectedPopulation := populations[tuning.BestAlgorithm]

	validation, err := comparePopulations(ctx, []int{20, selectedPopulation}, validationSeed)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("validation comparison: %w", err)
	}

	return selectedPopulation, tuning, validation, nil
}

func main() {
	selected, tuning, validation, err := tuneAndValidate(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "parameter tuning failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "selected population: %d\n", selected)
	fmt.Fprintf(os.Stdout, "tuning mean cost: %.8g (%s)\n",
		tuning.Statistics[tuning.BestAlgorithm].Mean,
		tuning.AlgorithmNames[tuning.BestAlgorithm],
	)
	fmt.Fprintf(os.Stdout, "held-out mean costs: default=%.8g selected=%.8g\n",
		validation.Statistics[0].Mean,
		validation.Statistics[1].Mean,
	)
	fmt.Fprintf(os.Stdout, "budget: %d evaluations/run; seeds: %d-%d and %d-%d\n",
		evaluationBudget,
		tuningSeed,
		tuningSeed+comparisonRuns-1,
		validationSeed,
		validationSeed+comparisonRuns-1,
	)
}
