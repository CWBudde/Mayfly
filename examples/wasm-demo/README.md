# Swarm Lab — the mayfly WebAssembly demo

Two pages that run [github.com/cwbudde/mayfly](https://github.com/CWBudde/Mayfly)
compiled to `js/wasm`:

- **`index.html` — Swarm Lab.** A benchmark landscape with the male and female
  populations flying over it, scrubbable frame by frame, next to the convergence
  curve, the swarm's diversity, and the population's cost distribution.
- **`benchmark.html` — Variant Shootout.** The library's `ComparisonRunner`
  across all eight variants: descriptive statistics, ranks, pairwise Wilcoxon
  signed-rank tests, and a Friedman test.

The organising rule is that **no optimization logic lives in JavaScript**. Every
position, cost and statistic on both pages comes out of the Go library. The JS
owns the DOM, the canvas, and the animation clock. A demo that reimplemented the
algorithm in JS would be demonstrating the JS.

Published from `main` by
[`.github/workflows/wasm-demo-pages.yml`](../../.github/workflows/wasm-demo-pages.yml).

## Build and run

```bash
just run-wasm-demo              # build into ./dist and serve on :8090
just build-wasm-demo            # build only
./scripts/build-wasm-demo.sh /tmp/somewhere   # build somewhere else
```

**An HTTP server is required.** The pages fetch `mayfly.wasm`, and the Shootout
runs its sweeps in a Web Worker; neither works from a `file://` URL.

`wasm_exec.js` is copied from your Go toolchain at build time and never
committed — it is version-locked to the compiler that produced the `.wasm`, and
a stale copy fails at runtime in ways that look like demo bugs.

## What it exercises

- **The dual population.** Males chase their own best and the swarm's; females
  are drawn to a better male or fly at random. On Rastrigin at seed 999 the
  swarm settles into a local minimum at x≈0.995 instead of the optimum — worth
  scrubbing through.
- **`WithPopulationObserver`.** The per-iteration snapshot hook the swarm view
  is built on. The Swarm Lab is its first consumer.
- **Reproducibility.** _Same seed_ re-runs the identical trail and final cost,
  because `Config.Rand` makes a run deterministic.
- **All eight variants.** Try DESMA against standard MA at one seed on Rastrigin
  in 10 dimensions, where DESMA's elite probes pay for themselves; on Sphere in
  two dimensions they do not.
- **The comparison framework.** Paired runs, `WithSeed`, and the significance
  tests, computed by the library rather than restated in JS.
- **Quasi-random initialization.** The _initialization_ control sets
  `Config.QMCInit`: `uniform`, `sobol` or `halton`. Iteration 0 on the heatmap
  is where the difference is visible — a Sobol population covers the box with no
  gaps or clumps. The Shootout applies the choice to every variant in the sweep,
  so its table stays a comparison of variants rather than of seedings. See
  `docs/qmc-initialization.md` for what it is worth on the benchmark suite.

## Reading the numbers

- **The timings are not native Go.** Under `js/wasm` everything runs on one
  thread with no SIMD, and `EnableParallel` is switched off because a goroutine
  pool buys nothing when the browser schedules every goroutine onto the same
  thread. Treat the millisecond column as a relative cost between variants, not
  as a benchmark of the library.
- **The quick preset proves little.** Five paired runs can only detect an
  enormous effect; the library defaults to 30. An absence of significance at
  this size is an absence of evidence.
- **Above two dimensions the swarm view is a projection.** Two axes are plotted
  and the rest of each position is pinned at the known optimum for the heatmap.
  The optimizer still searches every dimension.
- **The landscape is rank-normalized by default.** Benchmark value distributions
  are wildly uneven — a linear or log ramp paints most of Rastrigin one colour —
  so each sample is coloured by its rank among the others. That shows where the
  low ground is, not how deep it is; `mode: "log"` preserves magnitude.

## Layout

| File              | Role                                                    |
| ----------------- | ------------------------------------------------------- |
| `main.go`         | Export table; publishes `globalThis.mayfly`             |
| `main_stub.go`    | No-op `main()` so non-wasm builds still work            |
| `bridge.go`       | `guard()` and the tolerant option readers               |
| `marshal.go`      | `Float32Array` sinks over JS-owned `ArrayBuffer`s       |
| `benchmarks.go`   | The 15 benchmarks with bounds, optima and blurbs        |
| `variants.go`     | Variant lookup and per-run `Config` construction        |
| `run.go`          | `run` — one optimization, recorded frame by frame       |
| `landscape.go`    | `landscape` — the objective sampled on a grid           |
| `compare.go`      | `compare` — one `ComparisonRunner` sweep                |
| `info.go`         | `info` — the capability table the UI builds itself from |
| `index.html`      | Swarm Lab markup, with its DOM contract                 |
| `benchmark.html`  | Variant Shootout markup, with its DOM contract          |
| `style.css`       | The shared instrument-rack stylesheet                   |
| `render.js`       | `window.Render` — all Swarm Lab drawing                 |
| `app.js`          | Swarm Lab controller                                    |
| `bench.js`        | Shootout controller                                     |
| `bench-worker.js` | Classic worker owning the Shootout's wasm instance      |
| `bench-chart.js`  | `globalThis.BenchChart` — the grouped bar chart         |

## Two things that look odd and are not

**`guard()` wraps every export.** A Go panic that unwinds out of a `js.Func`
aborts the whole wasm instance, so one bad request would brick the page until a
reload. Every export returns its failures as `{error, panic}` data instead.

**The Shootout's loop lives in JavaScript, not Go.** A call into Go blocks its
thread's event loop for its whole duration, so a "stop" posted by the page
cannot be dispatched while one is running. `compare` therefore covers exactly
one benchmark function per call, and the worker yields between calls. The
chunking _is_ the cancellation mechanism.
