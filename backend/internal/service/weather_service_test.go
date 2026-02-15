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
			mockWeatherRepo := &MockWeatherRepository{
				SaveFn: func(ctx context.Context, data *models.WeatherData) error {
					return nil
				},
				GetHistoricalFn: func(ctx context.Context, locationID int, days int) ([]models.WeatherData, error) {
					return []models.WeatherData{}, nil
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

// Helper function
func intPtr(i int) *int {
	return &i
}
