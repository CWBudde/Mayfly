//go:build js && wasm

package main

import (
	"context"
	"math"
	"syscall/js"

	"github.com/cwbudde/mayfly"
)

const (
	maxDimensions = 30
	maxIterations = 1000
	maxPopulation = 200
)

// jsRun performs one optimization and returns the whole recorded history.
//
// Record-then-replay, rather than a stepped state machine driven from JS, is a
// deliberate choice and it is what keeps this page simple. At the sizes the
// Swarm Lab uses — two plotted dimensions, tens of mayflies, at most a thousand
// iterations — a complete run finishes in a handful of milliseconds, so there
// is nothing to chunk, nothing to cancel, and no partial state to reconcile.
// The page then replays the history on its own clock, which is what makes the
// scrubber possible: you cannot scrub backwards through a live computation.
//
// It also makes reproducibility demonstrable. The same seed produces the same
// history, byte for byte, and the page proves it by re-running.
func jsRun(opts js.Value) any {
	benchmarkName := readString(opts, "benchmark", "Rastrigin")

	spec, ok := lookupBenchmark(benchmarkName)
	if !ok {
		return errorResult("run: unknown benchmark %q", benchmarkName)
	}

	var (
		variantName = readString(opts, "variant", "MA")
		dimensions  = clampInt(readInt(opts, "dimensions", 2), 2, maxDimensions)
		iterations  = clampInt(readInt(opts, "iterations", 200), 1, maxIterations)
		npop        = clampInt(readInt(opts, "npop", 20), 2, maxPopulation)
		npopf       = clampInt(readInt(opts, "npopf", 20), 2, maxPopulation)
		seed        = int64(readFloat(opts, "seed", 42))
		lower       = readFloat(opts, "lower", spec.lower)
		upper       = readFloat(opts, "upper", spec.upper)
		axisX       = clampInt(readInt(opts, "axisX", 0), 0, dimensions-1)
		axisY       = clampInt(readInt(opts, "axisY", 1), 0, dimensions-1)
	)

	if lower >= upper {
		return errorResult("run: lower bound %v must be below upper bound %v", lower, upper)
	}

	config, err := configFor(variantName, benchmarkName, dimensions, iterations, npop, npopf, seed, lower, upper)
	if err != nil {
		return errorResult("run: %v", err)
	}

	history := newHistory(iterations, npop, npopf, axisX, axisY)

	result, err := mayfly.OptimizeContext(
		context.Background(),
		config,
		mayfly.WithPopulationObserver(history.observe),
	)
	if err != nil {
		return errorResult("run: %v", err)
	}

	out := opts.Get("out")

	response := map[string]any{
		"benchmark":         spec.name,
		"variant":           variantName,
		"dimensions":        dimensions,
		"axisX":             axisX,
		"axisY":             axisY,
		"npop":              npop,
		"npopf":             npopf,
		"iterations":        result.IterationCount,
		"evaluations":       result.FuncEvalCount,
		"terminationReason": string(result.TerminationReason),
		"bestCost":          jsNumber(result.GlobalBest.Cost),
		"bestPosition":      floatsToJS(result.GlobalBest.Position),
		"optimum":           optionalNumber(spec.optimumValue(dimensions)),
		"lower":             lower,
		"upper":             upper,

		// Result.Seed holds a time.Now() value that was never used once a
		// caller supplies its own *rand.Rand, so reporting it would be a lie.
		// This is the seed the run actually used.
		"seed": float64(seed),
	}

	history.truncate(result.IterationCount)

	putFloats(response, out, "convergence", toFloat32(result.ConvergenceCurve))
	putFloats(response, out, "males", history.males)
	putFloats(response, out, "females", history.females)
	putFloats(response, out, "maleCost", history.maleCost)
	putFloats(response, out, "femaleCost", history.femaleCost)
	putFloats(response, out, "bestTrail", history.bestTrail)
	putFloats(response, out, "maleDiversity", history.maleDiversity)
	putFloats(response, out, "femaleDiversity", history.femaleDiversity)

	return response
}

// history accumulates one run's population snapshots into the flat arrays the
// canvas wants. Everything is appended in iteration order, so an array of
// coordinates is iterations x population x 2 and the page indexes into it
// arithmetically rather than walking a nested structure across the boundary.
type history struct {
	males           []float32
	females         []float32
	maleCost        []float32
	femaleCost      []float32
	bestTrail       []float32
	maleDiversity   []float32
	femaleDiversity []float32
	npop            int
	npopf           int
	axisX           int
	axisY           int
}

func newHistory(iterations, npop, npopf, axisX, axisY int) *history {
	return &history{
		males:           make([]float32, 0, iterations*npop*2),
		females:         make([]float32, 0, iterations*npopf*2),
		maleCost:        make([]float32, 0, iterations*npop),
		femaleCost:      make([]float32, 0, iterations*npopf),
		bestTrail:       make([]float32, 0, iterations*2),
		maleDiversity:   make([]float32, 0, iterations),
		femaleDiversity: make([]float32, 0, iterations),
		npop:            npop,
		npopf:           npopf,
		axisX:           axisX,
		axisY:           axisY,
	}
}

func (h *history) observe(snapshot mayfly.PopulationSnapshot) {
	h.males, h.maleCost = h.appendPopulation(h.males, h.maleCost, snapshot.Males)
	h.females, h.femaleCost = h.appendPopulation(h.females, h.femaleCost, snapshot.Females)

	h.bestTrail = append(h.bestTrail,
		float32(coordinate(snapshot.Best.Position, h.axisX)),
		float32(coordinate(snapshot.Best.Position, h.axisY)),
	)

	h.maleDiversity = append(h.maleDiversity, float32(diversity(snapshot.Males)))
	h.femaleDiversity = append(h.femaleDiversity, float32(diversity(snapshot.Females)))
}

func (h *history) appendPopulation(
	positions, costs []float32,
	population []mayfly.Mayfly,
) ([]float32, []float32) {
	for i := range population {
		positions = append(positions,
			float32(coordinate(population[i].Position, h.axisX)),
			float32(coordinate(population[i].Position, h.axisY)),
		)
		costs = append(costs, float32(population[i].Cost))
	}

	return positions, costs
}

// truncate discards any tail the run never reached. Early stopping (a target
// cost or a stagnation window) ends the loop before MaxIterations, and the
// page's frame arithmetic assumes every array covers exactly the same span.
func (h *history) truncate(iterations int) {
	h.males = clip(h.males, iterations*h.npop*2)
	h.females = clip(h.females, iterations*h.npopf*2)
	h.maleCost = clip(h.maleCost, iterations*h.npop)
	h.femaleCost = clip(h.femaleCost, iterations*h.npopf)
	h.bestTrail = clip(h.bestTrail, iterations*2)
	h.maleDiversity = clip(h.maleDiversity, iterations)
	h.femaleDiversity = clip(h.femaleDiversity, iterations)
}

func clip(values []float32, length int) []float32 {
	if length < 0 || length > len(values) {
		return values
	}

	return values[:length]
}

func coordinate(position []float64, axis int) float64 {
	if axis < 0 || axis >= len(position) {
		return 0
	}

	return position[axis]
}

// diversity is the mean Euclidean distance from each member to the population
// centroid, over all dimensions rather than the two plotted ones. It is the
// demo's clearest single number for the exploration-to-exploitation shift: it
// starts high, and a converging swarm drives it to zero.
func diversity(population []mayfly.Mayfly) float64 {
	if len(population) == 0 || len(population[0].Position) == 0 {
		return 0
	}

	dimensions := len(population[0].Position)
	centroid := make([]float64, dimensions)

	for i := range population {
		for d := 0; d < dimensions && d < len(population[i].Position); d++ {
			centroid[d] += population[i].Position[d]
		}
	}

	for d := range centroid {
		centroid[d] /= float64(len(population))
	}

	total := 0.0

	for i := range population {
		sum := 0.0

		for d := 0; d < dimensions && d < len(population[i].Position); d++ {
			delta := population[i].Position[d] - centroid[d]
			sum += delta * delta
		}

		total += math.Sqrt(sum)
	}

	return total / float64(len(population))
}
