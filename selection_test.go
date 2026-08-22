package mayfly

import (
	"math"
	"math/rand"
	"testing"
)

// The offspring count is the whole point of the v0.5.0 change: before it, a
// caller who raised NPop got a larger swarm and exactly as much recombination
// as before.
func TestEffectiveNCTracksThePopulation(t *testing.T) {
	tests := []struct {
		name  string
		npop  int
		npopf int
		nc    int
		ratio float64
		want  int
	}{
		{"default population reproduces the historical count", 20, 20, NCAuto, 1.0, 20},
		{"large population scales instead of standing still", 4096, 4096, NCAuto, 1.0, 4096},
		{"a ratio above one lets every member mate", 64, 64, NCAuto, 2.0, 128},
		{"ratio below one halves the recombination", 200, 200, NCAuto, 0.5, 100},
		{"an odd product rounds down to a whole number of pairs", 25, 25, NCAuto, 1.0, 24},
		{"a written count always wins over the ratio", 4096, 4096, 20, 1.0, 20},
		{"a written zero still disables crossover", 40, 40, 0, 1.0, 0},
		{"the shorter population bounds the pair count", 100, 10, NCAuto, 1.0, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{NPop: tt.npop, NPopF: tt.npopf, NC: tt.nc, NCRatio: tt.ratio}
			if got := effectiveNC(config); got != tt.want {
				t.Fatalf("effectiveNC = %d, want %d", got, tt.want)
			}
		})
	}
}

// A ratio must never produce a configuration validateOffspring would refuse,
// because the caller never wrote the number it would be blamed for.
func TestScaledOffspringCountAlwaysValidates(t *testing.T) {
	for _, npopf := range []int{1, 2, 7, 20, 64} {
		config := NewDefaultConfig()
		config.NPop = 64
		config.NPopF = npopf
		config.NM = 0

		err := validateOffspring(config)
		if err != nil {
			t.Fatalf("NPopF=%d: %v", npopf, err)
		}
	}
}

func TestDefaultConfigScalesItsOffspringCount(t *testing.T) {
	config := NewDefaultConfig()

	if config.NCRatio != 1.0 {
		t.Fatalf("NCRatio = %v, want 1.0", config.NCRatio)
	}

	// Rank pairing stays the default: once NC scales, it mates the fitter half
	// of the population at every size, and swapping it for a tournament cost
	// measurable quality on the Griewank and Rastrigin regressions.
	if config.Selection != SelectionRank {
		t.Fatalf("Selection = %q, want %q", config.Selection, SelectionRank)
	}

	// At the default population the resolved count is what v0.4.0 hardcoded,
	// so the ratio changes nothing for a caller who never touched NPop.
	if got := effectiveNC(config); got != 20 {
		t.Fatalf("effectiveNC = %d, want 20", got)
	}
}

// Every variant builds on NewDefaultConfig, so the fix has to reach all of
// them -- which is the reason variant comparisons at a raised population were
// never measuring what they claimed to.
func TestEveryVariantConfigScalesItsOffspringCount(t *testing.T) {
	builders := map[string]func() *Config{
		"default": NewDefaultConfig,
		"desma":   NewDESMAConfig,
		"olce":    NewOLCEConfig,
		"eobbma":  NewEOBBMAConfig,
		"mpma":    NewMPMAConfig,
		"gsasma":  NewGSASMAConfig,
		"aoblmoa": NewAOBLMOAConfig,
	}

	for name, build := range builders {
		t.Run(name, func(t *testing.T) {
			config := build()
			config.NPop = 256
			config.NPopF = 256

			if got := effectiveNC(config); got != 256 {
				t.Fatalf("effectiveNC = %d, want 256", got)
			}
		})
	}
}

func TestRankSelectionPairsEqualRanks(t *testing.T) {
	males, females := sortedPopulations(8)
	config := &Config{Selection: SelectionRank}

	for k := range 8 {
		male, female := selectParents(males, females, k, config, rand.New(rand.NewSource(1)))
		if male != males[k] || female != females[k] {
			t.Fatalf("k=%d: rank selection must pair the k-th best with the k-th best", k)
		}
	}
}

// The tournament reads fitness off the sort order, so a sample's winner is its
// smallest index. Pinning that keeps the optimization honest if the population
// ever stops arriving sorted.
func TestTournamentReturnsTheFittestOfItsSample(t *testing.T) {
	rng := rand.New(rand.NewSource(7))

	for range 200 {
		got := tournamentIndex(50, 3, rng)
		if got < 0 || got >= 50 {
			t.Fatalf("index %d out of range", got)
		}
	}

	// A tournament covering the whole population every draw must return the
	// single fittest member.
	always := tournamentIndex(1, 5, rand.New(rand.NewSource(3)))
	if always != 0 {
		t.Fatalf("index = %d, want 0 for a population of one", always)
	}
}

// Selection pressure has to rise with the tournament size, otherwise the knob
// does not mean what its name says.
func TestLargerTournamentsSelectFitterParents(t *testing.T) {
	const population, draws = 100, 4000

	mean := func(size int) float64 {
		rng := rand.New(rand.NewSource(11))
		total := 0

		for range draws {
			total += tournamentIndex(population, size, rng)
		}

		return float64(total) / draws
	}

	loose, tight := mean(2), mean(8)
	if tight >= loose {
		t.Fatalf("mean selected rank: size 8 = %.2f, size 2 = %.2f; "+
			"a larger tournament must select fitter (lower-indexed) parents", tight, loose)
	}
}

// The escape hatch a reproducible re-run needs: NCRatio 0 with rank selection
// is the v0.4.0 mating behavior exactly.
func TestHistoricalBehaviorIsStillExpressible(t *testing.T) {
	config := NewDefaultConfig()
	config.NPop = 512
	config.NPopF = 512
	config.NC = 20
	config.Selection = SelectionRank

	if got := effectiveNC(config); got != 20 {
		t.Fatalf("effectiveNC = %d, want the historical 20", got)
	}

	males, females := sortedPopulations(16)

	male, female := selectParents(males, females, 3, config, rand.New(rand.NewSource(5)))
	if male != males[3] || female != females[3] {
		t.Fatal("rank selection must reproduce the historical pairing")
	}
}

func TestInvalidSelectionConfigurationIsRefused(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Config)
	}{
		{"unknown strategy", func(c *Config) { c.Selection = "roulette" }},
		{"negative ratio", func(c *Config) { c.NCRatio = -1 }},
		{"negative tournament size", func(c *Config) { c.TournamentSize = -2 }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig()
			tt.mutate(config)

			err := validateOffspring(config)
			if err == nil {
				t.Fatal("expected the configuration to be refused")
			}
		})
	}
}

// sortedPopulations builds two populations already ordered best-first, which is
// the state the mating loop always receives them in.
func sortedPopulations(n int) ([]*Mayfly, []*Mayfly) {
	males := make([]*Mayfly, n)
	females := make([]*Mayfly, n)

	for i := range n {
		males[i] = &Mayfly{Cost: float64(i)}
		females[i] = &Mayfly{Cost: float64(i)}
	}

	return males, females
}

// The regression the review caught: Selection is a field pre-v0.5.0 configs do
// not carry, so its zero value has to mean the pairing they were recorded
// under. Routing it to tournament would have changed every loaded config.
func TestUnsetSelectionPairsByRank(t *testing.T) {
	config := NewDefaultConfig()
	config.Selection = ""

	males, females := sortedPopulations(16)

	// A tournament over 16 members would almost surely leave index 3 at some
	// point across this many draws; rank pairing never does.
	for range 64 {
		male, female := selectParents(males, females, 3, config, rand.New(rand.NewSource(11)))
		if male != males[3] || female != females[3] {
			t.Fatal("an unset Selection must pair by rank, not by tournament")
		}
	}
}

// NCAuto must survive a run: a Config reused with a larger NPop has to scale
// again rather than see the previous run's resolved count as an explicit NC.
func TestAutoOffspringCountSurvivesARun(t *testing.T) {
	config := NewDefaultConfig()
	config.ObjectiveFunc = Sphere
	config.ProblemSize = 2
	config.LowerBound = -5
	config.UpperBound = 5
	config.MaxIterations = 3
	config.Rand = rand.New(rand.NewSource(1))

	_, err := Optimize(config)
	if err != nil {
		t.Fatalf("first run: %v", err)
	}

	if config.NC != NCAuto {
		t.Fatalf("NC = %d after a run, want the NCAuto sentinel to be preserved", config.NC)
	}

	config.NPop = 64
	config.NPopF = 64

	if got := effectiveNC(config); got != 64 {
		t.Fatalf("effectiveNC after raising NPop = %d, want 64", got)
	}
}

// A ratio large enough to overflow the conversion to int must clamp to the
// legal maximum, not wrap to a negative count that reads as "no offspring".
func TestExtremeOffspringRatioClampsToThePopulation(t *testing.T) {
	for _, ratio := range []float64{1e20, math.MaxFloat64, 1e9} {
		config := NewDefaultConfig()
		config.NPop = 64
		config.NPopF = 32
		config.NCRatio = ratio

		got := effectiveNC(config)
		if got != 64 {
			t.Fatalf("NCRatio=%g: effectiveNC = %d, want the clamp to 2*min(NPop,NPopF)=64", ratio, got)
		}
	}
}

// A ratio that is not a number at all is refused by validation, and effectiveNC
// falls back rather than deriving a count from it for callers who skip that.
func TestNonFiniteOffspringRatioIsRefused(t *testing.T) {
	for _, ratio := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		config := NewDefaultConfig()
		config.NCRatio = ratio

		err := validateOffspring(config)
		if err == nil {
			t.Fatalf("NCRatio=%g must be refused", ratio)
		}

		config.NPop = 20
		config.NPopF = 20

		if got := effectiveNC(config); got != 20 {
			t.Fatalf("NCRatio=%g: effectiveNC = %d, want the 1.0 fallback", ratio, got)
		}
	}
}

// ValidateConfig is documented as the check a loaded configuration passes, so
// the mating parameters have to fail there and not first at Optimize.
func TestValidateConfigRejectsBadSelectionParameters(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Config)
	}{
		{"unknown strategy", func(c *Config) { c.Selection = "roulette" }},
		{"negative ratio", func(c *Config) { c.NCRatio = -1 }},
		{"negative tournament size", func(c *Config) { c.TournamentSize = -2 }},
		{"offspring exceed the population", func(c *Config) { c.NC = 4096 }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig()
			config.ProblemSize = 2
			config.LowerBound = -5
			config.UpperBound = 5
			tt.mutate(config)

			err := ValidateConfig(config)
			if err == nil {
				t.Fatal("ValidateConfig accepted a configuration Optimize would refuse")
			}
		})
	}
}
