// Lévy Flight Implementation.
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
// alpha is the Lévy stability parameter (0 < alpha < 2). Mantegna's
// closed form degenerates at alpha=2 and must not be used there.
// beta is the Lévy scale parameter
// rng must not be nil (ensured by caller)
//
// Mantegna's algorithm approximates Lévy distribution using normal distributions.
// Returns a sanitized value (NaN/Inf checked by caller).
func levyFlight(alpha, beta float64, rng *rand.Rand) float64 {
	if rng == nil || math.IsNaN(alpha) || math.IsInf(alpha, 0) || alpha <= 0 || alpha >= 2 ||
		math.IsNaN(beta) || math.IsInf(beta, 0) || beta <= 0 {
		return 0
	}

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
	if math.Abs(v) < 1e-300 {
		v = math.Copysign(1e-300, v)
	}

	// Calculate Lévy flight step
	step := beta * u / math.Pow(math.Abs(v), 1/alpha)

	if math.IsNaN(step) {
		return 0
	}
	if math.IsInf(step, 0) {
		return math.Copysign(math.MaxFloat64, step)
	}

	return step
}

// levyFlightVec generates a vector of Lévy flight random numbers.
// rng must not be nil (ensured by caller).
func levyFlightVec(size int, alpha, beta float64, rng *rand.Rand) []float64 {
	if size <= 0 {
		return nil
	}

	vec := make([]float64, size)
	for i := range size {
		vec[i] = levyFlight(alpha, beta, rng)
	}

	return vec
}
