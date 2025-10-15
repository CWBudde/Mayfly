# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go implementation of the Mayfly Optimization Algorithm (MA) and its enhanced variants. This is a metaheuristic optimization library converted from MATLAB while maintaining research fidelity. The project is evolving into a comprehensive suite of Mayfly algorithm variants.

**Current Status**: Phase 1 complete (Standard MA + DESMA variant)
**Future Phases**: See PLAN.md for roadmap to implement 7+ additional variants

## Build & Development Commands

### Using Just (Task Runner)
```bash
# View all available commands
just

# Build the project
just build

# Run tests with coverage report (generates coverage.html)
just test

# Quick test without coverage
just test-quick

# Run benchmarks
just bench

# Run main examples
just run

# Run algorithm comparison
cd examples/comparison && go run main.go

# Format code
just fmt

# Lint (requires golangci-lint)
just lint

# Run full CI pipeline (format, lint, test)
just ci

# Clean build artifacts
just clean
```

### Direct Go Commands
```bash
# Build
go build -v ./...

# Test with race detection
go test -v -race ./...

# Run specific test
go test -v -run TestFunctionName

# Benchmark specific function
go test -bench=BenchmarkName -benchmem

# Run examples
cd examples && go run main.go
```

### Module Management
```bash
# Tidy dependencies
go mod tidy

# Verify dependencies
go mod verify

# Update all dependencies
go get -u ./...
```

## Architecture & Core Concepts

### Dual-Population Structure

The algorithm maintains two distinct populations with different behaviors:

**Male Population** (`males []*Mayfly`)
- Track personal best positions (`Mayfly.Best`)
- Update velocity based on personal best and global best
- Perform "nuptial dance" when at optimum
- Formula: `v = g*v + a1*exp(-β*r_pb²)*(pbest - x) + a2*exp(-β*r_gb²)*(gbest - x)`

**Female Population** (`females []*Mayfly`)
- No personal best tracking
- Attracted to males with better fitness OR fly randomly
- Female behavior: If `female.Cost > male.Cost` → move toward male, else random flight
- Formula: `v = g*v + a3*exp(-β*r_mf²)*(male - x)` or `v = g*v + fl*e`

### Main Optimization Flow

```
1. Initialize populations (males/females) randomly
2. For each iteration:
   a. Update females (attraction or random flight)
   b. Update males (personal/global best or nuptial dance)
   c. Sort populations by fitness
   d. Mating: Crossover between best males/females → offspring
   e. Mutation: Apply Gaussian mutation to offspring
   f. Merge offspring into populations
   g. Keep best individuals (selection)
   h. [DESMA only] Generate elite solutions around global best
   i. Update damping parameters (g, dance, fl)
3. Return best solution found
```

### DESMA Enhancement (Dynamic Elite Strategy)

Located in `mayfly.go:473-510` and `generateEliteMayflies()` at line 543.

**Key mechanism**: After selection, generates `EliteCount` (default: 5) candidate solutions around global best within `SearchRange`. Replaces worst male if elite is better.

**Adaptive search range**:
- If improving: `SearchRange *= EnlargeFactor` (default 1.05)
- If stagnating: `SearchRange *= ReductionFactor` (default 0.95)

**When to use**: DESMA excels on multimodal functions (Rastrigin, Rosenbrock) with 70%+ improvement over standard MA. Minimal overhead (~8% more function evaluations).

### Configuration System

All parameters centralized in `Config` struct. Use factory functions:

```go
// Standard MA
config := mayfly.NewDefaultConfig()

// DESMA variant
config := mayfly.NewDESMAConfig()
```

**Required fields** (must set before calling `Optimize()`):
- `ObjectiveFunc` - Function to minimize
- `ProblemSize` - Number of dimensions
- `LowerBound` / `UpperBound` - Search space bounds

**Auto-calculated if zero**:
- `NM` (mutants) = 5% of `NPop`
- `VelMax` / `VelMin` = ±10% of bounds
- `SearchRange` (DESMA) = 10% of bounds

### Helper Functions (Private API)

Critical internal functions in `mayfly.go`:

- `newMayfly(size)` - Constructor with pre-allocated slices
- `unifrnd()` / `unifrndVec()` - Uniform random generation
- `randn()` - Normal distribution generation
- `maxVec()` / `minVec()` - Element-wise boundary clamping (in-place)
- `sortMayflies()` - Bubble sort (appropriate for small populations)
- `Crossover()` - Genetic crossover operator (exported)
- `Mutate()` - Gaussian mutation operator (exported)

**Boundary handling**: Always applied after velocity updates and genetic operations via `maxVec()`/`minVec()`.

## Benchmark Functions

Located in `functions.go`. All functions are **minimization** problems with known global minima:

### Classic Benchmark Functions

| Function | Global Min | Optimal Point | Bounds | Type |
|----------|-----------|---------------|--------|------|
| Sphere | 0 | (0,...,0) | [-10,10] | Unimodal, convex |
| Rastrigin | 0 | (0,...,0) | [-5.12,5.12] | Highly multimodal |
| Rosenbrock | 0 | (1,...,1) | [-5,10] | Unimodal, narrow valley |
| Ackley | 0 | (0,...,0) | [-32.768,32.768] | Multimodal, flat outer |
| Griewank | 0 | (0,...,0) | [-600,600] | Many local minima |

### CEC-Style Benchmark Functions

| Function | Global Min | Optimal Point | Bounds | Type |
|----------|-----------|---------------|--------|------|
| Schwefel | 0 | (420.97,...,420.97) | [-500,500] | Highly multimodal, deceptive |
| Levy | 0 | (1,...,1) | [-10,10] | Multimodal |
| Zakharov | 0 | (0,...,0) | [-10,10] | Unimodal, polynomial |
| DixonPrice | 0 | Special pattern* | [-10,10] | Unimodal, valley |
| Michalewicz | -9.66 (10D) | Variable | [0,π] | Multimodal, steep valleys |
| BentCigar | 0 | (0,...,0) | [-100,100] | Unimodal, ill-conditioned |
| Discus | 0 | (0,...,0) | [-100,100] | Unimodal, ill-conditioned |
| Weierstrass | 0 | (0,...,0) | [-0.5,0.5] | Continuous, non-differentiable |
| HappyCat | 0 | (-1,...,-1) | [-2,2] | Multimodal, plate-shaped |
| ExpandedSchafferF6 | 0 | (0,...,0) | [-100,100] | Multimodal, composite |

*DixonPrice optimum: x_i = 2^(-(2^i - 2)/2^i)

**Performance expectations** (500 iterations):
- Sphere: ~1e-5 to 1e-10
- Rastrigin: 30-100 (multimodal, harder)
- Rosenbrock: 0.1-10 (narrow valley challenge)
- Schwefel: High variance due to deceptive landscape
- BentCigar/Discus: Test ill-conditioning handling

## Module Structure

```
.
├── mayfly.go              # Core algorithm (533 lines)
│   ├── Config struct      # All parameters
│   ├── Optimize()         # Main entry point
│   └── generateEliteMayflies()  # DESMA logic
├── functions.go           # Benchmark functions (15 total: 5 classic + 10 CEC-style)
├── functions_test.go      # Comprehensive function tests
├── benchmark_test.go      # Performance benchmark suite
├── go.mod                 # Root module
├── PLAN.md               # Development roadmap (10 phases)
├── justfile              # Task runner recipes
└── examples/
    ├── main.go           # Basic usage demo
    ├── go.mod            # Local replace directive
    └── comparison/
        └── main.go       # MA vs DESMA comparison
```

**Module replacement**: Examples use `replace github.com/cwbudde/mayfly => ../` in go.mod for local development.

## Testing Strategy (Future - Phase 1)

Per PLAN.md, comprehensive testing planned:

### Test Coverage Goals
- Unit tests: 90%+ coverage
- Integration tests: All variants
- Benchmark tests: All functions + CEC suites
- Regression tests: Performance baselines

### Test Structure (To be created)
```
mayfly_test.go          # Core algorithm tests
functions_test.go       # Benchmark function validation
benchmark_test.go       # Performance benchmarks
desma_test.go          # DESMA-specific tests
operators_test.go      # Crossover/Mutate tests
helpers_test.go        # Helper function tests
```

### Running Tests (When implemented)
```bash
# Run all tests
go test -v ./...

# Run with coverage
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific test
go test -v -run TestOptimizeSphere

# Benchmarks
go test -bench=. -benchmem
```

## Development Guidelines

### Adding New Algorithm Variants

Follow the pattern established by DESMA:

1. **Add parameters to `Config` struct** with `Use<Variant>` boolean flag
2. **Create factory function** `NewXXXConfig()` that returns configured `Config`
3. **Implement variant logic** in `Optimize()` function or separate helper
4. **Add to examples** with comparison against standard MA
5. **Document in README** with usage section and parameter descriptions

Example pattern:
```go
// In Config struct
UseNewVariant bool
NewVariantParam1 float64
NewVariantParam2 int

// Factory function
func NewVariantConfig() *Config {
    config := NewDefaultConfig()
    config.UseNewVariant = true
    config.NewVariantParam1 = 0.5
    return config
}

// In Optimize() main loop
if config.UseNewVariant {
    // Apply variant-specific logic
}
```

### Extending Benchmark Functions

Add to `functions.go` following the pattern:

```go
// FunctionName is a benchmark function.
// Global minimum is at f(x*, ..., x*) = optimal_value
// Typical bounds: [lower, upper]
func FunctionName(x []float64) float64 {
    // Implementation
    return result
}
```

### Reproducible Results

Use `Config.Rand` for deterministic behavior:

```go
import "math/rand"

config := mayfly.NewDefaultConfig()
config.Rand = rand.New(rand.NewSource(42))  // Fixed seed
result, _ := mayfly.Optimize(config)
```

### Maximization Problems

Algorithm minimizes by default. For maximization, negate the objective:

```go
func maximizeProfit(x []float64) float64 {
    profit := calculateProfit(x)
    return -profit  // Negate for maximization
}
```

## Future Development (PLAN.md Phases)

**Next Priority: Phase 1 - Testing Infrastructure**
- Create comprehensive test suite
- Establish CI/CD with GitHub Actions
- Benchmark against reference implementations

**Upcoming Variants** (Phases 2-6):
- OLCE-MA: Orthogonal learning + chaotic exploitation
- EOBBMA: Elite opposition-based with Lévy flight
- GSASMA: Golden sine + simulated annealing
- MPMA: Median position-based
- AOBLMOA: Multi-objective with Aquila optimizer

**Framework Unification** (Phase 7):
- Common variant interface
- Algorithm selection helper
- Comparison framework
- Configuration management (JSON/YAML)

## Research Citations

When implementing variants, maintain research fidelity and cite properly:

1. **Original MA**: Zervoudakis & Tsafarakis (2020). *Computers & Industrial Engineering*, 145, 106559.
2. **DESMA**: *PLOS One*, 2022. (Dynamic elite strategy)
3. **Future variants**: See PLAN.md references section

## Common Pitfalls

1. **Forgetting to set required Config fields**: `ObjectiveFunc`, `ProblemSize`, `LowerBound`, `UpperBound` must be set before calling `Optimize()`

2. **Not using factory functions**: Always start with `NewDefaultConfig()` or `NewDESMAConfig()` to get sensible defaults

3. **Boundary violations**: Helper functions like `Crossover()` and `Mutate()` handle boundaries, but custom operators must call `maxVec()`/`minVec()`

4. **Population size too small**: Default 20 works for simple problems, increase `NPop`/`NPopF` for complex/high-dimensional problems

5. **Comparing variants unfairly**: Use same iteration count and random seed for fair comparison. DESMA uses ~8% more function evaluations.

## Performance Profiling

```bash
# CPU profiling
cd examples
go run -cpuprofile=cpu.prof main.go
go tool pprof cpu.prof

# Memory profiling
go run -memprofile=mem.prof main.go
go tool pprof mem.prof

# Or use justfile
just profile-cpu
just profile-mem
```

## Key Files Reference

- `mayfly.go:244-526` - Main `Optimize()` function
- `mayfly.go:473-510` - DESMA elite strategy
- `mayfly.go:543-583` - Elite generation mechanism
- `mayfly.go:128-139` - Mayfly constructor
- `mayfly.go:201-242` - Genetic operators
- `justfile` - All build/test commands
- `PLAN.md` - Full development roadmap
- `.github/copilot-instructions.md` - Original AI guidance (historical)
