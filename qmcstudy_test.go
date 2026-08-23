//go:build qmcstudy

// The measurement behind Config.QMCInit.
//
// This file is behind a build tag because it is a study, not a test: it runs
// the standard benchmark suite three times over at thirty runs each and takes
// minutes, and it asserts nothing about which strategy wins. Asserting that
// would pin a statistical accident — the whole point of measuring is that the
// answer was not known in advance, and it is allowed to change.
//
//	go test -tags qmcstudy -run TestQMCInitStudy -timeout 60m -v
//
// The output is a table; docs/qmc-initialization.md records what it said.

package mayfly

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

const (
	qmcStudyRuns       = 30
	qmcStudyIterations = 500
)

// TestQMCInitStudy compares uniform, Sobol and Halton initial populations on
// every problem in benchmarkProblems.
//
// Runs are paired by seed: run r of each strategy starts from
// rand.NewSource(r), so the three strategies see the same sequence of random
// decisions up to the point where the initial population differs. They diverge
// immediately after that — a different starting population leads the algorithm
// to consume its generator differently — so the pairing removes the seed as a
// source of difference without making the runs step-for-step comparable. That
// is the usual arrangement in this literature and it is what the Wilcoxon
// signed-rank test below is being handed.
func TestQMCInitStudy(t *testing.T) {
	strategies := []string{QMCInitUniform, QMCInitSobol, QMCInitHalton}

	fmt.Printf("\nQMC initial population: %d runs, %d iterations, NPop=%d NPopF=%d\n\n",
		qmcStudyRuns, qmcStudyIterations, NewDefaultConfig().NPop, NewDefaultConfig().NPopF)
	fmt.Printf("%-18s %-8s %12s %12s %12s %7s %10s %8s\n",
		"problem", "init", "mean", "median", "stddev", "differ", "p vs unif", "better")

	wins := map[string]int{}
	losses := map[string]int{}

	for _, problem := range benchmarkProblems {
		results := make(map[string][]RunResult, len(strategies))

		for _, strategy := range strategies {
			results[strategy] = qmcStudyRunSuite(t, problem, strategy)
		}

		baseline := results[QMCInitUniform]

		for _, strategy := range strategies {
			stats := calculateAlgorithmStatistics(results[strategy], 0)

			verdict := ""
			pValue := math.NaN()
			differing := 0

			if strategy != QMCInitUniform {
				differing = countDiffering(results[strategy], baseline)
				test := wilcoxonSignedRankTest(strategy, QMCInitUniform, results[strategy], baseline)
				pValue = test.PValue

				if test.Significant {
					verdict = test.Winner

					if test.Winner == strategy {
						wins[strategy]++
					} else {
						losses[strategy]++
					}
				}
			}

			fmt.Printf("%-18s %-8s %12.4e %12.4e %12.4e %4d/%-2d %10.3f %8s\n",
				problem.Name, strategy, stats.Mean, stats.Median, stats.StdDev,
				differing, qmcStudyRuns, pValue, verdict)
		}

		fmt.Println()
	}

	fmt.Printf("Significant at p<0.05 over %d problems:\n", len(benchmarkProblems))

	for _, strategy := range strategies[1:] {
		fmt.Printf("  %-8s better on %d, worse on %d\n", strategy, wins[strategy], losses[strategy])
	}
}

// qmcStudyRunSuite runs one strategy on one problem, qmcStudyRuns times.
func qmcStudyRunSuite(t *testing.T, problem BenchmarkProblem, strategy string) []RunResult {
	t.Helper()

	runs := make([]RunResult, qmcStudyRuns)

	for run := range qmcStudyRuns {
		config := NewDefaultConfig()
		config.ObjectiveFunc = problem.Func
		config.ProblemSize = problem.Dimensions
		config.LowerBound = problem.LowerBound
		config.UpperBound = problem.UpperBound
		config.MaxIterations = qmcStudyIterations
		config.QMCInit = strategy
		config.Rand = rand.New(rand.NewSource(int64(run + 1)))

		result, err := Optimize(config)
		if err != nil {
			t.Fatalf("Optimize(%s, %s) error = %v", problem.Name, strategy, err)
		}

		runs[run] = RunResult{
			BestCost:   result.GlobalBest.Cost,
			FuncEvals:  result.FuncEvalCount,
			Iterations: result.IterationCount,
		}
	}

	return runs
}

// countDiffering reports how many of the paired runs the Wilcoxon test can
// actually see: it discards pairs closer together than 1e-10, so a problem
// both strategies solve to machine precision reduces to no evidence at all
// rather than to a verdict.
func countDiffering(runs, baseline []RunResult) int {
	differing := 0

	for i := range runs {
		if math.Abs(runs[i].BestCost-baseline[i].BestCost) > 1e-10 {
			differing++
		}
	}

	return differing
}
