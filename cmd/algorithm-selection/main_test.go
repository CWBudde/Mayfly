package main

import (
	"math"
	"testing"
)

func TestChooseAndOptimize(t *testing.T) {
	selected, result, err := chooseAndOptimize()
	if err != nil {
		t.Fatalf("chooseAndOptimize() error = %v", err)
	}

	if selected.Variant == nil || selected.Variant.Name() != "DESMA" {
		t.Fatalf("selected variant = %v, want DESMA", selected.Variant)
	}

	if result.Seed == nil || *result.Seed != 42 {
		t.Fatalf("result seed = %v, want 42", result.Seed)
	}

	if !isFinite(result.GlobalBest.Cost) {
		t.Fatalf("global best = %#v, want finite cost", result.GlobalBest)
	}

	if result.IterationCount == 0 || result.FuncEvalCount == 0 {
		t.Fatalf("work counters = (%d, %d), want non-zero", result.IterationCount, result.FuncEvalCount)
	}
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
