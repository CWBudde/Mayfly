package mayfly

import (
	"math"
	"math/rand"
	"testing"
)

// TestLevyFlight tests the Lévy flight distribution generator.
func TestLevyFlight(t *testing.T) {
	tests := []struct {
		name  string
		alpha float64
		beta  float64
		rng   *rand.Rand
	}{
		{"standard_levy", 1.5, 1.0, rand.New(rand.NewSource(42))},
		{"alpha_1.9", 1.9, 1.0, rand.New(rand.NewSource(123))},
		{"alpha_1.0", 1.0, 1.0, rand.New(rand.NewSource(456))},
		{"beta_0.5", 1.5, 0.5, rand.New(rand.NewSource(789))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate multiple samples
			samples := 1000

			values := make([]float64, 0, samples)

			for range samples {
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

func TestLevyFlightRejectsDegenerateParameters(t *testing.T) {
	for _, testCase := range []struct {
		name  string
		alpha float64
		beta  float64
	}{
		{name: "alpha zero", alpha: 0, beta: 1},
		{name: "alpha two", alpha: 2, beta: 1},
		{name: "alpha NaN", alpha: math.NaN(), beta: 1},
		{name: "beta zero", alpha: 1.5, beta: 0},
		{name: "beta infinity", alpha: 1.5, beta: math.Inf(1)},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if got := levyFlight(testCase.alpha, testCase.beta, rand.New(rand.NewSource(1))); got != 0 {
				t.Errorf("invalid levy parameters returned %v, want safe zero step", got)
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

	for range 100 {
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
		name  string
		size  int
		alpha float64
		beta  float64
		rng   *rand.Rand
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

			for i := range result {
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
		name       string
		current    []float64
		best       []float64
		lowerBound float64
		upperBound float64
		rng        *rand.Rand
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

	for i := range result1 {
		if result1[i] != result2[i] {
			t.Errorf("gaussianUpdate()[%d] with same seed produced different values: %v vs %v",
				i, result1[i], result2[i])
		}
	}
}

// TestEliteBounds checks the dynamic elite interval, including the fallback to
// the static search bounds when the elite collapses in a dimension.
func TestEliteBounds(t *testing.T) {
	makeMayfly := func(position ...float64) *Mayfly {
		mayfly := newMayfly(len(position))
		copy(mayfly.Position, position)

		return mayfly
	}

	population := []*Mayfly{
		makeMayfly(1.0, 4.0, -2.0),
		makeMayfly(3.0, 4.0, 0.5),
		makeMayfly(-1.0, 4.0, 7.0),
		makeMayfly(99.0, 99.0, 99.0), // outside the elite set
	}

	da, db := eliteBounds(population, 3, -10, 10)

	wantDa := []float64{-1.0, -10.0, -2.0}
	wantDb := []float64{3.0, 10.0, 7.0}

	for i := range wantDa {
		if da[i] != wantDa[i] || db[i] != wantDb[i] {
			t.Errorf("dimension %d: got [%v, %v], want [%v, %v]",
				i, da[i], db[i], wantDa[i], wantDb[i])
		}
	}

	if da, db := eliteBounds(population, 0, -10, 10); da != nil || db != nil {
		t.Errorf("eliteBounds() with count 0 = (%v, %v), want (nil, nil)", da, db)
	}
}

// TestEliteOppositionPointStaysInBounds checks that elite opposition never
// leaves the static search bounds, whichever branch of the rule fires.
func TestEliteOppositionPointStaysInBounds(t *testing.T) {
	rng := rand.New(rand.NewSource(11))
	da := []float64{-3, 0, 4}
	db := []float64{-1, 2, 9}
	position := []float64{-2, 1, 8}

	for range 500 {
		opposite := eliteOppositionPoint(position, da, db, -10, 10, rng)
		for i, value := range opposite {
			if value < -10 || value > 10 || math.IsNaN(value) {
				t.Fatalf("dimension %d out of bounds: %v", i, value)
			}
		}
	}
}

// TestEliteOppositionPointDiffersFromStaticOpposition documents the defect this
// operator fixes: static opposition mirrors an elite through the middle of the
// whole search space, which for an already-good elite is essentially always
// worse and therefore never accepted. Elite opposition reflects through the
// interval spanned by the elite set instead.
func TestEliteOppositionPointDiffersFromStaticOpposition(t *testing.T) {
	rng := rand.New(rand.NewSource(5))
	da := []float64{0.9, 0.9}
	db := []float64{1.1, 1.1}
	position := []float64{1.0, 1.0}

	static := oppositionPoint(position, -10, 10)

	matches := 0

	for range 100 {
		elite := eliteOppositionPoint(position, da, db, -10, 10, rng)
		if elite[0] == static[0] && elite[1] == static[1] {
			matches++
		}
	}

	if matches > 0 {
		t.Errorf("elite opposition coincided with static opposition %d times", matches)
	}
}

// TestEOBBMAOppositionRateIsNotInert is the inertness guard for EOBBMA: turning
// the opposition knob must change the search. Before the elite opposition fix
// the opposition candidates were evaluated but never accepted, so runs with
// OppositionRate 0 and 1 produced bit-identical results.
func TestEOBBMAOppositionRateIsNotInert(t *testing.T) {
	for _, parallel := range []bool{false, true} {
		run := func(oppositionRate float64, seed int64) float64 {
			config := NewEOBBMAConfig()
			config.ObjectiveFunc = Rastrigin
			config.ProblemSize = 6
			config.LowerBound = -5.12
			config.UpperBound = 5.12
			config.MaxIterations = 50
			config.EnableParallel = parallel
			config.Rand = rand.New(rand.NewSource(seed))
			config.OppositionRate = oppositionRate

			result, err := Optimize(config)
			if err != nil {
				t.Fatalf("Optimize() error = %v", err)
			}

			return result.GlobalBest.Cost
		}

		differed := 0

		for seed := int64(1); seed <= 5; seed++ {
			if run(0, seed) != run(1, seed) {
				differed++
			}
		}

		if differed == 0 {
			t.Errorf("parallel=%v: OppositionRate has no observable effect on EOBBMA results", parallel)
		}
	}
}
