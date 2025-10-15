package mayfly

import "math"

// Sphere is the Sphere benchmark function.
// Global minimum is at f(0, ..., 0) = 0
func Sphere(x []float64) float64 {
	sum := 0.0
	for _, val := range x {
		sum += val * val
	}
	return sum
}

// Rastrigin is the Rastrigin benchmark function.
// Global minimum is at f(0, ..., 0) = 0
func Rastrigin(x []float64) float64 {
	n := len(x)
	A := 10.0
	sum := 0.0
	for _, val := range x {
		sum += val*val - A*math.Cos(2*math.Pi*val)
	}
	return float64(n)*A + sum
}

// Rosenbrock is the Rosenbrock benchmark function (banana function).
// Global minimum is at f(1, ..., 1) = 0
func Rosenbrock(x []float64) float64 {
	sum := 0.0
	for i := 0; i < len(x)-1; i++ {
		sum += 100*math.Pow(x[i+1]-x[i]*x[i], 2) + math.Pow(1-x[i], 2)
	}
	return sum
}

// Ackley is the Ackley benchmark function.
// Global minimum is at f(0, ..., 0) = 0
func Ackley(x []float64) float64 {
	n := float64(len(x))
	sum1 := 0.0
	sum2 := 0.0
	for _, val := range x {
		sum1 += val * val
		sum2 += math.Cos(2 * math.Pi * val)
	}
	return -20*math.Exp(-0.2*math.Sqrt(sum1/n)) - math.Exp(sum2/n) + 20 + math.E
}

// Griewank is the Griewank benchmark function.
// Global minimum is at f(0, ..., 0) = 0
func Griewank(x []float64) float64 {
	sum := 0.0
	prod := 1.0
	for i, val := range x {
		sum += val * val
		prod *= math.Cos(val / math.Sqrt(float64(i+1)))
	}
	return sum/4000 - prod + 1
}

// Schwefel is the Schwefel benchmark function.
// Global minimum is at f(420.9687, ..., 420.9687) = 0
// Typical bounds: [-500, 500]
func Schwefel(x []float64) float64 {
	n := float64(len(x))
	sum := 0.0
	for _, val := range x {
		sum += val * math.Sin(math.Sqrt(math.Abs(val)))
	}
	return 418.9829*n - sum
}

// Levy is the Levy benchmark function.
// Global minimum is at f(1, ..., 1) = 0
// Typical bounds: [-10, 10]
func Levy(x []float64) float64 {
	n := len(x)
	w := make([]float64, n)
	for i := 0; i < n; i++ {
		w[i] = 1 + (x[i]-1)/4
	}

	term1 := math.Pow(math.Sin(math.Pi*w[0]), 2)
	term3 := math.Pow(w[n-1]-1, 2) * (1 + math.Pow(math.Sin(2*math.Pi*w[n-1]), 2))

	sum := 0.0
	for i := 0; i < n-1; i++ {
		wi := w[i]
		sum += math.Pow(wi-1, 2) * (1 + 10*math.Pow(math.Sin(math.Pi*wi+1), 2))
	}

	return term1 + sum + term3
}

// Zakharov is the Zakharov benchmark function.
// Global minimum is at f(0, ..., 0) = 0
// Typical bounds: [-5, 10] or [-10, 10]
func Zakharov(x []float64) float64 {
	sum1 := 0.0
	sum2 := 0.0
	for i, val := range x {
		sum1 += val * val
		sum2 += 0.5 * float64(i+1) * val
	}
	return sum1 + math.Pow(sum2, 2) + math.Pow(sum2, 4)
}

// Michalewicz is the Michalewicz benchmark function.
// Global minimum depends on dimension (e.g., -9.66 for 10D)
// Typical bounds: [0, pi]
func Michalewicz(x []float64) float64 {
	m := 10.0
	sum := 0.0
	for i, val := range x {
		sum += math.Sin(val) * math.Pow(math.Sin(float64(i+1)*val*val/math.Pi), 2*m)
	}
	return -sum
}

// Dixon-Price is the Dixon-Price benchmark function.
// Global minimum is at x_i = 2^(-(2^i - 2)/2^i) with f(x*) = 0
// Typical bounds: [-10, 10]
func DixonPrice(x []float64) float64 {
	n := len(x)
	if n == 0 {
		return 0
	}

	term1 := math.Pow(x[0]-1, 2)
	sum := 0.0
	for i := 1; i < n; i++ {
		sum += float64(i+1) * math.Pow(2*x[i]*x[i]-x[i-1], 2)
	}
	return term1 + sum
}

// Bent Cigar is a CEC-style benchmark function.
// Global minimum is at f(0, ..., 0) = 0
// Typical bounds: [-100, 100]
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

// Discus (Tablet) is a CEC-style benchmark function.
// Global minimum is at f(0, ..., 0) = 0
// Typical bounds: [-100, 100]
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

// Weierstrass is the Weierstrass benchmark function.
// Global minimum is at f(0, ..., 0) = 0
// Typical bounds: [-0.5, 0.5]
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

// HappyCat is the HappyCat benchmark function (CEC-style).
// Global minimum is at f(-1, ..., -1) = 0
// Typical bounds: [-2, 2]
func HappyCat(x []float64) float64 {
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

// Expanded Schaffer F6 is a composite CEC-style function.
// Global minimum is at f(0, ..., 0) = 0
// Typical bounds: [-100, 100]
func ExpandedSchafferF6(x []float64) float64 {
	n := len(x)
	if n < 2 {
		return 0
	}

	schafferF6 := func(x, y float64) float64 {
		sum := x*x + y*y
		numerator := math.Pow(math.Sin(math.Sqrt(sum)), 2) - 0.5
		denominator := math.Pow(1+0.001*sum, 2)
		return 0.5 + numerator/denominator
	}

	sum := 0.0
	for i := 0; i < n-1; i++ {
		sum += schafferF6(x[i], x[i+1])
	}
	// Close the loop
	sum += schafferF6(x[n-1], x[0])

	return sum
}
