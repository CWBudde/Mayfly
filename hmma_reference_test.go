package mayfly

import (
	_ "embed"
	"encoding/json"
	"reflect"
	"testing"
)

//go:embed docs/reference-data/hmma-2022-table1.json
var hmma2022Table1JSON []byte

func TestHMMA2022Table1ReferenceData(t *testing.T) {
	var reference struct {
		ProtocolID        string `json:"protocol_id"`
		ReproductionClaim bool   `json:"reproduction_claim"`
		Source            struct {
			DOI   string `json:"doi"`
			Table int    `json:"table"`
		} `json:"source"`
		Execution struct {
			Replications   int  `json:"replications"`
			Iterations     int  `json:"iterations"`
			PopulationSize *int `json:"population_size"`
		} `json:"execution"`
		ReportedParameters struct {
			Symbols []string  `json:"symbols"`
			Values  []float64 `json:"values"`
		} `json:"reported_parameter_tuple"`
		Equation10 struct {
			Expression                    string `json:"expression"`
			CurrentLibraryScheduleMatches bool   `json:"current_library_schedule_matches"`
		} `json:"published_equation_10"`
		Benchmarks []struct {
			ID        string   `json:"id"`
			Dimension *int     `json:"dimension"`
			Lower     *float64 `json:"lower_bound"`
			Upper     *float64 `json:"upper_bound"`
		} `json:"benchmarks"`
		StatisticsOrder             []string                        `json:"statistics_order"`
		PublishedResults            map[string]map[string][]float64 `json:"published_results"`
		UnresolvedProtocolSemantics []string                        `json:"unresolved_protocol_semantics"`
	}

	err := json.Unmarshal(hmma2022Table1JSON, &reference)
	if err != nil {
		t.Fatalf("decode HMMA Table 1 reference data: %v", err)
	}

	if reference.ProtocolID != "hmma-2022-table1" || reference.ReproductionClaim ||
		reference.Source.DOI != "10.1049/ell2.12568" || reference.Source.Table != 1 ||
		reference.Execution.Replications != 50 || reference.Execution.Iterations != 1000 ||
		reference.Execution.PopulationSize != nil {
		t.Fatalf("unexpected HMMA protocol metadata: %+v", reference)
	}

	wantSymbols := []string{
		"a1", "a2", "a3", "fl", "g", "ub", "lb", "theta", "d",
		"fl_damp", "g_damp", "d_damp", "rho",
	}
	if !reflect.DeepEqual(reference.ReportedParameters.Symbols, wantSymbols) ||
		len(reference.ReportedParameters.Values) != len(wantSymbols) ||
		reference.ReportedParameters.Values[5] != 0.1 ||
		reference.ReportedParameters.Values[6] != 10 ||
		reference.ReportedParameters.Values[7] != 0.005 {
		t.Fatalf("published parameter tuple was normalized or reordered: %+v",
			reference.ReportedParameters)
	}

	if reference.Equation10.Expression != "Ps = -exp((1 - t / Iter_MAX)^20) + theta" ||
		reference.Equation10.CurrentLibraryScheduleMatches {
		t.Fatalf("unexpected Equation 10 audit: %+v", reference.Equation10)
	}

	if len(reference.Benchmarks) != 3 || reference.Benchmarks[0].Dimension != nil ||
		reference.Benchmarks[1].Dimension != nil || reference.Benchmarks[2].Dimension == nil ||
		*reference.Benchmarks[2].Dimension != 4 {
		t.Fatalf("missing benchmark-dimension ambiguity: %+v", reference.Benchmarks)
	}

	for _, benchmark := range reference.Benchmarks {
		if benchmark.Lower != nil || benchmark.Upper != nil {
			t.Errorf("benchmark %s silently inferred bounds", benchmark.ID)
		}
	}

	wantOrder := []string{"best", "worst", "average", "standard_deviation", "median"}
	if !reflect.DeepEqual(reference.StatisticsOrder, wantOrder) {
		t.Fatalf("statistics order = %v, want %v", reference.StatisticsOrder, wantOrder)
	}

	for _, benchmarkID := range []string{"F3", "F7", "F15"} {
		rows := reference.PublishedResults[benchmarkID]
		for _, algorithm := range []string{"ma", "ima", "amma", "ocma", "hmma"} {
			if len(rows[algorithm]) != len(wantOrder) {
				t.Errorf("%s %s statistic count = %d, want %d",
					benchmarkID, algorithm, len(rows[algorithm]), len(wantOrder))
			}
		}
	}

	wantHMMAF3 := []float64{1.5321e-155, 3.0399e-123, 6.0807e-125, 4.2558e-124, 4.4956e-143}
	if !reflect.DeepEqual(reference.PublishedResults["F3"]["hmma"], wantHMMAF3) {
		t.Errorf("F3 HMMA row = %v, want %v", reference.PublishedResults["F3"]["hmma"], wantHMMAF3)
	}

	if len(reference.UnresolvedProtocolSemantics) < 6 {
		t.Fatalf("HMMA blockers were not preserved: %v", reference.UnresolvedProtocolSemantics)
	}
}
