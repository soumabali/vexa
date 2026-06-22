# WireGuard Deployment Guide

## Self-Hosted dengan Docker Compose (Recommended)

### Quick Start

```bash
# 1. Clone repo
git clone <your-repo> vexa
cd vexa

# 2. Setup environment
cp .env.example .env
# Edit .env dengan secrets yang kuat

# 3. Start services
docker compose -f docker-compose.prod.yml up -d

# 4. Verify
# - API: https://yourdomain.com/api/health
# - Web: https://yourdomain.com
# - WG: UDP ports 51820-51830 open
```

### Requirements

| Resource | Minimum | Recommended |
|----------|---------|-------------|
| CPU | 2 cores | 4 cores |
| RAM | 2 GB | 4 GB |
| Disk | 20 GB | 50 GB |
| Network | UDP 51820-51830 open |

### WireGuard Kernel Module

**Linux (bare metal / VM):**
```bash
# Check if WG kernel module available
modprobe wireguard
# or
wg --version  # should show version
```

**Kalau kernel module nggak tersedia:**
- App auto-fallback ke mock mode (simulasi, nggak bisa real networking)
- Untuk production, install wireguard kernel module:
  ```bash
  # Ubuntu/Debian
  apt install wireguard linux-headers-$(uname -r)
  modprobe wireguard
  ```

**Docker privileged mode:**
- `docker-compose.prod.yml` sudah set `privileged: true` + `network_mode: host`
- Container bisa akses host kernel WG module

### TLS / HTTPS

Caddy auto-managed Let's Encrypt untuk production domain. Untuk local testing, Caddy pakai self-signed certificate.

### Backup

```bash
# Backup database + WG configs
docker compose -f docker-compose.prod.yml exec postgres pg_dump -U vexa vexa > backup.sql
tar czf wg-config-backup.tar.gz /etc/wireguard
```

### Update

```bash
# Pull latest code, rebuild, restart
git pull origin main
docker compose -f docker-compose.prod.yml build api
docker compose -f docker-compose.prod.yml up -d
```

### Troubleshooting

| Problem | Check |
|---------|-------|
| WG tunnels nggak bisa enable | Kernel module loaded? `lsmod \| grep wireguard` |
| UDP ports blocked | Firewall: `ufw allow 51820:51830/udp` |
| Privilege error | Docker privileged mode enabled? |
| Mock mode aktif | Check logs: "running in mock mode" → install WG kernel module |
