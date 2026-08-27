package main

import (
	"encoding/json"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/cwbudde/mayfly"
)

func TestParseOptionsDESMATable3ModePinsProtocol(t *testing.T) {
	opts, err := parseOptions([]string{
		"-desma-table3-data", "cec-data",
		"-output", "results",
		"-workers", "3",
		"-seed", "17",
	})
	if err != nil {
		t.Fatalf("parseOptions() error = %v", err)
	}

	if opts.desmaTable3Data != "cec-data" || opts.outputDir != "results" || opts.workers != 3 || opts.seed != 17 {
		t.Fatalf("unexpected caller-controlled options: %+v", opts)
	}

	if !reflect.DeepEqual(opts.variants, []string{"desma"}) ||
		!reflect.DeepEqual(opts.dimensions, []int{desmaTable3Dimension}) ||
		len(opts.benchmarks) != 0 || opts.runs != desmaTable3Runs ||
		opts.maxEvals != desmaTable3Evaluations || opts.iterations != desmaTable3Iterations {
		t.Fatalf("DESMA Table 3 protocol was not pinned: %+v", opts)
	}
}

func TestParseOptionsDESMATable3ModeRejectsProtocolOverrides(t *testing.T) {
	for _, arguments := range [][]string{
		{"-desma-table3-data", "cec-data", "-runs", "51"},
		{"-desma-table3-data", "cec-data", "-max-evaluations", "300000"},
		{"-desma-table3-data", "cec-data", "-variants", "desma"},
		{"-desma-table3-data", "cec-data", "-dimensions", "30"},
		{"-desma-table3-data", "cec-data", "-published-reference", "reference.json"},
	} {
		_, err := parseOptions(arguments)
		if err == nil {
			t.Errorf("parseOptions(%v) accepted a fixed-protocol override", arguments)
		}
	}
}

func TestDESMATable3ManifestLabelsNonReproduction(t *testing.T) {
	variant, err := mayfly.NewVariantChecked("desma")
	if err != nil {
		t.Fatalf("NewVariantChecked() error = %v", err)
	}

	opts := options{
		outputDir:          "results",
		publishedReference: "",
		desmaTable3Data:    "cec-data",
		benchmarks:         nil,
		variants:           []string{"desma"},
		dimensions:         []int{desmaTable3Dimension},
		runs:               desmaTable3Runs,
		iterations:         desmaTable3Iterations,
		maxEvals:           desmaTable3Evaluations,
		workers:            1,
		seed:               42,
	}

	protocol, err := newDESMATable3Manifest(opts, variant, nil)
	if err != nil {
		t.Fatalf("newDESMATable3Manifest() error = %v", err)
	}

	if protocol.ProtocolID != desmaTable3ProtocolID || protocol.ComparisonKind != desmaTable3Comparison ||
		protocol.ReproductionClaim == nil || *protocol.ReproductionClaim {
		t.Fatalf("manifest mislabels reproduction status: %+v", protocol)
	}

	data, err := json.Marshal(protocol)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}

	if !strings.Contains(string(data), `"reproduction_claim":false`) {
		t.Fatalf("manifest omits explicit false reproduction claim: %s", data)
	}
}

func TestDESMATable3ExactPresetStatusExposesClarificationBlockers(t *testing.T) {
	status := newDESMAExactPresetStatus()

	if status.State != desmaExactPresetBlocked || status.Clarification != desmaClarificationPath ||
		!reflect.DeepEqual(status.BlockingQuestionIDs, desmaClarificationBlockerIDs) ||
		len(status.BlockingQuestionIDs) != 6 {
		t.Fatalf("unexpected exact-preset status: %+v", status)
	}

	summary := desmaTable3Summary{
		ComparisonKind:    desmaTable3Comparison,
		ReproductionClaim: false,
		ExactPresetStatus: status,
	}

	data, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("marshal DESMA Table 3 summary: %v", err)
	}

	encoded := string(data)
	for _, required := range append(
		[]string{desmaExactPresetBlocked, desmaClarificationPath},
		desmaClarificationBlockerIDs...,
	) {
		if !strings.Contains(encoded, required) {
			t.Errorf("summary omits %q: %s", required, encoded)
		}
	}

	if !strings.Contains(encoded, `"reproduction_claim":false`) ||
		!strings.Contains(encoded, `"comparison_kind":"descriptive_non_reproduction"`) {
		t.Fatalf("blocked summary mislabels reproduction status: %s", encoded)
	}
}

func TestValidateDESMATable3AccountingPinsCurrentDefaults(t *testing.T) {
	config := mayfly.NewDESMAConfig()

	err := validateDESMATable3Accounting(config)
	if err != nil {
		t.Fatalf("validateDESMATable3Accounting() error = %v", err)
	}

	config.NPop++

	err = validateDESMATable3Accounting(config)
	if err == nil {
		t.Fatal("validateDESMATable3Accounting() accepted changed evaluation accounting")
	}

	initialEvaluations := 20 + 20
	evaluationsPerIteration := 20 + 20 + 20 + 2 + 10

	iterations := (desmaTable3Evaluations - initialEvaluations + evaluationsPerIteration - 1) /
		evaluationsPerIteration
	if iterations != desmaTable3Iterations {
		t.Fatalf("derived iteration ceiling = %d, want %d", iterations, desmaTable3Iterations)
	}
}

func TestSummarizeDESMATable3FunctionComputesMeanAbsoluteError(t *testing.T) {
	runs := make([]mayfly.RunResult, desmaTable3Runs)
	want := 0.0

	for index := range runs {
		errorFromMinimum := float64(index - 25)
		runs[index] = mayfly.RunResult{
			Error:         "",
			BestCost:      -1400 + errorFromMinimum,
			ExecutionTime: 0,
			Seed:          int64(100 + index),
			FuncEvals:     desmaTable3Evaluations,
			Iterations:    desmaTable3Iterations,
			ConvergenceAt: 0,
		}
		want += math.Abs(errorFromMinimum)
	}

	want /= desmaTable3Runs
	result := &mayfly.ComparisonResult{
		FriedmanResult: nil,
		BenchmarkName:  "CEC2013 F1: Sphere",
		AlgorithmNames: []string{"DESMA"},
		RunResults:     [][]mayfly.RunResult{runs},
		Statistics:     nil,
		Rankings:       nil,
		WilcoxonTests:  nil,
		BestAlgorithm:  0,
		BaseSeed:       100,
	}

	summary, err := summarizeDESMATable3Function(
		1, "CEC2013 F1: Sphere", -1400, "f1.csv", "f1.json", result,
	)
	if err != nil {
		t.Fatalf("summarizeDESMATable3Function() error = %v", err)
	}

	if summary.FunctionID != "f1" || summary.FunctionNumber != 1 ||
		summary.MeanAbsoluteError != want || summary.AvailableRuns != desmaTable3Runs ||
		summary.FunctionEvaluationsPerRun != desmaTable3Evaluations {
		t.Fatalf("unexpected function summary: %+v; want mean absolute error %g", summary, want)
	}
}

func TestSummarizeDESMATable3FunctionRejectsInexactProtocol(t *testing.T) {
	validRuns := make([]mayfly.RunResult, desmaTable3Runs)
	for index := range validRuns {
		validRuns[index] = mayfly.RunResult{
			Error:         "",
			BestCost:      float64(index),
			ExecutionTime: 0,
			Seed:          int64(index),
			FuncEvals:     desmaTable3Evaluations,
			Iterations:    desmaTable3Iterations,
			ConvergenceAt: 0,
		}
	}

	tests := []struct {
		name string
		runs []mayfly.RunResult
	}{
		{name: "too few runs", runs: validRuns[:desmaTable3Runs-1]},
		{name: "short evaluation budget", runs: mutateDESMATable3Run(validRuns, func(run *mayfly.RunResult) {
			run.FuncEvals--
		})},
		{name: "failed run", runs: mutateDESMATable3Run(validRuns, func(run *mayfly.RunResult) {
			run.Error = "canceled"
		})},
		{name: "non-finite cost", runs: mutateDESMATable3Run(validRuns, func(run *mayfly.RunResult) {
			run.BestCost = math.Inf(1)
		})},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := &mayfly.ComparisonResult{
				FriedmanResult: nil,
				BenchmarkName:  "F1",
				AlgorithmNames: []string{"DESMA"},
				RunResults:     [][]mayfly.RunResult{test.runs},
				Statistics:     nil,
				Rankings:       nil,
				WilcoxonTests:  nil,
				BestAlgorithm:  0,
				BaseSeed:       0,
			}

			_, err := summarizeDESMATable3Function(1, "F1", -1400, "f1.csv", "f1.json", result)
			if err == nil {
				t.Fatal("summarizeDESMATable3Function() accepted an inexact protocol")
			}
		})
	}
}

func mutateDESMATable3Run(
	runs []mayfly.RunResult,
	mutate func(*mayfly.RunResult),
) []mayfly.RunResult {
	copyOfRuns := append([]mayfly.RunResult(nil), runs...)
	mutate(&copyOfRuns[0])

	return copyOfRuns
}
