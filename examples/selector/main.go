package main

import (
	"fmt"

	"github.com/cwbudde/mayfly"
)

func main() {
	fmt.Println("=================================================================")
	fmt.Println("   Mayfly Algorithm Selector - Intelligent Algorithm Selection")
	fmt.Println("=================================================================\n")

	// Example 1: Get recommendations for a highly multimodal problem
	fmt.Println("Example 1: Rastrigin Function (Highly Multimodal)")
	fmt.Println("-----------------------------------------------------------------")

	characteristics := mayfly.ProblemCharacteristics{
		Dimensionality:            30,
		Modality:                  mayfly.HighlyMultimodal,
		Landscape:                 mayfly.Rugged,
		ExpensiveEvaluations:      false,
		RequiresFastConvergence:   false,
		RequiresStableConvergence: false,
		MultiObjective:            false,
	}

	selector := mayfly.NewAlgorithmSelector()
	recommendations := selector.RecommendAlgorithms(characteristics)

	mayfly.PrintRecommendations(recommendations)
	fmt.Println()

	// Example 2: Using the best recommendation
	fmt.Println("\nExample 2: Running Optimization with Best Recommended Algorithm")
	fmt.Println("-----------------------------------------------------------------")

	best := recommendations[0]
	fmt.Printf("Selected: %s (%s)\n", best.Variant.Name(), best.Variant.FullName())
	fmt.Printf("Score: %.2f%%, Confidence: %.2f%%\n", best.Score*100, best.Confidence*100)
	fmt.Printf("Reason: %s\n\n", best.Reasoning)

	// Build and run with fluent API
	result, err := mayfly.NewBuilderFromVariant(best.Variant).
		ForProblem(mayfly.Rastrigin, 10, -5.12, 5.12).
		WithIterations(300).
		Optimize()

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Best Cost: %.6f\n", result.GlobalBest.Cost)
	fmt.Printf("Function Evaluations: %d\n", result.FuncEvalCount)
	fmt.Printf("Iterations: %d\n\n", result.IterationCount)

	// Example 3: Recommendations for different benchmark functions
	fmt.Println("\nExample 3: Algorithm Recommendations for Standard Benchmarks")
	fmt.Println("-----------------------------------------------------------------")

	benchmarks := []string{"Sphere", "Rastrigin", "Rosenbrock", "Schwefel", "Ackley"}

	for _, benchmark := range benchmarks {
		rec := mayfly.RecommendForBenchmark(benchmark)
		fmt.Printf("%-12s -> %-10s (Score: %5.1f%%, Conf: %5.1f%%)\n",
			benchmark,
			rec.Variant.Name(),
			rec.Score*100,
			rec.Confidence*100)
	}
	fmt.Println()

	// Example 4: Problem classification
	fmt.Println("\nExample 4: Automatic Problem Classification")
	fmt.Println("-----------------------------------------------------------------")

	fmt.Println("Classifying Schwefel function...")
	classified := mayfly.ClassifyProblem(mayfly.Schwefel, 10, -500, 500)

	fmt.Printf("  Dimensionality: %d\n", classified.Dimensionality)
	fmt.Printf("  Modality: ")
	switch classified.Modality {
	case mayfly.Unimodal:
		fmt.Println("Unimodal")
	case mayfly.Multimodal:
		fmt.Println("Multimodal")
	case mayfly.HighlyMultimodal:
		fmt.Println("Highly Multimodal")
	}

	fmt.Printf("  Landscape: ")
	switch classified.Landscape {
	case mayfly.Smooth:
		fmt.Println("Smooth")
	case mayfly.Rugged:
		fmt.Println("Rugged")
	case mayfly.Deceptive:
		fmt.Println("Deceptive")
	case mayfly.NarrowValley:
		fmt.Println("Narrow Valley")
	}

	bestForClassified := selector.RecommendBest(classified)
	fmt.Printf("\n  Recommended: %s (Score: %.1f%%)\n",
		bestForClassified.Variant.Name(),
		bestForClassified.Score*100)
	fmt.Printf("  Reasoning: %s\n\n", bestForClassified.Reasoning)

	// Example 5: List all available variants
	fmt.Println("\nExample 5: Available Algorithm Variants")
	fmt.Println("-----------------------------------------------------------------")

	variants := mayfly.GetAllVariants()
	for _, v := range variants {
		fmt.Printf("%-12s - %s\n", v.Name(), v.FullName())
		fmt.Printf("              %s\n", v.Description())
		fmt.Printf("              Overhead: ~%.0f%% extra evaluations\n",
			(v.EstimatedOverhead()-1)*100)
		fmt.Printf("              Best for: %v\n\n", v.RecommendedFor())
	}

	// Example 6: Using presets
	fmt.Println("\nExample 6: Configuration Presets")
	fmt.Println("-----------------------------------------------------------------")

	mayfly.PrintPresets()

	// Create a preset configuration
	fmt.Println("\nCreating preset configuration for deceptive problems...")
	config, err := mayfly.NewPresetConfig(mayfly.PresetDeceptive)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Set problem-specific parameters
	config.ObjectiveFunc = mayfly.Schwefel
	config.ProblemSize = 10
	config.LowerBound = -500
	config.UpperBound = 500

	fmt.Printf("Preset selected: EOBBMA (Elite Opposition-Based Bare Bones MA)\n")
	fmt.Printf("Running optimization...\n")

	presetResult, err := mayfly.Optimize(config)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Best Cost: %.6f\n", presetResult.GlobalBest.Cost)
	fmt.Printf("Function Evaluations: %d\n\n", presetResult.FuncEvalCount)

	// Example 7: Fluent builder API
	fmt.Println("\nExample 7: Fluent Builder API")
	fmt.Println("-----------------------------------------------------------------")

	fmt.Println("Building configuration with fluent API...")

	builderResult, err := mayfly.NewBuilder("gsasma").
		ForProblem(mayfly.Ackley, 20, -32.768, 32.768).
		WithIterations(400).
		WithPopulation(30, 30).
		WithConfig(func(c *mayfly.Config) {
			// Custom tuning
			c.CoolingRate = 0.97
			c.InitialTemperature = 200.0
		}).
		Optimize()

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Algorithm: GSASMA\n")
	fmt.Printf("Best Cost: %.6f\n", builderResult.GlobalBest.Cost)
	fmt.Printf("Function Evaluations: %d\n\n", builderResult.FuncEvalCount)

	fmt.Println("=================================================================")
	fmt.Println("                     Examples Complete")
	fmt.Println("=================================================================")
}
