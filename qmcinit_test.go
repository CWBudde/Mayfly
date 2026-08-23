package mayfly

import (
	"math"
	"math/rand"
	"strconv"
	"testing"
)

// qmcTestConfig is a minimal config for the initializer: forty individuals,
// the package default, over a box that is not symmetric about zero so that a
// scaling mistake cannot cancel out.
func qmcTestConfig(strategy string, dims int) *Config {
	config := NewDefaultConfig()
	config.ObjectiveFunc = Sphere
	config.ProblemSize = dims
	config.LowerBound = -3.0
	config.UpperBound = 7.0
	config.QMCInit = strategy

	return config
}

func TestValidateQMCInit(t *testing.T) {
	tests := []struct {
		name     string
		strategy string
		wantErr  bool
	}{
		{"empty means uniform", "", false},
		{"uniform", QMCInitUniform, false},
		{"sobol", QMCInitSobol, false},
		{"halton", QMCInitHalton, false},
		{"unknown name", "faure", true},
		{"case matters", "Sobol", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateQMCInit(qmcTestConfig(tt.strategy, 10))
			if (err != nil) != tt.wantErr {
				t.Errorf("validateQMCInit(%q) error = %v, wantErr %v", tt.strategy, err, tt.wantErr)
			}
		})
	}
}

func TestInitialPositionsWithinBounds(t *testing.T) {
	strategies := []string{QMCInitSobol, QMCInitHalton}
	dimensions := []int{1, 2, 10, 30, 64}

	for _, strategy := range strategies {
		for _, dims := range dimensions {
			t.Run(strategy+"/"+strconv.Itoa(dims), func(t *testing.T) {
				config := qmcTestConfig(strategy, dims)

				rows, err := quasiRandomPositions(config, rand.New(rand.NewSource(7)))
				if err != nil {
					t.Fatalf("quasiRandomPositions() error = %v", err)
				}

				if len(rows) != config.NPop+config.NPopF {
					t.Fatalf("quasiRandomPositions() returned %d rows, want %d", len(rows), config.NPop+config.NPopF)
				}

				for i, row := range rows {
					if len(row) != dims {
						t.Fatalf("row %d has %d coordinates, want %d", i, len(row), dims)
					}

					for j, v := range row {
						if math.IsNaN(v) || v < config.LowerBound || v >= config.UpperBound {
							t.Errorf("row %d coordinate %d = %v, want in [%v, %v)",
								i, j, v, config.LowerBound, config.UpperBound)
						}
					}
				}
			})
		}
	}
}

// TestInitialPositionsSeedPins checks that a caller who sets QMCSeed gets the
// same point set regardless of the run's RNG — the reproducibility half of the
// seeding rule.
// TestQuasiRandomPositionsUniformYieldsNoBlock pins the contract Optimize
// depends on: the uniform strategy returns no block at all, so the run keeps
// drawing positions one at a time and consumes its generator exactly as it did
// before this feature existed.
func TestQuasiRandomPositionsUniformYieldsNoBlock(t *testing.T) {
	for _, strategy := range []string{"", QMCInitUniform} {
		t.Run("strategy="+strategy, func(t *testing.T) {
			rows, err := quasiRandomPositions(qmcTestConfig(strategy, 10), rand.New(rand.NewSource(3)))
			if err != nil {
				t.Fatalf("quasiRandomPositions() error = %v", err)
			}

			if rows != nil {
				t.Errorf("quasiRandomPositions() returned %d rows for %q, want none", len(rows), strategy)
			}
		})
	}
}

func TestInitialPositionsSeedPins(t *testing.T) {
	for _, strategy := range []string{QMCInitSobol, QMCInitHalton} {
		t.Run(strategy, func(t *testing.T) {
			config := qmcTestConfig(strategy, 12)
			config.QMCSeed = 20260823

			first, err := quasiRandomPositions(config, rand.New(rand.NewSource(1)))
			if err != nil {
				t.Fatalf("quasiRandomPositions() error = %v", err)
			}

			second, err := quasiRandomPositions(config, rand.New(rand.NewSource(999)))
			if err != nil {
				t.Fatalf("quasiRandomPositions() error = %v", err)
			}

			for i := range first {
				for j := range first[i] {
					if first[i][j] != second[i][j] {
						t.Fatalf("row %d coordinate %d differs with QMCSeed pinned: %v vs %v",
							i, j, first[i][j], second[i][j])
					}
				}
			}
		})
	}
}

// TestInitialPositionsVaryWithoutSeed is the other half: with QMCSeed left at
// zero the scramble is drawn from the run's RNG, so two runs start from
// different point sets. Without this a thirty-run study would report a spread
// that measured nothing about the initial population.
func TestInitialPositionsVaryWithoutSeed(t *testing.T) {
	for _, strategy := range []string{QMCInitSobol, QMCInitHalton} {
		t.Run(strategy, func(t *testing.T) {
			config := qmcTestConfig(strategy, 12)

			first, err := quasiRandomPositions(config, rand.New(rand.NewSource(1)))
			if err != nil {
				t.Fatalf("quasiRandomPositions() error = %v", err)
			}

			second, err := quasiRandomPositions(config, rand.New(rand.NewSource(2)))
			if err != nil {
				t.Fatalf("quasiRandomPositions() error = %v", err)
			}

			if identicalRows(first, second) {
				t.Error("two RNG seeds produced the same initial population; the scramble is not being drawn from the RNG")
			}
		})
	}
}

// TestInitialPositionsDistinctIndividuals guards the mistake of seeding males
// and females from two generators with the same configuration, which would
// place every female exactly on top of a male.
func TestInitialPositionsDistinctIndividuals(t *testing.T) {
	for _, strategy := range []string{QMCInitSobol, QMCInitHalton} {
		t.Run(strategy, func(t *testing.T) {
			config := qmcTestConfig(strategy, 8)

			rows, err := quasiRandomPositions(config, rand.New(rand.NewSource(5)))
			if err != nil {
				t.Fatalf("quasiRandomPositions() error = %v", err)
			}

			for i := range rows {
				for j := i + 1; j < len(rows); j++ {
					if sameRow(rows[i], rows[j]) {
						t.Fatalf("individuals %d and %d have identical positions", i, j)
					}
				}
			}
		})
	}
}

// TestInitialPositionsStratification pins what the sequences buy over uniform
// draws: coverage of every part of every axis.
//
// Sobol's numbers are not a statistical accident. Owen scrambling preserves
// the net property, so an aligned block of forty points splits each axis into
// eight equal parts with none of them empty, for every seed and every
// dimension. Halton carries no such guarantee and is allowed one empty part.
// Uniform draws are not asserted on at all — over these same seeds they leave
// two parts of some axis empty, which is exactly the gap this feature exists
// to close, and pinning it would only make the test brittle.
func TestInitialPositionsStratification(t *testing.T) {
	const bins = 8

	tests := []struct {
		strategy    string
		minOccupied int
	}{
		{QMCInitSobol, bins},
		{QMCInitHalton, bins - 1},
	}

	for _, tt := range tests {
		t.Run(tt.strategy, func(t *testing.T) {
			for _, dims := range []int{2, 10, 30, 64} {
				for _, seed := range []int64{1, 2, 3, 17, 20260823} {
					config := qmcTestConfig(tt.strategy, dims)

					rows, err := quasiRandomPositions(config, rand.New(rand.NewSource(seed)))
					if err != nil {
						t.Fatalf("quasiRandomPositions() error = %v", err)
					}

					worst, axis := worstAxisOccupancy(rows, config.LowerBound, config.UpperBound, bins)
					if worst < tt.minOccupied {
						t.Errorf("dims=%d seed=%d: axis %d occupies %d of %d bins, want at least %d",
							dims, seed, axis, worst, bins, tt.minOccupied)
					}
				}
			}
		})
	}
}

// TestInitialPositionsRejectsBadStrategy checks the two ways this can fail:
// a name no sequence answers to, and a dimension count past the ceiling of the
// direction numbers qmc embeds. The second error is qmc's own, so the test
// asserts only that it arrives rather than restating its wording.
func TestInitialPositionsRejectsBadStrategy(t *testing.T) {
	tests := []struct {
		name     string
		strategy string
		dims     int
	}{
		{"unknown sequence", "faure", 10},
		{"sobol past the direction-number table", QMCInitSobol, 1025},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := quasiRandomPositions(qmcTestConfig(tt.strategy, tt.dims), rand.New(rand.NewSource(1)))
			if err == nil {
				t.Fatalf("quasiRandomPositions() with %q at %d dimensions: want an error, got none", tt.strategy, tt.dims)
			}
		})
	}
}

// TestOptimizeQMCInit runs the whole algorithm through the new path: a pinned
// scramble seed and a pinned RNG reproduce a run exactly, an unknown strategy
// is refused before any evaluation, and Halton works as well as Sobol.
func TestOptimizeQMCInit(t *testing.T) {
	run := func(strategy string, qmcSeed uint64) (*Result, error) {
		config := qmcTestConfig(strategy, 10)
		config.MaxIterations = 20
		config.QMCSeed = qmcSeed
		config.Rand = rand.New(rand.NewSource(42))

		return Optimize(config)
	}

	for _, strategy := range []string{QMCInitSobol, QMCInitHalton} {
		t.Run(strategy, func(t *testing.T) {
			first, err := run(strategy, 4711)
			if err != nil {
				t.Fatalf("Optimize() error = %v", err)
			}

			second, err := run(strategy, 4711)
			if err != nil {
				t.Fatalf("Optimize() error = %v", err)
			}

			if first.GlobalBest.Cost != second.GlobalBest.Cost {
				t.Errorf("pinned run is not reproducible: %v vs %v",
					first.GlobalBest.Cost, second.GlobalBest.Cost)
			}

			other, err := run(strategy, 1234)
			if err != nil {
				t.Fatalf("Optimize() error = %v", err)
			}

			if first.GlobalBest.Cost == other.GlobalBest.Cost {
				t.Error("a different QMCSeed produced an identical run; the seed is not reaching the sequence")
			}
		})
	}

	t.Run("unknown strategy", func(t *testing.T) {
		_, err := run("faure", 0)
		if err == nil {
			t.Error("Optimize() with an unknown qmc_init: want an error, got none")
		}
	})
}

// worstAxisOccupancy reports the smallest number of occupied bins over all
// axes, and which axis that was, after splitting [lower, upper) into bins
// equal parts.
func worstAxisOccupancy(rows [][]float64, lower, upper float64, bins int) (int, int) {
	dims := len(rows[0])
	worst := bins
	worstAxis := 0
	span := upper - lower

	for d := range dims {
		occupied := make([]bool, bins)

		for _, row := range rows {
			bin := int((row[d] - lower) / span * float64(bins))
			if bin >= 0 && bin < bins {
				occupied[bin] = true
			}
		}

		count := 0

		for _, seen := range occupied {
			if seen {
				count++
			}
		}

		if count < worst {
			worst = count
			worstAxis = d
		}
	}

	return worst, worstAxis
}

func identicalRows(a, b [][]float64) bool {
	for i := range a {
		if !sameRow(a[i], b[i]) {
			return false
		}
	}

	return true
}

func sameRow(a, b []float64) bool {
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
