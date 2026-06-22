# vexa — Responsive Design Strategy

> **Classification:** Security-Critical Application  
> **Owner:** Dhar  > **Status:** Draft — Design Review  > **Last Updated:** 2026-05-27

---

## 1. Breakpoint Definitions

| Name | Width | Target Devices | Primary Context |
|------|-------|---------------|-----------------|
| **Mobile Small** | 320px – 374px | Small phones, compact devices | Portrait only |
| **Mobile** | 375px – 639px | Standard phones | Portrait + landscape |
| **Tablet** | 640px – 1023px | Tablets, large phones landscape | Portrait + landscape |
| **Desktop Small** | 1024px – 1279px | Small laptops, tablet landscape | Landscape |
| **Desktop** | 1280px – 1439px | Standard laptops, monitors | Landscape |
| **Desktop Large** | 1440px – 1919px | Large monitors, ultrawide | Landscape |
| **Desktop XL** | 1920px+ | 4K, ultrawide, multi-monitor | Landscape |

**Implementation:** Mobile-first approach with `min-width` media queries.

---

## 2. Layout Transformations by Breakpoint

### 2.1 Navigation: Sidebar → Top Bar → Bottom Bar

```
Desktop (≥1024px)          Tablet (640-1023px)         Mobile (<640px)
┌────┬────────────────┐   ┌────────────────────┐      ┌────────────────┐
│ 🏠 │  Content       │   │ [≡] Title    [👤]  │      │ Content Area   │
│ 🖥️ │                │   ├────────────────────┤      │                │
│ 📋 │                │   │                    │      ├────────────────┤
│ 📊 │                │   │                    │      │ [🏠] [🖥️] [📋] │
│ ⚙️ │                │   │                    │      │ [📊] [⚙️]      │
└────┴────────────────┘   └────────────────────┘      └────────────────┘
  Sidebar expanded           Sidebar collapsed + hamburger   Bottom tabs
  240px / 64px toggle        Top bar only                    Bottom bar 64px
```

**Behavior:**
- **Desktop:** Sidebar always visible (240px default, collapsible to 64px)
- **Tablet:** Sidebar collapses to icon-only (64px), or hides behind hamburger menu
- **Mobile:** Bottom tab bar (4-5 items), no sidebar

### 2.2 Dashboard: Grid → Stack

```
Desktop (≥1280px)          Tablet (640-1279px)         Mobile (<640px)
┌────────────────────────┐  ┌────────────────┐          ┌──────────────┐
│ ┌────┐┌────┐┌────┐┌────┐│  │ ┌────┐ ┌────┐  │          │ ┌────┐       │
│ │Stat││Stat││Stat││Stat││  │ │Stat│ │Stat│  │          │ │Stat│       │
│ └────┘└────┘└────┘└────┘│  │ └────┘ └────┘  │          │ └────┘       │
├────────────────────────┤  ├────────────────┤          ├──────────────┤
│ ┌──────────┐┌────────┐ │  │ ┌────────────┐  │          │ ┌──────────┐ │
│ │ Active   ││ Recent │ │  │ │  Active      │  │          │ │ Active   │ │
│ │ Sessions ││ Sessions│ │  │ │  Sessions    │  │          │ │ Sessions │ │
│ │          ││        │ │  │ └──────────────┘  │          │ │          │ │
│ │          ││        │ │  │ ┌────────────┐  │          │ └──────────┘ │
│ └──────────┘└────────┘ │  │ │  Hosts       │  │          │ ┌──────────┐ │
├────────────────────────┤  │ └──────────────┘  │          │ │ Recent   │ │
│ Host Quick Access Grid │  └────────────────┘          │ │ Sessions │ │
│ ┌────┐┌────┐┌────┐    │                              │ └──────────┘ │
│ │Host││Host││Host│ ... │                              └──────────────┘
│ └────┘└────┘└────┘    │
└────────────────────────┘
```

**Grid columns:**
- Desktop Large: 4 stat cards, 3 host cards per row
- Desktop: 4 stat cards, 3 host cards per row
- Desktop Small: 4 stat cards, 2 host cards per row
- Tablet (landscape): 2 stat cards per row, 2 host cards per row
- Tablet (portrait): 2 stat cards per row, 2 host cards per row
- Mobile: 1 column stack, horizontal scroll for stat cards

### 2.3 Host List: Table → Cards

```
Desktop (≥1024px)          Tablet (<1024px)            Mobile (<640px)
┌────────────────────────┐  ┌────────────────┐          ┌──────────────┐
│ Name    IP    Status   │  │ ┌────────────┐  │          │ ┌──────────┐ │
│ ────────────────────── │  │ │ 🟢 prod-web│  │          │ │ 🟢 prod  │ │
│ prod-web 10.0.1.4  🟢  │  │ │ 10.0.1.4   │  │          │ │ SSH      │ │
│ win-dc   10.0.2.8  🔵  │  │ │ [Connect]  │  │          │ │ 10.0.1.4 │ │
│ lab-kvm  10.0.3.1  🟡  │  │ └────────────┘  │          │ │ [Connect]│ │
│ ...                    │  │ ┌────────────┐  │          │ └──────────┘ │
│                        │  │ │ 🔵 win-dc  │  │          │ ┌──────────┐ │
│                        │  │ │ 10.0.2.8   │  │          │ │ 🔵 win   │ │
│                        │  │ │ [Connect]  │  │          │ │ ...      │ │
│                        │  │ └────────────┘  │          │ └──────────┘ │
│                        │  └────────────────┘          └──────────────┘
```

**Transformation rules:**
- **≥1024px:** Full table with all columns (Name, IP, Protocol, Status, Last Seen, Actions)
- **640-1023px:** Card layout, 2 columns
- **<640px:** Card layout, 1 column, actions in ⋮ menu
- **<375px:** Compact cards, hide descriptions, smaller protocol badges

### 2.4 Terminal: Full-Screen Priority

```
Desktop (≥1024px)          Tablet (640-1023px)         Mobile (<640px)
┌────────────────────────┐  ┌────────────────┐          ┌──────────────┐
│ [Tab1] [Tab2] [+]      │  │ [Tab1] [Tab2]│          │ [prod-w...]│
├────────────────────────┤  ├────────────────┤          ├──────────────┤
│                        │  │                │          │              │
│     Terminal Area      │  │  Terminal Area │          │ Terminal     │
│                        │  │                │          │ Area (100%)  │
│                        │  │                │          │              │
├────────────────────────┤  ├────────────────┤          ├──────────────┤
│ 🔒TLS · 80x24 · 🟢     │  │ 🔒TLS · 80x24  │          │ ⌨️ [Ctrl]... │
└────────────────────────┘  └────────────────┘          └──────────────┘
  Sidebar hidden              Sidebar hidden             No sidebar
  Status bar full             Status bar compact        Status bar minimal
```

**Terminal behavior:**
- **All sizes:** Terminal content area gets maximum available space
- **<1024px:** Sidebar always hidden; navigation via hamburger/bottom bar
- **<640px:** Status bar reduced to essential only (security badge, menu button)
- **<375px:** Virtual keyboard accessory bar shown for special keys

### 2.5 RDP/VNC Viewer: Canvas Scaling

| Breakpoint | Default Zoom | Toolbar | Controls |
|-----------|--------------|---------|----------|
| Desktop (≥1280px) | Fit to window | Floating bottom-center | All visible |
| Desktop Small (1024-1279px) | Fit to window | Floating bottom-center | Collapse zoom |
| Tablet (640-1023px) | 75% or Fit | Floating bottom | Essential only |
| Mobile (<640px) | Fit to window | Bottom fixed bar | Minimal (disconnect, fullscreen, keyboard) |

**Canvas scaling modes:**
1. **Fit to window:** Scales proportionally to fill available space (letterbox if aspect mismatch)
2. **100% / 1:1:** Native resolution, scrollbars if larger than viewport
3. **Custom %:** User-selected, persists per session

### 2.6 Settings: Split → Stack

```
Desktop (≥1024px)          Mobile (<1024px)
┌────────────────────────┐  ┌────────────────┐
│ ┌──────┬──────────────┐│  │ Settings       │
│ │General│  Settings    ││  ├────────────────┤
│ │──────│  Content     ││  │ General       →│
│ │Security│             ││  │ Security      →│
│ │──────│             ││  │ Sessions      →│
│ │Sessions│             ││  │ Terminal      →│
│ │──────│             ││  │ Network       →│
│ │ ...  │             ││  │ Audit         →│
│ └──────┴──────────────┘│  └────────────────┘
│   240px   Flexible      │   Tap opens sub-screen
```

**Settings behavior:**
- **≥1024px:** Two-pane layout, persistent category list
- **640-1023px:** Two-pane with narrower category column (160px)
- **<640px:** Category list full-screen, tap pushes detail screen with back button

### 2.7 Audit Logs: Table → Cards

| Breakpoint | Layout | Columns Visible |
|-----------|--------|-----------------|
| ≥1280px | Full table | Time, User, Event, Host, Result, Actions |
| 1024-1279px | Full table | Time, User, Event, Result (Host hidden, hover to see) |
| 640-1023px | Cards | All data in card format, 1-2 columns |
| <640px | Cards | 1 column, expandable for details |

---

## 3. Component Behavior by Breakpoint

### 3.1 Modals/Dialogs

| Breakpoint | Width | Position |
|-----------|-------|----------|
| ≥640px | 480px max, centered | Center screen with backdrop |
| <640px | 100% width, max-height 90vh | Bottom sheet, slides up |

**Bottom sheet behavior (mobile):**
- Slides up from bottom with `transform: translateY(100%) → 0`
- Drag handle at top (centered, 40px wide, 4px tall)
- Swipe down to dismiss (if cancellable)
- Keyboard pushes content up, never covers primary action

### 3.2 Command Palette / Quick Search

| Breakpoint | Position | Size |
|-----------|----------|------|
| ≥640px | Center screen | 560px wide, max 600px tall |
| <640px | Top of screen | 100% width, 70% height |

### 3.3 Toasts

| Breakpoint | Position | Animation |
|-----------|----------|-----------|
| ≥1024px | Top-right, 400px max | Slide in from right |
| 640-1023px | Top-center, 80% width | Slide in from top |
| <640px | Bottom-center, 100% - 32px padding | Slide in from bottom |

### 3.4 Context Menus

| Breakpoint | Trigger | Style |
|-----------|---------|-------|
| Desktop | Right-click or ⋮ button | Dropdown, max 240px |
| Mobile | Long-press or ⋮ button | Bottom action sheet |

---

## 4. Touch Targets & Accessibility

### 4.1 Minimum Touch Targets

| Element | Desktop (mouse) | Tablet | Mobile |
|---------|----------------|--------|--------|
| Buttons | 32×24px | 44×44px | 48×48px |
| Icon buttons | 32×32px | 44×44px | 48×48px |
| List items | 36px height | 48px height | 56px height |
| Input fields | 32px height | 44px height | 48px height |
| Tab bar items | 36px height | 48px height | 56px height |
| Checkboxes | 16×16px | 24×24px | 28×28px |
| Close/× buttons | 24×24px | 32×32px | 44×44px |

### 4.2 Font Size Scaling

| Breakpoint | Base Size | Mono Size |
|-----------|-----------|-----------|
| ≥1280px | 14px | 13px |
| 1024-1279px | 14px | 13px |
| 640-1023px | 14px | 13px |
| 375-639px | 14px | 12px |
| 320-374px | 13px | 12px |

**Note:** User preference overrides — respect browser/system font size settings.

### 4.3 Safe Areas

```css
/* Mobile safe area handling */
.env-mobile {
  padding-top: env(safe-area-inset-top);
  padding-bottom: env(safe-area-inset-bottom);
  padding-left: env(safe-area-inset-left);
  padding-right: env(safe-area-inset-right);
}

/* Bottom bar on iOS */
.bottom-nav {
  padding-bottom: max(16px, env(safe-area-inset-bottom));
  height: calc(64px + env(safe-area-inset-bottom));
}
```

### 4.4 WCAG 2.1 AA Compliance Checklist

| Requirement | Implementation |
|------------|----------------|
| **Color contrast** | All text ≥4.5:1 against background (tested with APCA) |
| **Large text contrast** | Headings ≥3:1 |
| **UI component contrast** | Borders/focus indicators ≥3:1 |
| **Focus indicators** | 2px solid `border-focus`, visible on all interactive elements |
| **Keyboard navigation** | Full tab order, no keyboard traps, visible focus states |
| **Screen reader labels** | All icons have `aria-label`, all images have `alt` |
| **Touch target size** | Minimum 44×44px on touch devices |
| **Motion sensitivity** | `prefers-reduced-motion` respected, terminal cursor blink kept |
| **Zoom support** | 200% zoom without horizontal scroll on desktop |
| **Form labels** | All inputs have visible labels or `aria-label` |
| **Error identification** | Inline errors linked to inputs with `aria-describedby` |
| **Status announcements** | Live regions for connection state changes |

---

## 5. Responsive Typography Scale

| Token | Desktop | Tablet | Mobile |
|-------|---------|--------|--------|
| `display` | 32px | 28px | 24px |
| `h1` | 24px | 22px | 20px |
| `h2` | 20px | 18px | 18px |
| `h3` | 16px | 16px | 16px |
| `h4` | 14px | 14px | 14px |
| `body` | 14px | 14px | 14px |
| `body-sm` | 13px | 13px | 13px |
| `caption` | 12px | 12px | 12px |
| `mono` | 13px | 13px | 12px |
| `mono-sm` | 11px | 11px | 11px |

---

## 6. Responsive Spacing

| Token | Desktop | Tablet | Mobile |
|-------|---------|--------|--------|
| `space-1` (4px) | 4px | 4px | 4px |
| `space-2` (8px) | 8px | 8px | 8px |
| `space-3` (12px) | 12px | 12px | 8px |
| `space-4` (16px) | 16px | 16px | 12px |
| `space-5` (20px) | 20px | 16px | 12px |
| `space-6` (24px) | 24px | 20px | 16px |
| `space-8` (32px) | 32px | 24px | 20px |
| `space-10` (40px) | 40px | 32px | 24px |
| `space-12` (48px) | 48px | 40px | 32px |

**Page margins:**
- Desktop: 24px
- Tablet: 20px
- Mobile: 16px
- Mobile Small: 12px

---

## 7. Orientation Handling

### 7.1 Portrait vs Landscape (Mobile)

| Screen | Portrait | Landscape |
|--------|----------|-----------|
| Dashboard | Stacked single column | 2-column stat cards, 2-column hosts |
| Host List | Stacked cards | 2-column cards |
| Terminal | Full screen, accessory bar | Full screen, hide accessory bar |
| RDP/VNC | Fit to height, scroll width | Fit to width, scroll height |
| Settings | Single column categories | 2-pane if ≥640px |

### 7.2 Terminal Landscape Mode

When mobile is in landscape:
- Hide status bar (swipe up to reveal)
- Hide bottom keyboard bar (swipe up to reveal)
- Terminal takes 100% of viewport
- Quick-action overlay appears on two-finger tap

---

## 8. Print Styles (Optional)

```css
@media print {
  .sidebar,
  .bottom-nav,
  .top-bar,
  .action-buttons,
  .security-badges {
    display: none !important;
  }
  
  .terminal-content,
  .audit-log-table {
    background: white !important;
    color: black !important;
    font-family: monospace !important;
  }
  
  body {
    padding: 0.5in;
  }
}
```

**Use case:** Printing terminal output or audit logs for compliance documentation.

---

## 9. Document Control

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1 | 2026-05-27 | Ame (Designer Subagent) | Initial responsive strategy |
