# Vexa UI Redesign — Master Plan (v2 — Revised Gap Analysis)

> **Status**: PROPOSED — awaiting Dhar's approval  
> **Date**: 2026-06-30  
> **Scope**: Frontend-only redesign of Vexa web app to match `/home/ubuntu/vexa-redesign/` design system  
> **Backend (Go API)**: Untouched. All changes are in `apps/web/`

---

## 1. Mockup Coverage Audit

### Mockups → Current Page Mapping

| # | Mockup File | Current Page | Status |
|---|-------------|-------------|--------|
| 1 | `hosts_dashboard.html` | `app/hosts/page.tsx` | ✅ Direct mapping |
| 2 | `terminal_view.html` | `app/terminal/page.tsx` | ✅ Direct mapping |
| 3 | `vault_credentials.html` | `app/vault/page.tsx` | ✅ Direct mapping |
| 4 | `settings_overview.html` | `app/settings/page.tsx` | ✅ Direct mapping |
| 5 | `profile_settings.html` | `app/settings/profile/page.tsx` | ✅ Direct mapping |
| 6 | `security_settings.html` | `app/settings/security/page.tsx` | ✅ Direct mapping |
| 7 | `api_keys_settings.html` | `app/settings/api-keys/page.tsx` | ✅ Direct mapping |
| 8 | `appearance_settings.html` | `app/settings/appearance/page.tsx` | ✅ Direct mapping |
| 9 | `appearance_settings_active_state.html` | Same page (interaction state) | ✅ State variant |
| 10 | `appearance_settings_updated.html` | Same page (updated state) | ✅ State variant |
| 11 | `notifications_settings.html` | **NO current page** | 🆕 New page needed |
| 12 | `notifications_settings_interaction_states.html` | Same (interaction state) | 🆕 State variant |

### Pages WITHOUT Mockups (26 pages — infer from design system)

| Page | Category | Redesign Strategy |
|------|----------|-------------------|
| `app/admin/page.tsx` | Admin | Follow dashboard pattern from hosts_dashboard |
| `app/admin/audit/page.tsx` | Admin | Data table + filter bar pattern |
| `app/admin/hosts/page.tsx` | Admin | Host list + admin controls |
| `app/admin/metrics/page.tsx` | Admin | Stats cards + charts layout |
| `app/admin/sessions/page.tsx` | Admin | Session list + recordings |
| `app/admin/users/page.tsx` | Admin | User table + role badges |
| `app/discovery/page.tsx` | Infrastructure | Network scanner UI (unique) |
| `app/files/page.tsx` | Infrastructure | File manager (unique) |
| `app/hosts/[id]/page.tsx` | Infrastructure | Host detail (follows from hosts_dashboard) |
| `app/tunnels/page.tsx` | Infrastructure | Tunnel list + stats cards |
| `app/vault/share/page.tsx` | Vault | Sharing sub-page (follows from vault) |
| `app/sessions/page.tsx` | Session | Session list |
| `app/sessions/recordings/page.tsx` | Session | Recording player |
| `app/settings/sessions/page.tsx` | Settings | Follow settings card pattern |
| `app/settings/webauthn/page.tsx` | Settings | Follow settings card pattern |
| `app/login/page.tsx` | Auth | Auth layout pattern |
| `app/register/page.tsx` | Auth | Auth layout pattern |
| `app/forgot-password/page.tsx` | Auth | Auth layout pattern |
| `app/reset-password/page.tsx` | Auth | Auth layout pattern |
| `app/mfa-verify/page.tsx` | Auth | Auth layout pattern |
| `app/verify-email/page.tsx` | Auth | Auth layout pattern |
| `app/terms/page.tsx` | Legal | Static content page |
| `app/welcome/page.tsx` | Onboarding | Welcome/landing page |
| `app/page.tsx` | Root | Redirect or landing |
| `app/profile/page.tsx` | Profile | Old profile (may merge into settings/profile) |
| `app/security/page.tsx` | Security | Old security (may merge into settings/security) |

**Summary**: 9 pages have direct mockups, 26 pages need inferred design from the design system tokens + patterns established in mockups.

---

## 2. Design Token Gap Analysis (Detailed)

### 2.1 Color System

| Token | Current (globals.css) | Target (Design.md/Mockup) | Gap |
|-------|----------------------|--------------------------|-----|
| `--background` | `222 47% 5%` (HSL) | `#10131a` (hex) | Value differs slightly |
| `--surface` | `222 47% 5%` (same as bg) | `#10131a` | Same, but needs hex |
| `--surface-dim` | `222 47% 5%` | `#10131a` | ✅ Match |
| `--surface-bright` | `222 20% 25%` | `#363941` | **Different** — current is bluer, target is grayer |
| `--surface-container` | `222 30% 15%` | `#1d2027` | **Different** — target is more neutral |
| `--surface-container-high` | `222 25% 20%` | `#272a31` | **Different** — target is more neutral |
| **NEW** `--surface-container-lowest` | ❌ Missing | `#0b0e15` | **Add** |
| **NEW** `--surface-container-low` | ❌ Missing | `#191b23` | **Add** |
| **NEW** `--surface-container-highest` | ❌ Missing | `#32353c` | **Add** |
| **NEW** `--surface-variant` | ❌ Missing | `#32353c` | **Add** |
| `--primary` | `#3B82F6` (blue-500) | `#adc6ff` (lighter blue) | **Major change** — softer, more refined |
| **NEW** `--primary-container` | ❌ Missing | `#4d8eff` | **Add** |
| **NEW** `--on-primary` | ❌ Missing | `#002e6a` | **Add** |
| **NEW** `--on-primary-container` | ❌ Missing | `#00285d` | **Add** |
| `--secondary` | `156 62% 59%` (green!) | `#b9c8de` (blue-gray) | **Major change** — green → blue-gray |
| **NEW** `--secondary-container` | ❌ Missing | `#39485a` | **Add** |
| **NEW** `--on-secondary` | ❌ Missing | `#233143` | **Add** |
| **NEW** `--on-secondary-container` | ❌ Missing | `#a7b6cc` | **Add** |
| **NEW** `--tertiary` | ❌ Missing | `#ffb786` (warm orange) | **Add** — accent color |
| **NEW** `--tertiary-container` | ❌ Missing | `#df7412` | **Add** |
| **NEW** `--on-tertiary` | ❌ Missing | `#502400` | **Add** |
| **NEW** `--on-tertiary-container` | ❌ Missing | `#461f00` | **Add** |
| `--error` | `4 100% 83%` | `#ffb4ab` | **Different** — needs hex |
| **NEW** `--error-container` | ❌ Missing | `#93000a` | **Add** |
| **NEW** `--on-error` | ❌ Missing | `#690005` | **Add** |
| **NEW** `--on-error-container` | ❌ Missing | `#ffdad6` | **Add** |
| `--outline` | `220 14% 65%` | `#8c909f` | **Different** |
| `--outline-variant` | `220 18% 28%` | `#424754` | **Different** — target is bluer |
| `--on-surface` | `226 71% 92%` | `#e1e2ec` | Close, needs hex |
| `--on-surface-variant` | `226 20% 80%` | `#c2c6d6` | Close, needs hex |
| **NEW** `--on-background` | ❌ Missing | `#e1e2ec` | **Add** |
| **NEW** `--inverse-surface` | ❌ Missing | `#e1e2ec` | **Add** |
| **NEW** `--inverse-on-surface` | ❌ Missing | `#2e3038` | **Add** |
| **NEW** `--inverse-primary` | ❌ Missing | `#005ac2` | **Add** |
| `--border` | `220 18% 28%` | Same as `--outline-variant` | Needs alignment |
| `--muted-foreground` | `220 14% 65%` | Same as `--on-surface-variant` | Needs alignment |
| **NEW** `--success` | ❌ Missing | Implied (green) | **Add** — used in mockups |
| **NEW** `--warning` | ❌ Missing | Implied (yellow) | **Add** — used in mockups |

**Gap count**: ~25 tokens missing or different. Current system has ~20 CSS vars; target needs ~45+.

### 2.2 Tailwind Config Gap

| Area | Current | Target | Gap |
|------|---------|--------|-----|
| **Color format** | HSL via `hsl(var(--token))` | Direct hex values | **Approach decision needed** |
| **Surface levels** | 5 levels (dim, DEFAULT, bright, container, container-high) | 7 levels (dim, DEFAULT, bright, container-lowest/low/DEFAULT/high/highest) | **Missing 3 levels** + rename |
| **Primary/Secondary** | shadcn standard (primary, secondary, muted, accent, destructive) | Material 3 (primary, secondary, tertiary + containers + on-variants) | **Major restructure** |
| **Typography scale** | Not defined in Tailwind config | 8 type tokens (headline-lg/md/sm, body-lg/md, label-md, mono-code, headline-lg-mobile) | **Missing entirely** |
| **Spacing scale** | Tailwind default | Custom: base/xs/sm/md/lg/xl/gutter/margin | **Missing** |
| **Border radius** | `var(--radius)` based | Custom: DEFAULT/lg/xl/full | **Needs update** |
| **Font families** | Not in config (set in globals.css) | 8 font family tokens in config | **Missing** |

### 2.3 Layout Architecture Gap

| Element | Current | Target | Gap |
|---------|---------|--------|-----|
| **Top nav** | Logo + 5 nav links + (missing: search, notifications, avatar) | Logo + nav links + search bar + notification bell + help icon + user avatar | **Missing 3 elements** |
| **Sidebar** | ❌ None | 240px fixed sidebar with user context, primary action, section nav, footer | **Entire component missing** |
| **Content area** | `container py-8` (centered, max-width) | Full-width with left padding from sidebar | **Layout change** |
| **Footer** | ❌ None | Stats bar (Total Hosts, Active Sessions, version) | **Missing** |
| **Page wrapper** | `<DashboardLayout>` wraps with `<TopNav>` + centered `<main>` | `<DashboardLayout>` wraps with `<Sidebar>` + `<TopNav>` + full-width `<main>` | **Restructure** |

### 2.4 Icon System Gap

| Aspect | Current | Target | Gap |
|--------|---------|--------|-----|
| **Library** | Lucide React (`lucide-react`) | Material Symbols Outlined (Google Fonts) | **Full swap** |
| **Usage** | `<Icon className="w-5 h-5" />` | `<span className="material-symbols-outlined">icon_name</span>` | **API change** |
| **Icons found in mockups** | N/A | 40+ unique icons used | **All need mapping** |
| **Bundling** | npm package (tree-shakeable) | Google Fonts CDN link | **Loading strategy change** |

**Material Symbols identified in mockups**: `add`, `add_circle`, `admin_panel_settings`, `cloud`, `cloud_upload`, `code`, `contact_support`, `content_copy`, `delete_forever`, `description`, `dns`, `encrypted`, `extension`, `grid_view`, `history`, `key`, `laptop_mac`, `lock_open`, `lock`, `more_vert`, `palette`, `password`, `person`, `power`, `report`, `search`, `security`, `settings`, `settings_suggest`, `shield`, `terminal`, `text_fields`, `verified_user`, `view_list`, `view_quilt`, `visibility`, `warning`

**Lucide → Material Symbols mapping needed for**: All existing `lucide-react` imports across ~30+ component files.

### 2.5 Component Restyle Gap

| Component | Current Style | Target Style | Effort |
|-----------|--------------|-------------|--------|
| `Card` | `bg-card` + `border-border` | `bg-surface-container` + `border-outline-variant` + refined padding | Low |
| `Button` | `bg-primary` / `bg-secondary` / `bg-destructive` | `bg-primary` / `bg-primary-container` / `bg-error` + on-colors | Medium |
| `Input` | `bg-transparent` + `border-input` | `bg-surface-container-lowest` + `border-outline-variant` + focus ring `ring-primary` | Low |
| `Badge` | `bg-primary/10` + `text-primary` | Surface-variant backgrounds + on-surface-variant text + status colors | Medium |
| `Dialog` | `bg-popover` | `bg-surface-container-high` + `border-outline-variant` | Low |
| `Switch` | shadcn default toggle | Material-style with primary fill when on | Medium |
| `Tabs` | shadcn default (box-style) | Underline indicator with `border-primary` | Medium |
| `Table` | Basic table styles | `bg-surface-container` header + `border-outline-variant` rows | Low |
| `Tooltip` | `bg-popover` | `bg-surface-container-highest` | Low |
| `Alert` | `bg-destructive/10` etc. | `bg-error-container/5` + `border-error-container/20` + `text-error` | Medium |
| `Avatar` | Basic circular | Circular + `ring-2 ring-outline-variant` | Low |
| `Checkbox` | shadcn default | Material-style: `bg-primary` checked + `border-outline-variant` unchecked | Medium |
| `RadioGroup` | shadcn default | Material-style with `border-primary` selected ring | Medium |
| `Select` | shadcn default | `bg-surface-container-lowest` + refined dropdown | Medium |
| `Skeleton` | `bg-muted` | `bg-surface-container-low` with pulse | Low |
| `Progress` | shadcn default | `bg-surface-container` track + `bg-primary` fill | Low |
| `Slider` | shadcn default | Material-style with `bg-primary` thumb | Medium |
| `ScrollArea` | shadcn default | Refined scrollbar with `bg-surface-container-high` | Low |
| `Sonner/Toast` | shadcn default | `bg-surface-container-highest` + `border-outline-variant` | Low |
| `DropdownMenu` | `bg-popover` | `bg-surface-container-high` + `border-outline-variant` | Low |
| `ContextMenu` | `bg-popover` | Same as dropdown | Low |
| `Label` | Basic text | `text-on-surface-variant` + `font-medium` | Low |
| `Separator` | `bg-border` | `bg-outline-variant` | Low |
| `Textarea` | Same as Input | Same as Input | Low |
| `SecurityBadge` | Custom | Refine with new tokens | Low |

### 2.6 Page-Level Gap (Pages WITH Mockups)

| Page | Key Changes Needed |
|------|-------------------|
| **Hosts Dashboard** | Add sidebar filters, card grid, footer stats, search/toolbar, HostCard redesign |
| **Terminal** | Tabbed session bar, host dropdown + Connect button, terminal area bg refinement, toolbar |
| **Vault** | Locked vault state (master password modal), credential card grid, refined icon usage |
| **Settings Overview** | Categorized layout: Core Account (Profile, Security), Developer Access (API Keys, WebAuthn), System (Appearance, Notifications, Sessions) |
| **Settings/Profile** | Form layout with card sections, avatar upload, identity fields |
| **Settings/Security** | MFA section, login history, password audit, session management |
| **Settings/API Keys** | Key list table, create key modal, copy/revoke actions |
| **Settings/Appearance** | Theme toggle (active/updated states), density selector, font preview |
| **Settings/Notifications** | **NEW PAGE** — toggle switches, quiet hours, Slack integration, event triggers |

### 2.7 Page-Level Gap (Pages WITHOUT Mockups — 26 pages)

| Category | Pages | Strategy |
|----------|-------|----------|
| **Auth** (6 pages) | login, register, forgot-password, reset-password, mfa-verify, verify-email | Unified AuthLayout with branding, centered card, Material Symbols |
| **Admin** (6 pages) | dashboard, audit, hosts, metrics, sessions, users | Follow hosts_dashboard pattern: stats cards + data tables + filters |
| **Infrastructure** (5 pages) | discovery, files, hosts/[id], tunnels, vault/share | Follow respective parent page patterns |
| **Session** (2 pages) | sessions, sessions/recordings | Session list + recording player with new tokens |
| **Settings sub** (2 pages) | settings/sessions, settings/webauthn | Follow settings card pattern from mockups |
| **Legacy** (2 pages) | profile, security | Merge into settings/profile and settings/security, redirect old routes |
| **Static** (2 pages) | terms, welcome | Content pages with new typography + tokens |
| **Root** (1 page) | page.tsx (root) | Redirect to /hosts or /welcome |

---

## 3. Architecture Decisions Needed

### 3.1 Color Token Format: HSL vs Hex

**Current**: CSS vars in HSL (`222 47% 5%`), referenced as `hsl(var(--token))` in Tailwind.
**Mockup**: Direct hex values in Tailwind config (`"surface": "#10131a"`).

| Option | Pros | Cons |
|--------|------|------|
| **A: Keep HSL, add missing tokens** | shadcn/ui compatibility, opacity modifiers work | Complex to maintain, conversion needed |
| **B: Switch to hex (like mockup)** | Matches Design.md exactly, simpler | Breaks `hsl(var())` pattern, shadcn needs update |
| **C: Hybrid — hex in Tailwind, HSL for shadcn** | Best of both | Potential inconsistency, more work |

**Recommendation**: **Option B** (hex). The mockup uses hex directly in Tailwind config. shadcn/ui can work with hex — just change the Tailwind color values from `hsl(var(--token))` to direct hex. Keep CSS vars in globals.css for documentation, but Tailwind config uses hex directly.

### 3.2 Icon System: Lucide vs Material Symbols

**Current**: `lucide-react` npm package (~30+ files import it).
**Mockup**: Material Symbols Outlined via Google Fonts CDN.

| Option | Pros | Cons |
|--------|------|------|
| **A: Full swap to Material Symbols** | Matches mockup exactly, lighter bundle | ~30+ files to update, CDN dependency |
| **B: Keep Lucide, restyle** | No API change, tree-shakeable | Doesn't match mockup visual style |
| **C: Hybrid** | Pragmatic | Inconsistency, complexity |

**Recommendation**: **Option A** (full swap). Create a `<MaterialIcon name="search" className="..." />` wrapper component for ergonomics, then replace all Lucide imports. Add Material Symbols font link in `layout.tsx`.

### 3.3 shadcn/ui Compatibility

shadcn/ui components use specific CSS variable conventions: `--background`, `--foreground`, `--card`, `--primary`, etc. The redesign adds Material 3 tokens that don't map 1:1.

**Strategy**: Map shadcn tokens to Material 3 equivalents:
- `--background` → `--surface-dim` (or keep as alias)
- `--foreground` → `--on-surface` (or keep as alias)
- `--card` → `--surface-container`
- `--primary` → `--primary` (update value)
- `--secondary` → `--secondary` (update value)
- `--muted` → `--surface-container-low`
- `--muted-foreground` → `--on-surface-variant`
- `--border` → `--outline-variant`
- `--input` → `--outline-variant`
- `--ring` → `--primary`
- `--destructive` → `--error`
- `--accent` → `--tertiary` (or `--primary-container`)

This way shadcn components automatically get the new colors when we update the CSS variables.

---

## 4. Revised Phased Implementation

### Phase 1: Design Token Foundation
**Files**: `globals.css`, `tailwind.config.ts`, `layout.tsx`, new `components/ui/material-icon.tsx`
**Tasks**:
1. Replace CSS variables in `globals.css` with full Material 3 token set (~45 vars)
2. Map shadcn tokens to Material 3 equivalents (aliases)
3. Update `tailwind.config.ts` with hex colors, typography scale, spacing, radii
4. Add Material Symbols font to `layout.tsx`
5. Create `<MaterialIcon>` wrapper component
6. Verify: `npm run build` passes, visual sanity check

### Phase 2: Layout Shell
**Files**: `DashboardLayout.tsx`, `TopNav.tsx`, new `Sidebar.tsx`, new `Footer.tsx`
**Tasks**:
1. Create `Sidebar` component (user context, primary action, nav sections, footer)
2. Redesign `TopNav` (add search bar, notification bell, help icon, user avatar)
3. Restructure `DashboardLayout` (sidebar + topnav + full-width main)
4. Create `Footer` stats bar component
5. Verify: all pages still render (layout is wrapper, pages don't change yet)

### Phase 3: Component Library Restyle
**Files**: All 28 files in `components/ui/`
**Tasks**:
1. Card — surface-container bg, outline-variant border
2. Button — primary/secondary/ghost/error variants
3. Input/Textarea/Label — surface-container-lowest bg, outline-variant border
4. Badge — status colors (success/warning/error/info)
5. Dialog — surface-container-high bg
6. Switch — Material-style toggle
7. Tabs — underline indicator
8. Table — surface hierarchy
9. Tooltip — surface-container-highest bg
10. Alert — error/warning/success containers
11. Avatar — ring styling
12. Checkbox/RadioGroup — Material-style
13. Select/DropdownMenu/ContextMenu — surface-container-high
14. Skeleton/Progress/Slider/ScrollArea — surface hierarchy
15. Sonner/Toast — surface-container-highest
16. Verify: `npm run build` + `npm run test`

### Phase 4: Page Redesigns (Mockup Pages)
**Files**: 9 pages with mockups + 1 new page
**Tasks** (per page, each as separate dispatch task):
1. Hosts Dashboard — sidebar filters, card grid, footer stats
2. Terminal — tabbed sessions, connection dropdown
3. Vault — locked state, credential cards
4. Settings Overview — categorized sections
5. Settings/Profile — form sections
6. Settings/Security — MFA, history, audit
7. Settings/API Keys — key table, create modal
8. Settings/Appearance — theme toggle, density selector (with active/updated states)
9. Settings/Notifications — **NEW PAGE** — toggles, quiet hours, integrations

### Phase 5: Page Redesigns (Inferred Pages)
**Files**: 26 pages without mockups
**Tasks** (grouped by category):
1. Auth pages (6) — unified AuthLayout, branding, Material Symbols
2. Admin pages (6) — stats cards, data tables, filters
3. Infrastructure pages (5) — discovery, files, host detail, tunnels, vault/share
4. Session pages (2) — session list, recording player
5. Settings sub-pages (2) — sessions, webauthn
6. Legacy redirects (2) — profile → settings/profile, security → settings/security
7. Static pages (2) — terms, welcome
8. Root redirect (1) — page.tsx → /hosts

### Phase 6: Polish & QA
1. Responsive design — mobile breakpoints, collapsible sidebar
2. Accessibility — focus states, ARIA labels, contrast ratios
3. Loading/empty states — skeletons, no-data states
4. Test updates — fix tests broken by layout/structure changes
5. Final visual QA — compare against all 12 mockup screenshots

---

## 5. Execution Strategy

- **Tool**: Claude Code dispatches (Plan → Claude Code → Verify → Push)
- **Max errors per dispatch**: 30
- **Per-task commit**: Each task gets its own commit
- **Branch**: `feat/ui-redesign` from `main`
- **Verification**: `npm run build` + `npm run test` after each phase
- **Generated files**: Restore `.next/`, `tsconfig.tsbuildinfo` after dispatches
- **Estimated total dispatches**: 10-15 (across all phases)

---

## 6. Risk Assessment

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Token migration breaks shadcn components | High | Medium | Map shadcn tokens as aliases, test after Phase 1 |
| Material Symbols CDN dependency | Medium | Low | Self-host font files as fallback |
| 26 pages without mockups look inconsistent | Medium | Medium | Establish patterns in Phase 4, apply consistently in Phase 5 |
| Test failures from layout restructure | Medium | High | Update test setup to mock Sidebar, fix selectors |
| Large scope = long timeline | Medium | High | Phased approach, each phase independently shippable |
| HSL → hex migration breaks opacity modifiers | High | Low | Test `bg-primary/10` patterns early in Phase 1 |
| Icon swap misses some files | Low | Medium | Grep for all `lucide-react` imports, verify count |