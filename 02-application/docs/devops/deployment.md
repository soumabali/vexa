# vexa — Deployment Guide

> **Agent:** DevOps Engineer  
> **Status:** Generated 2026-05-28  
> **Version:** 1.0.0

---

## 1. Deployment Environments

| Environment | URL | Trigger | Strategy |
|-------------|-----|---------|----------|
| **Local** | `localhost:3000` / `localhost:8080` | Manual | Docker Compose |
| **Staging** | `https://staging.vexa.local` | Auto (push to `develop`) | Rolling |
| **Production** | `https://vexa.local` | Manual (`workflow_dispatch`) | Canary / Blue-Green |

---

## 2. Quick Start

### 2.1 Deploy to Staging (Auto)
```bash
# Simply push to develop — GitHub Actions handles the rest
git checkout develop
git merge feature/my-branch
git push origin develop
```

### 2.2 Deploy to Production (Manual)
```bash
# Via GitHub UI
# 1. Go to Actions → CD — Production
# 2. Click "Run workflow"
# 3. Enter image tag (e.g., commit SHA or version)
# 4. Select strategy (canary / blue-green / rolling)
# 5. Click "Run workflow"

# Or via CLI
gh workflow run cd-production.yml \
  -f image_tag=abc1234 \
  -f strategy=canary \
  -f canary_weight=10
```

---

## 3. Local Deployment

### 3.1 Docker Compose (Full Stack)
```bash
# Start all services
docker compose up -d

# View logs
docker compose logs -f api
docker compose logs -f web

# Scale API instances
docker compose up -d --scale api=3

# Stop all
docker compose down

# Stop and remove volumes (data loss!)
docker compose down -v
```

### 3.2 Makefile Commands
```bash
make build        # Build all
make test         # Run tests
make dev          # Start development
make deploy-staging    # Deploy to staging
make deploy-production # Deploy to production
make logs         # View logs
make db-migrate   # Run migrations
make db-seed      # Seed database
make db-backup    # Backup database
```

---

## 4. Staging Deployment

### 4.1 Prerequisites
- Kubernetes cluster (EKS/GKE/AKS or local k3d)
- `kubectl` configured with staging context
- `kustomize` installed
- Access to `ghcr.io/soumabali/vexa` images

### 4.2 Deploy via Script
```bash
./scripts/deploy-staging.sh $(git rev-parse --short HEAD)
```

### 4.3 Deploy via Kustomize
```bash
cd infra/k8s/staging
kustomize edit set image api=ghcr.io/soumabali/vexa/api:abc1234
kustomize edit set image web=ghcr.io/soumabali/vexa/web:abc1234
kustomize build . | kubectl apply -f -

# Verify
kubectl rollout status deployment/api -n vexa-staging
kubectl get pods -n vexa-staging -o wide
```

### 4.4 Staging Configuration
- **Replicas**: 2 (min) / 4 (max via HPA)
- **Resources**: 64Mi–256Mi RAM, 100m–500m CPU
- **Log level**: debug
- **MFA**: optional
- **TLS**: Let's Encrypt staging issuer
- **Network policies**: default (allow intra-namespace)

---

## 5. Production Deployment

### 5.1 Prerequisites
- Production Kubernetes cluster
- `kubectl` with production kubeconfig
- Helm 3.x (optional)
- Slack webhook configured

### 5.2 Deploy via Script
```bash
./scripts/deploy-production.sh abc1234 --strategy=canary --canary-weight=10
```

The script will:
1. Verify the image exists in GHCR
2. Capture pre-deploy snapshot
3. Apply production manifests via Kustomize
4. Wait for rollout
5. Run health checks
6. Verify security headers
7. Record deployment metadata

### 5.3 Deploy via GitHub Actions
```bash
gh workflow run cd-production.yml \
  -f image_tag=abc1234 \
  -f strategy=canary
```

### 5.4 Production Configuration
- **Replicas**: 3 (min) / 12 (max via HPA)
- **Resources**: 256Mi–1Gi RAM, 250m–1000m CPU
- **Log level**: info
- **MFA**: required
- **TLS**: 1.3 only, Let's Encrypt production issuer
- **Network policies**: strict (deny-all + explicit allow)
- **Pod Disruption Budget**: minAvailable 2
- **Topology spread**: across zones and nodes

### 5.5 Deployment Strategies

#### Canary
```bash
# Deploy 10% traffic
./scripts/deploy-production.sh abc1234 --strategy=canary --canary-weight=10

# Monitor metrics for 10 minutes
# If error rate < 1%, promote to 100%
./scripts/deploy-production.sh abc1234 --strategy=rolling
```

#### Blue-Green
```bash
# Deploy green environment
kubectl apply -k infra/k8s/overlays/production-green/

# Verify green
kubectl rollout status deployment/api-green -n vexa-prod

# Switch traffic (update ingress)
kubectl patch ingress api \
  -n vexa-prod \
  -p '{"spec":{"rules":[{"host":"vexa.local","http":{"paths":[{"path":"/","pathType":"Prefix","backend":{"service":{"name":"api-green","port":{"number":80}}}}}]}}]}'

# Tear down blue
kubectl delete -k infra/k8s/overlays/production-blue/
```

#### Rolling
```bash
# Standard rolling update (maxSurge=1, maxUnavailable=0)
./scripts/deploy-production.sh abc1234 --strategy=rolling
```

---

## 6. Rollback

### 6.1 Automatic Rollback (Script)
```bash
# Roll back to previous revision
./scripts/rollback.sh previous

# Roll back to specific revision
./scripts/rollback.sh 42

# Roll back to last (undo once)
./scripts/rollback.sh last

# Auto-confirm (for automation)
AUTO_CONFIRM=1 ./scripts/rollback.sh previous
```

### 6.2 Manual Rollback (kubectl)
```bash
# Undo last rollout
kubectl rollout undo deployment/api -n vexa-prod

# Undo to specific revision
kubectl rollout undo deployment/api -n vexa-prod --to-revision=3

# View history
kubectl rollout history deployment/api -n vexa-prod

# Roll back image only
kubectl set image deployment/api \
  api=ghcr.io/soumabali/vexa/api:previous-tag \
  -n vexa-prod
```

### 6.3 Emergency Rollback
```bash
# If deployment is broken and pods are crash-looping
kubectl rollout pause deployment/api -n vexa-prod
kubectl rollout undo deployment/api -n vexa-prod
kubectl rollout resume deployment/api -n vexa-prod
```

---

## 7. Health Checks

### 7.1 Probe Endpoints

| Probe | Endpoint | Expected | Timeout |
|-------|----------|----------|---------|
| Liveness | `/health` | HTTP 200 | 5s |
| Readiness | `/ready` | HTTP 200 | 3s |
| Startup | `/health` | HTTP 200 | 3s |
| Metrics | `/metrics` | HTTP 200 | 5s |

### 7.2 Probe Configuration
```yaml
# infra/k8s/monitoring/probes.yaml
livenessProbe:
  httpGet:
    path: /health
    port: health
    scheme: HTTPS
  initialDelaySeconds: 30
  periodSeconds: 15
  timeoutSeconds: 5
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /ready
    port: health
    scheme: HTTPS
  initialDelaySeconds: 5
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 3

startupProbe:
  httpGet:
    path: /health
    port: health
    scheme: HTTPS
  initialDelaySeconds: 10
  periodSeconds: 5
  failureThreshold: 30
```

### 7.3 Manual Health Check
```bash
# Via port-forward
kubectl port-forward svc/api 8080:80 -n vexa-prod
curl -sf http://localhost:8080/health
curl -sf http://localhost:8080/ready
curl -sf http://localhost:8080/metrics | head
```

---

## 8. SSL/TLS

### 8.1 cert-manager
```bash
# Install cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.0/cert-manager.yaml

# Create ClusterIssuer (staging)
cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-staging
spec:
  acme:
    server: https://acme-staging-v02.api.letsencrypt.org/directory
    email: admin@vexa.local
    privateKeySecretRef:
      name: letsencrypt-staging
    solvers:
    - http01:
        ingress:
          class: nginx
EOF

# Create ClusterIssuer (production)
cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@vexa.local
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
EOF
```

### 8.2 Certificate
```yaml
# Managed automatically by ingress annotation:
# cert-manager.io/cluster-issuer: "letsencrypt-prod"
```

---

## 9. Secrets Management

### 9.1 Sealed Secrets
```bash
# Install kubeseal
brew install kubeseal  # macOS
# or
wget https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.24.0/kubeseal-0.24.0-linux-amd64.tar.gz

# Seal a secret
cat <<EOF | kubeseal --controller-namespace=kube-system --format yaml > sealed-secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: vexa-secrets
  namespace: vexa
type: Opaque
stringData:
  DB_PASSWORD: "actual-password"
  JWT_SECRET: "actual-secret"
EOF

kubectl apply -f sealed-secret.yaml
```

### 9.2 Vault Integration
```bash
# Install Vault
helm install vault hashicorp/vault \
  --namespace vault \
  --create-namespace

# Enable Kubernetes auth
vault auth enable kubernetes
vault write auth/kubernetes/config \
  kubernetes_host="https://$KUBERNETES_PORT_443_TCP_ADDR:443" \
  token_reviewer_jwt="$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" \
  kubernetes_ca_cert=@/var/run/secrets/kubernetes.io/serviceaccount/ca.crt
```

---

## 10. Monitoring & Alerting

### 10.1 Prometheus Rules
Located in `infra/k8s/monitoring/prometheusrules.yaml`:
- Critical: Error rate, auth failures, vault decryption, crash loops
- High: Session proxy crashes, audit log lag, certificate expiry, API latency
- Medium: High CPU, high memory, low disk space

### 10.2 Grafana Dashboard
The base manifests include a ConfigMap with a Grafana dashboard (`infra/k8s/09-monitoring.yaml`).

### 10.3 Checking Alerts
```bash
# View firing alerts
kubectl get prometheusrules -n vexa

# Check Prometheus targets
kubectl port-forward svc/prometheus 9090:9090 -n monitoring
# Open http://localhost:9090/alerts
```

---

## 11. Troubleshooting

### 11.1 Pod CrashLoopBackOff
```bash
# Check logs
kubectl logs deployment/api -n vexa-prod --previous

# Check events
kubectl get events -n vexa-prod --sort-by='.lastTimestamp'

# Describe pod
kubectl describe pod <pod-name> -n vexa-prod
```

### 11.2 Database Connection Issues
```bash
# Check connectivity
kubectl run debug --rm -it --image=postgres:16-alpine -- \
  psql postgres://user:pass@postgres/vexa -c "SELECT 1;"

# Check pool status
kubectl exec -it deployment/api -n vexa-prod -- \
  psql "$DATABASE_URL" -c "SELECT * FROM pg_stat_activity;"
```

### 11.3 High Memory Usage
```bash
# Check metrics
kubectl top pods -n vexa-prod

# Check limits
kubectl get pods -n vexa-prod -o yaml | grep -A 5 resources

# Adjust limits
kubectl set resources deployment/api -n vexa-prod \
  --limits=memory=1Gi --requests=memory=512Mi
```

### 11.4 Deployment Stuck
```bash
# Check rollout status
kubectl rollout status deployment/api -n vexa-prod

# View replica sets
kubectl get rs -n vexa-prod

# Scale manually
kubectl scale deployment/api --replicas=3 -n vexa-prod
```

---

## 12. Maintenance Windows

### 12.1 Scheduled Maintenance
```bash
# Create maintenance window configmap
kubectl create configmap maintenance-window \
  --from-literal=start="2026-06-01T02:00:00Z" \
  --from-literal=end="2026-06-01T04:00:00Z" \
  -n vexa-prod

# Enable maintenance mode (app reads this configmap)
kubectl patch deployment api -n vexa-prod -p \
  '{"spec":{"template":{"metadata":{"annotations":{"maintenance":"true"}}}}}'

# Disable maintenance mode
kubectl patch deployment api -n vexa-prod -p \
  '{"spec":{"template":{"metadata":{"annotations":{"maintenance":"false"}}}}}'
```

---

## 13. References

- [Kustomize Docs](https://kubectl.docs.kubernetes.io/references/kustomize/)
- [Kubernetes Deployment Strategies](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#strategy)
- [cert-manager Docs](https://cert-manager.io/docs/)
- [Prometheus Operator](https://prometheus-operator.dev/)
- [Sealed Secrets](https://sealed-secrets.netlify.app/)
