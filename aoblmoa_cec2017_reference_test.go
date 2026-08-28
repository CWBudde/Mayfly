package mayfly

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"testing"
)

//go:embed docs/reference-data/aoblmoa-2023-table13.json
var aoblmoa2023Table13JSON []byte

type aoblmoaTable13PublishedRow struct {
	FunctionID        string     `json:"function_id"`
	Average           []*float64 `json:"average"`
	StandardDeviation []*float64 `json:"standard_deviation"`
	Ranks             []int      `json:"ranks"`
}

func TestAOBLMOA2023Table13ReferenceData(t *testing.T) {
	var reference struct {
		ProtocolID        string `json:"protocol_id"`
		Status            string `json:"status"`
		ReproductionClaim bool   `json:"reproduction_claim"`
		Source            struct {
			DOI   string `json:"doi"`
			Table int    `json:"table"`
		} `json:"source"`
		AuthorCodeAudit struct {
			Commit                     string `json:"commit"`
			CEC2017EvaluatorsAvailable bool   `json:"cec2017_evaluators_available"`
			BatchDriverAvailable       bool   `json:"batch_replication_driver_available"`
			RawRunsAvailable           bool   `json:"raw_run_outputs_available"`
			SeedScheduleAvailable      bool   `json:"published_seed_schedule_available"`
		} `json:"author_code_audit"`
		Execution struct {
			Replications           int `json:"replications_per_algorithm_and_function"`
			Iterations             int `json:"iterations"`
			ReportedPopulationSize int `json:"reported_population_size"`
			Dimension              any `json:"dimension"`
			MaxFunctionEvaluations any `json:"maximum_function_evaluations"`
			SeedSchedule           any `json:"seed_schedule"`
		} `json:"execution"`
		Benchmark struct {
			Suite                   string   `json:"suite_as_printed"`
			PublishedFunctionCount  int      `json:"published_function_count"`
			FunctionIDs             []string `json:"function_ids"`
			PaperIncludesF2         bool     `json:"paper_includes_f2"`
			FinalOfficialIncludesF2 bool     `json:"final_official_suite_includes_f2"`
		} `json:"benchmark"`
		ResultLayout struct {
			AlgorithmOrder    []string `json:"algorithm_order"`
			MeasureOrder      []string `json:"measure_order"`
			MissingValue      any      `json:"missing_numeric_value"`
			SourceMissingMark string   `json:"source_missing_marker"`
		} `json:"result_layout"`
		Rows     []aoblmoaTable13PublishedRow `json:"published_function_rows"`
		Friedman struct {
			PublishedMeanRanks  []float64 `json:"published_mean_ranks"`
			PublishedFinalRanks []int     `json:"published_final_ranks"`
		} `json:"friedman"`
		TranscriptionNotes   []string `json:"transcription_notes"`
		ReproductionBlockers []string `json:"reproduction_blockers"`
		CompanionArtifacts   []string `json:"companion_reference_artifacts"`
		RemainingOutputs     []string `json:"remaining_paper_outputs_not_encoded"`
	}

	err := json.Unmarshal(aoblmoa2023Table13JSON, &reference)
	if err != nil {
		t.Fatalf("decode AOBLMOA Table 13 reference data: %v", err)
	}

	if reference.ProtocolID != "aoblmoa-2023-table13" ||
		reference.Status != "source_transcribed_non_reproduction" || reference.ReproductionClaim ||
		reference.Source.DOI != "10.3390/biomimetics8040381" || reference.Source.Table != 13 {
		t.Fatalf("unexpected AOBLMOA CEC2017 reference metadata: %+v", reference.Source)
	}

	if reference.AuthorCodeAudit.Commit != "dd3b5b21fc4638cef3c4dde9fc04056296c574e6" ||
		reference.AuthorCodeAudit.CEC2017EvaluatorsAvailable || reference.AuthorCodeAudit.BatchDriverAvailable ||
		reference.AuthorCodeAudit.RawRunsAvailable || reference.AuthorCodeAudit.SeedScheduleAvailable {
		t.Fatalf("unexpected AOBLMOA author-code audit: %+v", reference.AuthorCodeAudit)
	}

	if reference.Execution.Replications != 30 || reference.Execution.Iterations != 1000 ||
		reference.Execution.ReportedPopulationSize != 30 || reference.Execution.Dimension != nil ||
		reference.Execution.MaxFunctionEvaluations != nil || reference.Execution.SeedSchedule != nil {
		t.Fatalf("unexpected AOBLMOA CEC2017 protocol: %+v", reference.Execution)
	}

	wantFunctionIDs := make([]string, 30)
	for function := 1; function <= 30; function++ {
		wantFunctionIDs[function-1] = fmt.Sprintf("f%d", function)
	}

	if reference.Benchmark.Suite != "CEC2017BC" || reference.Benchmark.PublishedFunctionCount != 30 ||
		!reflect.DeepEqual(reference.Benchmark.FunctionIDs, wantFunctionIDs) ||
		!reference.Benchmark.PaperIncludesF2 || reference.Benchmark.FinalOfficialIncludesF2 {
		t.Fatalf("unexpected AOBLMOA CEC2017 benchmark metadata: %+v", reference.Benchmark)
	}

	wantAlgorithms := []string{"AOBLMOA", "MOA", "AO", "LGCMFO", "RSA", "ESSA", "SGOA", "COLMA"}
	if !reflect.DeepEqual(reference.ResultLayout.AlgorithmOrder, wantAlgorithms) ||
		!reflect.DeepEqual(reference.ResultLayout.MeasureOrder, []string{"average", "standard_deviation", "rank"}) ||
		reference.ResultLayout.MissingValue != nil || reference.ResultLayout.SourceMissingMark != "-" {
		t.Fatalf("unexpected AOBLMOA Table 13 layout: %+v", reference.ResultLayout)
	}

	if len(reference.Rows) != len(wantFunctionIDs) {
		t.Fatalf("published function rows = %d, want %d", len(reference.Rows), len(wantFunctionIDs))
	}

	rowsByFunction := make(map[string]aoblmoaTable13PublishedRow, len(reference.Rows))
	averageSums := make([]float64, len(wantAlgorithms))
	standardDeviationSums := make([]float64, len(wantAlgorithms))
	numericCounts := make([]int, len(wantAlgorithms))
	rankSums := make([]int, len(wantAlgorithms))

	for rowIndex, row := range reference.Rows {
		if row.FunctionID != wantFunctionIDs[rowIndex] {
			t.Errorf("function row %d = %s, want %s", rowIndex, row.FunctionID, wantFunctionIDs[rowIndex])
		}

		if len(row.Average) != len(wantAlgorithms) || len(row.StandardDeviation) != len(wantAlgorithms) ||
			len(row.Ranks) != len(wantAlgorithms) {
			t.Fatalf("%s widths: average=%d standard deviation=%d ranks=%d", row.FunctionID,
				len(row.Average), len(row.StandardDeviation), len(row.Ranks))
		}

		for algorithmIndex, algorithm := range wantAlgorithms {
			average, deviation := row.Average[algorithmIndex], row.StandardDeviation[algorithmIndex]
			if (average == nil) != (deviation == nil) {
				t.Errorf("%s %s has only one missing numeric measure", row.FunctionID, algorithm)
			}

			if average != nil {
				if *average < 0 || *deviation < 0 || math.IsNaN(*average) || math.IsNaN(*deviation) ||
					math.IsInf(*average, 0) || math.IsInf(*deviation, 0) {
					t.Errorf("%s %s invalid measures: average=%v standard deviation=%v",
						row.FunctionID, algorithm, *average, *deviation)
				}

				averageSums[algorithmIndex] += *average
				standardDeviationSums[algorithmIndex] += *deviation
				numericCounts[algorithmIndex]++
			}

			if row.Ranks[algorithmIndex] < 1 || row.Ranks[algorithmIndex] > len(wantAlgorithms) {
				t.Errorf("%s %s rank = %d", row.FunctionID, algorithm, row.Ranks[algorithmIndex])
			}

			rankSums[algorithmIndex] += row.Ranks[algorithmIndex]
		}

		rowsByFunction[row.FunctionID] = row
	}

	if !reflect.DeepEqual(numericCounts, []int{30, 30, 29, 30, 29, 30, 30, 30}) {
		t.Errorf("numeric result counts = %v", numericCounts)
	}

	if !reflect.DeepEqual(averageSums, []float64{55693, 37615621, 360866, 26700001901104, 370759, 12203391241, 66102360409, 9254532}) {
		t.Errorf("average column checksums = %v", averageSums)
	}

	if !reflect.DeepEqual(standardDeviationSums, []float64{7513, 43493661, 63207, 38400002661750, 34291, 59901794255, 362001247760, 1339573}) {
		t.Errorf("standard-deviation column checksums = %v", standardDeviationSums)
	}

	if !reflect.DeepEqual(rankSums, []int{73, 183, 94, 194, 83, 181, 170, 100}) {
		t.Errorf("rank column checksums = %v", rankSums)
	}

	for algorithmIndex, rankSum := range rankSums {
		got := float64(rankSum) / float64(len(reference.Rows))
		if math.Abs(got-reference.Friedman.PublishedMeanRanks[algorithmIndex]) > 0.0000000005 {
			t.Errorf("%s mean rank from rows = %.12f, published %.9f",
				wantAlgorithms[algorithmIndex], got, reference.Friedman.PublishedMeanRanks[algorithmIndex])
		}
	}

	for algorithmIndex, meanRank := range reference.Friedman.PublishedMeanRanks {
		computedFinalRank := 1

		for _, otherMeanRank := range reference.Friedman.PublishedMeanRanks {
			if otherMeanRank < meanRank {
				computedFinalRank++
			}
		}

		if reference.Friedman.PublishedFinalRanks[algorithmIndex] != computedFinalRank {
			t.Errorf("%s final rank = %d, want %d from published means", wantAlgorithms[algorithmIndex],
				reference.Friedman.PublishedFinalRanks[algorithmIndex], computedFinalRank)
		}
	}

	assertAOBLMOATable13Row(t, rowsByFunction["f2"],
		[]any{200.0, 200.0, nil, 2.67e13, nil, 1.22e10, 6.61e10, 200.0},
		[]any{0.0, 0.0, nil, 3.84e13, nil, 5.99e10, 3.62e11, 0.0},
		[]int{1, 2, 8, 6, 8, 4, 5, 3})
	assertAOBLMOATable13Row(t, rowsByFunction["f14"],
		[]any{1442.0, 4426.0, 1449.0, 39100.0, 1450.0, 15949.0, 1851.0, 2680.0},
		[]any{3346.0, 5846.0, 54.0, 45000.0, 22.0, 14048.0, 214.0, 2260.0},
		[]int{1, 6, 2, 8, 3, 7, 4, 5})
	assertAOBLMOATable13Row(t, rowsByFunction["f28"],
		[]any{3100.0, 3125.0, 3211.0, 3230.0, 2300.0, 3273.0, 3300.0, 3169.0},
		[]any{148.0, 52.0, 47.0, 22.0, 124.0, 25.0, 0.0, 63.0},
		[]int{2, 3, 5, 6, 1, 7, 8, 4})
	assertAOBLMOATable13Row(t, rowsByFunction["f30"],
		[]any{3724.0, 7507.0, 290140.0, 33900.0, 296000.0, 36669.0, 196000.0, 7720.0},
		[]any{151.0, 796.0, 52314.0, 51600.0, 21400.0, 26129.0, 144000.0, 1050.0},
		[]int{1, 2, 7, 4, 8, 5, 6, 3})

	if len(reference.TranscriptionNotes) < 6 || len(reference.ReproductionBlockers) < 5 ||
		len(reference.CompanionArtifacts) != 4 || len(reference.RemainingOutputs) != 0 {
		t.Fatalf("incomplete AOBLMOA Table 13 provenance")
	}
}

func assertAOBLMOATable13Row(
	t *testing.T,
	got aoblmoaTable13PublishedRow,
	wantAverage, wantStandardDeviation []any,
	wantRanks []int,
) {
	t.Helper()

	for measure, pair := range map[string]struct {
		got  []*float64
		want []any
	}{
		"average":            {got.Average, wantAverage},
		"standard deviation": {got.StandardDeviation, wantStandardDeviation},
	} {
		for index, want := range pair.want {
			if want == nil {
				if pair.got[index] != nil {
					t.Errorf("%s %s[%d] = %v, want null", got.FunctionID, measure, index, *pair.got[index])
				}

				continue
			}

			wantValue, ok := want.(float64)
			if !ok {
				t.Fatalf("%s %s[%d] unexpected test value type %T", got.FunctionID, measure, index, want)
			}

			if pair.got[index] == nil || *pair.got[index] != wantValue {
				t.Errorf("%s %s[%d] = %v, want %v", got.FunctionID, measure, index, pair.got[index], want)
			}
		}
	}

	if !reflect.DeepEqual(got.Ranks, wantRanks) {
		t.Errorf("%s ranks = %v, want %v", got.FunctionID, got.Ranks, wantRanks)
	}
}
