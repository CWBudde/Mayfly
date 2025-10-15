package mayfly

import (
	"fmt"
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

// Modality describes the number of optima in the problem
type Modality int

const (
	Unimodal          Modality = iota // Single optimum
	Multimodal                         // Several optima
	HighlyMultimodal                   // Many optima (10+)
)

// Landscape describes the problem terrain
type Landscape int

const (
	Smooth       Landscape = iota // Few local features
	Rugged                        // Many local features
	Deceptive                     // Misleading gradients
	NarrowValley                  // Ill-conditioned
)

// variantRegistry holds all available algorithm variants
var variantRegistry = map[string]AlgorithmVariant{
	"ma":       &StandardMAVariant{},
	"desma":    &DESMAVariant{},
	"olce":     &OLCEVariant{},
	"olce-ma":  &OLCEVariant{}, // alias
	"eobbma":   &EOBBMAVariant{},
	"gsasma":   &GSASMAVariant{},
	"mpma":     &MPMAVariant{},
	"aoblmoa":  &AOBLMOAVariant{},
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
//   - "mpma" - Median Position-Based MA
//   - "aoblmoa" - Aquila Optimizer-Based Learning Multi-Objective Algorithm
func NewVariant(name string) AlgorithmVariant {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "standard" {
		name = "ma"
	}
	return variantRegistry[name]
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
	return variants
}

// GetAllVariants returns all available algorithm variants.
func GetAllVariants() []AlgorithmVariant {
	variants := make([]AlgorithmVariant, 0, 7)
	seen := make(map[AlgorithmVariant]bool)

	for _, variant := range variantRegistry {
		if !seen[variant] {
			variants = append(variants, variant)
			seen[variant] = true
		}
	}
	return variants
}

// =============================================================================
// Standard MA Variant
// =============================================================================

// StandardMAVariant represents the standard Mayfly Algorithm.
type StandardMAVariant struct{}

func (v *StandardMAVariant) Name() string {
	return "MA"
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
	return "DESMA"
}

func (v *DESMAVariant) FullName() string {
	return "Dynamic Elite Strategy Mayfly Algorithm"
}

func (v *DESMAVariant) Description() string {
	return "Enhanced with dynamic elite generation for better local optima escape. 70%+ improvement on multimodal problems."
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
	return 1.08 // ~8% more evaluations
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
	return "OLCE-MA"
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
	if characteristics.Modality == HighlyMultimodal {
		score += 0.4
	} else if characteristics.Modality == Multimodal {
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
	return 1.12 // ~12% more evaluations
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
	return "EOBBMA"
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

// GSASMAVariant represents the Golden Sine with Simulated Annealing MA.
type GSASMAVariant struct{}

func (v *GSASMAVariant) Name() string {
	return "GSASMA"
}

func (v *GSASMAVariant) FullName() string {
	return "Golden Sine Algorithm with Simulated Annealing Mayfly Algorithm"
}

func (v *GSASMAVariant) Description() string {
	return "Golden ratio + SA + hybrid mutation. 10-20% improvement with fast convergence on engineering problems."
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
	return 1.15 // ~15% more evaluations
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
	return "MPMA"
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

// AOBLMOAVariant represents the Aquila Optimizer-Based Learning Multi-Objective Algorithm.
type AOBLMOAVariant struct{}

func (v *AOBLMOAVariant) Name() string {
	return "AOBLMOA"
}

func (v *AOBLMOAVariant) FullName() string {
	return "Aquila Optimizer-Based Learning Multi-Objective Algorithm"
}

func (v *AOBLMOAVariant) Description() string {
	return "Hybrid Mayfly-Aquila with 4 hunting strategies. Adaptive multi-phase optimization + multi-objective support."
}

func (v *AOBLMOAVariant) GetConfig() *Config {
	return NewAOBLMOAConfig()
}

func (v *AOBLMOAVariant) ApplicableTo(characteristics ProblemCharacteristics) float64 {
	score := 0.5

	// Essential for multi-objective problems
	if characteristics.MultiObjective {
		score += 0.4
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
		"Multi-objective optimization",
		"Adaptive strategy switching",
		"Complex multi-phase problems",
		"Engineering design tradeoffs",
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
// Example: NewBuilder("desma").ForProblem(fn, 10, -5, 5).WithIterations(500).Build()
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

// NewBuilderFromVariant creates a builder from an existing variant instance.
func NewBuilderFromVariant(variant AlgorithmVariant) *VariantBuilder {
	return &VariantBuilder{
		variant: variant,
		config:  variant.GetConfig(),
	}
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

// WithConfig applies a custom configuration function.
// Example: WithConfig(func(c *Config) { c.A1 = 2.0; c.Beta = 3.0 })
func (b *VariantBuilder) WithConfig(fn func(*Config)) *VariantBuilder {
	if b == nil {
		return nil
	}
	fn(b.config)
	return b
}

// Build returns the configured Config ready for optimization.
func (b *VariantBuilder) Build() (*Config, error) {
	if b == nil {
		return nil, fmt.Errorf("builder is nil (unknown variant?)")
	}
	if b.config.ObjectiveFunc == nil {
		return nil, fmt.Errorf("objective function not set")
	}
	if b.config.ProblemSize <= 0 {
		return nil, fmt.Errorf("problem size must be positive")
	}
	return b.config, nil
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

// min returns the minimum of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
