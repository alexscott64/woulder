package calculator

import (
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// GetTotalPrecipitation sums up precipitation from weather data array
func GetTotalPrecipitation(weatherData []models.WeatherData) float64 {
	total := 0.0
	for _, data := range weatherData {
		total += data.Precipitation
	}
	return total
}

// GetPrecipitationInWindow calculates precipitation over last N hours from now
func GetPrecipitationInWindow(weatherData []models.WeatherData, hoursAgo int) float64 {
	cutoffTime := time.Now().Add(-time.Duration(hoursAgo) * time.Hour)
	total := 0.0

	for _, data := range weatherData {
		if data.Timestamp.After(cutoffTime) || data.Timestamp.Equal(cutoffTime) {
			total += data.Precipitation
		}
	}

	return total
}

// HasPersistentPrecipitation checks if multiple recent periods had measurable precipitation
func HasPersistentPrecipitation(recentWeather []models.WeatherData, threshold float64) bool {
	if len(recentWeather) == 0 {
		return false
	}

	periodsWithPrecip := 0
	for _, w := range recentWeather {
		if w.Precipitation > threshold {
			periodsWithPrecip++
		}
	}

	return periodsWithPrecip >= 2 // 2+ periods = persistent
}
