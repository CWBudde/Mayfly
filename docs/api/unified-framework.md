# Unified Framework & API

The library provides a unified framework for working with all algorithm variants through a consistent API, intelligent algorithm selection, and comprehensive comparison tools.

## Variant Interface

All algorithm variants implement the `AlgorithmVariant` interface, providing a consistent way to interact with any algorithm:

```go
// Get a variant by name
variant := mayfly.NewVariant("desma")

// Get variant information
fmt.Println(variant.Name())         // "DESMA"
fmt.Println(variant.FullName())     // "Dynamic Elite Strategy Mayfly Algorithm"
fmt.Println(variant.Description())  // Brief description
fmt.Println(variant.RecommendedFor()) // Problem types this variant excels at

// Get default configuration
config := variant.GetConfig()
```

### Available Variant Names

- `"ma"` - Standard Mayfly Algorithm
- `"desma"` - Dynamic Elite Strategy MA
- `"olce"` - Orthogonal Learning & Chaotic Exploitation MA
- `"eobbma"` - Elite Opposition-Based Bare Bones MA
- `"gsasma"` - Golden Sine with Simulated Annealing MA
- `"mpma"` - Median Position-Based MA
- `"aoblmoa"` - Aquila Optimizer-Based Learning MO Algorithm

## Fluent Builder API

Build and run optimizations with a fluent interface:

```go
result, err := mayfly.NewBuilder("gsasma").
    ForProblem(mayfly.Rastrigin, 20, -5.12, 5.12).
    WithIterations(500).
    WithPopulation(30, 30).
    WithConfig(func(c *mayfly.Config) {
        c.CoolingRate = 0.97
    }).
    Optimize()
```

### Builder Methods

| Method | Parameters | Description |
|--------|------------|-------------|
| `NewBuilder(variant string)` | variant name | Create builder for specific variant |
| `NewBuilderFromVariant(v AlgorithmVariant)` | variant instance | Create builder from variant object |
| `ForProblem(f, size, lower, upper)` | func, int, float64, float64 | Set problem definition |
| `WithIterations(n)` | int | Set max iterations |
| `WithPopulation(males, females)` | int, int | Set population sizes |
| `WithConfig(fn)` | func(*Config) | Modify config with custom function |
| `Optimize()` | - | Run optimization and return result |

### Examples

**Basic usage**:
```go
result, err := mayfly.NewBuilder("desma").
    ForProblem(myFunction, 10, -5, 5).
    WithIterations(1000).
    Optimize()
```

**With custom configuration**:
```go
result, err := mayfly.NewBuilder("olce").
    ForProblem(mayfly.Rastrigin, 30, -5.12, 5.12).
    WithIterations(500).
    WithPopulation(40, 40).
    WithConfig(func(c *mayfly.Config) {
        c.OrthogonalFactor = 0.4
        c.ChaosFactor = 0.15
    }).
    Optimize()
```

## Algorithm Selection

Let the library recommend the best algorithm for your problem:

### Define Problem Characteristics

```go
characteristics := mayfly.ProblemCharacteristics{
    Dimensionality:            30,
    Modality:                  mayfly.HighlyMultimodal,
    Landscape:                 mayfly.Rugged,
    ExpensiveEvaluations:      false,
    RequiresFastConvergence:   false,
    RequiresStableConvergence: false,
    MultiObjective:            false,
}
```

### Characteristic Enums

**Modality**:
- `mayfly.Unimodal` - Single optimum
- `mayfly.Multimodal` - Multiple local optima
- `mayfly.HighlyMultimodal` - Many local optima

**Landscape**:
- `mayfly.Smooth` - Well-behaved gradients
- `mayfly.Rugged` - Complex, irregular landscape
- `mayfly.Deceptive` - Misleading gradients (Schwefel-like)
- `mayfly.NarrowValley` - Ill-conditioned (Rosenbrock-like)

### Get Recommendations

```go
selector := mayfly.NewAlgorithmSelector()
recommendations := selector.RecommendAlgorithms(characteristics)

// recommendations is a slice sorted by score (best first)
for _, rec := range recommendations {
    fmt.Printf("%s: %.1f%% match\n", rec.Variant.Name(), rec.Score*100)
    fmt.Printf("Reason: %s\n", rec.Reason)
}

// Use the best recommendation
best := recommendations[0]
result, err := mayfly.NewBuilderFromVariant(best.Variant).
    ForProblem(mayfly.Rastrigin, 30, -5.12, 5.12).
    Optimize()
```

### Quick Recommendations

For standard benchmarks, use predefined recommendations:

```go
rec := mayfly.RecommendForBenchmark("Schwefel")
fmt.Printf("Recommended: %s (Score: %.1f%%)\n",
    rec.Variant.Name(), rec.Score*100)

// Available benchmark names:
// "Sphere", "Rastrigin", "Rosenbrock", "Ackley", "Griewank",
// "Schwefel", "Levy", "Zakharov", "DixonPrice", "Michalewicz",
// "BentCigar", "Discus", "Weierstrass", "HappyCat", "ExpandedSchafferF6"
```

## Automatic Problem Classification

Classify unknown problems automatically by sampling:

```go
characteristics := mayfly.ClassifyProblem(
    myFunction,  // Your objective function
    10,          // Problem size
    -10,         // Lower bound
    10,          // Upper bound
)

// Now use characteristics for algorithm selection
selector := mayfly.NewAlgorithmSelector()
best := selector.RecommendBest(characteristics)
```

**Classification Process**:
1. Samples function at random points
2. Analyzes gradient behavior
3. Detects modality through local optimization
4. Classifies landscape characteristics
5. Returns `ProblemCharacteristics` object

## Configuration Presets

Use predefined configurations for common problem types:

```go
config, err := mayfly.NewPresetConfig(mayfly.PresetDeceptive)
config.ObjectiveFunc = mayfly.Schwefel
config.ProblemSize = 10
config.LowerBound = -500
config.UpperBound = 500

result, err := mayfly.Optimize(config)
```

### Available Presets

| Preset | Algorithm | Best For |
|--------|-----------|----------|
| `PresetUnimodal` | Standard MA | Single-optimum problems |
| `PresetMultimodal` | DESMA | Multi-modal problems |
| `PresetHighlyMultimodal` | OLCE-MA | Many local optima |
| `PresetDeceptive` | EOBBMA | Deceptive landscapes |
| `PresetNarrowValley` | MPMA | Ill-conditioned problems |
| `PresetHighDimensional` | OLCE-MA | High-D problems (larger population) |
| `PresetFastConvergence` | GSASMA | Quick results needed |
| `PresetStableConvergence` | MPMA | Robust optimization |
| `PresetMultiObjective` | AOBLMOA | Multi-objective problems |

## Configuration Files

Save and load configurations from JSON:

### Save Configuration

```go
config := mayfly.NewOLCEConfig()
config.ProblemSize = 20
config.MaxIterations = 500
err := mayfly.SaveConfigToFile(config, "config.json")
```

### Load Configuration

```go
config, err := mayfly.LoadConfigFromFile("config.json")
if err != nil {
    panic(err)
}

// Set function separately (can't serialize functions)
config.ObjectiveFunc = mayfly.Rastrigin

result, err := mayfly.Optimize(config)
```

### JSON Format Example

```json
{
  "ProblemSize": 20,
  "LowerBound": -5.12,
  "UpperBound": 5.12,
  "MaxIterations": 500,
  "NPop": 30,
  "NPopF": 30,
  "UseOLCE": true,
  "OrthogonalFactor": 0.3,
  "ChaosFactor": 0.1
}
```

## Auto-Tuning

Automatically tune configuration based on problem characteristics:

```go
config := mayfly.NewGSASMAConfig()
characteristics := mayfly.ProblemCharacteristics{
    Dimensionality:            50,
    RequiresFastConvergence:   true,
}

mayfly.AutoTuneConfig(config, characteristics)
// Population and iterations adjusted automatically
// High-D problem → larger population
// Fast convergence → adjusted cooling schedule
```

**Auto-tuning adjustments**:
- **Dimensionality**: Scales population size
- **Modality**: Adjusts exploration parameters
- **Fast convergence**: Optimizes for speed
- **Stable convergence**: Optimizes for robustness
- **Expensive evaluations**: Reduces population/iterations

## Complete Working Examples

### Example 1: Automatic Algorithm Selection

```go
package main

import (
    "fmt"
    "github.com/cwbudde/mayfly"
)

func main() {
    // Let the library choose the best algorithm for your problem
    characteristics := mayfly.ProblemCharacteristics{
        Dimensionality:            30,
        Modality:                  mayfly.HighlyMultimodal,
        Landscape:                 mayfly.Rugged,
        ExpensiveEvaluations:      false,
        RequiresFastConvergence:   false,
        RequiresStableConvergence: false,
        MultiObjective:            false,
    }

    selector := mayfly.NewAlgorithmSelector()
    recommendations := selector.RecommendAlgorithms(characteristics)

    fmt.Println("=== Algorithm Recommendations ===")
    for i, rec := range recommendations[:3] {  // Show top 3
        fmt.Printf("%d. %s (%.1f%% match)\n",
            i+1, rec.Variant.FullName(), rec.Score*100)
        fmt.Printf("   Reason: %s\n\n", rec.Reason)
    }

    // Use the best recommendation
    best := recommendations[0]
    result, err := mayfly.NewBuilderFromVariant(best.Variant).
        ForProblem(mayfly.Rastrigin, 30, -5.12, 5.12).
        WithIterations(500).
        Optimize()

    if err != nil {
        panic(err)
    }

    fmt.Printf("Best Cost with %s: %.4f\n",
        best.Variant.Name(), result.GlobalBest.Cost)
}
```

### Example 2: Builder API for Quick Prototyping

```go
package main

import (
    "fmt"
    "math"
    "github.com/cwbudde/mayfly"
)

// Custom objective: minimize energy consumption of a system
func energyConsumption(params []float64) float64 {
    voltage := params[0]
    frequency := params[1]
    loadFactor := params[2]

    // Simplified energy model: E = V² * f * load
    energy := voltage * voltage * frequency * loadFactor

    // Add constraints as penalties
    penalty := 0.0

    // Voltage must be between 3.3V and 5V
    if voltage < 3.3 || voltage > 5.0 {
        penalty += 1000
    }

    // Frequency: 1-100 MHz
    if frequency < 1.0 || frequency > 100.0 {
        penalty += 1000
    }

    // Load factor: 0-1
    if loadFactor < 0 || loadFactor > 1.0 {
        penalty += 1000
    }

    return energy + penalty
}

func main() {
    fmt.Println("=== System Energy Optimization ===\n")

    // Use builder API for quick setup
    result, err := mayfly.NewBuilder("gsasma").  // Fast convergence
        ForProblem(energyConsumption, 3, 0, 100).
        WithIterations(300).
        WithPopulation(25, 25).
        WithConfig(func(c *mayfly.Config) {
            // Fine-tune GSASMA parameters
            c.CoolingRate = 0.97
            c.CauchyMutationRate = 0.3
        }).
        Optimize()

    if err != nil {
        panic(err)
    }

    voltage := result.GlobalBest.Position[0]
    frequency := result.GlobalBest.Position[1]
    loadFactor := result.GlobalBest.Position[2]

    fmt.Printf("Optimal Configuration:\n")
    fmt.Printf("  Voltage:    %.2f V\n", voltage)
    fmt.Printf("  Frequency:  %.2f MHz\n", frequency)
    fmt.Printf("  Load Factor: %.3f\n", loadFactor)
    fmt.Printf("\nMinimum Energy: %.4f units\n", result.GlobalBest.Cost)
}
```

### Example 3: Comparison Across Multiple Variants

```go
package main

import (
    "fmt"
    "github.com/cwbudde/mayfly"
)

func main() {
    fmt.Println("=== Comparing Algorithm Variants ===\n")

    // Define the problem
    problemFunc := mayfly.Rastrigin
    dimensions := 20
    lower, upper := -5.12, 5.12
    iterations := 400

    // Test multiple variants
    variants := []string{"ma", "desma", "olce", "eobbma", "gsasma"}
    results := make(map[string]float64)

    for _, variant := range variants {
        result, err := mayfly.NewBuilder(variant).
            ForProblem(problemFunc, dimensions, lower, upper).
            WithIterations(iterations).
            Optimize()

        if err != nil {
            fmt.Printf("Error with %s: %v\n", variant, err)
            continue
        }

        results[variant] = result.GlobalBest.Cost
        fmt.Printf("%s: %.4f (after %d evaluations)\n",
            variant, result.GlobalBest.Cost, result.FuncEvalCount)
    }

    // Find best
    bestVariant := ""
    bestCost := math.MaxFloat64
    for variant, cost := range results {
        if cost < bestCost {
            bestCost = cost
            bestVariant = variant
        }
    }

    fmt.Printf("\nBest variant: %s with cost %.4f\n", bestVariant, bestCost)
}
```

### Example 4: Configuration Presets

```go
package main

import (
    "fmt"
    "github.com/cwbudde/mayfly"
)

func main() {
    fmt.Println("=== Using Configuration Presets ===\n")

    // Automatically configure for deceptive landscapes
    config, err := mayfly.NewPresetConfig(mayfly.PresetDeceptive)
    if err != nil {
        panic(err)
    }

    // Just set the problem-specific parameters
    config.ObjectiveFunc = mayfly.Schwefel
    config.ProblemSize = 20
    config.LowerBound = -500
    config.UpperBound = 500

    result, err := mayfly.Optimize(config)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Using preset: PresetDeceptive\n")
    fmt.Printf("Algorithm: %s\n", "EOBBMA")  // Preset selects EOBBMA
    fmt.Printf("Best Cost: %.2f\n", result.GlobalBest.Cost)
    fmt.Printf("Function Evaluations: %d\n", result.FuncEvalCount)
}
```

### Example 5: Auto-Tuning Based on Problem Characteristics

```go
package main

import (
    "fmt"
    "github.com/cwbudde/mayfly"
)

func main() {
    fmt.Println("=== Auto-Tuning Configuration ===\n")

    // Start with a base configuration
    config := mayfly.NewOLCEConfig()
    config.ObjectiveFunc = mayfly.Rastrigin
    config.ProblemSize = 50  // High dimensionality
    config.LowerBound = -5.12
    config.UpperBound = 5.12

    // Define problem characteristics
    characteristics := mayfly.ProblemCharacteristics{
        Dimensionality:            50,
        Modality:                  mayfly.HighlyMultimodal,
        RequiresFastConvergence:   true,
        RequiresStableConvergence: false,
    }

    // Auto-tune the configuration
    mayfly.AutoTuneConfig(config, characteristics)

    // Configuration is now automatically adjusted:
    // - Population increased for high dimensionality
    // - Iterations optimized for fast convergence
    // - OLCE parameters tuned for multimodal landscape

    result, err := mayfly.Optimize(config)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Auto-tuned for:\n")
    fmt.Printf("  - High dimensionality (D=%d)\n", characteristics.Dimensionality)
    fmt.Printf("  - Highly multimodal landscape\n")
    fmt.Printf("  - Fast convergence requirement\n\n")

    fmt.Printf("Final population: %d males, %d females\n", config.NPop, config.NPopF)
    fmt.Printf("Max iterations: %d\n", config.MaxIterations)
    fmt.Printf("\nBest Cost: %.4f\n", result.GlobalBest.Cost)
}
```

## Running Examples

Complete examples available in the `examples/` directory:

### Algorithm Selection Demo

```bash
cd examples/selector && go run main.go
```

Shows:
- Automatic problem classification
- Algorithm recommendations with scores
- Performance comparison

### Benchmark Suite

```bash
cd examples/benchmark_suite && go run main.go
```

Runs comprehensive benchmarks across all variants and functions.

## Related Documentation

- [Configuration Guide](configuration.md) - Complete parameter reference
- [Algorithm Comparison](comparison-framework.md) - Statistical comparison tools
- [Algorithm Variants](../algorithms/) - Individual algorithm documentation
