package mayfly

import (
	_ "embed"
	"encoding/json"
	"reflect"
	"strconv"
	"testing"
)

//go:embed docs/reference-data/mpma-2022-tables1-10.json
var mpma2022Tables1To10JSON []byte

func TestMPMA2022Tables1To10ReferenceData(t *testing.T) {
	var reference struct {
		ProtocolID        string `json:"protocol_id"`
		Status            string `json:"status"`
		ReproductionClaim bool   `json:"reproduction_claim"`
		Source            struct {
			Authors        []string `json:"authors"`
			DOI            string   `json:"doi"`
			ProtocolTables []int    `json:"protocol_tables"`
			ResultTables   []int    `json:"result_tables"`
		} `json:"source"`
		Naming struct {
			PaperAcronym   string `json:"paper_acronym"`
			LibraryAcronym string `json:"library_acronym"`
		} `json:"naming"`
		PublicSourceAudit struct {
			VersionOfRecordAvailable  bool `json:"version_of_record_available"`
			PaperLinkedCodeOrData     bool `json:"paper_linked_code_or_data"`
			PublicImplementationFound bool `json:"public_author_implementation_found"`
			PublicRawRunsFound        bool `json:"public_raw_runs_found"`
			PublicSimulinkModelFound  bool `json:"public_simulink_model_found"`
		} `json:"public_source_audit"`
		BenchmarkExecution struct {
			Replications           int            `json:"replications"`
			MalePopulationSize     int            `json:"male_population_size"`
			FemalePopulationSize   int            `json:"female_population_size"`
			OffspringCount         any            `json:"offspring_count"`
			SeedSchedule           any            `json:"seed_schedule"`
			MaximumFunctionEvals   map[string]int `json:"maximum_function_evaluations"`
			FunctionEvalAccounting any            `json:"function_evaluation_accounting"`
			RawRunsAvailable       bool           `json:"raw_runs_available"`
		} `json:"benchmark_execution"`
		ReportedParameters struct {
			A1            float64 `json:"a1_personal_learning"`
			A2            float64 `json:"a2_global_learning"`
			Beta          float64 `json:"beta_distance_sight"`
			Dance         float64 `json:"nuptial_dance"`
			Flight        float64 `json:"random_flight"`
			GravityStart  float64 `json:"gravity_at_t0"`
			GravityEnd    float64 `json:"gravity_at_maxt"`
			MedianWeight  any     `json:"a4_median_attraction"`
			Crossover     any     `json:"crossover_coefficient_range"`
			MutationSigma any     `json:"mutation_sigma"`
		} `json:"reported_mma_parameters"`
		MedianDefinition struct {
			CoordinateWiseMedian bool `json:"coordinate_wise_population_median"`
			PopulationPool       any  `json:"population_pool"`
		} `json:"median_definition"`
		BenchmarkSuite []struct {
			ID               string  `json:"id"`
			Dimension        int     `json:"dimension"`
			RangePrinted     string  `json:"range_printed"`
			Optimum          float64 `json:"optimum"`
			ResultTable      int     `json:"result_table"`
			MaxFunctionEvals int     `json:"maximum_function_evaluations"`
		} `json:"benchmark_suite"`
		BenchmarkStatisticsOrder  []string             `json:"benchmark_statistics_order"`
		PublishedBenchmarkResults map[string][]float64 `json:"published_mma_benchmark_results"`
		GovernorProtocol          struct {
			DisturbancePercent int                  `json:"frequency_disturbance_percent"`
			SimulationSeconds  int                  `json:"simulation_time_seconds"`
			MaxIterations      int                  `json:"maximum_iterations"`
			Replications       int                  `json:"replications"`
			SeedSchedule       any                  `json:"seed_schedule"`
			Table8Order        []string             `json:"table8_parameter_order"`
			WorkingConditions  map[string][]float64 `json:"table8_working_conditions"`
		} `json:"governor_protocol"`
		GovernorITAEOrder      []string `json:"governor_itae_statistics_order"`
		GovernorIterationOrder []string `json:"governor_optimal_iteration_statistics_order"`
		GovernorTable9         map[string]struct {
			ITAE       []float64 `json:"itae"`
			Iterations []float64 `json:"iterations_to_within_0.1_percent"`
		} `json:"published_mma_governor_table9"`
		GovernorPerformanceOrder []string             `json:"governor_performance_statistics_order"`
		GovernorTable10          map[string][]float64 `json:"published_mma_governor_table10"`
		SourceInconsistencies    []struct {
			ID string `json:"id"`
		} `json:"source_inconsistencies"`
		ExactReproductionGates []struct {
			ID     string `json:"id"`
			Closed bool   `json:"closed"`
		} `json:"exact_reproduction_gates"`
		CurrentLibraryStatus struct {
			ExactPresetAvailable         bool     `json:"exact_preset_available"`
			PublishedComparisonAvailable bool     `json:"published_comparison_available"`
			Extensions                   []string `json:"documented_extensions_or_assumptions"`
		} `json:"current_library_status"`
		FigureOnlyOutputs []string `json:"figure_only_outputs_not_digitized"`
	}

	err := json.Unmarshal(mpma2022Tables1To10JSON, &reference)
	if err != nil {
		t.Fatalf("decode MPMA Tables 1-10 reference data: %v", err)
	}

	wantAuthors := []string{"Guo Lei", "Xu Chang", "Yu Tianhang", "Wumaier Tuerxun"}
	if reference.ProtocolID != "mpma-2022-tables1-10" ||
		reference.Status != "source_transcribed_non_reproduction" || reference.ReproductionClaim ||
		reference.Source.DOI != "10.1109/ACCESS.2022.3160714" ||
		!reflect.DeepEqual(reference.Source.Authors, wantAuthors) ||
		!reflect.DeepEqual(reference.Source.ProtocolTables, []int{1, 2, 3, 4, 8}) ||
		!reflect.DeepEqual(reference.Source.ResultTables, []int{5, 6, 7, 9, 10}) ||
		reference.Naming.PaperAcronym != "MMA" || reference.Naming.LibraryAcronym != "MPMA" {
		t.Fatalf("unexpected MPMA source metadata: source=%+v naming=%+v", reference.Source, reference.Naming)
	}

	audit := reference.PublicSourceAudit
	if !audit.VersionOfRecordAvailable || audit.PaperLinkedCodeOrData ||
		audit.PublicImplementationFound || audit.PublicRawRunsFound || audit.PublicSimulinkModelFound {
		t.Fatalf("unexpected MPMA public-source audit: %+v", audit)
	}

	execution := reference.BenchmarkExecution
	if execution.Replications != 30 || execution.MalePopulationSize != 20 ||
		execution.FemalePopulationSize != 20 || execution.OffspringCount != nil ||
		execution.SeedSchedule != nil || execution.FunctionEvalAccounting != nil ||
		execution.RawRunsAvailable ||
		!reflect.DeepEqual(execution.MaximumFunctionEvals, map[string]int{
			"f1-f10":  100000,
			"f11":     10000,
			"f12-f15": 1000,
			"f16-f18": 2000,
		}) {
		t.Fatalf("unexpected MPMA benchmark protocol: %+v", execution)
	}

	parameters := reference.ReportedParameters
	if parameters.A1 != 1 || parameters.A2 != 1.5 || parameters.Beta != 2 ||
		parameters.Dance != 5 || parameters.Flight != 1 ||
		parameters.GravityStart != 0.9 || parameters.GravityEnd != 0.4 ||
		parameters.MedianWeight != nil || parameters.Crossover != nil || parameters.MutationSigma != nil ||
		reference.MedianDefinition.CoordinateWiseMedian || reference.MedianDefinition.PopulationPool != nil {
		t.Fatalf("unexpected MPMA parameters or median definition: parameters=%+v median=%+v",
			parameters, reference.MedianDefinition)
	}

	wantBenchmarkOrder := []string{"best", "worst", "average", "median", "standard_deviation"}
	if len(reference.BenchmarkSuite) != 18 || len(reference.PublishedBenchmarkResults) != 18 ||
		!reflect.DeepEqual(reference.BenchmarkStatisticsOrder, wantBenchmarkOrder) {
		t.Fatalf("unexpected MPMA benchmark shape: suite=%d rows=%d order=%v",
			len(reference.BenchmarkSuite), len(reference.PublishedBenchmarkResults),
			reference.BenchmarkStatisticsOrder)
	}

	for index, benchmark := range reference.BenchmarkSuite {
		wantID := "f" + strconv.Itoa(index+1)
		wantTable := 7

		switch {
		case index < 5:
			wantTable = 5
		case index < 9:
			wantTable = 6
		}

		if benchmark.ID != wantID {
			t.Errorf("benchmark %d ID = %q, want %q", index, benchmark.ID, wantID)
		}

		if benchmark.Dimension <= 0 || benchmark.RangePrinted == "" ||
			benchmark.ResultTable != wantTable || benchmark.MaxFunctionEvals <= 0 {
			t.Errorf("%s invalid protocol row: %+v", benchmark.ID, benchmark)
		}

		if len(reference.PublishedBenchmarkResults[benchmark.ID]) != len(wantBenchmarkOrder) {
			t.Errorf("%s statistic count = %d, want %d", benchmark.ID,
				len(reference.PublishedBenchmarkResults[benchmark.ID]), len(wantBenchmarkOrder))
		}
	}

	if reference.BenchmarkSuite[9].RangePrinted != "[65.536,65.536]" ||
		reference.BenchmarkSuite[9].Optimum != 0.998 {
		t.Errorf("unexpected Table 3 F10 source row: %+v", reference.BenchmarkSuite[9])
	}

	for functionID, want := range map[string][]float64{
		"f1":  {1.0804e-43, 1.3864e-36, 6.21894e-38, 7.6066e-40, 2.53562e-37},
		"f6":  {-12489.2368, -8656.8406, -11043.75613, -11226.37685, 1071.345463},
		"f15": {-3.862789763, -3.862776763, -3.862780033, -3.862779263, 2.80829e-6},
		"f18": {-10.40280066, -2.765897324, -9.292606962, -10.40294052, 2.551279479},
	} {
		if !reflect.DeepEqual(reference.PublishedBenchmarkResults[functionID], want) {
			t.Errorf("%s MMA row = %v, want %v", functionID,
				reference.PublishedBenchmarkResults[functionID], want)
		}
	}

	governor := reference.GovernorProtocol

	wantTable8Order := []string{"Ty", "Ta", "Tw", "en", "eh", "ey", "eqh", "eqy"}
	if governor.DisturbancePercent != 10 || governor.SimulationSeconds != 20 ||
		governor.MaxIterations != 50 || governor.Replications != 35 || governor.SeedSchedule != nil ||
		!reflect.DeepEqual(governor.Table8Order, wantTable8Order) ||
		!reflect.DeepEqual(governor.WorkingConditions["1"], []float64{0.05, 10.54, 1.6, 1.3, 0.86, 1.12, 0.18, 1.4}) ||
		!reflect.DeepEqual(governor.WorkingConditions["2"], []float64{0.05, 10.54, 2.2, 1.6, 1.43, 0.65, 0.4, 0.85}) {
		t.Fatalf("unexpected MPMA governor protocol: %+v", governor)
	}

	if !reflect.DeepEqual(reference.GovernorITAEOrder, wantBenchmarkOrder) ||
		!reflect.DeepEqual(reference.GovernorIterationOrder, []string{"minimum", "maximum", "average"}) ||
		!reflect.DeepEqual(reference.GovernorPerformanceOrder, []string{"overshoot_percent", "adjustment_time_seconds"}) ||
		!reflect.DeepEqual(reference.GovernorTable9["working_condition_1"].ITAE,
			[]float64{0.205412036, 0.205412056, 0.205412047, 0.205412046, 4.68968e-9}) ||
		!reflect.DeepEqual(reference.GovernorTable9["working_condition_2"].Iterations,
			[]float64{13, 21, 18.143}) ||
		!reflect.DeepEqual(reference.GovernorTable10["working_condition_1"], []float64{3.1, 5.35}) ||
		!reflect.DeepEqual(reference.GovernorTable10["working_condition_2"], []float64{3.2, 8.135}) {
		t.Fatalf("unexpected MPMA governor outputs: table9=%+v table10=%+v",
			reference.GovernorTable9, reference.GovernorTable10)
	}

	wantInconsistencyIDs := []string{
		"foxholes_range_missing_minus",
		"f18_best_median_order",
		"governor_success_rate_missing",
	}

	gotInconsistencyIDs := make([]string, len(reference.SourceInconsistencies))
	for index, inconsistency := range reference.SourceInconsistencies {
		gotInconsistencyIDs[index] = inconsistency.ID
	}

	wantGateIDs := []string{
		"a4_median_attraction",
		"median_population_pool",
		"genetic_operator_parameters",
		"initialization_and_boundaries",
		"evaluation_accounting",
		"seeds_and_raw_runs",
		"governor_model",
		"source_inconsistencies",
	}

	gotGateIDs := make([]string, len(reference.ExactReproductionGates))
	for index, gate := range reference.ExactReproductionGates {
		gotGateIDs[index] = gate.ID
		if gate.Closed {
			t.Errorf("exact-reproduction gate %q is closed without source evidence", gate.ID)
		}
	}

	if !reflect.DeepEqual(gotInconsistencyIDs, wantInconsistencyIDs) ||
		!reflect.DeepEqual(gotGateIDs, wantGateIDs) ||
		reference.CurrentLibraryStatus.ExactPresetAvailable ||
		reference.CurrentLibraryStatus.PublishedComparisonAvailable ||
		len(reference.CurrentLibraryStatus.Extensions) != 4 || len(reference.FigureOnlyOutputs) != 4 {
		t.Fatalf("unexpected MPMA blockers/status: inconsistencies=%v gates=%v status=%+v figures=%v",
			gotInconsistencyIDs, gotGateIDs, reference.CurrentLibraryStatus, reference.FigureOnlyOutputs)
	}
}
