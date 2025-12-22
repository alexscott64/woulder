package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/alexscott64/woulder/backend/internal/models"
)

type Database struct {
	conn *sql.DB
}

func New() (*Database, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Database connection established")

	return &Database{conn: db}, nil
}

func (db *Database) Close() error {
	return db.conn.Close()
}

// GetAllLocations retrieves all saved locations
func (db *Database) GetAllLocations() ([]models.Location, error) {
	query := `SELECT id, name, latitude, longitude, created_at, updated_at FROM locations ORDER BY name`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []models.Location
	for rows.Next() {
		var loc models.Location
		if err := rows.Scan(&loc.ID, &loc.Name, &loc.Latitude, &loc.Longitude, &loc.CreatedAt, &loc.UpdatedAt); err != nil {
			return nil, err
		}
		locations = append(locations, loc)
	}

	return locations, nil
}

// GetLocation retrieves a location by ID
func (db *Database) GetLocation(id int) (*models.Location, error) {
	query := `SELECT id, name, latitude, longitude, created_at, updated_at FROM locations WHERE id = ?`

	var loc models.Location
	err := db.conn.QueryRow(query, id).Scan(&loc.ID, &loc.Name, &loc.Latitude, &loc.Longitude, &loc.CreatedAt, &loc.UpdatedAt)
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
		ON DUPLICATE KEY UPDATE
			temperature = VALUES(temperature),
			feels_like = VALUES(feels_like),
			precipitation = VALUES(precipitation),
			humidity = VALUES(humidity),
			wind_speed = VALUES(wind_speed),
			wind_direction = VALUES(wind_direction),
			cloud_cover = VALUES(cloud_cover),
			pressure = VALUES(pressure),
			description = VALUES(description),
			icon = VALUES(icon)
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
		WHERE location_id = ? AND timestamp >= DATE_SUB(NOW(), INTERVAL ? DAY)
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

// CleanOldWeatherData removes weather data older than specified days
func (db *Database) CleanOldWeatherData(days int) error {
	query := `DELETE FROM weather_data WHERE timestamp < DATE_SUB(NOW(), INTERVAL ? DAY)`
	_, err := db.conn.Exec(query, days)
	return err
}
