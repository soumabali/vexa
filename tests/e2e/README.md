# End-to-End + Integration Test Suite

End-to-end (Playwright) and integration (Go + Vitest) tests for the
terminal/SSH stack. The mock SSH server fixture lives at
`../integration/fixtures/ssh-server.ts` and is shared between suites.

## Layout

| Path                              | Tech       | Purpose                                                       |
| --------------------------------- | ---------- | ------------------------------------------------------------- |
| `specs/*.spec.ts`                 | Playwright | Browser-driven E2E for the desktop Next.js frontend          |
| `tests/integration/`              | Vitest     | WebSocket-level tests between Next.js and Go API              |
| `tests/integration/fixtures/`     | TypeScript | Reusable backend orchestrator + Docker sshd fixture           |
| `apps/api/tests/`                 | Go `go test` | Real SSH command execution against ephemeral sshd container |

## Prerequisites

- **Node.js** ≥ 18
- **Go** ≥ 1.22
- **Docker** (for the mock SSH server)
- A running `apps/web` dev server (default `:3000`) if running Playwright

## Quick Start

```bash
# 1. Bring up the mock SSH server + API + Postgres + Redis
cd tests/integration
docker compose -f docker-compose.dev.yml up -d

# 2. Run integration tests (Vitest, Node-side)
npm test

# 3. Run real SSH execution tests (Go)
cd ../../apps/api
go test -run TestSSHRealExecution -v ./tests/

# 4. Run E2E browser tests (Playwright)
cd ../../tests/e2e
npx playwright install --with-deps chromium
npx playwright test specs/terminal.spec.ts
```

## Skipping expensive paths

```bash
# Skip Docker-requiring tests in Go
cd apps/api && go test -short ./tests/...

# Skip Playwright browser install
PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1 npm test
```

## Mock SSH Server

`tests/integration/fixtures/ssh-server.ts` boots an `alpine:3.19`
container with `openssh-server` configured for password auth
(`testuser:testpass123`). The container exposes TCP/22 on a random host
port via `-p 0:22`. Host key is generated on the host and mounted read-only.

The fixture has two operating modes:

1. **Primary** — custom entrypoint script that runs `apk add openssh`
   inside the container (works with any `alpine:*` base).
2. **Fallback** — `sickp/alpine-sshd:7.5-r2` pre-built image (used when
   `apk` isn't available).

The Go integration test re-implements the same recipe in pure
`os/exec` so it does not depend on the TypeScript fixture
(`apps/api/tests/ssh_real_execution_test.go`).

## Coverage map

| Scenario                              | File                                                 |
| ------------------------------------- | ---------------------------------------------------- |
| Browser smoke (render, status, tabs)  | `specs/terminal.spec.ts`                             |
| Browser command input → output        | `specs/terminal.spec.ts` (last test)                 |
| WebSocket protocol (auth, frames)     | `tests/integration/terminal-ws.spec.ts`              |
| Real SSH `echo` + output verify       | `apps/api/tests/ssh_real_execution_test.go`          |
| Multiple commands (`pwd`, `whoami`)   | `…ssh_real_execution_test.go::TestSSHRealExecution_MultipleCommands` |
| stderr → stdout merge                 | `…ssh_real_execution_test.go::TestSSHRealExecution_StderrRedirect` |
| Disconnect / reconnect cycles         | `…ssh_real_execution_test.go::TestSSHRealExecution_DisconnectReconnect` |
| Non-zero exit propagation             | `…ssh_real_execution_test.go::TestSSHRealExecution_NonZeroExit` |

## Troubleshooting

| Symptom                              | Fix                                                       |
| ------------------------------------ | --------------------------------------------------------- |
| `docker port` returns multiple lines | Some Docker daemons list both IPv4 + IPv6 bindings; take first line. |
| `ssh-keygen not found`               | Install `openssh-client` on host (used to generate host key). |
| Connection refused right after start | Cold `apk add openssh` may take 60+ s; the fixture waits 180 s. |
| `Cannot connect to API :18080`       | Start `tests/integration/docker-compose.dev.yml`.         |
| Playwright "browser not found"       | `npx playwright install --with-deps chromium`.            |

## Extending the SSH fixture

When adding new test cases:

- Use `InsecureIgnoreHostKey` only in tests — never in production paths.
- Always `defer server.Stop()` to remove the container even on panic.
- Add `t.Skipf` branches when the environment cannot satisfy the
  precondition (no Docker, no ssh-keygen) so the suite stays green on
  minimal hosts.
