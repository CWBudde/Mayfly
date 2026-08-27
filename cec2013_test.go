package mayfly

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
)

func TestCEC2013SuiteLoadsAllFunctions(t *testing.T) {
	suite, err := CEC2013Suite(makeCEC2013Fixture(""), 30)
	if err != nil {
		t.Fatal(err)
	}

	if len(suite) != 28 {
		t.Fatalf("CEC2013 suite has %d functions, want 28", len(suite))
	}

	for i, problem := range suite {
		function := i + 1
		if problem.Suite() != "CEC2013" || problem.Number() != function || problem.Dimension() != 30 {
			t.Errorf("F%d metadata = suite %q, number %d, D=%d", function,
				problem.Suite(), problem.Number(), problem.Dimension())
		}

		if problem.MaxEvaluations() != 300000 {
			t.Errorf("F%d budget = %d, want 300000", function, problem.MaxEvaluations())
		}

		lower, upper := problem.Bounds()
		for coordinate := range 30 {
			if lower[coordinate] != -100 || upper[coordinate] != 100 {
				t.Fatalf("F%d bound %d = [%v,%v], want [-100,100]", function,
					coordinate, lower[coordinate], upper[coordinate])
			}
		}

		got, evaluateErr := problem.Evaluate(problem.Optimum())
		if evaluateErr != nil {
			t.Fatalf("F%d optimum: %v", function, evaluateErr)
		}

		assertCEC2013Close(t, function, got, cec2013Bias[function])
	}
}

func TestCEC2013MatchesOfficialSourceEvaluatorGoldens(t *testing.T) {
	// These outputs were generated from the official cec13_func.cpp using ten
	// identity matrices and ten constant shift blocks (0,10,...,90). Only the
	// documented Linux portability edit, %Lf to %lf, was applied. The fixture
	// isolates evaluator equations without redistributing organizer data.
	want := [...]float64{
		0,
		67013.793103448261,
		13495785411.916309,
		6.9574118357828823e26,
		6395605728.2010412,
		517096.6074663336,
		4546.153211337758,
		920763386.12100446,
		-678.4650181194354,
		-542.17492615169556,
		20186.699248832556,
		1641.7371829162937,
		1741.7371829162937,
		1843.4392865153543,
		11440.954264806882,
		11640.954264806882,
		205.22185633591741,
		3327.2016806882998,
		3427.2016806882998,
		2120351.3765346557,
		615.13479767873173,
		10232.915624509209,
		13539.764374979024,
		13639.764374979024,
		2853.1410346139392,
		1877.3879488385367,
		3505.9826067733684,
		7197.8295949509557,
		23260.640698313189,
	}

	x := make([]float64, 30)
	for i := range x {
		x[i] = -80 + 160*float64(i)/29
	}

	for function := 1; function <= 28; function++ {
		problem, err := NewCEC2013Problem(makeCEC2013Fixture(""), function, 30)
		if err != nil {
			t.Fatal(err)
		}

		got, err := problem.Evaluate(x)
		if err != nil {
			t.Fatalf("F%d: %v", function, err)
		}

		assertCEC2013Close(t, function, got, want[function])
	}
}

func TestCEC2013LoaderUsesOfficialFlatShiftPrefix(t *testing.T) {
	var shifts strings.Builder
	for value := range 1000 {
		fmt.Fprintf(&shifts, "%d ", value)

		if (value+1)%100 == 0 {
			shifts.WriteByte('\n')
		}
	}

	data := fstest.MapFS{
		"M_D30.txt":      &fstest.MapFile{Data: []byte(identityMatrices(10, 30))},
		"shift_data.txt": &fstest.MapFile{Data: []byte(shifts.String())},
	}

	shift, _, err := loadCEC2013Data(data, 30)
	if err != nil {
		t.Fatal(err)
	}

	if shift[29] != 29 || shift[30] != 30 || shift[299] != 299 {
		t.Fatalf("flat shift prefix tripwire = [%v,%v,%v], want [29,30,299]",
			shift[29], shift[30], shift[299])
	}
}

func TestCEC2013ConstructorsRejectInvalidOrIncompleteInputs(t *testing.T) {
	fixture := makeCEC2013Fixture("")

	for _, function := range []int{0, 29} {
		if _, err := NewCEC2013Problem(fixture, function, 30); err == nil {
			t.Errorf("CEC2013 F%d should be rejected", function)
		}
	}
	if _, err := NewCEC2013Problem(fixture, 1, 10); err == nil {
		t.Fatal("non-DESMA CEC2013 dimension should be rejected")
	}
	if _, err := NewCEC2013Problem(nil, 1, 30); err == nil {
		t.Fatal("nil data filesystem should be rejected")
	}

	truncatedMatrix := makeCEC2013Fixture("")
	truncatedMatrix["M_D30.txt"] = &fstest.MapFile{Data: []byte("1 0")}
	if _, err := NewCEC2013Problem(truncatedMatrix, 1, 30); err == nil {
		t.Fatal("truncated rotation data should be rejected")
	}

	truncatedShift := makeCEC2013Fixture("")
	truncatedShift["shift_data.txt"] = &fstest.MapFile{Data: []byte("0")}
	if _, err := NewCEC2013Problem(truncatedShift, 1, 30); err == nil {
		t.Fatal("truncated shift data should be rejected")
	}
}

func TestCEC2013AcceptsExtractedSupplementLayout(t *testing.T) {
	problem, err := NewCEC2013Problem(makeCEC2013Fixture("data/input_data/"), 1, 30)
	if err != nil {
		t.Fatal(err)
	}

	if problem.Name() != "CEC2013 F1: Sphere" {
		t.Fatalf("problem name = %q", problem.Name())
	}
}

func TestCEC2013ObjectiveIsConcurrencySafe(t *testing.T) {
	problem, err := NewCEC2013Problem(makeCEC2013Fixture(""), 28, 30)
	if err != nil {
		t.Fatal(err)
	}

	x := make([]float64, 30)
	for i := range x {
		x[i] = -80 + 160*float64(i)/29
	}

	want, err := problem.Evaluate(x)
	if err != nil {
		t.Fatal(err)
	}

	var wait sync.WaitGroup
	for range 32 {
		wait.Add(1)
		go func() {
			defer wait.Done()

			for range 20 {
				got, evaluateErr := problem.Evaluate(x)
				if evaluateErr != nil || got != want {
					t.Errorf("parallel evaluation = %v, %v; want %v", got, evaluateErr, want)
				}
			}
		}()
	}

	wait.Wait()

	if !math.IsInf(problem.Objective()([]float64{0}), 1) {
		t.Fatal("wrong-dimension objective should score +Inf")
	}
}

func makeCEC2013Fixture(prefix string) fstest.MapFS {
	return fstest.MapFS{
		prefix + "M_D30.txt": &fstest.MapFile{
			Data: []byte(identityMatrices(10, 30)),
		},
		prefix + "shift_data.txt": &fstest.MapFile{
			Data: []byte(distinctShiftRows(10, 30)),
		},
	}
}

func assertCEC2013Close(t *testing.T, function int, got, want float64) {
	t.Helper()

	tolerance := 1e-10 + 1e-12*math.Abs(want)
	// Ackley's cosine phase amplifies the last-bit difference between Go's
	// math.Pow and the glibc pow used to generate this source-evaluator golden.
	if function == 8 {
		tolerance = max(tolerance, 2e-6)
	}

	if math.Abs(got-want) > tolerance {
		t.Errorf("CEC2013 F%d = %.17g, want %.17g (tolerance %.3g)",
			function, got, want, tolerance)
	}
}
