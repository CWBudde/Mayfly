# API Quick Reference

This page is a compact map of Mayfly's public, application-level API. See the
[configuration guide](configuration.md) for every `Config` field and the
[run lifecycle guide](run-lifecycle.md) for cancellation, progress, logging,
initial populations, and convergence export.

## Minimal optimization

```go
config := mayfly.NewDefaultConfig()
config.ObjectiveFunc = mayfly.Sphere
config.ProblemSize = 10
config.LowerBound = -10
config.UpperBound = 10

result, err := mayfly.Optimize(config)
if err != nil {
    return err
}

fmt.Printf("cost=%g position=%v\n", result.GlobalBest.Cost, result.GlobalBest.Position)
```

Mayfly minimizes scalar objective functions. Negate the returned value to
maximize a quantity.

## Entry points

| API                                        | Use                                                         |
| ------------------------------------------ | ----------------------------------------------------------- |
| `Optimize(config)`                         | Run with a background context and no run options            |
| `OptimizeContext(ctx, config, options...)` | Add cancellation, initial populations, progress, or logging |
| `NewBuilder(name)`                         | Configure a registered variant with a fluent builder        |
| `NewBuilderFromVariant(variant)`           | Build from an `AlgorithmVariant` object                     |

`Optimize` and `OptimizeContext` return `(*Result, error)`. Configuration is
validated before objective evaluation starts.

## Configuration factories

| Factory              | Variant                                      | Flag enabled |
| -------------------- | -------------------------------------------- | ------------ |
| `NewDefaultConfig()` | Standard MA                                  | none         |
| `NewDESMAConfig()`   | Dynamic Elite Strategy                       | `UseDESMA`   |
| `NewOLCEConfig()`    | Orthogonal Learning and Chaotic Exploitation | `UseOLCE`    |
| `NewEOBBMAConfig()`  | Elite Opposition-Based Bare Bones            | `UseEOBBMA`  |
| `NewGSASMAConfig()`  | Golden Sine with Simulated Annealing         | `UseGSASMA`  |
| `NewHMMAConfig()`    | Hybrid Mutation                              | `UseHMMA`    |
| `NewMPMAConfig()`    | Median Position-Based                        | `UseMPMA`    |
| `NewAOBLMOAConfig()` | Aquila Optimizer-Based Learning              | `UseAOBLMOA` |

Every factory still requires `ObjectiveFunc`, `ProblemSize`, `LowerBound`, and
`UpperBound`. Only one `Use...` variant flag may be true.

## Run options

```go
result, err := mayfly.OptimizeContext(
    ctx,
    config,
    mayfly.WithInitialPopulation(maleSeeds, femaleSeeds),
    mayfly.WithProgressObserver(func(p mayfly.Progress) {
        fmt.Printf("iteration=%d cost=%g\n", p.Iteration, p.Best.Cost)
    }),
    mayfly.WithLogger(slog.Default()),
)
```

| Option                  | Parameters                                                               |
| ----------------------- | ------------------------------------------------------------------------ |
| `WithInitialPopulation` | Male and female position slices; either may be empty or partial          |
| `WithProgressObserver`  | Synchronous callback receiving a snapshot after each completed iteration |
| `WithLogger`            | `mayfly.Logger`; `*slog.Logger` implements the interface                 |

## Result fields and exports

| Field                            | Meaning                                          |
| -------------------------------- | ------------------------------------------------ |
| `GlobalBest.Position`            | Best decision vector found                       |
| `GlobalBest.Cost`                | Raw objective value at that position             |
| `GlobalBest.ConstraintViolation` | Aggregate constraint violation; zero is feasible |
| `ConvergenceCurve`               | Best cost after each completed iteration         |
| `IterationCount`                 | Number of completed iterations                   |
| `FuncEvalCount`                  | Number of objective evaluations                  |
| `TerminationReason`              | Maximum iterations, target cost, or stagnation   |
| `Seed`                           | Seed used by the run; nil for caller-owned RNGs  |

```go
if err := result.ExportConvergenceCSV("convergence.csv"); err != nil {
    return err
}
if err := result.ExportConvergenceJSON("convergence.json"); err != nil {
    return err
}
```

## Builder

```go
result, err := mayfly.NewBuilder("gsasma").
    ForProblem(mayfly.Rastrigin, 20, -5.12, 5.12).
    WithIterations(500).
    WithPopulation(30, 30).
    WithConfig(func(config *mayfly.Config) {
        config.CoolingRate = 0.97
    }).
    Optimize()
```

| Method                               | Parameters and result                               |
| ------------------------------------ | --------------------------------------------------- |
| `ForProblem(fn, size, lower, upper)` | Set scalar objective, dimensions, and common bounds |
| `WithIterations(iterations)`         | Set `MaxIterations`                                 |
| `WithPopulation(males, females)`     | Set `NPop` and `NPopF`                              |
| `WithConfig(func(*Config))`          | Apply arbitrary configuration changes               |
| `Build()`                            | Validate builder basics and return `*Config`        |
| `Optimize()`                         | Build and run                                       |
| `GetVariant()`                       | Return the underlying variant                       |

Recognized names are `ma` (alias `standard`), `desma`, `olce` (alias
`olce-ma`), `eobbma`, `gsasma`, `mpma`, and `aoblmoa`. `NewBuilder` returns
`nil` for an unknown name, so validate user-provided names before chaining.

## Algorithm selection

```go
characteristics := mayfly.ProblemCharacteristics{
    Dimensionality:       30,
    Modality:             mayfly.HighlyMultimodal,
    Landscape:            mayfly.Rugged,
    ExpensiveEvaluations: true,
}

recommendation := mayfly.NewAlgorithmSelector().RecommendBest(characteristics)
result, err := mayfly.NewBuilderFromVariant(recommendation.Variant).
    ForProblem(objective, 30, -10, 10).
    Optimize()
```

`AlgorithmRecommendation` reports `Variant`, `Reasoning`, `Score`, and
`Confidence`. `ClassifyProblem(fn, size, lower, upper, rng)` estimates
characteristics from scale-free line scans; `rng` may be nil, and `Landscape`
comes back only as `Smooth` or `Rugged`. `RecommendForBenchmark(name)` handles the bundled
benchmark names.

## Comparison framework

```go
runner := mayfly.NewComparisonRunner().
    WithVariantNames("ma", "desma", "olce").
    WithRuns(30).
    WithIterations(500).
    WithSeed(42).
    WithParallel(true).
    WithMaxWorkers(4)

comparison, err := runner.CompareContext(
    ctx, "Rastrigin-30D", mayfly.Rastrigin, 30, -5.12, 5.12,
)
```

Use `CompareContext` for cancellation and error reporting. `Compare` is the
compatibility entry point that continues after individual run errors. Results
can be written with `WriteComparisonResults`, `ExportToCSV`, and
`ExportToJSON`. See the [comparison guide](comparison-framework.md) for all
runner controls and result statistics.

## Constraints and convergence

```go
target := 1e-8
config.Convergence = &mayfly.ConvergenceConfig{
    TargetCost:           &target,
    StagnationIterations: 100,
    MinImprovement:       1e-10,
    MinIterations:        200,
}
config.Constraints = &mayfly.ConstraintConfig{
    Inequalities: []mayfly.ConstraintFunction{
        func(x []float64) float64 { return x[0] + x[1] - 1 },
    },
}
```

Inequalities are feasible at `g(x) <= 0`; equalities are feasible within
`EqualityTolerance`. Feasibility ranking is the default. Penalty ranking uses
`ConstraintHandlingPenalty`, `PenaltyLinear` or `PenaltyQuadratic`, and a
positive `PenaltyFactor`.

## Configuration files and presets

| API                                       | Use                                                               |
| ----------------------------------------- | ----------------------------------------------------------------- |
| `LoadConfigFromFile(path)`                | Load and validate JSON; restore function and RNG fields afterward |
| `SaveConfigToFile(config, path)`          | Save serializable fields as JSON                                  |
| `ValidateConfig(config)`                  | Validate without running optimization                             |
| `ExportConfigTemplate(path, variant)`     | Write a commented template for a variant                          |
| `NewPresetConfig(preset)`                 | Create a problem-oriented preset                                  |
| `ListPresets()`                           | Return preset descriptions                                        |
| `AutoTuneConfig(config, characteristics)` | Apply the built-in tuning heuristics                              |

Available presets are `PresetUnimodal`, `PresetMultimodal`,
`PresetHighlyMultimodal`, `PresetDeceptive`, `PresetNarrowValley`,
`PresetHighDimensional`, `PresetFastConvergence`, `PresetStableConvergence`,
The former `PresetMultiObjective` is deprecated and returns an error because
the optimizer accepts one scalar objective.

## Bundled objective functions

All bundled functions have signature `func([]float64) float64`:

`Sphere`, `Rastrigin`, `Rosenbrock`, `Ackley`, `Griewank`, `Eggcrate`, `Beale`,
`Schwefel`, `Levy`, `Zakharov`, `Michalewicz`, `DixonPrice`, `BentCigar`,
`Discus`, `Weierstrass`, `HappyCat`, `ExpandedSchafferF6`, and `Himmelblau`.

Typical bounds and known minima are listed in the
[benchmark reference](../benchmarks.md).

## Lower-level utilities

The package also exports operators and components for custom research code:

- constraints: `EvaluateConstraints`, `IsFeasible`, `PenalizedCost`, and
  `BetterConstrainedCandidate`;
- genetic operators: `Crossover`, `Mutate`, `MutateGaussian`, `MutateCauchy`,
  and `HybridMutate`;
- variant components: `AnnealingScheduler`, `LogisticMap`, `ParetoArchive`,
  `ApplyOrthogonalLearning`, and `ApplyOrthogonalLearningToElite`.

Consult [pkg.go.dev](https://pkg.go.dev/github.com/cwbudde/mayfly) for exact
signatures and type-level documentation for these lower-level APIs.
