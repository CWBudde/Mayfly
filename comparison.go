package mayfly

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// ComparisonResult holds the results of comparing multiple algorithms.
type ComparisonResult struct {
	FriedmanResult *FriedmanTestResult
	BenchmarkName  string
	AlgorithmNames []string
	RunResults     [][]RunResult
	Statistics     []AlgorithmStatistics
	Rankings       []int
	WilcoxonTests  [][]WilcoxonResult
	BestAlgorithm  int
}

// RunResult holds the result of a single optimization run.
type RunResult struct {
	BestCost      float64
	FuncEvals     int
	Iterations    int
	ConvergenceAt int     // Iteration where target was reached (0 if not reached)
	ExecutionTime float64 // Seconds
}

// AlgorithmStatistics holds statistical measures for an algorithm's performance.
type AlgorithmStatistics struct {
	Mean         float64
	Median       float64
	StdDev       float64
	Best         float64
	Worst        float64
	SuccessRate  float64 // Percentage of runs reaching target
	AvgFuncEvals float64
	AvgTime      float64 // Average execution time in seconds
}

// WilcoxonResult holds the result of a Wilcoxon signed-rank test.
type WilcoxonResult struct {
	Algorithm1  string
	Algorithm2  string
	Winner      string
	WStatistic  float64
	PValue      float64
	Significant bool
}

// FriedmanTestResult holds the result of a Friedman test.
type FriedmanTestResult struct {
	ChiSquare        float64
	PValue           float64
	Significant      bool // True if p < 0.05
	DegreesOfFreedom int
}

// ComparisonRunner orchestrates multi-algorithm comparisons.
type ComparisonRunner struct {
	Variants      []AlgorithmVariant
	Runs          int     // Number of runs per algorithm
	TargetCost    float64 // Success threshold (optional, 0 = unused)
	MaxIterations int     // Max iterations per run
	Verbose       bool    // Print progress
}

// NewComparisonRunner creates a new comparison runner.
func NewComparisonRunner() *ComparisonRunner {
	return &ComparisonRunner{
		Variants:      GetAllVariants(),
		Runs:          30, // Standard for statistical significance
		TargetCost:    0,
		MaxIterations: 500,
		Verbose:       false,
	}
}

// WithVariants sets the variants to compare.
func (cr *ComparisonRunner) WithVariants(variants ...AlgorithmVariant) *ComparisonRunner {
	cr.Variants = variants
	return cr
}

// WithVariantNames sets the variants to compare by name.
func (cr *ComparisonRunner) WithVariantNames(names ...string) *ComparisonRunner {
	variants := make([]AlgorithmVariant, 0, len(names))

	for _, name := range names {
		variant := NewVariant(name)
		if variant != nil {
			variants = append(variants, variant)
		}
	}

	cr.Variants = variants

	return cr
}

// WithRuns sets the number of runs per algorithm.
func (cr *ComparisonRunner) WithRuns(runs int) *ComparisonRunner {
	cr.Runs = runs
	return cr
}

// WithTarget sets the success threshold.
func (cr *ComparisonRunner) WithTarget(target float64) *ComparisonRunner {
	cr.TargetCost = target
	return cr
}

// WithIterations sets the maximum iterations.
func (cr *ComparisonRunner) WithIterations(iterations int) *ComparisonRunner {
	cr.MaxIterations = iterations
	return cr
}

// WithVerbose enables verbose output.
func (cr *ComparisonRunner) WithVerbose(verbose bool) *ComparisonRunner {
	cr.Verbose = verbose
	return cr
}

// Compare runs all algorithms on the given problem and returns comparison results.
func (cr *ComparisonRunner) Compare(
	benchmarkName string,
	fn ObjectiveFunction,
	problemSize int,
	lower, upper float64,
) *ComparisonResult {
	algorithmNames := make([]string, len(cr.Variants))
	runResults := make([][]RunResult, len(cr.Variants))

	// Run each algorithm
	for i, variant := range cr.Variants {
		algorithmNames[i] = variant.Name()
		runResults[i] = make([]RunResult, cr.Runs)

		if cr.Verbose {
			fmt.Printf("Running %s (%d runs)...\n", variant.Name(), cr.Runs)
		}

		for run := 0; run < cr.Runs; run++ {
			config := variant.GetConfig()
			config.ObjectiveFunc = fn
			config.ProblemSize = problemSize
			config.LowerBound = lower
			config.UpperBound = upper
			config.MaxIterations = cr.MaxIterations

			start := time.Now()
			result, err := Optimize(config)
			elapsed := time.Since(start).Seconds()

			if err != nil {
				runResults[i][run] = RunResult{
					BestCost:      math.Inf(1),
					FuncEvals:     0,
					Iterations:    0,
					ConvergenceAt: 0,
					ExecutionTime: elapsed,
				}

				continue
			}

			// Find convergence iteration
			convergenceAt := 0

			if cr.TargetCost > 0 {
				for iter, cost := range result.BestSolution {
					if cost <= cr.TargetCost {
						convergenceAt = iter + 1
						break
					}
				}
			}

			runResults[i][run] = RunResult{
				BestCost:      result.GlobalBest.Cost,
				FuncEvals:     result.FuncEvalCount,
				Iterations:    result.IterationCount,
				ConvergenceAt: convergenceAt,
				ExecutionTime: elapsed,
			}

			if cr.Verbose && (run+1)%10 == 0 {
				fmt.Printf("  Completed %d/%d runs\n", run+1, cr.Runs)
			}
		}
	}

	// Calculate statistics
	statistics := make([]AlgorithmStatistics, len(cr.Variants))
	for i := range cr.Variants {
		statistics[i] = calculateAlgorithmStatistics(runResults[i], cr.TargetCost)
	}

	// Rank algorithms by mean performance
	rankings := rankAlgorithms(statistics)
	bestAlgorithm := 0

	for i, rank := range rankings {
		if rank == 1 {
			bestAlgorithm = i
			break
		}
	}

	// Perform pairwise Wilcoxon tests
	wilcoxonTests := make([][]WilcoxonResult, len(cr.Variants))
	for i := range cr.Variants {
		wilcoxonTests[i] = make([]WilcoxonResult, len(cr.Variants))

		for j := range cr.Variants {
			if i != j {
				wilcoxonTests[i][j] = wilcoxonSignedRankTest(
					algorithmNames[i],
					algorithmNames[j],
					runResults[i],
					runResults[j],
				)
			}
		}
	}

	// Perform Friedman test
	friedmanResult := friedmanTest(runResults)

	return &ComparisonResult{
		AlgorithmNames: algorithmNames,
		BenchmarkName:  benchmarkName,
		RunResults:     runResults,
		Statistics:     statistics,
		Rankings:       rankings,
		WilcoxonTests:  wilcoxonTests,
		FriedmanResult: friedmanResult,
		BestAlgorithm:  bestAlgorithm,
	}
}

// calculateAlgorithmStatistics computes statistical measures for run results.
func calculateAlgorithmStatistics(runs []RunResult, targetCost float64) AlgorithmStatistics {
	if len(runs) == 0 {
		return AlgorithmStatistics{}
	}

	costs := make([]float64, len(runs))
	funcEvals := 0.0
	execTime := 0.0
	successCount := 0

	for i, run := range runs {
		costs[i] = run.BestCost
		funcEvals += float64(run.FuncEvals)
		execTime += run.ExecutionTime

		if targetCost > 0 && run.BestCost <= targetCost {
			successCount++
		}
	}

	// Sort for median and best/worst
	sortedCosts := make([]float64, len(costs))
	copy(sortedCosts, costs)
	sort.Float64s(sortedCosts)

	// Mean
	mean := 0.0
	for _, cost := range costs {
		mean += cost
	}

	mean /= float64(len(costs))

	// Median
	median := sortedCosts[len(sortedCosts)/2]
	if len(sortedCosts)%2 == 0 {
		median = (sortedCosts[len(sortedCosts)/2-1] + sortedCosts[len(sortedCosts)/2]) / 2.0
	}

	// Standard deviation
	variance := 0.0

	for _, cost := range costs {
		diff := cost - mean
		variance += diff * diff
	}

	variance /= float64(len(costs))
	stdDev := math.Sqrt(variance)

	// Best and worst
	best := sortedCosts[0]
	worst := sortedCosts[len(sortedCosts)-1]

	// Success rate
	successRate := float64(successCount) / float64(len(runs)) * 100.0

	return AlgorithmStatistics{
		Mean:         mean,
		Median:       median,
		StdDev:       stdDev,
		Best:         best,
		Worst:        worst,
		SuccessRate:  successRate,
		AvgFuncEvals: funcEvals / float64(len(runs)),
		AvgTime:      execTime / float64(len(runs)),
	}
}

// rankAlgorithms ranks algorithms based on mean performance (1 = best).
func rankAlgorithms(statistics []AlgorithmStatistics) []int {
	type indexedStat struct {
		index int
		mean  float64
	}

	indexed := make([]indexedStat, len(statistics))
	for i, stat := range statistics {
		indexed[i] = indexedStat{index: i, mean: stat.Mean}
	}

	// Sort by mean (ascending - lower is better)
	sort.Slice(indexed, func(i, j int) bool {
		return indexed[i].mean < indexed[j].mean
	})

	// Assign ranks
	rankings := make([]int, len(statistics))
	for rank, item := range indexed {
		rankings[item.index] = rank + 1
	}

	return rankings
}

// wilcoxonSignedRankTest performs a Wilcoxon signed-rank test between two algorithms.
func wilcoxonSignedRankTest(name1, name2 string, runs1, runs2 []RunResult) WilcoxonResult {
	if len(runs1) != len(runs2) {
		return WilcoxonResult{
			Algorithm1: name1,
			Algorithm2: name2,
			Winner:     "Error: unequal sample sizes",
		}
	}

	n := len(runs1)
	differences := make([]float64, 0, n)
	absDifferences := make([]float64, 0, n)

	// Calculate differences
	for i := 0; i < n; i++ {
		diff := runs1[i].BestCost - runs2[i].BestCost
		if math.Abs(diff) > 1e-10 { // Ignore ties
			differences = append(differences, diff)
			absDifferences = append(absDifferences, math.Abs(diff))
		}
	}

	if len(differences) == 0 {
		return WilcoxonResult{
			Algorithm1: name1,
			Algorithm2: name2,
			Winner:     "Tie",
		}
	}

	// Rank absolute differences
	ranks := rankValues(absDifferences)

	// Calculate W+ and W- (sum of positive and negative ranks)
	wPlus := 0.0
	wMinus := 0.0

	for i, diff := range differences {
		if diff > 0 {
			wPlus += ranks[i]
		} else {
			wMinus += ranks[i]
		}
	}

	// W statistic is the smaller of W+ and W-
	w := math.Min(wPlus, wMinus)

	// Approximate p-value using normal approximation for large n
	nEffective := float64(len(differences))
	meanW := nEffective * (nEffective + 1) / 4.0
	stdW := math.Sqrt(nEffective * (nEffective + 1) * (2*nEffective + 1) / 24.0)
	z := math.Abs((w - meanW) / stdW)
	pValue := 2.0 * (1.0 - normalCDF(z)) // Two-tailed

	significant := pValue < 0.05

	winner := "Tie"

	if significant {
		if wPlus < wMinus {
			winner = name1 // Algorithm 1 has lower costs (better)
		} else {
			winner = name2
		}
	}

	return WilcoxonResult{
		Algorithm1:  name1,
		Algorithm2:  name2,
		WStatistic:  w,
		PValue:      pValue,
		Significant: significant,
		Winner:      winner,
	}
}

// friedmanTest performs a Friedman test across all algorithms.
func friedmanTest(runResults [][]RunResult) *FriedmanTestResult {
	if len(runResults) < 2 {
		return nil
	}

	k := len(runResults)    // Number of algorithms
	n := len(runResults[0]) // Number of runs

	// Rank algorithms for each run
	ranks := make([][]float64, n)

	for run := 0; run < n; run++ {
		costs := make([]float64, k)
		for alg := 0; alg < k; alg++ {
			costs[alg] = runResults[alg][run].BestCost
		}

		ranks[run] = rankValues(costs)
	}

	// Calculate sum of ranks for each algorithm
	rankSums := make([]float64, k)

	for alg := 0; alg < k; alg++ {
		for run := 0; run < n; run++ {
			rankSums[alg] += ranks[run][alg]
		}
	}

	// Calculate Friedman statistic
	sumSquaredRanks := 0.0
	for _, rankSum := range rankSums {
		sumSquaredRanks += rankSum * rankSum
	}

	chiSquare := (12.0 / (float64(n) * float64(k) * float64(k+1))) * sumSquaredRanks
	chiSquare -= 3.0 * float64(n) * float64(k+1)

	// Degrees of freedom
	df := k - 1

	// Approximate p-value using chi-square distribution
	pValue := chiSquareCDF(chiSquare, df)

	return &FriedmanTestResult{
		ChiSquare:        chiSquare,
		PValue:           1.0 - pValue,
		Significant:      (1.0 - pValue) < 0.05,
		DegreesOfFreedom: df,
	}
}

// rankValues assigns ranks to values (1 = smallest).
func rankValues(values []float64) []float64 {
	type indexedValue struct {
		index int
		value float64
	}

	indexed := make([]indexedValue, len(values))
	for i, v := range values {
		indexed[i] = indexedValue{index: i, value: v}
	}

	// Sort by value
	sort.Slice(indexed, func(i, j int) bool {
		return indexed[i].value < indexed[j].value
	})

	// Assign ranks (handle ties by averaging)
	ranks := make([]float64, len(values))

	i := 0
	for i < len(indexed) {
		j := i
		// Find all tied values
		for j < len(indexed) && math.Abs(indexed[j].value-indexed[i].value) < 1e-10 {
			j++
		}
		// Average rank for ties
		avgRank := 0.0
		for k := i; k < j; k++ {
			avgRank += float64(k + 1)
		}

		avgRank /= float64(j - i)
		// Assign average rank
		for k := i; k < j; k++ {
			ranks[indexed[k].index] = avgRank
		}

		i = j
	}

	return ranks
}

// normalCDF computes the cumulative distribution function of the standard normal distribution.
func normalCDF(x float64) float64 {
	return 0.5 * (1.0 + math.Erf(x/math.Sqrt2))
}

// chiSquareCDF computes an approximation of the chi-square CDF.
// This is a simplified implementation for common use cases.
func chiSquareCDF(x float64, df int) float64 {
	if x <= 0 {
		return 0
	}
	// Use incomplete gamma function approximation
	// For simplicity, use normal approximation for large df
	if df > 30 {
		z := (x - float64(df)) / math.Sqrt(2.0*float64(df))
		return normalCDF(z)
	}
	// For small df, use a rough approximation
	return math.Min(math.Exp(-x/2.0)*math.Pow(x/2.0, float64(df)/2.0), 1.0)
}

// PrintComparisonResults prints a formatted comparison report.
func (cr *ComparisonResult) PrintComparisonResults() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Printf("Benchmark Comparison: %s\n", cr.BenchmarkName)
	fmt.Println(strings.Repeat("=", 80))

	// Statistics table
	fmt.Println("\nStatistical Summary:")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("%-10s | %8s | %8s | %8s | %8s | %8s | %5s\n",
		"Algorithm", "Mean", "Median", "StdDev", "Best", "Worst", "Rank")
	fmt.Println(strings.Repeat("-", 80))

	for i, name := range cr.AlgorithmNames {
		stats := cr.Statistics[i]
		rank := cr.Rankings[i]
		fmt.Printf("%-10s | %8.2e | %8.2e | %8.2e | %8.2e | %8.2e | %5d\n",
			name, stats.Mean, stats.Median, stats.StdDev, stats.Best, stats.Worst, rank)
	}

	fmt.Println(strings.Repeat("-", 80))

	// Best algorithm
	fmt.Printf("\n🏆 Best Algorithm: %s (Rank 1)\n", cr.AlgorithmNames[cr.BestAlgorithm])

	// Wilcoxon tests (only significant results)
	fmt.Println("\nSignificant Pairwise Differences (Wilcoxon signed-rank test, α=0.05):")
	fmt.Println(strings.Repeat("-", 80))

	foundSignificant := false

	for i := range cr.AlgorithmNames {
		for j := i + 1; j < len(cr.AlgorithmNames); j++ {
			test := cr.WilcoxonTests[i][j]
			if test.Significant {
				foundSignificant = true

				fmt.Printf("%s vs %s: p=%.4f, Winner: %s\n",
					test.Algorithm1, test.Algorithm2, test.PValue, test.Winner)
			}
		}
	}

	if !foundSignificant {
		fmt.Println("No significant differences found.")
	}

	// Friedman test
	if cr.FriedmanResult != nil {
		fmt.Println("\nFriedman Test (overall difference):")
		fmt.Printf("  χ² = %.4f, df = %d, p = %.4f",
			cr.FriedmanResult.ChiSquare,
			cr.FriedmanResult.DegreesOfFreedom,
			cr.FriedmanResult.PValue)

		if cr.FriedmanResult.Significant {
			fmt.Println(" (Significant at α=0.05)")
		} else {
			fmt.Println(" (Not significant)")
		}
	}

	fmt.Println(strings.Repeat("=", 80))
}
