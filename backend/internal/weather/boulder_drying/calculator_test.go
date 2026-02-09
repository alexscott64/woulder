package boulder_drying

import (
	"context"
	"testing"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

func TestCalculateBoulderDryingTime(t *testing.T) {
	calc := NewCalculator("test-api-key")

	tests := []struct {
		name                  string
		baseDryingTime        float64
		sunExposureHours      float64
		treeCoveragePercent   float64
		expectedModifier      float64
		expectedModifierRange [2]float64 // min, max
	}{
		{
			name:                  "Dry location",
			baseDryingTime:        0.0,
			sunExposureHours:      48.0,
			treeCoveragePercent:   0.0,
			expectedModifier:      0.0,
			expectedModifierRange: [2]float64{0.0, 0.0},
		},
		{
			name:                  "Exceptional sun, no trees",
			baseDryingTime:        24.0,
			sunExposureHours:      48.0, // 8 hours/day
			treeCoveragePercent:   0.0,
			expectedModifier:      0.7, // -30% drying time
			expectedModifierRange: [2]float64{16.0, 17.0},
		},
		{
			name:                  "Good sun, light trees",
			baseDryingTime:        24.0,
			sunExposureHours:      36.0, // 6 hours/day
			treeCoveragePercent:   30.0,
			expectedModifier:      0.85 * 1.05, // -15% sun, +5% trees
			expectedModifierRange: [2]float64{21.0, 22.0},
		},
		{
			name:                  "Poor sun, heavy trees",
			baseDryingTime:        24.0,
			sunExposureHours:      12.0, // 2 hours/day
			treeCoveragePercent:   80.0,
			expectedModifier:      1.15 * 1.3, // +15% sun, +30% trees
			expectedModifierRange: [2]float64{35.0, 37.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			locationDrying := &models.RockDryingStatus{
				IsWet:         tt.baseDryingTime > 0,
				HoursUntilDry: tt.baseDryingTime,
			}

			boulderStatus := &BoulderDryingStatus{
				SunExposureHours:    tt.sunExposureHours,
				TreeCoveragePercent: tt.treeCoveragePercent,
			}

			result := calc.calculateBoulderDryingTime(locationDrying, boulderStatus)

			if tt.expectedModifierRange[0] == 0 && tt.expectedModifierRange[1] == 0 {
				if result != 0 {
					t.Errorf("Expected dry (0h), got %.1fh", result)
				}
			} else {
				if result < tt.expectedModifierRange[0] || result > tt.expectedModifierRange[1] {
					t.Errorf("Expected %.1f-%.1fh, got %.1fh",
						tt.expectedModifierRange[0], tt.expectedModifierRange[1], result)
				}
			}
		})
	}
}

func TestDetermineDryingStatus(t *testing.T) {
	calc := NewCalculator("test-api-key")

	tests := []struct {
		name           string
		hoursUntilDry  float64
		rockType       string
		expectedStatus string
	}{
		{"Dry", 0.0, "granite", "good"},
		{"Wet - not sensitive", 12.0, "granite", "fair"},
		{"Wet - moderate", 30.0, "granite", "fair"},
		{"Wet - long", 50.0, "granite", "poor"},
		{"Wet sandstone - critical", 12.0, "sandstone", "critical"},
		{"Wet arkose - critical", 12.0, "arkose", "critical"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := &BoulderDryingStatus{
				IsWet:         tt.hoursUntilDry > 0,
				HoursUntilDry: tt.hoursUntilDry,
				RockType:      tt.rockType,
			}

			result := calc.determineDryingStatus(status)

			if result != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, result)
			}
		})
	}
}

func TestEstimateSunExposureFromAspect(t *testing.T) {
	calc := NewCalculator("test-api-key")

	tests := []struct {
		aspect          string
		expectedMinimum float64
		expectedMaximum float64
	}{
		{"N", 10.0, 14.0},   // 2 hours/day * 6 days
		{"NE", 16.0, 20.0},  // 3 hours/day * 6 days
		{"E", 22.0, 26.0},   // 4 hours/day * 6 days
		{"SE", 34.0, 38.0},  // 6 hours/day * 6 days
		{"S", 46.0, 50.0},   // 8 hours/day * 6 days (max)
		{"SW", 34.0, 38.0},  // 6 hours/day * 6 days
		{"W", 22.0, 26.0},   // 4 hours/day * 6 days
		{"NW", 16.0, 20.0},  // 3 hours/day * 6 days
		{"Unknown", 22.0, 26.0}, // Default fallback
	}

	for _, tt := range tests {
		t.Run(tt.aspect, func(t *testing.T) {
			result := calc.estimateSunExposureFromAspect(tt.aspect)

			if result < tt.expectedMinimum || result > tt.expectedMaximum {
				t.Errorf("Expected %.0f-%.0f hours, got %.0f hours",
					tt.expectedMinimum, tt.expectedMaximum, result)
			}
		})
	}
}

func TestAngleDifference(t *testing.T) {
	tests := []struct {
		angle1   float64
		angle2   float64
		expected float64
	}{
		{0, 0, 0},
		{90, 90, 0},
		{0, 90, 90},
		{90, 0, 90},
		{0, 180, 180},
		{0, 270, 90},  // Wraps around
		{350, 10, 20}, // Wraps around
		{10, 350, 20}, // Wraps around
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := angleDifference(tt.angle1, tt.angle2)

			if result != tt.expected {
				t.Errorf("angleDifference(%.0f, %.0f) = %.0f, expected %.0f",
					tt.angle1, tt.angle2, result, tt.expected)
			}
		})
	}
}

func TestCalculateBoulderDryingStatus(t *testing.T) {
	calc := NewCalculator("test-api-key")
	ctx := context.Background()

	route := &models.MPRoute{
		MPRouteID: 123456,
		Name:      "Test Boulder",
		Latitude:  func() *float64 { v := 47.8172; return &v }(),
		Longitude: func() *float64 { v := -121.6019; return &v }(),
		Aspect:    func() *string { v := "S"; return &v }(),
	}

	locationDrying := &models.RockDryingStatus{
		IsWet:              true,
		HoursUntilDry:      24.0,
		PrimaryRockType:    "granite",
		IsWetSensitive:     false,
		LastRainTimestamp:  time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
	}

	profile := &models.BoulderDryingProfile{
		TreeCoveragePercent: func() *float64 { v := 30.0; return &v }(),
	}

	// No hourly forecast for this test
	status, err := calc.CalculateBoulderDryingStatus(ctx, route, locationDrying, profile, 30.0, nil)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if status.MPRouteID != 123456 {
		t.Errorf("Expected route ID 123456, got %d", status.MPRouteID)
	}

	if status.Latitude != 47.8172 {
		t.Errorf("Expected latitude 47.8172, got %.4f", status.Latitude)
	}

	if status.Aspect != "S" {
		t.Errorf("Expected aspect S, got %s", status.Aspect)
	}

	if status.TreeCoveragePercent != 30.0 {
		t.Errorf("Expected tree coverage 30%%, got %.1f%%", status.TreeCoveragePercent)
	}

	if !status.IsWet {
		t.Error("Expected boulder to be wet")
	}

	// South-facing should have good sun exposure (realistic values with corrected sun calculations)
	if status.SunExposureHours < 25.0 || status.SunExposureHours > 40.0 {
		t.Errorf("Expected sun exposure 25-40 hours for south aspect, got %.1f hours", status.SunExposureHours)
	}

	// Confidence should be high when all data is available
	if status.ConfidenceScore < 90 {
		t.Errorf("Expected high confidence with all data available, got %d", status.ConfidenceScore)
	}
}

func TestCalculate6DayForecast_CurrentlyDry_StaysDry(t *testing.T) {
	calc := NewCalculator("test-api-key")
	now := time.Now()

	status := &BoulderDryingStatus{
		IsWet:         false,
		HoursUntilDry: 0,
	}

	// No rain in forecast - stays dry
	hourlyForecast := make([]models.WeatherData, 144) // 6 days
	for i := 0; i < 144; i++ {
		hourlyForecast[i] = models.WeatherData{
			Timestamp:     now.Add(time.Duration(i) * time.Hour),
			Precipitation: 0.0,
		}
	}

	forecast := calc.Calculate6DayForecast(status, hourlyForecast, 24.0)

	if len(forecast) != 1 {
		t.Fatalf("Expected 1 period (all dry), got %d", len(forecast))
	}

	if !forecast[0].IsDry {
		t.Error("Expected period to be dry")
	}

	if forecast[0].Status != "dry" {
		t.Errorf("Expected status 'dry', got '%s'", forecast[0].Status)
	}
}

func TestCalculate6DayForecast_CurrentlyDry_GetWet(t *testing.T) {
	calc := NewCalculator("test-api-key")
	now := time.Now()

	status := &BoulderDryingStatus{
		IsWet:         false,
		HoursUntilDry: 0,
	}

	// Rain starts at hour 24
	hourlyForecast := make([]models.WeatherData, 144)
	for i := 0; i < 144; i++ {
		precipitation := 0.0
		if i == 24 {
			precipitation = 0.25 // Quarter inch of rain
		}
		hourlyForecast[i] = models.WeatherData{
			Timestamp:     now.Add(time.Duration(i) * time.Hour),
			Precipitation: precipitation,
		}
	}

	forecast := calc.Calculate6DayForecast(status, hourlyForecast, 24.0)

	// Should have 2 periods: dry (0-24h), then wet (24h-48h), then dry (48h+)
	if len(forecast) < 2 {
		t.Fatalf("Expected at least 2 periods, got %d", len(forecast))
	}

	// First period should be dry
	if !forecast[0].IsDry {
		t.Error("Expected first period to be dry")
	}

	// Second period should be wet
	if forecast[1].IsDry {
		t.Error("Expected second period to be wet")
	}

	if forecast[1].Status != "wet" && forecast[1].Status != "drying" {
		t.Errorf("Expected wet or drying status, got '%s'", forecast[1].Status)
	}

	if forecast[1].RainAmount != 0.25 {
		t.Errorf("Expected 0.25 inches of rain, got %.2f", forecast[1].RainAmount)
	}
}

func TestCalculate6DayForecast_CurrentlyWet_DriesThenWetAgain(t *testing.T) {
	calc := NewCalculator("test-api-key")
	now := time.Now()

	lastRain := now.Add(-12 * time.Hour)
	status := &BoulderDryingStatus{
		IsWet:             true,
		HoursUntilDry:     12.0,
		LastRainTimestamp: &lastRain,
	}

	// No immediate rain, then rain at hour 36
	hourlyForecast := make([]models.WeatherData, 144)
	for i := 0; i < 144; i++ {
		precipitation := 0.0
		if i == 36 {
			precipitation = 0.5 // Half inch of rain
		}
		hourlyForecast[i] = models.WeatherData{
			Timestamp:     now.Add(time.Duration(i) * time.Hour),
			Precipitation: precipitation,
		}
	}

	forecast := calc.Calculate6DayForecast(status, hourlyForecast, 24.0)

	// Should have at least 3 periods: wet (drying), dry, wet
	if len(forecast) < 3 {
		t.Fatalf("Expected at least 3 periods, got %d", len(forecast))
	}

	// First period should be wet (drying)
	if forecast[0].IsDry {
		t.Error("Expected first period to be wet")
	}

	// Second period should be dry
	if !forecast[1].IsDry {
		t.Error("Expected second period to be dry")
	}

	// Third period should be wet again
	if forecast[2].IsDry {
		t.Error("Expected third period to be wet")
	}
}

func TestCalculate6DayForecast_HeavyRain_LongerDrying(t *testing.T) {
	calc := NewCalculator("test-api-key")
	now := time.Now()

	status := &BoulderDryingStatus{
		IsWet:         false,
		HoursUntilDry: 0,
	}

	// Heavy rain (1 inch) at hour 24
	hourlyForecast := make([]models.WeatherData, 144)
	for i := 0; i < 144; i++ {
		precipitation := 0.0
		if i == 24 {
			precipitation = 1.0 // 1 inch of rain
		}
		hourlyForecast[i] = models.WeatherData{
			Timestamp:     now.Add(time.Duration(i) * time.Hour),
			Precipitation: precipitation,
		}
	}

	forecast := calc.Calculate6DayForecast(status, hourlyForecast, 24.0)

	// Find the wet period
	var wetPeriod *DryingForecastPeriod
	for i := range forecast {
		if !forecast[i].IsDry {
			wetPeriod = &forecast[i]
			break
		}
	}

	if wetPeriod == nil {
		t.Fatal("Expected to find a wet period")
	}

	// Heavy rain (>0.5 inches) should accumulate
	if wetPeriod.RainAmount < 0.9 {
		t.Errorf("Expected at least 0.9 inches of rain accumulated, got %.2f", wetPeriod.RainAmount)
	}

	// Should take longer to dry - check that the wet period extends beyond base drying time
	// Heavy rain: Base 24h + (1.0 - 0.5) * 12 = 30h total
	dryTime := wetPeriod.StartTime.Add(30 * time.Hour)
	if !wetPeriod.EndTime.IsZero() && wetPeriod.EndTime.Before(dryTime.Add(-1*time.Hour)) {
		t.Errorf("Expected wet period to last at least 29h, but it dried at %v", wetPeriod.EndTime.Sub(wetPeriod.StartTime))
	}
}

func TestCalculate6DayForecast_ContinuousRain_AccumulatesAndExtends(t *testing.T) {
	calc := NewCalculator("test-api-key")
	now := time.Now()

	status := &BoulderDryingStatus{
		IsWet:         false,
		HoursUntilDry: 0,
	}

	// Continuous light rain for 12 hours
	hourlyForecast := make([]models.WeatherData, 144)
	for i := 0; i < 144; i++ {
		precipitation := 0.0
		if i >= 24 && i < 36 {
			precipitation = 0.05 // Light continuous rain
		}
		hourlyForecast[i] = models.WeatherData{
			Timestamp:     now.Add(time.Duration(i) * time.Hour),
			Precipitation: precipitation,
		}
	}

	forecast := calc.Calculate6DayForecast(status, hourlyForecast, 24.0)

	// Find the wet period
	var wetPeriod *DryingForecastPeriod
	for i := range forecast {
		if !forecast[i].IsDry {
			wetPeriod = &forecast[i]
			break
		}
	}

	if wetPeriod == nil {
		t.Fatal("Expected to find a wet period")
	}

	// Should accumulate all the rain
	expectedTotal := 0.05 * 12 // 0.6 inches total
	if wetPeriod.RainAmount < expectedTotal-0.1 || wetPeriod.RainAmount > expectedTotal+0.1 {
		t.Errorf("Expected ~%.2f inches accumulated, got %.2f", expectedTotal, wetPeriod.RainAmount)
	}
}

func TestCalculate6DayForecast_EmptyForecast_ReturnsNil(t *testing.T) {
	calc := NewCalculator("test-api-key")

	status := &BoulderDryingStatus{
		IsWet:         false,
		HoursUntilDry: 0,
	}

	forecast := calc.Calculate6DayForecast(status, []models.WeatherData{}, 24.0)

	if forecast != nil {
		t.Error("Expected nil forecast for empty hourly data")
	}
}

func TestCalculate6DayForecast_DryingTransition(t *testing.T) {
	calc := NewCalculator("test-api-key")
	now := time.Now()

	lastRain2 := now.Add(-6 * time.Hour)
	status := &BoulderDryingStatus{
		IsWet:             true,
		HoursUntilDry:     24.0,
		LastRainTimestamp: &lastRain2,
	}

	// No more rain in forecast
	hourlyForecast := make([]models.WeatherData, 144)
	for i := 0; i < 144; i++ {
		hourlyForecast[i] = models.WeatherData{
			Timestamp:     now.Add(time.Duration(i) * time.Hour),
			Precipitation: 0.0,
		}
	}

	forecast := calc.Calculate6DayForecast(status, hourlyForecast, 24.0)

	if len(forecast) < 1 {
		t.Fatal("Expected at least 1 period")
	}

	// First period should be wet but drying
	if forecast[0].IsDry {
		t.Error("Expected first period to be wet (still drying)")
	}

	if forecast[0].Status != "drying" && forecast[0].Status != "wet" {
		t.Errorf("Expected status 'drying' or 'wet', got '%s'", forecast[0].Status)
	}

	// Should eventually transition to dry
	foundDry := false
	for _, period := range forecast {
		if period.IsDry {
			foundDry = true
			break
		}
	}

	if !foundDry {
		t.Error("Expected to eventually find a dry period")
	}
}
