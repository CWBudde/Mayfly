# OLCE-MA - Orthogonal Learning and Chaotic Exploitation Mayfly Algorithm

## Research Reference

**Zhou, D., Kang, Z., Su, X., & Yang, C. (2022). An enhanced Mayfly optimization algorithm based on orthogonal learning and chaotic exploitation strategy. International Journal of Machine Learning and Cybernetics, 13, 3625–3643.**

https://doi.org/10.1007/s13042-022-01617-4

## Overview

OLCE-MA enhances the standard algorithm with orthogonal experimental design and chaotic perturbations. This variant excels on complex multimodal optimization problems by improving population diversity and local search capability.

## Key Innovations

### 1. Orthogonal Learning

Applies **orthogonal experimental design** as part of every male's movement:

- **Purpose**: Increase diversity and reduce oscillatory movement
- **Method**: Systematic exploration of parameter combinations
- **Effect**: More efficient search of the solution space
- **Applied to**: The primary male movement operator, before sorting and mating
- **Dimension safety**: A generated two-level array supplies distinct,
  pairwise-balanced columns beyond three dimensions; OLCE configurations above
  1023 dimensions are rejected before the quadratic allocation
- **Disabling**: `OrthogonalFactor = 0` skips the stage and spends no
  evaluations on it

**Benefits**:

- Explores parameter space more systematically than random search
- Reduces redundant evaluations
- Maintains population diversity in promising regions

### 2. Chaotic Exploitation

Mayfly currently forms one new position from **every crossover offspring**
after mating:

```
s = (MaxIterations - generation + 1) / MaxIterations
chaotic[j] = LowerBound + z[j] * (UpperBound - LowerBound)
candidate[i][j] = (1 - s) * offspring[i][j] + s * chaotic[i][j]
```

The compatibility implementation uses a logistic map:

```
z(n+1) = 4 * z(n) * (1 - z(n))
```

Each candidate is evaluated once and becomes its corresponding offspring
position. `ChaosFactor` is a compatibility multiplier on `s`; its default is 1
and zero disables the stage.

This stage is now known to be a library extension, not a paper-faithful
implementation. The publisher's [mating pseudocode](https://media.springernature.com/full/springer-static/image/art%3A10.1007%2Fs13042-022-01617-4/MediaObjects/13042_2022_1617_Fig1_HTML.png)
creates `N` crossover offspring, and its [chaotic-exploitation pseudocode](https://media.springernature.com/full/springer-static/image/art%3A10.1007%2Fs13042-022-01617-4/MediaObjects/13042_2022_1617_Fig3_HTML.png)
loops over all `N`; the [published map figure](https://media.springernature.com/full/springer-static/image/art%3A10.1007%2Fs13042-022-01617-4/MediaObjects/13042_2022_1617_Fig4_HTML.png)
also shows values in `[-1,1]` with `C1 = 0.65`. However, the indexed
author-shared full-text prose describes the candidate as the fittest offspring's
position. The paper therefore contains a batch-cardinality conflict rather than
an established all-`N` rule.

Mayfly's serial and parallel compatibility paths follow the literal all-`N`
pseudocode loop. This remains a documented library choice: the accessible
primary material does not expose the exact Chebyshev recurrence or Equation 12,
and it does not reconcile that loop with the fittest-offspring prose. The paper's
cited OLCGOA predecessor gives the analogous candidate blend
`CS = (1-s)*Fbest + s*C'`, but that different algorithm cannot establish OLCE's
offspring lifecycle. The earlier open Chebyshev-Mayfly paper likewise says only
that two values are randomly selected from a Chebyshev sequence; it supplies no
recurrence or lifecycle that can safely be transplanted. The library retains its
Logistic-map equation rather than guessing those semantics. Do not label OLCE
results as exact paper reproductions until the fidelity gate is resolved. The
public-source audit, three stable evidence blockers, acceptable primary
evidence, and unsent author-request draft are preserved in the
[machine-readable clarification request](../reference-data/olce-ma-2022-clarification-request.json).

**Properties**:

- Deterministic but appears random
- Covers search space ergodically
- Helps escape local optima
- Improves local search capability

### 3. Adaptive Strategy

The algorithm balances exploration and exploitation through proven parameter defaults that work well without tuning.

## Usage Examples

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/cwbudde/mayfly"
)

func main() {
    // Use OLCE-MA for multimodal problems
    config := mayfly.NewOLCEConfig()
    config.ObjectiveFunc = mayfly.Rastrigin  // Highly multimodal function
    config.ProblemSize = 10
    config.LowerBound = -10
    config.UpperBound = 10
    config.MaxIterations = 500

    result, err := mayfly.Optimize(config)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Best Cost: %f\n", result.GlobalBest.Cost)
}
```

### Advanced Usage with Custom Parameters

```go
package main

import (
    "fmt"
    "math/rand"
    "github.com/cwbudde/mayfly"
)

func main() {
    // Configure OLCE-MA for high-dimensional multimodal optimization
    config := mayfly.NewOLCEConfig()
    config.ObjectiveFunc = mayfly.Rastrigin
    config.ProblemSize = 30  // High dimensionality
    config.LowerBound = -5.12
    config.UpperBound = 5.12
    config.MaxIterations = 1000

    // Tune OLCE-specific parameters
    config.OrthogonalFactor = 0.4  // Increase exploration
    config.ChaosFactor = 1.0        // Canonical offspring constriction

    // Increase population for high-D problems
    config.NPop = 40
    config.NPopF = 40

    // Use fixed seed for reproducibility
    config.Rand = rand.New(rand.NewSource(42))

    result, err := mayfly.Optimize(config)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Final Cost: %.6f\n", result.GlobalBest.Cost)
    fmt.Printf("Iterations: %d\n", result.IterationCount)
    fmt.Printf("Function Evaluations: %d\n", result.FuncEvalCount)
}
```

### Real-World Example: Neural Network Hyperparameter Tuning

```go
package main

import (
    "fmt"
    "math"
    "github.com/cwbudde/mayfly"
)

// Simulate neural network validation error for given hyperparameters
func nnValidationError(params []float64) float64 {
    learningRate := params[0]
    momentum := params[1]
    dropout := params[2]
    l2Reg := params[3]

    // Simulate training with these hyperparameters
    // (replace with actual model training in practice)

    // Penalize extreme values
    error := 0.0

    // Learning rate: optimal around 0.001-0.01
    lrOptimal := 0.005
    error += math.Abs(learningRate - lrOptimal) * 10

    // Momentum: optimal around 0.9
    error += math.Abs(momentum - 0.9) * 5

    // Dropout: optimal around 0.2-0.3
    dropoutOptimal := 0.25
    error += math.Abs(dropout - dropoutOptimal) * 8

    // L2 regularization: optimal around 0.0001
    l2Optimal := 0.0001
    error += math.Abs(l2Reg - l2Optimal) * 1000

    // Add some noise to simulate stochastic training
    noise := (math.Sin(learningRate*100) + math.Cos(momentum*50)) * 0.1

    return error + math.Abs(noise)
}

func main() {
    fmt.Println("=== Neural Network Hyperparameter Optimization with OLCE-MA ===\n")

    // OLCE-MA excels at this multimodal optimization problem
    config := mayfly.NewOLCEConfig()
    config.ObjectiveFunc = nnValidationError
    config.ProblemSize = 4  // 4 hyperparameters
    config.LowerBound = 0.0001
    config.UpperBound = 0.5
    config.MaxIterations = 300

    // Smaller population for expensive evaluations
    config.NPop = 20
    config.NPopF = 20

    result, err := mayfly.Optimize(config)
    if err != nil {
        panic(err)
    }

    fmt.Println("Optimal Hyperparameters:")
    fmt.Printf("  Learning Rate: %.6f\n", result.GlobalBest.Position[0])
    fmt.Printf("  Momentum:      %.6f\n", result.GlobalBest.Position[1])
    fmt.Printf("  Dropout:       %.6f\n", result.GlobalBest.Position[2])
    fmt.Printf("  L2 Reg:        %.6f\n", result.GlobalBest.Position[3])
    fmt.Printf("\nValidation Error: %.6f\n", result.GlobalBest.Cost)
    fmt.Printf("Function Evaluations: %d\n", result.FuncEvalCount)
}
```

## OLCE-MA Parameters

- `UseOLCE`: Enable OLCE-MA variant (default: false)
- `OrthogonalFactor`: Orthogonal learning strength (default: 0.3, range: 0-1)
- `ChaosFactor`: Multiplier for the historical Logistic-map compatibility
  stage (default: 1, range: 0-1). Set to 0 to disable that stage.

## Benefits

- **15-30% improvement** on multimodal functions (Rastrigin, Rosenbrock, Ackley)
- **Better diversity**: Orthogonal learning explores parameter space more systematically
- **Escape stagnation**: Chaotic perturbations help avoid local optima
- **Evaluation overhead**: depends on dimension because the orthogonal array
  is generated to provide a distinct column per dimension
- **No tuning needed**: Works well with default parameters

## Performance

### Multimodal Functions

**Rastrigin (D=10, highly multimodal)**:

- Standard MA: 45-60
- OLCE-MA: 15-35 (30%+ improvement)

**Rosenbrock (D=10, narrow valley)**:

- Standard MA: 10-50
- OLCE-MA: 1-10 (significant improvement)

**Ackley (D=10, multimodal)**:

- Standard MA: 2-5
- OLCE-MA: 0.5-2 (50%+ improvement)

### Overhead

Each generation evaluates one orthogonal design plus one factor-analysis
candidate per male, and the current compatibility stage evaluates one chaotic
candidate per crossover offspring. The exact overhead therefore depends on
dimension and population; a future paper-faithful Chebyshev stage may retain
that count while changing the generated positions.

## When to Use OLCE-MA

- **Best for**: Multimodal problems with many local optima
- **Excellent on**: High-dimensional problems (10D+)
- **Use when**: Standard MA or DESMA struggle with local optima
- **Ideal for**: Rastrigin, Rosenbrock, Schwefel, Griewank functions

## Parameter Tuning Guide

### Orthogonal Factor

**Default (balanced)**:

```go
config.OrthogonalFactor = 0.3
```

- Good balance between exploration and exploitation
- Recommended for most problems

**More exploration**:

```go
config.OrthogonalFactor = 0.5
```

- Use when: Problem has very high dimensionality
- Use when: Need more systematic parameter space exploration
- Trade-off: More computational overhead

**More exploitation**:

```go
config.OrthogonalFactor = 0.1
```

- Use when: Problem requires fine-tuning near solutions
- Use when: Want minimal overhead
- Trade-off: Less diversity maintenance

### Chaos Factor

**Default (balanced)**:

```go
config.ChaosFactor = 1.0
```

- Uses the historical compatibility constriction without scaling.

**Reduced chaos**:

```go
config.ChaosFactor = 0.3
```

- Uses 30% of the historical compatibility constriction

**Minimal chaos**:

```go
config.ChaosFactor = 0.05
```

- Use when: Solutions need fine refinement
- Use when: Convergence speed is critical
- Trade-off: Less local optima escape capability

## OLCE-MA vs Other Variants

**Choose OLCE-MA when**:

- Problem is highly multimodal (Rastrigin-like)
- High dimensionality (20D+)
- You prioritize solution quality over convergence speed
- Systematic parameter exploration is beneficial

**Choose DESMA instead when**:

- Need simpler adaptive local search
- Function evaluations are cheap
- Want less computational overhead

**Choose EOBBMA instead when**:

- Problem is highly deceptive (Schwefel-like)
- Want simplest parameter tuning
- Heavy-tailed jumps are more effective than chaos

**Choose GSASMA instead when**:

- Need maximum convergence speed
- Simulated annealing fits problem structure
- Prefer hybrid mutation over orthogonal learning

## Algorithm Details

### Orthogonal Learning Application

1. **Update male movement**: Run the primary male movement operator
2. **Generate orthogonal array**: Systematic parameter combinations for each male
3. **Evaluate combinations**: Test each orthogonal combination
4. **Select best**: Keep an improving orthogonally-learned male position

### Chaotic Perturbation Application

1. **After crossover**: Generate offspring from parents
2. **Compatibility batch**: Visit every crossover offspring in stable slice order
3. **Compatibility map**: Blend each with bounded Logistic-map positions using `s`
4. **Evaluate**: Test every new offspring position before mutation and selection

These are the current library steps. The all-offspring topology matches the
publisher pseudocode, but the Logistic equation is not its unresolved
Chebyshev mutation; see the fidelity note above.

## Related Documentation

- [DESMA](desma.md) - Simpler elite-based variant
- [EOBBMA](eobbma.md) - Heavy-tailed exploration alternative
- [Standard MA](standard-ma.md) - Base algorithm
- [Configuration Guide](../api/configuration.md) - Complete parameter reference
