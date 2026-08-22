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

// aquilaPosition returns an Aquila-Optimizer candidate position for the mayfly,
// optionally refined by opposition-based learning.
//
// The caller decides whether the individual takes the Aquila branch at all;
// an individual that does not must still receive the ordinary Mayfly velocity
// and position update. See applyAOBLMOAToPopulationWithEvaluator.
//
// It returns the candidate position and the number of objective evaluations it
// consumed (two when opposition-based learning fired, zero otherwise).
func aquilaPosition(
	mayfly *Mayfly,
	globalBest Best,
	population []*Mayfly,
	currentIter, maxIter int,
	config *Config,
	evaluator *constraintEvaluator,
) ([]float64, int) {
	strategy := selectAquilaStrategy(currentIter, maxIter, config.Rand)
	newPosition := applyAquilaStrategy(mayfly, globalBest, population,
		strategy, currentIter, maxIter, config)

	// Apply opposition-based learning with probability OppositionProbability.
	if config.Rand.Float64() >= config.OppositionProbability {
		return newPosition, 0
	}

	oppositionPos := oppositionPoint(newPosition, config.LowerBound, config.UpperBound)

	original := evaluator.evaluate(newPosition, false)
	opposition := evaluator.evaluate(oppositionPos, false)

	if evaluator.better(opposition, original) {
		newPosition = oppositionPos
	}

	return newPosition, 2
}

// applyAOBLMOAToPopulation applies the AOBLMOA variant logic to a whole
// population, using a freshly built evaluator.
func applyAOBLMOAToPopulation(males, females []*Mayfly, globalBest Best,
	currentIter, maxIter int, g, dance, flight float64, config *Config,
) int {
	return applyAOBLMOAToPopulationWithEvaluator(
		males, females, globalBest, currentIter, maxIter, g, dance, flight, config,
		newConstraintEvaluator(config.ObjectiveFunc, nil),
	)
}

// applyAOBLMOAToPopulationWithEvaluator moves and evaluates every male and
// female exactly once.
//
// Each individual either takes an Aquila-Optimizer step (with probability
// config.AquilaWeight) or the ordinary Mayfly velocity and position update.
// Both branches move the individual: nobody is skipped. The Aquila strategy
// itself already switches on iteration progress, matching the paper's
// two-thirds exploration / one-third exploitation split.
//
// It returns the number of objective evaluations consumed.
func applyAOBLMOAToPopulationWithEvaluator(
	males, females []*Mayfly,
	globalBest Best,
	currentIter, maxIter int,
	g, dance, flight float64,
	config *Config,
	evaluator *constraintEvaluator,
) int {
	evaluations := 0

	for _, male := range males {
		if config.Rand.Float64() < config.AquilaWeight {
			newPos, oblEvals := aquilaPosition(
				male, globalBest, males, currentIter, maxIter, config, evaluator,
			)
			evaluations += oblEvals

			copy(male.Position, newPos)
			maxVec(male.Position, config.LowerBound)
			minVec(male.Position, config.UpperBound)
		} else {
			prepareStandardMale(
				male, globalBest, nil, g, dance, g, config, config.Rand, evaluator,
			)
		}

		evaluator.evaluateMayfly(male, false)

		evaluations++

		if evaluator.better(evaluationFromMayfly(male), evaluationFromBest(male.Best)) {
			male.Best.Cost = male.Cost
			male.Best.ConstraintViolation = male.ConstraintViolation
			copy(male.Best.Position, male.Position)
		}
	}

	for i, female := range females {
		if config.Rand.Float64() < config.AquilaWeight {
			newPos, oblEvals := aquilaPosition(
				female, globalBest, females, currentIter, maxIter, config, evaluator,
			)
			evaluations += oblEvals

			copy(female.Position, newPos)
			maxVec(female.Position, config.LowerBound)
			minVec(female.Position, config.UpperBound)
		} else {
			pairedMale := female
			if i < len(males) {
				pairedMale = males[i]
			}

			prepareStandardFemale(female, pairedMale, g, flight, config, config.Rand, evaluator)
		}

		evaluator.evaluateMayfly(female, false)

		evaluations++
	}

	return evaluations
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

// UpdateFromPopulation adds every member of the given populations to the
// archive.
//
// The optimizer does not call this itself. Nothing in the search reads the
// archive back, so maintaining it every iteration only bought NSGA-II pruning
// cost; callers that want a Pareto front build and feed one explicitly.
func (pa *ParetoArchive) UpdateFromPopulation(males, females []*Mayfly) {
	for _, m := range males {
		pa.AddFromMayfly(m)
	}

	for _, f := range females {
		pa.AddFromMayfly(f)
	}
}
