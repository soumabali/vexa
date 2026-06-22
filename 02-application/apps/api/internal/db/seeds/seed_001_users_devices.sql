-- seed_001_users_devices.sql
-- Seed data: admin, operator, and viewer users with their devices

BEGIN;

-- Password for ALL seed users is: Vexa2026!Secret (16+ chars, satisfies all policies)

-- Admin user
INSERT INTO users (id, email, password_hash, role, mfa_enabled, is_active, totp_enabled, created_at)
VALUES (
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'admin@vexa.local',
    '$argon2id$v=19$m=65536,t=3,p=4$' || encode(gen_random_bytes(32), 'base64') || '$' || encode(gen_random_bytes(32), 'base64'),
    'admin',
    FALSE,
    TRUE,
    FALSE,
    NOW() - INTERVAL '30 days'
);

-- Operator user
INSERT INTO users (id, email, password_hash, role, mfa_enabled, is_active, totp_enabled, created_at)
VALUES (
    'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22',
    'operator@vexa.local',
    '$argon2id$v=19$m=65536,t=3,p=4$' || encode(gen_random_bytes(32), 'base64') || '$' || encode(gen_random_bytes(32), 'base64'),
    'operator',
    FALSE,
    TRUE,
    FALSE,
    NOW() - INTERVAL '20 days'
);

-- Viewer user
INSERT INTO users (id, email, password_hash, role, mfa_enabled, is_active, totp_enabled, created_at)
VALUES (
    'c2eebc99-9c0b-4ef8-bb6d-6bb9bd380a33',
    'viewer@vexa.local',
    '$argon2id$v=19$m=65536,t=3,p=4$' || encode(gen_random_bytes(32), 'base64') || '$' || encode(gen_random_bytes(32), 'base64'),
    'viewer',
    FALSE,
    TRUE,
    FALSE,
    NOW() - INTERVAL '10 days'
);

-- Inactive user (for testing)
INSERT INTO users (id, email, password_hash, role, mfa_enabled, is_active, totp_enabled, created_at)
VALUES (
    'd3eebc99-9c0b-4ef8-bb6d-6bb9bd380a44',
    'inactive@vexa.local',
    '$argon2id$v=19$m=65536,t=3,p=4$' || encode(gen_random_bytes(32), 'base64') || '$' || encode(gen_random_bytes(32), 'base64'),
    'viewer',
    FALSE,
    FALSE,
    FALSE,
    NOW() - INTERVAL '5 days'
);

-- Admin's devices
INSERT INTO devices (id, user_id, name, fingerprint, public_key, trusted, last_seen, created_at)
VALUES (
    'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a55',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'Admin MacBook Pro',
    'SHA256:abcd1234efgh5678ijkl9012mnop3456qrst7890uvwx',
    decode('aGVsbG8gd29ybGQga2V5', 'base64'),
    TRUE,
    NOW(),
    NOW() - INTERVAL '25 days'
);

INSERT INTO devices (id, user_id, name, fingerprint, public_key, trusted, last_seen, created_at)
VALUES (
    'f5eebc99-9c0b-4ef8-bb6d-6bb9bd380a66',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'Admin iPhone',
    'SHA256:bcde2345fghi6789jklm0123nop4567qrst8901uvwx',
    decode('aGVsbG8gd29ybGQga2V5', 'base64'),
    TRUE,
    NOW(),
    NOW() - INTERVAL '15 days'
);

-- Operator's device
INSERT INTO devices (id, user_id, name, fingerprint, public_key, trusted, last_seen, created_at)
VALUES (
    'g6eebc99-9c0b-4ef8-bb6d-6bb9bd380a77',
    'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22',
    'Operator Linux Workstation',
    'SHA256:cdef3456ghij7890klmn1234opqr5678rst9012vwxy',
    decode('aGVsbG8gd29ybGQga2V5', 'base64'),
    TRUE,
    NOW(),
    NOW() - INTERVAL '18 days'
);

-- Viewer's device
INSERT INTO devices (id, user_id, name, fingerprint, public_key, trusted, last_seen, created_at)
VALUES (
    'h7eebc99-9c0b-4ef8-bb6d-6bb9bd380a88',
    'c2eebc99-9c0b-4ef8-bb6d-6bb9bd380a33',
    'Viewer Windows Desktop',
    'SHA256:defg4567hijk8901lmno2345pqrs6789tuvw9012xyz',
    decode('aGVsbG8gd29ybGQga2V5', 'base64'),
    TRUE,
    NOW(),
    NOW() - INTERVAL '8 days'
);

COMMIT;
