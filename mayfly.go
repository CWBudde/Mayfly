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

	rng := config.Rand

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

	// Main loop
	for it := 0; it < config.MaxIterations; it++ {
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

		// Update males
		for i := 0; i < config.NPop; i++ {
			e := unifrndVec(-1, 1, config.ProblemSize, rng)

			if males[i].Cost > globalBest.Cost {
				// Update velocity with personal and global best
				for j := 0; j < config.ProblemSize; j++ {
					rpbest := males[i].Best.Position[j] - males[i].Position[j]
					rgbest := globalBest.Position[j] - males[i].Position[j]
					males[i].Velocity[j] = g*males[i].Velocity[j] +
						config.A1*math.Exp(-config.Beta*rpbest*rpbest)*(males[i].Best.Position[j]-males[i].Position[j]) +
						config.A2*math.Exp(-config.Beta*rgbest*rgbest)*(globalBest.Position[j]-males[i].Position[j])
				}
			} else {
				// Nuptial dance
				for j := 0; j < config.ProblemSize; j++ {
					males[i].Velocity[j] = g*males[i].Velocity[j] + dance*e[j]
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

		// Sort populations by cost
		sortMayflies(males)
		sortMayflies(females)

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
		for k := 0; k < config.NM; k++ {
			// Select parent from offspring
			i := rng.Intn(len(offspring))
			p := offspring[i]

			mut := newMayfly(config.ProblemSize)
			mut.Position = Mutate(p.Position, config.Mu, config.LowerBound, config.UpperBound, rng)
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

		bestSolution[it] = globalBest.Cost

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
