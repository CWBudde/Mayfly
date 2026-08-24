//go:build js && wasm

package main

import (
	"runtime"
	"syscall/js"
)

// jsInfo is the capability table the page builds its controls from.
//
// Everything in it is derived from the library's own registries rather than
// restated here: the variant list, names, descriptions and overheads come from
// GetAllVariants(), and the benchmark metadata from the demo's table. The
// dropdowns in the static HTML are placeholders that the page replaces as soon
// as this call returns, so adding a variant to the library puts it in the UI
// without anyone editing the markup.
func jsInfo(_ js.Value) any {
	variants := variantOrder()
	variantList := make([]any, 0, len(variants))

	for _, variant := range variants {
		variantList = append(variantList, map[string]any{
			"key":            variantKey(variant),
			"name":           variant.Name(),
			"fullName":       variant.FullName(),
			"description":    variant.Description(),
			"recommendedFor": stringsToJS(variant.RecommendedFor()),
			"overhead":       jsNumber(variant.EstimatedOverhead()),
		})
	}

	names := benchmarkNames()
	benchmarkList := make([]any, 0, len(names))

	for _, name := range names {
		spec := benchmarks[name]
		benchmarkList = append(benchmarkList, map[string]any{
			"name":     spec.name,
			"blurb":    spec.blurb,
			"modality": spec.modality,
			"lower":    jsNumber(spec.lower),
			"upper":    jsNumber(spec.upper),

			// Reported per dimension, because Michalewicz's minimum is only
			// tabulated for a few and the page must not print a 2-D reference
			// value next to a 10-D run.
			"optimum2d": optionalNumber(spec.optimumValue(2)),
		})
	}

	qmcList := make([]any, 0, len(qmcInitOptions()))

	for _, option := range qmcInitOptions() {
		qmcList = append(qmcList, map[string]any{
			"key":   option.key,
			"label": option.label,
			"note":  option.note,
		})
	}

	return map[string]any{
		"goVersion":     runtime.Version(),
		"qmcInit":       qmcList,
		"goos":          runtime.GOOS,
		"goarch":        runtime.GOARCH,
		"variants":      variantList,
		"benchmarks":    benchmarkList,
		"maxDimensions": maxDimensions,
		"maxIterations": maxIterations,
		"maxPopulation": maxPopulation,
		"maxGrid":       maxGrid,
		"maxRuns":       maxCompareRuns,
	}
}
