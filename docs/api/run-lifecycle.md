# Run Lifecycle API

`Optimize` remains the simplest entry point. Use `OptimizeContext` when a run
needs cancellation, progress snapshots, structured logging, or a partially
supplied initial population. Completed results can export their convergence
curves for later analysis.

For a compact map of all application-level APIs, see the
[API quick reference](quick-reference.md).

## Cancellation and progress

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

result, err := mayfly.OptimizeContext(
    ctx,
    config,
    mayfly.WithProgressObserver(func(progress mayfly.Progress) {
        log.Printf(
            "iteration=%d evaluations=%d cost=%g",
            progress.Iteration,
            progress.EvaluationCount,
            progress.Best.Cost,
        )
    }),
)
```

The context is checked while initialization, population, crossover, mutation,
and variant-specific evaluation batches are dispatched, during parallel DESMA
candidate construction and MPMA median calculation, and at iteration
boundaries. An objective function that is already running is not interrupted;
the optimizer waits for in-flight calls before returning the cancellation
error. The observer runs synchronously after every completed iteration.
`Progress.Iteration` is one-based.

Each progress value owns its `Best.Position` slice. It is safe for an observer
to retain or modify the snapshot without changing the optimizer.

## Structured logging

`WithLogger` accepts any value implementing `mayfly.Logger`. A standard
`*slog.Logger` works directly:

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

result, err := mayfly.OptimizeContext(
    ctx,
    config,
    mayfly.WithLogger(logger),
)
```

Logging is opt-in and synchronous. Mayfly emits info-level events with stable
`event` attributes:

| Event                    | Important attributes                                                                   |
| ------------------------ | -------------------------------------------------------------------------------------- |
| `optimization_started`   | `problem_size`, `max_iterations`, `male_population`, `female_population`, `parallel`   |
| `iteration_completed`    | `iteration`, `evaluations`, `best_cost`, `constraint_violation`                        |
| `optimization_completed` | `iterations`, `evaluations`, `best_cost`, `constraint_violation`, `termination_reason` |

The logger and progress observer can be enabled together. A nil logger or
observer disables that output.

## Convergence curve export

Every successful result can export one sample per completed iteration:

```go
if err := result.ExportConvergenceCSV("convergence.csv"); err != nil {
    return err
}

if err := result.ExportConvergenceJSON("convergence.json"); err != nil {
    return err
}
```

CSV output has `iteration,best_cost` columns. JSON output is an array of
objects with `iteration` and `best_cost` fields. Iterations are one-based in
both formats, and both exporters preserve the full precision of the result's
`ConvergenceCurve` values.

## Initial population

Seed male and female populations independently:

```go
maleSeeds := [][]float64{
    savedBest,
    perturbedBest,
}
femaleSeeds := [][]float64{
    anotherCandidate,
}

result, err := mayfly.OptimizeContext(
    ctx,
    config,
    mayfly.WithInitialPopulation(maleSeeds, femaleSeeds),
)
```

A seed list may be empty or shorter than its configured population. Mayfly
fills the remaining positions with the run's random number generator. Supplying
more male positions than `Config.NPop`, or more female positions than
`Config.NPopF`, is an error. Every position must have exactly
`Config.ProblemSize` finite coordinates within the inclusive configured
bounds.

`WithInitialPopulation` snapshots both nested slices when the option is
created. Later changes to the caller's slices do not affect the run.
