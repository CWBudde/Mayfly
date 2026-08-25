# Benchmark Functions Reference

The library includes 18 standalone functions, the complete usable CEC2017 and
CEC2020 bound-constrained suites, and four constrained engineering-design problems.

## Function Categories

### Classic Benchmark Functions (7)

Standard test functions from optimization literature:

- **Sphere** - Unimodal, convex
- **Rastrigin** - Highly multimodal
- **Rosenbrock** - Unimodal, narrow valley
- **Ackley** - Multimodal, flat outer region
- **Griewank** - Many local minima
- **Eggcrate** - Fixed 2D, regularly oscillating
- **Beale** - Fixed 2D, non-convex

### CEC-Style Benchmark Functions (11)

Additional challenging functions from CEC competitions:

- **Schwefel** - Highly multimodal, deceptive
- **Levy** - Multimodal
- **Zakharov** - Unimodal, polynomial
- **DixonPrice** - Unimodal, valley
- **Michalewicz** - Multimodal, steep valleys
- **BentCigar** - Unimodal, ill-conditioned
- **Discus** - Unimodal, ill-conditioned
- **Weierstrass** - Continuous, non-differentiable
- **HappyCat** - Multimodal, plate-shaped
- **ExpandedSchafferF6** - Multimodal, composite
- **Himmelblau** - Multimodal, four equal global minima

### Official CEC2017 and CEC2020 Suites

`NewCEC2017Problem`, `CEC2017Suite`, `NewCEC2020Problem`, and `CEC2020Suite`
implement the numbered competition suites, including their shifts, rotations,
permutations, hybrid partitions, composition weights, biases, and evaluation budgets.
CEC2017 contains 29 usable functions because its organizers removed F2 for numerical
instability; CEC2020 contains ten.

The organizers did not attach a redistribution license to the transformation data, so
Mayfly does not silently copy it into the module. Download and extract the official
[CEC2017](https://github.com/P-N-Suganthan/CEC2017-BoundContrained) or
[CEC2020](https://github.com/P-N-Suganthan/2020-Bound-Constrained-Opt-Benchmark)
software, then pass an `fs.FS` rooted at either the archive or its `input_data` directory:

```go
data := os.DirFS("/path/to/CEC17_fast_pow")
problem, err := mayfly.NewCEC2017Problem(data, 10, 30)
if err != nil {
    log.Fatal(err)
}

config, err := problem.NewConfig(nil)
if err != nil {
    log.Fatal(err)
}
result, err := mayfly.Optimize(config)
```

Supported competition dimensions are 10, 30, 50, and 100 for CEC2017, and 5,
10, 15, and 20 for CEC2020. `BenchmarkCase.NewConfig` searches a normalized
`[0,1]^D` box; use `problem.Decode(result.GlobalBest.Position)` to recover the
suite coordinates.

Compatibility follows the released evaluator where it disagrees with descriptive prose.
In particular, CEC2017's Schaffer F7 reads its pre-rotation scratch and the released
"non-continuous" Rastrigin discards its rounded scratch. One numerical defect is not
reproduced: CEC2020 F7 at D=5 evaluates its one-dimensional elliptic partition normally
instead of dividing by zero and returning `NaN`.

### Engineering Design Suite

`EngineeringBenchmarkSuite` includes:

- tension/compression spring design;
- welded-beam design;
- mixed-variable pressure-vessel design; and
- mixed-variable speed-reducer design.

Each `BenchmarkCase` exposes physical bounds, objective, constraints, a published
reference solution, and defensive metadata. Pressure-vessel thickness indices and
speed-reducer tooth count are projected onto their discrete grids. Calling `NewConfig`
maps Mayfly's unit search box to the heterogeneous physical bounds and wraps the
constraints consistently; decode the returned best position before reporting physical
design variables.

## Function Details

### Sphere Function

```go
mayfly.Sphere(x []float64) float64
```

- **Global minimum**: f(0, ..., 0) = 0
- **Typical bounds**: [-10, 10]
- **Type**: Unimodal, convex
- **Characteristics**: Simplest test function, smooth gradient
- **Best variant**: Standard MA
- **Expected performance** (500 iter): 1e-5 to 1e-10

**Formula**: `f(x) = Σ(xi²)`

**Use for**: Testing basic convergence, baseline performance

---

### Rastrigin Function

```go
mayfly.Rastrigin(x []float64) float64
```

- **Global minimum**: f(0, ..., 0) = 0
- **Typical bounds**: [-5.12, 5.12]
- **Type**: Highly multimodal
- **Characteristics**: Many regularly distributed local minima
- **Best variant**: OLCE-MA
- **Expected performance** (500 iter): 30-100

**Formula**: `f(x) = 10n + Σ(xi² - 10cos(2πxi))`

**Use for**: Testing multimodal optimization, local optima escape

---

### Rosenbrock Function

```go
mayfly.Rosenbrock(x []float64) float64
```

- **Global minimum**: f(1, ..., 1) = 0
- **Typical bounds**: [-5, 10]
- **Type**: Unimodal, narrow valley
- **Characteristics**: Flat valley, hard to navigate
- **Best variant**: MPMA
- **Expected performance** (500 iter): 0.1-10

**Formula**: `f(x) = Σ(100(xi+1 - xi²)² + (xi - 1)²)`

**Use for**: Testing valley-following ability, ill-conditioning

---

### Ackley Function

```go
mayfly.Ackley(x []float64) float64
```

- **Global minimum**: f(0, ..., 0) = 0
- **Typical bounds**: [-32.768, 32.768]
- **Type**: Multimodal
- **Characteristics**: Nearly flat outer region, many local minima
- **Best variant**: OLCE-MA
- **Expected performance** (500 iter): 0.5-3

**Formula**:

```
f(x) = -20exp(-0.2√(Σxi²/n)) - exp(Σcos(2πxi)/n) + 20 + e
```

**Use for**: Testing exploration in flat regions

---

### Griewank Function

```go
mayfly.Griewank(x []float64) float64
```

- **Global minimum**: f(0, ..., 0) = 0
- **Typical bounds**: [-600, 600]
- **Type**: Multimodal
- **Characteristics**: Many local minima, product term creates interdependence
- **Best variant**: DESMA
- **Expected performance** (500 iter): 0.01-0.1

**Formula**: `f(x) = 1 + Σ(xi²/4000) - Π(cos(xi/√i))`

**Use for**: Testing ability to handle interdependent variables

---

### Eggcrate Function

```go
mayfly.Eggcrate(x []float64) float64
```

- **Global minimum**: f(0, 0) = 0
- **Typical bounds**: [-5, 5]
- **Type**: Fixed 2D, multimodal
- **Formula**: `f(x,y) = x² + y² + 25(sin²x + sin²y)`
- **Dimension handling**: Coordinates beyond the first two are ignored

This is F19 in Table 6 of the original MA paper.

---

### Beale Function

```go
mayfly.Beale(x []float64) float64
```

- **Global minimum**: f(3, 0.5) = 0
- **Typical bounds**: [-4.5, 4.5]
- **Type**: Fixed 2D, multimodal
- **Formula**: `(1.5-x+xy)² + (2.25-x+xy²)² + (2.625-x+xy³)²`
- **Dimension handling**: Coordinates beyond the first two are ignored

This is F20 in Table 6 of the original MA paper.

---

### Schwefel Function

```go
mayfly.Schwefel(x []float64) float64
```

- **Global minimum**: f(420.97, ..., 420.97) = 0
- **Typical bounds**: [-500, 500]
- **Type**: Highly multimodal, deceptive
- **Characteristics**: Global minimum far from origin, misleading gradients
- **Best variant**: EOBBMA
- **Expected performance** (500 iter): High variance (100-1000)

**Formula**: `f(x) = 418.9829n - Σ(xi·sin(√|xi|))`

**Use for**: Testing deceptive landscape handling

---

### Levy Function

```go
mayfly.Levy(x []float64) float64
```

- **Global minimum**: f(1, ..., 1) = 0
- **Typical bounds**: [-10, 10]
- **Type**: Multimodal
- **Characteristics**: Multiple local minima, similar to Rastrigin
- **Best variant**: OLCE-MA
- **Expected performance** (500 iter): 0.01-1.0

**Formula**: Complex formula involving wi = 1 + (xi-1)/4

**Use for**: Alternative multimodal test

---

### Zakharov Function

```go
mayfly.Zakharov(x []float64) float64
```

- **Global minimum**: f(0, ..., 0) = 0
- **Typical bounds**: [-10, 10]
- **Type**: Unimodal, polynomial
- **Characteristics**: Plate-shaped, easy for most algorithms
- **Best variant**: Standard MA
- **Expected performance** (500 iter): 1e-3 to 1e-6

**Formula**: `f(x) = Σ(xi²) + (Σ(0.5i·xi))² + (Σ(0.5i·xi))⁴`

**Use for**: Sanity check, should be easy

---

### DixonPrice Function

```go
mayfly.DixonPrice(x []float64) float64
```

- **Global minimum**: f(x*) = 0, where x*i = 2^(-(2^i - 2)/2^i)
- **Typical bounds**: [-10, 10]
- **Type**: Unimodal, valley
- **Characteristics**: Narrow ridge leading to minimum
- **Best variant**: MPMA
- **Expected performance** (500 iter): 0.1-5

**Formula**: `f(x) = (x1 - 1)² + Σ(i(2xi² - xi-1)²)`

**Use for**: Testing ridge-following capability

---

### Michalewicz Function

```go
mayfly.Michalewicz(x []float64) float64
```

- **Global minimum**: f(x\*) ≈ -9.66 (10D)
- **Typical bounds**: [0, π]
- **Type**: Multimodal, steep valleys
- **Characteristics**: Deep valleys, dimension-dependent minimum
- **Best variant**: EOBBMA
- **Expected performance** (500 iter): -9.0 to -9.5 (10D)

**Formula**: `f(x) = -Σ(sin(xi)·sin^20((i·xi²)/π))`

**Use for**: Testing steep valley navigation

---

### BentCigar Function

```go
mayfly.BentCigar(x []float64) float64
```

- **Global minimum**: f(0, ..., 0) = 0
- **Typical bounds**: [-100, 100]
- **Type**: Unimodal, ill-conditioned
- **Characteristics**: One dimension much more sensitive
- **Best variant**: MPMA
- **Expected performance** (500 iter): 1e2 to 1e4

**Formula**: `f(x) = x1² + 10^6·Σ(xi²)` for i=2..n

**Use for**: Testing ill-conditioning handling

---

### Discus Function

```go
mayfly.Discus(x []float64) float64
```

- **Global minimum**: f(0, ..., 0) = 0
- **Typical bounds**: [-100, 100]
- **Type**: Unimodal, ill-conditioned
- **Characteristics**: First dimension highly sensitive
- **Best variant**: MPMA
- **Expected performance** (500 iter): 1e2 to 1e4

**Formula**: `f(x) = 10^6·x1² + Σ(xi²)` for i=2..n

**Use for**: Testing sensitivity to conditioning

---

### Weierstrass Function

```go
mayfly.Weierstrass(x []float64) float64
```

- **Global minimum**: f(0, ..., 0) = 0
- **Typical bounds**: [-0.5, 0.5]
- **Type**: Continuous, non-differentiable
- **Characteristics**: Fractal-like, no smooth gradient
- **Best variant**: EOBBMA
- **Expected performance** (500 iter): 0.1-1.0

**Formula**: Complex sum of cosine terms with a=0.5, b=3, kmax=20

**Use for**: Testing gradient-free optimization

---

### HappyCat Function

```go
mayfly.HappyCat(x []float64) float64
```

- **Global minimum**: f(-1, ..., -1) = 0
- **Typical bounds**: [-2, 2]
- **Type**: Multimodal, plate-shaped
- **Characteristics**: Relatively flat, multiple local minima
- **Best variant**: GSASMA
- **Expected performance** (500 iter): 0.1-2.0

**Formula**: `f(x) = (|Σxi² - n|)^0.25 + (0.5Σxi² + Σxi)/n + 0.5`

**Use for**: Testing exploration on plateaus

---

### ExpandedSchafferF6 Function

```go
mayfly.ExpandedSchafferF6(x []float64) float64
```

- **Global minimum**: f(0, ..., 0) = 0
- **Typical bounds**: [-100, 100]
- **Type**: Multimodal, composite
- **Characteristics**: Composition of 2D Schaffer F6 functions
- **Best variant**: AOBLMOA
- **Expected performance** (500 iter): 0.5-5.0

**Formula**: Sum of 2D Schaffer F6 functions applied to consecutive pairs

**Use for**: Testing composite function optimization

---

### Himmelblau Function

```go
mayfly.Himmelblau(x []float64) float64
```

- **Global minimum**: f(3, 2, ..., 3, 2) = 0, with a trailing 0 in odd dimensions
- **Typical bounds**: [-5, 5]
- **Type**: Multimodal, four equal global minima
- **Characteristics**: 4^floor(n/2) equally good optima, none of them favoured
- **Best variant**: DESMA
- **Expected performance** (500 iter): 1e-6 to 1e-2

**Formula**: Sum of (a^2 + b - 11)^2 + (a + b^2 - 7)^2 over disjoint coordinate
pairs, plus the square of the unpaired coordinate when n is odd

**Use for**: Testing whether a swarm splits across equally good basins or
collapses into one

## Quick Reference Table

| Function           | Type               | Dimensionality | Best Variant | Difficulty         |
| ------------------ | ------------------ | -------------- | ------------ | ------------------ |
| Sphere             | Unimodal           | Any            | MA           | ⭐ Easy            |
| Zakharov           | Unimodal           | Any            | MA           | ⭐ Easy            |
| Rosenbrock         | Unimodal Valley    | Any            | MPMA         | ⭐⭐ Medium        |
| DixonPrice         | Unimodal Valley    | Any            | MPMA         | ⭐⭐ Medium        |
| BentCigar          | Ill-conditioned    | Any            | MPMA         | ⭐⭐⭐ Hard        |
| Discus             | Ill-conditioned    | Any            | MPMA         | ⭐⭐⭐ Hard        |
| Griewank           | Multimodal         | Any            | DESMA        | ⭐⭐ Medium        |
| Ackley             | Multimodal         | Any            | OLCE         | ⭐⭐⭐ Hard        |
| Rastrigin          | Highly Multimodal  | Any            | OLCE         | ⭐⭐⭐⭐ Very Hard |
| Levy               | Multimodal         | Any            | OLCE         | ⭐⭐⭐ Hard        |
| Schwefel           | Deceptive          | Any            | EOBBMA       | ⭐⭐⭐⭐⭐ Extreme |
| Michalewicz        | Steep Valleys      | Low-Medium     | EOBBMA       | ⭐⭐⭐⭐ Very Hard |
| Weierstrass        | Non-differentiable | Any            | EOBBMA       | ⭐⭐⭐⭐ Very Hard |
| HappyCat           | Plateau            | Any            | GSASMA       | ⭐⭐⭐ Hard        |
| ExpandedSchafferF6 | Composite          | Any            | AOBLMOA      | ⭐⭐⭐⭐ Very Hard |
| Himmelblau         | Multimodal         | Any            | DESMA        | ⭐⭐ Medium        |

## Usage Example

```go
package main

import (
    "fmt"
    "github.com/cwbudde/mayfly"
)

func main() {
    // Test on Rastrigin (highly multimodal)
    config := mayfly.NewOLCEConfig()  // Best for this function
    config.ObjectiveFunc = mayfly.Rastrigin
    config.ProblemSize = 30
    config.LowerBound = -5.12
    config.UpperBound = 5.12
    config.MaxIterations = 500

    result, err := mayfly.Optimize(config)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Rastrigin (D=30): Best Cost = %.4f\n", result.GlobalBest.Cost)
}
```

## Testing Strategy

### Progressive Testing

Test your implementation on functions in order of difficulty:

1. **Sphere** - Verify basic convergence
2. **Zakharov** - Verify unimodal performance
3. **Rosenbrock** - Test valley navigation
4. **Griewank** - Test multimodal handling
5. **Rastrigin** - Test local optima escape
6. **Schwefel** - Test deceptive landscape handling

### Dimensionality Testing

Start with low dimensions and scale up:

- **D=2**: Visualize the landscape
- **D=10**: Standard testing dimension
- **D=30**: Higher complexity
- **D=50-100**: Scalability testing

### Performance Baselines

Expected results for D=30, 500 iterations:

| Function   | MA   | DESMA | OLCE | EOBBMA | GSASMA | MPMA |
| ---------- | ---- | ----- | ---- | ------ | ------ | ---- |
| Sphere     | 1e-6 | 1e-8  | 1e-7 | 1e-6   | 1e-7   | 1e-6 |
| Rastrigin  | 55   | 40    | 30   | 38     | 36     | 48   |
| Rosenbrock | 25   | 15    | 12   | 18     | 20     | 8    |
| Schwefel   | 850  | 650   | 600  | 350    | 550    | 700  |

## Related Documentation

- [Algorithm Variants](algorithms/) - Individual algorithm documentation
- [Comparison Framework](api/comparison-framework.md) - Statistical testing
- [Getting Started](getting-started.md) - Usage tutorial
