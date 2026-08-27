package mayfly

import (
	_ "embed"
	"encoding/json"
	"reflect"
	"testing"
)

//go:embed docs/reference-data/aoblmoa-2023-tables5-6.json
var aoblmoa2023Tables5And6JSON []byte

func TestAOBLMOA2023Tables5And6ReferenceData(t *testing.T) {
	var reference struct {
		ProtocolID        string `json:"protocol_id"`
		Status            string `json:"status"`
		ReproductionClaim bool   `json:"reproduction_claim"`
		Source            struct {
			Authors []string `json:"authors"`
			DOI     string   `json:"doi"`
			Tables  []int    `json:"tables"`
		} `json:"source"`
		AuthorCodeAudit struct {
			Commit                       string   `json:"commit"`
			Files                        []string `json:"files"`
			MalePopulation               int      `json:"main_script_male_population"`
			FemalePopulation             int      `json:"main_script_female_population"`
			OffspringCount               int      `json:"main_script_offspring_count"`
			AvailableBenchmarkEvaluators []string `json:"available_benchmark_evaluators"`
			BatchDriverAvailable         bool     `json:"batch_replication_driver_available"`
			RawRunsAvailable             bool     `json:"raw_run_outputs_available"`
			SeedScheduleAvailable        bool     `json:"published_seed_schedule_available"`
		} `json:"author_code_audit"`
		Execution struct {
			Replications           int  `json:"replications"`
			Iterations             int  `json:"iterations"`
			ReportedPopulationSize int  `json:"reported_population_size"`
			MaxFunctionEvaluations *int `json:"maximum_function_evaluations"`
			SeedSchedule           any  `json:"seed_schedule"`
		} `json:"execution"`
		ReportedParameters struct {
			A1       float64 `json:"a1"`
			A2       float64 `json:"a2"`
			A3       float64 `json:"a3"`
			GInitial float64 `json:"g_initial"`
			GFinal   float64 `json:"g_final"`
			Alpha    float64 `json:"alpha"`
			Delta    float64 `json:"delta"`
		} `json:"reported_parameters"`
		BenchmarkSuite struct {
			Benchmarks []struct {
				ID          string  `json:"id"`
				Dimension   int     `json:"dimension"`
				LowerBound  float64 `json:"lower_bound"`
				UpperBound  float64 `json:"upper_bound"`
				Optimum     float64 `json:"optimum"`
				ResultTable int     `json:"result_table"`
			} `json:"benchmarks"`
		} `json:"benchmark_suite"`
		StatisticsOrder             []string             `json:"statistics_order"`
		ReportedTimeUnit            any                  `json:"reported_time_unit"`
		PublishedResults            map[string][]float64 `json:"published_aoblmoa_results"`
		ImplementationFidelityGates []struct {
			ID     string `json:"id"`
			Closed bool   `json:"closed"`
		} `json:"implementation_fidelity_gates"`
		UnresolvedProtocolSemantics []string `json:"unresolved_protocol_semantics"`
		RemainingOutputs            []string `json:"remaining_paper_outputs_not_encoded_here"`
	}

	err := json.Unmarshal(aoblmoa2023Tables5And6JSON, &reference)
	if err != nil {
		t.Fatalf("decode AOBLMOA Tables 5-6 reference data: %v", err)
	}

	wantAuthors := []string{"Yanpu Zhao", "Changsheng Huang", "Mengjie Zhang", "Yang Cui"}
	if reference.ProtocolID != "aoblmoa-2023-tables5-6" ||
		reference.Status != "source_transcribed_non_reproduction" || reference.ReproductionClaim ||
		reference.Source.DOI != "10.3390/biomimetics8040381" ||
		!reflect.DeepEqual(reference.Source.Authors, wantAuthors) ||
		!reflect.DeepEqual(reference.Source.Tables, []int{5, 6}) {
		t.Fatalf("unexpected AOBLMOA source metadata: %+v", reference.Source)
	}

	if reference.AuthorCodeAudit.Commit != "dd3b5b21fc4638cef3c4dde9fc04056296c574e6" ||
		!reflect.DeepEqual(reference.AuthorCodeAudit.Files, []string{"AOBLMOA.m", "main.m"}) ||
		reference.AuthorCodeAudit.MalePopulation != 30 ||
		reference.AuthorCodeAudit.FemalePopulation != 30 ||
		reference.AuthorCodeAudit.OffspringCount != 30 ||
		!reflect.DeepEqual(reference.AuthorCodeAudit.AvailableBenchmarkEvaluators, []string{"f1"}) ||
		reference.AuthorCodeAudit.BatchDriverAvailable || reference.AuthorCodeAudit.RawRunsAvailable ||
		reference.AuthorCodeAudit.SeedScheduleAvailable {
		t.Fatalf("unexpected AOBLMOA author-code audit: %+v", reference.AuthorCodeAudit)
	}

	if reference.Execution.Replications != 30 || reference.Execution.Iterations != 1000 ||
		reference.Execution.ReportedPopulationSize != 30 ||
		reference.Execution.MaxFunctionEvaluations != nil || reference.Execution.SeedSchedule != nil {
		t.Fatalf("unexpected AOBLMOA execution protocol: %+v", reference.Execution)
	}

	parameters := reference.ReportedParameters
	if parameters.A1 != 1 || parameters.A2 != 1.5 || parameters.A3 != 1.5 ||
		parameters.GInitial != 0.9 || parameters.GFinal != 0.4 ||
		parameters.Alpha != 0.1 || parameters.Delta != 0.1 {
		t.Fatalf("published AOBLMOA parameters changed: %+v", parameters)
	}

	if len(reference.BenchmarkSuite.Benchmarks) != 19 || len(reference.PublishedResults) != 19 {
		t.Fatalf("unexpected AOBLMOA benchmark shape: benchmarks=%d rows=%d",
			len(reference.BenchmarkSuite.Benchmarks), len(reference.PublishedResults))
	}

	wantOrder := []string{"best", "median", "worst", "mean", "standard_deviation", "reported_time"}
	if !reflect.DeepEqual(reference.StatisticsOrder, wantOrder) || reference.ReportedTimeUnit != nil {
		t.Fatalf("unexpected AOBLMOA statistics metadata: order=%v time_unit=%v",
			reference.StatisticsOrder, reference.ReportedTimeUnit)
	}

	seen := make(map[string]bool, len(reference.BenchmarkSuite.Benchmarks))
	for index, benchmark := range reference.BenchmarkSuite.Benchmarks {
		if seen[benchmark.ID] {
			t.Errorf("duplicate benchmark ID %q", benchmark.ID)
		}

		seen[benchmark.ID] = true

		wantTable := 5
		if index >= 10 {
			wantTable = 6
		}

		if benchmark.ResultTable != wantTable {
			t.Errorf("%s result table = %d, want %d", benchmark.ID, benchmark.ResultTable, wantTable)
		}

		if len(reference.PublishedResults[benchmark.ID]) != len(wantOrder) {
			t.Errorf("%s statistic count = %d, want %d",
				benchmark.ID, len(reference.PublishedResults[benchmark.ID]), len(wantOrder))
		}
	}

	for functionID, want := range map[string][]float64{
		"f1":  {0, 0, 0, 0, 0, 0.146875},
		"f5":  {5.63e-7, 2.931e-5, 0.000106, 3.76e-5, 2.79e-5, 0.1895833},
		"f12": {-1.03163, -1.03163, -1.03163, -1.03163, 0, 0.190104},
		"f19": {-10.5364, -10.5364, -10.5364, -10.5364, 2.57e-14, 0.202604},
	} {
		if !reflect.DeepEqual(reference.PublishedResults[functionID], want) {
			t.Errorf("%s AOBLMOA row = %v, want %v",
				functionID, reference.PublishedResults[functionID], want)
		}
	}

	wantGateIDs := []string{
		"gravity_schedule",
		"aquila_mean_semantics",
		"levy_scale_conflict",
		"author_rng_order",
	}

	gotGateIDs := make([]string, len(reference.ImplementationFidelityGates))
	for index, gate := range reference.ImplementationFidelityGates {
		gotGateIDs[index] = gate.ID
		if gate.Closed {
			t.Errorf("fidelity gate %q is marked closed without a reproducible resolution", gate.ID)
		}
	}

	if !reflect.DeepEqual(gotGateIDs, wantGateIDs) ||
		len(reference.UnresolvedProtocolSemantics) < 7 || len(reference.RemainingOutputs) != 4 {
		t.Fatalf("AOBLMOA blockers or remaining scope changed: gates=%v blockers=%v remaining=%v",
			gotGateIDs, reference.UnresolvedProtocolSemantics, reference.RemainingOutputs)
	}
}
