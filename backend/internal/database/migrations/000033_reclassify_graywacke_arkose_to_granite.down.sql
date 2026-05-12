-- 000033_reclassify_graywacke_arkose_to_granite.down.sql
--
-- Revert: move Graywacke and Arkose back to the "Sandstone" rock_type_group.
-- See up migration for the schema-drift note about `group_name` vs `name`.

UPDATE woulder.rock_types
SET rock_type_group_id = (
    SELECT id FROM woulder.rock_type_groups WHERE group_name = 'Sandstone'
)
WHERE name IN ('Graywacke', 'Arkose');
