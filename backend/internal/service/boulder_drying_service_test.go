package service

import (
	"context"
	"testing"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
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


func (m *mockRepository) GetRoutesWithGPSByArea(ctx context.Context, mpAreaID string) ([]*models.MPRoute, error) {
	if m.getRoutesWithGPSErr != nil {
		return nil, m.getRoutesWithGPSErr
	}
	return m.routes, nil
}

func (m *mockRepository) GetBoulderDryingProfile(ctx context.Context, mpRouteID string) (*models.BoulderDryingProfile, error) {
	return m.profile, nil
}

func (m *mockRepository) GetBoulderDryingProfilesByRouteIDs(ctx context.Context, mpRouteIDs []string) (map[string]*models.BoulderDryingProfile, error) {
	profiles := make(map[string]*models.BoulderDryingProfile)
	if m.profile != nil {
		// For testing, return the same profile for all routes
		for _, id := range mpRouteIDs {
			profiles[id] = m.profile
		}
	}
	return profiles, nil
}

func (m *mockRepository) GetMPRouteByID(ctx context.Context, mpRouteID string) (*models.MPRoute, error) {
	for _, route := range m.routes {
		if route.MPRouteID == mpRouteID {
			return route, nil
		}
	}
	return nil, nil
}

func (m *mockRepository) GetMPRoutesByIDs(ctx context.Context, mpRouteIDs []string) (map[string]*models.MPRoute, error) {
	routes := make(map[string]*models.MPRoute)
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
func (m *mockRepository) UpdateRouteGPS(ctx context.Context, routeID string, latitude, longitude float64, aspect string) error {
	return nil
}
func (m *mockRepository) GetLastClimbedForLocation(ctx context.Context, locationID int) (*models.LastClimbedInfo, error) {
	return nil, nil
}
func (m *mockRepository) GetClimbHistoryForLocation(ctx context.Context, locationID int, limit int) ([]models.ClimbHistoryEntry, error) {
	return nil, nil
}
func (m *mockRepository) GetMPAreaByID(ctx context.Context, mpAreaID string) (*models.MPArea, error) {
	return nil, nil
}
func (m *mockRepository) GetLastTickTimestampForRoute(ctx context.Context, routeID string) (*time.Time, error) {
	return nil, nil
}
func (m *mockRepository) GetAllRouteIDsForLocation(ctx context.Context, locationID int) ([]string, error) {
	return nil, nil
}
func (m *mockRepository) GetAreasOrderedByActivity(ctx context.Context, locationID int) ([]models.AreaActivitySummary, error) {
	return nil, nil
}
func (m *mockRepository) GetSubareasOrderedByActivity(ctx context.Context, parentAreaID string, locationID int) ([]models.AreaActivitySummary, error) {
	return nil, nil
}
func (m *mockRepository) GetRoutesOrderedByActivity(ctx context.Context, areaID string, locationID int, limit int) ([]models.RouteActivitySummary, error) {
	return nil, nil
}
func (m *mockRepository) GetRecentTicksForRoute(ctx context.Context, routeID string, limit int) ([]models.ClimbHistoryEntry, error) {
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
func (m *mockRepository) Close() error {
	return nil
}

// TestGetAreaDryingStats_AllDry tests area stats when all routes are dry
func TestGetAreaDryingStats_AllDry(t *testing.T) {
	now := time.Now()
	locationID := 1

	// Mock 3 routes, all dry (last rain was 48h ago)
	routes := []*models.MPRoute{
		{
			MPRouteID:  "route1",
			Name:       "Test Route 1",
			LocationID: &locationID,
			Latitude:   ptrFloat64(47.6),
			Longitude:  ptrFloat64(-120.9),
			Aspect:     ptrString("South"),
		},
		{
			MPRouteID:  "route2",
			Name:       "Test Route 2",
			LocationID: &locationID,
			Latitude:   ptrFloat64(47.61),
			Longitude:  ptrFloat64(-120.91),
			Aspect:     ptrString("East"),
		},
		{
			MPRouteID:  "route3",
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
		routes:            routes,
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

	service := NewBoulderDryingService(mock, nil)
	stats, err := service.GetAreaDryingStats(context.Background(), "area1", locationID)

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
			MPRouteID:  "route1",
			Name:       "Dry Route",
			LocationID: &locationID,
			Latitude:   ptrFloat64(47.6),
			Longitude:  ptrFloat64(-120.9),
			Aspect:     ptrString("South"),
		},
		{
			MPRouteID:  "route2",
			Name:       "Drying Route",
			LocationID: &locationID,
			Latitude:   ptrFloat64(47.61),
			Longitude:  ptrFloat64(-120.91),
			Aspect:     ptrString("East"),
		},
		{
			MPRouteID:  "route3",
			Name:       "Wet Route 1",
			LocationID: &locationID,
			Latitude:   ptrFloat64(47.62),
			Longitude:  ptrFloat64(-120.92),
			Aspect:     ptrString("North"),
		},
		{
			MPRouteID:  "route4",
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

	service := NewBoulderDryingService(mock, nil)
	stats, err := service.GetAreaDryingStats(context.Background(), "area1", locationID)

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

	service := NewBoulderDryingService(mock, nil)
	stats, err := service.GetAreaDryingStats(context.Background(), "area1", 1)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if stats != nil {
		t.Error("Expected nil stats for area with no routes, got non-nil")
	}
}

// Helper functions
func ptrFloat64(f float64) *float64 {
	return &f
}

func ptrString(s string) *string {
	return &s
}
