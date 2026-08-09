# vexa — Complete SSH Manager

> Self-hosted, security-first SSH access management for individuals and teams.

## What is vexa?

vexa is an open-source SSH manager that brings hosts, credentials, tunnels, and terminal access into one place. It is designed for DevOps teams, system administrators, and solo operators who want a modern web interface for managing SSH infrastructure without handing sensitive data to a third-party service.

The project is built as a monorepo: a **Next.js** web application, a **Go** API, and shared packages for UI, types, and configuration. Optional desktop (Tauri) and mobile (Flutter) clients are on the roadmap. Everything is self-hosted with Docker, using PostgreSQL for persistence and Redis for sessions and background work.

Security is treated as a first-class feature. Credentials are encrypted at rest, authentication uses Argon2-hashed passwords plus WebAuthn/passkeys, and every critical action is written to an audit log. Rate limiting, CORS enforcement, and TLS-ready deployment defaults help you run vexa safely in production.

## Quick Start (Docker)

The fastest way to try vexa is with Docker Compose:

```bash
git clone https://github.com/soumabali/vexa.git
cd vexa
cp .env.example .env
# Edit .env and set strong secrets for POSTGRES_PASSWORD, JWT_SECRET, and ENCRYPTION_KEY.
docker compose up -d
```

Open http://localhost:3000 and register the first admin account.

For production self-hosting, TLS, backups, and environment hardening, see [`docs/devops/self-hosted-deployment.md`](docs/devops/self-hosted-deployment.md).

## Screenshots

<!-- Coming in v1.1 -->

## Features

| Feature | Status | Notes |
|---------|--------|-------|
| Host management | Ready | SSH hosts, labels, inventory, activity log |
| Credential vault | Ready | AES-GCM encrypted master vault, team sharing |
| SSH terminal | Ready | xterm.js via WebSocket, host selector, multi-tab |
| SFTP file manager | Ready | Upload, download, navigate remote files |
| WireGuard tunnels | Ready | Per-user tunnels, enable/disable/rotate keys, live stats |
| WebAuthn / passkeys | Ready | FIDO2 security-key support, rename/delete |
| TOTP MFA | Ready | QR setup, backup codes, regenerate support |
| Audit logging | Ready | All critical actions logged |
| Session management | Ready | Active sessions list, remote revocation |
| Production hardening | Ready | Log rotation, auto backups, health checks, monitoring |
| Desktop app | Roadmap | Tauri scaffolding present |
| Mobile app | Roadmap | Flutter scaffolding present |
| Public SDK / CLI | Roadmap | Planned for future release |

## Architecture

```text
vexa/
├── apps/
│   ├── web/          # Next.js frontend (React 19, TypeScript)
│   ├── api/          # Go backend (Gin, JWT, WebAuthn, audit logging)
│   ├── desktop/      # Tauri v2 + Rust desktop client (roadmap)
│   └── mobile/       # Flutter mobile client (roadmap)
├── packages/
│   ├── ssh-core/     # Rust SSH/SFTP/tunnel core + FFI (roadmap)
│   ├── ui/           # Shared React components
│   ├── types/        # Shared TypeScript types
│   └── config/       # Shared eslint/tailwind/tsconfig
├── docs/             # Contributor & deployment docs
├── scripts/          # Operational scripts
├── docker-compose.yml
├── docker-compose.prod.yml
└── Makefile
```

The web app talks to the API over HTTP/WebSocket. The API stores data in PostgreSQL, caches sessions in Redis, and performs SSH/SFTP operations through the Go backend. Native performance-critical paths are planned to move into the Rust `ssh-core` package over time.

## Security Highlights

- **Authentication:** Argon2id password hashing + TOTP two-factor + WebAuthn/FIDO2 passkeys
- **Sessions:** JWT access tokens + refresh tokens, active session management with remote revocation
- **Encryption:** AES-256-GCM credential encryption at rest with unlockable master vault
- **Network:** Rate limiting, brute-force protection, CORS origin enforcement
- **Deployment:** TLS support in production (see `Caddyfile` and `docker-compose.prod.yml`)
- **Audit:** Comprehensive logging for all critical actions
- **Pipeline:** Pre-commit SAST, secret scanning, and dependency scanning

## Development

```bash
# Start database and cache
docker compose up -d postgres redis

# Run API
cd apps/api
go run cmd/server/main.go

# Run Web (another terminal)
cd apps/web
npm run dev
```

For the full contributor setup — including required tools, pre-commit hooks, and lint commands — see [`docs/dev/getting-started.md`](docs/dev/getting-started.md).

## Security

For vulnerability disclosure, see [`SECURITY.md`](SECURITY.md).

## Contributing

We welcome contributions! See [`CONTRIBUTING.md`](CONTRIBUTING.md) for the full developer guide — setup, branching, pre-commit hooks, verification gates, and PR process.

## License

[MIT](LICENSE) — © 2025-2026 vexa contributors.

## Support & Community

- **Bug reports & feature requests:** open a [GitHub issue](https://github.com/soumabali/vexa/issues).
- **Discussions & questions:** use [GitHub Discussions](https://github.com/soumabali/vexa/discussions).
- **Security:** email `sudharmika@gmail.com` — see [`SECURITY.md`](SECURITY.md).
