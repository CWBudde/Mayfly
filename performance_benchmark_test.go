package mayfly

import (
	"fmt"
	"math/rand"
	"testing"
)

// BenchmarkSortMayflies guards the population-ordering hot path. Swapping the
// endpoints keeps the population partially ordered between iterations, which
// mirrors selection after a small number of new candidates are introduced.
func BenchmarkSortMayflies(b *testing.B) {
	for _, populationSize := range []int{20, 100, 1000} {
		b.Run(fmt.Sprintf("population_%d", populationSize), func(b *testing.B) {
			rng := rand.New(rand.NewSource(42))
			population := make([]*Mayfly, populationSize)

			for i := range populationSize {
				candidate := newMayfly(1)
				candidate.Cost = rng.Float64()
				population[i] = candidate
			}

			b.ReportAllocs()

			for range b.N {
				population[0], population[populationSize-1] = population[populationSize-1], population[0]
				sortMayflies(population)
			}
		})
	}
}

// BenchmarkOptimizeBaseline is the representative end-to-end profiling
// workload. A fresh fixed-seed configuration per operation makes results
// reproducible and avoids carrying optimizer mutations between iterations.
func BenchmarkOptimizeBaseline(b *testing.B) {
	b.ReportAllocs()

	for range b.N {
		config := NewDefaultConfig()
		config.ObjectiveFunc = Sphere
		config.ProblemSize = 30
		config.LowerBound = -10
		config.UpperBound = 10
		config.MaxIterations = 100
		config.Rand = rand.New(rand.NewSource(42))

		result, err := Optimize(config)
		if err != nil {
			b.Fatal(err)
		}

		b.ReportMetric(float64(result.FuncEvalCount), "evals/op")
	}
}
