#!/usr/bin/env bash
# =============================================================================
# SSH Manager — Rollback Script
# Usage: ./scripts/rollback.sh [TARGET_REVISION|last|previous]
# =============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

K8S_NAMESPACE="${K8S_NAMESPACE:-ssh-manager-prod}"
KUBECONFIG="${KUBECONFIG:-${REPO_ROOT}/infra/k8s/kubeconfig-production}"
ROLLOUT_TIMEOUT="${ROLLOUT_TIMEOUT:-5m}"

# Colors
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

TARGET="${1:-previous}"

# ---------------------------------------------------------------------------
# Pre-flight
# ---------------------------------------------------------------------------
log "=== Rollback ==="
info  "Target: ${TARGET}"
info  "Namespace: ${K8S_NAMESPACE}"

if ! command -v kubectl >/dev/null 2>&1; then
  error "kubectl not found"
  exit 1
fi

if [ -f "$KUBECONFIG" ]; then
  export KUBECONFIG
fi

kubectl config current-context >/dev/null 2>&1 || {
  error "kubectl cannot connect to cluster"
  exit 1
}

# Interactive confirmation
if [ "${AUTO_CONFIRM:-}" != "1" ]; then
  echo -n "⚠️  This will ROLL BACK production. Are you sure? [yes/no]: "
  read -r CONFIRM
  if [ "$CONFIRM" != "yes" ]; then
    warn "Rollback aborted by user"
    exit 1
  fi
fi

# ---------------------------------------------------------------------------
# Determine rollback revision
# ---------------------------------------------------------------------------
log "Checking rollout history..."

# Get current revision
CURRENT_REV=$(kubectl get deployment/api -n "${K8S_NAMESPACE}" -o jsonpath='{.metadata.annotations.deployment\.kubernetes\.io/revision}' 2>/dev/null || echo "unknown")
info "Current API revision: ${CURRENT_REV}"

# Determine target revision
if [ "$TARGET" = "last" ]; then
  # Roll back to the immediately previous revision
  kubectl rollout undo deployment/api -n "${K8S_NAMESPACE}"
  kubectl rollout undo deployment/web -n "${K8S_NAMESPACE}"
  success "Rolled back to previous revision"

elif [ "$TARGET" = "previous" ]; then
  # Get previous revision number
  HISTORY=$(kubectl rollout history deployment/api -n "${K8S_NAMESPACE}" 2>/dev/null || echo "")
  PREV_REV=$(echo "$HISTORY" | tail -n 2 | head -n 1 | awk '{print $1}')
  if [ -z "$PREV_REV" ] || [ "$PREV_REV" = "$CURRENT_REV" ]; then
    PREV_REV=$((CURRENT_REV - 1))
  fi
  if [ "$PREV_REV" -lt 1 ]; then
    error "No previous revision available"
    exit 1
  fi
  log "Rolling back to revision ${PREV_REV}..."
  kubectl rollout undo deployment/api -n "${K8S_NAMESPACE}" --to-revision="${PREV_REV}"
  kubectl rollout undo deployment/web -n "${K8S_NAMESPACE}" --to-revision="${PREV_REV}"
  success "Rolled back to revision ${PREV_REV}"

elif echo "$TARGET" | grep -Eq '^[0-9]+$'; then
  log "Rolling back to specific revision ${TARGET}..."
  kubectl rollout undo deployment/api -n "${K8S_NAMESPACE}" --to-revision="${TARGET}"
  kubectl rollout undo deployment/web -n "${K8S_NAMESPACE}" --to-revision="${TARGET}"
  success "Rolled back to revision ${TARGET}"

else
  # Assume it's an image tag; revert by setting image
  log "Rolling back to image tag: ${TARGET}"
  kubectl set image deployment/api \
    api="${REGISTRY:-ghcr.io}/$(basename "$REPO_ROOT")/api:${TARGET}" \
    -n "${K8S_NAMESPACE}"
  kubectl set image deployment/web \
    web="${REGISTRY:-ghcr.io}/$(basename "$REPO_ROOT")/web:${TARGET}" \
    -n "${K8S_NAMESPACE}"
fi

# ---------------------------------------------------------------------------
# Wait for rollback to complete
# ---------------------------------------------------------------------------
log "Waiting for rollback to complete..."
kubectl rollout status deployment/api -n "${K8S_NAMESPACE}" --timeout "${ROLLOUT_TIMEOUT}" || {
  error "API rollback failed"
  kubectl get pods -n "${K8S_NAMESPACE}"
  exit 1
}

kubectl rollout status deployment/web -n "${K8S_NAMESPACE}" --timeout "${ROLLOUT_TIMEOUT}" || {
  error "Web rollback failed"
  kubectl get pods -n "${K8S_NAMESPACE}"
  exit 1
}

# ---------------------------------------------------------------------------
# Health checks
# ---------------------------------------------------------------------------
log "Running post-rollback health checks..."
kubectl port-forward svc/api 8080:80 -n "${K8S_NAMESPACE}" &
PF_PID=$!
sleep 5

cleanup() { kill "$PF_PID" 2>/dev/null || true; }
trap cleanup EXIT

for attempt in 1 2 3; do
  if curl -sf http://localhost:8080/health >/dev/null; then
    success "Health check passed"
    break
  fi
  if [ "$attempt" -eq 3 ]; then
    warn "Health check failed — rollback may be incomplete"
  fi
  sleep 5
done

curl -sf http://localhost:8080/ready >/dev/null || warn "Readiness check failed"

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
NEW_REV=$(kubectl get deployment/api -n "${K8S_NAMESPACE}" -o jsonpath='{.metadata.annotations.deployment\.kubernetes\.io/revision}' 2>/dev/null || echo "unknown")

log "=== Rollback Summary ==="
success "Rollback complete!"
info  "Previous revision: ${CURRENT_REV}"
info  "Current revision:  ${NEW_REV}"
info  "Namespace:       ${K8S_NAMESPACE}"

kubectl get pods -n "${K8S_NAMESPACE}" -o wide

log "If issues persist, investigate logs:"
echo "  kubectl logs -n ${K8S_NAMESPACE} deployment/api --tail=100"
echo "  kubectl describe deployment/api -n ${K8S_NAMESPACE}"
