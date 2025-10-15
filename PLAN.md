# Mayfly Algorithm Suite - Implementation Plan

## Vision

Transform this library into a comprehensive suite of Mayfly Optimization Algorithm variants, providing researchers and practitioners with state-of-the-art metaheuristic optimization tools.

---

## Phase 1: Testing Infrastructure ✅ PRIORITY

### Objectives

- Establish robust testing framework
- Ensure code quality and correctness
- Enable performance benchmarking

### Tasks

#### 1.1 Unit Tests

- [x] Test helper functions (unifrnd, unifrndVec, randn, maxVec, minVec)
- [x] Test genetic operators (Crossover, Mutate)
- [x] Test mayfly creation and cloning
- [x] Test sorting functions
- [x] Test elite generation mechanism (DESMA)

#### 1.2 Integration Tests

- [x] Test complete optimization runs with known solutions
- [x] Test convergence on simple benchmark functions
- [x] Test boundary constraint handling
- [x] Test configuration validation
- [x] Test both Standard MA and DESMA variants

#### 1.3 Benchmark Tests

- [ ] Create benchmark suite runner
- [ ] Test all provided benchmark functions (Sphere, Rastrigin, Rosenbrock, Ackley, Griewank)
- [ ] Add statistical significance testing (multiple runs)
- [ ] Generate performance reports
- [ ] Add CEC benchmark functions (CEC2017, CEC2020)

#### 1.4 Regression Tests

- [ ] Capture baseline performance metrics
- [ ] Create regression test suite
- [ ] Add CI/CD integration (GitHub Actions)

#### Deliverables
- [x] `mayfly_test.go` - Core algorithm tests (Phase 1.1 - Unit tests)
- [x] `integration_test.go` - Integration tests with Gherkin/Cucumber (Phase 1.2)
- [x] `features/*.feature` - Gherkin feature files (4 features, 20 scenarios, 94 steps)
- [ ] `functions_test.go` - Benchmark function tests
- [ ] `benchmark_test.go` - Performance benchmarks
- [ ] `.github/workflows/test.yml` - CI configuration

---

## Phase 2: Enhanced Mayfly with Orthogonal Learning & Chaotic Exploitation (OLCE-MA)

### Research Reference
*An enhanced Mayfly optimization algorithm based on orthogonal learning and chaotic exploitation strategy* (2022)
- International Journal of Machine Learning and Cybernetics

### Objectives
- Improve diversity of male mayfly operators
- Reduce oscillatory movement
- Enhance search capability through chaos

### Key Innovations
1. **Orthogonal Learning**
   - Apply orthogonal experimental design to male mayfly position updates
   - Increase population diversity
   - Guide steadier convergence

2. **Chaotic Exploitation**
   - Use chaotic maps (logistic map, tent map, etc.) for offspring position generation
   - Improve local search capability
   - Escape stagnation

### Tasks

#### 2.1 Core Implementation
- [ ] Implement orthogonal array generation
- [ ] Create orthogonal learning operator for male mayflies
- [ ] Implement chaotic map functions (logistic, tent, sine, circle)
- [ ] Add chaotic exploitation to offspring generation
- [ ] Create `NewOLCEConfig()` configuration

#### 2.2 Configuration Parameters
- [ ] `UseOLCE` - Enable OLCE variant
- [ ] `OrthogonalFactor` - Orthogonal learning strength
- [ ] `ChaosType` - Type of chaotic map ("logistic", "tent", "sine", "circle")
- [ ] `ChaosFactor` - Chaos influence factor

#### 2.3 Testing & Validation
- [ ] Unit tests for orthogonal learning
- [ ] Unit tests for chaotic maps
- [ ] Integration tests for OLCE-MA
- [ ] Performance comparison with Standard MA and DESMA
- [ ] Benchmark on high-dimensional problems

#### 2.4 Documentation
- [ ] Add OLCE-MA section to README
- [ ] Create usage examples
- [ ] Document parameter tuning guidelines
- [ ] Add research paper citation

#### Deliverables
- `olce.go` - OLCE-MA implementation
- `chaos.go` - Chaotic map functions
- `orthogonal.go` - Orthogonal learning operators
- `examples/olce/main.go` - Usage example

---

## Phase 3: Elite Opposition-Based Bare Bones Mayfly (EOBBMA)

### Research Reference
*Elite Opposition-Based Bare Bones Mayfly Algorithm* (2024)
- Arabian Journal for Science and Engineering

### Objectives
- Introduce Gaussian distribution for exploration
- Add Lévy flight for long-range jumps
- Implement opposition-based learning

### Key Innovations
1. **Gaussian Distribution**
   - Replace velocity-based updates with Gaussian sampling
   - Position: X_new = N(μ, σ²) where μ is between current and best

2. **Lévy Flight**
   - Add heavy-tailed distribution for exploration
   - Enable large jumps to escape local optima

3. **Elite Opposition-Based Learning**
   - Generate opposition points of elite solutions
   - Expand search space coverage

### Tasks

#### 3.1 Core Implementation
- [ ] Implement Bare Bones framework (Gaussian-based updates)
- [ ] Create Lévy flight distribution generator
- [ ] Implement opposition-based learning operator
- [ ] Add elite selection mechanism
- [ ] Create `NewEOBBMAConfig()` configuration

#### 3.2 Mathematical Components
- [ ] Gaussian sampling with adaptive parameters
- [ ] Lévy flight with Mantegna's algorithm
- [ ] Opposition generation for position vectors
- [ ] Dynamic Lévy step size adjustment

#### 3.3 Configuration Parameters
- [ ] `UseEOBBMA` - Enable EOBBMA variant
- [ ] `LevyAlpha` - Lévy stability parameter (default: 1.5)
- [ ] `LevyBeta` - Lévy scale parameter (default: 1.0)
- [ ] `OppositionRate` - Probability of opposition learning (default: 0.3)
- [ ] `EliteOppositionCount` - Number of elite solutions to oppose

#### 3.4 Testing & Validation
- [ ] Test Lévy flight distribution
- [ ] Test Gaussian sampling
- [ ] Test opposition-based learning
- [ ] Integration tests for complete algorithm
- [ ] Performance benchmarks on complex landscapes

#### 3.5 Documentation
- [ ] Add EOBBMA section to README
- [ ] Document Lévy flight and opposition concepts
- [ ] Create visualization examples
- [ ] Add usage examples

#### Deliverables
- `eobbma.go` - EOBBMA implementation
- `levy.go` - Lévy flight generator
- `opposition.go` - Opposition-based learning
- `examples/eobbma/main.go` - Usage example

---

## Phase 4: Golden Annealing Crossover-Mutation MA (GSASMA)

### Research Reference
*Improved mayfly algorithm based on hybrid mutation* (2022)
- Electronics Letters / IEEE

### Objectives
- Improve convergence speed
- Enhance crossover effectiveness
- Add simulated annealing for acceptance

### Key Innovations
1. **Golden Sine Algorithm Integration**
   - Use golden sine for position updates
   - Adaptive search strategy

2. **Simulated Annealing**
   - Accept worse solutions with probability
   - Escape local optima

3. **Hybrid Mutation**
   - Combination of Cauchy and Gaussian mutations
   - Opposition-based learning on global best

### Tasks

#### 4.1 Core Implementation
- [ ] Implement Golden Sine Algorithm operator
- [ ] Create simulated annealing acceptance mechanism
- [ ] Implement Cauchy mutation
- [ ] Add temperature scheduling
- [ ] Create `NewGSASMAConfig()` configuration

#### 4.2 Optimization Components
- [ ] Golden sine position update
- [ ] Adaptive temperature control
- [ ] Hybrid mutation operator (Cauchy + Gaussian)
- [ ] OBL on global best solution

#### 4.3 Configuration Parameters
- [ ] `UseGSASMA` - Enable GSASMA variant
- [ ] `InitialTemperature` - Starting temperature (default: 100)
- [ ] `CoolingRate` - Temperature decay rate (default: 0.95)
- [ ] `CauchyMutationRate` - Cauchy mutation probability (default: 0.3)
- [ ] `GoldenFactor` - Golden sine influence factor

#### 4.4 Testing & Validation
- [ ] Test simulated annealing acceptance
- [ ] Test Cauchy mutation
- [ ] Test golden sine operator
- [ ] Integration tests
- [ ] Performance comparison on engineering problems

#### 4.5 Documentation
- [ ] Add GSASMA section to README
- [ ] Document annealing schedule
- [ ] Create convergence analysis examples
- [ ] Add parameter tuning guide

#### Deliverables
- `gsasma.go` - GSASMA implementation
- `annealing.go` - Simulated annealing framework
- `golden_sine.go` - Golden sine algorithm
- `examples/gsasma/main.go` - Usage example

---

## Phase 5: Median Position-Based Mayfly (MPMA)

### Research Reference
*An Improved Mayfly Optimization Algorithm Based on Median Position* (2022)
- IEEE Access

### Objectives
- Use median position for better population guidance
- Add non-linear gravity coefficient
- Improve convergence stability

### Key Innovations
1. **Median Position Calculation**
   - Calculate median position of population
   - Use in velocity updates instead of mean

2. **Non-linear Gravity Coefficient**
   - Time-varying gravity based on iteration
   - Better exploration-exploitation balance

### Tasks

#### 5.1 Core Implementation
- [ ] Implement median position calculation
- [ ] Create non-linear gravity coefficient function
- [ ] Modify velocity update equations
- [ ] Create `NewMPMAConfig()` configuration

#### 5.2 Mathematical Components
- [ ] Efficient median calculation for large populations
- [ ] Non-linear coefficient functions (exponential, sigmoid, polynomial)
- [ ] Weighted median for elite solutions

#### 5.3 Configuration Parameters
- [ ] `UseMPMA` - Enable MPMA variant
- [ ] `MedianWeight` - Influence of median position (default: 0.5)
- [ ] `GravityType` - Type of gravity coefficient ("linear", "exponential", "sigmoid")
- [ ] `UseWeightedMedian` - Weight elite solutions more

#### 5.4 Testing & Validation
- [ ] Test median calculation accuracy
- [ ] Test gravity coefficient functions
- [ ] Integration tests
- [ ] Benchmark on control system problems

#### 5.5 Documentation
- [ ] Add MPMA section to README
- [ ] Document median vs mean differences
- [ ] Create gravity function visualizations
- [ ] Add application examples (PID tuning)

#### Deliverables
- `mpma.go` - MPMA implementation
- `gravity.go` - Gravity coefficient functions
- `examples/mpma/main.go` - Usage example

---

## Phase 6: Hybrid Multi-Objective Mayfly (AOBLMOA)

### Research Reference
*AOBLMOA: A Hybrid Biomimetic Optimization Algorithm* (2023)
- PubMed / Various journals

### Objectives
- Combine Mayfly with Aquila Optimizer
- Add opposition-based learning
- Support multi-objective optimization

### Key Innovations
1. **Aquila Optimizer Integration**
   - Four hunting strategies from Aquila Optimizer
   - Hybrid position updates

2. **Opposition-Based Learning**
   - Generate opposition population
   - Select better solutions

3. **Multi-Objective Support**
   - Pareto dominance
   - Crowding distance

### Tasks

#### 6.1 Core Implementation
- [ ] Implement Aquila Optimizer components
- [ ] Create hybrid operator switching mechanism
- [ ] Add opposition-based learning framework
- [ ] Implement Pareto dominance checking
- [ ] Calculate crowding distance
- [ ] Create `NewAOBLMOAConfig()` configuration

#### 6.2 Multi-Objective Components
- [ ] Multi-objective interface
- [ ] Non-dominated sorting
- [ ] Archive management for Pareto front
- [ ] Hypervolume indicator calculation
- [ ] IGD (Inverted Generational Distance) metric

#### 6.3 Aquila Strategies
- [ ] Expanded exploration (high soar)
- [ ] Narrowed exploration (low soar)
- [ ] Expanded exploitation (vertical dive)
- [ ] Narrowed exploitation (walk and grab)

#### 6.4 Configuration Parameters
- [ ] `UseAOBLMOA` - Enable AOBLMOA variant
- [ ] `AquilaWeight` - Aquila strategy weight (default: 0.5)
- [ ] `OppositionProbability` - OBL probability (default: 0.3)
- [ ] `ArchiveSize` - Max Pareto archive size
- [ ] `StrategySwitch` - Iteration threshold for strategy switching

#### 6.5 Testing & Validation
- [ ] Test each Aquila strategy
- [ ] Test Pareto dominance logic
- [ ] Test crowding distance calculation
- [ ] Integration tests
- [ ] Multi-objective benchmarks (ZDT, DTLZ)

#### 6.6 Documentation
- [ ] Add AOBLMOA section to README
- [ ] Document multi-objective usage
- [ ] Create Pareto front visualization examples
- [ ] Add engineering design examples

#### Deliverables
- `aoblmoa.go` - AOBLMOA implementation
- `aquila.go` - Aquila Optimizer strategies
- `multiobjective.go` - Multi-objective utilities
- `examples/aoblmoa/main.go` - Usage example

---

## Phase 7: Unified Framework & API

### Objectives
- Create consistent API across all variants
- Enable easy variant comparison
- Provide algorithm selection guidance

### Tasks

#### 7.1 Unified Interface
- [ ] Create `AlgorithmVariant` interface
- [ ] Implement common configuration base
- [ ] Add variant factory pattern
- [ ] Create fluent API for configuration

#### 7.2 Algorithm Selection Helper
- [ ] Implement problem classifier
- [ ] Create algorithm recommendation system
- [ ] Add performance prediction model

#### 7.3 Comparison Framework
- [ ] Multi-algorithm runner
- [ ] Statistical comparison tests (Wilcoxon, Friedman)
- [ ] Automated benchmark suite
- [ ] Performance visualization

#### 7.4 Configuration Management
- [ ] JSON/YAML configuration loading
- [ ] Configuration validation
- [ ] Parameter auto-tuning utilities
- [ ] Preset configurations for common problems

#### Deliverables
- `variants.go` - Variant interface and factory
- `selector.go` - Algorithm selection helper
- `comparison.go` - Comparison framework
- `config_loader.go` - Configuration utilities

---

## Phase 8: Advanced Features

### Objectives
- Add production-ready features
- Enable research applications
- Improve usability

### Tasks

#### 8.1 Parallel Execution
- [ ] Parallel population evaluation
- [ ] Multi-core support with goroutines
- [ ] Distributed optimization (optional)

#### 8.2 Convergence Detection
- [ ] Early stopping criteria
- [ ] Stagnation detection
- [ ] Adaptive iteration limits

#### 8.3 Logging & Monitoring
- [ ] Structured logging interface
- [ ] Progress callbacks
- [ ] Real-time visualization support (WebSocket)
- [ ] Convergence curve export

#### 8.4 Constraint Handling
- [ ] Penalty function methods
- [ ] Feasibility rules
- [ ] ε-constrained method
- [ ] Constraint-handling utilities

#### 8.5 Advanced Benchmarks
- [ ] CEC2017 benchmark suite
- [ ] CEC2020 benchmark suite
- [ ] CEC2022 benchmark suite
- [ ] Real-world engineering problems

#### Deliverables
- `parallel.go` - Parallel execution
- `convergence.go` - Convergence detection
- `logger.go` - Logging framework
- `constraints.go` - Constraint handling
- `benchmarks/` - Extended benchmark suite

---

## Phase 9: Documentation & Examples

### Objectives
- Comprehensive documentation
- Real-world examples
- Academic paper reproduction

### Tasks

#### 9.1 API Documentation
- [ ] Generate godoc for all packages
- [ ] Add code examples to docs
- [ ] Create quick reference guide
- [ ] Document all parameters and defaults

#### 9.2 Tutorials
- [ ] Getting started tutorial
- [ ] Algorithm selection guide
- [ ] Parameter tuning tutorial
- [ ] Multi-objective optimization tutorial
- [ ] Custom objective function guide

#### 9.3 Real-World Examples
- [ ] Engineering design optimization
- [ ] Neural network hyperparameter tuning
- [ ] Resource allocation problems
- [ ] Scheduling problems
- [ ] Feature selection
- [ ] Power system optimization

#### 9.4 Research Reproducibility
- [ ] Reproduce original MA paper results
- [ ] Reproduce DESMA paper results
- [ ] Reproduce OLCE-MA paper results
- [ ] Reproduce EOBBMA paper results
- [ ] Provide experiment scripts

#### 9.5 Performance Guidelines
- [ ] Algorithm comparison charts
- [ ] Problem-to-algorithm mapping
- [ ] Computational complexity analysis
- [ ] Scalability guidelines

#### Deliverables
- `docs/` - Documentation directory
- `examples/tutorials/` - Tutorial examples
- `examples/applications/` - Real-world applications
- `experiments/` - Research reproduction scripts
- `PERFORMANCE.md` - Performance guide

---

## Phase 10: Community & Release

### Objectives
- Prepare for public release
- Enable community contributions
- Establish maintenance process

### Tasks

#### 10.1 Code Quality
- [ ] Code review all implementations
- [ ] Run static analysis tools (golangci-lint)
- [ ] Ensure 80%+ test coverage
- [ ] Performance profiling and optimization

#### 10.2 Release Preparation
- [ ] Semantic versioning setup
- [ ] CHANGELOG.md
- [ ] Release notes template
- [ ] Migration guides

#### 10.3 Community Setup
- [ ] CONTRIBUTING.md
- [ ] CODE_OF_CONDUCT.md
- [ ] Issue templates
- [ ] Pull request templates
- [ ] Discussion forum setup

#### 10.4 CI/CD Pipeline
- [ ] Automated testing
- [ ] Code coverage reporting
- [ ] Benchmark performance tracking
- [ ] Automated releases

#### 10.5 Publication
- [ ] Publish to pkg.go.dev
- [ ] Create GitHub releases
- [ ] Write announcement blog post
- [ ] Submit to Awesome Go list

#### Deliverables
- `.github/` - GitHub community files
- `.github/workflows/` - CI/CD workflows
- `CHANGELOG.md` - Version history
- `CONTRIBUTING.md` - Contribution guide

---

## Testing Strategy (Cross-Phase)

### Test Coverage Goals
- **Unit Tests**: 90%+ coverage
- **Integration Tests**: All algorithm variants
- **Benchmark Tests**: All standard functions + CEC suites
- **Regression Tests**: Performance baselines

### Test Categories

#### Functional Tests
```
mayfly_test.go           - Core algorithm tests
desma_test.go           - DESMA-specific tests
olce_test.go            - OLCE-MA tests
eobbma_test.go          - EOBBMA tests
gsasma_test.go          - GSASMA tests
mpma_test.go            - MPMA tests
aoblmoa_test.go         - AOBLMOA tests
functions_test.go       - Benchmark function tests
operators_test.go       - Genetic operator tests
helpers_test.go         - Helper function tests
```

#### Performance Tests
```
benchmark_test.go       - Go benchmarks
convergence_test.go     - Convergence analysis
comparison_test.go      - Algorithm comparison
scaling_test.go         - Scalability tests
```

#### Integration Tests
```
integration_test.go     - End-to-end tests
config_test.go         - Configuration validation
boundary_test.go       - Boundary handling
constraints_test.go    - Constraint handling
```

### Test Data
- [ ] Create test fixtures directory
- [ ] Add known optimal solutions
- [ ] Generate synthetic test problems
- [ ] Add edge case datasets

### Continuous Testing
- [ ] Run tests on every commit
- [ ] Nightly performance benchmarks
- [ ] Weekly comprehensive suite
- [ ] Pre-release validation

---

## Success Metrics

### Phase Completion Criteria
Each phase is considered complete when:
1. All tasks checked off
2. All tests passing (coverage > 80%)
3. Documentation complete
4. Code reviewed
5. Examples working

### Overall Project Success
- [ ] 7+ algorithm variants implemented
- [ ] 90%+ test coverage
- [ ] Comprehensive documentation
- [ ] 10+ real-world examples
- [ ] Performance competitive with reference implementations
- [ ] Published and documented

---

## Timeline Estimate

| Phase | Estimated Duration | Priority |
|-------|-------------------|----------|
| Phase 1: Testing | 1-2 weeks | HIGH |
| Phase 2: OLCE-MA | 2-3 weeks | MEDIUM |
| Phase 3: EOBBMA | 2-3 weeks | MEDIUM |
| Phase 4: GSASMA | 2 weeks | MEDIUM |
| Phase 5: MPMA | 1-2 weeks | LOW |
| Phase 6: AOBLMOA | 3-4 weeks | HIGH (multi-obj) |
| Phase 7: Unified Framework | 2 weeks | HIGH |
| Phase 8: Advanced Features | 2-3 weeks | MEDIUM |
| Phase 9: Documentation | 2 weeks | HIGH |
| Phase 10: Release | 1 week | HIGH |

**Total Estimated Time**: 18-25 weeks (4-6 months)

---

## References

1. Zervoudakis, K., & Tsafarakis, S. (2020). A mayfly optimization algorithm. *Computers & Industrial Engineering*, 145, 106559.

2. Dynamic elite strategy mayfly algorithm. *PLOS One*, 2022.

3. An enhanced Mayfly optimization algorithm based on orthogonal learning and chaotic exploitation strategy. *International Journal of Machine Learning and Cybernetics*, 2022.

4. Elite Opposition-Based Bare Bones Mayfly Algorithm. *Arabian Journal for Science and Engineering*, 2024.

5. Improved mayfly algorithm based on hybrid mutation. *Electronics Letters*, 2022.

6. An Improved Mayfly Optimization Algorithm Based on Median Position. *IEEE Access*, 2022.

7. AOBLMOA: A Hybrid Biomimetic Optimization Algorithm. *PubMed*, 2023.

---

## Notes

- **Modularity**: Each variant should be in its own file with clear interfaces
- **Backward Compatibility**: Don't break existing Standard MA and DESMA APIs
- **Performance**: All variants should have similar or better performance than reference implementations
- **Documentation**: Every new feature must be documented before merge
- **Testing**: Write tests before implementation (TDD where practical)
- **Code Style**: Follow Go conventions and pass golangci-lint

---

**Last Updated**: 2025-10-15
**Version**: 1.1
**Status**: Phase 1.2 Complete (Integration Tests) ✅ | Phase 1.3 Next (Benchmark Tests)
