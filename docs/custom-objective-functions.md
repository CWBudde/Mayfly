# Custom Objective Functions

A Mayfly objective translates one candidate vector into one scalar cost. The
optimizer searches for the smallest cost, so the objective is the model: its
variables, units, bounds, constraints, and failure behavior determine what a
"good" result means. Test that translation independently before spending time
selecting or tuning an algorithm.

## Build a complete objective

This example fits the nonlinear decay model
`baseline + amplitude * exp(-decay * time)` to observations. Its three physical
parameters have different ranges, so the optimizer searches in `[0, 1]^3` and
the objective decodes each candidate before evaluating the model. The source is
also runnable as [`cmd/custom-objective`](../cmd/custom-objective/) and its first
code block is kept in sync by a test.

```go
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
```

Run it from the repository root:

```bash
go run ./cmd/custom-objective
```

The data were generated near amplitude `4.2`, decay `0.65`, and baseline
`0.8`. A seeded run should recover those values with a small mean squared
error. The returned `GlobalBest.Position` is still normalized; decode it before
using or reporting physical parameters. A low objective value establishes a
good fit to this data and loss only. It does not establish parameter
identifiability or out-of-sample accuracy.

`newDecayObjective` validates data once and copies the caller's slice before
creating the closure. The closure then reads immutable observations and has no
shared mutable state. That makes its result deterministic and lets the same
objective work safely with parallel evaluation.

## The objective contract

`mayfly.ObjectiveFunction` has this type:

```go
type ObjectiveFunction func(position []float64) float64
```

For a correctly configured run:

- `position` has exactly `Config.ProblemSize` components.
- Every component is finite and lies between the shared `LowerBound` and
  `UpperBound` when the objective is called.
- The slice is borrowed and read-only. Do not retain it or change its elements.
- Smaller return values are always better. Mayfly implements scalar
  minimization, not multi-objective optimization.
- The result should be deterministic and finite for every valid candidate.
- Calls are sequential by default. With `EnableParallel`, calls may overlap on
  distinct position slices, so all captured state must be safe for concurrent
  reads or access must be synchronized.

Mayfly cannot validate domain meaning. A swapped variable order, mismatched
`ProblemSize`, incorrect unit conversion, or wrong sign can run successfully
while optimizing the wrong problem. Put the decoder and objective under normal
Go unit tests.

## Choose the scalar cost deliberately

For fitting and prediction problems, common losses include mean squared error,
mean absolute error, and negative log likelihood. Use a loss whose units and
outlier behavior match the decision. Dividing a sum by the number of samples
does not change the optimizer's ranking, but it makes costs comparable across
datasets and gives target costs a stable interpretation.

If several goals must be combined, normalize their meaningful scales before
applying weights:

```go
cost := fitError/acceptableFitError + 0.25*energy/energyBudget
```

Document the weights and inspect every component at the returned solution. A
weighted sum encodes one tradeoff; it does not produce a Pareto front, and this
library has no multi-objective optimizer.

To maximize a finite score, negate it:

```go
func objective(position []float64) float64 {
	return -profit(position)
}

bestProfit := -result.GlobalBest.Cost
```

Bounds, convergence targets, logs, and comparisons still use the negated cost.
Do not return negative infinity to represent an excellent solution; both
infinities and `NaN` are invalid.

## Represent heterogeneous variables

`Config.LowerBound` and `Config.UpperBound` apply to every dimension. For
parameters with different ranges or units, search in `[0, 1]` and decode each
component with

`physical = lower + unit*(upper-lower)`.

This also keeps movement coefficients from being dominated merely because one
variable is measured in thousands and another in fractions. Maintain one
authoritative variable order shared by bounds, decoding, constraints, and
result reporting. Validate every lower/upper pair before starting the run.

For integer or categorical decisions, project in the decoder, for example by
rounding an integer and mapping an index to a fixed category list. Then many
continuous positions may represent the same physical choice and the objective
has plateaus or jumps. Always decode the final result before checking domain
rules. If most of the problem is discrete or combinatorial, a continuous
Mayfly search may be a poor modeling choice.

## Keep constraints separate

Use `ConstraintConfig` for actual feasibility rules instead of hiding every
rule in the loss. Inequalities are satisfied when `g(position) <= 0`;
equalities use `h(position) = 0` within `EqualityTolerance`. With normalized
search variables, constraint closures must decode the candidate in exactly the
same way as the objective.

The default feasibility handling ranks feasible candidates ahead of
infeasible ones. Penalty handling is available when the application has a
justified penalty scale. In either case, `GlobalBest.Cost` remains the raw
objective and `GlobalBest.ConstraintViolation` reports feasibility. See the
[configuration guide](api/configuration.md#constraint-handling) for the exact
ranking rules.

A failure to run a simulation or parse a model output is not automatically a
mathematical constraint. Validate static inputs before optimization. For an
unavoidable candidate-specific failure, return a documented, finite worst-case
cost that is large on the objective's scale without risking arithmetic
overflow. Record failures separately if they signal an operational defect.

## Handle non-finite values and panics

Mayfly treats `NaN`, positive infinity, and negative infinity as invalid costs
and ranks them as `math.MaxFloat64`; they never become a best solution. If the
entire initial population is invalid, optimization returns
`mayfly.ErrNoFiniteObjectiveValue`. Sporadic invalid values can therefore hide
a domain bug while merely making some candidates look bad, so test boundary
values and intermediate calculations yourself.

An objective has no error return. Use a constructor, as in the example, to
validate files, observations, dimensions, and fixed coefficients before
calling `Optimize`. Do not deliberately panic for an invalid candidate; a
panic from a normal optimization objective is not an error-reporting channel
and can terminate the run.

Common sources of non-finite results are division by zero, `log` outside its
domain, overflow in `exp`, and subtracting infinities. Prefer algebraically
stable formulas and return a finite domain penalty when the candidate is
genuinely outside the model's usable region.

## Make expensive objectives safe and measurable

Start sequentially. Enable parallel evaluation only after the objective is
correct and a benchmark shows that each call is expensive enough to amortize
coordination overhead:

```go
config.EnableParallel = true
config.MaxWorkers = 4
```

Captured lookup tables and immutable model data are safe. Shared counters,
scratch buffers, random generators, caches, files, and clients need their own
synchronization or one independent instance per call. If randomness is part of
the evaluation, fix it from the candidate or use repeated samples inside one
call; scheduling-dependent noise makes both reproducibility and comparisons
misleading.

Memoization helps only when identical physical candidates recur. Copy values
into a stable key rather than using or retaining the borrowed position slice,
and protect the cache when parallelism is enabled. Limit its size during long
runs. For external simulations, also bound resource use and apply timeouts in
the client: `OptimizeContext` allows in-flight objective calls to finish when a
run is cancelled.

Use `Result.FuncEvalCount` to measure completed objective work. When comparing
variants or configurations, prefer equal objective-evaluation budgets because
an iteration can contain different numbers of calls.

## Test before optimizing

A useful objective test suite covers:

1. A hand-calculated candidate and its expected cost.
2. Every bound endpoint and any singular or projected region.
3. Finite output across a representative random sample of the search box.
4. A known good candidate scoring better than a known bad one.
5. The variable order and normalized-to-physical decoder in both directions.
6. Repeated calls returning exactly the same value when determinism is
   expected.
7. Concurrent calls under `go test -race` before enabling parallel evaluation.
8. Final-result decoding and an independent recomputation of its cost and
   constraints.

Do not assess the objective only through optimizer convergence. A smooth,
decreasing curve can be compelling evidence that the wrong formula is being
optimized efficiently.

## Common mistakes

- Returning a reward directly even though Mayfly minimizes it.
- Setting one shared physical interval when dimensions have different units.
- Mutating or retaining the supplied position slice.
- Reading mutable data that another goroutine or caller can change.
- Returning `NaN` or infinity as a penalty or as a maximization shortcut.
- Mixing hard feasibility, soft preferences, and unrelated metrics into an
  undocumented penalty number.
- Rounding the reported normalized position instead of decoding and projecting
  it through the same function used during evaluation.
- Enabling parallel calls around a non-thread-safe simulator or random source.
- Treating one seeded optimizer result as validation of the model.

## Related documentation

- [Getting Started](getting-started.md) for the smallest complete optimizer.
- [Configuration Guide](api/configuration.md) for bounds, constraints,
  convergence, and parallel evaluation.
- [Parameter-tuning Tutorial](parameter-tuning.md) after the objective is
  tested and frozen.
- [Comparison Framework](api/comparison-framework.md) for paired,
  evaluation-budget-matched experiments.
- [Run Lifecycle](api/run-lifecycle.md) for cancellation and progress
  reporting.
