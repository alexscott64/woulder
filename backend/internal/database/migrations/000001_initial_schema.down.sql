-- Rollback initial schema migration
-- Drops all tables, functions, and schema

-- Drop triggers first
DROP TRIGGER IF EXISTS update_rivers_updated_at ON woulder.rivers;
DROP TRIGGER IF EXISTS update_locations_updated_at ON woulder.locations;
DROP TRIGGER IF EXISTS update_areas_updated_at ON woulder.areas;

-- Drop function
DROP FUNCTION IF EXISTS woulder.update_updated_at_column();

-- Drop tables in reverse order (respecting foreign keys)
DROP TABLE IF EXISTS woulder.rivers CASCADE;
DROP TABLE IF EXISTS woulder.weather_data CASCADE;
DROP TABLE IF EXISTS woulder.locations CASCADE;
DROP TABLE IF EXISTS woulder.areas CASCADE;

-- Drop schema
DROP SCHEMA IF EXISTS woulder CASCADE;
