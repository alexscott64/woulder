// Command backfill_location_timezone populates / corrects the per-location
// IANA timezone column on woulder.locations.
//
// Migration 000037_add_location_timezone added a `timezone TEXT NOT NULL
// DEFAULT 'America/Los_Angeles'` column. Every existing row therefore has the
// Pacific default. This tool re-derives each row's timezone from its
// (latitude, longitude) using the offline tzf polygon dataset wrapped by
// internal/geo.LookupTimezone, and UPDATEs rows whose derived value differs
// from what's currently stored.
//
// By default, only rows whose current timezone is the migration default
// 'America/Los_Angeles' are eligible for an update. Pass -force to re-derive
// every row regardless of its current value.
//
// See README.md in this directory for usage.
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/alexscott64/woulder/backend/internal/geo"
)

const migrationDefaultTimezone = "America/Los_Angeles"

type locationRow struct {
	ID        int
	Name      string
	Latitude  float64
	Longitude float64
	Timezone  string
}

func main() {
	dryRun := flag.Bool("dry-run", false, "Log planned UPDATEs without writing")
	locationID := flag.Int("location-id", 0, "Restrict to a single location_id (0 = all locations)")
	force := flag.Bool("force", false, "Re-derive every row, including those whose current timezone is not the migration default")
	flag.Parse()

	// Load .env (try cwd then parent, mirroring backfill_radiation_dewpoint).
	if err := godotenv.Load(); err != nil {
		if err2 := godotenv.Load("../.env"); err2 != nil {
			log.Printf("Warning: .env file not found: %v", err)
		}
	}

	log.Println("=== Location Timezone Backfill Tool ===")
	if *dryRun {
		log.Println("DRY RUN MODE: no rows will be modified")
	}
	if *locationID != 0 {
		log.Printf("Restricting to location_id=%d", *locationID)
	}
	if *force {
		log.Println("FORCE MODE: re-deriving every row, not just rows at the migration default")
	} else {
		log.Printf("Default mode: only updating rows whose current timezone = %q", migrationDefaultTimezone)
	}
	log.Println()

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

	rows, err := loadLocations(db, *locationID)
	if err != nil {
		log.Fatalf("Failed to load locations: %v", err)
	}
	log.Printf("✓ Loaded %d location(s)", len(rows))
	log.Println()

	scanned := 0
	updated := 0
	skipped := 0
	startedAt := time.Now()

	for _, row := range rows {
		scanned++
		derived := geo.LookupTimezone(row.Latitude, row.Longitude)

		if derived == row.Timezone {
			log.Printf("  id=%d name=%q (%.5f,%.5f) current=%q derived=%q action=skip-equal",
				row.ID, row.Name, row.Latitude, row.Longitude, row.Timezone, derived)
			skipped++
			continue
		}

		eligible := *force || row.Timezone == migrationDefaultTimezone
		if !eligible {
			log.Printf("  id=%d name=%q (%.5f,%.5f) current=%q derived=%q action=skip-not-default (use -force to override)",
				row.ID, row.Name, row.Latitude, row.Longitude, row.Timezone, derived)
			skipped++
			continue
		}

		action := "update"
		if *dryRun {
			action = "would-update"
		}
		log.Printf("  id=%d name=%q (%.5f,%.5f) current=%q -> derived=%q action=%s",
			row.ID, row.Name, row.Latitude, row.Longitude, row.Timezone, derived, action)

		if *dryRun {
			updated++
			continue
		}

		if _, err := db.Exec(
			`UPDATE woulder.locations SET timezone = $1 WHERE id = $2`,
			derived, row.ID,
		); err != nil {
			log.Printf("    ✗ UPDATE failed for id=%d: %v", row.ID, err)
			skipped++
			continue
		}
		updated++
	}

	elapsed := time.Since(startedAt)
	log.Println()
	log.Println("=== Backfill Complete ===")
	log.Printf("Summary: scanned=%d, updated=%d, skipped=%d, dry-run=%v",
		scanned, updated, skipped, *dryRun)
	log.Printf("Elapsed: %s", elapsed.Round(time.Millisecond))
}

// loadLocations reads (id, name, latitude, longitude, timezone) from
// woulder.locations using a bare database/sql query. We deliberately do not
// route through the locations repo because §1c of the per-location-timezone
// rollout has not landed yet, so the repo's Scan calls do not yet include
// the new column.
func loadLocations(db *sql.DB, onlyID int) ([]locationRow, error) {
	query := `SELECT id, name, latitude, longitude, timezone FROM woulder.locations`
	args := []any{}
	if onlyID != 0 {
		query += ` WHERE id = $1`
		args = append(args, onlyID)
	}
	query += ` ORDER BY id`

	rs, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query locations: %w", err)
	}
	defer rs.Close()

	var out []locationRow
	for rs.Next() {
		var r locationRow
		if err := rs.Scan(&r.ID, &r.Name, &r.Latitude, &r.Longitude, &r.Timezone); err != nil {
			return nil, fmt.Errorf("scan location: %w", err)
		}
		out = append(out, r)
	}
	return out, rs.Err()
}

func getEnvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
