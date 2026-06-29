# Plan: Fix GitHub Actions CI for vexa Application Repo

## Goal
Make GitHub Actions workflows run green on push/PR for the `soumabali/vexa` application repo.

## Scope
Inside `/home/ubuntu/projects/vexa/02-application/` only.

## Files to Modify
1. `apps/api/go.mod` — change `go 1.25.0` to `go 1.24` (latest stable)
2. `.github/workflows/ci.yml` — change `GO_VERSION: '1.25'` to `GO_VERSION: '1.24'`
3. `.github/workflows/pr-checks.yml` — change `GO_VERSION: '1.25'` to `GO_VERSION: '1.24'`; replace `pnpm` with `npm`
4. `.github/workflows/sast.yml` — change `GO_VERSION: '1.25'` to `GO_VERSION: '1.24'`
5. `.github/workflows/sca.yml` — change `GO_VERSION: '1.25'` to `GO_VERSION: '1.24'`
6. `.github/workflows/security-audit.yml` — change `GO_VERSION: '1.25'` to `GO_VERSION: '1.24'`

## Detailed Changes

### apps/api/go.mod
Change:
```
go 1.25.0
```
to:
```
go 1.24
```

### ci.yml
- `GO_VERSION: '1.25'` → `GO_VERSION: '1.24'`
- Already fixed pnpm→npm in a previous commit; verify it stays fixed.

### pr-checks.yml
- `GO_VERSION: '1.25'` → `GO_VERSION: '1.24'`
- `cache: 'pnpm'` → `cache: 'npm'`
- Add `cache-dependency-path: apps/web/package-lock.json`
- `pnpm install --frozen-lockfile || pnpm install` → `npm ci`
- `pnpm lint || true` → `npm run lint || true`

### sast.yml, sca.yml, security-audit.yml
- `GO_VERSION: '1.25'` → `GO_VERSION: '1.24'`

## Verification Gates (Claude must run)
1. `cd /home/ubuntu/projects/vexa/02-application/apps/api && go test ./...`
2. `cd /home/ubuntu/projects/vexa/02-application/apps/api && go build ./...`
3. `cd /home/ubuntu/projects/vexa/02-application/apps/web && npm run build`
4. `cp -r /home/ubuntu/projects/vexa/02-application/apps/web/.next/static /home/ubuntu/projects/vexa/02-application/apps/web/.next/standalone/.next/static`

## Compliance
- Use superpowers, caveman, and graphify. If E2E changes are needed, also use the playwright MCP.
- After execution, respond ONLY with valid JSON according to `/home/ubuntu/projects/vexa/00-meta/claude-response-schema.json`.
- Only edit inside `02-application/`.
- Do not push; Ame will handle root + subtree commits/push.
