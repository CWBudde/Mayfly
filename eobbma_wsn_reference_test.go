package mayfly

import (
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"hash/fnv"
	"math"
	"reflect"
	"testing"
)

//go:embed docs/reference-data/eobbma-2025-tables2-8.json
var eobbma2025Tables2To8JSON []byte

type eobbmaPublishedCaseResults struct {
	Case    int                  `json:"case"`
	Results map[string][]float64 `json:"results"`
}

func TestEOBBMA2025Tables2To8ReferenceData(t *testing.T) {
	var reference struct {
		ProtocolID        string `json:"protocol_id"`
		Status            string `json:"status"`
		ReproductionClaim bool   `json:"reproduction_claim"`
		VersionOfRecord   struct {
			DOI             string            `json:"doi"`
			PublishedOnline string            `json:"published_online"`
			IssueYear       int               `json:"issue_year"`
			Volume          int               `json:"volume"`
			Issue           int               `json:"issue"`
			Pages           string            `json:"pages"`
			TableURLs       map[string]string `json:"table_urls"`
		} `json:"version_of_record"`
		AuthorManuscript struct {
			DOI                    string `json:"doi"`
			FormalCrossrefRelation bool   `json:"formal_crossref_relation_to_version_of_record"`
			UsedForTranscription   bool   `json:"used_for_numeric_transcription"`
			ProtocolComparisonDone bool   `json:"full_protocol_comparison_complete"`
		} `json:"author_manuscript"`
		Execution struct {
			Algorithms         []string `json:"algorithms"`
			PopulationSize     int      `json:"population_size_as_printed"`
			PopulationSexSplit any      `json:"population_sex_split"`
			Iterations         int      `json:"iterations"`
			IndependentRuns    any      `json:"independent_runs"`
			SeedSchedule       any      `json:"seed_schedule"`
			RawRunsAvailable   bool     `json:"raw_runs_available"`
			RawCoordinates     bool     `json:"raw_sensor_coordinates_available"`
			MaxFunctionEvals   any      `json:"maximum_function_evaluations"`
			ObjectiveWeights   any      `json:"objective_weights"`
		} `json:"execution"`
		ProblemMapping struct {
			SensorSymbol       string `json:"number_of_sensors_symbol"`
			SearchDimension    string `json:"search_dimension"`
			CandidateSemantics string `json:"candidate_position_semantics"`
			OptimalSemantics   string `json:"optimal_candidate_semantics"`
			FitnessSemantics   string `json:"fitness_semantics"`
		} `json:"table_2_problem_mapping"`
		AlgorithmParameters map[string][]string `json:"table_3_algorithm_parameters"`
		Cases               []struct {
			Case      int   `json:"case"`
			Radius    int   `json:"radius_m"`
			Sensors   int   `json:"sensors"`
			Dimension int   `json:"dimension"`
			Region    []int `json:"region_m"`
		} `json:"table_4_cases"`
		Table5 struct {
			ValueOrder []string                     `json:"value_order"`
			Cases      []eobbmaPublishedCaseResults `json:"cases"`
		} `json:"table_5"`
		Table6 struct {
			ValueOrder []string                     `json:"value_order"`
			Cases      []eobbmaPublishedCaseResults `json:"cases"`
		} `json:"table_6"`
		Table7 struct {
			Baseline        string   `json:"baseline"`
			ComparatorOrder []string `json:"comparator_order"`
			PValues         []struct {
				Case   int      `json:"case"`
				Values []string `json:"values"`
			} `json:"p_values_as_printed"`
		} `json:"table_7"`
		Table8 struct {
			AlgorithmOrder   []string  `json:"algorithm_order"`
			FriedmanMeanRank []float64 `json:"friedman_mean_rank"`
			Rank             []float64 `json:"rank"`
		} `json:"table_8"`
		FigureOnlyOutputs    []any    `json:"figure_only_outputs"`
		TranscriptionNotes   []string `json:"transcription_notes"`
		ReproductionBlockers []string `json:"reproduction_blockers"`
	}

	err := json.Unmarshal(eobbma2025Tables2To8JSON, &reference)
	if err != nil {
		t.Fatalf("decode EOBBMA reference: %v", err)
	}

	if reference.ProtocolID != "eobbma-2025-tables2-8" ||
		reference.Status != "source_transcribed_non_reproduction" || reference.ReproductionClaim {
		t.Fatalf("unexpected EOBBMA reference status")
	}

	vor := reference.VersionOfRecord
	if vor.DOI != "10.1007/s13369-024-08899-6" || vor.PublishedOnline != "2024-03-25" ||
		vor.IssueYear != 2025 || vor.Volume != 50 || vor.Issue != 2 || vor.Pages != "719-739" ||
		len(vor.TableURLs) != 7 {
		t.Fatalf("unexpected version-of-record metadata: %+v", vor)
	}

	manuscript := reference.AuthorManuscript
	if manuscript.DOI != "10.2139/ssrn.4381249" || manuscript.FormalCrossrefRelation ||
		manuscript.UsedForTranscription || manuscript.ProtocolComparisonDone {
		t.Fatalf("unexpected author-manuscript provenance: %+v", manuscript)
	}

	wantAlgorithms := []string{"EOBBMA", "MA", "DE", "PSO", "FPA", "GWO", "SSA", "COA"}
	if !reflect.DeepEqual(reference.Execution.Algorithms, wantAlgorithms) ||
		reference.Execution.PopulationSize != 40 || reference.Execution.Iterations != 1000 ||
		reference.Execution.PopulationSexSplit != nil || reference.Execution.IndependentRuns != nil ||
		reference.Execution.SeedSchedule != nil || reference.Execution.RawRunsAvailable ||
		reference.Execution.RawCoordinates || reference.Execution.MaxFunctionEvals != nil ||
		reference.Execution.ObjectiveWeights != nil {
		t.Fatalf("unexpected execution protocol: %+v", reference.Execution)
	}

	if reference.ProblemMapping.SensorSymbol != "k" || reference.ProblemMapping.SearchDimension != "2*k" ||
		reference.ProblemMapping.CandidateSemantics != "sensor positions on a two-dimensional plane" ||
		reference.ProblemMapping.OptimalSemantics != "optimal sensor deployment scheme" ||
		reference.ProblemMapping.FitnessSemantics != "minimum value obtained by Equation 17" {
		t.Fatalf("unexpected Table 2 problem mapping: %+v", reference.ProblemMapping)
	}

	wantAlgorithmParameters := map[string][]string{
		"EOBBMA": {"population size = 40"},
		"MA":     {"population size = 40", "alpha1 = 1", "alpha2 = 1.5", "beta = 2", "d = 0.1", "fl = 0.1"},
		"DE":     {"population size = 40", "F = 0.5", "CR = 0.5"},
		"PSO":    {"population size = 40", "g_max = 0.9", "g_min = 0.4", "C1 = 0.8", "C2 = 0.4"},
		"FPA":    {"population size = 40", "p = 0.5"},
		"GWO":    {"population size = 40", "a ranges from 2 to 0"},
		"SSA":    {"population size = 40", "C1 ranges from 2 to 0", "C2 and C3 are random numbers between 0 and 1"},
		"COA":    {"population size = 40"},
	}
	if !reflect.DeepEqual(reference.AlgorithmParameters, wantAlgorithmParameters) {
		t.Fatalf("unexpected Table 3 parameters: %+v", reference.AlgorithmParameters)
	}

	wantSensors := []int{20, 25, 40, 100, 160, 240, 300, 500}
	wantRadii := []int{3, 5, 5, 5, 5, 6, 7, 8}
	wantRegions := []int{20, 40, 50, 80, 100, 150, 200, 300}

	if len(reference.Cases) != 8 {
		t.Fatalf("deployment case count = %d, want 8", len(reference.Cases))
	}

	for index, scenario := range reference.Cases {
		if scenario.Case != index+1 || scenario.Sensors != wantSensors[index] ||
			scenario.Radius != wantRadii[index] || scenario.Dimension != 2*scenario.Sensors ||
			!reflect.DeepEqual(scenario.Region, []int{wantRegions[index], wantRegions[index]}) {
			t.Errorf("unexpected deployment case %d: %+v", index+1, scenario)
		}
	}

	var (
		wantTable5Order = []string{"best", "worst", "average", "standard_deviation", "rank"}
		wantTable6Order = []string{"coverage_average", "coverage_standard_deviation", "redundancy_average", "redundancy_standard_deviation", "moving_distance_average", "moving_distance_standard_deviation"}
	)

	if !reflect.DeepEqual(reference.Table5.ValueOrder, wantTable5Order) ||
		!reflect.DeepEqual(reference.Table6.ValueOrder, wantTable6Order) ||
		len(reference.Table5.Cases) != 8 || len(reference.Table6.Cases) != 8 {
		t.Fatalf("unexpected result layouts")
	}

	hash := fnv.New64a()
	hashFloat := func(value float64) {
		var encoded [8]byte
		binary.LittleEndian.PutUint64(encoded[:], math.Float64bits(value))
		_, _ = hash.Write(encoded[:])
	}

	for caseIndex := range 8 {
		var (
			fitnessCase = reference.Table5.Cases[caseIndex]
			metricsCase = reference.Table6.Cases[caseIndex]
		)

		if fitnessCase.Case != caseIndex+1 || metricsCase.Case != caseIndex+1 ||
			len(fitnessCase.Results) != 8 || len(metricsCase.Results) != 8 {
			t.Fatalf("unexpected result case %d", caseIndex+1)
		}

		seenRanks := make([]bool, 9)

		for _, algorithm := range wantAlgorithms {
			var (
				fitness = fitnessCase.Results[algorithm]
				metrics = metricsCase.Results[algorithm]
			)

			if len(fitness) != 5 || len(metrics) != 6 {
				t.Fatalf("case %d %s result width", caseIndex+1, algorithm)
			}

			if fitness[0] > fitness[2] || fitness[2] > fitness[1] || fitness[3] < 0 ||
				fitness[4] < 1 || fitness[4] > 8 || fitness[4] != math.Trunc(fitness[4]) {
				t.Errorf("case %d %s invalid fitness summary %v", caseIndex+1, algorithm, fitness)
			}

			rank := int(fitness[4])
			if seenRanks[rank] {
				t.Errorf("case %d duplicate rank %d", caseIndex+1, rank)
			}

			seenRanks[rank] = true

			if algorithm == "EOBBMA" && rank != 1 {
				t.Errorf("case %d EOBBMA rank = %d, want 1", caseIndex+1, rank)
			}

			if metrics[0] < 0 || metrics[0] > 1 || metrics[1] < 0 ||
				metrics[2] < 0 || metrics[2] > 1 || metrics[3] < 0 || metrics[4] < 0 || metrics[5] < 0 {
				t.Errorf("case %d %s invalid deployment metrics %v", caseIndex+1, algorithm, metrics)
			}

			for _, value := range fitness {
				hashFloat(value)
			}

			for _, value := range metrics {
				hashFloat(value)
			}
		}
	}

	wantComparators := wantAlgorithms[1:]
	if reference.Table7.Baseline != "EOBBMA" || !reflect.DeepEqual(reference.Table7.ComparatorOrder, wantComparators) ||
		len(reference.Table7.PValues) != 8 {
		t.Fatalf("unexpected Table 7 layout")
	}

	for caseIndex, result := range reference.Table7.PValues {
		if result.Case != caseIndex+1 || len(result.Values) != 7 {
			t.Fatalf("unexpected Table 7 case %d", caseIndex+1)
		}

		for _, value := range result.Values {
			_, _ = hash.Write([]byte(value))
			_, _ = hash.Write([]byte{0})
		}
	}

	if reference.Table7.PValues[0].Values[4] != "1.0354–04" {
		t.Errorf("Table 7 malformed source cell was normalized")
	}

	if !reflect.DeepEqual(reference.Table8.AlgorithmOrder, wantAlgorithms) ||
		len(reference.Table8.FriedmanMeanRank) != 8 || len(reference.Table8.Rank) != 8 {
		t.Fatalf("unexpected Table 8 layout")
	}

	meanRankSum := 0.0
	seenOverallRanks := make([]bool, 9)

	for index := range 8 {
		meanRankSum += reference.Table8.FriedmanMeanRank[index]

		rank := reference.Table8.Rank[index]
		if rank < 1 || rank > 8 || rank != math.Trunc(rank) || seenOverallRanks[int(rank)] {
			t.Errorf("invalid overall rank %v", rank)
		} else {
			seenOverallRanks[int(rank)] = true
		}

		hashFloat(reference.Table8.FriedmanMeanRank[index])
		hashFloat(rank)
	}

	if meanRankSum != 36 || reference.Table8.Rank[0] != 1 {
		t.Errorf("unexpected Friedman summary")
	}

	const wantHash = uint64(0x940ddfddab4a6a0e)
	if got := hash.Sum64(); got != wantHash {
		t.Errorf("published value hash = %#x, want %#x", got, wantHash)
	}

	if len(reference.FigureOnlyOutputs) != 4 || len(reference.TranscriptionNotes) < 7 ||
		len(reference.ReproductionBlockers) < 9 {
		t.Fatalf("incomplete EOBBMA provenance")
	}
}
