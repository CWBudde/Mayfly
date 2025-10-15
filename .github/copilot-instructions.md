# Mayfly Optimization Algorithm - AI Coding Instructions

## Project Overview

This is a Go implementation of the Mayfly Optimization Algorithm (MA), a nature-inspired metaheuristic optimization algorithm. The project is a direct conversion from MATLAB to Go while maintaining the original research algorithm structure.

## Core Architecture

### Main Components

- **`mayfly.go`**: Core algorithm implementation with `Optimize()` function as main entry point
- **`functions.go`**: Standard benchmark functions (Sphere, Rastrigin, Rosenbrock, Ackley, Griewank)
- **`examples/`**: Demonstration module showing usage patterns with local replace directive

### Key Design Patterns

**Configuration-Driven Design**: All algorithm parameters are centralized in the `Config` struct. Use `NewDefaultConfig()` as starting point and override specific fields:

```go
config := mayfly.NewDefaultConfig()
config.ObjectiveFunc = myFunction
config.ProblemSize = 30
```

**Population-Based Structure**: Algorithm operates on two distinct populations (`males`, `females`) with different behavioral patterns. Males track personal best positions, females are attracted to males or fly randomly.

**Functional Interface**: Optimization problems defined via `ObjectiveFunction` type `func([]float64) float64` - always minimization problems.

## Development Workflows

### Testing the Algorithm

```bash
cd examples
go run main.go
```

This runs all benchmark functions with reduced iterations (500) for quick validation.

### Module Structure

- Root module: Core algorithm library
- Examples module: Uses local replace directive `replace github.com/cwbudde/mayfly => ../`
- No external dependencies beyond Go standard library

## Critical Implementation Details

### Memory Management

- All mayfly positions/velocities are pre-allocated slices
- Use `newMayfly(size)` constructor for proper initialization
- `clone()` method performs deep copy of mayfly state

### Algorithm-Specific Conventions

- **Population sizing**: `NPop` (males), `NPopF` (females), `NC` (offspring), `NM` (mutants auto-calculated as 5% of NPop)
- **Velocity clamping**: `VelMax/VelMin` auto-calculated as 10% of problem bounds
- **Parameter damping**: `GDamp`, `DanceDamp`, `FLDamp` reduce inertia/dance/flight coefficients over iterations
- **Sorting**: Simple bubble sort used (appropriate for small populations ~20-40)

### Boundary Handling

- `maxVec()`/`minVec()` helper functions clamp vectors element-wise
- Applied after velocity updates and crossover operations
- Critical for maintaining feasible solutions

### Randomization

- Optional `Config.Rand` field for reproducible results
- Functions like `unifrnd()`, `unifrndVec()`, `randn()` abstract random number generation
- If `Config.Rand` is nil, uses global `rand` package

## Common Extension Points

### Custom Objective Functions

Always implement as minimization problem. For maximization, negate the result:

```go
func maximizeProfit(x []float64) float64 {
    return -calculateProfit(x) // Negate for maximization
}
```

### Algorithm Parameters

- Increase `NPop`/`NPopF` for complex problems
- Adjust `A1`/`A2`/`A3` learning coefficients for convergence behavior
- Modify `Beta` (distance sight) for exploration/exploitation balance

## Testing and Validation

### Benchmark Functions

Each has known global minimum at specific points:

- `Sphere`: f(0,...,0) = 0, bounds [-10,10]
- `Rastrigin`: f(0,...,0) = 0, bounds [-5.12,5.12], highly multimodal
- `Rosenbrock`: f(1,...,1) = 0, bounds [-5,10], narrow valley
- `Ackley`: f(0,...,0) = 0, bounds [-32.768,32.768]
- `Griewank`: f(0,...,0) = 0, bounds [-600,600]

### Performance Expectations

- Sphere: Should converge to ~1e-10 within 500 iterations
- Complex functions: Expect slower convergence, focus on consistent improvement

## Research Context

Based on Zervoudakis & Tsafarakis (2020) paper. Maintain algorithm fidelity to original MATLAB implementation. When making improvements, preserve core behavioral patterns of male/female populations and mating mechanisms.
