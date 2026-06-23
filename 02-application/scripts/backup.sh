#!/usr/bin/env sh
# backup.sh — Daily backup runner for vexa inside container
# Backs up PostgreSQL database and WireGuard configs to /backups

set -euo pipefail

BACKUP_DIR="/backups/vexa"
DATE=$(date +%Y%m%d-%H%M%S)
RETENTION_DAYS=${RETENTION_DAYS:-7}

log() { echo "[vexa-backup] $*"; }

log "Starting backup at ${DATE}"

mkdir -p "${BACKUP_DIR}/db" "${BACKUP_DIR}/wireguard" "${BACKUP_DIR}/logs"

# PostgreSQL backup
DB_BACKUP="${BACKUP_DIR}/db/vexa-${DATE}.dump"
DB_URL="${DATABASE_URL:-}"
if [ -z "$DB_URL" ] && [ -f "/home/ubuntu/projects/vexa/02-application/.env" ]; then
  # Try to extract from .env file as fallback
  DB_URL=$(grep -E '^DATABASE_URL=' /home/ubuntu/projects/vexa/02-application/.env | cut -d= -f2-)
fi
if [ -z "$DB_URL" ]; then
  log "WARNING: DATABASE_URL not set and not found in .env, skipping PostgreSQL backup"
elif ! command -v pg_dump >/dev/null 2>&1; then
  log "WARNING: pg_dump not available, skipping PostgreSQL backup (run inside vexa-backup container or install PostgreSQL client)"
else
  log "Backing up PostgreSQL to ${DB_BACKUP}"
  pg_dump "${DB_URL}" -Fc -f "${DB_BACKUP}" || {
    log "ERROR: PostgreSQL backup failed"
    exit 1
  }
fi

# WireGuard configs backup
WG_BACKUP="${BACKUP_DIR}/wireguard/wg-${DATE}.tar.gz"
log "Backing up WireGuard configs to ${WG_BACKUP}"
if [ -d "/etc/wireguard" ]; then
  tar -czf "${WG_BACKUP}" -C /etc/wireguard . || {
    log "ERROR: WireGuard backup failed"
    exit 1
  }
else
  log "WARNING: /etc/wireguard not found, skipping WireGuard backup"
fi

# API logs backup
LOG_BACKUP="${BACKUP_DIR}/logs/api-logs-${DATE}.tar.gz"
log "Backing up API logs to ${LOG_BACKUP}"
if [ -d "/var/log/vexa" ]; then
  tar -czf "${LOG_BACKUP}" -C /var/log/vexa . || {
    log "ERROR: API logs backup failed"
    exit 1
  }
else
  log "WARNING: /var/log/vexa not found, skipping log backup"
fi

# Retention cleanup
log "Cleaning up backups older than ${RETENTION_DAYS} days"
find "${BACKUP_DIR}/db" -type f -mtime +${RETENTION_DAYS} -delete
find "${BACKUP_DIR}/wireguard" -type f -mtime +${RETENTION_DAYS} -delete
find "${BACKUP_DIR}/logs" -type f -mtime +${RETENTION_DAYS} -delete

log "Backup completed successfully"
