# HMMA - Hybrid Mutation Mayfly Algorithm

## Research reference

M. Zhang et al., “Improved mayfly algorithm based on hybrid mutation,”
*Electronics Letters* 58 (2022).
[DOI 10.1049/ell2.12568](https://doi.org/10.1049/ell2.12568)

## Lifecycle and equations

HMMA retains ordinary male/female movement. Immediately after positions are
updated, it generates one mutation of the current global optimum.

The scheduled probability is Eq. (10):

```text
Ps = -exp(-t / Iter_MAX) + theta
```

When `Ps > rand`, Eqs. (6)-(7) generate

```text
x_gbest' = ub + r3*(lb - x_gbest)
x_new = a4*(x_gbest - x_gbest')
```

where `r3` is drawn independently per coordinate. Eq. (7) has no additional
leading `x_gbest` term. Otherwise Eq. (8) generates

```text
x_new = Cauchy(0,1) * x_gbest
```

The candidate is repaired to the configured bounds, evaluated once, and kept
only when it improves the global optimum (Eq. (11)).

HMMA then performs ordinary convex male/female crossover and applies the
artificial mutation operator in Eq. (12) to every sibling pair:

```text
male_new   = (1-rho)*male_child   + rho*female_child
female_new = (1-rho)*female_child + rho*male_child
```

Both expressions use the original pair. These converted gender populations
enter survivor selection. Ordinary Gaussian mutants are not additionally
created for HMMA, so `NM` is inert for this variant.

## Configuration

```go
config := mayfly.NewHMMAConfig()
config.ObjectiveFunc = mayfly.Rastrigin
config.ProblemSize = 30
config.LowerBound = -5.12
config.UpperBound = 5.12

result, err := mayfly.Optimize(config)
```

| Field | Paper symbol | Default | Valid range |
| --- | --- | ---: | --- |
| `HMMAInformationExchange` | `a4` | 1.5 | finite, `> 0` |
| `HMMAScheduleOffset` | `theta` | 0.99 | finite, `[0,1]` |
| `HMMAArtificialMutation` | `rho` | 0.1 | finite, `[0,1]` |

The defaults are the values stated in the paper's simulation setup.

`CauchyMutationRate` and `ApplyOBLToGlobalBest` are deprecated compatibility
fields and are ignored by HMMA. The paper uses the iteration-dependent `Ps`
choice every iteration; it does not describe adaptive Cauchy/Gaussian offspring
mutation or a periodic opposition toggle.
