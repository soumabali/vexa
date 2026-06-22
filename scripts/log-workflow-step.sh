#!/usr/bin/env bash
# log-workflow-step.sh — Append a trace entry to the current workflow run log
# Usage: log-workflow-step.sh <phase> <status> "message"

set -euo pipefail

ROOT_DIR="/home/ubuntu/projects/vexa"
LOG_DIR="${ROOT_DIR}/03-history/sessions/run-logs"
PHASE="${1:-}"
STATUS="${2:-}"
MESSAGE="${3:-}"

if [ -z "$PHASE" ] || [ -z "$STATUS" ]; then
  echo "Usage: $0 <phase> <status> \"message\""
  exit 1
fi

mkdir -p "$LOG_DIR"

# If RUN_LOG env is not set, use today's active log or create new
if [ -z "${RUN_LOG:-}" ]; then
  RUN_LOG="${LOG_DIR}/workflow-$(date +%Y%m%d)-active.log"
fi

TIMESTAMP=$(date +%Y-%m-%dT%H:%M:%S%z)
echo "[$TIMESTAMP] [$PHASE] [$STATUS] $MESSAGE" >> "$RUN_LOG"
echo "[workflow-trace] Logged to $RUN_LOG"
