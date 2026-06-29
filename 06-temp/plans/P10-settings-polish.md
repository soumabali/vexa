# P10 Settings Polish

> Source of truth: `Obsidian/infra/vexa Roadmap.md` P10
> Status: BELUM DIMULAI
> Repo: `soumabali/vexa` (public), `02-application/` adalah working dir Claude Code

---

## Scope (4 sub-tasks per Obsidian)

| # | Sub-task | Backend | Frontend |
|---|----------|---------|----------|
| 1 | Active sessions list | ⚠️ ada stub handler (`GetActiveSessions` cuma return count + message) | ❌ belum ada page |
| 2 | Kill session | ✅ `RevokeSession` handler + route terdaftar | ❌ belum ada UI button |
| 3 | Recovery codes display | ✅ `RegenerateBackupCodes` handler | ❌ belum dialog display |
| 4 | WebAuthn rename/delete | ✅ `PATCH`+`DELETE` /auth/webauthn/credentials/:id | ✅ page `/settings/webauthn` sudah lengkap (commit sebelumnya) |

**Catatan:** Sub-task 4 SUDAH SELESAI dari sprint sebelumnya. Fokus sprint ini: 1+2+3 (semua frontend-only + 1 backend extension untuk Task 1).

---

## 🚨 Pitfalls (non-negotiable)

1. **HANYA edit `02-application/`.** No commit/push dari Claude Code. No edit root repo.
2. **JANGAN ubah backend behavior yang sudah production-tested** kecuali GetActiveSessions stub yang eksplisit disebutkan Task 1.
3. **MAX 30 errors/dispatch.** Jika ada lint error yang muncul dari pekerjaan ini, fix inline — jangan tulis `eslint-disable` sebagai escape hatch.
4. **Per-task commit dengan conventional prefix.** Format: `feat(settings)` / `feat(auth)` / `fix(auth)`.
5. **Restore generated artifacts** jika harus generate untuk verifikasi: `.next/`, `tsconfig.tsbuildinfo`. **JANGAN commit.**
6. **Pakailah `useAsyncData` (file `apps/web/src/hooks/useAsyncData.ts`) untuk fetching — jangan raw useEffect + setState yang antipattern React 18.**
7. **Untuk dialog baru, pakai komponen `Dialog` + `DialogContent` dari `@/components/ui/dialog` (sudah ada).** Jangan pakai library lain.
8. **Untuk toast, pakai `toast` dari `sonner`** (sudah dipakai di security/page.tsx).
9. **Loader pakai `Loader2` dari `lucide-react`** + class `animate-spin`.
10. **Empty state pakai copy singkat** ("No active sessions.", "No backup codes generated yet.") — jangan pakai placeholder atau hardcoded test data.
11. **Recovery codes display = one-time-show.** Modal/dialog HARUS punya warning "These codes will only be shown once." + tombol "I have saved them" sebelum user bisa menutup. Setelah close, codes TIDAK boleh ada di state, localStorage, atau console.
12. **Active session current (this device) HARUS visually distinct** dari session lain (misal: badge "This device" + tidak ada tombol revoke).
13. **JANGAN tampilkan token/secret apa pun** di UI active session. Hanya metadata (browser, OS, IP, last active).

---

## Sub-task 1 — Active Sessions List

### Tujuan
Tampilkan semua session aktif user, kasih kemampuan revoke session lain.

### Backend extension (`apps/api/internal/api/handlers/auth.go`)

**File:** `apps/api/internal/api/handlers/auth.go`
**Function:** `GetActiveSessions` (sekarang stub)

Ganti implementasi dengan enumerasi real dari `sessionStore`:

```go
// GetActiveSessions returns metadata for all active sessions of authenticated user.
func (h *AuthHandler) GetActiveSessions(c *gin.Context) {
    userIDVal, _ := c.Get("user_id")
    userID := userIDVal.(uuid.UUID)

    sessions, err := h.sessionStore.ListUserSessions(c.Request.Context(), userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list sessions"})
        return
    }

    type SessionMetadata struct {
        SessionID    string    `json:"session_id"`
        IPAddress    string    `json:"ip_address"`
        UserAgent    string    `json:"user_agent"`
        CreatedAt    time.Time `json:"created_at"`
        LastActiveAt time.Time `json:"last_active_at"`
        IsCurrent    bool      `json:"is_current"`
    }

    currentSessionID, _ := c.Get("session_id")
    currentIDStr, _ := currentSessionID.(string)

    metadata := make([]SessionMetadata, 0, len(sessions))
    for _, s := range sessions {
        metadata = append(metadata, SessionMetadata{
            SessionID:    s.ID,
            IPAddress:    s.IPAddress,
            UserAgent:    s.UserAgent,
            CreatedAt:    s.CreatedAt,
            LastActiveAt: s.LastActiveAt,
            IsCurrent:    s.ID == currentIDStr,
        })
    }

    c.JSON(http.StatusOK, gin.H{"sessions": metadata})
}
```

**Lokasi SessionStore:** cek `internal/auth/session_store.go` (atau nama setara). Field name mungkin `ID`, `IPAddress`, `UserAgent`, `CreatedAt`, `LastActiveAt`, `UserID`. Sesuaikan dengan struct actual.

**Wiring:** route sudah ada (`GET /auth/sessions`), TIDAK perlu edit `router.go`.

### Tests

Tambah unit test di `apps/api/tests/api/auth_test.go` atau per-file baru `auth_sessions_test.go`:
- `GetActiveSessions_ReturnsEmptyList_ForNewUser`
- `GetActiveSessions_ReturnsMultipleSessions_WithCurrentMarked`
- `RevokeSession_CannotRevokeCurrentSession`
- `RevokeSession_CannotRevokeOtherUsersSession`

### Acceptance
- `go test ./...` di `apps/api` hijau.
- `GetActiveSessions` mengembalikan array `sessions` dengan field wajib (`session_id`, `ip_address`, `user_agent`, `created_at`, `last_active_at`, `is_current`).

---

## Sub-task 2 — Frontend: Session Manager Page + Revoke Flow

### Tujuan
Page `/settings/sessions` dengan daftar session aktif + tombol revoke per session.

### Files

**BARU: `apps/web/src/app/settings/sessions/page.tsx`**
- Pakai `DashboardLayout` + `Card` + table/list pattern dari `security/page.tsx`.
- Fetch sessions via `useAsyncData(() => sessionsApi.listSessions(), [])` — tidak raw useEffect.
- Render list dengan field:
  - Browser/OS (parsed dari `user_agent` pakai library atau simple regex)
  - IP address (CopyableField dari P9)
  - Created at + Last active (relative time, format "2 minutes ago")
  - Badge "This device" untuk `is_current === true`
  - Tombol "Revoke" disabled untuk current session, danger-styled untuk yang lain
- Empty state: "No other active sessions."
- Dialog konfirmasi sebelum revoke (pattern `Dialog` dari `security/page.tsx`).
- Optimistic update: setelah revoke sukses, refresh list.

**EDIT: `apps/web/src/lib/api/auth.ts`**
Tambah type & methods:

```typescript
export interface ActiveSession {
  session_id: string;
  ip_address: string;
  user_agent: string;
  created_at: string;
  last_active_at: string;
  is_current: boolean;
}

export interface ActiveSessionsResponse {
  sessions: ActiveSession[];
}

// inside authApi object:
listSessions: () =>
  apiRequest<ActiveSessionsResponse>("/api/v1/auth/sessions"),

revokeSession: (sessionId: string) =>
  apiRequest<{ message: string }>("/api/v1/auth/sessions/revoke", {
    method: "POST",
    body: JSON.stringify({ session_id: sessionId }),
  }),
```

**EDIT: `apps/web/src/components/layouts/sidebar.tsx`** (atau nav equivalent)
Tambah link "Active Sessions" di section Settings — lihat posisi link Security/WebAuthn yang sudah ada.

### Tests

**BARU: `apps/web/src/app/settings/__tests__/sessions.test.tsx`**
- `SessionsPage renders active sessions list`
- `SessionsPage marks current session`
- `SessionsPage revokes non-current session via confirmation dialog`
- `SessionsPage does not show revoke button for current session`
- `SessionsPage shows empty state when no sessions`

### Acceptance
- `npm run typecheck` hijau.
- `npm run lint` 0 errors (warnings boleh).
- `npm test -- src/app/settings` hijau.
- Page reachable via `/settings/sessions`.

---

## Sub-task 3 — Frontend: Recovery Codes Display Dialog

### Tujuan
Dialog one-time-show untuk recovery codes dari TOTP MFA, dipicu dari settings security page.

### Files

**EDIT: `apps/web/src/app/settings/security/page.tsx`**

Di section `Card` untuk 2FA TOTP (sekitar `mfaEnabled && ...`), tambah:
- Tombol "Regenerate backup codes" yang membuka dialog.
- Dialog berisi:
  - Warning "These codes will replace any existing codes. Save them now — they will only be shown once."
  - Input "Verification code" (TOTP 6 digit) — required.
  - Tombol "Generate codes" submit.
- Setelah submit sukses:
  - Tutup dialog input.
  - Buka dialog kedua (one-time display):
    - Header: "Save your backup codes"
    - List codes dalam monospace, 8 codes (asumsi GenerateRecoveryCodes(8)).
    - Tombol "Copy all" + tombol "Download .txt" (generate Blob, `URL.createObjectURL`).
    - Tombol "I have saved them" untuk close.
- Setelah close dialog kedua, HAPUS codes dari state.

**BARU (atau tambahkan di security/page.tsx): `BackupCodesDialog.tsx`**
Component yang encapsulate dialog display, props: `open`, `onOpenChange`, `codes: string[]`, `onClose`.

### Tests

Tambah di `apps/web/src/app/settings/__tests__/settings.test.tsx`:
- `SecuritySettingsPage shows regenerate backup codes button when 2FA enabled`
- `BackupCodesDialog displays codes list`
- `BackupCodesDialog clears codes on close`
- `BackupCodesDialog has copy-all button`

### Acceptance
- `npm run typecheck` hijau.
- Page `/settings/security` punya tombol regenerate. Click → dialog input TOTP → submit → dialog display codes → close → state cleared.

---

## Sub-task 4 — WebAuthn rename/delete UI (VERIFICATION ONLY)

### Tujuan
**Verify sudah selesai.** Tidak ada code change.

### Verifikasi
- Buka `apps/web/src/app/settings/webauthn/page.tsx` (sudah ada).
- Confirm: ada tombol rename + delete per credential.
- Confirm: `renameCredential` + `deleteCredential` functions defined di `apps/web/src/lib/webauthn.ts`.
- Confirm: backend endpoint `PATCH /auth/webauthn/credentials/:id` + `DELETE /auth/webauthn/credentials/:id` wired di router.

### Acceptance
- Output dari Claude Code: "P10 sub-task 4 SUDAH COMPLETE. No changes made."

---

## Verification Gates (WAJIB hijau sebelum report)

```bash
# Backend
cd /home/ubuntu/projects/vexa/02-application/apps/api
go build ./... && echo "BUILD OK"
go test ./... -count=1
go vet ./...

# Frontend
cd /home/ubuntu/projects/vexa/02-application/apps/web
npm run typecheck && echo "TYPECHECK OK"
npm run lint        # 0 errors (warnings boleh)
npm test -- --run
```

**Target lint delta:** 0 → 0 (sub-task 1+2+3 strictly tambah fungsional, BUKAN modifikasi existing — pastikan tidak menambah unused vars).

---

## 📤 Output yang diharapkan

```
=== P10 SETTINGS POLISH SUMMARY ===

Task 1 (Active Sessions): <LOC, files modified>
Task 2 (Sessions Page UI): <LOC, files modified>
Task 3 (Recovery Codes UI): <LOC, files modified>
Task 4 (WebAuthn — verify only): ALREADY COMPLETE, 0 files changed

Commits:
- feat(auth)sessions list endpoint GET /auth/sessions real implementation (P10 task 1)
- test(auth)sessions handler tests (P10 task 1)
- feat(auth)lib sessions API list + revoke (P10 task 2)
- feat(settings)sessions manager page + sidebar link (P10 task 2)
- test(settings)sessions page tests (P10 task 2)
- feat(settings)backup codes regenerate dialog (P10 task 3)
- test(settings)backup codes dialog tests (P10 task 3)

Verification gates:
- go build: <PASS/FAIL>
- go test: <PASS/FAIL, #tests>
- npm typecheck: <PASS/FAIL>
- npm lint: <PASS/FAIL, errors: X>
- npm test: <PASS/FAIL, #tests>

Total commits: 7 (Target: 6-8, MAX 10)
```

**Kerjakan tanpa konfirmasi tambahan.** Ame verifikasi setelah selesai.

---

## Catatan Tambahan untuk Claude Code

1. **Untuk webauthn sub-task 4**, JANGAN modify apapun. Cukup inspect + output "ALREADY COMPLETE".
2. **Session store field names** mungkin berbeda dari asumsi — `internal/auth/session_store.go` (atau setara) adalah source of truth. Gunakan field names yang ada.
3. **Sidebar nav edit**: cari file `apps/web/src/components/layouts/sidebar.tsx` ATAU `nav.tsx` ATAU equivalent — multiple layouts mungkin ada, pilih yang dipakai DashboardLayout.
4. **Dialog pattern sudah ada** di `security/page.tsx` line 285+. Reference implementation.
5. **useAsyncData pattern**: lihat cara pakai di `apps/web/src/hooks/useAsyncData.ts` + contoh di `apps/web/src/app/hosts/[id]/page.tsx` (P9 hasil).
6. **Per-task commit, jangan squash.** Detailed history = audit value.
