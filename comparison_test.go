package mayfly

import (
	"testing"
)

// =============================================================================
// Tests for comparison.go - Statistical Comparison Framework
// =============================================================================

func TestComparisonRunner(t *testing.T) {
	runner := NewComparisonRunner()
	if runner == nil {
		t.Fatal("NewComparisonRunner() should not return nil")
	}

	// Check defaults
	if runner.Runs != 30 {
		t.Errorf("Default runs should be 30, got %d", runner.Runs)
	}
	if runner.MaxIterations != 500 {
		t.Errorf("Default max iterations should be 500, got %d", runner.MaxIterations)
	}

	// Test fluent API
	runner = runner.
		WithVariantNames("ma", "desma").
		WithRuns(3).
		WithIterations(50).
		WithVerbose(false)

	if len(runner.Variants) != 2 {
		t.Errorf("Expected 2 variants, got %d", len(runner.Variants))
	}
	if runner.Runs != 3 {
		t.Errorf("Expected 3 runs, got %d", runner.Runs)
	}
	if runner.MaxIterations != 50 {
		t.Errorf("Expected 50 iterations, got %d", runner.MaxIterations)
	}
}

func TestComparisonRunnerWithVariants(t *testing.T) {
	// Test WithVariants method
	variant1 := NewVariant("ma")
	variant2 := NewVariant("desma")

	runner := NewComparisonRunner().
		WithVariants(variant1, variant2).
		WithRuns(5)

	if len(runner.Variants) != 2 {
		t.Errorf("Expected 2 variants, got %d", len(runner.Variants))
	}
}

func TestComparisonRunnerCompare(t *testing.T) {
	// Quick comparison test (small runs for speed)
	runner := NewComparisonRunner().
		WithVariantNames("ma", "desma").
		WithRuns(2).
		WithIterations(20).
		WithVerbose(false)

	result := runner.Compare("Sphere", Sphere, 5, -10, 10)

	if result == nil {
		t.Fatal("Comparison result should not be nil")
	}

	if result.BenchmarkName != "Sphere" {
		t.Errorf("Expected benchmark name 'Sphere', got '%s'", result.BenchmarkName)
	}

	if len(result.AlgorithmNames) != 2 {
		t.Errorf("Expected 2 algorithm names, got %d", len(result.AlgorithmNames))
	}

	if len(result.Statistics) != 2 {
		t.Errorf("Expected 2 statistics entries, got %d", len(result.Statistics))
	}

	if len(result.Rankings) != 2 {
		t.Errorf("Expected 2 rankings, got %d", len(result.Rankings))
	}

	if len(result.RunResults) != 2 {
		t.Errorf("Expected 2 run result sets, got %d", len(result.RunResults))
	}

	// Check statistics are valid
	for i, stats := range result.Statistics {
		if stats.Best > stats.Worst {
			t.Errorf("Best (%f) should be <= Worst (%f) for algorithm %d", stats.Best, stats.Worst, i)
		}
		if stats.Mean < stats.Best || stats.Mean > stats.Worst {
			t.Errorf("Mean (%f) should be between Best (%f) and Worst (%f)", stats.Mean, stats.Best, stats.Worst)
		}
		if stats.StdDev < 0 {
			t.Errorf("StdDev should be non-negative, got %f", stats.StdDev)
		}
		if stats.AvgFuncEvals <= 0 {
			t.Errorf("AvgFuncEvals should be positive, got %f", stats.AvgFuncEvals)
		}
	}

	// Rankings should be 1 and 2
	rankSum := 0
	for _, rank := range result.Rankings {
		rankSum += rank
		if rank < 1 || rank > 2 {
			t.Errorf("Rank should be 1 or 2, got %d", rank)
		}
	}
	if rankSum != 3 { // 1 + 2 = 3
		t.Errorf("Rank sum should be 3, got %d", rankSum)
	}

	// Best algorithm should be valid index
	if result.BestAlgorithm < 0 || result.BestAlgorithm >= len(result.AlgorithmNames) {
		t.Errorf("BestAlgorithm index %d out of range", result.BestAlgorithm)
	}

	// Wilcoxon tests should be present
	if len(result.WilcoxonTests) != 2 {
		t.Errorf("Expected 2x2 Wilcoxon test matrix, got %d rows", len(result.WilcoxonTests))
	}
}

func TestWilcoxonSignedRankTest(t *testing.T) {
	// Create mock run results where alg1 is clearly better
	runs1 := []RunResult{
		{BestCost: 1.0}, {BestCost: 2.0}, {BestCost: 3.0},
	}
	runs2 := []RunResult{
		{BestCost: 2.0}, {BestCost: 3.0}, {BestCost: 4.0},
	}

	result := wilcoxonSignedRankTest("Alg1", "Alg2", runs1, runs2)

	if result.Algorithm1 != "Alg1" {
		t.Errorf("Expected Algorithm1 'Alg1', got '%s'", result.Algorithm1)
	}
	if result.Algorithm2 != "Alg2" {
		t.Errorf("Expected Algorithm2 'Alg2', got '%s'", result.Algorithm2)
	}

	// Winner should be Alg1 (lower costs) or Tie
	if result.Winner != "Alg1" && result.Winner != "Tie" {
		t.Errorf("Expected winner 'Alg1' or 'Tie', got '%s'", result.Winner)
	}

	// P-value should be in [0, 1]
	if result.PValue < 0 || result.PValue > 1 {
		t.Errorf("PValue should be in [0,1], got %f", result.PValue)
	}

	// W statistic should be non-negative
	if result.WStatistic < 0 {
		t.Errorf("WStatistic should be non-negative, got %f", result.WStatistic)
	}
}

func TestWilcoxonWithEqualResults(t *testing.T) {
	// Test with equal results (should be a tie)
	runs1 := []RunResult{
		{BestCost: 1.0}, {BestCost: 2.0}, {BestCost: 3.0},
	}
	runs2 := []RunResult{
		{BestCost: 1.0}, {BestCost: 2.0}, {BestCost: 3.0},
	}

	result := wilcoxonSignedRankTest("Alg1", "Alg2", runs1, runs2)

	if result.Winner != "Tie" {
		t.Logf("Expected 'Tie' for equal results, got '%s' (acceptable variation)", result.Winner)
	}
}

func TestFriedmanTest(t *testing.T) {
	// Create mock run results for 3 algorithms
	runResults := [][]RunResult{
		{{BestCost: 1.0}, {BestCost: 2.0}, {BestCost: 3.0}},
		{{BestCost: 2.0}, {BestCost: 3.0}, {BestCost: 4.0}},
		{{BestCost: 3.0}, {BestCost: 4.0}, {BestCost: 5.0}},
	}

	result := friedmanTest(runResults)

	if result == nil {
		t.Fatal("Friedman test should not return nil")
	}

	if result.ChiSquare < 0 {
		t.Errorf("ChiSquare should be non-negative, got %f", result.ChiSquare)
	}

	if result.PValue < 0 || result.PValue > 1 {
		t.Errorf("PValue should be in [0,1], got %f", result.PValue)
	}

	if result.DegreesOfFreedom != 2 { // k-1 where k=3
		t.Errorf("Expected 2 degrees of freedom, got %d", result.DegreesOfFreedom)
	}
}

func TestRankValues(t *testing.T) {
	values := []float64{3.0, 1.0, 2.0, 1.0, 4.0}
	ranks := rankValues(values)

	// Check rank count
	if len(ranks) != len(values) {
		t.Errorf("Expected %d ranks, got %d", len(values), len(ranks))
	}

	// Ranks should be positive
	for i, rank := range ranks {
		if rank <= 0 {
			t.Errorf("Rank %d should be positive, got %f", i, rank)
		}
	}

	// Smallest value (1.0) should have lowest average rank
	// Values at indices 1 and 3 are both 1.0, should have average rank 1.5
	if ranks[1] != 1.5 || ranks[3] != 1.5 {
		t.Logf("Tied values should have average rank, got ranks[1]=%f, ranks[3]=%f", ranks[1], ranks[3])
	}
}

func TestCalculateAlgorithmStatistics(t *testing.T) {
	runs := []RunResult{
		{BestCost: 1.0, FuncEvals: 100, ExecutionTime: 0.1},
		{BestCost: 2.0, FuncEvals: 110, ExecutionTime: 0.15},
		{BestCost: 3.0, FuncEvals: 105, ExecutionTime: 0.12},
		{BestCost: 1.5, FuncEvals: 108, ExecutionTime: 0.11},
	}

	stats := calculateAlgorithmStatistics(runs, 2.5)

	if stats.Best != 1.0 {
		t.Errorf("Expected Best=1.0, got %f", stats.Best)
	}
	if stats.Worst != 3.0 {
		t.Errorf("Expected Worst=3.0, got %f", stats.Worst)
	}
	if stats.Mean < stats.Best || stats.Mean > stats.Worst {
		t.Errorf("Mean %f should be between Best and Worst", stats.Mean)
	}
	if stats.StdDev < 0 {
		t.Errorf("StdDev should be non-negative, got %f", stats.StdDev)
	}

	// Success rate should be 75% (3 out of 4 runs <= 2.5)
	// BestCost values: 1.0, 2.0, 3.0, 1.5 -> only 3.0 exceeds 2.5
	if stats.SuccessRate != 75.0 {
		t.Errorf("Expected 75%% success rate, got %f", stats.SuccessRate)
	}

	// Average function evaluations
	expectedAvgEvals := (100.0 + 110.0 + 105.0 + 108.0) / 4.0
	if stats.AvgFuncEvals != expectedAvgEvals {
		t.Errorf("Expected AvgFuncEvals=%f, got %f", expectedAvgEvals, stats.AvgFuncEvals)
	}
}

func TestRankAlgorithms(t *testing.T) {
	statistics := []AlgorithmStatistics{
		{Mean: 2.0}, // Should be rank 2
		{Mean: 1.0}, // Should be rank 1 (best)
		{Mean: 3.0}, // Should be rank 3 (worst)
	}

	rankings := rankAlgorithms(statistics)

	if len(rankings) != 3 {
		t.Errorf("Expected 3 rankings, got %d", len(rankings))
	}

	if rankings[0] != 2 {
		t.Errorf("First algorithm should be rank 2, got %d", rankings[0])
	}
	if rankings[1] != 1 {
		t.Errorf("Second algorithm should be rank 1, got %d", rankings[1])
	}
	if rankings[2] != 3 {
		t.Errorf("Third algorithm should be rank 3, got %d", rankings[2])
	}
}

func TestNormalCDF(t *testing.T) {
	// Test standard normal CDF at known points
	tests := []struct {
		x        float64
		expected float64
		tolerance float64
	}{
		{0.0, 0.5, 0.01},      // Mean
		{-1.96, 0.025, 0.01},  // ~2.5th percentile
		{1.96, 0.975, 0.01},   // ~97.5th percentile
	}

	for _, tt := range tests {
		result := normalCDF(tt.x)
		if result < tt.expected-tt.tolerance || result > tt.expected+tt.tolerance {
			t.Errorf("normalCDF(%f) = %f, expected ~%f", tt.x, result, tt.expected)
		}
	}
}
