-- Migration 000016: Convert Mountain Project ID columns from VARCHAR to BIGINT
-- This improves performance and reduces storage for MP IDs

-- Step 1: Drop foreign key constraints (will be recreated after conversion)
ALTER TABLE woulder.mp_routes DROP CONSTRAINT IF EXISTS mp_routes_mp_area_id_fkey;
ALTER TABLE woulder.mp_ticks DROP CONSTRAINT IF EXISTS mp_ticks_mp_route_id_fkey;
ALTER TABLE woulder.boulder_drying_profiles DROP CONSTRAINT IF EXISTS boulder_drying_profiles_mp_route_id_fkey;

-- Step 2: Convert mp_areas columns
-- Convert mp_area_id first (no dependencies on it yet)
ALTER TABLE woulder.mp_areas
    ALTER COLUMN mp_area_id TYPE BIGINT USING mp_area_id::BIGINT;

-- Convert parent_mp_area_id
ALTER TABLE woulder.mp_areas
    ALTER COLUMN parent_mp_area_id TYPE BIGINT USING parent_mp_area_id::BIGINT;

-- Step 3: Convert mp_routes columns
-- Convert mp_route_id first
ALTER TABLE woulder.mp_routes
    ALTER COLUMN mp_route_id TYPE BIGINT USING mp_route_id::BIGINT;

-- Convert mp_area_id (references mp_areas.mp_area_id)
ALTER TABLE woulder.mp_routes
    ALTER COLUMN mp_area_id TYPE BIGINT USING mp_area_id::BIGINT;

-- Step 4: Convert mp_ticks.mp_route_id
ALTER TABLE woulder.mp_ticks
    ALTER COLUMN mp_route_id TYPE BIGINT USING mp_route_id::BIGINT;

-- Step 5: Convert mp_comments columns
ALTER TABLE woulder.mp_comments
    ALTER COLUMN mp_comment_id TYPE BIGINT USING mp_comment_id::BIGINT;

ALTER TABLE woulder.mp_comments
    ALTER COLUMN mp_area_id TYPE BIGINT USING mp_area_id::BIGINT;

ALTER TABLE woulder.mp_comments
    ALTER COLUMN mp_route_id TYPE BIGINT USING mp_route_id::BIGINT;

-- Step 6: Convert boulder_drying_profiles.mp_route_id
ALTER TABLE woulder.boulder_drying_profiles
    ALTER COLUMN mp_route_id TYPE BIGINT USING mp_route_id::BIGINT;

-- Step 7: Recreate foreign key constraints
ALTER TABLE woulder.mp_routes
    ADD CONSTRAINT mp_routes_mp_area_id_fkey
    FOREIGN KEY (mp_area_id) REFERENCES woulder.mp_areas(mp_area_id) ON DELETE CASCADE;

ALTER TABLE woulder.mp_ticks
    ADD CONSTRAINT mp_ticks_mp_route_id_fkey
    FOREIGN KEY (mp_route_id) REFERENCES woulder.mp_routes(mp_route_id) ON DELETE CASCADE;

ALTER TABLE woulder.boulder_drying_profiles
    ADD CONSTRAINT boulder_drying_profiles_mp_route_id_fkey
    FOREIGN KEY (mp_route_id) REFERENCES woulder.mp_routes(mp_route_id) ON DELETE CASCADE;

-- Step 8: Recreate indexes for better performance with BIGINT
-- Drop old indexes
DROP INDEX IF EXISTS woulder.idx_mp_areas_mp_area_id;
DROP INDEX IF EXISTS woulder.idx_mp_areas_parent;
DROP INDEX IF EXISTS woulder.idx_mp_routes_mp_route_id;
DROP INDEX IF EXISTS woulder.idx_mp_routes_area;
DROP INDEX IF EXISTS woulder.idx_mp_ticks_route;
DROP INDEX IF EXISTS woulder.idx_mp_comments_area;
DROP INDEX IF EXISTS woulder.idx_mp_comments_route;
DROP INDEX IF EXISTS woulder.idx_boulder_drying_mp_route_id;

-- Recreate indexes optimized for BIGINT
CREATE INDEX idx_mp_areas_mp_area_id ON woulder.mp_areas(mp_area_id);
CREATE INDEX idx_mp_areas_parent ON woulder.mp_areas(parent_mp_area_id) WHERE parent_mp_area_id IS NOT NULL;
CREATE INDEX idx_mp_routes_mp_route_id ON woulder.mp_routes(mp_route_id);
CREATE INDEX idx_mp_routes_area ON woulder.mp_routes(mp_area_id);
CREATE INDEX idx_mp_ticks_route ON woulder.mp_ticks(mp_route_id);
CREATE INDEX idx_mp_comments_area ON woulder.mp_comments(mp_area_id) WHERE comment_type = 'area';
CREATE INDEX idx_mp_comments_route ON woulder.mp_comments(mp_route_id) WHERE comment_type = 'route';
CREATE INDEX idx_boulder_drying_mp_route_id ON woulder.boulder_drying_profiles(mp_route_id);

-- Add comments to document the change
COMMENT ON COLUMN woulder.mp_areas.mp_area_id IS 'Mountain Project area ID (integer, converted from VARCHAR for performance)';
COMMENT ON COLUMN woulder.mp_areas.parent_mp_area_id IS 'Parent Mountain Project area ID (integer)';
COMMENT ON COLUMN woulder.mp_routes.mp_route_id IS 'Mountain Project route ID (integer, converted from VARCHAR for performance)';
COMMENT ON COLUMN woulder.mp_routes.mp_area_id IS 'Mountain Project area ID (integer)';
COMMENT ON COLUMN woulder.mp_ticks.mp_route_id IS 'Mountain Project route ID (integer)';
COMMENT ON COLUMN woulder.mp_comments.mp_comment_id IS 'Mountain Project comment ID (integer)';
COMMENT ON COLUMN woulder.boulder_drying_profiles.mp_route_id IS 'Mountain Project route ID (integer)';
