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

	// Validate required parameters
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}

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

	offspringErr := validateOffspring(config)
	if offspringErr != nil {
		return nil, offspringErr
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
		if config.OrthogonalFactor < 0 || config.OrthogonalFactor > 1 {
			return nil, fmt.Errorf("OLCE OrthogonalFactor must be in [0, 1], got %v", config.OrthogonalFactor)
		}

		if config.ChaosFactor < 0 {
			return nil, fmt.Errorf("OLCE ChaosFactor must be non-negative, got %v", config.ChaosFactor)
		}
	}

	if config.UseGSASMA {
		if config.InitialTemperature <= 0 {
			return nil, fmt.Errorf("GSASMA InitialTemperature must be positive, got %v", config.InitialTemperature)
		}

		if config.CoolingRate <= 0 || config.CoolingRate >= 1 {
			return nil, fmt.Errorf("GSASMA CoolingRate must be in (0, 1), got %v", config.CoolingRate)
		}

		if config.CauchyMutationRate < 0 || config.CauchyMutationRate > 1 {
			return nil, fmt.Errorf("GSASMA CauchyMutationRate must be in [0, 1], got %v", config.CauchyMutationRate)
		}
	}

	if config.UseMPMA {
		if config.MedianWeight < 0 || config.MedianWeight > 1 {
			return nil, fmt.Errorf("MPMA MedianWeight must be in [0, 1], got %v", config.MedianWeight)
		}
	}

	if config.UseAOBLMOA {
		if config.AquilaWeight < 0 || config.AquilaWeight > 1 {
			return nil, fmt.Errorf("AOBLMOA AquilaWeight must be in [0, 1], got %v", config.AquilaWeight)
		}

		if config.OppositionProbability < 0 || config.OppositionProbability > 1 {
			return nil, fmt.Errorf("AOBLMOA OppositionProbability must be in [0, 1], got %v", config.OppositionProbability)
		}
	}

	run, err := resolveRunOptions(options)
	if err != nil {
		return nil, err
	}

	initialPopulationErr := validateInitialPopulation(config, run)
	if initialPopulationErr != nil {
		return nil, initialPopulationErr
	}

	logOptimizationStarted(ctx, run.logger, config)

	// Initialize parameters
	if config.NM == 0 {
		config.NM = int(math.Round(0.05 * float64(config.NPop)))
	}

	if config.VelMax == 0 {
		config.VelMax = 0.1 * (config.UpperBound - config.LowerBound)
		config.VelMin = -config.VelMax
	}

	// Initialize random number generator if not provided
	rng := config.Rand
	// The seed is only tracked for reporting; a caller-provided *rand.Rand does
	// not expose its seed, so we record the time-based fallback in that case.
	seed := time.Now().UnixNano()

	if rng == nil {
		rng = rand.New(rand.NewSource(seed))
		// Share the fallback generator with the helpers that read config.Rand
		// directly, such as the sequential AOBLMOA path.
		config.Rand = rng
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
				males[i].Position = unifrndVec(config.LowerBound, config.UpperBound, config.ProblemSize, rng)
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
				females[i].Position = unifrndVec(config.LowerBound, config.UpperBound, config.ProblemSize, rng)
			}

			sanitizeVec(females[i].Position, config.LowerBound, config.UpperBound, rng)
		}

		_, evaluationErr = evaluator.evaluate(ctx, females, true, false)
		if evaluationErr != nil {
			return nil, evaluationErr
		}

		funcCount += len(females)
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
				males[i].Position = unifrndVec(config.LowerBound, config.UpperBound, config.ProblemSize, rng)
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
				females[i].Position = unifrndVec(config.LowerBound, config.UpperBound, config.ProblemSize, rng)
			}

			sanitizeVec(females[i].Position, config.LowerBound, config.UpperBound, rng)
			candidateEvaluator.evaluateMayfly(females[i], true)

			funcCount++
		}
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
	var (
		annealingScheduler  *AnnealingScheduler
		goldenSectionSearch *goldenSection
	)

	if config.UseGSASMA {
		annealingScheduler = NewAnnealingScheduler(
			config.InitialTemperature,
			config.CoolingRate,
			config.CoolingSchedule,
		)
		// The golden section search persists across iterations so that the
		// interval actually narrows over the run.
		goldenSectionSearch = newGoldenSection()
	}

	// Initialize AOBLMOA parameters if enabled
	if config.UseAOBLMOA {
		initializeAOBLMOA(config)
	}

	// Main loop
	for it := range config.MaxIterations {
		iterationErr := ctx.Err()
		if iterationErr != nil {
			return nil, iterationErr
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

			// Update global best from updated populations
			for i := range config.NPop {
				if candidateEvaluator.betterMayflyThanBest(males[i], globalBest) {
					copyMayflyToBest(&globalBest, males[i])
				}
			}
		case config.UseEOBBMA:
			if evaluator != nil {
				prepareEOBBMAFemales(females, males, config, rng)

				_, femaleEvaluationErr := evaluator.evaluate(ctx, females, false, false)
				if femaleEvaluationErr != nil {
					return nil, femaleEvaluationErr
				}

				funcCount += len(females)
				phaseBest := cloneBest(globalBest)
				prepareEOBBMAMales(males, phaseBest, config, rng)

				maleBest, maleEvaluationErr := evaluator.evaluate(ctx, males, false, true)
				if maleEvaluationErr != nil {
					return nil, maleEvaluationErr
				}

				funcCount += len(males)
				updatePersonalBests(males, candidateEvaluator)
				mergeBest(&globalBest, maleBest, candidateEvaluator)
			} else {
				// Update females with Gaussian sampling around best males.
				for i := range config.NPopF {
					if rng.Float64() < 0.5 {
						newPos := gaussianUpdate(females[i].Position, males[i].Position,
							config.LowerBound, config.UpperBound, rng)
						copy(females[i].Position, newPos)
					} else {
						levyStep := levyFlightVec(config.ProblemSize, config.LevyAlpha, config.LevyBeta, rng)
						for j := range config.ProblemSize {
							females[i].Position[j] += levyStep[j] * (config.UpperBound - config.LowerBound) * 0.01
						}

						maxVec(females[i].Position, config.LowerBound)
						minVec(females[i].Position, config.UpperBound)
					}

					candidateEvaluator.evaluateMayfly(females[i], false)

					funcCount++
				}

				// Update males with Gaussian sampling around personal and global best.
				for i := range config.NPop {
					if rng.Float64() < 0.5 {
						newPos := gaussianUpdate(males[i].Position, males[i].Best.Position,
							config.LowerBound, config.UpperBound, rng)
						copy(males[i].Position, newPos)
					} else {
						newPos := gaussianUpdate(males[i].Position, globalBest.Position,
							config.LowerBound, config.UpperBound, rng)
						copy(males[i].Position, newPos)
					}

					candidateEvaluator.evaluateMayfly(males[i], false)

					funcCount++

					if candidateEvaluator.better(
						evaluationFromMayfly(males[i]), evaluationFromBest(males[i].Best),
					) {
						copy(males[i].Best.Position, males[i].Position)
						males[i].Best.Cost = males[i].Cost
						males[i].Best.ConstraintViolation = males[i].ConstraintViolation

						if candidateEvaluator.better(
							evaluationFromBest(males[i].Best), evaluationFromBest(globalBest),
						) {
							globalBest = cloneBest(males[i].Best)
						}
					}
				}
			}
		default:
			// Standard velocity-based updates
			if evaluator != nil {
				prepareStandardFemales(females, males, g, fl, config, rng, candidateEvaluator)

				_, femaleEvaluationErr := evaluator.evaluate(ctx, females, false, false)
				if femaleEvaluationErr != nil {
					return nil, femaleEvaluationErr
				}

				funcCount += len(females)
			} else {
				// Update females sequentially for backward compatibility.
				for i := range config.NPopF {
					e := unifrndVec(-1, 1, config.ProblemSize, rng)

					if candidateEvaluator.betterMayfly(males[i], females[i]) {
						for j := range config.ProblemSize {
							rmf := males[i].Position[j] - females[i].Position[j]
							females[i].Velocity[j] = g*females[i].Velocity[j] +
								config.A3*math.Exp(-config.Beta*rmf*rmf)*(males[i].Position[j]-females[i].Position[j])
						}
					} else {
						for j := range config.ProblemSize {
							females[i].Velocity[j] = g*females[i].Velocity[j] + fl*e[j]
						}
					}

					maxVec(females[i].Velocity, config.VelMin)
					minVec(females[i].Velocity, config.VelMax)

					for j := range config.ProblemSize {
						females[i].Position[j] += females[i].Velocity[j]
					}

					maxVec(females[i].Position, config.LowerBound)
					minVec(females[i].Position, config.UpperBound)
					candidateEvaluator.evaluateMayfly(females[i], false)

					funcCount++
				}
			}

			// MPMA: Calculate median position if enabled
			var medianPos []float64

			var mpmaG float64 // MPMA-specific gravity coefficient

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
				// Calculate non-linear gravity coefficient
				mpmaG = calculateGravityCoefficient(config.GravityType, it, config.MaxIterations)
			}

			if evaluator != nil {
				phaseBest := cloneBest(globalBest)
				prepareStandardMales(
					males, phaseBest, medianPos, g, dance, mpmaG, config, rng, candidateEvaluator,
				)

				maleBest, maleEvaluationErr := evaluator.evaluate(ctx, males, false, true)
				if maleEvaluationErr != nil {
					return nil, maleEvaluationErr
				}

				funcCount += len(males)
				updatePersonalBests(males, candidateEvaluator)
				mergeBest(&globalBest, maleBest, candidateEvaluator)
			} else {
				// Update males sequentially for backward compatibility.
				for i := range config.NPop {
					e := unifrndVec(-1, 1, config.ProblemSize, rng)

					if candidateEvaluator.better(
						evaluationFromBest(globalBest), evaluationFromMayfly(males[i]),
					) {
						// Update velocity with personal and global best
						if config.UseMPMA {
							// MPMA: Include median position in velocity update
							for j := range config.ProblemSize {
								rpbest := males[i].Best.Position[j] - males[i].Position[j]
								rgbest := globalBest.Position[j] - males[i].Position[j]
								rmedian := medianPos[j] - males[i].Position[j]

								// Modified velocity update with median position and non-linear gravity
								males[i].Velocity[j] = mpmaG*males[i].Velocity[j] +
									config.A1*math.Exp(-config.Beta*rpbest*rpbest)*(males[i].Best.Position[j]-males[i].Position[j]) +
									config.A2*math.Exp(-config.Beta*rgbest*rgbest)*(globalBest.Position[j]-males[i].Position[j]) +
									config.MedianWeight*math.Exp(-config.Beta*rmedian*rmedian)*(medianPos[j]-males[i].Position[j])
							}
						} else {
							// Standard velocity update
							for j := range config.ProblemSize {
								rpbest := males[i].Best.Position[j] - males[i].Position[j]
								rgbest := globalBest.Position[j] - males[i].Position[j]
								males[i].Velocity[j] = g*males[i].Velocity[j] +
									config.A1*math.Exp(-config.Beta*rpbest*rpbest)*(males[i].Best.Position[j]-males[i].Position[j]) +
									config.A2*math.Exp(-config.Beta*rgbest*rgbest)*(globalBest.Position[j]-males[i].Position[j])
							}
						}
					} else {
						// Nuptial dance
						gVal := g
						if config.UseMPMA {
							gVal = mpmaG // Use MPMA gravity for dance too
						}

						for j := range config.ProblemSize {
							males[i].Velocity[j] = gVal*males[i].Velocity[j] + dance*e[j]
						}
					}

					// Apply velocity limits
					maxVec(males[i].Velocity, config.VelMin)
					minVec(males[i].Velocity, config.VelMax)

					// Update position
					for j := range config.ProblemSize {
						males[i].Position[j] += males[i].Velocity[j]
					}

					// Apply position limits
					maxVec(males[i].Position, config.LowerBound)
					minVec(males[i].Position, config.UpperBound)

					// Evaluate
					candidateEvaluator.evaluateMayfly(males[i], false)

					funcCount++

					// Update personal best
					if candidateEvaluator.better(
						evaluationFromMayfly(males[i]), evaluationFromBest(males[i].Best),
					) {
						copy(males[i].Best.Position, males[i].Position)
						males[i].Best.Cost = males[i].Cost
						males[i].Best.ConstraintViolation = males[i].ConstraintViolation

						// Update global best
						if candidateEvaluator.better(
							evaluationFromBest(males[i].Best), evaluationFromBest(globalBest),
						) {
							globalBest = cloneBest(males[i].Best)
						}
					}
				}
			}
		}

		// Sort populations by cost
		sortMayflies(males, candidateEvaluator)
		sortMayflies(females, candidateEvaluator)

		// OLCE-MA: Refine the elite males with orthogonal learning and
		// chaotic exploitation.
		if config.UseOLCE {
			numElite := olceEliteCount(len(males))
			refined := false

			// A zero factor collapses every orthogonal candidate onto the parent
			// male, so the whole stage is a guaranteed no-op. Skipping it keeps
			// the evaluation budget out of a step that cannot change anything.
			if config.OrthogonalFactor > 0 {
				// Prepare bounds vectors for orthogonal learning
				lb := make([]float64, config.ProblemSize)
				ub := make([]float64, config.ProblemSize)

				for j := range config.ProblemSize {
					lb[j] = config.LowerBound
					ub[j] = config.UpperBound
				}

				if evaluator != nil {
					orthogonalEvals, evaluationErr := evaluateParallelOrthogonalLearning(
						ctx,
						males,
						olceElitePercent,
						globalBest.Position,
						config.OrthogonalFactor,
						lb,
						ub,
						rng,
						evaluator,
					)
					if evaluationErr != nil {
						return nil, evaluationErr
					}

					funcCount += orthogonalEvals
				} else {
					// Apply to the elite males sequentially for backward compatibility.
					applyOrthogonalLearningToElite(
						males,
						olceElitePercent,
						globalBest.Position,
						config.OrthogonalFactor,
						lb, ub,
						candidateEvaluator,
						rng,
					)

					// Each elite male generates 4 candidates (L4 array).
					funcCount += numElite * len(L4Array)
				}

				refined = true
			}

			// Chaotic exploitation: one chaotic neighbor per elite male, with a
			// radius that decays over the run and greedy acceptance, so the step
			// can only improve the population.
			if config.ChaosFactor > 0 {
				if evaluator != nil {
					chaosEvals, evaluationErr := evaluateParallelChaoticExploitation(
						ctx, males, numElite, config, chaosMap, it, evaluator,
					)
					if evaluationErr != nil {
						return nil, evaluationErr
					}

					funcCount += chaosEvals
				} else {
					for i := range numElite {
						applyChaoticExploitation(males[i], config, chaosMap, it, candidateEvaluator)

						funcCount++
					}
				}

				refined = true
			}

			if refined {
				// Update global best if the refinement found a better solution
				for i := range numElite {
					if candidateEvaluator.betterMayflyThanBest(males[i], globalBest) {
						copyMayflyToBest(&globalBest, males[i])
					}
				}

				sortMayflies(males, candidateEvaluator)
			}
		}

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

		// GSASMA: Apply Golden Sine Algorithm with Simulated Annealing to elite males
		if config.UseGSASMA {
			if evaluator != nil {
				goldenSineEvals, evaluationErr := evaluateParallelGoldenSine(
					ctx,
					males,
					0.2,
					&globalBest,
					config.GoldenFactor,
					config.LowerBound,
					config.UpperBound,
					annealingScheduler,
					goldenSectionSearch,
					rng,
					evaluator,
				)
				if evaluationErr != nil {
					return nil, evaluationErr
				}

				funcCount += goldenSineEvals
			} else {
				// Apply GSA sequentially for backward compatibility.
				updatedGlobalBest, gsaFuncEvals := applyGSASMAToEliteMalesWithEvaluator(
					males,
					0.2, // Elite ratio: top 20%
					globalBest,
					config.GoldenFactor,
					config.LowerBound,
					config.UpperBound,
					annealingScheduler,
					goldenSectionSearch,
					candidateEvaluator,
					rng,
				)
				funcCount += gsaFuncEvals

				if candidateEvaluator.betterBest(updatedGlobalBest, globalBest) {
					globalBest = updatedGlobalBest
				}
			}

			// Re-sort after Golden Sine updates
			sortMayflies(males, candidateEvaluator)
		}

		// Mating - Create offspring
		var offspring []*Mayfly

		if evaluator != nil {
			parallelOffspring, offspringBest, evaluationErr := evaluateParallelGeneticOperators(
				ctx,
				males,
				females,
				config,
				rng,
				evaluator,
				it,
			)
			if evaluationErr != nil {
				return nil, evaluationErr
			}

			offspring = parallelOffspring
			funcCount += len(offspring)

			mergeBest(&globalBest, offspringBest, candidateEvaluator)
		} else {
			nc := effectiveNC(config)
			offspring = make([]*Mayfly, 0, nc)

			for k := range nc / 2 {
				p1, p2 := selectParents(males, females, k, config, rng)

				// Apply crossover
				off1Pos, off2Pos := Crossover(p1.Position, p2.Position, config.LowerBound, config.UpperBound, rng)

				// Create offspring 1
				off1 := newMayfly(config.ProblemSize)
				copy(off1.Position, off1Pos)

				candidateEvaluator.evaluateMayfly(off1, false)

				funcCount++

				if candidateEvaluator.betterMayflyThanBest(off1, globalBest) {
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

				if candidateEvaluator.betterMayflyThanBest(off2, globalBest) {
					copyMayflyToBest(&globalBest, off2)
				}

				copy(off2.Best.Position, off2.Position)
				off2.Best.Cost = off2.Cost
				off2.Best.ConstraintViolation = off2.ConstraintViolation

				offspring = append(offspring, off1, off2)
			}

			// Mutation
			// GSASMA: Use hybrid Cauchy-Gaussian mutation
			if config.UseGSASMA {
				// Apply hybrid mutation with adaptive Cauchy probability
				for range config.NM {
					// Select parent from offspring
					i := rng.Intn(len(offspring))
					p := offspring[i]

					mut := newMayfly(config.ProblemSize)

					// Calculate adaptive Cauchy probability based on iteration progress
					iterRatio := float64(it) / float64(config.MaxIterations)

					var cauchyProb float64

					switch {
					case iterRatio < 0.33:
						cauchyProb = 0.7 // Early: high exploration
					case iterRatio < 0.66:
						cauchyProb = 0.5 // Middle: balanced
					default:
						cauchyProb = config.CauchyMutationRate // Late: configured rate (default 0.3)
					}

					// Apply hybrid mutation
					mut.Position = HybridMutate(
						p.Position,
						config.Mu,
						config.LowerBound,
						config.UpperBound,
						cauchyProb,
						rng,
					)

					candidateEvaluator.evaluateMayfly(mut, false)

					funcCount++

					if candidateEvaluator.betterMayflyThanBest(mut, globalBest) {
						copyMayflyToBest(&globalBest, mut)
					}

					copy(mut.Best.Position, mut.Position)
					mut.Best.Cost = mut.Cost
					mut.Best.ConstraintViolation = mut.ConstraintViolation

					offspring = append(offspring, mut)
				}
			} else {
				// Standard mutation
				for range config.NM {
					// Select parent from offspring
					i := rng.Intn(len(offspring))
					p := offspring[i]

					mut := newMayfly(config.ProblemSize)
					mut.Position = Mutate(p.Position, config.Mu, config.LowerBound, config.UpperBound, rng)

					candidateEvaluator.evaluateMayfly(mut, false)

					funcCount++

					if candidateEvaluator.betterMayflyThanBest(mut, globalBest) {
						copyMayflyToBest(&globalBest, mut)
					}

					copy(mut.Best.Position, mut.Position)
					mut.Best.Cost = mut.Cost
					mut.Best.ConstraintViolation = mut.ConstraintViolation

					offspring = append(offspring, mut)
				}
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
				eliteMayfly, eliteFuncCount = generateEliteMayfliesWithEvaluator(
					globalBest,
					searchRange,
					config.EliteCount,
					config.ProblemSize,
					config.LowerBound,
					config.UpperBound,
					candidateEvaluator,
					rng,
				)
			}

			funcCount += eliteFuncCount

			// Replace worst male if elite is better
			if candidateEvaluator.betterMayfly(eliteMayfly, males[config.NPop-1]) {
				males[config.NPop-1] = eliteMayfly
				sortMayflies(males, candidateEvaluator) // Re-sort after replacement

				// Update global best if elite is the new best
				if candidateEvaluator.betterMayflyThanBest(eliteMayfly, globalBest) {
					copyMayflyToBest(&globalBest, eliteMayfly)
				}
			}

			lastGlobalBest = cloneBest(globalBest)
		}

		// GSASMA: Apply Opposition-Based Learning to global best
		if config.UseGSASMA && config.ApplyOBLToGlobalBest {
			// Apply OBL every 10 iterations to avoid excessive function evaluations
			if it%10 == 0 {
				updatedGlobalBest, improved := applyOBLToGlobalBestWithEvaluator(
					globalBest,
					config.LowerBound,
					config.UpperBound,
					candidateEvaluator,
				)
				funcCount++ // the opposition point evaluation

				if improved {
					globalBest = updatedGlobalBest
				}
			}
		}

		bestSolution[it] = globalBest.Cost
		iterationCount = it + 1

		// GSASMA: Update temperature schedule
		if config.UseGSASMA {
			annealingScheduler.Update()
		}

		// Update parameters
		g *= config.GDamp
		dance *= config.DanceDamp
		fl *= config.FLDamp

		notifyProgress(run.observer, it+1, funcCount, globalBest)
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
