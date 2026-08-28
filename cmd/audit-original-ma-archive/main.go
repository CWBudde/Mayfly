// Command audit-original-ma-archive creates a provenance and triage manifest
// for the three pre-simplification MATLAB File Exchange releases of the
// original Mayfly Algorithm implementation.
package main

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	protocolID              = "original-ma-2020-table6"
	sourceURL               = "https://www.mathworks.com/matlabcentral/fileexchange/76902-mayfly-optimization-algorithm"
	maxUncompressedFileSize = uint64(64 << 20)
	maxUncompressedTotal    = uint64(256 << 20)
)

var expectedReleases = []release{
	{Version: "1.0.0", Published: "2020-06-13"},
	{Version: "1.0.1", Published: "2020-06-13"},
	{Version: "1.0.2", Published: "2020-06-19"},
}

type release struct {
	Version   string `json:"version"`
	Published string `json:"published"`
}

type manifest struct {
	ProtocolID           string         `json:"protocol_id"`
	SourceURL            string         `json:"source_url"`
	RetrievedOn          string         `json:"retrieved_on"`
	AcquisitionNote      string         `json:"acquisition_note"`
	InterpretationStatus string         `json:"interpretation_status"`
	Packages             []packageAudit `json:"packages"`
	ManualReviewRequired []string       `json:"manual_review_required"`
	SchemaVersion        int            `json:"schema_version"`
}

type packageAudit struct {
	Version       string           `json:"version"`
	Published     string           `json:"published"`
	ArchiveName   string           `json:"archive_name"`
	ArchiveSHA256 string           `json:"archive_sha256"`
	ContentSHA256 string           `json:"content_sha256"`
	Files         []fileAudit      `json:"files"`
	Triage        triageIndicators `json:"triage_indicators"`
	ArchiveBytes  int64            `json:"archive_bytes"`
}

type fileAudit struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Bytes  uint64 `json:"bytes"`
}

type triageIndicators struct {
	MATLABFiles                   []string `json:"matlab_files"`
	BenchmarkNames                []string `json:"benchmark_names"`
	MentionsIMA                   bool     `json:"mentions_ima"`
	MentionsEvaluationBudget95000 bool     `json:"mentions_evaluation_budget_95000"`
	MentionsCrossoverRate095      bool     `json:"mentions_crossover_rate_0_95"`
	MentionsMutationRate01        bool     `json:"mentions_mutation_rate_0_1"`
}

func main() {
	var (
		output          string
		retrievedOn     string
		acquisitionNote string
	)

	flag.StringVar(&output, "output", "", "manifest output path (required)")
	flag.StringVar(&retrievedOn, "retrieved-on", "", "archive retrieval date in YYYY-MM-DD form (required)")
	flag.StringVar(&acquisitionNote, "acquisition-note", "", "how the official archives were obtained (required)")
	flag.Parse()

	if output == "" || acquisitionNote == "" {
		flag.Usage()
		os.Exit(2)
	}

	_, err := time.Parse("2006-01-02", retrievedOn)
	if err != nil {
		fatalf("invalid -retrieved-on date %q: %v", retrievedOn, err)
	}

	archives, err := parseArchives(flag.Args())
	if err != nil {
		fatalf("invalid archive arguments: %v", err)
	}

	report, err := auditArchives(archives, retrievedOn, acquisitionNote)
	if err != nil {
		fatalf("audit archives: %v", err)
	}

	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fatalf("encode manifest: %v", err)
	}

	encoded = append(encoded, '\n')

	err = os.WriteFile(output, encoded, 0o600)
	if err != nil {
		fatalf("write manifest: %v", err)
	}
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "audit-original-ma-archive: "+format+"\n", args...)
	os.Exit(1)
}

func parseArchives(args []string) (map[string]string, error) {
	if len(args) != len(expectedReleases) {
		return nil, fmt.Errorf("expected exactly %d version=archive.zip arguments", len(expectedReleases))
	}

	want := make(map[string]bool, len(expectedReleases))
	for _, item := range expectedReleases {
		want[item.Version] = true
	}

	archives := make(map[string]string, len(args))
	for _, arg := range args {
		version, archivePath, ok := strings.Cut(arg, "=")
		if !ok || version == "" || archivePath == "" {
			return nil, fmt.Errorf("argument %q must use version=archive.zip form", arg)
		}

		if !want[version] {
			return nil, fmt.Errorf("unexpected version %q", version)
		}

		if _, duplicate := archives[version]; duplicate {
			return nil, fmt.Errorf("duplicate version %q", version)
		}

		archives[version] = archivePath
	}

	return archives, nil
}

func auditArchives(archives map[string]string, retrievedOn, acquisitionNote string) (manifest, error) {
	report := manifest{
		SchemaVersion:        1,
		ProtocolID:           protocolID,
		SourceURL:            sourceURL,
		RetrievedOn:          retrievedOn,
		AcquisitionNote:      acquisitionNote,
		Packages:             make([]packageAudit, 0, len(expectedReleases)),
		InterpretationStatus: "manual_source_review_required_not_blocker_resolution",
		ManualReviewRequired: []string{
			"Determine whether any package contains the six-benchmark Table 6 experiment driver and exact " +
				"50-run, 95000-evaluation, 20-male/20-female protocol.",
			"Determine the executable crossover operator and the exact event controlled by rate 0.95.",
			"Determine the executable Gaussian mutation operator and the exact quantity controlled by rate 0.1.",
		},
	}

	for _, item := range expectedReleases {
		archivePath, ok := archives[item.Version]
		if !ok {
			return manifest{}, fmt.Errorf("missing archive for version %s", item.Version)
		}

		result, err := auditArchive(archivePath, item)
		if err != nil {
			return manifest{}, fmt.Errorf("version %s: %w", item.Version, err)
		}

		report.Packages = append(report.Packages, result)
	}

	return report, nil
}

func auditArchive(archivePath string, item release) (packageAudit, error) {
	archive, err := os.Open(archivePath)
	if err != nil {
		return packageAudit{}, err
	}
	defer archive.Close()

	info, err := archive.Stat()
	if err != nil {
		return packageAudit{}, err
	}

	if !info.Mode().IsRegular() {
		return packageAudit{}, errors.New("archive is not a regular file")
	}

	archiveHash := sha256.New()

	_, err = io.Copy(archiveHash, archive)
	if err != nil {
		return packageAudit{}, fmt.Errorf("hash archive: %w", err)
	}

	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return packageAudit{}, fmt.Errorf("open ZIP: %w", err)
	}
	defer reader.Close()

	files := make([]fileAudit, 0, len(reader.File))
	contents := make(map[string][]byte, len(reader.File))

	var totalUncompressed uint64

	seen := make(map[string]bool, len(reader.File))
	for _, zipped := range reader.File {
		cleaned, err := safeArchivePath(zipped.Name)
		if err != nil {
			return packageAudit{}, err
		}

		if zipped.FileInfo().IsDir() {
			continue
		}

		if zipped.Flags&0x1 != 0 {
			return packageAudit{}, fmt.Errorf("encrypted entry %q is unsupported", zipped.Name)
		}

		if seen[cleaned] {
			return packageAudit{}, fmt.Errorf("duplicate entry %q", cleaned)
		}

		seen[cleaned] = true

		if zipped.UncompressedSize64 > maxUncompressedFileSize ||
			totalUncompressed > maxUncompressedTotal-zipped.UncompressedSize64 {
			return packageAudit{}, fmt.Errorf("uncompressed size limit exceeded by %q", cleaned)
		}

		totalUncompressed += zipped.UncompressedSize64

		data, err := readZipFile(zipped)
		if err != nil {
			return packageAudit{}, fmt.Errorf("read %q: %w", cleaned, err)
		}

		digest := sha256.Sum256(data)
		files = append(files, fileAudit{
			Path:   cleaned,
			Bytes:  uint64(len(data)),
			SHA256: hex.EncodeToString(digest[:]),
		})
		contents[cleaned] = data
	}

	if len(files) == 0 {
		return packageAudit{}, errors.New("ZIP contains no files")
	}

	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })

	contentHash := sha256.New()
	for _, file := range files {
		fmt.Fprintf(contentHash, "%s\x00%d\x00", file.Path, file.Bytes)
		contentHash.Write(contents[file.Path])
	}

	return packageAudit{
		Version:       item.Version,
		Published:     item.Published,
		ArchiveName:   filepath.Base(archivePath),
		ArchiveBytes:  info.Size(),
		ArchiveSHA256: hex.EncodeToString(archiveHash.Sum(nil)),
		ContentSHA256: hex.EncodeToString(contentHash.Sum(nil)),
		Files:         files,
		Triage:        triage(contents),
	}, nil
}

func safeArchivePath(name string) (string, error) {
	normalized := strings.ReplaceAll(name, "\\", "/")
	cleaned := path.Clean(normalized)

	if cleaned == "." || strings.HasPrefix(cleaned, "/") || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", fmt.Errorf("unsafe archive entry %q", name)
	}

	return cleaned, nil
}

func readZipFile(file *zip.File) ([]byte, error) {
	reader, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	limited := io.LimitReader(reader, int64(maxUncompressedFileSize)+1)

	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}

	if uint64(len(data)) > maxUncompressedFileSize {
		return nil, errors.New("uncompressed file exceeds size limit")
	}

	return data, nil
}

func triage(contents map[string][]byte) triageIndicators {
	result := triageIndicators{
		MATLABFiles:                   []string{},
		BenchmarkNames:                []string{},
		MentionsIMA:                   false,
		MentionsEvaluationBudget95000: false,
		MentionsCrossoverRate095:      false,
		MentionsMutationRate01:        false,
	}
	allSource := strings.Builder{}

	for name, data := range contents {
		if !strings.EqualFold(filepath.Ext(name), ".m") {
			continue
		}

		result.MATLABFiles = append(result.MATLABFiles, name)

		allSource.Write(data)
		allSource.WriteByte('\n')
	}

	sort.Strings(result.MATLABFiles)

	source := strings.ToLower(allSource.String())
	for _, benchmark := range []string{"ackley", "beale", "eggcrate", "rastrigin", "rosenbrock", "sphere"} {
		if strings.Contains(source, benchmark) {
			result.BenchmarkNames = append(result.BenchmarkNames, benchmark)
		}
	}

	result.MentionsIMA = containsIdentifier(source, "ima") || strings.Contains(source, "improved mayfly")
	result.MentionsEvaluationBudget95000 = strings.Contains(source, "95000") || strings.Contains(source, "95,000")
	result.MentionsCrossoverRate095 = strings.Contains(source, "0.95") || strings.Contains(source, ".95")
	result.MentionsMutationRate01 = strings.Contains(source, "0.1") || strings.Contains(source, ".1")

	return result
}

func containsIdentifier(source, identifier string) bool {
	searchFrom := 0
	for searchFrom < len(source) {
		relative := strings.Index(source[searchFrom:], identifier)
		if relative < 0 {
			return false
		}

		start := searchFrom + relative
		end := start + len(identifier)
		beforeBoundary := start == 0 || !isIdentifierByte(source[start-1])

		afterBoundary := end == len(source) || !isIdentifierByte(source[end])
		if beforeBoundary && afterBoundary {
			return true
		}

		searchFrom = start + 1
	}

	return false
}

func isIdentifierByte(value byte) bool {
	return value >= 'a' && value <= 'z' || value >= '0' && value <= '9' || value == '_'
}
