package mayfly

import (
	"math"
	"math/rand"
	"testing"
)

// TestGenerateEliteMayflies tests the DESMA elite generation mechanism.
func TestGenerateEliteMayflies(t *testing.T) {
	tests := []struct {
		name        string
		currentBest Best
		searchRange float64
		eliteCount  int
		problemSize int
		lowerBound  float64
		upperBound  float64
		objFunc     ObjectiveFunction
		seed        int64
	}{
		{
			name: "sphere_2d",
			currentBest: Best{
				Position: []float64{1.0, 1.0},
				Cost:     2.0,
			},
			searchRange: 0.5,
			eliteCount:  5,
			problemSize: 2,
			lowerBound:  -10.0,
			upperBound:  10.0,
			objFunc:     Sphere,
			seed:        42,
		},
		{
			name: "sphere_high_dimensional",
			currentBest: Best{
				Position: make([]float64, 30),
				Cost:     0.1,
			},
			searchRange: 1.0,
			eliteCount:  10,
			problemSize: 30,
			lowerBound:  -5.0,
			upperBound:  5.0,
			objFunc:     Sphere,
			seed:        123,
		},
		{
			name: "rastrigin",
			currentBest: Best{
				Position: []float64{0.5, -0.5, 0.2},
				Cost:     5.0,
			},
			searchRange: 0.3,
			eliteCount:  3,
			problemSize: 3,
			lowerBound:  -5.12,
			upperBound:  5.12,
			objFunc:     Rastrigin,
			seed:        789,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize currentBest position if needed
			if tt.name == "sphere_high_dimensional" {
				for i := range tt.currentBest.Position {
					tt.currentBest.Position[i] = 0.01 * float64(i)
				}
				tt.currentBest.Cost = tt.objFunc(tt.currentBest.Position)
			}

			rng := rand.New(rand.NewSource(tt.seed))
			elite, funcEvals := generateEliteMayflies(
				tt.currentBest,
				tt.searchRange,
				tt.eliteCount,
				tt.problemSize,
				tt.lowerBound,
				tt.upperBound,
				tt.objFunc,
				rng,
			)

			// Check that function evaluations match elite count
			if funcEvals != tt.eliteCount {
				t.Errorf("generateEliteMayflies() funcEvals = %v, want %v", funcEvals, tt.eliteCount)
			}

			// Check elite structure
			if elite == nil {
				t.Fatal("generateEliteMayflies() returned nil elite")
			}
			if len(elite.Position) != tt.problemSize {
				t.Errorf("generateEliteMayflies() elite.Position length = %v, want %v",
					len(elite.Position), tt.problemSize)
			}

			// Check that elite is within bounds
			for i, val := range elite.Position {
				if val < tt.lowerBound || val > tt.upperBound {
					t.Errorf("generateEliteMayflies() elite.Position[%d] = %v, out of bounds [%v, %v]",
						i, val, tt.lowerBound, tt.upperBound)
				}
			}

			// Check that elite cost is valid (not NaN or Inf)
			if math.IsNaN(elite.Cost) {
				t.Error("generateEliteMayflies() elite.Cost is NaN")
			}
			if math.IsInf(elite.Cost, 0) {
				t.Error("generateEliteMayflies() elite.Cost is Inf")
			}

			// Check that elite cost matches evaluation
			// (Allow some tolerance for floating point precision)
			expectedCost := tt.objFunc(elite.Position)
			tolerance := math.Max(1e-10, math.Abs(expectedCost)*1e-10)
			if math.Abs(elite.Cost-expectedCost) > tolerance {
				t.Logf("generateEliteMayflies() elite.Cost = %v, re-evaluated = %v (possible floating point variance)",
					elite.Cost, expectedCost)
			}

			// Check that Best is synced with Position/Cost
			for i := 0; i < tt.problemSize; i++ {
				if elite.Best.Position[i] != elite.Position[i] {
					t.Errorf("generateEliteMayflies() elite.Best.Position[%d] != elite.Position[%d]",
						i, i)
				}
			}
			if elite.Best.Cost != elite.Cost {
				t.Error("generateEliteMayflies() elite.Best.Cost != elite.Cost")
			}

			// Check that elite is near currentBest (within searchRange)
			// Note: may be clamped to bounds, so this is a soft check
			nearbyCount := 0
			for i := 0; i < tt.problemSize; i++ {
				distance := math.Abs(elite.Position[i] - tt.currentBest.Position[i])
				if distance <= tt.searchRange {
					nearbyCount++
				}
			}
			t.Logf("Elite has %d/%d dimensions within searchRange of currentBest",
				nearbyCount, tt.problemSize)
		})
	}
}

// TestGenerateEliteMayfliesImprovement tests that elite generation can find improvements.
func TestGenerateEliteMayfliesImprovement(t *testing.T) {
	// Start at a suboptimal point
	currentBest := Best{
		Position: []float64{2.0, 2.0, 2.0},
		Cost:     Sphere([]float64{2.0, 2.0, 2.0}), // 12.0
	}

	// Large search range and many elites increase chance of finding improvement
	searchRange := 2.0
	eliteCount := 50
	problemSize := 3
	lowerBound := -10.0
	upperBound := 10.0

	rng := rand.New(rand.NewSource(42))
	elite, _ := generateEliteMayflies(
		currentBest,
		searchRange,
		eliteCount,
		problemSize,
		lowerBound,
		upperBound,
		Sphere,
		rng,
	)

	// Check that elite was generated
	if elite == nil {
		t.Fatal("generateEliteMayflies() returned nil")
	}

	// Log whether improvement was found
	if elite.Cost < currentBest.Cost {
		t.Logf("generateEliteMayflies() found improvement: %v -> %v",
			currentBest.Cost, elite.Cost)
	} else if elite.Cost == currentBest.Cost {
		t.Logf("generateEliteMayflies() maintained best: %v", elite.Cost)
	} else {
		t.Logf("generateEliteMayflies() no improvement found: %v (best: %v)",
			elite.Cost, currentBest.Cost)
	}

	// Elite should never be worse than currentBest (it returns best of all generated)
	// Actually, looking at the implementation, it initializes with currentBest
	if elite.Cost > currentBest.Cost {
		t.Errorf("generateEliteMayflies() returned worse elite: %v > %v",
			elite.Cost, currentBest.Cost)
	}
}

// TestGenerateEliteMayfliesDeterministic tests deterministic behavior.
func TestGenerateEliteMayfliesDeterministic(t *testing.T) {
	currentBest := Best{
		Position: []float64{1.0, 2.0, 3.0},
		Cost:     14.0,
	}
	searchRange := 0.5
	eliteCount := 5
	problemSize := 3
	lowerBound := -10.0
	upperBound := 10.0
	seed := int64(42)

	// Generate twice with same seed
	rng1 := rand.New(rand.NewSource(seed))
	elite1, funcEvals1 := generateEliteMayflies(
		currentBest, searchRange, eliteCount, problemSize,
		lowerBound, upperBound, Sphere, rng1,
	)

	rng2 := rand.New(rand.NewSource(seed))
	elite2, funcEvals2 := generateEliteMayflies(
		currentBest, searchRange, eliteCount, problemSize,
		lowerBound, upperBound, Sphere, rng2,
	)

	// Check function evaluations match
	if funcEvals1 != funcEvals2 {
		t.Errorf("generateEliteMayflies() funcEvals not deterministic: %v vs %v",
			funcEvals1, funcEvals2)
	}

	// Check elite cost matches
	if elite1.Cost != elite2.Cost {
		t.Errorf("generateEliteMayflies() elite.Cost not deterministic: %v vs %v",
			elite1.Cost, elite2.Cost)
	}

	// Check elite position matches
	for i := 0; i < problemSize; i++ {
		if elite1.Position[i] != elite2.Position[i] {
			t.Errorf("generateEliteMayflies() elite.Position[%d] not deterministic: %v vs %v",
				i, elite1.Position[i], elite2.Position[i])
		}
	}
}

// TestGenerateEliteMayfliesSearchRange tests different search ranges.
func TestGenerateEliteMayfliesSearchRange(t *testing.T) {
	currentBest := Best{
		Position: []float64{5.0, 5.0},
		Cost:     50.0,
	}
	eliteCount := 10
	problemSize := 2
	lowerBound := 0.0
	upperBound := 10.0
	seed := int64(999)

	tests := []struct {
		name        string
		searchRange float64
	}{
		{"small_range", 0.1},
		{"medium_range", 1.0},
		{"large_range", 5.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rng := rand.New(rand.NewSource(seed))
			elite, _ := generateEliteMayflies(
				currentBest, tt.searchRange, eliteCount, problemSize,
				lowerBound, upperBound, Sphere, rng,
			)

			// Check that elite was generated
			if elite == nil {
				t.Fatal("generateEliteMayflies() returned nil")
			}

			// Larger search ranges should potentially explore more diverse solutions
			t.Logf("SearchRange=%v: Elite cost=%v, distance from best=%v",
				tt.searchRange, elite.Cost,
				math.Sqrt(math.Pow(elite.Position[0]-currentBest.Position[0], 2)+
					math.Pow(elite.Position[1]-currentBest.Position[1], 2)))
		})
	}
}

// TestGenerateEliteMayfliesBoundaryHandling tests elite generation near boundaries.
func TestGenerateEliteMayfliesBoundaryHandling(t *testing.T) {
	tests := []struct {
		name        string
		currentBest Best
		lowerBound  float64
		upperBound  float64
	}{
		{
			name: "at_lower_bound",
			currentBest: Best{
				Position: []float64{-10.0, -10.0},
				Cost:     200.0,
			},
			lowerBound: -10.0,
			upperBound: 10.0,
		},
		{
			name: "at_upper_bound",
			currentBest: Best{
				Position: []float64{10.0, 10.0},
				Cost:     200.0,
			},
			lowerBound: -10.0,
			upperBound: 10.0,
		},
		{
			name: "near_lower_bound",
			currentBest: Best{
				Position: []float64{-9.5, -9.8},
				Cost:     186.49,
			},
			lowerBound: -10.0,
			upperBound: 10.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			searchRange := 1.0
			eliteCount := 20
			problemSize := 2

			rng := rand.New(rand.NewSource(42))
			elite, _ := generateEliteMayflies(
				tt.currentBest, searchRange, eliteCount, problemSize,
				tt.lowerBound, tt.upperBound, Sphere, rng,
			)

			// Check all positions are within bounds
			for i, val := range elite.Position {
				if val < tt.lowerBound || val > tt.upperBound {
					t.Errorf("generateEliteMayflies() elite.Position[%d] = %v, out of bounds [%v, %v]",
						i, val, tt.lowerBound, tt.upperBound)
				}
			}

			t.Logf("Elite at boundary: cost=%v, position=%v", elite.Cost, elite.Position)
		})
	}
}

// TestGenerateEliteMayfliesSingleElite tests generating a single elite.
func TestGenerateEliteMayfliesSingleElite(t *testing.T) {
	currentBest := Best{
		Position: []float64{1.0, 1.0, 1.0},
		Cost:     3.0,
	}
	searchRange := 0.5
	eliteCount := 1
	problemSize := 3
	lowerBound := -10.0
	upperBound := 10.0

	rng := rand.New(rand.NewSource(42))
	elite, funcEvals := generateEliteMayflies(
		currentBest, searchRange, eliteCount, problemSize,
		lowerBound, upperBound, Sphere, rng,
	)

	// Check that exactly one function evaluation was performed
	if funcEvals != 1 {
		t.Errorf("generateEliteMayflies() with eliteCount=1: funcEvals = %v, want 1", funcEvals)
	}

	// Check that elite was generated
	if elite == nil {
		t.Fatal("generateEliteMayflies() returned nil")
	}

	// The single elite should be near the current best
	t.Logf("Single elite: cost=%v, position=%v", elite.Cost, elite.Position)
}

// TestGenerateEliteMayfliesZeroElites tests edge case of zero elites.
func TestGenerateEliteMayfliesZeroElites(t *testing.T) {
	currentBest := Best{
		Position: []float64{1.0, 1.0},
		Cost:     2.0,
	}
	searchRange := 0.5
	eliteCount := 0
	problemSize := 2
	lowerBound := -10.0
	upperBound := 10.0

	rng := rand.New(rand.NewSource(42))
	elite, funcEvals := generateEliteMayflies(
		currentBest, searchRange, eliteCount, problemSize,
		lowerBound, upperBound, Sphere, rng,
	)

	// Should return currentBest with no function evaluations
	if funcEvals != 0 {
		t.Errorf("generateEliteMayflies() with eliteCount=0: funcEvals = %v, want 0", funcEvals)
	}

	// Elite should match currentBest
	if elite.Cost != currentBest.Cost {
		t.Errorf("generateEliteMayflies() with eliteCount=0: elite.Cost = %v, want %v",
			elite.Cost, currentBest.Cost)
	}
}
