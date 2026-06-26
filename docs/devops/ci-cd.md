# vexa — CI/CD Pipeline

> **Agent:** DevOps Engineer  
> **Status:** Active — 2026-06-26  
> **Version:** 2.0.0

---

## 1. Overview

The vexa CI/CD pipeline automatically builds Docker images and pushes them to
**GitHub Container Registry (GHCR)** on every push to `main`.

**Architecture:**

```
push to main
    → GitHub Actions
    → docker build (api, web, backup)
    → push to ghcr.io/soumabali/vexa-{api,web,backup}
    → server: docker compose pull && up -d
```

No local build or toolchain needed on the production server.

---

## 2. Images

| Image | Dockerfile | Context |
|-------|-----------|---------|
| `ghcr.io/soumabali/vexa-api` | `apps/api/Dockerfile` | `apps/api/` |
| `ghcr.io/soumabali/vexa-web` | `apps/web/Dockerfile` | `apps/web/` |
| `ghcr.io/soumabali/vexa-backup` | `Dockerfile.backup` | `.` (root) |

**Tags generated per push:**
- `latest` — latest main build
- `sha-<short_commit>` — specific commit (for pinning/rollback)
- `main` — branch tracking

---

## 3. Workflow File

**`.github/workflows/build-push.yml`**

- **Trigger:** push to `main`, manual `workflow_dispatch`
- **Strategy:** matrix build for api, web, backup (parallel)
- **Platform:** `linux/amd64`
- **Login:** `GITHUB_TOKEN` (auto-provided)
- **Cache:** GitHub Actions cache (`type=gha`)

---

## 4. Environment Variables (Build)

The web image requires build-time environment variables:

| Variable | Default | Notes |
|----------|---------|-------|
| `NEXT_PUBLIC_API_URL` | `https://api-vexa.nexigo.my.id` | Set via GitHub Repository Variables |
| `NEXT_PUBLIC_WS_URL` | `wss://api-vexa.nexigo.my.id` | Set via GitHub Repository Variables |

---

## 5. Per-Image Details

### 5.1 API (`vexa-api`)

- **Base:** `golang:1.25-alpine` (build), `debian:12-slim` (runtime)
- **Dependencies:** Go module download, static build with CGO disabled
- **Includes:** WireGuard tools, entrypoint script
- **Exposes:** `8080`, `51820/udp`
- **Size:** ~51MB final image

### 5.2 Web (`vexa-web`)

- **Base:** `node:22-alpine` (build + runtime)
- **Dependencies:** npm ci, Next.js production build (standalone mode)
- **Includes:** dumb-init for signal handling
- **Exposes:** `3000`
- **Size:** ~320MB final image

### 5.3 Backup (`vexa-backup`)

- **Base:** `postgres:16-alpine`
- **Tools:** bash, curl, tar, crond
- **Includes:** `scripts/backup.sh`, `scripts/backup.crontab`
- **Purpose:** Daily database backup at 02:00

---

## 6. Server Deployment

On the production server, no source code or build tools needed:

```bash
# Required only: docker-compose.prod.yml + .env
docker compose -f docker-compose.prod.yml pull
docker compose -f docker-compose.prod.yml up -d
```

---

## 7. Rollback

```bash
# Pin a specific commit
docker tag ghcr.io/soumabali/vexa-api:sha-abc1234 ghcr.io/soumabali/vexa-api:latest
docker compose -f docker-compose.prod.yml up -d
```

---

## 8. Required Permissions

| Permission | Purpose |
|------------|---------|
| `contents: read` | Checkout code |
| `packages: write` | Push to GHCR |
| `attestations: write` | Image provenance |
| `id-token: write` | OIDC token for attestation |

These are configured in the workflow file. No manual setup needed.

---

## 9. References

- [`.github/workflows/build-push.yml`](../../.github/workflows/build-push.yml)
- [`docker-compose.prod.yml`](../../docker-compose.prod.yml)
- [`docs/devops/self-hosted-deployment.md`](self-hosted-deployment.md)
- [GitHub Container Registry Docs](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
