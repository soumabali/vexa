# 20260622 — P5: TOTP MFA Real Backend (End-to-End)

## Goal

Mengaktifkan TOTP MFA end-to-end: dari setup di Settings, verifikasi kode, enable/disable MFA, sampai login flow yang menghandle `mfa_required`.

## Scope

Files di `02-application/` yang boleh diubah:
- `apps/api/internal/api/handlers/auth.go` — jika ada bug di MFA endpoint
- `apps/api/internal/api/router.go` — jika perlu tambah route
- `apps/api/internal/auth/user_service.go` — jika perlu tambah helper
- `apps/web/src/app/settings/security/page.tsx` — TOTP setup/verify/disable UI
- `apps/web/src/app/login/page.tsx` — TOTP second-step login UI
- `apps/web/src/lib/auth.ts` atau file baru — helper MFA API client
- `tests/e2e/specs/auth.spec.ts` — MFA flow E2E (jika memungkinkan tanpa mock SMTP/TOTP)
- `tests/e2e/fixtures/` — fixture data jika perlu

## Task Detail

### 1. Backend Verification (minimal changes)

1. Baca dan verifikasi endpoint berikut sudah benar:
   - `POST /api/v1/auth/mfa/setup` — return `{ qr_code, uri, backup_codes }`
   - `POST /api/v1/auth/mfa/enable` — terima `{ totp_code }`, validasi dari session, persist ke user
   - `DELETE /api/v1/auth/mfa/disable` — disable MFA user
   - `POST /api/v1/auth/mfa/verify` — login step 2 dengan `{ mfa_token, totp_code }`
2. Jika ditemukan bug (misal setup tidak mengembalikan `secret` plain untuk QR, atau `mfaService.ValidateTOTP` salah), perbaiki.
3. Pastikan `userService` methods: `EnableMFA`, `DisableMFA`, `StoreMFASetupSession`, `GetMFASetupSession`, `DeleteMFASetupSession`, `StoreMFAPendingSession`, `GetMFAPendingSession`, `DeleteMFAPendingSession` benar.

### 2. Frontend Settings Security TOTP Setup

1. Di `apps/web/src/app/settings/security/page.tsx`:
   - Ganti switch placeholder menjadi fetch real status MFA dari `/api/v1/users/me`.
   - Tambah dialog/modal untuk setup TOTP:
     - Step 1: call `POST /api/v1/auth/mfa/setup`, tampilkan QR code (gambar base64), URI, dan backup codes.
     - Step 2: input TOTP code 6 digit, call `POST /api/v1/auth/mfa/enable`.
     - Jika sukses, tampilkan backup codes dan beri tombol "I have saved the backup codes".
   - Jika MFA sudah enabled, tampilkan tombol "Disable MFA" dengan konfirmasi.
2. Pastikan UI tetap menggunakan Tailwind + komponen Shadcn yang sudah ada.

### 3. Frontend Login MFA Step

1. Di `apps/web/src/app/login/page.tsx`:
   - Setelah submit login, jika response `mfa_required: true`, simpan `mfa_token` dan tampilkan form input TOTP 6 digit.
   - Submit `POST /api/v1/auth/mfa/verify` dengan `{ mfa_token, totp_code }`.
   - Setelah sukses, redirect ke `/hosts` seperti login normal.
   - Handle error invalid code dengan toast.

### 4. E2E Tests (jika feasible)

1. Jika E2E bisa menggunakan authenticator fixture, tambahkan spec:
   - Register user → enable MFA → logout → login dengan TOTP → sukses.
2. Jika tidak feasible karena perlu secret TOTP di test, buat test minimal:
   - Endpoint `/auth/mfa/setup` mengembalikan qr_code dan backup_codes.
   - UI setup dialog muncul setelah klik setup.

### 5. Verification Gates

Setelah semua perubahan:
- `cd apps/api && go test ./...` — harus pass
- `cd apps/api && go build ./...` — harus pass
- `cd apps/web && npm run build` — harus pass
- Copy `.next/static` ke `.next/standalone/.next/static`
- Jalankan E2E dev jika memungkinkan: `npx playwright test --workers=1 tests/e2e/specs/auth.spec.ts` atau full suite.

## Mandatory Skill Stack

Gunakan `superpowers`, `caveman`, `graphify`. Jika ada perubahan E2E, gunakan `playwright` MCP.
Setelah eksekusi, respond ONLY dengan valid JSON sesuai `/home/ubuntu/projects/vexa/00-meta/claude-response-schema.json`.

## Important

- Do not edit root files.
- Do not push.
- Do not include real secrets in code or response.
- If a task is too large, focus on getting the backend wiring + settings UI + login flow functional; recovery code endpoint can be a TODO comment for P5-follow-up.

## Permissions

`--allowedTools "Read,Write,Edit,Bash(go test),Bash(go build),Bash(npm run build),Bash(npx playwright test),Bash(cp),Bash(cd),Bash(ls),Bash(make)"`
