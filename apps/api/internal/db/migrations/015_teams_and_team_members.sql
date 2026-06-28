-- Migration 015: Teams & Team Members
-- Adds support for team-based credential sharing (P6 P6.5 hardening)
-- Run order: 015 (after 014)
-- Rollback: ROLLBACK_015.sql

-- ============================================================
-- Tables: Teams & Team Members
-- ============================================================

CREATE TABLE IF NOT EXISTS teams (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name         VARCHAR(255) NOT NULL,
    description  TEXT,
    owner_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ DEFAULT now(),
    updated_at   TIMESTAMPTZ DEFAULT now(),
    is_active    BOOLEAN DEFAULT true
);

CREATE INDEX IF NOT EXISTS idx_teams_owner_id ON teams(owner_id);
CREATE INDEX IF NOT EXISTS idx_teams_active   ON teams(is_active) WHERE is_active = true;

CREATE TABLE IF NOT EXISTS team_members (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id     UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role        VARCHAR(20) NOT NULL DEFAULT 'member'
                    CHECK (role IN ('owner','admin','member','viewer')),
    joined_at   TIMESTAMPTZ DEFAULT now(),
    added_by    UUID REFERENCES users(id),
    CONSTRAINT unique_team_member UNIQUE (team_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_team_members_team_id ON team_members(team_id);
CREATE INDEX IF NOT EXISTS idx_team_members_user_id ON team_members(user_id);

-- ============================================================
-- Comments
-- ============================================================

COMMENT ON TABLE teams IS
  'Team/organization. Owner is the user who created the team. Soft-deleted via is_active=false. Rollback with ROLLBACK_015.sql';

COMMENT ON TABLE team_members IS
  'Team membership. One row per (team, user). Role determines permission level.';
