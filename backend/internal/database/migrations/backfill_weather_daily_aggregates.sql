INSERT INTO woulder.weather_daily_aggregates (
    location_id,
    local_date,
    min_temperature,
    max_temperature,
    avg_temperature,
    total_precipitation,
    avg_humidity,
    avg_wind_speed,
    snow_estimate_inches,
    sunrise_at,
    sunset_at,
    source_hour_count
)
SELECT
    wd.location_id,
    (wd.timestamp AT TIME ZONE 'America/Los_Angeles')::date AS local_date,
    MIN(wd.temperature) AS min_temperature,
    MAX(wd.temperature) AS max_temperature,
    AVG(wd.temperature) AS avg_temperature,
    SUM(wd.precipitation) AS total_precipitation,
    AVG(wd.humidity::numeric) AS avg_humidity,
    AVG(wd.wind_speed) AS avg_wind_speed,
    SUM(
        CASE
            WHEN wd.temperature <= 30 THEN wd.precipitation * 10.0
            WHEN wd.temperature >= 34 THEN 0
            ELSE wd.precipitation * ((34 - wd.temperature) / 4.0) * 10.0
        END
    ) AS snow_estimate_inches,
    MIN(CASE WHEN EXTRACT(HOUR FROM (wd.timestamp AT TIME ZONE 'America/Los_Angeles')) BETWEEN 4 AND 11 THEN wd.timestamp END) AS sunrise_at,
    MAX(CASE WHEN EXTRACT(HOUR FROM (wd.timestamp AT TIME ZONE 'America/Los_Angeles')) BETWEEN 15 AND 22 THEN wd.timestamp END) AS sunset_at,
    COUNT(*)::int AS source_hour_count
FROM woulder.weather_data wd
GROUP BY wd.location_id, (wd.timestamp AT TIME ZONE 'America/Los_Angeles')::date
ON CONFLICT (location_id, local_date) DO UPDATE SET
    min_temperature = EXCLUDED.min_temperature,
    max_temperature = EXCLUDED.max_temperature,
    avg_temperature = EXCLUDED.avg_temperature,
    total_precipitation = EXCLUDED.total_precipitation,
    avg_humidity = EXCLUDED.avg_humidity,
    avg_wind_speed = EXCLUDED.avg_wind_speed,
    snow_estimate_inches = EXCLUDED.snow_estimate_inches,
    sunrise_at = EXCLUDED.sunrise_at,
    sunset_at = EXCLUDED.sunset_at,
    source_hour_count = EXCLUDED.source_hour_count,
    updated_at = CURRENT_TIMESTAMP;