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
	SaveAreaComment(ctx context.Context, mpCommentID, mpAreaID int64, userName, commentText string, commentedAt time.Time) error
	SaveRouteComment(ctx context.Context, mpCommentID, mpRouteID int64, userName, commentText string, commentedAt time.Time) error
	UpdateRouteGPS(ctx context.Context, routeID int64, latitude, longitude float64, aspect string) error
	GetLastClimbedForLocation(ctx context.Context, locationID int) (*models.LastClimbedInfo, error) // DEPRECATED: Use GetClimbHistoryForLocation
	GetClimbHistoryForLocation(ctx context.Context, locationID int, limit int) ([]models.ClimbHistoryEntry, error)
	GetMPAreaByID(ctx context.Context, mpAreaID int64) (*models.MPArea, error)
	GetMPRouteByID(ctx context.Context, mpRouteID int64) (*models.MPRoute, error)
	GetLastTickTimestampForRoute(ctx context.Context, routeID int64) (*time.Time, error)
	GetAllRouteIDsForLocation(ctx context.Context, locationID int) ([]int64, error)
	GetAreasOrderedByActivity(ctx context.Context, locationID int) ([]models.AreaActivitySummary, error)
	GetSubareasOrderedByActivity(ctx context.Context, parentAreaID int64, locationID int) ([]models.AreaActivitySummary, error)
	GetRoutesOrderedByActivity(ctx context.Context, areaID int64, locationID int, limit int) ([]models.RouteActivitySummary, error)
	GetRecentTicksForRoute(ctx context.Context, routeID int64, limit int) ([]models.ClimbHistoryEntry, error)
	SearchInLocation(ctx context.Context, locationID int, searchQuery string, limit int) ([]models.SearchResult, error)
	SearchRoutesInLocation(ctx context.Context, locationID int, searchQuery string, limit int) ([]models.RouteActivitySummary, error)

	// Route count tracking for new route detection
	UpdateAreaRouteCount(ctx context.Context, mpAreaID string, total int) error
	GetAreaRouteCount(ctx context.Context, mpAreaID string) (int, error)
	GetChildAreas(ctx context.Context, parentMPAreaID string) ([]struct {
		MPAreaID string
		Name     string
	}, error)
	GetRouteIDsForArea(ctx context.Context, mpAreaID string) ([]string, error)
	GetAllStateConfigs(ctx context.Context) ([]struct {
		StateName string
		MPAreaID  string
		IsActive  bool
	}, error)

	// Upsert operations (used by mountainprojectsync)
	UpsertRoute(ctx context.Context, mpRouteID, mpAreaID int64, locationID *int, name, routeType, rating string, lat, lon *float64, aspect *string) error
	UpsertTick(ctx context.Context, mpRouteID int64, userName string, climbedAt time.Time, style string, comment *string) error
	UpsertAreaComment(ctx context.Context, mpCommentID, mpAreaID int64, userName string, userID *string, commentText string, commentedAt time.Time) error
	UpsertRouteComment(ctx context.Context, mpCommentID, mpRouteID int64, userName string, userID *string, commentText string, commentedAt time.Time) error

	// Priority-based sync operations for tick/comment optimization
	UpdateRouteSyncPriorities(ctx context.Context) error
	GetLocationRoutesDueForSync(ctx context.Context, syncType string) ([]int64, error)
	GetRoutesDueForTickSync(ctx context.Context, priority string) ([]int64, error)
	GetRoutesDueForCommentSync(ctx context.Context, priority string) ([]int64, error)
	GetPriorityDistribution(ctx context.Context) (map[string]int, error)

	// Boulder drying operations
	GetBoulderDryingProfile(ctx context.Context, mpRouteID int64) (*models.BoulderDryingProfile, error)
	GetBoulderDryingProfilesByRouteIDs(ctx context.Context, mpRouteIDs []int64) (map[int64]*models.BoulderDryingProfile, error)
	SaveBoulderDryingProfile(ctx context.Context, profile *models.BoulderDryingProfile) error
	GetLocationByID(ctx context.Context, locationID int) (*models.Location, error)
	GetRoutesWithGPSByArea(ctx context.Context, mpAreaID int64) ([]*models.MPRoute, error)
	GetMPRoutesByIDs(ctx context.Context, mpRouteIDs []int64) (map[int64]*models.MPRoute, error)

	// Heat map operations
	GetHeatMapData(ctx context.Context, startDate, endDate time.Time, bounds *GeoBounds, minActivity, limit int, routeTypes []string, lightweight bool) ([]models.HeatMapPoint, error)
	GetAreaActivityDetail(ctx context.Context, areaID int64, startDate, endDate time.Time) (*models.AreaActivityDetail, error)
	GetRoutesByBounds(ctx context.Context, bounds GeoBounds, startDate, endDate time.Time, limit int) ([]models.RouteActivity, error)

	// Health check
	Ping(ctx context.Context) error

	// Cleanup
	Close() error
}

// GeoBounds represents a geographic bounding box for filtering
type GeoBounds struct {
	MinLat float64
	MaxLat float64
	MinLon float64
	MaxLon float64
}

// Ensure Database implements Repository interface at compile time
var _ Repository = (*Database)(nil)
