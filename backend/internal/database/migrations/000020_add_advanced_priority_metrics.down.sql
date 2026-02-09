-- Rollback migration 000020: Remove advanced priority metrics

-- Drop index
DROP INDEX IF EXISTS woulder.idx_mp_routes_area_activity;

-- Drop columns
ALTER TABLE woulder.mp_routes
DROP COLUMN IF EXISTS tick_count_14d,
DROP COLUMN IF EXISTS area_percentile;
