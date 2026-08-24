package mayfly

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
)

const (
	eventOptimizationStarted   = "optimization_started"
	eventIterationCompleted    = "iteration_completed"
	eventOptimizationCompleted = "optimization_completed"
)

var errNilResult = errors.New("result cannot be nil")

// ConvergencePoint is one exported sample from a convergence curve. Iteration
// is one-based and BestCost is the best cost known after that iteration.
type ConvergencePoint struct {
	Iteration int     `json:"iteration"`
	BestCost  float64 `json:"best_cost"`
}

func logOptimizationStarted(ctx context.Context, logger Logger, config *Config) {
	if logger == nil {
		return
	}

	logger.Log(
		ctx,
		slog.LevelInfo,
		"optimization started",
		"event", eventOptimizationStarted,
		"problem_size", config.ProblemSize,
		"max_iterations", config.MaxIterations,
		"male_population", config.NPop,
		"female_population", config.NPopF,
		"parallel", config.EnableParallel,
	)
}

func logIterationCompleted(
	ctx context.Context,
	logger Logger,
	iteration, evaluationCount int,
	best Best,
) {
	if logger == nil {
		return
	}

	logger.Log(
		ctx,
		slog.LevelInfo,
		"optimization iteration completed",
		"event", eventIterationCompleted,
		"iteration", iteration,
		"evaluations", evaluationCount,
		"best_cost", best.Cost,
		"constraint_violation", best.ConstraintViolation,
	)
}

func logOptimizationCompleted(ctx context.Context, logger Logger, result *Result) {
	if logger == nil {
		return
	}

	logger.Log(
		ctx,
		slog.LevelInfo,
		"optimization completed",
		"event", eventOptimizationCompleted,
		"iterations", result.IterationCount,
		"evaluations", result.FuncEvalCount,
		"best_cost", result.GlobalBest.Cost,
		"constraint_violation", result.GlobalBest.ConstraintViolation,
		"termination_reason", result.TerminationReason,
	)
}

func (result *Result) convergencePoints() ([]ConvergencePoint, error) {
	if result == nil {
		return nil, errNilResult
	}

	points := make([]ConvergencePoint, len(result.ConvergenceCurve))
	for i, cost := range result.ConvergenceCurve {
		points[i] = ConvergencePoint{Iteration: i + 1, BestCost: cost}
	}

	return points, nil
}

// ExportConvergenceCSV writes the convergence curve to path as iteration and
// best_cost columns, with one row per completed iteration.
func (result *Result) ExportConvergenceCSV(path string) (returnErr error) {
	points, err := result.convergencePoints()
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create convergence CSV: %w", err)
	}

	defer func() {
		closeErr := file.Close()
		if returnErr == nil && closeErr != nil {
			returnErr = fmt.Errorf("close convergence CSV: %w", closeErr)
		}
	}()

	writer := csv.NewWriter(file)

	err = writer.Write([]string{"iteration", "best_cost"})
	if err != nil {
		return fmt.Errorf("write convergence CSV header: %w", err)
	}

	for _, point := range points {
		err = writer.Write([]string{
			strconv.Itoa(point.Iteration),
			strconv.FormatFloat(point.BestCost, 'g', -1, 64),
		})
		if err != nil {
			return fmt.Errorf("write convergence CSV row: %w", err)
		}
	}

	writer.Flush()

	err = writer.Error()
	if err != nil {
		return fmt.Errorf("flush convergence CSV: %w", err)
	}

	return nil
}

// ExportConvergenceJSON writes the convergence curve to path as an indented
// JSON array of ConvergencePoint values.
func (result *Result) ExportConvergenceJSON(path string) error {
	points, err := result.convergencePoints()
	if err != nil {
		return err
	}

	return writeJSONAtomic(points, path)
}
