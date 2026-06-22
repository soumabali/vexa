#!/usr/bin/env bash
set -euo pipefail

# Deploy SSH Manager to Kubernetes
# Usage: ./scripts/deploy-k8s.sh [environment] [namespace]

ENV=${1:-production}
NAMESPACE=${2:-ssh-manager}
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

echo "🚀 Deploying SSH Manager to Kubernetes"
echo "   Environment: $ENV"
echo "   Namespace: $NAMESPACE"
echo ""

# Check prerequisites
echo "📋 Checking prerequisites..."
command -v kubectl >/dev/null 2>&1 || { echo "❌ kubectl not found"; exit 1; }
command -v helm >/dev/null 2>&1 || { echo "❌ helm not found"; exit 1; }

# Verify cluster connection
echo "🔌 Verifying cluster connection..."
kubectl cluster-info >/dev/null || { echo "❌ Cannot connect to cluster"; exit 1; }

# Create namespace if not exists
echo "📦 Creating namespace..."
kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

# Apply K8s manifests
echo "📦 Applying K8s manifests..."
kubectl apply -f deploy/k8s/namespace.yaml
kubectl apply -f deploy/k8s/configmap.yaml -n "$NAMESPACE"
kubectl apply -f deploy/k8s/secret.yaml -n "$NAMESPACE"
kubectl apply -f deploy/k8s/deployment.yaml -n "$NAMESPACE"
kubectl apply -f deploy/k8s/service.yaml -n "$NAMESPACE"

# Apply ingress if domain is configured
if [[ -f deploy/k8s/ingress.yaml ]]; then
    echo "🌐 Applying ingress..."
    kubectl apply -f deploy/k8s/ingress.yaml -n "$NAMESPACE"
fi

# Wait for deployment
echo "⏳ Waiting for deployment to be ready..."
kubectl rollout status deployment/ssh-manager-api -n "$NAMESPACE" --timeout=300s
kubectl rollout status deployment/ssh-manager-web -n "$NAMESPACE" --timeout=300s

# Verify deployment
echo "✅ Verifying deployment..."
kubectl get pods -n "$NAMESPACE"
kubectl get svc -n "$NAMESPACE"

echo ""
echo "🎉 Deployment complete!"
echo ""
echo "Access your application at:"
echo "  API: kubectl port-forward svc/ssh-manager-api 8080:8080 -n $NAMESPACE"
echo "  Web: kubectl port-forward svc/ssh-manager-web 3000:80 -n $NAMESPACE"
