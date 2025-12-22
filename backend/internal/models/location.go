package models

import "time"

// Location represents a saved weather location
type Location struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Latitude  float64   `json:"latitude" db:"latitude"`
	Longitude float64   `json:"longitude" db:"longitude"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// WeatherData represents weather information for a location
type WeatherData struct {
	ID            int       `json:"id" db:"id"`
	LocationID    int       `json:"location_id" db:"location_id"`
	Timestamp     time.Time `json:"timestamp" db:"timestamp"`
	Temperature   float64   `json:"temperature" db:"temperature"`       // Fahrenheit
	FeelsLike     float64   `json:"feels_like" db:"feels_like"`         // Fahrenheit
	Precipitation float64   `json:"precipitation" db:"precipitation"`   // inches
	Humidity      int       `json:"humidity" db:"humidity"`             // percentage
	WindSpeed     float64   `json:"wind_speed" db:"wind_speed"`         // mph
	WindDirection int       `json:"wind_direction" db:"wind_direction"` // degrees
	CloudCover    int       `json:"cloud_cover" db:"cloud_cover"`       // percentage
	Pressure      int       `json:"pressure" db:"pressure"`             // hPa
	Description   string    `json:"description" db:"description"`       // e.g. "light rain"
	Icon          string    `json:"icon" db:"icon"`                     // OpenWeatherMap icon code
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// WeatherForecast represents forecast data
type WeatherForecast struct {
	LocationID int           `json:"location_id"`
	Location   Location      `json:"location"`
	Current    WeatherData   `json:"current"`
	Hourly     []WeatherData `json:"hourly"`
	Historical []WeatherData `json:"historical"`
}
