# Plan: Fix Rust Action Name in GitHub Actions Workflows

## Goal
Replace the non-existent `dtolnay/rust-action` action with the correct `dtolnay/rust-toolchain` action in all vexa GitHub Actions workflows.

## Background
GitHub Actions runners fail with:
```
Unable to resolve action dtolnay/rust-action, repository not found
```
The correct action for installing a Rust toolchain is `dtolnay/rust-toolchain@stable`.

## Scope
Inside `/home/ubuntu/projects/vexa/02-application/.github/workflows/` only.

## Files to Modify
1. `.github/workflows/ci.yml` — 2 occurrences
2. `.github/workflows/pr-checks.yml` — 2 occurrences
3. `.github/workflows/sast.yml` — 1 occurrence
4. `.github/workflows/sca.yml` — 1 occurrence
5. `.github/workflows/security-audit.yml` — 2 occurrences

## Change Detail
Replace:
```yaml
uses: dtolnay/rust-action@stable
```
with:
```yaml
uses: dtolnay/rust-toolchain@stable
```

Keep any `with: toolchain: ...` blocks unchanged.

## Verification Gates (run locally)
1. `cd /home/ubuntu/projects/vexa/02-application/apps/api && go test ./...`
2. `cd /home/ubuntu/projects/vexa/02-application/apps/api && go build ./...`
3. `cd /home/ubuntu/projects/vexa/02-application/apps/web && npm run build`
4. `cp -r /home/ubuntu/projects/vexa/02-application/apps/web/.next/static /home/ubuntu/projects/vexa/02-application/apps/web/.next/standalone/.next/static`

## Compliance
- Use superpowers, caveman, and graphify.
- After execution, respond ONLY with valid JSON according to `/home/ubuntu/projects/vexa/00-meta/claude-response-schema.json`.
- Only edit inside `02-application/`.
- Do not commit or push; Ame will handle root + subtree commits/push.
