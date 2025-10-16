# Getting Started with Mayfly Optimization

A practical tutorial for getting started with the Mayfly optimization library.

## Installation

```bash
go get github.com/CWBudde/mayfly
```

## Basic Usage

### 1. Simple Optimization

The simplest way to get started:

```go
package main

import (
    "fmt"
    "github.com/CWBudde/mayfly"
)

func main() {
    // Create configuration
    config := mayfly.NewDefaultConfig()

    // Define problem
    config.ObjectiveFunc = mayfly.Sphere  // Built-in test function
    config.ProblemSize = 30               // 30 dimensions
    config.LowerBound = -10              // Search bounds
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

### 2. Custom Objective Function

Optimize your own function:

```go
// Define your optimization problem
func myFunction(x []float64) float64 {
    // Example: Minimize sum of squares
    sum := 0.0
    for _, val := range x {
        sum += val * val
    }
    return sum
}

func main() {
    config := mayfly.NewDefaultConfig()
    config.ObjectiveFunc = myFunction  // Use your function
    config.ProblemSize = 10
    config.LowerBound = -5
    config.UpperBound = 5
    config.MaxIterations = 500

    result, err := mayfly.Optimize(config)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Minimum found: %f\n", result.GlobalBest.Cost)
}
```

### 3. Using Algorithm Variants

Try different variants for better performance:

```go
// Standard MA (baseline)
config := mayfly.NewDefaultConfig()

// DESMA (better for multimodal)
config := mayfly.NewDESMAConfig()

// OLCE-MA (best for highly multimodal)
config := mayfly.NewOLCEConfig()

// EOBBMA (best for deceptive landscapes)
config := mayfly.NewEOBBMAConfig()

// GSASMA (fastest convergence)
config := mayfly.NewGSASMAConfig()

// MPMA (most stable)
config := mayfly.NewMPMAConfig()

// AOBLMOA (adaptive multi-phase)
config := mayfly.NewAOBLMOAConfig()
```

## Real-World Examples

### Example 1: Function Fitting

Fit parameters to minimize error:

```go
// Observed data
observedX := []float64{1, 2, 3, 4, 5}
observedY := []float64{2.1, 4.3, 5.9, 8.2, 10.1}

// Model: y = a*x + b
func modelError(params []float64) float64 {
    a, b := params[0], params[1]

    totalError := 0.0
    for i := range observedX {
        predicted := a*observedX[i] + b
        error := predicted - observedY[i]
        totalError += error * error  // Sum of squared errors
    }

    return totalError
}

func main() {
    config := mayfly.NewDefaultConfig()
    config.ObjectiveFunc = modelError
    config.ProblemSize = 2  // Two parameters: a and b
    config.LowerBound = -10
    config.UpperBound = 10
    config.MaxIterations = 500

    result, err := mayfly.Optimize(config)
    if err != nil {
        panic(err)
    }

    a, b := result.GlobalBest.Position[0], result.GlobalBest.Position[1]
    fmt.Printf("Best fit: y = %.3f*x + %.3f\n", a, b)
    fmt.Printf("Error: %.6f\n", result.GlobalBest.Cost)
}
```

### Example 2: Resource Allocation

Optimize resource distribution:

```go
// Allocate budget across N projects to maximize ROI
func projectROI(allocation []float64) float64 {
    // Expected returns for each project (example data)
    returns := []float64{0.15, 0.12, 0.18, 0.10, 0.20}
    risks := []float64{0.05, 0.03, 0.08, 0.02, 0.10}

    totalReturn := 0.0
    totalRisk := 0.0

    for i := range allocation {
        totalReturn += allocation[i] * returns[i]
        totalRisk += allocation[i] * risks[i]
    }

    // Maximize return while minimizing risk
    // (negate for minimization)
    score := totalReturn - 0.5*totalRisk
    return -score
}

func main() {
    config := mayfly.NewGSASMAConfig()  // Fast convergence
    config.ObjectiveFunc = projectROI
    config.ProblemSize = 5  // 5 projects
    config.LowerBound = 0   // Minimum allocation
    config.UpperBound = 100 // Maximum allocation per project
    config.MaxIterations = 300

    result, err := mayfly.Optimize(config)
    if err != nil {
        panic(err)
    }

    fmt.Println("Optimal allocation:")
    for i, amount := range result.GlobalBest.Position {
        fmt.Printf("Project %d: $%.2f\n", i+1, amount)
    }
    fmt.Printf("Expected score: %.4f\n", -result.GlobalBest.Cost)
}
```

### Example 3: Hyperparameter Tuning

Optimize machine learning hyperparameters:

```go
import "math"

// Simulate model validation error
func modelPerformance(hyperparams []float64) float64 {
    learningRate := hyperparams[0]
    momentum := hyperparams[1]
    dropout := hyperparams[2]

    // Simulate training (replace with real model)
    // This is a synthetic error function
    error := math.Abs(learningRate - 0.01) +
             math.Abs(momentum - 0.9) +
             math.Abs(dropout - 0.2)

    return error
}

func main() {
    config := mayfly.NewOLCEConfig()  // Good for multimodal
    config.ObjectiveFunc = modelPerformance
    config.ProblemSize = 3  // 3 hyperparameters

    // Set appropriate bounds for each parameter
    config.LowerBound = 0.0001  // Will apply to all
    config.UpperBound = 0.5
    config.MaxIterations = 200

    result, err := mayfly.Optimize(config)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Best hyperparameters:\n")
    fmt.Printf("Learning rate: %.6f\n", result.GlobalBest.Position[0])
    fmt.Printf("Momentum: %.6f\n", result.GlobalBest.Position[1])
    fmt.Printf("Dropout: %.6f\n", result.GlobalBest.Position[2])
    fmt.Printf("Validation error: %.6f\n", result.GlobalBest.Cost)
}
```

## Advanced Features

### Reproducible Results

Use a fixed random seed for reproducibility:

```go
import "math/rand"

config := mayfly.NewDefaultConfig()
config.Rand = rand.New(rand.NewSource(42))  // Fixed seed
config.ObjectiveFunc = mayfly.Rastrigin
config.ProblemSize = 30
config.LowerBound = -5.12
config.UpperBound = 5.12

// Results will be identical across runs
result, _ := mayfly.Optimize(config)
```

### Maximization Problems

The library minimizes by default. For maximization, negate:

```go
func profit(x []float64) float64 {
    // Calculate profit
    p := calculateProfit(x)
    return -p  // Negate for maximization
}

config := mayfly.NewDefaultConfig()
config.ObjectiveFunc = profit
// ... rest of config

result, _ := mayfly.Optimize(config)
actualProfit := -result.GlobalBest.Cost  // Negate back
```

### Different Bounds Per Dimension

Currently, the library uses uniform bounds. For per-dimension bounds, use transformation:

```go
func transformedObjective(normalizedX []float64) float64 {
    // Define actual bounds per dimension
    lowerBounds := []float64{-10, 0, -5}
    upperBounds := []float64{10, 100, 5}

    // Transform from [0,1] to actual bounds
    actualX := make([]float64, len(normalizedX))
    for i := range normalizedX {
        actualX[i] = lowerBounds[i] + normalizedX[i]*(upperBounds[i]-lowerBounds[i])
    }

    // Evaluate with actual values
    return yourObjective(actualX)
}

config := mayfly.NewDefaultConfig()
config.ObjectiveFunc = transformedObjective
config.LowerBound = 0
config.UpperBound = 1
```

## Algorithm Selection Guide

### Quick Selection

**Not sure which variant to use?** Use the selector:

```go
// Define your problem characteristics
chars := mayfly.ProblemCharacteristics{
    Dimensionality:            30,
    Modality:                  mayfly.HighlyMultimodal,
    Landscape:                 mayfly.Rugged,
    ExpensiveEvaluations:      false,
    RequiresFastConvergence:   false,
    RequiresStableConvergence: false,
}

// Get recommendation
selector := mayfly.NewAlgorithmSelector()
best := selector.RecommendBest(chars)

// Use recommended variant
result, err := mayfly.NewBuilderFromVariant(best).
    ForProblem(myFunction, 30, -10, 10).
    WithIterations(500).
    Optimize()
```

### Rule of Thumb

| Problem Type | Recommended Variant | Why |
|--------------|---------------------|-----|
| Smooth, single optimum | **Standard MA** | Efficient baseline |
| Multiple local optima | **DESMA** | Adaptive elite search |
| Highly multimodal | **OLCE-MA** | Orthogonal learning + chaos |
| Deceptive landscape | **EOBBMA** | Lévy flights escape traps |
| Need fast results | **GSASMA** | Fastest convergence |
| Need stability | **MPMA** | Most robust |
| Complex/adaptive needs | **AOBLMOA** | 4 hunting strategies |

## Common Pitfalls

### 1. Not Setting Required Parameters

```go
// ❌ WRONG - Missing required fields
config := mayfly.NewDefaultConfig()
result, err := mayfly.Optimize(config)  // Will error!

// ✅ CORRECT - All required fields set
config := mayfly.NewDefaultConfig()
config.ObjectiveFunc = mayfly.Sphere
config.ProblemSize = 30
config.LowerBound = -10
config.UpperBound = 10
result, err := mayfly.Optimize(config)
```

### 2. Wrong Bounds Order

```go
// ❌ WRONG - Lower > Upper
config.LowerBound = 10
config.UpperBound = -10  // Will error!

// ✅ CORRECT
config.LowerBound = -10
config.UpperBound = 10
```

### 3. Too Few Iterations

```go
// ❌ WRONG - Too few iterations
config.MaxIterations = 10  // Won't converge well

// ✅ CORRECT - Sufficient iterations
config.MaxIterations = 500  // Start with at least 500
```

### 4. Inappropriate Population Size

```go
// ❌ WRONG - Too small for high dimensions
config.ProblemSize = 100
config.NPop = 10  // Insufficient

// ✅ CORRECT - Scale with dimensions
config.ProblemSize = 100
config.NPop = 50  // Better coverage
```

## Running Examples

The library includes complete examples in the `examples/` directory:

```bash
# Basic usage
cd examples
go run main.go

# Algorithm comparison
cd examples/comparison
go run main.go

# Algorithm selector demo
cd examples/selector
go run main.go
```

## Next Steps

- **[Algorithm Documentation](algorithms/)** - Learn about each variant in detail
- **[Configuration Guide](api/configuration.md)** - Full parameter reference
- **[Benchmark Functions](benchmarks.md)** - Test functions and expected results
- **[Comparison Framework](api/comparison-framework.md)** - Statistical testing

## Quick Tips

1. **Start simple**: Begin with Standard MA on a simple function
2. **Test on benchmarks**: Validate your setup with built-in functions
3. **Choose variant**: Pick specialized variant based on problem type
4. **Tune incrementally**: Start with defaults, tune only if needed
5. **Use reproducibility**: Set random seed for debugging
6. **Monitor convergence**: Check if solution improves over iterations
7. **Scale resources**: Increase population/iterations for complex problems

## Getting Help

- Check [GitHub Issues](https://github.com/CWBudde/mayfly/issues)
- Read [CLAUDE.md](../CLAUDE.md) for development guidance
- See [PLAN.md](../PLAN.md) for roadmap and future features
