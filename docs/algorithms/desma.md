# DESMA - Dynamic Elite Strategy Mayfly Algorithm

## Research Reference

**Dynamic elite strategy mayfly algorithm. PLOS One, 2022.**

## Overview

DESMA (Dynamic Elite Strategy Mayfly Algorithm) is an improved variant that addresses local optima trapping and slow convergence through adaptive elite generation around the current best position.

## Key Innovations

### Dynamic Elite Generation

After selection, DESMA generates `EliteCount` (default: 5) candidate solutions around the global best within a dynamic `SearchRange`. If any elite solution is better than the worst male, it replaces it.

**Implementation** (mayfly.go:473-510 and generateEliteMayflies() at line 543):

```go
// Generate elite solutions near global best
elites := generateEliteMayflies(config, result.GlobalBest, searchRange)

// Replace worst male if elite is better
for _, elite := range elites {
    if elite.Cost < males[NPop-1].Cost {
        males[NPop-1] = elite
        sortMayflies(males)
    }
}
```

### Adaptive Search Range

The search range adapts based on improvement:

- **If improving**: `SearchRange *= EnlargeFactor` (default 1.05)
- **If stagnating**: `SearchRange *= ReductionFactor` (default 0.95)

This creates a balance between exploration (large range) and exploitation (small range).

## Usage Example

```go
package main

import (
    "fmt"
    "github.com/CWBudde/mayfly"
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
- `EliteCount`: Number of elite mayflies to generate per iteration (default: 5)
- `SearchRange`: Initial search range for elite generation (default: auto-calculated as 10% of search space)
- `EnlargeFactor`: Factor to enlarge search range when improving (default: 1.05)
- `ReductionFactor`: Factor to reduce search range when not improving (default: 0.95)

## Benefits

- **Better convergence**: Escapes local optima more effectively
- **Adaptive search**: Dynamically adjusts search range based on improvement
- **Faster optimization**: Often achieves better results with the same number of iterations
- **Minimal overhead**: Only 5-10% more function evaluations (typically ~8%)

## Performance

### Multimodal Functions

DESMA excels on multimodal functions with 70%+ improvement over standard MA:

**Rastrigin Function (highly multimodal)**:
- Standard MA: 45-60 (typical result)
- DESMA: 15-30 (70%+ improvement)

**Rosenbrock Function (narrow valley)**:
- Standard MA: 10-50
- DESMA: 0.1-5 (significant improvement)

### Function Evaluation Overhead

DESMA uses approximately 8% more function evaluations than Standard MA:
- Standard MA: ~30,540 evaluations (500 iterations, pop=20)
- DESMA: ~33,000 evaluations (includes elite generation)

The small overhead is worthwhile for the significant quality improvement.

## When to Use DESMA

- **Best for**: Multimodal optimization problems with many local optima
- **Excellent on**: Rastrigin, Rosenbrock, Griewank, Ackley functions
- **Use when**: Standard MA gets trapped in local optima
- **Ideal for**: Problems where solution quality matters more than minimal evaluations

## Algorithm Workflow (Addition to Standard MA)

After the standard MA selection step:

1. **Generate Elite Solutions**: Create `EliteCount` solutions around global best
2. **Evaluate Elites**: Calculate fitness for each elite solution
3. **Replacement**: Replace worst male if any elite is better
4. **Adapt Search Range**:
   - If global best improved → increase range (exploration)
   - If stagnating → decrease range (exploitation)

## Parameter Tuning Guide

### Elite Count

**Default (balanced)**:
```go
config.EliteCount = 5
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

**Auto-calculated (recommended)**:
```go
// Leave SearchRange at 0 for automatic calculation
config.SearchRange = 0  // Auto: 10% of (UpperBound - LowerBound)
```

**Custom range**:
```go
config.SearchRange = 2.0  // Fixed range of ±2.0
```
- Use when: You know the optimal search radius
- Trade-off: Less adaptive behavior

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
