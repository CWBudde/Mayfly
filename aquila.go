// Package mayfly - Aquila Optimizer Implementation
//
// Implements the Aquila Optimizer (AO) hunting strategies for AOBLMOA variant.
//
// Reference:
// Abualigah, L., Yousri, D., Abd Elaziz, M., Ewees, A. A., Al-qaness, M. A., &
// Gandomi, A. H. (2021). Aquila Optimizer: A novel meta-heuristic optimization
// algorithm. Computers & Industrial Engineering, 157, 107250.
// DOI: 10.1016/j.cie.2021.107250
//
// The Aquila Optimizer is inspired by eagle hunting behavior with four strategies:
// 1. X1 - Expanded exploration: High soar with vertical stoop
// 2. X2 - Narrowed exploration: Contour flight with short glide attack
// 3. X3 - Expanded exploitation: Low flight with slow descent attack
// 4. X4 - Narrowed exploitation: Walk and grab prey
//
// Strategy selection adapts over iterations (2/3 exploration, 1/3 exploitation).
package mayfly

import (
	"math"
	"math/rand"
)

// AquilaStrategy represents which hunting strategy to use.
type AquilaStrategy int

const (
	ExpandedExploration  AquilaStrategy = iota // X1: High soar with vertical stoop
	NarrowedExploration                        // X2: Contour flight with short glide
	ExpandedExploitation                       // X3: Low flight with slow descent
	NarrowedExploitation                       // X4: Walk and grab
)

// selectAquilaStrategy determines which hunting strategy to use based on iteration progress.
// The first 2/3 of iterations use exploration strategies (X1, X2),
// the last 1/3 uses exploitation strategies (X3, X4).
func selectAquilaStrategy(currentIter, maxIter int, rng *rand.Rand) AquilaStrategy {
	t := float64(currentIter) / float64(maxIter)

	if t <= 2.0/3.0 {
		// Exploration phase: use X1 or X2
		if rng.Float64() < 0.5 {
			return ExpandedExploration
		}

		return NarrowedExploration
	}

	// Exploitation phase: use X3 or X4
	if rng.Float64() < 0.5 {
		return ExpandedExploitation
	}

	return NarrowedExploitation
}

// aquilaExpandedExploration implements the X1 strategy: High soar with vertical stoop.
// This strategy is used for global exploration in early iterations.
//
// Formula: X1(t+1) = Xbest(t) * (1 - t/T) + (XM(t) - Xbest(t) * rand)
// where:
//   - Xbest is the best solution found so far
//   - XM is the mean position of the current population
//   - t is current iteration, T is max iterations
//   - rand is a random number in [0, 1]
func aquilaExpandedExploration(current, best, mean []float64, currentIter, maxIter int,
	lowerBound, upperBound float64, rng *rand.Rand) []float64 {
	result := make([]float64, len(current))
	t := float64(currentIter) / float64(maxIter)

	for i := 0; i < len(current); i++ {
		// X1(t+1) = Xbest(t) * (1 - t/T) + (XM(t) - Xbest(t) * rand)
		result[i] = best[i]*(1.0-t) + (mean[i] - best[i]*rng.Float64())

		// Apply bounds
		if result[i] < lowerBound {
			result[i] = lowerBound
		}

		if result[i] > upperBound {
			result[i] = upperBound
		}
	}

	return result
}

// aquilaNarrowedExploration implements the X2 strategy: Contour flight with short glide attack.
// This strategy uses Lévy flight for exploration with smaller steps than X1.
//
// Formula: X2(t+1) = Xbest(t) * Levy(D) + XR(t) + (y - x) * rand
// where:
//   - Levy(D) is the Lévy flight distribution
//   - XR is a random solution from the population
//   - y, x are random position components
//   - D is the problem dimension
func aquilaNarrowedExploration(current, best []float64, population []*Mayfly, problemSize int,
	lowerBound, upperBound float64, rng *rand.Rand) []float64 {
	result := make([]float64, len(current))

	// Generate Lévy flight multiplier
	levyD := generateLevyFlight(problemSize, 1.5, rng)

	// Select a random solution from population
	randomIdx := rng.Intn(len(population))
	xr := population[randomIdx].Position

	for i := 0; i < len(current); i++ {
		// Generate random position components
		y := rng.Float64()*(upperBound-lowerBound) + lowerBound
		x := rng.Float64()*(upperBound-lowerBound) + lowerBound

		// X2(t+1) = Xbest(t) * Levy(D) + XR(t) + (y - x) * rand
		result[i] = best[i]*levyD + xr[i] + (y-x)*rng.Float64()

		// Apply bounds
		if result[i] < lowerBound {
			result[i] = lowerBound
		}

		if result[i] > upperBound {
			result[i] = upperBound
		}
	}

	return result
}

// aquilaExpandedExploitation implements the X3 strategy: Low flight with slow descent attack.
// This strategy performs local exploitation around the best solution.
//
// Formula: X3(t+1) = (Xbest(t) - XM(t)) * α - rand + ((UB - LB) * rand + LB) * δ
// where:
//   - α is a random adjustment parameter
//   - δ is a small exploitation parameter
//   - XM is the mean position
//   - UB, LB are upper and lower bounds
func aquilaExpandedExploitation(current, best, mean []float64, currentIter, maxIter int,
	lowerBound, upperBound float64, rng *rand.Rand) []float64 {
	result := make([]float64, len(current))
	t := float64(currentIter) / float64(maxIter)

	// α decreases from 2 to 0 over iterations
	alpha := 2.0 * (1.0 - t)

	// δ is a small value for fine-tuning
	delta := 0.1

	for i := 0; i < len(current); i++ {
		// X3(t+1) = (Xbest(t) - XM(t)) * α - rand + ((UB - LB) * rand + LB) * δ
		exploration := ((upperBound-lowerBound)*rng.Float64() + lowerBound) * delta
		result[i] = (best[i]-mean[i])*alpha - rng.Float64() + exploration

		// Apply bounds
		if result[i] < lowerBound {
			result[i] = lowerBound
		}

		if result[i] > upperBound {
			result[i] = upperBound
		}
	}

	return result
}

// aquilaNarrowedExploitation implements the X4 strategy: Walk and grab prey.
// This is the most intensive exploitation strategy, used for local refinement.
//
// Formula: X4(t+1) = QF * Xbest(t) - (G1 * X(t) * rand) - G2 * Levy(D) + rand * G1
// where:
//   - QF is a quality function that increases over time
//   - G1, G2 are control parameters
//   - Levy(D) provides small random walks
func aquilaNarrowedExploitation(current, best []float64, currentIter, maxIter, problemSize int,
	lowerBound, upperBound float64, rng *rand.Rand) []float64 {
	result := make([]float64, len(current))
	t := float64(currentIter) / float64(maxIter)

	// QF is quality function: QF(t) = t^((2*rand - 1))
	qf := math.Pow(t, 2.0*rng.Float64()-1.0)

	// G1 decreases from 2 to 0
	g1 := 2.0 * rng.Float64() * (1.0 - t)

	// G2 is a random value in [0, 1]
	g2 := 2.0 * (1.0 - t)

	// Generate Lévy flight
	levyD := generateLevyFlight(problemSize, 1.5, rng)

	for i := 0; i < len(current); i++ {
		// X4(t+1) = QF * Xbest(t) - (G1 * X(t) * rand) - G2 * Levy(D) + rand * G1
		result[i] = qf*best[i] - (g1 * current[i] * rng.Float64()) - g2*levyD + rng.Float64()*g1

		// Apply bounds
		if result[i] < lowerBound {
			result[i] = lowerBound
		}

		if result[i] > upperBound {
			result[i] = upperBound
		}
	}

	return result
}

// applyAquilaStrategy applies the selected Aquila Optimizer strategy to update a mayfly's position.
// The strategy is selected based on the current iteration and problem characteristics.
//
// Parameters:
//   - mayfly: The mayfly to update
//   - globalBest: The best solution found so far
//   - population: The entire population (for mean calculation and random selection)
//   - strategy: Which Aquila hunting strategy to use
//   - currentIter, maxIter: Iteration progress
//   - config: Algorithm configuration
//
// Returns:
//   - New position for the mayfly
func applyAquilaStrategy(mayfly *Mayfly, globalBest Best, population []*Mayfly,
	strategy AquilaStrategy, currentIter, maxIter int, config *Config) []float64 {
	// Calculate mean position of population
	mean := make([]float64, config.ProblemSize)

	for _, m := range population {
		for i := 0; i < config.ProblemSize; i++ {
			mean[i] += m.Position[i]
		}
	}

	for i := 0; i < config.ProblemSize; i++ {
		mean[i] /= float64(len(population))
	}

	// Apply the selected strategy
	switch strategy {
	case ExpandedExploration:
		return aquilaExpandedExploration(mayfly.Position, globalBest.Position, mean,
			currentIter, maxIter, config.LowerBound, config.UpperBound, config.Rand)

	case NarrowedExploration:
		return aquilaNarrowedExploration(mayfly.Position, globalBest.Position, population,
			config.ProblemSize, config.LowerBound, config.UpperBound, config.Rand)

	case ExpandedExploitation:
		return aquilaExpandedExploitation(mayfly.Position, globalBest.Position, mean,
			currentIter, maxIter, config.LowerBound, config.UpperBound, config.Rand)

	case NarrowedExploitation:
		return aquilaNarrowedExploitation(mayfly.Position, globalBest.Position,
			currentIter, maxIter, config.ProblemSize, config.LowerBound, config.UpperBound, config.Rand)

	default:
		// Should never happen, but return current position as fallback
		result := make([]float64, len(mayfly.Position))
		copy(result, mayfly.Position)

		return result
	}
}

// generateLevyFlight generates a Lévy flight step using Mantegna's algorithm.
// This is a simplified version used by Aquila Optimizer.
func generateLevyFlight(dim int, alpha float64, rng *rand.Rand) float64 {
	// Mantegna's algorithm for Lévy flight
	sigma := math.Pow(
		math.Gamma(1.0+alpha)*math.Sin(math.Pi*alpha/2.0)/
			(math.Gamma((1.0+alpha)/2.0)*alpha*math.Pow(2.0, (alpha-1.0)/2.0)),
		1.0/alpha,
	)

	u := rng.NormFloat64() * sigma
	v := rng.NormFloat64()

	step := u / math.Pow(math.Abs(v), 1.0/alpha)

	return step
}
