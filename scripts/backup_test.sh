#!/usr/bin/env bash
# backup_test.sh — Integration test for backup.sh (P4 #2)
# Spins up a throwaway postgres, runs backup.sh, restores to a fresh DB,
# then compares row counts between source and restored DB.
#
# Requires: docker (for ephemeral postgres), psql, pg_dump, gunzip.
# Tears everything down on exit.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
BACKUP_SH="${SCRIPT_DIR}/backup.sh"
TEST_NET="vexa-backup-test-net"
SRC_CONTAINER="vexa-pg-src"
DST_CONTAINER="vexa-pg-dst"
DB_NAME="vexa_test"
DB_USER="vexa"
DB_PASS="testpass"

if ! command -v docker >/dev/null 2>&1; then
  echo "ERROR: docker required" >&2
  exit 2
fi

cleanup() {
  docker rm -f "${SRC_CONTAINER}" "${DST_CONTAINER}" >/dev/null 2>&1 || true
  docker network rm "${TEST_NET}" >/dev/null 2>&1 || true
}
trap cleanup EXIT

log() { echo "[backup_test] $*"; }

log "Creating test network"
docker network create "${TEST_NET}" >/dev/null

log "Starting source postgres"
docker run -d --rm \
  --name "${SRC_CONTAINER}" \
  --network "${TEST_NET}" \
  -e POSTGRES_USER="${DB_USER}" \
  -e POSTGRES_PASSWORD="${DB_PASS}" \
  -e POSTGRES_DB="${DB_NAME}" \
  postgres:16-alpine >/dev/null

log "Starting restore-target postgres"
docker run -d --rm \
  --name "${DST_CONTAINER}" \
  --network "${TEST_NET}" \
  -e POSTGRES_USER="${DB_USER}" \
  -e POSTGRES_PASSWORD="${DB_PASS}" \
  -e POSTGRES_DB="${DB_NAME}_restored" \
  postgres:16-alpine >/dev/null

log "Waiting for source to be ready"
for _ in $(seq 1 30); do
  if docker exec "${SRC_CONTAINER}" pg_isready -U "${DB_USER}" >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

log "Seeding test schema + data"
docker exec -i "${SRC_CONTAINER}" psql -U "${DB_USER}" -d "${DB_NAME}" >/dev/null <<'SQL'
CREATE TABLE IF NOT EXISTS users (id serial PRIMARY KEY, email text NOT NULL);
INSERT INTO users (email) SELECT 'user_' || g || '@example.com' FROM generate_series(1, 50) g;
SELECT count(*) AS source_users FROM users;
SQL

SRC_COUNT=$(docker exec "${SRC_CONTAINER}" psql -U "${DB_USER}" -d "${DB_NAME}" -At -c "SELECT count(*) FROM users;")
log "Source row count: ${SRC_COUNT}"

# Run backup.sh with env pointing at the source container
TEST_BACKUP_DIR="$(mktemp -d)"
export BACKUP_DIR="${TEST_BACKUP_DIR}"
export LOG_DIR="${TEST_BACKUP_DIR}"
export DB_HOST="${SRC_CONTAINER}"
export DB_PORT="5432"
export DB_USER="${DB_USER}"
export DB_NAME="${DB_NAME}"
export PGPASSWORD="${DB_PASS}"
export BACKUP_RETENTION_DAYS="1"
export BACKUP_RETENTION_WEEKS="1"

log "Running backup.sh daily"
"${BACKUP_SH}" daily

DUMP_FILE="$(ls -1 "${TEST_BACKUP_DIR}"/vexa_${DB_NAME}_*.sql.gz | head -n 1)"
[[ -f "${DUMP_FILE}" ]] || { log "ERROR: no dump file produced"; exit 1; }
log "Dump artifact: ${DUMP_FILE}"

log "Restoring dump into target container"
gunzip -c "${DUMP_FILE}" \
  | docker exec -i "${DST_CONTAINER}" psql -U "${DB_USER}" -d "${DB_NAME}_restored"

log "Waiting for restore-target to be ready"
for _ in $(seq 1 30); do
  if docker exec "${DST_CONTAINER}" pg_isready -U "${DB_USER}" >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

DST_COUNT=$(docker exec "${DST_CONTAINER}" psql -U "${DB_USER}" -d "${DB_NAME}_restored" -At -c "SELECT count(*) FROM users;" || echo 0)
log "Restored row count: ${DST_COUNT}"

if [[ "${SRC_COUNT}" != "${DST_COUNT}" ]]; then
  log "FAIL: row count mismatch (src=${SRC_COUNT} dst=${DST_COUNT})"
  exit 1
fi

log "PASS: backup + restore round-trip OK (${SRC_COUNT} rows)"
exit 0