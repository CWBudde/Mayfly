package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"

	"github.com/cwbudde/mayfly"
)

const (
	csvPath  = "comparison_results.csv"
	jsonPath = "comparison_results.json"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "comparison failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	workers := min(4, runtime.NumCPU())
	runner := mayfly.NewComparisonRunner().
		WithVariantNames("ma", "desma", "olce", "eobbma", "gsasma", "mpma", "aoblmoa").
		WithRuns(8).
		WithIterations(200).
		WithSeed(42).
		WithParallel(true).
		WithMaxWorkers(workers)

	fmt.Printf("Comparing all Mayfly variants with at most %d concurrent runs...\n", workers)
	result, err := runner.CompareContext(ctx, "Rastrigin-20D", mayfly.Rastrigin, 20, -5.12, 5.12)
	if err != nil {
		return err
	}

	if err := result.WriteComparisonResults(os.Stdout); err != nil {
		return fmt.Errorf("write report: %w", err)
	}
	if err := result.ExportToCSV(csvPath); err != nil {
		return fmt.Errorf("export CSV: %w", err)
	}
	if err := result.ExportToJSON(jsonPath); err != nil {
		return fmt.Errorf("export JSON: %w", err)
	}

	fmt.Printf("\nDetailed results written to %s and %s.\n", csvPath, jsonPath)

	return nil
}
