package mayfly

import (
	"math"
	"math/rand"
)

// hmmaScheduleProbability implements Eq. (10) of Zhang et al. (2022):
//
//	Ps = -exp(-t/Iter_MAX) + theta
//
// iteration is one-based. Clamping is a numerical guard for caller-selected
// theta values at the endpoints; the paper's theta=0.99 naturally starts at
// zero and rises toward about 0.622.
func hmmaScheduleProbability(iteration, maxIterations int, theta float64) float64 {
	probability := -math.Exp(-float64(iteration)/float64(maxIterations)) + theta

	return min(max(probability, 0), 1)
}

// hmmaOppositionTarget implements Eqs. (6)-(7). The paper specifies r3 as a
// uniform random matrix, so each coordinate receives an independent draw.
func hmmaOppositionTarget(
	globalBest []float64,
	lowerBound, upperBound, informationExchange float64,
	rng *rand.Rand,
) []float64 {
	target := make([]float64, len(globalBest))
	for i, coordinate := range globalBest {
		opposition := upperBound + rng.Float64()*(lowerBound-coordinate)
		target[i] = informationExchange * (coordinate - opposition)
	}
	maxVec(target, lowerBound)
	minVec(target, upperBound)

	return target
}

// hmmaCauchyTarget implements Eq. (8), multiplying each global-best
// coordinate by an independent standard Cauchy variate.
func hmmaCauchyTarget(
	globalBest []float64,
	lowerBound, upperBound float64,
	rng *rand.Rand,
) []float64 {
	target := make([]float64, len(globalBest))
	for i, coordinate := range globalBest {
		target[i] = cauchyRand(0, 1, rng) * coordinate
	}
	maxVec(target, lowerBound)
	minVec(target, upperBound)

	return target
}

// hmmaGlobalMutation implements Eqs. (9)-(11): choose the opposition or
// Cauchy target with the scheduled probability, evaluate it once, and retain
// it only when it improves the global optimum.
func hmmaGlobalMutation(
	globalBest Best,
	iteration, maxIterations int,
	config *Config,
	evaluator *constraintEvaluator,
	rng *rand.Rand,
) Best {
	probability := hmmaScheduleProbability(
		iteration, maxIterations, config.HMMAScheduleOffset,
	)

	var position []float64
	if probability > rng.Float64() {
		position = hmmaOppositionTarget(
			globalBest.Position,
			config.LowerBound,
			config.UpperBound,
			config.HMMAInformationExchange,
			rng,
		)
	} else {
		position = hmmaCauchyTarget(
			globalBest.Position, config.LowerBound, config.UpperBound, rng,
		)
	}

	candidate := newMayfly(len(position))
	copy(candidate.Position, position)
	evaluator.evaluateMayfly(candidate, false)
	if evaluator.betterMayflyThanBest(candidate, globalBest) {
		return bestFromMayfly(candidate)
	}

	return globalBest
}

// hmmaArtificialMutation implements Eq. (12). Both outputs are calculated
// from the original pair, not from an already-mutated sibling.
func hmmaArtificialMutation(male, female []float64, rho float64) ([]float64, []float64) {
	maleMutated := make([]float64, len(male))
	femaleMutated := make([]float64, len(female))
	for i := range male {
		maleMutated[i] = (1-rho)*male[i] + rho*female[i]
		femaleMutated[i] = (1-rho)*female[i] + rho*male[i]
	}

	return maleMutated, femaleMutated
}
