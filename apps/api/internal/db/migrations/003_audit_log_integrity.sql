-- 003_audit_log_integrity.sql
-- Audit log enhancements: integrity verification, auto-partitioning

BEGIN;

-- Create function to generate row integrity hash (HMAC-SHA256)
CREATE OR REPLACE FUNCTION generate_audit_integrity_hash(
    p_id BIGINT,
    p_timestamp TIMESTAMPTZ,
    p_event_type TEXT,
    p_user_id UUID,
    p_device_id UUID,
    p_ip_address INET,
    p_details JSONB,
    p_secret TEXT
)
RETURNS TEXT AS $$
DECLARE
    data TEXT;
    mac TEXT;
BEGIN
    data := COALESCE(p_id::TEXT, '') || '|' ||
            COALESCE(p_timestamp::TEXT, '') || '|' ||
            COALESCE(p_event_type, '') || '|' ||
            COALESCE(p_user_id::TEXT, '') || '|' ||
            COALESCE(p_device_id::TEXT, '') || '|' ||
            COALESCE(p_ip_address::TEXT, '') || '|' ||
            COALESCE(p_details::TEXT, '');
    mac := 'hmac-sha256:' || encode(hmac(data::bytea, p_secret::bytea, 'sha256'), 'hex');
    RETURN mac;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create new partition for current month (auto-create partition)
DO $$
DECLARE
    partition_name TEXT;
    start_date TEXT;
    end_date TEXT;
BEGIN
    start_date := to_char(date_trunc('month', CURRENT_DATE), 'YYYY-MM-DD');
    end_date := to_char(date_trunc('month', CURRENT_DATE) + INTERVAL '1 month', 'YYYY-MM-DD');
    partition_name := 'audit_log_' || to_char(date_trunc('month', CURRENT_DATE), 'YYYY_MM');

    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF audit_log FOR VALUES FROM (%L) TO (%L)',
        partition_name, start_date, end_date
    );
END;
$$;

-- Create partition creation function for future months
CREATE OR REPLACE FUNCTION create_audit_log_partition(partition_date DATE)
RETURNS VOID AS $$
DECLARE
    partition_name TEXT;
    start_date TEXT;
    end_date TEXT;
BEGIN
    start_date := to_char(date_trunc('month', partition_date), 'YYYY-MM-DD');
    end_date := to_char(date_trunc('month', partition_date) + INTERVAL '1 month', 'YYYY-MM-DD');
    partition_name := 'audit_log_' || to_char(date_trunc('month', partition_date), 'YYYY_MM');

    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF audit_log FOR VALUES FROM (%L) TO (%L)',
        partition_name, start_date, end_date
    );
END;
$$ LANGUAGE plpgsql;

-- Create audit insert trigger with auto-include integrity hash
CREATE OR REPLACE FUNCTION audit_log_insert_trigger()
RETURNS TRIGGER AS $$
BEGIN
    -- Auto-generate integrity hash if not provided
    NEW.integrity_hash := generate_audit_integrity_hash(
        NEW.id,
        NEW.timestamp,
        NEW.event_type,
        NEW.user_id,
        NEW.device_id,
        NEW.ip_address,
        NEW.details,
        current_setting('app.audit_secret', true)
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Get current partition name and create trigger
DO $$
DECLARE
    partition_name TEXT;
BEGIN
    partition_name := 'audit_log_' || to_char(date_trunc('month', CURRENT_DATE), 'YYYY_MM');
    EXECUTE format('CREATE TRIGGER audit_log_insert_hash_trigger BEFORE INSERT ON %I FOR EACH ROW EXECUTE FUNCTION audit_log_insert_trigger()', partition_name);
END;
$$;

-- Add comments for documentation
COMMENT ON TABLE audit_log IS 'Immutable audit log table. Partitions created monthly. Rows must not be modified after insert.';
COMMENT ON FUNCTION generate_audit_integrity_hash IS 'Generates HMAC-SHA256 integrity hash for audit log rows. Uses app.audit_secret GUC.';

-- Record migration
INSERT INTO schema_migrations (version, description)
VALUES ('003_audit_log_integrity', 'Add audit log integrity functions and auto-partitioning')
ON CONFLICT (version) DO NOTHING;

COMMIT;
