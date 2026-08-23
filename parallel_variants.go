package mayfly

import (
	"context"
	"math/rand"
)

func largestParallelEvaluationBatch(config *Config) int {
	largest := max(config.NPop, config.NPopF, effectiveNC(config), effectiveNM(config))

	if config.UseDESMA {
		largest = max(largest, config.EliteCount)
	}

	if config.UseOLCE {
		numElite := min(max(int(float64(config.NPop)*0.2), 1), config.NPop)
		largest = max(largest, numElite*len(L4Array))
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

	bestElite := newMayfly(config.ProblemSize)
	copy(bestElite.Position, currentBest.Position)
	bestElite.Cost = currentBest.Cost
	bestElite.ConstraintViolation = currentBest.ConstraintViolation
	copy(bestElite.Best.Position, currentBest.Position)
	bestElite.Best.Cost = currentBest.Cost
	bestElite.Best.ConstraintViolation = currentBest.ConstraintViolation

	for _, elite := range elites {
		if evaluator.evaluator.betterMayfly(elite, bestElite) {
			bestElite = elite
			copy(bestElite.Best.Position, elite.Position)
			bestElite.Best.Cost = elite.Cost
			bestElite.Best.ConstraintViolation = elite.ConstraintViolation
		}
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
	// A zero factor collapses every candidate onto its parent male, so the
	// stage cannot change anything and must not spend evaluations.
	if factor <= 0 {
		return 0, nil
	}

	numElite := min(max(int(float64(len(males))*topPercent), 1), len(males))
	candidates := make([]*Mayfly, 0, numElite*len(L4Array))

	for i := range numElite {
		contextErr := ctx.Err()
		if contextErr != nil {
			return 0, contextErr
		}

		male := males[i]

		for _, row := range L4Array {
			candidate := newMayfly(len(male.Position))

			for j := range male.Position {
				var position float64
				if row[j%3] == 0 {
					position = male.Position[j] + factor*(male.Best.Position[j]-male.Position[j])
				} else {
					position = male.Position[j] + factor*(globalBest[j]-male.Position[j])
				}

				perturbation := (rng.Float64()*2 - 1) * factor * 0.1
				position += perturbation * (upperBounds[j] - lowerBounds[j])
				candidate.Position[j] = min(max(position, lowerBounds[j]), upperBounds[j])
			}

			candidates = append(candidates, candidate)
		}
	}

	_, evaluationErr := evaluator.evaluate(ctx, candidates, false, false)
	if evaluationErr != nil {
		return 0, evaluationErr
	}

	for i := range numElite {
		start := i * len(L4Array)
		bestCandidate := candidates[start]

		for _, candidate := range candidates[start+1 : start+len(L4Array)] {
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

	return len(candidates), nil
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

type goldenSineCandidate struct {
	mayfly         *Mayfly
	acceptanceDraw float64
}

func evaluateParallelGoldenSine(
	ctx context.Context,
	males []*Mayfly,
	eliteRatio float64,
	globalBest *Best,
	goldenFactor float64,
	lowerBound, upperBound float64,
	scheduler *AnnealingScheduler,
	section *goldenSection,
	rng *rand.Rand,
	evaluator *evaluationPool,
) (int, error) {
	numElite := min(max(int(float64(len(males))*eliteRatio), 1), len(males))
	// One snapshot of the section points is shared by the whole batch, because
	// the candidates are generated before any of them is evaluated. The section
	// is therefore advanced once per batch, after all results are in.
	sectionPoints := section.snapshot()
	candidates := make([]goldenSineCandidate, numElite)
	evaluationBatch := make([]*Mayfly, numElite)

	for i := range numElite {
		contextErr := ctx.Err()
		if contextErr != nil {
			return 0, contextErr
		}

		candidate := newMayfly(len(males[i].Position))
		candidate.Position = goldenSineUpdate(
			males[i].Position,
			globalBest.Position,
			goldenFactor,
			sectionPoints,
			lowerBound,
			upperBound,
			rng,
		)

		candidates[i] = goldenSineCandidate{
			mayfly:         candidate,
			acceptanceDraw: rng.Float64(),
		}
		evaluationBatch[i] = candidate
	}

	_, evaluationErr := evaluator.evaluate(ctx, evaluationBatch, false, false)
	if evaluationErr != nil {
		return 0, evaluationErr
	}

	temperature := scheduler.GetTemperature()

	batchImproved := false

	for i, candidate := range candidates {
		male := males[i]

		if evaluator.evaluator.betterMayfly(candidate.mayfly, male) {
			batchImproved = true
		}

		probability := evaluator.evaluator.acceptanceProbability(
			evaluationFromMayfly(male), evaluationFromMayfly(candidate.mayfly), temperature,
		)
		if !(candidate.acceptanceDraw < probability) {
			continue
		}

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

	// The whole batch was generated from a single section snapshot, so the
	// interval is narrowed exactly once for it. Narrowing per candidate would
	// judge later candidates against section points they were never generated
	// from.
	section.update(batchImproved)

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

// evaluateParallelChaoticExploitation runs the OLCE-MA chaotic exploitation
// step on the leading elites of males through the worker pool.
//
// Candidate generation stays on the caller goroutine so the LogisticMap is
// never shared, and acceptance is greedy: an elite only moves onto its chaotic
// neighbor when the neighbor is not worse. It returns the number of
// objective evaluations spent.
func evaluateParallelChaoticExploitation(
	ctx context.Context,
	males []*Mayfly,
	numElite int,
	config *Config,
	chaosMap *LogisticMap,
	iteration int,
	evaluator *evaluationPool,
) (int, error) {
	radius := chaoticExploitationRadius(config, iteration)
	candidates := make([]*Mayfly, 0, numElite)

	for i := range numElite {
		contextErr := ctx.Err()
		if contextErr != nil {
			return 0, contextErr
		}

		candidate := newMayfly(len(males[i].Position))
		chaoticExploitationCandidate(
			candidate.Position, males[i].Position, config, chaosMap, radius,
		)
		candidates = append(candidates, candidate)
	}

	_, evaluationErr := evaluator.evaluate(ctx, candidates, false, false)
	if evaluationErr != nil {
		return 0, evaluationErr
	}

	for i, candidate := range candidates {
		acceptChaoticCandidate(males[i], candidate, evaluator.evaluator)
	}

	return len(candidates), nil
}
