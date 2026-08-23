package mayfly

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"sync/atomic"
	"testing"
)

// aoblmoaStillMayfly builds a mayfly that sits exactly on top of its own
// personal best and is at rest.
//
// Such an individual is a fixed point of the Mayfly attraction branch: every
// attraction distance is zero, so the velocity stays zero and the position
// never changes. Any movement therefore proves it took the Aquila branch,
// which is what makes the branch decision observable without reimplementing
// either formula in the test.
func aoblmoaStillMayfly(size int, coordinate, cost float64) *Mayfly {
	member := newMayfly(size)

	for j := range size {
		member.Position[j] = coordinate
		member.Best.Position[j] = coordinate
	}

	member.Cost = cost
	member.Best.Cost = cost

	return member
}

func aoblmoaBranchConfig(seed int64) *Config {
	config := NewAOBLMOAConfig()
	config.Rand = rand.New(rand.NewSource(seed))
	config.ObjectiveFunc = Sphere
	config.ProblemSize = 3
	config.LowerBound = -5.0
	config.UpperBound = 5.0
	config.MaxIterations = 100
	config.VelMin = -2.0
	config.VelMax = 2.0

	return config
}

// TestAOBLMOAMaleBranchFollowsFitnessTest pins Eq. (29): a male keeps the
// Mayfly attraction term while the global best dominates him, and hunts as an
// Aquila otherwise.
//
// Through v0.5.1 the branch was a draw against AquilaWeight, whose default of
// 1.0 sent every male down the Aquila branch regardless of fitness, so the
// dominated male below would have moved.
func TestAOBLMOAMaleBranchFollowsFitnessTest(t *testing.T) {
	config := aoblmoaBranchConfig(101)

	// dominated is worse than the global best and must be attracted to it;
	// leader is better than it and must hunt.
	dominated := aoblmoaStillMayfly(config.ProblemSize, 1.0, 10.0)
	leader := aoblmoaStillMayfly(config.ProblemSize, 2.0, -10.0)
	males := []*Mayfly{dominated, leader}

	globalBest := Best{
		Position: make([]float64, config.ProblemSize),
		Cost:     0.0,
	}
	copy(globalBest.Position, dominated.Position)

	before := snapshotPositions(males)

	applyAOBLMOAMoves(
		males, nil, globalBest, 10, config.MaxIterations,
		config.G, config.Dance, config.FL, config, config.Rand,
		newConstraintEvaluator(config.ObjectiveFunc, nil),
	)

	if !samePosition(before[0], dominated.Position) {
		t.Errorf("a male the global best dominates moved from %v to %v; "+
			"Eq. (29) attracts him, and every attraction distance here is zero",
			before[0], dominated.Position)
	}

	if samePosition(before[1], leader.Position) {
		t.Errorf("a male the global best does not dominate stayed at %v; "+
			"Eq. (29) sends him down the Aquila branch", before[1])
	}
}

// TestAOBLMOAFemaleBranchFollowsFitnessTest pins Eq. (30): a female keeps the
// Mayfly attraction term while her paired male dominates her, and hunts as an
// Aquila otherwise.
//
// This is the direction aoblmoaFemaleTakesAttraction resolves. The paper's
// Algorithm 1 pseudocode states the opposite inequality; if that reading ever
// wins, this test's two expectations swap along with that one return.
func TestAOBLMOAFemaleBranchFollowsFitnessTest(t *testing.T) {
	config := aoblmoaBranchConfig(202)

	// Each female sits on her paired male, so the attraction term is exactly
	// zero and only the Aquila branch can move her.
	dominatedFemale := aoblmoaStillMayfly(config.ProblemSize, 1.0, 10.0)
	strongFemale := aoblmoaStillMayfly(config.ProblemSize, 2.0, -10.0)
	females := []*Mayfly{dominatedFemale, strongFemale}

	betterMale := aoblmoaStillMayfly(config.ProblemSize, 1.0, 0.0)
	worseMale := aoblmoaStillMayfly(config.ProblemSize, 2.0, 0.0)
	males := []*Mayfly{betterMale, worseMale}

	globalBest := Best{
		Position: make([]float64, config.ProblemSize),
		Cost:     math.Inf(-1),
	}

	before := snapshotPositions(females)

	applyAOBLMOAMoves(
		males, females, globalBest, 10, config.MaxIterations,
		config.G, config.Dance, config.FL, config, config.Rand,
		newConstraintEvaluator(config.ObjectiveFunc, nil),
	)

	if !samePosition(before[0], dominatedFemale.Position) {
		t.Errorf("a female her paired male dominates moved from %v to %v; "+
			"Eq. (30) attracts her, and the attraction distance here is zero",
			before[0], dominatedFemale.Position)
	}

	if samePosition(before[1], strongFemale.Position) {
		t.Errorf("a female who dominates her paired male stayed at %v; "+
			"Eq. (30) sends her down the Aquila branch", before[1])
	}
}

// TestAOBLMOAStrategyAssignmentIsSexAndPhaseDetermined pins the mapping that
// carries the paper's second contradiction.
//
// The equations give males the narrowed strategies and females the expanded
// ones; the abstract swaps them. Settling that question means swapping the two
// returned pairs in aoblmoaStrategyFor and the two halves of this table.
func TestAOBLMOAStrategyAssignmentIsSexAndPhaseDetermined(t *testing.T) {
	for _, testCase := range []struct {
		name        string
		isMale      bool
		exploration bool
		want        AquilaStrategy
	}{
		{name: "male exploring", isMale: true, exploration: true, want: NarrowedExploration},
		{name: "male exploiting", isMale: true, exploration: false, want: NarrowedExploitation},
		{name: "female exploring", isMale: false, exploration: true, want: ExpandedExploration},
		{name: "female exploiting", isMale: false, exploration: false, want: ExpandedExploitation},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			got := aoblmoaStrategyFor(testCase.isMale, testCase.exploration)
			if got != testCase.want {
				t.Errorf("aoblmoaStrategyFor(%t, %t) = %v, want %v",
					testCase.isMale, testCase.exploration, got, testCase.want)
			}
		})
	}
}

// TestAOBLMOAUpdatePhaseSpendsNoEvaluationsBeyondTheSwarm pins the budget of
// the update phase at exactly one evaluation per individual.
//
// Through v0.5.1 the phase also evaluated an Aquila candidate and its
// opposition point for a share of the swarm, because opposition-based learning
// ran there instead of on the offspring.
func TestAOBLMOAUpdatePhaseSpendsNoEvaluationsBeyondTheSwarm(t *testing.T) {
	const size = 10

	var calls atomic.Int64

	config := aoblmoaBranchConfig(303)
	config.ObjectiveFunc = func(position []float64) float64 {
		calls.Add(1)

		return Sphere(position)
	}

	males, females := aoblmoaTestPopulations(t, config, size)
	globalBest := Best{
		Position: append([]float64(nil), males[0].Best.Position...),
		Cost:     males[0].Cost,
	}

	// Only the update phase is being counted; building the populations already
	// evaluated everyone once.
	calls.Store(0)

	evaluations := applyAOBLMOAToPopulation(
		males, females, globalBest, 10, config.MaxIterations,
		config.G, config.Dance, config.FL, config,
	)

	if want := len(males) + len(females); evaluations != want {
		t.Errorf("AOBLMOA reported %d evaluations, want one per individual (%d)", evaluations, want)
	}

	if got := int(calls.Load()); got != len(males)+len(females) {
		t.Errorf("the objective was called %d times, want one call per individual (%d)",
			got, len(males)+len(females))
	}
}

// TestStochasticOppositionPointUsesOneUniformScalar pins the authors' source
// resolution of ambiguous Eq. (31): one U(0,1) scalar multiplies the complete
// reflected offspring vector.
func TestStochasticOppositionPointUsesOneUniformScalar(t *testing.T) {
	const (
		lower = -100.0
		upper = 100.0
		seed  = 404
	)

	position := []float64{1.0, -2.5, 7.25, 0.0}

	got := stochasticOppositionPoint(position, lower, upper, rand.New(rand.NewSource(seed)))

	// Replay the one draw to spell out the formula independently.
	replay := rand.New(rand.NewSource(seed))
	r := replay.Float64()
	for i, x := range position {
		want := (lower + upper - x) * r
		if got[i] != want {
			t.Errorf("dimension %d: got %v, want (lb+ub-x)*r = %v", i, got[i], want)
		}
	}

	plain := oppositionPoint(position, lower, upper)
	if samePosition(got, plain) {
		t.Error("the stochastic opposition point equals the plain reflection; " +
			"the Gaussian factor of Eq. (31) is missing")
	}
}

// TestStochasticOppositionPointStaysInBounds guards clipping for asymmetric
// bounds, where scaling the reflected coordinate can still leave the domain.
func TestStochasticOppositionPointStaysInBounds(t *testing.T) {
	const (
		lower = -5.0
		upper = 5.0
	)

	rng := rand.New(rand.NewSource(505))

	for range 200 {
		position := []float64{
			unifrnd(lower, upper, rng),
			unifrnd(lower, upper, rng),
		}

		for i, coordinate := range stochasticOppositionPoint(position, lower, upper, rng) {
			if coordinate < lower || coordinate > upper {
				t.Fatalf("dimension %d left the search bounds: %v", i, coordinate)
			}
		}
	}
}

// TestCommitStochasticOBLKeepsTheBetter pins the greedy selection of Eq. (32),
// including the personal-best refresh that keeps a replaced offspring
// self-consistent.
func TestCommitStochasticOBLKeepsTheBetter(t *testing.T) {
	evaluator := newConstraintEvaluator(Sphere, nil)

	improved := newMayfly(2)
	improved.Position = []float64{3, 3}
	improved.Cost = 18

	worsened := newMayfly(2)
	worsened.Position = []float64{0.5, 0.5}
	worsened.Cost = 0.5

	offspring := []*Mayfly{improved, worsened}

	betterOpposite := newMayfly(2)
	betterOpposite.Position = []float64{0, 0}
	betterOpposite.Cost = 0

	worseOpposite := newMayfly(2)
	worseOpposite.Position = []float64{4, 4}
	worseOpposite.Cost = 32

	commitStochasticOBL(offspring, []*Mayfly{betterOpposite, worseOpposite}, evaluator)

	if improved.Cost != 0 || !samePosition(improved.Position, []float64{0, 0}) {
		t.Errorf("the better opposition point was not kept: cost %v at %v",
			improved.Cost, improved.Position)
	}

	if improved.Best.Cost != 0 || !samePosition(improved.Best.Position, improved.Position) {
		t.Errorf("the replaced offspring's personal best was not refreshed: cost %v at %v",
			improved.Best.Cost, improved.Best.Position)
	}

	if worsened.Cost != 0.5 || !samePosition(worsened.Position, []float64{0.5, 0.5}) {
		t.Errorf("a worse opposition point replaced its offspring: cost %v at %v",
			worsened.Cost, worsened.Position)
	}
}

// TestAOBLMOAReplacesMutationWithOpposition pins the offspring stage: one
// opposition point per offspring, and no mutants at all.
//
// The arithmetic per iteration is NPop + NPopF (the update phase) + nc
// (crossover) + nc (one opposition point per offspring). NM is set high on
// purpose: under AOBLMOA it must buy nothing.
func TestAOBLMOAReplacesMutationWithOpposition(t *testing.T) {
	const (
		population = 6
		offspring  = 4
		iterations = 3
	)

	for _, parallel := range []bool{false, true} {
		t.Run(fmt.Sprintf("parallel_%t", parallel), func(t *testing.T) {
			var calls atomic.Int64

			config := NewAOBLMOAConfig()
			config.Rand = rand.New(rand.NewSource(606))
			config.ObjectiveFunc = func(position []float64) float64 {
				calls.Add(1)

				return Sphere(position)
			}
			config.ProblemSize = 3
			config.LowerBound = -5
			config.UpperBound = 5
			config.MaxIterations = iterations
			config.NPop = population
			config.NPopF = population
			config.NC = offspring
			config.NM = 5 // inert under AOBLMOA
			config.EnableParallel = parallel

			if got := effectiveNM(config); got != 0 {
				t.Errorf("effectiveNM = %d under AOBLMOA, want 0: the paper has no mutation stage", got)
			}

			result, err := Optimize(config)
			if err != nil {
				t.Fatalf("Optimize: %v", err)
			}

			// 2*population initial evaluations, then per iteration
			// 2*population moved individuals + offspring crossover children
			// + offspring opposition points.
			want := 2*population + iterations*(2*population+2*offspring)
			if result.FuncEvalCount != want {
				t.Errorf("FuncEvalCount = %d, want %d", result.FuncEvalCount, want)
			}

			if got := int(calls.Load()); got != want {
				t.Errorf("the objective was called %d times, want %d", got, want)
			}
		})
	}
}

// TestAOBLMOAHonoursStrategySwitch proves the knob is live. Through v0.5.1 it
// was declared, defaulted, documented and tested, but never read, so these two
// runs were identical.
func TestAOBLMOAHonoursStrategySwitch(t *testing.T) {
	config := NewAOBLMOAConfig()
	config.MaxIterations = 30
	config.StrategySwitch = 0
	if got := effectiveStrategySwitch(config); got != 20 {
		t.Fatalf("default strategy switch = %d, want 20", got)
	}
	if !aquilaExplorationPhase(10, effectiveStrategySwitch(config)) {
		t.Fatal("default switch leaves exploration too early")
	}
	config.StrategySwitch = 1
	if aquilaExplorationPhase(10, effectiveStrategySwitch(config)) {
		t.Fatal("configured switch is ignored")
	}
}

// TestAOBLMOAAquilaWeightOverrideChangesTheBranch pins the deprecated escape
// hatch: the sentinel takes the paper's deterministic branch, and any
// probability restores the old draw.
func TestAOBLMOAAquilaWeightOverrideChangesTheBranch(t *testing.T) {
	run := func(weight float64) [][]float64 {
		t.Helper()

		config := NewAOBLMOAConfig()
		config.Rand = rand.New(rand.NewSource(808))
		config.ObjectiveFunc = Sphere
		config.ProblemSize = 4
		config.LowerBound = -5
		config.UpperBound = 5
		config.MaxIterations = 100
		config.AquilaWeight = weight
		males, females := aoblmoaTestPopulations(t, config, 8)
		best := Best{Position: append([]float64(nil), males[0].Position...), Cost: males[0].Cost}
		applyAOBLMOAToPopulation(males, females, best, 10, config.MaxIterations,
			config.G, config.Dance, config.FL, config)
		return append(snapshotPositions(males), snapshotPositions(females)...)
	}

	equal := func(left, right [][]float64) bool {
		if len(left) != len(right) {
			return false
		}
		for i := range left {
			if !samePosition(left[i], right[i]) {
				return false
			}
		}
		return true
	}
	paper := run(AquilaWeightAuto)
	if equal(paper, run(0)) {
		t.Error("AquilaWeight = 0 matches the paper default; the override is not read")
	}

	if equal(paper, run(1)) {
		t.Error("AquilaWeight = 1 matches the paper default; the override is not read")
	}

	if equal(run(0), run(1)) {
		t.Error("AquilaWeight = 0 and 1 agree; the override does not select a branch")
	}
}

// TestAquilaWeightAutoSurvivesAConfigRoundTrip pins the sentinel through the
// JSON path, where an out-of-range value used to be rejected outright.
func TestAquilaWeightAutoSurvivesAConfigRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aoblmoa.json")

	config := NewAOBLMOAConfig()
	config.ObjectiveFunc = Sphere
	config.ProblemSize = 3
	config.LowerBound = -5
	config.UpperBound = 5

	saveErr := SaveConfigToFile(config, path)
	if saveErr != nil {
		t.Fatalf("SaveConfigToFile: %v", saveErr)
	}

	raw, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("ReadFile: %v", readErr)
	}

	var exported map[string]any

	unmarshalErr := json.Unmarshal(raw, &exported)
	if unmarshalErr != nil {
		t.Fatalf("Unmarshal: %v", unmarshalErr)
	}

	if exported["aquila_weight"] != AquilaWeightAuto {
		t.Errorf("aquila_weight exported as %v, want %v", exported["aquila_weight"], AquilaWeightAuto)
	}

	loaded, loadErr := LoadConfigFromFile(path)
	if loadErr != nil {
		t.Fatalf("LoadConfigFromFile rejected the sentinel: %v", loadErr)
	}

	if loaded.AquilaWeight != AquilaWeightAuto {
		t.Errorf("loaded AquilaWeight = %v, want the sentinel %v",
			loaded.AquilaWeight, AquilaWeightAuto)
	}
}

// TestAOBLMOAComparedWithStandardMayfly is advisory, not a gate.
//
// The paper claims a clear win over the plain Mayfly Algorithm. A loss is real
// evidence about the open question in aoblmoaStrategyFor -- whether the sexes
// are assigned the strategies the equations give them or the ones the abstract
// gives them -- and is the only empirical handle on it, so the numbers are
// logged rather than asserted.
func TestAOBLMOAComparedWithStandardMayfly(t *testing.T) {
	if testing.Short() {
		t.Skip("advisory comparison; run without -short")
	}

	const seeds = 15

	for _, problem := range []struct {
		objective    ObjectiveFunction
		name         string
		lower, upper float64
	}{
		{name: "Sphere", objective: Sphere, lower: -100, upper: 100},
		{name: "Rastrigin", objective: Rastrigin, lower: -5.12, upper: 5.12},
	} {
		t.Run(problem.name, func(t *testing.T) {
			run := func(newConfig func() *Config, seed int64) float64 {
				config := newConfig()
				config.Rand = rand.New(rand.NewSource(seed))
				config.ObjectiveFunc = problem.objective
				config.ProblemSize = 8
				config.LowerBound = problem.lower
				config.UpperBound = problem.upper
				config.MaxIterations = 300
				config.NPop = 20
				config.NPopF = 20

				result, err := Optimize(config)
				if err != nil {
					t.Fatalf("Optimize: %v", err)
				}

				return result.GlobalBest.Cost
			}

			standard := make([]float64, 0, seeds)
			aoblmoa := make([]float64, 0, seeds)

			for seed := range int64(seeds) {
				standard = append(standard, run(NewDefaultConfig, 1000+seed))
				aoblmoa = append(aoblmoa, run(NewAOBLMOAConfig, 1000+seed))
			}

			sort.Float64s(standard)
			sort.Float64s(aoblmoa)

			standardMedian := standard[seeds/2]
			aoblmoaMedian := aoblmoa[seeds/2]

			verdict := "AOBLMOA wins"
			if aoblmoaMedian > standardMedian {
				verdict = "AOBLMOA LOSES"
			}

			t.Logf("%s, D=8, 300 iterations, %d seeds: standard median %.6g, "+
				"AOBLMOA median %.6g -- %s",
				problem.name, seeds, standardMedian, aoblmoaMedian, verdict)
			t.Logf("%s best/worst: standard %.6g / %.6g, AOBLMOA %.6g / %.6g",
				problem.name, standard[0], standard[seeds-1], aoblmoa[0], aoblmoa[seeds-1])
		})
	}
}

// TestAOBLMOAParallelUpdatePhaseMatchesSequential is a direct, single-iteration
// check of the property TestAOBLMOASequentialAndParallelAgree covers over a
// whole run: both paths now call the same move function, so the swarms must be
// bit-identical after one update phase.
func TestAOBLMOAParallelUpdatePhaseMatchesSequential(t *testing.T) {
	const size = 8

	build := func(t *testing.T) (*Config, []*Mayfly, []*Mayfly, Best) {
		t.Helper()

		config := aoblmoaBranchConfig(909)
		males, females := aoblmoaTestPopulations(t, config, size)
		globalBest := Best{
			Position: append([]float64(nil), males[0].Best.Position...),
			Cost:     males[0].Cost,
		}

		return config, males, females, globalBest
	}

	sequentialConfig, sequentialMales, sequentialFemales, sequentialBest := build(t)
	sequentialEvaluations := applyAOBLMOAToPopulation(
		sequentialMales, sequentialFemales, sequentialBest, 10, sequentialConfig.MaxIterations,
		sequentialConfig.G, sequentialConfig.Dance, sequentialConfig.FL, sequentialConfig,
	)

	parallelConfig, parallelMales, parallelFemales, parallelBest := build(t)
	pool := newConstrainedEvaluationPool(
		newConstraintEvaluator(parallelConfig.ObjectiveFunc, nil), 4,
	)

	defer pool.close()

	parallelEvaluations, err := evaluateParallelAOBLMOA(
		context.Background(), parallelMales, parallelFemales, &parallelBest,
		10, parallelConfig.MaxIterations,
		parallelConfig.G, parallelConfig.Dance, parallelConfig.FL,
		parallelConfig, parallelConfig.Rand, pool,
	)
	if err != nil {
		t.Fatalf("evaluateParallelAOBLMOA: %v", err)
	}

	if sequentialEvaluations != parallelEvaluations {
		t.Errorf("evaluation budget differs: sequential %d, parallel %d",
			sequentialEvaluations, parallelEvaluations)
	}

	for i := range size {
		if !samePosition(sequentialMales[i].Position, parallelMales[i].Position) {
			t.Errorf("male %d differs: sequential %v, parallel %v",
				i, sequentialMales[i].Position, parallelMales[i].Position)
		}

		if !samePosition(sequentialFemales[i].Position, parallelFemales[i].Position) {
			t.Errorf("female %d differs: sequential %v, parallel %v",
				i, sequentialFemales[i].Position, parallelFemales[i].Position)
		}
	}
}
