// Opposition-Based Learning (OBL) Implementation.
//
// Implements Opposition-Based Learning for enhanced search space coverage.
//
// Reference:
// Tizhoosh, H. R. (2005). Opposition-Based Learning: A New Scheme for Machine
// Intelligence. In International Conference on Computational Intelligence for
// Modelling, Control and Automation (Vol. 1, pp. 695-701). IEEE.
// DOI: 10.1109/CIMCA.2005.1631345
//
// Xu, Q., Wang, L., Wang, N., Hei, X., & Zhao, L. (2014). A review of
// opposition-based learning from 2005 to 2012. Engineering Applications of
// Artificial Intelligence, 29, 1-12.
//
// OBL simultaneously considers a candidate solution and its opposition point
// (x_opp = a + b - x) to accelerate convergence and expand search coverage.
// Used in EOBBMA and AOBLMOA. Opposition is not a GSASMA stage.

package mayfly

import (
	"math/rand"
)

// oppositionPoint generates the opposition point of a given position.
// The opposition point is calculated as: x_opp = a + b - x
// where a is the lower bound and b is the upper bound.
func oppositionPoint(position []float64, lowerBound, upperBound float64) []float64 {
	result := make([]float64, len(position))
	for i := range position {
		result[i] = lowerBound + upperBound - position[i]
	}

	return result
}

// gaussianUpdate performs a Bare Bones update using Gaussian sampling.
// The new position is sampled from a Gaussian distribution with mean
// at the midpoint between current and best positions, and standard
// deviation based on the distance between them.
func gaussianUpdate(current, best []float64, lowerBound, upperBound float64, rng *rand.Rand) []float64 {
	result := make([]float64, len(current))

	for i := range current {
		// Mean is the midpoint between current and best
		mean := (current[i] + best[i]) * 0.5

		// Standard deviation is half the distance between current and best
		// If they're the same, use a small exploration factor
		stddev := (current[i] - best[i]) * 0.5
		if stddev < 0 {
			stddev = -stddev
		}

		if stddev < 1e-10 {
			// Small exploration when current and best are very close
			stddev = (upperBound - lowerBound) * 0.01
		}

		// Sample from Gaussian distribution
		result[i] = mean + randn(rng)*stddev

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

// eliteBounds computes the dynamic search interval spanned by the first count
// mayflies of a (sorted) population: da_j = min_i x_ij and db_j = max_i x_ij.
//
// Where the elite collapses onto a single point in a dimension the interval
// degenerates, so that dimension falls back to the static search bounds.
func eliteBounds(mayflies []*Mayfly, count int, lowerBound, upperBound float64) ([]float64, []float64) {
	count = max(0, min(count, len(mayflies)))
	if count == 0 {
		return nil, nil
	}

	size := len(mayflies[0].Position)
	da := make([]float64, size)
	db := make([]float64, size)

	for j := range size {
		lo := mayflies[0].Position[j]
		hi := lo

		for i := 1; i < count; i++ {
			lo = min(lo, mayflies[i].Position[j])
			hi = max(hi, mayflies[i].Position[j])
		}

		if hi-lo < 1e-12 {
			lo, hi = lowerBound, upperBound
		}

		da[j] = lo
		db[j] = hi
	}

	return da, db
}

// eliteOppositionPoint generates the elite opposition point of a position.
//
// Unlike the static opposition of oppositionPoint, elite opposition-based
// learning reflects through the dynamic interval [da_j, db_j] spanned by the
// elite set and scales the reflection by a random generalised coefficient:
//
//	x*_j = k * (da_j + db_j) - x_j,   k ~ U(0, 1)
//
// A reflection that leaves the static search bounds is resampled uniformly from
// the elite interval, as prescribed by the EOBL formulation.
//
// Reference:
// Zhou, X., Wu, Z., Wang, H., & Rahnamayan, S. (2013). Elite opposition-based
// differential evolution for solving large-scale optimization problems.
func eliteOppositionPoint(position, da, db []float64, lowerBound, upperBound float64,
	rng *rand.Rand,
) []float64 {
	k := rng.Float64()
	result := make([]float64, len(position))

	for j := range position {
		result[j] = k*(da[j]+db[j]) - position[j]
		if result[j] < lowerBound || result[j] > upperBound {
			result[j] = da[j] + rng.Float64()*(db[j]-da[j])
		}
	}

	return result
}

// stochasticOppositionPoint generates the stochastic opposition point of a
// position, AOBLMOA Eq. (31):
//
//	x̃_j = (lb + ub − x_j) × r,   r ~ U(0, 1)
//
// The article describes r inconsistently, but the authors' implementation
// resolves both ambiguities: one uniform scalar is drawn for the complete
// offspring and multiplies the reflected vector. Since r is in [0,1], a valid
// input can still leave asymmetric bounds after scaling, so the result is
// clipped defensively.
//
// Reference:
// Zhao, Y.; Huang, C.; Zhang, M.; Cui, Y. AOBLMOA. Biomimetics 2023, 8(4), 381.
func stochasticOppositionPoint(position []float64, lowerBound, upperBound float64,
	rng *rand.Rand,
) []float64 {
	result := make([]float64, len(position))
	r := rng.Float64()

	for i := range position {
		result[i] = (lowerBound + upperBound - position[i]) * r
		result[i] = min(max(result[i], lowerBound), upperBound)
	}

	return result
}
