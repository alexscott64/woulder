-- Woulder Seed Data
-- This file is tracked in git and contains initial location and river data

-- Areas
INSERT OR IGNORE INTO areas (name, description, region, display_order) VALUES
    ('Pacific Northwest', 'Climbing areas in Washington, Oregon, and British Columbia', 'Northwest', 1),
    ('Southern California', 'Desert climbing and high-alpine bouldering in Southern California', 'Southwest', 2);

-- Locations
-- Pacific Northwest locations (area_id = 1)
INSERT OR IGNORE INTO locations (name, latitude, longitude, elevation_ft, area_id) VALUES
    ('Skykomish - Money Creek', 47.69727769, -121.47884640, 1000, 1),
    ('Index', 47.82061272, -121.55492795, 500, 1),
    ('Gold Bar', 47.8468, -121.6970, 200, 1),
    ('Bellingham', 48.7519, -122.4787, 100, 1),
    ('Icicle Creek (Leavenworth)', 47.5962, -120.6615, 1200, 1),
    ('Squamish', 49.7016, -123.1558, 200, 1),
    ('Skykomish - Paradise', 47.64074805, -121.37822668, 1500, 1),
    ('Treasury', 47.76086166, -121.12877297, 3650, 1),
    ('Calendar Butte', 48.36202, -122.08273, 1600, 1);

-- Southern California locations (area_id = 2)
INSERT OR IGNORE INTO locations (name, latitude, longitude, elevation_ft, area_id) VALUES
    ('Joshua Tree', 34.01565, -116.16298, 2700, 2),
    ('Black Mountain', 33.82629, -116.7591, 7500, 2),
    ('Buttermilks', 37.3276, -118.5757, 6400, 2),
    ('Happy / Sad Boulders', 37.41601, -118.43994, 4400, 2),
    ('Yosemite', 37.7416, -119.60152, 3977, 2),
    ('Tramway', 33.81074, -116.65175, 8519, 2);

-- Rivers (using active USGS gauges)
-- Money Creek - estimated from South Fork Skykomish at Skykomish (12131500)
INSERT OR IGNORE INTO rivers (location_id, gauge_id, river_name, safe_crossing_cfs, caution_crossing_cfs, drainage_area_sq_mi, gauge_drainage_area_sq_mi, is_estimated, description)
SELECT id, '12131500', 'Money Creek (estimated)', 60, 90, 18.0, 355.0, 1,
       'Flow estimated from South Fork Skykomish at Skykomish. Watershed includes Lake Elizabeth, Goat Creek, Crosby Mountain, Apex Mine drainage'
FROM locations WHERE name = 'Skykomish - Money Creek';

-- Index - North Fork Skykomish estimated from Gold Bar gauge (12134500)
-- Per local bouldering guide: divide gauge reading by 2 for North Fork flow
INSERT OR IGNORE INTO rivers (location_id, gauge_id, river_name, safe_crossing_cfs, caution_crossing_cfs, flow_divisor, is_estimated, description)
SELECT id, '12134500', 'North Fork Skykomish River', 800, 900, 2.0, 1,
       'River crossing to Sasquatch Boulders. Flow estimated as gauge reading / 2. Below 800 CFS is safe, 800-900 CFS use caution, above 900 CFS is dangerous.'
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
