package service

import (
	"context"
	"time"

	"github.com/alexscott64/woulder/backend/internal/database/climbing"
	"github.com/alexscott64/woulder/backend/internal/database/heatmap"
	"github.com/alexscott64/woulder/backend/internal/database/mountainproject"
	"github.com/alexscott64/woulder/backend/internal/models"
)

// This file contains shared mock implementations used across service tests.
// Mocks are grouped by domain (weather, locations, climbing, etc.) for clarity.
//
// DESIGN DECISION: We use a single shared mocks file rather than per-test mocks because:
// 1. Reduces file proliferation - more maintainable
// 2. Follows Go testing conventions (single _test.go package with shared test helpers)
// 3. Mocks are simple and reusable across tests
// 4. Still keeps mocks isolated from production code (in _test.go files)

// ============================================================================
// WEATHER REPOSITORY MOCKS
// ============================================================================

// MockWeatherRepository implements weather.Repository
type MockWeatherRepository struct {
	SaveFn                 func(ctx context.Context, data *models.WeatherData) error
	GetHistoricalFn        func(ctx context.Context, locationID int, days int) ([]models.WeatherData, error)
	GetForecastFn          func(ctx context.Context, locationID int, hours int) ([]models.WeatherData, error)
	GetCurrentFn           func(ctx context.Context, locationID int) (*models.WeatherData, error)
	CleanOldFn             func(ctx context.Context, daysToKeep int) error
	DeleteOldForLocationFn func(ctx context.Context, locationID int, daysToKeep int) error
}

func (m *MockWeatherRepository) Save(ctx context.Context, data *models.WeatherData) error {
	if m.SaveFn != nil {
		return m.SaveFn(ctx, data)
	}
	return nil
}

func (m *MockWeatherRepository) GetHistorical(ctx context.Context, locationID int, days int) ([]models.WeatherData, error) {
	if m.GetHistoricalFn != nil {
		return m.GetHistoricalFn(ctx, locationID, days)
	}
	return []models.WeatherData{}, nil
}

func (m *MockWeatherRepository) GetForecast(ctx context.Context, locationID int, hours int) ([]models.WeatherData, error) {
	if m.GetForecastFn != nil {
		return m.GetForecastFn(ctx, locationID, hours)
	}
	return []models.WeatherData{}, nil
}

func (m *MockWeatherRepository) GetCurrent(ctx context.Context, locationID int) (*models.WeatherData, error) {
	if m.GetCurrentFn != nil {
		return m.GetCurrentFn(ctx, locationID)
	}
	return nil, nil
}

func (m *MockWeatherRepository) CleanOld(ctx context.Context, daysToKeep int) error {
	if m.CleanOldFn != nil {
		return m.CleanOldFn(ctx, daysToKeep)
	}
	return nil
}

func (m *MockWeatherRepository) DeleteOldForLocation(ctx context.Context, locationID int, daysToKeep int) error {
	if m.DeleteOldForLocationFn != nil {
		return m.DeleteOldForLocationFn(ctx, locationID, daysToKeep)
	}
	return nil
}

// ============================================================================
// LOCATIONS REPOSITORY MOCKS
// ============================================================================

// MockLocationsRepository implements locations.Repository
type MockLocationsRepository struct {
	GetAllFn    func(ctx context.Context) ([]models.Location, error)
	GetByIDFn   func(ctx context.Context, id int) (*models.Location, error)
	GetByAreaFn func(ctx context.Context, areaID int) ([]models.Location, error)
}

func (m *MockLocationsRepository) GetAll(ctx context.Context) ([]models.Location, error) {
	if m.GetAllFn != nil {
		return m.GetAllFn(ctx)
	}
	return []models.Location{}, nil
}

func (m *MockLocationsRepository) GetByID(ctx context.Context, id int) (*models.Location, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *MockLocationsRepository) GetByArea(ctx context.Context, areaID int) ([]models.Location, error) {
	if m.GetByAreaFn != nil {
		return m.GetByAreaFn(ctx, areaID)
	}
	return []models.Location{}, nil
}

// ============================================================================
// ROCKS REPOSITORY MOCKS
// ============================================================================

// MockRocksRepository implements rocks.Repository
type MockRocksRepository struct {
	GetRockTypesByLocationFn   func(ctx context.Context, locationID int) ([]models.RockType, error)
	GetPrimaryRockTypeFn       func(ctx context.Context, locationID int) (*models.RockType, error)
	GetSunExposureByLocationFn func(ctx context.Context, locationID int) (*models.LocationSunExposure, error)
}

func (m *MockRocksRepository) GetRockTypesByLocation(ctx context.Context, locationID int) ([]models.RockType, error) {
	if m.GetRockTypesByLocationFn != nil {
		return m.GetRockTypesByLocationFn(ctx, locationID)
	}
	return []models.RockType{}, nil
}

func (m *MockRocksRepository) GetPrimaryRockType(ctx context.Context, locationID int) (*models.RockType, error) {
	if m.GetPrimaryRockTypeFn != nil {
		return m.GetPrimaryRockTypeFn(ctx, locationID)
	}
	return nil, nil
}

func (m *MockRocksRepository) GetSunExposureByLocation(ctx context.Context, locationID int) (*models.LocationSunExposure, error) {
	if m.GetSunExposureByLocationFn != nil {
		return m.GetSunExposureByLocationFn(ctx, locationID)
	}
	return nil, nil
}

// ============================================================================
// RIVERS REPOSITORY MOCKS
// ============================================================================

// MockRiversRepository implements rivers.Repository
type MockRiversRepository struct {
	GetByLocationFn func(ctx context.Context, locationID int) ([]models.River, error)
	GetByIDFn       func(ctx context.Context, id int) (*models.River, error)
}

func (m *MockRiversRepository) GetByLocation(ctx context.Context, locationID int) ([]models.River, error) {
	if m.GetByLocationFn != nil {
		return m.GetByLocationFn(ctx, locationID)
	}
	return []models.River{}, nil
}

func (m *MockRiversRepository) GetByID(ctx context.Context, id int) (*models.River, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

// ============================================================================
// AREAS REPOSITORY MOCKS
// ============================================================================

// MockAreasRepository implements areas.Repository
type MockAreasRepository struct {
	GetAllFn                   func(ctx context.Context) ([]models.Area, error)
	GetByIDFn                  func(ctx context.Context, id int) (*models.Area, error)
	GetAllWithLocationCountsFn func(ctx context.Context) ([]models.AreaWithLocationCount, error)
}

func (m *MockAreasRepository) GetAll(ctx context.Context) ([]models.Area, error) {
	if m.GetAllFn != nil {
		return m.GetAllFn(ctx)
	}
	return []models.Area{}, nil
}

func (m *MockAreasRepository) GetByID(ctx context.Context, id int) (*models.Area, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *MockAreasRepository) GetAllWithLocationCounts(ctx context.Context) ([]models.AreaWithLocationCount, error) {
	if m.GetAllWithLocationCountsFn != nil {
		return m.GetAllWithLocationCountsFn(ctx)
	}
	return []models.AreaWithLocationCount{}, nil
}

// ============================================================================
// MOUNTAIN PROJECT REPOSITORY MOCKS
// ============================================================================

// MockMountainProjectRepository implements mountainproject.Repository with sub-repositories
type MockMountainProjectRepository struct {
	areas    *MockMPAreasRepository
	routes   *MockMPRoutesRepository
	ticks    *MockMPTicksRepository
	comments *MockMPCommentsRepository
	sync     *MockMPSyncRepository
}

func NewMockMountainProjectRepository() *MockMountainProjectRepository {
	return &MockMountainProjectRepository{
		areas:    &MockMPAreasRepository{},
		routes:   &MockMPRoutesRepository{},
		ticks:    &MockMPTicksRepository{},
		comments: &MockMPCommentsRepository{},
		sync:     &MockMPSyncRepository{},
	}
}

func (m *MockMountainProjectRepository) Areas() mountainproject.AreasRepository {
	return m.areas
}

func (m *MockMountainProjectRepository) Routes() mountainproject.RoutesRepository {
	return m.routes
}

func (m *MockMountainProjectRepository) Ticks() mountainproject.TicksRepository {
	return m.ticks
}

func (m *MockMountainProjectRepository) Comments() mountainproject.CommentsRepository {
	return m.comments
}

func (m *MockMountainProjectRepository) Sync() mountainproject.SyncRepository {
	return m.sync
}

// MockMPAreasRepository implements mountainproject.AreasRepository
type MockMPAreasRepository struct {
	SaveAreaFn           func(ctx context.Context, area *models.MPArea) error
	GetAreaByIDFn        func(ctx context.Context, mpAreaID int64) (*models.MPArea, error)
	UpdateRouteCountFn   func(ctx context.Context, mpAreaID string, total int) error
	GetRouteCountFn      func(ctx context.Context, mpAreaID string) (int, error)
	GetAllStateConfigsFn func(ctx context.Context) ([]mountainproject.StateConfig, error)
	GetChildAreasFn      func(ctx context.Context, parentMPAreaID string) ([]mountainproject.ChildArea, error)
}

func (m *MockMPAreasRepository) SaveArea(ctx context.Context, area *models.MPArea) error {
	if m.SaveAreaFn != nil {
		return m.SaveAreaFn(ctx, area)
	}
	return nil
}

func (m *MockMPAreasRepository) GetAreaByID(ctx context.Context, mpAreaID int64) (*models.MPArea, error) {
	if m.GetAreaByIDFn != nil {
		return m.GetAreaByIDFn(ctx, mpAreaID)
	}
	return nil, nil
}

func (m *MockMPAreasRepository) UpdateRouteCount(ctx context.Context, mpAreaID string, total int) error {
	if m.UpdateRouteCountFn != nil {
		return m.UpdateRouteCountFn(ctx, mpAreaID, total)
	}
	return nil
}

func (m *MockMPAreasRepository) GetRouteCount(ctx context.Context, mpAreaID string) (int, error) {
	if m.GetRouteCountFn != nil {
		return m.GetRouteCountFn(ctx, mpAreaID)
	}
	return 0, nil
}

func (m *MockMPAreasRepository) GetAllStateConfigs(ctx context.Context) ([]mountainproject.StateConfig, error) {
	if m.GetAllStateConfigsFn != nil {
		return m.GetAllStateConfigsFn(ctx)
	}
	return []mountainproject.StateConfig{}, nil
}

func (m *MockMPAreasRepository) GetChildAreas(ctx context.Context, parentMPAreaID string) ([]mountainproject.ChildArea, error) {
	if m.GetChildAreasFn != nil {
		return m.GetChildAreasFn(ctx, parentMPAreaID)
	}
	return []mountainproject.ChildArea{}, nil
}

// MockMPRoutesRepository implements mountainproject.RoutesRepository
type MockMPRoutesRepository struct {
	SaveRouteFn            func(ctx context.Context, route *models.MPRoute) error
	GetByIDFn              func(ctx context.Context, mpRouteID int64) (*models.MPRoute, error)
	GetByIDsFn             func(ctx context.Context, mpRouteIDs []int64) (map[int64]*models.MPRoute, error)
	GetAllIDsForLocationFn func(ctx context.Context, locationID int) ([]int64, error)
	UpdateGPSFn            func(ctx context.Context, routeID int64, latitude, longitude float64, aspect string) error
	GetIDsForAreaFn        func(ctx context.Context, mpAreaID string) ([]string, error)
	GetWithGPSByAreaFn     func(ctx context.Context, mpAreaID int64) ([]*models.MPRoute, error)
	UpsertRouteFn          func(ctx context.Context, mpRouteID, mpAreaID int64, locationID *int, name, routeType, rating string, lat, lon *float64, aspect *string) error
}

func (m *MockMPRoutesRepository) SaveRoute(ctx context.Context, route *models.MPRoute) error {
	if m.SaveRouteFn != nil {
		return m.SaveRouteFn(ctx, route)
	}
	return nil
}

func (m *MockMPRoutesRepository) GetByID(ctx context.Context, mpRouteID int64) (*models.MPRoute, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, mpRouteID)
	}
	return nil, nil
}

func (m *MockMPRoutesRepository) GetByIDs(ctx context.Context, mpRouteIDs []int64) (map[int64]*models.MPRoute, error) {
	if m.GetByIDsFn != nil {
		return m.GetByIDsFn(ctx, mpRouteIDs)
	}
	return map[int64]*models.MPRoute{}, nil
}

func (m *MockMPRoutesRepository) GetAllIDsForLocation(ctx context.Context, locationID int) ([]int64, error) {
	if m.GetAllIDsForLocationFn != nil {
		return m.GetAllIDsForLocationFn(ctx, locationID)
	}
	return []int64{}, nil
}

func (m *MockMPRoutesRepository) UpdateGPS(ctx context.Context, routeID int64, latitude, longitude float64, aspect string) error {
	if m.UpdateGPSFn != nil {
		return m.UpdateGPSFn(ctx, routeID, latitude, longitude, aspect)
	}
	return nil
}

func (m *MockMPRoutesRepository) GetIDsForArea(ctx context.Context, mpAreaID string) ([]string, error) {
	if m.GetIDsForAreaFn != nil {
		return m.GetIDsForAreaFn(ctx, mpAreaID)
	}
	return []string{}, nil
}

func (m *MockMPRoutesRepository) GetWithGPSByArea(ctx context.Context, mpAreaID int64) ([]*models.MPRoute, error) {
	if m.GetWithGPSByAreaFn != nil {
		return m.GetWithGPSByAreaFn(ctx, mpAreaID)
	}
	return []*models.MPRoute{}, nil
}

func (m *MockMPRoutesRepository) UpsertRoute(ctx context.Context, mpRouteID, mpAreaID int64, locationID *int, name, routeType, rating string, lat, lon *float64, aspect *string) error {
	if m.UpsertRouteFn != nil {
		return m.UpsertRouteFn(ctx, mpRouteID, mpAreaID, locationID, name, routeType, rating, lat, lon, aspect)
	}
	return nil
}

// MockMPTicksRepository implements mountainproject.TicksRepository
type MockMPTicksRepository struct {
	SaveTickFn                 func(ctx context.Context, tick *models.MPTick) error
	GetLastTimestampForRouteFn func(ctx context.Context, routeID int64) (*time.Time, error)
	UpsertTickFn               func(ctx context.Context, mpRouteID int64, userName string, climbedAt time.Time, style string, comment *string) error
}

func (m *MockMPTicksRepository) SaveTick(ctx context.Context, tick *models.MPTick) error {
	if m.SaveTickFn != nil {
		return m.SaveTickFn(ctx, tick)
	}
	return nil
}

func (m *MockMPTicksRepository) GetLastTimestampForRoute(ctx context.Context, routeID int64) (*time.Time, error) {
	if m.GetLastTimestampForRouteFn != nil {
		return m.GetLastTimestampForRouteFn(ctx, routeID)
	}
	return nil, nil
}

func (m *MockMPTicksRepository) UpsertTick(ctx context.Context, mpRouteID int64, userName string, climbedAt time.Time, style string, comment *string) error {
	if m.UpsertTickFn != nil {
		return m.UpsertTickFn(ctx, mpRouteID, userName, climbedAt, style, comment)
	}
	return nil
}

// MockMPCommentsRepository implements mountainproject.CommentsRepository
type MockMPCommentsRepository struct {
	SaveAreaCommentFn    func(ctx context.Context, mpCommentID, mpAreaID int64, userName, commentText string, commentedAt time.Time) error
	SaveRouteCommentFn   func(ctx context.Context, mpCommentID, mpRouteID int64, userName, commentText string, commentedAt time.Time) error
	UpsertAreaCommentFn  func(ctx context.Context, mpCommentID, mpAreaID int64, userName string, userID *string, commentText string, commentedAt time.Time) error
	UpsertRouteCommentFn func(ctx context.Context, mpCommentID, mpRouteID int64, userName string, userID *string, commentText string, commentedAt time.Time) error
}

func (m *MockMPCommentsRepository) SaveAreaComment(ctx context.Context, mpCommentID, mpAreaID int64, userName, commentText string, commentedAt time.Time) error {
	if m.SaveAreaCommentFn != nil {
		return m.SaveAreaCommentFn(ctx, mpCommentID, mpAreaID, userName, commentText, commentedAt)
	}
	return nil
}

func (m *MockMPCommentsRepository) SaveRouteComment(ctx context.Context, mpCommentID, mpRouteID int64, userName, commentText string, commentedAt time.Time) error {
	if m.SaveRouteCommentFn != nil {
		return m.SaveRouteCommentFn(ctx, mpCommentID, mpRouteID, userName, commentText, commentedAt)
	}
	return nil
}

func (m *MockMPCommentsRepository) UpsertAreaComment(ctx context.Context, mpCommentID, mpAreaID int64, userName string, userID *string, commentText string, commentedAt time.Time) error {
	if m.UpsertAreaCommentFn != nil {
		return m.UpsertAreaCommentFn(ctx, mpCommentID, mpAreaID, userName, userID, commentText, commentedAt)
	}
	return nil
}

func (m *MockMPCommentsRepository) UpsertRouteComment(ctx context.Context, mpCommentID, mpRouteID int64, userName string, userID *string, commentText string, commentedAt time.Time) error {
	if m.UpsertRouteCommentFn != nil {
		return m.UpsertRouteCommentFn(ctx, mpCommentID, mpRouteID, userName, userID, commentText, commentedAt)
	}
	return nil
}

// MockMPSyncRepository implements mountainproject.SyncRepository
type MockMPSyncRepository struct {
	UpdateRoutePrioritiesFn       func(ctx context.Context) error
	GetLocationRoutesDueForSyncFn func(ctx context.Context, syncType string) ([]int64, error)
	GetRoutesDueForTickSyncFn     func(ctx context.Context, priority string) ([]int64, error)
	GetRoutesDueForCommentSyncFn  func(ctx context.Context, priority string) ([]int64, error)
	GetPriorityDistributionFn     func(ctx context.Context) (map[string]int, error)
}

func (m *MockMPSyncRepository) UpdateRoutePriorities(ctx context.Context) error {
	if m.UpdateRoutePrioritiesFn != nil {
		return m.UpdateRoutePrioritiesFn(ctx)
	}
	return nil
}

func (m *MockMPSyncRepository) GetLocationRoutesDueForSync(ctx context.Context, syncType string) ([]int64, error) {
	if m.GetLocationRoutesDueForSyncFn != nil {
		return m.GetLocationRoutesDueForSyncFn(ctx, syncType)
	}
	return []int64{}, nil
}

func (m *MockMPSyncRepository) GetRoutesDueForTickSync(ctx context.Context, priority string) ([]int64, error) {
	if m.GetRoutesDueForTickSyncFn != nil {
		return m.GetRoutesDueForTickSyncFn(ctx, priority)
	}
	return []int64{}, nil
}

func (m *MockMPSyncRepository) GetRoutesDueForCommentSync(ctx context.Context, priority string) ([]int64, error) {
	if m.GetRoutesDueForCommentSyncFn != nil {
		return m.GetRoutesDueForCommentSyncFn(ctx, priority)
	}
	return []int64{}, nil
}

func (m *MockMPSyncRepository) GetPriorityDistribution(ctx context.Context) (map[string]int, error) {
	if m.GetPriorityDistributionFn != nil {
		return m.GetPriorityDistributionFn(ctx)
	}
	return map[string]int{}, nil
}

// ============================================================================
// CLIMBING REPOSITORY MOCKS
// ============================================================================

// MockClimbingRepository implements climbing.Repository with sub-repositories
type MockClimbingRepository struct {
	history  *MockClimbingHistoryRepository
	activity *MockClimbingActivityRepository
	search   *MockClimbingSearchRepository
}

func NewMockClimbingRepository() *MockClimbingRepository {
	return &MockClimbingRepository{
		history:  &MockClimbingHistoryRepository{},
		activity: &MockClimbingActivityRepository{},
		search:   &MockClimbingSearchRepository{},
	}
}

func (m *MockClimbingRepository) History() climbing.HistoryRepository {
	return m.history
}

func (m *MockClimbingRepository) Activity() climbing.ActivityRepository {
	return m.activity
}

func (m *MockClimbingRepository) Search() climbing.SearchRepository {
	return m.search
}

// MockClimbingHistoryRepository provides climb history methods
type MockClimbingHistoryRepository struct {
	GetLastClimbedForLocationFn   func(ctx context.Context, locationID int) (*models.LastClimbedInfo, error)
	GetClimbHistoryForLocationFn  func(ctx context.Context, locationID int, limit int) ([]models.ClimbHistoryEntry, error)
	GetClimbHistoryForLocationsFn func(ctx context.Context, locationIDs []int, limit int) (map[int][]models.ClimbHistoryEntry, error)
}

func (m *MockClimbingHistoryRepository) GetLastClimbedForLocation(ctx context.Context, locationID int) (*models.LastClimbedInfo, error) {
	if m.GetLastClimbedForLocationFn != nil {
		return m.GetLastClimbedForLocationFn(ctx, locationID)
	}
	return nil, nil
}

func (m *MockClimbingHistoryRepository) GetClimbHistoryForLocation(ctx context.Context, locationID int, limit int) ([]models.ClimbHistoryEntry, error) {
	if m.GetClimbHistoryForLocationFn != nil {
		return m.GetClimbHistoryForLocationFn(ctx, locationID, limit)
	}
	return []models.ClimbHistoryEntry{}, nil
}

func (m *MockClimbingHistoryRepository) GetClimbHistoryForLocations(ctx context.Context, locationIDs []int, limit int) (map[int][]models.ClimbHistoryEntry, error) {
	if m.GetClimbHistoryForLocationsFn != nil {
		return m.GetClimbHistoryForLocationsFn(ctx, locationIDs, limit)
	}
	return map[int][]models.ClimbHistoryEntry{}, nil
}

// MockClimbingActivityRepository provides climbing activity methods
type MockClimbingActivityRepository struct {
	GetAreasOrderedByActivityFn    func(ctx context.Context, locationID int) ([]models.AreaActivitySummary, error)
	GetSubareasOrderedByActivityFn func(ctx context.Context, parentAreaID int64, locationID int) ([]models.AreaActivitySummary, error)
	GetRoutesOrderedByActivityFn   func(ctx context.Context, areaID int64, locationID int, limit int) ([]models.RouteActivitySummary, error)
	GetRecentTicksForRouteFn       func(ctx context.Context, routeID int64, limit int) ([]models.ClimbHistoryEntry, error)
}

func (m *MockClimbingActivityRepository) GetAreasOrderedByActivity(ctx context.Context, locationID int) ([]models.AreaActivitySummary, error) {
	if m.GetAreasOrderedByActivityFn != nil {
		return m.GetAreasOrderedByActivityFn(ctx, locationID)
	}
	return []models.AreaActivitySummary{}, nil
}

func (m *MockClimbingActivityRepository) GetSubareasOrderedByActivity(ctx context.Context, parentAreaID int64, locationID int) ([]models.AreaActivitySummary, error) {
	if m.GetSubareasOrderedByActivityFn != nil {
		return m.GetSubareasOrderedByActivityFn(ctx, parentAreaID, locationID)
	}
	return []models.AreaActivitySummary{}, nil
}

func (m *MockClimbingActivityRepository) GetRoutesOrderedByActivity(ctx context.Context, areaID int64, locationID int, limit int) ([]models.RouteActivitySummary, error) {
	if m.GetRoutesOrderedByActivityFn != nil {
		return m.GetRoutesOrderedByActivityFn(ctx, areaID, locationID, limit)
	}
	return []models.RouteActivitySummary{}, nil
}

func (m *MockClimbingActivityRepository) GetRecentTicksForRoute(ctx context.Context, routeID int64, limit int) ([]models.ClimbHistoryEntry, error) {
	if m.GetRecentTicksForRouteFn != nil {
		return m.GetRecentTicksForRouteFn(ctx, routeID, limit)
	}
	return []models.ClimbHistoryEntry{}, nil
}

// MockClimbingSearchRepository provides search methods
type MockClimbingSearchRepository struct {
	SearchInLocationFn       func(ctx context.Context, locationID int, searchQuery string, limit int) ([]models.SearchResult, error)
	SearchRoutesInLocationFn func(ctx context.Context, locationID int, searchQuery string, limit int) ([]models.RouteActivitySummary, error)
}

func (m *MockClimbingSearchRepository) SearchInLocation(ctx context.Context, locationID int, searchQuery string, limit int) ([]models.SearchResult, error) {
	if m.SearchInLocationFn != nil {
		return m.SearchInLocationFn(ctx, locationID, searchQuery, limit)
	}
	return []models.SearchResult{}, nil
}

func (m *MockClimbingSearchRepository) SearchRoutesInLocation(ctx context.Context, locationID int, searchQuery string, limit int) ([]models.RouteActivitySummary, error) {
	if m.SearchRoutesInLocationFn != nil {
		return m.SearchRoutesInLocationFn(ctx, locationID, searchQuery, limit)
	}
	return []models.RouteActivitySummary{}, nil
}

// ============================================================================
// BOULDERS REPOSITORY MOCKS
// ============================================================================

// MockBouldersRepository implements boulders.Repository
type MockBouldersRepository struct {
	GetProfileFn       func(ctx context.Context, mpRouteID int64) (*models.BoulderDryingProfile, error)
	GetProfilesByIDsFn func(ctx context.Context, mpRouteIDs []int64) (map[int64]*models.BoulderDryingProfile, error)
	SaveProfileFn      func(ctx context.Context, profile *models.BoulderDryingProfile) error
}

func (m *MockBouldersRepository) GetProfile(ctx context.Context, mpRouteID int64) (*models.BoulderDryingProfile, error) {
	if m.GetProfileFn != nil {
		return m.GetProfileFn(ctx, mpRouteID)
	}
	return nil, nil
}

func (m *MockBouldersRepository) GetProfilesByIDs(ctx context.Context, mpRouteIDs []int64) (map[int64]*models.BoulderDryingProfile, error) {
	if m.GetProfilesByIDsFn != nil {
		return m.GetProfilesByIDsFn(ctx, mpRouteIDs)
	}
	return map[int64]*models.BoulderDryingProfile{}, nil
}

func (m *MockBouldersRepository) SaveProfile(ctx context.Context, profile *models.BoulderDryingProfile) error {
	if m.SaveProfileFn != nil {
		return m.SaveProfileFn(ctx, profile)
	}
	return nil
}

// ============================================================================
// HEATMAP REPOSITORY MOCKS
// ============================================================================

// MockHeatMapRepository implements heatmap.Repository
type MockHeatMapRepository struct {
	GetHeatMapDataFn           func(ctx context.Context, startDate, endDate time.Time, bounds *heatmap.GeoBounds, minActivity, limit int, routeTypes []string, lightweight bool) ([]models.HeatMapPoint, error)
	GetAreaActivityDetailFn    func(ctx context.Context, areaID int64, startDate, endDate time.Time) (*models.AreaActivityDetail, error)
	GetRoutesByBoundsFn        func(ctx context.Context, bounds heatmap.GeoBounds, startDate, endDate time.Time, limit int) ([]models.RouteActivity, error)
	GetRouteTicksInDateRangeFn func(ctx context.Context, routeID int64, startDate, endDate time.Time, limit int) ([]models.TickDetail, error)
	SearchRoutesInAreasFn      func(ctx context.Context, areaIDs []int64, searchQuery string, startDate, endDate time.Time, limit int) ([]models.RouteActivity, error)
}

func (m *MockHeatMapRepository) GetHeatMapData(ctx context.Context, startDate, endDate time.Time, bounds *heatmap.GeoBounds, minActivity, limit int, routeTypes []string, lightweight bool) ([]models.HeatMapPoint, error) {
	if m.GetHeatMapDataFn != nil {
		return m.GetHeatMapDataFn(ctx, startDate, endDate, bounds, minActivity, limit, routeTypes, lightweight)
	}
	return []models.HeatMapPoint{}, nil
}

func (m *MockHeatMapRepository) GetAreaActivityDetail(ctx context.Context, areaID int64, startDate, endDate time.Time) (*models.AreaActivityDetail, error) {
	if m.GetAreaActivityDetailFn != nil {
		return m.GetAreaActivityDetailFn(ctx, areaID, startDate, endDate)
	}
	return nil, nil
}

func (m *MockHeatMapRepository) GetRoutesByBounds(ctx context.Context, bounds heatmap.GeoBounds, startDate, endDate time.Time, limit int) ([]models.RouteActivity, error) {
	if m.GetRoutesByBoundsFn != nil {
		return m.GetRoutesByBoundsFn(ctx, bounds, startDate, endDate, limit)
	}
	return []models.RouteActivity{}, nil
}

func (m *MockHeatMapRepository) GetRouteTicksInDateRange(ctx context.Context, routeID int64, startDate, endDate time.Time, limit int) ([]models.TickDetail, error) {
	if m.GetRouteTicksInDateRangeFn != nil {
		return m.GetRouteTicksInDateRangeFn(ctx, routeID, startDate, endDate, limit)
	}
	return []models.TickDetail{}, nil
}

func (m *MockHeatMapRepository) SearchRoutesInAreas(ctx context.Context, areaIDs []int64, searchQuery string, startDate, endDate time.Time, limit int) ([]models.RouteActivity, error) {
	if m.SearchRoutesInAreasFn != nil {
		return m.SearchRoutesInAreasFn(ctx, areaIDs, searchQuery, startDate, endDate, limit)
	}
	return []models.RouteActivity{}, nil
}

// MockClimbTrackingService implements the ClimbTrackingService interface for testing
type MockClimbTrackingService struct {
	GetClimbHistoryForLocationsFn func(ctx context.Context, locationIDs []int, limit int) (map[int][]models.ClimbHistoryEntry, error)
}

func (m *MockClimbTrackingService) GetClimbHistoryForLocations(ctx context.Context, locationIDs []int, limit int) (map[int][]models.ClimbHistoryEntry, error) {
	if m.GetClimbHistoryForLocationsFn != nil {
		return m.GetClimbHistoryForLocationsFn(ctx, locationIDs, limit)
	}
	return map[int][]models.ClimbHistoryEntry{}, nil
}
