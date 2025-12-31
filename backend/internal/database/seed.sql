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
SELECT id, '12131500', 'Money Creek', 60, 90, 18.0, 355.0, 1,
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

-- Rock Type Groups
INSERT OR IGNORE INTO rock_type_groups (name, description) VALUES
('Wet-Sensitive Rocks', 'Soft rocks that are permanently damaged when climbed wet. DO NOT CLIMB WHEN WET.'),
('Fast-Drying Rocks', 'Hard, non-porous rocks that dry quickly after rain.'),
('Medium-Drying Rocks', 'Rocks with moderate porosity that take longer to dry.'),
('Slow-Drying Rocks', 'Rocks that absorb and retain water, requiring extended drying time.');

-- Rock Types (with group assignments)
INSERT OR IGNORE INTO rock_types (name, base_drying_hours, porosity_percent, is_wet_sensitive, description, rock_type_group_id) VALUES
-- Wet-sensitive rocks (group_id will be looked up)
('Sandstone', 36.0, 20.0, 1, 'Soft sedimentary rock that absorbs water and becomes friable when wet. DO NOT CLIMB WHEN WET.',
 (SELECT id FROM rock_type_groups WHERE name = 'Wet-Sensitive Rocks')),
('Arkose', 36.0, 18.0, 1, 'Feldspar-rich sandstone. Soft and water-absorbent. DO NOT CLIMB WHEN WET.',
 (SELECT id FROM rock_type_groups WHERE name = 'Wet-Sensitive Rocks')),
('Graywacke', 30.0, 15.0, 1, 'Hard sandstone with clay matrix. Absorbs water. DO NOT CLIMB WHEN WET.',
 (SELECT id FROM rock_type_groups WHERE name = 'Wet-Sensitive Rocks')),

-- Fast-drying rocks
('Granite', 6.0, 1.0, 0, 'Hard crystalline igneous rock. Non-porous, dries quickly.',
 (SELECT id FROM rock_type_groups WHERE name = 'Fast-Drying Rocks')),
('Granodiorite', 6.0, 1.2, 0, 'Coarse-grained igneous rock similar to granite. Dries quickly.',
 (SELECT id FROM rock_type_groups WHERE name = 'Fast-Drying Rocks')),
('Tonalite', 6.5, 1.5, 0, 'Plagioclase-rich igneous rock. Similar to granite, dries quickly.',
 (SELECT id FROM rock_type_groups WHERE name = 'Fast-Drying Rocks')),
('Rhyolite', 8.0, 7.0, 0, 'Fine-grained volcanic rock. Glassy texture sheds water well.',
 (SELECT id FROM rock_type_groups WHERE name = 'Fast-Drying Rocks')),

-- Medium-drying rocks
('Basalt', 10.0, 5.0, 0, 'Dense volcanic rock. May have vesicles that trap water.',
 (SELECT id FROM rock_type_groups WHERE name = 'Medium-Drying Rocks')),
('Andesite', 10.0, 6.0, 0, 'Intermediate volcanic rock. Moderate drying time.',
 (SELECT id FROM rock_type_groups WHERE name = 'Medium-Drying Rocks')),
('Schist', 12.0, 3.5, 0, 'Foliated metamorphic rock. Water can seep between layers.',
 (SELECT id FROM rock_type_groups WHERE name = 'Medium-Drying Rocks')),

-- Slow-drying rocks
('Schist', 12.0, 3.5, 0, 'Foliated metamorphic rock. Water can seep between layers.',
 (SELECT id FROM rock_type_groups WHERE name = 'Slow-Drying Rocks')),
('Phyllite', 20.0, 10.0, 0, 'Fine-grained metamorphic rock. Holds moisture in foliation.',
 (SELECT id FROM rock_type_groups WHERE name = 'Slow-Drying Rocks')),
('Argillite', 24.0, 12.0, 0, 'Clay-rich sedimentary rock. Absorbs and retains water.',
 (SELECT id FROM rock_type_groups WHERE name = 'Slow-Drying Rocks')),
('Chert', 14.0, 3.0, 0, 'Dense sedimentary rock. Micro-pores can hold water.',
 (SELECT id FROM rock_type_groups WHERE name = 'Slow-Drying Rocks')),
('Metavolcanic', 14.0, 4.0, 0, 'Metamorphosed volcanic rock. Moderate absorption.',
 (SELECT id FROM rock_type_groups WHERE name = 'Slow-Drying Rocks'));

-- Sun Exposure Profiles
-- Pacific Northwest locations - moderate tree coverage, mixed aspects
INSERT OR IGNORE INTO location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 35.0, 25.0, 25.0, 15.0, 40.0, 20.0, 60.0, 'Forest bouldering with mixed aspects and moderate tree shade'
FROM locations WHERE name = 'Skykomish - Money Creek';

INSERT OR IGNORE INTO location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 30.0, 30.0, 20.0, 20.0, 35.0, 25.0, 50.0, 'Wall climbing with good west/south exposure, moderate tree coverage'
FROM locations WHERE name = 'Index';

INSERT OR IGNORE INTO location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 40.0, 25.0, 20.0, 15.0, 45.0, 15.0, 55.0, 'Forest bouldering with good south exposure'
FROM locations WHERE name = 'Gold Bar';

INSERT OR IGNORE INTO location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 20.0, 30.0, 25.0, 25.0, 30.0, 30.0, 70.0, 'Dense forest bouldering, mostly sandstone with heavy tree coverage'
FROM locations WHERE name = 'Bellingham';

INSERT OR IGNORE INTO location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 45.0, 25.0, 20.0, 10.0, 50.0, 10.0, 35.0, 'Excellent south exposure, moderate tree coverage, many slabs'
FROM locations WHERE name = 'Icicle Creek (Leavenworth)';

INSERT OR IGNORE INTO location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 40.0, 30.0, 20.0, 10.0, 40.0, 20.0, 45.0, 'Mixed wall and boulder climbing with good sun exposure'
FROM locations WHERE name = 'Squamish';

INSERT OR IGNORE INTO location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 35.0, 25.0, 25.0, 15.0, 45.0, 15.0, 65.0, 'High elevation forest bouldering, good slab percentage but heavy tree shade'
FROM locations WHERE name = 'Skykomish - Paradise';

INSERT OR IGNORE INTO location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 30.0, 30.0, 25.0, 15.0, 40.0, 20.0, 60.0, 'High alpine bouldering with moderate tree coverage and snowmelt seepage risk'
FROM locations WHERE name = 'Treasury';

INSERT OR IGNORE INTO location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 45.0, 25.0, 20.0, 10.0, 55.0, 10.0, 40.0, 'Excellent south/slab exposure with moderate tree coverage'
FROM locations WHERE name = 'Calendar Butte';

-- Southern California locations - minimal tree coverage, excellent sun exposure
INSERT OR IGNORE INTO location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 40.0, 30.0, 20.0, 10.0, 35.0, 25.0, 5.0, 'Desert granite bouldering, excellent sun exposure, minimal shade'
FROM locations WHERE name = 'Joshua Tree';

INSERT OR IGNORE INTO location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 50.0, 25.0, 15.0, 10.0, 50.0, 10.0, 20.0, 'High alpine bouldering, excellent south/slab exposure, minimal trees'
FROM locations WHERE name = 'Black Mountain';

INSERT OR IGNORE INTO location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 45.0, 30.0, 15.0, 10.0, 45.0, 15.0, 10.0, 'High desert bouldering, excellent sun exposure, sparse tree coverage'
FROM locations WHERE name = 'Buttermilks';

INSERT OR IGNORE INTO location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 40.0, 30.0, 20.0, 10.0, 40.0, 20.0, 15.0, 'High desert bouldering, good sun exposure, minimal tree coverage'
FROM locations WHERE name = 'Happy / Sad Boulders';

INSERT OR IGNORE INTO location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 35.0, 30.0, 20.0, 15.0, 40.0, 20.0, 30.0, 'Mixed alpine climbing, moderate tree coverage at lower elevations'
FROM locations WHERE name = 'Yosemite';

INSERT OR IGNORE INTO location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 45.0, 30.0, 15.0, 10.0, 50.0, 10.0, 25.0, 'High alpine bouldering, excellent south exposure, sparse trees'
FROM locations WHERE name = 'Tramway';

-- Update seepage risk flags
UPDATE locations SET has_seepage_risk = 1 WHERE name IN ('Treasury', 'Skykomish - Paradise', 'Bellingham');

-- Location Rock Types
-- Skykomish - Money Creek: andesite basalt, phyllite, chert metavolcanic, granodiorite granite
INSERT OR IGNORE INTO location_rock_types (location_id, rock_type_id, is_primary)
SELECT l.id, rt.id, CASE WHEN rt.name = 'Andesite' THEN 1 ELSE 0 END
FROM locations l
CROSS JOIN rock_types rt
WHERE l.name = 'Skykomish - Money Creek' AND rt.name IN ('Andesite', 'Basalt', 'Phyllite', 'Chert', 'Metavolcanic', 'Granodiorite', 'Granite');

-- Index: Granite
INSERT OR IGNORE INTO location_rock_types (location_id, rock_type_id, is_primary)
SELECT l.id, rt.id, 1
FROM locations l
CROSS JOIN rock_types rt
WHERE l.name = 'Index' AND rt.name = 'Granite';

-- Gold Bar: Granodiorite Granite
INSERT OR IGNORE INTO location_rock_types (location_id, rock_type_id, is_primary)
SELECT l.id, rt.id, CASE WHEN rt.name = 'Granodiorite' THEN 1 ELSE 0 END
FROM locations l
CROSS JOIN rock_types rt
WHERE l.name = 'Gold Bar' AND rt.name IN ('Granodiorite', 'Granite');

-- Bellingham: Arkose Sandstone
INSERT OR IGNORE INTO location_rock_types (location_id, rock_type_id, is_primary)
SELECT l.id, rt.id, CASE WHEN rt.name = 'Arkose' THEN 1 ELSE 0 END
FROM locations l
CROSS JOIN rock_types rt
WHERE l.name = 'Bellingham' AND rt.name IN ('Arkose', 'Sandstone');

-- Icicle Creek (Leavenworth): Granodiorite
INSERT OR IGNORE INTO location_rock_types (location_id, rock_type_id, is_primary)
SELECT l.id, rt.id, 1
FROM locations l
CROSS JOIN rock_types rt
WHERE l.name = 'Icicle Creek (Leavenworth)' AND rt.name = 'Granodiorite';

-- Squamish: Granodiorite
INSERT OR IGNORE INTO location_rock_types (location_id, rock_type_id, is_primary)
SELECT l.id, rt.id, 1
FROM locations l
CROSS JOIN rock_types rt
WHERE l.name = 'Squamish' AND rt.name = 'Granodiorite';

-- Paradise: Granodiorite Granite Andesite
INSERT OR IGNORE INTO location_rock_types (location_id, rock_type_id, is_primary)
SELECT l.id, rt.id, CASE WHEN rt.name = 'Granodiorite' THEN 1 ELSE 0 END
FROM locations l
CROSS JOIN rock_types rt
WHERE l.name = 'Skykomish - Paradise' AND rt.name IN ('Granodiorite', 'Granite', 'Andesite');

-- Treasury: Granodiorite Granite
INSERT OR IGNORE INTO location_rock_types (location_id, rock_type_id, is_primary)
SELECT l.id, rt.id, CASE WHEN rt.name = 'Granodiorite' THEN 1 ELSE 0 END
FROM locations l
CROSS JOIN rock_types rt
WHERE l.name = 'Treasury' AND rt.name IN ('Granodiorite', 'Granite');

-- Calendar Butte: Granite
INSERT OR IGNORE INTO location_rock_types (location_id, rock_type_id, is_primary)
SELECT l.id, rt.id, 1
FROM locations l
CROSS JOIN rock_types rt
WHERE l.name = 'Calendar Butte' AND rt.name = 'Granite';

-- Joshua Tree: Granodiorite
INSERT OR IGNORE INTO location_rock_types (location_id, rock_type_id, is_primary)
SELECT l.id, rt.id, 1
FROM locations l
CROSS JOIN rock_types rt
WHERE l.name = 'Joshua Tree' AND rt.name = 'Granodiorite';

-- Black Mountain: Tonalite
INSERT OR IGNORE INTO location_rock_types (location_id, rock_type_id, is_primary)
SELECT l.id, rt.id, 1
FROM locations l
CROSS JOIN rock_types rt
WHERE l.name = 'Black Mountain' AND rt.name = 'Tonalite';

-- Buttermilks: Granodiorite
INSERT OR IGNORE INTO location_rock_types (location_id, rock_type_id, is_primary)
SELECT l.id, rt.id, 1
FROM locations l
CROSS JOIN rock_types rt
WHERE l.name = 'Buttermilks' AND rt.name = 'Granodiorite';

-- Happy / Sad Boulders: Rhyolite
INSERT OR IGNORE INTO location_rock_types (location_id, rock_type_id, is_primary)
SELECT l.id, rt.id, 1
FROM locations l
CROSS JOIN rock_types rt
WHERE l.name = 'Happy / Sad Boulders' AND rt.name = 'Rhyolite';

-- Yosemite: Granodiorite
INSERT OR IGNORE INTO location_rock_types (location_id, rock_type_id, is_primary)
SELECT l.id, rt.id, 1
FROM locations l
CROSS JOIN rock_types rt
WHERE l.name = 'Yosemite' AND rt.name = 'Granodiorite';

-- Tramway: Tonalite
INSERT OR IGNORE INTO location_rock_types (location_id, rock_type_id, is_primary)
SELECT l.id, rt.id, 1
FROM locations l
CROSS JOIN rock_types rt
WHERE l.name = 'Tramway' AND rt.name = 'Tonalite';
