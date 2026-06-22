# vexa Desktop

Tauri v2 desktop application for vexa — Complete SSH Manager (roadmap).

## Architecture

- **Backend**: Rust (Tauri v2)
  - Clipboard bridge with secure copy
  - Window state management (encrypted at rest)
  - System tray integration
  - Global shortcuts
  - Auto-updater

- **Frontend**: Next.js 14 (static export)
  - xterm.js terminal with WebGL renderer
  - WebSocket connection to Go backend
  - Custom dark theme
  - Resizable panels

## Security

- Strict CSP headers
- AES-256-GCM encrypted window state
- Clipboard auto-wipe after timeout
- No DevTools in production
- No inline scripts

## Development

```bash
# Install dependencies
cd src-frontend && pnpm install
cd .. && pnpm install

# Run in development mode
pnpm tauri dev

# Build for production
pnpm tauri build
```

## Global Shortcuts

- `Ctrl+Shift+T` - Quick Connect
- `Ctrl+Shift+H` - Toggle Window
- `Ctrl+Shift+C` - Copy to clipboard
- `Ctrl+Shift+V` - Paste from clipboard

## File Structure

```
apps/desktop/
├── src/
│   ├── main.rs              # Tauri entry point
│   ├── lib.rs               # Library exports
│   ├── commands/
│   │   ├── clipboard.rs     # Native clipboard bridge
│   │   ├── window.rs        # Window state management
│   │   └── system_tray.rs   # System tray integration
│   └── security/
│       └── permissions.rs   # Permission system
│       └── mod.rs           # Encryption utilities
├── src-frontend/            # Next.js frontend
│   ├── app/
│   │   ├── layout.tsx
│   │   └── page.tsx
│   └── components/
│       ├── Terminal.tsx
│       ├── HostList.tsx
│       └── ConnectionBar.tsx
├── tauri.conf.json          # Tauri configuration
├── Cargo.toml               # Rust dependencies
└── package.json             # Frontend dependencies
```
