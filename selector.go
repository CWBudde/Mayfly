package mayfly

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"
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
func (s *AlgorithmSelector) calculateConfidence(characteristics ProblemCharacteristics,
	variant AlgorithmVariant,
) float64 {
	confidence := 0.7 // Base confidence

	// Higher confidence for specific characteristics
	if characteristics.MultiObjective {
		if variant.Name() == nameAOBLMOA {
			confidence = 0.95 // Very confident for multi-objective
		} else {
			confidence = 0.3 // Low confidence for non-MO algorithms on MO problems
		}
	}

	if characteristics.Landscape == Deceptive && variant.Name() == nameEOBBMA {
		confidence = 0.9 // EOBBMA is proven on deceptive functions
	}

	if characteristics.RequiresFastConvergence && variant.Name() == nameGSASMA {
		confidence = 0.85 // GSASMA is designed for fast convergence
	}

	if characteristics.RequiresStableConvergence && variant.Name() == nameMPMA {
		confidence = 0.85 // MPMA is designed for stability
	}

	// Lower confidence for expensive evaluations with high-overhead algorithms
	if characteristics.ExpensiveEvaluations && variant.EstimatedOverhead() > 1.15 {
		confidence *= 0.7
	}

	return math.Min(confidence, 1.0)
}

// generateReasoning creates a human-readable explanation for the recommendation.
func (s *AlgorithmSelector) generateReasoning(characteristics ProblemCharacteristics,
	variant AlgorithmVariant, score float64,
) string {
	reasons := make([]string, 0, 3)

	// Analyze key characteristics
	if characteristics.MultiObjective && variant.Name() == nameAOBLMOA {
		reasons = append(reasons, "Multi-objective support required")
	}

	if characteristics.Modality == HighlyMultimodal {
		if variant.Name() == nameOLCEMA {
			reasons = append(reasons, "Highly multimodal problem benefits from orthogonal learning")
		} else if variant.Name() == nameDESMA {
			reasons = append(reasons, "Multimodal problem benefits from elite strategy")
		}
	}

	if characteristics.Landscape == Deceptive && variant.Name() == nameEOBBMA {
		reasons = append(reasons, "Lévy flight effective on deceptive landscapes")
	}

	if characteristics.Landscape == NarrowValley && variant.Name() == nameMPMA {
		reasons = append(reasons, "Median guidance handles ill-conditioned problems well")
	}

	if characteristics.RequiresFastConvergence && variant.Name() == nameGSASMA {
		reasons = append(reasons, "Fast convergence via simulated annealing")
	}

	if characteristics.RequiresStableConvergence && variant.Name() == nameMPMA {
		reasons = append(reasons, "Stable convergence via robust median guidance")
	}

	if characteristics.ExpensiveEvaluations && variant.EstimatedOverhead() <= 1.05 {
		reasons = append(reasons, "Low overhead suitable for expensive evaluations")
	}

	if characteristics.Dimensionality >= 20 && variant.Name() == nameOLCEMA {
		reasons = append(reasons, "High dimensionality benefits from diversity")
	}

	// Generate summary
	if len(reasons) == 0 {
		return fmt.Sprintf("Score: %.2f - %s", score, variant.Description())
	}

	return strings.Join(reasons, "; ")
}

// Sampling budget for ClassifyProblem. Deliberately small: classification runs
// before the optimization, and a classifier that costs a meaningful fraction of
// the run it is choosing for is not worth having.
const (
	// classifyLines is the number of random line scans across the box.
	classifyLines = 6
	// classifyLineSteps is the number of samples along each line. It has to be
	// dense enough not to alias a landscape with per-unit structure, such as
	// Rastrigin's lattice, across a box tens of units wide.
	classifyLineSteps = 65
	// classifyIterations, classifyRuns and classifyPopulation size the short
	// optimizations estimateStability uses to see how much the outcome depends
	// on the seed.
	classifyIterations = 20
	classifyRuns       = 3
	classifyPopulation = 10
)

// Classification thresholds, all applied to scale-free quantities so that they
// mean the same thing on Sphere over [-5,5] and on Schwefel over [-500,500].
const (
	// unimodalTurningPoints is the average number of direction changes per line
	// scan below which the landscape reads as single-basin. A quadratic bowl
	// crossed by a straight line turns exactly once.
	unimodalTurningPoints = 1.5
	// multimodalTurningPoints separates a handful of optima from a lattice.
	// Measured over forty seeds at d=10 the gap it sits in is wide: the
	// single-basin functions turn about once per line, Schwefel six to nine
	// times, Rastrigin ten or more. Six sat on Schwefel's lower edge and made
	// the verdict seed-dependent, so the threshold sits below it.
	multimodalTurningPoints = 5.0
	// smoothRoughness is the total variation along a line scan, in units of
	// that line's own value range, at or above which the landscape reads as
	// rugged. A line crossing a single basin descends once and climbs once, so
	// its total variation cannot exceed twice its range; anything above 2 has
	// turned more often than a single basin can explain. The margin over 2 is
	// small on purpose. Measured over eight seeds at d=10, the single-basin
	// functions top out at 1.8 (Sphere, Rosenbrock, BentCigar, Discus) while
	// Schwefel starts at 2.5 and Rastrigin at 4.2, so a higher threshold buys
	// no safety and costs Schwefel its Rugged verdict.
	smoothRoughness = 2.2
)

// ClassifyProblem samples an objective function to estimate its characteristics.
//
// It fills in Dimensionality, Modality, Landscape and
// RequiresStableConvergence. ExpensiveEvaluations, RequiresFastConvergence and
// MultiObjective are left at false, because they are facts about the caller's
// problem and budget rather than about the function's values; set them on the
// returned value before passing it to a selector.
//
// # What it can and cannot see
//
// Modality and Landscape come from a handful of straight-line scans across the
// box: how often the function changes direction along a line, and how much
// total variation it accumulates relative to that line's own range. Both
// quantities are scale-free, so the same thresholds apply whatever the bounds
// and whatever the units of the cost.
//
// Landscape is therefore only ever reported as Smooth or Rugged. Deceptive
// (gradients that lead away from the global optimum) and NarrowValley (an
// ill-conditioned basin) are statements about where the optimum sits relative
// to the terrain, and a few dozen samples cannot establish either. A caller who
// knows their problem is Schwefel-like or Rosenbrock-like should say so by
// setting Landscape on the returned value.
//
// The scans also have a resolution limit: structure much finer than the step
// spacing, or much smaller in amplitude than the range of the whole box, does
// not show up. Griewank over [-600,600] is the standard example -- its cosine
// ripples are a few units wide and of order one tall against a range of order
// 100000, so the box-scale verdict is a bowl even though an optimizer working
// near the optimum meets every ripple.
//
// The estimates are coarse heuristics. Treat a classification as a starting
// point a caller who knows their problem should override.
//
// Pass a seeded generator as rng to make a classification reproducible; nil
// draws a fresh one.
func ClassifyProblem(
	fn ObjectiveFunction,
	size int,
	lower, upper float64,
	rng *rand.Rand,
) ProblemCharacteristics {
	if rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	turningPoints, roughness := lineScanStatistics(fn, size, lower, upper, rng)
	stability := estimateStability(fn, size, lower, upper, rng)

	return ProblemCharacteristics{
		Dimensionality: size,
		Modality:       modalityFromTurningPoints(turningPoints),
		Landscape:      landscapeFromRoughness(roughness),
		// Low stability means the outcome depends heavily on the seed.
		RequiresStableConvergence: stability < 0.5,
	}
}

// modalityFromTurningPoints maps the average number of direction changes per
// line scan onto a modality.
func modalityFromTurningPoints(turningPoints float64) Modality {
	switch {
	case turningPoints >= multimodalTurningPoints:
		return HighlyMultimodal
	case turningPoints >= unimodalTurningPoints:
		return Multimodal
	default:
		return Unimodal
	}
}

// landscapeFromRoughness maps the average normalized total variation per line
// scan onto a landscape. It never returns Deceptive or NarrowValley; see
// ClassifyProblem.
func landscapeFromRoughness(roughness float64) Landscape {
	if roughness >= smoothRoughness {
		return Rugged
	}

	return Smooth
}

// lineScanStatistics walks several random straight lines across the search box
// and returns the average number of direction changes per line and the average
// total variation per line in units of that line's own value range.
//
// Both are scale-free by construction: the turning-point count does not look at
// magnitudes at all, and the roughness divides by the range it measured. That
// is the whole point -- the gradient-magnitude heuristic this replaced compared
// an absolute gradient magnitude against absolute thresholds, so it called
// Sphere over [-5,5] rugged and Sphere over [-500,500] deceptive, which says
// more about the bounds than about the function.
func lineScanStatistics(
	fn ObjectiveFunction,
	size int,
	lower, upper float64,
	rng *rand.Rand,
) (float64, float64) {
	totalTurns, totalRoughness := 0.0, 0.0

	for range classifyLines {
		values := scanLine(fn, size, lower, upper, rng)
		turns, roughness := lineShape(values)
		totalTurns += turns
		totalRoughness += roughness
	}

	return totalTurns / classifyLines, totalRoughness / classifyLines
}

// scanLine samples fn at evenly spaced points along the segment between two
// uniformly drawn points of the box.
func scanLine(fn ObjectiveFunction, size int, lower, upper float64, rng *rand.Rand) []float64 {
	from := unifrndVec(lower, upper, size, rng)
	to := unifrndVec(lower, upper, size, rng)

	point := make([]float64, size)
	values := make([]float64, classifyLineSteps)

	for step := range values {
		fraction := float64(step) / float64(classifyLineSteps-1)
		for j := range point {
			point[j] = from[j] + fraction*(to[j]-from[j])
		}

		values[step] = fn(point)
	}

	return values
}

// lineShape returns the number of direction changes along a scan and its total
// variation divided by its value range. A flat or non-finite scan contributes
// nothing to either.
func lineShape(values []float64) (float64, float64) {
	turns := 0.0
	totalVariation := 0.0
	lowest, highest := math.Inf(1), math.Inf(-1)
	previousSign := 0

	for i := 1; i < len(values); i++ {
		delta := values[i] - values[i-1]
		if !isFinite(delta) {
			return 0, 0
		}

		totalVariation += math.Abs(delta)

		sign := 0
		if delta > 0 {
			sign = 1
		} else if delta < 0 {
			sign = -1
		}

		if sign != 0 {
			if previousSign != 0 && sign != previousSign {
				turns++
			}

			previousSign = sign
		}
	}

	for _, value := range values {
		if !isFinite(value) {
			return 0, 0
		}

		lowest = math.Min(lowest, value)
		highest = math.Max(highest, value)
	}

	valueRange := highest - lowest
	if valueRange <= 0 {
		return turns, 0
	}

	return turns, totalVariation / valueRange
}

// estimateStability runs a few very short optimizations and reports 1/(1+cv) of
// their final costs, in [0,1]. A low value means the outcome depends heavily on
// the seed. Runs that fail are skipped rather than counted, so one failure does
// not drag the mean of the rest.
func estimateStability(
	fn ObjectiveFunction,
	size int,
	lower, upper float64,
	rng *rand.Rand,
) float64 {
	costs := make([]float64, 0, classifyRuns)

	for range classifyRuns {
		config := NewDefaultConfig()
		config.ObjectiveFunc = fn
		config.ProblemSize = size
		config.LowerBound = lower
		config.UpperBound = upper
		config.MaxIterations = classifyIterations
		config.NPop = classifyPopulation
		config.NPopF = classifyPopulation
		config.Rand = rand.New(rand.NewSource(rng.Int63()))

		result, err := Optimize(config)
		if err != nil {
			continue
		}

		costs = append(costs, result.GlobalBest.Cost)
	}

	if len(costs) == 0 {
		return 0
	}

	mean, stdDev := meanAndStdDev(costs)
	cv := stdDev / (math.Abs(mean) + 1e-10)

	return 1.0 / (1.0 + cv)
}

// meanAndStdDev returns the mean and the population standard deviation of
// values. An empty slice yields two zeros.
func meanAndStdDev(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 0
	}

	mean := 0.0
	for _, value := range values {
		mean += value
	}

	mean /= float64(len(values))

	variance := 0.0

	for _, value := range values {
		diff := value - mean
		variance += diff * diff
	}

	variance /= float64(len(values))

	return mean, math.Sqrt(variance)
}

// RecommendForBenchmark provides recommendations for standard benchmark functions.
func RecommendForBenchmark(benchmarkName string) AlgorithmRecommendation {
	selector := NewAlgorithmSelector()

	var characteristics ProblemCharacteristics

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
