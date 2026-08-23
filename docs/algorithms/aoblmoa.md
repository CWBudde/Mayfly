# AOBLMOA - Aquila Optimizer and Opposition-Based Learning Mayfly Optimization Algorithm

## Research Reference

Zhao, Y.; Huang, C.; Zhang, M.; Cui, Y. "AOBLMOA: A Hybrid Biomimetic Optimization
Algorithm for Numerical Optimization and Engineering Design Problems."
_Biomimetics_ **2023**, 8(4), 381. DOI:
[10.3390/biomimetics8040381](https://doi.org/10.3390/biomimetics8040381).

## Overview

AOBLMOA takes the Mayfly Algorithm and changes exactly two of its stages: the
position update, where the Aquila Optimizer's hunting strategies replace the
nuptial dance and the random flight, and the offspring stage, where stochastic
opposition-based learning replaces Gaussian mutation.

Everything else — attraction, crossover, sorting, truncation selection — is the
plain Mayfly Algorithm.

> **Changed in v0.6.0.** Through v0.5.1 this variant did not implement the
> paper: the branch was a probability rather than a fitness test, opposition
> ran inside the update phase rather than on the offspring, and `StrategySwitch`
> was never read. Results recorded under earlier releases were produced by a
> different algorithm. See the CHANGELOG.

## Key Innovations

### 1. Aquila Optimizer Integration

The **Aquila Optimizer** mimics the hunting behavior of eagles (Aquila genus) with four distinct strategies that adapt based on iteration progress:

#### X1 - Expanded Exploration (High soar with vertical stoop)

- **When**: First 1/3 of iterations
- **Purpose**: Global search across entire space
- **Formula**: `X₁ = Xbest * (1 - t/T) + (Xmean - Xbest * rand)`
- **Behavior**: Wide-ranging exploration using population mean

#### X2 - Narrowed Exploration (Contour flight with short glide)

- **When**: Iterations 1/3 to 2/3
- **Purpose**: Focused exploration with Lévy flight
- **Formula**: `X₂ = Xbest * Levy(D) + XR + (y - x) * rand`
- **Behavior**: Combines heavy-tailed jumps with local search

#### X3 - Expanded Exploitation (Low flight with slow descent)

- **When**: Last 1/3 of iterations
- **Purpose**: Convergence to promising regions
- **Formula**: `X₃ = (Xbest - Xmean) * α - rand + exploration`
- **Behavior**: Balances convergence with controlled exploration

#### X4 - Narrowed Exploitation (Walk and grab)

- **When**: Final iterations
- **Purpose**: Intensive local search
- **Formula**: `X₄ = QF * Xbest - (G1 * X * rand) - G2 * Levy(D)`
- **Behavior**: Fine-tunes solutions with quality function

**Which strategy applies**: there are two phases, not three, and no coin flip
inside a phase. The individual's sex picks the pair; the phase picks the member.

```
Iteration            Males              Females
──────────────────────────────────────────────────────────
before StrategySwitch   X2 (narrowed expl.)  X1 (expanded expl.)
from StrategySwitch on  X4 (narrowed expt.)  X3 (expanded expt.)
```

`StrategySwitch` defaults to two thirds of `MaxIterations`. See
`aoblmoaStrategyFor`, which carries one of the paper's open contradictions
(below).

### 2. The Mayfly/Aquila branch is a fitness test

Every individual moves every iteration. Which branch it takes is decided by
fitness, not by chance:

- A **male** keeps the Mayfly attraction term while the global best dominates
  him — Eq. (29). Otherwise he hunts as an Aquila. The Aquila step replaces the
  nuptial dance, nothing else.
- A **female** keeps the attraction term while her paired male dominates her —
  Eq. (30). Otherwise she hunts. The Aquila step replaces the random flight.

The Aquila branch is a position formula, so it leaves the individual's velocity
untouched.

`AquilaWeight` is **deprecated**. The paper has no such knob, and its old
default of `1.0` sent the whole swarm down the Aquila branch every iteration,
so the Mayfly attraction terms — the half of the hybrid the variant is named
for — never ran. It now defaults to `AquilaWeightAuto`, which selects the
fitness test. Setting a probability in `[0, 1]` restores the pre-v0.6.0
behavior of drawing the branch at random; it exists only to reproduce old runs.

### 3. Stochastic opposition-based learning replaces mutation

OBL is applied on the **offspring**, after crossover, to **every** offspring,
with no gate:

- **Opposition point**, Eq. (31): `x̃ = (lower + upper − x) × r`, `r ~ N(0, 1)`.
  The Gaussian factor is essential; it is what distinguishes this from
  Tizhoosh's plain reflection, and it means the result routinely leaves the
  search bounds and is clamped back.
- **Greedy selection**, Eq. (32): the better of the offspring and its opposition
  point survives.

This stage takes the slot Gaussian mutation occupies in the plain algorithm, so
**`NM` is inert under AOBLMOA** — `effectiveNM` reports `0` for it.
`OppositionProbability` is likewise unread by AOBLMOA; it is kept because the
other opposition-based variants use it.

The evaluation budget per iteration is `NPop + NPopF + 2·nc`.

### 3a. Open questions in the paper

The paper is ambiguous or self-contradictory at three points. Each is isolated
to a single function with a comment saying exactly what to change to flip it:

| Question                                                                     | Carried by                     | Current choice                                  |
| ---------------------------------------------------------------------------- | ------------------------------ | ----------------------------------------------- |
| Female branch inequality: Eq. (30) or the Algorithm 1 pseudocode?            | `aoblmoaFemaleTakesAttraction` | Eq. (30), matching `prepareStandardFemale`      |
| Which sex gets which strategy pair? The equations and the abstract disagree. | `aoblmoaStrategyFor`           | The equations: males narrowed, females expanded |
| Is Eq. (31)'s `r` drawn per solution or per dimension?                       | `stochasticOppositionPoint`    | Per dimension                                   |

### 4. Multi-Objective Building Blocks

The package ships a multi-objective toolkit that callers can drive themselves.
The optimizer does **not** maintain a Pareto archive of its own: nothing in the
search ever read one back, so the NSGA-II pruning was pure overhead. Build a
`ParetoArchive` and feed it with `UpdateFromPopulation` or `AddFromMayfly` when
you want a front.

#### Pareto Dominance

- Solution A dominates B if: A is no worse in all objectives AND strictly better in at least one
- Non-dominated solutions form the Pareto front
- Archive maintains best non-dominated solutions found

#### Crowding Distance

- Measures density of solutions in objective space
- Higher values = more isolated solutions (better diversity)
- Used for selection when archive exceeds size limit

#### NSGA-II Selection

- Combines Pareto ranking and crowding distance
- Maintains both convergence and diversity
- Automatic archive management

#### Performance Metrics

- **Hypervolume**: Volume dominated by Pareto front (higher is better)
- **IGD**: Inverted Generational Distance to true front (lower is better)

## Usage Examples

### Basic Single-Objective Optimization

```go
package main

import (
    "fmt"
    "github.com/cwbudde/mayfly"
)

func main() {
    // Use AOBLMOA for complex optimization with adaptive strategy switching
    config := mayfly.NewAOBLMOAConfig()
    config.ObjectiveFunc = mayfly.Rastrigin
    config.ProblemSize = 30
    config.LowerBound = -5.12
    config.UpperBound = 5.12
    config.MaxIterations = 500

    result, err := mayfly.Optimize(config)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Best Cost: %f\n", result.GlobalBest.Cost)
}
```

### Advanced Usage with Custom Strategy Weighting

```go
package main

import (
    "fmt"
    "github.com/cwbudde/mayfly"
)

func main() {
    // Configure AOBLMOA with custom Aquila/Mayfly blend
    config := mayfly.NewAOBLMOAConfig()
    config.ObjectiveFunc = mayfly.Schwefel  // Deceptive landscape
    config.ProblemSize = 30
    config.LowerBound = -500
    config.UpperBound = 500
    config.MaxIterations = 1000

    // Stay in the Aquila exploration phase longer than the default 2/3
    config.StrategySwitch = 800  // Exploit only over the last 200 iterations

    // Larger archive for a Pareto front you build yourself
    config.ArchiveSize = 150

    // Larger population for complex landscape
    config.NPop = 50
    config.NPopF = 50

    result, err := mayfly.Optimize(config)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Final Cost: %.2f\n", result.GlobalBest.Cost)
    fmt.Printf("Iterations: %d\n", result.IterationCount)
    fmt.Printf("Function Evaluations: %d\n", result.FuncEvalCount)
}
```

### Multi-Objective Optimization Example

```go
package main

import (
    "fmt"
    "math"
    "github.com/cwbudde/mayfly"
)

// Multi-objective function: minimize both objectives
// Objective 1: Distance from origin (Sphere)
// Objective 2: Rosenbrock function value
func multiObjective(x []float64) float64 {
    obj1 := mayfly.Sphere(x)      // Minimize distance to origin
    obj2 := mayfly.Rosenbrock(x)  // Minimize Rosenbrock valley

    // For single-objective interface, use weighted sum
    return 0.5*obj1 + 0.5*obj2
}

func main() {
    fmt.Println("=== Multi-Objective Optimization with AOBLMOA ===\n")

    config := mayfly.NewAOBLMOAConfig()
    config.ObjectiveFunc = multiObjective
    config.ProblemSize = 10
    config.LowerBound = -5
    config.UpperBound = 10
    config.MaxIterations = 500

    result, err := mayfly.Optimize(config)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Best Compromise Solution Cost: %.6f\n", result.GlobalBest.Cost)
}
```

### Real-World Example: Multi-Criteria Resource Allocation

```go
package main

import (
    "fmt"
    "math"
    "github.com/cwbudde/mayfly"
)

// Resource allocation with multiple conflicting objectives:
// 1. Maximize total performance
// 2. Minimize total cost
// 3. Balance resource distribution (minimize variance)
func resourceAllocation(allocation []float64) float64 {
    // Simulated performance gains per resource unit
    performance := []float64{1.5, 2.0, 1.2, 1.8, 2.5}

    // Simulated costs per resource unit
    costs := []float64{10.0, 15.0, 8.0, 12.0, 20.0}

    // Calculate objectives
    totalPerformance := 0.0
    totalCost := 0.0
    for i := range allocation {
        totalPerformance += allocation[i] * performance[i]
        totalCost += allocation[i] * costs[i]
    }

    // Calculate distribution balance (lower variance = better)
    mean := 0.0
    for _, a := range allocation {
        mean += a
    }
    mean /= float64(len(allocation))

    variance := 0.0
    for _, a := range allocation {
        diff := a - mean
        variance += diff * diff
    }
    variance /= float64(len(allocation))

    // Normalize objectives to similar scales
    perfObjective := -totalPerformance / 50.0 // Negate to minimize (was maximize)
    costObjective := totalCost / 100.0
    balanceObjective := math.Sqrt(variance) / 10.0

    // Weighted combination (can be adjusted based on priorities)
    // Higher weights = higher importance
    return 0.4*perfObjective + 0.4*costObjective + 0.2*balanceObjective
}

func main() {
    fmt.Println("=== Multi-Criteria Resource Allocation with AOBLMOA ===\n")

    // AOBLMOA excels at multi-criteria problems with conflicting objectives
    config := mayfly.NewAOBLMOAConfig()
    config.ObjectiveFunc = resourceAllocation
    config.ProblemSize = 5  // 5 resources to allocate
    config.LowerBound = 0.0  // Minimum allocation
    config.UpperBound = 20.0 // Maximum allocation per resource
    config.MaxIterations = 600

    // Larger archive for diverse Pareto solutions
    config.ArchiveSize = 120

    result, err := mayfly.Optimize(config)
    if err != nil {
        panic(err)
    }

    fmt.Println("Optimal Resource Allocation:")
    resourceNames := []string{"Server Capacity", "Network Bandwidth", "Storage", "Memory", "Processing"}
    for i, amount := range result.GlobalBest.Position {
        fmt.Printf("  %s: %.2f units\n", resourceNames[i], amount)
    }

    // Calculate final objectives for display
    performance := []float64{1.5, 2.0, 1.2, 1.8, 2.5}
    costs := []float64{10.0, 15.0, 8.0, 12.0, 20.0}

    totalPerf := 0.0
    totalCost := 0.0
    for i, a := range result.GlobalBest.Position {
        totalPerf += a * performance[i]
        totalCost += a * costs[i]
    }

    fmt.Printf("\nPerformance Metrics:\n")
    fmt.Printf("  Total Performance: %.2f\n", totalPerf)
    fmt.Printf("  Total Cost:        $%.2f\n", totalCost)
    fmt.Printf("  Combined Score:    %.6f (lower is better)\n", result.GlobalBest.Cost)
    fmt.Printf("\nFunction Evaluations: %d\n", result.FuncEvalCount)
}
```

**Note**: `Optimize` is single-objective. The Pareto helpers (`ParetoArchive`, `dominates`, NSGA-II selection, crowding distance, hypervolume, IGD) are exported building blocks you drive yourself; the optimizer does not maintain an archive during the run.

## AOBLMOA Parameters

- `UseAOBLMOA`: Enable AOBLMOA variant (default: false)
- `StrategySwitch`: First iteration of the Aquila exploitation phase (default: `0`, which resolves to 2/3 of `MaxIterations` per run and is never written back). A value at or beyond `MaxIterations` is legal and means "never exploit". Negative values are rejected.
- `AquilaWeight`: **Deprecated.** Default `AquilaWeightAuto` (`-1`) selects the paper's fitness test. A probability in `[0, 1]` restores the pre-v0.6.0 random branch draw.
- `OppositionProbability`: Unused by AOBLMOA, which opposes every offspring. Kept for the other opposition-based variants.
- `NM`: Inert under AOBLMOA; opposition replaces mutation.
- `ArchiveSize`: Maximum size for a `ParetoArchive` the caller builds (default: 100). The optimizer does not maintain one.

## Benefits

- **Adaptive Strategy**: Four Aquila strategies split across sexes and phases
- **Better Exploration**: Aquila strategies prevent premature convergence
- **Fitness-Directed Hybrid**: Only individuals a better solution dominates keep the social attraction term; the rest hunt
- **Robust Performance**: Works well across diverse problem types
- **Predictable Budget**: `NPop + NPopF + 2·nc` evaluations per iteration

Note that `Optimize` is single-objective. The Pareto helpers are exported
building blocks for callers who want a front; the search does not read one.

## When to Use AOBLMOA

- **Best for**: Complex multi-modal problems requiring adaptive strategies
- **Excellent on**: Problems with varying landscape characteristics (mix of smooth/rugged regions)
- **Use when**: Single algorithm struggles across all iterations
- **Ideal for**: Multi-objective optimization with conflicting objectives
- **Examples**: Engineering design tradeoffs, portfolio optimization, resource allocation with multiple criteria

## Parameter Tuning Guide

### Strategy Switch Settings

`StrategySwitch` is the only phase knob AOBLMOA has, and the one the paper
actually defines. It is the first iteration of the exploitation phase.

**Default** (two thirds of the budget):

```go
config.StrategySwitch = 0  // resolves to MaxIterations * 2 / 3
```

Leaving it at `0` means a `Config` reused with a different `MaxIterations`
rescales, because the resolution is never written back.

**Exploit earlier** (smooth or unimodal landscapes):

```go
config.StrategySwitch = config.MaxIterations / 3
```

**Never exploit** (pure exploration, legal and occasionally useful as a
baseline):

```go
config.StrategySwitch = config.MaxIterations
```

### Aquila Weight (deprecated)

`AquilaWeight` has no counterpart in the paper. Leave it at its default:

```go
config.AquilaWeight = mayfly.AquilaWeightAuto  // the default
```

Set a probability only to reproduce a run recorded before v0.6.0:

```go
config.AquilaWeight = 1.0  // the pre-v0.6.0 default: every individual hunts
config.AquilaWeight = 0.0  // every individual takes the standard Mayfly update
```

Note that even with the override the offspring stage is the paper's, so a
pre-v0.6.0 run is not reproduced exactly.

### Archive Size (Multi-Objective)

`ArchiveSize` only sizes a `ParetoArchive` you build yourself; it has no effect
on `Optimize`.

**Small Archive** (fast, focused):

```go
config.ArchiveSize = 50
```

- Use when: Want only best Pareto solutions
- Faster archive management
- Less diversity preservation

**Large Archive** (comprehensive):

```go
config.ArchiveSize = 200
```

- Use when: Need complete Pareto front representation
- Better diversity across objectives
- More computational cost for archive maintenance

## AOBLMOA vs Other Variants

**Choose AOBLMOA when**:

- Problem has distinct phases requiring different strategies
- You need multi-objective optimization capabilities
- Want adaptive exploration-exploitation without manual tuning
- Problem characteristics change across the search space

**Choose EOBBMA instead when**:

- Problem is purely deceptive (Schwefel-like)
- Want simpler Bare Bones framework
- Heavy-tailed jumps alone are sufficient

**Choose GSASMA instead when**:

- Need maximum convergence speed
- Simulated annealing fits problem structure
- Prefer hybrid mutation over strategy switching

**Choose OLCE-MA instead when**:

- Problem is highly multimodal with many local optima
- Orthogonal learning benefits parameter space exploration
- Chaotic perturbations are effective

## Related Documentation

- [EOBBMA](eobbma.md) - Lévy flight alternative
- [GSASMA](gsasma.md) - Fast convergence alternative
- [OLCE-MA](olce-ma.md) - Multimodal specialist
- [Configuration Guide](../api/configuration.md) - Complete parameter reference
