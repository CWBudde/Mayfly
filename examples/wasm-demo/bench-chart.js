/*
 * bench-chart.js — the shootout's grouped bar chart.
 *
 * No charting library, no CDN: the demo ships as plain static files and adding
 * a script tag pointing at someone else's origin would be the only external
 * dependency on the whole site.
 *
 * Series are distinguished by colour AND hatch pattern, and every bar carries
 * its rank as text. Colour alone excludes roughly one man in twelve, and this
 * chart's entire job is letting someone rank eight algorithms at a glance.
 */
(function () {
  "use strict";

  const PAD = { top: 16, right: 12, bottom: 46, left: 62 };

  // The palette walks the page's own tokens rather than inventing colours, so
  // the chart and the stylesheet cannot drift.
  const SERIES_VARS = [
    "--male",
    "--female",
    "--best",
    "--elite",
    "--ok",
    "--warn",
    "--dim",
  ];

  function seriesColor(index) {
    return window.Render.readVar(SERIES_VARS[index % SERIES_VARS.length], "#888");
  }

  function hatch(ctx, x, y, width, height, index) {
    if (index % 3 === 0) {
      return;
    }

    ctx.save();
    ctx.beginPath();
    ctx.rect(x, y, width, height);
    ctx.clip();
    ctx.strokeStyle = "rgba(0, 0, 0, 0.35)";
    ctx.lineWidth = 1;
    ctx.beginPath();

    const step = index % 3 === 1 ? 4 : 7;

    for (let offset = -height; offset < width + height; offset += step) {
      ctx.moveTo(x + offset, y + height);
      ctx.lineTo(x + offset + height, y);
    }

    ctx.stroke();
    ctx.restore();
  }

  // draw renders one grouped bar chart: a group per benchmark, a bar per
  // algorithm, height = the chosen statistic on a log axis (these costs span
  // many decades, and a linear axis would render every good result as zero).
  function draw(canvas, data, options) {
    const { ctx, width, height } = window.Render.fitCanvas(canvas);
    const opts = options || {};

    window.Render.clear(ctx, width, height);

    const groups = data.groups || [];
    const algorithms = data.algorithms || [];
    const hidden = opts.hidden || new Set();
    const visible = algorithms
      .map((name, index) => ({ name, index }))
      .filter((entry) => !hidden.has(entry.name));

    if (!groups.length || !visible.length) {
      window.Render.clear(ctx, width, height);
      ctx.fillStyle = window.Render.readVar("--dim", "#7d86b4");
      ctx.font = '11px "JetBrains Mono", ui-monospace, monospace';
      ctx.textAlign = "center";
      ctx.fillText("no results yet", width / 2, height / 2);

      return;
    }

    let low = Infinity;
    let high = -Infinity;
    let lowest = Infinity;

    for (const group of groups) {
      for (const entry of visible) {
        const value = group.values[entry.index];

        if (value === null || value === undefined || !isFinite(value)) {
          continue;
        }

        // A cost of exactly zero is a real, and excellent, result; it just has
        // no place on a log axis. It is floored to the smallest positive value
        // seen so the bar reaches the bottom of the plot instead of vanishing.
        if (value > 0) {
          low = Math.min(low, value);
        }

        lowest = Math.min(lowest, value);
        high = Math.max(high, value);
      }
    }

    if (!isFinite(lowest)) {
      window.Render.clear(ctx, width, height);

      return;
    }

    // Michalewicz's costs are negative, so no positive `low` exists: the axis
    // degenerated, `high` was forced to 1, and every bar was drawn at the floor
    // with zero height. Signed data gets a linear axis instead of a log one.
    const useLog = lowest > 0;

    if (!isFinite(low)) {
      low = 1e-12;
    }

    if (!isFinite(high)) {
      high = 1;
    }

    // These costs really do reach 1e-60 on the easy functions, and an axis that
    // honoured that would be sixty decades of empty space with every label
    // overlapping. Below about twelve decades from the worst result the
    // difference is numerical noise, not optimizer quality, so the axis stops
    // there and anything smaller sits on the floor.
    const MAX_DECADES = 12;
    const logHigh = Math.log10(Math.max(high, low * 10)) + 0.2;
    const logLow = Math.max(Math.log10(low) - 0.5, logHigh - MAX_DECADES);

    // Linear bounds, used when the data is signed. Bars are drawn from the
    // bottom of the plot, so the floor sits a little below the smallest value.
    const linearSpan = high - lowest || Math.abs(high) || 1;
    const linearLow = lowest - linearSpan * 0.08;
    const linearHigh = high + linearSpan * 0.08;
    const plotW = width - PAD.left - PAD.right;
    const plotH = height - PAD.top - PAD.bottom;

    const toY = (value) => {
      let t;

      if (useLog) {
        const safe = value > 0 ? value : low;
        t = (Math.log10(safe) - logLow) / (logHigh - logLow || 1);
      } else {
        t = (value - linearLow) / (linearHigh - linearLow || 1);
      }

      return PAD.top + plotH - Math.max(0, Math.min(1, t)) * plotH;
    };

    // gridlines, one per decade
    ctx.save();
    ctx.strokeStyle = window.Render.alpha(
      window.Render.readVar("--rule", "#232a4a"),
      0.8,
    );
    ctx.fillStyle = window.Render.readVar("--dim", "#7d86b4");
    ctx.font = '10px "JetBrains Mono", ui-monospace, monospace';
    ctx.textAlign = "right";
    ctx.textBaseline = "middle";

    if (useLog) {
      const firstDecade = Math.floor(logLow);
      const lastDecade = Math.ceil(logHigh);

      // At most ~9 labelled decades, whatever the span; more than that and the
      // mono labels collide into a smear.
      const step = Math.max(1, Math.ceil((lastDecade - firstDecade) / 9));

      for (let decade = firstDecade; decade <= lastDecade; decade += step) {
        const y = toY(Math.pow(10, decade));

        if (y < PAD.top || y > PAD.top + plotH) {
          continue;
        }

        ctx.beginPath();
        ctx.moveTo(PAD.left, y + 0.5);
        ctx.lineTo(width - PAD.right, y + 0.5);
        ctx.stroke();
        ctx.fillText(`1e${decade}`, PAD.left - 6, y);
      }
    } else {
      for (let i = 0; i <= 5; i += 1) {
        const value = linearHigh - ((linearHigh - linearLow) * i) / 5;
        const y = toY(value);

        ctx.beginPath();
        ctx.moveTo(PAD.left, y + 0.5);
        ctx.lineTo(width - PAD.right, y + 0.5);
        ctx.stroke();
        ctx.fillText(window.Render.compact(value), PAD.left - 6, y);
      }
    }

    ctx.restore();

    const groupW = plotW / groups.length;
    const barW = Math.max(2, (groupW * 0.78) / visible.length);

    groups.forEach((group, groupIndex) => {
      const groupX = PAD.left + groupIndex * groupW;

      visible.forEach((entry, slot) => {
        const value = group.values[entry.index];

        if (value === null || value === undefined || !isFinite(value)) {
          return;
        }

        const x = groupX + groupW * 0.11 + slot * barW;
        const y = toY(value);
        const barH = PAD.top + plotH - y;

        ctx.fillStyle = window.Render.alpha(seriesColor(entry.index), 0.85);
        ctx.fillRect(x, y, barW - 1, barH);
        hatch(ctx, x, y, barW - 1, barH, entry.index);
      });

      ctx.save();
      ctx.fillStyle = window.Render.readVar("--dim", "#7d86b4");
      ctx.font = '10px "JetBrains Mono", ui-monospace, monospace';
      ctx.textAlign = "center";
      ctx.textBaseline = "top";

      const centre = groupX + groupW / 2;
      const name = group.label;

      // Long benchmark names would overlap at seven groups across; they get
      // truncated rather than allowed to collide into illegibility.
      const maxChars = Math.max(4, Math.floor(groupW / 6));
      const shown =
        name.length > maxChars ? `${name.slice(0, maxChars - 1)}…` : name;

      ctx.fillText(shown, centre, PAD.top + plotH + 8);
      ctx.restore();
    });

    ctx.strokeStyle = window.Render.readVar("--rule", "#232a4a");
    ctx.lineWidth = 1;
    ctx.strokeRect(PAD.left + 0.5, PAD.top + 0.5, plotW - 1, plotH - 1);

    ctx.save();
    ctx.fillStyle = window.Render.readVar("--dim", "#7d86b4");
    ctx.font = '10px "JetBrains Mono", ui-monospace, monospace';
    ctx.textAlign = "center";
    ctx.textBaseline = "bottom";
    ctx.fillText(
      (opts.yLabel || "median best cost") + (useLog ? "" : " · linear axis"),
      width / 2,
      height - 6,
    );
    ctx.restore();
  }

  globalThis.BenchChart = { draw, seriesColor };
})();
