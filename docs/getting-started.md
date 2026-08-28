# Getting Started with Mayfly

Mayfly is a Go library for bounded, continuous, single-objective optimization.
Give it a function that maps a vector of numbers to a cost, and it searches for
the vector with the lowest cost.

This tutorial builds and runs a complete optimizer, explains its result, and
shows the changes needed for a real objective function.

## Prerequisites and installation

Use a supported Go release and start from a Go module:

```bash
mkdir mayfly-demo
cd mayfly-demo
go mod init example.com/mayfly-demo
go get github.com/cwbudde/mayfly
```

## Basic usage

Create `main.go` with the following complete program. The same source is kept
as the runnable [`cmd/getting-started`](../cmd/getting-started/)
example and is checked in CI.

```go
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
```

Run it:

```bash
go run .
```

The known minimum of this example is at `[1.5, -0.5]`, where the cost is zero.
Mayfly is stochastic, so expect a nearby position and a small nonnegative cost,
not necessarily the exact analytical answer.

## What the configuration means

Every run needs these four problem fields:

- `ObjectiveFunc` is the function to minimize.
- `ProblemSize` is the number of values in each candidate vector.
- `LowerBound` and `UpperBound` define one finite search interval shared by
  every dimension.

`NewDefaultConfig` supplies the algorithm parameters. The example lowers
`MaxIterations` to keep the first run quick. Start with those defaults; change
population and algorithm parameters only after measuring repeated runs on the
actual objective.

The fixed `Seed` makes repeated runs with the same Mayfly version and
configuration reproducible. It is useful for tests and debugging. Use several
independent seeds when judging solution quality, because one seeded run is not
a performance comparison.

## Reading the result

`Optimize` returns a `*mayfly.Result` and an error. Check the error before using
the result. Its main fields are:

- `GlobalBest.Position`: the best candidate vector found.
- `GlobalBest.Cost`: that candidate's objective value.
- `IterationCount` and `FuncEvalCount`: work actually performed.
- `TerminationReason`: whether the iteration limit, a target cost, or
  stagnation stopped the run.
- `ConvergenceCurve`: the best cost recorded at each completed iteration.

A finite result only means that the search completed. It does not prove a
global optimum. Validate the returned position against domain rules and, for
important work, compare repeated seeded runs with a baseline method.

## Use your own objective

Replace `objective` with a function of type
`func([]float64) float64`. For example, least-squares fitting of a line can use
two decision variables, slope and intercept:

```go
func squaredError(parameters []float64) float64 {
	slope, intercept := parameters[0], parameters[1]
	x := []float64{1, 2, 3, 4}
	y := []float64{2.2, 3.9, 6.1, 7.8}

	total := 0.0
	for i := range x {
		residual := slope*x[i] + intercept - y[i]
		total += residual * residual
	}

	return total
}
```

Set `ProblemSize` to `2`, choose bounds that are meaningful for both
parameters, and assign `squaredError` to `ObjectiveFunc`.

The optimizer minimizes. To maximize a score such as profit, return its
negative from the objective and negate `GlobalBest.Cost` when reporting it.
The objective should return a finite value for every vector inside the search
box. If `EnableParallel` is true, it and any constraint functions must also be
safe for concurrent calls.

## Different bounds and constraints

The core configuration uses the same bounds for every dimension. When
variables need different ranges, optimize normalized values in `[0, 1]` and
decode each component before evaluating the domain objective:

```go
func decode(unit, lower, upper []float64) []float64 {
	actual := make([]float64, len(unit))
	for i := range unit {
		actual[i] = lower[i] + unit[i]*(upper[i]-lower[i])
	}

	return actual
}
```

Set `LowerBound = 0` and `UpperBound = 1`; keep `ProblemSize` equal to the
number of domain variables. For constrained optimization, inequalities use
`g(x) <= 0` and equalities use `h(x) = 0`. The
[configuration guide](api/configuration.md#constraint-handling) covers feasibility
rules, penalties, and equality tolerance.

## Common first-run problems

- `ObjectiveFunc is required`: assign the objective before calling
  `Optimize`.
- `ProblemSize must be positive`: match it to the number of values read by the
  objective.
- `LowerBound must be less than UpperBound`: use finite bounds in ascending
  order.
- A panic while indexing `x`: the objective expects more components than
  `ProblemSize` provides.
- Results that look plausible but violate the domain: encode the constraint or
  decode/project the candidate before evaluation; do not silently accept it.

## Where to go next

- [Custom-objective Guide](custom-objective-functions.md) for modeling,
  validating, and safely evaluating a real objective.
- [API quick reference](api/quick-reference.md) for entry points and result
  fields.
- [Configuration guide](api/configuration.md) for constraints, stopping,
  parallel evaluation, and parameter definitions.
- [Run lifecycle](api/run-lifecycle.md) for cancellation, progress observers,
  logging, and initial populations.
- [Benchmark functions](benchmarks.md) for test objectives with known minima.
- [Algorithm documentation](algorithms/) before selecting a specialized
  variant.
