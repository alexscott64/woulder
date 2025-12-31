-- Drop the foreign key constraint and column from rock_types
ALTER TABLE woulder.rock_types DROP COLUMN IF EXISTS rock_type_group_id;

-- Drop the rock_type_groups table
DROP TABLE IF EXISTS woulder.rock_type_groups;
