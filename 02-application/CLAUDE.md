# Vexa / SSH Manager — Claude Code Context

## Project Identity

- **Name:** vexa
- **Type:** Monorepo (web + API + desktop + mobile + shared packages)
- **Server:** 43.156.128.55 (Tencent Cloud)
- **Model default untuk Claude Code:** `kimi-k2.7-code:cloud` via Ollama launcher
- **Permission mode:** `--dangerously-skip-permissions` (gunakan dengan hati-hati)
- **Lokasi project:** `/home/ubuntu/projects/vexa/02-application/`
- **Obsidian note:** `infra/vexa.md`
- **Credentials:** `credentials/vexa Credentials.md`

## Arsitektur

```
apps/
  api/          Go backend (Gin, JWT, MFA, WebAuthn)
  web/          Next.js 16 + React 19 + TypeScript
  desktop/      Tauri v2 + Rust
  mobile/       Flutter
packages/
  ssh-core/     Rust SSH/SFTP/tunnel library + FFI
  ui/           Shared UI components
  types/        Shared TypeScript types
  config/       Shared eslint/tailwind/tsconfig
deploy/         Helm charts + K8s manifests
infra/          Terraform + K8s + docker scripts
security/       Compliance + pentest + fuzz tests
tests/          Playwright E2E + integration tests
docs/           Documentation
```

## Ports & URLs

| Port | Service | Catatan |
|------|---------|---------|
| 3000 | web (Next.js dev) | http://localhost:3000 |
| 8080 | api (Go/Gin dev) | http://localhost:8080 |
| 5432 | PostgreSQL | localhost |
| 6379 | Redis | localhost |

## Commands Penting

```bash
# Jalankan dev stack
cd /home/ubuntu/projects/vexa/02-application
make dev

# Backend only
cd apps/api && go run cmd/server/main.go

# Frontend only
cd apps/web && npm run dev

# Tests
cd apps/api && go test ./...
cd apps/web && npm test
cd tests/e2e && npx playwright test

# Build production
cd apps/api && go build -o server cmd/server/main.go
cd apps/web && npm run build
```

## Standar yang Harus Dijaga

### Keamanan (P1)

- **JANGAN** hardcode secret/credential di source code.
- **JANGAN** set `AllowedOrigins: []string{"*"}` untuk production.
- **JANGAN** simpan secret di `/tmp`.
- Selalu gunakan env vars atau Docker secrets.
- Semua input user wajib divalidasi.
- Audit log wajib ditulis untuk setiap aksi kritis.

### Code Style

- **Go:** `gofmt`, konvensi Go standard, layered architecture `internal/`.
- **TypeScript/React:** ESLint + Prettier, Radix UI + Tailwind, vitest untuk unit test.
- **Rust:** `cargo fmt`, `cargo clippy`.
- **Flutter:** `dart format`.

### Git Workflow

- Semua perubahan lewat branch baru: `feature/<nama>`.
- Commit message dalam bahasa Inggris, gunakan conventional commits.
- Sebelum merge: test + lint + review.
- **Jangan** commit `.env`, binary, atau node_modules.

## Skill yang Aktif

Claude Code akan memuat skill-skill ini:

1. **superpowers** — coding superpowers dari obra/superpowers
2. **caveman** — caveman hooks dan workflows dari juliusbrussee/caveman
3. **graphify** — codebase graph understanding dari safishamsi/graphify
4. **playwright** — browser automation MCP (@executeautomation/playwright-mcp-server)

## Aturan Interaksi

- Setiap aktivitas di project ini harus dilakukan via Claude Code (`ollama launch claude`).
- Gunakan `--dangerously-skip-permissions` untuk workflow otomatis.
- Batasi max-turns sesuai kompleksitas task (default 10-20).
- Setelah selesai, laporkan ringkasan ke Hermes/Ame.

## Dokumentasi

- `01-documents/review-2026-06-19.md` — hasil review awal.
- `02-application/docs/` — dokumentasi asli upstream.
- `00-meta/` — ports, urls, credentials mapping.

## Notes

- Go belum terinstall di server; install terlebih dahulu sebelum build backend.
- Node deps `apps/web` memiliki conflict `react-simple-maps` vs React 19; perlu fix sebelum `npm install`.
- Rust/Flutter belum terinstall untuk desktop/mobile.
