-- Migration: 000028_add_weather_daily_aggregates
-- Purpose: Persist daily weather rollups for long-term historical analytics

CREATE TABLE IF NOT EXISTS woulder.weather_daily_aggregates (
    id SERIAL PRIMARY KEY,
    location_id INTEGER NOT NULL REFERENCES woulder.locations(id) ON DELETE CASCADE,
    local_date DATE NOT NULL,
    min_temperature DECIMAL(5, 2) NOT NULL,
    max_temperature DECIMAL(5, 2) NOT NULL,
    avg_temperature DECIMAL(5, 2) NOT NULL,
    total_precipitation DECIMAL(8, 3) NOT NULL DEFAULT 0 CHECK (total_precipitation >= 0),
    avg_humidity DECIMAL(5, 2) NOT NULL DEFAULT 0 CHECK (avg_humidity >= 0 AND avg_humidity <= 100),
    avg_wind_speed DECIMAL(6, 2) NOT NULL DEFAULT 0 CHECK (avg_wind_speed >= 0),
    snow_estimate_inches DECIMAL(8, 2) NOT NULL DEFAULT 0 CHECK (snow_estimate_inches >= 0),
    sunrise_at TIMESTAMPTZ,
    sunset_at TIMESTAMPTZ,
    source_hour_count INTEGER NOT NULL DEFAULT 0 CHECK (source_hour_count >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (location_id, local_date)
);

CREATE INDEX IF NOT EXISTS idx_weather_daily_aggregates_location_date
    ON woulder.weather_daily_aggregates(location_id, local_date DESC);

CREATE INDEX IF NOT EXISTS idx_weather_daily_aggregates_local_date
    ON woulder.weather_daily_aggregates(local_date DESC);

DROP TRIGGER IF EXISTS update_weather_daily_aggregates_updated_at ON woulder.weather_daily_aggregates;
CREATE TRIGGER update_weather_daily_aggregates_updated_at
    BEFORE UPDATE ON woulder.weather_daily_aggregates
    FOR EACH ROW EXECUTE FUNCTION woulder.update_updated_at_column();

COMMENT ON TABLE woulder.weather_daily_aggregates IS 'Daily weather rollups for long-term historical analytics';
COMMENT ON COLUMN woulder.weather_daily_aggregates.local_date IS 'Local calendar date in America/Los_Angeles timezone';
COMMENT ON COLUMN woulder.weather_daily_aggregates.source_hour_count IS 'Number of hourly samples included in the aggregate';