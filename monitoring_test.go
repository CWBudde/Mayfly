package mayfly

import (
	"bufio"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestStructuredLoggerReceivesLifecycleEvents(t *testing.T) {
	target := 1.0
	config := lifecycleConfig(func([]float64) float64 { return target })
	config.MaxIterations = 10
	config.Convergence = &ConvergenceConfig{TargetCost: &target}

	var output bytes.Buffer

	logger := slog.New(slog.NewJSONHandler(&output, nil))

	result, err := OptimizeContext(context.Background(), config, WithLogger(logger))
	if err != nil {
		t.Fatalf("OptimizeContext: %v", err)
	}

	var events []map[string]any

	scanner := bufio.NewScanner(&output)
	for scanner.Scan() {
		var event map[string]any

		err = json.Unmarshal(scanner.Bytes(), &event)
		if err != nil {
			t.Fatalf("decode log event: %v", err)
		}

		events = append(events, event)
	}

	err = scanner.Err()
	if err != nil {
		t.Fatalf("scan log output: %v", err)
	}

	if len(events) != 3 {
		t.Fatalf("received %d log events, want start, iteration, and completion", len(events))
	}

	wantEvents := []string{
		eventOptimizationStarted,
		eventIterationCompleted,
		eventOptimizationCompleted,
	}
	for i, want := range wantEvents {
		if events[i]["event"] != want {
			t.Errorf("event %d name = %v, want %q", i, events[i]["event"], want)
		}
	}

	if events[0]["problem_size"] != float64(config.ProblemSize) {
		t.Errorf("start problem_size = %v, want %d", events[0]["problem_size"], config.ProblemSize)
	}

	if events[1]["iteration"] != float64(1) {
		t.Errorf("iteration event iteration = %v, want 1", events[1]["iteration"])
	}

	if events[1]["best_cost"] != target {
		t.Errorf("iteration best_cost = %v, want %v", events[1]["best_cost"], target)
	}

	if events[2]["termination_reason"] != string(TerminationTargetCost) {
		t.Errorf("completion termination_reason = %v, want %q",
			events[2]["termination_reason"], TerminationTargetCost)
	}

	if events[2]["evaluations"] != float64(result.FuncEvalCount) {
		t.Errorf("completion evaluations = %v, want %d", events[2]["evaluations"], result.FuncEvalCount)
	}
}

func TestNilLoggerDisablesLogging(t *testing.T) {
	config := lifecycleConfig(sphere)

	_, err := OptimizeContext(context.Background(), config, WithLogger(nil))
	if err != nil {
		t.Fatalf("OptimizeContext: %v", err)
	}
}

func TestResultExportsConvergenceCurve(t *testing.T) {
	result := &Result{ConvergenceCurve: []float64{3, 2.5, 1}}
	directory := t.TempDir()

	csvPath := filepath.Join(directory, "convergence.csv")

	err := result.ExportConvergenceCSV(csvPath)
	if err != nil {
		t.Fatalf("ExportConvergenceCSV: %v", err)
	}

	csvContents, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("read convergence CSV: %v", err)
	}

	records, err := csv.NewReader(bytes.NewReader(csvContents)).ReadAll()
	if err != nil {
		t.Fatalf("decode convergence CSV: %v", err)
	}

	wantRecords := [][]string{
		{"iteration", "best_cost"},
		{"1", "3"},
		{"2", "2.5"},
		{"3", "1"},
	}
	if !reflect.DeepEqual(records, wantRecords) {
		t.Errorf("CSV records = %v, want %v", records, wantRecords)
	}

	jsonPath := filepath.Join(directory, "convergence.json")

	err = result.ExportConvergenceJSON(jsonPath)
	if err != nil {
		t.Fatalf("ExportConvergenceJSON: %v", err)
	}

	contents, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read convergence JSON: %v", err)
	}

	var points []ConvergencePoint

	err = json.Unmarshal(contents, &points)
	if err != nil {
		t.Fatalf("decode convergence JSON: %v", err)
	}

	wantPoints := []ConvergencePoint{
		{Iteration: 1, BestCost: 3},
		{Iteration: 2, BestCost: 2.5},
		{Iteration: 3, BestCost: 1},
	}
	if !reflect.DeepEqual(points, wantPoints) {
		t.Errorf("JSON points = %+v, want %+v", points, wantPoints)
	}
}

func TestNilResultCannotExportConvergence(t *testing.T) {
	var result *Result

	err := result.ExportConvergenceCSV(filepath.Join(t.TempDir(), "curve.csv"))
	if !errors.Is(err, errNilResult) {
		t.Errorf("ExportConvergenceCSV error = %v, want %v", err, errNilResult)
	}

	err = result.ExportConvergenceJSON(filepath.Join(t.TempDir(), "curve.json"))
	if !errors.Is(err, errNilResult) {
		t.Errorf("ExportConvergenceJSON error = %v, want %v", err, errNilResult)
	}
}
