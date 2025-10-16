package mayfly

import (
	"math"
	"testing"
)

// TestGravityCoefficient tests non-linear gravity coefficient functions.
func TestGravityCoefficient(t *testing.T) {
	tests := []struct {
		name           string
		gravityType    string
		iteration      int
		maxIterations  int
		expectedRange  [2]float64 // min, max expected values
		checkMonotonic bool       // should decrease monotonically
	}{
		{
			name:           "linear decay",
			gravityType:    "linear",
			iteration:      250,
			maxIterations:  500,
			expectedRange:  [2]float64{0.4, 0.6}, // at 50%, should be around 0.5
			checkMonotonic: true,
		},
		{
			name:           "exponential decay",
			gravityType:    "exponential",
			iteration:      250,
			maxIterations:  500,
			expectedRange:  [2]float64{0.13, 0.14}, // e^(-2) ≈ 0.135
			checkMonotonic: true,
		},
		{
			name:           "sigmoid decay",
			gravityType:    "sigmoid",
			iteration:      250,
			maxIterations:  500,
			expectedRange:  [2]float64{0.2, 0.8}, // sigmoid has S-curve
			checkMonotonic: true,
		},
		{
			name:          "linear at start",
			gravityType:   "linear",
			iteration:     0,
			maxIterations: 500,
			expectedRange: [2]float64{0.95, 1.0}, // should start near 1.0
		},
		{
			name:          "linear at end",
			gravityType:   "linear",
			iteration:     500,
			maxIterations: 500,
			expectedRange: [2]float64{0.0, 0.05}, // should end near 0.0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := calculateGravityCoefficient(tt.gravityType, tt.iteration, tt.maxIterations)

			// Check range
			if g < tt.expectedRange[0] || g > tt.expectedRange[1] {
				t.Errorf("expected gravity in range [%f, %f], got %f", tt.expectedRange[0], tt.expectedRange[1], g)
			}

			// Check monotonic decrease if required
			if tt.checkMonotonic && tt.iteration < tt.maxIterations-1 {
				gNext := calculateGravityCoefficient(tt.gravityType, tt.iteration+1, tt.maxIterations)
				if gNext > g {
					t.Errorf("expected monotonic decrease, but g(%d)=%f > g(%d)=%f", tt.iteration+1, gNext, tt.iteration, g)
				}
			}
		})
	}
}

// TestGravityCoefficientInvalidType tests handling of invalid gravity types.
func TestGravityCoefficientInvalidType(t *testing.T) {
	// Should default to linear for unknown types
	g := calculateGravityCoefficient("unknown", 250, 500)
	expected := calculateGravityCoefficient("linear", 250, 500)

	if math.Abs(g-expected) > 1e-10 {
		t.Errorf("expected default to linear, got different value: %f vs %f", g, expected)
	}
}

// TestNewMPMAConfig tests the MPMA configuration factory.
func TestNewMPMAConfig(t *testing.T) {
	config := NewMPMAConfig()

	// Verify MPMA is enabled
	if !config.UseMPMA {
		t.Error("expected UseMPMA to be true")
	}

	// Verify default median weight
	if config.MedianWeight <= 0 || config.MedianWeight > 1 {
		t.Errorf("expected MedianWeight in (0,1], got %f", config.MedianWeight)
	}

	// Verify default gravity type
	validTypes := []string{"linear", "exponential", "sigmoid"}
	found := false

	for _, validType := range validTypes {
		if config.GravityType == validType {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected GravityType to be one of %v, got %s", validTypes, config.GravityType)
	}

	// Verify base parameters are set
	if config.NPop == 0 {
		t.Error("expected NPop to be set")
	}

	if config.MaxIterations == 0 {
		t.Error("expected MaxIterations to be set")
	}
}

// TestMPMAOptimization tests MPMA on a simple optimization problem.
func TestMPMAOptimization(t *testing.T) {
	config := NewMPMAConfig()
	config.ObjectiveFunc = Sphere
	config.ProblemSize = 5
	config.LowerBound = -10
	config.UpperBound = 10
	config.MaxIterations = 100
	config.NPop = 10
	config.NPopF = 10

	result, err := Optimize(config)
	if err != nil {
		t.Fatalf("optimization failed: %v", err)
	}

	// Should converge reasonably well on Sphere function
	if result.GlobalBest.Cost > 1e-2 {
		t.Errorf("expected better convergence, got cost %e", result.GlobalBest.Cost)
	}

	// Verify result structure
	if len(result.GlobalBest.Position) != config.ProblemSize {
		t.Errorf("expected position length %d, got %d", config.ProblemSize, len(result.GlobalBest.Position))
	}

	if len(result.BestSolution) != config.MaxIterations {
		t.Errorf("expected solution history length %d, got %d", config.MaxIterations, len(result.BestSolution))
	}
}

// TestCalculateMedianPosition tests median position calculation.
func TestCalculateMedianPosition(t *testing.T) {
	tests := []struct {
		name       string
		population []*Mayfly
		expected   []float64
	}{
		{
			name: "odd number of mayflies",
			population: []*Mayfly{
				{Position: []float64{1.0, 2.0}},
				{Position: []float64{3.0, 4.0}},
				{Position: []float64{5.0, 6.0}},
			},
			expected: []float64{3.0, 4.0}, // median of [1,3,5] = 3, median of [2,4,6] = 4
		},
		{
			name: "even number of mayflies",
			population: []*Mayfly{
				{Position: []float64{1.0, 2.0}},
				{Position: []float64{3.0, 4.0}},
				{Position: []float64{5.0, 6.0}},
				{Position: []float64{7.0, 8.0}},
			},
			expected: []float64{4.0, 5.0}, // median of [1,3,5,7] = (3+5)/2 = 4, median of [2,4,6,8] = (4+6)/2 = 5
		},
		{
			name: "single mayfly",
			population: []*Mayfly{
				{Position: []float64{3.5, 7.2}},
			},
			expected: []float64{3.5, 7.2},
		},
		{
			name: "higher dimensions",
			population: []*Mayfly{
				{Position: []float64{1.0, 2.0, 3.0}},
				{Position: []float64{4.0, 5.0, 6.0}},
				{Position: []float64{7.0, 8.0, 9.0}},
			},
			expected: []float64{4.0, 5.0, 6.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateMedianPosition(tt.population)

			if len(result) != len(tt.expected) {
				t.Fatalf("expected length %d, got %d", len(tt.expected), len(result))
			}

			for i := range result {
				if math.Abs(result[i]-tt.expected[i]) > 1e-10 {
					t.Errorf("dimension %d: expected %f, got %f", i, tt.expected[i], result[i])
				}
			}
		})
	}
}

// TestCalculateWeightedMedianPosition tests weighted median calculation.
func TestCalculateWeightedMedianPosition(t *testing.T) {
	tests := []struct {
		name       string
		population []*Mayfly
		weights    []float64
		expected   []float64
	}{
		{
			name: "equal weights same as regular median",
			population: []*Mayfly{
				{Position: []float64{1.0}},
				{Position: []float64{3.0}},
				{Position: []float64{5.0}},
			},
			weights:  []float64{1.0, 1.0, 1.0},
			expected: []float64{3.0},
		},
		{
			name: "weighted toward better solutions",
			population: []*Mayfly{
				{Position: []float64{1.0}},
				{Position: []float64{3.0}},
				{Position: []float64{5.0}},
			},
			weights:  []float64{3.0, 2.0, 1.0}, // weight first solution more (total=6, half=3)
			expected: []float64{1.0},           // cumulative weight reaches 3 at value 1.0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateWeightedMedianPosition(tt.population, tt.weights)

			if len(result) != len(tt.expected) {
				t.Fatalf("expected length %d, got %d", len(tt.expected), len(result))
			}

			for i := range result {
				if math.Abs(result[i]-tt.expected[i]) > 1e-10 {
					t.Errorf("dimension %d: expected %f, got %f", i, tt.expected[i], result[i])
				}
			}
		})
	}
}
