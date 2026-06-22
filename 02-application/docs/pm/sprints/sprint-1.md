# Sprint 1 — Core Backend (Part 1)

> **Phase:** 1 | **Weeks:** 3–4 | **Duration:** 10 working days  
> **Sprint Goal:** Build authentication system, database layer, and API scaffolding with security-first design  
> **Story Points:** 55  
> **Status:** Not Started  
> **Depends On:** Sprint 0 (Foundation)

---

## User Stories

### US-006: User Authentication System
**As a** user, **I want** secure login with email and password, **so that** my account is protected.

**Acceptance Criteria:**
- [ ] User registration with email validation
- [ ] Password hashing with Argon2id (m=65536, t=3, p=4)
- [ ] Login with JWT access token (15 min expiry) and refresh token (7 days)
- [ ] Logout invalidates all tokens
- [ ] Password strength validation (min 12 chars, complexity requirements)
- [ ] Account lockout after 5 failed attempts (15 min)
- [ ] Email verification required before first login

**Definition of Done:**
- [ ] All auth endpoints return appropriate HTTP status codes
- [ ] Password hashes cannot be reversed
- [ ] Unit tests cover all auth flows (>90% coverage)
- [ ] Security review completed

**Story Points:** 13

---

### US-007: Multi-Factor Authentication (MFA)
**As a** security-conscious user, **I want** MFA support, **so that** my account has additional protection.

**Acceptance Criteria:**
- [ ] TOTP-based MFA (RFC 6238)
- [ ] QR code generation for authenticator apps
- [ ] Backup recovery codes (10 codes, single-use)
- [ ] MFA can be enabled/disabled in settings
- [ ] API requires MFA code after password validation when enabled
- [ ] Backup codes regenerated on MFA re-enable

**Definition of Done:**
- [ ] MFA works with Google Authenticator, Authy, 1Password
- [ ] Backup codes can recover account if device lost
- [ ] Unit tests for MFA verification logic
- [ ] Rate limiting on MFA verification endpoint

**Story Points:** 13

---

### US-008: Role-Based Access Control (RBAC)
**As an** administrator, **I want** role-based permissions, **so that** users have appropriate access levels.

**Acceptance Criteria:**
- [ ] Three roles: admin, operator, viewer
- [ ] Admin: full CRUD on users, hosts, credentials, audit logs
- [ ] Operator: CRUD on hosts and credentials, read audit logs
- [ ] Viewer: read-only access to hosts and credentials
- [ ] Permission checks on every API endpoint
- [ ] Role assignment during user creation

**Definition of Done:**
- [ ] Unauthorized requests return 403 with clear error message
- [ ] Role changes take effect immediately (no re-login required)
- [ ] Audit log records all permission changes
- [ ] Integration tests for each role's access boundaries

**Story Points:** 8

---

### US-009: Database Schema & Migrations
**As a** developer, **I want** a robust database schema with migrations, **so that** data integrity is maintained.

**Acceptance Criteria:**
- [ ] Migration 001: Users, devices, credentials, hosts, audit_log tables
- [ ] Migration 002: Sessions and vault_keys tables
- [ ] Migration 003: Audit log integrity functions
- [ ] Foreign key constraints with CASCADE rules
- [ ] Soft deletes (deleted_at timestamp)
- [ ] Auto-update triggers for updated_at
- [ ] Database seed data for development

**Definition of Done:**
- [ ] Migrations run forward and backward without errors
- [ ] Schema validated against design document
- [ ] Integration tests verify foreign key behavior
- [ ] Performance baseline established for common queries

**Story Points:** 8

---

### US-010: API Scaffolding & Middleware
**As a** frontend developer, **I want** a consistent API structure, **so that** integration is straightforward.

**Acceptance Criteria:**
- [ ] Health check endpoint (`GET /health`)
- [ ] Request ID middleware for tracing
- [ ] Structured logging (JSON format, correlation IDs)
- [ ] Error handling with consistent error responses
- [ ] Request validation middleware
- [ ] CORS configuration
- [ ] OpenAPI 3.0 spec with Swagger UI

**Definition of Done:**
- [ ] All endpoints return consistent JSON structure
- [ ] Error responses include error code, message, and details
- [ ] API documentation is auto-generated from code
- [ ] Load testing: 1000 req/s with < 100ms p95 latency

**Story Points:** 13

---

## Dependencies

| Dependency | Source | Impact |
|------------|--------|--------|
| Sprint 0 completion | Previous sprint | Blocks all development |
| PostgreSQL 16 | Infrastructure | Blocks database work |
| JWT library choice | Tech decision | Blocks auth implementation |
| Argon2 parameters | Security review | Blocks password hashing |

---

## Risks and Mitigations

| Risk | Probability | Impact | Mitigation | Owner |
|------|-------------|--------|------------|-------|
| Argon2 performance issues | Medium | High | Benchmark with production-like user load | Backend Lead |
| JWT secret management | Medium | Critical | Use Vault for secret storage, rotate quarterly | Security |
| Database migration failures | Low | High | Test migrations on copy of production data | DBA |
| RBAC complexity grows | Medium | Medium | Keep roles simple (3 only), document permissions matrix | Tech Lead |

---

## Technical Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Password hashing | Argon2id | Winner of Password Hashing Competition, memory-hard |
| JWT library | `golang-jwt/jwt/v5` | Most popular Go JWT library, actively maintained |
| TOTP library | `pquerna/otp` | Implements RFC 6238/4226, widely used |
| Database | PostgreSQL 16 | ACID compliance, excellent Go driver, JSON support |
| Migration tool | `golang-migrate` | Industry standard, CLI and Go API |

---

## Sprint Review Checklist

- [ ] Auth endpoints tested with 1000 concurrent users
- [ ] MFA verified with 3+ authenticator apps
- [ ] RBAC tested for all 3 roles
- [ ] Database migrations run successfully on staging
- [ ] API documentation reviewed by frontend team
- [ ] Security review signed off
- [ ] Sprint retrospective completed

---

## Sprint Retrospective Notes

*To be filled after sprint completion*

**What went well:**
- 

**What could be improved:**
- 

**Action items for next sprint:**
- 
