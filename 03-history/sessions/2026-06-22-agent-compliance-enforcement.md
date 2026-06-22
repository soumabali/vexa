# 2026-06-22 — Agent Compliance Enforcement

## Goal

Memastikan semua agent/subagent termasuk Claude Code **selalu** menggunakan aturan dan flow yang sudah dibuat. Sebelumnya skill dan flow terdaftar, tapi tidak ada enforcement eksplisit.

## Changes

### Root Orchestrator

- **`AGENTS.md` (baru)** — hard rules untuk semua AI agent/subagent/tool:
  - Wajib baca context files dulu.
  - Wajib ikuti Plan → Dispatch → Execute → Verify → Commit/Push → Document.
  - Wajib gunakan skill/MCP mandatory.
  - Tidak boleh skip verification gates.
  - Tidak boleh commit/push dari `02-application/`.
  - Jika ragu, stop dan tanya Ame.

- **`CLAUDE.md` root** — tambah referensi ke `AGENTS.md` dan penekanan bahwa semua agent harus mematuhinya.

- **`00-meta/skills.md`** — update dispatch template agar meminta compliance confirmation dari Claude Code sebelum eksekusi.

- **`scripts/audit-vexa.sh`** — tambah check untuk:
  - Keberadaan `AGENTS.md`
  - Compliance clauses di `AGENTS.md`, `CLAUDE.md`, `00-meta/skills.md`, `02-application/CLAUDE.md`, `02-application/.claude/rules/hermes-skills.md`

### Obsidian Vault

- **`vexa Rules.md`** — tambah section **9. Agent Compliance (New)** dengan 5 subsection:
  - Wajib baca AGENTS.md
  - Tidak boleh mengabaikan flow
  - If unsure, stop and ask
  - Agent must report compliance
  - Audit & rollback
- Renumber section Communication → 10 dan Enforcement → 11.

- **`vexa Workflow Runbook.md`** — tambah **Agent Compliance Check (WAJIB)** di checklist before dispatch, serta update checklist agar mencakup AGENTS.md dan root CLAUDE.md.

### Application Repo

- **`02-application/CLAUDE.md`** — tambah **AGENT COMPLIANCE NOTICE** di paling atas dan **Compliance Check (Do This First)** sebelum coding.

- **`02-application/.claude/rules/hermes-skills.md`** — update dispatch command dengan compliance check phrase.

## Expected First Response from Claude Code

Setiap kali di-dispatch, Claude Code harus membalas konfirmasi:

1. ✅ `AGENTS.md` dibaca
2. ✅ Root `CLAUDE.md` dibaca
3. ✅ `00-meta/skills.md` dibaca
4. ✅ `00-meta/git-structure.md` dibaca
5. ✅ `02-application/CLAUDE.md` dibaca
6. ✅ Plan dibaca
7. ✅ `superpowers`, `caveman`, `graphify` tersedia
8. ✅ `playwright` MCP tersedia (jika E2E)
9. ✅ Hanya akan edit file di `02-application/`
10. ✅ Tidak akan push/commit dari `02-application/`

## Verification

- `/home/ubuntu/projects/vexa/scripts/audit-vexa.sh` — ✅ **AUDIT PASSED**
- Root commit: `e105bb3`
- Application commit: `f194cc6`

## Why This Matters

- Menghilangkan ambiguitas siapa yang boleh apa.
- Memastikan Claude Code tidak "nyasar" mengabaikan rules.
- Audit script bisa secara otomatis mendeteksi pelanggaran.

## Links

- `AGENTS.md`
- `CLAUDE.md` root
- `00-meta/skills.md`
- `02-application/CLAUDE.md`
- `02-application/.claude/rules/hermes-skills.md`
- Obsidian: `infra/vexa Rules.md`
- Obsidian: `infra/vexa Workflow Runbook.md`
- `scripts/audit-vexa.sh`
