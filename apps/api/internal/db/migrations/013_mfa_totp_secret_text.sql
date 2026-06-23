-- 013_mfa_totp_secret_text.sql
-- Store TOTP secrets as TEXT so the application can read/write them as
-- base64-encoded encrypted strings without bytea driver mismatches.

ALTER TABLE users ALTER COLUMN totp_secret TYPE TEXT USING totp_secret::TEXT;
