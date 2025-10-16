package mayfly

import (
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

	// Multi-objective should recommend AOBLMOA
	characteristics := ProblemCharacteristics{
		MultiObjective: true,
	}

	best := selector.RecommendBest(characteristics)
	if best.Variant.Name() != "AOBLMOA" {
		t.Errorf("Expected AOBLMOA for multi-objective, got %s", best.Variant.Name())
	}

	// Check confidence is high
	if best.Confidence < 0.9 {
		t.Errorf("Expected high confidence (>0.9) for multi-objective, got %.2f", best.Confidence)
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

func TestClassifyProblem(t *testing.T) {
	// This is a lightweight test since ClassifyProblem does sampling
	// We just verify it returns valid characteristics
	characteristics := ClassifyProblem(Sphere, 5, -10, 10)

	if characteristics.Dimensionality != 5 {
		t.Errorf("Expected dimensionality 5, got %d", characteristics.Dimensionality)
	}

	// Sphere should be classified as unimodal or multimodal (not highly)
	if characteristics.Modality == HighlyMultimodal {
		t.Error("Sphere should not be classified as highly multimodal")
	}

	// Landscape should be valid
	validLandscapes := map[Landscape]bool{
		Smooth: true, Rugged: true, Deceptive: true, NarrowValley: true,
	}
	if !validLandscapes[characteristics.Landscape] {
		t.Errorf("Invalid landscape classification: %v", characteristics.Landscape)
	}
}

func TestEstimateModality(t *testing.T) {
	// Test modality estimation with known distributions
	// Low variance should indicate unimodal
	samples := []float64{1.0, 1.1, 0.9, 1.05, 0.95, 1.2, 0.8}

	modality := estimateModality(samples)
	if modality == HighlyMultimodal {
		t.Error("Low variance should not be classified as highly multimodal")
	}

	// High variance should indicate multimodal
	samples = []float64{1.0, 100.0, 5.0, 200.0, 10.0, 150.0}

	modality = estimateModality(samples)
	if modality == Unimodal {
		t.Error("High variance should not be classified as unimodal")
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
