package main

import (
	"fmt"

	"github.com/cwbudde/mayfly"
)

func main() {
	fmt.Println("=================================================================")
	fmt.Println("      Comprehensive Mayfly Algorithm Benchmark Suite")
	fmt.Println("=================================================================\n")

	// Configuration
	runs := 20 // Number of runs for statistical significance
	iterations := 300
	problemSize := 10

	// Define benchmarks to test
	benchmarks := []struct {
		name  string
		fn    mayfly.ObjectiveFunction
		lower float64
		upper float64
	}{
		{"Sphere", mayfly.Sphere, -10, 10},
		{"Rastrigin", mayfly.Rastrigin, -5.12, 5.12},
		{"Rosenbrock", mayfly.Rosenbrock, -5, 10},
		{"Ackley", mayfly.Ackley, -32.768, 32.768},
		{"Schwefel", mayfly.Schwefel, -500, 500},
	}

	// Example 1: Compare all variants on Rastrigin
	fmt.Println("Example 1: Comparing All Variants on Rastrigin Function")
	fmt.Println("=================================================================")

	runner := mayfly.NewComparisonRunner().
		WithRuns(runs).
		WithIterations(iterations).
		WithVerbose(true)

	result := runner.Compare(
		"Rastrigin",
		mayfly.Rastrigin,
		problemSize,
		-5.12,
		5.12,
	)

	result.PrintComparisonResults()

	// Example 2: Compare specific variants
	fmt.Println("\n\nExample 2: Comparing MA, DESMA, and OLCE-MA on Multiple Functions")
	fmt.Println("=================================================================")

	specificRunner := mayfly.NewComparisonRunner().
		WithVariantNames("ma", "desma", "olce").
		WithRuns(15).
		WithIterations(iterations).
		WithVerbose(false)

	for _, benchmark := range benchmarks {
		fmt.Printf("\n--- %s Function ---\n", benchmark.name)
		result := specificRunner.Compare(
			benchmark.name,
			benchmark.fn,
			problemSize,
			benchmark.lower,
			benchmark.upper,
		)

		// Print summary
		fmt.Printf("\nRankings:\n")
		for i, name := range result.AlgorithmNames {
			stats := result.Statistics[i]
			rank := result.Rankings[i]
			fmt.Printf("  %d. %-10s - Mean: %.6f, StdDev: %.6f\n",
				rank, name, stats.Mean, stats.StdDev)
		}

		if result.FriedmanResult != nil && result.FriedmanResult.Significant {
			fmt.Printf("\n  Significant difference detected (p=%.4f)\n", result.FriedmanResult.PValue)
		}
	}

	// Example 3: Compare variants with recommended algorithms
	fmt.Println("\n\n\nExample 3: Automatic Algorithm Selection and Comparison")
	fmt.Println("=================================================================")

	selector := mayfly.NewAlgorithmSelector()

	for _, benchmark := range benchmarks {
		fmt.Printf("\n--- %s Function ---\n", benchmark.name)

		// Get recommendation
		rec := mayfly.RecommendForBenchmark(benchmark.name)
		fmt.Printf("Recommended: %s (Score: %.1f%%, Confidence: %.1f%%)\n",
			rec.Variant.Name(),
			rec.Score*100,
			rec.Confidence*100)

		// Compare recommended vs standard MA
		compRunner := mayfly.NewComparisonRunner().
			WithVariants(&mayfly.StandardMAVariant{}, rec.Variant).
			WithRuns(10).
			WithIterations(iterations).
			WithVerbose(false)

		result := compRunner.Compare(
			benchmark.name,
			benchmark.fn,
			problemSize,
			benchmark.lower,
			benchmark.upper,
		)

		// Show improvement
		maStat := result.Statistics[0]
		recStat := result.Statistics[1]
		improvement := (maStat.Mean - recStat.Mean) / maStat.Mean * 100

		fmt.Printf("\nResults:\n")
		fmt.Printf("  Standard MA:   %.6f ± %.6f\n", maStat.Mean, maStat.StdDev)
		fmt.Printf("  %s: %.6f ± %.6f\n", rec.Variant.Name(), recStat.Mean, recStat.StdDev)

		if improvement > 0 {
			fmt.Printf("  Improvement: %.2f%% better\n", improvement)
		} else {
			fmt.Printf("  Improvement: %.2f%% worse\n", -improvement)
		}

		// Check statistical significance
		if len(result.WilcoxonTests) > 0 && len(result.WilcoxonTests[0]) > 1 {
			wilcoxon := result.WilcoxonTests[0][1]
			if wilcoxon.Significant {
				fmt.Printf("  Statistical Significance: YES (p=%.4f)\n", wilcoxon.PValue)
			} else {
				fmt.Printf("  Statistical Significance: NO (p=%.4f)\n", wilcoxon.PValue)
			}
		}
	}

	// Example 4: Performance summary table
	fmt.Println("\n\n\nExample 4: Performance Summary Table")
	fmt.Println("=================================================================")

	// Run each variant on each benchmark once for quick overview
	quickRunner := mayfly.NewComparisonRunner().
		WithRuns(5).
		WithIterations(200).
		WithVerbose(false)

	variants := []string{"ma", "desma", "olce", "eobbma", "gsasma", "mpma", "aoblmoa"}

	fmt.Println("\nAverage Performance (Mean Cost across 5 runs, 200 iterations):")
	fmt.Println("─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("%-12s", "Algorithm")
	for _, bm := range benchmarks {
		fmt.Printf(" | %-12s", bm.name)
	}
	fmt.Println()
	fmt.Println("─────────────────────────────────────────────────────────────────────────────")

	for _, variantName := range variants {
		fmt.Printf("%-12s", variantName)

		for _, benchmark := range benchmarks {
			result := quickRunner.
				WithVariantNames(variantName).
				Compare(
					benchmark.name,
					benchmark.fn,
					problemSize,
					benchmark.lower,
					benchmark.upper,
				)

			if len(result.Statistics) > 0 {
				fmt.Printf(" | %12.6f", result.Statistics[0].Mean)
			}
		}
		fmt.Println()
	}

	fmt.Println("─────────────────────────────────────────────────────────────────────────────")

	// Example 5: Export and load configuration
	fmt.Println("\n\n\nExample 5: Configuration Export/Import")
	fmt.Println("=================================================================")

	// Create a configuration
	config := mayfly.NewOLCEConfig()
	config.ProblemSize = 20
	config.LowerBound = -10
	config.UpperBound = 10
	config.MaxIterations = 500

	// Save to file
	configPath := "/tmp/mayfly_config.json"
	err := mayfly.SaveConfigToFile(config, configPath)
	if err != nil {
		fmt.Printf("Error saving config: %v\n", err)
	} else {
		fmt.Printf("Configuration saved to: %s\n", configPath)
	}

	// Load from file
	loadedConfig, err := mayfly.LoadConfigFromFile(configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
	} else {
		fmt.Printf("Configuration loaded successfully\n")
		fmt.Printf("  Problem Size: %d\n", loadedConfig.ProblemSize)
		fmt.Printf("  Max Iterations: %d\n", loadedConfig.MaxIterations)
		fmt.Printf("  Use OLCE: %t\n", loadedConfig.UseOLCE)
		fmt.Printf("  Orthogonal Factor: %.2f\n", loadedConfig.OrthogonalFactor)
	}

	// Example 6: Preset configurations
	fmt.Println("\n\n\nExample 6: Using Preset Configurations")
	fmt.Println("=================================================================")

	presets := []mayfly.ConfigPreset{
		mayfly.PresetMultimodal,
		mayfly.PresetDeceptive,
		mayfly.PresetFastConvergence,
	}

	for _, preset := range presets {
		config, err := mayfly.NewPresetConfig(preset)
		if err != nil {
			fmt.Printf("Error creating preset %s: %v\n", preset, err)
			continue
		}

		// Test on appropriate function
		var testFn mayfly.ObjectiveFunction
		var testName string

		switch preset {
		case mayfly.PresetMultimodal:
			testFn = mayfly.Rastrigin
			testName = "Rastrigin"
		case mayfly.PresetDeceptive:
			testFn = mayfly.Schwefel
			testName = "Schwefel"
		case mayfly.PresetFastConvergence:
			testFn = mayfly.Ackley
			testName = "Ackley"
		}

		config.ObjectiveFunc = testFn
		config.ProblemSize = 10
		config.LowerBound = -10
		config.UpperBound = 10
		config.MaxIterations = 200

		result, err := mayfly.Optimize(config)
		if err != nil {
			fmt.Printf("Error optimizing with preset %s: %v\n", preset, err)
			continue
		}

		fmt.Printf("\nPreset: %s (on %s)\n", preset, testName)
		fmt.Printf("  Best Cost: %.6f\n", result.GlobalBest.Cost)
		fmt.Printf("  Function Evaluations: %d\n", result.FuncEvalCount)
	}

	fmt.Println("\n=================================================================")
	fmt.Println("              Benchmark Suite Complete")
	fmt.Println("=================================================================")
}
