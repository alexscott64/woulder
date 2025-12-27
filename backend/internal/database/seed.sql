-- Woulder Seed Data
-- This file is tracked in git and contains initial location and river data

-- Locations
INSERT OR IGNORE INTO locations (name, latitude, longitude, elevation_ft) VALUES
    ('Skykomish - Money Creek', 47.69727769, -121.47884640, 1000),
    ('Index', 47.82061272, -121.55492795, 500),
    ('Gold Bar', 47.8468, -121.6970, 200),
    ('Bellingham', 48.7519, -122.4787, 100),
    ('Icicle Creek (Leavenworth)', 47.5962, -120.6615, 1200),
    ('Squamish', 49.7016, -123.1558, 200),
    ('Skykomish - Paradise', 47.64074805, -121.37822668, 1500);

-- Rivers (using active USGS gauges)
-- Money Creek - estimated from South Fork Skykomish at Skykomish (12131500)
INSERT OR IGNORE INTO rivers (location_id, gauge_id, river_name, safe_crossing_cfs, caution_crossing_cfs, drainage_area_sq_mi, gauge_drainage_area_sq_mi, is_estimated, description)
SELECT id, '12131500', 'Money Creek (estimated)', 60, 90, 18.0, 355.0, 1,
       'Flow estimated from South Fork Skykomish at Skykomish. Watershed includes Lake Elizabeth, Goat Creek, Crosby Mountain, Apex Mine drainage'
FROM locations WHERE name = 'Skykomish - Money Creek';

-- Index - North Fork Skykomish estimated from Gold Bar gauge (12134500)
INSERT OR IGNORE INTO rivers (location_id, gauge_id, river_name, safe_crossing_cfs, caution_crossing_cfs, drainage_area_sq_mi, gauge_drainage_area_sq_mi, is_estimated, description)
SELECT id, '12134500', 'North Fork Skykomish River', 1500, 2000, 357.0, 535.0, 1,
       'Flow estimated from Skykomish River at Gold Bar. North Fork drainage area.'
FROM locations WHERE name = 'Index';

-- Gold Bar - direct gauge reading (12134500)
INSERT OR IGNORE INTO rivers (location_id, gauge_id, river_name, safe_crossing_cfs, caution_crossing_cfs, is_estimated, description)
SELECT id, '12134500', 'Skykomish River', 3000, 4500, 0,
       'Direct gauge reading from USGS Skykomish River near Gold Bar'
FROM locations WHERE name = 'Gold Bar';

-- Paradise - West Fork Miller River estimated from Gold Bar (12134500)
INSERT OR IGNORE INTO rivers (location_id, gauge_id, river_name, safe_crossing_cfs, caution_crossing_cfs, drainage_area_sq_mi, gauge_drainage_area_sq_mi, is_estimated, description)
SELECT id, '12134500', 'West Fork Miller River', 600, 900, 28.0, 535.0, 1,
       'Flow estimated from Skykomish River at Gold Bar. West Fork Miller River drainage area.'
FROM locations WHERE name = 'Skykomish - Paradise';

-- Paradise - East Fork Miller River estimated from Gold Bar (12134500)
INSERT OR IGNORE INTO rivers (location_id, gauge_id, river_name, safe_crossing_cfs, caution_crossing_cfs, drainage_area_sq_mi, gauge_drainage_area_sq_mi, is_estimated, description)
SELECT id, '12134500', 'East Fork Miller River', 700, 1000, 22.0, 535.0, 1,
       'Flow estimated from Skykomish River at Gold Bar. East Fork Miller River drainage area.'
FROM locations WHERE name = 'Skykomish - Paradise';
