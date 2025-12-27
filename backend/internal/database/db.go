package database

import (
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
	"github.com/alexscott64/woulder/backend/internal/models"
)

//go:embed schema.sql
var schemaSQL string

//go:embed seed.sql
var seedSQL string

type Database struct {
	conn *sql.DB
}

func New() (*Database, error) {
	// Get database path from env or use default
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "woulder.db"
	}

	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	// Check if database exists (for initialization)
	isNewDB := false
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		isNewDB = true
	}

	// Open SQLite database
	db, err := sql.Open("sqlite", dbPath+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// SQLite connection settings
	db.SetMaxOpenConns(1) // SQLite works best with single writer
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0) // Connections don't expire

	database := &Database{conn: db}

	// Initialize schema if new database
	if isNewDB {
		log.Println("New database detected, initializing schema...")
		if err := database.initSchema(); err != nil {
			return nil, fmt.Errorf("failed to initialize schema: %w", err)
		}
		log.Println("Database schema initialized")

		log.Println("Seeding initial data...")
		if err := database.seedData(); err != nil {
			return nil, fmt.Errorf("failed to seed data: %w", err)
		}
		log.Println("Initial data seeded")
	}

	log.Printf("Database connection established: %s", dbPath)

	return database, nil
}

// initSchema creates the database tables
func (db *Database) initSchema() error {
	_, err := db.conn.Exec(schemaSQL)
	return err
}

// seedData inserts initial location and river data
func (db *Database) seedData() error {
	_, err := db.conn.Exec(seedSQL)
	return err
}

func (db *Database) Close() error {
	return db.conn.Close()
}

// GetAllLocations retrieves all saved locations
func (db *Database) GetAllLocations() ([]models.Location, error) {
	query := `SELECT id, name, latitude, longitude, elevation_ft, created_at, updated_at FROM locations ORDER BY name`

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
	query := `SELECT id, name, latitude, longitude, elevation_ft, created_at, updated_at FROM locations WHERE id = ?`

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
		INSERT INTO weather_data (
			location_id, timestamp, temperature, feels_like, precipitation,
			humidity, wind_speed, wind_direction, cloud_cover, pressure,
			description, icon
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(location_id, timestamp) DO UPDATE SET
			temperature = excluded.temperature,
			feels_like = excluded.feels_like,
			precipitation = excluded.precipitation,
			humidity = excluded.humidity,
			wind_speed = excluded.wind_speed,
			wind_direction = excluded.wind_direction,
			cloud_cover = excluded.cloud_cover,
			pressure = excluded.pressure,
			description = excluded.description,
			icon = excluded.icon
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

// GetHistoricalWeather retrieves historical weather data for a location
func (db *Database) GetHistoricalWeather(locationID int, days int) ([]models.WeatherData, error) {
	query := `
		SELECT id, location_id, timestamp, temperature, feels_like, precipitation,
			   humidity, wind_speed, wind_direction, cloud_cover, pressure,
			   description, icon, created_at
		FROM weather_data
		WHERE location_id = ? AND timestamp >= datetime('now', ? || ' days')
		ORDER BY timestamp ASC
	`

	rows, err := db.conn.Query(query, locationID, -days)
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
		FROM weather_data
		WHERE location_id = ? AND timestamp > datetime('now')
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
	query := `DELETE FROM weather_data WHERE timestamp < datetime('now', ? || ' days')`
	_, err := db.conn.Exec(query, -days)
	return err
}

// GetRiversByLocation retrieves all rivers for a location
func (db *Database) GetRiversByLocation(locationID int) ([]models.River, error) {
	query := `SELECT id, location_id, gauge_id, river_name, safe_crossing_cfs, caution_crossing_cfs,
			  drainage_area_sq_mi, gauge_drainage_area_sq_mi, is_estimated, description, created_at, updated_at
			  FROM rivers WHERE location_id = ? ORDER BY river_name`

	rows, err := db.conn.Query(query, locationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rivers []models.River
	for rows.Next() {
		var river models.River
		var isEstimated int
		if err := rows.Scan(&river.ID, &river.LocationID, &river.GaugeID, &river.RiverName,
			&river.SafeCrossingCFS, &river.CautionCrossingCFS, &river.DrainageAreaSqMi,
			&river.GaugeDrainageAreaSqMi, &isEstimated, &river.Description,
			&river.CreatedAt, &river.UpdatedAt); err != nil {
			return nil, err
		}
		river.IsEstimated = isEstimated == 1
		rivers = append(rivers, river)
	}

	return rivers, nil
}

// GetRiverByID retrieves a specific river by ID
func (db *Database) GetRiverByID(riverID int) (*models.River, error) {
	query := `SELECT id, location_id, gauge_id, river_name, safe_crossing_cfs, caution_crossing_cfs,
			  drainage_area_sq_mi, gauge_drainage_area_sq_mi, is_estimated, description, created_at, updated_at
			  FROM rivers WHERE id = ?`

	var river models.River
	var isEstimated int
	err := db.conn.QueryRow(query, riverID).Scan(&river.ID, &river.LocationID, &river.GaugeID,
		&river.RiverName, &river.SafeCrossingCFS, &river.CautionCrossingCFS, &river.DrainageAreaSqMi,
		&river.GaugeDrainageAreaSqMi, &isEstimated, &river.Description,
		&river.CreatedAt, &river.UpdatedAt)
	if err != nil {
		return nil, err
	}
	river.IsEstimated = isEstimated == 1

	return &river, nil
}
