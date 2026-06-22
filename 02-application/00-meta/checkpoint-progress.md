# Vexa / SSH Manager — Checkpoint Progress

> Updated: 2026-06-19 16:50+08:00
> Session: continuation from prompt-lanjut-vexa.md

## P0 — API bisa start ✅

- [x] Fix config DB URL / password fallback
  - `apps/api/config/config.go`: `buildDatabaseURL()` now composes `DATABASE_URL` from `DB_USER`, `DB_PASSWORD`, `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_SSL_MODE` when `DATABASE_URL` is empty; dev fallback password is `changeme`.
- [x] Fix secrets dir di dev
  - `.env`: added `SECRETS_DIR=/home/ubuntu/projects/vexa/05-config/secrets`
  - `config.go`: `resolveSecretsDir()` falls back to project config dir in development, `/var/lib/ssh-manager/secrets` in production.
  - Created `/home/ubuntu/projects/vexa/05-config/secrets` with `0700` permissions.
- [x] Fix encryption key length
  - `.env`: `ENCRYPTION_KEY=dev-encryption-key-for-local-dev` (exactly 32 bytes for AES-256-GCM).
- [x] Start API, verify `/health` 200
  - API running on `http://0.0.0.0:8080`
  - `GET /health/ready` → 200
  - `GET /health/live` → 200
  - `GET /health` → 200 (redirects to `/health/ready`)

## P1 — Host CRUD E2E stabil ✅

- [x] Fix edit host form validation error (`tags` null → default `[]`)
- [x] Fix delete host empty-body parsing (`204 No Content` handling)
- [x] Unique `data-testid` per host row/menu/card
- [x] Cleanup E2E host test prefix (`E2E*` dan `Debug*`)
- [x] Delete duplicate `debug-edit.spec.ts`
- [x] Refactor `HostList` ke `@tanstack/react-query` untuk menghindari setState di effect
- [x] Fix `NEXT_PUBLIC_API_URL` via `apps/web/.env.local`

## P2 — Terminal page ✅

- [x] Buat `/terminal` page dengan tabs, status indicator, terminal container
- [x] E2E terminal 5 test passing

## P3 — UI/UX alignment ✅

- [x] Shared `TopNav` dengan link Hosts, Terminal, Tunnels, Vault, Settings
- [x] Shared `DashboardLayout` untuk pages hosts dan terminal
- [x] Gunakan Next.js `Link` (bukan tag `<a>`)

## P4 — Audit & best practice ✅

- [x] Fix error ESLint pada file yang diedit (0 errors)
- [x] Refactor `LoginForm` pakai `useRouter().push()` alih-alih `window.location.href`
- [x] Refactor `HostList` pakai React Query
- [x] Install Go toolchain (`/usr/local/go/bin/go`)
- [x] Review backend Go — full `go test ./...` passing
  - `internal/crypto`: vault encrypt/decrypt format konsisten; `DecryptCredential` fix reconstruct ciphertext.
  - `internal/audit`: `NewTestLogger` + `NewNoOpLogger` untuk handler test.
  - `internal/api/handlers`: nil guard audit, terminal, hosts, auth login.
  - `internal/terminal`: input validation & sanitization (shell metacharacters, dangerous words, null bytes).
  - `tests/`: fix import paths, mock expectations, skip stubbed SSH proxy test.
- [x] Security hardening CORS, secrets, audit log
  - `ALLOWED_ORIGINS` wildcard dilarang di production.
  - Secrets disimpan di `/home/ubuntu/projects/vexa/05-config/secrets` (0700), bukan `/tmp`.
  - Audit logger menulis ke file + HMAC + opsi DB.

## P5 — End-to-end testing ✅

- [x] Verifikasi Playwright
- [x] Run `npx playwright test` di dev server Turbopack
- [x] Hasil dev: **15 passed**
- [x] Jalankan production standalone server dengan static chunks yang lengkap
- [x] Run `npx playwright test` di production build
- [x] Hasil production: **15 passed**

## P6 — Dokumentasi ✅

- [x] Update checkpoint `00-meta/checkpoint-progress.md`
- [x] Update `03-history/sessions/session-2026-06-19-go-audit.md`
- [x] Link perubahan ke Obsidian Vault
  - `infra/vexa.md` progress updated.
  - `credentials/vexa Credentials.md` values filled.

## P7 — Settings foundation ✅

- [x] Wire settings overview (`/settings`) dengan semua section cards
- [x] Consolidate profile page ke `/settings/profile` dengan tabs: Profile, Password, Delete
- [x] Link Security (`/settings/security`) dan WebAuthn (`/settings/webauthn`) sudah ada
- [x] Production build sukses
- [x] E2E settings spec `4 passed`
- [x] Stabilkan full E2E suite dengan shared storage state — **19 passed**

## P8 — Next Steps (Roadmap) ✅

- [x] Implementasi real SSH terminal via WebSocket + koneksi host aktual.

## P9 — Tunnels page integration ✅

- [x] Fix Tunnels list API response shape (`{ tunnels }`)
- [x] Wrap Tunnels page with `DashboardLayout`
- [x] Wire enable/disable/rotate/delete actions ke backend
- [x] Tambah `onDeleted` callback di `TunnelDeleteDialog`
- [x] Full E2E suite tetap **19 passed**

## P10 — Vault, Security, Host-Tunnel Integration ✅

- [x] Vault page: real backend list/unlock/create/delete/rotate credentials
- [x] Settings security: live WebAuthn passkey list, TOTP placeholder UI
- [x] Settings webauthn: wrap with DashboardLayout + fix API base path
- [x] Host detail page: add "Create Tunnel" dialog with allowed IPs/port/DNS
- [x] Full E2E suite tetap **19 passed**

## P11 — Next Steps (Roadmap) ⏳

- [ ] Implement real TOTP MFA backend + QR setup flow.
- [ ] Implement credential sharing flow backend + frontend (teams/permissions).
- [ ] WireGuard tunnel live status + handshake stats UI.
- [ ] SSH terminal connection E2E (dengan mock host atau dev SSH container).
- [ ] Obsidian Vault automation / session note generator integration.
