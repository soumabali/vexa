-- seed_002_hosts_credentials.sql
-- Seed data: host groups, hosts, and encrypted credentials

BEGIN;

-- Host groups (production infrastructure hierarchy)
INSERT INTO host_groups (id, owner_id, name, path, parent_id, created_at)
VALUES ('10000000-0000-0000-0000-000000000001', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Production', 'production', NULL, NOW() - INTERVAL '28 days');

INSERT INTO host_groups (id, owner_id, name, path, parent_id, created_at)
VALUES ('10000000-0000-0000-0000-000000000002', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Production Web', 'production.web', '10000000-0000-0000-0000-000000000001', NOW() - INTERVAL '27 days');

INSERT INTO host_groups (id, owner_id, name, path, parent_id, created_at)
VALUES ('10000000-0000-0000-0000-000000000003', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Production DB', 'production.db', '10000000-0000-0000-0000-000000000001', NOW() - INTERVAL '27 days');

INSERT INTO host_groups (id, owner_id, name, path, parent_id, created_at)
VALUES ('10000000-0000-0000-0000-000000000004', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Staging', 'staging', NULL, NOW() - INTERVAL '26 days');

INSERT INTO host_groups (id, owner_id, name, path, parent_id, created_at)
VALUES ('10000000-0000-0000-0000-000000000005', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Development', 'development', NULL, NOW() - INTERVAL '25 days');

-- Credentials (encrypted test data)
-- encrypted_data is AES-256-GCM ciphertext of: {"username":"admin","password":"SecretPass123!"}
-- For seed data, we use placeholder base64 values; real encryption happens at application layer

INSERT INTO credentials (id, owner_id, name, type, encrypted_data, nonce, salt, version, created_at)
VALUES (
    '20000000-0000-0000-0000-000000000001',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'Production Root SSH',
    'ssh_key',
    decode('U29tZVRlc3RDcnlwdG9kYXRh', 'base64'),
    decode('YWJjZGVmZ2hpamtsbW5vcA==', 'base64'),
    decode('c2FsdGZvcmVuY3J5cHRpb24=', 'base64'),
    1,
    NOW() - INTERVAL '25 days'
);

INSERT INTO credentials (id, owner_id, name, type, encrypted_data, nonce, salt, version, created_at)
VALUES (
    '20000000-0000-0000-0000-000000000002',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'Production DB Password',
    'password',
    decode('U29tZVRlc3RDcnlwdG9kYXRh', 'base64'),
    decode('YWJjZGVmZ2hpamtsbW5vcA==', 'base64'),
    decode('c2FsdGZvcmVuY3J5cHRpb24=', 'base64'),
    1,
    NOW() - INTERVAL '24 days'
);

INSERT INTO credentials (id, owner_id, name, type, encrypted_data, nonce, salt, version, created_at)
VALUES (
    '20000000-0000-0000-0000-000000000003',
    'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22',
    'Staging Deploy Token',
    'api_token',
    decode('VG9rZW5TZWVkRGF0YQ==', 'base64'),
    decode('dG9rZW5ub25jZQ==', 'base64'),
    decode('dG9rZW5zYWx0', 'base64'),
    1,
    NOW() - INTERVAL '20 days'
);

INSERT INTO credentials (id, owner_id, name, type, encrypted_data, nonce, salt, version, created_at)
VALUES (
    '20000000-0000-0000-0000-000000000004',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'RDP Admin Account',
    'password',
    decode('UkRQQWRtaW5QcGFzcw==', 'base64'),
    decode('cmRwbWFjcm9ubw==', 'base64'),
    decode('UkRQQWNj',
 'base64'),
    1,
    NOW() - INTERVAL '23 days'
);

INSERT INTO credentials (id, owner_id, name, type, encrypted_data, nonce, salt, version, created_at)
VALUES (
    '20000000-0000-0000-0000-000000000005',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'VNC Viewer Account',
    'password',
    decode('Vk5DVmlld2Vy', 'base64'),
    decode('dm5jbWluYw==', 'base64'),
    decode('Vk5D', 'base64'),
    1,
    NOW() - INTERVAL '22 days'
);

-- Production Web hosts
INSERT INTO hosts (id, owner_id, name, address, protocol, port, credentials_id, tags, group_path, description, is_active, created_at)
VALUES (
    '30000000-0000-0000-0000-000000000001',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'web-prod-01',
    '192.168.1.10'::inet,
    'ssh',
    22,
    '20000000-0000-0000-0000-000000000001',
    ARRAY['web', 'production', 'nginx'],
    'production.web',
    'Primary production web server',
    TRUE,
    NOW() - INTERVAL '25 days'
);

INSERT INTO hosts (id, owner_id, name, address, protocol, port, credentials_id, tags, group_path, description, is_active, created_at)
VALUES (
    '30000000-0000-0000-0000-000000000002',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'web-prod-02',
    '192.168.1.11'::inet,
    'ssh',
    22,
    '20000000-0000-0000-0000-000000000001',
    ARRAY['web', 'production', 'nginx'],
    'production.web',
    'Secondary production web server',
    TRUE,
    NOW() - INTERVAL '25 days'
);

-- Production DB hosts
INSERT INTO hosts (id, owner_id, name, address, protocol, port, credentials_id, tags, group_path, description, is_active, created_at)
VALUES (
    '30000000-0000-0000-0000-000000000003',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'db-prod-primary',
    '192.168.1.20'::inet,
    'ssh',
    22,
    '20000000-0000-0000-0000-000000000002',
    ARRAY['database', 'production', 'postgres'],
    'production.db',
    'Primary PostgreSQL database',
    TRUE,
    NOW() - INTERVAL '24 days'
);

INSERT INTO hosts (id, owner_id, name, address, protocol, port, credentials_id, tags, group_path, description, is_active, created_at)
VALUES (
    '30000000-0000-0000-0000-000000000004',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'db-prod-replica',
    '192.168.1.21'::inet,
    'ssh',
    22,
    '20000000-0000-0000-0000-000000000002',
    ARRAY['database', 'production', 'postgres'],
    'production.db',
    'Read replica PostgreSQL',
    TRUE,
    NOW() - INTERVAL '24 days'
);

INSERT INTO hosts (id, owner_id, name, address, protocol, port, credentials_id, tags, group_path, description, is_active, created_at)
VALUES (
    '30000000-0000-0000-0000-000000000005',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'db-prod-rdp',
    '192.168.1.22'::inet,
    'rdp',
    3389,
    '20000000-0000-0000-0000-000000000004',
    ARRAY['database', 'production', 'windows', 'rdp'],
    'production.db',
    'Windows RDS for database administration',
    TRUE,
    NOW() - INTERVAL '23 days'
);

-- Staging hosts
INSERT INTO hosts (id, owner_id, name, address, protocol, port, credentials_id, tags, group_path, description, is_active, created_at)
VALUES (
    '30000000-0000-0000-0000-000000000006',
    'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22',
    'staging-web-01',
    '10.0.1.10'::inet,
    'ssh',
    22,
    '20000000-0000-0000-0000-000000000003',
    ARRAY['web', 'staging'],
    'staging',
    'Staging web server',
    TRUE,
    NOW() - INTERVAL '20 days'
);

INSERT INTO hosts (id, owner_id, name, address, protocol, port, credentials_id, tags, group_path, description, is_active, created_at)
VALUES (
    '30000000-0000-0000-0000-000000000007',
    'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22',
    'staging-db-01',
    '10.0.1.20'::inet,
    'ssh',
    22,
    '20000000-0000-0000-0000-000000000003',
    ARRAY['database', 'staging'],
    'staging',
    'Staging database server',
    TRUE,
    NOW() - INTERVAL '20 days'
);

-- Development hosts
INSERT INTO hosts (id, owner_id, name, address, protocol, port, tags, group_path, description, is_active, created_at)
VALUES (
    '30000000-0000-0000-0000-000000000008',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'dev-workstation',
    '192.168.100.10'::inet,
    'ssh',
    22,
    ARRAY['dev', 'local'],
    'development',
    'Development workstation',
    TRUE,
    NOW() - INTERVAL '15 days'
);

INSERT INTO hosts (id, owner_id, name, address, protocol, port, tags, group_path, description, is_active, created_at)
VALUES (
    '30000000-0000-0000-0000-000000000009',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'dev-vnc',
    '192.168.100.20'::inet,
    'vnc',
    5900,
    ARRAY['dev', 'vnc', 'remote-desktop'],
    'development',
    'VNC server for remote Dev GUI access',
    TRUE,
    NOW() - INTERVAL '15 days'
);

-- An inactive host for testing /acl
INSERT INTO hosts (id, owner_id, name, address, protocol, port, tags, group_path, description, is_active, created_at)
VALUES (
    '30000000-0000-0000-0000-000000000010',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'decommissioned-server',
    '192.168.1.99'::inet,
    'ssh',
    22,
    ARRAY['decommissioned', 'inactive'],
    'production.web',
    'Old server ready for retirement',
    FALSE,
    NOW() - INTERVAL '60 days'
);

COMMIT;
