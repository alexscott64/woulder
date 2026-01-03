package calculator

import (
	"fmt"
	"testing"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// TestMoneyCreekSnowAccumulation tests the snow accumulation model against real-world observations
// User reported 3-5" of snow on ground at Money Creek on Jan 1, 2026
// Model should predict similar amounts
func TestMoneyCreekSnowAccumulation(t *testing.T) {
	elevationFt := 1000.0 // Money Creek elevation

	// Build historical data for the snowstorm period
	// Let's simulate Dec 26-Jan 1 conditions that would create 3-5" of snow
	var historical []models.WeatherData

	// Dec 26-28: Heavy snowfall period
	// Simulate a storm with accumulation
	for day := 26; day <= 28; day++ {
		baseTime := time.Date(2025, 12, day, 0, 0, 0, 0, time.UTC)
		for hour := 0; hour < 24; hour++ {
			historical = append(historical, models.WeatherData{
				Timestamp:     baseTime.Add(time.Duration(hour) * time.Hour),
				Temperature:   28.0, // Cold enough for snow
				Precipitation: 0.05, // Heavy snow rate
				WindSpeed:     8,
				Humidity:      85,
			})
		}
	}

	// Dec 29-31: Warmer temps with some melt
	for day := 29; day <= 31; day++ {
		baseTime := time.Date(2025, 12, day, 0, 0, 0, 0, time.UTC)
		for hour := 0; hour < 24; hour++ {
			historical = append(historical, models.WeatherData{
				Timestamp:     baseTime.Add(time.Duration(hour) * time.Hour),
				Temperature:   36.0, // Above freezing - melt
				Precipitation: 0,
				WindSpeed:     10,
				Humidity:      75,
			})
		}
	}

	// Jan 1: More mild temps
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for hour := 0; hour < 24; hour++ {
		historical = append(historical, models.WeatherData{
			Timestamp:     baseTime.Add(time.Duration(hour) * time.Hour),
			Temperature:   35.0, // Mild - slow melt
			Precipitation: 0,
			WindSpeed:     8,
			Humidity:      70,
		})
	}

	// Current data: Jan 2 (today)
	var current []models.WeatherData
	nowBase := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	for hour := 0; hour < 12; hour++ {
		current = append(current, models.WeatherData{
			Timestamp:     nowBase.Add(time.Duration(hour) * time.Hour),
			Temperature:   34.0, // Right at freezing - minimal melt
			Precipitation: 0,
			WindSpeed:     7,
			Humidity:      80,
		})
	}

	// Calculate snow depth using the model
	snowDepth := GetCurrentSnowDepth(historical, current, elevationFt)

	fmt.Printf("\n=== Money Creek Snow Test ===\n")
	fmt.Printf("Historical data: Dec 26 - Jan 1 (%d hours)\n", len(historical))
	fmt.Printf("Current data: Jan 2 (%d hours)\n", len(current))
	fmt.Printf("Calculated snow depth: %.2f\"\n", snowDepth)
	fmt.Printf("Expected range: 3-5\"\n")
	fmt.Printf("\n")

	// Also check the daily map
	dailySnowDepth := CalculateSnowAccumulation(historical, current, elevationFt)
	fmt.Printf("Daily snow depth map:\n")
	// Print in chronological order
	for day := 26; day <= 31; day++ {
		dateKey := fmt.Sprintf("2025-12-%02d", day)
		if depth, exists := dailySnowDepth[dateKey]; exists {
			fmt.Printf("  %s: %.2f\"\n", dateKey, depth)
		}
	}
	for day := 1; day <= 2; day++ {
		dateKey := fmt.Sprintf("2026-01-%02d", day)
		if depth, exists := dailySnowDepth[dateKey]; exists {
			fmt.Printf("  %s: %.2f\"\n", dateKey, depth)
		}
	}

	// Test assertions
	if snowDepth < 2.5 {
		t.Errorf("Snow depth too low: %.2f\" (expected 3-5\")", snowDepth)
	}
	if snowDepth > 6.0 {
		t.Errorf("Snow depth too high: %.2f\" (expected 3-5\")", snowDepth)
	}

	// Success range: 2.5-6" (allowing some margin)
	if snowDepth >= 2.5 && snowDepth <= 6.0 {
		fmt.Printf("✓ Snow depth %.2f\" is within acceptable range (2.5-6\")\n", snowDepth)
	}
}
