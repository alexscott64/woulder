package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/alexscott64/woulder/backend/internal/database"
	"github.com/alexscott64/woulder/backend/internal/mountainproject"
	"github.com/alexscott64/woulder/backend/internal/service"
	"github.com/joho/godotenv"
)

func main() {
	// Flags
	//   --discover-areas: skip the full per-root recursive seed below and
	//     instead invoke ClimbTrackingService.SyncLocationAreaDiscovery
	//     once. This is the on-demand counterpart to the scheduled
	//     `location_area_discovery` job and is the recommended way to pull
	//     in newly-added MP sub-areas (e.g. Fantasia Boulders) without
	//     waiting for the weekly schedule.
	discoverAreas := flag.Bool("discover-areas", false,
		"Run only the location_area_discovery job (crawl configured roots to pick up new MP sub-areas) and exit")
	flag.Parse()

	log.Println("Starting Mountain Project climb data sync...")

	// Load environment variables from backend directory
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: .env file not found, using system environment variables")
	}

	// Initialize database
	db, err := database.New()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Mountain Project client
	mpClient := mountainproject.NewClient()

	// Initialize climb tracking service (no job monitor for manual sync)
	climbService := service.NewClimbTrackingService(db.MountainProject(), db.Climbing(), mpClient, nil)

	ctx := context.Background()

	if *discoverAreas {
		log.Println("Running one-shot location_area_discovery (new sub-area pickup)...")
		startTime := time.Now()
		if err := climbService.SyncLocationAreaDiscovery(ctx); err != nil {
			log.Printf("location_area_discovery returned error: %v", err)
			log.Printf("Time elapsed: %s", time.Since(startTime).Round(time.Second))
			os.Exit(1)
		}
		log.Printf("location_area_discovery complete in %s", time.Since(startTime).Round(time.Second))
		return
	}

	// Default mode: full per-root recursive seed using the shared
	// LocationRoots() registry (single source of truth — also consumed by
	// the scheduled SyncLocationAreaDiscovery job).
	areaConfigs := service.LocationRoots()

	totalAreas := 0
	successCount := 0
	failCount := 0

	startTime := time.Now()

	// Process each location's areas
	for _, config := range areaConfigs {
		log.Printf("\n========================================")
		log.Printf("Processing location: %s (ID: %d)", config.LocationName, config.LocationID)
		log.Printf("========================================")

		// Local copy so we can take its address safely inside the loop.
		locationID := config.LocationID

		for _, areaID := range config.MPAreaIDs {
			totalAreas++
			log.Printf("\nSyncing area ID: %d for %s...", areaID, config.LocationName)

			// Convert int64 to string for API call
			areaIDStr := fmt.Sprintf("%d", areaID)
			err := climbService.SyncAreaRecursive(ctx, areaIDStr, &locationID)
			if err != nil {
				log.Printf("ERROR syncing area %d: %v", areaID, err)
				failCount++
				continue
			}

			successCount++
			log.Printf("✓ Successfully synced area %d", areaID)
		}
	}

	elapsed := time.Since(startTime)

	log.Printf("\n========================================")
	log.Printf("Sync Complete!")
	log.Printf("========================================")
	log.Printf("Total areas processed: %d", totalAreas)
	log.Printf("Successful: %d", successCount)
	log.Printf("Failed: %d", failCount)
	log.Printf("Time elapsed: %s", elapsed.Round(time.Second))
	log.Printf("========================================")

	if failCount > 0 {
		os.Exit(1)
	}
}
