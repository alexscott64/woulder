package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// KayaClimb represents a simplified climb for matching
type KayaClimb struct {
	ID        string
	Name      string
	Location  string
	Latitude  *float64
	Longitude *float64
	Grade     string
}

// MPRoute represents a Mountain Project route
type MPRoute struct {
	ID        int64
	Name      string
	Area      string
	Latitude  *float64
	Longitude *float64
}

// RouteMatch represents a potential match
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

func main() {
	log.Println("Starting Kaya ↔ Mountain Project route matching...")

	// Command-line flags
	locationFlag := flag.String("location", "", "Match routes for specific location (e.g., 'Leavenworth')")
	minConfidenceFlag := flag.Float64("min-confidence", 0.75, "Minimum confidence score (0.0-1.0)")
	dryRunFlag := flag.Bool("dry-run", false, "Show matches without saving to database")
	limitFlag := flag.Int("limit", 0, "Limit number of climbs to process (0 = all)")
	flag.Parse()

	// Load environment variables - try current directory first, then parent
	if err := godotenv.Load(".env"); err != nil {
		if err := godotenv.Load("../.env"); err != nil {
			log.Printf("Warning: .env file not found in . or .., using system environment variables")
		}
	}

	// Create raw SQL connection for matching queries
	sqlDB, err := createSQLConnection()
	if err != nil {
		log.Fatalf("Failed to create SQL connection: %v", err)
	}
	defer sqlDB.Close()

	ctx := context.Background()

	log.Printf("Configuration:")
	log.Printf("  - Location filter: %s", func() string {
		if *locationFlag != "" {
			return *locationFlag
		}
		return "All locations"
	}())
	log.Printf("  - Min confidence: %.2f", *minConfidenceFlag)
	log.Printf("  - Dry run: %v", *dryRunFlag)
	log.Printf("  - Limit: %d", *limitFlag)
	log.Println()

	// Get Kaya climbs to match
	climbs, err := getKayaClimbs(ctx, sqlDB, *locationFlag, *limitFlag)
	if err != nil {
		log.Fatalf("Failed to get Kaya climbs: %v", err)
	}

	log.Printf("Found %d Kaya climbs to match", len(climbs))

	if len(climbs) == 0 {
		log.Println("No climbs found. Ensure you've run the Kaya sync first.")
		return
	}

	matchCount := 0
	highConfidenceCount := 0

	// Process each climb
	for i, climb := range climbs {
		if i > 0 && i%10 == 0 {
			log.Printf("Progress: %d/%d climbs processed...", i, len(climbs))
		}

		// Find potential MP matches
		matches := findMPMatches(ctx, sqlDB, climb, *minConfidenceFlag)

		if len(matches) == 0 {
			continue
		}

		// Display matches
		for _, match := range matches {
			log.Printf("\n[%d/%d] Match found:", i+1, len(climbs))
			log.Printf("  Kaya: %s (%s)", match.KayaClimbName, match.KayaLocationName)
			log.Printf("  MP:   %s (%s)", match.MPRouteName, match.MPAreaName)
			log.Printf("  Confidence: %.2f | Type: %s", match.Confidence, match.MatchType)
			log.Printf("  Name similarity: %.2f | Distance: %s",
				match.NameSimilarity, formatDistance(match.DistanceKM))

			if match.Confidence >= 0.90 {
				highConfidenceCount++
			}

			matchCount++

			// Save match if not dry run
			if !*dryRunFlag {
				if err := saveMatch(ctx, sqlDB, match); err != nil {
					log.Printf("  ERROR saving match: %v", err)
				} else {
					log.Printf("  ✓ Saved to database")
				}
			}
		}
	}

	log.Printf("\n========================================")
	log.Printf("Matching Complete!")
	log.Printf("========================================")
	log.Printf("Climbs processed: %d", len(climbs))
	log.Printf("Total matches: %d", matchCount)
	log.Printf("High confidence (≥0.90): %d", highConfidenceCount)

	if *dryRunFlag {
		log.Printf("DRY RUN: No matches were saved to database")
	} else {
		log.Printf("Matches saved to kaya_mp_route_matches table")
	}
	log.Printf("========================================")
}

func getKayaClimbs(ctx context.Context, db *sql.DB, location string, limit int) ([]KayaClimb, error) {
	query := `
		SELECT
			c.slug,
			c.name,
			COALESCE(c.kaya_destination_name, c.kaya_area_name, 'Unknown') as location_name,
			l.latitude,
			l.longitude,
			c.grade_name
		FROM kaya_climbs c
		LEFT JOIN kaya_locations l ON c.kaya_destination_id = l.kaya_location_id
		WHERE c.slug IS NOT NULL
			AND c.slug != ''
	`

	args := []interface{}{}
	argNum := 1

	if location != "" {
		query += fmt.Sprintf(" AND (LOWER(c.kaya_destination_name) LIKE LOWER($%d) OR LOWER(c.kaya_area_name) LIKE LOWER($%d))", argNum, argNum)
		args = append(args, "%"+location+"%")
		argNum++
	}

	query += " ORDER BY c.kaya_destination_name, c.name"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query climbs: %w", err)
	}
	defer rows.Close()

	var climbs []KayaClimb
	for rows.Next() {
		var climb KayaClimb
		var lat, lon sql.NullFloat64
		var grade sql.NullString

		err := rows.Scan(&climb.ID, &climb.Name, &climb.Location, &lat, &lon, &grade)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if lat.Valid {
			climb.Latitude = &lat.Float64
		}
		if lon.Valid {
			climb.Longitude = &lon.Float64
		}
		if grade.Valid {
			climb.Grade = grade.String
		}

		climbs = append(climbs, climb)
	}

	return climbs, rows.Err()
}

func findMPMatches(ctx context.Context, db *sql.DB, climb KayaClimb, minConfidence float64) []RouteMatch {
	// Query MP routes with similar names, joining with areas to get area name
	query := `
		SELECT
			r.mp_route_id,
			r.name,
			COALESCE(a.name, 'Unknown') as area_name,
			a.latitude,
			a.longitude
		FROM woulder.mp_routes r
		LEFT JOIN woulder.mp_areas a ON r.mp_area_id = a.mp_area_id
		WHERE LOWER(r.name) LIKE LOWER($1)
		LIMIT 20
	`

	rows, err := db.QueryContext(ctx, query, "%"+climb.Name+"%")
	if err != nil {
		log.Printf("  Error querying MP routes: %v", err)
		return []RouteMatch{}
	}
	defer rows.Close()

	var matches []RouteMatch
	for rows.Next() {
		var mpID string
		var mpName, mpArea string
		var mpLat, mpLon sql.NullFloat64

		if err := rows.Scan(&mpID, &mpName, &mpArea, &mpLat, &mpLon); err != nil {
			continue
		}

		// Convert MP route ID to int64
		var mpIDInt int64
		if _, err := fmt.Sscanf(mpID, "%d", &mpIDInt); err != nil {
			continue
		}

		// Calculate name similarity
		nameSim := calculateNameSimilarity(climb.Name, mpName)

		// Check location name match
		locationMatch := matchLocationNames(climb.Location, mpArea)

		// Calculate GPS distance if both have coordinates
		var distKM *float64
		if climb.Latitude != nil && climb.Longitude != nil && mpLat.Valid && mpLon.Valid {
			dist := calculateGPSDistance(*climb.Latitude, *climb.Longitude, mpLat.Float64, mpLon.Float64)
			distKM = &dist
		}

		// Calculate overall confidence
		confidence := calculateMatchConfidence(nameSim, locationMatch, distKM)

		// Determine match type
		matchType := determineMatchType(nameSim, locationMatch, distKM)

		if confidence >= minConfidence {
			matches = append(matches, RouteMatch{
				KayaClimbID:       climb.ID,
				KayaClimbName:     climb.Name,
				KayaLocationName:  climb.Location,
				MPRouteID:         mpIDInt,
				MPRouteName:       mpName,
				MPAreaName:        mpArea,
				Confidence:        confidence,
				MatchType:         matchType,
				NameSimilarity:    nameSim,
				DistanceKM:        distKM,
				LocationNameMatch: locationMatch,
			})
		}
	}

	return matches
}

func saveMatch(ctx context.Context, db *sql.DB, match RouteMatch) error {
	query := `
		INSERT INTO kaya_mp_route_matches (
			kaya_climb_id, mp_route_id, match_confidence, match_type,
			kaya_climb_name, kaya_location_name,
			mp_route_name, mp_area_name,
			name_similarity, location_name_match, location_distance_km
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (kaya_climb_id, mp_route_id) DO UPDATE SET
			match_confidence = EXCLUDED.match_confidence,
			match_type = EXCLUDED.match_type,
			name_similarity = EXCLUDED.name_similarity,
			location_name_match = EXCLUDED.location_name_match,
			location_distance_km = EXCLUDED.location_distance_km,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := db.ExecContext(ctx, query,
		match.KayaClimbID,
		match.MPRouteID,
		match.Confidence,
		match.MatchType,
		match.KayaClimbName,
		match.KayaLocationName,
		match.MPRouteName,
		match.MPAreaName,
		match.NameSimilarity,
		match.LocationNameMatch,
		match.DistanceKM,
	)

	return err
}

// calculateNameSimilarity computes similarity between two route names using Levenshtein distance
func calculateNameSimilarity(name1, name2 string) float64 {
	// Normalize names
	n1 := normalizeRouteName(name1)
	n2 := normalizeRouteName(name2)

	if n1 == n2 {
		return 1.0
	}

	// Calculate Levenshtein distance
	distance := levenshteinDistance(n1, n2)
	maxLen := float64(max(len(n1), len(n2)))

	if maxLen == 0 {
		return 0.0
	}

	return 1.0 - (float64(distance) / maxLen)
}

// normalizeRouteName standardizes route names for comparison
func normalizeRouteName(name string) string {
	name = strings.ToLower(name)
	name = strings.TrimPrefix(name, "the ")
	name = strings.TrimPrefix(name, "a ")

	// Remove special characters but keep spaces
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == ' ' {
			return r
		}
		return -1
	}, name)

	return strings.TrimSpace(strings.Join(strings.Fields(name), " "))
}

// levenshteinDistance calculates edit distance
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,
				min(matrix[i][j-1]+1, matrix[i-1][j-1]+cost),
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

// matchLocationNames checks if location names match
func matchLocationNames(kayaLocation, mpArea string) bool {
	kayaLower := strings.ToLower(strings.TrimSpace(kayaLocation))
	mpLower := strings.ToLower(strings.TrimSpace(mpArea))

	return strings.Contains(mpLower, kayaLower) || strings.Contains(kayaLower, mpLower)
}

// calculateGPSDistance computes distance in kilometers using Haversine formula
func calculateGPSDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKm = 6371.0

	dLat := toRadians(lat2 - lat1)
	dLon := toRadians(lon2 - lon1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRadians(lat1))*math.Cos(toRadians(lat2))*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

// calculateMatchConfidence computes overall match confidence score
func calculateMatchConfidence(nameSim float64, locationMatch bool, distanceKM *float64) float64 {
	confidence := nameSim * 0.7 // Name similarity weighted 70%

	if locationMatch {
		confidence += 0.2 // Location match adds 20%
	}

	if distanceKM != nil && *distanceKM < 5.0 {
		proximityBonus := 0.1 * (1.0 - (*distanceKM / 5.0))
		confidence += proximityBonus
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// determineMatchType classifies the match
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

func toRadians(degrees float64) float64 {
	return degrees * math.Pi / 180.0
}

func createSQLConnection() (*sql.DB, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	if port == "" {
		port = "5432"
	}
	if sslmode == "" {
		sslmode = "require"
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)

	return sql.Open("postgres", connStr)
}

func formatDistance(distKM *float64) string {
	if distKM == nil {
		return "N/A"
	}
	if *distKM < 1.0 {
		return fmt.Sprintf("%.0fm", *distKM*1000)
	}
	return fmt.Sprintf("%.1fkm", *distKM)
}

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
