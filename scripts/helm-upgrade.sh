#!/usr/bin/env bash
set -euo pipefail

# Upgrade SSH Manager via Helm
# Usage: ./scripts/helm-upgrade.sh [release-name] [namespace] [values-file]

RELEASE_NAME=${1:-ssh-manager}
NAMESPACE=${2:-ssh-manager}
VALUES_FILE=${3:-deploy/helm/ssh-manager/values.yaml}
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

echo "🔄 Upgrading SSH Manager via Helm"
echo "   Release: $RELEASE_NAME"
echo "   Namespace: $NAMESPACE"
echo "   Values: $VALUES_FILE"
echo ""

# Check prerequisites
command -v helm >/dev/null 2>&1 || { echo "❌ helm not found"; exit 1; }

# Update dependencies
echo "📦 Updating Helm dependencies..."
cd deploy/helm/ssh-manager
helm dependency update
cd "$PROJECT_ROOT"

# Get current values (optional)
echo "📋 Current release info:"
helm status "$RELEASE_NAME" -n "$NAMESPACE" || true

# Upgrade with rollback on failure
echo "🔄 Upgrading..."
helm upgrade "$RELEASE_NAME" \
    ./deploy/helm/ssh-manager \
    --namespace "$NAMESPACE" \
    --values "$VALUES_FILE" \
    --wait \
    --timeout 10m \
    --atomic \
    --cleanup-on-fail

echo ""
echo "✅ Upgrade complete!"
echo ""
helm status "$RELEASE_NAME" -n "$NAMESPACE"
