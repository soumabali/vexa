## Bug Report: SSH Manager E2E Testing

### BUG-001: Host Edit Page Navigation
- **Severity:** Medium
- **Status:** Open
- **Description:** When navigating to host edit page, the URL requires a specific host ID format. The page doesn't handle navigation from host list properly.
- **Steps to Reproduce:**
  1. Login to SSH Manager
  2. Go to hosts list
  3. Click edit button on a host
  4. Page shows 404 or doesn't load
- **Expected:** Should navigate to `/hosts/[id]/edit` or show edit modal
- **Actual:** Navigation fails

### Notes
- Host creation works fine
- Host list displays correctly
- Issue is specifically with edit navigation

### BUG-008: E2E Validation Messages Mismatch
- **Severity:** Low
- **Status:** Confirmed
- **Description:** E2E test expects `"at least 8 characters"` but actual Zod validation uses `"Password must be at least 8 characters"`. Similarly, `"passwords do not match"` expectation may not match actual validation.
- **Steps to Reproduce:** Run `short password validation` or `mismatched passwords` e2e tests.
- **Expected:** Validation error text should be visible
- **Actual:** Locator not found

### BUG-009: Docker Web Build - GID Conflict
- **Severity:** Medium
- **Status:** Fixed
- **Description:** `RUN addgroup --system --gid 1000 nodejs` fails because `node:20-alpine` already has a group with gid 1000. Also pnpm v11 requires Node.js >= 22, but Dockerfile uses node:20-alpine.
- **Fix:** Changed gid to 1001 (committed). Need to update base image to node:22-alpine for pnpm compatibility.

### BUG-010: API - CGO Database Conflict
- **Severity:** High
- **Status:** Confirmed
- **Description:** `Dockerfile.dev` builds with `CGO_ENABLED=0` but `db/db.go` imports `go-sqlite3` which requires CGO. Additionally, `db/db.go` always opens SQLite even when `DATABASE_URL` is set to PostgreSQL. A separate `db/pool.go` has PostgreSQL support but is unused.
- **Workaround:** Run API natively (CGO enabled by default) with SQLite.
- **Impact:** Docker-based deployment of API fails on startup.

### BUG-011: Protected Pages Broken Without Auth
- **Severity:** High
- **Status:** Open
- **Description:** E2E tests for `/hosts`, `/settings`, and other protected pages fail because unauthenticated users are redirected to `/login` which returns 404 (BUG-007). Root page redirects to `/hosts` which then redirects to `/login` → 404.
- **Steps to Reproduce:**
  1. Visit `http://localhost:3000/` without auth cookie
  2. Expected: Either show content or redirect to working login page
  3. Actual: 404 page shown
- **Impact:** Blocks all E2E UI testing of protected features
- **Related:** BUG-007 (login page renders as 404)
