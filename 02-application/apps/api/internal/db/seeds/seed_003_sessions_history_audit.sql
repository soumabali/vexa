-- seed_003_sessions_connection_history.sql
-- Seed data: active/historical sessions and connection history

BEGIN;

-- Vault keys for each user
INSERT INTO vault_keys (id, user_id, version, algorithm, key_hash, salt, wrapping_salt, wrapped_key, is_active, rotated_at, created_at)
VALUES (
    '50000000-0000-0000-0000-000000000001',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    1,
    'argon2id-aes-256-gcm',
    encode(sha256('test-key-hash'::bytea), 'hex'),
    decode('c2FsdGZvcmVuY3J5cHRpb25n', 'base64'),
    decode('d3JhcHBpbmdzYWx0Zm9yZW5j', 'base64'),
    decode('d3JhcHBpbmdrZXlzZWNyZXQ=', 'base64'),
    TRUE,
    NOW() - INTERVAL '25 days',
    NOW() - INTERVAL '25 days'
);

INSERT INTO vault_keys (id, user_id, version, algorithm, key_hash, salt, wrapping_salt, wrapped_key, is_active, rotated_at, created_at)
VALUES (
    '50000000-0000-0000-0000-000000000002',
    'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22',
    1,
    'argon2id-aes-256-gcm',
    encode(sha256('test-key-hash'::bytea), 'hex'),
    decode('c2FsdGZvcmVuY3J5cHRpb25n', 'base64'),
    decode('d3JhcHBpbmdzYWx0Zm9yZW5j', 'base64'),
    decode('d3JhcHBpbmdrZXlzZWNyZXQ=', 'base64'),
    TRUE,
    NOW() - INTERVAL '20 days',
    NOW() - INTERVAL '20 days'
);

INSERT INTO vault_keys (id, user_id, version, algorithm, key_hash, salt, wrapping_salt, wrapped_key, is_active, rotated_at, created_at)
VALUES (
    '50000000-0000-0000-0000-000000000003',
    'c2eebc99-9c0b-4ef8-bb6d-6bb9bd380a33',
    1,
    'argon2id-aes-256-gcm',
    encode(sha256('test-key-hash'::bytea), 'hex'),
    decode('c2FsdGZvcmVuY3J5cHRpb25n', 'base64'),
    decode('d3JhcHBpbmdzYWx0Zm9yZW5j', 'base64'),
    decode('d3JhcHBpbmdrZXlzZWNyZXQ=', 'base64'),
    TRUE,
    NOW() - INTERVAL '10 days',
    NOW() - INTERVAL '10 days'
);

-- Active session (admin on web-prod-01)
INSERT INTO sessions (id, user_id, device_id, host_id, protocol, session_type, started_at, bytes_in, bytes_out, status, terminated_by)
VALUES (
    '60000000-0000-0000-0000-000000000001',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a55',
    '30000000-0000-0000-0000-000000000001',
    'ssh',
    'terminal',
    NOW() - INTERVAL '5 minutes',
    4096,
    2048,
    'active',
    NULL
);

-- Active SFTP session
INSERT INTO sessions (id, user_id, device_id, host_id, protocol, session_type, started_at, bytes_in, bytes_out, status)
VALUES (
    '60000000-0000-0000-0000-000000000002',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a55',
    '30000000-0000-0000-0000-000000000003',
    'ssh',
    'sftp',
    NOW() - INTERVAL '2 minutes',
    8192,
    1024,
    'active'
);

-- Active RDP session
INSERT INTO sessions (id, user_id, device_id, host_id, protocol, session_type, started_at, bytes_in, bytes_out, status)
VALUES (
    '60000000-0000-0000-0000-000000000003',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a55',
    '30000000-0000-0000-0000-000000000005',
    'rdp',
    'terminal',
    NOW() - INTERVAL '1 minute',
    20480,
    5120,
    'active'
);

-- Recent closed sessions (connection history)
INSERT INTO sessions (id, user_id, device_id, host_id, protocol, session_type, started_at, ended_at, bytes_in, bytes_out, status, terminated_by, termination_reason)
VALUES (
    '60000000-0000-0000-0000-000000000010',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a55',
    '30000000-0000-0000-0000-000000000001',
    'ssh',
    'terminal',
    NOW() - INTERVAL '1 hour',
    NOW() - INTERVAL '50 minutes',
    10240,
    8192,
    'closed',
    NULL,
    'normal'
);

INSERT INTO sessions (id, user_id, device_id, host_id, protocol, session_type, started_at, ended_at, bytes_in, bytes_out, status, terminated_by, termination_reason)
VALUES (
    '60000000-0000-0000-0000-000000000011',
    'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22',
    'g6eebc99-9c0b-4ef8-bb6d-6bb9bd380a77',
    '30000000-0000-0000-0000-000000000006',
    'ssh',
    'terminal',
    NOW() - INTERVAL '3 hours',
    NOW() - INTERVAL '2 hours',
    5120,
    4096,
    'closed',
    NULL,
    'normal'
);

INSERT INTO sessions (id, user_id, device_id, host_id, protocol, session_type, started_at, ended_at, bytes_in, bytes_out, status, terminated_by, termination_reason)
VALUES (
    '60000000-0000-0000-0000-000000000012',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a55',
    '30000000-0000-0000-0000-000000000002',
    'ssh',
    'terminal',
    NOW() - INTERVAL '2 hours',
    NOW() - INTERVAL '1 hour',
    3072,
    2560,
    'killed',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'timeout'
);

INSERT INTO sessions (id, user_id, device_id, host_id, protocol, session_type, started_at, ended_at, bytes_in, bytes_out, status, termination_reason)
VALUES (
    '60000000-0000-0000-0000-000000000013',
    'c2eebc99-9c0b-4ef8-bb6d-6bb9bd380a33',
    'h7eebc99-9c0b-4ef8-bb6d-6bb9bd380a88',
    '30000000-0000-0000-0000-000000000008',
    'ssh',
    'terminal',
    NOW() - INTERVAL '24 hours',
    NOW() - INTERVAL '23 hours',
    2048,
    1536,
    'closed',
    'normal'
);

-- Audit log entries
INSERT INTO audit_log (id, timestamp, event_type, user_id, device_id, ip_address, details, integrity_hash)
VALUES
    (1, NOW() - INTERVAL '30 days', 'user_login', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a55', '192.168.1.100'::inet, '{"result": "success"}', 'hmac-sha256:auto-generated'),
    (2, NOW() - INTERVAL '28 days', 'user_login', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'g6eebc99-9c0b-4ef8-bb6d-6bb9bd380a77', '10.0.0.50'::inet, '{"result": "success"}', 'hmac-sha256:auto-generated'),
    (3, NOW() - INTERVAL '25 days', 'host_created', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a55', '192.168.1.100'::inet, '{"host_id": "30000000-0000-0000-0000-000000000001", "name": "web-prod-01"}', 'hmac-sha256:auto-generated'),
    (4, NOW() - INTERVAL '20 days', 'credential_created', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'g6eebc99-9c0b-4ef8-bb6d-6bb9bd380a77', '10.0.0.50'::inet, '{"credential_id": "20000000-0000-0000-0000-000000000003"}', 'hmac-sha256:auto-generated'),
    (5, NOW() - INTERVAL '15 days', 'session_started', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a55', '192.168.1.100'::inet, '{"session_id": "60000000-0000-0000-0000-000000000010", "host_id": "30000000-0000-0000-0000-000000000001"}', 'hmac-sha256:auto-generated'),
    (6, NOW() - INTERVAL '14 days', 'vault_unlocked', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a55', '192.168.1.100'::inet, '{"result": "success"}', 'hmac-sha256:auto-generated'),
    (7, NOW() - INTERVAL '10 days', 'user_login', 'c2eebc99-9c0b-4ef8-bb6d-6bb9bd380a33', 'h7eebc99-9c0b-4ef8-bb6d-6bb9bd380a88', '192.168.100.5'::inet, '{"result": "success"}', 'hmac-sha256:auto-generated'),
    (8, NOW() - INTERVAL '5 minutes', 'session_started', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a55', '192.168.1.100'::inet, '{"session_id": "60000000-0000-0000-0000-000000000001", "protocol": "ssh"}', 'hmac-sha256:auto-generated'),
    (9, NOW() - INTERVAL '1 minute', 'session_started', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a55', '192.168.1.100'::inet, '{"session_id": "60000000-0000-0000-0000-000000000003", "protocol": "rdp"}', 'hmac-sha256:auto-generated');

COMMIT;
