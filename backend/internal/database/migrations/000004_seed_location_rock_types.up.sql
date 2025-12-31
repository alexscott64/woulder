-- Seed rock types for each location
-- Format: INSERT INTO woulder.location_rock_types (location_id, rock_type_id, is_primary)
--         SELECT id, (SELECT id FROM woulder.rock_types WHERE name = 'RockType'), true/false FROM woulder.locations WHERE name = 'LocationName';

-- Skykomish - Money Creek: andesite basalt, phyllite, chert metavolcanic, granodiorite granite
INSERT INTO woulder.location_rock_types (location_id, rock_type_id, is_primary)
SELECT l.id, rt.id, rt.name = 'Andesite'
FROM woulder.locations l
CROSS JOIN woulder.rock_types rt
WHERE l.name = 'Skykomish - Money Creek' AND rt.name IN ('Andesite', 'Basalt', 'Phyllite', 'Chert', 'Metavolcanic', 'Granodiorite', 'Granite');

-- Index: Granite
INSERT INTO woulder.location_rock_types (location_id, rock_type_id, is_primary)
SELECT l.id, rt.id, TRUE
FROM woulder.locations l
CROSS JOIN woulder.rock_types rt
WHERE l.name = 'Index' AND rt.name = 'Granite';

-- Gold Bar: Granodiorite Granite
INSERT INTO woulder.location_rock_types (location_id, rock_type_id, is_primary)
SELECT l.id, rt.id, rt.name = 'Granodiorite'
FROM woulder.locations l
CROSS JOIN woulder.rock_types rt
WHERE l.name = 'Gold Bar' AND rt.name IN ('Granodiorite', 'Granite');

-- Bellingham: Arkose Sandstone
INSERT INTO woulder.location_rock_types (location_id, rock_type_id, is_primary)
SELECT l.id, rt.id, rt.name = 'Arkose'
FROM woulder.locations l
CROSS JOIN woulder.rock_types rt
WHERE l.name = 'Bellingham' AND rt.name IN ('Arkose', 'Sandstone');

-- Icicle Creek (Leavenworth): Schist / Granite
INSERT INTO woulder.location_rock_types (location_id, rock_type_id, is_primary)
SELECT l.id, rt.id, rt.name = 'Granite'
FROM woulder.locations l
CROSS JOIN woulder.rock_types rt
WHERE l.name = 'Icicle Creek (Leavenworth)' AND rt.name IN ('Schist', 'Granite');

-- Squamish: Granite
INSERT INTO woulder.location_rock_types (location_id, rock_type_id, is_primary)
SELECT l.id, rt.id, TRUE
FROM woulder.locations l
CROSS JOIN woulder.rock_types rt
WHERE l.name = 'Squamish' AND rt.name = 'Granite';

-- Skykomish - Paradise: Granodiorite Granite
INSERT INTO woulder.location_rock_types (location_id, rock_type_id, is_primary)
SELECT l.id, rt.id, rt.name = 'Granodiorite'
FROM woulder.locations l
CROSS JOIN woulder.rock_types rt
WHERE l.name = 'Skykomish - Paradise' AND rt.name IN ('Granodiorite', 'Granite');

-- Treasury: Granite
INSERT INTO woulder.location_rock_types (location_id, rock_type_id, is_primary)
SELECT l.id, rt.id, TRUE
FROM woulder.locations l
CROSS JOIN woulder.rock_types rt
WHERE l.name = 'Treasury' AND rt.name = 'Granite';

-- Calendar Butte: Graywacke Argillite Phyllite
INSERT INTO woulder.location_rock_types (location_id, rock_type_id, is_primary)
SELECT l.id, rt.id, rt.name = 'Graywacke'
FROM woulder.locations l
CROSS JOIN woulder.rock_types rt
WHERE l.name = 'Calendar Butte' AND rt.name IN ('Graywacke', 'Argillite', 'Phyllite');

-- Joshua Tree: Granodiorite
INSERT INTO woulder.location_rock_types (location_id, rock_type_id, is_primary)
SELECT l.id, rt.id, TRUE
FROM woulder.locations l
CROSS JOIN woulder.rock_types rt
WHERE l.name = 'Joshua Tree' AND rt.name = 'Granodiorite';

-- Black Mountain: Tonalite
INSERT INTO woulder.location_rock_types (location_id, rock_type_id, is_primary)
SELECT l.id, rt.id, TRUE
FROM woulder.locations l
CROSS JOIN woulder.rock_types rt
WHERE l.name = 'Black Mountain' AND rt.name = 'Tonalite';

-- Buttermilks: Granodiorite
INSERT INTO woulder.location_rock_types (location_id, rock_type_id, is_primary)
SELECT l.id, rt.id, TRUE
FROM woulder.locations l
CROSS JOIN woulder.rock_types rt
WHERE l.name = 'Buttermilks' AND rt.name = 'Granodiorite';

-- Happy / Sad Boulders: Rhyolite
INSERT INTO woulder.location_rock_types (location_id, rock_type_id, is_primary)
SELECT l.id, rt.id, TRUE
FROM woulder.locations l
CROSS JOIN woulder.rock_types rt
WHERE l.name = 'Happy / Sad Boulders' AND rt.name = 'Rhyolite';

-- Yosemite: Granodiorite
INSERT INTO woulder.location_rock_types (location_id, rock_type_id, is_primary)
SELECT l.id, rt.id, TRUE
FROM woulder.locations l
CROSS JOIN woulder.rock_types rt
WHERE l.name = 'Yosemite' AND rt.name = 'Granodiorite';

-- Tramway: Tonalite
INSERT INTO woulder.location_rock_types (location_id, rock_type_id, is_primary)
SELECT l.id, rt.id, TRUE
FROM woulder.locations l
CROSS JOIN woulder.rock_types rt
WHERE l.name = 'Tramway' AND rt.name = 'Tonalite';
