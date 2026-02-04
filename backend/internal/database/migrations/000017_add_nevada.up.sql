-- Migration: Add Nevada (Red Rock Canyon) area and locations
-- Created: 2026-02-03

-- Add Nevada area
INSERT INTO areas (name, description, region, display_order)
VALUES ('Nevada', 'World-class sandstone bouldering in Red Rock Canyon National Conservation Area. Features stunning Aztec sandstone with crimps, slopers, and technical sequences across hundreds of classic problems. Rock is extremely wet-sensitive - DO NOT CLIMB when wet or damp.', 'Southwest', 3)
ON CONFLICT (name) DO NOTHING;

-- Add Nevada locations
DO $$
DECLARE
    nevada_area_id INTEGER;
BEGIN
    SELECT id INTO nevada_area_id FROM areas WHERE name = 'Nevada';

    -- Nevada locations (Red Rock Canyon) - Top-level areas
    INSERT INTO locations (name, latitude, longitude, elevation_ft, area_id)
    VALUES
        ('Oak Creek Canyon Boulders', 36.11096, -115.46617, 4200, nevada_area_id),
        ('Black Velvet Canyon Boulders', 36.03591, -115.46189, 4500, nevada_area_id),
        ('Windy Canyon Boulders', 36.01527, -115.45304, 4600, nevada_area_id),
        ('Second Pullout Boulders', 36.15197, -115.43752, 3800, nevada_area_id),
        ('Willow Spring Boulders', 36.16092, -115.49868, 4100, nevada_area_id),
        ('First Pullout Boulders', 36.14635, -115.43096, 3700, nevada_area_id),
        ('White Rock Spring Boulders', 36.17320, -115.47768, 4000, nevada_area_id),
        ('Pine Creek Canyon Boulders', 36.12851, -115.47299, 4300, nevada_area_id),
        ('Sandstone Quarry Boulders', 36.16235, -115.45040, 3900, nevada_area_id),
        ('Ice Box Canyon Boulders', 36.15010, -115.48401, 4200, nevada_area_id),
        ('Juniper Canyon Boulders', 36.11650, -115.48466, 4400, nevada_area_id),
        ('Southern Outcrops Boulders', 36.00599, -115.46113, 4700, nevada_area_id),
        ('Mustang Canyon Boulders', 36.04838, -115.46054, 4500, nevada_area_id),
        ('First Creek Canyon Boulders', 36.08122, -115.44808, 4400, nevada_area_id),
        ('Kraft Boulders', 36.15686, -115.42080, 3600, nevada_area_id),
        ('Red Spring Boulders', 36.14651, -115.41900, 3500, nevada_area_id),
        ('Gateway Canyon', 36.16367, -115.41210, 3700, nevada_area_id),
        ('Little Springs Canyon Boulders', 36.15369, -115.42569, 3600, nevada_area_id),
        ('Ash Spring Boulders', 36.15834, -115.42919, 3700, nevada_area_id)
    ON CONFLICT (name) DO NOTHING;
END $$;

-- Add rock types for Nevada locations (all Sandstone - wet-sensitive)
INSERT INTO location_rock_types (location_id, rock_type_id, is_primary)
SELECT l.id, rt.id, TRUE
FROM locations l
CROSS JOIN rock_types rt
WHERE l.name IN (
    'Oak Creek Canyon Boulders',
    'Black Velvet Canyon Boulders',
    'Windy Canyon Boulders',
    'Second Pullout Boulders',
    'Willow Spring Boulders',
    'First Pullout Boulders',
    'White Rock Spring Boulders',
    'Pine Creek Canyon Boulders',
    'Sandstone Quarry Boulders',
    'Ice Box Canyon Boulders',
    'Juniper Canyon Boulders',
    'Southern Outcrops Boulders',
    'Mustang Canyon Boulders',
    'First Creek Canyon Boulders',
    'Kraft Boulders',
    'Red Spring Boulders',
    'Gateway Canyon',
    'Little Springs Canyon Boulders',
    'Ash Spring Boulders'
) AND rt.name = 'Sandstone'
ON CONFLICT DO NOTHING;

-- Add sun exposure profiles for Nevada locations
INSERT INTO location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 40.0, 30.0, 20.0, 10.0, 30.0, 30.0, 5.0, 'Desert canyon sandstone, excellent sun exposure, minimal trees'
FROM locations WHERE name = 'Oak Creek Canyon Boulders'
ON CONFLICT DO NOTHING;

INSERT INTO location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 35.0, 30.0, 20.0, 15.0, 25.0, 35.0, 3.0, 'Desert canyon sandstone, excellent sun exposure, very sparse vegetation'
FROM locations WHERE name = 'Black Velvet Canyon Boulders'
ON CONFLICT DO NOTHING;

INSERT INTO location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 40.0, 35.0, 15.0, 10.0, 30.0, 30.0, 5.0, 'Windy desert canyon, excellent sun exposure, minimal shade'
FROM locations WHERE name = 'Windy Canyon Boulders'
ON CONFLICT DO NOTHING;

INSERT INTO location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 45.0, 30.0, 15.0, 10.0, 35.0, 25.0, 2.0, 'Roadside bouldering, excellent sun exposure, virtually no shade'
FROM locations WHERE name = 'Second Pullout Boulders'
ON CONFLICT DO NOTHING;

INSERT INTO location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 38.0, 32.0, 20.0, 10.0, 28.0, 32.0, 8.0, 'Desert canyon with some pinyon pine, good sun exposure'
FROM locations WHERE name = 'Willow Spring Boulders'
ON CONFLICT DO NOTHING;

INSERT INTO location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 45.0, 30.0, 15.0, 10.0, 35.0, 25.0, 2.0, 'Roadside bouldering, excellent sun exposure, virtually no shade'
FROM locations WHERE name = 'First Pullout Boulders'
ON CONFLICT DO NOTHING;

INSERT INTO location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 40.0, 30.0, 20.0, 10.0, 32.0, 28.0, 6.0, 'Desert canyon sandstone, excellent sun exposure, sparse trees'
FROM locations WHERE name = 'White Rock Spring Boulders'
ON CONFLICT DO NOTHING;

INSERT INTO location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 38.0, 32.0, 20.0, 10.0, 30.0, 30.0, 7.0, 'Desert canyon with some vegetation, good sun exposure'
FROM locations WHERE name = 'Pine Creek Canyon Boulders'
ON CONFLICT DO NOTHING;

INSERT INTO location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 42.0, 30.0, 18.0, 10.0, 33.0, 27.0, 4.0, 'Quarry area sandstone, excellent sun exposure, minimal vegetation'
FROM locations WHERE name = 'Sandstone Quarry Boulders'
ON CONFLICT DO NOTHING;

INSERT INTO location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 35.0, 30.0, 22.0, 13.0, 28.0, 32.0, 10.0, 'Shaded canyon with seasonal water, more vegetation than most Red Rock areas'
FROM locations WHERE name = 'Ice Box Canyon Boulders'
ON CONFLICT DO NOTHING;

INSERT INTO location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 38.0, 30.0, 20.0, 12.0, 30.0, 30.0, 8.0, 'Desert canyon with juniper, good sun exposure'
FROM locations WHERE name = 'Juniper Canyon Boulders'
ON CONFLICT DO NOTHING;

INSERT INTO location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 40.0, 32.0, 18.0, 10.0, 32.0, 28.0, 3.0, 'Desert outcrops, excellent sun exposure, very sparse vegetation'
FROM locations WHERE name = 'Southern Outcrops Boulders'
ON CONFLICT DO NOTHING;

INSERT INTO location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 38.0, 30.0, 20.0, 12.0, 30.0, 30.0, 5.0, 'Desert canyon sandstone, excellent sun exposure, minimal trees'
FROM locations WHERE name = 'Mustang Canyon Boulders'
ON CONFLICT DO NOTHING;

INSERT INTO location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 40.0, 30.0, 20.0, 10.0, 32.0, 28.0, 6.0, 'Desert canyon sandstone, excellent sun exposure, sparse vegetation'
FROM locations WHERE name = 'First Creek Canyon Boulders'
ON CONFLICT DO NOTHING;

-- Calico Basin sub-areas
INSERT INTO location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 42.0, 30.0, 18.0, 10.0, 30.0, 30.0, 4.0, 'Calico Basin sandstone, excellent sun exposure, minimal shade'
FROM locations WHERE name = 'Kraft Boulders'
ON CONFLICT DO NOTHING;

INSERT INTO location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 40.0, 30.0, 20.0, 10.0, 28.0, 32.0, 6.0, 'Calico Basin near spring, good sun exposure, some vegetation'
FROM locations WHERE name = 'Red Spring Boulders'
ON CONFLICT DO NOTHING;

INSERT INTO location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 38.0, 32.0, 20.0, 10.0, 30.0, 30.0, 5.0, 'Calico Basin gateway area, excellent sun exposure, sparse trees'
FROM locations WHERE name = 'Gateway Canyon'
ON CONFLICT DO NOTHING;

INSERT INTO location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 38.0, 30.0, 20.0, 12.0, 28.0, 32.0, 7.0, 'Calico Basin canyon with seasonal spring, some vegetation'
FROM locations WHERE name = 'Little Springs Canyon Boulders'
ON CONFLICT DO NOTHING;

INSERT INTO location_sun_exposure (location_id, south_facing_percent, west_facing_percent, east_facing_percent, north_facing_percent, slab_percent, overhang_percent, tree_coverage_percent, description)
SELECT id, 40.0, 30.0, 20.0, 10.0, 30.0, 30.0, 6.0, 'Calico Basin near spring, good sun exposure, some vegetation'
FROM locations WHERE name = 'Ash Spring Boulders'
ON CONFLICT DO NOTHING;
