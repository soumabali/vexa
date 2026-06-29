#!/usr/bin/env bash
# rotate-wireguard-keys.sh — Rotate WireGuard server keys for vexa (P4 #3)
#
# Behavior:
#   1. Generate a new keypair.
#   2. Back up the current server keys + wg0.conf to backups/wireguard/keys_<ts>/.
#   3. Atomically swap the new keys into wg0.conf.
#   4. Reload the live WireGuard interface via `wg syncconf`.
#   5. Append an audit entry to logs/wireguard-rotation.log.
#
# Atomicity: the old keys + config are kept under a timestamped backup dir.
# If the script aborts (set -e), the existing wg0.conf is NOT touched (we
# write a sibling file and `mv` it into place).
#
# Requirements: wg, wg-quick. Run as root (or with sudo) on the WG host.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
WG_DIR="${WG_DIR:-/etc/wireguard}"
WG_INTERFACE="${WG_INTERFACE:-wg0}"
BACKUP_ROOT="${BACKUP_ROOT:-${ROOT_DIR}/backups/wireguard}"
LOG_DIR="${LOG_DIR:-${ROOT_DIR}/logs}"
LOG_FILE="${LOG_DIR}/wireguard-rotation.log"

if [[ -f "${ROOT_DIR}/.env" ]]; then
  # shellcheck disable=SC1091
  set -a; . "${ROOT_DIR}/.env"; set +a
fi

mkdir -p "${BACKUP_ROOT}" "${LOG_DIR}"

log() {
  local ts
  ts="$(date '+%Y-%m-%d %H:%M:%S')"
  echo "[${ts}] $*" | tee -a "${LOG_FILE}"
}

fail() {
  log "ERROR: $*"
  exit 1
}

command -v wg >/dev/null 2>&1 || fail "wg not found in PATH"
command -v wg-quick >/dev/null 2>&1 || fail "wg-quick not found in PATH"

CONF_FILE="${WG_DIR}/${WG_INTERFACE}.conf"
[[ -r "${CONF_FILE}" ]] || fail "Cannot read ${CONF_FILE}"

TS="$(date +%s)"
BACKUP_DIR="${BACKUP_ROOT}/keys_${TS}"
mkdir -p "${BACKUP_DIR}"
chmod 700 "${BACKUP_DIR}"

# 1. Backup current state
cp -a "${CONF_FILE}" "${BACKUP_DIR}/${WG_INTERFACE}.conf.bak"
if [[ -f "${WG_DIR}/${WG_INTERFACE}.key" ]]; then
  cp -a "${WG_DIR}/${WG_INTERFACE}.key" "${BACKUP_DIR}/privatekey.bak"
fi
if [[ -f "${WG_DIR}/${WG_INTERFACE}.pub" ]]; then
  cp -a "${WG_DIR}/${WG_INTERFACE}.pub" "${BACKUP_DIR}/publickey.bak"
fi
chmod 600 "${BACKUP_DIR}"/* 2>/dev/null || true
log "Backup written to ${BACKUP_DIR}"

# 2. Generate new keypair
NEW_PRIV="$(umask 077 && mktemp)"
trap 'rm -f "${NEW_PRIV}" "${NEW_PUB:-}"' EXIT
NEW_PUB="$(mktemp)"
umask 077
wg genkey | tee "${NEW_PRIV}" | wg pubkey > "${NEW_PUB}"
umask 022
PRIV_KEY="$(cat "${NEW_PRIV}")"
PUB_KEY="$(cat "${NEW_PUB}")"
[[ -n "${PRIV_KEY}" && -n "${PUB_KEY}" ]] || fail "key generation produced empty output"
log "Generated new server public key: ${PUB_KEY}"

# 3. Rewrite conf file with new private key (preserve everything else)
NEW_CONF="$(mktemp)"
umask 077
awk -v newpriv="${PRIV_KEY}" '
  /^PrivateKey[[:space:]]*=/ { print "PrivateKey = " newpriv; next }
  { print }
' "${CONF_FILE}" > "${NEW_CONF}"
umask 022

# Validate config before swapping into place
chmod 600 "${NEW_CONF}"
if ! wg-quick strip "${WG_INTERFACE}" "${NEW_CONF}" >/dev/null 2>>"${LOG_FILE}"; then
  rm -f "${NEW_CONF}"
  fail "wg-quick strip validation failed — config NOT replaced"
fi

mv "${NEW_CONF}" "${CONF_FILE}"
chmod 600 "${CONF_FILE}"
log "Updated ${CONF_FILE} with new private key (atomic mv)"

# 4. Reload live interface (without dropping peers)
if ! wg syncconf "${WG_INTERFACE}" <(wg-quick strip "${WG_INTERFACE}") 2>>"${LOG_FILE}"; then
  log "WARNING: wg syncconf failed — config on disk is updated but interface is stale"
  log "Rollback: cp ${BACKUP_DIR}/${WG_INTERFACE}.conf.bak ${CONF_FILE} && wg syncconf ${WG_INTERFACE} <(wg-quick strip ${WG_INTERFACE})"
  exit 1
fi
log "WireGuard interface ${WG_INTERFACE} reloaded"

# 5. Audit entry (also machine-parseable for ingest)
log "ROTATE_OK ts=${TS} interface=${WG_INTERFACE} pubkey=${PUB_KEY} backup=${BACKUP_DIR}"

# Cleanup temp key files
rm -f "${NEW_PRIV}" "${NEW_PUB}"
trap - EXIT

log "Rotation complete"
exit 0