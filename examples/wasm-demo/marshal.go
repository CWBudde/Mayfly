//go:build js && wasm

package main

import (
	"syscall/js"
	"unsafe"
)

// The obvious way to hand a Go slice to JavaScript from wasm is a SetIndex
// loop. It costs one JS boundary crossing per element, and this demo moves a
// lot of elements: a 300-iteration run of 40 mayflies is 24,000 coordinates,
// and the landscape grid alone is 25,600 samples. One js.CopyBytesToJS per
// array replaces all of them with a single memcpy.
//
// Why the buffer must be JS-owned: Go's wasm heap can grow, and growing the
// WebAssembly.Memory detaches every typed array that was created over its
// buffer. A view over a JS-allocated ArrayBuffer is unaffected, so the page can
// hold its views across calls and reuse them frame after frame.
type float32Sink struct {
	f32      js.Value
	u8       js.Value
	capacity int // in float32 elements
}

// newFloat32Sink allocates a fresh JS-side buffer of n float32 elements.
func newFloat32Sink(n int) float32Sink {
	if n < 0 {
		n = 0
	}

	buffer := js.Global().Get("ArrayBuffer").New(n * 4)

	return float32Sink{
		f32:      js.Global().Get("Float32Array").New(buffer),
		u8:       js.Global().Get("Uint8Array").New(buffer),
		capacity: n,
	}
}

// sinkFor reuses the caller's view pair when it exists and is large enough,
// and otherwise allocates. The page passes opts.out = {males: {f32, u8}, ...}
// so a replaying animation allocates nothing per frame; a first call, or one
// that grew the population, silently gets a new buffer.
func sinkFor(out js.Value, key string, n int) float32Sink {
	if isObject(out) {
		candidate := out.Get(key)
		if isObject(candidate) {
			f32 := candidate.Get("f32")
			u8 := candidate.Get("u8")

			if isObject(f32) && isObject(u8) && f32.Length() >= n {
				return float32Sink{f32: f32, u8: u8, capacity: f32.Length()}
			}
		}
	}

	return newFloat32Sink(n)
}

// write copies data into the sink and returns the JS view to hand back. When
// the sink is larger than the payload (a reused buffer), the returned view is
// a subarray of exactly the right length, so the page never has to track how
// much of the buffer is live.
func (s float32Sink) write(data []float32) js.Value {
	if len(data) > 0 {
		js.CopyBytesToJS(s.u8, float32Bytes(data))
	}

	if s.capacity == len(data) {
		return s.f32
	}

	return s.f32.Call("subarray", 0, len(data))
}

func float32Bytes(data []float32) []byte {
	if len(data) == 0 {
		return nil
	}

	return unsafe.Slice((*byte)(unsafe.Pointer(&data[0])), len(data)*4)
}

// putFloats writes one named float32 array into result, reusing a caller-
// supplied view when one fits.
func putFloats(result map[string]any, out js.Value, key string, data []float32) {
	result[key] = sinkFor(out, key, len(data)).write(data)
}

// toFloat32 narrows a float64 slice for transport. Canvas rendering has no use
// for the extra precision and it halves the bytes moved.
func toFloat32(values []float64) []float32 {
	narrowed := make([]float32, len(values))
	for i, value := range values {
		narrowed[i] = float32(value)
	}

	return narrowed
}

// floatsToJS builds a plain JS array, for the short results (a position vector,
// a ranking) where a typed array would cost more in ceremony than it saves.
func floatsToJS(values []float64) []any {
	items := make([]any, len(values))
	for i, value := range values {
		items[i] = jsNumber(value)
	}

	return items
}
