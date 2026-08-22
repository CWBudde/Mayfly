//go:build js && wasm

package main

import (
	"fmt"
	"math"
	"syscall/js"
)

// guard wraps a demo entry point so that no failure inside Go can ever reach
// the JavaScript side as a trap.
//
// This matters more under js/wasm than it would anywhere else: a Go panic that
// unwinds out of a js.Func aborts the whole wasm instance. Every subsequent
// call into the module then fails, so a single bad request permanently bricks
// the page until the user reloads. Returning the failure as data costs one
// deferred recover per call and keeps a mistyped option from taking the demo
// down with it.
func guard(name string, fn func(js.Value) any) js.Func {
	return js.FuncOf(func(_ js.Value, args []js.Value) (result any) {
		defer func() {
			if r := recover(); r != nil {
				result = js.ValueOf(map[string]any{
					"error": fmt.Sprintf("%s: %v", name, r),
					"panic": true,
				})
			}
		}()

		opts := js.Undefined()
		if len(args) > 0 {
			opts = args[0]
		}

		return fn(opts)
	})
}

// errorResult is the shape every export returns on a rejected request. It
// carries panic:false to distinguish a request this code refused from one that
// crashed it — the page reports the two differently.
func errorResult(format string, args ...any) map[string]any {
	return map[string]any{
		"error": fmt.Sprintf(format, args...),
		"panic": false,
	}
}

func isObject(value js.Value) bool {
	return value.Type() == js.TypeObject && !value.IsNull()
}

// The read* helpers are deliberately tolerant: a missing key, a null, or a
// value of the wrong type yields the fallback rather than an error. The page
// sends partial option objects all the time (a control the user has not
// touched yet has nothing to send), and treating that as a failure would mean
// every caller had to fill in defaults the Go side already knows.

func readInt(opts js.Value, key string, fallback int) int {
	if !isObject(opts) {
		return fallback
	}

	value := opts.Get(key)
	if value.Type() != js.TypeNumber {
		return fallback
	}

	number := value.Float()
	if math.IsNaN(number) || math.IsInf(number, 0) {
		return fallback
	}

	return int(number)
}

func readFloat(opts js.Value, key string, fallback float64) float64 {
	if !isObject(opts) {
		return fallback
	}

	value := opts.Get(key)
	if value.Type() != js.TypeNumber {
		return fallback
	}

	number := value.Float()
	if math.IsNaN(number) || math.IsInf(number, 0) {
		return fallback
	}

	return number
}

func readString(opts js.Value, key, fallback string) string {
	if !isObject(opts) {
		return fallback
	}

	value := opts.Get(key)
	if value.Type() != js.TypeString {
		return fallback
	}

	return value.String()
}

func readBool(opts js.Value, key string, fallback bool) bool {
	if !isObject(opts) {
		return fallback
	}

	value := opts.Get(key)
	if value.Type() != js.TypeBoolean {
		return fallback
	}

	return value.Bool()
}

// readStrings reads an array of strings, skipping anything that is not one.
func readStrings(opts js.Value, key string, fallback []string) []string {
	if !isObject(opts) {
		return fallback
	}

	value := opts.Get(key)
	if !isObject(value) || value.Length() == 0 {
		return fallback
	}

	items := make([]string, 0, value.Length())

	for i := range value.Length() {
		item := value.Index(i)
		if item.Type() == js.TypeString {
			items = append(items, item.String())
		}
	}

	if len(items) == 0 {
		return fallback
	}

	return items
}

func clampInt(value, low, high int) int {
	if value < low {
		return low
	}

	if value > high {
		return high
	}

	return value
}

// optionalNumber renders a value that may not be known for the requested
// dimension — Michalewicz's optimum outside the tabulated 2, 5 and 10 — as null
// rather than as a plausible-looking number the page would then display.
func optionalNumber(value float64, known bool) any {
	if !known {
		return nil
	}

	return jsNumber(value)
}

// jsNumber renders a float for JavaScript. NaN and ±Inf are not representable
// in JSON and arrive in JS as unusable values, so they become null and the
// page renders them as "—" rather than "NaN".
func jsNumber(value float64) any {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return nil
	}

	return value
}
