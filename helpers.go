package mayfly

import (
	"fmt"
	"math"
	"math/rand"
	"slices"
)

// unifrnd generates a random float64 between lower and upper.
func unifrnd(lower, upper float64, rng *rand.Rand) float64 {
	if rng == nil {
		return lower + rand.Float64()*(upper-lower)
	}

	return lower + rng.Float64()*(upper-lower)
}

// unifrndVec generates a vector of random float64 values between lower and upper.
func unifrndVec(lower, upper float64, size int, rng *rand.Rand) []float64 {
	vec := make([]float64, size)
	for i := range vec {
		vec[i] = unifrnd(lower, upper, rng)
	}

	return vec
}

// randn generates a normally distributed random number.
func randn(rng *rand.Rand) float64 {
	if rng == nil {
		return rand.NormFloat64()
	}

	return rng.NormFloat64()
}

// maxVec returns element-wise maximum of vector and scalar.
func maxVec(vec []float64, bound float64) {
	for i := range vec {
		if vec[i] < bound {
			vec[i] = bound
		}
	}
}

// minVec returns element-wise minimum of vector and scalar.
func minVec(vec []float64, bound float64) {
	for i := range vec {
		if vec[i] > bound {
			vec[i] = bound
		}
	}
}

// sortMayflies sorts mayflies from most to least preferred.
func sortMayflies(mayflies []*Mayfly, evaluators ...*constraintEvaluator) {
	evaluator := newConstraintEvaluator(nil, nil)
	if len(evaluators) > 0 {
		evaluator = evaluators[0]
	}

	slices.SortStableFunc(mayflies, func(left, right *Mayfly) int {
		switch {
		case evaluator.betterMayfly(left, right):
			return -1
		case evaluator.betterMayfly(right, left):
			return 1
		default:
			return 0
		}
	})
}

// effectiveNM reports the mutant count Optimize will actually use, resolving
// the "0 means 5% of NPop" default the same way the main loop does.
func effectiveNM(config *Config) int {
	if config.NM != 0 {
		return config.NM
	}

	return int(math.Round(0.05 * float64(config.NPop)))
}

// effectiveNC reports the offspring count Optimize will actually use.
//
// A written NC always wins, including the zero that disables crossover: a
// caller who states a count gets it, and nothing the defaults carry may
// silently replace a field someone wrote. Only NCAuto defers to NCRatio.
//
// The ratio exists because NC was an absolute constant through v0.4.0, so
// raising NPop bought a larger swarm and not one extra crossover -- at NPop
// 4096 the same ten pairs mated while 4086 members only followed the global
// best. NCAuto with a ratio of 1 restores NC == NPop, which is the ratio the
// default configuration already expressed at its own NPop of 20.
//
// The result is rounded down to an even number because crossover consumes
// parents in pairs, and clamped so NC/2 never exceeds either population -- the
// same bound validateOffspring enforces for a written NC, applied here so a
// ratio can never derive a configuration the library would refuse.
func effectiveNC(config *Config) int {
	if config.NC != NCAuto {
		return config.NC
	}

	// A non-positive or non-finite ratio falls back to 1.0 rather than
	// deriving a count from it. Zero is included deliberately: it is the zero
	// value of the field, so honoring it literally would give every
	// partially-filled Config{NC: NCAuto} literal a silent no-crossover run --
	// the failure this whole change exists to remove. A caller who wants no
	// offspring writes NC = 0, which effectiveNC returns untouched above.
	ratio := config.NCRatio
	if math.IsNaN(ratio) || math.IsInf(ratio, 0) || ratio <= 0 {
		ratio = 1.0
	}

	// The population clamp is applied in floating point, before the conversion
	// to int, because a large ratio would otherwise overflow the conversion --
	// which is implementation-defined in Go and in practice yields a negative
	// count that the clamp below would read as "no offspring".
	pairs := min(config.NPop, config.NPopF)

	scaled := math.Round(ratio * float64(config.NPop))
	if scaled > float64(2*pairs) {
		scaled = float64(2 * pairs)
	}

	count := int(scaled)
	count -= count % 2

	if count < 0 {
		count = 0
	}

	return count
}

// effectiveTournamentSize reports the tournament size Optimize will use,
// resolving the zero value to the default of 2 -- the smallest tournament that
// still expresses a preference, and the one that applies the least selection
// pressure, which is the safer resolution for a caller who never set it.
func effectiveTournamentSize(config *Config) int {
	if config.TournamentSize > 0 {
		return config.TournamentSize
	}

	return 2
}

// selectParents chooses the parents for the k-th crossover of an iteration.
//
// Both populations arrive sorted best-first, which is what lets a tournament be
// decided without re-reading any cost: the fittest member of a uniform sample
// is simply the one with the smallest index. That keeps selection independent
// of the constraint-aware comparator the sort already applied.
func selectParents(males, females []*Mayfly, k int, config *Config, rng *rand.Rand) (*Mayfly, *Mayfly) {
	// Anything that is not explicitly a tournament pairs by rank, so that the
	// unset Selection of a pre-v0.5.0 configuration -- which has no such field
	// to load -- keeps the pairing it was recorded under. Validation refuses
	// the unknown strategies this would otherwise absorb.
	if config.Selection != SelectionTournament {
		return males[k], females[k]
	}

	size := effectiveTournamentSize(config)

	return males[tournamentIndex(len(males), size, rng)],
		females[tournamentIndex(len(females), size, rng)]
}

// tournamentIndex draws size candidates uniformly from a population of n and
// returns the index of the fittest, relying on the population being sorted
// best-first so that fittest means smallest index.
func tournamentIndex(n, size int, rng *rand.Rand) int {
	best := rng.Intn(n)

	for range size - 1 {
		if candidate := rng.Intn(n); candidate < best {
			best = candidate
		}
	}

	return best
}

// validateOffspring checks NC and NM against the population sizes.
//
// NC drives three separate index expressions in the main loop, none of which
// bounds-check: the mating loop reads males[k] and females[k] for k < NC/2, and
// the mutation step draws a uniform parent from the offspring slice. A caller
// who shrinks the population without also shrinking NC — the default NC of 20
// with any NPop below 10, for instance — used to get an out-of-range panic from
// inside the library rather than an error out of Optimize.
func validateOffspring(config *Config) error {
	if config.NC < 0 && config.NC != NCAuto {
		return fmt.Errorf(
			"NC (offspring count) must be non-negative or NCAuto (%d), got %d",
			NCAuto, config.NC,
		)
	}

	if config.NCRatio < 0 || math.IsNaN(config.NCRatio) || math.IsInf(config.NCRatio, 0) {
		return fmt.Errorf("NCRatio must be a non-negative finite number, got %f", config.NCRatio)
	}

	if config.CrossoverGamma < 0 || math.IsNaN(config.CrossoverGamma) ||
		math.IsInf(config.CrossoverGamma, 0) {
		return fmt.Errorf(
			"CrossoverGamma must be a non-negative finite number, got %f",
			config.CrossoverGamma,
		)
	}

	if config.TournamentSize < 0 {
		return fmt.Errorf("TournamentSize must be non-negative, got %d", config.TournamentSize)
	}

	switch config.Selection {
	case "", SelectionTournament, SelectionRank:
	default:
		return fmt.Errorf(
			"selection strategy must be %q or %q, got %q",
			SelectionTournament, SelectionRank, config.Selection,
		)
	}

	if config.NM < 0 {
		return fmt.Errorf("NM (mutant count) must be non-negative, got %d", config.NM)
	}

	// Mating pairs the k-th best male with the k-th best female, so neither
	// population may be shorter than the number of pairs.
	if pairs := effectiveNC(config) / 2; pairs > config.NPop || pairs > config.NPopF {
		return fmt.Errorf(
			"NC (offspring count) of %d needs %d parent pairs, "+
				"which exceeds NPop=%d or NPopF=%d; lower NC or raise the populations",
			effectiveNC(config), pairs, config.NPop, config.NPopF,
		)
	}

	// Mutants are drawn from the offspring, so there must be at least one.
	if effectiveNC(config) < 2 && effectiveNM(config) > 0 {
		return fmt.Errorf(
			"NC (offspring count) of %d produces no offspring for %d mutants to be drawn from; "+
				"raise NC to at least 2 (note that NM=0 does not disable mutants, "+
				"it selects the default of 5%% of NPop)",
			effectiveNC(config), effectiveNM(config),
		)
	}

	return nil
}

// effectiveCrossoverGamma reports the blend-crossover expansion factor
// Optimize will actually use.
//
// Unlike NC, the zero value is not honored literally. A gamma of zero confines
// the crossover coefficient to [0, 1], which makes every offspring a convex
// combination of its parents -- the contraction this field exists to remove --
// so a partially-filled Config literal that never mentions CrossoverGamma must
// not silently get it. Zero, negative values and NaN therefore all resolve to
// DefaultCrossoverGamma; only a positive, finite value is taken as written.
func effectiveCrossoverGamma(config *Config) float64 {
	gamma := config.CrossoverGamma
	if math.IsNaN(gamma) || math.IsInf(gamma, 0) || gamma <= 0 {
		return DefaultCrossoverGamma
	}

	return gamma
}
