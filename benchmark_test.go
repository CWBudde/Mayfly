package mayfly

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"testing"
)

// BenchmarkProblem defines a benchmark optimization problem.
type BenchmarkProblem struct {
	Func        ObjectiveFunction
	Name        string
	Dimensions  int
	LowerBound  float64
	UpperBound  float64
	GlobalOptim float64
}

// Standard benchmark problems suite.
var benchmarkProblems = []BenchmarkProblem{
	{
		Name:        "Sphere_10D",
		Func:        Sphere,
		Dimensions:  10,
		LowerBound:  -10,
		UpperBound:  10,
		GlobalOptim: 0,
	},
	{
		Name:        "Sphere_30D",
		Func:        Sphere,
		Dimensions:  30,
		LowerBound:  -10,
		UpperBound:  10,
		GlobalOptim: 0,
	},
	{
		Name:        "Rastrigin_10D",
		Func:        Rastrigin,
		Dimensions:  10,
		LowerBound:  -5.12,
		UpperBound:  5.12,
		GlobalOptim: 0,
	},
	{
		Name:        "Rastrigin_30D",
		Func:        Rastrigin,
		Dimensions:  30,
		LowerBound:  -5.12,
		UpperBound:  5.12,
		GlobalOptim: 0,
	},
	{
		Name:        "Rosenbrock_10D",
		Func:        Rosenbrock,
		Dimensions:  10,
		LowerBound:  -5,
		UpperBound:  10,
		GlobalOptim: 0,
	},
	{
		Name:        "Rosenbrock_30D",
		Func:        Rosenbrock,
		Dimensions:  30,
		LowerBound:  -5,
		UpperBound:  10,
		GlobalOptim: 0,
	},
	{
		Name:        "Ackley_10D",
		Func:        Ackley,
		Dimensions:  10,
		LowerBound:  -32.768,
		UpperBound:  32.768,
		GlobalOptim: 0,
	},
	{
		Name:        "Ackley_30D",
		Func:        Ackley,
		Dimensions:  30,
		LowerBound:  -32.768,
		UpperBound:  32.768,
		GlobalOptim: 0,
	},
	{
		Name:        "Griewank_10D",
		Func:        Griewank,
		Dimensions:  10,
		LowerBound:  -600,
		UpperBound:  600,
		GlobalOptim: 0,
	},
	{
		Name:        "Griewank_30D",
		Func:        Griewank,
		Dimensions:  30,
		LowerBound:  -600,
		UpperBound:  600,
		GlobalOptim: 0,
	},
	// CEC-style functions
	{
		Name:        "Schwefel_10D",
		Func:        Schwefel,
		Dimensions:  10,
		LowerBound:  -500,
		UpperBound:  500,
		GlobalOptim: 0,
	},
	{
		Name:        "Levy_10D",
		Func:        Levy,
		Dimensions:  10,
		LowerBound:  -10,
		UpperBound:  10,
		GlobalOptim: 0,
	},
	{
		Name:        "Zakharov_10D",
		Func:        Zakharov,
		Dimensions:  10,
		LowerBound:  -10,
		UpperBound:  10,
		GlobalOptim: 0,
	},
	{
		Name:        "BentCigar_10D",
		Func:        BentCigar,
		Dimensions:  10,
		LowerBound:  -100,
		UpperBound:  100,
		GlobalOptim: 0,
	},
	{
		Name:        "Discus_10D",
		Func:        Discus,
		Dimensions:  10,
		LowerBound:  -100,
		UpperBound:  100,
		GlobalOptim: 0,
	},
	{
		Name:        "HappyCat_10D",
		Func:        HappyCat,
		Dimensions:  10,
		LowerBound:  -2,
		UpperBound:  2,
		GlobalOptim: 0,
	},
}

// BenchmarkResult holds the results of a single benchmark run.
type BenchmarkResult struct {
	Problem      string
	Algorithm    string
	BestCost     float64
	WorstCost    float64
	MeanCost     float64
	StdDevCost   float64
	MedianCost   float64
	FuncEvals    int
	Iterations   int
	SuccessRate  float64 // Percentage of runs reaching near-optimal
	ErrorFromOpt float64 // Mean error from global optimum
}

// Statistics holds statistical measures for multiple runs.
type Statistics struct {
	Mean   float64
	StdDev float64
	Median float64
	Min    float64
	Max    float64
}

// calculateStatistics computes statistics from a slice of values.
func calculateStatistics(values []float64) Statistics {
	if len(values) == 0 {
		return Statistics{}
	}

	// Mean
	sum := 0.0
	for _, v := range values {
		sum += v
	}

	mean := sum / float64(len(values))

	// Standard deviation
	variance := 0.0

	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}

	stdDev := math.Sqrt(variance / float64(len(values)))

	// Median (sort a copy)
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	var median float64

	n := len(sorted)
	if n%2 == 0 {
		median = (sorted[n/2-1] + sorted[n/2]) / 2
	} else {
		median = sorted[n/2]
	}

	return Statistics{
		Mean:   mean,
		StdDev: stdDev,
		Median: median,
		Min:    sorted[0],
		Max:    sorted[n-1],
	}
}

// runBenchmarkSuite runs a benchmark problem multiple times and returns statistics.
func runBenchmarkSuite(problem BenchmarkProblem, useDESMA bool, runs int, iterations int) BenchmarkResult {
	costs := make([]float64, runs)
	funcEvals := 0
	successCount := 0
	threshold := 1e-2 // Consider success if within 0.01 of global optimum

	for i := 0; i < runs; i++ {
		var config *Config
		if useDESMA {
			config = NewDESMAConfig()
		} else {
			config = NewDefaultConfig()
		}

		// Set problem-specific parameters
		config.ObjectiveFunc = problem.Func
		config.ProblemSize = problem.Dimensions
		config.LowerBound = problem.LowerBound
		config.UpperBound = problem.UpperBound
		config.MaxIterations = iterations

		// Use different seed for each run
		config.Rand = rand.New(rand.NewSource(int64(i + 1)))

		result, err := Optimize(config)
		if err != nil {
			panic(fmt.Sprintf("Optimization failed: %v", err))
		}

		costs[i] = result.GlobalBest.Cost
		funcEvals += result.FuncEvalCount

		// Check if this run was successful
		if math.Abs(result.GlobalBest.Cost-problem.GlobalOptim) <= threshold {
			successCount++
		}
	}

	stats := calculateStatistics(costs)

	algorithmName := "Standard MA"
	if useDESMA {
		algorithmName = "DESMA"
	}

	return BenchmarkResult{
		Problem:      problem.Name,
		Algorithm:    algorithmName,
		BestCost:     stats.Min,
		WorstCost:    stats.Max,
		MeanCost:     stats.Mean,
		StdDevCost:   stats.StdDev,
		MedianCost:   stats.Median,
		FuncEvals:    funcEvals / runs,
		Iterations:   iterations,
		SuccessRate:  float64(successCount) / float64(runs) * 100,
		ErrorFromOpt: math.Abs(stats.Mean - problem.GlobalOptim),
	}
}

// BenchmarkOptimizeSphere_StandardMA benchmarks Standard MA on Sphere function.
func BenchmarkOptimizeSphere_StandardMA(b *testing.B) {
	config := NewDefaultConfig()
	config.ObjectiveFunc = Sphere
	config.ProblemSize = 30
	config.LowerBound = -10
	config.UpperBound = 10
	config.MaxIterations = 100
	config.Rand = rand.New(rand.NewSource(42))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = Optimize(config)
	}
}

// BenchmarkOptimizeSphere_DESMA benchmarks DESMA on Sphere function.
func BenchmarkOptimizeSphere_DESMA(b *testing.B) {
	config := NewDESMAConfig()
	config.ObjectiveFunc = Sphere
	config.ProblemSize = 30
	config.LowerBound = -10
	config.UpperBound = 10
	config.MaxIterations = 100
	config.Rand = rand.New(rand.NewSource(42))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = Optimize(config)
	}
}

// BenchmarkOptimizeRastrigin_StandardMA benchmarks Standard MA on Rastrigin function.
func BenchmarkOptimizeRastrigin_StandardMA(b *testing.B) {
	config := NewDefaultConfig()
	config.ObjectiveFunc = Rastrigin
	config.ProblemSize = 30
	config.LowerBound = -5.12
	config.UpperBound = 5.12
	config.MaxIterations = 100
	config.Rand = rand.New(rand.NewSource(42))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = Optimize(config)
	}
}

// BenchmarkOptimizeRastrigin_DESMA benchmarks DESMA on Rastrigin function.
func BenchmarkOptimizeRastrigin_DESMA(b *testing.B) {
	config := NewDESMAConfig()
	config.ObjectiveFunc = Rastrigin
	config.ProblemSize = 30
	config.LowerBound = -5.12
	config.UpperBound = 5.12
	config.MaxIterations = 100
	config.Rand = rand.New(rand.NewSource(42))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = Optimize(config)
	}
}

// BenchmarkOptimizeRosenbrock_StandardMA benchmarks Standard MA on Rosenbrock function.
func BenchmarkOptimizeRosenbrock_StandardMA(b *testing.B) {
	config := NewDefaultConfig()
	config.ObjectiveFunc = Rosenbrock
	config.ProblemSize = 30
	config.LowerBound = -5
	config.UpperBound = 10
	config.MaxIterations = 100
	config.Rand = rand.New(rand.NewSource(42))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = Optimize(config)
	}
}

// BenchmarkOptimizeRosenbrock_DESMA benchmarks DESMA on Rosenbrock function.
func BenchmarkOptimizeRosenbrock_DESMA(b *testing.B) {
	config := NewDESMAConfig()
	config.ObjectiveFunc = Rosenbrock
	config.ProblemSize = 30
	config.LowerBound = -5
	config.UpperBound = 10
	config.MaxIterations = 100
	config.Rand = rand.New(rand.NewSource(42))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = Optimize(config)
	}
}

// BenchmarkOptimizeAckley_StandardMA benchmarks Standard MA on Ackley function.
func BenchmarkOptimizeAckley_StandardMA(b *testing.B) {
	config := NewDefaultConfig()
	config.ObjectiveFunc = Ackley
	config.ProblemSize = 30
	config.LowerBound = -32.768
	config.UpperBound = 32.768
	config.MaxIterations = 100
	config.Rand = rand.New(rand.NewSource(42))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = Optimize(config)
	}
}

// BenchmarkOptimizeAckley_DESMA benchmarks DESMA on Ackley function.
func BenchmarkOptimizeAckley_DESMA(b *testing.B) {
	config := NewDESMAConfig()
	config.ObjectiveFunc = Ackley
	config.ProblemSize = 30
	config.LowerBound = -32.768
	config.UpperBound = 32.768
	config.MaxIterations = 100
	config.Rand = rand.New(rand.NewSource(42))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = Optimize(config)
	}
}

// BenchmarkOptimizeGriewank_StandardMA benchmarks Standard MA on Griewank function.
func BenchmarkOptimizeGriewank_StandardMA(b *testing.B) {
	config := NewDefaultConfig()
	config.ObjectiveFunc = Griewank
	config.ProblemSize = 30
	config.LowerBound = -600
	config.UpperBound = 600
	config.MaxIterations = 100
	config.Rand = rand.New(rand.NewSource(42))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = Optimize(config)
	}
}

// BenchmarkOptimizeGriewank_DESMA benchmarks DESMA on Griewank function.
func BenchmarkOptimizeGriewank_DESMA(b *testing.B) {
	config := NewDESMAConfig()
	config.ObjectiveFunc = Griewank
	config.ProblemSize = 30
	config.LowerBound = -600
	config.UpperBound = 600
	config.MaxIterations = 100
	config.Rand = rand.New(rand.NewSource(42))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = Optimize(config)
	}
}

// BenchmarkDimensionScaling tests how the algorithm scales with problem dimension.
func BenchmarkDimensionScaling(b *testing.B) {
	dimensions := []int{5, 10, 20, 30, 50}

	for _, dim := range dimensions {
		b.Run(fmt.Sprintf("Sphere_%dD", dim), func(b *testing.B) {
			config := NewDefaultConfig()
			config.ObjectiveFunc = Sphere
			config.ProblemSize = dim
			config.LowerBound = -10
			config.UpperBound = 10
			config.MaxIterations = 50
			config.Rand = rand.New(rand.NewSource(42))

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, _ = Optimize(config)
			}
		})
	}
}

// BenchmarkPopulationSize tests different population sizes.
func BenchmarkPopulationSize(b *testing.B) {
	populations := []int{10, 20, 40, 60}

	for _, pop := range populations {
		b.Run(fmt.Sprintf("Pop_%d", pop), func(b *testing.B) {
			config := NewDefaultConfig()
			config.ObjectiveFunc = Sphere
			config.ProblemSize = 30
			config.LowerBound = -10
			config.UpperBound = 10
			config.MaxIterations = 50
			config.NPop = pop
			config.NPopF = pop
			config.NC = pop / 2
			config.Rand = rand.New(rand.NewSource(42))

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, _ = Optimize(config)
			}
		})
	}
}

// TestBenchmarkSuite runs comprehensive benchmark suite with statistical analysis.
// This is a test function that generates a performance report.
func TestBenchmarkSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping benchmark suite in short mode")
	}

	runs := 30        // Number of independent runs for statistical significance
	iterations := 500 // Number of iterations per run

	t.Logf("\n" + strings.Repeat("=", 80))
	t.Logf("MAYFLY ALGORITHM BENCHMARK SUITE")
	t.Log(strings.Repeat("=", 80))
	t.Logf("Runs per problem: %d", runs)
	t.Logf("Iterations per run: %d", iterations)
	t.Logf(strings.Repeat("=", 80) + "\n")

	// Test a subset of problems for reasonable test time
	testProblems := []BenchmarkProblem{
		benchmarkProblems[0], // Sphere_10D
		benchmarkProblems[2], // Rastrigin_10D
		benchmarkProblems[4], // Rosenbrock_10D
		benchmarkProblems[6], // Ackley_10D
		benchmarkProblems[8], // Griewank_10D
	}

	for _, problem := range testProblems {
		t.Run(problem.Name, func(t *testing.T) {
			// Run Standard MA
			t.Logf("\nTesting %s with Standard MA...", problem.Name)
			resultMA := runBenchmarkSuite(problem, false, runs, iterations)

			// Run DESMA
			t.Logf("Testing %s with DESMA...", problem.Name)
			resultDESMA := runBenchmarkSuite(problem, true, runs, iterations)

			// Print comparison
			t.Logf("\n%s - BENCHMARK RESULTS", problem.Name)
			t.Log(strings.Repeat("-", 80))
			t.Logf("%-20s | %-15s | %-15s | Improvement", "Metric", "Standard MA", "DESMA")
			t.Log(strings.Repeat("-", 80))

			improvement := (resultMA.MeanCost - resultDESMA.MeanCost) / resultMA.MeanCost * 100
			t.Logf("%-20s | %15.6e | %15.6e | %+7.2f%%",
				"Mean Cost", resultMA.MeanCost, resultDESMA.MeanCost, improvement)

			t.Logf("%-20s | %15.6e | %15.6e |",
				"Std Dev", resultMA.StdDevCost, resultDESMA.StdDevCost)

			t.Logf("%-20s | %15.6e | %15.6e |",
				"Best Cost", resultMA.BestCost, resultDESMA.BestCost)

			t.Logf("%-20s | %15.6e | %15.6e |",
				"Worst Cost", resultMA.WorstCost, resultDESMA.WorstCost)

			t.Logf("%-20s | %15.6e | %15.6e |",
				"Median Cost", resultMA.MedianCost, resultDESMA.MedianCost)

			t.Logf("%-20s | %15d | %15d |",
				"Avg Func Evals", resultMA.FuncEvals, resultDESMA.FuncEvals)

			t.Logf("%-20s | %14.1f%% | %14.1f%% |",
				"Success Rate", resultMA.SuccessRate, resultDESMA.SuccessRate)

			t.Logf("%-20s | %15.6e | %15.6e |",
				"Error from Optim", resultMA.ErrorFromOpt, resultDESMA.ErrorFromOpt)

			t.Logf(strings.Repeat("-", 80) + "\n")

			// Verify both algorithms converge reasonably
			maxAcceptableError := 100.0 // Adjust based on problem difficulty
			if problem.Name == "Rastrigin_10D" {
				maxAcceptableError = 500.0 // Rastrigin is harder
			}

			if resultMA.MeanCost > maxAcceptableError {
				t.Logf("Warning: Standard MA mean cost %.2e exceeds threshold %.2e",
					resultMA.MeanCost, maxAcceptableError)
			}

			if resultDESMA.MeanCost > maxAcceptableError {
				t.Logf("Warning: DESMA mean cost %.2e exceeds threshold %.2e",
					resultDESMA.MeanCost, maxAcceptableError)
			}
		})
	}

	t.Logf("\n" + strings.Repeat("=", 80))
	t.Logf("BENCHMARK SUITE COMPLETED")
	t.Logf(strings.Repeat("=", 80) + "\n")
}

// TestBenchmarkSuiteQuick runs a quick benchmark suite with fewer runs.
func TestBenchmarkSuiteQuick(t *testing.T) {
	runs := 10
	iterations := 100

	t.Logf("\nQUICK BENCHMARK SUITE (runs=%d, iterations=%d)\n", runs, iterations)

	problem := BenchmarkProblem{
		Name:        "Sphere_10D_Quick",
		Func:        Sphere,
		Dimensions:  10,
		LowerBound:  -10,
		UpperBound:  10,
		GlobalOptim: 0,
	}

	resultMA := runBenchmarkSuite(problem, false, runs, iterations)
	resultDESMA := runBenchmarkSuite(problem, true, runs, iterations)

	t.Logf("Standard MA - Mean: %.6e, StdDev: %.6e, Best: %.6e",
		resultMA.MeanCost, resultMA.StdDevCost, resultMA.BestCost)
	t.Logf("DESMA       - Mean: %.6e, StdDev: %.6e, Best: %.6e",
		resultDESMA.MeanCost, resultDESMA.StdDevCost, resultDESMA.BestCost)

	improvement := (resultMA.MeanCost - resultDESMA.MeanCost) / resultMA.MeanCost * 100
	t.Logf("DESMA Improvement: %+.2f%%\n", improvement)

	// Verify both found reasonable solutions
	threshold := 1.0
	if resultMA.BestCost > threshold {
		t.Errorf("Standard MA best cost %.6e exceeds threshold %.2f", resultMA.BestCost, threshold)
	}

	if resultDESMA.BestCost > threshold {
		t.Errorf("DESMA best cost %.6e exceeds threshold %.2f", resultDESMA.BestCost, threshold)
	}
}

// TestBenchmarkConvergence tests convergence behavior over iterations.
func TestBenchmarkConvergence(t *testing.T) {
	config := NewDefaultConfig()
	config.ObjectiveFunc = Sphere
	config.ProblemSize = 10
	config.LowerBound = -10
	config.UpperBound = 10
	config.MaxIterations = 100
	config.Rand = rand.New(rand.NewSource(42))

	result, err := Optimize(config)
	if err != nil {
		t.Fatalf("Optimization failed: %v", err)
	}

	// Verify convergence: best solution should improve or stay same
	for i := 1; i < len(result.BestSolution); i++ {
		if result.BestSolution[i] > result.BestSolution[i-1]+1e-10 {
			t.Errorf("Convergence violated at iteration %d: %.6e > %.6e",
				i, result.BestSolution[i], result.BestSolution[i-1])
		}
	}

	// Verify final solution is reasonably good
	finalCost := result.GlobalBest.Cost
	if finalCost > 1e-2 {
		t.Logf("Warning: Final cost %.6e is higher than expected for Sphere function", finalCost)
	}

	t.Logf("Convergence test passed. Final cost: %.6e", finalCost)
}

// TestBenchmarkReproducibility tests that results are reproducible with same seed.
func TestBenchmarkReproducibility(t *testing.T) {
	runOptimization := func(seed int64) float64 {
		config := NewDefaultConfig()
		config.ObjectiveFunc = Sphere
		config.ProblemSize = 10
		config.LowerBound = -10
		config.UpperBound = 10
		config.MaxIterations = 50
		config.Rand = rand.New(rand.NewSource(seed))

		result, err := Optimize(config)
		if err != nil {
			t.Fatalf("Optimization failed: %v", err)
		}

		return result.GlobalBest.Cost
	}

	// Run with same seed twice
	cost1 := runOptimization(42)
	cost2 := runOptimization(42)

	if math.Abs(cost1-cost2) > 1e-10 {
		t.Errorf("Results not reproducible: run1=%.10e, run2=%.10e", cost1, cost2)
	}

	// Run with different seed should give different result (very likely)
	cost3 := runOptimization(123)
	if math.Abs(cost1-cost3) < 1e-10 {
		t.Logf("Warning: Different seeds gave same result (%.10e). This is very unlikely but possible.", cost1)
	}

	t.Logf("Reproducibility test passed. Same seed: %.6e = %.6e, Different seed: %.6e",
		cost1, cost2, cost3)
}
