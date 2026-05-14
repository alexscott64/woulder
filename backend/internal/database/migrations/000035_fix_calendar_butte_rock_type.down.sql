-- 000035_fix_calendar_butte_rock_type.down.sql
--
-- Revert: restore Calendar Butte's original Graywacke / Argillite / Phyllite
-- rock-type associations from migration 000004 and remove the Granite
-- association that the up-migration inserted.

DELETE FROM woulder.location_rock_types
WHERE location_id = (SELECT id FROM woulder.locations WHERE name = 'Calendar Butte')
  AND rock_type_id = (SELECT id FROM woulder.rock_types WHERE name = 'Granite');

INSERT INTO woulder.location_rock_types (location_id, rock_type_id, is_primary)
SELECT l.id, rt.id, rt.name = 'Graywacke'
FROM woulder.locations l
CROSS JOIN woulder.rock_types rt
WHERE l.name = 'Calendar Butte' AND rt.name IN ('Graywacke', 'Argillite', 'Phyllite')
ON CONFLICT (location_id, rock_type_id) DO NOTHING;
