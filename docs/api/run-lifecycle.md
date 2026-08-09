# Run Lifecycle API

`Optimize` remains the simplest entry point. Use `OptimizeContext` when a run
needs cancellation, progress snapshots, or a partially supplied initial
population.

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

The context is checked before every initialization evaluation and at both ends
of each iteration. An objective function that is already running is not
interrupted; cancellation takes effect at the next check. The observer runs
synchronously after every completed iteration. `Progress.Iteration` is
one-based.

Each progress value owns its `Best.Position` slice. It is safe for an observer
to retain or modify the snapshot without changing the optimizer.

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
