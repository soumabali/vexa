# Plan: Downgrade gin to v1.10.0 and stabilize Go 1.24

## Goal
Make `apps/api` build and test successfully with Go 1.24 stable in GitHub Actions by downgrading gin-gonic/gin from v1.12.0 to v1.10.0.

## Background
gin-gonic/gin v1.12.0 declares `go 1.25.0` in its go.mod, which conflicts with the project target of Go 1.24. v1.10.0 declares `go 1.20` and is fully API-compatible with the gin usage in this project (only standard gin APIs: gin.H, c.JSON, c.Redirect, c.Request.Context, c.Header, c.Data, middleware, router).

## Scope
Inside `/home/ubuntu/projects/vexa/02-application/` only.

## Files to Modify
1. `apps/api/go.mod`
   - Change `github.com/gin-gonic/gin v1.12.0` to `github.com/gin-gonic/gin v1.10.0`
   - Keep `go 1.24` directive (already set)
2. `apps/api/go.sum` (via `go mod tidy`)

## Steps to Execute
1. Read `apps/api/go.mod` and locate `github.com/gin-gonic/gin v1.12.0`.
2. Edit it to `github.com/gin-gonic/gin v1.10.0`.
3. Run `cd apps/api && go mod tidy`.
4. Run `cd apps/api && go test ./...`.
5. Run `cd apps/api && go build ./...`.
6. If any other dependency also requires Go 1.25, report it and stop.

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
