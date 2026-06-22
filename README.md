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

<!-- add screenshots later -->

## Features

| Feature | Status | Notes |
|---------|--------|-------|
| Host management | Ready | SSH hosts, labels, inventory |
| Credential vault | Ready | AES-GCM encrypted master vault |
| SSH terminal | Ready | xterm.js via WebSocket, host selector, multi-tab |
| SFTP file manager | Ready | Upload, download, navigate remote files |
| WireGuard tunnels | Ready | Per-user tunnels, enable/disable/rotate keys |
| WebAuthn / passkeys | Ready | FIDO2 security-key support |
| Audit logging | Ready | All critical actions logged |
| TOTP MFA | Beta | Backend scaffold exists; QR setup flow in progress |
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

- Argon2id password hashing
- AES-GCM credential encryption with an unlockable master vault
- JWT access tokens + refresh tokens
- WebAuthn / passkey support (FIDO2)
- Rate limiting and brute-force protection
- Comprehensive audit logging
- CORS enforced via `ALLOWED_ORIGINS`
- TLS support in production (see `Caddyfile` and `docker-compose.prod.yml`)
- Pre-commit hooks scan for secrets, SAST issues, and unsafe code patterns

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

## Contributing

Contributions are welcome. Please follow the steps below to keep the project consistent and secure.

1. Fork the repository and create a branch: `feature/<short-description>`.
2. Run the automated setup: `./scripts/setup-hooks.sh`.
3. Follow the development guide at [`docs/dev/getting-started.md`](docs/dev/getting-started.md).
4. Make sure the verification gates pass locally: `make verify`.
5. Open a pull request against `main` with a clear description and test notes.

By contributing, you agree to keep credentials and secrets out of the codebase and to follow the security-focused workflow described in the contributor docs.

## License

[MIT](LICENSE) — © vexa contributors.

## Support & Community

- **Bug reports & feature requests:** open a [GitHub issue](https://github.com/soumabali/vexa/issues).
- **Discussions & questions:** use [GitHub Discussions](https://github.com/soumabali/vexa/discussions).
- **Security issues:** please see [`SECURITY.md`](SECURITY.md) for responsible disclosure.
