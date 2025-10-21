# Standard Mayfly Algorithm (MA)

## Research Reference

**Zervoudakis, K., & Tsafarakis, S. (2020). A mayfly optimization algorithm. Computers & Industrial Engineering, 145, 106559.**

https://doi.org/10.1016/j.cie.2020.106559

Original MATLAB implementation by:
- K. Zervoudakis (kzervoudakis@isc.tuc.gr)
- S. Tsafarakis
- School of Production Engineering and Management, Technical University of Crete, Chania, Greece

## Overview

The Mayfly Algorithm is a swarm intelligence optimization algorithm inspired by the flight behavior and mating process of mayflies. The algorithm simulates:

- **Male mayflies**: Perform nuptial dances and are attracted to the global best position
- **Female mayflies**: Are attracted to males with better fitness
- **Mating process**: Crossover and mutation operations create offspring
- **Population evolution**: Best individuals survive to the next generation

## Key Features

- Clean, idiomatic Go implementation
- Dual-population structure (males and females with different behaviors)
- Velocity-based updates with exponential distance weighting
- Genetic operators (crossover and mutation)
- Configurable parameters with sensible defaults

## Usage Example

```go
package main

import (
    "fmt"
    "github.com/cwbudde/mayfly"
)

func main() {
    // Create configuration
    config := mayfly.NewDefaultConfig()
    config.ObjectiveFunc = mayfly.Sphere
    config.ProblemSize = 30
    config.LowerBound = -10
    config.UpperBound = 10
    config.MaxIterations = 1000

    // Run optimization
    result, err := mayfly.Optimize(config)
    if err != nil {
        panic(err)
    }

    // Print results
    fmt.Printf("Best Cost: %f\n", result.GlobalBest.Cost)
    fmt.Printf("Best Position: %v\n", result.GlobalBest.Position)
}
```

## Algorithm Parameters

### Problem Parameters

- `ObjectiveFunc`: The function to optimize (minimize)
- `ProblemSize`: Number of decision variables (dimensions)
- `LowerBound`: Lower bound for decision variables
- `UpperBound`: Upper bound for decision variables

### Population Parameters

- `NPop`: Population size for males (default: 20)
- `NPopF`: Population size for females (default: 20)
- `MaxIterations`: Maximum number of iterations (default: 2000)

### Velocity Parameters

- `G`: Inertia weight (default: 0.8)
- `GDamp`: Inertia weight damping ratio (default: 1.0)
- `A1`: Personal learning coefficient (default: 1.0)
- `A2`: Global learning coefficient for males (default: 1.5)
- `A3`: Global learning coefficient for females (default: 1.5)
- `Beta`: Distance sight coefficient (default: 2.0)
- `Dance`: Nuptial dance coefficient (default: 5.0)
- `FL`: Random flight coefficient (default: 1.0)
- `DanceDamp`: Dance damping ratio (default: 0.8)
- `FLDamp`: Flight damping ratio (default: 0.99)

### Mating Parameters

- `NC`: Number of offspring (default: 20)
- `NM`: Number of mutants (default: 5% of NPop)
- `Mu`: Mutation rate (default: 0.01)

## Algorithm Workflow

1. **Initialization**: Create random populations of male and female mayflies
2. **Female Update**: Females move toward males with better fitness or fly randomly
3. **Male Update**: Males fly toward their personal best and the global best, or perform nuptial dance
4. **Mating**: Best males and females produce offspring through crossover
5. **Mutation**: Random mutations introduce diversity
6. **Selection**: Best individuals survive to the next generation
7. **Repeat**: Steps 2-6 until convergence or maximum iterations

## Male Velocity Update

Males track personal best positions and update based on:

```
v = g*v + a1*exp(-β*r_pb²)*(pbest - x) + a2*exp(-β*r_gb²)*(gbest - x)
```

Where:
- `r_pb` = distance to personal best
- `r_gb` = distance to global best
- `β` = distance sight coefficient

When at optimum (very close to global best), males perform "nuptial dance":
```
v = g*v + dance*e  (random flight)
```

## Female Velocity Update

Females are attracted to males with better fitness OR fly randomly:

```
If female.Cost > male.Cost:
    v = g*v + a3*exp(-β*r_mf²)*(male - x)  (attraction)
Else:
    v = g*v + fl*e  (random flight)
```

## When to Use Standard MA

- **Best for**: General-purpose optimization problems
- **Excellent on**: Unimodal functions (Sphere, Rosenbrock)
- **Use when**: You need a balanced, well-tested baseline
- **Ideal for**: Initial exploration before trying specialized variants

## Performance Tips

- Start with default parameters and tune based on your problem
- Increase population size for more complex problems
- Reduce MaxIterations for faster convergence testing
- Use a custom random source (`Config.Rand`) for reproducibility
- For high-dimensional problems, consider increasing population sizes

## Comparison with Variants

| Aspect | Standard MA | When to Consider Variants |
|--------|-------------|---------------------------|
| **Convergence** | Good on unimodal | DESMA/OLCE for multimodal |
| **Local optima** | Can get trapped | EOBBMA/GSASMA for escape |
| **Stability** | Moderate | MPMA for more stable |
| **Speed** | Balanced | GSASMA for faster |
| **Complexity** | Simplest | AOBLMOA for adaptive |

See [Algorithm Comparison](../README.md#algorithm-comparison) for detailed performance metrics.

## Related Documentation

- [DESMA - Dynamic Elite Strategy](desma.md) - Enhanced variant for local optima escape
- [Configuration Guide](../api/configuration.md) - Complete parameter reference
- [Getting Started](../getting-started.md) - Tutorial and examples
