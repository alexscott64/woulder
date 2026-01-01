package database

import (
	"context"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// Repository defines the interface for all database operations.
// This interface allows for easy mocking in tests and decouples
// business logic from the database implementation.
type Repository interface {
	// Location operations
	GetAllLocations(ctx context.Context) ([]models.Location, error)
	GetLocation(ctx context.Context, id int) (*models.Location, error)
	GetLocationsByArea(ctx context.Context, areaID int) ([]models.Location, error)

	// Weather operations
	SaveWeatherData(ctx context.Context, data *models.WeatherData) error
	GetHistoricalWeather(ctx context.Context, locationID int, hours int) ([]models.WeatherData, error)
	GetForecastWeather(ctx context.Context, locationID int, hours int) ([]models.WeatherData, error)
	GetCurrentWeather(ctx context.Context, locationID int) (*models.WeatherData, error)
	CleanOldWeatherData(ctx context.Context, daysToKeep int) error

	// River operations
	GetRiversByLocation(ctx context.Context, locationID int) ([]models.River, error)
	GetRiverByID(ctx context.Context, id int) (*models.River, error)

	// Area operations
	GetAllAreas(ctx context.Context) ([]models.Area, error)
	GetAreasWithLocationCounts(ctx context.Context) ([]models.AreaWithLocationCount, error)
	GetAreaByID(ctx context.Context, id int) (*models.Area, error)

	// Rock type operations
	GetRockTypesByLocation(ctx context.Context, locationID int) ([]models.RockType, error)
	GetPrimaryRockType(ctx context.Context, locationID int) (*models.RockType, error)
	GetSunExposureByLocation(ctx context.Context, locationID int) (*models.LocationSunExposure, error)

	// Health check
	Ping(ctx context.Context) error

	// Cleanup
	Close() error
}

// Ensure Database implements Repository interface at compile time
var _ Repository = (*Database)(nil)
