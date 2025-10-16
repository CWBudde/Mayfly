package mayfly

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// AlgorithmRecommendation represents a recommended algorithm variant with a confidence score.
type AlgorithmRecommendation struct {
	Variant    AlgorithmVariant
	Reasoning  string
	Score      float64
	Confidence float64
}

// AlgorithmSelector provides intelligent algorithm selection based on problem characteristics.
type AlgorithmSelector struct {
	variants []AlgorithmVariant
}

// NewAlgorithmSelector creates a new algorithm selector with all available variants.
func NewAlgorithmSelector() *AlgorithmSelector {
	return &AlgorithmSelector{
		variants: GetAllVariants(),
	}
}

// RecommendAlgorithms returns ranked algorithm recommendations for the given problem.
// The results are sorted by score (highest first).
func (s *AlgorithmSelector) RecommendAlgorithms(characteristics ProblemCharacteristics) []AlgorithmRecommendation {
	recommendations := make([]AlgorithmRecommendation, 0, len(s.variants))

	for _, variant := range s.variants {
		score := variant.ApplicableTo(characteristics)
		confidence := s.calculateConfidence(characteristics, variant)
		reasoning := s.generateReasoning(characteristics, variant, score)

		recommendations = append(recommendations, AlgorithmRecommendation{
			Variant:    variant,
			Score:      score,
			Confidence: confidence,
			Reasoning:  reasoning,
		})
	}

	// Sort by score descending
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	return recommendations
}

// RecommendBest returns the single best algorithm for the given problem.
func (s *AlgorithmSelector) RecommendBest(characteristics ProblemCharacteristics) AlgorithmRecommendation {
	recommendations := s.RecommendAlgorithms(characteristics)
	if len(recommendations) == 0 {
		// Fallback to standard MA
		return AlgorithmRecommendation{
			Variant:    &StandardMAVariant{},
			Score:      0.5,
			Confidence: 0.5,
			Reasoning:  "Default fallback to Standard MA",
		}
	}

	return recommendations[0]
}

// calculateConfidence estimates how confident we are in the recommendation.
func (s *AlgorithmSelector) calculateConfidence(characteristics ProblemCharacteristics, variant AlgorithmVariant) float64 {
	confidence := 0.7 // Base confidence

	// Higher confidence for specific characteristics
	if characteristics.MultiObjective {
		if variant.Name() == "AOBLMOA" {
			confidence = 0.95 // Very confident for multi-objective
		} else {
			confidence = 0.3 // Low confidence for non-MO algorithms on MO problems
		}
	}

	if characteristics.Landscape == Deceptive && variant.Name() == "EOBBMA" {
		confidence = 0.9 // EOBBMA is proven on deceptive functions
	}

	if characteristics.RequiresFastConvergence && variant.Name() == "GSASMA" {
		confidence = 0.85 // GSASMA is designed for fast convergence
	}

	if characteristics.RequiresStableConvergence && variant.Name() == "MPMA" {
		confidence = 0.85 // MPMA is designed for stability
	}

	// Lower confidence for expensive evaluations with high-overhead algorithms
	if characteristics.ExpensiveEvaluations && variant.EstimatedOverhead() > 1.15 {
		confidence *= 0.7
	}

	return math.Min(confidence, 1.0)
}

// generateReasoning creates a human-readable explanation for the recommendation.
func (s *AlgorithmSelector) generateReasoning(characteristics ProblemCharacteristics, variant AlgorithmVariant, score float64) string {
	reasons := make([]string, 0, 3)

	// Analyze key characteristics
	if characteristics.MultiObjective && variant.Name() == "AOBLMOA" {
		reasons = append(reasons, "Multi-objective support required")
	}

	if characteristics.Modality == HighlyMultimodal {
		if variant.Name() == "OLCE-MA" {
			reasons = append(reasons, "Highly multimodal problem benefits from orthogonal learning")
		} else if variant.Name() == "DESMA" {
			reasons = append(reasons, "Multimodal problem benefits from elite strategy")
		}
	}

	if characteristics.Landscape == Deceptive && variant.Name() == "EOBBMA" {
		reasons = append(reasons, "Lévy flight effective on deceptive landscapes")
	}

	if characteristics.Landscape == NarrowValley && variant.Name() == "MPMA" {
		reasons = append(reasons, "Median guidance handles ill-conditioned problems well")
	}

	if characteristics.RequiresFastConvergence && variant.Name() == "GSASMA" {
		reasons = append(reasons, "Fast convergence via simulated annealing")
	}

	if characteristics.RequiresStableConvergence && variant.Name() == "MPMA" {
		reasons = append(reasons, "Stable convergence via robust median guidance")
	}

	if characteristics.ExpensiveEvaluations && variant.EstimatedOverhead() <= 1.05 {
		reasons = append(reasons, "Low overhead suitable for expensive evaluations")
	}

	if characteristics.Dimensionality >= 20 && variant.Name() == "OLCE-MA" {
		reasons = append(reasons, "High dimensionality benefits from diversity")
	}

	// Generate summary
	if len(reasons) == 0 {
		return fmt.Sprintf("Score: %.2f - %s", score, variant.Description())
	}

	summary := ""

	for i, reason := range reasons {
		if i > 0 {
			summary += "; "
		}

		summary += reason
	}

	return summary
}

// ClassifyProblem analyzes an objective function to determine its characteristics.
// This performs lightweight test runs to estimate problem properties.
func ClassifyProblem(fn ObjectiveFunction, size int, lower, upper float64) ProblemCharacteristics {
	const sampleSize = 50 // Number of random samples

	const testIterations = 20 // Short test runs

	// Sample random points to analyze landscape
	samples := make([]float64, sampleSize)

	for i := 0; i < sampleSize; i++ {
		point := make([]float64, size)
		for j := 0; j < size; j++ {
			point[j] = unifrnd(lower, upper, nil)
		}

		samples[i] = fn(point)
	}

	// Estimate modality from sample variance and distribution
	modality := estimateModality(samples)

	// Estimate landscape from gradient smoothness
	landscape := estimateLandscape(fn, size, lower, upper, sampleSize)

	// Test convergence behavior with a short run
	stability := testConvergenceStability(fn, size, lower, upper, testIterations)

	return ProblemCharacteristics{
		Dimensionality:            size,
		Modality:                  modality,
		Landscape:                 landscape,
		ExpensiveEvaluations:      false,           // User should set this
		RequiresFastConvergence:   false,           // User should set this
		RequiresStableConvergence: stability < 0.5, // Low stability suggests need for stable algorithm
		MultiObjective:            false,           // User should set this
	}
}

// estimateModality estimates the problem's modality from sample distribution.
func estimateModality(samples []float64) Modality {
	if len(samples) < 10 {
		return Multimodal // Conservative default
	}

	// Calculate coefficient of variation (CV = std/mean)
	mean := 0.0
	for _, s := range samples {
		mean += s
	}

	mean /= float64(len(samples))

	variance := 0.0

	for _, s := range samples {
		diff := s - mean
		variance += diff * diff
	}

	variance /= float64(len(samples))
	stdDev := math.Sqrt(variance)

	// Avoid division by zero
	if math.Abs(mean) < 1e-10 {
		return Multimodal
	}

	cv := stdDev / math.Abs(mean)

	// High CV suggests many local optima
	if cv > 2.0 {
		return HighlyMultimodal
	} else if cv > 0.5 {
		return Multimodal
	}

	return Unimodal
}

// estimateLandscape estimates landscape characteristics from gradient samples.
func estimateLandscape(fn ObjectiveFunction, size int, lower, upper float64, samples int) Landscape {
	// Sample random points and small perturbations
	gradientVariance := 0.0
	epsilon := (upper - lower) * 0.001 // Small step

	for i := 0; i < samples/2; i++ {
		point := make([]float64, size)
		for j := 0; j < size; j++ {
			point[j] = unifrnd(lower, upper, nil)
		}

		// Estimate gradient via finite differences
		f0 := fn(point)
		gradients := make([]float64, size)

		for j := 0; j < size; j++ {
			point[j] += epsilon
			f1 := fn(point)
			point[j] -= epsilon // Restore
			gradients[j] = (f1 - f0) / epsilon
		}

		// Calculate gradient magnitude
		gradMag := 0.0
		for _, g := range gradients {
			gradMag += g * g
		}

		gradMag = math.Sqrt(gradMag)

		gradientVariance += gradMag
	}

	gradientVariance /= float64(samples / 2)

	// High gradient variance suggests rugged landscape
	// This is a heuristic classification
	if gradientVariance > 100 {
		return Deceptive
	} else if gradientVariance > 10 {
		return Rugged
	} else if gradientVariance < 0.01 {
		return NarrowValley
	}

	return Smooth
}

// testConvergenceStability runs a short optimization to measure stability.
func testConvergenceStability(fn ObjectiveFunction, size int, lower, upper float64, iterations int) float64 {
	// Run multiple short optimizations
	const runs = 3
	results := make([]float64, runs)

	for i := 0; i < runs; i++ {
		config := NewDefaultConfig()
		config.ObjectiveFunc = fn
		config.ProblemSize = size
		config.LowerBound = lower
		config.UpperBound = upper
		config.MaxIterations = iterations
		config.NPop = 10 // Small population for speed
		config.NPopF = 10

		result, err := Optimize(config)
		if err != nil {
			results[i] = math.Inf(1)
		} else {
			results[i] = result.GlobalBest.Cost
		}
	}

	// Calculate coefficient of variation
	mean := 0.0

	for _, r := range results {
		if !math.IsInf(r, 1) {
			mean += r
		}
	}

	mean /= float64(runs)

	variance := 0.0

	for _, r := range results {
		if !math.IsInf(r, 1) {
			diff := r - mean
			variance += diff * diff
		}
	}

	variance /= float64(runs)

	// Return normalized stability (0 = unstable, 1 = very stable)
	cv := math.Sqrt(variance) / (math.Abs(mean) + 1e-10)
	stability := 1.0 / (1.0 + cv) // Map to [0,1]

	return stability
}

// RecommendForBenchmark provides recommendations for standard benchmark functions.
func RecommendForBenchmark(benchmarkName string) AlgorithmRecommendation {
	selector := NewAlgorithmSelector()

	characteristics := ProblemCharacteristics{}

	switch benchmarkName {
	case "Sphere":
		characteristics = ProblemCharacteristics{
			Dimensionality:            30,
			Modality:                  Unimodal,
			Landscape:                 Smooth,
			ExpensiveEvaluations:      false,
			RequiresFastConvergence:   false,
			RequiresStableConvergence: false,
			MultiObjective:            false,
		}

	case "Rastrigin":
		characteristics = ProblemCharacteristics{
			Dimensionality:            30,
			Modality:                  HighlyMultimodal,
			Landscape:                 Rugged,
			ExpensiveEvaluations:      false,
			RequiresFastConvergence:   false,
			RequiresStableConvergence: false,
			MultiObjective:            false,
		}

	case "Rosenbrock":
		characteristics = ProblemCharacteristics{
			Dimensionality:            30,
			Modality:                  Unimodal,
			Landscape:                 NarrowValley,
			ExpensiveEvaluations:      false,
			RequiresFastConvergence:   false,
			RequiresStableConvergence: true,
			MultiObjective:            false,
		}

	case "Ackley":
		characteristics = ProblemCharacteristics{
			Dimensionality:            30,
			Modality:                  Multimodal,
			Landscape:                 Rugged,
			ExpensiveEvaluations:      false,
			RequiresFastConvergence:   false,
			RequiresStableConvergence: false,
			MultiObjective:            false,
		}

	case "Griewank":
		characteristics = ProblemCharacteristics{
			Dimensionality:            30,
			Modality:                  HighlyMultimodal,
			Landscape:                 Rugged,
			ExpensiveEvaluations:      false,
			RequiresFastConvergence:   false,
			RequiresStableConvergence: false,
			MultiObjective:            false,
		}

	case "Schwefel":
		characteristics = ProblemCharacteristics{
			Dimensionality:            30,
			Modality:                  HighlyMultimodal,
			Landscape:                 Deceptive,
			ExpensiveEvaluations:      false,
			RequiresFastConvergence:   false,
			RequiresStableConvergence: false,
			MultiObjective:            false,
		}

	case "BentCigar", "Discus":
		characteristics = ProblemCharacteristics{
			Dimensionality:            30,
			Modality:                  Unimodal,
			Landscape:                 NarrowValley,
			ExpensiveEvaluations:      false,
			RequiresFastConvergence:   false,
			RequiresStableConvergence: true,
			MultiObjective:            false,
		}

	default:
		// Generic multimodal problem
		characteristics = ProblemCharacteristics{
			Dimensionality:            30,
			Modality:                  Multimodal,
			Landscape:                 Rugged,
			ExpensiveEvaluations:      false,
			RequiresFastConvergence:   false,
			RequiresStableConvergence: false,
			MultiObjective:            false,
		}
	}

	return selector.RecommendBest(characteristics)
}

// PrintRecommendations prints formatted recommendations to console.
func PrintRecommendations(recommendations []AlgorithmRecommendation) {
	fmt.Println("Algorithm Recommendations (ranked by score):")
	fmt.Println("=" + strings.Repeat("=", 79))
	fmt.Printf("%-12s | %-8s | %-10s | %s\n", "Algorithm", "Score", "Confidence", "Reasoning")
	fmt.Println(strings.Repeat("-", 80))

	for _, rec := range recommendations {
		fmt.Printf("%-12s | %6.2f%% | %8.2f%% | %s\n",
			rec.Variant.Name(),
			rec.Score*100,
			rec.Confidence*100,
			rec.Reasoning)
	}

	fmt.Println(strings.Repeat("=", 80))
}
