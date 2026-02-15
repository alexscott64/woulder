package database

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/alexscott64/woulder/backend/internal/database/areas"
	"github.com/alexscott64/woulder/backend/internal/database/boulders"
	"github.com/alexscott64/woulder/backend/internal/database/climbing"
	"github.com/alexscott64/woulder/backend/internal/database/heatmap"
	"github.com/alexscott64/woulder/backend/internal/database/locations"
	"github.com/alexscott64/woulder/backend/internal/database/mountainproject"
	"github.com/alexscott64/woulder/backend/internal/database/rivers"
	"github.com/alexscott64/woulder/backend/internal/database/rocks"
	"github.com/alexscott64/woulder/backend/internal/database/weather"
	_ "github.com/lib/pq"
)

//go:embed setup_postgres.sql
var setupSQL string

type Database struct {
	conn *sql.DB
}

func New() (*Database, error) {
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

	if host == "" || user == "" || password == "" || dbname == "" {
		return nil, fmt.Errorf("missing required database configuration")
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	database := &Database{conn: db}

	needsInit, err := database.needsInitialization()
	if err != nil {
		return nil, err
	}

	if needsInit {
		log.Println("Database schema not found, running setup...")
		if err := database.runSetup(); err != nil {
			return nil, err
		}
	}

	return database, nil
}

func (db *Database) needsInitialization() (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM information_schema.schemata WHERE schema_name = 'woulder')`
	err := db.conn.QueryRow(query).Scan(&exists)
	return !exists, err
}

func (db *Database) runSetup() error {
	_, err := db.conn.Exec(setupSQL)
	return err
}

func (db *Database) Close() error {
	return db.conn.Close()
}

func (db *Database) Ping(ctx context.Context) error {
	return db.conn.PingContext(ctx)
}

// Domain Repository Accessors
// These methods provide unified access to domain-specific repositories,
// breaking import cycles by having the database package import domains
// instead of domains importing database.

// Rivers returns the rivers repository for river data operations.
func (db *Database) Rivers() rivers.Repository {
	return rivers.NewPostgresRepository(db.conn)
}

// Weather returns the weather repository for weather data operations.
func (db *Database) Weather() weather.Repository {
	return weather.NewPostgresRepository(db.conn)
}

// Areas returns the areas repository for geographic area operations.
func (db *Database) Areas() areas.Repository {
	return areas.NewPostgresRepository(db.conn)
}

// Locations returns the locations repository for climbing location operations.
func (db *Database) Locations() locations.Repository {
	return locations.NewPostgresRepository(db.conn)
}

// Rocks returns the rocks repository for rock type and sun exposure operations.
func (db *Database) Rocks() rocks.Repository {
	return rocks.NewPostgresRepository(db.conn)
}

// Boulders returns the boulders repository for boulder drying profile operations.
func (db *Database) Boulders() boulders.Repository {
	return boulders.NewPostgresRepository(db.conn)
}

// HeatMap returns the heatmap repository for activity visualization operations.
func (db *Database) HeatMap() heatmap.Repository {
	return heatmap.NewPostgresRepository(db.conn)
}

// Climbing returns the climbing repository for activity and history operations.
func (db *Database) Climbing() climbing.Repository {
	return climbing.NewPostgresRepository(db.conn)
}

// MountainProject returns the Mountain Project repository for MP data operations.
func (db *Database) MountainProject() mountainproject.Repository {
	return mountainproject.NewPostgresRepository(db.conn)
}
