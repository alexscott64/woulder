-- Woulder Initial Schema Migration
-- Creates all tables, indexes, triggers, and functions

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
