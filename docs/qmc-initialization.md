# Quasi-random initial populations

`Config.QMCInit` seeds the first generation from a low-discrepancy sequence
instead of independent uniform draws. This page is the measurement behind it,
including where it makes no difference at all.

## What it does

The first generation is a sample of the search box. Forty uniform random points
in ten dimensions leave gaps and clusters, and nothing later in the algorithm is
told where the gaps are. A low-discrepancy sequence covers the same box more
evenly for the same number of function evaluations — that is the whole of the
idea, and it is why quasi-random initialization is a standard cheap addition to
population-based optimizers.

```go
config := mayfly.NewDefaultConfig()
config.ObjectiveFunc = mayfly.Rastrigin
config.ProblemSize = 30
config.LowerBound, config.UpperBound = -5.12, 5.12
config.QMCInit = mayfly.QMCInitSobol // or QMCInitHalton, or QMCInitUniform
```

or through the builder:

```go
result, err := mayfly.NewBuilder("MA").
    ForProblem(mayfly.Rastrigin, 30, -5.12, 5.12).
    WithQMCInitialPopulation(mayfly.QMCInitSobol).
    Optimize()
```

The sequences come from [github.com/cwbudde/qmc](https://github.com/CWBudde/qmc).
Males and females are drawn from **one** stream, not two: two generators with
the same configuration produce the same points, so seeding the halves separately
would place every female exactly on top of a male and halve the coverage.

### The seed

`Config.QMCSeed` pins the scramble. Left at zero it is drawn from the run's RNG,
which is what makes repeated runs differ.

This matters more than it looks. A quasi-random sequence is deterministic, so an
unrandomized initial population would be identical in every run, and a thirty-run
study would report a standard deviation that measured only what happens after
initialization. Scrambling (Owen for Sobol, nested for Halton) keeps the
low-discrepancy property while making each run a different point set — the
"randomized QMC" arrangement. Set `QMCSeed` to reproduce a single run without
also pinning `Config.Rand`.

### Why the population is one aligned block

Sobol's balance property covers a block of raw indices aligned on a power of two,
not any run of consecutive points. The initializer therefore skips to the next
power of two at or above `NPop+NPopF` — forty individuals sit on raw indices
64..103 — so the population is one balanced block rather than straddling two.
Halton takes the conventional burn-in of 64 instead, because its early points sit
in a corner of the box in every coordinate with a large base.

What that buys is visible directly: split each axis into eight equal parts and
count the parts that hold at least one of the forty individuals. Sobol occupies
all eight on every axis, at every seed and every dimension tested up to 64.
Halton leaves at most one part empty. Both are asserted in
`TestInitialPositionsStratification`. Uniform draws are not asserted on — over
the same seeds they leave up to two parts of some axis empty, which is the gap
this feature exists to close, and pinning it would only make the test brittle.

## The measurement

Reproduce with:

```bash
go test -tags qmcstudy -run TestQMCInitStudy -timeout 60m -v
```

Standard MA, 30 runs per problem, 500 iterations, `NPop = NPopF = 20`, over the
16 problems in the benchmark suite. Runs are paired by seed: run _r_ of each
strategy starts from `rand.NewSource(r)`. The pairing removes the seed as a
source of difference but does not make the runs step-for-step comparable — a
different starting population leads the algorithm to consume its generator
differently from the first iteration onwards.

`differ` counts the run pairs the Wilcoxon signed-rank test can actually see: it
discards pairs closer together than 1e-10, so a problem both strategies solve to
machine precision reduces to no evidence rather than to a verdict. `p` is against
the uniform baseline; `better` is filled in only at p < 0.05.

| problem        | init    |       mean |     median |     stddev | differ |     p | better    |
| -------------- | ------- | ---------: | ---------: | ---------: | -----: | ----: | --------- |
| Sphere_10D     | uniform | 5.4039e-54 | 1.2400e-56 | 2.5517e-53 |        |       |           |
|                | sobol   | 8.4844e-54 | 2.1517e-57 | 4.1845e-53 |    0/30 | 1.000 | all tied  |
|                | halton  | 8.2545e-55 | 6.3188e-57 | 3.3037e-54 |    0/30 | 1.000 | all tied  |
| Sphere_30D     | uniform | 1.4251e-11 | 5.2094e-12 | 2.0508e-11 |        |       |           |
|                | sobol   | 2.3248e-11 | 6.5013e-12 | 4.6923e-11 |    1/30 | 0.317 |           |
|                | halton  | 6.2788e-11 | 9.6731e-12 | 2.3293e-10 |    2/30 | 0.180 |           |
| Rastrigin_10D  | uniform | 1.8018e+00 | 1.4924e+00 | 1.6973e+00 |        |       |           |
|                | sobol   | 1.1540e+00 | 9.9496e-01 | 1.0181e+00 |   29/30 | 0.043 | **sobol** |
|                | halton  | 1.4334e+00 | 9.9496e-01 | 1.1212e+00 |   28/30 | 0.374 |           |
| Rastrigin_30D  | uniform | 1.8161e+01 | 1.8086e+01 | 6.8298e+00 |        |       |           |
|                | sobol   | 1.9317e+01 | 1.7417e+01 | 7.3451e+00 |   30/30 | 0.992 |           |
|                | halton  | 2.0220e+01 | 1.7748e+01 | 8.3592e+00 |   30/30 | 0.453 |           |
| Rosenbrock_10D | uniform | 6.1640e-01 | 5.8109e-04 | 1.6078e+00 |        |       |           |
|                | sobol   | 4.6833e-01 | 3.1576e-04 | 1.4325e+00 |   30/30 | 0.405 |           |
|                | halton  | 7.8931e-01 | 5.3569e-04 | 1.5035e+00 |   30/30 | 0.926 |           |
| Rosenbrock_30D | uniform | 5.8979e+01 | 7.3849e+01 | 3.4659e+01 |        |       |           |
|                | sobol   | 5.1131e+01 | 4.8677e+01 | 2.7877e+01 |   30/30 | 0.221 |           |
|                | halton  | 4.7713e+01 | 2.6562e+01 | 4.1769e+01 |   30/30 | 0.171 |           |
| Ackley_10D     | uniform | 2.8014e-01 | 3.9968e-15 | 5.7093e-01 |        |       |           |
|                | sobol   | 6.0116e-02 | 3.9968e-15 | 2.3257e-01 |   10/30 | 0.126 |           |
|                | halton  | 2.4740e-01 | 3.9968e-15 | 5.0152e-01 |   15/30 | 1.000 |           |
| Ackley_30D     | uniform | 2.4448e+00 | 2.6594e+00 | 8.5510e-01 |        |       |           |
|                | sobol   | 2.5184e+00 | 2.5813e+00 | 8.1525e-01 |   30/30 | 0.688 |           |
|                | halton  | 2.4095e+00 | 2.4957e+00 | 6.8859e-01 |   30/30 | 0.644 |           |
| Griewank_10D   | uniform | 1.4289e-01 | 1.1570e-01 | 1.2468e-01 |        |       |           |
|                | sobol   | 8.1382e-02 | 6.3910e-02 | 6.6895e-02 |   30/30 | 0.022 | **sobol** |
|                | halton  | 1.0769e-01 | 8.8592e-02 | 8.2289e-02 |   30/30 | 0.309 |           |
| Griewank_30D   | uniform | 3.5683e-02 | 2.9503e-02 | 3.0804e-02 |        |       |           |
|                | sobol   | 2.2819e-02 | 8.6304e-03 | 2.8712e-02 |   30/30 | 0.221 |           |
|                | halton  | 2.3604e-02 | 8.6267e-03 | 3.4539e-02 |   30/30 | 0.131 |           |
| Schwefel_10D   | uniform | 5.2123e+02 | 4.7375e+02 | 2.0401e+02 |        |       |           |
|                | sobol   | 5.4629e+02 | 4.7375e+02 | 1.5639e+02 |   26/30 | 0.576 |           |
|                | halton  | 5.7219e+02 | 5.8232e+02 | 1.8124e+02 |   28/30 | 0.350 |           |
| Levy_10D       | uniform | 6.0266e-24 | 1.4998e-32 | 3.2454e-23 |        |       |           |
|                | sobol   | 1.9393e-22 | 1.4998e-32 | 1.0444e-21 |    0/30 | 1.000 | all tied  |
|                | halton  | 2.8278e-32 | 1.4998e-32 | 7.1517e-32 |    0/30 | 1.000 | all tied  |
| Zakharov_10D   | uniform | 6.3555e-26 | 5.0468e-27 | 2.0875e-25 |        |       |           |
|                | sobol   | 5.5239e-25 | 2.4205e-27 | 2.1241e-24 |    0/30 | 1.000 | all tied  |
|                | halton  | 1.1584e-24 | 2.2004e-27 | 4.6888e-24 |    0/30 | 1.000 | all tied  |
| BentCigar_10D  | uniform | 2.7709e-01 | 9.5113e-48 | 1.4922e+00 |        |       |           |
|                | sobol   | 2.0437e-09 | 4.7919e-46 | 1.1006e-08 |    2/30 | 0.655 |           |
|                | halton  | 5.4290e-29 | 3.7326e-45 | 2.9210e-28 |    1/30 | 0.317 |           |
| Discus_10D     | uniform | 3.5067e-42 | 8.7267e-50 | 1.8882e-41 |        |       |           |
|                | sobol   | 5.9118e-42 | 3.9526e-51 | 3.1836e-41 |    0/30 | 1.000 | all tied  |
|                | halton  | 1.5075e-40 | 5.0827e-51 | 8.1009e-40 |    0/30 | 1.000 | all tied  |
| HappyCat_10D   | uniform | 1.5420e-01 | 1.2952e-01 | 6.5030e-02 |        |       |           |
|                | sobol   | 1.9001e-01 | 1.7026e-01 | 8.6637e-02 |   30/30 | 0.094 |           |
|                | halton  | 1.6542e-01 | 1.5606e-01 | 7.7019e-02 |   30/30 | 0.644 |           |

## What it says

**Two significant results out of sixteen problems, both for Sobol, none against.**
Rastrigin at 10 dimensions (mean 1.80 → 1.15, p = 0.043) and Griewank at 10
dimensions (0.143 → 0.081, p = 0.022). Halton never reaches p < 0.05 anywhere.

Two hits in thirty-two tests is close to what 0.05 produces by chance — the
expected number of false positives is about 1.6 — so the count on its own is not
evidence of anything, and neither result is far enough below 0.05 to survive any
correction for testing sixteen problems at once.

**An earlier run of this same study, against the pre-v0.6.0 algorithm, also
found exactly two significant Sobol results and none against — on two different
problems** (Rastrigin_30D at p < 0.001 and Ackley_10D at p = 0.007, both of
which now sit at p = 0.99 and p = 0.13). The count is stable and *which
problems* it lands on is not. That is what a chance-level effect looks like from
the inside, and it is the most useful thing this measurement has produced: it
rules out the reading that Sobol reliably helps on specific problems.

Counting means over the twelve problems that are not decided by ties, Sobol is
ahead on 7 and Halton on 7. Neither is significant as a sign test. **The
direction is mildly favorable and the magnitude is mostly nothing.**

**Four problems say nothing at all.** Sphere_10D, Levy_10D, Zakharov_10D and
Discus_10D are solved to machine precision — 1e-24 and below — by every strategy
in every run, so all thirty pairs tie and the test has no evidence to weigh.
BentCigar_10D is nearly the same story with 1 to 2 pairs differing. A table like
this invites reading the mean column and concluding that Sobol is 1e2 times
worse on Levy; it is not, both are zero, and the difference is which denormal
the last iteration happened to land on.

**What the two hits have in common is multimodality.** Rastrigin and Griewank
are the two problems in the suite with dense local minima that MA does not
escape, and both significant results in the earlier run were multimodal as well.
That is consistent with the mechanism — on a problem the algorithm solves
regardless, the starting sample is forgotten within a few dozen iterations; on
one where it gets stuck, where it started is still visible after 500 — but four
data points spread over two different algorithm versions is a hypothesis, not a
finding.

## Recommendation

**Uniform stays the default.** The measurement does not support a blanket
switch, and the default is the behavior every existing result in this repository
was produced under.

Reach for `QMCInitSobol` when the run does not converge — multimodal problems,
populations that are small relative to the dimension. That is where the
significant results have landed in both runs of this study and where the
mechanism predicts they should be, and it costs nothing worth measuring: one
generator construction per run against 500 iterations of the algorithm. Treat it
as worth trying on your own problem, not as an improvement you can assume.

`QMCInitHalton` is available and correct, but nothing here distinguishes it from
uniform. Sobol is the one with evidence behind it, and Sobol is limited to the
1024 dimensions of the direction numbers `qmc` embeds — past that it returns an
error naming the ceiling, and Halton has no ceiling.

## Caveats

- One algorithm (Standard MA) and one population size. The variants are
  untested here and a larger population would weaken the effect, since the gap
  a low-discrepancy sample closes is largest when points are scarce.
- 500 iterations. On a shorter budget the initial population is a larger share
  of the search and the effect should grow.
- The p-values use the normal approximation to the Wilcoxon statistic that
  `comparison.go` implements, with no correction for the sixteen problems tested
  in parallel. Taken as a family, neither p = 0.022 nor p = 0.043 among 32 tests
  is significant.
- The numbers above are tied to a version of the algorithm. They were remeasured
  when this work was rebased onto v0.6.0 and they moved; re-run the study rather
  than trusting the table after any change to the core update rules.
