# Threat Model: Multi-Protocol Terminal Manager

**Document ID:** SEC-TM-001  
**Version:** 1.0  
**Classification:** Internal — Security Critical  
**Owner:** Security Architect  
**Date:** 2026-05-27

---

## 1. Scope & System Overview

This threat model covers a **self-hosted multi-protocol terminal manager** that:
- Stores server credentials (SSH keys, passwords, certificates)
- Provides remote access via SSH, RDP, and VNC
- Is accessible from multiple platforms (web, desktop, mobile)
- Runs as a self-hosted application with a web-based management interface

### 1.1 System Components

| Component | Description | Trust Boundary |
|-----------|-------------|--------------|
| Web Application | Next.js 16 frontend, admin dashboard | Untrusted |
| API Gateway | Reverse proxy, rate limiting, TLS termination | Semi-Trusted |
| Authentication Service | OAuth2/OIDC, MFA, session management | Trusted |
| Credential Vault | Encrypted storage for keys/passwords | Trusted — Critical |
| Session Manager | WebSocket/terminal proxy for live sessions | Trusted — Critical |
| Database | PostgreSQL for metadata, audit logs | Trusted |
| Message Queue | Redis for session state, caching | Semi-Trusted |
| Target Servers | SSH/RDP/VNC endpoints managed by the system | External |

### 1.2 Data Flow Diagram (Level 0)

```
[User Browser] ←→ [API Gateway / WAF] ←→ [Web App + API Server]
                                      ↓
                            [Auth Service (MFA)]
                                      ↓
                        [Credential Vault (HSM-backed)]
                                      ↓
                        [Session Manager (WebSocket Proxy)]
                                      ↓
                        [Target Servers (SSH/RDP/VNC)]
```

---

## 2. STRIDE Threat Analysis

### 2.1 Spoofing (S) — Impersonation Threats

| Threat ID | Threat | Component | Risk | Mitigation |
|-----------|--------|-----------|------|------------|
| S-001 | Attacker spoofs legitimate user via stolen credentials | Auth Service | **Critical** | Enforce MFA (TOTP/WebAuthn), implement device fingerprinting, anomaly detection |
| S-002 | Attacker spoofs admin via session hijacking | Session Manager | **Critical** | Short-lived JWTs (15 min), HTTP-only Secure cookies, binding to IP + device |
| S-003 | Attacker spoofs target server (MITM) | Session Manager | **High** | Strict host key verification, certificate pinning for RDP/VNC |
| S-004 | Attacker spoofs API Gateway via DNS hijacking | API Gateway | **High** | DNSSEC, certificate transparency monitoring, HSTS preload |
| S-005 | Attacker spoofs internal service via compromised pod/container | Infrastructure | **High** | mTLS between services, service mesh (Istio/Linkerd), workload identity |

**Attack Tree: Credential Spoofing**
```
[Goal: Gain unauthorized access as legitimate user]
├── Steal username/password
│   ├── Phishing attack → MFA blocks this path
│   ├── Keylogger on client → Endpoint detection required
│   └── Credential stuffing → Rate limiting + breached password detection
├── Steal MFA token
│   ├── TOTP seed extraction → Secure enclave storage required
│   └── SIM swap (SMS MFA) → **Banned: only TOTP/WebAuthn**
├── Steal session token
│   ├── XSS exploitation → CSP + output encoding blocks this
│   └── Network sniffing (MITM) → TLS 1.3 + certificate pinning
└── Bypass authentication
    ├── OAuth2 implementation flaw → Strict state parameter, PKCE
    └── JWT signature bypass → Strong signing (RS256), key rotation
```

### 2.2 Tampering (T) — Data/State Modification Threats

| Threat ID | Threat | Component | Risk | Mitigation |
|-----------|--------|-----------|------|------------|
| T-001 | Attacker modifies credential vault data at rest | Credential Vault | **Critical** | AES-256-GCM encryption, HMAC-SHA256 integrity checks, sealed encryption |
| T-002 | Attacker modifies session data in transit | Session Manager | **Critical** | TLS 1.3 with AEAD ciphers, certificate pinning |
| T-003 | Attacker modifies audit logs to hide tracks | Database | **High** | Append-only log storage (WORM), cryptographic log signing, remote SIEM forwarding |
| T-004 | Attacker modifies application code/binaries | Web Application | **High** | Code signing, immutable infrastructure, runtime integrity monitoring (Falco) |
| T-005 | Attacker tampers with container images | Infrastructure | **High** | Image signing (Cosign), vulnerability scanning (Trivy), admission controllers |

### 2.3 Repudiation (R) — Denial of Actions

| Threat ID | Threat | Component | Risk | Mitigation |
|-----------|--------|-----------|------|------------|
| R-001 | User denies executing destructive command | Session Manager | **High** | Full session recording (asciinema), signed audit logs, non-repudiation via digital signatures |
| R-002 | Admin denies privilege escalation | Auth Service | **High** | Immutable audit trail, dual-control for sensitive operations, break-glass logging |
| R-003 | Attacker erases evidence of breach | Audit System | **Critical** | Forward logs to external SIEM in real-time, write-once storage, log integrity verification |

### 2.4 Information Disclosure (I) — Data Leakage Threats

| Threat ID | Threat | Component | Risk | Mitigation |
|-----------|--------|-----------|------|------------|
| I-001 | Credential vault breach exposes all secrets | Credential Vault | **Critical** | Envelope encryption with KMS, HSM for master keys, automatic credential rotation, Shamir's Secret Sharing for recovery |
| I-002 | Session recording leak exposes sensitive terminal output | Session Manager | **Critical** | Encrypt recordings at rest, role-based access, automatic PII redaction, retention policies |
| I-003 | Memory dump reveals decrypted credentials | Application Runtime | **Critical** | Secure memory handling (memset_s), mlock for sensitive buffers, avoid swapping, confidential computing (AMD SEV/Intel TDX) |
| I-004 | Error messages leak stack traces / internal paths | Web Application | **Medium** | Generic error responses, detailed logging server-side only |
| I-005 | Timing attacks reveal password validity | Auth Service | **Medium** | Constant-time comparison for passwords, randomized response delays |
| I-006 | Backup exposure reveals full credential database | Infrastructure | **High** | Encrypt backups independently, store in separate account, test restore procedures |

**Attack Tree: Credential Theft**
```
[Goal: Extract credentials from the vault]
├── Direct database breach
│   ├── SQL injection → Parameterized queries, WAF
│   ├── ORM exploitation → Input validation, least privilege DB user
│   └── Backup theft → Encrypted backups, access controls
├── Application layer extraction
│   ├── Memory dump → Secure memory, no core dumps
│   ├── Debug endpoint exposure → No debug in production, authentication
│   └── Deserialization attack → Safe deserialization, input validation
├── Cryptographic bypass
│   ├── Weak encryption key → HSM-backed keys, key rotation
│   ├── Side-channel attack → Constant-time crypto, blinding
│   └── Implementation flaw → Use vetted libraries (libsodium, OpenSSL)
└── Social engineering
    ├── Support staff impersonation → Verification protocols, training
    └── Admin credential phishing → MFA, security awareness
```

### 2.5 Denial of Service (DoS) — Availability Threats

| Threat ID | Threat | Component | Risk | Mitigation |
|-----------|--------|-----------|------|------------|
| D-001 | Resource exhaustion via connection flood | API Gateway | **High** | Rate limiting (per IP, per user), connection pooling, circuit breakers |
| D-002 | Cryptographic DoS (large JWT, expensive hash) | Auth Service | **Medium** | Input size limits, PBKDF2/Argon2 with bounded iterations, proof-of-work for anonymous endpoints |
| D-003 | Session starvation via hung connections | Session Manager | **High** | Connection timeouts (30s idle, 300s max), keep-alive limits, resource quotas per user |
| D-004 | Vault lockout via brute force attempts | Credential Vault | **High** | Exponential backoff, account lockout after 5 attempts, CAPTCHA after 3 failures |
| D-005 | Infrastructure DDoS | API Gateway | **High** | CDN (CloudFlare/Akamai), anycast, upstream DDoS protection, geographic filtering |

### 2.6 Elevation of Privilege (E) — Unauthorized Access Escalation

| Threat ID | Threat | Component | Risk | Mitigation |
|-----------|--------|-----------|------|------------|
| E-001 | Vertical escalation: regular user → admin | Auth Service | **Critical** | Role-based access control (RBAC), attribute-based access control (ABAC), deny-by-default, principle of least privilege |
| E-002 | Horizontal escalation: access another user's credentials | Credential Vault | **Critical** | Row-level security, encrypted per-user envelopes, strict authorization checks on every access |
| E-003 | Container escape to host | Infrastructure | **Critical** | Rootless containers, seccomp/apparmor profiles, no privileged containers, gVisor/Kata Containers |
| E-004 | Server-side request forgery (SSRF) to internal services | Web Application | **High** | URL allowlists, disable HTTP redirects, network policies (deny internal egress by default) |
| E-005 | Command injection via terminal input | Session Manager | **Critical** | Input sanitization, parameterized commands, shell escape prevention, readonly mode options |
| E-006 | Privilege escalation via misconfigured sudo on target | Target Servers | **High** | Harden target servers separately, audit sudoers, principle of least privilege |

**Attack Tree: Session Hijacking**
```
[Goal: Hijack an active terminal session]
├── Steal session identifier
│   ├── XSS injection → CSP, output encoding, HttpOnly cookies
│   ├── Network interception → TLS 1.3, HSTS, secure network
│   └── Session fixation → Regenerate ID on auth, secure random generation
├── Predict session token
│   └── Weak RNG → Cryptographically secure random (CSPRNG)
├── Hijack WebSocket connection
│   ├── CSWSH (Cross-Site WebSocket Hijacking) → Origin validation, CSRF tokens
│   └── Replay attack → Timestamp validation, nonce checks
└── Post-session access
    ├── Stolen recording → Encrypt recordings, access controls
    └── Browser cache exploitation → Cache-Control: no-store, clear on logout
```

---

## 3. Threat Risk Scoring (DREAD)

| Threat ID | Damage | Reproducibility | Exploitability | Affected Users | Discoverability | Score | Priority |
|-----------|--------|-----------------|---------------|----------------|-----------------|-------|----------|
| I-001 (Vault breach) | 10 | 7 | 4 | 10 | 3 | 6.8 | P1 |
| S-001 (Credential spoof) | 9 | 8 | 6 | 10 | 7 | 8.0 | P1 |
| E-001 (Vertical escalation) | 10 | 6 | 5 | 10 | 5 | 7.2 | P1 |
| I-003 (Memory dump) | 10 | 4 | 3 | 10 | 2 | 5.8 | P2 |
| T-001 (Vault tampering) | 9 | 5 | 4 | 10 | 4 | 6.4 | P2 |
| R-003 (Log tampering) | 8 | 6 | 5 | 10 | 5 | 6.8 | P2 |
| E-003 (Container escape) | 10 | 4 | 3 | 10 | 3 | 6.0 | P2 |
| D-001 (Connection flood) | 7 | 9 | 8 | 10 | 8 | 8.4 | P1 |

**Scoring Formula:** `(D + R + E + A + Dc) / 5` where each is 1-10.

---

## 4. Assumptions & Constraints

### 4.1 Assumptions
- The underlying operating system and container runtime are patched and hardened
- Physical access to the host is controlled and monitored
- Administrators are trusted but actions are logged (dual-control for sensitive ops)
- Target servers managed by the system have their own security hardening

### 4.2 Constraints
- Must support self-hosted deployment (no SaaS dependencies for core functions)
- Must work across multiple client platforms (Windows, macOS, Linux, mobile)
- Must comply with SOC 2 Type II and ISO 27001 requirements
- Must support air-gapped / offline deployments

---

## 5. Compliance Mapping

| Framework | Requirements Addressed |
|-----------|------------------------|
| SOC 2 Type II | CC6.1 (logical access), CC6.6 (encryption), CC7.2 (system monitoring) |
| ISO 27001 | A.9 (access control), A.10 (cryptography), A.12 (operations), A.16 (incident) |
| NIST 800-53 | AC-2, AC-3, AU-6, CM-7, IA-2, SC-8, SC-13, SC-28 |
| PCI DSS (if applicable) | Req 3 (protect stored data), Req 4 (encrypt transmission), Req 8 (identify users) |

---

## 6. Review Schedule

| Activity | Frequency | Owner |
|----------|-----------|-------|
| Full threat model review | Quarterly | Security Architect |
| Update after major feature | Per release | Security Team |
| Penetration test validation | Annual | External firm |
| STRIDE table review | Monthly | DevSecOps |

---

**Document History**
| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-05-27 | Security Architect | Initial threat model |
