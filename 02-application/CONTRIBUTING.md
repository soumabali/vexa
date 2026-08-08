# Contributing to vexa

Thanks for your interest in contributing! This guide covers everything you need to get started.

## Code of Conduct

Be respectful, constructive, and security-conscious. We're building infrastructure tools — safety and reliability come first.

## Getting Started

### Prerequisites

- **Go 1.25+** — backend API
- **Node.js 22+** — frontend web app
- **Docker** — for PostgreSQL, Redis, and production builds
- **pre-commit** — required for all contributors (installed via setup script)

### Setup

```bash
# 1. Clone and fork
git clone https://github.com/soumabali/vexa.git
cd vexa

# 2. Run the automated setup (installs pre-commit hooks, checks tools)
./scripts/setup-hooks.sh

# 3. Start development services
docker compose up -d postgres redis

# 4. Run the API (terminal 1)
cd apps/api
cp .env.example .env  # edit with your local values
go run cmd/server/main.go

# 5. Run the web app (terminal 2)
cd apps/web
npm install
npm run dev
```

Open http://localhost:3000 and register the first admin account.

For detailed setup, see [`docs/dev/getting-started.md`](docs/dev/getting-started.md).

## Development Workflow

### Branching

```bash
# Feature work
git checkout -b feature/short-description

# Bug fixes
git checkout -b fix/issue-number-short-description

# Docs
git checkout -b docs/what-changed
```

### Pre-commit Hooks

All contributors must install pre-commit hooks. These run automatically on every commit:

| Hook | Checks |
|------|--------|
| `trufflehog` | Secret detection |
| `semgrep` | SAST (static analysis) |
| `eslint-plugin-security` | JS/TS dangerous patterns |
| `golangci-lint` | Go code quality + security |
| `gitleaks` | Hardcoded secrets |

```bash
# Install hooks (run once)
pre-commit install

# Run manually against all files
pre-commit run --all-files
```

### Verification Gates

Before opening a PR, make sure all gates pass:

```bash
# Run from repo root
make verify
```

This runs:
- `go build ./...` — backend compiles
- `go test ./...` — all Go tests
- `npm run typecheck` — frontend type checks
- `npm test` — all frontend tests (Vitest)
- `npm run lint` — ESLint

## Pull Request Process

1. **Create a branch** from `main` using the naming convention above.
2. **Make your changes** — follow TDD, keep commits small and focused.
3. **Verify locally** — `make verify` must pass.
4. **Open a PR** against `main` with:
   - Clear description of what and why
   - Test notes (what you tested, how)
   - Screenshots for UI changes
   - Link to related issues
5. **CI must pass** — GitHub Actions runs the full verification suite.
6. **Review** — at least one maintainer reviews. Address feedback promptly.
7. **Merge** — squash-merge to keep history clean.

## Security

- **Never** commit credentials, API keys, or secrets.
- If you accidentally commit a secret, rotate it immediately and notify maintainers.
- For vulnerability reports, see [`SECURITY.md`](SECURITY.md) — do NOT open a public issue.
- Follow the security-focused workflow in [`docs/dev/getting-started.md`](docs/dev/getting-started.md).

## Code Style

### Go

- Follow standard Go conventions (`gofmt`, `go vet`)
- `golangci-lint` configuration in `apps/api/.golangci.yml`
- Use `context.Context` for request-scoped values
- Error handling: never ignore errors, wrap with context

### TypeScript / React

- ESLint + Prettier (see `packages/config/eslint`)
- Use functional components with hooks
- Type everything — no `any` without justification
- Material Design 3 tokens via `@/components/ui`

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add credential sharing for teams
fix: resolve session revoke 404 on DELETE method
docs: update README with wireguard stats info
chore: bump dependencies
security: patch XSS vector in host name display
test: add WebAuthn credential rename test
```

## Testing

We practice TDD (test-driven development) for all new features.

- **Frontend:** Vitest + React Testing Library (`apps/web/src/**/__tests__/`)
- **Backend:** Go standard library `testing` + `testify` (`apps/api/tests/`)
- **E2E:** Playwright (`tests/e2e/playwright/`)

Run specific tests:
```bash
# Frontend single file
cd apps/web && npx vitest run src/app/settings/__tests__/sessions.test.tsx

# Backend single package
cd apps/api && go test ./internal/vault/ -v -run TestShare
```

## Project Structure

```
02-application/
├── apps/
│   ├── api/          # Go backend (Gin, PostgreSQL, Redis)
│   ├── web/          # Next.js frontend (React 19, TypeScript)
│   ├── desktop/      # Tauri v2 + Rust (roadmap)
│   └── mobile/       # Flutter (roadmap)
├── packages/
│   ├── ui/           # Shared React components (Material Design 3)
│   ├── types/        # Shared TypeScript types
│   └── config/       # ESLint, Tailwind, tsconfig presets
├── docs/             # Architecture, dev, security, DevOps docs
├── scripts/          # Operational + CI helper scripts
└── tests/            # E2E + integration tests
```

## Getting Help

- **Questions:** [GitHub Discussions](https://github.com/soumabali/vexa/discussions)
- **Bugs:** [GitHub Issues](https://github.com/soumabali/vexa/issues)
- **Security:** Email `sudharmika@gmail.com` — see [`SECURITY.md`](SECURITY.md)

---

*Happy contributing! 🚀*
