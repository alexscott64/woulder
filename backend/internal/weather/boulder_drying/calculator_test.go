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
		MPRouteID: "123456",
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

	status, err := calc.CalculateBoulderDryingStatus(ctx, route, locationDrying, profile)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if status.MPRouteID != "123456" {
		t.Errorf("Expected route ID 123456, got %s", status.MPRouteID)
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

	// South-facing should have good sun exposure
	if status.SunExposureHours < 40.0 {
		t.Errorf("Expected high sun exposure for south aspect, got %.1f hours", status.SunExposureHours)
	}

	// Confidence should be reduced for missing sun API data
	if status.ConfidenceScore >= 100 {
		t.Errorf("Expected confidence < 100 due to missing sun API, got %d", status.ConfidenceScore)
	}
}
