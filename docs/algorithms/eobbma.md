# EOBBMA - Elite Opposition-Based Bare Bones Mayfly Algorithm

## Research Reference

**Elite Opposition-Based Bare Bones Mayfly Algorithm (2024). Arabian Journal for Science and Engineering.**

## Overview

EOBBMA replaces traditional velocity-based updates with Gaussian sampling and introduces Lévy flight for exploration. This "bare bones" approach provides excellent exploration-exploitation balance on complex, deceptive landscapes.

## Key Innovations

### 1. Bare Bones Framework

The Bare Bones approach eliminates velocity-based updates in favor of **Gaussian sampling**:

**Males**: Sample new positions from Gaussian distributions centered between current position and personal/global best

**Females**: Sample from Gaussian around best males or use Lévy flight

**Mathematical Foundation**:
```
X_new = N(μ, σ²)
where:
  μ = (X_current + X_best) / 2
  σ = |X_current - X_best| / 2
```

**Benefits**:
- Fewer parameters to tune
- More intuitive exploration behavior
- No velocity limits needed
- Automatic adaptation based on distance to best

### 2. Lévy Flight Distribution

**Lévy flights** generate heavy-tailed random jumps using Mantegna's algorithm:

- **Heavy tails**: Occasional large jumps help escape local optima
- **Stability parameter (α)**: Controls tail heaviness (default: 1.5)
- **Scale parameter (β)**: Controls jump magnitude (default: 1.0)

**What are Lévy Flights?**

Unlike normal random walks (Gaussian), Lévy flights produce a mix of many small steps and occasional very large jumps. This mimics foraging patterns in nature (albatross, honeybees) and is highly effective for global optimization.

**Visual comparison**:
```
Gaussian walk:    ○○○○○○○○○○○○○     (consistent small steps)
Lévy flight:      ○○○○○────────○○   (small steps + rare jumps)
```

### 3. Elite Opposition-Based Learning

**Opposition-based learning** explores the opposite side of the search space:

- For each elite solution, generate its **opposition point**: `x_opp = a + b - x`
- If opposition point is better, replace the elite
- Expands search coverage without additional population

**Example**: If elite is at x=7 in bounds [0,10], opposition point is at x=3

## Usage Example

```go
package main

import (
    "fmt"
    "github.com/CWBudde/mayfly"
)

func main() {
    // Use EOBBMA for complex, deceptive optimization landscapes
    config := mayfly.NewEOBBMAConfig()
    config.ObjectiveFunc = mayfly.Schwefel  // Highly deceptive function
    config.ProblemSize = 10
    config.LowerBound = -500
    config.UpperBound = 500
    config.MaxIterations = 500

    result, err := mayfly.Optimize(config)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Best Cost: %f\n", result.GlobalBest.Cost)
}
```

## EOBBMA Parameters

- `UseEOBBMA`: Enable EOBBMA variant (default: false)
- `LevyAlpha`: Lévy stability parameter (default: 1.5, range: 0 < α ≤ 2)
- `LevyBeta`: Lévy scale parameter (default: 1.0)
- `OppositionRate`: Probability of applying opposition learning (default: 0.3)
- `EliteOppositionCount`: Number of elite solutions to oppose (default: 3)

## Benefits

- **55%+ improvement** on deceptive functions (Schwefel, complex landscapes)
- **Better exploration**: Lévy flights enable efficient global search
- **Simpler tuning**: Fewer parameters than velocity-based approaches
- **Robust**: Works well across different problem types
- **Low overhead**: Comparable function evaluations to Standard MA

## Performance

### Deceptive Functions

**Schwefel (D=10, highly deceptive)**:
- Standard MA: 789.59 (~30,540 evals)
- EOBBMA: 355.32 (~31,011 evals)
- **Improvement: 55.00%**

**Michalewicz (D=10, steep valleys)**:
- Standard MA: -8.5 to -9.0
- EOBBMA: -9.3 to -9.6
- **Significantly better convergence**

## When to Use EOBBMA

- **Best for**: Highly deceptive functions with misleading gradients
- **Excellent on**: Problems where other algorithms plateau early
- **Use when**: Search space has many local optima at different scales
- **Ideal for**: Schwefel, Michalewicz, complex engineering problems

## Parameter Tuning Guide

### Lévy Alpha (Stability Parameter)

**Default (heavy-tailed)**:
```go
config.LevyAlpha = 1.5
```
- Produces good balance of small steps and large jumps
- Recommended for most problems

**More heavy-tailed (α < 1.5)**:
```go
config.LevyAlpha = 1.0
```
- Use when: Need more frequent large jumps
- Use when: Problem requires aggressive global exploration
- Caution: May converge slower

**Less heavy-tailed (α > 1.5)**:
```go
config.LevyAlpha = 1.8
```
- Use when: Want more stable, Gaussian-like behavior
- Use when: Solutions need refinement
- Note: α=2.0 is pure Gaussian (no heavy tails)

### Lévy Beta (Scale Parameter)

**Default**:
```go
config.LevyBeta = 1.0
```
- Good baseline scale

**Larger jumps**:
```go
config.LevyBeta = 2.0
```
- Use when: Search space is very large
- Use when: Need long-range exploration

**Smaller jumps**:
```go
config.LevyBeta = 0.5
```
- Use when: Search space is small
- Use when: Want more local exploration

### Opposition Rate

**Default (moderate)**:
```go
config.OppositionRate = 0.3  // 30% probability
```
- Balanced opposition application

**More opposition**:
```go
config.OppositionRate = 0.5  // 50% probability
```
- Use when: Search space is large and sparsely sampled
- Caution: Higher computational cost

**Less opposition**:
```go
config.OppositionRate = 0.1  // 10% probability
```
- Use when: Function evaluations are expensive
- Trade-off: Less search coverage

### Elite Opposition Count

**Default**:
```go
config.EliteOppositionCount = 3
```
- Applies to top 3 solutions

**More elites**:
```go
config.EliteOppositionCount = 5
```
- Better coverage of elite region opposites
- Higher computational cost

**Fewer elites**:
```go
config.EliteOppositionCount = 1
```
- Only global best gets opposed
- Minimal overhead

## EOBBMA vs Other Variants

**Choose EOBBMA when**:
- Problem is highly deceptive (Schwefel-like)
- Want simplest parameter tuning
- Heavy-tailed jumps are beneficial
- Other velocity-based algorithms struggle

**Choose OLCE-MA instead when**:
- Problem is highly multimodal but not deceptive
- Orthogonal learning benefits parameter space exploration
- Chaotic perturbations are effective

**Choose GSASMA instead when**:
- Need maximum convergence speed
- Simulated annealing fits problem structure
- Prefer gradual exploration-exploitation transition

**Choose DESMA instead when**:
- Want velocity-based framework
- Need adaptive elite local search
- Simpler local optima escape is sufficient

## Lévy Flight Details

### Mantegna's Algorithm

EOBBMA uses Mantegna's algorithm to generate Lévy-distributed random numbers:

```go
u = N(0, σ_u²)  // Gaussian with variance σ_u
v = N(0, σ_v²)  // Gaussian with variance σ_v

levy = u / |v|^(1/α)

where:
σ_u = [Γ(1+α) * sin(πα/2) / (Γ((1+α)/2) * α * 2^((α-1)/2))]^(1/α)
σ_v = 1
```

### Lévy Step Formula

```go
step = 0.01 * levy * (X_current - X_best)
X_new = X_current + LevyBeta * step
```

## Related Documentation

- [DESMA](desma.md) - Velocity-based elite variant
- [OLCE-MA](olce-ma.md) - Orthogonal and chaotic alternative
- [Standard MA](standard-ma.md) - Base algorithm
- [Configuration Guide](../api/configuration.md) - Complete parameter reference
