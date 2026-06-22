# Security Test Cases: vexa (vexa)

**Document ID:** QA-TESTCASES-001  
**Version:** 1.0  
**Date:** 2026-05-27  
**Owner:** QA / Security Tester

---

## Table of Contents

1. [Authentication Bypass Tests](#1-authentication-bypass-tests)
2. [Authorization Tests](#2-authorization-tests)
3. [Input Validation Tests](#3-input-validation-tests)
4. [Session Management Tests](#4-session-management-tests)
5. [Cryptographic Tests](#5-cryptographic-tests)
6. [WebSocket Security Tests](#6-websocket-security-tests)
7. [Credential Vault Tests](#7-credential-vault-tests)
8. [File Upload/Download Tests](#8-file-uploaddownload-tests)
9. [Rate Limiting Tests](#9-rate-limiting-tests)
10. [CSRF Tests](#10-csrf-tests)

---

## 1. Authentication Bypass Tests

### TC-AUTH-001: JWT Algorithm Confusion Attack
**Objective:** Verify that JWT tokens cannot be manipulated via algorithm switching.  
**Preconditions:** Application running, valid user credentials.  
**Steps:**
1. Login and obtain a valid JWT access token.
2. Decode the JWT header and change `"alg": "HS256"` to `"alg": "none"`.
3. Remove the signature portion and send the modified token to a protected endpoint.
4. Repeat with `"alg": "RS256"` and a forged public key.

**Expected Result:** Server rejects all modified tokens with HTTP 401.  
**Actual Result:** ❌ Server uses HS256 but does not explicitly reject "none" algorithm tokens.  
**Status:** FAIL

---

### TC-AUTH-002: Empty/Whitespace Password Login
**Objective:** Verify that empty or whitespace-only passwords are rejected.  
**Preconditions:** Login endpoint accessible.  
**Steps:**
1. Send login request with `password: ""`.
2. Send login request with `password: "   "`.
3. Send login request with `password: "\t\n"`.

**Expected Result:** HTTP 400 with validation error.  
**Actual Result:** ⚠️ Stub implementation; real validation not yet present.  
**Status:** NOT TESTED (stub)

---

### TC-AUTH-003: Timing Attack on Password Comparison
**Objective:** Verify that password verification uses constant-time comparison.  
**Preconditions:** Login endpoint with actual password hashing implemented.  
**Steps:**
1. Measure response time for valid username + valid password.
2. Measure response time for valid username + invalid password.
3. Measure response time for invalid username + any password.
4. Compare timing distributions statistically.

**Expected Result:** Timing differences are statistically insignificant (<5ms variance).  
**Actual Result:** ⚠️ Stub implementation.  
**Status:** NOT TESTED (stub)

---

### TC-AUTH-004: Brute Force Protection Effectiveness
**Objective:** Verify that brute force attacks are mitigated.  
**Preconditions:** Login endpoint accessible, rate limiter enabled.  
**Steps:**
1. Send 15 login requests with invalid credentials within 1 minute from the same IP.
2. Verify that after 10 attempts, requests are blocked.
3. Wait for rate limit window to expire.
4. Attempt valid login after window expires.

**Expected Result:** Requests 11-15 return HTTP 429. Valid login succeeds after window.  
**Actual Result:** ✅ Rate limiter active (10 req/min for auth).  
**Status:** PASS

---

### TC-AUTH-005: MFA Bypass via Concurrent Requests
**Objective:** Verify that MFA cannot be bypassed via race conditions.  
**Preconditions:** MFA-enabled account.  
**Steps:**
1. Initiate login to get MFA token.
2. Send multiple simultaneous MFA verification requests with different codes.
3. Check if any request succeeds without valid MFA.

**Expected Result:** Only valid MFA code succeeds; others return 401.  
**Actual Result:** ⚠️ Stub implementation.  
**Status:** NOT TESTED (stub)

---

### TC-AUTH-006: Token Replay After Logout
**Objective:** Verify that tokens are invalidated after logout.  
**Preconditions:** Valid session, token revocation list enabled.  
**Steps:**
1. Login and obtain access token.
2. Call logout endpoint.
3. Attempt to use the same access token on a protected endpoint.

**Expected Result:** Token is rejected with HTTP 401.  
**Actual Result:** ❌ Logout handler deletes Redis sessions but does not revoke JWT tokens.  
**Status:** FAIL

---

## 2. Authorization Tests

### TC-AUTHZ-001: Horizontal Privilege Escalation (IDOR) - Host Access
**Objective:** Verify that users cannot access other users' hosts.  
**Preconditions:** Two user accounts with distinct hosts.  
**Steps:**
1. Login as User A, note host ID.
2. Login as User B.
3. As User B, send GET /hosts/{UserA-HostID}.

**Expected Result:** HTTP 403 Forbidden.  
**Actual Result:** ❌ HostHandler.Get() does not check ownership. Returns host data.  
**Status:** FAIL

---

### TC-AUTHZ-002: Horizontal Privilege Escalation - Host Update
**Objective:** Verify that users cannot modify other users' hosts.  
**Preconditions:** Two user accounts with distinct hosts.  
**Steps:**
1. Login as User A, create a host.
2. Login as User B.
3. As User B, send PATCH /hosts/{UserA-HostID} to modify the host.

**Expected Result:** HTTP 403 Forbidden.  
**Actual Result:** ❌ HostHandler.Update() does not verify ownership.  
**Status:** FAIL

---

### TC-AUTHZ-003: Horizontal Privilege Escalation - Host Delete
**Objective:** Verify that users cannot delete other users' hosts.  
**Preconditions:** Two user accounts with distinct hosts.  
**Steps:**
1. Login as User A, create a host.
2. Login as User B.
3. As User B, send DELETE /hosts/{UserA-HostID}.

**Expected Result:** HTTP 403 Forbidden.  
**Actual Result:** ❌ HostHandler.Delete() does not verify ownership.  
**Status:** FAIL

---

### TC-AUTHZ-004: Vertical Privilege Escalation - Admin Access
**Objective:** Verify that non-admin users cannot access admin endpoints.  
**Preconditions:** Regular user account and admin account.  
**Steps:**
1. Login as regular user (role: viewer).
2. Attempt to access GET /api/v1/admin/audit.
3. Attempt to access POST /api/v1/admin/audit/log.

**Expected Result:** HTTP 403 Forbidden for all admin endpoints.  
**Actual Result:** ✅ RequireRole middleware blocks access.  
**Status:** PASS

---

### TC-AUTHZ-005: Vertical Privilege Escalation - Role Manipulation
**Objective:** Verify that users cannot escalate their own role.  
**Preconditions:** Regular user account.  
**Steps:**
1. Login as regular user.
2. Attempt to modify user profile to change role to "admin".
3. Attempt to send crafted JWT with modified "role" claim.

**Expected Result:** Role change rejected; forged JWT rejected.  
**Actual Result:** ⚠️ No role update endpoint exists yet.  
**Status:** NOT TESTED (feature missing)

---

## 3. Input Validation Tests

### TC-INPUT-001: SQL Injection in Host Search
**Objective:** Verify that host search parameters are SQL-safe.  
**Preconditions:** Authenticated user.  
**Steps:**
1. Send GET /hosts?tag='; DROP TABLE hosts; --
2. Send GET /hosts?group=' OR '1'='1
3. Verify database integrity.

**Expected Result:** Parameters sanitized; database unaffected.  
**Actual Result:** ⚠️ Host list uses parameterized queries but tag/group use dynamic string building.  
**Status:** CONDITIONAL PASS (needs review)

---

### TC-INPUT-002: Command Injection via Terminal
**Objective:** Verify that terminal input cannot inject shell commands.  
**Preconditions:** Active WebSocket terminal session.  
**Steps:**
1. Open WebSocket terminal to an SSH host.
2. Send terminal data containing: `; rm -rf / ;`
3. Send terminal data containing: `$(whoami)`
4. Send terminal data containing: `` `id` ``

**Expected Result:** Commands sent as literal text to SSH server; no local shell execution.  
**Actual Result:** ✅ Data written directly to SSH stdin via Go SSH library. No shell execution.  
**Status:** PASS

---

### TC-INPUT-003: XSS in Terminal Output
**Objective:** Verify that terminal output cannot execute JavaScript in the browser.  
**Preconditions:** Active WebSocket terminal session.  
**Steps:**
1. On SSH server, create file with: `<script>alert('XSS')</script>`.
2. Display file contents: `cat /tmp/xss_test`.
3. Observe frontend rendering.

**Expected Result:** Script tags rendered as text, not executed.  
**Actual Result:** ❌ Terminal output embedded directly in JSON without sanitization.  
**Status:** FAIL

---

### TC-INPUT-004: LDAP Injection (if applicable)
**Objective:** Verify LDAP queries are parameterized.  
**Preconditions:** LDAP integration enabled.  
**Steps:**
1. Not applicable — no LDAP integration found.  

**Status:** N/A

---

### TC-INPUT-005: Path Traversal in SFTP (if applicable)
**Objective:** Verify SFTP paths are restricted.  
**Preconditions:** SFTP file transfer enabled.  
**Steps:**
1. Not applicable — SFTP not yet implemented.  

**Status:** N/A

---

### TC-INPUT-006: NoSQL Injection (if applicable)
**Objective:** Verify NoSQL queries are safe.  
**Preconditions:** MongoDB or similar in use.  
**Steps:**
1. Not applicable — PostgreSQL only.  

**Status:** N/A

---

## 4. Session Management Tests

### TC-SESS-001: Session Fixation
**Objective:** Verify session ID is regenerated after authentication.  
**Preconditions:** Login endpoint functional.  
**Steps:**
1. Access login page and note any pre-authentication session ID.
2. Login with valid credentials.
3. Check if session ID has changed.

**Expected Result:** New session ID issued after successful authentication.  
**Actual Result:** ⚠️ Stub login implementation.  
**Status:** NOT TESTED (stub)

---

### TC-SESS-002: Session Hijacking via Token Theft
**Objective:** Verify stolen tokens can be detected or invalidated.  
**Preconditions:** Valid session on Device A.  
**Steps:**
1. Login on Device A, capture session token.
2. Use token on Device B with different IP and User-Agent.
3. Observe if server detects anomaly.

**Expected Result:** Session rejected or flagged for suspicious activity.  
**Actual Result:** ❌ Session not bound to IP or device fingerprint.  
**Status:** FAIL

---

### TC-SESS-003: Session Timeout
**Objective:** Verify sessions expire after idle timeout.  
**Preconditions:** Valid session with 30-minute idle timeout.  
**Steps:**
1. Login and perform an action.
2. Wait 31 minutes without activity.
3. Attempt to access a protected endpoint.

**Expected Result:** HTTP 401 with session expired message.  
**Actual Result:** ⚠️ SessionStore checks idle timeout but only on explicit Get() call.  
**Status:** CONDITIONAL PASS

---

### TC-SESS-004: Concurrent Session Limit
**Objective:** Verify users cannot create unlimited sessions.  
**Preconditions:** Valid user account.  
**Steps:**
1. Login repeatedly to create 10+ sessions.
2. Verify if new sessions are blocked after limit.

**Expected Result:** New sessions blocked after configured limit (default: 5).  
**Actual Result:** ❌ No concurrent session limit implemented.  
**Status:** FAIL

---

### TC-SESS-005: Secure Session Cookie Attributes
**Objective:** Verify session cookies have security attributes.  
**Preconditions:** Web application serving cookies.  
**Steps:**
1. Login and inspect Set-Cookie headers.
2. Verify HttpOnly, Secure, SameSite attributes.

**Expected Result:** All three attributes present.  
**Actual Result:** ⚠️ No cookie-based sessions implemented; JWT in headers only.  
**Status:** NOT TESTED (architecture choice)

---

## 5. Cryptographic Tests

### TC-CRYPTO-001: TLS Version Enforcement
**Objective:** Verify only TLS 1.3 is accepted.  
**Preconditions:** HTTPS endpoint accessible.  
**Steps:**
1. Test connection with TLS 1.0.
2. Test connection with TLS 1.1.
3. Test connection with TLS 1.2.
4. Test connection with TLS 1.3.

**Expected Result:** Only TLS 1.3 succeeds.  
**Actual Result:** ⚠️ Server uses `ListenAndServe()` (HTTP), TLS not configured in code.  
**Status:** FAIL

---

### TC-CRYPTO-002: Weak Password Hashing
**Objective:** Verify passwords use strong hashing (Argon2id).  
**Preconditions:** User registration functional.  
**Steps:**
1. Register a new user.
2. Inspect password hash in database.
3. Verify algorithm and parameters.

**Expected Result:** Argon2id with memory ≥ 64MB, iterations ≥ 3.  
**Actual Result:** ⚠️ Stub implementation; bcrypt and Argon2 both present but not wired to auth flow.  
**Status:** NOT TESTED (stub)

---

### TC-CRYPTO-003: JWT Secret Strength
**Objective:** Verify JWT signing secrets are strong.  
**Preconditions:** Application running.  
**Steps:**
1. Inspect JWT_SECRET environment variable or config.
2. Verify length ≥ 32 bytes.
3. Verify randomness (not dictionary words).

**Expected Result:** Cryptographically random, ≥ 32 bytes.  
**Actual Result:** ❌ Fallback secret is a predictable string.  
**Status:** FAIL

---

### TC-CRYPTO-004: Encryption Key Management
**Objective:** Verify encryption keys are not hardcoded.  
**Preconditions:** Application source code access.  
**Steps:**
1. Search source code for hardcoded keys.
2. Verify keys are loaded from secure sources (KMS, environment).

**Expected Result:** No hardcoded keys.  
**Actual Result:** ❌ Fallback encryption key is hardcoded in config.  
**Status:** FAIL

---

### TC-CRYPTO-005: Secure Random Number Generation
**Objective:** Verify CSPRNG is used for all security tokens.  
**Preconditions:** Code review.  
**Steps:**
1. Review all random generation points (tokens, salts, nonces).
2. Verify use of `crypto/rand`, not `math/rand`.

**Expected Result:** All security-sensitive random uses `crypto/rand`.  
**Actual Result:** ✅ Uses `crypto/rand` throughout.  
**Status:** PASS

---

## 6. WebSocket Security Tests

### TC-WS-001: Cross-Site WebSocket Hijacking (CSWSH)
**Objective:** Verify WebSocket connections cannot be hijacked from malicious sites.  
**Preconditions:** WebSocket endpoint accessible.  
**Steps:**
1. Host malicious HTML page on attacker.com.
2. Include JavaScript that opens WebSocket to victim's /ws/terminal.
3. Check if connection succeeds without explicit authorization.

**Expected Result:** Connection rejected due to origin validation.  
**Actual Result:** ❌ `CheckOrigin` returns `true` for all origins.  
**Status:** FAIL

---

### TC-WS-002: WebSocket Message Flooding
**Objective:** Verify WebSocket rate limiting prevents DoS.  
**Preconditions:** Active WebSocket connection.  
**Steps:**
1. Open WebSocket terminal.
2. Send 1000+ messages per second.
3. Monitor server resource usage.

**Expected Result:** Rate limiter blocks excessive messages.  
**Actual Result:** ❌ No WebSocket rate limiting implemented.  
**Status:** FAIL

---

### TC-WS-003: Malformed Binary Frames
**Objective:** Verify server handles malformed frames gracefully.  
**Preconditions:** WebSocket endpoint accessible.  
**Steps:**
1. Send invalid UTF-8 in text frames.
2. Send oversized binary frames (>10MB).
3. Send frames with reserved opcodes.

**Expected Result:** Server closes connection cleanly with appropriate close code.  
**Actual Result:** ⚠️ gorilla/websocket handles some cases; application does not add custom limits.  
**Status:** CONDITIONAL PASS

---

### TC-WS-004: WebSocket Authentication Validation
**Objective:** Verify JWT is validated during WebSocket handshake.  
**Preconditions:** Valid JWT token.  
**Steps:**
1. Attempt WebSocket upgrade with valid JWT in query parameter.
2. Attempt WebSocket upgrade with expired JWT.
3. Attempt WebSocket upgrade with no JWT.

**Expected Result:** Only valid JWT succeeds.  
**Actual Result:** ❌ No JWT validation during WebSocket upgrade.  
**Status:** FAIL

---

### TC-WS-005: Terminal Escape Sequence Injection
**Objective:** Verify ANSI escape sequences cannot abuse terminal.  
**Preconditions:** Active terminal session.  
**Steps:**
1. Send escape sequences: `\x1b[2J` (clear screen), `\x1b]0;malicious\x07` (set title).
2. Verify frontend terminal behavior.

**Expected Result:** Dangerous escape sequences filtered or sanitized.  
**Actual Result:** ❌ No escape sequence filtering.  
**Status:** FAIL

---

## 7. Credential Vault Tests

### TC-VAULT-001: Vault Lock After Inactivity
**Objective:** Verify vault locks after configured idle time.  
**Preconditions:** Vault unlocked.  
**Steps:**
1. Unlock vault with master password.
2. Wait for idle timeout (default: 15 minutes).
3. Attempt to encrypt a credential.

**Expected Result:** Vault is locked; encryption fails.  
**Actual Result:** ❌ No auto-lock timer implemented.  
**Status:** FAIL

---

### TC-VAULT-002: Master Password Brute Force
**Objective:** Verify vault unlock has brute force protection.  
**Preconditions:** Vault locked.  
**Steps:**
1. Attempt to unlock vault with 100 incorrect passwords.
2. Verify rate limiting or account lockout.

**Expected Result:** Attempts throttled or blocked after threshold.  
**Actual Result:** ❌ No brute force protection on vault unlock.  
**Status:** FAIL

---

### TC-VAULT-003: Memory Security of Decrypted Credentials
**Objective:** Verify decrypted credentials are securely zeroed from memory.  
**Preconditions:** Code review.  
**Steps:**
1. Review vault decryption code.
2. Verify `SecureZero` is called on plaintext after use.

**Expected Result:** All plaintext buffers zeroed after use.  
**Actual Result:** ✅ `SecureZero` exists and is used in `Lock()`.  
**Status:** PASS

---

### TC-VAULT-004: Key Rotation Integrity
**Objective:** Verify key rotation preserves data integrity.  
**Preconditions:** Vault with existing encrypted credentials.  
**Steps:**
1. Encrypt credential with old key.
2. Rotate to new key.
3. Decrypt credential with new key.
4. Verify data matches original.

**Expected Result:** Data intact after rotation.  
**Actual Result:** ✅ Tests pass for key rotation.  
**Status:** PASS

---

### TC-VAULT-005: Salt Uniqueness
**Objective:** Verify each encryption operation uses unique salt.  
**Preconditions:** Code review.  
**Steps:**
1. Call `EncryptCredential` multiple times.
2. Verify salts are unique across operations.

**Expected Result:** No salt reuse observed.  
**Actual Result:** ✅ `generateRandomBytes(32)` generates unique salts.  
**Status:** PASS

---

## 8. File Upload/Download Tests

### TC-FILE-001: Malicious File Upload
**Objective:** Verify file uploads reject executable content.  
**Preconditions:** File upload endpoint available.  
**Steps:**
1. Not applicable — no file upload endpoint found.  

**Status:** N/A

---

### TC-FILE-002: Path Traversal in Download
**Objective:** Verify download paths cannot escape allowed directories.  
**Preconditions:** File download endpoint available.  
**Steps:**
1. Not applicable — no file download endpoint found.  

**Status:** N/A

---

## 9. Rate Limiting Tests

### TC-RATE-001: API Rate Limit Enforcement
**Objective:** Verify general API rate limiting is effective.  
**Preconditions:** Application running, Redis available.  
**Steps:**
1. Send 101 authenticated requests within 1 minute.
2. Verify response code changes to 429.
3. Verify Retry-After header.

**Expected Result:** HTTP 429 after 100 requests.  
**Actual Result:** ✅ Rate limiter blocks at 100 req/min.  
**Status:** PASS

---

### TC-RATE-002: Auth Rate Limit Bypass
**Objective:** Verify auth rate limiting cannot be bypassed.  
**Preconditions:** Login endpoint accessible.  
**Steps:**
1. Send 11 login attempts within 1 minute from same IP.
2. Rotate source IP and repeat.
3. Check if distributed rate limiting catches rotated IPs.

**Expected Result:** All IPs limited independently; 11th attempt blocked.  
**Actual Result:** ⚠️ Uses `c.ClientIP()` which may be spoofed behind proxy.  
**Status:** CONDITIONAL PASS

---

### TC-RATE-003: Rate Limit Header Information
**Objective:** Verify rate limit headers inform clients correctly.  
**Preconditions:** Rate limited endpoint.  
**Steps:**
1. Send requests until rate limit triggered.
2. Inspect response headers for rate limit info.

**Expected Result:** `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset` headers present.  
**Actual Result:** ❌ No rate limit headers returned.  
**Status:** FAIL

---

## 10. CSRF Tests

### TC-CSRF-001: CSRF Token Validation
**Objective:** Verify CSRF tokens are validated on state-changing requests.  
**Preconditions:** Authenticated session.  
**Steps:**
1. Obtain valid CSRF token from server.
2. Send POST request with valid token.
3. Send POST request with invalid token.
4. Send POST request with no token.

**Expected Result:** Valid token succeeds; invalid/missing tokens return 403.  
**Actual Result:** ❌ CSRF middleware not applied to routes; only checks token presence.  
**Status:** FAIL

---

### TC-CSRF-002: Cross-Site POST Without Token
**Objective:** Verify cross-site POST requests are blocked.  
**Preconditions:** Authenticated session with cookies.  
**Steps:**
1. Create malicious HTML form on attacker.com.
2. Form submits POST to victim.com/api/v1/hosts/delete.
3. Trick victim into submitting form.

**Expected Result:** Request blocked due to missing CSRF token.  
**Actual Result:** ❌ No CSRF protection on API endpoints.  
**Status:** FAIL

---

### TC-CSRF-003: SameSite Cookie Protection
**Objective:** Verify cookies have SameSite attribute.  
**Preconditions:** Cookie-based session implemented.  
**Steps:**
1. Inspect Set-Cookie headers.
2. Verify SameSite=Strict or SameSite=Lax.

**Expected Result:** SameSite attribute present.  
**Actual Result:** ⚠️ No cookies used; JWT in Authorization header.  
**Status:** N/A (architecture uses Bearer tokens)

---

## Appendix A: Test Execution Summary

| Category | Total | Pass | Fail | Not Tested | N/A |
|----------|-------|------|------|------------|-----|
| Authentication | 6 | 1 | 2 | 3 | 0 |
| Authorization | 5 | 1 | 3 | 1 | 0 |
| Input Validation | 6 | 1 | 2 | 0 | 3 |
| Session Management | 5 | 0 | 3 | 2 | 0 |
| Cryptographic | 5 | 1 | 3 | 1 | 0 |
| WebSocket Security | 5 | 0 | 4 | 1 | 0 |
| Credential Vault | 5 | 2 | 2 | 1 | 0 |
| File Upload/Download | 2 | 0 | 0 | 0 | 2 |
| Rate Limiting | 3 | 1 | 2 | 0 | 0 |
| CSRF | 3 | 0 | 2 | 0 | 1 |
| **TOTAL** | **45** | **7** | **24** | **10** | **7** |

---

**Document History**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-05-27 | QA Subagent | Initial security test cases |
