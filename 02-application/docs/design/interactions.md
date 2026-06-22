# vexa — Interaction Design

> **Classification:** Security-Critical Application  
> **Owner:** Dhar  > **Status:** Draft — Design Review  > **Last Updated:** 2026-05-27

---

## 1. User Flows

### 1.1 Connect to Host (Primary Flow)

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  Dashboard  │───>│ Select Host │───>│  Protocol   │───>│ Credentials │
│   / Hosts   │    │  from List  │    │  Selector   │    │  (if needed)│
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
                                                              │
                                                              ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  Terminal/  │<───│   Session   │<───│  Security   │<───│   Unlock    │
│   Desktop   │    │  Established│    │   Verified  │    │   Vault     │
│    View     │    │             │    │             │    │  (if locked)│
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
```

**Step-by-Step:**

1. **Select Host** (0 interactions)
   - User clicks/taps host card in Dashboard or Hosts list
   - Host card hover state: `translateY(-1px)`, `border-focus/50`
   - Click: Card pulses once (150ms scale 1→0.98→1), then transition begins

2. **Protocol Selection** (0-1 interaction)
   - If host has single protocol defined: Skip this step
   - If host supports multiple protocols: Inline dropdown appears
   - Default protocol pre-selected (most recently used)

3. **Credentials** (0-2 interactions)
   - If vault is locked: Show vault unlock dialog (see Section 2.2)
   - If saved credential exists: Auto-select, user can override
   - If no saved credential: Inline credential form appears
   - **Security note:** Password fields always masked by default, show toggle

4. **Security Verification** (0-2 interactions)
   - If host key unknown/changed: Host key verification dialog (Section 3.3)
   - If connection requires MFA: Inline MFA prompt (rare for host connect, more for vault)
   - TLS status checked automatically; alert if certificate invalid

5. **Session Established** (system)
   - Terminal/Desktop view slides in from right (desktop) or bottom (mobile)
   - Tab appears in tab bar with pulsing protocol dot
   - Toast notification: "Connected to prod-web-01 via SSH"
   - Audit log entry created automatically

6. **Active Session** (ongoing)
   - Status bar shows live duration, security badges
   - Idle timer starts (configurable)
   - User can minimize (keeps running) or kill (ends session)

### 1.2 Add New Host

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Hosts     │───>│  Click [+]  │───>│ Fill Form   │───>│ Test / Save │
│   Screen    │    │   Button    │    │  (4 fields) │    │             │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
                                                              │
                                        ┌─────────────────────┘
                                        ▼
                              ┌─────────────────┐
                              │  [Test Connection] │
                              │  (optional)         │
                              └─────────────────┘
```

**Form Validation Flow:**

| Field | Validation | Feedback | Timing |
|-------|-----------|----------|--------|
| Display Name | Required, max 64 chars | Inline error | On blur + submit |
| Hostname/IP | DNS resolvable or valid IP | Spinner → Check/Alert | On blur, debounced 500ms |
| Port | 1-65535, default per protocol | Auto-correct | On change |
| Credentials | At least one method | Inline warning | On submit |
| Tags | Optional, max 10 | Tag limit warning | On add |

**Success States:**
- Form valid + [Save]: Dialog closes with 200ms fade, host appears in list
- Toast: "Host prod-web-01 added successfully"
- If "Connect after save" checked: Immediately opens connection (Section 1.1)

### 1.3 Configure MFA (First-Time Setup)

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  Settings   │───>│ Security   │───>│  Choose     │───>│   Setup     │
│             │    │  MFA Mgmt   │    │   Method    │    │  (TOTP/Key) │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
                                                              │
                                                              ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  MFA Active │<───│  Backup     │<───│   Verify    │
│  (Enabled)  │    │   Codes     │    │   (Test)    │
└─────────────┘    └─────────────┘    └─────────────┘
```

**TOTP Setup Detail:**
1. User clicks "Set up Authenticator App"
2. Dialog shows:
   - QR code (centered, 200×200px, with "Can't scan?" text reveal)
   - Secret key (mono, with copy button)
   - Instructions text
3. User scans QR with app
4. Input field for 6-digit verification code
5. On success: Auto-advance to backup codes screen
6. Backup codes displayed as 10 codes in 2×5 grid
7. Each code has individual copy button; [Copy All] button at top
8. User must check "I have saved these codes" to continue
9. MFA status updates across all devices immediately

### 1.4 Kill / Disconnect Session

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Active    │───>│ Click Kill │───>│  Confirm    │───>│  Session    │
│   Session   │    │   Button    │    │   Dialog    │    │   Ended     │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
```

**Two-Step Safety Pattern:**
1. First click on Kill button: Button arms
   - Visual: Background tints danger/10, icon pulses
   - Button text changes: "Kill" → "Click again to confirm"
   - 3-second progress bar appears at button bottom
2. Second click within 3 seconds: Session killed
3. If no second click: Auto-disarm after 3s, smooth fade back
4. **Critical sessions** (production hosts): Require typing host name in confirmation dialog

---

## 2. State Changes

### 2.1 Connection States

| State | Visual | Animation | Sound |
|-------|--------|-----------|-------|
| **Idle** | Host card static | None | None |
| **Initiating** | Spinner on host card, button disabled | Rotate 1s infinite | None |
| **Resolving DNS** | Hostname field shows spinner | Subtle pulse on field | None |
| **Handshaking** | Terminal tab appears with "Connecting..." | Tab slides in from right | None |
| **Authenticating** | "Authenticating..." with protocol icon | Spinner next to icon | None |
| **MFA Required** | Inline prompt or overlay | Fade in 150ms | Subtle ping (if enabled) |
| **Host Key Verify** | ShieldAlert icon, warning dialog | Shake dialog 300ms | Alert chime |
| **Connected** | Pulsing protocol dot, green status | Dot scale pulse 2s infinite | Success chime (optional) |
| **Reconnecting** | Yellow dot, "Reconnecting (1/5)" | Dot changes to spinner | None |
| **Idle Timeout** | Yellow warning badge, countdown | Badge fade in | Warning chime |
| **Disconnected** | Gray dot, strikethrough host name | Dot fades to gray | Disconnect chime |
| **Error** | Red dot, error tooltip | Badge shake 300ms | Error chime |

### 2.2 Vault States

| State | UI Indicator | User Action |
|-------|-------------|-------------|
| **Locked** | 🔐 Lock icon in sidebar, grayed credential fields | Enter master password or biometric |
| **Unlocking** | Spinner on lock icon, "Decrypting..." text | Wait |
| **Unlocked** | 🔓 Unlock icon, credential fields active | None |
| **Auto-lock Warning** | Yellow banner: "Vault locks in 60s" | Click to extend or ignore |
| **Auto-locked** | Lock icon reappears, fields grayed | Re-authenticate |

### 2.3 Host Health States

| State | Dot Color | Tooltip | Action |
|-------|-----------|---------|--------|
| **Online** | Green (solid) | "Reachable — last checked 2m ago" | None |
| **Checking** | Green (pulsing) | "Checking reachability..." | None |
| **Offline** | Red (solid) | "Unreachable — last successful 3h ago" | [Retry Check] |
| **Unknown** | Gray (solid) | "Not checked yet" | [Check Now] |
| **New Host** | Blue (solid) | "Never connected — host key unverified" | [Connect to Verify] |

---

## 3. Keyboard Shortcuts

### 3.1 Global Shortcuts

| Shortcut | Action | Context |
|----------|--------|---------|
| `Ctrl/Cmd + K` | Command palette / quick search | Everywhere |
| `Ctrl/Cmd + Shift + K` | Quick connect to host | Everywhere |
| `Ctrl/Cmd + ,` | Open Settings | Everywhere |
| `Ctrl/Cmd + L` | Lock vault / lock session | Everywhere |
| `Ctrl/Cmd + Shift + N` | Add new host | Everywhere |
| `Ctrl/Cmd + /` | Show keyboard shortcuts help | Everywhere |
| `Ctrl/Cmd + 1-9` | Switch to tab N | Terminal view |
| `Ctrl/Cmd + Tab` | Next tab | Terminal view |
| `Ctrl/Cmd + Shift + Tab` | Previous tab | Terminal view |
| `Ctrl/Cmd + W` | Close current tab | Terminal view |
| `Ctrl/Cmd + Shift + T` | Reopen last closed tab | Terminal view |
| `Ctrl/Cmd + D` | Duplicate current connection | Terminal view |
| `Ctrl/Cmd + +` | Zoom in (terminal/desktop) | Terminal view |
| `Ctrl/Cmd + -` | Zoom out (terminal/desktop) | Terminal view |
| `Ctrl/Cmd + 0` | Reset zoom | Terminal view |
| `Escape` | Close modal / cancel dialog | Modals |
| `Enter` | Primary action in dialog | Modals |
| `Ctrl/Cmd + Shift + F` | Search in current terminal | Terminal view |
| `Ctrl/Cmd + Shift + C` | Copy selection | Terminal view |
| `Ctrl/Cmd + Shift + V` | Paste | Terminal view |

### 3.2 Terminal-Specific Shortcuts

| Shortcut | Action | Notes |
|----------|--------|-------|
| `Ctrl + Shift + C` | Copy to clipboard | Overrides browser default |
| `Ctrl + Shift + V` | Paste from clipboard | |
| `Ctrl + Shift + F` | Find in terminal | Opens search overlay |
| `Ctrl + Shift + K` | Clear terminal | |
| `Ctrl + Shift + S` | Save terminal output | As .txt file |
| `Ctrl + Shift + R` | Rename current tab | Inline edit |
| `Ctrl + Shift + D` | Toggle compact density | More/less padding |
| `Ctrl + Shift + X` | Toggle ligatures | |
| `Ctrl + Shift + M` | Toggle fullscreen | Terminal only |

### 3.3 Navigation Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl/Cmd + Shift + H` | Go to Hosts |
| `Ctrl/Cmd + Shift + S` | Go to Sessions |
| `Ctrl/Cmd + Shift + A` | Go to Audit Logs |
| `Ctrl/Cmd + Shift + D` | Go to Dashboard |
| `Ctrl/Cmd + B` | Toggle sidebar collapse |

### 3.4 Shortcut Help Overlay
```
+----------------------------------------------------------+
|  Keyboard Shortcuts                             [×] [?]  |
|                                                          |
|  Search shortcuts: [________________________]            |
|                                                          |
|  ┌────────────────────────────────────────────────────┐  │
|  │  Global                                              │  │
|  │  ─────────────────────────────────────────           │  │
|  │  Cmd+K          Command palette / quick search       │  │
|  │  Cmd+Shift+K    Quick connect                        │  │
|  │  Cmd+,          Settings                           │  │
|  │  Cmd+L          Lock vault/session                   │  │
|  │                                                      │  │
|  │  Terminal                                            │  │
|  │  ─────────────────────────────────────────           │  │
|  │  Cmd+1-9        Switch to tab 1-9                    │  │
|  │  Cmd+W          Close tab                            │  │
|  │  Cmd+Shift+T    Reopen closed tab                    │  │
|  │  Cmd++          Zoom in                              │  │
|  │                                                      │  │
|  │  [View All 42 Shortcuts →]                           │  │
|  └────────────────────────────────────────────────────┘  │
|                                                          |
+----------------------------------------------------------+
```

---

## 4. Touch Gestures (Mobile/Tablet)

### 4.1 Dashboard/Lists

| Gesture | Action | Feedback |
|---------|--------|----------|
| **Tap** | Select / Open | Ripple effect, 200ms |
| **Long-press** | Context menu | Haptic feedback, card lifts (shadow increases) |
| **Swipe left** | Quick actions reveal | Card slides left, action buttons appear |
| **Swipe right** | Mark favorite / pin | Star icon appears, haptic |
| **Pull down** | Refresh list | Spinner appears, elastic pull |
| **Pinch out** | Zoom in on RDP/VNC | Scale transform |
| **Pinch in** | Zoom out on RDP/VNC | Scale transform |

### 4.2 Terminal View

| Gesture | Action | Feedback |
|---------|--------|----------|
| **Tap** | Move cursor / send click | Local cursor moves |
| **Double-tap** | Select word | Word highlighted |
| **Long-press** | Context menu (copy/paste/select all) | Haptic |
| **Swipe left (edge)** | Next tab | Tab slides in from right |
| **Swipe right (edge)** | Previous tab | Tab slides in from left |
| **Swipe up (bottom)** | Show full keyboard | Keyboard slides up |
| **Swipe down (top)** | Hide keyboard / show status | Keyboard slides down |
| **Two-finger tap** | Paste clipboard | Toast: "Pasted" |
| **Two-finger scroll** | Scroll terminal | Smooth scroll |

### 4.3 RDP/VNC Viewer

| Gesture | Action | Mode |
|---------|--------|------|
| **Single tap** | Left click | Direct |
| **Two-finger tap** | Right click | Direct |
| **Long-press drag** | Drag / select | Direct |
| **Swipe** | Scroll wheel | Direct |
| **Pan** | Move viewport | Trackpad (zoomed >100%) |
| **Pinch** | Zoom in/out | Always |
| **Three-finger swipe up** | Show toolbar | Always |
| **Double-tap with two fingers** | Fit to screen | Always |

---

## 5. Error States & Recovery

### 5.1 Connection Errors

| Error | Cause | Visual | Recovery |
|-------|-------|--------|----------|
| **Host unreachable** | Network failure, firewall | Red dot, "Connection failed" toast with [Retry] | Auto-retry 3× with backoff; manual retry button |
| **DNS resolution failed** | Bad hostname | Hostname field shows XCircle, tooltip with error | Inline edit hostname, re-resolve button |
| **Connection refused** | Wrong port, service down | Error toast, "Connection refused by host" | Check port, verify service status |
| **Authentication failed** | Wrong password/key | Shake password field, "Auth failed" | Retry password, select different credential |
| **Host key changed** | MITM or reinstalled server | Security dialog: ShieldAlert + details | [Verify New Key] [Cancel and Review] |
| **Host key unknown** | First connection | Yellow banner: "Verify host key fingerprint" | [Trust Once] [Trust Always] [Cancel] |
| **Certificate invalid** | Self-signed, expired | Red banner with cert details | [Trust Anyway] (requires admin) [Cancel] |
| **Session timeout** | Idle too long | Yellow banner with countdown | [Extend Session] [Disconnect] |
| **MFA code invalid** | Wrong TOTP | Shake TOTP fields, "Invalid code" | Clear fields, auto-focus first box |
| **Vault locked** | Auto-lock triggered | Gray overlay: "Vault locked" | Unlock button, biometric prompt |

### 5.2 Host Key Verification Dialog

```
+----------------------------------------------------------+
|  ⚠️ Security Warning                          [×]         |
+----------------------------------------------------------+
|                                                          |
|  The host key for prod-web-01 has changed.               │
│                                                          │
│  This could mean:                                        │
│  • The server was reinstalled                            │
│  • You're connecting to a different server                │
│  • A man-in-the-middle attack is in progress             │
│                                                          │
│  Previous fingerprint:                                   │
│  ┌──────────────────────────────────────────────────┐   │
│  │ SHA256:OLDfingerprint1234567890abcdef...         │   │
│  └──────────────────────────────────────────────────┘   │
│                                                          │
│  Current fingerprint:                                    │
│  ┌──────────────────────────────────────────────────┐   │
│  │ SHA256:NEWfingerprint0987654321fedcba...         │   │
│  └──────────────────────────────────────────────────┘   │
│                                                          │
│  Last verified: May 20, 2026 by admin@corp            │
│                                                          │
│  ┌──────────────────────────────────────────────────┐   │
│  │  [Cancel Connection]  [Trust Once]  [Trust Always]│   │
│  └──────────────────────────────────────────────────┘   │
│                                                          │
│  💡 Contact your administrator if you are unsure.       │
│                                                          │
+----------------------------------------------------------+
```

### 5.3 Vault Unlock Error

| Error | Visual | Recovery |
|-------|--------|----------|
| Wrong password | Shake dialog, "Incorrect master password" | Retry, show attempts remaining (3) |
| Biometric failed | Toast: "Biometric authentication failed" | Fallback to password |
| Vault corrupted | Red alert: "Vault integrity check failed" | [Restore from Backup] button |
| Too many attempts | Lockout: "Vault locked for 15 minutes" | Timer countdown, [Contact Admin] |

### 5.4 Network Error Recovery

| Scenario | Behavior |
|----------|----------|
| Client offline | Show offline banner: "No network connection" with [Retry] |
| Host list sync fail | Queue changes locally, show sync pending badge |
| Reconnect | Auto-reconnect with exponential backoff: 1s, 2s, 4s, 8s... max 60s |
| Sync conflict | Dialog: "This host was modified on another device" with [Keep Mine] [Keep Remote] [Merge] |

---

## 6. Security Confirmations

### 6.1 Destructive Actions Hierarchy

| Severity | Action | Confirmation |
|----------|--------|-------------|
| **Low** | Disconnect session | Two-step button (click ×2) |
| **Medium** | Delete host, Remove credential | Dialog + explicit confirm |
| **High** | Delete vault, Revoke MFA, Kill all sessions | Dialog + type confirm keyword |
| **Critical** | Disable audit logging, Change admin role, Reset user MFA | Dialog + type confirm keyword + admin password re-entry |

### 6.2 Two-Step Button Pattern

**Implementation:**
1. **Default state:** Standard danger/secondary button
2. **Armed state (after first click):**
   - Background: `danger/10` with `danger` border
   - Text: "Click again to confirm"
   - Icon: `AlertTriangle` with subtle pulse animation
   - Progress bar: 3-second countdown at bottom of button
3. **Confirmed state:** Action executes immediately
4. **Expired state:** Fades back to default after 3s with 200ms transition

**Visual Timeline:**
```
[  Kill Session  ]   ──click──▶   [  Click again to confirm  ▓▓░░░ ]   ──click──▶   ✓ Session killed
  default                       armed (3s countdown)                              success
```

### 6.3 Keyword Confirmation Dialog

For high/critical severity actions:
```
+----------------------------------------------------------+
|  🗑️ Delete Host                                        │
+----------------------------------------------------------+
│                                                          │
│  This action cannot be undone.                           │
│                                                          │
│  Host: prod-web-01 (10.0.1.4)                          │
│  3 saved credentials will also be deleted.               │
│  2 active sessions will be terminated.                   │
│                                                          │
│  To confirm, type the host name below:                 │
│  ┌──────────────────────────────────────────────────┐   │
│  │ prod-web-01                                      │   │
│  └──────────────────────────────────────────────────┘   │
│  ✅ Input matches                                       │
│                                                          │
│  [Cancel]  [🗑️ Delete Permanently]                       │
│                                                          │
+----------------------------------------------------------+
```

**Rules:**
- Confirm button disabled until exact match
- Case-sensitive for host names
- For admin-critical actions: also require current password re-entry
- Cancel is default action (Escape key, Enter does nothing)

### 6.4 Session Kill Confirmation

```
+----------------------------------------------------------+
|  ⚠️ Terminate Session                                    │
+----------------------------------------------------------+
│                                                          │
│  You are about to kill an active session.                │
│                                                          │
│  Host: prod-web-01                                       │
│  User: admin@corp                                        │
│  Duration: 45 minutes                                    │
│  Protocol: SSH                                           │
│                                                          │
│  Any unsaved work on the remote host will be lost.      │
│                                                          │
│  ┌──────────────────────────────────────────────────┐   │
│  │ [Cancel]              [⚠️ Kill Session]          │   │
│  └──────────────────────────────────────────────────┘   │
│                                                          │
+----------------------------------------------------------+
```

### 6.5 Vault Destruction Confirmation

```
+----------------------------------------------------------+
|  🔥 Destroy Vault                                        │
+----------------------------------------------------------+
│                                                          │
│  ⚠️ WARNING: This is irreversible and destructive.       │
│                                                          │
│  You will permanently lose:                              │
│  • All saved credentials (12 items)                      │
│  • All SSH keys and API tokens                           │
│  • Connection history and preferences                     │
│  • This CANNOT be recovered without a backup             │
│                                                          │
│  To confirm, type "DELETE MY VAULT" below:               │
│  ┌──────────────────────────────────────────────────┐   │
│  │ DELETE MY VAULT                                  │   │
│  └──────────────────────────────────────────────────┘   │
│                                                          │
│  Enter your current password:                            │
│  ┌──────────────────────────────────────────────────┐   │
│  │ ••••••••••                                       │   │
│  └──────────────────────────────────────────────────┘   │
│                                                          │
│  [Cancel]  [🔥 Destroy Permanently]                      │
│                                                          │
+----------------------------------------------------------+
```

---

## 7. Document Control

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1 | 2026-05-27 | Ame (Designer Subagent) | Initial interaction design |
