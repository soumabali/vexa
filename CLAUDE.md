# vexa — Orchestrator Claude Context

> Applies to root workspace `/home/ubuntu/projects/vexa`.  
> This file is for orchestration context only. For SSOT, see Obsidian Vault `infra/vexa — Project Index.md`.

---

## Role

- **Ame (Hermes):** plans, dispatches, verifies, commits, pushes, writes session notes, maintains root context files.
- **Claude Code:** executes code **only** inside `02-application/`.

---

## Rules

1. **Never write production code in root.** Root is for orchestration context and plans only.
2. **All coding tasks must use `02-application/CLAUDE.md`** as the project context.
3. **All plans live in `06-temp/plans/`.**
4. **After work, write session note** to `03-history/sessions/YYYY-MM-DD-<topic>.md` and mirror to Obsidian Vault.
5. **All credentials** live only in Obsidian Vault (`credentials/`); use `[REDACTED]` in chat/repo.
6. **Load required Hermes skills** before technical work:
   - `claude-code`
   - `production-e2e-regression`
   - `open-source-project-cleanup`
   - `requesting-code-review`
   - `plan`
   - `graphify`
   - `project-directory-structure`
   - `hermes-agent`

---

## Workflow

```text
Plan → Dispatch → Verify → Review → Commit/Push → Document
```

| Step | Siapa | Output |
|------|-------|--------|
| 1. Plan | Ame | `06-temp/plans/YYYYMMDD-<slug>.md` |
| 2. Dispatch | Ame | `ollama launch claude ...` ke `02-application/` |
| 3. Verify | Ame | go test, go build, npm build, E2E hijau |
| 4. Review | Claude Code | Independent review untuk perubahan non-trivial |
| 5. Commit/Push | Ame | Commit ke `02-application/`, push ke `soumabali/vexa` |
| 6. Document | Ame | Session note + Obsidian SSOT update |

---

## Permission Mode

- Workflow otomatis (print mode) menggunakan `ollama launch claude --model kimi-k2.7-code:cloud -- ... --dangerously-skip-permissions`.
- `--dangerously-skip-permissions` mempercepat eksekusi tetapi **tidak menghilangkan verification gates**.
- Pekerjaan interaktif berisiko tinggi boleh pakai plan/ask mode.

---

## Claude Code Skills/MCP (WAJIB)

Claude Code **wajib** memuat skill/MCP ini **pada setiap sesi coding** di `02-application/`.
Ini adalah stack standard project vexa dan **bukan optional**.

| Skill/MCP | Lokasi | Fungsi | Kapan digunakan |
|-----------|--------|--------|-----------------|
| `superpowers` | `~/.claude/skills/superpowers/` | Coding superpowers | Setiap sesi coding |
| `caveman` | `~/.claude/skills/caveman/` | Caveman hooks dan workflows | Setiap sesi coding |
| `graphify` | `~/.claude/skills/graphify/` | Codebase graph understanding | Saat eksplorasi, refactoring, audit arsitektur |
| `playwright` | MCP `@executeautomation/playwright-mcp-server` | Browser automation | Saat membuat/mengubah test E2E atau verifikasi UI |

**Prompt dispatch harus menyertakan:**

> "Gunakan skill superpowers, caveman, dan graphify. Jika ada perubahan E2E, gunakan playwright MCP."

Ame mengecek ketersediaan via `scripts/audit-vexa.sh` sebelum dispatch.

---

## Verification Gates

Before declaring a task complete:

1. `cd 02-application/apps/api && go test ./...`
2. `cd 02-application/apps/api && go build ./...`
3. `cd 02-application/apps/web && npm run build`
4. Copy `02-application/apps/web/.next/static` to `02-application/apps/web/.next/standalone/.next/static`
5. E2E production against `https://vexa.nexigo.my.id`

---

## Subtree Maintenance

```bash
# Pull latest application into root
cd /home/ubuntu/projects/vexa
git subtree pull --prefix=02-application \
  https://github.com/soumabali/vexa.git main --squash

# Push local application changes to GitHub
cd /home/ubuntu/projects/vexa
git subtree push --prefix=02-application \
  https://github.com/soumabali/vexa.git main
```

---

## Session Note Template

```markdown
# YYYY-MM-DD — Topic

## Goal
...

## Changes
- ...

## Verification
- go test: ...
- go build: ...
- npm build: ...
- E2E: .../19 passing

## Notes
- ...

## Links
- Obsidian: [[vexa Progress]]
```

See `03-history/sessions/` for examples.

---

## Obsidian SSOT

- [[vexa — Project Index]]
- [[vexa Progress]]
- [[vexa Runbook]]
- [[vexa Roadmap]]
- [[vexa Rules]]
- [[vexa Claude Code Setup]]
- [[vexa Workflow Runbook]]
