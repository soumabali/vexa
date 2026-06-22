# vexa — Self-Hosted Deployment Guide

> **Agent:** DevOps Engineer  
> **Status:** Generated 2026-05-28  
> **Version:** 1.0.0

---

## 1. Deployment Environments

| Environment | URL | Strategy |
|-------------|-----|----------|
| **Local** | `localhost:3000` / `localhost:8080` | Docker Compose |
| **Production self-hosted** | your own domain | Docker Compose + reverse proxy |

---

## 2. Quick Start

```bash
# 1. Clone the repository
git clone https://github.com/soumabali/vexa.git
cd vexa

# 2. Configure environment
cp .env.example .env
# Edit .env with strong secrets and your domains

# 3. Start the full stack
docker compose up -d

# 4. Open the web UI and register the first admin
# http://localhost:3000
```

For a production deployment behind a reverse proxy, see section 4 below.

---

## 3. Local Deployment

### 3.1 Docker Compose (Full Stack)

```bash
# Start all services
docker compose up -d

# View logs
docker compose logs -f api
docker compose logs -f web

# Stop all
docker compose down

# Stop and remove volumes (data loss!)
docker compose down -v
```

### 3.2 Makefile Commands

```bash
make dev             # Start development stack
make stop            # Stop services
make logs            # Follow logs
make build           # Build all production artifacts
make verify          # Run verification gates
make db-migrate      # Run database migrations
make db-seed         # Seed database
```

---

## 4. Production Self-Hosted Deployment

### 4.1 Requirements

| Resource | Minimum | Recommended |
|----------|---------|-------------|
| CPU | 2 cores | 4 cores |
| RAM | 2 GB | 4 GB |
| Disk | 20 GB | 50 GB |
| Network | UDP 51820-51830 open if using WireGuard |

### 4.2 Environment Variables

Copy `.env.example` to `.env` and set at least:

| Variable | Purpose |
|----------|---------|
| `WEB_DOMAIN` | Public web domain |
| `API_DOMAIN` | Public API domain |
| `DB_PASSWORD` | PostgreSQL password |
| `JWT_SECRET` | JWT signing key (≥32 bytes) |
| `JWT_REFRESH_SECRET` | Refresh-token signing key (≥32 bytes) |
| `ENCRYPTION_KEY` | Credential vault encryption key (≥32 bytes) |
| `ALLOWED_ORIGINS` | CORS origin(s), e.g. `https://vexa.example.com` |

Generate secrets with:

```bash
openssl rand -hex 32
```

### 4.3 Reverse Proxy

`docker-compose.prod.yml` exposes Traefik labels. You can use any reverse proxy (Traefik, Caddy, nginx) that terminates TLS and forwards to:

- Web container port `3000`
- API container port `8080`
- WebSocket path `/ws` on the API domain

Example Caddy snippet:

```
vexa.example.com {
    reverse_proxy web:3000
}

api-vexa.example.com {
    reverse_proxy api:8080
}
```

### 4.4 Start Production Stack

```bash
docker compose -f docker-compose.prod.yml up -d
```

### 4.5 Production Configuration Notes

- `LOG_LEVEL=info`
- MFA should be enabled for all admin accounts once TOTP is ready.
- TLS 1.2+ is required; TLS 1.3 recommended.
- Keep `ALLOWED_ORIGINS` explicit; never use `*` in production.

---

## 5. WireGuard Tunnels

If you use the built-in per-user WireGuard tunnels:

```bash
# Verify the kernel module is available
modprobe wireguard
wg --version
```

If the module is not available, the API falls back to mock mode until WireGuard is installed. On Ubuntu/Debian:

```bash
apt install wireguard linux-headers-$(uname -r)
modprobe wireguard
```

See `docs/devops/wireguard-deployment.md` for more details.

---

## 6. Health Checks

### 6.1 Probe Endpoints

| Probe | Endpoint | Expected |
|-------|----------|----------|
| Liveness | `/health/live` | HTTP 200 |
| Readiness | `/health/ready` | HTTP 200 |
| Startup | `/health/live` | HTTP 200 |

### 6.2 Manual Check

```bash
curl -sf https://api.yourdomain.com/health/live
curl -sf https://api.yourdomain.com/health/ready
```

---

## 7. Secrets Management

Use `.env` for self-hosted deployments. For multi-node or team-managed deployments, prefer Docker secrets or a Vault integration instead of committing `.env`.

Never store secrets in source control. Verify with:

```bash
gitleaks detect --source . -v
```

---

## 8. Updates

```bash
# Pull latest code
git pull origin main

# Rebuild and restart
docker compose -f docker-compose.prod.yml build api web
docker compose -f docker-compose.prod.yml up -d
```

---

## 9. Rollback

```bash
# Check previous image tag
docker images | grep vexa

# Re-deploy the previous image tag
docker compose -f docker-compose.prod.yml up -d --no-deps --build=false api web
```

For database-level rollback, restore from a backup before restarting the API.

---

## 10. Backup & Recovery

### 10.1 Database Backup

```bash
docker compose -f docker-compose.prod.yml exec postgres pg_dump -U vexa vexa > backup.sql
```

### 10.2 WireGuard Config Backup

```bash
tar czf wg-config-backup.tar.gz /etc/wireguard
```

### 10.3 Restore Database

```bash
docker compose -f docker-compose.prod.yml exec -T postgres psql -U vexa vexa < backup.sql
```

---

## 11. Troubleshooting

### 11.1 API Fails to Start

```bash
docker compose logs --tail=100 api
docker compose exec api wget -qO- http://127.0.0.1:8080/health/live
```

### 11.2 Database Connection Issues

```bash
docker compose exec postgres pg_isready -U vexa
docker compose exec api env | grep DATABASE
```

### 11.3 WebSocket Terminal Not Connecting

- Ensure the reverse proxy forwards WebSocket upgrades for `/ws`.
- Verify `NEXT_PUBLIC_WS_URL` matches the public API domain.

### 11.4 WireGuard Tunnels Not Working

- Confirm `privileged: true` and `cap_add: [NET_ADMIN, SYS_MODULE]` are set.
- Confirm UDP ports 51820-51830 are open in the host firewall.
- Check `WG_MOCK_FALLBACK=false` once the kernel module is installed.

---

## 12. References

- [`docs/devops/docker-guide.md`](docker-guide.md)
- [`docs/devops/wireguard-deployment.md`](wireguard-deployment.md)
- [`docs/dev/getting-started.md`](../dev/getting-started.md)
- [Docker Compose Docs](https://docs.docker.com/compose/)
