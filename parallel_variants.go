package mayfly

import (
	"context"
	"math/rand"
	"sort"
)

func largestParallelEvaluationBatch(config *Config) int {
	largest := max(config.NPop, config.NPopF, effectiveNC(config), 2*effectiveNM(config))

	if config.UseDESMA {
		largest = max(largest, config.EliteCount)
	}

	if config.UseOLCE {
		largest = max(largest, config.NPop*len(orthogonalArray(config.ProblemSize)))
	}

	if config.UseEOBBMA {
		largest = max(largest, min(config.EliteOppositionCount, config.NPop))
	}

	if config.UseGSASMA {
		numElite := min(max(int(float64(config.NPop)*0.2), 1), config.NPop)
		largest = max(largest, numElite)
	}

	if config.UseAOBLMOA {
		// The update phase evaluates the whole swarm as one batch, and
		// NPop+NPopF exceeds max(NPop, NPopF), so this clause genuinely
		// enlarges the pool rather than restating the general case.
		largest = max(largest, config.NPop+config.NPopF)
	}

	return largest
}

func evaluateParallelDESMAElites(
	ctx context.Context,
	currentBest Best,
	searchRange float64,
	config *Config,
	rng *rand.Rand,
	evaluator *evaluationPool,
) (*Mayfly, int, error) {
	if config.EliteCount <= 0 {
		return nil, 0, nil
	}

	randomOffsets := make([]float64, config.EliteCount*config.ProblemSize)
	for i := range config.EliteCount {
		contextErr := ctx.Err()
		if contextErr != nil {
			return nil, 0, contextErr
		}

		offsetStart := i * config.ProblemSize
		for j := range config.ProblemSize {
			randomOffsets[offsetStart+j] = unifrnd(-1, 1, rng)
		}
	}

	elites := make([]*Mayfly, config.EliteCount)

	generationErr := parallelFor(ctx, len(elites), effectiveMaxWorkers(config), func(i int) {
		elite := newMayfly(config.ProblemSize)
		offsetStart := i * config.ProblemSize

		for j := range config.ProblemSize {
			elite.Position[j] = currentBest.Position[j] + randomOffsets[offsetStart+j]*searchRange
		}

		maxVec(elite.Position, config.LowerBound)
		minVec(elite.Position, config.UpperBound)
		elites[i] = elite
	})
	if generationErr != nil {
		return nil, 0, generationErr
	}

	_, evaluationErr := evaluator.evaluate(ctx, elites, false, false)
	if evaluationErr != nil {
		return nil, 0, evaluationErr
	}

	bestElite := mayflyFromBest(currentBest, config.ProblemSize)
	improved := false

	for _, elite := range elites {
		if evaluator.evaluator.betterMayfly(elite, bestElite) {
			bestElite = elite
			copy(bestElite.Best.Position, elite.Position)
			bestElite.Best.Cost = elite.Cost
			bestElite.Best.ConstraintViolation = elite.ConstraintViolation
			improved = true
		}
	}

	if !improved {
		return nil, len(elites), nil
	}

	return bestElite, len(elites), nil
}

func evaluateParallelOrthogonalLearning(
	ctx context.Context,
	males []*Mayfly,
	topPercent float64,
	globalBest []float64,
	factor float64,
	lowerBounds, upperBounds []float64,
	rng *rand.Rand,
	evaluator *evaluationPool,
) (int, error) {
	if factor <= 0 {
		return 0, nil
	}

	numElite := min(max(int(float64(len(males))*topPercent), 1), len(males))
	dim := len(males[0].Position)
	array := orthogonalArray(dim)
	candidates := make([]*Mayfly, 0, numElite*len(array))
	_ = rng // The published orthogonal design is deterministic.

	for i := range numElite {
		err := ctx.Err()
		if err != nil {
			return 0, err
		}

		male := males[i]

		for _, row := range array {
			candidate := newMayfly(dim)

			for j := range male.Position {
				var position float64
				if row[j] == 0 {
					position = male.Position[j] + factor*(male.Best.Position[j]-male.Position[j])
				} else {
					position = male.Position[j] + factor*(globalBest[j]-male.Position[j])
				}

				candidate.Position[j] = min(max(position, lowerBounds[j]), upperBounds[j])
			}

			candidates = append(candidates, candidate)
		}
	}

	_, evaluationErr := evaluator.evaluate(ctx, candidates, false, false)
	if evaluationErr != nil {
		return 0, evaluationErr
	}

	predicted := make([]*Mayfly, numElite)
	for i := range numElite {
		start := i * len(array)
		group := candidates[start : start+len(array)]
		ranked := append([]*Mayfly(nil), group...)
		sort.SliceStable(ranked, func(left, right int) bool {
			return evaluator.evaluator.betterMayfly(ranked[left], ranked[right])
		})

		rankByCandidate := make(map[*Mayfly]float64, len(ranked))
		for rank, candidate := range ranked {
			rankByCandidate[candidate] = float64(rank + 1)
		}

		male := males[i]

		candidate := newMayfly(dim)
		for dimension := range dim {
			levelScores := [2]float64{}
			for experiment, rowCandidate := range group {
				levelScores[array[experiment][dimension]] += rankByCandidate[rowCandidate]
			}

			if levelScores[1] < levelScores[0] {
				candidate.Position[dimension] = male.Position[dimension] +
					factor*(globalBest[dimension]-male.Position[dimension])
			} else {
				candidate.Position[dimension] = male.Position[dimension] +
					factor*(male.Best.Position[dimension]-male.Position[dimension])
			}

			candidate.Position[dimension] = min(
				max(candidate.Position[dimension], lowerBounds[dimension]), upperBounds[dimension],
			)
		}

		predicted[i] = candidate
	}

	_, evaluationErr = evaluator.evaluate(ctx, predicted, false, false)
	if evaluationErr != nil {
		return 0, evaluationErr
	}

	for i := range numElite {
		start := i * len(array)

		bestCandidate := predicted[i]
		for _, candidate := range candidates[start : start+len(array)] {
			if evaluator.evaluator.betterMayfly(candidate, bestCandidate) {
				bestCandidate = candidate
			}
		}

		male := males[i]
		if evaluator.evaluator.betterMayfly(bestCandidate, male) {
			copy(bestCandidate.Velocity, male.Velocity)
			copy(bestCandidate.Best.Position, male.Best.Position)
			bestCandidate.Best.Cost = male.Best.Cost
			bestCandidate.Best.ConstraintViolation = male.Best.ConstraintViolation

			if evaluator.evaluator.better(
				evaluationFromMayfly(bestCandidate), evaluationFromBest(bestCandidate.Best),
			) {
				copy(bestCandidate.Best.Position, bestCandidate.Position)
				bestCandidate.Best.Cost = bestCandidate.Cost
				bestCandidate.Best.ConstraintViolation = bestCandidate.ConstraintViolation
			}

			males[i] = bestCandidate
		}
	}

	return len(candidates) + len(predicted), nil
}

type oppositionCandidate struct {
	mayfly     *Mayfly
	eliteIndex int
}

func evaluateParallelEOBBMAOpposition(
	ctx context.Context,
	males []*Mayfly,
	globalBest *Best,
	config *Config,
	rng *rand.Rand,
	evaluator *evaluationPool,
) (int, error) {
	numElite := max(0, min(config.EliteOppositionCount, len(males)))
	selected := make([]oppositionCandidate, 0, numElite)
	evaluationBatch := make([]*Mayfly, 0, numElite)
	eliteLower, eliteUpper := eliteBounds(males, numElite, config.LowerBound, config.UpperBound)

	for i := range numElite {
		contextErr := ctx.Err()
		if contextErr != nil {
			return 0, contextErr
		}

		if rng.Float64() >= config.OppositionRate {
			continue
		}

		candidate := newMayfly(config.ProblemSize)
		candidate.Position = eliteOppositionPoint(
			males[i].Position, eliteLower, eliteUpper,
			config.LowerBound, config.UpperBound, rng,
		)
		selected = append(selected, oppositionCandidate{eliteIndex: i, mayfly: candidate})
		evaluationBatch = append(evaluationBatch, candidate)
	}

	_, evaluationErr := evaluator.evaluate(ctx, evaluationBatch, false, false)
	if evaluationErr != nil {
		return 0, evaluationErr
	}

	for _, candidate := range selected {
		male := males[candidate.eliteIndex]
		if evaluator.evaluator.betterMayfly(candidate.mayfly, male) {
			copy(male.Position, candidate.mayfly.Position)
			male.Cost = candidate.mayfly.Cost
			male.ConstraintViolation = candidate.mayfly.ConstraintViolation

			if evaluator.evaluator.better(
				evaluationFromMayfly(male), evaluationFromBest(male.Best),
			) {
				copy(male.Best.Position, male.Position)
				male.Best.Cost = male.Cost
				male.Best.ConstraintViolation = male.ConstraintViolation
			}

			if evaluator.evaluator.betterMayflyThanBest(male, *globalBest) {
				copyMayflyToBest(globalBest, male)
			}
		}
	}

	return len(evaluationBatch), nil
}

// evaluateParallelAOBLMOA runs one AOBLMOA update phase with the swarm
// evaluated through the worker pool.
//
// The moves themselves come from applyAOBLMOAMoves, the same function the
// sequential path calls, so the two paths cannot drift: this function only
// decides how the resulting positions are evaluated. Opposition-based learning
// no longer lives in the update phase, so there is no candidate machinery left
// here either -- every individual is moved once and evaluated once.
func evaluateParallelAOBLMOA(
	ctx context.Context,
	males, females []*Mayfly,
	globalBest *Best,
	currentIteration, maxIterations int,
	g, dance, flight float64,
	config *Config,
	rng *rand.Rand,
	evaluator *evaluationPool,
) (int, error) {
	contextErr := ctx.Err()
	if contextErr != nil {
		return 0, contextErr
	}

	applyAOBLMOAMoves(males, females, *globalBest, currentIteration, maxIterations,
		g, dance, flight, config, rng, evaluator.evaluator)

	batch := make([]*Mayfly, 0, len(males)+len(females))
	batch = append(batch, males...)
	batch = append(batch, females...)

	_, evaluationErr := evaluator.evaluate(ctx, batch, false, false)
	if evaluationErr != nil {
		return 0, evaluationErr
	}

	updatePersonalBests(males, evaluator.evaluator)

	return len(batch), nil
}

// evaluateParallelChaoticExploitation is retained as an internal compatibility
// helper. Mayfly's historical stage applies chaos to one fittest offspring, so
// numElite is ignored and this function evaluates exactly one candidate. This
// is not the published all-offspring Chebyshev stage; see chaos.go.
func evaluateParallelChaoticExploitation(
	ctx context.Context,
	males []*Mayfly,
	numElite int,
	config *Config,
	chaosMap *LogisticMap,
	iteration int,
	evaluator *evaluationPool,
) (int, error) {
	_ = numElite

	err := ctx.Err()
	if err != nil {
		return 0, err
	}

	target := fittestMayfly(males, evaluator.evaluator)
	if target == nil {
		return 0, nil
	}

	candidate := newMayfly(len(target.Position))
	chaoticExploitationCandidate(
		candidate.Position, target.Position, config, chaosMap,
		chaoticConstrictionFactor(config, iteration),
	)

	_, evaluationErr := evaluator.evaluate(ctx, []*Mayfly{candidate}, false, false)
	if evaluationErr != nil {
		return 0, evaluationErr
	}

	commitChaoticOffspring(target, candidate)

	return 1, nil
}
