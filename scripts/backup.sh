#!/usr/bin/env bash
# backup.sh — PostgreSQL backup for vexa (P4 #2)
# Modes:
#   backup.sh daily   — keep last BACKUP_RETENTION_DAYS days
#   backup.sh weekly  — keep last BACKUP_RETENTION_WEEKS weeks (Sunday rotation)
#
# Env:
#   DB_HOST, DB_PORT, DB_USER, DB_PASSWORD (or PGPASSWORD), DB_NAME
#   BACKUP_DIR                  (default: /home/ubuntu/projects/vexa/02-application/backups/postgres)
#   BACKUP_RETENTION_DAYS       (default: 14)
#   BACKUP_RETENTION_WEEKS      (default: 8)
#   LOG_DIR                     (default: /home/ubuntu/projects/vexa/02-application/logs)

set -euo pipefail

MODE="${1:-daily}"

# Load .env if present (silent)
if [[ -f "$(cd "$(dirname "$0")/.." && pwd)/.env" ]]; then
  # shellcheck disable=SC1091
  set -a
  . "$(cd "$(dirname "$0")/.." && pwd)/.env"
  set +a
fi

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BACKUP_DIR="${BACKUP_DIR:-${ROOT_DIR}/backups/postgres}"
LOG_DIR="${LOG_DIR:-${ROOT_DIR}/logs}"
BACKUP_RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-14}"
BACKUP_RETENTION_WEEKS="${BACKUP_RETENTION_WEEKS:-8}"

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-vexa}"
DB_NAME="${DB_NAME:-vexa}"

# Derive connection params from DATABASE_URL when present and DB_* are unset.
# Format: postgres://user:pass@host:port/dbname?sslmode=...
if [[ -n "${DATABASE_URL:-}" && ( -z "${DB_HOST:-}" || "${DB_HOST}" == "localhost" ) ]]; then
  # shellcheck disable=SC2001
  _db_rest="${DATABASE_URL#*://}"                  # user:pass@host:port/db?...
  _db_userpass="${_db_rest%%@*}"                   # user:pass
  _db_hostport="${_db_rest#*@}"                    # host:port/db?...
  _db_hostport="${_db_hostport%%/*}"               # host:port
  _db_db="${_db_rest#*/}"                          # db?...
  _db_db="${_db_db%%\?*}"                          # db
  DB_USER="${_db_userpass%%:*}"
  DB_PASSWORD="${_db_userpass#*:}"
  DB_HOST="${_db_hostport%%:*}"
  DB_PORT="${_db_hostport##*:}"
  DB_NAME="${_db_db}"
fi

PGPASSWORD="${PGPASSWORD:-${DB_PASSWORD:-}}"
export PGPASSWORD

mkdir -p "${BACKUP_DIR}" "${LOG_DIR}"
LOG_FILE="${LOG_DIR}/backup.log"

log() {
  local ts
  ts="$(date '+%Y-%m-%d %H:%M:%S')"
  echo "[${ts}] [${MODE}] $*" | tee -a "${LOG_FILE}"
}

fail() {
  log "ERROR: $*"
  exit 1
}

# --- Preflight ---
command -v pg_dump >/dev/null 2>&1 || fail "pg_dump not found in PATH"
command -v gunzip >/dev/null 2>&1 || fail "gunzip not found in PATH"

[[ -n "${DB_NAME}" ]] || fail "DB_NAME not set"

TIMESTAMP="$(date +%Y%m%d_%H%M%S)"
OUTFILE="${BACKUP_DIR}/vexa_${DB_NAME}_${TIMESTAMP}.sql.gz"

log "Starting PostgreSQL dump (db=${DB_NAME} host=${DB_HOST}:${DB_PORT} user=${DB_USER})"

# pg_dump -> gzip (atomic write via temp file).
# Note: BusyBox mktemp requires the XXXXXX to be at the END of the template,
# so we use a plain temp file and gzip to the final .sql.gz name.
TMPFILE="$(mktemp "${BACKUP_DIR}/.dump.XXXXXX")"
trap 'rm -f "${TMPFILE}"' EXIT

if ! pg_dump \
    -h "${DB_HOST}" \
    -p "${DB_PORT}" \
    -U "${DB_USER}" \
    -d "${DB_NAME}" \
    --no-owner --no-privileges --clean --if-exists \
  | gzip -c > "${TMPFILE}"; then
  fail "pg_dump failed for database ${DB_NAME}"
fi

# Verify gzip integrity
if ! gunzip -t "${TMPFILE}" 2>>"${LOG_FILE}"; then
  fail "gzip integrity check failed for ${TMPFILE}"
fi

# Move to final .sql.gz location
mv "${TMPFILE}" "${OUTFILE}"
trap - EXIT
chmod 600 "${OUTFILE}"
log "Dump OK: ${OUTFILE} ($(du -h "${OUTFILE}" | cut -f1))"

# --- Retention ---
case "${MODE}" in
  daily)
    DELETED=$(find "${BACKUP_DIR}" -maxdepth 1 -type f -name "vexa_${DB_NAME}_*.sql.gz" \
      -mtime "+${BACKUP_RETENTION_DAYS}" -print -delete | wc -l || true)
    log "Retention: removed ${DELETED} files older than ${BACKUP_RETENTION_DAYS} days"
    ;;
  weekly)
    DELETED=$(find "${BACKUP_DIR}" -maxdepth 1 -type f -name "vexa_${DB_NAME}_*.sql.gz" \
      -mtime "+$((BACKUP_RETENTION_WEEKS * 7))" -print -delete | wc -l || true)
    log "Retention (weekly): removed ${DELETED} files older than ${BACKUP_RETENTION_WEEKS} weeks"
    ;;
  *)
    fail "Unknown mode: ${MODE} (use daily|weekly)"
    ;;
esac

log "Backup completed successfully"
exit 0