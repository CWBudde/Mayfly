package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"

	"github.com/cwbudde/mayfly"
)

const (
	desmaTable3SummarySchema = 1
	desmaTable3ProtocolID    = "desma-2022-table3"
	desmaTable3Comparison    = "descriptive_non_reproduction"
	desmaTable3Dimension     = 30
	desmaTable3Runs          = 51
	desmaTable3Evaluations   = 300_000
	desmaTable3Iterations    = 4_167
)

var desmaTable3FidelityGates = []string{
	"The initial search radius is not published.",
	"Population size 50 is not identified as per-sex or combined.",
	"Complete base-MA settings, evaluation accounting, seeds, and raw runs are unavailable.",
	"The Equation 16 elite lifecycle and the crossover coefficient L in [-1,1] remain unresolved.",
}

type desmaTable3FunctionSummary struct {
	BenchmarkName             string  `json:"benchmark_name"`
	ResultCSV                 string  `json:"result_csv"`
	ResultJSON                string  `json:"result_json"`
	FunctionID                string  `json:"function_id"`
	KnownMinimum              float64 `json:"known_minimum"`
	MeanAbsoluteError         float64 `json:"mean_absolute_error"`
	FunctionNumber            int     `json:"function_number"`
	AvailableRuns             int     `json:"available_runs"`
	FunctionEvaluationsPerRun int     `json:"function_evaluations_per_run"`
}

type desmaTable3Summary struct {
	ProtocolID             string                       `json:"protocol_id"`
	ComparisonKind         string                       `json:"comparison_kind"`
	ComputedAlgorithm      string                       `json:"computed_algorithm"`
	SeedSchedule           string                       `json:"seed_schedule"`
	Results                []desmaTable3FunctionSummary `json:"results"`
	ImplementationGates    []string                     `json:"implementation_fidelity_gates"`
	SchemaVersion          int                          `json:"schema_version"`
	Dimension              int                          `json:"dimension"`
	Runs                   int                          `json:"runs_per_function"`
	MaxFunctionEvaluations int                          `json:"max_function_evaluations_per_run"`
	BaseSeed               int64                        `json:"base_seed"`
	ReproductionClaim      bool                         `json:"reproduction_claim"`
}

func runDESMATable3(ctx context.Context, opts options, output io.Writer) error {
	problems, err := mayfly.CEC2013Suite(os.DirFS(opts.desmaTable3Data), desmaTable3Dimension)
	if err != nil {
		return fmt.Errorf("load DESMA Table 3 CEC2013 data: %w", err)
	}

	variant, err := mayfly.NewVariantChecked("desma")
	if err != nil {
		return err
	}

	err = validateDESMATable3Accounting(variant.GetConfig())
	if err != nil {
		return err
	}

	err = os.MkdirAll(opts.outputDir, 0o755)
	if err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	protocol, err := newDESMATable3Manifest(opts, variant, problems)
	if err != nil {
		return fmt.Errorf("build DESMA Table 3 manifest: %w", err)
	}

	err = writeJSON(filepath.Join(opts.outputDir, "manifest.json"), protocol)
	if err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}

	summary := desmaTable3Summary{
		SchemaVersion:          desmaTable3SummarySchema,
		ProtocolID:             desmaTable3ProtocolID,
		ComparisonKind:         desmaTable3Comparison,
		ReproductionClaim:      false,
		ComputedAlgorithm:      "current_library_desma",
		Dimension:              desmaTable3Dimension,
		Runs:                   desmaTable3Runs,
		MaxFunctionEvaluations: desmaTable3Evaluations,
		BaseSeed:               opts.seed,
		SeedSchedule:           "seed(run_index) = base_seed + zero_based_run_index",
		ImplementationGates:    append([]string(nil), desmaTable3FidelityGates...),
		Results:                make([]desmaTable3FunctionSummary, 0, len(problems)),
	}

	fmt.Fprintf(output, "running DESMA Table 3 non-reproduction (28 functions x %d runs x %d evaluations)\n",
		desmaTable3Runs, desmaTable3Evaluations)

	for _, problem := range problems {
		functionID, csvName, jsonName := desmaTable3ResultNames(problem.Number())
		fmt.Fprintf(output, "running %s (%d/28)\n", problem.Name(), problem.Number())

		runner, configureErr := mayfly.NewComparisonRunner().WithVariantsChecked(variant)
		if configureErr != nil {
			return fmt.Errorf("configure current DESMA: %w", configureErr)
		}

		runner.WithRuns(desmaTable3Runs).
			WithIterations(desmaTable3Iterations).
			WithMaxEvaluations(desmaTable3Evaluations).
			WithSeed(opts.seed).
			WithParallel(opts.workers > 1).
			WithMaxWorkers(opts.workers)

		result, compareErr := runner.CompareContext(ctx, problem.Name(), problem.Objective(),
			problem.Dimension(), -100, 100)
		if compareErr != nil {
			return fmt.Errorf("run %s: %w", functionID, compareErr)
		}

		exportErr := result.ExportToCSV(filepath.Join(opts.outputDir, csvName))
		if exportErr != nil {
			return fmt.Errorf("export %s CSV: %w", functionID, exportErr)
		}

		exportErr = result.ExportToJSON(filepath.Join(opts.outputDir, jsonName))
		if exportErr != nil {
			return fmt.Errorf("export %s JSON: %w", functionID, exportErr)
		}

		functionSummary, summaryErr := summarizeDESMATable3Function(
			problem.Number(), problem.Name(), problem.Minimum(), csvName, jsonName, result,
		)
		if summaryErr != nil {
			return fmt.Errorf("summarize %s: %w", functionID, summaryErr)
		}

		summary.Results = append(summary.Results, functionSummary)
	}

	err = writeJSON(filepath.Join(opts.outputDir, "desma-table3-summary.json"), summary)
	if err != nil {
		return fmt.Errorf("write DESMA Table 3 summary: %w", err)
	}

	fmt.Fprintf(output, "wrote descriptive_non_reproduction results to %s\n", opts.outputDir)

	return nil
}

func newDESMATable3Manifest(
	opts options,
	variant mayfly.AlgorithmVariant,
	problems []*mayfly.BenchmarkCase,
) (manifest, error) {
	protocol, err := newManifest(opts, []mayfly.AlgorithmVariant{variant})
	if err != nil {
		return manifest{}, err
	}

	noReproduction := false
	protocol.Experiment = "DESMA 2022 Table 3 descriptive current-library run"
	protocol.ProtocolID = desmaTable3ProtocolID
	protocol.ComparisonKind = desmaTable3Comparison
	protocol.ReproductionClaim = &noReproduction
	protocol.Notes = append(protocol.Notes,
		"This protocol is explicitly descriptive_non_reproduction.",
		"The CEC2013 geometry, 51-run count, and 300,000-call budget match Table 3.",
		"Current-library DESMA is not a paper-exact preset; unresolved fidelity gates are recorded in "+
			"desma-table3-summary.json.",
		"The 4,167 iteration value is the minimum safety ceiling that consumes 300,000 calls under the "+
			"recorded current-library DESMA defaults; the paper does not publish its iteration count.",
		"Current defaults use 20 males, 20 females, 20 crossover offspring, two mutants, and ten elite calls: "+
			"40 initialization calls plus 72 calls per full iteration and eight calls in the final partial iteration.",
	)

	for _, problem := range problems {
		_, csvName, jsonName := desmaTable3ResultNames(problem.Number())
		protocol.Benchmarks = append(protocol.Benchmarks, benchmarkProtocol{
			Name:       problem.Name(),
			ResultCSV:  csvName,
			ResultJSON: jsonName,
			Dimension:  problem.Dimension(),
			LowerBound: -100,
			UpperBound: 100,
			Minimum:    problem.Minimum(),
		})
	}

	return protocol, nil
}

func desmaTable3ResultNames(function int) (string, string, string) {
	functionID := fmt.Sprintf("f%d", function)
	stem := fmt.Sprintf("cec2013-f%02d-30d", function)

	return functionID, stem + ".csv", stem + ".json"
}

func validateDESMATable3Accounting(config *mayfly.Config) error {
	if config == nil {
		return errors.New("current DESMA configuration is nil")
	}

	if !config.UseDESMA || config.NPop != 20 || config.NPopF != 20 ||
		config.NC != mayfly.NCAuto || config.NCRatio != 1 || config.NM != 0 || config.EliteCount != 10 {
		return fmt.Errorf(
			"current DESMA defaults changed; the fixed %d-iteration Table 3 accounting must be updated",
			desmaTable3Iterations,
		)
	}

	return nil
}

func summarizeDESMATable3Function(
	function int,
	name string,
	minimum float64,
	csvName string,
	jsonName string,
	result *mayfly.ComparisonResult,
) (desmaTable3FunctionSummary, error) {
	if result == nil {
		return desmaTable3FunctionSummary{}, errors.New("comparison result is nil")
	}

	if len(result.AlgorithmNames) != 1 || result.AlgorithmNames[0] != "DESMA" || len(result.RunResults) != 1 {
		return desmaTable3FunctionSummary{}, errors.New("comparison result must contain only current-library DESMA")
	}

	runs := result.RunResults[0]
	if len(runs) != desmaTable3Runs {
		return desmaTable3FunctionSummary{}, fmt.Errorf("result has %d runs, want %d", len(runs), desmaTable3Runs)
	}

	meanAbsoluteError := 0.0

	for index, run := range runs {
		if run.Error != "" {
			return desmaTable3FunctionSummary{}, fmt.Errorf("run %d failed: %s", index+1, run.Error)
		}

		if run.FuncEvals != desmaTable3Evaluations {
			return desmaTable3FunctionSummary{}, fmt.Errorf(
				"run %d used %d objective evaluations, want %d",
				index+1, run.FuncEvals, desmaTable3Evaluations,
			)
		}

		if math.IsNaN(run.BestCost) || math.IsInf(run.BestCost, 0) {
			return desmaTable3FunctionSummary{}, fmt.Errorf("run %d best cost is not finite", index+1)
		}

		meanAbsoluteError += math.Abs(run.BestCost - minimum)
	}

	meanAbsoluteError /= float64(len(runs))

	return desmaTable3FunctionSummary{
		FunctionID:                fmt.Sprintf("f%d", function),
		FunctionNumber:            function,
		BenchmarkName:             name,
		KnownMinimum:              minimum,
		MeanAbsoluteError:         meanAbsoluteError,
		AvailableRuns:             len(runs),
		FunctionEvaluationsPerRun: desmaTable3Evaluations,
		ResultCSV:                 csvName,
		ResultJSON:                jsonName,
	}, nil
}
