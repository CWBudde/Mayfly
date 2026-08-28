# Choosing a Mayfly Algorithm

Mayfly provides one standard algorithm and seven specialized variants. Choose
with evidence from your objective: start with problem knowledge, use the
selector as a shortlist, and compare that shortlist across repeated paired
seeds. A selector score is a static rule-of-thumb match, not a measured success
probability or a guarantee that one variant will win.

All variants in this library optimize a bounded, continuous, scalar objective.
`MultiObjective: true` is rejected because no multi-objective optimizer is
implemented. For constraints and mixed-variable projection, first read the
[configuration guide](api/configuration.md#constraint-handling).

## Select and run a variant

The following complete program describes a known Ackley-like problem, asks
for ranked recommendations, and runs the first recommendation with a fixed
seed. It is also available as the runnable
[`cmd/algorithm-selection`](../cmd/algorithm-selection/) example and is kept in
sync by a test.

```go
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
```

From the repository root, run it with:

```bash
go run ./cmd/algorithm-selection
```

Use the checked selector and builder entry points for runtime data. They return
errors for invalid metadata, unsupported multi-objective problems, nil
variants, and invalid configurations instead of relying on compatibility
fallbacks.

## Describe the problem

`ProblemCharacteristics` records facts and priorities that the selector cannot
infer from the objective alone:

- `Dimensionality` is the number of decision variables.
- `Modality` is `Unimodal`, `Multimodal`, or `HighlyMultimodal`.
- `Landscape` is `Smooth`, `Rugged`, `Deceptive`, or `NarrowValley`.
- `ExpensiveEvaluations` favors variants with lower estimated evaluation
  overhead.
- `RequiresFastConvergence` and `RequiresStableConvergence` express different
  priorities; set them only when the application actually has that need.
- `MultiObjective` must remain false. Multiple objectives require a scalarization
  chosen by the caller or a different optimizer.

Use domain knowledge when it is available. For example, an objective can be
deceptive because its global optimum lies far from the basin suggested by most
samples; a short automatic scan cannot establish that fact.

## How the built-in variants differ

This table summarizes the rules currently used by `ApplicableTo`. The overhead
column is an approximate selector hint relative to Standard MA; actual
evaluation counts depend on configuration and stopping.

| Variant     | Selector favors it when                                         | Estimated overhead |
| ----------- | --------------------------------------------------------------- | -----------------: |
| Standard MA | unimodal, smooth, or at most 50 dimensions                      |              1.00x |
| DESMA       | multimodal or rugged, with inexpensive evaluations              |              1.08x |
| OLCE-MA     | highly multimodal, rugged, or at least 10 dimensions            |              1.33x |
| EOBBMA      | deceptive or highly multimodal, including expensive evaluations |             1.015x |
| GSASMA      | fast convergence is requested or the problem is multimodal      |              1.07x |
| HMMA        | non-unimodal, rugged, or deceptive                              |              1.02x |
| MPMA        | stable convergence, a narrow valley, or expensive evaluations   |              1.00x |
| AOBLMOA     | highly multimodal, rugged, or deceptive                         |              1.25x |

Keep Standard MA in every comparison as a baseline. Add the two or three
specialized variants whose update rules match the problem, then measure them
under the same objective-evaluation budget. The individual
[algorithm guides](algorithms/) explain each update rule and its parameters.

## Read a recommendation correctly

`RecommendAlgorithmsChecked` returns every variant in descending `Score`
order. Each item contains:

- `Variant`, which supplies defaults and can be passed to
  `NewBuilderFromVariantChecked`;
- `Score`, the result of hand-written applicability rules in `[0, 1]`;
- `Confidence`, another heuristic adjusted for a few known characteristics;
- `Reasoning`, a human-readable summary of matched rules.

Scores are useful for ranking candidates for the same description. They are
not objective values, benchmark results, calibrated probabilities, or valid
cross-problem performance measurements. Ties are possible, and small score
differences are not evidence of a meaningful performance difference.

For a name supplied by a user, resolve it with `NewVariantChecked`. For one of
the built-in benchmark names, `RecommendForBenchmarkChecked` is a convenient
metadata preset. Neither shortcut replaces validation on the real workload.

## Classify an unfamiliar objective

When domain knowledge is limited, `ClassifyProblemContext` can sample the
objective. Pass a seeded `rand.Rand` for a reproducible classification and a
context or evaluation budget when objective calls must be bounded. The
classifier estimates modality, reports only `Smooth` or `Rugged`, and probes
seed sensitivity. It cannot infer `Deceptive`, `NarrowValley`, expensive
evaluations, deadlines, or multi-objective intent; set those fields yourself.

Classification consumes objective evaluations before optimization. Its result
is coarse and scale-dependent details can still matter, so treat it as a way to
build a shortlist rather than as an automatic final decision. See the
[classifier API](api/unified-framework.md#automatic-problem-classification) for
the sampling call and budget controls.

## Validate the shortlist

Stochastic algorithms must be compared over multiple runs. Use
`ComparisonRunner` with:

1. the same variant list, objective, bounds, and stopping rules;
2. `WithSeed` so each run index uses the same paired seed across variants;
3. `WithMaxEvaluations` when variants perform different numbers of objective
   calls per iteration;
4. enough independent runs to inspect both central tendency and spread;
5. a representative validation workload separate from any cases used to tune
   parameters.

Do not select from one seed or from the lowest single cost. Review run-level
failures, feasibility, mean or median cost, dispersion, and practical effect
size. The [comparison framework](api/comparison-framework.md) covers paired
execution, statistical output, exact evaluation budgets, and exports.

## Common selection mistakes

- Treating the top score or `Confidence` as a performance guarantee.
- Marking a problem `Deceptive` or `NarrowValley` based only on the automatic
  classifier.
- Comparing equal iteration counts when variants consume different numbers of
  objective evaluations.
- Tuning on one benchmark or seed and reporting the same run as validation.
- Omitting Standard MA, which removes the simplest project-local baseline.
- Passing `MultiObjective: true`; Mayfly currently supports scalar objectives
  only.

After choosing a variant, keep its defaults initially. Change parameters only
through repeated, budget-matched measurements; the
[parameter-tuning tutorial](parameter-tuning.md) provides a complete workflow.
