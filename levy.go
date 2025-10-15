package mayfly

import (
	"math"
	"math/rand"
)

// levyFlight generates a Lévy flight random number using Mantegna's algorithm.
// alpha is the Lévy stability parameter (0 < alpha <= 2)
// beta is the Lévy scale parameter
// Mantegna's algorithm approximates Lévy distribution using normal distributions.
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

	// Calculate Lévy flight step
	step := beta * u / math.Pow(math.Abs(v), 1/alpha)

	return step
}

// levyFlightVec generates a vector of Lévy flight random numbers.
func levyFlightVec(size int, alpha, beta float64, rng *rand.Rand) []float64 {
	vec := make([]float64, size)
	for i := 0; i < size; i++ {
		vec[i] = levyFlight(alpha, beta, rng)
	}
	return vec
}
