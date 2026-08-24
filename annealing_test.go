package mayfly

import (
	"math"
	"math/rand"
	"testing"
)

func TestAnnealingSchedulerSchedulesAndReset(t *testing.T) {
	tests := []struct {
		name         string
		scheduleType string
		initial      float64
		coolingRate  float64
		want         float64
	}{
		{
			name:         "default_exponential",
			scheduleType: "",
			initial:      100,
			coolingRate:  0.5,
			want:         50,
		},
		{
			name:         "linear",
			scheduleType: CoolingLinear,
			initial:      10,
			coolingRate:  0.8,
			want:         8,
		},
		{
			name:         "logarithmic",
			scheduleType: "logarithmic",
			initial:      10,
			coolingRate:  0.5,
			want:         10 / (1 + 0.5*math.Log(2)),
		},
		{
			name:         "unknown_defaults_to_exponential",
			scheduleType: "unknown",
			initial:      100,
			coolingRate:  0.5,
			want:         50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheduler := NewAnnealingScheduler(tt.initial, tt.coolingRate, tt.scheduleType)
			scheduler.Update()

			if math.Abs(scheduler.GetTemperature()-tt.want) > epsilon {
				t.Errorf("temperature = %v, want %v", scheduler.GetTemperature(), tt.want)
			}

			if scheduler.Iteration != 1 {
				t.Errorf("iteration = %d, want 1", scheduler.Iteration)
			}

			scheduler.Reset()

			if scheduler.GetTemperature() != tt.initial || scheduler.Iteration != 0 {
				t.Errorf("reset scheduler = %+v, want temperature %v and iteration 0", scheduler, tt.initial)
			}
		})
	}
}

func TestAnnealingSchedulerTemperatureFloors(t *testing.T) {
	linear := NewAnnealingScheduler(0.02, 0.1, CoolingLinear)
	linear.Update()

	if linear.GetTemperature() != 0.01 {
		t.Errorf("linear floor = %v, want 0.01", linear.GetTemperature())
	}

	exponential := NewAnnealingScheduler(1e-10, 0.01, CoolingExponential)
	exponential.Update()

	if exponential.GetTemperature() != 1e-10 {
		t.Errorf("absolute floor = %v, want 1e-10", exponential.GetTemperature())
	}
}

func TestAnnealingNormalizesInvalidDirectInputs(t *testing.T) {
	scheduler := NewAnnealingScheduler(math.NaN(), math.Inf(1), "unknown")
	if scheduler.InitialTemperature != 1 || scheduler.CoolingRate != 0.95 ||
		scheduler.ScheduleType != CoolingExponential {
		t.Fatalf("normalized scheduler = %+v", scheduler)
	}

	if probability := acceptanceProbability(1, math.NaN(), 1); probability != 0 {
		t.Errorf("NaN candidate probability = %v, want 0", probability)
	}

	if probability := acceptanceProbability(1, 2, math.NaN()); probability != 0 {
		t.Errorf("invalid-temperature probability = %v, want 0", probability)
	}
}

func TestAnnealingAcceptance(t *testing.T) {
	if probability := acceptanceProbability(2, 1, 0.1); probability != 1 {
		t.Errorf("improvement probability = %v, want 1", probability)
	}

	worseProbability := acceptanceProbability(1, 2, 1)
	if math.Abs(worseProbability-math.Exp(-1)) > epsilon {
		t.Errorf("worse probability = %v, want %v", worseProbability, math.Exp(-1))
	}

	rng := rand.New(rand.NewSource(1))
	if !shouldAccept(2, 1, 1, rng) {
		t.Fatal("improving candidate was rejected")
	}

	if shouldAccept(1, 2, 1e-10, rng) {
		t.Fatal("effectively impossible worsening candidate was accepted")
	}

	scheduler := NewAnnealingScheduler(1, 0.9, CoolingExponential)
	if !simulatedAnnealingAcceptance(2, 1, scheduler, rng) {
		t.Fatal("scheduler-based acceptance rejected an improvement")
	}
}

func TestAnnealedUpdate(t *testing.T) {
	mayfly := &Mayfly{
		Position: []float64{2},
		Cost:     4,
		Best: Best{
			Position: []float64{2},
			Cost:     4,
		},
	}
	rng := rand.New(rand.NewSource(2))

	accepted, evaluations := annealedUpdate(mayfly, []float64{1}, 1, Sphere, rng)
	if !accepted || evaluations != 1 {
		t.Fatalf("improving update = (%t, %d), want (true, 1)", accepted, evaluations)
	}

	if mayfly.Cost != 1 || mayfly.Position[0] != 1 || mayfly.Best.Cost != 1 || mayfly.Best.Position[0] != 1 {
		t.Errorf("accepted mayfly = %+v, want current and best position 1 with cost 1", mayfly)
	}

	accepted, evaluations = annealedUpdate(mayfly, []float64{10}, 1e-10, Sphere, rng)
	if accepted || evaluations != 1 {
		t.Fatalf("worsening update = (%t, %d), want (false, 1)", accepted, evaluations)
	}

	if mayfly.Cost != 1 || mayfly.Position[0] != 1 {
		t.Errorf("rejected update changed mayfly to %+v", mayfly)
	}
}

func TestAdaptiveTemperatureControl(t *testing.T) {
	scheduler := NewAnnealingScheduler(100, 0.9, CoolingExponential)
	scheduler.CurrentTemperature = 95
	adaptiveTemperatureControl(scheduler, 0.05, 0.1, 0.9)

	if scheduler.CurrentTemperature != 100 {
		t.Errorf("reheated temperature = %v, want capped initial temperature 100", scheduler.CurrentTemperature)
	}

	adaptiveTemperatureControl(scheduler, 0.95, 0.1, 0.9)

	if scheduler.CurrentTemperature != 90 {
		t.Errorf("accelerated cooling temperature = %v, want 90", scheduler.CurrentTemperature)
	}

	adaptiveTemperatureControl(scheduler, 0.5, 0.1, 0.9)

	if scheduler.CurrentTemperature != 90 {
		t.Errorf("in-range acceptance changed temperature to %v, want 90", scheduler.CurrentTemperature)
	}
}
