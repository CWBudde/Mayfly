# Mayfly Algorithm Suite - Remaining Tasks

> ## ⚠️ Known defects in `functions.go` — benchmark edge-case handling
>
> Found 2026-08-23 while porting this file verbatim into the sibling
> [Dragonfly](https://github.com/MeKo-Christian/dragonfly) library. The same defects exist
> in both repositories and should be fixed in both, together, so the two benchmark suites
> stay numerically comparable.
>
> **1. `Levy([])` panics.** With an empty input, `n == 0` and the function indexes `w[n-1]`
> — `index out of range [0] with length 0`. It is the only benchmark function that panics
> rather than returning a value. A metaheuristic should never call it with an empty vector,
> but a panic is the wrong failure mode for a pure scoring function, and it makes the suite
> unsafe to fuzz.
>
> **2. Empty-input handling is inconsistent across the suite.** Measured behaviour of
> `f([])` for all 15 functions:
>
> | Result    | Functions                                                                                                                                |
> | --------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
> | `0`       | Sphere, Rastrigin, Rosenbrock, Griewank, Schwefel, Zakharov, DixonPrice, Michalewicz, BentCigar, Discus, Weierstrass, ExpandedSchafferF6 |
> | `NaN`     | Ackley, HappyCat — both divide by `n`                                                                                                    |
> | **panic** | Levy                                                                                                                                     |
>
> Pick one convention for the whole suite and apply it uniformly. Returning `0` for an
> empty vector is defensible (an empty sum is zero) and is already what 12 of the 15 do;
> whichever is chosen, document it in the file's package comment and add a table-driven
> test that asserts it for every function at once, so a new benchmark cannot land without
> a deliberate answer.
>
> **Not defects — checked and cleared.** Two things in this file look wrong at a glance and
> are not. Recording them here so they are not "fixed" into actual bugs later:
>
> - `Levy` uses `sin(math.Pi*wi + 1)`. The `+ 1` is correct and is part of the standard
>   definition (Surjanovic & Bingham:
>   `f = sin²(πw₁) + Σ(wᵢ−1)²[1+10sin²(πwᵢ+1)] + (w_d−1)²[1+sin²(2πw_d)]`, `wᵢ = 1+(xᵢ−1)/4`).
>   It is **not** a mistranscription of `π/4`.
> - `ExpandedSchafferF6` closes the loop with the wrap-around pair `g(x[n-1], x[0])`. That
>   cyclic term is part of the CEC definition of the expanded function, not an extra.
>
> ### Tasks -- resolved 2026-08-23, in both repositories together
>
> The convention is `f([]) == 0` for every single-objective function, stated in the
> `functions.go` file comment and asserted for all fifteen at once by
> `TestBenchmarkFunctionsEmptyInput`.
>
> - [x] Fix the `Levy([])` panic
> - [x] Choose and document a single empty-input convention for `functions.go`
> - [x] Add a table-driven test asserting that convention for all 15 functions
> - [x] Mirror the fix and the test in the Dragonfly library's ported `functions.go`

> ## ⚠️ Known defect in `selector.go` — `estimateLandscape` measures the box, not the function
>
> Found 2026-08-23 while building the sibling
> [Dragonfly](https://github.com/MeKo-Christian/dragonfly) library's own selector. Dragonfly
> deliberately did **not** port this heuristic; Mayfly still uses it.
>
> `estimateLandscape` (selector.go:239) averages the raw finite-difference gradient
> **magnitude** over random samples and compares that average against three absolute
> thresholds — `> 100 → Deceptive`, `> 10 → Rugged`, `< 0.01 → NarrowValley`, else `Smooth`.
> A gradient magnitude carries the units of the search space, so the classification scales
> with the width of the box rather than with the shape of the function.
>
> Measured, d = 30, calling `estimateLandscape` directly:
>
> | Function  | `[-5,5]`    | `[-50,50]`  | `[-500,500]` |
> | --------- | ----------- | ----------- | ------------ |
> | Sphere    | `Rugged`    | `Deceptive` | `Deceptive`  |
> | Rastrigin | `Deceptive` | `Deceptive` | `Deceptive`  |
>
> Sphere is a smooth convex bowl — the textbook `Smooth` case, and exactly what
> `RecommendForBenchmark` hard-codes it as at selector.go:353. The sampler never returns
> `Smooth` for it at any tested scale, and rates it _more_ deceptive than Rastrigin as the
> box grows. This matters because `Landscape` is not inert: `SelectVariant` routes on
> `== Deceptive` (selector.go:87, 126) and `== NarrowValley` (selector.go:130), so a
> misclassification silently changes which algorithm a caller is told to use.
>
> Two smaller things in the same function:
>
> - The accumulator is named `gradientVariance` but holds a **mean magnitude** — no variance
>   is ever computed, despite the comment at selector.go:274 saying "high gradient variance
>   suggests rugged landscape". Ruggedness is about how much the gradient _changes_, which
>   is the quantity the name promises and the code does not compute.
> - `ClassifyProblem` and `estimateLandscape` draw through `unifrnd(lower, upper, nil)`,
>   i.e. the package-level RNG, so a classification is not reproducible from a seed and
>   cannot be made so by the caller.
>
> **Suggested fix** — make the statistic scale-free. Dragonfly's replacement uses random
> line scans and reports (a) average direction changes per line for modality and (b) total
> variation normalised by _that line's own value range_ for landscape. Both are invariant
> under rescaling the box and under affine rescaling of the objective, and they separate
> Sphere/Zakharov from Rastrigin/Schwefel/Ackley. See `Dragonfly/selector.go`.
>
> Note that `Deceptive` and `NarrowValley` arguably cannot be established by sampling at
> all: both are claims about where the optimum sits relative to the surrounding terrain,
> not about local roughness. Dragonfly's version returns only `Smooth` or `Rugged` and
> documents that the other two are for the caller to set.
>
> ### Tasks -- resolved 2026-08-24
>
> Dragonfly's line-scan classifier is now Mayfly's too, with two thresholds retuned:
> `smoothRoughness` 3.0 -> 2.2 and `multimodalTurningPoints` 6.0 -> 5.0. Schwefel sat on
> both of Dragonfly's values, so its verdict flipped with the seed; the new ones sit in the
> measured gaps and hold over forty seeds. **Dragonfly likely has the same latent
> flakiness** -- its own `TestClassifyProblemSeparatesSmoothFromRugged` expects Schwefel
> `Rugged` at a single seed, which passes there by luck of the draw. Worth re-tuning that
> library to match.
>
> One reconciliation gap is documented rather than fixed: Griewank over `[-600,600]` reads
> `Unimodal`/`Smooth`, against the table's `HighlyMultimodal`/`Rugged`. Its cosine ripples
> are a few units wide and of order one tall against a value range of order 100000 -- both
> aliased by the scan spacing and negligible in the total variation. The table entry is the
> better answer for an optimizer working near the optimum and stays as it is.
>
> - [x] Replace the absolute gradient-magnitude thresholds with a scale-free statistic
> - [x] Rename `gradientVariance`, or compute the variance the name and comment promise
> - [x] Thread an `*rand.Rand` through `ClassifyProblem` / `estimateLandscape` instead of
>       drawing from the package-level RNG
> - [x] Add a test asserting a function classifies identically on `[-5,5]` and `[-500,500]`
> - [x] Reconcile `ClassifyProblem`'s answers with the hard-coded `RecommendForBenchmark`
>       table, which is the de facto expected output

> ## ⚠️ Correctness audit — release blocker for v0.7.0
>
> A repository-wide adversarial review completed 2026-08-24 found that the green root
> test suite overstates the library's readiness. The optimizer can discard a superior
> female solution, parallel execution can follow a different algorithm from sequential
> execution, and several named variants materially differ from their cited papers.
> AOBLMOA is also described as multi-objective even though the public optimizer accepts
> only a scalar objective. Historical seeded results for corrected algorithms must not be
> compared as though they came from the same implementation.
>
> ### Core correctness
>
> - [x] Include both populations in initialization, global-best updates, stopping, and reporting
> - [x] Sort populations before the first rank-dependent update
> - [x] Preserve offspring sex and generate mutations separately from incumbent males/females
> - [x] Give sequential and parallel execution one deterministic proposal/evaluate/commit model
> - [x] Treat configuration as immutable and report only truthful reproducibility metadata
> - [x] Centralize finite-value, bounds, velocity, variant, and cancellation validation
>
> ### Algorithm fidelity
>
> - [x] Restore the published scalar-distance and genetic stages of standard MA
> - [x] Correct MPMA ranked median, nonlinear gravity, application sites, and citation
> - [x] Correct AOBLMOA Aquila strategies, crossover, stochastic opposition, and naming
> - [x] Reimplement EOBBMA's fitness-dependent updates and mirror boundary handling
> - [x] Restore DESMA's dynamic elite schedule, replacement order, defaults, and no-op behavior
> - [ ] Move OLCE orthogonal/chaotic operators to their published lifecycle stages and support D > 3
> - [ ] Separate faithful GSASMA and HMMA implementations and citations
>
> ### Public API and tooling safety
>
> - [x] Remove fake multi-objective presets/recommendations; retain only honest Pareto utilities
> - [x] Make Pareto archives nondominated, validated, dimension-safe, and alias-safe
> - [x] Unify direct, builder, and JSON validation; make templates strict and round-trippable
> - [x] Make classifier failures/budgets explicit and its stability statistic translation-invariant
> - [x] Correct comparison targets, tied statistics, failed-run aggregation, and atomic export
> - [ ] Replace panic-prone public helpers and mutable global tables with validated APIs/copies
>
> ### Verification and release
>
> - [x] Add independent equation-level fixtures and seeded sequential/parallel parity tests
> - [x] Add non-finite, aliasing, cancellation, config-reuse, Pareto, and statistical regressions
> - [x] Build, vet, and test every nested `go.mod`; fix currently excluded broken examples
> - [ ] Pin the lint toolchain and run format, tidy, lint, unit, race, integration, and module checks
> - [x] Publish v0.7.0 migration notes and invalidate pre-fix deterministic benchmark baselines
>
> ### Remaining audit follow-ups (recorded 2026-08-24)
>
> - [ ] Place OLCE's chaotic initialization and orthogonal-learning operator at the exact
>       lifecycle points from the cited paper; the dimensional orthogonal array and factor
>       analysis are fixed, but the runtime still applies both stages to elite incumbents.
> - [ ] Validate GSASMA and HMMA end-to-end against their respective equations and lifecycle
>       descriptions. They now have separate variants, configuration, registry entries, and
>       documentation, but that separation alone is not proof of paper fidelity.
> - [ ] Finish replacing panic-prone exported helpers with explicit errors. The mutable
>       orthogonal-array global is gone and returned data is copied, but compatibility helpers
>       still sanitize some invalid input instead of reporting it.
> - [ ] Resolve the pinned golangci-lint v2.12.2 baseline (76 findings: 44 `wsl_v5`, plus
>       context propagation, bounds analysis, complexity, shadowing, field alignment, unused
>       helpers, formatting, and smaller style findings). Build, vet, tidy, unit, race,
>       integration, and nested-module checks pass; lint is the only incomplete check above.

## High Priority

### Phase 1: Advanced Features

#### 1.1 Parallel Fitness Evaluation (Core)

- [x] Implement worker pool for bounded concurrency
- [x] Parallelize male population fitness evaluation
- [x] Parallelize female population fitness evaluation
- [x] Thread-safe global best update mechanism (mutex/atomic)
- [x] Configuration: `Config.MaxWorkers` (default: runtime.NumCPU())
- [x] Configuration: `Config.EnableParallel` flag for backward compatibility
- [x] Benchmarks comparing sequential vs parallel performance

**Rationale**: For expensive objective functions (simulations, ML training), this provides 10-20x speedup on multi-core systems. Core populations have 20+ individuals evaluated per iteration.

#### 1.2 Parallel Genetic Operators

- [x] Parallel crossover offspring evaluation
- [x] Parallel mutation offspring evaluation
- [x] Thread-safe offspring slice management
- [x] Race detector tests (`go test -race`)

**Rationale**: Offspring generation (NC + NM individuals) happens every iteration. Parallelization reduces iteration time significantly.

#### 1.3 Parallel Variant-Specific Enhancements

- [x] DESMA: Parallel elite candidate generation and evaluation
- [x] OLCE-MA: Parallel orthogonal learning candidate evaluation (4 per elite)
- [x] EOBBMA: Parallel opposition point evaluation
- [x] GSASMA: Parallel Golden Sine candidate evaluation
- [x] AOBLMOA: Parallel Aquila strategy evaluation
- [x] MPMA: Thread-safe median position calculation

**Rationale**: Variant-specific operations add significant computational overhead. OLCE generates 4 candidates per elite (top 20%), DESMA generates 5+ elite candidates. These are natural parallelization targets.

#### 1.4 Multi-Algorithm Parallel Comparison Framework

- [x] Concurrent execution of multiple algorithms on same problem
- [x] Enhanced comparison example using goroutines
- [x] Statistical comparison utilities with parallel runs
- [x] Results aggregation and visualization

**Rationale**: Users often want to compare MA, DESMA, OLCE-MA, EOBBMA, GSASMA, MPMA, AOBLMOA on same problem. Running 7 algorithms sequentially takes 7x time; parallel execution is much faster.

#### 1.5 Parallel Infrastructure Testing & Validation

- [x] Comprehensive race condition tests
- [x] Verify deterministic results with same seed (challenging with parallel execution)
- [x] Performance benchmarks showing speedup vs core count
- [x] Validate no fitness evaluations are lost or duplicated
- [x] Test with cheap vs expensive objective functions
- [x] Document when parallel execution is beneficial vs overhead

**Rationale**: Parallel execution introduces complexity (race conditions, non-determinism). Thorough testing is critical to ensure correctness and measure actual performance gains.

**Phase 1 Total Effort Estimate**: Items 1.1-1.5 represent ~5-8x the original single "Parallel Execution" item. This is a major feature requiring careful design for thread-safety across all 7 algorithm variants.

#### 1.6 Convergence Detection

- [x] Early stopping criteria
- [x] Stagnation detection
- [x] Adaptive iteration limits

#### 1.7 Constraint Handling

- [x] Penalty function methods
- [x] Feasibility rules
- [x] Constraint-handling utilities

### Phase 2: Release Preparation

#### 2.1 Static Analysis & Lint Remediation

- [x] Capture and classify the current `golangci-lint` baseline
- [x] Resolve production-code layout and alignment findings
- [x] Resolve inline error-handling findings
- [x] Resolve whitespace/style findings in tests
- [x] Verify formatting, module tidiness, `go vet`, and `golangci-lint`

#### 2.2 Coverage & Test Quality

- [x] Measure package and function-level coverage
- [x] Add focused tests for the lowest-covered behavior
- [x] Verify 80%+ statement coverage
- [x] Run unit, race, and integration test suites

#### 2.3 Performance Profiling & Optimization

- [x] Establish reproducible benchmark baselines
- [x] Capture CPU and memory profiles for representative workloads
- [x] Identify and optimize material hot paths
- [x] Add regression benchmarks for optimized paths
- [x] Document profiling commands and results

#### 2.4 Release Engineering & Publishing

- [x] Define and document the semantic-versioning policy
- [x] Add a release checklist and validation workflow
- [x] Create `CHANGELOG.md` from the existing release history
- [x] Verify package metadata and documentation for pkg.go.dev
- [x] Tag a release and verify publication on pkg.go.dev

---

## Medium Priority

### Phase 3: Advanced Features (continued)

#### 3.1 Logging & Monitoring

- [x] Structured logging interface
- [x] Progress callbacks
- [x] Convergence curve export

#### 3.2 Advanced Benchmarks

- [ ] CEC2017 benchmark suite
- [ ] CEC2020 benchmark suite
- [ ] Real-world engineering problems

### Phase 4: Documentation

#### 4.1 API Documentation

- [x] Add code examples to docs
- [x] Create quick reference guide
- [x] Document all parameters

#### 4.2 Tutorials

- [ ] Getting started tutorial
- [ ] Algorithm selection guide
- [ ] Parameter tuning tutorial
- [ ] Custom objective function guide

#### 4.3 Real-World Examples

- [ ] Neural network hyperparameter tuning
- [ ] Resource allocation problems
- [ ] Scheduling problems
- [ ] Feature selection

---

## Low Priority

### Phase 5: Community Setup

- [ ] CONTRIBUTING.md
- [ ] Issue templates
- [ ] Pull request templates

### Phase 6: Research Reproducibility

- [ ] Reproduce original paper results (MA, DESMA, OLCE-MA, EOBBMA, GSASMA, MPMA, AOBLMOA)
- [ ] Provide experiment scripts
