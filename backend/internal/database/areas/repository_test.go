package areas_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alexscott64/woulder/backend/internal/database"
	"github.com/alexscott64/woulder/backend/internal/database/areas"
)

func TestPostgresRepository_GetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	desc1 := "Seattle, Portland, and surrounding areas"
	region1 := "West"
	desc2 := "Arizona, Nevada, Utah climbing areas"
	region2 := "Southwest"

	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "region",
		"display_order", "is_active", "created_at", "updated_at",
	}).AddRow(
		1, "Pacific Northwest", &desc1, &region1,
		1, true, now, now,
	).AddRow(
		2, "Southwest", &desc2, &region2,
		2, true, now, now,
	)

	mock.ExpectQuery("SELECT (.+) FROM woulder.areas WHERE is_active").
		WillReturnRows(rows)

	repo := areas.NewPostgresRepository(db)
	result, err := repo.GetAll(context.Background())

	if err != nil {
		t.Errorf("GetAll() error = %v", err)
	}

	if len(result) != 2 {
		t.Errorf("GetAll() returned %d areas, want 2", len(result))
	}

	if result[0].Name != "Pacific Northwest" {
		t.Errorf("GetAll() first area name = %v, want Pacific Northwest", result[0].Name)
	}

	if result[1].Name != "Southwest" {
		t.Errorf("GetAll() second area name = %v, want Southwest", result[1].Name)
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
		"id", "name", "description", "region",
		"display_order", "is_active", "created_at", "updated_at",
	})

	mock.ExpectQuery("SELECT (.+) FROM woulder.areas WHERE is_active").
		WillReturnRows(rows)

	repo := areas.NewPostgresRepository(db)
	result, err := repo.GetAll(context.Background())

	if err != nil {
		t.Errorf("GetAll() error = %v, want nil", err)
	}

	if len(result) != 0 {
		t.Errorf("GetAll() returned %d areas, want 0", len(result))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetAllWithLocationCounts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	desc1 := "Seattle, Portland areas"
	region1 := "West"
	desc2 := "Arizona, Nevada, Utah"
	region2 := "Southwest"

	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "region",
		"display_order", "is_active", "created_at", "updated_at",
		"location_count",
	}).AddRow(
		1, "Pacific Northwest", &desc1, &region1,
		1, true, now, now,
		15, // location count
	).AddRow(
		2, "Southwest", &desc2, &region2,
		2, true, now, now,
		23, // location count
	)

	mock.ExpectQuery("SELECT (.+) FROM woulder.areas a LEFT JOIN woulder.locations").
		WillReturnRows(rows)

	repo := areas.NewPostgresRepository(db)
	result, err := repo.GetAllWithLocationCounts(context.Background())

	if err != nil {
		t.Errorf("GetAllWithLocationCounts() error = %v", err)
	}

	if len(result) != 2 {
		t.Errorf("GetAllWithLocationCounts() returned %d areas, want 2", len(result))
	}

	if result[0].LocationCount != 15 {
		t.Errorf("GetAllWithLocationCounts() first area count = %v, want 15", result[0].LocationCount)
	}

	if result[1].LocationCount != 23 {
		t.Errorf("GetAllWithLocationCounts() second area count = %v, want 23", result[1].LocationCount)
	}

	if result[0].Name != "Pacific Northwest" {
		t.Errorf("GetAllWithLocationCounts() first area name = %v, want Pacific Northwest", result[0].Name)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetAllWithLocationCounts_ZeroLocations(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	desc := "Brand new area"
	region := "East"

	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "region",
		"display_order", "is_active", "created_at", "updated_at",
		"location_count",
	}).AddRow(
		1, "New Area", &desc, &region,
		1, true, now, now,
		0, // no locations yet
	)

	mock.ExpectQuery("SELECT (.+) FROM woulder.areas a LEFT JOIN woulder.locations").
		WillReturnRows(rows)

	repo := areas.NewPostgresRepository(db)
	result, err := repo.GetAllWithLocationCounts(context.Background())

	if err != nil {
		t.Errorf("GetAllWithLocationCounts() error = %v", err)
	}

	if len(result) != 1 {
		t.Errorf("GetAllWithLocationCounts() returned %d areas, want 1", len(result))
	}

	if result[0].LocationCount != 0 {
		t.Errorf("GetAllWithLocationCounts() area count = %v, want 0", result[0].LocationCount)
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
	desc := "Seattle, Portland, and surrounding areas"
	region := "West"

	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "region",
		"display_order", "is_active", "created_at", "updated_at",
	}).AddRow(
		5, "Pacific Northwest", &desc, &region,
		1, true, now, now,
	)

	mock.ExpectQuery("SELECT (.+) FROM woulder.areas WHERE id = (.+) AND is_active").
		WithArgs(5).
		WillReturnRows(rows)

	repo := areas.NewPostgresRepository(db)
	result, err := repo.GetByID(context.Background(), 5)

	if err != nil {
		t.Errorf("GetByID() error = %v", err)
	}

	if result.ID != 5 {
		t.Errorf("GetByID() ID = %v, want 5", result.ID)
	}

	if result.Name != "Pacific Northwest" {
		t.Errorf("GetByID() name = %v, want Pacific Northwest", result.Name)
	}

	if result.Region == nil || *result.Region != "West" {
		t.Errorf("GetByID() region = %v, want West", result.Region)
	}

	if !result.IsActive {
		t.Error("GetByID() expected is_active = true")
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

	mock.ExpectQuery("SELECT (.+) FROM woulder.areas WHERE id = (.+) AND is_active").
		WithArgs(999).
		WillReturnError(database.ErrNotFound)

	repo := areas.NewPostgresRepository(db)
	_, err = repo.GetByID(context.Background(), 999)

	if err == nil {
		t.Error("GetByID() expected error for non-existent area, got nil")
	}

	if !database.IsNotFound(err) {
		t.Errorf("GetByID() expected ErrNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetByID_Inactive(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	// Simulate that the area exists but is inactive (query returns no rows)
	mock.ExpectQuery("SELECT (.+) FROM woulder.areas WHERE id = (.+) AND is_active").
		WithArgs(10).
		WillReturnError(database.ErrNotFound)

	repo := areas.NewPostgresRepository(db)
	_, err = repo.GetByID(context.Background(), 10)

	if err == nil {
		t.Error("GetByID() expected error for inactive area, got nil")
	}

	if !database.IsNotFound(err) {
		t.Errorf("GetByID() expected ErrNotFound for inactive area, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
