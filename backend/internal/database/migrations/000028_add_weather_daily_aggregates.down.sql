-- Migration: 000028_add_weather_daily_aggregates (rollback)
-- Purpose: Remove daily weather rollups table

DROP TRIGGER IF EXISTS update_weather_daily_aggregates_updated_at ON woulder.weather_daily_aggregates;
DROP TABLE IF EXISTS woulder.weather_daily_aggregates;