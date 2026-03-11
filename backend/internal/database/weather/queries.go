package weather

// SQL queries for weather operations.
// These queries are separated from implementation for clarity and reusability.

const (
	// querySave inserts or updates weather data.
	// Uses ON CONFLICT to handle upserts efficiently.
	// Indexes: (location_id, timestamp) UNIQUE for upsert performance
	querySave = `
		INSERT INTO woulder.weather_data (
			location_id, timestamp, temperature, feels_like, precipitation,
			humidity, wind_speed, wind_direction, cloud_cover, pressure,
			description, icon
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT(location_id, timestamp) DO UPDATE SET
			temperature = EXCLUDED.temperature,
			feels_like = EXCLUDED.feels_like,
			precipitation = EXCLUDED.precipitation,
			humidity = EXCLUDED.humidity,
			wind_speed = EXCLUDED.wind_speed,
			wind_direction = EXCLUDED.wind_direction,
			cloud_cover = EXCLUDED.cloud_cover,
			pressure = EXCLUDED.pressure,
			description = EXCLUDED.description,
			icon = EXCLUDED.icon
	`

	// queryGetHistorical retrieves past weather data for a location.
	// Filters by location_id and date range (past N days).
	// Index: (location_id, timestamp) for efficient range scans
	queryGetHistorical = `
		SELECT id, location_id, timestamp, temperature, feels_like,
		       precipitation, humidity, wind_speed, wind_direction,
		       cloud_cover, pressure, description, icon, created_at
		FROM woulder.weather_data
		WHERE location_id = $1
		  AND timestamp >= NOW() - INTERVAL '1 day' * $2
		  AND timestamp <= NOW()
		ORDER BY timestamp ASC
	`

	// queryGetForecast retrieves future weather data (forecast) for a location.
	// Filters by location_id and date range (next N hours).
	// Index: (location_id, timestamp) for efficient range scans
	queryGetForecast = `
		SELECT id, location_id, timestamp, temperature, feels_like,
		       precipitation, humidity, wind_speed, wind_direction,
		       cloud_cover, pressure, description, icon, created_at
		FROM woulder.weather_data
		WHERE location_id = $1
		  AND timestamp > NOW()
		  AND timestamp <= NOW() + ($2 * INTERVAL '1 hour')
		ORDER BY timestamp ASC
	`

	// queryGetCurrent retrieves the weather record closest to NOW().
	// Uses ABS(EXTRACT(EPOCH FROM ...)) to find nearest timestamp.
	// Can return past or future record depending on data availability.
	// Index: location_id for filtering, then full scan for MIN(time_diff)
	queryGetCurrent = `
		SELECT id, location_id, timestamp, temperature, feels_like,
		       precipitation, humidity, wind_speed, wind_direction,
		       cloud_cover, pressure, description, icon, created_at
		FROM woulder.weather_data
		WHERE location_id = $1
		ORDER BY ABS(EXTRACT(EPOCH FROM (timestamp - NOW())))
		LIMIT 1
	`

	// queryCleanOld deletes weather data older than specified days (global).
	// Deletes based on timestamp alone.
	// Should be run during off-peak hours as it can affect many rows.
	queryCleanOld = `
		DELETE FROM woulder.weather_data
		WHERE timestamp < NOW() - INTERVAL '1 day' * $1
	`

	// queryDeleteOldForLocation deletes stale weather data for a location.
	// Deletes records where timestamp is old OR created_at is old (stale forecasts).
	// Index: location_id, timestamp, created_at for efficient deletion
	queryDeleteOldForLocation = `
		DELETE FROM woulder.weather_data
		WHERE location_id = $1
		  AND (timestamp < NOW() - INTERVAL '1 day' * $2
		   OR created_at < NOW() - INTERVAL '1 day' * $2)
	`

	// queryDeleteFutureForLocation deletes ALL future forecast data for a specific location.
	// This purges stale forecasts before saving fresh data, ensuring no stale
	// future values persist (e.g., from timezone bugs or model changes).
	queryDeleteFutureForLocation = `
		DELETE FROM woulder.weather_data
		WHERE location_id = $1
		  AND timestamp > NOW()
	`

	// queryUpsertDailyAggregates rolls up hourly weather_data rows into daily aggregates
	// for a location over an inclusive local-date range in America/Los_Angeles.
	queryUpsertDailyAggregates = `
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
		WHERE wd.location_id = $1
		  AND (wd.timestamp AT TIME ZONE 'America/Los_Angeles')::date >= $2::date
		  AND (wd.timestamp AT TIME ZONE 'America/Los_Angeles')::date <= $3::date
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
			updated_at = CURRENT_TIMESTAMP
	`

	// queryGetDailyAggregates returns daily aggregate rows for a location and date range.
	queryGetDailyAggregates = `
		SELECT id, location_id, local_date::text, min_temperature, max_temperature,
		       avg_temperature, total_precipitation, avg_humidity, avg_wind_speed,
		       snow_estimate_inches, sunrise_at, sunset_at, source_hour_count,
		       created_at, updated_at
		FROM woulder.weather_daily_aggregates
		WHERE location_id = $1
		  AND local_date >= $2::date
		  AND local_date <= $3::date
		ORDER BY local_date ASC
	`
)
