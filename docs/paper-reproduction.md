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

## Output contract

Each benchmark/dimension pair produces a raw CSV and the complete comparison
result as JSON. `manifest.json` records the exact benchmark bounds, seed
schedule, run and iteration counts, runtime, source revision when available,
and a JSON snapshot of every variant's non-problem parameters. Common run
settings and benchmark-specific problem fields are recorded separately.

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
