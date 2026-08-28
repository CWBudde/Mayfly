package mayfly

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"testing"
)

//go:embed docs/reference-data/aoblmoa-2023-tables10-11.json
var aoblmoa2023Tables10To11JSON []byte

func TestAOBLMOA2023Tables10To11ReferenceData(t *testing.T) {
	var reference struct {
		ProtocolID        string `json:"protocol_id"`
		Status            string `json:"status"`
		ReproductionClaim bool   `json:"reproduction_claim"`
		Source            struct {
			DOI    string `json:"doi"`
			Tables []int  `json:"tables"`
		} `json:"source"`
		AuthorCodeAudit struct {
			Commit                    string `json:"commit"`
			BatchDriverAvailable      bool   `json:"batch_replication_driver_available"`
			StatisticsDriverAvailable bool   `json:"statistical_analysis_driver_available"`
			RawRunsAvailable          bool   `json:"raw_run_outputs_available"`
			SeedScheduleAvailable     bool   `json:"published_seed_schedule_available"`
		} `json:"author_code_audit"`
		Execution struct {
			Replications           int  `json:"replications_per_algorithm_and_case"`
			Iterations             int  `json:"iterations"`
			ReportedPopulationSize int  `json:"reported_population_size"`
			MaxFunctionEvaluations *int `json:"maximum_function_evaluations"`
			SeedSchedule           any  `json:"seed_schedule"`
		} `json:"execution"`
		BenchmarkCases struct {
			Count                      int `json:"count"`
			VariableDimensionFunctions struct {
				FunctionIDs []string `json:"function_ids"`
				Dimensions  []int    `json:"dimensions"`
			} `json:"variable_dimension_functions"`
			FixedDimensions map[string]int `json:"fixed_dimensions"`
		} `json:"benchmark_cases"`
		Wilcoxon struct {
			SourceTable        int              `json:"source_table"`
			ReferenceAlgorithm string           `json:"reference_algorithm"`
			ComparatorOrder    []string         `json:"comparator_order"`
			Threshold          float64          `json:"significance_threshold"`
			SummaryOrder       []string         `json:"summary_symbol_order_as_printed"`
			SummaryCounts      map[string][]int `json:"published_summary_counts"`
		} `json:"wilcoxon_rank_sum"`
		Friedman struct {
			SourceTable         int       `json:"source_table"`
			AlgorithmOrder      []string  `json:"algorithm_order"`
			PublishedMeanRanks  []float64 `json:"published_mean_ranks"`
			PublishedFinalRanks []int     `json:"published_final_ranks"`
		} `json:"friedman"`
		Rows []struct {
			FunctionID      string    `json:"function_id"`
			Dimension       int       `json:"dimension"`
			WilcoxonPValues []float64 `json:"wilcoxon_p_values"`
			FriedmanRanks   []int     `json:"friedman_ranks"`
		} `json:"published_case_rows"`
		TranscriptionNotes   []string `json:"transcription_notes"`
		ReproductionBlockers []string `json:"reproduction_blockers"`
		CompanionArtifacts   []string `json:"companion_reference_artifacts"`
		RemainingOutputs     []string `json:"remaining_paper_outputs_not_encoded"`
	}

	err := json.Unmarshal(aoblmoa2023Tables10To11JSON, &reference)
	if err != nil {
		t.Fatalf("decode AOBLMOA Tables 10-11 reference data: %v", err)
	}

	if reference.ProtocolID != "aoblmoa-2023-tables10-11" ||
		reference.Status != "source_transcribed_non_reproduction" || reference.ReproductionClaim ||
		reference.Source.DOI != "10.3390/biomimetics8040381" ||
		!reflect.DeepEqual(reference.Source.Tables, []int{10, 11}) {
		t.Fatalf("unexpected AOBLMOA statistics-reference metadata: %+v", reference.Source)
	}

	if reference.AuthorCodeAudit.Commit != "dd3b5b21fc4638cef3c4dde9fc04056296c574e6" ||
		reference.AuthorCodeAudit.BatchDriverAvailable ||
		reference.AuthorCodeAudit.StatisticsDriverAvailable ||
		reference.AuthorCodeAudit.RawRunsAvailable || reference.AuthorCodeAudit.SeedScheduleAvailable {
		t.Fatalf("unexpected AOBLMOA author-code audit: %+v", reference.AuthorCodeAudit)
	}

	if reference.Execution.Replications != 30 || reference.Execution.Iterations != 1000 ||
		reference.Execution.ReportedPopulationSize != 30 ||
		reference.Execution.MaxFunctionEvaluations != nil || reference.Execution.SeedSchedule != nil {
		t.Fatalf("unexpected AOBLMOA statistical protocol: %+v", reference.Execution)
	}

	wantVariableFunctions := []string{"f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f9", "f10"}

	wantFixedDimensions := map[string]int{
		"f11": 4, "f12": 2, "f13": 2, "f14": 2, "f15": 3,
		"f16": 6, "f17": 4, "f18": 4, "f19": 4,
	}

	if reference.BenchmarkCases.Count != 49 ||
		!reflect.DeepEqual(reference.BenchmarkCases.VariableDimensionFunctions.FunctionIDs, wantVariableFunctions) ||
		!reflect.DeepEqual(reference.BenchmarkCases.VariableDimensionFunctions.Dimensions, []int{10, 30, 50, 100}) ||
		!reflect.DeepEqual(reference.BenchmarkCases.FixedDimensions, wantFixedDimensions) {
		t.Fatalf("unexpected AOBLMOA benchmark cases: %+v", reference.BenchmarkCases)
	}

	wantComparators := []string{"AO", "MOA", "AMOA", "OBLAO", "OBLMOA"}

	wantAlgorithms := append(append([]string(nil), wantComparators...), "AOBLMOA")

	if reference.Wilcoxon.SourceTable != 10 || reference.Wilcoxon.ReferenceAlgorithm != "AOBLMOA" ||
		!reflect.DeepEqual(reference.Wilcoxon.ComparatorOrder, wantComparators) ||
		reference.Wilcoxon.Threshold != 0.05 ||
		!reflect.DeepEqual(reference.Wilcoxon.SummaryOrder, []string{"+", "-", "="}) ||
		reference.Friedman.SourceTable != 11 ||
		!reflect.DeepEqual(reference.Friedman.AlgorithmOrder, wantAlgorithms) {
		t.Fatalf("unexpected statistical-test metadata: wilcoxon=%+v friedman=%+v",
			reference.Wilcoxon, reference.Friedman)
	}

	expectedCases := make([]string, 0, reference.BenchmarkCases.Count)

	for _, functionID := range wantVariableFunctions {
		for _, dimension := range []int{10, 30, 50, 100} {
			expectedCases = append(expectedCases, fmt.Sprintf("%s/%d", functionID, dimension))
		}
	}

	for functionNumber := 11; functionNumber <= 19; functionNumber++ {
		functionID := fmt.Sprintf("f%d", functionNumber)
		expectedCases = append(expectedCases, fmt.Sprintf("%s/%d", functionID, wantFixedDimensions[functionID]))
	}

	if len(reference.Rows) != len(expectedCases) {
		t.Fatalf("published case rows = %d, want %d", len(reference.Rows), len(expectedCases))
	}

	significantCounts := make([]int, len(wantComparators))
	rankSums := make([]int, len(wantAlgorithms))

	rowsByCase := make(map[string]struct {
		PValues []float64
		Ranks   []int
	}, len(reference.Rows))
	for rowIndex, row := range reference.Rows {
		caseID := fmt.Sprintf("%s/%d", row.FunctionID, row.Dimension)

		if caseID != expectedCases[rowIndex] {
			t.Errorf("case row %d = %s, want %s", rowIndex, caseID, expectedCases[rowIndex])
		}

		if len(row.WilcoxonPValues) != len(wantComparators) || len(row.FriedmanRanks) != len(wantAlgorithms) {
			t.Fatalf("case %s widths: p-values=%d ranks=%d", caseID,
				len(row.WilcoxonPValues), len(row.FriedmanRanks))
		}

		for algorithmIndex, pValue := range row.WilcoxonPValues {
			if pValue < 0 || pValue > 1 {
				t.Errorf("case %s comparator %s p-value = %g", caseID, wantComparators[algorithmIndex], pValue)
			}

			if pValue < reference.Wilcoxon.Threshold {
				significantCounts[algorithmIndex]++
			}
		}

		for algorithmIndex, rank := range row.FriedmanRanks {
			if rank < 1 || rank > len(wantAlgorithms) {
				t.Errorf("case %s algorithm %s rank = %d", caseID, wantAlgorithms[algorithmIndex], rank)
			}

			rankSums[algorithmIndex] += rank
		}

		if row.FriedmanRanks[len(row.FriedmanRanks)-1] != 1 {
			t.Errorf("case %s AOBLMOA rank = %d, want 1", caseID, row.FriedmanRanks[len(row.FriedmanRanks)-1])
		}

		rowsByCase[caseID] = struct {
			PValues []float64
			Ranks   []int
		}{row.WilcoxonPValues, row.FriedmanRanks}
	}

	for algorithmIndex, algorithm := range wantComparators {
		wantSummary := []int{0, significantCounts[algorithmIndex], len(reference.Rows) - significantCounts[algorithmIndex]}
		if !reflect.DeepEqual(reference.Wilcoxon.SummaryCounts[algorithm], wantSummary) {
			t.Errorf("%s summary = %v, want %v", algorithm,
				reference.Wilcoxon.SummaryCounts[algorithm], wantSummary)
		}
	}

	for algorithmIndex, rankSum := range rankSums {
		got := float64(rankSum) / float64(len(reference.Rows))
		if math.Abs(got-reference.Friedman.PublishedMeanRanks[algorithmIndex]) > 0.0000005 {
			t.Errorf("%s mean rank from rows = %.9f, published %.6f",
				wantAlgorithms[algorithmIndex], got, reference.Friedman.PublishedMeanRanks[algorithmIndex])
		}
	}

	if !reflect.DeepEqual(reference.Friedman.PublishedFinalRanks, []int{5, 6, 4, 2, 3, 1}) {
		t.Errorf("published final ranks = %v", reference.Friedman.PublishedFinalRanks)
	}

	for caseID, want := range map[string]struct {
		PValues []float64
		Ranks   []int
	}{
		"f5/100": {[]float64{0.42843, 1.73e-6, 0.338843, 0.557743, 0.002585}, []int{3, 6, 4, 2, 5, 1}},
		"f13/2":  {[]float64{0.000453, 0.25, 0.25, 0.000241, 0.25}, []int{6, 1, 1, 5, 1, 1}},
		"f19/4":  {[]float64{1.73e-6, 0.003906, 3.79e-6, 1.73e-6, 1}, []int{3, 5, 6, 2, 4, 1}},
	} {
		got := rowsByCase[caseID]
		if !reflect.DeepEqual(got.PValues, want.PValues) || !reflect.DeepEqual(got.Ranks, want.Ranks) {
			t.Errorf("case %s = p-values %v, ranks %v; want %v, %v",
				caseID, got.PValues, got.Ranks, want.PValues, want.Ranks)
		}
	}

	if len(reference.TranscriptionNotes) < 4 || len(reference.ReproductionBlockers) < 4 ||
		!reflect.DeepEqual(reference.CompanionArtifacts, []string{
			"docs/reference-data/aoblmoa-2023-tables5-6.json",
			"docs/reference-data/aoblmoa-2023-tables7-9.json",
		}) || len(reference.RemainingOutputs) != 2 {
		t.Fatalf("AOBLMOA provenance or remaining scope changed: notes=%v blockers=%v companions=%v remaining=%v",
			reference.TranscriptionNotes, reference.ReproductionBlockers,
			reference.CompanionArtifacts, reference.RemainingOutputs)
	}
}
