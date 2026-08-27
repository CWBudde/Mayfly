// Command paper-reproduction runs a fixed-seed, paired comparison of all
// current Mayfly variants on the library's classic benchmark functions.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"

	"github.com/cwbudde/mayfly"
)

const schemaVersion = 2

type options struct {
	outputDir          string
	publishedReference string
	desmaTable3Data    string
	benchmarks         []string
	variants           []string
	dimensions         []int
	runs               int
	iterations         int
	maxEvals           int
	workers            int
	seed               int64
}

type benchmark struct {
	fn      mayfly.ObjectiveFunction
	name    string
	lower   float64
	upper   float64
	minimum float64
}

type benchmarkProtocol struct {
	Name       string  `json:"name"`
	ResultCSV  string  `json:"result_csv"`
	ResultJSON string  `json:"result_json"`
	Dimension  int     `json:"dimension"`
	LowerBound float64 `json:"lower_bound"`
	UpperBound float64 `json:"upper_bound"`
	Minimum    float64 `json:"known_minimum"`
}

type variantProtocol struct {
	Parameters map[string]any `json:"parameters"`
	Name       string         `json:"name"`
	FullName   string         `json:"full_name"`
}

type manifest struct {
	Experiment        string              `json:"experiment"`
	Module            string              `json:"module"`
	Revision          string              `json:"revision,omitempty"`
	GoVersion         string              `json:"go_version"`
	GOOS              string              `json:"goos"`
	GOARCH            string              `json:"goarch"`
	SeedSchedule      string              `json:"seed_schedule"`
	ProtocolID        string              `json:"protocol_id,omitempty"`
	ComparisonKind    string              `json:"comparison_kind,omitempty"`
	ReproductionClaim *bool               `json:"reproduction_claim,omitempty"`
	Variants          []variantProtocol   `json:"variants"`
	Benchmarks        []benchmarkProtocol `json:"benchmarks"`
	Notes             []string            `json:"notes"`
	BaseSeed          int64               `json:"base_seed"`
	SchemaVersion     int                 `json:"schema_version"`
	Runs              int                 `json:"runs_per_algorithm"`
	Iterations        int                 `json:"iterations_per_run"`
	MaxEvaluations    int                 `json:"max_function_evaluations,omitempty"`
	Workers           int                 `json:"comparison_workers"`
	SourceDirty       bool                `json:"source_dirty,omitempty"`
}

var benchmarkRegistry = map[string]benchmark{
	"ackley":     {name: "Ackley", fn: mayfly.Ackley, lower: -32.768, upper: 32.768, minimum: 0},
	"beale":      {name: "Beale", fn: mayfly.Beale, lower: -4.5, upper: 4.5, minimum: 0},
	"eggcrate":   {name: "Eggcrate", fn: mayfly.Eggcrate, lower: -5, upper: 5, minimum: 0},
	"griewank":   {name: "Griewank", fn: mayfly.Griewank, lower: -600, upper: 600, minimum: 0},
	"rastrigin":  {name: "Rastrigin", fn: mayfly.Rastrigin, lower: -5.12, upper: 5.12, minimum: 0},
	"rosenbrock": {name: "Rosenbrock", fn: mayfly.Rosenbrock, lower: -30, upper: 30, minimum: 0},
	"schwefel":   {name: "Schwefel", fn: mayfly.Schwefel, lower: -500, upper: 500, minimum: 0},
	"sphere":     {name: "Sphere", fn: mayfly.Sphere, lower: -100, upper: 100, minimum: 0},
}

func main() {
	os.Exit(runMain(os.Args[1:], os.Stdout, os.Stderr))
}

func runMain(arguments []string, output, errorOutput io.Writer) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	opts, err := parseOptions(arguments)
	if err != nil {
		fmt.Fprintf(errorOutput, "paper reproduction: %v\n", err)

		return 2
	}

	err = runExperiment(ctx, opts, output)
	if err != nil {
		fmt.Fprintf(errorOutput, "paper reproduction: %v\n", err)

		return 1
	}

	return 0
}

func parseOptions(arguments []string) (options, error) {
	var (
		benchmarkNames string
		variantNames   string
		dimensions     string
	)

	opts := options{
		outputDir:          "",
		publishedReference: "",
		desmaTable3Data:    "",
		benchmarks:         nil,
		variants:           nil,
		dimensions:         nil,
		runs:               0,
		iterations:         0,
		maxEvals:           0,
		workers:            0,
		seed:               0,
	}
	flags := flag.NewFlagSet("paper-reproduction", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&opts.outputDir, "output", "paper-results", "directory for the manifest and result files")
	flags.StringVar(&opts.publishedReference, "published-reference", "",
		"Table 6 reference JSON for a descriptive current-MA comparison")
	flags.StringVar(&opts.desmaTable3Data, "desma-table3-data", "",
		"CEC2013 input-data directory for the fixed DESMA Table 3 non-reproduction protocol")
	flags.StringVar(&benchmarkNames, "benchmarks", "all", "comma-separated benchmark names, or all")
	flags.StringVar(&variantNames, "variants", "all", "comma-separated variant names, or all")
	flags.StringVar(&dimensions, "dimensions", "10,30", "comma-separated positive dimensions")
	flags.IntVar(&opts.runs, "runs", 30, "paired runs per algorithm")
	flags.IntVar(&opts.iterations, "iterations", 2000, "iterations per run")
	flags.IntVar(&opts.maxEvals, "max-evaluations", 0, "exact objective-evaluation budget per run; zero disables")
	flags.IntVar(&opts.workers, "workers", 1, "concurrent optimization runs")
	flags.Int64Var(&opts.seed, "seed", 20260825, "base seed; run i uses base seed+i")

	err := flags.Parse(arguments)
	if err != nil {
		return options{}, err
	}

	if flags.NArg() != 0 {
		return options{}, fmt.Errorf("unexpected positional arguments: %s", strings.Join(flags.Args(), " "))
	}

	if opts.outputDir == "" {
		return options{}, errors.New("output directory must not be empty")
	}

	if opts.runs <= 0 || opts.iterations <= 0 || opts.workers <= 0 || opts.maxEvals < 0 {
		return options{}, errors.New("runs, iterations, and workers must be positive; max-evaluations must be non-negative")
	}

	if opts.desmaTable3Data != "" {
		return configureDESMATable3Options(flags, opts)
	}

	if opts.publishedReference != "" {
		if strings.EqualFold(strings.TrimSpace(variantNames), "all") {
			variantNames = "ma"
		}

		opts.variants, err = selectVariantNames(variantNames)
		if err != nil {
			return options{}, err
		}

		if len(opts.variants) != 1 || opts.variants[0] != "ma" {
			return options{}, errors.New("published-reference mode supports only the current ma variant")
		}

		return opts, nil
	}

	opts.dimensions, err = parseDimensions(dimensions)
	if err != nil {
		return options{}, err
	}

	opts.benchmarks, err = selectNames(benchmarkNames, sortedBenchmarkNames(), "benchmark")
	if err != nil {
		return options{}, err
	}

	opts.variants, err = selectVariantNames(variantNames)
	if err != nil {
		return options{}, err
	}

	return opts, nil
}

func configureDESMATable3Options(flags *flag.FlagSet, opts options) (options, error) {
	if opts.publishedReference != "" {
		return options{}, errors.New("desma-table3-data and published-reference modes are mutually exclusive")
	}

	var fixedFlag string

	flags.Visit(func(value *flag.Flag) {
		switch value.Name {
		case "benchmarks", "dimensions", "iterations", "max-evaluations", "runs", "variants":
			if fixedFlag == "" {
				fixedFlag = value.Name
			}
		}
	})

	if fixedFlag != "" {
		return options{}, fmt.Errorf("-%s is fixed by desma-table3-data mode", fixedFlag)
	}

	opts.benchmarks = nil
	opts.variants = []string{"desma"}
	opts.dimensions = []int{desmaTable3Dimension}
	opts.runs = desmaTable3Runs
	opts.maxEvals = desmaTable3Evaluations
	// The paper does not report an iteration count. This is the minimum
	// ceiling under the current recorded DESMA defaults that consumes the
	// exact call budget, including its partial final generation.
	opts.iterations = desmaTable3Iterations

	return opts, nil
}

func parseDimensions(value string) ([]int, error) {
	parts := strings.Split(value, ",")
	dimensions := make([]int, 0, len(parts))
	seen := make(map[int]bool, len(parts))

	for _, part := range parts {
		dimension, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || dimension <= 0 {
			return nil, fmt.Errorf("invalid dimension %q", part)
		}

		if !seen[dimension] {
			dimensions = append(dimensions, dimension)
			seen[dimension] = true
		}
	}

	if len(dimensions) == 0 {
		return nil, errors.New("at least one dimension is required")
	}

	return dimensions, nil
}

func selectNames(value string, available []string, kind string) ([]string, error) {
	if strings.EqualFold(strings.TrimSpace(value), "all") {
		return append([]string(nil), available...), nil
	}

	known := make(map[string]bool, len(available))
	for _, name := range available {
		known[name] = true
	}

	selected := make([]string, 0)
	seen := make(map[string]bool)

	for _, rawName := range strings.Split(value, ",") {
		name := strings.ToLower(strings.TrimSpace(rawName))
		if !known[name] {
			return nil, fmt.Errorf("unknown %s %q (available: %s)", kind, rawName, strings.Join(available, ", "))
		}

		if !seen[name] {
			selected = append(selected, name)
			seen[name] = true
		}
	}

	if len(selected) == 0 {
		return nil, fmt.Errorf("at least one %s is required", kind)
	}

	return selected, nil
}

func selectVariantNames(value string) ([]string, error) {
	available := []string{"ma", "desma", "olce-ma", "eobbma", "gsasma", "hmma", "mpma", "aoblmoa"}
	if strings.EqualFold(strings.TrimSpace(value), "all") {
		return available, nil
	}

	aliases := map[string]string{"olce": "olce-ma", "standard": "ma"}
	parts := strings.Split(value, ",")

	for index, part := range parts {
		name := strings.ToLower(strings.TrimSpace(part))
		if canonical, ok := aliases[name]; ok {
			name = canonical
		}

		parts[index] = name
	}

	return selectNames(strings.Join(parts, ","), available, "variant")
}

func sortedBenchmarkNames() []string {
	names := make([]string, 0, len(benchmarkRegistry))
	for name := range benchmarkRegistry {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}

func runExperiment(ctx context.Context, opts options, output io.Writer) error {
	if opts.desmaTable3Data != "" {
		return runDESMATable3(ctx, opts, output)
	}

	if opts.publishedReference != "" {
		return runPublishedReferenceComparison(ctx, opts, output)
	}

	variants := make([]mayfly.AlgorithmVariant, len(opts.variants))
	for index, name := range opts.variants {
		variant, err := mayfly.NewVariantChecked(name)
		if err != nil {
			return err
		}

		variants[index] = variant
	}

	err := os.MkdirAll(opts.outputDir, 0o755)
	if err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	protocol, err := newManifest(opts, variants)
	if err != nil {
		return fmt.Errorf("build manifest: %w", err)
	}

	err = writeJSON(filepath.Join(opts.outputDir, "manifest.json"), protocol)
	if err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}

	for _, benchmarkName := range opts.benchmarks {
		bench := benchmarkRegistry[benchmarkName]
		for _, dimension := range opts.dimensions {
			caseName := fmt.Sprintf("%s-%dD", bench.name, dimension)
			stem := fmt.Sprintf("%s-%dd", benchmarkName, dimension)
			csvName := stem + ".csv"
			jsonName := stem + ".json"

			fmt.Fprintf(output, "running %s (%d variants x %d paired runs)\n", caseName, len(variants), opts.runs)

			runner, err := mayfly.NewComparisonRunner().WithVariantsChecked(variants...)
			if err != nil {
				return fmt.Errorf("configure variants: %w", err)
			}

			runner.WithRuns(opts.runs).
				WithIterations(opts.iterations).
				WithMaxEvaluations(opts.maxEvals).
				WithSeed(opts.seed).
				WithParallel(opts.workers > 1).
				WithMaxWorkers(opts.workers)

			result, err := runner.CompareContext(ctx, caseName, bench.fn, dimension, bench.lower, bench.upper)
			if err != nil {
				return fmt.Errorf("run %s: %w", caseName, err)
			}

			err = result.ExportToCSV(filepath.Join(opts.outputDir, csvName))
			if err != nil {
				return fmt.Errorf("export %s CSV: %w", caseName, err)
			}

			err = result.ExportToJSON(filepath.Join(opts.outputDir, jsonName))
			if err != nil {
				return fmt.Errorf("export %s JSON: %w", caseName, err)
			}
		}
	}

	fmt.Fprintf(output, "wrote protocol and raw results to %s\n", opts.outputDir)

	return nil
}

func runPublishedReferenceComparison(ctx context.Context, opts options, output io.Writer) error {
	if len(opts.variants) != 1 || opts.variants[0] != "ma" {
		return errors.New("published-reference mode supports only the current ma variant")
	}

	reference, err := loadOriginalMAReference(opts.publishedReference)
	if err != nil {
		return err
	}

	variant, err := mayfly.NewVariantChecked("ma")
	if err != nil {
		return err
	}

	err = os.MkdirAll(opts.outputDir, 0o755)
	if err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	protocol, err := newManifest(opts, []mayfly.AlgorithmVariant{variant})
	if err != nil {
		return fmt.Errorf("build manifest: %w", err)
	}

	protocol.Experiment = "Descriptive current-library MA comparison with original MA 2020 Table 6"
	protocol.Notes = append(protocol.Notes,
		"The published comparison is explicitly descriptive_non_reproduction.",
		"Only current-library MA is compared with the published Basic MA row; "+
			"VGMA, SMA, and IMA are not implemented as historical variants.",
		"The paper's crossover and Gaussian-mutation semantics remain unresolved, and its replication seeds are unavailable.",
	)

	for _, referenceBenchmark := range reference.Benchmarks {
		stem := referenceResultStem(referenceBenchmark)
		protocol.Benchmarks = append(protocol.Benchmarks, benchmarkProtocol{
			Name:       referenceBenchmark.Name,
			ResultCSV:  stem + ".csv",
			ResultJSON: stem + ".json",
			Dimension:  referenceBenchmark.Dimension,
			LowerBound: referenceBenchmark.LowerBound,
			UpperBound: referenceBenchmark.UpperBound,
			Minimum:    referenceBenchmark.KnownMinimum,
		})
	}

	err = writeJSON(filepath.Join(opts.outputDir, "manifest.json"), protocol)
	if err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}

	inputs := make([]basicMAComparisonInput, 0, len(reference.Benchmarks))
	for _, referenceBenchmark := range reference.Benchmarks {
		bench, ok := benchmarkRegistry[strings.ToLower(referenceBenchmark.Name)]
		if !ok {
			return fmt.Errorf("reference benchmark %s (%s) has no objective implementation",
				referenceBenchmark.ID, referenceBenchmark.Name)
		}

		caseName := fmt.Sprintf("%s %s-%dD", referenceBenchmark.ID,
			referenceBenchmark.Name, referenceBenchmark.Dimension)
		fmt.Fprintf(output, "running %s (current MA x %d runs)\n", caseName, opts.runs)

		runner, configureErr := mayfly.NewComparisonRunner().WithVariantsChecked(variant)
		if configureErr != nil {
			return fmt.Errorf("configure current MA: %w", configureErr)
		}

		runner.WithRuns(opts.runs).
			WithIterations(opts.iterations).
			WithMaxEvaluations(opts.maxEvals).
			WithSeed(opts.seed).
			WithParallel(opts.workers > 1).
			WithMaxWorkers(opts.workers)

		result, compareErr := runner.CompareContext(ctx, caseName, bench.fn,
			referenceBenchmark.Dimension, referenceBenchmark.LowerBound, referenceBenchmark.UpperBound)
		if compareErr != nil {
			return fmt.Errorf("run %s: %w", caseName, compareErr)
		}

		stem := referenceResultStem(referenceBenchmark)

		exportErr := result.ExportToCSV(filepath.Join(opts.outputDir, stem+".csv"))
		if exportErr != nil {
			return fmt.Errorf("export %s CSV: %w", caseName, exportErr)
		}

		exportErr = result.ExportToJSON(filepath.Join(opts.outputDir, stem+".json"))
		if exportErr != nil {
			return fmt.Errorf("export %s JSON: %w", caseName, exportErr)
		}

		inputs = append(inputs, basicMAComparisonInput{
			Result:      result,
			Config:      variant.GetConfig(),
			BenchmarkID: referenceBenchmark.ID,
			Dimension:   referenceBenchmark.Dimension,
			LowerBound:  referenceBenchmark.LowerBound,
			UpperBound:  referenceBenchmark.UpperBound,
		})
	}

	summary, err := buildBasicMAComparisonSummary(reference, inputs)
	if err != nil {
		return fmt.Errorf("build published comparison: %w", err)
	}

	err = writeJSON(filepath.Join(opts.outputDir, "published-comparison.json"), summary)
	if err != nil {
		return fmt.Errorf("write published comparison: %w", err)
	}

	fmt.Fprintf(output, "wrote descriptive published comparison and raw results to %s\n", opts.outputDir)

	return nil
}

func referenceResultStem(benchmark originalMABenchmark) string {
	return fmt.Sprintf("%s-%s-%dd", strings.ToLower(benchmark.ID),
		strings.ToLower(benchmark.Name), benchmark.Dimension)
}

func newManifest(opts options, variants []mayfly.AlgorithmVariant) (manifest, error) {
	revision, dirty := sourceRevision()
	protocol := manifest{
		SchemaVersion:     schemaVersion,
		Experiment:        "Mayfly post-v0.7 classic benchmark baseline",
		Module:            "github.com/cwbudde/mayfly",
		Revision:          revision,
		SourceDirty:       dirty,
		GoVersion:         runtime.Version(),
		GOOS:              runtime.GOOS,
		GOARCH:            runtime.GOARCH,
		BaseSeed:          opts.seed,
		SeedSchedule:      "seed(run_index) = base_seed + zero_based_run_index; the same seed is paired across variants",
		Runs:              opts.runs,
		Iterations:        opts.iterations,
		MaxEvaluations:    opts.maxEvals,
		Workers:           opts.workers,
		Variants:          []variantProtocol{},
		Benchmarks:        []benchmarkProtocol{},
		ProtocolID:        "",
		ComparisonKind:    "",
		ReproductionClaim: nil,
		Notes: []string{
			"No target-cost early stopping is enabled.",
			"CSV execution_seconds is machine-dependent; costs, seeds, evaluation counts, " +
				"and iterations are the reproducibility data.",
			"Different variants can consume different function-evaluation counts, so compare " +
				"function_evaluations as well as best_cost.",
			"This is a current-library baseline, not a claim that every source paper used this common protocol.",
		},
	}

	if opts.maxEvals > 0 {
		protocol.Notes = append(protocol.Notes,
			"Each successful run invokes the objective exactly max_function_evaluations times; "+
				"iterations_per_run is a safety ceiling and must be high enough to consume that budget.",
		)
	} else {
		protocol.Notes = append(protocol.Notes,
			"Every successful run uses the configured iteration count; no objective-evaluation cap is enabled.",
		)
	}

	for _, variant := range variants {
		config := *variant.GetConfig()

		parameters, err := variantParameters(config)
		if err != nil {
			return manifest{}, fmt.Errorf("snapshot %s parameters: %w", variant.Name(), err)
		}

		protocol.Variants = append(protocol.Variants, variantProtocol{
			Name:       variant.Name(),
			FullName:   variant.FullName(),
			Parameters: parameters,
		})
	}

	for _, benchmarkName := range opts.benchmarks {
		bench := benchmarkRegistry[benchmarkName]
		for _, dimension := range opts.dimensions {
			stem := fmt.Sprintf("%s-%dd", benchmarkName, dimension)
			protocol.Benchmarks = append(protocol.Benchmarks, benchmarkProtocol{
				Name:       bench.name,
				Dimension:  dimension,
				LowerBound: bench.lower,
				UpperBound: bench.upper,
				Minimum:    bench.minimum,
				ResultCSV:  stem + ".csv",
				ResultJSON: stem + ".json",
			})
		}
	}

	return protocol, nil
}

func variantParameters(config mayfly.Config) (map[string]any, error) {
	data, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("marshal configuration: %w", err)
	}

	parameters := make(map[string]any)

	err = json.Unmarshal(data, &parameters)
	if err != nil {
		return nil, fmt.Errorf("decode configuration: %w", err)
	}
	// These run-specific values are recorded once at the protocol or benchmark
	// level instead of being duplicated in every variant's parameter snapshot.
	for _, key := range []string{"lower_bound", "max_iterations", "problem_size", "seed", "upper_bound"} {
		delete(parameters, key)
	}

	return parameters, nil
}

func sourceRevision() (string, bool) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "", false
	}

	var (
		revision string
		dirty    bool
	)

	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			revision = setting.Value
		case "vcs.modified":
			dirty = setting.Value == "true"
		}
	}

	return revision, dirty
}

func writeJSON(path string, value any) (returnErr error) {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := file.Close()
		if returnErr == nil && closeErr != nil {
			returnErr = closeErr
		}
	}()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)

	return encoder.Encode(value)
}
