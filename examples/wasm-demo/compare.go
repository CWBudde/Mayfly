//go:build js && wasm

package main

import (
	"context"
	"syscall/js"

	"github.com/cwbudde/mayfly"
)

const (
	maxCompareRuns       = 50
	maxCompareIterations = 2000
)

// jsCompare runs the library's statistical comparison framework over one
// benchmark function and returns everything it produced.
//
// One benchmark per call is the chunking, and the chunking is the cancellation
// mechanism. A call into Go blocks its thread's event loop for its entire
// duration, so a "stop" message posted by the page cannot be dispatched while
// this function is running; the worker gets its chance to see it in the gap
// between two calls. Sizing each call at one benchmark keeps that gap a second
// or two away, which is why the sweep can be stopped at all without touching
// the library's own loop.
func jsCompare(opts js.Value) any {
	benchmarkName := readString(opts, "benchmark", "Rastrigin")

	spec, ok := lookupBenchmark(benchmarkName)
	if !ok {
		return errorResult("compare: unknown benchmark %q", benchmarkName)
	}

	var (
		dimensions = clampInt(readInt(opts, "dimensions", 10), 2, maxDimensions)
		runs       = clampInt(readInt(opts, "runs", 5), 1, maxCompareRuns)
		iterations = clampInt(readInt(opts, "iterations", 200), 1, maxCompareIterations)
		seed       = int64(readFloat(opts, "seed", 42))
		target     = readFloat(opts, "target", 1e-8)
		lower      = readFloat(opts, "lower", spec.lower)
		upper      = readFloat(opts, "upper", spec.upper)
	)

	names := readStrings(opts, "variants", nil)

	runner := mayfly.NewComparisonRunner().
		WithRuns(runs).
		WithIterations(iterations).
		WithSeed(seed).
		WithTarget(target).
		WithVerbose(false).
		// Same reasoning as a single run: js/wasm schedules every goroutine
		// onto one thread, so a worker pool adds coordination and no speed.
		WithParallel(false)

	if len(names) > 0 {
		runner = runner.WithVariantNames(names...)

		// WithVariantNames silently drops names it does not recognize, which
		// would show up as a table quietly missing a column.
		if len(runner.Variants) != len(names) {
			return errorResult("compare: unknown variant in %v", names)
		}
	}

	result, err := runner.CompareContext(
		context.Background(),
		spec.name,
		spec.fn,
		dimensions,
		lower,
		upper,
	)
	if err != nil {
		return errorResult("compare: %v", err)
	}

	return compareResponse(result, spec, dimensions, runs, iterations, target)
}

func compareResponse(
	result *mayfly.ComparisonResult,
	spec benchmark,
	dimensions, runs, iterations int,
	target float64,
) map[string]any {
	statistics := make([]any, len(result.Statistics))

	for i, stat := range result.Statistics {
		statistics[i] = map[string]any{
			"algorithm":    result.AlgorithmNames[i],
			"mean":         jsNumber(stat.Mean),
			"median":       jsNumber(stat.Median),
			"stdDev":       jsNumber(stat.StdDev),
			"best":         jsNumber(stat.Best),
			"worst":        jsNumber(stat.Worst),
			"successRate":  jsNumber(stat.SuccessRate),
			"avgFuncEvals": jsNumber(stat.AvgFuncEvals),
			"avgTime":      jsNumber(stat.AvgTime),
			"rank":         result.Rankings[i],
			"costs":        runCosts(result.RunResults[i]),
		}
	}

	response := map[string]any{
		"benchmark":  result.BenchmarkName,
		"dimensions": dimensions,
		"runs":       runs,
		"iterations": iterations,
		"target":     jsNumber(target),
		"optimum":    jsNumber(spec.optimum),
		"baseSeed":   float64(result.BaseSeed),
		"algorithms": stringsToJS(result.AlgorithmNames),
		"best":       result.BestAlgorithm,
		"statistics": statistics,
		"wilcoxon":   wilcoxonToJS(result.WilcoxonTests),
	}

	if result.FriedmanResult != nil {
		response["friedman"] = map[string]any{
			"chiSquare":        jsNumber(result.FriedmanResult.ChiSquare),
			"pValue":           jsNumber(result.FriedmanResult.PValue),
			"significant":      result.FriedmanResult.Significant,
			"degreesOfFreedom": result.FriedmanResult.DegreesOfFreedom,
		}
	}

	return response
}

func runCosts(runResults []mayfly.RunResult) []any {
	costs := make([]any, len(runResults))
	for i, run := range runResults {
		costs[i] = jsNumber(run.BestCost)
	}

	return costs
}

func wilcoxonToJS(tests [][]mayfly.WilcoxonResult) []any {
	rows := make([]any, len(tests))

	for i, row := range tests {
		cells := make([]any, len(row))

		for j, test := range row {
			cells[j] = map[string]any{
				"algorithm1":  test.Algorithm1,
				"algorithm2":  test.Algorithm2,
				"winner":      test.Winner,
				"wStatistic":  jsNumber(test.WStatistic),
				"pValue":      jsNumber(test.PValue),
				"significant": test.Significant,
			}
		}

		rows[i] = cells
	}

	return rows
}

func stringsToJS(values []string) []any {
	items := make([]any, len(values))
	for i, value := range values {
		items[i] = value
	}

	return items
}
