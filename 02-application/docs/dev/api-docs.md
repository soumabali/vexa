# vexa — API Documentation

> **Agent:** Full Stack Developer  
> **Status:** Generated 2026-05-28  
> **Version:** 0.1.0

---

## 1. Base URL

| Environment | URL |
|-------------|-----|
| Development | `http://localhost:8080` |
| Production | `https://api.vexa.local` |

All endpoints require `Content-Type: application/json` unless specified.

---

## 2. Authentication

### 2.1 Login
```http
POST /api/auth/login
```

**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecureP@ss123",
  "mfa_code": "123456"
}
```

**Response (200):**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJSUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 900,
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "role": "user",
    "mfa_enabled": true
  }
}
```

**Response (401 - Invalid credentials):**
```json
{
  "error": "AUTH_001",
  "message": "Invalid email or password"
}
```

**Response (401 - MFA required):**
```json
{
  "error": "AUTH_002",
  "message": "MFA verification required",
  "mfa_required": true
}
```

### 2.2 Register
```http
POST /api/auth/register
```

**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecureP@ss123",
  "confirm_password": "SecureP@ss123"
}
```

**Response (201):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "role": "user",
  "created_at": "2026-05-28T10:00:00Z"
}
```

### 2.3 Refresh Token
```http
POST /api/auth/refresh
```

**Request:**
```json
{
  "refresh_token": "eyJhbGciOiJSUzI1NiIs..."
}
```

**Response (200):**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "expires_in": 900
}
```

### 2.4 Logout
```http
POST /api/auth/logout
Authorization: Bearer {access_token}
```

**Response (204):** No content

---

## 3. MFA

### 3.1 Setup MFA
```http
POST /api/auth/mfa/setup
Authorization: Bearer {access_token}
```

**Response (200):**
```json
{
  "secret": "JBSWY3DPEHPK3PXP",
  "qr_code": "data:image/png;base64,iVBORw0KGgo...",
  "backup_codes": ["12345678", "87654321", "11112222"]
}
```

### 3.2 Verify MFA
```http
POST /api/auth/mfa/verify
Authorization: Bearer {access_token}
```

**Request:**
```json
{
  "code": "123456"
}
```

### 3.3 Disable MFA
```http
DELETE /api/auth/mfa
Authorization: Bearer {access_token}
```

**Request:**
```json
{
  "password": "SecureP@ss123",
  "code": "123456"
}
```

---

## 4. Users

### 4.1 Get Current User
```http
GET /api/users/me
Authorization: Bearer {access_token}
```

**Response (200):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "role": "user",
  "mfa_enabled": true,
  "created_at": "2026-05-28T10:00:00Z",
  "last_login": "2026-05-28T10:05:00Z"
}
```

### 4.2 Update Profile
```http
PATCH /api/users/me
Authorization: Bearer {access_token}
```

**Request:**
```json
{
  "email": "new@example.com"
}
```

### 4.3 Change Password
```http
POST /api/users/me/password
Authorization: Bearer {access_token}
```

**Request:**
```json
{
  "current_password": "OldP@ss123",
  "new_password": "NewP@ss456",
  "confirm_password": "NewP@ss456"
}
```

### 4.4 Delete Account
```http
DELETE /api/users/me
Authorization: Bearer {access_token}
```

**Request:**
```json
{
  "password": "SecureP@ss123"
}
```

---

## 5. Hosts

### 5.1 List Hosts
```http
GET /api/hosts
Authorization: Bearer {access_token}
```

**Query Parameters:**
- `page` (default: 1)
- `limit` (default: 20, max: 100)
- `search` (optional)
- `protocol` (optional: ssh|rdp|vnc)

**Response (200):**
```json
{
  "hosts": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "name": "Production Server",
      "hostname": "prod.example.com",
      "port": 22,
      "protocol": "ssh",
      "tags": ["production", "linux"],
      "created_at": "2026-05-28T10:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 20
}
```

### 5.2 Create Host
```http
POST /api/hosts
Authorization: Bearer {access_token}
```

**Request:**
```json
{
  "name": "Production Server",
  "hostname": "prod.example.com",
  "port": 22,
  "protocol": "ssh",
  "credential_id": "550e8400-e29b-41d4-a716-446655440002",
  "tags": ["production", "linux"]
}
```

### 5.3 Get Host
```http
GET /api/hosts/{id}
Authorization: Bearer {access_token}
```

### 5.4 Update Host
```http
PATCH /api/hosts/{id}
Authorization: Bearer {access_token}
```

### 5.5 Delete Host
```http
DELETE /api/hosts/{id}
Authorization: Bearer {access_token}
```

---

## 6. Credentials (Vault)

### 6.1 List Credentials
```http
GET /api/credentials
Authorization: Bearer {access_token}
```

**Response (200):**
```json
{
  "credentials": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "name": "Production SSH Key",
      "type": "ssh_key",
      "created_at": "2026-05-28T10:00:00Z"
    }
  ]
}
```

### 6.2 Create Credential
```http
POST /api/credentials
Authorization: Bearer {access_token}
```

**Request (Password):**
```json
{
  "name": "Root Password",
  "type": "password",
  "data": "SuperSecret123!"
}
```

**Request (SSH Key):**
```json
{
  "name": "Production Key",
  "type": "ssh_key",
  "data": "-----BEGIN OPENSSH PRIVATE KEY-----\n..."
}
```

### 6.3 Get Credential (decrypted)
```http
GET /api/credentials/{id}
Authorization: Bearer {access_token}
```

**Response (200):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440002",
  "name": "Root Password",
  "type": "password",
  "data": "SuperSecret123!",
  "created_at": "2026-05-28T10:00:00Z"
}
```

> **Note:** Credential data is decrypted using the user's master key (derived from password + Argon2).

### 6.4 Update Credential
```http
PATCH /api/credentials/{id}
Authorization: Bearer {access_token}
```

### 6.5 Delete Credential
```http
DELETE /api/credentials/{id}
Authorization: Bearer {access_token}
```

---

## 7. Gateway (Protocol Proxy)

### 7.1 Connect to Host
```http
POST /api/gateway/connect
Authorization: Bearer {access_token}
```

**Request:**
```json
{
  "host_id": "550e8400-e29b-41d4-a716-446655440001",
  "protocol": "ssh"
}
```

**Response (200):**
```json
{
  "session_id": "550e8400-e29b-41d4-a716-446655440003",
  "websocket_url": "wss://api.vexa.local/ws/terminal/550e8400-e29b-41d4-a716-446655440003",
  "expires_at": "2026-05-28T11:00:00Z"
}
```

### 7.2 WebSocket Terminal
```
GET /ws/terminal/{session_id}
Authorization: Bearer {access_token}
Upgrade: websocket
```

**Message Format:**
```json
{
  "type": "input",
  "data": "ls -la\n"
}
```

**Response Stream:**
```json
{
  "type": "output",
  "data": "total 128\ndrwxr-xr-x  5 user user  4096 May 28 10:00 .\n"
}
```

### 7.3 Disconnect
```http
POST /api/gateway/disconnect
Authorization: Bearer {access_token}
```

**Request:**
```json
{
  "session_id": "550e8400-e29b-41d4-a716-446655440003"
}
```

---

## 8. SFTP (File Manager)

### 8.1 List Directory
```http
GET /api/sftp/{host_id}/list?path=/home/user
Authorization: Bearer {access_token}
```

**Response (200):**
```json
{
  "path": "/home/user",
  "entries": [
    {
      "name": "documents",
      "type": "directory",
      "size": 4096,
      "modified": "2026-05-28T10:00:00Z",
      "permissions": "drwxr-xr-x"
    },
    {
      "name": "file.txt",
      "type": "file",
      "size": 1024,
      "modified": "2026-05-28T09:00:00Z",
      "permissions": "-rw-r--r--"
    }
  ]
}
```

### 8.2 Upload File
```http
POST /api/sftp/{host_id}/upload
Authorization: Bearer {access_token}
Content-Type: multipart/form-data
```

**Form Data:**
- `file`: binary
- `path`: `/home/user/uploads/`
- `overwrite`: `false`

### 8.3 Download File
```http
GET /api/sftp/{host_id}/download?path=/home/user/file.txt
Authorization: Bearer {access_token}
```

**Response:** Binary stream with `Content-Disposition: attachment`

### 8.4 Delete File/Directory
```http
DELETE /api/sftp/{host_id}/delete
Authorization: Bearer {access_token}
```

**Request:**
```json
{
  "path": "/home/user/file.txt"
}
```

### 8.5 WebSocket File Transfer
```
GET /ws/sftp/{host_id}
Authorization: Bearer {access_token}
Upgrade: websocket
```

**Chunked Upload:**
```json
{
  "type": "upload_chunk",
  "path": "/home/user/large-file.zip",
  "chunk_index": 0,
  "total_chunks": 100,
  "data": "base64encodedchunk..."
}
```

---

## 9. Sessions

### 9.1 List Active Sessions
```http
GET /api/sessions
Authorization: Bearer {access_token}
```

**Response (200):**
```json
{
  "sessions": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440003",
      "host_id": "550e8400-e29b-41d4-a716-446655440001",
      "host_name": "Production Server",
      "protocol": "ssh",
      "started_at": "2026-05-28T10:00:00Z",
      "duration": 3600,
      "bytes_sent": 1024000,
      "bytes_received": 2048000
    }
  ]
}
```

### 9.2 Get Session History
```http
GET /api/sessions/history
Authorization: Bearer {access_token}
```

### 9.3 Kill Session
```http
DELETE /api/sessions/{id}
Authorization: Bearer {access_token}
```

---

## 10. Audit Logs

### 10.1 Query Audit Logs
```http
GET /api/audit/logs
Authorization: Bearer {access_token}
```

**Query Parameters:**
- `start_date` (ISO 8601)
- `end_date` (ISO 8601)
- `action` (optional: login|logout|connect|disconnect|create|update|delete)
- `resource` (optional)

**Response (200):**
```json
{
  "logs": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440004",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "action": "connect",
      "resource": "host:550e8400-e29b-41d4-a716-446655440001",
      "ip_address": "192.168.1.100",
      "user_agent": "Mozilla/5.0...",
      "timestamp": "2026-05-28T10:00:00Z",
      "success": true
    }
  ]
}
```

### 10.2 Export Audit Logs
```http
GET /api/audit/logs/export?format=json|csv
Authorization: Bearer {access_token}
```

**Response:** File download

---

## 11. Error Codes

| Code | Description | HTTP Status |
|------|-------------|-------------|
| `AUTH_001` | Invalid credentials | 401 |
| `AUTH_002` | MFA required | 401 |
| `AUTH_003` | Invalid MFA code | 401 |
| `AUTH_004` | Token expired | 401 |
| `AUTH_005` | Token invalid | 401 |
| `HOST_001` | Host not found | 404 |
| `HOST_002` | Host unreachable | 502 |
| `HOST_003` | Connection refused | 502 |
| `VAULT_001` | Decryption failed | 500 |
| `VAULT_002` | Invalid master key | 401 |
| `RATE_001` | Rate limit exceeded | 429 |
| `PERM_001` | Insufficient permissions | 403 |
| `VAL_001` | Validation error | 400 |

---

## 12. Rate Limits

| Endpoint | Limit | Window |
|----------|-------|--------|
| `POST /api/auth/*` | 5 req | 1 minute |
| `GET /api/*` | 100 req | 1 minute |
| `POST /api/*` | 50 req | 1 minute |
| `WS /ws/*` | 10 connections | Per user |

---

## 13. WebSocket Protocol

### 13.1 Connection
All WebSocket connections require the access token in the `Authorization` header during handshake.

### 13.2 Message Format
```typescript
interface WebSocketMessage {
  type: 'input' | 'output' | 'error' | 'resize' | 'heartbeat' | 'file_chunk';
  data: string | object;
  timestamp?: string;
}
```

### 13.3 Heartbeat
Client must send heartbeat every 30 seconds:
```json
{
  "type": "heartbeat",
  "data": "ping"
}
```

Server responds:
```json
{
  "type": "heartbeat",
  "data": "pong"
}
```

### 13.4 Terminal Resize
```json
{
  "type": "resize",
  "data": {
    "cols": 120,
    "rows": 30
  }
}
```
