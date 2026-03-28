#!/usr/bin/env bash
set -euo pipefail

PROF_DIR="${1:-metrics/profiles}"
PLOT_DIR="${2:-metrics/plots}"
PORT=18084

mkdir -p "$PLOT_DIR"

for prof in "$@"; do
  shift; shift
  break
done

# accept list of profile base-names from env or scan dir
if [ -n "${PROF_NAMES:-}" ]; then
  IFS=',' read -ra NAMES <<< "$PROF_NAMES"
else
  NAMES=()
  for f in "${PROF_DIR}"/*.prof; do
    [ -f "$f" ] || continue
    base=$(basename "$f" .prof)
    NAMES+=("$base")
  done
fi

for prof in "${NAMES[@]}"; do
  echo "→ flamegraph: ${prof}"

  fuser -k ${PORT}/tcp 2>/dev/null || true
  sleep 1

  go tool pprof -http=":${PORT}" "${PROF_DIR}/${prof}.prof" &
  PPROF_PID=$!
  sleep 2

  curl -sf "http://localhost:${PORT}/ui/flamegraph" \
    -o "${PLOT_DIR}/flamegraph_${prof}.html" || { kill $PPROF_PID 2>/dev/null; continue; }

  kill $PPROF_PID 2>/dev/null || true
  wait $PPROF_PID 2>/dev/null || true
  sleep 1

  chromium-browser --headless --no-sandbox --disable-gpu \
    --screenshot="${PLOT_DIR}/flamegraph_${prof}.png" \
    --window-size=1600,900 \
    "file://$(pwd)/${PLOT_DIR}/flamegraph_${prof}.html" 2>/dev/null || true
  sleep 2

  echo "  saved: ${PLOT_DIR}/flamegraph_${prof}.png"
done
