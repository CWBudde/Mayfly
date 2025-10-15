package mayfly

import (
	"math"
	"sort"
)

// MultiObjectiveFunction represents a multi-objective optimization function.
// It takes a position vector and returns multiple objective values.
type MultiObjectiveFunction func([]float64) []float64

// ParetoSolution represents a solution in the Pareto archive.
type ParetoSolution struct {
	Position          []float64 // Decision variables
	ObjectiveValues   []float64 // Multiple objective values
	Rank              int       // Pareto rank (1 = non-dominated front)
	CrowdingDistance  float64   // Crowding distance for diversity
	DominationCount   int       // Number of solutions that dominate this one
	DominatedSolutions []int    // Indices of solutions dominated by this one
}

// dominates checks if solution a dominates solution b (for minimization).
// Solution a dominates b if:
//   - a is no worse than b in all objectives
//   - a is strictly better than b in at least one objective
//
// For minimization: a[i] <= b[i] for all i, and a[j] < b[j] for at least one j
func dominates(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}

	strictlyBetter := false
	for i := 0; i < len(a); i++ {
		if a[i] > b[i] {
			// a is worse in this objective
			return false
		}
		if a[i] < b[i] {
			// a is strictly better in this objective
			strictlyBetter = true
		}
	}

	return strictlyBetter
}

// fastNonDominatedSort performs fast non-dominated sorting on a population.
// This is a key component of NSGA-II and other multi-objective algorithms.
//
// Returns:
//   - fronts: List of Pareto fronts, where fronts[0] is the first (best) front
//
// Algorithm complexity: O(MN²) where M is number of objectives, N is population size
func fastNonDominatedSort(solutions []*ParetoSolution) [][]int {
	n := len(solutions)
	if n == 0 {
		return nil
	}

	// Initialize domination data
	for i := 0; i < n; i++ {
		solutions[i].DominationCount = 0
		solutions[i].DominatedSolutions = make([]int, 0)
	}

	// First front (non-dominated solutions)
	firstFront := make([]int, 0)

	// Compare all pairs of solutions
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if i == j {
				continue
			}

			if dominates(solutions[i].ObjectiveValues, solutions[j].ObjectiveValues) {
				// i dominates j
				solutions[i].DominatedSolutions = append(solutions[i].DominatedSolutions, j)
			} else if dominates(solutions[j].ObjectiveValues, solutions[i].ObjectiveValues) {
				// j dominates i
				solutions[i].DominationCount++
			}
		}

		// If no one dominates this solution, it's in the first front
		if solutions[i].DominationCount == 0 {
			solutions[i].Rank = 1
			firstFront = append(firstFront, i)
		}
	}

	// Build subsequent fronts
	fronts := make([][]int, 0)
	fronts = append(fronts, firstFront)

	rank := 1
	for len(fronts[rank-1]) > 0 {
		nextFront := make([]int, 0)

		for _, i := range fronts[rank-1] {
			// For each solution dominated by i
			for _, j := range solutions[i].DominatedSolutions {
				solutions[j].DominationCount--

				// If j is now non-dominated in remaining solutions
				if solutions[j].DominationCount == 0 {
					solutions[j].Rank = rank + 1
					nextFront = append(nextFront, j)
				}
			}
		}

		if len(nextFront) > 0 {
			fronts = append(fronts, nextFront)
			rank++
		} else {
			break
		}
	}

	return fronts
}

// calculateCrowdingDistance calculates the crowding distance for solutions in a front.
// Crowding distance measures how close a solution is to its neighbors.
// Higher values indicate more isolated solutions (better diversity).
//
// The crowding distance for solution i is the sum of distances to neighbors
// in each objective, normalized by the objective range.
func calculateCrowdingDistance(solutions []*ParetoSolution, frontIndices []int) {
	frontSize := len(frontIndices)
	if frontSize == 0 {
		return
	}

	// Initialize all crowding distances to 0
	for _, idx := range frontIndices {
		solutions[idx].CrowdingDistance = 0
	}

	// If only 1 or 2 solutions, set to infinity (maximum diversity)
	if frontSize <= 2 {
		for _, idx := range frontIndices {
			solutions[idx].CrowdingDistance = math.Inf(1)
		}
		return
	}

	// Number of objectives
	numObjectives := len(solutions[frontIndices[0]].ObjectiveValues)

	// For each objective
	for m := 0; m < numObjectives; m++ {
		// Sort front by this objective
		sortedIndices := make([]int, frontSize)
		copy(sortedIndices, frontIndices)

		sort.Slice(sortedIndices, func(i, j int) bool {
			return solutions[sortedIndices[i]].ObjectiveValues[m] <
				solutions[sortedIndices[j]].ObjectiveValues[m]
		})

		// Boundary solutions get infinite distance (always select them)
		solutions[sortedIndices[0]].CrowdingDistance = math.Inf(1)
		solutions[sortedIndices[frontSize-1]].CrowdingDistance = math.Inf(1)

		// Calculate objective range
		objMin := solutions[sortedIndices[0]].ObjectiveValues[m]
		objMax := solutions[sortedIndices[frontSize-1]].ObjectiveValues[m]
		objRange := objMax - objMin

		// Avoid division by zero
		if objRange < 1e-10 {
			objRange = 1e-10
		}

		// Calculate crowding distance for intermediate solutions
		for i := 1; i < frontSize-1; i++ {
			if !math.IsInf(solutions[sortedIndices[i]].CrowdingDistance, 1) {
				// Add normalized distance to neighbors
				distance := (solutions[sortedIndices[i+1]].ObjectiveValues[m] -
					solutions[sortedIndices[i-1]].ObjectiveValues[m]) / objRange

				solutions[sortedIndices[i]].CrowdingDistance += distance
			}
		}
	}
}

// crowdingDistanceComparison compares two solutions based on crowding distance.
// Returns true if solution a is preferred over solution b.
// Preference order:
//  1. Lower rank (better Pareto front)
//  2. Higher crowding distance (more diversity)
func crowdingDistanceComparison(a, b *ParetoSolution) bool {
	if a.Rank < b.Rank {
		return true
	}
	if a.Rank > b.Rank {
		return false
	}

	// Same rank: prefer higher crowding distance
	return a.CrowdingDistance > b.CrowdingDistance
}

// calculateHypervolume calculates the hypervolume indicator for a Pareto front.
// The hypervolume is the volume of the objective space dominated by the front.
// Higher values indicate better convergence and diversity.
//
// This is a simplified 2D implementation. For 3+ objectives, more sophisticated
// algorithms like WFG or FPRAS should be used.
//
// Parameters:
//   - solutions: Solutions in the Pareto front
//   - referencePoint: Worst acceptable point (usually slightly worse than nadir point)
//
// Note: Currently only supports 2 objectives for simplicity.
func calculateHypervolume(solutions []*ParetoSolution, referencePoint []float64) float64 {
	if len(solutions) == 0 {
		return 0
	}

	numObjectives := len(solutions[0].ObjectiveValues)
	if numObjectives != 2 {
		// For now, only 2D hypervolume is supported
		// For higher dimensions, need more complex algorithms
		return 0
	}

	// Sort solutions by first objective
	sorted := make([]*ParetoSolution, len(solutions))
	copy(sorted, solutions)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ObjectiveValues[0] < sorted[j].ObjectiveValues[0]
	})

	hypervolume := 0.0
	previousY := referencePoint[1]

	for _, sol := range sorted {
		width := referencePoint[0] - sol.ObjectiveValues[0]
		height := previousY - sol.ObjectiveValues[1]

		if width > 0 && height > 0 {
			hypervolume += width * height
		}

		// Update for next iteration
		if sol.ObjectiveValues[1] < previousY {
			previousY = sol.ObjectiveValues[1]
		}
	}

	return hypervolume
}

// calculateIGD calculates the Inverted Generational Distance (IGD) metric.
// IGD measures both convergence and diversity by computing the average
// distance from each point in the true Pareto front to the nearest
// point in the obtained front.
//
// Lower values indicate better performance (closer to true Pareto front).
//
// Parameters:
//   - obtainedFront: The Pareto front obtained by the algorithm
//   - trueFront: The true/reference Pareto front (if known)
//
// Returns:
//   - IGD value (lower is better)
func calculateIGD(obtainedFront, trueFront []*ParetoSolution) float64 {
	if len(trueFront) == 0 || len(obtainedFront) == 0 {
		return math.Inf(1)
	}

	totalDistance := 0.0

	// For each point in the true front
	for _, truePoint := range trueFront {
		// Find minimum distance to obtained front
		minDistance := math.Inf(1)

		for _, obtainedPoint := range obtainedFront {
			// Calculate Euclidean distance in objective space
			distance := 0.0
			for i := 0; i < len(truePoint.ObjectiveValues); i++ {
				diff := truePoint.ObjectiveValues[i] - obtainedPoint.ObjectiveValues[i]
				distance += diff * diff
			}
			distance = math.Sqrt(distance)

			if distance < minDistance {
				minDistance = distance
			}
		}

		totalDistance += minDistance
	}

	// Average distance
	return totalDistance / float64(len(trueFront))
}

// selectByNSGA2 selects the best N solutions using NSGA-II selection.
// This combines Pareto ranking and crowding distance.
//
// Parameters:
//   - solutions: Pool of solutions to select from
//   - n: Number of solutions to select
//
// Returns:
//   - Selected solutions (size n)
func selectByNSGA2(solutions []*ParetoSolution, n int) []*ParetoSolution {
	if len(solutions) <= n {
		return solutions
	}

	// Perform non-dominated sorting
	fronts := fastNonDominatedSort(solutions)

	// Calculate crowding distance for each front
	for _, front := range fronts {
		calculateCrowdingDistance(solutions, front)
	}

	// Select solutions front by front
	selected := make([]*ParetoSolution, 0, n)

	for _, front := range fronts {
		if len(selected)+len(front) <= n {
			// Add entire front
			for _, idx := range front {
				selected = append(selected, solutions[idx])
			}
		} else {
			// Need to select some solutions from this front
			remaining := n - len(selected)

			// Sort front by crowding distance (descending)
			sortedFront := make([]*ParetoSolution, len(front))
			for i, idx := range front {
				sortedFront[i] = solutions[idx]
			}

			sort.Slice(sortedFront, func(i, j int) bool {
				return sortedFront[i].CrowdingDistance > sortedFront[j].CrowdingDistance
			})

			// Add most diverse solutions
			for i := 0; i < remaining; i++ {
				selected = append(selected, sortedFront[i])
			}

			break
		}
	}

	return selected
}
