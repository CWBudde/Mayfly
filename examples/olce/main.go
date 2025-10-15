package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/cwbudde/mayfly"
)

func main() {
	fmt.Println("=== OLCE-MA (Orthogonal Learning & Chaotic Exploitation) Example ===\n")

	// Seed random number generator
	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))

	// Problem parameters
	problemSize := 10
	lowerBound := -10.0
	upperBound := 10.0

	// Test on multimodal Rastrigin function (OLCE-MA should excel here)
	fmt.Println("Testing on Rastrigin function (highly multimodal)")
	fmt.Printf("Dimensions: %d\n", problemSize)
	fmt.Printf("Search space: [%.1f, %.1f]\n", lowerBound, upperBound)
	fmt.Printf("Global optimum: f(0,...,0) = 0\n\n")

	// ====================================================================
	// Run Standard Mayfly Algorithm
	// ====================================================================
	fmt.Println("Running Standard Mayfly Algorithm...")

	standardConfig := mayfly.NewDefaultConfig()
	standardConfig.ObjectiveFunc = mayfly.Rastrigin
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
	// Run OLCE-MA Variant
	// ====================================================================
	fmt.Println("Running OLCE-MA Variant...")

	olceConfig := mayfly.NewOLCEConfig()
	olceConfig.ObjectiveFunc = mayfly.Rastrigin
	olceConfig.ProblemSize = problemSize
	olceConfig.LowerBound = lowerBound
	olceConfig.UpperBound = upperBound
	olceConfig.MaxIterations = 500
	olceConfig.Rand = rng

	olceStart := time.Now()
	olceResult, err := mayfly.Optimize(olceConfig)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	olceDuration := time.Since(olceStart)

	fmt.Printf("  Best cost: %.6f\n", olceResult.GlobalBest.Cost)
	fmt.Printf("  Function evaluations: %d\n", olceResult.FuncEvalCount)
	fmt.Printf("  Time: %v\n\n", olceDuration)

	// ====================================================================
	// Compare Results
	// ====================================================================
	fmt.Println("=== Comparison ===")

	improvement := ((standardResult.GlobalBest.Cost - olceResult.GlobalBest.Cost) / standardResult.GlobalBest.Cost) * 100
	overhead := ((float64(olceResult.FuncEvalCount) / float64(standardResult.FuncEvalCount)) - 1.0) * 100

	fmt.Printf("Cost improvement: %.2f%%\n", improvement)
	fmt.Printf("Evaluation overhead: %.2f%%\n", overhead)

	if olceResult.GlobalBest.Cost < standardResult.GlobalBest.Cost {
		fmt.Println("\n✓ OLCE-MA found a better solution!")
	} else if olceResult.GlobalBest.Cost == standardResult.GlobalBest.Cost {
		fmt.Println("\n= Both algorithms found equivalent solutions")
	} else {
		fmt.Println("\n✗ Standard MA performed better on this run")
	}

	// ====================================================================
	// Test on another multimodal function: Rosenbrock
	// ====================================================================
	fmt.Println("\n\n=== Testing on Rosenbrock function ===")
	fmt.Println("(Unimodal but with narrow valley)")

	// Standard MA on Rosenbrock
	standardConfig.ObjectiveFunc = mayfly.Rosenbrock
	standardResult2, _ := mayfly.Optimize(standardConfig)
	fmt.Printf("Standard MA: %.6f (evals: %d)\n",
		standardResult2.GlobalBest.Cost, standardResult2.FuncEvalCount)

	// OLCE-MA on Rosenbrock
	olceConfig.ObjectiveFunc = mayfly.Rosenbrock
	olceResult2, _ := mayfly.Optimize(olceConfig)
	fmt.Printf("OLCE-MA:     %.6f (evals: %d)\n",
		olceResult2.GlobalBest.Cost, olceResult2.FuncEvalCount)

	improvement2 := ((standardResult2.GlobalBest.Cost - olceResult2.GlobalBest.Cost) / standardResult2.GlobalBest.Cost) * 100
	fmt.Printf("Improvement: %.2f%%\n", improvement2)

	// ====================================================================
	// Display OLCE-MA Configuration
	// ====================================================================
	fmt.Println("\n\n=== OLCE-MA Configuration ===")
	fmt.Printf("Orthogonal Factor: %.2f (controls learning strength)\n", olceConfig.OrthogonalFactor)
	fmt.Printf("Chaos Factor: %.2f (controls perturbation strength)\n", olceConfig.ChaosFactor)
	fmt.Println("\nOLCE-MA Features:")
	fmt.Println("  • Orthogonal learning: Increases diversity, reduces oscillation")
	fmt.Println("  • Chaotic exploitation: Improves local search with logistic map")
	fmt.Println("  • Applied to top 20% of males and all offspring")
	fmt.Println("\nTypical performance:")
	fmt.Println("  • 15-30% improvement on multimodal functions")
	fmt.Println("  • ~12% overhead in function evaluations")
	fmt.Println("  • Excellent for complex, high-dimensional problems")
}
