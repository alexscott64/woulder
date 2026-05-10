-- Rollback for 000032_add_radiation_dewpoint
-- Drops the radiation and dewpoint columns added to weather_data.

ALTER TABLE woulder.weather_data
    DROP COLUMN IF EXISTS shortwave_radiation,
    DROP COLUMN IF EXISTS direct_radiation,
    DROP COLUMN IF EXISTS diffuse_radiation,
    DROP COLUMN IF EXISTS dewpoint_f;
