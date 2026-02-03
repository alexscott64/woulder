-- Migration 000014: Add sync progress tracking table
-- Track sync progress for recovery and monitoring

CREATE TABLE woulder.sync_progress (
    id SERIAL PRIMARY KEY,
    area_id INTEGER REFERENCES woulder.areas(id),    -- Nullable - not used for state syncs
    mp_area_id VARCHAR(50) NOT NULL,
    sync_type VARCHAR(20) NOT NULL, -- 'full' or 'incremental'
    status VARCHAR(20) NOT NULL,     -- 'pending', 'in_progress', 'completed', 'failed'
    routes_synced INTEGER DEFAULT 0,
    ticks_synced INTEGER DEFAULT 0,
    areas_synced INTEGER DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for efficient progress queries
CREATE INDEX idx_sync_progress_status ON woulder.sync_progress(status, started_at DESC);
CREATE INDEX idx_sync_progress_area ON woulder.sync_progress(area_id, created_at DESC);
CREATE INDEX idx_sync_progress_mp_area ON woulder.sync_progress(mp_area_id);

-- Add auto-update trigger for updated_at
CREATE TRIGGER update_sync_progress_updated_at BEFORE UPDATE ON woulder.sync_progress
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
