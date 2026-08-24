package mayfly

import (
	"math"
	"testing"
)

func TestEngineeringBenchmarkPublishedSolutions(t *testing.T) {
	problems, err := EngineeringBenchmarkSuite()
	if err != nil {
		t.Fatal(err)
	}

	if len(problems) != 4 {
		t.Fatalf("engineering suite has %d problems, want 4", len(problems))
	}

	for _, problem := range problems {
		value, err := problem.Evaluate(problem.Optimum())
		if err != nil {
			t.Fatalf("%s optimum: %v", problem.Name(), err)
		}

		if math.Abs(value-problem.Minimum()) > 5e-4 {
			t.Errorf("%s reference value = %.12g, want %.12g", problem.Name(), value, problem.Minimum())
		}

		evaluation, err := EvaluateConstraintsChecked(problem.Optimum(), problem.Constraints())
		if err != nil {
			t.Fatalf("%s constraints: %v", problem.Name(), err)
		}

		if !evaluation.Feasible && evaluation.Violation > 1e-12 {
			t.Errorf("%s reference solution violation = %.12g", problem.Name(), evaluation.Violation)
		}
	}
}

func TestEngineeringDiscreteProjection(t *testing.T) {
	pressure, err := NewPressureVesselDesignProblem()
	if err != nil {
		t.Fatal(err)
	}

	integer := []float64{13, 7, 42, 177}
	fractional := []float64{13.1, 6.9, 42, 177}

	want, err := pressure.Evaluate(integer)
	if err != nil {
		t.Fatal(err)
	}

	got, err := pressure.Evaluate(fractional)
	if err != nil {
		t.Fatal(err)
	}

	if got != want {
		t.Fatalf("pressure-vessel thickness projection = %v, want %v", got, want)
	}

	reducer, err := NewSpeedReducerDesignProblem()
	if err != nil {
		t.Fatal(err)
	}

	integer = reducer.Optimum()
	fractional = reducer.Optimum()
	fractional[2] = 17.1
	want, _ = reducer.Evaluate(integer)

	got, _ = reducer.Evaluate(fractional)
	if got != want {
		t.Fatalf("speed-reducer tooth projection = %v, want %v", got, want)
	}
}

func TestEngineeringNormalizedConfigPreservesPhysicalEvaluation(t *testing.T) {
	problem, err := NewWeldedBeamDesignProblem()
	if err != nil {
		t.Fatal(err)
	}

	physical := problem.Optimum()
	lower, upper := problem.Bounds()

	normalized := make([]float64, problem.Dimension())
	for i := range normalized {
		normalized[i] = (physical[i] - lower[i]) / (upper[i] - lower[i])
	}

	decoded, err := problem.Decode(normalized)
	if err != nil {
		t.Fatal(err)
	}

	for i := range physical {
		if math.Abs(decoded[i]-physical[i]) > 1e-12 {
			t.Fatalf("decoded[%d] = %v, want %v", i, decoded[i], physical[i])
		}
	}

	config, err := problem.NewConfig(nil)
	if err != nil {
		t.Fatal(err)
	}

	physicalValue, _ := problem.Evaluate(physical)
	if got := config.ObjectiveFunc(normalized); math.Abs(got-physicalValue) > 1e-12 {
		t.Fatalf("normalized objective = %v, want %v", got, physicalValue)
	}

	physicalConstraints, _ := EvaluateConstraintsChecked(physical, problem.Constraints())

	normalizedConstraints, _ := EvaluateConstraintsChecked(normalized, config.Constraints)
	if math.Abs(physicalConstraints.Violation-normalizedConstraints.Violation) > 1e-12 {
		t.Fatalf("normalized constraint violation = %v, want %v", normalizedConstraints.Violation, physicalConstraints.Violation)
	}
}

func TestBenchmarkCaseDefensiveCopiesAndValidation(t *testing.T) {
	problem, err := NewSpringDesignProblem()
	if err != nil {
		t.Fatal(err)
	}

	lower, upper := problem.Bounds()
	lower[0], upper[0] = -100, 100

	lowerAgain, upperAgain := problem.Bounds()
	if lowerAgain[0] == -100 || upperAgain[0] == 100 {
		t.Fatal("Bounds returned mutable aliases")
	}

	optimum := problem.Optimum()
	optimum[0] = 1

	if problem.Optimum()[0] == 1 {
		t.Fatal("Optimum returned a mutable alias")
	}

	constraints := problem.Constraints()

	constraints.Inequalities[0] = nil
	if problem.Constraints().Inequalities[0] == nil {
		t.Fatal("Constraints returned a mutable alias")
	}

	if _, err := problem.Decode([]float64{-0.1, 0, 0}); err == nil {
		t.Fatal("Decode should reject coordinates outside the unit box")
	}

	if _, err := problem.Evaluate([]float64{math.NaN(), 1, 2}); err == nil {
		t.Fatal("Evaluate should reject non-finite coordinates")
	}
}
