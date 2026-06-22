-- ROLLBACK_003.sql
-- Rollback: audit log integrity enhancements

BEGIN;

-- Drop partition insert trigger (动态创建的分区trigger无法通过固定名称删除，统一处理)
DO $$
DECLARE
    partition_name TEXT;
    trig_name TEXT;
BEGIN
    partition_name := 'audit_log_' || to_char(date_trunc('month', CURRENT_DATE), 'YYYY_MM');
    trig_name := partition_name || '_insert_hash_trigger';
    EXECUTE format('DROP TRIGGER IF EXISTS %I ON %I', trig_name, partition_name);
END;
$$;

-- Drop functions
DROP FUNCTION IF EXISTS audit_log_insert_trigger();
DROP FUNCTION IF EXISTS create_audit_log_partition(DATE);
DROP FUNCTION IF EXISTS generate_audit_integrity_hash(BIGSERIAL, TIMESTAMPTZ, TEXT, UUID, UUID, INET, JSONB, TEXT);

-- Note: Partitions are not dropped — data would be lost
-- Dropping partitions should be done manually or via a separate data migration

DELETE FROM schema_migrations WHERE version = '003_audit_log_integrity';

COMMIT;
