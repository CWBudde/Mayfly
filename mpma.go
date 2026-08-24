package mayfly

import (
	"context"
	"math"
	"sort"
)

// calculateMedianPosition returns the position of the median fitness-ranked
// mayfly. The caller must supply the population in fitness order, as MPMA does.
// For an even population the two middle ranked position vectors are averaged.
// This deliberately does not calculate coordinate-wise medians: doing so can
// synthesize a point that no middle-ranked mayfly represents.
func calculateMedianPosition(population []*Mayfly) []float64 {
	if !validPopulationDimensions(population) {
		return nil
	}

	size := len(population[0].Position)
	median := make([]float64, size)

	middle := len(population) / 2
	if len(population)%2 == 1 {
		copy(median, population[middle].Position)

		return median
	}

	for dimension := range size {
		median[dimension] = (population[middle-1].Position[dimension] +
			population[middle].Position[dimension]) / 2
	}

	return median
}

func validPopulationDimensions(population []*Mayfly) bool {
	if len(population) == 0 || population[0] == nil || len(population[0].Position) == 0 {
		return false
	}

	size := len(population[0].Position)
	for _, mayfly := range population {
		if mayfly == nil || len(mayfly.Position) != size {
			return false
		}
	}

	return true
}

// calculateWeightedMedianPosition calculates the weighted median position.
// Higher weights give more influence to certain positions (typically better solutions).
func calculateWeightedMedianPosition(population []*Mayfly, weights []float64) []float64 {
	if !validPopulationDimensions(population) || len(weights) != len(population) {
		return nil
	}

	maxWeight := 0.0

	for _, weight := range weights {
		if math.IsNaN(weight) || math.IsInf(weight, 0) || weight < 0 {
			return nil
		}

		maxWeight = max(maxWeight, weight)
	}

	if maxWeight == 0 {
		return nil
	}

	size := len(population[0].Position)
	median := make([]float64, size)

	// For each dimension, calculate the weighted median
	for dim := range size {
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
				weight: weights[i] / maxWeight,
			}
			totalWeight += pairs[i].weight
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

func calculateMedianPositionParallel(
	ctx context.Context,
	population []*Mayfly,
	maxWorkers int,
) ([]float64, error) {
	if !validPopulationDimensions(population) {
		return nil, nil
	}

	size := len(population[0].Position)
	median := make([]float64, size)

	err := parallelFor(ctx, size, maxWorkers, func(dimension int) {
		middle := len(population) / 2
		if len(population)%2 == 1 {
			median[dimension] = population[middle].Position[dimension]

			return
		}

		median[dimension] = (population[middle-1].Position[dimension] +
			population[middle].Position[dimension]) / 2
	})
	if err != nil {
		return nil, err
	}

	return median, nil
}

func calculateWeightedMedianPositionParallel(
	ctx context.Context,
	population []*Mayfly,
	weights []float64,
	maxWorkers int,
) ([]float64, error) {
	if !validPopulationDimensions(population) || len(weights) != len(population) {
		return nil, nil
	}

	maxWeight := 0.0

	for _, weight := range weights {
		if math.IsNaN(weight) || math.IsInf(weight, 0) || weight < 0 {
			return nil, nil
		}

		maxWeight = max(maxWeight, weight)
	}

	if maxWeight == 0 {
		return nil, nil
	}

	type valueWeight struct {
		value  float64
		weight float64
	}

	size := len(population[0].Position)
	median := make([]float64, size)

	err := parallelFor(ctx, size, maxWorkers, func(dimension int) {
		pairs := make([]valueWeight, len(population))
		totalWeight := 0.0

		for i, mayfly := range population {
			weight := weights[i] / maxWeight
			pairs[i] = valueWeight{value: mayfly.Position[dimension], weight: weight}
			totalWeight += weight
		}

		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].value < pairs[j].value
		})

		halfWeight := totalWeight / 2
		cumulativeWeight := 0.0

		for _, pair := range pairs {
			cumulativeWeight += pair.weight
			if cumulativeWeight >= halfWeight {
				median[dimension] = pair.value

				return
			}
		}
	})
	if err != nil {
		return nil, err
	}

	return median, nil
}

// calculateGravityCoefficient computes a time-varying gravity coefficient
// that controls exploration-exploitation balance.
// Returns a value that typically decreases from 1.0 to 0.0 over iterations.
func calculateGravityCoefficient(gravityType string, iteration, maxIterations int) float64 {
	if maxIterations <= 0 {
		return math.NaN()
	}

	// Normalize iteration to [0, 1]
	t := min(max(float64(iteration)/float64(maxIterations), 0), 1)

	switch gravityType {
	case GravityPaper:
		// MPMA Eq. (17): g(t) = 0.5*sqrt(1-(t/T)^2) + 0.4.
		return 0.5*math.Sqrt(max(1.0-t*t, 0)) + 0.4

	case GravityExponential:
		// Exponential decay: g = e^(-4t)
		// Decays faster than linear, good for quick convergence
		return math.Exp(-4.0 * t)

	case GravitySigmoid:
		// Sigmoid decay: g = 1 / (1 + e^(10(t-0.5)))
		// S-curve: slow decay at start, rapid in middle, slow at end
		return 1.0 / (1.0 + math.Exp(10.0*(t-0.5)))

	default: // GravityLinear
		// Linear decay: g = 1 - t
		// Simple linear decrease from 1 to 0
		return 1.0 - t
	}
}
