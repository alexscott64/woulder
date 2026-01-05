package rock_drying

import (
	"testing"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

func TestEstimateSnowMeltTime_WarmWeather(t *testing.T) {
	rockType := models.RockType{
		Name:            "Granite",
		BaseDryingHours: 6.0,
		PorosityPercent: 1.0,
	}

	currentWeather := &models.WeatherData{
		Temperature: 50.0, // Above freezing
		WindSpeed:   10.0,
		Timestamp:   time.Now(),
	}

	// Test with 2 inches of snow in warm weather
	hoursToMelt := estimateSnowMeltTime(2.0, currentWeather, []models.WeatherData{}, rockType, nil)

	// Should be less than 100 hours (warm weather melts relatively fast)
	if hoursToMelt > 100 {
		t.Errorf("Expected snow to melt in < 100 hours at 50°F, got %.2f hours", hoursToMelt)
	}

	// Should be more than rock drying time (snow must melt first)
	if hoursToMelt < rockType.BaseDryingHours {
		t.Errorf("Expected snow melt time (%.2f) to be > rock drying time (%.2f)", hoursToMelt, rockType.BaseDryingHours)
	}

	t.Logf("Snow melt time at 50°F with 2\" snow: %.2f hours", hoursToMelt)
}

func TestEstimateSnowMeltTime_FreezingWithWarmingTrend(t *testing.T) {
	rockType := models.RockType{
		Name:            "Granite",
		BaseDryingHours: 6.0,
		PorosityPercent: 1.0,
	}

	currentWeather := &models.WeatherData{
		Temperature: 30.0, // Below freezing
		WindSpeed:   5.0,
		Timestamp:   time.Now(),
	}

	// Historical shows warming trend (average > 34°F)
	historical := []models.WeatherData{
		{Temperature: 38.0, Timestamp: time.Now().Add(-1 * time.Hour)},
		{Temperature: 36.0, Timestamp: time.Now().Add(-2 * time.Hour)},
		{Temperature: 35.0, Timestamp: time.Now().Add(-3 * time.Hour)},
	}

	hoursToMelt := estimateSnowMeltTime(1.0, currentWeather, historical, rockType, nil)

	// With warming trend, should use average temp and be reasonable
	if hoursToMelt > 150 {
		t.Errorf("Expected reasonable melt time with warming trend, got %.2f hours", hoursToMelt)
	}

	t.Logf("Snow melt time with warming trend: %.2f hours", hoursToMelt)
}

func TestEstimateSnowMeltTime_FreezingWinter(t *testing.T) {
	rockType := models.RockType{
		Name:            "Granite",
		BaseDryingHours: 6.0,
		PorosityPercent: 1.0,
	}

	currentWeather := &models.WeatherData{
		Temperature: 28.0, // Below freezing
		WindSpeed:   5.0,
		Timestamp:   time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC), // January (winter)
	}

	// No warming trend
	historical := []models.WeatherData{
		{Temperature: 30.0, Timestamp: time.Now().Add(-1 * time.Hour)},
		{Temperature: 28.0, Timestamp: time.Now().Add(-2 * time.Hour)},
	}

	hoursToMelt := estimateSnowMeltTime(3.0, currentWeather, historical, rockType, nil)

	// Winter + no warming = long estimate (1-2 weeks base + 36h per inch)
	expectedMin := 168.0 + (3.0 * 36.0) // 276 hours minimum
	if hoursToMelt < expectedMin {
		t.Errorf("Expected winter snow to persist > %.2f hours, got %.2f", expectedMin, hoursToMelt)
	}

	t.Logf("Snow melt time in winter (3\"): %.2f hours (%.1f days)", hoursToMelt, hoursToMelt/24)
}

func TestEstimateSnowMeltTime_FreezingSummer(t *testing.T) {
	rockType := models.RockType{
		Name:            "Granite",
		BaseDryingHours: 6.0,
		PorosityPercent: 1.0,
	}

	currentWeather := &models.WeatherData{
		Temperature: 30.0, // Below freezing (unusual for summer)
		WindSpeed:   5.0,
		Timestamp:   time.Date(2026, 7, 15, 10, 0, 0, 0, time.UTC), // July (summer)
	}

	// No warming trend
	historical := []models.WeatherData{
		{Temperature: 32.0, Timestamp: time.Now().Add(-1 * time.Hour)},
		{Temperature: 30.0, Timestamp: time.Now().Add(-2 * time.Hour)},
	}

	hoursToMelt := estimateSnowMeltTime(1.0, currentWeather, historical, rockType, nil)

	// Summer snow melts fast even if cold (48h base + 12h per inch)
	expectedMax := 48.0 + (1.0 * 12.0) // 60 hours maximum
	if hoursToMelt > expectedMax*1.1 { // Allow 10% margin
		t.Errorf("Expected summer snow to melt within ~%.2f hours, got %.2f", expectedMax, hoursToMelt)
	}

	t.Logf("Snow melt time in summer (1\"): %.2f hours", hoursToMelt)
}

func TestEstimateSnowMeltTime_SunExposure(t *testing.T) {
	rockType := models.RockType{
		Name:            "Granite",
		BaseDryingHours: 6.0,
		PorosityPercent: 1.0,
	}

	currentWeather := &models.WeatherData{
		Temperature: 40.0,
		WindSpeed:   8.0,
		Timestamp:   time.Now(),
	}

	// South-facing with minimal tree cover
	sunExposure := &models.LocationSunExposure{
		SouthFacingPercent:  80.0,
		WestFacingPercent:   10.0,
		EastFacingPercent:   5.0,
		NorthFacingPercent:  5.0,
		TreeCoveragePercent: 10.0,
	}

	hoursWithSun := estimateSnowMeltTime(2.0, currentWeather, []models.WeatherData{}, rockType, sunExposure)
	hoursWithoutSun := estimateSnowMeltTime(2.0, currentWeather, []models.WeatherData{}, rockType, nil)

	// Sun exposure should reduce melt time
	if hoursWithSun >= hoursWithoutSun {
		t.Errorf("Expected sun exposure to reduce melt time. With sun: %.2f, Without: %.2f", hoursWithSun, hoursWithoutSun)
	}

	t.Logf("Snow melt with sun: %.2f hours, without sun: %.2f hours", hoursWithSun, hoursWithoutSun)
}

func TestEstimateSnowMeltTime_RockType(t *testing.T) {
	currentWeather := &models.WeatherData{
		Temperature: 45.0,
		WindSpeed:   8.0,
		Timestamp:   time.Now(),
	}

	// Dark granite (low porosity, absorbs heat)
	granite := models.RockType{
		Name:            "Granite",
		BaseDryingHours: 6.0,
		PorosityPercent: 1.0, // Very low porosity
	}

	// Porous sandstone (high porosity, reflects more light)
	sandstone := models.RockType{
		Name:            "Sandstone",
		BaseDryingHours: 12.0,
		PorosityPercent: 15.0, // High porosity
	}

	hoursGranite := estimateSnowMeltTime(2.0, currentWeather, []models.WeatherData{}, granite, nil)
	hoursSandstone := estimateSnowMeltTime(2.0, currentWeather, []models.WeatherData{}, sandstone, nil)

	// Granite should melt snow faster than sandstone
	if hoursGranite >= hoursSandstone {
		t.Errorf("Expected granite to melt snow faster. Granite: %.2f, Sandstone: %.2f", hoursGranite, hoursSandstone)
	}

	t.Logf("Granite melt time: %.2f hours, Sandstone melt time: %.2f hours", hoursGranite, hoursSandstone)
}
