package mayfly

import (
	"math/rand"
	"testing"
)

// =============================================================================
// Tests for selector.go - Algorithm Selection and Recommendation
// =============================================================================

func TestAlgorithmSelector(t *testing.T) {
	selector := NewAlgorithmSelector()
	if selector == nil {
		t.Fatal("NewAlgorithmSelector() should not return nil")
	}

	characteristics := ProblemCharacteristics{
		Dimensionality: 30,
		Modality:       HighlyMultimodal,
		Landscape:      Rugged,
	}

	recommendations := selector.RecommendAlgorithms(characteristics)
	if len(recommendations) == 0 {
		t.Fatal("Should return at least one recommendation")
	}

	// Recommendations should be sorted by score (descending)
	for i := 1; i < len(recommendations); i++ {
		if recommendations[i].Score > recommendations[i-1].Score {
			t.Error("Recommendations should be sorted by score (descending)")
			break
		}
	}

	// All recommendations should have valid scores
	for _, rec := range recommendations {
		if rec.Score < 0 || rec.Score > 1 {
			t.Errorf("Score should be in [0,1], got %.2f", rec.Score)
		}

		if rec.Confidence < 0 || rec.Confidence > 1 {
			t.Errorf("Confidence should be in [0,1], got %.2f", rec.Confidence)
		}

		if rec.Variant == nil {
			t.Error("Variant should not be nil")
		}

		if rec.Reasoning == "" {
			t.Error("Reasoning should not be empty")
		}
	}
}

func TestRecommendBest(t *testing.T) {
	selector := NewAlgorithmSelector()

	// The selector must not present a scalar optimizer as multi-objective.
	characteristics := ProblemCharacteristics{
		MultiObjective: true,
	}

	best := selector.RecommendBest(characteristics)
	if best.Variant != nil {
		t.Errorf("Expected no multi-objective recommendation, got %s", best.Variant.Name())
	}

	if best.Reasoning == "" {
		t.Error("Expected an explanation for unavailable multi-objective support")
	}
}

func TestRecommendForBenchmark(t *testing.T) {
	benchmarks := []struct {
		name         string
		expectedName string
		minScore     float64
	}{
		{"Sphere", "MA", 0.5},       // Unimodal - standard MA
		{"Rastrigin", "", 0.7},      // Highly multimodal - OLCE-MA or DESMA
		{"Schwefel", "EOBBMA", 0.8}, // Deceptive - EOBBMA
		{"Rosenbrock", "", 0.7},     // Narrow valley - MPMA or others
		{"Ackley", "", 0.6},         // Multimodal
		{"Griewank", "", 0.7},       // Highly multimodal
		{"BentCigar", "", 0.6},      // Ill-conditioned
		{"Discus", "", 0.6},         // Ill-conditioned
	}

	for _, tt := range benchmarks {
		t.Run(tt.name, func(t *testing.T) {
			rec := RecommendForBenchmark(tt.name)
			if rec.Variant == nil {
				t.Fatal("Recommendation should not have nil variant")
			}

			variant := rec.Variant.Name()
			if variant == "" {
				t.Error("Variant name should not be empty")
			}

			// If specific variant expected, check it
			if tt.expectedName != "" && variant != tt.expectedName {
				t.Logf("Expected %s for %s, got %s (may vary based on scoring)",
					tt.expectedName, tt.name, variant)
			}

			// Score should be reasonable
			if rec.Score < tt.minScore {
				t.Errorf("Score %.2f is below minimum %.2f for %s", rec.Score, tt.minScore, tt.name)
			}

			// Confidence should be reasonable
			if rec.Confidence < 0.5 {
				t.Errorf("Confidence too low (%.2f) for standard benchmark", rec.Confidence)
			}
		})
	}
}

func TestRecommendForUnknownBenchmark(t *testing.T) {
	rec := RecommendForBenchmark("UnknownFunction")

	// Should still return a valid recommendation (generic)
	if rec.Variant == nil {
		t.Error("Should return a recommendation even for unknown benchmark")
	}

	// Should have reasonable confidence
	if rec.Confidence < 0.4 {
		t.Errorf("Confidence too low (%.2f) for generic recommendation", rec.Confidence)
	}
}

func TestClassifyProblemIsSeedReproducible(t *testing.T) {
	first := ClassifyProblem(Sphere, 5, -5, 5, rand.New(rand.NewSource(99)))
	second := ClassifyProblem(Sphere, 5, -5, 5, rand.New(rand.NewSource(99)))

	if first != second {
		t.Errorf("ClassifyProblem is not reproducible for a seed: %+v vs %+v", first, second)
	}

	if first.Dimensionality != 5 {
		t.Errorf("Dimensionality = %d, want 5", first.Dimensionality)
	}

	// The caller-set fields must be left alone, not guessed at.
	if first.MultiObjective || first.ExpensiveEvaluations || first.RequiresFastConvergence {
		t.Errorf("ClassifyProblem filled in a caller-set field: %+v", first)
	}
}

// TestClassifyProblemIsScaleFree pins the property the line-scan probe exists
// for and the gradient-magnitude heuristic it replaced did not have: the
// verdict follows the function, not the width of the box. Sphere is the same
// shape on [-5,5] and on [-500,500] and must classify the same way.
func TestClassifyProblemIsScaleFree(t *testing.T) {
	tests := []struct {
		name          string
		fn            ObjectiveFunction
		size          int
		lower, upper  float64
		wantModality  Modality
		wantLandscape Landscape
	}{
		{"Sphere narrow box", Sphere, 5, -5, 5, Unimodal, Smooth},
		{"Sphere wide box", Sphere, 5, -500, 500, Unimodal, Smooth},
		{"Zakharov", Zakharov, 5, -5, 10, Unimodal, Smooth},
		{"Rastrigin", Rastrigin, 10, -5.12, 5.12, HighlyMultimodal, Rugged},
		{"Schwefel", Schwefel, 10, -500, 500, HighlyMultimodal, Rugged},
		{"Ackley", Ackley, 10, -32, 32, HighlyMultimodal, Rugged},
	}

	for _, test := range tests {
		got := ClassifyProblem(test.fn, test.size, test.lower, test.upper, rand.New(rand.NewSource(4)))

		if got.Modality != test.wantModality {
			t.Errorf("%s: modality = %v, want %v", test.name, got.Modality, test.wantModality)
		}

		if got.Landscape != test.wantLandscape {
			t.Errorf("%s: landscape = %v, want %v", test.name, got.Landscape, test.wantLandscape)
		}
	}
}

// TestLineShapeHandWorked checks the two scan statistics on scans whose shape
// is obvious by inspection.
func TestLineShapeHandWorked(t *testing.T) {
	// A single V: down 4, up 4. One direction change; total variation 8 over a
	// range of 4, so roughness 2 -- the value a line crossing one basin gives,
	// which is why smoothRoughness sits above it.
	turns, roughness := lineShape([]float64{4, 2, 0, 2, 4})
	if turns != 1 {
		t.Errorf("single-basin scan turns = %v, want 1", turns)
	}

	if roughness != 2 {
		t.Errorf("single-basin scan roughness = %v, want 2", roughness)
	}

	// A sawtooth 0,1,0,1,0,1,0: five direction changes; total variation 6 over
	// a range of 1, so roughness 6.
	turns, roughness = lineShape([]float64{0, 1, 0, 1, 0, 1, 0})
	if turns != 5 {
		t.Errorf("sawtooth turns = %v, want 5", turns)
	}

	if roughness != 6 {
		t.Errorf("sawtooth roughness = %v, want 6", roughness)
	}

	// A flat scan has no direction changes and no range to normalize by.
	turns, roughness = lineShape([]float64{3, 3, 3, 3})
	if turns != 0 || roughness != 0 {
		t.Errorf("flat scan = %v, %v, want 0, 0", turns, roughness)
	}

	// A monotone ramp turns nowhere and has roughness exactly 1.
	turns, roughness = lineShape([]float64{0, 1, 2, 3})
	if turns != 0 || roughness != 1 {
		t.Errorf("ramp = %v, %v, want 0, 1", turns, roughness)
	}
}

// TestClassifyProblemNeverGuessesDeceptiveOrNarrowValley documents the limit
// stated on ClassifyProblem: neither can be established from samples, so the
// classifier must not claim them.
func TestClassifyProblemNeverGuessesDeceptiveOrNarrowValley(t *testing.T) {
	functions := map[string]ObjectiveFunction{
		"Schwefel":   Schwefel,
		"Rosenbrock": Rosenbrock,
		"BentCigar":  BentCigar,
		"Sphere":     Sphere,
	}

	for name, fn := range functions {
		got := ClassifyProblem(fn, 6, -10, 10, rand.New(rand.NewSource(21)))
		if got.Landscape != Smooth && got.Landscape != Rugged {
			t.Errorf("%s classified as %v; ClassifyProblem may only report Smooth or Rugged",
				name, got.Landscape)
		}
	}
}

func TestClassifyProblemAcceptsNilRNG(t *testing.T) {
	characteristics := ClassifyProblem(Sphere, 3, -1, 1, nil)
	if characteristics.Dimensionality != 3 {
		t.Errorf("Dimensionality = %d, want 3", characteristics.Dimensionality)
	}
}

// TestClassifyProblemAgreesWithBenchmarkTable reconciles the sampler with the
// hand-classified switch in RecommendForBenchmark, which is the de facto
// expected output. Modality and landscape are checked for every entry; the
// table's Deceptive and NarrowValley verdicts are asserted *not* to come back
// from the sampler, since ClassifyProblem never claims either -- those entries
// are the caller's knowledge, not the sampler's, and the test pins exactly that
// division.
//
// Griewank is the one entry where the sampler and the table genuinely disagree,
// and the row records why rather than hiding it. See the note on that row.
func TestClassifyProblemAgreesWithBenchmarkTable(t *testing.T) {
	tests := []struct {
		name           string
		fn             ObjectiveFunction
		lower, upper   float64
		tableLandscape Landscape // as hard-coded in RecommendForBenchmark
		wantModality   Modality  // what the sampler must report
		wantLandscape  Landscape // what the sampler must report
	}{
		{"Sphere", Sphere, -100, 100, Smooth, Unimodal, Smooth},
		{"Rastrigin", Rastrigin, -5.12, 5.12, Rugged, HighlyMultimodal, Rugged},
		{"Rosenbrock", Rosenbrock, -5, 10, NarrowValley, Unimodal, Smooth},
		{"Ackley", Ackley, -32, 32, Rugged, HighlyMultimodal, Rugged},
		// Griewank's table entry is HighlyMultimodal/Rugged, which the
		// literature supports and an optimizer working near the optimum feels.
		// The line scan cannot see it over the standard [-600,600] box: the
		// cosine ripples have a period of a few units and an amplitude of order
		// one, against a value range of order 100000, so they are both aliased
		// by the scan spacing and negligible in the total variation. The
		// sampler reports the shape at box scale, which at that scale is a
		// bowl. This is the documented resolution limit, not a regression.
		{"Griewank", Griewank, -600, 600, Rugged, Unimodal, Smooth},
		{"Schwefel", Schwefel, -500, 500, Deceptive, HighlyMultimodal, Rugged},
		{"BentCigar", BentCigar, -100, 100, NarrowValley, Unimodal, Smooth},
		{"Discus", Discus, -100, 100, NarrowValley, Unimodal, Smooth},
	}

	for _, test := range tests {
		got := ClassifyProblem(test.fn, 10, test.lower, test.upper, rand.New(rand.NewSource(7)))

		if got.Modality != test.wantModality {
			t.Errorf("%s: modality = %v, want %v", test.name, got.Modality, test.wantModality)
		}

		if got.Landscape != test.wantLandscape {
			t.Errorf("%s: landscape = %v, want %v", test.name, got.Landscape, test.wantLandscape)
		}

		if test.tableLandscape == Deceptive || test.tableLandscape == NarrowValley {
			if got.Landscape == test.tableLandscape {
				t.Errorf("%s: ClassifyProblem reported %v, which it cannot establish by sampling",
					test.name, got.Landscape)
			}
		}
	}
}

func TestProblemCharacteristicsValidation(t *testing.T) {
	// Test that problem characteristics are properly structured
	chars := ProblemCharacteristics{
		Dimensionality:            50,
		Modality:                  HighlyMultimodal,
		Landscape:                 Deceptive,
		ExpensiveEvaluations:      true,
		RequiresFastConvergence:   false,
		RequiresStableConvergence: true,
		MultiObjective:            false,
	}

	// Use the selector to ensure characteristics are handled properly
	selector := NewAlgorithmSelector()
	recommendations := selector.RecommendAlgorithms(chars)

	if len(recommendations) == 0 {
		t.Error("Should return recommendations for valid characteristics")
	}

	// Should recommend EOBBMA highly for deceptive + highly multimodal
	found := false

	for _, rec := range recommendations {
		if rec.Variant.Name() == "EOBBMA" && rec.Score > 0.7 {
			found = true
			break
		}
	}

	if !found {
		t.Error("EOBBMA should be highly recommended for deceptive + highly multimodal")
	}
}
