# P9 Dispatch — Host Detail Enhancement

**Tanggal:** 2026-06-28
**Branch:** `feat/p9-host-detail-enhancement`
**Dispatcher:** Ame (Hermes)

---

## 📜 Konteks

Audit P9:
- ✅ HostDetailPage ada, 436 baris, sudah 11 useState (kompleks)
- ✅ Sudah ada: edit dialog, delete confirm, tunnel create dialog, connection button
- **Gap:** Quick stats card belum ada, copy-to-clipboard utilities belum ada, layout bisa di-improve

**Strategi:** Refactor HostDetailPage jadi modular dengan sub-components + tambah quick stats.

---

## 🎯 3 Task

### Task 1 — Extract Sub-Components

**File baru:**
- `apps/web/src/components/hosts/HostStatsCard.tsx` — Quick stats (last connected, tunnel count, host type, status)
- `apps/web/src/components/hosts/HostMetadataPanel.tsx` — IP, port, user, OS, tags
- `apps/web/src/components/hosts/CopyableField.tsx` — Reusable copy-to-clipboard component

### Task 2 — Tambah Quick Stats Endpoint Backend

**File:** `apps/api/internal/api/handlers/hosts.go` (extend)

Tambah endpoint `GET /api/v1/hosts/:id/stats`:
```go
type HostStats struct {
    TotalSessions   int       `json:"total_sessions"`
    LastConnectedAt *time.Time `json:"last_connected_at,omitempty"`
    TunnelCount     int       `json:"tunnel_count"`
    ActiveTunnels   int       `json:"active_tunnels"`
}

func (h *HostHandler) GetStats(c *gin.Context) {
    // ...
}
```

Register di router.go.

### Task 3 — Refactor HostDetailPage

**File:** `apps/web/src/app/hosts/[id]/page.tsx`

- Pakai `HostStatsCard`, `HostMetadataPanel`, `CopyableField`
- Reduce dari 436 baris ke ~250 baris
- Better layout: stats card di atas, metadata panel di bawah, action buttons prominent

---

## 🚨 Pitfalls

1. HANYA edit `02-application/`. No commit/push/.git.
2. JANGAN ubah business logic existing (edit/delete/create tunnel sudah jalan).
3. Pakai `useAsyncData` pattern untuk fetch stats — JANGAN raw useEffect+setState.
4. CopyableField harus accessible (proper aria-label, keyboard support).

## ✅ Verification

```bash
cd apps/api && go build ./...
cd apps/api && go test ./tests/... -count=1
cd apps/web && npm run typecheck
cd apps/web && npm run build
```

## 📦 Commit (3-4 commit)

```
feat(hosts): HostStatsCard HostMetadataPanel CopyableField components
feat(hosts): backend endpoint GET /api/v1/hosts/:id/stats
feat(hosts): wire stats endpoint in router.go
refactor(hosts): HostDetailPage uses sub-components
```

Print summary.
