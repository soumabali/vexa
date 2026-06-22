# 2026-06-22 — Cleanup App Context, Docs, CI/CD for Open Source

## Goal

Rapikan `02-application/` sebagai repo open-source self-hosted yang jelas, tidak ambigu, dan tidak mengandung dokumentasi/CI/scripts K8s/Terraform/Helm legacy.

## Changes

### Root Orchestrator

- Buat `00-meta/urls.md`, `00-meta/credentials-index.md`, `00-meta/skills.md` sebagai SSOT pointer.
- Update `01-documents/roadmap.md` menjadi pointer ke Obsidian Vault.
- Buat `06-temp/plans/README.md` sebagai index plan.
- Update `03-history/sessions/README.md` sebagai index session.
- Rewrite root `CLAUDE.md` dan `.claude/rules/hermes-skills.md` agar konsisten dengan SSOT.

### Obsidian Vault SSOT

- Buat `infra/vexa Workflow Runbook.md` — step-by-step workflow Ame → Claude Code.
- Update `infra/vexa Rules.md` — fix penomoran section, tambah link ke runbook.
- Update `infra/vexa — Project Index.md` — tambah link runbook.

### Application Repo (via Claude Code)

- Rewrite `02-application/CLAUDE.md` — up-to-date app context, skill mapping, verification gates.
- Create `02-application/.claude/settings.json` dan `.claude/rules/hermes-skills.md`.
- Update `02-application/README.md` — open-source contributor version.
- Update `apps/desktop/README.md` dan `apps/mobile/README.md` — roadmap/experimental.
- Archive K8s/Terraform/Helm/sprint/QA/research docs ke `docs/_archive/`.
- Replace `docs/devops/deployment.md` dengan `docs/devops/self-hosted-deployment.md`.
- Remove K8s/Terraform/CD workflows (keep CI, PR checks, security-audit, SAST, SCA).
- Remove K8s/staging/production/helm scripts.
- Rewrite `scripts/rollback.sh` untuk Docker Compose self-hosted.

## Verification

- `go test ./...` in `apps/api` — ✅ passing
- `go build ./...` in `apps/api` — ✅ exit 0
- `npm run build` in `apps/web` — ✅ passing
- `cp -r apps/web/.next/static apps/web/.next/standalone/.next/static` — ✅ done
- E2E production 19/19 passing — ✅

## Commits

- Application: `eafa67d` on `https://github.com/soumabali/vexa`
- Root: `01170b3` on `https://github.com/soumabali/vexa-root`

## Notes

- `.env`, `next-env.d.ts`, dan `tests/e2e/playwright/.auth/user.json` tidak di-commit.
- GitHub push via `git subtree push` setelah setup `~/.netrc` dengan token dari `gh` CLI.
- **Lesson learned:** `02-application/` adalah git subtree, bukan repo terpisah. Ame harusnya sudah mendokumentasikan struktur ini sejak awal. Dokumentasi sekarang sudah diperbarui di `00-meta/git-structure.md`, `README.md`, `02-application/CLAUDE.md`, dan Obsidian Vault.

## Links

- Obsidian: [[vexa Workflow Runbook]]
- Obsidian: [[vexa Rules]]
- Obsidian: [[vexa Progress]]
- Plan: `06-temp/plans/20260622-cleanup-app-context-docs.md`
