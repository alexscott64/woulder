package database

import (
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/alexscott64/woulder/backend/internal/models"
)

//go:embed setup_postgres.sql
var setupSQL string

type Database struct {
	conn *sql.DB
}

func New() (*Database, error) {
	// Build PostgreSQL connection string from environment variables
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	// Set defaults
	if port == "" {
		port = "5432"
	}
	if sslmode == "" {
		sslmode = "require"
	}

	// Validate required fields
	if host == "" || user == "" || password == "" || dbname == "" {
		return nil, fmt.Errorf("missing required database configuration: DB_HOST, DB_USER, DB_PASSWORD, DB_NAME must be set")
	}

	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	// Open PostgreSQL database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// PostgreSQL connection pool settings
	db.SetMaxOpenConns(25)                 // Maximum number of open connections
	db.SetMaxIdleConns(5)                  // Maximum number of idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Maximum lifetime of a connection

	database := &Database{conn: db}

	// Check if schema needs initialization
	needsInit, err := database.needsInitialization()
	if err != nil {
		return nil, fmt.Errorf("failed to check initialization status: %w", err)
	}

	if needsInit {
		log.Println("Database schema not found, running setup...")
		if err := database.runSetup(); err != nil {
			return nil, fmt.Errorf("failed to run database setup: %w", err)
		}
		log.Println("Database setup complete (schema + seed data)")
	}

	log.Printf("PostgreSQL connection established: %s@%s:%s/%s", user, host, port, dbname)

	return database, nil
}

// needsInitialization checks if the woulder schema exists
func (db *Database) needsInitialization() (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM information_schema.schemata WHERE schema_name = 'woulder')`
	err := db.conn.QueryRow(query).Scan(&exists)
	if err != nil {
		return false, err
	}
	return !exists, nil
}

// runSetup creates the database schema and seeds initial data
func (db *Database) runSetup() error {
	_, err := db.conn.Exec(setupSQL)
	return err
}

func (db *Database) Close() error {
	return db.conn.Close()
}

// GetAllLocations retrieves all saved locations
func (db *Database) GetAllLocations() ([]models.Location, error) {
	query := `SELECT id, name, latitude, longitude, elevation_ft, created_at, updated_at
	          FROM woulder.locations ORDER BY name`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []models.Location
	for rows.Next() {
		var loc models.Location
		if err := rows.Scan(&loc.ID, &loc.Name, &loc.Latitude, &loc.Longitude, &loc.ElevationFt, &loc.CreatedAt, &loc.UpdatedAt); err != nil {
			return nil, err
		}
		locations = append(locations, loc)
	}

	return locations, nil
}

// GetLocation retrieves a location by ID
func (db *Database) GetLocation(id int) (*models.Location, error) {
	query := `SELECT id, name, latitude, longitude, elevation_ft, created_at, updated_at
	          FROM woulder.locations WHERE id = $1`

	var loc models.Location
	err := db.conn.QueryRow(query, id).Scan(&loc.ID, &loc.Name, &loc.Latitude, &loc.Longitude, &loc.ElevationFt, &loc.CreatedAt, &loc.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &loc, nil
}

// SaveWeatherData saves weather data to the database
func (db *Database) SaveWeatherData(data *models.WeatherData) error {
	query := `
		INSERT INTO woulder.weather_data (
			location_id, timestamp, temperature, feels_like, precipitation,
			humidity, wind_speed, wind_direction, cloud_cover, pressure,
			description, icon
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
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

	_, err := db.conn.Exec(query,
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

// GetHistoricalWeather retrieves historical weather data for a location (past only, not future)
func (db *Database) GetHistoricalWeather(locationID int, days int) ([]models.WeatherData, error) {
	query := `
		SELECT id, location_id, timestamp, temperature, feels_like, precipitation,
			   humidity, wind_speed, wind_direction, cloud_cover, pressure,
			   description, icon, created_at
		FROM woulder.weather_data
		WHERE location_id = $1
		  AND timestamp >= NOW() - INTERVAL '1 day' * $2
		  AND timestamp <= NOW()
		ORDER BY timestamp ASC
	`

	rows, err := db.conn.Query(query, locationID, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var weatherData []models.WeatherData
	for rows.Next() {
		var data models.WeatherData
		if err := rows.Scan(
			&data.ID,
			&data.LocationID,
			&data.Timestamp,
			&data.Temperature,
			&data.FeelsLike,
			&data.Precipitation,
			&data.Humidity,
			&data.WindSpeed,
			&data.WindDirection,
			&data.CloudCover,
			&data.Pressure,
			&data.Description,
			&data.Icon,
			&data.CreatedAt,
		); err != nil {
			return nil, err
		}
		weatherData = append(weatherData, data)
	}

	return weatherData, nil
}

// GetForecastWeather retrieves future weather data (forecast) for a location
func (db *Database) GetForecastWeather(locationID int) ([]models.WeatherData, error) {
	query := `
		SELECT id, location_id, timestamp, temperature, feels_like, precipitation,
			   humidity, wind_speed, wind_direction, cloud_cover, pressure,
			   description, icon, created_at
		FROM woulder.weather_data
		WHERE location_id = $1 AND timestamp > NOW()
		ORDER BY timestamp ASC
	`

	rows, err := db.conn.Query(query, locationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var weatherData []models.WeatherData
	for rows.Next() {
		var data models.WeatherData
		if err := rows.Scan(
			&data.ID,
			&data.LocationID,
			&data.Timestamp,
			&data.Temperature,
			&data.FeelsLike,
			&data.Precipitation,
			&data.Humidity,
			&data.WindSpeed,
			&data.WindDirection,
			&data.CloudCover,
			&data.Pressure,
			&data.Description,
			&data.Icon,
			&data.CreatedAt,
		); err != nil {
			return nil, err
		}
		weatherData = append(weatherData, data)
	}

	return weatherData, nil
}

// CleanOldWeatherData removes weather data older than specified days
func (db *Database) CleanOldWeatherData(days int) error {
	query := `DELETE FROM woulder.weather_data WHERE timestamp < NOW() - INTERVAL '1 day' * $1`
	_, err := db.conn.Exec(query, days)
	return err
}

// GetCurrentWeather retrieves the most recent weather data for a location (closest to now)
func (db *Database) GetCurrentWeather(locationID int) (*models.WeatherData, error) {
	// Use PostgreSQL's ABS and ORDER BY to find the closest timestamp
	query := `
		SELECT id, location_id, timestamp, temperature, feels_like, precipitation,
			   humidity, wind_speed, wind_direction, cloud_cover, pressure,
			   description, icon, created_at
		FROM woulder.weather_data
		WHERE location_id = $1
		ORDER BY ABS(EXTRACT(EPOCH FROM (timestamp - NOW())))
		LIMIT 1
	`

	var data models.WeatherData
	err := db.conn.QueryRow(query, locationID).Scan(
		&data.ID,
		&data.LocationID,
		&data.Timestamp,
		&data.Temperature,
		&data.FeelsLike,
		&data.Precipitation,
		&data.Humidity,
		&data.WindSpeed,
		&data.WindDirection,
		&data.CloudCover,
		&data.Pressure,
		&data.Description,
		&data.Icon,
		&data.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

// GetLastRefreshTime returns the most recent weather data timestamp for any location
func (db *Database) GetLastRefreshTime() (string, error) {
	query := `SELECT MAX(created_at) FROM woulder.weather_data`
	var lastRefresh sql.NullTime
	err := db.conn.QueryRow(query).Scan(&lastRefresh)
	if err != nil {
		return "", err
	}
	if !lastRefresh.Valid {
		return "", nil
	}
	return lastRefresh.Time.Format(time.RFC3339), nil
}

// GetRiversByLocation retrieves all rivers for a location
func (db *Database) GetRiversByLocation(locationID int) ([]models.River, error) {
	query := `SELECT id, location_id, gauge_id, river_name, safe_crossing_cfs, caution_crossing_cfs,
			  drainage_area_sq_mi, gauge_drainage_area_sq_mi, flow_divisor, is_estimated, description, created_at, updated_at
			  FROM woulder.rivers WHERE location_id = $1 ORDER BY river_name`

	rows, err := db.conn.Query(query, locationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rivers []models.River
	for rows.Next() {
		var river models.River
		if err := rows.Scan(&river.ID, &river.LocationID, &river.GaugeID, &river.RiverName,
			&river.SafeCrossingCFS, &river.CautionCrossingCFS, &river.DrainageAreaSqMi,
			&river.GaugeDrainageAreaSqMi, &river.FlowDivisor, &river.IsEstimated, &river.Description,
			&river.CreatedAt, &river.UpdatedAt); err != nil {
			return nil, err
		}
		rivers = append(rivers, river)
	}

	return rivers, nil
}

// GetRiverByID retrieves a specific river by ID
func (db *Database) GetRiverByID(riverID int) (*models.River, error) {
	query := `SELECT id, location_id, gauge_id, river_name, safe_crossing_cfs, caution_crossing_cfs,
			  drainage_area_sq_mi, gauge_drainage_area_sq_mi, flow_divisor, is_estimated, description, created_at, updated_at
			  FROM woulder.rivers WHERE id = $1`

	var river models.River
	err := db.conn.QueryRow(query, riverID).Scan(&river.ID, &river.LocationID, &river.GaugeID,
		&river.RiverName, &river.SafeCrossingCFS, &river.CautionCrossingCFS, &river.DrainageAreaSqMi,
		&river.GaugeDrainageAreaSqMi, &river.FlowDivisor, &river.IsEstimated, &river.Description,
		&river.CreatedAt, &river.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &river, nil
}
