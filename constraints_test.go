package mayfly

import (
	"context"
	"encoding/json"
	"math"
	"math/rand"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
)

func TestEvaluateConstraints(t *testing.T) {
	config := &ConstraintConfig{
		Inequalities: []ConstraintFunction{
			func(position []float64) float64 { return position[0] - 2 },
			func(position []float64) float64 { return -position[0] - 1 },
		},
		Equalities: []ConstraintFunction{
			func(position []float64) float64 { return position[1] - 3 },
		},
		EqualityTolerance: 0.25,
	}

	result := EvaluateConstraints([]float64{3, 3.5}, config)
	if result.Feasible {
		t.Fatal("constraint evaluation is feasible, want infeasible")
	}

	if result.Violation != 1.25 {
		t.Errorf("violation = %v, want 1.25", result.Violation)
	}

	feasible := EvaluateConstraints([]float64{1, 3.1}, config)
	if !feasible.Feasible || feasible.Violation != 0 {
		t.Errorf("feasible evaluation = %+v, want zero violation", feasible)
	}
}

func TestEvaluateConstraintsRejectsNonFiniteAndNilResults(t *testing.T) {
	tests := []ConstraintFunction{
		nil,
		func([]float64) float64 { return math.NaN() },
		func([]float64) float64 { return math.Inf(1) },
	}

	for i, constraint := range tests {
		result := EvaluateConstraints(nil, &ConstraintConfig{
			Inequalities: []ConstraintFunction{constraint},
		})
		if !math.IsInf(result.Violation, 1) || result.Feasible {
			t.Errorf("case %d = %+v, want infeasible infinite violation", i, result)
		}
	}
}

func TestPenaltyMethods(t *testing.T) {
	if got := PenalizedCost(10, 2, 3, PenaltyLinear); got != 16 {
		t.Errorf("linear penalty = %v, want 16", got)
	}

	if got := PenalizedCost(10, 2, 3, PenaltyQuadratic); got != 22 {
		t.Errorf("quadratic penalty = %v, want 22", got)
	}

	if got := PenalizedCost(10, 2, 3, ""); got != 22 {
		t.Errorf("default penalty = %v, want 22", got)
	}
}

func TestConstraintRankingMethods(t *testing.T) {
	feasible := CandidateEvaluation{Cost: 100}
	infeasible := CandidateEvaluation{Cost: 0, ConstraintViolation: 1}

	if !BetterConstrainedCandidate(feasible, infeasible, &ConstraintConfig{}) {
		t.Fatal("default feasibility rules did not prefer feasible candidate")
	}

	penalty := &ConstraintConfig{
		Handling:      ConstraintHandlingPenalty,
		PenaltyMethod: PenaltyLinear,
		PenaltyFactor: 0.1,
	}
	if !BetterConstrainedCandidate(infeasible, feasible, penalty) {
		t.Fatal("penalty method did not prefer lower penalized cost")
	}
}

func TestOptimizeConstraintHandlingMethods(t *testing.T) {
	run := func(constraints *ConstraintConfig) *Result {
		t.Helper()

		config := staticConstraintConfig()
		config.Constraints = constraints

		result, err := OptimizeContext(
			context.Background(),
			config,
			WithInitialPopulation(
				[][]float64{{0}, {1}},
				[][]float64{{0}, {1}},
			),
		)
		if err != nil {
			t.Fatalf("OptimizeContext: %v", err)
		}

		return result
	}

	constraint := func(position []float64) float64 { return 1 - position[0] }

	feasibility := run(&ConstraintConfig{Inequalities: []ConstraintFunction{constraint}})
	if feasibility.GlobalBest.ConstraintViolation != 0 || feasibility.GlobalBest.Cost != 1 {
		t.Errorf("feasibility best = %+v, want feasible cost 1", feasibility.GlobalBest)
	}

	penalty := run(&ConstraintConfig{
		Inequalities:  []ConstraintFunction{constraint},
		Handling:      ConstraintHandlingPenalty,
		PenaltyMethod: PenaltyLinear,
		PenaltyFactor: 0.01,
	})
	if penalty.GlobalBest.ConstraintViolation != 1 || penalty.GlobalBest.Cost != 0 {
		t.Errorf("penalty best = %+v, want infeasible raw cost 0 and violation 1", penalty.GlobalBest)
	}
}

func TestTargetCostRequiresFeasibleIncumbent(t *testing.T) {
	target := 0.0
	config := staticConstraintConfig()
	config.MaxIterations = 2
	config.Convergence = &ConvergenceConfig{TargetCost: &target}
	config.Constraints = &ConstraintConfig{
		Inequalities: []ConstraintFunction{
			func(position []float64) float64 { return 1 - position[0] },
		},
	}

	result, err := OptimizeContext(
		context.Background(),
		config,
		WithInitialPopulation(
			[][]float64{{0}, {0}},
			[][]float64{{0}, {0}},
		),
	)
	if err != nil {
		t.Fatalf("OptimizeContext: %v", err)
	}

	if result.TerminationReason != TerminationMaxIterations {
		t.Errorf("TerminationReason = %q, want %q",
			result.TerminationReason, TerminationMaxIterations)
	}
}

func TestConstraintViolationDrivesStagnation(t *testing.T) {
	tracker := newConvergenceTracker(
		&ConvergenceConfig{MinImprovement: 0.5, StagnationIterations: 2},
		Best{Cost: 0, ConstraintViolation: 2},
		newConstraintEvaluator(nil, &ConstraintConfig{}),
	)

	if _, stop := tracker.observe(1, Best{Cost: 10, ConstraintViolation: 1.4}); stop {
		t.Fatal("significant violation improvement stopped run")
	}

	if _, stop := tracker.observe(2, Best{Cost: 5, ConstraintViolation: 1.2}); stop {
		t.Fatal("first stagnant iteration stopped run")
	}

	reason, stop := tracker.observe(3, Best{Cost: 1, ConstraintViolation: 1.1})
	if !stop || reason != TerminationStagnation {
		t.Errorf("termination = (%q, %t), want (%q, true)", reason, stop, TerminationStagnation)
	}
}

func TestConstraintHandlingAppliesToEveryVariantAndSchedule(t *testing.T) {
	factories := []struct {
		name string
		new  func() *Config
	}{
		{name: "standard", new: NewDefaultConfig},
		{name: "DESMA", new: NewDESMAConfig},
		{name: "OLCE", new: NewOLCEConfig},
		{name: "EOBBMA", new: NewEOBBMAConfig},
		{name: "GSASMA", new: NewGSASMAConfig},
		{name: "MPMA", new: NewMPMAConfig},
		{name: "AOBLMOA", new: NewAOBLMOAConfig},
	}

	for _, factory := range factories {
		for _, parallel := range []bool{false, true} {
			name := factory.name + "/sequential"
			if parallel {
				name = factory.name + "/parallel"
			}

			t.Run(name, func(t *testing.T) {
				var objectiveCalls, constraintCalls atomic.Int64

				config := factory.new()
				config.ObjectiveFunc = func(position []float64) float64 {
					objectiveCalls.Add(1)

					return -position[0]
				}
				config.Constraints = &ConstraintConfig{
					Inequalities: []ConstraintFunction{func(position []float64) float64 {
						constraintCalls.Add(1)

						return position[0]
					}},
				}
				config.ProblemSize = 2
				config.LowerBound = -1
				config.UpperBound = 1
				config.MaxIterations = 1
				config.NPop = 6
				config.NPopF = 6
				config.NC = 0
				config.NM = 0
				config.Rand = rand.New(rand.NewSource(17))
				config.EnableParallel = parallel
				config.MaxWorkers = 3

				result, err := OptimizeContext(
					context.Background(),
					config,
					WithInitialPopulation([][]float64{{-0.5, 0}}, nil),
				)
				if err != nil {
					t.Fatalf("OptimizeContext: %v", err)
				}

				if result.GlobalBest.ConstraintViolation != 0 {
					t.Errorf("GlobalBest.ConstraintViolation = %v, want 0",
						result.GlobalBest.ConstraintViolation)
				}

				if constraintCalls.Load() != objectiveCalls.Load() {
					t.Errorf("constraint calls = %d, objective calls = %d",
						constraintCalls.Load(), objectiveCalls.Load())
				}
			})
		}
	}
}

func TestConstrainedParallelEvaluationIsDeterministicAcrossWorkerCounts(t *testing.T) {
	run := func(workers int) *Result {
		t.Helper()

		config := NewDefaultConfig()
		config.ObjectiveFunc = sphere
		config.Constraints = &ConstraintConfig{
			Inequalities: []ConstraintFunction{
				func(position []float64) float64 {
					return position[0] + position[1] + position[2]
				},
			},
		}
		config.ProblemSize = 3
		config.LowerBound = -2
		config.UpperBound = 2
		config.MaxIterations = 3
		config.NPop = 8
		config.NPopF = 8
		config.NC = 4
		config.NM = 2
		config.Rand = rand.New(rand.NewSource(44))
		config.EnableParallel = true
		config.MaxWorkers = workers

		result, err := Optimize(config)
		if err != nil {
			t.Fatalf("Optimize: %v", err)
		}

		return result
	}

	oneWorker := run(1)
	fourWorkers := run(4)

	if !reflect.DeepEqual(oneWorker.GlobalBest, fourWorkers.GlobalBest) ||
		!reflect.DeepEqual(oneWorker.ConvergenceCurve, fourWorkers.ConvergenceCurve) ||
		oneWorker.FuncEvalCount != fourWorkers.FuncEvalCount {
		t.Errorf("constrained results differ by worker count: one=%+v four=%+v",
			oneWorker, fourWorkers)
	}
}

func TestConstraintConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		constraints *ConstraintConfig
	}{
		{name: "negative tolerance", constraints: &ConstraintConfig{EqualityTolerance: -1}},
		{name: "nil inequality", constraints: &ConstraintConfig{Inequalities: []ConstraintFunction{nil}}},
		{name: "unknown handling", constraints: &ConstraintConfig{Handling: "unknown"}},
		{name: "non-finite penalty", constraints: &ConstraintConfig{PenaltyFactor: math.NaN()}},
		{name: "zero penalty", constraints: &ConstraintConfig{Handling: ConstraintHandlingPenalty}},
		{name: "unknown penalty", constraints: &ConstraintConfig{
			PenaltyMethod: "unknown",
		}},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			config := staticConstraintConfig()
			config.Constraints = testCase.constraints

			err := ValidateConfig(config)
			if err == nil ||
				!strings.Contains(err.Error(), "invalid constraint config") {
				t.Errorf("ValidateConfig error = %v, want constraint error", err)
			}

			_, err = Optimize(config)
			if err == nil ||
				!strings.Contains(err.Error(), "invalid constraint config") {
				t.Errorf("Optimize error = %v, want constraint error", err)
			}
		})
	}
}

func TestConstraintConfigJSONRoundTrip(t *testing.T) {
	config := staticConstraintConfig()
	config.Constraints = &ConstraintConfig{
		Inequalities:      []ConstraintFunction{func([]float64) float64 { return 0 }},
		Equalities:        []ConstraintFunction{func([]float64) float64 { return 0 }},
		Handling:          ConstraintHandlingPenalty,
		PenaltyMethod:     PenaltyLinear,
		PenaltyFactor:     12,
		EqualityTolerance: 1e-4,
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded Config

	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.Constraints == nil ||
		decoded.Constraints.Handling != ConstraintHandlingPenalty ||
		decoded.Constraints.PenaltyMethod != PenaltyLinear ||
		decoded.Constraints.PenaltyFactor != 12 ||
		decoded.Constraints.EqualityTolerance != 1e-4 {
		t.Fatalf("decoded constraints = %+v", decoded.Constraints)
	}

	if decoded.Constraints.Inequalities != nil || decoded.Constraints.Equalities != nil {
		t.Fatal("constraint functions were serialized")
	}
}

func staticConstraintConfig() *Config {
	config := NewDefaultConfig()
	config.ObjectiveFunc = func(position []float64) float64 { return position[0] * position[0] }
	config.ProblemSize = 1
	config.LowerBound = 0
	config.UpperBound = 2
	config.MaxIterations = 1
	config.NPop = 2
	config.NPopF = 2
	config.NC = 0
	config.NM = 0
	config.G = 0
	config.A1 = 0
	config.A2 = 0
	config.A3 = 0
	config.Dance = 0
	config.FL = 0
	config.Rand = rand.New(rand.NewSource(1))

	return config
}
