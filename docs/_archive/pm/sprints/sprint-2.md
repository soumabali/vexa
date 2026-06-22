# Sprint 2 — Core Backend (Part 2)

> **Phase:** 1 | **Weeks:** 5–6 | **Duration:** 10 working days  
> **Sprint Goal:** Implement protocol gateway, session management, and credential vault with encryption  
> **Story Points:** 55  > **Status:** Not Started  
> **Depends On:** Sprint 1 (Auth + Database)

---

## User Stories

### US-011: Protocol Gateway (SSH/RDP/VNC)
**As a** user, **I want** to connect to remote servers via SSH, RDP, and VNC, **so that** I can manage heterogeneous infrastructure.

**Acceptance Criteria:**
- [ ] SSH proxy using `golang.org/x/crypto/ssh`
- [ ] RDP proxy using FreeRDP subprocess
- [ ] VNC proxy using `libvncclient` subprocess
- [ ] Connection multiplexing (multiple sessions per user)
- [ ] Protocol auto-detection from host configuration
- [ ] Graceful error handling for connection failures
- [ ] Timeout configuration per protocol (default: 30s)

**Definition of Done:**
- [ ] Can connect to SSH server (password + key auth)
- [ ] Can connect to RDP server (NLA + standard)
- [ ] Can connect to VNC server (standard auth)
- [ ] Failed connections return descriptive error messages
- [ ] Load test: 100 concurrent sessions

**Story Points:** 13

---

### US-012: Session Management
**As a** user, **I want** persistent session tracking, **so that** I can resume or review past connections.

**Acceptance Criteria:**
- [ ] Session start/end logging with timestamps
- [ ] Active session list with real-time status
- [ ] Session recording (text for SSH, optional video for RDP/VNC)
- [ ] Session timeout (configurable, default: 30 min idle)
- [ ] Force disconnect capability
- [ ] Session metadata: bytes sent/received, duration, client IP

**Definition of Done:**
- [ ] Sessions survive gateway restart (stored in Redis)
- [ ] Recording files stored securely with restricted access
- [ ] Audit log entries for all session events
- [ ] GDPR-compliant data retention (auto-delete after 90 days)

**Story Points:** 13

---

### US-013: Credential Vault
**As a** user, **I want** encrypted credential storage, **so that** my passwords and keys are secure.

**Acceptance Criteria:**
- [ ] Master password derived key using Argon2id
- [ ] AES-256-GCM encryption for all credentials
- [ ] Vault lock after configurable timeout (default: 5 min)
- [ ] Credential types: password, SSH key, API token
- [ ] Credential organization with tags and groups
- [ ] Secure export/import (encrypted JSON)

**Definition of Done:**
- [ ] Vault encryption verified by security review
- [ ] Memory securely cleared after vault lock
- [ ] No plaintext credentials in logs or memory dumps
- [ ] Key rotation capability (re-encrypt all credentials)

**Story Points:** 13

---

### US-014: Audit Logging
**As an** administrator, **I want** comprehensive audit logs, **so that** all actions are traceable.

**Acceptance Criteria:**
- [ ] All CRUD operations logged with user, timestamp, IP
- [ ] Immutable audit log with integrity hash chain
- [ ] Log categories: auth, host, credential, session, admin
- [ ] Query interface with filters (user, date range, action)
- [ ] Export to JSON/CSV
- [ ] Tamper detection (hash chain verification)

**Definition of Done:**
- [ ] Audit log cannot be modified by application users
- [ ] Query performance: < 100ms for last 30 days
- [ ] Integrity verification passes for all records
- [ ] Retention policy: 1 year active, 7 years archive

**Story Points:** 8

---

### US-015: Host Management API
**As a** user, **I want** CRUD operations for hosts, **so that** I can organize my server inventory.

**Acceptance Criteria:**
- [ ] Create host with name, address, protocol, port, tags
- [ ] Read host list with filtering and pagination
- [ ] Update host details
- [ ] Soft delete host (recoverable)
- [ ] Host grouping with hierarchy
- [ ] Health check endpoint (ping + port check)

**Definition of Done:**
- [ ] All endpoints protected by RBAC
- [ ] Input validation prevents injection attacks
- [ ] Pagination handles 10,000+ hosts efficiently
- [ ] Audit log records all host changes

**Story Points:** 8

---

## Dependencies

| Dependency | Source | Impact |
|------------|--------|--------|
| Sprint 1 completion | Previous sprint | Blocks auth-protected endpoints |
| FreeRDP binary | System package | Blocks RDP proxy |
| libvnc | System package | Blocks VNC proxy |
| Redis 7 | Infrastructure | Blocks session persistence |

---

## Risks and Mitigations

| Risk | Probability | Impact | Mitigation | Owner |
|------|-------------|--------|------------|-------|
| Protocol proxy security | High | Critical | Sandboxed subprocess, seccomp, network namespace | Security |
| Session recording storage | Medium | High | Compressed storage, auto-cleanup after retention | DevOps |
| Vault key loss | Low | Critical | Encrypted backup, recovery codes | Security |
| Performance with 1000+ hosts | Medium | Medium | Pagination, indexing, connection pooling | Backend |

---

## Technical Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| SSH library | `golang.org/x/crypto/ssh` | Official Go crypto library, maintained by Google |
| RDP implementation | FreeRDP subprocess | Full RDP protocol support, sandboxed |
| VNC implementation | `libvncclient` subprocess | Lightweight, well-tested |
| Session storage | Redis | Fast in-memory, TTL support, persistence |
| Vault encryption | AES-256-GCM | Authenticated encryption, Go stdlib support |

---

## Sprint Review Checklist

- [ ] Protocol connections tested with real servers
- [ ] Session recording playback verified
- [ ] Vault encryption audited by security
- [ ] Audit log integrity verified
- [ ] Host CRUD tested with 1000 hosts
- [ ] Performance tests pass (p95 < 100ms)
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
