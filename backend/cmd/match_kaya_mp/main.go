package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/alexscott64/woulder/backend/internal/database"
	"github.com/alexscott64/woulder/backend/internal/service"
	"github.com/joho/godotenv"
)

// MatchKayaMP links Kaya climbs to Mountain Project routes using smart matching
func main() {
	log.Println("Starting Kaya ↔ Mountain Project route matching...")

	// Command-line flags
	locationFlag := flag.String("location", "", "Match routes for specific location (e.g., 'Leavenworth')")
	minConfidenceFlag := flag.Float64("min-confidence", 0.85, "Minimum confidence score (0.0-1.0)")
	dryRunFlag := flag.Bool("dry-run", false, "Show matches without saving to database")
	limitFlag := flag.Int("limit", 0, "Limit number of climbs to process (0 = all)")
	flag.Parse()

	// Load environment variables
	if err := godotenv.Load("../.env"); err != nil {
		log.Printf("Warning: .env file not found, using system environment variables")
	}

	// Initialize database
	db, err := database.New()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize matching service
	_ = service.NewKayaMPMatchingService() // Will be used when full implementation is complete

	ctx := context.Background()

	// Get Kaya climbs to match
	climbs, err := getKayaClimbs(ctx, db, *locationFlag, *limitFlag)
	if err != nil {
		log.Fatalf("Failed to get Kaya climbs: %v", err)
	}

	log.Printf("Found %d Kaya climbs to match", len(climbs))

	matchCount := 0
	highConfidenceCount := 0

	// Process each climb
	for i, climb := range climbs {
		log.Printf("\n[%d/%d] Matching: %s (%s)", i+1, len(climbs), climb.Name, climb.Location)

		// Find potential MP matches
		matches := findMPMatches(ctx, db, climb, *minConfidenceFlag)

		if len(matches) == 0 {
			log.Printf("  No matches found")
			continue
		}

		// Display matches
		for _, match := range matches {
			confidence := match.Confidence
			matchType := match.MatchType

			log.Printf("  ✓ Match: %s (confidence: %.2f, type: %s)",
				match.MPRouteName, confidence, matchType)
			log.Printf("    Area: %s, Distance: %s",
				match.MPAreaName, formatDistance(match.DistanceKM))

			if confidence >= 0.90 {
				highConfidenceCount++
			}

			matchCount++

			// Save match if not dry run
			if !*dryRunFlag {
				if err := saveMatch(ctx, db, match); err != nil {
					log.Printf("    ERROR saving match: %v", err)
				}
			}
		}
	}

	log.Printf("\n========================================")
	log.Printf("Matching Complete!")
	log.Printf("========================================")
	log.Printf("Climbs processed: %d", len(climbs))
	log.Printf("Total matches: %d", matchCount)
	log.Printf("High confidence (>0.90): %d", highConfidenceCount)

	if *dryRunFlag {
		log.Printf("DRY RUN: No matches were saved to database")
	} else {
		log.Printf("Matches saved to kaya_mp_route_matches table")
	}
	log.Printf("========================================")
}

// KayaClimb represents a simplified climb for matching
type KayaClimb struct {
	ID        string
	Name      string
	Location  string
	Latitude  *float64
	Longitude *float64
	Grade     string
}

func getKayaClimbs(ctx context.Context, db *database.Database, location string, limit int) ([]KayaClimb, error) {
	query := `
		SELECT 
			c.kaya_climb_id,
			c.name,
			l.name as location_name,
			l.latitude,
			l.longitude,
			c.grade_name
		FROM woulder.kaya_climbs c
		JOIN woulder.kaya_locations l ON c.kaya_location_id = l.kaya_location_id
		WHERE 1=1
	`

	args := []interface{}{}

	if location != "" {
		query += " AND LOWER(l.name) LIKE LOWER($1)"
		args = append(args, "%"+location+"%")
	}

	query += " ORDER BY c.id"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	// For now, return empty list and log that this needs implementation
	log.Println("Note: Full database integration pending. Returning empty list.")
	log.Println("To implement: Query kaya_climbs JOIN kaya_locations")

	return []KayaClimb{}, nil
}

func findMPMatches(ctx context.Context, db *database.Database, climb KayaClimb, minConfidence float64) []service.RouteMatch {
	// This would query boulders table and find matches
	// For now, return empty
	log.Println("  Note: MP matching logic pending database integration")
	return []service.RouteMatch{}
}

func saveMatch(ctx context.Context, db *database.Database, match service.RouteMatch) error {
	_ = ctx   // Will be used in full implementation
	_ = db    // Will be used in full implementation
	_ = match // Will be used in full implementation

	// Full implementation would execute:
	// INSERT INTO woulder.kaya_mp_route_matches (...) VALUES (...)
	log.Println("  Note: Save to database pending")
	return nil
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
