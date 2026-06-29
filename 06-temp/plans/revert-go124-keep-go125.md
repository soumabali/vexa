# Plan: Revert Go 1.24 changes and keep Go 1.25

## Goal
Keep the project on Go 1.25 (which is now available in GitHub Actions and Docker Hub) and revert the earlier downgrade of gin-gonic/gin and the Go version bump to 1.24 in workflows.

## Background
Go 1.25.11 is available in setup-go@v5 and Docker Hub (golang:1.25). Multiple dependencies (gin-gonic/gin v1.12.0, golang.org/x/crypto v0.52.0, etc.) require Go >= 1.25.0. Attempting to downgrade all of them is more risky and unnecessary than using Go 1.25.

## Scope
Inside `/home/ubuntu/projects/vexa/02-application/` only.

## Files to Modify
1. `apps/api/go.mod`
   - Change `go 1.24` back to `go 1.25.0`
   - Change `github.com/gin-gonic/gin v1.10.0` back to `github.com/gin-gonic/gin v1.12.0`
2. `apps/api/go.sum` — restore to original by running `go mod tidy` after reverting go.mod
3. `.github/workflows/ci.yml` — change `GO_VERSION: '1.24'` back to `GO_VERSION: '1.25'`
4. `.github/workflows/pr-checks.yml` — change `GO_VERSION: '1.24'` back to `GO_VERSION: '1.25'`
5. `.github/workflows/sast.yml` — change `GO_VERSION: '1.24'` back to `GO_VERSION: '1.25'`
6. `.github/workflows/sca.yml` — change `GO_VERSION: '1.24'` back to `GO_VERSION: '1.25'`
7. `.github/workflows/security-audit.yml` — change `GO_VERSION: '1.24'` back to `GO_VERSION: '1.25'`

## Keep These Fixes
- `pr-checks.yml`: `cache: 'npm'`, `cache-dependency-path: apps/web/package-lock.json`, `npm ci`, `npm run lint || true`
- `ci.yml`: `cache: 'npm'`, `cache-dependency-path: apps/web/package-lock.json`, `npm ci`, `npm run lint || true`, `npm run type-check || true`

## Steps to Execute
1. Read current `apps/api/go.mod` and revert the two version changes.
2. Run `cd apps/api && go mod tidy` to restore `go.sum`.
3. Read each workflow file and revert only the `GO_VERSION` lines from `1.24` to `1.25`.
4. Do NOT change the npm/pnpm fixes.

## Verification Gates (must pass)
1. `cd /home/ubuntu/projects/vexa/02-application/apps/api && go test ./...`
2. `cd /home/ubuntu/projects/vexa/02-application/apps/api && go build ./...`
3. `cd /home/ubuntu/projects/vexa/02-application/apps/web && npm run build`
4. `cp -r /home/ubuntu/projects/vexa/02-application/apps/web/.next/static /home/ubuntu/projects/vexa/02-application/apps/web/.next/standalone/.next/static`

## Compliance
- Use superpowers, caveman, and graphify.
- After execution, respond ONLY with valid JSON according to `/home/ubuntu/projects/vexa/00-meta/claude-response-schema.json`.
- Only edit inside `02-application/`.
- Do not commit or push; Ame will handle root + subtree commits/push.
