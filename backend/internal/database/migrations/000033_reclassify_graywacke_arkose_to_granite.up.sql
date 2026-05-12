-- 000033_reclassify_graywacke_arkose_to_granite.up.sql
--
-- Reclassify Graywacke and Arkose from the "Sandstone" rock_type_group into "Granite".
-- These hard, quartz-cemented sandstones behave thermally and (in PNW use cases) frictionally
-- much closer to granite than to soft desert sandstones with desert varnish. The previous
-- classification caused the rock-temp solver to apply Sandstone's high absorptivity (α=0.75)
-- and inflated surface-temperature predictions for Calendar Butte (location_id=9) and any
-- future Graywacke/Arkose locations.
--
-- NOTE: schema drift — the live `woulder.rock_type_groups` table uses `group_name`, not `name`
-- as declared in migration 000005. Live queries already use `group_name`, so this migration
-- matches production reality.

UPDATE woulder.rock_types
SET rock_type_group_id = (
    SELECT id FROM woulder.rock_type_groups WHERE group_name = 'Granite'
)
WHERE name IN ('Graywacke', 'Arkose');
