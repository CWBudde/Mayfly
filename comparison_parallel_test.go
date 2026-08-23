package mayfly

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type comparisonTestVariant struct {
	name          string
	invalidConfig bool
}

func (v comparisonTestVariant) Name() string        { return v.name }
func (v comparisonTestVariant) FullName() string    { return v.name }
func (v comparisonTestVariant) Description() string { return "comparison test variant" }
func (v comparisonTestVariant) ApplicableTo(ProblemCharacteristics) float64 {
	return 1
}
func (v comparisonTestVariant) EstimatedOverhead() float64 { return 1 }
func (v comparisonTestVariant) RecommendedFor() []string   { return nil }
func (v comparisonTestVariant) GetConfig() *Config {
	config := NewDefaultConfig()
	config.NPop = 2
	config.NPopF = 2
	config.NC = 2

	config.NM = 0
	if v.invalidConfig {
		config.NPop = 0
	}

	return config
}

func TestGetAllVariantsHasStableCanonicalOrder(t *testing.T) {
	variants := GetAllVariants()
	names := make([]string, len(variants))

	for i, variant := range variants {
		names[i] = variant.Name()
	}

	want := []string{"MA", "DESMA", "OLCE-MA", "EOBBMA", "GSASMA", "HMMA", "MPMA", "AOBLMOA"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("GetAllVariants names = %v, want %v", names, want)
	}
}

func TestComparisonRunnerParallelOptionsAndDefaults(t *testing.T) {
	runner := NewComparisonRunner()
	if runner.Parallel {
		t.Fatal("new comparison runner unexpectedly enables parallel execution")
	}

	if runner.MaxWorkers != runtime.NumCPU() {
		t.Fatalf("MaxWorkers = %d, want %d", runner.MaxWorkers, runtime.NumCPU())
	}

	got := runner.WithParallel(true).WithMaxWorkers(3).WithSeed(42)
	if got != runner {
		t.Fatal("parallel option methods should preserve the fluent runner")
	}

	if !runner.Parallel || runner.MaxWorkers != 3 || runner.Seed != 42 {
		t.Fatalf("parallel options not retained: %+v", runner)
	}
}

func TestComparisonRunnerBoundsOuterConcurrency(t *testing.T) {
	var active atomic.Int64

	var maximum atomic.Int64

	objective := func(position []float64) float64 {
		current := active.Add(1)
		defer active.Add(-1)

		for {
			observed := maximum.Load()
			if current <= observed || maximum.CompareAndSwap(observed, current) {
				break
			}
		}

		time.Sleep(time.Millisecond)

		return Sphere(position)
	}

	runner := NewComparisonRunner().
		WithVariants(comparisonTestVariant{name: "A"}, comparisonTestVariant{name: "B"}).
		WithRuns(3).
		WithIterations(1).
		WithSeed(100).
		WithParallel(true).
		WithMaxWorkers(2)

	result, err := runner.CompareContext(context.Background(), "bounded", objective, 2, -1, 1)
	if err != nil {
		t.Fatalf("CompareContext: %v", err)
	}

	if got := maximum.Load(); got < 2 || got > 2 {
		t.Fatalf("maximum concurrent objective calls = %d, want exactly 2", got)
	}

	for i, runs := range result.RunResults {
		if len(runs) != 3 {
			t.Fatalf("algorithm %d has %d runs, want 3", i, len(runs))
		}

		for j, run := range runs {
			if run.Error != "" || run.FuncEvals == 0 {
				t.Fatalf("run [%d][%d] was not completed: %+v", i, j, run)
			}
		}
	}
}

func TestComparisonSeedMakesSequentialAndParallelRunsEqualAndPaired(t *testing.T) {
	newRunner := func(parallel bool) *ComparisonRunner {
		return NewComparisonRunner().
			WithVariants(comparisonTestVariant{name: "A"}, comparisonTestVariant{name: "B"}).
			WithRuns(3).
			WithIterations(2).
			WithSeed(9876).
			WithParallel(parallel).
			WithMaxWorkers(3)
	}

	sequential, err := newRunner(false).CompareContext(context.Background(), "seeded", Sphere, 3, -2, 2)
	if err != nil {
		t.Fatalf("sequential CompareContext: %v", err)
	}

	parallel, err := newRunner(true).CompareContext(context.Background(), "seeded", Sphere, 3, -2, 2)
	if err != nil {
		t.Fatalf("parallel CompareContext: %v", err)
	}

	for algorithm := range sequential.RunResults {
		for runIndex, sequentialRun := range sequential.RunResults[algorithm] {
			parallelRun := parallel.RunResults[algorithm][runIndex]

			wantSeed := int64(9876 + runIndex)
			if sequentialRun.Seed != wantSeed || parallelRun.Seed != wantSeed {
				t.Fatalf("run %d seeds = (%d, %d), want paired seed %d",
					runIndex, sequentialRun.Seed, parallelRun.Seed, wantSeed)
			}

			sequentialRun.ExecutionTime = 0
			parallelRun.ExecutionTime = 0

			if !reflect.DeepEqual(sequentialRun, parallelRun) {
				t.Fatalf("run [%d][%d] differs by scheduling:\nsequential: %+v\nparallel:   %+v",
					algorithm, runIndex, sequentialRun, parallelRun)
			}
		}
	}

	for runIndex := range sequential.RunResults[0] {
		if sequential.RunResults[0][runIndex].Seed != sequential.RunResults[1][runIndex].Seed {
			t.Fatalf("run %d is not paired across algorithms", runIndex)
		}
	}
}

func TestComparisonContextValidationCancellationAndRunErrors(t *testing.T) {
	valid := NewComparisonRunner().WithVariants(comparisonTestVariant{name: "A"}).WithRuns(1).WithIterations(1)

	tests := []struct {
		name       string
		runner     *ComparisonRunner
		fn         ObjectiveFunction
		size       int
		lower      float64
		upper      float64
		nilContext bool
	}{
		{name: "nil context", runner: valid, fn: Sphere, size: 2, lower: -1, upper: 1, nilContext: true},
		{name: "no variants", runner: NewComparisonRunner().WithVariants(), fn: Sphere, size: 2, lower: -1, upper: 1},
		{name: "zero runs", runner: NewComparisonRunner().WithRuns(0), fn: Sphere, size: 2, lower: -1, upper: 1},
		{name: "negative workers", runner: NewComparisonRunner().WithMaxWorkers(-1), fn: Sphere, size: 2, lower: -1, upper: 1},
		{name: "nil objective", runner: valid, size: 2, lower: -1, upper: 1},
		{name: "bad size", runner: valid, fn: Sphere, size: 0, lower: -1, upper: 1},
		{name: "bad bounds", runner: valid, fn: Sphere, size: 2, lower: 1, upper: 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			if test.nilContext {
				ctx = nil
			}

			_, err := test.runner.CompareContext(ctx, "invalid", test.fn, test.size, test.lower, test.upper)
			if err == nil {
				t.Fatal("CompareContext accepted invalid input")
			}
		})
	}

	canceled, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := valid.CompareContext(canceled, "canceled", Sphere, 2, -1, 1)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("pre-canceled CompareContext error = %v, want context.Canceled", err)
	}

	broken := NewComparisonRunner().
		WithVariants(comparisonTestVariant{name: "broken", invalidConfig: true}).
		WithRuns(1).
		WithIterations(1)

	result, err := broken.CompareContext(context.Background(), "broken", Sphere, 2, -1, 1)
	if err == nil || result != nil {
		t.Fatalf("failed run returned result=%v, err=%v; want nil result and error", result, err)
	}
}

func comparisonReportFixture() *ComparisonResult {
	return &ComparisonResult{
		BenchmarkName:  "fixture",
		AlgorithmNames: []string{"finite", "failed"},
		RunResults: [][]RunResult{
			{{BestCost: 1, FuncEvals: 10, Iterations: 2, Seed: 7}},
			{{BestCost: math.Inf(1), Seed: 7, Error: "boom"}},
		},
		Statistics: []AlgorithmStatistics{
			{Mean: 1, Median: 1, Best: 1, Worst: 1, AvgFuncEvals: 10},
			{Mean: math.Inf(1), Median: math.Inf(1), Best: math.Inf(1), Worst: math.Inf(1)},
		},
		Rankings:      []int{1, 2},
		WilcoxonTests: [][]WilcoxonResult{{{}, {}}, {{}, {}}},
		BestAlgorithm: 0,
		BaseSeed:      7,
	}
}

func TestWriteComparisonResultsIncludesChartAndHandlesFailures(t *testing.T) {
	var output bytes.Buffer

	err := comparisonReportFixture().WriteComparisonResults(&output)
	if err != nil {
		t.Fatalf("WriteComparisonResults: %v", err)
	}

	for _, want := range []string{"Benchmark Comparison: fixture", "Relative Quality", "finite", "########################", "failed/unavailable"} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("report does not contain %q:\n%s", want, output.String())
		}
	}

	err = comparisonReportFixture().WriteComparisonResults(errorWriter{})
	if err == nil {
		t.Fatal("WriteComparisonResults did not propagate writer failure")
	}

	err = comparisonReportFixture().WriteComparisonResults(nil)
	if err == nil {
		t.Fatal("WriteComparisonResults accepted a nil writer")
	}
}

type errorWriter struct{}

func (errorWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }

func TestComparisonResultExportsCSVAndJSON(t *testing.T) {
	result := comparisonReportFixture()
	result.RunResults[1][0].BestCost = 2
	result.RunResults[1][0].Error = ""
	result.Statistics[1] = AlgorithmStatistics{Mean: 2, Median: 2, Best: 2, Worst: 2}
	dir := t.TempDir()

	csvPath := filepath.Join(dir, "comparison.csv")

	err := result.ExportToCSV(csvPath)
	if err != nil {
		t.Fatalf("ExportToCSV: %v", err)
	}

	file, err := os.Open(csvPath)
	if err != nil {
		t.Fatalf("open CSV: %v", err)
	}

	records, err := csv.NewReader(file).ReadAll()

	closeErr := file.Close()
	if err != nil || closeErr != nil {
		t.Fatalf("read CSV: read=%v close=%v", err, closeErr)
	}

	if len(records) != 3 || records[0][0] != "benchmark" || records[1][0] != "fixture" || records[2][1] != "failed" {
		t.Fatalf("unexpected CSV records: %v", records)
	}

	jsonPath := filepath.Join(dir, "comparison.json")

	err = result.ExportToJSON(jsonPath)
	if err != nil {
		t.Fatalf("ExportToJSON: %v", err)
	}

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read JSON: %v", err)
	}

	var decoded ComparisonResult

	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("decode JSON: %v", err)
	}

	if decoded.BenchmarkName != result.BenchmarkName || decoded.BaseSeed != result.BaseSeed ||
		!reflect.DeepEqual(decoded.AlgorithmNames, result.AlgorithmNames) {
		t.Fatalf("unexpected JSON export: %+v", decoded)
	}
}
