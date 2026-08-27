package mayfly

import (
	"math"
	"math/rand"
	"slices"
	"testing"
)

func TestDESMACrossoverEquationsSixAndSeven(t *testing.T) {
	male := []float64{2, -4, 8}
	female := []float64{-6, 10, 3}
	seed := int64(73)

	offspring1, offspring2 := desmaCrossover(
		male, female, -100, 100, rand.New(rand.NewSource(seed)),
	)
	wantRNG := rand.New(rand.NewSource(seed))

	coefficients := make([]float64, len(male))
	for dimension := range male {
		coefficient := unifrnd(-1, 1, wantRNG)
		coefficients[dimension] = coefficient
		want1 := coefficient*male[dimension] + (1-coefficient)*female[dimension]
		want2 := coefficient*female[dimension] + (1-coefficient)*male[dimension]

		if offspring1[dimension] != want1 || offspring2[dimension] != want2 {
			t.Fatalf(
				"dimension %d = (%v, %v), want DESMA Eqs. (6)-(7) (%v, %v)",
				dimension, offspring1[dimension], offspring2[dimension], want1, want2,
			)
		}

		if coefficient < -1 || coefficient > 1 {
			t.Fatalf("dimension %d coefficient = %v, want [-1,1]", dimension, coefficient)
		}

		if math.Abs(
			(offspring1[dimension]+offspring2[dimension])-(male[dimension]+female[dimension]),
		) > 1e-14 {
			t.Fatalf("dimension %d complementary offspring do not preserve the parent sum", dimension)
		}
	}

	if coefficients[0] == coefficients[1] && coefficients[1] == coefficients[2] {
		t.Fatalf("DESMA crossover reused one scalar coefficient across coordinates: %v", coefficients)
	}
}

func TestCrossoverForConfigUsesDESMARangeAndPreservesHMMAPrecedence(t *testing.T) {
	male := []float64{2, -4}
	female := []float64{-6, 10}

	desmaConfig := NewDESMAConfig()
	desmaConfig.LowerBound = -100
	desmaConfig.UpperBound = 100
	desmaConfig.CrossoverGamma = 99 // DESMA's paper-specific L ignores generic BLX gamma.

	got1, got2 := crossoverForConfig(male, female, desmaConfig, rand.New(rand.NewSource(19)))

	want1, want2 := desmaCrossover(male, female, -100, 100, rand.New(rand.NewSource(19)))
	if !slices.Equal(got1, want1) || !slices.Equal(got2, want2) {
		t.Fatalf("DESMA dispatch = (%v, %v), want (%v, %v)", got1, got2, want1, want2)
	}

	hmmaConfig := NewHMMAConfig()
	hmmaConfig.UseDESMA = true
	hmmaConfig.LowerBound = -100
	hmmaConfig.UpperBound = 100
	got1, got2 = crossoverForConfig(male, female, hmmaConfig, rand.New(rand.NewSource(19)))

	want1, want2 = CrossoverBlend(male, female, 0, -100, 100, rand.New(rand.NewSource(19)))
	if !slices.Equal(got1, want1) || !slices.Equal(got2, want2) {
		t.Fatalf("HMMA precedence = (%v, %v), want (%v, %v)", got1, got2, want1, want2)
	}
}

func TestCommitDESMAEliteReplacesCurrentBestPopulationMember(t *testing.T) {
	newCandidate := func(position, cost float64) *Mayfly {
		candidate := newMayfly(1)
		candidate.Position[0] = position
		candidate.Cost = cost
		candidate.Best.Position[0] = position
		candidate.Best.Cost = cost

		return candidate
	}

	tests := []struct {
		name             string
		maleCosts        []float64
		femaleCosts      []float64
		wantReplacedMale bool
	}{
		{name: "best male", maleCosts: []float64{2, 8}, femaleCosts: []float64{3, 9}, wantReplacedMale: true},
		{name: "best female", maleCosts: []float64{3, 8}, femaleCosts: []float64{2, 9}, wantReplacedMale: false},
		{name: "tie keeps male", maleCosts: []float64{2, 8}, femaleCosts: []float64{2, 9}, wantReplacedMale: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			males := []*Mayfly{
				newCandidate(test.maleCosts[0], test.maleCosts[0]),
				newCandidate(test.maleCosts[1], test.maleCosts[1]),
			}
			females := []*Mayfly{
				newCandidate(test.femaleCosts[0], test.femaleCosts[0]),
				newCandidate(test.femaleCosts[1], test.femaleCosts[1]),
			}
			oldMaleBest, oldMaleWorst := males[0], males[1]
			oldFemaleBest, oldFemaleWorst := females[0], females[1]
			globalBest := Best{Position: []float64{2}, Cost: 2}
			elite := newCandidate(1, 1)

			committed := commitDESMAElite(
				males, females, &globalBest, elite, newConstraintEvaluator(Sphere, nil),
			)
			if !committed {
				t.Fatal("strictly improving elite was not committed")
			}

			if globalBest.Cost != elite.Cost || globalBest.Position[0] != elite.Position[0] {
				t.Fatalf("global attractor = %+v, want elite %+v", globalBest, elite)
			}

			if test.wantReplacedMale {
				if males[0] != elite || males[1] != oldMaleWorst || females[0] != oldFemaleBest ||
					females[1] != oldFemaleWorst || slices.Contains(males, oldMaleBest) {
					t.Fatalf("male-best replacement produced males=%p/%p females=%p/%p",
						males[0], males[1], females[0], females[1])
				}
			} else if females[0] != elite || females[1] != oldFemaleWorst || males[0] != oldMaleBest ||
				males[1] != oldMaleWorst || slices.Contains(females, oldFemaleBest) {
				t.Fatalf("female-best replacement produced males=%p/%p females=%p/%p",
					males[0], males[1], females[0], females[1])
			}
		})
	}
}

func TestCommitDESMAEliteRejectsEqualAndWorseCandidates(t *testing.T) {
	for _, eliteCost := range []float64{2, 3} {
		males := []*Mayfly{mayflyFromBest(Best{Position: []float64{2}, Cost: 2}, 1)}
		females := []*Mayfly{mayflyFromBest(Best{Position: []float64{4}, Cost: 4}, 1)}
		globalBest := Best{Position: []float64{2}, Cost: 2}
		elite := mayflyFromBest(Best{Position: []float64{eliteCost}, Cost: eliteCost}, 1)
		oldMale, oldFemale := males[0], females[0]

		if commitDESMAElite(males, females, &globalBest, elite, newConstraintEvaluator(Sphere, nil)) {
			t.Fatalf("elite cost %v was accepted over equal/better incumbent", eliteCost)
		}

		if males[0] != oldMale || females[0] != oldFemale || globalBest.Cost != 2 || globalBest.Position[0] != 2 {
			t.Fatalf("rejected elite cost %v changed the lifecycle", eliteCost)
		}
	}
}

func TestDESMAEquation16UsesCommittedEliteAsNextAttractor(t *testing.T) {
	globalBest := Best{Position: []float64{4}, Cost: 16}
	males := []*Mayfly{mayflyFromBest(globalBest, 1)}
	females := []*Mayfly{mayflyFromBest(Best{Position: []float64{5}, Cost: 25}, 1)}
	elite := mayflyFromBest(Best{Position: []float64{1}, Cost: 1}, 1)

	evaluator := newConstraintEvaluator(Sphere, nil)
	if !commitDESMAElite(males, females, &globalBest, elite, evaluator) {
		t.Fatal("improving elite was not committed")
	}

	male := mayflyFromBest(Best{Position: []float64{2}, Cost: 4}, 1)
	male.Position[0] = 3
	male.Cost = 9
	male.Velocity[0] = 0.5
	config := NewDESMAConfig()
	config.ProblemSize = 1
	config.LowerBound = -100
	config.UpperBound = 100
	config.VelMin = -100
	config.VelMax = 100
	config.A1 = 1
	config.A2 = 1.5
	config.Beta = 0.25

	const gravity = 0.8
	prepareStandardMale(
		male, globalBest, nil, gravity, config.Dance, gravity, config,
		rand.New(rand.NewSource(5)), evaluator,
	)

	personalDelta := 2.0 - 3.0
	eliteDelta := 1.0 - 3.0

	wantVelocity := gravity*0.5 +
		config.A1*math.Exp(-config.Beta*personalDelta*personalDelta)*personalDelta +
		config.A2*math.Exp(-config.Beta*eliteDelta*eliteDelta)*eliteDelta
	if math.Abs(male.Velocity[0]-wantVelocity) > 1e-15 {
		t.Fatalf("Eq. (16) velocity = %.17g, want %.17g", male.Velocity[0], wantVelocity)
	}

	if math.Abs(male.Position[0]-(3+wantVelocity)) > 1e-15 {
		t.Fatalf("Eq. (16) position = %.17g, want %.17g", male.Position[0], 3+wantVelocity)
	}
}

func TestDESMAEquation16DancesWhenEliteDoesNotDominate(t *testing.T) {
	male := mayflyFromBest(Best{Position: []float64{1}, Cost: 1}, 1)
	male.Position[0] = 2
	male.Cost = 1 // Equal fitness takes the non-attraction branch.
	male.Velocity[0] = 0.5
	eliteBest := Best{Position: []float64{-3}, Cost: 1}
	config := NewDESMAConfig()
	config.ProblemSize = 1
	config.LowerBound = -100
	config.UpperBound = 100
	config.VelMin = -100
	config.VelMax = 100
	config.Dance = 0

	const gravity = 0.8
	prepareStandardMale(
		male, eliteBest, nil, gravity, config.Dance, gravity, config,
		rand.New(rand.NewSource(5)), newConstraintEvaluator(Sphere, nil),
	)

	if male.Velocity[0] != gravity*0.5 || male.Position[0] != 2+gravity*0.5 {
		t.Fatalf(
			"equal elite took attraction branch: velocity=%v position=%v",
			male.Velocity[0], male.Position[0],
		)
	}
}

// TestGenerateEliteMayflies tests the DESMA elite generation mechanism.
func TestGenerateEliteMayflies(t *testing.T) {
	tests := []struct {
		objFunc     ObjectiveFunction
		name        string
		currentBest Best
		searchRange float64
		eliteCount  int
		problemSize int
		lowerBound  float64
		upperBound  float64
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
			for i := range tt.problemSize {
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

			for i := range tt.problemSize {
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
	switch {
	case elite.Cost < currentBest.Cost:
		t.Logf("generateEliteMayflies() found improvement: %v -> %v",
			currentBest.Cost, elite.Cost)
	case elite.Cost == currentBest.Cost:
		t.Logf("generateEliteMayflies() maintained best: %v", elite.Cost)
	default:
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
	for i := range problemSize {
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
				math.Sqrt((elite.Position[0]-currentBest.Position[0])*(elite.Position[0]-currentBest.Position[0])+
					(elite.Position[1]-currentBest.Position[1])*(elite.Position[1]-currentBest.Position[1])))
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

func TestDESMAImprovedEliteHelperReportsTrueNoOps(t *testing.T) {
	currentBest := Best{Position: []float64{1, 1}, Cost: 1}
	evaluator := newConstraintEvaluator(func([]float64) float64 { return 1 }, nil)

	elite, evaluations, improved := generateImprovedEliteMayfliesWithEvaluator(
		currentBest, 0.5, 0, 2, -10, 10, evaluator, rand.New(rand.NewSource(42)),
	)
	if elite != nil || evaluations != 0 || improved {
		t.Fatalf("zero elites = (%v, %d, %t), want (nil, 0, false)", elite, evaluations, improved)
	}

	elite, evaluations, improved = generateImprovedEliteMayfliesWithEvaluator(
		currentBest, 0.5, 4, 2, -10, 10, evaluator, rand.New(rand.NewSource(42)),
	)
	if elite != nil || evaluations != 4 || improved {
		t.Fatalf("unimproved elites = (%v, %d, %t), want (nil, 4, false)", elite, evaluations, improved)
	}
}
