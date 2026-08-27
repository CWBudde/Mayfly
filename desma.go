// DESMA (Dynamic Elite Strategy Mayfly Algorithm).
//
// Implements the DESMA variant with adaptive elite generation around global best.
//
// Reference:
// Du, Q., & Zhu, H. (2022). Dynamic elite strategy mayfly algorithm.
// PLOS ONE, 17(8), e0273155.
// DOI: 10.1371/journal.pone.0273155
// PMC: https://pmc.ncbi.nlm.nih.gov/articles/PMC9409577/
//
// DESMA enhances the standard Mayfly Algorithm with:
// - Elite solution generation within adaptive search range
// - Dynamic range adjustment based on improvement (enlarge if improving, reduce if stagnating)
//
// The paper does not report the initial search radius. The library's automatic
// 10%-of-span radius remains a compatibility choice, not part of a paper-exact
// preset. See docs/reference-data/desma-2022-table3.json.

package mayfly

import "math/rand"

// desmaCrossover implements DESMA Eqs. (6)-(7). Each coordinate draws its
// coefficient from U[-1,1], and the same coefficient is shared by the two
// complementary siblings. The per-coordinate draw follows the cited original
// MA authors' crossover implementation; the DESMA paper itself specifies the
// interval but does not publish code or state the draw granularity.
func desmaCrossover(
	male, female []float64,
	lowerBound, upperBound float64,
	rng *rand.Rand,
) ([]float64, []float64) {
	offspring1 := make([]float64, len(male))
	offspring2 := make([]float64, len(male))

	for dimension := range male {
		coefficient := unifrnd(-1, 1, rng)
		offspring1[dimension] = coefficient*male[dimension] +
			(1-coefficient)*female[dimension]
		offspring2[dimension] = coefficient*female[dimension] +
			(1-coefficient)*male[dimension]
	}

	maxVec(offspring1, lowerBound)
	minVec(offspring1, upperBound)
	maxVec(offspring2, lowerBound)
	minVec(offspring2, upperBound)

	return offspring1, offspring2
}

// crossoverForConfig selects the paper-specific crossover without changing
// the exported generic BLX operator. HMMA keeps precedence for configurations
// that combine flags because its Eq. (4) explicitly confines L to [0,1].
func crossoverForConfig(
	male, female []float64,
	config *Config,
	rng *rand.Rand,
) ([]float64, []float64) {
	if config.UseHMMA {
		return CrossoverBlend(
			male, female, 0, config.LowerBound, config.UpperBound, rng,
		)
	}

	if config.UseDESMA {
		return desmaCrossover(
			male, female, config.LowerBound, config.UpperBound, rng,
		)
	}

	return CrossoverBlend(
		male,
		female,
		effectiveCrossoverGamma(config),
		config.LowerBound,
		config.UpperBound,
		rng,
	)
}

// commitDESMAElite implements the lifecycle described around DESMA Eq. (16):
// a strictly improving elite replaces the current best population member and
// becomes the global-best attractor for the next male update. Both populations
// are sorted best-first when this helper is called. A tie keeps the male slot.
func commitDESMAElite(
	males, females []*Mayfly,
	globalBest *Best,
	elite *Mayfly,
	evaluator *constraintEvaluator,
) bool {
	if elite == nil || globalBest == nil || evaluator == nil ||
		!evaluator.betterMayflyThanBest(elite, *globalBest) {
		return false
	}

	switch {
	case len(males) > 0 && (len(females) == 0 || !evaluator.betterMayfly(females[0], males[0])):
		males[0] = elite
	case len(females) > 0:
		females[0] = elite
	default:
		return false
	}

	copyMayflyToBest(globalBest, elite)

	return true
}

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
