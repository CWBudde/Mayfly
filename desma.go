package mayfly

import "math/rand"

// generateEliteMayflies implements the DESMA dynamic elite strategy.
// It generates elite mayflies around the current global best position.
func generateEliteMayflies(currentBest Best, searchRange float64, eliteCount, problemSize int,
	lowerBound, upperBound float64, objFunc ObjectiveFunction, rng *rand.Rand) (*Mayfly, int) {

	bestElite := newMayfly(problemSize)
	copy(bestElite.Position, currentBest.Position)
	bestElite.Cost = currentBest.Cost
	copy(bestElite.Best.Position, currentBest.Position)
	bestElite.Best.Cost = currentBest.Cost

	funcEvals := 0

	// Generate elite mayflies around current best
	for i := 0; i < eliteCount; i++ {
		elite := newMayfly(problemSize)

		// Generate elite position: egbest = cgbest + r1 * R
		// where r1 is random vector in [-1, 1]
		for j := 0; j < problemSize; j++ {
			r1 := unifrnd(-1, 1, rng)
			elite.Position[j] = currentBest.Position[j] + r1*searchRange
		}

		// Apply boundary constraints
		maxVec(elite.Position, lowerBound)
		minVec(elite.Position, upperBound)

		// Evaluate elite mayfly
		elite.Cost = objFunc(elite.Position)
		funcEvals++

		// Update best elite if this one is better
		if elite.Cost < bestElite.Cost {
			copy(bestElite.Position, elite.Position)
			bestElite.Cost = elite.Cost
			copy(bestElite.Best.Position, elite.Position)
			bestElite.Best.Cost = elite.Cost
		}
	}

	return bestElite, funcEvals
}
