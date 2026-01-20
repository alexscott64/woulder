package boulder_drying

import (
	"math"
)

const (
	// Default radius for boulder distribution (in degrees)
	// ~100 meters at mid-latitudes ≈ 0.0009 degrees
	DefaultRadiusDegrees = 0.0009

	// Earth's radius in kilometers for distance calculations
	EarthRadiusKm = 6371.0
)

// BoulderPosition represents the calculated GPS position and aspect of a boulder
type BoulderPosition struct {
	Latitude  float64
	Longitude float64
	Aspect    string // Cardinal direction: N, NE, E, SE, S, SW, W, NW
}

// CalculateBoulderPositions distributes boulders in an even circle around area center
// Routes are ordered left-to-right based on their index in the children array
// Returns array of BoulderPosition structs with GPS coordinates and aspect
func CalculateBoulderPositions(
	centerLat, centerLon float64,
	totalBoulders int,
	radiusDegrees float64,
) []BoulderPosition {
	if totalBoulders == 0 {
		return []BoulderPosition{}
	}

	// If only one boulder, place it at the center facing south
	if totalBoulders == 1 {
		return []BoulderPosition{{
			Latitude:  centerLat,
			Longitude: centerLon,
			Aspect:    "S",
		}}
	}

	positions := make([]BoulderPosition, totalBoulders)

	for i := 0; i < totalBoulders; i++ {
		// Calculate angle for this boulder (0° = North, clockwise)
		// Distribute evenly around the circle
		angle := (2.0 * math.Pi * float64(i)) / float64(totalBoulders)

		// Calculate latitude and longitude offsets
		// North/South: latitude offset = radius * cos(angle)
		// East/West: longitude offset = radius * sin(angle) / cos(centerLat)
		latOffset := radiusDegrees * math.Cos(angle)
		lonOffset := radiusDegrees * math.Sin(angle) / math.Cos(centerLat*math.Pi/180.0)

		positions[i] = BoulderPosition{
			Latitude:  centerLat + latOffset,
			Longitude: centerLon + lonOffset,
			Aspect:    AngleToAspect(angle),
		}
	}

	return positions
}

// AngleToAspect converts an angle in radians (0° = North, clockwise) to cardinal direction
// Returns one of: N, NE, E, SE, S, SW, W, NW
func AngleToAspect(angleRadians float64) string {
	// Normalize angle to 0-2π range
	angle := math.Mod(angleRadians, 2*math.Pi)
	if angle < 0 {
		angle += 2 * math.Pi
	}

	// Convert to degrees for easier calculation
	degrees := angle * 180.0 / math.Pi

	// Map to 8 cardinal directions (45° each)
	// 0° = N, 45° = NE, 90° = E, 135° = SE, 180° = S, 225° = SW, 270° = W, 315° = NW
	switch {
	case degrees < 22.5 || degrees >= 337.5:
		return "N"
	case degrees >= 22.5 && degrees < 67.5:
		return "NE"
	case degrees >= 67.5 && degrees < 112.5:
		return "E"
	case degrees >= 112.5 && degrees < 157.5:
		return "SE"
	case degrees >= 157.5 && degrees < 202.5:
		return "S"
	case degrees >= 202.5 && degrees < 247.5:
		return "SW"
	case degrees >= 247.5 && degrees < 292.5:
		return "W"
	case degrees >= 292.5 && degrees < 337.5:
		return "NW"
	default:
		return "N" // Fallback
	}
}

// AspectToDegrees converts a cardinal direction to degrees (0° = North, clockwise)
// Inverse of AngleToAspect for use in sun exposure calculations
func AspectToDegrees(aspect string) float64 {
	switch aspect {
	case "N":
		return 0.0
	case "NE":
		return 45.0
	case "E":
		return 90.0
	case "SE":
		return 135.0
	case "S":
		return 180.0
	case "SW":
		return 225.0
	case "W":
		return 270.0
	case "NW":
		return 315.0
	default:
		return 0.0 // Default to North
	}
}

// CalculateDistance calculates the great-circle distance between two GPS coordinates
// Returns distance in kilometers using the Haversine formula
func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// Convert to radians
	lat1Rad := lat1 * math.Pi / 180.0
	lat2Rad := lat2 * math.Pi / 180.0
	deltaLat := (lat2 - lat1) * math.Pi / 180.0
	deltaLon := (lon2 - lon1) * math.Pi / 180.0

	// Haversine formula
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return EarthRadiusKm * c
}

// CalculateRadiusForArea calculates an appropriate radius based on number of boulders
// More boulders = larger radius to avoid overlap
// Returns radius in degrees
func CalculateRadiusForArea(boulderCount int) float64 {
	if boulderCount <= 5 {
		return DefaultRadiusDegrees // ~100 meters
	} else if boulderCount <= 20 {
		return DefaultRadiusDegrees * 1.5 // ~150 meters
	} else if boulderCount <= 50 {
		return DefaultRadiusDegrees * 2.0 // ~200 meters
	} else {
		return DefaultRadiusDegrees * 3.0 // ~300 meters
	}
}
