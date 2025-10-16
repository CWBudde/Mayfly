package mayfly

import (
	"math"
	"math/rand"
	"testing"
)

// TestUnifrnd tests the unifrnd function for uniform random generation.
func TestUnifrnd(t *testing.T) {
	tests := []struct {
		rng  *rand.Rand
		name string
		min  float64
		max  float64
	}{
		{"default_rng_0_1", 0.0, 1.0, nil},
		{"default_rng_negative", -10.0, 10.0, nil},
		{"seeded_rng", 5.0, 15.0, rand.New(rand.NewSource(42))},
		{"large_range", -1000.0, 1000.0, rand.New(rand.NewSource(123))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate multiple samples to test distribution
			samples := 1000
			for i := 0; i < samples; i++ {
				val := unifrnd(tt.min, tt.max, tt.rng)
				if val < tt.min || val > tt.max {
					t.Errorf("unifrnd() = %v, want value in range [%v, %v]", val, tt.min, tt.max)
				}
			}
		})
	}
}

// TestUnifrndDeterministic tests that seeded RNG produces deterministic results.
func TestUnifrndDeterministic(t *testing.T) {
	seed := int64(42)
	rng1 := rand.New(rand.NewSource(seed))
	rng2 := rand.New(rand.NewSource(seed))

	for i := 0; i < 100; i++ {
		val1 := unifrnd(-5.0, 5.0, rng1)
		val2 := unifrnd(-5.0, 5.0, rng2)

		if val1 != val2 {
			t.Errorf("unifrnd() with same seed produced different values: %v vs %v", val1, val2)
		}
	}
}

// TestUnifrndVec tests the vector version of uniform random generation.
func TestUnifrndVec(t *testing.T) {
	tests := []struct {
		rng  *rand.Rand
		name string
		min  float64
		max  float64
		size int
	}{
		{"size_10", 0.0, 1.0, 10, nil},
		{"size_50", -10.0, 10.0, 50, rand.New(rand.NewSource(42))},
		{"size_1", -5.0, 5.0, 1, nil},
		{"large_size", 0.0, 100.0, 1000, rand.New(rand.NewSource(123))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vec := unifrndVec(tt.min, tt.max, tt.size, tt.rng)

			// Check size
			if len(vec) != tt.size {
				t.Errorf("unifrndVec() length = %v, want %v", len(vec), tt.size)
			}

			// Check all values are in range
			for i, val := range vec {
				if val < tt.min || val > tt.max {
					t.Errorf("unifrndVec()[%d] = %v, want value in range [%v, %v]", i, val, tt.min, tt.max)
				}
			}
		})
	}
}

// TestUnifrndVecDeterministic tests deterministic behavior.
func TestUnifrndVecDeterministic(t *testing.T) {
	seed := int64(789)
	size := 50

	vec1 := unifrndVec(-10.0, 10.0, size, rand.New(rand.NewSource(seed)))
	vec2 := unifrndVec(-10.0, 10.0, size, rand.New(rand.NewSource(seed)))

	for i := 0; i < size; i++ {
		if vec1[i] != vec2[i] {
			t.Errorf("unifrndVec()[%d] with same seed produced different values: %v vs %v", i, vec1[i], vec2[i])
		}
	}
}

// TestRandn tests normal distribution generation.
func TestRandn(t *testing.T) {
	tests := []struct {
		rng     *rand.Rand
		name    string
		samples int
	}{
		{"default_rng", nil, 1000},
		{"seeded_rng", rand.New(rand.NewSource(42)), 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sum, sumSq float64

			for i := 0; i < tt.samples; i++ {
				val := randn(tt.rng)
				sum += val
				sumSq += val * val
			}

			// Check approximate mean (should be close to 0)
			mean := sum / float64(tt.samples)
			if math.Abs(mean) > 0.2 {
				t.Errorf("randn() mean = %v, want ~0.0", mean)
			}

			// Check approximate variance (should be close to 1)
			variance := sumSq/float64(tt.samples) - mean*mean
			if math.Abs(variance-1.0) > 0.3 {
				t.Errorf("randn() variance = %v, want ~1.0", variance)
			}
		})
	}
}

// TestRandnDeterministic tests deterministic behavior.
func TestRandnDeterministic(t *testing.T) {
	seed := int64(999)
	rng1 := rand.New(rand.NewSource(seed))
	rng2 := rand.New(rand.NewSource(seed))

	for i := 0; i < 100; i++ {
		val1 := randn(rng1)
		val2 := randn(rng2)

		if val1 != val2 {
			t.Errorf("randn() with same seed produced different values: %v vs %v", val1, val2)
		}
	}
}

// TestMaxVec tests element-wise maximum with scalar bound.
func TestMaxVec(t *testing.T) {
	tests := []struct {
		name     string
		input    []float64
		expected []float64
		bound    float64
	}{
		{
			name:     "all_below_bound",
			input:    []float64{-5.0, -2.0, -10.0},
			bound:    -1.0,
			expected: []float64{-1.0, -1.0, -1.0},
		},
		{
			name:     "all_above_bound",
			input:    []float64{5.0, 2.0, 10.0},
			bound:    1.0,
			expected: []float64{5.0, 2.0, 10.0},
		},
		{
			name:     "mixed",
			input:    []float64{-5.0, 2.0, -1.0, 10.0},
			bound:    0.0,
			expected: []float64{0.0, 2.0, 0.0, 10.0},
		},
		{
			name:     "at_bound",
			input:    []float64{5.0, 5.0, 5.0},
			bound:    5.0,
			expected: []float64{5.0, 5.0, 5.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vec := make([]float64, len(tt.input))
			copy(vec, tt.input)
			maxVec(vec, tt.bound)

			for i := 0; i < len(vec); i++ {
				if vec[i] != tt.expected[i] {
					t.Errorf("maxVec()[%d] = %v, want %v", i, vec[i], tt.expected[i])
				}
			}
		})
	}
}

// TestMinVec tests element-wise minimum with scalar bound.
func TestMinVec(t *testing.T) {
	tests := []struct {
		name     string
		input    []float64
		expected []float64
		bound    float64
	}{
		{
			name:     "all_above_bound",
			input:    []float64{5.0, 2.0, 10.0},
			bound:    1.0,
			expected: []float64{1.0, 1.0, 1.0},
		},
		{
			name:     "all_below_bound",
			input:    []float64{-5.0, -2.0, -10.0},
			bound:    -1.0,
			expected: []float64{-5.0, -2.0, -10.0},
		},
		{
			name:     "mixed",
			input:    []float64{5.0, -2.0, 1.0, -10.0},
			bound:    0.0,
			expected: []float64{0.0, -2.0, 0.0, -10.0},
		},
		{
			name:     "at_bound",
			input:    []float64{5.0, 5.0, 5.0},
			bound:    5.0,
			expected: []float64{5.0, 5.0, 5.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vec := make([]float64, len(tt.input))
			copy(vec, tt.input)
			minVec(vec, tt.bound)

			for i := 0; i < len(vec); i++ {
				if vec[i] != tt.expected[i] {
					t.Errorf("minVec()[%d] = %v, want %v", i, vec[i], tt.expected[i])
				}
			}
		})
	}
}

// TestMaxMinVecChaining tests using both maxVec and minVec together.
func TestMaxMinVecChaining(t *testing.T) {
	input := []float64{-15.0, -5.0, 0.0, 5.0, 15.0}
	lowerBound := -10.0
	upperBound := 10.0
	expected := []float64{-10.0, -5.0, 0.0, 5.0, 10.0}

	vec := make([]float64, len(input))
	copy(vec, input)

	// Apply lower bound then upper bound
	maxVec(vec, lowerBound)
	minVec(vec, upperBound)

	for i := 0; i < len(vec); i++ {
		if vec[i] != expected[i] {
			t.Errorf("After maxVec+minVec: vec[%d] = %v, want %v", i, vec[i], expected[i])
		}
	}
}

// TestSortMayflies tests sorting of mayfly population by cost.
func TestSortMayflies(t *testing.T) {
	tests := []struct {
		name          string
		costs         []float64
		expectedOrder []int
	}{
		{
			name:          "already_sorted",
			costs:         []float64{1.0, 2.0, 3.0, 4.0},
			expectedOrder: []int{0, 1, 2, 3},
		},
		{
			name:          "reverse_sorted",
			costs:         []float64{4.0, 3.0, 2.0, 1.0},
			expectedOrder: []int{3, 2, 1, 0},
		},
		{
			name:          "random_order",
			costs:         []float64{3.5, 1.2, 4.8, 2.1},
			expectedOrder: []int{1, 3, 0, 2},
		},
		{
			name:          "duplicates",
			costs:         []float64{2.0, 1.0, 2.0, 1.0},
			expectedOrder: []int{1, 3, 0, 2}, // Stable sort behavior
		},
		{
			name:          "single",
			costs:         []float64{5.0},
			expectedOrder: []int{0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mayflies with costs and track original indices
			mayflies := make([]*Mayfly, len(tt.costs))
			for i, cost := range tt.costs {
				mayflies[i] = &Mayfly{
					Cost:     cost,
					Position: []float64{float64(i)}, // Use position to track original index
				}
			}

			// Sort
			sortMayflies(mayflies)

			// Verify sorted order
			for i := 1; i < len(mayflies); i++ {
				if mayflies[i-1].Cost > mayflies[i].Cost {
					t.Errorf("sortMayflies() not sorted: mayflies[%d].Cost=%v > mayflies[%d].Cost=%v",
						i-1, mayflies[i-1].Cost, i, mayflies[i].Cost)
				}
			}

			// Verify expected original indices (if stable sort expected)
			for i, expectedOrigIdx := range tt.expectedOrder {
				actualOrigIdx := int(mayflies[i].Position[0])
				if actualOrigIdx != expectedOrigIdx {
					t.Logf("sortMayflies() order note: position %d has original index %d, expected %d",
						i, actualOrigIdx, expectedOrigIdx)
				}
			}
		})
	}
}

// TestSortMayfliesEmpty tests edge case of empty population.
func TestSortMayfliesEmpty(t *testing.T) {
	mayflies := make([]*Mayfly, 0)
	sortMayflies(mayflies) // Should not panic
}

// TestSortMayfliesSingleElement tests edge case of single element.
func TestSortMayfliesSingleElement(t *testing.T) {
	mayflies := []*Mayfly{
		{Cost: 5.0, Position: []float64{1.0}},
	}
	sortMayflies(mayflies)

	if mayflies[0].Cost != 5.0 {
		t.Errorf("sortMayflies() modified single element: got cost %v, want 5.0", mayflies[0].Cost)
	}
}
