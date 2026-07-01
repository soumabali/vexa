# Vexa Redesign Dispatch-Ready Execution Plan

> **Status**: APPROVED — autonomous execution 
> **Date**: 2026-06-30 
> **Author**:Ame (Hermes)
> **Scope**: Frontend-only redesign`02-application/apps/web/`
> **Backend**: UNTOUCHED 
> **Reference**:`/home/ubuntu/vexa-redesign/`(Design.md 12 HTML mockups) 
> **Workflow**: PlanDispatch (Claude Code)Verify Commit/Push Document 
> **Execution Mode**: **Autonomous** — Ame runs all phases sequentially, phase-level checkpoints only

---

## Autonomous Execution Protocol

### Execution Model
Ame menjalankan semua dispatch dalam phase secara berurutan TANPA checkpoint per-dispatch.
**Phase-level checkpoint**: Ame hanya berhenti di akhir setiap phase untuk:
1. Verify build + test hijau
2. Commit semua perubahan phase
3. Update progress file
4. Lanjut ke phase berikutnya (atau stop jika error)

### Checkpoint File: `.hermes/vexa-redesign-progress.json`
**Location**: `02-application/.hermes/vexa-redesign-progress.json`
**Purpose**: Single source of truth untuk resume. Dibaca sebelum setiap dispatch.

```json
{
  "plan": "2026-06-30_ui-redesign",
  "branch": "feat/ui-redesign",
  "started_at": "2026-06-30T...",
  "updated_at": "2026-06-30T...",
  "current_phase": 1,
  "current_dispatch": 1,
  "status": "in_progress",
  "phases": {
    "0": { "status": "completed", "completed_at": "..." },
    "1": { "status": "in_progress", "dispatches": { "1": "completed", "2": "pending" } },
    "2": { "status": "pending" },
    "3": { "status": "pending", "subphases": { "3a": "pending", "3b": "pending" } },
    "4": { "status": "pending", "dispatches": { "4": "pending", "5": "pending", "6": "pending", "7": "pending" } },
    "5": { "status": "pending", "dispatches": { "8": "pending", "9": "pending", "10": "pending" } },
    "6": { "status": "pending" }
  },
  "last_error": null,
  "git_commits": []
}
```

### Pre-Dispatch Verification Gate (RUN BEFORE EVERY DISPATCH)
Sebelum setiap dispatch, Ame jalankan:

```bash
# 1. Read progress file
cat 02-application/.hermes/vexa-redesign-progress.json

# 2. Verify branch
cd /home/ubuntu/projects/vexa
git branch --show-current
# Expected: feat/ui-redesign

# 3. Verify clean working tree (no leftover from previous dispatch)
git status --porcelain 02-application/apps/web/
# Expected: empty OR only .hermes/ files

# 4. If previous dispatch touched files, verify they were committed
git log --oneline -3
# Expected: last commit matches previous dispatch work

# 5. Quick build check (if Phase >= 2)
cd 02-application/apps/web && npm run build 2>&1 | tail -5
# If build fails: DO NOT proceed. Fix first.
```

### Resume Protocol (when session crashes / context resets)
Jika Ame kehilangan context (new session, crash, `/new`):

1. **Read progress file**:
   ```bash
   cat /home/ubuntu/projects/vexa/02-application/.hermes/vexa-redesign-progress.json
   ```

2. **Determine resume point**:
   - `status: "completed"` → nothing to do
   - `status: "in_progress"` → resume from `current_dispatch`
   - `last_error: non-null` → see Error Recovery

3. **Verify git state**:
   ```bash
   cd /home/ubuntu/projects/vexa
   git branch --show-current  # must be feat/ui-redesign
   git log --oneline -5       # verify last commit
   git status --porcelain      # check for uncommitted work
   ```

4. **If uncommitted work exists**:
   - Run build: `cd 02-application/apps/web && npm run build`
   - If build passes: commit the work, update progress, resume
   - If build fails: stash the work, note in `last_error`, resume from current_dispatch

5. **Resume dispatch**:
   - Read plan section for `current_dispatch`
   - Dispatch Claude Code with the task
   - On success: update progress file, proceed to next dispatch

### Error Recovery Decision Tree

```
Dispatch fails (build error / test failure / Claude Code error > 30)
│
├── Error is TypeScript/lint error in files from THIS dispatch?
│   ├── YES: Re-dispatch with error context ("fix these errors: ...")
│   │   └── Still fails after 2 retries: STOP, report to Dhar
│   └── NO: Error is in files from PREVIOUS dispatch?
│       ├── YES: Fix inline (root-cause, not eslint-disable)
│       │   └── Fix breaks other things: STOP, report to Dhar
│       └── NO: Error is environment/dependency issue?
│           └── Fix environment, re-dispatch. If unfixable: STOP
│
Dispatch succeeds but test failures appear
│
├── Tests broken by THIS dispatch's changes?
│   ├── YES: Fix tests in same dispatch (re-dispatch with test fix task)
│   └── NO: Tests were already broken (pre-existing)?
│       └── Note in progress file, proceed (fix in Phase 6)
│
Build succeeds but visual doesn't match mockup
│
├── Minor difference (spacing, color shade)?
│   └── Note for Phase 6 polish, proceed
└── Major difference (wrong layout, missing component)?
    └── Re-dispatch with specific correction instruction
```

### Context Management Strategy

**Problem**: 11 dispatches × (plan + verify + commit) = massive context
**Solution**: Delegate dispatches to Claude Code subagents

**Per-dispatch flow**:
1. Ame reads plan section for dispatch N (from progress file)
2. Ame creates Claude Code dispatch with:
   - Exact task description (from plan)
   - File list to modify
   - Acceptance criteria
   - Pitfalls list
3. Claude Code executes and returns result
4. Ame verifies: `npm run build` + `npm run test`
5. Ame commits with conventional message
6. Ame updates progress file
7. Ame proceeds to dispatch N+1

**Key**: Ame does NOT read Claude Code's intermediate output. Only:
- Dispatch prompt (from plan)
- Final result summary
- Build/test output
This keeps Ame's context lean.

### Phase-Level Checkpoint Review

Setelah semua dispatch dalam phase selesai, Ame:

1. **Full verification**:
   ```bash
   cd 02-application/apps/web
   npm run build 2>&1 | tail -20
   npm run test 2>&1 | tail -30
   ```

2. **Visual spot-check** (if dev server available):
   - Start: `npm run dev &`
   - Check 1-2 key pages against mockup screenshots
   - Stop: `kill %1`

3. **Commit all phase work**:
   ```bash
   cd /home/ubuntu/projects/vexa
   git add 02-application/apps/web/
   git commit -m "feat(ui): Phase N complete — [description]"
   ```

4. **Update progress file**:
   - Mark phase as `completed`
   - Set `current_phase` to N+1
   - Clear `current_dispatch` 

5. **Continue to next phase** (autonomous — no Dhar checkpoint needed)

### When Ame STOPS (hard stops only)
Ame berhenti dan report ke Dhar HANYA jika:
- ❌ Build fails after 2 retry attempts on same dispatch
- ❌ Test failures that can't be fixed in 1 retry
- ❌ Claude Code dispatch returns error > 30 errors
- ❌ Environment issue (npm install fails, disk full, etc.)
- ❌ Dhar sends `/stop` or explicit stop message

**Ame does NOT stop for**:
- ✅ Minor visual differences (defer to Phase 6)
- ✅ Pre-existing test failures (note and proceed)
- ✅ Warnings (lint warnings OK, errors must be fixed)
- ✅ Phase boundary reached (continue to next phase)

---

## Compliance

Plan ini mengikuti aturan:
- `AGENTS.md` — workflow mandatory, tidak commit dari dalam `02-application/`
- `02-application/CLAUDE.md` — Claude Code hanya edit di dalam `02-application/`
- `00-meta/pre-push-checklist.md` — verification gates sebelum push
- `00-meta/skills.md` — superpowers, caveman, graphify, playwright wajib aktif
- MAX 30 errors per dispatch
- Per-task commit dengan conventional prefix
- Restore generated artifacts (`.next/`, `tsconfig.tsbuildinfo`) — JANGAN commit
- Pakai `useAsyncData` untuk fetching, bukan raw useEffect
- Pakai komponen `Dialog` dari `@/components/ui/dialog` untuk modal
- Pakai `toast` dari `sonner` untuk notifikasi

---

## Architecture Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Color format | **Hex langsung di Tailwind config** | Mockup pakai hex, lebih simple, shadcn tetap jalan dengan alias |
| Icon library | **Full swap ke Material Symbols** | Mockup semua pakai Material Symbols, harus match persis |
| shadcn compatibility | **Map shadcn tokens sebagai alias** | `--background` = `--surface-dim`, `--primary` = value baru, dll |
| HSL → hex | **Keep CSS vars di globals.css sebagai documentation** | Tailwind config pakai hex, CSS vars jadi reference |
| Branch | **`feat/ui-redesign`** dari `main` | Isolasi perubahan besar |

---

## Phase 0 — Pre-Flight (Ame, no dispatch)

### 0.1 Create branch
```bash
cd /home/ubuntu/projects/vexa
git checkout main
git pull origin main
git checkout -b feat/ui-redesign
```

### 0.2 Backup current globals.css and tailwind.config.ts
```bash
cp 02-application/apps/web/src/app/globals.css 02-application/apps/web/src/app/globals.css.bak
cp 02-application/apps/web/tailwind.config.ts 02-application/apps/web/tailwind.config.ts.bak
```

### 0.3 Copy mockup reference files ke working directory
```bash
cp -r /home/ubuntu/vexa-redesign 02-application/.hermes/vexa-redesign-ref
```

---

## Phase 1 — Design Token Foundation

> **Dispatch 1** | **Files**: 4 | **Est. errors**: <10

### Objective
Replace current CSS variables and Tailwind config with the full Material 3 design token set from Design.md. Add Material Symbols font. Create icon wrapper component.

### Tasks

#### Task 1.1 — Rewrite `globals.css`
**File**: `apps/web/src/app/globals.css`

Replace ENTIRE file with:

```css
@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
  :root {
    /* Material 3 Design Tokens — Vexa Core (from Design.md) */
    --surface-dim: #10131a;
    --surface: #10131a;
    --surface-bright: #363941;
    --surface-container-lowest: #0b0e15;
    --surface-container-low: #191b23;
    --surface-container: #1d2027;
    --surface-container-high: #272a31;
    --surface-container-highest: #32353c;
    --surface-variant: #32353c;
    --on-surface: #e1e2ec;
    --on-surface-variant: #c2c6d6;
    --inverse-surface: #e1e2ec;
    --inverse-on-surface: #2e3038;
    --outline: #8c909f;
    --outline-variant: #424754;
    --surface-tint: #adc6ff;
    
    /* Primary */
    --primary: #adc6ff;
    --on-primary: #002e6a;
    --primary-container: #4d8eff;
    --on-primary-container: #00285d;
    --inverse-primary: #005ac2;
    --primary-fixed: #d8e2ff;
    --primary-fixed-dim: #adc6ff;
    --on-primary-fixed: #001a42;
    --on-primary-fixed-variant: #004395;
    
    /* Secondary */
    --secondary: #b9c8de;
    --on-secondary: #233143;
    --secondary-container: #39485a;
    --on-secondary-container: #a7b6cc;
    --secondary-fixed: #d4e4fa;
    --secondary-fixed-dim: #b9c8de;
    --on-secondary-fixed: #0d1c2d;
    --on-secondary-fixed-variant: #39485a;
    
    /* Tertiary */
    --tertiary: #ffb786;
    --on-tertiary: #502400;
    --tertiary-container: #df7412;
    --on-tertiary-container: #461f00;
    --tertiary-fixed: #ffdcc6;
    --tertiary-fixed-dim: #ffb786;
    --on-tertiary-fixed: #311400;
    --on-tertiary-fixed-variant: #723600;
    
    /* Error */
    --error: #ffb4ab;
    --on-error: #690005;
    --error-container: #93000a;
    --on-error-container: #ffdad6;
    
    /* Background */
    --background: #10131a;
    --on-background: #e1e2ec;
    
    /* shadcn/ui aliases (map to Material 3 tokens) */
    --foreground: var(--on-surface);
    --card: var(--surface-container);
    --card-foreground: var(--on-surface);
    --popover: var(--surface-container-high);
    --popover-foreground: var(--on-surface);
    --primary-foreground: var(--on-primary);
    --secondary-foreground: var(--on-secondary);
    --muted: var(--surface-container-low);
    --muted-foreground: var(--on-surface-variant);
    --accent: var(--primary-container);
    --accent-foreground: var(--on-primary-container);
    --destructive: var(--error);
    --destructive-foreground: var(--on-error);
    --border: var(--outline-variant);
    --input: var(--outline-variant);
    --ring: var(--primary);
    --radius: 0.5rem;
    
    /* Status colors (used in mockups) */
    --success: #4ade80;
    --warning: #facc15;
    --info: #adc6ff;
  }

  * {
    @apply border-border;
  }

  body {
    @apply bg-background text-on-surface antialiased;
    font-family: 'Inter', system-ui, -apple-system, sans-serif;
  }

  code, pre, .font-mono, .mono-label {
    font-family: 'JetBrains Mono', 'Fira Code', monospace;
  }

  /* Material Symbols base */
  .material-symbols-outlined {
    font-variation-settings: 'FILL' 0, 'wght' 400, 'GRAD' 0, 'opsz' 24;
  }
}

@layer components {
  /* Glass card (from mockup) */
  .glass-card {
    background: rgba(30, 41, 59, 0.4);
    backdrop-filter: blur(8px);
    border: 1px solid rgba(51, 65, 85, 0.5);
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }
  .glass-card:hover {
    background: rgba(51, 65, 85, 0.6);
    border-color: #adc6ff;
    transform: translateY(-2px);
  }
  
  /* No scrollbar */
  .no-scrollbar::-webkit-scrollbar {
    display: none;
  }
  .no-scrollbar {
    -ms-overflow-style: none;
    scrollbar-width: none;
  }
  
  /* Terminal dot */
  .terminal-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
  }
}
```

**Acceptance**: File compiles, no CSS errors.

#### Task 1.2 — Rewrite `tailwind.config.ts`
**File**: `apps/web/tailwind.config.ts`

Replace with:

```typescript
import type { Config } from "tailwindcss";

const config: Config = {
  darkMode: "class",
  content: [
    "./src/pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/components/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        // shadcn aliases (reference CSS vars for compatibility)
        background: "#10131a",
        foreground: "#e1e2ec",
        card: "#1d2027",
        "card-foreground": "#e1e2ec",
        popover: "#272a31",
        "popover-foreground": "#e1e2ec",
        primary: "#adc6ff",
        "primary-foreground": "#002e6a",
        secondary: "#b9c8de",
        "secondary-foreground": "#233143",
        muted: "#191b23",
        "muted-foreground": "#c2c6d6",
        accent: "#4d8eff",
        "accent-foreground": "#00285d",
        destructive: "#ffb4ab",
        "destructive-foreground": "#690005",
        border: "#424754",
        input: "#424754",
        ring: "#adc6ff",
        
        // Material 3 surface hierarchy
        surface: {
          DEFAULT: "#10131a",
          dim: "#10131a",
          bright: "#363941",
          "container-lowest": "#0b0e15",
          "container-low": "#191b23",
          container: "#1d2027",
          "container-high": "#272a31",
          "container-highest": "#32353c",
          variant: "#32353c",
        },
        "on-surface": "#e1e2ec",
        "on-surface-variant": "#c2c6d6",
        "inverse-surface": "#e1e2ec",
        "inverse-on-surface": "#2e3038",
        outline: {
          DEFAULT: "#8c909f",
          variant: "#424754",
        },
        "surface-tint": "#adc6ff",
        
        // Primary
        "primary-container": "#4d8eff",
        "on-primary": "#002e6a",
        "on-primary-container": "#00285d",
        "inverse-primary": "#005ac2",
        "primary-fixed": "#d8e2ff",
        "primary-fixed-dim": "#adc6ff",
        "on-primary-fixed": "#001a42",
        "on-primary-fixed-variant": "#004395",
        
        // Secondary
        "secondary-container": "#39485a",
        "on-secondary": "#233143",
        "on-secondary-container": "#a7b6cc",
        "secondary-fixed": "#d4e4fa",
        "secondary-fixed-dim": "#b9c8de",
        "on-secondary-fixed": "#0d1c2d",
        "on-secondary-fixed-variant": "#39485a",
        
        // Tertiary
        tertiary: "#ffb786",
        "on-tertiary": "#502400",
        "tertiary-container": "#df7412",
        "on-tertiary-container": "#461f00",
        "tertiary-fixed": "#ffdcc6",
        "tertiary-fixed-dim": "#ffb786",
        "on-tertiary-fixed": "#311400",
        "on-tertiary-fixed-variant": "#723600",
        
        // Error
        error: "#ffb4ab",
        "on-error": "#690005",
        "error-container": "#93000a",
        "on-error-container": "#ffdad6",
        
        // Status
        success: "#4ade80",
        warning: "#facc15",
        info: "#adc6ff",
      },
      fontFamily: {
        sans: ["Inter", "system-ui", "sans-serif"],
        mono: ["JetBrains Mono", "Fira Code", "monospace"],
        "headline-lg": ["Inter"],
        "headline-md": ["Inter"],
        "headline-sm": ["Inter"],
        "headline-lg-mobile": ["Inter"],
        "body-lg": ["Inter"],
        "body-md": ["Inter"],
        "label-md": ["Inter"],
        "mono-code": ["JetBrains Mono"],
      },
      fontSize: {
        "headline-lg": ["30px", { lineHeight: "38px", letterSpacing: "-0.02em", fontWeight: "700" }],
        "headline-md": ["24px", { lineHeight: "32px", letterSpacing: "-0.01em", fontWeight: "600" }],
        "headline-sm": ["20px", { lineHeight: "28px", fontWeight: "600" }],
        "headline-lg-mobile": ["24px", { lineHeight: "32px", fontWeight: "700" }],
        "body-lg": ["16px", { lineHeight: "24px", fontWeight: "400" }],
        "body-md": ["14px", { lineHeight: "20px", fontWeight: "400" }],
        "label-md": ["12px", { lineHeight: "16px", letterSpacing: "0.02em", fontWeight: "500" }],
        "mono-code": ["13px", { lineHeight: "20px", fontWeight: "400" }],
      },
      spacing: {
        base: "4px",
        xs: "4px",
        sm: "8px",
        md: "16px",
        lg: "24px",
        xl: "32px",
        gutter: "20px",
        margin: "24px",
      },
      borderRadius: {
        DEFAULT: "0.25rem",
        sm: "0.25rem",
        md: "0.75rem",
        lg: "0.5rem",
        xl: "0.75rem",
        "2xl": "1rem",
        "3xl": "1.5rem",
        full: "9999px",
      },
    },
  },
  plugins: [],
};

export default config;
```

**Acceptance**: `npx tailwindcss --content './src/**/*.tsx' --output /dev/null` no errors.

#### Task 1.3 — Add Material Symbols font to `layout.tsx`
**File**: `apps/web/src/app/layout.tsx`

Add to `<head>` (via Next.js metadata or direct link):

```tsx
// Add to existing imports:
import MaterialSymbols from "@/components/ui/material-icon";

// In the <html> tag, add the font link:
// Either via next/font or CDN link in <head>

// Update the <html> tag:
<html
  lang="en"
  className={`${inter.variable} ${jetbrainsMono.variable} h-full antialiased dark`}
>
  <head>
    <link
      href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:wght,FILL@100..700,0..1&display=swap"
      rel="stylesheet"
    />
  </head>
  <body className="min-h-full flex flex-col bg-background text-on-surface font-sans">
    <QueryProvider>
      {children}
    </QueryProvider>
  </body>
</html>
```

**Acceptance**: Material Symbols font loads (check Network tab).

#### Task 1.4 — Create `MaterialIcon` wrapper component
**File**: `apps/web/src/components/ui/material-icon.tsx` (NEW)

```tsx
import { cn } from "@/lib/utils";

interface MaterialIconProps {
  name: string;
  className?: string;
  fill?: boolean;
  weight?: 100 | 200 | 300 | 400 | 500 | 600 | 700;
  size?: "sm" | "md" | "lg" | "xl";
}

const sizeMap = {
  sm: "text-[14px]",
  md: "text-[18px]",
  lg: "text-[20px]",
  xl: "text-[24px]",
};

export function MaterialIcon({
  name,
  className,
  fill = false,
  weight = 400,
  size = "md",
}: MaterialIconProps) {
  return (
    <span
      className={cn("material-symbols-outlined", sizeMap[size], className)}
      style={{
        fontVariationSettings: `'FILL' ${fill ? 1 : 0}, 'wght' ${weight}, 'GRAD' 0, 'opsz' 24`,
      }}
      aria-hidden="true"
    >
      {name}
    </span>
  );
}
```

**Acceptance**: Component renders, TypeScript compiles.

### Phase 1 Verification
```bash
cd 02-application/apps/web
npm run build 2>&1 | tail -20
# Expected: Build succeeds. May have warnings about unused vars — OK.
# Restore .next/ after build:
rm -rf .next
```

### Phase 1 Commit
```bash
cd /home/ubuntu/projects/vexa
git add 02-application/apps/web/src/app/globals.css \
       02-application/apps/web/tailwind.config.ts \
       02-application/apps/web/src/app/layout.tsx \
       02-application/apps/web/src/components/ui/material-icon.tsx
git commit -m "feat(ui): replace design tokens with Material 3 system + add Material Symbols"
```

---

## Phase 2 — Layout Shell

> **Dispatch 2** | **Files**: 5 (3 new, 2 modified) | **Est. errors**: <15

### Objective
Replace top-nav-only layout with sidebar + top nav shell matching mockup pattern.

### Tasks

#### Task 2.1 — Create `Sidebar.tsx`
**File**: `apps/web/src/components/layouts/Sidebar.tsx` (NEW)

Structure (from mockup):
```
<aside> (w-260px, fixed, bg-surface-container, border-r border-outline-variant)
  ├── UserContext (icon + name + cluster label)
  ├── NewConnection button (bg-primary-container text-on-primary-container)
  ├── NavSection (flex-1, scrollable)
  │   ├── All Hosts (dns icon, active: text-primary bg-secondary-container border-l-2 border-primary)
  │   ├── Groups (folder icon)
  │   ├── Cloud Instances (cloud icon)
  │   ├── Recent Sessions (history icon)
  │   └── Key Management (vpn_key icon)
  └── Footer (border-t pt-4)
      ├── Help Center (help_outline icon)
      └── Log out (logout icon, text-error)
```

Props:
```tsx
interface SidebarProps {
  activeItem?: "hosts" | "terminal" | "tunnels" | "vault" | "settings" | "sessions";
}
```

**Active state pattern**: `text-primary bg-secondary-container border-l-2 border-primary rounded-lg`
**Inactive state pattern**: `text-on-surface-variant hover:bg-surface-variant rounded-lg`

Each nav item: `flex items-center gap-3 px-3 py-2 font-label-md text-label-md`

**Acceptance**: Sidebar renders with all nav items, active state highlighted.

#### Task 2.2 — Redesign `TopNav.tsx`
**File**: `apps/web/src/components/layouts/TopNav.tsx`

Structure (from mockup):
```
<header> (h-16, fixed top-0, bg-background, border-b border-outline-variant, z-50)
  ├── Left: Logo (w-8 h-8 bg-primary rounded-lg + "terminal" icon) + "Vexa" text-headline-sm font-bold
  ├── Center: Nav links (Hosts/Terminal/Tunnels/Vault/Settings)
  │   active: text-primary border-b-2 border-primary pb-1
  │   inactive: text-on-surface-variant hover:text-on-surface
  └── Right: 
      ├── notifications button (w-10 h-10 rounded-lg hover:bg-surface-container-highest)
      ├── help button (same)
      └── avatar (w-8 h-8 rounded-full bg-secondary-container border border-outline-variant)
```

Replace all `lucide-react` imports with `MaterialIcon`.

**Acceptance**: TopNav renders, nav links highlight active route.

#### Task 2.3 — Restructure `DashboardLayout.tsx`
**File**: `apps/web/src/components/layouts/DashboardLayout.tsx`

```tsx
import TopNav from "./TopNav";
import Sidebar from "./Sidebar";
import { usePathname } from "next/navigation";

interface DashboardLayoutProps {
  children: React.ReactNode;
}

export function DashboardLayout({ children }: DashboardLayoutProps) {
  const pathname = usePathname();
  
  // Determine active sidebar item from pathname
  const activeItem = pathname?.startsWith("/hosts") ? "hosts" :
    pathname?.startsWith("/terminal") ? "terminal" :
    pathname?.startsWith("/tunnels") ? "tunnels" :
    pathname?.startsWith("/vault") ? "vault" :
    pathname?.startsWith("/settings") ? "settings" :
    pathname?.startsWith("/sessions") ? "sessions" : "hosts";

  return (
    <div className="min-h-screen bg-background">
      <TopNav />
      <div className="flex pt-16 h-screen">
        <Sidebar activeItem={activeItem} />
        <main className="flex-1 md:ml-[260px] p-8 overflow-y-auto bg-background h-full">
          {children}
        </main>
      </div>
    </div>
  );
}
```

**Acceptance**: All pages using `<DashboardLayout>` still render with new layout.

#### Task 2.4 — Create `AuthLayout.tsx`
**File**: `apps/web/src/components/auth/AuthLayout.tsx` (MODIFY existing)

Structure (Pattern F — split screen):
```
<div class="min-h-screen flex">
  Left panel (hidden md:flex, w-1/2, bg-surface-container, flex-col justify-center p-12):
    - Logo + "Vexa" headline-lg
    - Tagline
    - Feature highlights (3 items with icons)
    - Footer: version
  Right panel (flex-1, bg-background, flex items-center justify-center p-8):
    - Card (w-full max-w-md, bg-surface-container, border border-outline-variant, rounded-xl, p-8)
    - {children}
</div>
```

**Acceptance**: Auth pages render with split layout.

#### Task 2.5 — Add mobile FAB for sidebar-less pages
**File**: `apps/web/src/components/layouts/DashboardLayout.tsx`

For mobile (< md): sidebar hidden, show floating action button (FAB):
```tsx
<button className="md:hidden fixed bottom-6 right-6 w-14 h-14 bg-primary text-on-primary rounded-full shadow-xl z-[100]">
  <MaterialIcon name="add" size="lg" />
</button>
```

### Phase 2 Verification
```bash
cd 02-application/apps/web
npm run build 2>&1 | tail -20
# Expected: Build succeeds. Pages render with new layout.
rm -rf .next
```

### Phase 2 Commit
```bash
git add 02-application/apps/web/src/components/layouts/Sidebar.tsx \
       02-application/apps/web/src/components/layouts/TopNav.tsx \
       02-application/apps/web/src/components/layouts/DashboardLayout.tsx \
       02-application/apps/web/src/components/auth/AuthLayout.tsx
git commit -m "feat(ui): add sidebar + redesign topnav for layout shell"
```

---

## Phase 3 — Component Library Restyle

> **Dispatch 3** | **Files**: 28 | **Est. errors**: 20-30  
> **Split**: 2 dispatches (3a: core components, 3b: form + overlay components)

### Dispatch 3a — Core Components (14 files)

#### Task 3.1 — Card
**File**: `apps/web/src/components/ui/card.tsx`
Changes:
- `bg-card` → `bg-surface-container`
- `border-border` → `border-outline-variant`
- Add `rounded-xl` (from `rounded-lg`)
- Text: `text-card-foreground` → `text-on-surface`

#### Task 3.2 — Button
**File**: `apps/web/src/components/ui/button.tsx`
Variants:
- `default`: `bg-primary text-on-primary hover:bg-primary/90`
- `secondary`: `bg-primary-container text-on-primary-container hover:opacity-90`
- `outline`: `border border-outline-variant text-on-surface-variant hover:bg-surface-variant`
- `ghost`: `hover:bg-surface-variant text-on-surface-variant`
- `destructive`: `bg-error text-on-error hover:bg-error/90`
- `link`: `text-primary underline-offset-4 hover:underline`
Sizes: Keep existing (default, sm, lg, icon)

#### Task 3.3 — Badge
**File**: `apps/web/src/components/ui/badge.tsx`
Variants:
- `default`: `bg-primary/10 text-primary border-primary/20`
- `secondary`: `bg-surface-variant text-on-surface-variant`
- `destructive`: `bg-error/10 text-error border-error/20`
- `outline`: `border-outline-variant text-on-surface-variant`
- `success`: `bg-success/10 text-success`
- `warning`: `bg-warning/10 text-warning`

#### Task 3.4 — Input
**File**: `apps/web/src/components/ui/input.tsx`
- `bg-transparent` → `bg-surface-container-lowest`
- `border-input` → `border-outline-variant`
- `focus:ring-ring` → `focus:border-primary focus:ring-1 focus:ring-primary`
- `placeholder:text-muted-foreground` → `placeholder:text-on-surface-variant/50`

#### Task 3.5 — Label
**File**: `apps/web/src/components/ui/label.tsx`
- `text-foreground` → `text-on-surface-variant`
- Keep `font-medium`

#### Task 3.6 — Textarea
**File**: `apps/web/src/components/ui/textarea.tsx`
- Same changes as Input

#### Task 3.7 — Separator
**File**: `apps/web/src/components/ui/separator.tsx`
- `bg-border` → `bg-outline-variant`

#### Task 3.8 — Skeleton
**File**: `apps/web/src/components/ui/skeleton.tsx`
- `bg-muted` → `bg-surface-container-low`

#### Task 3.9 — Avatar
**File**: `apps/web/src/components/ui/avatar.tsx`
- Add `ring-2 ring-outline-variant` to default
- `bg-muted` → `bg-secondary-container`

#### Task 3.10 — Progress
**File**: `apps/web/src/components/ui/progress.tsx`
- Track: `bg-surface-container`
- Fill: `bg-primary`

#### Task 3.11 — Alert
**File**: `apps/web/src/components/ui/alert.tsx`
Variants:
- `default`: `bg-surface-container text-on-surface border-outline-variant`
- `destructive`: `bg-error-container/5 text-error border-error-container/20`
- `success`: `bg-success/5 text-success border-success/20`
- `warning`: `bg-warning/5 text-warning border-warning/20`

#### Task 3.12 — SecurityBadge
**File**: `apps/web/src/components/ui/SecurityBadge.tsx`
- Update colors to new tokens

#### Task 3.13 — Switch
**File**: `apps/web/src/components/ui/switch.tsx`
- Checked: `bg-primary`
- Unchecked: `bg-surface-variant`
- Thumb: `bg-on-surface` (light) when checked, `bg-on-surface-variant` when unchecked

#### Task 3.14 — Tabs
**File**: `apps/web/src/components/ui/tabs.tsx`
- Active: `border-b-2 border-primary text-on-surface`
- Inactive: `border-b-2 border-transparent text-on-surface-variant hover:text-on-surface`
- Tab content: `bg-surface-container` or transparent

### Dispatch 3b — Form + Overlay Components (14 files)

#### Task 3.15 — Dialog
**File**: `apps/web/src/components/ui/dialog.tsx`
- Overlay: `bg-background/80 backdrop-blur-sm`
- Content: `bg-surface-container-high border border-outline-variant rounded-xl`
- Header: `text-on-surface`
- Close: `text-on-surface-variant hover:text-on-surface`

#### Task 3.16 — DropdownMenu
**File**: `apps/web/src/components/ui/dropdown-menu.tsx`
- Content: `bg-surface-container-high border border-outline-variant`
- Item: `hover:bg-surface-variant text-on-surface`
- Separator: `bg-outline-variant`

#### Task 3.17 — ContextMenu
**File**: `apps/web/src/components/ui/context-menu.tsx`
- Same as DropdownMenu

#### Task 3.18 — Select
**File**: `apps/web/src/components/ui/select.tsx`
- Trigger: `bg-surface-container-highest border border-outline-variant`
- Content: `bg-surface-container-high border border-outline-variant`
- Item: `hover:bg-surface-variant`

#### Task 3.19 — Checkbox
**File**: `apps/web/src/components/ui/checkbox.tsx`
- Checked: `bg-primary text-on-primary border-primary`
- Unchecked: `border-outline-variant`
- Indicator: white checkmark

#### Task 3.20 — RadioGroup
**File**: `apps/web/src/components/ui/radio-group.tsx`
- Selected: `border-primary border-2`
- Unselected: `border-outline-variant`
- Indicator: `bg-primary`

#### Task 3.21 — Slider
**File**: `apps/web/src/components/ui/slider.tsx`
- Track: `bg-surface-container`
- Range: `bg-primary`
- Thumb: `bg-primary border-2 border-on-primary`

#### Task 3.22 — ScrollArea
**File**: `apps/web/src/components/ui/scroll-area.tsx`
- Bar: `bg-surface-container-high`

#### Task 3.23 — Tooltip
**File**: `apps/web/src/components/ui/tooltip.tsx`
- Content: `bg-surface-container-highest text-on-surface border border-outline-variant`

#### Task 3.24 — Table
**File**: `apps/web/src/components/ui/table.tsx`
- Header: `bg-surface-container-high text-on-surface-variant`
- Row: `border-b border-outline-variant`
- Row hover: `hover:bg-surface-variant/30`

#### Task 3.25 — Sonner/Toast
**File**: `apps/web/src/components/ui/sonner.tsx` + `toast.tsx`
- Toast: `bg-surface-container-highest border border-outline-variant text-on-surface`

#### Task 3.26 — Toaster
**File**: `apps/web/src/components/ui/toaster.tsx`
- Same styling as Sonner

#### Task 3.27 — use-toast hook
**File**: `apps/web/src/components/ui/use-toast.ts`
- No visual changes, just ensure compatibility

#### Task 3.28 — All remaining UI files
Check and update any files not covered above:
- `hover:bg-popover` → `hover:bg-surface-container-high`
- `text-muted-foreground` → `text-on-surface-variant`
- `bg-muted` → `bg-surface-container-low`
- `border-border` → `border-outline-variant`
- `bg-popover` → `bg-surface-container-high`
- `bg-destructive` → `bg-error`
- `text-destructive` → `text-error`

### Phase 3 Verification
```bash
cd 02-application/apps/web
npm run build 2>&1 | tail -20
# Expected: Build succeeds. All components use new tokens.
npm run test 2>&1 | tail -20
# Expected: Tests pass (may need test updates for new class names)
rm -rf .next
```

### Phase 3 Commit (per dispatch)
```bash
# Dispatch 3a
git add 02-application/apps/web/src/components/ui/
git commit -m "feat(ui): restyle core components with Material 3 tokens"

# Dispatch 3b
git add 02-application/apps/web/src/components/ui/
git commit -m "feat(ui): restyle form + overlay components with Material 3 tokens"
```

---

## Phase 4 — Mockup Pages (9 pages + 1 new)

> **Dispatches 4-7** | **Files**: 10 pages + ~30 child components | **Est. errors**: 15-25 per dispatch

### Dispatch 4 — Hosts Dashboard + Terminal

#### Task 4.1 — Hosts Dashboard
**File**: `apps/web/src/app/hosts/page.tsx` + child components

Layout from mockup `hosts_dashboard.html`:
```
Page header:
  h1.text-headline-lg "Hosts"
  p.text-body-md.text-on-surface-variant "Securely manage remote environments across SSH, RDP, VNC."
  Actions: grid_view toggle + "Add Host" button (bg-primary text-on-primary)

Filter bar:
  Search input (bg-surface-container-low border-outline-variant, search icon absolute left-3)
  Filter pills: All (bg-secondary-container text-on-secondary-container) | SSH | RDP | VNC
  Filters button (border-outline-variant)

Card grid: grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6
  Each card: glass-card rounded-xl p-5 flex flex-col gap-4
    Icon (w-12 h-12, bg-primary/10 or bg-tertiary-container/10)
    Status badge (Online: bg-green-500/10 text-green-400, Offline: bg-red-500/10 text-red-400, Busy: bg-yellow-500/10 text-yellow-400)
    Hostname (font-headline-sm text-on-surface)
    IP (font-mono-code text-on-surface-variant/70)
    Tags (px-2 py-0.5 bg-surface-variant text-on-surface-variant text-[11px] rounded-md)
    Footer: border-t border-outline-variant + "Connect" (text-primary) or disabled
    Offline: opacity-75 grayscale-[0.5]
  Add card placeholder: border-2 border-dashed border-outline-variant hover:border-primary

Footer stats: border-t border-outline-variant
  "Total Hosts: 24" + "Active Sessions: 3" + version "v2.4.1-stable"
```

**Files to modify**:
- `app/hosts/page.tsx` — page layout
- `components/hosts/HostList.tsx` — grid layout, filter bar
- `components/hosts/HostCard.tsx` — glass-card redesign
- `components/hosts/HostStatsCard.tsx` — footer stats
- `components/hosts/HostForm.tsx` — form with new tokens
- `components/hosts/HostImportExport.tsx` — buttons with new tokens
- `components/hosts/ConnectionButton.tsx` — button style
- `components/hosts/CopyableField.tsx` — mono-code styling

**Acceptance**: Hosts page matches mockup screenshot.

#### Task 4.2 — Terminal View
**File**: `apps/web/src/app/terminal/page.tsx` + child components

Layout from mockup `terminal_view.html`:
```
Tab bar (h-14, bg-surface-container, border-b border-outline-variant):
  Active tab: border-b-2 border-primary bg-surface-container-high + icon + label + close (hover)
  Inactive tab: border-b-2 border-transparent hover:bg-surface-variant
  "+" button: w-8 h-8 rounded hover:bg-surface-variant
  Right: host dropdown (bg-surface-container-highest border-outline-variant) + Connect (bg-primary-container text-on-primary-container)

Terminal pane (bg-black p-4 font-mono-code):
  Status overlay: absolute top-6 left-6, bg-surface-container-lowest/80, pulse dot
  Output: text-on-surface-variant, text-primary for highlights, text-tertiary for user@host
  SSH blocks: pl-4 border-l border-outline-variant/30 bg-surface-variant/10
  Prompt: text-tertiary + text-on-surface + text-primary-container (path) + text-on-surface ($)
```

**Files to modify**:
- `app/terminal/page.tsx` — page layout
- `components/terminal/terminal.tsx` — terminal pane
- `components/terminal/terminal-tabs.tsx` — tab bar
- `components/terminal/terminal-pane.tsx` — terminal output styling
- `components/terminal/terminal-toolbar.tsx` — controls bar

**Acceptance**: Terminal page matches mockup screenshot.

### Dispatch 5 — Vault + Settings Overview

#### Task 4.3 — Vault
**File**: `apps/web/src/app/vault/page.tsx` + child components

Layout from mockup `vault_credentials.html`:
```
Header: h1 "Secure Vault" + "encrypted" icon (text-[40px]) + subtitle
Toolbar: "Add Credential" (bg-primary text-on-primary) + import/export (border-outline-variant)

Locked state: centered card max-w-md mx-auto mt-20
  Shield icon (w-20 h-20 bg-surface-container-high rounded-full, text-[40px])
  "Vault Locked" title
  "Your credentials protected AES-256 encryption..."
  Password input + "Unlock" button (bg-primary)

Unlocked state: credential card grid (glass-card pattern)
  Type icon (SSH: key, Password: password, API: code) with bg tint
  Name + metadata (type, last used, tags)
  Actions: copy, view, delete (icon buttons, hover reveal)
  Footer: border-t + last modified
```

**Files to modify**:
- `app/vault/page.tsx`
- `components/vault/credential-card.tsx`
- `components/vault/credential-form.tsx`
- `components/vault/credential-list.tsx`
- `components/vault/import-export.tsx`

#### Task 4.4 — Settings Overview
**File**: `apps/web/src/app/settings/page.tsx`

Layout from mockup `settings_overview.html`:
```
H1 "Settings" + subtitle "Manage account configurations, security preferences, system behavior."

Category sections (3 groups, each with section label text-label-md uppercase tracking-widest text-on-surface-variant):

Core Account:
  Profile card: person icon + "Profile" + "Update identity, email, public-facing details."
  Security card: security icon + "Security" + "Multi-factor authentication, login history, password audits."

Developer Access:
  API Keys card: key icon + "API Keys" + "Manage programmatic access third-party integrations."
  WebAuthn card: verified_user icon + "WebAuthn" + "Register security keys passwordless authentication."

System:
  Appearance card: palette icon + "Appearance" + "Customize theme, density, typography preferences."
  Notifications card: notifications icon + "Notifications" + "Configure alerts channels, event triggers."
  Sessions card: history icon + "Sessions" + "View manage active past sessions."

Each card: bg-surface-container border border-outline-variant rounded-xl p-5
  Hover: border-primary + translateY(-1px) transition
  Layout: flex items-center gap-4
  Left: icon w-12 h-12 bg-surface-container-high rounded-xl
  Right: title font-headline-sm + description text-body-md text-on-surface-variant
  Click → navigate to sub-page
```

**Acceptance**: Settings overview matches mockup.

### Dispatch 6 — Settings Sub-Pages (Profile, Security, API Keys)

#### Task 4.5 — Settings/Profile
**File**: `app/settings/profile/page.tsx`

From mockup `profile_settings.html`:
```
H1 "Profile Settings" + subtitle
Avatar card: bg-surface-container rounded-xl p-6
  Avatar (w-20 h-20 rounded-full bg-secondary-container border-2 border-outline-variant)
  Upload + Remove buttons
Identity card: bg-surface-container rounded-xl p-6
  Full name, email, bio — bg-surface-container-lowest border-outline-variant
  Focus: focus:border-primary focus:ring-1 focus:ring-primary
Public profile card: visibility toggle + public URL
Save button: bg-primary text-on-primary
Cancel: border-outline-variant
```

#### Task 4.6 — Settings/Security
**File**: `app/settings/security/page.tsx`

From mockup `security_settings.html`:
```
H1 "Security Settings" + subtitle
MFA card: toggle (Material switch) + methods list (icon + name + status badge + Configure link)
Password card: current + new + confirm fields + strength bar
Login history card: table (timestamp, IP, device, location, status badge)
Active sessions card: session list with Revoke buttons
Security audit card: checklist with check/warning icons
```

#### Task 4.7 — Settings/API Keys
**File**: `app/settings/api-keys/page.tsx`

From mockup `api_keys_settings.html`:
```
H1 "API Keys" + subtitle "Manage programmatic access third-party integrations."
"Create New Key" button: bg-primary text-on-primary
Keys table: bg-surface-container rounded-xl border border-outline-variant
  Headers: bg-surface-container-high text-label-md text-on-surface-variant border-b border-outline-variant
  Columns: Name, Key (mono-code masked), Created, Last Used, Scopes, Status, Actions
  Row: border-b border-outline-variant hover:bg-surface-variant/30
  Actions: copy, rotate (refresh), revoke (delete_forever text-error)
Create key modal: bg-surface-container-high border-outline-variant
  Name input, scope checkboxes, expiry dropdown
  Generated key display: mono-code + copy button
Empty state: "No API keys" + key icon + "Create your first key"
```

### Dispatch 7 — Settings/Appearance + Settings/Notifications (NEW)

#### Task 4.8 — Settings/Appearance
**File**: `app/settings/appearance/page.tsx`

From mockups `appearance_settings.html` + active + updated states:
```
H1 "Appearance" + subtitle
Theme card: bg-surface-container rounded-xl p-6
  Theme options: Dark (active), Light, System — clickable preview cards
  Active: border-2 border-primary bg-primary/5 + check icon
  Inactive: border border-outline-variant hover:border-outline
Density card: Comfortable / Compact — radio-style
  Selected: bg-secondary-container text-on-secondary-container
Font size card: slider (Material-style) + preview text
Accent color card: color swatches (clickable circles)
Updated toast: bg-surface-container-highest border-outline-variant — "Appearance updated" + check
```

#### Task 4.9 — Settings/Notifications (NEW PAGE)
**File**: `app/settings/notifications/page.tsx` (NEW)

From mockup `notifications_settings.html` + interaction states:
```
H1 "Notifications" (or "Admin Console" per mockup) + subtitle
Notification channels card: bg-surface-container rounded-xl p-6
  Desktop Notifications: Material checkbox toggle
  Email Alerts: same toggle
  Quiet Hours: toggle + time pickers (FROM/TO, bg-surface-container-lowest border-outline-variant)
  Focus state: border-primary
Event triggers card:
  SSH login alerts, Host down, Credential access, Tunnel status — each with toggle + description
Integrations card:
  Slack: slider-style toggle + "Connect Workspace" button (bg-tertiary text-on-tertiary)
Recent activity card:
  Timeline list (icon + event + timestamp) + "Clear all" link
```

### Phase 4 Verification
After EACH dispatch:
```bash
cd 02-application/apps/web
npm run build 2>&1 | tail -20
npm run test 2>&1 | tail -30
rm -rf .next
```

### Phase 4 Commits (per dispatch)
```bash
git add 02-application/apps/web/src/app/hosts/ 02-application/apps/web/src/components/hosts/
git commit -m "feat(ui): redesign hosts dashboard with glass-card grid + filter bar"

git add 02-application/apps/web/src/app/terminal/ 02-application/apps/web/src/components/terminal/
git commit -m "feat(ui): redesign terminal with tabbed sessions + control bar"

# ... etc per page
```

---

## Phase 5 — Inferred Pages (26 pages)

> **Dispatches 8-10** | **Files**: 26 pages + child components | **Est. errors**: 10-20 per dispatch

### Dispatch 8 — Auth Pages (6) + Redirects (3) + Static (2)

#### Task 5.1 — Login
**File**: `app/login/page.tsx` + `LoginForm.tsx`
Pattern F (split screen): Left branding panel + Right form card
Fields: email, password, remember me, sign in button
Links: forgot password, register

#### Task 5.2 — Register
**File**: `app/register/page.tsx` + `RegisterForm.tsx`
Pattern F: Name, email, password, confirm, terms checkbox, create button

#### Task 5.3 — Forgot Password
**File**: `app/forgot-password/page.tsx` + `ForgotPasswordForm.tsx`
Pattern F: Email field, send reset link button

#### Task 5.4 — Reset Password
**File**: `app/reset-password/page.tsx` + `ResetPasswordForm.tsx`
Pattern F: New password, confirm, strength bar, reset button

#### Task 5.5 — MFA Verify
**File**: `app/mfa-verify/page.tsx`
Pattern F: Centered card, shield icon, 6-digit code input, verify button

#### Task 5.6 — Verify Email
**File**: `app/verify-email/page.tsx`
Pattern F: Centered card, email icon, "check your email", resend button

#### Task 5.7 — Redirects
- `app/profile/page.tsx` → `redirect('/settings/profile')`
- `app/security/page.tsx` → `redirect('/settings/security')`
- `app/page.tsx` → `redirect('/hosts')`

#### Task 5.8 — Terms
**File**: `app/terms/page.tsx`
DashboardLayout wrapper, H1 + prose content in bg-surface-container rounded-xl

#### Task 5.9 — Welcome
**File**: `app/welcome/page.tsx`
Full-screen centered, logo, checklist, get started button

### Dispatch 9 — Admin Pages (6) + Infrastructure (5)

#### Task 5.10 — Admin Dashboard
**File**: `app/admin/page.tsx`
Pattern A: Stats cards (6), recent activity feed, quick actions

#### Task 5.11 — Admin Audit
**File**: `app/admin/audit/page.tsx`
Data table: timestamp, user, action, resource, IP, status badges
Filter bar + export button

#### Task 5.12 — Admin Hosts
**File**: `app/admin/hosts/page.tsx`
Host table with admin controls, bulk actions

#### Task 5.13 — Admin Metrics
**File**: `app/admin/metrics/page.tsx`
Time range selector + chart cards (4) + resource table

#### Task 5.14 — Admin Sessions
**File**: `app/admin/sessions/page.tsx`
Session table with terminate actions

#### Task 5.15 — Admin Users
**File**: `app/admin/users/page.tsx`
User table with role badges + add/invite buttons

#### Task 5.16 — Discovery
**File**: `app/discovery/page.tsx`
Scan controls + progress bar + results table + scan history

#### Task 5.17 — Files
**File**: `app/files/page.tsx` + `components/file-manager/*`
3-panel layout: file tree + file list/grid + preview pane
Toolbar: breadcrumb, view toggle, upload, new folder

#### Task 5.18 — Host Detail
**File**: `app/hosts/[id]/page.tsx`
Host header card + tab bar (Overview/Sessions/Credentials/Settings)
Overview: stats grid (4 cards)

#### Task 5.19 — Tunnels
**File**: `app/tunnels/page.tsx` + `components/tunnels/*`
Stats cards (4) + tunnel table with status badges + create button

#### Task 5.20 — Vault Share
**File**: `app/vault/share/page.tsx`
Share form card + active shares table

### Dispatch 10 — Session + Settings Sub-Pages + Tests

#### Task 5.21 — Sessions List
**File**: `app/sessions/page.tsx`
Stats cards + session table with recording links

#### Task 5.22 — Session Recordings
**File**: `app/sessions/recordings/page.tsx`
Recording list + player with timeline scrubber

#### Task 5.23 — Settings/Sessions
**File**: `app/settings/sessions/page.tsx`
Timeout card, recording settings, concurrent sessions, security card

#### Task 5.24 — Settings/WebAuthn
**File**: `app/settings/webauthn/page.tsx`
Registered keys list + register new key + test authentication

#### Task 5.25 — Update Tests
**Files**: All `__tests__/` files
- Update selectors for new layout structure (sidebar, new class names)
- Mock Sidebar component in test setup
- Fix any snapshot tests

### Phase 5 Verification
```bash
cd 02-application/apps/web
npm run build 2>&1 | tail -20
npm run test 2>&1 | tail -30
rm -rf .next
```

### Phase 5 Commits (per dispatch)
```bash
git add 02-application/apps/web/src/app/login/ 02-application/apps/web/src/components/auth/LoginForm.tsx
git commit -m "feat(ui): redesign login with split-screen auth layout"
# ... etc per page group
```

---

## Phase 6 — Polish & QA

> **Dispatch 11** | **Files**: various | **Est. errors**: <10

### Tasks

#### Task 6.1 — Responsive design
- Test all pages at breakpoints: sm (640), md (768), lg (1024), xl (1280)
- Sidebar: `hidden md:flex` (collapsible on mobile)
- Card grids: responsive column counts
- TopNav: hide nav links on mobile, show hamburger menu
- FAB: visible only on mobile (`md:hidden`)

#### Task 6.2 — Accessibility
- All interactive elements have `aria-label` or visible text
- Focus states: `focus:ring-2 focus:ring-primary`
- Color contrast: verify all text/background combos meet WCAG AA
- Material Icons: `aria-hidden="true"` (already in wrapper)

#### Task 6.3 — Loading states
- Page-level: `<Skeleton>` components matching final layout shape
- Card-level: skeleton with `bg-surface-container-low animate-pulse`
- Table-level: skeleton rows

#### Task 6.4 — Empty states
- No hosts: icon + "No hosts configured" + "Add your first host" button
- No sessions: icon + "No active sessions"
- No credentials: icon + "Vault is empty" + "Add credential"
- No tunnels: icon + "No tunnels configured"
- No API keys: icon + "No API keys" + "Create your first key"
- No recordings: icon + "No recordings available"

#### Task 6.5 — Test fixes
- Update all `__tests__/` files:
  - Mock `Sidebar` and `TopNav` in test setup
  - Update CSS class assertions
  - Update DOM selectors for new layout
- Run: `npm run test`
- Target: 0 failing tests

#### Task 6.6 — Final visual QA
- Open dev server: `npm run dev`
- Compare each mockup page against screenshot
- Check: colors, spacing, typography, icon rendering, hover states, focus states
- Document any discrepancies

### Phase 6 Verification
```bash
cd 02-application/apps/web
npm run build 2>&1 | tail -20
npm run test 2>&1
rm -rf .next
```

### Phase 6 Commit
```bash
git add 02-application/apps/web/
git commit -m "feat(ui): polish responsive, accessibility, loading/empty states, fix tests"
```

---

## Post-Execution (Ame, no dispatch)

### Final Verification
```bash
cd /home/ubuntu/projects/vexa/02-application/apps/web
npm run build
npm run test
# All green

cd /home/ubuntu/projects/vexa/02-application/apps/api
go test ./...
go build ./...
# All green (untouched, should pass)
```

### Merge to main
```bash
cd /home/ubuntu/projects/vexa
git checkout main
git merge feat/ui-redesign --no-ff
```

### Push
```bash
git push origin main
git subtree push --prefix=02-application https://github.com/soumabali/vexa.git main
```

### Document
1. Write session note: `03-history/sessions/2026-06-30-ui-redesign.md`
2. Update Obsidian vault: `infra/vexa — UI Redesign.md`
3. Update plan index: `06-temp/plans/README.md`

### Cleanup
```bash
rm 02-application/apps/web/src/app/globals.css.bak
rm 02-application/apps/web/tailwind.config.ts.bak
rm -rf 02-application/.hermes/vexa-redesign-ref
git checkout -- 02-application/apps/web/.next  # restore if needed
```

---

## Dispatch Summary & Progress Tracker

### Dispatch Table

| # | Phase | Scope | Files | Est. Errors | Status |
|---|-------|-------|-------|-------------|--------|
| 0 | Phase 0 | Pre-flight (branch, backup) | — | 0 | ⬜ pending |
| 1 | Phase 1 | Design tokens + Material Symbols | 4 | <10 | ⬜ pending |
| 2 | Phase 2 | Layout shell (sidebar + topnav) | 5 | <15 | ⬜ pending |
| 3a | Phase 3a | Core components restyle | 14 | 15-20 | ⬜ pending |
| 3b | Phase 3b | Form + overlay components | 14 | 15-20 | ⬜ pending |
| 4 | Phase 4 | Hosts + Terminal | 13 | 15-25 | ⬜ pending |
| 5 | Phase 4 | Vault + Settings Overview | 8 | 10-15 | ⬜ pending |
| 6 | Phase 4 | Settings: Profile + Security + API Keys | 6 | 10-15 | ⬜ pending |
| 7 | Phase 4 | Settings: Appearance + Notifications (new) | 3 | 5-10 | ⬜ pending |
| 8 | Phase 5 | Auth (6) + Redirects (3) + Static (2) | 13 | 10-15 | ⬜ pending |
| 9 | Phase 5 | Admin (6) + Infrastructure (5) | 20 | 15-25 | ⬜ pending |
| 10 | Phase 5 | Sessions (2) + Settings sub (2) + Tests | 10 | 10-20 | ⬜ pending |
| 11 | Phase 6 | Polish + test fixes | various | <10 | ⬜ pending |

**Total**: 11 dispatches, ~167 files touched

### Status Legend
- ⬜ pending — belum dimulai
- 🔄 in_progress — dispatch sedang berjalan
- ✅ completed — dispatch selesai, build+test hijau, committed
- ❌ failed — dispatch gagal, butuh intervention
- ⏸️ blocked — menunggu dependency

### Autonomous Dispatch Sequence

```
Phase 0 (Ame): branch + backup
  ↓
Phase 1 → Dispatch 1 → verify → commit → update progress
  ↓
Phase 2 → Dispatch 2 → verify → commit → update progress
  ↓
Phase 3 → Dispatch 3a → verify → commit → update progress
         Dispatch 3b → verify → commit → update progress
  ↓
Phase 4 → Dispatch 4 → verify → commit → update progress
         Dispatch 5 → verify → commit → update progress
         Dispatch 6 → verify → commit → update progress
         Dispatch 7 → verify → commit → update progress
  ↓
Phase 5 → Dispatch 8 → verify → commit → update progress
         Dispatch 9 → verify → commit → update progress
         Dispatch 10 → verify → commit → update progress
  ↓
Phase 6 → Dispatch 11 → verify → commit → update progress
  ↓
Post-Execution (Ame): merge + push + document
```

### Per-Dispatch Execution Checklist

Untuk setiap dispatch, Ame jalankan urutan ini:

```text
1. READ progress file → determine current dispatch
2. PRE-FLIGHT: verify branch + clean working tree
3. READ plan section for dispatch N
4. CREATE Claude Code dispatch prompt (from plan task description)
5. DISPATCH to Claude Code
6. WAIT for result
7. VERIFY: npm run build
8. VERIFY: npm run test
9. IF FAIL: retry (max 2x) → if still fail: STOP, report to Dhar
10. IF PASS: commit with conventional message
11. UPDATE progress file (mark dispatch completed, advance current_dispatch)
12. PROCEED to next dispatch
```

### Context Budget Per Dispatch

| Step | Context Cost |
|------|-------------|
| Read progress file | ~200 tokens |
| Read plan section | ~500-1000 tokens |
| Claude Code dispatch prompt | ~800 tokens |
| Claude Code result summary | ~500-1000 tokens |
| Build output (tail -20) | ~200 tokens |
| Test output (tail -30) | ~300 tokens |
| Commit + progress update | ~200 tokens |
| **Total per dispatch** | **~3000-4000 tokens** |

11 dispatches × ~3500 tokens = ~38,500 tokens untuk Ame context.
Well within limits — context overflow risk minimal.

### Fallback: Manual Resume Command

Jika Ame crash dan perlu resume di session baru, Dhar bisa bilang:

```
Resume vexa UI redesign dari progress file
```

Dan Ame akan:
1. Read `02-application/.hermes/vexa-redesign-progress.json`
2. Verify git state
3. Resume from `current_dispatch`

---

## Pitfalls (Non-Negotiable)

1. **HANYA edit `02-application/apps/web/`.** Tidak commit/push dari Claude Code. Tidak edit root repo. Tidak edit backend.
2. **MAX 30 errors per dispatch.** Fix inline, jangan `eslint-disable` sebagai escape hatch.
3. **Per-task commit dengan conventional prefix.** Format: `feat(ui): ...`
4. **Restore generated artifacts** jika harus generate untuk verifikasi: `.next/`, `tsconfig.tsbuildinfo`. **JANGAN commit.**
5. **Pakai `useAsyncData`** untuk fetching, bukan raw useEffect + setState.
6. **Pakai `Dialog` dari `@/components/ui/dialog`** untuk modal, jangan library lain.
7. **Pakai `toast` dari `sonner`** untuk notifikasi.
8. **Jangan hapus atau rename file** kecuali eksplisit disebutkan di plan.
9. **Jangan ubah import paths** yang sudah ada (kecuali `lucide-react` → `MaterialIcon`).
10. **Test harus passing** setelah setiap dispatch. Jika test break, fix di dispatch yang sama.
11. **Jangan commit `.bak` files** — hapus setelah verification.
12. **Material Symbols**: pakai `font-variation-settings` untuk fill/weight, bukan CSS classes.