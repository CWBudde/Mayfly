// DESMA (Dynamic Elite Strategy Mayfly Algorithm).
//
// Implements the DESMA variant with adaptive elite generation around global best.
//
// Reference:
// Du, P., Wang, J., Hao, Y., Niu, T., & Yang, W. (2022). Dynamic elite strategy
// mayfly algorithm. PLOS One, 17(8), e0273155.
// DOI: 10.1371/journal.pone.0273155
// PMC: https://pmc.ncbi.nlm.nih.gov/articles/PMC9409577/
//
// DESMA enhances the standard Mayfly Algorithm with:
// - Elite solution generation within adaptive search range
// - Dynamic range adjustment based on improvement (enlarge if improving, reduce if stagnating)
// - 70%+ improvement on multimodal functions with ~8% overhead

package mayfly

import "math/rand"

// generateEliteMayflies implements the DESMA dynamic elite strategy.
// It generates elite mayflies around the current global best position.
func generateEliteMayflies(currentBest Best, searchRange float64, eliteCount, problemSize int,
	lowerBound, upperBound float64, objFunc ObjectiveFunction, rng *rand.Rand,
) (*Mayfly, int) {
	return generateEliteMayfliesWithEvaluator(
		currentBest, searchRange, eliteCount, problemSize, lowerBound, upperBound,
		newConstraintEvaluator(objFunc, nil), rng,
	)
}

func generateEliteMayfliesWithEvaluator(
	currentBest Best,
	searchRange float64,
	eliteCount, problemSize int,
	lowerBound, upperBound float64,
	evaluator *constraintEvaluator,
	rng *rand.Rand,
) (*Mayfly, int) {
	bestElite, funcEvals, _ := generateImprovedEliteMayfliesWithEvaluator(
		currentBest, searchRange, eliteCount, problemSize, lowerBound, upperBound,
		evaluator, rng,
	)
	if bestElite != nil {
		return bestElite, funcEvals
	}

	return mayflyFromBest(currentBest, problemSize), funcEvals
}

// generateImprovedEliteMayfliesWithEvaluator performs the DESMA elite search
// while explicitly reporting whether it found a strict improvement. This lets
// the optimizer make EliteCount==0 and unsuccessful searches true no-ops
// instead of replacing an unrelated population member with a clone of gbest.
func generateImprovedEliteMayfliesWithEvaluator(
	currentBest Best,
	searchRange float64,
	eliteCount, problemSize int,
	lowerBound, upperBound float64,
	evaluator *constraintEvaluator,
	rng *rand.Rand,
) (*Mayfly, int, bool) {
	if eliteCount <= 0 {
		return nil, 0, false
	}

	bestElite := mayflyFromBest(currentBest, problemSize)
	improved := false
	funcEvals := 0

	// Generate elite mayflies around current best
	for range eliteCount {
		elite := newMayfly(problemSize)

		// Generate elite position: egbest = cgbest + r1 * R
		// where r1 is random vector in [-1, 1]
		for j := range problemSize {
			r1 := unifrnd(-1, 1, rng)
			elite.Position[j] = currentBest.Position[j] + r1*searchRange
		}

		maxVec(elite.Position, lowerBound)
		minVec(elite.Position, upperBound)
		evaluator.evaluateMayfly(elite, false)

		funcEvals++

		if evaluator.betterMayfly(elite, bestElite) {
			bestElite = elite
			copy(bestElite.Best.Position, elite.Position)
			bestElite.Best.Cost = elite.Cost
			bestElite.Best.ConstraintViolation = elite.ConstraintViolation
			improved = true
		}
	}

	if !improved {
		return nil, funcEvals, false
	}

	return bestElite, funcEvals, true
}

func mayflyFromBest(currentBest Best, problemSize int) *Mayfly {
	bestElite := newMayfly(problemSize)
	copy(bestElite.Position, currentBest.Position)
	bestElite.Cost = currentBest.Cost
	bestElite.ConstraintViolation = currentBest.ConstraintViolation
	copy(bestElite.Best.Position, currentBest.Position)
	bestElite.Best.Cost = currentBest.Cost
	bestElite.Best.ConstraintViolation = currentBest.ConstraintViolation

	return bestElite
}
