package mayfly

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const wilcoxonTie = "Tie"

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
	BaseSeed       int64
}

// RunResult holds the result of a single optimization run.
type RunResult struct {
	Error         string
	BestCost      float64
	ExecutionTime float64 // Seconds
	Seed          int64
	FuncEvals     int
	Iterations    int
	ConvergenceAt int // Iteration where target was reached (0 if not reached)
}

// AlgorithmStatistics holds statistical measures for an algorithm's performance.
type AlgorithmStatistics struct {
	Mean          float64
	Median        float64
	StdDev        float64
	Best          float64
	Worst         float64
	SuccessRate   float64 // Percentage of runs reaching target
	AvgFuncEvals  float64
	AvgTime       float64 // Average execution time in seconds
	AvailableRuns int
	FailedRuns    int
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

// ComparisonRunner orchestrates multi-algorithm comparisons. When Parallel is
// true, the objective function must be safe for concurrent use. MaxWorkers
// limits concurrent optimization runs; any Config-level parallel evaluation is
// an independent inner limit.
type ComparisonRunner struct {
	TargetCost    *float64
	Variants      []AlgorithmVariant
	Runs          int
	MaxIterations int
	MaxWorkers    int
	Seed          int64
	Verbose       bool
	Parallel      bool
}

// NewComparisonRunner creates a new comparison runner.
func NewComparisonRunner() *ComparisonRunner {
	return &ComparisonRunner{
		Variants:      GetAllVariants(),
		Runs:          30, // Standard for statistical significance
		TargetCost:    nil,
		MaxIterations: 500,
		Verbose:       false,
		Parallel:      false,
		MaxWorkers:    runtime.NumCPU(),
		Seed:          time.Now().UnixNano(),
	}
}

// WithVariants sets the variants to compare.
// Deprecated: use WithVariantsChecked when variants are user supplied.
func (cr *ComparisonRunner) WithVariants(variants ...AlgorithmVariant) *ComparisonRunner {
	cr.Variants = variants
	return cr
}

// WithVariantsChecked rejects a nil runner, empty list, or nil variant.
func (cr *ComparisonRunner) WithVariantsChecked(variants ...AlgorithmVariant) (*ComparisonRunner, error) {
	if cr == nil {
		return nil, errors.New("comparison runner is nil")
	}

	if len(variants) == 0 {
		return nil, errors.New("at least one comparison variant is required")
	}

	for i, variant := range variants {
		if variant == nil {
			return nil, fmt.Errorf("comparison variant %d is nil", i)
		}
	}

	cr.Variants = append([]AlgorithmVariant(nil), variants...)

	return cr, nil
}

// WithVariantNames sets the variants to compare by name.
// Deprecated: use WithVariantNamesChecked. This compatibility method silently
// omits unknown names.
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

// WithVariantNamesChecked resolves every requested name or leaves the runner
// unchanged and returns an error.
func (cr *ComparisonRunner) WithVariantNamesChecked(names ...string) (*ComparisonRunner, error) {
	if cr == nil {
		return nil, errors.New("comparison runner is nil")
	}

	if len(names) == 0 {
		return nil, errors.New("at least one comparison variant name is required")
	}

	variants := make([]AlgorithmVariant, len(names))
	for i, name := range names {
		variant, err := NewVariantChecked(name)
		if err != nil {
			return nil, fmt.Errorf("comparison variant %d: %w", i, err)
		}

		variants[i] = variant
	}

	cr.Variants = variants

	return cr, nil
}

// WithRuns sets the number of runs per algorithm.
func (cr *ComparisonRunner) WithRuns(runs int) *ComparisonRunner {
	cr.Runs = runs
	return cr
}

// WithTarget sets the success threshold.
func (cr *ComparisonRunner) WithTarget(target float64) *ComparisonRunner {
	cr.TargetCost = &target
	return cr
}

// WithoutTarget disables success tracking.
func (cr *ComparisonRunner) WithoutTarget() *ComparisonRunner {
	cr.TargetCost = nil
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

// WithParallel enables or disables concurrent comparison runs.
func (cr *ComparisonRunner) WithParallel(parallel bool) *ComparisonRunner {
	cr.Parallel = parallel

	return cr
}

// WithMaxWorkers sets the maximum number of concurrent optimization runs.
// Zero uses runtime.NumCPU().
func (cr *ComparisonRunner) WithMaxWorkers(workers int) *ComparisonRunner {
	cr.MaxWorkers = workers

	return cr
}

// WithSeed sets the base seed used to derive one paired seed per run index.
func (cr *ComparisonRunner) WithSeed(seed int64) *ComparisonRunner {
	cr.Seed = seed

	return cr
}

// Compare runs all algorithms on the given problem and returns comparison results.
// Deprecated: use CompareContext; Compare preserves the historical behavior of
// returning a partial result while discarding validation and run errors.
func (cr *ComparisonRunner) Compare(
	benchmarkName string,
	fn ObjectiveFunction,
	problemSize int,
	lower, upper float64,
) *ComparisonResult {
	result, _ := cr.compare(context.Background(), benchmarkName, fn, problemSize, lower, upper, true)

	return result
}

// CompareContext runs all configured algorithms with cancellation and explicit
// error reporting. It returns no partial aggregate when any run fails.
func (cr *ComparisonRunner) CompareContext(
	ctx context.Context,
	benchmarkName string,
	fn ObjectiveFunction,
	problemSize int,
	lower, upper float64,
) (*ComparisonResult, error) {
	return cr.compare(ctx, benchmarkName, fn, problemSize, lower, upper, false)
}

type comparisonJob struct {
	config       *Config
	variantIndex int
	runIndex     int
	seed         int64
}

type comparisonJobResult struct {
	err          error
	run          RunResult
	variantIndex int
	runIndex     int
}

func (cr *ComparisonRunner) compare(
	ctx context.Context,
	benchmarkName string,
	fn ObjectiveFunction,
	problemSize int,
	lower, upper float64,
	continueOnError bool,
) (*ComparisonResult, error) {
	err := cr.validate(ctx, fn, problemSize, lower, upper)
	if err != nil {
		return &ComparisonResult{BenchmarkName: benchmarkName, BestAlgorithm: -1, BaseSeed: cr.Seed}, err
	}

	algorithmNames, runResults, jobs, err := cr.prepareJobs(fn, problemSize, lower, upper)
	if err != nil {
		return nil, err
	}

	err = cr.runJobs(ctx, jobs, algorithmNames, runResults, continueOnError)
	if err != nil {
		return nil, err
	}

	return cr.aggregate(benchmarkName, algorithmNames, runResults), nil
}

func (cr *ComparisonRunner) prepareJobs(
	fn ObjectiveFunction,
	problemSize int,
	lower, upper float64,
) ([]string, [][]RunResult, []comparisonJob, error) {
	algorithmNames := make([]string, len(cr.Variants))
	runResults := make([][]RunResult, len(cr.Variants))

	for i, variant := range cr.Variants {
		algorithmNames[i] = variant.Name()
		runResults[i] = make([]RunResult, cr.Runs)
	}

	jobs := make([]comparisonJob, 0, len(cr.Variants)*cr.Runs)
	for run := range cr.Runs {
		seed := cr.Seed + int64(run)

		for variantIndex, variant := range cr.Variants {
			config := cloneComparisonConfig(variant.GetConfig())
			if config == nil {
				return nil, nil, nil, fmt.Errorf("variant %s returned a nil config", variant.Name())
			}

			config.ObjectiveFunc = fn
			config.ProblemSize = problemSize
			config.LowerBound = lower
			config.UpperBound = upper
			config.MaxIterations = cr.MaxIterations
			config.Rand = rand.New(rand.NewSource(seed))
			jobs = append(jobs, comparisonJob{
				config: config, variantIndex: variantIndex, runIndex: run, seed: seed,
			})
		}
	}

	return algorithmNames, runResults, jobs, nil
}

func (cr *ComparisonRunner) runJobs(
	ctx context.Context,
	jobs []comparisonJob,
	algorithmNames []string,
	runResults [][]RunResult,
	continueOnError bool,
) error {
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobCh := make(chan comparisonJob, len(jobs))
	resultCh := make(chan comparisonJobResult, len(jobs))

	for _, job := range jobs {
		jobCh <- job
	}

	close(jobCh)

	var workers sync.WaitGroup
	workers.Add(cr.comparisonWorkerCount(len(jobs)))

	for range cr.comparisonWorkerCount(len(jobs)) {
		go cr.comparisonWorker(workerCtx, jobCh, resultCh, &workers)
	}

	go func() {
		workers.Wait()
		close(resultCh)
	}()

	firstErr := cr.collectJobResults(resultCh, algorithmNames, runResults, len(jobs), continueOnError, cancel)
	if firstErr != nil && !continueOnError {
		return firstErr
	}

	contextErr := ctx.Err()
	if contextErr != nil && !continueOnError {
		return contextErr
	}

	return nil
}

func (cr *ComparisonRunner) comparisonWorkerCount(jobCount int) int {
	if !cr.Parallel {
		return 1
	}

	workerCount := cr.MaxWorkers
	if workerCount == 0 {
		workerCount = runtime.NumCPU()
	}

	return min(workerCount, jobCount)
}

func (cr *ComparisonRunner) comparisonWorker(
	ctx context.Context,
	jobs <-chan comparisonJob,
	results chan<- comparisonJobResult,
	workers *sync.WaitGroup,
) {
	defer workers.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-jobs:
			if !ok {
				return
			}

			results <- cr.executeJob(ctx, job)
		}
	}
}

func (cr *ComparisonRunner) collectJobResults(
	results <-chan comparisonJobResult,
	algorithmNames []string,
	runResults [][]RunResult,
	jobCount int,
	continueOnError bool,
	cancel context.CancelFunc,
) error {
	var firstErr error

	completed := 0

	for jobResult := range results {
		completed++

		if jobResult.err != nil && firstErr == nil {
			name := algorithmNames[jobResult.variantIndex]
			firstErr = fmt.Errorf("compare %s run %d: %w", name, jobResult.runIndex+1, jobResult.err)

			if !continueOnError {
				cancel()
			}
		}

		runResults[jobResult.variantIndex][jobResult.runIndex] = jobResult.run

		if cr.Verbose {
			fmt.Printf("Completed %d/%d comparison runs\n", completed, jobCount)
		}
	}

	return firstErr
}

func (cr *ComparisonRunner) validate(
	ctx context.Context,
	fn ObjectiveFunction,
	problemSize int,
	lower, upper float64,
) error {
	if ctx == nil {
		return errNilContext
	}

	if len(cr.Variants) == 0 {
		return errors.New("at least one comparison variant is required")
	}

	for i, variant := range cr.Variants {
		if variant == nil {
			return fmt.Errorf("comparison variant %d is nil", i)
		}
	}

	if cr.Runs <= 0 {
		return fmt.Errorf("comparison runs must be positive, got %d", cr.Runs)
	}

	if cr.MaxIterations <= 0 {
		return fmt.Errorf("comparison iterations must be positive, got %d", cr.MaxIterations)
	}

	if cr.MaxWorkers < 0 {
		return fmt.Errorf("comparison MaxWorkers must be non-negative, got %d", cr.MaxWorkers)
	}

	if cr.TargetCost != nil && (math.IsNaN(*cr.TargetCost) || math.IsInf(*cr.TargetCost, 0)) {
		return errors.New("comparison target cost must be finite")
	}

	if fn == nil {
		return errors.New("comparison objective function is required")
	}

	if problemSize <= 0 {
		return fmt.Errorf("comparison problem size must be positive, got %d", problemSize)
	}

	if math.IsNaN(lower) || math.IsInf(lower, 0) || math.IsNaN(upper) || math.IsInf(upper, 0) {
		return errors.New("comparison bounds must be finite")
	}

	if lower >= upper {
		return fmt.Errorf("comparison lower bound %v must be less than upper bound %v", lower, upper)
	}

	return ctx.Err()
}

func (cr *ComparisonRunner) executeJob(ctx context.Context, job comparisonJob) comparisonJobResult {
	start := time.Now()
	result, err := OptimizeContext(ctx, job.config)

	run := RunResult{ExecutionTime: time.Since(start).Seconds(), Seed: job.seed}
	if err != nil {
		run.Error = err.Error()
		return comparisonJobResult{run: run, variantIndex: job.variantIndex, runIndex: job.runIndex, err: err}
	}

	if cr.TargetCost != nil {
		for iter, cost := range result.ConvergenceCurve {
			if cost <= *cr.TargetCost {
				run.ConvergenceAt = iter + 1
				break
			}
		}
	}

	run.BestCost = result.GlobalBest.Cost
	run.FuncEvals = result.FuncEvalCount
	run.Iterations = result.IterationCount

	return comparisonJobResult{run: run, variantIndex: job.variantIndex, runIndex: job.runIndex}
}

func (cr *ComparisonRunner) aggregate(
	benchmarkName string,
	algorithmNames []string,
	runResults [][]RunResult,
) *ComparisonResult {
	// Calculate statistics
	statistics := make([]AlgorithmStatistics, len(cr.Variants))
	for i := range cr.Variants {
		statistics[i] = calculateAlgorithmStatisticsWithTarget(runResults[i], cr.TargetCost)
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
		BaseSeed:       cr.Seed,
	}
}

// calculateAlgorithmStatistics computes statistical measures for run results.
func calculateAlgorithmStatistics(runs []RunResult, targetCost float64) AlgorithmStatistics {
	var target *float64
	if targetCost != 0 {
		target = &targetCost
	}

	return calculateAlgorithmStatisticsWithTarget(runs, target)
}

func calculateAlgorithmStatisticsWithTarget(runs []RunResult, targetCost *float64) AlgorithmStatistics {
	if len(runs) == 0 {
		return AlgorithmStatistics{}
	}

	costs := make([]float64, 0, len(runs))
	funcEvals := 0.0
	execTime := 0.0
	successCount := 0

	for _, run := range runs {
		if run.Error != "" || math.IsNaN(run.BestCost) || math.IsInf(run.BestCost, 0) {
			continue
		}

		costs = append(costs, run.BestCost)
		funcEvals += float64(run.FuncEvals)
		execTime += run.ExecutionTime

		if targetCost != nil && run.BestCost <= *targetCost {
			successCount++
		}
	}

	if len(costs) == 0 {
		return AlgorithmStatistics{FailedRuns: len(runs)}
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
	successRate := 0.0
	if targetCost != nil {
		successRate = float64(successCount) / float64(len(costs)) * 100.0
	}

	return AlgorithmStatistics{
		Mean:          mean,
		Median:        median,
		StdDev:        stdDev,
		Best:          best,
		Worst:         worst,
		SuccessRate:   successRate,
		AvgFuncEvals:  funcEvals / float64(len(costs)),
		AvgTime:       execTime / float64(len(costs)),
		AvailableRuns: len(costs),
		FailedRuns:    len(runs) - len(costs),
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
		mean := stat.Mean
		if stat.AvailableRuns == 0 && stat.FailedRuns > 0 {
			mean = math.Inf(1)
		}

		indexed[i] = indexedStat{index: i, mean: mean}
	}

	// Sort by mean (ascending - lower is better)
	sort.Slice(indexed, func(i, j int) bool {
		return indexed[i].mean < indexed[j].mean
	})

	// Assign equal competition ranks to exact ties.
	rankings := make([]int, len(statistics))
	for start := 0; start < len(indexed); {
		end := start + 1
		for end < len(indexed) && indexed[end].mean == indexed[start].mean {
			end++
		}

		for _, item := range indexed[start:end] {
			rankings[item.index] = start + 1
		}

		start = end
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
	for i := range n {
		if runs1[i].Error != "" || runs2[i].Error != "" ||
			!isFinite(runs1[i].BestCost) || !isFinite(runs2[i].BestCost) {
			continue
		}

		diff := runs1[i].BestCost - runs2[i].BestCost
		if math.Abs(diff) > 1e-10 { // Ignore ties
			differences = append(differences, diff)
			absDifferences = append(absDifferences, math.Abs(diff))
		}
	}

	if len(differences) == 0 {
		// Every pair tied, so there is no evidence of a difference at all.
		// PValue has to say so explicitly: its zero value reads as p = 0,
		// which is the most significant result the field can hold, and a
		// caller printing it sees "0.000" where the truth is that the two
		// algorithms produced the same costs on every run.
		return WilcoxonResult{
			Algorithm1: name1,
			Algorithm2: name2,
			Winner:     wilcoxonTie,
			PValue:     1.0,
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

	nEffective := float64(len(differences))
	meanW := nEffective * (nEffective + 1) / 4.0

	var pValue float64
	if len(differences) <= 20 {
		pValue = exactWilcoxonPValue(ranks, w)
	} else {
		variance := nEffective * (nEffective + 1) * (2*nEffective + 1) / 24.0

		variance -= wilcoxonTieCorrection(absDifferences) / 48.0
		if variance <= 0 {
			pValue = 1
		} else {
			z := max(math.Abs(wPlus-meanW)-0.5, 0) / math.Sqrt(variance)
			pValue = 2.0 * (1.0 - normalCDF(z))
		}
	}

	significant := pValue < 0.05

	winner := wilcoxonTie

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

func exactWilcoxonPValue(ranks []float64, observedW float64) float64 {
	total := 0.0
	for _, rank := range ranks {
		total += rank
	}

	assignments := uint64(1) << len(ranks)
	extreme := uint64(0)

	for mask := range assignments {
		positive := 0.0

		for i, rank := range ranks {
			if mask&(uint64(1)<<i) != 0 {
				positive += rank
			}
		}

		if min(positive, total-positive) <= observedW+1e-12 {
			extreme++
		}
	}

	return float64(extreme) / float64(assignments)
}

func wilcoxonTieCorrection(values []float64) float64 {
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)

	correction := 0.0

	for start := 0; start < len(sorted); {
		end := start + 1
		for end < len(sorted) && math.Abs(sorted[end]-sorted[start]) < 1e-10 {
			end++
		}

		tie := float64(end - start)
		correction += tie*tie*tie - tie
		start = end
	}

	return correction
}

// friedmanTest performs a Friedman test across all algorithms.
func friedmanTest(runResults [][]RunResult) *FriedmanTestResult {
	if len(runResults) < 2 {
		return nil
	}

	k := len(runResults) // Number of algorithms
	ranks := make([][]float64, 0, len(runResults[0]))
	tieSum := 0.0

	for run := range len(runResults[0]) {
		costs := make([]float64, k)
		valid := true

		for alg := range k {
			if run >= len(runResults[alg]) || runResults[alg][run].Error != "" ||
				!isFinite(runResults[alg][run].BestCost) {
				valid = false
				break
			}

			costs[alg] = runResults[alg][run].BestCost
		}

		if !valid {
			continue
		}

		ranks = append(ranks, rankValues(costs))
		tieSum += friedmanTieSum(costs)
	}

	n := len(ranks)
	if n == 0 {
		return nil
	}

	// Calculate sum of ranks for each algorithm
	rankSums := make([]float64, k)

	for alg := range k {
		for run := range n {
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

	tieCorrection := 1.0 - tieSum/(float64(n)*float64(k*k*k-k))
	if tieCorrection <= 0 {
		chiSquare = 0
	} else {
		chiSquare /= tieCorrection
	}

	// Degrees of freedom
	df := k - 1

	// The p-value is the upper tail: the chance of a statistic at least this
	// extreme when every algorithm performs identically.
	pValue := chiSquareSurvival(chiSquare, df)

	return &FriedmanTestResult{
		ChiSquare:        chiSquare,
		PValue:           pValue,
		Significant:      pValue < 0.05,
		DegreesOfFreedom: df,
	}
}

func friedmanTieSum(values []float64) float64 {
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)

	sum := 0.0

	for start := 0; start < len(sorted); {
		end := start + 1
		for end < len(sorted) && math.Abs(sorted[end]-sorted[start]) < 1e-10 {
			end++
		}

		tie := float64(end - start)
		sum += tie*tie*tie - tie
		start = end
	}

	return sum
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

// chiSquareSurvival returns the upper-tail probability of the chi-square
// distribution: P(X > x) with df degrees of freedom. That is the p-value the
// Friedman test needs.
//
// It replaces an earlier "chiSquareCDF" whose small-df branch returned
// exp(-x/2) * (x/2)^(df/2) — a curve that is neither a CDF nor monotonic, and
// which was additionally used as a p-value without taking the complement. The
// two errors did not cancel: the reported significance was inverted, so a
// strong result read as no difference and vice versa.
func chiSquareSurvival(x float64, df int) float64 {
	if df <= 0 {
		return math.NaN()
	}

	if x <= 0 {
		return 1
	}

	return regularizedGammaQ(float64(df)/2.0, x/2.0)
}

// regularizedGammaQ is Q(a, x), the regularized upper incomplete gamma
// function. The series converges quickly for x below a+1 and the continued
// fraction for x above it, which is the standard split.
func regularizedGammaQ(a, x float64) float64 {
	if x < 0 || a <= 0 {
		return math.NaN()
	}

	if x == 0 {
		return 1
	}

	if x < a+1 {
		return 1 - lowerGammaSeries(a, x)
	}

	return upperGammaContinuedFraction(a, x)
}

const (
	gammaMaxIterations = 300
	gammaEpsilon       = 3e-14
	gammaTiny          = 1e-300
)

// lowerGammaSeries evaluates P(a, x) by its series expansion.
func lowerGammaSeries(a, x float64) float64 {
	logPrefactor := -x + a*math.Log(x) - logGamma(a)
	term := 1.0 / a
	sum := term
	next := a

	for range gammaMaxIterations {
		next++
		term *= x / next
		sum += term

		if math.Abs(term) < math.Abs(sum)*gammaEpsilon {
			break
		}
	}

	return sum * math.Exp(logPrefactor)
}

// upperGammaContinuedFraction evaluates Q(a, x) by its continued fraction,
// using the modified Lentz algorithm.
func upperGammaContinuedFraction(a, x float64) float64 {
	logPrefactor := -x + a*math.Log(x) - logGamma(a)

	b := x + 1 - a
	c := 1 / gammaTiny
	d := 1 / b
	h := d

	for i := 1; i <= gammaMaxIterations; i++ {
		an := -float64(i) * (float64(i) - a)
		b += 2

		d = an*d + b
		if math.Abs(d) < gammaTiny {
			d = gammaTiny
		}

		c = b + an/c
		if math.Abs(c) < gammaTiny {
			c = gammaTiny
		}

		d = 1 / d
		delta := d * c
		h *= delta

		if math.Abs(delta-1) < gammaEpsilon {
			break
		}
	}

	return h * math.Exp(logPrefactor)
}

func logGamma(x float64) float64 {
	value, _ := math.Lgamma(x)

	return value
}

// PrintComparisonResults prints a formatted comparison report to stdout.
// Deprecated: use WriteComparisonResults to receive validation/write errors.
func (cr *ComparisonResult) PrintComparisonResults() {
	_ = cr.WriteComparisonResults(os.Stdout)
}

// WriteComparisonResults writes a formatted statistical report and relative
// quality chart. Longer bars indicate better (lower) finite mean costs.
func (cr *ComparisonResult) WriteComparisonResults(w io.Writer) error {
	if w == nil {
		return errors.New("comparison report writer cannot be nil")
	}

	err := cr.validateShape()
	if err != nil {
		return err
	}

	var writeErr error

	writef := func(format string, args ...any) {
		if writeErr != nil {
			return
		}

		_, writeErr = fmt.Fprintf(w, format, args...)
	}

	line := strings.Repeat("=", 80)
	writef("\n%s\nBenchmark Comparison: %s\n%s\n", line, cr.BenchmarkName, line)
	writef("\nStatistical Summary:\n%s\n", strings.Repeat("-", 80))
	writef("%-10s | %8s | %8s | %8s | %8s | %8s | %5s\n",
		"Algorithm", "Mean", "Median", "StdDev", "Best", "Worst", "Rank")
	writef("%s\n", strings.Repeat("-", 80))

	for i, name := range cr.AlgorithmNames {
		stats := cr.Statistics[i]
		writef("%-10s | %8.2e | %8.2e | %8.2e | %8.2e | %8.2e | %5d\n",
			name, stats.Mean, stats.Median, stats.StdDev, stats.Best, stats.Worst, cr.Rankings[i])
	}

	writef("%s\n", strings.Repeat("-", 80))
	writef("\nBest Algorithm: %s (Rank 1)\n", cr.AlgorithmNames[cr.BestAlgorithm])

	writef("\nRelative Quality (lower mean cost is better):\n")

	for _, index := range cr.rankedIndices() {
		mean := cr.Statistics[index].Mean
		bar, label := cr.qualityBar(mean, 24)
		writef("%2d. %-10s |%-24s| %s\n", cr.Rankings[index], cr.AlgorithmNames[index], bar, label)
	}

	writef("\nSignificant Pairwise Differences (Wilcoxon signed-rank test, alpha=0.05):\n")
	writef("%s\n", strings.Repeat("-", 80))

	foundSignificant := false

	for i := range cr.AlgorithmNames {
		for j := i + 1; j < len(cr.AlgorithmNames); j++ {
			test := cr.WilcoxonTests[i][j]
			if test.Significant {
				foundSignificant = true

				writef("%s vs %s: p=%.4f, Winner: %s\n",
					test.Algorithm1, test.Algorithm2, test.PValue, test.Winner)
			}
		}
	}

	if !foundSignificant {
		writef("No significant differences found.\n")
	}

	if cr.FriedmanResult != nil {
		significance := "Not significant"
		if cr.FriedmanResult.Significant {
			significance = "Significant at alpha=0.05"
		}

		writef("\nFriedman Test (overall difference):\n")
		writef("  chi-square = %.4f, df = %d, p = %.4f (%s)\n",
			cr.FriedmanResult.ChiSquare,
			cr.FriedmanResult.DegreesOfFreedom,
			cr.FriedmanResult.PValue,
			significance)
	}

	writef("%s\n", line)

	return writeErr
}

// ExportToCSV writes one deterministic row per algorithm run, including the
// corresponding aggregate statistics.
func (cr *ComparisonResult) ExportToCSV(path string) (returnErr error) {
	err := cr.validateShape()
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create comparison CSV: %w", err)
	}

	defer func() {
		closeErr := file.Close()
		if returnErr == nil && closeErr != nil {
			returnErr = fmt.Errorf("close comparison CSV: %w", closeErr)
		}
	}()

	w := csv.NewWriter(file)
	header := []string{
		"benchmark", "algorithm", "rank", "run", "seed", "best_cost", "function_evaluations",
		"iterations", "convergence_at", "execution_seconds", "error", "mean", "median", "stddev",
		"best", "worst", "success_rate", "avg_function_evaluations", "avg_execution_seconds",
	}

	err = w.Write(header)
	if err != nil {
		return fmt.Errorf("write comparison CSV header: %w", err)
	}

	for algorithm, name := range cr.AlgorithmNames {
		stats := cr.Statistics[algorithm]
		for runIndex, run := range cr.RunResults[algorithm] {
			record := []string{
				cr.BenchmarkName, name, strconv.Itoa(cr.Rankings[algorithm]), strconv.Itoa(runIndex + 1),
				strconv.FormatInt(run.Seed, 10), strconv.FormatFloat(run.BestCost, 'g', -1, 64),
				strconv.Itoa(run.FuncEvals), strconv.Itoa(run.Iterations), strconv.Itoa(run.ConvergenceAt),
				strconv.FormatFloat(run.ExecutionTime, 'g', -1, 64), run.Error,
				strconv.FormatFloat(stats.Mean, 'g', -1, 64), strconv.FormatFloat(stats.Median, 'g', -1, 64),
				strconv.FormatFloat(stats.StdDev, 'g', -1, 64), strconv.FormatFloat(stats.Best, 'g', -1, 64),
				strconv.FormatFloat(stats.Worst, 'g', -1, 64), strconv.FormatFloat(stats.SuccessRate, 'g', -1, 64),
				strconv.FormatFloat(stats.AvgFuncEvals, 'g', -1, 64), strconv.FormatFloat(stats.AvgTime, 'g', -1, 64),
			}

			err = w.Write(record)
			if err != nil {
				return fmt.Errorf("write comparison CSV row: %w", err)
			}
		}
	}

	w.Flush()

	err = w.Error()
	if err != nil {
		return fmt.Errorf("flush comparison CSV: %w", err)
	}

	return nil
}

// ExportToJSON writes the complete comparison result as indented JSON.
func (cr *ComparisonResult) ExportToJSON(path string) error {
	err := cr.validateShape()
	if err != nil {
		return err
	}

	return writeJSONAtomic(cr, path)
}

func cloneComparisonConfig(config *Config) *Config {
	if config == nil {
		return nil
	}

	clone := *config
	if config.Convergence != nil {
		convergence := *config.Convergence
		if config.Convergence.TargetCost != nil {
			target := *config.Convergence.TargetCost
			convergence.TargetCost = &target
		}

		clone.Convergence = &convergence
	}

	if config.Constraints != nil {
		constraints := *config.Constraints
		constraints.Inequalities = append([]ConstraintFunction(nil), config.Constraints.Inequalities...)
		constraints.Equalities = append([]ConstraintFunction(nil), config.Constraints.Equalities...)
		clone.Constraints = &constraints
	}

	return &clone
}

func (cr *ComparisonResult) rankedIndices() []int {
	indices := make([]int, len(cr.AlgorithmNames))

	for i := range indices {
		indices[i] = i
	}

	sort.SliceStable(indices, func(i, j int) bool {
		return cr.Rankings[indices[i]] < cr.Rankings[indices[j]]
	})

	return indices
}

func (cr *ComparisonResult) qualityBar(mean float64, width int) (string, string) {
	if math.IsNaN(mean) || math.IsInf(mean, 0) {
		return strings.Repeat(" ", width), "failed/unavailable"
	}

	best, worst := math.Inf(1), math.Inf(-1)

	for _, stats := range cr.Statistics {
		if math.IsNaN(stats.Mean) || math.IsInf(stats.Mean, 0) {
			continue
		}

		best = min(best, stats.Mean)
		worst = max(worst, stats.Mean)
	}

	quality := 1.0
	if worst > best {
		quality = (worst - mean) / (worst - best)
	}

	quality = max(0, min(1, quality))
	filled := int(math.Round(quality * float64(width)))

	return strings.Repeat("#", filled) + strings.Repeat(" ", width-filled), fmt.Sprintf("mean=%g", mean)
}

func (cr *ComparisonResult) validateShape() error {
	if cr == nil {
		return errors.New("comparison result cannot be nil")
	}

	count := len(cr.AlgorithmNames)
	if count == 0 {
		return errors.New("comparison result has no algorithms")
	}

	if len(cr.RunResults) != count || len(cr.Statistics) != count || len(cr.Rankings) != count ||
		len(cr.WilcoxonTests) != count {
		return errors.New("comparison result fields have inconsistent algorithm counts")
	}

	if cr.BestAlgorithm < 0 || cr.BestAlgorithm >= count {
		return fmt.Errorf("comparison best algorithm index %d is out of range", cr.BestAlgorithm)
	}

	for i := range count {
		if len(cr.WilcoxonTests[i]) != count {
			return fmt.Errorf("comparison Wilcoxon row %d has inconsistent length", i)
		}
	}

	return nil
}
