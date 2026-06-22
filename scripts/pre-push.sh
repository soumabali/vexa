#!/usr/bin/env bash
# pre-push hook for vexa root orchestrator
# Installation: ln -s /home/ubuntu/projects/vexa/scripts/pre-push.sh .git/hooks/pre-push

set -euo pipefail

ROOT_DIR="/home/ubuntu/projects/vexa"

echo "[vexa pre-push] Running audit..."
cd "${ROOT_DIR}"
if ! /home/ubuntu/projects/vexa/scripts/audit-vexa.sh; then
  echo "[vexa pre-push] ❌ Audit failed. Push aborted."
  exit 1
fi

echo "[vexa pre-push] ✅ Audit passed. Continuing push..."
exit 0
