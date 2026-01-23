-- Rollback Migration 000008: Remove boulder GPS coordinates and drying profiles

-- Drop boulder_drying_profiles table
DROP TRIGGER IF EXISTS update_boulder_drying_profiles_updated_at ON woulder.boulder_drying_profiles;
DROP TABLE IF EXISTS woulder.boulder_drying_profiles;

-- Drop GPS indexes
DROP INDEX IF EXISTS woulder.idx_mp_routes_lat_lon;
DROP INDEX IF EXISTS woulder.idx_mp_areas_lat_lon;

-- Remove GPS columns from mp_routes
ALTER TABLE woulder.mp_routes
    DROP COLUMN IF EXISTS aspect,
    DROP COLUMN IF EXISTS longitude,
    DROP COLUMN IF EXISTS latitude;

-- Remove GPS columns from mp_areas
ALTER TABLE woulder.mp_areas
    DROP COLUMN IF EXISTS longitude,
    DROP COLUMN IF EXISTS latitude;
