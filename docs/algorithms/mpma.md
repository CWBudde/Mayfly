# MPMA - Median Position-Based Mayfly Algorithm

## Research Reference

Guo, L.; Xu, C.; Yu, T.; Tuerxun, W. “An Improved Mayfly Optimization
Algorithm Based on Median Position and Its Application in the Optimization of
PID Parameters of Hydro-Turbine Governor.” _IEEE Access_ **2022**, 10,
36335–36349. DOI:
[10.1109/ACCESS.2022.3160714](https://doi.org/10.1109/ACCESS.2022.3160714).

The paper abbreviates the method as MMA. Mayfly calls it MPMA to avoid an
ambiguous acronym.

## Overview

MPMA adds a fitness-ranked middle position and a nonlinear gravity coefficient
to standard MA. The paper evaluates it on classic benchmarks and hydro-turbine
governor PID tuning. Those published results are evidence for that protocol,
not a general performance guarantee.

## Key Innovations

### 1. Median Position Guidance

The paper sorts mayflies by objective value. For an odd population it takes the
middle-ranked mayfly's complete position vector; for an even population it
averages the complete position vectors of the two middle-ranked mayflies. It is
not a coordinate-wise population median.

**Mathematical Foundation**:

```
Sort mayflies by objective value.
pm = x[(n+1)/2]                         when n is odd
pm = (x[n/2] + x[n/2+1]) / 2           when n is even

Velocity update includes median attraction:
  v = g*v + a1*exp(-β*r_pb²)*(pbest - x)
        + a2*exp(-β*r_gb²)*(gbest - x)
        + a4*exp(-β*r_med²)*(pm - x)

where:
  r_pb  = distance to personal best
  r_gb  = distance to global best
  r_med = distance to median position
  a4    = median-attraction coefficient
```

The paper does not publish `a4` and does not clearly state whether `n` covers
males only or both sexes. Mayfly's `MedianWeight = 0.5` and male-only pool are
compatibility choices, not exact paper settings.

### 2. Non-linear Gravity Coefficient

The paper's schedule is:

```
g(t) = 0.5*sqrt(1-(t/T)^2) + 0.4
```

It decreases from 0.9 to 0.4. `GravityPaper` is the default. Linear,
exponential, and sigmoid schedules remain available as explicit library
extensions.

### 3. Weighted Median (Optional)

Fitness-weighted median guidance is a Mayfly extension and is not part of the
paper protocol:

- **Weight calculation**: Better fitness → higher weight
- **Weighted median**: Cumulative weight ≥ 50% determines median
- **Effect**: Population guidance shifts toward elite solutions

**Example**:

```
Population: [x1=0.1, x2=0.3, x3=0.5]  (sorted by fitness)
Costs:      [1.0,    5.0,    10.0]    (lower is better)

Regular median: 0.3
Weighted median: ~0.2 (shifted toward better solution x1)
```

Enable it only for an exploratory library experiment, not for a paper
reproduction.

The complete source-audited protocol and all 110 tabular MMA output cells are
available in the
[machine-readable reference](../reference-data/mpma-2022-tables1-10.json).
The artifact records the remaining exact-reproduction gates and makes no
reproduction claim. A separate
[machine-readable clarification request](../reference-data/mpma-2022-clarification-request.json)
pins the five exact-preset and three comparison/model questions and includes an
unsent correspondence draft.

## Usage Examples

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/cwbudde/mayfly"
)

func main() {
    // Use MPMA for stable convergence on control problems
    config := mayfly.NewMPMAConfig()
    config.ObjectiveFunc = mayfly.Rosenbrock  // Narrow valley function
    config.ProblemSize = 10
    config.LowerBound = -5
    config.UpperBound = 10
    config.MaxIterations = 500

    result, err := mayfly.Optimize(config)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Best Cost: %f\n", result.GlobalBest.Cost)
}
```

### Advanced Usage with Sigmoid Gravity

```go
package main

import (
    "fmt"
    "github.com/cwbudde/mayfly"
)

func main() {
    // Configure MPMA with sigmoid gravity for smooth phase transition
    config := mayfly.NewMPMAConfig()
    config.ObjectiveFunc = mayfly.BentCigar  // Ill-conditioned function
    config.ProblemSize = 20
    config.LowerBound = -10
    config.UpperBound = 10
    config.MaxIterations = 800

    // Use sigmoid gravity for balanced exploration-exploitation
    config.GravityType = "sigmoid"

    // Strong median influence for stability
    config.MedianWeight = 0.7

    // Use fitness-weighted median
    config.UseWeightedMedian = true

    // Larger population for ill-conditioned problems
    config.NPop = 30
    config.NPopF = 30

    result, err := mayfly.Optimize(config)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Final Cost: %.6f\n", result.GlobalBest.Cost)
    fmt.Printf("Iterations: %d\n", result.IterationCount)
    fmt.Printf("Function Evaluations: %d\n", result.FuncEvalCount)
}
```

### Real-World Example: System Identification

```go
package main

import (
    "fmt"
    "math"
    "github.com/cwbudde/mayfly"
)

// System identification: fit transfer function parameters to measured data
// This is a classic control engineering problem requiring stable convergence
func systemIdentificationError(params []float64) float64 {
    // Transfer function: H(s) = K / (τ₁s + 1)(τ₂s + 1)
    // Parameters: K (gain), τ₁ (time constant 1), τ₂ (time constant 2)
    K := params[0]
    tau1 := params[1]
    tau2 := params[2]

    // Simulated "measured" frequency response data
    // (In practice, this would come from experiments)
    frequencies := []float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0}
    measuredMagnitudes := []float64{0.95, 0.85, 0.65, 0.35, 0.10, 0.03}
    measuredPhases := []float64{-5, -25, -45, -75, -110, -145} // degrees

    totalError := 0.0

    for i, omega := range frequencies {
        // Calculate model response at this frequency
        s := complex(0, omega) // s = jω

        // Transfer function: H(jω)
        numerator := complex(K, 0)
        denominator := (complex(1, 0) + complex(tau1, 0)*s) *
            (complex(1, 0) + complex(tau2, 0)*s)

        H := numerator / denominator

        // Extract magnitude and phase
        modelMagnitude := math.Abs(real(H)*real(H) + imag(H)*imag(H))
        modelMagnitude = math.Sqrt(modelMagnitude)
        modelPhase := math.Atan2(imag(H), real(H)) * 180 / math.Pi

        // Compute error
        magError := math.Abs(modelMagnitude - measuredMagnitudes[i])
        phaseError := math.Abs(modelPhase - measuredPhases[i])

        // Weight magnitude error more heavily
        totalError += magError*10 + phaseError*0.1
    }

    // Add penalty for physically unrealistic parameters
    if K < 0 || tau1 < 0 || tau2 < 0 {
        totalError += 1000
    }

    return totalError
}

func main() {
    fmt.Println("=== System Identification with MPMA ===\n")

    // MPMA is ideal for system identification
    // (requires stable convergence, handles ill-conditioning well)
    config := mayfly.NewMPMAConfig()
    config.ObjectiveFunc = systemIdentificationError
    config.ProblemSize = 3  // K, τ₁, τ₂
    config.LowerBound = 0.01
    config.UpperBound = 10.0
    config.MaxIterations = 600

    // Use exponential gravity for quick convergence
    config.GravityType = "exponential"

    // Moderate median influence
    config.MedianWeight = 0.5

    result, err := mayfly.Optimize(config)
    if err != nil {
        panic(err)
    }

    fmt.Println("Identified Transfer Function Parameters:")
    fmt.Printf("  K (Gain):          %.4f\n", result.GlobalBest.Position[0])
    fmt.Printf("  τ₁ (Time const 1): %.4f s\n", result.GlobalBest.Position[1])
    fmt.Printf("  τ₂ (Time const 2): %.4f s\n", result.GlobalBest.Position[2])
    fmt.Printf("\nFitting Error: %.6f\n", result.GlobalBest.Cost)
    fmt.Printf("Function Evaluations: %d\n", result.FuncEvalCount)

    fmt.Println("\nTransfer Function:")
    K := result.GlobalBest.Position[0]
    tau1 := result.GlobalBest.Position[1]
    tau2 := result.GlobalBest.Position[2]
    fmt.Printf("H(s) = %.4f / ((%.4fs + 1)(%.4fs + 1))\n", K, tau1, tau2)
}
```

## MPMA Parameters

- `UseMPMA`: Enable MPMA variant (default: false)
- `MedianWeight`: Influence of the ranked median position (default: 0.5,
  range: 0-1). The paper's corresponding `a4` is unpublished.
- `GravityType`: Type of gravity coefficient (default: "paper")
  - Options: "paper", "linear", "exponential", "sigmoid"
  - Only "paper" is the published MPMA schedule.
- `UseWeightedMedian`: Use the library's fitness-weighted extension (default:
  false)

## Published Evidence

The source paper reports MMA aggregates for 18 classic functions and two
hydro-turbine governor conditions. It reports stronger results on 16 of the 18
functions, but does not publish seeds or raw trials. Treat those tables as
descriptive evidence for the reported setup, not as universal improvement
percentages or variance guarantees.

## When to Use MPMA

- **Best for**: Control system optimization (PID tuning, system identification)
- **Excellent on**: Ill-conditioned problems with narrow valleys
- **Use when**: Standard MA shows oscillatory behavior
- **Ideal for**: Problems where convergence stability matters more than speed
- **Examples**: Rosenbrock, BentCigar, Discus, engineering design optimization

## Parameter Tuning Guide

### Median Weight Settings

**Balanced Guidance** (default):

```go
config.MedianWeight = 0.5  // Equal influence with personal/global best
```

**More Population Influence**:

```go
config.MedianWeight = 0.8  // Strong median influence
// Use when: Population diversity is important
```

**Less Population Influence**:

```go
config.MedianWeight = 0.2  // Weak median influence
// Use when: Global best is already near optimum
```

### Gravity Type Selection

**Paper schedule** (default):

```go
config.GravityType = "paper"
```

- Exact published gravity equation
- Required when investigating the paper protocol

**Exponential** (fast convergence):

```go
config.GravityType = "exponential"
```

- Use for: Unimodal functions (Sphere, Rosenbrock)
- Use for: Time-critical applications
- Caution: May converge prematurely on multimodal functions

**Sigmoid** (smooth transition):

```go
config.GravityType = "sigmoid"
```

- Use for: Problems requiring careful phase transition
- Use for: Control system tuning (PID parameters)
- Best for: Maintaining exploration longer before exploitation

### Weighted Median

**Standard Median** (default):

```go
config.UseWeightedMedian = false
```

- True robust median, completely outlier-resistant

**Weighted Median** (elite emphasis):

```go
config.UseWeightedMedian = true
```

- Use when: Elite solutions are significantly better
- Use when: Population has large fitness variance
- Caution: Less robust to outliers

## MPMA vs Other Variants

**Choose MPMA when**:

- You need stable, predictable convergence
- Problem has outliers or noisy evaluations
- Oscillatory behavior is observed in standard algorithms
- Working on control system optimization

**Choose DESMA instead when**:

- Problem has many local optima
- You want adaptive local search
- Function evaluations are cheap

**Choose GSASMA instead when**:

- You need fast convergence speed
- Problem requires aggressive exploration
- You can tolerate higher variance

**Choose OLCE-MA instead when**:

- Problem is highly multimodal
- Need systematic parameter space exploration
- High dimensionality (20D+)

## Reproduction Status

No exact preset or published comparison is exposed. The missing `a4`, median
population pool, genetic-operator parameters, initialization and boundary
semantics, function-evaluation accounting, seeds, raw runs, and Simulink model
must be resolved first. See the
[paper-reproduction audit](../paper-reproduction.md#mpma-protocol-audit).
The corresponding clarification draft is explicitly marked `not_sent`.

## Related Documentation

- [GSASMA](gsasma.md) - For faster convergence alternative
- [Standard MA](standard-ma.md) - Base algorithm
- [DESMA](desma.md) - For elite-based local search
- [Configuration Guide](../api/configuration.md) - Complete parameter reference
