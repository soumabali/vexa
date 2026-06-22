# vexa — CI/CD Documentation

> **Agent:** DevOps Engineer  
> **Status:** Generated 2026-05-28  
> **Version:** 1.0.0

---

## 1. Overview

The vexa CI/CD pipeline is fully automated and security-first. It covers:

- **Code quality** (lint, format, type check)
- **Testing** (unit, integration, race detection)
- **Security scanning** (SAST, SCA, container scan, secrets)
- **Build & package** (multi-arch Docker images)
- **Infrastructure validation** (K8s, Terraform, Checkov)
- **Deployment** (staging auto, production manual)
- **Monitoring** (health checks, metrics, alerting)

---

## 2. Workflow Files

| File | Trigger | Purpose |
|------|---------|---------|
| `.github/workflows/ci.yml` | Push to `main`/`develop`, PRs, nightly cron | Full CI: lint → test → build → scan → infra validate |
| `.github/workflows/pr-checks.yml` | Every PR | Fast feedback: lint, unit tests, dependency review, secrets |
| `.github/workflows/cd-staging.yml` | Push to `develop` | Auto build, scan, and deploy to staging |
| `.github/workflows/cd-production.yml` | `workflow_dispatch` | Manual deploy to production with approval gates |

---

## 3. CI Pipeline Stages

```
┌─────────┐   ┌─────────┐   ┌─────────┐   ┌─────────┐   ┌─────────┐
│  Lint   │ → │  Test   │ → │  Build  │ → │  Scan   │ → │  Infra  │
│         │   │         │   │         │   │         │   │ Validate│
└─────────┘   └─────────┘   └─────────┘   └─────────┘   └─────────┘
```

### 3.1 Lint & Format
- **Go**: `gofmt`, `golangci-lint`, `go vet`
- **TypeScript**: `ESLint`, type check
- **Rust**: `cargo clippy`, `cargo fmt`
- **Shell**: `shellcheck`

### 3.2 Test
- Go unit tests with race detection (`-race`)
- Rust `cargo test --lib`
- Coverage reports uploaded as artifacts

### 3.3 Build
- Multi-platform Docker images via Buildx
- Images pushed to GHCR (`ghcr.io/soumabali/vexa/{api,web}`)
- Caching via GitHub Actions cache (`type=gha`)

### 3.4 Security Scan
- **SAST**: `gosec` (Go), GitHub CodeQL
- **Secrets**: GitLeaks, TruffleHog
- **SCA**: Trivy filesystem scan
- **Container**: Trivy image scan (CRITICAL/HIGH)
- **Licenses**: `go-licenses`
- **Infra**: Checkov (K8s, Terraform, Dockerfile)

### 3.5 Infrastructure Validation
- `kubeval` for K8s manifest validation
- `terraform fmt` + `terraform validate`
- `helm lint` (if Helm charts present)

---

## 4. CD Pipeline

### 4.1 Staging (Auto)
```
develop push
    → Build images (api + web)
    → Trivy scan
    → kustomize build infra/k8s/staging
    → kubectl apply
    → kubectl rollout status
    → Smoke tests (health + readiness + metrics)
    → Slack notification
```

### 4.2 Production (Manual)
```
workflow_dispatch (image_tag + strategy)
    → Validate image exists in registry
    → Pre-deploy snapshot (record current revision)
    → kustomize build infra/k8s/production
    → kubectl apply
    → kubectl rollout status
    → Post-deploy health checks
    → Security headers verification
    → Record deployment metadata
    → Slack notification
```

**Deployment strategies supported:**
- `canary` (default, configurable weight)
- `blue-green`
- `rolling`

---

## 5. Scripts

| Script | Purpose |
|--------|---------|
| `scripts/deploy-staging.sh [TAG]` | Deploy to staging from local CLI |
| `scripts/deploy-production.sh [TAG]` | Deploy to production with confirmation |
| `scripts/rollback.sh [revision\|last\|previous]` | Roll back production deployment |

### Usage Examples

```bash
# Deploy latest to staging
./scripts/deploy-staging.sh $(git rev-parse --short HEAD)

# Deploy specific tag to production
./scripts/deploy-production.sh abc1234 --strategy=canary --canary-weight=10

# Roll back to previous revision
./scripts/rollback.sh previous

# Roll back to specific revision
./scripts/rollback.sh 42

# Auto-confirm (for automation)
AUTO_CONFIRM=1 ./scripts/deploy-production.sh abc1234
```

---

## 6. Environment Configs

### 6.1 Kustomize Overlays

```
infra/k8s/
├── base/                          # Shared base resources
│   ├── kustomization.yaml
│   ├── 00-namespace.yaml
│   ├── 01-configmap.yaml
│   ├── 02-secrets.yaml
│   ├── 03-postgres.yaml
│   ├── 04-redis.yaml
│   ├── 05-app.yaml
│   ├── 06-ingress.yaml
│   ├── 07-network-policies.yaml
│   ├── 08-security-policies.yaml
│   └── 09-monitoring.yaml
├── staging/                       # Staging overlay
│   ├── kustomization.yaml
│   ├── patch-replicas.yaml
│   ├── patch-resources.yaml
│   ├── patch-ingress.yaml
│   └── patch-env.yaml
└── production/                    # Production overlay
    ├── kustomization.yaml
    ├── patch-replicas.yaml
    ├── patch-resources.yaml
    ├── patch-ingress.yaml
    ├── patch-env.yaml
    ├── patch-pdb.yaml
    └── patch-network-policies.yaml
```

### 6.2 Key Differences

| Aspect | Staging | Production |
|--------|---------|------------|
| Replicas | 2 | 3 |
| HPA max | 4 | 12 |
| Resource limits | 256Mi / 500m CPU | 1Gi / 1000m CPU |
| Log level | debug | info |
| MFA required | false | true |
| TLS min version | any | 1.3 |
| Ingress issuer | letsencrypt-staging | letsencrypt-prod |
| Network policies | default | strict deny-all + explicit allow |
| PDB | minAvailable 1 | minAvailable 2 |

---

## 7. Monitoring Integration

### 7.1 Health Checks

| Probe | Endpoint | Purpose |
|-------|----------|---------|
| Liveness | `/health` | Restart unhealthy pods |
| Readiness | `/ready` | Remove from service endpoints |
| Startup | `/health` | Delay liveness/readiness until ready |
| Metrics | `/metrics` | Prometheus scrape |

### 7.2 Prometheus Rules

Located in `infra/k8s/monitoring/prometheusrules.yaml`:

| Alert | Severity | Condition |
|-------|----------|-----------|
| `SSHManagerHighErrorRate` | critical | Error rate > 5% for 5m |
| `SSHManagerAuthFailures` | critical | > 5 auth failures/sec for 2m |
| `SSHManagerVaultDecryptionFailures` | critical | > 5 failures/min |
| `SSHManagerContainerCrashLoop` | critical | Container restarting |
| `SSHManagerPodNotReady` | critical | Pod not Running for 10m |
| `SSHManagerSessionProxyCrash` | high | Proxy crash detected |
| `SSHManagerAPILatencyHigh` | high | P95 latency > 500ms for 10m |
| `SSHManagerCertificateExpiring` | high | Cert expires in < 7 days |
| `SSHManagerHighCPU` | medium | CPU > 80% for 10m |
| `SSHManagerHighMemory` | medium | Memory > 85% for 10m |
| `SSHManagerDiskSpaceLow` | medium | Disk < 15% free |

### 7.3 ServiceMonitor

```yaml
# infra/k8s/monitoring/servicemonitor.yaml
endpoints:
  - port: metrics
    path: /metrics
    interval: 15s
    scrapeTimeout: 10s
```

---

## 8. Required Secrets

Configure these in GitHub repository settings:

| Secret | Used By | Description |
|--------|---------|-------------|
| `GITHUB_TOKEN` | All workflows | Auto-provided, for GHCR push |
| `KUBECONFIG_STAGING` | cd-staging | Base64-encoded staging kubeconfig |
| `KUBECONFIG_PROD` | cd-production, rollback | Base64-encoded production kubeconfig |
| `SLACK_WEBHOOK_URL` | cd-staging, cd-production | Slack notifications |
| `GITLEAKS_LICENSE` | pr-checks | GitLeaks license key |

---

## 9. Troubleshooting

### CI Failures
```bash
# Re-run lint locally
cd apps/api && golangci-lint run ./...
cd apps/web && pnpm lint
cd packages/ssh-core && cargo clippy

# Run tests locally
cd apps/api && go test -race ./...
```

### Deployment Failures
```bash
# Check rollout status
kubectl rollout status deployment/api -n vexa-staging

# View events
kubectl get events -n vexa-staging --sort-by='.lastTimestamp'

# Describe pod
kubectl describe pod <pod-name> -n vexa-staging

# Check logs
kubectl logs -f deployment/api -n vexa-staging
```

### Rollback
```bash
# Via script
./scripts/rollback.sh previous

# Via kubectl directly
kubectl rollout undo deployment/api -n vexa-prod
```

---

## 10. References

- [GitHub Actions Docs](https://docs.github.com/en/actions)
- [Kustomize Docs](https://kubectl.docs.kubernetes.io/references/kustomize/)
- [Trivy Docs](https://aquasecurity.github.io/trivy/)
- [Prometheus Operator](https://prometheus-operator.dev/)
