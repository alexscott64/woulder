package calculator

import (
	"testing"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// TestSnowAccumulation_MoneyCreekScenario tests the real-world scenario reported by the user
// User visited Money Creek on Dec 30, 2024 and observed:
// - 2 inches of snow on the ground
// - Icy roads
// But the app showed 0 inches of snow
func TestSnowAccumulation_MoneyCreekScenario(t *testing.T) {
	// Money Creek elevation: ~1000 ft
	elevationFt := 1000.0

	// Simulate historical weather leading up to Dec 30, 2024
	// User saw 2 inches on ground - this suggests modest snowfall
	// Typical snow-to-liquid ratio: 10:1, so 2 inches snow = 0.2 inches liquid
	baseTime := time.Date(2024, 12, 27, 0, 0, 0, 0, time.UTC)

	historical := []models.WeatherData{
		// Dec 27 - Clear and cold
		{Timestamp: baseTime, Temperature: 35, Precipitation: 0, WindSpeed: 5, Humidity: 60, CloudCover: 10},
		{Timestamp: baseTime.Add(3 * time.Hour), Temperature: 34, Precipitation: 0, WindSpeed: 5, Humidity: 60, CloudCover: 10},
		{Timestamp: baseTime.Add(6 * time.Hour), Temperature: 32, Precipitation: 0, WindSpeed: 5, Humidity: 60, CloudCover: 20},

		// Dec 28 - Light snow event
		{Timestamp: baseTime.Add(24 * time.Hour), Temperature: 30, Precipitation: 0.1, WindSpeed: 8, Humidity: 80, CloudCover: 90},
		{Timestamp: baseTime.Add(27 * time.Hour), Temperature: 28, Precipitation: 0.15, WindSpeed: 10, Humidity: 85, CloudCover: 100},
		{Timestamp: baseTime.Add(30 * time.Hour), Temperature: 29, Precipitation: 0.1, WindSpeed: 12, Humidity: 85, CloudCover: 100},

		// Dec 29 - Clearing, cold temps preserve snow
		{Timestamp: baseTime.Add(48 * time.Hour), Temperature: 26, Precipitation: 0, WindSpeed: 8, Humidity: 75, CloudCover: 50},
		{Timestamp: baseTime.Add(51 * time.Hour), Temperature: 24, Precipitation: 0, WindSpeed: 5, Humidity: 70, CloudCover: 30},
		{Timestamp: baseTime.Add(54 * time.Hour), Temperature: 25, Precipitation: 0, WindSpeed: 5, Humidity: 65, CloudCover: 20},
		{Timestamp: baseTime.Add(57 * time.Hour), Temperature: 26, Precipitation: 0, WindSpeed: 5, Humidity: 60, CloudCover: 10},
	}

	// Dec 30 - Current conditions (when user visited)
	currentData := []models.WeatherData{
		{Timestamp: baseTime.Add(72 * time.Hour), Temperature: 28, Precipitation: 0, WindSpeed: 5, Humidity: 60, CloudCover: 20}, // Clear, cold
	}

	// Calculate snow depth
	snowDepth := GetCurrentSnowDepth(historical, currentData, elevationFt)

	// User reported 2 inches of snow on ground
	// With ~0.35 inches of precipitation at 28-30°F, we should see 1-4 inches of settled snow
	if snowDepth < 0.5 {
		t.Errorf("Expected at least 0.5 inch of snow (user saw 2 inches), got %.2f inches", snowDepth)
	}

	if snowDepth > 6.0 {
		t.Errorf("Expected reasonable snow depth (<6 inches), got %.2f inches", snowDepth)
	}

	t.Logf("Money Creek scenario: %.2f inches of snow on ground (user reported 2 inches)", snowDepth)
}

// TestSnowAccumulation_WarmWeatherMelt tests that snow melts when temperatures rise
func TestSnowAccumulation_WarmWeatherMelt(t *testing.T) {
	baseTime := time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)
	elevationFt := 1000.0

	historical := []models.WeatherData{
		// Day 1: Heavy snowfall
		{Timestamp: baseTime, Temperature: 28, Precipitation: 1.0, WindSpeed: 5, Humidity: 80, CloudCover: 100},
		{Timestamp: baseTime.Add(3 * time.Hour), Temperature: 28, Precipitation: 0.5, WindSpeed: 5, Humidity: 80, CloudCover: 100},

		// Day 2: Warming trend - snow should melt
		{Timestamp: baseTime.Add(24 * time.Hour), Temperature: 40, Precipitation: 0, WindSpeed: 10, Humidity: 50, CloudCover: 30},
		{Timestamp: baseTime.Add(27 * time.Hour), Temperature: 45, Precipitation: 0, WindSpeed: 10, Humidity: 40, CloudCover: 20},
		{Timestamp: baseTime.Add(30 * time.Hour), Temperature: 50, Precipitation: 0, WindSpeed: 10, Humidity: 35, CloudCover: 10},
	}

	currentData := []models.WeatherData{
		{Timestamp: baseTime.Add(48 * time.Hour), Temperature: 50, Precipitation: 0, WindSpeed: 5, Humidity: 40, CloudCover: 10},
	}

	snowDepth := GetCurrentSnowDepth(historical, currentData, elevationFt)

	// Snow should have mostly or completely melted
	if snowDepth > 2.0 {
		t.Errorf("Expected snow to melt significantly at 50°F, got %.2f inches remaining", snowDepth)
	}

	t.Logf("After warming to 50°F: %.2f inches of snow remaining", snowDepth)
}

// TestSnowAccumulation_RainOnSnow tests rain-on-snow compaction and melt
func TestSnowAccumulation_RainOnSnow(t *testing.T) {
	baseTime := time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)
	elevationFt := 1000.0

	historical := []models.WeatherData{
		// Day 1: Snowfall
		{Timestamp: baseTime, Temperature: 28, Precipitation: 1.0, WindSpeed: 5, Humidity: 80, CloudCover: 100},
		{Timestamp: baseTime.Add(3 * time.Hour), Temperature: 28, Precipitation: 0.5, WindSpeed: 5, Humidity: 80, CloudCover: 100},

		// Day 2: Rain on snow (32-34°F transition zone)
		{Timestamp: baseTime.Add(24 * time.Hour), Temperature: 36, Precipitation: 0.5, WindSpeed: 5, Humidity: 90, CloudCover: 100}, // Rain
		{Timestamp: baseTime.Add(27 * time.Hour), Temperature: 38, Precipitation: 0.3, WindSpeed: 5, Humidity: 85, CloudCover: 100}, // Rain
	}

	currentData := []models.WeatherData{
		{Timestamp: baseTime.Add(48 * time.Hour), Temperature: 35, Precipitation: 0, WindSpeed: 5, Humidity: 70, CloudCover: 50},
	}

	// Calculate snow before and after rain
	snowDepthBeforeRain := GetCurrentSnowDepth(historical[:2], []models.WeatherData{historical[2]}, elevationFt)
	snowDepthAfterRain := GetCurrentSnowDepth(historical, currentData, elevationFt)

	// Rain should compact and melt some snow
	if snowDepthAfterRain >= snowDepthBeforeRain {
		t.Errorf("Expected rain to reduce snow depth, but got before=%.2f, after=%.2f", snowDepthBeforeRain, snowDepthAfterRain)
	}

	t.Logf("Before rain: %.2f inches, After rain: %.2f inches", snowDepthBeforeRain, snowDepthAfterRain)
}

// TestSnowAccumulation_FreezingLevelTransition tests the 30-34°F transition zone
func TestSnowAccumulation_FreezingLevelTransition(t *testing.T) {
	baseTime := time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)
	elevationFt := 1000.0

	// Test different temperatures in transition zone
	testCases := []struct {
		temp     float64
		precip   float64
		expected string
	}{
		{28, 1.0, "all snow"},        // Below 30°F = all snow
		{30, 1.0, "all snow"},        // At 30°F = all snow
		{32, 1.0, "mostly snow"},     // Mid-transition = mixed
		{34, 1.0, "all rain"},        // At 34°F = all rain
		{36, 1.0, "all rain"},        // Above 34°F = all rain
	}

	for _, tc := range testCases {
		historical := []models.WeatherData{
			{Timestamp: baseTime, Temperature: tc.temp, Precipitation: tc.precip, WindSpeed: 5, Humidity: 80, CloudCover: 100},
		}

		currentData := []models.WeatherData{
			{Timestamp: baseTime.Add(3 * time.Hour), Temperature: tc.temp - 5, Precipitation: 0, WindSpeed: 5, Humidity: 70, CloudCover: 50},
		}

		snowDepth := GetCurrentSnowDepth(historical, currentData, elevationFt)

		t.Logf("Temp %.0f°F with %.1f\" precip: %.2f\" snow depth (%s)", tc.temp, tc.precip, snowDepth, tc.expected)
	}
}

// TestSnowAccumulation_ElevationAdjustment tests temperature lapse rate
func TestSnowAccumulation_ElevationAdjustment(t *testing.T) {
	baseTime := time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)

	// Same weather conditions at different elevations
	historical := []models.WeatherData{
		{Timestamp: baseTime, Temperature: 36, Precipitation: 1.0, WindSpeed: 5, Humidity: 80, CloudCover: 100}, // 36°F at base
	}

	currentData := []models.WeatherData{
		{Timestamp: baseTime.Add(3 * time.Hour), Temperature: 35, Precipitation: 0, WindSpeed: 5, Humidity: 70, CloudCover: 50},
	}

	// At sea level (0 ft): 36°F = rain
	snowDepthSeaLevel := GetCurrentSnowDepth(historical, currentData, 0)

	// At 2000 ft: 36 - (2000/1000 * 3.5) = 36 - 7 = 29°F = snow
	snowDepth2000ft := GetCurrentSnowDepth(historical, currentData, 2000)

	// At 4000 ft: 36 - (4000/1000 * 3.5) = 36 - 14 = 22°F = snow
	snowDepth4000ft := GetCurrentSnowDepth(historical, currentData, 4000)

	if snowDepth2000ft <= snowDepthSeaLevel {
		t.Errorf("Expected more snow at higher elevation, got sea_level=%.2f, 2000ft=%.2f", snowDepthSeaLevel, snowDepth2000ft)
	}

	if snowDepth4000ft <= snowDepth2000ft {
		t.Errorf("Expected more snow at 4000ft than 2000ft, got 2000ft=%.2f, 4000ft=%.2f", snowDepth2000ft, snowDepth4000ft)
	}

	t.Logf("Sea level: %.2f\", 2000ft: %.2f\", 4000ft: %.2f\"", snowDepthSeaLevel, snowDepth2000ft, snowDepth4000ft)
}

// TestSnowAccumulation_NoSnowScenario tests that algorithm returns 0 when there's no snow
func TestSnowAccumulation_NoSnowScenario(t *testing.T) {
	baseTime := time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)
	elevationFt := 1000.0

	// Warm weather with rain (no snow)
	historical := []models.WeatherData{
		{Timestamp: baseTime, Temperature: 50, Precipitation: 0.5, WindSpeed: 10, Humidity: 80, CloudCover: 100},
		{Timestamp: baseTime.Add(3 * time.Hour), Temperature: 52, Precipitation: 0.3, WindSpeed: 10, Humidity: 75, CloudCover: 90},
	}

	currentData := []models.WeatherData{
		{Timestamp: baseTime.Add(6 * time.Hour), Temperature: 48, Precipitation: 0, WindSpeed: 5, Humidity: 70, CloudCover: 50},
	}

	snowDepth := GetCurrentSnowDepth(historical, currentData, elevationFt)

	if snowDepth > 0.1 {
		t.Errorf("Expected no snow with warm temps, got %.2f inches", snowDepth)
	}

	t.Logf("No snow scenario: %.2f inches (expected ~0)", snowDepth)
}
