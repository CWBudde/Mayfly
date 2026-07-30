package mayfly

import (
	"math/rand"
	"strings"
	"testing"
)

// sphere is a trivially minimizable objective, used here only to make Optimize
// runnable; these tests are about validation, not about convergence.
func sphere(position []float64) float64 {
	sum := 0.0
	for _, value := range position {
		sum += value * value
	}

	return sum
}

// smallConfig builds a valid configuration with the given populations and
// offspring count, deterministic so a failure is reproducible.
func smallConfig(npop, npopf, nc, nm int) *Config {
	config := NewDefaultConfig()
	config.ObjectiveFunc = sphere
	config.ProblemSize = 3
	config.LowerBound = -1
	config.UpperBound = 1
	config.MaxIterations = 2
	config.NPop = npop
	config.NPopF = npopf
	config.NC = nc
	config.NM = nm
	config.Rand = rand.New(rand.NewSource(1))

	return config
}

// TestOptimizeRejectsMoreParentPairsThanPopulation covers the crash that made
// this validation necessary: mating reads males[k] and females[k] for
// k < NC/2, so a population smaller than the pair count indexed out of range
// inside the library. The default NC of 20 with any population below 10 hit it,
// which is exactly what a caller shrinking the swarm to go faster would write.
func TestOptimizeRejectsMoreParentPairsThanPopulation(t *testing.T) {
	for _, testCase := range []struct {
		name            string
		npop, npopf, nc int
	}{
		{"males too few", 4, 20, 20},
		{"females too few", 20, 4, 20},
		{"both too few", 4, 4, 20},
		{"one pair short", 9, 9, 20},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			// A panic here is the regression, so let it fail the test loudly
			// rather than be recovered into a pass.
			result, err := Optimize(smallConfig(testCase.npop, testCase.npopf, testCase.nc, 1))
			if err == nil {
				t.Fatalf("Optimize accepted NC=%d with NPop=%d, NPopF=%d, want an error",
					testCase.nc, testCase.npop, testCase.npopf)
			}

			if result != nil {
				t.Errorf("Optimize returned a result alongside an error: %+v", result)
			}

			if !strings.Contains(err.Error(), "parent pairs") {
				t.Errorf("error %q does not explain the parent-pair constraint", err)
			}
		})
	}
}

// TestOptimizeAcceptsExactlyEnoughParents pins the boundary: NC/2 equal to the
// population is the largest legal offspring count, and must not be rejected.
func TestOptimizeAcceptsExactlyEnoughParents(t *testing.T) {
	_, err := Optimize(smallConfig(10, 10, 20, 1))
	if err != nil {
		t.Fatalf("Optimize rejected NC=20 with populations of 10: %v", err)
	}
}

// TestOptimizeRejectsMutantsWithoutOffspring covers the second way NC crashed
// the loop: mutants are drawn from the offspring slice with rng.Intn, which
// panics on an empty slice.
func TestOptimizeRejectsMutantsWithoutOffspring(t *testing.T) {
	for _, nc := range []int{0, 1} {
		result, err := Optimize(smallConfig(10, 10, nc, 3))
		if err == nil {
			t.Fatalf("Optimize accepted NC=%d with NM=3, want an error", nc)
		}

		if result != nil {
			t.Errorf("Optimize returned a result alongside an error: %+v", result)
		}
	}
}

// TestOptimizeAllowsNoOffspringWithoutMutants is the same configuration with
// nothing to draw, which is degenerate but not broken.
func TestOptimizeAllowsNoOffspringWithoutMutants(t *testing.T) {
	_, err := Optimize(smallConfig(10, 10, 0, -1))
	if err == nil {
		t.Fatal("Optimize accepted a negative NM")
	}

	config := smallConfig(10, 10, 0, 0)
	// NM of 0 means "5% of NPop", which is 1 here — not "no mutants" — so the
	// only way to ask for none is a population small enough to round to zero.
	config.NPop = 4
	config.NPopF = 4

	_, err = Optimize(config)
	if err != nil {
		t.Fatalf("Optimize rejected NC=0 with no mutants to draw: %v", err)
	}
}

// TestOptimizeRejectsNegativeOffspringCount keeps NC/2 from being negative,
// which would silently skip mating rather than fail.
func TestOptimizeRejectsNegativeOffspringCount(t *testing.T) {
	if _, err := Optimize(smallConfig(10, 10, -2, 1)); err == nil {
		t.Fatal("Optimize accepted a negative NC")
	}
}

// TestEffectiveNMResolvesTheDefault documents the "0 means 5% of NPop" rule the
// validation has to reproduce, since it runs before the main loop applies it.
func TestEffectiveNMResolvesTheDefault(t *testing.T) {
	for _, testCase := range []struct {
		nm, npop, want int
	}{
		{0, 20, 1},
		{0, 100, 5},
		{0, 4, 0},
		{7, 20, 7},
	} {
		config := &Config{NM: testCase.nm, NPop: testCase.npop}
		if got := effectiveNM(config); got != testCase.want {
			t.Errorf("effectiveNM(NM=%d, NPop=%d) = %d, want %d",
				testCase.nm, testCase.npop, got, testCase.want)
		}
	}
}

// TestConvergenceCurveIsACostHistory pins what the field actually holds, which
// the name it used to carry did not: it has MaxIterations entries rather than
// ProblemSize, and it is non-increasing. Read as the position vector the old
// name advertised, it is nonsense of exactly the plausible-looking kind.
func TestConvergenceCurveIsACostHistory(t *testing.T) {
	config := smallConfig(10, 10, 20, 1)
	config.MaxIterations = 12
	config.ProblemSize = 3

	result, err := Optimize(config)
	if err != nil {
		t.Fatalf("Optimize: %v", err)
	}

	if len(result.ConvergenceCurve) != config.MaxIterations {
		t.Fatalf("ConvergenceCurve has %d entries, want MaxIterations=%d",
			len(result.ConvergenceCurve), config.MaxIterations)
	}

	for i := 1; i < len(result.ConvergenceCurve); i++ {
		if result.ConvergenceCurve[i] > result.ConvergenceCurve[i-1] {
			t.Errorf("ConvergenceCurve rose at %d: %v > %v",
				i, result.ConvergenceCurve[i], result.ConvergenceCurve[i-1])
		}
	}

	if len(result.GlobalBest.Position) != config.ProblemSize {
		t.Errorf("GlobalBest.Position has %d entries, want ProblemSize=%d",
			len(result.GlobalBest.Position), config.ProblemSize)
	}

	// The final curve entry is the best cost, so it must agree with the best
	// solution's cost — this is what makes the two fields tell one story.
	final := result.ConvergenceCurve[len(result.ConvergenceCurve)-1]
	if final != result.GlobalBest.Cost {
		t.Errorf("curve ends at %v but GlobalBest.Cost is %v", final, result.GlobalBest.Cost)
	}
}
