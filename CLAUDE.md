# vexa — Orchestrator Claude Context

> Applies to root workspace `/home/ubuntu/projects/vexa`.

## Role

- **Ame (Hermes):** plans, dispatches, verifies, commits, pushes, writes session notes.
- **Claude Code:** executes code **only** inside `02-application/`.

## Rules

- Never write code in root; root is for orchestration context and plans.
- All coding tasks must use `02-application/CLAUDE.md` as the project context.
- All plans live in `06-temp/plans/`.
- After work, write session note to `03-history/sessions/YYYY-MM-DD-<topic>.md` and mirror to Obsidian Vault.

## Subtree Maintenance

```bash
# Add / update application subtree
git subtree pull --prefix=02-application https://github.com/soumabali/vexa.git main --squash

git subtree push --prefix=02-application https://github.com/soumabali/vexa.git main
```

## Verification Gates

Before declaring a task complete:

1. `cd 02-application/apps/api && go test ./...`
2. `cd 02-application/apps/api && go build ./...`
3. `cd 02-application/apps/web && npm run build`
4. Copy `.next/static` to `.next/standalone/.next/static`
5. E2E production against `https://vexa.nexigo.my.id`

## Session Note Template

See `03-history/sessions/` for examples.
