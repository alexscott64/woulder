-- Migration 000020: Add advanced priority metrics for hybrid multi-signal system
-- Enables seasonal detection, activity surges, and per-area percentile ranking

-- Add new columns for advanced priority calculation
ALTER TABLE woulder.mp_routes
ADD COLUMN tick_count_14d INTEGER DEFAULT 0,        -- Ticks in last 14 days (surge detection)
ADD COLUMN area_percentile NUMERIC(5,4);            -- Percentile rank within area (0.0-1.0)

-- Add index for efficient percentile queries
CREATE INDEX idx_mp_routes_area_activity ON woulder.mp_routes(mp_area_id, tick_count_90d);

-- Comment on new columns
COMMENT ON COLUMN woulder.mp_routes.tick_count_14d IS 'Number of ticks in last 14 days, used for activity surge detection (season starts)';
COMMENT ON COLUMN woulder.mp_routes.area_percentile IS 'Percentile rank (0.0-1.0) within area based on 90-day activity. Used for per-capita adjustment.';
