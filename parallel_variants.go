package mayfly

import (
	"context"
	"math/rand"
)

func largestParallelEvaluationBatch(config *Config) int {
	largest := max(config.NPop, config.NPopF, effectiveNC(config), config.NM)

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
		largest = max(largest, 2*(config.NPop+config.NPopF))
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

	for i := range numElite {
		contextErr := ctx.Err()
		if contextErr != nil {
			return 0, contextErr
		}

		if rng.Float64() >= config.OppositionRate {
			continue
		}

		candidate := newMayfly(config.ProblemSize)
		candidate.Position = oppositionPoint(males[i].Position, config.LowerBound, config.UpperBound)
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
	currentIteration, maxIterations int,
	lowerBound, upperBound float64,
	scheduler *AnnealingScheduler,
	rng *rand.Rand,
	evaluator *evaluationPool,
) (int, error) {
	numElite := min(max(int(float64(len(males))*eliteRatio), 1), len(males))
	candidates := make([]goldenSineCandidate, numElite)
	evaluationBatch := make([]*Mayfly, numElite)

	for i := range numElite {
		contextErr := ctx.Err()
		if contextErr != nil {
			return 0, contextErr
		}

		candidate := newMayfly(len(males[i].Position))
		candidate.Position = goldenSineUpdateAdaptive(
			males[i].Position,
			globalBest.Position,
			goldenFactor,
			currentIteration,
			maxIterations,
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

	for i, candidate := range candidates {
		male := males[i]

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

	return len(evaluationBatch), nil
}

type aquilaCandidate struct {
	target     *Mayfly
	original   *Mayfly
	opposition *Mayfly
	isMale     bool
}

// evaluateParallelAOBLMOA moves and evaluates every male and female exactly
// once per iteration.
//
// An individual either takes an Aquila-Optimizer step (with probability
// config.AquilaWeight) or the ordinary Mayfly velocity and position update.
// Both branches move the individual, so the batch handed to the evaluator
// always covers the whole population.
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
	strategyConfig := *config
	strategyConfig.Rand = rng
	candidates := make([]aquilaCandidate, 0, len(males)+len(females))
	comparisonBatch := make([]*Mayfly, 0, 2*(len(males)+len(females)))

	aquilaCandidateFor := func(target *Mayfly, population []*Mayfly, isMale bool) aquilaCandidate {
		strategy := selectAquilaStrategy(currentIteration, maxIterations, rng)
		position := applyAquilaStrategy(
			target,
			*globalBest,
			population,
			strategy,
			currentIteration,
			maxIterations,
			&strategyConfig,
		)

		original := newMayfly(config.ProblemSize)
		copy(original.Position, position)
		maxVec(original.Position, config.LowerBound)
		minVec(original.Position, config.UpperBound)

		candidate := aquilaCandidate{target: target, original: original, isMale: isMale}

		if rng.Float64() < config.OppositionProbability {
			opposition := newMayfly(config.ProblemSize)
			opposition.Position = oppositionPoint(
				original.Position,
				config.LowerBound,
				config.UpperBound,
			)
			candidate.opposition = opposition
			comparisonBatch = append(comparisonBatch, original, opposition)
		}

		return candidate
	}

	// mayflyCandidate captures the position produced by the ordinary Mayfly
	// update, which the caller has already applied to target in place.
	mayflyCandidate := func(target *Mayfly, isMale bool) aquilaCandidate {
		moved := newMayfly(config.ProblemSize)
		copy(moved.Position, target.Position)

		return aquilaCandidate{target: target, original: moved, isMale: isMale}
	}

	for _, target := range males {
		if ctx.Err() != nil {
			break
		}

		if rng.Float64() < config.AquilaWeight {
			candidates = append(candidates, aquilaCandidateFor(target, males, true))

			continue
		}

		prepareStandardMale(
			target, *globalBest, nil, g, dance, g, config, rng, evaluator.evaluator,
		)
		candidates = append(candidates, mayflyCandidate(target, true))
	}

	for i, target := range females {
		if ctx.Err() != nil {
			break
		}

		if rng.Float64() < config.AquilaWeight {
			candidates = append(candidates, aquilaCandidateFor(target, females, false))

			continue
		}

		pairedMale := target
		if i < len(males) {
			pairedMale = males[i]
		}

		prepareStandardFemale(target, pairedMale, g, flight, config, rng, evaluator.evaluator)
		candidates = append(candidates, mayflyCandidate(target, false))
	}

	contextErr := ctx.Err()
	if contextErr != nil {
		return 0, contextErr
	}

	_, comparisonErr := evaluator.evaluate(ctx, comparisonBatch, false, false)
	if comparisonErr != nil {
		return 0, comparisonErr
	}

	finalBatch := selectAquilaWinners(candidates, evaluator)

	_, finalErr := evaluator.evaluate(ctx, finalBatch, false, false)
	if finalErr != nil {
		return 0, finalErr
	}

	commitAquilaCandidates(candidates, finalBatch, globalBest, evaluator)

	return len(comparisonBatch) + len(finalBatch), nil
}

// selectAquilaWinners picks, for every candidate, the better of the Aquila
// position and its opposition point, where opposition-based learning fired.
func selectAquilaWinners(candidates []aquilaCandidate, evaluator *evaluationPool) []*Mayfly {
	winners := make([]*Mayfly, len(candidates))

	for i := range candidates {
		selected := candidates[i].original
		if candidates[i].opposition != nil &&
			evaluator.evaluator.betterMayfly(candidates[i].opposition, selected) {
			selected = candidates[i].opposition
		}

		winners[i] = selected
	}

	return winners
}

// commitAquilaCandidates writes the evaluated positions back onto the swarm and
// refreshes the personal and global bests.
func commitAquilaCandidates(
	candidates []aquilaCandidate,
	finalBatch []*Mayfly,
	globalBest *Best,
	evaluator *evaluationPool,
) {
	for i, candidate := range candidates {
		selected := finalBatch[i]
		copy(candidate.target.Position, selected.Position)
		candidate.target.Cost = selected.Cost
		candidate.target.ConstraintViolation = selected.ConstraintViolation

		if !candidate.isMale {
			continue
		}

		if evaluator.evaluator.better(
			evaluationFromMayfly(candidate.target), evaluationFromBest(candidate.target.Best),
		) {
			copy(candidate.target.Best.Position, candidate.target.Position)
			candidate.target.Best.Cost = candidate.target.Cost
			candidate.target.Best.ConstraintViolation = candidate.target.ConstraintViolation
		}

		if evaluator.evaluator.betterMayflyThanBest(candidate.target, *globalBest) {
			copyMayflyToBest(globalBest, candidate.target)
		}
	}
}
