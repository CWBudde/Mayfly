package mayfly

import (
	"errors"
	"fmt"
	"math"
)

// BenchmarkCase describes a reproducible optimization problem. Coordinates
// accepted by Objective and Constraints are physical coordinates in Bounds.
// Use NewConfig to adapt heterogeneous bounds to Mayfly's scalar search box.
type BenchmarkCase struct {
	objective      ObjectiveFunction
	constraints    *ConstraintConfig
	project        func([]float64) []float64
	suite          string
	name           string
	lower          []float64
	upper          []float64
	optimum        []float64
	number         int
	dimension      int
	minimum        float64
	maxEvaluations int
}

// Suite returns the benchmark collection name, such as "CEC2017".
func (problem *BenchmarkCase) Suite() string { return problem.suite }

// Name returns the problem's descriptive name.
func (problem *BenchmarkCase) Name() string { return problem.name }

// Number returns the one-based function number within a numbered suite. It is
// zero for unnumbered engineering problems.
func (problem *BenchmarkCase) Number() int { return problem.number }

// Dimension returns the number of decision variables.
func (problem *BenchmarkCase) Dimension() int { return problem.dimension }

// Bounds returns defensive copies of the physical lower and upper bounds.
func (problem *BenchmarkCase) Bounds() ([]float64, []float64) {
	return append([]float64(nil), problem.lower...), append([]float64(nil), problem.upper...)
}

// Optimum returns a defensive copy of a known global minimizer. A nil result
// means that the suite does not publish one for this case.
func (problem *BenchmarkCase) Optimum() []float64 {
	return append([]float64(nil), problem.optimum...)
}

// Minimum returns the objective value at the published global minimizer.
func (problem *BenchmarkCase) Minimum() float64 { return problem.minimum }

// MaxEvaluations returns the competition evaluation budget, or zero when the
// source does not define a single budget.
func (problem *BenchmarkCase) MaxEvaluations() int { return problem.maxEvaluations }

// Objective returns a concurrency-safe objective over physical coordinates.
// Invalid inputs score +Inf so the function can be passed directly to an
// optimizer; Evaluate provides detailed validation errors for callers.
func (problem *BenchmarkCase) Objective() ObjectiveFunction {
	return func(position []float64) float64 {
		value, err := problem.Evaluate(position)
		if err != nil {
			return math.Inf(1)
		}

		return value
	}
}

// Constraints returns a defensive snapshot of the physical-coordinate
// constraints. A nil result means that the problem is bound-constrained only.
func (problem *BenchmarkCase) Constraints() *ConstraintConfig {
	return cloneBenchmarkConstraints(problem.constraints)
}

// Evaluate validates and evaluates a physical-coordinate position.
func (problem *BenchmarkCase) Evaluate(position []float64) (float64, error) {
	if problem == nil {
		return 0, errors.New("benchmark problem is nil")
	}

	if len(position) != problem.dimension {
		return 0, fmt.Errorf(
			"position dimension %d does not match benchmark dimension %d",
			len(position), problem.dimension,
		)
	}

	physical := append([]float64(nil), position...)
	for i, coordinate := range physical {
		if !isFinite(coordinate) {
			return 0, fmt.Errorf("position %d is not finite", i)
		}

		if coordinate < problem.lower[i] || coordinate > problem.upper[i] {
			return 0, fmt.Errorf(
				"position %d=%v is outside [%v,%v]",
				i, coordinate, problem.lower[i], problem.upper[i],
			)
		}
	}

	if problem.project != nil {
		physical = problem.project(physical)
	}

	value := problem.objective(physical)
	if !isFinite(value) {
		return 0, errors.New("benchmark objective returned a non-finite value")
	}

	return value, nil
}

// Decode maps a point in the unit hypercube to physical coordinates and
// applies any discrete-variable projection used by the problem.
func (problem *BenchmarkCase) Decode(normalized []float64) ([]float64, error) {
	if problem == nil {
		return nil, errors.New("benchmark problem is nil")
	}

	if len(normalized) != problem.dimension {
		return nil, fmt.Errorf(
			"position dimension %d does not match benchmark dimension %d",
			len(normalized), problem.dimension,
		)
	}

	physical := make([]float64, problem.dimension)

	for i, coordinate := range normalized {
		if !isFinite(coordinate) || coordinate < 0 || coordinate > 1 {
			return nil, fmt.Errorf("normalized position %d=%v is outside [0,1]", i, coordinate)
		}

		physical[i] = problem.lower[i] + coordinate*(problem.upper[i]-problem.lower[i])
	}

	if problem.project != nil {
		physical = problem.project(physical)
	}

	return physical, nil
}

// NewConfig returns a copy of base configured for this benchmark. The search
// takes place in [0,1]^D and objective/constraint calls are decoded to physical
// coordinates, which supports heterogeneous and discrete engineering bounds.
// Passing nil uses NewDefaultConfig.
func (problem *BenchmarkCase) NewConfig(base *Config) (*Config, error) {
	if problem == nil {
		return nil, errors.New("benchmark problem is nil")
	}

	if base == nil {
		base = NewDefaultConfig()
	}

	config := cloneComparisonConfig(base)
	config.ProblemSize = problem.dimension
	config.LowerBound = 0
	config.UpperBound = 1
	config.VelMin = 0
	config.VelMax = 0
	config.ObjectiveFunc = func(normalized []float64) float64 {
		physical, err := problem.Decode(normalized)
		if err != nil {
			return math.Inf(1)
		}

		return problem.objective(physical)
	}

	config.Constraints = normalizedBenchmarkConstraints(problem)

	err := ValidateConfig(config)
	if err != nil {
		return nil, fmt.Errorf("configure benchmark: %w", err)
	}

	return config, nil
}

func normalizedBenchmarkConstraints(problem *BenchmarkCase) *ConstraintConfig {
	if problem.constraints == nil {
		return nil
	}

	result := cloneBenchmarkConstraints(problem.constraints)

	wrap := func(constraint ConstraintFunction) ConstraintFunction {
		return func(normalized []float64) float64 {
			physical, err := problem.Decode(normalized)
			if err != nil {
				return math.Inf(1)
			}

			return constraint(physical)
		}
	}
	for i, constraint := range result.Inequalities {
		result.Inequalities[i] = wrap(constraint)
	}

	for i, constraint := range result.Equalities {
		result.Equalities[i] = wrap(constraint)
	}

	return result
}

func cloneBenchmarkConstraints(config *ConstraintConfig) *ConstraintConfig {
	if config == nil {
		return nil
	}

	clone := *config
	clone.Inequalities = append([]ConstraintFunction(nil), config.Inequalities...)
	clone.Equalities = append([]ConstraintFunction(nil), config.Equalities...)

	return &clone
}

func newBenchmarkCase(problem BenchmarkCase) (*BenchmarkCase, error) {
	if problem.suite == "" || problem.name == "" {
		return nil, errors.New("benchmark suite and name are required")
	}

	if problem.dimension <= 0 || len(problem.lower) != problem.dimension || len(problem.upper) != problem.dimension {
		return nil, errors.New("benchmark bounds do not match its dimension")
	}

	if problem.objective == nil {
		return nil, errors.New("benchmark objective is nil")
	}

	for i := range problem.dimension {
		if !isFinite(problem.lower[i]) || !isFinite(problem.upper[i]) || problem.lower[i] >= problem.upper[i] {
			return nil, fmt.Errorf("invalid benchmark bounds at dimension %d", i)
		}
	}

	if len(problem.optimum) != 0 && len(problem.optimum) != problem.dimension {
		return nil, errors.New("benchmark optimum does not match its dimension")
	}

	if !isFinite(problem.minimum) {
		return nil, errors.New("benchmark minimum is not finite")
	}

	err := validateConstraintConfig(problem.constraints)
	if err != nil {
		return nil, fmt.Errorf("invalid benchmark constraints: %w", err)
	}

	problem.lower = append([]float64(nil), problem.lower...)
	problem.upper = append([]float64(nil), problem.upper...)
	problem.optimum = append([]float64(nil), problem.optimum...)
	problem.constraints = cloneBenchmarkConstraints(problem.constraints)

	return &problem, nil
}

func repeatedBounds(dimension int, lower, upper float64) ([]float64, []float64) {
	lowers := make([]float64, dimension)

	uppers := make([]float64, dimension)
	for i := range dimension {
		lowers[i] = lower
		uppers[i] = upper
	}

	return lowers, uppers
}
