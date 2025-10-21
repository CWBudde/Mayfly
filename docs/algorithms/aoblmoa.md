# AOBLMOA - Aquila Optimizer-Based Learning Multi-Objective Algorithm

## Research Reference

**AOBLMOA: A Hybrid Biomimetic Optimization Algorithm (2023). PubMed / Various journals**

## Overview

AOBLMOA is a powerful hybrid metaheuristic that combines the social behavior of the Mayfly Algorithm with the hunting strategies of the Aquila Optimizer. This variant excels at complex single-objective problems and provides built-in support for multi-objective optimization with Pareto dominance.

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

**Adaptive Strategy Switching**:
```
Iteration Progress     Strategy    Mode
───────────────────────────────────────────
0-33%                 X1/X2       Exploration
33-66%                X2/X3       Transition
66-100%               X3/X4       Exploitation
```

### 2. Hybrid Operator Switching

AOBLMOA creates a **hybrid** between Mayfly and Aquila behaviors:

- **AquilaWeight** parameter controls the blend (default: 0.5)
- Each mayfly has probability `AquilaWeight` of using Aquila strategies
- Otherwise uses standard Mayfly velocity updates
- **Best of both worlds**: Mayfly's social learning + Aquila's hunting intelligence

**Example with AquilaWeight = 0.5**:
```
50% of mayflies → Use Aquila hunting strategies (adaptive)
50% of mayflies → Use Mayfly velocity updates (social)
```

### 3. Opposition-Based Learning Framework

**Opposition-Based Learning** (OBL) expands search coverage:

- **Opposition Point**: `x_opp = lower + upper - x`
- **Applied with probability**: `OppositionProbability` (default: 0.3)
- **Evaluation**: If opposition point is better, accept it
- **Benefit**: Searches both sides of space simultaneously

**When Applied**:
- After Aquila strategy updates
- Before accepting new positions
- Only to solutions selected by probability threshold

### 4. Multi-Objective Support

AOBLMOA includes **complete multi-objective optimization** framework:

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

    // More Aquila for aggressive exploration
    config.AquilaWeight = 0.7  // 70% Aquila, 30% Mayfly

    // Aggressive opposition learning
    config.OppositionProbability = 0.4

    // Larger archive for diverse solutions
    config.ArchiveSize = 150

    // Custom strategy switch point (default is 2/3 of iterations)
    config.StrategySwitch = 600  // Switch at iteration 600

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
    // (Pareto front is still maintained internally)
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
    config.ArchiveSize = 100  // Store up to 100 Pareto-optimal solutions

    result, err := mayfly.Optimize(config)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Best Compromise Solution Cost: %.6f\n", result.GlobalBest.Cost)
    fmt.Println("\nNote: Internal Pareto archive maintained during optimization")
    fmt.Println("using NSGA-II selection with crowding distance.")
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

    // Balanced Aquila/Mayfly blend
    config.AquilaWeight = 0.5

    // Moderate opposition
    config.OppositionProbability = 0.3

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

**Note**: Full multi-objective interface (accepting `MultiObjectiveFunction` that returns multiple values) is available through the internal archive system. The Pareto front is maintained during optimization using NSGA-II selection with crowding distance.

## AOBLMOA Parameters

- `UseAOBLMOA`: Enable AOBLMOA variant (default: false)
- `AquilaWeight`: Probability of using Aquila strategies vs Mayfly (default: 0.5, range: 0-1)
- `OppositionProbability`: Probability of applying OBL (default: 0.3, range: 0-1)
- `ArchiveSize`: Maximum Pareto archive size for multi-objective (default: 100)
- `StrategySwitch`: Iteration threshold for strategy switching (default: auto-set to 2/3 of MaxIterations)

## Benefits

- **Adaptive Strategy**: Four distinct strategies for different search phases
- **Better Exploration**: Aquila strategies prevent premature convergence
- **Multi-Objective Native**: No additional code needed for MO problems
- **Flexible Hybrid**: AquilaWeight parameter controls algorithm balance
- **Robust Performance**: Works well across diverse problem types
- **Moderate Overhead**: ~20-30% more function evaluations for better solutions

## When to Use AOBLMOA

- **Best for**: Complex multi-modal problems requiring adaptive strategies
- **Excellent on**: Problems with varying landscape characteristics (mix of smooth/rugged regions)
- **Use when**: Single algorithm struggles across all iterations
- **Ideal for**: Multi-objective optimization with conflicting objectives
- **Examples**: Engineering design tradeoffs, portfolio optimization, resource allocation with multiple criteria

## Parameter Tuning Guide

### Aquila Weight Settings

**Balanced Hybrid** (default):
```go
config.AquilaWeight = 0.5  // 50% Aquila, 50% Mayfly
```
- Best starting point for most problems
- Combines strengths of both algorithms

**More Aquila** (aggressive exploration):
```go
config.AquilaWeight = 0.7  // 70% Aquila, 30% Mayfly
```
- Use when: Problem has many deceptive local optima
- Use when: Need strong exploration capability
- Trade-off: May converge slower

**More Mayfly** (social learning):
```go
config.AquilaWeight = 0.3  // 30% Aquila, 70% Mayfly
```
- Use when: Problem benefits from swarm intelligence
- Use when: Social learning is effective (smooth landscapes)
- Trade-off: Less adaptive strategy switching

**Pure Strategies**:
```go
config.AquilaWeight = 1.0  // 100% Aquila (pure Aquila Optimizer)
config.AquilaWeight = 0.0  // 100% Mayfly (standard Mayfly Algorithm)
```

### Opposition Probability Settings

**Moderate Opposition** (default):
```go
config.OppositionProbability = 0.3  // 30% of updates use OBL
```
- Balanced exploration of opposite regions
- Minimal computational overhead

**Aggressive Opposition**:
```go
config.OppositionProbability = 0.5  // 50% of updates use OBL
```
- Use when: Search space is large and sparsely sampled
- Use when: Initial solutions are far from optimum
- Caution: Doubles function evaluations for OBL

**Conservative Opposition**:
```go
config.OppositionProbability = 0.1  // 10% of updates use OBL
```
- Use when: Function evaluations are expensive
- Use when: Initial population is well-distributed
- Lower overhead, less exploration

### Archive Size (Multi-Objective)

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
