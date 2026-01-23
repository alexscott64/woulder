package boulder_drying

import (
	"context"
	"os"
	"testing"
)

func TestTreeCoverClient_GetTreeCoverage_GPS(t *testing.T) {
	tests := []struct {
		name           string
		lat            float64
		lon            float64
		expectedMin    float64
		expectedMax    float64
		description    string
	}{
		{
			name:        "Leavenworth Icicle Creek - Dense Forest",
			lat:         47.6,
			lon:         -120.9,
			expectedMin: 55.0,
			expectedMax: 65.0,
			description: "Icicle Creek area should have high tree coverage (60%)",
		},
		{
			name:        "Leavenworth Town Walls - Mixed",
			lat:         47.6,
			lon:         -120.6,
			expectedMin: 20.0,
			expectedMax: 30.0,
			description: "Town walls should have moderate tree coverage (25%)",
		},
		{
			name:        "Bishop Buttermilks - Sparse Desert",
			lat:         37.35,
			lon:         -118.7,
			expectedMin: 3.0,
			expectedMax: 7.0,
			description: "High desert should have minimal tree coverage (5%)",
		},
		{
			name:        "Bishop Happy Boulders - Desert Canyon",
			lat:         37.2,
			lon:         -118.7,
			expectedMin: 10.0,
			expectedMax: 20.0,
			description: "Canyon areas should have slightly more vegetation (15%)",
		},
		{
			name:        "Squamish - Coastal Rainforest",
			lat:         49.7,
			lon:         -123.2,
			expectedMin: 65.0,
			expectedMax: 75.0,
			description: "Coastal rainforest should have high tree coverage (70%)",
		},
		{
			name:        "Red Rocks - Desert",
			lat:         36.15,
			lon:         -115.45,
			expectedMin: 0.0,
			expectedMax: 5.0,
			description: "Desert should have minimal tree coverage (2%)",
		},
		{
			name:        "Smith Rock - High Desert",
			lat:         44.35,
			lon:         -121.15,
			expectedMin: 8.0,
			expectedMax: 12.0,
			description: "High desert with juniper (10%)",
		},
		{
			name:        "Joshua Tree - Desert",
			lat:         34.0,
			lon:         -116.2,
			expectedMin: 1.0,
			expectedMax: 5.0,
			description: "Desert with joshua trees (3%)",
		},
		{
			name:        "Yosemite Valley - Mixed Conifer Forest",
			lat:         37.75,
			lon:         -119.6,
			expectedMin: 50.0,
			expectedMax: 60.0,
			description: "Conifer forest should have high tree coverage (55%)",
		},
		{
			name:        "Unknown Location - Default",
			lat:         40.0,
			lon:         -100.0,
			expectedMin: 25.0,
			expectedMax: 35.0,
			description: "Unknown locations should return moderate default (30%)",
		},
	}

	// Create client without API key (will use location-based estimates)
	client := NewTreeCoverClient()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			coverage, err := client.GetTreeCoverage(ctx, tt.lat, tt.lon)

			if err != nil {
				t.Fatalf("GetTreeCoverage() error = %v", err)
			}

			if coverage < tt.expectedMin || coverage > tt.expectedMax {
				t.Errorf("GetTreeCoverage() = %v, want between %v and %v for %s (lat: %.2f, lon: %.2f)",
					coverage, tt.expectedMin, tt.expectedMax, tt.description, tt.lat, tt.lon)
			}

			t.Logf("✓ %s: %.1f%% tree coverage (expected %.0f-%.0f%%)",
				tt.name, coverage, tt.expectedMin, tt.expectedMax)
		})
	}
}

func TestTreeCoverClient_Disabled(t *testing.T) {
	// Ensure API key is not set for this test
	originalKey := os.Getenv("GOOGLE_EARTH_ENGINE_KEY_PATH")
	os.Unsetenv("GOOGLE_EARTH_ENGINE_KEY_PATH")
	defer func() {
		if originalKey != "" {
			os.Setenv("GOOGLE_EARTH_ENGINE_KEY_PATH", originalKey)
		}
	}()

	client := NewTreeCoverClient()

	if client.enabled {
		t.Error("Client should be disabled when GOOGLE_EARTH_ENGINE_KEY_PATH not set")
	}

	// Should still work with fallback estimates
	ctx := context.Background()
	coverage, err := client.GetTreeCoverage(ctx, 47.6, -120.9)

	if err != nil {
		t.Fatalf("GetTreeCoverage() should work with fallback, error = %v", err)
	}

	if coverage < 0 || coverage > 100 {
		t.Errorf("GetTreeCoverage() = %v, want valid percentage (0-100)", coverage)
	}
}

func TestTreeCoverClient_EstimateTreeCoverageFromLocation(t *testing.T) {
	client := &TreeCoverClient{enabled: false}

	tests := []struct {
		name     string
		lat      float64
		lon      float64
		expected float64
	}{
		{"Leavenworth Icicle Creek", 47.6, -120.9, 60.0},
		{"Leavenworth Town Walls", 47.6, -120.6, 25.0},
		{"Bishop Buttermilks", 37.35, -118.7, 5.0},
		{"Bishop Canyons", 37.2, -118.7, 15.0},
		{"Squamish", 49.7, -123.2, 70.0},
		{"Red Rocks", 36.15, -115.45, 2.0},
		{"Smith Rock", 44.35, -121.15, 10.0},
		{"Joshua Tree", 34.0, -116.2, 3.0},
		{"Yosemite Valley", 37.75, -119.6, 55.0},
		{"Unknown Location", 40.0, -100.0, 30.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coverage := client.estimateTreeCoverageFromLocation(tt.lat, tt.lon)

			if coverage != tt.expected {
				t.Errorf("estimateTreeCoverageFromLocation(%v, %v) = %v, want %v",
					tt.lat, tt.lon, coverage, tt.expected)
			}
		})
	}
}

func TestTreeCoverClient_GPSPrecision(t *testing.T) {
	client := NewTreeCoverClient()
	ctx := context.Background()

	// Test that slight GPS variations within same zone give consistent results
	tests := []struct {
		name     string
		coords   []struct{ lat, lon float64 }
		expected float64
	}{
		{
			name: "Leavenworth Icicle Creek Zone",
			coords: []struct{ lat, lon float64 }{
				{47.55, -120.85},
				{47.60, -120.90},
				{47.65, -120.95},
			},
			expected: 60.0,
		},
		{
			name: "Bishop Buttermilks Zone",
			coords: []struct{ lat, lon float64 }{
				{37.31, -118.65},
				{37.35, -118.70},
				{37.40, -118.75},
			},
			expected: 5.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, coord := range tt.coords {
				coverage, err := client.GetTreeCoverage(ctx, coord.lat, coord.lon)
				if err != nil {
					t.Fatalf("GetTreeCoverage() error = %v", err)
				}

				if coverage != tt.expected {
					t.Errorf("GetTreeCoverage(%.2f, %.2f) = %v, want %v (zone consistency check)",
						coord.lat, coord.lon, coverage, tt.expected)
				}
			}
		})
	}
}

func TestTreeCoverClient_BoundaryConditions(t *testing.T) {
	client := NewTreeCoverClient()
	ctx := context.Background()

	// Test boundary between Leavenworth zones
	coverage1, _ := client.GetTreeCoverage(ctx, 47.6, -120.81) // Just west of boundary
	coverage2, _ := client.GetTreeCoverage(ctx, 47.6, -120.79) // Just east of boundary

	if coverage1 == coverage2 {
		t.Errorf("Expected different coverage on opposite sides of zone boundary, got %v for both", coverage1)
	}

	if coverage1 != 60.0 {
		t.Errorf("West of -120.8 should be Icicle Creek (60%%), got %v", coverage1)
	}

	if coverage2 != 25.0 {
		t.Errorf("East of -120.8 should be Town Walls (25%%), got %v", coverage2)
	}
}
