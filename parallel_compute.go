package mayfly

import (
	"context"
	"sync"
	"sync/atomic"
)

// parallelFor runs independent CPU work with bounded concurrency. Callers must
// ensure each invocation owns its writes and only reads immutable shared data.
func parallelFor(ctx context.Context, count, maxWorkers int, work func(int)) error {
	if count == 0 {
		return ctx.Err()
	}

	workerCount := min(count, maxWorkers)

	var next atomic.Int64

	var workers sync.WaitGroup
	workers.Add(workerCount)

	for range workerCount {
		go func() {
			defer workers.Done()

			for {
				if ctx.Err() != nil {
					return
				}

				index := int(next.Add(1) - 1)
				if index >= count {
					return
				}

				work(index)
			}
		}()
	}

	workers.Wait()

	return ctx.Err()
}
