package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/alexscott64/woulder/backend/internal/database"
	kayaClient "github.com/alexscott64/woulder/backend/internal/kaya"
	"github.com/alexscott64/woulder/backend/internal/service"
	"github.com/joho/godotenv"
)

// LocationConfig defines locations to sync from Kaya
type LocationConfig struct {
	Name      string
	Slug      string
	Recursive bool // Whether to sync sub-locations
}

func main() {
	log.Println("Starting Kaya climb data sync...")

	// Command-line flags
	slugFlag := flag.String("slug", "", "Specific location slug to sync (e.g., 'Leavenworth-344933')")
	recursiveFlag := flag.Bool("recursive", true, "Sync sub-locations recursively")
	testFlag := flag.Bool("test", false, "Test mode: only sync Leavenworth")
	tokenFlag := flag.String("token", "", "Kaya API JWT token (or set KAYA_AUTH_TOKEN env var)")
	flag.Parse()

	// Load environment variables from backend directory
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: .env file not found, using system environment variables")
	}

	// Get auth token from flag or environment variable
	authToken := *tokenFlag
	if authToken == "" {
		authToken = os.Getenv("KAYA_AUTH_TOKEN")
	}

	if authToken == "" {
		log.Println("WARNING: No auth token provided. Set KAYA_AUTH_TOKEN env var or use -token flag.")
		log.Println("Without authentication, API calls may fail or return limited data.")
	}

	// Initialize database
	db, err := database.New()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Kaya HTTP client (Cloudflare doesn't block normal HTTP requests!)
	log.Println("Initializing Kaya API client...")
	client := kayaClient.NewClient()
	log.Println("✓ Client initialized successfully")

	// Initialize Kaya sync service (no job monitor for manual sync)
	kayaService := service.NewKayaSyncService(db.Kaya(), client, nil)

	ctx := context.Background()

	// Handle specific slug sync
	if *slugFlag != "" {
		log.Printf("Syncing specific location: %s (recursive: %v)", *slugFlag, *recursiveFlag)
		if err := syncLocation(ctx, kayaService, *slugFlag, *recursiveFlag); err != nil {
			log.Fatalf("Failed to sync location %s: %v", *slugFlag, err)
		}
		log.Println("✓ Sync completed successfully!")
		return
	}

	// Define location mappings
	locationConfigs := []LocationConfig{
		{
			Name:      "Leavenworth",
			Slug:      "Leavenworth-344933",
			Recursive: true,
		},
	}

	// Test mode: only sync Leavenworth
	if *testFlag {
		log.Println("TEST MODE: Only syncing Leavenworth")
		locationConfigs = locationConfigs[:1]
	}

	totalLocations := len(locationConfigs)
	successCount := 0
	failCount := 0
	startTime := time.Now()

	// Process each location
	for i, config := range locationConfigs {
		log.Printf("\n========================================")
		log.Printf("Processing location %d/%d: %s", i+1, totalLocations, config.Name)
		log.Printf("Slug: %s (recursive: %v)", config.Slug, config.Recursive)
		log.Printf("========================================")

		if err := syncLocation(ctx, kayaService, config.Slug, config.Recursive); err != nil {
			log.Printf("ERROR syncing %s: %v", config.Name, err)
			failCount++
			continue
		}

		successCount++
		log.Printf("✓ Successfully synced %s", config.Name)
	}

	elapsed := time.Since(startTime)

	log.Printf("\n========================================")
	log.Printf("Sync Complete!")
	log.Printf("========================================")
	log.Printf("Total locations processed: %d", totalLocations)
	log.Printf("Successful: %d", successCount)
	log.Printf("Failed: %d", failCount)
	log.Printf("Time elapsed: %s", elapsed.Round(time.Second))
	log.Printf("========================================")

	if failCount > 0 {
		os.Exit(1)
	}
}

// syncLocation syncs a single location with error handling
func syncLocation(ctx context.Context, service *service.KayaSyncService, slug string, recursive bool) error {
	log.Printf("Starting sync for slug: %s", slug)

	err := service.SyncLocationBySlug(ctx, slug, recursive)
	if err != nil {
		// Check if it's a transient error
		if isTransientError(err) {
			log.Printf("Transient error detected, retrying in 5 seconds...")
			time.Sleep(5 * time.Second)
			err = service.SyncLocationBySlug(ctx, slug, recursive)
		}
	}

	return err
}

// isTransientError checks if an error is likely transient and worth retrying
func isTransientError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	transientKeywords := []string{
		"timeout",
		"connection refused",
		"temporary failure",
		"502",
		"503",
		"504",
	}

	for _, keyword := range transientKeywords {
		if strings.Contains(errMsg, keyword) {
			return true
		}
	}
	return false
}
