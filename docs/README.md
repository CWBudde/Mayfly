# Mayfly Optimization Library Documentation

Welcome to the Mayfly optimization library documentation. This folder contains comprehensive guides for all aspects of the library.

## Quick Links

- **New to Mayfly?** Start with [Getting Started](getting-started.md)
- **Want to understand algorithms?** Check [Algorithm Documentation](#algorithm-documentation)
- **Need API reference?** See [API Documentation](#api-documentation)
- **Looking for benchmarks?** Visit [Benchmark Functions](benchmarks.md)
- **Interested in research?** Read [Research References](research.md)

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
   - +70% improvement on multimodal functions

3. **[OLCE-MA](algorithms/olce-ma.md)** - Orthogonal Learning & Chaotic Exploitation
   - Best for: Highly multimodal problems
   - +15-30% improvement on Rastrigin-like functions

4. **[EOBBMA](algorithms/eobbma.md)** - Elite Opposition-Based Bare Bones
   - Best for: Deceptive landscapes
   - +55% improvement on Schwefel function

5. **[GSASMA](algorithms/gsasma.md)** - Golden Sine with Simulated Annealing
   - Best for: Fast convergence
   - +10-20% improvement on engineering problems

6. **[MPMA](algorithms/mpma.md)** - Median Position-Based
   - Best for: Stable convergence
   - +10-30% improvement on ill-conditioned problems

7. **[AOBLMOA](algorithms/aoblmoa.md)** - Aquila Optimizer-Based Learning
   - Best for: Adaptive multi-phase optimization
   - Built-in multi-objective support

### API Documentation

Complete API reference:

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

### Reference Documentation

- **[Benchmark Functions](benchmarks.md)** - Test function reference
  - 15+ standard benchmark functions
  - Function characteristics
  - Expected performance
  - Usage examples

- **[Research References](research.md)** - Academic papers
  - Original research citations
  - Variant-specific papers
  - BibTeX entries
  - Research trends

## Navigation Guide

### By User Type

**Beginner Users:**
1. [Getting Started](getting-started.md)
2. [Benchmark Functions](benchmarks.md)
3. [Standard MA](algorithms/standard-ma.md)

**Intermediate Users:**
1. [Algorithm Documentation](algorithms/)
2. [Configuration Guide](api/configuration.md)
3. [Unified Framework](api/unified-framework.md)

**Advanced Users:**
1. [Comparison Framework](api/comparison-framework.md)
2. [Research References](research.md)
3. All algorithm variants

### By Task

**Optimize a function:**
- [Getting Started](getting-started.md#basic-usage)
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

| Document | Lines | Focus |
|----------|-------|-------|
| getting-started.md | ~340 | Tutorial and examples |
| benchmarks.md | ~350 | Function reference |
| research.md | ~280 | Academic citations |
| configuration.md | ~250 | Parameter reference |
| unified-framework.md | ~230 | Advanced API |
| comparison-framework.md | ~290 | Statistical testing |
| standard-ma.md | ~170 | Algorithm guide |
| desma.md | ~170 | Algorithm guide |
| olce-ma.md | ~180 | Algorithm guide |
| eobbma.md | ~230 | Algorithm guide |
| gsasma.md | ~250 | Algorithm guide |
| mpma.md | ~200 | Algorithm guide |
| aoblmoa.md | ~260 | Algorithm guide |

**Total:** ~3,700 lines of documentation (was 1,314 lines in README.md alone)

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
