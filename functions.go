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
