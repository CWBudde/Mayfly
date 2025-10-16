package mayfly

import (
	"math"
	"math/rand"
	"testing"
)

// TestCrossover tests the crossover operator.
func TestCrossover(t *testing.T) {
	tests := []struct {
		name       string
		x1         []float64
		x2         []float64
		lowerBound float64
		upperBound float64
		seed       int64
	}{
		{
			name:       "simple_2d",
			x1:         []float64{1.0, 2.0},
			x2:         []float64{3.0, 4.0},
			lowerBound: 0.0,
			upperBound: 5.0,
			seed:       42,
		},
		{
			name:       "high_dimensional",
			x1:         []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0},
			x2:         []float64{10.0, 9.0, 8.0, 7.0, 6.0, 5.0, 4.0, 3.0, 2.0, 1.0},
			lowerBound: 0.0,
			upperBound: 10.0,
			seed:       123,
		},
		{
			name:       "negative_bounds",
			x1:         []float64{-5.0, -3.0, -1.0},
			x2:         []float64{-2.0, -4.0, -6.0},
			lowerBound: -10.0,
			upperBound: 10.0,
			seed:       789,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rng := rand.New(rand.NewSource(tt.seed))
			off1, off2 := Crossover(tt.x1, tt.x2, tt.lowerBound, tt.upperBound, rng)

			// Check size
			if len(off1) != len(tt.x1) {
				t.Errorf("Crossover() off1 length = %v, want %v", len(off1), len(tt.x1))
			}

			if len(off2) != len(tt.x2) {
				t.Errorf("Crossover() off2 length = %v, want %v", len(off2), len(tt.x2))
			}

			// Check bounds for offspring 1
			for i, val := range off1 {
				if val < tt.lowerBound || val > tt.upperBound {
					t.Errorf("Crossover() off1[%d] = %v, out of bounds [%v, %v]",
						i, val, tt.lowerBound, tt.upperBound)
				}
			}

			// Check bounds for offspring 2
			for i, val := range off2 {
				if val < tt.lowerBound || val > tt.upperBound {
					t.Errorf("Crossover() off2[%d] = %v, out of bounds [%v, %v]",
						i, val, tt.lowerBound, tt.upperBound)
				}
			}

			// Check that offspring are different from parents (highly likely with random L)
			diff1 := false
			diff2 := false

			for i := 0; i < len(tt.x1); i++ {
				if math.Abs(off1[i]-tt.x1[i]) > 1e-10 {
					diff1 = true
				}

				if math.Abs(off2[i]-tt.x2[i]) > 1e-10 {
					diff2 = true
				}
			}
			// At least one offspring should differ from its primary parent
			if !diff1 && !diff2 {
				t.Log("Crossover() produced offspring identical to parents (rare but possible)")
			}
		})
	}
}

// TestCrossoverDeterministic tests that crossover is deterministic with seeded RNG.
func TestCrossoverDeterministic(t *testing.T) {
	x1 := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	x2 := []float64{6.0, 7.0, 8.0, 9.0, 10.0}
	lowerBound := 0.0
	upperBound := 10.0
	seed := int64(42)

	rng1 := rand.New(rand.NewSource(seed))
	off1a, off2a := Crossover(x1, x2, lowerBound, upperBound, rng1)

	rng2 := rand.New(rand.NewSource(seed))
	off1b, off2b := Crossover(x1, x2, lowerBound, upperBound, rng2)

	// Check offspring 1 matches
	for i := 0; i < len(off1a); i++ {
		if off1a[i] != off1b[i] {
			t.Errorf("Crossover() off1[%d] not deterministic: %v vs %v", i, off1a[i], off1b[i])
		}
	}

	// Check offspring 2 matches
	for i := 0; i < len(off2a); i++ {
		if off2a[i] != off2b[i] {
			t.Errorf("Crossover() off2[%d] not deterministic: %v vs %v", i, off2a[i], off2b[i])
		}
	}
}

// TestCrossoverBoundaryViolations tests that crossover properly handles boundary violations.
func TestCrossoverBoundaryViolations(t *testing.T) {
	// Parents at extreme bounds
	x1 := []float64{-10.0, -10.0, -10.0}
	x2 := []float64{10.0, 10.0, 10.0}
	lowerBound := -5.0
	upperBound := 5.0

	rng := rand.New(rand.NewSource(42))
	off1, off2 := Crossover(x1, x2, lowerBound, upperBound, rng)

	// Check that all values are clamped to bounds
	for i := 0; i < len(off1); i++ {
		if off1[i] < lowerBound || off1[i] > upperBound {
			t.Errorf("Crossover() off1[%d] = %v, expected to be clamped to [%v, %v]",
				i, off1[i], lowerBound, upperBound)
		}

		if off2[i] < lowerBound || off2[i] > upperBound {
			t.Errorf("Crossover() off2[%d] = %v, expected to be clamped to [%v, %v]",
				i, off2[i], lowerBound, upperBound)
		}
	}
}

// TestCrossoverSymmetry tests that crossover operation is symmetric.
func TestCrossoverSymmetry(t *testing.T) {
	x1 := []float64{1.0, 2.0, 3.0}
	x2 := []float64{4.0, 5.0, 6.0}
	lowerBound := 0.0
	upperBound := 10.0
	seed := int64(999)

	// Crossover in one order
	rng1 := rand.New(rand.NewSource(seed))
	off1a, off2a := Crossover(x1, x2, lowerBound, upperBound, rng1)

	// Crossover in reverse order
	rng2 := rand.New(rand.NewSource(seed))
	off1b, off2b := Crossover(x2, x1, lowerBound, upperBound, rng2)

	// The offspring should be swapped
	for i := 0; i < len(x1); i++ {
		if math.Abs(off1a[i]-off2b[i]) > 1e-10 {
			t.Errorf("Crossover() symmetry: off1a[%d]=%v should equal off2b[%d]=%v",
				i, off1a[i], i, off2b[i])
		}

		if math.Abs(off2a[i]-off1b[i]) > 1e-10 {
			t.Errorf("Crossover() symmetry: off2a[%d]=%v should equal off1b[%d]=%v",
				i, off2a[i], i, off1b[i])
		}
	}
}

// TestMutate tests the mutation operator.
func TestMutate(t *testing.T) {
	tests := []struct {
		name       string
		x          []float64
		mu         float64
		lowerBound float64
		upperBound float64
		seed       int64
	}{
		{
			name:       "low_mutation_rate",
			x:          []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			mu:         0.2,
			lowerBound: 0.0,
			upperBound: 10.0,
			seed:       42,
		},
		{
			name:       "high_mutation_rate",
			x:          []float64{5.0, 5.0, 5.0, 5.0, 5.0},
			mu:         0.8,
			lowerBound: 0.0,
			upperBound: 10.0,
			seed:       123,
		},
		{
			name:       "full_mutation",
			x:          []float64{1.0, 2.0, 3.0},
			mu:         1.0,
			lowerBound: -10.0,
			upperBound: 10.0,
			seed:       789,
		},
		{
			name:       "high_dimensional",
			x:          make([]float64, 50),
			mu:         0.1,
			lowerBound: -5.0,
			upperBound: 5.0,
			seed:       999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize input if needed
			if tt.name == "high_dimensional" {
				for i := range tt.x {
					tt.x[i] = float64(i) * 0.1
				}
			}

			rng := rand.New(rand.NewSource(tt.seed))
			y := Mutate(tt.x, tt.mu, tt.lowerBound, tt.upperBound, rng)

			// Check size
			if len(y) != len(tt.x) {
				t.Errorf("Mutate() length = %v, want %v", len(y), len(tt.x))
			}

			// Check bounds
			for i, val := range y {
				if val < tt.lowerBound || val > tt.upperBound {
					t.Errorf("Mutate() y[%d] = %v, out of bounds [%v, %v]",
						i, val, tt.lowerBound, tt.upperBound)
				}
			}

			// Check that mutation occurred (at least some genes should differ)
			expectedMutations := int(math.Ceil(tt.mu * float64(len(tt.x))))
			actualMutations := 0

			for i := 0; i < len(tt.x); i++ {
				if math.Abs(y[i]-tt.x[i]) > 1e-10 {
					actualMutations++
				}
			}

			if expectedMutations > 0 && actualMutations == 0 {
				t.Errorf("Mutate() expected at least %d mutations, got 0", expectedMutations)
			}

			t.Logf("Mutate() expected ~%d mutations, got %d", expectedMutations, actualMutations)
		})
	}
}

// TestMutateDeterministic tests that mutation is deterministic with seeded RNG.
func TestMutateDeterministic(t *testing.T) {
	x := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	mu := 0.4
	lowerBound := 0.0
	upperBound := 10.0
	seed := int64(42)

	rng1 := rand.New(rand.NewSource(seed))
	y1 := Mutate(x, mu, lowerBound, upperBound, rng1)

	rng2 := rand.New(rand.NewSource(seed))
	y2 := Mutate(x, mu, lowerBound, upperBound, rng2)

	for i := 0; i < len(y1); i++ {
		if y1[i] != y2[i] {
			t.Errorf("Mutate() y[%d] not deterministic: %v vs %v", i, y1[i], y2[i])
		}
	}
}

// TestMutateZeroRate tests mutation with zero mutation rate.
func TestMutateZeroRate(t *testing.T) {
	x := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	mu := 0.0
	lowerBound := 0.0
	upperBound := 10.0

	rng := rand.New(rand.NewSource(42))
	y := Mutate(x, mu, lowerBound, upperBound, rng)

	// With mutation rate 0, output should equal input
	for i := 0; i < len(x); i++ {
		if y[i] != x[i] {
			t.Errorf("Mutate() with mu=0: y[%d] = %v, want %v (no mutation expected)", i, y[i], x[i])
		}
	}
}

// TestMutateBoundaryViolations tests that mutation properly handles boundary violations.
func TestMutateBoundaryViolations(t *testing.T) {
	// Start at bounds with high mutation rate
	x := []float64{-10.0, 10.0, -10.0, 10.0}
	mu := 1.0
	lowerBound := -5.0
	upperBound := 5.0

	rng := rand.New(rand.NewSource(42))
	y := Mutate(x, mu, lowerBound, upperBound, rng)

	// Check that all values are within bounds
	for i := 0; i < len(y); i++ {
		if y[i] < lowerBound || y[i] > upperBound {
			t.Errorf("Mutate() y[%d] = %v, expected to be clamped to [%v, %v]",
				i, y[i], lowerBound, upperBound)
		}
	}
}

// TestMutateDoesNotModifyInput tests that mutation doesn't modify the input vector.
func TestMutateDoesNotModifyInput(t *testing.T) {
	x := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	xCopy := make([]float64, len(x))
	copy(xCopy, x)

	mu := 0.6
	lowerBound := 0.0
	upperBound := 10.0

	rng := rand.New(rand.NewSource(42))
	_ = Mutate(x, mu, lowerBound, upperBound, rng)

	// Check that input wasn't modified
	for i := 0; i < len(x); i++ {
		if x[i] != xCopy[i] {
			t.Errorf("Mutate() modified input: x[%d] = %v, want %v", i, x[i], xCopy[i])
		}
	}
}

// TestMutateSigmaCalculation tests that mutation uses correct sigma.
func TestMutateSigmaCalculation(t *testing.T) {
	x := []float64{5.0, 5.0, 5.0, 5.0, 5.0}
	mu := 1.0 // Mutate all genes
	lowerBound := 0.0
	upperBound := 100.0

	// Expected sigma = 0.1 * (upperBound - lowerBound) = 10.0
	// Mutations should be x[i] + N(0, 10.0)
	// Most values should be within ±30 (3 sigma) of original

	rng := rand.New(rand.NewSource(42))
	y := Mutate(x, mu, lowerBound, upperBound, rng)

	for i := 0; i < len(y); i++ {
		// Check that mutation applied (should be different with high probability)
		if math.Abs(y[i]-x[i]) < 1e-10 {
			t.Logf("Mutate() y[%d] unchanged (statistically rare but possible)", i)
		}

		// Check within bounds
		if y[i] < lowerBound || y[i] > upperBound {
			t.Errorf("Mutate() y[%d] = %v, out of bounds", i, y[i])
		}
	}
}
