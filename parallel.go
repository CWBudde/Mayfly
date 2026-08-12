package mayfly

import (
	"context"
	"math"
	"sync"
)

type evaluationJob struct {
	mayfly       *Mayfly
	best         *batchBest
	done         *sync.WaitGroup
	index        int
	sanitizeCost bool
}

type evaluationPool struct {
	evaluator *constraintEvaluator
	jobs      chan evaluationJob
	workers   sync.WaitGroup
}

func newEvaluationPool(objective ObjectiveFunction, maxWorkers int) *evaluationPool {
	return newConstrainedEvaluationPool(newConstraintEvaluator(objective, nil), maxWorkers)
}

func newConstrainedEvaluationPool(evaluator *constraintEvaluator, maxWorkers int) *evaluationPool {
	pool := &evaluationPool{
		evaluator: evaluator,
		jobs:      make(chan evaluationJob),
	}

	pool.workers.Add(maxWorkers)

	for range maxWorkers {
		go pool.worker()
	}

	return pool
}

func (pool *evaluationPool) worker() {
	defer pool.workers.Done()

	for job := range pool.jobs {
		pool.evaluator.evaluateMayfly(job.mayfly, job.sanitizeCost)

		if job.best != nil {
			job.best.consider(job.index, job.mayfly)
		}

		job.done.Done()
	}
}

func (pool *evaluationPool) evaluate(
	ctx context.Context,
	mayflies []*Mayfly,
	sanitizeCosts bool,
	trackBest bool,
) (Best, error) {
	var best *batchBest
	if trackBest {
		best = newBatchBest(pool.evaluator)
	}

	var done sync.WaitGroup

	for i, mayfly := range mayflies {
		contextErr := ctx.Err()
		if contextErr != nil {
			done.Wait()

			return Best{}, contextErr
		}

		done.Add(1)
		job := evaluationJob{
			mayfly:       mayfly,
			best:         best,
			done:         &done,
			index:        i,
			sanitizeCost: sanitizeCosts,
		}

		select {
		case pool.jobs <- job:
		case <-ctx.Done():
			done.Done()
			done.Wait()

			return Best{}, ctx.Err()
		}
	}

	done.Wait()

	contextErr := ctx.Err()
	if contextErr != nil {
		return Best{}, contextErr
	}

	if best == nil {
		return Best{}, nil
	}

	return best.snapshot(), nil
}

func (pool *evaluationPool) close() {
	close(pool.jobs)
	pool.workers.Wait()
}

type batchBest struct {
	evaluator *constraintEvaluator
	best      Best
	mu        sync.Mutex
	index     int
}

func newBatchBest(evaluator *constraintEvaluator) *batchBest {
	return &batchBest{
		best: Best{
			Cost:                math.Inf(1),
			ConstraintViolation: math.Inf(1),
		},
		evaluator: evaluator,
		index:     -1,
	}
}

func (best *batchBest) consider(index int, candidate *Mayfly) {
	if math.IsNaN(candidate.Cost) {
		return
	}

	best.mu.Lock()
	defer best.mu.Unlock()

	better := best.evaluator.betterMayflyThanBest(candidate, best.best)

	equal := !better && !best.evaluator.better(
		evaluationFromBest(best.best),
		evaluationFromMayfly(candidate),
	)
	if !better && (!equal || (best.index >= 0 && index >= best.index)) {
		return
	}

	best.best.Cost = candidate.Cost
	best.best.ConstraintViolation = candidate.ConstraintViolation
	best.best.Position = append(best.best.Position[:0], candidate.Position...)
	best.index = index
}

func (best *batchBest) snapshot() Best {
	best.mu.Lock()
	defer best.mu.Unlock()

	return Best{
		Position:            append([]float64(nil), best.best.Position...),
		Cost:                best.best.Cost,
		ConstraintViolation: best.best.ConstraintViolation,
	}
}

func effectiveMaxWorkers(config *Config) int {
	if config.MaxWorkers > 0 {
		return config.MaxWorkers
	}

	return defaultMaxWorkers()
}

func mergeBest(globalBest *Best, candidate Best, evaluator *constraintEvaluator) {
	if !evaluator.betterBest(candidate, *globalBest) {
		return
	}

	globalBest.Cost = candidate.Cost
	globalBest.ConstraintViolation = candidate.ConstraintViolation
	copy(globalBest.Position, candidate.Position)
}
