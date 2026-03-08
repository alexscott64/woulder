SELECT COUNT(*) AS aggregate_rows
FROM woulder.weather_daily_aggregates;

SELECT
    location_id,
    MIN(local_date) AS first_day,
    MAX(local_date) AS last_day,
    COUNT(*) AS day_count
FROM woulder.weather_daily_aggregates
GROUP BY location_id
ORDER BY location_id
LIMIT 10;