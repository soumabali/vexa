# 2026-06-22 — Remove Orchestrator Folders from Application Repo

## Problem

GitHub application repo (`soumabali/vexa`) masih memiliki folder-folder yang seharusnya hanya ada di root orchestrator:

- `.status/` — 53 status files legacy
- `00-meta/` — checkpoint progress lama
- `03-history/` — session notes dan upstream metadata lama
- `EOF` — file kosong aneh

Folder-folder ini membuat repo aplikasi terlihat berantakan, duplikat dengan `vexa-root`, dan membingungkan kontributor open source.

## Goal

Bersihkan repo aplikasi sehingga hanya berisi **production code + contributor docs + operational scripts**. Semua orchestrator context dipindahkan ke `vexa-root` dan Obsidian Vault.

## SSOT Boundary Final

| Layer | SSOT For | Location |
|-------|----------|----------|
| Production code | Application code | `02-application/` → `https://github.com/soumabali/vexa` |
| Workflow, plans, session notes, decisions | Orchestrator context | Root (`/home/ubuntu/projects/vexa`) → `https://github.com/soumabali/vexa-root` |
| Credentials, secrets | Private vault | `/home/ubuntu/Documents/Obsidian Vault/credentials/` |
| URLs, ports, runbook | Obsidian Vault + root `00-meta/urls.md` | `00-meta/urls.md` + Obsidian `infra/` |

## Changes

### Application Repo

- **Hapus** dari `02-application/`:
  - `.status/` (53 files)
  - `00-meta/` (1 file)
  - `03-history/` (16 files)
  - `EOF`
- **Update** `02-application/.gitignore` untuk eksplisit ignore `.status/`, `00-meta/`, `03-history/`, `EOF`.
- **Update** `02-application/README.md`:
  - Tambah section "Documentation" yang menjelaskan pemisahan docs aplikasi vs orchestrator.
  - Perjelas bahwa repo ini hanya open-source contributor docs.
- **Update** `02-application/CLAUDE.md`:
  - Tambah "SSOT Boundary Summary" di bagian Documentation.
  - Perjelas tidak boleh membuat `.status/`, `00-meta/`, atau `03-history/` di application repo.
- **Update** `02-application/Makefile`:
  - Hapus Rust target dari `test` dan `build`.
  - Tambah target `backup` dan `backup-cron`.

### Root Orchestrator

- **Update** `/home/ubuntu/projects/vexa/CLAUDE.md`:
  - Tambah section "Single Source of Truth Boundaries".
  - Jelaskan bahwa `.status/`, `00-meta/`, `03-history/` hanya di root, tidak di application.

## Verification

| Check | Hasil |
|-------|-------|
| `02-application/.status` exists? | ❌ Removed |
| `02-application/00-meta` exists? | ❌ Removed |
| `02-application/03-history` exists? | ❌ Removed |
| `02-application/EOF` exists? | ❌ Removed |
| `docker-compose.prod.yml` config valid | ✅ Valid |
| `scripts/audit-vexa.sh` | ✅ PASSED |
| Pre-push hook | ✅ Executed |
| Subtree push | ✅ Berhasil |

## Git Commits

| Repo | Commit |
|------|--------|
| Root Orchestrator | `de67ddb` → `https://github.com/soumabali/vexa-root` |
| Application | `71bf135` → `https://github.com/soumabali/vexa` |

## Notes

- Application repo sekarang lebih bersih untuk open-source contributor.
- Tidak ada duplikasi folder context antara root dan application.
- Workflow tetap 10/10 karena audit dan pre-push hook masih aktif.

## Links

- Root orchestrator: `https://github.com/soumabali/vexa-root`
- Application repo: `https://github.com/soumabali/vexa`
- Obsidian: [[vexa — Project Index]]
- Obsidian: [[vexa Progress]]
