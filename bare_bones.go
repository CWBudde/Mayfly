package mayfly

import (
	"math"
	"math/rand"
)

// eobbmaGaussianMalePosition implements BBMA Eqs. (26)-(28) for a non-best
// male. peer1 and peer2 are two distinct randomly selected male positions.
func eobbmaGaussianMalePosition(
	pbest, globalBest, peer1, peer2 []float64,
	currentCost, globalBestCost float64,
	rng *rand.Rand,
) ([]float64, []float64) {
	if rng == nil || len(pbest) == 0 || len(globalBest) != len(pbest) ||
		len(peer1) != len(pbest) || len(peer2) != len(pbest) {
		return nil, nil
	}

	position := make([]float64, len(pbest))
	mean := make([]float64, len(pbest))
	fitnessFactor := math.Exp(globalBestCost - currentCost)
	for dimension := range pbest {
		mean[dimension] = (globalBest[dimension] + pbest[dimension]) / 2
		disturbance := rng.Float64() * math.Abs(peer1[dimension]-peer2[dimension]) * fitnessFactor
		deviation := math.Abs(globalBest[dimension]-pbest[dimension]) + disturbance
		position[dimension] = mean[dimension] + rng.NormFloat64()*deviation
	}

	return position, mean
}

// eobbmaGaussianFemalePosition implements BBMA Eqs. (34)-(35) for a female
// whose paired male is fitter. It also returns the Gaussian mean required by
// the paper's boundary pullback rule.
func eobbmaGaussianFemalePosition(female, pairedMale []float64, rng *rand.Rand) ([]float64, []float64) {
	if rng == nil || len(female) == 0 || len(pairedMale) != len(female) {
		return nil, nil
	}

	position := make([]float64, len(female))
	mean := make([]float64, len(female))
	for dimension := range female {
		mean[dimension] = (female[dimension] + pairedMale[dimension]) / 2
		deviation := math.Sqrt(math.Abs(pairedMale[dimension] - female[dimension]))
		position[dimension] = mean[dimension] + rng.NormFloat64()*deviation
	}

	return position, mean
}

// eobbmaLevyPosition implements the multiplicative Lévy update used by both
// the best male and females at least as fit as their paired male:
// x(t+1) = x(t) + x(t)*Levy(beta).
func eobbmaLevyPosition(position []float64, alpha, beta float64, rng *rand.Rand) []float64 {
	if len(position) == 0 {
		return nil
	}

	step := levyFlightVec(len(position), alpha, beta, rng)
	result := make([]float64, len(position))
	for dimension := range position {
		result[dimension] = position[dimension] + position[dimension]*step[dimension]
	}

	return result
}

// eobbmaRepairPosition implements BBMA Eq. (37), pulling an escaped Gaussian
// sample back toward its Gaussian mean instead of pinning it to a boundary.
func eobbmaRepairPosition(position, mean []float64, lowerBound, upperBound float64) []float64 {
	if len(position) == 0 || len(mean) != len(position) || !isFinite(lowerBound) ||
		!isFinite(upperBound) || lowerBound > upperBound {
		return nil
	}

	repaired := append([]float64(nil), position...)
	for dimension, coordinate := range repaired {
		if coordinate >= lowerBound && coordinate <= upperBound && isFinite(coordinate) {
			continue
		}

		center := min(max(mean[dimension], lowerBound), upperBound)
		border := lowerBound
		if coordinate > upperBound || math.IsInf(coordinate, 1) {
			border = upperBound
		}
		denominator := coordinate - center
		if denominator == 0 || math.IsNaN(denominator) {
			repaired[dimension] = center
			continue
		}

		repaired[dimension] = center + (border-center)*(border-center)/denominator
		if !isFinite(repaired[dimension]) {
			repaired[dimension] = center
		}
		repaired[dimension] = min(max(repaired[dimension], lowerBound), upperBound)
	}

	return repaired
}
