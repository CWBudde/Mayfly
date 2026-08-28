# Configuration Guide

Complete reference for every field in `mayfly.Config` and its nested
configuration types. Defaults in this guide come from the matching
`New...Config` factory. Fields described as automatic start at zero and are
resolved when optimization begins.

## Complete field index

Every top-level `Config` field is covered in the sections below:

| Group                 | Fields                                                                                                                            |
| --------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| Problem               | `ObjectiveFunc`, `ProblemSize`, `LowerBound`, `UpperBound`                                                                        |
| Population and budget | `NPop`, `NPopF`, `MaxIterations`                                                                                                  |
| Movement              | `G`, `GDamp`, `A1`, `A2`, `A3`, `Beta`, `Dance`, `FL`, `DanceDamp`, `FLDamp`, `VelMax`, `VelMin`                                  |
| Genetic operators     | `NC`, `NM`, `Mu`                                                                                                                  |
| DESMA                 | `UseDESMA`, `EliteCount`, `SearchRange`, `EnlargeFactor`, `ReductionFactor`                                                       |
| OLCE-MA               | `UseOLCE`, `OrthogonalFactor`, `ChaosFactor`                                                                                      |
| EOBBMA                | `UseEOBBMA`, `LevyAlpha`, `LevyBeta`, `OppositionRate`, `EliteOppositionCount`                                                    |
| GSASMA                | `UseGSASMA`, `InitialTemperature`, `CoolingRate`, `CoolingSchedule` |
| HMMA                  | `UseHMMA`, `HMMAInformationExchange`, `HMMAScheduleOffset`, `HMMAArtificialMutation` |
| MPMA                  | `UseMPMA`, `MedianWeight`, `GravityType`, `UseWeightedMedian`                                                                     |
| AOBLMOA               | `UseAOBLMOA`, `AquilaWeight`, `OppositionProbability`, `ArchiveSize`, `StrategySwitch`                                            |
| Run behavior          | `Convergence`, `Constraints`, `EnableParallel`, `MaxWorkers`, `Rand`                                                              |

`ObjectiveFunc`, `Rand`, and the function slices inside `Constraints` are not
JSON-serializable. Assign them after `LoadConfigFromFile`.

For common entry points and return types, see the
[API quick reference](quick-reference.md).

## Problem Parameters

These parameters define the optimization problem:

| Parameter       | Type                      | Required | Description                               |
| --------------- | ------------------------- | -------- | ----------------------------------------- |
| `ObjectiveFunc` | `func([]float64) float64` | **Yes**  | The function to minimize                  |
| `ProblemSize`   | `int`                     | **Yes**  | Number of decision variables (dimensions) |
| `LowerBound`    | `float64`                 | **Yes**  | Lower bound for all decision variables    |
| `UpperBound`    | `float64`                 | **Yes**  | Upper bound for all decision variables    |

Bounds are shared by all dimensions, must be finite, and must satisfy
`LowerBound < UpperBound`. The objective is minimized and must be safe for
concurrent calls when `EnableParallel` is true.

### Example

```go
config := mayfly.NewDefaultConfig()
config.ObjectiveFunc = mayfly.Sphere
config.ProblemSize = 30
config.LowerBound = -10
config.UpperBound = 10
```

## Population Parameters

Control the size and behavior of the mayfly populations:

| Parameter       | Type  | Default | Description                  |
| --------------- | ----- | ------- | ---------------------------- |
| `NPop`          | `int` | 20      | Population size for males    |
| `NPopF`         | `int` | 20      | Population size for females  |
| `MaxIterations` | `int` | 2000    | Maximum number of iterations |

**Recommendations**:

- Increase population for complex/high-dimensional problems (30-50)
- Decrease for simple problems or quick testing (10-15)
- `NPop` and `NPopF` are typically equal
- `NPopF` must not exceed `NPop`: every female is paired with the male at the
  same index

## Velocity Parameters

Control movement behavior of mayflies:

| Parameter   | Type      | Default | Description                             |
| ----------- | --------- | ------- | --------------------------------------- |
| `G`         | `float64` | 0.8     | Inertia weight                          |
| `GDamp`     | `float64` | 1.0     | Inertia weight damping ratio            |
| `A1`        | `float64` | 1.0     | Personal learning coefficient           |
| `A2`        | `float64` | 1.5     | Global learning coefficient for males   |
| `A3`        | `float64` | 1.5     | Global learning coefficient for females |
| `Beta`      | `float64` | 2.0     | Distance sight coefficient              |
| `Dance`     | `float64` | 5.0     | Nuptial dance coefficient               |
| `FL`        | `float64` | 1.0     | Random flight coefficient               |
| `DanceDamp` | `float64` | 0.8     | Dance damping ratio                     |
| `FLDamp`    | `float64` | 0.99    | Flight damping ratio                    |
| `VelMax`    | `float64` | Auto\*  | Maximum velocity (auto: 10% of bounds)  |
| `VelMin`    | `float64` | Auto\*  | Minimum velocity (auto: -10% of bounds) |

\*Auto-calculated if left at 0

## Mating Parameters

Control genetic operators:

| Parameter        | Type                | Default  | Description                                   |
| ---------------- | ------------------- | -------- | --------------------------------------------- |
| `NC`             | `int`               | `NCAuto` | Crossover offspring count                     |
| `NCRatio`        | `float64`           | 1.0      | Offspring per population member when `NCAuto` |
| `NM`             | `int`               | Auto\*   | Number of mutants (auto: 5% of NPop)          |
| `Mu`             | `float64`           | 0.01     | Fraction of dimensions mutated, in `[0, 1]`   |
| `CrossoverGamma` | `float64`           | 0.4      | Blend-crossover expansion factor              |
| `Selection`      | `SelectionStrategy` | `"rank"` | Parent selection rule                         |
| `TournamentSize` | `int`               | 3        | Candidates per tournament draw                |

\*`NM == 0` resolves to `round(0.05 * NPop)` and therefore does not disable
mutation. Crossover creates pairs, so use an even `NC`; `NC/2` may not exceed
either population size. If the effective mutant count is positive, `NC` must
be at least 2 because mutants are sampled from crossover offspring.

`CrossoverGamma` applies to the generic MA blend operator. DESMA ignores it and
uses its paper-specific per-coordinate `L` in `[-1,1]`; HMMA likewise uses its
own `L` in `[0,1]` semantics.

### How `CrossoverGamma` is resolved

Crossover is the blend (BLX-style) operator of the reference implementation: a
per-dimension coefficient `L` is drawn from `U(-gamma, 1+gamma)` and the
offspring are `L*x1 + (1-L)*x2` and `L*x2 + (1-L)*x1`, clamped to the problem
bounds. With `gamma = 0.4` — the reference default, exported as
`DefaultCrossoverGamma` — an offspring may land up to 40 % of the parental
interval outside it on either side.

Unlike `NC`, the zero value is _not_ honored literally. `gamma = 0` confines
`L` to `[0, 1]`, which makes every offspring a convex combination of its
parents; the population's convex hull then shrinks monotonically and mating can
never restore lost spread. A partially-filled `Config` literal that never
mentions `CrossoverGamma` must not silently get that, so `0`, negative values,
`NaN` and `Inf` all resolve to `DefaultCrossoverGamma`. Only a positive finite
value is taken as written.

### How `Mu` is resolved

`Mu` is a _fraction of dimensions_, not a per-gene probability: `MutateGaussian`
mutates exactly `ceil(Mu * ProblemSize)` randomly chosen dimensions. At the
default `0.01` that is a single dimension for any problem of 100 variables or
fewer.

Unlike `CrossoverGamma` and `NCRatio`, an out-of-range `Mu` is rejected rather
than resolved to a default: `Optimize` returns an error for anything outside
`[0, 1]`, including `NaN` and `Inf`. There is no sensible count to fall back to,
and the count is a slice bound, so the previous behaviour was a panic partway
through the run. Both endpoints are valid — `0` mutates nothing and `1` mutates
every dimension.

The exported operators (`MutateGaussian`, `MutateCauchy`, `HybridMutate`) take
the rate directly and so can still be reached with an out-of-range one. They
saturate instead of panicking: at or below `0` — and at `NaN` — they mutate
nothing, and at or above `1` they mutate every dimension.

### How `NC` is resolved

A written `NC` always wins, including the `0` that disables crossover. Only the
sentinel `NCAuto` defers to `NCRatio`, which derives the count from `NPop`:

```
NC == NCAuto  ->  round(NCRatio * NPop), rounded down to an even number
                  and clamped so NC/2 never exceeds either population
NC >= 0       ->  exactly NC
```

An `NCRatio` that is not a positive finite number falls back to 1.0 rather than
deriving a count from it. Zero is included deliberately: it is the zero value of
the field, so a `Config` literal that sets `NC` to `NCAuto` without also setting
`NCRatio` would otherwise run with no crossover at all — the failure this change
exists to remove. **`NCRatio` of 0 does not disable crossover; `NC = 0` does.**

`NCAuto` is the default because through v0.4.0 `NC` was an absolute `20` that
no caller had reason to revisit. Raising `NPop` therefore bought a larger swarm
and not one additional crossover: at `NPop` 4096 the same ten pairs mated while
4086 members only followed the global best, which quietly turned the algorithm
into plain PSO at large populations. `NCRatio` of 1.0 restores `NC == NPop`,
the ratio the default configuration already expressed at its own `NPop` of 20.

To reproduce a run recorded before v0.5.0, write the count the old default
carried: `config.NC = 20`.

### Parent selection

| Strategy       | Behaviour                                                   |
| -------------- | ----------------------------------------------------------- |
| `"rank"`       | Pairs the k-th best male with the k-th best female          |
| `"tournament"` | Draws `TournamentSize` candidates uniformly, mates the best |

`"rank"` is the default and is what the algorithm has always done. It is not
the elitism trap it can look like: with `NC` scaling, it mates the fitter half
of the population at every size. The trap was `NC` standing still, not the
pairing.

`"tournament"` is offered for experiments that want lower selection pressure.
It is not the default because it measurably reduced solution quality on this
library's own regression suite — Griewank 10D success fell from 70%+ to 60%
(Standard MA) and 20% (DESMA) at `NPop` 20. Measure before adopting it.

## Variant-Specific Parameters

### DESMA Parameters

| Parameter         | Type      | Default | Description                            |
| ----------------- | --------- | ------- | -------------------------------------- |
| `UseDESMA`        | `bool`    | false   | Enable DESMA variant                   |
| `EliteCount`      | `int`     | 10      | Number of elite solutions to generate  |
| `SearchRange`     | `float64` | Auto\*  | Search range for elite generation      |
| `EnlargeFactor`   | `float64` | 1.05    | Factor to enlarge range when improving |
| `ReductionFactor` | `float64` | 0.95    | Factor to reduce range when stagnating |

\*Auto: 10% of (UpperBound - LowerBound). The DESMA paper does not report its
initial radius, so this is a library compatibility default rather than a
paper-derived value. A non-zero value is also an initial radius and continues
to be adapted each iteration.

### OLCE-MA Parameters

| Parameter          | Type      | Default | Description                                             |
| ------------------ | --------- | ------- | ------------------------------------------------------- |
| `UseOLCE`          | `bool`    | false   | Enable OLCE-MA variant                                  |
| `OrthogonalFactor` | `float64` | 0.3     | Orthogonal learning strength (0-1)                      |
| `ChaosFactor`      | `float64` | 1.0     | Multiplier on chaotic-offspring constriction (0 disables) |

### EOBBMA Parameters

| Parameter              | Type      | Default | Description                           |
| ---------------------- | --------- | ------- | ------------------------------------- |
| `UseEOBBMA`            | `bool`    | false   | Enable EOBBMA variant                 |
| `LevyAlpha`            | `float64` | 1.5     | Lévy stability parameter (0 < α ≤ 2)  |
| `LevyBeta`             | `float64` | 1.0     | Lévy scale parameter                  |
| `OppositionRate`       | `float64` | 0.3     | Opposition learning probability (0-1) |
| `EliteOppositionCount` | `int`     | 3       | Number of elites for opposition       |

### GSASMA Parameters

| Parameter              | Type      | Default       | Description                                      |
| ---------------------- | --------- | ------------- | ------------------------------------------------ |
| `UseGSASMA`            | `bool`    | false         | Enable GSASMA variant                            |
| `InitialTemperature`   | `float64` | 100.0         | Extension temperature held until annealing starts |
| `CoolingRate`          | `float64` | 0.95          | Library-extension temperature decay rate         |
| `CoolingSchedule`      | `string`  | "exponential" | Extension: exponential, linear, or logarithmic   |
| `GoldenFactor`         | `float64` | 1.0           | **Deprecated and ignored**; absent from Eq. (10) |

### HMMA Parameters

| Parameter                 | Type      | Default | Description                                  |
| ------------------------- | --------- | ------- | -------------------------------------------- |
| `UseHMMA`                 | `bool`    | false   | Enable Hybrid Mutation MA                    |
| `HMMAInformationExchange` | `float64` | 1.5     | Eq. (7) `a4`; extension default because the paper omits its value |
| `HMMAScheduleOffset`      | `float64` | 0.99    | Historical compatibility-schedule offset; not the paper's unresolved Eq. (10) |
| `HMMAArtificialMutation`  | `float64` | 0.1     | Eq. (12) gender-exchange coefficient `rho`   |
| `CauchyMutationRate`      | `float64` | 0.3     | **Deprecated and ignored by HMMA**           |
| `ApplyOBLToGlobalBest`    | `bool`    | false   | **Deprecated and ignored by HMMA**           |

### MPMA Parameters

| Parameter           | Type      | Default  | Description                                      |
| ------------------- | --------- | -------- | ------------------------------------------------ |
| `UseMPMA`           | `bool`    | false    | Enable MPMA variant                              |
| `MedianWeight`      | `float64` | 0.5      | Median position influence (0-1)                  |
| `GravityType`       | `string`  | "paper"  | Published schedule; alternatives are extensions       |
| `UseWeightedMedian` | `bool`    | false    | Use fitness-weighted median                      |

### AOBLMOA Parameters

| Parameter               | Type      | Default            | Description                                                                   |
| ----------------------- | --------- | ------------------ | ----------------------------------------------------------------------------- |
| `UseAOBLMOA`            | `bool`    | false              | Enable AOBLMOA variant                                                        |
| `StrategySwitch`        | `int`     | Auto\*             | First iteration of the Aquila exploitation phase                              |
| `AquilaWeight`          | `float64` | `AquilaWeightAuto` | **Deprecated.** Restores the pre-v0.6.0 random branch draw when set to [0, 1] |
| `OppositionProbability` | `float64` | 0.3                | Unused by AOBLMOA, which opposes every offspring                              |
| `ArchiveSize`           | `int`     | 100                | Max size of a `ParetoArchive` the caller builds                               |

\*Auto: `0` resolves to 2/3 of `MaxIterations` per run and is never written
back. A value at or beyond `MaxIterations` is legal and means "never exploit".

Note that `NM` is inert under AOBLMOA: stochastic opposition-based learning
takes the slot Gaussian mutation occupies in the plain algorithm.

## Advanced Parameters

`Config.Convergence` and `Config.Constraints` are nil by default. A nil pointer
disables that feature; a non-nil pointer uses the fields documented below.

### Constraint Handling

`Config.Constraints` enables constrained optimization. Inequality functions
return `g(x)` and are feasible at `g(x) <= 0`. Equality functions return
`h(x)` and contribute violation only when `abs(h(x))` exceeds
`EqualityTolerance`. Violations are summed across all constraints.

| Parameter           | Type                       | Default       | Description                                     |
| ------------------- | -------------------------- | ------------- | ----------------------------------------------- |
| `Inequalities`      | `[]ConstraintFunction`     | nil           | Functions defining `g(x) <= 0`                  |
| `Equalities`        | `[]ConstraintFunction`     | nil           | Functions defining `h(x) = 0`                   |
| `EqualityTolerance` | `float64`                  | 0             | Allowed absolute equality residual              |
| `Handling`          | `ConstraintHandlingMethod` | `feasibility` | Candidate ranking strategy                      |
| `PenaltyMethod`     | `PenaltyMethod`            | `quadratic`   | `linear` or `quadratic` when using penalty mode |
| `PenaltyFactor`     | `float64`                  | required      | Positive multiplier when using penalty mode     |

Feasibility ranking applies these rules in order: feasible candidates outrank
infeasible candidates; feasible candidates are ordered by raw objective cost;
infeasible candidates are ordered by total violation and then raw cost.

```go
config.Constraints = &mayfly.ConstraintConfig{
    Inequalities: []mayfly.ConstraintFunction{
        func(x []float64) float64 { return x[0]*x[0] + x[1]*x[1] - 1 },
    },
    Equalities: []mayfly.ConstraintFunction{
        func(x []float64) float64 { return x[0] - x[1] },
    },
    EqualityTolerance: 1e-6,
}
```

To rank by a penalty score instead, set `Handling` to
`ConstraintHandlingPenalty`, select `PenaltyLinear` or `PenaltyQuadratic`, and
provide a positive `PenaltyFactor`. `Result.GlobalBest.Cost` remains the raw
objective value in both modes; `ConstraintViolation` reports feasibility.

Constraint function slices are omitted from JSON configuration files, like
`ObjectiveFunc`, and must be assigned after loading. Serializable constraint
settings are preserved. When parallel evaluation is enabled, objective and
constraint functions must be safe for concurrent calls and treat positions as
read-only.

Target-cost convergence requires a feasible incumbent. While the incumbent is
infeasible, stagnation measures reductions in constraint violation. A move
from infeasible to feasible always resets stagnation. Because feasibility can
take priority over raw cost, `ConvergenceCurve` can rise during a constrained
run even though the incumbent ranking improves.

### Convergence Detection

`Config.Convergence` is optional. When it is `nil` (the default), every run
executes exactly `MaxIterations`. When configured, `MaxIterations` remains a
hard upper bound while convergence criteria adapt the actual run length.

| Parameter              | Type       | Default   | Description                                                  |
| ---------------------- | ---------- | --------- | ------------------------------------------------------------ |
| `TargetCost`           | `*float64` | nil       | Stop when the best cost is at most this value                |
| `StagnationIterations` | `int`      | 0         | Stop after this many iterations without sufficient progress  |
| `MinImprovement`       | `float64`  | 0         | Absolute cost reduction required to reset stagnation         |
| `MinIterations`        | `int`      | 0 (one\*) | Lower bound on completed iterations before convergence stops |

\*Convergence checks happen after completed iterations, so zero permits the
first check after iteration one. `MinIterations` cannot exceed
`MaxIterations`.

```go
targetCost := 1e-8
config.Convergence = &mayfly.ConvergenceConfig{
    TargetCost:           &targetCost,
    StagnationIterations: 100,
    MinImprovement:       1e-10,
    MinIterations:        200,
}
```

Target and stagnation checks can be used independently or together. A target
pointer is used so zero and negative target costs remain valid. Stagnation is
measured against the last improvement larger than `MinImprovement`; smaller
improvements accumulate and can eventually reset the counter once their total
passes that threshold.

`Result.IterationCount` and the length of `Result.ConvergenceCurve` report the
actual number of completed iterations. `Result.TerminationReason` is one of
`TerminationMaxIterations`, `TerminationTargetCost`, or
`TerminationStagnation`. Progress observers receive exactly one update for
each completed iteration, including the final iteration of an early-stopped
run. The policy applies uniformly to the standard algorithm and all variants.

### Parallel Fitness Evaluation

| Parameter        | Type   | Default            | Description                                     |
| ---------------- | ------ | ------------------ | ----------------------------------------------- |
| `EnableParallel` | `bool` | false              | Evaluate objective-function batches in parallel |
| `MaxWorkers`     | `int`  | `runtime.NumCPU()` | Maximum concurrent objective-function calls     |

Parallel evaluation is opt-in so existing configurations retain their current
sequential behavior. A `MaxWorkers` value of zero also resolves to
`runtime.NumCPU()`; negative values are invalid. The optimizer caps its worker
pool at the largest configured core, offspring, or variant-specific batch, and
active objective calls never exceed either the current batch size or
`MaxWorkers`.

Custom objective functions must be safe for concurrent calls when parallel
evaluation is enabled and must treat the supplied position as read-only. Avoid
shared mutable state, or protect it with your own synchronization. Random
movement, candidate coefficients, selection decisions, and result commits stay
on the optimizer goroutine, so pure, deterministic objective functions produce
scheduling-independent batch results.

```go
config := mayfly.NewDefaultConfig()
config.EnableParallel = true
config.MaxWorkers = 4
```

Objective calls for initialization, male/female population updates, crossover
offspring, and mutation offspring are parallelized. Variant-specific batches
are covered too: DESMA elites, OLCE male-movement candidates, EOBBMA
opposition points, GSASMA male/female movement batches, and AOBLMOA Aquila and
opposition candidates. OLCE evaluates one chaotic candidate for the fittest
crossover offspring before mutation and survivor selection.

DESMA pre-draws random offsets and constructs independent elite positions with
bounded workers before evaluating them. MPMA computes independent position
dimensions concurrently, including weighted medians. Both use at most
`MaxWorkers`; they do not overlap objective-evaluation batches. All population,
personal-best, global-best, annealing, and archive updates remain serialized.

#### When parallel evaluation helps

Parallel evaluation is most useful when each objective call performs enough
CPU work to outweigh goroutine scheduling and synchronization, and when the
population or candidate batch contains several independent evaluations. Leave
it disabled for very cheap objectives, small populations, or workloads already
parallelized internally; in those cases the coordination overhead can make a
parallel run slower.

For CPU-bound objectives, start with `MaxWorkers` set to the smaller of
`runtime.NumCPU()` and the largest typical evaluation batch. More workers than
the current batch cannot improve throughput. Reduce the limit if each call uses
substantial memory or other constrained resources. Measure the complete
optimization workload because the profitable worker count depends on objective
cost, population sizes, and the selected variant.

Comparison-run parallelism and fitness-evaluation parallelism are separate
layers. Usually enable one layer at a time. If both are enabled, keep their
combined concurrency within the machine's capacity to avoid oversubscription.

The repository includes a benchmark matrix for cheap and CPU-expensive
objectives across the available worker counts:

```sh
go test -run '^$' -bench '^BenchmarkParallelFitnessEvaluation$' -benchmem -count=3 ./...
```

Compare `ns/op` within each objective group. Multiple runs help distinguish a
consistent speedup from ordinary timing noise; benchmark the real objective
before choosing a production worker limit.

### Random Number Generation

| Parameter | Type         | Default | Description                                                   |
| --------- | ------------ | ------- | ------------------------------------------------------------- |
| `Rand`    | `*rand.Rand` | nil     | Custom run-owned random generator; nil uses a time-based seed |

**Example for reproducible results**:

```go
import "math/rand"

config := mayfly.NewDefaultConfig()
config.Rand = rand.New(rand.NewSource(42)) // Fixed seed
```

Do not share a `*rand.Rand` between concurrent optimization runs. The standard
generator is not safe for concurrent use. `Result.Seed` records the generated
seed for a nil `Rand`; Go's `rand.Rand` does not expose a caller-provided seed.

## Factory Functions

Pre-configured factory functions for each variant:

```go
// Standard MA
config := mayfly.NewDefaultConfig()

// DESMA
config := mayfly.NewDESMAConfig()

// OLCE-MA
config := mayfly.NewOLCEConfig()

// EOBBMA
config := mayfly.NewEOBBMAConfig()

// GSASMA
config := mayfly.NewGSASMAConfig()

// MPMA
config := mayfly.NewMPMAConfig()

// AOBLMOA
config := mayfly.NewAOBLMOAConfig()
```

All factory functions set sensible defaults. You only need to set the required problem parameters.

## Configuration Validation

The `Optimize()` function validates configuration:

**Required fields**:

- `ObjectiveFunc`
- `ProblemSize`
- finite `LowerBound` and `UpperBound` with `LowerBound < UpperBound`
- positive `MaxIterations`, `NPop`, and `NPopF` (set by all factories)
- `NPopF` no larger than `NPop`
- at most one of `UseAOBLMOA`, `UseEOBBMA`, and `UseMPMA`: they replace the same
  position-update phase, so combining them left the losing one inert

**Auto-calculated fields** (if zero):

- `VelMax` = 0.1 \* (UpperBound - LowerBound)
- `VelMin` = -VelMax
- `NM` = max(1, int(0.05 \* NPop))
- `SearchRange` (DESMA) = 0.1 \* (UpperBound - LowerBound)
- `MaxWorkers` = runtime.NumCPU()

**Validation errors**:

```go
result, err := mayfly.Optimize(config)
if err != nil {
    // Handle errors:
    // - "objective function is required"
    // - "problem size must be positive"
    // - "invalid bounds: lower bound must be less than upper bound"
    // - etc.
}
```

## Configuration Tips

### For Quick Testing

```go
config := mayfly.NewDefaultConfig()
config.MaxIterations = 100  // Reduce for speed
config.NPop = 10
config.NPopF = 10
```

### For High-Dimensional Problems

```go
config := mayfly.NewDefaultConfig()
config.NPop = 50  // Increase population
config.NPopF = 50
config.MaxIterations = 2000
```

### For Expensive Function Evaluations

```go
config := mayfly.NewDESMAConfig()
config.EnableParallel = true
config.MaxWorkers = 4
config.EliteCount = 3       // Reduce elite count
config.NPop = 15            // Smaller population
config.MaxIterations = 500  // Fewer iterations
```

### For Maximization Problems

```go
// Negate the objective function
func maximize(x []float64) float64 {
    profit := calculateProfit(x)
    return -profit  // Negate for maximization
}

config.ObjectiveFunc = maximize
```

## Related Documentation

- [Parameter-tuning Tutorial](../parameter-tuning.md) - Empirical tuning workflow
- [Unified Framework](unified-framework.md) - Builder API and algorithm selection
- [Algorithm Comparison](comparison-framework.md) - Statistical comparison tools
- [Getting Started](../getting-started.md) - Tutorial and examples
