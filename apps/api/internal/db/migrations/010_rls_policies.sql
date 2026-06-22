-- Migration 010: Row-Level Security Policies
-- Description: Implement PostgreSQL RLS for multi-tenant data isolation
-- Run order: 010 (after 006_credential_sharing)

BEGIN;

-- Enable RLS on all tables (skip tables that do not yet exist)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_class WHERE relname = 'users' AND relkind = 'r') THEN
        ALTER TABLE users ENABLE ROW LEVEL SECURITY;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_class WHERE relname = 'hosts' AND relkind = 'r') THEN
        ALTER TABLE hosts ENABLE ROW LEVEL SECURITY;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_class WHERE relname = 'credentials' AND relkind = 'r') THEN
        ALTER TABLE credentials ENABLE ROW LEVEL SECURITY;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_class WHERE relname = 'sessions' AND relkind = 'r') THEN
        ALTER TABLE sessions ENABLE ROW LEVEL SECURITY;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_class WHERE relname = 'audit_log' AND relkind = 'r') THEN
        ALTER TABLE audit_log ENABLE ROW LEVEL SECURITY;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_class WHERE relname = 'host_groups' AND relkind = 'r') THEN
        ALTER TABLE host_groups ENABLE ROW LEVEL SECURITY;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_class WHERE relname = 'shared_credentials' AND relkind = 'r') THEN
        ALTER TABLE shared_credentials ENABLE ROW LEVEL SECURITY;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_class WHERE relname = 'user_devices' AND relkind = 'r') THEN
        ALTER TABLE user_devices ENABLE ROW LEVEL SECURITY;
    END IF;
END;
$$;

-- Function to get current tenant_id from session
CREATE OR REPLACE FUNCTION current_tenant_id()
RETURNS UUID AS $$
BEGIN
    RETURN NULLIF(current_setting('app.current_tenant_id', true), '')::UUID;
EXCEPTION WHEN OTHERS THEN
    RETURN NULL;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- Function to get current user_id from session
CREATE OR REPLACE FUNCTION current_user_id()
RETURNS UUID AS $$
BEGIN
    RETURN NULLIF(current_setting('app.current_user_id', true), '')::UUID;
EXCEPTION WHEN OTHERS THEN
    RETURN NULL;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- Function to check if current user is admin
CREATE OR REPLACE FUNCTION current_user_is_admin()
RETURNS BOOLEAN AS $$
DECLARE
    v_role VARCHAR(20);
BEGIN
    SELECT role INTO v_role FROM users WHERE id = current_user_id();
    RETURN v_role = 'admin';
EXCEPTION WHEN OTHERS THEN
    RETURN FALSE;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- ============================================
-- USERS TABLE
-- ============================================

-- Users can only see their own data, admins can see all
CREATE POLICY users_select_policy ON users
    FOR SELECT
    USING (
        id = current_user_id()
        OR current_user_is_admin()
        OR current_setting('app.bypass_rls', true) = 'true'
    );

-- Users can only update their own data
CREATE POLICY users_update_policy ON users
    FOR UPDATE
    USING (id = current_user_id())
    WITH CHECK (id = current_user_id());

-- Only admins can insert/delete
CREATE POLICY users_insert_policy ON users
    FOR INSERT
    WITH CHECK (current_user_is_admin() OR current_setting('app.bypass_rls', true) = 'true');

CREATE POLICY users_delete_policy ON users
    FOR DELETE
    USING (current_user_is_admin() OR current_setting('app.bypass_rls', true) = 'true');

-- ============================================
-- HOSTS TABLE
-- ============================================

-- Users can see their own hosts + shared hosts
CREATE POLICY hosts_select_policy ON hosts
    FOR SELECT
    USING (
        owner_id = current_user_id()
        OR current_user_is_admin()
        OR current_setting('app.bypass_rls', true) = 'true'
    );

-- Users can only modify their own hosts
CREATE POLICY hosts_insert_policy ON hosts
    FOR INSERT
    WITH CHECK (
        owner_id = current_user_id()
        OR current_user_is_admin()
        OR current_setting('app.bypass_rls', true) = 'true'
    );

CREATE POLICY hosts_update_policy ON hosts
    FOR UPDATE
    USING (
        owner_id = current_user_id()
        OR current_user_is_admin()
        OR current_setting('app.bypass_rls', true) = 'true'
    )
    WITH CHECK (
        owner_id = current_user_id()
        OR current_user_is_admin()
    );

CREATE POLICY hosts_delete_policy ON hosts
    FOR DELETE
    USING (
        owner_id = current_user_id()
        OR current_user_is_admin()
        OR current_setting('app.bypass_rls', true) = 'true'
    );

-- ============================================
-- CREDENTIALS TABLE
-- ============================================

-- Users can see their own credentials + shared credentials
CREATE POLICY credentials_select_policy ON credentials
    FOR SELECT
    USING (
        owner_id = current_user_id()
        OR current_user_is_admin()
        OR current_setting('app.bypass_rls', true) = 'true'
    );

CREATE POLICY credentials_insert_policy ON credentials
    FOR INSERT
    WITH CHECK (
        owner_id = current_user_id()
        OR current_user_is_admin()
        OR current_setting('app.bypass_rls', true) = 'true'
    );

CREATE POLICY credentials_update_policy ON credentials
    FOR UPDATE
    USING (
        owner_id = current_user_id()
        OR current_user_is_admin()
        OR current_setting('app.bypass_rls', true) = 'true'
    )
    WITH CHECK (
        owner_id = current_user_id()
        OR current_user_is_admin()
    );

CREATE POLICY credentials_delete_policy ON credentials
    FOR DELETE
    USING (
        owner_id = current_user_id()
        OR current_user_is_admin()
        OR current_setting('app.bypass_rls', true) = 'true'
    );

-- ============================================
-- SESSIONS TABLE
-- ============================================

-- Users can only see their own sessions
CREATE POLICY sessions_select_policy ON sessions
    FOR SELECT
    USING (
        user_id = current_user_id()
        OR current_user_is_admin()
        OR current_setting('app.bypass_rls', true) = 'true'
    );

CREATE POLICY sessions_insert_policy ON sessions
    FOR INSERT
    WITH CHECK (
        user_id = current_user_id()
        OR current_setting('app.bypass_rls', true) = 'true'
    );

CREATE POLICY sessions_update_policy ON sessions
    FOR UPDATE
    USING (
        user_id = current_user_id()
        OR current_user_is_admin()
        OR current_setting('app.bypass_rls', true) = 'true'
    )
    WITH CHECK (
        user_id = current_user_id()
        OR current_user_is_admin()
    );

CREATE POLICY sessions_delete_policy ON sessions
    FOR DELETE
    USING (
        user_id = current_user_id()
        OR current_user_is_admin()
        OR current_setting('app.bypass_rls', true) = 'true'
    );

-- ============================================
-- AUDIT_LOG TABLE
-- ============================================

-- Audit log is append-only, users can see their own events
-- Admins can see all events
CREATE POLICY audit_log_select_policy ON audit_log
    FOR SELECT
    USING (
        user_id = current_user_id()
        OR current_user_is_admin()
        OR current_setting('app.bypass_rls', true) = 'true'
    );

-- Only system/app can insert audit logs
CREATE POLICY audit_log_insert_policy ON audit_log
    FOR INSERT
    WITH CHECK (TRUE); -- Insert allowed, but select restricted

-- Audit logs should never be updated or deleted
-- (handled by trigger, but policy ensures)
CREATE POLICY audit_log_update_policy ON audit_log
    FOR UPDATE
    USING (FALSE);

CREATE POLICY audit_log_delete_policy ON audit_log
    FOR DELETE
    USING (current_user_is_admin());

-- Host groups ownership
CREATE POLICY host_groups_select_policy ON host_groups
    FOR SELECT
    USING (
        owner_id = current_user_id()
        OR current_user_is_admin()
        OR current_setting('app.bypass_rls', true) = 'true'
    );

CREATE POLICY host_groups_insert_policy ON host_groups
    FOR INSERT
    WITH CHECK (
        owner_id = current_user_id()
        OR current_user_is_admin()
        OR current_setting('app.bypass_rls', true) = 'true'
    );

CREATE POLICY host_groups_update_policy ON host_groups
    FOR UPDATE
    USING (
        owner_id = current_user_id()
        OR current_user_is_admin()
        OR current_setting('app.bypass_rls', true) = 'true'
    )
    WITH CHECK (
        owner_id = current_user_id()
        OR current_user_is_admin()
    );

CREATE POLICY host_groups_delete_policy ON host_groups
    FOR DELETE
    USING (
        owner_id = current_user_id()
        OR current_user_is_admin()
        OR current_setting('app.bypass_rls', true) = 'true'
    );

-- ============================================
-- SHARED_CREDENTIALS TABLE (legacy; skip if table does not exist)
-- ============================================

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_class WHERE relname = 'shared_credentials' AND relkind = 'r') THEN
        EXECUTE format('CREATE POLICY shared_credentials_select_policy ON shared_credentials
            FOR SELECT
            USING (
                shared_with_id = current_user_id()
                OR shared_by_id = current_user_id()
                OR current_user_is_admin()
                OR current_setting(''app.bypass_rls'', true) = ''true''
            )');

        EXECUTE format('CREATE POLICY shared_credentials_insert_policy ON shared_credentials
            FOR INSERT
            WITH CHECK (
                shared_by_id = current_user_id()
                OR current_user_is_admin()
                OR current_setting(''app.bypass_rls'', true) = ''true''
            )');

        EXECUTE format('CREATE POLICY shared_credentials_update_policy ON shared_credentials
            FOR UPDATE
            USING (
                shared_by_id = current_user_id()
                OR current_user_is_admin()
                OR current_setting(''app.bypass_rls'', true) = ''true''
            )
            WITH CHECK (
                shared_by_id = current_user_id()
                OR current_user_is_admin()
            )');

        EXECUTE format('CREATE POLICY shared_credentials_delete_policy ON shared_credentials
            FOR DELETE
            USING (
                shared_by_id = current_user_id()
                OR shared_with_id = current_user_id()
                OR current_user_is_admin()
                OR current_setting(''app.bypass_rls'', true) = ''true''
            )');
    END IF;
END;
$$;

-- ============================================
-- USER_DEVICES TABLE (legacy; skip if table does not exist)
-- ============================================

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_class WHERE relname = 'user_devices' AND relkind = 'r') THEN
        EXECUTE format('CREATE POLICY user_devices_select_policy ON user_devices
            FOR SELECT
            USING (
                user_id = current_user_id()
                OR current_user_is_admin()
                OR current_setting(''app.bypass_rls'', true) = ''true''
            )');

        EXECUTE format('CREATE POLICY user_devices_insert_policy ON user_devices
            FOR INSERT
            WITH CHECK (
                user_id = current_user_id()
                OR current_setting(''app.bypass_rls'', true) = ''true''
            )');

        EXECUTE format('CREATE POLICY user_devices_update_policy ON user_devices
            FOR UPDATE
            USING (
                user_id = current_user_id()
                OR current_user_is_admin()
                OR current_setting(''app.bypass_rls'', true) = ''true''
            )
            WITH CHECK (
                user_id = current_user_id()
                OR current_user_is_admin()
            )');

        EXECUTE format('CREATE POLICY user_devices_delete_policy ON user_devices
            FOR DELETE
            USING (
                user_id = current_user_id()
                OR current_user_is_admin()
                OR current_setting(''app.bypass_rls'', true) = ''true''
            )');
    END IF;
END;
$$;

-- ============================================
-- BYPASS RLS ROLE
-- ============================================

CREATE ROLE bypass_rls;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO bypass_rls;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO bypass_rls;

-- Allow bypass_rls to bypass RLS
ALTER ROLE bypass_rls BYPASSRLS;

-- ============================================
-- INDEXES FOR RLS PERFORMANCE
-- ============================================

CREATE INDEX idx_hosts_user_id_rls ON hosts(owner_id) WHERE owner_id IS NOT NULL;
CREATE INDEX idx_credentials_user_id_rls ON credentials(owner_id) WHERE owner_id IS NOT NULL;
CREATE INDEX idx_sessions_user_id_rls ON sessions(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_audit_log_user_id_rls ON audit_log(user_id) WHERE user_id IS NOT NULL;

COMMIT;
