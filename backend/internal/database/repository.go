package database

import (
	"context"
	"time"

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
	DeleteOldWeatherData(ctx context.Context, locationID int, daysToKeep int) error

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

	// Mountain Project operations
	SaveMPArea(ctx context.Context, area *models.MPArea) error
	SaveMPRoute(ctx context.Context, route *models.MPRoute) error
	SaveMPTick(ctx context.Context, tick *models.MPTick) error
	GetLastClimbedForLocation(ctx context.Context, locationID int) (*models.LastClimbedInfo, error) // DEPRECATED: Use GetClimbHistoryForLocation
	GetClimbHistoryForLocation(ctx context.Context, locationID int, limit int) ([]models.ClimbHistoryEntry, error)
	GetMPAreaByID(ctx context.Context, mpAreaID string) (*models.MPArea, error)
	GetLastTickTimestampForRoute(ctx context.Context, routeID string) (*time.Time, error)
	GetAllRouteIDsForLocation(ctx context.Context, locationID int) ([]string, error)
	GetAreasOrderedByActivity(ctx context.Context, locationID int) ([]models.AreaActivitySummary, error)
	GetSubareasOrderedByActivity(ctx context.Context, parentAreaID string, locationID int) ([]models.AreaActivitySummary, error)
	GetRoutesOrderedByActivity(ctx context.Context, areaID string, locationID int, limit int) ([]models.RouteActivitySummary, error)
	GetRecentTicksForRoute(ctx context.Context, routeID string, limit int) ([]models.ClimbHistoryEntry, error)
	SearchInLocation(ctx context.Context, locationID int, searchQuery string, limit int) ([]models.SearchResult, error)
	SearchRoutesInLocation(ctx context.Context, locationID int, searchQuery string, limit int) ([]models.RouteActivitySummary, error)

	// Health check
	Ping(ctx context.Context) error

	// Cleanup
	Close() error
}

// Ensure Database implements Repository interface at compile time
var _ Repository = (*Database)(nil)
