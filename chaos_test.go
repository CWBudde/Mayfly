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
		{name: "not_a_number", seed: math.NaN(), want: 0.314159},
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

// TestChaoticConstrictionFactorUsesOneBasedGeneration pins s=(G-g+1)/G
// from the cited OLCE-MA chaotic-offspring strategy. The optimizer uses
// zero-based iteration indices, while the paper uses generations 1..G.
func TestChaoticConstrictionFactorUsesOneBasedGeneration(t *testing.T) {
	config := NewOLCEConfig()
	config.MaxIterations = 4

	testCases := []struct {
		name      string
		iteration int
		want      float64
	}{
		{name: "generation one", iteration: 0, want: 1},
		{name: "generation two", iteration: 1, want: 0.75},
		{name: "generation four", iteration: 3, want: 0.25},
		{name: "clamped after run", iteration: 9, want: 0.25},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := chaoticConstrictionFactor(config, testCase.iteration)
			if math.Abs(got-testCase.want) > 1e-12 {
				t.Errorf("constriction at iteration %d = %v, want %v",
					testCase.iteration, got, testCase.want)
			}
		})
	}

	config.ChaosFactor = 0.4
	if got := chaoticConstrictionFactor(config, 1); math.Abs(got-0.3) > 1e-12 {
		t.Errorf("compatibility-scaled constriction = %v, want 0.3", got)
	}
}

// TestChaoticExploitationCandidateEquation pins the published offspring
// construction: C'=LB+C(UB-LB), O'=(1-s)O+sC'. The paper says "the fittest
// offspring's position" (singular), which is why the optimizer applies this
// once to the best crossover child rather than to parents or every child.
func TestChaoticExploitationCandidateEquation(t *testing.T) {
	config := NewOLCEConfig()
	config.LowerBound = -2
	config.UpperBound = 6
	config.MaxIterations = 4
	destination := make([]float64, 2)

	// Seed .25 advances to .75 for both first two logistic-map samples.
	chaoticExploitationCandidate(
		destination, []float64{0, 2}, config, NewLogisticMap(0.25),
		chaoticConstrictionFactor(config, 1),
	)

	want := []float64{3, 3.5} // .25*offspring + .75*4
	for dimension := range destination {
		if math.Abs(destination[dimension]-want[dimension]) > 1e-12 {
			t.Errorf("dimension %d = %v, want %v", dimension, destination[dimension], want[dimension])
		}
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

// TestChaoticExploitationFormsNewOffspringUnconditionally pins the paper's
// lifecycle semantics: Eq. (12) forms the new offspring position; it does not
// describe choosing the better of that position and the crossover source.
func TestChaoticExploitationFormsNewOffspringUnconditionally(t *testing.T) {
	config := NewOLCEConfig()
	config.ProblemSize = 1
	config.LowerBound = -1
	config.UpperBound = 1
	config.MaxIterations = 1

	evaluator := newConstraintEvaluator(Sphere, nil)

	target := newMayfly(config.ProblemSize)
	evaluator.evaluateMayfly(target, false)
	copy(target.Best.Position, target.Position)
	target.Best.Cost = target.Cost

	if !applyChaoticExploitation(target, config, NewLogisticMap(0.25), 0, evaluator) {
		t.Fatal("finite chaotic offspring was not committed")
	}
	// The first logistic value is .75, mapping to .5 in [-1,1]. This is
	// deliberately worse than the crossover offspring at zero.
	if target.Position[0] != 0.5 || target.Cost != 0.25 {
		t.Fatalf("offspring = (%v, cost %v), want (0.5, 0.25)", target.Position[0], target.Cost)
	}

	if target.Best.Position[0] != 0.5 || target.Best.Cost != 0.25 {
		t.Fatalf("offspring best was not initialized from its formed position: %+v", target.Best)
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

func TestOptimizeOLCEAppliesChaosOncePerGeneration(t *testing.T) {
	newConfig := func(chaosFactor float64) *Config {
		config := NewOLCEConfig()
		config.ObjectiveFunc = Sphere
		config.ProblemSize = 2
		config.LowerBound = -5
		config.UpperBound = 5
		config.MaxIterations = 6
		config.NPop = 4
		config.NPopF = 4
		config.OrthogonalFactor = 0
		config.ChaosFactor = chaosFactor
		config.Rand = rand.New(rand.NewSource(41))

		return config
	}

	disabled, err := Optimize(newConfig(0))
	if err != nil {
		t.Fatalf("Optimize without chaotic offspring: %v", err)
	}

	enabled, err := Optimize(newConfig(1))
	if err != nil {
		t.Fatalf("Optimize with chaotic offspring: %v", err)
	}

	if difference := enabled.FuncEvalCount - disabled.FuncEvalCount; difference != 6 {
		t.Fatalf("chaotic offspring evaluations = %d, want one for each of 6 generations", difference)
	}
}
