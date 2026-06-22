#!/usr/bin/env bash
# =============================================================================
# vexa — Self-Hosted Rollback Script
# Usage: ./scripts/rollback.sh [git-revision|image-tag|previous]
# =============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

COMPOSE_FILE="${COMPOSE_FILE:-${REPO_ROOT}/docker-compose.prod.yml}"
TARGET="${1:-previous}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

log() { echo -e "${BLUE}[$(date +%Y-%m-%d\ %H:%M:%S)]${NC} $*"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $*"; }
error() { echo -e "${RED}[ERROR]${NC} $*"; }
success() { echo -e "${GREEN}[OK]${NC} $*"; }
info() { echo -e "${CYAN}[INFO]${NC} $*"; }

log "=== vexa Self-Hosted Rollback ==="
info "Compose file: ${COMPOSE_FILE}"
info "Target: ${TARGET}"

if [ ! -f "${COMPOSE_FILE}" ]; then
  error "Compose file not found: ${COMPOSE_FILE}"
  exit 1
fi

if ! command -v docker >/dev/null 2>&1; then
  error "docker not found"
  exit 1
fi

# Interactive confirmation
if [ "${AUTO_CONFIRM:-}" != "1" ]; then
  echo -n "⚠️  This will roll back the running vexa stack. Are you sure? [yes/no]: "
  read -r CONFIRM
  if [ "$CONFIRM" != "yes" ]; then
    warn "Rollback aborted by user"
    exit 1
  fi
fi

# -----------------------------------------------------------------------------
# Rollback strategy
# -----------------------------------------------------------------------------
if [ "$TARGET" = "previous" ]; then
  # Roll back to the previous Git commit on the current branch
  if ! command -v git >/dev/null 2>&1; then
    error "git not found"
    exit 1
  fi
  cd "${REPO_ROOT}"
  PREV_COMMIT="$(git rev-parse HEAD~1)"
  log "Rolling back code to previous commit: ${PREV_COMMIT}"
  git checkout "${PREV_COMMIT}" -- .
  docker compose -f "${COMPOSE_FILE}" up -d --build --no-deps api web
elif echo "$TARGET" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+'; then
  # Roll back to a specific released image tag
  log "Rolling back to image tag: ${TARGET}"
  API_IMAGE="${API_IMAGE:-ghcr.io/soumabali/vexa/api:${TARGET}}"
  WEB_IMAGE="${WEB_IMAGE:-ghcr.io/soumabali/vexa/web:${TARGET}}"
  cat > /tmp/vexa-rollback.yml <<EOF
services:
  api:
    image: ${API_IMAGE}
    build: !reset null
  web:
    image: ${WEB_IMAGE}
    build: !reset null
EOF
  docker compose -f "${COMPOSE_FILE}" -f /tmp/vexa-rollback.yml up -d --no-deps --build=false api web
else
  # Assume TARGET is a Git revision
  if ! command -v git >/dev/null 2>&1; then
    error "git not found"
    exit 1
  fi
  cd "${REPO_ROOT}"
  log "Rolling back code to revision: ${TARGET}"
  git checkout "${TARGET}" -- .
  docker compose -f "${COMPOSE_FILE}" up -d --build --no-deps api web
fi

# -----------------------------------------------------------------------------
# Health checks
# -----------------------------------------------------------------------------
log "Running post-rollback health checks..."
sleep 5

for attempt in 1 2 3; do
  if docker compose -f "${COMPOSE_FILE}" exec -T api wget -qO- http://127.0.0.1:8080/health/live >/dev/null 2>&1; then
    success "API health check passed"
    break
  fi
  if [ "$attempt" -eq 3 ]; then
    warn "API health check failed — rollback may be incomplete"
  fi
  sleep 5
done

log "=== Rollback Summary ==="
success "Rollback complete!"
docker compose -f "${COMPOSE_FILE}" ps api web

log "If issues persist, investigate logs:"
echo "  docker compose -f ${COMPOSE_FILE} logs --tail=100 api"
echo "  docker compose -f ${COMPOSE_FILE} logs --tail=100 web"
