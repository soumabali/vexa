# vexa — vexa: Project Requirements

> **Classification:** Security-Critical Application  
> **Owner:** Dhar  
> **Status:** Draft — PM Review  
> **Last Updated:** 2026-05-27

---

## 1. Project Overview

**vexa** is a self-hosted, security-first server management platform that provides unified terminal access across multiple protocols (SSH, RDP, VNC). It is designed for system administrators, DevOps engineers, and security professionals who need centralized, audited, and encrypted remote access to heterogeneous infrastructure.

### 1.1 Vision
One secure hub. Every protocol. Every device. Full audit trail.

### 1.2 Target Users
- System Administrators managing Linux/Windows fleets
- DevOps / SRE teams with hybrid infrastructure
- Security-conscious individuals running homelabs or small data centers
- MSPs (Managed Service Providers) requiring client isolation

### 1.3 Core Value Propositions
1. **Unified Access:** Single interface for SSH, RDP, and VNC
2. **Zero-Trust Architecture:** Never trust, always verify, always log
3. **Self-Hosted Sovereignty:** Your data, your keys, your network
4. **Cross-Platform Mobility:** Android, Windows, Linux, Web

---

## 2. Functional Requirements

### 2.1 vexa (Master Server)
- **FR-CT-001:** Self-hosted central hub deployable as Docker container, bare metal binary, or VM image
- **FR-CT-002:** Web-based administrative dashboard for host management, user management, and audit review
- **FR-CT-003:** RESTful API for all operations with OpenAPI documentation
- **FR-CT-004:** WebSocket server for real-time session streaming to web clients
- **FR-CT-005:** Configurable listen addresses, ports, and TLS termination options
- **FR-CT-006:** Health check and readiness endpoints for load balancer integration
- **FR-CT-007:** Backup/restore of configuration and encrypted vault data

### 2.2 Multi-Protocol Support
- **FR-MP-001:** Native SSH client (port 22) with key-based and password authentication
- **FR-MP-002:** RDP gateway supporting NLA and standard RDP security
- **FR-MP-003:** VNC proxy supporting standard VNC authentication
- **FR-MP-004:** Protocol translation layer to normalize session metadata (start time, duration, bytes transferred)
- **FR-MP-005:** Concurrent multi-protocol sessions with per-protocol resource limits
- **FR-MP-006:** Session recording (text for SSH; optional video for RDP/VNC)

### 2.3 Cross-Platform Clients
- **FR-CP-001:** Mobile app (Flutter) for iOS/Android with biometric unlock
- **FR-CP-002:** Desktop client (Tauri) for Windows with system tray integration
- **FR-CP-003:** Desktop client (Tauri) for Linux compatible with major distributions
- **FR-CP-004:** Web app (Next.js) for browser-based access
- **FR-CP-005:** All clients support offline host list caching with encrypted local store
- **FR-CP-006:** Push notification support for session alerts (configurable)

### 2.4 Credential Vault
- **FR-CV-001:** Centralized encrypted storage of credentials (passwords, SSH keys, API tokens)
- **FR-CV-002:** Client-side encryption — server stores only encrypted blobs, cannot decrypt
- **FR-CV-003:** Master password / biometric key derivation for vault unlock
- **FR-CV-004:** Credential versioning and history (last N versions)
- **FR-CV-005:** Auto-fill integration for client applications
- **FR-CV-006:** Secure credential sharing between authorized users (team/role-based)
- **FR-CV-007:** Automatic credential rotation reminders and audit trail

### 2.5 Host Management & Sync
- **FR-HM-001:** CRUD operations for host definitions (name, address, protocol, port, credentials reference)
- **FR-HM-002:** Host grouping with tags and folders
- **FR-HM-003:** Real-time sync of host definitions across all authenticated devices
- **FR-HM-004:** Offline queue for host changes with conflict resolution on reconnect
- **FR-HM-005:** Host reachability health checks (ICMP/TCP probe)
- **FR-HM-006:** Import/export from CSV, JSON, and `.ssh/config`
- **FR-HM-007:** Host discovery via network scanning (opt-in, with rate limiting)

### 2.6 UI/UX Design
- **FR-UX-001:** Clean, minimal interface with dark/light themes
- **FR-UX-002:** Host list with quick-connect and protocol badges
- **FR-UX-003:** Tabbed multi-session interface
- **FR-UX-004:** Search and filter across hosts, sessions, and audit logs
- **FR-UX-005:** Responsive layout down to 320px width
- **FR-UX-006:** Keyboard shortcut system for power users
- **FR-UX-007:** Accessibility compliance (WCAG 2.1 AA minimum)

---

## 3. Security Requirements

> Every functional requirement must satisfy the following security controls.

### 3.1 Authentication & Authorization
- **SR-AUTH-001:** Multi-factor authentication (TOTP, WebAuthn/FIDO2, backup codes)
- **SR-AUTH-002:** Role-Based Access Control (RBAC): Admin, Operator, Viewer
- **SR-AUTH-003:** Per-host access control lists (ACLs)
- **SR-AUTH-004:** Session timeout and idle lock (configurable)
- **SR-AUTH-005:** Brute-force protection with exponential backoff
- **SR-AUTH-006:** Single Sign-On (SSO) via OIDC / SAML 2.0 (optional enterprise)
- **SR-AUTH-007:** API key management with scoped permissions and expiration

### 3.2 Credential Storage & Encryption
- **SR-CRYPTO-001:** AES-256-GCM for data at rest with unique IVs
- **SR-CRYPTO-002:** Argon2id for password/key derivation
- **SR-CRYPTO-003:** Private keys stored in HSM-backed or software-backed secure enclave where available
- **SR-CRYPTO-004:** No plaintext secrets in logs, memory dumps, or swap
- **SR-CRYPTO-005:** Secure zeroing of memory after credential use

### 3.3 Network Security & Tunneling
- **SR-NET-001:** TLS 1.3 mandatory for all client-server communication
- **SR-NET-002:** mTLS option for agent-to-tower connections
- **SR-NET-003:** Built-in SSH tunnel / jump host support
- **SR-NET-004:** WireGuard VPN integration for remote agent connectivity
- **SR-NET-005:** IP allowlist/denylist per user and per host
- **SR-NET-006:** GEO-IP blocking (configurable)

### 3.4 Session Security
- **SR-SESS-001:** Cryptographically random session IDs (256-bit)
- **SR-SESS-002:** Session binding to device fingerprint
- **SR-SESS-003:** Real-time session termination capability by admin or user
- **SR-SESS-004:** Idle session timeout with configurable grace period
- **SR-SESS-005:** Clipboard and file transfer controls (allow/deny per host)
- **SR-SESS-006:** Watermarking of terminal output with user/session ID (optional)

### 3.5 Audit Logging
- **SR-AUDIT-001:** Immutable append-only audit log (WORM storage or blockchain anchoring)
- **SR-AUDIT-002:** Log all authentication attempts (success and failure)
- **SR-AUDIT-003:** Log all session lifecycle events (start, end, protocol, host, user)
- **SR-AUDIT-004:** Log all credential access and vault unlock events
- **SR-AUDIT-005:** Log all administrative actions (user create, role change, host modify)
- **SR-AUDIT-006:** Structured logging (JSON) with integrity verification hashes
- **SR-AUDIT-007:** Exportable audit reports with tamper-evident signatures

### 3.6 Input Validation
- **SR-INPUT-001:** Strict allowlist validation on all API inputs
- **SR-INPUT-002:** Hostname/IP validation with DNS resolution limits
- **SR-INPUT-003:** Rate limiting on all endpoints (per-user, per-IP)
- **SR-INPUT-004:** File upload restrictions (no executables, size limits, MIME type validation)
- **SR-INPUT-005:** SQL/NoSQL injection prevention through parameterized queries

### 3.7 Sandboxing
- **SR-SAND-001:** Protocol handlers run in isolated subprocesses with seccomp/AppArmor/SELinux profiles
- **SR-SAND-002:** Network namespace isolation for protocol proxy processes
- **SR-SAND-003:** Resource limits (CPU, memory, file descriptors) on session containers
- **SR-SAND-004:** Mandatory session cleanup on crash or abnormal termination

### 3.8 Penetration Testing Requirements
- **SR-PENTEST-001:** Annual third-party penetration test with public summary
- **SR-PENTEST-002:** Bug bounty program (minimum critical payout defined)
- **SR-PENTEST-003:** Automated SAST/DAST in CI/CD pipeline
- **SR-PENTEST-004:** Dependency vulnerability scanning (SCA) with fail-fast thresholds
- **SR-PENTEST-005:** Fuzz testing on protocol parsers and API endpoints
- **SR-PENTEST-006:** Red team exercise before every major release

---

## 4. Non-Functional Requirements

### 4.1 Performance
- **NFR-PERF-001:** Tower supports 1000 concurrent sessions on 4 vCPU / 8 GB RAM
- **NFR-PERF-002:** Session establishment latency < 2 seconds under normal load
- **NFR-PERF-003:** Host list sync time < 5 seconds for 1000 hosts

### 4.2 Reliability
- **NFR-REL-001:** 99.9% uptime target (self-hosted; dependent on user infrastructure)
- **NFR-REL-002:** Graceful degradation under load with queue and backpressure
- **NFR-REL-003:** Automatic reconnection with exponential backoff for clients

### 4.3 Scalability
- **NFR-SCALE-001:** Horizontal scaling via multiple Tower instances behind load balancer
- **NFR-SCALE-002:** Stateless session design enabling easy scaling
- **NFR-SCALE-003:** Read replicas for audit log queries

### 4.4 Compliance
- **NFR-COMP-001:** SOC 2 Type II controls mapping (for enterprise deployments)
- **NFR-COMP-002:** GDPR data minimization and right-to-erasure support
- **NFR-COMP-003:** ISO 27001 controls alignment documentation

---

## 5. Open Questions / Decisions Needed

1. **Stack Selection:** Go backend, Next.js frontend, Tauri desktop, Flutter mobile
2. **Database:** PostgreSQL vs SQLite for single-node? How to handle encrypted vault with sync?
3. **Session Recording Storage:** Local filesystem vs S3-compatible object store?
4. **Licensing Model:** Open source core + enterprise plugins vs fully commercial?
5. **On-Prem vs Cloud Bridge:** Should Tower support hybrid mode with relay for NAT traversal?

---

## 6. Document Control

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1 | 2026-05-27 | Ame (PM Subagent) | Initial draft |
