/*
 * bench.js — the Variant Shootout controller.
 *
 * Two invariants worth stating up front, because both are easy to break:
 *
 *   Stop. The worker's sweep loop yields between bounded Go calls, so a posted
 *   "cancel" is dispatched at the next gap and the run ends with its partial
 *   results intact. worker.terminate() plus a respawn is kept behind a watchdog
 *   for the case where one call runs long enough that the gap never comes.
 *
 *   runId. Every sweep carries a monotonic id and every worker message is
 *   checked against it. A re-run started while a stale worker is still draining
 *   its queue must not append rows to the new sweep's table.
 *
 * This page owns no wasm instance of its own. Everything statistical — the
 * means, the ranks, the Wilcoxon and Friedman tests — is computed by the Go
 * library in the worker and arrives here as finished numbers.
 */
(function () {
  "use strict";

  const el = (id) => document.getElementById(id);

  const rack = el("rack");
  const statusEl = el("status");
  const swarmRing = el("swarmRing");

  const effortSelect = el("effort");
  const runsInput = el("runs");
  const iterationsInput = el("iterations");
  const dimensionsInput = el("dimensions");
  const seedInput = el("seed");
  const benchmarkChips = el("benchmarkChips");
  const variantChips = el("variantChips");
  const startButton = el("start");
  const stopButton = el("stop");
  const estimateEl = el("estimate");
  const progressBar = el("progressBar");
  const progressText = el("progressText");

  const chartCanvas = el("chart");
  const chartLegend = el("chartLegend");
  const metricSelect = el("metric");
  const statsTable = el("statsTable");
  const matrixEl = el("matrix");
  const friedmanEl = el("friedman");
  const exportCsvButton = el("exportCsv");
  const exportJsonButton = el("exportJson");
  const liveRegion = el("liveRegion");
  const buildInfo = el("buildInfo");

  const CANCEL_TIMEOUT_MS = 4000;
  const LIVE_THROTTLE_MS = 700;

  // Effort presets trade confidence for wall-clock. The library's own default
  // is 30 runs; "quick" is explicitly below what proves anything, and the page
  // says so in its caveats rather than hiding it.
  const EFFORT = {
    quick: { runs: 5, iterations: 200, dimensions: 10 },
    standard: { runs: 12, iterations: 400, dimensions: 10 },
    thorough: { runs: 30, iterations: 600, dimensions: 20 },
  };

  const DEFAULT_BENCHMARKS = ["Sphere", "Rastrigin", "Rosenbrock", "Ackley"];

  const state = {
    info: null,
    worker: null,
    ready: false,
    runId: 0,
    running: false,
    results: [],
    rows: [],
    hidden: new Set(),
    sort: { key: "benchmark", direction: 1 },
    lastAnnounce: 0,
    cancelTimer: null,
  };

  function setStatus(message, tone) {
    statusEl.textContent = message;
    statusEl.dataset.state = tone || "";
  }

  function announce(message) {
    const now = Date.now();

    if (now - state.lastAnnounce < LIVE_THROTTLE_MS) {
      return;
    }

    state.lastAnnounce = now;
    liveRegion.textContent = message;
  }

  function intValue(input, fallback) {
    const value = parseInt(input.value, 10);

    return Number.isFinite(value) ? value : fallback;
  }

  // --- chips -------------------------------------------------------------

  function buildChips(container, items, selected) {
    container.innerHTML = "";

    for (const item of items) {
      const chip = document.createElement("button");
      chip.type = "button";
      chip.className = "chip";
      chip.textContent = item.label;
      chip.dataset.value = item.value;
      chip.title = item.title || item.label;
      chip.setAttribute("aria-pressed", String(selected.includes(item.value)));

      chip.addEventListener("click", () => {
        const pressed = chip.getAttribute("aria-pressed") === "true";
        chip.setAttribute("aria-pressed", String(!pressed));
        updateEstimate();
      });

      container.append(chip);
    }
  }

  function chipValues(container) {
    return Array.from(container.querySelectorAll(".chip"))
      .filter((chip) => chip.getAttribute("aria-pressed") === "true")
      .map((chip) => chip.dataset.value);
  }

  // --- estimate ----------------------------------------------------------

  // A crude model, calibrated on one machine, and labelled as crude on the
  // page. Its job is only to stop someone launching a twenty-minute sweep by
  // accident.
  function updateEstimate() {
    const benchmarks = chipValues(benchmarkChips);
    const variants = chipValues(variantChips);
    const runs = intValue(runsInput, 5);
    const iterations = intValue(iterationsInput, 200);
    const dimensions = intValue(dimensionsInput, 10);
    const trials = benchmarks.length * variants.length * runs;
    const seconds = (trials * iterations * dimensions * 40) / 1e6;

    startButton.disabled = !state.ready || !benchmarks.length || !variants.length;

    if (!benchmarks.length || !variants.length) {
      estimateEl.textContent = "Select at least one function and one variant.";

      return;
    }

    estimateEl.innerHTML = `<b>${trials.toLocaleString("en-US")}</b> optimization runs — roughly ${
      seconds < 1 ? "under a second" : `${Math.ceil(seconds)} s`
    }. Rough estimate.`;
  }

  // --- sweep -------------------------------------------------------------

  function startSweep() {
    if (state.running || !state.ready) {
      return;
    }

    const benchmarks = chipValues(benchmarkChips);
    const variants = chipValues(variantChips);

    if (!benchmarks.length || !variants.length) {
      return;
    }

    state.runId += 1;
    state.running = true;
    state.results = [];
    state.rows = [];

    renderAll();

    startButton.disabled = true;
    stopButton.disabled = false;
    exportCsvButton.disabled = true;
    exportJsonButton.disabled = true;

    progressBar.style.width = "0%";
    progressText.textContent = `0 / ${benchmarks.length} functions`;
    setStatus("Running sweep…", "loading");

    state.worker.postMessage({
      type: "run",
      runId: state.runId,
      benchmarks,
      variants,
      runs: intValue(runsInput, 5),
      iterations: intValue(iterationsInput, 200),
      dimensions: intValue(dimensionsInput, 10),
      seed: intValue(seedInput, 42),
      target: 1e-8,
    });
  }

  function stopSweep() {
    if (!state.running) {
      return;
    }

    stopButton.disabled = true;
    setStatus("Stopping after the current function…", "loading");
    state.worker.postMessage({ type: "cancel", runId: state.runId });

    // If the gap between calls never comes — one benchmark running far longer
    // than expected — the worker is killed and replaced instead of leaving the
    // page stuck with a dead Stop button.
    window.clearTimeout(state.cancelTimer);
    state.cancelTimer = window.setTimeout(() => {
      if (!state.running) {
        return;
      }

      setStatus("Worker did not yield; restarting it.", "error");
      respawnWorker();
      finishSweep(true);
    }, CANCEL_TIMEOUT_MS);
  }

  function finishSweep(cancelled) {
    window.clearTimeout(state.cancelTimer);
    state.running = false;
    startButton.disabled = false;
    stopButton.disabled = true;

    const done = state.results.length;
    exportCsvButton.disabled = done === 0;
    exportJsonButton.disabled = done === 0;

    setStatus(
      cancelled
        ? `Stopped — ${done} function${done === 1 ? "" : "s"} completed`
        : `Sweep complete — ${done} function${done === 1 ? "" : "s"}`,
      done ? "ready" : "error",
    );
    liveRegion.textContent = cancelled ? "Sweep stopped" : "Sweep complete";
  }

  // --- worker ------------------------------------------------------------

  function handleMessage(event) {
    const message = event.data || {};

    if (message.type === "ready") {
      state.info = message.info;
      state.ready = true;
      swarmRing.dataset.state = "ready";
      rack.dataset.boot = "ready";
      Render.ring(swarmRing, 1);

      buildChips(
        benchmarkChips,
        message.info.benchmarks.map((b) => ({
          value: b.name,
          label: b.name,
          title: b.blurb,
        })),
        DEFAULT_BENCHMARKS,
      );

      buildChips(
        variantChips,
        message.info.variants.map((v) => ({
          value: v.key,
          label: v.name,
          title: v.fullName,
        })),
        message.info.variants.map((v) => v.key),
      );

      buildInfo.textContent = `${message.info.goVersion} · ${message.info.goos}/${message.info.goarch}`;
      setStatus("WASM ready", "ready");
      updateEstimate();

      return;
    }

    if (message.type === "fatal") {
      setStatus(message.error, "error");
      finishSweep(true);

      return;
    }

    // Everything below belongs to a specific sweep. A message from a stale
    // worker still draining its queue is dropped rather than mixed in.
    if (message.runId !== state.runId) {
      return;
    }

    if (message.type === "jobStarted") {
      announce(`Running ${message.benchmark}`);
      progressText.textContent = `${message.benchmark}…`;

      return;
    }

    if (message.type === "jobResult") {
      state.results.push(message.result);
      appendRows(message.result);
      renderAll();

      return;
    }

    if (message.type === "jobError") {
      setStatus(message.error, "error");

      return;
    }

    if (message.type === "jobProgress") {
      const percent = (message.completed / message.total) * 100;
      progressBar.style.width = `${percent}%`;
      progressText.textContent = `${message.completed} / ${message.total} functions`;

      return;
    }

    if (message.type === "runDone") {
      progressBar.style.width = "100%";
      finishSweep(message.cancelled);
    }
  }

  function respawnWorker() {
    if (state.worker) {
      state.worker.terminate();
    }

    state.worker = new Worker("bench-worker.js");
    state.worker.onmessage = handleMessage;
    state.worker.onerror = (err) => {
      console.error(err);
      setStatus(`worker error: ${err.message || "unknown"}`, "error");
    };
  }

  // --- table -------------------------------------------------------------

  function appendRows(result) {
    for (const stat of result.statistics) {
      state.rows.push({
        benchmark: result.benchmark,
        algorithm: stat.algorithm,
        rank: stat.rank,
        mean: stat.mean,
        median: stat.median,
        stdDev: stat.stdDev,
        best: stat.best,
        worst: stat.worst,
        successRate: stat.successRate,
        avgFuncEvals: stat.avgFuncEvals,
        avgTime: stat.avgTime === null ? null : stat.avgTime * 1000,
      });
    }
  }

  function renderTable() {
    const body = statsTable.querySelector("tbody");
    const { key, direction } = state.sort;

    const sorted = state.rows.slice().sort((a, b) => {
      const left = a[key];
      const right = b[key];

      if (typeof left === "string" || typeof right === "string") {
        return String(left).localeCompare(String(right)) * direction;
      }

      const safeLeft = left === null || !isFinite(left) ? Infinity : left;
      const safeRight = right === null || !isFinite(right) ? Infinity : right;

      return (safeLeft - safeRight) * direction;
    });

    body.innerHTML = "";

    for (const row of sorted) {
      const tr = document.createElement("tr");

      const cells = [
        row.benchmark,
        row.algorithm,
        String(row.rank),
        Render.compact(row.mean),
        Render.compact(row.median),
        Render.compact(row.stdDev),
        Render.compact(row.best),
        Render.compact(row.worst),
        row.successRate === null ? "—" : `${row.successRate.toFixed(0)}%`,
        row.avgFuncEvals === null ? "—" : Math.round(row.avgFuncEvals).toLocaleString("en-US"),
        row.avgTime === null ? "—" : row.avgTime.toFixed(1),
      ];

      cells.forEach((value, index) => {
        const td = document.createElement("td");
        td.textContent = value;

        if (index === 2) {
          td.dataset.rank = String(row.rank);
        }

        tr.append(td);
      });

      body.append(tr);
    }

    for (const th of statsTable.querySelectorAll("thead th")) {
      if (th.dataset.key === key) {
        th.setAttribute("aria-sort", direction === 1 ? "ascending" : "descending");
      } else {
        th.removeAttribute("aria-sort");
      }
    }
  }

  // --- chart -------------------------------------------------------------

  function algorithmNames() {
    if (!state.results.length) {
      return [];
    }

    return state.results[0].algorithms;
  }

  function renderChart() {
    const algorithms = algorithmNames();
    const metric = metricSelect.value;

    const groups = state.results.map((result) => ({
      label: result.benchmark,
      values: result.statistics.map((stat) => stat[metric]),
    }));

    BenchChart.draw(
      chartCanvas,
      { groups, algorithms },
      { hidden: state.hidden, yLabel: `${metric} best cost — log scale, shorter is better` },
    );

    renderChartLegend(algorithms);
  }

  function renderChartLegend(algorithms) {
    if (chartLegend.childElementCount === algorithms.length) {
      return;
    }

    chartLegend.innerHTML = "";

    algorithms.forEach((name, index) => {
      const chip = document.createElement("button");
      chip.type = "button";
      chip.className = "chip";
      chip.textContent = name;
      chip.style.color = BenchChart.seriesColor(index);
      chip.style.borderColor = BenchChart.seriesColor(index);
      chip.setAttribute("aria-pressed", String(!state.hidden.has(name)));

      chip.addEventListener("click", () => {
        if (state.hidden.has(name)) {
          state.hidden.delete(name);
        } else {
          state.hidden.add(name);
        }

        chip.setAttribute("aria-pressed", String(!state.hidden.has(name)));
        renderChart();
      });

      chartLegend.append(chip);
    });
  }

  // --- significance ------------------------------------------------------

  function renderSignificance() {
    if (!state.results.length) {
      friedmanEl.textContent = "Run a sweep to see the tests.";
      matrixEl.innerHTML = "";

      return;
    }

    // The matrix is shown for the last completed function; showing one per
    // function would be seven tables nobody reads.
    const result = state.results[state.results.length - 1];
    const friedman = result.friedman;

    if (friedman) {
      friedmanEl.innerHTML = friedman.significant
        ? `<b>${result.benchmark}:</b> the variants are not equivalent — Friedman χ² = ${Render.compact(
            friedman.chiSquare,
          )}, df ${friedman.degreesOfFreedom}, p = ${Render.compact(friedman.pValue)}. Best here: <b>${
            result.algorithms[result.best]
          }</b>.`
        : `<b>${result.benchmark}:</b> no significant difference between variants — Friedman χ² = ${Render.compact(
            friedman.chiSquare,
          )}, df ${friedman.degreesOfFreedom}, p = ${Render.compact(
            friedman.pValue,
          )}. With ${result.runs} runs that is weak evidence either way.`;
    }

    const algorithms = result.algorithms;
    const table = document.createElement("table");
    table.className = "matrix";

    const head = document.createElement("thead");
    const headRow = document.createElement("tr");
    headRow.append(document.createElement("th"));

    for (const name of algorithms) {
      const th = document.createElement("th");
      th.textContent = name;
      headRow.append(th);
    }

    head.append(headRow);
    table.append(head);

    const body = document.createElement("tbody");

    algorithms.forEach((rowName, i) => {
      const tr = document.createElement("tr");
      const label = document.createElement("td");
      label.textContent = rowName;
      tr.append(label);

      algorithms.forEach((columnName, j) => {
        const td = document.createElement("td");

        if (i === j) {
          td.textContent = "·";
          td.className = "tie";
        } else {
          const test = result.wilcoxon[i] && result.wilcoxon[i][j];

          if (!test) {
            td.textContent = "—";
            td.className = "tie";
          } else if (!test.significant) {
            td.textContent = "=";
            td.className = "tie";
            td.title = `p = ${Render.compact(test.pValue)} — not significant`;
          } else {
            const won = test.winner === rowName;
            td.textContent = won ? "win" : "loss";
            td.className = won ? "win" : "lose";
            td.title = `${test.winner} wins, p = ${Render.compact(test.pValue)}`;
          }
        }

        tr.append(td);
      });

      body.append(tr);
    });

    table.append(body);
    matrixEl.innerHTML = "";
    matrixEl.append(table);
  }

  function renderAll() {
    renderTable();
    renderChart();
    renderSignificance();
  }

  // --- export ------------------------------------------------------------

  function download(filename, text, type) {
    const blob = new Blob([text], { type });
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = filename;
    anchor.click();
    URL.revokeObjectURL(url);
  }

  function exportCsv() {
    const header = [
      "benchmark",
      "variant",
      "rank",
      "mean",
      "median",
      "std_dev",
      "best",
      "worst",
      "success_rate",
      "avg_func_evals",
      "avg_time_ms",
    ];

    const lines = [header.join(",")];

    for (const row of state.rows) {
      lines.push(
        [
          row.benchmark,
          row.algorithm,
          row.rank,
          row.mean,
          row.median,
          row.stdDev,
          row.best,
          row.worst,
          row.successRate,
          row.avgFuncEvals,
          row.avgTime,
        ].join(","),
      );
    }

    download("mayfly-shootout.csv", lines.join("\n"), "text/csv");
  }

  function exportJson() {
    download(
      "mayfly-shootout.json",
      JSON.stringify(
        {
          generated: new Date().toISOString(),
          settings: {
            runs: intValue(runsInput, 5),
            iterations: intValue(iterationsInput, 200),
            dimensions: intValue(dimensionsInput, 10),
            seed: intValue(seedInput, 42),
          },
          results: state.results,
        },
        null,
        2,
      ),
      "application/json",
    );
  }

  // --- wiring ------------------------------------------------------------

  function wire() {
    startButton.addEventListener("click", startSweep);
    stopButton.addEventListener("click", stopSweep);
    metricSelect.addEventListener("change", renderChart);
    exportCsvButton.addEventListener("click", exportCsv);
    exportJsonButton.addEventListener("click", exportJson);

    effortSelect.addEventListener("change", () => {
      const preset = EFFORT[effortSelect.value];

      if (preset) {
        runsInput.value = String(preset.runs);
        iterationsInput.value = String(preset.iterations);
        dimensionsInput.value = String(preset.dimensions);
      }

      updateEstimate();
    });

    for (const input of [runsInput, iterationsInput, dimensionsInput, seedInput]) {
      input.addEventListener("change", updateEstimate);
    }

    for (const th of statsTable.querySelectorAll("thead th")) {
      th.addEventListener("click", () => {
        const key = th.dataset.key;

        if (state.sort.key === key) {
          state.sort.direction *= -1;
        } else {
          state.sort = { key, direction: 1 };
        }

        renderTable();
      });
    }

    let resizeTimer = null;

    window.addEventListener("resize", () => {
      window.clearTimeout(resizeTimer);
      resizeTimer = window.setTimeout(renderChart, 120);
    });
  }

  // --- boot --------------------------------------------------------------

  Render.ring(swarmRing, 0.15);
  setStatus("Starting worker…", "loading");
  respawnWorker();
  wire();
  renderChart();
})();
