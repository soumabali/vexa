# 2026-06-22 — Backup Automation Daily

## Goal

Menambahkan backup automation harian untuk PostgreSQL, WireGuard configs, dan API logs sebagai item terakhir agar workflow vexa mencapai rating 10/10.

## Changes

### Application Repo

- **`02-application/Dockerfile.backup`** — image container backup berbasis `postgres:16-alpine` dengan `pg_dump`, `tar`, `curl`.
- **`02-application/scripts/backup.sh`** — runner backup utama:
  - Dump PostgreSQL ke `/backups/vexa/db/vexa-YYYYMMDD-HHMMSS.dump`
  - Archive WireGuard configs ke `/backups/vexa/wireguard/wg-YYYYMMDD-HHMMSS.tar.gz`
  - Archive API logs ke `/backups/vexa/logs/api-logs-YYYYMMDD-HHMMSS.tar.gz`
  - Cleanup backup lebih tua dari `RETENTION_DAYS` (default 7 hari)
- **`02-application/scripts/install-backup-cron.sh`** — install cron daily jam 02:00.
- **`02-application/docker-compose.prod.yml`** — tambah service `backup` yang mount `/backups`, `wg_config`, dan `vexa_logs`.

### Documentation

- **Obsidian `infra/vexa Runbook.md`** — update section Backup Rutin dengan instruksi installasi, manual backup, restore PostgreSQL, retention.
- **`00-meta/urls.md`** — tambah path backup dan daftar automation scripts.

## Verification

- `docker compose -f docker-compose.prod.yml config` → ✅ valid
- `audit-vexa.sh` → ✅ PASSED (backup dir check masih OK)

## Backup Schedule

- Default: **setiap hari jam 02:00** via cron pada host production.
- Manual trigger: `docker compose -f docker-compose.prod.yml run --rm backup`
- Retention: 7 hari (configurable via `BACKUP_RETENTION_DAYS`)

## Installasi di Production

```bash
cd /home/ubuntu/projects/vexa/02-application
sudo mkdir -p /backups/vexa/db /backups/vexa/wireguard /backups/vexa/logs
sudo chown -R ubuntu:ubuntu /backups
./scripts/install-backup-cron.sh
```

## Links

- `02-application/Dockerfile.backup`
- `02-application/scripts/backup.sh`
- `02-application/scripts/install-backup-cron.sh`
- `02-application/docker-compose.prod.yml`
- `00-meta/urls.md`
- Obsidian: `infra/vexa Runbook.md`
