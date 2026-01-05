package pests

import (
	"testing"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

func TestCalculateMosquitoScore(t *testing.T) {
	analyzer := &PestAnalyzer{}

	tests := []struct {
		name            string
		currentTemp     float64
		currentHumidity int
		currentWind     float64
		recentRainfall  float64
		month           time.Month
		expectedLevel   PestLevel
		minScore        int
		maxScore        int
	}{
		{
			name:            "Too cold - dormant",
			currentTemp:     45,
			currentHumidity: 60,
			currentWind:     5,
			recentRainfall:  1.0,
			month:           time.January,
			expectedLevel:   PestLevelLow,
			minScore:        0,
			maxScore:        5,
		},
		{
			name:            "Optimal conditions - summer",
			currentTemp:     75,
			currentHumidity: 75,
			currentWind:     3,
			recentRainfall:  2.0,
			month:           time.July,
			expectedLevel:   PestLevelExtreme,
			minScore:        85,
			maxScore:        100,
		},
		{
			name:            "Moderate conditions - spring",
			currentTemp:     65,
			currentHumidity: 55,
			currentWind:     8,
			recentRainfall:  0.5,
			month:           time.May,
			expectedLevel:   PestLevelHigh,
			minScore:        40,
			maxScore:        65,
		},
		{
			name:            "Low activity - dry and windy",
			currentTemp:     70,
			currentHumidity: 30,
			currentWind:     18,
			recentRainfall:  0,
			month:           time.August,
			expectedLevel:   PestLevelHigh,
			minScore:        40,
			maxScore:        55,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.CalculateMosquitoScore(
				tt.currentTemp,
				tt.currentHumidity,
				tt.currentWind,
				tt.recentRainfall,
				tt.month,
			)

			if result.Score < tt.minScore || result.Score > tt.maxScore {
				t.Errorf("Expected score between %d and %d, got %d", tt.minScore, tt.maxScore, result.Score)
			}

			level := ScoreToLevel(result.Score)
			if level != tt.expectedLevel {
				t.Errorf("Expected level %s, got %s (score: %d)", tt.expectedLevel, level, result.Score)
			}

			if len(result.Factors) == 0 {
				t.Error("Expected at least one factor")
			}
		})
	}
}

func TestCalculateOutdoorPestScore(t *testing.T) {
	analyzer := &PestAnalyzer{}

	tests := []struct {
		name            string
		currentTemp     float64
		currentHumidity int
		recentRainfall  float64
		month           time.Month
		expectedLevel   PestLevel
		minScore        int
		maxScore        int
	}{
		{
			name:            "Too cold - minimal activity",
			currentTemp:     45,
			currentHumidity: 60,
			recentRainfall:  1.0,
			month:           time.January,
			expectedLevel:   PestLevelLow,
			minScore:        0,
			maxScore:        10,
		},
		{
			name:            "Optimal conditions - peak season",
			currentTemp:     80,
			currentHumidity: 70,
			recentRainfall:  1.5,
			month:           time.July,
			expectedLevel:   PestLevelExtreme,
			minScore:        80,
			maxScore:        100,
		},
		{
			name:            "Moderate conditions - spring",
			currentTemp:     68,
			currentHumidity: 50,
			recentRainfall:  0.7,
			month:           time.May,
			expectedLevel:   PestLevelVeryHigh,
			minScore:        60,
			maxScore:        80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.CalculateOutdoorPestScore(
				tt.currentTemp,
				tt.currentHumidity,
				tt.recentRainfall,
				tt.month,
			)

			if result.Score < tt.minScore || result.Score > tt.maxScore {
				t.Errorf("Expected score between %d and %d, got %d", tt.minScore, tt.maxScore, result.Score)
			}

			level := ScoreToLevel(result.Score)
			if level != tt.expectedLevel {
				t.Errorf("Expected level %s, got %s (score: %d)", tt.expectedLevel, level, result.Score)
			}

			if len(result.Factors) == 0 {
				t.Error("Expected at least one factor")
			}
		})
	}
}

func TestAssessConditions(t *testing.T) {
	analyzer := &PestAnalyzer{}
	now := time.Now()

	tests := []struct {
		name       string
		current    *models.WeatherData
		historical []models.WeatherData
		expectLow  bool // If true, expect both levels to be low
	}{
		{
			name: "Good conditions with recent rain",
			current: &models.WeatherData{
				Temperature:   75,
				Humidity:      70,
				WindSpeed:     5,
				Precipitation: 0,
				Timestamp:     now,
			},
			historical: []models.WeatherData{
				{Timestamp: now.Add(-48 * time.Hour), Precipitation: 1.0},
				{Timestamp: now.Add(-24 * time.Hour), Precipitation: 0.5},
			},
			expectLow: false,
		},
		{
			name: "Cold winter conditions",
			current: &models.WeatherData{
				Temperature:   40,
				Humidity:      50,
				WindSpeed:     10,
				Precipitation: 0,
				Timestamp:     now,
			},
			historical: []models.WeatherData{},
			expectLow:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.AssessConditions(tt.current, tt.historical)

			if tt.expectLow {
				if result.MosquitoLevel != PestLevelLow {
					t.Errorf("Expected low mosquito level, got %s", result.MosquitoLevel)
				}
				if result.OutdoorPestLevel != PestLevelLow {
					t.Errorf("Expected low outdoor pest level, got %s", result.OutdoorPestLevel)
				}
			} else {
				// For non-low conditions, just check that we got valid results
				if result.MosquitoScore < 0 || result.MosquitoScore > 100 {
					t.Errorf("Invalid mosquito score: %d", result.MosquitoScore)
				}
				if result.OutdoorPestScore < 0 || result.OutdoorPestScore > 100 {
					t.Errorf("Invalid outdoor pest score: %d", result.OutdoorPestScore)
				}
			}

			// Check that factors are limited to 4
			if len(result.Factors) > 4 {
				t.Errorf("Expected max 4 factors, got %d", len(result.Factors))
			}
		})
	}
}

func TestGetSeasonalMosquitoFactor(t *testing.T) {
	tests := []struct {
		month    time.Month
		expected float64
	}{
		{time.January, 0.0},
		{time.February, 0.0},
		{time.March, 0.1},
		{time.April, 0.3},
		{time.May, 0.6},
		{time.June, 0.9},
		{time.July, 1.0},
		{time.August, 1.0},
		{time.September, 0.7},
		{time.October, 0.3},
		{time.November, 0.1},
		{time.December, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.month.String(), func(t *testing.T) {
			result := GetSeasonalMosquitoFactor(tt.month)
			if result != tt.expected {
				t.Errorf("Expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestScoreToLevel(t *testing.T) {
	tests := []struct {
		score    int
		expected PestLevel
	}{
		{0, PestLevelLow},
		{15, PestLevelLow},
		{20, PestLevelModerate},
		{39, PestLevelModerate},
		{40, PestLevelHigh},
		{59, PestLevelHigh},
		{60, PestLevelVeryHigh},
		{79, PestLevelVeryHigh},
		{80, PestLevelExtreme},
		{100, PestLevelExtreme},
	}

	for _, tt := range tests {
		t.Run(string(tt.expected), func(t *testing.T) {
			result := ScoreToLevel(tt.score)
			if result != tt.expected {
				t.Errorf("Score %d: expected %s, got %s", tt.score, tt.expected, result)
			}
		})
	}
}

func TestCalculateRecentRainfall(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		historical []models.WeatherData
		daysAgo    int
		expected   float64
	}{
		{
			name: "Rain within 7 days",
			historical: []models.WeatherData{
				{Timestamp: now.Add(-48 * time.Hour), Precipitation: 1.0},
				{Timestamp: now.Add(-24 * time.Hour), Precipitation: 0.5},
			},
			daysAgo:  7,
			expected: 1.5,
		},
		{
			name: "Rain outside window",
			historical: []models.WeatherData{
				{Timestamp: now.Add(-10 * 24 * time.Hour), Precipitation: 2.0},
			},
			daysAgo:  7,
			expected: 0.0,
		},
		{
			name:       "No rain",
			historical: []models.WeatherData{},
			daysAgo:    7,
			expected:   0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateRecentRainfall(tt.historical, tt.daysAgo)
			if result != tt.expected {
				t.Errorf("Expected %.2f, got %.2f", tt.expected, result)
			}
		})
	}
}

func TestGetWorstLevel(t *testing.T) {
	tests := []struct {
		level1   PestLevel
		level2   PestLevel
		expected PestLevel
	}{
		{PestLevelLow, PestLevelModerate, PestLevelModerate},
		{PestLevelHigh, PestLevelLow, PestLevelHigh},
		{PestLevelExtreme, PestLevelVeryHigh, PestLevelExtreme},
		{PestLevelModerate, PestLevelModerate, PestLevelModerate},
	}

	for _, tt := range tests {
		t.Run(string(tt.expected), func(t *testing.T) {
			result := GetWorstLevel(tt.level1, tt.level2)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}
