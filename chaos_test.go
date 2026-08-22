package mayfly

import (
	"math"
	"math/rand"
	"testing"
)

func TestLogisticMapSequenceAndState(t *testing.T) {
	logisticMap := NewLogisticMap(0.25)
	if logisticMap.Current() != 0.25 {
		t.Errorf("initial state = %v, want 0.25", logisticMap.Current())
	}

	if next := logisticMap.Next(); next != 0.75 {
		t.Errorf("next state = %v, want 0.75", next)
	}

	if logisticMap.Current() != 0.75 {
		t.Errorf("current state = %v, want 0.75", logisticMap.Current())
	}
}

func TestLogisticMapNormalizesInvalidSeeds(t *testing.T) {
	tests := []struct {
		name string
		seed float64
		want float64
	}{
		{name: "zero", seed: 0, want: 0.1},
		{name: "positive_fraction", seed: 1.25, want: 0.3},
		{name: "negative_fraction", seed: -0.5, want: 0.314159},
		{name: "positive_infinity", seed: math.Inf(1), want: 0.271828},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logisticMap := NewLogisticMap(tt.seed)
			if math.Abs(logisticMap.Current()-tt.want) > epsilon {
				t.Errorf("normalized state = %v, want %v", logisticMap.Current(), tt.want)
			}
		})
	}
}

func TestLogisticMapResetAndBoundarySafeguards(t *testing.T) {
	logisticMap := NewLogisticMap(0.25)
	logisticMap.Reset(0.4)

	if logisticMap.Current() != 0.4 {
		t.Errorf("reset state = %v, want 0.4", logisticMap.Current())
	}

	logisticMap.Reset(-0.5)

	if logisticMap.Current() != 0.314159 {
		t.Errorf("normalized reset state = %v, want 0.314159", logisticMap.Current())
	}

	logisticMap.Reset(math.Inf(1))

	if logisticMap.Current() != 0.271828 {
		t.Errorf("infinite reset state = %v, want 0.271828", logisticMap.Current())
	}

	logisticMap.x = 0
	if next := logisticMap.Next(); next != 1e-10 {
		t.Errorf("lower boundary safeguard = %v, want 1e-10", next)
	}

	logisticMap.x = 0.5
	if next := logisticMap.Next(); next != 1-1e-10 {
		t.Errorf("upper boundary safeguard = %v, want %v", next, 1-1e-10)
	}
}

func TestChaoticExploitationRadiusDecays(t *testing.T) {
	config := NewOLCEConfig()
	config.ChaosFactor = 0.4
	config.MaxIterations = 100

	// The loop applies the iterations 0 to MaxIterations-1, so the decay is
	// measured against the last applied iteration, 99.
	testCases := []struct {
		name      string
		iteration int
		want      float64
	}{
		{name: "start", iteration: 0, want: 0.4},
		{name: "half", iteration: 50, want: 0.4 * (1.0 - 50.0/99.0)},
		{name: "last applied iteration", iteration: 99, want: 0},
		{name: "beyond end", iteration: 150, want: 0},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := chaoticExploitationRadius(config, testCase.iteration)
			if math.Abs(got-testCase.want) > 1e-12 {
				t.Errorf("radius at iteration %d = %v, want %v",
					testCase.iteration, got, testCase.want)
			}
		})
	}

	// A two-iteration run must span the full range instead of stopping at half
	// the factor.
	config.MaxIterations = 2

	if got := chaoticExploitationRadius(config, 0); math.Abs(got-config.ChaosFactor) > 1e-12 {
		t.Errorf("radius of first of two iterations = %v, want %v", got, config.ChaosFactor)
	}

	if got := chaoticExploitationRadius(config, 1); math.Abs(got) > 1e-12 {
		t.Errorf("radius of last of two iterations = %v, want 0", got)
	}

	// A single iteration has no decay to spread out and keeps the full factor,
	// because a zero radius would make its only exploitation step a no-op.
	config.MaxIterations = 1

	if got := chaoticExploitationRadius(config, 0); got != config.ChaosFactor {
		t.Errorf("radius of a one-iteration run = %v, want %v", got, config.ChaosFactor)
	}

	config.MaxIterations = 0

	if got := chaoticExploitationRadius(config, 7); got != config.ChaosFactor {
		t.Errorf("radius without iteration budget = %v, want %v", got, config.ChaosFactor)
	}
}

// TestChaoticExploitationRejectsNaNCandidate pins that a candidate the
// objective cannot evaluate is rejected instead of being copied into the
// individual, where its NaN cost would spread through mating and sorting.
func TestChaoticExploitationRejectsNaNCandidate(t *testing.T) {
	config := NewOLCEConfig()
	config.ProblemSize = 3
	config.LowerBound = -5
	config.UpperBound = 5
	config.MaxIterations = 10
	config.ChaosFactor = 0.5

	evaluator := newConstraintEvaluator(func(_ []float64) float64 {
		return math.NaN()
	}, nil)

	target := newMayfly(config.ProblemSize)
	target.Cost = 42
	copy(target.Best.Position, target.Position)
	target.Best.Cost = target.Cost

	position := make([]float64, config.ProblemSize)
	copy(position, target.Position)

	if applyChaoticExploitation(target, config, NewLogisticMap(0.19), 0, evaluator) {
		t.Fatal("chaotic exploitation accepted a candidate with a NaN cost")
	}

	if target.Cost != 42 {
		t.Errorf("cost = %v, want 42", target.Cost)
	}

	for j, value := range target.Position {
		if value != position[j] {
			t.Errorf("dimension %d moved to %v, want %v", j, value, position[j])
		}
	}
}

// TestChaoticExploitationNeverWorsensIndividual pins the greedy acceptance of
// the chaotic exploitation step. Before this was a greedy step it displaced
// every individual unconditionally, which is a persistent random walk.
func TestChaoticExploitationNeverWorsensIndividual(t *testing.T) {
	config := NewOLCEConfig()
	config.ProblemSize = 5
	config.LowerBound = -10
	config.UpperBound = 10
	config.MaxIterations = 200
	config.ChaosFactor = 0.5

	evaluator := newConstraintEvaluator(Rastrigin, nil)
	chaosMap := NewLogisticMap(0.37)

	target := newMayfly(config.ProblemSize)
	for j := range target.Position {
		target.Position[j] = 3.5 - 0.25*float64(j)
	}

	evaluator.evaluateMayfly(target, false)
	copy(target.Best.Position, target.Position)
	target.Best.Cost = target.Cost

	accepted := 0

	for iteration := range config.MaxIterations {
		previousCost := target.Cost

		if applyChaoticExploitation(target, config, chaosMap, iteration, evaluator) {
			accepted++
		}

		if target.Cost > previousCost {
			t.Fatalf("iteration %d: cost rose from %v to %v", iteration, previousCost, target.Cost)
		}

		if target.Cost != Rastrigin(target.Position) {
			t.Fatalf("iteration %d: cost %v does not match position", iteration, target.Cost)
		}

		if target.Best.Cost > target.Cost {
			t.Fatalf("iteration %d: personal best %v worse than current %v",
				iteration, target.Best.Cost, target.Cost)
		}

		for j, value := range target.Position {
			if value < config.LowerBound || value > config.UpperBound {
				t.Fatalf("iteration %d: dimension %d left bounds: %v", iteration, j, value)
			}
		}
	}

	if accepted == 0 {
		t.Error("chaotic exploitation never accepted a candidate")
	}
}

// TestChaoticExploitationSpendsOneEvaluation guards the evaluation budget of
// the step: exactly one objective call per individual and iteration.
func TestChaoticExploitationSpendsOneEvaluation(t *testing.T) {
	config := NewOLCEConfig()
	config.ProblemSize = 3
	config.LowerBound = -5
	config.UpperBound = 5
	config.MaxIterations = 10

	calls := 0
	evaluator := newConstraintEvaluator(func(position []float64) float64 {
		calls++

		return Sphere(position)
	}, nil)

	target := newMayfly(config.ProblemSize)
	evaluator.evaluateMayfly(target, false)

	calls = 0

	applyChaoticExploitation(target, config, NewLogisticMap(0.21), 0, evaluator)

	if calls != 1 {
		t.Errorf("objective calls = %d, want 1", calls)
	}
}

// TestOLCEConvergesOnUnimodalProblem is the end-to-end guard for the same
// property. An unconditional chaotic kick keeps displacing the population
// every iteration, which caps the reachable precision several orders of
// magnitude above what greedy, decaying exploitation reaches.
func TestOLCEConvergesOnUnimodalProblem(t *testing.T) {
	const tolerance = 1e-20

	for _, seed := range []int64{1, 2, 3} {
		config := NewOLCEConfig()
		config.ObjectiveFunc = Sphere
		config.ProblemSize = 5
		config.LowerBound = -5
		config.UpperBound = 5
		config.MaxIterations = 200
		config.Rand = rand.New(rand.NewSource(seed))

		result, err := Optimize(config)
		if err != nil {
			t.Fatalf("Optimize with seed %d: %v", seed, err)
		}

		if result.GlobalBest.Cost > tolerance {
			t.Errorf("seed %d: best cost = %v, want at most %v",
				seed, result.GlobalBest.Cost, tolerance)
		}
	}
}
