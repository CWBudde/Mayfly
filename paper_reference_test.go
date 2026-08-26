package mayfly

import (
	_ "embed"
	"encoding/json"
	"testing"
)

//go:embed docs/reference-data/original-ma-2020-table6.json
var originalMA2020Table6JSON []byte

func TestOriginalMA2020Table6ReferenceData(t *testing.T) {
	var reference struct {
		ProtocolID string `json:"protocol_id"`
		AuthorCode struct {
			DatasetURL           string `json:"dataset_url"`
			Version1DOI          string `json:"version_1_doi"`
			DeclaredScope        string `json:"declared_scope"`
			Table6ExperimentCode bool   `json:"table_6_experiment_code"`
			AuditedVersions      []struct {
				Version      int    `json:"version"`
				MatlabSHA256 string `json:"matlab_sha256"`
			} `json:"audited_versions"`
			Version1 struct {
				Dimension                      int     `json:"dimension"`
				MutantsPerIteration            int     `json:"mutants_per_iteration"`
				MutationCoordinateFraction     float64 `json:"mutation_coordinate_fraction"`
				MutationStandardDeviationRange float64 `json:"mutation_standard_deviation_fraction_of_search_range"`
				ActualFunctionEvaluations      int     `json:"actual_function_evaluations_at_2000_iterations"`
				DisplayedFunctionEvaluations   int     `json:"displayed_function_evaluations_at_2000_iterations"`
			} `json:"version_1_observations"`
		} `json:"author_code_archive"`
		Execution struct {
			Replications        int `json:"replications"`
			FunctionEvaluations int `json:"function_evaluations_per_replication"`
			MalePopulation      int `json:"male_population"`
			FemalePopulation    int `json:"female_population"`
		} `json:"execution"`
		Benchmarks []struct {
			ID        string  `json:"id"`
			Name      string  `json:"name"`
			Dimension int     `json:"dimension"`
			Lower     float64 `json:"lower_bound"`
			Upper     float64 `json:"upper_bound"`
		} `json:"benchmarks"`
		PublishedResults map[string]map[string][]float64 `json:"published_results"`
	}

	err := json.Unmarshal(originalMA2020Table6JSON, &reference)
	if err != nil {
		t.Fatalf("decode original MA Table 6 reference data: %v", err)
	}

	if reference.ProtocolID != "original-ma-2020-table6" ||
		reference.Execution.Replications != 50 ||
		reference.Execution.FunctionEvaluations != 95_000 ||
		reference.Execution.MalePopulation != 20 ||
		reference.Execution.FemalePopulation != 20 {
		t.Fatalf("unexpected execution protocol: %+v", reference.Execution)
	}

	if reference.AuthorCode.DatasetURL != "https://data.mendeley.com/datasets/5w58s8hhz2" ||
		reference.AuthorCode.Version1DOI != "10.17632/5w58s8hhz2.1" ||
		reference.AuthorCode.DeclaredScope == "" || reference.AuthorCode.Table6ExperimentCode ||
		len(reference.AuthorCode.AuditedVersions) != 4 ||
		reference.AuthorCode.AuditedVersions[2].MatlabSHA256 !=
			reference.AuthorCode.AuditedVersions[3].MatlabSHA256 ||
		reference.AuthorCode.Version1.Dimension != 50 ||
		reference.AuthorCode.Version1.MutantsPerIteration != 1 ||
		reference.AuthorCode.Version1.MutationCoordinateFraction != 0.01 ||
		reference.AuthorCode.Version1.MutationStandardDeviationRange != 0.1 ||
		reference.AuthorCode.Version1.ActualFunctionEvaluations != 122_040 ||
		reference.AuthorCode.Version1.DisplayedFunctionEvaluations != 120_040 {
		t.Fatalf("unexpected author-code audit: %+v", reference.AuthorCode)
	}

	wantBenchmarks := map[string]struct {
		name      string
		dimension int
		lower     float64
		upper     float64
	}{
		"F1":  {name: "Sphere", dimension: 5, lower: -10, upper: 10},
		"F2":  {name: "Rosenbrock", dimension: 5, lower: -5, upper: 10},
		"F10": {name: "Rastrigin", dimension: 5, lower: -5.12, upper: 5.12},
		"F11": {name: "Ackley", dimension: 5, lower: -32, upper: 32},
		"F19": {name: "Eggcrate", dimension: 2, lower: -5, upper: 5},
		"F20": {name: "Beale", dimension: 2, lower: -4.5, upper: 4.5},
	}

	if len(reference.Benchmarks) != len(wantBenchmarks) {
		t.Fatalf("benchmark count = %d, want %d", len(reference.Benchmarks), len(wantBenchmarks))
	}

	for _, benchmark := range reference.Benchmarks {
		want, ok := wantBenchmarks[benchmark.ID]
		if !ok {
			t.Fatalf("unexpected benchmark ID %q", benchmark.ID)
		}

		if benchmark.Name != want.name || benchmark.Dimension != want.dimension ||
			benchmark.Lower != want.lower || benchmark.Upper != want.upper {
			t.Errorf("benchmark %s = %+v, want %+v", benchmark.ID, benchmark, want)
		}

		rows, ok := reference.PublishedResults[benchmark.ID]
		if !ok {
			t.Errorf("benchmark %s has no published result rows", benchmark.ID)
			continue
		}

		for _, algorithm := range []string{"basic_ma", "vgma", "sma", "ima"} {
			if got := len(rows[algorithm]); got != 5 {
				t.Errorf("%s %s statistic count = %d, want 5", benchmark.ID, algorithm, got)
			}
		}
	}

	wantIMAF2 := []float64{0, 1.7995e-29, 2.0798e-30, 0, 4.3249e-30}
	for index, want := range wantIMAF2 {
		if got := reference.PublishedResults["F2"]["ima"][index]; got != want {
			t.Errorf("F2 IMA statistic %d = %g, want %g", index, got, want)
		}
	}
}
