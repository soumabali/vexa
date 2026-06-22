# Secure Coding Guide: Multi-Protocol Terminal Manager

**Document ID:** SEC-CODE-001
**Version:** 1.0
**Classification:** Internal — Security Critical
**Owner:** Security Architect
**Date:** 2026-05-27

---

## 1. Input Validation Rules

### 1.1 Golden Rule

**NEVER trust any input. Validate every byte.**

All input is untrusted, regardless of source:
- HTTP request parameters, headers, body
- WebSocket messages
- Database records (other apps may have written them)
- File uploads
- External API responses
- Environment variables
- Configuration files

### 1.2 Validation Strategy: Allowlist Over Blocklist

```python
# BAD: Blocklist (incomplete, bypassable)
def is_valid_username(username):
    forbidden = ['admin', 'root', 'select', 'drop']
    return username.lower() not in forbidden

# GOOD: Allowlist (explicit, complete)
def is_valid_username(username):
    """
    Usernames must be:
    - 3-32 characters
    - Alphanumeric + underscore + hyphen
    - Cannot start with number
    """
    return bool(re.match(r'^[a-zA-Z][a-zA-Z0-9_-]{2,31}$', username))
```

### 1.3 Validation by Data Type

| Data Type | Validation Rules | Example |
|-----------|---------------|---------|
| **Username** | 3-32 chars, alphanumeric + `_-`, no spaces, case-insensitive unique | `john_doe-123` |
| **Password** | 16+ chars (admin: 24+), check against HIBP, no dictionary words | Not stored, only hashed |
| **Email** | RFC 5322 compliant, MX record verification, disposable email block | `user@example.com` |
| **UUID** | RFC 4122 v4 format, validate in application logic | `550e8400-e29b-41d4-a716-446655440000` |
| **IP Address** | Valid IPv4/IPv6, no private ranges for public endpoints | `203.0.113.1` |
| **Port** | Integer 1-65535, no well-known ports without justification | `8080` |
| **URL** | Parse and validate scheme, host, no file://, no javascript: | `https://example.com/path` |
| **File Path** | Canonicalize, no `..`, restrict to allowed directories | `/uploads/uuid-filename.txt` |
| **Command** | No shell execution, parameterized only, allowlist commands | `['ls', '-la', '/path']` |
| **Terminal Input** | Escape sequence filtering, length limits, rate limiting | Sanitized byte stream |

### 1.4 Framework-Specific Validation

```python
# Python (Pydantic) - Preferred
from pydantic import BaseModel, Field, validator, EmailStr
from typing import Optional

class CredentialCreateRequest(BaseModel):
    name: str = Field(..., min_length=1, max_length=128, regex=r'^[\w\s-]+$')
    host: str = Field(..., min_length=1, max_length=255)
    port: int = Field(..., ge=1, le=65535)
    protocol: str = Field(..., regex=r'^(ssh|rdp|vnc)$')
    username: str = Field(..., min_length=1, max_length=128)
    password: Optional[str] = Field(None, max_length=512)
    private_key: Optional[str] = Field(None, max_length=8192)
    
    @validator('host')
    def validate_host(cls, v):
        # No private IPs for public-facing credentials
        if is_private_ip(v):
            raise ValueError('Private IP addresses require admin approval')
        return v
    
    @validator('password', 'private_key')
    def validate_credential_material(cls, v, values):
        if not v and not values.get('private_key') and not values.get('password'):
            raise ValueError('Either password or private_key required')
        return v

# TypeScript (Zod) - Preferred
import { z } from 'zod';

const CredentialCreateSchema = z.object({
  name: z.string().min(1).max(128).regex(/^[\w\s-]+$/),
  host: z.string().min(1).max(255),
  port: z.number().int().min(1).max(65535),
  protocol: z.enum(['ssh', 'rdp', 'vnc']),
  username: z.string().min(1).max(128),
  password: z.string().max(512).optional(),
  privateKey: z.string().max(8192).optional(),
}).refine(data => data.password || data.privateKey, {
  message: "Either password or privateKey required"
});
```

### 1.5 File Upload Validation

```python
# Secure file upload handling
import magic
import hashlib
from pathlib import Path

class SecureFileUpload:
    ALLOWED_TYPES = {
        'application/x-pem-file',      # SSH keys
        'text/plain',                    # Config files
        'application/json',              # Settings
    }
    MAX_SIZE = 1024 * 1024  # 1MB
    
    @staticmethod
    def process_upload(file_stream, filename):
        # 1. Validate size BEFORE reading
        file_stream.seek(0, 2)
        size = file_stream.tell()
        file_stream.seek(0)
        
        if size > SecureFileUpload.MAX_SIZE:
            raise ValidationError("File too large")
        
        # 2. Read content
        content = file_stream.read()
        
        # 3. Validate magic number (NOT extension)
        mime = magic.from_buffer(content, mime=True)
        if mime not in SecureFileUpload.ALLOWED_TYPES:
            raise ValidationError(f"File type not allowed: {mime}")
        
        # 4. Additional content validation for SSH keys
        if mime == 'application/x-pem-file':
            if not content.startswith(b'-----BEGIN'):
                raise ValidationError("Invalid key format")
        
        # 5. Generate safe filename (not user-provided)
        safe_name = f"{uuid.uuid4()}.upload"
        
        # 6. Store outside web root, no execution permissions
        upload_path = Path('/secure/uploads') / safe_name
        upload_path.write_bytes(content)
        upload_path.chmod(0o640)
        
        # 7. Return reference (not path)
        return hashlib.sha256(content).hexdigest()
```

---

## 2. Output Encoding Requirements

### 2.1 Context-Specific Encoding

| Context | Encoding | Example |
|---------|----------|---------|
| **HTML Body** | HTML entities | `&lt;script&gt;` → `&lt;script&gt;` |
| **HTML Attribute** | Attribute encoding | `" onclick="alert(1)"` → `" onclick="alert(1)"` |
| **JavaScript** | JS encoding | `';alert(1);//` → `\x27\x3b\x61\x6c\x65\x72\x74\x28\x31\x29\x3b\x2f\x2f` |
| **CSS** | CSS encoding | `</style><script>...` → `\3c /style\3e...` |
| **URL** | Percent encoding | `<script>` → `%3Cscript%3E` |
| **SQL** | Parameterized queries | Never concatenate |
| **Shell** | Shell escaping | `$(rm -rf /)` → Quoted/parameterized |
| **LDAP** | LDAP encoding | `*` → `\2a`, `(` → `\28` |
| **XML** | XML entities | `<` → `&lt;`, `&` → `&amp;` |
| **JSON** | JSON string encoding | Control chars → `\uXXXX` |

### 2.2 Implementation Examples

```python
# Python - Using markupsafe for HTML
from markupsafe import Markup, escape

def render_user_content(user_input):
    # Automatically escaped when used in template
    return escape(user_input)

# NEVER use Markup() on untrusted input
dangerous = Markup(user_input)  # BAD!

# Template (Jinja2 auto-escapes by default)
# {{ user_input }}  ← Automatically escaped
# {{ user_input | safe }}  ← Explicit opt-in (DANGEROUS)
```

```typescript
// TypeScript/React - JSX auto-escapes
function UserDisplay({ username }: { username: string }) {
  // Automatically escaped
  return <div>{username}</div>;
  
  // DANGEROUS - only use with sanitized input
  return <div dangerouslySetInnerHTML={{ __html: username }} />;
}

// For terminal output in browser
function TerminalOutput({ text }: { text: string }) {
  // Terminal escape sequence filtering
  const sanitized = text
    .replace(/\x1b\[[0-9;]*[a-zA-Z]/g, '')  // ANSI escapes
    .replace(/\x07/g, '')                    // Bell
    .replace(/\x00-\x08\x0b\x0c\x0e-\x1f/g, ''); // Control chars
  
  return <pre>{sanitized}</pre>;
}
```

### 2.3 Terminal-Specific Output Handling

```python
# Terminal escape sequence handling
import re

class TerminalSanitizer:
    """
    Sanitize terminal output for safe display in browser.
    
    Strategy:
    - Allow: Basic formatting (colors, bold, underline)
    - Block: Cursor movement, screen clearing, alternate buffer
    - Block: OSC sequences (can execute code in some terminals)
    """
    
    # Allowed sequences (colors, styles)
    ALLOWED_SGR = re.compile(
        r'\x1b\['
        r'(?:[0-9;]*m)'  # SGR (colors, styles)
        r')
    
    # Blocked sequences
    BLOCKED_SEQUENCES = [
        re.compile(r'\x1b\[[0-9;]*[ABCDEHJKLSTf]'),  # Cursor movement, clear
        re.compile(r'\x1b\[[?][0-9]*[hl]'),          # Set/reset mode
        re.compile(r'\x1b\][0-9;]*\x07'),           # OSC ( Operating System Command )
        re.compile(r'\x1b[\(\)][AB012]'),            # Character set selection
        re.compile(r'\x1b[c]'),                      # Device attributes
        re.compile(r'\x1b[78]'),                     # Save/restore cursor
        re.compile(r'\x1b[M]'),                       # Mouse tracking
    ]
    
    @classmethod
    def sanitize(cls, data: bytes) -> bytes:
        """Sanitize terminal output for safe display."""
        # First, remove blocked sequences
        for pattern in cls.BLOCKED_SEQUENCES:
            data = pattern.sub(b'', data)
        
        # Verify no remaining dangerous sequences
        if b'\x1b' in data:
            # Remove all remaining escape sequences
            data = re.sub(rb'\x1b\[[0-9;]*[a-zA-Z]', b'', data)
            data = re.sub(rb'\x1b][0-9;]*\x07', b'', data)
        
        return data
    
    @classmethod
    def sanitize_for_logs(cls, data: bytes) -> str:
        """Sanitize for audit logs - more restrictive."""
        # Remove ALL escape sequences for logs
        data = re.sub(rb'\x1b\[[0-9;]*[a-zA-Z]', b'', data)
        data = re.sub(rb'\x07', b'', data)
        
        # Truncate if needed
        if len(data) > 10000:
            data = data[:10000] + b'...[truncated]'
        
        return data.decode('utf-8', errors='replace')
```

---

## 3. Cryptographic Best Practices

### 3.1 Algorithm Selection

| Purpose | Recommended | Minimum Acceptable | NEVER Use |
|---------|-------------|-------------------|-----------|
| Symmetric encryption | AES-256-GCM | AES-128-GCM | DES, 3DES, RC4, CBC mode |
| Stream encryption | ChaCha20-Poly1305 | - | CTR without MAC |
| Password hashing | Argon2id | scrypt, PBKDF2 | MD5, SHA1, SHA256 alone |
| Digital signatures | Ed25519 | ECDSA P-256 | RSA < 2048, DSA |
| Key derivation | HKDF-SHA256 | - | Simple hash, MD5, SHA1 |
| Hashing (non-password) | SHA-256 | SHA-256 | MD5, SHA1 |
| Random generation | /dev/urandom, getrandom() | CSPRNG | Math.random(), rand() |
| Key exchange | X25519 | ECDH P-256 | Static DH, RSA key exchange |
| MAC | HMAC-SHA256 | - | CBC-MAC, MD5-based |
| Checksum | BLAKE3 | SHA-256 | MD5, CRC32 |

### 3.2 Password Hashing (Argon2id)

```python
# Python - Using argon2-cffi
from argon2 import PasswordHasher
from argon2.exceptions import VerifyMismatchError

# Production parameters (tune based on your hardware)
ph = PasswordHasher(
    time_cost=3,        # iterations
    memory_cost=65536,  # 64 MB
    parallelism=4,      # threads
    hash_len=32,        # output length
    salt_len=16         # salt length
)

# Hash password
def hash_password(password: str) -> str:
    """Hash a password for storage."""
    # Validate password meets policy FIRST
    if not validate_password_policy(password):
        raise ValueError("Password does not meet policy")
    
    return ph.hash(password)

# Verify password
def verify_password(password: str, hash_string: str) -> bool:
    """Verify password against stored hash."""
    try:
        ph.verify(hash_string, password)
        return True
    except VerifyMismatchError:
        return False

# Check if rehash needed (parameters upgraded)
def needs_rehash(hash_string: str) -> bool:
    return ph.check_needs_rehash(hash_string)
```

```typescript
// TypeScript - Using argon2-browser or Node.js argon2
import argon2 from 'argon2';

const hashPassword = async (password: string): Promise<string> => {
  // Validate policy first
  if (!validatePasswordPolicy(password)) {
    throw new Error('Password does not meet policy');
  }
  
  return argon2.hash(password, {
    type: argon2id,
    memoryCost: 65536,
    timeCost: 3,
    parallelism: 4,
    hashLength: 32,
    saltLength: 16
  });
};

const verifyPassword = async (password: string, hash: string): Promise<boolean> => {
  try {
    return await argon2.verify(hash, password);
  } catch {
    return false;
  }
};
```

### 3.3 Symmetric Encryption (AES-256-GCM)

```python
# Python - Using cryptography library
from cryptography.hazmat.primitives.ciphers.aead import AESGCM
from cryptography.hazmat.backends import default_backend
import os
import base64

class SecureEncryption:
    @staticmethod
    def encrypt(plaintext: bytes, key: bytes, associated_data: bytes = None) -> bytes:
        """
        Encrypt with AES-256-GCM.
        
        Args:
            plaintext: Data to encrypt
            key: 32-byte key
            associated_data: Additional authenticated data (optional)
        
        Returns:
            nonce (12 bytes) + ciphertext + tag (16 bytes)
        """
        if len(key) != 32:
            raise ValueError("Key must be 32 bytes for AES-256")
        
        # NEVER reuse nonce with same key
        nonce = os.urandom(12)  # 96 bits as recommended
        
        aesgcm = AESGCM(key)
        ciphertext = aesgcm.encrypt(nonce, plaintext, associated_data)
        
        # Format: nonce || ciphertext (includes tag)
        return nonce + ciphertext
    
    @staticmethod
    def decrypt(ciphertext: bytes, key: bytes, associated_data: bytes = None) -> bytes:
        """Decrypt AES-256-GCM ciphertext."""
        if len(key) != 32:
            raise ValueError("Key must be 32 bytes for AES-256")
        
        if len(ciphertext) < 28:  # 12 nonce + 16 tag minimum
            raise ValueError("Ciphertext too short")
        
        nonce = ciphertext[:12]
        encrypted = ciphertext[12:]
        
        aesgcm = AESGCM(key)
        return aesgcm.decrypt(nonce, encrypted, associated_data)

# Example: Envelope encryption for credential storage
class CredentialVault:
    def __init__(self, master_key: bytes):
        self.master_key = master_key
    
    def store_credential(self, credential: bytes) -> dict:
        """Store credential with envelope encryption."""
        # Generate unique data key
        data_key = os.urandom(32)
        
        # Encrypt credential with data key
        encrypted_credential = SecureEncryption.encrypt(credential, data_key)
        
        # Encrypt data key with master key
        encrypted_key = SecureEncryption.encrypt(data_key, self.master_key)
        
        # Clear plaintext data key from memory
        import ctypes
        ctypes.memset(id(data_key) + 20, 0, len(data_key))
        
        return {
            'encrypted_credential': base64.b64encode(encrypted_credential).decode(),
            'encrypted_key': base64.b64encode(encrypted_key).decode(),
            'algorithm': 'AES-256-GCM',
            'key_wrapping': 'AES-256-GCM'
        }
```

### 3.4 Digital Signatures (Ed25519)

```python
# Python - Using cryptography
from cryptography.hazmat.primitives.asymmetric.ed25519 import (
    Ed25519PrivateKey, Ed25519PublicKey
)
import base64

class MessageSigner:
    @staticmethod
    def generate_keypair() -> tuple[str, str]:
        """Generate Ed25519 keypair. Returns (private_b64, public_b64)."""
        private_key = Ed25519PrivateKey.generate()
        public_key = private_key.public_key()
        
        private_bytes = private_key.private_bytes(
            encoding=serialization.Encoding.Raw,
            format=serialization.PrivateFormat.Raw,
            encryption_algorithm=serialization.NoEncryption()
        )
        public_bytes = public_key.public_bytes(
            encoding=serialization.Encoding.Raw,
            format=serialization.PublicFormat.Raw
        )
        
        return (
            base64.b64encode(private_bytes).decode(),
            base64.b64encode(public_bytes).decode()
        )
    
    @staticmethod
    def sign(message: bytes, private_key_b64: str) -> str:
        """Sign message with Ed25519."""
        private_bytes = base64.b64decode(private_key_b64)
        private_key = Ed25519PrivateKey.from_private_bytes(private_bytes)
        
        signature = private_key.sign(message)
        return base64.b64encode(signature).decode()
    
    @staticmethod
    def verify(message: bytes, signature_b64: str, public_key_b64: str) -> bool:
        """Verify Ed25519 signature."""
        try:
            public_bytes = base64.b64decode(public_key_b64)
            public_key = Ed25519PublicKey.from_public_bytes(public_bytes)
            
            signature = base64.b64decode(signature_b64)
            public_key.verify(signature, message)
            return True
        except Exception:
            return False
```

### 3.5 Secure Random Generation

```python
# Python - Using secrets module (os.urandom wrapper)
import secrets
import string

def generate_secure_token(length: int = 32) -> str:
    """Generate cryptographically secure random token."""
    # Use secrets, NOT random
    alphabet = string.ascii_letters + string.digits
    return ''.join(secrets.choice(alphabet) for _ in range(length))

def generate_session_id() -> str:
    """Generate secure session identifier."""
    # 256-bit entropy
    return secrets.token_urlsafe(32)

def generate_nonce(length: int = 12) -> bytes:
    """Generate cryptographic nonce."""
    return secrets.token_bytes(length)
```

```typescript
// TypeScript/Node.js
import { randomBytes } from 'crypto';

const generateSecureToken = (length: number = 32): string => {
  return randomBytes(length).toString('base64url');
};

const generateSessionId = (): string => {
  return randomBytes(32).toString('hex');
};
```

### 3.6 Cryptographic Anti-Patterns

| Anti-Pattern | Why Dangerous | Correct Approach |
|-------------|---------------|------------------|
| Rolling own crypto | Subtle vulnerabilities | Use vetted libraries (libsodium, cryptography) |
| ECB mode | Leaks patterns, no integrity | Use GCM or ChaCha20-Poly1305 |
| Hardcoded keys | Key exposure in source | Key management service, environment injection |
| Weak key derivation | Brute forceable | Argon2id, scrypt |
| Unauthenticated encryption | Tampering possible | Always use AEAD modes |
| Static IV/nonce | Complete security break | Unique nonce per encryption |
| Compression + encryption | CRIME/BREACH attacks | Compress then encrypt, or disable compression |
| RSA without OAEP | Padding oracle attacks | Use OAEP or switch to ECDH/Ed25519 |
| Short RSA keys | Factorable | Minimum 2048-bit, prefer ECDH |
| SHA1 for signatures | Collision attacks | SHA-256 or Ed25519 |

---

## 4. Error Handling (No Information Leakage)

### 4.1 Error Response Strategy

```python
# Error response categories
class ErrorResponse:
    """
    Secure error responses - never leak internal details.
    
    Internal: Full details logged server-side
    External: Generic message to client
    """
    
    # Client-facing messages (generic)
    CLIENT_MESSAGES = {
        'auth_failure': 'Authentication failed',
        'not_found': 'Resource not found',
        'permission_denied': 'Access denied',
        'validation_error': 'Invalid input',
        'rate_limited': 'Too many requests',
        'server_error': 'Internal server error',
        'service_unavailable': 'Service temporarily unavailable',
    }
    
    @staticmethod
    def handle(exception: Exception, public: bool = True) -> dict:
        """
        Convert exception to response.
        
        If public=True: Return safe, generic message
        If public=False: Return detailed message (internal APIs only)
        """
        # Log full details internally
        import logging
        logger = logging.getLogger('security')
        
        error_id = generate_secure_token(8)
        logger.error(
            f"Error {error_id}: {type(exception).__name__}: {str(exception)}",
            exc_info=True,
            extra={
                'error_id': error_id,
                'error_type': type(exception).__name__,
                'stack_trace': traceback.format_exc()
            }
        )
        
        if public:
            # Return generic message with error ID for support reference
            error_key = map_exception_to_key(exception)
            return {
                'error': ErrorResponse.CLIENT_MESSAGES.get(
                    error_key, 
                    'An error occurred'
                ),
                'error_id': error_id,  # For support to look up internally
                'timestamp': datetime.utcnow().isoformat()
            }
        else:
            # Internal API - more detail (but still no secrets)
            return {
                'error': type(exception).__name__,
                'message': str(exception),
                'error_id': error_id
            }
```

### 4.2 Exception Mapping

```python
# Map internal exceptions to safe external messages
def map_exception_to_key(exception: Exception) -> str:
    """Map exception to client-safe error key."""
    
    exception_map = {
        # Authentication errors
        AuthenticationError: 'auth_failure',
        InvalidTokenError: 'auth_failure',
        ExpiredTokenError: 'auth_failure',
        
        # Authorization errors
        PermissionDenied: 'permission_denied',
        InsufficientPrivileges: 'permission_denied',
        
        # Not found errors
        NotFoundError: 'not_found',
        DoesNotExist: 'not_found',
        
        # Validation errors
        ValidationError: 'validation_error',
        SchemaError: 'validation_error',
        
        # Rate limiting
        RateLimitExceeded: 'rate_limited',
        TooManyRequests: 'rate_limited',
        
        # Database errors -> generic server error
        DatabaseError: 'server_error',
        IntegrityError: 'server_error',
        
        # Catch-all
        Exception: 'server_error'
    }
    
    for exc_type, key in exception_map.items():
        if isinstance(exception, exc_type):
            return key
    
    return 'server_error'
```

### 4.3 What NOT to Include in Error Messages

| Never Include | Why | What to Do Instead |
|-------------|-----|-------------------|
| Stack traces | Reveals code structure, libraries | Log internally, return error ID |
| Database errors | Reveals schema, ORM, SQL | Generic "server error" |
| File paths | Reveals directory structure | Generic error |
| Internal IP addresses | Reveals network topology | Generic error |
| "User not found" vs "Password incorrect" | User enumeration | Always "Authentication failed" |
| Version numbers | Helps attacker target vulnerabilities | Remove from headers |
| "Timed out after 5000ms" | Reveals timing internals | "Service unavailable" |
| Detailed validation errors | Information gathering for attackers | Generic "Invalid input" |
| "You don't have permission to access X" | Information disclosure | "Access denied" |

### 4.4 HTTP Security Headers

```python
# Secure HTTP headers for all responses
SECURITY_HEADERS = {
    # Prevent MIME type sniffing
    'X-Content-Type-Options': 'nosniff',
    
    # Prevent clickjacking
    'X-Frame-Options': 'SAMEORIGIN',
    
    # XSS protection (legacy browsers)
    'X-XSS-Protection': '1; mode=block',
    
    # Referrer policy
    'Referrer-Policy': 'strict-origin-when-cross-origin',
    
    # Permissions policy
    'Permissions-Policy': (
        'accelerometer=(), camera=(), geolocation=(), gyroscope=(), '
        'magnetometer=(), microphone=(), payment=(), usb=()'
    ),
    
    # Content Security Policy (strict for terminal manager)
    'Content-Security-Policy': (
        "default-src 'self'; "
        "script-src 'self'; "
        "style-src 'self' 'unsafe-inline'; "
        "img-src 'self' data:; "
        "font-src 'self'; "
        "connect-src 'self' wss:; "
        "media-src 'none'; "
        "object-src 'none'; "
        "frame-ancestors 'none'; "
        "base-uri 'self'; "
        "form-action 'self'; "
        "upgrade-insecure-requests;"
    ),
    
    # HSTS (only on HTTPS)
    'Strict-Transport-Security': 'max-age=31536000; includeSubDomains; preload',
    
    # Expect CT
    'Expect-CT': 'max-age=86400, enforce',
    
    # Remove server identification
    'Server': 'Web Server',
}
```

---

## 5. Logging Requirements

### 5.1 What to Log

| Event Type | Data to Log | Sensitivity | Retention |
|-----------|-------------|-------------|-----------|
| **Authentication success** | User ID, timestamp, IP, MFA method, device fingerprint | Medium | 1 year |
| **Authentication failure** | User ID, timestamp, IP, failure reason (generic) | Medium | 1 year |
| **Session start/end** | User ID, session ID, timestamp, IP, device | Medium | 1 year |
| **Privilege change** | Admin ID, target user, old/new roles, timestamp | High | 7 years |
| **Credential access** | User ID, credential ID (not value!), action, timestamp | High | 7 years |
| **Terminal session start** | User ID, target server (ID only), protocol, timestamp | High | 7 years |
| **Terminal session end** | User ID, duration, bytes transferred, termination reason | Medium | 1 year |
| **Config change** | Admin ID, setting changed, old/new values (not secrets) | High | 7 years |
| **Vault operations** | Operation type, key ID, success/failure, timestamp | Critical | 7 years |
| **Backup/restore** | Initiator, scope, timestamp, success/failure | High | 7 years |
| **Security alerts** | Alert type, severity, triggered rules, timestamp | Critical | 7 years |
| **Admin actions** | Admin ID, action, target, timestamp | High | 7 years |
| **API errors** | Endpoint, error type (generic), error ID, timestamp | Low | 90 days |
| **Rate limit triggered** | User/IP, endpoint, limit, timestamp | Medium | 90 days |

### 5.2 What NOT to Log

| Never Log | Reason | Alternative |
|-----------|--------|-------------|
| Passwords (plaintext or hashed) | Credential exposure | Log "password changed" event only |
| MFA secrets / TOTP seeds | Account takeover | Log "MFA enrolled/verified" only |
| Private keys | Complete cryptographic compromise | Log "key accessed/rotated" only |
| Session tokens / JWTs | Session hijacking | Log token hash (first 8 chars) |
| Credit card numbers | PCI DSS violation | Log last 4 digits only |
| Social security numbers | Identity theft | Use anonymized ID |
| Decryption keys | Complete data exposure | Log "decryption performed" only |
| API keys | Unauthorized access | Log "API key used" with key ID |
| Full HTTP Authorization header | Token exposure | Log "Authorization: Bearer ***" |
| Database connection strings | Database compromise | Use DSN references |
| Environment variables with secrets | Secret exposure | Log variable names only |
| Stack traces (in production) | Information disclosure | Log internally, return error ID |
| User input without sanitization | Log injection attacks | Sanitize/escape before logging |

### 5.3 Secure Logging Implementation

```python
# Python - Secure structured logging
import logging
import json
from datetime import datetime
from typing import Any, Dict

class SecurityLogger:
    """
    Security-aware logger that prevents secret leakage.
    
    Features:
    - Structured JSON output
    - Automatic secret redaction
    - Tamper-evident log signing
    - Async delivery to SIEM
    """
    
    # Patterns to redact
    SENSITIVE_PATTERNS = [
        (r'password[=:]\s*\S+', 'password=***'),
        (r'token[=:]\s*\S+', 'token=***'),
        (r'key[=:]\s*\S+', 'key=***'),
        (r'secret[=:]\s*\S+', 'secret=***'),
        (r'Bearer\s+\S+', 'Bearer ***'),
        (r'Basic\s+\S+', 'Basic ***'),
    ]
    
    # Fields to always remove
    SENSITIVE_FIELDS = {
        'password', 'passwd', 'pwd', 'secret', 'token', 
        'api_key', 'private_key', 'mfa_secret', 'credit_card',
        'ssn', 'session_token', 'jwt', 'bearer'
    }
    
    def __init__(self, name: str):
        self.logger = logging.getLogger(name)
    
    def _sanitize_data(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Remove sensitive fields from log data."""
        sanitized = {}
        for key, value in data.items():
            # Check if key is sensitive
            if any(sensitive in key.lower() for sensitive in self.SENSITIVE_FIELDS):
                sanitized[key] = '***REDACTED***'
            elif isinstance(value, dict):
                sanitized[key] = self._sanitize_data(value)
            elif isinstance(value, list):
                sanitized[key] = [
                    self._sanitize_data(item) if isinstance(item, dict) else item
                    for item in value
                ]
            else:
                sanitized[key] = value
        return sanitized
    
    def _redact_message(self, message: str) -> str:
        """Redact sensitive patterns from message string."""
        import re
        for pattern, replacement in self.SENSITIVE_PATTERNS:
            message = re.sub(pattern, replacement, message, flags=re.IGNORECASE)
        return message
    
    def security_event(self, event_type: str, **kwargs):
        """Log a security event with structured data."""
        
        # Build log entry
        entry = {
            'timestamp': datetime.utcnow().isoformat() + 'Z',
            'event_type': event_type,
            'event_id': generate_secure_token(8),
            'environment': os.environ.get('ENV', 'unknown'),
            'service': 'terminal-manager',
            'data': self._sanitize_data(kwargs)
        }
        
        # Sign the entry (for integrity)
        entry['_signature'] = self._sign_entry(entry)
        
        # Log as JSON
        self.logger.info(json.dumps(entry))
        
        # Also forward to SIEM (async)
        self._forward_to_siem(entry)
    
    def _sign_entry(self, entry: dict) -> str:
        """Create integrity signature for log entry."""
        from cryptography.hazmat.primitives import hashes, hmac
        
        # Use HMAC with log signing key
        signing_key = get_log_signing_key()
        
        canonical = json.dumps(entry, sort_keys=True)
        h = hmac.HMAC(signing_key, hashes.SHA256())
        h.update(canonical.encode())
        return base64.b64encode(h.finalize()).decode()
    
    def audit(self, action: str, user_id: str, resource: str, result: str, **details):
        """
        Log audit event (immutable, high retention).
        
        These go to separate append-only stream.
        """
        self.security_event(
            'AUDIT',
            action=action,
            user_id=user_id,
            resource=resource,
            result=result,
            details=details,
            _immutable=True
        )

# Usage examples
logger = SecurityLogger('terminal-manager')

# Authentication event
logger.security_event(
    'AUTH_SUCCESS',
    user_id='user-123',
    method='password+mfa_totp',
    ip_address='203.0.113.42',
    device_fingerprint='abc123...',
    mfa_verified=True
)

# Credential access (audit trail)
logger.audit(
    action='CREDENTIAL_READ',
    user_id='user-123',
    resource='credential-456',
    result='SUCCESS',
    access_reason='terminal_session',
    session_id='sess-789'
)

# Error (with no sensitive data)
logger.security_event(
    'AUTH_FAILURE',
    user_id='user-123',
    reason='invalid_credentials',  # Generic
    ip_address='203.0.113.42',
    error_id='err-abc123'  # For internal lookup
)
```

### 5.4 Log Integrity Verification

```python
# Verify log integrity
def verify_log_chain(logs: list[dict]) -> bool:
    """
    Verify integrity of a sequence of log entries.
    Each entry should be signed and verifiable.
    """
    previous_hash = None
    
    for entry in logs:
        # Verify entry signature
        signature = entry.pop('_signature', None)
        if not signature:
            return False
        
        # Recalculate signature
        expected = calculate_signature(entry)
        if not hmac.compare_digest(signature, expected):
            return False
        
        # Verify chain (optional - link to previous)
        if previous_hash:
            if entry.get('_previous_hash') != previous_hash:
                return False
        
        # Calculate hash for chain
        previous_hash = hash_entry(entry)
    
    return True
```

---

## 6. Terminal-Specific Secure Coding

### 6.1 Command Execution

```python
# NEVER use shell=True or os.system()
# BAD:
os.system(f"ssh {user}@{host}")  # Command injection!
subprocess.run(f"ssh {user}@{host}", shell=True)  # Command injection!

# GOOD:
import subprocess

# Use parameterized execution
subprocess.run(
    ['ssh', '-o', 'StrictHostKeyChecking=yes', f'{user}@{host}'],
    capture_output=True,
    text=True,
    timeout=30
)

# Even better - use library instead of shelling out
import paramiko

client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.RejectPolicy())  # Strict!
client.connect(
    hostname=host,
    username=user,
    pkey=private_key,  # Never use password if possible
    look_for_keys=False,  # Don't use agent keys
    allow_agent=False,    # Don't use SSH agent
    timeout=30
)
```

### 6.2 WebSocket Security

```typescript
// Secure WebSocket server implementation
import { WebSocketServer } from 'ws';
import { verifyToken } from './auth';

const wss = new WebSocketServer({ 
  port: 8443,
  // Validate origin
  verifyClient: async (info, callback) => {
    const origin = info.origin;
    const allowedOrigins = [
      'https://terminal.example.com',
      'https://app.terminal.example.com'
    ];
    
    if (!allowedOrigins.includes(origin)) {
      callback(false, 403, 'Forbidden');
      return;
    }
    
    // Validate JWT from query or header
    const token = extractToken(info.req);
    try {
      const user = await verifyToken(token);
      info.req.user = user;  // Attach for later use
      callback(true);
    } catch {
      callback(false, 401, 'Unauthorized');
    }
  }
});

wss.on('connection', (ws, req) => {
  const user = req.user;
  
  // Rate limiting per user
  if (!checkRateLimit(user.id, 'websocket')) {
    ws.close(1008, 'Rate limit exceeded');
    return;
  }
  
  // Message size limits
  ws.on('message', (data) => {
    if (data.length > 65536) {  // 64KB max
      ws.close(1009, 'Message too large');
      return;
    }
    
    // Validate message format
    let message;
    try {
      message = JSON.parse(data.toString());
    } catch {
      ws.close(1007, 'Invalid message format');
      return;
    }
    
    // Validate message schema
    if (!validateTerminalMessage(message)) {
      ws.close(1007, 'Invalid message schema');
      return;
    }
    
    // Process message...
  });
  
  // Timeout for idle connections
  const idleTimeout = setTimeout(() => {
    ws.close(1001, 'Idle timeout');
  }, 15 * 60 * 1000);  // 15 minutes
  
  ws.on('pong', () => {
    clearTimeout(idleTimeout);
    // Reset timeout...
  });
});
```

### 6.3 Session Recording Security

```python
# Secure session recording
class SessionRecorder:
    """
    Record terminal sessions securely.
    
    Security requirements:
    - Encrypt recordings at rest
    - Access control (who can view)
    - Automatic PII redaction
    - Retention policies
    - Tamper detection
    """
    
    def __init__(self):
        self.encryption_key = get_recording_encryption_key()
    
    def start_recording(self, session_id: str, user_id: str) -> str:
        """Start recording a new session."""
        recording_id = generate_secure_token(16)
        
        # Create encrypted recording stream
        recording_path = self._get_recording_path(recording_id)
        
        # Log recording start
        logger.audit(
            'RECORDING_START',
            user_id=user_id,
            resource=recording_id,
            result='SUCCESS',
            session_id=session_id
        )
        
        return recording_id
    
    def write_frame(self, recording_id: str, data: bytes, direction: str):
        """
        Write a frame to the recording.
        
        direction: 'client_to_server' or 'server_to_client'
        """
        # Redact potential PII before storage
        data = self._redact_pii(data)
        
        # Add frame header
        frame = {
            'timestamp': datetime.utcnow().isoformat(),
            'direction': direction,
            'data': base64.b64encode(data).decode()
        }
        
        # Encrypt and append
        encrypted = self._encrypt_frame(json.dumps(frame).encode())
        
        # Append to recording file
        path = self._get_recording_path(recording_id)
        with open(path, 'ab') as f:
            # Write length-prefixed encrypted frame
            f.write(struct.pack('>I', len(encrypted)))
            f.write(encrypted)
    
    def _redact_pii(self, data: bytes) -> bytes:
        """Redact potential PII from terminal output."""
        text = data.decode('utf-8', errors='replace')
        
        # Redact patterns
        patterns = [
            (r'\b\d{3}-\d{2}-\d{4}\b', '[SSN-REDACTED]'),      # SSN
            (r'\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b', '[CC-REDACTED]'),  # Credit card
            (r'[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}', '[EMAIL-REDACTED]'),  # Email
            (r'password[=:]\s*\S+', 'password=***'),             # Passwords
            (r'ssh-rsa\s+AAAA[0-9A-Za-z+/]+[=]{0,3}', '[SSH-KEY-REDACTED]'),  # SSH keys
        ]
        
        for pattern, replacement in patterns:
            text = re.sub(pattern, replacement, text)
        
        return text.encode('utf-8')
    
    def access_recording(self, recording_id: str, requester_id: str) -> Optional[bytes]:
        """
        Access a recording with authorization check.
        """
        # Check authorization
        if not self._can_access_recording(recording_id, requester_id):
            logger.audit(
                'RECORDING_ACCESS_DENIED',
                user_id=requester_id,
                resource=recording_id,
                result='DENIED'
            )
            return None
        
        # Log access
        logger.audit(
            'RECORDING_ACCESS',
            user_id=requester_id,
            resource=recording_id,
            result='SUCCESS'
        )
        
        # Decrypt and return
        return self._decrypt_recording(recording_id)
```

---

## 7. Dependency Security

### 7.1 Dependency Management

```yaml
# .github/dependabot.yml
version: 2
updates:
  - package-ecosystem: "pip"
    directory: "/"
    schedule:
      interval: "daily"
    open-pull-requests-limit: 10
    security-updates-only: false
    
  - package-ecosystem: "npm"
    directory: "/frontend"
    schedule:
      interval: "daily"
    
  - package-ecosystem: "docker"
    directory: "/"
    schedule:
      interval: "weekly"
```

### 7.2 Supply Chain Security

| Control | Implementation | Frequency |
|---------|---------------|-----------|
| SBOM generation | Syft, CycloneDX | Every build |
| Vulnerability scanning | Trivy, Snyk | Every build + daily |
| License compliance | FOSSA, Snyk | Every build |
| Signed commits | GPG signing mandatory | Every commit |
| Verified builds | Reproducible builds, SLSA | Every release |
| Container signing | Cosign | Every build |
| Dependency pinning | Lock files committed | Always |

---

## 8. Code Review Security Checklist

### 8.1 Pre-Merge Checklist

| Check | Question | Severity |
|-------|----------|----------|
| Input validation | Are all inputs validated with allowlists? | Blocking |
| Output encoding | Is all output encoded for its context? | Blocking |
| Cryptography | Are vetted algorithms and libraries used? | Blocking |
| Authentication | Are all endpoints authenticated? | Blocking |
| Authorization | Is authorization checked on every access? | Blocking |
| Secrets | Are no secrets hardcoded or logged? | Blocking |
| SQL/NoSQL | Are parameterized queries used everywhere? | Blocking |
| Commands | Is shell execution avoided? | Blocking |
| Error handling | Do errors leak no sensitive information? | Blocking |
| Logging | Are sensitive fields redacted from logs? | Blocking |
| Session management | Are sessions securely managed? | Blocking |
| Rate limiting | Are sensitive endpoints rate limited? | Blocking |
| Dependencies | Are all dependencies scanned? | Blocking |
| Tests | Are security tests included? | Blocking |
| Documentation | Are security considerations documented? | Non-blocking |

---

**Document History**
| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-05-27 | Security Architect | Initial secure coding guide |
