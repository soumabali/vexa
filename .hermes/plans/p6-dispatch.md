# P6 Dispatch Prompt — Credential Sharing (Teams) End-to-End

**Tanggal:** 2026-06-28
**Branch:** `feat/p6-credential-sharing`
**Dispatcher:** Ame (Hermes)
**Executor:** Claude Code di `02-application/`

---

## 📜 Konteks untuk Claude Code

Kamu Claude Code, coding executor **vexa**. WAJIB baca sebelum eksekusi:
1. `02-application/CLAUDE.md`
2. `00-meta/skills.md`
3. `apps/api/internal/team/service.go` — Team model sudah ada
4. `apps/api/internal/vault/share_handlers.go` — Handler sudah ada tapi BELUM di-wire
5. `apps/api/internal/vault/sharing.go` — ShareRepository ada, real implementation
6. `apps/web/src/app/vault/share/page.tsx` — Frontend sudah ada, 354 baris

**P6 = wire up + implement stubs**. Bukan greenfield.

---

## 🔍 Audit Real — Yang Sudah Ada (JANGAN tulis ulang)

| Komponen | Status | File |
|----------|--------|------|
| `Team`, `TeamMember`, `SharedHost` model | ✅ | `apps/api/internal/team/service.go` |
| `CreateTeam/GetTeam/UpdateTeam/DeleteTeam/ListTeams` | ✅ WORKING | `team/service.go` |
| `AddMember/RemoveMember/UpdateMemberRole` | ⚠️ STUB (audit only, no DB) | `team/service.go` |
| `ShareHost/UnshareHost` | ⚠️ STUB | `team/service.go` |
| `ShareRepository` (real impl with E2E crypto) | ✅ 470 LOC | `vault/sharing.go` |
| `ShareHandler` (Create/Accept/Revoke/List/Get/Reject) | ⚠️ NOT WIRED | `vault/share_handlers.go` |
| `RegisterShareRoutes()` defined | ⚠️ Never called | `vault/share_handlers.go:199` |
| `ShareWithTeam` in CredentialService | ❌ Placeholder "not implemented" | `vault/credential_service.go:631` |
| Frontend `vault/share/page.tsx` | ✅ 354 LOC UI | `apps/web/src/app/vault/share/page.tsx` |
| Migrations `008_credential_sharing.sql` | ✅ Tables exist | `apps/api/internal/db/migrations/` |
| Tests: `team/service_test.go` (19 tests) | ✅ Most PASS | `tests/team/service_test.go` |
| Tests: `vault_share_handlers_test.go` | ⏭ SKIPPED | `tests/vault_share_handlers_test.go` |
| Tests: `vault_sharing_test.go` | ⏭ SKIPPED | `tests/vault_sharing_test.go` |

---

## 🎯 Scope 4 Task

### Task 1 — Wire Share Routes (Quick Win)

**File:** `apps/api/internal/api/router.go`

Tambah inisialisasi `ShareHandler` dan panggil `RegisterShareRoutes`:

```go
// Di setup function router.go (cari tempat setelah credHandler)
// ShareHandler untuk E2E credential sharing
shareRepo := vault.NewShareRepository(db)
shareHandler := vault.NewShareHandler(shareRepo)

// Daftarkan routes
vault.RegisterShareRoutes(authenticated, shareHandler)
```

**Acceptance:**
- Backend `go build ./...` clean
- Endpoint `/api/v1/vault/shares` (GET, POST) accessible dengan auth
- Endpoint `/api/v1/vault/shares/:id/{accept,reject,revoke}` accessible

---

### Task 2 — Fix Frontend-Backend Request Mismatch

**Problem:** Frontend POSTs `{recipient_email, permission, expiry_days}` ke
`/vault/credentials/:id/share`. Backend expects `{team_id, permissions[]}`.

**Pilih salah satu approach (pilih yang paling clean):**

**Option A (recommended)**: Update frontend `vault/share/page.tsx` line ~140-150
- Ganti `recipient_email` → `team_id` (butuh UI ubah dari email input ke team selector)
- Ganti `permission` (string) → `permissions` (array of strings)

**Option B**: Update backend `handlers/credentials.go` Share() untuk terima email
dan resolve ke user_id internal (lebih simple, less frontend changes)

**Rekomendasi:** Option A karena:
- Email-based share lebih cocok untuk individual user-to-user (sebelum ada teams UI)
- Backend `team_id` field暗示 udah siap untuk P7/P8 yang fokus teams

**Acceptance:**
- POST `/vault/credentials/:id/share` accepts `{team_id, permissions[]}`
- Frontend tidak break typecheck
- ShareWithTeam masih return error "not implemented" — itu task 3

---

### Task 3 — Implement ShareWithTeam di CredentialService

**File:** `apps/api/internal/vault/credential_service.go:630`

Ganti placeholder dengan real implementation yang delegates ke ShareRepository:

```go
func (s *CredentialService)ShareWithTeam(ctx context.Context, userID uuid.UUID, credID uuid.UUID, teamID uuid.UUID, permissions string)error {
 // Verify user owns credential
 // Verify user is admin/owner of team
 // Verify team members can access (use TeamService.IsTeamMember)
 // Call ShareRepository.CreateShare with proper E2E encryption

 shareRepo := NewShareRepository(s.db)
 // ... (delegate ke share_repo dengan proper crypto)

 return nil
}
```

**Pattern:** Mirror logic dari `team/service.go` `ShareHost` tapi untuk credentials.

**Acceptance:**
- Go test `TestShareWithTeam_Success` PASS
- Audit log entry dibuat (`audit.EventCredentialShare`)
- Error handling untuk: credential not owned, not team member, etc.

---

### Task 4 — Implement Team Member Stubs

**File:** `apps/api/internal/team/service.go:201-254`

Replace stubs dengan real DB operations:

- `AddMember`: INSERT ke `team_members` table (cek migration untuk schema)
- `RemoveMember`: DELETE dengan cek owner tidak bisa dihapus
- `UpdateMemberRole`: UPDATE role

**Cek dulu:**
```bash
grep -A5 "CREATE TABLE team_members" apps/api/internal/db/migrations/*.sql
```

**Acceptance:**
- Go tests `TestTeamService_AddMember` PASS (tambah test baru)
- Audit log tetap dipanggil

---

## 🚨 Pitfalls (KRUSIAL)

1. **HANYA edit `02-application/`.** Jangan commit, push, atau `.git` di sana.
2. **JANGAN tulis ulang kode yang sudah ada.** ShareRepository sudah punya E2E crypto — pakai saja.
3. **JANGAN ubah migration files.** Tabel `team_members`, `credential_shares` sudah ada.
4. **JANGAN hapus stub tests yang skip.** Hanya fix agar tidak skip.
5. **Hormati permission model**: owner/admin only untuk AddMember/RemoveMember/UpdateRole.
6. **Audit log TETAP dipanggil** untuk setiap perubahan.
7. **ShareWithTeam HARUS return nil atau wrapped error** — jangan panic.

---

## ✅ Verification Gates (WAJIB hijau)

```bash
# Backend
cd apps/api && go build ./...
cd apps/api && go test ./tests/team/... -v -count=1
cd apps/api && go test ./tests/vault_share_handlers_test.go -v -count=1
cd apps/api && go test ./tests/ -run 'TestShare|TestTeam|TestMember' -v -count=1

# Frontend
cd apps/web && npm run typecheck
cd apps/web && npm run lint
```

**Expected:** semua PASS. Skipped tests harus jadi PASS (unskip dengan implementasi).

---

## 📦 Commit Strategy (4 commit)

```
wip(share): wire ShareHandler routes RegisterShareRoutes (P6 task 1)
fix(share): frontend recipient_email team_id migration (P6 task 2)
feat(share): ShareWithTeam real implementation ShareRepository delegate (P6 task 3)
feat(team): AddMember RemoveMember UpdateMemberRole real DB (P6 task 4)
```

---

## 📤 Output

Print summary:
```
=== P6 SUMMARY ===
Task 1 (wire routes): PASS/FAIL + endpoints registered
Task 2 (request mismatch): PASS/FAIL + approach used
Task 3 (ShareWithTeam impl): PASS/FAIL + test result
Task 4 (team member stubs): PASS/FAIL + test result

Verification:
- go build: PASS/FAIL
- go test team: PASS/FAIL + #tests
- go test share: PASS/FAIL + #tests
- npm typecheck: PASS/FAIL
- npm lint: PASS/FAIL

Commits: 4
Files modified: <count>
```

Kerjakan tanpa konfirmasi tambahan. Ame verifikasi setelah selesai.
