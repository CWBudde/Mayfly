//go:build js && wasm

package main

import (
	"math"
	"sort"
	"syscall/js"
)

const (
	maxGrid     = 320
	defaultGrid = 160
)

// jsLandscape samples the objective on a grid, for the heatmap the swarm flies
// over.
//
// It calls the same mayfly.ObjectiveFunction the optimizer minimizes, so the
// picture and the search cannot disagree — a heatmap drawn from a JavaScript
// transcription of Rastrigin would be a plausible-looking lie the moment either
// definition drifted.
func jsLandscape(opts js.Value) any {
	benchmarkName := readString(opts, "benchmark", "Rastrigin")

	spec, ok := lookupBenchmark(benchmarkName)
	if !ok {
		return errorResult("landscape: unknown benchmark %q", benchmarkName)
	}

	var (
		width      = clampInt(readInt(opts, "width", defaultGrid), 8, maxGrid)
		height     = clampInt(readInt(opts, "height", defaultGrid), 8, maxGrid)
		dimensions = clampInt(readInt(opts, "dimensions", 2), 2, maxDimensions)
		lower      = readFloat(opts, "lower", spec.lower)
		upper      = readFloat(opts, "upper", spec.upper)
		axisX      = clampInt(readInt(opts, "axisX", 0), 0, dimensions-1)
		axisY      = clampInt(readInt(opts, "axisY", 1), 0, dimensions-1)
		mode       = readString(opts, "mode", "rank")
	)

	if lower >= upper {
		return errorResult("landscape: lower bound %v must be below upper bound %v", lower, upper)
	}

	// Above two dimensions the grid is a slice, not the whole function. The
	// unplotted axes are pinned at the known minimizer so the slice passes
	// through the global minimum and the picture stays meaningful.
	//
	// That minimizer is a vector, not a scalar: Dixon-Price's coordinates
	// depend on their index, so pinning them all to 1 produced a slice through
	// a point costing 54 in ten dimensions rather than through the optimum.
	// Michalewicz has no known minimizer above two dimensions at all; there the
	// slice is taken through the middle of the domain and reported as
	// arbitrary, because claiming it runs through an optimum nobody can name
	// would be the more misleading of the two options.
	position, throughOptimum := spec.optimumAt(dimensions)
	if !throughOptimum {
		position = make([]float64, dimensions)
		for i := range position {
			position[i] = (lower + upper) / 2
		}
	}

	samples := make([]float32, 0, width*height)
	minimum := math.Inf(1)
	maximum := math.Inf(-1)

	for row := range height {
		// Row 0 is the top of the canvas, which is the *high* y value.
		y := upper - (float64(row)+0.5)/float64(height)*(upper-lower)

		for column := range width {
			x := lower + (float64(column)+0.5)/float64(width)*(upper-lower)

			position[axisX] = x
			position[axisY] = y

			value := spec.fn(position)
			if math.IsNaN(value) {
				value = math.Inf(1)
			}

			if !math.IsInf(value, 0) {
				minimum = math.Min(minimum, value)
				maximum = math.Max(maximum, value)
			}

			samples = append(samples, float32(value))
		}
	}

	if math.IsInf(minimum, 0) || math.IsInf(maximum, 0) {
		return errorResult("landscape: %s produced no finite samples", spec.name)
	}

	normalized := normalize(samples, minimum, maximum, mode)

	out := opts.Get("out")

	response := map[string]any{
		"benchmark":  spec.name,
		"width":      width,
		"height":     height,
		"lower":      lower,
		"upper":      upper,
		"min":        jsNumber(minimum),
		"max":        jsNumber(maximum),
		"mode":       mode,
		"projected":  dimensions > 2,
		"axisX":      axisX,
		"axisY":      axisY,
		"dimensions": dimensions,

		// Whether the slice actually passes through the global minimum. The
		// page says which kind of picture it is showing.
		"throughOptimum": throughOptimum,
		"optimum":        optionalNumber(spec.optimumValue(dimensions)),
	}

	putFloats(response, out, "values", normalized)
	putFloats(response, out, "raw", samples)

	return response
}

// normalize maps raw objective samples onto [0,1] for the colour ramp.
//
// "rank" is the default, and it is the only mode that reliably shows structure.
// Benchmark landscapes have wildly different value distributions: Rastrigin
// spends most of its domain in a narrow band near the mean, so both a linear
// and a log ramp paint almost the whole picture one colour and the basins the
// swarm is actually hunting disappear. Ranking each sample against the others
// spreads the ramp evenly by construction, whatever the distribution.
//
// The trade is real and the page says so: a rank-normalized map shows where the
// low ground is, not how deep it is. The "log" and "linear" modes preserve
// magnitude for anyone who wants to compare depths instead.
func normalize(samples []float32, minimum, maximum float64, mode string) []float32 {
	normalized := make([]float32, len(samples))
	span := maximum - minimum

	switch mode {
	case "linear", "log":
		for i, value := range samples {
			var t float64

			switch {
			case span <= 0:
				t = 0
			case mode == "log":
				t = math.Log1p(float64(value)-minimum) / math.Log1p(span)
			default:
				t = (float64(value) - minimum) / span
			}

			if math.IsNaN(t) {
				t = 1
			}

			normalized[i] = float32(math.Max(0, math.Min(1, t)))
		}

	default:
		order := make([]int, len(samples))
		for i := range order {
			order[i] = i
		}

		sort.Slice(order, func(a, b int) bool {
			return samples[order[a]] < samples[order[b]]
		})

		last := float64(len(order) - 1)
		if last <= 0 {
			last = 1
		}

		for rank, index := range order {
			normalized[index] = float32(float64(rank) / last)
		}
	}

	return normalized
}
