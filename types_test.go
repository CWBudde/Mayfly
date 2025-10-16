package mayfly

import (
	"math"
	"math/rand"
	"testing"
)

// TestNewMayfly tests the mayfly constructor.
func TestNewMayfly(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{"small", 5},
		{"medium", 30},
		{"large", 100},
		{"single", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newMayfly(tt.size)

			// Check Position slice
			if m.Position == nil {
				t.Error("newMayfly() Position is nil")
			}

			if len(m.Position) != tt.size {
				t.Errorf("newMayfly() Position length = %v, want %v", len(m.Position), tt.size)
			}

			// Check Velocity slice
			if m.Velocity == nil {
				t.Error("newMayfly() Velocity is nil")
			}

			if len(m.Velocity) != tt.size {
				t.Errorf("newMayfly() Velocity length = %v, want %v", len(m.Velocity), tt.size)
			}

			// Check Cost is initialized to infinity
			if !math.IsInf(m.Cost, 1) {
				t.Errorf("newMayfly() Cost = %v, want +Inf", m.Cost)
			}

			// Check Best.Position slice
			if m.Best.Position == nil {
				t.Error("newMayfly() Best.Position is nil")
			}

			if len(m.Best.Position) != tt.size {
				t.Errorf("newMayfly() Best.Position length = %v, want %v", len(m.Best.Position), tt.size)
			}

			// Check Best.Cost is initialized to infinity
			if !math.IsInf(m.Best.Cost, 1) {
				t.Errorf("newMayfly() Best.Cost = %v, want +Inf", m.Best.Cost)
			}

			// Check that slices are zero-initialized
			for i := 0; i < tt.size; i++ {
				if m.Position[i] != 0.0 {
					t.Errorf("newMayfly() Position[%d] = %v, want 0.0", i, m.Position[i])
				}

				if m.Velocity[i] != 0.0 {
					t.Errorf("newMayfly() Velocity[%d] = %v, want 0.0", i, m.Velocity[i])
				}

				if m.Best.Position[i] != 0.0 {
					t.Errorf("newMayfly() Best.Position[%d] = %v, want 0.0", i, m.Best.Position[i])
				}
			}
		})
	}
}

// TestMayflyClone tests the clone method.
func TestMayflyClone(t *testing.T) {
	original := &Mayfly{
		Position: []float64{1.0, 2.0, 3.0},
		Velocity: []float64{0.1, 0.2, 0.3},
		Cost:     5.5,
		Best: Best{
			Position: []float64{0.5, 1.0, 1.5},
			Cost:     3.2,
		},
	}

	clone := original.clone()

	// Check that values are copied
	if clone.Cost != original.Cost {
		t.Errorf("clone() Cost = %v, want %v", clone.Cost, original.Cost)
	}

	if clone.Best.Cost != original.Best.Cost {
		t.Errorf("clone() Best.Cost = %v, want %v", clone.Best.Cost, original.Best.Cost)
	}

	// Check Position values
	for i := 0; i < len(original.Position); i++ {
		if clone.Position[i] != original.Position[i] {
			t.Errorf("clone() Position[%d] = %v, want %v", i, clone.Position[i], original.Position[i])
		}
	}

	// Check Velocity values
	for i := 0; i < len(original.Velocity); i++ {
		if clone.Velocity[i] != original.Velocity[i] {
			t.Errorf("clone() Velocity[%d] = %v, want %v", i, clone.Velocity[i], original.Velocity[i])
		}
	}

	// Check Best.Position values
	for i := 0; i < len(original.Best.Position); i++ {
		if clone.Best.Position[i] != original.Best.Position[i] {
			t.Errorf("clone() Best.Position[%d] = %v, want %v", i, clone.Best.Position[i], original.Best.Position[i])
		}
	}

	// Check deep copy: modify clone and verify original unchanged
	clone.Position[0] = 999.0
	clone.Velocity[1] = 888.0
	clone.Cost = 777.0
	clone.Best.Position[2] = 666.0
	clone.Best.Cost = 555.0

	if original.Position[0] == 999.0 {
		t.Error("clone() Position is not a deep copy")
	}

	if original.Velocity[1] == 888.0 {
		t.Error("clone() Velocity is not a deep copy")
	}

	if original.Cost == 777.0 {
		t.Error("clone() Cost should be independent")
	}

	if original.Best.Position[2] == 666.0 {
		t.Error("clone() Best.Position is not a deep copy")
	}

	if original.Best.Cost == 555.0 {
		t.Error("clone() Best.Cost should be independent")
	}
}

// TestNewDefaultConfig tests the default configuration factory.
func TestNewDefaultConfig(t *testing.T) {
	config := NewDefaultConfig()

	// Check non-zero default values
	if config.MaxIterations != 2000 {
		t.Errorf("NewDefaultConfig() MaxIterations = %v, want 2000", config.MaxIterations)
	}

	if config.NPop != 20 {
		t.Errorf("NewDefaultConfig() NPop = %v, want 20", config.NPop)
	}

	if config.NPopF != 20 {
		t.Errorf("NewDefaultConfig() NPopF = %v, want 20", config.NPopF)
	}

	if config.G != 0.8 {
		t.Errorf("NewDefaultConfig() G = %v, want 0.8", config.G)
	}

	if config.A1 != 1.0 {
		t.Errorf("NewDefaultConfig() A1 = %v, want 1.0", config.A1)
	}

	if config.A2 != 1.5 {
		t.Errorf("NewDefaultConfig() A2 = %v, want 1.5", config.A2)
	}

	if config.A3 != 1.5 {
		t.Errorf("NewDefaultConfig() A3 = %v, want 1.5", config.A3)
	}

	if config.Beta != 2.0 {
		t.Errorf("NewDefaultConfig() Beta = %v, want 2.0", config.Beta)
	}

	if config.Dance != 5.0 {
		t.Errorf("NewDefaultConfig() Dance = %v, want 5.0", config.Dance)
	}

	if config.FL != 1.0 {
		t.Errorf("NewDefaultConfig() FL = %v, want 1.0", config.FL)
	}

	if config.DanceDamp != 0.8 {
		t.Errorf("NewDefaultConfig() DanceDamp = %v, want 0.8", config.DanceDamp)
	}

	if config.FLDamp != 0.99 {
		t.Errorf("NewDefaultConfig() FLDamp = %v, want 0.99", config.FLDamp)
	}

	if config.NC != 20 {
		t.Errorf("NewDefaultConfig() NC = %v, want 20", config.NC)
	}

	if config.Mu != 0.01 {
		t.Errorf("NewDefaultConfig() Mu = %v, want 0.01", config.Mu)
	}

	// Check DESMA defaults
	if config.UseDESMA != false {
		t.Errorf("NewDefaultConfig() UseDESMA = %v, want false", config.UseDESMA)
	}

	if config.EliteCount != 5 {
		t.Errorf("NewDefaultConfig() EliteCount = %v, want 5", config.EliteCount)
	}

	if config.EnlargeFactor != 1.05 {
		t.Errorf("NewDefaultConfig() EnlargeFactor = %v, want 1.05", config.EnlargeFactor)
	}

	if config.ReductionFactor != 0.95 {
		t.Errorf("NewDefaultConfig() ReductionFactor = %v, want 0.95", config.ReductionFactor)
	}

	// Check zero-initialized values
	if config.NM != 0 {
		t.Errorf("NewDefaultConfig() NM = %v, want 0 (auto-calculated)", config.NM)
	}

	if config.SearchRange != 0 {
		t.Errorf("NewDefaultConfig() SearchRange = %v, want 0 (auto-calculated)", config.SearchRange)
	}

	// Check nil values
	if config.ObjectiveFunc != nil {
		t.Error("NewDefaultConfig() ObjectiveFunc should be nil (user must set)")
	}

	if config.Rand != nil {
		t.Error("NewDefaultConfig() Rand should be nil (optional)")
	}
}

// TestNewDESMAConfig tests the DESMA configuration factory.
func TestNewDESMAConfig(t *testing.T) {
	config := NewDESMAConfig()

	// Check that DESMA is enabled
	if !config.UseDESMA {
		t.Error("NewDESMAConfig() UseDESMA = false, want true")
	}

	// Check that it inherits default values
	if config.MaxIterations != 2000 {
		t.Errorf("NewDESMAConfig() MaxIterations = %v, want 2000", config.MaxIterations)
	}

	if config.NPop != 20 {
		t.Errorf("NewDESMAConfig() NPop = %v, want 20", config.NPop)
	}

	// Check DESMA-specific defaults
	if config.EliteCount != 5 {
		t.Errorf("NewDESMAConfig() EliteCount = %v, want 5", config.EliteCount)
	}

	if config.EnlargeFactor != 1.05 {
		t.Errorf("NewDESMAConfig() EnlargeFactor = %v, want 1.05", config.EnlargeFactor)
	}

	if config.ReductionFactor != 0.95 {
		t.Errorf("NewDESMAConfig() ReductionFactor = %v, want 0.95", config.ReductionFactor)
	}
}

// TestConfigModification tests that config can be modified after creation.
func TestConfigModification(t *testing.T) {
	config := NewDefaultConfig()

	// Modify values
	config.MaxIterations = 1000
	config.NPop = 50
	config.G = 0.9
	config.ObjectiveFunc = Sphere
	config.ProblemSize = 10
	config.LowerBound = -5.0
	config.UpperBound = 5.0
	config.Rand = rand.New(rand.NewSource(42))

	// Verify modifications
	if config.MaxIterations != 1000 {
		t.Errorf("Modified MaxIterations = %v, want 1000", config.MaxIterations)
	}

	if config.NPop != 50 {
		t.Errorf("Modified NPop = %v, want 50", config.NPop)
	}

	if config.G != 0.9 {
		t.Errorf("Modified G = %v, want 0.9", config.G)
	}

	if config.ObjectiveFunc == nil {
		t.Error("Modified ObjectiveFunc is nil")
	}

	if config.ProblemSize != 10 {
		t.Errorf("Modified ProblemSize = %v, want 10", config.ProblemSize)
	}

	if config.Rand == nil {
		t.Error("Modified Rand is nil")
	}
}

// TestConfigIndependence tests that multiple configs are independent.
func TestConfigIndependence(t *testing.T) {
	config1 := NewDefaultConfig()
	config2 := NewDefaultConfig()

	// Modify config1
	config1.MaxIterations = 500
	config1.NPop = 100

	// Check that config2 is unchanged
	if config2.MaxIterations == 500 {
		t.Error("Config instances are not independent")
	}

	if config2.NPop == 100 {
		t.Error("Config instances are not independent")
	}

	// Verify config2 has default values
	if config2.MaxIterations != 2000 {
		t.Errorf("config2 MaxIterations = %v, want 2000", config2.MaxIterations)
	}

	if config2.NPop != 20 {
		t.Errorf("config2 NPop = %v, want 20", config2.NPop)
	}
}

// TestBestStruct tests the Best struct.
func TestBestStruct(t *testing.T) {
	best := Best{
		Position: []float64{1.0, 2.0, 3.0},
		Cost:     5.5,
	}

	if len(best.Position) != 3 {
		t.Errorf("Best.Position length = %v, want 3", len(best.Position))
	}

	if best.Cost != 5.5 {
		t.Errorf("Best.Cost = %v, want 5.5", best.Cost)
	}

	// Test modification
	best.Position[0] = 10.0
	best.Cost = 1.0

	if best.Position[0] != 10.0 {
		t.Errorf("Modified Best.Position[0] = %v, want 10.0", best.Position[0])
	}

	if best.Cost != 1.0 {
		t.Errorf("Modified Best.Cost = %v, want 1.0", best.Cost)
	}
}

// TestResultStruct tests the Result struct.
func TestResultStruct(t *testing.T) {
	result := Result{
		GlobalBest: Best{
			Position: []float64{1.0, 2.0},
			Cost:     3.5,
		},
		BestSolution:   []float64{10.0, 5.0, 2.0, 1.0},
		FuncEvalCount:  1000,
		IterationCount: 100,
	}

	// Check GlobalBest
	if len(result.GlobalBest.Position) != 2 {
		t.Errorf("Result.GlobalBest.Position length = %v, want 2", len(result.GlobalBest.Position))
	}

	if result.GlobalBest.Cost != 3.5 {
		t.Errorf("Result.GlobalBest.Cost = %v, want 3.5", result.GlobalBest.Cost)
	}

	// Check BestSolution convergence history
	if len(result.BestSolution) != 4 {
		t.Errorf("Result.BestSolution length = %v, want 4", len(result.BestSolution))
	}

	// Check counts
	if result.FuncEvalCount != 1000 {
		t.Errorf("Result.FuncEvalCount = %v, want 1000", result.FuncEvalCount)
	}

	if result.IterationCount != 100 {
		t.Errorf("Result.IterationCount = %v, want 100", result.IterationCount)
	}

	// Check that convergence values are decreasing (typical for minimization)
	for i := 1; i < len(result.BestSolution); i++ {
		if result.BestSolution[i] > result.BestSolution[i-1] {
			t.Logf("Result.BestSolution[%d]=%v > BestSolution[%d]=%v (not monotonically decreasing)",
				i, result.BestSolution[i], i-1, result.BestSolution[i-1])
		}
	}
}

// TestObjectiveFunctionType tests the ObjectiveFunction type.
func TestObjectiveFunctionType(t *testing.T) {
	// Define a simple objective function
	var objFunc ObjectiveFunction = func(x []float64) float64 {
		sum := 0.0
		for _, val := range x {
			sum += val * val
		}
		return sum
	}

	// Test the function
	x := []float64{1.0, 2.0, 3.0}
	result := objFunc(x)
	expected := 14.0 // 1^2 + 2^2 + 3^2

	if result != expected {
		t.Errorf("ObjectiveFunction result = %v, want %v", result, expected)
	}

	// Test with built-in function
	objFunc = Sphere
	x = []float64{0.0, 0.0, 0.0}

	result = objFunc(x)
	if result != 0.0 {
		t.Errorf("Sphere at origin = %v, want 0.0", result)
	}
}
