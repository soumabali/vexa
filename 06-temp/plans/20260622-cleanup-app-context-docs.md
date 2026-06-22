# 20260622 — Cleanup App Context, Docs, CI/CD for Open Source

## Goal

Rapikan `02-application/` agar siap menjadi open-source self-hosted repo yang jelas, tidak ambigu, dan tidak mengandung dokumentasi/CI/scripts yang tidak relevan.

## Scope

### In Scope
1. Rewrite `02-application/CLAUDE.md` menjadi app-level context yang up-to-date.
2. Buat `02-application/.claude/settings.json`.
3. Buat `02-application/.claude/rules/hermes-skills.md`.
4. Update `02-application/README.md` untuk open-source contributor final.
5. Cleanup dokumentasi `docs/`:
   - Hapus/rename file yang masih bicara K8s/Terraform/Helm sebagai primary deployment.
   - Arsipkan sprint notes internal ke `docs/_archive/` atau hapus jika benar-benar tidak relevan.
   - Pertahankan: `dev/getting-started.md`, `devops/docker-guide.md`, `devops/self-hosted-deployment.md` (baru), `architecture/*.md`, `api/*`.
6. Cleanup CI/CD:
   - Hapus workflow K8s/Terraform dari `.github/workflows/`.
   - Update `ci.yml` agar tidak merujuk path `infra/` yang sudah tidak ada.
   - Pertahankan `ci.yml`, `pr-checks.yml`, `security-audit.yml`, `sast.yml`, `sca.yml`.
7. Cleanup scripts:
   - Hapus `deploy-k8s.sh`, `deploy-staging.sh`, `deploy-production.sh`, `helm-install.sh`, `helm-upgrade.sh`.
   - Pertahankan script yang relevan untuk Docker/Compose/local.
8. Cleanup desktop/mobile scaffolding:
   - Hapus file meta OpenClaw (`AGENTS.md`, `BOOTSTRAP.md`, `SOUL.md`, `IDENTITY.md`, `USER.md`, `TOOLS.md`, `HEARTBEAT.md`) dari `apps/desktop/`.
   - Update `apps/desktop/README.md` menjadi roadmap.
   - Update `apps/mobile/README.md` menjadi roadmap.
9. Pastikan branding konsisten: `vexa — Complete SSH Manager`.
10. Update `Makefile` agar tidak merujuk K8s/Terraform.

### Out of Scope
- Perubahan fitur production code (`apps/api`, `apps/web`).
- Refactor source code `.go/.ts/.tsx`.
- Perubahan Docker Compose production config yang sudah berfungsi.

## Files to Modify

- `02-application/CLAUDE.md` → rewrite
- `02-application/.claude/settings.json` → create
- `02-application/.claude/rules/hermes-skills.md` → create
- `02-application/README.md` → update
- `02-application/Makefile` → update
- `02-application/docs/devops/deployment.md` → rewrite/rename
- `02-application/docs/devops/docker-guide.md` → verify/update
- `02-application/docs/pm/sprints/*.md` → archive or delete
- `02-application/docs/security/pentest/*.md` → review, archive if placeholder
- `02-application/docs/qa/*.md` → review, archive if placeholder
- `02-application/docs/research/security-research.md` → review, archive if placeholder
- `02-application/.github/workflows/ci.yml` → remove K8s/Terraform refs
- `02-application/.github/workflows/cd-staging.yml` → delete or archive
- `02-application/.github/workflows/*.yml` → review individually
- `02-application/scripts/deploy-k8s.sh` → delete
- `02-application/scripts/deploy-staging.sh` → delete
- `02-application/scripts/deploy-production.sh` → delete
- `02-application/scripts/helm-install.sh` → delete
- `02-application/scripts/helm-upgrade.sh` → delete
- `02-application/apps/desktop/AGENTS.md` → delete
- `02-application/apps/desktop/BOOTSTRAP.md` → delete
- `02-application/apps/desktop/SOUL.md` → delete
- `02-application/apps/desktop/IDENTITY.md` → delete
- `02-application/apps/desktop/USER.md` → delete
- `02-application/apps/desktop/TOOLS.md` → delete
- `02-application/apps/desktop/HEARTBEAT.md` → delete
- `02-application/apps/desktop/README.md` → update
- `02-application/apps/mobile/README.md` → update

## Verification Gates

1. `cd 02-application/apps/api && go test ./...` passing
2. `cd 02-application/apps/api && go build ./...` exit 0
3. `cd 02-application/apps/web && npm run build` passing
4. `cp -r 02-application/apps/web/.next/static 02-application/apps/web/.next/standalone/.next/static`
5. `cd 02-application/tests/e2e && SKIP_WEBSERVER=true BASE_URL=https://vexa.nexigo.my.id API_BASE_URL=https://api-vexa.nexigo.my.id npx playwright test --workers=1` → 19/19 passing

## Rollback Plan

Jika E2E gagal:
1. Cek log container.
2. `git revert HEAD` jika perlu.
3. Re-deploy dari commit sebelumnya.
4. Re-run E2E.

## Notes

- Jangan hapus file source code production.
- Jangan hapus `tests/e2e/`, `docker-compose.yml`, `docker-compose.prod.yml`, `apps/api/Dockerfile`, `apps/web/Dockerfile`.
- Semua kredensial tetap `[REDACTED]` di file mana pun.

## Links

- Obsidian: [[vexa Workflow Runbook]]
- Obsidian: [[vexa Rules]]
- Obsidian: [[vexa Progress]]
- Root plan location: `/home/ubuntu/projects/vexa/06-temp/plans/20260622-cleanup-app-context-docs.md`
