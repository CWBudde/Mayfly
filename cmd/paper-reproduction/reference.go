package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"sort"

	"github.com/cwbudde/mayfly"
)

const (
	originalMAReferenceSchema = 1
	publishedComparisonSchema = 1
)

var originalMAStatisticOrder = []string{"best", "worst", "mean", "median", "standard_deviation"}

type originalMAReference struct {
	PublishedResults map[string]map[string][]float64 `json:"published_results"`
	ProtocolID       string                          `json:"protocol_id"`
	StatisticsOrder  []string                        `json:"statistics_order"`
	Benchmarks       []originalMABenchmark           `json:"benchmarks"`
	SHA256           string                          `json:"-"`
	Source           originalMAReferenceSource       `json:"source"`
	IMA              originalMAParameters            `json:"ima"`
	Execution        struct {
		Replications        int `json:"replications"`
		FunctionEvaluations int `json:"function_evaluations_per_replication"`
		MalePopulation      int `json:"male_population"`
		FemalePopulation    int `json:"female_population"`
	} `json:"execution"`
	SchemaVersion int `json:"schema_version"`
}

type originalMAReferenceSource struct { //nolint:govet // Keep source-provenance fields in JSON order.
	Authors      []string `json:"authors"`
	Title        string   `json:"title"`
	DOI          string   `json:"doi"`
	PublisherURL string   `json:"publisher_url"`
	Table        int      `json:"table"`
}

type originalMAParameters struct {
	UnresolvedOperatorSemantics []string                   `json:"unresolved_operator_semantics"`
	CommonParameters            originalMACommonParameters `json:"common_parameters"`
}

type originalMACommonParameters struct {
	Selection    string  `json:"selection"`
	A1           float64 `json:"a1"`
	A2           float64 `json:"a2"`
	Beta         float64 `json:"beta"`
	Dance        float64 `json:"nuptial_dance"`
	RandomFlight float64 `json:"random_flight"`
}

type originalMABenchmark struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Dimension    int     `json:"dimension"`
	LowerBound   float64 `json:"lower_bound"`
	UpperBound   float64 `json:"upper_bound"`
	KnownMinimum float64 `json:"known_minimum"`
}

type computedStatistics struct {
	Best                  float64 `json:"best"`
	Worst                 float64 `json:"worst"`
	Mean                  float64 `json:"mean"`
	Median                float64 `json:"median"`
	PopulationStandardDev float64 `json:"population_standard_deviation"`
}

type publishedStatistics struct {
	Best                      float64 `json:"best"`
	Worst                     float64 `json:"worst"`
	Mean                      float64 `json:"mean"`
	Median                    float64 `json:"median"`
	ReportedStandardDeviation float64 `json:"reported_standard_deviation"`
}

type statisticDifferences struct {
	Best   float64 `json:"best"`
	Worst  float64 `json:"worst"`
	Mean   float64 `json:"mean"`
	Median float64 `json:"median"`
}

type protocolAlignment struct {
	Reasons             []string `json:"reasons"`
	BenchmarkGeometry   bool     `json:"benchmark_geometry"`
	Replications        bool     `json:"replications"`
	EvaluationBudget    bool     `json:"evaluation_budget"`
	Population          bool     `json:"population"`
	KnownParameters     bool     `json:"known_parameters"`
	OperatorSemantics   bool     `json:"operator_semantics"`
	PublishedSeedsKnown bool     `json:"published_seeds_known"`
}

type publishedBenchmarkComparison struct { //nolint:govet // Keep the JSON fields grouped by meaning.
	Computed                  computedStatistics   `json:"computed"`
	Published                 publishedStatistics  `json:"published"`
	ComputedMinusPublished    statisticDifferences `json:"computed_minus_published"`
	Alignment                 protocolAlignment    `json:"protocol_alignment"`
	BenchmarkID               string               `json:"benchmark_id"`
	BenchmarkName             string               `json:"benchmark_name"`
	ComputedAlgorithm         string               `json:"computed_algorithm"`
	PublishedAlgorithm        string               `json:"published_algorithm"`
	PublishedStdDevConvention string               `json:"published_standard_deviation_convention"`
	Dimension                 int                  `json:"dimension"`
	RequestedRuns             int                  `json:"requested_runs"`
	AvailableRuns             int                  `json:"available_runs"`
	FailedRuns                int                  `json:"failed_runs"`
	ExpectedEvaluationsPerRun int                  `json:"expected_function_evaluations_per_run"`
}

type publishedComparisonSummary struct { //nolint:govet // Keep provenance and status fields together.
	Comparisons                       []publishedBenchmarkComparison `json:"comparisons"`
	UnresolvedOperatorSemantics       []string                       `json:"unresolved_operator_semantics"`
	UnavailablePublishedAlgorithms    []string                       `json:"unavailable_published_algorithms"`
	ComparisonKind                    string                         `json:"comparison_kind"`
	ProtocolID                        string                         `json:"protocol_id"`
	ReferenceDOI                      string                         `json:"reference_doi"`
	ReferenceSHA256                   string                         `json:"reference_sha256"`
	PublishedStandardDeviationMeaning string                         `json:"published_standard_deviation_convention"`
	SchemaVersion                     int                            `json:"schema_version"`
	ReproductionClaim                 bool                           `json:"reproduction_claim"`
}

type basicMAComparisonInput struct {
	Result      *mayfly.ComparisonResult
	Config      *mayfly.Config
	BenchmarkID string
	Dimension   int
	LowerBound  float64
	UpperBound  float64
}

func loadOriginalMAReference(path string) (originalMAReference, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return originalMAReference{}, fmt.Errorf("read published reference: %w", err)
	}

	var reference originalMAReference

	err = json.Unmarshal(data, &reference)
	if err != nil {
		return originalMAReference{}, fmt.Errorf("decode published reference: %w", err)
	}

	digest := sha256.Sum256(data)
	reference.SHA256 = hex.EncodeToString(digest[:])

	err = validateOriginalMAReference(reference)
	if err != nil {
		return originalMAReference{}, fmt.Errorf("validate published reference: %w", err)
	}

	return reference, nil
}

func validateOriginalMAReference(reference originalMAReference) error {
	err := validateOriginalMAReferenceMetadata(reference)
	if err != nil {
		return err
	}

	seenIDs, err := validateOriginalMABenchmarks(reference)
	if err != nil {
		return err
	}

	for benchmarkID := range reference.PublishedResults {
		if !seenIDs[benchmarkID] {
			return fmt.Errorf("published results contain unknown benchmark %q", benchmarkID)
		}
	}

	if len(reference.IMA.UnresolvedOperatorSemantics) == 0 {
		return errors.New("unresolved operator semantics must remain explicit")
	}

	return nil
}

func validateOriginalMAReferenceMetadata(reference originalMAReference) error {
	if reference.SchemaVersion != originalMAReferenceSchema {
		return fmt.Errorf("schema_version = %d, want %d", reference.SchemaVersion, originalMAReferenceSchema)
	}

	if reference.ProtocolID != "original-ma-2020-table6" {
		return fmt.Errorf("unexpected protocol_id %q", reference.ProtocolID)
	}

	if reference.Source.DOI == "" || reference.Source.PublisherURL == "" || reference.Source.Table != 6 {
		return errors.New("source DOI, publisher URL, and Table 6 provenance are required")
	}

	if reference.Execution.Replications <= 0 || reference.Execution.FunctionEvaluations <= 0 ||
		reference.Execution.MalePopulation <= 0 || reference.Execution.FemalePopulation <= 0 {
		return errors.New("execution counts must be positive")
	}

	err := validateOriginalMACommonParameters(reference.IMA.CommonParameters)
	if err != nil {
		return err
	}

	if !equalStrings(reference.StatisticsOrder, originalMAStatisticOrder) {
		return fmt.Errorf("statistics_order = %v, want %v", reference.StatisticsOrder, originalMAStatisticOrder)
	}

	if len(reference.Benchmarks) == 0 {
		return errors.New("at least one benchmark is required")
	}

	return nil
}

func validateOriginalMACommonParameters(parameters originalMACommonParameters) error {
	if !finite(parameters.A1) || !finite(parameters.A2) || !finite(parameters.Beta) ||
		!finite(parameters.Dance) || !finite(parameters.RandomFlight) ||
		parameters.A1 <= 0 || parameters.A2 <= 0 || parameters.Beta <= 0 ||
		parameters.Dance <= 0 || parameters.RandomFlight <= 0 || parameters.Selection == "" {
		return errors.New("common MA parameters must be finite and positive and selection must be named")
	}

	return nil
}

func validateOriginalMABenchmarks(reference originalMAReference) (map[string]bool, error) {
	seenIDs := make(map[string]bool, len(reference.Benchmarks))
	seenNames := make(map[string]bool, len(reference.Benchmarks))

	for _, benchmark := range reference.Benchmarks {
		if benchmark.ID == "" || benchmark.Name == "" {
			return nil, errors.New("benchmark ID and name are required")
		}

		if seenIDs[benchmark.ID] {
			return nil, fmt.Errorf("duplicate benchmark ID %q", benchmark.ID)
		}

		if seenNames[benchmark.Name] {
			return nil, fmt.Errorf("duplicate benchmark name %q", benchmark.Name)
		}

		seenIDs[benchmark.ID] = true
		seenNames[benchmark.Name] = true

		if benchmark.Dimension <= 0 || !finite(benchmark.LowerBound) || !finite(benchmark.UpperBound) ||
			!finite(benchmark.KnownMinimum) || benchmark.LowerBound >= benchmark.UpperBound {
			return nil, fmt.Errorf("benchmark %s has invalid dimension, bounds, or minimum", benchmark.ID)
		}

		rows, ok := reference.PublishedResults[benchmark.ID]
		if !ok {
			return nil, fmt.Errorf("benchmark %s has no published results", benchmark.ID)
		}

		for _, algorithm := range []string{"basic_ma", "vgma", "sma", "ima"} {
			values, rowOK := rows[algorithm]
			if !rowOK {
				return nil, fmt.Errorf("benchmark %s has no %s result", benchmark.ID, algorithm)
			}

			err := validatePublishedStatistics(benchmark.ID, algorithm, values)
			if err != nil {
				return nil, err
			}
		}
	}

	return seenIDs, nil
}

func validatePublishedStatistics(benchmarkID, algorithm string, values []float64) error {
	if len(values) != len(originalMAStatisticOrder) {
		return fmt.Errorf("benchmark %s %s has %d statistics, want %d",
			benchmarkID, algorithm, len(values), len(originalMAStatisticOrder))
	}

	for index, value := range values {
		if !finite(value) {
			return fmt.Errorf("benchmark %s %s statistic %s is not finite",
				benchmarkID, algorithm, originalMAStatisticOrder[index])
		}
	}

	best, worst, mean, median, standardDeviation := values[0], values[1], values[2], values[3], values[4]
	if best > worst || mean < best || mean > worst || median < best || median > worst || standardDeviation < 0 {
		return fmt.Errorf("benchmark %s %s statistics are inconsistent", benchmarkID, algorithm)
	}

	return nil
}

func buildBasicMAComparisonSummary(
	reference originalMAReference,
	inputs []basicMAComparisonInput,
) (publishedComparisonSummary, error) {
	summary := publishedComparisonSummary{
		SchemaVersion:                     publishedComparisonSchema,
		ComparisonKind:                    "descriptive_non_reproduction",
		ProtocolID:                        reference.ProtocolID,
		ReferenceDOI:                      reference.Source.DOI,
		ReferenceSHA256:                   reference.SHA256,
		ReproductionClaim:                 false,
		PublishedStandardDeviationMeaning: "unknown",
		UnresolvedOperatorSemantics:       append([]string(nil), reference.IMA.UnresolvedOperatorSemantics...),
		UnavailablePublishedAlgorithms:    []string{"vgma", "sma", "ima"},
		Comparisons:                       make([]publishedBenchmarkComparison, 0, len(inputs)),
	}

	for _, input := range inputs {
		comparison, err := buildBasicMAComparison(reference, input)
		if err != nil {
			return publishedComparisonSummary{}, err
		}

		summary.Comparisons = append(summary.Comparisons, comparison)
	}

	return summary, nil
}

func buildBasicMAComparison(
	reference originalMAReference,
	input basicMAComparisonInput,
) (publishedBenchmarkComparison, error) {
	benchmark, ok := findOriginalMABenchmark(reference.Benchmarks, input.BenchmarkID)
	if !ok {
		return publishedBenchmarkComparison{}, fmt.Errorf("unknown reference benchmark %q", input.BenchmarkID)
	}

	if input.Result == nil {
		return publishedBenchmarkComparison{}, fmt.Errorf("benchmark %s comparison result is nil", benchmark.ID)
	}

	algorithmIndex := -1

	for index, name := range input.Result.AlgorithmNames {
		if name == "MA" {
			algorithmIndex = index
			break
		}
	}

	if algorithmIndex < 0 || algorithmIndex >= len(input.Result.RunResults) {
		return publishedBenchmarkComparison{}, fmt.Errorf("benchmark %s has no current-library MA runs", benchmark.ID)
	}

	computed, availableRuns, failedRuns := summarizeRuns(input.Result.RunResults[algorithmIndex])
	if availableRuns == 0 {
		return publishedBenchmarkComparison{}, fmt.Errorf("benchmark %s has no successful finite MA runs", benchmark.ID)
	}

	publishedValues := reference.PublishedResults[benchmark.ID]["basic_ma"]
	published := publishedStatistics{
		Best:                      publishedValues[0],
		Worst:                     publishedValues[1],
		Mean:                      publishedValues[2],
		Median:                    publishedValues[3],
		ReportedStandardDeviation: publishedValues[4],
	}

	runs := input.Result.RunResults[algorithmIndex]
	alignment := assessBasicMAAlignment(reference, benchmark, input, runs)

	return publishedBenchmarkComparison{
		BenchmarkID:               benchmark.ID,
		BenchmarkName:             benchmark.Name,
		Dimension:                 input.Dimension,
		ComputedAlgorithm:         "current_library_ma",
		PublishedAlgorithm:        "basic_ma",
		RequestedRuns:             len(runs),
		AvailableRuns:             availableRuns,
		FailedRuns:                failedRuns,
		ExpectedEvaluationsPerRun: reference.Execution.FunctionEvaluations,
		Computed:                  computed,
		Published:                 published,
		ComputedMinusPublished: statisticDifferences{
			Best:   computed.Best - published.Best,
			Worst:  computed.Worst - published.Worst,
			Mean:   computed.Mean - published.Mean,
			Median: computed.Median - published.Median,
		},
		PublishedStdDevConvention: "unknown",
		Alignment:                 alignment,
	}, nil
}

func assessBasicMAAlignment(
	reference originalMAReference,
	benchmark originalMABenchmark,
	input basicMAComparisonInput,
	runs []mayfly.RunResult,
) protocolAlignment {
	config := input.Config
	alignment := protocolAlignment{
		BenchmarkGeometry: input.Dimension == benchmark.Dimension &&
			input.LowerBound == benchmark.LowerBound && input.UpperBound == benchmark.UpperBound,
		Replications:     len(runs) == reference.Execution.Replications,
		EvaluationBudget: len(runs) > 0,
		Population: config != nil && config.NPop == reference.Execution.MalePopulation &&
			config.NPopF == reference.Execution.FemalePopulation,
		KnownParameters: config != nil && config.A1 == reference.IMA.CommonParameters.A1 &&
			config.A2 == reference.IMA.CommonParameters.A2 &&
			config.Beta == reference.IMA.CommonParameters.Beta &&
			config.Dance == reference.IMA.CommonParameters.Dance &&
			config.FL == reference.IMA.CommonParameters.RandomFlight &&
			reference.IMA.CommonParameters.Selection == "linear rank pairing" &&
			config.Selection == mayfly.SelectionRank,
		OperatorSemantics:   false,
		PublishedSeedsKnown: false,
		Reasons:             []string{},
	}

	for _, run := range runs {
		if run.Error != "" || run.FuncEvals != reference.Execution.FunctionEvaluations {
			alignment.EvaluationBudget = false
			break
		}
	}

	if !alignment.BenchmarkGeometry {
		alignment.Reasons = append(alignment.Reasons, "benchmark dimension or bounds differ from Table 6")
	}

	if !alignment.Replications {
		alignment.Reasons = append(alignment.Reasons, "replication count differs from Table 6")
	}

	if !alignment.EvaluationBudget {
		alignment.Reasons = append(alignment.Reasons, "one or more runs failed or did not use exactly 95000 evaluations")
	}

	if !alignment.Population {
		alignment.Reasons = append(alignment.Reasons, "male or female population differs from Table 6")
	}

	if !alignment.KnownParameters {
		alignment.Reasons = append(alignment.Reasons, "current MA differs from one or more published numeric parameters")
	}

	alignment.Reasons = append(alignment.Reasons,
		"published crossover and mutation semantics are unresolved",
		"published replication seeds are unavailable, so samples are not paired",
	)

	return alignment
}

func summarizeRuns(runs []mayfly.RunResult) (computedStatistics, int, int) {
	costs := make([]float64, 0, len(runs))
	for _, run := range runs {
		if run.Error == "" && finite(run.BestCost) {
			costs = append(costs, run.BestCost)
		}
	}

	if len(costs) == 0 {
		var statistics computedStatistics

		return statistics, 0, len(runs)
	}

	sort.Float64s(costs)

	mean := 0.0
	for _, cost := range costs {
		mean += cost
	}

	mean /= float64(len(costs))

	median := costs[len(costs)/2]
	if len(costs)%2 == 0 {
		median = (costs[len(costs)/2-1] + costs[len(costs)/2]) * 0.5
	}

	return computedStatistics{
		Best:                  costs[0],
		Worst:                 costs[len(costs)-1],
		Mean:                  mean,
		Median:                median,
		PopulationStandardDev: populationStandardDeviation(costs, mean),
	}, len(costs), len(runs) - len(costs)
}

func populationStandardDeviation(values []float64, mean float64) float64 {
	scale := 0.0

	for _, value := range values {
		scale = max(scale, math.Abs(value-mean))
	}

	if scale == 0 {
		return 0
	}

	scaledSquares := 0.0

	for _, value := range values {
		scaledDifference := (value - mean) / scale
		scaledSquares += scaledDifference * scaledDifference
	}

	return scale * math.Sqrt(scaledSquares/float64(len(values)))
}

func findOriginalMABenchmark(benchmarks []originalMABenchmark, id string) (originalMABenchmark, bool) {
	for _, benchmark := range benchmarks {
		if benchmark.ID == id {
			return benchmark, true
		}
	}

	var benchmark originalMABenchmark

	return benchmark, false
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}

	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}

	return true
}
