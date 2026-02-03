-- Migration 000013: Rollback state-level indexes

DROP INDEX IF EXISTS woulder.idx_mp_areas_name_trgm;
DROP INDEX IF EXISTS woulder.idx_mp_ticks_route_climbed;
DROP INDEX IF EXISTS woulder.idx_mp_routes_location_null;
DROP INDEX IF EXISTS woulder.idx_mp_areas_location_parent;

-- Note: pg_trgm extension is left in place as it may be used by other features
