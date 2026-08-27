package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func csvRows(path string) ([][]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return csv.NewReader(file).ReadAll()
}

func TestParseOptions(t *testing.T) {
	opts, err := parseOptions([]string{
		"-output", "results", "-benchmarks", "sphere,ackley", "-variants", "standard,olce",
		"-dimensions", "2,10,2", "-runs", "3", "-iterations", "4", "-workers", "2", "-seed", "7",
		"-max-evaluations", "90",
	})
	if err != nil {
		t.Fatalf("parseOptions() error = %v", err)
	}

	if !reflect.DeepEqual(opts.benchmarks, []string{"sphere", "ackley"}) {
		t.Fatalf("benchmarks = %v", opts.benchmarks)
	}

	if !reflect.DeepEqual(opts.variants, []string{"ma", "olce-ma"}) {
		t.Fatalf("variants = %v", opts.variants)
	}

	if !reflect.DeepEqual(opts.dimensions, []int{2, 10}) {
		t.Fatalf("dimensions = %v", opts.dimensions)
	}

	if opts.runs != 3 || opts.iterations != 4 || opts.maxEvals != 90 || opts.workers != 2 || opts.seed != 7 {
		t.Fatalf("unexpected scalar options: %+v", opts)
	}
}

func TestParseOptionsRejectsUnknownNames(t *testing.T) {
	_, err := parseOptions([]string{"-benchmarks", "missing"})
	if err == nil {
		t.Fatal("parseOptions() accepted an unknown benchmark")
	}

	_, err = parseOptions([]string{"-variants", "missing"})
	if err == nil {
		t.Fatal("parseOptions() accepted an unknown variant")
	}
}

func TestParseOptionsPublishedReferenceUsesOnlyCurrentMA(t *testing.T) {
	opts, err := parseOptions([]string{
		"-published-reference", "reference.json",
		"-runs", "1", "-iterations", "2",
	})
	if err != nil {
		t.Fatalf("parseOptions() error = %v", err)
	}

	if opts.publishedReference != "reference.json" ||
		!reflect.DeepEqual(opts.variants, []string{"ma"}) ||
		len(opts.benchmarks) != 0 || len(opts.dimensions) != 0 {
		t.Fatalf("unexpected published-reference options: %+v", opts)
	}

	_, err = parseOptions([]string{
		"-published-reference", "reference.json",
		"-variants", "ma,desma",
	})
	if err == nil {
		t.Fatal("parseOptions() accepted a historical comparison against a non-historical variant")
	}
}

func TestRunExperimentWritesAllVariantRowsAndManifest(t *testing.T) {
	outputDir := t.TempDir()
	opts := options{
		outputDir:          outputDir,
		publishedReference: "",
		desmaTable3Data:    "",
		benchmarks:         []string{"sphere"},
		variants:           []string{"ma", "desma", "olce-ma", "eobbma", "gsasma", "hmma", "mpma", "aoblmoa"},
		dimensions:         []int{2},
		runs:               1,
		iterations:         1,
		maxEvals:           0,
		workers:            1,
		seed:               42,
	}

	err := runExperiment(context.Background(), opts, io.Discard)
	if err != nil {
		t.Fatalf("runExperiment() error = %v", err)
	}

	rows, err := csvRows(filepath.Join(outputDir, "sphere-2d.csv"))
	if err != nil {
		t.Fatalf("read result CSV: %v", err)
	}

	if got, want := len(rows), 9; got != want {
		t.Fatalf("CSV rows = %d, want %d", got, want)
	}

	for index, row := range rows[1:] {
		if row[4] != "42" {
			t.Fatalf("row %d seed = %q, want 42", index+1, row[4])
		}
	}

	data, err := os.ReadFile(filepath.Join(outputDir, "manifest.json"))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}

	var protocol manifest

	err = json.Unmarshal(data, &protocol)
	if err != nil {
		t.Fatalf("decode manifest: %v", err)
	}

	if protocol.SchemaVersion != schemaVersion || len(protocol.Variants) != 8 || len(protocol.Benchmarks) != 1 {
		t.Fatalf("unexpected manifest: %+v", protocol)
	}

	if protocol.Benchmarks[0].ResultCSV != "sphere-2d.csv" {
		t.Fatalf("result CSV = %q", protocol.Benchmarks[0].ResultCSV)
	}
}

func TestRunPublishedReferenceComparisonWritesExplicitNonReproductionSummary(t *testing.T) {
	outputDir := t.TempDir()
	opts := options{
		outputDir:          outputDir,
		publishedReference: filepath.Join("..", "..", "docs", "reference-data", "original-ma-2020-table6.json"),
		desmaTable3Data:    "",
		benchmarks:         nil,
		variants:           []string{"ma"},
		dimensions:         nil,
		runs:               1,
		iterations:         1,
		maxEvals:           0,
		workers:            1,
		seed:               42,
	}

	err := runExperiment(context.Background(), opts, io.Discard)
	if err != nil {
		t.Fatalf("runExperiment() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outputDir, "published-comparison.json"))
	if err != nil {
		t.Fatalf("read published comparison: %v", err)
	}

	var summary publishedComparisonSummary

	err = json.Unmarshal(data, &summary)
	if err != nil {
		t.Fatalf("decode published comparison: %v", err)
	}

	if summary.ReproductionClaim || summary.ComparisonKind != "descriptive_non_reproduction" ||
		len(summary.Comparisons) != 6 {
		t.Fatalf("unexpected published comparison: %+v", summary)
	}

	_, err = os.Stat(filepath.Join(outputDir, "f19-eggcrate-2d.csv"))
	if err != nil {
		t.Fatalf("reference-specific fixed-dimension result is missing: %v", err)
	}

	manifestData, err := os.ReadFile(filepath.Join(outputDir, "manifest.json"))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}

	var protocol manifest

	err = json.Unmarshal(manifestData, &protocol)
	if err != nil {
		t.Fatalf("decode manifest: %v", err)
	}

	if len(protocol.Variants) != 1 || protocol.Variants[0].Name != "MA" ||
		len(protocol.Benchmarks) != 6 {
		t.Fatalf("unexpected published-reference manifest: %+v", protocol)
	}
}
