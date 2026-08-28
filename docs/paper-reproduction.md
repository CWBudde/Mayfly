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
`ackley`, `beale`, `eggcrate`, `griewank`, and `schwefel`. Variant names are
`ma`, `desma`, `olce-ma`, `eobbma`, `gsasma`, `hmma`, `mpma`, and `aoblmoa`.

Use `-max-evaluations N` when a source protocol specifies an objective-call
budget. `-iterations` remains a safety ceiling and must be high enough to
consume the budget. A run that falls short is rejected. If the budget ends
partway through a generation, the remaining candidates are not passed to the
objective and cannot replace the best solution found within the budget.

## Original MA protocol audit

The original MA paper reports 50 replications at 95,000 function evaluations
per run. It uses a population of 40 (20 male and 20 female mayflies) and reports
the tuned attraction constants `a1 = 1`, `a2 = 1.5`, visibility coefficient
`beta = 2`, and basic-MA dance and random-flight values of `0.1`. Its IMA adds
the `0.1 * (upper-lower)` velocity limit, `g = 0.8`, dance/flight damping of
`0.77`, and Gaussian mutation with a reported rate of `0.1`. These facts come
from the publisher's [validation and comparison section](https://www.sciencedirect.com/science/article/pii/S036083522030293X).

The complete Table 6 protocol, Appendix A tuning grid, benchmark bounds, and
all Basic MA/VGMA/SMA/IMA reference statistics are transcribed in the
[machine-readable reference file](reference-data/original-ma-2020-table6.json).
The scalable cases are 5D Sphere, Rosenbrock, Rastrigin, and Ackley; Eggcrate
and Beale remain the fixed 2D functions defined by the paper's benchmark table,
despite Table 6's generic “at 5 dimensions” caption.

Use that file to produce a descriptive comparison between the corrected
current-library MA and the published Basic MA row:

```bash
./scripts/run-paper-experiments.sh \
  -published-reference docs/reference-data/original-ma-2020-table6.json \
  -output paper-results/original-ma-table6-current-ma \
  -runs 50 \
  -max-evaluations 95000 \
  -iterations 2000 \
  -workers 8
```

This mode derives all six dimensions and bounds from the reference file and
accepts only `ma`. Besides the raw result files and manifest, it writes
`published-comparison.json`. That file is deliberately labeled
`descriptive_non_reproduction` and `reproduction_claim: false`; it reports
protocol-alignment fields, current-minus-published descriptive statistics, the
unknown convention behind the paper's reported standard deviation, and the
unavailable historical VGMA/SMA/IMA implementations. The paper did not publish
its seeds, so no paired test or pass/fail threshold is inferred.

The authors' [Mendeley code archive](https://data.mendeley.com/datasets/5w58s8hhz2)
also does not contain the Table 6 experiment. Its own metadata calls every
version a “simplified Matlab demo”; version 1 (published 2020-07-10) runs only
50D Sphere/Rastrigin. It crosses every pair with a separate coefficient per
coordinate sampled from `[-1,1]`, creates one mutant per iteration, treats
`mu = 0.01` as the fraction of coordinates to mutate, and uses
`sigma = 0.1 * (upper-lower)`. Version 2 changes the crossover coefficient to
`[0,1]` and changes scalar attraction distances to component-wise distances.
Those changes are useful provenance for a separate demo protocol, but they do
not explain the paper's crossover rate `0.95` or Gaussian mutation rate `0.1`.
The authors' later [Python implementation](https://github.com/KZervoudakis/Mayfly-Optimization-Algorithm-Python)
likewise exposes a demo configuration rather than the tuned IMA protocol.

The harness can now enforce the paper's exact 95,000-call budget. An
`original-ma-2020` preset is intentionally not yet exposed. The remaining
blocker is operator semantics, not missing numerical transcription:

- Section 3.2.3 defines arithmetic offspring using a coefficient `L`, while
  Section 4.2 calls the configured operator “single point uniform crossover” at
  rate `0.95`; it does not say how that rate gates mating.
- Equation 22 defines additive Gaussian mutation, but the reported mutation
  rate `0.1` is not identified as a candidate probability, coordinate
  probability, or Gaussian standard deviation.

The later author code has separate knobs for those interpretations and uses
different tuned parameters. Selecting one silently would create a plausible
experiment, not a reproduction. The reference file records these ambiguities so
that a future author clarification or archived experiment implementation can
close them without redoing the protocol audit.

The repository also provides a
[machine-readable clarification request](reference-data/original-ma-2020-clarification-request.json).
It pins the two unresolved questions under the stable IDs
`crossover_operator_and_rate` and `gaussian_mutation_rate_semantics`, records
acceptable primary evidence, and includes an unsent correspondence draft. The
generated `published-comparison.json` carries the same artifact path, target
algorithm (`ima`), and blocker IDs in `exact_preset_status`. An exact historical
IMA preset remains blocked until both answers are supported by primary evidence
and equation-level tests; the unpublished Table 6 seeds remain a separate
limitation on paired comparison.

## DESMA protocol audit

The DESMA paper's experiment protocol and complete Table 3 DESMA column are
transcribed in the
[machine-readable DESMA reference](reference-data/desma-2022-table3.json). The
paper reports 51 independent runs of all 28 CEC2013 functions at 30 dimensions,
a 300,000-objective-evaluation limit and population size 50. Section 4.2 selects
`k = 10`; the artifact explicitly records that applying it to Table 3 is an
inference because Section 4.3 does not restate `k`. The paper also
specifies the dynamic-inertia, dance/flight, and radius multiplier values. The
artifact records the published mean error and rank for every function, the
overall ranks, and the reported DESMA-versus-comparator t-test counts.

This is a source audit, not a reproduction preset. The paper omits the initial
search radius, does not say whether population 50 means each sex or both sexes
combined, and does not give enough base-MA operator settings to derive exact
evaluation accounting. Seeds and raw runs are absent. The official supporting
RAR supplies CEC2013 evaluator/input files but no DESMA implementation.

The follow-up requirements are pinned in a
[machine-readable clarification request](reference-data/desma-2022-clarification-request.json).
It records the 2026-08-28 public-source audit, six stable blocker IDs, acceptable
primary evidence, affected configuration/output fields, and an unsent
correspondence draft for the paper's corresponding author. The exact-preset gate
can close only when every blocker has a non-null answer backed by primary
evidence. Preparing this request does not imply that it was sent or that an
author response was received.

The implementation-level ambiguities now have a documented source-guided
resolution. An improving elite replaces the current best population member and
becomes the next iteration's global attractor. Equation 16's printed `>` is
treated as the same minimization-sign error found in the paper's base Equation
3. DESMA's Equations 6-7 crossover independently samples each coordinate's `L`
uniformly from `[-1,1]` and shares it between the complementary siblings; the
draw granularity follows the
[cited original MA authors' implementation](https://github.com/KZervoudakis/Mayfly-Optimization-Algorithm-Python/blob/749251dfd95fe3606fde0c67bbef4c042d4202e8/operators.py#L3-L9)
because the DESMA authors did not publish code. These resolutions do not fill the
remaining experiment-protocol gaps, so results stay non-reproductions.
Mayfly's automatic 10%-of-span initial radius is still a compatibility choice.
The official CEC2013 D=30 suite is available through `CEC2013Suite`. Run the
fixed current-library DESMA protocol with an already extracted copy of the
official input data:

```bash
./scripts/run-paper-experiments.sh \
  -desma-table3-data /path/to/CEC2013-or-extracted-S1 \
  -output paper-results/desma-table3-current \
  -workers 8 \
  -seed 20260825
```

The mode accepts supplement roots containing `data/input_data/`, organizer
roots containing `input_data/`, or the input-data directory itself. It does not
download or redistribute the official files. The 28 functions, D=30, current
DESMA variant, 51 runs, and 300,000 objective calls per run are fixed; generic
`-benchmarks`, `-dimensions`, `-variants`, `-runs`, `-iterations`, and
`-max-evaluations` overrides are rejected.

This full command makes 428.4 million objective calls. Each function produces
raw CSV and JSON, while `desma-table3-summary.json` computes the mean of
`abs(best_cost - known_minimum)` across its 51 runs. Both that summary and the
manifest carry `protocol_id: desma-2022-table3`,
`comparison_kind: descriptive_non_reproduction`, and
`reproduction_claim: false`.
The summary also embeds `exact_preset_status`, including the clarification
artifact path and blocker IDs, so archived runs remain self-describing if they
are separated from these docs.

The command uses 4,167 iterations as the minimum ceiling under the recorded
current defaults: 40 initialization calls, 72 calls per complete iteration,
and eight objective calls in the final partial iteration. That ceiling affects
DESMA's inertia schedule and is current-implementation accounting, not a value
reported by the paper. An exact preset and published comparison remain blocked
until the implementation and protocol gaps below are closed.

## HMMA protocol audit

The HMMA paper's Table 1 protocol and aggregate rows are transcribed in the
[machine-readable HMMA reference](reference-data/hmma-2022-table1.json). The
paper reports 50 independent runs, 1,000 iterations, a parameter tuple, and
Best/Worst/Average/Std/Median rows for MA, IMA, AMMA, OCMA, and HMMA on F3, F7,
and F15.

That artifact is deliberately labeled `reproduction_claim: false`. The paper
does not publish population size, F3/F7 dimensions, seeds, or raw per-run
results. Its parameter tuple prints `ub = 0.1`, `lb = 10`, omits the `a4`
coefficient used by Equation (7), and reports `theta = 0.005` while its printed
Equation (10), `Ps = -exp((1-t/Iter_MAX)^20) + theta`, stays negative throughout
the stated run. Mayfly's existing `-exp(-t/Iter_MAX) + theta` schedule and
defaults are now explicitly documented as compatibility extensions rather than
paper-calibrated settings. An exact HMMA preset and comparison remain blocked
on author clarification or the supporting data, which the paper makes available
on request. The
[machine-readable clarification request](reference-data/hmma-2022-clarification-request.json)
turns that dependency into seven stable exact-preset questions covering the
population, base-MA/crossover settings, benchmark geometry, `a4`, Equation (10),
and Equation (12). Two separate question IDs request the historical seed
schedule and raw runs/reference code needed to elevate an aggregate-only
comparison into a reproduction claim. The included correspondence draft is
explicitly marked `not_sent`.

## EOBBMA protocol audit

EOBBMA's version-of-record wireless-sensor targets are transcribed in the
[machine-readable Tables 2–8 reference](reference-data/eobbma-2025-tables2-8.json).
The authoritative journal record is Zhou, Zhang, and Zhou, volume 50 issue 2,
pages 719–739 (2025), published online 25 March 2024 under DOI
`10.1007/s13369-024-08899-6`. The similarly titled 2023 SSRN manuscript has a
different title, author order/list, and DOI (`10.2139/ssrn.4381249`); neither
DOI registry record formally links the two, so the artifact labels their
preprint/version-of-record relationship as an inference rather than metadata.

The artifact records all eight two-dimensional deployment cases from 20 to 500
sensors and dimensions 40 to 1,000. Table 3 prints population size 40, while
the canonical Figure 4 shows 1,000 iterations for every case. All 320 fitness
summary/rank cells from Table 5, 384 coverage/redundancy/moving-distance cells
from Table 6, 56 printed Wilcoxon cells from Table 7, and 16 Friedman cells from
Table 8 are encoded. The malformed Case 1/GWO p-value `1.0354–04` is preserved
as text instead of silently repaired. Figure-only convergence, ANOVA,
histogram, and final deployment plots are identified but not digitized into
invented samples.

This remains a source transcription, not an exact preset or reproduction. The
public primary-source tables and figures do not provide the independent-run
count, seed schedule, raw samples, Equation 17 objective weights, machine
environment, raw initial/final coordinates, or a common objective-evaluation
budget. They also do not clarify whether population 40 means total mayflies or
40 of each sex. The complete preprint and version-of-record prose still need a
cell-by-cell protocol comparison, and no public author implementation or batch
driver has been identified. Mayfly does not yet implement the paper's WSN
objective or expose an EOBBMA paper preset.

The
[machine-readable clarification request](reference-data/eobbma-2025-clarification-request.json)
records the completed publisher/DOI/SSRN/ORCID audit and assigns stable IDs to
seven exact-preset gates and three historical-comparison gates. The exact gate
covers the versioned protocol, population split, Lévy and EOBL semantics, WSN
objective, coordinate initialization/boundaries, and evaluation accounting.
The comparison gate separately covers the run/statistics protocol, seed
schedule, and raw trials/reference code. Its correspondence draft is marked
`not_sent`; no request has been transmitted.

## MPMA protocol audit

The library name MPMA corresponds to the paper's “Mayfly Algorithm Based on
Median Position,” which the authors abbreviate as MMA. The exact article is
Guo Lei, Xu Chang, Yu Tianhang, and Wumaier Tuerxun, “An Improved Mayfly
Optimization Algorithm Based on Median Position and Its Application in the
Optimization of PID Parameters of Hydro-Turbine Governor,” _IEEE Access_ 10
(2022), 36335–36349, DOI
[`10.1109/ACCESS.2022.3160714`](https://doi.org/10.1109/ACCESS.2022.3160714).

The complete public protocol and every tabular MMA output are transcribed in
the [machine-readable Tables 1–10 reference](reference-data/mpma-2022-tables1-10.json).
The benchmark experiment uses 20 male and 20 female mayflies, 30 independent
runs, and function-specific limits: 100,000 evaluations for F1–F10, 10,000 for
F11, 1,000 for F12–F15, and 2,000 for F16–F18. The artifact contains all 90 MMA
Best/Worst/Average/Median/Std cells from Tables 5–7. It also records the
35-run, 50-iteration hydro-turbine governor protocol, both Table 8 working
conditions, all 16 MMA ITAE/iteration cells from Table 9, and all four MMA
overshoot/adjustment-time cells from Table 10.

This is source transcription, not a reproduction preset. Equation 16 introduces
the median-attraction coefficient `a4`, but Table 4 does not assign it a value.
The paper also omits the offspring count, crossover-coefficient range, mutation
scale, initialization details, boundary handling, seeds, raw runs, and exact
objective-call accounting. Its prose does not settle whether the fitness-ranked
median covers males only or both sexes. The Simulink model, solver settings, and
raw governor trials are not publicly archived. No paper-linked or public author
implementation was found in the code/data registries audited on 28 August 2026.

The artifact preserves three source problems instead of guessing corrections:
Table 3 prints the F10 Foxholes range as `[65.536,65.536]`; Table 7 prints an
F18 MMA median better than its “best”; and the Table 9 prose promises a success
rate that the table does not contain. Figures 1–4 and 8–9 remain identified as
figure-only outputs and are not converted into invented raw samples. Mayfly's
`MedianWeight = 0.5`, male-only median pool, alternative gravity schedules, and
weighted median are therefore documented compatibility choices or extensions,
not paper-derived settings.

The remaining questions are pinned in a
[machine-readable clarification request](reference-data/mpma-2022-clarification-request.json).
Its exact-preset gate separately identifies `a4`, median-pool lifecycle,
genetic-operator parameters, initialization/boundary behavior, and objective-
call accounting. Its historical-comparison gate covers the versioned source,
ordered seeds and raw trials, the executable Simulink model, and the three
printed source inconsistencies. The request is send-ready but marked
`not_sent`; no correspondence has been transmitted.

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

AOBLMOA's paper targets are now machine-readable in
[`reference-data/aoblmoa-2023-tables5-6.json`](reference-data/aoblmoa-2023-tables5-6.json)
and
[`reference-data/aoblmoa-2023-tables7-9.json`](reference-data/aoblmoa-2023-tables7-9.json),
with their statistical analysis in
[`reference-data/aoblmoa-2023-tables10-11.json`](reference-data/aoblmoa-2023-tables10-11.json),
and the CEC2017 comparison in
[`reference-data/aoblmoa-2023-table13.json`](reference-data/aoblmoa-2023-table13.json),
and the CEC2020 constrained comparison in
[`reference-data/aoblmoa-2023-tables14-23.json`](reference-data/aoblmoa-2023-tables14-23.json).
Together they record the 30-run, 1,000-iteration protocols, the reported
population and parameters, all 19 original benchmark rows, all 30 F1-F10 rows
at dimensions 30, 50, and 100, and all 49 published Wilcoxon/Friedman rows and
summary values. Table 13 contributes all 30 CEC2017 average,
standard-deviation, and rank rows for eight algorithms plus its Friedman
summaries; Tables 14-23 add 200 objective-summary values for four algorithms
on ten selected CEC2020 constrained problems. They remain descriptive
references rather than runnable reproductions: the paper-linked MATLAB
repository supplies only F1 and a single-run entry point, while seeds, raw
runs, the table driver, and the statistical-analysis driver are absent. The
paper and source also disagree on the Aquila mean and Levy scale, and the
current Go default does not yet use the reported `g=0.9` to `0.4` schedule. The
Tables 7-9 artifact preserves and flags the published Table 8 F10 standard
deviation even though it exceeds that row's reported worst value. The Tables
10-11 artifact also preserves the paper's rank-sum method and non-average tie
ranks; it cannot be regenerated by Mayfly's paired signed-rank comparison
without implementing the paper's method and recovering all raw samples.

Table 13 has additional protocol blockers: the paper does not identify the
CEC2017 dimension, objective-evaluation budget, or seeds, and the author code
contains neither the CEC2017 evaluators nor the comparison driver. The table
also includes the officially removed F2 and assigns rank 8 to its missing AO
and RSA results. Mayfly's final official `CEC2017Suite` therefore has 29 usable
functions and cannot regenerate the paper's 30-function Friedman summary. The
reference artifact also flags the printed F14 and F28 anomalies instead of
silently correcting them.

Tables 14-23 report 25 runs and a static penalty method, but do not report the
AOBLMOA population, iterations, evaluation budget, seeds, penalty definition,
or raw samples. The reference artifact maps the paper acronyms to official
CEC2020 RC problem metadata and dimension-derived budgets without claiming
that AOBLMOA followed them. The article publishes objective summaries only,
omitting the official protocol's constraint violations, feasibility rates, and
violation-count vectors. It therefore cannot support an exact or even
feasibility-aware current-library comparison. The artifact also preserves the
paper's RC21/RC22 description conflicts and its Table 16 sCMAgES maximum rather
than replacing them with values inferred from the official archive.

In particular, GSASMA's publication does not define its cooling-coefficient
sequence, temperature update, initial temperature, or four SMA
crossover/mutation probability bounds. Its data-availability statement offers
supporting data only on request, and an August 2026 public artifact search found
no author code or raw seeded runs. The official OLCE-MA pseudocode figures
specify Chebyshev mutation over all `N`
crossover offspring, invalidating Mayfly's earlier fittest-offspring tie
assumption, but the accessible figures do not expose the exact recurrence and
component mutation equation. Until that stage is corrected from an
authoritative equation or reference implementation, results using Mayfly's
documented defaults must be labeled as current-library baselines, not exact
paper replications.
