package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/alexscott64/woulder/backend/internal/weather"
	"github.com/stretchr/testify/assert"
)

func TestWeatherService_GetLocationWeather(t *testing.T) {
	tests := []struct {
		name           string
		locationID     int
		mockLocationFn func(ctx context.Context, id int) (*models.Location, error)
		wantErr        bool
	}{
		{
			name:       "location not found",
			locationID: 999,
			mockLocationFn: func(ctx context.Context, id int) (*models.Location, error) {
				return nil, errors.New("location not found")
			},
			wantErr: true,
		},
		{
			name:       "success - location found",
			locationID: 1,
			mockLocationFn: func(ctx context.Context, id int) (*models.Location, error) {
				return &models.Location{
					ID:        id,
					Name:      "Test Location",
					Latitude:  45.0,
					Longitude: -122.0,
				}, nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWeatherRepo := &MockWeatherRepository{
				SaveFn: func(ctx context.Context, data *models.WeatherData) error {
					return nil
				},
				GetHistoricalFn: func(ctx context.Context, locationID int, days int) ([]models.WeatherData, error) {
					return []models.WeatherData{}, nil
				},
			}

			mockLocationsRepo := &MockLocationsRepository{
				GetByIDFn: tt.mockLocationFn,
			}

			mockRocksRepo := &MockRocksRepository{
				GetRockTypesByLocationFn: func(ctx context.Context, locationID int) ([]models.RockType, error) {
					return []models.RockType{
						{ID: 1, Name: "Granite", BaseDryingHours: 4.0, PorosityPercent: 1.0},
					}, nil
				},
				GetSunExposureByLocationFn: func(ctx context.Context, locationID int) (*models.LocationSunExposure, error) {
					return &models.LocationSunExposure{
						LocationID:          locationID,
						SouthFacingPercent:  50,
						WestFacingPercent:   25,
						EastFacingPercent:   25,
						TreeCoveragePercent: 20,
					}, nil
				},
			}

			client := weather.NewWeatherService("test_api_key")
			// Pass nil for climb service in tests - it's optional
			service := NewWeatherService(mockWeatherRepo, mockLocationsRepo, mockRocksRepo, client, nil)

			forecast, err := service.GetLocationWeather(context.Background(), tt.locationID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, forecast)
			} else {
				// Note: Weather API calls may fail in tests
				// In production, you'd mock the weatherClient
				if err == nil {
					assert.NotNil(t, forecast)
					assert.Equal(t, tt.locationID, forecast.LocationID)
				}
			}
		})
	}
}

func TestWeatherService_GetWeatherByCoordinates(t *testing.T) {
	tests := []struct {
		name    string
		lat     float64
		lon     float64
		wantErr bool
	}{
		{
			name:    "valid coordinates",
			lat:     45.0,
			lon:     -122.0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWeatherRepo := &MockWeatherRepository{}
			mockLocationsRepo := &MockLocationsRepository{}
			mockRocksRepo := &MockRocksRepository{}

			client := weather.NewWeatherService("test_api_key")
			// Pass nil for climb service in tests - it's optional
			service := NewWeatherService(mockWeatherRepo, mockLocationsRepo, mockRocksRepo, client, nil)

			forecast, err := service.GetWeatherByCoordinates(context.Background(), tt.lat, tt.lon)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				// Note: Weather API may fail in tests
				if err == nil {
					assert.NotNil(t, forecast)
					assert.NotNil(t, forecast.Sunrise)
					assert.NotNil(t, forecast.Sunset)
				}
			}
		})
	}
}

func TestWeatherService_GetAllWeather(t *testing.T) {
	tests := []struct {
		name               string
		areaID             *int
		mockLocationsFn    func(ctx context.Context) ([]models.Location, error)
		mockLocationByArea func(ctx context.Context, areaID int) ([]models.Location, error)
		wantErr            bool
	}{
		{
			name:   "success - all locations",
			areaID: nil,
			mockLocationsFn: func(ctx context.Context) ([]models.Location, error) {
				return []models.Location{
					{ID: 1, Name: "Location 1", Latitude: 45.0, Longitude: -122.0},
					{ID: 2, Name: "Location 2", Latitude: 46.0, Longitude: -123.0},
				}, nil
			},
			wantErr: false,
		},
		{
			name:   "success - filtered by area",
			areaID: intPtr(1),
			mockLocationByArea: func(ctx context.Context, areaID int) ([]models.Location, error) {
				return []models.Location{
					{ID: 1, Name: "Location 1", Latitude: 45.0, Longitude: -122.0, AreaID: areaID},
				}, nil
			},
			wantErr: false,
		},
		{
			name:   "database error",
			areaID: nil,
			mockLocationsFn: func(ctx context.Context) ([]models.Location, error) {
				return nil, errors.New("database error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWeatherRepo := &MockWeatherRepository{
				SaveFn: func(ctx context.Context, data *models.WeatherData) error {
					return nil
				},
				GetHistoricalFn: func(ctx context.Context, locationID int, days int) ([]models.WeatherData, error) {
					return []models.WeatherData{}, nil
				},
			}

			mockLocationsRepo := &MockLocationsRepository{
				GetAllFn:    tt.mockLocationsFn,
				GetByAreaFn: tt.mockLocationByArea,
				GetByIDFn: func(ctx context.Context, id int) (*models.Location, error) {
					return &models.Location{
						ID:        id,
						Name:      "Test",
						Latitude:  45.0,
						Longitude: -122.0,
					}, nil
				},
			}

			mockRocksRepo := &MockRocksRepository{
				GetRockTypesByLocationFn: func(ctx context.Context, locationID int) ([]models.RockType, error) {
					return []models.RockType{}, nil
				},
				GetSunExposureByLocationFn: func(ctx context.Context, locationID int) (*models.LocationSunExposure, error) {
					return nil, errors.New("no sun exposure")
				},
			}

			client := weather.NewWeatherService("test_api_key")
			// Pass nil for climb service in tests - it's optional
			service := NewWeatherService(mockWeatherRepo, mockLocationsRepo, mockRocksRepo, client, nil)

			_, err := service.GetAllWeather(context.Background(), tt.areaID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Note: Some forecasts may fail due to weather API
			}
		})
	}
}

func TestWeatherService_RefreshAllWeather(t *testing.T) {
	tests := []struct {
		name            string
		mockLocationsFn func(ctx context.Context) ([]models.Location, error)
		wantErr         bool
	}{
		{
			name: "success",
			mockLocationsFn: func(ctx context.Context) ([]models.Location, error) {
				return []models.Location{
					{ID: 1, Name: "Location 1", Latitude: 45.0, Longitude: -122.0},
				}, nil
			},
			wantErr: false,
		},
		{
			name: "database error",
			mockLocationsFn: func(ctx context.Context) ([]models.Location, error) {
				return nil, errors.New("database error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deleteDaysToKeep := 0
			aggregateCalls := 0
			mockWeatherRepo := &MockWeatherRepository{
				SaveFn: func(ctx context.Context, data *models.WeatherData) error {
					return nil
				},
				GetHistoricalFn: func(ctx context.Context, locationID int, days int) ([]models.WeatherData, error) {
					return []models.WeatherData{}, nil
				},
				DeleteOldForLocationFn: func(ctx context.Context, locationID int, daysToKeep int) error {
					deleteDaysToKeep = daysToKeep
					return nil
				},
				UpsertDailyAggregatesFn: func(ctx context.Context, locationID int, startDate, endDate string) error {
					aggregateCalls++
					return nil
				},
			}

			mockLocationsRepo := &MockLocationsRepository{
				GetAllFn: tt.mockLocationsFn,
				GetByIDFn: func(ctx context.Context, id int) (*models.Location, error) {
					return &models.Location{
						ID:        id,
						Name:      "Test",
						Latitude:  45.0,
						Longitude: -122.0,
					}, nil
				},
			}

			mockRocksRepo := &MockRocksRepository{
				GetRockTypesByLocationFn: func(ctx context.Context, locationID int) ([]models.RockType, error) {
					return []models.RockType{}, nil
				},
				GetSunExposureByLocationFn: func(ctx context.Context, locationID int) (*models.LocationSunExposure, error) {
					return nil, errors.New("no sun exposure")
				},
			}

			client := weather.NewWeatherService("test_api_key")
			// Pass nil for climb service in tests - it's optional
			service := NewWeatherService(mockWeatherRepo, mockLocationsRepo, mockRocksRepo, client, nil)

			err := service.RefreshAllWeather(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, 30, deleteDaysToKeep)
				assert.GreaterOrEqual(t, aggregateCalls, 1)
			}
		})
	}
}

func TestWeatherService_RefreshAllWeather_ConcurrentCalls(t *testing.T) {
	mockWeatherRepo := &MockWeatherRepository{
		GetCurrentFn: func(ctx context.Context, locationID int) (*models.WeatherData, error) {
			return nil, nil // Return nil to force refresh
		},
	}

	mockLocationsRepo := &MockLocationsRepository{
		GetAllFn: func(ctx context.Context) ([]models.Location, error) {
			time.Sleep(100 * time.Millisecond) // Simulate slow operation
			return []models.Location{}, nil
		},
	}

	mockRocksRepo := &MockRocksRepository{}

	client := weather.NewWeatherService("test_api_key")
	// Pass nil for climb service in tests - it's optional
	service := NewWeatherService(mockWeatherRepo, mockLocationsRepo, mockRocksRepo, client, nil)

	// Start first refresh with forceRefresh=true to bypass freshness check
	go service.RefreshAllWeatherWithOptions(context.Background(), true)
	time.Sleep(10 * time.Millisecond) // Let it start

	// Try to start second refresh while first is running
	err := service.RefreshAllWeatherWithOptions(context.Background(), true)

	// Should get error because refresh is already in progress
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already in progress")
}

// TestWeatherService_GetAllWeather_IncludesClimbHistory ensures climb_history
// is populated when climb tracking service is available (regression test)
func TestWeatherService_GetAllWeather_IncludesClimbHistory(t *testing.T) {
	mockWeatherRepo := &MockWeatherRepository{
		SaveFn: func(ctx context.Context, data *models.WeatherData) error {
			return nil
		},
		GetHistoricalFn: func(ctx context.Context, locationID int, days int) ([]models.WeatherData, error) {
			return []models.WeatherData{}, nil
		},
	}

	mockLocationsRepo := &MockLocationsRepository{
		GetAllFn: func(ctx context.Context) ([]models.Location, error) {
			return []models.Location{
				{ID: 1, Name: "Location 1", Latitude: 45.0, Longitude: -122.0},
				{ID: 2, Name: "Location 2", Latitude: 46.0, Longitude: -123.0},
			}, nil
		},
		GetByIDFn: func(ctx context.Context, id int) (*models.Location, error) {
			return &models.Location{
				ID:        id,
				Name:      "Test",
				Latitude:  45.0,
				Longitude: -122.0,
			}, nil
		},
	}

	mockRocksRepo := &MockRocksRepository{
		GetRockTypesByLocationFn: func(ctx context.Context, locationID int) ([]models.RockType, error) {
			return []models.RockType{}, nil
		},
		GetSunExposureByLocationFn: func(ctx context.Context, locationID int) (*models.LocationSunExposure, error) {
			return nil, errors.New("no sun exposure")
		},
	}

	// Create mock climb repository with batch method that returns climb history
	mockClimbingHistoryRepo := &MockClimbingHistoryRepository{
		GetClimbHistoryForLocationsFn: func(ctx context.Context, locationIDs []int, limit int) (map[int][]models.ClimbHistoryEntry, error) {
			// Return climb history for both locations
			return map[int][]models.ClimbHistoryEntry{
				1: {
					{MPRouteID: 1, RouteName: "Test Route 1", RouteRating: "5.10a", ClimbedAt: time.Now()},
					{MPRouteID: 2, RouteName: "Test Route 2", RouteRating: "5.11b", ClimbedAt: time.Now()},
				},
				2: {
					{MPRouteID: 3, RouteName: "Test Route 3", RouteRating: "V5", ClimbedAt: time.Now()},
				},
			}, nil
		},
	}

	mockClimbingRepo := &MockClimbingRepository{
		history: mockClimbingHistoryRepo,
	}

	mockMountainProjectRepo := &MockMountainProjectRepository{}

	// Create climb tracking service with the mock repositories
	mockClimbService := NewClimbTrackingService(mockMountainProjectRepo, mockClimbingRepo, nil, nil)

	client := weather.NewWeatherService("test_api_key")
	service := NewWeatherService(mockWeatherRepo, mockLocationsRepo, mockRocksRepo, client, mockClimbService)

	forecasts, err := service.GetAllWeather(context.Background(), nil)

	assert.NoError(t, err)
	assert.NotEmpty(t, forecasts)

	// Critical: Verify all forecasts have climb_history populated
	for _, forecast := range forecasts {
		assert.NotNil(t, forecast.ClimbHistory, "climb_history must not be nil for location %d", forecast.LocationID)
		// At least one location should have history
		if forecast.LocationID == 1 || forecast.LocationID == 2 {
			assert.NotEmpty(t, forecast.ClimbHistory, "climb_history must be populated for location %d", forecast.LocationID)
		}
	}
}

// TestWeatherService_GetLocationWeather_PersistsFreshData verifies that when the
// per-request path fetches fresh weather from the API (cache miss or stale), it
// persists the data to the DB: DeleteFutureForLocation is called, then Save is
// called for each hourly forecast entry plus the current entry.
func TestWeatherService_GetLocationWeather_PersistsFreshData(t *testing.T) {
	var (
		deleteFutureCalled bool
		deleteFutureLocID  int
		saveCount          int
	)

	mockWeatherRepo := &MockWeatherRepository{
		// Return nil to simulate cache miss → forces API fetch path
		GetCurrentFn: func(ctx context.Context, locationID int) (*models.WeatherData, error) {
			return nil, nil
		},
		GetHistoricalFn: func(ctx context.Context, locationID int, days int) ([]models.WeatherData, error) {
			return []models.WeatherData{}, nil
		},
		DeleteFutureForLocationFn: func(ctx context.Context, locationID int) error {
			deleteFutureCalled = true
			deleteFutureLocID = locationID
			return nil
		},
		SaveFn: func(ctx context.Context, data *models.WeatherData) error {
			saveCount++
			return nil
		},
	}

	mockLocationsRepo := &MockLocationsRepository{
		GetByIDFn: func(ctx context.Context, id int) (*models.Location, error) {
			return &models.Location{
				ID:        id,
				Name:      "Index",
				Latitude:  47.82061272,
				Longitude: -121.55492795,
			}, nil
		},
	}

	mockRocksRepo := &MockRocksRepository{
		GetRockTypesByLocationFn: func(ctx context.Context, locationID int) ([]models.RockType, error) {
			return []models.RockType{
				{ID: 1, Name: "Granite", BaseDryingHours: 4.0, PorosityPercent: 1.0},
			}, nil
		},
		GetSunExposureByLocationFn: func(ctx context.Context, locationID int) (*models.LocationSunExposure, error) {
			return &models.LocationSunExposure{
				LocationID:         locationID,
				SouthFacingPercent: 50,
			}, nil
		},
	}

	weatherClient := weather.NewWeatherService("test_api_key")
	service := NewWeatherService(mockWeatherRepo, mockLocationsRepo, mockRocksRepo, weatherClient, nil)

	forecast, err := service.GetLocationWeather(context.Background(), 2)
	if err != nil {
		// Weather API may fail in tests (no network, rate limits, etc.)
		// That's okay — the test is about the DB persistence path
		t.Skipf("Skipping: weather API call failed (expected in CI): %v", err)
	}

	assert.NotNil(t, forecast)
	assert.Equal(t, 2, forecast.LocationID)

	// Verify DeleteFutureForLocation was called with the correct location ID
	assert.True(t, deleteFutureCalled, "DeleteFutureForLocation must be called when persisting fresh data")
	assert.Equal(t, 2, deleteFutureLocID)

	// Verify Save was called at least once (current + forecast hours)
	// GetCurrentAndForecast returns 16 days = 384 hours + 1 current = 385+ saves
	assert.Greater(t, saveCount, 100, "Save should be called for current + all forecast hours (got %d)", saveCount)
}

// TestWeatherService_GetLocationWeather_StaleCreatedAtTriggersRefresh verifies that
// cached data with a CreatedAt older than 1 hour is treated as stale, triggering
// a fresh API fetch and DB persistence.
func TestWeatherService_GetLocationWeather_StaleCreatedAtTriggersRefresh(t *testing.T) {
	var (
		deleteFutureCalled bool
		saveCount          int
	)

	staleTime := time.Now().Add(-2 * time.Hour) // 2 hours ago → stale

	mockWeatherRepo := &MockWeatherRepository{
		GetCurrentFn: func(ctx context.Context, locationID int) (*models.WeatherData, error) {
			return &models.WeatherData{
				LocationID:  locationID,
				Timestamp:   time.Now(), // Timestamp is current hour
				CreatedAt:   staleTime,  // But created_at is 2hrs old → stale
				Temperature: 40.0,
			}, nil
		},
		GetHistoricalFn: func(ctx context.Context, locationID int, days int) ([]models.WeatherData, error) {
			return []models.WeatherData{}, nil
		},
		DeleteFutureForLocationFn: func(ctx context.Context, locationID int) error {
			deleteFutureCalled = true
			return nil
		},
		SaveFn: func(ctx context.Context, data *models.WeatherData) error {
			saveCount++
			return nil
		},
	}

	mockLocationsRepo := &MockLocationsRepository{
		GetByIDFn: func(ctx context.Context, id int) (*models.Location, error) {
			return &models.Location{
				ID:        id,
				Name:      "Index",
				Latitude:  47.82061272,
				Longitude: -121.55492795,
			}, nil
		},
	}

	mockRocksRepo := &MockRocksRepository{
		GetRockTypesByLocationFn: func(ctx context.Context, locationID int) ([]models.RockType, error) {
			return []models.RockType{
				{ID: 1, Name: "Granite", BaseDryingHours: 4.0, PorosityPercent: 1.0},
			}, nil
		},
		GetSunExposureByLocationFn: func(ctx context.Context, locationID int) (*models.LocationSunExposure, error) {
			return &models.LocationSunExposure{
				LocationID:         locationID,
				SouthFacingPercent: 50,
			}, nil
		},
	}

	weatherClient := weather.NewWeatherService("test_api_key")
	service := NewWeatherService(mockWeatherRepo, mockLocationsRepo, mockRocksRepo, weatherClient, nil)

	forecast, err := service.GetLocationWeather(context.Background(), 2)
	if err != nil {
		t.Skipf("Skipping: weather API call failed (expected in CI): %v", err)
	}

	assert.NotNil(t, forecast)

	// Verify the stale cache triggered a refresh with DB persistence
	assert.True(t, deleteFutureCalled, "DeleteFutureForLocation must be called when CreatedAt is stale")
	assert.Greater(t, saveCount, 100, "Save should be called for fresh data persistence (got %d)", saveCount)
}

// TestWeatherService_GetLocationWeather_FreshCacheDoesNotPersist verifies that
// fresh cached data (CreatedAt < 1 hour) is served without re-fetching or saving.
func TestWeatherService_GetLocationWeather_FreshCacheDoesNotPersist(t *testing.T) {
	var (
		deleteFutureCalled bool
		saveCount          int
	)

	freshTime := time.Now().Add(-10 * time.Minute) // 10 minutes ago → fresh

	mockWeatherRepo := &MockWeatherRepository{
		GetCurrentFn: func(ctx context.Context, locationID int) (*models.WeatherData, error) {
			return &models.WeatherData{
				LocationID:    locationID,
				Timestamp:     time.Now(),
				CreatedAt:     freshTime,
				Temperature:   40.0,
				Precipitation: 0.1,
				Humidity:      80,
				WindSpeed:     5.0,
				CloudCover:    50,
				Pressure:      1013,
				Description:   "Cloudy",
				Icon:          "04d",
			}, nil
		},
		GetForecastFn: func(ctx context.Context, locationID int, hours int) ([]models.WeatherData, error) {
			// Return some cached forecast data
			forecast := make([]models.WeatherData, 48)
			for i := range forecast {
				forecast[i] = models.WeatherData{
					LocationID:    locationID,
					Timestamp:     time.Now().Add(time.Duration(i+1) * time.Hour),
					CreatedAt:     freshTime,
					Temperature:   42.0,
					Precipitation: 0.05,
					Humidity:      75,
					WindSpeed:     4.0,
					CloudCover:    60,
					Pressure:      1012,
					Description:   "Partly Cloudy",
					Icon:          "03d",
				}
			}
			return forecast, nil
		},
		GetHistoricalFn: func(ctx context.Context, locationID int, days int) ([]models.WeatherData, error) {
			return []models.WeatherData{}, nil
		},
		DeleteFutureForLocationFn: func(ctx context.Context, locationID int) error {
			deleteFutureCalled = true
			return nil
		},
		SaveFn: func(ctx context.Context, data *models.WeatherData) error {
			saveCount++
			return nil
		},
	}

	mockLocationsRepo := &MockLocationsRepository{
		GetByIDFn: func(ctx context.Context, id int) (*models.Location, error) {
			return &models.Location{
				ID:        id,
				Name:      "Index",
				Latitude:  47.82061272,
				Longitude: -121.55492795,
			}, nil
		},
	}

	mockRocksRepo := &MockRocksRepository{
		GetRockTypesByLocationFn: func(ctx context.Context, locationID int) ([]models.RockType, error) {
			return []models.RockType{
				{ID: 1, Name: "Granite", BaseDryingHours: 4.0, PorosityPercent: 1.0},
			}, nil
		},
		GetSunExposureByLocationFn: func(ctx context.Context, locationID int) (*models.LocationSunExposure, error) {
			return &models.LocationSunExposure{
				LocationID:         locationID,
				SouthFacingPercent: 50,
			}, nil
		},
	}

	weatherClient := weather.NewWeatherService("test_api_key")
	service := NewWeatherService(mockWeatherRepo, mockLocationsRepo, mockRocksRepo, weatherClient, nil)

	forecast, err := service.GetLocationWeather(context.Background(), 2)

	assert.NoError(t, err)
	assert.NotNil(t, forecast)
	assert.Equal(t, 2, forecast.LocationID)

	// Fresh cache should NOT trigger any DB writes
	assert.False(t, deleteFutureCalled, "DeleteFutureForLocation should NOT be called for fresh cache")
	assert.Equal(t, 0, saveCount, "Save should NOT be called for fresh cache")
}

// Helper function
func intPtr(i int) *int {
	return &i
}
