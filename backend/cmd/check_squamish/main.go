package main

import (
	"context"
	"fmt"
	"log"

	"github.com/alexscott64/woulder/backend/internal/database"
	"github.com/joho/godotenv"
)

func main() {
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

	ctx := context.Background()
	locationID := 6 // Squamish

	fmt.Println("\n========================================")
	fmt.Println("Squamish Mountain Project Areas")
	fmt.Println("========================================\n")

	// Get areas ordered by activity
	areas, err := db.GetAreasOrderedByActivity(ctx, locationID)
	if err != nil {
		log.Fatalf("Failed to get areas: %v", err)
	}

	fmt.Printf("Found %d root areas:\n\n", len(areas))

	for i, area := range areas {
		fmt.Printf("%d. %s (ID: %s)\n", i+1, area.Name, area.MPAreaID)
		fmt.Printf("   • Total ticks: %d\n", area.TotalTicks)
		fmt.Printf("   • Unique routes: %d\n", area.UniqueRoutes)
		fmt.Printf("   • Has subareas: %v", area.HasSubareas)
		if area.HasSubareas {
			fmt.Printf(" (%d sub-areas)", area.SubareaCount)
		}
		fmt.Println()
		fmt.Printf("   • Last climbed: %d days ago\n", area.DaysSinceClimb)

		// If it has subareas, fetch them
		if area.HasSubareas {
			subareas, err := db.GetSubareasOrderedByActivity(ctx, area.MPAreaID, locationID)
			if err != nil {
				log.Printf("   Error fetching subareas: %v", err)
			} else {
				fmt.Printf("   • Sub-areas (%d):\n", len(subareas))
				for j, sub := range subareas {
					fmt.Printf("     %d. %s (ID: %s) - %d routes, %d ticks\n",
						j+1, sub.Name, sub.MPAreaID, sub.UniqueRoutes, sub.TotalTicks)
				}
			}
		}
		fmt.Println()
	}

	fmt.Println("========================================")
}
