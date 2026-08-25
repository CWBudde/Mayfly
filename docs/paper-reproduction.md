# Paper-reproduction experiments

The `paper-reproduction` command provides the repeatable experiment protocol
for MA, DESMA, OLCE-MA, EOBBMA, GSASMA, HMMA, MPMA, and AOBLMOA. It uses the
post-v0.7 implementation state produced by the correctness audit. Do not mix
its output with historical Mayfly results generated before that audit.

Run a quick end-to-end check:

```bash
./scripts/run-paper-experiments.sh \
  -output paper-results/smoke \
  -benchmarks sphere \
  -dimensions 2 \
  -runs 1 \
  -iterations 5
```

Run the complete classic baseline (six functions at 10D and 30D, 30 paired
runs, 2,000 iterations, all eight variants):

```bash
./scripts/run-paper-experiments.sh -output paper-results/classic
```

Use `-workers N` to execute independent runs concurrently. Each run index uses
`base_seed + run_index`, and that seed is paired across every selected variant,
so changing the worker count does not change the cost samples. For example:

```bash
./scripts/run-paper-experiments.sh \
  -output paper-results/classic \
  -workers 8 \
  -seed 20260825
```

The command accepts comma-separated `-benchmarks`, `-variants`, and
`-dimensions`. Benchmark names are `sphere`, `rastrigin`, `rosenbrock`,
`ackley`, `griewank`, and `schwefel`. Variant names are `ma`, `desma`,
`olce-ma`, `eobbma`, `gsasma`, `hmma`, `mpma`, and `aoblmoa`.

Use `-max-evaluations N` when a source protocol specifies an objective-call
budget. `-iterations` remains a safety ceiling and must be high enough to
consume the budget. A run that falls short is rejected. If the budget ends
partway through a generation, the remaining candidates are not passed to the
objective and cannot replace the best solution found within the budget.

## Original MA protocol audit

The original MA paper reports 50 replications at 95,000 function evaluations
per run. It uses a population of 40 (20 male and 20 female mayflies) and reports
the tuned attraction constants `a1 = 1`, `a2 = 1.5`, visibility coefficient
`beta = 2`, and basic-MA dance and random-flight values of `0.1`. These facts
come from the publisher's [validation and comparison section](https://www.sciencedirect.com/science/article/pii/S036083522030293X).
The authors' [reference implementation](https://github.com/KZervoudakis/Mayfly-Optimization-Algorithm-Python)
also confirms the 20/20 population shape and exposes a later demo configuration.

The harness can now enforce the paper's exact 95,000-call budget. An
`original-ma-2020` preset is intentionally not yet exposed: the paper compares
its fully improved MA after a basic/VGMA/SMA/IMA tuning sequence, while the
current default configuration reflects the authors' later reference code.
Encoding a preset before the exact Appendix A configuration and the reference
rows have been transcribed would conflate those two protocols.

## Output contract

Each benchmark/dimension pair produces a raw CSV and the complete comparison
result as JSON. `manifest.json` records the exact benchmark bounds, seed
schedule, run, iteration, and optional function-evaluation limits, runtime,
source revision when available, and a JSON snapshot of every variant's
non-problem parameters. Common run settings and benchmark-specific problem
fields are recorded separately.

Treat these fields as the primary reproducibility record:

- `seed`
- `best_cost`
- `function_evaluations`
- `iterations`
- `error`

`execution_seconds` depends on the machine and workload. Variants may use
different numbers of objective evaluations per iteration, so a paper comparison
must report `function_evaluations` rather than relying only on iteration count.

## Scope and paper-specific protocols

The default command is a common, controlled classic-function baseline. It does
not assert that all eight source papers used the same functions, bounds,
dimensions, populations, stopping criteria, or evaluation budgets. Reproducing
a paper's published table requires a protocol derived from that paper or its
author implementation, followed by a separate output directory and archived
manifest.

In particular, GSASMA's temperature recurrence and crossover/mutation
probability bounds are not fully specified in the accessible publication. The
official OLCE-MA pseudocode figures specify Chebyshev mutation over all `N`
crossover offspring, invalidating Mayfly's earlier fittest-offspring tie
assumption, but the accessible figures do not expose the exact recurrence and
component mutation equation. Until that stage is corrected from an
authoritative equation or reference implementation, results using Mayfly's
documented defaults must be labeled as current-library baselines, not exact
paper replications.
