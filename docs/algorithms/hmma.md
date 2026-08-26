# HMMA - Hybrid Mutation Mayfly Algorithm

## Research reference

H. Zhang et al., “Improved mayfly algorithm based on hybrid mutation,”
*Electronics Letters* 58 (2022).
[DOI 10.1049/ell2.12568](https://doi.org/10.1049/ell2.12568)

## Lifecycle and equations

HMMA retains ordinary male/female movement. Immediately after positions are
updated, it generates one mutation of the current global optimum.

The current library uses this historical compatibility schedule:

```text
Ps = -exp(-t / Iter_MAX) + theta
```

It is not Equation (10) as printed by the publisher, which is:

```text
Ps = -exp((1 - t / Iter_MAX)^20) + theta
```

The paper's parameter tuple reports `theta = 0.005`; together with that printed
equation, `Ps` is negative throughout the stated iteration interval. The tuple
also prints `ub = 0.1` and `lb = 10`, and omits the Equation (7) coefficient
`a4`. Mayfly therefore preserves its established schedule and defaults as an
explicit extension until author clarification is available; it does not claim
that they reproduce Table 1.

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

| Field | Symbol | Default | Valid range | Status |
| --- | --- | ---: | --- | --- |
| `HMMAInformationExchange` | `a4` | 1.5 | finite, `> 0` | extension default; paper omits `a4` |
| `HMMAScheduleOffset` | `theta` | 0.99 | finite, `[0,1]` | compatibility schedule; paper reports `0.005` with an unusable printed equation |
| `HMMAArtificialMutation` | `rho` | 0.1 | finite, `[0,1]` | paper value |

## Published experiment reference

The source-audited [Table 1 reference](../reference-data/hmma-2022-table1.json)
records the 50 replications, 1,000 iterations, parameter tuple, and all five
published statistics for MA, IMA, AMMA, OCMA, and HMMA on F3, F7, and F15.
It is labeled `reproduction_claim: false`: the paper does not report population
size, F3/F7 dimensions, seeds, raw runs, consistent bounds, or a usable
Equation (10) probability. Its data-availability statement offers the missing
material from the corresponding author on reasonable request.

`CauchyMutationRate` and `ApplyOBLToGlobalBest` are deprecated compatibility
fields and are ignored by HMMA. The paper uses the iteration-dependent `Ps`
choice every iteration; it does not describe adaptive Cauchy/Gaussian offspring
mutation or a periodic opposition toggle.
