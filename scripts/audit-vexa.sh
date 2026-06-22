#!/usr/bin/env bash
# vexa-audit.sh — Automated workflow audit for vexa project
# Run from /home/ubuntu/projects/vexa

set -euo pipefail

ROOT_DIR="/home/ubuntu/projects/vexa"
APP_DIR="${ROOT_DIR}/02-application"
FAIL=0

log() { echo "[vexa-audit] $*"; }
warn() { echo "[vexa-audit] WARNING: $*"; FAIL=1; }

log "Starting vexa workflow audit..."

# 1. Check 02-application is subtree, not separate repo
if [ -d "${APP_DIR}/.git" ]; then
  warn "02-application/.git exists — should be subtree, not separate repo"
else
  log "OK: 02-application/ is subtree (no .git)"
fi

# 2. Check git remotes are valid
cd "${ROOT_DIR}"
REMOTES=$(git remote -v)
if echo "$REMOTES" | grep -q "02-application-remote"; then
  warn "Invalid remote '02-application-remote' still present"
else
  log "OK: no invalid 02-application-remote"
fi

if echo "$REMOTES" | grep -q "origin.*soumabali/vexa-root"; then
  log "OK: origin points to vexa-root"
else
  warn "origin remote does not point to soumabali/vexa-root"
fi

if echo "$REMOTES" | grep -q "app-vexa.*soumabali/vexa"; then
  log "OK: app-vexa points to vexa application repo"
else
  warn "app-vexa remote does not point to soumabali/vexa"
fi

# 3. Check root context files exist
for f in README.md CLAUDE.md 00-meta/urls.md 00-meta/skills.md 00-meta/git-structure.md 00-meta/pre-push-checklist.md 00-meta/audit-checklist.md 00-meta/credentials-index.md; do
  if [ -f "${ROOT_DIR}/${f}" ]; then
    log "OK: ${f} exists"
  else
    warn "Missing root context file: ${f}"
  fi
done

# 4. Check Obsidian SSOT files exist
OBSIDIAN_DIR="/home/ubuntu/Documents/Obsidian Vault/infra"
for f in "vexa — Project Index.md" "vexa Progress.md" "vexa Rules.md" "vexa Workflow Runbook.md" "vexa Claude Code Setup.md" "vexa Runbook.md"; do
  if [ -f "${OBSIDIAN_DIR}/${f}" ]; then
    log "OK: Obsidian ${f} exists"
  else
    warn "Missing Obsidian SSOT: ${f}"
  fi
done

# 5. Check application context files exist
for f in CLAUDE.md README.md .claude/settings.json .claude/rules/hermes-skills.md; do
  if [ -f "${APP_DIR}/${f}" ]; then
    log "OK: 02-application/${f} exists"
  else
    warn "Missing app context file: 02-application/${f}"
  fi
done

# 6. Check docker-compose.prod.yml mounts /backups
if grep -q "/backups:/backups" "${APP_DIR}/docker-compose.prod.yml"; then
  log "OK: docker-compose.prod.yml mounts /backups"
else
  warn "docker-compose.prod.yml does not mount /backups"
fi

# 7. Check backup directories exist
for d in /backups/vexa/db /backups/vexa/wireguard /backups/vexa/logs; do
  if [ -d "$d" ]; then
    log "OK: ${d} exists"
  else
    warn "Missing backup directory: ${d}"
  fi
done

# 8. Check Claude skills exist
for skill in superpowers caveman graphify; do
  if [ -d "${HOME}/.claude/skills/${skill}" ]; then
    log "OK: skill ${skill} installed"
  else
    warn "Missing Claude skill: ${skill}"
  fi
done

# 9. Check playwright MCP in global settings
if grep -q "playwright" "${HOME}/.claude/settings.json"; then
  log "OK: playwright MCP configured"
else
  warn "playwright MCP not found in ~/.claude/settings.json"
fi

# 10. Check no K8s/Terraform refs in active docs
if grep -R "kubernetes\|terraform\|helm\|k8s" "${APP_DIR}/docs" 2>/dev/null | grep -v "_archive" | grep -v "\.md:" | head -1; then
  warn "Found K8s/Terraform/Helm refs in active docs"
else
  log "OK: no K8s/Terraform/Helm refs in active docs"
fi

# 11. Check AGENTS.md exists
if [ -f "${ROOT_DIR}/AGENTS.md" ]; then
  log "OK: AGENTS.md exists"
else
  warn "Missing AGENTS.md — all agents must read this file"
fi

# 12. Check agent compliance phrases in context files
COMPLIANCE_FILES=(
  "${ROOT_DIR}/AGENTS.md"
  "${ROOT_DIR}/CLAUDE.md"
  "${ROOT_DIR}/00-meta/skills.md"
  "${APP_DIR}/CLAUDE.md"
  "${APP_DIR}/.claude/rules/hermes-skills.md"
)
for f in "${COMPLIANCE_FILES[@]}"; do
  if [ -f "$f" ]; then
    if grep -qi "AGENT COMPLIANCE\|wajib\|mandatory\|must read" "$f"; then
      log "OK: ${f} contains agent compliance clause"
    else
      warn "${f} missing agent compliance clause"
    fi
  fi
done

# 13. Check schema and compliance script exist
for f in "${ROOT_DIR}/00-meta/claude-response-schema.json" "${ROOT_DIR}/scripts/claude-compliance-check.sh"; do
  if [ -f "$f" ]; then
    log "OK: ${f} exists"
  else
    warn "Missing compliance tool: ${f}"
  fi
done

# Summary
if [ $FAIL -eq 0 ]; then
  log "AUDIT PASSED ✅"
  exit 0
else
  log "AUDIT FAILED ❌ — see warnings above"
  exit 1
fi
