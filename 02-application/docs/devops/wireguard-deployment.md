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

## Key Rotation (P4 #3)

The server's WireGuard private key should be rotated on a regular cadence
to limit the blast radius of any leaked key material.

### Schedule

**Recommended: quarterly.** The script lives at
`scripts/rotate-wireguard-keys.sh` and is safe to run unattended.

### Manual run (on the WG host)

```bash
# As root (or with sudo) on the WG host
sudo /home/ubuntu/projects/vexa/02-application/scripts/rotate-wireguard-keys.sh
```

What it does:

1. Backs up the current `wg0.conf` + private/public keys to
   `backups/wireguard/keys_<unix_ts>/`.
2. Generates a fresh keypair.
3. Rewrites `wg0.conf` with the new private key (atomic `mv` from a
   sibling temp file — `set -euo pipefail` aborts on any failure).
4. Validates the new config with `wg-quick strip` BEFORE swapping.
5. Calls `wg syncconf wg0 <(wg-quick strip wg0)` to push the change to
   the running interface without dropping existing peer handshakes.
6. Appends a structured `ROTATE_OK` line to `logs/wireguard-rotation.log`.

### Trigger via API (admin only)

```bash
TOKEN=$(admin-login-and-jwt)
curl -X POST -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/admin/wg/rotate
```

The endpoint shells out to the same script and writes an audit entry
(`event=admin.wg.rotate`).

### Client notification

After rotation, the **server** public key changes. Active peers keep
working until their next handshake (~2 minutes) because `wg syncconf`
preserves existing peer configs. However, the rotation script leaves the
**peer entries untouched** — clients do NOT need a new config unless the
operator chose to reissue.

If you do reissue peer configs:

```bash
# Pull the new server public key
sudo wg show wg0 public-key
# Distribute via your standard peer onboarding flow
```

### Rollback

If the new interface state is bad:

```bash
sudo ls /home/ubuntu/projects/vexa/02-application/backups/wireguard/
# pick the most recent timestamped dir
sudo cp backups/wireguard/keys_<TS>/wg0.conf.bak /etc/wireguard/wg0.conf
sudo chmod 600 /etc/wireguard/wg0.conf
sudo wg syncconf wg0 <(wg-quick strip wg0)
```

### Failure modes

- `wg-quick strip` validation failed → config NOT replaced. Restore from
  backup if the live interface is stale.
- `wg syncconf` failed → live interface keeps old keys, config on disk is
  new. Manual rollback per above.
- Permission denied → run as root or via sudo.
