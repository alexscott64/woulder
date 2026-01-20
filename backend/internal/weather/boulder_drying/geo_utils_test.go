package boulder_drying

import (
	"math"
	"testing"
)

func TestCalculateBoulderPositions(t *testing.T) {
	centerLat := 37.3276  // Buttermilks
	centerLon := -118.5757
	radius := DefaultRadiusDegrees

	t.Run("Single boulder at center", func(t *testing.T) {
		positions := CalculateBoulderPositions(centerLat, centerLon, 1, radius)

		if len(positions) != 1 {
			t.Errorf("Expected 1 position, got %d", len(positions))
		}

		if positions[0].Latitude != centerLat || positions[0].Longitude != centerLon {
			t.Errorf("Single boulder should be at center: got (%.6f, %.6f)", positions[0].Latitude, positions[0].Longitude)
		}

		if positions[0].Aspect != "S" {
			t.Errorf("Single boulder should face South, got %s", positions[0].Aspect)
		}
	})

	t.Run("Four boulders cardinal directions", func(t *testing.T) {
		positions := CalculateBoulderPositions(centerLat, centerLon, 4, radius)

		if len(positions) != 4 {
			t.Errorf("Expected 4 positions, got %d", len(positions))
		}

		// First boulder should be North
		if positions[0].Aspect != "N" {
			t.Errorf("First boulder should face North, got %s", positions[0].Aspect)
		}

		// Second boulder should be East
		if positions[1].Aspect != "E" {
			t.Errorf("Second boulder should face East, got %s", positions[1].Aspect)
		}

		// Third boulder should be South
		if positions[2].Aspect != "S" {
			t.Errorf("Third boulder should face South, got %s", positions[2].Aspect)
		}

		// Fourth boulder should be West
		if positions[3].Aspect != "W" {
			t.Errorf("Fourth boulder should face West, got %s", positions[3].Aspect)
		}
	})

	t.Run("Eight boulders all directions", func(t *testing.T) {
		positions := CalculateBoulderPositions(centerLat, centerLon, 8, radius)

		expected := []string{"N", "NE", "E", "SE", "S", "SW", "W", "NW"}
		for i, expectedAspect := range expected {
			if positions[i].Aspect != expectedAspect {
				t.Errorf("Boulder %d: expected %s, got %s", i, expectedAspect, positions[i].Aspect)
			}
		}
	})

	t.Run("Boulders are evenly distributed", func(t *testing.T) {
		positions := CalculateBoulderPositions(centerLat, centerLon, 12, radius)

		// All boulders should be approximately same distance from center
		for i, pos := range positions {
			dist := CalculateDistance(centerLat, centerLon, pos.Latitude, pos.Longitude)
			expectedDist := radius * 111.32 // degrees to km (approximate)

			if math.Abs(dist-expectedDist) > 0.05 { // 50 meter tolerance
				t.Errorf("Boulder %d distance %.3f km differs from expected %.3f km", i, dist, expectedDist)
			}
		}
	})

	t.Run("Zero boulders returns empty array", func(t *testing.T) {
		positions := CalculateBoulderPositions(centerLat, centerLon, 0, radius)

		if len(positions) != 0 {
			t.Errorf("Expected empty array for 0 boulders, got %d", len(positions))
		}
	})
}

func TestAngleToAspect(t *testing.T) {
	tests := []struct {
		angle    float64 // radians
		expected string
	}{
		{0.0, "N"},
		{math.Pi / 4, "NE"},     // 45°
		{math.Pi / 2, "E"},      // 90°
		{3 * math.Pi / 4, "SE"}, // 135°
		{math.Pi, "S"},          // 180°
		{5 * math.Pi / 4, "SW"}, // 225°
		{3 * math.Pi / 2, "W"},  // 270°
		{7 * math.Pi / 4, "NW"}, // 315°
		{2 * math.Pi, "N"},      // 360° = 0°
	}

	for _, tt := range tests {
		result := AngleToAspect(tt.angle)
		if result != tt.expected {
			degrees := tt.angle * 180 / math.Pi
			t.Errorf("Angle %.0f°: expected %s, got %s", degrees, tt.expected, result)
		}
	}
}

func TestAspectToDegrees(t *testing.T) {
	tests := []struct {
		aspect   string
		expected float64
	}{
		{"N", 0.0},
		{"NE", 45.0},
		{"E", 90.0},
		{"SE", 135.0},
		{"S", 180.0},
		{"SW", 225.0},
		{"W", 270.0},
		{"NW", 315.0},
		{"Invalid", 0.0}, // Default to North
	}

	for _, tt := range tests {
		result := AspectToDegrees(tt.aspect)
		if result != tt.expected {
			t.Errorf("Aspect %s: expected %.0f°, got %.0f°", tt.aspect, tt.expected, result)
		}
	}
}

func TestCalculateDistance(t *testing.T) {
	// Test known distances
	tests := []struct {
		name     string
		lat1     float64
		lon1     float64
		lat2     float64
		lon2     float64
		expected float64 // km
		tolerance float64
	}{
		{
			name:      "Same point",
			lat1:      37.3276,
			lon1:      -118.5757,
			lat2:      37.3276,
			lon2:      -118.5757,
			expected:  0.0,
			tolerance: 0.001,
		},
		{
			name:      "Buttermilks to Peabody (approx 1km)",
			lat1:      37.3276,
			lon1:      -118.5757,
			lat2:      37.3365,
			lon2:      -118.5757,
			expected:  1.0,
			tolerance: 0.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateDistance(tt.lat1, tt.lon1, tt.lat2, tt.lon2)
			if math.Abs(result-tt.expected) > tt.tolerance {
				t.Errorf("Expected ~%.2f km, got %.2f km", tt.expected, result)
			}
		})
	}
}

func TestCalculateRadiusForArea(t *testing.T) {
	tests := []struct {
		boulderCount int
		expectedMin  float64
		expectedMax  float64
	}{
		{1, DefaultRadiusDegrees * 0.9, DefaultRadiusDegrees * 1.1},
		{5, DefaultRadiusDegrees * 0.9, DefaultRadiusDegrees * 1.1},
		{10, DefaultRadiusDegrees * 1.4, DefaultRadiusDegrees * 1.6},
		{50, DefaultRadiusDegrees * 1.9, DefaultRadiusDegrees * 2.1},
		{100, DefaultRadiusDegrees * 2.9, DefaultRadiusDegrees * 3.1},
	}

	for _, tt := range tests {
		result := CalculateRadiusForArea(tt.boulderCount)
		if result < tt.expectedMin || result > tt.expectedMax {
			t.Errorf("Boulder count %d: expected radius %.4f-%.4f, got %.4f",
				tt.boulderCount, tt.expectedMin, tt.expectedMax, result)
		}
	}
}

func BenchmarkCalculateBoulderPositions(b *testing.B) {
	centerLat := 37.3276
	centerLon := -118.5757
	radius := DefaultRadiusDegrees

	for i := 0; i < b.N; i++ {
		CalculateBoulderPositions(centerLat, centerLon, 50, radius)
	}
}
