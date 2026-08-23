# HMMA - Hybrid Mutation Mayfly Algorithm

## Research reference

_An improved mayfly optimization algorithm based on hybrid mutation_ (2022),
Electronics Letters. DOI:
[10.1049/ell2.12568](https://doi.org/10.1049/ell2.12568).

## Implementation

HMMA retains the Mayfly movement and mating stages, replaces ordinary mutation
with an adaptive Cauchy/Gaussian mutation, and periodically tests the opposition
of the global best. These stages were incorrectly documented and enabled under
GSASMA through v0.6.

```go
config := mayfly.NewHMMAConfig()
config.ObjectiveFunc = mayfly.Rastrigin
config.ProblemSize = 30
config.LowerBound = -5.12
config.UpperBound = 5.12

result, err := mayfly.Optimize(config)
```

`CauchyMutationRate` controls the base exploratory mutation probability, and
`ApplyOBLToGlobalBest` controls the periodic opposition stage. HMMA is mutually
exclusive with every other named variant.

