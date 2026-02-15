-- Migration 000022: Add job monitoring system
-- Track background job executions and progress for monitoring

CREATE TABLE woulder.job_executions (
    id SERIAL PRIMARY KEY,
    job_name VARCHAR(100) NOT NULL,          -- 'high_priority_tick_sync', 'medium_priority_comment_sync', etc.
    job_type VARCHAR(50) NOT NULL,           -- 'tick_sync', 'comment_sync', 'priority_calc', 'route_backfill'
    status VARCHAR(20) NOT NULL,              -- 'running', 'completed', 'failed', 'cancelled'
    total_items INTEGER DEFAULT 0,            -- Total items to process
    items_processed INTEGER DEFAULT 0,        -- Items completed so far
    items_succeeded INTEGER DEFAULT 0,        -- Items that succeeded
    items_failed INTEGER DEFAULT 0,           -- Items that failed
    error_message TEXT,
    started_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    
    -- Progress metadata (JSON) - store job-specific data
    metadata JSONB DEFAULT '{}'::jsonb
);

-- Indexes for efficient queries
CREATE INDEX idx_job_executions_status ON woulder.job_executions(status, started_at DESC);
CREATE INDEX idx_job_executions_job_name ON woulder.job_executions(job_name, started_at DESC);
CREATE INDEX idx_job_executions_active ON woulder.job_executions(status) WHERE status = 'running';
CREATE INDEX idx_job_executions_started_at ON woulder.job_executions(started_at DESC);

-- Auto-update trigger for updated_at
CREATE TRIGGER update_job_executions_updated_at 
    BEFORE UPDATE ON woulder.job_executions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Comments for documentation
COMMENT ON TABLE woulder.job_executions IS 'Tracks background job execution and progress for monitoring';
COMMENT ON COLUMN woulder.job_executions.job_name IS 'Unique identifier for job type + priority (e.g., high_priority_tick_sync)';
COMMENT ON COLUMN woulder.job_executions.job_type IS 'Category of job (tick_sync, comment_sync, priority_calc, route_backfill)';
COMMENT ON COLUMN woulder.job_executions.status IS 'Current status: running, completed, failed, cancelled';
COMMENT ON COLUMN woulder.job_executions.metadata IS 'Job-specific metadata (priority level, routes/sec, etc.)';
