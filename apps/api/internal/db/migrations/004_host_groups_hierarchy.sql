-- 004_host_groups_hierarchy.sql
-- Enhanced host groups with LTREE support for efficient hierarchical queries

BEGIN;

-- Add ltree column for efficient path queries (optional enhancement)
-- This adds a computed/generated column using the text path

-- Create function to convert text path to ltree-compatible format
CREATE OR REPLACE FUNCTION path_to_ltree(text)
RETURNS ltree AS $$
BEGIN
    -- Replace dots and hyphens with underscores to avoid ltree special chars
    RETURN cast(regexp_replace($1, '[.\-]', '_', 'g') as ltree);
EXCEPTION WHEN OTHERS THEN
    -- Fallback: return simple label
    RETURN cast('root' as ltree);
END;
$$ LANGUAGE plpgsql IMMUTABLE STRICT;

-- Add ltree_path column to host_groups (materialized path)
ALTER TABLE host_groups ADD COLUMN IF NOT EXISTS ltree_path ltree;

-- Backfill ltree_path from existing paths
UPDATE host_groups SET ltree_path = path_to_ltree(path) WHERE ltree_path IS NULL;

-- Add NOT NULL constraint
ALTER TABLE host_groups ALTER COLUMN ltree_path SET NOT NULL;

-- Create GiST index for ltree path queries (ancestors/descendants)
CREATE INDEX IF NOT EXISTS idx_host_groups_ltree_path ON host_groups USING GIST (ltree_path);

-- Create index for path prefix search (e.g., path like 'production.%')
CREATE INDEX IF NOT EXISTS idx_host_groups_path_pattern ON host_groups (path text_pattern_ops);

-- Convenience view for hierarchy
CREATE OR REPLACE VIEW v_host_groups_hierarchy AS
SELECT
    h.id,
    h.owner_id,
    h.name,
    h.path,
    h.parent_id,
    h.created_at,
    (SELECT COUNT(*) FROM host_groups child WHERE child.parent_id = h.id) AS child_count,
    (SELECT COUNT(*) FROM hosts ho WHERE ho.group_path = h.path) AS host_count
FROM host_groups h;

-- Create function to get all descendants of a group by path
CREATE OR REPLACE FUNCTION get_group_descendants(p_path TEXT)
RETURNS TABLE(id UUID, name TEXT, path TEXT, depth INT) AS $$
BEGIN
    RETURN QUERY
    WITH RECURSIVE group_tree AS (
        SELECT h.id, h.name, h.path, 0 AS depth
        FROM host_groups h
        WHERE h.path = p_path
        UNION ALL
        SELECT h.id, h.name, h.path, gt.depth + 1
        FROM host_groups h
        JOIN group_tree gt ON h.parent_id = gt.id
    )
    SELECT gt.id, gt.name, gt.path, gt.depth FROM group_tree gt;
END;
$$ LANGUAGE plpgsql;

-- Record migration
INSERT INTO schema_migrations (version, description)
VALUES ('004_host_groups_hierarchy', 'Add LTREE index and hierarchy helpers for host_groups')
ON CONFLICT (version) DO NOTHING;

COMMIT;
