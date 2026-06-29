# P7 Dispatch — WireGuard Live Stats End-to-End

**Tanggal:** 2026-06-28
**Branch:** `feat/p7-wireguard-live-stats`
**Dispatcher:** Ame (Hermes)

---

## 📜 Konteks

Audit P7 state:
- ✅ Backend `wireguard/stats.go` (100 LOC) sudah ada
- ✅ Endpoint `GET /api/v1/tunnels/:id/stats` sudah di-wire di router.go
- ✅ Handler `tunnelHandler.Stats()` ada
- ❌ Frontend `tunnels/page.tsx` TIDAK pakai stats endpoint (0 grep hits)

**Real gap:** Frontend perlu integrasi stats endpoint + UI live stats panel dengan auto-refresh.

---

## 🎯 3 Task

### Task 1 — Frontend Stats API Client
**File baru:** `apps/web/src/lib/api/tunnels.ts` (extend existing)

Pastikan ada typed method:
```typescript
export interface TunnelStats {
  bytes_received: number;
  bytes_sent: number;
  last_handshake: string | null;
}

export const tunnelsApi = {
  // ...existing methods
  getStats: (id: string) => apiRequest<{ stats: TunnelStats }>(`/api/v1/tunnels/${id}/stats`).then(d => d.stats),
};
```

### Task 2 — Live Stats Panel Component
**File baru:** `apps/web/src/components/tunnels/LiveStatsPanel.tsx`

```typescript
interface LiveStatsPanelProps {
  tunnelId: string;
  refreshIntervalMs?: number; // default 10000 (10s)
}

// Shows:
// - Bytes received (formatted: KB/MB/GB)
// - Bytes sent (formatted)
// - Last handshake (relative time: "5 min ago")
// - Auto-refresh toggle
// - Connection status indicator (green/red dot)
```

Requirements:
- Use `useAsyncData` hook pattern (yang sudah ada dari refactor phase 2)
- AbortController untuk cleanup polling saat unmount
- Pause polling saat tab tidak visible (Page Visibility API)
- Error state dengan retry button

### Task 3 — Integrate ke Tunnel Detail Page
**File:** `apps/web/src/app/tunnels/page.tsx`

Tambah tab atau section yang menampilkan `LiveStatsPanel` untuk tunnel yang dipilih.

---

## 🚨 Pitfalls

1. HANYA edit `02-application/`. No commit, no push.
2. JANGAN ubah backend (sudah real). Fokus frontend.
3. JANGAN ubah existing tunnels API client methods.
4. Pakai `useAsyncData` pattern dari refactor phase 2 — JANGAN buat ulang dengan useEffect+setState.

## ✅ Verification

```bash
cd apps/web && npm run typecheck
cd apps/web && npm run build
cd apps/web && npm run lint
```

## 📦 Commit (3 commit)

```
feat(tunnels): stats API client typed getStats method
feat(tunnels): LiveStatsPanel component with auto-refresh
feat(tunnels): integrate LiveStatsPanel ke tunnel detail page
```

Print summary dengan final state.
