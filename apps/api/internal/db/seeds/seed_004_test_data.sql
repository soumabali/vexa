-- Seed 004: Comprehensive Test Data
-- Description: Generate realistic test data for development and testing
-- Users: 50, Hosts: 200, Credentials: 500, Sessions: 1000, Audit Logs: 10000

BEGIN;

-- ============================================
-- USERS (50 users)
-- ============================================

INSERT INTO users (id, email, password_hash, name, role, is_active, mfa_enabled, failed_login_count, created_at, updated_at) VALUES
(gen_random_uuid(), 'admin@vexa.local', '$argon2id$v=19$m=65536,t=3,p=4$...', 'System Administrator', 'admin', true, true, 0, NOW() - INTERVAL '30 days', NOW()),
(gen_random_uuid(), 'operator1@vexa.local', '$argon2id$v=19$m=65536,t=3,p=4$...', 'Operator One', 'operator', true, true, 0, NOW() - INTERVAL '25 days', NOW()),
(gen_random_uuid(), 'operator2@vexa.local', '$argon2id$v=19$m=65536,t=3,p=4$...', 'Operator Two', 'operator', true, false, 1, NOW() - INTERVAL '20 days', NOW()),
(gen_random_uuid(), 'viewer1@vexa.local', '$argon2id$v=19$m=65536,t=3,p=4$...', 'Viewer One', 'viewer', true, false, 0, NOW() - INTERVAL '15 days', NOW()),
(gen_random_uuid(), 'viewer2@vexa.local', '$argon2id$v=19$m=65536,t=3,p=4$...', 'Viewer Two', 'viewer', true, true, 2, NOW() - INTERVAL '10 days', NOW());

-- Generate remaining 45 users with realistic data
DO $$
DECLARE
    i INT;
    v_user_id UUID;
    roles TEXT[] := ARRAY['admin', 'operator', 'viewer'];
    domains TEXT[] := ARRAY['company.com', 'org.local', 'dev.net', 'lab.io'];
BEGIN
    FOR i IN 6..50 LOOP
        v_user_id := gen_random_uuid();
        INSERT INTO users (id, email, password_hash, name, role, is_active, mfa_enabled, failed_login_count, created_at, updated_at)
        VALUES (
            v_user_id,
            'user' || i || '@' || domains[(i % 4) + 1],
            '$argon2id$v=19$m=65536,t=3,p=4$' || encode(gen_random_bytes(32), 'base64'),
            'Test User ' || i,
            roles[(i % 3) + 1],
            (i % 10 != 0),  -- 10% inactive
            (i % 3 = 0),    -- 33% MFA enabled
            (i % 7),        -- Some failed logins
            NOW() - (INTERVAL '1 day' * (i * 2)),
            NOW() - (INTERVAL '1 hour' * i)
        );
    END LOOP;
END $$;

-- ============================================
-- HOSTS (200 hosts)
-- ============================================

DO $$
DECLARE
    i INT;
    v_user_id UUID;
    protocols TEXT[] := ARRAY['ssh', 'rdp', 'vnc'];
    tags TEXT[] := ARRAY['production', 'staging', 'development', 'database', 'web-server', 'load-balancer', 'monitoring', 'backup'];
BEGIN
    FOR i IN 1..200 LOOP
        SELECT id INTO v_user_id FROM users ORDER BY random() LIMIT 1;
        
        INSERT INTO hosts (id, user_id, name, address, port, protocol, tags, description, health_status, last_seen, created_at, updated_at)
        VALUES (
            gen_random_uuid(),
            v_user_id,
            'host-' || i || '-' || protocols[(i % 3) + 1],
            '192.168.' || (i % 255) || '.' || (i % 254 + 1),
            CASE 
                WHEN i % 3 = 1 THEN 22
                WHEN i % 3 = 2 THEN 3389
                ELSE 5900
            END,
            protocols[(i % 3) + 1],
            ARRAY[tags[(i % 8) + 1], tags[((i + 3) % 8) + 1]],
            'Test host ' || i || ' for ' || protocols[(i % 3) + 1] || ' connections',
            CASE WHEN i % 10 != 0 THEN 'healthy' ELSE 'unhealthy' END,
            NOW() - (INTERVAL '1 hour' * (i % 48)),
            NOW() - (INTERVAL '1 day' * (i % 30)),
            NOW() - (INTERVAL '1 hour' * (i % 24))
        );
    END LOOP;
END $$;

-- ============================================
-- CREDENTIALS (500 credentials)
-- ============================================

DO $$
DECLARE
    i INT;
    v_user_id UUID;
    v_host_id UUID;
    types TEXT[] := ARRAY['password', 'ssh_key', 'api_token', 'certificate'];
BEGIN
    FOR i IN 1..500 LOOP
        SELECT id INTO v_user_id FROM users ORDER BY random() LIMIT 1;
        SELECT id INTO v_host_id FROM hosts WHERE user_id = v_user_id ORDER BY random() LIMIT 1;
        
        INSERT INTO credentials (id, user_id, host_id, name, type, data_encrypted, is_active, created_at, updated_at)
        VALUES (
            gen_random_uuid(),
            v_user_id,
            v_host_id,
            'cred-' || i || '-' || types[(i % 4) + 1],
            types[(i % 4) + 1],
            encrypt(encode(gen_random_bytes(64), 'base64'), 'master-key'),
            (i % 20 != 0),  -- 5% inactive
            NOW() - (INTERVAL '1 day' * (i % 60)),
            NOW() - (INTERVAL '1 hour' * (i % 100))
        );
    END LOOP;
END $$;

-- ============================================
-- SESSIONS (1000 sessions)
-- ============================================

DO $$
DECLARE
    i INT;
    v_user_id UUID;
    v_host_id UUID;
    statuses TEXT[] := ARRAY['active', 'closed', 'timeout', 'error'];
BEGIN
    FOR i IN 1..1000 LOOP
        SELECT id INTO v_user_id FROM users ORDER BY random() LIMIT 1;
        SELECT id INTO v_host_id FROM hosts WHERE user_id = v_user_id ORDER BY random() LIMIT 1;
        
        INSERT INTO sessions (id, user_id, host_id, protocol, status, started_at, ended_at, bytes_sent, bytes_received, recording_path, created_at)
        VALUES (
            gen_random_uuid(),
            v_user_id,
            v_host_id,
            CASE 
                WHEN i % 3 = 1 THEN 'ssh'
                WHEN i % 3 = 2 THEN 'rdp'
                ELSE 'vnc'
            END,
            CASE 
                WHEN i <= 50 THEN 'active'
                ELSE statuses[(i % 4) + 1]
            END,
            NOW() - (INTERVAL '1 hour' * (i % 168)),  -- Up to 1 week ago
            CASE 
                WHEN i <= 50 THEN NULL  -- Active sessions
                ELSE NOW() - (INTERVAL '1 hour' * ((i % 168) - (i % 24)))
            END,
            (i * 1024 * 1024)::BIGINT,  -- MB sent
            (i * 512 * 1024)::BIGINT,     -- MB received
            CASE 
                WHEN i % 5 = 0 THEN '/recordings/session_' || i || '.cast'
                ELSE NULL
            END,
            NOW() - (INTERVAL '1 hour' * (i % 168))
        );
    END LOOP;
END $$;

-- ============================================
-- AUDIT_LOG (10000 audit logs)
-- ============================================

DO $$
DECLARE
    i INT;
    v_user_id UUID;
    actions TEXT[] := ARRAY['login', 'logout', 'host_create', 'host_update', 'host_delete', 'credential_create', 'credential_update', 'credential_delete', 'session_start', 'session_end', 'vault_unlock', 'vault_lock', 'mfa_enable', 'mfa_disable', 'settings_update'];
    ips TEXT[] := ARRAY['192.168.1.', '10.0.0.', '172.16.0.', '203.0.113.'];
BEGIN
    FOR i IN 1..10000 LOOP
        SELECT id INTO v_user_id FROM users ORDER BY random() LIMIT 1;
        
        INSERT INTO audit_log (id, user_id, action, entity_type, entity_id, ip_address, user_agent, details, timestamp, integrity_hash)
        VALUES (
            gen_random_uuid(),
            v_user_id,
            actions[(i % 15) + 1],
            CASE 
                WHEN i % 15 IN (0,1,2) THEN 'user'
                WHEN i % 15 IN (3,4,5) THEN 'host'
                WHEN i % 15 IN (6,7,8) THEN 'credential'
                WHEN i % 15 IN (9,10) THEN 'session'
                ELSE 'system'
            END,
            gen_random_uuid(),
            ips[(i % 4) + 1] || (i % 255),
            'Mozilla/5.0 (Test-Agent)',
            jsonb_build_object(
                'test_data', true,
                'index', i,
                'metadata', jsonb_build_object('source', 'seed', 'batch', '004')
            ),
            NOW() - (INTERVAL '1 minute' * (i % 10080)),  -- Up to 1 week
            encode(digest(gen_random_bytes(64), 'sha256'), 'hex')
        );
    END LOOP;
END $$;

-- ============================================
-- HOST_GROUPS (20 groups with hierarchy)
-- ============================================

INSERT INTO host_groups (id, user_id, name, path, description, created_at) VALUES
(gen_random_uuid(), (SELECT id FROM users WHERE role = 'admin' LIMIT 1), 'Production', 'production', 'Production servers', NOW()),
(gen_random_uuid(), (SELECT id FROM users WHERE role = 'admin' LIMIT 1), 'Staging', 'production.staging', 'Staging environment', NOW()),
(gen_random_uuid(), (SELECT id FROM users WHERE role = 'admin' LIMIT 1), 'Development', 'production.development', 'Development environment', NOW()),
(gen_random_uuid(), (SELECT id FROM users WHERE role = 'admin' LIMIT 1), 'Web Servers', 'production.web', 'Web server cluster', NOW()),
(gen_random_uuid(), (SELECT id FROM users WHERE role = 'admin' LIMIT 1), 'Database', 'production.database', 'Database servers', NOW());

-- ============================================
-- SHARED_CREDENTIALS (50 shares)
-- ============================================

DO $$
DECLARE
    i INT;
    v_shared_by UUID;
    v_shared_with UUID;
    v_credential_id UUID;
BEGIN
    FOR i IN 1..50 LOOP
        SELECT id INTO v_shared_by FROM users ORDER BY random() LIMIT 1;
        SELECT id INTO v_shared_with FROM users WHERE id != v_shared_by ORDER BY random() LIMIT 1;
        SELECT id INTO v_credential_id FROM credentials WHERE user_id = v_shared_by ORDER BY random() LIMIT 1;
        
        INSERT INTO shared_credentials (id, credential_id, shared_by_id, shared_with_id, encrypted_data, expires_at, access_level, created_at)
        VALUES (
            gen_random_uuid(),
            v_credential_id,
            v_shared_by,
            v_shared_with,
            encrypt(encode(gen_random_bytes(32), 'base64'), 'master-key'),
            NOW() + (INTERVAL '1 day' * (i % 30)),
            CASE 
                WHEN i % 3 = 0 THEN 'read'
                WHEN i % 3 = 1 THEN 'write'
                ELSE 'admin'
            END,
            NOW()
        );
    END LOOP;
END $$;

-- ============================================
-- USER_DEVICES (100 devices)
-- ============================================

DO $$
DECLARE
    i INT;
    v_user_id UUID;
    device_types TEXT[] := ARRAY['desktop', 'mobile', 'tablet'];
BEGIN
    FOR i IN 1..100 LOOP
        SELECT id INTO v_user_id FROM users ORDER BY random() LIMIT 1;
        
        INSERT INTO user_devices (id, user_id, device_name, device_type, fingerprint, trusted, last_used, created_at)
        VALUES (
            gen_random_uuid(),
            v_user_id,
            'Device-' || i || '-' || device_types[(i % 3) + 1],
            device_types[(i % 3) + 1],
            encode(gen_random_bytes(16), 'hex'),
            (i % 10 != 0),  -- 10% untrusted
            NOW() - (INTERVAL '1 hour' * (i % 168)),
            NOW() - (INTERVAL '1 day' * (i % 30))
        );
    END LOOP;
END $$;

COMMIT;

-- Verify counts
SELECT 'users' as table_name, COUNT(*) as count FROM users
UNION ALL SELECT 'hosts', COUNT(*) FROM hosts
UNION ALL SELECT 'credentials', COUNT(*) FROM credentials
UNION ALL SELECT 'sessions', COUNT(*) FROM sessions
UNION ALL SELECT 'audit_log', COUNT(*) FROM audit_log
UNION ALL SELECT 'host_groups', COUNT(*) FROM host_groups
UNION ALL SELECT 'shared_credentials', COUNT(*) FROM shared_credentials
UNION ALL SELECT 'user_devices', COUNT(*) FROM user_devices;
