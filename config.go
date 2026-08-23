package mayfly

import "runtime"

// Gravity coefficient decay schedules for MPMA (Config.GravityType).
const (
	GravityLinear      = "linear"
	GravityExponential = "exponential"
	GravitySigmoid     = "sigmoid"
)

// Cooling schedules for the simulated annealing component (Config.CoolingSchedule).
const (
	CoolingExponential = "exponential"
	CoolingLinear      = "linear"
	CoolingLogarithmic = "logarithmic"
)

// NewDefaultConfig creates a default configuration for the standard Mayfly Algorithm.
// You must set ObjectiveFunc, ProblemSize, LowerBound, and UpperBound.
func NewDefaultConfig() *Config {
	return &Config{
		MaxIterations: 2000,
		NPop:          20,
		NPopF:         20,
		G:             0.8,
		GDamp:         1.0,
		A1:            1.0,
		A2:            1.5,
		A3:            1.5,
		Beta:          2.0,
		Dance:         5.0,
		FL:            1.0,
		DanceDamp:     0.8,
		FLDamp:        0.99,
		NC:            NCAuto,
		NM:            0, // Will be calculated as 5% of NPop
		// NC defers to NCRatio, so the offspring count tracks the population
		// instead of standing still at the 20 that v0.4.0 hardcoded. A ratio
		// of 1 reproduces NC == NPop, which is what this configuration already
		// expressed at its own NPop of 20; write NC: 20 to pin the historical
		// count regardless of population.
		NCRatio: 1.0,
		// Selection decides which parents those crossovers use. See
		// SelectionTournament and SelectionRank.
		Selection:      SelectionRank,
		TournamentSize: 3,
		// CrossoverGamma is the blend-crossover expansion factor from the
		// reference implementation; see DefaultCrossoverGamma.
		CrossoverGamma: DefaultCrossoverGamma,
		Mu:             0.01,
		MaxWorkers:     defaultMaxWorkers(),
		// DESMA defaults
		UseDESMA:        false,
		EliteCount:      5,
		SearchRange:     0, // Will be auto-calculated
		EnlargeFactor:   1.05,
		ReductionFactor: 0.95,
	}
}

func defaultMaxWorkers() int {
	return runtime.NumCPU()
}

// NewDESMAConfig creates a default configuration for the DESMA variant.
// You must set ObjectiveFunc, ProblemSize, LowerBound, and UpperBound.
func NewDESMAConfig() *Config {
	config := NewDefaultConfig()
	config.UseDESMA = true

	return config
}

// NewOLCEConfig creates a default configuration for the OLCE-MA variant
// (Orthogonal Learning and Chaotic Exploitation Mayfly Algorithm).
// You must set ObjectiveFunc, ProblemSize, LowerBound, and UpperBound.
//
// OLCE-MA enhances the standard Mayfly Algorithm with:
// - Orthogonal learning to increase diversity and reduce oscillatory movement
// - Chaotic exploitation to improve local search capability
//
// Both stages operate on the elite males only and use greedy acceptance, so
// neither can make the population worse. Setting OrthogonalFactor or
// ChaosFactor to zero disables the corresponding stage completely, including
// its share of the evaluation budget.
func NewOLCEConfig() *Config {
	config := NewDefaultConfig()
	config.UseOLCE = true
	config.OrthogonalFactor = 0.3 // Balanced exploration/exploitation
	config.ChaosFactor = 0.1      // Initial chaotic search radius, decays to zero

	return config
}

// NewEOBBMAConfig creates a default configuration for the EOBBMA variant
// (Elite Opposition-Based Bare Bones Mayfly Algorithm).
// You must set ObjectiveFunc, ProblemSize, LowerBound, and UpperBound.
//
// EOBBMA enhances the standard Mayfly Algorithm with:
// - Gaussian distribution-based "bare bones" framework for exploration
// - Lévy flight for heavy-tailed random jumps to escape local optima
// - Elite opposition-based learning to expand search space coverage
//
// The Bare Bones approach replaces velocity-based updates with Gaussian sampling,
// which can provide better exploration on complex landscapes while reducing
// the number of parameters to tune.
//
// Reference: Elite Opposition-Based Bare Bones Mayfly Algorithm (2024),
// Arabian Journal for Science and Engineering.
func NewEOBBMAConfig() *Config {
	config := NewDefaultConfig()
	config.UseEOBBMA = true
	config.LevyAlpha = 1.5          // Standard Lévy parameter (heavy-tailed)
	config.LevyBeta = 1.0           // Unit scale
	config.OppositionRate = 0.3     // Apply opposition to 30% of elite solutions
	config.EliteOppositionCount = 3 // Top 3 solutions get opposition

	return config
}

// NewMPMAConfig creates a default configuration for the MPMA variant
// (Median Position-Based Mayfly Algorithm).
// You must set ObjectiveFunc, ProblemSize, LowerBound, and UpperBound.
//
// MPMA enhances the standard Mayfly Algorithm with:
// - Median position guidance for better population-level convergence
// - Non-linear gravity coefficient for adaptive exploration/exploitation balance
// - Optional fitness-weighted median for emphasizing better solutions
//
// The Median Position approach uses the population's median rather than just
// the global best, which can provide more stable convergence and better
// resistance to premature convergence on multimodal problems.
//
// Reference: An Improved Mayfly Optimization Algorithm Based on Median Position (2022),
// IEEE Access, DOI: 10.1109/ACCESS.2022.XXXXXXX.
func NewMPMAConfig() *Config {
	config := NewDefaultConfig()
	config.UseMPMA = true
	config.MedianWeight = 0.5          // Balanced influence of median vs global best
	config.GravityType = GravityLinear // Linear decay by default (simplest)
	config.UseWeightedMedian = false   // Standard median by default

	return config
}

// NewGSASMAConfig creates a default configuration for the GSASMA variant
// (Golden Sine Algorithm with Simulated Annealing Mayfly Algorithm).
// You must set ObjectiveFunc, ProblemSize, LowerBound, and UpperBound.
//
// GSASMA enhances the standard Mayfly Algorithm with:
// - Golden Sine Algorithm for adaptive exploration using golden ratio and sine function
// - Simulated Annealing for probabilistic acceptance to escape local optima
// - Hybrid Cauchy-Gaussian mutation for balanced exploration/exploitation
// - Opposition-Based Learning on global best for expanded search coverage
//
// This variant is particularly effective for:
// - Engineering optimization problems with many local optima
// - Problems requiring fast convergence speed
// - Complex multimodal landscapes where standard algorithms plateau
//
// Key advantages:
// - 10-20% improvement in convergence speed on engineering problems
// - Better escape from local optima through SA acceptance
// - Adaptive mutation strategy that transitions from exploration to exploitation
// - Minimal tuning required with sensible defaults
//
// Reference: Improved mayfly algorithm based on hybrid mutation (2022),
// Electronics Letters / IEEE.
func NewGSASMAConfig() *Config {
	config := NewDefaultConfig()
	config.UseGSASMA = true
	config.InitialTemperature = 100.0           // High initial temp for early exploration
	config.CoolingRate = 0.95                   // Gradual cooling (95% per iteration)
	config.CauchyMutationRate = 0.3             // 30% Cauchy, 70% Gaussian by late phase
	config.GoldenFactor = 1.0                   // Standard golden sine influence
	config.CoolingSchedule = CoolingExponential // Fast early cooling, slow late cooling
	config.ApplyOBLToGlobalBest = true          // Enable OBL for better coverage

	return config
}

// NewAOBLMOAConfig creates a default configuration for the AOBLMOA variant
// (Aquila Optimizer and Opposition-Based Learning Mayfly Optimization
// Algorithm).
// You must set ObjectiveFunc, ProblemSize, LowerBound, and UpperBound.
//
// AOBLMOA changes two stages of the standard Mayfly Algorithm:
//   - In the update phase it keeps the attraction branches and replaces the
//     nuptial dance (males) and the random flight (females) with Aquila
//     Optimizer hunting strategies. Which branch an individual takes is a
//     deterministic fitness test, not a draw.
//   - In the offspring phase it replaces Gaussian mutation with stochastic
//     opposition-based learning applied to every offspring, followed by greedy
//     selection.
//
// The Aquila Optimizer contributes four hunting behaviors:
// 1. Expanded exploration (X1): High soar with vertical stoop for global search
// 2. Narrowed exploration (X2): Contour flight with short glide for local exploration
// 3. Expanded exploitation (X3): Low flight with slow descent for convergence
// 4. Narrowed exploitation (X4): Walk and grab for intensive local search
//
// The individual's sex and the iteration phase together fix which of the four
// applies; see aoblmoaStrategyFor, which carries the paper's one unresolved
// contradiction on that mapping.
//
// Note that Optimize is single-objective. AOBLMOA's Pareto helpers are
// exported for callers who build a front themselves; the search does not read
// one.
//
// AquilaWeight is deprecated and defaults to AquilaWeightAuto. The published
// algorithm has no such knob: it decides each individual's branch by a
// deterministic fitness test, not by chance. Setting a probability in [0, 1]
// restores the pre-v0.6.0 behavior of drawing the branch at random.
//
// StrategySwitch is the first iteration of the Aquila exploitation phase. Zero
// defers to two thirds of MaxIterations, the split the Aquila Optimizer paper
// prescribes; a value at or beyond MaxIterations keeps the run in exploration
// throughout.
//
// OppositionProbability is unused by AOBLMOA, which applies stochastic
// opposition-based learning to every offspring unconditionally. It is kept
// because the other opposition-based variants read it.
//
// ArchiveSize sizes the exported ParetoArchive helper. The optimizer no longer
// maintains an archive of its own, because nothing in the search ever read one.
//
// Reference:
// Zhao, Y.; Huang, C.; Zhang, M.; Cui, Y. AOBLMOA: A Hybrid Biomimetic
// Optimization Algorithm for Numerical Optimization and Engineering Design
// Problems. Biomimetics 2023, 8(4), 381. DOI: 10.3390/biomimetics8040381.
func NewAOBLMOAConfig() *Config {
	config := NewDefaultConfig()
	config.UseAOBLMOA = true
	config.AquilaWeight = AquilaWeightAuto // Deterministic branch test, as in the paper
	config.OppositionProbability = 0.3     // Unused by AOBLMOA; see the doc comment
	config.ArchiveSize = 100               // Capacity of a caller-managed ParetoArchive
	config.StrategySwitch = 0              // Resolved to MaxIterations * 2/3 per run

	return config
}
