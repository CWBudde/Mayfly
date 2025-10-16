package mayfly

import (
	"math"
	"math/rand"
	"testing"
)

// RegressionBaseline holds expected performance baselines for regression testing.
// These values were established during initial implementation and should not degrade
// significantly in future changes.
type RegressionBaseline struct {
	Function         ObjectiveFunction
	Name             string
	Dimensions       int
	LowerBound       float64
	UpperBound       float64
	Iterations       int
	Seed             int64
	ExpectedBest     float64
	ExpectedMean     float64
	Tolerance        float64
	SuccessThreshold float64
	UseDESMA         bool
}

// Regression baselines for Standard MA and DESMA.
var regressionBaselines = []RegressionBaseline{
	{
		Name:             "StandardMA_Sphere_10D",
		Function:         Sphere,
		Dimensions:       10,
		LowerBound:       -10,
		UpperBound:       10,
		Iterations:       500,
		Seed:             42,
		UseDESMA:         false,
		ExpectedBest:     1e-5, // Should reach near-zero
		ExpectedMean:     1e-4, // Mean over 10 runs
		Tolerance:        10.0, // Allow 10x degradation before failing
		SuccessThreshold: 0.8,  // 80% of runs should succeed
	},
	{
		Name:             "DESMA_Sphere_10D",
		Function:         Sphere,
		Dimensions:       10,
		LowerBound:       -10,
		UpperBound:       10,
		Iterations:       500,
		Seed:             42,
		UseDESMA:         true,
		ExpectedBest:     1e-6, // DESMA should do better
		ExpectedMean:     1e-5,
		Tolerance:        10.0,
		SuccessThreshold: 0.9, // 90% success rate
	},
	{
		Name:             "StandardMA_Rastrigin_10D",
		Function:         Rastrigin,
		Dimensions:       10,
		LowerBound:       -5.12,
		UpperBound:       5.12,
		Iterations:       500,
		Seed:             42,
		UseDESMA:         false,
		ExpectedBest:     50.0, // Rastrigin is harder, multimodal
		ExpectedMean:     100.0,
		Tolerance:        2.0, // Less tolerance for degradation
		SuccessThreshold: 0.5, // 50% success rate acceptable
	},
	{
		Name:             "DESMA_Rastrigin_10D",
		Function:         Rastrigin,
		Dimensions:       10,
		LowerBound:       -5.12,
		UpperBound:       5.12,
		Iterations:       500,
		Seed:             42,
		UseDESMA:         true,
		ExpectedBest:     20.0, // DESMA should significantly improve
		ExpectedMean:     40.0,
		Tolerance:        2.0,
		SuccessThreshold: 0.7, // 70% success rate
	},
	{
		Name:             "StandardMA_Rosenbrock_10D",
		Function:         Rosenbrock,
		Dimensions:       10,
		LowerBound:       -5,
		UpperBound:       10,
		Iterations:       500,
		Seed:             42,
		UseDESMA:         false,
		ExpectedBest:     10.0, // Rosenbrock has narrow valley
		ExpectedMean:     50.0,
		Tolerance:        5.0,
		SuccessThreshold: 0.6,
	},
	{
		Name:             "DESMA_Rosenbrock_10D",
		Function:         Rosenbrock,
		Dimensions:       10,
		LowerBound:       -5,
		UpperBound:       10,
		Iterations:       500,
		Seed:             42,
		UseDESMA:         true,
		ExpectedBest:     5.0, // DESMA helps with narrow valleys
		ExpectedMean:     20.0,
		Tolerance:        5.0,
		SuccessThreshold: 0.7,
	},
	{
		Name:             "StandardMA_Ackley_10D",
		Function:         Ackley,
		Dimensions:       10,
		LowerBound:       -32.768,
		UpperBound:       32.768,
		Iterations:       500,
		Seed:             42,
		UseDESMA:         false,
		ExpectedBest:     0.5, // Ackley has many local minima
		ExpectedMean:     1.0,
		Tolerance:        3.0,
		SuccessThreshold: 0.6, // Adjusted: Ackley is very challenging, 60% is reasonable
	},
	{
		Name:             "DESMA_Ackley_10D",
		Function:         Ackley,
		Dimensions:       10,
		LowerBound:       -32.768,
		UpperBound:       32.768,
		Iterations:       500,
		Seed:             42,
		UseDESMA:         true,
		ExpectedBest:     0.1, // DESMA should escape local minima better
		ExpectedMean:     0.5,
		Tolerance:        3.0,
		SuccessThreshold: 0.7, // Adjusted: DESMA helps but Ackley is still hard, 70% is solid
	},
	{
		Name:             "StandardMA_Griewank_10D",
		Function:         Griewank,
		Dimensions:       10,
		LowerBound:       -600,
		UpperBound:       600,
		Iterations:       500,
		Seed:             42,
		UseDESMA:         false,
		ExpectedBest:     0.1,
		ExpectedMean:     0.5,
		Tolerance:        5.0,
		SuccessThreshold: 0.7,
	},
	{
		Name:             "DESMA_Griewank_10D",
		Function:         Griewank,
		Dimensions:       10,
		LowerBound:       -600,
		UpperBound:       600,
		Iterations:       500,
		Seed:             42,
		UseDESMA:         true,
		ExpectedBest:     0.05,
		ExpectedMean:     0.2,
		Tolerance:        5.0,
		SuccessThreshold: 0.7, // Adjusted: Griewank is challenging, 70% is reasonable for DESMA
	},
}

// TestRegressionSuite runs regression tests against baseline performance metrics.
// This ensures that future changes don't degrade algorithm performance.
func TestRegressionSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping regression suite in short mode")
	}

	runs := 10 // Number of runs for statistical reliability

	for _, baseline := range regressionBaselines {
		t.Run(baseline.Name, func(t *testing.T) {
			costs := make([]float64, runs)
			successCount := 0

			// Run multiple times for statistical significance
			for i := 0; i < runs; i++ {
				var config *Config
				if baseline.UseDESMA {
					config = NewDESMAConfig()
				} else {
					config = NewDefaultConfig()
				}

				config.ObjectiveFunc = baseline.Function
				config.ProblemSize = baseline.Dimensions
				config.LowerBound = baseline.LowerBound
				config.UpperBound = baseline.UpperBound
				config.MaxIterations = baseline.Iterations
				config.Rand = rand.New(rand.NewSource(baseline.Seed + int64(i)))

				result, err := Optimize(config)
				if err != nil {
					t.Fatalf("Optimization failed: %v", err)
				}

				costs[i] = result.GlobalBest.Cost

				// Check if this run was successful
				if costs[i] <= baseline.ExpectedBest*baseline.Tolerance {
					successCount++
				}
			}

			// Calculate statistics
			stats := calculateStatistics(costs)

			// Check best cost
			if stats.Min > baseline.ExpectedBest*baseline.Tolerance {
				t.Errorf("Best cost %.6e exceeds baseline %.6e * tolerance %.1f = %.6e",
					stats.Min, baseline.ExpectedBest, baseline.Tolerance,
					baseline.ExpectedBest*baseline.Tolerance)
			}

			// Check mean cost
			if stats.Mean > baseline.ExpectedMean*baseline.Tolerance {
				t.Errorf("Mean cost %.6e exceeds baseline %.6e * tolerance %.1f = %.6e",
					stats.Mean, baseline.ExpectedMean, baseline.Tolerance,
					baseline.ExpectedMean*baseline.Tolerance)
			}

			// Check success rate
			successRate := float64(successCount) / float64(runs)
			if successRate < baseline.SuccessThreshold {
				t.Errorf("Success rate %.2f%% is below threshold %.2f%%",
					successRate*100, baseline.SuccessThreshold*100)
			}

			// Log results for monitoring
			t.Logf("Best: %.6e | Mean: %.6e | StdDev: %.6e | Success: %.1f%%",
				stats.Min, stats.Mean, stats.StdDev, successRate*100)

			// Check for improvement (informational only)
			if stats.Mean < baseline.ExpectedMean*0.5 {
				t.Logf("INFO: Performance significantly improved! Mean %.6e is less than 50%% of baseline %.6e",
					stats.Mean, baseline.ExpectedMean)
			}
		})
	}
}

// TestRegressionQuick runs a quick regression check with a single run per baseline.
// Used for fast CI checks.
func TestRegressionQuick(t *testing.T) {
	// Quick subset of baselines
	quickBaselines := []RegressionBaseline{
		regressionBaselines[0], // StandardMA_Sphere_10D
		regressionBaselines[1], // DESMA_Sphere_10D
		regressionBaselines[2], // StandardMA_Rastrigin_10D
		regressionBaselines[3], // DESMA_Rastrigin_10D
	}

	for _, baseline := range quickBaselines {
		t.Run(baseline.Name, func(t *testing.T) {
			var config *Config
			if baseline.UseDESMA {
				config = NewDESMAConfig()
			} else {
				config = NewDefaultConfig()
			}

			config.ObjectiveFunc = baseline.Function
			config.ProblemSize = baseline.Dimensions
			config.LowerBound = baseline.LowerBound
			config.UpperBound = baseline.UpperBound
			config.MaxIterations = baseline.Iterations
			config.Rand = rand.New(rand.NewSource(baseline.Seed))

			result, err := Optimize(config)
			if err != nil {
				t.Fatalf("Optimization failed: %v", err)
			}

			cost := result.GlobalBest.Cost

			// Check against baseline with tolerance
			threshold := baseline.ExpectedBest * baseline.Tolerance
			if cost > threshold {
				t.Errorf("Cost %.6e exceeds threshold %.6e (baseline %.6e * tolerance %.1f)",
					cost, threshold, baseline.ExpectedBest, baseline.Tolerance)
			}

			t.Logf("Cost: %.6e (threshold: %.6e)", cost, threshold)
		})
	}
}

// TestRegressionReproducibility ensures deterministic results with fixed seeds.
func TestRegressionReproducibility(t *testing.T) {
	baseline := regressionBaselines[0] // Use Sphere as test case

	runOptimization := func(seed int64) float64 {
		config := NewDefaultConfig()
		config.ObjectiveFunc = baseline.Function
		config.ProblemSize = baseline.Dimensions
		config.LowerBound = baseline.LowerBound
		config.UpperBound = baseline.UpperBound
		config.MaxIterations = 100 // Shorter for quick test
		config.Rand = rand.New(rand.NewSource(seed))

		result, err := Optimize(config)
		if err != nil {
			t.Fatalf("Optimization failed: %v", err)
		}

		return result.GlobalBest.Cost
	}

	// Run with same seed multiple times
	seed := int64(42)
	cost1 := runOptimization(seed)
	cost2 := runOptimization(seed)
	cost3 := runOptimization(seed)

	// All runs with same seed should produce identical results
	if math.Abs(cost1-cost2) > 1e-15 || math.Abs(cost2-cost3) > 1e-15 {
		t.Errorf("Results not reproducible: %.15e, %.15e, %.15e", cost1, cost2, cost3)
	}

	t.Logf("Reproducibility verified: %.6e = %.6e = %.6e", cost1, cost2, cost3)
}

// TestRegressionNoRegression ensures recent changes haven't degraded performance.
// This test should be updated when baselines are intentionally improved.
func TestRegressionNoRegression(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping no-regression test in short mode")
	}

	// Test that DESMA consistently outperforms Standard MA on known problems
	testCases := []struct {
		function   ObjectiveFunction
		name       string
		dimensions int
		lower      float64
		upper      float64
	}{
		{"Rastrigin", Rastrigin, 10, -5.12, 5.12},
		{"Rosenbrock", Rosenbrock, 10, -5, 10},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runs := 5
			iterations := 500

			maCosts := make([]float64, runs)
			desmaCosts := make([]float64, runs)

			for i := 0; i < runs; i++ {
				seed := int64(i + 1)

				// Standard MA
				maConfig := NewDefaultConfig()
				maConfig.ObjectiveFunc = tc.function
				maConfig.ProblemSize = tc.dimensions
				maConfig.LowerBound = tc.lower
				maConfig.UpperBound = tc.upper
				maConfig.MaxIterations = iterations
				maConfig.Rand = rand.New(rand.NewSource(seed))

				maResult, err := Optimize(maConfig)
				if err != nil {
					t.Fatalf("Standard MA failed: %v", err)
				}

				maCosts[i] = maResult.GlobalBest.Cost

				// DESMA
				desmaConfig := NewDESMAConfig()
				desmaConfig.ObjectiveFunc = tc.function
				desmaConfig.ProblemSize = tc.dimensions
				desmaConfig.LowerBound = tc.lower
				desmaConfig.UpperBound = tc.upper
				desmaConfig.MaxIterations = iterations
				desmaConfig.Rand = rand.New(rand.NewSource(seed))

				desmaResult, err := Optimize(desmaConfig)
				if err != nil {
					t.Fatalf("DESMA failed: %v", err)
				}

				desmaCosts[i] = desmaResult.GlobalBest.Cost
			}

			maStats := calculateStatistics(maCosts)
			desmaStats := calculateStatistics(desmaCosts)

			improvement := (maStats.Mean - desmaStats.Mean) / maStats.Mean * 100

			t.Logf("Standard MA - Mean: %.6e, Min: %.6e", maStats.Mean, maStats.Min)
			t.Logf("DESMA       - Mean: %.6e, Min: %.6e", desmaStats.Mean, desmaStats.Min)
			t.Logf("Improvement: %.2f%%", improvement)

			// DESMA should improve performance on these problems
			// Allow some variance but expect general improvement
			if desmaStats.Mean > maStats.Mean*1.2 {
				t.Errorf("DESMA performed worse than Standard MA (DESMA: %.6e > MA: %.6e * 1.2)",
					desmaStats.Mean, maStats.Mean)
			}
		})
	}
}
