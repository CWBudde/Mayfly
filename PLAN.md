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
  - [x] Source-audit the original MA paper's 50-run, 95,000-evaluation, 20/20
        population protocol; encode its Appendix A tuning grid, Table 6 benchmarks,
        and Basic MA/VGMA/SMA/IMA rows in
        `docs/reference-data/original-ma-2020-table6.json`.
  - [x] Audit all four versions of the original MA authors' Mendeley archive and
        record why its simplified 50D Sphere/Rastrigin demos are not the Table 6
        implementation.
  - [x] Add a machine-labeled `descriptive_non_reproduction` comparison of corrected
        current MA against the published Basic MA row without presenting modern
        variants as the unavailable historical VGMA/SMA/IMA implementations.
  - [x] Exhaust the original MA publisher and author-linked public artifacts;
        encode the unresolved crossover-rate and Gaussian-mutation semantics as
        two stable blocker IDs in a machine-readable, send-ready clarification
        request, and expose the same gate in descriptive comparison output.
  - [ ] Obtain primary author or archival evidence resolving both original MA
        IMA operator blockers.
  - [ ] Validate that evidence, implement an explicit original-MA-2020 IMA
        preset, and add equation-, accounting-, and source-conformance tests.
  - [x] Source-audit HMMA's 50-run, 1,000-iteration Table 1 protocol and encode all
        published MA/IMA/AMMA/OCMA/HMMA aggregates in
        `docs/reference-data/hmma-2022-table1.json`. The artifact deliberately makes
        no reproduction claim and preserves the missing population, F3/F7
        dimensions, seeds, raw runs, `a4`, inconsistent `ub`/`lb` tuple, and
        Equation 10/`theta` conflict.
  - [x] Exhaust HMMA's publisher, DOI-registry, corresponding-author ORCID, and
        institutional public artifacts; encode seven stable exact-preset blockers
        and two historical-comparison data blockers in a machine-readable,
        send-ready clarification request linked from the Table 1 reference.
  - [ ] Obtain primary HMMA author or archival evidence resolving all nine
        clarification blockers.
  - [ ] Validate that evidence, implement an explicit HMMA-2022 preset, add
        equation-, accounting-, and source-conformance tests, and compare against
        Table 1 with claims limited by the recovered seeds and raw runs.
  - [x] Source-audit DESMA's 51-run, 30-dimensional, 300,000-evaluation CEC2013
        protocol and encode all 28 published DESMA Table 3 mean-error/rank rows in
        `docs/reference-data/desma-2022-table3.json`. The artifact makes no
        reproduction claim and preserves the missing initial radius, population
        split, seeds, base-MA settings, raw runs, and best-versus-worst replacement
        mismatch.
  - [x] Add official CEC2013 D=30 benchmark support for DESMA's Table 3 protocol,
        including all 28 source-compatible evaluators, caller-supplied official
        data, the 300,000-evaluation budget, and source-derived conformance tests.
  - [x] Add a DESMA Table 3 runner mode that loads CEC2013 data, performs 51 runs
        under the exact objective-call budget, computes mean absolute error, and
        labels its output non-reproduction until the remaining fidelity gates close.
  - [x] Resolve and implement DESMA's Equation 16 elite-attractor/replacement
        lifecycle and its `L` in `[-1,1]` crossover semantics; add equation-level
        regression tests.
  - [x] Audit the DESMA paper, supplement, publisher exports, metadata, and public
        code/data registries; encode a machine-readable, send-ready clarification
        request for the missing protocol and run data, and expose its stable blocker
        IDs in fixed-run output.
  - [ ] Obtain author or archival answers for every DESMA clarification blocker,
        validate the supplied code/data, then add an exact preset and comparison.
  - [x] Source-audit and encode paper-specific protocols/reference outputs for
        EOBBMA, MPMA, and AOBLMOA.
    - [x] Resolve EOBBMA's version-of-record identity and encode its public
          wireless-sensor protocol plus all 776 aggregate/statistical cells
          from Tables 5-8 in
          `docs/reference-data/eobbma-2025-tables2-8.json`. Distinguish the
          formally unlinked 2023 SSRN preprint from the 2025 journal issue,
          preserve the malformed Table 7 Case 1/GWO p-value, identify
          figure-only outputs, and make no reproduction claim.
    - [x] Exhaust EOBBMA's publisher, both DOI-registry records, SSRN record,
          and corresponding-author ORCID; encode seven stable exact-preset
          blockers and three historical-comparison blockers in a
          machine-readable, send-ready clarification request linked from the
          Tables 2-8 reference.
    - [ ] Obtain primary author or archival evidence resolving all ten EOBBMA
          clarification blockers, including the complete versioned protocol,
          reference code, seeds, raw trials, and sensor coordinates.
    - [ ] Validate that evidence, implement an explicit EOBBMA-2025 WSN preset
          and runner, add equation-, accounting-, statistics-, and
          source-conformance tests, and compare against Tables 5-8 with claims
          limited by the recovered seeds and raw trials.
    - [x] Resolve MPMA's paper identity and MMA-versus-MPMA naming; encode the
          30-run benchmark protocol, 35-run governor protocol, and all 110
          published MMA aggregate/performance cells from Tables 5-7 and 9-10
          in `docs/reference-data/mpma-2022-tables1-10.json`. Preserve the
          malformed Table 3 Foxholes range, inconsistent Table 7 F18
          best/median cells, missing Table 9 success rate, and figure-only
          outputs without making a reproduction claim.
    - [x] Exhaust MPMA's publisher, DOI registries, all four author ORCIDs,
          current institutional contact path, and public code/data registries;
          encode five stable exact-preset blockers and three historical-
          comparison/model blockers in a machine-readable, send-ready
          clarification request linked from the Tables 1-10 reference.
    - [ ] Obtain primary MPMA author or archival evidence resolving all eight
          clarification blockers, including the implementation, complete
          configuration, raw trials, seed schedule, and Simulink model.
    - [ ] Validate that evidence, implement an explicit MPMA-2022 preset and
          benchmark/governor runners, add equation-, accounting-, model-,
          statistics-, and source-conformance tests, and compare against
          Tables 5-10 with claims limited by the recovered seeds and raw trials.
    - [x] Source-audit AOBLMOA's open article and paper-linked MATLAB code;
          encode the 30-run, 1,000-iteration Tables 5-6 protocol and all 19
          published AOBLMOA rows in
          `docs/reference-data/aoblmoa-2023-tables5-6.json`. The artifact makes
          no reproduction claim and preserves the paper/code conflicts, absent
          seeds/raw runs, incomplete benchmark source, and current-library
          fidelity gates.
    - [x] Encode AOBLMOA's 30-, 50-, and 100-dimensional stability protocol
          and all 30 published AOBLMOA rows from Tables 7-9 in
          `docs/reference-data/aoblmoa-2023-tables7-9.json`, preserving the
          source's apparent Table 8 f10 standard-deviation inconsistency and
          making no reproduction claim.
    - [x] Encode all 49 AOBLMOA benchmark-case Wilcoxon p-value rows and
          Friedman rank rows, plus both published summary rows, from Tables
          10-11 in `docs/reference-data/aoblmoa-2023-tables10-11.json`. Preserve
          the source's rank-sum method and non-average tie ranks, and make no
          reproduction claim while raw samples and the statistics driver are
          unavailable.
    - [x] Encode all 30 AOBLMOA CEC2017 average, standard-deviation, and
          per-function rank rows plus both Friedman summary rows from Table 13
          in `docs/reference-data/aoblmoa-2023-table13.json`. Preserve the
          missing AO/RSA f2 values and printed ranks, the apparent f14/f28
          inconsistencies, and make no reproduction claim while the dimension,
          evaluation budget, seeds, raw runs, and CEC2017 driver are unavailable.
    - [x] Encode all 200 AOBLMOA/SASS/COLSHADE/sCMAgES objective-summary
          values for the ten selected CEC2020 real-world constrained problems
          from Tables 14-23 in
          `docs/reference-data/aoblmoa-2023-tables14-23.json`. Map the paper
          acronyms to official RC15/16/17/21/22/25/26/29/30/33 metadata and
          budgets without inferring that AOBLMOA followed those budgets, and
          preserve the paper/official-source inconsistencies and missing
          feasibility, penalty, seed, driver, and raw-run evidence as explicit
          non-reproduction gates.
    - [x] Exhaust AOBLMOA's open article, DOI/ORCID records, paper-linked
          MATLAB repository, and public code/data registries; encode five stable
          exact-preset blockers and five experiment/comparison blockers in a
          machine-readable, send-ready clarification request linked from the
          Tables 5-6 reference.
    - [ ] Obtain primary AOBLMOA author or archival evidence resolving all ten
          clarification blockers, including the exact table-generating code,
          complete benchmark/CEC drivers, seeds, raw trials, and statistics.
    - [ ] Validate that evidence, implement an explicit AOBLMOA-2023 preset and
          paper-family runners, add equation-, lifecycle-, accounting-,
          feasibility-, statistics-, and source-conformance tests, and compare
          against Tables 5-23 with claims limited by the recovered evidence.
  - [ ] Run and archive exact trials only for variants whose implementation and
        protocol fidelity gates are closed.
- [x] Provide experiment scripts. `cmd/paper-reproduction` and
      `scripts/run-paper-experiments.sh` run paired seeded trials of all eight variants,
      export raw CSV/JSON, and record the protocol, revision, runtime, bounds, and
      configuration in a manifest.
- [ ] Correct OLCE's chaotic-exploitation stage against the complete published equations.
  - [x] Audit the publisher's authoritative pseudocode figures: they specify
        Chebyshev mutation over all `N` crossover offspring, invalidating the earlier
        equal-fitness/fittest-offspring premise.
  - [x] Document the current Logistic-map, fittest-offspring stage as a library extension.
  - [x] Exhaust the publisher article and figures, DOI/discovery records,
        current institutional contact path, and public code/data registries;
        encode three stable recurrence/seed, sequence-lifecycle, and component-
        mutation blockers in a machine-readable, send-ready clarification request.
  - [ ] Obtain primary author or archival evidence resolving all three OLCE-MA
        clarification blockers, preferably the original reference implementation
        plus a deterministic chaotic-mutation trace.
  - [ ] Implement Chebyshev mutation over all `N` offspring and add equation-level
        regression tests.
- [ ] Calibrate GSASMA's undocumented annealing recurrence/defaults and SMA crossover and
      mutation probability bounds.
  - [x] Document the current cooling and ordinary configured mating as library
        extensions rather than inventing constants absent from the paper.
  - [x] Hold the extension's initial temperature through the ordinary first half and
        begin cooling at the exact `2*iteration >= MaxIterations` annealing boundary.
  - [x] Audit the version-of-record article, publisher flowchart, Crossref metadata,
        GitHub, Zenodo, Mendeley Data, Figshare, and OSF; no public author
        implementation or raw seeded results were available in August 2026.
  - [ ] Obtain the authors' `T0`, `T` update, undefined `tau_i` sequence, four SMA
        probability bounds, fitness orientation, seeds, and raw 30-run outputs.
  - [ ] Implement the exact recurrence and SMA mating, then validate it against the
        author material.
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
