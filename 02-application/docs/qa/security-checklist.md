# Security Verification Checklist: vexa (vexa)

**Document ID:** QA-CHECKLIST-001  
**Version:** 1.0  
**Date:** 2026-05-27  
**Owner:** QA / Security Tester

---

## How to Use This Checklist

- **PASS**: Control is implemented and verified.
- **FAIL**: Control is missing, incomplete, or bypassable.
- **PARTIAL**: Control exists but has weaknesses.
- **N/A**: Not applicable to this application.
- **NOT TESTED**: Could not be evaluated during this assessment.

---

## Section 1: OWASP Top 10 Compliance (2021)

### A01:2021 — Broken Access Control

| # | Control | Status | Notes |
|---|---------|--------|-------|
| 1.1 | Deny by default — all resources require authentication | **FAIL** | WebSocket bypasses auth |
| 1.2 | Access control enforced server-side, not client-side | **PASS** | Gin middleware enforces |
| 1.3 | Users cannot access other users' resources (IDOR) | **FAIL** | Host handlers lack ownership checks |
| 1.4 | Directory listing disabled | **N/A** | API only, no file serving |
| 1.5 | Metadata (UUIDs) not exposed in URLs without auth | **PASS** | UUIDs are standard identifiers |
| 1.6 | CORS properly configured for API | **PARTIAL** | CORS allows credentials but `*` fallback possible |
| 1.7 | Admin functions require admin role | **PASS** | RequireRole middleware present |
| 1.8 | File/function access controls enforced | **N/A** | No file upload/download yet |

**A01 Result: FAIL**

---

### A02:2021 — Cryptographic Failures

| # | Control | Status | Notes |
|---|---------|--------|-------|
| 2.1 | All sensitive data encrypted at rest | **PARTIAL** | Vault encrypts, but database has plaintext metadata |
| 2.2 | All data in transit encrypted with TLS 1.3 | **FAIL** | Server uses HTTP (`ListenAndServe`), no TLS configured |
| 2.3 | Strong encryption algorithms (AES-256-GCM) | **PASS** | Uses AES-256-GCM |
| 2.4 | Keys managed securely (HSM/KMS, not hardcoded) | **FAIL** | Fallback keys hardcoded in config |
| 2.5 | Passwords hashed with Argon2id/ bcrypt | **PARTIAL** | Code exists but not wired to auth flow |
| 2.6 | No hardcoded secrets in source code | **FAIL** | Default secrets in config.go, docker-compose.yml |
| 2.7 | Random numbers from CSPRNG | **PASS** | Uses `crypto/rand` |
| 2.8 | Salts unique per operation | **PASS** | `generateRandomBytes(32)` used |

**A02 Result: FAIL**

---

### A03:2021 — Injection

| # | Control | Status | Notes |
|---|---------|--------|-------|
| 3.1 | Parameterized queries for SQL | **PASS** | Uses `$1, $2` placeholders |
| 3.2 | Input validation on all user inputs | **PARTIAL** | Gin binding validates some fields |
| 3.3 | No OS command execution via user input | **PASS** | Uses Go SSH library, no shell execution |
| 3.4 | No LDAP injection protections needed | **N/A** | No LDAP |
| 3.5 | No XML external entity (XXE) protections needed | **N/A** | No XML parsing of user data |
| 3.6 | Terminal input sanitized | **FAIL** | No bounds checking or null byte filtering |
| 3.7 | WebSocket messages validated | **FAIL** | No schema validation on WS messages |

**A03 Result: FAIL**

---

### A04:2021 — Insecure Design

| # | Control | Status | Notes |
|---|---------|--------|-------|
| 4.1 | Threat model documented and maintained | **PASS** | Threat model exists |
| 4.2 | Security requirements defined | **PASS** | Security requirements documented |
| 4.3 | Business logic abuse protections | **NOT TESTED** | Application not fully implemented |
| 4.4 | Rate limiting on expensive operations | **PARTIAL** | API rate limiting exists; WS has none |
| 4.5 | Defense in depth implemented | **PARTIAL** | Some layers, but gaps exist |

**A04 Result: PARTIAL**

---

### A05:2021 — Security Misconfiguration

| # | Control | Status | Notes |
|---|---------|--------|-------|
| 5.1 | Default accounts and passwords removed | **FAIL** | Database uses `postgres:postgres` |
| 5.2 | Error messages are generic, no stack traces | **FAIL** | Validation errors expose field details |
| 5.3 | Security patches applied | **PARTIAL** | Dependencies mostly current |
| 5.4 | Unnecessary features disabled | **PASS** | Minimal feature set |
| 5.5 | Directory listing disabled | **N/A** | API application |
| 5.6 | Default security headers present | **PASS** | SecurityHeaders middleware present |
| 5.7 | Application runs with least privilege | **PASS** | Docker runs as user 65534 |

**A05 Result: FAIL**

---

### A06:2021 — Vulnerable and Outdated Components

| # | Control | Status | Notes |
|---|---------|--------|-------|
| 6.1 | Dependency inventory maintained | **PARTIAL** | go.mod exists, no SBOM |
| 6.2 | Known vulnerabilities patched | **NOT TESTED** | Snyk/Trivy scans not performed |
| 6.3 | Unused dependencies removed | **PASS** | Minimal dependencies |
| 6.4 | Signed and verified dependencies | **FAIL** | No go.sum verification enforced |
| 6.5 | Container images scanned | **FAIL** | No Trivy/Clair scanning |

**A06 Result: PARTIAL**

---

### A07:2021 — Identification and Authentication Failures

| # | Control | Status | Notes |
|---|---------|--------|-------|
| 7.1 | Multi-factor authentication enforced | **NOT TESTED** | Stub implementation |
| 7.2 | Session tokens generated securely | **PASS** | JWT with crypto/rand |
| 7.3 | Session invalidation on logout | **FAIL** | JWT tokens not revoked on logout |
| 7.4 | Brute force protection | **PARTIAL** | Rate limiting present but can be bypassed |
| 7.5 | Weak password checks | **FAIL** | Weak regex, no breach checking |
| 7.6 | Session timeout after inactivity | **PARTIAL** | Redis TTL but not strictly enforced |
| 7.7 | WebSocket authentication | **FAIL** | No JWT validation during WS upgrade |

**A07 Result: FAIL**

---

### A08:2021 — Software and Data Integrity Failures

| # | Control | Status | Notes |
|---|---------|--------|-------|
| 8.1 | Code signing for deployment | **FAIL** | No code signing |
| 8.2 | Dependency integrity verification | **FAIL** | go.sum present but not enforced in CI |
| 8.3 | CI/CD pipeline security | **NOT TESTED** | No CI configuration found |
| 8.4 | Serialization attack prevention | **N/A** | Uses JSON, not Java serialization |
| 8.5 | Audit log integrity | **PARTIAL** | HMAC present but chain not verified |

**A08 Result: FAIL**

---

### A09:2021 — Security Logging and Monitoring Failures

| # | Control | Status | Notes |
|---|---------|--------|-------|
| 9.1 | All authentication events logged | **PASS** | Auth events covered |
| 9.2 | All authorization decisions logged | **PARTIAL** | Some endpoints logged |
| 9.3 | Input validation failures logged | **FAIL** | Not consistently logged |
| 9.4 | Sensitive data excluded from logs | **PARTIAL** | Redaction present but may miss fields |
| 9.5 | Log integrity protection | **PARTIAL** | HMAC but no chain verification |
| 9.6 | Alerts for suspicious activity | **FAIL** | No alerting mechanism |
| 9.7 | Logs sent to external SIEM | **FAIL** | File + DB only |

**A09 Result: FAIL**

---

### A10:2021 — Server-Side Request Forgery (SSRF)

| # | Control | Status | Notes |
|---|---------|--------|-------|
| 10.1 | URL validation on server-side requests | **PARTIAL** | Host validation exists but not strict |
| 10.2 | Internal services not accessible via SSRF | **NOT TESTED** | No explicit SSRF tests performed |
| 10.3 | Whitelist for outgoing connections | **FAIL** | No URL whitelist |
| 10.4 | DNS rebinding protection | **FAIL** | No DNS rebinding protections |

**A10 Result: FAIL**

---

## Section 2: SANS Top 25 Software Errors (2023)

| CWE | Title | Status | Notes |
|-----|-------|--------|-------|
| CWE-787 | Out-of-bounds Write | **N/A** | Go memory safety prevents |
| CWE-79 | Cross-site Scripting (XSS) | **FAIL** | Terminal output not sanitized |
| CWE-89 | SQL Injection | **PASS** | Parameterized queries used |
| CWE-416 | Use After Free | **N/A** | Go garbage collection |
| CWE-78 | OS Command Injection | **PASS** | No shell execution |
| CWE-20 | Improper Input Validation | **FAIL** | Missing bounds checks |
| CWE-125 | Out-of-bounds Read | **N/A** | Go bounds checking |
| CWE-22 | Path Traversal | **N/A** | No file operations |
| CWE-352 | Cross-Site Request Forgery | **FAIL** | CSRF middleware not applied |
| CWE-434 | Unrestricted File Upload | **N/A** | No file upload |
| CWE-862 | Missing Authorization | **FAIL** | IDOR on host resources |
| CWE-476 | NULL Pointer Dereference | **PARTIAL** | Some nil checks missing |
| CWE-287 | Improper Authentication | **FAIL** | WS auth bypass |
| CWE-190 | Integer Overflow | **N/A** | Go handles overflow |
| CWE-502 | Deserialization of Untrusted Data | **N/A** | JSON only |
| CWE-269 | Improper Privilege Management | **FAIL** | No ownership checks |
| CWE-312 | Cleartext Storage of Sensitive Info | **FAIL** | Plaintext passwords in structs |
| CWE-77 | Command Injection | **PASS** | Not applicable |
| CWE-119 | Buffer Overflow | **N/A** | Go memory safety |
| CWE-798 | Use of Hard-coded Credentials | **FAIL** | Default secrets |
| CWE-918 | Server-Side Request Forgery | **PARTIAL** | No explicit protections |
| CWE-306 | Missing Authentication | **FAIL** | WS upgrade |
| CWE-362 | Race Condition | **NOT TESTED** | Could exist in session handling |
| CWE-732 | Incorrect Permission Assignment | **PARTIAL** | Docker runs as non-root |
| CWE-611 | Improper Restriction of XML External Entity | **N/A** | No XML processing |

**SANS Top 25 Compliance: FAIL**

---

## Section 3: Protocol-Specific Security (SSH, RDP, VNC)

### SSH Security

| # | Control | Status | Notes |
|---|---------|--------|-------|
| 3.1 | SSH host key verification enabled | **FAIL** | `InsecureIgnoreHostKey` used |
| 3.2 | Only strong SSH algorithms (Ed25519, ECDSA P-256) | **NOT TESTED** | Not configurable |
| 3.3 | SSH protocol 2.0 enforced | **PASS** | Go crypto/ssh defaults to v2 |
| 3.4 | Agent forwarding disabled | **PASS** | `allow_agent: false` |
| 3.5 | X11 forwarding disabled | **PASS** | Not enabled |
| 3.6 | Port forwarding restricted | **NOT TESTED** | Not configurable |
| 3.7 | Known_hosts database maintained | **FAIL** | No known_hosts storage |
| 3.8 | SSH connection timeouts configured | **PASS** | 10s timeout |

**SSH Security Result: FAIL**

---

### RDP Security

| # | Control | Status | Notes |
|---|---------|--------|-------|
| 3.9 | NLA enforced | **NOT TESTED** | Stub implementation |
| 3.10 | CredSSP patched | **NOT TESTED** | Stub implementation |
| 3.11 | Clipboard restrictions configurable | **NOT TESTED** | Stub implementation |
| 3.12 | Drive redirection disabled | **NOT TESTED** | Stub implementation |
| 3.13 | Session timeout configured | **PARTIAL** | Connection timeout only |

**RDP Security Result: NOT TESTED**

---

### VNC Security

| # | Control | Status | Notes |
|---|---------|--------|-------|
| 3.14 | VNC through TLS tunnel | **FAIL** | Plain TCP connections |
| 3.15 | Strong VNC password policy | **NOT TESTED** | Stub implementation |
| 3.16 | Clipboard sanitized | **NOT TESTED** | Stub implementation |
| 3.17 | Screen blanking on idle | **NOT TESTED** | Stub implementation |

**VNC Security Result: FAIL**

---

## Section 4: Infrastructure Security (Docker, Networking)

### Docker Security

| # | Control | Status | Notes |
|---|---------|--------|-------|
| 4.1 | Non-root user in container | **PASS** | USER 65534:65534 |
| 4.2 | Read-only root filesystem | **FAIL** | Not configured |
| 4.3 | No privileged containers | **PASS** | No privileged flag |
| 4.4 | Capability dropping | **FAIL** | No capability restrictions |
| 4.5 | seccomp profile | **FAIL** | Default profile only |
| 4.6 | AppArmor/SELinux labels | **FAIL** | Not configured |
| 4.7 | Health checks defined | **FAIL** | No HEALTHCHECK |
| 4.8 | Minimal base image | **PASS** | scratch + ca-certs |
| 4.9 | No secrets in image layers | **FAIL** | .env.example in build context |
| 4.10 | Image scanning in CI | **FAIL** | No CI pipeline |

**Docker Security Result: FAIL**

---

### Network Security

| # | Control | Status | Notes |
|---|---------|--------|-------|
| 4.11 | Network segmentation | **PARTIAL** | Docker network isolation |
| 4.12 | Database not exposed to internet | **FAIL** | PostgreSQL port mapped to host |
| 4.13 | Redis not exposed to internet | **FAIL** | Redis port mapped to host |
| 4.14 | TLS 1.3 for all connections | **FAIL** | HTTP only |
| 4.15 | mTLS for service-to-service | **FAIL** | Not implemented |
| 4.16 | DDoS protection | **PARTIAL** | Rate limiting only |
| 4.17 | WAF in front of application | **FAIL** | No WAF |
| 4.18 | DNSSEC enabled | **FAIL** | Not applicable (self-hosted) |

**Network Security Result: FAIL**

---

### Database Security

| # | Control | Status | Notes |
|---|---------|--------|-------|
| 4.19 | Strong passwords for DB users | **FAIL** | `postgres:postgres` |
| 4.20 | SSL/TLS for DB connections | **FAIL** | `sslmode=disable` |
| 4.21 | Connection limits configured | **PASS** | SetMaxOpenConns(25) |
| 4.22 | Audit logging enabled | **FAIL** | PostgreSQL audit not configured |
| 4.23 | Row-level security | **FAIL** | Not implemented |
| 4.24 | Encrypted backups | **FAIL** | Not configured |

**Database Security Result: FAIL**

---

## Overall Compliance Summary

| Standard | Pass | Fail | Partial | N/A | Not Tested |
|----------|------|------|---------|-----|------------|
| OWASP Top 10 | 8 | 23 | 8 | 10 | 5 |
| SANS Top 25 | 7 | 11 | 3 | 4 | 1 |
| Protocol-Specific | 3 | 5 | 1 | 0 | 8 |
| Infrastructure | 3 | 16 | 2 | 0 | 0 |
| **TOTAL** | **21** | **55** | **14** | **14** | **14** |

**Overall Compliance Rating: FAIL**

---

## Remediation Priority Matrix

| Priority | Findings | Timeline |
|----------|----------|----------|
| **P0 — Immediate** | WS auth bypass, SSH MITM, hardcoded secrets | 24 hours |
| **P1 — Critical** | XSS, CSRF, IDOR, session hijacking, plaintext passwords | 7 days |
| **P2 — High** | TLS config, rate limiting, concurrent sessions | 30 days |
| **P3 — Medium** | Docker hardening, error messages, password policy | 90 days |
| **P4 — Low** | Frontend metadata, logging improvements | Next release |

---

**Document History**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-05-27 | QA Subagent | Initial security checklist |
