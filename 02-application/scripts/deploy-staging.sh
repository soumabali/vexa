#!/usr/bin/env bash
# =============================================================================
# SSH Manager — Deploy to Staging
# Usage: ./scripts/deploy-staging.sh [IMAGE_TAG]
# =============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

REGISTRY="${REGISTRY:-ghcr.io}"
IMAGE_NAME="${IMAGE_NAME:-$(basename "$REPO_ROOT")}"
K8S_NAMESPACE="${K8S_NAMESPACE:-ssh-manager-staging}"
IMAGE_TAG="${1:-latest}"
KUBECONFIG="${KUBECONFIG:-${REPO_ROOT}/infra/k8s/kubeconfig-staging}"
HELM_TIMEOUT="${HELM_TIMEOUT:-5m}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log() {
  echo -e "${BLUE}[$(date +%Y-%m-%d\ %H:%M:%S)]${NC} $*"
}
warn() {
  echo -e "${YELLOW}[WARN]${NC} $*"
}
error() {
  echo -e "${RED}[ERROR]${NC} $*"
}
success() {
  echo -e "${GREEN}[OK]${NC} $*"
}

# ---------------------------------------------------------------------------
# Pre-flight checks
# ---------------------------------------------------------------------------
log "=== Deploying to Staging ==="
log "Image tag: ${IMAGE_TAG}"
log "Namespace: ${K8S_NAMESPACE}"

if ! command -v kubectl >/dev/null 2>&1; then
  error "kubectl not found. Please install kubectl."
  exit 1
fi

if ! command -v kustomize >/dev/null 2>&1; then
  warn "kustomize not found. Attempting to use kubectl built-in..."
  KUSTOMIZE="kubectl kustomize"
else
  KUSTOMIZE="kustomize"
fi

if [ -f "$KUBECONFIG" ]; then
  export KUBECONFIG
  kubectl config current-context >/dev/null 2>&1 || {
    error "Invalid kubeconfig at ${KUBECONFIG}"
    exit 1
  }
else
  warn "Kubeconfig file not found at ${KUBECONFIG}. Using current context..."
fi

# ---------------------------------------------------------------------------
# Ensure namespace exists
# ---------------------------------------------------------------------------
log "Ensuring namespace ${K8S_NAMESPACE} exists..."
kubectl create namespace "${K8S_NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f - || true
kubectl label namespace "${K8S_NAMESPACE}" \
  pod-security.kubernetes.io/enforce=restricted \
  pod-security.kubernetes.io/enforce-version=latest \
  --overwrite 2>/dev/null || true

# ---------------------------------------------------------------------------
# Update image tags in staging overlay
# ---------------------------------------------------------------------------
log "Updating image tags in staging overlay..."
cd "${REPO_ROOT}/infra/k8s/staging"
$KUSTOMIZE edit set image "api=${REGISTRY}/${IMAGE_NAME}/api:${IMAGE_TAG}" || true
$KUSTOMIZE edit set image "web=${REGISTRY}/${IMAGE_NAME}/web:${IMAGE_TAG}" || true

# ---------------------------------------------------------------------------
# Build and apply manifests
# ---------------------------------------------------------------------------
log "Building manifests..."
$KUSTOMIZE build . > /tmp/staging-manifests.yaml

log "Applying manifests..."
kubectl apply -f /tmp/staging-manifests.yaml

# ---------------------------------------------------------------------------
# Wait for rollout
# ---------------------------------------------------------------------------
log "Waiting for deployments to roll out..."
kubectl rollout status deployment/api -n "${K8S_NAMESPACE}" --timeout "${HELM_TIMEOUT}" || {
  error "API deployment failed to roll out"
  kubectl get pods -n "${K8S_NAMESPACE}"
  kubectl describe deployment/api -n "${K8S_NAMESPACE}"
  exit 1
}

kubectl rollout status deployment/web -n "${K8S_NAMESPACE}" --timeout "${HELM_TIMEOUT}" || {
  error "Web deployment failed to roll out"
  kubectl get pods -n "${K8S_NAMESPACE}"
  kubectl describe deployment/web -n "${K8S_NAMESPACE}"
  exit 1
}

# ---------------------------------------------------------------------------
# Health checks
# ---------------------------------------------------------------------------
log "Running health checks..."
sleep 5

# Port-forward for checks
kubectl port-forward svc/api 8080:80 -n "${K8S_NAMESPACE}" &
PF_PID=$!
sleep 3

cleanup() {
  kill "$PF_PID" 2>/dev/null || true
}
trap cleanup EXIT

for attempt in 1 2 3; do
  if curl -sf http://localhost:8080/health >/dev/null; then
    success "Health check passed (attempt ${attempt})"
    break
  fi
  if [ "$attempt" -eq 3 ]; then
    error "Health check failed after 3 attempts"
    exit 1
  fi
  warn "Health check failed, retrying..."
  sleep 5
done

curl -sf http://localhost:8080/ready >/dev/null || {
  error "Readiness check failed"
  exit 1
}

success "Readiness check passed"

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
log "=== Deployment Summary ==="
success "Staging deployment complete!"
echo "  Namespace: ${K8S_NAMESPACE}"
echo "  API image: ${REGISTRY}/${IMAGE_NAME}/api:${IMAGE_TAG}"
echo "  Web image: ${REGISTRY}/${IMAGE_NAME}/web:${IMAGE_TAG}"
echo "  URL:       https://staging.ssh-manager.local"

kubectl get pods -n "${K8S_NAMESPACE}" -o wide
