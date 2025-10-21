// Package mayfly - Lévy Flight Implementation
//
// Implements Lévy flights using Mantegna's algorithm for heavy-tailed exploration.
//
// Reference:
// Mantegna, R.N. (1994). Fast, Accurate Algorithm for Numerical Simulation of
// Lévy Stable Stochastic Processes. Physical Review E, 49(5), 4677-4683.
// DOI: 10.1103/PhysRevE.49.4677
//
// Used in EOBBMA and AOBLMOA variants for occasional large jumps to escape
// local optima. The stability parameter α ∈ (0, 2] controls tail heaviness.
package mayfly

import (
	"math"
	"math/rand"
)

// levyFlight generates a Lévy flight random number using Mantegna's algorithm.
// alpha is the Lévy stability parameter (0 < alpha <= 2)
// beta is the Lévy scale parameter
// rng must not be nil (ensured by caller)
//
// Mantegna's algorithm approximates Lévy distribution using normal distributions.
// Returns a sanitized value (NaN/Inf checked by caller).
func levyFlight(alpha, beta float64, rng *rand.Rand) float64 {
	// Mantegna's algorithm for Lévy flight
	// Calculate sigma_u and sigma_v
	numerator := math.Gamma(1+alpha) * math.Sin(math.Pi*alpha/2)
	denominator := math.Gamma((1+alpha)/2) * alpha * math.Pow(2, (alpha-1)/2)
	sigmaU := math.Pow(numerator/denominator, 1/alpha)
	sigmaV := 1.0

	// Generate two Gaussian random numbers
	u := randn(rng) * sigmaU
	v := randn(rng) * sigmaV

	// Avoid division by zero or very small values
	if math.Abs(v) < 1e-10 {
		v = 1e-10
		if rng.Float64() < 0.5 {
			v = -v
		}
	}

	// Calculate Lévy flight step
	step := beta * u / math.Pow(math.Abs(v), 1/alpha)

	// Sanitize: check for NaN/Inf from Gamma functions or pow operations
	if math.IsNaN(step) || math.IsInf(step, 0) {
		// Return a moderate random step as fallback
		return beta * randn(rng)
	}

	return step
}

// levyFlightVec generates a vector of Lévy flight random numbers.
// rng must not be nil (ensured by caller).
func levyFlightVec(size int, alpha, beta float64, rng *rand.Rand) []float64 {
	vec := make([]float64, size)
	for i := 0; i < size; i++ {
		vec[i] = levyFlight(alpha, beta, rng)
	}

	return vec
}
