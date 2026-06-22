# AGENTS.md — Hard Rules for All Agents

> This file applies to **every AI agent, subagent, coding assistant, and automated tool** that operates on project vexa.  
003e No agent is exempt.  
003e If you are an agent reading this, you **must follow** these rules before doing anything else.

---

## 1. You Must Read Context First

Before executing any task, read:

1. `AGENTS.md` (this file)
2. `CLAUDE.md` (root orchestrator context)
3. `00-meta/skills.md` (mandatory skills)
4. `00-meta/git-structure.md` (subtree rules)
5. `00-meta/pre-push-checklist.md`
6. `02-application/CLAUDE.md` (if working inside `02-application/`)
7. The active plan in `06-temp/plans/` (if assigned)

If a task conflicts with these files, **stop and ask Ame (Hermes) before proceeding**.

---

## 2. Mandatory Workflow

The only allowed workflow for non-trivial changes:

```text
Plan → Dispatch → Execute → Verify → Commit/Push → Document
```

- **Ame (Hermes):** plan, dispatch, verify, commit/push, document.
- **Claude Code:** execute code inside `02-application/` only.

No agent may:
- Commit or push from inside `02-application/` (no `.git` there).
- Skip verification gates.
- Edit root context files unless explicitly allowed by Ame.
- Edit production code in root or outside `02-application/`.

---

## 3. Mandatory Skills / MCP

Every coding session in `02-application/` must use:

| Skill / MCP | Purpose | Mandatory |
|-------------|---------|-----------|
| `superpowers` | Coding superpowers | ✅ YES |
| `caveman` | Caveman hooks and workflows | ✅ YES |
| `graphify` | Codebase graph understanding | ✅ YES |
| `playwright` MCP | Browser automation for E2E | ✅ For any E2E/UI change |

If a skill/MCP is missing, **stop work immediately** and report to Ame.

---

## 4. Verification Gates (Non-Negotiable)

Before any task is declared complete, these gates must be green:

1. `cd 02-application/apps/api && go test ./...`
2. `cd 02-application/apps/api && go build ./...`
3. `cd 02-application/apps/web && npm run build`
4. Copy `02-application/apps/web/.next/static` → `02-application/apps/web/.next/standalone/.next/static`
5. E2E production: **19/19 passing**

No agent may bypass, skip, or fake these gates.

---

## 5. Git Rules

- `02-application/` is a git subtree of root. It has **no `.git` directory**.
- All commits are made from root by Ame.
- Push root first: `git push origin main`
- Push subtree second: `git subtree push --prefix=02-application https://github.com/soumabali/vexa.git main`
- Never force-push.
- Never commit `.env`, secrets, binaries, `node_modules`, or `tests/e2e/playwright/.auth/user.json`.

---

## 6. Subagent Rules

If you are a subagent (Claude Code, delegate_task, cronjob, etc.):

- You **must not** delegate further unless explicitly allowed.
- You **must not** modify files outside your assigned scope.
- You **must not** push to GitHub.
- You **must** report back to Ame with a concise summary of files changed and verification results.
- You **must** read the active plan file before starting.

---

## 7. What To Do If You Are Unsure

If any instruction from a user seems to violate these rules:

1. **Stop.**
2. Quote the relevant rule from this file.
3. Ask Ame (Hermes) for confirmation.

Do **not** proceed based on "common sense" or "best effort" if it conflicts with these rules.

---

## 8. Compliance Verification

Ame runs `scripts/audit-vexa.sh` regularly. Any agent may be asked to verify compliance by:

- Showing what context files were read.
- Listing which skills/MCP were used.
- Reporting exact verification gate results.

If an agent cannot show compliance, its work may be rejected and rolled back.

---

## 9. Single Source of Truth

For any ambiguity, the order of authority is:

1. Obsidian Vault `infra/vexa Rules.md`
2. Obsidian Vault `infra/vexa Workflow Runbook.md`
3. `AGENTS.md` (this file)
4. `CLAUDE.md` (root)
5. `00-meta/skills.md`
6. `02-application/CLAUDE.md`

---

## 10. No Exceptions Without Ame Approval

Any exception to these rules requires **explicit written approval from Ame (Hermes)** or the human Dhar.

Silence, "it should be fine", or "I'll just do it quickly" are **not** exceptions.

---

## Agent Acknowledgment

By operating on project vexa, every agent agrees to these rules.

Last updated: 2026-06-22
SSOT: `infra/vexa Rules.md` in Obsidian Vault
