# 20260622 — Workflow Improvement Plan (Target: 10/10)

## Hasil Dry-Run Workflow

- **Rating saat ini: 8/10**
- Claude Code berhasil memberikan **10-point compliance confirmation**.
- Tidak ada file yang diubah setelah dry-run.
- `scripts/audit-vexa.sh` tetap **PASSED ✅**.
- Response Claude Code: 5 turns, 1.742 output tokens, exit success.

## Mengapa 8/10, Bukan 10/10?

| # | Kekuatan | Kelemahan |
|---|----------|-----------|
| 1 | Compliance notice dan hard rules ada di semua layer. | Belum ada **verifikasi otomatis** bahwa Claude Code benar-bener membaca semua file (hanya self-reported). |
| 2 | Dispatch template sudah meminta compliance confirmation. | Belum ada **schema validation** response — Ame masih harus baca manual. |
| 3 | Audit script mengecek keberadaan file dan clause. | Audit script **tidak menjalankan** dry-run real atau parse response compliance. |
| 4 | Skill stack mandatory jelas. | Belum ada **fallback** kalau Claude Code tidak membalas dengan format yang benar. |
| 5 | Workflow step-by-step jelas. | Belum ada **visual dashboard/progress** untuk melacak status setiap phase. |
| 6 | Pre-push checklist ada. | Checklist masih manual, belum ada **git hook** atau **CI gate** yang memaksa. |
| 7 | Backup volume sudah di-setup. | Belum ada **backup automation** (cron/container). |
| 8 | Session notes dan Obsidian SSOT rapi. | Belum ada **template/session generator** otomatis. |

## Target: Workflow 10/10

### 1. Compliance Verification Otomatis

- Buat `scripts/claude-compliance-check.sh` yang:
  - Menerima output JSON dari Claude Code dry-run.
  - Parse 10-point confirmation.
  - Return exit 0 hanya jika semua poin terkonfirmasi.
- Integrasikan ke `scripts/audit-vexa.sh`.

### 2. Schema-Validated Response

- Buat `00-meta/claude-response-schema.json` — format JSON yang harus dikembalikan Claude Code setiap kali di-dispatch:
  - `compliance.confirmation[10]`
  - `context_files_read[]`
  - `skills_used[]`
  - `files_changed[]`
  - `verification_results{}`
  - `summary`
- Update dispatch template untuk meminta `--output-format json` dengan schema ini.

### 3. Pre-Push Git Hook

- Buat `.git/hooks/pre-push` (atau `scripts/install-hooks.sh`) yang:
  - Menjalankan `audit-vexa.sh`.
  - Menjalankan verification gates lokal.
  - Block push kalau ada failure.

### 4. Dashboard / Progress Tracker

- Buat `00-meta/progress-dashboard.md` atau script `scripts/vexa-status.sh` yang menampilkan:
  - Status terakhir setiap phase (Plan → Dispatch → Verify → Commit → Document).
  - Commit hash terakhir root + app.
  - Health production (API/web respond).
  - Backup terakhir.

### 5. Backup Automation

- Tambahkan cron job atau container `vexa-backup` di `docker-compose.prod.yml`:
  - PostgreSQL dump harian ke `/backups/vexa/db`.
  - WireGuard config copy harian.
  - Log rotation harian.

### 6. Session Note Generator

- Buat `scripts/new-session.sh <topic>` yang:
  - Membuat file session note dari template.
  - Update `03-history/sessions/README.md`.
  - Update Obsidian Progress.

### 7. Error Recovery Playbook

- Tambahkan `infra/vexa Error Recovery Playbook.md` di Obsidian:
  - Kalau Claude Code hang.
  - Kalau verification gate merah.
  - Kalau subtree push conflict.
  - Kalau production container down.

### 8. CI Gate di Application Repo

- Update `.github/workflows/ci.yml` agar menjalankan:
  - `scripts/audit-vexa.sh`.
  - Schema validation untuk Claude Code response (jika ada PR dari bot/agent).

## Acceptance Criteria untuk 10/10

- [ ] `scripts/claude-compliance-check.sh` exit 0 setelah dry-run.
- [ ] `scripts/audit-vexa.sh` mencakup compliance check.
- [ ] Pre-push hook terpasang dan memblokir push yang gagal audit.
- [ ] `00-meta/claude-response-schema.json` ada dan digunakan.
- [ ] Backup automation berjalan (minimal PostgreSQL harian).
- [ ] Dashboard/status script bisa menampilkan health snapshot.
- [ ] Session note generator otomatis.
- [ ] Error recovery playbook di Obsidian.

## Out of Scope (biarkan untuk roadmap)

- Perubahan fitur aplikasi (P5 TOTP, P6 teams, dll).
- Redesign UI.
- Multi-region deployment.

## Links

- Obsidian: [[vexa Workflow Runbook]]
- Obsidian: [[vexa Rules]]
- Root: `AGENTS.md`
- Root: `00-meta/skills.md`
- Root: `scripts/audit-vexa.sh`
- Session dry-run: `2026-06-22-workflow-dryrun-test.md`
