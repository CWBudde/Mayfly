package main

import (
	"fmt"
	"math"

	"github.com/cwbudde/mayfly"
)

func main() {
	fmt.Println("=================================================")
	fmt.Println("   Mayfly vs DESMA Performance Comparison")
	fmt.Println("=================================================\n")

	// Test on Sphere function
	fmt.Println("--- Sphere Function (50D) ---")
	compareAlgorithms("Sphere", mayfly.Sphere, -10, 10, 50, 500)

	fmt.Println("\n--- Rastrigin Function (50D) ---")
	compareAlgorithms("Rastrigin", mayfly.Rastrigin, -10, 10, 50, 500)

	fmt.Println("\n--- Rosenbrock Function (30D) ---")
	compareAlgorithms("Rosenbrock", mayfly.Rosenbrock, -5, 10, 30, 500)

	fmt.Println("\n--- Ackley Function (30D) ---")
	compareAlgorithms("Ackley", mayfly.Ackley, -32.768, 32.768, 30, 500)
}

func compareAlgorithms(name string, fn mayfly.ObjectiveFunction, lower, upper float64, problemSize, iterations int) {
	// Run standard Mayfly Algorithm
	configMA := mayfly.NewDefaultConfig()
	configMA.ObjectiveFunc = fn
	configMA.ProblemSize = problemSize
	configMA.LowerBound = lower
	configMA.UpperBound = upper
	configMA.MaxIterations = iterations

	resultMA, _ := mayfly.Optimize(configMA)

	// Run DESMA variant
	configDESMA := mayfly.NewDESMAConfig()
	configDESMA.ObjectiveFunc = fn
	configDESMA.ProblemSize = problemSize
	configDESMA.LowerBound = lower
	configDESMA.UpperBound = upper
	configDESMA.MaxIterations = iterations

	resultDESMA, _ := mayfly.Optimize(configDESMA)

	// Display comparison
	fmt.Printf("\nAlgorithm          | Best Cost        | Func Evals | Improvement\n")
	fmt.Printf("-------------------|------------------|------------|------------\n")
	fmt.Printf("Standard MA        | %.10f | %6d     | -\n", resultMA.GlobalBest.Cost, resultMA.FuncEvalCount)
	fmt.Printf("DESMA              | %.10f | %6d     | ", resultDESMA.GlobalBest.Cost, resultDESMA.FuncEvalCount)

	if resultDESMA.GlobalBest.Cost < resultMA.GlobalBest.Cost {
		improvement := (resultMA.GlobalBest.Cost - resultDESMA.GlobalBest.Cost) / resultMA.GlobalBest.Cost * 100
		fmt.Printf("%.2f%%\n", improvement)
	} else {
		fmt.Printf("%.2f%%\n", 0.0)
	}

	// Show convergence comparison
	fmt.Println("\nConvergence Comparison:")
	showComparison(resultMA, resultDESMA)
}

func showComparison(resultMA, resultDESMA *mayfly.Result) {
	n := len(resultMA.BestSolution)
	iterations := []int{n / 4, n / 2, 3 * n / 4, n}

	fmt.Printf("\nIteration | Standard MA      | DESMA            | Winner\n")
	fmt.Printf("----------|------------------|------------------|--------\n")

	for _, it := range iterations {
		idx := it - 1
		maVal := resultMA.BestSolution[idx]
		desmaVal := resultDESMA.BestSolution[idx]

		winner := "MA"
		if desmaVal < maVal {
			winner := "DESMA"
			fmt.Printf("%9d | %.10f | %.10f | %s\n", it, maVal, desmaVal, winner)
		} else {
			fmt.Printf("%9d | %.10f | %.10f | %s\n", it, maVal, desmaVal, winner)
		}
	}

	// Calculate convergence rate
	maConvergence := calculateConvergenceRate(resultMA.BestSolution)
	desmaConvergence := calculateConvergenceRate(resultDESMA.BestSolution)

	fmt.Printf("\nConvergence Rate (per 100 iter):\n")
	fmt.Printf("  Standard MA: %.6f\n", maConvergence)
	fmt.Printf("  DESMA:       %.6f\n", desmaConvergence)
}

func calculateConvergenceRate(solution []float64) float64 {
	n := len(solution)
	if n < 2 {
		return 0
	}

	// Calculate average improvement per iteration in first half
	improvements := 0.0
	count := 0
	for i := 1; i < n/2; i++ {
		if solution[i-1] != 0 {
			improvement := (solution[i-1] - solution[i]) / math.Abs(solution[i-1])
			if improvement > 0 {
				improvements += improvement
				count++
			}
		}
	}

	if count == 0 {
		return 0
	}

	return (improvements / float64(count)) * 100
}
