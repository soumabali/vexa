# 2026-06-22 — Make Claude Skills/MCP Mandatory

## Goal

Memperjelas bahwa skill/MCP tertentu **wajib** digunakan oleh Claude Code pada setiap sesi coding vexa. Sebelumnya hanya "terdaftar" tapi tidak eksplisit wajib.

## Changes

### Obsidian Vault

- `infra/vexa Workflow Runbook.md`
  - Tambah section **Claude Code Skill Stack (WAJIB)**.
  - Tambah section **Hermes Skills (Ame)**.
  - Jelaskan cara verifikasi dan cara menggunakan setiap skill.

- `infra/vexa Claude Code Setup.md`
  - Ganti section skill menjadi **"Skill & MCP yang Harus Aktif (WAJIB)"**.
  - Tambah penjelasan: stack standard, bukan optional, harus aktif setiap sesi.
  - Tambah cara verifikasi dan cara menggunakan.

### Root Orchestrator

- `CLAUDE.md`
  - Update section skill menjadi **"Claude Code Skills/MCP (WAJIB)"**.
  - Tambah required prompt phrase.
  - Tambah verifikasi via `scripts/audit-vexa.sh`.

- `00-meta/skills.md`
  - Rewrite: mandatory table, verification commands, required prompt phrase.
  - Update dispatch template dengan kalimat wajib.

### Application Repo

- `02-application/CLAUDE.md`
  - Update section skill menjadi **"Skills / MCP (WAJIB)"**.
  - Tambah required prompt phrase, verification commands, stop-if-missing rule.

- `02-application/.claude/rules/hermes-skills.md`
  - Update section skill menjadi **"Required Skills / MCP for Claude Code (WAJIB)"**.
  - Tambah mandatory table, verification, required prompt phrase.
  - Update dispatch command dengan kalimat wajib.

## Required Prompt Phrase

Semua dispatch ke Claude Code harus menyertakan:

> "Use superpowers, caveman, and graphify. If E2E changes are needed, also use the playwright MCP."

Atau dalam bahasa Indonesia:

> "Gunakan skill superpowers, caveman, dan graphify. Jika ada perubahan E2E, gunakan playwright MCP."

## Verification

- `scripts/audit-vexa.sh` — ✅ **AUDIT PASSED**
- GitHub root commit: `53ea1f3`
- GitHub app commit: `8d3be5b`

## Why This Matters

- Menghindari situasi di mana Claude Code tidak menggunakan toolset penuh.
- Memastikan konsistensi antara sesi coding.
- Memudahkan audit dan troubleshooting.

## Links

- Obsidian: [[vexa Workflow Runbook]]
- Obsidian: [[vexa Claude Code Setup]]
- Root: `00-meta/skills.md`
- Root: `CLAUDE.md`
- App: `02-application/CLAUDE.md`
- App: `02-application/.claude/rules/hermes-skills.md`
- Script: `scripts/audit-vexa.sh`
