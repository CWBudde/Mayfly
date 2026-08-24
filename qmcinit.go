// Quasi-random initialization of the initial population.
//
// A metaheuristic's first generation is a sample of the search box, and a
// uniform random sample of forty points in ten dimensions leaves gaps and
// clusters that no amount of later movement is told about. A low-discrepancy
// sequence fills the same box more evenly for the same number of evaluations,
// which is why quasi-random initialization is a standard cheap addition to
// population-based optimizers.
//
// Whether it helps *this* algorithm on *these* problems is a measurement, not
// an assumption. See docs/qmc-initialization.md for the numbers.

package mayfly

import (
	"context"
	"fmt"
	"math/bits"
	"math/rand"

	"github.com/cwbudde/qmc"
)

// Initial-population strategies (Config.QMCInit).
//
// QMCInitUniform is the zero value's behavior and the historical one: every
// coordinate an independent draw from the run's *rand.Rand. The other two seed
// the population from a low-discrepancy sequence instead — Sobol up to the
// 1024 dimensions its direction numbers cover, Halton with no ceiling.
//
// Config.QMCSeed goes with these: it pins the scramble, and left at zero the
// scramble is drawn from the run's RNG so that repeated runs start from
// different point sets. See newQMCSequence, and docs/qmc-initialization.md for
// what the choice is worth on the benchmark suite.
const (
	QMCInitUniform = "uniform"
	QMCInitSobol   = "sobol"
	QMCInitHalton  = "halton"
)

// haltonBurnIn is the number of Halton points discarded before the first one
// used. The early points of a Halton sequence sit in a corner of the box in
// every coordinate with a large base — point 1 is (1/2, 1/3, 1/5, ...) — and
// skipping a few dozen is the standard remedy.
const haltonBurnIn = 64

// validateQMCInit reports whether config.QMCInit names a strategy this package
// implements.
//
// It deliberately does not check the problem size against the dimension
// ceiling of the underlying sequence. That ceiling belongs to qmc — Sobol
// stops where its embedded direction numbers stop — and restating the number
// here would leave two copies to disagree the next time qmc's table grows.
// initialPositions lets the constructor answer instead, and its error names
// the ceiling.
func validateQMCInit(config *Config) error {
	switch config.QMCInit {
	case "", QMCInitUniform, QMCInitSobol, QMCInitHalton:
		return nil
	default:
		return fmt.Errorf("qmc_init must be %q, %q or %q (got %q)",
			QMCInitUniform, QMCInitSobol, QMCInitHalton, config.QMCInit)
	}
}

// quasiRandomPositions draws the starting positions for the whole population:
// config.NPop male rows followed by config.NPopF female rows, each of
// config.ProblemSize coordinates in [config.LowerBound, config.UpperBound].
//
// It returns nil for the uniform strategy rather than a block of independent
// draws. Optimize takes each uniform position at the moment it needs it, and
// only for the individuals no caller-supplied initial population already
// covers; returning a full block here would consume the run's generator for
// positions that are then discarded, and change every subsequent draw.
//
// Males and females come out of one stream rather than two. Two generators
// with the same configuration produce the same points, so seeding the halves
// separately would place every female exactly on top of a male and halve the
// initial coverage — the opposite of the point.
func quasiRandomPositions(config *Config, rng *rand.Rand) ([][]float64, error) {
	return quasiRandomPositionsContext(context.Background(), config, rng)
}

func quasiRandomPositionsContext(
	ctx context.Context,
	config *Config,
	rng *rand.Rand,
) ([][]float64, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	total := config.NPop + config.NPopF

	if config.QMCInit == "" || config.QMCInit == QMCInitUniform {
		// No block is the uniform strategy's answer, not a missing result.
		return nil, nil
	}

	seq, err := newQMCSequence(config, total, rng)
	if err != nil {
		return nil, err
	}

	span := config.UpperBound - config.LowerBound
	rows := make([][]float64, total)
	point := make([]float64, config.ProblemSize)

	for i := range rows {
		err := ctx.Err()
		if err != nil {
			return nil, err
		}

		seq.NextInto(point)

		row := make([]float64, config.ProblemSize)
		for j, u := range point {
			// qmc guarantees u in [0,1), so row[j] lands in [lower, upper).
			row[j] = config.LowerBound + u*span
		}

		rows[i] = row
	}

	return rows, nil
}

// newQMCSequence builds the generator named by config.QMCInit, randomized so
// that repeated runs of the same configuration explore different point sets.
//
// The randomization seed is the load-bearing detail. A quasi-random sequence
// is deterministic, so an unrandomized initial population would be identical
// in every run and a thirty-run study would report a standard deviation that
// measured only the algorithm's own stochasticity downstream of it. Drawing
// the seed from the run's RNG makes each run a distinct *randomized* QMC
// sample — still low-discrepancy, no longer the same one twice — while
// Config.QMCSeed pins it. Note that it pins only the initial population: the
// rest of the run reads Config.Rand, which Optimize seeds from the clock when
// the caller leaves it nil, so reproducing a whole run means pinning both.
func newQMCSequence(config *Config, total int, rng *rand.Rand) (qmc.Sequence, error) {
	seed := config.QMCSeed
	if seed == 0 {
		// Same nil-rng fallback as unifrnd: the helpers in this package accept
		// a nil generator and reach for the global one.
		if rng == nil {
			seed = rand.Uint64()
		} else {
			seed = rng.Uint64()
		}
	}

	switch config.QMCInit {
	case QMCInitSobol:
		// Sobol's balance property covers a block of raw indices aligned on a
		// power of two, not any run of consecutive points, and qmc maps point
		// i to raw index skip+1+i. Skipping 2^m-1 with 2^m >= total therefore
		// puts the whole population inside one aligned block; the default skip
		// of 0 would straddle two and give up the stratification this is here
		// for.
		block := 1 << bits.Len(uint(total-1))

		seq, err := qmc.NewSobol(config.ProblemSize, qmc.WithSkip(block-1), qmc.WithOwenScrambling(seed))
		if err != nil {
			return nil, fmt.Errorf("qmc_init %q: %w", config.QMCInit, err)
		}

		return seq, nil

	case QMCInitHalton:
		seq, err := qmc.NewHalton(config.ProblemSize, qmc.WithSkip(haltonBurnIn), qmc.WithNestedScrambling(seed))
		if err != nil {
			return nil, fmt.Errorf("qmc_init %q: %w", config.QMCInit, err)
		}

		return seq, nil

	default:
		return nil, fmt.Errorf("qmc_init %q is not a known sequence", config.QMCInit)
	}
}

// fillInitialPosition writes the starting position for population index i into
// dst, which is the slice newMayfly already allocated. The male index is i and
// the female index is NPop+i.
//
// positions is what quasiRandomPositions returned, so nil means the uniform
// strategy and the coordinates are drawn here. The indexing is deliberate
// rather than a running cursor: a caller-supplied initial population overrides
// individual indices, and row i has to stay row i so that the points actually
// used remain a subset of one low-discrepancy block.
func fillInitialPosition(config *Config, positions [][]float64, index int, dst []float64, rng *rand.Rand) {
	if positions == nil {
		// One draw per coordinate, in the same order unifrndVec would use, so
		// the uniform path consumes the generator exactly as it always has.
		for j := range dst {
			dst[j] = unifrnd(config.LowerBound, config.UpperBound, rng)
		}

		return
	}

	copy(dst, positions[index])
}
