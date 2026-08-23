// Simulated Annealing Implementation.
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
	if math.IsNaN(initialTemp) || math.IsInf(initialTemp, 0) || initialTemp <= 0 {
		initialTemp = 1
	}
	if math.IsNaN(coolingRate) || math.IsInf(coolingRate, 0) || coolingRate <= 0 || coolingRate >= 1 {
		coolingRate = 0.95
	}
	switch scheduleType {
	case CoolingExponential, CoolingLinear, CoolingLogarithmic:
	default:
		scheduleType = CoolingExponential
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
	if as == nil {
		return
	}

	as.Iteration++

	switch as.ScheduleType {
	case CoolingExponential:
		// T(k) = T₀ * α^k
		// Most common schedule, provides fast early cooling and slow late cooling
		as.CurrentTemperature *= as.CoolingRate

	case CoolingLinear:
		// CoolingRate is consistently a retention factor across schedules.
		// Linear cooling therefore subtracts T₀*(1-α) per iteration.
		as.CurrentTemperature = as.InitialTemperature *
			(1.0 - float64(as.Iteration)*(1.0-as.CoolingRate))
		if as.CurrentTemperature < 0.01 {
			as.CurrentTemperature = 0.01 // Minimum temperature
		}

	case CoolingLogarithmic:
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
	if as == nil {
		return math.NaN()
	}

	return as.CurrentTemperature
}

// Reset resets the scheduler to initial temperature.
func (as *AnnealingScheduler) Reset() {
	if as == nil {
		return
	}

	as.CurrentTemperature = as.InitialTemperature
	as.Iteration = 0
}

// Returns: acceptance probability in [0, 1].
func acceptanceProbability(oldCost, newCost, temperature float64) float64 {
	if math.IsNaN(oldCost) || math.IsNaN(newCost) {
		return 0
	}

	// If new solution is better, always accept
	if newCost <= oldCost {
		return 1.0
	}
	if math.IsNaN(temperature) || math.IsInf(temperature, 0) || temperature <= 0 {
		return 0
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
	if prob >= 1 {
		return true
	}
	if rng == nil {
		return false
	}

	return rng.Float64() < prob
}

// Returns: (accepted bool, funcEvals int).
//
//nolint:unparam // reserved API returns evaluation counts for future batched updates.
func annealedUpdate(mayfly *Mayfly, candidatePos []float64, temperature float64,
	objectiveFunc ObjectiveFunction, rng *rand.Rand,
) (bool, int) {
	if mayfly == nil || objectiveFunc == nil || len(candidatePos) != len(mayfly.Position) {
		return false, 0
	}

	// Evaluate candidate
	candidateCost := objectiveFunc(candidatePos)
	funcEvals := 1
	if math.IsNaN(candidateCost) || math.IsInf(candidateCost, 0) {
		return false, funcEvals
	}

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
	if scheduler == nil || math.IsNaN(acceptanceRate) || math.IsInf(acceptanceRate, 0) ||
		math.IsNaN(minRate) || math.IsInf(minRate, 0) || math.IsNaN(maxRate) || math.IsInf(maxRate, 0) {
		return
	}
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
	if scheduler == nil {
		return false
	}

	return shouldAccept(oldCost, newCost, scheduler.GetTemperature(), rng)
}
