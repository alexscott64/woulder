-- Rollback: Remove is_active column and index from areas table

DROP INDEX IF EXISTS woulder.idx_areas_is_active;

ALTER TABLE woulder.areas
DROP COLUMN IF EXISTS is_active;
