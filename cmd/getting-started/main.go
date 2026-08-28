package main

import (
	"fmt"
	"os"

	"github.com/cwbudde/mayfly"
)

func objective(x []float64) float64 {
	dx := x[0] - 1.5
	dy := x[1] + 0.5

	return dx*dx + dy*dy
}

func optimize() (*mayfly.Result, error) {
	seed := int64(42)
	config := mayfly.NewDefaultConfig()
	config.ObjectiveFunc = objective
	config.ProblemSize = 2
	config.LowerBound = -5
	config.UpperBound = 5
	config.MaxIterations = 200
	config.Seed = &seed

	return mayfly.Optimize(config)
}

func main() {
	result, err := optimize()
	if err != nil {
		fmt.Fprintf(os.Stderr, "optimization failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "best cost: %.8g\n", result.GlobalBest.Cost)
	fmt.Fprintf(os.Stdout, "best position: %.6f, %.6f\n",
		result.GlobalBest.Position[0],
		result.GlobalBest.Position[1],
	)
	fmt.Fprintf(os.Stdout, "iterations: %d; evaluations: %d; stopped: %s\n",
		result.IterationCount,
		result.FuncEvalCount,
		result.TerminationReason,
	)
}
