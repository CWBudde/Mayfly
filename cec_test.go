package mayfly

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"testing/fstest"
)

func TestCEC2017SuiteLoadsAllUsableFunctions(t *testing.T) {
	suite, err := CEC2017Suite(makeCECFixture(2017, 10), 10)
	if err != nil {
		t.Fatal(err)
	}

	if len(suite) != 29 {
		t.Fatalf("CEC2017 suite has %d functions, want 29", len(suite))
	}

	for _, problem := range suite {
		got, err := problem.Evaluate(problem.Optimum())
		if err != nil {
			t.Fatalf("%s optimum: %v", problem.Name(), err)
		}

		if math.Abs(got-problem.Minimum()) > 1e-7 {
			t.Errorf("%s optimum = %.12g, want %.12g", problem.Name(), got, problem.Minimum())
		}

		if problem.MaxEvaluations() != 100000 {
			t.Errorf("%s budget = %d, want 100000", problem.Name(), problem.MaxEvaluations())
		}
	}
}

func TestCEC2020SuiteLoadsAllFunctions(t *testing.T) {
	suite, err := CEC2020Suite(makeCECFixture(2020, 5), 5)
	if err != nil {
		t.Fatal(err)
	}

	if len(suite) != 10 {
		t.Fatalf("CEC2020 suite has %d functions, want 10", len(suite))
	}

	for _, problem := range suite {
		got, err := problem.Evaluate(problem.Optimum())
		if err != nil {
			t.Fatalf("%s optimum: %v", problem.Name(), err)
		}

		if math.Abs(got-problem.Minimum()) > 1e-7 {
			t.Errorf("%s optimum = %.12g, want %.12g", problem.Name(), got, problem.Minimum())
		}

		if problem.MaxEvaluations() != 50000 {
			t.Errorf("%s budget = %d, want 50000", problem.Name(), problem.MaxEvaluations())
		}
	}
}

func TestCECConstructorsRejectInvalidOrIncompleteInputs(t *testing.T) {
	if _, err := NewCEC2017Problem(makeCECFixture(2017, 10), 2, 10); err == nil {
		t.Fatal("CEC2017 F2 should be rejected")
	}

	if _, err := NewCEC2017Problem(makeCECFixture(2017, 10), 1, 20); err == nil {
		t.Fatal("non-competition CEC2017 dimension should be rejected")
	}

	if _, err := NewCEC2020Problem(makeCECFixture(2020, 5), 11, 5); err == nil {
		t.Fatal("CEC2020 F11 should be rejected")
	}

	if _, err := NewCEC2020Problem(nil, 1, 5); err == nil {
		t.Fatal("nil data filesystem should be rejected")
	}

	truncated := fstest.MapFS{
		"M_1_D10.txt":            &fstest.MapFile{Data: []byte("1 0")},
		"shift_data_1.txt":       &fstest.MapFile{Data: []byte(strings.Repeat("0 ", 10))},
		"shuffle_data_1_D10.txt": &fstest.MapFile{Data: []byte("1 2 3 4 5 6 7 8 9 10")},
	}
	if _, err := NewCEC2017Problem(truncated, 1, 10); err == nil {
		t.Fatal("truncated rotation data should be rejected")
	}
}

func TestCECReferenceEvaluatorQuirksRemainCompatible(t *testing.T) {
	x := []float64{1.25, -0.75, 0.5}
	shift := []float64{0.1, -0.2, 0.3}
	rotation := []float64{
		0, 1, 0,
		1, 0, 0,
		0, 0, 1,
	}
	// The released Schaffer F7 evaluator uses its pre-rotation scratch.
	gotSchaffer := cecEvaluateBase(cecSchafferF7, x, shift, rotation)

	wantSchaffer := cecSchafferF7Value(cecTransform(x, shift, nil, 1))
	if gotSchaffer != wantSchaffer {
		t.Fatalf("Schaffer F7 reference compatibility = %v, want %v", gotSchaffer, wantSchaffer)
	}
	// Its non-continuous Rastrigin rounding writes a discarded scratch vector.
	gotStep := cecEvaluateBase(cecNonContinuousRastrigin, x, shift, rotation)

	wantRastrigin := cecEvaluateBase(cecRastrigin, x, shift, rotation)
	if gotStep != wantRastrigin {
		t.Fatalf("non-continuous Rastrigin reference compatibility = %v, want %v", gotStep, wantRastrigin)
	}
}

func TestCECProblemIsSafeForOptimizerUse(t *testing.T) {
	problem, err := NewCEC2020Problem(makeCECFixture(2020, 5), 1, 5)
	if err != nil {
		t.Fatal(err)
	}

	if !math.IsInf(problem.Objective()([]float64{0}), 1) {
		t.Fatal("wrong-dimension objective should score +Inf")
	}

	if _, err := problem.Evaluate([]float64{101, 0, 0, 0, 0}); err == nil {
		t.Fatal("out-of-bounds physical position should be rejected")
	}

	base := NewDefaultConfig()

	config, err := problem.NewConfig(base)
	if err != nil {
		t.Fatal(err)
	}

	if config.ProblemSize != 5 || config.LowerBound != 0 || config.UpperBound != 1 {
		t.Fatalf("unexpected normalized config: D=%d bounds=[%v,%v]", config.ProblemSize, config.LowerBound, config.UpperBound)
	}

	if base.ProblemSize != 0 || base.ObjectiveFunc != nil {
		t.Fatal("NewConfig mutated its base configuration")
	}

	center := []float64{0.5, 0.5, 0.5, 0.5, 0.5}

	physical, err := problem.Decode(center)
	if err != nil {
		t.Fatal(err)
	}

	want, err := problem.Evaluate(physical)
	if err != nil {
		t.Fatal(err)
	}

	if got := config.ObjectiveFunc(center); got != want {
		t.Fatalf("normalized objective = %v, want physical value %v", got, want)
	}
}

func makeCECFixture(year, dimension int) fstest.MapFS {
	data := fstest.MapFS{}
	internalFunctions := []int{}

	if year == 2017 {
		for function := 1; function <= 30; function++ {
			if function != 2 {
				internalFunctions = append(internalFunctions, function)
			}
		}
	} else {
		internalFunctions = append(internalFunctions, cec2020Internal[1:]...)
	}

	for _, internal := range internalFunctions {
		components := 1
		if internal > 20 {
			components = 10
		}

		data[fmt.Sprintf("M_%d_D%d.txt", internal, dimension)] = &fstest.MapFile{
			Data: []byte(identityMatrices(components, dimension)),
		}
		data[fmt.Sprintf("shift_data_%d.txt", internal)] = &fstest.MapFile{
			Data: []byte(distinctShiftRows(components, dimension)),
		}
		shuffleCount := 0

		if year == 2017 {
			if internal >= 11 && internal <= 20 {
				shuffleCount = dimension
			} else if internal == 29 || internal == 30 {
				shuffleCount = 10 * dimension
			}
		} else if internal == 4 || internal == 6 || (internal >= 11 && internal <= 20) {
			shuffleCount = dimension
		}

		if shuffleCount > 0 {
			var shuffle strings.Builder
			for i := range shuffleCount {
				fmt.Fprintf(&shuffle, "%d ", i%dimension+1)
			}

			data[fmt.Sprintf("shuffle_data_%d_D%d.txt", internal, dimension)] = &fstest.MapFile{Data: []byte(shuffle.String())}
		}
	}

	return data
}

func identityMatrices(count, dimension int) string {
	var result strings.Builder

	for range count {
		for row := range dimension {
			for column := range dimension {
				if row == column {
					result.WriteString("1 ")
				} else {
					result.WriteString("0 ")
				}
			}
		}
	}

	return result.String()
}

func distinctShiftRows(count, dimension int) string {
	var result strings.Builder

	for row := range count {
		for range dimension {
			fmt.Fprintf(&result, "%d ", row*10)
		}

		result.WriteByte('\n')
	}

	return result.String()
}
