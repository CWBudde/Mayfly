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
	objective ObjectiveFunction
	jobs      chan evaluationJob
	workers   sync.WaitGroup
}

func newEvaluationPool(objective ObjectiveFunction, maxWorkers int) *evaluationPool {
	pool := &evaluationPool{
		objective: objective,
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
		cost := pool.objective(job.mayfly.Position)
		if job.sanitizeCost {
			cost = sanitizeCost(cost)
		}

		job.mayfly.Cost = cost
		if job.best != nil {
			job.best.consider(job.index, job.mayfly.Position, cost)
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
		best = newBatchBest()
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
	best  Best
	mu    sync.Mutex
	index int
}

func newBatchBest() *batchBest {
	return &batchBest{
		best:  Best{Cost: math.Inf(1)},
		index: -1,
	}
}

func (best *batchBest) consider(index int, position []float64, cost float64) {
	if math.IsNaN(cost) {
		return
	}

	best.mu.Lock()
	defer best.mu.Unlock()

	if cost > best.best.Cost || (cost == best.best.Cost && best.index >= 0 && index >= best.index) {
		return
	}

	best.best.Cost = cost
	best.best.Position = append(best.best.Position[:0], position...)
	best.index = index
}

func (best *batchBest) snapshot() Best {
	best.mu.Lock()
	defer best.mu.Unlock()

	return Best{
		Position: append([]float64(nil), best.best.Position...),
		Cost:     best.best.Cost,
	}
}

func effectiveMaxWorkers(config *Config) int {
	if config.MaxWorkers > 0 {
		return config.MaxWorkers
	}

	return defaultMaxWorkers()
}

func mergeBest(globalBest *Best, candidate Best) {
	if candidate.Cost >= globalBest.Cost {
		return
	}

	globalBest.Cost = candidate.Cost
	copy(globalBest.Position, candidate.Position)
}
