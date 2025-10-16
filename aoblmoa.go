package mayfly

// AOBLMOA - Aquila Optimizer-Based Learning Multi-Objective Algorithm
//
// This file implements the AOBLMOA variant, which hybridizes the Mayfly Algorithm
// with the Aquila Optimizer and adds opposition-based learning.
//
// Key Features:
// 1. Hybrid operator switching between Mayfly and Aquila strategies
// 2. Opposition-based learning for expanded search space
// 3. Multi-objective optimization support (Pareto dominance)
// 4. Adaptive strategy selection based on iteration progress
//
// Reference:
// AOBLMOA: A Hybrid Biomimetic Optimization Algorithm (2023)

// applyAOBLMOA applies the AOBLMOA variant logic to update a mayfly population.
// This function is called during the main optimization loop when UseAOBLMOA is enabled.
//
// AOBLMOA combines:
// - Standard Mayfly position updates (weighted by 1 - AquilaWeight)
// - Aquila Optimizer strategies (weighted by AquilaWeight)
// - Opposition-based learning for select solutions
//
// Parameters:
//   - mayfly: The mayfly to update
//   - globalBest: Current best solution
//   - population: All mayflies (males or females)
//   - isMale: Whether this is a male mayfly
//   - currentIter, maxIter: Iteration progress
//   - config: Algorithm configuration
//
// Returns:
//   - Updated position for the mayfly
func applyAOBLMOA(mayfly *Mayfly, globalBest Best, population []*Mayfly,
	isMale bool, currentIter, maxIter int, config *Config) []float64 {
	// Determine if we should apply Aquila strategy or standard Mayfly update
	useAquilaStrategy := config.Rand.Float64() < config.AquilaWeight

	var newPosition []float64

	if useAquilaStrategy {
		// Use Aquila Optimizer strategy
		strategy := selectAquilaStrategy(currentIter, maxIter, config.Rand)
		newPosition = applyAquilaStrategy(mayfly, globalBest, population,
			strategy, currentIter, maxIter, config)
	} else {
		// Use standard Mayfly update (this will be done by the main loop)
		// Return nil to signal that standard update should be used
		return nil
	}

	// Apply opposition-based learning with probability OppositionProbability
	if config.Rand.Float64() < config.OppositionProbability {
		// Generate opposition point
		oppositionPos := oppositionPoint(newPosition, config.LowerBound, config.UpperBound)

		// Evaluate both positions and keep the better one
		originalCost := config.ObjectiveFunc(newPosition)
		oppositionCost := config.ObjectiveFunc(oppositionPos)

		if oppositionCost < originalCost {
			newPosition = oppositionPos
		}
	}

	return newPosition
}

// 4. Updates positions and evaluates fitness.
func applyAOBLMOAToPopulation(males, females []*Mayfly, globalBest Best,
	currentIter, maxIter int, config *Config) {
	// Update males with AOBLMOA
	for i := 0; i < len(males); i++ {
		newPos := applyAOBLMOA(males[i], globalBest, males, true, currentIter, maxIter, config)

		if newPos != nil {
			// AOBLMOA provided a new position, use it
			copy(males[i].Position, newPos)

			// Clamp to bounds
			maxVec(males[i].Position, config.LowerBound)
			minVec(males[i].Position, config.UpperBound)

			// Evaluate
			males[i].Cost = config.ObjectiveFunc(males[i].Position)

			// Update personal best
			if males[i].Cost < males[i].Best.Cost {
				males[i].Best.Cost = males[i].Cost
				copy(males[i].Best.Position, males[i].Position)
			}
		}
		// If nil, the standard Mayfly update will be used instead
	}

	// Update females with AOBLMOA
	for i := 0; i < len(females); i++ {
		newPos := applyAOBLMOA(females[i], globalBest, females, false, currentIter, maxIter, config)

		if newPos != nil {
			// AOBLMOA provided a new position, use it
			copy(females[i].Position, newPos)

			// Clamp to bounds
			maxVec(females[i].Position, config.LowerBound)
			minVec(females[i].Position, config.UpperBound)

			// Evaluate
			females[i].Cost = config.ObjectiveFunc(females[i].Position)
		}
		// If nil, the standard Mayfly update will be used instead
	}
}

// initializeAOBLMOA initializes AOBLMOA-specific parameters.
// This is called once at the start of optimization.
func initializeAOBLMOA(config *Config) {
	// Set strategy switch point if not already set
	if config.StrategySwitch == 0 {
		config.StrategySwitch = (config.MaxIterations * 2) / 3
	}

	// Ensure opposition probability is in valid range
	if config.OppositionProbability < 0 {
		config.OppositionProbability = 0
	}

	if config.OppositionProbability > 1 {
		config.OppositionProbability = 1
	}

	// Ensure Aquila weight is in valid range
	if config.AquilaWeight < 0 {
		config.AquilaWeight = 0
	}

	if config.AquilaWeight > 1 {
		config.AquilaWeight = 1
	}

	// Ensure archive size is positive
	if config.ArchiveSize <= 0 {
		config.ArchiveSize = 100
	}
}

// ParetoArchive maintains a set of non-dominated solutions for multi-objective problems.
type ParetoArchive struct {
	Solutions []*ParetoSolution
	MaxSize   int
}

// NewParetoArchive creates a new Pareto archive with specified maximum size.
func NewParetoArchive(maxSize int) *ParetoArchive {
	return &ParetoArchive{
		Solutions: make([]*ParetoSolution, 0, maxSize),
		MaxSize:   maxSize,
	}
}

// Add adds a solution to the Pareto archive.
// If the archive is full, it uses NSGA-II selection to maintain diversity.
func (pa *ParetoArchive) Add(solution *ParetoSolution) {
	// Add the new solution
	pa.Solutions = append(pa.Solutions, solution)

	// If archive exceeds max size, select best solutions
	if len(pa.Solutions) > pa.MaxSize {
		pa.Solutions = selectByNSGA2(pa.Solutions, pa.MaxSize)
	}
}

// AddFromMayfly converts a Mayfly to a ParetoSolution and adds it to the archive.
// For single-objective problems, the objective value is just the cost.
func (pa *ParetoArchive) AddFromMayfly(mayfly *Mayfly) {
	solution := &ParetoSolution{
		Position:         make([]float64, len(mayfly.Position)),
		ObjectiveValues:  []float64{mayfly.Cost},
		Rank:             0,
		CrowdingDistance: 0,
	}
	copy(solution.Position, mayfly.Position)
	pa.Add(solution)
}

// GetBestSolution returns the solution with the lowest first objective value.
// This is useful for single-objective optimization.
func (pa *ParetoArchive) GetBestSolution() *ParetoSolution {
	if len(pa.Solutions) == 0 {
		return nil
	}

	best := pa.Solutions[0]
	for _, sol := range pa.Solutions[1:] {
		if sol.ObjectiveValues[0] < best.ObjectiveValues[0] {
			best = sol
		}
	}

	return best
}

// updateParetoArchive updates the Pareto archive with current population.
// This is called at the end of each iteration to maintain the best solutions found.
func updateParetoArchive(archive *ParetoArchive, males, females []*Mayfly) {
	// Add all males
	for _, m := range males {
		archive.AddFromMayfly(m)
	}

	// Add all females
	for _, f := range females {
		archive.AddFromMayfly(f)
	}
}
