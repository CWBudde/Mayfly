# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go implementation of the Mayfly Optimization Algorithm (MA) and its enhanced variants. This is a metaheuristic optimization library converted from MATLAB while maintaining research fidelity.

**Current Status**: All 7 algorithm variants implemented (Standard MA, DESMA, OLCE-MA, EOBBMA, GSASMA, MPMA, AOBLMOA)
**Documentation**: Complete docs/ folder with algorithm guides, API reference, and tutorials
**Future Phases**: See PLAN.md for roadmap

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

**Full documentation**: See `docs/benchmarks.md` for complete benchmark function reference.

Located in `functions.go`. All functions are **minimization** problems with known global minima.

### Quick Reference

**Classic Functions** (5): Sphere, Rastrigin, Rosenbrock, Ackley, Griewank
**CEC-Style Functions** (10): Schwefel, Levy, Zakharov, DixonPrice, Michalewicz, BentCigar, Discus, Weierstrass, HappyCat, ExpandedSchafferF6

**Performance expectations** (500 iterations, D=30):
- Sphere: ~1e-5 to 1e-10
- Rastrigin: 30-100 (multimodal, harder)
- Rosenbrock: 0.1-10 (narrow valley challenge)
- Schwefel: High variance due to deceptive landscape
- BentCigar/Discus: Test ill-conditioning handling

For detailed function descriptions, parameters, and expected results, see `docs/benchmarks.md`.

## Module Structure

```
.
├── mayfly.go              # Core algorithm
├── functions.go           # Benchmark functions (15 total)
├── variants.go            # Variant interface and implementations
├── selector.go            # Algorithm selection
├── comparison.go          # Statistical comparison framework
├── config.go              # Configuration structures
├── config_loader.go       # JSON config support
├── types.go               # Core type definitions
├── *_test.go              # Comprehensive test suite
├── benchmark_test.go      # Performance benchmarks
├── go.mod                 # Root module
├── README.md              # Concise overview (323 lines, was 1314)
├── CLAUDE.md              # Development guide (this file)
├── PLAN.md                # Development roadmap
├── justfile               # Task runner recipes
├── docs/                  # Documentation folder
│   ├── getting-started.md # Tutorial
│   ├── benchmarks.md      # Benchmark function reference
│   ├── research.md        # Academic citations
│   ├── algorithms/        # Algorithm-specific docs
│   │   ├── standard-ma.md
│   │   ├── desma.md
│   │   ├── olce-ma.md
│   │   ├── eobbma.md
│   │   ├── gsasma.md
│   │   ├── mpma.md
│   │   └── aoblmoa.md
│   └── api/               # API documentation
│       ├── configuration.md
│       ├── unified-framework.md
│       └── comparison-framework.md
└── examples/
    ├── main.go            # Basic usage
    ├── comparison/        # Algorithm comparison
    ├── selector/          # Algorithm selection demo
    └── benchmark_suite/   # Comprehensive benchmarks
```

**Module replacement**: Examples use `replace github.com/cwbudde/mayfly => ../` in go.mod for local development.

**Documentation structure**: All technical details have been moved to `docs/` folder for better organization.

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

**Full citations**: See `docs/research.md` for complete academic references.

All variants maintain research fidelity to original papers:

1. **Original MA**: Zervoudakis & Tsafarakis (2020). *Computers & Industrial Engineering*, 145, 106559.
2. **DESMA**: *PLOS One*, 2022
3. **OLCE-MA**: Zhou et al. (2022). *International Journal of Machine Learning and Cybernetics*, 13, 3625–3643
4. **EOBBMA**: *Arabian Journal for Science and Engineering*, 2024
5. **GSASMA**: *Electronics Letters / IEEE*, 2022
6. **MPMA**: *IEEE Access*, 2022
7. **AOBLMOA**: *PubMed / Various journals*, 2023

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

### Core Implementation
- `mayfly.go` - Main optimization algorithm
- `variants.go` - Variant interface and implementations
- `selector.go` - Algorithm selection logic
- `comparison.go` - Statistical comparison framework
- `config.go` / `config_loader.go` - Configuration management
- `functions.go` - Benchmark functions

### Documentation
- `README.md` - Concise project overview (323 lines)
- `docs/getting-started.md` - Tutorial and examples
- `docs/algorithms/` - Individual algorithm documentation
- `docs/api/` - API reference documentation
- `docs/benchmarks.md` - Benchmark function reference
- `docs/research.md` - Academic citations

### Development
- `justfile` - All build/test commands
- `PLAN.md` - Development roadmap
- `CLAUDE.md` - This file (development guide)
