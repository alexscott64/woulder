-- Rollback migration for Kaya <-> Mountain Project route matching

-- Drop trigger first
DROP TRIGGER IF EXISTS update_kaya_mp_matches_timestamp ON kaya_mp_route_matches;

-- Drop function
DROP FUNCTION IF EXISTS update_kaya_mp_matches_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_kaya_mp_matches_type;
DROP INDEX IF EXISTS idx_kaya_mp_matches_verified;
DROP INDEX IF EXISTS idx_kaya_mp_matches_confidence;
DROP INDEX IF EXISTS idx_kaya_mp_matches_mp_route;
DROP INDEX IF EXISTS idx_kaya_mp_matches_kaya_climb;

-- Drop table
DROP TABLE IF EXISTS kaya_mp_route_matches;
