# Migration Manifest
# Run order: 001 → 002 → 003 → 004 → 005 → 006
#
# Rollback order (reverse): ROLLBACK_006 → ROLLBACK_004_005 → ROLLBACK_003 → ROLLBACK_002 → ROLLBACK_001
#
# Quick reference:
#   make migrate-up          # Apply all pending migrations
#   make migrate-status     # Show migration status
#   make migrate-seed        # Load seed data
#   make migrate-down       # Rollback last migration
#   make migrate-reset      # Reset all migrations (WARNING: drops all data)
#   make migrate-rollback-to VERSION=001  # Rollback to specific version
#
# Docker development (auto-runs migrations on container start):
#   Scripts in migrations/ directory are mounted to /docker-entrypoint-initdb.d/
#
# Seed data order:
#   seed_000_mark_migrations_applied.sql  (only in dev/test — skip in production)
#   seed_001_users_devices.sql
#   seed_002_hosts_credentials.sql
#   seed_003_sessions_history_audit.sql
#
# Seed users and passwords (dev only — must be changed in production):
#   admin@vexa.local    — role: admin    — any 16+ char password
#   operator@vexa.local — role: operator — any 16+ char password
#   viewer@vexa.local   — role: viewer   — any 16+ char password
#   inactive@vexa.local — role: viewer   — is_active = FALSE

## Tables created

| Migration | Table(s) | Description |
|-----------|----------|-------------|
| 001_initial | users, devices, credentials, host_groups, hosts, connection_history, audit_log | Base schema with FKs, indexes, triggers |
| 002_sessions_and_vault_keys | sessions, vault_keys | Active session tracking + vault key management |
| 003_audit_log_integrity | — (functions only) | Integrity hash functions, auto-partition trigger |
| 004_host_groups_hierarchy | — (columns/indexes only) | LTREE path column + GiST index + hierarchy helpers |
| 005_final_indexes | — (indexes only) | Query optimization indexes |
| 006_credential_sharing | credential_shares, master_key_versions | Encrypted cross-user credential sharing + key rotation audit |
