-- Migration 000026 rollback: Remove detailed route information fields

-- Drop indexes
DROP INDEX IF EXISTS woulder.idx_mp_routes_rating;
DROP INDEX IF EXISTS woulder.idx_mp_routes_difficulty;

-- Remove columns from mp_routes table
ALTER TABLE woulder.mp_routes
    DROP COLUMN IF EXISTS difficulty,
    DROP COLUMN IF EXISTS pitches,
    DROP COLUMN IF EXISTS height_feet,
    DROP COLUMN IF EXISTS mp_rating,
    DROP COLUMN IF EXISTS popularity,
    DROP COLUMN IF EXISTS description_text,
    DROP COLUMN IF EXISTS location_text,
    DROP COLUMN IF EXISTS protection_text,
    DROP COLUMN IF EXISTS safety_text;
