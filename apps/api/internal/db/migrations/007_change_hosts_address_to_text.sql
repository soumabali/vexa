-- 007_change_hosts_address_to_text.sql
-- Change hosts.address from INET to TEXT to support hostnames and IPs
ALTER TABLE hosts ALTER COLUMN address TYPE TEXT;
