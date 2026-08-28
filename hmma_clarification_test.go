package mayfly

import (
	_ "embed"
	"encoding/json"
	"reflect"
	"testing"
)

const hmmaClarificationTarget = "hmma"

//go:embed docs/reference-data/hmma-2022-clarification-request.json
var hmma2022ClarificationJSON []byte

func TestHMMA2022ClarificationRequestPinsEvidenceGates(t *testing.T) {
	t.Parallel()

	type question struct {
		Answer                    any      `json:"answer"`
		ID                        string   `json:"id"`
		Gate                      string   `json:"gate"`
		Question                  string   `json:"question"`
		KnownEvidence             string   `json:"known_evidence"`
		RequiredAnswerShape       string   `json:"required_answer_shape"`
		Units                     string   `json:"units"`
		Status                    string   `json:"status"`
		AcceptablePrimaryEvidence []string `json:"acceptable_primary_evidence"`
		AffectedFields            []string `json:"affected_fields"`
		Evidence                  []string `json:"evidence"`
	}

	var clarification struct {
		SchemaVersion     int        `json:"schema_version"`
		ProtocolID        string     `json:"protocol_id"`
		Status            string     `json:"status"`
		ReproductionClaim bool       `json:"reproduction_claim"`
		BlockingQuestions []question `json:"blocking_questions"`
		PublicSourceAudit struct {
			Outcome string `json:"outcome"`
			Sources []struct {
				URL     string `json:"url"`
				Finding string `json:"finding"`
			} `json:"sources"`
		} `json:"public_source_audit"`
		ExactPresetGate struct {
			State               string   `json:"state"`
			TargetAlgorithm     string   `json:"target_algorithm"`
			RequiresQuestionIDs []string `json:"requires_question_ids"`
			Rule                string   `json:"rule"`
		} `json:"exact_preset_gate"`
		ReproductionComparisonGate struct {
			State               string   `json:"state"`
			RequiresQuestionIDs []string `json:"requires_question_ids"`
			Rule                string   `json:"rule"`
		} `json:"reproduction_comparison_gate"`
		CorrespondenceDraft struct {
			Status  string `json:"status"`
			Subject string `json:"subject"`
			Body    string `json:"body"`
		} `json:"correspondence_draft"`
	}

	err := json.Unmarshal(hmma2022ClarificationJSON, &clarification)
	if err != nil {
		t.Fatalf("decode HMMA clarification request: %v", err)
	}

	if clarification.SchemaVersion != 1 || clarification.ProtocolID != "hmma-2022-table1" ||
		clarification.Status != "awaiting_author_or_archival_data" || clarification.ReproductionClaim {
		t.Fatalf("unexpected clarification metadata: %+v", clarification)
	}

	wantExactIDs := []string{
		"population_size_and_sex_split",
		"base_ma_parameters_and_crossover",
		"f3_f7_dimensions",
		"benchmark_bounds_and_parameter_tuple",
		"equation_7_information_exchange_coefficient",
		"equation_10_probability_schedule",
		"artificial_mutation_selection_semantics",
	}
	wantComparisonIDs := []string{
		"published_seed_schedule",
		"raw_per_run_results_and_reference_code",
	}
	wantAllIDs := append(append([]string(nil), wantExactIDs...), wantComparisonIDs...)
	gotIDs := make([]string, 0, len(clarification.BlockingQuestions))
	seen := make(map[string]bool, len(clarification.BlockingQuestions))

	for _, blocker := range clarification.BlockingQuestions {
		if seen[blocker.ID] {
			t.Errorf("duplicate blocker ID %q", blocker.ID)
		}

		seen[blocker.ID] = true
		gotIDs = append(gotIDs, blocker.ID)

		if blocker.Status != "unresolved" || blocker.Answer != nil || len(blocker.Evidence) != 0 {
			t.Errorf("blocker %q was resolved without evidence: %+v", blocker.ID, blocker)
		}

		if blocker.Gate == "" || blocker.Question == "" || blocker.KnownEvidence == "" ||
			blocker.RequiredAnswerShape == "" || blocker.Units == "" ||
			len(blocker.AcceptablePrimaryEvidence) == 0 || len(blocker.AffectedFields) == 0 {
			t.Errorf("blocker %q is not actionable: %+v", blocker.ID, blocker)
		}
	}

	if !reflect.DeepEqual(gotIDs, wantAllIDs) ||
		!reflect.DeepEqual(clarification.ExactPresetGate.RequiresQuestionIDs, wantExactIDs) ||
		!reflect.DeepEqual(clarification.ReproductionComparisonGate.RequiresQuestionIDs, wantComparisonIDs) {
		t.Fatalf("clarification blocker IDs drifted: questions=%v exact=%v comparison=%v want=%v",
			gotIDs, clarification.ExactPresetGate.RequiresQuestionIDs,
			clarification.ReproductionComparisonGate.RequiresQuestionIDs, wantAllIDs)
	}

	if clarification.ExactPresetGate.State != "blocked_missing_author_or_archival_data" ||
		clarification.ExactPresetGate.TargetAlgorithm != hmmaClarificationTarget ||
		clarification.ExactPresetGate.Rule == "" ||
		clarification.ReproductionComparisonGate.State != "blocked_missing_historical_run_data" ||
		clarification.ReproductionComparisonGate.Rule == "" || clarification.PublicSourceAudit.Outcome == "" ||
		len(clarification.PublicSourceAudit.Sources) < 4 {
		t.Fatalf("clarification does not preserve its evidence gates: %+v", clarification.ExactPresetGate)
	}

	for _, source := range clarification.PublicSourceAudit.Sources {
		if source.URL == "" || source.Finding == "" {
			t.Errorf("incomplete public-source audit entry: %+v", source)
		}
	}

	if clarification.CorrespondenceDraft.Status != "not_sent" ||
		clarification.CorrespondenceDraft.Subject == "" || clarification.CorrespondenceDraft.Body == "" {
		t.Fatalf("clarification request is not send-ready: %+v", clarification.CorrespondenceDraft)
	}
}

func TestHMMA2022Table1ReferenceLinksClarificationRequest(t *testing.T) {
	t.Parallel()

	var reference struct {
		ProtocolID           string `json:"protocol_id"`
		ClarificationRequest struct {
			Artifact               string   `json:"artifact"`
			Status                 string   `json:"status"`
			TargetAlgorithm        string   `json:"target_algorithm"`
			BlockingQuestionIDs    []string `json:"blocking_question_ids"`
			ExactPresetQuestionIDs []string `json:"exact_preset_question_ids"`
		} `json:"clarification_request"`
	}

	err := json.Unmarshal(hmma2022Table1JSON, &reference)
	if err != nil {
		t.Fatalf("decode HMMA Table 1 reference data: %v", err)
	}

	if reference.ProtocolID != "hmma-2022-table1" ||
		reference.ClarificationRequest.Artifact !=
			"docs/reference-data/hmma-2022-clarification-request.json" ||
		reference.ClarificationRequest.Status != "awaiting_author_or_archival_data" ||
		reference.ClarificationRequest.TargetAlgorithm != hmmaClarificationTarget ||
		len(reference.ClarificationRequest.BlockingQuestionIDs) != 9 ||
		len(reference.ClarificationRequest.ExactPresetQuestionIDs) != 7 {
		t.Fatalf("Table 1 reference does not link the clarification request: %+v", reference)
	}
}
