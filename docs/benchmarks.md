# Benchmark Functions Reference

The library includes 15+ standard benchmark functions for testing and comparing algorithms.

## Function Categories

### Classic Benchmark Functions (5)

Standard test functions from optimization literature:
- **Sphere** - Unimodal, convex
- **Rastrigin** - Highly multimodal
- **Rosenbrock** - Unimodal, narrow valley
- **Ackley** - Multimodal, flat outer region
- **Griewank** - Many local minima

### CEC-Style Benchmark Functions (10)

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

- **Global minimum**: f(x*) ≈ -9.66 (10D)
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

## Quick Reference Table

| Function | Type | Dimensionality | Best Variant | Difficulty |
|----------|------|----------------|--------------|------------|
| Sphere | Unimodal | Any | MA | ⭐ Easy |
| Zakharov | Unimodal | Any | MA | ⭐ Easy |
| Rosenbrock | Unimodal Valley | Any | MPMA | ⭐⭐ Medium |
| DixonPrice | Unimodal Valley | Any | MPMA | ⭐⭐ Medium |
| BentCigar | Ill-conditioned | Any | MPMA | ⭐⭐⭐ Hard |
| Discus | Ill-conditioned | Any | MPMA | ⭐⭐⭐ Hard |
| Griewank | Multimodal | Any | DESMA | ⭐⭐ Medium |
| Ackley | Multimodal | Any | OLCE | ⭐⭐⭐ Hard |
| Rastrigin | Highly Multimodal | Any | OLCE | ⭐⭐⭐⭐ Very Hard |
| Levy | Multimodal | Any | OLCE | ⭐⭐⭐ Hard |
| Schwefel | Deceptive | Any | EOBBMA | ⭐⭐⭐⭐⭐ Extreme |
| Michalewicz | Steep Valleys | Low-Medium | EOBBMA | ⭐⭐⭐⭐ Very Hard |
| Weierstrass | Non-differentiable | Any | EOBBMA | ⭐⭐⭐⭐ Very Hard |
| HappyCat | Plateau | Any | GSASMA | ⭐⭐⭐ Hard |
| ExpandedSchafferF6 | Composite | Any | AOBLMOA | ⭐⭐⭐⭐ Very Hard |

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

| Function | MA | DESMA | OLCE | EOBBMA | GSASMA | MPMA |
|----------|-------|-------|------|--------|--------|------|
| Sphere | 1e-6 | 1e-8 | 1e-7 | 1e-6 | 1e-7 | 1e-6 |
| Rastrigin | 55 | 40 | 30 | 38 | 36 | 48 |
| Rosenbrock | 25 | 15 | 12 | 18 | 20 | 8 |
| Schwefel | 850 | 650 | 600 | 350 | 550 | 700 |

## Related Documentation

- [Algorithm Variants](algorithms/) - Individual algorithm documentation
- [Comparison Framework](api/comparison-framework.md) - Statistical testing
- [Getting Started](getting-started.md) - Usage tutorial
