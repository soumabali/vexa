# Plan: Fix npm ci in GitHub Actions CI

## Goal
Make `npm ci` succeed in GitHub Actions by enabling legacy peer dependency resolution, matching the local dev environment where `npm install` already works.

## Background
GitHub Actions runners fail with:
```
npm error ERESOLVE could not resolve
npm error While resolving: react-simple-maps@4.0.0-beta.6
npm error Found: react@19.2.7
npm error Could not resolve dependency:
npm error peer react@"^16.8.0 || 17.x || 18.x" from react-simple-maps@4.0.0-beta.6
```
Local `npm ci --legacy-peer-deps` succeeds.

## Scope
Inside `/home/ubuntu/projects/vexa/02-application/` only.

## Files to Modify
1. `apps/web/.npmrc`
   - Create file with content:
   ```
   legacy-peer-deps=true
   ```
2. `.github/workflows/ci.yml`
   - Change `npm ci` to `npm ci --legacy-peer-deps` (line in Lint TypeScript step)
   - Change `npm run lint || true` and `npm run type-check || true` remain unchanged.
3. `.github/workflows/pr-checks.yml`
   - Change `npm ci` to `npm ci --legacy-peer-deps` (line in TS lint step)
   - Change `npm run lint || true` remain unchanged.

## Steps
1. Create `apps/web/.npmrc` with `legacy-peer-deps=true`.
2. Edit `ci.yml` and `pr-checks.yml` to use `npm ci --legacy-peer-deps`.
3. Verify local `npm ci --legacy-peer-deps` works (already confirmed).

## Verification Gates
1. `cd /home/ubuntu/projects/vexa/02-application/apps/api && go test ./...`
2. `cd /home/ubuntu/projects/vexa/02-application/apps/api && go build ./...`
3. `cd /home/ubuntu/projects/vexa/02-application/apps/web && npm run build`
4. `cp -r /home/ubuntu/projects/vexa/02-application/apps/web/.next/static /home/ubuntu/projects/vexa/02-application/apps/web/.next/standalone/.next/static`

## Compliance
- Use superpowers, caveman, and graphify.
- After execution, respond ONLY with valid JSON according to `/home/ubuntu/projects/vexa/00-meta/claude-response-schema.json`.
- Only edit inside `02-application/`.
- Do not commit or push; Ame will handle root + subtree commits/push.
