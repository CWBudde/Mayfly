//go:build js && wasm

package main

import (
	"math"
	"sort"

	"github.com/cwbudde/mayfly"
)

// benchmark pairs one of the library's objective functions with the metadata a
// UI needs: where to search, where the answer is, and what makes the function
// worth showing.
//
// The library exposes its 15 benchmarks as plain functions with no registry and
// no programmatic bounds — those live only in doc comments, and inconsistently.
// This table is the demo's own reading of them, kept here rather than pushed
// into the library because the demo is currently the only caller that needs it.
type benchmark struct {
	fn   mayfly.ObjectiveFunction
	name string

	// optimumAt returns the minimizing position in the requested dimension,
	// and whether one is known. It is not a scalar because two of these
	// functions do not have a uniform minimizer: Dixon-Price's coordinates
	// depend on their index, and Michalewicz has no closed form at all. Pinning
	// every hidden coordinate to a single value produced heatmap slices that
	// missed the global minimum while the UI claimed they passed through it.
	optimumAt func(dimensions int) ([]float64, bool)

	// optimumValue returns the objective value at that minimum, and whether it
	// is known for this dimension. Michalewicz's is tabulated for a few
	// dimensions and unknown for the rest.
	optimumValue func(dimensions int) (float64, bool)

	blurb    string
	modality string
	lower    float64
	upper    float64
}

// uniformOptimum describes a function minimized where every coordinate takes
// the same value, which is all of them except Dixon-Price and Michalewicz.
func uniformOptimum(coordinate float64) func(int) ([]float64, bool) {
	return func(dimensions int) ([]float64, bool) {
		position := make([]float64, dimensions)
		for i := range position {
			position[i] = coordinate
		}

		return position, true
	}
}

// knownValue describes a function whose minimum is the same in every dimension.
func knownValue(value float64) func(int) (float64, bool) {
	return func(int) (float64, bool) { return value, true }
}

// dixonPriceOptimum is minimized at x_i = 2^(-(2^i - 2) / 2^i) for one-based i,
// so only the first coordinate is 1. Evaluating the all-ones vector instead
// gives 2, 14 and 54 in 2, 5 and 10 dimensions rather than 0.
func dixonPriceOptimum(dimensions int) ([]float64, bool) {
	position := make([]float64, dimensions)

	for i := range position {
		power := math.Pow(2, float64(i+1))
		position[i] = math.Pow(2, -(power-2)/power)
	}

	return position, true
}

// michalewiczOptima are the published minima for the steepness m = 10 that
// functions.go implements. There is no closed form, and no tabulated value for
// the dimensions in between, so those are reported as unknown rather than
// guessed at.
var michalewiczOptima = map[int]float64{2: -1.8013, 5: -4.687658, 10: -9.66015}

func michalewiczValue(dimensions int) (float64, bool) {
	value, ok := michalewiczOptima[dimensions]

	return value, ok
}

// michalewiczOptimumAt knows the minimizer only in two dimensions, where it is
// approximately (2.20, 1.57). Above that the demo has no position to pin a
// projection to and says so instead of inventing one.
func michalewiczOptimumAt(dimensions int) ([]float64, bool) {
	if dimensions != 2 {
		return nil, false
	}

	return []float64{2.20, 1.57}, true
}

var benchmarks = map[string]benchmark{
	"Sphere": {
		fn: mayfly.Sphere, name: "Sphere", lower: -10, upper: 10,
		optimumAt: uniformOptimum(0), optimumValue: knownValue(0), modality: "unimodal",
		blurb: "The smooth bowl. One minimum, no structure to get lost in — a sanity check, not a challenge.",
	},
	"Rastrigin": {
		fn: mayfly.Rastrigin, name: "Rastrigin", lower: -5.12, upper: 5.12,
		optimumAt: uniformOptimum(0), optimumValue: knownValue(0), modality: "highly multimodal",
		blurb: "A cosine egg carton over a bowl. Local minima everywhere; the classic test of whether a swarm escapes them.",
	},
	"Rosenbrock": {
		fn: mayfly.Rosenbrock, name: "Rosenbrock", lower: -5, upper: 10,
		optimumAt: uniformOptimum(1), optimumValue: knownValue(0), modality: "unimodal valley",
		blurb: "The banana valley. Finding the valley is easy; following its curved floor to (1,1) is not.",
	},
	"Ackley": {
		fn: mayfly.Ackley, name: "Ackley", lower: -32.768, upper: 32.768,
		optimumAt: uniformOptimum(0), optimumValue: knownValue(0), modality: "multimodal",
		blurb: "A near-flat plain with a narrow central funnel. Punishes swarms that converge before they explore.",
	},
	"Griewank": {
		fn: mayfly.Griewank, name: "Griewank", lower: -600, upper: 600,
		optimumAt: uniformOptimum(0), optimumValue: knownValue(0), modality: "multimodal",
		blurb: "Product-of-cosines ripple on a wide bowl. Gets easier, not harder, as dimensions rise.",
	},
	"Schwefel": {
		fn: mayfly.Schwefel, name: "Schwefel", lower: -500, upper: 500,
		optimumAt: uniformOptimum(420.9687), optimumValue: knownValue(0), modality: "deceptive",
		blurb: "Deceptive: the global minimum sits far from the second best, so the gradient actively misleads.",
	},
	"Levy": {
		fn: mayfly.Levy, name: "Levy", lower: -10, upper: 10,
		optimumAt: uniformOptimum(1), optimumValue: knownValue(0), modality: "multimodal",
		blurb: "Sinusoidal ridges with a single global basin at (1,1).",
	},
	"Zakharov": {
		fn: mayfly.Zakharov, name: "Zakharov", lower: -5, upper: 10,
		optimumAt: uniformOptimum(0), optimumValue: knownValue(0), modality: "unimodal",
		blurb: "No local minima, but strongly coupled dimensions — a test of coordinated movement.",
	},
	"Michalewicz": {
		fn: mayfly.Michalewicz, name: "Michalewicz", lower: 0, upper: 3.141592653589793,
		optimumAt: michalewiczOptimumAt, optimumValue: michalewiczValue, modality: "multimodal",
		blurb: "Steep valleys separated by flat plateaus. The steepness parameter makes the basins nearly invisible.",
	},
	"DixonPrice": {
		fn: mayfly.DixonPrice, name: "DixonPrice", lower: -10, upper: 10,
		optimumAt: dixonPriceOptimum, optimumValue: knownValue(0), modality: "unimodal valley",
		blurb: "A curved valley whose optimum shifts with the dimension index.",
	},
	"BentCigar": {
		fn: mayfly.BentCigar, name: "BentCigar", lower: -100, upper: 100,
		optimumAt: uniformOptimum(0), optimumValue: knownValue(0), modality: "ill-conditioned",
		blurb: "One direction is a million times cheaper than the rest. Tests handling of ill-conditioning.",
	},
	"Discus": {
		fn: mayfly.Discus, name: "Discus", lower: -100, upper: 100,
		optimumAt: uniformOptimum(0), optimumValue: knownValue(0), modality: "ill-conditioned",
		blurb: "BentCigar inverted: one direction dominates the cost entirely.",
	},
	"Weierstrass": {
		fn: mayfly.Weierstrass, name: "Weierstrass", lower: -0.5, upper: 0.5,
		optimumAt: uniformOptimum(0), optimumValue: knownValue(0), modality: "highly multimodal",
		blurb: "Continuous everywhere, differentiable nowhere. Fractal roughness at every scale.",
	},
	"HappyCat": {
		fn: mayfly.HappyCat, name: "HappyCat", lower: -2, upper: 2,
		optimumAt: uniformOptimum(-1), optimumValue: knownValue(0), modality: "multimodal",
		blurb: "A thin curved shell of near-optimal points around a sphere of radius sqrt(n).",
	},
	"ExpandedSchafferF6": {
		fn: mayfly.ExpandedSchafferF6, name: "ExpandedSchafferF6", lower: -100, upper: 100,
		optimumAt: uniformOptimum(0), optimumValue: knownValue(0), modality: "highly multimodal",
		blurb: "Concentric ripples around the origin — every ring is a local minimum.",
	},
}

// benchmarkNames returns the table's keys in a stable, didactic order: the
// five classics first, in rising difficulty, then the CEC-style additions.
// Map iteration order would reshuffle the UI's dropdown on every page load.
func benchmarkNames() []string {
	ordered := []string{
		"Sphere", "Rastrigin", "Rosenbrock", "Ackley", "Griewank",
		"Schwefel", "Levy", "Zakharov", "Michalewicz", "DixonPrice",
		"BentCigar", "Discus", "Weierstrass", "HappyCat", "ExpandedSchafferF6",
	}

	seen := make(map[string]bool, len(ordered))
	names := make([]string, 0, len(benchmarks))

	for _, name := range ordered {
		if _, ok := benchmarks[name]; ok {
			names = append(names, name)
			seen[name] = true
		}
	}

	// Anything added to the table but forgotten in the list above still shows
	// up, sorted, rather than silently vanishing from the UI.
	rest := make([]string, 0)

	for name := range benchmarks {
		if !seen[name] {
			rest = append(rest, name)
		}
	}

	sort.Strings(rest)

	return append(names, rest...)
}

func lookupBenchmark(name string) (benchmark, bool) {
	found, ok := benchmarks[name]

	return found, ok
}

// successTarget derives the cost that counts as solving this function in this
// dimension, and whether such a target can be expressed at all.
//
// The library treats a non-positive target as "no target set", so a function
// whose optimum is negative has no representable success threshold. Reporting
// that honestly is the only correct option: the alternative, a fixed 1e-8,
// silently scored every Michalewicz run as a success because every negative
// cost is below it.
func successTarget(spec benchmark, dimensions int) (float64, bool) {
	optimum, known := spec.optimumValue(dimensions)
	if !known {
		return 0, false
	}

	const tolerance = 1e-8

	target := optimum + tolerance
	if target <= 0 {
		return 0, false
	}

	return target, true
}
