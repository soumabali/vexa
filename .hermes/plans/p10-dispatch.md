# P10 Dispatch — Settings Polish

**Tanggal:** 2026-06-29
**Branch:** `feat/p10-settings-polish`
**Dispatcher:** Ame (Hermes)
**Executor:** Claude Code di `02-application/`

---

## Konteks

Audit P10 (post-merge P9 commit `f41c32f`):
- Backend SEMUA endpoint P10 sudah live: `/auth/sessions`, `/auth/sessions/revoke`, `/auth/mfa/backup-codes/regenerate`, `PATCH/DELETE /auth/webauthn/credentials/:id`.
- Frontend backend interface (`apps/web/src/lib/api/auth.ts`) **belum** punya method sessions/backup-codes display.
- Page `/settings/sessions` belum ada.
- WebAuthn sub-task sudah selesai — verify only.
- Backend `GetActiveSessions` masih stub (return count + message, bukan real list). Perlu extension.

## Strategi

Sprint ini frontend-heavy + 1 backend extension. Backend stub di-replace dengan real implementation, frontend tambah 2 page/component + 1 dialog. WebAuthn sub-task verify-only.

## Sub-tasks (4)

Detail lengkap: `06-temp/plans/P10-settings-polish.md`

### Task 1: Backend Sessions List Endpoint
- Replace stub `GetActiveSessions` di `apps/api/internal/api/handlers/auth.go`.
- Return array `sessions` dengan metadata + `is_current` flag.
- Tambah tests di `apps/api/tests/api/`.

### Task 2: Frontend Sessions Manager Page
- Page baru: `apps/web/src/app/settings/sessions/page.tsx`.
- Edit `apps/web/src/lib/api/auth.ts` tambah `listSessions` + `revokeSession`.
- Edit sidebar nav tambah link "Active Sessions".
- Tests di `apps/web/src/app/settings/__tests__/sessions.test.tsx`.

### Task 3: Frontend Recovery Codes Dialog
- Edit `apps/web/src/app/settings/security/page.tsx` tambah tombol regenerate.
- Component baru (atau inline): dialog one-time display backup codes dengan Copy + Download .txt + close warning.
- Tests extend `settings.test.tsx`.

### Task 4: WebAuthn (Verify Only)
- Confirm `apps/web/src/app/settings/webauthn/page.tsx` sudah lengkap rename/delete.
- Output: "P10 sub-task 4 ALREADY COMPLETE. No changes."
- **JANGAN modify file.**

## Estimate Commits

Target 6-8 commits, MAX 10.

Per-task commit convention:
```
feat(auth)sessions list endpoint real implementation (P10 task 1)
test(auth)sessions handler tests (P10 task 1)
feat(auth)lib sessions API client (P10 task 2)
feat(settings)sessions manager page (P10 task 2)
test(settings)sessions page tests (P10 task 2)
feat(settings)backup codes regenerate dialog (P10 task 3)
test(settings)backup codes tests (P10 task 3)
fix(...)(optional, hanya jika ada lint cepat)
```

## Pitfalls

1. HANYA edit `02-application/`. No commit/push.
2. JANGAN modify webauthn page (sub-task 4 verify-only).
3. Per-task commit. JANGAN squash.
4. Pakai `useAsyncData` untuk fetching — lihat `apps/web/src/hooks/useAsyncData.ts`.
5. Pakai `Dialog` component existing dari `@/components/ui/dialog`.
6. Pakai `toast` dari `sonner`.
7. MAX 30 errors/dispatch — jika ada lint error baru, fix inline.
8. Recovery codes HAPUS dari state setelah user close dialog display.

## Verify Gates (WAJIB hijau)

```bash
cd 02-application/apps/api && go build ./... && go test ./... -count=1
cd 02-application/apps/web && npm run typecheck && npm run lint && npm test -- --run
```

Lint: 0 errors (216 warnings existing boleh, jangan naik).

## Output

Print summary menggunakan template dari `06-temp/plans/P10-settings-polish.md` section "Output yang diharapkan".

Exit tanpa push. Ame yang push dari root setelah verifikasi.

---

## End

Setelah print summary, exit.
