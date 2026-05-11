-- Migration: 000032_add_radiation_dewpoint
-- Purpose: Add solar radiation and dewpoint columns to weather_data for the
--          rock_temp (rock surface temperature) calculator. The calculator
--          performs an energy-balance model that needs direct/diffuse radiation
--          on the face plus dewpoint for condensation detection.
--
-- Performance: ALTER TABLE ... ADD COLUMN ... DEFAULT 0 is metadata-only on
-- Postgres 11+, so this is safe on a populated weather_data table.

ALTER TABLE woulder.weather_data
    ADD COLUMN IF NOT EXISTS shortwave_radiation DECIMAL(7, 2) NOT NULL DEFAULT 0
        CONSTRAINT weather_data_shortwave_radiation_check
            CHECK (shortwave_radiation >= 0),
    ADD COLUMN IF NOT EXISTS direct_radiation    DECIMAL(7, 2) NOT NULL DEFAULT 0
        CONSTRAINT weather_data_direct_radiation_check
            CHECK (direct_radiation >= 0),
    ADD COLUMN IF NOT EXISTS diffuse_radiation   DECIMAL(7, 2) NOT NULL DEFAULT 0
        CONSTRAINT weather_data_diffuse_radiation_check
            CHECK (diffuse_radiation >= 0),
    ADD COLUMN IF NOT EXISTS dewpoint_f          DECIMAL(5, 2) NOT NULL DEFAULT 0;

COMMENT ON COLUMN woulder.weather_data.shortwave_radiation IS 'Total shortwave (solar) radiation on horizontal surface, W/m^2';
COMMENT ON COLUMN woulder.weather_data.direct_radiation    IS 'Direct beam solar radiation on horizontal surface, W/m^2';
COMMENT ON COLUMN woulder.weather_data.diffuse_radiation   IS 'Diffuse (scattered) solar radiation on horizontal surface, W/m^2';
COMMENT ON COLUMN woulder.weather_data.dewpoint_f          IS 'Dewpoint temperature in Fahrenheit (used for condensation detection)';
