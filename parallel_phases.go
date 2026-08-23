package mayfly

import (
	"math"
	"math/rand"
)

func prepareEOBBMAFemales(females, males []*Mayfly, config *Config, rng *rand.Rand) {
	for i := range females {
		if rng.Float64() < 0.5 {
			newPosition := gaussianUpdate(females[i].Position, males[i].Position,
				config.LowerBound, config.UpperBound, rng)
			copy(females[i].Position, newPosition)

			continue
		}

		levyStep := levyFlightVec(config.ProblemSize, config.LevyAlpha, config.LevyBeta, rng)
		for j := range config.ProblemSize {
			females[i].Position[j] += levyStep[j] * (config.UpperBound - config.LowerBound) * 0.01
		}

		maxVec(females[i].Position, config.LowerBound)
		minVec(females[i].Position, config.UpperBound)
	}
}

func prepareEOBBMAMales(males []*Mayfly, globalBest Best, config *Config, rng *rand.Rand) {
	for _, male := range males {
		var target []float64
		if rng.Float64() < 0.5 {
			target = male.Best.Position
		} else {
			target = globalBest.Position
		}

		newPosition := gaussianUpdate(male.Position, target,
			config.LowerBound, config.UpperBound, rng)
		copy(male.Position, newPosition)
	}
}

func prepareStandardFemales(
	females, males []*Mayfly,
	g, flight float64,
	config *Config,
	rng *rand.Rand,
	evaluator *constraintEvaluator,
) {
	for i, female := range females {
		prepareStandardFemale(female, males[i], g, flight, config, rng, evaluator)
	}
}

// prepareStandardFemale performs the ordinary Mayfly velocity and position
// update for a single female, attracted to its paired male when that male is
// the better of the two and flying randomly otherwise.
func prepareStandardFemale(
	female, male *Mayfly,
	g, flight float64,
	config *Config,
	rng *rand.Rand,
	evaluator *constraintEvaluator,
) {
	randomFlight := unifrndVec(-1, 1, config.ProblemSize, rng)

	if evaluator.betterMayfly(male, female) {
		prepareAttractedFemale(female, male, g, config)
	} else {
		for j := range config.ProblemSize {
			female.Velocity[j] = g*female.Velocity[j] + flight*randomFlight[j]
		}
	}

	applyVelocityAndMove(female, config)
}

// prepareAttractedFemale applies the female attraction term of the Mayfly
// velocity update: the female is drawn toward the male she is paired with.
//
// It mirrors prepareAttractedMale and, like it, only writes the velocity; the
// caller clamps it and moves the individual through applyVelocityAndMove.
func prepareAttractedFemale(female, male *Mayfly, g float64, config *Config) {
	for j := range config.ProblemSize {
		distance := male.Position[j] - female.Position[j]
		female.Velocity[j] = g*female.Velocity[j] +
			config.A3*math.Exp(-config.Beta*distance*distance)*distance
	}
}

func prepareStandardMales(
	males []*Mayfly,
	globalBest Best,
	medianPosition []float64,
	g, dance, mpmaG float64,
	config *Config,
	rng *rand.Rand,
	evaluator *constraintEvaluator,
) {
	for _, male := range males {
		prepareStandardMale(male, globalBest, medianPosition, g, dance, mpmaG, config, rng, evaluator)
	}
}

// useMedianPosition reports whether the MPMA median term applies to this
// update. Variant paths that reuse the standard Mayfly update as a fallback —
// AOBLMOA, for instance — pass no median position, and must fall back to the
// plain Mayfly formula even when config.UseMPMA is set.
func useMedianPosition(medianPosition []float64, config *Config) bool {
	return config.UseMPMA && medianPosition != nil
}

// prepareStandardMale performs the ordinary Mayfly velocity and position update
// for a single male, attracted to its personal and the global best when the
// global best dominates it and dancing randomly otherwise.
func prepareStandardMale(
	male *Mayfly,
	globalBest Best,
	medianPosition []float64,
	g, dance, mpmaG float64,
	config *Config,
	rng *rand.Rand,
	evaluator *constraintEvaluator,
) {
	randomDance := unifrndVec(-1, 1, config.ProblemSize, rng)

	if evaluator.better(evaluationFromBest(globalBest), evaluationFromMayfly(male)) {
		prepareAttractedMale(male, globalBest, medianPosition, g, mpmaG, config)
	} else {
		gravity := g
		if useMedianPosition(medianPosition, config) {
			gravity = mpmaG
		}

		for j := range config.ProblemSize {
			male.Velocity[j] = gravity*male.Velocity[j] + dance*randomDance[j]
		}
	}

	applyVelocityAndMove(male, config)
}

// applyVelocityAndMove clamps a mayfly's velocity to the configured limits,
// steps its position by that velocity and clamps the position to the search
// bounds.
//
// It is the tail every Mayfly-style position update shares, extracted so that
// the attraction branches can be reused by variants that replace only the
// random branch.
func applyVelocityAndMove(mayfly *Mayfly, config *Config) {
	maxVec(mayfly.Velocity, config.VelMin)
	minVec(mayfly.Velocity, config.VelMax)

	for j := range config.ProblemSize {
		mayfly.Position[j] += mayfly.Velocity[j]
	}

	maxVec(mayfly.Position, config.LowerBound)
	minVec(mayfly.Position, config.UpperBound)
}

func prepareAttractedMale(
	male *Mayfly,
	globalBest Best,
	medianPosition []float64,
	g, mpmaG float64,
	config *Config,
) {
	useMedian := useMedianPosition(medianPosition, config)

	for j := range config.ProblemSize {
		personalDistance := male.Best.Position[j] - male.Position[j]
		globalDistance := globalBest.Position[j] - male.Position[j]

		if useMedian {
			medianDistance := medianPosition[j] - male.Position[j]
			male.Velocity[j] = mpmaG*male.Velocity[j] +
				config.A1*math.Exp(-config.Beta*personalDistance*personalDistance)*personalDistance +
				config.A2*math.Exp(-config.Beta*globalDistance*globalDistance)*globalDistance +
				config.MedianWeight*math.Exp(-config.Beta*medianDistance*medianDistance)*medianDistance

			continue
		}

		male.Velocity[j] = g*male.Velocity[j] +
			config.A1*math.Exp(-config.Beta*personalDistance*personalDistance)*personalDistance +
			config.A2*math.Exp(-config.Beta*globalDistance*globalDistance)*globalDistance
	}
}

func updatePersonalBests(males []*Mayfly, evaluator *constraintEvaluator) {
	for _, male := range males {
		if !evaluator.better(
			evaluationFromMayfly(male),
			evaluationFromBest(male.Best),
		) {
			continue
		}

		copy(male.Best.Position, male.Position)
		male.Best.Cost = male.Cost
		male.Best.ConstraintViolation = male.ConstraintViolation
	}
}
