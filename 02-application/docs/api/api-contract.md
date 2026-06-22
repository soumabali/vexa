# vexa — API Contract Documentation

> **Agent:** Full Stack Developer
> **Status:** Generated 2026-05-28
> **Version:** 0.1.0
> **OpenAPI Spec:** `docs/api/openapi.yaml`

---

## 1. API Overview

| Field | Value |
|-------|-------|
| **Base URL (Dev)** | `http://localhost:8080/api/v1` |
| **Base URL (Prod)** | `https://api.vexa.local/api/v1` |
| **Authentication** | JWT Bearer Token + Optional MFA |
| **Content-Type** | `application/json` |
| **WebSocket** | `ws://localhost:8080/ws/terminal` |
| **OpenAPI Version** | 3.0.3 |

---

## 2. Authentication Flow

### 2.1 Login Flow (with MFA)

```
┌─────────┐                    ┌─────────┐
│  Client │─── POST /auth/login ───>│  Server │
│         │◄─── MFA Required ───│         │
│         │                     │         │
│         │─── POST /auth/mfa/verify ──>│         │
│         │◄─── Token Pair ───│         │
└─────────┘                     └─────────┘
```

### 2.2 Token Refresh

```
┌─────────┐                          ┌─────────┐
│  Client │─── POST /auth/refresh ───>│  Server │
│         │  (with refresh_token)   │         │
│         │◄─── New Token Pair ─────│         │
└─────────┘                          └─────────┘
```

### 2.3 Authorization Header

```http
Authorization: Bearer eyJhbG…9…
```

---

## 3. Endpoint Summary

### 3.1 Authentication

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/auth/login` | ❌ | Step 1: Email + password |
| POST | `/auth/mfa/verify` | ❌ | Step 2: TOTP code |
| POST | `/auth/refresh` | ❌ | Refresh access token |
| POST | `/auth/logout` | ✅ | Invalidate all sessions |
| POST | `/auth/mfa/setup` | ✅ | Generate TOTP secret |
| POST | `/auth/mfa/enable` | ✅ | Verify + enable MFA |
| DELETE | `/auth/mfa/disable` | ✅ | Disable MFA |
| GET | `/auth/sessions` | ✅ | Get active session count |
| POST | `/auth/sessions` | ✅ | Revoke specific session |

### 3.2 Vault

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/vault/unlock` | ✅ | Unlock with master password |
| POST | `/vault/lock` | ✅ | Lock vault |
| GET | `/vault/status` | ✅ | Check unlock status |
| POST | `/vault/key/rotate` | ✅ | Rotate master key |
| GET | `/vault/credentials` | ✅ | List credentials (metadata) |
| POST | `/vault/credentials` | ✅ | Create credential |
| GET | `/vault/credentials/{id}` | ✅ | Get credential metadata |
| PATCH | `/vault/credentials/{id}` | ✅ | Update credential |
| DELETE | `/vault/credentials/{id}` | ✅ | Delete credential |
| GET | `/vault/credentials/{id}/decrypt` | ✅ | Decrypt credential data |
| POST | `/vault/credentials/{id}/share` | ✅ | Share credential |
| DELETE | `/vault/credentials/{id}/share` | ✅ | Revoke share |

### 3.3 Hosts

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/hosts` | ✅ | List hosts |
| POST | `/hosts` | ✅ | Create host |
| GET | `/hosts/{id}` | ✅ | Get host |
| PATCH | `/hosts/{id}` | ✅ | Update host |
| DELETE | `/hosts/{id}` | ✅ | Delete host |
| GET | `/hosts/{id}/health` | ✅ | Health check |

### 3.4 Terminal

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/ws/terminal` | ✅ WS | WebSocket terminal session |
| GET | `/sessions` | ✅ | List active sessions |
| DELETE | `/sessions/{id}` | ✅ | Close session |

### 3.5 Users

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/users/me` | ✅ | Get profile |
| PATCH | `/users/me` | ✅ | Update profile |
| POST | `/users/me/password` | ✅ | Change password |

### 3.6 Admin

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/admin/users` | ✅ Admin | List users |
| POST | `/admin/users` | ✅ Admin | Create user |
| GET | `/admin/users/{id}` | ✅ Admin | Get user |
| PATCH | `/admin/users/{id}` | ✅ Admin | Update user |
| DELETE | `/admin/users/{id}` | ✅ Admin | Deactivate user |
| GET | `/admin/audit` | ✅ Admin | Query audit log |
| GET | `/admin/audit/export` | ✅ Admin | Export audit log |
| POST | `/admin/audit/log` | ✅ Admin | Custom audit event |

### 3.7 Gateway

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/v1/gateway/ssh/connect` | ✅ | SSH connection |
| POST | `/api/v1/gateway/rdp/connect` | ✅ | RDP connection |
| POST | `/api/v1/gateway/vnc/connect` | ✅ | VNC connection |

---

## 4. Request/Response Examples

### 4.1 Login

**Request:**
```json
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "admin@vexa.local",
  "password": "***"
}
```

**Response (MFA Required):**
```json
{
  "message": "MFA required",
  "mfa_required": true,
  "mfa_token": "mfa_temp_abc123"
}
```

**Response (Success - no MFA):**
```json
{
  "access_token": "eyJhbG…s...",
  "refresh_token": "eyJhbG…s...",
  "expires_in": 3600,
  "token_type": "Bearer"
}
```

### 4.2 MFA Verification

**Request:**
```json
POST /api/v1/auth/mfa/verify
Content-Type: application/json

{
  "email": "admin@vexa.local",
  "totp_code": "123456"
}
```

### 4.3 Create Host

**Request:**
```json
POST /api/v1/hosts
Authorization: Bearer eyJhbG…s…
Content-Type: application/json

{
  "name": "Production Server",
  "address": "192.168.1.100",
  "protocol": "ssh",
  "port": 22,
  "tags": ["production", "web"],
  "description": "Main production web server"
}
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Production Server",
  "address": "192.168.1.100",
  "protocol": "ssh",
  "port": 22,
  "tags": ["production", "web"],
  "is_active": true,
  "created_at": "2026-05-28T12:00:00Z",
  "updated_at": "2026-05-28T12:00:00Z"
}
```

### 4.4 WebSocket Terminal

**Connection:**
```
GET ws://localhost:8080/ws/terminal?host_id=550e8400...&protocol=ssh
Authorization: Bearer eyJhbG…s…
```

**Message Protocol:**
```json
// Client > Server
{
  "type": "resize",
  "cols": 120,
  "rows": 40
}

// Server > Client
{
  "type": "data",
  "data": "base64encodedterminaldata"
}
```

---

## 5. Error Responses

### 5.1 Standard Error Format

```json
{
  "error": "error_code",
  "message": "Human readable description",
  "details": "Additional context (optional)"
}
```

### 5.2 HTTP Status Codes

| Code | Meaning | Example |
|------|---------|---------|
| 200 | OK | Success |
| 201 | Created | Resource created |
| 204 | No Content | Delete success |
| 400 | Bad Request | Invalid input |
| 401 | Unauthorized | Missing/invalid token |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource doesn't exist |
| 409 | Conflict | Duplicate/constraint violation |
| 429 | Rate Limited | Too many requests |
| 500 | Server Error | Internal error |
| 503 | Service Unavailable | Gateway down |

### 5.3 Validation Errors

```json
{
  "error": "validation_failed",
  "message": "Request validation failed",
  "details": {
    "email": ["Invalid email format"],
    "password": ["Must be at least 12 characters"]
  }
}
```

---

## 6. Rate Limiting

| Endpoint Group | Limit | Window |
|----------------|-------|--------|
| Authentication | 5 | 1 minute |
| API (general) | 100 | 1 minute |
| Admin endpoints | 30 | 1 minute |
| WebSocket | 10 connections | Per user |

**Headers:**
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 87
X-RateLimit-Reset: 1716897600
```

---

## 7. Pagination

List endpoints support cursor-based pagination:

```http
GET /api/v1/hosts?limit=20&offset=40
```

**Response:**
```json
{
  "data": [...],
  "pagination": {
    "total": 156,
    "limit": 20,
    "offset": 40,
    "has_more": true
  }
}
```

---

## 8. Filtering

### 8.1 Host Filters

```http
GET /api/v1/hosts?tag=production&group=webservers&protocol=ssh
```

### 8.2 Audit Log Filters

```http
GET /api/v1/admin/audit?event_type=login&user_id=xxx&start_time=2026-05-01&end_time=2026-05-28
```

---

## 9. WebSocket Protocol

### 9.1 Connection

```javascript
const ws = new WebSocket(
  'ws://localhost:8080/ws/terminal?host_id=xxx&protocol=ssh',
  [],
  { headers: { Authorization: '***' + token } }
);
```

### 9.2 Message Types

| Type | Direction | Description |
|------|-----------|-------------|
| `auth` | Client → Server | Send JWT token |
| `resize` | Client → Server | Terminal resize |
| `data` | Bidirectional | Terminal I/O |
| `heartbeat` | Client → Server | Keep-alive ping |
| `error` | Server → Client | Error notification |
| `close` | Server → Client | Session closed |

### 9.3 Binary Protocol (Desktop)

Desktop client uses binary WebSocket frames:

```
[1 byte: message type][payload...]

Message Types:
0x01 - Terminal data
0x02 - Resize
0x03 - Heartbeat
0x04 - File transfer
0x05 - SFTP command
0x06 - Close
```

---

## 10. SDK Examples

### 10.1 JavaScript/TypeScript

```typescript
import { SSHManagerClient } from '@vexa/sdk';

const client = new SSHManagerClient({
  baseURL: 'https://api.vexa.local',
  token: 'your-jwt-token'
});

// Login
const { accessToken } = await client.auth.login({
  email: 'user@example.com',
  password: 'password'
});

// List hosts
const hosts = await client.hosts.list({
  tag: 'production',
  limit: 20
});

// Create host
const host = await client.hosts.create({
  name: 'Web Server',
  address: '192.168.1.100',
  protocol: 'ssh',
  port: 22
});
```

### 10.2 Go

```go
package main

import (
    "context"
    "fmt"
    "github.com/soumabali/vexa/sdk-go"
)

func main() {
    client, _ := vexa.NewClient("https://api.vexa.local")
    
    // Login
    token, _ := client.Auth.Login(context.Background(), vexa.LoginRequest{
        Email:    "user@example.com",
        Password: "password",
    })
    
    // Use token
    client.SetToken(token.AccessToken)
    
    // List hosts
    hosts, _ := client.Hosts.List(context.Background(), vexa.ListHostsRequest{
        Limit: 20,
    })
    
    fmt.Printf("Found %d hosts\n", len(hosts.Data))
}
```

### 10.3 cURL

```bash
# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@vexa.local","password":"***"}'

# List hosts (with token)
curl http://localhost:8080/api/v1/hosts \
  -H "Authorization: Bearer ***"

# Create host
curl -X POST http://localhost:8080/api/v1/hosts \
  -H "Authorization: Bearer ***" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Web Server",
    "address": "192.168.1.100",
    "protocol": "ssh",
    "port": 22
  }'
```

---

## 11. Change Log

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-05-28 | Initial API release |
| 0.1.0 | 2026-05-28 | Alpha — MVP endpoints |

---

## Appendix: OpenAPI Tools

### Generate Client SDK
```bash
# TypeScript
openapi-generator-cli generate \
  -i docs/api/openapi.yaml \
  -g typescript-fetch \
  -o sdk/typescript

# Go
openapi-generator-cli generate \
  -i docs/api/openapi.yaml \
  -g go \
  -o sdk/go

# Python
openapi-generator-cli generate \
  -i docs/api/openapi.yaml \
  -g python \
  -o sdk/python
```

### Validate Spec
```bash
swagger-cli validate docs/api/openapi.yaml
```

### Preview UI
```bash
swagger-ui-express docs/api/openapi.yaml
```
