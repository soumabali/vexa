# Cleanup Sprint Dispatch — Web Lint & Warnings

**Tanggal:** 2026-06-28
**Branch:** `chore/web-lint-cleanup`
**Dispatcher:** Ame (Hermes)
**Executor:** Claude Code di `02-application/`

---

## 📜 Konteks untuk Claude Code

Kamu adalah Claude Code, **coding executor** untuk project **vexa**.
**WAJIB** baca sebelum eksekusi:

1. `02-application/CLAUDE.md`
2. `00-meta/skills.md`
3. `apps/web/.eslintrc.json` — config strict (security + jsx-a11y)

**Tidak seperti P4/P5 yang feature-focused**, sprint ini purely **code quality**.
- ZERO functional change
- ZERO business logic change
- HANYA refactor cosmetic / type narrowing / unused cleanup

---

## 🔍 Audit Hasil Lint

```
266 errors total:
  205 @typescript-eslint/no-unused-vars     (Paling banyak)
   42 @typescript-eslint/no-explicit-any
   16 react-hooks/exhaustive-deps
    2 @typescript-eslint/no-empty-object-type
    1 @next/next

+ 224 warnings (dependencies, console.log, dll)
```

**Top 10 file:** file-list-view (9), rdp-viewer (8), vnc-viewer (7), useSession (5),
discovery/page (5), admin/metrics/page (5), terminal.tsx (3), file-upload (3),
file-preview (3), recordings/page (3).

---

## 🎯 Strategi Per Kategori

### Kategori 1: `no-unused-vars` (205 errors — bulk of work)

**Strategy:**
- **Variabel yang memang tidak dipakai**: HAPUS.
- **Variabel yang dipakai tapi prefix `_`**: rename `_varName`.
- **Function parameter tidak dipakai** (callback signature harus match): prefix `_param`.
- **Import tidak dipakai**: HAPUS import-nya.
- **Type import tidak terpakai**: HAPUS.
- **Function yang return error tapi error tidak dicek** (idiomatic Go-style tapi
  tidak perlu di TS): cek apakah perlu di-handle atau prefix `_`.

**Jangan:**
- Jangan ubah business logic
- Jangan rename exported identifier (akan break import di tempat lain)
- Jangan disable rule untuk "lebih cepat"

### Kategori 2: `no-explicit-any` (42 errors)

**Strategy (urutan preferensi):**
1. **Ganti ke tipe spesifik** jika interface/type sudah ada.
2. **Ganti ke `unknown`** + type narrowing dengan `if (typeof x === 'string')` dll.
3. **Gunakan generic** jika pola generic.
4. **Last resort**: `// eslint-disable-next-line @typescript-eslint/no-explicit-any`
  dengan komentar alasan (e.g. "third-party API tidak punya types").

**Contoh patterns:**
```typescript
// ❌ any
function parse(data: any) { ... }

// ✅ unknown + narrowing
function parse(data: unknown) {
  if (typeof data === 'object' && data !== null) { ... }
}

// ✅ generic
function parse<T>(data: T) { ... }

// ⚠️ last resort (third-party)
function parse(data: any) { // eslint-disable-line @typescript-eslint/no-explicit-any
  // some-lib has no types
}
```

### Kategori 3: `react-hooks/exhaustive-deps` (16 errors)

**Strategy:**
- **Missing dep yang memang dibutuhkan**: tambahkan ke array.
- **Missing dep yang tidak perlu** (e.g. ref, dispatch stable): gunakan pattern
  `useEffect(() => {...}, [/* eslint-disable-line */])` JANGAN disable kalau
  bisa fix dengan benar.
- **Function yang dibuat inline** dan menjadi dep: wrap dengan `useCallback`
  atau extract keluar component.

### Kategori 4: `no-empty-object-type` (2 errors)

**Strategy:** Ganti `interface Foo {}` dengan `type Foo = Record<string, never>`
atau tambahkan property. Cek konteksnya.

### Kategori 5: `@next/next` (1 error)

**Strategy:** Biasanya `<a>` yang harusnya `<Link>`. Fix sesuai dokumentasi Next.js.

### Warnings (224 — JANGAN fokus utama)

Tujuan sprint ini: **eliminate semua errors** agar `--max-warnings=0` someday possible.
Tapi **prioritas #1 adalah 266 errors**. Untuk warnings:
- Hanya fix warning yang ** mudah dan obvious** (e.g. unused import jika sekalian
  sudah di-handle).
- **JANGAN** spent time refactor besar untuk warning — itu sprint terpisah.

---

## 🚨 Pitfalls (KRUSIAL untuk cleanup sprint)

1. **NO business logic change.** Diff harusnya mostly deletions + type narrowing.
   Jika butuh ubah logic, **STOP, tanya Ame** — itu di luar scope.
2. **NO breaking exports.** Jangan rename exported function/variable/interface.
   Lint akan pass di satu file tapi break import di 5 file lain.
3. **NO mass disable.** Dilarang `/* eslint-disable */` global atau per-file.
   Boleh **per-line** dengan komentar alasan yang jelas (max 5% dari total fix).
4. **NO prettier-only changes.** Jangan ubah formatting yang tidak terkait rule
   yang violated.
5. **NO generated file changes.** File di `.next/`, `dist/`, dll.
6. **HANYA edit di `02-application/`.**
7. **TIDAK commit, TIDAK push, TIDAK `.git` di dalam `02-application/`.**

---

## 📦 Commit Strategy (3-4 commit)

Karena 266 errors itu besar, pecah jadi **3-4 commit** berdasarkan kategori:

### Commit 1: `chore(lint): remove unused vars imports (266 → ~61)`
Fix semua `no-unused-vars` (205 errors) + bulk import cleanup.

### Commit 2: `chore(lint): type narrowing no-explicit-any (61 → ~19)`
Fix `no-explicit-any` (42 errors).

### Commit 3: `chore(lint): fix react-hooks deps + minor rules (19 → 0)`
Fix `react-hooks/exhaustive-deps` (16) + `no-empty-object-type` (2) + `@next/next` (1).
Plus easy warnings kalau sempat.

**Format commit:**
```
chore(lint): <kategori> (<sebelum> → <sesudah>)
```

---

## ✅ Verification Gates (WAJIB hijau setiap commit)

```bash
# 1. Typecheck (WAJIB hijau — jangan break types)
cd apps/web && npm run typecheck

# 2. Lint (harus turun setiap commit)
cd apps/web && npm run lint
# Target akhir: 0 errors (warnings boleh tetap untuk sprint berikutnya)

# 3. Build (WAJIB hijau — pastikan tidak ada broken refactor)
cd apps/web && npm run build

# 4. Tests (vitest)
cd apps/web && npm test
```

**Untuk verifikasi setiap sub-bagian** (bukan cuma akhir):
- Setelah commit 1: `npm run lint` harus turun drastis
- Setelah commit 2: errors harusnya ~19
- Setelah commit 3: errors harusnya 0

---

## 📤 Output yang diharapkan

Setelah selesai, print summary:

```
=== CLEANUP SPRINT SUMMARY ===
Initial state: 266 errors, 224 warnings

Commit 1 (no-unused-vars): 205 → 0 (target)
Commit 2 (no-explicit-any): 42 → 0 (target)
Commit 3 (hooks deps + minor): 19 → 0 (target)

Final state: 0 errors, X warnings (X >= 0, ideally turun tapi bukan blocker)

Verification gates:
- npm typecheck: <PASS/FAIL>
- npm lint: <PASS/FAIL, errors: X>
- npm build: <PASS/FAIL>
- npm test: <PASS/FAIL, #tests>

Disables added (eslint-disable): <count, harusnya <5% dari total fix>
Files modified: <count>
LOC delta: <negative karena mostly deletion>
```

**Kerjakan tanpa konfirmasi tambahan.** Ame verifikasi setelah kamu selesai.

---

## 🔚 End

Setelah print summary di atas, exit. Jangan push (Ame yang push dari root).
