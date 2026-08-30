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

- [x] Write a getting-started tutorial with a complete, reproducibly seeded
      optimizer, result interpretation, custom-objective guidance, normalized
      per-dimension bounds, troubleshooting, and a CI-tested runnable example.
- [x] Write an algorithm-selection guide covering selector scope, explicit
      problem metadata, all variant heuristics and overhead hints, checked APIs,
      automatic-classifier limits, budget-matched validation, common mistakes,
      and a CI-tested reproducibly seeded example.
- [x] Write a parameter-tuning tutorial covering baseline-first experimental
      design, parameter prioritization and coupling, checked heuristic
      auto-tuning limits, paired seeds, exact evaluation budgets, held-out
      validation, reproducible configuration handoff, common mistakes, and a
      CI-tested runnable population-size sweep.
- [x] Write a custom-objective-function guide covering the scalar minimization
      contract, normalized heterogeneous and projected variables, objective
      scaling and maximization, constraints, finite failure handling,
      concurrency and expensive evaluations, validation tests, common
      mistakes, and a CI-tested runnable nonlinear curve-fitting example.

#### Real-world examples

- [ ] Add a neural-network hyperparameter-tuning example.
- [ ] Add a resource-allocation example.
- [ ] Add a scheduling example.
- [ ] Add a feature-selection example.

### Phase 6: Community setup — low priority

- [ ] Add `CONTRIBUTING.md`.
- [ ] Add issue templates.
- [ ] Add pull-request templates.

### Phase 7: Research reproducibility — low priority

Completed audits, reference transcriptions, fidelity decisions, correspondence
readiness, and classifier follow-ups are summarized in the
[Phase 7 reproducibility status](docs/phase-7-status.md). The tasks below are
open and ordered by dependency.

#### 1. Evidence access and correspondence

- [ ] Retrieve MATLAB Central releases 1.0.0–1.0.2 through an authorized
      account, run `cmd/audit-original-ma-archive`, and determine whether any
      release contains the original MA Table 6 driver or resolves either IMA
      operator blocker.
- [ ] After explicit outbound-correspondence authorization, create the prepared
      GSASMA request in Drafts for user review and manual transmission. Record
      its selected recipient, Drafts UID, and Message-ID.
- [ ] After explicit outbound-correspondence authorization, create the prepared
      OLCE-MA request in Drafts for user review and manual transmission. Record
      its selected recipient, Drafts UID, and Message-ID.
- [ ] After either request is manually transmitted, record its send date and
      correspondence status without treating dispatch as evidence resolution.

#### 2. Primary evidence recovery

- [ ] Original MA: resolve the crossover-rate and Gaussian-mutation-rate
      semantics with provenance-checked author or archival evidence.
- [ ] AOBLMOA: resolve all ten blockers, including the exact table-generating
      code, complete benchmark/CEC drivers, seeds, raw trials, and statistics.
- [ ] DESMA: resolve the initial search radius, population split, base-MA
      settings, evaluation accounting, seeds, and raw per-run results.
- [ ] EOBBMA: recover the complete versioned WSN protocol, reference code,
      seeds, raw trials, and sensor coordinates.
- [ ] GSASMA: resolve `T0`, the temperature update, `tau_i`, all four SMA
      probability bounds, fitness orientation, seeds, and raw 30-run outputs.
- [ ] HMMA: resolve all nine clarification blockers, including the seven preset
      questions plus seeds and raw runs.
- [ ] MPMA: recover the implementation, complete configuration, seed schedule,
      raw benchmark/governor trials, and Simulink model.
- [ ] OLCE-MA: resolve the Chebyshev recurrence/seed, sequence lifecycle,
      component equation, and offspring-cardinality conflict, preferably with
      reference code and a deterministic mutation trace.

#### 3. Evidence-gated implementation and validation

- [ ] Original MA: validate the recovered evidence, add an explicit 2020 IMA
      preset, and add equation-, accounting-, and source-conformance tests.
- [ ] AOBLMOA: add an explicit 2023 preset and paper-family runners plus
      equation-, lifecycle-, accounting-, feasibility-, statistics-, and
      source-conformance tests; compare against Tables 5–23 only as supported
      by recovered evidence.
- [ ] DESMA: validate recovered code/data, then add an exact Table 3 preset and
      comparison.
- [ ] EOBBMA: add an explicit 2025 WSN preset and runner plus equation-,
      accounting-, statistics-, and source-conformance tests; compare against
      Tables 5–8 only as supported by recovered trials.
- [ ] GSASMA: implement the confirmed annealing recurrence and SMA mating, then
      validate them against the author material.
- [ ] HMMA: add an explicit 2022 preset plus equation-, accounting-, and
      source-conformance tests; compare against Table 1 only as supported by
      recovered trials.
- [ ] MPMA: add an explicit 2022 preset and benchmark/governor runners plus
      equation-, accounting-, model-, statistics-, and source-conformance
      tests; compare against Tables 5–10 only as supported by recovered trials.
- [ ] OLCE-MA: implement author-confirmed Chebyshev mutation over the confirmed
      offspring set and add equation-level regression tests.

#### 4. Exact trials

- [ ] Run and archive exact trials only for variants whose implementation and
      protocol fidelity gates are closed. Keep results from before the
      correctness audit separate from corrected-implementation results.
