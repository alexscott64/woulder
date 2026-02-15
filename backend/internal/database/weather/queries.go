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
)
