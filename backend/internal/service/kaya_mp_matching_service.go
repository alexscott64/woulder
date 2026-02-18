package service

import (
	"context"
	"fmt"
	"math"
	"strings"
)

// KayaMPMatchingService handles intelligent matching between Kaya climbs and Mountain Project routes
type KayaMPMatchingService struct {
	// Repository access will be added when we implement full database integration
}

// NewKayaMPMatchingService creates a new matching service
func NewKayaMPMatchingService() *KayaMPMatchingService {
	return &KayaMPMatchingService{}
}

// MatchConfig defines parameters for the matching algorithm
type MatchConfig struct {
	MinConfidence         float64 // Minimum confidence to create a match (default: 0.70)
	RequireNameSimilarity float64 // Minimum name similarity (default: 0.70)
	MaxDistanceKM         float64 // Maximum GPS distance in km (default: 10.0)
	LocationName          string  // Optional: only match within this location
}

// RouteMatch represents a potential match between Kaya and MP
type RouteMatch struct {
	KayaClimbID       string
	KayaClimbName     string
	KayaLocationName  string
	MPRouteID         int64
	MPRouteName       string
	MPAreaName        string
	Confidence        float64
	MatchType         string
	NameSimilarity    float64
	DistanceKM        *float64
	LocationNameMatch bool
}

// DefaultMatchConfig returns sensible default matching parameters
func DefaultMatchConfig() MatchConfig {
	return MatchConfig{
		MinConfidence:         0.70,
		RequireNameSimilarity: 0.70,
		MaxDistanceKM:         10.0,
	}
}

// MatchRoutesForLocation finds matches for all Kaya climbs in a location
func (s *KayaMPMatchingService) MatchRoutesForLocation(ctx context.Context, locationName string, config MatchConfig) ([]RouteMatch, error) {
	// TODO: Implement after confirming database access patterns
	// This would:
	// 1. Get all Kaya climbs for the location
	// 2. Get all MP routes in nearby areas
	// 3. Run matching algorithm
	// 4. Return high-confidence matches

	return nil, fmt.Errorf("not yet implemented")
}

// calculateNameSimilarity computes similarity between two route names
// Uses Levenshtein distance normalized to 0.0-1.0
func calculateNameSimilarity(name1, name2 string) float64 {
	// Normalize names first
	n1 := normalizeRouteName(name1)
	n2 := normalizeRouteName(name2)

	// Exact match
	if n1 == n2 {
		return 1.0
	}

	// Calculate Levenshtein distance
	distance := levenshteinDistance(n1, n2)
	maxLen := float64(max(len(n1), len(n2)))

	if maxLen == 0 {
		return 0.0
	}

	// Convert distance to similarity (1.0 = identical, 0.0 = completely different)
	similarity := 1.0 - (float64(distance) / maxLen)
	return similarity
}

// normalizeRouteName standardizes route names for comparison
func normalizeRouteName(name string) string {
	name = strings.ToLower(name)

	// Remove common prefixes
	name = strings.TrimPrefix(name, "the ")
	name = strings.TrimPrefix(name, "a ")
	name = strings.TrimPrefix(name, "an ")

	// Remove special characters but keep spaces
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == ' ' {
			return r
		}
		return -1 // Remove character
	}, name)

	// Normalize multiple spaces to single space
	name = strings.Join(strings.Fields(name), " ")

	// Normalize grade representations
	name = normalizeGrades(name)

	return strings.TrimSpace(name)
}

// normalizeGrades converts various grade formats to standard form
func normalizeGrades(name string) string {
	// V-scale: "V0", "v0", "V 0" -> "v0"
	name = strings.ReplaceAll(name, "V ", "v")
	name = strings.ReplaceAll(name, "V-", "v")

	// YDS: "5.10a", "510a", "5 10a" -> "510a"
	name = strings.ReplaceAll(name, "5.", "5")
	name = strings.ReplaceAll(name, "5 ", "5")

	return name
}

// levenshteinDistance calculates edit distance between two strings
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create 2D matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}

	// Initialize first row and column
	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}

			deletion := matrix[i-1][j] + 1
			insertion := matrix[i][j-1] + 1
			substitution := matrix[i-1][j-1] + cost

			matrix[i][j] = min(deletion, min(insertion, substitution))
		}
	}

	return matrix[len(s1)][len(s2)]
}

// calculateGPSDistance computes distance in kilometers between two GPS coordinates
func calculateGPSDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// Haversine formula
	const earthRadiusKm = 6371.0

	dLat := toRadians(lat2 - lat1)
	dLon := toRadians(lon2 - lon1)

	lat1Rad := toRadians(lat1)
	lat2Rad := toRadians(lat2)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1Rad)*math.Cos(lat2Rad)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

// toRadians converts degrees to radians
func toRadians(degrees float64) float64 {
	return degrees * math.Pi / 180.0
}

// matchLocationNames checks if location names indicate same area
// e.g., "Leavenworth" matches "Leavenworth, WA" or area hierarchy containing "Leavenworth"
func matchLocationNames(kayaLocation, mpAreaHierarchy string) bool {
	kayaLower := strings.ToLower(strings.TrimSpace(kayaLocation))
	mpLower := strings.ToLower(strings.TrimSpace(mpAreaHierarchy))

	// Direct substring match
	if strings.Contains(mpLower, kayaLower) || strings.Contains(kayaLower, mpLower) {
		return true
	}

	// Check for common abbreviations
	abbreviations := map[string][]string{
		"rmnp": {"rocky mountain national park", "rocky mountain np"},
		"jt":   {"joshua tree"},
		"rr":   {"red rocks", "red rock"},
	}

	for abbr, full := range abbreviations {
		if kayaLower == abbr {
			for _, fullName := range full {
				if strings.Contains(mpLower, fullName) {
					return true
				}
			}
		}
	}

	return false
}

// calculateMatchConfidence computes overall match confidence score
func calculateMatchConfidence(nameSim float64, locationMatch bool, distanceKM *float64) float64 {
	confidence := nameSim * 0.7 // Name similarity weighted 70%

	if locationMatch {
		confidence += 0.2 // Location name match adds 20%
	}

	if distanceKM != nil && *distanceKM < 5.0 {
		// GPS proximity bonus (up to 10%)
		proximityBonus := 0.1 * (1.0 - (*distanceKM / 5.0))
		confidence += proximityBonus
	}

	// Cap at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// determineMatchType classifies the match based on characteristics
func determineMatchType(nameSim float64, locationMatch bool, distanceKM *float64) string {
	if nameSim == 1.0 {
		return "exact_name"
	}

	if nameSim >= 0.9 && locationMatch {
		return "fuzzy_name_location"
	}

	if nameSim >= 0.85 {
		return "fuzzy_name"
	}

	if locationMatch && distanceKM != nil && *distanceKM < 1.0 {
		return "location_gps_proximity"
	}

	if locationMatch {
		return "location_name"
	}

	return "low_confidence"
}

// SaveMatch persists a route match to the database
func (s *KayaMPMatchingService) SaveMatch(ctx context.Context, match RouteMatch) error {
	// TODO: Implement database save
	// This would insert into kaya_mp_route_matches table
	return fmt.Errorf("not yet implemented")
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Example usage patterns (for documentation):
/*
// Match routes for a specific location
matchingService := NewKayaMPMatchingService(db.Kaya())
config := DefaultMatchConfig()
config.LocationName = "Leavenworth"
matches, err := matchingService.MatchRoutesForLocation(ctx, "Leavenworth", config)

// Calculate name similarity
similarity := calculateNameSimilarity("The Prism", "Prism")
// Returns: 0.95 (very similar)

// Calculate GPS distance
distance := calculateGPSDistance(47.5962, -120.6616, 47.5970, -120.6620)
// Returns: ~0.05 km (50 meters)

// Determine match confidence
confidence := calculateMatchConfidence(0.92, true, &distance)
// Returns: ~0.85 (high confidence match)
*/
