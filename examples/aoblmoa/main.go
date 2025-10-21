package main

import (
	"fmt"
	"math/rand"

	"github.com/cwbudde/mayfly"
)

func main() {
	fmt.Println("================================================================================")
	fmt.Println("AOBLMOA Variant - Aquila Optimizer Hybrid Mayfly Algorithm")
	fmt.Println("================================================================================")
	fmt.Println()

	// Test parameters
	problemSize := 10
	maxIterations := 500
	seed := int64(42)

	// Benchmark functions to test
	testFunctions := []struct {
		name        string
		function    mayfly.ObjectiveFunction
		lowerBound  float64
		upperBound  float64
		optimalCost float64
	}{
		{"Rastrigin (Multimodal)", mayfly.Rastrigin, -5.12, 5.12, 0},
		{"Schwefel (Deceptive)", mayfly.Schwefel, -500, 500, 0},
		{"Ackley (Multimodal)", mayfly.Ackley, -32.768, 32.768, 0},
		{"Rosenbrock (Valley)", mayfly.Rosenbrock, -5, 10, 0},
	}

	fmt.Println("Testing AOBLMOA vs Standard MA on multiple benchmark functions")
	fmt.Println("Problem Size:", problemSize, "| Max Iterations:", maxIterations)
	fmt.Println()

	// Run tests
	for _, test := range testFunctions {
		fmt.Println("================================================================================")
		fmt.Println("Testing:", test.name)
		fmt.Println("================================================================================")

		// Run Standard MA
		configStd := mayfly.NewDefaultConfig()
		configStd.ObjectiveFunc = test.function
		configStd.ProblemSize = problemSize
		configStd.LowerBound = test.lowerBound
		configStd.UpperBound = test.upperBound
		configStd.MaxIterations = maxIterations
		configStd.Rand = rand.New(rand.NewSource(seed))

		resultStd, err := mayfly.Optimize(configStd)
		if err != nil {
			panic(err)
		}

		// Run AOBLMOA
		configAOBLMOA := mayfly.NewAOBLMOAConfig()
		configAOBLMOA.ObjectiveFunc = test.function
		configAOBLMOA.ProblemSize = problemSize
		configAOBLMOA.LowerBound = test.lowerBound
		configAOBLMOA.UpperBound = test.upperBound
		configAOBLMOA.MaxIterations = maxIterations
		configAOBLMOA.Rand = rand.New(rand.NewSource(seed))

		resultAOBLMOA, err := mayfly.Optimize(configAOBLMOA)
		if err != nil {
			panic(err)
		}

		// Calculate improvement
		improvement := ((resultStd.GlobalBest.Cost - resultAOBLMOA.GlobalBest.Cost) / resultStd.GlobalBest.Cost) * 100

		// Print results
		fmt.Printf("\nResults:\n")
		fmt.Printf("  Standard MA:\n")
		fmt.Printf("    Best Cost:        %.6e\n", resultStd.GlobalBest.Cost)
		fmt.Printf("    Function Evals:   %d\n", resultStd.FuncEvalCount)
		fmt.Printf("    Error from Optimum: %.6e\n", resultStd.GlobalBest.Cost-test.optimalCost)
		fmt.Printf("\n")
		fmt.Printf("  AOBLMOA:\n")
		fmt.Printf("    Best Cost:        %.6e\n", resultAOBLMOA.GlobalBest.Cost)
		fmt.Printf("    Function Evals:   %d\n", resultAOBLMOA.FuncEvalCount)
		fmt.Printf("    Error from Optimum: %.6e\n", resultAOBLMOA.GlobalBest.Cost-test.optimalCost)
		fmt.Printf("\n")

		if improvement > 0 {
			fmt.Printf("  Improvement:      +%.2f%% (AOBLMOA is better)\n", improvement)
		} else {
			fmt.Printf("  Improvement:      %.2f%% (Standard MA is better)\n", improvement)
		}

		overhead := float64(resultAOBLMOA.FuncEvalCount-resultStd.FuncEvalCount) / float64(resultStd.FuncEvalCount) * 100
		fmt.Printf("  Overhead:         +%.1f%% function evaluations\n", overhead)
		fmt.Println()
	}

	// Demonstrate parameter tuning
	fmt.Println("================================================================================")
	fmt.Println("AOBLMOA Parameter Tuning Examples")
	fmt.Println("================================================================================")
	fmt.Println()

	// Example 1: Balanced hybrid (default)
	fmt.Println("Example 1: Balanced Hybrid Configuration (Default)")
	config1 := mayfly.NewAOBLMOAConfig()
	config1.ObjectiveFunc = mayfly.Rastrigin
	config1.ProblemSize = 10
	config1.LowerBound = -5.12
	config1.UpperBound = 5.12
	config1.MaxIterations = 500
	// AquilaWeight = 0.5 (default)
	// OppositionProbability = 0.3 (default)
	fmt.Printf("  AquilaWeight:          %.1f (50%% Aquila, 50%% Mayfly)\n", config1.AquilaWeight)
	fmt.Printf("  OppositionProbability: %.1f (30%% use OBL)\n", config1.OppositionProbability)
	fmt.Printf("  ArchiveSize:           %d (Pareto archive capacity)\n", config1.ArchiveSize)
	fmt.Println("  Use when: General problems, unsure which strategy is best")
	fmt.Println()

	// Example 2: More Aquila (aggressive exploration)
	fmt.Println("Example 2: Aggressive Exploration Configuration")
	config2 := mayfly.NewAOBLMOAConfig()
	config2.ObjectiveFunc = mayfly.Schwefel
	config2.ProblemSize = 10
	config2.LowerBound = -500
	config2.UpperBound = 500
	config2.MaxIterations = 500
	config2.AquilaWeight = 0.7          // 70% Aquila
	config2.OppositionProbability = 0.5 // 50% OBL
	fmt.Printf("  AquilaWeight:          %.1f (70%% Aquila, 30%% Mayfly)\n", config2.AquilaWeight)
	fmt.Printf("  OppositionProbability: %.1f (50%% use OBL)\n", config2.OppositionProbability)
	fmt.Println("  Use when: Highly deceptive problems, many local optima")
	fmt.Println()

	// Example 3: More Mayfly (social learning)
	fmt.Println("Example 3: Social Learning Configuration")
	config3 := mayfly.NewAOBLMOAConfig()
	config3.ObjectiveFunc = mayfly.Rosenbrock
	config3.ProblemSize = 10
	config3.LowerBound = -5
	config3.UpperBound = 10
	config3.MaxIterations = 500
	config3.AquilaWeight = 0.3          // 30% Aquila
	config3.OppositionProbability = 0.1 // 10% OBL
	fmt.Printf("  AquilaWeight:          %.1f (30%% Aquila, 70%% Mayfly)\n", config3.AquilaWeight)
	fmt.Printf("  OppositionProbability: %.1f (10%% use OBL)\n", config3.OppositionProbability)
	fmt.Println("  Use when: Problems benefit from swarm intelligence, smooth landscapes")
	fmt.Println()

	// Example 4: Multi-objective setup
	fmt.Println("Example 4: Multi-Objective Optimization Setup")
	config4 := mayfly.NewAOBLMOAConfig()
	// For multi-objective, would need MultiObjectiveFunction interface
	// For now, demonstrate weighted sum approach
	config4.ObjectiveFunc = func(x []float64) float64 {
		obj1 := mayfly.Sphere(x)     // Objective 1: minimize distance
		obj2 := mayfly.Rosenbrock(x) // Objective 2: minimize valley
		return 0.5*obj1 + 0.5*obj2   // Weighted combination
	}
	config4.ProblemSize = 10
	config4.LowerBound = -5
	config4.UpperBound = 10
	config4.MaxIterations = 500
	config4.ArchiveSize = 100 // Store up to 100 Pareto-optimal solutions
	fmt.Printf("  ArchiveSize:           %d (Pareto front capacity)\n", config4.ArchiveSize)
	fmt.Println("  Use when: Multiple conflicting objectives need optimization")
	fmt.Println("  Note: Pareto archive maintains non-dominated solutions internally")
	fmt.Println()

	// Demonstrate adaptive strategy switching
	fmt.Println("================================================================================")
	fmt.Println("AOBLMOA Adaptive Strategy Switching")
	fmt.Println("================================================================================")
	fmt.Println()
	fmt.Println("AOBLMOA uses four Aquila hunting strategies that adapt based on iteration:")
	fmt.Println()
	fmt.Println("  Phase 1 (0-33% iterations): Expanded Exploration (X1)")
	fmt.Println("    - High soar with vertical stoop")
	fmt.Println("    - Global search across entire space")
	fmt.Println("    - Uses population mean for guidance")
	fmt.Println()
	fmt.Println("  Phase 2 (33-66% iterations): Narrowed Exploration (X2)")
	fmt.Println("    - Contour flight with short glide")
	fmt.Println("    - Focused exploration with Lévy flight")
	fmt.Println("    - Combines heavy-tailed jumps with local search")
	fmt.Println()
	fmt.Println("  Phase 3 (66-90% iterations): Expanded Exploitation (X3)")
	fmt.Println("    - Low flight with slow descent")
	fmt.Println("    - Convergence to promising regions")
	fmt.Println("    - Balanced convergence with exploration")
	fmt.Println()
	fmt.Println("  Phase 4 (90-100% iterations): Narrowed Exploitation (X4)")
	fmt.Println("    - Walk and grab prey")
	fmt.Println("    - Intensive local search")
	fmt.Println("    - Fine-tunes solutions with quality function")
	fmt.Println()
	fmt.Println("Strategy switching is automatic and requires no manual tuning!")
	fmt.Println()

	fmt.Println("================================================================================")
	fmt.Println("AOBLMOA Features Summary")
	fmt.Println("================================================================================")
	fmt.Println()
	fmt.Println("  ✓ Four adaptive hunting strategies (Aquila Optimizer)")
	fmt.Println("  ✓ Hybrid Mayfly-Aquila operator switching")
	fmt.Println("  ✓ Opposition-based learning for expanded coverage")
	fmt.Println("  ✓ Multi-objective optimization with Pareto dominance")
	fmt.Println("  ✓ NSGA-II selection with crowding distance")
	fmt.Println("  ✓ Automatic strategy switching (exploration → exploitation)")
	fmt.Println("  ✓ Configurable hybrid balance (AquilaWeight parameter)")
	fmt.Println("  ✓ Archive management for Pareto-optimal solutions")
	fmt.Println()
	fmt.Println("Best for:")
	fmt.Println("  • Complex multi-modal problems with varying landscapes")
	fmt.Println("  • Multi-objective optimization with conflicting criteria")
	fmt.Println("  • Problems requiring adaptive exploration-exploitation")
	fmt.Println("  • Engineering design with multiple performance objectives")
	fmt.Println()
}
