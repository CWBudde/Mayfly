package mayfly

import (
	"math"
	"math/rand"
	"testing"
)

func TestEOBBMAFemaleEquations(t *testing.T) {
	const seed = 3031

	female := []float64{1, 4}
	male := []float64{5, -5}
	got, mean := eobbmaGaussianFemalePosition(female, male, rand.New(rand.NewSource(seed)))

	replay := rand.New(rand.NewSource(seed))

	for dimension := range female {
		wantMean := (female[dimension] + male[dimension]) / 2

		want := wantMean + replay.NormFloat64()*math.Sqrt(math.Abs(male[dimension]-female[dimension]))
		if mean[dimension] != wantMean || got[dimension] != want {
			t.Errorf("dimension %d = (position=%g, mean=%g), want (%g, %g)",
				dimension, got[dimension], mean[dimension], want, wantMean)
		}
	}
}

func TestEOBBMAMaleGaussianIncludesAdaptiveDisturbance(t *testing.T) {
	const seed = 3032

	pbest := []float64{2}
	gbest := []float64{0}
	peer1 := []float64{-3}
	peer2 := []float64{5}
	got, mean := eobbmaGaussianMalePosition(
		pbest, gbest, peer1, peer2, 4, 1, rand.New(rand.NewSource(seed)),
	)

	replay := rand.New(rand.NewSource(seed))
	delta := replay.Float64() * 8 * math.Exp(-3)
	wantMean := 1.0

	want := wantMean + replay.NormFloat64()*(2+delta)
	if mean[0] != wantMean || got[0] != want {
		t.Fatalf("male Gaussian = (position=%g, mean=%g), want (%g, %g)", got[0], mean[0], want, wantMean)
	}
}

func TestEOBBMAMirrorWallPullback(t *testing.T) {
	got := eobbmaRepairPosition([]float64{12, -14, 3}, []float64{2, -2, 0}, -10, 10)

	want := []float64{
		2 + float64((10-2)*(10-2))/(12-2),
		-2 + float64((-10 - -2)*(-10 - -2))/(-14-(-2)),
		3,
	}
	for dimension := range want {
		if got[dimension] != want[dimension] {
			t.Errorf("dimension %d = %g, want %g", dimension, got[dimension], want[dimension])
		}
	}
}

func TestEOBBMALevyUpdateIsMultiplicative(t *testing.T) {
	const seed = 3033

	position := []float64{2, -3}
	got := eobbmaLevyPosition(position, 1.5, 0.25, rand.New(rand.NewSource(seed)))

	replay := rand.New(rand.NewSource(seed))

	steps := referenceLevyStepsForBareBones(len(position), 1.5, 0.25, replay)
	for dimension := range position {
		want := position[dimension] + position[dimension]*steps[dimension]
		if got[dimension] != want {
			t.Errorf("dimension %d = %g, want %g", dimension, got[dimension], want)
		}
	}
}

func referenceLevyStepsForBareBones(size int, alpha, scale float64, replay *rand.Rand) []float64 {
	sigma := math.Pow(
		math.Gamma(1+alpha)*math.Sin(math.Pi*alpha/2)/
			(math.Gamma((1+alpha)/2)*alpha*math.Pow(2, (alpha-1)/2)),
		1/alpha,
	)

	steps := make([]float64, size)
	for dimension := range steps {
		u := replay.NormFloat64() * sigma

		v := replay.NormFloat64()
		if math.Abs(v) < 1e-300 {
			v = math.Copysign(1e-300, v)
		}

		steps[dimension] = scale * u / math.Pow(math.Abs(v), 1/alpha)
	}

	return steps
}
