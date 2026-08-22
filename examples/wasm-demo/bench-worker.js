/*
 * bench-worker.js — the shootout's wasm instance, off the main thread.
 *
 * CLASSIC worker, deliberately. wasm_exec.js is a classic script that assigns
 * globalThis.Go, and module workers cannot importScripts(), so a
 * type:"module" worker has no way to load it short of rewriting it.
 *
 * Two facts shape everything below.
 *
 * 1. go.run() must NOT be awaited. The demo's main() ends in select{}, so the
 *    promise it returns never resolves. Start it and poll for self.mayfly to
 *    appear instead.
 *
 * 2. A call into Go is synchronous and blocks this worker's event loop for its
 *    whole duration, so a "cancel" message posted by the page cannot be
 *    dispatched while a Go call is in flight. That is why the sweep loop lives
 *    here rather than in Go: each compare() call covers exactly one benchmark
 *    function, and between calls we yield with setTimeout so the queued cancel
 *    actually lands. The chunking IS the cancellation mechanism.
 */

"use strict";

importScripts("wasm_exec.js");

const READY_TIMEOUT_MS = 30000;
const READY_POLL_MS = 10;

let api = null;
let cancelRequested = -1;

function post(message) {
  self.postMessage(message);
}

function yieldToLoop() {
  return new Promise((resolve) => setTimeout(resolve, 0));
}

async function boot() {
  const go = new Go();
  const response = await fetch("mayfly.wasm");

  if (!response.ok) {
    throw new Error(`fetch mayfly.wasm: ${response.status}`);
  }

  const result = await WebAssembly.instantiate(
    await response.arrayBuffer(),
    go.importObject,
  );

  go.run(result.instance).catch((err) => {
    post({ type: "fatal", error: `go runtime exited: ${err && err.message}` });
  });

  const deadline = Date.now() + READY_TIMEOUT_MS;

  while (!self.mayfly) {
    if (Date.now() > deadline) {
      throw new Error("timed out waiting for the wasm module to publish mayfly");
    }

    await new Promise((resolve) => setTimeout(resolve, READY_POLL_MS));
  }

  api = self.mayfly;
  post({ type: "ready", info: api.info() });
}

async function runSweep(request) {
  const runId = request.runId;
  const benchmarks = request.benchmarks || [];
  let completed = 0;
  let cancelled = false;

  post({ type: "runStarted", runId, total: benchmarks.length });

  for (let i = 0; i < benchmarks.length; i += 1) {
    if (cancelRequested === runId) {
      cancelled = true;
      break;
    }

    const benchmark = benchmarks[i];

    post({ type: "jobStarted", runId, benchmark, index: i });

    const result = api.compare({
      benchmark,
      variants: request.variants,
      dimensions: request.dimensions,
      runs: request.runs,
      iterations: request.iterations,
      seed: request.seed,
    });

    if (!result || result.error) {
      post({
        type: "jobError",
        runId,
        benchmark,
        error: (result && result.error) || "compare returned nothing",
        panic: Boolean(result && result.panic),
        // A panic has aborted the instance: every later call would fail too,
        // so the sweep stops rather than filling the table with errors.
        fatal: Boolean(result && result.panic),
      });

      if (result && result.panic) {
        cancelled = true;
        break;
      }
    } else {
      completed += 1;
      post({ type: "jobResult", runId, benchmark, result });
    }

    post({
      type: "jobProgress",
      runId,
      completed,
      total: benchmarks.length,
      index: i,
    });

    // The yield that makes Stop work. Without it this loop never returns to
    // the event loop and a queued cancel waits for the entire sweep.
    await yieldToLoop();
  }

  post({ type: "runDone", runId, completed, cancelled });
}

self.onmessage = async (event) => {
  const message = event.data || {};

  if (message.type === "cancel") {
    cancelRequested = message.runId;

    return;
  }

  if (message.type !== "run") {
    return;
  }

  if (!api) {
    post({ type: "fatal", error: "worker is not ready" });

    return;
  }

  try {
    await runSweep(message);
  } catch (err) {
    post({ type: "fatal", error: String((err && err.message) || err) });
  }
};

boot().catch((err) => {
  post({ type: "fatal", error: String((err && err.message) || err) });
});
