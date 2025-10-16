# Algorithm Comparison Framework

Statistical comparison tools for evaluating and comparing algorithm variants.

## Overview

The comparison framework provides tools to:
- Run multiple algorithms on the same problem
- Collect statistical data across multiple runs
- Perform statistical significance tests
- Analyze convergence behavior
- Generate comprehensive reports

## Quick Start

```go
runner := mayfly.NewComparisonRunner().
    WithVariantNames("ma", "desma", "olce", "eobbma").
    WithRuns(30).  // 30 runs for statistical significance
    WithIterations(500).
    WithVerbose(true)

result := runner.Compare(
    "Rastrigin",
    mayfly.Rastrigin,
    30,      // problem size
    -5.12, 5.12,
)

// Print comprehensive statistical analysis
result.PrintComparisonResults()
```

## ComparisonRunner API

### Constructor

```go
runner := mayfly.NewComparisonRunner()
```

### Configuration Methods

| Method | Parameters | Description |
|--------|------------|-------------|
| `WithVariantNames(...string)` | variant names | Set algorithms to compare |
| `WithVariants(...AlgorithmVariant)` | variant objects | Set algorithms using variant objects |
| `WithRuns(n int)` | number of runs | Set number of independent runs (default: 30) |
| `WithIterations(n int)` | max iterations | Set iterations per run (default: 500) |
| `WithVerbose(v bool)` | verbose flag | Enable/disable progress output |
| `WithSeed(s int64)` | random seed | Set seed for reproducibility |

### Comparison Methods

#### Compare Single Problem

```go
result := runner.Compare(
    name string,           // Problem name (for reporting)
    fn ObjectiveFunction,  // Function to optimize
    problemSize int,       // Number of dimensions
    lower, upper float64,  // Search bounds
)
```

#### Compare Multiple Problems

```go
problems := []mayfly.BenchmarkProblem{
    {Name: "Sphere", Func: mayfly.Sphere, Size: 30, Lower: -10, Upper: 10},
    {Name: "Rastrigin", Func: mayfly.Rastrigin, Size: 30, Lower: -5.12, Upper: 5.12},
}

results := runner.CompareMultiple(problems)
```

## ComparisonResult

The result object contains comprehensive statistical data:

### Fields

```go
type ComparisonResult struct {
    ProblemName string
    Variants    []string
    Statistics  map[string]AlgorithmStats
    Rankings    []AlgorithmRanking

    // Statistical tests
    WilcoxonTests  map[string]float64  // Pairwise p-values
    FriedmanTest   FriedmanResult
}
```

### AlgorithmStats

Statistics for each algorithm:

```go
type AlgorithmStats struct {
    Mean              float64
    Median            float64
    StdDev            float64
    Best              float64
    Worst             float64
    SuccessRate       float64  // % runs below threshold
    AvgConvergenceIter int     // Avg iterations to converge
    AllResults        []float64
}
```

### Methods

```go
// Print formatted results
result.PrintComparisonResults()

// Print summary table only
result.PrintSummary()

// Print statistical tests
result.PrintStatisticalTests()

// Print convergence analysis
result.PrintConvergenceAnalysis()

// Get best algorithm
best := result.GetBestAlgorithm()  // Based on mean cost
```

## Statistical Analysis

### Wilcoxon Signed-Rank Test

Pairwise comparisons between algorithms:

```go
result.PrintStatisticalTests()
```

**Output**:
```
Wilcoxon Signed-Rank Tests (p-values):
  DESMA vs MA:     0.0023 ** (significant)
  OLCE vs MA:      0.0001 *** (highly significant)
  EOBBMA vs DESMA: 0.1234 (not significant)
```

**Interpretation**:
- p < 0.01: Highly significant difference (***)
- p < 0.05: Significant difference (**)
- p < 0.10: Marginally significant (*)
- p ≥ 0.10: Not significant

### Friedman Test

Overall test for differences across all algorithms:

```go
fmt.Printf("Friedman Test: χ² = %.2f, p = %.4f\n",
    result.FriedmanTest.ChiSquare,
    result.FriedmanTest.PValue)
```

**Interpretation**:
- p < 0.05: At least one algorithm is significantly different
- p ≥ 0.05: No significant differences detected

## Convergence Analysis

Track how algorithms converge over iterations:

```go
runner := mayfly.NewComparisonRunner().
    WithVariantNames("ma", "desma").
    WithRuns(30).
    WithIterations(1000).
    WithConvergenceTracking(true)  // Enable convergence tracking

result := runner.Compare("Rastrigin", mayfly.Rastrigin, 30, -5.12, 5.12)

// Get convergence curves
curves := result.GetConvergenceCurves()

// curves[variant][iteration] = average cost at that iteration
for variant, curve := range curves {
    fmt.Printf("%s: Final avg cost = %.4f\n", variant, curve[len(curve)-1])
}
```

## Example Output

### Summary Table

```
Algorithm Comparison: Rastrigin (D=30, 30 runs)
─────────────────────────────────────────────────────────────
Algorithm    Mean      Median    Std Dev   Best      Worst
─────────────────────────────────────────────────────────────
OLCE-MA      28.32     27.45     4.23      18.92     38.15
DESMA        35.67     34.89     5.67      24.31     48.22
EOBBMA       38.12     37.54     6.12      26.78     52.34
MA           45.23     44.67     7.89      32.45     62.11
─────────────────────────────────────────────────────────────
```

### Statistical Rankings

```
Rankings (by mean cost):
1. OLCE-MA    (mean: 28.32, rank: 1.2)
2. DESMA      (mean: 35.67, rank: 2.3)
3. EOBBMA     (mean: 38.12, rank: 2.8)
4. MA         (mean: 45.23, rank: 3.7)
```

### Convergence Analysis

```
Convergence to cost < 40.0:
  OLCE-MA:  avg 287 iterations (25% faster than MA)
  DESMA:    avg 312 iterations (18% faster than MA)
  MA:       avg 382 iterations (baseline)
```

## Advanced Usage

### Custom Success Threshold

Define what counts as "success":

```go
runner := mayfly.NewComparisonRunner().
    WithVariantNames("ma", "desma").
    WithSuccessThreshold(10.0)  // Cost < 10 = success

result := runner.Compare("Sphere", mayfly.Sphere, 30, -10, 10)
fmt.Printf("Success rate: %.1f%%\n", result.Statistics["desma"].SuccessRate)
```

### Parallel Execution

Run comparisons in parallel for speed:

```go
runner := mayfly.NewComparisonRunner().
    WithVariantNames("ma", "desma", "olce").
    WithRuns(30).
    WithParallel(true)  // Enable parallel execution

result := runner.Compare("Rastrigin", mayfly.Rastrigin, 30, -5.12, 5.12)
```

### Export Results

Export to CSV or JSON:

```go
// Export to CSV
err := result.ExportToCSV("results.csv")

// Export to JSON
err := result.ExportToJSON("results.json")
```

## Benchmark Suite Example

Complete example in `examples/benchmark_suite/main.go`:

```go
package main

import (
    "fmt"
    "github.com/CWBudde/mayfly"
)

func main() {
    runner := mayfly.NewComparisonRunner().
        WithVariantNames("ma", "desma", "olce", "eobbma", "gsasma", "mpma").
        WithRuns(30).
        WithIterations(500).
        WithVerbose(true)

    problems := []mayfly.BenchmarkProblem{
        {Name: "Sphere", Func: mayfly.Sphere, Size: 30, Lower: -10, Upper: 10},
        {Name: "Rastrigin", Func: mayfly.Rastrigin, Size: 30, Lower: -5.12, Upper: 5.12},
        {Name: "Rosenbrock", Func: mayfly.Rosenbrock, Size: 30, Lower: -5, Upper: 10},
        {Name: "Schwefel", Func: mayfly.Schwefel, Size: 30, Lower: -500, Upper: 500},
    }

    results := runner.CompareMultiple(problems)

    for _, result := range results {
        fmt.Printf("\n=== %s ===\n", result.ProblemName)
        result.PrintComparisonResults()
    }
}
```

## Performance Metrics

The framework tracks multiple performance metrics:

| Metric | Description | Use Case |
|--------|-------------|----------|
| **Mean** | Average final cost | Overall quality |
| **Median** | Middle value | Robust central tendency |
| **Std Dev** | Variance | Algorithm stability |
| **Best** | Best run result | Peak performance |
| **Worst** | Worst run result | Worst-case behavior |
| **Success Rate** | % below threshold | Reliability |
| **Convergence** | Iterations to target | Speed |

## Statistical Significance

### Choosing Sample Size

For statistical validity:
- **Minimum**: 20 runs
- **Recommended**: 30 runs (standard)
- **High confidence**: 50+ runs

### Interpreting p-values

- **p < 0.001**: Very strong evidence
- **p < 0.01**: Strong evidence
- **p < 0.05**: Moderate evidence
- **p < 0.10**: Weak evidence
- **p ≥ 0.10**: Insufficient evidence

### Effect Size

Consider practical significance:
- Small difference but p < 0.05: Statistically significant but may not matter
- Large difference but p > 0.05: Practically important but needs more data

## Related Documentation

- [Configuration Guide](configuration.md) - Parameter reference
- [Unified Framework](unified-framework.md) - Builder API and selection
- [Algorithm Variants](../algorithms/) - Individual algorithm docs
