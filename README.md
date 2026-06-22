# vexa — Orchestrator Root

> Private orchestration workspace for Hermes Agent and Claude Code.
> Application code lives in `02-application/` (subtree → https://github.com/soumabali/vexa).

## Directory Map

| Folder | Purpose |
|--------|---------|
| `00-meta/` | SSOT links to credentials, URLs, ports, contacts, git structure |
| `01-documents/` | Architecture, roadmap, decisions, standards |
| `02-application/` | Production application code (git subtree) |
| `03-history/sessions/` | Per-session notes and reports |
| `04-data/` | Reference data, exports |
| `05-config/` | Local config and secrets (gitignored) |
| `06-temp/plans/` | Active Claude Code execution plans |
| `.claude/` | Root-level Claude Code context |

## Git Structure

- `02-application/` is a **git subtree**, not a separate local repo.
- It does **not** contain a `.git` directory.
- All commits and pushes are done from this root workspace.
- Push application subtree: `git subtree push --prefix=02-application https://github.com/soumabali/vexa.git main`
- Detail: see `00-meta/git-structure.md`.

## Workflow

1. Ame writes plans in `06-temp/plans/` and session notes in `03-history/sessions/`.
2. Claude Code executes inside `02-application/`.
3. Ame verifies gates (go test, go build, npm build, E2E).
4. Ame commits from root and pushes both root + subtree.

## Claude Code Rule

1. Do **not** edit code in this root directory.
2. All coding work happens inside `02-application/`.
3. Load `02-application/CLAUDE.md` and `02-application/.claude/rules/hermes-skills.md` before coding.
4. Ame writes plans here; Claude Code executes inside `02-application/`.
5. Do not create a `.git` directory inside `02-application/`.

## Links

- **Public application repo:** https://github.com/soumabali/vexa
- **Private root repo:** https://github.com/soumabali/vexa-root
- **Obsidian vault:** `Documents/Obsidian Vault/infra/vexa.md`
- **Credentials:** `Documents/Obsidian Vault/credentials/vexa Credentials.md`
