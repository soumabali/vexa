# vexa — vexa: Task Breakdown

> **Classification:** Security-Critical Application  
> **Owner:** Dhar  
> **Status:** Draft — PM Review  
> **Last Updated:** 2026-05-27

---

## Phase Overview

| Phase | Name | Duration | Goal |
|-------|------|----------|------|
| 0 | Foundation | 2 weeks | Project scaffolding, CI/CD, security baseline |
| 1 | Core Backend | 4 weeks | Tower API, auth, vault, database |
| 2 | Protocol Layer | 4 weeks | SSH/RDP/VNC proxies with sandboxing |
| 3 | Web Client | 3 weeks | PWA dashboard + terminal |
| 4 | Desktop Clients | 3 weeks | Windows + Linux native apps |
| 5 | Mobile Client | 3 weeks | Android app |
| 6 | Security Hardening | 3 weeks | Penetration tests, hardening, compliance |
| 7 | Release | 2 weeks | Packaging, docs, launch |

**Total Estimated Duration:** 24 weeks (~6 months)

---

## Phase 0: Foundation (Weeks 1–2)

### 0.1 Project Scaffolding
- [ ] Initialize monorepo structure (`/apps/api`, `/apps/web`, `/apps/desktop`, `/apps/mobile`, `/packages/ssh-core`, `/docs`, `/infra`)
- [ ] Set up Go modules and dependency management
- [ ] Set up Next.js + TypeScript project for web
- [ ] Set up Tauri v2 project for desktop (Rust + WebView)
- [ ] Set up Flutter project for mobile (iOS/Android)
- [ ] Define coding standards (Go: `gofmt` + `golangci-lint`, TS: `eslint` + `prettier`, Rust: `rustfmt` + `clippy`)

### 0.2 CI/CD Pipeline
- [ ] GitHub Actions / GitLab CI configuration
- [ ] Automated unit test execution for all components
- [ ] SAST integration (`semgrep`, `gosec`, `eslint-security`)
- [ ] SCA integration (`snyk`, `dependabot`, `renovate`)
- [ ] Container image build + scan with `trivy`
- [ ] Artifact signing with Cosign

### 0.3 Security Baseline
- [ ] Create `SECURITY.md` with vulnerability disclosure policy
- [ ] Define secret management strategy (1Password / Vault for team)
- [ ] Generate initial threat model (STRIDE)
- [ ] Set up pre-commit hooks for secret scanning (`gitleaks`)
- [ ] Establish secure coding guidelines document

### 0.4 Infrastructure Skeleton
- [ ] Docker + Docker Compose development environment
- [ ] PostgreSQL schema migration tool (`golang-migrate`)
- [ ] Local TLS certificate generation (mkcert)
- [ ] Prometheus + Grafana dev stack

### 0.5 Deliverables
- Working development environment (`docker-compose up`)
- CI pipeline green on empty projects
- Security policy published

---

## Phase 1: Core Backend (Weeks 3–6)

### 1.1 Database Layer
- [ ] Implement PostgreSQL schema (users, devices, hosts, credentials, sessions, audit_log)
- [ ] Migration scripts with rollback tests
- [ ] Row-level security (RLS) policies for multi-tenancy
- [ ] Connection pooling and health checks
- [ ] Database seeding for development

### 1.2 Authentication Service
- [ ] User registration with email verification
- [ ] Password login with Argon2id hashing
- [ ] TOTP enrollment and verification (`otp` library)
- [ ] WebAuthn / FIDO2 registration and authentication
- [ ] JWT access token + secure refresh token rotation
- [ ] Session binding to device fingerprint
- [ ] Brute-force protection (exponential backoff per IP/user)
- [ ] Rate limiting middleware (per endpoint, per user, per IP)

### 1.3 Authorization & RBAC
- [ ] Role definitions: admin, operator, viewer
- [ ] Middleware to enforce RBAC on every API endpoint
- [ ] Per-host ACL system
- [ ] API key generation with scoped permissions

### 1.4 Credential Vault
- [ ] Client-side encryption protocol design
- [ ] Server-side encrypted blob storage
- [ ] Vault unlock endpoint (master password verification)
- [ ] Credential CRUD with versioning
- [ ] Secure memory handling (zeroing after use)
- [ ] Credential sharing between users (E2E encrypted)

### 1.5 Host Management
- [ ] Host CRUD API with input validation
- [ ] Host grouping with tags and hierarchical paths
- [ ] Host import from CSV, JSON, `.ssh/config`
- [ ] Host health check probes (TCP/ICMP)
- [ ] Host discovery scanner (configurable, rate-limited)

### 1.6 Audit Logging
- [ ] Structured JSON logger with integrity hashes
- [ ] Append-only audit log table with partitioning
- [ ] Event types: auth, session, vault, admin, system
- [ ] Tamper-evident export with HMAC signatures
- [ ] Async ingestion via message queue

### 1.7 API Gateway
- [ ] RESTful API implementation (all endpoints)
- [ ] OpenAPI 3.1 specification + Swagger UI
- [ ] Request/response validation middleware
- [ ] CORS configuration
- [ ] TLS 1.3 enforcement

### 1.8 Deliverables
- All API endpoints functional with Postman/curl tests
- Authentication flow working (password + TOTP)
- Vault encryption/decryption verified
- Audit log recording all events
- 80%+ unit test coverage on backend

---

## Phase 2: Protocol Layer (Weeks 7–10)

### 2.1 Session Management Service
- [ ] Session lifecycle: create, track, terminate
- [ ] Session metadata normalization (protocol-agnostic)
- [ ] Concurrent session limits per user/host
- [ ] Session recording directory structure

### 2.2 SSH Proxy
- [ ] SSH client integration (`golang.org/x/crypto/ssh`)
- [ ] Key-based authentication forwarding
- [ ] Password authentication with vault integration
- [ ] Terminal emulation (xterm-256color)
- [ ] SSH session text recording (asciinema format)
- [ ] SFTP file transfer proxy (optional)

### 2.3 RDP Gateway
- [ ] FreeRDP subprocess wrapper
- [ ] NLA authentication forwarding
- [ ] Bitmap encoding and streaming to web client
- [ ] Input redirection (keyboard, mouse)
- [ ] Session video recording (optional)
- [ ] Clipboard redirection controls

### 2.4 VNC Proxy
- [ ] libvncclient subprocess wrapper
- [ ] VNC authentication methods
- [ ] Framebuffer encoding and streaming
- [ ] Input redirection
- [ ] Session recording (optional)

### 2.5 WebSocket Streaming
- [ ] WebSocket server for terminal sessions
- [ ] Binary message framing for protocol data
- [ ] Heartbeat and timeout handling
- [ ] Connection resilience (reconnect logic)

### 2.6 Sandboxing Implementation
- [ ] Subprocess isolation with seccomp-bpf profiles
- [ ] Network namespace creation per session
- [ ] cgroups v2 resource limits
- [ ] AppArmor/SELinux profile definitions
- [ ] Process monitoring and auto-cleanup

### 2.7 Deliverables
- Working SSH connections through Tower
- Working RDP connections (Windows target)
- Working VNC connections
- Session recording playback
- Sandbox escape testing passed

---

## Phase 3: Web Client (Weeks 11–13)

### 3.1 UI Framework
- [ ] Design system setup (colors, typography, spacing)
- [ ] Dark/light theme implementation
- [ ] Component library (buttons, inputs, tables, modals)
- [ ] Responsive layout (320px+)

### 3.2 Authentication Flow
- [ ] Login screen (email + password)
- [ ] TOTP input screen
- [ ] WebAuthn prompt integration
- [ ] Vault unlock screen
- [ ] Session timeout warning and auto-lock

### 3.3 Host Management UI
- [ ] Host list with search/filter
- [ ] Host detail view
- [ ] Host creation/editing form
- [ ] Group/tag management sidebar
- [ ] Quick-connect button
- [ ] Import dialog (CSV, JSON, ssh config)

### 3.4 Terminal Interface
- [ ] xterm.js integration for SSH sessions
- [ ] Tabbed multi-session interface
- [ ] Session status indicators
- [ ] Keyboard shortcut system (Ctrl+Shift+T, etc.)
- [ ] Resize handling
- [ ] Copy/paste integration

### 3.5 Credential Vault UI
- [ ] Vault unlock prompt
- [ ] Credential list with type icons
- [ ] Credential creation/editing (encrypted before leaving client)
- [ ] SSH key generation in browser (Web Crypto API)
- [ ] Auto-fill for host connections

### 3.6 Admin Dashboard
- [ ] User management table
- [ ] Role assignment interface
- [ ] Audit log viewer with search
- [ ] System health metrics display
- [ ] Active sessions monitor with kill capability

### 3.7 Deliverables
- Functional PWA with all core features
- Lighthouse score > 90
- WCAG 2.1 AA compliance verified

---

## Phase 4: Desktop Clients (Weeks 14–16)

### 4.1 Tauri Shell
- [ ] Tauri project configuration
- [ ] Native menu and system tray integration
- [ ] Window management (minimize to tray)
- [ ] Auto-updater integration

### 4.2 Shared Web Core
- [ ] Reuse web client code in Tauri WebView
- [ ] Native bridge for file system access (import/export)
- [ ] Native bridge for secure credential storage (OS keychain)
- [ ] Native bridge for system notifications

### 4.3 Platform-Specific Polish
- [ ] Windows: Installer (MSI/MSIX), registry integration
- [ ] Linux: AppImage, .deb, .rpm packages
- [ ] macOS: (Future phase) Notary service, Keychain access

### 4.4 Deliverables
- Windows installer with working app
- Linux packages for Ubuntu/Fedora/Arch
- Biometric/OS keychain integration for vault

---

## Phase 5: Mobile Client — Flutter (Weeks 17–19)

### 5.1 Cross-Platform Mobile App
- [ ] Flutter project with Dart
- [ ] Navigation with Material Design 3 / Cupertino
- [ ] Adaptive theming for iOS/Android
- [ ] FFI bindings to `ssh-core` Rust library

### 5.2 Authentication
- [ ] Biometric unlock (fingerprint/face)
- [ ] Secure credential storage (Android Keystore)
- [ ] Background token refresh

### 5.3 Host Management
- [ ] Host list with search
- [ ] Host detail with connection button
- [ ] Sync with Tower (background + manual)
- [ ] Offline caching of host list

### 5.4 Terminal Experience
- [ ] SSH terminal with virtual keyboard optimization
- [ ] Terminal color scheme support
- [ ] Pinch-to-zoom for readability
- [ ] Hardware keyboard support

### 5.5 Credential Vault
- [ ] Vault unlock with biometric
- [ ] Credential viewer (decrypted in memory only)
- [ ] Auto-fill integration (Android Autofill Framework)

### 5.6 Deliverables
- APK signed and installable
- Play Store ready (if desired)
- Biometric auth working

---

## Phase 6: Security Hardening (Weeks 20–22)

### 6.1 Penetration Testing
- [ ] Internal security review with checklist
- [ ] Third-party penetration test engagement
- [ ] OWASP ASVS Level 2 assessment
- [ ] Fuzz testing on API endpoints (`go-fuzz`, `ffuf`)
- [ ] Protocol fuzzing (SSH, RDP, VNC parsers)

### 6.2 Hardening
- [ ] Security headers review (HSTS, CSP, X-Frame-Options)
- [ ] Dependency update and vulnerability remediation
- [ ] Secrets scanning across entire codebase
- [ ] Memory safety audit (Go race detector, valgrind on C components)
- [ ] TLS configuration audit (testssl.sh)

### 6.3 Compliance Documentation
- [ ] ISO 27001 controls mapping
- [ ] SOC 2 Type II controls mapping
- [ ] GDPR data processing documentation
- [ ] Security architecture review document

### 6.4 Incident Response Prep
- [ ] Incident response runbook
- [ ] Log analysis procedures
- [ ] Contact tree and escalation procedures
- [ ] Automated threat detection rules

### 6.5 Deliverables
- Penetration test report with remediation
- Compliance documentation package
- Hardened production build
- Incident response procedures

---

## Phase 7: Release (Weeks 23–24)

### 7.1 Packaging
- [ ] Docker images for x86_64 and ARM64
- [ ] Docker Compose production template
- [ ] Kubernetes Helm chart
- [ ] Desktop installer signing (Windows: EV cert, Linux: GPG)

### 7.2 Documentation
- [ ] User manual (installation, configuration, usage)
- [ ] Administrator guide (scaling, backup, monitoring)
- [ ] API documentation (OpenAPI + examples)
- [ ] Security whitepaper
- [ ] Changelog and release notes

### 7.3 Launch Preparation
- [ ] Beta testing program (invite-only)
- [ ] Bug bounty program setup
- [ ] Support channel establishment (Discord/forum)
- [ ] Website and landing page

### 7.4 Post-Launch
- [ ] Telemetry and error reporting (opt-in, anonymized)
- [ ] Update server for desktop clients
- [ ] Community feedback collection

### 7.5 Deliverables
- Public release v1.0.0
- Signed artifacts
- Documentation site live
- Support channels active

---

## Dependencies & Critical Path

```
Phase 0 ────────────────────────────────────────────
    │
    ▼
Phase 1 ────────────────────────────────────────────
    │         │
    ▼         ▼
Phase 2 ──> Phase 3 (Web) ──> Phase 4 (Desktop)
    │                           │
    ▼                           ▼
Phase 5 (Mobile) ◄─────────────┘
    │
    ▼
Phase 6 ────────────────────────────────────────────
    │
    ▼
Phase 7 ────────────────────────────────────────────
```

**Critical Path:** 0 → 1 → 2 → 3 → 6 → 7 (Desktop and Mobile can parallelize after Web)

---

## Risk Register

| ID | Risk | Probability | Impact | Mitigation |
|----|------|-------------|--------|------------|
| R1 | Protocol library vulnerability (libssh2, FreeRDP) | Medium | Critical | SCA monitoring, rapid patch process, sandboxing |
| R2 | Vault encryption flaw | Low | Critical | External crypto review, use standard libraries only |
| R3 | Sandbox escape | Low | Critical | Regular penetration testing, minimal syscall profiles |
| R4 | Key personnel departure | Medium | High | Documentation, pair programming, no bus factor of 1 |
| R5 | Performance under load | Medium | Medium | Early load testing, horizontal scaling design |
| R6 | Mobile app store rejection | Medium | Low | Follow platform guidelines, early review submission |

---

## Document Control

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1 | 2026-05-27 | Ame (PM Subagent) | Initial task breakdown |
