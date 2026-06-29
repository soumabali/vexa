# Refactor Strategis Lint — Phase 2 Dispatch

**Tanggal:** 2026-06-28
**Branch:** `chore/refactor-strategic-lint`
**Dispatcher:** Ame (Hermes)
**Audit hasil:** 80 errors teridentifikasi, root cause per kategori ditemukan

---

## 📜 Konteks untuk Claude Code

Kamu adalah Claude Code, coding executor vexa. **WAJIB** baca sebelum eksekusi:
1. `02-application/CLAUDE.md`
2. `apps/web/.eslintrc.json`
3. File-file di bawah (lihat setiap task)

**3 task diurutkan**: Task 1 (types, paling foundational) → Task 2 (hooks) → Task 3 (components).

**Commit per task**. Jangan batch. Push dari Ame nanti.

---

## Task 1 — Type Definitions & `no-explicit-any` (42 errors)

### 1a. Buat `apps/web/src/types/discovery.ts`

Ambil dari backend Go types sebagai source of truth:

```typescript
// apps/web/src/types/discovery.ts

export interface PortResult {
  port: number;
  service?: string;
  banner?: string;
  protocol: 'tcp' | 'udp';
  tls?: boolean;
}

export interface ScanResult {
  id: string;
  ip_address: string;
  hostname?: string;
  open_ports: PortResult[];
  os_guess?: string;
  response_time_ms: number;
  status: 'up' | 'down' | 'filtered';
  scanned_at: string;
}

export interface ScanOptions {
  timeout_ms: number;
  rate_limit_ms: number;
  concurrency: number;
  icmp_first: boolean;
  os_scan: boolean;
  resolve_hostname: boolean;
}

export interface ScanJob {
  id: string;
  network: string;
  ports: number[];
  status: 'pending' | 'running' | 'completed' | 'cancelled' | 'failed';
  progress: number;
  results: ScanResult[];
  total_hosts: number;
  scanned_hosts: number;
  created_at: string;
  started_at?: string;
  completed_at?: string;
  error?: string;
  options: ScanOptions;
}

export interface StartScanRequest {
  network: string;
  ports: number[];
  options: Partial<ScanOptions>;
}

export interface StartScanResponse {
  id: string;
  status: string;
  network: string;
  message: string;
  created_at: string;
}

export interface ScanResultsResponse {
  scan: ScanJob;
  results: ScanResult[];
}

export interface ScanHistoryResponse {
  jobs: ScanJob[];
}

export interface ImportResultsRequest {
  import_all: boolean;
  selected_ips?: string[];
}

export interface ImportResultsResponse {
  imported: number;
  failed: number;
  host_ids: string[];
}
```

### 1b. Augment next-auth Session di `apps/web/src/types/next-auth.d.ts`

```typescript
// apps/web/src/types/next-auth.d.ts
import 'next-auth';

declare module 'next-auth' {
  interface Session {
    accessToken?: string;
    user?: {
      id?: string;
      name?: string | null;
      email?: string | null;
      image?: string | null;
    };
  }
}
```

### 1c. Refactor `apps/web/src/app/discovery/page.tsx` (8 errors)

Ganti semua `: any` dengan type dari `discovery.ts`:

```typescript
// L91: const response: StartScanResponse = await api<StartScanResponse>('/api/v1/discovery/scan', {...})
// L120: const response: ScanResultsResponse = await api<ScanResultsResponse>(...)
// L129: const response: ScanHistoryResponse = await api<ScanHistoryResponse>(...)
// L139: const response: ScanJob = await api<ScanJob>(...)
// L174: const response: ImportResultsResponse = await api<ImportResultsResponse>(...)
// catch: catch (error) { const message = error instanceof Error ? error.message : 'Failed'; ... }
```

Note: `api()` function adalah generic, pakai `api<T>()` untuk type hasil.

### 1d. Refactor `apps/web/src/app/vault/share/page.tsx` (5 errors)

Ganti `(session as any)?.accessToken` dengan:
```typescript
import { useSession } from 'next-auth/react';
// session sekarang Session (dari next-auth) + accessToken (dari augment)
const token = session?.accessToken;
```

### 1e. Refactor `apps/web/src/components/rdp-viewer.tsx` & `vnc-viewer.tsx`

L173/L177 rdp-viewer `handleMouseMove(e as any)` → cek signature, fix tanpa any:
```typescript
// Jika handler butuh React.MouseEvent, gunakan langsung
// Jangan cast any — refactor function signature
```

L304 rdp-viewer `onConnect: (params: any) => void` → cek struct `params` dan type:
```typescript
// Cari tipe ConnectParams dari RDPViewerProps.connect usage atau define:
interface ConnectParams {
  hostname: string;
  port: number;
  username: string;
  password: string;
  domain?: string;
}
onConnect: (params: ConnectParams) => void;
```

Commit 1:
```
refactor(types): discovery types + next-auth augment, replace any (42 → ~10 errors)
```

---

## Task 2 — React Hooks Patterns (25 errors total)

### 2a. `set-state-in-effect` (16 errors)

Pattern: `useEffect(() => { fetchData(); setX(data); }, [])` 

Fix pattern: **derive state during render, not in effect**. Contoh untuk admin pages:

```typescript
// ❌ Before
useEffect(() => { fetchLogs(); }, []);
const fetchLogs = async () => {
  const res = await fetch('/api/admin/audit-logs');
  const data = await res.json();
  setLogs(data.logs || []);
};

// ✅ After: use a data fetching hook or initialization pattern
const [logs, setLogs] = useState<Log[] | null>(null);
useEffect(() => {
  let cancelled = false;
  (async () => {
    const res = await fetch('/api/admin/audit-logs');
    const data = await res.json();
    if (!cancelled) setLogs(data.logs || []);
  })();
  return () => { cancelled = true; };
}, []);
```

File target (semua admin pages + recordings): admin/audit, admin/hosts, admin/metrics, admin/sessions, admin/users, sessions/recordings, settings/api-keys, settings/webauthn, vault/share, dst.

### 2b. `immutability` (9 errors)

File target: rdp-viewer.tsx, vnc-viewer.tsx, file-upload.tsx, file-list-view.tsx.

Pattern violations:
- `rdp-viewer L74`: `handleBinaryMessage(event.data as ArrayBuffer)` referenced before declaration
- `vnc-viewer L111`: same
- `file-upload L45`: `addFiles(droppedFiles)` accessed before declared

**Fix**: Reorder declarations so functions are defined BEFORE `useCallback` that references them. Atau gunakan lazy reference pattern.

### 2c. `static-components` (4 errors)

File: `file-list-view.tsx` L65/L72.
Pattern: Inline component `SortIcon` defined inside render causes reset state.

**Fix**: Extract `SortIcon` ke module-level component (di atas `FileListView`).

### 2d. `purity` (2 errors)

File: `rdp-viewer.tsx` L33 `Date.now()` di body component.
**Fix**: `const lastTimeRef = useRef<number>(0);` initialized lazily:
```typescript
if (lastTimeRef.current === 0) lastTimeRef.current = Date.now();
// atau
useEffect(() => { lastTimeRef.current = Date.now(); }, []);
```

Commit 2:
```
refactor(hooks): fix set-state-in-effect, immutability, static-components (25 → 0 errors)
```

---

## Task 3 — Cosmetic & Empty Types (7 errors)

### 3a. `no-unescaped-entities` (5 errors)

File: `app/settings/api-keys/page.tsx` L310/L338, `components/profile/DeleteAccountForm.tsx`.
Fix:
```jsx
// ❌ "won't" atau 'I've'
// ✅ {"won't"} atau {'I\u2019ve'}
// atau gunakan template literal
// "won\u2019t"
```

### 3b. `no-empty-object-type` (2 errors)

File: `components/ui/input.tsx` L4 dan file lain.
```typescript
// ❌ interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {}
// ✅ type InputProps = React.InputHTMLAttributes<HTMLInputElement>;
```

Commit 3:
```
chore(lint): unescaped entities + empty object types (7 → 0 errors)
```

---

## 🚨 Pitfalls

1. **HANYA edit `02-application/`** — tidak commit, tidak push, tidak `.git`
2. **JANGAN ubah backend types** — types/discovery.ts di frontend adalah **copy** dari backend
3. **JANGAN disable rule** untuk skip
4. **JANGAN rename exported identifiers**
5. **Pakai type yang sudah ada** jika ada di `src/types/` — `tunnel.ts`, `file-manager.ts` bisa jadi reference
6. **api() function signature**: `apiRequest<T>(path, options): Promise<T>` — pakai generic, bukan cast

## ✅ Verification Gates

Setiap task WAJIB hijau sebelum commit:

```bash
cd apps/web && npm run typecheck   # WAJIB exit 0
cd apps/web && npm run lint         # errors harus TURUN
cd apps/web && npm run build        # WAJIB exit 0
```

Final state target: **0 errors**. Warnings boleh tetap.

## 📤 Output

Print summary per task:
```
=== TASK 1 (types) ===
- discovery.ts created: <lines>
- next-auth.d.ts created: <lines>
- files modified: <list>
- errors before: 42, after: <count>

=== TASK 2 (hooks) ===
- files modified: <list>
- errors before: 25, after: <count>

=== TASK 3 (cosmetic) ===
- files modified: <list>
- errors before: 7, after: 0

=== FINAL ===
- total errors: 0
- total warnings: <count>
- typecheck: PASS
- build: PASS
- commits: 3
```

Kerjakan tanpa konfirmasi. Ame verifikasi setelah selesai.
