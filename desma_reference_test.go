package mayfly

import (
	_ "embed"
	"encoding/json"
	"reflect"
	"testing"
)

//go:embed docs/reference-data/desma-2022-table3.json
var desma2022Table3JSON []byte

func TestDESMA2022Table3ReferenceData(t *testing.T) {
	var reference struct {
		ProtocolID        string `json:"protocol_id"`
		ReproductionClaim bool   `json:"reproduction_claim"`
		Source            struct {
			Authors []string `json:"authors"`
			DOI     string   `json:"doi"`
			Table   int      `json:"table"`
		} `json:"source"`
		Execution struct {
			Replications           int  `json:"replications"`
			Dimension              int  `json:"dimension"`
			MaxFunctionEvaluations int  `json:"max_function_evaluations"`
			Iterations             *int `json:"iterations"`
			ReportedPopulationSize int  `json:"reported_population_size"`
			MalePopulationSize     *int `json:"male_population_size"`
			FemalePopulationSize   *int `json:"female_population_size"`
		} `json:"execution"`
		ReportedParameters struct {
			EliteCount          int      `json:"elite_count"`
			EnlargementFactor   float64  `json:"radius_enlargement_factor"`
			ReductionFactor     float64  `json:"radius_reduction_factor"`
			InitialSearchRadius *float64 `json:"initial_search_radius"`
		} `json:"reported_parameters"`
		EliteStrategy struct {
			Equation13                       string `json:"equation_13"`
			Equation14                       string `json:"equation_14"`
			CurrentLibraryReplacementMatches bool   `json:"current_library_replacement_matches"`
		} `json:"elite_strategy"`
		BenchmarkSuite struct {
			FunctionIDs []string `json:"function_ids"`
			Benchmarks  []struct {
				ID      string  `json:"id"`
				Lower   float64 `json:"lower_bound"`
				Upper   float64 `json:"upper_bound"`
				Optimum float64 `json:"optimum"`
			} `json:"benchmarks"`
			SupportingArchive struct {
				SHA256 string `json:"sha256"`
			} `json:"supporting_archive"`
		} `json:"benchmark_suite"`
		StatisticsOrder      []string             `json:"statistics_order"`
		PublishedResults     map[string][]float64 `json:"published_desma_results"`
		PublishedAverageRank map[string]float64   `json:"published_average_ranks"`
		PublishedTTests      struct {
			PrintedOrder              []string `json:"printed_order"`
			ReportedSignificanceLevel float64  `json:"reported_significance_level"`
			MA                        []int    `json:"ma"`
		} `json:"published_t_test_counts"`
		ImplementationFidelityGates []string `json:"implementation_fidelity_gates"`
		UnresolvedProtocolSemantics []string `json:"unresolved_protocol_semantics"`
	}

	if err := json.Unmarshal(desma2022Table3JSON, &reference); err != nil {
		t.Fatalf("decode DESMA Table 3 reference data: %v", err)
	}

	wantAuthors := []string{"Qianhang Du", "Honghao Zhu"}
	if reference.ProtocolID != "desma-2022-table3" || reference.ReproductionClaim ||
		!reflect.DeepEqual(reference.Source.Authors, wantAuthors) ||
		reference.Source.DOI != "10.1371/journal.pone.0273155" || reference.Source.Table != 3 {
		t.Fatalf("unexpected DESMA source metadata: %+v", reference)
	}

	if reference.Execution.Replications != 51 || reference.Execution.Dimension != 30 ||
		reference.Execution.MaxFunctionEvaluations != 300000 ||
		reference.Execution.ReportedPopulationSize != 50 || reference.Execution.Iterations != nil ||
		reference.Execution.MalePopulationSize != nil || reference.Execution.FemalePopulationSize != nil {
		t.Fatalf("unexpected DESMA execution protocol: %+v", reference.Execution)
	}

	if reference.ReportedParameters.EliteCount != 10 ||
		reference.ReportedParameters.EnlargementFactor != 1.05 ||
		reference.ReportedParameters.ReductionFactor != 0.95 ||
		reference.ReportedParameters.InitialSearchRadius != nil {
		t.Fatalf("published DESMA parameters were inferred or changed: %+v", reference.ReportedParameters)
	}

	if reference.EliteStrategy.Equation13 != "r1 = 2*rand(1,n) - 1" ||
		reference.EliteStrategy.Equation14 != "egbest = cgbest + r1*R" ||
		reference.EliteStrategy.CurrentLibraryReplacementMatches {
		t.Fatalf("unexpected DESMA equation audit: %+v", reference.EliteStrategy)
	}

	if len(reference.BenchmarkSuite.FunctionIDs) != 28 || len(reference.BenchmarkSuite.Benchmarks) != 28 ||
		reference.BenchmarkSuite.FunctionIDs[0] != "f1" ||
		reference.BenchmarkSuite.FunctionIDs[27] != "f28" ||
		reference.BenchmarkSuite.SupportingArchive.SHA256 !=
			"7d311e26d5b98ae6bd292ff271be13b0bb929e9f8fc41f3dcc2724363d238bbd" {
		t.Fatalf("unexpected DESMA benchmark metadata: %+v", reference.BenchmarkSuite)
	}

	for index, benchmark := range reference.BenchmarkSuite.Benchmarks {
		wantOptimum := float64(-1400 + index*100)
		if index >= 14 {
			wantOptimum += 100
		}

		if benchmark.ID != reference.BenchmarkSuite.FunctionIDs[index] ||
			benchmark.Lower != -100 || benchmark.Upper != 100 ||
			benchmark.Optimum != wantOptimum {
			t.Errorf("unexpected CEC2013 benchmark at index %d: %+v", index, benchmark)
		}
	}

	wantOrder := []string{"mean_error", "rank"}
	if !reflect.DeepEqual(reference.StatisticsOrder, wantOrder) || len(reference.PublishedResults) != 28 {
		t.Fatalf("unexpected DESMA Table 3 shape: order=%v rows=%d",
			reference.StatisticsOrder, len(reference.PublishedResults))
	}

	for _, functionID := range reference.BenchmarkSuite.FunctionIDs {
		if len(reference.PublishedResults[functionID]) != len(wantOrder) {
			t.Errorf("%s statistic count = %d, want %d",
				functionID, len(reference.PublishedResults[functionID]), len(wantOrder))
		}
	}

	for functionID, want := range map[string][]float64{
		"f1":  {0, 1},
		"f12": {112, 1},
		"f23": {4110, 1},
		"f28": {401, 3},
	} {
		if !reflect.DeepEqual(reference.PublishedResults[functionID], want) {
			t.Errorf("%s DESMA row = %v, want %v",
				functionID, reference.PublishedResults[functionID], want)
		}
	}

	if reference.PublishedAverageRank["desma"] != 2.57 ||
		!reflect.DeepEqual(reference.PublishedTTests.PrintedOrder, []string{"+", "=", "-"}) ||
		reference.PublishedTTests.ReportedSignificanceLevel != 0.05 ||
		!reflect.DeepEqual(reference.PublishedTTests.MA, []int{15, 6, 7}) {
		t.Fatalf("unexpected DESMA Table 3 summaries: ranks=%v t-test=%+v",
			reference.PublishedAverageRank, reference.PublishedTTests)
	}

	if len(reference.ImplementationFidelityGates) < 4 || len(reference.UnresolvedProtocolSemantics) < 6 {
		t.Fatalf("DESMA blockers were not preserved: %v", reference.UnresolvedProtocolSemantics)
	}
}
