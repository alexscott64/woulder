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
				Temperature:   32,
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
				Temperature:   38,
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
				Temperature:   78,
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
			name: "High humidity at 70°F - marginal (warm + humidity)",
			weather: &models.WeatherData{
				Temperature:   70,
				Precipitation: 0,
				WindSpeed:     5,
				Humidity:      90,
				Timestamp:     time.Now().In(pacificLoc),
			},
			expected: "marginal",
			reasons:  2, // Warm temp (> 65) + humidity (> 65)
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
				Temperature:   60,
				Precipitation: 0,
				WindSpeed:     8,
				Humidity:      60,
				Timestamp:     today,
			},
			hourlyForecast: []models.WeatherData{
				{Timestamp: today.Add(1 * time.Hour), Temperature: 62, Precipitation: 0, WindSpeed: 8, Humidity: 60},
				{Timestamp: today.Add(2 * time.Hour), Temperature: 64, Precipitation: 0, WindSpeed: 7, Humidity: 58},
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
			name:        "High humidity at 39°F - marginal (just cold, not humidity)",
			temperature: 39,
			humidity:    90,
			expected:    "marginal",
			reasons:     1, // Just cold temp (< 41)
		},
		{
			name:        "High humidity at 41°F - good (at lower threshold)",
			temperature: 41,
			humidity:    90,
			expected:    "good",
			reasons:     0, // At threshold, no cold warning
		},
		{
			name:        "High humidity at 50°F - good (comfortable range)",
			temperature: 50,
			humidity:    90,
			expected:    "good",
			reasons:     0, // No issues
		},
		{
			name:        "High humidity at 65°F - good (at upper threshold)",
			temperature: 65,
			humidity:    90,
			expected:    "good",
			reasons:     0, // At threshold, humidity doesn't matter yet
		},
		// Above comfortable range - humidity should matter
		{
			name:        "High humidity at 66°F - marginal (warm + humidity)",
			temperature: 66,
			humidity:    90,
			expected:    "marginal",
			reasons:     2, // Warm temp (> 65) + humidity (> 65)
		},
		{
			name:        "High humidity at 70°F - marginal (warm + humidity)",
			temperature: 70,
			humidity:    90,
			expected:    "marginal",
			reasons:     2, // Warm temp (> 65) + humidity (> 65)
		},
		{
			name:        "High humidity at 78°F - bad (too hot + humidity)",
			temperature: 78,
			humidity:    90,
			expected:    "bad",
			reasons:     2, // Too hot (> 75) + humidity
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

func TestFilterPastHourReasons(t *testing.T) {
	calc := &ConditionCalculator{}

	tests := []struct {
		name          string
		input         models.ClimbingCondition
		expectedLevel string
		expectedCount int
	}{
		{
			name: "Keep precipitation, remove temp",
			input: models.ClimbingCondition{
				Level:   "bad",
				Reasons: []string{"Heavy rain (0.15in/hr)", "Cold (38°F)"},
			},
			expectedLevel: "bad",
			expectedCount: 1, // Only rain reason kept
		},
		{
			name: "Remove all temp/wind, return good",
			input: models.ClimbingCondition{
				Level:   "marginal",
				Reasons: []string{"Cold (38°F)", "Moderate winds (15mph)"},
			},
			expectedLevel: "good",
			expectedCount: 0, // All filtered out
		},
		{
			name: "Keep light rain",
			input: models.ClimbingCondition{
				Level:   "marginal",
				Reasons: []string{"Light rain (0.02in/hr)", "Strong winds (22mph)"},
			},
			expectedLevel: "marginal",
			expectedCount: 1, // Only rain kept
		},
		{
			name: "Keep persistent drizzle",
			input: models.ClimbingCondition{
				Level:   "marginal",
				Reasons: []string{"Persistent drizzle (0.08in over 2h)", "Cold (39°F)"},
			},
			expectedLevel: "marginal",
			expectedCount: 1, // Only drizzle kept
		},
		{
			name: "No reasons - return good",
			input: models.ClimbingCondition{
				Level:   "good",
				Reasons: []string{},
			},
			expectedLevel: "good",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.filterPastHourReasons(tt.input)

			if result.Level != tt.expectedLevel {
				t.Errorf("Expected level %s, got %s", tt.expectedLevel, result.Level)
			}

			if len(result.Reasons) != tt.expectedCount {
				t.Errorf("Expected %d reasons, got %d: %v", tt.expectedCount, len(result.Reasons), result.Reasons)
			}
		})
	}
}

func TestWindTemperatureThresholds(t *testing.T) {
	calc := &ConditionCalculator{}
	pacificLoc, _ := time.LoadLocation("America/Los_Angeles")

	tests := []struct {
		name        string
		temperature float64
		windSpeed   float64
		expected    string
		hasReason   bool
	}{
		// Cold temps (< 38°F) - lower wind threshold (18mph)
		{
			name:        "Cold temp with 17mph wind - good",
			temperature: 37,
			windSpeed:   17,
			expected:    "marginal", // marginal due to cold, but not wind
			hasReason:   true,       // Cold reason only
		},
		{
			name:        "Cold temp with 19mph wind - marginal",
			temperature: 37,
			windSpeed:   19,
			expected:    "marginal",
			hasReason:   true, // Cold + wind
		},
		{
			name:        "Cold temp with 22mph wind - marginal",
			temperature: 37,
			windSpeed:   22,
			expected:    "marginal", // marginal due to cold + strong winds
			hasReason:   true,
		},
		// Warm temps (>= 38°F) - higher wind threshold (20mph)
		{
			name:        "Warm temp with 19mph wind - good",
			temperature: 60,
			windSpeed:   19,
			expected:    "good",
			hasReason:   false,
		},
		{
			name:        "Warm temp with 21mph wind - marginal",
			temperature: 60,
			windSpeed:   21,
			expected:    "marginal",
			hasReason:   true, // Wind only
		},
		{
			name:        "Warm temp with 35mph wind - bad",
			temperature: 60,
			windSpeed:   35,
			expected:    "bad",
			hasReason:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			weather := &models.WeatherData{
				Temperature:   tt.temperature,
				Precipitation: 0,
				WindSpeed:     tt.windSpeed,
				Humidity:      60,
				Timestamp:     time.Now().In(pacificLoc),
			}

			result := calc.calculateInstantCondition(weather, []models.WeatherData{})

			if result.Level != tt.expected {
				t.Errorf("Expected level %s, got %s. Reasons: %v", tt.expected, result.Level, result.Reasons)
			}

			if tt.hasReason && len(result.Reasons) == 0 {
				t.Errorf("Expected reasons but got none")
			}
		})
	}
}

func TestTodayConditionIgnoresPastTempAndWind(t *testing.T) {
	calc := &ConditionCalculator{}
	pacificLoc, _ := time.LoadLocation("America/Los_Angeles")
	if pacificLoc == nil {
		pacificLoc = time.UTC
	}

	// Simulate: It's 2pm, and it was cold at 8am (38°F) but now it's 60°F
	now := time.Now().In(pacificLoc)
	currentTime := time.Date(now.Year(), now.Month(), now.Day(), 14, 0, 0, 0, pacificLoc) // 2pm

	tests := []struct {
		name           string
		current        *models.WeatherData
		hourlyForecast []models.WeatherData
		expectedLevel  string
		shouldNotHave  string // reason that should NOT be in result
	}{
		{
			name: "Past cold temp should be ignored",
			current: &models.WeatherData{
				Temperature:   60,
				Precipitation: 0,
				WindSpeed:     8,
				Humidity:      60,
				Timestamp:     currentTime, // 2pm now, 60°F
			},
			hourlyForecast: []models.WeatherData{
				// Past hour at 8am was cold (should be filtered out)
				{Timestamp: currentTime.Add(-6 * time.Hour), Temperature: 38, Precipitation: 0, WindSpeed: 8, Humidity: 60}, // 8am
				// Future hours are good
				{Timestamp: currentTime.Add(1 * time.Hour), Temperature: 62, Precipitation: 0, WindSpeed: 8, Humidity: 60}, // 3pm
				{Timestamp: currentTime.Add(2 * time.Hour), Temperature: 61, Precipitation: 0, WindSpeed: 7, Humidity: 58}, // 4pm
			},
			expectedLevel: "good",
			shouldNotHave: "Cold",
		},
		{
			name: "Past high wind should be ignored",
			current: &models.WeatherData{
				Temperature:   60,
				Precipitation: 0,
				WindSpeed:     10,
				Humidity:      60,
				Timestamp:     currentTime, // 2pm now, calm
			},
			hourlyForecast: []models.WeatherData{
				// Past hour at 8am was windy (should be filtered out)
				{Timestamp: currentTime.Add(-6 * time.Hour), Temperature: 60, Precipitation: 0, WindSpeed: 25, Humidity: 60}, // 8am
				// Future hours are good
				{Timestamp: currentTime.Add(1 * time.Hour), Temperature: 62, Precipitation: 0, WindSpeed: 8, Humidity: 60}, // 3pm
			},
			expectedLevel: "good",
			shouldNotHave: "wind",
		},
		{
			name: "Past rain should still affect condition",
			current: &models.WeatherData{
				Temperature:   60,
				Precipitation: 0,
				WindSpeed:     8,
				Humidity:      60,
				Timestamp:     currentTime, // 2pm now, no rain
			},
			hourlyForecast: []models.WeatherData{
				// Past hour at 10am had rain (should be kept because rock might still be wet)
				{Timestamp: currentTime.Add(-4 * time.Hour), Temperature: 60, Precipitation: 0.08, WindSpeed: 8, Humidity: 75}, // 10am
				// Future hours are good
				{Timestamp: currentTime.Add(1 * time.Hour), Temperature: 62, Precipitation: 0, WindSpeed: 8, Humidity: 60}, // 3pm
			},
			expectedLevel: "bad", // Should be bad because of rain
			shouldNotHave: "",    // We DO want rain in the reasons
		},
		{
			name: "Future cold temp should affect condition",
			current: &models.WeatherData{
				Temperature:   60,
				Precipitation: 0,
				WindSpeed:     8,
				Humidity:      60,
				Timestamp:     currentTime, // 2pm now, good
			},
			hourlyForecast: []models.WeatherData{
				// Future hour at 6pm will be cold
				{Timestamp: currentTime.Add(4 * time.Hour), Temperature: 38, Precipitation: 0, WindSpeed: 8, Humidity: 60}, // 6pm
			},
			expectedLevel: "marginal", // Should be marginal because it will be cold later
			shouldNotHave: "",         // We DO want cold warning for future
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.CalculateTodayCondition(tt.current, tt.hourlyForecast, []models.WeatherData{})

			if result.Level != tt.expectedLevel {
				t.Errorf("Expected level %s, got %s. Reasons: %v", tt.expectedLevel, result.Level, result.Reasons)
			}

			if tt.shouldNotHave != "" {
				for _, reason := range result.Reasons {
					if contains(toLower(reason), toLower(tt.shouldNotHave)) {
						t.Errorf("Should not have reason containing '%s', but got: %v", tt.shouldNotHave, result.Reasons)
					}
				}
			}
		})
	}
}
