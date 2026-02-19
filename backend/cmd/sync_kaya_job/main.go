package main

import (
	"bufio"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/alexscott64/woulder/backend/internal/database"
	kayaClient "github.com/alexscott64/woulder/backend/internal/kaya"
	"github.com/alexscott64/woulder/backend/internal/monitoring"
	"github.com/alexscott64/woulder/backend/internal/service"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// KayaSyncJob runs scheduled syncs of Kaya data with monitoring
func main() {
	log.Println("Starting Kaya scheduled sync job...")

	// Command-line flags
	incrementalFlag := flag.Bool("incremental", true, "Only sync new data since last sync")
	testFlag := flag.Bool("test", false, "Test mode: only sync 3 destinations")
	delayFlag := flag.Int("delay", 3, "Delay in seconds between destinations")
	flag.Parse()

	// Load environment variables - try current directory first, then parent
	if err := godotenv.Load(".env"); err != nil {
		if err := godotenv.Load("../.env"); err != nil {
			log.Printf("Warning: .env file not found in . or .., using system environment variables")
		}
	}

	// Initialize database
	db, err := database.New()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create a separate SQL connection for job monitoring
	monitorDB, err := createMonitoringDB()
	if err != nil {
		log.Fatalf("Failed to create monitoring database connection: %v", err)
	}
	defer monitorDB.Close()

	// Initialize job monitor
	jobMonitor := monitoring.NewJobMonitor(monitorDB)

	// Load destinations to determine total items
	destinations, err := loadDestinations()
	if err != nil {
		log.Fatalf("Failed to load destinations: %v", err)
	}

	// Test mode: only sync first 3
	if *testFlag {
		log.Println("TEST MODE: Only syncing first 3 destinations")
		if len(destinations) > 3 {
			destinations = destinations[:3]
		}
	}

	// Start job run
	jobName := "kaya_sync"
	jobType := "full"
	if *incrementalFlag {
		jobType = "incremental"
	}

	jobExec, err := jobMonitor.StartJob(context.Background(), jobName, jobType, len(destinations), map[string]interface{}{
		"incremental": *incrementalFlag,
		"test_mode":   *testFlag,
		"delay":       *delayFlag,
	})
	if err != nil {
		log.Fatalf("Failed to start job tracking: %v", err)
	}

	log.Printf("Job started with ID: %d", jobExec.ID)
	startTime := time.Now()

	// Run sync
	successCount, failCount := runSync(db, jobMonitor, jobExec.ID, destinations, *incrementalFlag, *delayFlag)

	// Complete job tracking
	duration := time.Since(startTime)

	if failCount > 0 && successCount == 0 {
		// Complete failure
		errMsg := "All destinations failed to sync"
		jobMonitor.FailJob(context.Background(), jobExec.ID, errMsg)
		log.Fatalf("Kaya sync job failed after %s: %s", duration, errMsg)
	}

	// Complete successfully (even with partial failures)
	jobMonitor.CompleteJob(context.Background(), jobExec.ID)
	log.Printf("✓ Kaya sync job completed in %s (success: %d, failed: %d)", duration, successCount, failCount)
}

func runSync(db *database.Database, jobMonitor *monitoring.JobMonitor, jobID int64, destinations []string, incremental bool, delay int) (int, int) {
	ctx := context.Background()

	// Initialize Kaya client
	client := kayaClient.NewClient()

	// Initialize Kaya sync service
	kayaService := service.NewKayaSyncService(db.Kaya(), client, nil)

	successCount := 0
	failCount := 0
	processed := 0

	for i, slug := range destinations {
		log.Printf("\n[%d/%d] Syncing %s...", i+1, len(destinations), slug)

		// For incremental sync, check if we need to sync this location
		if incremental {
			shouldSync, err := shouldSyncLocation(ctx, db, slug)
			if err != nil {
				log.Printf("Error checking sync status for %s: %v", slug, err)
			} else if !shouldSync {
				log.Printf("Skipping %s (recently synced)", slug)
				processed++
				jobMonitor.UpdateProgress(ctx, jobID, processed, successCount, failCount)
				continue
			}
		}

		// Sync location
		err := kayaService.SyncLocationBySlug(ctx, slug, true)
		processed++

		if err != nil {
			log.Printf("ERROR syncing %s: %v", slug, err)
			failCount++
		} else {
			successCount++
			log.Printf("✓ Synced %s", slug)
		}

		// Update progress
		jobMonitor.UpdateProgress(ctx, jobID, processed, successCount, failCount)

		// Rate limiting
		if i < len(destinations)-1 {
			time.Sleep(time.Duration(delay) * time.Second)
		}
	}

	log.Printf("\n========================================")
	log.Printf("Sync Summary:")
	log.Printf("Total: %d, Success: %d, Failed: %d", len(destinations), successCount, failCount)
	log.Printf("========================================")

	return successCount, failCount
}

func loadDestinations() ([]string, error) {
	// Load from kaya-destinations.txt
	file, err := os.Open("../docs/kaya-destinations.txt")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var slugs []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "Total:") {
			continue
		}
		if strings.Contains(line, "-") && !strings.HasPrefix(line, "Format:") {
			slugs = append(slugs, line)
		}
	}

	return slugs, scanner.Err()
}

func shouldSyncLocation(ctx context.Context, db *database.Database, slug string) (bool, error) {
	// Check last sync time from kaya_sync_progress
	// For now, always sync (incremental logic can be added later)
	// Full implementation would check: last_synced_at > NOW() - INTERVAL '24 hours'
	return true, nil
}

func createMonitoringDB() (*sql.DB, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	if port == "" {
		port = "5432"
	}
	if sslmode == "" {
		sslmode = "require"
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)

	return sql.Open("postgres", connStr)
}
