-- Migration: Session recordings (asciinema format)
-- Created: 2026-05-28

CREATE TYPE recording_status AS ENUM ('recording', 'paused', 'completed', 'failed', 'deleted');

CREATE TABLE recordings (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id      UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    host_id         UUID NOT NULL REFERENCES hosts(id) ON DELETE CASCADE,
    
    -- Recording metadata
    status          recording_status NOT NULL DEFAULT 'recording',
    started_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at        TIMESTAMPTZ,
    duration_ms     INTEGER, -- actual duration in milliseconds
    
    -- File info
    filename        TEXT NOT NULL, -- e.g., "session-uuid-timestamp.cast"
    file_size_bytes BIGINT,
    file_path       TEXT, -- S3 key or local path
    checksum_sha256 TEXT, -- integrity check
    
    -- asciinema v2 header
    asciinema_version INTEGER DEFAULT 2,
    terminal_width  INTEGER,
    terminal_height INTEGER,
    terminal_type     TEXT DEFAULT 'xterm-256color',
    shell           TEXT DEFAULT '/bin/bash',
    
    -- Content indexing for search
    content_index   TSVECTOR, -- full-text search index
    command_history TEXT[], -- extracted commands
    
    -- Encryption
    encrypted       BOOLEAN DEFAULT true,
    encryption_key_id TEXT, -- reference to key management
    
    -- Retention
    retention_days  INTEGER DEFAULT 90,
    expires_at      TIMESTAMPTZ,
    
    -- Audit
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,
    deleted_by      UUID REFERENCES users(id)
);

-- Indexes for common queries
CREATE INDEX idx_recordings_session ON recordings(session_id);
CREATE INDEX idx_recordings_user ON recordings(user_id);
CREATE INDEX idx_recordings_host ON recordings(host_id);
CREATE INDEX idx_recordings_status ON recordings(status);
CREATE INDEX idx_recordings_started ON recordings(started_at DESC);
CREATE INDEX idx_recordings_expires ON recordings(expires_at);
CREATE INDEX idx_recordings_content_search ON recordings USING GIN(content_index);

-- Partition by month for performance (optional, for high volume)
-- CREATE TABLE recordings_2026_01 PARTITION OF recordings
--     FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

-- Recording events (for detailed timeline)
CREATE TABLE recording_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recording_id    UUID NOT NULL REFERENCES recordings(id) ON DELETE CASCADE,
    event_time      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    event_type      TEXT NOT NULL, -- 'output', 'resize', 'marker', 'command', 'error'
    event_data      JSONB, -- event-specific data
    offset_ms       INTEGER -- offset from recording start in ms
);

CREATE INDEX idx_recording_events_recording ON recording_events(recording_id);
CREATE INDEX idx_recording_events_type ON recording_events(event_type);

-- Recording bookmarks (user annotations)
CREATE TABLE recording_bookmarks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recording_id    UUID NOT NULL REFERENCES recordings(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    offset_ms       INTEGER NOT NULL,
    label           TEXT,
    color           TEXT DEFAULT '#FFD700',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bookmarks_recording ON recording_bookmarks(recording_id);

-- Update trigger
CREATE OR REPLACE FUNCTION update_recordings_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER recordings_updated_at
    BEFORE UPDATE ON recordings
    FOR EACH ROW
    EXECUTE FUNCTION update_recordings_updated_at();

-- RLS Policies
ALTER TABLE recordings ENABLE ROW LEVEL SECURITY;
ALTER TABLE recording_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE recording_bookmarks ENABLE ROW LEVEL SECURITY;

-- Recordings: users can see their own, admins can see all
CREATE POLICY recordings_user_isolation ON recordings
    FOR ALL
    USING (user_id = current_setting('app.current_user_id')::UUID
           OR current_setting('app.is_admin')::BOOLEAN);

-- Recording events: same as recordings
CREATE POLICY recording_events_user_isolation ON recording_events
    FOR ALL
    USING (EXISTS (
        SELECT 1 FROM recordings r
        WHERE r.id = recording_events.recording_id
          AND (r.user_id = current_setting('app.current_user_id')::UUID
               OR current_setting('app.is_admin')::BOOLEAN)
    ));

-- Bookmarks: users manage their own
CREATE POLICY bookmarks_user_isolation ON recording_bookmarks
    FOR ALL
    USING (user_id = current_setting('app.current_user_id')::UUID
           OR current_setting('app.is_admin')::BOOLEAN);

-- Comments / documentation
COMMENT ON TABLE recordings IS 'Session recordings in asciinema v2 format';
COMMENT ON COLUMN recordings.content_index IS 'Full-text search index of terminal output content';
COMMENT ON COLUMN recordings.command_history IS 'Extracted shell commands from the session';
