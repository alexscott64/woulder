-- Migration: 000027_optimize_tick_metrics_query (rollback)
-- Purpose: Rollback composite index optimization

-- Drop the new composite indexes (CONCURRENTLY to avoid blocking)
DROP INDEX CONCURRENTLY IF EXISTS woulder.idx_mp_ticks_route_date_composite;
DROP INDEX CONCURRENTLY IF EXISTS woulder.idx_mp_ticks_location_date;

-- Restore the original partial index (CONCURRENTLY to avoid blocking)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_mp_ticks_route_climbed
    ON woulder.mp_ticks(mp_route_id, climbed_at DESC);

-- Analyze the table to update statistics
ANALYZE woulder.mp_ticks;
