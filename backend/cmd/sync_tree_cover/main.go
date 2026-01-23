package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/alexscott64/woulder/backend/internal/weather/boulder_drying"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Route struct {
	MPRouteID  string
	Name       string
	Latitude   float64
	Longitude  float64
	LocationID int
}

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	log.Println("=== Boulder Tree Coverage Sync Tool ===")
	log.Println("This tool fetches and stores tree coverage data for all boulders with GPS coordinates")
	log.Println()

	// Connect to database
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		getEnvOrDefault("DB_PORT", "5432"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		getEnvOrDefault("DB_SSLMODE", "require"),
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("✓ Database connected")

	// Initialize tree cover client
	treeClient := boulder_drying.NewTreeCoverClient()
	if !treeClient.IsEnabled() {
		log.Println("Warning: Google Earth Engine not configured - will use location-based estimates only")
	} else {
		log.Println("✓ Google Earth Engine client initialized")
	}

	// Get all routes with GPS coordinates
	log.Println()
	log.Println("Fetching routes with GPS coordinates...")
	routes, err := getRoutesWithGPS(db)
	if err != nil {
		log.Fatalf("Failed to fetch routes: %v", err)
	}

	log.Printf("Found %d routes with GPS coordinates", len(routes))
	if len(routes) == 0 {
		log.Println("No routes found. Make sure Mountain Project sync has run with GPS distribution.")
		return
	}

	// Process routes
	log.Println()
	log.Println("Syncing tree coverage...")
	log.Println()

	successCount := 0
	skipCount := 0
	errorCount := 0

	for i, route := range routes {
		// Check if tree coverage already exists
		var existingCoverage *float64
		err := db.QueryRow(`
			SELECT tree_coverage_percent
			FROM woulder.boulder_drying_profiles
			WHERE mp_route_id = $1
		`, route.MPRouteID).Scan(&existingCoverage)

		if err == nil && existingCoverage != nil {
			skipCount++
			if i%10 == 0 {
				log.Printf("[%d/%d] Skipping %s (already has tree coverage: %.1f%%)",
					i+1, len(routes), route.Name, *existingCoverage)
			}
			continue
		}

		// Fetch tree coverage for this route
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		treeCoverage, err := treeClient.GetTreeCoverage(ctx, route.Latitude, route.Longitude)
		cancel()

		if err != nil {
			log.Printf("[%d/%d] ✗ Error fetching tree coverage for %s: %v",
				i+1, len(routes), route.Name, err)
			errorCount++
			continue
		}

		// Save or update boulder drying profile
		_, err = db.Exec(`
			INSERT INTO woulder.boulder_drying_profiles
				(mp_route_id, tree_coverage_percent, created_at, updated_at)
			VALUES ($1, $2, NOW(), NOW())
			ON CONFLICT (mp_route_id)
			DO UPDATE SET
				tree_coverage_percent = EXCLUDED.tree_coverage_percent,
				updated_at = NOW()
		`, route.MPRouteID, treeCoverage)

		if err != nil {
			log.Printf("[%d/%d] ✗ Error saving profile for %s: %v",
				i+1, len(routes), route.Name, err)
			errorCount++
			continue
		}

		successCount++
		log.Printf("[%d/%d] ✓ %s (location %d): %.1f%% tree coverage",
			i+1, len(routes), route.Name, route.LocationID, treeCoverage)

		// Rate limiting: pause every 50 routes
		if (i+1)%50 == 0 && i+1 < len(routes) {
			log.Printf("... Processed %d routes, pausing 2 seconds ...", i+1)
			time.Sleep(2 * time.Second)
		}
	}

	// Print summary
	log.Println()
	log.Println("=== Sync Complete ===")
	log.Printf("Total routes: %d", len(routes))
	log.Printf("✓ Successfully synced: %d", successCount)
	log.Printf("⊙ Skipped (already exists): %d", skipCount)
	log.Printf("✗ Errors: %d", errorCount)
	log.Println()

	if errorCount > 0 {
		os.Exit(1)
	}
}

func getRoutesWithGPS(db *sql.DB) ([]Route, error) {
	query := `
		SELECT mp_route_id, name, latitude, longitude, location_id
		FROM woulder.mp_routes
		WHERE latitude IS NOT NULL
		  AND longitude IS NOT NULL
		ORDER BY location_id, name
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var routes []Route
	for rows.Next() {
		var r Route
		if err := rows.Scan(&r.MPRouteID, &r.Name, &r.Latitude, &r.Longitude, &r.LocationID); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		routes = append(routes, r)
	}

	return routes, rows.Err()
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
