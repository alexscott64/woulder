package service

import (
	"testing"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// TestCalculateBoulderRockTempStatus_RockTypeOverride verifies that when a
// boulder profile sets rock_type_override, the rock_temp calculator receives
// the overridden rock type group instead of the location-level default.
func TestCalculateBoulderRockTempStatus_RockTypeOverride(t *testing.T) {
	now := time.Now().UTC()

	wctx := newRockTempTestContext(now)

	override := "Sandstone"
	profile := &models.BoulderDryingProfile{
		MPRouteID:        9001,
		RockTypeOverride: &override,
	}

	svc := newRockTempTestService()
	status := svc.calculateBoulderRockTempStatus(wctx, profile)

	if status == nil {
		t.Fatalf("expected non-nil RockTemperatureStatus when rock types are present")
	}
	if status.RockType != "Sandstone" {
		t.Errorf("expected RockType to honor override 'Sandstone', got %q", status.RockType)
	}
}

// TestCalculateBoulderRockTempStatus_NoOverride_UsesLocationDefault verifies
// the location's primary rock type group is used when no override is set.
func TestCalculateBoulderRockTempStatus_NoOverride_UsesLocationDefault(t *testing.T) {
	now := time.Now().UTC()

	wctx := newRockTempTestContext(now)

	svc := newRockTempTestService()
	// Profile is nil — no overrides at all.
	status := svc.calculateBoulderRockTempStatus(wctx, nil)

	if status == nil {
		t.Fatalf("expected non-nil RockTemperatureStatus when rock types are present")
	}
	if status.RockType != "Granite" {
		t.Errorf("expected RockType to be location default 'Granite', got %q", status.RockType)
	}
}

// TestCalculateBoulderRockTempStatus_TreeCoverageOverride verifies that when
// the boulder profile sets tree_coverage_percent, a copy of the location's
// SunExposure is built and the calculator sees the boulder-specific value
// (without mutating the cached/shared location pointer).
func TestCalculateBoulderRockTempStatus_TreeCoverageOverride(t *testing.T) {
	now := time.Now().UTC()
	wctx := newRockTempTestContext(now)

	originalTree := wctx.sunExposure.TreeCoveragePercent

	tree := 90.0
	profile := &models.BoulderDryingProfile{
		MPRouteID:           9001,
		TreeCoveragePercent: &tree,
	}

	svc := newRockTempTestService()
	status := svc.calculateBoulderRockTempStatus(wctx, profile)
	if status == nil {
		t.Fatalf("expected non-nil RockTemperatureStatus")
	}

	// Critical: the cached/shared location SunExposure must not be mutated.
	if wctx.sunExposure.TreeCoveragePercent != originalTree {
		t.Errorf("location sun exposure tree coverage was mutated: got %v, want %v",
			wctx.sunExposure.TreeCoveragePercent, originalTree)
	}
}

// TestCalculateBoulderRockTempStatus_NoRockTypes_ReturnsNil verifies the
// function skips the calculation entirely when the location has no rock
// type data (matches weather_service.calculateRockTempStatus behavior).
func TestCalculateBoulderRockTempStatus_NoRockTypes_ReturnsNil(t *testing.T) {
	now := time.Now().UTC()
	wctx := newRockTempTestContext(now)
	wctx.rockTypes = nil // simulate location with no rock-type rows

	svc := newRockTempTestService()
	status := svc.calculateBoulderRockTempStatus(wctx, nil)
	if status != nil {
		t.Errorf("expected nil status when location has no rock types, got %+v", status)
	}
}

// --- helpers --------------------------------------------------------------

func newRockTempTestService() *BoulderDryingService {
	return NewBoulderDryingService(
		&MockBouldersRepository{},
		&MockWeatherRepository{},
		&MockLocationsRepository{},
		&MockRocksRepository{},
		NewMockMountainProjectRepository(),
		&mockBoulderWeatherClient{},
	)
}

// newRockTempTestContext builds a locationWeatherContext with enough weather
// data and rock type info that the rock_temp Calculator will produce a
// non-degraded response. Defaults: location-level rock = Granite, tree
// coverage = 30%.
func newRockTempTestContext(now time.Time) *locationWeatherContext {
	location := &models.Location{
		ID:          1,
		Name:        "Test Location",
		Latitude:    47.6,
		Longitude:   -120.9,
		ElevationFt: 1000,
	}

	current := &models.WeatherData{
		LocationID:    location.ID,
		Timestamp:     now,
		Temperature:   60.0,
		Humidity:      40.0,
		WindSpeed:     5.0,
		DewpointF:     35.0,
		Precipitation: 0.0,
		CloudCover:    20,
	}

	// 12h of past hours for thermal-lag spin-up.
	historical := make([]models.WeatherData, 0, 12)
	for i := 12; i >= 1; i-- {
		historical = append(historical, models.WeatherData{
			LocationID:  location.ID,
			Timestamp:   now.Add(-time.Duration(i) * time.Hour),
			Temperature: 55.0,
			Humidity:    45.0,
			WindSpeed:   5.0,
			DewpointF:   34.0,
			CloudCover:  30,
		})
	}

	// 24h of future forecast.
	forecast := make([]models.WeatherData, 0, 24)
	for i := 1; i <= 24; i++ {
		forecast = append(forecast, models.WeatherData{
			LocationID:  location.ID,
			Timestamp:   now.Add(time.Duration(i) * time.Hour),
			Temperature: 62.0,
			Humidity:    35.0,
			WindSpeed:   6.0,
			DewpointF:   33.0,
			CloudCover:  20,
		})
	}

	rockTypes := []models.RockType{
		{
			Name:            "Granite",
			BaseDryingHours: 12.0,
			GroupName:       "Granite",
		},
	}

	sunExposure := &models.LocationSunExposure{
		LocationID:          location.ID,
		SouthFacingPercent:  80,
		TreeCoveragePercent: 30,
	}

	return &locationWeatherContext{
		location:          location,
		current:           current,
		historicalWeather: historical,
		hourlyForecast:    forecast,
		rockTypes:         rockTypes,
		sunExposure:       sunExposure,
	}
}
