package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/cwbudde/mayfly"
)

func main() {
	fmt.Println("=== EOBBMA (Elite Opposition-Based Bare Bones MA) Example ===\n")

	// Seed random number generator
	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))

	// Problem parameters
	problemSize := 10
	lowerBound := -10.0
	upperBound := 10.0

	// Test on multimodal Ackley function (EOBBMA should excel here)
	fmt.Println("Testing on Ackley function (multimodal with flat outer region)")
	fmt.Printf("Dimensions: %d\n", problemSize)
	fmt.Printf("Search space: [%.1f, %.1f]\n", lowerBound, upperBound)
	fmt.Printf("Global optimum: f(0,...,0) = 0\n\n")

	// ====================================================================
	// Run Standard Mayfly Algorithm
	// ====================================================================
	fmt.Println("Running Standard Mayfly Algorithm...")

	standardConfig := mayfly.NewDefaultConfig()
	standardConfig.ObjectiveFunc = mayfly.Ackley
	standardConfig.ProblemSize = problemSize
	standardConfig.LowerBound = lowerBound
	standardConfig.UpperBound = upperBound
	standardConfig.MaxIterations = 500
	standardConfig.Rand = rng

	standardStart := time.Now()
	standardResult, err := mayfly.Optimize(standardConfig)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	standardDuration := time.Since(standardStart)

	fmt.Printf("  Best cost: %.6f\n", standardResult.GlobalBest.Cost)
	fmt.Printf("  Function evaluations: %d\n", standardResult.FuncEvalCount)
	fmt.Printf("  Time: %v\n\n", standardDuration)

	// ====================================================================
	// Run EOBBMA Variant
	// ====================================================================
	fmt.Println("Running EOBBMA Variant...")

	eobbmaConfig := mayfly.NewEOBBMAConfig()
	eobbmaConfig.ObjectiveFunc = mayfly.Ackley
	eobbmaConfig.ProblemSize = problemSize
	eobbmaConfig.LowerBound = lowerBound
	eobbmaConfig.UpperBound = upperBound
	eobbmaConfig.MaxIterations = 500
	eobbmaConfig.Rand = rng

	eobbmaStart := time.Now()
	eobbmaResult, err := mayfly.Optimize(eobbmaConfig)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	eobbmaDuration := time.Since(eobbmaStart)

	fmt.Printf("  Best cost: %.6f\n", eobbmaResult.GlobalBest.Cost)
	fmt.Printf("  Function evaluations: %d\n", eobbmaResult.FuncEvalCount)
	fmt.Printf("  Time: %v\n\n", eobbmaDuration)

	// ====================================================================
	// Compare Results
	// ====================================================================
	fmt.Println("=== Comparison ===")

	improvement := ((standardResult.GlobalBest.Cost - eobbmaResult.GlobalBest.Cost) / standardResult.GlobalBest.Cost) * 100
	overhead := ((float64(eobbmaResult.FuncEvalCount) / float64(standardResult.FuncEvalCount)) - 1.0) * 100

	fmt.Printf("Cost improvement: %.2f%%\n", improvement)
	fmt.Printf("Evaluation overhead: %.2f%%\n", overhead)

	if eobbmaResult.GlobalBest.Cost < standardResult.GlobalBest.Cost {
		fmt.Println("\n✓ EOBBMA found a better solution!")
	} else if eobbmaResult.GlobalBest.Cost == standardResult.GlobalBest.Cost {
		fmt.Println("\n= Both algorithms found equivalent solutions")
	} else {
		fmt.Println("\n✗ Standard MA performed better on this run")
	}

	// ====================================================================
	// Test on another complex function: Schwefel
	// ====================================================================
	fmt.Println("\n\n=== Testing on Schwefel function ===")
	fmt.Println("(Highly deceptive with many local minima)")

	lowerSchwefel := -500.0
	upperSchwefel := 500.0

	// Standard MA on Schwefel
	standardConfig.ObjectiveFunc = mayfly.Schwefel
	standardConfig.LowerBound = lowerSchwefel
	standardConfig.UpperBound = upperSchwefel
	standardResult2, _ := mayfly.Optimize(standardConfig)
	fmt.Printf("Standard MA: %.6f (evals: %d)\n",
		standardResult2.GlobalBest.Cost, standardResult2.FuncEvalCount)

	// EOBBMA on Schwefel
	eobbmaConfig.ObjectiveFunc = mayfly.Schwefel
	eobbmaConfig.LowerBound = lowerSchwefel
	eobbmaConfig.UpperBound = upperSchwefel
	eobbmaResult2, _ := mayfly.Optimize(eobbmaConfig)
	fmt.Printf("EOBBMA:      %.6f (evals: %d)\n",
		eobbmaResult2.GlobalBest.Cost, eobbmaResult2.FuncEvalCount)

	improvement2 := ((standardResult2.GlobalBest.Cost - eobbmaResult2.GlobalBest.Cost) / standardResult2.GlobalBest.Cost) * 100
	fmt.Printf("Improvement: %.2f%%\n", improvement2)

	// ====================================================================
	// Test on Rastrigin (heavy multimodality)
	// ====================================================================
	fmt.Println("\n\n=== Testing on Rastrigin function ===")
	fmt.Println("(Highly multimodal with many local minima)")

	lowerRastrigin := -5.12
	upperRastrigin := 5.12

	// Standard MA on Rastrigin
	standardConfig.ObjectiveFunc = mayfly.Rastrigin
	standardConfig.LowerBound = lowerRastrigin
	standardConfig.UpperBound = upperRastrigin
	standardResult3, _ := mayfly.Optimize(standardConfig)
	fmt.Printf("Standard MA: %.6f (evals: %d)\n",
		standardResult3.GlobalBest.Cost, standardResult3.FuncEvalCount)

	// EOBBMA on Rastrigin
	eobbmaConfig.ObjectiveFunc = mayfly.Rastrigin
	eobbmaConfig.LowerBound = lowerRastrigin
	eobbmaConfig.UpperBound = upperRastrigin
	eobbmaResult3, _ := mayfly.Optimize(eobbmaConfig)
	fmt.Printf("EOBBMA:      %.6f (evals: %d)\n",
		eobbmaResult3.GlobalBest.Cost, eobbmaResult3.FuncEvalCount)

	improvement3 := ((standardResult3.GlobalBest.Cost - eobbmaResult3.GlobalBest.Cost) / standardResult3.GlobalBest.Cost) * 100
	fmt.Printf("Improvement: %.2f%%\n", improvement3)

	// ====================================================================
	// Display EOBBMA Configuration
	// ====================================================================
	fmt.Println("\n\n=== EOBBMA Configuration ===")
	fmt.Printf("Lévy Alpha: %.2f (stability parameter, controls tail heaviness)\n", eobbmaConfig.LevyAlpha)
	fmt.Printf("Lévy Beta: %.2f (scale parameter)\n", eobbmaConfig.LevyBeta)
	fmt.Printf("Opposition Rate: %.2f (probability of applying opposition)\n", eobbmaConfig.OppositionRate)
	fmt.Printf("Elite Opposition Count: %d (top solutions to consider)\n", eobbmaConfig.EliteOppositionCount)
	fmt.Println("\nEOBBMA Features:")
	fmt.Println("  • Bare Bones framework: Gaussian sampling replaces velocity updates")
	fmt.Println("  • Lévy flight: Heavy-tailed jumps for global exploration")
	fmt.Println("  • Elite opposition: Generates opposite solutions to expand search")
	fmt.Println("  • Reduces parameters: No velocity limits or inertia weights")
	fmt.Println("\nTypical performance:")
	fmt.Println("  • Excellent on complex, multimodal landscapes")
	fmt.Println("  • Better exploration-exploitation balance")
	fmt.Println("  • Good for problems with deceptive local optima")
	fmt.Println("  • Comparable or better than standard MA with fewer parameters")
}
