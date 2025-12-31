-- Add sun exposure profiles for locations
-- This enables accurate rock drying calculations based on sun exposure geometry

-- Create location_sun_exposure table
CREATE TABLE IF NOT EXISTS woulder.location_sun_exposure (
    id SERIAL PRIMARY KEY,
    location_id INTEGER NOT NULL UNIQUE REFERENCES woulder.locations(id) ON DELETE CASCADE,

    -- Sun aspect percentages (should sum to ~100%)
    south_facing_percent DECIMAL(5,2) DEFAULT 0 CHECK (south_facing_percent >= 0 AND south_facing_percent <= 100),
    west_facing_percent DECIMAL(5,2) DEFAULT 0 CHECK (west_facing_percent >= 0 AND west_facing_percent <= 100),
    east_facing_percent DECIMAL(5,2) DEFAULT 0 CHECK (east_facing_percent >= 0 AND east_facing_percent <= 100),
    north_facing_percent DECIMAL(5,2) DEFAULT 0 CHECK (north_facing_percent >= 0 AND north_facing_percent <= 100),

    -- Rock angle percentages (affects water runoff and sun exposure)
    slab_percent DECIMAL(5,2) DEFAULT 0 CHECK (slab_percent >= 0 AND slab_percent <= 100),
    overhang_percent DECIMAL(5,2) DEFAULT 0 CHECK (overhang_percent >= 0 AND overhang_percent <= 100),

    -- Tree coverage (affects shade and reduced drying)
    tree_coverage_percent DECIMAL(5,2) DEFAULT 0 CHECK (tree_coverage_percent >= 0 AND tree_coverage_percent <= 100),

    -- Optional description for manual overrides or notes
    description TEXT,

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Add seepage risk flag to locations table
ALTER TABLE woulder.locations ADD COLUMN has_seepage_risk BOOLEAN DEFAULT FALSE;

-- Create indexes for performance
CREATE INDEX idx_location_sun_exposure_location ON woulder.location_sun_exposure(location_id);

-- Add comments for documentation
COMMENT ON TABLE woulder.location_sun_exposure IS 'Sun exposure profiles for each climbing location. Used to calculate sun exposure factor in rock drying algorithm.';
COMMENT ON COLUMN woulder.location_sun_exposure.south_facing_percent IS 'Percentage of climbing area facing south (optimal sun exposure)';
COMMENT ON COLUMN woulder.location_sun_exposure.west_facing_percent IS 'Percentage of climbing area facing west (afternoon sun)';
COMMENT ON COLUMN woulder.location_sun_exposure.east_facing_percent IS 'Percentage of climbing area facing east (morning sun)';
COMMENT ON COLUMN woulder.location_sun_exposure.north_facing_percent IS 'Percentage of climbing area facing north (minimal sun exposure)';
COMMENT ON COLUMN woulder.location_sun_exposure.slab_percent IS 'Percentage of climbing on slabs (water runs off faster, more sun exposure)';
COMMENT ON COLUMN woulder.location_sun_exposure.overhang_percent IS 'Percentage of climbing on overhangs (stays dry longer, less sun exposure)';
COMMENT ON COLUMN woulder.location_sun_exposure.tree_coverage_percent IS 'Percentage of climbing area shaded by trees (reduces drying speed)';
COMMENT ON COLUMN woulder.locations.has_seepage_risk IS 'Location has seepage, snowmelt, or water table issues that slow rock drying';

-- Insert sun exposure profiles for all locations
-- Pacific Northwest locations - moderate tree coverage, mixed aspects
INSERT INTO woulder.location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 35.0, 25.0, 25.0, 15.0, 40.0, 20.0, 60.0, 'Forest bouldering with mixed aspects and moderate tree shade'
FROM woulder.locations WHERE name = 'Skykomish - Money Creek';

INSERT INTO woulder.location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 30.0, 30.0, 20.0, 20.0, 35.0, 25.0, 50.0, 'Wall climbing with good west/south exposure, moderate tree coverage'
FROM woulder.locations WHERE name = 'Index';

INSERT INTO woulder.location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 40.0, 25.0, 20.0, 15.0, 45.0, 15.0, 55.0, 'Forest bouldering with good south exposure'
FROM woulder.locations WHERE name = 'Gold Bar';

INSERT INTO woulder.location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 20.0, 30.0, 25.0, 25.0, 30.0, 30.0, 70.0, 'Dense forest bouldering, mostly sandstone with heavy tree coverage'
FROM woulder.locations WHERE name = 'Bellingham';

INSERT INTO woulder.location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 45.0, 25.0, 20.0, 10.0, 50.0, 10.0, 35.0, 'Excellent south exposure, moderate tree coverage, many slabs'
FROM woulder.locations WHERE name = 'Icicle Creek (Leavenworth)';

INSERT INTO woulder.location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 40.0, 30.0, 20.0, 10.0, 40.0, 20.0, 45.0, 'Mixed wall and boulder climbing with good sun exposure'
FROM woulder.locations WHERE name = 'Squamish';

INSERT INTO woulder.location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 35.0, 25.0, 25.0, 15.0, 45.0, 15.0, 65.0, 'High elevation forest bouldering, good slab percentage but heavy tree shade'
FROM woulder.locations WHERE name = 'Skykomish - Paradise';

INSERT INTO woulder.location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 30.0, 30.0, 25.0, 15.0, 40.0, 20.0, 60.0, 'High alpine bouldering with moderate tree coverage and snowmelt seepage risk'
FROM woulder.locations WHERE name = 'Treasury';

INSERT INTO woulder.location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 45.0, 25.0, 20.0, 10.0, 55.0, 10.0, 40.0, 'Excellent south/slab exposure with moderate tree coverage'
FROM woulder.locations WHERE name = 'Calendar Butte';

-- Southern California locations - minimal tree coverage, excellent sun exposure
INSERT INTO woulder.location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 40.0, 30.0, 20.0, 10.0, 35.0, 25.0, 5.0, 'Desert granite bouldering, excellent sun exposure, minimal shade'
FROM woulder.locations WHERE name = 'Joshua Tree';

INSERT INTO woulder.location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 50.0, 25.0, 15.0, 10.0, 50.0, 10.0, 20.0, 'High alpine bouldering, excellent south/slab exposure, minimal trees'
FROM woulder.locations WHERE name = 'Black Mountain';

INSERT INTO woulder.location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 45.0, 30.0, 15.0, 10.0, 45.0, 15.0, 10.0, 'High desert bouldering, excellent sun exposure, sparse tree coverage'
FROM woulder.locations WHERE name = 'Buttermilks';

INSERT INTO woulder.location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 40.0, 30.0, 20.0, 10.0, 40.0, 20.0, 15.0, 'High desert bouldering, good sun exposure, minimal tree coverage'
FROM woulder.locations WHERE name = 'Happy / Sad Boulders';

INSERT INTO woulder.location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 35.0, 30.0, 20.0, 15.0, 40.0, 20.0, 30.0, 'Mixed alpine climbing, moderate tree coverage at lower elevations'
FROM woulder.locations WHERE name = 'Yosemite';

INSERT INTO woulder.location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 45.0, 30.0, 15.0, 10.0, 50.0, 10.0, 25.0, 'High alpine bouldering, excellent south exposure, sparse trees'
FROM woulder.locations WHERE name = 'Tramway';

-- Update seepage risk flags for locations with groundwater/snowmelt issues
UPDATE woulder.locations SET has_seepage_risk = TRUE WHERE name IN ('Treasury', 'Skykomish - Paradise', 'Bellingham', 'Icicle Creek (Leavenworth)');
