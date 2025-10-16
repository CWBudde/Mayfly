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

## Usage Examples

### Basic Usage

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

### Advanced Usage with Lévy Flight Tuning

```go
package main

import (
    "fmt"
    "github.com/CWBudde/mayfly"
)

func main() {
    // Configure EOBBMA for deceptive landscape with custom Lévy parameters
    config := mayfly.NewEOBBMAConfig()
    config.ObjectiveFunc = mayfly.Schwefel
    config.ProblemSize = 30
    config.LowerBound = -500
    config.UpperBound = 500
    config.MaxIterations = 1000

    // Tune Lévy flight parameters
    config.LevyAlpha = 1.3    // More heavy-tailed for aggressive exploration
    config.LevyBeta = 1.5      // Larger jumps for wide search space

    // Aggressive opposition learning
    config.OppositionRate = 0.4          // 40% probability
    config.EliteOppositionCount = 5       // Apply to top 5 solutions

    // Larger population for complex landscape
    config.NPop = 40
    config.NPopF = 40

    result, err := mayfly.Optimize(config)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Final Cost: %.2f\n", result.GlobalBest.Cost)
    fmt.Printf("Iterations: %d\n", result.IterationCount)
    fmt.Printf("Function Evaluations: %d\n", result.FuncEvalCount)

    // Show first few dimensions of solution
    fmt.Printf("Solution (first 5 dims): ")
    for i := 0; i < min(5, len(result.GlobalBest.Position)); i++ {
        fmt.Printf("%.2f ", result.GlobalBest.Position[i])
    }
    fmt.Println()
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
```

### Real-World Example: Portfolio Optimization with Risk

```go
package main

import (
    "fmt"
    "math"
    "github.com/CWBudde/mayfly"
)

// Portfolio optimization: maximize return while managing risk
// This creates a deceptive landscape due to risk-return tradeoffs
func portfolioObjective(allocation []float64) float64 {
    // Simulated asset returns (annual %)
    returns := []float64{0.12, 0.08, 0.15, 0.06, 0.18, 0.10}

    // Simulated asset volatilities (risk)
    volatilities := []float64{0.20, 0.10, 0.25, 0.05, 0.30, 0.12}

    // Correlation effects (diversification)
    correlations := [][]float64{
        {1.0, 0.3, 0.5, 0.1, 0.6, 0.2},
        {0.3, 1.0, 0.2, 0.4, 0.3, 0.5},
        {0.5, 0.2, 1.0, 0.1, 0.7, 0.3},
        {0.1, 0.4, 0.1, 1.0, 0.2, 0.6},
        {0.6, 0.3, 0.7, 0.2, 1.0, 0.4},
        {0.2, 0.5, 0.3, 0.6, 0.4, 1.0},
    }

    // Normalize allocations to sum to 1 (full investment)
    totalAlloc := 0.0
    for _, a := range allocation {
        totalAlloc += a
    }
    normalized := make([]float64, len(allocation))
    for i := range allocation {
        normalized[i] = allocation[i] / totalAlloc
    }

    // Calculate expected return
    expectedReturn := 0.0
    for i := range normalized {
        expectedReturn += normalized[i] * returns[i]
    }

    // Calculate portfolio risk (variance)
    portfolioRisk := 0.0
    for i := range normalized {
        for j := range normalized {
            portfolioRisk += normalized[i] * normalized[j] *
                volatilities[i] * volatilities[j] * correlations[i][j]
        }
    }
    portfolioStdDev := math.Sqrt(portfolioRisk)

    // Penalty for concentration (encourage diversification)
    concentrationPenalty := 0.0
    for _, a := range normalized {
        if a > 0.4 {  // Penalize if more than 40% in one asset
            concentrationPenalty += (a - 0.4) * 2.0
        }
    }

    // Sharpe-like objective: maximize return per unit risk
    // Negate for minimization; risk-free rate = 0.02
    riskAdjustedReturn := (expectedReturn - 0.02) / portfolioStdDev

    // Minimize: -Sharpe + concentration penalty
    return -riskAdjustedReturn + concentrationPenalty
}

func main() {
    fmt.Println("=== Portfolio Optimization with EOBBMA ===\n")

    // EOBBMA is excellent for this deceptive problem
    // (risk-return landscape has many local optima)
    config := mayfly.NewEOBBMAConfig()
    config.ObjectiveFunc = portfolioObjective
    config.ProblemSize = 6  // 6 assets
    config.LowerBound = 0.0  // Min allocation
    config.UpperBound = 1.0  // Max allocation
    config.MaxIterations = 500

    // Moderate Lévy flights for financial optimization
    config.LevyAlpha = 1.5   // Balanced exploration
    config.LevyBeta = 1.0
    config.OppositionRate = 0.3

    result, err := mayfly.Optimize(config)
    if err != nil {
        panic(err)
    }

    // Normalize final allocation
    totalAlloc := 0.0
    for _, a := range result.GlobalBest.Position {
        totalAlloc += a
    }

    fmt.Println("Optimal Portfolio Allocation:")
    assetNames := []string{"Stocks", "Bonds", "Real Estate", "Cash", "Commodities", "Crypto"}
    for i, a := range result.GlobalBest.Position {
        percentage := (a / totalAlloc) * 100
        fmt.Printf("  %s: %.1f%%\n", assetNames[i], percentage)
    }

    fmt.Printf("\nObjective Value: %.6f\n", result.GlobalBest.Cost)
    fmt.Printf("Function Evaluations: %d\n", result.FuncEvalCount)
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
