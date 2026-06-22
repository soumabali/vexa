-- seed_000_mark_migrations_applied.sql
-- Mark all migrations as applied so the migrator skips them
-- This is for DEVELOPMENT/TEST environments only

BEGIN;

INSERT INTO schema_migrations (version, description) VALUES
    ('001_initial', 'Base schema: users, devices, credentials, host_groups, hosts, connection_history, audit_log'),
    ('002_sessions_and_vault_keys', 'Add sessions and vault_keys tables'),
    ('003_audit_log_integrity', 'Add audit log integrity functions and auto-partitioning'),
    ('004_host_groups_hierarchy', 'Add LTREE index and hierarchy helpers for host_groups'),
    ('006_credential_sharing', 'Add credential_shares, master_key_versions, and credential metadata');

COMMIT;
