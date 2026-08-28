package main

import (
	"archive/zip"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestAuditArchivesProducesDeterministicProvenanceAndTriage(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	archives := make(map[string]string, len(expectedReleases))

	for index, item := range expectedReleases {
		archivePath := filepath.Join(dir, item.Version+".zip")

		files := map[string]string{
			"MA/main.m":      "benchmarks = {'Sphere','Rosenbrock','Rastrigin','Ackley','Eggcrate','Beale'}; maxeval = 95000; crossover = 0.95; mutation = 0.1; % IMA\n",
			"MA/version.txt": item.Version,
		}
		if index == 1 {
			files["MA/notes.txt"] = "second release"
		}

		writeTestZIP(t, archivePath, files)
		archives[item.Version] = archivePath
	}

	report, err := auditArchives(archives, "2026-08-28", "Downloaded while signed in to MathWorks")
	if err != nil {
		t.Fatalf("audit archives: %v", err)
	}

	if report.SchemaVersion != 1 || report.ProtocolID != protocolID || report.SourceURL != sourceURL ||
		report.InterpretationStatus != "manual_source_review_required_not_blocker_resolution" ||
		len(report.Packages) != 3 || len(report.ManualReviewRequired) != 3 {
		t.Fatalf("unexpected manifest metadata: %+v", report)
	}

	wantBenchmarks := []string{"ackley", "beale", "eggcrate", "rastrigin", "rosenbrock", "sphere"}

	for index, item := range report.Packages {
		if item.Version != expectedReleases[index].Version || item.Published != expectedReleases[index].Published ||
			len(item.ArchiveSHA256) != 64 || len(item.ContentSHA256) != 64 || item.ArchiveBytes == 0 {
			t.Errorf("incomplete package provenance: %+v", item)
		}

		if !reflect.DeepEqual(item.Triage.BenchmarkNames, wantBenchmarks) || !item.Triage.MentionsIMA ||
			!item.Triage.MentionsEvaluationBudget95000 || !item.Triage.MentionsCrossoverRate095 ||
			!item.Triage.MentionsMutationRate01 {
			t.Errorf("unexpected triage indicators: %+v", item.Triage)
		}

		for fileIndex := 1; fileIndex < len(item.Files); fileIndex++ {
			if item.Files[fileIndex-1].Path >= item.Files[fileIndex].Path {
				t.Errorf("file inventory is not sorted: %+v", item.Files)
			}
		}
	}
}

func TestParseArchivesRequiresEachExpectedVersionOnce(t *testing.T) {
	t.Parallel()

	valid := []string{"1.0.2=c.zip", "1.0.0=a.zip", "1.0.1=b.zip"}

	got, err := parseArchives(valid)
	if err != nil {
		t.Fatalf("parse valid archives: %v", err)
	}

	if got["1.0.0"] != "a.zip" || got["1.0.1"] != "b.zip" || got["1.0.2"] != "c.zip" {
		t.Fatalf("unexpected parsed archives: %v", got)
	}

	for _, args := range [][]string{
		{"1.0.0=a.zip", "1.0.1=b.zip"},
		{"1.0.0=a.zip", "1.0.1=b.zip", "2.0.1=c.zip"},
		{"1.0.0=a.zip", "1.0.0=b.zip", "1.0.2=c.zip"},
		{"1.0.0=a.zip", "1.0.1=b.zip", "not-an-assignment"},
	} {
		_, err := parseArchives(args)
		if err == nil {
			t.Errorf("parseArchives(%v) unexpectedly succeeded", args)
		}
	}
}

func TestAuditArchiveRejectsUnsafePaths(t *testing.T) {
	t.Parallel()

	archivePath := filepath.Join(t.TempDir(), "unsafe.zip")
	writeTestZIP(t, archivePath, map[string]string{"../escape.m": "disp('unsafe')"})

	_, err := auditArchive(archivePath, expectedReleases[0])
	if err == nil || !strings.Contains(err.Error(), "unsafe archive entry") {
		t.Fatalf("unsafe archive error = %v", err)
	}
}

func TestTriageDoesNotTreatIdentifierSubstringsAsIMA(t *testing.T) {
	t.Parallel()

	result := triage(map[string][]byte{
		"plot.m": []byte("imagesc(values); title('primary result');\n"),
	})

	if result.MentionsIMA {
		t.Fatal("image/primary substrings were treated as an IMA identifier")
	}
}

func writeTestZIP(t *testing.T, archivePath string, files map[string]string) {
	t.Helper()

	archive, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("create test ZIP: %v", err)
	}

	writer := zip.NewWriter(archive)

	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}

	// Deliberately write the map order as provided by iteration. The auditor's
	// output must sort entries independently of ZIP member order.
	for _, name := range names {
		entry, createErr := writer.Create(name)
		if createErr != nil {
			t.Fatalf("create ZIP entry: %v", createErr)
		}

		_, err = entry.Write([]byte(files[name]))
		if err != nil {
			t.Fatalf("write ZIP entry: %v", err)
		}
	}

	err = writer.Close()
	if err != nil {
		t.Fatalf("close ZIP writer: %v", err)
	}

	err = archive.Close()
	if err != nil {
		t.Fatalf("close ZIP: %v", err)
	}
}
