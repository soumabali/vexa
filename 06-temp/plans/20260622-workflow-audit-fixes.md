# 20260622 — Workflow Audit Fixes

## Goal

Perbaiki issue yang ditemukan saat audit workflow vexa dan buat mekanisme pencegahan agar issue serupa tidak terulang.

## Scope

### In Scope

1. Fix double section di `vexa Progress.md`.
2. Update root `06-temp/plans/README.md` agar mencakup semua plan.
3. Hapus git remote invalid `02-application-remote`.
4. Verifikasi / setup volume backup untuk DB dan WireGuard.
5. Buat pre-push checklist dan audit script.

### Out of Scope

- Perubahan kode aplikasi (P5 TOTP, fitur baru).
- Arsip `apps/desktop/src-frontend/` (bisa jadi plan terpisah).
- Redeploy production.

## Files

- `/home/ubuntu/Documents/Obsidian Vault/infra/vexa Progress.md`
- `/home/ubuntu/projects/vexa/06-temp/plans/README.md`
- `/home/ubuntu/projects/vexa/.git/config`
- `/home/ubuntu/projects/vexa/02-application/docker-compose.prod.yml`
- `/home/ubuntu/projects/vexa/00-meta/audit-checklist.md` (baru)
- `/home/ubuntu/projects/vexa/00-meta/pre-push-checklist.md` (baru)
- `/home/ubuntu/projects/vexa/scripts/audit-vexa.sh` (baru)

## Verification

- `vexa Progress.md` tidak ada double section.
- `06-temp/plans/README.md` mencakup semua plan aktif/arsip.
- `git remote -v` tidak ada `02-application-remote`.
- Volume `/backups` tersedia untuk container production.
- `scripts/audit-vexa.sh` bisa dijalankan tanpa error.

## Rollback

- Git remote bisa ditambahkan ulang kalau dibutuhkan (tidak diperlukan).
- `vexa Progress.md` lama bisa di-recover dari git history.

## Links

- Obsidian: [[vexa Progress]]
- Obsidian: [[vexa Rules]]
- Obsidian: [[vexa Workflow Runbook]]
- Root: `00-meta/git-structure.md`
