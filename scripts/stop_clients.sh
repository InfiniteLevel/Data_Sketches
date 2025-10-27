#!/usr/bin/env bash
set -euo pipefail

pkill -TERM -x sketcher || true
sleep 1
pkill -KILL -x sketcher || true

echo "[INFO] remaining:"
ps -eo pid,user,comm,args | grep -E '[s]ketcher|[g]rpc|[n]ohup' || true