# Research References

Academic papers and research behind the Mayfly algorithm variants implemented in this library.

## Original Mayfly Algorithm

**Zervoudakis, K., & Tsafarakis, S. (2020). A mayfly optimization algorithm. Computers & Industrial Engineering, 145, 106559.**

**DOI**: https://doi.org/10.1016/j.cie.2020.106559

### Original Implementation

- **Authors**: K. Zervoudakis (kzervoudakis@isc.tuc.gr), S. Tsafarakis
- **Institution**: School of Production Engineering and Management, Technical University of Crete, Chania, Greece
- **Language**: MATLAB
- **Year**: 2020

The Table 6 IMA exact preset remains blocked on the paper's crossover-rate and
Gaussian-mutation-rate semantics. The
[machine-readable clarification request](reference-data/original-ma-2020-clarification-request.json)
pins both questions and the primary evidence required to resolve them.

### Key Contributions

- Introduced mayfly-inspired swarm intelligence algorithm
- Dual-population structure (males and females with different behaviors)
- Nuptial dance mechanism for intensification
- Genetic operators (crossover and mutation) for offspring generation
- Demonstrated competitive performance on benchmark functions

---

## DESMA - Dynamic Elite Strategy

Qianhang Du and Honghao Zhu, **Dynamic elite strategy mayfly algorithm**,
[PLOS ONE 17(8), 2022](https://doi.org/10.1371/journal.pone.0273155).

### Key Contributions

- Adaptive elite generation around global best
- Dynamic search range adjustment based on improvement
- Addresses local optima trapping and slow convergence
- Ranked first overall in the paper's 28-function CEC2013 comparison
- Adds `k` elite evaluations per iteration; percentage overhead depends on the configuration

### Enhancement Strategy

- Generates elite solutions within adaptive search range
- Enlarges range when improving (exploration)
- Reduces range when stagnating (exploitation)
- A strictly improving elite replaces the current best population member and
  becomes the next iteration's male attractor
- DESMA-specific crossover samples per-coordinate `L` uniformly from `[-1,1]`
- The exact Table 3 preset remains blocked on six author/archival questions
  pinned in the
  [machine-readable clarification request](reference-data/desma-2022-clarification-request.json)

---

## OLCE-MA - Orthogonal Learning and Chaotic Exploitation

**Zhou, D., Kang, Z., Su, X., & Yang, C. (2022). An enhanced Mayfly optimization algorithm based on orthogonal learning and chaotic exploitation strategy. International Journal of Machine Learning and Cybernetics, 13, 3625–3643.**

**DOI**: https://doi.org/10.1007/s13042-022-01617-4

### Key Contributions

- Orthogonal experimental design for systematic parameter exploration
- Chebyshev-map mutation in the published chaotic-exploitation pseudocode
- Improved diversity maintenance
- Better performance on highly multimodal problems
- 15-30% improvement on Rastrigin, Rosenbrock, Ackley

### Technical Details

- **Orthogonal learning**: Applied to the primary male movement operator
- **Chaos exploitation**: The publisher pseudocode applies Chebyshev-based
  mutation to all crossover offspring. Mayfly's current one-offspring
  Logistic-map stage predates access to that figure and remains a documented
  extension until the exact component equation is available.
- **Target problems**: High-dimensional multimodal optimization

---

## EOBBMA - Elite Opposition-Based Bare Bones

**Zhou, G., Zhang, T., & Zhou, Y. (2025). Elite Opposition-Based Bare Bones
Mayfly Algorithm for Optimization Wireless Sensor Networks Coverage Problem.
Arabian Journal for Science and Engineering, 50(2), 719–739.** Published online
25 March 2024. [DOI 10.1007/s13369-024-08899-6](https://doi.org/10.1007/s13369-024-08899-6)

### Key Contributions

- Bare Bones framework: Gaussian sampling instead of velocity
- Lévy flight distribution for heavy-tailed exploration
- Elite opposition-based learning for search coverage
- Evaluated on eight wireless-sensor deployment scales
- Fewer parameters to tune than velocity-based approaches

### Protocol Audit

The [machine-readable source audit](reference-data/eobbma-2025-tables2-8.json)
encodes the version of record's eight deployment scenarios and every published
cell from Tables 5–8. Table 3 prints population size 40 and Figure 4 shows 1,000
iterations. The artifact separately identifies the 2023 SSRN preprint, which
has a different title, author list, and DOI but no formal DOI-registry relation
to the journal article.

This is not an exact reproduction protocol. The public primary-source tables
and figures omit the independent-run count, seeds, objective weights, raw
sensor coordinates, and whether population 40 is total or per sex. The current
library therefore has no paper-exact EOBBMA preset or WSN comparison.

The linked
[machine-readable clarification request](reference-data/eobbma-2025-clarification-request.json)
captures the completed publisher, DOI-registry, SSRN, and ORCID audit. It
separates seven exact-preset blockers from three historical-comparison data
blockers, pins them to stable IDs, and includes a send-ready draft that remains
explicitly unsent.

### Current-Library Technical Details

- **Gaussian sampling**: μ = (X_current + X_best)/2, σ = |X_current - X_best|/2
- **Lévy flights**: Mantegna's algorithm with α=1.5 (stability parameter)
- **Opposition-based learning**: x_opp = a + b - x
- **Application**: Elite solutions (top 3 by default)

### Mathematical Foundation

Lévy flights follow power-law distribution:

- Stability parameter α ∈ (0, 2]: controls tail heaviness
- α=1.5 provides good balance of local/global search
- Heavy tails enable occasional large jumps to escape local optima

---

## GSASMA - Golden Annealing Crossover-Mutation Mayfly Algorithm

**An improved mayfly algorithm and its application (2022). AIP Advances.**

**DOI**: https://doi.org/10.1063/5.0108278

### Key Contributions

- Simulated annealing selects late-stage velocities for both populations
- Golden-sine Eq. (10) updates every male and female position
- SMA crossover/mutation was selected, but its probability bounds are omitted

### Technical Components

1. **Golden Sine Algorithm**:
   - Uses fixed coefficients derived from `(sqrt(5)-1)/2`
   - Applied to both complete populations
   - Has no configurable `GoldenFactor` or recurrent interval narrowing

2. **Simulated Annealing**:
   - Metropolis criterion: P(accept) = exp(-ΔE/T)
   - The paper does not publish the temperature recurrence or defaults; the
     library's cooling schedules are documented extensions

HMMA instead schedules OBL/Cauchy mutation of the global optimum and applies
artificial gender mutation to offspring pairs.

---

## HMMA - Hybrid Mutation Mayfly Algorithm

H. Zhang et al., “Improved mayfly algorithm based on hybrid mutation,”
*Electronics Letters* 58 (2022), 687–689.
[DOI 10.1049/ell2.12568](https://doi.org/10.1049/ell2.12568).

The paper combines scheduled opposition-based/Cauchy mutation of the global
best with Equation (12) artificial mutation of offspring. Its Table 1 aggregates
are preserved in the
[machine-readable reference](reference-data/hmma-2022-table1.json). The exact
preset is blocked by seven protocol questions, while historical seeds and raw
runs are two additional reproduction-comparison limitations. All nine are
pinned in the
[machine-readable clarification request](reference-data/hmma-2022-clarification-request.json).

---

## MPMA - Median Position-Based

Guo, L.; Xu, C.; Yu, T.; Tuerxun, W. “An Improved Mayfly Optimization
Algorithm Based on Median Position and Its Application in the Optimization of
PID Parameters of Hydro-Turbine Governor.” _IEEE Access_ **2022**, 10,
36335–36349. DOI:
[10.1109/ACCESS.2022.3160714](https://doi.org/10.1109/ACCESS.2022.3160714).

The paper calls the method MMA; this library uses MPMA to disambiguate it.

### Key Contributions

- Ranks mayflies by objective value and uses the middle-ranked position vector,
  or the average of the two middle-ranked vectors, in male velocity updates
- Adds the nonlinear gravity schedule
  `g(t) = 0.5*sqrt(1-(t/T)^2)+0.4`
- Evaluates 18 classic functions and a hydro-turbine governor PID model

### Technical Details

- The benchmark protocol uses 20 males, 20 females, 30 runs, and
  function-specific evaluation budgets from 1,000 to 100,000.
- The governor protocol uses 35 runs and 50 iterations under two operating
  conditions.
- The paper does not report `a4`, the median pool, genetic parameters, seeds,
  raw runs, or exact evaluation accounting, so no exact preset is available.
- Linear, exponential, and sigmoid gravity choices and the weighted median are
  library extensions. `MedianWeight = 0.5` and the male-only median pool are
  compatibility choices, not values resolved by the paper.

All 110 tabular MMA output cells and the public protocol are preserved in
[`mpma-2022-tables1-10.json`](reference-data/mpma-2022-tables1-10.json), along
with the printed Foxholes-range and F18 best/median inconsistencies. The
artifact is explicitly a source transcription, not a reproduction.

Five exact-preset questions and three historical-comparison/model questions
are pinned in the
[machine-readable clarification request](reference-data/mpma-2022-clarification-request.json).
Its correspondence draft is marked `not_sent`; no request has been
transmitted.

### Paper Application

- Control system optimization (PID tuning)
- Hydro-turbine governor frequency-disturbance response

---

## AOBLMOA - Aquila Optimizer and Opposition-Based Learning Mayfly Algorithm

Zhao, Y.; Huang, C.; Zhang, M.; Cui, Y. "AOBLMOA: A Hybrid Biomimetic
Optimization Algorithm for Numerical Optimization and Engineering Design
Problems." _Biomimetics_ **2023**, 8(4), 381. DOI:
[10.3390/biomimetics8040381](https://doi.org/10.3390/biomimetics8040381).
Open access (PMC10452254).

### Key Contributions

- Replaces the Mayfly nuptial dance (males) and random flight (females) with
  Aquila Optimizer hunting strategies, keeping the attraction branches intact
- Decides the branch by a deterministic fitness test, Eq. (29) and Eq. (30)
- Fixes the hunting strategy by the individual's sex and the iteration phase,
  rather than flipping a coin as plain AO does
- Replaces Gaussian offspring mutation with stochastic opposition-based
  learning, Eq. (31), and greedy selection, Eq. (32)

### Aquila Optimizer Strategies

1. **X1 - Expanded Exploration**: High soar with vertical stoop
   - Uses population mean for global search
   - Females, exploration phase

2. **X2 - Narrowed Exploration**: Contour flight with short glide
   - Lévy flight for focused exploration
   - Males, exploration phase

3. **X3 - Expanded Exploitation**: Low flight with slow descent
   - Convergence with controlled exploration
   - Females, exploitation phase

4. **X4 - Narrowed Exploitation**: Walk and grab
   - Intensive local search with quality function
   - Males, exploitation phase

The sex-to-strategy mapping above follows the paper's equations. Its abstract
states the opposite assignment; see `aoblmoaStrategyFor`, which carries that
resolution. The paper's Tables 5-6 protocol and 19 AOBLMOA result rows are
transcribed in
[`aoblmoa-2023-tables5-6.json`](reference-data/aoblmoa-2023-tables5-6.json).
Its Tables 7-9 dimension-stability protocol and all 30 AOBLMOA rows for F1-F10
at dimensions 30, 50, and 100 are transcribed separately in
[`aoblmoa-2023-tables7-9.json`](reference-data/aoblmoa-2023-tables7-9.json).
Tables 10-11 add all 49 published Wilcoxon p-value and Friedman-rank rows and
their summary rows in
[`aoblmoa-2023-tables10-11.json`](reference-data/aoblmoa-2023-tables10-11.json).
Table 13's 30 CEC2017 average, standard-deviation, and rank rows for all eight
algorithms, plus its Friedman summaries, are in
[`aoblmoa-2023-table13.json`](reference-data/aoblmoa-2023-table13.json).
Tables 14-23 complete the published-output transcription with all 200
objective-summary values for four algorithms on ten selected CEC2020
constrained problems in
[`aoblmoa-2023-tables14-23.json`](reference-data/aoblmoa-2023-tables14-23.json).
All five artifacts are explicitly non-reproducing because the paper-linked
MATLAB repository omits the 30-run driver, statistics driver, seeds, raw
outputs, and all benchmark evaluators except F1, and conflicts with the article
on the population mean and Levy scale. The dimension artifact also preserves
an apparent source inconsistency in Table 8's F10 standard deviation. The
statistics artifact preserves the paper's unpaired rank-sum test and
non-average tie ranks, which differ from Mayfly's paired signed-rank comparison.
The CEC2017 artifact additionally records that the paper omits the dimension
and evaluation budget, includes the officially removed F2 with missing AO/RSA
values, and contains apparent F14/F28 numerical inconsistencies.
The CEC2020 artifact maps the paper's problem acronyms to the official RC
metadata and budgets, but does not infer AOBLMOA adherence: the article omits
its population, iteration/evaluation budget, penalty definition, seeds, raw
runs, constraint violations, and feasibility outputs. It also preserves the
article's RC21/RC22 description conflicts and Table 16 comparator discrepancy.

The completed publisher/DOI/ORCID/repository audit and ten unresolved evidence
questions are consolidated in a
[machine-readable clarification request](reference-data/aoblmoa-2023-clarification-request.json).
It separates five algorithm-definition questions from five experiment and
historical-comparison questions, and its send-ready author draft remains marked
`not_sent`.

### Multi-Objective Helpers

`Optimize` is single-objective. The package exports a Pareto toolkit for
callers who want a front:

- **Pareto dominance**: Solution A dominates B if no worse in all objectives and better in ≥1
- **Crowding distance**: Measures solution density in objective space
- **NSGA-II selection**: Maintains convergence and diversity
- **Archive management**: Stores non-dominated solutions
- **Performance metrics**: Hypervolume, IGD (Inverted Generational Distance)

---

## Comparative Studies

### Performance Benchmarks

Source papers use different protocols, so fixed percentage improvements are not
directly comparable across variants:

| Variant | Best Problem Type | Improvement     | Overhead            |
| ------- | ----------------- | --------------- | ------------------- |
| DESMA   | CEC2013           | Table 3 rank 1st | `k` evals/iteration |
| OLCE-MA | Highly Multimodal | 15-30%          | Dimension-dependent |
| EOBBMA  | Deceptive         | 55%+            | +1.5% evals         |
| GSASMA  | Multimodal        | Paper-dependent | baseline batches    |
| MPMA    | Classic/PID paper cases | Better on 16/18 paper functions | Exact accounting unpublished |
| AOBLMOA | Complex/Adaptive  | Variable        | +20-30% evals       |

### Common Benchmark Functions

Papers typically evaluate on:

- **CEC competitions**: CEC 2014, CEC 2017, CEC 2020 suites
- **Classic functions**: Sphere, Rastrigin, Rosenbrock, Ackley, Griewank,
  Eggcrate, Beale
- **Deceptive functions**: Schwefel, Michalewicz
- **Engineering problems**: Spring design, welded beam, pressure vessel

---

## Research Trends

### Evolution of Mayfly Algorithm (2020-2024)

1. **2020**: Original MA introduced
2. **2022**: Multiple variants emerge (DESMA, OLCE-MA, GSASMA, MPMA)
3. **2023**: Hybrid approaches (AOBLMOA)
4. **2024**: Advanced variants (EOBBMA with Bare Bones framework)

### Common Enhancement Strategies

1. **Elite strategies**: Generate/maintain high-quality solutions
2. **Chaotic maps**: Improve diversity and local search
3. **Opposition-based learning**: Expand search coverage
4. **Lévy flights**: Heavy-tailed exploration
5. **Hybrid approaches**: Combine multiple metaheuristics
6. **Adaptive mechanisms**: Self-adjust parameters during search

---

## Implementation Notes

This Go implementation targets **research fidelity** while recording known
source ambiguities and compatibility extensions explicitly. In particular,
DESMA's missing Table 3 protocol/data, OLCE chaotic exploitation, and GSASMA's
undocumented schedules are open Phase 7 items. The library also provides:

- Idiomatic Go code structure
- Accessible paper components covered by equation-level fixtures
- Known deviations isolated behind documented compatibility behavior
- Consistent API across variants
- Classic, CEC, and engineering benchmark support
- Statistical comparison and raw-result export

### Validation Approach

Validation currently includes:

- equation-level fixtures for accessible published formulas;
- deterministic lifecycle, accounting, and sequential/parallel parity tests;
- controlled benchmark studies produced by the
  [paper-reproduction harness](paper-reproduction.md).

The harness does not yet reproduce every source paper's complete protocol or
published result table. Its manifests and raw outputs are the basis for that
remaining work; reported improvement percentages above are literature claims,
not new measurements from the corrected v0.7 implementations.

---

## Citations

When using this library in research, please cite:

### Standard MA

```bibtex
@article{zervoudakis2020mayfly,
  title={A mayfly optimization algorithm},
  author={Zervoudakis, Konstantinos and Tsafarakis, Sifis},
  journal={Computers \& Industrial Engineering},
  volume={145},
  pages={106559},
  year={2020},
  publisher={Elsevier}
}
```

### OLCE-MA

```bibtex
@article{zhou2022enhanced,
  title={An enhanced Mayfly optimization algorithm based on orthogonal learning and chaotic exploitation strategy},
  author={Zhou, Donglin and Kang, Zheng and Su, Xiangdong and Yang, Cong},
  journal={International Journal of Machine Learning and Cybernetics},
  volume={13},
  pages={3625--3643},
  year={2022},
  publisher={Springer}
}
```

### This Implementation

```bibtex
@software{mayfly_go,
  title={Mayfly Optimization Algorithm: Go Implementation},
  author={Budde, Christian},
  year={2024},
  url={https://github.com/cwbudde/mayfly}
}
```

---

## Further Reading

### Related Metaheuristics

- **Particle Swarm Optimization (PSO)**: Velocity-based swarm intelligence
- **Genetic Algorithms (GA)**: Evolutionary computation with crossover/mutation
- **Aquila Optimizer (AO)**: Eagle hunting strategies
- **Grey Wolf Optimizer (GWO)**: Wolf pack hierarchy
- **Whale Optimization Algorithm (WOA)**: Bubble-net feeding

### Survey Papers

- Metaheuristic optimization surveys (2020-2024)
- Swarm intelligence comparisons
- Hybrid algorithm design patterns
- No Free Lunch theorem implications

---

## Contact Information

For questions about the implementation or to report issues:

- **GitHub**: https://github.com/cwbudde/mayfly/issues
- **Email**: Contact repository maintainers

For questions about the original algorithms:

- Refer to author contact information in respective papers
- Check paper citations for latest contact details
