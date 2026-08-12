package mayfly

import (
	"math"
	"testing"
)

func TestLogisticMapSequenceAndState(t *testing.T) {
	logisticMap := NewLogisticMap(0.25)
	if logisticMap.Current() != 0.25 {
		t.Errorf("initial state = %v, want 0.25", logisticMap.Current())
	}

	if next := logisticMap.Next(); next != 0.75 {
		t.Errorf("next state = %v, want 0.75", next)
	}

	if logisticMap.Current() != 0.75 {
		t.Errorf("current state = %v, want 0.75", logisticMap.Current())
	}
}

func TestLogisticMapNormalizesInvalidSeeds(t *testing.T) {
	tests := []struct {
		name string
		seed float64
		want float64
	}{
		{name: "zero", seed: 0, want: 0.1},
		{name: "positive_fraction", seed: 1.25, want: 0.3},
		{name: "negative_fraction", seed: -0.5, want: 0.314159},
		{name: "positive_infinity", seed: math.Inf(1), want: 0.271828},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logisticMap := NewLogisticMap(tt.seed)
			if math.Abs(logisticMap.Current()-tt.want) > epsilon {
				t.Errorf("normalized state = %v, want %v", logisticMap.Current(), tt.want)
			}
		})
	}
}

func TestLogisticMapResetAndBoundarySafeguards(t *testing.T) {
	logisticMap := NewLogisticMap(0.25)
	logisticMap.Reset(0.4)

	if logisticMap.Current() != 0.4 {
		t.Errorf("reset state = %v, want 0.4", logisticMap.Current())
	}

	logisticMap.Reset(-0.5)

	if logisticMap.Current() != 0.314159 {
		t.Errorf("normalized reset state = %v, want 0.314159", logisticMap.Current())
	}

	logisticMap.Reset(math.Inf(1))

	if logisticMap.Current() != 0.271828 {
		t.Errorf("infinite reset state = %v, want 0.271828", logisticMap.Current())
	}

	logisticMap.x = 0
	if next := logisticMap.Next(); next != 1e-10 {
		t.Errorf("lower boundary safeguard = %v, want 1e-10", next)
	}

	logisticMap.x = 0.5
	if next := logisticMap.Next(); next != 1-1e-10 {
		t.Errorf("upper boundary safeguard = %v, want %v", next, 1-1e-10)
	}
}
