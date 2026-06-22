# Security Architecture: Multi-Protocol Terminal Manager

**Document ID:** SEC-ARCH-001  
**Version:** 1.0  
**Classification:** Internal — Security Critical  
**Owner:** Security Architect  
**Date:** 2026-05-27

---

## 1. Security Architecture Principles

### 1.1 Core Principles

| Principle | Implementation |
|-----------|---------------|
| **Zero Trust** | Never trust, always verify — every request authenticated and authorized, regardless of origin |
| **Defense in Depth** | Multiple independent security controls at each layer |
| **Least Privilege** | Minimum access required, time-bound, just-in-time where possible |
| **Assume Breach** | Design for detection and containment, not just prevention |
| **Secure by Default** | Secure configurations out of the box, explicit opt-in for risky features |
| **Fail Securely** | Any failure defaults to denied access, not granted |

### 1.2 Security Layers (Onion Model)

```
┌─────────────────────────────────────────────────────────────┐
│  Layer 6: Data           │ Encrypted credentials, audit logs│
│  Layer 5: Application    │ Input validation, authZ, secure  │
│  Layer 4: Transport      │ TLS 1.3, mTLS, certificate pinning│
│  Layer 3: Network        │ Segmentation, micro-segmentation   │
│  Layer 2: Compute        │ Hardened containers, runtime sec   │
│  Layer 1: Physical       │ Host hardening, encrypted volumes  │
└─────────────────────────────────────────────────────────────┘
```

---

## 2. Zero-Trust Network Architecture

### 2.1 Zero-Trust Tenets Applied

**1. All resources are accessed securely, regardless of location**
- All API calls require mutual TLS (mTLS) between services
- End-user access requires TLS 1.3 + strong authentication
- No "trusted network" exceptions for internal services

**2. Strict access control on a per-request basis**
- Every API request carries JWT with identity and claims
- Service-to-service calls verified via SPIFFE/SPIRE or Istio mTLS
- Resource access evaluated against ABAC policies

**3. Assume breach — inspect and log all traffic**
- All traffic inspected by service mesh (Istio/Linkerd)
- Anomaly detection on traffic patterns
- East-west traffic monitored, not just north-south

### 2.2 Trust Boundaries

```
┌──────────────────────────────────────────────────────────────────────┐
│                        UNTRUSTED ZONE                                │
│  [Internet] → [CDN/WAF] → [API Gateway]                              │
│                                                                      │
│   Trust Boundary: TLS 1.3 termination + WAF rules                    │
└──────────────────────────────────────────────────────────────────────┘
                              ↓
┌──────────────────────────────────────────────────────────────────────┐
│                     SEMI-TRUSTED ZONE                                │
│  [Load Balancer] → [Web App] → [API Server]                          │
│                                                                      │
│   Trust Boundary: mTLS + JWT validation + rate limiting              │
└──────────────────────────────────────────────────────────────────────┘
                              ↓
┌──────────────────────────────────────────────────────────────────────┐
│                        TRUSTED ZONE                                    │
│  [Auth Service] → [Credential Vault] → [Session Manager]             │
│  [Database] → [Message Queue]                                        │
│                                                                      │
│   Trust Boundary: mTLS + service identity + network policies           │
└──────────────────────────────────────────────────────────────────────┘
                              ↓
┌──────────────────────────────────────────────────────────────────────┐
│                     HIGHLY TRUSTED ZONE                               │
│  [HSM/KMS] → [Master Key Storage] → [Audit Log Storage (Immutable)] │
│                                                                      │
│   Trust Boundary: Hardware-backed, physical access controls           │
└──────────────────────────────────────────────────────────────────────┘
```

### 2.3 Service-to-Service Authentication

```yaml
# Istio PeerAuthentication example
apiVersion: security.istio.io/v1beta1
kind: PeerAuthentication
metadata:
  name: default
  namespace: terminal-manager
spec:
  mtls:
    mode: STRICT  # All traffic must use mTLS
---
# AuthorizationPolicy example
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: vault-access
  namespace: terminal-manager
spec:
  selector:
    matchLabels:
      app: credential-vault
  action: ALLOW
  rules:
  - from:
    - source:
        principals: ["cluster.local/ns/terminal-manager/sa/api-server"]
    to:
    - operation:
        methods: ["POST"]
        paths: ["/vault/credentials/*"]
```

---

## 3. Defense in Depth Strategy

### 3.1 Layered Defenses by Component

| Layer | Control | Implementation |
|-------|---------|---------------|
| **Perimeter** | DDoS Protection | CloudFlare/Akamai Magic Transit, anycast |
| **Perimeter** | WAF | OWASP CRS, rate limiting, geo-blocking |
| **Perimeter** | Bot Detection | JavaScript challenges, CAPTCHA integration |
| **Network** | Segmentation | VLANs/VPCs, network policies, micro-segmentation |
| **Network** | Traffic Inspection | Istio/Linkerd with telemetry |
| **Application** | Authentication | MFA (TOTP/WebAuthn), SSO integration |
| **Application** | Authorization | RBAC + ABAC, deny-by-default |
| **Application** | Input Validation | Schema validation, allowlists, parameterized queries |
| **Data** | Encryption at Rest | AES-256-GCM with envelope encryption |
| **Data** | Encryption in Transit | TLS 1.3, mTLS, perfect forward secrecy |
| **Data** | Encryption in Use | AMD SEV/Intel TDX for sensitive workloads |
| **Monitoring** | Logging | Immutable audit trail, real-time SIEM forwarding |
| **Monitoring** | Alerting | Anomaly detection, threshold-based alerts |
| **Recovery** | Backup | Encrypted, tested, geographically distributed |

### 3.2 Defense in Depth: Credential Access Example

```
User requests credential for Server-A:
┌────────────────────────────────────────────────────────────┐
│ 1. User authenticates with MFA (TOTP/WebAuthn)             │
├────────────────────────────────────────────────────────────┤
│ 2. JWT validated (signature, expiry, issuer)               │
├────────────────────────────────────────────────────────────┤
│ 3. RBAC check: User has "read:credentials:server-a" role?    │
├────────────────────────────────────────────────────────────┤
│ 4. ABAC check: Is request during allowed hours? From        │
│    allowed IP range? Device trusted?                         │
├────────────────────────────────────────────────────────────┤
│ 5. Rate limit: Not exceeding credential access threshold    │
├────────────────────────────────────────────────────────────┤
│ 6. Audit log: Request logged with full context             │
├────────────────────────────────────────────────────────────┤
│ 7. Credential decrypted using envelope encryption           │
│    (Data key → KMS → HSM-backed master key)                  │
├────────────────────────────────────────────────────────────┤
│ 8. Credential delivered over TLS 1.3, single-use if possible│
├────────────────────────────────────────────────────────────┤
│ 9. Session recording begins if credential used for connection│
└────────────────────────────────────────────────────────────┘
```

---

## 4. Encryption Strategy

### 4.1 Encryption at Rest

#### Credential Vault Encryption (Envelope Encryption)

```
┌─────────────────────────────────────────────────────────────┐
│                    ENVELOPE ENCRYPTION                         │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│   ┌──────────────┐         ┌──────────────┐                 │
│   │  Credential  │ ──enc──▶│ Encrypted    │                 │
│   │  (plaintext) │   +     │ Credential   │                 │
│   └──────────────┘  DEK    └──────────────┘                 │
│         ↑                                                     │
│   ┌──────────────┐         ┌──────────────┐                 │
│   │  Data Key    │ ──enc──▶│ Encrypted    │                 │
│   │  (DEK)       │   +     │ DEK          │                 │
│   │  (per-cred)  │  KEK     │              │                 │
│   └──────────────┘         └──────────────┘                 │
│         ↑                        ↓                            │
│   ┌──────────────┐         ┌──────────────┐                 │
│   │  KMS/Hashi   │ ──enc──▶│ Master Key   │                 │
│   │  Vault/AWS   │   +     │ (HSM-backed) │                 │
│   │  KMS         │  Master  │              │                 │
│   └──────────────┘         └──────────────┘                 │
│                                                               │
│   Key Rotation:                                               │
│   - DEK: Per-credential, rotated on each update              │
│   - KEK: Monthly rotation, re-encrypt all DEKs               │
│   - Master: Quarterly rotation via HSM ceremony              │
└─────────────────────────────────────────────────────────────┘
```

**Algorithm Selection:**
| Data Type | Algorithm | Key Size | Mode |
|-----------|-----------|----------|------|
| Credentials | AES-256 | 256-bit | GCM (authenticated) |
| Session Keys | ChaCha20-Poly1305 | 256-bit | AEAD |
| Password Hashing | Argon2id | - | Memory=64MB, iterations=3, parallelism=4 |
| Digital Signatures | Ed25519 | 256-bit | Pure EdDSA |
| Key Derivation | HKDF-SHA256 | Variable | Extract-then-Expand |

### 4.2 Encryption in Transit

#### TLS Configuration

```nginx
# Nginx TLS configuration (API Gateway)
server {
    listen 443 ssl http2;
    
    # TLS 1.3 only (drop 1.2 for internal, allow 1.2 for compatibility if needed)
    ssl_protocols TLSv1.3;
    
    # Strong cipher suites
    ssl_ciphers TLS_AES_256_GCM_SHA384:TLS_CHACHA20_POLY1305_SHA256;
    ssl_prefer_server_ciphers off;  # Let client choose (both are strong)
    
    # Perfect Forward Secrecy (mandatory in TLS 1.3)
    # Certificates
    ssl_certificate /etc/ssl/certs/server.crt;
    ssl_certificate_key /etc/ssl/private/server.key;
    
    # OCSP Stapling
    ssl_stapling on;
    ssl_stapling_verify on;
    
    # HSTS
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;
    
    # Certificate pinning (optional, report-only first)
    add_header Expect-CT "max-age=86400, enforce" always;
}
```

#### Service-to-Service mTLS

```yaml
# Every internal service must present valid certificate
# Certificates issued by internal CA (Vault/Istio CA)
# Short-lived (24-72 hours), auto-rotated

Certificate Requirements:
- CN = service identity (e.g., "credential-vault.terminal-manager.svc")
- SAN = DNS names, service account
- Key type: ECDSA P-256
- Signature: SHA-384
- Validity: 24 hours (auto-rotation)
- Issued by: Internal CA with intermediate protection
```

### 4.3 Encryption in Use (Memory/Runtime)

```
┌─────────────────────────────────────────────────────────────┐
│              CONFIDENTIAL COMPUTING (where available)          │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  AMD SEV / Intel TDX / ARM CCA                               │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │  Encrypted Memory Region                                  │ │
│  │  ┌─────────────────────────────────────────────────────┐│ │
│  │  │  Credential Vault Process                            ││ │
│  │  │  • Decrypted credentials only in encrypted memory     ││ │
│  │  │  • Host OS cannot read secrets                       ││ │
│  │  │  • Memory encrypted with VM-specific key             ││ │
│  │  └─────────────────────────────────────────────────────┘│ │
│  └─────────────────────────────────────────────────────────┘ │
│                                                               │
│  Fallback (when confidential computing unavailable):          │
│  • mlock() sensitive memory pages (prevent swapping)          │
│  • Disable core dumps (ulimit -c 0)                          │
│  • Clear memory immediately after use (explicit_bzero)      │
│  • Run in isolated containers with restricted capabilities    │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

---

## 5. Authentication Flow with MFA Support

### 5.1 Authentication Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                        AUTHENTICATION FLOW                           │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  [User] ──(1)──▶ [Client App]                                      │
│                    │                                                 │
│                    ▼ (2)                                             │
│              [API Gateway]                                           │
│                    │                                                 │
│                    ▼ (3) Authenticate                                │
│              [Auth Service]                                          │
│                    │                                                 │
│              ┌─────┴─────┐                                           │
│              ▼           ▼                                           │
│        [Identity    [MFA                                            │
│         Provider]    Service]                                        │
│         (OIDC)       (TOTP/WebAuthn)                                 │
│              │           │                                           │
│              └─────┬─────┘                                           │
│                    ▼ (4) Success                                     │
│              [Token Service]                                         │
│                    │                                                 │
│              ┌─────┴─────┐                                           │
│              ▼           ▼                                           │
│        [Access Token] [Refresh Token]                                │
│        (JWT, 15 min)  (Rotating, 7 days)                             │
│              │           │                                           │
│              └─────┬─────┘                                           │
│                    ▼                                                 │
│              [Session Manager]                                       │
│                    │                                                 │
│                    ▼ (5) API Access                                  │
│              [Credential Vault]                                      │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### 5.2 Multi-Factor Authentication

#### Supported Methods (Priority Order)

| Priority | Method | Security Level | User Experience | Implementation |
|----------|--------|---------------|-----------------|----------------|
| 1 | WebAuthn/FIDO2 | **Highest** | Excellent (touch/face) | Hardware keys, platform authenticators |
| 2 | TOTP (RFC 6238) | High | Good (authenticator app) | Time-based 6-digit codes |
| 3 | Push Notification | Medium-High | Good (tap to approve) | Mobile app integration |
| **X** | SMS OTP | **Banned** | - | Vulnerable to SIM swap, interception |

#### MFA Flow

```
┌─────────────────────────────────────────────────────────────┐
│                    MFA ENFORCEMENT FLOW                      │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  1. Primary Authentication                                   │
│     ├── Username + Password (Argon2id verification)         │
│     ├── OR SSO (OIDC/SAML via IdP)                          │
│     └── Password policy: 16+ chars, breached password check   │
│                                                              │
│  2. MFA Challenge (if primary succeeds)                       │
│     ├── Check: User has MFA enrolled?                        │
│     │   ├── No → Redirect to MFA enrollment (mandatory)     │
│     │   └── Yes → Issue MFA challenge                       │
│     │                                                       │
│     ├── WebAuthn: Return challenge, client signs with key    │
│     ├── TOTP: Request 6-digit code, validate window        │
│     └── Hardware token: Request signature, validate           │
│                                                              │
│  3. Step-Up Authentication (for sensitive operations)       │
│     ├── Accessing credential vault → Re-authenticate + MFA │
│     ├── Deleting credentials → Require second approver       │
│     └── Privilege escalation → Hardware token required       │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

#### WebAuthn Implementation

```javascript
// Client-side registration flow
const credential = await navigator.credentials.create({
  publicKey: {
    challenge: Uint8Array.from(atob(serverChallenge), c => c.charCodeAt(0)),
    rp: { name: "Terminal Manager", id: "terminal.example.com" },
    user: { id: userId, name: "user@example.com", displayName: "User" },
    pubKeyCredParams: [{ alg: -7, type: "public-key" }], // ES256
    authenticatorSelection: {
      authenticatorAttachment: "platform", // or "cross-platform"
      userVerification: "required",
      residentKey: "required" // Passkey support
    },
    attestation: "direct"
  }
});
```

---

## 6. Session Management and Timeouts

### 6.1 Session Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                      SESSION LIFECYCLE                               │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │  BROWSER SESSION                                               │ │
│  │  ├── Cookie: session_id (HttpOnly, Secure, SameSite=Strict)  │ │
│  │  ├── Max-Age: 8 hours                                        │ │
│  │  ├── Absolute timeout: 8 hours (force re-auth)                │ │
│  │  ├── Idle timeout: 30 minutes (warn at 25 min)               │ │
│  │  └── Revocation: Instant via session blacklist (Redis)         │ │
│  └────────────────────────────────────────────────────────────────┘ │
│                                                                      │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │  API ACCESS TOKEN (JWT)                                        │ │
│  │  ├── Lifetime: 15 minutes                                       │ │
│  │  ├── Refresh: Via refresh token (rotating, 7 days)            │ │
│  │  ├── Binding: Issued for specific device + IP                  │ │
│  │  └── Claims: user_id, roles, mfa_verified, device_fingerprint │ │
│  └────────────────────────────────────────────────────────────────┘ │
│                                                                      │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │  TERMINAL SESSION (WebSocket)                                  │ │
│  │  ├── Lifetime: Max 4 hours                                      │ │
│  │  ├── Idle timeout: 15 minutes (auto-disconnect)                 │ │
│  │  ├── Recording: All sessions recorded by default                │ │
│  │  └── Concurrent: Max 3 per user (configurable)                 │ │
│  └────────────────────────────────────────────────────────────────┘ │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### 6.2 Timeout Configuration

| Timeout Type | Duration | Action | Configuration |
|-------------|----------|--------|--------------|
| Browser session absolute | 8 hours | Force re-authentication | Fixed, non-configurable |
| Browser session idle | 30 minutes | Warn at 25m, logout at 30m | Admin-configurable (15-120 min) |
| API access token | 15 minutes | Refresh required | Fixed for security |
| API refresh token | 7 days | Re-authentication required | Admin-configurable (1-30 days) |
| Terminal session max | 4 hours | Force disconnect | Admin-configurable (1-24 hours) |
| Terminal session idle | 15 minutes | Auto-disconnect | Admin-configurable (5-60 minutes) |
| MFA remembered device | 30 days | Require MFA again | User-configurable (7-90 days) |

### 6.3 Session Security Controls

```yaml
# Session security implementation
SessionSecurity:
  # Prevent session fixation
  regenerateOnAuth: true
  
  # Bind session to network context
  binding:
    ipSubnet: true      # Match /24 subnet (allow minor network changes)
    userAgent: true     # Detect major UA changes
    deviceFingerprint: true  # Canvas + font fingerprinting
  
  # Concurrent session limits
  maxConcurrent:
    web: 3              # Browser sessions
    terminal: 3         # Active terminal connections
    api: 10             # API clients
  
  # Revocation
  revocation:
    blacklistStore: redis
    blacklistTTL: 8h    # Match max session duration
    immediatePropagation: true  # All nodes aware within 1 second
  
  # Secure cookie settings
  cookie:
    name: __Host-session  # __Host- prefix for additional security
    httpOnly: true
    secure: true
    sameSite: Strict
    path: /
    partition: true     # CHIPS (Cookies Having Independent Partitioned State)
```

---

## 7. Audit and Logging Architecture

### 7.1 Logging Requirements

#### What MUST Be Logged

| Event | Data Captured | Sensitivity |
|-------|--------------|-------------|
| Authentication success | User ID, timestamp, IP, device, MFA method | Medium |
| Authentication failure | User ID, timestamp, IP, failure reason | Medium |
| Privilege escalation | User ID, target role, approver, timestamp | High |
| Credential access | User ID, credential ID, action (read/rotate), timestamp | High |
| Terminal session start | User ID, target server, protocol, timestamp | High |
| Terminal session end | User ID, duration, bytes transferred | Medium |
| Configuration change | Admin ID, setting changed, old/new value | High |
| Vault key rotation | Key ID, triggered by, timestamp | Critical |
| MFA enrollment/deletion | User ID, method, timestamp | High |
| Bulk operations | User ID, operation type, count affected | High |

#### What MUST NOT Be Logged

| Data Type | Reason | Alternative |
|-----------|--------|-------------|
| Passwords / credentials | Exposure risk | Log "credential accessed" only |
| MFA secrets / TOTP seeds | Complete account takeover | Log "MFA verified" only |
| Decryption keys | Cryptographic compromise | Log "decryption performed" only |
| Session tokens / JWTs | Session hijacking | Log token hash (first 8 chars) |
| Full credit card numbers | PCI violation | Last 4 digits only |
| Personal health information | HIPAA/GDPR violation | Anonymized identifiers |

### 7.2 Logging Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                      AUDIT LOGGING PIPELINE                          │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐     │
│  │ Application │──▶│ Log Shipper│──▶│ Kafka/   │──▶│ External │     │
│  │ (structured │    │ (Fluentd/ │    │ Redis    │    │ SIEM     │     │
│  │  JSON)      │    │  Vector)  │    │ Stream   │    │ (Splunk/ │     │
│  └──────────┘    └──────────┘    └──────────┘    │  Datadog)│     │
│        │                              │            └──────────┘     │
│        │                              │                             │
│        ▼                              ▼                             │
│  ┌──────────┐                 ┌──────────┐                         │
│  │ Local    │                 │ Immutable│                         │
│  │ PostgreSQL│                │ Storage  │                         │
│  │ (7 days) │                 │ (WORM)   │                         │
│  └──────────┘                 └──────────┘                         │
│                                                                      │
│  Log Integrity:                                                      │
│  • Each log entry signed with Ed25519                                │
│  • Merkle tree for batch verification                                │
│  • Tamper detection via periodic hash chain verification             │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### 7.3 Real-Time Alerting Rules

```yaml
alerting_rules:
  # Authentication anomalies
  - name: Multiple Failed Logins
    condition: failed_logins > 5 in 5 minutes per user
    severity: high
    action: rate_limit + notify_admin
    
  - name: Impossible Travel
    condition: login_from_new_country within 1 hour of different_country
    severity: critical
    action: force_mfa + notify_admin
    
  # Credential access anomalies
  - name: Bulk Credential Access
    condition: credential_access_count > 50 in 10 minutes per user
    severity: high
    action: alert_admin + require_reauth
    
  - name: After Hours Vault Access
    condition: credential_access AND time NOT IN business_hours
    severity: medium
    action: notify_admin + log_only
    
  # Infrastructure
  - name: Container Escape Attempt
    condition: falco_alert_type == "syscall_anomaly"
    severity: critical
    action: isolate_container + page_oncall
    
  - name: Unusual Egress Traffic
    condition: outbound_bytes > 10x_baseline
    severity: high
    action: trigger_captures + notify_soc
```

---

## 8. Secrets Management Architecture

### 8.1 Secret Tiers

```
┌─────────────────────────────────────────────────────────────────────┐
│                      SECRET TIER HIERARCHY                           │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  Tier 1: Master Keys (HSM-backed)                                    │
│  ├── Key Encryption Keys (KEKs)                                      │
│  ├── Root CA private key                                             │
│  └── HSM admin credentials                                           │
│  Storage: Hardware Security Module (YubiHSM, AWS CloudHSM)          │
│  Access: M-of-N Shamir split, physical ceremony required            │
│                                                                      │
│  Tier 2: Service Credentials                                         │
│  ├── Database credentials                                            │
│  ├── Message queue credentials                                       │
│  ├── External API keys                                               │
│  └── Internal service tokens                                         │
│  Storage: HashiCorp Vault / AWS Secrets Manager                      │
│  Access: Service identity (SPIFFE), dynamic short-lived credentials  │
│                                                                      │
│  Tier 3: Application Secrets                                         │
│  ├── JWT signing keys                                                │
│  ├── Session encryption keys                                         │
│  └── API rate limit salts                                            │
│  Storage: Injected at runtime, never committed                       │
│  Access: Application runtime only, memory-resident                   │
│                                                                      │
│  Tier 4: User-Provided Credentials                                   │
│  ├── SSH keys for target servers                                     │
│  ├── RDP passwords                                                   │
│  ├── VNC passwords                                                   │
│  └── Server certificates                                             │
│  Storage: Credential Vault (envelope encrypted)                      │
│  Access: Authorized users via RBAC + ABAC                            │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### 8.2 Secret Rotation Schedule

| Secret Type | Rotation Frequency | Automated | Procedure |
|-------------|-------------------|-----------|-----------|
| Master Key (KEK) | Quarterly | Partial | HSM ceremony, M-of-N required |
| Data Encryption Keys | On credential update | Yes | Auto-generate new DEK, re-encrypt |
| Service passwords | Monthly | Yes | Vault dynamic credentials |
| JWT signing keys | Weekly | Yes | Blue/green rotation with overlap |
| TLS certificates | 60 days before expiry | Yes | cert-manager + Let's Encrypt or internal CA |
| User SSH keys | Every 90 days or on suspicion | User-triggered | Email notification + one-click rotation |
| API credentials | On employee offboarding | Yes | Immediate revocation + audit |

---

## 9. Security Monitoring and Incident Response

### 9.1 Detection Capabilities

| Control | Implementation | Coverage |
|---------|---------------|----------|
| Runtime threat detection | Falco/Tetragon | Container syscalls, privilege escalation |
| Network anomaly detection | Istio telemetry + ML | East-west traffic patterns |
| Log anomaly detection | Splunk/Datadog ML | Authentication patterns, access anomalies |
| Credential usage monitoring | Custom rules | Unusual credential access patterns |
| Session behavior analysis | Custom + eBPF | Terminal command patterns, data exfiltration |

### 9.2 Incident Response Integration

```
Detection → Triage → Containment → Eradication → Recovery → Lessons
    │         │           │            │           │          │
    ▼         ▼           ▼            ▼           ▼          ▼
  SIEM     On-call    Auto:        Forensic    Restore    Post-
  Alert    Engineer   Isolate      Analysis    from       mortem
                      Affected     (evidence   backup     review
                      Containers   preservation)          + update
                      + Revoke
                        tokens
```

---

**Document History**
| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-05-27 | Security Architect | Initial security architecture |
