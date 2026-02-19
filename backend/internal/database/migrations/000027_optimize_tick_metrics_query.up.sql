-- Migration: 000027_optimize_tick_metrics_query
-- Purpose: Add composite index to speed up route_metrics CTE in priority calculation
-- 
-- The route priority calculation (queryUpdateRouteSyncPriorities) performs multiple
-- COUNT aggregations filtered by date ranges (14 days, 90 days) which causes full
-- table scans on mp_ticks. This composite index allows the database to efficiently
-- filter ticks by route_id and climbed_at together.
--
-- Query pattern being optimized:
--   COUNT(CASE WHEN t.climbed_at >= NOW() - INTERVAL '14 days' THEN 1 END)
--   COUNT(CASE WHEN t.climbed_at >= NOW() - INTERVAL '90 days' THEN 1 END)
--
-- Before: Sequential scan on mp_ticks with date filtering in memory
-- After: Index-only scan using composite index for efficient date range filtering

-- Note: We keep the existing index and add a new one instead of dropping first
-- This ensures queries can still use the old index during creation of the new one

-- Create composite index optimized for date range aggregations (non-blocking)
-- Ordering: (mp_route_id, climbed_at DESC)
-- - Groups all ticks by route (efficient for GROUP BY r.mp_route_id)
-- - Sorts by date descending (efficient for MAX(climbed_at) and date range filters)
-- - Allows index-only scans for COUNT operations with date filters
-- CONCURRENTLY allows this to build without blocking writes (takes longer but safer)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_mp_ticks_route_date_composite
    ON woulder.mp_ticks(mp_route_id, climbed_at DESC);

-- Add covering index for location-based queries (used by other parts of the system)
-- This supports queries that filter on location_id and need tick counts
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_mp_ticks_location_date
    ON woulder.mp_ticks(location_id, climbed_at DESC)
    WHERE location_id IS NOT NULL;

-- After new indexes are created, drop the old partial index if it exists
-- This is safe because the new composite index covers the same use cases
DROP INDEX CONCURRENTLY IF EXISTS woulder.idx_mp_ticks_route_climbed;

-- Analyze the table to update statistics for the query planner
ANALYZE woulder.mp_ticks;
