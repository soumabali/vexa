# Sprint 5 — Desktop Clients

> **Phase:** 4 | **Weeks:** 14–16 | **Duration:** 15 working days
> **Sprint Goal:** Build cross-platform desktop app with native terminal and system integration
> **Story Points:** 55
> **Status:** Not Started
> **Depends On:** Sprint 4 (Web Client)

---

## User Stories

### US-028: Tauri Desktop Shell
**As a** user, **I want** a native desktop app, **so that** I have better performance than browser.

**Acceptance Criteria:**
- [ ] Tauri v2 project setup
- [ ] Window management (minimize, maximize, fullscreen)
- [ ] System tray integration (minimize to tray)
- [ ] Native menu bar (File, Edit, View, Help)
- [ ] Keyboard shortcuts (Ctrl+T new tab, Ctrl+W close tab)
- [ ] Multiple window support
- [ ] Auto-updater integration

**Definition of Done:**
- [ ] App launches in < 3s
- [ ] Memory usage < 200MB idle
- [ ] Works on Windows 10+, macOS 12+, Linux Ubuntu 20+
- [ ] Code-signed binaries (Windows, macOS)

**Story Points:** 13

---

### US-029: Native Terminal
**As a** user, **I want** a native terminal in desktop app, **so that** I have better performance than web terminal.

**Acceptance Criteria:**
- [ ] Native terminal widget (not WebView-based)
- [ ] Tabbed interface with drag-and-drop reordering
- [ ] Split panes (horizontal/vertical)
- [ ] Search in terminal (Ctrl+Shift+F)
- [ ] Export terminal output to file
- [ ] Custom color schemes

**Definition of Done:**
- [ ] Terminal latency < 10ms (vs < 50ms web)
- [ ] Supports Unicode and emoji
- [ ] Copy/paste works with system clipboard
- [ ] Font rendering crisp on HiDPI displays

**Story Points:** 13

---

### US-030: Desktop File Manager
**As a** user, **I want** native file manager, **so that** I can transfer files efficiently.

**Acceptance Criteria:**
- [ ] Two-pane layout (local + remote)
- [ ] Drag-and-drop between panes
- [ ] Native file picker integration
- [ ] Context menus (right-click)
- [ ] File preview (images, text)
- [ ] Bulk operations

**Definition of Done:**
- [ ] Transfer speed > 90% of native SFTP client
- [ ] Handles 10,000+ files in directory
- [ ] Progress bar with cancel button
- [ ] Queued transfers with priority

**Story Points:** 13

---

### US-031: System Integration
**As a** user, **I want** OS-level integration, **so that** the app feels native.

**Acceptance Criteria:**
- [ ] URL scheme handler (`vexa://connect/...`)
- [ ] SSH config import (`~/.ssh/config`)
- [ ] System notifications (connection status, transfers complete)
- [ ] Global hotkey (Ctrl+Shift+S to show/hide)
- [ ] Auto-start on login (configurable)
- [ ] Dark/light mode follows OS preference

**Definition of Done:**
- [ ] URL scheme opens app and connects to host
- [ ] SSH config parsed correctly (includes, wildcards)
- [ ] Notifications use native OS style
- [ ] Hotkey works even when app minimized

**Story Points:** 8

---

### US-032: Offline Support
**As a** user, **I want** offline access to host list, **so that** I can work without internet.

**Acceptance Criteria:**
- [ ] Host list cached locally (SQLite or file)
- [ ] Credential cache (encrypted, time-limited)
- [ ] Offline indicator when disconnected
- [ ] Sync on reconnection
- [ ] Conflict resolution for concurrent edits

**Definition of Done:**
- [ ] App works fully when backend unreachable
- [ ] Cache encrypted with device key
- [ ] Sync resolves conflicts without data loss
- [ ] Cache auto-expires after 7 days

**Story Points:** 8

---

## Dependencies

| Dependency | Source | Impact |
|------------|--------|--------|
| Sprint 4 completion | Previous sprint | Blocks desktop UI |
| Tauri v2 | Framework | Blocks desktop features |
| Code signing certs | Apple/Microsoft | Blocks distribution |
| Native terminal lib | `alacritty`/`wezterm` | Blocks native terminal |

---

## Risks and Mitigations

| Risk | Probability | Impact | Mitigation | Owner |
|------|-------------|--------|------------|-------|
| Tauri v2 stability | Medium | High | Use latest stable, have fallback to v1 | Desktop Lead |
| Cross-platform差异 | Medium | Medium | Platform-specific code branches, CI testing | Desktop |
| Code signing complexity | Medium | High | Start signing process early, use CI | DevOps |
| Native terminal performance | Medium | High | Benchmark early, fallback to web terminal | Desktop |

---

## Technical Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Desktop framework | Tauri v2 | Smaller bundle, Rust performance |
| Terminal widget | `alacritty` backend | GPU-accelerated, cross-platform |
| Local storage | SQLite | SQL queries, native performance |
| Auto-updater | Tauri updater | Built-in, secure |
| Notifications | Tauri notification API | Cross-platform native |

---

## Sprint Review Checklist

- [ ] Desktop app launches on Windows, macOS, Linux
- [ ] Native terminal latency < 10ms
- [ ] File manager tested with 10GB transfers
- [ ] System integration works (URL, hotkey, notifications)
- [ ] Offline mode tested for 24h
- [ ] Auto-updater tested with staged rollout
- [ ] Code signing verified on all platforms
- [ ] Sprint retrospective completed

---

## Sprint Retrospective Notes

*To be filled after sprint completion*

**What went well:**
-

**What could be improved:**
-

**Action items for next sprint:**
-
