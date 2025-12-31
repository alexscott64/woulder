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
COMMENT ON COLUMN woulder.areas.display_order IS 'Order in which areas should be displayed (lower = first)';
COMMENT ON COLUMN woulder.locations.area_id IS 'Foreign key to areas table';
COMMENT ON COLUMN woulder.rivers.is_estimated IS 'Whether flow is estimated from nearby gauge or direct reading';
COMMENT ON COLUMN woulder.rivers.flow_divisor IS 'Divisor to apply to gauge reading (e.g., 2.0 means divide by 2)';
COMMENT ON COLUMN woulder.rivers.drainage_area_sq_mi IS 'Drainage area of the river at crossing point';
COMMENT ON COLUMN woulder.rivers.gauge_drainage_area_sq_mi IS 'Drainage area at the gauge location';

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

-- ============================================================================
-- VERIFICATION
-- ============================================================================

-- Display setup results
DO $$
DECLARE
    area_count INTEGER;
    location_count INTEGER;
    river_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO area_count FROM woulder.areas;
    SELECT COUNT(*) INTO location_count FROM woulder.locations;
    SELECT COUNT(*) INTO river_count FROM woulder.rivers;

    RAISE NOTICE '';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'Woulder Database Setup Complete!';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'Areas created: %', area_count;
    RAISE NOTICE 'Locations created: %', location_count;
    RAISE NOTICE 'Rivers created: %', river_count;
    RAISE NOTICE '';
    RAISE NOTICE 'Schema: woulder';
    RAISE NOTICE 'Tables: areas, locations, weather_data, rivers';
    RAISE NOTICE '========================================';
END $$;
