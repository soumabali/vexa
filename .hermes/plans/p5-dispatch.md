# P5 Dispatch Prompt — TOTP/MFA End-to-End + Web Gate Fixes

**Tanggal:** 2026-06-28
**Branch:** `feat/p5-totp-mfa-end-to-end`
**Dispatcher:** Ame (Hermes)
**Executor:** Claude Code di `02-application/`

---

## 📜 Konteks untuk Claude Code

Kamu adalah Claude Code, **coding executor** untuk project **vexa** (Complete SSH Manager).
**WAJIB** baca sebelum eksekusi:

1. `02-application/CLAUDE.md`
2. `AGENTS.md` (root) + `02-application/AGENTS.md` (jika ada)
3. `00-meta/skills.md`
4. `06-temp/plans/20260622-p5-totp-mfa-end-to-end.md` — plan P5 (referensi)

**BANYAK kode P5 SUDAH ADA** (backend MFA service, login flow, security page, API client).
Tugasmu: **verifikasi end-to-end + fix gap** — JANGAN tulis ulang dari nol.

---

## 🔍 Audit Hasil — Yang Sudah Ada (JANGAN rebuild)

| Komponen | Status | File |
|----------|--------|------|
| `MFAService` (TOTP generate/validate/encrypt) | ✅ | `apps/api/internal/auth/mfa.go` |
| Backup codes generator | ✅ | `apps/api/internal/auth/mfa.go` |
| MFA setup session store | ✅ | `apps/api/internal/auth/mfa_store.go` |
| Auth handler MFA endpoints | ✅ | `apps/api/internal/api/handlers/auth.go` |
| Login 2-step flow | ✅ | `apps/web/src/components/auth/LoginForm.tsx` |
| Security settings page (TOTP toggle + dialog) | ✅ | `apps/web/src/app/settings/security/page.tsx` |
| TwoFASetup component (QR + backup codes) | ✅ | `apps/web/src/components/security/TwoFASetup.tsx` |
| API client (`setupMFA`, `verify2FASetup`, `verifyMFA`, `verifyBackupCode`) | ✅ | `apps/web/src/lib/api/auth.ts` |
| MFA service Go tests | ✅ 3/3 PASS | `apps/api/tests/mfa_test.go` |
| **E2E MFA flow test** | ⚠️ MINIMAL (15 baris) | `tests/e2e/specs/auth.spec.ts` |

---

## 🎯 Scope Tugas (5 task)

### Task 1 — Web Gate Blockers (WAJIB fix sebelum task lain)

**Bloker:** typecheck dan lint gagal karena error pre-existing.

#### 1a. Fix `apps/web/src/components/terminal/__tests__/terminal.test.tsx`
**Error:** `Module '"@testing-library/react"' has no exported member 'screen'/'fireEvent'/'waitFor'`
**Fix:** Ganti ke named imports benar atau gunakan `@testing-library/react` v14+ syntax. Cek versi di `package.json` dan samakan.

#### 1b. Fix `apps/web/src/components/vault/__tests__/vault.test.tsx`
**Error:** sama seperti 1a.
**Fix:** sama.

#### 1c. Fix lint `no-explicit-any` di:
- `apps/web/src/lib/file-utils.ts` (lines 78, 81, 108)
- `apps/web/src/lib/webauthn.ts` (lines 6, 7)

**Fix:** Ganti `any` dengan tipe yang sesuai (interface atau `unknown` + narrowing).

**Acceptance:**
- `cd apps/web && npm run typecheck` exit code 0
- `cd apps/web && npm run lint` exit code 0

---

### Task 2 — E2E MFA Full Flow Test

**File:** `02-application/tests/e2e/specs/auth.spec.ts` (sekarang 15 baris, terlalu minimal).

**Tulis ulang dengan flow lengkap:**

```typescript
test.describe('MFA TOTP end-to-end', () => {
  test('user can enable MFA, logout, login with TOTP', async ({ page }) => {
    // 1. Login sebagai user tanpa MFA
    // 2. Navigate ke /settings/security
    // 3. Click "Set up authenticator" → dialog muncul
    // 4. Verify dialog menampilkan QR code + secret + backup codes
    // 5. Generate TOTP code dari secret (pakai otplib jika available, ATAU
    //    hit endpoint backend khusus test yang accept secret + return valid code)
    // 6. Input TOTP code → submit → success toast
    // 7. Verify MFA enabled (badge "Enabled" muncul)
    // 8. Logout
    // 9. Login dengan kredensial sama
    // 10. Expect MFA step muncul
    // 11. Submit TOTP code → expect redirect ke /hosts
  });

  test('user can disable MFA dengan valid TOTP code', async ({ page }) => {
    // 1. Login user yang sudah MFA enabled
    // 2. Navigate ke /settings/security
    // 3. Click "Disable" → confirm dialog
    // 4. Input TOTP code + confirm
    // 5. Verify MFA disabled (badge "Disabled" muncul)
  });

  test('invalid TOTP code ditolak dengan error message', async ({ page }) => {
    // 1. Login user MFA enabled
    // 2. Trigger MFA step (logout + login lagi)
    // 3. Input kode salah (e.g. "000000")
    // 4. Expect error toast + form tetap visible
  });
});
```

**CATATAN PENTING untuk testing TOTP:**
- Pakai library `otplib` jika sudah ada di `package.json` (cek deps)
- ATAU buat test helper endpoint di backend (HANYA di test mode) yang accept secret dan return valid code
- JANGAN hardcode TOTP code di test (akan expired)

**Acceptance:**
- File `auth.spec.ts` punya minimal 3 test cases di atas
- `npx playwright test tests/e2e/specs/auth.spec.ts --workers=1` PASS

---

### Task 3 — Security Page: Disable MFA Flow Audit

**File:** `apps/web/src/app/settings/security/page.tsx`

Audit & fix (jika ditemukan gap):

1. Apakah `disableMFA` endpoint benar-benar dipanggil dengan TOTP code valid?
2. Apakah ada loading state saat disable?
3. Apakah error handling tampil dengan benar?
4. Apakah setelah disable, UI re-fetch profile (badge update)?

**Acceptance:** Disable MFA flow berfungsi dari UI → backend → UI refresh.

---

### Task 4 — Recovery Code Endpoint (jika belum ada)

Cek `apps/api/internal/api/handlers/auth.go` apakah ada endpoint untuk regenerate backup codes.

**Jika belum ada:** tambah `POST /api/v1/auth/mfa/backup-codes/regenerate`:
- Require auth + MFA verified
- Validate TOTP code dulu
- Generate backup codes baru
- Return `{ backup_codes: [...] }`
- Audit log entry

**Jika sudah ada:** skip task ini (tulis "sudah ada" di summary).

---

### Task 5 — Hardcoded Secret Audit di Test

Cari test file yang hardcode TOTP secret / encryption key di source:

```bash
grep -rn "JBSWY3DPEHPK3PXP\|totp.Generate\|encryptionKey.*=" apps/api/tests/ apps/api/internal/auth/
```

**Fix:** Ganti semua hardcoded secret dengan env var atau generate random per test.

**Acceptance:** Tidak ada TOTP secret atau encryption key hardcoded di non-test code path.

---

## 🚨 Pitfalls (JANGAN ULANGI)

1. **Aturan baja**: edit HANYA di `02-application/`. Tidak commit, push, atau `.git` di sana.
2. **Pre-existing gate errors**: scope task 1 spesifik di file yang listed. Jangan fix noise lain.
3. **Web Dockerfile pitfall**: jangan ubah Dockerfile.
4. **Generated files**: jangan commit `.next/`, `dist/`, dll.
5. **MFA encryption key**: HARUS 32 bytes (sudah divalidasi di NewMFAService).
6. **Backup code regeneration**: HARUS invalidate backup code lama.
7. **E2E test**: jangan hardcode TOTP code — pakai otplib atau test helper.

---

## ✅ Verification Gates (WAJIB hijau)

```bash
# 1. Compose config (jika ada perubahan compose)
cd 02-application && docker compose config --quiet

# 2. Backend
cd 02-application/apps/api && go build ./...
cd 02-application/apps/api && go test ./tests/... -v

# 3. Web
cd 02-application/apps/web && npm run typecheck
cd 02-application/apps/web && npm run lint

# 4. E2E (jika backend running locally)
cd 02-application && npx playwright test tests/e2e/specs/auth.spec.ts --workers=1
```

**Semua gate HARUS hijau** sebelum commit.

---

## 📋 Commit Strategy

Commit per task (5 commit), conventional:

```
fix(test): resolve testing-library imports di terminal vault test (P5 task 1)
test(e2e): full MFA TOTP flow test setup disable invalid (P5 task 2)
fix(security): disable MFA flow audit hardening (P5 task 3)
feat(mfa): regenerate backup codes endpoint + audit (P5 task 4)
chore(security): replace hardcoded TOTP secrets env vars (P5 task 5)
```

Jika task 4 sudah ada → skip commit 4 (catat di summary).

---

## 📤 Output yang diharapkan

Setelah semua 5 task selesai, print summary:

```
=== P5 SUMMARY ===
Task 1 (web gate fixes): <PASS/FAIL + detail>
Task 2 (E2E MFA full flow): <PASS/FAIL + #tests>
Task 3 (disable MFA audit): <PASS/FAIL + finding>
Task 4 (backup code regen): <status: added/already-exists>
Task 5 (hardcoded secret audit): <# found + fixed>

Verification gates:
- go build: <PASS/FAIL>
- go test: <PASS/FAIL + #tests>
- npm typecheck: <PASS/FAIL>
- npm lint: <PASS/FAIL>
- playwright: <PASS/FAIL + #tests>

Commits made: <list of commit messages>
Files modified: <count>
```

**Kerjakan tanpa konfirmasi tambahan.** Ame verifikasi setelah kamu selesai.

---

## 🔚 End

Setelah print summary di atas, exit. Jangan push (Ame yang push dari root).
