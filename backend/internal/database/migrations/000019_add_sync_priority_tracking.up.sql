-- Migration 000019: Add sync priority tracking for optimized tick/comment syncing
-- This enables a dual-track sync system:
-- - Location routes (location_id NOT NULL): Always sync daily
-- - Non-location routes (location_id NULL): Priority-based syncing (high/medium/low)

-- Add sync priority tracking columns to mp_routes
ALTER TABLE woulder.mp_routes
ADD COLUMN sync_priority VARCHAR(10) DEFAULT 'medium',    -- 'high', 'medium', 'low'
ADD COLUMN last_tick_sync_at TIMESTAMPTZ,                 -- When we last checked for new ticks
ADD COLUMN last_comment_sync_at TIMESTAMPTZ,              -- When we last checked for new comments
ADD COLUMN tick_count_90d INTEGER DEFAULT 0,              -- Ticks in last 90 days (for scoring)
ADD COLUMN total_tick_count INTEGER DEFAULT 0,            -- Lifetime tick count (for scoring)
ADD COLUMN days_since_last_tick INTEGER;                  -- Days since most recent tick

-- Indexes for efficient priority-based queries
CREATE INDEX idx_mp_routes_sync_priority ON woulder.mp_routes(sync_priority, last_tick_sync_at);
CREATE INDEX idx_mp_routes_location_priority ON woulder.mp_routes(location_id, sync_priority);
CREATE INDEX idx_mp_routes_comment_sync ON woulder.mp_routes(sync_priority, last_comment_sync_at);

-- Comment on new columns for documentation
COMMENT ON COLUMN woulder.mp_routes.sync_priority IS 'Dynamic priority tier (high/medium/low) based on recent activity. Only applies to routes with location_id NULL.';
COMMENT ON COLUMN woulder.mp_routes.last_tick_sync_at IS 'Timestamp when we last checked Mountain Project API for new ticks';
COMMENT ON COLUMN woulder.mp_routes.last_comment_sync_at IS 'Timestamp when we last checked Mountain Project API for new comments';
COMMENT ON COLUMN woulder.mp_routes.tick_count_90d IS 'Number of ticks in last 90 days, used for priority calculation';
COMMENT ON COLUMN woulder.mp_routes.total_tick_count IS 'Lifetime tick count, used for priority calculation';
COMMENT ON COLUMN woulder.mp_routes.days_since_last_tick IS 'Days since most recent tick, used for priority calculation';
