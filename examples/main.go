package main

import (
	"fmt"
	"math"

	"github.com/cwbudde/mayfly"
)

func main() {
	fmt.Println("=== Standard Mayfly Algorithm Examples ===\n")

	// Example 1: Sphere function
	fmt.Println("=== Optimizing Sphere Function ===")
	runOptimization("Sphere", mayfly.Sphere, -10, 10, 50, false)

	fmt.Println("\n=== Optimizing Rastrigin Function ===")
	runOptimization("Rastrigin", mayfly.Rastrigin, -10, 10, 50, false)

	fmt.Println("\n\n=== DESMA (Dynamic Elite Strategy) Examples ===\n")

	fmt.Println("=== Optimizing Sphere Function with DESMA ===")
	runOptimization("Sphere", mayfly.Sphere, -10, 10, 50, true)

	fmt.Println("\n=== Optimizing Rastrigin Function with DESMA ===")
	runOptimization("Rastrigin", mayfly.Rastrigin, -10, 10, 50, true)
}

func runOptimization(name string, fn mayfly.ObjectiveFunction, lower, upper float64, problemSize int, useDESMA bool) {
	// Create configuration
	var config *mayfly.Config
	if useDESMA {
		config = mayfly.NewDESMAConfig()
	} else {
		config = mayfly.NewDefaultConfig()
	}
	config.ObjectiveFunc = fn
	config.ProblemSize = problemSize
	config.LowerBound = lower
	config.UpperBound = upper
	config.MaxIterations = 500 // Reduce for faster demo

	// Run optimization
	result, err := mayfly.Optimize(config)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Print results
	variant := "Standard MA"
	if useDESMA {
		variant = "DESMA"
	}
	fmt.Printf("Algorithm: %s\n", variant)
	fmt.Printf("Function: %s\n", name)
	fmt.Printf("Problem Size: %d dimensions\n", problemSize)
	fmt.Printf("Bounds: [%.2f, %.2f]\n", lower, upper)
	fmt.Printf("Iterations: %d\n", result.IterationCount)
	fmt.Printf("Function Evaluations: %d\n", result.FuncEvalCount)
	fmt.Printf("Best Cost: %.10f\n", result.GlobalBest.Cost)
	fmt.Printf("Best Position (first 5): ")
	for i := 0; i < min(5, len(result.GlobalBest.Position)); i++ {
		fmt.Printf("%.6f ", result.GlobalBest.Position[i])
	}
	fmt.Println()

	// Show convergence statistics
	showConvergence(result)
}

func showConvergence(result *mayfly.Result) {
	// Show improvement over iterations
	n := len(result.BestSolution)
	fmt.Printf("Convergence:\n")
	fmt.Printf("  Iteration 1: %.6f\n", result.BestSolution[0])
	fmt.Printf("  Iteration %d: %.6f\n", n/4, result.BestSolution[n/4-1])
	fmt.Printf("  Iteration %d: %.6f\n", n/2, result.BestSolution[n/2-1])
	fmt.Printf("  Iteration %d: %.6f\n", 3*n/4, result.BestSolution[3*n/4-1])
	fmt.Printf("  Iteration %d: %.6f\n", n, result.BestSolution[n-1])

	// Calculate improvement rate
	initial := result.BestSolution[0]
	final := result.BestSolution[n-1]
	if initial != 0 {
		improvement := (initial - final) / math.Abs(initial) * 100
		fmt.Printf("  Improvement: %.2f%%\n", improvement)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
