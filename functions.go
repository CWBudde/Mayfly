// Benchmark objective functions for the Mayfly Algorithm library.
//
// The suite is algorithm-agnostic: every function is pure math over a
// []float64 position vector, so results are directly comparable with the
// sibling Dragonfly library.
//
// Every single-objective function returns 0 for an empty position vector.
// An empty sum is zero, and the alternatives -- NaN from dividing by a zero
// dimension count, or a panic from indexing an empty slice -- are the wrong
// failure mode for a pure scoring function.

package mayfly

import "math"

// Sphere is the Sphere benchmark function: a smooth, convex, unimodal bowl.
// Global minimum is at f(0, ..., 0) = 0.
func Sphere(x []float64) float64 {
	sum := 0.0
	for _, val := range x {
		sum += val * val
	}

	return sum
}

// Rastrigin is the Rastrigin benchmark function: highly multimodal with a regular lattice of local minima.
// Global minimum is at f(0, ..., 0) = 0.
func Rastrigin(x []float64) float64 {
	n := len(x)
	A := 10.0
	sum := 0.0

	for _, val := range x {
		sum += val*val - A*math.Cos(2*math.Pi*val)
	}

	return float64(n)*A + sum
}

// Rosenbrock is the Rosenbrock benchmark function (the "banana" function): a narrow, curved valley.
// Global minimum is at f(1, ..., 1) = 0.
func Rosenbrock(x []float64) float64 {
	sum := 0.0

	for i := range len(x) - 1 {
		valley := x[i+1] - x[i]*x[i]
		offset := 1 - x[i]
		sum += 100*valley*valley + offset*offset
	}

	return sum
}

// Ackley is the Ackley benchmark function: a nearly flat outer region with a deep central basin.
// Global minimum is at f(0, ..., 0) = 0.
func Ackley(x []float64) float64 {
	// An empty position vector scores 0, per the convention documented at the
	// top of this file.
	if len(x) == 0 {
		return 0
	}

	n := float64(len(x))
	sum1 := 0.0
	sum2 := 0.0

	for _, val := range x {
		sum1 += val * val
		sum2 += math.Cos(2 * math.Pi * val)
	}

	return -20*math.Exp(-0.2*math.Sqrt(sum1/n)) - math.Exp(sum2/n) + 20 + math.E
}

// Griewank is the Griewank benchmark function: a product term creates many regularly spaced local minima.
// Global minimum is at f(0, ..., 0) = 0.
func Griewank(x []float64) float64 {
	sum := 0.0
	prod := 1.0

	for i, val := range x {
		sum += val * val
		prod *= math.Cos(val / math.Sqrt(float64(i+1)))
	}

	return sum/4000 - prod + 1
}

// Schwefel is the Schwefel benchmark function: deceptive, with the global minimum far from the next best local minima.
// Typical bounds: [-500, 500].
func Schwefel(x []float64) float64 {
	n := float64(len(x))

	sum := 0.0
	for _, val := range x {
		sum += val * math.Sin(math.Sqrt(math.Abs(val)))
	}

	return 418.9829*n - sum
}

// Levy is the Levy benchmark function: multimodal with a strongly oscillating surface.
// Typical bounds: [-10, 10].
func Levy(x []float64) float64 {
	// An empty position vector scores 0, per the convention documented at the
	// top of this file.
	if len(x) == 0 {
		return 0
	}

	n := len(x)
	w := make([]float64, n)

	for i := range n {
		w[i] = 1 + (x[i]-1)/4
	}

	sinFirst := math.Sin(math.Pi * w[0])
	term1 := sinFirst * sinFirst

	lastOffset := w[n-1] - 1
	sinLast := math.Sin(2 * math.Pi * w[n-1])
	term3 := lastOffset * lastOffset * (1 + sinLast*sinLast)

	sum := 0.0

	for i := range n - 1 {
		wi := w[i]
		offset := wi - 1
		sinTerm := math.Sin(math.Pi*wi + 1)
		sum += offset * offset * (1 + 10*sinTerm*sinTerm)
	}

	return term1 + sum + term3
}

// Zakharov is the Zakharov benchmark function: unimodal, with no local minima besides the global one.
// Typical bounds: [-5, 10] or [-10, 10].
func Zakharov(x []float64) float64 {
	sum1 := 0.0
	sum2 := 0.0

	for i, val := range x {
		sum1 += val * val
		sum2 += 0.5 * float64(i+1) * val
	}

	sum2Sq := sum2 * sum2

	return sum1 + sum2Sq + sum2Sq*sum2Sq
}

// Michalewicz is the Michalewicz benchmark function: steep valleys and ridges controlled by a steepness parameter.
// Typical bounds: [0, pi].
func Michalewicz(x []float64) float64 {
	m := 10.0
	sum := 0.0

	for i, val := range x {
		sum += math.Sin(val) * math.Pow(math.Sin(float64(i+1)*val*val/math.Pi), 2*m)
	}

	return -sum
}

// DixonPrice is the Dixon-Price benchmark function: a valley-shaped, unimodal landscape.
// Typical bounds: [-10, 10].
func DixonPrice(x []float64) float64 {
	n := len(x)
	if n == 0 {
		return 0
	}

	firstOffset := x[0] - 1
	term1 := firstOffset * firstOffset

	sum := 0.0

	for i := 1; i < n; i++ {
		inner := 2*x[i]*x[i] - x[i-1]
		sum += float64(i+1) * inner * inner
	}

	return term1 + sum
}

// BentCigar is the Bent Cigar benchmark function: unimodal and severely ill-conditioned.
// Typical bounds: [-100, 100].
func BentCigar(x []float64) float64 {
	if len(x) == 0 {
		return 0
	}

	sum := x[0] * x[0]
	for i := 1; i < len(x); i++ {
		sum += 1e6 * x[i] * x[i]
	}

	return sum
}

// Discus is the Discus benchmark function: unimodal and ill-conditioned along a single direction.
// Typical bounds: [-100, 100].
func Discus(x []float64) float64 {
	if len(x) == 0 {
		return 0
	}

	sum := 1e6 * x[0] * x[0]
	for i := 1; i < len(x); i++ {
		sum += x[i] * x[i]
	}

	return sum
}

// Weierstrass is the Weierstrass benchmark function: continuous everywhere but differentiable nowhere.
// Typical bounds: [-0.5, 0.5].
func Weierstrass(x []float64) float64 {
	n := len(x)
	a := 0.5
	b := 3.0
	kmax := 20

	sum := 0.0

	for _, xi := range x {
		innerSum := 0.0

		for k := 0; k <= kmax; k++ {
			ak := math.Pow(a, float64(k))
			bk := math.Pow(b, float64(k))
			innerSum += ak * math.Cos(2*math.Pi*bk*(xi+0.5))
		}

		sum += innerSum
	}

	// Subtract the constant term
	constantSum := 0.0

	for k := 0; k <= kmax; k++ {
		ak := math.Pow(a, float64(k))
		bk := math.Pow(b, float64(k))
		constantSum += ak * math.Cos(2*math.Pi*bk*0.5)
	}

	return sum - float64(n)*constantSum
}

// HappyCat is the HappyCat benchmark function: multimodal with a curved, thin optimal region.
// Typical bounds: [-2, 2].
func HappyCat(x []float64) float64 {
	// An empty position vector scores 0, per the convention documented at the
	// top of this file.
	if len(x) == 0 {
		return 0
	}

	n := float64(len(x))
	alpha := 0.125

	sumSquares := 0.0
	sumValues := 0.0

	for _, val := range x {
		sumSquares += val * val
		sumValues += val
	}

	return math.Pow(math.Abs(sumSquares-n), 2*alpha) + (0.5*sumSquares+sumValues)/n + 0.5
}

// ExpandedSchafferF6 is the Expanded Schaffer F6 benchmark function: multimodal with concentric ripples.
// Typical bounds: [-100, 100].
func ExpandedSchafferF6(x []float64) float64 {
	n := len(x)
	if n < 2 {
		return 0
	}

	schafferF6 := func(x, y float64) float64 {
		sum := x*x + y*y
		numerator := math.Pow(math.Sin(math.Sqrt(sum)), 2) - 0.5
		denominator := (1 + 0.001*sum) * (1 + 0.001*sum)

		return 0.5 + numerator/denominator
	}

	sum := 0.0
	for i := range n - 1 {
		sum += schafferF6(x[i], x[i+1])
	}
	// Close the loop
	sum += schafferF6(x[n-1], x[0])

	return sum
}

// Himmelblau is the Himmelblau benchmark function: multimodal with four equal global
// minima, extended to n dimensions by summing over disjoint coordinate pairs.
// Typical bounds: [-5, 5].
//
// In two dimensions this is the textbook function, with minima of 0 at (3, 2),
// (-2.805118, 3.131312), (-3.779310, -3.283186) and (3.584428, -1.848126).
// An odd dimension leaves one coordinate without a partner; it contributes its
// square, so the minimum stays 0 and Himmelblau([3, 2, 0]) == 0 rather than
// silently ignoring the last coordinate.
func Himmelblau(x []float64) float64 {
	n := len(x)

	sum := 0.0

	for k := 0; k < n-1; k += 2 {
		a, b := x[k], x[k+1]
		first := a*a + b - 11
		second := a + b*b - 7
		sum += first*first + second*second
	}

	if n%2 == 1 {
		sum += x[n-1] * x[n-1]
	}

	return sum
}
