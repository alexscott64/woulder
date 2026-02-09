-- Remove route count tracking columns from mp_areas table
DROP INDEX IF EXISTS woulder.idx_mp_areas_route_count;

ALTER TABLE woulder.mp_areas
DROP COLUMN IF EXISTS route_count_total,
DROP COLUMN IF EXISTS route_count_last_checked;
