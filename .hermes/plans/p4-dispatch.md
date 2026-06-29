# P4 Dispatch Prompt — Production Hardening

**Tanggal:** 2026-06-28
**Branch:** `feat/p4-production-hardening`
**Dispatcher:** Ame (Hermes)
**Executor:** Claude Code di `02-application/`

---

## 📜 Konteks untuk Claude Code

Kamu adalah Claude Code, **coding executor** untuk project **vexa** (Complete SSH Manager).
Kamu WAJIB membaca dan patuhi sebelum eksekusi:

1. `02-application/CLAUDE.md` — identitas project, role split
2. `02-application/AGENTS.md` — aturan baja (jika ada copy di dalam)
3. `00-meta/skills.md` — skill mapping
4. `06-temp/plans/P4-production-hardening.md` — **plan utama** untuk task ini

Kamu **HANYA** boleh edit file di dalam `02-application/`. Dilarang commit, push, atau
membuat `.git` di dalam `02-application/`. Ame yang akan commit & push dari root.

---

## 🎯 Tujuan

Eksekusi **5 sub-task** P4 Production Hardening secara **berurutan**. Setiap sub-task
**HARUS** punya commit-nya sendiri dengan pesan konvensional yang sudah ditentukan.

---

## 📋 Sub-task 1 — Log rotation (compose)

**File:** `02-application/docker-compose.yml`

Tambahkan konfigurasi logging untuk service `api:` dan `web:` saja (JANGAN ubah service
postgres/redis/ssh-core yang sudah punya setup sendiri):

```yaml
logging:
  driver: json-file
  options:
    max-size: "10m"
    max-file: "3"
    labels: "service,env"
```

Tambah komentar singkat di file compose: `# P4 #1: log rotation`.

**Commit:**
```
chore(compose): log rotation untuk api+web (P4 #1)
```

---

## 📋 Sub-task 2 — PostgreSQL automated backups

**File existing:** `02-application/scripts/backup.sh`, `02-application/scripts/backup.crontab`

### Enhance `scripts/backup.sh`:
- Output: `vexa_${DB_NAME}_$(date +%Y%m%d_%H%M%S).sql.gz` di `backups/postgres/`
- Retention: 14 hari daily + 8 minggu weekly (Sunday)
- Env vars: `BACKUP_DIR`, `BACKUP_RETENTION_DAYS`, `BACKUP_RETENTION_WEEKS`
- Set `set -euo pipefail` di atas
- Logging ke `logs/backup.log` via `tee`
- Verify dump dengan `gunzip -t` sebelum dianggap sukses
- Return exit code 0 sukses / non-zero gagal

### Enhance `scripts/backup.crontab`:
- Daily jam 03:00 — `0 3 * * * backup.sh daily`
- Weekly Minggu jam 04:00 — `0 4 * * 0 backup.sh weekly`

### Tambah `docs/devops/postgres-backup-restore.md`:
- Cara install crontab server production
- Cara restore: `gunzip -c backup.sql.gz | psql ...`
- Runbook failure (cek log, alert)

### Tambah `scripts/backup_test.sh` (integration test):
- Buat postgres test, jalankan dump, restore, bandingkan row count

**Commit:**
```
feat(backup): postgresql automated backups + retention (P4 #2)
```

---

## 📋 Sub-task 3 — WireGuard key rotation workflow

### Tambah `scripts/rotate-wireguard-keys.sh`:
- Generate keypair baru: `wg genkey | tee privatekey | wg pubkey > publickey`
- Backup key lama: `backups/wireguard/keys_$(date +%s)/`
- Update server config `wg0.conf` atomically (mv dengan nama temp)
- Reload WireGuard: `wg syncconf wg0 <(wg-quick strip wg0)`
- Audit log entry di `logs/wireguard-rotation.log`
- `set -euo pipefail`

### API endpoint optional `POST /admin/wg/rotate` (jika admin auth sudah ada):
- File: `apps/api/internal/api/handlers/admin_wireguard.go` (atau path yang sesuai struktur existing — JANGAN paksa jika endpoint group admin tidak ada)
- Trigger rotation script via `os/exec`
- Pakai auditlogger existing

### Tests `apps/api/tests/wireguard_rotation_test.go`:
- Mock `exec.Command` untuk verify command dipanggil dengan args benar
- Verify audit log entry ter-create

### Update `docs/devops/wireguard-deployment.md`:
- Tambah section "Key Rotation"
- Schedule (quarterly recommended)
- Client notification procedure
- Rollback jika gagal

**Commit:**
```
feat(wireguard): key rotation workflow + audit (P4 #3)
```

---

## 📋 Sub-task 4 — Health check endpoint expansion

### Backend (`apps/api`):

**Path handler existing:** `apps/api/internal/api/handlers/health.go` (cek — sesuaikan dengan struktur actual)

- Endpoint `/health/live` — liveness, return 200 jika process hidup
- Endpoint `/health/ready` — readiness, cek dependency:
  - Postgres: `db.Ping()`
  - Redis: `redis.Ping()`
  - `JWT_SECRET` env present
  - Disk space: `>= 10% free` di data dir
  - WireGuard interface aktif (opsional, flag `WG_CHECK_ENABLED`)
- Response JSON: `{ status, checks: { db: {ok, latency_ms}, redis: {...}, disk: {...}, wg: {...} } }`
- HTTP 200 jika semua ok, 503 jika ada yang gagal

### Tests (`apps/api/tests/health_test.go`):
- Tambah test cases untuk setiap dependency failure mode
- Mock DB dan Redis (testify/mock atau library existing)

### Web (`apps/web/src/components/health-status.tsx`):
- Tambah panel per-check (bukan hanya aggregate)
- Warna hijau/merah per check
- Auto-refresh setiap 30 detik

**Commit:**
```
feat(health): expanded health endpoints + tests (P4 #4)
```

---

## 📋 Sub-task 5 — Monitoring dasar (container metrics)

### Prometheus metrics API (`apps/api`):

Tambah endpoint `GET /metrics` (text format Prometheus).

**Metrics minimum:**
- `vexa_http_requests_total{method,path,status}` counter
- `vexa_http_request_duration_seconds{method,path}` histogram
- `vexa_db_pool_active_connections` gauge
- `vexa_redis_pool_active_connections` gauge
- `vexa_uptime_seconds` gauge

Gunakan library `prometheus/client_golang` (cek `go.mod` — tambahkan jika belum ada).
**PENTING: Jangan expose `/metrics` di public tanpa auth** — risk data leak.

### Docker compose monitoring profile:

Tambah di `docker-compose.yml`:
- Profile `monitoring`
- Service `prometheus` (port 9090)
- Service `grafana` (port 3001, `http://localhost:3001`)
- Konfigurasi scrape `/metrics` dari API

### Tambah `docs/devops/monitoring.md`:
- Cara jalankan: `docker compose --profile monitoring up -d`
- Dashboard recommendation
- Alert rules minimum

**CORS**: tambah `http://localhost:3001` di `ALLOWED_ORIGINS` API jika monitoring profile dipakai.

**Commit:**
```
feat(monitoring): prometheus metrics + monitoring profile (P4 #5)
```

---

## 🚨 Pitfalls (JANGAN ULANGI)

1. **Aturan baja**: jangan ubah file di luar `02-application/`
2. **Web Dockerfile pitfall**: `pnpm install --frozen-lockfile` di prod stage, `--omit=dev` di build
3. **Jangan commit generated files** (next build, dist, dll)
4. **CORS**: tambah `http://localhost:3001` jika monitoring profile dipakai
5. **WireGuard**: script HARUS `set -euo pipefail` — partial rotation bisa lock admin keluar
6. **Backup retention**: JANGAN hapus backup lebih baru dari retention — bug fatal
7. **Metrics endpoint**: JANGAN expose di public tanpa auth — risk data leak
8. **Tidak ada `.git`** di dalam `02-application/`
9. **Tidak push dari dalam `02-application/`**

---

## ✅ Verification Gates (wajib dijalankan sebelum selesai)

```bash
# Compose config valid
cd 02-application && docker compose config --quiet

# Bash syntax check
bash -n scripts/backup.sh
bash -n scripts/rotate-wireguard-keys.sh
bash -n scripts/backup_test.sh

# Go tests
cd apps/api && go test ./tests/... -run 'TestHealth|TestWireguard' -v

# Web typecheck
cd apps/web && pnpm typecheck

# Web lint
cd apps/web && pnpm lint
```

Jika ada gate yang gagal, **fix dulu** sebelum menulis commit. Laporkan ke Ame jika ada
kompromi yang harus diambil.

---

## 📤 Output yang diharapkan

1. 5 commit di branch `feat/p4-production-hardening` (saat ini sudah aktif)
2. Semua file baru ada di tempat yang ditentukan
3. Output summary: daftar commit + verifikasi gates status

**Kerjakan tanpa konfirmasi tambahan.** Ame akan verifikasi setelah kamu selesai.

---

## 🔚 End

Setelah semua 5 sub-task selesai, tulis summary di stdout:
```
=== P4 SUMMARY ===
[ ] Sub-task 1: log rotation — <status>
[ ] Sub-task 2: postgres backup — <status>
[ ] Sub-task 3: wireguard rotation — <status>
[ ] Sub-task 4: health endpoint — <status>
[ ] Sub-task 5: monitoring — <status>

Commits: <jumlah>
Verification: <pass/fail + detail>
```
