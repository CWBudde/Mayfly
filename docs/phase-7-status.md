# Phase 7 Reproducibility Status

Phase 7 tracks paper-fidelity work for the eight Mayfly variants and the
completed classifier follow-ups. This page records completed work and current
evidence gates; [`PLAN.md`](../PLAN.md) contains only the remaining tasks, in
dependency order.

## Reproduction policy

- Results produced before the correctness audit are not comparable to results
  from the corrected implementations.
- A published aggregate table is a reference transcription, not a reproduced
  result.
- A paper-exact preset or comparison is added only after its algorithm and
  experiment-protocol gates are supported by primary author or archival
  evidence.
- Sending a clarification request does not resolve an evidence gate.
- Exact trials are run and archived only after the relevant implementation and
  protocol gates close.

## Shared infrastructure completed

- [`cmd/paper-reproduction`](../cmd/paper-reproduction) and
  [`scripts/run-paper-experiments.sh`](../scripts/run-paper-experiments.sh) run
  paired seeded trials for all eight variants. They export raw CSV/JSON and a
  manifest containing the protocol, revision, runtime, bounds, and
  configuration.
- Reference-data tests keep paper transcriptions, blocker IDs, non-reproduction
  labels, and clarification status machine-checkable.
- The [paper-reproduction guide](paper-reproduction.md) documents the harness,
  supported protocols, and the distinction between descriptive comparisons and
  exact reproductions.

## Variant status

| Variant            | Completed Phase 7 work                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            | Evidence still required                                                                                                                                                                                   |
| ------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Original MA (2020) | Audited the 50-run, 95,000-evaluation protocol, Appendix A grid, Table 6, four Mendeley versions, MATLAB Central history, and institutional record. Added a descriptive corrected-MA comparison and a provenance-safe intake auditor for releases 1.0.0–1.0.2. See [Table 6](reference-data/original-ma-2020-table6.json) and the [clarification request](reference-data/original-ma-2020-clarification-request.json).                                                                                                                                                                            | Recover releases 1.0.0–1.0.2 and determine whether they contain the Table 6 driver or resolve the crossover-rate and Gaussian-mutation semantics. Otherwise obtain primary answers to those two blockers. |
| AOBLMOA (2023)     | Audited the paper-linked MATLAB code and transcribed Tables 5–11, 13, and 14–23, including protocols, statistical cells, official CEC metadata mappings, and source inconsistencies. See the [Tables 5–6](reference-data/aoblmoa-2023-tables5-6.json), [Tables 7–9](reference-data/aoblmoa-2023-tables7-9.json), [Tables 10–11](reference-data/aoblmoa-2023-tables10-11.json), [Table 13](reference-data/aoblmoa-2023-table13.json), [Tables 14–23](reference-data/aoblmoa-2023-tables14-23.json), and [clarification request](reference-data/aoblmoa-2023-clarification-request.json) artifacts. | Five exact-preset and five experiment/comparison blockers: table-generating code, complete benchmark/CEC drivers, seeds, raw trials, feasibility/penalty handling, and statistics details.                |
| DESMA (2022)       | Transcribed all Table 3 rows; added official CEC2013 D=30 support, the 300,000-evaluation/51-run runner, and equation-level elite lifecycle and crossover tests. See [Table 3](reference-data/desma-2022-table3.json) and the [clarification request](reference-data/desma-2022-clarification-request.json).                                                                                                                                                                                                                                                                                      | Initial search radius, population split, base-MA settings, evaluation accounting, seeds, and raw per-run results.                                                                                         |
| EOBBMA (2025)      | Resolved the journal/preprint identity and transcribed all 776 cells from Tables 5–8 while preserving source defects and figure-only outputs. See [Tables 2–8](reference-data/eobbma-2025-tables2-8.json) and the [clarification request](reference-data/eobbma-2025-clarification-request.json).                                                                                                                                                                                                                                                                                                 | The complete versioned protocol, reference code, run count, seeds, objective weights, raw trials, sensor coordinates, and population interpretation.                                                      |
| GSASMA (2022)      | Corrected the annealing boundary, documented cooling and ordinary mating as extensions, and audited the article, flowchart, registries, and archives. The [clarification request](reference-data/gsasma-2022-clarification-request.json) is complete, and a non-marking Inbox/Sent/Drafts search on 2026-08-29 found no matching prior thread.                                                                                                                                                                                                                                                    | Initial temperature, temperature recurrence, `tau_i` definition/lifecycle, four SMA probability bounds, fitness orientation, seed schedule, and raw 30-run outputs.                                       |
| HMMA (2022)        | Transcribed all Table 1 aggregates and audited the publisher, DOI, ORCID, and institutional sources. See [Table 1](reference-data/hmma-2022-table1.json) and the [clarification request](reference-data/hmma-2022-clarification-request.json).                                                                                                                                                                                                                                                                                                                                                    | Seven exact-preset blockers, including population/dimension and equation inconsistencies, plus seeds and raw runs for historical comparison.                                                              |
| MPMA (2022)        | Resolved MMA/MPMA naming and transcribed all 110 cells from Tables 5–7 and 9–10, preserving malformed and missing source values. See [Tables 1–10](reference-data/mpma-2022-tables1-10.json) and the [clarification request](reference-data/mpma-2022-clarification-request.json).                                                                                                                                                                                                                                                                                                                | Complete configuration, implementation, evaluation accounting, seeds, raw benchmark/governor trials, and the Simulink model.                                                                              |
| OLCE-MA (2022)     | Audited the publisher figures, indexed prose, and strategy lineage; preserved the all-offspring versus fittest-offspring conflict. The Logistic-map behavior is documented and tested as a compatibility extension. The [clarification request](reference-data/olce-ma-2022-clarification-request.json) is complete, and a non-marking Inbox/Sent/Drafts search on 2026-08-29 found no prior DOI-matched thread.                                                                                                                                                                                  | Author-confirmed Chebyshev recurrence and seed, sequence lifecycle, component equation, and offspring cardinality, preferably with reference code and a deterministic trace.                              |

All clarification artifacts contain stable blocker IDs and prepared
correspondence. They remain explicitly unsent. Dispatch readiness has been
verified only for OLCE-MA and GSASMA; creating either draft still requires
explicit outbound-correspondence authorization.

## Classifier follow-ups completed

- Dragonfly and Mayfly now share `smoothRoughness = 2.2` and
  `multimodalTurningPoints = 5.0`. The aligned selector classifies Schwefel
  correctly across all 40 regression seeds; the old thresholds reached 37/40
  for modality and 25/40 for landscape.
- Griewank keeps the cheap generic box-scale classifier and the
  literature-backed `HighlyMultimodal`/`Rugged` benchmark override. A 200-seed
  sweep from 65 to 1,025 points per line recovered sampled modality only at much
  higher cost and still classified every run as `Smooth`, so increasing the
  generic sampling cost was not justified.
