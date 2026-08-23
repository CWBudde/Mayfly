# Changelog

All notable changes to Mayfly are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and releases use [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
  section points of a golden section search over `[-π, π]` that narrows by
  `1/φ` after every candidate. `GoldenFactor` keeps its meaning as a scale on
  the second term and its default of 1.0 reproduces the published rule.

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

[Unreleased]: https://github.com/CWBudde/Mayfly/compare/v0.5.1...HEAD
[0.5.1]: https://github.com/CWBudde/Mayfly/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/CWBudde/Mayfly/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/CWBudde/Mayfly/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/CWBudde/Mayfly/compare/v0.2.1...v0.3.0
[0.2.1]: https://github.com/CWBudde/Mayfly/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/CWBudde/Mayfly/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/CWBudde/Mayfly/releases/tag/v0.1.0
