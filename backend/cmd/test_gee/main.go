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

	log.Println("=== Google Earth Engine API Test ===")
	log.Println()

	// Create tree cover client
	client := boulder_drying.NewTreeCoverClient()
	if !client.IsEnabled() {
		log.Fatal("ERROR: Google Earth Engine client not initialized. Check your .env credentials.")
	}

	log.Println("✓ Google Earth Engine client initialized successfully")
	log.Println()

	// Test coordinates
	testCases := []struct {
		name string
		lat  float64
		lon  float64
	}{
		{"Leavenworth, WA (Icicle Creek)", 47.6, -120.9},
		{"Bishop, CA (Buttermilks)", 37.35, -118.7},
		{"Squamish, BC (Malamute)", 49.7, -123.2},
		{"Red Rocks, NV", 36.15, -115.45},
	}

	for _, tc := range testCases {
		log.Printf("Testing: %s (%.6f, %.6f)", tc.name, tc.lat, tc.lon)

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		coverage, err := client.GetTreeCoverage(ctx, tc.lat, tc.lon)
		cancel()

		if err != nil {
			log.Printf("  ✗ ERROR: %v", err)
		} else {
			log.Printf("  ✓ Tree coverage: %.1f%%", coverage)
		}
		log.Println()
	}

	log.Println("=== Test Complete ===")
}
