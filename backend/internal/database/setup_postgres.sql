-- Woulder PostgreSQL Database Setup
-- This file contains the complete schema and seed data
-- Run this file once to set up a new Woulder database
--
-- Usage: psql -h your-host -U your-user -d your-database -f setup_postgres.sql
--
-- PostgreSQL Best Practices:
-- - Uses 'woulder' schema for organization
-- - DECIMAL for precise numbers (lat/lon, temperature, precipitation)
-- - TIMESTAMPTZ for timezone-aware timestamps
-- - snake_case naming (PostgreSQL standard)
-- - SERIAL for auto-incrementing PKs
-- - Explicit CASCADE rules for foreign keys
-- - B-tree indexes on frequently queried columns
-- - CHECK constraints for data validation
-- - Automatic updated_at triggers

-- ============================================================================
-- SCHEMA CREATION
-- ============================================================================

-- Create schema if it doesn't exist
CREATE SCHEMA IF NOT EXISTS woulder;

-- Set search path for this session
SET search_path TO woulder, public;

-- ============================================================================
-- TABLES
-- ============================================================================

-- Areas table
-- Stores geographic areas for grouping climbing locations
CREATE TABLE IF NOT EXISTS woulder.areas (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    region VARCHAR(100),
    display_order INTEGER DEFAULT 0 CHECK (display_order >= 0),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Locations table
-- Stores climbing locations with coordinates and elevation
CREATE TABLE IF NOT EXISTS woulder.locations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    latitude DECIMAL(10, 8) NOT NULL CHECK (latitude >= -90 AND latitude <= 90),
    longitude DECIMAL(11, 8) NOT NULL CHECK (longitude >= -180 AND longitude <= 180),
    elevation_ft INTEGER DEFAULT 0 CHECK (elevation_ft >= -1000 AND elevation_ft <= 30000),
    area_id INTEGER NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (area_id) REFERENCES woulder.areas(id) ON DELETE RESTRICT
);

-- Weather data table
-- Stores historical and forecast weather data
CREATE TABLE IF NOT EXISTS woulder.weather_data (
    id SERIAL PRIMARY KEY,
    location_id INTEGER NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    temperature DECIMAL(5, 2) NOT NULL,
    feels_like DECIMAL(5, 2) NOT NULL,
    precipitation DECIMAL(6, 3) DEFAULT 0 CHECK (precipitation >= 0),
    humidity INTEGER DEFAULT 0 CHECK (humidity >= 0 AND humidity <= 100),
    wind_speed DECIMAL(5, 2) DEFAULT 0 CHECK (wind_speed >= 0),
    wind_direction INTEGER DEFAULT 0 CHECK (wind_direction >= 0 AND wind_direction <= 360),
    cloud_cover INTEGER DEFAULT 0 CHECK (cloud_cover >= 0 AND cloud_cover <= 100),
    pressure INTEGER DEFAULT 0 CHECK (pressure >= 0),
    description VARCHAR(255) DEFAULT '',
    icon VARCHAR(10) DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (location_id) REFERENCES woulder.locations(id) ON DELETE CASCADE,
    UNIQUE(location_id, timestamp)
);

-- Rivers table
-- Stores river crossing information with USGS gauge data
CREATE TABLE IF NOT EXISTS woulder.rivers (
    id SERIAL PRIMARY KEY,
    location_id INTEGER NOT NULL,
    gauge_id VARCHAR(50) NOT NULL,
    river_name VARCHAR(255) NOT NULL,
    safe_crossing_cfs INTEGER NOT NULL CHECK (safe_crossing_cfs > 0),
    caution_crossing_cfs INTEGER NOT NULL CHECK (caution_crossing_cfs >= safe_crossing_cfs),
    drainage_area_sq_mi DECIMAL(10, 2),
    gauge_drainage_area_sq_mi DECIMAL(10, 2),
    flow_divisor DECIMAL(5, 2),
    is_estimated BOOLEAN DEFAULT FALSE,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (location_id) REFERENCES woulder.locations(id) ON DELETE CASCADE
);

-- Rock Type Groups table
-- Stores categories of rock types based on drying characteristics
CREATE TABLE IF NOT EXISTS woulder.rock_type_groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Rock Types table
-- Stores individual rock types with drying characteristics
CREATE TABLE IF NOT EXISTS woulder.rock_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    base_drying_hours DECIMAL(4,1) NOT NULL,
    porosity_percent DECIMAL(4,1),
    is_wet_sensitive BOOLEAN DEFAULT FALSE,
    description TEXT,
    rock_type_group_id INTEGER NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (rock_type_group_id) REFERENCES woulder.rock_type_groups(id) ON DELETE RESTRICT
);

-- Location Rock Types junction table
-- Links locations with their rock types (many-to-many relationship)
CREATE TABLE IF NOT EXISTS woulder.location_rock_types (
    location_id INTEGER NOT NULL,
    rock_type_id INTEGER NOT NULL,
    is_primary BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (location_id, rock_type_id),
    FOREIGN KEY (location_id) REFERENCES woulder.locations(id) ON DELETE CASCADE,
    FOREIGN KEY (rock_type_id) REFERENCES woulder.rock_types(id) ON DELETE CASCADE
);

-- ============================================================================
-- INDEXES
-- ============================================================================

-- Area indexes
CREATE INDEX IF NOT EXISTS idx_areas_display_order ON woulder.areas(display_order);

-- Location indexes
CREATE INDEX IF NOT EXISTS idx_locations_name ON woulder.locations(name);
CREATE INDEX IF NOT EXISTS idx_locations_area_id ON woulder.locations(area_id);

-- Weather data indexes (for performance)
CREATE INDEX IF NOT EXISTS idx_weather_data_location_timestamp ON woulder.weather_data(location_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_weather_data_timestamp ON woulder.weather_data(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_weather_data_created_at ON woulder.weather_data(created_at DESC);

-- River indexes
CREATE INDEX IF NOT EXISTS idx_rivers_location ON woulder.rivers(location_id);
CREATE INDEX IF NOT EXISTS idx_rivers_gauge_id ON woulder.rivers(gauge_id);

-- Rock type indexes
CREATE INDEX IF NOT EXISTS idx_rock_types_group ON woulder.rock_types(rock_type_group_id);
CREATE INDEX IF NOT EXISTS idx_location_rock_types_location ON woulder.location_rock_types(location_id);
CREATE INDEX IF NOT EXISTS idx_location_rock_types_rock_type ON woulder.location_rock_types(rock_type_id);

-- ============================================================================
-- TRIGGERS
-- ============================================================================

-- Function to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION woulder.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger for areas table
DROP TRIGGER IF EXISTS update_areas_updated_at ON woulder.areas;
CREATE TRIGGER update_areas_updated_at BEFORE UPDATE ON woulder.areas
    FOR EACH ROW EXECUTE FUNCTION woulder.update_updated_at_column();

-- Trigger for locations table
DROP TRIGGER IF EXISTS update_locations_updated_at ON woulder.locations;
CREATE TRIGGER update_locations_updated_at BEFORE UPDATE ON woulder.locations
    FOR EACH ROW EXECUTE FUNCTION woulder.update_updated_at_column();

-- Trigger for rivers table
DROP TRIGGER IF EXISTS update_rivers_updated_at ON woulder.rivers;
CREATE TRIGGER update_rivers_updated_at BEFORE UPDATE ON woulder.rivers
    FOR EACH ROW EXECUTE FUNCTION woulder.update_updated_at_column();

-- ============================================================================
-- COMMENTS (Documentation)
-- ============================================================================

COMMENT ON SCHEMA woulder IS 'Woulder climbing weather application schema';
COMMENT ON TABLE woulder.areas IS 'Geographic areas for grouping climbing locations';
COMMENT ON TABLE woulder.locations IS 'Climbing locations with coordinates and elevation';
COMMENT ON TABLE woulder.weather_data IS 'Historical and forecast weather data';
COMMENT ON TABLE woulder.rivers IS 'River crossing information with USGS gauge data';
COMMENT ON TABLE woulder.rock_type_groups IS 'Categories of rock types based on drying characteristics';
COMMENT ON TABLE woulder.rock_types IS 'Individual rock types with drying and porosity data';
COMMENT ON TABLE woulder.location_rock_types IS 'Junction table linking locations to their rock types';
COMMENT ON COLUMN woulder.areas.display_order IS 'Order in which areas should be displayed (lower = first)';
COMMENT ON COLUMN woulder.locations.area_id IS 'Foreign key to areas table';
COMMENT ON COLUMN woulder.rivers.is_estimated IS 'Whether flow is estimated from nearby gauge or direct reading';
COMMENT ON COLUMN woulder.rivers.flow_divisor IS 'Divisor to apply to gauge reading (e.g., 2.0 means divide by 2)';
COMMENT ON COLUMN woulder.rivers.drainage_area_sq_mi IS 'Drainage area of the river at crossing point';
COMMENT ON COLUMN woulder.rivers.gauge_drainage_area_sq_mi IS 'Drainage area at the gauge location';
COMMENT ON COLUMN woulder.rock_types.base_drying_hours IS 'Base hours required to dry after 0.1" of rain';
COMMENT ON COLUMN woulder.rock_types.is_wet_sensitive IS 'True for rocks that are permanently damaged when climbed wet';
COMMENT ON COLUMN woulder.location_rock_types.is_primary IS 'Marks the primary rock type for the location';

-- ============================================================================
-- SEED DATA
-- ============================================================================

-- Areas
-- Create Pacific Northwest and Southern California areas
INSERT INTO woulder.areas (name, description, region, display_order)
VALUES
    ('Pacific Northwest', 'Climbing areas in Washington, Oregon, and British Columbia', 'Northwest', 1),
    ('Southern California', 'Desert climbing and high-alpine bouldering in Southern California', 'Southwest', 2)
ON CONFLICT (name) DO NOTHING;

-- Locations
-- Using ON CONFLICT DO NOTHING to skip if location already exists
DO $$
DECLARE
    pnw_area_id INTEGER;
    socal_area_id INTEGER;
BEGIN
    SELECT id INTO pnw_area_id FROM woulder.areas WHERE name = 'Pacific Northwest';
    SELECT id INTO socal_area_id FROM woulder.areas WHERE name = 'Southern California';

    -- Pacific Northwest locations
    INSERT INTO woulder.locations (name, latitude, longitude, elevation_ft, area_id)
    VALUES
        ('Skykomish - Money Creek', 47.69727769, -121.47884640, 1000, pnw_area_id),
        ('Index', 47.82061272, -121.55492795, 500, pnw_area_id),
        ('Gold Bar', 47.8468, -121.6970, 200, pnw_area_id),
        ('Bellingham', 48.7519, -122.4787, 100, pnw_area_id),
        ('Icicle Creek (Leavenworth)', 47.5962, -120.6615, 1200, pnw_area_id),
        ('Squamish', 49.7016, -123.1558, 200, pnw_area_id),
        ('Skykomish - Paradise', 47.64074805, -121.37822668, 1500, pnw_area_id),
        ('Treasury', 47.76086166, -121.12877297, 3650, pnw_area_id),
        ('Calendar Butte', 48.36202, -122.08273, 1600, pnw_area_id)
    ON CONFLICT (name) DO NOTHING;

    -- Southern California locations
    INSERT INTO woulder.locations (name, latitude, longitude, elevation_ft, area_id)
    VALUES
        ('Joshua Tree', 34.01565, -116.16298, 2700, socal_area_id),
        ('Black Mountain', 33.82629, -116.7591, 7500, socal_area_id),
        ('Buttermilks', 37.3276, -118.5757, 6400, socal_area_id),
        ('Happy / Sad Boulders', 37.41601, -118.43994, 4400, socal_area_id),
        ('Yosemite', 37.7416, -119.60152, 3977, socal_area_id),
        ('Tramway', 33.81074, -116.65175, 8519, socal_area_id)
    ON CONFLICT (name) DO NOTHING;
END $$;

-- Rivers (using active USGS gauges)

-- Money Creek - estimated from South Fork Skykomish at Skykomish (12131500)
INSERT INTO woulder.rivers (location_id, gauge_id, river_name, safe_crossing_cfs, caution_crossing_cfs, drainage_area_sq_mi, gauge_drainage_area_sq_mi, is_estimated, description)
SELECT
    l.id,
    '12131500',
    'Money Creek',
    60,
    90,
    18.0,
    355.0,
    TRUE,
    'Flow estimated from South Fork Skykomish at Skykomish. Watershed includes Lake Elizabeth, Goat Creek, Crosby Mountain, Apex Mine drainage'
FROM woulder.locations l
WHERE l.name = 'Skykomish - Money Creek'
ON CONFLICT DO NOTHING;

-- Index - North Fork Skykomish estimated from Gold Bar gauge (12134500)
-- Per local bouldering guide: divide gauge reading by 2 for North Fork flow
INSERT INTO woulder.rivers (location_id, gauge_id, river_name, safe_crossing_cfs, caution_crossing_cfs, flow_divisor, is_estimated, description)
SELECT
    l.id,
    '12134500',
    'North Fork Skykomish River',
    800,
    900,
    2.0,
    TRUE,
    'River crossing to Sasquatch Boulders. Flow estimated as gauge reading / 2. Below 800 CFS is safe, 800-900 CFS use caution, above 900 CFS is dangerous.'
FROM woulder.locations l
WHERE l.name = 'Index'
ON CONFLICT DO NOTHING;

-- Gold Bar - direct gauge reading (12134500)
INSERT INTO woulder.rivers (location_id, gauge_id, river_name, safe_crossing_cfs, caution_crossing_cfs, is_estimated, description)
SELECT
    l.id,
    '12134500',
    'Skykomish River',
    3000,
    4500,
    FALSE,
    'Direct gauge reading from USGS Skykomish River near Gold Bar'
FROM woulder.locations l
WHERE l.name = 'Gold Bar'
ON CONFLICT DO NOTHING;

-- Paradise - West Fork Miller River estimated from Gold Bar (12134500)
INSERT INTO woulder.rivers (location_id, gauge_id, river_name, safe_crossing_cfs, caution_crossing_cfs, drainage_area_sq_mi, gauge_drainage_area_sq_mi, is_estimated, description)
SELECT
    l.id,
    '12134500',
    'West Fork Miller River',
    600,
    900,
    28.0,
    535.0,
    TRUE,
    'Flow estimated from Skykomish River at Gold Bar. West Fork Miller River drainage area.'
FROM woulder.locations l
WHERE l.name = 'Skykomish - Paradise'
ON CONFLICT DO NOTHING;

-- Paradise - East Fork Miller River estimated from Gold Bar (12134500)
INSERT INTO woulder.rivers (location_id, gauge_id, river_name, safe_crossing_cfs, caution_crossing_cfs, drainage_area_sq_mi, gauge_drainage_area_sq_mi, is_estimated, description)
SELECT
    l.id,
    '12134500',
    'East Fork Miller River',
    700,
    1000,
    22.0,
    535.0,
    TRUE,
    'Flow estimated from Skykomish River at Gold Bar. East Fork Miller River drainage area.'
FROM woulder.locations l
WHERE l.name = 'Skykomish - Paradise'
ON CONFLICT DO NOTHING;

-- Rock Type Groups
INSERT INTO woulder.rock_type_groups (name, description)
VALUES
    ('Wet-Sensitive Rocks', 'Soft rocks that are permanently damaged when climbed wet. DO NOT CLIMB WHEN WET.'),
    ('Fast-Drying Rocks', 'Hard, non-porous rocks that dry quickly after rain.'),
    ('Medium-Drying Rocks', 'Rocks with moderate porosity that take longer to dry.'),
    ('Slow-Drying Rocks', 'Rocks that absorb and retain water, requiring extended drying time.')
ON CONFLICT (name) DO NOTHING;

-- Rock Types
DO $$
BEGIN
    -- Wet-sensitive rocks
    INSERT INTO woulder.rock_types (name, base_drying_hours, porosity_percent, is_wet_sensitive, description, rock_type_group_id)
    VALUES
        ('Sandstone', 36.0, 20.0, TRUE, 'Soft sedimentary rock that absorbs water and becomes friable when wet. DO NOT CLIMB WHEN WET.',
         (SELECT id FROM woulder.rock_type_groups WHERE name = 'Wet-Sensitive Rocks')),
        ('Arkose', 36.0, 18.0, TRUE, 'Feldspar-rich sandstone. Soft and water-absorbent. DO NOT CLIMB WHEN WET.',
         (SELECT id FROM woulder.rock_type_groups WHERE name = 'Wet-Sensitive Rocks')),
        ('Graywacke', 30.0, 15.0, TRUE, 'Hard sandstone with clay matrix. Absorbs water. DO NOT CLIMB WHEN WET.',
         (SELECT id FROM woulder.rock_type_groups WHERE name = 'Wet-Sensitive Rocks')),

        -- Fast-drying rocks
        ('Granite', 6.0, 1.0, FALSE, 'Hard crystalline igneous rock. Non-porous, dries quickly.',
         (SELECT id FROM woulder.rock_type_groups WHERE name = 'Fast-Drying Rocks')),
        ('Granodiorite', 6.0, 1.2, FALSE, 'Coarse-grained igneous rock similar to granite. Dries quickly.',
         (SELECT id FROM woulder.rock_type_groups WHERE name = 'Fast-Drying Rocks')),
        ('Tonalite', 6.5, 1.5, FALSE, 'Plagioclase-rich igneous rock. Similar to granite, dries quickly.',
         (SELECT id FROM woulder.rock_type_groups WHERE name = 'Fast-Drying Rocks')),
        ('Rhyolite', 8.0, 7.0, FALSE, 'Fine-grained volcanic rock. Glassy texture sheds water well.',
         (SELECT id FROM woulder.rock_type_groups WHERE name = 'Fast-Drying Rocks')),

        -- Medium-drying rocks
        ('Basalt', 10.0, 5.0, FALSE, 'Dense volcanic rock. May have vesicles that trap water.',
         (SELECT id FROM woulder.rock_type_groups WHERE name = 'Medium-Drying Rocks')),
        ('Andesite', 10.0, 6.0, FALSE, 'Intermediate volcanic rock. Moderate drying time.',
         (SELECT id FROM woulder.rock_type_groups WHERE name = 'Medium-Drying Rocks')),
        ('Schist', 12.0, 3.5, FALSE, 'Foliated metamorphic rock. Water can seep between layers.',
         (SELECT id FROM woulder.rock_type_groups WHERE name = 'Medium-Drying Rocks')),

        -- Slow-drying rocks
        ('Phyllite', 20.0, 10.0, FALSE, 'Fine-grained metamorphic rock. Holds moisture in foliation.',
         (SELECT id FROM woulder.rock_type_groups WHERE name = 'Slow-Drying Rocks')),
        ('Argillite', 24.0, 12.0, FALSE, 'Clay-rich sedimentary rock. Absorbs and retains water.',
         (SELECT id FROM woulder.rock_type_groups WHERE name = 'Slow-Drying Rocks')),
        ('Chert', 14.0, 3.0, FALSE, 'Dense sedimentary rock. Micro-pores can hold water.',
         (SELECT id FROM woulder.rock_type_groups WHERE name = 'Slow-Drying Rocks')),
        ('Metavolcanic', 14.0, 4.0, FALSE, 'Metamorphosed volcanic rock. Moderate absorption.',
         (SELECT id FROM woulder.rock_type_groups WHERE name = 'Slow-Drying Rocks'))
    ON CONFLICT (name) DO NOTHING;
END $$;

-- Location Rock Types
DO $$
BEGIN
    -- Skykomish - Money Creek: andesite basalt, phyllite, chert metavolcanic, granodiorite granite
    INSERT INTO woulder.location_rock_types (location_id, rock_type_id, is_primary)
    SELECT l.id, rt.id, CASE WHEN rt.name = 'Andesite' THEN TRUE ELSE FALSE END
    FROM woulder.locations l
    CROSS JOIN woulder.rock_types rt
    WHERE l.name = 'Skykomish - Money Creek' AND rt.name IN ('Andesite', 'Basalt', 'Phyllite', 'Chert', 'Metavolcanic', 'Granodiorite', 'Granite')
    ON CONFLICT DO NOTHING;

    -- Index: Granite
    INSERT INTO woulder.location_rock_types (location_id, rock_type_id, is_primary)
    SELECT l.id, rt.id, TRUE
    FROM woulder.locations l
    CROSS JOIN woulder.rock_types rt
    WHERE l.name = 'Index' AND rt.name = 'Granite'
    ON CONFLICT DO NOTHING;

    -- Gold Bar: Granodiorite Granite
    INSERT INTO woulder.location_rock_types (location_id, rock_type_id, is_primary)
    SELECT l.id, rt.id, CASE WHEN rt.name = 'Granodiorite' THEN TRUE ELSE FALSE END
    FROM woulder.locations l
    CROSS JOIN woulder.rock_types rt
    WHERE l.name = 'Gold Bar' AND rt.name IN ('Granodiorite', 'Granite')
    ON CONFLICT DO NOTHING;

    -- Bellingham: Arkose Sandstone
    INSERT INTO woulder.location_rock_types (location_id, rock_type_id, is_primary)
    SELECT l.id, rt.id, CASE WHEN rt.name = 'Arkose' THEN TRUE ELSE FALSE END
    FROM woulder.locations l
    CROSS JOIN woulder.rock_types rt
    WHERE l.name = 'Bellingham' AND rt.name IN ('Arkose', 'Sandstone')
    ON CONFLICT DO NOTHING;

    -- Icicle Creek (Leavenworth): Granodiorite
    INSERT INTO woulder.location_rock_types (location_id, rock_type_id, is_primary)
    SELECT l.id, rt.id, TRUE
    FROM woulder.locations l
    CROSS JOIN woulder.rock_types rt
    WHERE l.name = 'Icicle Creek (Leavenworth)' AND rt.name = 'Granodiorite'
    ON CONFLICT DO NOTHING;

    -- Squamish: Granodiorite
    INSERT INTO woulder.location_rock_types (location_id, rock_type_id, is_primary)
    SELECT l.id, rt.id, TRUE
    FROM woulder.locations l
    CROSS JOIN woulder.rock_types rt
    WHERE l.name = 'Squamish' AND rt.name = 'Granodiorite'
    ON CONFLICT DO NOTHING;

    -- Paradise: Granodiorite Granite Andesite
    INSERT INTO woulder.location_rock_types (location_id, rock_type_id, is_primary)
    SELECT l.id, rt.id, CASE WHEN rt.name = 'Granodiorite' THEN TRUE ELSE FALSE END
    FROM woulder.locations l
    CROSS JOIN woulder.rock_types rt
    WHERE l.name = 'Skykomish - Paradise' AND rt.name IN ('Granodiorite', 'Granite', 'Andesite')
    ON CONFLICT DO NOTHING;

    -- Treasury: Granodiorite Granite
    INSERT INTO woulder.location_rock_types (location_id, rock_type_id, is_primary)
    SELECT l.id, rt.id, CASE WHEN rt.name = 'Granodiorite' THEN TRUE ELSE FALSE END
    FROM woulder.locations l
    CROSS JOIN woulder.rock_types rt
    WHERE l.name = 'Treasury' AND rt.name IN ('Granodiorite', 'Granite')
    ON CONFLICT DO NOTHING;

    -- Calendar Butte: Granite
    INSERT INTO woulder.location_rock_types (location_id, rock_type_id, is_primary)
    SELECT l.id, rt.id, TRUE
    FROM woulder.locations l
    CROSS JOIN woulder.rock_types rt
    WHERE l.name = 'Calendar Butte' AND rt.name = 'Granite'
    ON CONFLICT DO NOTHING;

    -- Joshua Tree: Granodiorite
    INSERT INTO woulder.location_rock_types (location_id, rock_type_id, is_primary)
    SELECT l.id, rt.id, TRUE
    FROM woulder.locations l
    CROSS JOIN woulder.rock_types rt
    WHERE l.name = 'Joshua Tree' AND rt.name = 'Granodiorite'
    ON CONFLICT DO NOTHING;

    -- Black Mountain: Tonalite
    INSERT INTO woulder.location_rock_types (location_id, rock_type_id, is_primary)
    SELECT l.id, rt.id, TRUE
    FROM woulder.locations l
    CROSS JOIN woulder.rock_types rt
    WHERE l.name = 'Black Mountain' AND rt.name = 'Tonalite'
    ON CONFLICT DO NOTHING;

    -- Buttermilks: Granodiorite
    INSERT INTO woulder.location_rock_types (location_id, rock_type_id, is_primary)
    SELECT l.id, rt.id, TRUE
    FROM woulder.locations l
    CROSS JOIN woulder.rock_types rt
    WHERE l.name = 'Buttermilks' AND rt.name = 'Granodiorite'
    ON CONFLICT DO NOTHING;

    -- Happy / Sad Boulders: Rhyolite
    INSERT INTO woulder.location_rock_types (location_id, rock_type_id, is_primary)
    SELECT l.id, rt.id, TRUE
    FROM woulder.locations l
    CROSS JOIN woulder.rock_types rt
    WHERE l.name = 'Happy / Sad Boulders' AND rt.name = 'Rhyolite'
    ON CONFLICT DO NOTHING;

    -- Yosemite: Granodiorite
    INSERT INTO woulder.location_rock_types (location_id, rock_type_id, is_primary)
    SELECT l.id, rt.id, TRUE
    FROM woulder.locations l
    CROSS JOIN woulder.rock_types rt
    WHERE l.name = 'Yosemite' AND rt.name = 'Granodiorite'
    ON CONFLICT DO NOTHING;

    -- Tramway: Tonalite
    INSERT INTO woulder.location_rock_types (location_id, rock_type_id, is_primary)
    SELECT l.id, rt.id, TRUE
    FROM woulder.locations l
    CROSS JOIN woulder.rock_types rt
    WHERE l.name = 'Tramway' AND rt.name = 'Tonalite'
    ON CONFLICT DO NOTHING;
END $$;

-- ============================================================================
-- VERIFICATION
-- ============================================================================

-- Display setup results
DO $$
DECLARE
    area_count INTEGER;
    location_count INTEGER;
    river_count INTEGER;
    rock_group_count INTEGER;
    rock_type_count INTEGER;
    location_rock_type_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO area_count FROM woulder.areas;
    SELECT COUNT(*) INTO location_count FROM woulder.locations;
    SELECT COUNT(*) INTO river_count FROM woulder.rivers;
    SELECT COUNT(*) INTO rock_group_count FROM woulder.rock_type_groups;
    SELECT COUNT(*) INTO rock_type_count FROM woulder.rock_types;
    SELECT COUNT(*) INTO location_rock_type_count FROM woulder.location_rock_types;

    RAISE NOTICE '';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'Woulder Database Setup Complete!';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'Areas created: %', area_count;
    RAISE NOTICE 'Locations created: %', location_count;
    RAISE NOTICE 'Rivers created: %', river_count;
    RAISE NOTICE 'Rock type groups created: %', rock_group_count;
    RAISE NOTICE 'Rock types created: %', rock_type_count;
    RAISE NOTICE 'Location-rock type associations: %', location_rock_type_count;
    RAISE NOTICE '';
    RAISE NOTICE 'Schema: woulder';
    RAISE NOTICE 'Tables: areas, locations, weather_data, rivers, rock_type_groups, rock_types, location_rock_types';
    RAISE NOTICE '========================================';
END $$;
