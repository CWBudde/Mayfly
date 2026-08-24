package mayfly

import (
	"math"
	"math/rand"
	"testing"
)

func TestGSASMAGoldenPositionMatchesEquation10(t *testing.T) {
	const seed = int64(71)

	position := []float64{1.25, -0.75}
	personalBest := []float64{-0.5, 2}

	got := gsasmaGoldenPosition(
		position, personalBest, -10, 10, rand.New(rand.NewSource(seed)),
	)

	reference := rand.New(rand.NewSource(seed))
	r1 := reference.Float64() * 2 * math.Pi
	r2 := reference.Float64() * math.Pi
	tau := (math.Sqrt(5) - 1) / 2
	c1 := -math.Pi + (1-tau)*2*math.Pi

	c2 := -math.Pi + tau*2*math.Pi
	for i := range position {
		want := position[i]*math.Abs(math.Sin(r1)) -
			r2*math.Sin(r1)*math.Abs(c1*personalBest[i]-c2*position[i])

		want = max(-10.0, min(10.0, want))
		if math.Abs(got[i]-want) > 1e-15 {
			t.Errorf("dimension %d: Eq. (10) = %v, want %v", i, got[i], want)
		}
	}
}

func TestGSASMAVelocityMateriallyFeedsPosition(t *testing.T) {
	const seed = int64(83)

	position := []float64{0.4, -0.2}
	personalBest := []float64{1.2, -1.5}

	stationary := gsasmaPositionStep(
		position, []float64{0, 0}, personalBest, -10, 10,
		rand.New(rand.NewSource(seed)),
	)
	moved := gsasmaPositionStep(
		position, []float64{0.75, -0.5}, personalBest, -10, 10,
		rand.New(rand.NewSource(seed)),
	)

	if slicesEqual(stationary, moved) {
		t.Fatalf("SA-selected velocity is inert: stationary=%v moved=%v", stationary, moved)
	}

	want := gsasmaGoldenPosition(
		[]float64{1.15, -0.7}, personalBest, -10, 10,
		rand.New(rand.NewSource(seed)),
	)
	for i := range want {
		if math.Abs(moved[i]-want[i]) > 1e-15 {
			t.Fatalf("dimension %d: composed position = %v, want %v", i, moved[i], want[i])
		}
	}
}

func TestGSASMAPreviousCostsFollowSurvivorIdentityAndStayBounded(t *testing.T) {
	a := newMayfly(1)
	b := newMayfly(1)
	c := newMayfly(1)
	discarded := newMayfly(1)
	a.Cost, b.Cost, c.Cost, discarded.Cost = 1, 2, 3, 99

	previous := map[*Mayfly]float64{a: 11, b: 22, discarded: 999}
	// Deliberately reorder the incumbents and introduce one new offspring.
	retained := retainGSASMAPreviousCosts(previous, []*Mayfly{b, a}, []*Mayfly{c})

	if len(retained) != 3 {
		t.Fatalf("retained state length = %d, want population size 3", len(retained))
	}

	if retained[a] != 11 || retained[b] != 22 {
		t.Fatalf("previous costs did not follow identity after reorder: a=%v b=%v",
			retained[a], retained[b])
	}

	if retained[c] != c.Cost {
		t.Fatalf("new survivor previous cost = %v, want current cost %v", retained[c], c.Cost)
	}

	if _, ok := retained[discarded]; ok {
		t.Fatal("discarded offspring remained in previous-fitness state")
	}
}

func slicesEqual(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
