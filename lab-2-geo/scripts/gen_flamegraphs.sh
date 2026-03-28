#!/usr/bin/env bash
set -euo pipefail

PROF_DIR="${1:-metrics/profiles}"
PLOT_DIR="${2:-metrics/plots}"
PORT=18083

mkdir -p "$PLOT_DIR"

for prof in cpu_findnearby cpu_insert mem_findnearby mem_insert; do
  echo "→ flamegraph: ${prof}"

  # kill anything still on port
  fuser -k ${PORT}/tcp 2>/dev/null || true

  # start pprof HTTP server
  go tool pprof -http=":${PORT}" "${PROF_DIR}/${prof}.prof" &
  PPROF_PID=$!
  sleep 2

  # download flamegraph HTML
  curl -sf "http://localhost:${PORT}/ui/flamegraph" \
    -o "${PLOT_DIR}/flamegraph_${prof}.html" || { kill $PPROF_PID 2>/dev/null; continue; }

  kill $PPROF_PID 2>/dev/null || true
  wait $PPROF_PID 2>/dev/null || true
  sleep 1

  # render to PNG
  chromium-browser --headless --no-sandbox --disable-gpu \
    --screenshot="${PLOT_DIR}/flamegraph_${prof}.png" \
    --window-size=1600,900 \
    "file://$(pwd)/${PLOT_DIR}/flamegraph_${prof}.html" 2>/dev/null || true
  sleep 2

  echo "  saved: ${PLOT_DIR}/flamegraph_${prof}.png"
done
