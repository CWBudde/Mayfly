package mayfly

import "math"

type cecBase int

const (
	cecSphere cecBase = iota
	cecElliptic
	cecBentCigar
	cecDiscus
	cecRosenbrock
	cecSchafferF7
	cecAckley
	cecRastrigin
	cecWeierstrass
	cecGriewank
	cecSchwefel
	cecKatsuura
	cecLunacek
	cecGriewankRosenbrock
	cecExpandedSchaffer6
	cecNonContinuousRastrigin
	cecHappyCat
	cecHGBat
	cecZakharov
	cecLevy
)

func (instance *cecInstance) objective() ObjectiveFunction {
	return func(position []float64) float64 {
		if len(position) != instance.dimension {
			return math.Inf(1)
		}

		var value float64
		if instance.year == 2017 {
			value = instance.evaluate2017(position)
			value += float64(instance.function * 100)
		} else {
			value = instance.evaluate2020(position)
			value += cec2020Bias[instance.function]
		}

		return value
	}
}

func (instance *cecInstance) evaluate2017(x []float64) float64 {
	switch instance.function {
	case 1:
		return cecEvaluateBase(cecBentCigar, x, instance.shift, instance.rotation)
	case 3:
		return cecEvaluateBase(cecZakharov, x, instance.shift, instance.rotation)
	case 4:
		return cecEvaluateBase(cecRosenbrock, x, instance.shift, instance.rotation)
	case 5:
		return cecEvaluateBase(cecRastrigin, x, instance.shift, instance.rotation)
	case 6:
		return cecEvaluateBase(cecSchafferF7, x, instance.shift, instance.rotation)
	case 7:
		return cecEvaluateBase(cecLunacek, x, instance.shift, instance.rotation)
	case 8:
		return cecEvaluateBase(cecNonContinuousRastrigin, x, instance.shift, instance.rotation)
	case 9:
		return cecEvaluateBase(cecLevy, x, instance.shift, instance.rotation)
	case 10:
		return cecEvaluateBase(cecSchwefel, x, instance.shift, instance.rotation)
	case 11, 12, 13, 14, 15, 16, 17, 18, 19, 20:
		return cecHybrid2017(instance.function-10, x, instance.shift[:instance.dimension],
			instance.rotation[:instance.dimension*instance.dimension], instance.shuffle[:instance.dimension])
	case 21, 22, 23, 24, 25, 26, 27, 28, 29, 30:
		return instance.composition2017(instance.function-20, x)
	default:
		return math.Inf(1)
	}
}

func (instance *cecInstance) evaluate2020(x []float64) float64 {
	switch instance.function {
	case 1:
		return cecEvaluateBase(cecBentCigar, x, instance.shift, instance.rotation)
	case 2:
		return cecEvaluateBase(cecSchwefel, x, instance.shift, instance.rotation)
	case 3:
		return cecEvaluateBase(cecLunacek, x, instance.shift, instance.rotation)
	case 4:
		return cecEvaluateBase(cecGriewankRosenbrock, x, instance.shift, instance.rotation)
	case 5:
		return cecHybrid2020(1, x, instance.shift, instance.rotation, instance.shuffle)
	case 6:
		return cecHybrid2020(6, x, instance.shift, instance.rotation, instance.shuffle)
	case 7:
		return cecHybrid2020(5, x, instance.shift, instance.rotation, instance.shuffle)
	case 8:
		return instance.composition2017(2, x)
	case 9:
		return instance.composition2017(4, x)
	case 10:
		return instance.composition2017(5, x)
	default:
		return math.Inf(1)
	}
}

func cecEvaluateBase(kind cecBase, x, shift, rotation []float64) float64 {
	if len(x) == 0 {
		return 0
	}

	if kind == cecNonContinuousRastrigin {
		// The released evaluator computes a rounded scratch vector and then
		// immediately transforms x again. Reproduce that observable behavior:
		// CEC comparisons depend on the executable, even where its prose calls
		// this function non-continuous.
		z := cecTransform(x, shift, rotation, 5.12/100)
		return cecRastriginValue(z)
	}

	if kind == cecLunacek {
		return cecLunacekValue(x, shift, rotation, true)
	}

	rate := cecTransformRate(kind)

	z := cecTransform(x, shift, rotation, rate)
	if kind == cecSchafferF7 && len(rotation) == len(x)*len(x) {
		// This intentionally follows the official evaluator, which computes
		// the rotated vector but applies Schaffer F7 to the pre-rotation scratch.
		z = cecTransform(x, shift, nil, rate)
	}

	return cecEvaluateTransformed(kind, z)
}

func cecTransformRate(kind cecBase) float64 {
	switch kind {
	case cecRosenbrock:
		return 2.048 * 0.01
	case cecRastrigin:
		return 5.12 * 0.01
	case cecWeierstrass:
		return 0.5 * 0.01
	case cecGriewank:
		return 600.0 * 0.01
	case cecSchwefel:
		return 1000.0 * 0.01
	case cecKatsuura, cecGriewankRosenbrock, cecHappyCat, cecHGBat:
		return 5.0 * 0.01
	default:
		return 1
	}
}

func cecEvaluateTransformed(kind cecBase, z []float64) float64 {
	switch kind {
	case cecSphere:
		return cecSphereValue(z)
	case cecElliptic:
		return cecEllipticValue(z)
	case cecBentCigar:
		return cecBentCigarValue(z)
	case cecDiscus:
		return cecDiscusValue(z)
	case cecRosenbrock:
		return cecRosenbrockValue(z)
	case cecSchafferF7:
		return cecSchafferF7Value(z)
	case cecAckley:
		return cecAckleyValue(z)
	case cecRastrigin:
		return cecRastriginValue(z)
	case cecWeierstrass:
		return cecWeierstrassValue(z)
	case cecGriewank:
		return cecGriewankValue(z)
	case cecSchwefel:
		return cecSchwefelValue(z)
	case cecKatsuura:
		return cecKatsuuraValue(z)
	case cecGriewankRosenbrock:
		return cecGriewankRosenbrockValue(z)
	case cecExpandedSchaffer6:
		return cecExpandedSchaffer6Value(z)
	case cecHappyCat:
		return cecHappyCatValue(z)
	case cecHGBat:
		return cecHGBatValue(z)
	case cecZakharov:
		return cecZakharovValue(z)
	case cecLevy:
		return cecLevyValue(z)
	default:
		return math.Inf(1)
	}
}

func cecTransform(x, shift, rotation []float64, rate float64) []float64 {
	n := len(x)

	scaled := make([]float64, n)
	for i, value := range x {
		if len(shift) == n {
			value -= shift[i]
		}

		scaled[i] = value * rate
	}

	if len(rotation) != n*n {
		return scaled
	}

	rotated := make([]float64, n)
	for row := range n {
		for column, value := range scaled {
			rotated[row] += rotation[row*n+column] * value
		}
	}

	return rotated
}

func cecSphereValue(x []float64) float64 {
	result := 0.0
	for _, value := range x {
		result += value * value
	}

	return result
}

func cecEllipticValue(x []float64) float64 {
	if len(x) == 1 {
		return x[0] * x[0]
	}

	result := 0.0
	for i, value := range x {
		result += math.Pow(1e6, float64(i)/float64(len(x)-1)) * value * value
	}

	return result
}

func cecBentCigarValue(x []float64) float64 {
	result := x[0] * x[0]
	for _, value := range x[1:] {
		result += 1e6 * value * value
	}

	return result
}

func cecDiscusValue(x []float64) float64 {
	result := 1e6 * x[0] * x[0]
	for _, value := range x[1:] {
		result += value * value
	}

	return result
}

func cecRosenbrockValue(x []float64) float64 {
	z := append([]float64(nil), x...)
	for i := range z {
		z[i]++
	}

	result := 0.0

	for i := 0; i+1 < len(z); i++ {
		a := z[i]*z[i] - z[i+1]
		b := z[i] - 1
		result += 100*a*a + b*b
	}

	return result
}

func cecSchafferF7Value(x []float64) float64 {
	if len(x) < 2 {
		return 0
	}

	result := 0.0

	for i := 0; i+1 < len(x); i++ {
		radius := math.Hypot(x[i], x[i+1])
		root := math.Sqrt(radius)
		wave := math.Sin(50 * math.Pow(radius, 0.2))
		result += root * (1 + wave*wave)
	}

	denominator := float64(len(x) - 1)

	return result * result / (denominator * denominator)
}

func cecAckleyValue(x []float64) float64 {
	squares, cosines := 0.0, 0.0
	for _, value := range x {
		squares += value * value
		cosines += math.Cos(2 * math.Pi * value)
	}

	n := float64(len(x))

	return math.E - 20*math.Exp(-0.2*math.Sqrt(squares/n)) - math.Exp(cosines/n) + 20
}

func cecRastriginValue(x []float64) float64 {
	result := 0.0
	for _, value := range x {
		result += value*value - 10*math.Cos(2*math.Pi*value) + 10
	}

	return result
}

func cecWeierstrassValue(x []float64) float64 {
	constant := 0.0
	for k := range 21 {
		constant += math.Pow(0.5, float64(k)) * math.Cos(2*math.Pi*math.Pow(3, float64(k))*0.5)
	}

	result := 0.0

	for _, value := range x {
		for k := range 21 {
			result += math.Pow(0.5, float64(k)) * math.Cos(2*math.Pi*math.Pow(3, float64(k))*(value+0.5))
		}
	}

	return result - float64(len(x))*constant
}

func cecGriewankValue(x []float64) float64 {
	sum, product := 0.0, 1.0
	for i, value := range x {
		sum += value * value
		product *= math.Cos(value / math.Sqrt(float64(i+1)))
	}

	return 1 + sum/4000 - product
}

func cecSchwefelValue(x []float64) float64 {
	result := 418.9828872724338 * float64(len(x))
	for _, value := range x {
		z := value + 420.9687462275036
		switch {
		case z > 500:
			wrapped := 500 - math.Mod(z, 500)
			result -= wrapped * math.Sin(math.Sqrt(math.Abs(wrapped)))
			offset := (z - 500) * 0.01
			result += offset * offset / float64(len(x))
		case z < -500:
			wrapped := -500 + math.Mod(math.Abs(z), 500)
			result -= wrapped * math.Sin(math.Sqrt(math.Abs(500-math.Mod(math.Abs(z), 500))))
			offset := (z + 500) * 0.01
			result += offset * offset / float64(len(x))
		default:
			result -= z * math.Sin(math.Sqrt(math.Abs(z)))
		}
	}

	return result
}

func cecKatsuuraValue(x []float64) float64 {
	product := 1.0

	power := 10 / math.Pow(float64(len(x)), 1.2)
	for i, value := range x {
		sum := 0.0

		for j := 1; j <= 32; j++ {
			scale := math.Ldexp(1, j)
			sum += math.Abs(scale*value-math.Floor(scale*value+0.5)) / scale
		}

		product *= math.Pow(1+float64(i+1)*sum, power)
	}

	factor := 10 / float64(len(x)*len(x))

	return factor*product - factor
}

func cecLunacekValue(x, shift, rotation []float64, applyShift bool) float64 {
	n := len(x)
	mu0 := 2.5
	s := 1 - 1/(2*math.Sqrt(float64(n)+20)-8.2)
	mu1 := -math.Sqrt((mu0*mu0 - 1) / s)
	z := make([]float64, n)

	shifted := make([]float64, n)
	for i, value := range x {
		if applyShift && len(shift) == n {
			value -= shift[i]
		}

		value *= 0.1

		value *= 2
		if len(shift) == n && shift[i] < 0 {
			value = -value
		}

		z[i] = value
		shifted[i] = value + mu0
	}

	first, second := 0.0, 0.0
	for _, value := range shifted {
		first += (value - mu0) * (value - mu0)
		second += (value - mu1) * (value - mu1)
	}

	second = s*second + float64(n)

	rotated := z
	if len(rotation) == n*n {
		rotated = cecTransform(z, nil, rotation, 1)
	}

	cosines := 0.0
	for _, value := range rotated {
		cosines += math.Cos(2 * math.Pi * value)
	}

	return min(first, second) + 10*(float64(n)-cosines)
}

func cecGriewankRosenbrockValue(x []float64) float64 {
	z := append([]float64(nil), x...)
	for i := range z {
		z[i]++
	}

	result := 0.0

	for i := range z {
		next := z[(i+1)%len(z)]
		a := z[i]*z[i] - next
		b := z[i] - 1
		rosenbrock := 100*a*a + b*b
		result += rosenbrock*rosenbrock/4000 - math.Cos(rosenbrock) + 1
	}

	return result
}

func cecExpandedSchaffer6Value(x []float64) float64 {
	result := 0.0

	for i := range x {
		a, b := x[i], x[(i+1)%len(x)]
		radiusSquared := a*a + b*b
		sine := math.Sin(math.Sqrt(radiusSquared))
		denominator := 1 + 0.001*radiusSquared
		result += 0.5 + (sine*sine-0.5)/(denominator*denominator)
	}

	return result
}

func cecHappyCatValue(x []float64) float64 {
	radius, sum := 0.0, 0.0

	for _, value := range x {
		value--
		radius += value * value
		sum += value
	}

	n := float64(len(x))

	return math.Pow(math.Abs(radius-n), 0.25) + (0.5*radius+sum)/n + 0.5
}

func cecHGBatValue(x []float64) float64 {
	radius, sum := 0.0, 0.0

	for _, value := range x {
		value--
		radius += value * value
		sum += value
	}

	n := float64(len(x))

	return math.Sqrt(math.Abs(radius*radius-sum*sum)) + (0.5*radius+sum)/n + 0.5
}

func cecZakharovValue(x []float64) float64 {
	squares, weighted := 0.0, 0.0
	for i, value := range x {
		squares += value * value
		weighted += 0.5 * float64(i+1) * value
	}

	weightedSquared := weighted * weighted

	return squares + weightedSquared + weightedSquared*weightedSquared
}

func cecLevyValue(x []float64) float64 {
	w := make([]float64, len(x))
	for i, value := range x {
		w[i] = 1 + (value-1)/4
	}

	first := math.Sin(math.Pi * w[0])
	lastOffset := w[len(w)-1] - 1
	lastSine := math.Sin(2 * math.Pi * w[len(w)-1])
	result := first*first + lastOffset*lastOffset*(1+lastSine*lastSine)

	for _, value := range w[:len(w)-1] {
		offset := value - 1
		sine := math.Sin(math.Pi*value + 1)
		result += offset * offset * (1 + 10*sine*sine)
	}

	return result
}

type cecHybridSpec struct {
	proportions []float64
	functions   []cecBase
	rightSized  bool
}

var cec2017Hybrids = map[int]cecHybridSpec{
	1: {[]float64{0.2, 0.4, 0.4}, []cecBase{cecZakharov, cecRosenbrock, cecRastrigin}, false},
	2: {[]float64{0.3, 0.3, 0.4}, []cecBase{cecElliptic, cecSchwefel, cecBentCigar}, false},
	3: {[]float64{0.3, 0.3, 0.4}, []cecBase{cecBentCigar, cecRosenbrock, cecLunacek}, false},
	4: {[]float64{0.2, 0.2, 0.2, 0.4}, []cecBase{cecElliptic, cecAckley, cecSchafferF7, cecRastrigin}, false},
	5: {[]float64{0.2, 0.2, 0.3, 0.3}, []cecBase{cecBentCigar, cecHGBat, cecRastrigin, cecRosenbrock}, false},
	6: {[]float64{0.2, 0.2, 0.3, 0.3}, []cecBase{cecExpandedSchaffer6, cecHGBat, cecRosenbrock, cecSchwefel}, false},
	7: {
		[]float64{0.1, 0.2, 0.2, 0.2, 0.3},
		[]cecBase{cecKatsuura, cecAckley, cecGriewankRosenbrock, cecSchwefel, cecRastrigin},
		false,
	},
	8: {[]float64{0.2, 0.2, 0.2, 0.2, 0.2}, []cecBase{cecElliptic, cecAckley, cecRastrigin, cecHGBat, cecDiscus}, false},
	9: {
		[]float64{0.2, 0.2, 0.2, 0.2, 0.2},
		[]cecBase{cecBentCigar, cecRastrigin, cecGriewankRosenbrock, cecWeierstrass, cecExpandedSchaffer6},
		false,
	},
	10: {
		[]float64{0.1, 0.1, 0.2, 0.2, 0.2, 0.2},
		[]cecBase{cecHGBat, cecKatsuura, cecAckley, cecRastrigin, cecSchwefel, cecSchafferF7},
		false,
	},
}

var cec2020Hybrids = map[int]cecHybridSpec{
	1: {[]float64{0.3, 0.3, 0.4}, []cecBase{cecSchwefel, cecRastrigin, cecElliptic}, true},
	5: {
		[]float64{0.1, 0.2, 0.2, 0.2, 0.3},
		[]cecBase{cecExpandedSchaffer6, cecHGBat, cecRosenbrock, cecSchwefel, cecElliptic},
		true,
	},
	6: {[]float64{0.2, 0.2, 0.3, 0.3}, []cecBase{cecExpandedSchaffer6, cecHGBat, cecRosenbrock, cecSchwefel}, true},
}

func cecHybrid2017(number int, x, shift, rotation []float64, shuffle []int) float64 {
	return cecHybrid(x, shift, rotation, shuffle, cec2017Hybrids[number])
}

func cecHybrid2020(number int, x, shift, rotation []float64, shuffle []int) float64 {
	return cecHybrid(x, shift, rotation, shuffle, cec2020Hybrids[number])
}

func cecHybrid(x, shift, rotation []float64, shuffle []int, spec cecHybridSpec) float64 {
	z := cecTransform(x, shift, rotation, 1)

	permuted := make([]float64, len(z))
	for i, index := range shuffle {
		permuted[i] = z[index-1]
	}

	sizes := cecPartitionSizes(len(x), spec.proportions, spec.rightSized)

	result, offset := 0.0, 0
	for i, size := range sizes {
		part := permuted[offset : offset+size]

		switch spec.functions[i] {
		case cecSchafferF7:
			// Schaffer F7 reads the shared pre-rotation buffer in the reference
			// evaluator. Inside a hybrid this means the first group, regardless
			// of the component's offset.
			result += cecSchafferF7Value(permuted[:size])
		case cecLunacek:
			// Even with shifting disabled, the released hybrid evaluator uses
			// the first shift-vector signs inside Lunacek's asymmetric transform.
			result += cecLunacekValue(part, shift[:size], nil, false)
		default:
			result += cecEvaluateBase(spec.functions[i], part, nil, nil)
		}

		offset += size
	}

	return result
}

func cecPartitionSizes(dimension int, proportions []float64, rightSized bool) []int {
	sizes := make([]int, len(proportions))
	used := 0

	if rightSized {
		for i := 1; i < len(proportions); i++ {
			sizes[i] = int(math.Ceil(proportions[i] * float64(dimension)))
			used += sizes[i]
		}

		sizes[0] = dimension - used
		switch {
		case dimension == 5 && len(sizes) == 5:
			for i := range sizes {
				sizes[i] = 1
			}
		case dimension == 5 && len(sizes) == 4:
			sizes = []int{1, 1, 1, 2}
		}

		return sizes
	}

	for i := 0; i+1 < len(proportions); i++ {
		sizes[i] = int(math.Ceil(proportions[i] * float64(dimension)))
		used += sizes[i]
	}

	sizes[len(sizes)-1] = dimension - used

	return sizes
}

type cecComponent struct {
	base        cecBase
	hybrid      int
	scale       float64
	delta, bias float64
}

var cecCompositions = map[int][]cecComponent{
	1: {
		{base: cecRosenbrock, scale: 1, delta: 10},
		{base: cecElliptic, scale: 1e4 / 1e10, delta: 20, bias: 100},
		{base: cecRastrigin, scale: 1, delta: 30, bias: 200},
	},
	2: {
		{base: cecRastrigin, scale: 1, delta: 10},
		{base: cecGriewank, scale: 1e3 * 0.01, delta: 20, bias: 100},
		{base: cecSchwefel, scale: 1, delta: 30, bias: 200},
	},
	3: {
		{base: cecRosenbrock, scale: 1, delta: 10},
		{base: cecAckley, scale: 1e3 * 0.01, delta: 20, bias: 100},
		{base: cecSchwefel, scale: 1, delta: 30, bias: 200},
		{base: cecRastrigin, scale: 1, delta: 40, bias: 300},
	},
	4: {
		{base: cecAckley, scale: 1e3 * 0.01, delta: 10},
		{base: cecElliptic, scale: 1e4 / 1e10, delta: 20, bias: 100},
		{base: cecGriewank, scale: 1e3 * 0.01, delta: 30, bias: 200},
		{base: cecRastrigin, scale: 1, delta: 40, bias: 300},
	},
	5: {
		{base: cecRastrigin, scale: 1e4 / 1e3, delta: 10},
		{base: cecHappyCat, scale: 1, delta: 20, bias: 100},
		{base: cecAckley, scale: 1e3 * 0.01, delta: 30, bias: 200},
		{base: cecDiscus, scale: 1e4 / 1e10, delta: 40, bias: 300},
		{base: cecRosenbrock, scale: 1, delta: 50, bias: 400},
	},
	6: {
		{base: cecExpandedSchaffer6, scale: 1e4 / 2e7, delta: 10},
		{base: cecSchwefel, scale: 1, delta: 20, bias: 100},
		{base: cecGriewank, scale: 1e3 * 0.01, delta: 20, bias: 200},
		{base: cecRosenbrock, scale: 1, delta: 30, bias: 300},
		{base: cecRastrigin, scale: 1e4 / 1e3, delta: 40, bias: 400},
	},
	7: {
		{base: cecHGBat, scale: 1e4 / 1e3, delta: 10},
		{base: cecRastrigin, scale: 1e4 / 1e3, delta: 20, bias: 100},
		{base: cecSchwefel, scale: 1e4 / 4e3, delta: 30, bias: 200},
		{base: cecBentCigar, scale: 1e4 / 1e30, delta: 40, bias: 300},
		{base: cecElliptic, scale: 1e4 / 1e10, delta: 50, bias: 400},
		{base: cecExpandedSchaffer6, scale: 1e4 / 2e7, delta: 60, bias: 500},
	},
	8: {
		{base: cecAckley, scale: 1e3 * 0.01, delta: 10},
		{base: cecGriewank, scale: 1e3 * 0.01, delta: 20, bias: 100},
		{base: cecDiscus, scale: 1e4 / 1e10, delta: 30, bias: 200},
		{base: cecRosenbrock, scale: 1, delta: 40, bias: 300},
		{base: cecHappyCat, scale: 1, delta: 50, bias: 400},
		{base: cecExpandedSchaffer6, scale: 1e4 / 2e7, delta: 60, bias: 500},
	},
	9: {
		{hybrid: 5, scale: 1, delta: 10},
		{hybrid: 6, scale: 1, delta: 30, bias: 100},
		{hybrid: 7, scale: 1, delta: 50, bias: 200},
	},
	10: {
		{hybrid: 5, scale: 1, delta: 10},
		{hybrid: 8, scale: 1, delta: 30, bias: 100},
		{hybrid: 9, scale: 1, delta: 50, bias: 200},
	},
}

func (instance *cecInstance) composition2017(number int, x []float64) float64 {
	components := cecCompositions[number]
	values := make([]float64, len(components))
	weights := make([]float64, len(components))

	n := instance.dimension
	for i, component := range components {
		shift := instance.shift[i*n : (i+1)*n]

		rotation := instance.rotation[i*n*n : (i+1)*n*n]
		if component.hybrid == 0 {
			values[i] = cecEvaluateBase(component.base, x, shift, rotation)
		} else {
			shuffle := instance.shuffle[i*n : (i+1)*n]
			values[i] = cecHybrid2017(component.hybrid, x, shift, rotation, shuffle)
		}

		values[i] = values[i]*component.scale + component.bias
		distance := 0.0

		for j, coordinate := range x {
			offset := coordinate - shift[j]
			distance += offset * offset
		}

		if distance == 0 {
			weights[i] = 1e99
		} else {
			weights[i] = math.Exp(-distance/(2*float64(n)*component.delta*component.delta)) / math.Sqrt(distance)
		}
	}

	weightSum, result := 0.0, 0.0
	for _, weight := range weights {
		weightSum += weight
	}

	if weightSum == 0 {
		weightSum = float64(len(weights))
		for i := range weights {
			weights[i] = 1
		}
	}

	for i, weight := range weights {
		result += weight / weightSum * values[i]
	}

	return result
}
