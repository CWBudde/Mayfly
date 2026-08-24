# Migrating to v0.7

v0.7 is a pre-1.0 correctness release. Recorded trajectories and benchmark
comparisons from older releases must be treated as results from different
implementations.

## Reproducibility

Prefer an explicit seed:

```go
seed := int64(42)
config.Seed = &seed
```

`Config.Seed` and `Config.Rand` are mutually exclusive. `Result.Seed` is now
`*int64`; it is nil when the caller supplies an opaque `*rand.Rand`.
`Optimize` no longer writes RNG or derived velocity values into its input.

## Configuration and execution

- Enable at most one named `Use...` variant flag.
- `NM` is the number of mutants generated for each sex, not a combined count.
- Numeric configuration values must be finite. Velocity limits must both be
  zero for automatic bounds or form an explicit ordered pair.
- Objective functions must return finite values. Invalid proposals are rejected;
  an entirely invalid initial population returns `ErrNoFiniteObjectiveValue`.
- Parallel mode is seeded-equivalent to sequential mode. Recurrent stages may
  intentionally execute sequentially.

## Variant names and behavior

- GSASMA now applies simulated annealing to late-stage velocities and the fixed
  Eq. (10) golden-sine position update to both populations. `GoldenFactor` is
  deprecated and ignored.
- HMMA now follows the paper's every-iteration scheduled OBL/Cauchy mutation of
  the global optimum and Eq. (12) artificial gender mutation. Its new fields
  are `HMMAInformationExchange`, `HMMAScheduleOffset`, and
  `HMMAArtificialMutation`; `CauchyMutationRate` and `ApplyOBLToGlobalBest` are
  deprecated and ignored by HMMA.
- AOBLMOA means Aquila Optimizer and Opposition-Based Learning Mayfly
  Optimization Algorithm. It is scalar, not multi-objective.
- `PresetMultiObjective` is deprecated and returns an error. Pareto utilities
  remain available independently, but no optimizer currently returns a front.
- Standard MA, MPMA, AOBLMOA, EOBBMA, DESMA, and OLCE trajectories change due
  to equation, lifecycle, ranking, or boundary corrections.

## API changes

- `ComparisonRunner.TargetCost` is `*float64`; use `WithTarget` to enable it.
- Use `ClassifyProblemContext` for validated, cancelable, budgeted
  classification. The compatibility `ClassifyProblem` wrapper cannot report
  probe failures.
- Pareto archive additions return acceptance and errors; archive reads return
  defensive copies.
- JSON configuration loading is strict, so misspelled or duplicate keys now
  fail instead of silently selecting defaults.
