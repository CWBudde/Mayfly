// Core type definitions for the Mayfly Optimization Algorithm.

package mayfly

import (
	"math"
	"math/rand"
)

// ObjectiveFunction represents a function to be optimized.
// It takes a read-only position vector and returns a fitness cost.
type ObjectiveFunction func([]float64) float64

// ConstraintFunction evaluates a constraint at a position. Inequality
// constraints are satisfied when the returned value is less than or equal to
// zero. Equality constraints are satisfied when the absolute returned value is
// within the configured equality tolerance.
type ConstraintFunction func([]float64) float64

// ConstraintHandlingMethod selects how constrained candidates are ranked.
type ConstraintHandlingMethod string

const (
	// ConstraintHandlingFeasibility applies Deb's feasibility rules.
	ConstraintHandlingFeasibility ConstraintHandlingMethod = "feasibility"
	// ConstraintHandlingPenalty ranks candidates by their penalized cost.
	ConstraintHandlingPenalty ConstraintHandlingMethod = "penalty"
)

// PenaltyMethod selects how aggregate constraint violation is penalized.
type PenaltyMethod string

const (
	// PenaltyLinear adds factor * violation to the objective cost.
	PenaltyLinear PenaltyMethod = "linear"
	// PenaltyQuadratic adds factor * violation squared to the objective cost.
	PenaltyQuadratic PenaltyMethod = "quadratic"
)

// ConstraintConfig configures optional problem constraints. Function fields
// are not serialized and must be restored after loading a configuration.
type ConstraintConfig struct {
	Handling          ConstraintHandlingMethod `json:"handling,omitempty"`
	PenaltyMethod     PenaltyMethod            `json:"penalty_method,omitempty"`
	Inequalities      []ConstraintFunction     `json:"-"`
	Equalities        []ConstraintFunction     `json:"-"`
	PenaltyFactor     float64                  `json:"penalty_factor,omitempty"`
	EqualityTolerance float64                  `json:"equality_tolerance,omitempty"`
}

// Best represents the best position and cost found.
type Best struct {
	Position            []float64
	Cost                float64
	ConstraintViolation float64
}

// Mayfly represents a single mayfly (male or female) in the population.
type Mayfly struct {
	Position            []float64
	Velocity            []float64
	Best                Best
	Cost                float64
	ConstraintViolation float64
}

// ConvergenceConfig controls optional early termination. MaxIterations remains
// the hard upper bound; successful target or stagnation checks may shorten a
// run after MinIterations completed iterations.
type ConvergenceConfig struct {
	// TargetCost stops the run when the best cost is less than or equal to the
	// pointed-to value. A pointer distinguishes a disabled target from a target
	// of zero.
	TargetCost *float64 `json:"target_cost,omitempty"`

	// MinImprovement is the absolute cost, penalty score, or constraint-
	// violation reduction required to reset the stagnation counter. It must be
	// non-negative.
	MinImprovement float64 `json:"min_improvement"`

	// StagnationIterations stops the run after this many consecutive iterations
	// without a sufficient improvement. Zero disables stagnation detection.
	StagnationIterations int `json:"stagnation_iterations"`

	// MinIterations is the minimum number of iterations completed before either
	// stopping criterion can terminate the run. Zero behaves as one because
	// convergence is checked at iteration boundaries.
	MinIterations int `json:"min_iterations"`
}

// NCAuto makes Optimize derive the crossover offspring count from NPop and
// NCRatio instead of taking it literally. It is the default, and it is a
// distinct sentinel rather than the zero value because zero already means
// "produce no offspring" and a caller who asked for that must keep getting it.
const NCAuto = -1

// AquilaWeightAuto makes the AOBLMOA update phase pick each individual's branch
// the way the source paper picks it -- by a deterministic fitness test -- rather
// than by a random draw against Config.AquilaWeight. It is the default.
//
// It is a distinct sentinel rather than the zero value because zero is a
// meaningful probability ("never take an Aquila step"), and a caller who wrote
// it must keep getting it. See effectiveAquilaWeight.
const AquilaWeightAuto = -1.0

// SelectionStrategy names the rule that pairs parents for crossover.
type SelectionStrategy string

const (
	// SelectionTournament draws TournamentSize candidates uniformly from a
	// population and mates the fittest of them. It lets every member reproduce
	// with probability proportional to its rank while still favoring the fit,
	// but it is not the default: see the note in CHANGELOG.md for the Griewank
	// regression it caused at the default population.
	SelectionTournament SelectionStrategy = "tournament"
	// SelectionRank pairs the k-th best male with the k-th best female. It is
	// the rule the algorithm shipped with through v0.4.0, kept because it is
	// the only way to reproduce a run recorded before v0.5.0 and because it is
	// the faithful reading of the original description.
	SelectionRank SelectionStrategy = "rank"
)

// Config holds the configuration parameters for the Mayfly Algorithm.
// When EnableParallel is true, ObjectiveFunc and configured constraint
// functions may be called concurrently with distinct position vectors and
// must be safe for concurrent use.
type Config struct {
	ObjectiveFunc         ObjectiveFunction  `json:"-"`
	Rand                  *rand.Rand         `json:"-"`
	Convergence           *ConvergenceConfig `json:"convergence,omitempty"`
	Constraints           *ConstraintConfig  `json:"constraints,omitempty"`
	CoolingSchedule       string             `json:"cooling_schedule"`
	GravityType           string             `json:"gravity_type"`
	Selection             SelectionStrategy  `json:"selection"`
	ReductionFactor       float64            `json:"reduction_factor"`
	Dance                 float64            `json:"dance"`
	NPop                  int                `json:"npop"`
	NPopF                 int                `json:"npopf"`
	G                     float64            `json:"g"`
	GDamp                 float64            `json:"g_damp"`
	A1                    float64            `json:"a1"`
	A2                    float64            `json:"a2"`
	A3                    float64            `json:"a3"`
	ChaosFactor           float64            `json:"chaos_factor"`
	OrthogonalFactor      float64            `json:"orthogonal_factor"`
	FL                    float64            `json:"fl"`
	DanceDamp             float64            `json:"dance_damp"`
	FLDamp                float64            `json:"fl_damp"`
	NC                    int                `json:"nc"`
	NM                    int                `json:"nm"`
	TournamentSize        int                `json:"tournament_size"`
	NCRatio               float64            `json:"nc_ratio"`
	CrossoverGamma        float64            `json:"crossover_gamma"` // 0/negative/NaN/Inf: DefaultCrossoverGamma
	Mu                    float64            `json:"mu"`
	VelMax                float64            `json:"vel_max"`
	VelMin                float64            `json:"vel_min"`
	EliteCount            int                `json:"elite_count"`
	SearchRange           float64            `json:"search_range"`
	EnlargeFactor         float64            `json:"enlarge_factor"`
	MaxIterations         int                `json:"max_iterations"`
	UpperBound            float64            `json:"upper_bound"`
	Beta                  float64            `json:"beta"`
	LevyAlpha             float64            `json:"levy_alpha"`
	StrategySwitch        int                `json:"strategy_switch"`
	ArchiveSize           int                `json:"archive_size"`
	LevyBeta              float64            `json:"levy_beta"`
	OppositionRate        float64            `json:"opposition_rate"`
	EliteOppositionCount  int                `json:"elite_opposition_count"`
	OppositionProbability float64            `json:"opposition_probability"`
	// Deprecated: the AOBLMOA paper has no such knob. Its update phase moves
	// every individual by Mayfly attraction or by an Aquila strategy, decided
	// by a fitness test, not by chance. Leave this at AquilaWeightAuto for the
	// published algorithm; set a probability in [0, 1] only to approximate the
	// pre-v0.6.0 behavior, which drew the branch at random.
	AquilaWeight         float64 `json:"aquila_weight"`
	MedianWeight         float64 `json:"median_weight"`
	LowerBound           float64 `json:"lower_bound"`
	ProblemSize          int     `json:"problem_size"`
	MaxWorkers           int     `json:"max_workers"`
	GoldenFactor         float64 `json:"golden_factor"`
	InitialTemperature   float64 `json:"initial_temperature"`
	CoolingRate          float64 `json:"cooling_rate"`
	CauchyMutationRate   float64 `json:"cauchy_mutation_rate"`
	UseGSASMA            bool    `json:"use_gsasma"`
	UseWeightedMedian    bool    `json:"use_weighted_median"`
	ApplyOBLToGlobalBest bool    `json:"apply_obl_to_global_best"`
	UseAOBLMOA           bool    `json:"use_aoblmoa"`
	UseMPMA              bool    `json:"use_mpma"`
	UseEOBBMA            bool    `json:"use_eobbma"`
	UseOLCE              bool    `json:"use_olce"`
	UseDESMA             bool    `json:"use_desma"`
	EnableParallel       bool    `json:"enable_parallel"`
}

// TerminationReason describes why an optimization run ended.
type TerminationReason string

const (
	// TerminationMaxIterations means the configured iteration cap was reached.
	TerminationMaxIterations TerminationReason = "maximum_iterations"
	// TerminationTargetCost means the configured target cost was reached.
	TerminationTargetCost TerminationReason = "target_cost"
	// TerminationStagnation means the best cost did not improve sufficiently
	// within the configured stagnation window.
	TerminationStagnation TerminationReason = "stagnation"
)

// Result holds the results of the optimization.
type Result struct {
	// ConvergenceCurve holds the best cost known at the end of each completed
	// iteration, so it has IterationCount entries. It is non-increasing for
	// unconstrained optimization; a constrained incumbent's raw cost may rise
	// when feasibility or lower violation takes priority. Without early
	// stopping, IterationCount equals MaxIterations. It is a history of costs,
	// not a point in the search space.
	//
	// The solution itself is GlobalBest.Position.
	//
	// It replaces the former BestSolution, whose name was the defect: the field
	// has always held this cost history, but reads as a position vector and is
	// a []float64 exactly like one, so using it as a solution compiled, ran and
	// produced nonsense. The rename is deliberately breaking — an alias would
	// have left the misleading name in place.
	ConvergenceCurve []float64

	TerminationReason TerminationReason
	GlobalBest        Best
	FuncEvalCount     int
	IterationCount    int
	Seed              int64 // Random seed used for reproducibility
}

// newMayfly creates an empty mayfly with allocated slices.
func newMayfly(size int) *Mayfly {
	return &Mayfly{
		Position:            make([]float64, size),
		Velocity:            make([]float64, size),
		Cost:                math.Inf(1),
		ConstraintViolation: math.Inf(1),
		Best: Best{
			Position:            make([]float64, size),
			Cost:                math.Inf(1),
			ConstraintViolation: math.Inf(1),
		},
	}
}

// clone creates a deep copy of a mayfly.
func (m *Mayfly) clone() *Mayfly {
	clone := &Mayfly{
		Position:            make([]float64, len(m.Position)),
		Velocity:            make([]float64, len(m.Velocity)),
		Cost:                m.Cost,
		ConstraintViolation: m.ConstraintViolation,
		Best: Best{
			Position:            make([]float64, len(m.Best.Position)),
			Cost:                m.Best.Cost,
			ConstraintViolation: m.Best.ConstraintViolation,
		},
	}
	copy(clone.Position, m.Position)
	copy(clone.Velocity, m.Velocity)
	copy(clone.Best.Position, m.Best.Position)

	return clone
}

// sanitizeVec checks and fixes NaN/Inf values in a vector.
// This prevents numerical issues from heavy-tailed distributions (Lévy, Cauchy)
// and operations that can produce invalid values (log, exp, division by small numbers).
// Invalid values are replaced with random values within bounds.
func sanitizeVec(vec []float64, lowerBound, upperBound float64, rng *rand.Rand) {
	for i := range vec {
		if math.IsNaN(vec[i]) || math.IsInf(vec[i], 0) {
			// Replace invalid value with random value in bounds
			vec[i] = unifrnd(lowerBound, upperBound, rng)
		}
	}
}

// sanitizeCost checks and fixes NaN/Inf cost values.
// Returns a very large finite value if the cost is invalid.
func sanitizeCost(cost float64) float64 {
	if math.IsNaN(cost) || math.IsInf(cost, 1) {
		return 1e100 // Very large but finite penalty
	}

	if math.IsInf(cost, -1) {
		return -1e100 // Very small but finite (for maximization if ever needed)
	}

	return cost
}

// evaluateWithSanitization evaluates the objective function after sanitizing the position.
// This ensures all heavy-tailed operators (Lévy, Cauchy) don't pass invalid values.
//
//nolint:unused // retained for package compatibility and focused helper tests.
func evaluateWithSanitization(objFunc ObjectiveFunction, position []float64,
	lowerBound, upperBound float64, rng *rand.Rand,
) float64 {
	sanitizeVec(position, lowerBound, upperBound, rng)
	return sanitizeCost(objFunc(position))
}
