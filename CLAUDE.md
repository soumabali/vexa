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

## Single Source of Truth Boundaries

| Layer | SSOT For | Location |
|-------|----------|----------|
| Production code | Application code | `02-application/` → `https://github.com/soumabali/vexa` |
| Workflow, plans, session notes, decisions | Orchestrator context | Root (`/home/ubuntu/projects/vexa`) → `https://github.com/soumabali/vexa-root` |
| Credentials, secrets | Private vault | `/home/ubuntu/Documents/Obsidian Vault/credentials/` |
| URLs, ports, runbook | Obsidian Vault + root `00-meta/` | `00-meta/urls.md` + Obsidian `infra/` |

**Do not create `.status/`, `00-meta/`, or `03-history/` inside `02-application/`**. Those folders belong to the root orchestrator only.

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

- Workflow otomatis (print mode) menggunakan `scripts/dispatch-claude.sh` yang memanggil `ollama launch claude ... --print`.
- Permission default untuk Claude Code: `--allowedTools "Read,Write,Edit,Bash(go test),Bash(go build),Bash(npm run build),Bash(make),Bash(cp),Bash(rm -f),Bash(mkdir),Bash(ls),Bash(grep),Bash(cd)"`.
- Untuk command Bash di luar allowlist, gunakan `scripts/safe-exec.sh` atau jalankan manual oleh Ame.
- `--dangerously-skip-permissions` atau `--allow-dangerously-skip-permissions` tidak lagi digunakan secara otomatis; setiap permission harus melalui allowlist eksplisit.
- Pekerjaan interaktif berisiko tinggi boleh pakai plan/ask mode.

---

## Claude Code Skills/MCP (WAJIB)

Claude Code **wajib** memuat skill/MCP ini **pada setiap sesi coding** di `02-application/`.
Detail lengkap ada di `00-meta/skills.md`. Ini adalah stack standard project vexa dan **bukan optional**.

**Prompt dispatch harus menyertakan:**

> "Gunakan skill superpowers, caveman, dan graphify. Jika ada perubahan E2E, gunakan playwright MCP. Setelah selesai, respond ONLY dengan valid JSON sesuai `00-meta/claude-response-schema.json`."

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
- [[vexa Error Recovery Playbook]]
