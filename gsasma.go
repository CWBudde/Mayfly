package mayfly

import (
	"math"
	"math/rand"
)

// GSASMA is the Golden Annealing Crossover-Mutation Mayfly Algorithm.
// Simulated annealing controls its late-stage velocity branch and the fixed
// golden-sine equation updates every male and female position.

func gsasmaAnnealedAttraction(
	currentCost, previousCost, temperature float64,
	rng *rand.Rand,
) bool {
	if currentCost < previousCost {
		return true
	}

	eta := math.Exp(-(currentCost - previousCost) / temperature)

	return rng.Float64() < eta
}

// gsasmaGoldenPosition implements Eq. (10) with fixed golden coefficients.
// r1 and r2 are scalars drawn once per individual, as in the paper.
func gsasmaGoldenPosition(
	position, personalBest []float64,
	lowerBound, upperBound float64,
	rng *rand.Rand,
) []float64 {
	r1 := rng.Float64() * 2 * math.Pi
	r2 := rng.Float64() * math.Pi
	sinR1 := math.Sin(r1)
	absSinR1 := math.Abs(sinR1)
	c1 := -math.Pi + (1-goldenRatioConjugate)*2*math.Pi
	c2 := -math.Pi + goldenRatioConjugate*2*math.Pi

	updated := make([]float64, len(position))
	for i := range position {
		updated[i] = position[i]*absSinR1 -
			r2*sinR1*math.Abs(c1*personalBest[i]-c2*position[i])
	}

	maxVec(updated, lowerBound)
	minVec(updated, upperBound)

	return updated
}

// gsasmaPositionStep composes the two GSASMA improvements in paper order:
// first apply the selected velocity, then refine that moved position with Eq.
// (10). Applying Eq. (10) to the old position would make the SA velocity stage
// behaviorally inert.
func gsasmaPositionStep(
	position, velocity, personalBest []float64,
	lowerBound, upperBound float64,
	rng *rand.Rand,
) []float64 {
	stepped := make([]float64, len(position))
	for i := range position {
		stepped[i] = position[i] + velocity[i]
	}

	maxVec(stepped, lowerBound)
	minVec(stepped, upperBound)

	return gsasmaGoldenPosition(stepped, personalBest, lowerBound, upperBound, rng)
}

func prepareGSASMAPopulations(
	males, females []*Mayfly,
	globalBest Best,
	previousCosts map[*Mayfly]float64,
	iteration, maxIterations int,
	g, dance, flight float64,
	config *Config,
	scheduler *AnnealingScheduler,
	evaluator *constraintEvaluator,
	rng *rand.Rand,
) {
	early := 2*iteration < maxIterations
	temperature := scheduler.GetTemperature()

	for i, female := range females {
		currentCost := female.Cost

		attracted := evaluator.betterMayfly(males[i], female)
		if !early {
			previousCost, ok := previousCosts[female]
			if !ok {
				previousCost = currentCost
			}

			attracted = gsasmaAnnealedAttraction(
				currentCost, previousCost, temperature, rng,
			)
		}

		if attracted {
			prepareAttractedFemale(female, males[i], g, config)
		} else {
			randomFlight := unifrndVec(-1, 1, config.ProblemSize, rng)
			for j := range config.ProblemSize {
				female.Velocity[j] = g*female.Velocity[j] + flight*randomFlight[j]
			}
		}

		maxVec(female.Velocity, config.VelMin)
		minVec(female.Velocity, config.VelMax)
		female.Position = gsasmaPositionStep(
			female.Position, female.Velocity, males[i].Best.Position,
			config.LowerBound, config.UpperBound, rng,
		)
		previousCosts[female] = currentCost
	}

	for _, male := range males {
		currentCost := male.Cost

		attracted := evaluator.better(
			evaluationFromBest(globalBest), evaluationFromMayfly(male),
		)
		if !early {
			previousCost, ok := previousCosts[male]
			if !ok {
				previousCost = currentCost
			}

			attracted = gsasmaAnnealedAttraction(
				currentCost, previousCost, temperature, rng,
			)
		}

		if attracted {
			prepareAttractedMale(male, globalBest, nil, g, 0, config)
		} else {
			randomDance := unifrndVec(-1, 1, config.ProblemSize, rng)
			for j := range config.ProblemSize {
				male.Velocity[j] = g*male.Velocity[j] + dance*randomDance[j]
			}
		}

		maxVec(male.Velocity, config.VelMin)
		minVec(male.Velocity, config.VelMax)
		male.Position = gsasmaPositionStep(
			male.Position, male.Velocity, male.Best.Position,
			config.LowerBound, config.UpperBound, rng,
		)
		previousCosts[male] = currentCost
	}
}

// retainGSASMAPreviousCosts preserves prior fitness for current survivors and
// releases entries for discarded offspring, bounding state to O(population).
func retainGSASMAPreviousCosts(
	previousCosts map[*Mayfly]float64,
	males, females []*Mayfly,
) map[*Mayfly]float64 {
	retained := make(map[*Mayfly]float64, len(males)+len(females))
	for _, population := range [][]*Mayfly{males, females} {
		for _, mayfly := range population {
			previousCost, ok := previousCosts[mayfly]
			if !ok {
				previousCost = mayfly.Cost
			}

			retained[mayfly] = previousCost
		}
	}

	return retained
}
