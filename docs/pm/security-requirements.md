# vexa — vexa: Security Requirements

> **Classification:** Security-Critical Application  
> **Owner:** Dhar  
> **Status:** Draft — Security Review  
> **Last Updated:** 2026-05-27  
> **Confidentiality:** Internal — Contains Threat Model

---

## 1. Executive Summary

vexa is a self-hosted, multi-protocol remote access platform. Because it acts as a **privileged intermediary** between administrators and infrastructure, it represents a **high-value target** for attackers. A compromise of vexa could lead to:

- Unauthorized access to all managed hosts
- Lateral movement across entire networks
- Credential theft and replay attacks
- Complete loss of confidentiality, integrity, and availability

This document defines the security requirements, threat model, and compliance framework necessary to protect the platform and its users.

---

## 2. Security Objectives

| Objective | Description | Priority |
|-----------|-------------|----------|
| SO-1 | **Confidentiality of Credentials** — Server must never have plaintext access to user credentials | Critical |
| SO-2 | **Authentication Integrity** — Only legitimate users with valid MFA can access the system | Critical |
| SO-3 | **Session Isolation** — Sessions must not leak data or privileges to other sessions | Critical |
| SO-4 | **Audit Immutability** — Audit logs must be tamper-evident and tamper-resistant | Critical |
| SO-5 | **Availability Under Attack** — System must degrade gracefully under DoS, not collapse | High |
| SO-6 | **Supply Chain Integrity** — Dependencies must be verified, reproducible builds required | High |
| SO-7 | **Recovery from Compromise** — Clear procedures to rotate keys, revoke access, and recover | High |

---

## 3. Threat Model

### 3.1 Threat Actors

| Actor | Motivation | Capability | Risk Level |
|-------|-----------|------------|------------|
| **External Attacker** | Financial gain, data theft, ransomware | Automated tools, zero-days, social engineering | High |
| **Malicious Insider** | Sabotage, data exfiltration, revenge | Legitimate credentials, knowledge of internals | High |
| **Compromised Client Device** | Pivot to infrastructure | Malware, keyloggers, browser extensions | Medium |
| **Nation-State / APT** | Espionage, critical infrastructure | Advanced persistent threats, supply chain | Medium |
| **Accidental Insider** | Misconfiguration, human error | Limited technical knowledge | Medium |

### 3.2 STRIDE Analysis

#### Spoofing (S)
- **T-SPOOF-01:** Attacker impersonates legitimate user via stolen credentials
  - *Mitigation:* MFA mandatory, WebAuthn preferred, device fingerprinting
- **T-SPOOF-02:** Attacker impersonates vexa server (MITM)
  - *Mitigation:* TLS 1.3 with certificate pinning, mTLS for agents
- **T-SPOOF-03:** Attacker impersonates target host
  - *Mitigation:* Host key verification, known_hosts equivalent storage

#### Tampering (T)
- **T-TAMP-01:** Attacker modifies API requests in transit
  - *Mitigation:* TLS 1.3 + request signing with HMAC
- **T-TAMP-02:** Attacker modifies audit logs to cover tracks
  - *Mitigation:* Immutable append-only storage, integrity hashes, WORM or blockchain anchoring
- **T-TAMP-03:** Attacker modifies host configurations to redirect connections
  - *Mitigation:* Config change notifications, dual-control for critical hosts

#### Repudiation (R)
- **T-REPU-01:** User denies performing an action
  - *Mitigation:* Immutable audit log with user identity, session recordings
- **T-REPU-02:** Admin denies making configuration changes
  - *Mitigation:* Admin action logging with approval workflows

#### Information Disclosure (I)
- **T-DISC-01:** Attacker extracts credentials from server database
  - *Mitigation:* Client-side encryption; server stores only ciphertext
- **T-DISC-02:** Attacker intercepts session traffic
  - *Mitigation:* End-to-end encryption where possible; TLS for all transit
- **T-DISC-03:** Attacker reads session recordings
  - *Mitigation:* Recordings encrypted at rest, access controlled by ACL
- **T-DISC-04:** Memory dump reveals plaintext secrets
  - *Mitigation:* Secure memory handling, swap disabled or encrypted, core dumps disabled

#### Denial of Service (D)
- **T-DOS-01:** Attacker floods authentication endpoints
  - *Mitigation:* Rate limiting, CAPTCHA after N failures, IP-based backoff
- **T-DOS-02:** Attacker opens unlimited sessions
  - *Mitigation:* Per-user and global session limits, connection pooling
- **T-DOS-03:** Attacker exhausts disk with session recordings
  - *Mitigation:* Quotas, automatic rotation, size limits per recording

#### Elevation of Privilege (E)
- **T-ELEV-01:** Low-privilege user gains admin access
  - *Mitigation:* RBAC enforced at every layer, no implicit elevation
- **T-ELEV-02:** Attacker escapes session sandbox to host OS
  - *Mitigation:* seccomp, namespaces, cgroups, minimal syscall profiles
- **T-ELEV-03:** Attacker exploits protocol proxy to gain Tower shell
  - *Mitigation:* Proxies run as unprivileged user, no shell access, chroot

---

## 4. Detailed Security Requirements

### 4.1 Authentication & Authorization (SR-AUTH)

#### SR-AUTH-001: Multi-Factor Authentication
- All user accounts MUST require at least one second factor
- Supported factors: TOTP (RFC 6238), WebAuthn/FIDO2, hardware security keys
- Backup recovery codes MUST be provided and single-use
- Enforced at first login; cannot be disabled by user

#### SR-AUTH-002: Role-Based Access Control
- Predefined roles: **Admin**, **Operator**, **Viewer**
- Custom roles MAY be defined with granular permissions
- Role changes MUST be logged and require admin approval
- Permission checks MUST occur at API gateway AND service layer (defense in depth)

#### SR-AUTH-003: Host-Level Access Control
- Each host MUST have an ACL specifying which users/groups can connect
- Default deny: no user can connect unless explicitly allowed
- Time-based restrictions supported (e.g., business hours only)
- IP-based restrictions per user per host

#### SR-AUTH-004: Session Security
- Sessions MUST expire after configurable idle timeout (default: 15 minutes)
- Sessions MUST have absolute maximum lifetime (default: 8 hours)
- Sessions MUST be bound to device fingerprint
- Concurrent session limits per user (default: 5)
- Real-time session termination capability for admins and users

#### SR-AUTH-005: Brute-Force Protection
- Exponential backoff: 1s, 2s, 4s, 8s... up to 1 hour
- Account lockout after 10 consecutive failures (requires admin unlock or time decay)
- IP-based rate limiting: 10 attempts per minute, 100 per hour
- Distributed rate limiting via Redis for multi-node deployments

#### SR-AUTH-006: API Key Security
- API keys MUST be cryptographically random (256-bit entropy)
- API keys MUST have scoped permissions (read-only, host-specific, etc.)
- API keys MUST have expiration dates (max 1 year)
- API keys MUST be revocable immediately
- API keys MUST NOT be loggable or displayable after creation (show once)

---

### 4.2 Credential Storage & Encryption (SR-CRYPTO)

#### SR-CRYPTO-001: Encryption at Rest
- All sensitive data MUST be encrypted with AES-256-GCM
- Unique IV per encryption operation
- Authenticated encryption preventing tampering

#### SR-CRYPTO-002: Key Derivation
- User master passwords MUST be derived via Argon2id
- Parameters: memory ≥ 64 MB, iterations ≥ 3, parallelism ≥ 4
- Salt MUST be unique per user and stored alongside hash

#### SR-CRYPTO-003: Client-Side Encryption Model
- Server MUST NEVER have access to plaintext credentials
- Encryption/decryption MUST occur on authenticated client device
- Server stores only encrypted blobs and cannot decrypt without master key
- Sync between devices uses E2E encryption with device public keys

#### SR-CRYPTO-004: Secure Memory
- Sensitive data in memory MUST be zeroed after use
- `memset_s` or equivalent platform-specific secure zeroing
- Disable core dumps (`PR_SET_DUMPABLE`)
- Recommend encrypted swap or swap disabled

#### SR-CRYPTO-005: Key Rotation
- Support for credential versioning
- Automatic reminders for key rotation (configurable interval)
- One-click rotation with automatic distribution to authorized clients
- Audit log entry for every rotation event

#### SR-CRYPTO-006: Hardware Security
- Use platform secure enclaves when available:
  - Android: Android Keystore (hardware-backed if available)
  - Windows: DPAPI / TPM
  - Linux: TPM2 + keyring
  - macOS: Keychain (future)
- Web: Web Crypto API for client-side operations

---

### 4.3 Network Security (SR-NET)

#### SR-NET-001: Transport Security
- TLS 1.3 MANDATORY for all client-server communication
- No fallback to TLS 1.2 or lower
- Certificate validation MUST be strict (no insecure skip)

#### SR-NET-002: Mutual TLS
- Optional mTLS for agent-to-tower connections
- Client certificates issued by internal CA
- Certificate revocation checking

#### SR-NET-003: Tunneling & Jump Hosts
- Built-in SSH tunnel support for hosts behind firewalls
- Configurable jump host chains
- WireGuard VPN integration for remote sites

#### SR-NET-004: Network Segmentation
- Tower MUST support deployment in DMZ with restricted outbound
- Protocol proxies MUST have no direct internet access
- Database MUST be on private network only

#### SR-NET-005: IP Controls
- Per-user IP allowlist/denylist
- GEO-IP blocking with configurable regions
- Automatic blocking of TOR exit nodes and known VPNs (configurable)

---

### 4.4 Session Security (SR-SESS)

#### SR-SESS-001: Session Isolation
- Each protocol session MUST run in independent OS process
- Network namespace isolation preventing cross-session communication
- Filesystem isolation via chroot or overlayfs

#### SR-SESS-002: Session Recording
- SSH sessions MUST be recorded as text (asciinema format)
- RDP/VNC sessions MAY be recorded as video (configurable)
- Recordings MUST be encrypted at rest
- Access to recordings MUST require host-level ACL permission

#### SR-SESS-003: Clipboard & File Transfer
- Configurable per-host: allow clipboard, deny clipboard, or one-way only
- File transfer through SFTP/SCP MUST be logged with file names and sizes
- Malware scanning on uploaded files (optional integration)

#### SR-SESS-004: Session Watermarking
- Optional: Embed invisible/visible session ID in terminal output
- Deters screenshot leaks and provides attribution

---

### 4.5 Audit Logging (SR-AUDIT)

#### SR-AUDIT-001: Event Coverage
ALL of the following MUST be logged:
- Authentication attempts (success and failure)
- MFA verification events
- Vault unlock/lock events
- Credential access (view, use, modify, delete)
- Host CRUD operations
- Session lifecycle (start, end, protocol, host, user, bytes)
- Administrative actions (user create, role change, config change)
- System events (startup, shutdown, error, security alert)

#### SR-AUDIT-002: Log Integrity
- Each log entry MUST include HMAC-SHA256 integrity hash
- Hash chain: previous entry hash included in current entry
- Asymmetric signature option for regulatory compliance
- Append-only storage with no update or delete operations permitted

#### SR-AUDIT-003: Retention & Export
- Minimum retention: 1 year (configurable)
- Exportable formats: JSON, CSV, syslog
- Tamper-evident export with manifest and signature
- Real-time streaming to SIEM via syslog/HEC

---

### 4.6 Input Validation & Rate Limiting (SR-INPUT)

#### SR-INPUT-001: API Input Validation
- Strict allowlist validation on all API parameters
- JSON schema validation before processing
- Reject oversized payloads (>1MB default)
- Reject malformed requests at edge

#### SR-INPUT-002: Host Validation
- Hostname validation against RFC 1123
- IP address validation (reject private ranges if configured)
- DNS resolution with timeout and query limit
- No arbitrary command injection via host fields

#### SR-INPUT-003: File Upload Restrictions
- Maximum file size: 10MB default
- MIME type allowlist (text/plain, application/json, image/png, etc.)
- Magic number verification beyond extension
- Quarantine and scan before processing

#### SR-INPUT-004: Injection Prevention
- Parameterized queries for all database access
- No dynamic SQL or command construction
- HTML encoding for all user-supplied data in UI
- CSP headers to mitigate XSS

---

### 4.7 Sandboxing (SR-SAND)

#### SR-SAND-001: Process Isolation
- Protocol proxies MUST spawn as separate processes
- Processes MUST run under dedicated unprivileged user (`tower-proxy`)
- No shell access, no login access
- `NO_NEW_PRIVS` set to prevent privilege escalation

#### SR-SAND-002: seccomp-bpf
- Custom seccomp profiles allowing only necessary syscalls
- Default deny for all non-whitelisted syscalls
- Profile per protocol (SSH needs different syscalls than RDP)

#### SR-SAND-003: Namespaces & cgroups
- Each proxy in new network namespace (no external connectivity)
- PID namespace to prevent process visibility
- cgroups v2 for CPU, memory, and IO limits
- OOM killer protection for Tower core processes

#### SR-SAND-004: Filesystem Restrictions
- Read-only root filesystem for proxies (overlayfs)
- Temporary directories with `noexec`, `nosuid`, `nodev`
- Landlock LSM or AppArmor profile restricting file access

---

## 5. Compliance Requirements

### 5.1 SOC 2 Type II
- CC6.1: Logical access controls (RBAC, MFA)
- CC6.2: Access removal (immediate revocation)
- CC6.3: Access monitoring (audit logs)
- CC7.1: Security operations (monitoring, alerting)
- CC7.2: Incident detection (automated alerts)
- CC8.1: Change management (approval workflows)

### 5.2 GDPR
- Data minimization: only collect necessary data
- Right to erasure: user deletion cascades all personal data
- Data portability: export user data in standard format
- Processing records: maintain records of all data processing

### 5.3 ISO 27001
- A.9.1: Access control policy
- A.9.2: User access management
- A.9.4: System and application access control
- A.12.4: Logging and monitoring
- A.12.6: Technical vulnerability management

---

## 6. Security Testing Requirements

### 6.1 Automated Testing (CI/CD)
- **SAST:** Semgrep, Gosec, ESLint Security — run on every commit
- **SCA:** Snyk, OWASP Dependency-Check — weekly scan minimum
- **DAST:** OWASP ZAP — run against staging environment
- **Container Scan:** Trivy — scan every built image
- **Secret Scan:** Gitleaks — pre-commit hook + CI scan

### 6.2 Manual Testing
- **Annual Penetration Test:** Third-party CREST/OWASP certified firm
- **Red Team Exercise:** Before every major release (x.y.0)
- **Code Review:** Security-focused review for all auth/crypto code

### 6.3 Fuzz Testing
- API endpoints fuzzed with `go-fuzz` / `ffuf`
- Protocol parsers fuzzed with `AFL++` or `libfuzzer`
- Fuzzing corpus maintained and expanded continuously

### 6.4 Bug Bounty
- Public program minimum payouts:
  - Critical: $5,000
  - High: $2,000
  - Medium: $500
  - Low: $100
- Response SLA: 24 hours for critical, 72 hours for high

---

## 7. Incident Response

### 7.1 Severity Classification
| Severity | Criteria | Response Time |
|----------|----------|---------------|
| Critical | Active exploitation, credential breach, full system compromise | 1 hour |
| High | Potential breach, vulnerability with PoC | 4 hours |
| Medium | Vulnerability without known exploitation | 24 hours |
| Low | Defense-in-depth improvement | 7 days |

### 7.2 Response Procedures
1. **Detection:** Automated alerts + manual reporting
2. **Containment:** Isolate affected systems, revoke compromised credentials
3. **Eradication:** Patch vulnerability, rotate keys
4. **Recovery:** Restore services, verify integrity
5. **Lessons Learned:** Post-mortem within 1 week for critical/high

---

## 8. Document Control

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1 | 2026-05-27 | Ame (PM Subagent) | Initial security requirements and threat model |
