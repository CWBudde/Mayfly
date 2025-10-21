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

### Key Contributions

- Introduced mayfly-inspired swarm intelligence algorithm
- Dual-population structure (males and females with different behaviors)
- Nuptial dance mechanism for intensification
- Genetic operators (crossover and mutation) for offspring generation
- Demonstrated competitive performance on benchmark functions

---

## DESMA - Dynamic Elite Strategy

**Dynamic elite strategy mayfly algorithm. PLOS One, 2022.**

### Key Contributions

- Adaptive elite generation around global best
- Dynamic search range adjustment based on improvement
- Addresses local optima trapping and slow convergence
- 70%+ improvement on multimodal functions
- Minimal overhead (~8% more function evaluations)

### Enhancement Strategy

- Generates elite solutions within adaptive search range
- Enlarges range when improving (exploration)
- Reduces range when stagnating (exploitation)
- Replaces worst population members with better elites

---

## OLCE-MA - Orthogonal Learning and Chaotic Exploitation

**Zhou, D., Kang, Z., Su, X., & Yang, C. (2022). An enhanced Mayfly optimization algorithm based on orthogonal learning and chaotic exploitation strategy. International Journal of Machine Learning and Cybernetics, 13, 3625–3643.**

**DOI**: https://doi.org/10.1007/s13042-022-01617-4

### Key Contributions

- Orthogonal experimental design for systematic parameter exploration
- Chaotic maps (logistic map) for perturbation
- Improved diversity maintenance
- Better performance on highly multimodal problems
- 15-30% improvement on Rastrigin, Rosenbrock, Ackley

### Technical Details

- **Orthogonal learning**: Applied to top 20% of population
- **Chaos perturbation**: Logistic map with factor 0.1-0.3
- **Target problems**: High-dimensional multimodal optimization

---

## EOBBMA - Elite Opposition-Based Bare Bones

**Elite Opposition-Based Bare Bones Mayfly Algorithm (2024). Arabian Journal for Science and Engineering.**

### Key Contributions

- Bare Bones framework: Gaussian sampling instead of velocity
- Lévy flight distribution for heavy-tailed exploration
- Elite opposition-based learning for search coverage
- 55%+ improvement on deceptive functions (Schwefel)
- Fewer parameters to tune than velocity-based approaches

### Technical Details

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

## GSASMA - Golden Sine with Simulated Annealing

**Improved mayfly algorithm based on hybrid mutation (2022). Electronics Letters / IEEE**

### Key Contributions

- Golden Sine Algorithm integration using golden ratio (φ ≈ 1.618)
- Simulated Annealing for probabilistic local optima escape
- Hybrid Cauchy-Gaussian mutation (adaptive exploration-exploitation)
- Opposition-Based Learning for expanded coverage
- 10-20% improvement on engineering optimization problems

### Technical Components

1. **Golden Sine Algorithm**:
   - Uses sine function with golden ratio scaling
   - Applied to elite males (top 20%)
   - Adaptive step sizing decreases over iterations

2. **Simulated Annealing**:
   - Three cooling schedules: exponential, linear, logarithmic
   - Metropolis criterion: P(accept) = exp(-ΔE/T)
   - Temperature controls exploration-exploitation balance

3. **Hybrid Mutation**:
   - Early iterations: 70% Cauchy (exploration)
   - Late iterations: 70% Gaussian (exploitation)
   - Smooth transition based on progress

4. **Opposition-Based Learning**:
   - Applied to global best every 10 iterations
   - x_opp = lower + upper - x

---

## MPMA - Median Position-Based

**An Improved Mayfly Optimization Algorithm Based on Median Position (2022). IEEE Access**

### Key Contributions

- Median position guidance for robust convergence
- Non-linear gravity coefficients (linear, exponential, sigmoid)
- Weighted median option for elite emphasis
- 10-30% improvement on ill-conditioned problems
- Lower variance across runs (more stable)

### Technical Details

- **Median guidance**: More robust than mean to outliers
- **Gravity types**:
  - Linear: g(t) = 1 - t/T
  - Exponential: g(t) = exp(-4t/T)
  - Sigmoid: g(t) = 1/(1 + exp(10(t/T - 0.5)))
- **Weighted median**: Fitness-weighted for elite emphasis

### Target Applications

- Control system optimization (PID tuning)
- System identification
- Problems requiring stable, predictable convergence
- Ill-conditioned optimization (narrow valleys)

---

## AOBLMOA - Aquila Optimizer-Based Learning

**AOBLMOA: A Hybrid Biomimetic Optimization Algorithm (2023). PubMed / Various journals**

### Key Contributions

- Hybrid algorithm combining Mayfly and Aquila Optimizer
- Four adaptive hunting strategies (X1, X2, X3, X4)
- Built-in multi-objective optimization (Pareto dominance)
- Automatic strategy switching based on iteration progress
- Opposition-Based Learning framework

### Aquila Optimizer Strategies

1. **X1 - Expanded Exploration**: High soar with vertical stoop
   - Uses population mean for global search
   - First 1/3 of iterations

2. **X2 - Narrowed Exploration**: Contour flight with short glide
   - Lévy flight for focused exploration
   - Iterations 1/3 to 2/3

3. **X3 - Expanded Exploitation**: Low flight with slow descent
   - Convergence with controlled exploration
   - Last 1/3 of iterations

4. **X4 - Narrowed Exploitation**: Walk and grab
   - Intensive local search with quality function
   - Final iterations

### Multi-Objective Framework

- **Pareto dominance**: Solution A dominates B if no worse in all objectives and better in ≥1
- **Crowding distance**: Measures solution density in objective space
- **NSGA-II selection**: Maintains convergence and diversity
- **Archive management**: Stores non-dominated solutions
- **Performance metrics**: Hypervolume, IGD (Inverted Generational Distance)

---

## Comparative Studies

### Performance Benchmarks

Research papers report the following improvements over Standard MA:

| Variant | Best Problem Type | Improvement | Overhead |
|---------|------------------|-------------|----------|
| DESMA | Multimodal | 70%+ | +8% evals |
| OLCE-MA | Highly Multimodal | 15-30% | +12% evals |
| EOBBMA | Deceptive | 55%+ | +1.5% evals |
| GSASMA | Engineering | 10-20% | +15% evals |
| MPMA | Ill-conditioned | 10-30% | 0% (baseline) |
| AOBLMOA | Complex/Adaptive | Variable | +20-30% evals |

### Common Benchmark Functions

Papers typically evaluate on:
- **CEC competitions**: CEC 2014, CEC 2017, CEC 2020 suites
- **Classic functions**: Sphere, Rastrigin, Rosenbrock, Ackley, Griewank
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

This Go implementation maintains **research fidelity** while providing:

- Idiomatic Go code structure
- All variants implemented as described in papers
- Consistent API across variants
- Comprehensive benchmark validation
- Performance metrics matching published results

### Validation Approach

Each variant has been validated against:
- Benchmark functions from original papers
- Expected performance ranges
- Algorithm behavior characteristics
- Parameter sensitivity analysis

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
