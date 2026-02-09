package database

import (
	"context"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// MockRepository implements Repository interface for testing
type MockRepository struct {
	// Location mocks
	GetAllLocationsFn    func(ctx context.Context) ([]models.Location, error)
	GetLocationFn        func(ctx context.Context, id int) (*models.Location, error)
	GetLocationsByAreaFn func(ctx context.Context, areaID int) ([]models.Location, error)

	// Weather mocks
	SaveWeatherDataFn        func(ctx context.Context, data *models.WeatherData) error
	GetHistoricalWeatherFn   func(ctx context.Context, locationID int, hours int) ([]models.WeatherData, error)
	GetForecastWeatherFn     func(ctx context.Context, locationID int, hours int) ([]models.WeatherData, error)
	GetCurrentWeatherFn      func(ctx context.Context, locationID int) (*models.WeatherData, error)
	CleanOldWeatherDataFn    func(ctx context.Context, daysToKeep int) error
	DeleteOldWeatherDataFn   func(ctx context.Context, locationID int, daysToKeep int) error

	// River mocks
	GetRiversByLocationFn func(ctx context.Context, locationID int) ([]models.River, error)
	GetRiverByIDFn        func(ctx context.Context, id int) (*models.River, error)

	// Area mocks
	GetAllAreasFn                func(ctx context.Context) ([]models.Area, error)
	GetAreasWithLocationCountsFn func(ctx context.Context) ([]models.AreaWithLocationCount, error)
	GetAreaByIDFn                func(ctx context.Context, id int) (*models.Area, error)

	// Rock type mocks
	GetRockTypesByLocationFn   func(ctx context.Context, locationID int) ([]models.RockType, error)
	GetPrimaryRockTypeFn       func(ctx context.Context, locationID int) (*models.RockType, error)
	GetSunExposureByLocationFn func(ctx context.Context, locationID int) (*models.LocationSunExposure, error)

	// Mountain Project mocks
	SaveMPAreaFn                       func(ctx context.Context, area *models.MPArea) error
	SaveMPRouteFn                      func(ctx context.Context, route *models.MPRoute) error
	SaveMPTickFn                       func(ctx context.Context, tick *models.MPTick) error
	SaveAreaCommentFn                  func(ctx context.Context, mpCommentID, mpAreaID int64, userName, commentText string, commentedAt time.Time) error
	SaveRouteCommentFn                 func(ctx context.Context, mpCommentID, mpRouteID int64, userName, commentText string, commentedAt time.Time) error
	UpdateRouteGPSFn                   func(ctx context.Context, routeID int64, latitude, longitude float64, aspect string) error
	GetLastClimbedForLocationFn        func(ctx context.Context, locationID int) (*models.LastClimbedInfo, error)
	GetClimbHistoryForLocationFn       func(ctx context.Context, locationID int, limit int) ([]models.ClimbHistoryEntry, error)
	GetMPAreaByIDFn                    func(ctx context.Context, mpAreaID int64) (*models.MPArea, error)
	GetMPRouteByIDFn                   func(ctx context.Context, mpRouteID int64) (*models.MPRoute, error)
	GetLastTickTimestampForRouteFn     func(ctx context.Context, routeID int64) (*time.Time, error)
	GetAllRouteIDsForLocationFn        func(ctx context.Context, locationID int) ([]int64, error)
	GetAreasOrderedByActivityFn        func(ctx context.Context, locationID int) ([]models.AreaActivitySummary, error)
	GetSubareasOrderedByActivityFn     func(ctx context.Context, parentAreaID int64, locationID int) ([]models.AreaActivitySummary, error)
	GetRoutesOrderedByActivityFn       func(ctx context.Context, areaID int64, locationID int, limit int) ([]models.RouteActivitySummary, error)
	GetRecentTicksForRouteFn           func(ctx context.Context, routeID int64, limit int) ([]models.ClimbHistoryEntry, error)
	SearchInLocationFn                 func(ctx context.Context, locationID int, searchQuery string, limit int) ([]models.SearchResult, error)
	SearchRoutesInLocationFn           func(ctx context.Context, locationID int, searchQuery string, limit int) ([]models.RouteActivitySummary, error)

	// Route count tracking mocks
	UpdateAreaRouteCountFn func(ctx context.Context, mpAreaID string, total int) error
	GetAreaRouteCountFn    func(ctx context.Context, mpAreaID string) (int, error)
	GetChildAreasFn        func(ctx context.Context, parentMPAreaID string) ([]struct {
		MPAreaID string
		Name     string
	}, error)
	GetRouteIDsForAreaFn   func(ctx context.Context, mpAreaID string) ([]string, error)
	GetAllStateConfigsFn   func(ctx context.Context) ([]struct {
		StateName string
		MPAreaID  string
		IsActive  bool
	}, error)

	// Priority-based sync mocks
	UpdateRouteSyncPrioritiesFn      func(ctx context.Context) error
	GetLocationRoutesDueForSyncFn    func(ctx context.Context, syncType string) ([]int64, error)
	GetRoutesDueForTickSyncFn        func(ctx context.Context, priority string) ([]int64, error)
	GetRoutesDueForCommentSyncFn     func(ctx context.Context, priority string) ([]int64, error)
	GetPriorityDistributionFn        func(ctx context.Context) (map[string]int, error)

	// Upsert operations mocks
	UpsertRouteFn        func(ctx context.Context, mpRouteID, mpAreaID int64, locationID *int, name, routeType, rating string, lat, lon *float64, aspect *string) error
	UpsertTickFn         func(ctx context.Context, mpRouteID int64, userName string, climbedAt time.Time, style string, comment *string) error
	UpsertAreaCommentFn  func(ctx context.Context, mpCommentID, mpAreaID int64, userName string, userID *string, commentText string, commentedAt time.Time) error
	UpsertRouteCommentFn func(ctx context.Context, mpCommentID, mpRouteID int64, userName string, userID *string, commentText string, commentedAt time.Time) error

	// Boulder drying mocks
	GetBoulderDryingProfileFn          func(ctx context.Context, mpRouteID int64) (*models.BoulderDryingProfile, error)
	GetBoulderDryingProfilesByRouteIDsFn func(ctx context.Context, mpRouteIDs []int64) (map[int64]*models.BoulderDryingProfile, error)
	SaveBoulderDryingProfileFn         func(ctx context.Context, profile *models.BoulderDryingProfile) error
	GetLocationByIDFn                  func(ctx context.Context, locationID int) (*models.Location, error)
	GetRoutesWithGPSByAreaFn           func(ctx context.Context, mpAreaID int64) ([]*models.MPRoute, error)
	GetMPRoutesByIDsFn                 func(ctx context.Context, mpRouteIDs []int64) (map[int64]*models.MPRoute, error)

	// Health
	PingFn  func(ctx context.Context) error
	CloseFn func() error
}

// GetAllLocations mock
func (m *MockRepository) GetAllLocations(ctx context.Context) ([]models.Location, error) {
	if m.GetAllLocationsFn != nil {
		return m.GetAllLocationsFn(ctx)
	}
	return nil, nil
}

// GetLocation mock
func (m *MockRepository) GetLocation(ctx context.Context, id int) (*models.Location, error) {
	if m.GetLocationFn != nil {
		return m.GetLocationFn(ctx, id)
	}
	return nil, nil
}

func (m *MockRepository) GetLocationsByArea(ctx context.Context, areaID int) ([]models.Location, error) {
	if m.GetLocationsByAreaFn != nil {
		return m.GetLocationsByAreaFn(ctx, areaID)
	}
	return nil, nil
}

// SaveWeatherData mock
func (m *MockRepository) SaveWeatherData(ctx context.Context, data *models.WeatherData) error {
	if m.SaveWeatherDataFn != nil {
		return m.SaveWeatherDataFn(ctx, data)
	}
	return nil
}

// GetHistoricalWeather mock
func (m *MockRepository) GetHistoricalWeather(ctx context.Context, locationID int, hours int) ([]models.WeatherData, error) {
	if m.GetHistoricalWeatherFn != nil {
		return m.GetHistoricalWeatherFn(ctx, locationID, hours)
	}
	return nil, nil
}

// GetForecastWeather mock
func (m *MockRepository) GetForecastWeather(ctx context.Context, locationID int, hours int) ([]models.WeatherData, error) {
	if m.GetForecastWeatherFn != nil {
		return m.GetForecastWeatherFn(ctx, locationID, hours)
	}
	return nil, nil
}

// GetCurrentWeather mock
func (m *MockRepository) GetCurrentWeather(ctx context.Context, locationID int) (*models.WeatherData, error) {
	if m.GetCurrentWeatherFn != nil {
		return m.GetCurrentWeatherFn(ctx, locationID)
	}
	return nil, nil
}

// CleanOldWeatherData mock
func (m *MockRepository) CleanOldWeatherData(ctx context.Context, daysToKeep int) error {
	if m.CleanOldWeatherDataFn != nil {
		return m.CleanOldWeatherDataFn(ctx, daysToKeep)
	}
	return nil
}

// DeleteOldWeatherData mock
func (m *MockRepository) DeleteOldWeatherData(ctx context.Context, locationID int, daysToKeep int) error {
	if m.DeleteOldWeatherDataFn != nil {
		return m.DeleteOldWeatherDataFn(ctx, locationID, daysToKeep)
	}
	return nil
}

// GetRiversByLocation mock
func (m *MockRepository) GetRiversByLocation(ctx context.Context, locationID int) ([]models.River, error) {
	if m.GetRiversByLocationFn != nil {
		return m.GetRiversByLocationFn(ctx, locationID)
	}
	return nil, nil
}

// GetRiverByID mock
func (m *MockRepository) GetRiverByID(ctx context.Context, id int) (*models.River, error) {
	if m.GetRiverByIDFn != nil {
		return m.GetRiverByIDFn(ctx, id)
	}
	return nil, nil
}

// GetAllAreas mock
func (m *MockRepository) GetAllAreas(ctx context.Context) ([]models.Area, error) {
	if m.GetAllAreasFn != nil {
		return m.GetAllAreasFn(ctx)
	}
	return nil, nil
}

// GetAreasWithLocationCounts mock
func (m *MockRepository) GetAreasWithLocationCounts(ctx context.Context) ([]models.AreaWithLocationCount, error) {
	if m.GetAreasWithLocationCountsFn != nil {
		return m.GetAreasWithLocationCountsFn(ctx)
	}
	return nil, nil
}

// GetAreaByID mock
func (m *MockRepository) GetAreaByID(ctx context.Context, id int) (*models.Area, error) {
	if m.GetAreaByIDFn != nil {
		return m.GetAreaByIDFn(ctx, id)
	}
	return nil, nil
}

// GetRockTypesByLocation mock
func (m *MockRepository) GetRockTypesByLocation(ctx context.Context, locationID int) ([]models.RockType, error) {
	if m.GetRockTypesByLocationFn != nil {
		return m.GetRockTypesByLocationFn(ctx, locationID)
	}
	return nil, nil
}

// GetPrimaryRockType mock
func (m *MockRepository) GetPrimaryRockType(ctx context.Context, locationID int) (*models.RockType, error) {
	if m.GetPrimaryRockTypeFn != nil {
		return m.GetPrimaryRockTypeFn(ctx, locationID)
	}
	return nil, nil
}

// GetSunExposureByLocation mock
func (m *MockRepository) GetSunExposureByLocation(ctx context.Context, locationID int) (*models.LocationSunExposure, error) {
	if m.GetSunExposureByLocationFn != nil {
		return m.GetSunExposureByLocationFn(ctx, locationID)
	}
	return nil, nil
}

// Ping mock
func (m *MockRepository) Ping(ctx context.Context) error {
	if m.PingFn != nil {
		return m.PingFn(ctx)
	}
	return nil
}

// Close mock
func (m *MockRepository) Close() error {
	if m.CloseFn != nil {
		return m.CloseFn()
	}
	return nil
}

// SaveMPArea mock
func (m *MockRepository) SaveMPArea(ctx context.Context, area *models.MPArea) error {
	if m.SaveMPAreaFn != nil {
		return m.SaveMPAreaFn(ctx, area)
	}
	return nil
}

// SaveMPRoute mock
func (m *MockRepository) SaveMPRoute(ctx context.Context, route *models.MPRoute) error {
	if m.SaveMPRouteFn != nil {
		return m.SaveMPRouteFn(ctx, route)
	}
	return nil
}

// SaveMPTick mock
func (m *MockRepository) SaveMPTick(ctx context.Context, tick *models.MPTick) error {
	if m.SaveMPTickFn != nil {
		return m.SaveMPTickFn(ctx, tick)
	}
	return nil
}

// SaveAreaComment mock
func (m *MockRepository) SaveAreaComment(ctx context.Context, mpCommentID, mpAreaID int64, userName, commentText string, commentedAt time.Time) error {
	if m.SaveAreaCommentFn != nil {
		return m.SaveAreaCommentFn(ctx, mpCommentID, mpAreaID, userName, commentText, commentedAt)
	}
	return nil
}

// SaveRouteComment mock
func (m *MockRepository) SaveRouteComment(ctx context.Context, mpCommentID, mpRouteID int64, userName, commentText string, commentedAt time.Time) error {
	if m.SaveRouteCommentFn != nil {
		return m.SaveRouteCommentFn(ctx, mpCommentID, mpRouteID, userName, commentText, commentedAt)
	}
	return nil
}

// GetLastClimbedForLocation mock
func (m *MockRepository) GetLastClimbedForLocation(ctx context.Context, locationID int) (*models.LastClimbedInfo, error) {
	if m.GetLastClimbedForLocationFn != nil {
		return m.GetLastClimbedForLocationFn(ctx, locationID)
	}
	return nil, nil
}

// GetClimbHistoryForLocation mock
func (m *MockRepository) GetClimbHistoryForLocation(ctx context.Context, locationID int, limit int) ([]models.ClimbHistoryEntry, error) {
	if m.GetClimbHistoryForLocationFn != nil {
		return m.GetClimbHistoryForLocationFn(ctx, locationID, limit)
	}
	return nil, nil
}

// GetMPAreaByID mock
func (m *MockRepository) GetMPAreaByID(ctx context.Context, mpAreaID int64) (*models.MPArea, error) {
	if m.GetMPAreaByIDFn != nil {
		return m.GetMPAreaByIDFn(ctx, mpAreaID)
	}
	return nil, nil
}

// GetLastTickTimestampForRoute mock
func (m *MockRepository) GetLastTickTimestampForRoute(ctx context.Context, routeID int64) (*time.Time, error) {
	if m.GetLastTickTimestampForRouteFn != nil {
		return m.GetLastTickTimestampForRouteFn(ctx, routeID)
	}
	return nil, nil
}

// GetAllRouteIDsForLocation mock
func (m *MockRepository) GetAllRouteIDsForLocation(ctx context.Context, locationID int) ([]int64, error) {
	if m.GetAllRouteIDsForLocationFn != nil {
		return m.GetAllRouteIDsForLocationFn(ctx, locationID)
	}
	return nil, nil
}

// GetAreasOrderedByActivity mock
func (m *MockRepository) GetAreasOrderedByActivity(ctx context.Context, locationID int) ([]models.AreaActivitySummary, error) {
	if m.GetAreasOrderedByActivityFn != nil {
		return m.GetAreasOrderedByActivityFn(ctx, locationID)
	}
	return nil, nil
}

// GetSubareasOrderedByActivity mock
func (m *MockRepository) GetSubareasOrderedByActivity(ctx context.Context, parentAreaID int64, locationID int) ([]models.AreaActivitySummary, error) {
	if m.GetSubareasOrderedByActivityFn != nil {
		return m.GetSubareasOrderedByActivityFn(ctx, parentAreaID, locationID)
	}
	return nil, nil
}

// GetRoutesOrderedByActivity mock
func (m *MockRepository) GetRoutesOrderedByActivity(ctx context.Context, areaID int64, locationID int, limit int) ([]models.RouteActivitySummary, error) {
	if m.GetRoutesOrderedByActivityFn != nil {
		return m.GetRoutesOrderedByActivityFn(ctx, areaID, locationID, limit)
	}
	return nil, nil
}

// GetRecentTicksForRoute mock
func (m *MockRepository) GetRecentTicksForRoute(ctx context.Context, routeID int64, limit int) ([]models.ClimbHistoryEntry, error) {
	if m.GetRecentTicksForRouteFn != nil {
		return m.GetRecentTicksForRouteFn(ctx, routeID, limit)
	}
	return nil, nil
}

// SearchInLocation mock
func (m *MockRepository) SearchInLocation(ctx context.Context, locationID int, searchQuery string, limit int) ([]models.SearchResult, error) {
	if m.SearchInLocationFn != nil {
		return m.SearchInLocationFn(ctx, locationID, searchQuery, limit)
	}
	return nil, nil
}

// SearchRoutesInLocation mock
func (m *MockRepository) SearchRoutesInLocation(ctx context.Context, locationID int, searchQuery string, limit int) ([]models.RouteActivitySummary, error) {
	if m.SearchRoutesInLocationFn != nil {
		return m.SearchRoutesInLocationFn(ctx, locationID, searchQuery, limit)
	}
	return nil, nil
}

// UpdateRouteGPS mock
func (m *MockRepository) UpdateRouteGPS(ctx context.Context, routeID int64, latitude, longitude float64, aspect string) error {
	if m.UpdateRouteGPSFn != nil {
		return m.UpdateRouteGPSFn(ctx, routeID, latitude, longitude, aspect)
	}
	return nil
}

// GetMPRouteByID mock
func (m *MockRepository) GetMPRouteByID(ctx context.Context, mpRouteID int64) (*models.MPRoute, error) {
	if m.GetMPRouteByIDFn != nil {
		return m.GetMPRouteByIDFn(ctx, mpRouteID)
	}
	return nil, nil
}

// GetMPRoutesByIDs mock
func (m *MockRepository) GetMPRoutesByIDs(ctx context.Context, mpRouteIDs []int64) (map[int64]*models.MPRoute, error) {
	if m.GetMPRoutesByIDsFn != nil {
		return m.GetMPRoutesByIDsFn(ctx, mpRouteIDs)
	}
	return make(map[int64]*models.MPRoute), nil
}

// GetBoulderDryingProfile mock
func (m *MockRepository) GetBoulderDryingProfile(ctx context.Context, mpRouteID int64) (*models.BoulderDryingProfile, error) {
	if m.GetBoulderDryingProfileFn != nil {
		return m.GetBoulderDryingProfileFn(ctx, mpRouteID)
	}
	return nil, nil
}

// GetBoulderDryingProfilesByRouteIDs mock
func (m *MockRepository) GetBoulderDryingProfilesByRouteIDs(ctx context.Context, mpRouteIDs []int64) (map[int64]*models.BoulderDryingProfile, error) {
	if m.GetBoulderDryingProfilesByRouteIDsFn != nil {
		return m.GetBoulderDryingProfilesByRouteIDsFn(ctx, mpRouteIDs)
	}
	return make(map[int64]*models.BoulderDryingProfile), nil
}

// SaveBoulderDryingProfile mock
func (m *MockRepository) SaveBoulderDryingProfile(ctx context.Context, profile *models.BoulderDryingProfile) error {
	if m.SaveBoulderDryingProfileFn != nil {
		return m.SaveBoulderDryingProfileFn(ctx, profile)
	}
	return nil
}

// GetLocationByID mock
func (m *MockRepository) GetLocationByID(ctx context.Context, locationID int) (*models.Location, error) {
	if m.GetLocationByIDFn != nil {
		return m.GetLocationByIDFn(ctx, locationID)
	}
	return nil, nil
}

// GetRoutesWithGPSByArea mock
func (m *MockRepository) GetRoutesWithGPSByArea(ctx context.Context, mpAreaID int64) ([]*models.MPRoute, error) {
	if m.GetRoutesWithGPSByAreaFn != nil {
		return m.GetRoutesWithGPSByAreaFn(ctx, mpAreaID)
	}
	return nil, nil
}

// UpdateAreaRouteCount mock
func (m *MockRepository) UpdateAreaRouteCount(ctx context.Context, mpAreaID string, total int) error {
	if m.UpdateAreaRouteCountFn != nil {
		return m.UpdateAreaRouteCountFn(ctx, mpAreaID, total)
	}
	return nil
}

// GetAreaRouteCount mock
func (m *MockRepository) GetAreaRouteCount(ctx context.Context, mpAreaID string) (int, error) {
	if m.GetAreaRouteCountFn != nil {
		return m.GetAreaRouteCountFn(ctx, mpAreaID)
	}
	return -1, nil
}

// GetChildAreas mock
func (m *MockRepository) GetChildAreas(ctx context.Context, parentMPAreaID string) ([]struct {
	MPAreaID string
	Name     string
}, error) {
	if m.GetChildAreasFn != nil {
		return m.GetChildAreasFn(ctx, parentMPAreaID)
	}
	return nil, nil
}

// GetRouteIDsForArea mock
func (m *MockRepository) GetRouteIDsForArea(ctx context.Context, mpAreaID string) ([]string, error) {
	if m.GetRouteIDsForAreaFn != nil {
		return m.GetRouteIDsForAreaFn(ctx, mpAreaID)
	}
	return nil, nil
}

// GetAllStateConfigs mock
func (m *MockRepository) GetAllStateConfigs(ctx context.Context) ([]struct {
	StateName string
	MPAreaID  string
	IsActive  bool
}, error) {
	if m.GetAllStateConfigsFn != nil {
		return m.GetAllStateConfigsFn(ctx)
	}
	return nil, nil
}

// UpsertRoute mock
func (m *MockRepository) UpsertRoute(ctx context.Context, mpRouteID, mpAreaID int64, locationID *int, name, routeType, rating string, lat, lon *float64, aspect *string) error {
	if m.UpsertRouteFn != nil {
		return m.UpsertRouteFn(ctx, mpRouteID, mpAreaID, locationID, name, routeType, rating, lat, lon, aspect)
	}
	return nil
}

// UpsertTick mock
func (m *MockRepository) UpsertTick(ctx context.Context, mpRouteID int64, userName string, climbedAt time.Time, style string, comment *string) error {
	if m.UpsertTickFn != nil {
		return m.UpsertTickFn(ctx, mpRouteID, userName, climbedAt, style, comment)
	}
	return nil
}

// UpsertAreaComment mock
func (m *MockRepository) UpsertAreaComment(ctx context.Context, mpCommentID, mpAreaID int64, userName string, userID *string, commentText string, commentedAt time.Time) error {
	if m.UpsertAreaCommentFn != nil {
		return m.UpsertAreaCommentFn(ctx, mpCommentID, mpAreaID, userName, userID, commentText, commentedAt)
	}
	return nil
}

// UpsertRouteComment mock
func (m *MockRepository) UpsertRouteComment(ctx context.Context, mpCommentID, mpRouteID int64, userName string, userID *string, commentText string, commentedAt time.Time) error {
	if m.UpsertRouteCommentFn != nil {
		return m.UpsertRouteCommentFn(ctx, mpCommentID, mpRouteID, userName, userID, commentText, commentedAt)
	}
	return nil
}

// UpdateRouteSyncPriorities mock
func (m *MockRepository) UpdateRouteSyncPriorities(ctx context.Context) error {
	if m.UpdateRouteSyncPrioritiesFn != nil {
		return m.UpdateRouteSyncPrioritiesFn(ctx)
	}
	return nil
}

// GetLocationRoutesDueForSync mock
func (m *MockRepository) GetLocationRoutesDueForSync(ctx context.Context, syncType string) ([]int64, error) {
	if m.GetLocationRoutesDueForSyncFn != nil {
		return m.GetLocationRoutesDueForSyncFn(ctx, syncType)
	}
	return nil, nil
}

// GetRoutesDueForTickSync mock
func (m *MockRepository) GetRoutesDueForTickSync(ctx context.Context, priority string) ([]int64, error) {
	if m.GetRoutesDueForTickSyncFn != nil {
		return m.GetRoutesDueForTickSyncFn(ctx, priority)
	}
	return nil, nil
}

// GetRoutesDueForCommentSync mock
func (m *MockRepository) GetRoutesDueForCommentSync(ctx context.Context, priority string) ([]int64, error) {
	if m.GetRoutesDueForCommentSyncFn != nil {
		return m.GetRoutesDueForCommentSyncFn(ctx, priority)
	}
	return nil, nil
}

// GetPriorityDistribution mock
func (m *MockRepository) GetPriorityDistribution(ctx context.Context) (map[string]int, error) {
	if m.GetPriorityDistributionFn != nil {
		return m.GetPriorityDistributionFn(ctx)
	}
	return make(map[string]int), nil
}
