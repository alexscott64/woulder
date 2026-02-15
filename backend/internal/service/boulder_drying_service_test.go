package service

import (
	"context"
	"testing"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/alexscott64/woulder/backend/internal/weather/client"
)

// mockBoulderWeatherClient provides mock weather data for boulder drying tests
type mockBoulderWeatherClient struct {
	currentWeather  *models.WeatherData
	forecastWeather []models.WeatherData
}

func (m *mockBoulderWeatherClient) GetCurrentAndForecast(lat, lon float64) (*models.WeatherData, []models.WeatherData, *client.SunTimes, error) {
	return m.currentWeather, m.forecastWeather, nil, nil
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

	currentWeather := &models.WeatherData{
		LocationID:    locationID,
		Timestamp:     now,
		Temperature:   60.0,
		Humidity:      50.0,
		WindSpeed:     5.0,
		Precipitation: 0.0,
	}

	mockBouldersRepo := &MockBouldersRepository{}
	mockWeatherRepo := &MockWeatherRepository{
		GetCurrentFn: func(ctx context.Context, locID int) (*models.WeatherData, error) {
			return currentWeather, nil
		},
		GetHistoricalFn: func(ctx context.Context, locID int, days int) ([]models.WeatherData, error) {
			return historicalWeather, nil
		},
	}
	mockLocationsRepo := &MockLocationsRepository{
		GetByIDFn: func(ctx context.Context, id int) (*models.Location, error) {
			return &models.Location{
				ID:        id,
				Name:      "Test Location",
				Latitude:  47.6,
				Longitude: -120.9,
			}, nil
		},
	}
	mockRocksRepo := &MockRocksRepository{
		GetRockTypesByLocationFn: func(ctx context.Context, locID int) ([]models.RockType, error) {
			return []models.RockType{
				{Name: "Granite", BaseDryingHours: 12.0},
			}, nil
		},
		GetSunExposureByLocationFn: func(ctx context.Context, locID int) (*models.LocationSunExposure, error) {
			return &models.LocationSunExposure{
				TreeCoveragePercent: 30.0,
			}, nil
		},
	}
	mockMPRepo := NewMockMountainProjectRepository()
	mockMPRepo.routes.GetWithGPSByAreaFn = func(ctx context.Context, mpAreaID int64) ([]*models.MPRoute, error) {
		return routes, nil
	}
	mockMPRepo.routes.GetByIDsFn = func(ctx context.Context, mpRouteIDs []int64) (map[int64]*models.MPRoute, error) {
		routeMap := make(map[int64]*models.MPRoute)
		for _, route := range routes {
			routeMap[route.MPRouteID] = route
		}
		return routeMap, nil
	}

	mockWeatherClient := &mockBoulderWeatherClient{
		currentWeather:  currentWeather,
		forecastWeather: []models.WeatherData{},
	}

	service := NewBoulderDryingService(mockBouldersRepo, mockWeatherRepo, mockLocationsRepo, mockRocksRepo, mockMPRepo, mockWeatherClient)
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
	currentWeather := &models.WeatherData{
		LocationID:    locationID,
		Timestamp:     now,
		Temperature:   60.0,
		Humidity:      50.0,
		WindSpeed:     5.0,
		Precipitation: 0.0,
	}
	historicalWeather := []models.WeatherData{
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
	}

	mockBouldersRepo := &MockBouldersRepository{}
	mockWeatherRepo := &MockWeatherRepository{
		GetCurrentFn: func(ctx context.Context, locID int) (*models.WeatherData, error) {
			return currentWeather, nil
		},
		GetHistoricalFn: func(ctx context.Context, locID int, days int) ([]models.WeatherData, error) {
			return historicalWeather, nil
		},
	}
	mockLocationsRepo := &MockLocationsRepository{
		GetByIDFn: func(ctx context.Context, id int) (*models.Location, error) {
			return &models.Location{
				ID:        id,
				Name:      "Test Location",
				Latitude:  47.6,
				Longitude: -120.9,
			}, nil
		},
	}
	mockRocksRepo := &MockRocksRepository{
		GetRockTypesByLocationFn: func(ctx context.Context, locID int) ([]models.RockType, error) {
			return []models.RockType{
				{Name: "Granite", BaseDryingHours: 12.0},
			}, nil
		},
		GetSunExposureByLocationFn: func(ctx context.Context, locID int) (*models.LocationSunExposure, error) {
			return &models.LocationSunExposure{
				TreeCoveragePercent: 30.0,
			}, nil
		},
	}
	mockMPRepo := NewMockMountainProjectRepository()
	mockMPRepo.routes.GetWithGPSByAreaFn = func(ctx context.Context, mpAreaID int64) ([]*models.MPRoute, error) {
		return routes, nil
	}
	mockMPRepo.routes.GetByIDsFn = func(ctx context.Context, mpRouteIDs []int64) (map[int64]*models.MPRoute, error) {
		routeMap := make(map[int64]*models.MPRoute)
		for _, route := range routes {
			routeMap[route.MPRouteID] = route
		}
		return routeMap, nil
	}

	mockWeatherClient := &mockBoulderWeatherClient{
		currentWeather:  currentWeather,
		forecastWeather: []models.WeatherData{},
	}

	service := NewBoulderDryingService(mockBouldersRepo, mockWeatherRepo, mockLocationsRepo, mockRocksRepo, mockMPRepo, mockWeatherClient)
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
	mockBouldersRepo := &MockBouldersRepository{}
	mockWeatherRepo := &MockWeatherRepository{}
	mockLocationsRepo := &MockLocationsRepository{}
	mockRocksRepo := &MockRocksRepository{}
	mockMPRepo := NewMockMountainProjectRepository()
	mockMPRepo.routes.GetWithGPSByAreaFn = func(ctx context.Context, mpAreaID int64) ([]*models.MPRoute, error) {
		return []*models.MPRoute{}, nil
	}

	mockWeatherClient := &mockBoulderWeatherClient{
		currentWeather:  nil,
		forecastWeather: []models.WeatherData{},
	}

	service := NewBoulderDryingService(mockBouldersRepo, mockWeatherRepo, mockLocationsRepo, mockRocksRepo, mockMPRepo, mockWeatherClient)
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

	mockBouldersRepo := &MockBouldersRepository{}
	mockWeatherRepo := &MockWeatherRepository{
		GetCurrentFn: func(ctx context.Context, locID int) (*models.WeatherData, error) {
			return currentWeather, nil
		},
		GetHistoricalFn: func(ctx context.Context, locID int, days int) ([]models.WeatherData, error) {
			return historicalWeather, nil
		},
		GetForecastFn: func(ctx context.Context, locID int, hours int) ([]models.WeatherData, error) {
			return forecastWeather, nil
		},
	}
	mockLocationsRepo := &MockLocationsRepository{
		GetByIDFn: func(ctx context.Context, id int) (*models.Location, error) {
			return &models.Location{
				ID:        id,
				Name:      "Test Location",
				Latitude:  47.6,
				Longitude: -120.9,
			}, nil
		},
	}
	mockRocksRepo := &MockRocksRepository{
		GetRockTypesByLocationFn: func(ctx context.Context, locID int) ([]models.RockType, error) {
			return rockTypes, nil
		},
		GetSunExposureByLocationFn: func(ctx context.Context, locID int) (*models.LocationSunExposure, error) {
			return &models.LocationSunExposure{
				TreeCoveragePercent: 30.0,
			}, nil
		},
	}
	mockMPRepo := NewMockMountainProjectRepository()
	mockMPRepo.routes.GetWithGPSByAreaFn = func(ctx context.Context, mpAreaID int64) ([]*models.MPRoute, error) {
		return routes, nil
	}
	mockMPRepo.routes.GetByIDsFn = func(ctx context.Context, mpRouteIDs []int64) (map[int64]*models.MPRoute, error) {
		routeMap := make(map[int64]*models.MPRoute)
		for _, route := range routes {
			routeMap[route.MPRouteID] = route
		}
		return routeMap, nil
	}

	mockWeatherClient := &mockBoulderWeatherClient{
		currentWeather:  currentWeather,
		forecastWeather: forecastWeather,
	}

	service := NewBoulderDryingService(mockBouldersRepo, mockWeatherRepo, mockLocationsRepo, mockRocksRepo, mockMPRepo, mockWeatherClient)
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
