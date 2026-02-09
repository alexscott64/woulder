package main

import (
	"context"
	"fmt"
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
	MPAreaIDs    []int64
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
			MPAreaIDs:    []int64{120714486},
		},
		{
			LocationName: "Index",
			LocationID:   2,
			MPAreaIDs:    []int64{108123669},
		},
		{
			LocationName: "Gold Bar",
			LocationID:   3,
			MPAreaIDs:    []int64{105805788},
		},
		{
			LocationName: "Bellingham",
			LocationID:   4,
			MPAreaIDs:    []int64{107627792,125093900,108045031,118561215},
		},
		{
			LocationName: "Icicle Creek (Leavenworth)",
			LocationID:   5,
			MPAreaIDs:    []int64{105790237,105794001,105790727},
		},
		{
			LocationName: "Squamish",
			LocationID:   6,
			// Stawamus Chief boulder areas (from within 105805895):
			//   - Grand Wall Boulders (112842712)
			//   - North Wall Boulders (108506197)
			//   - Apron Boulders (106025685)
			// Paradise Valley Boulders (110937821) - contains sub-areas
			// Powerline Boulders (121199811)
			MPAreaIDs:    []int64{112842712,108506197,106025685,110937821,121199811},
		},
		{
			LocationName: "Skykomish - Paradise",
			LocationID:   7,
			MPAreaIDs:    []int64{120379690},
		},
		{
			LocationName: "Treasury",
			LocationID:   8,
			MPAreaIDs:    []int64{119589316},
		},
		{
			LocationName: "Calendar Butte",
			LocationID:   9,
			MPAreaIDs:    []int64{127029858},
		},
		{
			LocationName: "Joshua Tree",
			LocationID:   10,
			MPAreaIDs:    []int64{106098051},
		},
		{
			LocationName: "Black Mountain",
			LocationID:   11,
			MPAreaIDs:    []int64{105991127},
		},
		{
			LocationName: "Buttermilks",
			LocationID:   12,
			MPAreaIDs:    []int64{106132808},
		},
		{
			LocationName: "Happy / Sad Boulders",
			LocationID:   13,
			MPAreaIDs:    []int64{105799640,106068462},
		},
		{
			LocationName: "Yosemite",
			LocationID:   14,
			MPAreaIDs:    []int64{107457415},
		},
		{
			LocationName: "Tramway",
			LocationID:   15,
			MPAreaIDs:    []int64{105991060},
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
			log.Printf("\nSyncing area ID: %d for %s...", areaID, config.LocationName)

			// Convert int64 to string for API call
			areaIDStr := fmt.Sprintf("%d", areaID)
			err := climbService.SyncAreaRecursive(ctx, areaIDStr, &config.LocationID)
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
