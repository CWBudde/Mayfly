package mayfly

import (
	"math/rand"
)

// GSASMA (Golden Sine Algorithm with Simulated Annealing Mayfly Algorithm)
// is an enhanced variant that combines three powerful optimization techniques:
//
// 1. Golden Sine Algorithm (GSA): For adaptive exploration using golden ratio
// 2. Simulated Annealing (SA): For escaping local optima via probabilistic acceptance
// 3. Hybrid Mutation: Combining Cauchy (exploration) and Gaussian (exploitation)
// 4. Opposition-Based Learning (OBL): For expanding search space coverage
//
// This file contains integration logic specific to GSASMA that orchestrates
// these components within the main Mayfly optimization loop.

// Returns: (updatedGlobalBest, updatedGlobalBestCost, funcEvals).
//
//nolint:unused // retained for focused GSASMA helper compatibility.
func applyGSASMAToEliteMales(males []*Mayfly, eliteRatio float64, globalBest []float64,
	globalBestCost float64, goldenFactor float64, currentIter, maxIter int,
	lowerBound, upperBound float64, scheduler *AnnealingScheduler,
	objectiveFunc ObjectiveFunction, rng *rand.Rand,
) ([]float64, float64, int) {
	numElite := min(max(int(float64(len(males))*eliteRatio), 1), len(males))

	funcEvals := 0
	updatedGlobalBest := globalBest
	updatedGlobalBestCost := globalBestCost

	// Apply Golden Sine Algorithm with Simulated Annealing to elite males
	for i := range numElite {
		// Generate candidate position using adaptive Golden Sine
		candidatePos := goldenSineUpdateAdaptive(
			males[i].Position,
			globalBest,
			goldenFactor,
			currentIter,
			maxIter,
			lowerBound,
			upperBound,
			rng,
		)

		// Evaluate candidate
		candidateCost := objectiveFunc(candidatePos)
		funcEvals++

		// Use simulated annealing acceptance criterion
		if shouldAccept(males[i].Cost, candidateCost, scheduler.GetTemperature(), rng) {
			// Accept: update male position
			copy(males[i].Position, candidatePos)
			males[i].Cost = candidateCost

			// Update personal best if better
			if candidateCost < males[i].Best.Cost {
				copy(males[i].Best.Position, candidatePos)
				males[i].Best.Cost = candidateCost
			}

			// Update global best if this is the new best
			if candidateCost < updatedGlobalBestCost {
				updatedGlobalBest = make([]float64, len(candidatePos))
				copy(updatedGlobalBest, candidatePos)

				updatedGlobalBestCost = candidateCost
			}
		}
	}

	return updatedGlobalBest, updatedGlobalBestCost, funcEvals
}

func applyGSASMAToEliteMalesWithEvaluator(
	males []*Mayfly,
	eliteRatio float64,
	globalBest Best,
	goldenFactor float64,
	currentIter, maxIter int,
	lowerBound, upperBound float64,
	scheduler *AnnealingScheduler,
	evaluator *constraintEvaluator,
	rng *rand.Rand,
) (Best, int) {
	numElite := min(max(int(float64(len(males))*eliteRatio), 1), len(males))
	updatedGlobalBest := cloneBest(globalBest)

	for i := range numElite {
		candidate := newMayfly(len(males[i].Position))
		candidate.Position = goldenSineUpdateAdaptive(
			males[i].Position,
			globalBest.Position,
			goldenFactor,
			currentIter,
			maxIter,
			lowerBound,
			upperBound,
			rng,
		)
		evaluator.evaluateMayfly(candidate, false)

		probability := evaluator.acceptanceProbability(
			evaluationFromMayfly(males[i]), evaluationFromMayfly(candidate), scheduler.GetTemperature(),
		)
		if rng.Float64() >= probability {
			continue
		}

		copy(males[i].Position, candidate.Position)
		males[i].Cost = candidate.Cost
		males[i].ConstraintViolation = candidate.ConstraintViolation

		if evaluator.better(evaluationFromMayfly(males[i]), evaluationFromBest(males[i].Best)) {
			copy(males[i].Best.Position, males[i].Position)
			males[i].Best.Cost = males[i].Cost
			males[i].Best.ConstraintViolation = males[i].ConstraintViolation
		}

		if evaluator.betterMayflyThanBest(males[i], updatedGlobalBest) {
			copyMayflyToBest(&updatedGlobalBest, males[i])
		}
	}

	return updatedGlobalBest, numElite
}

func applyOBLToGlobalBestWithEvaluator(
	globalBest Best,
	lowerBound, upperBound float64,
	evaluator *constraintEvaluator,
) (Best, bool) {
	opposition := newMayfly(len(globalBest.Position))
	opposition.Position = oppositionPoint(globalBest.Position, lowerBound, upperBound)
	evaluator.evaluateMayfly(opposition, false)

	if !evaluator.betterMayflyThanBest(opposition, globalBest) {
		return globalBest, false
	}

	return bestFromMayfly(opposition), true
}

// Returns: mutated offspring.
//
//nolint:unused // reserved for the GSASMA variant; not wired into Optimize() yet.
func applyHybridMutationGSASMA(offspring []*Mayfly, nMutants int, mutationRate float64,
	currentIter, maxIter int, cauchyMutationRate, lowerBound, upperBound float64,
	rng *rand.Rand,
) []*Mayfly {
	// Calculate adaptive Cauchy probability based on iteration progress
	iterRatio := float64(currentIter) / float64(maxIter)

	var cauchyProb float64

	switch {
	case iterRatio < 0.33:
		// Early phase: high Cauchy for exploration
		cauchyProb = 0.7
	case iterRatio < 0.66:
		// Middle phase: balanced
		cauchyProb = 0.5
	default:
		// Late phase: low Cauchy for exploitation
		cauchyProb = cauchyMutationRate // Use configured rate (default 0.3)
	}

	// Apply hybrid mutation to create mutants
	for range nMutants {
		// Select random parent from offspring
		i := rng.Intn(len(offspring))
		parent := offspring[i]

		// Create mutant
		mutant := newMayfly(len(parent.Position))

		// Apply hybrid mutation
		mutant.Position = HybridMutate(
			parent.Position,
			mutationRate,
			lowerBound,
			upperBound,
			cauchyProb,
			rng,
		)

		// Note: Cost will be evaluated in main loop
		// Setting to infinity to ensure it gets evaluated
		mutant.Cost = parent.Cost // Placeholder, will be updated

		offspring = append(offspring, mutant)
	}

	return offspring
}

// Returns: (updatedGlobalBest, updatedGlobalBestCost, improved).
// The caller must account for the single evaluation of the opposition point.
//
//nolint:unused // retained for focused GSASMA helper compatibility.
func applyOBLToGlobalBest(globalBest []float64, globalBestCost float64,
	lowerBound, upperBound float64, objectiveFunc ObjectiveFunction,
) ([]float64, float64, bool) {
	// Generate opposition point
	oppPos := oppositionPoint(globalBest, lowerBound, upperBound)

	// Evaluate opposition point
	oppCost := objectiveFunc(oppPos)

	// If opposition is better, update global best
	if oppCost < globalBestCost {
		updatedGlobalBest := make([]float64, len(oppPos))
		copy(updatedGlobalBest, oppPos)

		return updatedGlobalBest, oppCost, true
	}

	// No improvement
	return globalBest, globalBestCost, false
}

// Returns: adaptive Cauchy rate.
//
//nolint:unused // reserved for the GSASMA variant; not wired into Optimize() yet.
func calculateAdaptiveCauchyRate(males []*Mayfly, baseCauchyRate float64) float64 {
	if len(males) < 2 {
		return baseCauchyRate
	}

	// Calculate population diversity (average std dev across dimensions)
	problemSize := len(males[0].Position)
	totalStdDev := 0.0

	for dim := range problemSize {
		// Calculate mean for this dimension
		mean := 0.0
		for _, m := range males {
			mean += m.Position[dim]
		}

		mean /= float64(len(males))

		// Calculate standard deviation
		variance := 0.0

		for _, m := range males {
			diff := m.Position[dim] - mean
			variance += diff * diff
		}

		variance /= float64(len(males))
		stdDev := variance // Simplified: using variance as proxy

		totalStdDev += stdDev
	}

	avgStdDev := totalStdDev / float64(problemSize)

	// Normalize diversity (assume search space is [0, 1] normalized)
	// Higher diversity → lower rate multiplier
	// Lower diversity → higher rate multiplier
	diversityFactor := 1.0 / (1.0 + avgStdDev)

	// Adaptive rate: increase when diversity is low
	adaptiveRate := baseCauchyRate * (1.0 + diversityFactor)

	// Cap at 0.9 to prevent excessive mutation
	if adaptiveRate > 0.9 {
		adaptiveRate = 0.9
	}

	return adaptiveRate
}
