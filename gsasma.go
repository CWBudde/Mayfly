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

// applyGSASMAToEliteMales applies GSASMA-specific enhancements to elite male mayflies.
// This combines Golden Sine Algorithm updates with simulated annealing acceptance.
//
// Strategy:
//  1. Select elite males (top 20% by default)
//  2. Apply Golden Sine Algorithm to generate candidate positions
//  3. Use simulated annealing to accept/reject candidates
//  4. Update global best if better solutions found
//
// Parameters:
//   - males: male population (sorted by cost, best first)
//   - eliteRatio: proportion considered elite (e.g., 0.2 for top 20%)
//   - globalBest: current global best position
//   - globalBestCost: current global best cost
//   - goldenFactor: GSA scaling factor
//   - currentIter: current iteration number
//   - maxIter: maximum iterations
//   - lowerBound: lower bound for variables
//   - upperBound: upper bound for variables
//   - scheduler: simulated annealing scheduler
//   - objectiveFunc: objective function
//   - rng: random number generator
//
// Returns: (updatedGlobalBest, updatedGlobalBestCost, funcEvals)
func applyGSASMAToEliteMales(males []*Mayfly, eliteRatio float64, globalBest []float64,
	globalBestCost float64, goldenFactor float64, currentIter, maxIter int,
	lowerBound, upperBound float64, scheduler *AnnealingScheduler,
	objectiveFunc ObjectiveFunction, rng *rand.Rand) ([]float64, float64, int) {

	numElite := int(float64(len(males)) * eliteRatio)
	if numElite < 1 {
		numElite = 1
	}
	if numElite > len(males) {
		numElite = len(males)
	}

	funcEvals := 0
	updatedGlobalBest := globalBest
	updatedGlobalBestCost := globalBestCost

	// Apply Golden Sine Algorithm with Simulated Annealing to elite males
	for i := 0; i < numElite; i++ {
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

// applyHybridMutationGSASMA applies hybrid Cauchy-Gaussian mutation with
// adaptive probability based on iteration progress.
//
// Strategy:
//   - Early iterations (0-33%): High Cauchy probability (0.7) for exploration
//   - Middle iterations (33-66%): Balanced probability (0.5)
//   - Late iterations (66-100%): Low Cauchy probability (0.3) for exploitation
//
// This adaptive approach naturally transitions from exploration to exploitation
// as the optimization progresses.
//
// Parameters:
//   - offspring: offspring population to mutate
//   - nMutants: number of mutants to create
//   - mutationRate: mutation rate (proportion of dimensions)
//   - currentIter: current iteration
//   - maxIter: maximum iterations
//   - cauchyMutationRate: base Cauchy mutation probability
//   - lowerBound: lower bound for variables
//   - upperBound: upper bound for variables
//   - rng: random number generator
//
// Returns: mutated offspring
func applyHybridMutationGSASMA(offspring []*Mayfly, nMutants int, mutationRate float64,
	currentIter, maxIter int, cauchyMutationRate, lowerBound, upperBound float64,
	rng *rand.Rand) []*Mayfly {

	// Calculate adaptive Cauchy probability based on iteration progress
	iterRatio := float64(currentIter) / float64(maxIter)
	var cauchyProb float64

	if iterRatio < 0.33 {
		// Early phase: high Cauchy for exploration
		cauchyProb = 0.7
	} else if iterRatio < 0.66 {
		// Middle phase: balanced
		cauchyProb = 0.5
	} else {
		// Late phase: low Cauchy for exploitation
		cauchyProb = cauchyMutationRate // Use configured rate (default 0.3)
	}

	// Apply hybrid mutation to create mutants
	for k := 0; k < nMutants; k++ {
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

// applyOBLToGlobalBest applies Opposition-Based Learning to the global best solution.
// This generates an opposition point and evaluates it, potentially finding a better solution.
//
// Opposition-Based Learning (OBL) is based on the concept that the opposite
// of a candidate solution might be closer to the optimum than the candidate itself.
// For a point x in [a, b], its opposition is: x_opp = a + b - x
//
// Strategy:
//  1. Generate opposition point of global best
//  2. Evaluate opposition point
//  3. If better than global best, replace it
//
// This is particularly effective when the algorithm is converging, as it can
// help discover solutions on the opposite side of the search space.
//
// Parameters:
//   - globalBest: current global best position
//   - globalBestCost: current global best cost
//   - lowerBound: lower bound for variables
//   - upperBound: upper bound for variables
//   - objectiveFunc: objective function
//   - rng: random number generator
//
// Returns: (updatedGlobalBest, updatedGlobalBestCost, funcEvals, improved)
func applyOBLToGlobalBest(globalBest []float64, globalBestCost float64,
	lowerBound, upperBound float64, objectiveFunc ObjectiveFunction,
	rng *rand.Rand) ([]float64, float64, int, bool) {

	// Generate opposition point
	oppPos := oppositionPoint(globalBest, lowerBound, upperBound)

	// Evaluate opposition point
	oppCost := objectiveFunc(oppPos)
	funcEvals := 1

	// If opposition is better, update global best
	if oppCost < globalBestCost {
		updatedGlobalBest := make([]float64, len(oppPos))
		copy(updatedGlobalBest, oppPos)
		return updatedGlobalBest, oppCost, funcEvals, true
	}

	// No improvement
	return globalBest, globalBestCost, funcEvals, false
}

// calculateAdaptiveCauchyRate calculates an adaptive Cauchy mutation rate
// based on population diversity. This helps maintain exploration when diversity
// drops too low.
//
// Strategy:
//   - High diversity: use base rate (don't need extra exploration)
//   - Low diversity: increase rate (need more exploration to escape)
//
// Diversity measure: average standard deviation across all dimensions
//
// Parameters:
//   - males: male population
//   - baseCauchyRate: configured Cauchy mutation rate
//
// Returns: adaptive Cauchy rate
func calculateAdaptiveCauchyRate(males []*Mayfly, baseCauchyRate float64) float64 {
	if len(males) < 2 {
		return baseCauchyRate
	}

	// Calculate population diversity (average std dev across dimensions)
	problemSize := len(males[0].Position)
	totalStdDev := 0.0

	for dim := 0; dim < problemSize; dim++ {
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
