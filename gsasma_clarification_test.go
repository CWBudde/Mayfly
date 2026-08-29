package mayfly

import (
	_ "embed"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

//go:embed docs/reference-data/gsasma-2022-clarification-request.json
var gsasma2022ClarificationJSON []byte

//go:embed docs/algorithms/gsasma.md
var gsasmaAlgorithmGuide string

func TestGSASMA2022ClarificationRequest(t *testing.T) {
	var clarification struct {
		SchemaVersion     int    `json:"schema_version"`
		ProtocolID        string `json:"protocol_id"`
		Status            string `json:"status"`
		ReproductionClaim bool   `json:"reproduction_claim"`
		SourceContact     struct {
			CorrespondingAuthor string `json:"corresponding_author"`
			PublicEmail         string `json:"public_email"`
			ContactSource       string `json:"contact_source"`
			SelectionBasis      string `json:"selection_basis"`
		} `json:"source_contact"`
		PublicSourceAudit struct {
			CheckedOn string `json:"checked_on"`
			Outcome   string `json:"outcome"`
			Sources   []struct {
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
			Units                     string   `json:"units"`
			AcceptablePrimaryEvidence []string `json:"acceptable_primary_evidence"`
			AffectedFields            []string `json:"affected_fields"`
			Status                    string   `json:"status"`
			Answer                    *string  `json:"answer"`
			Evidence                  []string `json:"evidence"`
		} `json:"blocking_questions"`
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
		CurrentLibraryStatus struct {
			PaperFaithfulAnnealing bool     `json:"paper_faithful_annealing_recurrence_available"`
			PaperFaithfulSMA       bool     `json:"paper_faithful_sma_mating_available"`
			DocumentedExtensions   []string `json:"documented_extensions"`
			AllowedClaim           string   `json:"allowed_claim"`
			ForbiddenClaim         string   `json:"forbidden_claim"`
		} `json:"current_library_status"`
		CorrespondenceDraft struct {
			Status  string `json:"status"`
			Subject string `json:"subject"`
			Body    string `json:"body"`
		} `json:"correspondence_draft"`
	}

	if err := json.Unmarshal(gsasma2022ClarificationJSON, &clarification); err != nil {
		t.Fatalf("decode GSASMA clarification request: %v", err)
	}

	if clarification.SchemaVersion != 1 ||
		clarification.ProtocolID != "gsasma-2022-algorithm-and-benchmarks" ||
		clarification.Status != "awaiting_author_or_archival_data" ||
		clarification.ReproductionClaim {
		t.Fatalf("unexpected clarification metadata: %+v", clarification)
	}

	contact := clarification.SourceContact
	if contact.CorrespondingAuthor != "Mengling Zhao" || contact.PublicEmail == "" ||
		!strings.HasPrefix(contact.ContactSource, "https://") || contact.SelectionBasis == "" {
		t.Fatalf("clarification request has no auditable public contact path: %+v", contact)
	}

	wantExactIDs := []string{
		"initial_temperature",
		"temperature_update_recurrence",
		"tau_sequence_and_lifecycle",
		"sma_probability_bounds",
		"fitness_orientation_and_annealing_delta",
	}
	wantAllIDs := append(append([]string{}, wantExactIDs...),
		"seed_schedule_and_rng_lifecycle",
		"raw_30_run_outputs",
	)

	var (
		gotIDs = make([]string, 0, len(clarification.BlockingQuestions))
		seen   = make(map[string]bool, len(clarification.BlockingQuestions))
	)

	for _, blocker := range clarification.BlockingQuestions {
		if seen[blocker.ID] {
			t.Fatalf("duplicate blocker ID %q", blocker.ID)
		}

		seen[blocker.ID] = true

		gotIDs = append(gotIDs, blocker.ID)

		if blocker.Gate == "" || blocker.Question == "" || blocker.KnownEvidence == "" ||
			blocker.RequiredAnswerShape == "" || blocker.Units == "" ||
			len(blocker.AcceptablePrimaryEvidence) == 0 || len(blocker.AffectedFields) == 0 ||
			blocker.Status != "unresolved" || blocker.Answer != nil || len(blocker.Evidence) != 0 {
			t.Fatalf("blocker %q is not an unresolved actionable gate: %+v", blocker.ID, blocker)
		}
	}

	if !reflect.DeepEqual(gotIDs, wantAllIDs) ||
		!reflect.DeepEqual(clarification.ExactPresetGate.RequiresQuestionIDs, wantExactIDs) ||
		!reflect.DeepEqual(clarification.ReproductionComparisonGate.RequiresQuestionIDs, wantAllIDs) {
		t.Fatalf("clarification blocker IDs drifted: questions=%v exact=%v comparison=%v want=%v",
			gotIDs, clarification.ExactPresetGate.RequiresQuestionIDs,
			clarification.ReproductionComparisonGate.RequiresQuestionIDs, wantAllIDs)
	}

	if clarification.ExactPresetGate.State != "blocked_missing_author_or_archival_data" ||
		clarification.ExactPresetGate.TargetAlgorithm != "gsasma" ||
		clarification.ExactPresetGate.Rule == "" ||
		clarification.ReproductionComparisonGate.State != "blocked_missing_historical_run_data" ||
		clarification.ReproductionComparisonGate.Rule == "" ||
		clarification.PublicSourceAudit.CheckedOn != "2026-08-29" ||
		clarification.PublicSourceAudit.Outcome == "" || len(clarification.PublicSourceAudit.Sources) < 9 {
		t.Fatalf("clarification does not preserve its evidence gates: %+v", clarification.ExactPresetGate)
	}

	sourceKinds := make(map[string]bool, len(clarification.PublicSourceAudit.Sources))
	for _, source := range clarification.PublicSourceAudit.Sources {
		if source.Kind == "" || !strings.HasPrefix(source.URL, "https://") || source.Finding == "" {
			t.Fatalf("incomplete public-source audit entry: %+v", source)
		}

		sourceKinds[source.Kind] = true
	}

	for _, kind := range []string{
		"publisher_article",
		"publisher_flowchart",
		"crossref_record",
		"institutional_contact_profile",
		"github_doi_repository_search",
		"zenodo_doi_record_search",
		"figshare_resource_doi_search",
	} {
		if !sourceKinds[kind] {
			t.Fatalf("GSASMA public-source audit does not pin %q", kind)
		}
	}

	status := clarification.CurrentLibraryStatus
	if status.PaperFaithfulAnnealing || status.PaperFaithfulSMA ||
		len(status.DocumentedExtensions) != 2 || status.AllowedClaim != "current-library GSASMA baseline" ||
		status.ForbiddenClaim == "" {
		t.Fatalf("clarification mislabels current GSASMA behavior: %+v", status)
	}

	if clarification.CorrespondenceDraft.Status != "not_sent" ||
		clarification.CorrespondenceDraft.Subject == "" || clarification.CorrespondenceDraft.Body == "" {
		t.Fatalf("clarification request is not send-ready: %+v", clarification.CorrespondenceDraft)
	}
}

func TestGSASMAAlgorithmGuideLinksClarificationRequest(t *testing.T) {
	const clarificationPath = "../reference-data/gsasma-2022-clarification-request.json"

	if !strings.Contains(gsasmaAlgorithmGuide, clarificationPath) {
		t.Fatalf("GSASMA guide does not link %q", clarificationPath)
	}

	if !strings.Contains(gsasmaAlgorithmGuide, "five exact-algorithm blockers") ||
		!strings.Contains(gsasmaAlgorithmGuide, "two historical-comparison blockers") {
		t.Fatal("GSASMA guide does not preserve the clarification gate scope")
	}
}
