-- Migration 000007 Rollback: Remove Mountain Project climb tracking tables

-- Drop triggers first
DROP TRIGGER IF EXISTS update_mp_ticks_updated_at ON woulder.mp_ticks;
DROP TRIGGER IF EXISTS update_mp_routes_updated_at ON woulder.mp_routes;
DROP TRIGGER IF EXISTS update_mp_areas_updated_at ON woulder.mp_areas;

-- Drop tables in reverse order (handle foreign keys)
DROP TABLE IF EXISTS woulder.mp_ticks;
DROP TABLE IF EXISTS woulder.mp_routes;
DROP TABLE IF EXISTS woulder.mp_areas;
