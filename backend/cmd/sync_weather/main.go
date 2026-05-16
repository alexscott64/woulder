// Command sync_weather is a standalone CLI for manually refreshing weather
// data in the woulder.weather_data table. It is the recommended dev workflow
// when paired with WEATHER_OFFLINE_MODE=true on the API server: the server
// stops calling Open-Meteo on every request, and you call this tool on demand
// to bring the DB up to date.
//
// It bypasses the offlineMode flag entirely — this command IS the manual
// refresh path. It always calls Open-Meteo directly via the existing
// weather/client.OpenMeteoClient, then persists rows via the standard
// weather.PostgresRepository (same code path the server uses).
//
// See README.md in this directory for usage.
package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	weatherRepo "github.com/alexscott64/woulder/backend/internal/database/weather"
	"github.com/alexscott64/woulder/backend/internal/weather/client"
)

// minForecastHoursForCacheReplacement is the minimum number of FUTURE hourly
// forecast rows that an Open-Meteo response must contain before we destroy and
// replace the existing cache for a location.
//
// IMPORTANT: this constant is duplicated in
// `internal/service/weather_service.go` (kept in sync manually). If you change
// one, change the other. See that file for full background on the truncated-
// forecast bug this guard prevents.
const minForecastHoursForCacheReplacement = 14 * 24 // 336 hours = 14 days

// LocationInfo holds the lat/lon of a single location.
type LocationInfo struct {
	ID        int
	Name      string
	Latitude  float64
	Longitude float64
}

func main() {
	all := flag.Bool("all", false, "Sync every location in woulder.locations")
	locationID := flag.Int("location-id", 0, "Sync a single location by ID")
	rateLimitMs := flag.Int("rate-limit-ms", 1100, "Sleep between Open-Meteo calls (free tier ~600 req/min)")
	dryRun := flag.Bool("dry-run", false, "Log what would be fetched/written without touching Open-Meteo or the DB")
	flag.Parse()

	// Validate exactly one of --all / --location-id.
	if *all && *locationID != 0 {
		log.Fatal("Error: pass exactly one of --all or --location-id, not both")
	}
	if !*all && *locationID == 0 {
		log.Fatal("Error: must pass either --all or --location-id <ID>")
	}

	// Load .env (try cwd then parent), mirroring sibling commands.
	if err := godotenv.Load(); err != nil {
		if err2 := godotenv.Load("../.env"); err2 != nil {
			log.Printf("Warning: .env file not found: %v", err)
		}
	}

	log.Println("=== Weather Sync Tool ===")
	if *dryRun {
		log.Println("DRY RUN MODE: no API calls and no DB writes will be performed")
	}
	if *all {
		log.Println("Mode: ALL locations")
	} else {
		log.Printf("Mode: single location_id=%d", *locationID)
	}
	log.Printf("Rate limit: %dms between Open-Meteo calls", *rateLimitMs)
	log.Println()

	// DB connection (same env-driven pattern as cmd/backfill_radiation_dewpoint).
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		getEnvOrDefault("DB_PORT", "5432"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		getEnvOrDefault("DB_SSLMODE", "require"),
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("✓ Database connected")

	// SIGINT handling — finish the in-flight location, then exit cleanly.
	var stopRequested atomic.Bool
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("⚠️  Interrupt received — finishing current location then exiting...")
		stopRequested.Store(true)
	}()

	// Load locations to sync.
	locations, err := loadLocations(db, *all, *locationID)
	if err != nil {
		log.Fatalf("Failed to load locations: %v", err)
	}
	if len(locations) == 0 {
		log.Fatal("No matching locations found in woulder.locations")
	}
	log.Printf("✓ Loaded %d location(s) to sync", len(locations))
	log.Println()

	// Build the Open-Meteo client + repository wrapper.
	openMeteo := client.NewOpenMeteoClient()
	repo := weatherRepo.NewPostgresRepository(db)

	totalRowsSaved := 0
	totalCallsMade := 0
	totalFailures := 0
	startedAt := time.Now()
	ctx := context.Background()

	for i, loc := range locations {
		if stopRequested.Load() {
			log.Println("Stopping due to interrupt.")
			break
		}

		prefix := fmt.Sprintf("[%d/%d] location=%d (%s)", i+1, len(locations), loc.ID, loc.Name)

		if *dryRun {
			log.Printf("%s WOULD fetch lat=%.5f lon=%.5f and persist current+forecast rows",
				prefix, loc.Latitude, loc.Longitude)
		} else {
			rowsSaved, err := syncLocation(ctx, repo, openMeteo, loc)
			totalCallsMade++
			if err != nil {
				log.Printf("%s ✗ failed: %v", prefix, err)
				totalFailures++
			} else {
				log.Printf("%s rows_saved=%d", prefix, rowsSaved)
				totalRowsSaved += rowsSaved
			}
		}

		// Rate limit before next call (skip on last iteration / if interrupted).
		if i < len(locations)-1 && !stopRequested.Load() {
			time.Sleep(time.Duration(*rateLimitMs) * time.Millisecond)
		}
	}

	elapsed := time.Since(startedAt)
	log.Println()
	log.Println("=== Sync Complete ===")
	if *dryRun {
		log.Printf("Locations that would be synced: %d", len(locations))
	} else {
		log.Printf("Open-Meteo calls made: %d", totalCallsMade)
		log.Printf("Rows saved: %d", totalRowsSaved)
		log.Printf("Failures: %d", totalFailures)
	}
	log.Printf("Elapsed: %s", elapsed.Round(time.Second))

	if totalFailures > 0 {
		os.Exit(1)
	}
}

// syncLocation fetches current + forecast from Open-Meteo for a single location
// and persists every row via the standard weather repository. Returns the count
// of rows saved (current + each hourly forecast).
func syncLocation(ctx context.Context, repo *weatherRepo.PostgresRepository, om *client.OpenMeteoClient, loc LocationInfo) (int, error) {
	current, forecast, _, err := om.GetCurrentAndForecast(loc.Latitude, loc.Longitude)
	if err != nil {
		return 0, fmt.Errorf("open-meteo: %w", err)
	}
	if current == nil {
		return 0, errors.New("open-meteo returned nil current weather")
	}

	// FIX: Validate response length BEFORE replacing the cache.
	// Open-Meteo intermittently returns short hourly arrays (HTTP 200, no
	// error) which would otherwise poison the cache. Count future-only rows
	// (the slice may include past spin-up rows from past_hours=12 in
	// GetCurrentAndForecast) and only replace the cache if long enough.
	now := time.Now().UTC()
	futureHours := 0
	for i := range forecast {
		if forecast[i].Timestamp.After(now) {
			futureHours++
		}
	}

	rowsSaved := 0
	current.LocationID = loc.ID

	if futureHours < minForecastHoursForCacheReplacement {
		log.Printf(
			"    WARN: Open-Meteo returned truncated forecast for location %d (lat=%.5f lon=%.5f): future_hours=%d, threshold=%d. "+
				"Skipping cache replacement to preserve previously-cached forecast; saving current observation only.",
			loc.ID, loc.Latitude, loc.Longitude,
			futureHours, minForecastHoursForCacheReplacement,
		)
		// Save just the current observation row — it's safe and useful.
		if err := repo.Save(ctx, current); err != nil {
			return rowsSaved, fmt.Errorf("save current (truncated forecast path): %w", err)
		}
		rowsSaved++
		return rowsSaved, nil
	}

	// Atomically replace the future-forecast cache (delete + save in a single
	// transaction) — matches the server's getLocationWeatherWithOptions path.
	if err := repo.ReplaceFutureForLocation(ctx, loc.ID, forecast); err != nil {
		return rowsSaved, fmt.Errorf("replace future forecast: %w", err)
	}
	rowsSaved += len(forecast)

	if err := repo.Save(ctx, current); err != nil {
		return rowsSaved, fmt.Errorf("save current: %w", err)
	}
	rowsSaved++

	return rowsSaved, nil
}

// loadLocations returns the locations to sync. If all is true, returns every
// row in woulder.locations; otherwise returns just the row matching onlyID.
func loadLocations(db *sql.DB, all bool, onlyID int) ([]LocationInfo, error) {
	query := `SELECT id, name, latitude, longitude FROM woulder.locations`
	args := []any{}
	if !all {
		query += ` WHERE id = $1`
		args = append(args, onlyID)
	}
	query += ` ORDER BY id`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query locations: %w", err)
	}
	defer rows.Close()

	var out []LocationInfo
	for rows.Next() {
		var li LocationInfo
		if err := rows.Scan(&li.ID, &li.Name, &li.Latitude, &li.Longitude); err != nil {
			return nil, fmt.Errorf("scan location: %w", err)
		}
		out = append(out, li)
	}
	return out, rows.Err()
}

func getEnvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
