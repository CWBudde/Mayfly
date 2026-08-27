# DESMA - Dynamic Elite Strategy Mayfly Algorithm

## Research Reference

Qianhang Du and Honghao Zhu, **Dynamic elite strategy mayfly algorithm**,
PLOS ONE 17(8), 2022, e0273155,
[doi:10.1371/journal.pone.0273155](https://doi.org/10.1371/journal.pone.0273155).

## Overview

DESMA (Dynamic Elite Strategy Mayfly Algorithm) is an improved variant that addresses local optima trapping and slow convergence through adaptive elite generation around the current best position.

## Key Innovations

### Dynamic Elite Generation

The paper generates `k` candidates around the current global best:

```text
r1     = 2*rand(1,n) - 1
egbest = cgbest + r1*R
```

It clips each coordinate to the problem bounds, selects the best candidate,
and describes replacing the current global-best mayfly only when that candidate
improves it. The next male-velocity update then uses `egbest`.

The current library generates `EliteCount` candidates after survivor selection
and inserts an improving candidate at the worst-male slot. This retains the old
best population member, unlike the published replacement description. That
population-insertion choice is a documented compatibility extension pending an
authoritative implementation or clarification.

### Adaptive Search Range

The search range adapts based on improvement:

- **If improving**: `SearchRange *= EnlargeFactor` (default 1.05)
- **If stagnating**: `SearchRange *= ReductionFactor` (default 0.95)

This creates a balance between exploration (large range) and exploitation (small range).

The paper specifies the two multipliers but not the initial value of `R`.
Mayfly's automatic initial value, 10% of the search-space span, is a library
default rather than a value recovered from the paper.

## Usage Example

```go
package main

import (
    "fmt"
    "github.com/cwbudde/mayfly"
)

func main() {
    // Use DESMA for better performance
    config := mayfly.NewDESMAConfig()
    config.ObjectiveFunc = mayfly.Rastrigin
    config.ProblemSize = 50
    config.LowerBound = -10
    config.UpperBound = 10
    config.MaxIterations = 1000

    result, err := mayfly.Optimize(config)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Best Cost: %f\n", result.GlobalBest.Cost)
}
```

## DESMA-Specific Parameters

- `UseDESMA`: Enable DESMA variant (default: false)
- `EliteCount`: Number of elite mayflies to generate per iteration (default: 10)
- `SearchRange`: Initial search range for elite generation (library default: 10% of the search-space span; not reported by the paper)
- `EnlargeFactor`: Factor to enlarge search range when improving (default: 1.05)
- `ReductionFactor`: Factor to reduce search range when not improving (default: 0.95)

## Published experiment and fidelity status

The paper evaluates 28 CEC2013 functions in 30 dimensions, using 51 independent
runs and at most 300,000 function evaluations. It reports population size 50,
`k = 10`, `g_max = 0.9`, `g_min = 0.4`, `a1 = 1`, `a2 = 1.5`, dance/attenuation
`5/0.8`, flight/attenuation `1/0.99`, and radius factors `1.05/0.95`. DESMA's
Table 3 average rank is 2.57, first among the eight algorithms in that table.

The complete CEC2013 function metadata and DESMA Table 3 mean-error/rank column
are transcribed in the
[machine-readable reference artifact](../reference-data/desma-2022-table3.json).
It is deliberately marked `reproduction_claim: false`: the paper does not give
the initial radius, clarify whether population 50 is per sex or combined, or
publish the full base-MA operator settings, seeds, or raw runs. Its supplement
contains CEC2013 evaluator/input data but no DESMA implementation. The library
does not yet expose that CEC2013 suite, so the paper's exact benchmark set
cannot currently be selected by the reproduction command.

Two additional implementation gates remain. The paper's crossover uses `L` in
`[-1,1]`, whereas the current generic BLX operator uses `[-0.4,1.4]` at its
default `gamma = 0.4`. Equation 16 also prints a minimization-direction
condition that conflicts with the surrounding prose. These must be resolved
rather than silently normalized for an exact preset.

Evaluation overhead is configuration-dependent. Each iteration adds up to
`EliteCount` objective calls; the resulting percentage depends on population,
crossover, mutation, stopping, and no-op behavior. The paper does not support a
universal “8% overhead” value.

## When to Use DESMA

- **Designed for**: Search landscapes where ordinary MA becomes trapped
- **Published evaluation**: The shifted/rotated CEC2013 suite, not the classic
  unshifted functions used by Mayfly's default reproduction harness
- **Use when**: The adaptive elite search is worth its additional objective calls
- **Measure first**: The paper does not establish a universal improvement for a
  named classic function or arbitrary configuration

## Algorithm Workflow (Addition to Standard MA)

In the current library, after the standard MA selection step:

1. **Adapt Search Range**:
   - If global best improved → increase range (exploration)
   - If stagnating → decrease range (exploitation)
2. **Generate Elite Solutions**: Create `EliteCount` solutions around global best
3. **Evaluate Elites**: Calculate fitness for each elite solution
4. **Library replacement**: Insert the best improving elite at the worst-male slot

## Parameter Tuning Guide

### Elite Count

**Default (balanced)**:

```go
config.EliteCount = 10
```

- Good balance between exploration and computational cost

**More exploration**:

```go
config.EliteCount = 10
```

- Use when: Problem has many local optima
- Trade-off: Higher computational cost

**Less overhead**:

```go
config.EliteCount = 3
```

- Use when: Function evaluations are expensive
- Trade-off: Less intensive local search

### Search Range

**Auto-calculated library default**:

```go
// Leave SearchRange at 0 for automatic calculation
config.SearchRange = 0  // Auto: 10% of (UpperBound - LowerBound)
```

**Custom range**:

```go
config.SearchRange = 2.0  // Initial range of ±2.0; still adapted each iteration
```

- Use when: You know an appropriate initial radius

### Adaptation Factors

**More aggressive adaptation**:

```go
config.EnlargeFactor = 1.1    // Faster exploration increase
config.ReductionFactor = 0.90  // Faster exploitation focus
```

**More conservative adaptation**:

```go
config.EnlargeFactor = 1.02   // Slower exploration increase
config.ReductionFactor = 0.98  // Slower exploitation focus
```

## DESMA vs Other Variants

**Choose DESMA when**:

- Problem has many local optima
- You want adaptive local search
- Function evaluations are cheap
- Standard MA plateaus early

**Choose OLCE-MA instead when**:

- Problem is highly multimodal (Rastrigin-like)
- High dimensionality (20D+)
- Need systematic parameter space exploration

**Choose EOBBMA instead when**:

- Problem is highly deceptive (Schwefel-like)
- Want simpler parameter tuning
- Heavy-tailed jumps are beneficial

## Related Documentation

- [Standard MA](standard-ma.md) - Base algorithm
- [OLCE-MA](olce-ma.md) - For highly multimodal problems
- [Configuration Guide](../api/configuration.md) - Complete parameter reference
