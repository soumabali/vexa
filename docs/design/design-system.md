# vexa — Design System

> **Classification:** Security-Critical Application  
> **Owner:** Dhar  
> **Status:** Draft — Design Review  
> **Last Updated:** 2026-05-27

---

## 1. Design Philosophy

### 1.1 Core Principles
1. **Content First:** Chrome is minimal; the terminal/remote desktop is the hero
2. **Security at a Glance:** Every connection's security posture is immediately visible
3. **Trust Through Transparency:** Audit trails, encryption status, and session metadata are always accessible
4. **Power User Friendly:** Extensive keyboard shortcuts; zero-click workflows for common actions
5. **Calm Under Pressure:** Error states and security alerts are clear but never alarming

### 1.2 Visual Language
- **Density:** Medium-high (information-dense tools for sysadmins)
- **Shapes:** Rounded rectangles (4-8px) for UI chrome; sharp corners (0-2px) for terminal panels
- **Depth:** Flat with subtle layered shadows for elevation; no glassmorphism
- **Motion:** Fast (150-250ms), functional, never decorative

---

## 2. Color Palette

### 2.1 Base Palette (Dark Mode — Primary)

| Token | Hex | Usage |
|-------|-----|-------|
| `bg-base` | `#0B0F19` | Main background |
| `bg-surface` | `#111827` | Cards, panels, elevated surfaces |
| `bg-surface-raised` | `#1A2235` | Modals, popovers, dropdowns |
| `bg-inset` | `#070A10` | Terminal backgrounds, code blocks |
| `border-subtle` | `#1E293B` | Dividers, inactive borders |
| `border-default` | `#334155` | Input borders, table borders |
| `border-focus` | `#3B82F6` | Focus rings |

### 2.2 Text Colors (Dark)

| Token | Hex | Usage |
|-------|-----|-------|
| `text-primary` | `#F1F5F9` | Headings, primary text |
| `text-secondary` | `#94A3B8` | Labels, metadata, placeholders |
| `text-tertiary` | `#64748B` | Timestamps, disabled state |
| `text-inverse` | `#0B0F19` | Text on colored buttons |

### 2.3 Semantic Colors (Security-Critical)

| State | Token | Hex | Light Token | Light Hex |
|-------|-------|-----|-------------|-----------|
| **Secure / Connected** | `success` | `#10B981` | `success-light` | `#059669` |
| **Warning / Unverified** | `warning` | `#F59E0B` | `warning-light` | `#D97706` |
| **Danger / Insecure** | `danger` | `#EF4444` | `danger-light` | `#DC2626` |
| **Info / Neutral** | `info` | `#3B82F6` | `info-light` | `#2563EB` |
| **MFA Required** | `mfa` | `#8B5CF6` | `mfa-light` | `#7C3AED` |

### 2.4 Security State Indicators

Every connection/host/session MUST display a security badge:

| Badge | Color | Icon | Meaning |
|-------|-------|------|---------|
| 🔒 **Encrypted** | `success` + `bg-base` | ShieldCheck | TLS 1.3 active, key verified |
| 🔐 **MFA Verified** | `mfa` | Fingerprint | User authenticated with 2FA |
| ⚠️ **Unverified Host** | `warning` | AlertTriangle | New/ changed host key |
| ❌ **Insecure** | `danger` | ShieldX | Plaintext or failed TLS |
| 🕐 **Session Expiring** | `warning` | Clock | Idle timeout approaching |
| ⏸️ **Locked** | `info` | Lock | Session locked, auth required |

### 2.5 Protocol Colors

| Protocol | Accent | Hex | Usage |
|----------|--------|-----|-------|
| SSH | `protocol-ssh` | `#10B981` | Badges, tab indicators, activity dots |
| RDP | `protocol-rdp` | `#3B82F6` | Badges, tab indicators, activity dots |
| VNC | `protocol-vnc` | `#F59E0B` | Badges, tab indicators, activity dots |

### 2.6 Light Mode (Optional)

| Token | Hex |
|-------|-----|
| `bg-base-light` | `#F8FAFC` |
| `bg-surface-light` | `#FFFFFF` |
| `bg-inset-light` | `#F1F5F9` |
| `text-primary-light` | `#0F172A` |
| `text-secondary-light` | `#475569` |
| `border-subtle-light` | `#E2E8F0` |
| `border-default-light` | `#CBD5E1` |

**Note:** Light mode inverts surface hierarchy but keeps semantic colors identical. Terminal always renders dark (`bg-inset`) regardless of mode.

---

## 3. Typography

### 3.1 Font Stack

```css
--font-sans: 'Inter', system-ui, -apple-system, sans-serif;
--font-mono: 'JetBrains Mono', 'Fira Code', 'SF Mono', monospace;
--font-display: 'Inter', sans-serif; /* For large numbers/stats */
```

### 3.2 Type Scale

| Token | Size | Weight | Line-Height | Letter-Spacing | Usage |
|-------|------|--------|-------------|----------------|-------|
| `display` | 32px | 700 | 1.1 | -0.02em | Empty state headings |
| `h1` | 24px | 600 | 1.2 | -0.01em | Page titles |
| `h2` | 20px | 600 | 1.25 | -0.01em | Section headers |
| `h3` | 16px | 600 | 1.3 | 0 | Card titles, dialog titles |
| `h4` | 14px | 600 | 1.35 | 0 | Subsection, table headers |
| `body` | 14px | 400 | 1.5 | 0 | Primary body text |
| `body-sm` | 13px | 400 | 1.5 | 0 | Secondary text, descriptions |
| `caption` | 12px | 500 | 1.4 | 0.01em | Labels, timestamps, badges |
| `mono` | 13px | 400 | 1.6 | 0 | Code, terminal, logs |
| `mono-sm` | 11px | 400 | 1.5 | 0 | Compact log output |

### 3.3 Terminal Typography
- **Font:** `JetBrains Mono` (ensures box-drawing alignment)
- **Size:** User-configurable 10-20px (default 13px)
- **Line-height:** 1.2 (tight for density)
- **Ligatures:** Optional (user toggle in Settings)
- **Anti-aliasing:** Subpixel on LCD; grayscale on HiDPI

---

## 4. Spacing System

### 4.1 Base Unit: 4px

| Token | Value | Usage |
|-------|-------|-------|
| `space-1` | 4px | Icon padding, tight gaps |
| `space-2` | 8px | Inline spacing, badge padding |
| `space-3` | 12px | Small component padding |
| `space-4` | 16px | Standard padding, card gaps |
| `space-5` | 20px | Form section spacing |
| `space-6` | 24px | Dialog padding, list gaps |
| `space-8` | 32px | Section separation |
| `space-10` | 40px | Page-level margins |
| `space-12` | 48px | Large layout gaps |

### 4.2 Layout Grid
- **Desktop:** 12-column, 24px gutter, max-width 1440px
- **Tablet:** 8-column, 20px gutter
- **Mobile:** 4-column, 16px gutter
- **Terminal:** No grid — full-bleed content area

---

## 5. Component Library

### 5.1 Buttons

#### Primary Button
```
Background:    info (#3B82F6)
Text:          text-inverse (#0B0F19)
Padding:       space-3 space-4 (12px 16px)
Border-radius: 6px
Font:          body, weight 500
Hover:         bg darken 10%, translateY(-1px), shadow-sm
Active:        bg darken 15%, translateY(0)
Disabled:      opacity 50%, cursor not-allowed
Focus:         ring-2 ring-info ring-offset-2 ring-offset-bg-base
```

#### Secondary Button
```
Background:    transparent
Border:        1px solid border-default
Text:          text-primary
Hover:         bg-surface-raised, border-focus
```

#### Danger Button
```
Background:    danger (#EF4444)
Text:          text-inverse
Hover:         bg darken 10%
Focus:         ring-2 ring-danger
```

#### Ghost Button (Icon-only)
```
Background:    transparent
Text:          text-secondary
Size:          32px x 32px (touch target)
Border-radius: 6px
Hover:         bg-surface, text-primary
Active:        bg-surface-raised
```

#### Security Action Button
Special variant for destructive security actions (disconnect, kill session, revoke access):
```
Requires:      Two-step activation (click once to arm, click again to confirm within 3s)
Armed state:   bg-danger/10, border-danger, text-danger, icon pulses 1s
Confirm:       Full danger button style
Timeout:       Auto-disarm after 3s with fade transition
```

### 5.2 Inputs

#### Text Input
```
Background:    bg-base
Border:        1px solid border-default
Border-radius: 6px
Padding:       space-3 (12px)
Font:          body
Focus:         border-focus, ring-2 ring-focus/20
Error:         border-danger, ring-2 ring-danger/20
Placeholder:   text-tertiary
Disabled:      opacity 60%, bg-surface
```

#### Secure Input (Passwords, Keys)
```
Inherits:      Text Input base
Extra:
  - Toggle visibility button (eye icon) at right
  - Strength indicator bar below (if applicable)
  - Copy button for generated keys
  - "Show as QR" option for TOTP secrets
```

#### Host Address Input
```
Inherits:      Text Input base
Extra:
  - Protocol selector dropdown prepended (SSH/RDP/VNC)
  - Port field inline (auto-populates default per protocol)
  - Validation icon: spinner while resolving, checkmark/alert on result
  - Reachability indicator dot (green = pingable, red = timeout, gray = unknown)
```

### 5.3 Cards

#### Host Card
```
Background:    bg-surface
Border:        1px solid border-subtle
Border-radius: 8px
Padding:       space-4 (16px)
Shadow:        none (flat)
Hover:         border-focus/50, translateY(-1px), shadow-sm
Selected:      border-focus, ring-1 ring-focus, bg-surface-raised
Layout:        Horizontal — protocol icon | host info | status | quick-actions
```

#### Session Card
```
Inherits:      Host Card base
Extra:
  - Live indicator: 8px pulsing dot in protocol color (animation: pulse 2s infinite)
  - Duration timer: caption mono, updating every second
  - Security badge: top-right corner
  - "Kill" action: ghost button with skull icon, requires two-step
```

### 5.4 Modals / Dialogs

#### Standard Dialog
```
Overlay:       bg-base/80, backdrop-blur-sm
Panel:         bg-surface-raised, border-subtle
Border-radius: 12px
Padding:       space-6 (24px)
Shadow:        shadow-xl
Max-width:     480px (standard), 640px (wide), 960px (xl)
Animation:     Fade in 150ms + scale from 0.96 to 1.00
```

#### Security Confirmation Dialog
```
Inherits:      Standard Dialog base
Extra:
  - Icon: Large (48px) in semantic color at top center
  - Title: h2 in semantic color
  - Body: Clear description of irreversible action
  - Input: Must type confirm keyword (e.g., host name) for destructive actions
  - Actions: Secondary "Cancel" left, Danger "Confirm" right
  - Escape key: Closes (equivalent to Cancel)
```

#### MFA Challenge Dialog
```
Inherits:      Standard Dialog base
Extra:
  - Step indicator: 1-2-3 dots at top
  - Large TOTP input: 6 boxes, auto-focus, auto-advance
  - WebAuthn button: Prominent with fingerprint/key icon
  - Backup codes: Collapsible section
  - Timer: 30s countdown for TOTP validity
```

### 5.5 Terminal Panel

```
Background:    bg-inset (#070A10)
Border:        none (flush with container)
Padding:       space-4 (16px) on all sides
Font:          mono, 13px, line-height 1.2
Scrollbar:     8px width, bg-transparent, thumb border-default/50
Selection:     info/30 (blue-tinted highlight)
Cursor:        Block or bar (user preference), blinks at 530ms
```

#### Terminal Tab
```
Height:        36px
Background:    bg-surface (inactive) / bg-inset (active)
Border:        1px solid border-subtle, bottom-border transparent on active
Border-radius: 6px 6px 0 0
Padding:       space-2 space-3
Content:       Protocol dot (6px) + Host name (truncated) + Close button (hover reveal)
Unread badge:  16px circle with count, bg-danger
```

### 5.6 RDP/VNC Viewer Panel

```
Background:    bg-inset
Border:        none
Cursor:        Remote cursor rendered via canvas (local cursor hidden)
Toolbar:       Floating, auto-hide after 3s inactivity
  - Position: Bottom-center
  - Background: bg-surface-raised/90, backdrop-blur
  - Height: 48px, border-radius 24px
  - Buttons: Disconnect, Fullscreen, Clipboard, Keyboard, Zoom
```

### 5.7 Navigation

#### Sidebar (Desktop)
```
Width:         64px (collapsed) / 240px (expanded)
Background:    bg-surface
Border-right:  1px solid border-subtle
Items:         Icon + Label (expanded) / Icon only (collapsed)
Active item:   Left border 3px solid info, bg-surface-raised
Hover:         bg-surface-raised
Transition:    width 250ms ease-in-out
```

#### Top Bar (Mobile)
```
Height:        56px
Background:    bg-surface
Border-bottom: 1px solid border-subtle
Content:       Logo left, contextual actions right, title center
Shadow:        shadow-sm on scroll
```

#### Bottom Bar (Mobile)
```
Height:        64px (includes safe area)
Background:    bg-surface
Border-top:    1px solid border-subtle
Items:         4-5 tabs, icon + label
Active:        Icon in info color, label in text-primary
```

### 5.8 Data Tables (Audit Logs, Host Lists)

```
Header:        h4, text-secondary, bg-surface, border-bottom border-subtle
Row:           body, text-primary, padding space-3 space-4
Stripe:        Even rows bg-base/50
Hover:         bg-surface-raised/50
Selected:      bg-info/10
Cell:          First cell left-aligned, others as appropriate
Sortable:      Header has up/down arrow, active column has info color
```

#### Audit Log Row Special
```
Extra:
  - Event type badge: semantic color (AUTH=info, SESSION=success, ADMIN=warning, SECURITY=danger)
  - Integrity indicator: Lock icon if HMAC verified, AlertTriangle if hash mismatch
  - Expandable: Click to see full JSON details in mono font
  - User avatar + name always visible
```

### 5.9 Toasts & Notifications

```
Position:      Top-right desktop, bottom-center mobile
Max-width:     400px
Background:    bg-surface-raised
Border:        1px solid border-subtle
Border-radius: 8px
Shadow:        shadow-lg
Padding:       space-3 space-4
Animation:     Slide in from right (desktop) / bottom (mobile) 200ms + fade
Auto-dismiss:  5s (unless action required)
Types:
  - Success: Left border 3px success + icon
  - Error:   Left border 3px danger + icon (no auto-dismiss)
  - Warning: Left border 3px warning + icon
  - Info:    Left border 3px info + icon
```

### 5.10 Tooltips

```
Background:    bg-surface-raised
Text:          caption, text-primary
Border:        1px solid border-subtle
Border-radius: 6px
Padding:       space-2 space-3
Arrow:         6px, matching background
Delay:         300ms hover before show
```

---

## 6. Iconography

### 6.1 Icon Set: Lucide-React
All icons from [Lucide](https://lucide.dev/) for consistency, clarity, and security-appropriate semantics.

### 6.2 Security Icons (Critical)

| Icon | Name | Usage |
|------|------|-------|
| 🔒 | `Shield` | Base security icon |
| 🔒✓ | `ShieldCheck` | Verified secure connection |
| 🔒✗ | `ShieldX` | Insecure / failed verification |
| 🔒⚡ | `ShieldAlert` | Security warning |
| 🔑 | `Key` | Credentials, SSH keys |
| 🔐 | `Lock` | Locked session, vault |
| 🔓 | `Unlock` | Unlocked vault |
| 👁️ | `Eye` | Show password / visibility toggle |
| 👁️‍🗨️ | `EyeOff` | Hide password |
| ✓ | `CheckCircle` | Success, verified |
| ✗ | `XCircle` | Error, failed |
| ⚠️ | `AlertTriangle` | Warning, unverified |
| 🖐️ | `Fingerprint` | Biometric, WebAuthn |
| 📱 | `Smartphone` | TOTP, mobile device |
| 🔢 | `Hash` | Backup codes |

### 6.3 Navigation Icons

| Icon | Name | Screen |
|------|------|--------|
| 🏠 | `LayoutDashboard` | Dashboard |
| 🖥️ | `Monitor` | Hosts |
| 📋 | `Terminal` | Active Sessions |
| 📊 | `BarChart3` | Audit Logs |
| ⚙️ | `Settings` | Settings |
| 🔑 | `KeyRound` | Vault |
| 👤 | `User` | Profile |
| 🚪 | `LogOut` | Sign Out |

### 6.4 Protocol Icons

| Protocol | Icon | Color |
|----------|------|-------|
| SSH | `Terminal` | success (#10B981) |
| RDP | `Monitor` | info (#3B82F6) |
| VNC | `Cast` | warning (#F59E0B) |

### 6.5 Icon Sizes

| Context | Size | Stroke Width |
|---------|------|--------------|
| Inline with text | 16px | 2px |
| Buttons / table cells | 18px | 2px |
| Navigation items | 20px | 2px |
| Empty state illustrations | 48px | 1.5px |
| Security status badges | 20px | 2px |

---

## 7. Animation & Motion

### 7.1 Principles
- **Purposeful:** Every animation communicates state change
- **Fast:** Never slower than 300ms for UI feedback
- **Subtle:** Easing should feel natural, not bouncy

### 7.2 Timing Tokens

| Token | Duration | Easing | Usage |
|-------|----------|--------|-------|
| `instant` | 0ms | — | Terminal output, real-time data |
| `fast` | 100ms | ease-out | Button states, icon toggles |
| `normal` | 150ms | ease-in-out | Color transitions, opacity |
| `smooth` | 200ms | cubic-bezier(0.4, 0, 0.2, 1) | Panel slides, modals |
| `gentle` | 300ms | cubic-bezier(0.4, 0, 0.2, 1) | Sidebar expand, page transitions |

### 7.3 Specific Animations

#### Connection State Spinner
```
Icon: RefreshCw
Animation: Rotate 360deg, 1s, linear, infinite
Context: Connecting state, host reachability check
```

#### Session Live Indicator
```
Element: 8px circle
Animation: Scale 1.0 → 1.4 → 1.0, opacity 1 → 0.5 → 1
Duration: 2s, ease-in-out, infinite
Colors: Protocol color
```

#### Security Alert Pulse
```
Element: Border or badge
Animation: Box-shadow 0 0 0 0 → 0 0 0 8px (color/0) + opacity 1 → 0
Duration: 1.5s, ease-out, 3 pulses then rest
Trigger: Security event (unverified host, failed auth)
```

#### Terminal Cursor
```
Element: Block or bar
Animation: Opacity 1 ↔ 0
Duration: 530ms, steps(1), infinite
```

#### Modal Entrance
```
Overlay: opacity 0 → 1, 100ms
Panel: opacity 0 → 1 + scale 0.96 → 1.00, 150ms, ease-out
```

#### Toast Slide
```
Desktop: translateX(120%) → 0, 200ms, ease-out
Mobile: translateY(100%) → 0, 200ms, ease-out
Exit: Reverse, 150ms
```

#### Two-Step Security Button
```
Arming: Background color 150ms normal, icon shake 200ms
Confirmation window: Progress bar 3s linear countdown at bottom of button
Auto-cancel: Fade to default, 100ms
```

### 7.4 Reduced Motion
```css
@media (prefers-reduced-motion: reduce) {
  * {
    animation-duration: 0.01ms !important;
    transition-duration: 0.01ms !important;
  }
  /* Keep only essential: terminal cursor blink */
  .terminal-cursor {
    animation-duration: 530ms !important;
  }
}
```

---

## 8. Z-Index Scale

| Layer | Z-Index | Element |
|-------|---------|---------|
| Base | 0 | Page content |
| Sticky | 10 | Sticky headers |
| Dropdown | 20 | Select menus, tooltips |
| Overlay | 30 | Backdrops, drawers |
| Modal | 40 | Dialogs, MFA challenge |
| Toast | 50 | Notifications |
| Critical | 60 | Security alerts that demand attention |

---

## 9. Shadows

| Token | Value | Usage |
|-------|-------|-------|
| `shadow-sm` | `0 1px 2px rgba(0,0,0,0.3)` | Buttons, cards hover |
| `shadow-md` | `0 4px 6px -1px rgba(0,0,0,0.4)` | Dropdowns, popovers |
| `shadow-lg` | `0 10px 15px -3px rgba(0,0,0,0.5)` | Modals, toasts |
| `shadow-xl` | `0 20px 25px -5px rgba(0,0,0,0.6)` | Full-screen overlays |
| `shadow-focus` | `0 0 0 3px rgba(59,130,246,0.3)` | Focus rings |

---

## 10. Document Control

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1 | 2026-05-27 | Ame (Designer Subagent) | Initial design system |
