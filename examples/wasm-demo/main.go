//go:build js && wasm

// Command wasm-demo is the browser front end for github.com/cwbudde/mayfly.
//
// The rule this package exists to enforce: no optimization logic lives in
// JavaScript. Every mayfly position, every cost, every statistic on the two
// demo pages is produced by the library compiled to js/wasm. The JavaScript
// side owns the DOM, the canvas and the animation clock, and nothing else. A
// demo that reimplemented the algorithm in JS would demonstrate the JS, not
// the library.
package main

import "syscall/js"

// exports lists every function the demo publishes, by its name on the
// namespaced globalThis.mayfly object. Each one is wrapped by guard, which is
// the single rule this bridge has: nothing reaches JavaScript without a
// recover() in front of it (see bridge.go).
var exports = map[string]func(js.Value) any{
	"info":      jsInfo,
	"run":       jsRun,
	"landscape": jsLandscape,
	"compare":   jsCompare,
}

// live keeps the js.Func values referenced so they are never released.
var live []js.Func

func main() {
	namespace := js.Global().Get("Object").New()

	for name, fn := range exports {
		wrapped := guard(name, fn)
		live = append(live, wrapped)
		namespace.Set(name, wrapped)
	}

	js.Global().Set("mayfly", namespace)

	// main must not return: the Go runtime tears the instance down when it
	// does, taking every exported function with it. The JavaScript side knows
	// this and never awaits go.run().
	select {}
}
