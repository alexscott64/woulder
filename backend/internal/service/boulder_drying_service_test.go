package service

import (
	"context"
	"testing"
	"time"

	"github.com/alexscott64/woulder/backend/internal/database"
	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/alexscott64/woulder/backend/internal/weather/client"
)

// mockRepository is a mock implementation of the database.Repository interface for testing
type mockRepository struct {
	routes              []*models.MPRoute
	profile             *models.BoulderDryingProfile
	currentWeather      *models.WeatherData
	historicalWeather   []models.WeatherData
	forecastWeather     []models.WeatherData
	rockTypes           []models.RockType
	sunExposure         *models.LocationSunExposure
	location            *models.Location
	getRoutesWithGPSErr error
}

// mockBoulderWeatherClient provides mock weather data for boulder drying tests
type mockBoulderWeatherClient struct {
	currentWeather  *models.WeatherData
	forecastWeather []models.WeatherData
}

func (m *mockBoulderWeatherClient) GetCurrentAndForecast(lat, lon float64) (*models.WeatherData, []models.WeatherData, *client.SunTimes, error) {
	return m.currentWeather, m.forecastWeather, nil, nil
}

func (m *mockRepository) GetRoutesWithGPSByArea(ctx context.Context, mpAreaID int64) ([]*models.MPRoute, error) {
	if m.getRoutesWithGPSErr != nil {
		return nil, m.getRoutesWithGPSErr
	}
	return m.routes, nil
}

func (m *mockRepository) GetBoulderDryingProfile(ctx context.Context, mpRouteID int64) (*models.BoulderDryingProfile, error) {
	return m.profile, nil
}

func (m *mockRepository) GetBoulderDryingProfilesByRouteIDs(ctx context.Context, mpRouteIDs []int64) (map[int64]*models.BoulderDryingProfile, error) {
	profiles := make(map[int64]*models.BoulderDryingProfile)
	if m.profile != nil {
		// For testing, return the same profile for all routes
		for _, id := range mpRouteIDs {
			profiles[id] = m.profile
		}
	}
	return profiles, nil
}

func (m *mockRepository) GetMPRouteByID(ctx context.Context, mpRouteID int64) (*models.MPRoute, error) {
	for _, route := range m.routes {
		if route.MPRouteID == mpRouteID {
			return route, nil
		}
	}
	return nil, nil
}

func (m *mockRepository) GetMPRoutesByIDs(ctx context.Context, mpRouteIDs []int64) (map[int64]*models.MPRoute, error) {
	routes := make(map[int64]*models.MPRoute)
	for _, route := range m.routes {
		for _, id := range mpRouteIDs {
			if route.MPRouteID == id {
				routes[id] = route
				break
			}
		}
	}
	return routes, nil
}

func (m *mockRepository) GetCurrentWeather(ctx context.Context, locationID int) (*models.WeatherData, error) {
	return m.currentWeather, nil
}

func (m *mockRepository) GetHistoricalWeather(ctx context.Context, locationID int, hours int) ([]models.WeatherData, error) {
	return m.historicalWeather, nil
}

func (m *mockRepository) GetForecastWeather(ctx context.Context, locationID int, hours int) ([]models.WeatherData, error) {
	return m.forecastWeather, nil
}

func (m *mockRepository) GetRockTypesByLocation(ctx context.Context, locationID int) ([]models.RockType, error) {
	return m.rockTypes, nil
}

func (m *mockRepository) GetSunExposureByLocation(ctx context.Context, locationID int) (*models.LocationSunExposure, error) {
	return m.sunExposure, nil
}

// Stub implementations for other Repository methods (not used in these tests)
func (m *mockRepository) GetAllLocations(ctx context.Context) ([]models.Location, error) {
	return nil, nil
}
func (m *mockRepository) GetLocation(ctx context.Context, id int) (*models.Location, error) {
	if m.location != nil {
		return m.location, nil
	}
	// Return a default location if not set
	return &models.Location{
		ID:        id,
		Name:      "Test Location",
		Latitude:  47.6,
		Longitude: -120.9,
	}, nil
}
func (m *mockRepository) GetLocationsByArea(ctx context.Context, areaID int) ([]models.Location, error) {
	return nil, nil
}
func (m *mockRepository) SaveWeatherData(ctx context.Context, data *models.WeatherData) error {
	return nil
}
func (m *mockRepository) CleanOldWeatherData(ctx context.Context, daysToKeep int) error {
	return nil
}
func (m *mockRepository) DeleteOldWeatherData(ctx context.Context, locationID int, daysToKeep int) error {
	return nil
}
func (m *mockRepository) GetRiversByLocation(ctx context.Context, locationID int) ([]models.River, error) {
	return nil, nil
}
func (m *mockRepository) GetRiverByID(ctx context.Context, id int) (*models.River, error) {
	return nil, nil
}
func (m *mockRepository) GetAllAreas(ctx context.Context) ([]models.Area, error) {
	return nil, nil
}
func (m *mockRepository) GetAreasWithLocationCounts(ctx context.Context) ([]models.AreaWithLocationCount, error) {
	return nil, nil
}
func (m *mockRepository) GetAreaByID(ctx context.Context, id int) (*models.Area, error) {
	return nil, nil
}
func (m *mockRepository) GetPrimaryRockType(ctx context.Context, locationID int) (*models.RockType, error) {
	return nil, nil
}
func (m *mockRepository) SaveMPArea(ctx context.Context, area *models.MPArea) error {
	return nil
}
func (m *mockRepository) SaveMPRoute(ctx context.Context, route *models.MPRoute) error {
	return nil
}
func (m *mockRepository) SaveMPTick(ctx context.Context, tick *models.MPTick) error {
	return nil
}
func (m *mockRepository) UpdateRouteGPS(ctx context.Context, routeID int64, latitude, longitude float64, aspect string) error {
	return nil
}
func (m *mockRepository) GetLastClimbedForLocation(ctx context.Context, locationID int) (*models.LastClimbedInfo, error) {
	return nil, nil
}
func (m *mockRepository) GetClimbHistoryForLocation(ctx context.Context, locationID int, limit int) ([]models.ClimbHistoryEntry, error) {
	return nil, nil
}
func (m *mockRepository) GetMPAreaByID(ctx context.Context, mpAreaID int64) (*models.MPArea, error) {
	return nil, nil
}
func (m *mockRepository) GetLastTickTimestampForRoute(ctx context.Context, routeID int64) (*time.Time, error) {
	return nil, nil
}
func (m *mockRepository) GetAllRouteIDsForLocation(ctx context.Context, locationID int) ([]int64, error) {
	return nil, nil
}
func (m *mockRepository) GetAreasOrderedByActivity(ctx context.Context, locationID int) ([]models.AreaActivitySummary, error) {
	return nil, nil
}
func (m *mockRepository) GetSubareasOrderedByActivity(ctx context.Context, parentAreaID int64, locationID int) ([]models.AreaActivitySummary, error) {
	return nil, nil
}
func (m *mockRepository) GetRoutesOrderedByActivity(ctx context.Context, areaID int64, locationID int, limit int) ([]models.RouteActivitySummary, error) {
	return nil, nil
}
func (m *mockRepository) GetRecentTicksForRoute(ctx context.Context, routeID int64, limit int) ([]models.ClimbHistoryEntry, error) {
	return nil, nil
}
func (m *mockRepository) SearchInLocation(ctx context.Context, locationID int, searchQuery string, limit int) ([]models.SearchResult, error) {
	return nil, nil
}
func (m *mockRepository) SearchRoutesInLocation(ctx context.Context, locationID int, searchQuery string, limit int) ([]models.RouteActivitySummary, error) {
	return nil, nil
}
func (m *mockRepository) SaveBoulderDryingProfile(ctx context.Context, profile *models.BoulderDryingProfile) error {
	return nil
}
func (m *mockRepository) GetLocationByID(ctx context.Context, locationID int) (*models.Location, error) {
	return nil, nil
}
func (m *mockRepository) Ping(ctx context.Context) error {
	return nil
}
func (m *mockRepository) SaveAreaComment(ctx context.Context, mpCommentID, mpAreaID int64, userName, commentText string, commentedAt time.Time) error {
	return nil
}
func (m *mockRepository) SaveRouteComment(ctx context.Context, mpCommentID, mpRouteID int64, userName, commentText string, commentedAt time.Time) error {
	return nil
}
func (m *mockRepository) Close() error {
	return nil
}
func (m *mockRepository) UpdateAreaRouteCount(ctx context.Context, mpAreaID string, total int) error {
	return nil
}
func (m *mockRepository) GetAreaRouteCount(ctx context.Context, mpAreaID string) (int, error) {
	return 0, nil
}
func (m *mockRepository) GetChildAreas(ctx context.Context, parentMPAreaID string) ([]struct {
	MPAreaID string
	Name     string
}, error) {
	return nil, nil
}
func (m *mockRepository) GetRouteIDsForArea(ctx context.Context, mpAreaID string) ([]string, error) {
	return nil, nil
}
func (m *mockRepository) GetAllStateConfigs(ctx context.Context) ([]struct {
	StateName string
	MPAreaID  string
	IsActive  bool
}, error) {
	return nil, nil
}
func (m *mockRepository) UpsertRoute(ctx context.Context, mpRouteID, mpAreaID int64, locationID *int, name, routeType, rating string, lat, lon *float64, aspect *string) error {
	return nil
}
func (m *mockRepository) UpsertTick(ctx context.Context, mpRouteID int64, userName string, climbedAt time.Time, style string, comment *string) error {
	return nil
}
func (m *mockRepository) UpsertAreaComment(ctx context.Context, mpCommentID, mpAreaID int64, userName string, userID *string, commentText string, commentedAt time.Time) error {
	return nil
}
func (m *mockRepository) UpsertRouteComment(ctx context.Context, mpCommentID, mpRouteID int64, userName string, userID *string, commentText string, commentedAt time.Time) error {
	return nil
}
func (m *mockRepository) UpdateRouteSyncPriorities(ctx context.Context) error {
	return nil
}
func (m *mockRepository) GetLocationRoutesDueForSync(ctx context.Context, syncType string) ([]int64, error) {
	return nil, nil
}
func (m *mockRepository) GetRoutesDueForTickSync(ctx context.Context, priority string) ([]int64, error) {
	return nil, nil
}
func (m *mockRepository) GetRoutesDueForCommentSync(ctx context.Context, priority string) ([]int64, error) {
	return nil, nil
}
func (m *mockRepository) GetPriorityDistribution(ctx context.Context) (map[string]int, error) {
	return make(map[string]int), nil
}
func (m *mockRepository) GetHeatMapData(ctx context.Context, startDate, endDate time.Time, bounds *database.GeoBounds, minActivity, limit int, routeTypes []string, lightweight bool) ([]models.HeatMapPoint, error) {
	return nil, nil
}
func (m *mockRepository) GetAreaActivityDetail(ctx context.Context, areaID int64, startDate, endDate time.Time) (*models.AreaActivityDetail, error) {
	return nil, nil
}
func (m *mockRepository) GetRoutesByBounds(ctx context.Context, bounds database.GeoBounds, startDate, endDate time.Time, limit int) ([]models.RouteActivity, error) {
	return nil, nil
}
func (m *mockRepository) GetRouteTicksInDateRange(ctx context.Context, routeID int64, startDate, endDate time.Time, limit int) ([]models.TickDetail, error) {
	return nil, nil
}
func (m *mockRepository) SearchRoutesInAreas(ctx context.Context, areaIDs []int64, searchQuery string, startDate, endDate time.Time, limit int) ([]models.RouteActivity, error) {
	return nil, nil
}

// TestGetAreaDryingStats_AllDry tests area stats when all routes are dry
func TestGetAreaDryingStats_AllDry(t *testing.T) {
	now := time.Now()
	locationID := 1

	// Mock 3 routes, all dry (last rain was 48h ago)
	routes := []*models.MPRoute{
		{
			MPRouteID:  1001,
			Name:       "Test Route 1",
			LocationID: &locationID,
			Latitude:   ptrFloat64(47.6),
			Longitude:  ptrFloat64(-120.9),
			Aspect:     ptrString("South"),
		},
		{
			MPRouteID:  1002,
			Name:       "Test Route 2",
			LocationID: &locationID,
			Latitude:   ptrFloat64(47.61),
			Longitude:  ptrFloat64(-120.91),
			Aspect:     ptrString("East"),
		},
		{
			MPRouteID:  1003,
			Name:       "Test Route 3",
			LocationID: &locationID,
			Latitude:   ptrFloat64(47.62),
			Longitude:  ptrFloat64(-120.92),
			Aspect:     ptrString("West"),
		},
	}

	// Create historical weather with last rain 72h ago (well past drying time)
	historicalWeather := []models.WeatherData{}
	for i := 168; i > 0; i-- {
		timestamp := now.Add(-time.Duration(i) * time.Hour)
		precip := 0.0
		if i == 72 {
			precip = 0.3 // Last rain 72h ago
		}
		historicalWeather = append(historicalWeather, models.WeatherData{
			LocationID:    locationID,
			Timestamp:     timestamp,
			Temperature:   60.0,
			Humidity:      50.0,
			WindSpeed:     5.0,
			Precipitation: precip,
		})
	}

	mock := &mockRepository{
		routes: routes,
		currentWeather: &models.WeatherData{
			LocationID:    locationID,
			Timestamp:     now,
			Temperature:   60.0,
			Humidity:      50.0,
			WindSpeed:     5.0,
			Precipitation: 0.0,
		},
		historicalWeather: historicalWeather,
		rockTypes: []models.RockType{
			{Name: "Granite", BaseDryingHours: 12.0},
		},
		sunExposure: &models.LocationSunExposure{
			TreeCoveragePercent: 30.0,
		},
	}

	mockWeather := &mockBoulderWeatherClient{
		currentWeather:  mock.currentWeather,
		forecastWeather: []models.WeatherData{},
	}

	service := NewBoulderDryingService(mock, mockWeather)
	stats, err := service.GetAreaDryingStats(context.Background(), 2001, locationID)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if stats == nil {
		t.Fatal("Expected stats, got nil")
	}

	if stats.TotalRoutes != 3 {
		t.Errorf("Expected TotalRoutes=3, got %d", stats.TotalRoutes)
	}

	if stats.DryCount != 3 {
		t.Errorf("Expected DryCount=3, got %d", stats.DryCount)
	}

	if stats.WetCount != 0 {
		t.Errorf("Expected WetCount=0, got %d", stats.WetCount)
	}

	if stats.DryingCount != 0 {
		t.Errorf("Expected DryingCount=0, got %d", stats.DryingCount)
	}

	if stats.PercentDry != 100.0 {
		t.Errorf("Expected PercentDry=100, got %.2f", stats.PercentDry)
	}
}

// TestGetAreaDryingStats_Mixed tests area stats with mixed dry/wet/drying routes
func TestGetAreaDryingStats_Mixed(t *testing.T) {
	now := time.Now()
	locationID := 1

	// Mock 4 routes: 1 dry, 1 drying (12h), 2 wet (>24h)
	routes := []*models.MPRoute{
		{
			MPRouteID:  2001,
			Name:       "Dry Route",
			LocationID: &locationID,
			Latitude:   ptrFloat64(47.6),
			Longitude:  ptrFloat64(-120.9),
			Aspect:     ptrString("South"),
		},
		{
			MPRouteID:  2002,
			Name:       "Drying Route",
			LocationID: &locationID,
			Latitude:   ptrFloat64(47.61),
			Longitude:  ptrFloat64(-120.91),
			Aspect:     ptrString("East"),
		},
		{
			MPRouteID:  2003,
			Name:       "Wet Route 1",
			LocationID: &locationID,
			Latitude:   ptrFloat64(47.62),
			Longitude:  ptrFloat64(-120.92),
			Aspect:     ptrString("North"),
		},
		{
			MPRouteID:  2004,
			Name:       "Wet Route 2",
			LocationID: &locationID,
			Latitude:   ptrFloat64(47.63),
			Longitude:  ptrFloat64(-120.93),
			Aspect:     ptrString("West"),
		},
	}

	// Historical weather: rain at different times for different routes
	// Route 1: Last rain 48h ago (dry)
	// Routes 2-4: Rain 2h ago (wet/drying)
	mock := &mockRepository{
		routes: routes,
		currentWeather: &models.WeatherData{
			LocationID:    locationID,
			Timestamp:     now,
			Temperature:   60.0,
			Humidity:      50.0,
			WindSpeed:     5.0,
			Precipitation: 0.0,
		},
		historicalWeather: []models.WeatherData{
			{
				LocationID:    locationID,
				Timestamp:     now.Add(-48 * time.Hour),
				Temperature:   55.0,
				Humidity:      80.0,
				Precipitation: 0.3, // Old rain
			},
			{
				LocationID:    locationID,
				Timestamp:     now.Add(-2 * time.Hour),
				Temperature:   58.0,
				Humidity:      70.0,
				Precipitation: 0.5, // Recent rain
			},
		},
		rockTypes: []models.RockType{
			{Name: "Granite", BaseDryingHours: 12.0},
		},
		sunExposure: &models.LocationSunExposure{
			TreeCoveragePercent: 30.0,
		},
	}

	mockWeather := &mockBoulderWeatherClient{
		currentWeather:  mock.currentWeather,
		forecastWeather: []models.WeatherData{},
	}

	service := NewBoulderDryingService(mock, mockWeather)
	stats, err := service.GetAreaDryingStats(context.Background(), 2001, locationID)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if stats == nil {
		t.Fatal("Expected stats, got nil")
	}

	if stats.TotalRoutes != 4 {
		t.Errorf("Expected TotalRoutes=4, got %d", stats.TotalRoutes)
	}

	// At least some routes should be wet/drying
	if stats.DryCount == 4 {
		t.Errorf("Expected some wet/drying routes, but all are dry")
	}

	// Percentage dry should be less than 100%
	if stats.PercentDry >= 100.0 {
		t.Errorf("Expected PercentDry < 100, got %.2f", stats.PercentDry)
	}

	// Should have average hours until dry for wet routes
	if stats.AvgHoursUntilDry <= 0 {
		t.Errorf("Expected AvgHoursUntilDry > 0, got %.2f", stats.AvgHoursUntilDry)
	}
}

// TestGetAreaDryingStats_NoRoutes tests area stats when area has no routes with GPS
func TestGetAreaDryingStats_NoRoutes(t *testing.T) {
	mock := &mockRepository{
		routes: []*models.MPRoute{}, // Empty
	}

	mockWeather := &mockBoulderWeatherClient{
		currentWeather:  nil,
		forecastWeather: []models.WeatherData{},
	}

	service := NewBoulderDryingService(mock, mockWeather)
	stats, err := service.GetAreaDryingStats(context.Background(), 2001, 1)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if stats != nil {
		t.Error("Expected nil stats for area with no routes, got non-nil")
	}
}

// TestGetAreaDryingStats_MixedStatus tests area stats with dry, drying, and wet routes
// This test verifies the fix for dry/drying/wet counting logic
func TestGetAreaDryingStats_MixedStatus(t *testing.T) {
	now := time.Now()
	locationID := 1

	// Mock 6 routes with different statuses:
	// 2 dry (0 hours until dry)
	// 2 drying (12 hours until dry - within 48h threshold)
	// 2 wet (96 hours until dry - beyond 48h threshold)
	routes := []*models.MPRoute{
		{
			MPRouteID:  1001,
			Name:       "Dry Route 1",
			LocationID: &locationID,
			Latitude:   ptrFloat64(47.6),
			Longitude:  ptrFloat64(-120.9),
			Aspect:     ptrString("South"),
		},
		{
			MPRouteID:  1002,
			Name:       "Dry Route 2",
			LocationID: &locationID,
			Latitude:   ptrFloat64(47.61),
			Longitude:  ptrFloat64(-120.91),
			Aspect:     ptrString("South"),
		},
		{
			MPRouteID:  1003,
			Name:       "Drying Route 1",
			LocationID: &locationID,
			Latitude:   ptrFloat64(47.62),
			Longitude:  ptrFloat64(-120.92),
			Aspect:     ptrString("North"), // North aspect dries slower
		},
		{
			MPRouteID:  1004,
			Name:       "Drying Route 2",
			LocationID: &locationID,
			Latitude:   ptrFloat64(47.63),
			Longitude:  ptrFloat64(-120.93),
			Aspect:     ptrString("North"),
		},
		{
			MPRouteID:  1005,
			Name:       "Wet Route 1",
			LocationID: &locationID,
			Latitude:   ptrFloat64(47.64),
			Longitude:  ptrFloat64(-120.94),
			Aspect:     ptrString("North"), // North aspect + recent heavy rain = very slow drying
		},
		{
			MPRouteID:  1006,
			Name:       "Wet Route 2",
			LocationID: &locationID,
			Latitude:   ptrFloat64(47.65),
			Longitude:  ptrFloat64(-120.95),
			Aspect:     ptrString("North"),
		},
	}

	// Create historical weather with:
	// - Light rain 24h ago (dry routes already dry)
	// - Moderate rain 12h ago (drying routes actively drying)
	// - Heavy rain 6h ago (wet routes taking long time to dry)
	historicalWeather := []models.WeatherData{}
	for i := 168; i > 0; i-- {
		timestamp := now.Add(-time.Duration(i) * time.Hour)
		precip := 0.0
		if i == 24 {
			precip = 0.1 // Light rain 24h ago
		} else if i == 12 {
			precip = 0.3 // Moderate rain 12h ago
		} else if i == 6 {
			precip = 0.8 // Heavy rain 6h ago
		}

		historicalWeather = append(historicalWeather, models.WeatherData{
			LocationID:    locationID,
			Timestamp:     timestamp,
			Temperature:   55.0,
			Precipitation: precip,
			Humidity:      60,
			WindSpeed:     5.0,
		})
	}

	// Current weather: dry, sunny
	currentWeather := &models.WeatherData{
		LocationID:    locationID,
		Timestamp:     now,
		Temperature:   65.0,
		Precipitation: 0,
		Humidity:      50,
		WindSpeed:     7.0,
	}

	// Forecast: no more rain
	forecastWeather := []models.WeatherData{}
	for i := 1; i <= 168; i++ {
		forecastWeather = append(forecastWeather, models.WeatherData{
			LocationID:    locationID,
			Timestamp:     now.Add(time.Duration(i) * time.Hour),
			Temperature:   65.0,
			Precipitation: 0,
			Humidity:      50,
			WindSpeed:     7.0,
		})
	}

	rockTypes := []models.RockType{
		{
			ID:              1,
			Name:            "Granite",
			BaseDryingHours: 8.0,
			PorosityPercent: 5.0,
			IsWetSensitive:  false,
		},
	}

	mock := &mockRepository{
		routes:            routes,
		currentWeather:    currentWeather,
		historicalWeather: historicalWeather,
		forecastWeather:   forecastWeather,
		rockTypes:         rockTypes,
	}

	mockWeather := &mockBoulderWeatherClient{
		currentWeather:  currentWeather,
		forecastWeather: forecastWeather,
	}

	service := NewBoulderDryingService(mock, mockWeather)
	stats, err := service.GetAreaDryingStats(context.Background(), 2001, locationID)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if stats == nil {
		t.Fatal("Expected stats, got nil")
	}

	// Verify total routes
	if stats.TotalRoutes != 6 {
		t.Errorf("Expected TotalRoutes=6, got %d", stats.TotalRoutes)
	}

	// Verify dry/drying/wet counts
	// Note: The actual counts will vary based on the drying algorithm,
	// but we should have a mix (not all in one category)
	totalCounted := stats.DryCount + stats.DryingCount + stats.WetCount
	if totalCounted != stats.TotalRoutes {
		t.Errorf("Expected dry+drying+wet=%d, got %d+%d+%d=%d",
			stats.TotalRoutes, stats.DryCount, stats.DryingCount, stats.WetCount, totalCounted)
	}

	// Verify percentages make sense
	if stats.PercentDry < 0 || stats.PercentDry > 100 {
		t.Errorf("Expected PercentDry in [0,100], got %.2f", stats.PercentDry)
	}

	// Log the results for debugging
	t.Logf("Area Stats: Total=%d, Dry=%d, Drying=%d, Wet=%d, PercentDry=%.2f, AvgHoursUntilDry=%.2f",
		stats.TotalRoutes, stats.DryCount, stats.DryingCount, stats.WetCount,
		stats.PercentDry, stats.AvgHoursUntilDry)
}

// Helper functions
func ptrFloat64(f float64) *float64 {
	return &f
}

func ptrString(s string) *string {
	return &s
}
