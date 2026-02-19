package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Migration struct {
	Version  int
	Name     string
	UpPath   string
	DownPath string
}

func main() {
	// Load .env file if it exists
	// Try loading from backend/.env first (when running from backend/cmd/migrate)
	envPath := ".env"
	if err := godotenv.Load(envPath); err != nil {
		// Try loading from project root (when running from backend directory)
		envPath = filepath.Join("..", ".env")
		if err := godotenv.Load(envPath); err != nil {
			log.Println("Warning: .env file not found, using environment variables")
		}
	}

	// Get database connection details from environment
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

	// Construct connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Create schema_migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		log.Fatalf("Failed to create migrations table: %v", err)
	}

	// Parse command
	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	// Get migrations directory
	migrationsPath := filepath.Join("..", "..", "internal", "database", "migrations")

	// Execute command
	switch command {
	case "up":
		if err := migrateUp(db, migrationsPath); err != nil {
			log.Fatalf("Migration up failed: %v", err)
		}

	case "down":
		if err := migrateDown(db, migrationsPath); err != nil {
			log.Fatalf("Migration down failed: %v", err)
		}

	case "version":
		version, err := getCurrentVersion(db)
		if err != nil {
			log.Fatalf("Failed to get version: %v", err)
		}
		if version == 0 {
			log.Println("No migrations have been run yet")
		} else {
			log.Printf("Current version: %d\n", version)
		}

	case "force":
		if len(os.Args) < 3 {
			log.Fatal("Usage: migrate force <version>")
		}
		version, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatalf("Invalid version number: %v", err)
		}
		if err := forceVersion(db, version); err != nil {
			log.Fatalf("Force version failed: %v", err)
		}
		log.Printf("✓ Forced version to %d\n", version)

	case "step":
		if len(os.Args) < 3 {
			log.Fatal("Usage: migrate step <n>")
		}
		steps, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatalf("Invalid step number: %v", err)
		}
		if err := migrateSteps(db, migrationsPath, steps); err != nil {
			log.Fatalf("Migration step failed: %v", err)
		}

	case "help":
		printHelp()

	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printHelp()
		os.Exit(1)
	}
}

func createMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

func getCurrentVersion(db *sql.DB) (int, error) {
	var version int
	err := db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

func forceVersion(db *sql.DB, version int) error {
	_, err := db.Exec("DELETE FROM schema_migrations")
	if err != nil {
		return err
	}
	if version > 0 {
		_, err = db.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version)
	}
	return err
}

func loadMigrations(migrationsPath string) ([]Migration, error) {
	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		return nil, err
	}

	migrationsMap := make(map[int]*Migration)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}

		// Parse filename: 000001_initial_schema.up.sql
		parts := strings.SplitN(name, "_", 2)
		if len(parts) < 2 {
			continue
		}

		version, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}

		migration, exists := migrationsMap[version]
		if !exists {
			migration = &Migration{Version: version}
			migrationsMap[version] = migration
		}

		fullPath := filepath.Join(migrationsPath, name)

		if strings.HasSuffix(name, ".up.sql") {
			migration.UpPath = fullPath
			migration.Name = strings.TrimSuffix(parts[1], ".up.sql")
		} else if strings.HasSuffix(name, ".down.sql") {
			migration.DownPath = fullPath
			if migration.Name == "" {
				migration.Name = strings.TrimSuffix(parts[1], ".down.sql")
			}
		}
	}

	// Convert map to sorted slice
	migrations := make([]Migration, 0, len(migrationsMap))
	for _, m := range migrationsMap {
		migrations = append(migrations, *m)
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func migrateUp(db *sql.DB, migrationsPath string) error {
	log.Println("Running migrations up...")

	currentVersion, err := getCurrentVersion(db)
	if err != nil {
		return err
	}

	migrations, err := loadMigrations(migrationsPath)
	if err != nil {
		return err
	}

	appliedCount := 0
	for _, migration := range migrations {
		if migration.Version <= currentVersion {
			continue
		}

		if migration.UpPath == "" {
			return fmt.Errorf("missing up migration for version %d", migration.Version)
		}

		log.Printf("Applying migration %d: %s...", migration.Version, migration.Name)

		sqlContent, err := os.ReadFile(migration.UpPath)
		if err != nil {
			return fmt.Errorf("failed to read migration file: %v", err)
		}

		sqlString := string(sqlContent)

		// Check if migration contains CONCURRENTLY (cannot run in transaction)
		usesConcurrently := strings.Contains(strings.ToUpper(sqlString), "CONCURRENTLY")

		if usesConcurrently {
			// Execute without transaction for CONCURRENT operations
			log.Printf("  (running without transaction due to CONCURRENTLY)")

			if _, err := db.Exec(sqlString); err != nil {
				return fmt.Errorf("migration %d failed: %v", migration.Version, err)
			}

			if _, err := db.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", migration.Version); err != nil {
				return fmt.Errorf("failed to record migration: %v", err)
			}
		} else {
			// Execute migration in a transaction (normal case)
			tx, err := db.Begin()
			if err != nil {
				return err
			}

			if _, err := tx.Exec(sqlString); err != nil {
				tx.Rollback()
				return fmt.Errorf("migration %d failed: %v", migration.Version, err)
			}

			if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", migration.Version); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to record migration: %v", err)
			}

			if err := tx.Commit(); err != nil {
				return fmt.Errorf("failed to commit migration: %v", err)
			}
		}

		log.Printf("✓ Applied migration %d", migration.Version)
		appliedCount++
	}

	if appliedCount == 0 {
		log.Println("✓ No migrations to run (already up to date)")
	} else {
		log.Printf("✓ Successfully applied %d migration(s)", appliedCount)
	}

	return nil
}

func migrateDown(db *sql.DB, migrationsPath string) error {
	log.Println("Rolling back migrations...")

	currentVersion, err := getCurrentVersion(db)
	if err != nil {
		return err
	}

	if currentVersion == 0 {
		log.Println("✓ No migrations to roll back")
		return nil
	}

	migrations, err := loadMigrations(migrationsPath)
	if err != nil {
		return err
	}

	// Rollback in reverse order
	for i := len(migrations) - 1; i >= 0; i-- {
		migration := migrations[i]
		if migration.Version > currentVersion {
			continue
		}

		if migration.DownPath == "" {
			return fmt.Errorf("missing down migration for version %d", migration.Version)
		}

		log.Printf("Rolling back migration %d: %s...", migration.Version, migration.Name)

		sqlContent, err := os.ReadFile(migration.DownPath)
		if err != nil {
			return fmt.Errorf("failed to read migration file: %v", err)
		}

		sqlString := string(sqlContent)

		// Check if rollback contains CONCURRENTLY (cannot run in transaction)
		usesConcurrently := strings.Contains(strings.ToUpper(sqlString), "CONCURRENTLY")

		if usesConcurrently {
			// Execute without transaction for CONCURRENT operations
			log.Printf("  (running without transaction due to CONCURRENTLY)")

			if _, err := db.Exec(sqlString); err != nil {
				return fmt.Errorf("rollback %d failed: %v", migration.Version, err)
			}

			if _, err := db.Exec("DELETE FROM schema_migrations WHERE version = $1", migration.Version); err != nil {
				return fmt.Errorf("failed to remove migration record: %v", err)
			}
		} else {
			// Execute rollback in a transaction (normal case)
			tx, err := db.Begin()
			if err != nil {
				return err
			}

			if _, err := tx.Exec(sqlString); err != nil {
				tx.Rollback()
				return fmt.Errorf("rollback %d failed: %v", migration.Version, err)
			}

			if _, err := tx.Exec("DELETE FROM schema_migrations WHERE version = $1", migration.Version); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to remove migration record: %v", err)
			}

			if err := tx.Commit(); err != nil {
				return fmt.Errorf("failed to commit rollback: %v", err)
			}
		}

		log.Printf("✓ Rolled back migration %d", migration.Version)
	}

	log.Println("✓ Rollback completed successfully!")
	return nil
}

func migrateSteps(db *sql.DB, migrationsPath string, steps int) error {
	if steps == 0 {
		log.Println("✓ No migrations to run")
		return nil
	}

	currentVersion, err := getCurrentVersion(db)
	if err != nil {
		return err
	}

	migrations, err := loadMigrations(migrationsPath)
	if err != nil {
		return err
	}

	if steps > 0 {
		// Step up
		log.Printf("Stepping up %d migration(s)...", steps)
		count := 0
		for _, migration := range migrations {
			if migration.Version <= currentVersion {
				continue
			}
			if count >= steps {
				break
			}

			if migration.UpPath == "" {
				return fmt.Errorf("missing up migration for version %d", migration.Version)
			}

			log.Printf("Applying migration %d: %s...", migration.Version, migration.Name)

			sqlContent, err := os.ReadFile(migration.UpPath)
			if err != nil {
				return fmt.Errorf("failed to read migration file: %v", err)
			}

			tx, err := db.Begin()
			if err != nil {
				return err
			}

			if _, err := tx.Exec(string(sqlContent)); err != nil {
				tx.Rollback()
				return fmt.Errorf("migration %d failed: %v", migration.Version, err)
			}

			if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", migration.Version); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to record migration: %v", err)
			}

			if err := tx.Commit(); err != nil {
				return fmt.Errorf("failed to commit migration: %v", err)
			}

			log.Printf("✓ Applied migration %d", migration.Version)
			count++
		}
		log.Printf("✓ Stepped up %d migration(s)", count)
	} else {
		// Step down (negative steps)
		steps = -steps
		log.Printf("Stepping down %d migration(s)...", steps)
		count := 0
		for i := len(migrations) - 1; i >= 0; i-- {
			migration := migrations[i]
			if migration.Version > currentVersion {
				continue
			}
			if count >= steps {
				break
			}

			if migration.DownPath == "" {
				return fmt.Errorf("missing down migration for version %d", migration.Version)
			}

			log.Printf("Rolling back migration %d: %s...", migration.Version, migration.Name)

			sqlContent, err := os.ReadFile(migration.DownPath)
			if err != nil {
				return fmt.Errorf("failed to read migration file: %v", err)
			}

			tx, err := db.Begin()
			if err != nil {
				return err
			}

			if _, err := tx.Exec(string(sqlContent)); err != nil {
				tx.Rollback()
				return fmt.Errorf("rollback %d failed: %v", migration.Version, err)
			}

			if _, err := tx.Exec("DELETE FROM schema_migrations WHERE version = $1", migration.Version); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to remove migration record: %v", err)
			}

			if err := tx.Commit(); err != nil {
				return fmt.Errorf("failed to commit rollback: %v", err)
			}

			log.Printf("✓ Rolled back migration %d", migration.Version)
			count++
		}
		log.Printf("✓ Stepped down %d migration(s)", count)
	}

	return nil
}

func printHelp() {
	fmt.Println("Woulder Database Migration Tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run cmd/migrate/main.go [command]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  up               Apply all pending migrations (default)")
	fmt.Println("  down             Rollback all migrations")
	fmt.Println("  version          Show current migration version")
	fmt.Println("  step <n>         Apply next n migrations (or rollback if negative)")
	fmt.Println("  force <version>  Force database to specific version (use with caution)")
	fmt.Println("  help             Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run cmd/migrate/main.go up")
	fmt.Println("  go run cmd/migrate/main.go down")
	fmt.Println("  go run cmd/migrate/main.go version")
	fmt.Println("  go run cmd/migrate/main.go step 1")
	fmt.Println("  go run cmd/migrate/main.go step -1")
	fmt.Println("  go run cmd/migrate/main.go force 2")
}
