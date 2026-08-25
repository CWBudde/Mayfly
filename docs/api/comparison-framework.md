# Algorithm Comparison Framework

The comparison framework runs Mayfly variants repeatedly on the same problem,
aggregates their performance, applies paired statistical tests, and produces a
text report or machine-readable exports.

## Quick Start

```go
ctx := context.Background()
runner := mayfly.NewComparisonRunner().
    WithVariantNames("ma", "desma", "olce", "eobbma").
    WithRuns(30).
    WithIterations(500).
    WithSeed(42).
    WithParallel(true).
    WithMaxWorkers(4)

result, err := runner.CompareContext(
    ctx,
    "Rastrigin-30D",
    mayfly.Rastrigin,
    30,
    -5.12,
    5.12,
)
if err != nil {
    return err
}

if err := result.WriteComparisonResults(os.Stdout); err != nil {
    return err
}
```

The report includes aggregate statistics, rankings, a relative-quality bar
chart, significant pairwise Wilcoxon results, and the overall Friedman test.
Lower objective costs rank better.

## Configuring a Comparison

`NewComparisonRunner` compares all eight registered variants by default, using
30 runs, 500 iterations, sequential execution, and a time-based seed.

| Method                              | Purpose                                                                                     |
| ----------------------------------- | ------------------------------------------------------------------------------------------- |
| `WithVariants(...AlgorithmVariant)` | Select variants as objects                                                                  |
| `WithVariantNames(...string)`       | Select variants by registered name                                                          |
| `WithRuns(int)`                     | Set independent runs per variant                                                            |
| `WithIterations(int)`               | Set the iteration limit for every run                                                       |
| `WithMaxEvaluations(int)`           | Cap actual objective calls per run; zero disables the cap                                   |
| `WithTarget(float64)`               | Set a positive cost threshold used for success and convergence statistics; zero disables it |
| `WithSeed(int64)`                   | Set the base seed for reproducible, paired runs                                             |
| `WithParallel(bool)`                | Run independent optimization trials concurrently                                            |
| `WithMaxWorkers(int)`               | Bound concurrent trials; zero uses `runtime.NumCPU()`                                       |
| `WithVerbose(bool)`                 | Print completed-run progress to standard output                                             |

Recognized variant names are `ma`, `desma`, `olce` (or `olce-ma`), `eobbma`,
`gsasma`, `mpma`, and `aoblmoa`. `standard` is an alias for `ma`. Unknown names
are ignored by `WithVariantNames`; `CompareContext` returns an error if no valid
variant remains.

### Function-evaluation budgets

Paper comparisons often require the same number of objective evaluations even
when algorithms evaluate different numbers of candidates per iteration. Use an
evaluation budget together with an iteration safety ceiling:

```go
runner.WithMaxEvaluations(95_000).WithIterations(2_000)
```

Each run invokes the objective no more than the configured budget and reports
the actual invocation count in `RunResult.FuncEvals`. If the iteration ceiling
is reached before the complete budget is consumed, `CompareContext` returns an
error instead of silently accepting an incomparable run. If the budget ends
partway through a generation, remaining candidates receive an infinite cost
without invoking the objective, preserving the best solution found within the
exact budget. Set `MaxIterations` close to the expected final generation to
avoid unnecessary post-budget work.

### Parallelism and objective safety

Comparison parallelism schedules complete optimization trials concurrently and
is separate from `Config.EnableParallel`, which parallelizes fitness batches
inside one optimization. `ComparisonRunner` uses each variant's default config,
so internal fitness parallelism remains disabled. Avoid enabling both layers at
high worker counts unless the extra CPU demand is intentional.

When comparison parallelism is enabled, the objective function can be called
concurrently by different trials. Custom objectives must therefore be safe for
concurrent use. Each trial receives its own random generator, and all variants
use the same derived seed for a given run index. A fixed base seed makes results
independent of goroutine scheduling and preserves pairing for statistical tests.

## Running Comparisons

Prefer `CompareContext` for new code:

```go
result, err := runner.CompareContext(ctx, name, objective, dimensions, lower, upper)
```

It validates the runner and problem, supports cancellation, and returns an
error if a trial fails. No partial aggregate is returned after a trial failure.
In-flight objective calls are allowed to finish before cancellation completes.

The compatibility method has the original signature:

```go
result := runner.Compare(name, objective, dimensions, lower, upper)
```

It uses a background context and continues after individual trial failures.
Failed trials are represented by an infinite cost and an error string in their
`RunResult`. Because it cannot return an error, use `CompareContext` when input,
cancellation, or optimization failures must be handled explicitly.

## Results

`ComparisonResult` keeps all slices aligned by algorithm index:

```go
type ComparisonResult struct {
    BenchmarkName  string
    AlgorithmNames []string
    RunResults     [][]RunResult
    Statistics     []AlgorithmStatistics
    Rankings       []int
    WilcoxonTests  [][]WilcoxonResult
    FriedmanResult *FriedmanTestResult
    BestAlgorithm  int
    BaseSeed       int64
}
```

`RunResults[algorithm][run]` records the best cost, function evaluations,
iterations, target-convergence iteration, execution time, seed, and any error.
The matching `Statistics` entry contains mean, median, standard deviation, best
and worst cost, success rate, average evaluation count, and average execution
time. `Rankings` uses `1` for the lowest mean cost, while `BestAlgorithm` is the
index of that entry.

The Wilcoxon matrix compares paired run results for each algorithm pair. The
Friedman result summarizes whether the algorithms differ overall. Both report
significance at alpha 0.05.

## Reporting and Export

Write a report to any `io.Writer` when errors need to be handled:

```go
if err := result.WriteComparisonResults(os.Stdout); err != nil {
    return err
}
```

`result.PrintComparisonResults()` is a convenience wrapper that writes the same
report to standard output and discards write errors.

Export detailed results for further analysis:

```go
if err := result.ExportToCSV("comparison_results.csv"); err != nil {
    return err
}
if err := result.ExportToJSON("comparison_results.json"); err != nil {
    return err
}
```

CSV contains one row per algorithm run together with its aggregate statistics.
JSON contains the complete `ComparisonResult`. Both methods create or truncate
the destination file and return filesystem or encoding errors.

## Complete Parallel Example

The runnable example compares all eight variants on Rastrigin, limits the number
of concurrent trials, handles Ctrl+C cancellation, prints the report and chart,
and exports CSV and JSON:

```bash
cd examples/comparison
go run .
```

For statistically defensible studies, use at least 20 paired runs and choose an
iteration budget appropriate to the problem. Execution time is useful for
operational comparisons, but it naturally varies with machine load and the
selected worker count.

## Related Documentation

- [Configuration Guide](configuration.md)
- [Unified Framework](unified-framework.md)
- [Algorithm Variants](../algorithms/)
