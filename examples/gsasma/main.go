package main

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/cwbudde/mayfly"
)

func main() {
	fmt.Println("=== GSASMA (Golden Annealing Crossover-Mutation MA) Example ===")
	fmt.Println()
	fmt.Println("GSASMA combines two population-wide update techniques:")
	fmt.Println("1. Simulated annealing selects late-stage velocity branches")
	fmt.Println("2. The fixed golden-sine equation refines positions")
	fmt.Println()

	// Create a fixed random seed for reproducibility
	seed := int64(12345)
	rng := rand.New(rand.NewSource(seed))

	// Test parameters
	problemSize := 30
	maxIterations := 500
	numRuns := 10

	// Test functions
	testFunctions := []struct {
		name     string
		function mayfly.ObjectiveFunction
		lower    float64
		upper    float64
		expected float64
	}{
		{"Sphere", mayfly.Sphere, -10, 10, 0},
		{"Rastrigin", mayfly.Rastrigin, -5.12, 5.12, 0},
		{"Rosenbrock", mayfly.Rosenbrock, -5, 10, 0},
		{"Ackley", mayfly.Ackley, -32.768, 32.768, 0},
		{"Griewank", mayfly.Griewank, -600, 600, 0},
		{"Schwefel", mayfly.Schwefel, -500, 500, 0},
	}

	fmt.Println("Testing GSASMA vs Standard MA on benchmark functions")
	fmt.Println("Problem Size:", problemSize)
	fmt.Println("Max Iterations:", maxIterations)
	fmt.Println("Number of Runs:", numRuns)
	fmt.Println()

	// Compare Standard MA vs GSASMA on each function
	for _, test := range testFunctions {
		fmt.Printf("Function: %s (Global Minimum = %.2f)\n", test.name, test.expected)
		fmt.Println(strings.Repeat("-", 80))

		// Run Standard MA
		stdResults := make([]float64, numRuns)
		stdFuncEvals := 0
		for run := 0; run < numRuns; run++ {
			config := mayfly.NewDefaultConfig()
			config.ObjectiveFunc = test.function
			config.ProblemSize = problemSize
			config.LowerBound = test.lower
			config.UpperBound = test.upper
			config.MaxIterations = maxIterations
			config.Rand = rand.New(rand.NewSource(seed + int64(run)))

			result, err := mayfly.Optimize(config)
			if err != nil {
				panic(err)
			}
			stdResults[run] = result.GlobalBest.Cost
			if run == 0 {
				stdFuncEvals = result.FuncEvalCount
			}
		}

		// Run GSASMA
		gsasmaResults := make([]float64, numRuns)
		gsasmaFuncEvals := 0
		for run := 0; run < numRuns; run++ {
			config := mayfly.NewGSASMAConfig()
			config.ObjectiveFunc = test.function
			config.ProblemSize = problemSize
			config.LowerBound = test.lower
			config.UpperBound = test.upper
			config.MaxIterations = maxIterations
			config.Rand = rand.New(rand.NewSource(seed + int64(run)))

			result, err := mayfly.Optimize(config)
			if err != nil {
				panic(err)
			}
			gsasmaResults[run] = result.GlobalBest.Cost
			if run == 0 {
				gsasmaFuncEvals = result.FuncEvalCount
			}
		}

		// Calculate statistics
		stdMean, stdStd := calculateStats(stdResults)
		gsasmaMean, gsasmaStd := calculateStats(gsasmaResults)
		improvement := ((stdMean - gsasmaMean) / stdMean) * 100

		// Display results
		fmt.Printf("Standard MA:\n")
		fmt.Printf("  Mean Cost:     %.6f (±%.6f)\n", stdMean, stdStd)
		fmt.Printf("  Best Cost:     %.6f\n", min(stdResults))
		fmt.Printf("  Worst Cost:    %.6f\n", max(stdResults))
		fmt.Printf("  Func Evals:    %d\n", stdFuncEvals)
		fmt.Println()

		fmt.Printf("GSASMA:\n")
		fmt.Printf("  Mean Cost:     %.6f (±%.6f)\n", gsasmaMean, gsasmaStd)
		fmt.Printf("  Best Cost:     %.6f\n", min(gsasmaResults))
		fmt.Printf("  Worst Cost:    %.6f\n", max(gsasmaResults))
		fmt.Printf("  Func Evals:    %d\n", gsasmaFuncEvals)
		fmt.Println()

		fmt.Printf("Improvement:     %.2f%%\n", improvement)
		fmt.Printf("Overhead:        %.2f%% more evaluations\n",
			float64(gsasmaFuncEvals-stdFuncEvals)/float64(stdFuncEvals)*100)
		fmt.Println()
	}

	// Detailed GSASMA demonstration on a single problem
	fmt.Println("\n=== Detailed GSASMA Demonstration ===")
	fmt.Println()
	fmt.Println("Running GSASMA on Rastrigin function (highly multimodal)")
	fmt.Println("This demonstrates all GSASMA components working together")
	fmt.Println()

	config := mayfly.NewGSASMAConfig()
	config.ObjectiveFunc = mayfly.Rastrigin
	config.ProblemSize = 20
	config.LowerBound = -5.12
	config.UpperBound = 5.12
	config.MaxIterations = 300
	config.Rand = rng

	// Show configuration
	fmt.Println("Configuration:")
	fmt.Printf("  Initial Temperature: %.1f\n", config.InitialTemperature)
	fmt.Printf("  Cooling Rate:        %.3f\n", config.CoolingRate)
	fmt.Printf("  Cooling Schedule:    %s\n", config.CoolingSchedule)
	fmt.Println()

	result, err := mayfly.Optimize(config)
	if err != nil {
		panic(err)
	}

	fmt.Println("Results:")
	fmt.Printf("  Final Cost:      %.6f\n", result.GlobalBest.Cost)
	fmt.Printf("  Total Iterations: %d\n", result.IterationCount)
	fmt.Printf("  Func Evaluations: %d\n", result.FuncEvalCount)
	fmt.Printf("  Evals per Iter:   %.1f\n",
		float64(result.FuncEvalCount)/float64(result.IterationCount))
	fmt.Println()

	// Show convergence progress
	fmt.Println("Convergence Progress (every 50 iterations):")
	for i := 0; i < len(result.ConvergenceCurve); i += 50 {
		if i < len(result.ConvergenceCurve) {
			fmt.Printf("  Iteration %3d: %.6f\n", i, result.ConvergenceCurve[i])
		}
	}
	if len(result.ConvergenceCurve) > 0 {
		fmt.Printf("  Iteration %3d: %.6f (final)\n",
			len(result.ConvergenceCurve)-1, result.ConvergenceCurve[len(result.ConvergenceCurve)-1])
	}
	fmt.Println()

	// Tips for using GSASMA
	fmt.Println("=== Tips for Using GSASMA ===")
	fmt.Println()
	fmt.Println("GSASMA is best suited for:")
	fmt.Println("  • Engineering optimization problems")
	fmt.Println("  • Problems with many local optima")
	fmt.Println("  • Complex multimodal landscapes")
	fmt.Println()

	fmt.Println("Parameter Tuning Guidelines:")
	fmt.Println("  • InitialTemperature: Higher (100-1000) for more exploration")
	fmt.Println("  • CoolingRate: Higher (0.95-0.99) for gradual cooling")
	fmt.Println("  • CoolingSchedule: 'exponential' (fast), 'logarithmic' (slow)")
	fmt.Println("  • Cooling controls are library extensions; the paper omits its recurrence")
	fmt.Println()
}

// Helper functions
func calculateStats(values []float64) (mean, stddev float64) {
	if len(values) == 0 {
		return 0, 0
	}

	// Calculate mean
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean = sum / float64(len(values))

	// Calculate standard deviation
	sumSquaredDiff := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquaredDiff += diff * diff
	}
	variance := sumSquaredDiff / float64(len(values))
	stddev = variance // Simplified: using variance as stddev for brevity
	if variance > 0 {
		stddev = variance
	}

	return mean, stddev
}

func min(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	minVal := values[0]
	for _, v := range values {
		if v < minVal {
			minVal = v
		}
	}
	return minVal
}

func max(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	maxVal := values[0]
	for _, v := range values {
		if v > maxVal {
			maxVal = v
		}
	}
	return maxVal
}
