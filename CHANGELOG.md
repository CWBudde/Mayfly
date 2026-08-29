# Changelog

All notable changes to Mayfly are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and releases use [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- A custom-objective-function guide covering the scalar minimization contract,
  normalized heterogeneous and projected variables, objective scaling and
  maximization, constraints, finite failure handling, concurrency and
  expensive evaluations, validation tests, and common mistakes, with a
  CI-tested runnable nonlinear curve-fitting example.
- A parameter-tuning tutorial covering baseline-first experimental design,
  parameter prioritization and coupling, heuristic auto-tuning limits, paired
  seeds, exact evaluation budgets, held-out validation, configuration handoff,
  and common mistakes, with a CI-tested runnable population-size sweep.
- An algorithm-selection guide covering the selector's scalar-objective scope,
  problem metadata, variant heuristics, score interpretation, classifier
  limits, budget-matched empirical validation, and common mistakes, with a
  reproducibly seeded runnable example kept in sync by a documentation test.
- A focused getting-started tutorial with a complete reproducibly seeded
  optimizer, result interpretation, custom-objective and normalized-bound
  guidance, troubleshooting, and a runnable example kept in sync by a
  documentation regression test.
- A machine-readable, send-ready OLCE-MA chaotic-exploitation clarification
  request with three stable evidence blockers for the Chebyshev recurrence and
  seed, sequence lifecycle, and offspring-component mutation equation. The
  unsent request records the publisher, DOI, institutional-contact, and public
  code/data audit without treating Mayfly's Logistic-map extension as the
  published algorithm.
- A machine-readable, send-ready AOBLMOA Tables 5-23 clarification request
  with five stable exact-preset blocker IDs for strategy/branch, Aquila mean,
  Levy scale, population/lifecycle, and evaluation accounting semantics, plus
  five separately gated benchmark/CEC driver, seed/raw-statistics, and source-
  inconsistency questions. The Tables 5-6 reference links the unsent request
  without claiming reproduction.
- A machine-readable, send-ready MPMA/MMA Tables 5-10 clarification request
  with five stable exact-preset blocker IDs for `a4`, median-pool semantics,
  genetic operators, initialization/boundaries, and evaluation accounting,
  plus three separately gated historical run, Simulink-model, and source-
  inconsistency questions. The Tables 1-10 reference links the unsent request
  without claiming reproduction.
- A machine-readable, send-ready EOBBMA Tables 5-8 clarification request with
  seven stable exact-preset blocker IDs for the versioned protocol,
  population, Lévy/EOBL semantics, WSN objective, coordinates/boundaries, and
  evaluation accounting, plus three separately gated historical run,
  seed, and raw-data questions. The source reference links the unsent request
  without claiming reproduction.
- A machine-readable, send-ready HMMA Table 1 clarification request with seven
  stable exact-preset blocker IDs for population, base-MA/crossover settings,
  benchmark geometry, `a4`, and Equations 10 and 12, plus two separately gated
  historical seed/raw-data questions. The Table 1 reference links the unsent
  request and preserves all nine IDs without claiming reproduction.
- A machine-readable, send-ready original MA Table 6 clarification request
  with two stable IMA exact-preset blocker IDs for the conflicting
  crossover-rate and Gaussian-mutation-rate semantics. Descriptive comparison
  summaries now carry the same blocker IDs and artifact path without claiming
  reproduction. Its source audit now also records the metadata-only Technical
  University of Crete entry and the authentication-gated MATLAB Central
  releases 1.0.0-1.0.2 that predate the explicitly simplified demo as an
  unrecovered archival lead, not resolved evidence. A deterministic intake
  command now validates recovered ZIPs, records archive/content/per-file
  SHA-256 provenance, and emits explicitly non-resolving source-review
  indicators for the Table 6 protocol and ambiguous operators.
- Machine-readable MPMA/MMA Tables 1-10 provenance with the 30-run,
  function-evaluation-limited benchmark protocol, the 35-run governor
  protocol, and all 110 published MMA aggregate/performance cells. The
  artifact preserves the source's malformed Foxholes range, inconsistent F18
  best/median values, and missing promised success rate while recording the
  unpublished `a4`, median-pool, operator, seed, raw-data, and Simulink gates;
  it makes no reproduction claim.
- Machine-readable EOBBMA wireless-sensor provenance for the canonical
  version-of-record Tables 2-8: eight deployment scenarios, the reported
  40-member and 1,000-iteration settings, all 64 fitness summaries, all 64
  coverage/redundancy/distance summaries, 56 Wilcoxon cells, and the complete
  Friedman result. The artifact distinguishes the formally unlinked SSRN
  preprint, preserves a malformed printed p-value, and records the remaining
  protocol and raw-data gates instead of claiming a reproduction.
- Machine-readable AOBLMOA Tables 14-23 CEC2020 constrained-problem
  provenance with all 200 objective-summary values for AOBLMOA, SASS,
  COLSHADE, and sCMAgES. The artifact maps the ten paper acronyms to official
  RC problem metadata and budgets while recording that AOBLMOA budget
  adherence, feasibility outputs, penalty semantics, seeds, drivers, and raw
  runs are unavailable, so it makes no reproduction claim.
- Machine-readable AOBLMOA Table 13 CEC2017 provenance with all 30 functions'
  averages, standard deviations, per-function ranks, and both Friedman summary
  rows for eight algorithms. The artifact preserves missing F2 comparator
  values and apparent F14/F28 source anomalies, and records the omitted
  dimension, evaluation budget, seeds, raw runs, and CEC2017 driver as explicit
  non-reproduction gates.
- Machine-readable AOBLMOA Tables 10-11 statistical provenance with all 49
  Wilcoxon rank-sum p-value and Friedman-rank rows, both published summaries,
  and explicit non-reproduction gates for missing raw samples/statistics code
  and Mayfly's different paired signed-rank method.
- Machine-readable AOBLMOA Tables 7-9 dimension-stability provenance with all
  30 published AOBLMOA rows for F1-F10 at dimensions 30, 50, and 100, plus an
  explicit note preserving the source's apparent Table 8 F10 inconsistency and
  the existing non-reproduction blockers.
- Machine-readable AOBLMOA Tables 5-6 provenance with the paper's 30-run,
  1,000-iteration classic-function protocol, all 19 published AOBLMOA rows,
  the exact paper-linked MATLAB commit, and explicit non-reproduction gates for
  missing seeds/raw runs/source functions plus paper/code/library conflicts.
- A machine-readable, send-ready DESMA Table 3 clarification request that
  records the public-source audit, six stable exact-preset blockers,
  acceptable primary evidence, and an unsent corresponding-author draft.
  Fixed-run summaries now carry the same blocker IDs and clarification path.
- Official CEC2013 D=30 benchmark support for DESMA's protocol, with all 28
  functions, caller-supplied official transformation data, a 300,000-evaluation
  budget, and source-evaluator conformance coverage.
- A fixed DESMA Table 3 experiment mode that runs all 28 CEC2013 D=30 functions
  for 51 seeded trials at exactly 300,000 objective calls, exports raw results
  and mean absolute errors, and machine-labels them as a descriptive
  non-reproduction while DESMA fidelity gates remain open.
- Added machine-readable provenance for every version of the original MA
  authors' Mendeley code archive, explicitly distinguishing its simplified
  50D demo protocol from the paper's 5D Table 6 experiment.
- Added an opt-in published-reference mode to the experiment command. It runs
  corrected current MA on Table 6's exact case geometry and emits a strictly
  validated, SHA-256-linked descriptive comparison that explicitly disclaims
  historical reproduction and unavailable VGMA/SMA/IMA implementations.
- A reproducible experiment command and wrapper for paired seeded comparisons
  of all eight variants, with CSV/JSON run data and a protocol manifest that
  records revision, runtime, bounds, seed schedule, and configuration.
- Exact per-run objective-evaluation budgets for `ComparisonRunner` and the
  paper-reproduction command, including rejection when the iteration safety
  ceiling is too short to consume the requested budget.
- Machine-readable transcription of the original MA paper's Appendix A tuning
  grid, Table 6 protocol, and published Basic MA/VGMA/SMA/IMA statistics, plus
  the fixed-dimension Eggcrate and Beale benchmark functions used by that table.
- A forty-seed classifier calibration regression covering smooth and rugged
  benchmark landscapes.
- `Himmelblau` benchmark function, the sixteenth in the standalone suite:
  multimodal with four equal global minima, extended to n dimensions by summing
  over disjoint coordinate pairs, with the unpaired coordinate in an odd
  dimension scored as its square. Available in the WebAssembly demo.
- Exact CEC2017 (29 usable functions; F2 was officially removed) and CEC2020
  (10 functions) benchmark evaluators with validated `fs.FS` loaders for the
  organizers' shift, rotation, and shuffle data, official biases and budgets,
  normalized Mayfly configurations, and regression coverage for every function.
- Constrained spring, welded-beam, pressure-vessel, and speed-reducer engineering
  benchmarks with physical bounds, published reference designs, mixed-variable
  projection, and normalized configuration adapters.

### Changed

- Align OLCE's documented Logistic-map compatibility extension with the
  publisher pseudocode's all-offspring loop in serial and parallel execution.
  Every crossover child now receives one chaotic candidate and one objective
  evaluation; the unavailable Chebyshev recurrence and component equation
  remain explicit reproduction blockers.
- Correct DESMA's source-guided Equation 16 lifecycle: a strictly improving
  elite now replaces the current best population member and becomes the next
  male attractor, while DESMA crossover uses per-coordinate `L` in `[-1,1]` in
  both sequential and parallel execution.
- OLCE documentation now preserves the source's internal cardinality conflict:
  the publisher pseudocode loops over all crossover offspring, while indexed
  full-text prose names the fittest offspring. The cited OLCGOA blend and the
  earlier Chebyshev-Mayfly paper narrow the strategy lineage but leave the
  recurrence, lifecycle, and exact Equation 12 unresolved. Logistic-map behavior
  remains explicitly labeled a library extension.

## [0.7.0] - 2026-08-24

### Added

- `Config.Seed` and nullable `Result.Seed` provide truthful reproducibility;
  opaque caller-owned `*rand.Rand` values no longer produce an invented seed.
- `ClassifyProblemContext` adds validation, cancellation, evaluation budgets,
  scan-only operation, and explicit errors to the landscape classifier.
- HMMA is a separately registered variant with the paper's scheduled
  OBL/Cauchy global-best mutation and artificial gender mutation.
- Validated exported Pareto helpers and defensive archive snapshots.
- `just test-modules` and CI coverage for every nested example module.
- `Config.QMCInit` seeds the initial population from a low-discrepancy sequence
  instead of independent uniform draws: `QMCInitUniform` (the zero value and the
  historical behaviour), `QMCInitSobol`, or `QMCInitHalton`. The builder exposes
  it as `WithQMCInitialPopulation`. Males and females come out of one stream, so
  the halves do not land on top of each other, and the population sits on one
  balanced block of the sequence rather than straddling two.
- `Config.QMCSeed` pins that sequence's scramble. Left at zero the scramble is
  drawn from the run's generator, which is what makes repeated runs start from
  different point sets while a pinned `Config.Rand` still reproduces a whole run.
  Sobol is limited to the 1024 dimensions of the direction numbers `qmc` embeds
  and reports an error naming the ceiling past that; Halton has no ceiling.
- `docs/qmc-initialization.md` is the measurement behind that feature, including
  where it makes no difference. Across the sixteen benchmark problems the effect
  is at chance level -- two significant results for Sobol, none against, against
  about 1.6 expected by chance, and an earlier run of the same study found two
  hits on two *different* problems. **Uniform stays the default**; the sequence
  is worth trying on a problem that does not converge, not assumed to be an
  improvement.
- Dependency: `github.com/cwbudde/qmc v0.2.0`.

### Changed

- **Correctness release:** seeded trajectories from older versions are not
  comparable to v0.7 results. Standard attraction now uses the paper's scalar
  Cartesian distances; crossover retains offspring sex; mutation generates
  `NM` candidates per sex from the matching incumbent population.
- Configuration is immutable during a run, every named variant is exclusive,
  all numeric fields must be finite, explicit velocity limits must be a valid
  pair, and non-finite objective values are rejected rather than rewarded.
  `CrossoverGamma` no longer accepts `NaN` or `Inf` as a fallback signal; zero
  and negative values still resolve to `DefaultCrossoverGamma`.
- Sequential and parallel modes share proposal and commit semantics. Recurrent
  golden-section updates remain sequential so enabling workers does not select
  a different algorithm.
- `ComparisonRunner.TargetCost` is now `*float64`, allowing zero and negative
  targets. Failed runs are explicit and excluded from numeric aggregates.

- `ClassifyProblem` takes a trailing `*rand.Rand`. It and its helpers drew
  through the package-level generator, so a classification could not be
  reproduced from a seed and the caller had no way to make it so. `nil` keeps
  the old behaviour of a fresh generator. **Breaking**: existing calls need a
  fifth argument.
- `ClassifyProblem` estimates modality and landscape from random straight-line
  scans across the box rather than from sample variance and finite-difference
  gradients. Modality counts direction changes per line; landscape sums each
  line's total variation in units of that line's own value range. Both are
  scale-free, so the verdict follows the function instead of the bounds. This is
  the classifier the sibling Dragonfly library uses, with two thresholds
  retuned: `smoothRoughness` sits just above the single-basin bound of 2 rather
  than at 3, and `multimodalTurningPoints` at 5 rather than 6, because Schwefel
  straddled both and its verdict flipped with the seed.
- `ClassifyProblem` no longer returns `Deceptive` or `NarrowValley`. Both are
  claims about where the optimum sits relative to the surrounding terrain, which
  sampling cannot establish; callers who know their problem should set
  `Landscape` on the returned value. The `RecommendForBenchmark` table still
  reports them, since those landscapes are known from the literature.
- The short runs behind `RequiresStableConvergence` are seeded from the passed
  generator, and a failed run is skipped instead of being counted as `+Inf` and
  dragged into the mean.
- Every benchmark function scores an empty position vector as 0, stated as the
  suite's convention in the file comment. `Levy([])` panicked on `w[n-1]`,
  `Ackley([])` and `HappyCat([])` returned NaN from dividing by a zero dimension
  count, and the other twelve already returned 0.
  `TestBenchmarkFunctionsEmptyInput` asserts it for all fifteen against both a
  nil and an allocated empty slice. The same change lands in the sibling
  Dragonfly library, which ported the file verbatim, so the two suites stay
  numerically comparable.

### Fixed

- Corrected DESMA's source attribution and removed unsupported fixed
  performance/overhead claims. The official 51-run CEC2013 protocol and all 28
  Table 3 DESMA aggregates are now preserved in a non-reproduction reference
  artifact, together with the missing settings and population-insertion mismatch.
- Corrected HMMA fidelity documentation after auditing the publisher's Equation
  (10) and Table 1 protocol. The established probability schedule is now
  labeled a compatibility extension, and the published aggregates and all
  unresolved protocol fields are preserved in a non-reproduction reference
  artifact.
- GSASMA's library-extension temperature schedule no longer cools during the
  ordinary-MA first half of a run. Cooling now begins with the paper's
  second-half annealed velocity phase, so the default schedule does not reach
  its temperature floor before it is first used.

- Females now participate in `GlobalBest`, convergence, target termination, and
  stagnation decisions. Populations are sorted before the first ranked update.
- Corrected MPMA's fitness-ranked median and published gravity schedule;
  AOBLMOA's Aquila equations, crossover, and one-scalar uniform opposition;
  EOBBMA's fitness-dependent bare-bones/Lévy updates and boundary pullback;
  DESMA's dynamic inertia, elite no-op behavior, default count, and citation;
  and OLCE's dimension-safe orthogonal design and factor analysis.
- All objective phases consistently reject NaN and both infinities. A run whose
  entire initial population is invalid returns `ErrNoFiniteObjectiveValue`.
- Strict JSON loading rejects unknown, duplicate, and trailing data; generated
  templates are complete valid JSON and are written atomically.
- Pareto archives now retain only finite nondominated solutions, reject mixed
  dimensions and duplicates, and cannot be mutated through aliased pointers.
- Comparison ranking handles ties, Wilcoxon uses exact small-sample or
  tie-corrected asymptotic probabilities, and Friedman applies tie correction.
- Corrected the false claim that AOBLMOA is a multi-objective optimizer. The
  deprecated multi-objective preset now returns an error.
- `wilcoxonSignedRankTest` reported `PValue` as 0 when every pair tied. It
  discards pairs closer than 1e-10, and the early return for an empty remainder
  never touched the field, so the zero value stood -- the most significant
  result the test can express, where the truth in that branch is that there is
  no evidence of any difference at all. It reports 1.

- `ClassifyProblem` classified by the width of the search box rather than by the
  shape of the function. The old `estimateLandscape` averaged the raw
  finite-difference gradient *magnitude* and compared it against absolute
  thresholds, but a gradient magnitude carries the units of the search space, so
  Sphere -- a smooth convex bowl, and hard-coded as `Smooth` in
  `RecommendForBenchmark` -- came back `Rugged` on `[-5,5]` and `Deceptive` on
  `[-500,500]`, rating it more deceptive than Rastrigin as the box grew. The
  verdict is not inert: the selector routes on `Deceptive` and `NarrowValley`,
  so the misclassification silently changed which algorithm a caller was told to
  use. `TestClassifyProblemIsScaleFree` now pins Sphere to the same answer on
  both boxes.
- The accumulator behind that classification was named `gradientVariance` and
  documented as a variance, but held a mean magnitude; no variance was ever
  computed. The statistic it was reaching for -- how much the landscape turns --
  is now measured directly as turning points per line scan.

## [0.6.0] - 2026-08-23

### Added

- `AquilaWeightAuto`, the sentinel that selects AOBLMOA's published branch rule.
  See the Changed and Deprecated entries below.
- `WithPopulationObserver` reports both populations after every completed
  iteration, as a `PopulationSnapshot` of deep copies. `WithProgressObserver`
  only ever carried the global best, so there was no way to watch the swarm
  itself — to animate it, to measure its diversity, or to see how a variant
  actually moves. The hook is opt-in and separate from `Progress` precisely
  because copying `NPop+NPopF` position and velocity vectors once per iteration
  is not free, and most observers only want the best cost.
- A browser demo of the library in `examples/wasm-demo`, published to
  <https://cwbudde.github.io/Mayfly/> by the new `wasm-demo-pages` workflow. Two
  pages — a Swarm Lab that animates both populations over a benchmark landscape,
  and a Variant Shootout that runs the comparison framework across all seven
  variants. Everything shown is computed by this library compiled to `js/wasm`;
  no part of the algorithm is reimplemented in JavaScript. `just run-wasm-demo`
  builds and serves it locally.

### Changed

- **AOBLMOA now implements its source paper.** This is a behavior change, not a
  fix: results for a given seed differ from every prior release, and
  **recorded AOBLMOA results should be regarded as produced by a different
  algorithm.** Three structural deviations from Zhao, Y.; Huang, C.; Zhang, M.;
  Cui, Y., "AOBLMOA: A Hybrid Biomimetic Optimization Algorithm for Numerical
  Optimization and Engineering Design Problems", _Biomimetics_ 2023, 8(4), 381
  ([10.3390/biomimetics8040381](https://doi.org/10.3390/biomimetics8040381)):
  - The Mayfly/Aquila switch was a probability; it is a deterministic fitness
    test. The paper keeps the attraction branches of the Mayfly Algorithm
    verbatim and replaces only the nuptial dance (males, Eq. 29) and the random
    flight (females, Eq. 30). Which branch an individual takes follows from
    whether a better solution dominates it. At the old default of
    `AquilaWeight = 1.0` the attraction terms — the Mayfly half of the hybrid —
    never ran at all. Measured before the change on D=8 Sphere over 500
    iterations and 15 seeds: median `6.66e-05` at weight 1.0 against
    `2.12e-82` at weight 0.0.
  - Opposition-based learning was in the wrong pipeline slot. It fired at
    `p = 0.3` on a post-Aquila candidate inside the update phase, used the plain
    reflection `lb + ub − x`, and left Gaussian mutation in place. The paper
    replaces _offspring mutation_ with stochastic opposition on every offspring
    after crossover, ungated: Eq. (31) `x̃ = (lb + ub − x) × r` with
    `r ~ N(0,1)` — the `× r` is essential — followed by the greedy selection of
    Eq. (32). Gaussian mutation is removed rather than gated, so `NM` is inert
    under AOBLMOA; `effectiveNM` reports `0` for it. `OppositionProbability` is
    likewise unread by AOBLMOA. The evaluation budget per iteration is now
    `NPop + NPopF + 2·nc`.
  - Within a phase, the Aquila strategy is fixed by the individual's sex rather
    than by a coin flip, so AOBLMOA consumes no randomness on that decision.

  `StrategySwitch` was declared, defaulted, documented and tested as a tunable
  through v0.5.1, but nothing read it: the phase split was hard-coded at two
  thirds. It is now honored. `0` still resolves to `MaxIterations * 2 / 3`, and
  the resolution is never written back, so a `Config` reused with a different
  budget rescales. A value at or beyond `MaxIterations` is legal and means
  "never exploit"; negative values are rejected by both `Optimize` and
  `ValidateConfig`.

  `initializeAOBLMOA` is gone. It mutated the `Config` in place on the first
  iteration, which would have clamped the new `AquilaWeightAuto` sentinel away —
  the same reused-`Config` hazard `NCAuto` already forbids. Validation covers
  every range it clamped.

  Both AOBLMOA paths now share one move function that evaluates nothing, so the
  sequential and parallel implementations agree by construction rather than by
  two hand-maintained copies of the same branch cascade — the arrangement that
  drifted twice before.

  **Three points where the paper is ambiguous or self-contradictory** are each
  isolated to one function, with a comment naming the question and saying
  exactly what to change to flip it:
  - The female branch inequality — Eq. (30) or the Algorithm 1 pseudocode,
    which states the opposite? Carried by `aoblmoaFemaleTakesAttraction`.
    Resolved for Eq. (30), which matches `prepareStandardFemale`: AOBLMOA
    replaces branches of exactly that algorithm.
  - Which sex gets which strategy pair? The equations give males X2/X4 and
    females X1/X3; the abstract swaps them. Carried by `aoblmoaStrategyFor`.
    **Open**; it follows the equations, because a formal specification outranks
    prose.
  - Is Eq. (31)'s `r` drawn per solution or per dimension? The subscript
    notation is inconsistent. Carried by `stochasticOppositionPoint`.
    **Open**; per dimension.

### Deprecated

- `Config.AquilaWeight`. The published AOBLMOA has no such knob; the branch is a
  fitness test. The field now defaults to the `AquilaWeightAuto` sentinel
  (`-1`), which selects the paper's behavior, and both `Optimize` and
  `ValidateConfig` accept it. Setting a probability in `[0, 1]` restores the
  pre-v0.6.0 random branch draw, complete with the standard update's own dance
  and flight branches. Note that it restores the branch choice only: the
  offspring stage is the paper's either way, so a pre-v0.6.0 run is not
  reproduced exactly.

### Fixed

- A `Mu` outside the documented `[0, 1]` panicked the optimizer partway through
  a run instead of being rejected. `Mu` is a fraction of the dimensions, and the
  mutation operators turn it into a count with `ceil(Mu * ProblemSize)` and then
  slice a permutation of the dimensions to that length. A `Mu` above 1 asks for
  more dimensions than exist, a negative one asks for a negative count, and
  `NaN` converts to the most negative `int`; all three ended in
  `slice bounds out of range`. `Mu = 1.0000001` was enough. The range was
  already documented and already enforced by `LoadConfig`, so only configs built
  programmatically were exposed — `Optimize` never checked it. The check now
  lives in `validateOffspring` alongside the other genetic-operator parameters,
  so both paths agree and the run fails before it starts. The exported operators
  `MutateGaussian`, `MutateCauchy` and `HybridMutate` take the rate directly and
  now saturate rather than panic: at or below `0`, and at `NaN`, they mutate
  nothing; at or above `1` they mutate every dimension.
- The two AOBLMOA implementations computed different iterations. The sequential
  path moved and re-evaluated every male before the females ran, so a female
  compared herself against her paired male's fresh cost; the parallel path moved
  the males but deferred their evaluation, so the same comparison saw this
  iteration's position with last iteration's cost, and the Aquila mean position
  and random peer mixed moved and unmoved members. Both paths now update the
  whole swarm against the state it had when the iteration began — the same
  pre-move pairing the standard variant uses — and produce identical results for
  a given seed and configuration.
- `Optimize` rejects `NPopF` greater than `NPop` instead of panicking. Every
  female update phase pairs `females[i]` with `males[i]`, so a larger female
  population indexed past the end of the male slice; the standard and EOBBMA
  paths crashed with `index out of range`, sequentially and in parallel.
- `ValidateConfig` rejects `NPopF` greater than `NPop` as well. It checked that
  both populations were positive but not that they could pair, so a
  configuration loaded from a file passed validation and then failed at the
  start of a run its caller believed was already checked. Both entry points now
  report the pairing failure before the offspring checks, so the same
  configuration produces the same error whichever one sees it first.
- `Optimize` rejects `UseAOBLMOA`, `UseEOBBMA`, and `UseMPMA` in combination
  instead of silently ignoring all but one of them. They replace the same
  position-update phase of the main loop, which is a switch, so enabling MPMA
  alongside AOBLMOA never computed a median position and the median term was
  dropped while the configuration still claimed MPMA was in use.
- The Friedman test reported inverted significance. Its p-value came from
  `chiSquareCDF`, whose small-`df` branch returned
  `exp(-x/2) * (x/2)^(df/2)` — a curve that is neither a cumulative
  distribution nor monotonic — and the result was then used as a lower-tail
  probability. The two errors did not cancel: a strongly significant result
  read as "no difference" and vice versa. The function is now
  `chiSquareSurvival`, the true upper-tail probability via the regularized
  incomplete gamma function. For chi-square 14.207 on 6 degrees of freedom the
  reported p-value was 0.7053; it is 0.0274. **Any `Significant` verdict from
  `FriedmanTestResult` recorded before this release should be discarded.**

## [0.5.1] - 2026-08-22

### Fixed

- Crossover is blend crossover again. The coefficient was drawn from `U(0, 1)`,
  which makes every offspring a convex combination of its two parents: the
  population's convex hull can only shrink from one generation to the next, and
  mating can never restore spread the swarm has lost. The reference
  implementation draws it from `U(-gamma, 1+gamma)` with `gamma = 0.4`
  (Zervoudakis & Tsafarakis 2020; the author's Python port,
  `KZervoudakis/Mayfly-Optimization-Algorithm-Python`, `operators.py`
  `ContinousCrossover` and `ma.py` `MA(..., gamma=0.4)`), so offspring may land
  outside the parental interval. **This changes the search trajectory of every
  run and every variant**, including previously seeded ones.
- OLCE-MA's chaotic exploitation is a greedy local search again instead of an
  unconditional random kick. Every offspring was displaced in every dimension,
  every iteration, with no acceptance test and no decay — and because the
  logistic map at `r = 4` has an arcsine stationary distribution concentrated
  near 0 and 1, the typical displacement was near-maximal rather than
  near-zero. The stage now generates one chaotic neighbor per elite male, with
  a radius that decays linearly to zero over the run, and accepts it only when
  it is not worse. On 10-dimensional Rastrigin over 30 seeds (500 iterations,
  default population) the mean best cost of `NewOLCEConfig` improves from
  8.4392 to 2.8035; the standard algorithm scores 3.6211 on the same seeds, so
  OLCE-MA was previously more than twice as bad as the variant it enhances.
- OLCE-MA's orthogonal learning no longer spends its evaluation budget when
  `OrthogonalFactor` is 0. Both the blend and the perturbation scale with the
  factor, so all four L4 candidates were bit-identical to the parent male and
  the stage was a guaranteed no-op that still cost four evaluations per elite
  male per iteration — 8000 of 38540 evaluations (20.8%) in the run above.
  `ApplyOrthogonalLearning`, `ApplyOrthogonalLearningToElite` and the pooled
  path now return immediately for a non-positive factor.
- **AOBLMOA moved only part of its swarm.** Every individual drew against
  `AquilaWeight` to decide whether it took an Aquila-Optimizer step, and an
  individual that lost the draw was skipped outright — no Aquila step, and no
  Mayfly velocity update either. A comment claimed the main loop would apply
  the standard update, but the AOBLMOA branch of that loop replaces the
  standard update rather than following it, so a losing individual simply stood
  still, unevaluated, until crossover happened to touch it. At the former
  default `AquilaWeight` of 0.5, half the population was frozen every
  iteration. Both the sequential and the parallel evaluation paths were
  affected. Every individual now moves every iteration, taking either an Aquila
  step or the ordinary Mayfly update.
- AOBLMOA's function-evaluation count is now the number of evaluations actually
  performed rather than a formula that assumed one evaluation per individual —
  an assumption the skipping made false.
- EOBBMA's elite opposition-based learning had no observable effect on the
  search. It generated the *static* opposition point `a + b - x` of each elite
  and accepted it only when it beat that elite. Reflecting a good solution
  through the middle of the search space lands in the mirror region, where it
  essentially never wins, so the candidates were evaluated and always
  discarded: runs with `OppositionRate` 0 and 1 produced bit-identical results
  on every seed tried, differing only in the wasted evaluation count. The
  operator now implements elite opposition-based learning as published,
  reflecting through the dynamic interval spanned by the elite set with a
  random coefficient `k ~ U(0, 1)` and resampling out-of-bounds reflections
  from that interval.
- GSASMA's "Golden Sine Algorithm" component was not the Golden Sine Algorithm.
  The implemented rule `x + r1*sin(r2)*|r3*best - x|` is the Sine Cosine
  Algorithm update; the `GoldenRatio` constant was never read anywhere in the
  package, so nothing in the variant used the golden ratio. The update now
  follows Tanyildizi & Demir (2017):
  `x*|sin(r1)| - r2*sin(r1)*|x1*best - x2*x|`, where `x1` and `x2` are the
  fixed coefficients derived from `τ=(sqrt(5)-1)/2`. The published Eq. (10)
  has no scale parameter or recurrent narrowing; `GoldenFactor` is deprecated
  and ignored.

### Added

- `Config.CrossoverGamma`, the blend-crossover expansion factor, defaulting to
  the reference `DefaultCrossoverGamma` of `0.4`. Zero, negative, `NaN` and
  `Inf` all resolve to the default, because zero reproduces the interpolation
  bug this release fixes.
- `CrossoverBlend`, the general form of `Crossover` taking an explicit gamma.
  `Crossover` keeps its signature and now delegates with
  `DefaultCrossoverGamma`.

### Changed

- **Behaviour change for `UseOLCE` runs.** Chaotic exploitation moved off the
  crossover and mutation offspring and onto the elite males, so an OLCE-MA run
  no longer reproduces a result recorded under v0.5.0. `ChaosFactor` keeps its
  default of 0.1 but now means the *initial* radius of a decaying neighborhood
  rather than a constant displacement; setting it to 0 disables the stage.
  The stage costs one evaluation per elite male per iteration.
- **`NewAOBLMOAConfig` now defaults `AquilaWeight` to 1.0** (was 0.5). The
  published algorithm has no such probability: it moves every individual either
  by Mayfly attraction or by an Aquila strategy chosen from the iteration
  phase. 1.0 is the closest match to the paper, and it was the only setting the
  skipping defect left unharmed. Set `config.AquilaWeight = 0.5` to keep a
  half-and-half blend, which now genuinely blends instead of freezing half the
  swarm.
- **The optimizer no longer maintains a Pareto archive.** `Optimize` built one
  for AOBLMOA and pushed the whole population into it every iteration, paying
  NSGA-II pruning for it, while nothing in the search or in `Result` ever read
  it back. Removing it leaves results bit-for-bit identical and returns the
  time. The Pareto toolkit stays exported for callers who want a front:
  `NewParetoArchive`, `Add`, `AddFromMayfly`, `GetBestSolution` and the new
  `(*ParetoArchive).UpdateFromPopulation`, which replaces the unexported
  `updateParetoArchive`. `ArchiveSize` now only sizes an archive the caller
  builds.

### Removed

- `applyGSASMAToEliteMales`, `applyGoldenSineToElite` and
  `goldenSineConvergence`, unexported helpers that were unreachable and carried
  a second copy of the mislabelled Sine-Cosine update.

## [0.5.0] - 2026-08-22

### Fixed

- The crossover offspring count now tracks the population. `NC` was an absolute
  `20` that no caller had reason to revisit, so raising `NPop` bought a larger
  swarm and not one additional crossover. At `NPop` 4096 the same ten pairs
  mated while 4086 members only followed the global best, which quietly reduced
  the algorithm to plain PSO — and, because every variant configuration derives
  from `NewDefaultConfig`, did so for DESMA, OLCE-MA, EOBBMA, MPMA, GSASMA and
  AOBLMOA alike. Any comparison between variants run at a raised population was
  therefore measuring a handicapped algorithm.

### Added

- `NCAuto`, the default value of `NC`, deriving the offspring count from `NPop`
  and the new `NCRatio` (default 1.0, giving `NC == NPop`). A written `NC` still
  wins, including the `0` that disables crossover, so no field a caller sets is
  silently replaced.
- `Selection` with `SelectionRank` and `SelectionTournament`, plus
  `TournamentSize`, making parent selection configurable rather than hardcoded.
  `SelectionRank` remains the default and the historical behaviour.

### Changed

- **Breaking for callers that raised `NPop`.** A run at any population other
  than the default 20 now performs a different number of crossovers and will
  not reproduce a result recorded under v0.4.0. Set `config.NC = 20` to restore
  the old count exactly. At the default `NPop` of 20, `NCAuto` resolves to 20
  and results are unchanged.
- `validateOffspring` judges the resolved offspring count rather than the
  written `NC`, and rejects an unknown `Selection`, a negative `NCRatio`, and a
  negative `TournamentSize`.

### Notes

- `SelectionTournament` is implemented and configurable but is **not** the
  default: at `NPop` 20 it reduced Griewank 10D success from above 70% to 60%
  (Standard MA) and 20% (DESMA) on this repository's regression suite. Once
  `NC` scales, rank pairing mates the fitter half of the population at every
  size, so the elitism it appeared to cause was the `NC` bug rather than the
  pairing rule.


## [0.4.0] - 2026-08-12

### Added

- Structured optimization lifecycle logging compatible with `log/slog`.
- CSV and JSON convergence-curve export on optimization results.
- Optional bounded parallel evaluation across the core optimizer, genetic
  operators, and all algorithm-specific enhancement phases.
- Parallel multi-algorithm comparison with statistical aggregation and result
  visualization utilities.
- Target-cost, stagnation, and adaptive-iteration convergence controls.
- Inequality and equality constraints with feasibility-rule and penalty-based
  candidate ranking.

### Changed

- Added compile-checked package examples, an API quick reference, and a
  complete configuration field index.
- Documented when parallel evaluation offsets its scheduling overhead.
- Expanded race, determinism, evaluation-accounting, convergence, and
  constraint-handling tests.

## [0.3.0] - 2026-08-09

### Added

- Context cancellation, progress observation, and caller-provided initial
  populations for optimization runs.
- Explicit termination reasons and run-lifecycle documentation.

## [0.2.1] - 2026-07-30

### Fixed

- Updated CI and lint configuration for reliable race and coverage runs.
- Stabilized seeded integration scenarios.
- Restored package documentation and resolved static-analysis findings.

## [0.2.0] - 2026-07-30

### Changed

- Renamed `Result.BestSolution` to `Result.ConvergenceCurve` to reflect that
  the field contains per-iteration costs. This is a breaking API change.

### Fixed

- Validate offspring count `NC` before optimization, preventing out-of-range
  access for incompatible population and offspring sizes.

## [0.1.0] - 2025-10-21

### Added

- Initial implementation of seven Mayfly Algorithm variants.
- Benchmark functions, algorithm selection, comparison utilities, examples,
  JSON configuration, and algorithm documentation.

[Unreleased]: https://github.com/CWBudde/Mayfly/compare/v0.7.0...HEAD
[0.7.0]: https://github.com/CWBudde/Mayfly/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/CWBudde/Mayfly/compare/v0.5.1...v0.6.0
[0.5.1]: https://github.com/CWBudde/Mayfly/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/CWBudde/Mayfly/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/CWBudde/Mayfly/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/CWBudde/Mayfly/compare/v0.2.1...v0.3.0
[0.2.1]: https://github.com/CWBudde/Mayfly/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/CWBudde/Mayfly/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/CWBudde/Mayfly/releases/tag/v0.1.0
