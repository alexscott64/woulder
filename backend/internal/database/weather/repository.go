// Package weather provides repository operations for weather data management.
// Weather data includes historical observations, current conditions, and forecasts
// for climbing locations.
package weather

import (
	"context"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// Repository defines operations for weather data management.
// All methods are safe for concurrent use.
type Repository interface {
	// Save saves or updates weather data for a location.
	// Uses upsert logic - if a record exists for the same location_id and timestamp,
	// it will be updated with new values.
	Save(ctx context.Context, data *models.WeatherData) error

	// GetHistorical retrieves weather data from the past N days for a location.
	// Results are ordered by timestamp ascending (oldest first).
	// Returns an empty slice if no data is found.
	GetHistorical(ctx context.Context, locationID int, days int) ([]models.WeatherData, error)

	// GetForecast retrieves forecast weather data for the next N hours for a location.
	// Results are ordered by timestamp ascending (nearest first).
	// Returns an empty slice if no forecast data is found.
	GetForecast(ctx context.Context, locationID int, hours int) ([]models.WeatherData, error)

	// GetCurrent retrieves the most recent weather data for a location.
	// Finds the record with timestamp closest to NOW() (can be past or future).
	// Returns database.ErrNotFound if no weather data exists for the location.
	GetCurrent(ctx context.Context, locationID int) (*models.WeatherData, error)

	// CleanOld deletes weather data older than the specified number of days.
	// This is a global cleanup operation across all locations.
	// Used for maintenance to remove stale historical data.
	CleanOld(ctx context.Context, daysToKeep int) error

	// DeleteOldForLocation deletes stale weather data for a specific location.
	// Deletes records where EITHER:
	// - The timestamp (observation time) is older than daysToKeep, OR
	// - The record was created more than daysToKeep ago (stale forecasts)
	DeleteOldForLocation(ctx context.Context, locationID int, daysToKeep int) error
}
