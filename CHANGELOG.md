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

### Added

- `Config.CrossoverGamma`, the blend-crossover expansion factor, defaulting to
  the reference `DefaultCrossoverGamma` of `0.4`. Zero, negative, `NaN` and
  `Inf` all resolve to the default, because zero reproduces the interpolation
  bug this release fixes.
- `CrossoverBlend`, the general form of `Crossover` taking an explicit gamma.
  `Crossover` keeps its signature and now delegates with
  `DefaultCrossoverGamma`.

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
