#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
DEMO_DIR="$ROOT_DIR/examples/wasm-demo"
OUT_DIR="${1:-$ROOT_DIR/dist}"

mkdir -p "$OUT_DIR"

# Resolve OUT_DIR to an absolute path before anything uses it. The build below
# runs with go's working directory set to the demo module, so a relative -o
# would be written under examples/wasm-demo/ while mkdir and the asset copy
# here operate from the repository root. The Pages workflow passes a relative
# "dist", so that split silently produced an upload with every static asset and
# no mayfly.wasm.
OUT_DIR="$(cd "$OUT_DIR" && pwd)"

# The demo is its own module with a replace back to the library, matching every
# other directory under examples/. -C keeps that working from any cwd.
GOOS=js GOARCH=wasm go build -C "$DEMO_DIR" -o "$OUT_DIR/mayfly.wasm" .

# Copy every static asset by glob rather than by name. Listing files explicitly
# means each new page, script or stylesheet has to be remembered here, and
# forgetting one produces a broken GitHub Pages deploy with no build failure.
# Go sources are matched by none of these patterns, so nothing unwanted ships.
shopt -s nullglob
assets=("$DEMO_DIR"/*.html "$DEMO_DIR"/*.css "$DEMO_DIR"/*.js "$DEMO_DIR"/*.svg)
shopt -u nullglob

if [ ${#assets[@]} -eq 0 ]; then
    echo "error: no static assets found in $DEMO_DIR" >&2
    exit 1
fi

cp "${assets[@]}" "$OUT_DIR/"

# wasm_exec.js is copied from the toolchain, never vendored: it is version-
# locked to the compiler that produced the .wasm, and a stale committed copy
# fails at runtime in ways that look like demo bugs. Go moved it in 1.24, and
# go.mod still says 1.23, so both locations have to be probed.
wasm_exec=""

for candidate in "$(go env GOROOT)/lib/wasm/wasm_exec.js" "$(go env GOROOT)/misc/wasm/wasm_exec.js"; do
    if [ -f "$candidate" ]; then
        wasm_exec="$candidate"
        break
    fi
done

if [ -z "$wasm_exec" ]; then
    echo "error: wasm_exec.js not found under $(go env GOROOT)" >&2
    exit 1
fi

cp "$wasm_exec" "$OUT_DIR/"

printf "Copied %d static asset(s)\n" "${#assets[@]}"
printf "WASM demo built at %s\n" "$OUT_DIR"
