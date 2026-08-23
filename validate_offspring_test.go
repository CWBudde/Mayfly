package mayfly

import (
	"fmt"
	"math"
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
		// wantError pins which validation reports the case. "males too few"
		// also has more females than males, and that pairing failure is the
		// more fundamental one, so Optimize reports it first. Naming the
		// expected message per case keeps that precedence deliberate instead
		// of an accident of validation order.
		wantError string
	}{
		{"males too few", 4, 20, 20, "must not exceed NPop"},
		{"females too few", 20, 4, 20, "parent pairs"},
		{"both too few", 4, 4, 20, "parent pairs"},
		{"one pair short", 9, 9, 20, "parent pairs"},
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

			if !strings.Contains(err.Error(), testCase.wantError) {
				t.Errorf("error %q does not contain %q", err, testCase.wantError)
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
	_, err := Optimize(smallConfig(10, 10, -2, 1))
	if err == nil {
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
// the name it used to carry did not: it has one entry per completed iteration
// rather than ProblemSize entries, and it is non-increasing. Read as the
// position vector the old name advertised, it is nonsense of exactly the
// plausible-looking kind.
func TestConvergenceCurveIsACostHistory(t *testing.T) {
	config := smallConfig(10, 10, 20, 1)
	config.MaxIterations = 12
	config.ProblemSize = 3

	result, err := Optimize(config)
	if err != nil {
		t.Fatalf("Optimize: %v", err)
	}

	if len(result.ConvergenceCurve) != result.IterationCount {
		t.Fatalf("ConvergenceCurve has %d entries, want IterationCount=%d",
			len(result.ConvergenceCurve), result.IterationCount)
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

// TestOptimizeRejectsMutationRateOutsideUnitInterval covers the panic that made
// this validation necessary. Mu is a fraction of the dimensions, and the
// mutation operators turn it into a count with ceil(Mu*ProblemSize) and then
// slice a permutation of the dimensions to that length. A Mu above 1 asks for
// more dimensions than exist, a negative Mu asks for a negative count, and NaN
// converts to the most negative int, so all three used to panic inside the
// library partway through a run.
//
// The documented range has always been [0,1] and LoadConfig already enforced
// it, so only the programmatic path was exposed.
func TestOptimizeRejectsMutationRateOutsideUnitInterval(t *testing.T) {
	for _, testCase := range []struct {
		name string
		mu   float64
	}{
		{"just above one", 1.0000001},
		{"above one", 1.5},
		{"twice the dimensions", 2},
		{"negative", -0.1},
		{"not a number", math.NaN()},
		{"positive infinity", math.Inf(1)},
		{"negative infinity", math.Inf(-1)},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			// A panic here is the regression, so let it fail the test loudly
			// rather than be recovered into a pass.
			config := smallConfig(20, 20, 20, 1)
			config.Mu = testCase.mu

			result, err := Optimize(config)
			if err == nil {
				t.Fatalf("Optimize accepted Mu=%v, want an error", testCase.mu)
			}

			if result != nil {
				t.Errorf("Optimize returned a result alongside an error: %+v", result)
			}

			if !strings.Contains(err.Error(), "Mu") {
				t.Errorf("error %q does not name the offending field", err)
			}
		})
	}
}

// TestOptimizeAcceptsMutationRatesInsideUnitInterval pins the other half of the
// boundary: every rate the documentation permits must still run, including both
// endpoints. Mu=1 mutates every dimension, which is the case that sits directly
// against the rejected 1.0000001 above.
func TestOptimizeAcceptsMutationRatesInsideUnitInterval(t *testing.T) {
	for _, mu := range []float64{0, 0.01, 0.5, 1} {
		t.Run(fmt.Sprintf("mu=%v", mu), func(t *testing.T) {
			config := smallConfig(20, 20, 20, 1)
			config.Mu = mu

			result, err := Optimize(config)
			if err != nil {
				t.Fatalf("Optimize rejected the documented Mu=%v: %v", mu, err)
			}

			if math.IsNaN(result.GlobalBest.Cost) || math.IsInf(result.GlobalBest.Cost, 0) {
				t.Errorf("Optimize returned a non-finite cost %v for Mu=%v",
					result.GlobalBest.Cost, mu)
			}
		})
	}
}

// TestMutationCountSaturatesOutOfRangeRates covers the exported mutation
// operators, which take a rate directly and so stay reachable with an
// out-of-range one even though Optimize now rejects it first. Saturating keeps
// them from panicking on a bound they cannot slice with.
func TestMutationCountSaturatesOutOfRangeRates(t *testing.T) {
	const nVar = 10

	for _, testCase := range []struct {
		name string
		mu   float64
		want int
	}{
		{"zero mutates nothing", 0, 0},
		{"a tenth rounds up to one", 0.05, 1},
		{"a fraction rounds up", 0.42, 5},
		{"one mutates everything", 1, nVar},
		{"above one saturates", 2.5, nVar},
		{"negative floors at zero", -0.5, 0},
		{"not a number floors at zero", math.NaN(), 0},
		{"positive infinity saturates", math.Inf(1), nVar},
		{"negative infinity floors at zero", math.Inf(-1), 0},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if got := mutationCount(testCase.mu, nVar); got != testCase.want {
				t.Errorf("mutationCount(%v, %d) = %d, want %d",
					testCase.mu, nVar, got, testCase.want)
			}
		})
	}
}

// TestMutationOperatorsSurviveOutOfRangeRates is the end-to-end form of the
// case above: both exported operators used to panic on these rates.
func TestMutationOperatorsSurviveOutOfRangeRates(t *testing.T) {
	for _, mu := range []float64{-1, 1.0000001, 2, math.NaN(), math.Inf(1), math.Inf(-1)} {
		t.Run(fmt.Sprintf("mu=%v", mu), func(t *testing.T) {
			position := []float64{0, 0.25, -0.5, 1, -1}

			for name, mutate := range map[string]func() []float64{
				"MutateGaussian": func() []float64 {
					return MutateGaussian(position, mu, -1, 1, rand.New(rand.NewSource(1)))
				},
				"MutateCauchy": func() []float64 {
					return MutateCauchy(position, mu, -1, 1, rand.New(rand.NewSource(1)))
				},
			} {
				mutated := mutate()
				if len(mutated) != len(position) {
					t.Errorf("%s(mu=%v) returned %d dimensions, want %d",
						name, mu, len(mutated), len(position))
				}

				for i, value := range mutated {
					if value < -1 || value > 1 {
						t.Errorf("%s(mu=%v) put dimension %d at %v, outside [-1,1]",
							name, mu, i, value)
					}
				}
			}
		})
	}
}

// TestOptimizeRejectsMoreFemalesThanMales covers a second index-out-of-range
// crash in the same family: every female update phase pairs females[i] with
// males[i], so NPopF above NPop indexed past the end of the male slice. The
// standard path and the EOBBMA path both panicked, sequentially and in
// parallel; only AOBLMOA carried an ad-hoc guard.
func TestOptimizeRejectsMoreFemalesThanMales(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		variant func() *Config
	}{
		{"standard", NewDefaultConfig},
		{"eobbma", NewEOBBMAConfig},
		{"aoblmoa", NewAOBLMOAConfig},
	} {
		for _, parallel := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s_parallel_%t", testCase.name, parallel), func(t *testing.T) {
				// A panic here is the regression, so let it fail the test
				// loudly rather than be recovered into a pass.
				config := testCase.variant()
				config.ObjectiveFunc = Sphere
				config.ProblemSize = 3
				config.LowerBound = -5.0
				config.UpperBound = 5.0
				config.MaxIterations = 5
				config.NPop = 6
				config.NPopF = 12
				config.NC = 6
				config.NM = 1
				config.EnableParallel = parallel
				config.Rand = rand.New(rand.NewSource(3))

				result, err := Optimize(config)
				if err == nil {
					t.Fatal("Optimize accepted NPopF=12 with NPop=6, want an error")
				}

				if result != nil {
					t.Errorf("Optimize returned a result alongside an error: %+v", result)
				}

				if !strings.Contains(err.Error(), "must not exceed NPop") {
					t.Errorf("error %q does not explain the pairing constraint", err)
				}
			})
		}
	}
}

// TestOptimizeAcceptsFewerFemalesThanMales pins the boundary from the other
// side: a smaller female population is legal and must keep working.
func TestOptimizeAcceptsFewerFemalesThanMales(t *testing.T) {
	config := NewDefaultConfig()
	config.ObjectiveFunc = Sphere
	config.ProblemSize = 3
	config.LowerBound = -5.0
	config.UpperBound = 5.0
	config.MaxIterations = 5
	config.NPop = 12
	config.NPopF = 6
	config.NC = 6
	config.NM = 1
	config.Rand = rand.New(rand.NewSource(3))

	_, err := Optimize(config)
	if err != nil {
		t.Fatalf("Optimize rejected NPopF=6 with NPop=12: %v", err)
	}
}
