# Validated helper APIs

Mayfly's optimizer entry points validate their complete configuration. Direct
operator and utility calls do not have that surrounding context, so v0.7 adds
checked APIs that return errors instead of panicking, using package-global
randomness, or replacing invalid values with defaults.

Use the checked forms for all new code:

| Compatibility API | Checked API |
| --- | --- |
| `NewVariant`, `NewBuilder`, `NewBuilderFromVariant` | `NewVariantChecked`, `NewBuilderChecked`, `NewBuilderFromVariantChecked` |
| `NewAnnealingScheduler`, `Update`, `GetTemperature`, `Reset` | checked constructor, checked methods, and `AnnealingScheduler.Validate` |
| `NewLogisticMap`, `Next`, `Current`, `Reset` | checked constructor and corresponding checked methods |
| `NewParetoArchive` | `NewParetoArchiveChecked` or `NewParetoArchiveWithObjectives` |
| `Crossover`, `CrossoverBlend` | `CrossoverChecked`, `CrossoverBlendChecked` |
| `Mutate`, `MutateGaussian`, `MutateCauchy`, `HybridMutate` | the corresponding `...Checked` mutation function |
| `OrthogonalArray`, `ApplyOrthogonalLearning`, `ApplyOrthogonalLearningToElite` | the corresponding `...Checked` function |
| `EvaluateConstraints`, `PenalizedCost`, `BetterConstrainedCandidate` | the corresponding `...Checked` function |
| `AlgorithmSelector.RecommendAlgorithms`, `RecommendBest` | `RecommendAlgorithmsChecked`, `RecommendBestChecked` |
| `ComparisonRunner.WithVariants`, `WithVariantNames` | `WithVariantsChecked`, `WithVariantNamesChecked` |
| `RecommendForBenchmark` | `RecommendForBenchmarkChecked` |
| `PrintRecommendations` | `WriteRecommendations` |
| `PrintPresets` | `WritePresets` |
| `AutoTuneConfig` | `AutoTuneConfigChecked` |

Checked stochastic operators require a non-nil `*rand.Rand`, finite increasing
bounds, finite in-bounds vectors of equal dimension, and probabilities or rates
in their documented ranges. This makes randomness explicit and reproducible.

The legacy APIs remain for source compatibility and are deprecated. Their old
fallback behavior is documented on each function; callers should not use it to
validate user-controlled data.

`ParetoArchive.Solutions` and `ParetoArchive.MaxSize` are deprecated snapshots.
Use `GetSolutions` and `Capacity`; archive inputs and outputs are deep copied.
`AnnealingScheduler` retains exported fields for compatibility, so call
`Validate` after modifying them directly.
