//nolint:wsl_v5 // Compact stage groupings mirror the official evaluator pipelines.
package mayfly

import "math"

// This file independently ports Jane Jing Liang's CEC13 evaluator as
// distributed with Du and Zhu's DESMA paper: cec13_func.cpp from the paper's
// S1 Data archive (PLOS ONE, 2022, doi:10.1371/journal.pone.0273155). The
// implementation deliberately follows that executable source where its
// behavior is surprising. In particular:
//
//   - shift_data.txt is consumed as sequential shift blocks by the loader;
//   - the asymmetric transform leaves non-positive output coordinates at their
//     prior scratch values instead of copying its input;
//   - Different Powers computes its exponent with integer arithmetic;
//   - the Griewank-Rosenbrock function computes, then discards, its rotation;
//   - the composition functions retain the source's component definitions.
//
// The C evaluator stores its work arrays in mutable globals. Here every
// objective call owns its scratch arrays, making an immutable instance safe for
// concurrent use while preserving scratch-buffer fallback behavior within one
// evaluation (including across composition components).

type cec2013Scratch struct {
	y []float64
	z []float64
}

func newCEC2013Scratch(dimension int) *cec2013Scratch {
	return &cec2013Scratch{
		y: make([]float64, dimension),
		z: make([]float64, dimension),
	}
}

func (instance *cec2013Instance) objective() ObjectiveFunction {
	return func(position []float64) float64 {
		if len(position) != instance.dimension {
			return math.Inf(1)
		}

		scratch := newCEC2013Scratch(instance.dimension)

		return instance.evaluate(position, scratch) + cec2013Bias[instance.function]
	}
}

func (instance *cec2013Instance) evaluate(position []float64, scratch *cec2013Scratch) float64 {
	if instance.function <= 10 {
		return instance.evaluateFirstTen(position, scratch)
	}

	if instance.function <= 20 {
		return instance.evaluateSecondTen(position, scratch)
	}

	return cec2013Composition(instance.function-20, position, instance.shift, instance.rotation, scratch)
}

func (instance *cec2013Instance) evaluateFirstTen(position []float64, scratch *cec2013Scratch) float64 {
	shift := instance.shift
	rotation := instance.rotation

	switch instance.function {
	case 1:
		return cec2013Sphere(position, shift, scratch)
	case 2:
		return cec2013Elliptic(position, shift, rotation, true, scratch)
	case 3:
		return cec2013BentCigar(position, shift, rotation, true, scratch)
	case 4:
		return cec2013Discus(position, shift, rotation, true, scratch)
	case 5:
		return cec2013DifferentPowers(position, shift, rotation, false, scratch)
	case 6:
		return cec2013Rosenbrock(position, shift, rotation, true, scratch)
	case 7:
		return cec2013SchafferF7(position, shift, rotation, true, scratch)
	case 8:
		return cec2013Ackley(position, shift, rotation, true, scratch)
	case 9:
		return cec2013Weierstrass(position, shift, rotation, scratch)
	case 10:
		return cec2013Griewank(position, shift, rotation, true, scratch)
	default:
		return math.Inf(1)
	}
}

func (instance *cec2013Instance) evaluateSecondTen(position []float64, scratch *cec2013Scratch) float64 {
	shift := instance.shift
	rotation := instance.rotation

	switch instance.function {
	case 11:
		return cec2013Rastrigin(position, shift, rotation, false, scratch)
	case 12:
		return cec2013Rastrigin(position, shift, rotation, true, scratch)
	case 13:
		return cec2013StepRastrigin(position, shift, rotation, true, scratch)
	case 14:
		return cec2013Schwefel(position, shift, rotation, false, scratch)
	case 15:
		return cec2013Schwefel(position, shift, rotation, true, scratch)
	case 16:
		return cec2013Katsuura(position, shift, rotation, true, scratch)
	case 17:
		return cec2013Lunacek(position, shift, rotation, false, scratch)
	case 18:
		return cec2013Lunacek(position, shift, rotation, true, scratch)
	case 19:
		return cec2013GriewankRosenbrock(position, shift, rotation, true, scratch)
	case 20:
		return cec2013ExpandedSchaffer6(position, shift, rotation, true, scratch)
	default:
		return math.Inf(1)
	}
}

func cec2013Shift(x, shifted, shift []float64) {
	for i, value := range x {
		shifted[i] = value - shift[i]
	}
}

func cec2013Rotate(x, rotated, matrix []float64) {
	n := len(x)
	for row := range n {
		value := 0.0
		for column, coordinate := range x {
			value += coordinate * matrix[row*n+column]
		}

		rotated[row] = value
	}
}

func cec2013RotateOrCopy(x, out, matrix []float64, rotate bool) {
	if rotate {
		cec2013Rotate(x, out, matrix)
		return
	}

	copy(out, x)
}

// cec2013Asymmetry intentionally writes only positive input coordinates. The
// official C routine leaves every other output coordinate untouched, so the
// caller's previous scratch contents are observable and must remain intact.
func cec2013Asymmetry(x, out []float64, beta float64) {
	denominator := float64(len(x) - 1)
	for i, value := range x {
		if value > 0 {
			// Keep pow(value, 0.5), rather than the equivalent sqrt(value):
			// the reference C evaluator uses pow here and its last-bit result is
			// amplified by the outer power for Ackley and related functions.
			exponent := 1 + beta*float64(i)/denominator*math.Pow(value, 0.5)
			out[i] = math.Pow(value, exponent)
		}
	}
}

func cec2013Oscillatory(x, out []float64) {
	for i, value := range x {
		if i != 0 && i != len(x)-1 {
			out[i] = value
			continue
		}

		if value == 0 {
			out[i] = 0
			continue
		}

		xx := math.Log(math.Abs(value))
		c1, c2 := 5.5, 3.1
		sign := -1.0

		if value > 0 {
			c1, c2 = 10, 7.9
			sign = 1
		}

		out[i] = sign * math.Exp(xx+0.049*(math.Sin(c1*xx)+math.Sin(c2*xx)))
	}
}

func cec2013Sphere(
	x, shift []float64,
	scratch *cec2013Scratch,
) float64 {
	cec2013Shift(x, scratch.y, shift)
	copy(scratch.z, scratch.y)

	result := 0.0
	for _, value := range scratch.z {
		result += value * value
	}

	return result
}

func cec2013Elliptic(
	x, shift, rotation []float64,
	rotate bool,
	scratch *cec2013Scratch,
) float64 {
	cec2013Shift(x, scratch.y, shift)
	cec2013RotateOrCopy(scratch.y, scratch.z, rotation, rotate)
	cec2013Oscillatory(scratch.z, scratch.y)

	result := 0.0
	denominator := float64(len(x) - 1)

	for i, value := range scratch.y {
		result += math.Pow(10, 6*float64(i)/denominator) * value * value
	}

	return result
}

func cec2013BentCigar(
	x, shift, rotation []float64,
	rotate bool,
	scratch *cec2013Scratch,
) float64 {
	n := len(x)
	cec2013Shift(x, scratch.y, shift)
	cec2013RotateOrCopy(scratch.y, scratch.z, rotation, rotate)
	cec2013Asymmetry(scratch.z, scratch.y, 0.5)
	cec2013RotateOrCopy(scratch.y, scratch.z, rotation[n*n:], rotate)

	result := scratch.z[0] * scratch.z[0]
	for _, value := range scratch.z[1:] {
		result += 1e6 * value * value
	}

	return result
}

func cec2013Discus(
	x, shift, rotation []float64,
	rotate bool,
	scratch *cec2013Scratch,
) float64 {
	cec2013Shift(x, scratch.y, shift)
	cec2013RotateOrCopy(scratch.y, scratch.z, rotation, rotate)
	cec2013Oscillatory(scratch.z, scratch.y)

	result := 1e6 * scratch.y[0] * scratch.y[0]
	for _, value := range scratch.y[1:] {
		result += value * value
	}

	return result
}

func cec2013DifferentPowers(
	x, shift, rotation []float64,
	rotate bool,
	scratch *cec2013Scratch,
) float64 {
	cec2013Shift(x, scratch.y, shift)
	cec2013RotateOrCopy(scratch.y, scratch.z, rotation, rotate)

	result := 0.0

	for i, value := range scratch.z {
		// Both operands are int in the official source. Retaining integer
		// division produces the stepwise exponent sequence it evaluates.
		exponent := 2 + 4*i/(len(x)-1)
		result += math.Pow(math.Abs(value), float64(exponent))
	}

	return math.Sqrt(result)
}

func cec2013Rosenbrock(
	x, shift, rotation []float64,
	rotate bool,
	scratch *cec2013Scratch,
) float64 {
	cec2013Shift(x, scratch.y, shift)

	for i := range scratch.y {
		scratch.y[i] *= 2.048 / 100
	}

	cec2013RotateOrCopy(scratch.y, scratch.z, rotation, rotate)

	for i := range scratch.z {
		scratch.z[i]++
	}

	result := 0.0

	for i := 0; i+1 < len(x); i++ {
		a := scratch.z[i]*scratch.z[i] - scratch.z[i+1]
		b := scratch.z[i] - 1
		result += 100*a*a + b*b
	}

	return result
}

func cec2013SchafferF7(
	x, shift, rotation []float64,
	rotate bool,
	scratch *cec2013Scratch,
) float64 {
	n := len(x)
	cec2013Shift(x, scratch.y, shift)
	cec2013RotateOrCopy(scratch.y, scratch.z, rotation, rotate)
	cec2013Asymmetry(scratch.z, scratch.y, 0.5)

	denominator := float64(n - 1)
	for i := range scratch.z {
		scratch.z[i] = scratch.y[i] * math.Pow(10, float64(i)/denominator/2)
	}

	cec2013RotateOrCopy(scratch.z, scratch.y, rotation[n*n:], rotate)

	for i := 0; i+1 < n; i++ {
		scratch.z[i] = math.Hypot(scratch.y[i], scratch.y[i+1])
	}

	result := 0.0

	for _, radius := range scratch.z[:n-1] {
		root := math.Sqrt(radius)
		wave := math.Sin(50 * math.Pow(radius, 0.2))
		result += root + root*wave*wave
	}

	return result * result / denominator / denominator
}

func cec2013Ackley(
	x, shift, rotation []float64,
	rotate bool,
	scratch *cec2013Scratch,
) float64 {
	n := len(x)
	cec2013Shift(x, scratch.y, shift)
	cec2013RotateOrCopy(scratch.y, scratch.z, rotation, rotate)
	cec2013Asymmetry(scratch.z, scratch.y, 0.5)

	denominator := float64(n - 1)
	for i := range scratch.z {
		scratch.z[i] = scratch.y[i] * math.Pow(10, float64(i)/denominator/2)
	}

	cec2013RotateOrCopy(scratch.z, scratch.y, rotation[n*n:], rotate)

	squares, cosines := 0.0, 0.0
	for _, value := range scratch.y {
		squares += value * value
		cosines += math.Cos(2 * math.Pi * value)
	}

	count := float64(n)

	return math.E - 20*math.Exp(-0.2*math.Sqrt(squares/count)) - math.Exp(cosines/count) + 20
}

func cec2013Weierstrass(
	x, shift, rotation []float64,
	scratch *cec2013Scratch,
) float64 {
	n := len(x)
	cec2013Shift(x, scratch.y, shift)

	for i := range scratch.y {
		scratch.y[i] *= 0.5 / 100
	}

	cec2013Rotate(scratch.y, scratch.z, rotation)
	cec2013Asymmetry(scratch.z, scratch.y, 0.5)

	denominator := float64(n - 1)
	for i := range scratch.z {
		scratch.z[i] = scratch.y[i] * math.Pow(10, float64(i)/denominator/2)
	}

	cec2013Rotate(scratch.z, scratch.y, rotation[n*n:])

	result, constant := 0.0, 0.0
	for _, value := range scratch.y {
		constant = 0

		for k := range 21 {
			a := math.Pow(0.5, float64(k))
			b := math.Pow(3, float64(k))
			result += a * math.Cos(2*math.Pi*b*(value+0.5))
			constant += a * math.Cos(2*math.Pi*b*0.5)
		}
	}

	return result - float64(n)*constant
}

func cec2013Griewank(
	x, shift, rotation []float64,
	rotate bool,
	scratch *cec2013Scratch,
) float64 {
	n := len(x)
	cec2013Shift(x, scratch.y, shift)

	for i := range scratch.y {
		scratch.y[i] *= 6
	}

	cec2013RotateOrCopy(scratch.y, scratch.z, rotation, rotate)

	denominator := float64(n - 1)

	for i := range scratch.z {
		scratch.z[i] *= math.Pow(100, float64(i)/denominator/2)
	}

	squares, product := 0.0, 1.0
	for i, value := range scratch.z {
		squares += value * value
		product *= math.Cos(value / math.Sqrt(float64(i+1)))
	}

	return 1 + squares/4000 - product
}

func cec2013Rastrigin(
	x, shift, rotation []float64,
	rotate bool,
	scratch *cec2013Scratch,
) float64 {
	n := len(x)
	cec2013Shift(x, scratch.y, shift)

	for i := range scratch.y {
		scratch.y[i] *= 5.12 / 100
	}

	cec2013RotateOrCopy(scratch.y, scratch.z, rotation, rotate)
	cec2013Oscillatory(scratch.z, scratch.y)
	cec2013Asymmetry(scratch.y, scratch.z, 0.2)
	cec2013RotateOrCopy(scratch.z, scratch.y, rotation[n*n:], rotate)

	denominator := float64(n - 1)
	for i := range scratch.y {
		scratch.y[i] *= math.Pow(10, float64(i)/denominator/2)
	}

	cec2013RotateOrCopy(scratch.y, scratch.z, rotation, rotate)

	result := 0.0
	for _, value := range scratch.z {
		result += value*value - 10*math.Cos(2*math.Pi*value) + 10
	}

	return result
}

func cec2013StepRastrigin(
	x, shift, rotation []float64,
	rotate bool,
	scratch *cec2013Scratch,
) float64 {
	n := len(x)
	cec2013Shift(x, scratch.y, shift)

	for i := range scratch.y {
		scratch.y[i] *= 5.12 / 100
	}

	cec2013RotateOrCopy(scratch.y, scratch.z, rotation, rotate)
	for i, value := range scratch.z {
		if math.Abs(value) > 0.5 {
			scratch.z[i] = math.Floor(2*value+0.5) / 2
		}
	}

	cec2013Oscillatory(scratch.z, scratch.y)
	cec2013Asymmetry(scratch.y, scratch.z, 0.2)
	cec2013RotateOrCopy(scratch.z, scratch.y, rotation[n*n:], rotate)

	denominator := float64(n - 1)
	for i := range scratch.y {
		scratch.y[i] *= math.Pow(10, float64(i)/denominator/2)
	}

	cec2013RotateOrCopy(scratch.y, scratch.z, rotation, rotate)

	result := 0.0
	for _, value := range scratch.z {
		result += value*value - 10*math.Cos(2*math.Pi*value) + 10
	}

	return result
}

func cec2013Schwefel(
	x, shift, rotation []float64,
	rotate bool,
	scratch *cec2013Scratch,
) float64 {
	n := len(x)
	cec2013Shift(x, scratch.y, shift)

	for i := range scratch.y {
		scratch.y[i] *= 10
	}

	cec2013RotateOrCopy(scratch.y, scratch.z, rotation, rotate)
	denominator := float64(n - 1)
	for i := range scratch.y {
		scratch.y[i] = scratch.z[i] * math.Pow(10, float64(i)/denominator/2)
		scratch.z[i] = scratch.y[i] + 420.9687462275036
	}

	result := 0.0

	for _, value := range scratch.z {
		switch {
		case value > 500:
			remainder := math.Mod(value, 500)
			reflected := 500 - remainder
			result -= reflected * math.Sin(math.Sqrt(reflected))
			offset := (value - 500) / 100
			result += offset * offset / float64(n)
		case value < -500:
			remainder := math.Mod(math.Abs(value), 500)
			reflected := -500 + remainder
			result -= reflected * math.Sin(math.Sqrt(500-remainder))
			offset := (value + 500) / 100
			result += offset * offset / float64(n)
		default:
			result -= value * math.Sin(math.Sqrt(math.Abs(value)))
		}
	}

	return 418.9828872724338*float64(n) + result
}

func cec2013Katsuura(
	x, shift, rotation []float64,
	rotate bool,
	scratch *cec2013Scratch,
) float64 {
	n := len(x)
	cec2013Shift(x, scratch.y, shift)
	for i := range scratch.y {
		scratch.y[i] *= 5.0 / 100
	}

	cec2013RotateOrCopy(scratch.y, scratch.z, rotation, rotate)
	denominator := float64(n - 1)

	for i := range scratch.z {
		scratch.z[i] *= math.Pow(100, float64(i)/denominator/2)
	}

	cec2013RotateOrCopy(scratch.z, scratch.y, rotation[n*n:], rotate)
	power := 10 / math.Pow(float64(n), 1.2)
	product := 1.0

	for i, value := range scratch.y {
		sum := 0.0

		for j := 1; j <= 32; j++ {
			scale := math.Ldexp(1, j)
			sum += math.Abs(scale*value-math.Floor(scale*value+0.5)) / scale
		}

		product *= math.Pow(1+float64(i+1)*sum, power)
	}

	factor := 10 / float64(n*n)

	return product*factor - factor
}

func cec2013Lunacek(
	x, shift, rotation []float64,
	rotate bool,
	scratch *cec2013Scratch,
) float64 {
	n := len(x)
	mu0 := 2.5
	s := 1 - 1/(2*math.Sqrt(float64(n)+20)-8.2)
	mu1 := -math.Sqrt((mu0*mu0 - 1) / s)
	tmpx := make([]float64, n)

	cec2013Shift(x, scratch.y, shift)
	for i := range scratch.y {
		scratch.y[i] *= 0.1

		tmpx[i] = 2 * scratch.y[i]
		if shift[i] < 0 {
			tmpx[i] = -tmpx[i]
		}

		scratch.z[i] = tmpx[i]
		tmpx[i] += mu0
	}

	cec2013RotateOrCopy(scratch.z, scratch.y, rotation, rotate)

	denominator := float64(n - 1)

	for i := range scratch.y {
		scratch.y[i] *= math.Pow(100, float64(i)/denominator/2)
	}

	cec2013RotateOrCopy(scratch.y, scratch.z, rotation[n*n:], rotate)
	first, second := 0.0, 0.0

	for _, value := range tmpx {
		firstOffset := value - mu0
		first += firstOffset * firstOffset
		secondOffset := value - mu1
		second += secondOffset * secondOffset
	}

	second = second*s + float64(n)
	cosines := 0.0

	for _, value := range scratch.z {
		cosines += math.Cos(2 * math.Pi * value)
	}

	return min(first, second) + 10*(float64(n)-cosines)
}

func cec2013GriewankRosenbrock(
	x, shift, rotation []float64,
	rotate bool,
	scratch *cec2013Scratch,
) float64 {
	cec2013Shift(x, scratch.y, shift)

	for i := range scratch.y {
		scratch.y[i] *= 5.0 / 100
	}

	// The official source computes this rotation into z, then immediately
	// overwrites z from the unrotated y buffer. Preserve that discarded work.
	cec2013RotateOrCopy(scratch.y, scratch.z, rotation, rotate)

	for i := range scratch.z {
		scratch.z[i] = scratch.y[i] + 1
	}

	result := 0.0

	for i := range scratch.z {
		next := scratch.z[(i+1)%len(scratch.z)]
		a := scratch.z[i]*scratch.z[i] - next
		b := scratch.z[i] - 1
		rosenbrock := 100*a*a + b*b
		result += rosenbrock*rosenbrock/4000 - math.Cos(rosenbrock) + 1
	}

	return result
}

func cec2013ExpandedSchaffer6(
	x, shift, rotation []float64,
	rotate bool,
	scratch *cec2013Scratch,
) float64 {
	n := len(x)
	cec2013Shift(x, scratch.y, shift)
	cec2013RotateOrCopy(scratch.y, scratch.z, rotation, rotate)
	cec2013Asymmetry(scratch.z, scratch.y, 0.5)
	cec2013RotateOrCopy(scratch.y, scratch.z, rotation[n*n:], rotate)

	result := 0.0

	for i, value := range scratch.z {
		next := scratch.z[(i+1)%n]
		radiusSquared := value*value + next*next
		sine := math.Sin(math.Sqrt(radiusSquared))
		denominator := 1 + 0.001*radiusSquared
		result += 0.5 + (sine*sine-0.5)/(denominator*denominator)
	}

	return result
}

func cec2013Composition(
	number int,
	x, shift, rotation []float64,
	scratch *cec2013Scratch,
) float64 {
	switch number {
	case 1:
		return cec2013Composition1(x, shift, rotation, scratch)
	case 2:
		return cec2013Composition2(x, shift, rotation, false, scratch)
	case 3:
		return cec2013Composition2(x, shift, rotation, true, scratch)
	case 4:
		return cec2013Composition4(x, shift, rotation, []float64{20, 20, 20}, scratch)
	case 5:
		return cec2013Composition4(x, shift, rotation, []float64{10, 30, 50}, scratch)
	case 6:
		return cec2013Composition6(x, shift, rotation, scratch)
	case 7:
		return cec2013Composition7(x, shift, rotation, scratch)
	case 8:
		return cec2013Composition8(x, shift, rotation, scratch)
	default:
		return math.Inf(1)
	}
}

func cec2013ComponentData(shift, rotation []float64, component, dimension int) ([]float64, []float64) {
	return shift[component*dimension:], rotation[component*dimension*dimension:]
}

func cec2013Composition1(x, shift, rotation []float64, scratch *cec2013Scratch) float64 {
	n := len(x)
	fit := make([]float64, 5)

	componentShift, componentRotation := cec2013ComponentData(shift, rotation, 0, n)
	fit[0] = cec2013Rosenbrock(x, componentShift, componentRotation, true, scratch)
	componentShift, componentRotation = cec2013ComponentData(shift, rotation, 1, n)
	fit[1] = cec2013DifferentPowers(x, componentShift, componentRotation, true, scratch) * 1e4 / 1e10
	componentShift, componentRotation = cec2013ComponentData(shift, rotation, 2, n)
	fit[2] = cec2013BentCigar(x, componentShift, componentRotation, true, scratch) * 1e4 / 1e30
	componentShift, componentRotation = cec2013ComponentData(shift, rotation, 3, n)
	fit[3] = cec2013Discus(x, componentShift, componentRotation, true, scratch) * 1e4 / 1e10
	componentShift, _ = cec2013ComponentData(shift, rotation, 4, n)
	fit[4] = cec2013Sphere(x, componentShift, scratch) * 1e4 / 1e5

	return cec2013CompositionValue(
		x, shift,
		[]float64{10, 20, 30, 40, 50},
		[]float64{0, 100, 200, 300, 400},
		fit,
	)
}

func cec2013Composition2(
	x, shift, rotation []float64,
	rotate bool,
	scratch *cec2013Scratch,
) float64 {
	n := len(x)
	fit := make([]float64, 3)

	for component := range fit {
		componentShift, componentRotation := cec2013ComponentData(shift, rotation, component, n)
		fit[component] = cec2013Schwefel(x, componentShift, componentRotation, rotate, scratch)
	}

	return cec2013CompositionValue(
		x, shift,
		[]float64{20, 20, 20},
		[]float64{0, 100, 200},
		fit,
	)
}

func cec2013Composition4(
	x, shift, rotation, deltas []float64,
	scratch *cec2013Scratch,
) float64 {
	n := len(x)
	fit := make([]float64, 3)
	componentShift, componentRotation := cec2013ComponentData(shift, rotation, 0, n)
	fit[0] = cec2013Schwefel(x, componentShift, componentRotation, true, scratch) * 1e3 / 4e3
	componentShift, componentRotation = cec2013ComponentData(shift, rotation, 1, n)
	fit[1] = cec2013Rastrigin(x, componentShift, componentRotation, true, scratch) * 1e3 / 1e3
	componentShift, componentRotation = cec2013ComponentData(shift, rotation, 2, n)
	fit[2] = cec2013Weierstrass(x, componentShift, componentRotation, scratch) * 1e3 / 400

	return cec2013CompositionValue(x, shift, deltas, []float64{0, 100, 200}, fit)
}

func cec2013Composition6(x, shift, rotation []float64, scratch *cec2013Scratch) float64 {
	n := len(x)
	fit := make([]float64, 5)
	componentShift, componentRotation := cec2013ComponentData(shift, rotation, 0, n)
	fit[0] = cec2013Schwefel(x, componentShift, componentRotation, true, scratch) * 1e3 / 4e3
	componentShift, componentRotation = cec2013ComponentData(shift, rotation, 1, n)
	fit[1] = cec2013Rastrigin(x, componentShift, componentRotation, true, scratch) * 1e3 / 1e3
	componentShift, componentRotation = cec2013ComponentData(shift, rotation, 2, n)
	fit[2] = cec2013Elliptic(x, componentShift, componentRotation, true, scratch) * 1e3 / 1e10
	componentShift, componentRotation = cec2013ComponentData(shift, rotation, 3, n)
	fit[3] = cec2013Weierstrass(x, componentShift, componentRotation, scratch) * 1e3 / 400
	componentShift, _ = cec2013ComponentData(shift, rotation, 4, n)
	fit[4] = cec2013Griewank(x, componentShift, componentRotation, true, scratch) * 1e3 / 100

	return cec2013CompositionValue(
		x, shift,
		[]float64{10, 10, 10, 10, 10},
		[]float64{0, 100, 200, 300, 400},
		fit,
	)
}

func cec2013Composition7(x, shift, rotation []float64, scratch *cec2013Scratch) float64 {
	n := len(x)
	fit := make([]float64, 5)
	componentShift, componentRotation := cec2013ComponentData(shift, rotation, 0, n)
	fit[0] = cec2013Griewank(x, componentShift, componentRotation, true, scratch) * 1e4 / 100
	componentShift, componentRotation = cec2013ComponentData(shift, rotation, 1, n)
	fit[1] = cec2013Rastrigin(x, componentShift, componentRotation, true, scratch) * 1e4 / 1e3
	componentShift, componentRotation = cec2013ComponentData(shift, rotation, 2, n)
	fit[2] = cec2013Schwefel(x, componentShift, componentRotation, true, scratch) * 1e4 / 4e3
	componentShift, componentRotation = cec2013ComponentData(shift, rotation, 3, n)
	fit[3] = cec2013Weierstrass(x, componentShift, componentRotation, scratch) * 1e4 / 400
	componentShift, _ = cec2013ComponentData(shift, rotation, 4, n)
	fit[4] = cec2013Sphere(x, componentShift, scratch) * 1e4 / 1e5

	return cec2013CompositionValue(
		x, shift,
		[]float64{10, 10, 10, 20, 20},
		[]float64{0, 100, 200, 300, 400},
		fit,
	)
}

func cec2013Composition8(x, shift, rotation []float64, scratch *cec2013Scratch) float64 {
	n := len(x)
	fit := make([]float64, 5)
	componentShift, componentRotation := cec2013ComponentData(shift, rotation, 0, n)
	fit[0] = cec2013GriewankRosenbrock(x, componentShift, componentRotation, true, scratch) * 1e4 / 4e3
	componentShift, componentRotation = cec2013ComponentData(shift, rotation, 1, n)
	fit[1] = cec2013SchafferF7(x, componentShift, componentRotation, true, scratch) * 1e4 / 4e6
	componentShift, componentRotation = cec2013ComponentData(shift, rotation, 2, n)
	fit[2] = cec2013Schwefel(x, componentShift, componentRotation, true, scratch) * 1e4 / 4e3
	componentShift, componentRotation = cec2013ComponentData(shift, rotation, 3, n)
	fit[3] = cec2013ExpandedSchaffer6(x, componentShift, componentRotation, true, scratch) * 1e4 / 2e7
	componentShift, _ = cec2013ComponentData(shift, rotation, 4, n)
	fit[4] = cec2013Sphere(x, componentShift, scratch) * 1e4 / 1e5

	return cec2013CompositionValue(
		x, shift,
		[]float64{10, 20, 30, 40, 50},
		[]float64{0, 100, 200, 300, 400},
		fit,
	)
}

func cec2013CompositionValue(x, shift, deltas, biases, fit []float64) float64 {
	components := len(fit)
	n := len(x)
	weights := make([]float64, components)
	weightSum, maximumWeight := 0.0, 0.0

	for component := range components {
		fit[component] += biases[component]
		distance := 0.0

		componentShift := shift[component*n : (component+1)*n]
		for i, value := range x {
			offset := value - componentShift[i]
			distance += offset * offset
		}

		if distance == 0 {
			weights[component] = 1e99
		} else {
			weights[component] = math.Sqrt(1/distance) *
				math.Exp(-distance/(2*float64(n)*deltas[component]*deltas[component]))
		}

		maximumWeight = max(maximumWeight, weights[component])
		weightSum += weights[component]
	}

	if maximumWeight == 0 {
		weightSum = float64(components)

		for i := range weights {
			weights[i] = 1
		}
	}

	result := 0.0
	for i, weight := range weights {
		result += weight / weightSum * fit[i]
	}

	return result
}
