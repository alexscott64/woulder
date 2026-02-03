-- Migration 000013: Add indexes for state-level Mountain Project queries
-- Optimize queries for state-level data (location_id IS NULL)

-- Composite index for area hierarchy traversal with NULL locations
-- This speeds up queries that traverse the MP area hierarchy for state data
CREATE INDEX idx_mp_areas_location_parent ON woulder.mp_areas(location_id, parent_mp_area_id)
  WHERE location_id IS NULL;

-- Index for route queries without locations
-- Speeds up queries that find routes for state areas
CREATE INDEX idx_mp_routes_location_null ON woulder.mp_routes(mp_area_id)
  WHERE location_id IS NULL;

-- Optimize tick sync (find latest tick per route)
-- Used by incremental sync to find the last tick timestamp for each route
CREATE INDEX idx_mp_ticks_route_climbed ON woulder.mp_ticks(mp_route_id, climbed_at DESC);

-- Text search for area names (optional - for future search features)
-- Enable fuzzy text search across all MP area names
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_mp_areas_name_trgm ON woulder.mp_areas USING gin(name gin_trgm_ops);
