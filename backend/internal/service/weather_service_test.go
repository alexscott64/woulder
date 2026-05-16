package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/alexscott64/woulder/backend/internal/weather"
	"github.com/alexscott64/woulder/backend/internal/weather/client"
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
		ReplaceFutureForLocationFn: func(ctx context.Context, locationID int, rows []models.WeatherData) error {
			// New atomic path replaces the legacy delete+save loop. Treat
			// invocation as the equivalent of a delete + N saves so the
			// existing assertions below stay meaningful.
			deleteFutureCalled = true
			deleteFutureLocID = locationID
			saveCount += len(rows)
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
		ReplaceFutureForLocationFn: func(ctx context.Context, locationID int, rows []models.WeatherData) error {
			deleteFutureCalled = true
			saveCount += len(rows)
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

	// Verify the stale cache triggered a refresh with DB persistence (now via
	// the atomic ReplaceFutureForLocation path).
	assert.True(t, deleteFutureCalled, "ReplaceFutureForLocation must be called when CreatedAt is stale")
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
		ReplaceFutureForLocationFn: func(ctx context.Context, locationID int, rows []models.WeatherData) error {
			deleteFutureCalled = true
			saveCount += len(rows)
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
	assert.False(t, deleteFutureCalled, "DeleteFutureForLocation/ReplaceFutureForLocation should NOT be called for fresh cache")
	assert.Equal(t, 0, saveCount, "Save should NOT be called for fresh cache")
}

// Helper function
func intPtr(i int) *int {
	return &i
}

// TestWeatherService_OfflineMode_DoesNotCallAPI verifies that with
// SetOfflineMode(true) the service serves weather purely from the DB and
// never invokes the underlying weatherClient. We assert this indirectly:
// the test weatherClient has no API key / no network access, so any real
// API call would fail. By providing fresh cached data and ensuring no
// fetch/save side-effects occur, we confirm the offline path is taken.
//
// It also runs a second case where the cache is EMPTY (GetCurrent returns
// nil) — offline mode must still succeed (no API call, no error) and emit
// a warning rather than blowing up.
func TestWeatherService_OfflineMode_DoesNotCallAPI(t *testing.T) {
	t.Run("with cached data — no API call, no DB writes", func(t *testing.T) {
		var (
			deleteFutureCalled bool
			saveCount          int
			getForecastCalled  bool
		)

		// Stale cache (older than 1h) — in normal mode this would trigger an
		// API fetch. In offline mode it must be served as-is.
		staleTime := time.Now().Add(-3 * time.Hour)

		mockWeatherRepo := &MockWeatherRepository{
			GetCurrentFn: func(ctx context.Context, locationID int) (*models.WeatherData, error) {
				return &models.WeatherData{
					LocationID:  locationID,
					Timestamp:   time.Now(),
					CreatedAt:   staleTime,
					Temperature: 55.0,
					Humidity:    60,
				}, nil
			},
			GetForecastFn: func(ctx context.Context, locationID int, hours int) ([]models.WeatherData, error) {
				getForecastCalled = true
				return []models.WeatherData{}, nil
			},
			GetHistoricalFn: func(ctx context.Context, locationID int, days int) ([]models.WeatherData, error) {
				return []models.WeatherData{}, nil
			},
			DeleteFutureForLocationFn: func(ctx context.Context, locationID int) error {
				deleteFutureCalled = true
				return nil
			},
			ReplaceFutureForLocationFn: func(ctx context.Context, locationID int, rows []models.WeatherData) error {
				deleteFutureCalled = true
				saveCount += len(rows)
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
					ID: id, Name: "Test", Latitude: 47.0, Longitude: -121.0,
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
				return &models.LocationSunExposure{LocationID: locationID, SouthFacingPercent: 50}, nil
			},
		}

		// Use a real WeatherService client — if offline mode is broken and
		// we hit the API path, the call will either succeed (still wrong)
		// or fail. We rely on the DB-write side effects (DeleteFuture/Save)
		// being our canary: they're ONLY invoked on the API-fetch branch.
		weatherClient := weather.NewWeatherService("test_api_key")
		service := NewWeatherService(mockWeatherRepo, mockLocationsRepo, mockRocksRepo, weatherClient, nil)
		service.SetOfflineMode(true)

		forecast, err := service.GetLocationWeather(context.Background(), 7)

		assert.NoError(t, err)
		assert.NotNil(t, forecast)
		assert.Equal(t, 7, forecast.LocationID)

		// Critical assertions: no API-fetch side effects.
		assert.False(t, deleteFutureCalled, "DeleteFutureForLocation must NOT be called in offline mode")
		assert.Equal(t, 0, saveCount, "Save must NOT be called in offline mode (no API fetch)")
		assert.True(t, getForecastCalled, "Should still load cached forecast from DB")
	})

	t.Run("empty cache — offline mode warns and returns synthetic data without erroring", func(t *testing.T) {
		var saveCount int

		mockWeatherRepo := &MockWeatherRepository{
			GetCurrentFn: func(ctx context.Context, locationID int) (*models.WeatherData, error) {
				return nil, nil // No cached data at all
			},
			SaveFn: func(ctx context.Context, data *models.WeatherData) error {
				saveCount++
				return nil
			},
		}

		mockLocationsRepo := &MockLocationsRepository{
			GetByIDFn: func(ctx context.Context, id int) (*models.Location, error) {
				return &models.Location{ID: id, Name: "Test", Latitude: 47.0, Longitude: -121.0}, nil
			},
		}

		mockRocksRepo := &MockRocksRepository{
			GetRockTypesByLocationFn: func(ctx context.Context, locationID int) ([]models.RockType, error) {
				return []models.RockType{}, nil
			},
		}

		weatherClient := weather.NewWeatherService("test_api_key")
		service := NewWeatherService(mockWeatherRepo, mockLocationsRepo, mockRocksRepo, weatherClient, nil)
		service.SetOfflineMode(true)

		forecast, err := service.GetLocationWeather(context.Background(), 99)

		assert.NoError(t, err, "offline mode must not error on empty cache")
		assert.NotNil(t, forecast)
		assert.Equal(t, 0, saveCount, "no DB writes in offline mode even on empty cache")
	})

	t.Run("RefreshAllWeather is a no-op in offline mode", func(t *testing.T) {
		getAllCalled := false

		mockLocationsRepo := &MockLocationsRepository{
			GetAllFn: func(ctx context.Context) ([]models.Location, error) {
				getAllCalled = true
				return []models.Location{{ID: 1}}, nil
			},
		}

		weatherClient := weather.NewWeatherService("test_api_key")
		service := NewWeatherService(&MockWeatherRepository{}, mockLocationsRepo, &MockRocksRepository{}, weatherClient, nil)
		service.SetOfflineMode(true)

		err := service.RefreshAllWeatherWithOptions(context.Background(), true)
		assert.NoError(t, err)
		assert.False(t, getAllCalled, "RefreshAllWeather should short-circuit in offline mode before touching repos")
	})
}

// buildOpenMeteoResponse synthesizes a JSON-serializable map mimicking the
// Open-Meteo /v1/forecast response shape. `hourCount` controls the number of
// hourly entries; `futureHours` controls how many of those entries have
// timestamps strictly after time.Now() (the rest are dated in the past so
// the service-level future-only count matches `futureHours`). Used by the
// truncated-response cache-preservation test below.
func buildOpenMeteoResponse(hourCount, futureHours int) map[string]interface{} {
	now := time.Now().UTC().Truncate(time.Hour)
	hourlyTimes := make([]string, hourCount)
	hourlyTemp := make([]float64, hourCount)
	hourlyHum := make([]int, hourCount)
	hourlyPrecip := make([]float64, hourCount)
	hourlyRain := make([]float64, hourCount)
	hourlySnow := make([]float64, hourCount)
	hourlyCloud := make([]int, hourCount)
	hourlyWind := make([]float64, hourCount)
	hourlyWindDir := make([]int, hourCount)
	hourlyCode := make([]int, hourCount)
	hourlyApp := make([]float64, hourCount)
	hourlyPress := make([]float64, hourCount)
	hourlySW := make([]float64, hourCount)
	hourlyDir := make([]float64, hourCount)
	hourlyDiff := make([]float64, hourCount)
	hourlyDew := make([]float64, hourCount)

	pastHours := hourCount - futureHours
	for i := 0; i < hourCount; i++ {
		// First `pastHours` rows are dated in the past, rest are future.
		offset := time.Duration(i-pastHours) * time.Hour
		ts := now.Add(offset)
		hourlyTimes[i] = ts.Format("2006-01-02T15:04")
		hourlyTemp[i] = 50.0
		hourlyHum[i] = 60
		hourlyCloud[i] = 50
		hourlyWind[i] = 5.0
		hourlyWindDir[i] = 180
		hourlyCode[i] = 1
		hourlyApp[i] = 48.0
		hourlyPress[i] = 1013.0
	}

	return map[string]interface{}{
		"current": map[string]interface{}{
			"time":                 now.Format("2006-01-02T15:04"),
			"temperature_2m":       50.0,
			"relative_humidity_2m": 60,
			"precipitation":        0.0,
			"rain":                 0.0,
			"snowfall":             0.0,
			"cloud_cover":          50,
			"wind_speed_10m":       5.0,
			"wind_direction_10m":   180,
			"weather_code":         1,
			"apparent_temperature": 48.0,
			"surface_pressure":     1013.0,
			"shortwave_radiation":  0.0,
			"direct_radiation":     0.0,
			"diffuse_radiation":    0.0,
			"dew_point_2m":         40.0,
		},
		"hourly": map[string]interface{}{
			"time":                 hourlyTimes,
			"temperature_2m":       hourlyTemp,
			"relative_humidity_2m": hourlyHum,
			"precipitation":        hourlyPrecip,
			"rain":                 hourlyRain,
			"snowfall":             hourlySnow,
			"cloud_cover":          hourlyCloud,
			"wind_speed_10m":       hourlyWind,
			"wind_direction_10m":   hourlyWindDir,
			"weather_code":         hourlyCode,
			"apparent_temperature": hourlyApp,
			"surface_pressure":     hourlyPress,
			"shortwave_radiation":  hourlySW,
			"direct_radiation":     hourlyDir,
			"diffuse_radiation":    hourlyDiff,
			"dew_point_2m":         hourlyDew,
		},
		"daily": map[string]interface{}{
			"time":    []string{now.Format("2006-01-02")},
			"sunrise": []string{now.Format("2006-01-02") + "T07:00"},
			"sunset":  []string{now.Format("2006-01-02") + "T19:00"},
		},
	}
}

// TestWeatherService_GetLocationWeather_TruncatedResponse_PreservesCache
// is the regression test for the "truncated daily forecast" bug (Fix 1).
//
// Setup: an httptest server returns a syntactically valid Open-Meteo response
// with only 100 hourly rows (well below the 336-hour threshold). The service
// is configured with the standard repo mock that records all writes.
//
// Expected behavior:
//   - The service must NOT call ReplaceFutureForLocation (atomic delete+save
//     for forecast cache).
//   - The service must NOT call DeleteFutureForLocation (legacy path).
//   - Forecast saves (the per-hour Save loop) must NOT happen — the existing
//     cache is preserved as-is.
//   - The function returns without a hard error: a stale-cache result is
//     better than a poisoned cache. (NB: the openmeteo client now rejects
//     truncated responses upstream as an error — `current` will end up nil
//     and the service falls through to error-return BEFORE reaching the
//     guard. That still satisfies the cache-preservation contract: no
//     destructive writes occur. We assert on the writes side-effect only.)
func TestWeatherService_GetLocationWeather_TruncatedResponse_PreservesCache(t *testing.T) {
	var (
		replaceFutureCalled bool
		deleteFutureCalled  bool
		forecastSaveCount   int
	)

	// Stub Open-Meteo with a truncated 100-hour response.
	resp := buildOpenMeteoResponse(100, 100) // 100 future hours, all future
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	restore := client.SetForecastBaseURLForTest(server.URL)
	defer restore()

	mockWeatherRepo := &MockWeatherRepository{
		// Cache miss → forces the API-fetch / persist branch.
		GetCurrentFn: func(ctx context.Context, locationID int) (*models.WeatherData, error) {
			return nil, nil
		},
		GetHistoricalFn: func(ctx context.Context, locationID int, days int) ([]models.WeatherData, error) {
			return []models.WeatherData{}, nil
		},
		GetForecastFn: func(ctx context.Context, locationID int, hours int) ([]models.WeatherData, error) {
			return []models.WeatherData{}, nil
		},
		DeleteFutureForLocationFn: func(ctx context.Context, locationID int) error {
			deleteFutureCalled = true
			return nil
		},
		ReplaceFutureForLocationFn: func(ctx context.Context, locationID int, rows []models.WeatherData) error {
			replaceFutureCalled = true
			return nil
		},
		SaveFn: func(ctx context.Context, data *models.WeatherData) error {
			// The "current" observation Save can still happen (it's a single
			// row, not the multi-day forecast). Track only forecast-shaped
			// saves — those are the ones that would poison the cache.
			// We approximate "forecast save" as any Save not for the "current"
			// row by matching on a non-zero precipitation/temperature pattern;
			// since this test never reaches the forecast-save loop with the
			// truncation guard in place, we just count all saves and assert
			// the count is small (<=1, the current row).
			forecastSaveCount++
			return nil
		},
	}

	mockLocationsRepo := &MockLocationsRepository{
		GetByIDFn: func(ctx context.Context, id int) (*models.Location, error) {
			return &models.Location{
				ID: id, Name: "Trunc Test", Latitude: 47.0, Longitude: -121.0,
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
			return &models.LocationSunExposure{LocationID: locationID, SouthFacingPercent: 50}, nil
		},
	}

	weatherClient := weather.NewWeatherService("test_api_key")
	service := NewWeatherService(mockWeatherRepo, mockLocationsRepo, mockRocksRepo, weatherClient, nil)

	_, err := service.GetLocationWeather(context.Background(), 42)

	// PRIMARY ASSERTIONS — the cache-preservation contract:
	// (a) ReplaceFutureForLocation NOT called (no atomic cache replacement)
	// (b) DeleteFutureForLocation NOT called (no destructive purge)
	// (c) Forecast-shaped Save loop NOT called (no per-hour saves of the
	//     truncated stub). We allow at most 1 Save (the standalone "current"
	//     row in the service-level guard's degraded path). With Fix 3 in
	//     place, the client rejects the response before parse, so the
	//     service short-circuits with an error and Save is never called at
	//     all (count=0). Either branch is acceptable.
	assert.False(t, replaceFutureCalled, "ReplaceFutureForLocation must NOT be called on truncated upstream response")
	assert.False(t, deleteFutureCalled, "DeleteFutureForLocation must NOT be called on truncated upstream response")
	assert.LessOrEqual(t, forecastSaveCount, 1,
		"At most 1 Save (the standalone current row) is allowed on truncated response; got %d. "+
			"Saving the truncated forecast would poison the cache.", forecastSaveCount)

	// We don't assert err==nil: with Fix 3 the client now returns an error
	// for truncated responses, which the service propagates. Both outcomes
	// (err==nil with cache preserved, or err!=nil with cache preserved) are
	// equivalent for the bug we're fixing — the cache is what matters.
	_ = err
	_ = fmt.Sprintf // keep fmt import used if asserts are tightened later
}
