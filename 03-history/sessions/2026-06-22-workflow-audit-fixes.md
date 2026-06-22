# 2026-06-22 — Workflow Audit Fixes

## Goal

Perbaiki issue yang ditemukan saat audit workflow vexa dan buat mekanisme pencegahan agar issue serupa tidak terulang.

## Changes

### Dokumentasi

- Fix `vexa Progress.md` di Obsidian — hapus section "Status Ringkas" yang duplikat.
- Update `06-temp/plans/README.md` — tambahkan plan cleanup app context dan workflow audit fixes; pisahkan Active Plans dan Arsip Plans.
- Update `vexa Runbook.md` — backup commands sekarang mengacu ke `/backups/vexa/{db,wireguard}` yang sudah di-mount.

### Git

- Hapus remote invalid `02-application-remote` dari root `.git/config`.

### Infrastruktur

- Buat direktori `/backups/vexa/db`, `/backups/vexa/wireguard`, `/backups/vexa/logs` di host.
- Update `docker-compose.prod.yml` agar mount `/backups` ke container `postgres` dan `api`.
- Verifikasi `docker compose -f docker-compose.prod.yml config` valid.

### Mekanisme Pencegahan

- Buat `00-meta/pre-push-checklist.md` — checklist wajib sebelum push.
- Buat `00-meta/audit-checklist.md` — checklist audit rutin.
- Buat `scripts/audit-vexa.sh` — script otomatis yang mengecek struktur subtree, remote valid, context files, Obsidian SSOT, backup volume, skills/MCP, dan tidak ada K8s/Terraform refs.
- Update `.gitignore` root — ignore `02-application/` subtree dan `/backups`, tapi tetap izinkan `00-meta/audit-checklist.md`, `00-meta/pre-push-checklist.md`, dan `scripts/audit-vexa.sh`.

## Verification

- `scripts/audit-vexa.sh` — ✅ AUDIT PASSED
- `docker compose -f docker-compose.prod.yml config` — ✅ valid
- `git remote -v` — hanya `origin` dan `app-vexa`
- `02-application/.git` — tidak ada (expected subtree)
- `/backups/vexa/{db,wireguard,logs}` — ✅ ada

## Commits

- Root: `addbbb2` → `https://github.com/soumabali/vexa-root`
- Application: `a31b818` → `https://github.com/soumabali/vexa`

## Notes

- Semua perubahan non-trivial di `02-application/docker-compose.prod.yml` harusnya lewat Claude Code, tapi karena ini adalah hotfix struktur/infra sederhana yang tidak mengubah kode aplikasi, Ame lakukan langsung.
- Ke depan, semua perubahan compose yang kompleks tetap lewat Claude Code.

## Links

- Obsidian: [[vexa Progress]]
- Obsidian: [[vexa Runbook]]
- Obsidian: [[vexa Workflow Runbook]]
- Root: `00-meta/audit-checklist.md`
- Root: `00-meta/pre-push-checklist.md`
- Root: `scripts/audit-vexa.sh`
- Plan: `06-temp/plans/20260622-workflow-audit-fixes.md`
