-- 000035_fix_calendar_butte_rock_type.up.sql
--
-- Remove the "do not climb when wet" restriction from Calendar Butte.
--
-- Root cause: migration 000004 seeded Calendar Butte (location_id=9) with three
-- rock types — Graywacke, Argillite, and Phyllite. Per migration 000003,
-- `Graywacke` is flagged `is_wet_sensitive = TRUE` ("Hard sandstone with clay
-- matrix. Absorbs water. DO NOT CLIMB WHEN WET."). The frontend
-- (ConditionsModal.tsx, WeatherCard.tsx, ForecastView.tsx) drives the
-- "do_not_climb" / "Wet-Sensitive Rock Warning" condition off
-- `rockStatus.is_wet_sensitive`, which is true for the location whenever ANY
-- linked rock_type has `is_wet_sensitive = TRUE`. The Graywacke association is
-- therefore the source of the wet-rain restriction.
--
-- Calendar Butte is NOT sandstone. Migration 000033 already reclassified
-- Graywacke into the "Granite" rock_type_group for thermal-modeling reasons,
-- but did not flip its `is_wet_sensitive` flag (correctly — there are still
-- legitimate Graywacke locations elsewhere). To clear the restriction for this
-- specific location we instead align the location-to-rock-type associations
-- with the canonical state declared in setup_postgres.sql lines 670-676, which
-- assigns Calendar Butte to Granite only.
--
-- Effect: removes Graywacke / Argillite / Phyllite associations for Calendar
-- Butte and inserts Granite as the single primary rock type. None of those
-- removed types were marked `is_wet_sensitive = TRUE` except Graywacke, but we
-- align fully with the canonical seed for consistency.

DELETE FROM woulder.location_rock_types
WHERE location_id = (SELECT id FROM woulder.locations WHERE name = 'Calendar Butte')
  AND rock_type_id IN (
      SELECT id FROM woulder.rock_types WHERE name IN ('Graywacke', 'Argillite', 'Phyllite')
  );

INSERT INTO woulder.location_rock_types (location_id, rock_type_id, is_primary)
SELECT l.id, rt.id, TRUE
FROM woulder.locations l
CROSS JOIN woulder.rock_types rt
WHERE l.name = 'Calendar Butte' AND rt.name = 'Granite'
ON CONFLICT (location_id, rock_type_id) DO UPDATE SET is_primary = TRUE;
