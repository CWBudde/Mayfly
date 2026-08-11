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
	}

	var evaluator *evaluationPool

	if config.EnableParallel {
		largestBatch := largestParallelEvaluationBatch(config)
		workerCount := min(effectiveMaxWorkers(config), largestBatch)

		evaluator = newEvaluationPool(config.ObjectiveFunc, workerCount)
		defer evaluator.close()
	}

	// Initialize populations
	males := make([]*Mayfly, config.NPop)
	females := make([]*Mayfly, config.NPopF)

	globalBest := Best{
		Position: make([]float64, config.ProblemSize),
		Cost:     math.Inf(1),
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
		}

		mergeBest(&globalBest, initialBest)

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

			males[i].Cost = evaluateWithSanitization(config.ObjectiveFunc, males[i].Position,
				config.LowerBound, config.UpperBound, rng)
			funcCount++

			copy(males[i].Best.Position, males[i].Position)
			males[i].Best.Cost = males[i].Cost

			if males[i].Best.Cost < globalBest.Cost {
				globalBest.Cost = males[i].Best.Cost
				globalBest.Position = make([]float64, config.ProblemSize)
				copy(globalBest.Position, males[i].Best.Position)
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

			females[i].Cost = evaluateWithSanitization(config.ObjectiveFunc, females[i].Position,
				config.LowerBound, config.UpperBound, rng)
			funcCount++
		}
	}

	bestSolution := make([]float64, config.MaxIterations)
	g := config.G
	dance := config.Dance
	fl := config.FL

	// Initialize DESMA parameters if enabled
	var searchRange float64

	var lastGlobalBestCost float64

	if config.UseDESMA {
		if config.SearchRange == 0 {
			// Auto-calculate initial search range as 10% of the search space
			searchRange = 0.1 * (config.UpperBound - config.LowerBound)
		} else {
			searchRange = config.SearchRange
		}

		lastGlobalBestCost = globalBest.Cost
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

	// Initialize AOBLMOA parameters if enabled
	var paretoArchive *ParetoArchive

	if config.UseAOBLMOA {
		initializeAOBLMOA(config)
		paretoArchive = NewParetoArchive(config.ArchiveSize)
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
				applyAOBLMOAToPopulation(males, females, globalBest, it, config.MaxIterations, config)

				// Count function evaluations (approximation)
				// Aquila strategies: 1 eval per mayfly
				// Opposition learning: OppositionProbability * population size * 2 (original + opposition)
				aoblmoaEvals := config.NPop + config.NPopF
				oppositionEvals := int(config.OppositionProbability * float64(config.NPop+config.NPopF) * 2)
				funcCount += aoblmoaEvals + oppositionEvals
			}

			// Update global best from updated populations
			for i := range config.NPop {
				if males[i].Cost < globalBest.Cost {
					globalBest.Cost = males[i].Cost
					copy(globalBest.Position, males[i].Position)
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
				updatePersonalBests(males)
				mergeBest(&globalBest, maleBest)
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

					females[i].Cost = config.ObjectiveFunc(females[i].Position)
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

					males[i].Cost = config.ObjectiveFunc(males[i].Position)
					funcCount++

					if males[i].Cost < males[i].Best.Cost {
						copy(males[i].Best.Position, males[i].Position)
						males[i].Best.Cost = males[i].Cost

						if males[i].Best.Cost < globalBest.Cost {
							globalBest.Cost = males[i].Best.Cost
							copy(globalBest.Position, males[i].Best.Position)
						}
					}
				}
			}
		default:
			// Standard velocity-based updates
			if evaluator != nil {
				prepareStandardFemales(females, males, g, fl, config, rng)

				_, femaleEvaluationErr := evaluator.evaluate(ctx, females, false, false)
				if femaleEvaluationErr != nil {
					return nil, femaleEvaluationErr
				}

				funcCount += len(females)
			} else {
				// Update females sequentially for backward compatibility.
				for i := range config.NPopF {
					e := unifrndVec(-1, 1, config.ProblemSize, rng)

					if females[i].Cost > males[i].Cost {
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
					females[i].Cost = config.ObjectiveFunc(females[i].Position)
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
						if maxCost > minCost {
							weights[i] = 1.0 - (males[i].Cost-minCost)/(maxCost-minCost)
						} else {
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
				prepareStandardMales(males, phaseBest, medianPos, g, dance, mpmaG, config, rng)

				maleBest, maleEvaluationErr := evaluator.evaluate(ctx, males, false, true)
				if maleEvaluationErr != nil {
					return nil, maleEvaluationErr
				}

				funcCount += len(males)
				updatePersonalBests(males)
				mergeBest(&globalBest, maleBest)
			} else {
				// Update males sequentially for backward compatibility.
				for i := range config.NPop {
					e := unifrndVec(-1, 1, config.ProblemSize, rng)

					if males[i].Cost > globalBest.Cost {
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
					males[i].Cost = config.ObjectiveFunc(males[i].Position)
					funcCount++

					// Update personal best
					if males[i].Cost < males[i].Best.Cost {
						copy(males[i].Best.Position, males[i].Position)
						males[i].Best.Cost = males[i].Cost

						// Update global best
						if males[i].Best.Cost < globalBest.Cost {
							globalBest.Cost = males[i].Best.Cost
							copy(globalBest.Position, males[i].Best.Position)
						}
					}
				}
			}
		}

		// Sort populations by cost
		sortMayflies(males)
		sortMayflies(females)

		// OLCE-MA: Apply orthogonal learning to elite males
		if config.UseOLCE {
			// Prepare bounds vectors for orthogonal learning
			lb := make([]float64, config.ProblemSize)
			ub := make([]float64, config.ProblemSize)

			for j := range config.ProblemSize {
				lb[j] = config.LowerBound
				ub[j] = config.UpperBound
			}

			numElite := max(int(float64(len(males))*0.2), 1)

			if evaluator != nil {
				orthogonalEvals, evaluationErr := evaluateParallelOrthogonalLearning(
					ctx,
					males,
					0.2,
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
				// Apply to top 20% of males sequentially for backward compatibility.
				ApplyOrthogonalLearningToElite(
					males,
					0.2, // Top 20%
					globalBest.Position,
					config.OrthogonalFactor,
					lb, ub,
					config.ObjectiveFunc,
					rng,
				)

				// Each elite male generates 4 candidates (L4 array).
				funcCount += numElite * len(L4Array)
			}

			// Update global best if orthogonal learning found better solution
			for i := range numElite {
				if males[i].Cost < globalBest.Cost {
					globalBest.Cost = males[i].Cost
					copy(globalBest.Position, males[i].Position)
				}
			}

			// Re-sort after orthogonal learning
			sortMayflies(males)
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

				for i := range numEliteOpposition {
					if rng.Float64() < config.OppositionRate {
						oppPos := oppositionPoint(males[i].Position, config.LowerBound, config.UpperBound)
						oppCost := config.ObjectiveFunc(oppPos)
						funcCount++

						if oppCost < males[i].Cost {
							copy(males[i].Position, oppPos)
							males[i].Cost = oppCost

							if oppCost < males[i].Best.Cost {
								copy(males[i].Best.Position, oppPos)
								males[i].Best.Cost = oppCost
							}

							if oppCost < globalBest.Cost {
								globalBest.Cost = oppCost
								copy(globalBest.Position, oppPos)
							}
						}
					}
				}
			}

			// Re-sort after opposition learning
			sortMayflies(males)
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
					it,
					config.MaxIterations,
					config.LowerBound,
					config.UpperBound,
					annealingScheduler,
					rng,
					evaluator,
				)
				if evaluationErr != nil {
					return nil, evaluationErr
				}

				funcCount += goldenSineEvals
			} else {
				// Apply GSA sequentially for backward compatibility.
				updatedGlobalBest, updatedGlobalBestCost, gsaFuncEvals := applyGSASMAToEliteMales(
					males,
					0.2, // Elite ratio: top 20%
					globalBest.Position,
					globalBest.Cost,
					config.GoldenFactor,
					it,
					config.MaxIterations,
					config.LowerBound,
					config.UpperBound,
					annealingScheduler,
					config.ObjectiveFunc,
					rng,
				)
				funcCount += gsaFuncEvals

				if updatedGlobalBestCost < globalBest.Cost {
					globalBest.Cost = updatedGlobalBestCost
					copy(globalBest.Position, updatedGlobalBest)
				}
			}

			// Re-sort after Golden Sine updates
			sortMayflies(males)
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
				chaosMap,
				evaluator,
				it,
			)
			if evaluationErr != nil {
				return nil, evaluationErr
			}

			offspring = parallelOffspring
			funcCount += len(offspring)

			mergeBest(&globalBest, offspringBest)
		} else {
			offspring = make([]*Mayfly, 0, config.NC)

			for k := range config.NC / 2 {
				// Select parents (best males and females)
				p1 := males[k]
				p2 := females[k]

				// Apply crossover
				off1Pos, off2Pos := Crossover(p1.Position, p2.Position, config.LowerBound, config.UpperBound, rng)

				// Create offspring 1
				off1 := newMayfly(config.ProblemSize)
				copy(off1.Position, off1Pos)

				// OLCE-MA: Apply chaotic exploitation to offspring
				if config.UseOLCE {
					for j := range config.ProblemSize {
						chaosValue := chaosMap.Next()
						perturbation := config.ChaosFactor * (chaosValue - 0.5) * (config.UpperBound - config.LowerBound)
						off1.Position[j] += perturbation

						// Apply bounds
						if off1.Position[j] < config.LowerBound {
							off1.Position[j] = config.LowerBound
						}

						if off1.Position[j] > config.UpperBound {
							off1.Position[j] = config.UpperBound
						}
					}
				}

				off1.Cost = config.ObjectiveFunc(off1.Position)
				funcCount++

				if off1.Cost < globalBest.Cost {
					globalBest.Cost = off1.Cost
					copy(globalBest.Position, off1.Position)
				}

				copy(off1.Best.Position, off1.Position)
				off1.Best.Cost = off1.Cost

				// Create offspring 2
				off2 := newMayfly(config.ProblemSize)
				copy(off2.Position, off2Pos)

				// OLCE-MA: Apply chaotic exploitation to offspring
				if config.UseOLCE {
					for j := range config.ProblemSize {
						chaosValue := chaosMap.Next()
						perturbation := config.ChaosFactor * (chaosValue - 0.5) * (config.UpperBound - config.LowerBound)
						off2.Position[j] += perturbation

						// Apply bounds
						if off2.Position[j] < config.LowerBound {
							off2.Position[j] = config.LowerBound
						}

						if off2.Position[j] > config.UpperBound {
							off2.Position[j] = config.UpperBound
						}
					}
				}

				off2.Cost = config.ObjectiveFunc(off2.Position)
				funcCount++

				if off2.Cost < globalBest.Cost {
					globalBest.Cost = off2.Cost
					copy(globalBest.Position, off2.Position)
				}

				copy(off2.Best.Position, off2.Position)
				off2.Best.Cost = off2.Cost

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

					// OLCE-MA: Apply chaotic exploitation to mutated offspring if OLCE is also enabled
					if config.UseOLCE {
						for j := range config.ProblemSize {
							// Apply chaotic perturbation
							chaosValue := chaosMap.Next()
							perturbation := config.ChaosFactor * (chaosValue - 0.5) * (config.UpperBound - config.LowerBound)
							mut.Position[j] += perturbation

							// Apply bounds
							if mut.Position[j] < config.LowerBound {
								mut.Position[j] = config.LowerBound
							}

							if mut.Position[j] > config.UpperBound {
								mut.Position[j] = config.UpperBound
							}
						}
					}

					mut.Cost = config.ObjectiveFunc(mut.Position)
					funcCount++

					if mut.Cost < globalBest.Cost {
						globalBest.Cost = mut.Cost
						copy(globalBest.Position, mut.Position)
					}

					copy(mut.Best.Position, mut.Position)
					mut.Best.Cost = mut.Cost

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

					// OLCE-MA: Apply chaotic exploitation to mutated offspring
					if config.UseOLCE {
						for j := range config.ProblemSize {
							// Apply chaotic perturbation
							chaosValue := chaosMap.Next()
							perturbation := config.ChaosFactor * (chaosValue - 0.5) * (config.UpperBound - config.LowerBound)
							mut.Position[j] += perturbation

							// Apply bounds
							if mut.Position[j] < config.LowerBound {
								mut.Position[j] = config.LowerBound
							}

							if mut.Position[j] > config.UpperBound {
								mut.Position[j] = config.UpperBound
							}
						}
					}

					mut.Cost = config.ObjectiveFunc(mut.Position)
					funcCount++

					if mut.Cost < globalBest.Cost {
						globalBest.Cost = mut.Cost
						copy(globalBest.Position, mut.Position)
					}

					copy(mut.Best.Position, mut.Position)
					mut.Best.Cost = mut.Cost

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
		sortMayflies(males)
		sortMayflies(females)

		males = males[:config.NPop]
		females = females[:config.NPopF]

		// DESMA: Apply dynamic elite strategy
		if config.UseDESMA {
			// Dynamically adjust search range based on improvement
			if globalBest.Cost < lastGlobalBestCost {
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
				eliteMayfly, eliteFuncCount = generateEliteMayflies(
					globalBest,
					searchRange,
					config.EliteCount,
					config.ProblemSize,
					config.LowerBound,
					config.UpperBound,
					config.ObjectiveFunc,
					rng,
				)
			}

			funcCount += eliteFuncCount

			// Replace worst male if elite is better
			if eliteMayfly.Cost < males[config.NPop-1].Cost {
				males[config.NPop-1] = eliteMayfly
				sortMayflies(males) // Re-sort after replacement

				// Update global best if elite is the new best
				if eliteMayfly.Cost < globalBest.Cost {
					globalBest.Cost = eliteMayfly.Cost
					copy(globalBest.Position, eliteMayfly.Position)
				}
			}

			lastGlobalBestCost = globalBest.Cost
		}

		// GSASMA: Apply Opposition-Based Learning to global best
		if config.UseGSASMA && config.ApplyOBLToGlobalBest {
			// Apply OBL every 10 iterations to avoid excessive function evaluations
			if it%10 == 0 {
				updatedGlobalBest, updatedGlobalBestCost, improved := applyOBLToGlobalBest(
					globalBest.Position,
					globalBest.Cost,
					config.LowerBound,
					config.UpperBound,
					config.ObjectiveFunc,
				)
				funcCount++ // the opposition point evaluation

				if improved {
					globalBest.Cost = updatedGlobalBestCost
					copy(globalBest.Position, updatedGlobalBest)
				}
			}
		}

		// AOBLMOA: Update Pareto archive
		if config.UseAOBLMOA {
			updateParetoArchive(paretoArchive, males, females)
		}

		bestSolution[it] = globalBest.Cost

		// GSASMA: Update temperature schedule
		if config.UseGSASMA {
			annealingScheduler.Update()
		}

		// Update parameters
		g *= config.GDamp
		dance *= config.DanceDamp
		fl *= config.FLDamp

		notifyProgress(run.observer, it+1, funcCount, globalBest)

		iterationErr = ctx.Err()
		if iterationErr != nil {
			return nil, iterationErr
		}
	}

	return &Result{
		GlobalBest:       globalBest,
		ConvergenceCurve: bestSolution,
		FuncEvalCount:    funcCount,
		IterationCount:   config.MaxIterations,
		Seed:             seed,
	}, nil
}
