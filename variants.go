package mayfly

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// AlgorithmVariant represents a specific variant of the Mayfly Algorithm.
// This interface provides a unified way to work with all algorithm variants.
type AlgorithmVariant interface {
	// Name returns the short name of the variant (e.g., "MA", "DESMA", "OLCE-MA")
	Name() string

	// FullName returns the full descriptive name of the variant
	FullName() string

	// Description returns a brief description of the variant's key features
	Description() string

	// GetConfig returns a default configuration for this variant.
	// You must still set ObjectiveFunc, ProblemSize, LowerBound, and UpperBound.
	GetConfig() *Config

	// ApplicableTo returns a score (0-1) indicating how well this variant
	// suits the given problem characteristics. Higher scores indicate better fit.
	ApplicableTo(characteristics ProblemCharacteristics) float64

	// EstimatedOverhead returns the approximate overhead (as a multiplier)
	// compared to standard MA. For example, 1.15 means ~15% more function evaluations.
	EstimatedOverhead() float64

	// RecommendedFor returns a slice of problem types this variant excels at
	RecommendedFor() []string
}

// ProblemCharacteristics describes the properties of an optimization problem.
type ProblemCharacteristics struct {
	// Dimensionality indicates problem size
	Dimensionality int // Number of decision variables

	// Modality describes the landscape
	Modality Modality // Unimodal, Multimodal, HighlyMultimodal

	// Landscape describes the terrain
	Landscape Landscape // Smooth, Rugged, Deceptive, NarrowValley

	// ExpensiveEvaluations indicates if function evaluations are costly
	ExpensiveEvaluations bool

	// RequiresFastConvergence indicates if quick results are needed
	RequiresFastConvergence bool

	// RequiresStableConvergence indicates if low variance is important
	RequiresStableConvergence bool

	// MultiObjective indicates if there are multiple objectives
	MultiObjective bool
}

// Modality describes the number of optima in the problem.
type Modality int

const (
	Unimodal         Modality = iota // Single optimum
	Multimodal                       // Several optima
	HighlyMultimodal                 // Many optima (10+)
)

// Landscape describes the problem terrain.
type Landscape int

const (
	Smooth       Landscape = iota // Few local features
	Rugged                        // Many local features
	Deceptive                     // Misleading gradients
	NarrowValley                  // Ill-conditioned
)

// Canonical variant names, as returned by AlgorithmVariant.Name().
const (
	nameStandardMA = "MA"
	nameDESMA      = "DESMA"
	nameOLCEMA     = "OLCE-MA"
	nameEOBBMA     = "EOBBMA"
	nameGSASMA     = "GSASMA"
	nameHMMA       = "HMMA"
	nameMPMA       = "MPMA"
	nameAOBLMOA    = "AOBLMOA"
)

// variantRegistry holds all available algorithm variants.
var variantRegistry = map[string]AlgorithmVariant{
	"ma":      &StandardMAVariant{},
	"desma":   &DESMAVariant{},
	"olce":    &OLCEVariant{},
	"olce-ma": &OLCEVariant{}, // alias
	"eobbma":  &EOBBMAVariant{},
	"gsasma":  &GSASMAVariant{},
	"hmma":    &HMMAVariant{},
	"mpma":    &MPMAVariant{},
	"aoblmoa": &AOBLMOAVariant{},
}

// NewVariant creates an algorithm variant by name.
// Returns nil if the variant name is not recognized.
//
// Available variants:
//   - "ma" or "standard" - Standard Mayfly Algorithm
//   - "desma" - Dynamic Elite Strategy MA
//   - "olce" or "olce-ma" - Orthogonal Learning and Chaotic Exploitation MA
//   - "eobbma" - Elite Opposition-Based Bare Bones MA
//   - "gsasma" - Golden Sine with Simulated Annealing MA
//   - "hmma" - Hybrid Mutation Mayfly Algorithm
//   - "mpma" - Median Position-Based MA
//   - "aoblmoa" - Aquila Optimizer and Opposition-Based Learning Mayfly Optimization Algorithm
//
// Deprecated: use NewVariantChecked when the name is not a compile-time
// constant; this compatibility lookup returns nil for unknown names.
func NewVariant(name string) AlgorithmVariant {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "standard" {
		name = "ma"
	}

	return variantRegistry[name]
}

// NewVariantChecked resolves a variant name or returns an error listing the
// invalid lookup. It is preferred for user-controlled names.
func NewVariantChecked(name string) (AlgorithmVariant, error) {
	variant := NewVariant(name)
	if variant == nil {
		return nil, fmt.Errorf("unknown algorithm variant %q", name)
	}

	return variant, nil
}

// ListVariants returns a list of all available algorithm variant names.
func ListVariants() []string {
	variants := make([]string, 0, len(variantRegistry))
	seen := make(map[string]bool)

	for name := range variantRegistry {
		// Skip aliases (only include primary names)
		if name == "olce-ma" || name == "standard" {
			continue
		}

		if !seen[name] {
			variants = append(variants, name)
			seen[name] = true
		}
	}

	sort.Strings(variants)

	return variants
}

// GetAllVariants returns all available algorithm variants.
func GetAllVariants() []AlgorithmVariant {
	return []AlgorithmVariant{
		variantRegistry["ma"],
		variantRegistry["desma"],
		variantRegistry["olce"],
		variantRegistry["eobbma"],
		variantRegistry["gsasma"],
		variantRegistry["hmma"],
		variantRegistry["mpma"],
		variantRegistry["aoblmoa"],
	}
}

// =============================================================================
// Standard MA Variant
// =============================================================================

// StandardMAVariant represents the standard Mayfly Algorithm.
type StandardMAVariant struct{}

func (v *StandardMAVariant) Name() string {
	return nameStandardMA
}

func (v *StandardMAVariant) FullName() string {
	return "Standard Mayfly Algorithm"
}

func (v *StandardMAVariant) Description() string {
	return "Original Mayfly Algorithm with balanced exploration-exploitation. Good baseline for most problems."
}

func (v *StandardMAVariant) GetConfig() *Config {
	return NewDefaultConfig()
}

func (v *StandardMAVariant) ApplicableTo(characteristics ProblemCharacteristics) float64 {
	// Multi-objective problems require specialized algorithms
	if characteristics.MultiObjective {
		return 0.2 // Not suitable for multi-objective
	}

	score := 0.5 // Baseline applicability

	// Performs well on general problems
	if characteristics.Modality == Unimodal {
		score += 0.2
	}

	if characteristics.Landscape == Smooth {
		score += 0.2
	}

	if characteristics.Dimensionality <= 50 {
		score += 0.1
	}

	return min(score, 1.0)
}

func (v *StandardMAVariant) EstimatedOverhead() float64 {
	return 1.0 // Baseline
}

func (v *StandardMAVariant) RecommendedFor() []string {
	return []string{
		"General optimization problems",
		"Unimodal functions",
		"Smooth landscapes",
		"Baseline comparison",
	}
}

// =============================================================================
// DESMA Variant
// =============================================================================

// DESMAVariant represents the Dynamic Elite Strategy Mayfly Algorithm.
type DESMAVariant struct{}

func (v *DESMAVariant) Name() string {
	return nameDESMA
}

func (v *DESMAVariant) FullName() string {
	return "Dynamic Elite Strategy Mayfly Algorithm"
}

func (v *DESMAVariant) Description() string {
	return "Enhanced with dynamic elite generation around the current global best " +
		"and an adaptive search radius."
}

func (v *DESMAVariant) GetConfig() *Config {
	return NewDESMAConfig()
}

func (v *DESMAVariant) ApplicableTo(characteristics ProblemCharacteristics) float64 {
	// Multi-objective problems require specialized algorithms
	if characteristics.MultiObjective {
		return 0.2 // Not suitable for multi-objective
	}

	score := 0.5

	// Excels on multimodal problems
	if characteristics.Modality == Multimodal || characteristics.Modality == HighlyMultimodal {
		score += 0.3
	}

	if characteristics.Landscape == Rugged {
		score += 0.2
	}

	if !characteristics.ExpensiveEvaluations {
		score += 0.1 // Overhead is acceptable
	}

	return min(score, 1.0)
}

func (v *DESMAVariant) EstimatedOverhead() float64 {
	// A fixed selector hint only. Actual evaluation overhead depends on
	// EliteCount and on the configured population and genetic operators.
	return 1.08
}

func (v *DESMAVariant) RecommendedFor() []string {
	return []string{
		"Multimodal problems",
		"Local optima escape",
		"Rastrigin, Rosenbrock functions",
		"Problems with many basins of attraction",
	}
}

// =============================================================================
// OLCE-MA Variant
// =============================================================================

// OLCEVariant represents the Orthogonal Learning and Chaotic Exploitation MA.
type OLCEVariant struct{}

func (v *OLCEVariant) Name() string {
	return nameOLCEMA
}

func (v *OLCEVariant) FullName() string {
	return "Orthogonal Learning and Chaotic Exploitation Mayfly Algorithm"
}

func (v *OLCEVariant) Description() string {
	return "Orthogonal experimental design + chaotic perturbations. 15-30% improvement on highly multimodal problems."
}

func (v *OLCEVariant) GetConfig() *Config {
	return NewOLCEConfig()
}

func (v *OLCEVariant) ApplicableTo(characteristics ProblemCharacteristics) float64 {
	// Multi-objective problems require specialized algorithms
	if characteristics.MultiObjective {
		return 0.2 // Not suitable for multi-objective
	}

	score := 0.5

	// Best for highly multimodal problems
	switch characteristics.Modality {
	case HighlyMultimodal:
		score += 0.4
	case Multimodal:
		score += 0.2
	}

	if characteristics.Dimensionality >= 10 {
		score += 0.2 // Benefits from diversity
	}

	if characteristics.Landscape == Rugged {
		score += 0.1
	}

	return min(score, 1.0)
}

func (v *OLCEVariant) EstimatedOverhead() float64 {
	return 1.33 // Default orthogonal-learning evaluations add about 33%.
}

func (v *OLCEVariant) RecommendedFor() []string {
	return []string{
		"Highly multimodal problems",
		"High-dimensional problems (10D+)",
		"Rastrigin, Griewank, Schwefel",
		"Problems requiring diversity",
	}
}

// =============================================================================
// EOBBMA Variant
// =============================================================================

// EOBBMAVariant represents the Elite Opposition-Based Bare Bones MA.
type EOBBMAVariant struct{}

func (v *EOBBMAVariant) Name() string {
	return nameEOBBMA
}

func (v *EOBBMAVariant) FullName() string {
	return "Elite Opposition-Based Bare Bones Mayfly Algorithm"
}

func (v *EOBBMAVariant) Description() string {
	return "Gaussian sampling + Lévy flight + opposition learning. 55%+ improvement on deceptive functions."
}

func (v *EOBBMAVariant) GetConfig() *Config {
	return NewEOBBMAConfig()
}

func (v *EOBBMAVariant) ApplicableTo(characteristics ProblemCharacteristics) float64 {
	// Multi-objective problems require specialized algorithms
	if characteristics.MultiObjective {
		return 0.2 // Not suitable for multi-objective
	}

	score := 0.5

	// Excellent for deceptive landscapes
	if characteristics.Landscape == Deceptive {
		score += 0.4
	}

	if characteristics.Modality == HighlyMultimodal {
		score += 0.2
	}

	if characteristics.ExpensiveEvaluations {
		score += 0.1 // Low overhead
	}

	return min(score, 1.0)
}

func (v *EOBBMAVariant) EstimatedOverhead() float64 {
	return 1.015 // ~1.5% more evaluations
}

func (v *EOBBMAVariant) RecommendedFor() []string {
	return []string{
		"Deceptive functions (Schwefel, Michalewicz)",
		"Complex landscapes",
		"Problems plateauing with other algorithms",
		"Fewer parameters to tune",
	}
}

// =============================================================================
// GSASMA Variant
// =============================================================================

// GSASMAVariant represents the Golden Annealing Crossover-Mutation MA.
type GSASMAVariant struct{}

func (v *GSASMAVariant) Name() string {
	return nameGSASMA
}

func (v *GSASMAVariant) FullName() string {
	return "Golden Annealing Crossover-Mutation Mayfly Algorithm"
}

func (v *GSASMAVariant) Description() string {
	return "Annealed late-stage velocity selection and golden-sine position updates for both mayfly populations."
}

// HMMAVariant represents the Hybrid Mutation Mayfly Algorithm.
type HMMAVariant struct{}

func (v *HMMAVariant) Name() string { return nameHMMA }

func (v *HMMAVariant) FullName() string { return "Hybrid Mutation Mayfly Algorithm" }

func (v *HMMAVariant) Description() string {
	return "Scheduled OBL/Cauchy mutation of the global optimum with artificial gender mutation of offspring."
}

func (v *HMMAVariant) GetConfig() *Config { return NewHMMAConfig() }

func (v *HMMAVariant) ApplicableTo(characteristics ProblemCharacteristics) float64 {
	if characteristics.MultiObjective {
		return 0.2
	}

	score := 0.5
	if characteristics.Modality != Unimodal {
		score += 0.2
	}

	if characteristics.Landscape == Rugged || characteristics.Landscape == Deceptive {
		score += 0.2
	}

	return min(score, 1)
}

func (v *HMMAVariant) EstimatedOverhead() float64 { return 1.02 }

func (v *HMMAVariant) RecommendedFor() []string {
	return []string{"Rugged scalar-objective landscapes", "Stagnation escape", "Multimodal problems"}
}

func (v *GSASMAVariant) GetConfig() *Config {
	return NewGSASMAConfig()
}

func (v *GSASMAVariant) ApplicableTo(characteristics ProblemCharacteristics) float64 {
	// Multi-objective problems require specialized algorithms
	if characteristics.MultiObjective {
		return 0.2 // Not suitable for multi-objective
	}

	score := 0.5

	// Best for fast convergence needs
	if characteristics.RequiresFastConvergence {
		score += 0.3
	}

	if characteristics.Modality == Multimodal {
		score += 0.2
	}

	if !characteristics.ExpensiveEvaluations {
		score += 0.1 // Moderate overhead acceptable
	}

	return min(score, 1.0)
}

func (v *GSASMAVariant) EstimatedOverhead() float64 {
	return 1.07 // Default golden/annealing stages add about 7%.
}

func (v *GSASMAVariant) RecommendedFor() []string {
	return []string{
		"Engineering optimization",
		"Fast convergence requirements",
		"PID tuning, hyperparameter optimization",
		"Problems with time/budget constraints",
	}
}

// =============================================================================
// MPMA Variant
// =============================================================================

// MPMAVariant represents the Median Position-Based MA.
type MPMAVariant struct{}

func (v *MPMAVariant) Name() string {
	return nameMPMA
}

func (v *MPMAVariant) FullName() string {
	return "Median Position-Based Mayfly Algorithm"
}

func (v *MPMAVariant) Description() string {
	return "Median guidance + non-linear gravity. 10-30% improvement with stable, robust convergence."
}

func (v *MPMAVariant) GetConfig() *Config {
	return NewMPMAConfig()
}

func (v *MPMAVariant) ApplicableTo(characteristics ProblemCharacteristics) float64 {
	// Multi-objective problems require specialized algorithms
	if characteristics.MultiObjective {
		return 0.2 // Not suitable for multi-objective
	}

	score := 0.5

	// Excellent for stable convergence needs
	if characteristics.RequiresStableConvergence {
		score += 0.3
	}

	if characteristics.Landscape == NarrowValley {
		score += 0.3
	}

	if characteristics.ExpensiveEvaluations {
		score += 0.1 // No overhead
	}

	return min(score, 1.0)
}

func (v *MPMAVariant) EstimatedOverhead() float64 {
	return 1.0 // No additional overhead
}

func (v *MPMAVariant) RecommendedFor() []string {
	return []string{
		"Control system optimization",
		"Ill-conditioned problems",
		"Rosenbrock, BentCigar, Discus",
		"Stable, predictable convergence",
	}
}

// =============================================================================
// AOBLMOA Variant
// =============================================================================

// AOBLMOAVariant represents the Aquila Optimizer and Opposition-Based Learning
// Mayfly Optimization Algorithm. It is a scalar-objective optimizer.
type AOBLMOAVariant struct{}

func (v *AOBLMOAVariant) Name() string {
	return nameAOBLMOA
}

func (v *AOBLMOAVariant) FullName() string {
	return "Aquila Optimizer and Opposition-Based Learning Mayfly Optimization Algorithm"
}

func (v *AOBLMOAVariant) Description() string {
	return "Mayfly with the dance and flight branches replaced by Aquila hunting " +
		"strategies, and offspring mutation replaced by stochastic opposition-based learning."
}

func (v *AOBLMOAVariant) GetConfig() *Config {
	return NewAOBLMOAConfig()
}

func (v *AOBLMOAVariant) ApplicableTo(characteristics ProblemCharacteristics) float64 {
	score := 0.5

	// This library does not currently expose a multi-objective optimizer.
	if characteristics.MultiObjective {
		return 0.2
	}

	if characteristics.Modality == HighlyMultimodal {
		score += 0.2
	}

	if characteristics.Landscape == Rugged || characteristics.Landscape == Deceptive {
		score += 0.1
	}

	return min(score, 1.0)
}

func (v *AOBLMOAVariant) EstimatedOverhead() float64 {
	return 1.25 // ~25% more evaluations
}

func (v *AOBLMOAVariant) RecommendedFor() []string {
	return []string{
		"Adaptive strategy switching",
		"Complex multi-phase problems",
		"Rugged scalar-objective landscapes",
	}
}

// =============================================================================
// Fluent Builder API
// =============================================================================

// VariantBuilder provides a fluent API for configuring algorithm variants.
type VariantBuilder struct {
	variant AlgorithmVariant
	config  *Config
}

// NewBuilder creates a new builder for the specified variant.
// Returns nil if the variant name is not recognized.
//
// Example: NewBuilder("desma").ForProblem(fn, 10, -5, 5).WithIterations(500).Build().
//
// Deprecated: use NewBuilderChecked to receive an explicit lookup or config
// error.
func NewBuilder(variantName string) *VariantBuilder {
	variant := NewVariant(variantName)
	if variant == nil {
		return nil
	}

	return &VariantBuilder{
		variant: variant,
		config:  variant.GetConfig(),
	}
}

// NewBuilderChecked creates a builder or reports an invalid variant/default
// configuration explicitly.
func NewBuilderChecked(variantName string) (*VariantBuilder, error) {
	variant, err := NewVariantChecked(variantName)
	if err != nil {
		return nil, err
	}

	return NewBuilderFromVariantChecked(variant)
}

// NewBuilderFromVariant creates a builder from an existing variant instance.
// Deprecated: use NewBuilderFromVariantChecked.
func NewBuilderFromVariant(variant AlgorithmVariant) *VariantBuilder {
	if variant == nil {
		return nil
	}

	config := variant.GetConfig()
	if config == nil {
		return nil
	}

	return &VariantBuilder{
		variant: variant,
		config:  cloneComparisonConfig(config),
	}
}

// NewBuilderFromVariantChecked creates a builder from an implementation and
// validates that it provides a non-nil default configuration.
func NewBuilderFromVariantChecked(variant AlgorithmVariant) (*VariantBuilder, error) {
	if variant == nil {
		return nil, errors.New("algorithm variant is nil")
	}

	config := variant.GetConfig()
	if config == nil {
		return nil, fmt.Errorf("variant %q returned a nil config", variant.Name())
	}

	return &VariantBuilder{variant: variant, config: cloneComparisonConfig(config)}, nil
}

// ForProblem sets the objective function and problem parameters.
func (b *VariantBuilder) ForProblem(fn ObjectiveFunction, size int, lower, upper float64) *VariantBuilder {
	if b == nil {
		return nil
	}

	b.config.ObjectiveFunc = fn
	b.config.ProblemSize = size
	b.config.LowerBound = lower
	b.config.UpperBound = upper

	return b
}

// WithIterations sets the maximum number of iterations.
func (b *VariantBuilder) WithIterations(iterations int) *VariantBuilder {
	if b == nil {
		return nil
	}

	b.config.MaxIterations = iterations

	return b
}

// WithPopulation sets the population sizes for males and females.
func (b *VariantBuilder) WithPopulation(males, females int) *VariantBuilder {
	if b == nil {
		return nil
	}

	b.config.NPop = males
	b.config.NPopF = females

	return b
}

// WithQMCInitialPopulation seeds the initial population from a quasi-random
// low-discrepancy sequence instead of independent uniform draws.
//
// Sequence is QMCInitSobol, QMCInitHalton, or QMCInitUniform for the default.
// An unknown name is not rejected here, because a fluent setter has nowhere to
// put an error. It is caught later: Optimize and ValidateConfig both run
// validateQMCInit, which is where this repository validates a whole config in
// one place. Build checks only that the builder itself is usable.
//
// The scramble seed is drawn from the run's RNG, so repeated runs start from
// different low-discrepancy point sets. Set Config.QMCSeed through WithConfig
// to pin it.
func (b *VariantBuilder) WithQMCInitialPopulation(sequence string) *VariantBuilder {
	if b == nil {
		return nil
	}

	b.config.QMCInit = sequence

	return b
}

// WithConfig applies a custom configuration function to the builder.
//
// Example: WithConfig(func(c *Config) { c.A1 = 2.0; c.Beta = 3.0 }).
func (b *VariantBuilder) WithConfig(fn func(*Config)) *VariantBuilder {
	if b == nil || fn == nil {
		return nil
	}

	fn(b.config)

	return b
}

// WithConfigChecked applies a customizer or returns a nil-input error.
func (b *VariantBuilder) WithConfigChecked(fn func(*Config)) (*VariantBuilder, error) {
	if b == nil {
		return nil, errors.New("builder is nil")
	}

	if fn == nil {
		return nil, errors.New("config customizer is nil")
	}

	fn(b.config)

	return b, nil
}

// Build returns the configured Config ready for optimization.
func (b *VariantBuilder) Build() (*Config, error) {
	if b == nil {
		return nil, errors.New("builder is nil (unknown variant?)")
	}

	if b.config.ObjectiveFunc == nil {
		return nil, errors.New("objective function not set")
	}

	if b.config.ProblemSize <= 0 {
		return nil, errors.New("problem size must be positive")
	}

	err := ValidateConfig(b.config)
	if err != nil {
		return nil, err
	}

	return cloneComparisonConfig(b.config), nil
}

// Optimize is a convenience method that builds the config and runs optimization.
func (b *VariantBuilder) Optimize() (*Result, error) {
	config, err := b.Build()
	if err != nil {
		return nil, err
	}

	return Optimize(config)
}

// GetVariant returns the underlying variant.
func (b *VariantBuilder) GetVariant() AlgorithmVariant {
	if b == nil {
		return nil
	}

	return b.variant
}
