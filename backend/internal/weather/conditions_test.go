package weather

import (
	"testing"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

func TestCalculateInstantCondition(t *testing.T) {
	calc := &ConditionCalculator{}
	pacificLoc, _ := time.LoadLocation("America/Los_Angeles")

	tests := []struct {
		name     string
		weather  *models.WeatherData
		expected string // expected level
		reasons  int    // expected number of reasons
	}{
		{
			name: "Good conditions",
			weather: &models.WeatherData{
				Temperature:   60,
				Precipitation: 0,
				WindSpeed:     8,
				Humidity:      60,
				Timestamp:     time.Now().In(pacificLoc),
			},
			expected: "good",
			reasons:  0,
		},
		{
			name: "Heavy rain - bad",
			weather: &models.WeatherData{
				Temperature:   60,
				Precipitation: 0.15,
				WindSpeed:     8,
				Humidity:      80,
				Timestamp:     time.Now().In(pacificLoc),
			},
			expected: "bad",
			reasons:  1,
		},
		{
			name: "Light rain - marginal",
			weather: &models.WeatherData{
				Temperature:   60,
				Precipitation: 0.02,
				WindSpeed:     8,
				Humidity:      70,
				Timestamp:     time.Now().In(pacificLoc),
			},
			expected: "marginal",
			reasons:  1,
		},
		{
			name: "Too cold - bad",
			weather: &models.WeatherData{
				Temperature:   35,
				Precipitation: 0,
				WindSpeed:     5,
				Humidity:      60,
				Timestamp:     time.Now().In(pacificLoc),
			},
			expected: "bad",
			reasons:  1,
		},
		{
			name: "Cold - marginal",
			weather: &models.WeatherData{
				Temperature:   43,
				Precipitation: 0,
				WindSpeed:     5,
				Humidity:      60,
				Timestamp:     time.Now().In(pacificLoc),
			},
			expected: "marginal",
			reasons:  1,
		},
		{
			name: "Too hot - bad",
			weather: &models.WeatherData{
				Temperature:   95,
				Precipitation: 0,
				WindSpeed:     5,
				Humidity:      40,
				Timestamp:     time.Now().In(pacificLoc),
			},
			expected: "bad",
			reasons:  1,
		},
		{
			name: "Dangerous winds - bad",
			weather: &models.WeatherData{
				Temperature:   60,
				Precipitation: 0,
				WindSpeed:     35,
				Humidity:      60,
				Timestamp:     time.Now().In(pacificLoc),
			},
			expected: "bad",
			reasons:  1,
		},
		{
			name: "Strong winds - marginal",
			weather: &models.WeatherData{
				Temperature:   60,
				Precipitation: 0,
				WindSpeed:     22,
				Humidity:      60,
				Timestamp:     time.Now().In(pacificLoc),
			},
			expected: "marginal",
			reasons:  1,
		},
		{
			name: "High humidity at 60°F - should be GOOD (not marginal)",
			weather: &models.WeatherData{
				Temperature:   60,
				Precipitation: 0,
				WindSpeed:     5,
				Humidity:      87,
				Timestamp:     time.Now().In(pacificLoc),
			},
			expected: "good",
			reasons:  0,
		},
		{
			name: "High humidity at 57°F - should be GOOD (user's example)",
			weather: &models.WeatherData{
				Temperature:   57,
				Precipitation: 0,
				WindSpeed:     5,
				Humidity:      90,
				Timestamp:     time.Now().In(pacificLoc),
			},
			expected: "good",
			reasons:  0,
		},
		{
			name: "High humidity at 70°F - marginal (above 65°F threshold)",
			weather: &models.WeatherData{
				Temperature:   70,
				Precipitation: 0,
				WindSpeed:     5,
				Humidity:      90,
				Timestamp:     time.Now().In(pacificLoc),
			},
			expected: "marginal",
			reasons:  1,
		},
		{
			name: "High humidity at 30°F - bad (cold temp + humidity warning)",
			weather: &models.WeatherData{
				Temperature:   30,
				Precipitation: 0,
				WindSpeed:     5,
				Humidity:      90,
				Timestamp:     time.Now().In(pacificLoc),
			},
			expected: "bad",
			reasons:  2, // Cold temp + humidity with freezing
		},
		{
			name: "Multiple issues - rain and cold",
			weather: &models.WeatherData{
				Temperature:   38,
				Precipitation: 0.08,
				WindSpeed:     10,
				Humidity:      75,
				Timestamp:     time.Now().In(pacificLoc),
			},
			expected: "bad",
			reasons:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.calculateInstantCondition(tt.weather, []models.WeatherData{})

			if result.Level != tt.expected {
				t.Errorf("Expected level %s, got %s", tt.expected, result.Level)
			}

			if len(result.Reasons) != tt.reasons {
				t.Errorf("Expected %d reasons, got %d: %v", tt.reasons, len(result.Reasons), result.Reasons)
			}
		})
	}
}

func TestCalculateRainLast48h(t *testing.T) {
	calc := &ConditionCalculator{}
	now := time.Now()

	tests := []struct {
		name       string
		historical []models.WeatherData
		hourly     []models.WeatherData
		expected   float64
	}{
		{
			name: "No rain",
			historical: []models.WeatherData{
				{Timestamp: now.Add(-40 * time.Hour), Precipitation: 0},
				{Timestamp: now.Add(-30 * time.Hour), Precipitation: 0},
			},
			hourly: []models.WeatherData{
				{Timestamp: now.Add(-10 * time.Hour), Precipitation: 0},
			},
			expected: 0,
		},
		{
			name: "Rain within 48h",
			historical: []models.WeatherData{
				{Timestamp: now.Add(-40 * time.Hour), Precipitation: 0.5},
				{Timestamp: now.Add(-30 * time.Hour), Precipitation: 0.3},
			},
			hourly: []models.WeatherData{
				{Timestamp: now.Add(-10 * time.Hour), Precipitation: 0.2},
			},
			expected: 1.0,
		},
		{
			name: "Rain outside 48h window",
			historical: []models.WeatherData{
				{Timestamp: now.Add(-60 * time.Hour), Precipitation: 1.0},
			},
			hourly: []models.WeatherData{
				{Timestamp: now.Add(-10 * time.Hour), Precipitation: 0.2},
			},
			expected: 0.2,
		},
		{
			name: "Deduplicate by timestamp",
			historical: []models.WeatherData{
				{Timestamp: now.Add(-10 * time.Hour), Precipitation: 0.5},
			},
			hourly: []models.WeatherData{
				{Timestamp: now.Add(-10 * time.Hour), Precipitation: 0.5}, // Same timestamp
			},
			expected: 0.5, // Should only count once
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.CalculateRainLast48h(tt.historical, tt.hourly)

			if result != tt.expected {
				t.Errorf("Expected %.2f, got %.2f", tt.expected, result)
			}
		})
	}
}

func TestIsClimbingHour(t *testing.T) {
	pacificLoc, _ := time.LoadLocation("America/Los_Angeles")
	if pacificLoc == nil {
		pacificLoc = time.UTC
	}

	tests := []struct {
		name     string
		hour     int
		expected bool
	}{
		{"8am - not climbing", 8, false},
		{"9am - climbing", 9, true},
		{"12pm - climbing", 12, true},
		{"7pm - climbing", 19, true},
		{"8pm - not climbing", 20, false},
		{"9pm - not climbing", 21, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a timestamp at the specified hour in Pacific timezone
			timestamp := time.Date(2026, 1, 5, tt.hour, 0, 0, 0, pacificLoc)
			result := isClimbingHour(timestamp, pacificLoc)

			if result != tt.expected {
				t.Errorf("Hour %d: expected %v, got %v", tt.hour, tt.expected, result)
			}
		})
	}
}

func TestCalculateTodayCondition(t *testing.T) {
	calc := &ConditionCalculator{}
	pacificLoc, _ := time.LoadLocation("America/Los_Angeles")
	if pacificLoc == nil {
		pacificLoc = time.UTC
	}

	now := time.Now().In(pacificLoc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, pacificLoc)

	tests := []struct {
		name           string
		current        *models.WeatherData
		hourlyForecast []models.WeatherData
		historical     []models.WeatherData
		expectedLevel  string
	}{
		{
			name: "Good day - no issues",
			current: &models.WeatherData{
				Temperature:   65,
				Precipitation: 0,
				WindSpeed:     8,
				Humidity:      60,
				Timestamp:     today,
			},
			hourlyForecast: []models.WeatherData{
				{Timestamp: today.Add(1 * time.Hour), Temperature: 65, Precipitation: 0, WindSpeed: 8, Humidity: 60},
				{Timestamp: today.Add(2 * time.Hour), Temperature: 67, Precipitation: 0, WindSpeed: 7, Humidity: 58},
			},
			historical:    []models.WeatherData{},
			expectedLevel: "good",
		},
		{
			name: "Bad day - rain during climbing hours",
			current: &models.WeatherData{
				Temperature:   60,
				Precipitation: 0.1,
				WindSpeed:     8,
				Humidity:      75,
				Timestamp:     today,
			},
			hourlyForecast: []models.WeatherData{
				{Timestamp: today.Add(1 * time.Hour), Temperature: 60, Precipitation: 0.08, WindSpeed: 8, Humidity: 75},
			},
			historical:    []models.WeatherData{},
			expectedLevel: "bad",
		},
		{
			name: "Marginal day - light rain",
			current: &models.WeatherData{
				Temperature:   60,
				Precipitation: 0.02,
				WindSpeed:     8,
				Humidity:      70,
				Timestamp:     today,
			},
			hourlyForecast: []models.WeatherData{
				{Timestamp: today.Add(1 * time.Hour), Temperature: 60, Precipitation: 0.01, WindSpeed: 8, Humidity: 70},
			},
			historical:    []models.WeatherData{},
			expectedLevel: "marginal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.CalculateTodayCondition(tt.current, tt.hourlyForecast, tt.historical)

			if result.Level != tt.expectedLevel {
				t.Errorf("Expected level %s, got %s. Reasons: %v", tt.expectedLevel, result.Level, result.Reasons)
			}
		})
	}
}

func TestHumidityTemperatureThresholds(t *testing.T) {
	calc := &ConditionCalculator{}
	pacificLoc, _ := time.LoadLocation("America/Los_Angeles")

	tests := []struct {
		name        string
		temperature float64
		humidity    int
		expected    string // expected level
		reasons     int    // expected number of reasons
	}{
		// Below freezing - humidity should matter
		{
			name:        "High humidity at 31°F - bad (below freezing)",
			temperature: 31,
			humidity:    90,
			expected:    "bad",
			reasons:     2, // Cold temp + humidity with freezing
		},
		{
			name:        "High humidity at 32°F - bad (exactly freezing)",
			temperature: 32,
			humidity:    90,
			expected:    "bad",
			reasons:     1, // Just cold temp, humidity not triggered at exactly 32
		},
		// Comfortable range - humidity should NOT matter
		{
			name:        "High humidity at 44°F - marginal (just cold, not humidity)",
			temperature: 44,
			humidity:    90,
			expected:    "marginal",
			reasons:     1, // Just cold temp (< 45)
		},
		{
			name:        "High humidity at 50°F - good (comfortable range)",
			temperature: 50,
			humidity:    90,
			expected:    "good",
			reasons:     0, // No issues
		},
		{
			name:        "High humidity at 65°F - good (at threshold)",
			temperature: 65,
			humidity:    90,
			expected:    "good",
			reasons:     0, // Not above 65, so humidity doesn't matter
		},
		// Above comfortable range - humidity should matter
		{
			name:        "High humidity at 66°F - marginal (above 65°F)",
			temperature: 66,
			humidity:    90,
			expected:    "marginal",
			reasons:     1, // Humidity above 65°F
		},
		{
			name:        "High humidity at 86°F - marginal (warm + humidity)",
			temperature: 86,
			humidity:    90,
			expected:    "marginal",
			reasons:     2, // Warm temp (> 85) + humidity (> 65)
		},
		{
			name:        "High humidity at 95°F - bad (too hot + humidity)",
			temperature: 95,
			humidity:    90,
			expected:    "bad",
			reasons:     2, // Too hot + humidity
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			weather := &models.WeatherData{
				Temperature:   tt.temperature,
				Precipitation: 0,
				WindSpeed:     5,
				Humidity:      tt.humidity,
				Timestamp:     time.Now().In(pacificLoc),
			}

			result := calc.calculateInstantCondition(weather, []models.WeatherData{})

			if result.Level != tt.expected {
				t.Errorf("Expected level %s, got %s. Reasons: %v", tt.expected, result.Level, result.Reasons)
			}

			if len(result.Reasons) != tt.reasons {
				t.Errorf("Expected %d reasons, got %d: %v", tt.reasons, len(result.Reasons), result.Reasons)
			}
		})
	}
}

func TestHumidityMessages(t *testing.T) {
	calc := &ConditionCalculator{}
	pacificLoc, _ := time.LoadLocation("America/Los_Angeles")

	tests := []struct {
		name            string
		temperature     float64
		humidity        int
		expectedMessage string
	}{
		{
			name:            "Freezing with high humidity - detailed message",
			temperature:     30,
			humidity:        90,
			expectedMessage: "High humidity with freezing temps (90%, 30°F)",
		},
		{
			name:            "Hot with high humidity - simple message",
			temperature:     70,
			humidity:        88,
			expectedMessage: "High humidity (88%)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			weather := &models.WeatherData{
				Temperature:   tt.temperature,
				Precipitation: 0,
				WindSpeed:     5,
				Humidity:      tt.humidity,
				Timestamp:     time.Now().In(pacificLoc),
			}

			result := calc.calculateInstantCondition(weather, []models.WeatherData{})

			// Check if expected message appears in reasons
			found := false
			for _, reason := range result.Reasons {
				if reason == tt.expectedMessage {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Expected message '%s' not found in reasons: %v", tt.expectedMessage, result.Reasons)
			}
		})
	}
}
