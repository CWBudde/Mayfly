package mayfly

import (
	_ "embed"
	"encoding/json"
	"reflect"
	"testing"
)

//go:embed docs/reference-data/aoblmoa-2023-tables7-9.json
var aoblmoa2023Tables7To9JSON []byte

func TestAOBLMOA2023Tables7To9ReferenceData(t *testing.T) {
	var reference struct {
		ProtocolID        string `json:"protocol_id"`
		Status            string `json:"status"`
		ReproductionClaim bool   `json:"reproduction_claim"`
		Source            struct {
			DOI    string `json:"doi"`
			Tables []int  `json:"tables"`
		} `json:"source"`
		AuthorCodeAudit struct {
			Commit                       string   `json:"commit"`
			Files                        []string `json:"files"`
			AvailableBenchmarkEvaluators []string `json:"available_benchmark_evaluators"`
			BatchDriverAvailable         bool     `json:"batch_replication_driver_available"`
			RawRunsAvailable             bool     `json:"raw_run_outputs_available"`
			SeedScheduleAvailable        bool     `json:"published_seed_schedule_available"`
		} `json:"author_code_audit"`
		Execution struct {
			Replications           int   `json:"replications"`
			Iterations             int   `json:"iterations"`
			ReportedPopulationSize int   `json:"reported_population_size"`
			Dimensions             []int `json:"dimensions"`
			MaxFunctionEvaluations *int  `json:"maximum_function_evaluations"`
			SeedSchedule           any   `json:"seed_schedule"`
		} `json:"execution"`
		BenchmarkSuite struct {
			SourceTable int                  `json:"source_table"`
			FunctionIDs []string             `json:"function_ids"`
			Bounds      map[string][]float64 `json:"bounds"`
			Optimum     float64              `json:"optimum"`
		} `json:"benchmark_suite"`
		StatisticsOrder      []string                        `json:"statistics_order"`
		PublishedResults     map[string]map[string][]float64 `json:"published_aoblmoa_results_by_dimension"`
		TranscriptionNotes   []string                        `json:"transcription_notes"`
		ReproductionBlockers []string                        `json:"reproduction_blockers"`
		CompanionArtifact    string                          `json:"companion_reference_artifact"`
		RemainingOutputs     []string                        `json:"remaining_paper_outputs_not_encoded"`
	}

	err := json.Unmarshal(aoblmoa2023Tables7To9JSON, &reference)
	if err != nil {
		t.Fatalf("decode AOBLMOA Tables 7-9 reference data: %v", err)
	}

	if reference.ProtocolID != "aoblmoa-2023-tables7-9" ||
		reference.Status != "source_transcribed_non_reproduction" || reference.ReproductionClaim ||
		reference.Source.DOI != "10.3390/biomimetics8040381" ||
		!reflect.DeepEqual(reference.Source.Tables, []int{7, 8, 9}) {
		t.Fatalf("unexpected AOBLMOA dimension-reference metadata: %+v", reference.Source)
	}

	if reference.AuthorCodeAudit.Commit != "dd3b5b21fc4638cef3c4dde9fc04056296c574e6" ||
		!reflect.DeepEqual(reference.AuthorCodeAudit.Files, []string{"AOBLMOA.m", "main.m"}) ||
		!reflect.DeepEqual(reference.AuthorCodeAudit.AvailableBenchmarkEvaluators, []string{"f1"}) ||
		reference.AuthorCodeAudit.BatchDriverAvailable || reference.AuthorCodeAudit.RawRunsAvailable ||
		reference.AuthorCodeAudit.SeedScheduleAvailable {
		t.Fatalf("unexpected AOBLMOA author-code audit: %+v", reference.AuthorCodeAudit)
	}

	if reference.Execution.Replications != 30 || reference.Execution.Iterations != 1000 ||
		reference.Execution.ReportedPopulationSize != 30 ||
		!reflect.DeepEqual(reference.Execution.Dimensions, []int{30, 50, 100}) ||
		reference.Execution.MaxFunctionEvaluations != nil || reference.Execution.SeedSchedule != nil {
		t.Fatalf("unexpected AOBLMOA dimension protocol: %+v", reference.Execution)
	}

	wantFunctions := []string{"f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f9", "f10"}

	wantOrder := []string{"best", "median", "worst", "mean", "standard_deviation"}
	if reference.BenchmarkSuite.SourceTable != 3 || reference.BenchmarkSuite.Optimum != 0 ||
		!reflect.DeepEqual(reference.BenchmarkSuite.FunctionIDs, wantFunctions) ||
		len(reference.BenchmarkSuite.Bounds) != len(wantFunctions) ||
		!reflect.DeepEqual(reference.StatisticsOrder, wantOrder) {
		t.Fatalf("unexpected benchmark or statistics metadata: suite=%+v order=%v",
			reference.BenchmarkSuite, reference.StatisticsOrder)
	}

	for _, dimension := range []string{"30", "50", "100"} {
		results := reference.PublishedResults[dimension]
		if len(results) != len(wantFunctions) {
			t.Fatalf("dimension %s result rows = %d, want %d", dimension, len(results), len(wantFunctions))
		}

		for _, functionID := range wantFunctions {
			if len(results[functionID]) != len(wantOrder) {
				t.Errorf("dimension %s %s statistic count = %d, want %d",
					dimension, functionID, len(results[functionID]), len(wantOrder))
			}
		}
	}

	for dimension, rows := range map[string]map[string][]float64{
		"30": {
			"f5":  {2.59e-6, 1.68e-5, 0.000145, 3.24e-5, 3.1e-5},
			"f10": {1.35e-32, 1.84e-32, 1.09e-29, 4.02e-31, 1.95e-30},
		},
		"50": {
			"f7":  {8.882e-16, 8.882e-16, 8.882e-16, 8.882e-16, 9.86e-32},
			"f10": {1.97e-19, 1.98e-17, 1.39e-15, 1.92e-16, 5.76e-15},
		},
		"100": {
			"f5": {2.2e-7, 2.18e-5, 0.000102, 3.04e-5, 2.73e-5},
			"f9": {3.28e-13, 4.08e-11, 1.56e-9, 2.94e-10, 3.97e-10},
		},
	} {
		for functionID, want := range rows {
			if !reflect.DeepEqual(reference.PublishedResults[dimension][functionID], want) {
				t.Errorf("dimension %s %s AOBLMOA row = %v, want %v",
					dimension, functionID, reference.PublishedResults[dimension][functionID], want)
			}
		}
	}

	if len(reference.TranscriptionNotes) < 3 || len(reference.ReproductionBlockers) < 3 ||
		reference.CompanionArtifact != "docs/reference-data/aoblmoa-2023-tables5-6.json" ||
		len(reference.RemainingOutputs) != 3 {
		t.Fatalf("AOBLMOA provenance or remaining scope changed: notes=%v blockers=%v companion=%q remaining=%v",
			reference.TranscriptionNotes, reference.ReproductionBlockers,
			reference.CompanionArtifact, reference.RemainingOutputs)
	}
}
