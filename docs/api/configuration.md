# Configuration Guide

Complete reference for all configuration parameters in the Mayfly optimization library.

## Problem Parameters

These parameters define the optimization problem:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `ObjectiveFunc` | `func([]float64) float64` | **Yes** | The function to minimize |
| `ProblemSize` | `int` | **Yes** | Number of decision variables (dimensions) |
| `LowerBound` | `float64` | **Yes** | Lower bound for all decision variables |
| `UpperBound` | `float64` | **Yes** | Upper bound for all decision variables |

### Example
```go
config := mayfly.NewDefaultConfig()
config.ObjectiveFunc = mayfly.Sphere
config.ProblemSize = 30
config.LowerBound = -10
config.UpperBound = 10
```

## Population Parameters

Control the size and behavior of the mayfly populations:

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `NPop` | `int` | 20 | Population size for males |
| `NPopF` | `int` | 20 | Population size for females |
| `MaxIterations` | `int` | 2000 | Maximum number of iterations |

**Recommendations**:
- Increase population for complex/high-dimensional problems (30-50)
- Decrease for simple problems or quick testing (10-15)
- `NPop` and `NPopF` are typically equal

## Velocity Parameters

Control movement behavior of mayflies:

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `G` | `float64` | 0.8 | Inertia weight |
| `GDamp` | `float64` | 1.0 | Inertia weight damping ratio |
| `A1` | `float64` | 1.0 | Personal learning coefficient |
| `A2` | `float64` | 1.5 | Global learning coefficient for males |
| `A3` | `float64` | 1.5 | Global learning coefficient for females |
| `Beta` | `float64` | 2.0 | Distance sight coefficient |
| `Dance` | `float64` | 5.0 | Nuptial dance coefficient |
| `FL` | `float64` | 1.0 | Random flight coefficient |
| `DanceDamp` | `float64` | 0.8 | Dance damping ratio |
| `FLDamp` | `float64` | 0.99 | Flight damping ratio |
| `VelMax` | `float64` | Auto* | Maximum velocity (auto: 10% of bounds) |
| `VelMin` | `float64` | Auto* | Minimum velocity (auto: -10% of bounds) |

*Auto-calculated if left at 0

## Mating Parameters

Control genetic operators:

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `NC` | `int` | 20 | Number of offspring per iteration |
| `NM` | `int` | Auto* | Number of mutants (auto: 5% of NPop) |
| `Mu` | `float64` | 0.01 | Mutation rate |

*Auto-calculated if left at 0

## Variant-Specific Parameters

### DESMA Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `UseDESMA` | `bool` | false | Enable DESMA variant |
| `EliteCount` | `int` | 5 | Number of elite solutions to generate |
| `SearchRange` | `float64` | Auto* | Search range for elite generation |
| `EnlargeFactor` | `float64` | 1.05 | Factor to enlarge range when improving |
| `ReductionFactor` | `float64` | 0.95 | Factor to reduce range when stagnating |

*Auto: 10% of (UpperBound - LowerBound)

### OLCE-MA Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `UseOLCE` | `bool` | false | Enable OLCE-MA variant |
| `OrthogonalFactor` | `float64` | 0.3 | Orthogonal learning strength (0-1) |
| `ChaosFactor` | `float64` | 0.1 | Chaos perturbation strength (0-1) |

### EOBBMA Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `UseEOBBMA` | `bool` | false | Enable EOBBMA variant |
| `LevyAlpha` | `float64` | 1.5 | Lévy stability parameter (0 < α ≤ 2) |
| `LevyBeta` | `float64` | 1.0 | Lévy scale parameter |
| `OppositionRate` | `float64` | 0.3 | Opposition learning probability (0-1) |
| `EliteOppositionCount` | `int` | 3 | Number of elites for opposition |

### GSASMA Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `UseGSASMA` | `bool` | false | Enable GSASMA variant |
| `InitialTemperature` | `float64` | 100.0 | Starting temperature for SA |
| `CoolingRate` | `float64` | 0.95 | Temperature decay rate |
| `CoolingSchedule` | `string` | "exponential" | Schedule: "exponential", "linear", "logarithmic" |
| `CauchyMutationRate` | `float64` | 0.3 | Base Cauchy mutation probability |
| `GoldenFactor` | `float64` | 1.0 | GSA influence factor (0.5-2.0) |
| `ApplyOBLToGlobalBest` | `bool` | true | Enable OBL on global best |

### MPMA Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `UseMPMA` | `bool` | false | Enable MPMA variant |
| `MedianWeight` | `float64` | 0.5 | Median position influence (0-1) |
| `GravityType` | `string` | "linear" | Gravity type: "linear", "exponential", "sigmoid" |
| `UseWeightedMedian` | `bool` | false | Use fitness-weighted median |

### AOBLMOA Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `UseAOBLMOA` | `bool` | false | Enable AOBLMOA variant |
| `AquilaWeight` | `float64` | 0.5 | Aquila vs Mayfly blend (0-1) |
| `OppositionProbability` | `float64` | 0.3 | OBL application probability (0-1) |
| `ArchiveSize` | `int` | 100 | Max Pareto archive size |
| `StrategySwitch` | `int` | Auto* | Iteration threshold for strategy switch |

*Auto: 2/3 of MaxIterations

## Advanced Parameters

### Random Number Generation

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `Rand` | `*rand.Rand` | nil | Custom random number generator for reproducibility |

**Example for reproducible results**:
```go
import "math/rand"

config := mayfly.NewDefaultConfig()
config.Rand = rand.New(rand.NewSource(42))  // Fixed seed
```

## Factory Functions

Pre-configured factory functions for each variant:

```go
// Standard MA
config := mayfly.NewDefaultConfig()

// DESMA
config := mayfly.NewDESMAConfig()

// OLCE-MA
config := mayfly.NewOLCEConfig()

// EOBBMA
config := mayfly.NewEOBBMAConfig()

// GSASMA
config := mayfly.NewGSASMAConfig()

// MPMA
config := mayfly.NewMPMAConfig()

// AOBLMOA
config := mayfly.NewAOBLMOAConfig()
```

All factory functions set sensible defaults. You only need to set the required problem parameters.

## Configuration Validation

The `Optimize()` function validates configuration:

**Required fields** (must be non-zero):
- `ObjectiveFunc`
- `ProblemSize`
- `LowerBound` (can be negative)
- `UpperBound`

**Auto-calculated fields** (if zero):
- `VelMax` = 0.1 * (UpperBound - LowerBound)
- `VelMin` = -VelMax
- `NM` = max(1, int(0.05 * NPop))
- `SearchRange` (DESMA) = 0.1 * (UpperBound - LowerBound)

**Validation errors**:
```go
result, err := mayfly.Optimize(config)
if err != nil {
    // Handle errors:
    // - "objective function is required"
    // - "problem size must be positive"
    // - "invalid bounds: lower bound must be less than upper bound"
    // - etc.
}
```

## Configuration Tips

### For Quick Testing
```go
config := mayfly.NewDefaultConfig()
config.MaxIterations = 100  // Reduce for speed
config.NPop = 10
config.NPopF = 10
```

### For High-Dimensional Problems
```go
config := mayfly.NewDefaultConfig()
config.NPop = 50  // Increase population
config.NPopF = 50
config.MaxIterations = 2000
```

### For Expensive Function Evaluations
```go
config := mayfly.NewDESMAConfig()
config.EliteCount = 3  // Reduce elite count
config.NPop = 15  // Smaller population
config.MaxIterations = 500  // Fewer iterations
```

### For Maximization Problems
```go
// Negate the objective function
func maximize(x []float64) float64 {
    profit := calculateProfit(x)
    return -profit  // Negate for maximization
}

config.ObjectiveFunc = maximize
```

## Related Documentation

- [Unified Framework](unified-framework.md) - Builder API and algorithm selection
- [Algorithm Comparison](comparison-framework.md) - Statistical comparison tools
- [Getting Started](../getting-started.md) - Tutorial and examples
