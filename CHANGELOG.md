# Changelog

All notable changes to Mayfly are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and releases use [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/CWBudde/Mayfly/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/CWBudde/Mayfly/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/CWBudde/Mayfly/compare/v0.2.1...v0.3.0
[0.2.1]: https://github.com/CWBudde/Mayfly/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/CWBudde/Mayfly/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/CWBudde/Mayfly/releases/tag/v0.1.0
