package database

import (
	"context"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// SaveWeatherData saves or updates weather data for a location
func (db *Database) SaveWeatherData(ctx context.Context, data *models.WeatherData) error {
	query := `
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

	_, err := db.conn.ExecContext(ctx, query,
		data.LocationID,
		data.Timestamp,
		data.Temperature,
		data.FeelsLike,
		data.Precipitation,
		data.Humidity,
		data.WindSpeed,
		data.WindDirection,
		data.CloudCover,
		data.Pressure,
		data.Description,
		data.Icon,
	)

	return err
}

// GetHistoricalWeather retrieves historical weather data for a location
func (db *Database) GetHistoricalWeather(ctx context.Context, locationID, days int) ([]models.WeatherData, error) {
	query := `
		SELECT id, location_id, timestamp, temperature, feels_like,
		       precipitation, humidity, wind_speed, wind_direction,
		       cloud_cover, pressure, description, icon, created_at
		FROM woulder.weather_data
		WHERE location_id = $1
		  AND timestamp >= NOW() - INTERVAL '1 day' * $2
		  AND timestamp <= NOW()
		ORDER BY timestamp ASC
	`

	rows, err := db.conn.QueryContext(ctx, query, locationID, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.WeatherData
	for rows.Next() {
		var d models.WeatherData
		if err := rows.Scan(
			&d.ID, &d.LocationID, &d.Timestamp,
			&d.Temperature, &d.FeelsLike, &d.Precipitation,
			&d.Humidity, &d.WindSpeed, &d.WindDirection,
			&d.CloudCover, &d.Pressure, &d.Description,
			&d.Icon, &d.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, d)
	}

	return result, nil
}

// GetForecastWeather retrieves forecast weather data for a location
func (db *Database) GetForecastWeather(
	ctx context.Context,
	locationID int,
	hours int,
) ([]models.WeatherData, error) {

	query := `
		SELECT id, location_id, timestamp, temperature, feels_like,
		       precipitation, humidity, wind_speed, wind_direction,
		       cloud_cover, pressure, description, icon, created_at
		FROM woulder.weather_data
		WHERE location_id = $1
		  AND timestamp > NOW()
		  AND timestamp <= NOW() + ($2 * INTERVAL '1 hour')
		ORDER BY timestamp ASC
	`

	rows, err := db.conn.QueryContext(ctx, query, locationID, hours)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.WeatherData
	for rows.Next() {
		var d models.WeatherData
		if err := rows.Scan(
			&d.ID,
			&d.LocationID,
			&d.Timestamp,
			&d.Temperature,
			&d.FeelsLike,
			&d.Precipitation,
			&d.Humidity,
			&d.WindSpeed,
			&d.WindDirection,
			&d.CloudCover,
			&d.Pressure,
			&d.Description,
			&d.Icon,
			&d.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, d)
	}

	return result, nil
}

// GetCurrentWeather retrieves the current weather data for a location
func (db *Database) GetCurrentWeather(ctx context.Context, locationID int) (*models.WeatherData, error) {
	query := `
		SELECT id, location_id, timestamp, temperature, feels_like,
		       precipitation, humidity, wind_speed, wind_direction,
		       cloud_cover, pressure, description, icon, created_at
		FROM woulder.weather_data
		WHERE location_id = $1
		ORDER BY ABS(EXTRACT(EPOCH FROM (timestamp - NOW())))
		LIMIT 1
	`

	var d models.WeatherData
	err := db.conn.QueryRowContext(ctx, query, locationID).Scan(
		&d.ID, &d.LocationID, &d.Timestamp,
		&d.Temperature, &d.FeelsLike, &d.Precipitation,
		&d.Humidity, &d.WindSpeed, &d.WindDirection,
		&d.CloudCover, &d.Pressure, &d.Description,
		&d.Icon, &d.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &d, nil
}

// CleanOldWeatherData deletes weather data older than specified days
func (db *Database) CleanOldWeatherData(ctx context.Context, days int) error {
	query := `DELETE FROM woulder.weather_data WHERE timestamp < NOW() - INTERVAL '1 day' * $1`
	_, err := db.conn.ExecContext(ctx, query, days)
	return err
}

// DeleteOldWeatherData deletes weather data older than specified days for a specific location
// Deletes based on both timestamp AND created_at to handle stale forecast data
func (db *Database) DeleteOldWeatherData(ctx context.Context, locationID int, daysToKeep int) error {
	// Delete records where either:
	// 1. The timestamp (observation time) is old, OR
	// 2. The record was created more than daysToKeep ago (stale forecast data)
	query := `DELETE FROM woulder.weather_data
	          WHERE location_id = $1
	          AND (timestamp < NOW() - INTERVAL '1 day' * $2
	               OR created_at < NOW() - INTERVAL '1 day' * $2)`
	_, err := db.conn.ExecContext(ctx, query, locationID, daysToKeep)
	return err
}
