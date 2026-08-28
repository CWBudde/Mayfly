package mayfly

import (
	_ "embed"
	"encoding/json"
	"math"
	"reflect"
	"testing"
)

//go:embed docs/reference-data/aoblmoa-2023-tables14-23.json
var aoblmoa2023Tables14To23JSON []byte

type aoblmoaCEC2020PublishedProblem struct {
	Table                              int       `json:"table"`
	PaperAcronym                       string    `json:"paper_acronym"`
	OfficialID                         string    `json:"official_id"`
	OfficialName                       string    `json:"official_name"`
	OfficialDimension                  int       `json:"official_dimension"`
	OfficialInequalityConstraints      int       `json:"official_inequality_constraints"`
	OfficialEqualityConstraints        int       `json:"official_equality_constraints"`
	OfficialMaximumFunctionEvaluations int       `json:"official_maximum_function_evaluations"`
	Average                            []float64 `json:"average"`
	StandardDeviation                  []float64 `json:"standard_deviation"`
	Median                             []float64 `json:"median"`
	Minimum                            []float64 `json:"minimum"`
	Maximum                            []float64 `json:"maximum"`
}

func TestAOBLMOA2023Tables14To23ReferenceData(t *testing.T) {
	var reference struct {
		ProtocolID        string `json:"protocol_id"`
		Status            string `json:"status"`
		ReproductionClaim bool   `json:"reproduction_claim"`
		Source            struct {
			DOI    string `json:"doi"`
			Tables []int  `json:"tables"`
		} `json:"source"`
		AuthorCodeAudit struct {
			Commit                 string `json:"commit"`
			CEC2020Available       bool   `json:"cec2020_evaluators_available"`
			PenaltyDriverAvailable bool   `json:"constraint_penalty_driver_available"`
			BatchDriverAvailable   bool   `json:"batch_replication_driver_available"`
			RawRunsAvailable       bool   `json:"raw_run_outputs_available"`
			SeedScheduleAvailable  bool   `json:"published_seed_schedule_available"`
		} `json:"author_code_audit"`
		Execution struct {
			Replications           int    `json:"replications_per_algorithm_and_problem"`
			ConstraintHandling     string `json:"constraint_handling_as_printed"`
			PenaltyDefinition      any    `json:"static_penalty_definition"`
			PenaltyCoefficient     any    `json:"static_penalty_coefficient"`
			PopulationSize         any    `json:"reported_population_size"`
			Iterations             any    `json:"iterations"`
			MaxFunctionEvaluations any    `json:"maximum_function_evaluations"`
			SeedSchedule           any    `json:"seed_schedule"`
			ValueSemantics         any    `json:"reported_value_semantics"`
		} `json:"execution"`
		OfficialContext struct {
			ArchiveCommit              string   `json:"archive_commit"`
			GuidelineReplications      int      `json:"guideline_replications"`
			RequiredFinalReporting     []string `json:"required_final_reporting"`
			PaperReportsBudget         bool     `json:"paper_reports_budget_adherence"`
			PaperReportsViolation      bool     `json:"paper_reports_constraint_violation"`
			PaperReportsFeasibility    bool     `json:"paper_reports_feasibility_rate"`
			PaperReportsViolationCount bool     `json:"paper_reports_violation_count_vector"`
			ComparatorRawRuns          bool     `json:"comparator_raw_outputs_present_in_archive"`
			AOBLMOARawRuns             bool     `json:"aoblmoa_raw_outputs_present_in_archive"`
		} `json:"official_cec2020_context"`
		ResultLayout struct {
			AlgorithmOrder []string `json:"algorithm_order"`
			MeasureOrder   []string `json:"measure_order"`
		} `json:"result_layout"`
		Problems             []aoblmoaCEC2020PublishedProblem `json:"problems"`
		TranscriptionNotes   []string                         `json:"transcription_notes"`
		ReproductionBlockers []string                         `json:"reproduction_blockers"`
		CompanionArtifacts   []string                         `json:"companion_reference_artifacts"`
		RemainingOutputs     []string                         `json:"remaining_aoblmoa_paper_outputs_not_encoded"`
	}

	err := json.Unmarshal(aoblmoa2023Tables14To23JSON, &reference)
	if err != nil {
		t.Fatalf("decode AOBLMOA Tables 14-23 reference data: %v", err)
	}

	if reference.ProtocolID != "aoblmoa-2023-tables14-23" ||
		reference.Status != "source_transcribed_non_reproduction" || reference.ReproductionClaim ||
		reference.Source.DOI != "10.3390/biomimetics8040381" ||
		!reflect.DeepEqual(reference.Source.Tables, []int{14, 15, 16, 17, 18, 19, 20, 21, 22, 23}) {
		t.Fatalf("unexpected AOBLMOA CEC2020 reference metadata: %+v", reference.Source)
	}

	if reference.AuthorCodeAudit.Commit != "dd3b5b21fc4638cef3c4dde9fc04056296c574e6" ||
		reference.AuthorCodeAudit.CEC2020Available || reference.AuthorCodeAudit.PenaltyDriverAvailable ||
		reference.AuthorCodeAudit.BatchDriverAvailable || reference.AuthorCodeAudit.RawRunsAvailable ||
		reference.AuthorCodeAudit.SeedScheduleAvailable {
		t.Fatalf("unexpected AOBLMOA author-code audit: %+v", reference.AuthorCodeAudit)
	}

	if reference.Execution.Replications != 25 || reference.Execution.ConstraintHandling != "static penalty function" ||
		reference.Execution.PenaltyDefinition != nil || reference.Execution.PenaltyCoefficient != nil ||
		reference.Execution.PopulationSize != nil || reference.Execution.Iterations != nil ||
		reference.Execution.MaxFunctionEvaluations != nil || reference.Execution.SeedSchedule != nil ||
		reference.Execution.ValueSemantics != nil {
		t.Fatalf("unexpected AOBLMOA constrained protocol: %+v", reference.Execution)
	}

	wantReporting := []string{
		"objective value",
		"constraint violation",
		"feasibility rate",
		"median-solution violation-count vector",
	}
	if reference.OfficialContext.ArchiveCommit != "a148874a233f9fb8fa82fdae3b5739d1bba7020d" ||
		reference.OfficialContext.GuidelineReplications != 25 ||
		!reflect.DeepEqual(reference.OfficialContext.RequiredFinalReporting, wantReporting) ||
		reference.OfficialContext.PaperReportsBudget || reference.OfficialContext.PaperReportsViolation ||
		reference.OfficialContext.PaperReportsFeasibility || reference.OfficialContext.PaperReportsViolationCount ||
		!reference.OfficialContext.ComparatorRawRuns || reference.OfficialContext.AOBLMOARawRuns {
		t.Fatalf("unexpected official CEC2020 context: %+v", reference.OfficialContext)
	}

	wantAlgorithms := []string{"AOBLMOA", "SASS", "COLSHADE", "sCMAgES"}

	wantMeasures := []string{"average", "standard_deviation", "median", "minimum", "maximum"}
	if !reflect.DeepEqual(reference.ResultLayout.AlgorithmOrder, wantAlgorithms) ||
		!reflect.DeepEqual(reference.ResultLayout.MeasureOrder, wantMeasures) {
		t.Fatalf("unexpected AOBLMOA Tables 14-23 layout: %+v", reference.ResultLayout)
	}

	wantProblems := []struct {
		table, dimension, inequalities, equalities, evaluations int
		acronym, officialID, name                               string
	}{
		{14, 7, 11, 0, 100000, "WMSR", "RC15", "Weight Minimization of a Speed Reducer"},
		{15, 14, 15, 0, 200000, "ODIRS", "RC16", "Optimal Design of Industrial Refrigeration System"},
		{16, 3, 3, 0, 100000, "TCSD1", "RC17", "Tension/compression spring design (case 1)"},
		{17, 5, 7, 0, 100000, "MDCBDP", "RC21", "Multiple disk clutch brake design problem"},
		{18, 9, 10, 1, 100000, "PGTDO", "RC22", "Planetary gear train design optimization problem"},
		{19, 4, 7, 0, 100000, "HTBDP", "RC25", "Hydro-static thrust bearing design problem"},
		{20, 22, 86, 0, 200000, "FGBP", "RC26", "Four-stage gear box problem"},
		{21, 4, 1, 0, 100000, "GTCD", "RC29", "Gas Transmission Compressor Design"},
		{22, 3, 8, 0, 100000, "TCSD2", "RC30", "Tension/compression spring design (case 2)"},
		{23, 30, 30, 0, 200000, "TO", "RC33", "Topology Optimization"},
	}

	if len(reference.Problems) != len(wantProblems) {
		t.Fatalf("published constrained problems = %d, want %d", len(reference.Problems), len(wantProblems))
	}

	const (
		fnvOffset64 = uint64(14695981039346656037)
		fnvPrime64  = uint64(1099511628211)
		wantHash    = uint64(0xfa858413fcb4c244)
	)

	valueHash := fnvOffset64

	for index, problem := range reference.Problems {
		want := wantProblems[index]
		if problem.Table != want.table || problem.PaperAcronym != want.acronym ||
			problem.OfficialID != want.officialID || problem.OfficialName != want.name ||
			problem.OfficialDimension != want.dimension ||
			problem.OfficialInequalityConstraints != want.inequalities ||
			problem.OfficialEqualityConstraints != want.equalities ||
			problem.OfficialMaximumFunctionEvaluations != want.evaluations {
			t.Errorf("problem %d metadata = %+v", index, problem)
		}

		measures := [][]float64{
			problem.Average,
			problem.StandardDeviation,
			problem.Median,
			problem.Minimum,
			problem.Maximum,
		}
		for measureIndex, values := range measures {
			if len(values) != len(wantAlgorithms) {
				t.Fatalf("%s %s width = %d, want %d", problem.PaperAcronym,
					wantMeasures[measureIndex], len(values), len(wantAlgorithms))
			}

			for _, value := range values {
				if value < 0 || math.IsNaN(value) || math.IsInf(value, 0) {
					t.Errorf("%s %s contains invalid value %v", problem.PaperAcronym,
						wantMeasures[measureIndex], value)
				}

				bits := math.Float64bits(value)
				for byteIndex := range 8 {
					valueHash ^= uint64(byte(bits >> (8 * byteIndex)))
					valueHash *= fnvPrime64
				}
			}
		}

		for algorithmIndex, algorithm := range wantAlgorithms {
			if problem.Minimum[algorithmIndex] > problem.Median[algorithmIndex] ||
				problem.Median[algorithmIndex] > problem.Maximum[algorithmIndex] ||
				problem.Average[algorithmIndex] < problem.Minimum[algorithmIndex] ||
				problem.Average[algorithmIndex] > problem.Maximum[algorithmIndex] {
				t.Errorf("%s %s inconsistent location statistics", problem.PaperAcronym, algorithm)
			}
		}
	}

	if valueHash != wantHash {
		t.Errorf("published value hash = %#x, want %#x", valueHash, wantHash)
	}

	if len(reference.TranscriptionNotes) < 7 || len(reference.ReproductionBlockers) < 6 ||
		len(reference.CompanionArtifacts) != 4 || len(reference.RemainingOutputs) != 0 {
		t.Fatalf("incomplete AOBLMOA Tables 14-23 provenance")
	}
}
