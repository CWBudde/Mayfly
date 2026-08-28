package mayfly

import (
	_ "embed"
	"encoding/json"
	"reflect"
	"testing"
)

//go:embed docs/reference-data/original-ma-2020-clarification-request.json
var originalMA2020ClarificationJSON []byte

func TestOriginalMA2020ClarificationRequestPinsExactPresetBlockers(t *testing.T) {
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
				Kind    string `json:"kind"`
				URL     string `json:"url"`
				Finding string `json:"finding"`
			} `json:"sources"`
		} `json:"public_source_audit"`
		UnrecoveredArchivalLeads []struct {
			ID        string `json:"id"`
			SourceURL string `json:"source_url"`
			Versions  []struct {
				Version     string `json:"version"`
				Published   string `json:"published"`
				ReleaseNote string `json:"release_note"`
			} `json:"versions"`
			BoundaryRelease struct {
				Version     string `json:"version"`
				Published   string `json:"published"`
				ReleaseNote string `json:"release_note"`
			} `json:"boundary_release"`
			AccessStatus       string `json:"access_status"`
			EvidentiaryStatus  string `json:"evidentiary_status"`
			IntakeTool         string `json:"intake_tool"`
			RequiredNextAction string `json:"required_next_action"`
		} `json:"unrecovered_archival_leads"`
		ExactPresetGate struct {
			State               string   `json:"state"`
			TargetAlgorithm     string   `json:"target_algorithm"`
			RequiresQuestionIDs []string `json:"requires_question_ids"`
			Rule                string   `json:"rule"`
		} `json:"exact_preset_gate"`
		CorrespondenceDraft struct {
			Status  string `json:"status"`
			Subject string `json:"subject"`
			Body    string `json:"body"`
		} `json:"correspondence_draft"`
	}

	err := json.Unmarshal(originalMA2020ClarificationJSON, &clarification)
	if err != nil {
		t.Fatalf("decode original MA clarification request: %v", err)
	}

	if clarification.SchemaVersion != 1 || clarification.ProtocolID != "original-ma-2020-table6" ||
		clarification.Status != "awaiting_author_or_archival_data" || clarification.ReproductionClaim {
		t.Fatalf("unexpected clarification metadata: %+v", clarification)
	}

	wantIDs := []string{"crossover_operator_and_rate", "gaussian_mutation_rate_semantics"}
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
		clarification.ExactPresetGate.TargetAlgorithm != "ima" ||
		clarification.ExactPresetGate.Rule == "" || clarification.PublicSourceAudit.Outcome == "" ||
		len(clarification.PublicSourceAudit.Sources) < 4 {
		t.Fatalf("clarification does not preserve its evidence gate: %+v", clarification.ExactPresetGate)
	}

	sourceKinds := make(map[string]bool, len(clarification.PublicSourceAudit.Sources))
	for _, source := range clarification.PublicSourceAudit.Sources {
		if source.Kind == "" || source.URL == "" || source.Finding == "" {
			t.Errorf("incomplete public-source audit entry: %+v", source)
		}

		sourceKinds[source.Kind] = true
	}

	for _, kind := range []string{
		"author_matlab_file_exchange_version_history",
		"author_institutional_repository",
	} {
		if !sourceKinds[kind] {
			t.Errorf("public-source audit does not include %q", kind)
		}
	}

	if len(clarification.UnrecoveredArchivalLeads) != 1 {
		t.Fatalf("unexpected archival leads: %+v", clarification.UnrecoveredArchivalLeads)
	}

	lead := clarification.UnrecoveredArchivalLeads[0]

	gotVersions := make([]string, 0, len(lead.Versions))
	for _, version := range lead.Versions {
		if version.Published == "" {
			t.Errorf("archival version lacks publication date: %+v", version)
		}

		gotVersions = append(gotVersions, version.Version)
	}

	if lead.ID != "matlab_file_exchange_pre_simplification_versions" || lead.SourceURL == "" ||
		!reflect.DeepEqual(gotVersions, []string{"1.0.0", "1.0.1", "1.0.2"}) ||
		lead.BoundaryRelease.Version != "2.0.1" ||
		lead.BoundaryRelease.Published != "2020-06-23" ||
		lead.BoundaryRelease.ReleaseNote != "The code has been simplified." ||
		lead.AccessStatus != "authentication_required_not_retrieved" ||
		lead.EvidentiaryStatus != "uninspected_lead_not_blocker_resolution" ||
		lead.IntakeTool != "cmd/audit-original-ma-archive" ||
		lead.RequiredNextAction == "" {
		t.Fatalf("archival lead drifted or was treated as evidence: %+v", lead)
	}

	if clarification.CorrespondenceDraft.Status != "not_sent" ||
		clarification.CorrespondenceDraft.Subject == "" || clarification.CorrespondenceDraft.Body == "" {
		t.Fatalf("clarification request is not send-ready: %+v", clarification.CorrespondenceDraft)
	}
}
