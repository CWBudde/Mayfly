package mayfly

import (
	"context"
	"math"
	"math/rand"
	"testing"
)

// TestGoldenSectionInitialCoefficients checks the two interior section points
// against hand-computed values. With a = -π, b = π and τ = (√5-1)/2:
//
//	x1 = aτ + b(1-τ) = π(1 - 2τ) = -π·0.2360679774997897 ≈ -0.7416
//	x2 = a(1-τ) + bτ = π(2τ - 1) =  π·0.2360679774997897 ≈  0.7416
func TestGoldenSectionInitialCoefficients(t *testing.T) {
	const tau = 0.6180339887498949

	if math.Abs(goldenRatioConjugate-tau) > 1e-15 {
		t.Fatalf("goldenRatioConjugate = %v, want %v", goldenRatioConjugate, tau)
	}

	if math.Abs(GoldenRatio-goldenRatioConjugate-1) > 1e-15 {
		t.Errorf("golden ratio identity φ - 1/φ = 1 violated: %v", GoldenRatio-goldenRatioConjugate)
	}

	section := newGoldenSection()

	wantX1 := -math.Pi * (2*tau - 1)
	wantX2 := -wantX1

	if math.Abs(section.x1-wantX1) > 1e-12 {
		t.Errorf("x1 = %v, want %v", section.x1, wantX1)
	}

	if math.Abs(section.x2-wantX2) > 1e-12 {
		t.Errorf("x2 = %v, want %v", section.x2, wantX2)
	}

	if math.Abs(section.x1+0.7416294238611396) > 1e-12 {
		t.Errorf("x1 = %v, want -0.7416294238611396", section.x1)
	}

	if section.x1 >= section.x2 {
		t.Errorf("expected x1 < x2, got x1 = %v, x2 = %v", section.x1, section.x2)
	}
}

// TestGoldenSectionNarrows verifies that the interval shrinks by the golden
// ratio on every update and resets once it has collapsed.
func TestGoldenSectionNarrows(t *testing.T) {
	for _, improved := range []bool{true, false} {
		section := newGoldenSection()
		width := section.b - section.a

		for step := range 10 {
			section.update(improved)

			newWidth := section.b - section.a
			if newWidth > width {
				// A reset restores the full interval, which is expected.
				if math.Abs(newWidth-2*math.Pi) > 1e-12 {
					t.Fatalf("improved=%v step %d: interval grew to %v", improved, step, newWidth)
				}

				width = newWidth

				continue
			}

			// Both branches keep the τ share of the interval, exactly as a
			// golden section search does.
			if math.Abs(newWidth-width*goldenRatioConjugate) > 1e-12 {
				t.Fatalf("improved=%v step %d: width %v, want %v",
					improved, step, newWidth, width*goldenRatioConjugate)
			}

			if section.x1 < section.a || section.x1 > section.b ||
				section.x2 < section.a || section.x2 > section.b {
				t.Fatalf("improved=%v step %d: section points outside [%v, %v]",
					improved, step, section.a, section.b)
			}

			width = newWidth
		}
	}
}

// TestGoldenSineUpdateMatchesPublishedRule reproduces the update rule by hand
// for a fixed RNG stream. This is the guard that keeps the golden section
// coefficients wired into the update: an implementation that ignores x1 and x2
// (as the Sine-Cosine-style rule it replaced did) cannot pass it.
func TestGoldenSineUpdateMatchesPublishedRule(t *testing.T) {
	const seed = 7

	position := []float64{1.5, -2.0, 0.25}
	best := []float64{0.5, 0.5, 0.5}
	section := newGoldenSection()

	got := goldenSineUpdate(position, best, 1.0, section.snapshot(), -5, 5, rand.New(rand.NewSource(seed)))

	reference := rand.New(rand.NewSource(seed))
	r1 := reference.Float64() * 2 * math.Pi
	r2 := reference.Float64() * math.Pi

	for i := range position {
		want := position[i]*math.Abs(math.Sin(r1)) -
			r2*math.Sin(r1)*math.Abs(section.x1*best[i]-section.x2*position[i])
		want = math.Max(-5, math.Min(5, want))

		if math.Abs(got[i]-want) > 1e-12 {
			t.Errorf("dimension %d: got %v, want %v", i, got[i], want)
		}
	}
}

// TestGoldenSineUpdateRespectsBounds checks the clamping of the update.
func TestGoldenSineUpdateRespectsBounds(t *testing.T) {
	rng := rand.New(rand.NewSource(3))
	position := []float64{4.9, -4.9, 0}
	best := []float64{-5, 5, 5}

	for range 200 {
		out := goldenSineUpdate(position, best, 2.0, newGoldenSection().snapshot(), -5, 5, rng)
		for i, v := range out {
			if v < -5 || v > 5 || math.IsNaN(v) {
				t.Fatalf("dimension %d out of bounds: %v", i, v)
			}
		}
	}
}

// TestGSASMAGoldenFactorIsNotInert guards against a GSASMA whose characteristic
// knob has no observable effect on the search.
func TestGSASMAGoldenFactorIsNotInert(t *testing.T) {
	run := func(goldenFactor float64, seed int64) float64 {
		config := NewGSASMAConfig()
		config.ObjectiveFunc = Rosenbrock
		config.ProblemSize = 8
		config.LowerBound = -2.048
		config.UpperBound = 2.048
		config.MaxIterations = 80
		config.Rand = rand.New(rand.NewSource(seed))
		config.GoldenFactor = goldenFactor

		result, err := Optimize(config)
		if err != nil {
			t.Fatalf("Optimize() error = %v", err)
		}

		return result.GlobalBest.Cost
	}

	differed := 0

	for seed := int64(1); seed <= 5; seed++ {
		if run(0.5, seed) != run(2.0, seed) {
			differed++
		}
	}

	if differed == 0 {
		t.Error("GoldenFactor has no observable effect on GSASMA results")
	}
}

// TestParallelGoldenSineAdvancesSectionOncePerBatch guards the batch semantics
// of the parallel Golden Sine step: every candidate is generated from one
// section snapshot, so the interval must be narrowed exactly one step for the
// whole batch instead of once per candidate.
func TestParallelGoldenSineAdvancesSectionOncePerBatch(t *testing.T) {
	pool := newEvaluationPool(sphere, 2)
	defer pool.close()

	males := make([]*Mayfly, 4)
	for i := range males {
		males[i] = newMayfly(2)
		males[i].Position = []float64{float64(i + 1), float64(i + 2)}
		males[i].Cost = sphere(males[i].Position)
		copy(males[i].Best.Position, males[i].Position)
		males[i].Best.Cost = males[i].Cost
	}

	section := newGoldenSection()
	initial := section.snapshot()

	improved, worse := initial, initial
	improved.update(true)
	worse.update(false)

	globalBest := Best{Position: []float64{0, 0}, Cost: 0}

	_, err := evaluateParallelGoldenSine(
		context.Background(),
		males,
		1.0,
		&globalBest,
		1.0,
		-5, 5,
		NewAnnealingScheduler(100, 0.95, "exponential"),
		section,
		rand.New(rand.NewSource(7)),
		pool,
	)
	if err != nil {
		t.Fatalf("evaluateParallelGoldenSine: %v", err)
	}

	got := section.snapshot()
	if got != improved && got != worse {
		t.Errorf("section = %+v, want one single step from %+v (improved=%+v, worse=%+v)",
			got, initial, improved, worse)
	}
}
