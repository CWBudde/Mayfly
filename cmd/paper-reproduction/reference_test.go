package main

import (
	"math"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/mayfly"
)

func loadTestOriginalMAReference(t *testing.T) originalMAReference {
	t.Helper()

	path := filepath.Join("..", "..", "docs", "reference-data", "original-ma-2020-table6.json")

	reference, err := loadOriginalMAReference(path)
	if err != nil {
		t.Fatalf("loadOriginalMAReference() error = %v", err)
	}

	return reference
}

func referenceTestRun(cost float64, evaluations int, runError string) mayfly.RunResult {
	return mayfly.RunResult{
		Error:         runError,
		BestCost:      cost,
		ExecutionTime: 0,
		Seed:          0,
		FuncEvals:     evaluations,
		Iterations:    0,
		ConvergenceAt: 0,
	}
}

func referenceTestResult(name string, runs []mayfly.RunResult) *mayfly.ComparisonResult {
	return &mayfly.ComparisonResult{
		FriedmanResult: nil,
		BenchmarkName:  "test",
		AlgorithmNames: []string{name},
		RunResults:     [][]mayfly.RunResult{runs},
		Statistics:     nil,
		Rankings:       nil,
		WilcoxonTests:  nil,
		BestAlgorithm:  0,
		BaseSeed:       0,
	}
}

func TestLoadOriginalMAReference(t *testing.T) {
	reference := loadTestOriginalMAReference(t)

	if reference.ProtocolID != "original-ma-2020-table6" || len(reference.Benchmarks) != 6 {
		t.Fatalf("unexpected reference: protocol=%q benchmarks=%d", reference.ProtocolID, len(reference.Benchmarks))
	}

	if len(reference.SHA256) != 64 {
		t.Fatalf("SHA-256 length = %d, want 64", len(reference.SHA256))
	}

	benchmark, ok := findOriginalMABenchmark(reference.Benchmarks, "F20")
	if !ok || benchmark.Name != "Beale" || benchmark.Dimension != 2 {
		t.Fatalf("F20 benchmark = %+v, found=%v", benchmark, ok)
	}
}

func TestValidateOriginalMAReferenceRejectsMalformedRows(t *testing.T) {
	reference := loadTestOriginalMAReference(t)
	reference.PublishedResults["F1"]["basic_ma"] = []float64{0, 1, 0.5}

	err := validateOriginalMAReference(reference)
	if err == nil || !strings.Contains(err.Error(), "has 3 statistics, want 5") {
		t.Fatalf("validateOriginalMAReference() error = %v", err)
	}
}

func TestValidateOriginalMAReferenceRejectsReorderedStatistics(t *testing.T) {
	reference := loadTestOriginalMAReference(t)
	reference.StatisticsOrder[0], reference.StatisticsOrder[1] =
		reference.StatisticsOrder[1], reference.StatisticsOrder[0]

	err := validateOriginalMAReference(reference)
	if err == nil || !strings.Contains(err.Error(), "statistics_order") {
		t.Fatalf("validateOriginalMAReference() error = %v", err)
	}
}

func TestBuildBasicMAComparisonSummaryIsExplicitlyNonReproduction(t *testing.T) {
	reference := loadTestOriginalMAReference(t)
	runs := make([]mayfly.RunResult, reference.Execution.Replications)

	for index := range runs {
		runs[index] = referenceTestRun(float64(index), reference.Execution.FunctionEvaluations, "")
	}

	config := mayfly.NewDefaultConfig()
	input := basicMAComparisonInput{
		BenchmarkID: "F1",
		Dimension:   5,
		LowerBound:  -10,
		UpperBound:  10,
		Config:      config,
		Result:      referenceTestResult("MA", runs),
	}

	summary, err := buildBasicMAComparisonSummary(reference, []basicMAComparisonInput{input})
	if err != nil {
		t.Fatalf("buildBasicMAComparisonSummary() error = %v", err)
	}

	if summary.ReproductionClaim || summary.ComparisonKind != "descriptive_non_reproduction" {
		t.Fatalf("summary mislabels reproduction status: %+v", summary)
	}

	if len(summary.Comparisons) != 1 || summary.Comparisons[0].PublishedAlgorithm != "basic_ma" {
		t.Fatalf("comparisons = %+v", summary.Comparisons)
	}

	comparison := summary.Comparisons[0]
	if !comparison.Alignment.BenchmarkGeometry || !comparison.Alignment.Replications ||
		!comparison.Alignment.EvaluationBudget || !comparison.Alignment.Population {
		t.Fatalf("unexpected protocol alignment: %+v", comparison.Alignment)
	}

	if comparison.Alignment.KnownParameters || comparison.Alignment.OperatorSemantics ||
		comparison.Alignment.PublishedSeedsKnown {
		t.Fatalf("ambiguous protocol was marked aligned: %+v", comparison.Alignment)
	}

	if comparison.Computed.Mean != 24.5 || comparison.Computed.Median != 24.5 ||
		comparison.Computed.Best != 0 || comparison.Computed.Worst != 49 {
		t.Fatalf("computed statistics = %+v", comparison.Computed)
	}

	wantPopulationStdDev := math.Sqrt(208.25)
	if math.Abs(comparison.Computed.PopulationStandardDev-wantPopulationStdDev) > 1e-12 {
		t.Fatalf("population standard deviation = %g, want %g",
			comparison.Computed.PopulationStandardDev, wantPopulationStdDev)
	}

	if comparison.PublishedStdDevConvention != "unknown" ||
		comparison.Published.ReportedStandardDeviation != 0 {
		t.Fatalf("published standard deviation metadata = %+v", comparison)
	}
}

func TestBuildBasicMAComparisonSummaryReportsFailedAndShortRuns(t *testing.T) {
	reference := loadTestOriginalMAReference(t)
	config := mayfly.NewDefaultConfig()
	input := basicMAComparisonInput{
		BenchmarkID: "F10",
		Dimension:   10,
		LowerBound:  -5.12,
		UpperBound:  5.12,
		Config:      config,
		Result: referenceTestResult("MA", []mayfly.RunResult{
			referenceTestRun(2, reference.Execution.FunctionEvaluations-1, ""),
			referenceTestRun(1, reference.Execution.FunctionEvaluations, "failed"),
		}),
	}

	summary, err := buildBasicMAComparisonSummary(reference, []basicMAComparisonInput{input})
	if err != nil {
		t.Fatalf("buildBasicMAComparisonSummary() error = %v", err)
	}

	comparison := summary.Comparisons[0]
	if comparison.AvailableRuns != 1 || comparison.FailedRuns != 1 {
		t.Fatalf("run counts = available %d, failed %d", comparison.AvailableRuns, comparison.FailedRuns)
	}

	if comparison.Alignment.BenchmarkGeometry || comparison.Alignment.Replications ||
		comparison.Alignment.EvaluationBudget {
		t.Fatalf("mismatched protocol was marked aligned: %+v", comparison.Alignment)
	}
}

func TestPopulationStandardDeviationPreservesSubnormalScale(t *testing.T) {
	values := []float64{0, 1e-240}

	got := populationStandardDeviation(values, 5e-241)
	if got != 5e-241 {
		t.Fatalf("population standard deviation = %g, want %g", got, 5e-241)
	}
}

func TestBuildBasicMAComparisonSummaryRejectsMissingMA(t *testing.T) {
	reference := loadTestOriginalMAReference(t)

	_, err := buildBasicMAComparisonSummary(reference, []basicMAComparisonInput{{
		BenchmarkID: "F1",
		Dimension:   5,
		LowerBound:  -10,
		UpperBound:  10,
		Config:      mayfly.NewDefaultConfig(),
		Result:      referenceTestResult("DESMA", []mayfly.RunResult{referenceTestRun(1, 1, "")}),
	}})
	if err == nil || !strings.Contains(err.Error(), "no current-library MA runs") {
		t.Fatalf("buildBasicMAComparisonSummary() error = %v", err)
	}
}
