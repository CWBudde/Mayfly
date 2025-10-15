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
	Cost     float64
	Best     Best
}

// Config holds the configuration parameters for the Mayfly Algorithm.
type Config struct {
	// Problem parameters
	ObjectiveFunc ObjectiveFunction `json:"-"` // Cannot serialize functions
	ProblemSize   int               `json:"problem_size"`
	LowerBound    float64           `json:"lower_bound"`
	UpperBound    float64           `json:"upper_bound"`

	// Algorithm parameters
	MaxIterations int     `json:"max_iterations"`
	NPop          int     `json:"npop"`
	NPopF         int     `json:"npopf"`
	G             float64 `json:"g"`
	GDamp         float64 `json:"g_damp"`
	A1            float64 `json:"a1"`
	A2            float64 `json:"a2"`
	A3            float64 `json:"a3"`
	Beta          float64 `json:"beta"`
	Dance         float64 `json:"dance"`
	FL            float64 `json:"fl"`
	DanceDamp     float64 `json:"dance_damp"`
	FLDamp        float64 `json:"fl_damp"`

	// Mating parameters
	NC int     `json:"nc"`
	NM int     `json:"nm"`
	Mu float64 `json:"mu"`

	// Velocity limits
	VelMax float64 `json:"vel_max"`
	VelMin float64 `json:"vel_min"`

	// Random source (optional, will use default if nil)
	Rand *rand.Rand `json:"-"` // Cannot serialize

	// DESMA (Dynamic Elite Strategy) parameters
	UseDESMA        bool    `json:"use_desma"`
	EliteCount      int     `json:"elite_count"`
	SearchRange     float64 `json:"search_range"`
	EnlargeFactor   float64 `json:"enlarge_factor"`
	ReductionFactor float64 `json:"reduction_factor"`

	// OLCE-MA (Orthogonal Learning and Chaotic Exploitation) parameters
	UseOLCE          bool    `json:"use_olce"`
	OrthogonalFactor float64 `json:"orthogonal_factor"`
	ChaosFactor      float64 `json:"chaos_factor"`

	// EOBBMA (Elite Opposition-Based Bare Bones Mayfly Algorithm) parameters
	UseEOBBMA            bool    `json:"use_eobbma"`
	LevyAlpha            float64 `json:"levy_alpha"`
	LevyBeta             float64 `json:"levy_beta"`
	OppositionRate       float64 `json:"opposition_rate"`
	EliteOppositionCount int     `json:"elite_opposition_count"`

	// MPMA (Median Position-Based Mayfly Algorithm) parameters
	UseMPMA           bool    `json:"use_mpma"`
	MedianWeight      float64 `json:"median_weight"`
	GravityType       string  `json:"gravity_type"`
	UseWeightedMedian bool    `json:"use_weighted_median"`

	// GSASMA (Golden Sine Algorithm with Simulated Annealing MA) parameters
	UseGSASMA            bool    `json:"use_gsasma"`
	InitialTemperature   float64 `json:"initial_temperature"`
	CoolingRate          float64 `json:"cooling_rate"`
	CauchyMutationRate   float64 `json:"cauchy_mutation_rate"`
	GoldenFactor         float64 `json:"golden_factor"`
	CoolingSchedule      string  `json:"cooling_schedule"`
	ApplyOBLToGlobalBest bool    `json:"apply_obl_to_global_best"`

	// AOBLMOA (Aquila Optimizer-Based Learning Multi-Objective Algorithm) parameters
	UseAOBLMOA            bool    `json:"use_aoblmoa"`
	AquilaWeight          float64 `json:"aquila_weight"`
	OppositionProbability float64 `json:"opposition_probability"`
	ArchiveSize           int     `json:"archive_size"`
	StrategySwitch        int     `json:"strategy_switch"`
}

// Result holds the results of the optimization.
type Result struct {
	GlobalBest     Best
	BestSolution   []float64 // Best cost at each iteration
	FuncEvalCount  int
	IterationCount int
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
