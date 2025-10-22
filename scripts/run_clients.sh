#!/usr/bin/env bash
set -euo pipefail

ENV_FILE="${1:-}"
if [[ -z "${ENV_FILE}" || ! -f "${ENV_FILE}" ]]; then
  echo "Usage: $0 <env_file>"
  exit 1
fi
# shellcheck disable=SC1090
source "$ENV_FILE"

BIN="${BIN_PATH:-$HOME/DS/app/sketcher}"
DATA="${DATA_PATH:-$HOME/DS/Data_Sketches/database/test.csv}"
LOG_DIR="${LOG_DIR:-$HOME/DS/logs}"
mkdir -p "$LOG_DIR"

echo "[INFO] launching $NUM_CLIENTS clients -> server ${SERVER_IP}:${SERVER_PORT} | sketch=${SKETCH} | field=${FIELD_NAME} | type=${FIELD_TYPE} | merge=${MERGE_EVERY} | stream=${STREAM_RATE}"

for ((i=1; i<=NUM_CLIENTS; i++)); do
  LOG="$LOG_DIR/client-${i}.log"
  sleep $(( (i-1) * START_STAGGER_SECS )) &
  (
    wait %1 2>/dev/null || true
    nohup "$BIN" \
      -client -a "$SERVER_IP" -port "$SERVER_PORT" \
      -sketch "$SKETCH" \
      -d "$DATA" \
      -name "$FIELD_NAME" \
      -type "$FIELD_TYPE" \
      -merge "$MERGE_EVERY" \
      -stream "$STREAM_RATE" \
      > "$LOG" 2>&1 &
    echo "[OK] client $i started; log: $LOG"
  ) &
done
wait