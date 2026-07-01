# Vexa UI Redesign — Per-Page Implementation Plan

> **Status**: PROPOSED — awaiting Dhar's approval  
> **Date**: 2026-06-30  
> **Scope**: All 35 current pages + 1 new page = 36 pages total  
> **Reference**: `/home/ubuntu/vexa-redesign/` (Design.md + 12 HTML mockups + images)  
> **Backend**: Untouched — frontend only (`apps/web/`)

---

## Design Patterns Extracted from Mockups

### Pattern A: Dashboard Page (from hosts_dashboard.html)
```
[TopNav: Logo + Nav + Notifications/Help/Avatar]
[Sidebar: UserContext + NewConnection + NavSection + Footer(Help/Logout)]
[Main:
  Header (H1 headline-lg + subtitle body-md + action buttons)
  FilterBar (search input + filter pills + filter button)
  CardGrid (glass-card, 1-4 cols responsive)
  FooterStats (total count + active count + version)
]
```

### Pattern B: Terminal Page (from terminal_view.html)
```
[TopNav: same shell]
[Sidebar: same shell, "Recent Sessions" active]
[Main:
  TabBar (h-14, surface-container bg, tabs with border-b-2 primary active)
  TerminalPane (bg-black, mono-code font, colored prompt)
  ControlsBar (host dropdown + Connect button)
]
```

### Pattern C: Vault Page (from vault_credentials.html)
```
[TopNav: same shell]
[Sidebar: same shell, "Key Management" active]
[Main:
  Header (H1 "Secure Vault" + lock icon + subtitle)
  LockModal (centered card: shield icon, "Vault Locked", password input, unlock button)
  OR CredentialGrid (credential cards with type icon, name, metadata, actions)
]
```

### Pattern D: Settings Overview (from settings_overview.html)
```
[TopNav: same shell, "Settings" active in top nav]
[Sidebar: same shell, settings sub-nav active]
[Main:
  Header (H1 "Settings" + subtitle)
  SettingsNav (sidebar within main: Profile/Security/API Keys/WebAuthn/Appearance/Notifications/Sessions)
  CategoryCards (grid of category sections, each with icon + title + description → link)
]
```

### Pattern E: Settings Sub-Page (from profile/security/api_keys/appearance/notifications)
```
[TopNav: same shell, "Settings" active]
[Sidebar: same shell + settings sub-nav with active item highlighted]
[Main:
  BackLink + H1 (headline-lg) + subtitle
  Card sections (surface-container bg, rounded-xl, p-6)
    - Each section: icon header + form fields or toggles
    - Form fields: bg-surface-container-lowest, border-outline-variant
    - Toggles: Material-style checkbox or slider
  SaveBar (sticky bottom or inline button)
]
```

### Pattern F: Auth Page (inferred from auth component structure)
```
[No TopNav/Sidebar — full-screen centered]
[SplitLayout:
  Left: Brand panel (bg-surface-container, logo, tagline, illustration)
  Right: Form panel (bg-background, centered card)
    - Logo + title (headline-lg)
    - Form fields (same input pattern)
    - Primary button (bg-primary text-on-primary)
    - Links to other auth pages
]
```

---

## Shared Shell Components (Phase 2 — all pages use these)

### TopNav (all pages except auth)
```html
<header class="bg-background border-b border-outline-variant h-16 flex items-center justify-between px-6 fixed top-0 left-0 right-0 z-50">
  Left: Logo (w-8 h-8 bg-primary rounded-lg + "terminal" icon) + "Vexa" text-headline-sm
  Center: Nav links (Hosts/Terminal/Tunnels/Vault/Settings) — active: text-primary border-b-2 border-primary
  Right: notifications icon | help icon | avatar (w-8 h-8 rounded-full bg-secondary-container border-outline-variant)
</header>
```

### Sidebar (all pages except auth)
```html
<aside class="fixed left-0 top-16 bottom-0 w-[260px] bg-surface-container border-r border-outline-variant hidden md:flex flex-col p-4 gap-2">
  UserContext: w-10 h-10 rounded-lg bg-surface-container-high border-outline-variant + "admin_panel_settings" icon | "System Admin" + "Production Cluster"
  PrimaryAction: bg-primary-container text-on-primary-container rounded-lg h-10 — "add" icon + "New Connection"
  NavSection: flex-1, gap-1 — each item: px-3 py-2 rounded-lg, active: text-primary bg-secondary-container border-l-2 border-primary
  Footer: border-t pt-4 — "Help Center" (help_outline) + "Log out" (logout, text-error)
</aside>
```

### Page Wrapper
```html
<div class="flex pt-16 h-screen">
  <Sidebar />
  <main class="flex-1 md:ml-[260px] p-8 overflow-y-auto bg-background h-full">
    {page content}
  </main>
</div>
```

---

## Per-Page Plan: 36 Pages

### === PAGES WITH MOCKUPS (9 pages + 1 new) ===

---

### Page 1: `/hosts` — Hosts Dashboard
**Mockup**: `hosts_dashboard.html` ✅
**Current**: `app/hosts/page.tsx` → wraps `<DashboardLayout><HostList/></DashboardLayout>`

**Layout**: Pattern A (Dashboard)
**Changes**:
1. Page header: `h1.text-headline-lg` "Hosts" + `p.text-body-md.text-on-surface-variant` "Securely manage remote environments across SSH, RDP, VNC."
2. Action buttons: grid_view toggle (border-outline-variant) + "Add Host" (bg-primary text-on-primary)
3. Filter bar: search input (bg-surface-container-low border-outline-variant) + filter pills (All/SSH/RDP/VNC in bg-surface-container) + Filters button
4. Card grid: `grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6`
5. HostCard redesign: `glass-card rounded-xl p-5` with:
   - Icon (12x12, bg-primary/10 or bg-tertiary-container/10 based on type)
   - Status badge (Online: green, Offline: red, Busy: yellow — `bg-{color}-500/10 text-{color}-400`)
   - Hostname (font-headline-sm) + IP (font-mono-code text-on-surface-variant/70)
   - Tags (px-2 py-0.5 bg-surface-variant text-[11px] rounded-md)
   - Footer: border-t border-outline-variant + "Connect" link (text-primary)
   - Offline cards: `opacity-75 grayscale-[0.5]`, Connect button disabled
6. Add placeholder card: border-2 border-dashed border-outline-variant, hover:border-primary
7. Footer stats: border-t border-outline-variant + "Total Hosts: 24" + "Active Sessions: 3" + version

**Components touched**: `HostList.tsx`, `HostCard.tsx`, `HostStatsCard.tsx`, `HostForm.tsx`, `HostImportExport.tsx`
**Icons**: dns, terminal, desktop_windows, screen_share, add, grid_view, filter_list, search, folder, cloud, history, vpn_key

---

### Page 2: `/terminal` — Terminal View
**Mockup**: `terminal_view.html` ✅
**Current**: `app/terminal/page.tsx`

**Layout**: Pattern B (Terminal)
**Changes**:
1. Tab bar: `h-14 bg-surface-container border-b border-outline-variant`
   - Active tab: `border-b-2 border-primary bg-surface-container-high` + icon + label + close button (hover)
   - Inactive tab: `border-b-2 border-transparent hover:bg-surface-variant`
   - "+" button: `w-8 h-8 rounded hover:bg-surface-variant`
2. Controls bar (right side of tab bar): host dropdown (`bg-surface-container-highest border-outline-variant`) + "Connect" button (`bg-primary-container text-on-primary-container`)
3. Terminal pane: `bg-black p-4 font-mono-code`
   - Status overlay: absolute top-6 left-6, `bg-surface-container-lowest/80 border-outline-variant`, pulse dot
   - Terminal output: `text-on-surface-variant` for regular, `text-primary` for prompt highlights, `text-tertiary` for user@host, `text-primary-container` for path
   - SSH output blocks: `pl-4 border-l border-outline-variant/30 bg-surface-variant/10`
4. No footer stats bar (terminal uses full height)

**Components touched**: `terminal.tsx`, `terminal-tabs.tsx`, `terminal-pane.tsx`, `terminal-toolbar.tsx`
**Icons**: terminal, computer, dns, add, close, power, expand_more

---

### Page 3: `/vault` — Secure Vault
**Mockup**: `vault_credentials.html` ✅
**Current**: `app/vault/page.tsx`

**Layout**: Pattern C (Vault)
**Changes**:
1. Page header: `h1.text-headline-lg` "Secure Vault" + "encrypted" icon (text-[40px]) + subtitle "Manage organize encrypted credentials private SSH keys."
2. Toolbar: "Add Credential" button (bg-primary text-on-primary) + import/export buttons (border-outline-variant)
3. Locked state: centered card `max-w-md mx-auto mt-20`:
   - Shield icon in circle (w-20 h-20 bg-surface-container-high rounded-full)
   - "Vault Locked" title + "Your credentials protected AES-256 encryption..."
   - Password input (bg-surface-container-lowest border-outline-variant) + "Unlock" button (bg-primary)
4. Unlocked state: credential card grid (same glass-card pattern as hosts):
   - Type icon (SSH key, password, API key — each with different bg tint)
   - Credential name + metadata (type, last used, tags)
   - Actions: copy, view, delete (icon buttons, hover reveal)
   - Footer: border-t + "Last modified" date

**Components touched**: `vault/credential-card.tsx`, `vault/credential-form.tsx`, `vault/credential-list.tsx`, `vault/import-export.tsx`
**Icons**: encrypted, lock_open, key, password, content_copy, visibility, delete_forever, shield, add

---

### Page 4: `/vault/share` — Vault Share
**Mockup**: None — inferred from vault pattern
**Current**: `app/vault/share/page.tsx`

**Layout**: Pattern C variant (share sub-page)
**Changes**:
1. Back link: "← Back to Vault" (text-on-surface-variant hover:text-on-surface)
2. H1 "Share Credential" + subtitle
3. Share form card: bg-surface-container rounded-xl p-6:
   - Credential selector dropdown
   - Recipient email/username input
   - Expiry selector (1h, 24h, 7d, never)
   - Permission level (view-only, full-access)
   - "Generate Share Link" button (bg-primary)
4. Active shares table: bg-surface-container rounded-xl, headers with bg-surface-container-high
   - Columns: Recipient, Credential, Expires, Status, Actions (revoke)
5. Empty state: "No active shares" with share icon

**Components touched**: `vault/share` page + `vault/credential-list.tsx` for reference
**Icons**: share, link, content_copy, delete_forever, schedule, lock

---

### Page 5: `/settings` — Settings Overview
**Mockup**: `settings_overview.html` ✅
**Current**: `app/settings/page.tsx`

**Layout**: Pattern D (Settings Overview)
**Changes**:
1. H1 "Settings" (text-headline-lg) + subtitle "Manage account configurations, security preferences, system behavior."
2. Category sections (3 groups):
   - **Core Account**: Profile (person icon), Security (security icon)
   - **Developer Access**: API Keys (key icon), WebAuthn (verified_user icon)
   - **System**: Appearance (palette icon), Notifications (notifications icon), Sessions (history icon)
3. Each category card: `bg-surface-container border border-outline-variant rounded-xl p-5 hover:border-primary transition-all`
   - Left: icon in `w-12 h-12 bg-surface-container-high rounded-xl`
   - Right: title (font-headline-sm) + description (text-body-md text-on-surface-variant)
   - Click → navigate to sub-page
   - Hover: `hover:border-primary` + slight `translateY(-1px)`
4. Sidebar sub-nav visible (Profile/Security/API Keys/WebAuthn/Appearance/Notifications/Sessions)

**Components touched**: `app/settings/page.tsx` + new `SettingsSidebar` sub-component
**Icons**: person, security, key, verified_user, palette, notifications, history, settings_suggest

---

### Page 6: `/settings/profile` — Profile Settings
**Mockup**: `profile_settings.html` ✅
**Current**: `app/settings/profile/page.tsx`

**Layout**: Pattern E (Settings Sub-Page)
**Changes**:
1. Back link + H1 "Profile Settings" + subtitle
2. Avatar section card: `bg-surface-container rounded-xl p-6`:
   - Avatar (w-20 h-20 rounded-full bg-secondary-container border-2 border-outline-variant)
   - Upload button + Remove button
3. Identity form card: `bg-surface-container rounded-xl p-6`:
   - Full name, email, bio fields — `bg-surface-container-lowest border border-outline-variant rounded-lg py-2.5 px-4`
   - Focus: `focus:border-primary focus:ring-1 focus:ring-primary`
   - Labels: `text-label-md text-on-surface-variant`
4. Public profile card: visibility toggle, public URL field
5. Save button: `bg-primary text-on-primary` + Cancel: `border-outline-variant`

**Components touched**: `ProfileForm.tsx`, `components/profile/*`
**Icons**: person, upload, content_copy, visibility, lock

---

### Page 7: `/settings/security` — Security Settings
**Mockup**: `security_settings.html` ✅
**Current**: `app/settings/security/page.tsx`

**Layout**: Pattern E (Settings Sub-Page)
**Changes**:
1. H1 "Security Settings" + subtitle
2. MFA card: `bg-surface-container rounded-xl p-6`:
   - Toggle: Material-style switch (checked: bg-primary)
   - MFA methods list (Authenticator app, Security key, Backup codes)
   - Each method row: icon + name + status badge + "Configure" link
3. Password card: current password + new password + confirm fields
   - Strength indicator bar (bg-surface-container → bg-primary for strong)
4. Login history card: `bg-surface-container rounded-xl`:
   - Table: timestamp, IP, device, location, status (success/denied badge)
   - `border-outline-variant` row borders
5. Active sessions card: session list with "Revoke" buttons
6. Security audit card: checklist with check/warning icons

**Components touched**: `TwoFASetup.tsx`, `BiometricSettings.tsx`, `LoginHistory.tsx`, `ChangePasswordForm.tsx`
**Icons**: security, lock, verified_user, shield, password, history, warning, check_circle, cancel, devices

---

### Page 8: `/settings/api-keys` — API Keys
**Mockup**: `api_keys_settings.html` ✅
**Current**: `app/settings/api-keys/page.tsx`

**Layout**: Pattern E (Settings Sub-Page)
**Changes**:
1. H1 "API Keys" + subtitle "Manage programmatic access third-party integrations."
2. "Create New Key" button: `bg-primary text-on-primary`
3. Keys table: `bg-surface-container rounded-xl border border-outline-variant`:
   - Headers: `bg-surface-container-high text-label-md text-on-surface-variant border-b border-outline-variant`
   - Columns: Name, Key (masked, mono-code font), Created, Last Used, Scopes, Status, Actions
   - Row: `border-b border-outline-variant hover:bg-surface-variant/30`
   - Actions: copy, rotate (refresh icon), revoke (delete_forever text-error)
4. Create key modal: `bg-surface-container-high border border-outline-variant rounded-xl`:
   - Name input, scope checkboxes, expiry dropdown
   - Generated key display (mono-code, copy button) — shown once
5. Empty state: "No API keys" with key icon + "Create your first key" button

**Components touched**: `app/settings/api-keys/page.tsx` + API client functions
**Icons**: key, content_copy, refresh, delete_forever, add, lock, schedule, code

---

### Page 9: `/settings/appearance` — Appearance Settings
**Mockup**: `appearance_settings.html` + `appearance_settings_active_state.html` + `appearance_settings_updated.html` ✅
**Current**: `app/settings/appearance/page.tsx`

**Layout**: Pattern E (Settings Sub-Page) with interactive states
**Changes**:
1. H1 "Appearance" + subtitle
2. Theme card: `bg-surface-container rounded-xl p-6`:
   - Theme options: Dark (active), Light, System — rendered as clickable preview cards
   - Active: `border-2 border-primary bg-primary/5` with check icon
   - Inactive: `border border-outline-variant hover:border-outline`
   - Each: mini preview rectangle (showing color scheme)
3. Density card: Comfortable / Compact — radio-style selection
   - Selected: `bg-secondary-container text-on-secondary-container`
4. Font size card: slider (Material-style, bg-primary thumb) + preview text
5. Accent color card: color swatches (primary, tertiary, custom) — clickable circles
6. "Updated" toast state: `bg-surface-container-highest border-outline-variant` — "Appearance updated" with check icon

**Components touched**: `app/settings/appearance/page.tsx`
**Icons**: palette, check, text_fields, view_quilt, dark_mode, light_mode, settings_suggest

---

### Page 10: `/settings/notifications` — Notifications Settings (NEW PAGE)
**Mockup**: `notifications_settings.html` + `notifications_settings_interaction_states.html` ✅
**Current**: Does not exist → **CREATE NEW**

**Layout**: Pattern E (Settings Sub-Page)
**Changes**:
1. Create `app/settings/notifications/page.tsx`
2. H1 "Admin Console" or "Notifications" + subtitle
3. Notification channels card: `bg-surface-container rounded-xl p-6`:
   - Desktop Notifications: Material checkbox toggle (checked: bg-primary white checkmark)
   - Email Alerts: same toggle style
   - Quiet Hours: toggle + time pickers (FROM/TO, `bg-surface-container-lowest border-outline-variant`)
   - Focus: time inputs get `border-primary` when focused
4. Event triggers card:
   - SSH login alerts, Host down alerts, Credential access, Tunnel status changes
   - Each: toggle switch + description
5. Integrations card:
   - Slack Integration: slider-style toggle (off: gray, on: primary)
   - "Connect Workspace" button: `bg-tertiary text-on-tertiary` (purple accent)
6. Recent activity card:
   - Timeline list: icon + event + timestamp
   - "Clear all" text link

**Components touched**: NEW page + new toggle components if needed
**Icons**: notifications, email, schedule, toggle_on, toggle_off, event, clear_all, forum

---

### === PAGES WITHOUT MOCKUPS — INFERRED (26 pages) ===

---

### Page 11: `/` — Root Page
**Mockup**: None
**Current**: `app/page.tsx`

**Strategy**: Redirect to `/hosts` (authenticated) or `/login` (unauthenticated)
**Changes**: Replace page content with `redirect('/hosts')` or conditional redirect

---

### Page 12: `/welcome` — Welcome/Onboarding
**Mockup**: None — inferred from auth pattern
**Current**: `app/welcome/page.tsx`

**Layout**: Full-screen centered, no sidebar
**Changes**:
1. Centered card: `max-w-2xl mx-auto mt-20`:
   - Logo (large, w-16 h-16 bg-primary rounded-2xl + terminal icon)
   - "Welcome to Vexa" (text-headline-lg)
   - Subtitle: "Complete SSH access management..."
   - Getting started checklist (4 items): Add first host, Configure SSH key, Start session, Explore vault
   - Each item: icon circle + text + "Skip" / "Start" button
2. "Get Started" button: `bg-primary text-on-primary`
3. "Skip for now" link: text-on-surface-variant

**Icons**: terminal, dns, key, play_circle, shield, arrow_forward

---

### Page 13: `/login` — Login
**Mockup**: None — inferred (Pattern F: Auth)
**Current**: `app/login/page.tsx`

**Layout**: Pattern F (Auth — split screen)
**Changes**:
1. Left panel (hidden md:flex): `bg-surface-container flex-col justify-center p-12`:
   - Logo + "Vexa" headline-lg
   - Tagline: "Self-hosted SSH access management..."
   - Feature highlights: 3 items with icons (terminal, shield, dns)
   - Footer: version info
2. Right panel: `bg-background flex items-center justify-center p-8`:
   - Card: `w-full max-w-md bg-surface-container border border-outline-variant rounded-xl p-8`
   - H2 "Sign In" (text-headline-sm)
   - Email + password fields (bg-surface-container-lowest border-outline-variant)
   - "Remember me" checkbox (Material style)
   - "Sign In" button: `w-full bg-primary text-on-primary h-11 rounded-lg`
   - "Forgot password?" link (text-primary)
   - "Don't have an account? Register" link
3. No top nav or sidebar

**Components touched**: `LoginForm.tsx`, `AuthLayout.tsx`
**Icons**: terminal, shield, dns, email, lock, login

---

### Page 14: `/register` — Register
**Mockup**: None — inferred (Pattern F: Auth)
**Current**: `app/register/page.tsx`

**Layout**: Pattern F (Auth — split screen)
**Changes**: Same as login with registration fields:
1. Left panel: same branding
2. Right panel:
   - H2 "Create Account"
   - Name, email, password, confirm password fields
   - Password strength indicator
   - Terms checkbox + "I agree to Terms" link
   - "Create Account" button
   - "Already have an account? Sign In" link

**Components touched**: `RegisterForm.tsx`, `AuthLayout.tsx`
**Icons**: person_add, email, lock, check, terminal, shield

---

### Page 15: `/forgot-password` — Forgot Password
**Mockup**: None — inferred (Pattern F: Auth)
**Current**: `app/forgot-password/page.tsx`

**Layout**: Pattern F (Auth — split screen, narrower)
**Changes**:
1. Left panel: same branding
2. Right panel:
   - H2 "Reset Password"
   - Subtitle: "Enter your email..."
   - Email field
   - "Send Reset Link" button
   - "Back to Sign In" link

**Components touched**: `ForgotPasswordForm.tsx`, `AuthLayout.tsx`
**Icons**: email, send, arrow_back

---

### Page 16: `/reset-password` — Reset Password
**Mockup**: None — inferred (Pattern F: Auth)
**Current**: `app/reset-password/page.tsx`

**Layout**: Pattern F (Auth — split screen)
**Changes**:
1. H2 "Set New Password"
2. New password + confirm password fields
3. Password strength bar
4. "Reset Password" button
5. "Back to Sign In" link

**Components touched**: `ResetPasswordForm.tsx`, `AuthLayout.tsx`
**Icons**: lock, check, arrow_back

---

### Page 17: `/mfa-verify` — MFA Verify
**Mockup**: None — inferred (Pattern F: Auth)
**Current**: `app/mfa-verify/page.tsx`

**Layout**: Pattern F (Auth — centered card)
**Changes**:
1. Centered card: `max-w-md mx-auto mt-20`
2. Shield icon (w-16 h-16 bg-surface-container-high rounded-full)
3. H2 "Two-Factor Authentication"
4. 6-digit code input (6 separate boxes or single input)
5. "Verify" button
6. "Resend code" + "Use backup code" links

**Components touched**: `mfa-verify/page.tsx`
**Icons**: shield, verified_user, password, refresh

---

### Page 18: `/verify-email` — Email Verification
**Mockup**: None — inferred (Pattern F: Auth)
**Current**: `app/verify-email/page.tsx`

**Layout**: Pattern F (Auth — centered card)
**Changes**:
1. Centered card: `max-w-md mx-auto mt-20`
2. Email icon (w-16 h-16 bg-surface-container-high rounded-full)
3. H2 "Verify Your Email"
4. "We sent a verification link to..."
5. "Resend email" button + "Change email" link

**Icons**: email, send, check_circle

---

### Page 19: `/terms` — Terms of Service
**Mockup**: None — inferred (static content)
**Current**: `app/terms/page.tsx`

**Layout**: Page with sidebar + topnav
**Changes**:
1. H1 "Terms of Service" (text-headline-lg)
2. Content: `prose prose-invert max-w-3xl`:
   - Sections with h2 (text-headline-sm) + body text (text-body-md text-on-surface-variant)
   - Legal text in `bg-surface-container rounded-xl p-6`
   - Last updated date in text-label-md text-on-surface-variant
3. "Accept and Continue" button at bottom

**Icons**: description, gavel, check

---

### Page 20: `/profile` — Legacy Profile (REDIRECT)
**Mockup**: None
**Current**: `app/profile/page.tsx`

**Strategy**: Redirect to `/settings/profile`
**Changes**: Replace with `redirect('/settings/profile')`

---

### Page 21: `/security` — Legacy Security (REDIRECT)
**Mockup**: None
**Current**: `app/security/page.tsx`

**Strategy**: Redirect to `/settings/security`
**Changes**: Replace with `redirect('/settings/security')`

---

### Page 22: `/hosts/[id]` — Host Detail
**Mockup**: None — inferred from hosts_dashboard card pattern
**Current**: `app/hosts/[id]/page.tsx`

**Layout**: Pattern A variant (detail view)
**Changes**:
1. Back link: "← Back to Hosts"
2. Host header card: `bg-surface-container rounded-xl p-6`:
   - Large icon (w-16 h-16 based on type) + hostname (headline-lg) + IP (mono-code)
   - Status badge (same as HostCard)
   - Action buttons: Connect (bg-primary), Edit (border-outline-variant), Delete (text-error)
3. Tab bar (underline style): Overview | Sessions | Credentials | Settings
4. Overview tab: stats grid (4 cards: CPU, Memory, Disk, Uptime — each `bg-surface-container rounded-xl p-5`)
5. Sessions tab: session list
6. Credentials tab: linked credentials from vault
7. Settings tab: host config form (same input pattern)

**Components touched**: `HostMetadataPanel.tsx`, `HostForm.tsx`, `ConnectionButton.tsx`, `CopyableField.tsx`
**Icons**: dns, terminal, desktop_windows, screen_share, edit, delete, content_copy, play_circle, settings, history, key

---

### Page 23: `/tunnels` — Tunnels
**Mockup**: None — inferred from dashboard pattern
**Current**: `app/tunnels/page.tsx`

**Layout**: Pattern A (Dashboard variant)
**Changes**:
1. H1 "Tunnels" + subtitle "Manage WireGuard VPN tunnels secure remote access."
2. Stats cards row (4): Active Tunnels, Total Traffic, Connected Peers, Avg Latency
   - Each: `bg-surface-container rounded-xl p-5` + icon + value (headline-sm) + label (text-label-md)
3. "Create Tunnel" button: bg-primary text-on-primary
4. Tunnel list/table: `bg-surface-container rounded-xl border border-outline-variant`:
   - Columns: Name, Status, Peers, Traffic, Uptime, Actions
   - Status badges: Active (green), Inactive (gray), Error (red)
   - Row hover: bg-surface-variant/30
   - Actions: configure (settings), view details (more_vert), delete (delete_forever text-error)
5. Empty state: tunnel icon + "No tunnels configured" + create button

**Components touched**: `tunnels/TunnelTab.tsx`, `TunnelStatsCards.tsx`, `TunnelStatusBadge.tsx`, `TunnelConfigModal.tsx`, `TunnelDetailDialog.tsx`, `TunnelDeleteDialog.tsx`, `LiveStatsPanel.tsx`
**Icons**: vpn_key, cloud, add, settings, delete_forever, more_vert, speed, people, timer

---

### Page 24: `/discovery` — Network Discovery
**Mockup**: None — inferred
**Current**: `app/discovery/page.tsx`

**Layout**: Pattern A (Dashboard variant)
**Changes**:
1. H1 "Network Discovery" + subtitle "Scan discover hosts local remote networks."
2. Scan controls bar: IP range input (bg-surface-container-low border-outline-variant) + port range + "Start Scan" button (bg-primary)
3. Progress bar: `bg-surface-container` track + `bg-primary` fill
4. Results table: `bg-surface-container rounded-xl border border-outline-variant`:
   - Columns: IP, Hostname, OS, Open Ports, Status, Actions
   - Status: Reachable (green), Unreachable (red)
   - Actions: "Add as Host" (text-primary link)
5. Scan history card: recent scans with timestamp + IP range + hosts found

**Components touched**: `app/discovery/page.tsx`
**Icons**: radar, search, dns, add, play_circle, schedule, cloud_off, check_circle

---

### Page 25: `/files` — File Manager
**Mockup**: None — inferred
**Current**: `app/files/page.tsx`

**Layout**: Full-width with sidebar
**Changes**:
1. H1 "File Manager" + subtitle
2. Host selector dropdown (which host's filesystem)
3. File manager layout (3-panel):
   - Left: file tree (`bg-surface-container border-r border-outline-variant w-64`)
   - Center: file list/grid (`bg-surface-container-lowest`)
   - Right: preview pane (`bg-surface-container border-l border-outline-variant w-80`)
4. Toolbar: path breadcrumb, view toggle (grid/list), upload button, new folder
5. File items: icon (by type), name (text-body-md), size (text-label-md text-on-surface-variant), modified date
6. Selected file: `bg-primary/10 border-primary`
7. Context menu: open, rename, delete, download, properties

**Components touched**: `file-manager/*` (all 8 files)
**Icons**: folder, description, image, insert_drive_file, download, upload, delete, create_new_folder, grid_view, view_list, more_vert

---

### Page 26: `/sessions` — Sessions List
**Mockup**: None — inferred from dashboard pattern
**Current**: `app/sessions/page.tsx`

**Layout**: Pattern A (Dashboard variant)
**Changes**:
1. H1 "Sessions" + subtitle "View manage active past SSH sessions."
2. Stats cards: Active Sessions, Total Today, Avg Duration, Peak Concurrent
3. Session table: `bg-surface-container rounded-xl border border-outline-variant`:
   - Columns: User, Host, Protocol, Started, Duration, Status, Actions
   - Status: Active (green pulse), Ended (gray), Terminated (red)
   - Actions: view recording (play_circle), terminate (cancel text-error)
4. Filter bar: by user, by host, by status, by date range

**Components touched**: `SessionList.tsx`, `LoginHistory.tsx`
**Icons**: history, terminal, desktop_windows, play_circle, cancel, timer, people, calendar_today

---

### Page 27: `/sessions/recordings` — Session Recordings
**Mockup**: None — inferred
**Current**: `app/sessions/recordings/page.tsx`

**Layout**: Pattern A (Dashboard variant)
**Changes**:
1. H1 "Session Recordings" + subtitle
2. Recording list: card grid or table:
   - Each recording: session info, duration, date, file size
   - Thumbnail/preview: `bg-black rounded-lg` terminal screenshot
3. Player: `bg-surface-container-high rounded-xl p-4`:
   - Timeline scrubber (bg-surface-container + bg-primary fill)
   - Play/pause button (bg-primary)
   - Speed selector
   - Download button

**Components touched**: `recording-player.tsx`
**Icons**: play_circle, pause, download, speed, timeline, terminal, history

---

### Page 28: `/settings/sessions` — Session Management Settings
**Mockup**: None — inferred from settings pattern
**Current**: `app/settings/sessions/page.tsx`

**Layout**: Pattern E (Settings Sub-Page)
**Changes**:
1. H1 "Session Management" + subtitle
2. Session timeout card: timeout duration input + "Auto-terminate idle sessions" toggle
3. Recording settings card: "Record all sessions" toggle + retention period + storage path
4. Concurrent sessions card: max sessions per user + warning threshold
5. Security card: "Require MFA for new sessions" toggle + "IP allowlist" textarea

**Icons**: history, timer, record, security, lock, schedule

---

### Page 29: `/settings/webauthn` — WebAuthn Settings
**Mockup**: None — inferred from settings pattern
**Current**: `app/settings/webauthn/page.tsx`

**Layout**: Pattern E (Settings Sub-Page)
**Changes**:
1. H1 "WebAuthn / Passkeys" + subtitle "Register security keys passwordless authentication."
2. Registered keys card: list of registered devices:
   - Each: device name, type (icon), registered date, last used
   - "Rename" + "Remove" actions
3. Register new key button: `bg-primary text-on-primary`
   - Flow: browser WebAuthn API → success toast
4. Security key test card: "Test authentication" button

**Components touched**: `WebAuthnSettings.tsx`
**Icons**: verified_user, key, security, add, delete, check_circle, devices

---

### Page 30: `/admin` — Admin Dashboard
**Mockup**: None — inferred from dashboard pattern
**Current**: `app/admin/page.tsx`

**Layout**: Pattern A (Dashboard)
**Changes**:
1. H1 "Admin Dashboard" + subtitle
2. Stats cards row (6): Total Users, Active Sessions, Total Hosts, Tunnels, Vault Entries, System Health
   - Each: `bg-surface-container rounded-xl p-5` + icon + value + label
3. Recent activity feed: timeline list in `bg-surface-container rounded-xl`
4. Quick actions: "Add User", "Add Host", "View Audit", "Configure System"

**Components touched**: `admin/stats-cards.tsx`
**Icons**: admin_panel_settings, people, dns, vpn_key, shield, terminal, health_and_safety, add, history

---

### Page 31: `/admin/audit` — Audit Log
**Mockup**: None — inferred from data table pattern
**Current**: `app/admin/audit/page.tsx`

**Layout**: Pattern A (Dashboard variant — data table)
**Changes**:
1. H1 "Audit Log" + subtitle
2. Filter bar: user dropdown, action type, date range, search
3. Audit table: `bg-surface-container rounded-xl border border-outline-variant`:
   - Columns: Timestamp, User, Action, Resource, IP, Status
   - `text-mono-code` for timestamps and IPs
   - Status: Success (green), Denied (red), Warning (yellow) badges
4. Export button: `border-outline-variant`
5. Pagination: page numbers + total count

**Components touched**: `admin/tables/data-table.tsx`
**Icons**: history, filter_list, download, search, admin_panel_settings, person, action, check_circle, cancel, warning

---

### Page 32: `/admin/hosts` — Admin Hosts
**Mockup**: None — inferred
**Current**: `app/admin/hosts/page.tsx`

**Layout**: Pattern A (Dashboard variant — admin table)
**Changes**:
1. H1 "Host Management" + subtitle
2. Stats: Total, Active, Offline, by protocol (SSH/RDP/VNC)
3. Hosts table: Name, IP, Protocol, Owner, Status, Assigned Groups, Actions
4. Bulk actions: select all, delete selected, assign group
5. "Add Host" button

**Icons**: dns, admin_panel_settings, add, delete, group_work, terminal, desktop_windows, screen_share

---

### Page 33: `/admin/metrics` — System Metrics
**Mockup**: None — inferred
**Current**: `app/admin/metrics/page.tsx`

**Layout**: Pattern A (Dashboard — metrics)
**Changes**:
1. H1 "System Metrics" + subtitle
2. Time range selector: 1h, 24h, 7d, 30d (pill buttons, active: bg-secondary-container)
3. Chart cards (4): CPU Usage (line), Memory (line), Disk I/O (bar), Network (area)
   - Each: `bg-surface-container rounded-xl p-6` + chart canvas
   - Chart colors: primary for main, tertiary for secondary
4. Resource table: per-host resource usage

**Components touched**: `admin/charts/line-chart.tsx`, `bar-chart.tsx`, `pie-chart.tsx`
**Icons**: metrics, memory, storage, network_check, timer, trending_up, show_chart

---

### Page 34: `/admin/sessions` — Admin Sessions
**Mockup**: None — inferred
**Current**: `app/admin/sessions/page.tsx`

**Layout**: Pattern A (Dashboard variant — admin table)
**Changes**:
1. H1 "Session Management" + subtitle
2. Stats: Active, Ended, Terminated, Avg Duration
3. Sessions table: User, Host, Protocol, Started, Duration, Status, Actions (terminate, view recording)
4. Bulk: terminate selected

**Icons**: admin_panel_settings, history, terminal, play_circle, cancel, people, timer

---

### Page 35: `/admin/users` — User Management
**Mockup**: None — inferred
**Current**: `app/admin/users/page.tsx`

**Layout**: Pattern A (Dashboard variant — admin table)
**Changes**:
1. H1 "User Management" + subtitle
2. Stats: Total Users, Active, Admins, Suspended
3. Users table: Avatar + Name, Email, Role (badge), Last Login, Status, Actions
   - Role badges: Admin (bg-primary/10 text-primary), User (bg-surface-variant), Suspended (bg-error/10 text-error)
4. "Add User" button + "Invite User" button
5. User modal: name, email, role select, MFA required toggle

**Components touched**: `admin/tables/data-table.tsx`, `admin/stats-cards.tsx`
**Icons**: people, person_add, admin_panel_settings, email, shield, delete, edit, more_vert

---

### Page 36: All `/admin/*` tests
**Mockup**: N/A
**Current**: `app/admin/__tests__/dashboard.test.tsx`

**Changes**: Update test selectors to match new layout structure (sidebar, new class names)

---

## Execution Order (Prioritized)

### Phase 1: Foundation (1 dispatch)
1. globals.css — full Material 3 token set
2. tailwind.config.ts — hex colors, typography, spacing, radii
3. layout.tsx — Material Symbols font
4. material-icon.tsx — wrapper component

### Phase 2: Shell (1 dispatch)
5. Sidebar.tsx — new
6. TopNav.tsx — redesign
7. DashboardLayout.tsx — restructure
8. AuthLayout.tsx — new auth layout

### Phase 3: Components (2 dispatches)
9-26. All 28 UI components restyled

### Phase 4: Mockup Pages (3 dispatches)
27. hosts/page.tsx + HostCard + HostList
28. terminal/page.tsx + terminal components
29. vault/page.tsx + vault components
30. settings/page.tsx (overview)
31. settings/profile + settings/security
32. settings/api-keys + settings/appearance
33. settings/notifications (NEW)

### Phase 5: Inferred Pages (3 dispatches)
34. Auth pages (6): login, register, forgot-password, reset-password, mfa-verify, verify-email
35. Admin pages (6): dashboard, audit, hosts, metrics, sessions, users
36. Infrastructure pages (5): discovery, files, hosts/[id], tunnels, vault/share
37. Remaining (4): sessions, sessions/recordings, settings/sessions, settings/webauthn
38. Redirects (3): profile→settings/profile, security→settings/security, root→hosts
39. Static (2): terms, welcome

### Phase 6: Polish (1 dispatch)
40. Responsive, accessibility, loading/empty states, test fixes

**Total estimated dispatches: 10-12**