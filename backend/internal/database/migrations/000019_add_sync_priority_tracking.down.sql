-- Rollback migration 000019: Remove sync priority tracking columns

-- Drop indexes
DROP INDEX IF EXISTS woulder.idx_mp_routes_sync_priority;
DROP INDEX IF EXISTS woulder.idx_mp_routes_location_priority;
DROP INDEX IF EXISTS woulder.idx_mp_routes_comment_sync;

-- Drop columns
ALTER TABLE woulder.mp_routes
DROP COLUMN IF EXISTS sync_priority,
DROP COLUMN IF EXISTS last_tick_sync_at,
DROP COLUMN IF EXISTS last_comment_sync_at,
DROP COLUMN IF EXISTS tick_count_90d,
DROP COLUMN IF EXISTS total_tick_count,
DROP COLUMN IF EXISTS days_since_last_tick;
