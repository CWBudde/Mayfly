package mayfly

// LogisticMap implements the logistic chaotic map.
// The logistic map is defined by: x_{n+1} = r * x_n * (1 - x_n)
// where r is the control parameter. When r = 4.0, the map exhibits
// fully chaotic behavior with good ergodic properties.
//
// The logistic map is widely used in chaos-based optimization algorithms
// due to its simplicity, well-understood dynamics, and ability to generate
// pseudo-random sequences with better uniformity than standard PRNGs.
type LogisticMap struct {
	x float64 // Current state value in [0, 1]
	r float64 // Control parameter (typically 4.0 for full chaos)
}

// NewLogisticMap creates a new logistic map with the given seed.
// The seed should be in the range (0, 1), exclusive of boundaries.
// If seed is outside this range, it will be normalized to (0, 1).
// The control parameter r is set to 4.0 for fully chaotic behavior.
func NewLogisticMap(seed float64) *LogisticMap {
	// Ensure seed is in valid range (0, 1)
	if seed <= 0.0 || seed >= 1.0 {
		// Normalize to (0, 1) using simple hash-like function
		seed = 0.1 + 0.8*(seed-float64(int(seed)))
		if seed <= 0.0 {
			seed = 0.314159 // Safe default
		}

		if seed >= 1.0 {
			seed = 0.271828 // Safe default
		}
	}

	return &LogisticMap{
		x: seed,
		r: 4.0, // Standard value for fully chaotic behavior
	}
}

// Next generates and returns the next value in the chaotic sequence.
// The returned value is in the range [0, 1].
// This method updates the internal state and should be called
// sequentially to generate a chaotic sequence.
func (lm *LogisticMap) Next() float64 {
	// Apply logistic map equation: x_{n+1} = r * x_n * (1 - x_n)
	lm.x = lm.r * lm.x * (1.0 - lm.x)

	// Safeguard against numerical drift to boundaries
	// The logistic map should naturally stay in (0,1) but floating point
	// errors can occasionally push values to exactly 0 or 1, which would
	// cause the sequence to collapse to a fixed point.
	if lm.x <= 0.0 {
		lm.x = 1e-10
	}

	if lm.x >= 1.0 {
		lm.x = 1.0 - 1e-10
	}

	return lm.x
}

// Current returns the current state value without advancing the sequence.
// This is useful for debugging or when you need to inspect the state
// without modifying it.
func (lm *LogisticMap) Current() float64 {
	return lm.x
}

// Reset resets the map to a new seed value.
// This allows reusing the same LogisticMap instance with a different
// starting point.
func (lm *LogisticMap) Reset(seed float64) {
	// Ensure seed is in valid range (0, 1)
	if seed <= 0.0 || seed >= 1.0 {
		seed = 0.1 + 0.8*(seed-float64(int(seed)))
		if seed <= 0.0 {
			seed = 0.314159
		}

		if seed >= 1.0 {
			seed = 0.271828
		}
	}

	lm.x = seed
}
