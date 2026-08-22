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
) ([]*Mayfly, Best, error) {
	geneticBest := Best{Cost: math.Inf(1), ConstraintViolation: math.Inf(1)}
	nc := effectiveNC(config)
	offspring := make([]*Mayfly, 0, 2*(nc/2)+config.NM)

	for k := range nc / 2 {
		contextErr := ctx.Err()
		if contextErr != nil {
			return nil, Best{}, contextErr
		}

		// Optimize validates that both populations contain every requested
		// parent pair before this internal helper is called.
		male, female := selectParents(males, females, k, config, rng)
		off1Pos, off2Pos := Crossover(
			male.Position,
			female.Position,
			config.LowerBound,
			config.UpperBound,
			rng,
		)

		off1 := newMayfly(config.ProblemSize)
		copy(off1.Position, off1Pos)

		off2 := newMayfly(config.ProblemSize)
		copy(off2.Position, off2Pos)

		offspring = append(offspring, off1, off2)
	}

	crossoverBest, err := evaluator.evaluate(ctx, offspring, false, true)
	if err != nil {
		return nil, Best{}, err
	}

	initializeOffspringBests(offspring)

	if evaluator.evaluator.betterBest(crossoverBest, geneticBest) {
		geneticBest = crossoverBest
	}

	mutationStart := len(offspring)

	for range config.NM {
		contextErr := ctx.Err()
		if contextErr != nil {
			return nil, Best{}, contextErr
		}

		parent := offspring[rng.Intn(len(offspring))]
		mutant := newMayfly(config.ProblemSize)

		if config.UseGSASMA {
			mutant.Position = HybridMutate(
				parent.Position,
				config.Mu,
				config.LowerBound,
				config.UpperBound,
				adaptiveCauchyProbability(iteration, config),
				rng,
			)
		} else {
			mutant.Position = Mutate(
				parent.Position,
				config.Mu,
				config.LowerBound,
				config.UpperBound,
				rng,
			)
		}

		// Append immediately so later mutants retain the existing ability to
		// select an earlier mutant as their parent.
		offspring = append(offspring, mutant)
	}

	mutants := offspring[mutationStart:]

	mutationBest, err := evaluator.evaluate(ctx, mutants, false, true)
	if err != nil {
		return nil, Best{}, err
	}

	initializeOffspringBests(mutants)

	if evaluator.evaluator.betterBest(mutationBest, geneticBest) {
		geneticBest = mutationBest
	}

	return offspring, geneticBest, nil
}

func adaptiveCauchyProbability(iteration int, config *Config) float64 {
	iterRatio := float64(iteration) / float64(config.MaxIterations)

	switch {
	case iterRatio < 0.33:
		return 0.7
	case iterRatio < 0.66:
		return 0.5
	default:
		return config.CauchyMutationRate
	}
}

func initializeOffspringBests(offspring []*Mayfly) {
	for _, mayfly := range offspring {
		copy(mayfly.Best.Position, mayfly.Position)
		mayfly.Best.Cost = mayfly.Cost
		mayfly.Best.ConstraintViolation = mayfly.ConstraintViolation
	}
}
