#!/usr/bin/env bash
# =============================================================================
# SSH Manager — Deploy to Production (Manual, with approval gate)
# Usage: ./scripts/deploy-production.sh [IMAGE_TAG] [--strategy=canary|blue-green|rolling]
# =============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

REGISTRY="${REGISTRY:-ghcr.io}"
IMAGE_NAME="${IMAGE_NAME:-$(basename "$REPO_ROOT")}"
K8S_NAMESPACE="${K8S_NAMESPACE:-ssh-manager-prod}"
STRATEGY="${STRATEGY:-canary}"
CANARY_WEIGHT="${CANARY_WEIGHT:-10}"
ROLLOUT_TIMEOUT="${ROLLOUT_TIMEOUT:-10m}"
KUBECONFIG="${KUBECONFIG:-${REPO_ROOT}/infra/k8s/kubeconfig-production}"

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

# ---------------------------------------------------------------------------
# Parse arguments
# ---------------------------------------------------------------------------
IMAGE_TAG=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --strategy=*)
      STRATEGY="${1#*=}"
      shift
      ;;
    --canary-weight=*)
      CANARY_WEIGHT="${1#*=}"
      shift
      ;;
    --skip-health-check)
      SKIP_HEALTH_CHECK=1
      shift
      ;;
    -*)
      error "Unknown option: $1"
      exit 1
      ;;
    *)
      if [ -z "$IMAGE_TAG" ]; then
        IMAGE_TAG="$1"
      fi
      shift
      ;;
  esac
done

if [ -z "$IMAGE_TAG" ]; then
  error "Usage: $0 [IMAGE_TAG] [--strategy=canary|blue-green|rolling] [--canary-weight=10]"
  exit 1
fi

# ---------------------------------------------------------------------------
# Pre-flight
# ---------------------------------------------------------------------------
log "=== Production Deployment ==="
info  "Strategy: ${STRATEGY}"
info  "Image:    ${REGISTRY}/${IMAGE_NAME}/api:${IMAGE_TAG}"
info  "Image:    ${REGISTRY}/${IMAGE_NAME}/web:${IMAGE_TAG}"
info  "Namespace: ${K8S_NAMESPACE}"

# Interactive confirmation
if [ "${AUTO_CONFIRM:-}" != "1" ]; then
  echo -n "Are you sure you want to deploy to production? [yes/no]: "
  read -r CONFIRM
  if [ "$CONFIRM" != "yes" ]; then
    warn "Deployment aborted by user"
    exit 1
  fi
fi

if ! command -v kubectl >/dev/null 2>&1; then
  error "kubectl not found"
  exit 1
fi

if [ -f "$KUBECONFIG" ]; then
  export KUBECONFIG
fi

# Verify kubeconfig works
kubectl config current-context >/dev/null 2>&1 || {
  error "kubectl cannot connect to cluster"
  exit 1
}

# Verify images exist
log "Verifying images in registry..."
docker manifest inspect "${REGISTRY}/${IMAGE_NAME}/api:${IMAGE_TAG}" >/dev/null 2>&1 || {
  error "API image not found in registry: ${REGISTRY}/${IMAGE_NAME}/api:${IMAGE_TAG}"
  exit 1
}
docker manifest inspect "${REGISTRY}/${IMAGE_NAME}/web:${IMAGE_TAG}" >/dev/null 2>&1 || {
  error "Web image not found in registry: ${REGISTRY}/${IMAGE_NAME}/web:${IMAGE_TAG}"
  exit 1
}
success "Images verified in registry"

# ---------------------------------------------------------------------------
# Pre-deploy snapshot
# ---------------------------------------------------------------------------
log "Capturing pre-deploy snapshot..."
kubectl get deployments -n "${K8S_NAMESPACE}" -o wide
CURRENT_API_REV=$(kubectl get deployment/api -n "${K8S_NAMESPACE}" -o jsonpath='{.metadata.annotations.deployment\.kubernetes\.io/revision}' 2>/dev/null || echo "unknown")
log "Current API revision: ${CURRENT_API_REV}"

# ---------------------------------------------------------------------------
# Deploy via Kustomize
# ---------------------------------------------------------------------------
log "Applying production manifests..."
cd "${REPO_ROOT}/infra/k8s/production"

if command -v kustomize >/dev/null 2>&1; then
  KUSTOMIZE="kustomize"
else
  KUSTOMIZE="kubectl kustomize"
fi

$KUSTOMIZE edit set image "api=${REGISTRY}/${IMAGE_NAME}/api:${IMAGE_TAG}" || true
$KUSTOMIZE edit set image "web=${REGISTRY}/${IMAGE_NAME}/web:${IMAGE_TAG}" || true

if [ "$STRATEGY" = "canary" ]; then
  # Add canary annotation to ingress or deployment
  $KUSTOMIZE edit add annotation canary-weight:"${CANARY_WEIGHT}" || true
fi

$KUSTOMIZE build . > /tmp/production-manifests.yaml
kubectl apply -f /tmp/production-manifests.yaml

# ---------------------------------------------------------------------------
# Wait for rollout
# ---------------------------------------------------------------------------
log "Waiting for deployments to roll out (timeout: ${ROLLOUT_TIMEOUT})..."
kubectl rollout status deployment/api -n "${K8S_NAMESPACE}" --timeout "${ROLLOUT_TIMEOUT}" || {
  error "API deployment failed"
  kubectl get pods -n "${K8S_NAMESPACE}"
  kubectl describe deployment/api -n "${K8S_NAMESPACE}"
  exit 1
}

kubectl rollout status deployment/web -n "${K8S_NAMESPACE}" --timeout "${ROLLOUT_TIMEOUT}" || {
  error "Web deployment failed"
  kubectl get pods -n "${K8S_NAMESPACE}"
  kubectl describe deployment/web -n "${K8S_NAMESPACE}"
  exit 1
}

# ---------------------------------------------------------------------------
# Health checks
# ---------------------------------------------------------------------------
if [ "${SKIP_HEALTH_CHECK:-0}" != "1" ]; then
  log "Running post-deploy health checks..."
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
      error "Health check failed"
      exit 1
    fi
    warn "Retrying health check..."
    sleep 5
  done

  curl -sf http://localhost:8080/ready >/dev/null || {
    error "Readiness check failed"
    exit 1
  }
  success "Readiness check passed"

  # Security headers
  headers=$(curl -sI http://localhost:8080/health || true)
  echo "$headers" | grep -i "strict-transport-security" >/dev/null || warn "HSTS header missing"
  echo "$headers" | grep -i "x-content-type-options" >/dev/null || warn "X-Content-Type-Options missing"
  echo "$headers" | grep -i "x-frame-options" >/dev/null || warn "X-Frame-Options missing"
fi

# ---------------------------------------------------------------------------
# Record deployment metadata
# ---------------------------------------------------------------------------
kubectl annotate deployment/api -n "${K8S_NAMESPACE}" \
  deployment.ssh-manager.io/deployed-at="$(date -Iseconds)" \
  deployment.ssh-manager.io/deployed-by="$(whoami)" \
  deployment.ssh-manager.io/git-sha="${IMAGE_TAG}" \
  deployment.ssh-manager.io/strategy="${STRATEGY}" \
  --overwrite 2>/dev/null || true

kubectl annotate deployment/web -n "${K8S_NAMESPACE}" \
  deployment.ssh-manager.io/deployed-at="$(date -Iseconds)" \
  deployment.ssh-manager.io/deployed-by="$(whoami)" \
  deployment.ssh-manager.io/git-sha="${IMAGE_TAG}" \
  --overwrite 2>/dev/null || true

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
log "=== Deployment Summary ==="
success "Production deployment complete!"
info  "Strategy:     ${STRATEGY}"
info  "Image tag:    ${IMAGE_TAG}"
info  "Namespace:    ${K8S_NAMESPACE}"
info  "URL:          https://ssh-manager.local"

kubectl get pods -n "${K8S_NAMESPACE}" -o wide

log "To rollback if needed, run:"
echo "  ./scripts/rollback.sh ${CURRENT_API_REV}"
