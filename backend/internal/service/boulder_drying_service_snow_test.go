package service

import (
	"context"
	"testing"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/alexscott64/woulder/backend/internal/weather/client"
)

// mockWeatherClient provides mock weather data for testing
type mockWeatherClient struct {
	currentWeather  *models.WeatherData
	forecastWeather []models.WeatherData
}

func (m *mockWeatherClient) GetCurrentAndForecast(lat, lon float64) (*models.WeatherData, []models.WeatherData, *client.SunTimes, error) {
	return m.currentWeather, m.forecastWeather, nil, nil
}

// TestGetLocationRockDryingStatus_WithSnow verifies snow depth is calculated and passed to rock drying
func TestGetLocationRockDryingStatus_WithSnow(t *testing.T) {
	now := time.Now()
	locationID := 15 // Tramway

	location := &models.Location{
		ID:          locationID,
		Name:        "Tramway",
		Latitude:    33.8,
		Longitude:   -116.6,
		ElevationFt: 8500, // High elevation for snow
	}

	currentWeather := &models.WeatherData{
		LocationID:    locationID,
		Timestamp:     now,
		Temperature:   30.0, // Freezing
		Humidity:      80,
		WindSpeed:     10.0,
		Precipitation: 0.5, // Active snow
		Description:   "Slight snow",
	}

	historicalWeather := []models.WeatherData{}
	// Add historical weather with snow accumulation
	for i := 168; i > 0; i-- {
		timestamp := now.Add(-time.Duration(i) * time.Hour)
		temp := 28.0
		precip := 0.0
		desc := "Clear"

		// Add snow 24-48h ago
		if i >= 24 && i <= 48 {
			precip = 0.3 // Snow
			desc = "Snow"
		}

		historicalWeather = append(historicalWeather, models.WeatherData{
			LocationID:    locationID,
			Timestamp:     timestamp,
			Temperature:   temp,
			Humidity:      70,
			WindSpeed:     5.0,
			Precipitation: precip,
			Description:   desc,
		})
	}

	// Add forecast weather (needed for snow calculation)
	forecastWeather := []models.WeatherData{}
	for i := 1; i <= 144; i++ {
		timestamp := now.Add(time.Duration(i) * time.Hour)
		forecastWeather = append(forecastWeather, models.WeatherData{
			LocationID:    locationID,
			Timestamp:     timestamp,
			Temperature:   32.0,
			Humidity:      60,
			WindSpeed:     5.0,
			Precipitation: 0.0,
			Description:   "Clear",
		})
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
			return location, nil
		},
	}
	mockRocksRepo := &MockRocksRepository{
		GetRockTypesByLocationFn: func(ctx context.Context, locID int) ([]models.RockType, error) {
			return []models.RockType{
				{Name: "Tonalite", BaseDryingHours: 12.0, GroupName: "Granite"},
			}, nil
		},
		GetSunExposureByLocationFn: func(ctx context.Context, locID int) (*models.LocationSunExposure, error) {
			return &models.LocationSunExposure{
				TreeCoveragePercent: 30.0,
			}, nil
		},
	}
	mockMPRepo := NewMockMountainProjectRepository()

	// Create mock weather client with the same data
	mockWeatherClient := &mockWeatherClient{
		currentWeather:  currentWeather,
		forecastWeather: forecastWeather,
	}

	service := NewBoulderDryingService(mockBouldersRepo, mockWeatherRepo, mockLocationsRepo, mockRocksRepo, mockMPRepo, mockWeatherClient)
	dryingStatus, _, err := service.getLocationRockDryingStatus(context.Background(), locationID)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if dryingStatus == nil {
		t.Fatal("Expected drying status, got nil")
	}

	// CRITICAL: With snow accumulation and active precipitation, location MUST show wet
	if !dryingStatus.IsWet {
		t.Errorf("CRITICAL BUG: Location shows DRY with active snow (HoursUntilDry=%.2f)",
			dryingStatus.HoursUntilDry)
	}

	// Should have significant hours until dry due to snow
	if dryingStatus.HoursUntilDry < 24.0 {
		t.Errorf("Expected HoursUntilDry >= 24 with snow, got %.2f", dryingStatus.HoursUntilDry)
	}

	t.Logf("SUCCESS: Location correctly shows wet with %.2f hours until dry", dryingStatus.HoursUntilDry)
	t.Logf("Message: %s", dryingStatus.Message)
}
