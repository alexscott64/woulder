package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/alexscott64/woulder/backend/internal/database"
	"github.com/alexscott64/woulder/backend/internal/mountainproject"
	"github.com/alexscott64/woulder/backend/internal/service"
	"github.com/joho/godotenv"
)

// AreaConfig defines the mapping between locations and their Mountain Project area IDs
type AreaConfig struct {
	LocationName string
	LocationID   int
	MPAreaIDs    []string
}

func main() {
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

	// Initialize climb tracking service
	climbService := service.NewClimbTrackingService(db, mpClient)

	// Define area mappings (location ID -> Mountain Project area IDs)
	// IMPORTANT: These IDs must match the actual database location IDs
	areaConfigs := []AreaConfig{
		{
			LocationName: "Skykomish - Money Creek",
			LocationID:   1,
			MPAreaIDs:    []string{"120714486"},
		},
		{
			LocationName: "Index",
			LocationID:   2,
			MPAreaIDs:    []string{"108123669"},
		},
		{
			LocationName: "Gold Bar",
			LocationID:   3,
			MPAreaIDs:    []string{"105805788"},
		},
		{
			LocationName: "Bellingham",
			LocationID:   4,
			MPAreaIDs:    []string{"107627792", "125093900", "108045031", "118561215"},
		},
		{
			LocationName: "Icicle Creek (Leavenworth)",
			LocationID:   5,
			MPAreaIDs:    []string{"105790237", "105794001", "105790727"},
		},
		{
			LocationName: "Squamish",
			LocationID:   6,
			MPAreaIDs:    []string{"110937821", "105808584", "105805895", "121199811"},
		},
		{
			LocationName: "Skykomish - Paradise",
			LocationID:   7,
			MPAreaIDs:    []string{"120379690"},
		},
		{
			LocationName: "Treasury",
			LocationID:   8,
			MPAreaIDs:    []string{"119589316"},
		},
		{
			LocationName: "Calendar Butte",
			LocationID:   9,
			MPAreaIDs:    []string{"127029858"},
		},
		{
			LocationName: "Joshua Tree",
			LocationID:   10,
			MPAreaIDs:    []string{"106098051"},
		},
		{
			LocationName: "Black Mountain",
			LocationID:   11,
			MPAreaIDs:    []string{"105991127"},
		},
		{
			LocationName: "Buttermilks",
			LocationID:   12,
			MPAreaIDs:    []string{"106132808"},
		},
		{
			LocationName: "Happy / Sad Boulders",
			LocationID:   13,
			MPAreaIDs:    []string{"105799640", "106068462"},
		},
		{
			LocationName: "Yosemite",
			LocationID:   14,
			MPAreaIDs:    []string{"107457415"},
		},
		{
			LocationName: "Tramway",
			LocationID:   15,
			MPAreaIDs:    []string{"105991060"},
		},
	}

	ctx := context.Background()
	totalAreas := 0
	successCount := 0
	failCount := 0

	startTime := time.Now()

	// Process each location's areas
	for _, config := range areaConfigs {
		log.Printf("\n========================================")
		log.Printf("Processing location: %s (ID: %d)", config.LocationName, config.LocationID)
		log.Printf("========================================")

		for _, areaID := range config.MPAreaIDs {
			totalAreas++
			log.Printf("\nSyncing area ID: %s for %s...", areaID, config.LocationName)

			err := climbService.SyncAreaRecursive(ctx, areaID, &config.LocationID)
			if err != nil {
				log.Printf("ERROR syncing area %s: %v", areaID, err)
				failCount++
				continue
			}

			successCount++
			log.Printf("✓ Successfully synced area %s", areaID)
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
