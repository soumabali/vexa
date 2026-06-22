# vexa Desktop (Roadmap / Experimental)

Tauri v2 desktop application for **vexa — Complete SSH Manager**.

## Status

- **Phase:** Roadmap / experimental scaffolding
- **Milestone:** MVP planned after core web + API features stabilize
- **Current contents:** Rust + Tauri v2 skeleton and a placeholder Next.js frontend

## Planned Architecture

- **Backend:** Rust (Tauri v2)
  - Secure clipboard bridge
  - Encrypted window state
  - System tray integration
  - Global shortcuts
  - Auto-updater

- **Frontend:** Next.js static export
  - xterm.js terminal with WebGL renderer
  - WebSocket connection to the Go backend
  - Custom dark theme
  - Resizable panels

## Development

```bash
# Install dependencies (when Rust/Node are available)
cd src-frontend && pnpm install
cd .. && pnpm install

# Run in development mode
pnpm tauri dev

# Build for production
pnpm tauri build
```

## Roadmap

1. Native SSH terminal window
2. Secure credential copy/paste bridge
3. System tray quick-connect
4. Packaged installers (Windows, macOS, Linux)

## Notes

- This package is not required for the web or API builds.
- Rust toolchain must be installed to build or test this app.
