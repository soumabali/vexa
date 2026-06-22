# vexa — Complete SSH Manager

Self-hosted SSH access management with a security-first architecture.

- **Web app** (Next.js) — manage hosts, credentials, tunnels, and browser-based SSH terminal
- **API** (Go + Gin + PostgreSQL + Redis) — REST, WebSocket terminal, WireGuard tunnels, vault encryption, audit logging
- **Desktop app** (Tauri) — planned / experimental
- **Mobile app** (Flutter) — planned / experimental

## Quick Start (Docker)

```bash
git clone https://github.com/soumabali/vexa.git
cd vexa
cp .env.example .env
# edit .env with strong secrets
docker compose up -d
```

Open http://localhost:3000 and register the first admin account.

For production self-hosting, see [`docs/devops/self-hosted-deployment.md`](docs/devops/self-hosted-deployment.md).

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

See `docs/dev/getting-started.md` for the full setup guide.

## Documentation

> **Orchestrator-level documentation** (workflow, roadmap, credentials, runbook) is maintained privately in the root orchestrator repo: `https://github.com/soumabali/vexa-root`.  
> This application repo contains only **open-source contributor docs** (`docs/`) and **operational scripts** (`scripts/`).

## Features

| Feature | Status | Notes |
|---------|--------|-------|
| Host management | ✅ Ready | SSH hosts, labels, inventory |
| Credential vault | ✅ Ready | AES-GCM encrypted, unlockable master vault |
| SSH terminal | ✅ Ready | xterm.js via WebSocket, host selector, multi-tab |
| SFTP file manager | ✅ Ready | Upload, download, navigate remote files |
| WireGuard tunnels | ✅ Ready | Per-user tunnels, enable/disable/rotate |
| WebAuthn / passkeys | ✅ Ready | FIDO2 security keys |
| TOTP MFA | 🚧 Planned | Backend scaffold exists, UI placeholder |
| Audit logging | ✅ Ready | All critical actions logged |
| Desktop app | 🚧 Roadmap | Tauri scaffolding present |
| Mobile app | 🚧 Roadmap | Flutter scaffolding present |

## Architecture

```
vexa/
├── apps/
│   ├── web/          # Next.js frontend
│   ├── api/          # Go backend
│   ├── desktop/      # Tauri desktop app (roadmap)
│   └── mobile/       # Flutter mobile app (roadmap)
├── packages/
│   ├── ssh-core/     # Rust SSH/SFTP core (roadmap)
│   ├── ui/           # Shared React components
│   ├── types/        # Shared TypeScript types
│   └── config/       # Shared eslint/tailwind/tsconfig
├── docs/             # Contributor & deployment docs
├── scripts/          # Operational scripts
├── docker-compose.yml
└── Makefile
```

## Security

- Argon2 password hashing
- AES-GCM credential encryption
- JWT access + refresh tokens
- WebAuthn / passkey support
- Rate limiting and audit logging
- CORS enforced via `ALLOWED_ORIGINS`
- TLS support in production

## Roadmap

1. TOTP MFA real backend + QR setup flow
2. Desktop app (Tauri) MVP
3. Mobile app (Flutter) MVP
4. Credential team sharing / rotation
5. SSH terminal mobile integration
6. Public SDK / CLI

## Contributing

1. Fork the repository and create a branch: `feature/<description>`.
2. Follow the verification gates (`make verify`).
3. Open a PR against `main`.

See `docs/dev/getting-started.md` for the development setup.

## License

MIT — see `LICENSE`.

## Support

For bugs and feature requests, please open a GitHub issue.
