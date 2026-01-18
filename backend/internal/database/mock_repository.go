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
	SaveWeatherDataFn      func(ctx context.Context, data *models.WeatherData) error
	GetHistoricalWeatherFn func(ctx context.Context, locationID int, hours int) ([]models.WeatherData, error)
	GetForecastWeatherFn   func(ctx context.Context, locationID int, hours int) ([]models.WeatherData, error)
	GetCurrentWeatherFn    func(ctx context.Context, locationID int) (*models.WeatherData, error)
	CleanOldWeatherDataFn  func(ctx context.Context, daysToKeep int) error

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
	GetLastClimbedForLocationFn        func(ctx context.Context, locationID int) (*models.LastClimbedInfo, error)
	GetClimbHistoryForLocationFn       func(ctx context.Context, locationID int, limit int) ([]models.ClimbHistoryEntry, error)
	GetMPAreaByIDFn                    func(ctx context.Context, mpAreaID string) (*models.MPArea, error)
	GetLastTickTimestampForRouteFn     func(ctx context.Context, routeID string) (*time.Time, error)
	GetAllRouteIDsForLocationFn        func(ctx context.Context, locationID int) ([]string, error)
	GetAreasOrderedByActivityFn        func(ctx context.Context, locationID int) ([]models.AreaActivitySummary, error)
	GetSubareasOrderedByActivityFn     func(ctx context.Context, parentAreaID string, locationID int) ([]models.AreaActivitySummary, error)
	GetRoutesOrderedByActivityFn       func(ctx context.Context, areaID string, locationID int, limit int) ([]models.RouteActivitySummary, error)
	GetRecentTicksForRouteFn           func(ctx context.Context, routeID string, limit int) ([]models.ClimbHistoryEntry, error)
	SearchInLocationFn                 func(ctx context.Context, locationID int, searchQuery string, limit int) ([]models.SearchResult, error)
	SearchRoutesInLocationFn           func(ctx context.Context, locationID int, searchQuery string, limit int) ([]models.RouteActivitySummary, error)

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
func (m *MockRepository) GetMPAreaByID(ctx context.Context, mpAreaID string) (*models.MPArea, error) {
	if m.GetMPAreaByIDFn != nil {
		return m.GetMPAreaByIDFn(ctx, mpAreaID)
	}
	return nil, nil
}

// GetLastTickTimestampForRoute mock
func (m *MockRepository) GetLastTickTimestampForRoute(ctx context.Context, routeID string) (*time.Time, error) {
	if m.GetLastTickTimestampForRouteFn != nil {
		return m.GetLastTickTimestampForRouteFn(ctx, routeID)
	}
	return nil, nil
}

// GetAllRouteIDsForLocation mock
func (m *MockRepository) GetAllRouteIDsForLocation(ctx context.Context, locationID int) ([]string, error) {
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
func (m *MockRepository) GetSubareasOrderedByActivity(ctx context.Context, parentAreaID string, locationID int) ([]models.AreaActivitySummary, error) {
	if m.GetSubareasOrderedByActivityFn != nil {
		return m.GetSubareasOrderedByActivityFn(ctx, parentAreaID, locationID)
	}
	return nil, nil
}

// GetRoutesOrderedByActivity mock
func (m *MockRepository) GetRoutesOrderedByActivity(ctx context.Context, areaID string, locationID int, limit int) ([]models.RouteActivitySummary, error) {
	if m.GetRoutesOrderedByActivityFn != nil {
		return m.GetRoutesOrderedByActivityFn(ctx, areaID, locationID, limit)
	}
	return nil, nil
}

// GetRecentTicksForRoute mock
func (m *MockRepository) GetRecentTicksForRoute(ctx context.Context, routeID string, limit int) ([]models.ClimbHistoryEntry, error) {
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
