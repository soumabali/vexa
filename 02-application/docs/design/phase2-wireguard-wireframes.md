# vexa — Phase 2: WireGuard Tunnel Wireframes

> **Feature ID:** FEAT-024
> **Owner:** Designer (Sprint 2.1)
> **Status:** Draft
> **Last Updated:** 2026-06-06

---

## 1. Navigation Update

### 1.1 Sidebar Addition

Insert "Tunnels" between Vault (🔑) and Settings (⚙️):

```
| Icon | Name | Lucide | Route |
|------|------|--------|-------|
| 🔗 | `Waypoints` | `Waypoints` | `/tunnels` |
```

**Desktop sidebar (expanded):**

```
[🏠]  Dashboard
[🖥️]  Hosts
[📋]  Sessions
[🔑]  Vault
[🔗]  Tunnels        ← NEW
[📊]  Audit
[⚙️]  Settings
```

**Mobile bottom bar:** Replace Settings tab with Tunnels; move Settings to overflow or sub-page.

---

## 2. Host Detail — Tunnel Tab (DES-001)

### 2.1 Host Detail Layout with Tabs

The current flat card layout gets a horizontal tab bar added between the header and content area.

```
+------------------------------------------------------------------+
|  Hosts  >  prod-web-01                                     [⋯]  |
|                                                                  |
|  [Overview]  [Sessions]  [Tunnel ●]                              |
|  ┌────────────────────────────────────────────────────────────┐  |
|  │                    Tab Content Below                        │  |
|  └────────────────────────────────────────────────────────────┘  |
+------------------------------------------------------------------+
```

**Tab bar spec:**
- Inherits from `Tabs` component (already in ui library)
- Active tab: bottom-border 2px solid `info` (`#3B82F6`)
- Inactive tab: `text-tertiary`, hover `text-secondary`
- "Tunnel" tab only visible when: host supports WireGuard check succeeds (or always visible with disabled state if WG not available on server)

### 2.2 Tunnel Tab — No Tunnel (Empty State)

Shown when host has no tunnel created yet.

```
+------------------------------------------------------------------+
|  [Tunnel]                                                        |
|                                                                  |
|                       🔗                                         |
|             No WireGuard tunnel configured                       |
|                                                                  |
|     Create a tunnel to connect to this host's network            |
|     securely over an encrypted WireGuard tunnel.                |
|                                                                  |
|  ┌────────────────────────────────────────────────────────────┐  |
|  │  Tunnel Configuration                                     │  |
|  │                                                           │  |
|  │  Allowed IPs (CIDR)        Listen Port                    │  |
|  │  ┌─────────────────────┐  ┌────────────┐                 │  |
|  │  │ 0.0.0.0/0, ::/0     │  │ 51820      │                 │  |
|  │  └─────────────────────┘  └────────────┘                 │  |
|  │                                                           │  |
|  │  DNS Servers (optional, comma-separated)                  │  |
|  │  ┌─────────────────────────────────────────────────────┐  │  |
|  │  │ 1.1.1.1, 8.8.8.8                                   │  │  |
|  │  └─────────────────────────────────────────────────────┘  │  |
|  │                                                           │  |
|  │  □ Use Preshared Key (PSK) for extra layer               │  |
|  │                                                           │  |
|  │  ┌─────────────────────────────────────────────────────┐  │  |
|  │  │           🔗 Create Tunnel                          │  │  |
|  │  └─────────────────────────────────────────────────────┘  │  |
|  └────────────────────────────────────────────────────────────┘  |
|                                                                  |
|  💡 Tunnel config can be exported after creation.                |
|     Client keys are generated locally for security.             |
+------------------------------------------------------------------+
```

**Create tunnel form fields:**

| Field | Type | Default | Validation |
|-------|------|---------|------------|
| Allowed IPs | Text (CIDR list) | `0.0.0.0/0, ::/0` | Valid CIDR notation, max 10 entries |
| Listen Port | Number | 51820 | 1-65535, unique per host |
| DNS Servers | Text (comma-separated) | empty | Valid IP addresses |
| PSK | Checkbox | false | — |

**Actions:**
- [Create Tunnel]: Calls `POST /api/v1/tunnels`, shows loading spinner, transitions to active state
- Cancel: Not needed (tab switch = abandon)

### 2.3 Tunnel Tab — Active Tunnel

Shown when a tunnel exists for this host.

```
+------------------------------------------------------------------+
|  [Tunnel]                                                        |
|                                                                  |
|  ┌────────────────────────────────────────────────────────────┐  |
|  │  🔗 wg0    ● Connected   10.200.200.2 → 10.200.200.1     │  |
|  │                                                           │  |
|  │  ┌─────────────┐  ┌─────────────┐  ┌───────────────────┐  │  |
|  │  │ ⬇ 1.2 GB    │  │ ⬆ 342 MB   │  │ 🕐 14:32:05       │  │  |
|  │  │ Received     │  │ Sent         │  │ Last handshake    │  │  |
|  │  └─────────────┘  └─────────────┘  └───────────────────┘  │  |
|  │                                                           │  |
|  │  ┌────────────────────────────────────────────────────────┤  |
|  │  │ Details                                                │  │
|  │  │ ─────────────────────────────────────────               │  │
|  │  │ Interface      wg0                                    │  │
|  │  │ Server IP      10.200.200.1/24                        │  │
|  │  │ Client IP      10.200.200.2/24                        │  │
|  │  │ Listen Port    51820 (UDP)                            │  │
|  │  │ Allowed IPs    0.0.0.0/0                              │  │
|  │  │                 ::/0                                   │  │
|  │  │ DNS Servers    1.1.1.1, 8.8.8.8                      │  │
|  │  │ PSK            Enabled                                │  │
|  │  │ Created        2026-06-06 10:30 UTC                    │  │
|  │  │ Last Rotated   2026-06-06 10:30 UTC                    │  │
|  │  └────────────────────────────────────────────────────────┘  │
|  │                                                           │  │
|  │  ┌────────────────────────────────────────────────────────┤  │
|  │  │ Actions                                                │  │
|  │  │ ─────────────────────────────────────────               │  │
|  │  │                                                         │  |
|  │  │ [● Toggle: Enabled]  [🔄 Rotate Keys]  [📋 Export]     │  │
|  │  │                         [🗑 Delete Tunnel]              │  │
|  │  └────────────────────────────────────────────────────────┘  │
|  └────────────────────────────────────────────────────────────┘  |
+------------------------------------------------------------------+
```

**Status badge** (top of tunnel card):
- **🟢 Connected** — `success` (#10B981), handshake within last 2 minutes
- **🔴 Disconnected** — `danger` (#EF4444), no handshake > 2 minutes
- **🟡 Error** — `warning` (#F59E0B), interface exists but peer not responding
- **⚪ Disabled** — `text-tertiary` (#64748B), interface down by user action
- **🔄 Creating** — `info` (#3B82F6), spinner animation

**Traffic stats row:**
- Three stat mini-cards in horizontal row
- Updates via WebSocket `tunnel:stats` (every 5s when tab active, every 30s when backgrounded)
- Values in human-readable format (bytes → KB/MB/GB)

**Actions bar:**
- [Toggle]: `Switch` component, ON = enabled, OFF = disabled (calls enable/disable API)
- [Rotate Keys]: Two-step safety button pattern (see interactions doc)
- [Export]: Opens config export modal (see DES-002)
- [Delete Tunnel]: Opens security confirmation dialog with hostname typing

### 2.4 Delete Tunnel Confirmation

```
+------------------------------------------------------------------+
|  🗑️ Delete Tunnel                                    [×]         |
+------------------------------------------------------------------+
|                                                                  |
|  ⚠️ This will permanently remove the WireGuard tunnel            |
|     and destroy the interface on the server.                     |
|                                                                  |
|  Host:        prod-web-01                                        |
|  Interface:   wg0                                                |
|  Client IP:   10.200.200.2/24                                    |
|  Created:     2026-06-06 10:30 UTC                              |
|                                                                  |
|  Active connections through this tunnel will be interrupted.    |
|                                                                  |
|  To confirm, type the host name below:                          |
|  ┌──────────────────────────────────────────────────────────┐   |
|  │ prod-web-01                                              │   |
|  └──────────────────────────────────────────────────────────┘   |
|  ✅ Input matches                                              |
|                                                                  |
|  ┌──────────────────────────────────────────────────────────┐   |
|  │  [Cancel]                    🗑️ Delete Tunnel            │   |
|  └──────────────────────────────────────────────────────────┘   |
|                                                                  |
+------------------------------------------------------------------+
```

- Follows `Security Confirmation Dialog` pattern from design system
- Confirm disabled until `host.name` exact match
- Destructive action: can only delete if tunnel is disabled (or force with additional warning)

### 2.5 Mobile Tunnel Tab

```
+------------------------------------------------------------------+
|  [←]  prod-web-01                        [Edit] [Delete]         |
+------------------------------------------------------------------+
|  [Overview]  [Sessions]  [Tunnel ●]                              |
+------------------------------------------------------------------+
|  🔗 wg0                                        ● Connected       |
|                                                                  |
|  ⬇ 1.2 GB  │  ⬆ 342 MB  │  🕐 14:32:05                        |
|                                                                  |
|  ┌─ Details ─────────────────────────────────────────────────┐  |
|  │ Interface      wg0                                        │  |
|  │ Client IP      10.200.200.2/24                            │  |
|  │ Server IP      10.200.200.1/24                            │  |
|  │ Port           51820                                      │  |
|  │ Allowed IPs    0.0.0.0/0, ::/0                           │  |
|  │ DNS            1.1.1.1                                    │  |
|  └──────────────────────────────────────────────────────────┘  |
|                                                                  |
|  [● Enabled]   [🔄 Rotate]   [📋 Export]                        |
|                                                                  |
+------------------------------------------------------------------+
```

- Stats row: horizontal scrollable if overflow
- Actions: full-width buttons stacked vertically on very small screens
- Bottom sheet for config export on mobile

---

## 3. Tunnel List Page (DES-002)

### 3.1 Desktop Layout

```
+------------------------------------------------------------------+
|  Sidebar |  Tunnels                                      [+ New] |
|          |                                                       |
| [🏠]     |  ┌─────────────────────────────────────────────────┐  |
| [🖥️]     |  │ 🔍 Search tunnels...    [Host ▼]  [Status ▼]   │  |
| [📋]     |  └─────────────────────────────────────────────────┘  |
| [🔑]     |                                                       |
| [🔗]     |  ┌─────────────────────────────────────────────────┐  |
| [📊]     |  │ Interface  Host         Client IP      Status   │  |
| [⚙️]     |  ├─────────────────────────────────────────────────┤  |
|          |  │ 🔗 wg0  │ prod-web-01 │ 10.200.200.2 │ 🟢 Up   │  |
|          |  │          │ 51820       │ 1.2 GB/342 MB│ 14:32   │  |
|          |  ├─────────────────────────────────────────────────┤  |
|          |  │ 🔗 wg1  │ dev-db-01   │ 10.200.200.3 │ 🔴 Down │  |
|          |  │          │ 51821       │ 0 B/0 B      │ —       │  |
|          |  ├─────────────────────────────────────────────────┤  |
|          |  │ 🔗 wg2  │ staging-web │ 10.200.200.4 │ 🟢 Up   │  |
|          |  │          │ 51822       │ 512 MB/128 MB│ 08:15   │  |
|          |  └─────────────────────────────────────────────────┘  |
|          |                                                       |
|          |  Showing 3 tunnels · 7 remaining of 10 max           |
|          |                                                       |
+------------------------------------------------------------------+
```

**Table columns:**

| Column | Width | Content |
|--------|-------|---------|
| Interface | 80px | `mono`, icon + name |
| Host | Flexible | Host name (link), port below in `caption` |
| Client IP | 180px | `mono`, traffic stats below in `caption` |
| Status | 120px | Badge + label, handshake time below |
| Actions | 60px | `[⋯]` dropdown menu |

**Row click:** Navigate to host detail → Tunnel tab.

**Row actions (`[⋯]` menu):**
- View Host
- Export Config
- Rotate Keys
- Enable/Disable
- Delete

**Empty state:**
```
+------------------------------------------------------------------+
|                       🔗                                         |
|             No WireGuard tunnels configured                      |
|                                                                  |
|     Create a tunnel from a host's detail page to get             |
|     started. Tunnels provide secure encrypted access.            |
|                                                                  |
|  💡 Go to any host and open the "Tunnel" tab to create one.     |
+------------------------------------------------------------------+
```

### 3.2 Mobile Tunnel List

```
+------------------------------------------------------------------+
|  [≡]  Tunnels                                          [+ New]  |
+------------------------------------------------------------------+
|  🔍 [Search tunnels...]                                         |
|  [Host ▼]  [Status ▼]                                           |
+------------------------------------------------------------------+
|  ┌──────────────────────────────────────────────────────────┐   |
|  │ 🔗 wg0  prod-web-01               🟢 Up                   │  |
|  │    10.200.200.2 · 1.2 GB/342 MB  14:32                    │  |
|  │    [📋 Export] [⋯]                                        │  |
|  └──────────────────────────────────────────────────────────┘   |
|  ┌──────────────────────────────────────────────────────────┐   |
|  │ 🔗 wg1  dev-db-01                   🔴 Down               │  |
|  │    10.200.200.3 · 0 B/0 B           —                     │  |
|  │    [📋 Export] [⋯]                                        │  |
|  └──────────────────────────────────────────────────────────┘   |
|                                                                  |
|  3 tunnels · 7 remaining of 10 max                               |
+------------------------------------------------------------------+
```

- Card-based layout per breakpoint rules (< 1024px)
- Swipe left: reveal [Export] [Delete]
- Long-press: context menu

---

## 4. Config Export Modal (DES-002)

### 4.1 Desktop Config Modal

```
+------------------------------------------------------------------+
|  🔗 Tunnel Configuration                           [×]           |
+------------------------------------------------------------------+
|                                                                  |
|  Host: prod-web-01   Interface: wg0   Client: 10.200.200.2/24   |
|                                                                  |
|  ┌────────────────────────────────────────────────────────────┐  |
|  │ [Interface]                                               │  |
|  │ PrivateKey = <client_private_key>                         │  |
|  │ Address = 10.200.200.2/24                                 │  |
|  │ DNS = 1.1.1.1, 8.8.8.8                                  │  |
|  │                                                           │  |
|  │ [Peer]                                                    │  |
|  │ PublicKey = <server_public_key>                           │  |
|  │ PresharedKey = <preshared_key>                           │  |
|  │ AllowedIPs = 0.0.0.0/0, ::/0                             │  |
|  │ Endpoint = vpn.corp.example.com:51820                     │  |
|  │ PersistentKeepalive = 25                                  │  |
|  └────────────────────────────────────────────────────────────┘  |
|                                                                  |
|  ⚠️ Private key shown only on the device where it was            |
|     generated. Store it securely.                                |
|                                                                  |
|  ┌──────────────┐  ┌───────────────┐  ┌──────────────────────┐  |
|  │ 📋 Copy       │  │ ⬇ Download    │  │ 📱 Share to App     │  |
|  └──────────────┘  └───────────────┘  └──────────────────────┘  |
|                                                                  |
|  ──────────────── OR scan QR code on mobile ──────────────────  |
|                                                                  |
|          ┌────────────────────────────────────┐                 |
|          │                                    │                 |
|          │           [QR Code]                │                 |
|          │      200×200 px, error-correct     │                 |
|          │                                    │                 |
|          └────────────────────────────────────┘                 |
|                                                                  |
|  💡 Open your WireGuard mobile app and tap + to scan.           |
|                                                                  |
+------------------------------------------------------------------+
```

**Modal spec:**
- Max-width: 640px (wide dialog per design system)
- .conf viewer: `bg-inset`, `mono` font, 13px, 16px padding, border-radius 6px
- QR code: 200×200px, error correction level M (15%), centered
- Actions row: 3 buttons in horizontal layout, wrap to stack on mobile
- Warning banner above actions if `.conf` contains private key

### 4.2 Mobile Config Modal (Bottom Sheet)

```
+------------------------------------------------------------------+
|  ═══  (drag handle)                                              |
+------------------------------------------------------------------+
|  Tunnel Configuration                                             |
|                                                                  |
|  prod-web-01 · wg0 · 10.200.200.2/24                            |
|                                                                  |
|  ┌────────────────────────────────────────────────────────────┐  |
|  │ [Interface]     Copy ▲                                      │  |
|  │ PrivateKey = ...        [📋]                                │  |
|  │ Address = 10.200.200.2/24                                   │  |
|  │ DNS = 1.1.1.1                                               │  |
|  │ [Peer]         Copy ▲                                       │  |
|  │ PublicKey = ...        [📋]                                  │  |
|  │ ...                                                         │  |
|  └────────────────────────────────────────────────────────────┘  |
|                                                                  |
|  ┌────────────┐  ┌──────────────────────┐                      │
|  │ 📋 Copy All │  │ ⬇ Download .conf     │                      │
|  └────────────┘  └──────────────────────┘                      │
|                                                                  |
|  ┌──────────────────────────────────────────────────────────┐   │
|  │  📱 Open in WireGuard App                                │   │
|  └──────────────────────────────────────────────────────────┘   |
|                                                                  |
|  ┌──────────────────────────────────────────────────────────┐   │
|  │  📸 Show QR Code for Mobile Import                       │   |
|  └──────────────────────────────────────────────────────────┘   |
|                                                                  |
+------------------------------------------------------------------+
```

- Bottom sheet slides up (per responsive rules for < 640px)
- QR code on separate full-screen view when tapped
- "Open in WireGuard App" triggers `wireguard://` deep link

---

## 5. Component Specifications

### 5.1 Tunnel Status Badges

```
┌─────────────────────┐  ┌─────────────────────┐
│  ● Connected        │  │  ● Disconnected     │
│  bg-success/10      │  │  bg-danger/10       │
│  text-success       │  │  text-danger        │
└─────────────────────┘  └─────────────────────┘

┌─────────────────────┐  ┌─────────────────────┐
│  ● Error            │  │  ○ Disabled         │
│  bg-warning/10      │  │  bg-surface         │
│  text-warning       │  │  text-tertiary      │
└─────────────────────┘  └─────────────────────┘

┌─────────────────────┐
│  ◌ Creating...      │
│  bg-info/10         │
│  text-info          │
│  (spinner anim)     │
└─────────────────────┘
```

- Height: 24px, horizontal padding 8px (space-2), border-radius 6px
- Left dot: 6px circle, pulsing only for Connected state (2s pulse)
- Text: `caption` (12px), weight 500

### 5.2 Stats Display Component

```
┌──────────────────────────────────────────────────────────────────┐
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│  │ ⬇ 1.2 GB    │  │ ⬆ 342 MB    │  │ 🕐 14:32:05          │  │
│  │ Received     │  │ Sent         │  │ Last Handshake        │  │
│  └──────────────┘  └──────────────┘  └──────────────────────┘  │
└──────────────────────────────────────────────────────────────────┘
```

- Three mini stat cards in a row
- Each card: icon (16px) + value (`mono`, 16px, weight 600) + label (`caption`)
- Background: `bg-surface`, border: `border-subtle`, border-radius: 6px

### 5.3 Config Copy Button

```
┌────────────────────┐
│  📋 Copy Config    │
│  bg-surface        │
│  border-default    │
│  text-primary      │
│  Hover: bg-surface-│
│  raised            │
└────────────────────┘

Copy success state (1.5s):
┌────────────────────┐
│  ✅ Copied!        │
│  bg-success/10     │
│  border-success    │
│  text-success      │
└────────────────────┘
```

- Secondary button style, or ghost with icon
- `aria-label`: "Copy tunnel configuration to clipboard"
- On click: copy to clipboard, show "Copied!" toast + button visual feedback for 1.5s

### 5.4 QR Code Component

```
┌────────────────────────────────────┐
│                                    │
│         ██████████████            │
│         ██  QR CODE  ██           │
│         ██████████████            │
│         ██ ██  ████ ██           │
│         ██  ████  ████           │
│         ██████████████            │
│                                    │
│      200×200px (desktop)          │
│      280×280px (mobile modal)     │
│                                    │
└────────────────────────────────────┘
```

- Container: `bg-white` (QR needs light background), border-radius 12px, padding 16px
- Size: 200×200px desktop, 280×280px mobile (touch target for scanning)
- Loading state: Skeleton 200×200px gray box with spinner
- Error state: AlertTriangle icon + "Failed to generate QR code" + [Retry]
- Caption below: "Scan with WireGuard mobile app"
- Generated server-side via `go-qrcode` library, returned as PNG base64 or blob URL

---

## 6. Route Structure

| Route | Screen | Layout |
|-------|--------|--------|
| `/tunnels` | Tunnel list | Full page with sidebar |
| `/tunnels/[id]` | Tunnel detail (redirects to host Tunnel tab) | Redirect |
| `hosts/[id]` (with tab param) | Host detail Tunnel tab | `#tunnel` hash or `?tab=tunnel` |

No separate tunnel detail page — tunnel details are viewed on the host detail "Tunnel" tab and the config export modal.

---

## 7. Document Control

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1 | 2026-06-06 | Designer Subagent | Initial Phase 2 WireGuard wireframes |
