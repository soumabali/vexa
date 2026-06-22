# Sprint 4 — Web Client

> **Phase:** 3 | **Weeks:** 11–13 | **Duration:** 15 working days  
> **Sprint Goal:** Build functional web client with terminal, host management, and credential vault UI  > **Story Points:** 55  > **Status:** Not Started  > **Depends On:** Sprint 3 (Protocol Layer)

---

## User Stories

### US-023: Host Management UI
**As a** user, **I want** a web interface to manage hosts, **so that** I can organize my server inventory visually.

**Acceptance Criteria:**
- [ ] Host list with search, filter, and sort
- [ ] Host creation form with validation
- [ ] Host editing inline or modal
- [ ] Host deletion with confirmation
- [ ] Group/ folder organization (tree view)
- [ ] Favorite hosts (pin to top)
- [ ] Bulk operations (delete, move to group)

**Definition of Done:**
- [ ] Renders 1000 hosts without performance issues
- [ ] Search responds in < 100ms
- [ ] Mobile responsive (works on tablet)
- [ ] Accessible (WCAG 2.1 AA)

**Story Points:** 13

---

### US-024: Terminal Component
**As a** user, **I want** an embedded terminal in the browser, **so that** I can execute commands without desktop client.

**Acceptance Criteria:**
- [ ] Terminal component using `xterm.js`
- [ ] Tabbed interface (multiple sessions)
- [ ] Color scheme support (dark/light themes)
- [ ] Font size adjustment
- [ ] Copy/paste support
- [ ] Terminal resize (responsive)
- [ ] Status bar (connection state, latency)

**Definition of Done:**
- [ ] Terminal renders 80x24 correctly
- [ ] Supports ANSI color codes
- [ ] Copy/paste works across browsers
- [ ] Memory usage stable with 10 tabs

**Story Points:** 13

---

### US-025: Credential Vault UI
**As a** user, **I want** a secure credential manager UI, **so that** I can store and use passwords/keys.

**Acceptance Criteria:**
- [ ] Vault unlock with master password
- [ ] Credential list with type icons
- [ ] Credential creation (password, SSH key, API token)
- [ ] Credential editing
- [ ] Copy to clipboard (with timeout)
- [ ] Credential sharing with team members
- [ ] Auto-lock after timeout

**Definition of Done:**
- [ ] Credentials never displayed in plaintext (reveal on demand)
- [ ] Clipboard cleared after 30s
- [ ] Auto-lock after 5 min inactivity
- [ ] Audit log shows all credential access

**Story Points:** 13

---

### US-026: Dashboard & Analytics
**As an** administrator, **I want** a dashboard with usage analytics, **so that** I can monitor team activity.

**Acceptance Criteria:**
- [ ] Active sessions count (real-time)
- [ ] Connection history (last 24h)
- [ ] Top connected hosts
- [ ] User activity summary
- [ ] Audit log viewer with filters
- [ ] Export reports (PDF, CSV)
- [ ] Health status of all hosts

**Definition of Done:**
- [ ] Dashboard loads in < 2s
- [ ] Real-time updates via WebSocket
- [ ] Charts render correctly on mobile
- [ ] Data refreshes every 30s

**Story Points:** 8

---

### US-027: User Settings
**As a** user, **I want** to customize my experience, **so that** the app works the way I prefer.

**Acceptance Criteria:**
- [ ] Profile editing (name, email, avatar)
- [ ] Password change
- [ ] MFA setup/disable
- [ ] Theme selection (dark/light/system)
- [ ] Default terminal settings (font, size)
- [ ] Notification preferences
- [ ] Session timeout preference

**Definition of Done:**
- [ ] All settings persist across sessions
- [ ] Theme changes apply immediately
- [ ] MFA setup tested with authenticator apps
- [ ] Settings accessible from any page

**Story Points:** 8

---

## Dependencies

| Dependency | Source | Impact |
|------------|--------|--------|
| Sprint 3 completion | Previous sprint | Blocks terminal and file manager |
| Design system | Sprint 0 | Blocks UI consistency |
| API endpoints | Sprint 2 | Blocks all data operations |
| WebSocket terminal | Sprint 3 | Blocks terminal component |

---

## Risks and Mitigations

| Risk | Probability | Impact | Mitigation | Owner |
|------|-------------|--------|------------|-------|
| `xterm.js` performance | Medium | High | Lazy loading, virtual scrolling | Frontend |
| Browser compatibility | Medium | Medium | Test on Chrome, Firefox, Safari, Edge | QA |
| Mobile usability | Medium | Medium | Touch-friendly design, responsive testing | Design |
| Credential UI security | High | Critical | No plaintext display, clipboard timeout | Security |

---

## Technical Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| UI framework | Next.js 16 + React 19 | App Router, Server Components |
| Component library | shadcn/ui + Tailwind | Headless, customizable |
| State management | Zustand | Lightweight, TypeScript-friendly |
| Terminal | `xterm.js` + `xterm-addon-fit` | Industry standard |
| Charts | Recharts | React-native, customizable |
| Icons | Lucide React | Lightweight, consistent |

---

## Sprint Review Checklist

- [ ] Host management tested with 1000 hosts
- [ ] Terminal works with SSH, RDP, VNC
- [ ] Credential vault UI security reviewed
- [ ] Dashboard renders on mobile devices
- [ ] Settings persist across sessions
- [ ] Accessibility audit passed
- [ ] Cross-browser testing completed
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
