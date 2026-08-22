//go:build js && wasm

package main

import (
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
	fn       mayfly.ObjectiveFunction
	name     string
	blurb    string
	lower    float64
	upper    float64
	optimum  float64 // objective value at the global minimum
	optimumX float64 // the coordinate every dimension takes there, where it is uniform
	modality string
}

var benchmarks = map[string]benchmark{
	"Sphere": {
		fn: mayfly.Sphere, name: "Sphere", lower: -10, upper: 10,
		optimum: 0, optimumX: 0, modality: "unimodal",
		blurb: "The smooth bowl. One minimum, no structure to get lost in — a sanity check, not a challenge.",
	},
	"Rastrigin": {
		fn: mayfly.Rastrigin, name: "Rastrigin", lower: -5.12, upper: 5.12,
		optimum: 0, optimumX: 0, modality: "highly multimodal",
		blurb: "A cosine egg carton over a bowl. Local minima everywhere; the classic test of whether a swarm escapes them.",
	},
	"Rosenbrock": {
		fn: mayfly.Rosenbrock, name: "Rosenbrock", lower: -5, upper: 10,
		optimum: 0, optimumX: 1, modality: "unimodal valley",
		blurb: "The banana valley. Finding the valley is easy; following its curved floor to (1,1) is not.",
	},
	"Ackley": {
		fn: mayfly.Ackley, name: "Ackley", lower: -32.768, upper: 32.768,
		optimum: 0, optimumX: 0, modality: "multimodal",
		blurb: "A near-flat plain with a narrow central funnel. Punishes swarms that converge before they explore.",
	},
	"Griewank": {
		fn: mayfly.Griewank, name: "Griewank", lower: -600, upper: 600,
		optimum: 0, optimumX: 0, modality: "multimodal",
		blurb: "Product-of-cosines ripple on a wide bowl. Gets easier, not harder, as dimensions rise.",
	},
	"Schwefel": {
		fn: mayfly.Schwefel, name: "Schwefel", lower: -500, upper: 500,
		optimum: 0, optimumX: 420.9687, modality: "deceptive",
		blurb: "Deceptive: the global minimum sits far from the second best, so the gradient actively misleads.",
	},
	"Levy": {
		fn: mayfly.Levy, name: "Levy", lower: -10, upper: 10,
		optimum: 0, optimumX: 1, modality: "multimodal",
		blurb: "Sinusoidal ridges with a single global basin at (1,1).",
	},
	"Zakharov": {
		fn: mayfly.Zakharov, name: "Zakharov", lower: -5, upper: 10,
		optimum: 0, optimumX: 0, modality: "unimodal",
		blurb: "No local minima, but strongly coupled dimensions — a test of coordinated movement.",
	},
	"Michalewicz": {
		fn: mayfly.Michalewicz, name: "Michalewicz", lower: 0, upper: 3.141592653589793,
		optimum: -1.8013, optimumX: 2.20, modality: "multimodal",
		blurb: "Steep valleys separated by flat plateaus. The steepness parameter makes the basins nearly invisible.",
	},
	"DixonPrice": {
		fn: mayfly.DixonPrice, name: "DixonPrice", lower: -10, upper: 10,
		optimum: 0, optimumX: 1, modality: "unimodal valley",
		blurb: "A curved valley whose optimum shifts with the dimension index.",
	},
	"BentCigar": {
		fn: mayfly.BentCigar, name: "BentCigar", lower: -100, upper: 100,
		optimum: 0, optimumX: 0, modality: "ill-conditioned",
		blurb: "One direction is a million times cheaper than the rest. Tests handling of ill-conditioning.",
	},
	"Discus": {
		fn: mayfly.Discus, name: "Discus", lower: -100, upper: 100,
		optimum: 0, optimumX: 0, modality: "ill-conditioned",
		blurb: "BentCigar inverted: one direction dominates the cost entirely.",
	},
	"Weierstrass": {
		fn: mayfly.Weierstrass, name: "Weierstrass", lower: -0.5, upper: 0.5,
		optimum: 0, optimumX: 0, modality: "highly multimodal",
		blurb: "Continuous everywhere, differentiable nowhere. Fractal roughness at every scale.",
	},
	"HappyCat": {
		fn: mayfly.HappyCat, name: "HappyCat", lower: -2, upper: 2,
		optimum: 0, optimumX: -1, modality: "multimodal",
		blurb: "A thin curved shell of near-optimal points around a sphere of radius sqrt(n).",
	},
	"ExpandedSchafferF6": {
		fn: mayfly.ExpandedSchafferF6, name: "ExpandedSchafferF6", lower: -100, upper: 100,
		optimum: 0, optimumX: 0, modality: "highly multimodal",
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
