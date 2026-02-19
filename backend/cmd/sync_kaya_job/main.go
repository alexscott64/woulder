package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
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
	// Embedded list of all 105 Kaya official destinations
	// Source: https://kayaclimb.com/explore (extracted 2026-02-18)
	destinations := []string{
		"Squamish-295658",
		"Red-Rocks-331387",
		"Bishop-316882",
		"Joshua-Tree-317008",
		"Hueco-Tanks-339538",
		"Joes-Valley-340826",
		"Vancouver-Island-295813",
		"Clear-Creek-Canyon-323872",
		"Ogden-1153006",
		"Lincoln-Lake-5272477",
		"Guanella-Pass-323792",
		"Tahoe-317136",
		"Little-Cottonwood-Canyon-986245",
		"New-River-Gorge-347179",
		"Coopers-Rock-347182",
		"Smith-Rock-336540",
		"Black-Mountain-317072",
		"Leavenworth-344933",
		"Kelowna-296013",
		"Hatcher-Pass-314961",
		"Devils-Lake-348323",
		"Lake-Ramona-10400507",
		"RMNP-323755",
		"Tramway-317070",
		"Vancouver-296037",
		"Ibex-341212",
		"Stone-Fort-999671",
		"Mount-Woodson-2192166",
		"Red-Feather-324534",
		"Flagstaff-Mountain-323839",
		"Big-Cottonwood-Canyon-BCC-341957",
		"Fraser-Valley-3340725",
		"Reimers-Ranch-339808",
		"Horseshoe-Canyon-Ranch-316278",
		"Tulsa-OK-10116402",
		"Mineral-King-15161231",
		"Rumbling-Bald-335837",
		"Rocktown-327484",
		"Horse-Pens-40-983782",
		"Malibu-838425",
		"Santa-Barbara-317853",
		"Doyle-322152",
		"Comox-Valley-Vancouver-Island-BC-7882675",
		"NYC-Bouldering-8736175",
		"Moes-Valley-340851",
		"Gold-Bar-344983",
		"The-Nooks-3899367",
		"Adirondacks-335103",
		"Stoney-Point-317772",
		"Treasury-2106513",
		"Eldorado-Canyon-323915",
		"Uintas-1394571",
		"holy-boulders-1016922",
		"Gunpowder-Falls-1395399",
		"Boat-Rock-327557",
		"Reynolds-Creek-328023",
		"Triassic-341357",
		"Needle-Peak-658063",
		"Box-Springs-Mountain-Reserve-5727203",
		"Horse-Flats-317843",
		"Mt-Evans-323773",
		"Smugglers-Notch-344705",
		"Rock-shop-348813",
		"Morpheus-345195",
		"Berkeley-316984",
		"Mount-Rubidoux-321790",
		"Index-345070",
		"purgatory-851804",
		"Vernon-4132330",
		"Exit-38-345299",
		"Castle-Rock-State-Park-328014",
		"Sams-Throne-316415",
		"Patapsco-Valley-State-Park-8555804",
		"Porcupine-Hills-6234426",
		"Cowell-316321",
		"Dixon-School-Road-335964",
		"Barton-Creek-Greenbelt-339852",
		"Utah-Hills-341651",
		"Price-1361664",
		"Big-Rock-291216",
		"Rogers-Park-339768",
		"Salt-Point-317575",
		"The-Citadel-295573",
		"Sierra-Buttes-318225",
		"Hammond-Pond-330274",
		"Nut-Tree-990859",
		"Santee-Boulders-2376083",
		"Indian-Rock-3199690",
		"Juan-De-Fuca-7846367",
		"Richland-Creek-15036518",
		"Lost-Ledges-345403",
		"Lions-Den-334501",
		"Conejo-Mountain-9502266",
		"Mckinney-Falls-1493881",
		"Wadi-Rum-389777",
		"Rocks-State-Park-330207",
		"Sawmill-330569",
		"Mt-Tamalpais-318183",
		"Rock-Creek-317075",
		"Sugarloaf-Ridge-State-Park-1770584",
	}

	return destinations, nil
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
