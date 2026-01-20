-- Migration 000008: Add boulder GPS coordinates and drying profiles
-- Adds GPS coordinates to areas and routes, plus boulder-specific drying data

-- Step 1: Add GPS coordinates to mp_areas table
ALTER TABLE woulder.mp_areas
    ADD COLUMN latitude DECIMAL(10, 7),
    ADD COLUMN longitude DECIMAL(10, 7);

-- Step 2: Add GPS coordinates and aspect to mp_routes table
ALTER TABLE woulder.mp_routes
    ADD COLUMN latitude DECIMAL(10, 7),
    ADD COLUMN longitude DECIMAL(10, 7),
    ADD COLUMN aspect VARCHAR(20);  -- "N", "NE", "E", "SE", "S", "SW", "W", "NW"

-- Step 3: Create boulder_drying_profiles table
-- Stores boulder-specific drying metadata (tree cover, rock type overrides, cached sun data)
CREATE TABLE woulder.boulder_drying_profiles (
    id SERIAL PRIMARY KEY,
    mp_route_id VARCHAR(50) UNIQUE NOT NULL REFERENCES woulder.mp_routes(mp_route_id) ON DELETE CASCADE,
    tree_coverage_percent DECIMAL(5, 2) CHECK (tree_coverage_percent >= 0 AND tree_coverage_percent <= 100),
    rock_type_override VARCHAR(100),          -- Optional boulder-specific rock type
    last_sun_calc_at TIMESTAMPTZ,             -- Last time sun exposure was calculated
    sun_exposure_hours_cache JSONB,           -- Cache of sun hours by date
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for boulder_drying_profiles
CREATE INDEX idx_boulder_drying_mp_route_id ON woulder.boulder_drying_profiles(mp_route_id);
CREATE INDEX idx_boulder_drying_last_sun_calc ON woulder.boulder_drying_profiles(last_sun_calc_at);

-- Step 4: Add indexes for GPS coordinates (for spatial queries)
CREATE INDEX idx_mp_areas_lat_lon ON woulder.mp_areas(latitude, longitude);
CREATE INDEX idx_mp_routes_lat_lon ON woulder.mp_routes(latitude, longitude);

-- Step 5: Add comment on aspect column for clarity
COMMENT ON COLUMN woulder.mp_routes.aspect IS 'Cardinal direction boulder faces (calculated from position in circular distribution): N, NE, E, SE, S, SW, W, NW';

-- Step 6: Add auto-update trigger for boulder_drying_profiles
CREATE TRIGGER update_boulder_drying_profiles_updated_at BEFORE UPDATE ON woulder.boulder_drying_profiles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
