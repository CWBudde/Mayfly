# Mayfly Optimization Algorithm (Go)

A Go implementation of the Mayfly Optimization Algorithm (MA), a nature-inspired metaheuristic optimization algorithm based on the mating behavior of mayflies.

## Original Research

This implementation is based on:

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

## Features

- Clean, idiomatic Go implementation
- **DESMA variant included** - Dynamic Elite Strategy Mayfly Algorithm for improved performance
- Configurable algorithm parameters
- Multiple benchmark functions included (Sphere, Rastrigin, Rosenbrock, Ackley, Griewank)
- Easy to use with custom objective functions
- Thread-safe (with proper configuration)

## Installation

```bash
go get github.com/CWBudde/mayfly
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/CWBudde/mayfly"
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

## Custom Objective Function

```go
// Define your own objective function
func myFunction(x []float64) float64 {
    // Your optimization problem here
    sum := 0.0
    for _, val := range x {
        sum += val * val
    }
    return sum
}

// Use it in the configuration
config := mayfly.NewDefaultConfig()
config.ObjectiveFunc = myFunction
config.ProblemSize = 10
config.LowerBound = -5
config.UpperBound = 5
```

## Configuration Parameters

### Problem Parameters

- `ObjectiveFunc`: The function to optimize (minimize)
- `ProblemSize`: Number of decision variables (dimensions)
- `LowerBound`: Lower bound for decision variables
- `UpperBound`: Upper bound for decision variables

### Algorithm Parameters

- `MaxIterations`: Maximum number of iterations (default: 2000)
- `NPop`: Population size for males (default: 20)
- `NPopF`: Population size for females (default: 20)
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

### DESMA Parameters (Dynamic Elite Strategy)

- `UseDESMA`: Enable DESMA variant (default: false)
- `EliteCount`: Number of elite mayflies to generate per iteration (default: 5)
- `SearchRange`: Initial search range for elite generation (default: auto-calculated as 10% of search space)
- `EnlargeFactor`: Factor to enlarge search range when improving (default: 1.05)
- `ReductionFactor`: Factor to reduce search range when not improving (default: 0.95)

## Using DESMA Variant

DESMA (Dynamic Elite Strategy Mayfly Algorithm) is an improved variant that addresses local optima trapping and slow convergence. It dynamically generates elite solutions around the current best position and adapts the search range based on performance.

```go
// Use DESMA for better performance
config := mayfly.NewDESMAConfig()
config.ObjectiveFunc = mayfly.Rastrigin
config.ProblemSize = 50
config.LowerBound = -10
config.UpperBound = 10
config.MaxIterations = 1000

result, err := mayfly.Optimize(config)
```

### DESMA Benefits

- **Better convergence**: Escapes local optima more effectively
- **Adaptive search**: Dynamically adjusts search range based on improvement
- **Faster optimization**: Often achieves better results with the same number of iterations
- **Minimal overhead**: Only 5-10% more function evaluations

## Benchmark Functions

The library includes several standard benchmark functions:

### Sphere Function

- Global minimum: f(0, ..., 0) = 0
- Typical bounds: [-10, 10]
- Characteristics: Unimodal, convex

### Rastrigin Function

- Global minimum: f(0, ..., 0) = 0
- Typical bounds: [-5.12, 5.12]
- Characteristics: Multimodal, highly complex

### Rosenbrock Function

- Global minimum: f(1, ..., 1) = 0
- Typical bounds: [-5, 10]
- Characteristics: Unimodal, narrow valley

### Ackley Function

- Global minimum: f(0, ..., 0) = 0
- Typical bounds: [-32.768, 32.768]
- Characteristics: Multimodal, nearly flat outer region

### Griewank Function

- Global minimum: f(0, ..., 0) = 0
- Typical bounds: [-600, 600]
- Characteristics: Multimodal, many local minima

## Running the Examples

```bash
cd examples
go run main.go
```

This will run the optimization algorithm on multiple benchmark functions and display the results.

## Algorithm Workflow

1. **Initialization**: Create random populations of male and female mayflies
2. **Female Update**: Females move toward males with better fitness or fly randomly
3. **Male Update**: Males fly toward their personal best and the global best, or perform nuptial dance
4. **Mating**: Best males and females produce offspring through crossover
5. **Mutation**: Random mutations introduce diversity
6. **Selection**: Best individuals survive to the next generation
7. **Repeat**: Steps 2-6 until convergence or maximum iterations

## Performance Tips

- Start with default parameters and tune based on your problem
- Increase population size for more complex problems
- Reduce MaxIterations for faster convergence testing
- Use a custom random source (`Config.Rand`) for reproducibility
- For high-dimensional problems, consider increasing population sizes
