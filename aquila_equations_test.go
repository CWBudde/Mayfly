package mayfly

import (
	"math"
	"math/rand"
	"testing"
)

func referenceAquilaLevy(size int, replay *rand.Rand) []float64 {
	const alpha = 1.5
	sigma := math.Pow(
		math.Gamma(1+alpha)*math.Sin(math.Pi*alpha/2)/
			(math.Gamma((1+alpha)/2)*alpha*math.Pow(2, (alpha-1)/2)),
		1/alpha,
	)
	result := make([]float64, size)
	for i := range result {
		u := replay.NormFloat64() * sigma
		v := replay.NormFloat64()
		if math.Abs(v) < 1e-300 {
			v = math.Copysign(1e-300, v)
		}
		result[i] = 0.01 * u / math.Pow(math.Abs(v), 1/alpha)
	}

	return result
}

func TestAquilaX1UsesOneEquationScalar(t *testing.T) {
	const seed = 711
	current := []float64{8, 9}
	best := []float64{2, -4}
	mean := []float64{1, 3}
	got := aquilaExpandedExploration(
		current, best, mean, 0, 10, -100, 100, rand.New(rand.NewSource(seed)),
	)

	replay := rand.New(rand.NewSource(seed))
	r := replay.Float64()
	for dimension := range current {
		want := best[dimension]*0.9 + mean[dimension] - best[dimension]*r
		if math.Abs(got[dimension]-want) > 1e-15 {
			t.Errorf("X1 dimension %d = %.17g, want %.17g", dimension, got[dimension], want)
		}
	}
}

func TestAquilaX2MatchesPublishedEquation(t *testing.T) {
	const seed = 812
	current := []float64{3, 4, 5}
	best := []float64{0.5, -1, 2}
	population := []*Mayfly{
		{Position: []float64{1, 2, 3}},
		{Position: []float64{-2, 0.25, 4}},
	}

	got := aquilaNarrowedExploration(
		current, best, population, -1e6, 1e6, rand.New(rand.NewSource(seed)),
	)

	replay := rand.New(rand.NewSource(seed))
	levy := referenceAquilaLevy(len(current), replay)
	xr := population[replay.Intn(len(population))].Position
	randomScale := replay.Float64()
	for dimension := range current {
		d := float64(dimension + 1)
		radius := 10.0 + 0.00565*d
		theta := -0.005*d + 3*math.Pi/2
		x := radius * math.Sin(theta)
		y := radius * math.Cos(theta)
		want := best[dimension]*levy[dimension] + xr[dimension] + (y-x)*randomScale
		if got[dimension] != want {
			t.Errorf("X2 dimension %d = %.17g, want %.17g", dimension, got[dimension], want)
		}
	}
}

func TestAquilaX3UsesPublishedConstantsAndBroadcastRandoms(t *testing.T) {
	const seed = 913
	current := []float64{1, 2}
	best := []float64{4, -3}
	mean := []float64{1, 1}
	got := aquilaExpandedExploitation(
		current, best, mean, 79, 100, -10, 10, rand.New(rand.NewSource(seed)),
	)

	replay := rand.New(rand.NewSource(seed))
	r1, r2 := replay.Float64(), replay.Float64()
	for dimension := range current {
		want := (best[dimension]-mean[dimension])*0.1 - r1 + ((20*r2)-10)*0.1
		if got[dimension] != want {
			t.Errorf("X3 dimension %d = %.17g, want %.17g", dimension, got[dimension], want)
		}
	}
}

func TestAquilaX4MatchesPublishedEquationAtFinalIteration(t *testing.T) {
	const (
		seed    = 1014
		maxIter = 100
	)
	current := []float64{1, -2}
	best := []float64{0.5, 0.25}
	got := aquilaNarrowedExploitation(
		current, best, maxIter-1, maxIter, -10, 10, rand.New(rand.NewSource(seed)),
	)

	replay := rand.New(rand.NewSource(seed))
	exponent := (2*replay.Float64() - 1) / math.Pow(1-maxIter, 2)
	qf := math.Pow(maxIter, exponent)
	g1 := 2*replay.Float64() - 1
	r1, r2 := replay.Float64(), replay.Float64()
	levy := referenceAquilaLevy(len(current), replay)
	for dimension := range current {
		// At t=T, G2 is exactly zero.
		want := qf*best[dimension] - g1*current[dimension]*r1 + r2*g1
		if got[dimension] != want {
			t.Errorf("X4 dimension %d = %.17g, want %.17g (levy=%g)",
				dimension, got[dimension], want, levy[dimension])
		}
	}
}

func TestAquilaProgressUsesOneBasedAppliedIterations(t *testing.T) {
	if got := aquilaProgress(0, 10); got != 0.1 {
		t.Errorf("first iteration progress = %v, want 0.1", got)
	}
	if got := aquilaProgress(9, 10); got != 1 {
		t.Errorf("final iteration progress = %v, want 1", got)
	}
}
