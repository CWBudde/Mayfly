// Package mayfly implements the Mayfly Optimization Algorithm (MA).
//
// Developers: K. Zervoudakis & S. Tsafarakis
//
// Contact Info: kzervoudakis@isc.tuc.gr
//
//	School of Production Engineering and Management,
//	Technical University of Crete, Chania, Greece
//
// Please cite as:
// Zervoudakis, K., & Tsafarakis, S. (2020). A mayfly optimization algorithm.
// Computers & Industrial Engineering, 145, 106559.
// https://doi.org/10.1016/j.cie.2020.106559
//
// Go implementation by [Your Name]
package mayfly

import (
	"math"
	"math/rand"
)

// ObjectiveFunction represents a function to be optimized.
// It takes a position vector and returns a fitness cost.
type ObjectiveFunction func([]float64) float64

// Best represents the best position and cost found.
type Best struct {
	Position []float64
	Cost     float64
}

// Mayfly represents a single mayfly (male or female) in the population.
type Mayfly struct {
	Position []float64
	Velocity []float64
	Best     Best
	Cost     float64
}

// Config holds the configuration parameters for the Mayfly Algorithm.
type Config struct {
	ObjectiveFunc         ObjectiveFunction `json:"-"`
	Rand                  *rand.Rand        `json:"-"`
	CoolingSchedule       string            `json:"cooling_schedule"`
	GravityType           string            `json:"gravity_type"`
	ReductionFactor       float64           `json:"reduction_factor"`
	Dance                 float64           `json:"dance"`
	NPop                  int               `json:"npop"`
	NPopF                 int               `json:"npopf"`
	G                     float64           `json:"g"`
	GDamp                 float64           `json:"g_damp"`
	A1                    float64           `json:"a1"`
	A2                    float64           `json:"a2"`
	A3                    float64           `json:"a3"`
	ChaosFactor           float64           `json:"chaos_factor"`
	OrthogonalFactor      float64           `json:"orthogonal_factor"`
	FL                    float64           `json:"fl"`
	DanceDamp             float64           `json:"dance_damp"`
	FLDamp                float64           `json:"fl_damp"`
	NC                    int               `json:"nc"`
	NM                    int               `json:"nm"`
	Mu                    float64           `json:"mu"`
	VelMax                float64           `json:"vel_max"`
	VelMin                float64           `json:"vel_min"`
	EliteCount            int               `json:"elite_count"`
	SearchRange           float64           `json:"search_range"`
	EnlargeFactor         float64           `json:"enlarge_factor"`
	MaxIterations         int               `json:"max_iterations"`
	UpperBound            float64           `json:"upper_bound"`
	Beta                  float64           `json:"beta"`
	LevyAlpha             float64           `json:"levy_alpha"`
	StrategySwitch        int               `json:"strategy_switch"`
	ArchiveSize           int               `json:"archive_size"`
	LevyBeta              float64           `json:"levy_beta"`
	OppositionRate        float64           `json:"opposition_rate"`
	EliteOppositionCount  int               `json:"elite_opposition_count"`
	OppositionProbability float64           `json:"opposition_probability"`
	AquilaWeight          float64           `json:"aquila_weight"`
	MedianWeight          float64           `json:"median_weight"`
	LowerBound            float64           `json:"lower_bound"`
	ProblemSize           int               `json:"problem_size"`
	GoldenFactor          float64           `json:"golden_factor"`
	InitialTemperature    float64           `json:"initial_temperature"`
	CoolingRate           float64           `json:"cooling_rate"`
	CauchyMutationRate    float64           `json:"cauchy_mutation_rate"`
	UseGSASMA             bool              `json:"use_gsasma"`
	UseWeightedMedian     bool              `json:"use_weighted_median"`
	ApplyOBLToGlobalBest  bool              `json:"apply_obl_to_global_best"`
	UseAOBLMOA            bool              `json:"use_aoblmoa"`
	UseMPMA               bool              `json:"use_mpma"`
	UseEOBBMA             bool              `json:"use_eobbma"`
	UseOLCE               bool              `json:"use_olce"`
	UseDESMA              bool              `json:"use_desma"`
}

// Result holds the results of the optimization.
type Result struct {
	BestSolution   []float64
	GlobalBest     Best
	FuncEvalCount  int
	IterationCount int
	Seed           int64 // Random seed used for reproducibility
}

// newMayfly creates an empty mayfly with allocated slices.
func newMayfly(size int) *Mayfly {
	return &Mayfly{
		Position: make([]float64, size),
		Velocity: make([]float64, size),
		Cost:     math.Inf(1),
		Best: Best{
			Position: make([]float64, size),
			Cost:     math.Inf(1),
		},
	}
}

// clone creates a deep copy of a mayfly.
func (m *Mayfly) clone() *Mayfly {
	clone := &Mayfly{
		Position: make([]float64, len(m.Position)),
		Velocity: make([]float64, len(m.Velocity)),
		Cost:     m.Cost,
		Best: Best{
			Position: make([]float64, len(m.Best.Position)),
			Cost:     m.Best.Cost,
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
func evaluateWithSanitization(objFunc ObjectiveFunction, position []float64,
	lowerBound, upperBound float64, rng *rand.Rand) float64 {
	sanitizeVec(position, lowerBound, upperBound, rng)
	return sanitizeCost(objFunc(position))
}
