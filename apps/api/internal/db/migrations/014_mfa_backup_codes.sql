-- 014_mfa_backup_codes.sql
-- Persist hashed backup codes on the user row so they can be regenerated
-- and survive across the original MFA setup session in Redis.
--
-- Format: JSON array of bcrypt-style hashes; verification compares with
-- the same hashing scheme used at generation time.

ALTER TABLE users ADD COLUMN IF NOT EXISTS backup_codes_hashes JSONB DEFAULT '[]'::JSONB;