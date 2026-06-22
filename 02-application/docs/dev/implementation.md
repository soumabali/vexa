# vexa — Developer Implementation Guide

> **Agent:** Full Stack Developer  
> **Status:** Generated 2026-05-28  
> **Version:** 0.1.0

---

## 1. Architecture Overview

### 1.1 Monorepo Layout

```
vexa/
├── apps/
│   ├── api/          # Go 1.25 backend (Gin + PostgreSQL + Redis)
│   ├── web/          # Next.js 16 + React 19 + TypeScript
│   ├── desktop/      # Tauri v2 (Rust + WebView)
│   └── mobile/       # Flutter 3.x (Dart)
├── packages/
│   ├── ssh-core/     # Rust SSH/SFTP/VNC library with FFI
│   ├── ui/           # Shared React UI components (Shadcn/ui)
│   ├── types/        # Shared TypeScript type definitions
│   └── config/       # Shared ESLint, Tailwind, TSConfig presets
├── docs/             # Development team documentation
└── infra/            # Docker, K8s, Terraform
```

### 1.2 Tech Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| **Backend** | Go 1.25 + Gin + GORM | REST API, WebSocket, protocol proxies |
| **Frontend** | Next.js 16 + React 19 + TS | Web dashboard, PWA |
| **Desktop** | Tauri v2 (Rust + WebView) | Cross-platform desktop app |
| **Mobile** | Flutter 3.x (Dart) | iOS/Android app |
| **Core** | Rust + libssh2 | SSH/SFTP/VNC protocol handling |
| **Database** | PostgreSQL 16 | Primary datastore |
| **Cache** | Redis 7 | Sessions, rate limiting, pub/sub |
| **Auth** | JWT + bcrypt + TOTP | Token + MFA authentication |
| **Crypto** | Argon2 + AES-256-GCM | Password hashing, credential vault |

---

## 2. Backend (apps/api)

### 2.1 Module Structure

```
internal/
├── api/
│   ├── router.go          # Gin router + middleware chain
│   ├── validators.go      # Request validation helpers
│   └── handlers/
│       ├── auth.go        # Login, register, MFA, session
│       ├── users.go       # User CRUD, profile
│       ├── hosts.go       # Host management
│       ├── credentials.go # Credential vault
│       ├── gateway.go     # SSH/RDP/VNC proxy
│       ├── terminal.go    # WebSocket terminal
│       └── audit.go       # Audit log queries
├── auth/
│   ├── jwt.go             # JWT token generation/validation
│   ├── mfa.go             # TOTP MFA implementation
│   ├── session.go         # Session management (Redis-backed)
│   └── user_service.go    # User business logic
├── crypto/
│   ├── crypto.go          # AES-256-GCM encryption
│   └── password.go        # Argon2 password hashing
├── db/
│   └── db.go              # PostgreSQL connection + migrations
├── gateway/
│   ├── ssh.go             # SSH proxy with connection pooling
│   ├── rdp.go             # RDP proxy (FreeRDP subprocess)
│   └── vnc.go             # VNC proxy (libvncclient subprocess)
├── hosts/
│   └── repository.go        # Host CRUD operations
├── middleware/
│   ├── auth.go            # JWT auth middleware
│   ├── logging.go         # Structured request logging
│   ├── ratelimit.go       # Redis-based rate limiting
│   ├── recovery.go        # Panic recovery
│   └── security.go        # CORS, CSP, HSTS headers
├── models/
│   ├── user.go            # User entity
│   ├── host.go            # Host entity
│   ├── session.go         # Session entity
│   └── credential.go      # Credential entity
├── sftp/
│   ├── api.go             # SFTP REST API endpoints
│   ├── service.go         # SFTP business logic
│   ├── websocket.go       # WebSocket SFTP streaming
│   ├── bandwidth.go       # Bandwidth throttling
│   └── path.go            # Path sanitization
├── terminal/
│   └── session.go         # WebSocket terminal sessions
├── tunnel/
│   └── manager.go         # WireGuard tunnel management
├── vault/
│   └── vault.go           # Encrypted credential storage
├── audit/
│   └── logger.go          # Immutable audit logging
└── team/
    └── service.go         # Team/organization management
```

### 2.2 Key Patterns

#### Dependency Injection
All services receive dependencies via constructor, not globals:

```go
func NewAuthService(cfg *config.Config, db *sql.DB, redis *redis.Client, vault *vault.Service) *AuthService
```

#### Middleware Chain
```
Recovery → Logging → SecurityHeaders → RateLimit → CORS → Auth → Handler
```

#### Error Handling
Centralized error responses with structured logging:
```go
api.HandleError(c, http.StatusUnauthorized, "AUTH_001", "Invalid credentials")
```

#### Connection Pooling
SSH connections pooled per host with configurable limits:
```go
type SSHPool struct {
    mu          sync.Mutex
    connections []*SSHConnection
    host        string
    port        int
    maxConn     int
    idleTimeout time.Duration
}
```

---

## 3. Frontend (apps/web)

### 3.1 Directory Structure

```
src/
├── app/                   # Next.js App Router pages
│   ├── page.tsx           # Dashboard
│   ├── login/page.tsx
│   ├── register/page.tsx
│   ├── hosts/page.tsx
│   ├── sessions/page.tsx
│   ├── security/page.tsx
│   └── profile/page.tsx
├── components/
│   ├── auth/              # Login, register, MFA forms
│   ├── hosts/             # Host cards, lists
│   ├── file-manager/      # SFTP file browser (tree, grid, list)
│   ├── profile/           # Profile, password change
│   ├── security/          # 2FA, biometric settings
│   ├── sessions/          # Session history, active sessions
│   ├── ui/                # Shadcn/ui components (30+)
│   └── providers/         # QueryProvider, AuthProvider
├── hooks/
│   ├── useAuth.ts         # Auth state management
│   └── use-file-manager.tsx
├── lib/
│   ├── api/               # API client functions
│   ├── validations/       # Zod schemas
│   ├── file-utils.ts
│   └── utils.ts
└── types/
    └── file-manager.ts
```

### 3.2 State Management
- **Server State:** React Query (TanStack Query) for API caching
- **Client State:** Zustand for auth, UI preferences
- **Form State:** React Hook Form + Zod validation

### 3.3 API Client Pattern
```typescript
// lib/api/auth.ts
export async function login(credentials: LoginInput) {
  const res = await fetch('/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(credentials),
  });
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}
```

---

## 4. Desktop (apps/desktop)

### 4.1 Tauri Architecture
- **Backend:** Rust commands (src/commands/*.rs)
- **Frontend:** Reuses apps/web build output
- **Security:** OS keychain integration, sandboxed protocol handlers

### 4.2 Key Commands
```rust
// src/commands/ssh.rs
#[tauri::command]
async fn ssh_connect(host_id: String) -> Result<ConnectionInfo, Error>

// src/commands/keychain.rs
#[tauri::command]
async fn store_credential(key: String, value: String) -> Result<(), Error>
```

---

## 5. Core Library (packages/ssh-core)

### 5.1 Rust FFI Architecture
```
lib.rs          # Public API
├── ssh.rs      # SSH connection + channel management
├── sftp.rs     # SFTP operations
├── tunnel.rs   # WireGuard tunnel
├── vault.rs    # Credential encryption (AES-256-GCM)
├── crypto.rs   # Cryptographic primitives
└── ffi/
    ├── mod.rs   # FFI module exports
    ├── tauri.rs # Tauri bindings
    └── dart.rs  # Flutter FFI bindings
```

### 5.2 FFI Consumers
- **Tauri:** `ssh_core_tauri_` prefix functions
- **Flutter:** `ssh_core_dart_` prefix functions via `dart:ffi`

---

## 6. Database Schema (PostgreSQL)

### 6.1 Core Tables
```sql
users
├── id (UUID, PK)
├── email (UNIQUE)
├── password_hash (Argon2)
├── role (admin|user|viewer)
├── mfa_enabled
├── mfa_secret (encrypted)
├── created_at, updated_at

hosts
├── id (UUID, PK)
├── user_id (FK)
├── name
├── hostname
├── port (default 22)
├── protocol (ssh|rdp|vnc)
├── credential_id (FK, nullable)
└── tags (JSONB)

credentials
├── id (UUID, PK)
├── user_id (FK)
├── type (password|ssh_key|api_token)
├── data (AES-256-GCM encrypted)
└── created_at

sessions
├── id (UUID, PK)
├── user_id (FK)
├── host_id (FK)
├── protocol
├── started_at
├── ended_at
├── bytes_sent
└── bytes_received
```

---

## 7. Security Implementation

### 7.1 Authentication Flow
```
1. User submits credentials
2. Server validates with Argon2 hash
3. If MFA enabled → TOTP challenge
4. Issue JWT access token (15 min) + refresh token (7 days)
5. Store session in Redis
6. Return token pair to client
```

### 7.2 Authorization
- RBAC with 3 roles: admin, user, viewer
- Middleware checks `role` claim in JWT
- Host-level access control per user

### 7.3 Credential Vault
- Master key derived from user password + Argon2
- Credentials encrypted with AES-256-GCM + random nonce
- Key never stored server-side (derived on auth)

### 7.4 Audit Trail
- Every action logged with: user_id, action, resource, timestamp, IP, user_agent
- Logs append-only (no delete/update)
- Exported to SIEM-compatible format

---

## 8. Protocol Gateway

### 8.1 SSH Proxy
- Connection pooling with configurable limits
- WebSocket bridge for browser terminals
- Session recording (text)
- Bandwidth throttling per session

### 8.2 RDP/VNC Proxy
- Spawn subprocess per connection (Firejail sandboxed)
- Video stream via WebRTC or WebSocket binary
- Input forwarding (keyboard/mouse)

### 8.3 SFTP
- WebSocket-based file operations
- Chunked upload/download
- Bandwidth limiting
- Path traversal protection

---

## 9. Performance Considerations

### 9.1 Caching Strategy
| Data | Cache | TTL |
|------|-------|-----|
| User sessions | Redis | 15 min |
| Host lists | Redis | 5 min |
| Rate limits | Redis | Window-based |
| Credentials | Memory (decrypted) | Session lifetime |

### 9.2 Database Optimization
- Connection pooling (max 25)
- Indexed: users.email, hosts.user_id, sessions.user_id
- Partitioned: audit_logs by month

### 9.3 Connection Limits
- SSH: 50 concurrent per host
- RDP/VNC: 10 concurrent per host
- API: Rate limited to 100 req/min per IP
