#!/usr/bin/env bash
# audit-vexa.sh — Application-level workflow audit for vexa
# Run from the root of the application repo (02-application/ in local layout).

set -euo pipefail

APP_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FAIL=0

log() { echo "[vexa-app-audit] $*"; }
warn() { echo "[vexa-app-audit] WARNING: $*"; FAIL=1; }

cd "${APP_DIR}"
log "Starting application-level vexa audit..."

# 1. Check application context files exist
for f in CLAUDE.md README.md .claude/settings.json .claude/rules/hermes-skills.md; do
  if [ -f "${APP_DIR}/${f}" ]; then
    log "OK: ${f} exists"
  else
    warn "Missing app context file: ${f}"
  fi
done

# 2. Check docker-compose.prod.yml mounts /backups
if grep -q "/backups:/backups" "${APP_DIR}/docker-compose.prod.yml"; then
  log "OK: docker-compose.prod.yml mounts /backups"
else
  warn "docker-compose.prod.yml does not mount /backups"
fi

# 3. Check backup artifacts exist
for f in scripts/backup.sh scripts/backup.crontab Dockerfile.backup; do
  if [ -f "${APP_DIR}/${f}" ]; then
    log "OK: ${f} exists"
  else
    warn "Missing backup artifact: ${f}"
  fi
done

# 4. Check agent compliance phrases in context files
COMPLIANCE_FILES=(
  "${APP_DIR}/CLAUDE.md"
  "${APP_DIR}/.claude/rules/hermes-skills.md"
)
for f in "${COMPLIANCE_FILES[@]}"; do
  if [ -f "$f" ]; then
    if grep -qi "AGENT COMPLIANCE\|wajib\|mandatory\|must read" "$f"; then
      log "OK: $(basename "$f") contains agent compliance clause"
    else
      warn "$(basename "$f") missing agent compliance clause"
    fi
  fi
done

# 5. Check no K8s/Terraform refs in active docs
if grep -R "kubernetes\|terraform\|helm\|k8s" "${APP_DIR}/docs" 2>/dev/null | grep -v "_archive" | grep -v "\.md:" | head -1; then
  warn "Found K8s/Terraform/Helm refs in active docs"
else
  log "OK: no K8s/Terraform/Helm refs in active docs"
fi

# 6. Check required skills are referenced in context
for skill in superpowers caveman graphify; do
  if grep -Rq "${skill}" "${APP_DIR}/.claude/" "${APP_DIR}/CLAUDE.md" 2>/dev/null; then
    log "OK: skill ${skill} referenced in app context"
  else
    warn "skill ${skill} not referenced in app context"
  fi
done

# Summary
if [ $FAIL -eq 0 ]; then
  log "APPLICATION AUDIT PASSED ✅"
  exit 0
else
  log "APPLICATION AUDIT FAILED ❌ — see warnings above"
  exit 1
fi
