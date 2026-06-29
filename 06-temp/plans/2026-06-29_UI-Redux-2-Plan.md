# UI Redesign Plan – Align Hosts & Navigation with Design System & Security  
**Project:** vexa (02‑application)  
**Author:** Hermes Agent (UI Redesign)  
**Created:** 2026‑06‑29 10:45 UTC+8  
**Branch:** `ui-redesign/hosts-auth-guard`  

> **Workflow Rule:** Must follow **Plan → Dispatch → Verify → Commit/Push → Document**.  
> **Atomic Commits:** Commit after every bite‑sized task.  
> **Verification Gates:** CI must stay green (19/19 E2E) before any push.  

---  

## 📋 Progress Tracker (Markdown Checklist)

> **Status key:**  
> - ✅ = completed  
> - ☐ = not started  

- [x] **T1 – Capture Baseline Screenshots** – assets copied to `.hermes/plans/assets/` (completed)  
- [ ] **T2 – Update Primary Color Token** – replace `--primary-color` with `#3B82F6` in `src/theme.css` (next)  
- [ ] **T3 – Adjust Heading Typography** – set heading “Welcome Back” to `font-size: 32px; font-weight: 700;` in `src/components/LoginCard.jsx`  
- [ ] **T4 – Fix Primary Button Style** – apply `background-color: var(--primary-color); border-radius: 6px;` in `src/components/PrimaryButton.module.css`  
- [ ] **T5 – Implement Security Badge Component** – create `Badge.jsx` with semantic colors  
- [ ] **T6 – Insert Badges on Host Cards** – add `<Badge status={host.security} />` in `src/pages/Hosts/HostCard.jsx`  
- [ ] **T7 – Add Route‑Guard Redirect** – client‑side redirect from `/hosts` to `/login` when `document.cookie` empty (in `src/App.jsx` or `src/routes/HostsRoute.jsx`)  
- [ ] **T8 – Add “Dashboard” Link to Top Bar** – insert Dashboard link in `src/components/TopBar.jsx`  
- [ ] **T9 – Add “Back‑to‑Dashboard” Button** – add button in `src/pages/Hosts/HostsPage.jsx`  
- [ ] **T10 – Write Playwright Visual‑Regression Test** – in `tests/e2e/hosts-login.page.spec.ts` (snapshot vs baseline)  
- [ ] **T11 – Run Full Verification Gates** – execute `make verify` (API tests, build, E2E)  
- [ ] **T12 – Update README with Login‑Flow Note** – add short description of new redirect behavior  
- [ ] **T13 – Atomic Git Commits (per task)** – `git add <files> && git commit -m "feat: <short desc>"`  
- [ ] **T14 – Push Feature Branch & Open PR** – `git push origin ui-redesign/hosts-auth-guard` then open PR for Dhar review  

---  

## 1️⃣ Context & Assumptions  

| Item | Detail |
|------|--------|
| **Repo location** | `/home/ubuntu/projects/vexa/02‑application/` |
| **Design‑system reference** | `docs/design/design-system.md` (tokens: `info #3B82F6`, radius `6px`, heading `display 32px`) |
| **Security rule** | AGENTS.md: `/hosts` **must** redirect to `/login` when no session cookie |
| **Current state** | Top‑bar nav (Hosts, Terminal, Tunnels, Vault, Settings) shown; `/hosts` loads unauthenticated; primary button uses lavender (`#A3B8FF`) instead of spec blue; heading “Welcome Back” ≈ 24‑28 px; missing security badges; no visible “Back to Dashboard” affordance |
| **Verification tools** | Playwright E2E suite, Jest/React testing library, `make verify` runs all gates (19/19) |

---  

## 2️⃣ Bite‑Sized Task Breakdown (aligned with **Plan → Dispatch → Verify**)  

| Phase | Task ID | Description | File / Path | Command / Action | Success Indicator |
|-------|---------|-------------|-------------|------------------|-------------------|
| **1 – Plan Capture** | **T1** | Create baseline assets folder & copy screenshots | `.hermes/plans/assets/` | `mkdir -p .hermes/plans/assets && cp /home/ubuntu/.hermes/cache/screenshots/*.png .hermes/plans/assets/` | Assets stored |
| **2 – Theme & Typography** | **T2** | Update primary color token to spec blue | `src/theme.css` | `sed -i 's/--primary-color:.*/--primary-color: #3B82F6;/' src/theme.css` | CSS token updated, linter passes |
| | **T3** | Adjust “Welcome Back” heading size & weight | `src/components/LoginCard.jsx` | Add `font-size: 32px; font-weight: 700;` | Heading matches `display 32px` spec |
| | **T4** | Fix primary button style (color + radius) | `src/components/PrimaryButton.module.css` | `background-color: var(--primary-color); border-radius: 6px;` | Button now blue, radius 6 px |
| **3 – Security Badges** | **T5** | Implement badge component (semantic colors) | `src/components/Badge.jsx` | Create component with `semanticColor={status}` | Badge renders correctly |
| | **T6** | Insert badge into host‑card markup | `src/pages/Hosts/HostCard.jsx` | `<Badge status={host.security} />` | Badge appears on each host card |
| **4 – Route Guard & Navigation** | **T7** | Add client‑side redirect when no cookie | `src/App.jsx` (or `src/routes/HostsRoute.jsx`) | `if (!document.cookie) return <Redirect to="/login" />;` | Unauthenticated users redirected |
| | **T8** | Add “Dashboard” link to top bar | `src/components/TopBar.jsx` | Insert `<Link to="/" className={styles.dashboard}>Dashboard</Link>` | Dashboard link visible |
| | **T9** | Add “Back‑to‑Dashboard” button on hosts page | `src/pages/Hosts/HostsPage.jsx` | `<Button onClick={backToDashboard}>← Dashboard</Button>` | Button visible |
| **5 – Visual Regression Test** | **T10** | Write Playwright test for auth guard & snapshot | `tests/e2e/hosts-login.page.spec.ts` | (see annex) – visit `/hosts`, assert redirect, screenshot vs baseline | Snapshot passes |
| **6 – Verification & Gates** | **T11** | Run full verification suite | — | `make verify` (runs API tests, build, E2E) | All 5 gates green |
| **7 – Documentation & Hygiene** | **T12** | Update README with login‑flow change | `README.md` | Add short note | Doc updated |
| | **T13** | Atomic Git commits per task | — | `git add <files> && git commit -m "feat: <short desc>"` | Commit created |
| | **T13** | Push feature branch & open PR for Dhar review | — | `git push origin ui-redesign/hosts-auth-guard` then open PR | PR created |

---  

## 3️⃣ File‑Change Summary  

| Path | Purpose |
|------|---------|
| `.hermes/plans/assets/hosts-baseline.png` | Baseline screenshot for regression |
| `src/theme.css` | Replace primary color with spec `#3B82F6` |
| `src/components/LoginCard.jsx` | Adjust heading typography |
| `src/components/PrimaryButton.module.css` | Update button color & radius |
| `src/components/Badge.jsx` | New badge component |
| `src/pages/Hosts/HostCard.jsx` | Insert badge |
| `src/App.jsx` / `src/routes/HostsRoute.jsx` | Add route‑guard redirect |
| `src/components/TopBar.jsx` | Add Dashboard link |
| `src/pages/Hosts/HostsPage.jsx` | Add Back‑to‑Dashboard button |
| `tests/e2e/hosts-login.page.spec.ts` | Visual regression & auth guard test |
| `README.md` | Document login‑flow change |

---  

## 4️⃣ Verification & Acceptance Criteria  

- **All UI changes** must use tokens from `design-system.md`.  
- **Route guard** must redirect to `/login` when `document.cookie` is empty.  
- **Playwright test** must pass; snapshot must match baseline in `.hermes/plans/assets/`.  
- **CI gate** (`make verify`) must report **19/19 passing**.  
- **Atomic commits** must follow conventional‑commit format (`feat: align hosts UI…`).  
- **PR** must receive Dhar’s explicit **👍** before merge.  

---  

## 5️⃣ Execution Handoff  

1. **T2 is the next task** – run the `sed` command to replace the primary‑color token.  
2. After each task:  
   - Run the verification command listed.  
   - If green, mark the checklist item ✅.  
   - Commit, push (when appropriate), and update the checklist.  
3. When all checkboxes are ✅, the feature is ready for final review.  

---  

*Plan saved at:*  
`/home/ubuntu/projects/vexa/06-temp/plans/2026-06-29_UI-Redux-2-Plan.md`  

Ready for **subagent‑driven** execution. Use `subagent-driven-development` to spin a fresh child agent for each `T#` task, pass the exact file path and verification command in the task context, then await spec‑compliance and code‑quality reviews before moving to the next task.  
