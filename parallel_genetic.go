package mayfly

import (
	"context"
	"math"
	"math/rand"
)

// evaluateParallelGeneticOperators prepares genetic candidates on the caller
// goroutine, then evaluates the fixed crossover and mutation batches through
// the run-scoped worker pool. Keeping random generation and slice growth here
// avoids sharing rand.Rand or a slice header between goroutines.
func evaluateParallelGeneticOperators(
	ctx context.Context,
	males, females []*Mayfly,
	config *Config,
	rng *rand.Rand,
	evaluator *evaluationPool,
	iteration int,
	chaosMaps ...*LogisticMap,
) ([]*Mayfly, Best, int, error) {
	geneticBest := Best{Cost: math.Inf(1), ConstraintViolation: math.Inf(1)}
	nc := effectiveNC(config)
	maleOffspring := make([]*Mayfly, 0, nc/2+effectiveNM(config))
	femaleOffspring := make([]*Mayfly, 0, nc/2+effectiveNM(config))

	for k := range nc / 2 {
		contextErr := ctx.Err()
		if contextErr != nil {
			return nil, Best{}, 0, contextErr
		}

		// Optimize validates that both populations contain every requested
		// parent pair before this internal helper is called.
		male, female := selectParents(males, females, k, config, rng)
		off1Pos, off2Pos := crossoverForConfig(
			male.Position,
			female.Position,
			config,
			rng,
		)

		off1 := newMayfly(config.ProblemSize)
		copy(off1.Position, off1Pos)

		off2 := newMayfly(config.ProblemSize)
		copy(off2.Position, off2Pos)

		if config.UseHMMA {
			off1.Position, off2.Position = hmmaArtificialMutation(
				off1.Position, off2.Position, config.HMMAArtificialMutation,
			)
		}

		maleOffspring = append(maleOffspring, off1)
		femaleOffspring = append(femaleOffspring, off2)
	}

	offspring := append(append(make([]*Mayfly, 0, nc), maleOffspring...), femaleOffspring...)

	crossoverBest, err := evaluator.evaluate(ctx, offspring, false, true)
	if err != nil {
		return nil, Best{}, 0, err
	}

	evaluations := len(offspring)

	initializeOffspringBests(offspring)

	// The historical OLCE compatibility stage forms one new position from the
	// fittest crossover offspring. Publisher pseudocode discovered later uses
	// Chebyshev mutation over the whole offspring batch; see chaos.go.
	if config.UseOLCE && config.ChaosFactor > 0 && len(offspring) > 0 {
		var chaosMap *LogisticMap
		if len(chaosMaps) > 0 {
			chaosMap = chaosMaps[0]
		}

		if chaosMap == nil {
			chaosMap = NewLogisticMap(rng.Float64())
		}

		target := fittestMayfly(offspring, evaluator.evaluator)
		candidate := newMayfly(config.ProblemSize)
		chaoticExploitationCandidate(
			candidate.Position, target.Position, config, chaosMap,
			chaoticConstrictionFactor(config, iteration),
		)

		if _, evaluationErr := evaluator.evaluate(ctx, []*Mayfly{candidate}, false, false); evaluationErr != nil {
			return nil, Best{}, 0, evaluationErr
		}

		commitChaoticOffspring(target, candidate)

		if evaluator.evaluator.betterMayflyThanBest(target, crossoverBest) {
			copyMayflyToBest(&crossoverBest, target)
		}

		evaluations++
	}

	if config.UseAOBLMOA {
		// AOBLMOA replaces offspring mutation with stochastic
		// opposition-based learning on every offspring; see
		// prepareStochasticOBL. The crossover best is deliberately not merged
		// above: Eq. (32) always keeps the better of each pair, so the global
		// best is taken from the kept offspring once the selection is done.
		oblEvaluations, oblErr := applyParallelStochasticOBL(ctx, offspring, config, rng, evaluator)
		if oblErr != nil {
			return nil, Best{}, 0, oblErr
		}

		return offspring, keptOffspringBest(offspring, evaluator), evaluations + oblEvaluations, nil
	}

	if evaluator.evaluator.betterBest(crossoverBest, geneticBest) {
		geneticBest = crossoverBest
	}

	for range effectiveNM(config) {
		contextErr := ctx.Err()
		if contextErr != nil {
			return nil, Best{}, 0, contextErr
		}

		maleParent := males[rng.Intn(len(males))]
		femaleParent := females[rng.Intn(len(females))]
		maleOffspring = append(maleOffspring,
			prepareGeneticMutant(maleParent, iteration, config, rng))
		femaleOffspring = append(femaleOffspring,
			prepareGeneticMutant(femaleParent, iteration, config, rng))
	}

	mutants := make([]*Mayfly, 0, 2*effectiveNM(config))
	mutants = append(mutants, maleOffspring[len(maleOffspring)-effectiveNM(config):]...)
	mutants = append(mutants, femaleOffspring[len(femaleOffspring)-effectiveNM(config):]...)

	mutationBest, err := evaluator.evaluate(ctx, mutants, false, true)
	if err != nil {
		return nil, Best{}, 0, err
	}

	initializeOffspringBests(mutants)

	if evaluator.evaluator.betterBest(mutationBest, geneticBest) {
		geneticBest = mutationBest
	}

	offspring = make([]*Mayfly, 0, len(maleOffspring)+len(femaleOffspring))
	offspring = append(offspring, maleOffspring...)
	offspring = append(offspring, femaleOffspring...)

	return offspring, geneticBest, evaluations + len(mutants), nil
}

func prepareGeneticMutant(
	parent *Mayfly,
	iteration int,
	config *Config,
	rng *rand.Rand,
) *Mayfly {
	mutant := newMayfly(config.ProblemSize)
	mutant.Position = Mutate(
		parent.Position,
		config.Mu,
		config.LowerBound,
		config.UpperBound,
		rng,
	)

	return mutant
}

// applyParallelStochasticOBL runs the AOBLMOA opposition stage over the worker
// pool. The candidates are built on the caller goroutine by the very function
// the sequential path calls, so both paths consume the same random numbers in
// the same order.
func applyParallelStochasticOBL(
	ctx context.Context,
	offspring []*Mayfly,
	config *Config,
	rng *rand.Rand,
	evaluator *evaluationPool,
) (int, error) {
	opposites := prepareStochasticOBL(offspring, config, rng)

	_, err := evaluator.evaluate(ctx, opposites, false, false)
	if err != nil {
		return 0, err
	}

	commitStochasticOBL(offspring, opposites, evaluator.evaluator)

	return len(opposites), nil
}

// keptOffspringBest reports the best of the offspring that survived greedy
// opposition selection.
func keptOffspringBest(offspring []*Mayfly, evaluator *evaluationPool) Best {
	best := Best{Cost: math.Inf(1), ConstraintViolation: math.Inf(1)}

	for _, child := range offspring {
		if !evaluator.evaluator.betterMayflyThanBest(child, best) {
			continue
		}

		best = cloneBest(Best{
			Position:            child.Position,
			Cost:                child.Cost,
			ConstraintViolation: child.ConstraintViolation,
		})
	}

	return best
}

func initializeOffspringBests(offspring []*Mayfly) {
	for _, mayfly := range offspring {
		copy(mayfly.Best.Position, mayfly.Position)
		mayfly.Best.Cost = mayfly.Cost
		mayfly.Best.ConstraintViolation = mayfly.ConstraintViolation
	}
}
