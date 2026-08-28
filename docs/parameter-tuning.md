# Tuning Mayfly Parameters

Parameter tuning is an experiment, not a way to make one lucky run look good.
Start from the selected variant's factory defaults, change a small number of
parameters with a stated rationale, compare candidates under paired random
seeds and equal objective-evaluation budgets, and confirm the choice on seeds
or problem instances that did not select it.

Mayfly minimizes scalar objectives. Before tuning, make sure the objective,
bounds, constraints, and variant are correct; tuning cannot repair a model that
rewards the wrong behavior. The [custom-objective guide](custom-objective-functions.md)
covers that modeling and validation step.

## Tune one parameter reproducibly

This complete example tunes the male and female population sizes together for
Standard MA on 10-dimensional Rastrigin. It uses five runs to keep the example
quick; use at least 20 to 30 runs for a consequential decision. The example is
also available as [`cmd/parameter-tuning`](../cmd/parameter-tuning/) and its
first code block is kept in sync by a test.

```go
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
```

Run it from the repository root:

```bash
go run ./cmd/parameter-tuning
```

`populationVariant` is a small adapter: `ComparisonRunner` asks each
`AlgorithmVariant` for a configuration, so the adapter can expose several
configurations of the same algorithm as separately named candidates. Its
`GetConfig` starts from a fresh factory configuration on every call. This
avoids sharing mutable `Config` or `rand.Rand` values across runs.

The runner derives the same seed for every candidate at a given run index.
`WithMaxEvaluations` also gives each candidate exactly 3,000 objective calls;
equal iteration counts would favor population sizes that do less work per
iteration. `WithIterations(1_000)` is only a safety ceiling and must be high
enough for every candidate to consume the evaluation budget.

The first seed range selects a candidate. The disjoint second range compares
that choice with the population-20 default. If population 20 wins the tuning
runs, the held-out comparison deliberately contains equivalent configurations;
the correct conclusion is to keep the simpler default, not to claim an
improvement.

## A practical tuning workflow

1. Freeze the objective definition, dimensions, bounds, constraint handling,
   selected variant, stopping rule, and hardware policy.
2. Record the untouched factory configuration as the baseline.
3. Choose one influential parameter and a small range that includes the
   default. Use domain knowledge or coarse logarithmic spacing rather than a
   dense arbitrary grid.
4. Reject invalid candidates with `ValidateConfig` before spending an
   experimental budget.
5. Compare every candidate with paired seeds and the same objective-evaluation
   budget. Inspect run failures and spread as well as mean or median cost.
6. Select only a practically meaningful improvement. When results are close,
   prefer the default or the cheaper configuration.
7. Confirm once on held-out seeds and, ideally, representative problem
   instances. Do not tune again on the held-out results.
8. Save the selected configuration, seed schedules, library version, objective
   budget, and raw results so the decision can be reproduced.

After a one-parameter sweep, tune a second parameter only if the first result
is stable. A full Cartesian grid grows rapidly and repeated selection increases
the chance of finding a candidate that is lucky on the tuning data. For
important studies, use nested validation or multiple problem instances and
report the complete search space, not only the winner.

## What to tune first

| Parameter group                               | Tune when                                            | Important coupling                                                 |
| --------------------------------------------- | ---------------------------------------------------- | ------------------------------------------------------------------ |
| `NPop`, `NPopF`                               | Diversity or per-generation cost is the main concern | Keep `NPopF <= NPop`; tune under an evaluation budget              |
| `MaxIterations`                               | There is no exact evaluation budget                  | It is a stopping budget, not an algorithm-quality parameter        |
| `G`, `GDamp`, `A1`, `A2`, `A3`, `Beta`        | Movement is clearly too exploratory or exploitative  | Change one coefficient family at a time                            |
| `Dance`, `FL`, `DanceDamp`, `FLDamp`          | Unpaired movement needs adjustment                   | Effects change over the run because of damping                     |
| `NC`, `NCRatio`, `NM`, `Mu`, `CrossoverGamma` | Offspring diversity or cost needs adjustment         | `NC` work scales with population; `Mu` is a fraction of dimensions |
| Variant-specific fields                       | The chosen variant's mechanism is relevant           | Use that variant's algorithm guide and documented ranges           |

Start with population size because it has a clear cost/diversity tradeoff. Keep
`NPop` and `NPopF` equal unless the application gives a reason not to; females
cannot outnumber males. Leave `NC` at `NCAuto` when tuning population size so
`NCRatio` continues to scale crossover work with the population.

Do not tune `Seed`: seeds describe experimental replication, not model
parameters. Similarly, `EnableParallel` and `MaxWorkers` are throughput
settings. Benchmark those for wall-clock performance while holding the
algorithm configuration fixed.

## Use heuristic auto-tuning carefully

`AutoTuneConfigChecked` is a deterministic convenience heuristic, not an
empirical tuner. It rejects a nil configuration, negative dimensionality, and
unknown modality or landscape enum values, then adjusts only a few fields:
population sizes for at least 20 or 50 dimensions, the iteration limit for
fast or high-dimensional runs, and selected OLCE-MA, GSASMA, or MPMA fields.
`ExpensiveEvaluations`, `RequiresStableConvergence`, and `MultiObjective` do
not affect these heuristics; the optimizer still requires a scalar objective.
The helper does not run the objective, search a parameter space, validate the
complete resulting configuration, estimate uncertainty, or prove that the
configuration is better.

```go
config := mayfly.NewOLCEConfig()
characteristics := mayfly.ProblemCharacteristics{
	Dimensionality:            50,
	Modality:                  mayfly.HighlyMultimodal,
	Landscape:                 mayfly.Rugged,
	ExpensiveEvaluations:      false,
	RequiresFastConvergence:   false,
	RequiresStableConvergence: false,
	MultiObjective:            false,
}

if err := mayfly.AutoTuneConfigChecked(config, characteristics); err != nil {
	return err
}
```

Treat that output as another candidate and validate it against the factory
default with the same process as any hand-written configuration. Note that
`AutoTuneConfigChecked` can change `MaxIterations`; an explicit
`ComparisonRunner.WithIterations` overrides it for the comparison.

## Validate and preserve a configuration

Builder customizers are convenient for a final known configuration:

```go
seed := int64(42)
builder, err := mayfly.NewBuilderChecked("ma")
if err != nil {
	return err
}

config, err := builder.
	ForProblem(objective, dimensions, lower, upper).
	WithPopulation(12, 12).
	WithConfig(func(config *mayfly.Config) {
		config.Seed = &seed
	}).
	Build()
if err != nil {
	return err
}
```

`Build` returns a validated clone. For a directly constructed configuration,
call `ValidateConfig` explicitly. `SaveConfigToFile` can preserve serializable
fields, but objective and constraint functions and `Rand` are not JSON values;
reattach them after loading. Record `Result.FuncEvalCount`, `IterationCount`,
`Seed`, and `TerminationReason` with the final cost.

## Common tuning mistakes

- Choosing the lowest cost from a single stochastic run.
- Comparing equal iterations when candidates perform different work per
  iteration.
- Selecting and reporting performance on the same seeds or instances.
- Sweeping many interacting parameters without enough runs for each candidate.
- Ignoring failed, infeasible, `NaN`, or infinite runs in an aggregate.
- Treating a small mean difference as meaningful without checking dispersion,
  paired results, and application impact.
- Changing the objective, bounds, variant, parallelism, and parameters in one
  experiment, making the source of any difference unknowable.
- Assuming `AutoTuneConfigChecked` measured the real objective.

For field semantics and valid ranges, use the
[configuration reference](api/configuration.md). For statistical output,
paired tests, exact budgets, and exports, use the
[comparison framework](api/comparison-framework.md).
