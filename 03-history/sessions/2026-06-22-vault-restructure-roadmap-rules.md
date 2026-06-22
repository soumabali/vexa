# 2026-06-22 — Vault Restructure + Roadmap + Rules

## Goal

1. Rapikan Obsidian Vault sebagai SSOT yang terstruktur dan mudah dibaca.
2. Buat roadmap lengkap project vexa (P3–P12).
3. Buat aturan baja bahwa Ame hanya orkestrasi, semua eksekusi kode via Claude Code.

## Changes

### Obsidian Vault

- **New/updated SSOT files:**
  - `infra/Index.md` — index utama infrastruktur
  - `infra/vexa — Project Index.md` — SSOT project vexa
  - `infra/vexa Progress.md` — progress dan milestone
  - `infra/vexa Runbook.md` — runbook operasional
  - `infra/vexa Roadmap.md` — rencana fitur P3–P12
  - `infra/vexa Rules.md` — aturan baja project
  - `infra/vexa.md` — pointer singkat ke SSOT
  - `infra/vexa/history/` — arsip catatan lama hasil review awal

- **Archived old notes:**
  - vexa Audit Ketidaksesuaian.md
  - vexa Cleanup and Readability.md
  - vexa E2E Playwright Readiness.md
  - vexa Requirements Patch Tracker.md
  - vexa Standards and Best Practices.md
  - vexa Task 1–5 review notes
  - vexa UI-UX Design Alignment.md
  - vexa Claude Code Reviews Index.md

### Root Orchestrator

- Updated `CLAUDE.md` dengan aturan baja, skill mapping, verification gates, dan link ke Obsidian SSOT.

## Verification

- Semua file vault tertulis dengan benar.
- Struktur vault bersih dan tidak ada duplikat di root infra.
- Tidak ada kode diubah di `02-application/` (sesuai rules: Ame hanya orkestrasi).

## Notes

- Backup lama di `/home/ubuntu/projects/vexa-backup-20260622-112822` tetap dipertahankan.
- Next step: Finalisasi README open source untuk kontributor.

## Links

- [[vexa — Project Index]]
- [[vexa Progress]]
- [[vexa Roadmap]]
- [[vexa Rules]]
- [[vexa Runbook]]
