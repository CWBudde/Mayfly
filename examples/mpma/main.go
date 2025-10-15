// Package main demonstrates the Median Position-Based Mayfly Algorithm (MPMA).
//
// MPMA enhances the standard Mayfly Algorithm by using the median position
// of the population instead of just the global best, combined with a non-linear
// gravity coefficient for better exploration-exploitation balance.
package main

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/cwbudde/mayfly"
)

func main() {
	fmt.Println("================================================================================")
	fmt.Println("Median Position-Based Mayfly Algorithm (MPMA) Demonstration")
	fmt.Println("================================================================================")
	fmt.Println()

	// Define test problems
	problems := []struct {
		name      string
		objective func([]float64) float64
		dim       int
		lower     float64
		upper     float64
		optimum   float64
	}{
		{"Sphere", mayfly.Sphere, 10, -10, 10, 0},
		{"Rastrigin", mayfly.Rastrigin, 10, -5.12, 5.12, 0},
		{"Rosenbrock", mayfly.Rosenbrock, 10, -5, 10, 0},
		{"Ackley", mayfly.Ackley, 10, -32.768, 32.768, 0},
		{"Schwefel", mayfly.Schwefel, 10, -500, 500, 0},
	}

	// Run comparison for each problem
	for _, prob := range problems {
		fmt.Printf("\n%s\n", prob.name)
		fmt.Println(repeat("=", 80))
		fmt.Printf("Problem: %s (%dD)\n", prob.name, prob.dim)
		fmt.Printf("Bounds: [%.2f, %.2f]\n", prob.lower, prob.upper)
		fmt.Printf("Known optimum: %.2f\n\n", prob.optimum)

		// Fixed seed for fair comparison
		seed := int64(42)
		iterations := 500
		runs := 10

		// Test Standard MA
		fmt.Println("Testing Standard MA...")
		standardResults := runMultiple(prob.objective, prob.dim, prob.lower, prob.upper, iterations, runs, seed, "standard")
		standardMean, standardStd, standardBest := calculateStats(standardResults)

		// Test DESMA
		fmt.Println("Testing DESMA...")
		desmaResults := runMultiple(prob.objective, prob.dim, prob.lower, prob.upper, iterations, runs, seed+1, "desma")
		desmaMean, desmaStd, desmaBest := calculateStats(desmaResults)

		// Test MPMA with different gravity types
		fmt.Println("Testing MPMA (Linear Gravity)...")
		mpmaLinearResults := runMultiple(prob.objective, prob.dim, prob.lower, prob.upper, iterations, runs, seed+2, "mpma-linear")
		mpmaLinearMean, mpmaLinearStd, mpmaLinearBest := calculateStats(mpmaLinearResults)

		fmt.Println("Testing MPMA (Exponential Gravity)...")
		mpmaExpResults := runMultiple(prob.objective, prob.dim, prob.lower, prob.upper, iterations, runs, seed+3, "mpma-exp")
		mpmaExpMean, mpmaExpStd, mpmaExpBest := calculateStats(mpmaExpResults)

		fmt.Println("Testing MPMA (Sigmoid Gravity)...")
		mpmaSigmoidResults := runMultiple(prob.objective, prob.dim, prob.lower, prob.upper, iterations, runs, seed+4, "mpma-sigmoid")
		mpmaSigmoidMean, mpmaSigmoidStd, mpmaSigmoidBest := calculateStats(mpmaSigmoidResults)

		fmt.Println("Testing MPMA (Weighted Median)...")
		mpmaWeightedResults := runMultiple(prob.objective, prob.dim, prob.lower, prob.upper, iterations, runs, seed+5, "mpma-weighted")
		mpmaWeightedMean, mpmaWeightedStd, mpmaWeightedBest := calculateStats(mpmaWeightedResults)

		// Print results
		fmt.Println("\nResults Summary:")
		fmt.Println(repeat("-", 80))
		fmt.Printf("%-25s | %15s | %15s | %15s\n", "Algorithm", "Mean Cost", "Std Dev", "Best Cost")
		fmt.Println(repeat("-", 80))
		fmt.Printf("%-25s | %15.6e | %15.6e | %15.6e\n", "Standard MA", standardMean, standardStd, standardBest)
		fmt.Printf("%-25s | %15.6e | %15.6e | %15.6e\n", "DESMA", desmaMean, desmaStd, desmaBest)
		fmt.Printf("%-25s | %15.6e | %15.6e | %15.6e\n", "MPMA (Linear)", mpmaLinearMean, mpmaLinearStd, mpmaLinearBest)
		fmt.Printf("%-25s | %15.6e | %15.6e | %15.6e\n", "MPMA (Exponential)", mpmaExpMean, mpmaExpStd, mpmaExpBest)
		fmt.Printf("%-25s | %15.6e | %15.6e | %15.6e\n", "MPMA (Sigmoid)", mpmaSigmoidMean, mpmaSigmoidStd, mpmaSigmoidBest)
		fmt.Printf("%-25s | %15.6e | %15.6e | %15.6e\n", "MPMA (Weighted Median)", mpmaWeightedMean, mpmaWeightedStd, mpmaWeightedBest)
		fmt.Println(repeat("-", 80))

		// Calculate improvements
		fmt.Println("\nImprovement Analysis:")
		fmt.Println(repeat("-", 80))
		fmt.Printf("MPMA (Linear) vs Standard MA:    %+.2f%%\n", improvement(standardMean, mpmaLinearMean))
		fmt.Printf("MPMA (Exponential) vs Standard:  %+.2f%%\n", improvement(standardMean, mpmaExpMean))
		fmt.Printf("MPMA (Sigmoid) vs Standard:      %+.2f%%\n", improvement(standardMean, mpmaSigmoidMean))
		fmt.Printf("MPMA (Weighted) vs Standard:     %+.2f%%\n", improvement(standardMean, mpmaWeightedMean))
		fmt.Printf("MPMA (Linear) vs DESMA:          %+.2f%%\n", improvement(desmaMean, mpmaLinearMean))
		fmt.Println(repeat("-", 80))

		// Identify best performer
		bestMean := math.Min(standardMean, math.Min(desmaMean, math.Min(mpmaLinearMean, math.Min(mpmaExpMean, math.Min(mpmaSigmoidMean, mpmaWeightedMean)))))
		var bestAlgo string
		switch bestMean {
		case standardMean:
			bestAlgo = "Standard MA"
		case desmaMean:
			bestAlgo = "DESMA"
		case mpmaLinearMean:
			bestAlgo = "MPMA (Linear)"
		case mpmaExpMean:
			bestAlgo = "MPMA (Exponential)"
		case mpmaSigmoidMean:
			bestAlgo = "MPMA (Sigmoid)"
		case mpmaWeightedMean:
			bestAlgo = "MPMA (Weighted Median)"
		}
		fmt.Printf("\nBest performer: %s (Mean: %.6e)\n", bestAlgo, bestMean)
	}

	fmt.Println("\n================================================================================")
	fmt.Println("Key Insights:")
	fmt.Println("================================================================================")
	fmt.Println("1. Linear Gravity: Simple linear decay, balanced exploration-exploitation")
	fmt.Println("2. Exponential Gravity: Fast convergence, good for exploitation")
	fmt.Println("3. Sigmoid Gravity: S-curve balance, smooth transition between phases")
	fmt.Println("4. Weighted Median: Emphasizes better solutions, robust to outliers")
	fmt.Println("5. Median Position: More robust to outliers than mean position")
	fmt.Println("6. Non-linear Gravity: Better control of exploration-exploitation balance")
	fmt.Println("================================================================================")
}

// runMultiple runs the specified algorithm multiple times and returns all results
func runMultiple(objective func([]float64) float64, dim int, lower, upper float64, iterations, runs int, seed int64, variant string) []float64 {
	results := make([]float64, runs)

	for i := 0; i < runs; i++ {
		var config *mayfly.Config

		switch variant {
		case "standard":
			config = mayfly.NewDefaultConfig()
		case "desma":
			config = mayfly.NewDESMAConfig()
		case "mpma-linear":
			config = mayfly.NewMPMAConfig()
			config.GravityType = "linear"
			config.UseWeightedMedian = false
		case "mpma-exp":
			config = mayfly.NewMPMAConfig()
			config.GravityType = "exponential"
			config.UseWeightedMedian = false
		case "mpma-sigmoid":
			config = mayfly.NewMPMAConfig()
			config.GravityType = "sigmoid"
			config.UseWeightedMedian = false
		case "mpma-weighted":
			config = mayfly.NewMPMAConfig()
			config.GravityType = "linear"
			config.UseWeightedMedian = true
		default:
			config = mayfly.NewDefaultConfig()
		}

		config.ObjectiveFunc = objective
		config.ProblemSize = dim
		config.LowerBound = lower
		config.UpperBound = upper
		config.MaxIterations = iterations
		config.Rand = rand.New(rand.NewSource(seed + int64(i)))

		result, err := mayfly.Optimize(config)
		if err != nil {
			fmt.Printf("Error in run %d: %v\n", i+1, err)
			continue
		}

		results[i] = result.GlobalBest.Cost
	}

	return results
}

// calculateStats calculates mean, standard deviation, and best from results
func calculateStats(results []float64) (mean, std, best float64) {
	n := float64(len(results))
	best = results[0]

	// Calculate mean
	sum := 0.0
	for _, v := range results {
		sum += v
		if v < best {
			best = v
		}
	}
	mean = sum / n

	// Calculate standard deviation
	sumSq := 0.0
	for _, v := range results {
		diff := v - mean
		sumSq += diff * diff
	}
	std = math.Sqrt(sumSq / n)

	return mean, std, best
}

// improvement calculates percentage improvement (positive = better)
func improvement(baseline, current float64) float64 {
	if baseline == 0 {
		if current == 0 {
			return 0
		}
		return -100 // Worse if baseline is 0 but current is not
	}
	return ((baseline - current) / baseline) * 100
}

// repeat creates a string of repeated characters
func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
