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
// Go implementation by Christian-W. Budde.
package mayfly

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// Optimize runs the Mayfly Optimization Algorithm with the given configuration.
func Optimize(config *Config) (*Result, error) {
	return OptimizeContext(context.Background(), config)
}

// OptimizeContext runs the Mayfly Optimization Algorithm with cancellation,
// optional initial populations, and progress reporting. Cancellation is
// checked while parallel evaluation batches are dispatched and at iteration
// boundaries. In-flight objective calls are allowed to finish before return.
func OptimizeContext(ctx context.Context, config *Config, options ...RunOption) (*Result, error) {
	if ctx == nil {
		return nil, errNilContext
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Validate required parameters
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}

	// A run resolves automatic values and owns its fallback RNG. Keep those
	// mutations off the caller's Config so the same value can safely be reused
	// with different bounds or by another goroutine.
	runConfig := *config
	config = &runConfig

	if config.ObjectiveFunc == nil {
		return nil, errors.New("ObjectiveFunc is required")
	}

	if config.ProblemSize <= 0 {
		return nil, fmt.Errorf("ProblemSize must be positive, got %d", config.ProblemSize)
	}

	// Validate bounds are finite and properly ordered
	if math.IsNaN(config.LowerBound) || math.IsInf(config.LowerBound, 0) {
		return nil, fmt.Errorf("LowerBound must be finite, got %v", config.LowerBound)
	}

	if math.IsNaN(config.UpperBound) || math.IsInf(config.UpperBound, 0) {
		return nil, fmt.Errorf("UpperBound must be finite, got %v", config.UpperBound)
	}

	if config.LowerBound >= config.UpperBound {
		return nil, fmt.Errorf("LowerBound (%v) must be less than UpperBound (%v)",
			config.LowerBound, config.UpperBound)
	}

	if config.MaxIterations <= 0 {
		return nil, fmt.Errorf("MaxIterations must be positive, got %d", config.MaxIterations)
	}

	convergenceErr := validateConvergenceConfig(config.Convergence, config.MaxIterations)
	if convergenceErr != nil {
		return nil, fmt.Errorf("invalid convergence config: %w", convergenceErr)
	}

	constraintErr := validateConstraintConfig(config.Constraints)
	if constraintErr != nil {
		return nil, fmt.Errorf("invalid constraint config: %w", constraintErr)
	}

	// Validate population sizes
	if config.NPop <= 0 {
		return nil, fmt.Errorf("NPop (male population) must be positive, got %d", config.NPop)
	}

	if config.NPopF <= 0 {
		return nil, fmt.Errorf("NPopF (female population) must be positive, got %d", config.NPopF)
	}

	if config.MaxWorkers < 0 {
		return nil, fmt.Errorf("MaxWorkers must be non-negative, got %d", config.MaxWorkers)
	}

	pairingErr := validateFemalePairing(config)
	if pairingErr != nil {
		return nil, pairingErr
	}

	offspringErr := validateOffspring(config)
	if offspringErr != nil {
		return nil, offspringErr
	}

	qmcInitErr := validateQMCInit(config)
	if qmcInitErr != nil {
		return nil, qmcInitErr
	}

	// Validate variant-specific parameters
	if config.UseDESMA {
		if config.SearchRange < 0 {
			return nil, fmt.Errorf("DESMA SearchRange must be non-negative, got %v", config.SearchRange)
		}

		if config.EliteCount < 0 {
			return nil, fmt.Errorf("DESMA EliteCount must be non-negative, got %d", config.EliteCount)
		}

		if config.EnlargeFactor <= 0 {
			return nil, fmt.Errorf("DESMA EnlargeFactor must be positive, got %v", config.EnlargeFactor)
		}

		if config.ReductionFactor <= 0 {
			return nil, fmt.Errorf("DESMA ReductionFactor must be positive, got %v", config.ReductionFactor)
		}
	}

	if config.UseEOBBMA {
		if config.LevyAlpha <= 0 || config.LevyAlpha > 2 {
			return nil, fmt.Errorf("EOBBMA LevyAlpha must be in (0, 2], got %v", config.LevyAlpha)
		}

		if config.LevyBeta <= 0 {
			return nil, fmt.Errorf("EOBBMA LevyBeta must be positive, got %v", config.LevyBeta)
		}

		if config.OppositionRate < 0 || config.OppositionRate > 1 {
			return nil, fmt.Errorf("EOBBMA OppositionRate must be in [0, 1], got %v", config.OppositionRate)
		}
	}

	if config.UseOLCE {
		if config.ProblemSize > MaxOrthogonalArrayDimensions {
			return nil, fmt.Errorf("OLCE ProblemSize must not exceed %d, got %d",
				MaxOrthogonalArrayDimensions, config.ProblemSize)
		}

		if config.OrthogonalFactor < 0 || config.OrthogonalFactor > 1 {
			return nil, fmt.Errorf("OLCE OrthogonalFactor must be in [0, 1], got %v", config.OrthogonalFactor)
		}

		if config.ChaosFactor < 0 || config.ChaosFactor > 1 {
			return nil, fmt.Errorf("OLCE ChaosFactor must be in [0, 1], got %v", config.ChaosFactor)
		}
	}

	if config.UseGSASMA {
		if config.InitialTemperature <= 0 {
			return nil, fmt.Errorf("GSASMA InitialTemperature must be positive, got %v", config.InitialTemperature)
		}

		if config.CoolingRate <= 0 || config.CoolingRate >= 1 {
			return nil, fmt.Errorf("GSASMA CoolingRate must be in (0, 1), got %v", config.CoolingRate)
		}
	}

	if config.UseHMMA {
		if !isFinite(config.HMMAInformationExchange) || config.HMMAInformationExchange <= 0 {
			return nil, fmt.Errorf("HMMA information exchange coefficient must be positive, got %v",
				config.HMMAInformationExchange)
		}

		if !isFinite(config.HMMAScheduleOffset) ||
			config.HMMAScheduleOffset < 0 || config.HMMAScheduleOffset > 1 {
			return nil, fmt.Errorf("HMMA schedule offset must be in [0, 1], got %v",
				config.HMMAScheduleOffset)
		}

		if !isFinite(config.HMMAArtificialMutation) ||
			config.HMMAArtificialMutation < 0 || config.HMMAArtificialMutation > 1 {
			return nil, fmt.Errorf("HMMA artificial mutation coefficient must be in [0, 1], got %v",
				config.HMMAArtificialMutation)
		}
	}

	if config.UseMPMA {
		if config.MedianWeight < 0 || config.MedianWeight > 1 {
			return nil, fmt.Errorf("MPMA MedianWeight must be in [0, 1], got %v", config.MedianWeight)
		}
	}

	if config.UseAOBLMOA {
		if config.AquilaWeight != AquilaWeightAuto &&
			(config.AquilaWeight < 0 || config.AquilaWeight > 1) {
			return nil, fmt.Errorf(
				"AOBLMOA AquilaWeight must be in [0, 1] or AquilaWeightAuto (%v), got %v",
				AquilaWeightAuto, config.AquilaWeight,
			)
		}

		if config.OppositionProbability < 0 || config.OppositionProbability > 1 {
			return nil, fmt.Errorf("AOBLMOA OppositionProbability must be in [0, 1], got %v", config.OppositionProbability)
		}

		// A StrategySwitch at or beyond MaxIterations is deliberately legal:
		// it keeps the run in the Aquila exploration phase throughout.
		if config.StrategySwitch < 0 {
			return nil, fmt.Errorf("AOBLMOA StrategySwitch must be non-negative, got %d", config.StrategySwitch)
		}
	}

	updatePhaseErr := validateUpdatePhaseVariants(config)
	if updatePhaseErr != nil {
		return nil, updatePhaseErr
	}

	if err := ValidateConfig(config); err != nil {
		return nil, err
	}

	run, err := resolveRunOptions(options)
	if err != nil {
		return nil, err
	}

	initialPopulationErr := validateInitialPopulation(config, run)
	if initialPopulationErr != nil {
		return nil, initialPopulationErr
	}

	if config.VelMax == 0 {
		config.VelMax = 0.1 * (config.UpperBound - config.LowerBound)
		config.VelMin = -config.VelMax
	}

	// Initialize the run-local random number generator. A caller-provided
	// *rand.Rand is intentionally reported with no seed: math/rand does not
	// expose the seed used to construct it.
	rng := config.Rand

	var seed *int64

	if rng == nil {
		seedValue := time.Now().UnixNano()
		if config.Seed != nil {
			seedValue = *config.Seed
		}

		seed = &seedValue
		rng = rand.New(rand.NewSource(seedValue))
		// Some variant helpers read Config.Rand directly. This is the run-local
		// copy, not the caller's configuration.
		config.Rand = rng
	}

	logOptimizationStarted(ctx, run.logger, config)

	// The quasi-random block, if one was asked for, comes off the generator
	// before anything else consumes it, so that the population's coverage does
	// not depend on how much randomness the setup above happened to use.
	qmcPositions, qmcErr := quasiRandomPositionsContext(ctx, config, rng)
	if qmcErr != nil {
		return nil, qmcErr
	}

	candidateEvaluator := newConstraintEvaluator(config.ObjectiveFunc, config.Constraints)

	var evaluator *evaluationPool

	if config.EnableParallel {
		largestBatch := largestParallelEvaluationBatch(config)
		workerCount := min(effectiveMaxWorkers(config), largestBatch)

		evaluator = newConstrainedEvaluationPool(candidateEvaluator, workerCount)
		defer evaluator.close()
	}

	// Initialize populations
	males := make([]*Mayfly, config.NPop)
	females := make([]*Mayfly, config.NPopF)

	globalBest := Best{
		Position:            make([]float64, config.ProblemSize),
		Cost:                math.Inf(1),
		ConstraintViolation: math.Inf(1),
	}

	funcCount := 0

	if evaluator != nil {
		for i := range config.NPop {
			contextErr := ctx.Err()
			if contextErr != nil {
				return nil, contextErr
			}

			males[i] = newMayfly(config.ProblemSize)
			if i < len(run.initialMales) {
				copy(males[i].Position, run.initialMales[i])
			} else {
				fillInitialPosition(config, qmcPositions, i, males[i].Position, rng)
			}

			sanitizeVec(males[i].Position, config.LowerBound, config.UpperBound, rng)
		}

		initialBest, evaluationErr := evaluator.evaluate(ctx, males, true, true)
		if evaluationErr != nil {
			return nil, evaluationErr
		}

		funcCount += len(males)
		for _, male := range males {
			copy(male.Best.Position, male.Position)
			male.Best.Cost = male.Cost
			male.Best.ConstraintViolation = male.ConstraintViolation
		}

		mergeBest(&globalBest, initialBest, candidateEvaluator)

		for i := range config.NPopF {
			contextErr := ctx.Err()
			if contextErr != nil {
				return nil, contextErr
			}

			females[i] = newMayfly(config.ProblemSize)
			if i < len(run.initialFemales) {
				copy(females[i].Position, run.initialFemales[i])
			} else {
				fillInitialPosition(config, qmcPositions, config.NPop+i, females[i].Position, rng)
			}

			sanitizeVec(females[i].Position, config.LowerBound, config.UpperBound, rng)
		}

		femaleBest, evaluationErr := evaluator.evaluate(ctx, females, true, true)
		if evaluationErr != nil {
			return nil, evaluationErr
		}

		funcCount += len(females)

		mergeBest(&globalBest, femaleBest, candidateEvaluator)
	} else {
		// Initialize male population sequentially for backward compatibility.
		for i := range config.NPop {
			contextErr := ctx.Err()
			if contextErr != nil {
				return nil, contextErr
			}

			males[i] = newMayfly(config.ProblemSize)
			if i < len(run.initialMales) {
				copy(males[i].Position, run.initialMales[i])
			} else {
				fillInitialPosition(config, qmcPositions, i, males[i].Position, rng)
			}

			sanitizeVec(males[i].Position, config.LowerBound, config.UpperBound, rng)
			candidateEvaluator.evaluateMayfly(males[i], true)

			funcCount++

			copy(males[i].Best.Position, males[i].Position)
			males[i].Best.Cost = males[i].Cost
			males[i].Best.ConstraintViolation = males[i].ConstraintViolation

			if candidateEvaluator.better(evaluationFromBest(males[i].Best), evaluationFromBest(globalBest)) {
				globalBest = cloneBest(males[i].Best)
			}
		}

		// Initialize female population sequentially for backward compatibility.
		for i := range config.NPopF {
			contextErr := ctx.Err()
			if contextErr != nil {
				return nil, contextErr
			}

			females[i] = newMayfly(config.ProblemSize)
			if i < len(run.initialFemales) {
				copy(females[i].Position, run.initialFemales[i])
			} else {
				fillInitialPosition(config, qmcPositions, config.NPop+i, females[i].Position, rng)
			}

			sanitizeVec(females[i].Position, config.LowerBound, config.UpperBound, rng)
			candidateEvaluator.evaluateMayfly(females[i], true)

			funcCount++

			if candidateEvaluator.betterMayflyThanBest(females[i], globalBest) {
				copyMayflyToBest(&globalBest, females[i])
			}
		}
	}

	// Pairing, weighted ranks and several variants assume best-first input on
	// the very first iteration, not only after the first update.
	sortMayflies(males, candidateEvaluator)
	sortMayflies(females, candidateEvaluator)

	if globalBest.Cost == math.MaxFloat64 {
		return nil, ErrNoFiniteObjectiveValue
	}

	bestSolution := make([]float64, config.MaxIterations)
	convergence := newConvergenceTracker(config.Convergence, globalBest, candidateEvaluator)
	iterationCount := 0
	terminationReason := TerminationMaxIterations
	g := config.G
	dance := config.Dance
	fl := config.FL

	// Initialize DESMA parameters if enabled
	var searchRange float64

	var lastGlobalBest Best

	if config.UseDESMA {
		if config.SearchRange == 0 {
			// Auto-calculate initial search range as 10% of the search space
			searchRange = 0.1 * (config.UpperBound - config.LowerBound)
		} else {
			searchRange = config.SearchRange
		}

		lastGlobalBest = cloneBest(globalBest)
	}

	// Initialize OLCE-MA parameters if enabled
	var chaosMap *LogisticMap

	if config.UseOLCE {
		// Initialize chaotic map with random seed
		seed := rng.Float64()
		chaosMap = NewLogisticMap(seed)
	}

	// Initialize GSASMA parameters if enabled
	var annealingScheduler *AnnealingScheduler

	if config.UseGSASMA {
		annealingScheduler = NewAnnealingScheduler(
			config.InitialTemperature,
			config.CoolingRate,
			config.CoolingSchedule,
		)
	}

	previousGSASMACosts := make(map[*Mayfly]float64, len(males)+len(females))
	for _, male := range males {
		previousGSASMACosts[male] = male.Cost
	}

	for _, female := range females {
		previousGSASMACosts[female] = female.Cost
	}

	// Main loop
	for it := range config.MaxIterations {
		iterationErr := ctx.Err()
		if iterationErr != nil {
			return nil, iterationErr
		}

		iterationGravity := g
		if config.UseMPMA {
			iterationGravity = calculateGravityCoefficient(config.GravityType, it+1, config.MaxIterations)
		}

		if config.UseDESMA {
			const (
				desmaGravityMax = 0.9
				desmaGravityMin = 0.4
			)

			progress := float64(it+1) / float64(config.MaxIterations)
			iterationGravity = desmaGravityMax - (desmaGravityMax-desmaGravityMin)*progress
		}

		// AOBLMOA: Use hybrid Mayfly-Aquila updates with opposition-based learning
		switch {
		case config.UseAOBLMOA:
			if evaluator != nil {
				aoblmoaEvals, evaluationErr := evaluateParallelAOBLMOA(
					ctx,
					males,
					females,
					&globalBest,
					it,
					config.MaxIterations,
					g,
					dance,
					fl,
					config,
					rng,
					evaluator,
				)
				if evaluationErr != nil {
					return nil, evaluationErr
				}

				funcCount += aoblmoaEvals
			} else {
				// Apply AOBLMOA to populations sequentially for backward compatibility.
				funcCount += applyAOBLMOAToPopulationWithEvaluator(
					males, females, globalBest, it, config.MaxIterations,
					g, dance, fl, config, candidateEvaluator,
				)
			}

		case config.UseEOBBMA:
			prepareEOBBMAFemales(females, males, config, rng, candidateEvaluator)

			if evaluator != nil {
				_, femaleEvaluationErr := evaluator.evaluate(ctx, females, false, false)
				if femaleEvaluationErr != nil {
					return nil, femaleEvaluationErr
				}
			} else {
				for _, female := range females {
					candidateEvaluator.evaluateMayfly(female, false)
				}
			}

			funcCount += len(females)

			phaseBest := cloneBest(globalBest)
			prepareEOBBMAMales(males, phaseBest, config, rng)

			if evaluator != nil {
				maleBest, maleEvaluationErr := evaluator.evaluate(ctx, males, false, true)
				if maleEvaluationErr != nil {
					return nil, maleEvaluationErr
				}

				mergeBest(&globalBest, maleBest, candidateEvaluator)
			} else {
				for _, male := range males {
					candidateEvaluator.evaluateMayfly(male, false)
				}
			}

			funcCount += len(males)
			updatePersonalBests(males, candidateEvaluator)
		case config.UseGSASMA:
			prepareGSASMAPopulations(
				males, females, cloneBest(globalBest),
				previousGSASMACosts,
				it, config.MaxIterations,
				g, dance, fl, config, annealingScheduler,
				candidateEvaluator, rng,
			)

			if evaluator != nil {
				if _, evaluationErr := evaluator.evaluate(ctx, females, false, false); evaluationErr != nil {
					return nil, evaluationErr
				}

				maleBest, evaluationErr := evaluator.evaluate(ctx, males, false, true)
				if evaluationErr != nil {
					return nil, evaluationErr
				}

				mergeBest(&globalBest, maleBest, candidateEvaluator)
			} else {
				for _, female := range females {
					candidateEvaluator.evaluateMayfly(female, false)
				}

				for _, male := range males {
					candidateEvaluator.evaluateMayfly(male, false)
				}
			}

			funcCount += len(males) + len(females)
			updatePersonalBests(males, candidateEvaluator)
		default:
			// Standard velocity-based updates
			prepareStandardFemales(females, males, iterationGravity, fl, config, rng, candidateEvaluator)

			if evaluator != nil {
				_, femaleEvaluationErr := evaluator.evaluate(ctx, females, false, false)
				if femaleEvaluationErr != nil {
					return nil, femaleEvaluationErr
				}
			} else {
				for _, female := range females {
					candidateEvaluator.evaluateMayfly(female, false)
				}
			}

			funcCount += len(females)

			// MPMA: Calculate median position if enabled
			var medianPos []float64

			mpmaG := iterationGravity

			if config.UseMPMA {
				if config.UseWeightedMedian {
					// Create fitness weights (better fitness = higher weight)
					weights := make([]float64, len(males))
					maxCost := males[len(males)-1].Cost // Worst cost (sorted)
					minCost := males[0].Cost            // Best cost

					for i := range males {
						// Normalize and invert (better solutions get higher weight)
						switch {
						case config.Constraints != nil && len(males) > 1:
							weights[i] = 1.0 - float64(i)/float64(len(males)-1)
						case maxCost > minCost:
							weights[i] = 1.0 - (males[i].Cost-minCost)/(maxCost-minCost)
						default:
							weights[i] = 1.0 // All equal
						}
					}

					if evaluator != nil {
						var medianErr error

						medianPos, medianErr = calculateWeightedMedianPositionParallel(
							ctx,
							males,
							weights,
							effectiveMaxWorkers(config),
						)
						if medianErr != nil {
							return nil, medianErr
						}
					} else {
						medianPos = calculateWeightedMedianPosition(males, weights)
					}
				} else {
					if evaluator != nil {
						var medianErr error

						medianPos, medianErr = calculateMedianPositionParallel(
							ctx,
							males,
							effectiveMaxWorkers(config),
						)
						if medianErr != nil {
							return nil, medianErr
						}
					} else {
						medianPos = calculateMedianPosition(males)
					}
				}
				// Alternative gravity schedules are explicit MPMA extensions.
				mpmaG = calculateGravityCoefficient(config.GravityType, it+1, config.MaxIterations)
			}

			phaseBest := cloneBest(globalBest)
			prepareStandardMales(
				males, phaseBest, medianPos, iterationGravity, dance, mpmaG, config, rng, candidateEvaluator,
			)

			if evaluator != nil {
				maleBest, maleEvaluationErr := evaluator.evaluate(ctx, males, false, true)
				if maleEvaluationErr != nil {
					return nil, maleEvaluationErr
				}

				mergeBest(&globalBest, maleBest, candidateEvaluator)
			} else {
				for _, male := range males {
					candidateEvaluator.evaluateMayfly(male, false)
				}
			}

			funcCount += len(males)
			updatePersonalBests(males, candidateEvaluator)

			// OLCE-MA applies orthogonal learning to the primary male
			// movement operator. It therefore runs for the male population
			// here, before sorting and mating; it is not an elite local-search
			// pass over already selected incumbents.
			if config.UseOLCE && config.OrthogonalFactor > 0 {
				lowerBounds := make([]float64, config.ProblemSize)

				upperBounds := make([]float64, config.ProblemSize)
				for dimension := range config.ProblemSize {
					lowerBounds[dimension] = config.LowerBound
					upperBounds[dimension] = config.UpperBound
				}

				if evaluator != nil {
					orthogonalEvals, evaluationErr := evaluateParallelOrthogonalLearning(
						ctx, males, 1, phaseBest.Position, config.OrthogonalFactor,
						lowerBounds, upperBounds, rng, evaluator,
					)
					if evaluationErr != nil {
						return nil, evaluationErr
					}

					funcCount += orthogonalEvals
				} else {
					applyOrthogonalLearningToElite(
						males, 1, phaseBest.Position, config.OrthogonalFactor,
						lowerBounds, upperBounds, candidateEvaluator, rng,
					)
					funcCount += len(males) * (len(orthogonalArray(config.ProblemSize)) + 1)
				}
			}
		}

		// Females are evaluated candidates too. Historically only males were
		// allowed to update GlobalBest, which could make Optimize return a worse
		// point than one it had already evaluated.
		mergePopulationBest(&globalBest, males, candidateEvaluator)
		mergePopulationBest(&globalBest, females, candidateEvaluator)

		// Sort populations by cost
		sortMayflies(males, candidateEvaluator)
		sortMayflies(females, candidateEvaluator)

		// EOBBMA: Apply elite opposition-based learning
		if config.UseEOBBMA {
			if evaluator != nil {
				oppositionEvals, evaluationErr := evaluateParallelEOBBMAOpposition(
					ctx,
					males,
					&globalBest,
					config,
					rng,
					evaluator,
				)
				if evaluationErr != nil {
					return nil, evaluationErr
				}

				funcCount += oppositionEvals
			} else {
				// Apply opposition sequentially for backward compatibility.
				numEliteOpposition := min(config.EliteOppositionCount, len(males))
				eliteLower, eliteUpper := eliteBounds(
					males, numEliteOpposition, config.LowerBound, config.UpperBound,
				)

				for i := range numEliteOpposition {
					if rng.Float64() >= config.OppositionRate {
						continue
					}

					oppPos := eliteOppositionPoint(
						males[i].Position, eliteLower, eliteUpper,
						config.LowerBound, config.UpperBound, rng,
					)
					opposition := newMayfly(config.ProblemSize)
					copy(opposition.Position, oppPos)
					candidateEvaluator.evaluateMayfly(opposition, false)

					funcCount++

					if candidateEvaluator.betterMayfly(opposition, males[i]) {
						copy(males[i].Position, oppPos)
						males[i].Cost = opposition.Cost
						males[i].ConstraintViolation = opposition.ConstraintViolation

						if candidateEvaluator.better(
							evaluationFromMayfly(males[i]), evaluationFromBest(males[i].Best),
						) {
							copy(males[i].Best.Position, oppPos)
							males[i].Best.Cost = opposition.Cost
							males[i].Best.ConstraintViolation = opposition.ConstraintViolation
						}

						if candidateEvaluator.betterMayflyThanBest(males[i], globalBest) {
							copyMayflyToBest(&globalBest, males[i])
						}
					}
				}
			}

			// Re-sort after opposition learning
			sortMayflies(males, candidateEvaluator)
		}

		// HMMA Eqs. (6)-(11): after the mayfly positions have been
		// updated, mutate the global optimum once through the scheduled
		// OBL/Cauchy cascade and greedily retain the better point.
		if config.UseHMMA {
			globalBest = hmmaGlobalMutation(
				globalBest, it+1, config.MaxIterations, config, candidateEvaluator, rng,
			)
			funcCount++
		}

		// Mating - Create offspring
		var offspring []*Mayfly

		if evaluator != nil {
			parallelOffspring, offspringBest, geneticEvaluations, evaluationErr := evaluateParallelGeneticOperators(
				ctx,
				males,
				females,
				config,
				rng,
				evaluator,
				it,
				chaosMap,
			)
			if evaluationErr != nil {
				return nil, evaluationErr
			}

			offspring = parallelOffspring
			// Not len(offspring): AOBLMOA evaluates an opposition point for
			// every offspring without appending it, so the stage spends more
			// evaluations than it returns individuals.
			funcCount += geneticEvaluations

			mergeBest(&globalBest, offspringBest, candidateEvaluator)
		} else {
			nc := effectiveNC(config)

			maleOffspring := make([]*Mayfly, 0, nc/2+effectiveNM(config))
			femaleOffspring := make([]*Mayfly, 0, nc/2+effectiveNM(config))

			for k := range nc / 2 {
				p1, p2 := selectParents(males, females, k, config, rng)

				// Apply crossover
				off1Pos, off2Pos := crossoverForConfig(
					p1.Position, p2.Position, config, rng,
				)
				if config.UseHMMA {
					off1Pos, off2Pos = hmmaArtificialMutation(
						off1Pos, off2Pos, config.HMMAArtificialMutation,
					)
				}

				// Create offspring 1
				off1 := newMayfly(config.ProblemSize)
				copy(off1.Position, off1Pos)

				candidateEvaluator.evaluateMayfly(off1, false)

				funcCount++

				if !config.UseAOBLMOA &&
					candidateEvaluator.betterMayflyThanBest(off1, globalBest) {
					copyMayflyToBest(&globalBest, off1)
				}

				copy(off1.Best.Position, off1.Position)
				off1.Best.Cost = off1.Cost
				off1.Best.ConstraintViolation = off1.ConstraintViolation

				// Create offspring 2
				off2 := newMayfly(config.ProblemSize)
				copy(off2.Position, off2Pos)

				candidateEvaluator.evaluateMayfly(off2, false)

				funcCount++

				if !config.UseAOBLMOA &&
					candidateEvaluator.betterMayflyThanBest(off2, globalBest) {
					copyMayflyToBest(&globalBest, off2)
				}

				copy(off2.Best.Position, off2.Position)
				off2.Best.Cost = off2.Cost
				off2.Best.ConstraintViolation = off2.ConstraintViolation

				maleOffspring = append(maleOffspring, off1)
				femaleOffspring = append(femaleOffspring, off2)
			}

			offspring = make([]*Mayfly, 0, len(maleOffspring)+len(femaleOffspring))
			offspring = append(offspring, maleOffspring...)
			offspring = append(offspring, femaleOffspring...)

			// The OLCE compatibility stage retains Mayfly's documented
			// Logistic-map equation, but follows the publisher pseudocode's
			// verified all-offspring batch topology. The exact Chebyshev map
			// and component equation remain blocked on primary evidence.
			if config.UseOLCE && config.ChaosFactor > 0 && len(offspring) > 0 {
				for _, child := range offspring {
					applyChaoticExploitation(
						child, config, chaosMap, it, candidateEvaluator,
					)

					funcCount++

					if candidateEvaluator.betterMayflyThanBest(child, globalBest) {
						copyMayflyToBest(&globalBest, child)
					}
				}
			}

			// Offspring refinement: AOBLMOA applies stochastic opposition;
			// HMMA already converted every sibling pair with Eq. (12) and
			// effectiveNM disables extra mutants. Other variants use Gaussian
			// mutation here.
			switch {
			case config.UseAOBLMOA:
				// The paper replaces offspring mutation with stochastic
				// opposition-based learning on every offspring, ungated:
				// Eq. (31) builds the opposition point and Eq. (32) greedily
				// keeps the better of the pair. Config.NM is inert here; see
				// effectiveNM.
				opposites := prepareStochasticOBL(offspring, config, rng)

				for _, opposite := range opposites {
					candidateEvaluator.evaluateMayfly(opposite, false)

					funcCount++
				}

				commitStochasticOBL(offspring, opposites, candidateEvaluator)

				// The global best is merged only now, from the kept
				// offspring. Eq. (32) always keeps the better of the pair, so
				// a discarded candidate could never have been the global best
				// and nothing is lost by waiting.
				for _, child := range offspring {
					if candidateEvaluator.betterMayflyThanBest(child, globalBest) {
						copyMayflyToBest(&globalBest, child)
					}
				}

			default:
				// Standard mutation, independently for the two populations.
				for range effectiveNM(config) {
					maleParent := males[rng.Intn(len(males))]
					femaleParent := females[rng.Intn(len(females))]
					maleMutant := prepareGeneticMutant(
						maleParent, it, config, rng,
					)
					femaleMutant := prepareGeneticMutant(
						femaleParent, it, config, rng,
					)

					candidateEvaluator.evaluateMayfly(maleMutant, false)
					candidateEvaluator.evaluateMayfly(femaleMutant, false)

					funcCount += 2

					initializeOffspringBests([]*Mayfly{maleMutant, femaleMutant})
					mergePopulationBest(&globalBest, []*Mayfly{maleMutant, femaleMutant}, candidateEvaluator)
					maleOffspring = append(maleOffspring, maleMutant)
					femaleOffspring = append(femaleOffspring, femaleMutant)
				}
			}

			if !config.UseAOBLMOA {
				offspring = offspring[:0]
				offspring = append(offspring, maleOffspring...)
				offspring = append(offspring, femaleOffspring...)
			}
		}

		// Merge offspring into populations. Both slices are full-length
		// populations here; the append deliberately grows them before the
		// selection step trims back to NPop/NPopF.
		split := len(offspring) / 2
		males = append(males, offspring[:split]...)     //nolint:makezero // intentional growth, trimmed below
		females = append(females, offspring[split:]...) //nolint:makezero // intentional growth, trimmed below

		// Sort and keep best
		sortMayflies(males, candidateEvaluator)
		sortMayflies(females, candidateEvaluator)

		males = males[:config.NPop]
		females = females[:config.NPopF]

		// DESMA: Apply dynamic elite strategy
		if config.UseDESMA {
			// Dynamically adjust search range based on improvement
			if candidateEvaluator.betterBest(globalBest, lastGlobalBest) {
				// Improving: enlarge search range
				searchRange *= config.EnlargeFactor
			} else {
				// Not improving: reduce search range
				searchRange *= config.ReductionFactor
			}

			var eliteMayfly *Mayfly

			var eliteFuncCount int

			if evaluator != nil {
				var evaluationErr error

				eliteMayfly, eliteFuncCount, evaluationErr = evaluateParallelDESMAElites(
					ctx,
					globalBest,
					searchRange,
					config,
					rng,
					evaluator,
				)
				if evaluationErr != nil {
					return nil, evaluationErr
				}
			} else {
				var improved bool

				eliteMayfly, eliteFuncCount, improved = generateImprovedEliteMayfliesWithEvaluator(
					globalBest,
					searchRange,
					config.EliteCount,
					config.ProblemSize,
					config.LowerBound,
					config.UpperBound,
					candidateEvaluator,
					rng,
				)
				if !improved {
					eliteMayfly = nil
				}
			}

			funcCount += eliteFuncCount

			commitDESMAElite(males, females, &globalBest, eliteMayfly, candidateEvaluator)

			lastGlobalBest = cloneBest(globalBest)
		}

		if config.UseGSASMA {
			previousGSASMACosts = retainGSASMAPreviousCosts(
				previousGSASMACosts, males, females,
			)
		}

		bestSolution[it] = globalBest.Cost
		iterationCount = it + 1

		// GSASMA: Cool only after the annealed second-half phase begins.
		if config.UseGSASMA {
			advanceGSASMATemperature(annealingScheduler, it, config.MaxIterations)
		}

		// Update parameters
		g *= config.GDamp
		dance *= config.DanceDamp
		fl *= config.FLDamp

		notifyProgress(run.observer, it+1, funcCount, globalBest)
		notifyPopulation(run.populationObserver, it+1, funcCount, globalBest, males, females)
		logIterationCompleted(ctx, run.logger, it+1, funcCount, globalBest)

		iterationErr = ctx.Err()
		if iterationErr != nil {
			return nil, iterationErr
		}

		if reason, stop := convergence.observe(iterationCount, globalBest); stop {
			terminationReason = reason

			break
		}
	}

	result := &Result{
		GlobalBest:        globalBest,
		ConvergenceCurve:  bestSolution[:iterationCount],
		TerminationReason: terminationReason,
		FuncEvalCount:     funcCount,
		IterationCount:    iterationCount,
		Seed:              seed,
	}

	logOptimizationCompleted(ctx, run.logger, result)

	return result, nil
}
