# Mayfly Optimization Library Documentation

Welcome to the Mayfly optimization library documentation. This folder contains comprehensive guides for all aspects of the library.

## Quick Links

- **New to Mayfly?** Start with [Getting Started](getting-started.md)
- **Want to understand algorithms?** Check [Algorithm Documentation](#algorithm-documentation)
- **Need an API summary?** Open the [API Quick Reference](api/quick-reference.md)
- **Looking for benchmarks?** Visit [Benchmark Functions](benchmarks.md)
- **Wondering about the initial population?** See [Quasi-Random Initial Populations](qmc-initialization.md)
- **Interested in research?** Read [Research References](research.md)
- **Reproducing experiments?** Use the [Paper-reproduction Experiments](paper-reproduction.md)

## Documentation Structure

### Getting Started

- **[Getting Started Guide](getting-started.md)** - Tutorial with practical examples
  - Installation
  - Basic usage
  - Custom objective functions
  - Real-world examples
  - Common pitfalls

### Algorithm Documentation

Detailed guides for each algorithm variant:

1. **[Standard MA](algorithms/standard-ma.md)** - Original Mayfly Algorithm
   - Best for: General-purpose optimization
   - Baseline performance reference

2. **[DESMA](algorithms/desma.md)** - Dynamic Elite Strategy
   - Best for: Multimodal problems
   - Published CEC2013 Table 3 average rank: 1st of 8 algorithms

3. **[OLCE-MA](algorithms/olce-ma.md)** - Orthogonal Learning & Chaotic Exploitation
   - Best for: Highly multimodal problems
   - +15-30% improvement on Rastrigin-like functions

4. **[EOBBMA](algorithms/eobbma.md)** - Elite Opposition-Based Bare Bones
   - Best for: Deceptive landscapes
   - +55% improvement on Schwefel function

5. **[GSASMA](algorithms/gsasma.md)** - Golden Sine with Simulated Annealing
   - Best for: Fast convergence
   - Annealed velocity and golden-sine updates for both populations

6. **[MPMA](algorithms/mpma.md)** - Median Position-Based
   - Paper scope: 18 classic functions and hydro-turbine PID tuning
   - Paper reports better performance on 16 of 18 functions; raw runs are unavailable

7. **[AOBLMOA](algorithms/aoblmoa.md)** - Aquila Optimizer-Based Learning
   - Best for: Adaptive multi-phase optimization
   - Built-in multi-objective support

### API Documentation

Complete API reference:

- **[API Quick Reference](api/quick-reference.md)** - Task-oriented API map
  - Optimization entry points and factories
  - Run options and result fields
  - Builder, selection, and comparison APIs
  - Configuration files, presets, and utilities

- **[Configuration Guide](api/configuration.md)** - All parameters explained
  - Problem parameters
  - Population parameters
  - Velocity parameters
  - Variant-specific parameters
  - Configuration validation

- **[Unified Framework](api/unified-framework.md)** - Advanced features
  - Variant interface
  - Fluent builder API
  - Algorithm selection
  - Automatic problem classification
  - Configuration presets
  - JSON configuration files

- **[Comparison Framework](api/comparison-framework.md)** - Statistical analysis
  - ComparisonRunner API
  - Statistical tests (Wilcoxon, Friedman)
  - Convergence analysis
  - Result export (CSV, JSON)

- **[Run Lifecycle](api/run-lifecycle.md)** - Runtime control and monitoring
  - Context cancellation and initial populations
  - Progress callbacks and structured logging
  - Convergence curve export (CSV, JSON)

### Reference Documentation

- **[Benchmark Functions](benchmarks.md)** - Test function reference
  - 16+ standard benchmark functions
  - Function characteristics
  - Expected performance
  - Usage examples

- **[Performance and Profiling](performance.md)** - Developer performance guide
  - Reproducible optimizer benchmarks
  - CPU and memory profiling commands
  - Recorded optimization results

- **[Quasi-Random Initial Populations](qmc-initialization.md)** - Config.QMCInit
  - Sobol and Halton initial populations
  - The measurement over the benchmark suite
  - Where it helps and where it says nothing

- **[Research References](research.md)** - Academic papers
  - Original research citations
  - Variant-specific papers
  - BibTeX entries
  - Research trends
- **[Paper-reproduction Experiments](paper-reproduction.md)** - Reproducible post-v0.7 baseline
  - Paired seed schedule for all eight variants
  - Machine-readable protocol, raw run data, and configuration snapshots
- **[Release Guide](releasing.md)** - Version policy, validation, and
  publication checklist

## Navigation Guide

### By User Type

**Beginner Users:**

1. [Getting Started](getting-started.md)
2. [Benchmark Functions](benchmarks.md)
3. [Standard MA](algorithms/standard-ma.md)

**Intermediate Users:**

1. [API Quick Reference](api/quick-reference.md)
2. [Algorithm Documentation](algorithms/)
3. [Configuration Guide](api/configuration.md)

**Advanced Users:**

1. [Comparison Framework](api/comparison-framework.md)
2. [Research References](research.md)
3. All algorithm variants

### By Task

**Optimize a function:**

- [Getting Started](getting-started.md#basic-usage)
- [API Quick Reference](api/quick-reference.md#minimal-optimization)
- [Configuration Guide](api/configuration.md)

**Choose an algorithm:**

- [Algorithm comparison table](../README.md#algorithm-variants)
- [Unified Framework - Algorithm Selection](api/unified-framework.md#algorithm-selection)

**Compare algorithms statistically:**

- [Comparison Framework](api/comparison-framework.md)

**Tune parameters:**

- [Configuration Guide](api/configuration.md)
- Individual algorithm docs (parameter tuning sections)

**Understand research:**

- [Research References](research.md)
- Algorithm-specific papers in each variant doc

## Document Sizes

| Document                | Lines | Focus                 |
| ----------------------- | ----- | --------------------- |
| quick-reference.md      | ~230  | Compact API map       |
| getting-started.md      | ~440  | Tutorial and examples |
| benchmarks.md           | ~410  | Function reference    |
| research.md             | ~350  | Academic citations    |
| configuration.md        | ~470  | Parameter reference   |
| unified-framework.md    | ~570  | Advanced API          |
| comparison-framework.md | ~180  | Statistical testing   |
| standard-ma.md          | ~170  | Algorithm guide       |
| desma.md                | ~220  | Algorithm guide       |
| olce-ma.md              | ~360  | Algorithm guide       |
| eobbma.md               | ~480  | Algorithm guide       |
| gsasma.md               | ~430  | Algorithm guide       |
| mpma.md                 | ~440  | Algorithm guide       |
| aoblmoa.md              | ~490  | Algorithm guide       |

**Total:** more than 5,000 lines across these reference documents.

## Contributing to Documentation

When updating documentation:

1. Keep examples concise and runnable
2. Include line references to source code where appropriate
3. Update this README.md if adding new documents
4. Cross-reference related documentation
5. Test all code examples

## External Resources

- **Main README**: [../README.md](../README.md)
- **Development Guide**: [../CLAUDE.md](../CLAUDE.md)
- **Roadmap**: [../PLAN.md](../PLAN.md)
- **Examples**: [../examples/](../examples/)
- **Source Code**: Root directory

---

**Questions?** Check [GitHub Issues](https://github.com/cwbudde/mayfly/issues)
