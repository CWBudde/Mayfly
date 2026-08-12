# Changelog

All notable changes to Mayfly are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and releases use [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
