# vexa — Database Schema Documentation

> **Agent:** Database Architect  
> **Status:** Generated 2026-05-28  
> **Version:** 0.1.0  
> **Last Updated:** 2026-05-28

---

## 1. Overview

The vexa database is designed around PostgreSQL 16 with the following design principles:

- **Security-first:** All credentials encrypted at rest, audit logs immutable
- **Relational integrity:** Foreign keys with CASCADE rules for data consistency
- **Performance:** Strategic indexes for common query patterns
- **Scalability:** Partitioning-ready for audit logs, soft deletes with `deleted_at`
- **Auditability:** Every table has `created_at`, `updated_at`, `created_by`, `updated_by`

---

## 2. Entity Relationship Diagram (Textual)

```
┌──────────────────────────────────────────────────────────────┐
│                         CORE TABLES                            │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐   │
│  │   users     │◄────┤   devices   │     │   sessions  │   │
│  │─────────────│     │─────────────│     │─────────────│   │
│  │ id (PK)     │     │ id (PK)     │     │ id (PK)     │   │
│  │ email (UQ)  │     │ user_id(FK) │     │ user_id(FK) │   │
│  │ password_hash│    │ fingerprint │     │ token       │   │
│  │ role        │     │ last_seen   │     │ expires_at  │   │
│  │ mfa_secret  │     └─────────────┘     └─────────────┘   │
│  │ is_active   │                                            │
│  └──────┬──────┘                                            │
│         │                                                     │
│         │ 1:N                                                 │
│         ▼                                                     │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐   │
│  │ credentials │◄────┤  hosts      │◄────┤ host_groups │   │
│  │─────────────│     │─────────────│     │─────────────│   │
│  │ id (PK)     │     │ id (PK)     │     │ id (PK)     │   │
│  │ user_id(FK) │     │ group_id(FK)│     │ parent_id   │   │
│  │ type        │     │ name        │     │ name        │   │
│  │ encrypted_data│    │ hostname    │     │ path (ltree)│  │
│  │ iv          │     │ port        │     └─────────────┘   │
│  └──────┬──────┘     │ protocol    │                       │
│         │            └─────────────┘                       │
│         │                                                   │
│         │ 1:N                                               │
│         ▼                                                   │
│  ┌─────────────┐     ┌─────────────┐                       │
│  │connection_  │     │  audit_log  │                       │
│  │   history   │     │─────────────│                       │
│  │─────────────│     │ id (PK)     │                       │
│  │ id (PK)     │     │ user_id(FK) │                       │
│  │ host_id(FK) │     │ action      │                       │
│  │ credential_ │     │ entity_type │                       │
│  │   id (FK)   │     │ entity_id   │                       │
│  │ started_at  │     │ old_value   │                       │
│  │ ended_at    │     │ new_value   │                       │
│  │ bytes_sent  │     │ integrity_hash│                      │
│  │ bytes_recv  │     └─────────────┘                       │
│  └─────────────┘                                            │
│                                                              │
└──────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────┐
│                    SECURITY TABLES                           │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────┐     ┌─────────────────┐                    │
│  │ vault_keys  │     │credential_shares│                   │
│  │─────────────│     │─────────────────│                   │
│  │ id (PK)     │     │ id (PK)         │                   │
│  │ user_id(FK) │     │ credential_id   │                   │
│  │ master_key  │     │ shared_with     │                   │
│  │ key_version │     │ encrypted_copy  │                   │
│  │ created_at  │     │ expires_at      │                   │
│  └─────────────┘     └─────────────────┘                   │
│                                                              │
│  ┌─────────────────┐                                        │
│  │master_key_versions│                                      │
│  │─────────────────│                                        │
│  │ id (PK)         │                                        │
│  │ version         │                                        │
│  │ reason          │                                        │
│  │ rotated_at      │                                        │
│  └─────────────────┘                                        │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

---

## 3. Table Definitions

### 3.1 users

**Purpose:** Authentication and authorization master table

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | UUID | PRIMARY KEY | Unique identifier |
| `email` | VARCHAR(255) | UNIQUE, NOT NULL | Login email |
| `password_hash` | VARCHAR(255) | NOT NULL | Argon2id hash |
| `display_name` | VARCHAR(100) | | User display name |
| `role` | VARCHAR(20) | DEFAULT 'viewer' | admin, operator, viewer |
| `mfa_secret` | VARCHAR(255) | | TOTP secret (encrypted) |
| `mfa_enabled` | BOOLEAN | DEFAULT FALSE | MFA active flag |
| `email_verified` | BOOLEAN | DEFAULT FALSE | Email verification |
| `is_active` | BOOLEAN | DEFAULT TRUE | Soft delete flag |
| `last_login_at` | TIMESTAMPTZ | | Last successful login |
| `failed_login_count` | INTEGER | DEFAULT 0 | Brute force protection |
| `locked_until` | TIMESTAMPTZ | | Account lockout |
| `created_at` | TIMESTAMPTZ | DEFAULT NOW() | Creation timestamp |
| `updated_at` | TIMESTAMPTZ | DEFAULT NOW() | Last update |
| `deleted_at` | TIMESTAMPTZ | | Soft delete timestamp |

**Indexes:**
- `idx_users_email` (email) — login lookups
- `idx_users_role` (role) — RBAC queries
- `idx_users_active` (is_active) — active user filtering
- `idx_users_deleted_at` (deleted_at) — soft delete queries

**Triggers:**
- `trg_users_updated_at` — auto-update `updated_at` on modify
- `trg_users_audit` — insert audit_log entry on change

---

### 3.2 devices

**Purpose:** Track registered devices for session management

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | UUID | PRIMARY KEY | Unique identifier |
| `user_id` | UUID | FK → users.id, ON DELETE CASCADE | Device owner |
| `device_name` | VARCHAR(100) | | Human-readable name |
| `device_type` | VARCHAR(20) | | desktop, mobile, web |
| `fingerprint` | VARCHAR(255) | UNIQUE | Device fingerprint |
| `public_key` | TEXT | | Device public key (for mTLS) |
| `last_ip` | INET | | Last known IP |
| `last_seen` | TIMESTAMPTZ | | Last activity |
| `is_trusted` | BOOLEAN | DEFAULT FALSE | Trusted device flag |
| `created_at` | TIMESTAMPTZ | DEFAULT NOW() | Registration date |

**Indexes:**
- `idx_devices_user` (user_id) — user device lookups
- `idx_devices_fingerprint` (fingerprint) — device identification

---

### 3.3 credentials

**Purpose:** Encrypted credential storage

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | UUID | PRIMARY KEY | Unique identifier |
| `user_id` | UUID | FK → users.id, ON DELETE CASCADE | Owner |
| `host_id` | UUID | FK → hosts.id, ON DELETE SET NULL | Associated host |
| `name` | VARCHAR(100) | NOT NULL | Credential label |
| `type` | VARCHAR(20) | NOT NULL | password, ssh_key, api_token |
| `encrypted_data` | BYTEA | NOT NULL | AES-256-GCM encrypted |
| `iv` | BYTEA | NOT NULL | Initialization vector |
| `auth_tag` | BYTEA | NOT NULL | GCM authentication tag |
| `key_version` | INTEGER | DEFAULT 1 | Encryption key version |
| `created_at` | TIMESTAMPTZ | DEFAULT NOW() | Creation timestamp |
| `updated_at` | TIMESTAMPTZ | DEFAULT NOW() | Last update |
| `deleted_at` | TIMESTAMPTZ | | Soft delete |

**Indexes:**
- `idx_credentials_user` (user_id) — user credential listings
- `idx_credentials_host` (host_id) — host credential lookup
- `idx_credentials_type` (type) — type filtering

**Security Notes:**
- Data encrypted with user-specific key derived from master key
- Never store plaintext passwords
- `key_version` enables re-encryption on rotation

---

### 3.4 hosts

**Purpose:** Server/connection target definitions

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | UUID | PRIMARY KEY | Unique identifier |
| `user_id` | UUID | FK → users.id, ON DELETE CASCADE | Owner |
| `group_id` | UUID | FK → host_groups.id, ON DELETE SET NULL | Organization |
| `name` | VARCHAR(100) | NOT NULL | Display name |
| `hostname` | VARCHAR(255) | NOT NULL | IP or domain |
| `port` | INTEGER | DEFAULT 22 | Connection port |
| `protocol` | VARCHAR(10) | DEFAULT 'ssh' | ssh, rdp, vnc |
| `username` | VARCHAR(100) | | Login username |
| `credential_id` | UUID | FK → credentials.id | Default credential |
| `tags` | TEXT[] | | Array of tags |
| `jump_host_id` | UUID | FK → hosts.id | Bastion/jump host |
| `notes` | TEXT | | User notes |
| `is_favorite` | BOOLEAN | DEFAULT FALSE | Favorite flag |
| `created_at` | TIMESTAMPTZ | DEFAULT NOW() | Creation |
| `updated_at` | TIMESTAMPTZ | DEFAULT NOW() | Last update |
| `deleted_at` | TIMESTAMPTZ | | Soft delete |

**Indexes:**
- `idx_hosts_user` (user_id) — user host listings
- `idx_hosts_group` (group_id) — group filtering
- `idx_hosts_protocol` (protocol) — protocol filtering
- `idx_hosts_tags` USING GIN (tags) — tag search
- `idx_hosts_favorite` (is_favorite) — favorite filtering

---

### 3.5 host_groups

**Purpose:** Hierarchical organization of hosts

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | UUID | PRIMARY KEY | Unique identifier |
| `user_id` | UUID | FK → users.id, ON DELETE CASCADE | Owner |
| `parent_id` | UUID | FK → host_groups.id, ON DELETE CASCADE | Parent group |
| `name` | VARCHAR(100) | NOT NULL | Group name |
| `path` | LTREE | | Hierarchical path |
| `color` | VARCHAR(7) | | UI color (#RRGGBB) |
| `icon` | VARCHAR(50) | | UI icon name |
| `created_at` | TIMESTAMPTZ | DEFAULT NOW() | Creation |
| `updated_at` | TIMESTAMPTZ | DEFAULT NOW() | Last update |

**Indexes:**
- `idx_host_groups_user` (user_id)
- `idx_host_groups_path` USING GIST (path) — tree queries
- `idx_host_groups_parent` (parent_id)

**LTREE Operations:**
```sql
-- Get all descendants
SELECT * FROM host_groups WHERE path <@ 'root.webservers';

-- Get ancestors
SELECT * FROM host_groups WHERE path @> 'root.webservers.prod';

-- Get immediate children
SELECT * FROM host_groups WHERE path ~ 'root.webservers.{1}';
```

---

### 3.6 sessions

**Purpose:** Active and historical session tracking

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | UUID | PRIMARY KEY | Unique identifier |
| `user_id` | UUID | FK → users.id, ON DELETE CASCADE | Session owner |
| `host_id` | UUID | FK → hosts.id, ON DELETE SET NULL | Target host |
| `credential_id` | UUID | FK → credentials.id | Used credential |
| `token` | VARCHAR(255) | UNIQUE, NOT NULL | Session JWT |
| `type` | VARCHAR(20) | | ssh, rdp, vnc |
| `status` | VARCHAR(20) | DEFAULT 'active' | active, closed, error |
| `started_at` | TIMESTAMPTZ | DEFAULT NOW() | Session start |
| `ended_at` | TIMESTAMPTZ | | Session end |
| `bytes_sent` | BIGINT | DEFAULT 0 | Traffic sent |
| `bytes_received` | BIGINT | DEFAULT 0 | Traffic received |
| `recording_path` | VARCHAR(500) | | Session recording file |
| `client_ip` | INET | | Client IP |
| `user_agent` | TEXT | | Client user agent |
| `created_at` | TIMESTAMPTZ | DEFAULT NOW() | Creation |

**Indexes:**
- `idx_sessions_user` (user_id) — user session history
- `idx_sessions_host` (host_id) — host session history
- `idx_sessions_token` (token) — session validation
- `idx_sessions_status` (status) — active session filtering
- `idx_sessions_started` (started_at DESC) — recent sessions

---

### 3.7 audit_log

**Purpose:** Immutable audit trail (append-only)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | BIGSERIAL | PRIMARY KEY | Auto-increment |
| `timestamp` | TIMESTAMPTZ | DEFAULT NOW() | Event time |
| `user_id` | UUID | FK → users.id | Actor |
| `action` | VARCHAR(50) | NOT NULL | create, update, delete, login, etc. |
| `entity_type` | VARCHAR(50) | NOT NULL | users, hosts, credentials, sessions |
| `entity_id` | UUID | | Affected entity |
| `old_value` | JSONB | | Previous state |
| `new_value` | JSONB | | New state |
| `ip_address` | INET | | Source IP |
| `user_agent` | TEXT | | Client info |
| `session_id` | UUID | | Session reference |
| `integrity_hash` | VARCHAR(64) | NOT NULL | SHA-256 chain hash |
| `previous_hash` | VARCHAR(64) | | Previous record hash |

**Indexes:**
- `idx_audit_user` (user_id) — user activity
- `idx_audit_entity` (entity_type, entity_id) — entity history
- `idx_audit_action` (action) — action filtering
- `idx_audit_timestamp` (timestamp DESC) — time range queries

**Partitioning:**
- Monthly partitions: `audit_log_2026_01`, `audit_log_2026_02`, etc.
- Auto-created by trigger on insert

**Integrity:**
```sql
-- Verify chain integrity
SELECT 
  id,
  integrity_hash = encode(
    digest(
      previous_hash || timestamp::text || action || entity_type || 
      COALESCE(old_value::text, '') || COALESCE(new_value::text, ''),
      'sha256'
    ),
    'hex'
  ) as is_valid
FROM audit_log
ORDER BY id;
```

---

### 3.8 vault_keys

**Purpose:** Per-user encryption key management

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | UUID | PRIMARY KEY | Unique identifier |
| `user_id` | UUID | FK → users.id, ON DELETE CASCADE | Key owner |
| `encrypted_master_key` | BYTEA | NOT NULL | Master key (encrypted with user password) |
| `key_version` | INTEGER | DEFAULT 1 | Key generation |
| `created_at` | TIMESTAMPTZ | DEFAULT NOW() | Creation |
| `rotated_at` | TIMESTAMPTZ | | Last rotation |

**Indexes:**
- `idx_vault_keys_user` (user_id) — key lookup

---

### 3.9 credential_shares

**Purpose:** Secure credential sharing between users

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | UUID | PRIMARY KEY | Unique identifier |
| `credential_id` | UUID | FK → credentials.id, ON DELETE CASCADE | Shared credential |
| `shared_by` | UUID | FK → users.id | Owner |
| `shared_with` | UUID | FK → users.id | Recipient |
| `encrypted_copy` | BYTEA | NOT NULL | Re-encrypted for recipient |
| `expires_at` | TIMESTAMPTZ | | Share expiry |
| `permissions` | VARCHAR(20) | DEFAULT 'read' | read, use |
| `created_at` | TIMESTAMPTZ | DEFAULT NOW() | Creation |

**Indexes:**
- `idx_shares_owner` (shared_by) — owner shares
- `idx_shares_recipient` (shared_with) — received shares
- `idx_shares_expiry` (expires_at) — cleanup queries

---

### 3.10 master_key_versions

**Purpose:** Track master key rotations for audit

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | UUID | PRIMARY KEY | Unique identifier |
| `version` | INTEGER | NOT NULL | Key version number |
| `reason` | TEXT | | Rotation reason |
| `rotated_by` | UUID | FK → users.id | Initiator |
| `rotated_at` | TIMESTAMPTZ | DEFAULT NOW() | Rotation time |
| `previous_key_hash` | VARCHAR(64) | | Hash of old key |

---

## 4. Migration History

### Migration 001: Initial Schema
**File:** `001_initial.sql`
**Tables Created:** users, devices, credentials, host_groups, hosts, connection_history, audit_log
**Key Features:**
- Base schema with all core tables
- Foreign key relationships
- Initial indexes
- Audit triggers

### Migration 002: Sessions & Vault
**File:** `002_sessions_and_vault_keys.sql`
**Tables Created:** sessions, vault_keys
**Key Features:**
- Session tracking for active connections
- Vault key storage per user
- Session-credential relationship

### Migration 003: Audit Integrity
**File:** `003_audit_log_integrity.sql`
**Key Features:**
- Integrity hash functions
- Auto-partition trigger for audit_log
- Hash chain verification

### Migration 004: Host Group Hierarchy
**File:** `004_host_groups_hierarchy.sql`
**Key Features:**
- LTREE extension for hierarchical data
- Path column with GiST index
- Hierarchy helper functions

### Migration 005: Query Optimization
**File:** `005_final_indexes.sql`
**Key Features:**
- Additional indexes for performance
- Covering indexes for common queries

### Migration 006: Credential Sharing
**File:** `006_credential_sharing.sql`
**Tables Created:** credential_shares, master_key_versions
**Key Features:**
- Cross-user credential sharing
- Master key rotation tracking
- Permission-based sharing

---

## 5. Seed Data

### Development Seeds

**File:** `seed_001_users_devices.sql`
- 4 test users (admin, operator, viewer, inactive)
- 2 test devices per user

**File:** `seed_002_hosts_credentials.sql`
- Sample hosts (SSH, RDP, VNC)
- Sample credentials for each host
- Host group hierarchy

**File:** `seed_003_sessions_history_audit.sql`
- Historical session records
- Sample audit log entries
- Connection history

**⚠️ WARNING:** Seed data must NOT be used in production. Change all passwords.

---

## 6. Query Patterns

### Common Queries

```sql
-- Get user's hosts with groups
SELECT h.*, g.name as group_name, g.path
FROM hosts h
LEFT JOIN host_groups g ON h.group_id = g.id
WHERE h.user_id = $1 AND h.deleted_at IS NULL
ORDER BY h.is_favorite DESC, h.name;

-- Get active sessions
SELECT s.*, h.name as host_name, h.hostname
FROM sessions s
JOIN hosts h ON s.host_id = h.id
WHERE s.user_id = $1 AND s.status = 'active'
ORDER BY s.started_at DESC;

-- Search audit log
SELECT * FROM audit_log
WHERE user_id = $1 
  AND timestamp BETWEEN $2 AND $3
  AND action = ANY($4)
ORDER BY timestamp DESC
LIMIT 100;

-- Get credential with decrypted preview (application-level)
SELECT id, name, type, key_version
FROM credentials
WHERE user_id = $1 AND deleted_at IS NULL;
```

### Performance Notes
- All queries use indexed columns
- Audit log queries should include time range
- Tag searches use GIN index
- Soft deletes checked in all listings

---

## 7. Backup & Recovery

### Backup Strategy
```bash
# Full backup
pg_dump -Fc vexa > backup_$(date +%Y%m%d).dump

# Schema only
pg_dump -s vexa > schema_backup.sql

# Specific tables
pg_dump -t audit_log -t sessions vexa > audit_backup.sql
```

### Recovery
```bash
# Restore full backup
pg_restore -d vexa backup_20260528.dump

# Restore with clean slate
dropdb vexa && createdb vexa
pg_restore -d vexa backup_20260528.dump
```

---

## 8. Monitoring

### Key Metrics
| Metric | Query | Alert Threshold |
|--------|-------|-----------------|
| Table sizes | `SELECT pg_size_pretty(pg_total_relation_size('users'))` | > 1GB |
| Index bloat | `SELECT * FROM pg_stat_user_indexes` | > 50% |
| Lock waits | `SELECT * FROM pg_locks WHERE NOT granted` | > 10 |
| Slow queries | `SELECT * FROM pg_stat_statements ORDER BY mean_time DESC LIMIT 10` | > 100ms |
| Connection count | `SELECT count(*) FROM pg_stat_activity` | > 80% max |

---

## Appendix A: Complete Column Reference

See individual migration files in `apps/api/internal/db/migrations/` for full column details.

## Appendix B: Index Reference

See `005_final_indexes.sql` for complete index definitions with `EXPLAIN ANALYZE` rationale.

## Appendix C: Trigger Reference

| Trigger | Table | Event | Function |
|---------|-------|-------|----------|
| `trg_users_updated_at` | users | UPDATE | auto-update timestamp |
| `trg_users_audit` | users | ALL | audit log entry |
| `trg_audit_partition` | audit_log | INSERT | auto-partition |
| `trg_audit_integrity` | audit_log | INSERT | compute hash chain |
| `trg_sessions_cleanup` | sessions | UPDATE | cleanup on close |
