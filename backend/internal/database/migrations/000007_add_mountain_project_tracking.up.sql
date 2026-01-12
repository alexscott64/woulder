-- Migration 000007: Add Mountain Project climb tracking tables
-- Stores areas, routes (boulders), and tick data from Mountain Project API

-- Table 1: Mountain Project Areas
-- Stores hierarchical area data from Mountain Project
CREATE TABLE woulder.mp_areas (
    id SERIAL PRIMARY KEY,
    mp_area_id VARCHAR(50) UNIQUE NOT NULL,           -- Mountain Project area ID
    name VARCHAR(255) NOT NULL,
    parent_mp_area_id VARCHAR(50),                    -- For hierarchy (null = root)
    area_type VARCHAR(50),                            -- "area" or terminal type
    location_id INTEGER REFERENCES woulder.locations(id) ON DELETE SET NULL,
    last_synced_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for mp_areas
CREATE INDEX idx_mp_areas_mp_area_id ON woulder.mp_areas(mp_area_id);
CREATE INDEX idx_mp_areas_location_id ON woulder.mp_areas(location_id);
CREATE INDEX idx_mp_areas_parent ON woulder.mp_areas(parent_mp_area_id);

-- Table 2: Mountain Project Routes (Boulders)
-- Stores individual boulder problems from Mountain Project
CREATE TABLE woulder.mp_routes (
    id SERIAL PRIMARY KEY,
    mp_route_id VARCHAR(50) UNIQUE NOT NULL,          -- Mountain Project route ID
    mp_area_id VARCHAR(50) REFERENCES woulder.mp_areas(mp_area_id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    route_type VARCHAR(100),                          -- "Boulder", "Trad", etc.
    rating VARCHAR(50),                               -- "V4", "5.10a", etc.
    location_id INTEGER REFERENCES woulder.locations(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for mp_routes
CREATE INDEX idx_mp_routes_mp_route_id ON woulder.mp_routes(mp_route_id);
CREATE INDEX idx_mp_routes_location_id ON woulder.mp_routes(location_id);
CREATE INDEX idx_mp_routes_area ON woulder.mp_routes(mp_area_id);

-- Table 3: Mountain Project Ticks (Climb Activity)
-- Stores individual climb logs (ticks) from Mountain Project users
CREATE TABLE woulder.mp_ticks (
    id SERIAL PRIMARY KEY,
    mp_route_id VARCHAR(50) REFERENCES woulder.mp_routes(mp_route_id) ON DELETE CASCADE,
    user_name VARCHAR(255),
    climbed_at TIMESTAMPTZ NOT NULL,
    style VARCHAR(50),                                -- "Lead", "Flash", "Send", etc.
    comment TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for mp_ticks
CREATE INDEX idx_mp_ticks_route ON woulder.mp_ticks(mp_route_id);
CREATE INDEX idx_mp_ticks_climbed_at ON woulder.mp_ticks(climbed_at DESC);
-- Unique constraint to prevent duplicate ticks (same route, user, timestamp)
CREATE UNIQUE INDEX idx_mp_ticks_unique ON woulder.mp_ticks(mp_route_id, user_name, climbed_at);

-- Add auto-update triggers for updated_at columns
CREATE TRIGGER update_mp_areas_updated_at BEFORE UPDATE ON woulder.mp_areas
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_mp_routes_updated_at BEFORE UPDATE ON woulder.mp_routes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_mp_ticks_updated_at BEFORE UPDATE ON woulder.mp_ticks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
