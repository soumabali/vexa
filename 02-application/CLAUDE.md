# vexa — Complete SSH Manager / Claude Code Context

## Project Identity

- **Name:** vexa
- **Tagline:** Complete SSH Manager
- **Type:** Open-source, self-hosted monorepo (web + API + desktop + mobile + shared packages)
- **Project location:** `/home/ubuntu/projects/vexa/02-application/`
- **Root orchestrator:** `../CLAUDE.md`
- **Git structure:** `02-application/` is a git subtree of the root repo (no `.git` directory inside).  
  Commit/push are handled by Ame from the root workspace.
- **Obsidian SSOT:** `infra/vexa — Project Index.md`

## Role Split

- **Ame (Hermes):** plans, dispatches, verifies, commits from root, pushes root + subtree, writes session notes.
- **Claude Code:** executes code **inside `02-application/`** only.

When starting work in this directory, load:

- `02-application/CLAUDE.md` (this file)
- `02-application/.claude/settings.json`
- `02-application/.claude/rules/hermes-skills.md`

## Push Rules

- Do not create a `.git` directory inside `02-application/`.
- Do not push from inside `02-application/` (it has no git context).
- Ame commits all changes from root and runs:
  - `git push origin main` for root orchestrator
  - `git subtree push --prefix=02-application https://github.com/soumabali/vexa.git main` for application repo

## Architecture

```
apps/
  api/          Go backend (Gin, JWT, WebAuthn, audit logging)
  web/          Next.js 16 + React 19 + TypeScript
  desktop/      Tauri v2 + Rust (roadmap / experimental)
  mobile/       Flutter (roadmap / experimental)
packages/
  ssh-core/     Rust SSH/SFTP/tunnel library + FFI (roadmap)
  ui/           Shared UI components
  types/        Shared TypeScript types
  config/       Shared eslint/tailwind/tsconfig
tests/          Playwright E2E + integration tests
docs/           Documentation
docker-compose.yml
docker-compose.prod.yml
Makefile
```

## Ports & URLs (development)

| Port | Service | Notes |
|------|---------|-------|
| 3000 | web (Next.js dev) | http://localhost:3000 |
| 8080 | api (Go/Gin dev) | http://localhost:8080 |
| 5432 | PostgreSQL | localhost |
| 6379 | Redis | localhost |

## Important Commands

```bash
# Run the dev stack
make dev

# Backend only
cd apps/api && go run cmd/server/main.go

# Frontend only
cd apps/web && npm run dev

# Verification gates
make verify
```

## Verification Gates

Before declaring a task complete:

1. `cd apps/api && go test ./...`
2. `cd apps/api && go build ./...`
3. `cd apps/web && npm run build`
4. `cp -r apps/web/.next/static apps/web/.next/standalone/.next/static`
5. E2E production 19/19 passing (run by Ame/Hermes)

## Standards

### Security (P1)

- **DO NOT** hardcode secrets/credentials in source code.
- **DO NOT** set `AllowedOrigins: []string{"*"}` in production.
- **DO NOT** store secrets in `/tmp`.
- Always use env vars or Docker secrets.
- All user input must be validated.
- Audit logs must be written for every critical action.

### Code Style

- **Go:** `gofmt`, standard Go conventions, layered `internal/` architecture.
- **TypeScript/React:** ESLint + Prettier, Radix UI + Tailwind, vitest for unit tests.
- **Rust:** `cargo fmt`, `cargo clippy`.
- **Flutter:** `dart format`.

### Git Workflow

- All changes through a new branch: `feature/<name>`.
- Commit messages in English, conventional commits.
- Before merge: test + lint + review.
- **Do not** commit `.env`, binaries, or `node_modules`.

## Skills / MCP (WAJIB)

Claude Code **wajib** memuat skill/MCP ini **pada setiap sesi coding** di `02-application/`.
Ini adalah stack standard project vexa dan **bukan optional**.

| Skill/MCP | Location | Purpose | When to use |
|-----------|----------|---------|-------------|
| `superpowers` | `~/.claude/skills/superpowers/` | Coding superpowers | Every coding session |
| `caveman` | `~/.claude/skills/caveman/` | Caveman hooks and workflows | Every coding session |
| `graphify` | `~/.claude/skills/graphify/` | Codebase graph understanding | Exploration, refactoring, architecture audit |
| `playwright` | MCP `@executeautomation/playwright-mcp-server` | Browser automation | E2E tests or UI verification |

### Required Prompt Phrase

When Ame dispatches a task to Claude Code, the prompt must include:

> "Use superpowers, caveman, and graphify. If E2E changes are needed, also use the playwright MCP."

### Verification

```bash
ls -d ~/.claude/skills/{superpowers,caveman,graphify}
grep -A3 '"playwright"' ~/.claude/settings.json
```

If any skill/MCP is missing, stop and report to Ame before coding.

## Interaction Rules

- All coding activity in this project is done via Claude Code (`ollama launch claude`).
- Use `--dangerously-skip-permissions` only for approved automated workflows.
- Limit max-turns to task complexity (default 10–20).
- When finished, report a concise summary to Hermes/Ame.

## Documentation

- `README.md` — open-source contributor overview.
- `docs/dev/getting-started.md` — development setup.
- `docs/devops/docker-guide.md` — Docker operations.
- `docs/devops/self-hosted-deployment.md` — self-hosted deployment.
- `docs/devops/wireguard-deployment.md` — WireGuard tunnel setup.
- `docs/architecture/` — architecture and security.
- `docs/api/` — API contract.
- Root orchestrator: `../CLAUDE.md`.
