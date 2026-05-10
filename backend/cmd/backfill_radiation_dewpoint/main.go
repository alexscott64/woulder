// Command backfill_radiation_dewpoint fills in historical radiation and
// dewpoint data for rows in woulder.weather_data that were inserted before
// migration 000032_add_radiation_dewpoint added the corresponding columns.
//
// Existing historical rows have shortwave_radiation, direct_radiation,
// diffuse_radiation, and dewpoint_f all set to 0 (the migration default).
// This tool finds those rows, looks up each location's lat/lon, calls
// Open-Meteo's free historical archive API, and UPDATEs the rows with the
// correct values.
//
// See README.md in this directory for usage.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const (
	openMeteoArchiveURL = "https://archive-api.open-meteo.com/v1/archive"
)

// LocationInfo holds the lat/lon of a single location.
type LocationInfo struct {
	ID        int
	Latitude  float64
	Longitude float64
}

// MissingRange represents the contiguous date range of missing data for a
// single location.
type MissingRange struct {
	LocationID int
	StartDate  time.Time
	EndDate    time.Time
	NumDates   int
}

// archiveResponse mirrors the Open-Meteo archive endpoint hourly response.
type archiveResponse struct {
	Hourly struct {
		Time               []string  `json:"time"`
		ShortwaveRadiation []float64 `json:"shortwave_radiation"`
		DirectRadiation    []float64 `json:"direct_radiation"`
		DiffuseRadiation   []float64 `json:"diffuse_radiation"`
		Dewpoint2m         []float64 `json:"dew_point_2m"`
	} `json:"hourly"`
	Error  bool   `json:"error"`
	Reason string `json:"reason"`
}

func main() {
	batchSize := flag.Int("batch-size", 100, "Number of (location_id, date) groups processed per batch (informational; one API call per location range)")
	startDateStr := flag.String("start-date", "", "Earliest date to backfill (YYYY-MM-DD). Default: oldest row in DB")
	endDateStr := flag.String("end-date", "", "Latest date to backfill (YYYY-MM-DD). Default: today")
	locationID := flag.Int("location-id", 0, "Restrict to a single location_id (0 = all locations)")
	dryRun := flag.Bool("dry-run", false, "Log planned UPDATE counts without writing")
	rateLimitMs := flag.Int("rate-limit-ms", 1100, "Sleep between Open-Meteo calls (free tier ~600 req/min)")
	flag.Parse()

	// Load .env (try cwd then parent, mirroring sync_tree_cover/sync_kaya).
	if err := godotenv.Load(); err != nil {
		if err2 := godotenv.Load("../.env"); err2 != nil {
			log.Printf("Warning: .env file not found: %v", err)
		}
	}

	log.Println("=== Radiation & Dewpoint Backfill Tool ===")
	if *dryRun {
		log.Println("DRY RUN MODE: no rows will be modified")
	}
	if *locationID != 0 {
		log.Printf("Restricting to location_id=%d", *locationID)
	}
	log.Printf("Rate limit: %dms between Open-Meteo calls", *rateLimitMs)
	log.Printf("Batch size: %d (informational)", *batchSize)
	log.Println()

	// DB connection (same env-driven pattern as sync_tree_cover/main.go).
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

	// Parse optional date bounds.
	var startDate, endDate time.Time
	if *startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", *startDateStr)
		if err != nil {
			log.Fatalf("Invalid --start-date %q: %v", *startDateStr, err)
		}
	}
	if *endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", *endDateStr)
		if err != nil {
			log.Fatalf("Invalid --end-date %q: %v", *endDateStr, err)
		}
	} else {
		endDate = time.Now().UTC().Truncate(24 * time.Hour)
	}

	// SIGINT handling — finish current batch then exit.
	var stopRequested atomic.Bool
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("⚠️  Interrupt received — finishing current batch then exiting...")
		stopRequested.Store(true)
	}()

	// Load locations.
	log.Println("Fetching location coordinates...")
	locations, err := loadLocations(db, *locationID)
	if err != nil {
		log.Fatalf("Failed to load locations: %v", err)
	}
	log.Printf("✓ Loaded %d location(s)", len(locations))

	// Find missing ranges per location.
	log.Println("Scanning for weather_data rows missing radiation/dewpoint data...")
	ranges, err := findMissingRanges(db, *locationID, startDate, endDate)
	if err != nil {
		log.Fatalf("Failed to find missing ranges: %v", err)
	}

	if len(ranges) == 0 {
		log.Println("✓ Nothing to backfill — all rows already have data.")
		return
	}

	log.Printf("Found missing data for %d location(s)", len(ranges))
	totalDates := 0
	for _, r := range ranges {
		totalDates += r.NumDates
	}
	log.Printf("Total distinct (location, date) pairs missing: %d", totalDates)
	log.Println()

	// Process each location's range.
	httpClient := &http.Client{Timeout: 60 * time.Second}
	totalRowsUpdated := 0
	totalCallsMade := 0
	startedAt := time.Now()

	for i, r := range ranges {
		if stopRequested.Load() {
			log.Println("Stopping due to interrupt.")
			break
		}

		loc, ok := locations[r.LocationID]
		if !ok {
			log.Printf("[%d/%d] ✗ location_id=%d has no entry in woulder.locations — skipping",
				i+1, len(ranges), r.LocationID)
			continue
		}

		log.Printf("[%d/%d] location=%d (%.5f, %.5f) dates=%s..%s (%d distinct dates)",
			i+1, len(ranges), r.LocationID, loc.Latitude, loc.Longitude,
			r.StartDate.Format("2006-01-02"), r.EndDate.Format("2006-01-02"), r.NumDates)

		// Fetch from Open-Meteo archive.
		resp, err := fetchArchive(httpClient, loc.Latitude, loc.Longitude, r.StartDate, r.EndDate)
		totalCallsMade++
		if err != nil {
			log.Printf("    ✗ Open-Meteo fetch failed: %v", err)
			time.Sleep(time.Duration(*rateLimitMs) * time.Millisecond)
			continue
		}
		if len(resp.Hourly.Time) == 0 {
			log.Printf("    ⊙ No hourly data returned")
			time.Sleep(time.Duration(*rateLimitMs) * time.Millisecond)
			continue
		}
		log.Printf("    ✓ Fetched %d hourly samples", len(resp.Hourly.Time))

		rowsUpdated, err := applyUpdates(context.Background(), db, r.LocationID, resp, *dryRun)
		if err != nil {
			log.Printf("    ✗ Update failed: %v", err)
		} else {
			verb := "rows_updated"
			if *dryRun {
				verb = "rows_would_update"
			}
			log.Printf("    ✓ %s=%d", verb, rowsUpdated)
			totalRowsUpdated += rowsUpdated
		}

		// Rate limit before next call.
		if i < len(ranges)-1 && !stopRequested.Load() {
			time.Sleep(time.Duration(*rateLimitMs) * time.Millisecond)
		}
	}

	elapsed := time.Since(startedAt)
	log.Println()
	log.Println("=== Backfill Complete ===")
	log.Printf("Open-Meteo calls made: %d", totalCallsMade)
	if *dryRun {
		log.Printf("Rows that would have been updated: %d", totalRowsUpdated)
	} else {
		log.Printf("Rows updated: %d", totalRowsUpdated)
	}
	log.Printf("Elapsed: %s", elapsed.Round(time.Second))
}

// loadLocations loads lat/lon for either all locations or a single one.
func loadLocations(db *sql.DB, onlyID int) (map[int]LocationInfo, error) {
	query := `SELECT id, latitude, longitude FROM woulder.locations`
	args := []any{}
	if onlyID != 0 {
		query += ` WHERE id = $1`
		args = append(args, onlyID)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query locations: %w", err)
	}
	defer rows.Close()

	out := make(map[int]LocationInfo)
	for rows.Next() {
		var li LocationInfo
		if err := rows.Scan(&li.ID, &li.Latitude, &li.Longitude); err != nil {
			return nil, fmt.Errorf("scan location: %w", err)
		}
		out[li.ID] = li
	}
	return out, rows.Err()
}

// findMissingRanges queries weather_data for rows where ALL of the four
// columns are NULL or 0, then collapses each location's missing dates into a
// single contiguous (min, max) range. We use one range per location because
// Open-Meteo's archive API can return arbitrary date ranges in a single call,
// and over-fetching is harmless (the UPDATE WHERE clause filters out rows
// that already have data).
func findMissingRanges(db *sql.DB, onlyID int, startDate, endDate time.Time) ([]MissingRange, error) {
	// "Missing" = all four columns are NULL OR 0. Migration 000032 sets them to
	// NOT NULL DEFAULT 0, so historical rows will be 0 (not NULL), but we keep
	// the NULL check for safety.
	query := `
        SELECT location_id,
               (timestamp AT TIME ZONE 'UTC')::date AS day
        FROM woulder.weather_data
        WHERE (shortwave_radiation IS NULL OR shortwave_radiation = 0)
          AND (direct_radiation    IS NULL OR direct_radiation    = 0)
          AND (diffuse_radiation   IS NULL OR diffuse_radiation   = 0)
          AND (dewpoint_f          IS NULL OR dewpoint_f          = 0)
    `
	args := []any{}
	argIdx := 1

	if onlyID != 0 {
		query += fmt.Sprintf(" AND location_id = $%d", argIdx)
		args = append(args, onlyID)
		argIdx++
	}
	if !startDate.IsZero() {
		query += fmt.Sprintf(" AND timestamp >= $%d", argIdx)
		args = append(args, startDate)
		argIdx++
	}
	if !endDate.IsZero() {
		// endDate is inclusive day boundary — add a day to be exclusive.
		query += fmt.Sprintf(" AND timestamp < $%d", argIdx)
		args = append(args, endDate.Add(24*time.Hour))
		argIdx++
	}

	query += `
        GROUP BY location_id, day
        ORDER BY location_id, day
    `

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query missing rows: %w", err)
	}
	defer rows.Close()

	// Collect per location.
	perLoc := make(map[int][]time.Time)
	for rows.Next() {
		var locID int
		var day time.Time
		if err := rows.Scan(&locID, &day); err != nil {
			return nil, fmt.Errorf("scan missing row: %w", err)
		}
		perLoc[locID] = append(perLoc[locID], day)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Build one MissingRange per location: min..max date.
	out := make([]MissingRange, 0, len(perLoc))
	for locID, days := range perLoc {
		if len(days) == 0 {
			continue
		}
		sort.Slice(days, func(i, j int) bool { return days[i].Before(days[j]) })
		out = append(out, MissingRange{
			LocationID: locID,
			StartDate:  days[0],
			EndDate:    days[len(days)-1],
			NumDates:   len(days),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].LocationID < out[j].LocationID })
	return out, nil
}

// fetchArchive calls the Open-Meteo historical archive endpoint.
func fetchArchive(client *http.Client, lat, lon float64, start, end time.Time) (*archiveResponse, error) {
	url := fmt.Sprintf(
		"%s?latitude=%.6f&longitude=%.6f&start_date=%s&end_date=%s&hourly=shortwave_radiation,direct_radiation,diffuse_radiation,dew_point_2m&timezone=UTC",
		openMeteoArchiveURL,
		lat, lon,
		start.Format("2006-01-02"),
		end.Format("2006-01-02"),
	)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	var ar archiveResponse
	if err := json.Unmarshal(body, &ar); err != nil {
		return nil, fmt.Errorf("decode: %w (body: %s)", err, string(body))
	}
	if ar.Error {
		return nil, fmt.Errorf("api error: %s", ar.Reason)
	}
	return &ar, nil
}

// applyUpdates iterates the hourly samples and UPDATEs matching weather_data
// rows. Runs inside a transaction (per location) for atomicity.
//
// The Open-Meteo "time" field is a string like "2024-01-01T00:00" in UTC
// (because we requested timezone=UTC). We parse it as UTC and match on
// equality with weather_data.timestamp.
//
// Dewpoint is converted from Celsius to Fahrenheit. Radiation is W/m² as-is.
func applyUpdates(ctx context.Context, db *sql.DB, locationID int, resp *archiveResponse, dryRun bool) (int, error) {
	hourly := resp.Hourly
	n := len(hourly.Time)
	if n == 0 {
		return 0, nil
	}

	// Sanity: all hourly arrays should be the same length.
	if len(hourly.ShortwaveRadiation) != n ||
		len(hourly.DirectRadiation) != n ||
		len(hourly.DiffuseRadiation) != n ||
		len(hourly.Dewpoint2m) != n {
		return 0, fmt.Errorf("inconsistent hourly array lengths from open-meteo")
	}

	if dryRun {
		// Count rows that *would* be updated without modifying anything.
		// We check existence per timestamp to give a realistic estimate.
		count := 0
		for i := 0; i < n; i++ {
			ts, err := time.ParseInLocation("2006-01-02T15:04", hourly.Time[i], time.UTC)
			if err != nil {
				continue
			}
			var exists bool
			err = db.QueryRowContext(ctx, `
                SELECT EXISTS (
                    SELECT 1 FROM woulder.weather_data
                    WHERE location_id = $1 AND timestamp = $2
                      AND (shortwave_radiation IS NULL OR shortwave_radiation = 0)
                      AND (direct_radiation    IS NULL OR direct_radiation    = 0)
                      AND (diffuse_radiation   IS NULL OR diffuse_radiation   = 0)
                      AND (dewpoint_f          IS NULL OR dewpoint_f          = 0)
                )
            `, locationID, ts).Scan(&exists)
			if err == nil && exists {
				count++
			}
		}
		return count, nil
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, `
        UPDATE woulder.weather_data
        SET shortwave_radiation = $1,
            direct_radiation    = $2,
            diffuse_radiation   = $3,
            dewpoint_f          = $4
        WHERE location_id = $5
          AND timestamp = $6
    `)
	if err != nil {
		return 0, fmt.Errorf("prepare: %w", err)
	}
	defer stmt.Close()

	totalRows := 0
	for i := 0; i < n; i++ {
		ts, err := time.ParseInLocation("2006-01-02T15:04", hourly.Time[i], time.UTC)
		if err != nil {
			// Try alternate format with seconds, just in case.
			ts, err = time.ParseInLocation("2006-01-02T15:04:05", hourly.Time[i], time.UTC)
			if err != nil {
				continue
			}
		}

		dewpointF := celsiusToFahrenheit(hourly.Dewpoint2m[i])

		res, err := stmt.ExecContext(ctx,
			hourly.ShortwaveRadiation[i],
			hourly.DirectRadiation[i],
			hourly.DiffuseRadiation[i],
			dewpointF,
			locationID,
			ts,
		)
		if err != nil {
			return totalRows, fmt.Errorf("update at %s: %w", ts, err)
		}
		ra, _ := res.RowsAffected()
		totalRows += int(ra)
	}

	if err := tx.Commit(); err != nil {
		return totalRows, fmt.Errorf("commit: %w", err)
	}
	return totalRows, nil
}

func celsiusToFahrenheit(c float64) float64 {
	return c*9.0/5.0 + 32.0
}

func getEnvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
