#!/usr/bin/env bash
# vexa-status.sh — Dashboard snapshot untuk project vexa
# Run dari /home/ubuntu/projects/vexa atau langsung dengan path penuh

set -euo pipefail

ROOT_DIR="/home/ubuntu/projects/vexa"
APP_DIR="${ROOT_DIR}/02-application"

echo ""
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║                    vexa Status Dashboard                       ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

# Git
echo "── Git ────────────────────────────────────────────────────────"
cd "${ROOT_DIR}"
echo "Root branch:      $(git rev-parse --abbrev-ref HEAD)"
echo "Root commit:      $(git rev-parse --short HEAD)"
echo "Root remote:      $(git remote get-url origin 2>/dev/null || echo 'N/A')"
echo "App repo commit:  $(git ls-remote https://github.com/soumabali/vexa.git main 2>/dev/null | awk '{print $1}' | cut -c1-7 || echo 'N/A')"
echo "Subtree OK:       $(test ! -d "${APP_DIR}/.git" && echo '✅' || echo '❌')"
echo ""

# Audit
echo "── Audit ───────────────────────────────────────────────────────"
if /home/ubuntu/projects/vexa/scripts/audit-vexa.sh > /tmp/vexa-audit.log 2>&1; then
  echo "Audit:            ✅ PASSED"
else
  echo "Audit:            ❌ FAILED"
  tail -5 /tmp/vexa-audit.log
fi
echo ""

# Production Health (opsional, bisa fail kalau network down)
echo "── Production Health (best effort) ──────────────────────────────"
WEB_STATUS=$(curl -s -o /dev/null -w "%{http_code}" https://vexa.nexigo.my.id || echo "000")
API_STATUS=$(curl -s -o /dev/null -w "%{http_code}" https://api-vexa.nexigo.my.id/health || echo "000")
echo "Web (vexa.nexigo.my.id):      ${WEB_STATUS}"
echo "API (api-vexa.nexigo.my.id):    ${API_STATUS}"
echo ""

# Containers
echo "── Containers ─────────────────────────────────────────────────"
if command -v docker &> /dev/null; then
  docker ps --filter "name=vexa-" --format "{{.Names}}: {{.Status}}" 2>/dev/null || echo "No vexa containers running"
else
  echo "Docker not available"
fi
echo ""

# Backup dirs
echo "── Backup Directories ───────────────────────────────────────────"
for d in /backups/vexa/db /backups/vexa/wireguard /backups/vexa/logs; do
  if [ -d "$d" ]; then
    echo "${d}: ✅"
  else
    echo "${d}: ❌"
  fi
done
echo ""

# Skills
echo "── Claude Skills/MCP ────────────────────────────────────────────"
for skill in superpowers caveman graphify; do
  if [ -d "${HOME}/.claude/skills/${skill}" ]; then
    echo "${skill}: ✅"
  else
    echo "${skill}: ❌"
  fi
done
if grep -q "playwright" "${HOME}/.claude/settings.json" 2>/dev/null; then
  echo "playwright MCP:   ✅"
else
  echo "playwright MCP:   ❌"
fi
echo ""

echo "══════════════════════════════════════════════════════════════════"
echo "Run /home/ubuntu/projects/vexa/scripts/vexa-status.sh anytime."
echo "══════════════════════════════════════════════════════════════════"
