package mayfly

import (
	"testing"
)

// =============================================================================
// Tests for variants.go - Unified Interface and Factory Pattern
// =============================================================================

func TestNewVariant(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Standard MA", "ma", "MA"},
		{"Standard alias", "standard", "MA"},
		{"DESMA", "desma", "DESMA"},
		{"OLCE", "olce", "OLCE-MA"},
		{"OLCE alias", "olce-ma", "OLCE-MA"},
		{"EOBBMA", "eobbma", "EOBBMA"},
		{"GSASMA", "gsasma", "GSASMA"},
		{"MPMA", "mpma", "MPMA"},
		{"AOBLMOA", "aoblmoa", "AOBLMOA"},
		{"Case insensitive", "DESMA", "DESMA"},
		{"With spaces", " ma ", "MA"},
		{"Unknown", "unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			variant := NewVariant(tt.input)
			if tt.expected == "" {
				if variant != nil {
					t.Errorf("Expected nil for unknown variant, got %v", variant)
				}
			} else {
				if variant == nil {
					t.Fatalf("Expected variant %s, got nil", tt.expected)
				}
				if variant.Name() != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, variant.Name())
				}
			}
		})
	}
}

func TestListVariants(t *testing.T) {
	variants := ListVariants()

	// Should have exactly 7 variants (excluding aliases)
	if len(variants) != 7 {
		t.Errorf("Expected 7 variants, got %d", len(variants))
	}

	// Check for required variants
	required := map[string]bool{
		"ma": false, "desma": false, "olce": false, "eobbma": false,
		"gsasma": false, "mpma": false, "aoblmoa": false,
	}

	for _, name := range variants {
		required[name] = true
	}

	for name, found := range required {
		if !found {
			t.Errorf("Variant %s not found in list", name)
		}
	}
}

func TestGetAllVariants(t *testing.T) {
	variants := GetAllVariants()

	// Should have exactly 7 unique variants
	if len(variants) != 7 {
		t.Errorf("Expected 7 variants, got %d", len(variants))
	}

	// Each should have valid methods
	for _, v := range variants {
		if v.Name() == "" {
			t.Error("Variant name should not be empty")
		}
		if v.FullName() == "" {
			t.Error("Variant full name should not be empty")
		}
		if v.Description() == "" {
			t.Error("Variant description should not be empty")
		}
		if len(v.RecommendedFor()) == 0 {
			t.Errorf("Variant %s should have recommended use cases", v.Name())
		}
		if v.EstimatedOverhead() < 1.0 {
			t.Errorf("Variant %s overhead should be >= 1.0, got %.2f", v.Name(), v.EstimatedOverhead())
		}
	}
}

func TestVariantApplicability(t *testing.T) {
	// Test that variants score appropriately for their target problems

	// DESMA should score high for multimodal
	desma := NewVariant("desma")
	multimodal := ProblemCharacteristics{
		Modality:  Multimodal,
		Landscape: Rugged,
	}
	score := desma.ApplicableTo(multimodal)
	if score < 0.7 {
		t.Errorf("DESMA should score high (>0.7) for multimodal, got %.2f", score)
	}

	// EOBBMA should score high for deceptive
	eobbma := NewVariant("eobbma")
	deceptive := ProblemCharacteristics{
		Landscape: Deceptive,
		Modality:  HighlyMultimodal,
	}
	score = eobbma.ApplicableTo(deceptive)
	if score < 0.8 {
		t.Errorf("EOBBMA should score high (>0.8) for deceptive, got %.2f", score)
	}

	// MPMA should score high for stable convergence needs
	mpma := NewVariant("mpma")
	stable := ProblemCharacteristics{
		RequiresStableConvergence: true,
		Landscape:                 NarrowValley,
	}
	score = mpma.ApplicableTo(stable)
	if score < 0.8 {
		t.Errorf("MPMA should score high (>0.8) for stable convergence, got %.2f", score)
	}

	// AOBLMOA essential for multi-objective
	aoblmoa := NewVariant("aoblmoa")
	multiObj := ProblemCharacteristics{
		MultiObjective: true,
	}
	score = aoblmoa.ApplicableTo(multiObj)
	if score < 0.9 {
		t.Errorf("AOBLMOA should score very high (>0.9) for multi-objective, got %.2f", score)
	}
}

func TestVariantBuilder(t *testing.T) {
	// Test fluent builder API
	builder := NewBuilder("desma")
	if builder == nil {
		t.Fatal("Builder should not be nil for valid variant")
	}

	config, err := builder.
		ForProblem(Sphere, 10, -10, 10).
		WithIterations(100).
		WithPopulation(20, 20).
		Build()

	if err != nil {
		t.Fatalf("Builder failed: %v", err)
	}

	if config.ProblemSize != 10 {
		t.Errorf("Expected ProblemSize 10, got %d", config.ProblemSize)
	}
	if config.MaxIterations != 100 {
		t.Errorf("Expected MaxIterations 100, got %d", config.MaxIterations)
	}
	if config.NPop != 20 {
		t.Errorf("Expected NPop 20, got %d", config.NPop)
	}
	if !config.UseDESMA {
		t.Error("Expected UseDESMA to be true")
	}
}

func TestVariantBuilderWithConfig(t *testing.T) {
	// Test custom config modification
	builder := NewBuilder("gsasma")
	config, err := builder.
		ForProblem(Rastrigin, 20, -5.12, 5.12).
		WithIterations(200).
		WithConfig(func(c *Config) {
			c.CoolingRate = 0.97
			c.InitialTemperature = 150.0
		}).
		Build()

	if err != nil {
		t.Fatalf("Builder with custom config failed: %v", err)
	}

	if config.CoolingRate != 0.97 {
		t.Errorf("Expected CoolingRate 0.97, got %f", config.CoolingRate)
	}
	if config.InitialTemperature != 150.0 {
		t.Errorf("Expected InitialTemperature 150.0, got %f", config.InitialTemperature)
	}
}

func TestVariantBuilderErrors(t *testing.T) {
	// Nil builder for unknown variant
	builder := NewBuilder("unknown")
	if builder != nil {
		t.Error("Builder should be nil for unknown variant")
	}

	// Missing objective function
	builder = NewBuilder("ma")
	builder.config.ObjectiveFunc = nil // Force nil
	_, err := builder.Build()
	if err == nil {
		t.Error("Build should fail without objective function")
	}

	// Invalid problem size
	builder = NewBuilder("ma").ForProblem(Sphere, -1, -10, 10)
	_, err = builder.Build()
	if err == nil {
		t.Error("Build should fail with negative problem size")
	}
}

func TestNewBuilderFromVariant(t *testing.T) {
	variant := NewVariant("olce")
	builder := NewBuilderFromVariant(variant)

	if builder == nil {
		t.Fatal("Builder should not be nil")
	}

	if builder.GetVariant() != variant {
		t.Error("Builder should return the same variant")
	}

	config, err := builder.
		ForProblem(Sphere, 15, -10, 10).
		WithIterations(300).
		Build()

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if !config.UseOLCE {
		t.Error("Expected UseOLCE to be true")
	}
}
