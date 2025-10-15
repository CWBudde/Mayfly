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
	ObjectiveFunc ObjectiveFunction
	ProblemSize   int     // Number of decision variables
	LowerBound    float64 // Lower bound for decision variables
	UpperBound    float64 // Upper bound for decision variables

	// Algorithm parameters
	MaxIterations int     // Maximum number of iterations
	NPop          int     // Population size for males
	NPopF         int     // Population size for females
	G             float64 // Inertia weight
	GDamp         float64 // Inertia weight damping ratio
	A1            float64 // Personal learning coefficient
	A2            float64 // Global learning coefficient (males)
	A3            float64 // Global learning coefficient (females)
	Beta          float64 // Distance sight coefficient
	Dance         float64 // Nuptial dance coefficient
	FL            float64 // Random flight coefficient
	DanceDamp     float64 // Dance damping ratio
	FLDamp        float64 // Flight damping ratio

	// Mating parameters
	NC int     // Number of offspring
	NM int     // Number of mutants
	Mu float64 // Mutation rate

	// Velocity limits
	VelMax float64
	VelMin float64

	// Random source (optional, will use default if nil)
	Rand *rand.Rand

	// DESMA (Dynamic Elite Strategy) parameters
	UseDESMA        bool    // Enable DESMA variant
	EliteCount      int     // Number of elite mayflies to generate (default: 5)
	SearchRange     float64 // Initial search range for elite generation (default: auto-calculated)
	EnlargeFactor   float64 // Factor to enlarge search range when improving (default: 1.05)
	ReductionFactor float64 // Factor to reduce search range when not improving (default: 0.95)

	// OLCE-MA (Orthogonal Learning and Chaotic Exploitation) parameters
	UseOLCE          bool    // Enable OLCE-MA variant
	OrthogonalFactor float64 // Orthogonal learning strength (default: 0.3)
	ChaosFactor      float64 // Chaos perturbation strength for offspring (default: 0.1)

	// EOBBMA (Elite Opposition-Based Bare Bones Mayfly Algorithm) parameters
	UseEOBBMA            bool    // Enable EOBBMA variant
	LevyAlpha            float64 // Lévy stability parameter (default: 1.5, range: 0 < alpha <= 2)
	LevyBeta             float64 // Lévy scale parameter (default: 1.0)
	OppositionRate       float64 // Probability of opposition learning (default: 0.3)
	EliteOppositionCount int     // Number of elite solutions to apply opposition (default: 3)

	// MPMA (Median Position-Based Mayfly Algorithm) parameters
	UseMPMA           bool    // Enable MPMA variant
	MedianWeight      float64 // Influence of median position on velocity (default: 0.5)
	GravityType       string  // Type of gravity coefficient: "linear", "exponential", "sigmoid" (default: "linear")
	UseWeightedMedian bool    // Use fitness-weighted median (default: false)

	// GSASMA (Golden Sine Algorithm with Simulated Annealing MA) parameters
	UseGSASMA            bool    // Enable GSASMA variant
	InitialTemperature   float64 // Starting temperature for simulated annealing (default: 100)
	CoolingRate          float64 // Temperature decay rate (default: 0.95)
	CauchyMutationRate   float64 // Probability of Cauchy mutation vs Gaussian (default: 0.3)
	GoldenFactor         float64 // Golden sine influence factor (default: 1.0)
	CoolingSchedule      string  // Temperature schedule: "exponential", "linear", "logarithmic" (default: "exponential")
	ApplyOBLToGlobalBest bool    // Apply opposition-based learning to global best (default: true)

	// AOBLMOA (Aquila Optimizer-Based Learning Multi-Objective Algorithm) parameters
	UseAOBLMOA           bool    // Enable AOBLMOA variant
	AquilaWeight         float64 // Weight for Aquila strategy influence (default: 0.5)
	OppositionProbability float64 // Probability of applying opposition-based learning (default: 0.3)
	ArchiveSize          int     // Maximum size of Pareto archive for multi-objective (default: 100)
	StrategySwitch       int     // Iteration threshold for switching strategies (default: MaxIterations * 2/3)
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
