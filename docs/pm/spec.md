# vexa — vexa: Technical Specification

> **Classification:** Security-Critical Application  
> **Owner:** Dhar  
> **Status:** Monorepo Consolidated — Ready for Build  
> **Last Updated:** 2026-05-28 (Monorepo Structure Applied)

---

## 1. Architecture Overview

### 1.1 High-Level Diagram (Textual)

```
┌─────────────────────────────────────────────────────────────┐
│                      CLIENT LAYER                            │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌────────────────┐ │
│  │  Mobile  │ │ Desktop  │ │ Desktop  │ │   Web (PWA)    │ │
│  │(Flutter) │ │(Tauri)   │ │(Tauri)   │ │  (Next.js)     │ │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └───────┬────────┘ │
│       │            │            │               │          │
│       └────────────┴────────────┴───────────────┘          │
│                         TLS 1.3 / mTLS                     │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    CENTRAL TOWER (Master)                    │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │   API GW    │  │   WebSocket │  │   Admin Dashboard   │ │
│  │  (REST)     │  │   Server    │  │   (Next.js SPA)     │ │
│  └──────┬──────┘  └──────┬──────┘  └─────────────────────┘ │
│         │                │                                  │
│         └────────────────┼────────────────────────────────┘
│                          │                                  │
│  ┌───────────────────────┴───────────────────────────────┐ │
│  │              CORE SERVICES (Go / Rust)                 │ │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐   │ │
│  │  │   Auth Svc  │ │  Vault Svc  │ │  Host Svc   │   │ │
│  │  │  (RBAC+2FA) │ │(Client Encr)│ │ (CRUD+Sync) │   │ │
│  │  └─────────────┘ └─────────────┘ └─────────────┘   │ │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐   │ │
│  │  │ Session Svc │ │  Audit Svc  │ │  Tunnel Svc │   │ │
│  │  │ (Recording) │ │ (Immutable) │ │ (WireGuard) │   │ │
│  │  └─────────────┘ └─────────────┘ └─────────────┘   │ │
│  └─────────────────────────────────────────────────────┘ │
│                                                          │
│  ┌───────────────────────────────────────────────────────┐│
│  │           PROTOCOL PROXY LAYER (Sandboxed)            ││
│  │  ┌────────────┐ ┌────────────┐ ┌────────────┐        ││
│  │  │ SSH Proxy  │ │ RDP Proxy  │ │ VNC Proxy  │        ││
│  │  │ (libssh2)  │ │ (FreeRDP)  │ │ (libvnc)   │        ││
│  │  └─────┬──────┘ └─────┬──────┘ └─────┬──────┘        ││
│  └────────┼──────────────┼──────────────┼───────────────┘│
└───────────┼──────────────┼──────────────┼────────────────┘
            │              │              │
            ▼              ▼              ▼
        ┌──────┐     ┌────────┐     ┌──────┐
        │ SSH  │     │  RDP   │     │ VNC  │
        │:22   │     │ :3389  │     │ :5900│
        └──────┘     └────────┘     └──────┘
        Target Servers
```

### 1.2 Design Principles
1. **Defense in Depth:** Multiple independent security layers
2. **Zero Trust:** Verify every request, every session, every packet
3. **Least Privilege:** Components run with minimum required permissions
4. **Fail Secure:** Any failure defaults to closed/denied state
5. **Observability:** Everything is logged, measured, and alerted

---

## 2. Technology Stack (Proposed)

### 2.1 Backend — vexa
| Component | Technology | Justification |
|-----------|------------|---------------|
| Primary Language | **Go** | Strong concurrency, memory safety, fast compile, large stdlib |
| Crypto | Go `crypto` + `libsodium` bindings | Battle-tested, constant-time operations |
| HTTP/API | `net/http` + `gorilla/mux` or `chi` | Lightweight, middleware-rich |
| WebSocket | `gorilla/websocket` | Mature, production-proven |
| Database | **PostgreSQL 15+** | ACID, JSON support, row-level security |
| Cache | **Redis** (optional) | Session state, rate limiting |
| Message Queue | **NATS** or **Redis Streams** | Async audit log ingestion |
| Vault Storage | Encrypted blobs in PostgreSQL | Client-held keys = server cannot decrypt |

### 2.2 Protocol Proxies
| Protocol | Library | Sandbox |
|----------|---------|---------|
| SSH | `libssh2` (CGo) or `golang.org/x/crypto/ssh` | seccomp + network namespace |
| RDP | `FreeRDP` (subprocess) | Firejail / AppArmor profile |
| VNC | `libvncclient` (subprocess) | Firejail / AppArmor profile |

### 2.3 Frontend
| Platform | Framework | Notes |
|----------|-----------|-------|
| Web / PWA | **Next.js** 16 + React 19 + TypeScript | Zustand for state |
| Desktop (Win/Mac/Linux) | **Tauri** v2 (Rust + WebView) | Better security than Electron |
| Mobile (iOS/Android) | **Flutter** 3.x (Dart) | Single codebase, native performance |

### 2.4 Infrastructure
- **Container:** Docker + Docker Compose for single-node; Kubernetes for clustered
- **Reverse Proxy:** Traefik or Caddy (auto-TLS)
- **Monitoring:** Prometheus + Grafana
- **Log Aggregation:** Loki or ELK stack
- **Alerting:** Alertmanager + PagerDuty/webhook

---

## 3. Data Architecture

### 3.1 Database Schema (Core Tables)

```sql
-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email CITEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,           -- Argon2id
    totp_secret BYTEA,                     -- Encrypted at rest
    webauthn_credentials JSONB,
    role ENUM('admin','operator','viewer') DEFAULT 'viewer',
    mfa_enabled BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    last_login TIMESTAMPTZ,
    is_active BOOLEAN DEFAULT TRUE
);

-- Devices (for vault sync + session binding)
CREATE TABLE devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    fingerprint BYTEA NOT NULL,            -- SHA-256 of device cert + hardware IDs
    public_key BYTEA NOT NULL,             -- For E2E sync
    last_seen TIMESTAMPTZ,
    trusted BOOLEAN DEFAULT FALSE
);

-- Hosts
CREATE TABLE hosts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    address INET NOT NULL,
    protocol ENUM('ssh','rdp','vnc') NOT NULL,
    port INT CHECK (port > 0 AND port <= 65535),
    credentials_id UUID REFERENCES credentials(id),
    tags TEXT[],
    group_path LTREE,                      -- Hierarchical groups
    allowed_users UUID[],                  -- ACL
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Credentials (encrypted blobs)
CREATE TABLE credentials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    type ENUM('password','ssh_key','api_token'),
    encrypted_data BYTEA NOT NULL,         -- AES-256-GCM, key derived from user master
    iv BYTEA NOT NULL,
    version INT DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Sessions
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    device_id UUID REFERENCES devices(id),
    host_id UUID REFERENCES hosts(id),
    protocol ENUM('ssh','rdp','vnc') NOT NULL,
    started_at TIMESTAMPTZ DEFAULT NOW(),
    ended_at TIMESTAMPTZ,
    bytes_in BIGINT DEFAULT 0,
    bytes_out BIGINT DEFAULT 0,
    recording_path TEXT,                     -- If session recorded
    terminated_by UUID REFERENCES users(id),
    status ENUM('active','closed','killed') DEFAULT 'active'
);

-- Audit Log (append-only, partitioned by month)
CREATE TABLE audit_log (
    id BIGSERIAL,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    event_type TEXT NOT NULL,
    user_id UUID,
    device_id UUID,
    ip_address INET,
    details JSONB,
    integrity_hash TEXT NOT NULL            -- HMAC-SHA256 of row
) PARTITION BY RANGE (timestamp);

-- Indexes
CREATE INDEX idx_audit_time ON audit_log(timestamp);
CREATE INDEX idx_sessions_user ON sessions(user_id, started_at);
CREATE INDEX idx_hosts_owner ON hosts(owner_id, group_path);
```

### 3.2 Encryption Key Hierarchy

```
User Master Password
        │
        ▼
   Argon2id (memory=64MB, iterations=3, parallelism=4)
        │
        ▼
   Master Key (MK) ──────┐
        │                 │
        ▼                 ▼
   Vault Key (VK)    Device Key (DK)
   (per-credential)  (E2E sync)
        │
        ▼
   AES-256-GCM Encryption
   of credential blobs
```

---

## 4. API Specification (REST + WebSocket)

### 4.1 REST API (OpenAPI 3.1)
Base path: `/api/v1`

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/auth/login` | POST | None | Username/password + TOTP |
| `/auth/mfa/challenge` | POST | Partial | WebAuthn challenge |
| `/auth/refresh` | POST | Refresh Token | Rotate access token |
| `/vault/unlock` | POST | Access Token | Decrypt vault with master password |
| `/vault/credentials` | CRUD | Access + Vault | Manage encrypted credentials |
| `/hosts` | CRUD | Access Token | Host management |
| `/hosts/{id}/connect` | POST | Access + Host ACL | Initiate protocol session |
| `/sessions` | GET | Access Token | List active/historical sessions |
| `/sessions/{id}` | DELETE | Access Token | Kill active session |
| `/audit` | GET | Admin | Query audit log |
| `/ws/terminal` | WebSocket | Access Token | Bidirectional terminal stream |
| `/ws/events` | WebSocket | Access Token | Real-time notifications |

### 4.2 Authentication Flow

```
Client                           Tower
  │                                │
  ├─ POST /auth/login ────────────>│
  │  {email, password}             │
  │                                │
  │<────────── 200 + MFA required ─┤
  │  {mfa_token, methods: [totp,   │
  │            webauthn]}          │
  │                                │
  ├─ POST /auth/mfa/verify ───────>│
  │  {mfa_token, totp_code}        │
  │                                │
  │<──────── 200 + {access_token,  │
  │          refresh_token}        │
  │                                │
  ├─ POST /vault/unlock ──────────>│
  │  {master_password}             │
  │                                │
  │<──────── 200 + vault_key_wrap  │
  │                                │
```

### 4.3 WebSocket Protocol
- **Path:** `/ws/terminal?session={session_id}`
- **Subprotocol:** `centraltower.terminal.v1`
- **Message Types:**
  - `INIT` — Client sends terminal dimensions
  - `DATA` — Bidirectional raw protocol bytes (base64 encoded)
  - `RESIZE` — Terminal resize event
  - `HEARTBEAT` — Keepalive every 30s
  - `CLOSE` — Graceful session end

---

## 5. Security Architecture

### 5.1 Threat Model (STRIDE)

| Threat | Component | Mitigation |
|--------|-----------|------------|
| **Spoofing** | Client identity | mTLS + device fingerprinting |
| **Tampering** | Audit logs | HMAC-SHA256 per row + WORM storage |
| **Repudiation** | User actions | Immutable audit log with user signatures |
| **Information Disclosure** | Credential vault | Client-side encryption; server never holds keys |
| **Denial of Service** | Tower availability | Rate limiting, connection limits, backpressure |
| **Elevation of Privilege** | Role escalation | RBAC enforced at every API layer |

### 5.2 Network Security Layers

```
Internet
   │
   ▼
┌─────────────┐  DDoS / WAF (if exposed)
│   CDN/WAF   │
└──────┬──────┘
       │
       ▼
┌─────────────┐  TLS 1.3 termination
│   Traefik   │  + mTLS optional
└──────┬──────┘
       │
       ▼
┌─────────────┐  API Gateway
│    Tower    │  Rate limiting + Auth
│    (Go)     │
└──────┬──────┘
       │
   ┌───┴───┐
   ▼       ▼
┌──────┐ ┌────────┐
│ PSQL │ │ Redis  │
└──────┘ └────────┘
```

### 5.3 Session Sandboxing
- Each protocol proxy runs as a separate OS process
- Processes are launched with:
  - `seccomp-bpf` filter allowing only necessary syscalls
  - `clone()` into new network namespace (no direct internet)
  - `cgroups` v2 for CPU/memory limits
  - `PR_SET_DUMPABLE` disabled to prevent core dumps
  - Landlock or AppArmor for filesystem access restrictions

### 5.4 Secret Lifecycle
1. **Generation:** Cryptographically random via `crypto/rand`
2. **Transmission:** TLS 1.3 or E2E encrypted sync
3. **Storage:** AES-256-GCM encrypted blob in PostgreSQL
4. **Use:** Decrypted in memory only, zeroed after use
5. **Rotation:** Versioned credentials with automatic expiry alerts
6. **Deletion:** Secure wipe + audit log entry

---

## 6. Deployment Architecture

### 6.1 Single-Node (Homelab / Small Team)
```yaml
# docker-compose.yml (conceptual)
services:
  tower:
    image: centraltower:latest
    ports:
      - "443:8443"
    environment:
      - DATABASE_URL=postgres://tower@db/tower
      - TLS_CERT=/certs/cert.pem
    volumes:
      - ./certs:/certs:ro
      - ./recordings:/recordings
  db:
    image: postgres:15-alpine
    volumes:
      - pgdata:/var/lib/postgresql/data
```

### 6.2 High-Availability (Enterprise)
- 3+ Tower instances behind HAProxy/Traefik
- PostgreSQL with Patroni (HA) + connection pooling via PgBouncer
- Redis Sentinel for session state
- Shared object storage (MinIO) for recordings
- Separate read replicas for audit log analytics

---

## 7. Monitoring & Alerting

### 7.1 Metrics (Prometheus)
- `tower_sessions_active` — Gauge by protocol
- `tower_auth_attempts_total` — Counter with result label
- `tower_audit_log_lag_seconds` — Histogram
- `tower_vault_unlock_duration_seconds` — Histogram

### 7.2 Alerts
- **Critical:** Vault decryption failure spike > 5 in 1 minute
- **Critical:** Authentication brute-force detected
- **High:** Session proxy process crash loop
- **Medium:** Audit log ingestion lag > 30 seconds
- **Low:** Certificate expiry < 7 days

---

## 8. Document Control

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1 | 2026-05-27 | Ame (PM Subagent) | Initial architecture spec |
