package main

import (
	"math"
	"testing"
)

func TestOptimize(t *testing.T) {
	result, err := optimize()
	if err != nil {
		t.Fatalf("optimize: %v", err)
	}

	if len(result.GlobalBest.Position) != 2 {
		t.Fatalf("position dimensions = %d, want 2", len(result.GlobalBest.Position))
	}

	if result.Seed == nil || *result.Seed != 42 {
		t.Fatalf("seed = %v, want 42", result.Seed)
	}

	if result.IterationCount == 0 || result.FuncEvalCount == 0 {
		t.Fatalf("work counters = %d iterations, %d evaluations", result.IterationCount, result.FuncEvalCount)
	}

	if math.IsNaN(result.GlobalBest.Cost) || math.IsInf(result.GlobalBest.Cost, 0) {
		t.Fatalf("cost is not finite: %v", result.GlobalBest.Cost)
	}

	if result.GlobalBest.Cost >= objective([]float64{0, 0}) {
		t.Fatalf("cost = %g, did not improve on the origin", result.GlobalBest.Cost)
	}
}
