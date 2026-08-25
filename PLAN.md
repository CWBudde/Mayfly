# Mayfly Algorithm Suite Roadmap

This file is the source of truth for remaining work. Completed phases are intentionally
condensed; unchecked items retain the details needed to act on them.

## Completed phases

### Phase 1: Core and advanced infrastructure — complete

- Added bounded parallel evaluation for both populations, genetic operators, and all
  variant-specific stages. `Config.MaxWorkers` defaults to `runtime.NumCPU()`, while
  `Config.EnableParallel` preserves the sequential compatibility path.
- Added deterministic sequential/parallel proposal, evaluation, and commit behavior;
  race, accounting, parity, and cheap/expensive-objective tests; performance benchmarks;
  and guidance on when parallel execution offsets its overhead.
- Added concurrent multi-algorithm comparison, statistical aggregation, and result
  visualization.
- Added convergence/stagnation detection, adaptive iteration limits, penalty methods,
  feasibility rules, and constraint utilities.

### Phase 2: Release preparation — complete

- Completed the static-analysis cleanup, coverage work (80%+), unit/race/integration
  testing, reproducible profiling, hot-path optimization, and regression benchmarks.
- Added semantic-versioning and release documentation, `CHANGELOG.md`, package metadata,
  the release workflow, and pkg.go.dev publication checks.
- The final correctness-audit run of `just lint-fix` completed its formatter stage, but
  the installed golangci-lint v2.11.4 analyzer reported `no go files to analyze`. No tool
  was installed or upgraded. Formatting, module tidiness, unit tests, `go vet`, race tests,
  integration tests, and nested example-module checks passed independently.

### Phase 3: Correctness and algorithm-fidelity audit — complete (2026-08-24)

- Standardized all 15 single-objective benchmark functions on `f([]) == 0`, fixed the
  `Levy([])` panic, added one table-driven regression test, and mirrored the change in
  Dragonfly. Do not "correct" two intentional formulas: Levy's middle term contains
  `sin(math.Pi*wi + 1)`, and Expanded Schaffer F6 includes the wrap-around pair
  `g(x[n-1], x[0])`.
- Replaced the selector's scale-dependent mean-gradient heuristic with seeded,
  scale-free line scans. Retuned `smoothRoughness` from 3.0 to 2.2 and
  `multimodalTurningPoints` from 6.0 to 5.0 after forty-seed testing, made classification
  reproducible, and reconciled sampled results with benchmark recommendations where the
  sampling resolution supports doing so.
- Corrected core lifecycle behavior: both populations participate in initialization,
  best tracking, stopping, and reporting; populations are sorted before rank-dependent
  updates; offspring retain sex; configuration is immutable; metadata is truthful; and
  finite values, bounds, velocity, variants, and cancellation share validation paths.
- Restored the published scalar-distance and genetic stages of MA; corrected MPMA's
  ranked median and nonlinear gravity; corrected AOBLMOA's Aquila strategies, crossover,
  stochastic opposition, and scalar-objective naming; restored EOBBMA's
  fitness-dependent updates and mirror boundaries; and fixed DESMA's dynamic elite
  schedule, replacement order, defaults, and no-op behavior.
- Moved OLCE orthogonal learning and chaotic exploitation to their published lifecycle
  stages, made orthogonal arrays dimension-safe beyond three dimensions, and corrected
  evaluation accounting. Separated GSASMA and HMMA into faithful implementations:
  GSASMA uses annealed velocity selection and golden-sine position updates for both
  populations; HMMA uses its scheduled OBL/Cauchy global-best cascade and artificial
  gender mutation.
- Removed misleading multi-objective presets while retaining validated Pareto utilities;
  unified direct, builder, JSON, and template validation; corrected comparison statistics
  and atomic exports; and added checked error-returning alternatives for panic-prone
  helpers. Deprecated compatibility APIs remain defensive and source-compatible.
- Added equation fixtures plus seeded parity, non-finite, aliasing, cancellation,
  configuration-reuse, Pareto, and statistical regressions. Updated v0.7.0 migration
  notes and invalidated deterministic benchmark baselines produced by older algorithms.

### Phase 4: Benchmark-suite expansion — complete (2026-08-24)

- Added the complete usable CEC2017 and CEC2020 bound-constrained suites through
  validated loaders for the organizers' official transformation data, plus four
  constrained engineering-design benchmarks with mixed-variable projection and
  normalized Mayfly adapters.

## Remaining phases

### Phase 5: Documentation and examples — medium priority

The API examples, quick-reference guide, and parameter documentation are complete.

#### Tutorials

- [ ] Write a getting-started tutorial.
- [ ] Write an algorithm-selection guide.
- [ ] Write a parameter-tuning tutorial.
- [ ] Write a custom-objective-function guide.

#### Real-world examples

- [ ] Add a neural-network hyperparameter-tuning example.
- [ ] Add a resource-allocation example.
- [ ] Add a scheduling example.
- [ ] Add a feature-selection example.

### Phase 6: Community setup — low priority

- [ ] Add `CONTRIBUTING.md`.
- [ ] Add issue templates.
- [ ] Add pull-request templates.

### Phase 7: Research reproducibility and classifier follow-ups — low priority

- [ ] Reproduce the original paper results for MA, DESMA, OLCE-MA, EOBBMA, GSASMA,
      HMMA, MPMA, and AOBLMOA. Historical seeded results from before the correctness
      audit must not be compared as if they came from the corrected implementations.
      The post-audit harness now provides a controlled baseline and can enforce
      exact per-run objective-evaluation budgets. The original MA paper's 50-run,
      95,000-evaluation target and 20/20 population have been source-audited.
      Its Appendix A tuning grid, Table 6 benchmark protocol, and published
      Basic MA/VGMA/SMA/IMA rows are encoded in
      `docs/reference-data/original-ma-2020-table6.json`. An exact preset and
      comparison remain blocked because the paper does not resolve how crossover
      rate 0.95 applies to its conflicting crossover descriptions or what the
      Gaussian mutation rate 0.1 controls.
- [x] Provide experiment scripts. `cmd/paper-reproduction` and
      `scripts/run-paper-experiments.sh` run paired seeded trials of all eight variants,
      export raw CSV/JSON, and record the protocol, revision, runtime, bounds, and
      configuration in a manifest.
- [ ] Correct OLCE's chaotic-exploitation stage against the complete published equations.
      The publisher's authoritative pseudocode figures became accessible on 2026-08-25
      and invalidate the earlier equal-fitness premise: they specify Chebyshev mutation
      over all `N` crossover offspring, so there is no fittest-offspring tie. The current
      Logistic-map, first-best stage is now documented as an extension. The figures do
      not expose the exact recurrence and component mutation equation, so implementation
      remains blocked on an authoritative full equation or reference implementation.
- [ ] Calibrate GSASMA's undocumented annealing recurrence/defaults and SMA crossover and
      mutation probability bounds against the authors' reference implementation or
      reproducible experimental data. The current cooling schedule is explicitly a
      library extension, and ordinary configured mating remains in place rather than
      inventing constants absent from the paper.
- [x] Re-test Dragonfly's selector thresholds across multiple seeds and, if it still uses
      the old values, align them with Mayfly's `smoothRoughness = 2.2` and
      `multimodalTurningPoints = 5.0`. Across 40 seeds, the old thresholds classified
      Schwefel correctly only 37/40 times for modality and 25/40 for landscape; the
      aligned thresholds and new multi-seed regression pass 40/40.
- [x] Decide whether to improve sampled classification for Griewank on `[-600,600]`.
      Mayfly currently measures it as `Unimodal`/`Smooth`, while
      `RecommendForBenchmark` deliberately remains `HighlyMultimodal`/`Rugged`: the
      order-one cosine ripples are only a few units wide against a value range around
      100,000, so the current line spacing aliases them and normalized total variation
      treats them as negligible. The hard-coded recommendation is more useful to an
      optimizer working near the optimum. A 200-seed sweep found that increasing from
      65 to 1025 points per line raises sampling cost from 390 to 6,150 evaluations and
      recovers modality, but still reports `Smooth` in every run. Keep the cheap generic
      box-scale classifier and the literature-backed benchmark override.
