# SSH Manager — Integration Test Suite

End-to-end integration tests for the **WebSocket Terminal** connecting the desktop frontend (Next.js + xterm.js) to the Go backend (Gin + Gorilla WebSocket).

## Prerequisites

- **Node.js** ≥ 18
- **Docker + Docker Compose** (for PostgreSQL, Redis, test SSH server)
- **Go** ≥ 1.22 (for running the API server natively)
- **OpenSSL** (for JWT signing in tests)

## Stack Under Test

| Component | Tech | Port (test) |
|-----------|------|-------------|
| API Server | Go / Gin / Gorilla WS | `18080` |
| Database | PostgreSQL 16 | `15432` |
| Cache | Redis 7 | `16379` |
| Test SSH Server | Docker (alpine-sshd) | dynamic |

## File Layout

```
tests/integration/
├── terminal-ws.spec.ts              # Main test suite
├── fixtures/
│   ├── backend-orchestrator.ts      # Start/stop Go + DB + Redis
│   └── ssh-server.ts                # ephemeral SSH container fixture
├── docker-compose.dev.yml           # Test infrastructure services
├── vitest.config.ts                 # Test runner config
└── README.md                        # This file
```

## How It Works

1. **Infrastructure** (`backend-orchestrator.ts`)
   - Starts PostgreSQL + Redis via `docker-compose.dev.yml`
   - Starts the Go API server **natively** (`go run cmd/server/main.go`) on port `18080`
   - Waits for `/health` to return `200`

2. **Auth**
   - Tests generate HS256 JWTs signed with the same secret the backend uses
   - Token is passed via `?token=<jwt>` query parameter (or `Authorization` header)

3. **Protocol**
   - The Go backend uses **JSON text messages** (not the binary constants in `websocket.ts`):
     ```json
     { "type": "data",    "data": "echo hello\\r" }
     { "type": "resize",  "data": { "cols": 80, "rows": 24 } }
     { "type": "heartbeat" }
     { "type": "error",   "data": "..." }
     ```

4. **SSH Server**
   - `ssh-server.ts` spins up a Docker container with `sshd`
   - Used for full I/O tests when wired into a host record
   - Falls back through multiple Docker images (alpine → sickp/alpine-sshd → linuxserver)

## Running the Tests

```bash
cd tests/integration

# Install dependencies
npm install

# Run all tests (headless)
npm run test

# Run with UI
npm run test:ui

# Debug verbose output
npm run test:debug
```

## Environment Variables

Set in `docker-compose.dev.yml` and mirrored in `backend-orchestrator.ts`:

| Variable | Default (test) |
|----------|----------------|
| `DATABASE_URL` | `postgres://vexa:testpassword@localhost:15432/vexa_test?sslmode=disable` |
| `REDIS_ADDR` | `localhost:16379` |
| `JWT_SECRET` | `test-jwt-secret-minimum-32-bytes-long!!` |
| `JWT_REFRESH_SECRET` | `test-refresh-secret-32-bytes-long!!!` |
| `SERVER_PORT` | `18080` |
| `ALLOWED_ORIGINS` | `*` |

## Test Suites

| Suite | Coverage |
|-------|----------|
| **Backend Startup** | Health check, infra readiness |
| **Auth Handshake** | Missing token, expired JWT, invalid JWT, valid token via query param & header |
| **Session Lifecycle** | Open, input/output, resize, close, reconnect, heartbeat, rapid connect/disconnect |
| **Concurrent Sessions** | 3+ parallel WebSockets, data isolation |
| **Auth Failures** | Empty token, malformed JWT, missing claims, REST 401 |
| **Edge Cases** | Invalid host ID, missing host_id, oversized input (4096 limit), large output |

## Known Limitations

- **SSH I/O tests** verify the WebSocket stays open and does not crash. Real shell output requires the test SSH server to be registered as a Host in the database and the backend to dial it.
- **Binary protocol** (`WS_MSG_INPUT`, `WS_MSG_OUTPUT`, etc.) is defined in the frontend library (`websocket.ts`) but the Go backend currently uses JSON text frames. Tests target the actual backend protocol.
- The `ws` library (`node_modules/ws`) is used for Node.js WebSocket clients; headers may be stripped on some connections.

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| `PostgreSQL failed to start` | Check Docker is running; `docker compose -f docker-compose.dev.yml -p ssh-manager-test up -d postgres` |
| `API server failed to start` | Ensure Go is installed and `../../apps/api/go.mod` exists; check port `18080` is free |
| `WebSocket open timeout` | API may not be ready; check `[API]` logs in test output |
| `JWT validation failed` | Verify `JWT_SECRET` matches between orchestrator and backend |
| `SSH server start failed` | Docker may not support port mapping; check `docker port` output |

## Next Steps

- Wire test data seeding to create a Host record pointing to the ephemeral SSH server for full end-to-end shell I/O assertions.
- Add Playwright tests for the actual desktop frontend (Tauri WebView) once the app is buildable in CI.
