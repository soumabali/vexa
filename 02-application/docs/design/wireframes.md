# vexa — Wireframes & Screen Specifications

> **Classification:** Security-Critical Application  
> **Owner:** Dhar  > **Status:** Draft — Design Review  > **Last Updated:** 2026-05-27

---

## 1. Login / Authentication Screen

### 1.1 Layout Structure
```
+----------------------------------------------------------+
|                                                          |
|                    [Logo: Shield+Tower]                  |
|                    "CENTRAL TOWER"                       |
|                    "Secure Remote Access"                |
|                                                          |
|  +----------------------------------------------------+  |
|  |  Sign In                                           |  |
|  |                                                    |  |
|  |  Email Address                                     |  |
|  |  ┌──────────────────────────────────────────────┐  |  |
|  |  │ admin@corp.local                               │  |  |
|  |  └──────────────────────────────────────────────┘  |  |
|  |                                                    |  |
|  |  Password                                          |  |
|  |  ┌──────────────────────────────────────────────┐  |  |
|  |  │ ••••••••••••••                  [Eye icon]     │  |  |
|  |  └──────────────────────────────────────────────┘  |  |
|  |  [ ] Remember this device for 30 days              |  |
|  |                                                    |  |
|  |  ┌──────────────────────────────────────────────┐  |  |
|  |  │           🔐 Sign In                         │  |  |
|  |  └──────────────────────────────────────────────┘  |  |
|  |                                                    |  |
|  |  ──────────────── OR ─────────────────────────────   |  |
|  |                                                    |  |
|  |  ┌──────────────────────────────────────────────┐  |  |
|  |  │   🖐️ Sign in with WebAuthn / Security Key    │  |  |
|  |  └──────────────────────────────────────────────┘  |  |
|  |                                                    |  |
|  |  🔑 Trouble signing in? · Reset password           |  |
|  |                                                    |  |
|  +----------------------------------------------------+  |
|                                                          |
|  🔒 End-to-end encrypted · TLS 1.3 · v2.4.1            |
|                                                          |
+----------------------------------------------------------+
```

### 1.2 Elements
- **Background:** `bg-base` with subtle gradient (dark blue radial from center)
- **Logo:** Animated SVG — shield with tower, lock icon in center
- **Card:** Centered, max-width 420px, `bg-surface`, `shadow-lg`, border-radius 12px
- **Security Footer:** Fixed at bottom, small text with lock icon, version number

### 1.3 MFA Flow — Step 2 (After Password Verified)
```
+----------------------------------------------------------+
|  🔐 Two-Factor Authentication                            |
|                                                          |
|  Step 1 ✓  →  Step 2 ●  →  Step 3 ○                    |
|  ───────────────────────────────────────────             |
|                                                          |
|  [Primary: TOTP]  [WebAuthn]  [Backup Codes]             |
|                                                          |
|  Enter the 6-digit code from your authenticator app:    |
|                                                          |
|  ┌────┐┌────┐┌────┐┌────┐┌────┐┌────┐                 |
|  │ 5  ││ 2  ││ 9  ││    ││    ││    │  ← Auto-focus    |
|  └────┘└────┘└────┘└────┘└────┘└────┘  Auto-advance    |
|                                                          |
|  Time remaining: [████████░░░░░░░░░░░░] 18s              |
|                                                          |
|  ┌──────────────────────────────────────────────┐       |
|  │           Verify Code                        │       |
|  └──────────────────────────────────────────────┘       |
|                                                          |
|  🔑 Use a backup code instead                            |
|                                                          |
+----------------------------------------------------------+
```

### 1.4 MFA — WebAuthn Tab
```
┌──────────────────────────────────────────────┐
│  [TOTP]  [WebAuthn ●]  [Backup Codes]        │
│                                              │
│     ┌──────────────────────┐                │
│     │                      │                │
│     │    🖐️                │                │
│     │   Fingerprint        │                │
│     │                      │                │
│     │  Touch your          │                │
│     │  security key or     │                │
│     │  use biometrics      │                │
│     │                      │                │
│     └──────────────────────┘                │
│                                              │
│  ┌────────────────────────────────────────┐  │
│  │   🖐️ Authenticate with Security Key   │  │
│  └────────────────────────────────────────┘  │
└──────────────────────────────────────────────┘
```

### 1.5 MFA — Backup Codes Tab
```
┌──────────────────────────────────────────────┐
│  [TOTP]  [WebAuthn]  [Backup Codes ●]        │
│                                              │
│  Enter one of your saved backup codes:       │
│                                              │
│  ┌──────────────────────────────────────┐  │
│  │  a1b2-c3d4-e5f6                      │  │
│  └──────────────────────────────────────┘  │
│                                              │
│  💡 Each code can only be used once.         │
│     Generate new codes in Settings.          │
│                                              │
│  ┌──────────────────────────────────────┐  │
│  │   Verify Backup Code                 │  │
│  └──────────────────────────────────────┘  │
└──────────────────────────────────────────────┘
```

---

## 2. Dashboard

### 2.1 Layout (Desktop — Sidebar Expanded)
```
+----------------------------------------------------------+
|  Sidebar |  Top Bar                                       |
|  [🏠]    |  "Dashboard"            [🔍] [🔔] [👤]         |
|  [🖥️]    |                                                |
|  [📋]    |  +------------------------------------------+  |
|  [📊]    |  |  Quick Stats                             |  |
|  [🔑]    |  |  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐  |  |
|  [⚙️]    |  |  │ 🔒 3 │ │ 📋 2 │ │ ⚠️ 1 │ │ 🕐 0 │  |  |
|  [🚪]    |  |  │Active│ │Sess. │ │Alerts│ │Expir.│  |  |
|          |  |  └──────┘ └──────┘ └──────┘ └──────┘  |  |
|  64px    |  +------------------------------------------+  |
|  /240px  |                                                |
|          |  +-------------------+ +--------------------+  |
|          |  |  🔒 Active        | |  🕐 Recent Sessions  |  |
|          |  |  Connections      | |                      |  |
|          |  |                   | |  ┌────────────────┐  |  |
|          |  |  ┌────────────┐   | |  │ 🟢 SSH prod-web-01 │  |
|          |  |  │ 🟢 prod-web│   | |  │   12 min ago       │  |
|          |  |  │    ● SSH   │   | |  │   admin@prod       │  |
|          |  |  │   14:32:05 │   | |  └────────────────┘  |  |
|          |  |  │   🔒 TLS   │   | |                      |  |
|          |  |  └────────────┘   | |  ┌────────────────┐  |  |
|          |  |                   | |  │ 🔵 RDP win-dc-01   │  |
|          |  |  ┌────────────┐   | |  │   2 hours ago      │  |
|          |  |  │ 🔵 win-dc01│   | |  │   ops@corp         │  |
|          |  |  │    ● RDP   │   | |  └────────────────┘  |  |
|          |  |  │   09:15:22 │   | |                      |  |
|          |  |  │   🔒 TLS   │   | |  [View All Sessions →]  |
|          |  |  └────────────┘   | |                      |  |
|          |  |                   | +--------------------+  |
|          |  +-------------------+                          |
|          |                                                |
|          |  +------------------------------------------+  |
|          |  |  🖥️ Host Quick Access                    |  |
|          |  |  [🔍 Search hosts...]                    |  |
|          |  |                                          |  |
|          |  |  ┌────────┐ ┌────────┐ ┌────────┐      |  |
|          |  |  │[SSH]🟢│ │[RDP]🔵│ │[VNC]🟡│ ...  |  |
|          |  |  │prod-db │ │win-term│ │lab-kvm │      |  |
|          |  |  │10.0.1.5│ │10.0.2.8│ │10.0.3.1│      |  |
|          |  |  │[Connect]│[Connect]│[Connect]│      |  |
|          |  |  └────────┘ └────────┘ └────────┘      |  |
|          |  +------------------------------------------+  |
|          |                                                |
|          |  +------------------------------------------+  |
|          |  |  ⚠️ Security Alerts                      |  |
|          |  |  ┌──────────────────────────────────┐  |  |
|          |  |  │ ⚠️ Unverified host key for        │  |
|          |  |  │    prod-web-02 (10.0.1.6)         │  |
|          |  |  │    [Verify] [Dismiss]              │  |
|          |  |  └──────────────────────────────────┘  |  |
|          |  +------------------------------------------+  |
|          |                                                |
+----------------------------------------------------------+
```

### 2.2 Dashboard Elements Detail

#### Quick Stats Bar
- Four stat cards in a horizontal row (desktop), 2x2 grid (mobile)
- Each card: Icon + number + label + trend arrow (optional)
- Colors: Active=success, Sessions=info, Alerts=warning, Expiring=default
- Clicking a card filters the relevant section below

#### Active Connections Panel
- Vertical card list, each card representing one live session
- Card content (left to right):
  1. Protocol color dot (pulsing animation)
  2. Host name (bold) + protocol label
  3. Session duration (updating live, mono font)
  4. Security badge (🔒 TLS, 🖐️ MFA)
  5. Quick actions: [Minimize] [Kill] (two-step)
- Max height: 400px, scrollable
- Empty state: "No active connections" with [Quick Connect] CTA

#### Recent Sessions Panel
- Vertical list, read-only
- Each row: Protocol dot + Host name + "X time ago" + User
- Shows last 5 sessions, [View All] link at bottom
- Empty state: "No recent sessions"

#### Host Quick Access
- Search bar at top, filters hosts in real-time
- Host cards in responsive grid: 4 cols (lg), 3 cols (md), 2 cols (sm), 1 col (xs)
- Each card:
  - Protocol badge top-left (SSH/RDP/VNC)
  - Host name (h3)
  - IP address (caption, text-secondary)
  - [Connect] button (primary small)
  - [⋯] menu: Edit, Duplicate, Delete
- Contextual hover: Shows tags, last connected, health status dot

#### Security Alerts Panel
- Collapsible, starts expanded if alerts exist
- Each alert row: Icon + message + timestamp + actions
- Severity colors: warning yellow, danger red
- Dismissed alerts slide out 200ms then collapse
- Critical alerts (e.g., brute-force detected) cannot be dismissed, only acknowledged

### 2.3 Mobile Dashboard Layout
```
+----------------------------------------------------------+
|  [≡]  vexa      [🔔] [👤]                       |
+----------------------------------------------------------+
|                                                          |
|  Quick Stats (Horizontal scroll)                           |
|  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐                   |
|  │ 🔒 3 │ │ 📋 2 │ │ ⚠️ 1 │ │ 🕐 0 │                   |
|  │Active│ │Sess. │ │Alerts│ │Expir.│                   |
|  └──────┘ └──────┘ └──────┘ └──────┘                   |
|                                                          |
|  🔍 Search hosts...                                      |
|                                                          |
|  🖥️ Hosts                                                |
|  ┌──────────────────────┐                                |
|  │ 🟢 SSH prod-web-01   │                                |
|  │ 10.0.1.4            │ [Connect]                      |
|  └──────────────────────┘                                |
|  ┌──────────────────────┐                                |
|  │ 🔵 RDP win-dc-01     │                                |
|  │ 10.0.2.8            │ [Connect]                      |
|  └──────────────────────┘                                |
|                                                          |
|  📋 Active Sessions                                      |
|  ┌──────────────────────┐                                |
|  │ 🟢 prod-web-01 ●     │                                |
|  │ SSH · 14:32:05 · 🔒  │ [Kill]                         |
|  └──────────────────────┘                                |
|                                                          |
|  ⚠️ Security Alerts                                      |
|  ┌──────────────────────┐                                |
|  │ ⚠️ Unverified host... │                                |
|  │ [Verify] [Dismiss]   │                                |
|  └──────────────────────┘                                |
|                                                          |
+----------------------------------------------------------+
|  [🏠]  [🖥️]  [📋]  [📊]  [⚙️]                           |
+----------------------------------------------------------+
```

---

## 3. Terminal View (SSH)

### 3.1 Layout (Desktop)
```
+----------------------------------------------------------+
|  Sidebar |  Terminal Area                                  |
|  collapsed|                                                |
|          |  +------------------------------------------+  |
|  [🏠]    |  | Tab Bar                                    |  |
|  [🖥️]    |  | [🟢 prod-web-01 ● SSH] [🔵 db-master ●   |  |
|  [📋]    |  |  SSH] [🟡 lab-kvm ● VNC] [+]              |  |
|          |  +------------------------------------------+  |
|          |  |                                            |  |
|          |  |  Terminal Content Area                     |  |
|          |  |                                            |  |
|          |  |  $ ls -la                                  |  |
|          |  |  total 128                                 |  |
|          |  |  drwxr-xr-x  5 root root 4096 May 27 10:00 |  |
|          |  |  -rw-r--r--  1 root root  220 May 27 09:45 |  |
|          |  |  $ _                                       |  |
|          |  |  █                                         |  |
|          |  |                                            |  |
|          |  |                                            |  |
|          |  |                                            |  |
|          |  +------------------------------------------+  |
|          |  | Status Bar                                 |  |
|          |  | 🔒 TLS 1.3 | 🖐️ MFA | 80x24 | UTF-8 | 🟢   |  |
|          |  +------------------------------------------+  |
|          |                                                |
+----------------------------------------------------------+
```

### 3.2 Tab Bar Detail
- **Height:** 36px
- **Active tab:** `bg-inset`, no bottom border, sharp bottom corners
- **Inactive tab:** `bg-surface`, bottom border `border-subtle`
- **Tab content:** Protocol dot (6px, pulsing) + host name (truncated to 20 chars) + close button (×, visible on hover)
- **New tab button (+):** Ghost button, right-aligned
- **Tab overflow:** Horizontal scroll with < > arrows if >8 tabs, or dropdown menu
- **Unread/pending:** Red badge dot on tab if new output while tab inactive
- **Drag to reorder:** Supported on desktop (touch on tablet)

### 3.3 Terminal Content Area
- **Background:** `bg-inset` (#070A10) — ALWAYS dark regardless of theme
- **Padding:** 16px
- **Font:** `JetBrains Mono`, 13px, line-height 1.2
- **Scrollback:** User-configurable (default 10000 lines)
- **Scrollbar:** 8px, styled, auto-hide after 2s inactivity
- **Selection:** Click-drag to select, `bg-info/30` highlight
- **Right-click context menu:** Copy, Paste, Select All, Clear, Search

### 3.4 Status Bar (Bottom)
- **Height:** 28px, `bg-surface`, border-top `border-subtle`
- **Left section:** Security badges (🔒 TLS, 🖐️ MFA), host address
- **Center:** Terminal dimensions (cols × rows), encoding
- **Right section:** Connection health dot, session duration, [⋮] menu
- **Status menu:** Download recording, Session info, Kill session, Keyboard shortcuts

### 3.5 Mobile Terminal View
```
+----------------------------------------------------------+
|  [←] prod-web-01  [🔒] [⋯]                              |
+----------------------------------------------------------+
|                                                          |
|  $ ls -la                                                |
|  total 128                                               |
|  drwxr-xr-x  5 root root 4096 May 27 10:00              |
|  -rw-r--r--  1 root root  220 May 27 09:45              |
|  $ _                                                     |
|  █                                                       |
|                                                          |
|                                                          |
|                                                          |
|                                                          |
+----------------------------------------------------------+
|  ⌨️  [Ctrl]  [Alt]  [Tab]  [Esc]  [Paste]               |
+----------------------------------------------------------+
```

- **Top bar:** Back button, truncated host name, security badge, actions menu
- **Screen keyboard bar:** Essential keys that are hard to type on mobile keyboards
- **Swipe gestures:**
  - Swipe left from right edge: Switch to next tab
  - Swipe right from left edge: Switch to previous tab
  - Swipe up from bottom: Show full keyboard
  - Two-finger tap: Paste clipboard

### 3.6 RDP/VNC Viewer (Remote Desktop)

```
+----------------------------------------------------------+
|  [<-] win-dc-01    🔵 RDP · 🔒 TLS · 🖐️ MFA  [⋯]      |
+----------------------------------------------------------+
|                                                          |
|  +----------------------------------------------------+  |
|  |                                                    |  |
|  |              Remote Desktop Canvas                 |  |
|  |                                                    |  |
|  |              (Scalable / Scrollable)               |  |
|  |                                                    |  |
|  |                                                    |  |
|  |                                                    |  |
|  +----------------------------------------------------+  |
|                                                          |
|  ┌────────────────────────────────────────────────────┐  |
|  │ [🚪 Disconnect] [⛶ Fullscreen] [📋 Clipboard]      │  |
|  │ [⌨️ Keyboard] [🔍 Zoom: 100%] [🖱️ Mouse]           │  |
|  └────────────────────────────────────────────────────┘  |
|                                                          |
+----------------------------------------------------------+
```

- **Canvas:** Renders remote framebuffer, scales to fit or 1:1 mode
- **Toolbar:** Floating, bottom-center, auto-hide after 3s (tap to reveal)
- **Zoom:** 50%, 75%, 100%, 125%, 150%, Fit to window
- **Clipboard:** Bidirectional sync toggle (with security confirmation)
- **Mouse modes:** Direct (touch = mouse), Trackpad (relative), Pen tablet
- **Performance indicator:** Optional FPS / latency overlay (corner, toggleable)

---

## 4. Host Management

### 4.1 Host List Screen
```
+----------------------------------------------------------+
|  Hosts                                          [+ Add]  |
|                                                          |
|  ┌────────────────────────────────────────────────────┐  |
|  │ 🔍 Search hosts, tags, or IP...                    │  |
|  │ [Filter ▼] [Group by ▼] [Import] [Export]          │  |
|  └────────────────────────────────────────────────────┘  |
|                                                          |
|  📁 Production  (12)    [▼]                              |
|  ┌────────────────────────────────────────────────────┐  |
|  │ 🟢 │ SSH │ prod-web-01     │ 10.0.1.4 │ 🔒 │ [⋯] │  │
|  │    │     │ web server cluster                    │  │
|  ├────────────────────────────────────────────────────┤  |
|  │ 🟢 │ SSH │ prod-web-02     │ 10.0.1.5 │ ⚠️ │ [⋯] │  │
|  │    │     │ web server cluster                    │  │
|  ├────────────────────────────────────────────────────┤  |
|  │ 🔵 │ RDP │ win-dc-01       │ 10.0.2.8 │ 🔒 │ [⋯] │  │
|  │    │     │ Active Directory DC                   │  │
|  └────────────────────────────────────────────────────┘  |
|                                                          |
|  📁 Development  (4)    [▼]                              |
|  ┌────────────────────────────────────────────────────┐  |
|  │ 🟢 │ SSH │ dev-app-01      │ 192.168.1.10 │ 🔒 │   │  │
|  ├────────────────────────────────────────────────────┤  |
|  │ 🟡 │ VNC │ lab-kvm         │ 192.168.1.50 │ 🔒 │   │  │
|  └────────────────────────────────────────────────────┘  |
|                                                          |
|  📁 Archive  (2)    [▶]                                  |
|                                                          |
+----------------------------------------------------------+
```

### 4.2 Host Card Detail
Each row/card in the list:
- **Left edge:** Health/reachability dot (green=online, red=offline, gray=unknown, pulsing spinner=checking)
- **Protocol badge:** Colored badge (SSH/RDP/VNC)
- **Host name:** h4, bold
- **Description/Tags:** caption, text-secondary
- **IP/Address:** mono font, text-secondary
- **Security badge:** 🔒, ⚠️, or ❌
- **Actions menu (⋯):** Edit, Duplicate, Connect, Test Reachability, Move to Group, Delete

### 4.3 Add/Edit Host Dialog
```
+----------------------------------------------------------+
|  🖥️ Add New Host                              [×]        |
+----------------------------------------------------------+
|                                                          |
|  Connection Details                                      |
|  ─────────────────────────────────────────               |
|                                                          |
|  Protocol                           Display Name          │
|  ┌────────────┐                    ┌──────────────────┐  │
|  │ SSH ▼      │                    │ Production Web 01│  │
|  └────────────┘                    └──────────────────┘  │
|                                                          |
|  Hostname / IP Address            Port                   │
|  ┌──────────────────────────┐    ┌──────────┐          │
|  │ prod-web-01.corp.local    │    │ 22       │          │
|  │ ● Resolving...            │    │          │          │
|  └──────────────────────────┘    └──────────┘          │
|  192.168.1.4 (resolved)                                  │
|                                                          |
|  Credentials                                             │
|  ─────────────────────────────────────────               |
|                                                          |
|  [● Use saved credential]  [○ Enter manually]            │
|                                                          |
|  ┌────────────────────────────┐                          │
|  │ 🔑 prod-web-key (SSH Key)  │                          │
|  └────────────────────────────┘                          │
|                                                          |
|  [+ Add New Credential]                                  │
|                                                          |
|  Advanced Options                                        |
|  ─────────────────────────────────────────               |
|  [▶]                                                     |
|                                                          |
|  Tags: [production] [web] [x]                            │
|  Group: ┌──────────────┐                                 │
|         │ Production ▼ │                                 │
|         └──────────────┘                                 │
|                                                          |
|  ┌────────────────────────────────────────────────────┐  │
|  │  [Cancel]              [💾 Save Host]             │  │
|  └────────────────────────────────────────────────────┘  │
+----------------------------------------------------------+
```

### 4.4 Credential Selector Dropdown
```
┌────────────────────────────────────┐
│ 🔍 Search credentials...           │
├────────────────────────────────────┤
│ 🔑 prod-web-key (SSH Key)        │
│    Fingerprint: SHA256:abc...     │
├────────────────────────────────────┤
│ 🔐 admin-password (Password)     │
│    Last used: 2 hours ago         │
├────────────────────────────────────┤
│ 🔑 api-token (API Token)         │
├────────────────────────────────────┤
│ [+ Add New Credential]           │
└────────────────────────────────────┘
```

### 4.5 Host Group Management
```
+----------------------------------------------------------+
|  Manage Groups                                           |
|                                                          |
|  Drag to reorder. Right-click for options.               |
|                                                          |
|  ┌────────────────────────────────────────────────────┐  │
|  │ ☰ │ 📁 Production        │ 12 hosts │ [⋯]        │  │
|  │ ☰ │ 📁 Development       │ 4 hosts  │ [⋯]        │  │
|  │ ☰ │ 📁 Staging           │ 3 hosts  │ [⋯]        │  │
|  │ ☰ │ 📁 Archive           │ 2 hosts  │ [⋯]        │  │
|  └────────────────────────────────────────────────────┘  │
|                                                          |
|  [+ Create New Group]                                    │
|                                                          |
+----------------------------------------------------------+
```

---

## 5. Settings

### 5.1 Settings Layout
```
+----------------------------------------------------------+
|  Settings                                                |
|                                                          |
|  ┌──────────┐  +--------------------------------------+  │
|  │ General  │  │  General Settings                    │  │
|  │ ──────── │  │                                      │  │
|  │ Security │  │  Appearance                          │  │
|  │ ──────── │  │  [● Dark] [○ Light] [○ System]     │  │
|  │ Sessions │  │                                      │  │
|  │ ──────── │  │  Language                            │  │
|  │ Terminal │  │  ┌────────────────────────────┐     │  │
|  │ ──────── │  │  │ English (US) ▼              │     │  │
|  │ Network  │  │  └────────────────────────────┘     │  │
|  │ ──────── │  │                                      │  │
|  │ Auditing │  │  Sidebar                              │  │
|  │ ──────── │  │  [✓] Collapse sidebar by default    │  │
|  │          │  │                                      │  │
|  │          │  │  Notifications                        │  │
|  │          │  │  [✓] Desktop notifications            │  │
|  │          │  │  [✓] Sound alerts                   │  │
|  │          │  │                                      │  │
|  +──────────+  +--------------------------------------+  │
|                                                          |
+----------------------------------------------------------+
```

### 5.2 Settings Categories

#### General
- Theme (Dark/Light/System)
- Language selector
- Sidebar behavior
- Notifications (desktop + sound)
- Date/time format
- Start page (Dashboard/Last Session/Host List)

#### Security
```
+----------------------------------------------------------+
│  Security Settings                                       │
│                                                          │
│  Authentication                                          │
│  ─────────────────────────────────────────               │
│                                                          │
│  Two-Factor Authentication                               │
│  Status: 🖐️ Enabled (TOTP + WebAuthn)                   │
│  [Manage 2FA →]                                        │
│                                                          │
│  Session Security                                        │
│  ─────────────────────────────────────────               │
│                                                          │
│  Idle timeout                                            │
│  ┌──────────────────────────────────────────┐           │
│  │ After 30 minutes of inactivity           ▼│           │
│  └──────────────────────────────────────────┘           │
│                                                          │
│  [✓] Lock session instead of disconnect                  │
│  [✓] Require password to unlock                        │
│                                                          │
│  Vault                                                   │
│  ─────────────────────────────────────────               │
│                                                          │
│  Auto-lock vault                                         │
│  ┌──────────────────────────────────────────┐           │
│  │ After 15 minutes of inactivity         ▼│           │
│  └──────────────────────────────────────────┘           │
│                                                          │
│  [🔐 Change Master Password]                             │
│  [🔑 Export Vault Backup]                                │
│  [⚠️ Delete Vault Data]                                  │
│                                                          │
│  Audit                                                   │
│  ─────────────────────────────────────────               │
│  [✓] Include my sessions in audit log                  │
│  [View My Audit Trail →]                                 │
│                                                          │
+----------------------------------------------------------+
```

#### MFA Management Screen
```
+----------------------------------------------------------+
│  🔐 Two-Factor Authentication                            │
│                                                          │
│  Configured Methods                                        │
│  ─────────────────────────────────────────               │
│                                                          │
│  ┌────────────────────────────────────────────────────┐  │
│  │ 📱 TOTP (Authenticator App)                        │  │
│  │ Status: Active · Added May 20, 2026                │  │
│  │ [Regenerate Backup Codes] [Disable]                │  │
│  └────────────────────────────────────────────────────┘  │
│                                                          │
│  ┌────────────────────────────────────────────────────┐  │
│  │ 🖐️ WebAuthn / Security Key                       │  │
│  │ Status: Active · YubiKey 5 NFC · Added May 21     │  │
│  │ [Rename] [Remove]                                  │  │
│  └────────────────────────────────────────────────────┘  │
│                                                          │
│  Backup Codes                                            │
│  ─────────────────────────────────────────               │
│  Remaining: 8 of 10                                     │
│  [📋 View Codes] [🔄 Generate New Codes]                 │
│                                                          │
│  Add New Method                                          │
│  ─────────────────────────────────────────               │
│  [📱 Set up Authenticator App]                           │
│  [🖐️ Register Security Key]                               │
│                                                          │
+----------------------------------------------------------+
```

#### Terminal Settings
- Font family (JetBrains Mono, Fira Code, SF Mono, Custom)
- Font size (10-20px slider)
- Line height
- Cursor style (Block/Bar/Underline)
- Cursor blink (toggle)
- Color scheme (Presets: Dracula, Solarized, Monokai, Custom)
- Scrollback buffer size
- Bell (visual/audible/none)
- Ligatures (toggle)
- Copy on select (toggle)

#### Network Settings
- Default SSH port
- Default RDP port
- Default VNC port
- Connection timeout
- Keepalive interval
- Jump host / Bastion configuration
- SOCKS proxy
- WireGuard tunnel settings

#### Audit Settings (Admin only)
- Log retention period
- Real-time log streaming
- Export format (JSON, CSV, PDF)
- Alert webhooks
- Integrity verification schedule

---

## 6. Audit Logs Viewer

### 6.1 Layout
```
+----------------------------------------------------------+
|  Audit Logs                          [Export] [Refresh]    |
|                                                          |
|  ┌────────────────────────────────────────────────────┐  │
|  │ 🔍 Search events...                                │  │
│  │ [Event Type ▼] [User ▼] [Date Range ▼] [Severity │  │
│  │ ▼] [🔍 Apply Filters]                              │  │
│  └────────────────────────────────────────────────────┘  │
│                                                          │
│  ┌────────────────────────────────────────────────────┐  │
│  │ Time        User      Event        Host      Result│  │
│  ├────────────────────────────────────────────────────┤  │
│  │ 19:05:22  admin@corp AUTH_LOGIN  —          ✅   │  │
│  │ 19:05:45  admin@corp SESSION_START prod-web-01 🟢 │  │
│  │ 19:06:12  admin@corp HOST_VERIFY   prod-web-02 ⚠️   │  │
│  │ 19:06:15  admin@corp HOST_VERIFY   prod-web-02 ✅   │  │
│  │ 19:10:33  ops@corp   VAULT_UNLOCK  —         ✅   │  │
│  │ 19:12:01  ops@corp   SESSION_END   win-dc-01   🟢   │  │
│  │ 19:15:44  admin@corp HOST_CREATE   lab-kvm-03  ✅   │  │
│  │ 19:22:10  (unknown)  AUTH_FAIL    —          ❌   │  │
│  │ 19:22:11  (unknown)  AUTH_FAIL    —          ❌   │  │
│  │ 19:22:12  (unknown)  AUTH_BANNED  —          🚫   │  │
│  ├────────────────────────────────────────────────────┤  │
│  │ [Load More...]                                     │  │
│  └────────────────────────────────────────────────────┘  │
│                                                          │
+----------------------------------------------------------+
```

### 6.2 Audit Row Detail (Expanded)
```
┌──────────────────────────────────────────────────────────┐
│  19:22:12  (unknown)  AUTH_BANNED  —            🚫       │
├──────────────────────────────────────────────────────────┤
│  🔒 Integrity: Verified (SHA-256: a3f2...)              │
│  📍 IP: 192.168.1.100 · 🔑 Method: password            │
│  ⚠️ Reason: Too many failed attempts (5/5)             │
│  🚫 Ban expires: 19:27:12 (in 5 minutes)               │
│                                                          │
│  Raw Event:                                              │
│  ┌──────────────────────────────────────────────────┐  │
│  │ {                                                  │  │
│  │   "event": "AUTH_BANNED",                         │  │
│  │   "timestamp": "2026-05-27T19:22:12Z",            │  │
│  │   "ip": "192.168.1.100",                          │  │
│  │   "attempts": 5,                                  │  │
│  │   "ban_duration_minutes": 5,                      │  │
│  │   "integrity_hash": "sha256:a3f2b1..."             │  │
│  │ }                                                  │  │
│  └──────────────────────────────────────────────────┘  │
│  [📋 Copy JSON] [🔗 Permalink]                         │
└──────────────────────────────────────────────────────────┘
```

### 6.3 Audit Event Types & Badges

| Event Type | Badge Color | Icon |
|------------|-------------|------|
| AUTH_LOGIN | success | CheckCircle |
| AUTH_FAIL | danger | XCircle |
| AUTH_MFA | info | ShieldCheck |
| AUTH_BANNED | danger | Ban |
| SESSION_START | success | Play |
| SESSION_END | info | Square |
| SESSION_KILL | warning | AlertTriangle |
| HOST_CREATE | info | Plus |
| HOST_UPDATE | info | Pencil |
| HOST_DELETE | danger | Trash2 |
| HOST_VERIFY | warning | ShieldAlert |
| VAULT_UNLOCK | success | Unlock |
| VAULT_LOCK | info | Lock |
| CRED_ACCESS | warning | Key |
| ADMIN_ACTION | warning | Settings |

### 6.4 Mobile Audit View
```
+----------------------------------------------------------+
|  Audit Logs                     [Filter] [Export]        |
+----------------------------------------------------------+
|                                                          |
|  ┌──────────────────────┐                                |
|  │ 19:05:22  admin@corp │                                |
|  │ ✅ AUTH_LOGIN         │                                |
|  └──────────────────────┘                                |
|  ┌──────────────────────┐                                |
|  │ 19:05:45  admin@corp │                                |
|  │ 🟢 SESSION_START      │                                |
|  │ prod-web-01           │                                |
|  └──────────────────────┘                                |
|  ┌──────────────────────┐                                |
|  │ 19:22:12  (unknown)  │                                |
|  │ 🚫 AUTH_BANNED        │                                |
|  │ 192.168.1.100         │                                |
|  │ Too many failed       │                                |
|  │ attempts              │                                │
|  └──────────────────────┘                                |
|                                                          |
+----------------------------------------------------------+
```

- Cards replace table rows on mobile
- Tap to expand and see full JSON details
- Swipe left on card to quick-copy event ID

---

## 7. Empty States

### 7.1 No Hosts
```
+----------------------------------------------------------+
|                                                          |
|                    🖥️                                    |
|            No hosts configured yet                       │
|                                                          |
|       Add your first host to start managing               |
|       your infrastructure securely.                       │
│                                                          │
│         ┌──────────────────────────────┐                │
│         │     🖥️ Add Your First Host    │                │
│         └──────────────────────────────┘                │
│                                                          │
│              [📄 Import from .ssh/config]               │
│                                                          │
+----------------------------------------------------------+
```

### 7.2 No Active Sessions
```
+----------------------------------------------------------+
|                                                          |
|                    📋                                    │
│          No active connections                            │
│                                                          │
│     Select a host from the dashboard or hosts list        │
│     to start a new session.                               │
│                                                          │
│            ┌──────────────────────┐                     │
│            │   🖥️ View Hosts      │                     │
│            └──────────────────────┘                     │
│                                                          │
+----------------------------------------------------------+
```

### 7.3 Audit Log Empty
```
+----------------------------------------------------------+
|                                                          |
|                    📊                                    │
│            No audit events yet                            │
│                                                          │
│     Events will appear here once activity begins.         │
│     All authentication, session, and admin actions        │
│     are logged with integrity verification.               │
│                                                          │
+----------------------------------------------------------+
```

---

## 8. Document Control

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1 | 2026-05-27 | Ame (Designer Subagent) | Initial wireframes |
