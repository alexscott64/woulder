package locations_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alexscott64/woulder/backend/internal/database"
	"github.com/alexscott64/woulder/backend/internal/database/locations"
)

func TestPostgresRepository_GetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "name", "latitude", "longitude", "elevation_ft",
		"area_id", "has_seepage_risk", "created_at", "updated_at",
	}).AddRow(
		1, "Smith Rock", 44.3672, -121.1423, 3200,
		1, false, now, now,
	).AddRow(
		2, "Index Town Wall", 47.8203, -121.5565, 1500,
		1, true, now, now,
	)

	mock.ExpectQuery("SELECT (.+) FROM woulder.locations").
		WillReturnRows(rows)

	repo := locations.NewPostgresRepository(db)
	result, err := repo.GetAll(context.Background())

	if err != nil {
		t.Errorf("GetAll() error = %v", err)
	}

	if len(result) != 2 {
		t.Errorf("GetAll() returned %d locations, want 2", len(result))
	}

	if result[0].Name != "Smith Rock" {
		t.Errorf("GetAll() first location name = %v, want Smith Rock", result[0].Name)
	}

	if result[1].Name != "Index Town Wall" {
		t.Errorf("GetAll() second location name = %v, want Index Town Wall", result[1].Name)
	}

	if result[1].HasSeepageRisk != true {
		t.Error("GetAll() second location should have seepage risk")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetAll_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"id", "name", "latitude", "longitude", "elevation_ft",
		"area_id", "has_seepage_risk", "created_at", "updated_at",
	})

	mock.ExpectQuery("SELECT (.+) FROM woulder.locations").
		WillReturnRows(rows)

	repo := locations.NewPostgresRepository(db)
	result, err := repo.GetAll(context.Background())

	if err != nil {
		t.Errorf("GetAll() error = %v, want nil", err)
	}

	if len(result) != 0 {
		t.Errorf("GetAll() returned %d locations, want 0", len(result))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "name", "latitude", "longitude", "elevation_ft",
		"area_id", "has_seepage_risk", "created_at", "updated_at",
	}).AddRow(
		5, "Smith Rock", 44.3672, -121.1423, 3200,
		1, false, now, now,
	)

	mock.ExpectQuery("SELECT (.+) FROM woulder.locations WHERE id").
		WithArgs(5).
		WillReturnRows(rows)

	repo := locations.NewPostgresRepository(db)
	result, err := repo.GetByID(context.Background(), 5)

	if err != nil {
		t.Errorf("GetByID() error = %v", err)
	}

	if result.ID != 5 {
		t.Errorf("GetByID() ID = %v, want 5", result.ID)
	}

	if result.Name != "Smith Rock" {
		t.Errorf("GetByID() name = %v, want Smith Rock", result.Name)
	}

	if result.ElevationFt != 3200 {
		t.Errorf("GetByID() elevation = %v, want 3200", result.ElevationFt)
	}

	if result.Latitude != 44.3672 {
		t.Errorf("GetByID() latitude = %v, want 44.3672", result.Latitude)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT (.+) FROM woulder.locations WHERE id").
		WithArgs(999).
		WillReturnError(database.ErrNotFound)

	repo := locations.NewPostgresRepository(db)
	_, err = repo.GetByID(context.Background(), 999)

	if err == nil {
		t.Error("GetByID() expected error for non-existent location, got nil")
	}

	if !database.IsNotFound(err) {
		t.Errorf("GetByID() expected ErrNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetByArea(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "name", "latitude", "longitude", "elevation_ft",
		"area_id", "has_seepage_risk", "created_at", "updated_at",
	}).AddRow(
		1, "Smith Rock", 44.3672, -121.1423, 3200,
		10, false, now, now,
	).AddRow(
		2, "Lower Index", 47.8203, -121.5565, 1500,
		10, true, now, now,
	).AddRow(
		3, "Upper Index", 47.8250, -121.5580, 1700,
		10, true, now, now,
	)

	mock.ExpectQuery("SELECT (.+) FROM woulder.locations WHERE area_id").
		WithArgs(10).
		WillReturnRows(rows)

	repo := locations.NewPostgresRepository(db)
	result, err := repo.GetByArea(context.Background(), 10)

	if err != nil {
		t.Errorf("GetByArea() error = %v", err)
	}

	if len(result) != 3 {
		t.Errorf("GetByArea() returned %d locations, want 3", len(result))
	}

	// Verify all locations belong to the same area
	for i, loc := range result {
		if loc.AreaID != 10 {
			t.Errorf("GetByArea() location %d has area_id = %v, want 10", i, loc.AreaID)
		}
	}

	if result[0].Name != "Smith Rock" {
		t.Errorf("GetByArea() first location name = %v, want Smith Rock", result[0].Name)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetByArea_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"id", "name", "latitude", "longitude", "elevation_ft",
		"area_id", "has_seepage_risk", "created_at", "updated_at",
	})

	mock.ExpectQuery("SELECT (.+) FROM woulder.locations WHERE area_id").
		WithArgs(999).
		WillReturnRows(rows)

	repo := locations.NewPostgresRepository(db)
	result, err := repo.GetByArea(context.Background(), 999)

	if err != nil {
		t.Errorf("GetByArea() error = %v, want nil", err)
	}

	if len(result) != 0 {
		t.Errorf("GetByArea() returned %d locations, want 0", len(result))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
