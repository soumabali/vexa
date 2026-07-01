---
name: Vexa Core
colors:
  surface: '#10131a'
  surface-dim: '#10131a'
  surface-bright: '#363941'
  surface-container-lowest: '#0b0e15'
  surface-container-low: '#191b23'
  surface-container: '#1d2027'
  surface-container-high: '#272a31'
  surface-container-highest: '#32353c'
  on-surface: '#e1e2ec'
  on-surface-variant: '#c2c6d6'
  inverse-surface: '#e1e2ec'
  inverse-on-surface: '#2e3038'
  outline: '#8c909f'
  outline-variant: '#424754'
  surface-tint: '#adc6ff'
  primary: '#adc6ff'
  on-primary: '#002e6a'
  primary-container: '#4d8eff'
  on-primary-container: '#00285d'
  inverse-primary: '#005ac2'
  secondary: '#b9c8de'
  on-secondary: '#233143'
  secondary-container: '#39485a'
  on-secondary-container: '#a7b6cc'
  tertiary: '#ffb786'
  on-tertiary: '#502400'
  tertiary-container: '#df7412'
  on-tertiary-container: '#461f00'
  error: '#ffb4ab'
  on-error: '#690005'
  error-container: '#93000a'
  on-error-container: '#ffdad6'
  primary-fixed: '#d8e2ff'
  primary-fixed-dim: '#adc6ff'
  on-primary-fixed: '#001a42'
  on-primary-fixed-variant: '#004395'
  secondary-fixed: '#d4e4fa'
  secondary-fixed-dim: '#b9c8de'
  on-secondary-fixed: '#0d1c2d'
  on-secondary-fixed-variant: '#39485a'
  tertiary-fixed: '#ffdcc6'
  tertiary-fixed-dim: '#ffb786'
  on-tertiary-fixed: '#311400'
  on-tertiary-fixed-variant: '#723600'
  background: '#10131a'
  on-background: '#e1e2ec'
  surface-variant: '#32353c'
typography:
  headline-lg:
    fontFamily: Inter
    fontSize: 30px
    fontWeight: '700'
    lineHeight: 38px
    letterSpacing: -0.02em
  headline-md:
    fontFamily: Inter
    fontSize: 24px
    fontWeight: '600'
    lineHeight: 32px
    letterSpacing: -0.01em
  headline-sm:
    fontFamily: Inter
    fontSize: 20px
    fontWeight: '600'
    lineHeight: 28px
  body-lg:
    fontFamily: Inter
    fontSize: 16px
    fontWeight: '400'
    lineHeight: 24px
  body-md:
    fontFamily: Inter
    fontSize: 14px
    fontWeight: '400'
    lineHeight: 20px
  label-md:
    fontFamily: Inter
    fontSize: 12px
    fontWeight: '500'
    lineHeight: 16px
    letterSpacing: 0.02em
  mono-code:
    fontFamily: JetBrains Mono
    fontSize: 13px
    fontWeight: '400'
    lineHeight: 20px
  headline-lg-mobile:
    fontFamily: Inter
    fontSize: 24px
    fontWeight: '700'
    lineHeight: 32px
rounded:
  sm: 0.25rem
  DEFAULT: 0.5rem
  md: 0.75rem
  lg: 1rem
  xl: 1.5rem
  full: 9999px
spacing:
  base: 4px
  xs: 4px
  sm: 8px
  md: 16px
  lg: 24px
  xl: 32px
  gutter: 20px
  margin: 24px
---

## Brand & Style

The design system for this SSH and server management application is built on a foundation of **Modern Minimalism** with a **Technical/Corporate** edge. It is designed specifically for DevOps engineers and system administrators who require a high-density, low-friction environment that feels both high-end and utilitarian.

The aesthetic avoids the "generic SaaS" look by utilizing a deep, monochromatic base layered with precision-engineered components. It uses a "Dark-First" philosophy, where depth is communicated through subtle border highlights and tonal shifts rather than heavy drop shadows. The emotional goal is to evoke a sense of **calm control, security, and technical mastery.** 

Key stylistic pillars include:
- **Precision Engineering:** Sharp, intentional spacing and consistent geometric alignment.
- **Subtle Depth:** Utilizing 1px strokes and low-opacity fills to create a layered "glass-on-dark" effect.
- **Content-Focused:** Minimizing decorative elements to ensure terminal outputs and server logs remain the primary focal point.

## Colors

The palette is anchored by a sophisticated deep navy/slate foundation. This reduces eye strain during long terminal sessions while providing a premium, developer-centric environment.

- **Primary (Vexa Blue):** An electric, high-visibility blue used exclusively for primary actions, active states, and focus indicators.
- **Neutral/Surface:** A range of slates (`#0F172A` to `#334155`) define the hierarchy. The deepest shade is reserved for the global background, with progressively lighter shades used for cards, modals, and input fields.
- **Functional Colors:** Success, Error, and Warning colors are desaturated slightly to prevent "vibrating" against the dark background, ensuring they remain legible and professional.
- **Accents:** Subtle grays are used for secondary text and non-interactive borders to maintain a low-contrast, non-distracting UI layout.

## Typography

This design system employs a dual-font strategy. **Inter** provides a clean, highly legible interface for all UI controls and navigation, while **JetBrains Mono** is utilized for terminal outputs, code snippets, and configuration values.

- **Scale:** The system uses a tight scale to accommodate information-dense screens.
- **Weights:** Use 'Regular' (400) for body text and 'SemiBold' (600) or 'Bold' (700) for headings to create clear contrast without needing excessive size changes.
- **Mono-spacing:** The terminal font is slightly smaller (13px) to maximize the amount of code visible on screen while maintaining a generous line height (20px) for legibility.

## Layout & Spacing

The design system utilizes a **Fluid Grid** with fixed-width sidebars for navigation. The primary workspace expands to fill the screen, supporting the multi-pane environment necessary for server management.

- **Sidebar:** Fixed at 260px for desktop, collapsible to 64px (icons only).
- **Rhythm:** A 4px baseline grid ensures vertical rhythm. Spacing is intentionally generous (24px - 32px) between major sections to prevent the technical content from feeling cluttered.
- **Density:** In "Terminal" or "List" views, a high-density mode can be toggled which reduces vertical padding from `16px` to `8px`.
- **Breakpoints:**
  - Mobile (< 768px): Single column, hidden sidebar (hamburger menu).
  - Tablet (768px - 1024px): Collapsed sidebar, fluid main content.
  - Desktop (> 1024px): Full expanded sidebar, 12-column internal grid for dashboard widgets.

## Elevation & Depth

Visual hierarchy is achieved through **Tonal Layering** and **Low-Contrast Outlines**.

1.  **Level 0 (Background):** Deepest Navy (`#0F172A`). Used for the main app canvas.
2.  **Level 1 (Surface):** Slate (`#1E293B`). Used for cards and secondary panels. Features a subtle 1px border (`#334155`).
3.  **Level 2 (Active/Floating):** Lighter Slate (`#334155`). Used for hovered states or active tabs.
4.  **Level 3 (Modals/Popovers):** Surface color with a soft, 20% opacity black shadow (20px blur) and a slightly brighter 1px border to distinguish from the background.

Shadows are never pure black; they are tinted with the primary navy tone to ensure they feel like a natural part of the environment.

## Shapes

The shape language is controlled and modern. 

- **Containers & Cards:** Use a standard `8px` (0.5rem) radius.
- **Buttons & Inputs:** Follow the standard `8px` radius to maintain a consistent internal rhythm.
- **Inner Elements:** Elements nested inside containers (like small tags or chips) should use a `4px` or `6px` radius to maintain visual harmony (the "inner radius < outer radius" rule).
- **Terminal View:** The main terminal viewport remains sharp or with very minimal `4px` rounding to emphasize its technical, "full-screen" nature.

## Components

### Buttons
- **Primary:** Solid 'Vexa Blue' with white text. 
- **Secondary:** Transparent background with a 1px border (`#334155`) and secondary gray text. Turns slightly lighter on hover.
- **Danger:** Ghost style with red text and border, or solid red for destructive final actions.

### Input Fields
- Darker background than the card they sit on.
- 1px border (`#334155`) that transitions to 'Vexa Blue' on focus.
- Labels are always `label-md` and placed above the field.

### Cards
- Use for grouping host details or settings. 
- No shadow by default; defined by a 1px border. 
- Header section separated by a subtle 1px horizontal rule.

### Navigation Sidebar
- High contrast between active and inactive states.
- Active item uses a "Left Indicator" (2px blue vertical line) and a subtle background tint.

### Terminal Pane
- Pure black background (`#000000`) or the deepest navy.
- Syntax highlighting uses the brand's functional colors (Success for directories, Warning for executable files, etc.).