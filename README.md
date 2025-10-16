# Mayfly Optimization Algorithm (Go)

A Go implementation of the Mayfly Optimization Algorithm (MA), a nature-inspired metaheuristic optimization algorithm based on the mating behavior of mayflies.

[![Go Reference](https://pkg.go.dev/badge/github.com/CWBudde/mayfly.svg)](https://pkg.go.dev/github.com/CWBudde/mayfly)
[![Go Report Card](https://goreportcard.com/badge/github.com/CWBudde/mayfly)](https://goreportcard.com/report/github.com/CWBudde/mayfly)

## Overview

The Mayfly Algorithm is a swarm intelligence optimization algorithm inspired by the flight behavior and mating process of mayflies. This implementation includes the standard algorithm and 6 enhanced variants for different optimization scenarios.

**Key Features:**
- Clean, idiomatic Go implementation
- 7 algorithm variants for different problem types
- 15+ benchmark functions included
- Unified API with intelligent algorithm selection
- Statistical comparison framework
- Thread-safe with proper configuration

## Quick Start

### Installation

```bash
go get github.com/CWBudde/mayfly
```

### Basic Usage

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

### Custom Objective Function

```go
func myFunction(x []float64) float64 {
    sum := 0.0
    for _, val := range x {
        sum += val * val
    }
    return sum
}

config := mayfly.NewDefaultConfig()
config.ObjectiveFunc = myFunction
config.ProblemSize = 10
config.LowerBound = -5
config.UpperBound = 5
```

## Algorithm Variants

| Variant | Best For | Improvement | Key Features |
|---------|----------|-------------|--------------|
| **[Standard MA](docs/algorithms/standard-ma.md)** | General problems | Baseline | Balanced, well-tested |
| **[DESMA](docs/algorithms/desma.md)** | Multimodal | +70% | Adaptive elite search |
| **[OLCE-MA](docs/algorithms/olce-ma.md)** | Highly multimodal | +15-30% | Orthogonal learning + chaos |
| **[EOBBMA](docs/algorithms/eobbma.md)** | Deceptive landscapes | +55% | Lévy flights, Bare Bones |
| **[GSASMA](docs/algorithms/gsasma.md)** | Fast convergence | +10-20% | Golden Sine + Simulated Annealing |
| **[MPMA](docs/algorithms/mpma.md)** | Stable convergence | +10-30% | Median guidance, robust |
| **[AOBLMOA](docs/algorithms/aoblmoa.md)** | Adaptive/Multi-objective | Variable | 4 hunting strategies |

### Using Variants

```go
// Standard MA (baseline)
config := mayfly.NewDefaultConfig()

// DESMA (better for multimodal)
config := mayfly.NewDESMAConfig()

// OLCE-MA (best for highly multimodal)
config := mayfly.NewOLCEConfig()

// EOBBMA (best for deceptive)
config := mayfly.NewEOBBMAConfig()

// GSASMA (fastest convergence)
config := mayfly.NewGSASMAConfig()

// MPMA (most stable)
config := mayfly.NewMPMAConfig()

// AOBLMOA (adaptive)
config := mayfly.NewAOBLMOAConfig()
```

## Intelligent Algorithm Selection

Let the library recommend the best algorithm for your problem:

```go
// Define problem characteristics
chars := mayfly.ProblemCharacteristics{
    Dimensionality:            30,
    Modality:                  mayfly.HighlyMultimodal,
    Landscape:                 mayfly.Rugged,
}

// Get recommendation
selector := mayfly.NewAlgorithmSelector()
best := selector.RecommendBest(chars)

// Use recommended variant
result, err := mayfly.NewBuilderFromVariant(best).
    ForProblem(mayfly.Rastrigin, 30, -5.12, 5.12).
    WithIterations(500).
    Optimize()
```

Or use the fluent builder API directly:

```go
result, err := mayfly.NewBuilder("olce").
    ForProblem(mayfly.Rastrigin, 30, -5.12, 5.12).
    WithIterations(500).
    WithPopulation(30, 30).
    Optimize()
```

## Statistical Comparison

Compare algorithms with comprehensive statistical analysis:

```go
runner := mayfly.NewComparisonRunner().
    WithVariantNames("ma", "desma", "olce", "eobbma").
    WithRuns(30).
    WithIterations(500).
    WithVerbose(true)

result := runner.Compare(
    "Rastrigin",
    mayfly.Rastrigin,
    30, -5.12, 5.12,
)

result.PrintComparisonResults()
```

## Benchmark Functions

15+ standard test functions included:

**Classic Functions:** Sphere, Rastrigin, Rosenbrock, Ackley, Griewank

**CEC-Style Functions:** Schwefel, Levy, Zakharov, DixonPrice, Michalewicz, BentCigar, Discus, Weierstrass, HappyCat, ExpandedSchafferF6

See [Benchmark Functions](docs/benchmarks.md) for details.

## Documentation

### Getting Started
- **[Getting Started Guide](docs/getting-started.md)** - Tutorial and examples
- **[Configuration Guide](docs/api/configuration.md)** - Complete parameter reference
- **[Benchmark Functions](docs/benchmarks.md)** - Test functions and expected results

### Algorithms
- **[Standard MA](docs/algorithms/standard-ma.md)** - Original Mayfly Algorithm
- **[DESMA](docs/algorithms/desma.md)** - Dynamic Elite Strategy
- **[OLCE-MA](docs/algorithms/olce-ma.md)** - Orthogonal Learning & Chaotic Exploitation
- **[EOBBMA](docs/algorithms/eobbma.md)** - Elite Opposition-Based Bare Bones
- **[GSASMA](docs/algorithms/gsasma.md)** - Golden Sine with Simulated Annealing
- **[MPMA](docs/algorithms/mpma.md)** - Median Position-Based
- **[AOBLMOA](docs/algorithms/aoblmoa.md)** - Aquila Optimizer-Based Learning

### API Reference
- **[Unified Framework](docs/api/unified-framework.md)** - Builder API, algorithm selection, presets
- **[Comparison Framework](docs/api/comparison-framework.md)** - Statistical testing and analysis
- **[Configuration Guide](docs/api/configuration.md)** - All parameters explained

### Research
- **[Research References](docs/research.md)** - Academic papers and citations

## Running Examples

```bash
# Basic usage
cd examples && go run main.go

# Algorithm comparison
cd examples/comparison && go run main.go

# Algorithm selector demo
cd examples/selector && go run main.go

# Comprehensive benchmark suite
cd examples/benchmark_suite && go run main.go
```

## Build Commands

Using [Just](https://github.com/casey/just) task runner:

```bash
# View all commands
just

# Build the project
just build

# Run tests with coverage
just test

# Run benchmarks
just bench

# Format code
just fmt

# Run full CI pipeline
just ci
```

Or use Go commands directly:

```bash
go build -v ./...
go test -v -race ./...
go test -bench=. -benchmem
```

## Research & Citations

Based on the following research:

**Original Mayfly Algorithm:**
Zervoudakis, K., & Tsafarakis, S. (2020). A mayfly optimization algorithm. *Computers & Industrial Engineering*, 145, 106559. [DOI: 10.1016/j.cie.2020.106559](https://doi.org/10.1016/j.cie.2020.106559)

**OLCE-MA Variant:**
Zhou, D., et al. (2022). An enhanced Mayfly optimization algorithm based on orthogonal learning and chaotic exploitation strategy. *International Journal of Machine Learning and Cybernetics*, 13, 3625–3643. [DOI: 10.1007/s13042-022-01617-4](https://doi.org/10.1007/s13042-022-01617-4)

**Other Variants:**
- DESMA: *PLOS One*, 2022
- EOBBMA: *Arabian Journal for Science and Engineering*, 2024
- GSASMA: *Electronics Letters / IEEE*, 2022
- MPMA: *IEEE Access*, 2022
- AOBLMOA: *PubMed / Various journals*, 2023

See [Research References](docs/research.md) for complete citations.

## Performance

### Expected Results (D=30, 500 iterations)

| Function | Standard MA | DESMA | OLCE-MA | EOBBMA | Best Variant |
|----------|-------------|-------|---------|--------|--------------|
| Sphere | 1e-6 | 1e-8 | 1e-7 | 1e-6 | DESMA |
| Rastrigin | 55 | 40 | **30** | 38 | **OLCE-MA** |
| Rosenbrock | 25 | 15 | 12 | 18 | **MPMA (8)** |
| Schwefel | 850 | 650 | 600 | **350** | **EOBBMA** |

### Algorithm Overhead

| Variant | Additional Evaluations | When Worth It |
|---------|------------------------|---------------|
| DESMA | +8% | Multimodal problems |
| OLCE-MA | +12% | Highly multimodal |
| EOBBMA | +1.5% | Deceptive landscapes |
| GSASMA | +15% | Need fast convergence |
| MPMA | 0% (baseline) | Need stability |
| AOBLMOA | +20-30% | Complex/adaptive needs |

## Development Status

**Current:** Phase 1 complete - All 7 variants implemented

**Future:** See [PLAN.md](PLAN.md) for roadmap

## Contributing

Contributions welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass (`just test`)
5. Submit a pull request

## License

MIT License - See [LICENSE](LICENSE) file for details

## Support

- **Issues:** [GitHub Issues](https://github.com/CWBudde/mayfly/issues)
- **Documentation:** [docs/](docs/)
- **Examples:** [examples/](examples/)

---

**Quick Links:**
[Getting Started](docs/getting-started.md) |
[API Docs](docs/api/) |
[Algorithms](docs/algorithms/) |
[Benchmarks](docs/benchmarks.md) |
[Research](docs/research.md)
