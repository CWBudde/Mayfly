package main

import (
	"errors"
	"fmt"
	"math"
	"os"

	"github.com/cwbudde/mayfly"
)

const parameterCount = 3

type observation struct {
	time  float64
	value float64
}

type interval struct {
	lower float64
	upper float64
}

type decayParameters struct {
	amplitude float64
	decay     float64
	baseline  float64
}

var parameterBounds = [parameterCount]interval{
	{lower: 0, upper: 10},
	{lower: 0.05, upper: 2},
	{lower: -2, upper: 2},
}

func scale(unit float64, bounds interval) float64 {
	return bounds.lower + unit*(bounds.upper-bounds.lower)
}

func decode(unit []float64) decayParameters {
	return decayParameters{
		amplitude: scale(unit[0], parameterBounds[0]),
		decay:     scale(unit[1], parameterBounds[1]),
		baseline:  scale(unit[2], parameterBounds[2]),
	}
}

func newDecayObjective(data []observation) (mayfly.ObjectiveFunction, error) {
	if len(data) == 0 {
		return nil, errors.New("at least one observation is required")
	}

	frozen := append([]observation(nil), data...)
	for _, sample := range frozen {
		if math.IsNaN(sample.time) || math.IsInf(sample.time, 0) ||
			math.IsNaN(sample.value) || math.IsInf(sample.value, 0) {
			return nil, errors.New("observations must be finite")
		}
	}

	return func(unit []float64) float64 {
		parameters := decode(unit)
		squaredError := 0.0

		for _, sample := range frozen {
			prediction := parameters.baseline +
				parameters.amplitude*math.Exp(-parameters.decay*sample.time)
			residual := prediction - sample.value
			squaredError += residual * residual
		}

		return squaredError / float64(len(frozen))
	}, nil
}

func fit(data []observation) (decayParameters, *mayfly.Result, error) {
	objective, err := newDecayObjective(data)
	if err != nil {
		return decayParameters{}, nil, err
	}

	seed := int64(42)
	config := mayfly.NewDefaultConfig()
	config.ObjectiveFunc = objective
	config.ProblemSize = parameterCount
	config.LowerBound = 0
	config.UpperBound = 1
	config.MaxIterations = 400
	config.Seed = &seed

	result, err := mayfly.Optimize(config)
	if err != nil {
		return decayParameters{}, nil, err
	}

	return decode(result.GlobalBest.Position), result, nil
}

func main() {
	data := []observation{
		{time: 0, value: 5.000000},
		{time: 0.5, value: 3.834613},
		{time: 1, value: 2.992588},
		{time: 2, value: 1.944636},
		{time: 3, value: 1.397553},
		{time: 4, value: 1.111950},
	}

	parameters, result, err := fit(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "curve fit failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "mean squared error: %.8g\n", result.GlobalBest.Cost)
	fmt.Fprintf(os.Stdout, "amplitude: %.6f; decay: %.6f; baseline: %.6f\n",
		parameters.amplitude,
		parameters.decay,
		parameters.baseline,
	)
	fmt.Fprintf(os.Stdout, "normalized position: %.6f, %.6f, %.6f\n",
		result.GlobalBest.Position[0],
		result.GlobalBest.Position[1],
		result.GlobalBest.Position[2],
	)
}
