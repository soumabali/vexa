#!/usr/bin/env bash
set -euo pipefail

# Install SSH Manager via Helm
# Usage: ./scripts/helm-install.sh [release-name] [namespace] [values-file]

RELEASE_NAME=${1:-ssh-manager}
NAMESPACE=${2:-ssh-manager}
VALUES_FILE=${3:-deploy/helm/ssh-manager/values.yaml}
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

echo "📦 Installing SSH Manager via Helm"
echo "   Release: $RELEASE_NAME"
echo "   Namespace: $NAMESPACE"
echo "   Values: $VALUES_FILE"
echo ""

# Check prerequisites
command -v helm >/dev/null 2>&1 || { echo "❌ helm not found"; exit 1; }

# Add dependencies
echo "📦 Adding Helm dependencies..."
cd deploy/helm/ssh-manager
helm dependency update
cd "$PROJECT_ROOT"

# Create namespace
kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

# Install/upgrade chart
echo "📦 Installing Helm chart..."
helm upgrade --install "$RELEASE_NAME" \
    ./deploy/helm/ssh-manager \
    --namespace "$NAMESPACE" \
    --values "$VALUES_FILE" \
    --wait \
    --timeout 10m \
    --atomic

echo ""
echo "✅ Installation complete!"
echo ""
helm status "$RELEASE_NAME" -n "$NAMESPACE"
