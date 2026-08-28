package mayfly

import (
	_ "embed"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

//go:embed docs/reference-data/olce-ma-2022-clarification-request.json
var olceMA2022ClarificationJSON []byte

//go:embed docs/algorithms/olce-ma.md
var olceMAAlgorithmGuide string

func TestOLCEMA2022ClarificationRequest(t *testing.T) {
	var clarification struct {
		SchemaVersion     int    `json:"schema_version"`
		ProtocolID        string `json:"protocol_id"`
		Status            string `json:"status"`
		ReproductionClaim bool   `json:"reproduction_claim"`
		SourceContact     struct {
			ContactAuthor            string `json:"contact_author"`
			PaperCorrespondenceEmail string `json:"paper_correspondence_email"`
			CurrentPublicEmail       string `json:"current_public_email"`
			PaperContactSource       string `json:"paper_contact_source"`
			CurrentContactSource     string `json:"current_contact_source"`
		} `json:"source_contact"`
		PublicSourceAudit struct {
			Outcome string `json:"outcome"`
			Sources []struct {
				Kind    string `json:"kind"`
				URL     string `json:"url"`
				Finding string `json:"finding"`
			} `json:"sources"`
		} `json:"public_source_audit"`
		BlockingQuestions []struct {
			ID                        string   `json:"id"`
			Gate                      string   `json:"gate"`
			Question                  string   `json:"question"`
			KnownEvidence             string   `json:"known_evidence"`
			RequiredAnswerShape       string   `json:"required_answer_shape"`
			AcceptablePrimaryEvidence []string `json:"acceptable_primary_evidence"`
			Status                    string   `json:"status"`
			Answer                    *string  `json:"answer"`
			Evidence                  []string `json:"evidence"`
		} `json:"blocking_questions"`
		ExactGate struct {
			State               string   `json:"state"`
			TargetAlgorithm     string   `json:"target_algorithm"`
			RequiresQuestionIDs []string `json:"requires_question_ids"`
			Rule                string   `json:"rule"`
		} `json:"exact_chaotic_exploitation_gate"`
		CurrentLibraryStatus struct {
			PaperFaithfulAvailable bool   `json:"paper_faithful_chaotic_exploitation_available"`
			DocumentedExtension    string `json:"documented_extension"`
			AllowedClaim           string `json:"allowed_claim"`
			ForbiddenClaim         string `json:"forbidden_claim"`
		} `json:"current_library_status"`
		CorrespondenceDraft struct {
			Status  string `json:"status"`
			Subject string `json:"subject"`
			Body    string `json:"body"`
		} `json:"correspondence_draft"`
	}

	err := json.Unmarshal(olceMA2022ClarificationJSON, &clarification)
	if err != nil {
		t.Fatalf("decode OLCE-MA clarification request: %v", err)
	}

	if clarification.SchemaVersion != 1 ||
		clarification.ProtocolID != "olce-ma-2022-chaotic-exploitation" ||
		clarification.Status != "awaiting_author_or_archival_data" ||
		clarification.ReproductionClaim {
		t.Fatalf("unexpected clarification metadata: %+v", clarification)
	}

	contact := clarification.SourceContact
	if contact.ContactAuthor != "Xiaoping Su" || contact.PaperCorrespondenceEmail == "" ||
		contact.CurrentPublicEmail == "" || !strings.HasPrefix(contact.PaperContactSource, "https://") ||
		!strings.HasPrefix(contact.CurrentContactSource, "https://") {
		t.Fatalf("clarification request has no auditable public contact path: %+v", contact)
	}

	wantIDs := []string{
		"chebyshev_recurrence_and_seed",
		"chaotic_sequence_lifecycle",
		"offspring_component_mutation",
	}
	gotIDs := make([]string, 0, len(clarification.BlockingQuestions))
	seen := make(map[string]bool, len(clarification.BlockingQuestions))

	for _, blocker := range clarification.BlockingQuestions {
		gotIDs = append(gotIDs, blocker.ID)
		if blocker.ID == "" || seen[blocker.ID] {
			t.Fatalf("empty or duplicate clarification blocker ID %q", blocker.ID)
		}

		seen[blocker.ID] = true

		if blocker.Gate != "exact_chaotic_exploitation" || blocker.Question == "" ||
			blocker.KnownEvidence == "" || blocker.RequiredAnswerShape == "" ||
			len(blocker.AcceptablePrimaryEvidence) == 0 || blocker.Status != "unresolved" ||
			blocker.Answer != nil || len(blocker.Evidence) != 0 {
			t.Fatalf("blocker %q is not an unresolved evidence gate: %+v", blocker.ID, blocker)
		}
	}

	if !reflect.DeepEqual(gotIDs, wantIDs) ||
		!reflect.DeepEqual(clarification.ExactGate.RequiresQuestionIDs, wantIDs) {
		t.Fatalf("clarification blocker IDs drifted: questions=%v gate=%v want=%v",
			gotIDs, clarification.ExactGate.RequiresQuestionIDs, wantIDs)
	}

	if clarification.ExactGate.State != "blocked_missing_author_or_archival_data" ||
		clarification.ExactGate.TargetAlgorithm != "olce-ma" || clarification.ExactGate.Rule == "" ||
		clarification.PublicSourceAudit.Outcome == "" || len(clarification.PublicSourceAudit.Sources) < 8 {
		t.Fatalf("clarification does not preserve its evidence gate: %+v", clarification.ExactGate)
	}

	for _, source := range clarification.PublicSourceAudit.Sources {
		if source.Kind == "" || !strings.HasPrefix(source.URL, "https://") || source.Finding == "" {
			t.Fatalf("incomplete public-source audit entry: %+v", source)
		}
	}

	status := clarification.CurrentLibraryStatus
	if status.PaperFaithfulAvailable || status.DocumentedExtension == "" ||
		status.AllowedClaim != "current-library baseline" || status.ForbiddenClaim == "" {
		t.Fatalf("clarification mislabels current OLCE-MA behavior: %+v", status)
	}

	if clarification.CorrespondenceDraft.Status != "not_sent" ||
		clarification.CorrespondenceDraft.Subject == "" || clarification.CorrespondenceDraft.Body == "" {
		t.Fatalf("clarification request is not send-ready: %+v", clarification.CorrespondenceDraft)
	}
}

func TestOLCEMAAlgorithmGuideLinksClarificationRequest(t *testing.T) {
	const clarificationPath = "../reference-data/olce-ma-2022-clarification-request.json"

	if !strings.Contains(olceMAAlgorithmGuide, clarificationPath) {
		t.Fatalf("OLCE-MA guide does not link %q", clarificationPath)
	}
}
