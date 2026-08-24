package mayfly

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"strconv"
	"strings"
)

var cec2017Names = map[int]string{
	1:  "Shifted and Rotated Bent Cigar",
	3:  "Shifted and Rotated Zakharov",
	4:  "Shifted and Rotated Rosenbrock",
	5:  "Shifted and Rotated Rastrigin",
	6:  "Shifted and Rotated Schaffer F7",
	7:  "Shifted and Rotated Lunacek Bi-Rastrigin",
	8:  "Shifted and Rotated Non-Continuous Rastrigin",
	9:  "Shifted and Rotated Levy",
	10: "Shifted and Rotated Schwefel",
	11: "Hybrid Function 1", 12: "Hybrid Function 2", 13: "Hybrid Function 3",
	14: "Hybrid Function 4", 15: "Hybrid Function 5", 16: "Hybrid Function 6",
	17: "Hybrid Function 7", 18: "Hybrid Function 8", 19: "Hybrid Function 9",
	20: "Hybrid Function 10",
	21: "Composition Function 1", 22: "Composition Function 2",
	23: "Composition Function 3", 24: "Composition Function 4",
	25: "Composition Function 5", 26: "Composition Function 6",
	27: "Composition Function 7", 28: "Composition Function 8",
	29: "Composition Function 9", 30: "Composition Function 10",
}

var cec2020Names = map[int]string{
	1:  "Shifted and Rotated Bent Cigar",
	2:  "Shifted and Rotated Schwefel",
	3:  "Shifted and Rotated Lunacek Bi-Rastrigin",
	4:  "Expanded Rosenbrock plus Griewank",
	5:  "Hybrid Function 1",
	6:  "Hybrid Function 2",
	7:  "Hybrid Function 3",
	8:  "Composition Function 1",
	9:  "Composition Function 2",
	10: "Composition Function 3",
}

var (
	cec2020Internal = [...]int{0, 1, 2, 3, 7, 4, 16, 6, 22, 24, 25}
	cec2020Bias     = [...]float64{0, 100, 1100, 700, 1900, 1700, 1600, 2100, 2200, 2400, 2500}
)

type cecInstance struct {
	shift      []float64
	rotation   []float64
	shuffle    []int
	year       int
	function   int
	internal   int
	dimension  int
	components int
}

// NewCEC2017Problem loads one exact CEC2017 bound-constrained problem from the
// organizers' input_data files. data may point at the input_data directory or
// at an extracted official archive containing it. Function 2 is rejected
// because the organizers removed it for numerical instability.
func NewCEC2017Problem(data fs.FS, function, dimension int) (*BenchmarkCase, error) {
	name, exists := cec2017Names[function]
	if !exists {
		if function == 2 {
			return nil, errors.New("CEC2017 F2 was removed by the organizers for numerical instability")
		}

		return nil, fmt.Errorf("CEC2017 function must be F1 or F3-F30, got F%d", function)
	}

	if !containsInt([]int{10, 30, 50, 100}, dimension) {
		return nil, fmt.Errorf("CEC2017 competition dimension must be one of 10, 30, 50, or 100, got %d", dimension)
	}

	instance, err := loadCECInstance(data, 2017, function, function, dimension)
	if err != nil {
		return nil, err
	}

	lower, upper := repeatedBounds(dimension, -100, 100)
	optimum := append([]float64(nil), instance.shift[:dimension]...)
	// The released CEC2017 evaluator applies the ordinary Levy function to
	// M(x-o), whose minimizer is the all-ones vector rather than zero. Unlike
	// the report's blanket x*=o statement, F9 therefore needs the inverse
	// rotation here. This preserves the executable benchmark's actual optimum.
	if function == 9 {
		target := make([]float64, dimension)
		for i := range target {
			target[i] = 1
		}

		offset, solveErr := solveCECLinearSystem(instance.rotation, target, dimension)
		if solveErr != nil {
			return nil, fmt.Errorf("derive CEC2017 F9 optimum: %w", solveErr)
		}

		for i := range optimum {
			optimum[i] += offset[i]
		}
	}

	return newBenchmarkCase(BenchmarkCase{
		suite:          "CEC2017",
		name:           fmt.Sprintf("CEC2017 F%d: %s", function, name),
		number:         function,
		dimension:      dimension,
		lower:          lower,
		upper:          upper,
		optimum:        optimum,
		minimum:        float64(function * 100),
		maxEvaluations: 10000 * dimension,
		objective:      instance.objective(),
	})
}

// CEC2017Suite loads all 29 usable CEC2017 problems for dimension. F2 is
// deliberately absent, matching the final official suite.
func CEC2017Suite(data fs.FS, dimension int) ([]*BenchmarkCase, error) {
	problems := make([]*BenchmarkCase, 0, len(cec2017Names))

	for function := 1; function <= 30; function++ {
		if function == 2 {
			continue
		}

		problem, err := NewCEC2017Problem(data, function, dimension)
		if err != nil {
			return nil, err
		}

		problems = append(problems, problem)
	}

	return problems, nil
}

// NewCEC2020Problem loads one exact CEC2020 single-objective bound-constrained
// problem from the organizers' input_data files.
func NewCEC2020Problem(data fs.FS, function, dimension int) (*BenchmarkCase, error) {
	name, exists := cec2020Names[function]
	if !exists {
		return nil, fmt.Errorf("CEC2020 function must be in F1-F10, got F%d", function)
	}

	if !containsInt([]int{5, 10, 15, 20}, dimension) {
		return nil, fmt.Errorf("CEC2020 competition dimension must be one of 5, 10, 15, or 20, got %d", dimension)
	}

	internal := cec2020Internal[function]

	instance, err := loadCECInstance(data, 2020, function, internal, dimension)
	if err != nil {
		return nil, err
	}

	lower, upper := repeatedBounds(dimension, -100, 100)
	optimum := append([]float64(nil), instance.shift[:dimension]...)

	return newBenchmarkCase(BenchmarkCase{
		suite:          "CEC2020",
		name:           fmt.Sprintf("CEC2020 F%d: %s", function, name),
		number:         function,
		dimension:      dimension,
		lower:          lower,
		upper:          upper,
		optimum:        optimum,
		minimum:        cec2020Bias[function],
		maxEvaluations: cec2020EvaluationBudget(dimension),
		objective:      instance.objective(),
	})
}

// CEC2020Suite loads all ten CEC2020 problems for dimension.
func CEC2020Suite(data fs.FS, dimension int) ([]*BenchmarkCase, error) {
	problems := make([]*BenchmarkCase, 0, 10)

	for function := 1; function <= 10; function++ {
		problem, err := NewCEC2020Problem(data, function, dimension)
		if err != nil {
			return nil, err
		}

		problems = append(problems, problem)
	}

	return problems, nil
}

func cec2020EvaluationBudget(dimension int) int {
	switch dimension {
	case 5:
		return 50000
	case 10:
		return 1000000
	case 15:
		return 3000000
	case 20:
		return 10000000
	default:
		return 0
	}
}

func loadCECInstance(data fs.FS, year, function, internal, dimension int) (*cecInstance, error) {
	if data == nil {
		return nil, fmt.Errorf("CEC%d input data filesystem is nil", year)
	}

	components := 1
	if internal > 20 {
		components = 10
	}

	matrixCount := components * dimension * dimension

	rotation, err := readCECFloatFile(data, fmt.Sprintf("M_%d_D%d.txt", internal, dimension), matrixCount)
	if err != nil {
		return nil, fmt.Errorf("load CEC%d F%d rotation: %w", year, function, err)
	}

	shift, err := readCECShiftFile(data, fmt.Sprintf("shift_data_%d.txt", internal), components, dimension)
	if err != nil {
		return nil, fmt.Errorf("load CEC%d F%d shift: %w", year, function, err)
	}

	shuffleCount := 0

	if year == 2017 {
		if internal >= 11 && internal <= 20 {
			shuffleCount = dimension
		} else if internal == 29 || internal == 30 {
			shuffleCount = 10 * dimension
		}
	} else {
		if internal == 4 || internal == 6 || (internal >= 11 && internal <= 20) {
			shuffleCount = dimension
		} else if internal == 29 || internal == 30 {
			shuffleCount = 10 * dimension
		}
	}

	var shuffle []int
	if shuffleCount > 0 {
		shuffle, err = readCECIntFile(data, fmt.Sprintf("shuffle_data_%d_D%d.txt", internal, dimension), shuffleCount)
		if err != nil {
			return nil, fmt.Errorf("load CEC%d F%d shuffle: %w", year, function, err)
		}

		for i, index := range shuffle {
			if index < 1 || index > dimension {
				return nil, fmt.Errorf("CEC%d F%d shuffle index %d at offset %d is outside 1..%d", year, function, index, i, dimension)
			}
		}
	}

	return &cecInstance{
		year: year, function: function, internal: internal, dimension: dimension,
		shift: shift, rotation: rotation, shuffle: shuffle, components: components,
	}, nil
}

func readCECShiftFile(data fs.FS, name string, rows, dimension int) ([]float64, error) {
	file, err := openCECData(data, name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	result := make([]float64, 0, rows*dimension)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024), 1024*1024)

	for scanner.Scan() && len(result) < rows*dimension {
		fields := strings.Fields(scanner.Text())
		if len(fields) < dimension {
			return nil, fmt.Errorf("%s row has %d values; need %d", name, len(fields), dimension)
		}

		for _, field := range fields[:dimension] {
			value, parseErr := strconv.ParseFloat(field, 64)
			if parseErr != nil {
				return nil, fmt.Errorf("parse %s: %w", name, parseErr)
			}

			result = append(result, value)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(result) != rows*dimension {
		return nil, fmt.Errorf("%s has %d selected values; need %d", name, len(result), rows*dimension)
	}

	return result, nil
}

func readCECFloatFile(data fs.FS, name string, count int) ([]float64, error) {
	file, err := openCECData(data, name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	result := make([]float64, count)
	for i := range count {
		if _, err := fmt.Fscan(reader, &result[i]); err != nil {
			return nil, fmt.Errorf("%s value %d: %w", name, i, err)
		}
	}

	return result, nil
}

func readCECIntFile(data fs.FS, name string, count int) ([]int, error) {
	file, err := openCECData(data, name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	result := make([]int, count)
	for i := range count {
		if _, err := fmt.Fscan(reader, &result[i]); err != nil {
			return nil, fmt.Errorf("%s value %d: %w", name, i, err)
		}
	}

	return result, nil
}

func openCECData(data fs.FS, name string) (fs.File, error) {
	paths := []string{
		name,
		"input_data/" + name,
		"CEC17_fast_pow/input_data/" + name,
		"Matlab version/input_data/" + name,
	}

	var lastErr error

	for _, path := range paths {
		file, err := data.Open(path)
		if err == nil {
			return file, nil
		}

		lastErr = err
	}

	if errors.Is(lastErr, fs.ErrNotExist) {
		return nil, fmt.Errorf("%s: %w", name, fs.ErrNotExist)
	}

	return nil, fmt.Errorf("open %s: %w", name, lastErr)
}

func containsInt(values []int, value int) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}

	return false
}

func solveCECLinearSystem(matrix, target []float64, dimension int) ([]float64, error) {
	augmented := make([]float64, dimension*(dimension+1))
	for row := range dimension {
		copy(augmented[row*(dimension+1):row*(dimension+1)+dimension], matrix[row*dimension:(row+1)*dimension])
		augmented[row*(dimension+1)+dimension] = target[row]
	}

	for column := range dimension {
		pivot := column
		for row := column + 1; row < dimension; row++ {
			if abs := math.Abs(augmented[row*(dimension+1)+column]); abs > math.Abs(augmented[pivot*(dimension+1)+column]) {
				pivot = row
			}
		}

		pivotValue := augmented[pivot*(dimension+1)+column]
		if math.Abs(pivotValue) < 1e-14 {
			return nil, errors.New("rotation matrix is singular")
		}

		if pivot != column {
			for entry := column; entry <= dimension; entry++ {
				a := column*(dimension+1) + entry
				b := pivot*(dimension+1) + entry
				augmented[a], augmented[b] = augmented[b], augmented[a]
			}
		}

		for row := column + 1; row < dimension; row++ {
			factor := augmented[row*(dimension+1)+column] / augmented[column*(dimension+1)+column]
			for entry := column; entry <= dimension; entry++ {
				augmented[row*(dimension+1)+entry] -= factor * augmented[column*(dimension+1)+entry]
			}
		}
	}

	result := make([]float64, dimension)
	for row := dimension - 1; row >= 0; row-- {
		value := augmented[row*(dimension+1)+dimension]
		for column := row + 1; column < dimension; column++ {
			value -= augmented[row*(dimension+1)+column] * result[column]
		}

		result[row] = value / augmented[row*(dimension+1)+row]
	}

	return result, nil
}
