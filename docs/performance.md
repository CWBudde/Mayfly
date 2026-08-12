# Performance and Profiling

Mayfly keeps a deterministic end-to-end benchmark and a focused population
ordering benchmark in `performance_benchmark_test.go`. Use the same machine,
Go version, power profile, and idle-system conditions when comparing changes.
Run several samples because optimizer timing is sensitive to scheduler and CPU
frequency variation.

## Reproducible baseline

Run the representative baselines with fixed inputs and report allocations:

```sh
go test -run '^$' \
  -bench '^(BenchmarkSortMayflies|BenchmarkOptimizeBaseline)$' \
  -benchmem -count=5 -benchtime=500ms .
```

For formal comparisons, save the results before and after a change and use
`benchstat`:

```sh
go test -run '^$' -bench '^Benchmark' -benchmem -count=10 . > before.bench
# Apply the change.
go test -run '^$' -bench '^Benchmark' -benchmem -count=10 . > after.bench
benchstat before.bench after.bench
```

The `.bench` extension is ignored so local measurements do not enter commits.

## CPU and memory profiles

The task runner captures profiles for the fixed-seed, 30-dimensional Sphere
workload with 20 male and 20 female candidates over 100 iterations:

```sh
just profile-cpu
just profile-mem
```

The equivalent combined command is:

```sh
go test -run '^$' -bench '^BenchmarkOptimizeBaseline$' -benchtime=5s \
  -cpuprofile=cpu.pprof -memprofile=memory.pprof .
```

Inspect cumulative CPU cost and allocated bytes, respectively:

```sh
go tool pprof -top cpu.pprof
go tool pprof -top -alloc_space memory.pprof
go tool pprof -http=:0 cpu.pprof
```

Profile files and the test binary produced by `go test` are ignored by Git.
Remove them after analysis with
`rm -f cpu.pprof memory.pprof mayfly.test`.

## August 2026 optimization baseline

The initial profile exposed `sortMayflies` as a material hot path. Selection
orders both populations several times per iteration, and the previous stable
bubble sort performed O(n²) comparisons even on nearly ordered populations.
It accounted for 500 ms of 4.43 s sampled CPU (11.3% cumulative) in the
representative workload. The implementation now uses Go's stable O(n log n)
slice sort, retaining deterministic ordering for equally preferred candidates.

Measurements were taken on Linux/amd64 with Go 1.26.0 on an AMD Ryzen 5 4600H.
The table reports the median of five 500 ms samples; lower is better.

| Benchmark              |          Before |           After | Change |
| ---------------------- | --------------: | --------------: | -----: |
| Sort, population 20    |     1,300 ns/op |       555 ns/op | -57.3% |
| Sort, population 100   |    27,676 ns/op |     2,409 ns/op | -91.3% |
| Sort, population 1,000 | 2,727,024 ns/op |    20,682 ns/op | -99.2% |
| Optimize baseline      | 9,129,232 ns/op | 8,658,945 ns/op |  -5.2% |

The end-to-end workload stayed at 6,140 objective evaluations, 15,180
allocations, and approximately 3.37 MB allocated per operation. In the
post-change CPU profile, cumulative `sortMayflies` time fell to 330 ms of
4.35 s sampled CPU (7.6%). Allocation profiles remained dominated by candidate
and position creation (`newMayfly`, `unifrndVec`, and `Crossover`); population
ordering introduced no per-operation allocation regression.

Absolute timings are machine-specific. Treat the recorded results as evidence
for this change and rerun the commands above on the target platform before
using them as a release performance threshold.
