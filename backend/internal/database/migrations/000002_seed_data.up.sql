-- Woulder Seed Data Migration
-- Inserts initial areas, locations, and river data

-- Set search path
SET search_path TO woulder, public;

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
    'Money Creek (estimated)',
    60,
    90,
    18.0,
    355.0,
    TRUE,
    'Flow estimated from South Fork Skykomish at Skykomish. Watershed includes Lake Elizabeth, Goat Creek, Crosby Mountain, Apex Mine drainage'
FROM woulder.locations l
WHERE l.name = 'Skykomish - Money Creek'
AND NOT EXISTS (
    SELECT 1 FROM woulder.rivers r
    WHERE r.location_id = l.id AND r.gauge_id = '12131500' AND r.river_name = 'Money Creek (estimated)'
);

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
AND NOT EXISTS (
    SELECT 1 FROM woulder.rivers r
    WHERE r.location_id = l.id AND r.gauge_id = '12134500' AND r.river_name = 'North Fork Skykomish River'
);

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
AND NOT EXISTS (
    SELECT 1 FROM woulder.rivers r
    WHERE r.location_id = l.id AND r.gauge_id = '12134500' AND r.river_name = 'Skykomish River'
);

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
AND NOT EXISTS (
    SELECT 1 FROM woulder.rivers r
    WHERE r.location_id = l.id AND r.gauge_id = '12134500' AND r.river_name = 'West Fork Miller River'
);

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
AND NOT EXISTS (
    SELECT 1 FROM woulder.rivers r
    WHERE r.location_id = l.id AND r.gauge_id = '12134500' AND r.river_name = 'East Fork Miller River'
);
