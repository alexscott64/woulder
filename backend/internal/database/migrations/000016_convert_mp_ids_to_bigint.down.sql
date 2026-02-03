-- Migration 000016 DOWN: Revert Mountain Project ID columns from BIGINT to VARCHAR

-- Step 1: Drop foreign key constraints
ALTER TABLE woulder.mp_routes DROP CONSTRAINT IF EXISTS mp_routes_mp_area_id_fkey;
ALTER TABLE woulder.mp_ticks DROP CONSTRAINT IF EXISTS mp_ticks_mp_route_id_fkey;
ALTER TABLE woulder.boulder_drying_profiles DROP CONSTRAINT IF EXISTS boulder_drying_profiles_mp_route_id_fkey;

-- Step 2: Convert columns back to VARCHAR
ALTER TABLE woulder.boulder_drying_profiles
    ALTER COLUMN mp_route_id TYPE VARCHAR(50) USING mp_route_id::VARCHAR;

ALTER TABLE woulder.mp_comments
    ALTER COLUMN mp_route_id TYPE VARCHAR(50) USING mp_route_id::VARCHAR;

ALTER TABLE woulder.mp_comments
    ALTER COLUMN mp_area_id TYPE VARCHAR(50) USING mp_area_id::VARCHAR;

ALTER TABLE woulder.mp_comments
    ALTER COLUMN mp_comment_id TYPE VARCHAR(50) USING mp_comment_id::VARCHAR;

ALTER TABLE woulder.mp_ticks
    ALTER COLUMN mp_route_id TYPE VARCHAR(50) USING mp_route_id::VARCHAR;

ALTER TABLE woulder.mp_routes
    ALTER COLUMN mp_area_id TYPE VARCHAR(50) USING mp_area_id::VARCHAR;

ALTER TABLE woulder.mp_routes
    ALTER COLUMN mp_route_id TYPE VARCHAR(50) USING mp_route_id::VARCHAR;

ALTER TABLE woulder.mp_areas
    ALTER COLUMN parent_mp_area_id TYPE VARCHAR(50) USING parent_mp_area_id::VARCHAR;

ALTER TABLE woulder.mp_areas
    ALTER COLUMN mp_area_id TYPE VARCHAR(50) USING mp_area_id::VARCHAR;

-- Step 3: Recreate foreign key constraints
ALTER TABLE woulder.mp_routes
    ADD CONSTRAINT mp_routes_mp_area_id_fkey
    FOREIGN KEY (mp_area_id) REFERENCES woulder.mp_areas(mp_area_id) ON DELETE CASCADE;

ALTER TABLE woulder.mp_ticks
    ADD CONSTRAINT mp_ticks_mp_route_id_fkey
    FOREIGN KEY (mp_route_id) REFERENCES woulder.mp_routes(mp_route_id) ON DELETE CASCADE;

ALTER TABLE woulder.boulder_drying_profiles
    ADD CONSTRAINT boulder_drying_profiles_mp_route_id_fkey
    FOREIGN KEY (mp_route_id) REFERENCES woulder.mp_routes(mp_route_id) ON DELETE CASCADE;

-- Step 4: Recreate indexes
DROP INDEX IF EXISTS woulder.idx_mp_areas_mp_area_id;
DROP INDEX IF EXISTS woulder.idx_mp_areas_parent;
DROP INDEX IF EXISTS woulder.idx_mp_routes_mp_route_id;
DROP INDEX IF EXISTS woulder.idx_mp_routes_area;
DROP INDEX IF EXISTS woulder.idx_mp_ticks_route;
DROP INDEX IF EXISTS woulder.idx_mp_comments_area;
DROP INDEX IF EXISTS woulder.idx_mp_comments_route;
DROP INDEX IF EXISTS woulder.idx_boulder_drying_mp_route_id;

CREATE INDEX idx_mp_areas_mp_area_id ON woulder.mp_areas(mp_area_id);
CREATE INDEX idx_mp_areas_parent ON woulder.mp_areas(parent_mp_area_id);
CREATE INDEX idx_mp_routes_mp_route_id ON woulder.mp_routes(mp_route_id);
CREATE INDEX idx_mp_routes_area ON woulder.mp_routes(mp_area_id);
CREATE INDEX idx_mp_ticks_route ON woulder.mp_ticks(mp_route_id);
CREATE INDEX idx_mp_comments_area ON woulder.mp_comments(mp_area_id) WHERE comment_type = 'area';
CREATE INDEX idx_mp_comments_route ON woulder.mp_comments(mp_route_id) WHERE comment_type = 'route';
CREATE INDEX idx_boulder_drying_mp_route_id ON woulder.boulder_drying_profiles(mp_route_id);

-- Remove comments
COMMENT ON COLUMN woulder.mp_areas.mp_area_id IS NULL;
COMMENT ON COLUMN woulder.mp_areas.parent_mp_area_id IS NULL;
COMMENT ON COLUMN woulder.mp_routes.mp_route_id IS NULL;
COMMENT ON COLUMN woulder.mp_routes.mp_area_id IS NULL;
COMMENT ON COLUMN woulder.mp_ticks.mp_route_id IS NULL;
COMMENT ON COLUMN woulder.mp_comments.mp_comment_id IS NULL;
COMMENT ON COLUMN woulder.boulder_drying_profiles.mp_route_id IS NULL;
