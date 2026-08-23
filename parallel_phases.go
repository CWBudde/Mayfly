package mayfly

import (
	"math"
	"math/rand"
)

func prepareEOBBMAFemales(
	females, males []*Mayfly,
	config *Config,
	rng *rand.Rand,
	evaluator *constraintEvaluator,
) {
	femaleSnapshot := cloneMayflies(females)
	maleSnapshot := cloneMayflies(males)
	for i := range females {
		female := &femaleSnapshot[i]
		male := &maleSnapshot[i]
		var position, center []float64
		if evaluator.betterMayfly(male, female) {
			position, center = eobbmaGaussianFemalePosition(female.Position, male.Position, rng)
		} else {
			position = eobbmaLevyPosition(female.Position, config.LevyAlpha, config.LevyBeta, rng)
			center = female.Position
		}
		copy(females[i].Position, eobbmaRepairPosition(
			position, center, config.LowerBound, config.UpperBound,
		))
	}
}

func prepareEOBBMAMales(males []*Mayfly, globalBest Best, config *Config, rng *rand.Rand) {
	snapshot := cloneMayflies(males)
	for i := range males {
		male := &snapshot[i]
		var position, center []float64
		if i == 0 {
			position = eobbmaLevyPosition(male.Position, config.LevyAlpha, config.LevyBeta, rng)
			center = male.Position
		} else {
			peer1, peer2 := eobbmaDistinctPeers(snapshot, i, rng)
			position, center = eobbmaGaussianMalePosition(
				male.Best.Position, globalBest.Position, peer1, peer2,
				male.Cost, globalBest.Cost, rng,
			)
		}
		copy(males[i].Position, eobbmaRepairPosition(
			position, center, config.LowerBound, config.UpperBound,
		))
	}
}

func eobbmaDistinctPeers(population []Mayfly, excluded int, rng *rand.Rand) ([]float64, []float64) {
	indices := make([]int, 0, len(population)-1)
	for i := range population {
		if i != excluded {
			indices = append(indices, i)
		}
	}
	if len(indices) < 2 {
		return population[excluded].Position, population[excluded].Best.Position
	}
	firstOffset := rng.Intn(len(indices))
	first := indices[firstOffset]
	indices[firstOffset] = indices[len(indices)-1]
	indices = indices[:len(indices)-1]
	second := indices[rng.Intn(len(indices))]
	return population[first].Position, population[second].Position
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
	distanceSquared := 0.0
	for j := range config.ProblemSize {
		delta := male.Position[j] - female.Position[j]
		distanceSquared += delta * delta
	}
	attraction := config.A3 * math.Exp(-config.Beta*distanceSquared)
	for j := range config.ProblemSize {
		distance := male.Position[j] - female.Position[j]
		female.Velocity[j] = g*female.Velocity[j] +
			attraction*distance
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
	personalDistanceSquared := 0.0
	globalDistanceSquared := 0.0
	medianDistanceSquared := 0.0
	for j := range config.ProblemSize {
		personalDelta := male.Best.Position[j] - male.Position[j]
		globalDelta := globalBest.Position[j] - male.Position[j]
		personalDistanceSquared += personalDelta * personalDelta
		globalDistanceSquared += globalDelta * globalDelta
		if useMedian {
			medianDelta := medianPosition[j] - male.Position[j]
			medianDistanceSquared += medianDelta * medianDelta
		}
	}
	personalAttraction := config.A1 * math.Exp(-config.Beta*personalDistanceSquared)
	globalAttraction := config.A2 * math.Exp(-config.Beta*globalDistanceSquared)
	medianAttraction := config.MedianWeight * math.Exp(-config.Beta*medianDistanceSquared)

	for j := range config.ProblemSize {
		personalDistance := male.Best.Position[j] - male.Position[j]
		globalDistance := globalBest.Position[j] - male.Position[j]

		if useMedian {
			medianDistance := medianPosition[j] - male.Position[j]
			male.Velocity[j] = mpmaG*male.Velocity[j] +
				personalAttraction*personalDistance +
				globalAttraction*globalDistance +
				medianAttraction*medianDistance

			continue
		}

		male.Velocity[j] = g*male.Velocity[j] +
			personalAttraction*personalDistance +
			globalAttraction*globalDistance
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
