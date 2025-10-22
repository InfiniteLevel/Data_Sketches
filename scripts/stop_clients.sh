#!/usr/bin/env bash
set -euo pipefail
# 停止所有正在运行的 client 进程
pkill -f "DS/app/sketcher.*-client" || true
sleep 1
pgrep -fa sketcher || true
echo "[OK] all clients stopped (if any)."