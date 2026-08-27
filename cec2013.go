package mayfly

import (
	"errors"
	"fmt"
	"io/fs"
)

var cec2013Names = [...]string{
	"",
	"Sphere",
	"Rotated High Conditioned Elliptic",
	"Rotated Bent Cigar",
	"Rotated Discus",
	"Different Powers",
	"Rotated Rosenbrock",
	"Rotated Schaffers F7",
	"Rotated Ackley",
	"Rotated Weierstrass",
	"Rotated Griewank",
	"Rastrigin",
	"Rotated Rastrigin",
	"Non-Continuous Rotated Rastrigin",
	"Schwefel",
	"Rotated Schwefel",
	"Rotated Katsuura",
	"Lunacek Bi-Rastrigin",
	"Rotated Lunacek Bi-Rastrigin",
	"Expanded Griewank plus Rosenbrock",
	"Expanded Schaffer F6",
	"Composition 1 (N=5)",
	"Composition 2 (N=3)",
	"Composition 3 (N=3)",
	"Composition 4 (N=3)",
	"Composition 5 (N=3)",
	"Composition 6 (N=5)",
	"Composition 7 (N=5)",
	"Composition 8 (N=5)",
}

var cec2013Bias = [...]float64{
	0,
	-1400, -1300, -1200, -1100, -1000, -900, -800,
	-700, -600, -500, -400, -300, -200, -100,
	100, 200, 300, 400, 500, 600, 700,
	800, 900, 1000, 1100, 1200, 1300, 1400,
}

type cec2013Instance struct {
	shift     []float64
	rotation  []float64
	function  int
	dimension int
}

// NewCEC2013Problem loads one official CEC2013 real-parameter problem for the
// 30-dimensional DESMA protocol. data may point at the supplement's data
// directory, its input_data directory, or an extracted supplement root.
func NewCEC2013Problem(data fs.FS, function, dimension int) (*BenchmarkCase, error) {
	if function < 1 || function >= len(cec2013Names) {
		return nil, fmt.Errorf("CEC2013 function must be in F1-F28, got F%d", function)
	}

	if dimension != 30 {
		return nil, fmt.Errorf("CEC2013 DESMA protocol dimension must be 30, got %d", dimension)
	}

	shift, rotation, err := loadCEC2013Data(data, dimension)
	if err != nil {
		return nil, fmt.Errorf("load CEC2013 F%d: %w", function, err)
	}

	return newCEC2013Benchmark(function, dimension, shift, rotation)
}

// CEC2013Suite loads all 28 official CEC2013 problems for the 30-dimensional
// DESMA protocol. The shared shift and rotation files are parsed only once.
func CEC2013Suite(data fs.FS, dimension int) ([]*BenchmarkCase, error) {
	if dimension != 30 {
		return nil, fmt.Errorf("CEC2013 DESMA protocol dimension must be 30, got %d", dimension)
	}

	shift, rotation, err := loadCEC2013Data(data, dimension)
	if err != nil {
		return nil, fmt.Errorf("load CEC2013 suite: %w", err)
	}

	problems := make([]*BenchmarkCase, 0, 28)

	for function := 1; function <= 28; function++ {
		problem, problemErr := newCEC2013Benchmark(function, dimension, shift, rotation)
		if problemErr != nil {
			return nil, problemErr
		}

		problems = append(problems, problem)
	}

	return problems, nil
}

func loadCEC2013Data(data fs.FS, dimension int) ([]float64, []float64, error) {
	if data == nil {
		return nil, nil, errors.New("CEC2013 input data filesystem is nil")
	}

	rotation, err := readCECFloatFile(data, fmt.Sprintf("M_D%d.txt", dimension), 10*dimension*dimension)
	if err != nil {
		return nil, nil, fmt.Errorf("load rotation: %w", err)
	}

	// The released evaluator reads the first 10*D values sequentially from
	// shift_data.txt. Do not select the first D columns from each 100-D row.
	shift, err := readCECFloatFile(data, "shift_data.txt", 10*dimension)
	if err != nil {
		return nil, nil, fmt.Errorf("load shift: %w", err)
	}

	return shift, rotation, nil
}

func newCEC2013Benchmark(
	function, dimension int,
	shift, rotation []float64,
) (*BenchmarkCase, error) {
	instance := &cec2013Instance{
		shift: shift, rotation: rotation, function: function, dimension: dimension,
	}
	lower, upper := repeatedBounds(dimension, -100, 100)

	return newBenchmarkCase(BenchmarkCase{
		suite:          "CEC2013",
		name:           fmt.Sprintf("CEC2013 F%d: %s", function, cec2013Names[function]),
		number:         function,
		dimension:      dimension,
		lower:          lower,
		upper:          upper,
		optimum:        append([]float64(nil), shift[:dimension]...),
		minimum:        cec2013Bias[function],
		maxEvaluations: 10000 * dimension,
		objective:      instance.objective(),
		constraints:    nil,
		project:        nil,
	})
}
