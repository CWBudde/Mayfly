package mayfly

import (
	_ "embed"
	"encoding/json"
	"reflect"
	"testing"
)

//go:embed docs/reference-data/desma-2022-clarification-request.json
var desma2022ClarificationJSON []byte

func TestDESMA2022ClarificationRequestPinsExactPresetBlockers(t *testing.T) {
	t.Parallel()

	type question struct {
		Answer                    any      `json:"answer"`
		ID                        string   `json:"id"`
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
			RequiresQuestionIDs []string `json:"requires_question_ids"`
			Rule                string   `json:"rule"`
		} `json:"exact_preset_gate"`
		CorrespondenceDraft struct {
			Status  string `json:"status"`
			Subject string `json:"subject"`
			Body    string `json:"body"`
		} `json:"correspondence_draft"`
	}

	if err := json.Unmarshal(desma2022ClarificationJSON, &clarification); err != nil {
		t.Fatalf("decode DESMA clarification request: %v", err)
	}

	if clarification.SchemaVersion != 1 || clarification.ProtocolID != "desma-2022-table3" ||
		clarification.Status != "awaiting_author_or_archival_data" || clarification.ReproductionClaim {
		t.Fatalf("unexpected clarification metadata: %+v", clarification)
	}

	wantIDs := []string{
		"initial_search_radius",
		"population_split",
		"base_ma_settings",
		"evaluation_accounting",
		"published_seed_schedule",
		"raw_per_run_results",
	}
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

		if blocker.Question == "" || blocker.KnownEvidence == "" || blocker.RequiredAnswerShape == "" ||
			blocker.Units == "" || len(blocker.AcceptablePrimaryEvidence) == 0 || len(blocker.AffectedFields) == 0 {
			t.Errorf("blocker %q is not actionable: %+v", blocker.ID, blocker)
		}
	}

	if !reflect.DeepEqual(gotIDs, wantIDs) ||
		!reflect.DeepEqual(clarification.ExactPresetGate.RequiresQuestionIDs, wantIDs) {
		t.Fatalf("clarification blocker IDs drifted: questions=%v gate=%v want=%v",
			gotIDs, clarification.ExactPresetGate.RequiresQuestionIDs, wantIDs)
	}

	if clarification.ExactPresetGate.State != "blocked_missing_author_or_archival_data" ||
		clarification.ExactPresetGate.Rule == "" || clarification.PublicSourceAudit.Outcome == "" ||
		len(clarification.PublicSourceAudit.Sources) < 4 {
		t.Fatalf("clarification does not preserve its evidence gate: %+v", clarification.ExactPresetGate)
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

func TestDESMA2022Table3ReferenceLinksClarificationRequest(t *testing.T) {
	t.Parallel()

	var reference struct {
		ProtocolID           string `json:"protocol_id"`
		ClarificationRequest struct {
			Artifact            string   `json:"artifact"`
			Status              string   `json:"status"`
			BlockingQuestionIDs []string `json:"blocking_question_ids"`
		} `json:"clarification_request"`
	}

	if err := json.Unmarshal(desma2022Table3JSON, &reference); err != nil {
		t.Fatalf("decode DESMA Table 3 reference data: %v", err)
	}

	if reference.ProtocolID != "desma-2022-table3" ||
		reference.ClarificationRequest.Artifact !=
			"docs/reference-data/desma-2022-clarification-request.json" ||
		reference.ClarificationRequest.Status != "awaiting_author_or_archival_data" ||
		len(reference.ClarificationRequest.BlockingQuestionIDs) != 6 {
		t.Fatalf("Table 3 reference does not link the clarification request: %+v", reference)
	}
}
