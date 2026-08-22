package mayfly

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
)

// Progress describes the best solution known after a completed iteration.
// Iteration is one-based. Best and its Position are snapshots: observers may
// retain or modify them without affecting the optimizer.
type Progress struct { //nolint:govet // Preserve public field order for unkeyed composite literals.
	Iteration       int
	EvaluationCount int
	Best            Best
}

// ProgressObserver receives a snapshot after each completed iteration.
// OptimizeContext invokes observers synchronously on the calling goroutine.
type ProgressObserver func(Progress)

// PopulationSnapshot is the state of both populations after a completed
// iteration. Iteration is one-based. Every mayfly, and Best, is a deep copy:
// observers may retain or modify them without affecting the optimizer.
//
// It is deliberately separate from Progress rather than an extension of it.
// Copying NPop+NPopF position and velocity vectors once per iteration is not
// free, and the overwhelmingly common reason to observe a run is to watch the
// best cost fall, which Progress already answers. Callers who want the swarm
// itself — to animate it, to measure diversity, to debug a variant's search
// behavior — opt in with WithPopulationObserver and pay for it there.
type PopulationSnapshot struct {
	Males           []Mayfly
	Females         []Mayfly
	Best            Best
	Iteration       int
	EvaluationCount int
}

// PopulationObserver receives both populations after each completed iteration.
// OptimizeContext invokes observers synchronously on the calling goroutine.
type PopulationObserver func(PopulationSnapshot)

// Logger receives structured optimization lifecycle events. *slog.Logger
// implements Logger. OptimizeContext invokes loggers synchronously on the
// calling goroutine.
type Logger interface {
	Log(ctx context.Context, level slog.Level, message string, args ...any)
}

// RunOption customizes one optimization run. Its fields are intentionally
// private; construct options with WithInitialPopulation,
// WithProgressObserver, and WithLogger.
type RunOption struct {
	apply func(*runOptions) error
}

type runOptions struct {
	observer           ProgressObserver
	populationObserver PopulationObserver
	logger             Logger
	initialMales       [][]float64
	initialFemales     [][]float64
}

// WithInitialPopulation seeds the start of the male and female populations.
// Each argument may contain fewer positions than its configured population;
// unfilled slots are initialized randomly. The positions are copied when this
// function is called and again when applied to a run.
func WithInitialPopulation(males, females [][]float64) RunOption {
	maleSnapshot := clonePositions(males)
	femaleSnapshot := clonePositions(females)

	return RunOption{apply: func(options *runOptions) error {
		options.initialMales = clonePositions(maleSnapshot)
		options.initialFemales = clonePositions(femaleSnapshot)

		return nil
	}}
}

// WithProgressObserver registers an observer for iteration progress. Passing a
// nil observer disables progress reporting.
func WithProgressObserver(observer ProgressObserver) RunOption {
	return RunOption{apply: func(options *runOptions) error {
		options.observer = observer

		return nil
	}}
}

// WithPopulationObserver registers an observer for the male and female
// populations. It is called once per completed iteration, after
// WithProgressObserver's observer. Passing a nil observer disables population
// reporting, which is the default: no copying happens unless an observer is
// registered.
func WithPopulationObserver(observer PopulationObserver) RunOption {
	return RunOption{apply: func(options *runOptions) error {
		options.populationObserver = observer

		return nil
	}}
}

// WithLogger registers a structured logger for run lifecycle events. Passing
// nil disables logging. The logger receives optimization_started,
// iteration_completed, and optimization_completed events.
func WithLogger(logger Logger) RunOption {
	return RunOption{apply: func(options *runOptions) error {
		options.logger = logger

		return nil
	}}
}

func resolveRunOptions(options []RunOption) (runOptions, error) {
	var resolved runOptions

	for i, option := range options {
		if option.apply == nil {
			return runOptions{}, fmt.Errorf("run option %d is invalid", i)
		}

		err := option.apply(&resolved)
		if err != nil {
			return runOptions{}, fmt.Errorf("apply run option %d: %w", i, err)
		}
	}

	return resolved, nil
}

func validateInitialPopulation(config *Config, options runOptions) error {
	if len(options.initialMales) > config.NPop {
		return fmt.Errorf("initial male population has %d positions, exceeds NPop=%d",
			len(options.initialMales), config.NPop)
	}

	if len(options.initialFemales) > config.NPopF {
		return fmt.Errorf("initial female population has %d positions, exceeds NPopF=%d",
			len(options.initialFemales), config.NPopF)
	}

	err := validateInitialPositions("male", options.initialMales, config)
	if err != nil {
		return err
	}

	return validateInitialPositions("female", options.initialFemales, config)
}

func validateInitialPositions(kind string, positions [][]float64, config *Config) error {
	for i, position := range positions {
		if len(position) != config.ProblemSize {
			return fmt.Errorf("initial %s position %d has dimension %d, want %d",
				kind, i, len(position), config.ProblemSize)
		}

		for dimension, value := range position {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return fmt.Errorf("initial %s position %d dimension %d must be finite, got %v",
					kind, i, dimension, value)
			}

			if value < config.LowerBound || value > config.UpperBound {
				return fmt.Errorf(
					"initial %s position %d dimension %d is outside bounds [%v, %v]: %v",
					kind, i, dimension, config.LowerBound, config.UpperBound, value,
				)
			}
		}
	}

	return nil
}

func clonePositions(positions [][]float64) [][]float64 {
	if positions == nil {
		return nil
	}

	cloned := make([][]float64, len(positions))
	for i, position := range positions {
		cloned[i] = append([]float64(nil), position...)
	}

	return cloned
}

func cloneBest(best Best) Best {
	return Best{
		Position:            append([]float64(nil), best.Position...),
		Cost:                best.Cost,
		ConstraintViolation: best.ConstraintViolation,
	}
}

func notifyProgress(observer ProgressObserver, iteration, evaluationCount int, best Best) {
	if observer == nil {
		return
	}

	observer(Progress{
		Iteration:       iteration,
		EvaluationCount: evaluationCount,
		Best:            cloneBest(best),
	})
}

func notifyPopulation(
	observer PopulationObserver,
	iteration, evaluationCount int,
	best Best,
	males, females []*Mayfly,
) {
	if observer == nil {
		return
	}

	observer(PopulationSnapshot{
		Males:           cloneMayflies(males),
		Females:         cloneMayflies(females),
		Best:            cloneBest(best),
		Iteration:       iteration,
		EvaluationCount: evaluationCount,
	})
}

// cloneMayflies flattens a population of pointers into independent values, so
// an observer cannot reach back into the running optimizer through a shared
// slice header or a retained pointer.
func cloneMayflies(population []*Mayfly) []Mayfly {
	if population == nil {
		return nil
	}

	cloned := make([]Mayfly, len(population))
	for i, mayfly := range population {
		if mayfly == nil {
			continue
		}

		cloned[i] = *mayfly.clone()
	}

	return cloned
}

var errNilContext = errors.New("context cannot be nil")
