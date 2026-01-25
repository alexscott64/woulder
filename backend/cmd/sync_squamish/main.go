package main

import (
	"context"
	"log"
	"time"

	"github.com/alexscott64/woulder/backend/internal/database"
	"github.com/alexscott64/woulder/backend/internal/mountainproject"
	"github.com/alexscott64/woulder/backend/internal/service"
	"github.com/joho/godotenv"
)

func main() {
	log.Println("Starting Squamish Mountain Project sync...")

	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// Initialize database
	db, err := database.New()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Mountain Project client
	mpClient := mountainproject.NewClient()

	// Initialize climb tracking service
	climbService := service.NewClimbTrackingService(db, mpClient)

	// Squamish configuration
	locationID := 6
	locationName := "Squamish"

	// ONLY the boulder areas we want:
	// - Grand Wall Boulders (112842712)
	// - North Wall Boulders (108506197)
	// - Apron Boulders (106025685)
	// - Paradise Valley Boulders (110937821) - has sub-areas
	// - Powerline Boulders (121199811)
	areaIDs := []string{
		"112842712", // Grand Wall Boulders
		"108506197", // North Wall Boulders
		"106025685", // Apron Boulders
		"110937821", // Paradise Valley Boulders
		"121199811", // Powerline Boulders
	}

	ctx := context.Background()
	startTime := time.Now()
	successCount := 0
	failCount := 0

	log.Printf("\n========================================")
	log.Printf("Processing location: %s (ID: %d)", locationName, locationID)
	log.Printf("Areas to sync: %d", len(areaIDs))
	log.Printf("========================================")

	for i, areaID := range areaIDs {
		log.Printf("\n[%d/%d] Syncing area ID: %s...", i+1, len(areaIDs), areaID)

		err := climbService.SyncAreaRecursive(ctx, areaID, &locationID)
		if err != nil {
			log.Printf("ERROR syncing area %s: %v", areaID, err)
			failCount++
			continue
		}

		successCount++
		log.Printf("✓ Successfully synced area %s", areaID)
	}

	elapsed := time.Since(startTime)

	log.Printf("\n========================================")
	log.Printf("Squamish Sync Complete!")
	log.Printf("========================================")
	log.Printf("Total areas processed: %d", len(areaIDs))
	log.Printf("Successful: %d", successCount)
	log.Printf("Failed: %d", failCount)
	log.Printf("Time elapsed: %s", elapsed.Round(time.Second))
	log.Printf("========================================")

	// Now run the check script to see what we have
	log.Println("\nChecking synced areas...")
	areas, err := db.GetAreasOrderedByActivity(ctx, locationID)
	if err != nil {
		log.Printf("Error fetching areas: %v", err)
	} else {
		log.Printf("\nFound %d root areas in database:\n", len(areas))
		for i, area := range areas {
			log.Printf("  %d. %s (ID: %s) - %d routes, %d ticks",
				i+1, area.Name, area.MPAreaID, area.UniqueRoutes, area.TotalTicks)
			if area.HasSubareas {
				log.Printf("     Has %d sub-areas", area.SubareaCount)
			}
		}
	}
}
