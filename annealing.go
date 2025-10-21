// Package mayfly - Simulated Annealing Implementation
//
// Implements Simulated Annealing cooling schedules and acceptance criteria
// for the GSASMA variant.
//
// Reference:
// Kirkpatrick, S., Gelatt, C. D., & Vecchi, M. P. (1983). Optimization by
// Simulated Annealing. Science, 220(4598), 671-680.
// DOI: 10.1126/science.220.4598.671
//
// Simulated Annealing is inspired by the annealing process in metallurgy,
// where controlled cooling allows atoms to settle into a low-energy state.
// The Metropolis criterion allows probabilistic acceptance of worse solutions
// to escape local optima.
package mayfly

import (
	"math"
	"math/rand"
)

// AnnealingScheduler manages the temperature schedule for simulated annealing.
type AnnealingScheduler struct {
	ScheduleType       string
	InitialTemperature float64
	CurrentTemperature float64
	CoolingRate        float64
	Iteration          int
}

// NewAnnealingScheduler creates a new annealing scheduler.
// Parameters:
//   - initialTemp: starting temperature (typically 100-1000)
//   - coolingRate: cooling rate (0 < rate < 1, typically 0.8-0.99)
//   - scheduleType: type of cooling schedule ("exponential", "linear", "logarithmic")
func NewAnnealingScheduler(initialTemp, coolingRate float64, scheduleType string) *AnnealingScheduler {
	if scheduleType == "" {
		scheduleType = "exponential"
	}

	return &AnnealingScheduler{
		InitialTemperature: initialTemp,
		CurrentTemperature: initialTemp,
		CoolingRate:        coolingRate,
		ScheduleType:       scheduleType,
		Iteration:          0,
	}
}

// Update updates the temperature according to the cooling schedule.
// This should be called once per iteration.
func (as *AnnealingScheduler) Update() {
	as.Iteration++

	switch as.ScheduleType {
	case "exponential":
		// T(k) = T₀ * α^k
		// Most common schedule, provides fast early cooling and slow late cooling
		as.CurrentTemperature *= as.CoolingRate

	case "linear":
		// T(k) = T₀ - k * α
		// Linear decrease, simpler but less effective
		as.CurrentTemperature = as.InitialTemperature - float64(as.Iteration)*as.CoolingRate
		if as.CurrentTemperature < 0.01 {
			as.CurrentTemperature = 0.01 // Minimum temperature
		}

	case "logarithmic":
		// T(k) = T₀ / (1 + α * log(1 + k))
		// Slowest cooling, best for highly multimodal problems
		as.CurrentTemperature = as.InitialTemperature / (1.0 + as.CoolingRate*math.Log(1.0+float64(as.Iteration)))

	default:
		// Default to exponential
		as.CurrentTemperature *= as.CoolingRate
	}

	// Ensure temperature doesn't go to zero (causes numerical issues)
	if as.CurrentTemperature < 1e-10 {
		as.CurrentTemperature = 1e-10
	}
}

// GetTemperature returns the current temperature.
func (as *AnnealingScheduler) GetTemperature() float64 {
	return as.CurrentTemperature
}

// Reset resets the scheduler to initial temperature.
func (as *AnnealingScheduler) Reset() {
	as.CurrentTemperature = as.InitialTemperature
	as.Iteration = 0
}

// Returns: acceptance probability in [0, 1].
func acceptanceProbability(oldCost, newCost, temperature float64) float64 {
	// If new solution is better, always accept
	if newCost < oldCost {
		return 1.0
	}

	// If new solution is worse, accept with probability exp(-ΔE/T)
	deltaE := newCost - oldCost
	probability := math.Exp(-deltaE / temperature)

	return probability
}

// shouldAccept implements the Metropolis criterion for simulated annealing.
// rng must not be nil (ensured by caller).
// Returns: true if the new solution should be accepted.
func shouldAccept(oldCost, newCost, temperature float64, rng *rand.Rand) bool {

	// Calculate acceptance probability
	prob := acceptanceProbability(oldCost, newCost, temperature)

	// Accept if probability is greater than random number
	return rng.Float64() < prob
}

// Returns: (accepted bool, funcEvals int).
func annealedUpdate(mayfly *Mayfly, candidatePos []float64, temperature float64,
	objectiveFunc ObjectiveFunction, rng *rand.Rand) (bool, int) {
	// Evaluate candidate
	candidateCost := objectiveFunc(candidatePos)
	funcEvals := 1

	// Decide acceptance using Metropolis criterion
	if shouldAccept(mayfly.Cost, candidateCost, temperature, rng) {
		// Accept: update mayfly position and cost
		copy(mayfly.Position, candidatePos)
		mayfly.Cost = candidateCost

		// Update personal best if better
		if candidateCost < mayfly.Best.Cost {
			copy(mayfly.Best.Position, candidatePos)
			mayfly.Best.Cost = candidateCost
		}

		return true, funcEvals
	}

	// Reject: keep current position
	return false, funcEvals
}

// adaptiveTemperatureControl adjusts temperature based on acceptance rate.
// This is an advanced feature that can help maintain exploration when
// acceptance rate drops too low.
//
// Strategy:
//   - If acceptance rate < minRate: increase temperature (reheat)
//   - If acceptance rate > maxRate: decrease temperature faster
//   - Otherwise: use normal cooling schedule
//
// Parameters:
//   - scheduler: annealing scheduler
//   - acceptanceRate: current acceptance rate (accepted/total)
//   - minRate: minimum desired acceptance rate (e.g., 0.1)
//   - maxRate: maximum desired acceptance rate (e.g., 0.9)
//
// This helps prevent premature convergence or excessive wandering.
func adaptiveTemperatureControl(scheduler *AnnealingScheduler, acceptanceRate, minRate, maxRate float64) {
	if acceptanceRate < minRate {
		// Too few acceptances: reheat to increase exploration
		scheduler.CurrentTemperature *= 1.1
		if scheduler.CurrentTemperature > scheduler.InitialTemperature {
			scheduler.CurrentTemperature = scheduler.InitialTemperature
		}
	} else if acceptanceRate > maxRate {
		// Too many acceptances: cool faster for more exploitation
		scheduler.CurrentTemperature *= 0.9
	}
	// Otherwise, maintain current temperature (will cool naturally on next Update)
}

// Returns: true if candidate should be accepted.
func simulatedAnnealingAcceptance(oldCost, newCost float64, scheduler *AnnealingScheduler, rng *rand.Rand) bool {
	return shouldAccept(oldCost, newCost, scheduler.GetTemperature(), rng)
}
