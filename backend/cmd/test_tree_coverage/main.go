package main

import (
	"context"
	"log"
	"time"

	"github.com/alexscott64/woulder/backend/internal/weather/boulder_drying"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	log.Println("=== Tree Coverage Test ===")
	log.Println()

	// Initialize tree cover client
	client := boulder_drying.NewTreeCoverClient()

	// Test coordinates (various climbing areas)
	testLocations := []struct {
		name string
		lat  float64
		lon  float64
	}{
		{"Index/Zelda Boulders, WA", 47.8213, -121.5601},
		{"Leavenworth, WA (Icicle Creek)", 47.6, -120.9},
		{"Leavenworth, WA (Peshastin Pinnacles)", 47.6, -120.65},
		{"Bishop, CA (Buttermilks)", 37.35, -118.7},
	}

	for _, loc := range testLocations {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		coverage, err := client.GetTreeCoverage(ctx, loc.lat, loc.lon)
		cancel()

		if err != nil {
			log.Printf("✗ %s: ERROR - %v", loc.name, err)
		} else {
			log.Printf("✓ %s: %.1f%% tree coverage", loc.name, coverage)
		}
	}

	log.Println()
	log.Println("=== Test Complete ===")

	if client.IsEnabled() {
		log.Println("✓ Google Earth Engine is working!")
	} else {
		log.Println("⚠ Using location-based estimates (Earth Engine not configured)")
	}
}
