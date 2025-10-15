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
	"fmt"
	"math"
	"math/rand"
	"time"
)

// Optimize runs the Mayfly Optimization Algorithm with the given configuration.
func Optimize(config *Config) (*Result, error) {
	// Validate required parameters
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if config.ObjectiveFunc == nil {
		return nil, fmt.Errorf("ObjectiveFunc is required")
	}
	if config.ProblemSize <= 0 {
		return nil, fmt.Errorf("ProblemSize must be positive")
	}
	if config.LowerBound >= config.UpperBound {
		return nil, fmt.Errorf("LowerBound must be less than UpperBound")
	}
	if config.MaxIterations <= 0 {
		return nil, fmt.Errorf("MaxIterations must be positive")
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
	if rng == nil {
		rng = defaultRand()
	}

	// Initialize populations
	males := make([]*Mayfly, config.NPop)
	females := make([]*Mayfly, config.NPopF)

	globalBest := Best{
		Position: make([]float64, config.ProblemSize),
		Cost:     math.Inf(1),
	}

	funcCount := 0

	// Initialize male population
	for i := 0; i < config.NPop; i++ {
		males[i] = newMayfly(config.ProblemSize)
		males[i].Position = unifrndVec(config.LowerBound, config.UpperBound, config.ProblemSize, rng)
		males[i].Cost = config.ObjectiveFunc(males[i].Position)
		funcCount++

		// Update personal best
		copy(males[i].Best.Position, males[i].Position)
		males[i].Best.Cost = males[i].Cost

		// Update global best
		if males[i].Best.Cost < globalBest.Cost {
			globalBest.Cost = males[i].Best.Cost
			globalBest.Position = make([]float64, config.ProblemSize)
			copy(globalBest.Position, males[i].Best.Position)
		}
	}

	// Initialize female population
	for i := 0; i < config.NPopF; i++ {
		females[i] = newMayfly(config.ProblemSize)
		females[i].Position = unifrndVec(config.LowerBound, config.UpperBound, config.ProblemSize, rng)
		females[i].Cost = config.ObjectiveFunc(females[i].Position)
		funcCount++
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
	for it := 0; it < config.MaxIterations; it++ {
		// AOBLMOA: Use hybrid Mayfly-Aquila updates with opposition-based learning
		if config.UseAOBLMOA {
			// Apply AOBLMOA to populations
			applyAOBLMOAToPopulation(males, females, globalBest, it, config.MaxIterations, config)

			// Count function evaluations (approximation)
			// Aquila strategies: 1 eval per mayfly
			// Opposition learning: OppositionProbability * population size * 2 (original + opposition)
			aoblmoaEvals := config.NPop + config.NPopF
			oppositionEvals := int(config.OppositionProbability * float64(config.NPop+config.NPopF) * 2)
			funcCount += aoblmoaEvals + oppositionEvals

			// Update global best from updated populations
			for i := 0; i < config.NPop; i++ {
				if males[i].Cost < globalBest.Cost {
					globalBest.Cost = males[i].Cost
					copy(globalBest.Position, males[i].Position)
				}
			}
		} else if config.UseEOBBMA {
			// Update females with Gaussian sampling around best males
			for i := 0; i < config.NPopF; i++ {
				// Decide whether to use Lévy flight or Gaussian update
				if rng.Float64() < 0.5 {
					// Use Gaussian update toward best male
					newPos := gaussianUpdate(females[i].Position, males[i].Position,
						config.LowerBound, config.UpperBound, rng)
					copy(females[i].Position, newPos)
				} else {
					// Use Lévy flight for exploration
					levyStep := levyFlightVec(config.ProblemSize, config.LevyAlpha, config.LevyBeta, rng)
					for j := 0; j < config.ProblemSize; j++ {
						females[i].Position[j] += levyStep[j] * (config.UpperBound - config.LowerBound) * 0.01
					}
					maxVec(females[i].Position, config.LowerBound)
					minVec(females[i].Position, config.UpperBound)
				}

				females[i].Cost = config.ObjectiveFunc(females[i].Position)
				funcCount++
			}

			// Update males with Gaussian sampling around personal and global best
			for i := 0; i < config.NPop; i++ {
				// Decide whether to use Gaussian toward personal best or global best
				if rng.Float64() < 0.5 {
					// Gaussian toward personal best
					newPos := gaussianUpdate(males[i].Position, males[i].Best.Position,
						config.LowerBound, config.UpperBound, rng)
					copy(males[i].Position, newPos)
				} else {
					// Gaussian toward global best
					newPos := gaussianUpdate(males[i].Position, globalBest.Position,
						config.LowerBound, config.UpperBound, rng)
					copy(males[i].Position, newPos)
				}

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
		} else {
			// Standard velocity-based updates
			// Update females
			for i := 0; i < config.NPopF; i++ {
				e := unifrndVec(-1, 1, config.ProblemSize, rng)

				if females[i].Cost > males[i].Cost {
					// Attracted to male
					for j := 0; j < config.ProblemSize; j++ {
						rmf := males[i].Position[j] - females[i].Position[j]
						females[i].Velocity[j] = g*females[i].Velocity[j] +
							config.A3*math.Exp(-config.Beta*rmf*rmf)*(males[i].Position[j]-females[i].Position[j])
					}
				} else {
					// Random flight
					for j := 0; j < config.ProblemSize; j++ {
						females[i].Velocity[j] = g*females[i].Velocity[j] + fl*e[j]
					}
				}

				// Apply velocity limits
				maxVec(females[i].Velocity, config.VelMin)
				minVec(females[i].Velocity, config.VelMax)

				// Update position
				for j := 0; j < config.ProblemSize; j++ {
					females[i].Position[j] += females[i].Velocity[j]
				}

				// Apply position limits
				maxVec(females[i].Position, config.LowerBound)
				minVec(females[i].Position, config.UpperBound)

				// Evaluate
				females[i].Cost = config.ObjectiveFunc(females[i].Position)
				funcCount++
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
					medianPos = calculateWeightedMedianPosition(males, weights)
				} else {
					medianPos = calculateMedianPosition(males)
				}
				// Calculate non-linear gravity coefficient
				mpmaG = calculateGravityCoefficient(config.GravityType, it, config.MaxIterations)
			}

			// Update males
			for i := 0; i < config.NPop; i++ {
				e := unifrndVec(-1, 1, config.ProblemSize, rng)

				if males[i].Cost > globalBest.Cost {
					// Update velocity with personal and global best
					if config.UseMPMA {
						// MPMA: Include median position in velocity update
						for j := 0; j < config.ProblemSize; j++ {
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
						for j := 0; j < config.ProblemSize; j++ {
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
					for j := 0; j < config.ProblemSize; j++ {
						males[i].Velocity[j] = gVal*males[i].Velocity[j] + dance*e[j]
					}
				}

				// Apply velocity limits
				maxVec(males[i].Velocity, config.VelMin)
				minVec(males[i].Velocity, config.VelMax)

				// Update position
				for j := 0; j < config.ProblemSize; j++ {
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

		// Sort populations by cost
		sortMayflies(males)
		sortMayflies(females)

		// OLCE-MA: Apply orthogonal learning to elite males
		if config.UseOLCE {
			// Prepare bounds vectors for orthogonal learning
			lb := make([]float64, config.ProblemSize)
			ub := make([]float64, config.ProblemSize)
			for j := 0; j < config.ProblemSize; j++ {
				lb[j] = config.LowerBound
				ub[j] = config.UpperBound
			}

			// Apply to top 20% of males
			ApplyOrthogonalLearningToElite(
				males,
				0.2, // Top 20%
				globalBest.Position,
				config.OrthogonalFactor,
				lb, ub,
				config.ObjectiveFunc,
				rng,
			)

			// Count function evaluations from orthogonal learning
			// Each elite male generates 4 candidates (L4 array)
			numElite := int(float64(len(males)) * 0.2)
			if numElite < 1 {
				numElite = 1
			}
			funcCount += numElite * 4

			// Update global best if orthogonal learning found better solution
			for i := 0; i < numElite; i++ {
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
			// Apply opposition to top elite solutions with probability OppositionRate
			numEliteOpposition := config.EliteOppositionCount
			if numEliteOpposition > len(males) {
				numEliteOpposition = len(males)
			}

			for i := 0; i < numEliteOpposition; i++ {
				if rng.Float64() < config.OppositionRate {
					// Generate opposition point
					oppPos := oppositionPoint(males[i].Position, config.LowerBound, config.UpperBound)

					// Evaluate opposition point
					oppCost := config.ObjectiveFunc(oppPos)
					funcCount++

					// If opposition is better, replace the elite
					if oppCost < males[i].Cost {
						copy(males[i].Position, oppPos)
						males[i].Cost = oppCost

						// Update personal best
						if oppCost < males[i].Best.Cost {
							copy(males[i].Best.Position, oppPos)
							males[i].Best.Cost = oppCost
						}

						// Update global best
						if oppCost < globalBest.Cost {
							globalBest.Cost = oppCost
							copy(globalBest.Position, oppPos)
						}
					}
				}
			}

			// Re-sort after opposition learning
			sortMayflies(males)
		}

		// GSASMA: Apply Golden Sine Algorithm with Simulated Annealing to elite males
		if config.UseGSASMA {
			// Apply GSA to elite males (top 20%)
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

			// Update global best if GSA found better solution
			if updatedGlobalBestCost < globalBest.Cost {
				globalBest.Cost = updatedGlobalBestCost
				copy(globalBest.Position, updatedGlobalBest)
			}

			// Re-sort after Golden Sine updates
			sortMayflies(males)
		}

		// Mating - Create offspring
		offspring := make([]*Mayfly, 0, config.NC)
		for k := 0; k < config.NC/2; k++ {
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
				for j := 0; j < config.ProblemSize; j++ {
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
				for j := 0; j < config.ProblemSize; j++ {
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
			for k := 0; k < config.NM; k++ {
				// Select parent from offspring
				i := rng.Intn(len(offspring))
				p := offspring[i]

				mut := newMayfly(config.ProblemSize)

				// Calculate adaptive Cauchy probability based on iteration progress
				iterRatio := float64(it) / float64(config.MaxIterations)
				var cauchyProb float64
				if iterRatio < 0.33 {
					cauchyProb = 0.7 // Early: high exploration
				} else if iterRatio < 0.66 {
					cauchyProb = 0.5 // Middle: balanced
				} else {
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
					for j := 0; j < config.ProblemSize; j++ {
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
			for k := 0; k < config.NM; k++ {
				// Select parent from offspring
				i := rng.Intn(len(offspring))
				p := offspring[i]

				mut := newMayfly(config.ProblemSize)
				mut.Position = Mutate(p.Position, config.Mu, config.LowerBound, config.UpperBound, rng)

				// OLCE-MA: Apply chaotic exploitation to mutated offspring
				if config.UseOLCE {
					for j := 0; j < config.ProblemSize; j++ {
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

		// Merge offspring into populations
		split := len(offspring) / 2
		males = append(males, offspring[:split]...)
		females = append(females, offspring[split:]...)

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

			// Generate elite mayflies around global best
			eliteMayfly, eliteFuncCount := generateEliteMayflies(
				globalBest,
				searchRange,
				config.EliteCount,
				config.ProblemSize,
				config.LowerBound,
				config.UpperBound,
				config.ObjectiveFunc,
				rng,
			)
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
				updatedGlobalBest, updatedGlobalBestCost, oblFuncEvals, improved := applyOBLToGlobalBest(
					globalBest.Position,
					globalBest.Cost,
					config.LowerBound,
					config.UpperBound,
					config.ObjectiveFunc,
					rng,
				)
				funcCount += oblFuncEvals

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
	}

	return &Result{
		GlobalBest:     globalBest,
		BestSolution:   bestSolution,
		FuncEvalCount:  funcCount,
		IterationCount: config.MaxIterations,
	}, nil
}

// defaultRand creates a default random number generator
func defaultRand() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}
