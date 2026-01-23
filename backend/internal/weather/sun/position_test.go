package sun

import (
	"math"
	"testing"
	"time"
)

func TestCalculate(t *testing.T) {
	// Test for Seattle area on summer solstice at noon
	// Seattle: 47.6062°N, 122.3321°W
	lat := 47.6062
	lon := -122.3321

	// June 21, 2024, 12:00 PM PDT (19:00 UTC)
	testTime := time.Date(2024, 6, 21, 19, 0, 0, 0, time.UTC)

	pos := Calculate(lat, lon, testTime)

	// On summer solstice at noon, sun should be:
	// - High in the sky (elevation 60-70°)
	// - Roughly south (azimuth ~180°)
	if pos.Elevation < 60.0 || pos.Elevation > 70.0 {
		t.Errorf("Expected elevation 60-70°, got %.2f°", pos.Elevation)
	}

	if pos.Azimuth < 170.0 || pos.Azimuth > 190.0 {
		t.Errorf("Expected azimuth ~180° (south), got %.2f°", pos.Azimuth)
	}

	t.Logf("Summer solstice noon: Elevation=%.2f°, Azimuth=%.2f°", pos.Elevation, pos.Azimuth)
}

func TestCalculate_WinterSolstice(t *testing.T) {
	// Test for Seattle area on winter solstice at noon
	lat := 47.6062
	lon := -122.3321

	// December 21, 2024, 12:00 PM PST (20:00 UTC)
	testTime := time.Date(2024, 12, 21, 20, 0, 0, 0, time.UTC)

	pos := Calculate(lat, lon, testTime)

	// On winter solstice at noon, sun should be:
	// - Lower in the sky (elevation 15-25°)
	// - Roughly south (azimuth ~180°)
	if pos.Elevation < 15.0 || pos.Elevation > 25.0 {
		t.Errorf("Expected elevation 15-25°, got %.2f°", pos.Elevation)
	}

	if pos.Azimuth < 170.0 || pos.Azimuth > 190.0 {
		t.Errorf("Expected azimuth ~180° (south), got %.2f°", pos.Azimuth)
	}

	t.Logf("Winter solstice noon: Elevation=%.2f°, Azimuth=%.2f°", pos.Elevation, pos.Azimuth)
}

func TestCalculate_Sunset(t *testing.T) {
	// Test for Seattle area at sunset
	lat := 47.6062
	lon := -122.3321

	// June 21, 2024, 9:00 PM PDT (04:00 UTC next day)
	testTime := time.Date(2024, 6, 22, 4, 0, 0, 0, time.UTC)

	pos := Calculate(lat, lon, testTime)

	// At sunset:
	// - Elevation should be close to 0° (slightly negative)
	// - Azimuth should be northwest (~280-310°)
	if pos.Elevation < -5.0 || pos.Elevation > 5.0 {
		t.Errorf("Expected elevation near 0°, got %.2f°", pos.Elevation)
	}

	if pos.Azimuth < 280.0 || pos.Azimuth > 310.0 {
		t.Errorf("Expected azimuth 280-310° (NW), got %.2f°", pos.Azimuth)
	}

	t.Logf("Sunset: Elevation=%.2f°, Azimuth=%.2f°", pos.Elevation, pos.Azimuth)
}

func TestHourlyPositions(t *testing.T) {
	lat := 47.6062
	lon := -122.3321
	startTime := time.Date(2024, 6, 21, 12, 0, 0, 0, time.UTC)

	positions := HourlyPositions(lat, lon, startTime, 12)

	if len(positions) != 12 {
		t.Errorf("Expected 12 positions, got %d", len(positions))
	}

	// Check that elevation changes over time (sun moves)
	firstElevation := positions[0].Elevation
	lastElevation := positions[11].Elevation

	if math.Abs(firstElevation-lastElevation) < 10.0 {
		t.Errorf("Expected significant elevation change over 12 hours, got %.2f° to %.2f°",
			firstElevation, lastElevation)
	}

	t.Logf("First elevation: %.2f°, Last elevation: %.2f°", firstElevation, lastElevation)
}

func TestAspectToDegrees(t *testing.T) {
	tests := []struct {
		aspect   string
		expected float64
	}{
		{"North", 0.0},
		{"N", 0.0},
		{"East", 90.0},
		{"E", 90.0},
		{"South", 180.0},
		{"S", 180.0},
		{"West", 270.0},
		{"W", 270.0},
		{"North-East", 45.0},
		{"NE", 45.0},
		{"South-West", 225.0},
		{"SW", 225.0},
		{"Unknown", 180.0}, // Default to south
	}

	for _, tt := range tests {
		result := aspectToDegrees(tt.aspect)
		if result != tt.expected {
			t.Errorf("aspectToDegrees(%s) = %.1f, want %.1f", tt.aspect, result, tt.expected)
		}
	}
}

func TestAngleDifference(t *testing.T) {
	tests := []struct {
		a1       float64
		a2       float64
		expected float64
	}{
		{0, 0, 0},
		{0, 90, 90},
		{90, 0, 90},
		{0, 180, 180},
		{0, 270, 90},  // Wraps around
		{350, 10, 20}, // Wraps around
		{10, 350, 20}, // Wraps around
	}

	for _, tt := range tests {
		result := angleDifference(tt.a1, tt.a2)
		if math.Abs(result-tt.expected) > 0.1 {
			t.Errorf("angleDifference(%.1f, %.1f) = %.1f, want %.1f",
				tt.a1, tt.a2, result, tt.expected)
		}
	}
}

func TestCalculateSunExposure(t *testing.T) {
	lat := 47.6062
	lon := -122.3321
	startTime := time.Date(2024, 6, 21, 12, 0, 0, 0, time.UTC)

	// Test south-facing boulder with no tree coverage
	sunHours := CalculateSunExposure(lat, lon, "South", 0.0, startTime, 144)

	// South-facing boulder should get substantial sun
	if sunHours < 50.0 {
		t.Errorf("Expected south-facing boulder to get >50 sun hours, got %.2f", sunHours)
	}

	t.Logf("South-facing (0%% trees): %.2f sun hours", sunHours)

	// Test north-facing boulder with no tree coverage
	sunHoursNorth := CalculateSunExposure(lat, lon, "North", 0.0, startTime, 144)

	// North-facing should get less sun than south-facing
	if sunHoursNorth >= sunHours {
		t.Errorf("Expected north-facing to get less sun than south-facing")
	}

	t.Logf("North-facing (0%% trees): %.2f sun hours", sunHoursNorth)

	// Test south-facing boulder with 75% tree coverage
	sunHoursWithTrees := CalculateSunExposure(lat, lon, "South", 75.0, startTime, 144)

	// Trees should reduce sun exposure
	if sunHoursWithTrees >= sunHours*0.5 {
		t.Errorf("Expected 75%% tree coverage to significantly reduce sun hours")
	}

	t.Logf("South-facing (75%% trees): %.2f sun hours", sunHoursWithTrees)
}

func TestIsAboveHorizon(t *testing.T) {
	tests := []struct {
		elevation float64
		expected  bool
	}{
		{10.0, true},
		{0.1, true},
		{0.0, false},
		{-0.1, false},
		{-10.0, false},
	}

	for _, tt := range tests {
		pos := Position{Elevation: tt.elevation}
		result := pos.IsAboveHorizon()
		if result != tt.expected {
			t.Errorf("IsAboveHorizon(%.1f) = %v, want %v", tt.elevation, result, tt.expected)
		}
	}
}

func TestGetSunriseAndSunset(t *testing.T) {
	lat := 47.6062
	lon := -122.3321

	// Summer solstice
	date := time.Date(2024, 6, 21, 12, 0, 0, 0, time.UTC)
	sunrise, sunset := GetSunriseAndSunset(lat, lon, date)

	// Check that sunrise is before sunset
	if !sunrise.Before(sunset) {
		t.Errorf("Sunrise should be before sunset")
	}

	// Check that sunrise is in the morning hours (roughly 4-7 AM UTC for Seattle in summer)
	sunriseHour := sunrise.Hour()
	if sunriseHour < 10 || sunriseHour > 14 {
		t.Errorf("Expected sunrise around 10-14 UTC, got %d:00", sunriseHour)
	}

	t.Logf("Sunrise: %s, Sunset: %s", sunrise.Format("15:04"), sunset.Format("15:04"))
}
