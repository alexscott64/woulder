-- Rollback migration 000029: Remove grade_order column
DROP INDEX IF EXISTS woulder.idx_mp_routes_grade_order;
ALTER TABLE woulder.mp_routes DROP COLUMN IF EXISTS grade_order;
