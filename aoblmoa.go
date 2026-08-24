package mayfly

import (
	"errors"
	"fmt"
	"math/rand"
)

// AOBLMOA - Aquila Optimizer and Opposition-Based Learning Mayfly Optimization
// Algorithm.
//
// This file implements the AOBLMOA variant, which hybridizes the Mayfly
// Algorithm with the Aquila Optimizer and replaces offspring mutation with
// stochastic opposition-based learning.
//
// Structure of one iteration, following the paper:
//
//  1. Update phase. Every individual either keeps the Mayfly attraction term
//     or takes an Aquila step. The choice is a deterministic fitness test, not
//     a probability: a male is attracted while the global best dominates him
//     and hunts otherwise; a female is attracted while her paired male
//     dominates her and hunts otherwise. The Aquila strategy is fixed by the
//     individual's sex and by the iteration phase, so no coin is flipped
//     inside a phase.
//  2. Crossover, unchanged.
//  3. Stochastic opposition-based learning on every offspring, Eq. (31),
//     followed by greedy selection, Eq. (32). This replaces Gaussian mutation,
//     which the paper does not have.
//
// Reference:
// Zhao, Y.; Huang, C.; Zhang, M.; Cui, Y. AOBLMOA: A Hybrid Biomimetic
// Optimization Algorithm for Numerical Optimization and Engineering Design
// Problems. Biomimetics 2023, 8(4), 381. DOI: 10.3390/biomimetics8040381.

// aoblmoaMaleTakesAttraction reports whether a male keeps the ordinary Mayfly
// attraction term this iteration instead of taking an Aquila step.
//
// Eq. (29): the male is attracted while the global best dominates him, and
// hunts otherwise. This is exactly the test prepareStandardMale applies before
// its nuptial dance, which is the branch AOBLMOA replaces.
func aoblmoaMaleTakesAttraction(male *Mayfly, globalBest Best, evaluator *constraintEvaluator) bool {
	return evaluator.better(evaluationFromBest(globalBest), evaluationFromMayfly(male))
}

// aoblmoaFemaleTakesAttraction reports whether a female keeps the ordinary
// Mayfly attraction term this iteration instead of taking an Aquila step.
//
// OPEN PAPER QUESTION (resolved, but deliberately one line to flip).
// The paper contradicts itself: Eq. (30) attracts the female while her paired
// male dominates her, whereas the Algorithm 1 pseudocode states the opposite
// inequality. Resolved in favor of Eq. (30), because AOBLMOA replaces
// branches of the plain Mayfly Algorithm and prepareStandardFemale attracts on
// exactly this test.
//
// To adopt the pseudocode reading instead, invert this one return and change
// nothing else.
func aoblmoaFemaleTakesAttraction(female, male *Mayfly, evaluator *constraintEvaluator) bool {
	return evaluator.betterMayfly(male, female)
}

// aoblmoaStrategyFor maps an individual to its Aquila hunting strategy.
//
// OPEN PAPER QUESTION.
// The paper contradicts itself about which sex gets which pair of strategies.
// Its equations give males the narrowed strategies (X2 contour flight, X4 walk
// and grab) and females the expanded ones (X1 high soar, X3 low flight); its
// abstract swaps them. Both readings are internally coherent. The equations
// win here, because a formal specification outranks prose and its terms can be
// checked one by one against aquila.go.
//
// To adopt the abstract's reading instead, swap the two returned pairs and
// change nothing else.
//
// Note that AOBLMOA, unlike plain AO, does not flip a coin within a phase: the
// sex fixes the strategy, so this decision consumes no randomness.
func aoblmoaStrategyFor(isMale, exploration bool) AquilaStrategy {
	if isMale {
		if exploration {
			return NarrowedExploration
		}

		return NarrowedExploitation
	}

	if exploration {
		return ExpandedExploration
	}

	return ExpandedExploitation
}

// snapshotPopulation freezes the position and fitness of a population at the
// start of an AOBLMOA iteration.
//
// Both AOBLMOA paths update the whole swarm against this frozen state: the
// Aquila strategies read population positions (the mean position and a random
// peer), and every female is paired against a male. Reading the live slices
// instead makes the outcome depend on how far the update loop has already
// progressed, which is where the sequential and the parallel path used to
// disagree — the sequential path moved and re-evaluated every male before the
// females ran, while the parallel path deferred the Aquila moves and the
// evaluations to the end of the iteration.
//
// Freezing also matches the plain standard variant, which updates the females
// against the males' previous-iteration position and cost.
func snapshotPopulation(population []*Mayfly) []*Mayfly {
	snapshot := make([]*Mayfly, len(population))

	for i, member := range population {
		frozen := newMayfly(len(member.Position))
		copy(frozen.Position, member.Position)
		frozen.Cost = member.Cost
		frozen.ConstraintViolation = member.ConstraintViolation
		snapshot[i] = frozen
	}

	return snapshot
}

// applyAOBLMOAToPopulation applies the AOBLMOA variant logic to a whole
// population, using a freshly built evaluator.
func applyAOBLMOAToPopulation(males, females []*Mayfly, globalBest Best,
	currentIter, maxIter int, g, dance, flight float64, config *Config,
) int {
	return applyAOBLMOAToPopulationWithEvaluator(
		males, females, globalBest, currentIter, maxIter, g, dance, flight, config,
		newConstraintEvaluator(config.ObjectiveFunc, nil),
	)
}

// applyAOBLMOAMoves moves every male and every female exactly once and
// evaluates nothing.
//
// It is the single implementation of the AOBLMOA update phase. Both the
// sequential and the parallel path call it and then evaluate the swarm their
// own way, which is what makes the two paths bit-identical by construction
// rather than by two hand-synchronized copies of the same branch cascade --
// the arrangement that let them drift apart twice before.
//
// Every branch test reads the frozen snapshots, so the outcome cannot depend
// on how far the loop has already progressed.
//
// dance and flight are consumed only by the deprecated AquilaWeight override.
// The published algorithm replaces the nuptial-dance and random-flight
// branches wholesale, so under the default neither coefficient can reach the
// update.
//
// Note that the Aquila branch deliberately leaves the velocity untouched.
// Branches two and three of Eq. (29) and Eq. (30) are position formulas, not
// velocity formulas: the individual is teleported by the hunting strategy and
// keeps the velocity it carried into the iteration.
func applyAOBLMOAMoves(
	males, females []*Mayfly,
	globalBest Best,
	currentIter, maxIter int,
	g, dance, flight float64,
	config *Config,
	rng *rand.Rand,
	evaluator *constraintEvaluator,
) {
	// Every individual is updated against the state the swarm had when the
	// iteration began, so the result does not depend on the update order.
	frozenMales := snapshotPopulation(males)
	frozenFemales := snapshotPopulation(females)

	exploration := aquilaExplorationPhase(currentIter, effectiveStrategySwitch(config))
	weight, overridden := effectiveAquilaWeight(config)

	for i, male := range males {
		// The deprecated AquilaWeight override draws once per individual,
		// before the branch and in the order the pre-v0.6.0 code drew, and
		// falls back to the complete standard update -- attraction or nuptial
		// dance -- so a caller who sets it gets the old behavior back rather
		// than an approximation of it. dance and flight reach nothing else.
		if overridden {
			if rng.Float64() >= weight {
				prepareStandardMale(male, globalBest, nil, g, dance, g, config, rng, evaluator)

				continue
			}

			aoblmoaAquilaStep(male, globalBest, frozenMales, true, exploration,
				currentIter, maxIter, config, rng)

			continue
		}

		if aoblmoaMaleTakesAttraction(frozenMales[i], globalBest, evaluator) {
			prepareAttractedMale(male, globalBest, nil, g, g, config)
			applyVelocityAndMove(male, config)

			continue
		}

		aoblmoaAquilaStep(male, globalBest, frozenMales, true, exploration,
			currentIter, maxIter, config, rng)
	}

	for i, female := range females {
		if overridden {
			if rng.Float64() >= weight {
				prepareStandardFemale(female, frozenMales[i], g, flight, config, rng, evaluator)

				continue
			}

			aoblmoaAquilaStep(female, globalBest, frozenFemales, false, exploration,
				currentIter, maxIter, config, rng)

			continue
		}

		if aoblmoaFemaleTakesAttraction(frozenFemales[i], frozenMales[i], evaluator) {
			prepareAttractedFemale(female, frozenMales[i], g, config)
			applyVelocityAndMove(female, config)

			continue
		}

		aoblmoaAquilaStep(female, globalBest, frozenFemales, false, exploration,
			currentIter, maxIter, config, rng)
	}
}

// aoblmoaAquilaStep replaces an individual's position with the one its Aquila
// hunting strategy prescribes, clamped to the search bounds. The velocity is
// left as it was; see applyAOBLMOAMoves.
func aoblmoaAquilaStep(
	target *Mayfly,
	globalBest Best,
	frozen []*Mayfly,
	isMale, exploration bool,
	currentIter, maxIter int,
	config *Config,
	rng *rand.Rand,
) {
	position := applyAquilaStrategy(
		target, globalBest, frozen, aoblmoaStrategyFor(isMale, exploration),
		currentIter, maxIter, config, rng,
	)

	copy(target.Position, position)
	maxVec(target.Position, config.LowerBound)
	minVec(target.Position, config.UpperBound)
}

// applyAOBLMOAToPopulationWithEvaluator runs one sequential AOBLMOA update
// phase: it moves every male and female exactly once, evaluates each of them
// once and refreshes the males' personal bests.
//
// It returns the number of objective evaluations consumed, which is always the
// population size: opposition-based learning has moved to the offspring stage,
// where the paper puts it, so the update phase spends nothing extra.
func applyAOBLMOAToPopulationWithEvaluator(
	males, females []*Mayfly,
	globalBest Best,
	currentIter, maxIter int,
	g, dance, flight float64,
	config *Config,
	evaluator *constraintEvaluator,
) int {
	applyAOBLMOAMoves(males, females, globalBest, currentIter, maxIter, g, dance, flight,
		config, config.Rand, evaluator)

	for _, male := range males {
		evaluator.evaluateMayfly(male, false)
	}

	for _, female := range females {
		evaluator.evaluateMayfly(female, false)
	}

	updatePersonalBests(males, evaluator)

	return len(males) + len(females)
}

// prepareStochasticOBL builds the stochastic opposition point of every
// offspring, Eq. (31). It draws every random number the stage needs and
// evaluates nothing, so the sequential and the parallel path can share it
// verbatim and differ only in how they evaluate the result.
func prepareStochasticOBL(offspring []*Mayfly, config *Config, rng *rand.Rand) []*Mayfly {
	opposites := make([]*Mayfly, len(offspring))

	for i, child := range offspring {
		opposite := newMayfly(len(child.Position))
		opposite.Position = stochasticOppositionPoint(
			child.Position, config.LowerBound, config.UpperBound, rng,
		)
		opposites[i] = opposite
	}

	return opposites
}

// commitStochasticOBL applies the greedy selection of Eq. (32): each offspring
// is replaced by its opposition point when that point is the better of the
// two, and keeps its own position otherwise.
//
// Both slices must already be evaluated.
func commitStochasticOBL(offspring, opposites []*Mayfly, evaluator *constraintEvaluator) {
	for i, child := range offspring {
		opposite := opposites[i]
		if !evaluator.betterMayfly(opposite, child) {
			continue
		}

		copy(child.Position, opposite.Position)
		child.Cost = opposite.Cost
		child.ConstraintViolation = opposite.ConstraintViolation

		copy(child.Best.Position, child.Position)
		child.Best.Cost = child.Cost
		child.Best.ConstraintViolation = child.ConstraintViolation
	}
}

// ParetoArchive maintains a set of non-dominated solutions for multi-objective problems.
type ParetoArchive struct {
	initErr            error
	Solutions          []*ParetoSolution
	solutions          []*ParetoSolution
	MaxSize            int
	maxSize            int
	objectiveDimension int
}

// NewParetoArchive creates a new Pareto archive with specified maximum size.
func NewParetoArchive(maxSize int) *ParetoArchive {
	archive := &ParetoArchive{MaxSize: maxSize, maxSize: maxSize}
	if maxSize <= 0 {
		archive.initErr = fmt.Errorf("Pareto archive capacity must be positive, got %d", maxSize)
		archive.maxSize = 0

		return archive
	}

	archive.solutions = make([]*ParetoSolution, 0, maxSize)
	archive.syncSnapshot()

	return archive
}

// NewParetoArchiveWithObjectives creates an archive with an explicit objective
// dimension. Use it when the dimension is known before the first insertion.
func NewParetoArchiveWithObjectives(maxSize, objectiveDimension int) (*ParetoArchive, error) {
	archive := NewParetoArchive(maxSize)
	if archive.initErr != nil {
		return nil, archive.initErr
	}

	if objectiveDimension <= 0 {
		return nil, fmt.Errorf("objective dimension must be positive, got %d", objectiveDimension)
	}

	archive.objectiveDimension = objectiveDimension

	return archive, nil
}

// Add adds a solution to the Pareto archive.
// If the archive is full, it uses NSGA-II selection to maintain diversity.
func (pa *ParetoArchive) Add(solution *ParetoSolution) (bool, error) {
	if pa == nil {
		return false, errors.New("Pareto archive is nil")
	}

	if pa.initErr != nil {
		return false, pa.initErr
	}

	if solution == nil {
		return false, errors.New("Pareto solution is nil")
	}

	dimension := pa.objectiveDimension
	if dimension == 0 {
		dimension = len(solution.ObjectiveValues)
	}

	err := validateObjectiveVector(solution.ObjectiveValues, dimension)
	if err != nil {
		return false, err
	}

	for i, coordinate := range solution.Position {
		if !isFinite(coordinate) {
			return false, fmt.Errorf("position %d is not finite", i)
		}
	}

	for _, incumbent := range pa.solutions {
		if objectiveVectorsEqual(incumbent.ObjectiveValues, solution.ObjectiveValues) {
			return false, nil
		}

		if dominates(incumbent.ObjectiveValues, solution.ObjectiveValues) {
			return false, nil
		}
	}

	kept := make([]*ParetoSolution, 0, len(pa.solutions)+1)
	for _, incumbent := range pa.solutions {
		if !dominates(solution.ObjectiveValues, incumbent.ObjectiveValues) {
			kept = append(kept, incumbent)
		}
	}

	kept = append(kept, cloneParetoSolution(solution))
	if len(kept) > pa.maxSize {
		kept = selectByNSGA2(kept, pa.maxSize)
	}

	pa.solutions = kept
	pa.objectiveDimension = dimension
	pa.syncSnapshot()

	return true, nil
}

// AddFromMayfly converts a Mayfly to a ParetoSolution and adds it to the archive.
// For single-objective problems, the objective value is just the cost.
func (pa *ParetoArchive) AddFromMayfly(mayfly *Mayfly) (bool, error) {
	if mayfly == nil {
		return false, errors.New("mayfly is nil")
	}

	solution := &ParetoSolution{
		Position:         make([]float64, len(mayfly.Position)),
		ObjectiveValues:  []float64{mayfly.Cost},
		Rank:             0,
		CrowdingDistance: 0,
	}
	copy(solution.Position, mayfly.Position)

	return pa.Add(solution)
}

// GetBestSolution returns the solution with the lowest first objective value.
// This is useful for single-objective optimization.
func (pa *ParetoArchive) GetBestSolution() *ParetoSolution {
	if pa == nil || len(pa.solutions) == 0 {
		return nil
	}

	best := pa.solutions[0]
	for _, sol := range pa.solutions[1:] {
		if sol.ObjectiveValues[0] < best.ObjectiveValues[0] {
			best = sol
		}
	}

	return cloneParetoSolution(best)
}

// GetSolutions returns a defensive snapshot of the archive.
func (pa *ParetoArchive) GetSolutions() []*ParetoSolution {
	if pa == nil {
		return nil
	}

	return cloneParetoSolutions(pa.solutions)
}

// Capacity returns the validated maximum number of retained solutions.
func (pa *ParetoArchive) Capacity() int {
	if pa == nil {
		return 0
	}

	return pa.maxSize
}

// UpdateFromPopulation adds every member of the given populations to the
// archive.
//
// The optimizer does not call this itself. Nothing in the search reads the
// archive back, so maintaining it every iteration only bought NSGA-II pruning
// cost; callers that want a Pareto front build and feed one explicitly.
func (pa *ParetoArchive) UpdateFromPopulation(males, females []*Mayfly) error {
	for _, m := range males {
		if _, err := pa.AddFromMayfly(m); err != nil {
			return err
		}
	}

	for _, f := range females {
		if _, err := pa.AddFromMayfly(f); err != nil {
			return err
		}
	}

	return nil
}

func (pa *ParetoArchive) syncSnapshot() {
	pa.Solutions = cloneParetoSolutions(pa.solutions)
}

func objectiveVectorsEqual(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
