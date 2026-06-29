# P4 — Production Hardening

> Source of truth: `Obsidian/infra/vexa Roadmap.md` § P4
> Status: ⬜ belum dimulai → 🟡 in-progress setelah Claude Code selesai
> Repo: `soumabali/vexa` (public), `02-application/` adalah working dir Claude Code

## Scope (5 sub-tasks per Obsidian)

| # | Sub-task | Lokasi | Status awal |
|---|----------|--------|-------------|
| 1 | Log rotation API + web | `02-application/docker-compose.yml` | 🟡 healthcheck sudah ada, log rotation belum |
| 2 | PostgreSQL automated backups | `02-application/scripts/backup.sh` (67 baris) + `backup.crontab` | 🟡 script dasar sudah ada, perlu enhance |
| 3 | WireGuard key rotation workflow | `apps/api/internal/models/wireguard.go` | 🟡 model sudah ada, workflow belum |
| 4 | Health check endpoint expansion | `apps/api/tests/health_test.go` + `apps/web/src/components/health-status.tsx` | 🟡 endpoint dasar sudah ada |
| 5 | Monitoring dasar (container metrics) | belum ada | 🔴 hijau |

---

## Sub-task 1 — Log rotation API + web container

**Tujuan:** mencegah log file menumpuk di host production.

**Yang harus dilakukan Claude Code:**
- Tambah konfigurasi `logging:` driver di service `api:` dan `web:` di `docker-compose.yml`:
  ```yaml
  logging:
    driver: json-file
    options:
      max-size: "10m"
      max-file: "3"
      labels: "service,env"
  ```
- Pastikan TIDAK mengubah service lain (postgres/redis/ssh-core) yang sudah punya setup sendiri.
- Tambah komentar singkat di file compose bahwa ini memenuhi P4 #1.

**Acceptance:**
- `docker compose config` lulus tanpa error.
- Service api/web punya `logging.options.max-size = 10m`.

---

## Sub-task 2 — PostgreSQL automated backups

**Tujuan:** backup harian otomatis + retention policy + error handling.

**File existing:** `02-application/scripts/backup.sh` (dasar), `backup.crontab` (dasar).

**Yang harus dilakukan Claude Code:**

1. **Enhance `scripts/backup.sh`:**
   - Output filename: `vexa_${DB_NAME}_$(date +%Y%m%d_%H%M%S).sql.gz` ke `backups/postgres/`.
   - Retention policy: hapus backup lebih dari 14 hari (daily) + 8 minggu (weekly Sunday).
   - Variable env: `BACKUP_DIR`, `BACKUP_RETENTION_DAYS`, `BACKUP_RETENTION_WEEKS`.
   - Set `set -euo pipefail` di atas.
   - Logging ke `logs/backup.log` via `tee`.
   - Verify dump dengan `gunzip -t` sebelum dianggap sukses.
   - Return exit code 0 sukses / non-zero gagal.

2. **Enhance `scripts/backup.crontab`:**
   - Daily jam 03:00 WIB (`0 3 * * *`) — `backup.sh daily`.
   - Weekly Minggu jam 04:00 (`0 4 * * 0`) — `backup.sh weekly`.

3. **Dokumentasi di `docs/devops/postgres-backup-restore.md`:**
   - Cara install crontab di server production.
   - Cara restore: `gunzip -c backup.sql.gz | psql ...`.
   - Runbook failure (cek log, alert).

4. **Tests:** tambah `scripts/backup_test.sh` (integration test) — jalankan dump → restore ke db test → bandingkan row count.

**Acceptance:**
- `bash -n scripts/backup.sh` lulus.
- Script dijalankan lokal tanpa error (skip jika tidak ada postgres running).
- Test file ada dan executable.

---

## Sub-task 3 — WireGuard key rotation workflow

**Tujuan:** workflow yang aman dan terdokumentasi untuk rotate WireGuard keys.

**Yang harus dilakukan Claude Code:**

1. **Tambah `scripts/rotate-wireguard-keys.sh`:**
   - Generate keypair baru via `wg genkey | tee privatekey | wg pubkey > publickey`.
   - Backup key lama ke `backups/wireguard/keys_$(date +%s)/`.
   - Update server config `wg0.conf` atomically (mv dengan nama temp).
   - Reload WireGuard interface: `wg syncconf wg0 <(wg-quick strip wg0)`.
   - Audit log entry ke `logs/wireguard-rotation.log`.
   - `set -euo pipefail`.

2. **API endpoint optional `POST /admin/wg/rotate`** (jika admin auth sudah ada):
   - Trigger rotation script via `os/exec`.
   - Require admin role check.
   - Audit log ke existing audit logger (lihat P2 security foundation).
   - Tambah di `apps/api/internal/handlers/admin_wireguard.go` (atau path yang sesuai struktur existing).

3. **Tests:**
   - `apps/api/tests/wireguard_rotation_test.go` — mock `exec.Command`, verify command dipanggil dengan args benar.
   - Verify audit log entry ter-create.

4. **Dokumentasi di `docs/devops/wireguard-deployment.md`** (existing) — tambah section "Key Rotation":
   - Schedule (quarterly recommended).
   - Client notification procedure (jika ada).
   - Rollback jika gagal.

**Acceptance:**
- Script `bash -n` lulus.
- Test rotasi pass.
- Docs updated.

---

## Sub-task 4 — Health check endpoint expansion

**Tujuan:** health endpoint informatif untuk readiness probe + monitoring.

**Existing:** `apps/api/tests/health_test.go` (4166 bytes), `apps/web/src/components/health-status.tsx` (5885 bytes).

**Yang harus dilakukan Claude Code:**

1. **Backend (`apps/api`):**
   - Endpoint `/health/live` — liveness, return 200 jika process hidup (minimal check).
   - Endpoint `/health/ready` — readiness, cek dependency:
     - Postgres: `db.Ping()`.
     - Redis: `redis.Ping()`.
     - JWT_SECRET env present.
     - Disk space: `>= 10% free` di data dir.
     - WireGuard interface aktif (opsional, flag `WG_CHECK_ENABLED`).
   - Response JSON: `{ status, checks: { db: {ok, latency_ms}, redis: {...}, disk: {...}, wg: {...} } }`.
   - HTTP 200 jika semua ok, 503 jika ada yang gagal.

2. **Tests (`apps/api/tests/health_test.go`):**
   - Tambah test cases untuk setiap dependency failure mode.
   - Mock DB dan Redis dengan testify/mock (atau library existing).

3. **Web (`apps/web/src/components/health-status.tsx`):**
   - Tambah panel per-check (bukan hanya aggregate).
   - Warna hijau/merah per check.
   - Auto-refresh setiap 30 detik.

**Acceptance:**
- `go test ./apps/api/tests/...` lulus untuk semua health test.
- Web component render tanpa error.
- Response shape konsisten dengan dokumentasi.

---

## Sub-task 5 — Monitoring dasar (container metrics)

**Tujuan:** observability minimal untuk production.

**Yang harus dilakukan Claude Code:**

1. **Prometheus metrics di API (`apps/api`):**
   - Tambah endpoint `GET /metrics` (text format Prometheus).
   - Metrics minimum:
     - `vexa_http_requests_total{method,path,status}` — counter.
     - `vexa_http_request_duration_seconds{method,path}` — histogram.
     - `vexa_db_pool_active_connections` — gauge.
     - `vexa_redis_pool_active_connections` — gauge.
     - `vexa_uptime_seconds` — gauge.
   - Gunakan library `prometheus/client_golang` (cek apakah sudah ada di `go.mod`, jika belum tambahkan).

2. **Docker compose monitoring profile (opsional, dokumentasikan):**
   - Di `docker-compose.yml`, tambah profile `monitoring` dengan service:
     - `prometheus` — image `prom/prometheus:latest`, mount `./monitoring/prometheus.yml`.
     - `grafana` — image `grafana/grafana:latest`, port `3001:3000`.
   - Contoh `monitoring/prometheus.yml` di repo.

3. **Dokumentasi `docs/devops/monitoring.md`:**
   - Cara enable: `docker compose --profile monitoring up -d`.
   - Default dashboards yang berguna.
   - Alert dasar (jika ada AlertManager).

**Acceptance:**
- `/metrics` endpoint return Prometheus text format.
- `go build ./...` lulus setelah tambah dependency.
- Docs ditulis.

---

## Verification gates (Ame jalankan setelah Claude Code selesai)

```bash
cd /home/ubuntu/projects/vexa/02-application

# 1. Compose syntax
docker compose config -q

# 2. Shell syntax
bash -n scripts/backup.sh
bash -n scripts/rotate-wireguard-keys.sh

# 3. Go tests
cd apps/api && go test ./tests/... -count=1

# 4. Web lint + typecheck
cd ../web && pnpm lint && pnpm typecheck

# 5. Smoke test endpoint
docker compose up -d postgres redis api
curl -sf http://localhost:8080/health/ready | jq .
curl -sf http://localhost:8080/metrics | head
docker compose down -v
```

---

## Commit strategy

- Branch: `feat/p4-production-hardening`
- 5 commits (1 per sub-task), conventional commit prefix:
  - `chore(compose): log rotation untuk api+web (P4 #1)`
  - `feat(backup): postgresql automated backups + retention (P4 #2)`
  - `feat(wireguard): key rotation workflow + audit (P4 #3)`
  - `feat(health): expanded health endpoints + tests (P4 #4)`
  - `feat(monitoring): prometheus metrics + monitoring profile (P4 #5)`
- Setelah semua commit + verification gates hijau → push.
- Tulis session note di `03-history/sessions/2026-06-28-p4-production-hardening.md`.

---

## Pitfalls untuk Claude Code

- **Aturan baja**: jangan ubah file di luar scope `02-application/`.
- **Web Dockerfile pitfall** (memory): `pnpm install --frozen-lockfile` di prod stage, `--omit=dev` di build.
- **Jangan commit generated files** (next build, dist, dll).
- **CORS**: tambah `http://localhost:3001` (grafana) di `ALLOWED_ORIGINS` jika monitoring profile dipakai.
- **WireGuard**: script harus `set -euo pipefail` — partial rotation bisa lock admin keluar.
- **Backup retention**: JANGAN hapus backup lebih baru dari retention — bug fatal.
- **Metrics endpoint**: JANGAN expose di public tanpa auth — risk data leak (path lengths, dll).

---

## Definition of Done

- [ ] Semua 5 sub-task punya commit di branch.
- [ ] Verification gates hijau (compose, shell, go test, web lint/typecheck, smoke test).
- [ ] Docs updated (postgres-backup-restore.md, wireguard-deployment.md, monitoring.md).
- [ ] Push ke remote.
- [ ] Session note ditulis ke `03-history/sessions/`.
- [ ] Status P4 di Obsidian roadmap diupdate ke ✅.
