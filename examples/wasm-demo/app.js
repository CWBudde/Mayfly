/*
 * app.js — the Swarm Lab controller.
 *
 * Its only jobs are: collect control values, ask Go for a run, and replay the
 * history Go returns. It never computes a position, a cost, or a statistic.
 * If you find yourself about to write a benchmark function here, that is the
 * signal to add an export in the Go bridge instead — a demo that reimplements
 * the library in JavaScript demonstrates the JavaScript.
 *
 * No modules; this file is an IIFE and depends only on window.Render.
 */
(function () {
  "use strict";

  // --- DOM ---------------------------------------------------------------

  const el = (id) => document.getElementById(id);

  const rack = el("rack");
  const statusEl = el("status");
  const swarmRing = el("swarmRing");

  const stage = el("stage");
  const convergenceCanvas = el("convergence");
  const diversityCanvas = el("diversity");
  const spreadCanvas = el("spread");

  const benchmarkSelect = el("benchmark");
  const variantSelect = el("variant");
  const compareSelect = el("compareVariant");
  const dimensionsInput = el("dimensions");
  const npopInput = el("npop");
  const npopfInput = el("npopf");
  const iterationsInput = el("iterations");
  const seedInput = el("seed");
  const lowerInput = el("lower");
  const upperInput = el("upper");

  const runButton = el("run");
  const rerunButton = el("rerun");
  const newSeedButton = el("newSeed");

  const playButton = el("play");
  const scrub = el("scrub");
  const speedSelect = el("speed");
  const frameReadout = el("frameReadout");

  const benchNote = el("benchNote");
  const variantNote = el("variantNote");
  const buildInfo = el("buildInfo");

  const telemetry = {
    best: el("tBest"),
    evals: el("tEvals"),
    iterations: el("tIterations"),
    termination: el("tTermination"),
    seed: el("tSeed"),
    diversity: el("tDiversity"),
  };

  const reducedMotion =
    window.matchMedia &&
    window.matchMedia("(prefers-reduced-motion: reduce)").matches;

  // --- state -------------------------------------------------------------

  const state = {
    ready: false,
    info: null,
    land: null,
    run: null,
    compare: null,
    frame: 0,
    playing: false,
    lastTick: 0,
  };

  // Reusable views over JS-owned ArrayBuffers. Go copies into these instead of
  // allocating a new typed array per call; see marshal.go for why the buffer
  // has to be owned by this side.
  const sinks = { run: {}, compare: {}, land: {} };

  function setStatus(message, tone) {
    statusEl.textContent = message;
    statusEl.dataset.state = tone || "";
  }

  // --- the wasm call wrapper ---------------------------------------------

  // Every call into Go goes through here. A missing export, a thrown error and
  // an {error} result all become the same thing: a message on the status line
  // and a null return. Nothing from the wasm side is ever allowed to throw into
  // the render loop, because a half-drawn frame is much harder to diagnose than
  // a status line that says what went wrong.
  function call(name, opts, callOpts) {
    const silent = callOpts && callOpts.silent;
    const api = window.mayfly;

    if (!api || typeof api[name] !== "function") {
      if (!silent) {
        setStatus(`export "${name}" is unavailable`, "error");
      }

      return null;
    }

    let result;

    try {
      result = api[name](opts);
    } catch (err) {
      console.error(err);

      if (!silent) {
        setStatus(`${name} failed: ${err && err.message}`, "error");
      }

      return null;
    }

    if (result && result.error) {
      console.error(result.error);

      if (!silent) {
        setStatus(result.error, "error");
      }

      return null;
    }

    return result;
  }

  // --- control readers ---------------------------------------------------

  function intValue(input, fallback) {
    const value = parseInt(input.value, 10);

    return Number.isFinite(value) ? value : fallback;
  }

  function floatValue(input, fallback) {
    const value = parseFloat(input.value);

    return Number.isFinite(value) ? value : fallback;
  }

  function benchmarkSpec(name) {
    if (!state.info) {
      return null;
    }

    return state.info.benchmarks.find((b) => b.name === name) || null;
  }

  function currentRequest(variantName, out) {
    return {
      benchmark: benchmarkSelect.value,
      variant: variantName,
      dimensions: intValue(dimensionsInput, 2),
      iterations: intValue(iterationsInput, 60),
      npop: intValue(npopInput, 20),
      npopf: intValue(npopfInput, 20),
      seed: intValue(seedInput, 42),
      lower: floatValue(lowerInput, -5.12),
      upper: floatValue(upperInput, 5.12),
      axisX: 0,
      axisY: 1,
      out,
    };
  }

  // --- population of controls from info() --------------------------------

  // The <select>s in the HTML ship empty. Everything in them comes from the Go
  // side's capability table, so adding a variant to the library puts it in the
  // UI without anyone editing markup.
  function populateControls() {
    const info = state.info;

    benchmarkSelect.innerHTML = "";

    for (const spec of info.benchmarks) {
      const option = document.createElement("option");
      option.value = spec.name;
      option.textContent = spec.name;
      benchmarkSelect.append(option);
    }

    benchmarkSelect.value = "Rastrigin";

    variantSelect.innerHTML = "";
    compareSelect.innerHTML = "";

    const none = document.createElement("option");
    none.value = "";
    none.textContent = "— off —";
    compareSelect.append(none);

    for (const variant of info.variants) {
      for (const target of [variantSelect, compareSelect]) {
        const option = document.createElement("option");
        option.value = variant.key;
        option.textContent = variant.name;
        option.title = variant.fullName;
        target.append(option);
      }
    }

    variantSelect.value = "MA";
    compareSelect.value = "";

    dimensionsInput.max = String(info.maxDimensions);
    iterationsInput.max = String(info.maxIterations);
    npopInput.max = String(info.maxPopulation);
    npopfInput.max = String(info.maxPopulation);

    buildInfo.textContent = `${info.goVersion} · ${info.goos}/${info.goarch}`;

    applyBenchmarkDefaults();
    updateVariantNote();
  }

  function applyBenchmarkDefaults() {
    const spec = benchmarkSpec(benchmarkSelect.value);

    if (!spec) {
      return;
    }

    lowerInput.value = String(spec.lower);
    upperInput.value = String(spec.upper);

    benchNote.innerHTML = `<b>${spec.modality}.</b> ${spec.blurb} Optimum ${Render.compact(spec.optimum)}.`;
  }

  function updateVariantNote() {
    if (!state.info) {
      return;
    }

    const variant = state.info.variants.find((v) => v.key === variantSelect.value);

    if (!variant) {
      return;
    }

    const overhead = Math.round((variant.overhead - 1) * 100);
    const cost =
      overhead > 0
        ? ` About ${overhead}% more work than standard MA.`
        : " No measurable overhead over standard MA.";

    variantNote.innerHTML = `<b>${variant.fullName}.</b> ${variant.description}${cost}`;
  }

  // --- running -----------------------------------------------------------

  function refreshLandscape() {
    const spec = benchmarkSpec(benchmarkSelect.value);

    if (!spec) {
      return;
    }

    const land = call("landscape", {
      benchmark: benchmarkSelect.value,
      dimensions: intValue(dimensionsInput, 2),
      lower: floatValue(lowerInput, spec.lower),
      upper: floatValue(upperInput, spec.upper),
      width: 180,
      height: 180,
      mode: "rank",
      out: sinks.land,
    });

    if (land) {
      state.land = land;
      cacheSinks(sinks.land, land, ["values", "raw"]);
    }
  }

  // cacheSinks remembers the views Go handed back so the next call can reuse
  // their buffers. A returned view may be a subarray of the buffer, so the
  // whole ArrayBuffer is re-wrapped rather than the view stored directly.
  function cacheSinks(store, result, keys) {
    for (const key of keys) {
      const view = result[key];

      if (view && view.buffer) {
        store[key] = {
          f32: new Float32Array(view.buffer),
          u8: new Uint8Array(view.buffer),
        };
      }
    }
  }

  const RUN_ARRAYS = [
    "convergence",
    "males",
    "females",
    "maleCost",
    "femaleCost",
    "bestTrail",
    "maleDiversity",
    "femaleDiversity",
  ];

  function doRun() {
    if (!state.ready) {
      return;
    }

    const started = performance.now();

    refreshLandscape();

    const run = call("run", currentRequest(variantSelect.value, sinks.run));

    if (!run) {
      return;
    }

    cacheSinks(sinks.run, run, RUN_ARRAYS);
    state.run = run;

    if (compareSelect.value) {
      const compare = call(
        "run",
        currentRequest(compareSelect.value, sinks.compare),
      );

      if (compare) {
        cacheSinks(sinks.compare, compare, RUN_ARRAYS);
      }

      state.compare = compare;
    } else {
      state.compare = null;
    }

    const elapsed = performance.now() - started;

    scrub.max = String(Math.max(0, run.iterations - 1));
    scrub.disabled = false;
    playButton.disabled = false;

    setFrame(run.iterations - 1);
    updateTelemetry(run, elapsed);
    drawAll();

    // Autoplay from the start is the point of the page, but it is motion the
    // user did not ask for, so it is skipped when they have asked for less.
    if (!reducedMotion) {
      setFrame(0);
      setPlaying(true);
    }
  }

  function updateTelemetry(run, elapsed) {
    telemetry.best.textContent = Render.compact(run.bestCost);
    telemetry.best.dataset.tone =
      run.optimum !== null && Math.abs(run.bestCost - run.optimum) < 1e-6
        ? "good"
        : "";
    telemetry.evals.textContent = run.evaluations.toLocaleString("en-US");
    telemetry.iterations.textContent = String(run.iterations);
    telemetry.termination.textContent = run.terminationReason.replace(/_/g, " ");
    telemetry.seed.textContent = String(run.seed);

    setStatus(
      `${run.variant} on ${run.benchmark} — ${run.iterations} iterations in ${elapsed.toFixed(1)} ms`,
      "ready",
    );
  }

  // --- replay ------------------------------------------------------------

  function setFrame(index) {
    const run = state.run;

    if (!run) {
      return;
    }

    const clamped = Math.max(0, Math.min(index, run.iterations - 1));
    state.frame = clamped;
    scrub.value = String(clamped);
    frameReadout.textContent = `${clamped + 1} / ${run.iterations}`;

    const spread = run.maleDiversity ? run.maleDiversity[clamped] : null;
    telemetry.diversity.textContent = Render.compact(spread);
  }

  function setPlaying(playing) {
    state.playing = playing;
    playButton.setAttribute("aria-pressed", String(playing));
    playButton.textContent = playing ? "Pause" : "Play";
    state.lastTick = performance.now();
  }

  function drawAll() {
    const run = state.run;

    if (!run) {
      return;
    }

    Render.drawSwarm(stage, {
      land: state.land,
      run,
      compare: state.compare,
      frameIndex: state.frame,
    });

    const series = [
      {
        values: run.convergence,
        color: Render.readVar("--best", "#ff5d8f"),
        width: 1.8,
      },
    ];

    if (state.compare) {
      series.push({
        values: state.compare.convergence,
        color: Render.readVar("--elite", "#8b7dff"),
        dash: [5, 4],
        width: 1.6,
      });
    }

    Render.drawSeries(convergenceCanvas, series, {
      log: true,
      marker: state.frame,
    });

    Render.drawSeries(
      diversityCanvas,
      [
        {
          values: run.maleDiversity,
          color: Render.readVar("--male", "#ffb347"),
        },
        {
          values: run.femaleDiversity,
          color: Render.readVar("--female", "#35d6c4"),
        },
      ],
      { log: false, marker: state.frame },
    );

    Render.drawSpread(spreadCanvas, { run, frameIndex: state.frame });
  }

  // --- animation ---------------------------------------------------------

  // Frames advance on wall-clock time rather than once per rAF, so the replay
  // runs at the same speed on a 60 Hz and a 144 Hz display.
  const BASE_FPS = 24;

  function tick(now) {
    if (state.playing && state.run) {
      const speed = parseFloat(speedSelect.value) || 1;
      const interval = 1000 / (BASE_FPS * speed);

      if (now - state.lastTick >= interval) {
        state.lastTick = now;

        if (state.frame >= state.run.iterations - 1) {
          setPlaying(false);
        } else {
          setFrame(state.frame + 1);
          drawAll();
        }
      }
    }

    requestAnimationFrame(tick);
  }

  // --- controls ----------------------------------------------------------

  function wireControls() {
    runButton.addEventListener("click", doRun);
    rerunButton.addEventListener("click", doRun);

    newSeedButton.addEventListener("click", () => {
      seedInput.value = String(Math.floor(Math.random() * 100000));
      doRun();
    });

    benchmarkSelect.addEventListener("change", () => {
      applyBenchmarkDefaults();
      doRun();
    });

    variantSelect.addEventListener("change", () => {
      updateVariantNote();
      doRun();
    });

    compareSelect.addEventListener("change", doRun);

    for (const input of [
      dimensionsInput,
      npopInput,
      npopfInput,
      iterationsInput,
      lowerInput,
      upperInput,
      seedInput,
    ]) {
      input.addEventListener("change", doRun);
    }

    playButton.addEventListener("click", () => {
      if (!state.run) {
        return;
      }

      // Pressing play at the end restarts rather than doing nothing.
      if (!state.playing && state.frame >= state.run.iterations - 1) {
        setFrame(0);
      }

      setPlaying(!state.playing);
    });

    scrub.addEventListener("input", () => {
      setPlaying(false);
      setFrame(parseInt(scrub.value, 10) || 0);
      drawAll();
    });

    let resizeTimer = null;

    window.addEventListener("resize", () => {
      window.clearTimeout(resizeTimer);
      resizeTimer = window.setTimeout(drawAll, 120);
    });

    Render.watchDPR(() => {
      Render.invalidateTheme();
      drawAll();
    });
  }

  // --- boot --------------------------------------------------------------

  async function loadWasmWithProgress(onProgress) {
    if (!WebAssembly.instantiateStreaming) {
      WebAssembly.instantiateStreaming = async (resp, importObject) => {
        const source = await (await resp).arrayBuffer();

        return WebAssembly.instantiate(source, importObject);
      };
    }

    const go = new Go();
    const response = await fetch("mayfly.wasm");

    if (!response.ok) {
      throw new Error(`fetch mayfly.wasm: ${response.status}`);
    }

    if (!response.body || !response.body.getReader || reducedMotion) {
      onProgress(1);

      return {
        go,
        result: await WebAssembly.instantiateStreaming(response, go.importObject),
      };
    }

    const total = Number(response.headers.get("content-length")) || 0;
    const reader = response.body.getReader();
    const chunks = [];
    let received = 0;

    for (;;) {
      const { done, value } = await reader.read();

      if (done) {
        break;
      }

      chunks.push(value);
      received += value.length;

      if (total > 0) {
        onProgress(Math.min(0.98, received / total));
      }
    }

    onProgress(1);

    const bytes = new Uint8Array(received);
    let offset = 0;

    for (const chunk of chunks) {
      bytes.set(chunk, offset);
      offset += chunk.length;
    }

    return {
      go,
      result: await WebAssembly.instantiate(bytes, go.importObject),
    };
  }

  async function initWasm() {
    setStatus("Loading WebAssembly…", "loading");

    const { go, result } = await loadWasmWithProgress((progress) => {
      Render.ring(swarmRing, progress);
    });

    // Deliberately not awaited: the demo's main() ends in select{} so this
    // promise never resolves. Awaiting it would hang the page forever.
    go.run(result.instance);

    // Give the Go side one turn of the event loop to publish globalThis.mayfly.
    await new Promise((resolve) => setTimeout(resolve, 0));

    const info = call("info", undefined);

    if (!info) {
      throw new Error("the wasm module did not publish its capability table");
    }

    state.info = info;
    state.ready = true;

    swarmRing.dataset.state = "ready";
    rack.dataset.boot = "ready";

    populateControls();
    wireControls();

    runButton.disabled = false;
    rerunButton.disabled = false;
    newSeedButton.disabled = false;

    setStatus("WASM ready", "ready");
    requestAnimationFrame(tick);
    doRun();
  }

  initWasm().catch((err) => {
    console.error(err);
    setStatus(
      "WebAssembly failed to load. Serve this page over HTTP and check that mayfly.wasm is sent as application/wasm.",
      "error",
    );
  });
})();
