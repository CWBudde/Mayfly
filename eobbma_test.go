package mayfly

import (
	"math"
	"math/rand"
	"testing"
)

// TestLevyFlight tests the Lévy flight distribution generator.
func TestLevyFlight(t *testing.T) {
	tests := []struct {
		rng   *rand.Rand
		name  string
		alpha float64
		beta  float64
	}{
		{"standard_levy", 1.5, 1.0, rand.New(rand.NewSource(42))},
		{"alpha_2.0", 2.0, 1.0, rand.New(rand.NewSource(123))},
		{"alpha_1.0", 1.0, 1.0, rand.New(rand.NewSource(456))},
		{"beta_0.5", 1.5, 0.5, rand.New(rand.NewSource(789))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate multiple samples
			samples := 1000

			var values []float64

			for i := 0; i < samples; i++ {
				val := levyFlight(tt.alpha, tt.beta, tt.rng)
				values = append(values, val)

				// Lévy flight should produce finite values
				if math.IsNaN(val) || math.IsInf(val, 0) {
					t.Errorf("levyFlight() produced non-finite value: %v", val)
				}
			}

			// Check that distribution has heavy tails (some large values)
			// At least one value should be > 5 in 1000 samples for heavy-tailed distribution
			hasLargeValue := false

			for _, v := range values {
				if math.Abs(v) > 5.0 {
					hasLargeValue = true
					break
				}
			}

			if !hasLargeValue {
				t.Logf("Warning: levyFlight() may not be generating heavy-tailed distribution")
			}
		})
	}
}

// TestLevyFlightDeterministic tests that seeded RNG produces deterministic results.
func TestLevyFlightDeterministic(t *testing.T) {
	seed := int64(42)
	alpha := 1.5
	beta := 1.0

	rng1 := rand.New(rand.NewSource(seed))
	rng2 := rand.New(rand.NewSource(seed))

	for i := 0; i < 100; i++ {
		val1 := levyFlight(alpha, beta, rng1)
		val2 := levyFlight(alpha, beta, rng2)

		if val1 != val2 {
			t.Errorf("levyFlight() with same seed produced different values: %v vs %v", val1, val2)
		}
	}
}

// TestLevyFlightVector tests vector Lévy flight generation.
func TestLevyFlightVector(t *testing.T) {
	tests := []struct {
		rng   *rand.Rand
		name  string
		size  int
		alpha float64
		beta  float64
	}{
		{"size_10", 10, 1.5, 1.0, rand.New(rand.NewSource(42))},
		{"size_50", 50, 1.5, 1.0, rand.New(rand.NewSource(123))},
		{"size_1", 1, 1.5, 1.0, rand.New(rand.NewSource(456))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vec := levyFlightVec(tt.size, tt.alpha, tt.beta, tt.rng)

			// Check size
			if len(vec) != tt.size {
				t.Errorf("levyFlightVec() length = %v, want %v", len(vec), tt.size)
			}

			// Check all values are finite
			for i, val := range vec {
				if math.IsNaN(val) || math.IsInf(val, 0) {
					t.Errorf("levyFlightVec()[%d] = %v, want finite value", i, val)
				}
			}
		})
	}
}

// TestOppositionLearning tests the opposition-based learning operator.
func TestOppositionLearning(t *testing.T) {
	tests := []struct {
		name       string
		position   []float64
		expected   []float64
		lowerBound float64
		upperBound float64
	}{
		{
			name:       "center_point",
			position:   []float64{0.0, 0.0, 0.0},
			lowerBound: -10.0,
			upperBound: 10.0,
			expected:   []float64{0.0, 0.0, 0.0}, // Opposition of center is center
		},
		{
			name:       "lower_bound",
			position:   []float64{-10.0, -10.0},
			lowerBound: -10.0,
			upperBound: 10.0,
			expected:   []float64{10.0, 10.0},
		},
		{
			name:       "upper_bound",
			position:   []float64{10.0, 10.0},
			lowerBound: -10.0,
			upperBound: 10.0,
			expected:   []float64{-10.0, -10.0},
		},
		{
			name:       "arbitrary_point",
			position:   []float64{5.0, -3.0, 2.0},
			lowerBound: -10.0,
			upperBound: 10.0,
			expected:   []float64{-5.0, 3.0, -2.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := oppositionPoint(tt.position, tt.lowerBound, tt.upperBound)

			if len(result) != len(tt.expected) {
				t.Errorf("oppositionPoint() length = %v, want %v", len(result), len(tt.expected))
			}

			for i := 0; i < len(result); i++ {
				if math.Abs(result[i]-tt.expected[i]) > 1e-10 {
					t.Errorf("oppositionPoint()[%d] = %v, want %v", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

// TestGaussianUpdate tests the Bare Bones Gaussian-based update.
func TestGaussianUpdate(t *testing.T) {
	tests := []struct {
		rng        *rand.Rand
		name       string
		current    []float64
		best       []float64
		lowerBound float64
		upperBound float64
	}{
		{
			name:       "standard_case",
			current:    []float64{1.0, 2.0, 3.0},
			best:       []float64{0.0, 0.0, 0.0},
			lowerBound: -10.0,
			upperBound: 10.0,
			rng:        rand.New(rand.NewSource(42)),
		},
		{
			name:       "same_position",
			current:    []float64{5.0, 5.0},
			best:       []float64{5.0, 5.0},
			lowerBound: -10.0,
			upperBound: 10.0,
			rng:        rand.New(rand.NewSource(123)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gaussianUpdate(tt.current, tt.best, tt.lowerBound, tt.upperBound, tt.rng)

			// Check size
			if len(result) != len(tt.current) {
				t.Errorf("gaussianUpdate() length = %v, want %v", len(result), len(tt.current))
			}

			// Check bounds
			for i, val := range result {
				if val < tt.lowerBound || val > tt.upperBound {
					t.Errorf("gaussianUpdate()[%d] = %v, out of bounds [%v, %v]",
						i, val, tt.lowerBound, tt.upperBound)
				}
			}

			// Check not all zeros (very unlikely with Gaussian)
			allSame := true

			for i := 1; i < len(result); i++ {
				if result[i] != result[0] {
					allSame = false
					break
				}
			}

			if allSame && len(result) > 1 {
				t.Logf("Warning: gaussianUpdate() produced identical values across all dimensions")
			}
		})
	}
}

// TestGaussianUpdateDeterministic tests deterministic behavior with seeded RNG.
func TestGaussianUpdateDeterministic(t *testing.T) {
	seed := int64(999)
	current := []float64{1.0, 2.0, 3.0}
	best := []float64{0.0, 0.0, 0.0}
	lowerBound := -10.0
	upperBound := 10.0

	rng1 := rand.New(rand.NewSource(seed))
	rng2 := rand.New(rand.NewSource(seed))

	result1 := gaussianUpdate(current, best, lowerBound, upperBound, rng1)
	result2 := gaussianUpdate(current, best, lowerBound, upperBound, rng2)

	for i := 0; i < len(result1); i++ {
		if result1[i] != result2[i] {
			t.Errorf("gaussianUpdate()[%d] with same seed produced different values: %v vs %v",
				i, result1[i], result2[i])
		}
	}
}
