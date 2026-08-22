/*
 * render.js — every pixel the Swarm Lab draws.
 *
 * This file owns colour. It reads the swarm hues and the landscape ramp out of
 * the page's CSS custom properties rather than hardcoding them, so the canvas
 * and the stylesheet cannot drift apart: change --male in style.css and the
 * glyphs, the legend and the diversity curve all follow.
 *
 * It owns no state about the run. Everything it draws is passed in, so the
 * animation loop can render any frame in any order — which is what makes the
 * scrubber work.
 *
 * No modules — everything hangs off window.Render.
 */
(function () {
  "use strict";

  // --- theme -------------------------------------------------------------

  let varCache = null;

  function readVar(name, fallback) {
    if (!varCache) {
      varCache = new Map();
    }

    if (varCache.has(name)) {
      return varCache.get(name);
    }

    const raw = getComputedStyle(document.documentElement)
      .getPropertyValue(name)
      .trim();
    const value = raw || fallback;
    varCache.set(name, value);

    return value;
  }

  function invalidateTheme() {
    varCache = null;
  }

  // parseColor handles the #rgb / #rrggbb forms this stylesheet actually uses.
  // Anything else falls back to mid grey rather than throwing inside a draw
  // call, because a broken colour should not take the whole canvas down.
  function parseColor(input) {
    const value = String(input || "").trim();
    const short = /^#([0-9a-f])([0-9a-f])([0-9a-f])$/i.exec(value);

    if (short) {
      return [
        parseInt(short[1] + short[1], 16),
        parseInt(short[2] + short[2], 16),
        parseInt(short[3] + short[3], 16),
      ];
    }

    const long = /^#([0-9a-f]{2})([0-9a-f]{2})([0-9a-f]{2})$/i.exec(value);

    if (long) {
      return [
        parseInt(long[1], 16),
        parseInt(long[2], 16),
        parseInt(long[3], 16),
      ];
    }

    return [128, 128, 128];
  }

  function mix(a, b, t) {
    return [
      Math.round(a[0] + (b[0] - a[0]) * t),
      Math.round(a[1] + (b[1] - a[1]) * t),
      Math.round(a[2] + (b[2] - a[2]) * t),
    ];
  }

  // landColor maps a normalised cost in [0,1] onto the three-stop landscape
  // ramp. Low cost is dark, so the basins the swarm is hunting read as wells
  // and the bright glyphs stay legible on top of them.
  function landColor(t) {
    const low = parseColor(readVar("--land-low", "#0d1230"));
    const mid = parseColor(readVar("--land-mid", "#2b2a63"));
    const high = parseColor(readVar("--land-high", "#6a5aa8"));

    if (t <= 0.5) {
      return mix(low, mid, t * 2);
    }

    return mix(mid, high, (t - 0.5) * 2);
  }

  function alpha(color, a) {
    const [r, g, b] = parseColor(color);

    return `rgba(${r}, ${g}, ${b}, ${a})`;
  }

  // --- canvas sizing -----------------------------------------------------

  function currentDPR() {
    return Math.min(window.devicePixelRatio || 1, 2);
  }

  // fitCanvas sizes the backing store to the CSS box times DPR, and returns a
  // context already scaled so every draw call below can work in CSS pixels.
  function fitCanvas(canvas) {
    const dpr = currentDPR();
    const rect = canvas.getBoundingClientRect();
    const width = Math.max(1, Math.round(rect.width));
    const height = Math.max(1, Math.round(rect.height));

    // Round before comparing. canvas.width is an integer attribute, so at a
    // fractional device pixel ratio (1.5 on many Windows displays) the raw
    // product never equals it, the comparison stayed true on every animation
    // frame, and the backing store was reallocated and cleared 60 times a
    // second.
    const backingWidth = Math.round(width * dpr);
    const backingHeight = Math.round(height * dpr);

    if (canvas.width !== backingWidth || canvas.height !== backingHeight) {
      canvas.width = backingWidth;
      canvas.height = backingHeight;
    }

    const ctx = canvas.getContext("2d");
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);

    return { ctx, width, height };
  }

  function clear(ctx, width, height) {
    ctx.clearRect(0, 0, width, height);
  }

  // --- number formatting -------------------------------------------------

  function compact(value) {
    if (value === null || value === undefined || !isFinite(value)) {
      return "—";
    }

    const magnitude = Math.abs(value);

    if (magnitude === 0) {
      return "0";
    }

    if (magnitude < 1e-4 || magnitude >= 1e5) {
      return value.toExponential(2);
    }

    if (magnitude < 1) {
      return value.toFixed(4);
    }

    return value.toFixed(magnitude < 100 ? 3 : 1);
  }

  // --- axis frame --------------------------------------------------------

  function frame(ctx, width, height, pad) {
    ctx.strokeStyle = readVar("--rule", "#232a4a");
    ctx.lineWidth = 1;
    ctx.strokeRect(
      pad.left + 0.5,
      pad.top + 0.5,
      width - pad.left - pad.right - 1,
      height - pad.top - pad.bottom - 1,
    );
  }

  function label(ctx, text, x, y, align, color) {
    ctx.fillStyle = color || readVar("--dim", "#7d86b4");
    ctx.font = '10px "JetBrains Mono", ui-monospace, monospace';
    ctx.textAlign = align || "left";
    ctx.textBaseline = "middle";
    ctx.fillText(text, x, y);
  }

  // --- landscape ---------------------------------------------------------

  // Cached because the landscape only changes when the function or the bounds
  // do, while the swarm on top of it redraws every animation frame. Rebuilding
  // a 160x160 ImageData 60 times a second would dominate the frame budget for
  // a picture that never changed.
  let landCache = { key: null, bitmap: null };

  function landscapeKey(land) {
    return [
      land.benchmark,
      land.width,
      land.height,
      land.lower,
      land.upper,
      land.axisX,
      land.axisY,
      land.dimensions,
      land.mode,
    ].join("|");
  }

  function buildLandscape(land) {
    const key = landscapeKey(land);

    if (landCache.key === key && landCache.bitmap) {
      return landCache.bitmap;
    }

    const { width, height, values } = land;
    const offscreen = document.createElement("canvas");
    offscreen.width = width;
    offscreen.height = height;

    const ctx = offscreen.getContext("2d");
    const image = ctx.createImageData(width, height);
    const data = image.data;

    // Precompute the ramp once per build: 256 lookups instead of one colour
    // mix per pixel.
    const ramp = new Uint8ClampedArray(256 * 3);

    for (let i = 0; i < 256; i += 1) {
      const [r, g, b] = landColor(i / 255);
      ramp[i * 3] = r;
      ramp[i * 3 + 1] = g;
      ramp[i * 3 + 2] = b;
    }

    for (let i = 0; i < width * height; i += 1) {
      const t = Math.max(0, Math.min(1, values[i] || 0));
      const index = Math.round(t * 255) * 3;

      data[i * 4] = ramp[index];
      data[i * 4 + 1] = ramp[index + 1];
      data[i * 4 + 2] = ramp[index + 2];
      data[i * 4 + 3] = 255;
    }

    ctx.putImageData(image, 0, 0);
    landCache = { key, bitmap: offscreen };

    return offscreen;
  }

  function drawLandscape(ctx, land, width, height) {
    const bitmap = buildLandscape(land);

    ctx.imageSmoothingEnabled = true;
    ctx.drawImage(bitmap, 0, 0, width, height);

    drawContours(ctx, land, width, height);
  }

  // Contour bands are drawn from the same normalised samples as the fill. They
  // are what turn a smooth gradient into a landscape you can read distances
  // off, which matters most on the deceptive functions.
  function drawContours(ctx, land, width, height) {
    const { width: cols, height: rows, values } = land;
    const cellW = width / cols;
    const cellH = height / rows;
    const levels = [0.12, 0.25, 0.4, 0.55, 0.7, 0.85];

    ctx.save();
    ctx.strokeStyle = "rgba(255, 255, 255, 0.07)";
    ctx.lineWidth = 1;
    ctx.beginPath();

    for (let row = 0; row < rows - 1; row += 1) {
      for (let col = 0; col < cols - 1; col += 1) {
        const here = values[row * cols + col];
        const right = values[row * cols + col + 1];
        const down = values[(row + 1) * cols + col];

        for (let i = 0; i < levels.length; i += 1) {
          const level = levels[i];

          if ((here < level) !== (right < level)) {
            ctx.moveTo((col + 1) * cellW, row * cellH);
            ctx.lineTo((col + 1) * cellW, (row + 1) * cellH);
          }

          if ((here < level) !== (down < level)) {
            ctx.moveTo(col * cellW, (row + 1) * cellH);
            ctx.lineTo((col + 1) * cellW, (row + 1) * cellH);
          }
        }
      }
    }

    ctx.stroke();
    ctx.restore();
  }

  // --- swarm -------------------------------------------------------------

  function projector(land, width, height) {
    const span = land.upper - land.lower;

    return {
      x: (value) => ((value - land.lower) / span) * width,
      // Canvas y grows downward; the landscape's top row is the high y value.
      y: (value) => height - ((value - land.lower) / span) * height,
    };
  }

  function diamond(ctx, x, y, radius) {
    ctx.moveTo(x, y - radius);
    ctx.lineTo(x + radius, y);
    ctx.lineTo(x, y + radius);
    ctx.lineTo(x - radius, y);
    ctx.closePath();
  }

  // Males are squares, females are circles, the global best is a diamond with
  // a ring. Shape carries the same information as hue, so the picture survives
  // being printed in grey or read by someone who cannot separate amber from
  // teal.
  function drawSwarm(canvas, options) {
    const { ctx, width, height } = fitCanvas(canvas);
    const { land, frameIndex, run, compare } = options;

    clear(ctx, width, height);

    if (!land) {
      return;
    }

    if (compare) {
      // A/B: two landscapes side by side, each with its own swarm, so the same
      // seed on two variants can be compared frame for frame.
      const half = width / 2;

      ctx.save();
      ctx.beginPath();
      ctx.rect(0, 0, half, height);
      ctx.clip();
      drawLandscape(ctx, land, half, height);
      drawPopulations(ctx, land, run, frameIndex, half, height, options);
      ctx.restore();

      ctx.save();
      ctx.translate(half, 0);
      ctx.beginPath();
      ctx.rect(0, 0, half, height);
      ctx.clip();
      drawLandscape(ctx, land, half, height);
      drawPopulations(ctx, land, compare, frameIndex, half, height, options);
      ctx.restore();

      ctx.strokeStyle = readVar("--rule", "#232a4a");
      ctx.beginPath();
      ctx.moveTo(half + 0.5, 0);
      ctx.lineTo(half + 0.5, height);
      ctx.stroke();

      caption(ctx, run.variant, 8, 8, "left", readVar("--lume", "#e9ecff"));
      caption(ctx, compare.variant, width - 8, 8, "right", readVar("--lume", "#e9ecff"));

      return;
    }

    drawLandscape(ctx, land, width, height);
    drawPopulations(ctx, land, run, frameIndex, width, height, options);
  }

  function caption(ctx, text, x, y, align, color) {
    ctx.fillStyle = color;
    ctx.font = '11px "JetBrains Mono", ui-monospace, monospace';
    ctx.textAlign = align;
    ctx.textBaseline = "top";
    ctx.fillText(String(text).toUpperCase(), x, y);
  }

  function drawPopulations(ctx, land, run, frameIndex, width, height, options) {
    if (!run || !run.males) {
      return;
    }

    const project = projector(land, width, height);
    const index = Math.max(0, Math.min(frameIndex, run.iterations - 1));
    const showTrail = options.trail !== false;

    if (showTrail) {
      drawTrail(ctx, run, index, project);
    }

    drawPopulation(
      ctx,
      run.females,
      run.npopf,
      index,
      project,
      readVar("--female", "#35d6c4"),
      "circle",
    );
    drawPopulation(
      ctx,
      run.males,
      run.npop,
      index,
      project,
      readVar("--male", "#ffb347"),
      "square",
    );

    drawBest(ctx, run, index, project);
  }

  function drawPopulation(ctx, flat, count, frameIndex, project, color, shape) {
    if (!flat || !count) {
      return;
    }

    const base = frameIndex * count * 2;

    ctx.save();
    ctx.fillStyle = alpha(color, 0.85);
    ctx.strokeStyle = "rgba(10, 13, 26, 0.65)";
    ctx.lineWidth = 1;
    ctx.beginPath();

    for (let i = 0; i < count; i += 1) {
      const x = project.x(flat[base + i * 2]);
      const y = project.y(flat[base + i * 2 + 1]);

      if (!isFinite(x) || !isFinite(y)) {
        continue;
      }

      // The female circle is drawn a little wider than the male square on
      // purpose. A converged swarm puts both populations on the same point,
      // and these radii leave the circle's rim visible around the square
      // instead of one population silently erasing the other.
      if (shape === "circle") {
        ctx.moveTo(x + 4.4, y);
        ctx.arc(x, y, 4.4, 0, Math.PI * 2);
      } else {
        ctx.rect(x - 2.8, y - 2.8, 5.6, 5.6);
      }
    }

    ctx.fill();
    ctx.stroke();
    ctx.restore();
  }

  function drawTrail(ctx, run, frameIndex, project) {
    const trail = run.bestTrail;

    if (!trail || frameIndex < 1) {
      return;
    }

    ctx.save();
    ctx.strokeStyle = alpha(readVar("--best", "#ff5d8f"), 0.55);
    ctx.lineWidth = 1.4;
    ctx.beginPath();

    for (let i = 0; i <= frameIndex; i += 1) {
      const x = project.x(trail[i * 2]);
      const y = project.y(trail[i * 2 + 1]);

      if (i === 0) {
        ctx.moveTo(x, y);
      } else {
        ctx.lineTo(x, y);
      }
    }

    ctx.stroke();
    ctx.restore();
  }

  function drawBest(ctx, run, frameIndex, project) {
    const trail = run.bestTrail;

    if (!trail) {
      return;
    }

    const x = project.x(trail[frameIndex * 2]);
    const y = project.y(trail[frameIndex * 2 + 1]);
    const color = readVar("--best", "#ff5d8f");

    // Outlined, never filled. A converged swarm stacks every mayfly on one
    // point, and a solid marker would sit on top of the whole population at
    // exactly the moment the picture is trying to show that they all arrived.
    ctx.save();
    ctx.strokeStyle = color;
    ctx.lineWidth = 1.6;
    ctx.beginPath();
    diamond(ctx, x, y, 7);
    ctx.stroke();

    ctx.strokeStyle = alpha(color, 0.45);
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.arc(x, y, 13, 0, Math.PI * 2);
    ctx.stroke();
    ctx.restore();
  }

  // --- line charts -------------------------------------------------------

  const PAD = { top: 12, right: 12, bottom: 22, left: 52 };

  // Costs on these benchmarks fall through many orders of magnitude, so a
  // linear axis shows one drop and then a flat line along zero. A log axis
  // keeps the late refinement — which is where the variants actually differ —
  // visible.
  //
  // It only works for strictly positive costs, though, and Michalewicz's are
  // negative throughout. Asking for a log axis there collapsed both bounds
  // onto the floor and drew a flat line instead of the run's progress, so a
  // series with any non-positive value falls back to a linear axis and says so
  // in its label.
  function logScale(value, floor) {
    return Math.log10(Math.max(value, floor));
  }

  function allPositive(series) {
    for (const s of series) {
      for (let i = 0; i < s.values.length; i += 1) {
        const value = s.values[i];

        if (isFinite(value) && value <= 0) {
          return false;
        }
      }
    }

    return true;
  }

  function drawSeries(canvas, series, options) {
    const { ctx, width, height } = fitCanvas(canvas);
    const opts = options || {};

    clear(ctx, width, height);

    const live = series.filter((s) => s && s.values && s.values.length);

    if (!live.length) {
      label(ctx, "no data", width / 2, height / 2, "center");

      return;
    }

    const useLog = Boolean(opts.log) && allPositive(live);
    let floor = Infinity;
    let low = Infinity;
    let high = -Infinity;
    let longest = 0;

    for (const s of live) {
      longest = Math.max(longest, s.values.length);

      for (let i = 0; i < s.values.length; i += 1) {
        const value = s.values[i];

        if (!isFinite(value)) {
          continue;
        }

        low = Math.min(low, value);
        high = Math.max(high, value);

        if (value > 0) {
          floor = Math.min(floor, value);
        }
      }
    }

    if (!isFinite(low) || !isFinite(high)) {
      label(ctx, "no finite data", width / 2, height / 2, "center");

      return;
    }

    // A run that reaches the optimum exactly gives log10(0) = -Infinity. Clamp
    // to a decade below the smallest positive value so the curve still lands
    // on the axis instead of vanishing.
    if (!isFinite(floor)) {
      floor = 1e-12;
    }

    floor = Math.min(floor, 1) * 0.1;

    const toValue = useLog ? (v) => logScale(v, floor) : (v) => v;
    let min = useLog ? toValue(Math.max(low, floor)) : low;
    let max = useLog ? toValue(Math.max(high, floor)) : high;

    if (max - min < 1e-9) {
      max = min + 1;
    }

    const plotW = width - PAD.left - PAD.right;
    const plotH = height - PAD.top - PAD.bottom;
    const toX = (i, count) =>
      PAD.left + (count <= 1 ? plotW / 2 : (i / (count - 1)) * plotW);
    const toY = (v) =>
      PAD.top + plotH - ((toValue(v) - min) / (max - min)) * plotH;

    // grid + y labels
    ctx.save();
    ctx.strokeStyle = alpha(readVar("--rule", "#232a4a"), 0.7);
    ctx.lineWidth = 1;

    for (let i = 0; i <= 4; i += 1) {
      const y = PAD.top + (plotH * i) / 4;
      ctx.beginPath();
      ctx.moveTo(PAD.left, y + 0.5);
      ctx.lineTo(width - PAD.right, y + 0.5);
      ctx.stroke();

      const raw = max - ((max - min) * i) / 4;
      label(
        ctx,
        compact(useLog ? Math.pow(10, raw) : raw),
        PAD.left - 6,
        y,
        "right",
      );
    }

    ctx.restore();

    for (const s of live) {
      ctx.save();
      ctx.strokeStyle = s.color;
      ctx.lineWidth = s.width || 1.6;

      if (s.dash) {
        ctx.setLineDash(s.dash);
      }

      ctx.beginPath();

      for (let i = 0; i < s.values.length; i += 1) {
        const value = s.values[i];

        if (!isFinite(value)) {
          continue;
        }

        const x = toX(i, s.values.length);
        const y = toY(value);

        if (i === 0) {
          ctx.moveTo(x, y);
        } else {
          ctx.lineTo(x, y);
        }
      }

      ctx.stroke();
      ctx.restore();
    }

    // playhead: the iteration the swarm view is showing, so the two panels
    // read as one instrument
    if (typeof opts.marker === "number" && longest > 1) {
      const x = toX(Math.min(opts.marker, longest - 1), longest);

      ctx.save();
      ctx.strokeStyle = alpha(readVar("--lume", "#e9ecff"), 0.35);
      ctx.setLineDash([3, 3]);
      ctx.beginPath();
      ctx.moveTo(x, PAD.top);
      ctx.lineTo(x, PAD.top + plotH);
      ctx.stroke();
      ctx.restore();
    }

    frame(ctx, width, height, PAD);

    label(ctx, "1", PAD.left, height - PAD.bottom / 2, "left");
    label(ctx, String(longest), width - PAD.right, height - PAD.bottom / 2, "right");
    label(
      ctx,
      (opts.xLabel || "iteration") +
        (opts.log && !useLog ? " · linear axis (signed costs)" : ""),
      PAD.left + plotW / 2,
      height - PAD.bottom / 2,
      "center",
    );
  }

  // --- cost spread -------------------------------------------------------

  // One column per population member, sorted best to worst, drawn as a bar
  // whose height is its cost. Watching it flatten is watching selection
  // pressure do its work.
  function drawSpread(canvas, options) {
    const { ctx, width, height } = fitCanvas(canvas);
    const { run, frameIndex } = options;

    clear(ctx, width, height);

    if (!run || !run.maleCost) {
      label(ctx, "no data", width / 2, height / 2, "center");

      return;
    }

    const index = Math.max(0, Math.min(frameIndex, run.iterations - 1));
    const males = slice(run.maleCost, index, run.npop);
    const females = slice(run.femaleCost, index, run.npopf);
    const all = males.concat(females).filter(isFinite);

    if (!all.length) {
      return;
    }

    const lowest = Math.min.apply(null, all);
    const ceiling = Math.max.apply(null, all);
    const plotW = width - PAD.left - PAD.right;
    const plotH = height - PAD.top - PAD.bottom;

    // Michalewicz's costs are negative, and a log domain cannot hold them: the
    // floor and the ceiling both collapsed onto 1e-12 and every bar was drawn
    // flat on the baseline, leaving this panel blank for a benchmark the UI
    // offers. Signed costs get a linear axis instead.
    const useLog = lowest > 0;
    const floor = useLog ? Math.max(lowest, 1e-12) : lowest;
    const logFloor = useLog ? Math.log10(floor * 0.5) : 0;
    const logCeiling = useLog ? Math.log10(Math.max(ceiling, floor * 10)) : 0;

    // A linear axis needs headroom so an all-equal population is not a row of
    // zero-height bars.
    const linearSpan = ceiling - lowest || Math.abs(ceiling) || 1;
    const linearFloor = lowest - linearSpan * 0.05;
    const linearCeiling = ceiling + linearSpan * 0.05;

    const toY = (value) => {
      const t = useLog
        ? (Math.log10(Math.max(value, floor * 0.5)) - logFloor) /
          (logCeiling - logFloor || 1)
        : (value - linearFloor) / (linearCeiling - linearFloor || 1);

      return PAD.top + plotH - Math.max(0, Math.min(1, t)) * plotH;
    };

    const total = males.length + females.length;
    const barW = Math.max(1, plotW / total - 1);

    males.sort((a, b) => a - b);
    females.sort((a, b) => a - b);

    let column = 0;

    const paint = (values, color) => {
      ctx.fillStyle = alpha(color, 0.75);

      for (const value of values) {
        if (!isFinite(value)) {
          column += 1;
          continue;
        }

        const x = PAD.left + (column / total) * plotW;
        const y = toY(value);
        ctx.fillRect(x, y, barW, PAD.top + plotH - y);
        column += 1;
      }
    };

    paint(males, readVar("--male", "#ffb347"));
    paint(females, readVar("--female", "#35d6c4"));

    for (let i = 0; i <= 3; i += 1) {
      const y = PAD.top + (plotH * i) / 3;
      const raw = useLog
        ? Math.pow(10, logCeiling - ((logCeiling - logFloor) * i) / 3)
        : linearCeiling - ((linearCeiling - linearFloor) * i) / 3;
      label(ctx, axisLabel(raw, linearCeiling - linearFloor, useLog), PAD.left - 6, y, "right");
    }

    frame(ctx, width, height, PAD);
    label(
      ctx,
      "population sorted by cost — males then females" +
        (useLog ? "" : " · linear axis"),
      PAD.left + plotW / 2,
      height - PAD.bottom / 2,
      "center",
    );
  }

  // A converged population spans a range far smaller than its own magnitude —
  // four significant digits then print the same string for every tick, which
  // reads as a broken axis rather than a tight one. Scale the precision to the
  // span being labelled.
  function axisLabel(value, span, useLog) {
    if (useLog || !isFinite(span) || span <= 0) {
      return compact(value);
    }

    const magnitude = Math.abs(value);

    if (magnitude === 0) {
      return "0";
    }

    const digits = Math.ceil(Math.log10(magnitude / span)) + 2;

    if (digits > 6) {
      return value.toExponential(2);
    }

    return value.toPrecision(Math.max(2, Math.min(15, digits)));
  }

  function slice(flat, frameIndex, count) {
    const out = [];
    const base = frameIndex * count;

    for (let i = 0; i < count; i += 1) {
      out.push(flat[base + i]);
    }

    return out;
  }

  // --- boot ring ---------------------------------------------------------

  function ring(element, progress) {
    if (element) {
      element.style.setProperty("--boot-progress", String(progress));
    }
  }

  // --- DPR watch ---------------------------------------------------------

  function watchDPR(onChange) {
    if (!window.matchMedia) {
      return;
    }

    let query = null;

    const listen = () => {
      query = window.matchMedia(`(resolution: ${currentDPR()}dppx)`);
      query.addEventListener("change", () => {
        listen();
        onChange();
      }, { once: true });
    };

    listen();
  }

  window.Render = {
    readVar,
    invalidateTheme,
    parseColor,
    alpha,
    fitCanvas,
    clear,
    compact,
    drawSwarm,
    drawSeries,
    drawSpread,
    ring,
    watchDPR,
  };
})();
