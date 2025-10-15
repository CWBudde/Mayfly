package mayfly

import "math/rand"

// unifrnd generates a random float64 between min and max.
func unifrnd(min, max float64, rng *rand.Rand) float64 {
	if rng == nil {
		return min + rand.Float64()*(max-min)
	}
	return min + rng.Float64()*(max-min)
}

// unifrndVec generates a vector of random float64 values between min and max.
func unifrndVec(min, max float64, size int, rng *rand.Rand) []float64 {
	vec := make([]float64, size)
	for i := range vec {
		vec[i] = unifrnd(min, max, rng)
	}
	return vec
}

// randn generates a normally distributed random number.
func randn(rng *rand.Rand) float64 {
	if rng == nil {
		return rand.NormFloat64()
	}
	return rng.NormFloat64()
}

// maxVec returns element-wise maximum of vector and scalar.
func maxVec(vec []float64, bound float64) {
	for i := range vec {
		if vec[i] < bound {
			vec[i] = bound
		}
	}
}

// minVec returns element-wise minimum of vector and scalar.
func minVec(vec []float64, bound float64) {
	for i := range vec {
		if vec[i] > bound {
			vec[i] = bound
		}
	}
}

// sortMayflies sorts mayflies by cost (ascending).
func sortMayflies(mayflies []*Mayfly) {
	// Simple bubble sort for small populations
	n := len(mayflies)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if mayflies[j].Cost > mayflies[j+1].Cost {
				mayflies[j], mayflies[j+1] = mayflies[j+1], mayflies[j]
			}
		}
	}
}
