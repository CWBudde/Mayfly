package main

import (
	"math"
	"testing"
)

func TestDecayObjectiveCopiesAndFitsData(t *testing.T) {
	data := []observation{
		{time: 0, value: 5.000000},
		{time: 0.5, value: 3.834613},
		{time: 1, value: 2.992588},
		{time: 2, value: 1.944636},
		{time: 3, value: 1.397553},
		{time: 4, value: 1.111950},
	}

	objective, err := newDecayObjective(data)
	if err != nil {
		t.Fatalf("newDecayObjective() error = %v", err)
	}

	before := objective([]float64{0.42, (0.65 - 0.05) / 1.95, 0.7})
	data[0].value = -100

	after := objective([]float64{0.42, (0.65 - 0.05) / 1.95, 0.7})
	if before != after {
		t.Fatalf("objective changed after caller mutated data: before=%g after=%g", before, after)
	}

	if before > 1e-10 {
		t.Fatalf("objective at generating parameters = %g, want near zero", before)
	}
}

func TestFitRecoversDecayParameters(t *testing.T) {
	parameters, result, err := fit([]observation{
		{time: 0, value: 5.000000},
		{time: 0.5, value: 3.834613},
		{time: 1, value: 2.992588},
		{time: 2, value: 1.944636},
		{time: 3, value: 1.397553},
		{time: 4, value: 1.111950},
	})
	if err != nil {
		t.Fatalf("fit() error = %v", err)
	}

	if result.Seed == nil || *result.Seed != 42 {
		t.Fatalf("result seed = %v, want 42", result.Seed)
	}

	if result.GlobalBest.Cost > 1e-6 {
		t.Fatalf("best cost = %g, want <= 1e-6", result.GlobalBest.Cost)
	}

	checks := []struct {
		name string
		got  float64
		want float64
	}{
		{name: "amplitude", got: parameters.amplitude, want: 4.2},
		{name: "decay", got: parameters.decay, want: 0.65},
		{name: "baseline", got: parameters.baseline, want: 0.8},
	}
	for _, check := range checks {
		if math.Abs(check.got-check.want) > 0.01 {
			t.Errorf("%s = %g, want %g ± 0.01", check.name, check.got, check.want)
		}
	}
}

func TestNewDecayObjectiveRejectsInvalidData(t *testing.T) {
	_, emptyErr := newDecayObjective(nil)
	if emptyErr == nil {
		t.Fatal("newDecayObjective(nil) succeeded")
	}

	_, nonFiniteErr := newDecayObjective([]observation{{time: 0, value: math.NaN()}})
	if nonFiniteErr == nil {
		t.Fatal("newDecayObjective() accepted non-finite data")
	}
}
