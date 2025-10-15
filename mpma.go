package mayfly

import (
	"math"
	"sort"
)

// calculateMedianPosition calculates the median position across all mayflies in the population.
// For each dimension, it computes the median value across all mayflies.
func calculateMedianPosition(population []*Mayfly) []float64 {
	if len(population) == 0 {
		return nil
	}

	// Get problem size from first mayfly
	size := len(population[0].Position)
	median := make([]float64, size)

	// For each dimension, calculate the median
	for dim := 0; dim < size; dim++ {
		// Collect values for this dimension
		values := make([]float64, len(population))
		for i, mayfly := range population {
			values[i] = mayfly.Position[dim]
		}

		// Sort values to find median
		sort.Float64s(values)

		// Calculate median
		n := len(values)
		if n%2 == 1 {
			// Odd number: take middle value
			median[dim] = values[n/2]
		} else {
			// Even number: average of two middle values
			median[dim] = (values[n/2-1] + values[n/2]) / 2.0
		}
	}

	return median
}

// calculateWeightedMedianPosition calculates the weighted median position.
// Higher weights give more influence to certain positions (typically better solutions).
func calculateWeightedMedianPosition(population []*Mayfly, weights []float64) []float64 {
	if len(population) == 0 || len(weights) != len(population) {
		return nil
	}

	size := len(population[0].Position)
	median := make([]float64, size)

	// For each dimension, calculate the weighted median
	for dim := 0; dim < size; dim++ {
		// Create pairs of (value, weight) and sort by value
		type valueWeight struct {
			value  float64
			weight float64
		}

		pairs := make([]valueWeight, len(population))
		totalWeight := 0.0

		for i, mayfly := range population {
			pairs[i] = valueWeight{
				value:  mayfly.Position[dim],
				weight: weights[i],
			}
			totalWeight += weights[i]
		}

		// Sort by value
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].value < pairs[j].value
		})

		// Find weighted median (smallest value where cumulative weight >= 50%)
		halfWeight := totalWeight / 2.0
		cumWeight := 0.0

		for _, pair := range pairs {
			cumWeight += pair.weight
			if cumWeight >= halfWeight {
				median[dim] = pair.value
				break
			}
		}
	}

	return median
}

// calculateGravityCoefficient computes a time-varying gravity coefficient
// that controls exploration-exploitation balance.
// Returns a value that typically decreases from 1.0 to 0.0 over iterations.
func calculateGravityCoefficient(gravityType string, iteration, maxIterations int) float64 {
	// Normalize iteration to [0, 1]
	t := float64(iteration) / float64(maxIterations)

	switch gravityType {
	case "exponential":
		// Exponential decay: g = e^(-4t)
		// Decays faster than linear, good for quick convergence
		return math.Exp(-4.0 * t)

	case "sigmoid":
		// Sigmoid decay: g = 1 / (1 + e^(10(t-0.5)))
		// S-curve: slow decay at start, rapid in middle, slow at end
		return 1.0 / (1.0 + math.Exp(10.0*(t-0.5)))

	case "linear":
		fallthrough
	default:
		// Linear decay: g = 1 - t
		// Simple linear decrease from 1 to 0
		return 1.0 - t
	}
}
