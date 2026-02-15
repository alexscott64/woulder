package rocks_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alexscott64/woulder/backend/internal/database/rocks"
)

func TestPostgresRepository_GetRockTypesByLocation(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"id", "name", "base_drying_hours", "porosity_percent",
		"is_wet_sensitive", "description", "rock_type_group_id", "group_name",
	}).AddRow(
		1, "Granite", 4.0, 1.5,
		false, "Hard igneous rock", 1, "Igneous",
	).AddRow(
		2, "Basalt", 3.5, 2.0,
		false, "Volcanic rock", 1, "Igneous",
	)

	mock.ExpectQuery("SELECT (.+) FROM woulder.rock_types rt").
		WithArgs(10).
		WillReturnRows(rows)

	repo := rocks.NewPostgresRepository(db)
	result, err := repo.GetRockTypesByLocation(context.Background(), 10)

	if err != nil {
		t.Errorf("GetRockTypesByLocation() error = %v", err)
	}

	if len(result) != 2 {
		t.Errorf("GetRockTypesByLocation() returned %d rock types, want 2", len(result))
	}

	if result[0].Name != "Granite" {
		t.Errorf("GetRockTypesByLocation() first rock type = %v, want Granite", result[0].Name)
	}

	if result[0].BaseDryingHours != 4.0 {
		t.Errorf("GetRockTypesByLocation() granite drying hours = %v, want 4.0", result[0].BaseDryingHours)
	}

	if result[0].IsWetSensitive {
		t.Error("GetRockTypesByLocation() granite should not be wet sensitive")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetRockTypesByLocation_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"id", "name", "base_drying_hours", "porosity_percent",
		"is_wet_sensitive", "description", "rock_type_group_id", "group_name",
	})

	mock.ExpectQuery("SELECT (.+) FROM woulder.rock_types rt").
		WithArgs(999).
		WillReturnRows(rows)

	repo := rocks.NewPostgresRepository(db)
	result, err := repo.GetRockTypesByLocation(context.Background(), 999)

	if err != nil {
		t.Errorf("GetRockTypesByLocation() error = %v, want nil", err)
	}

	if len(result) != 0 {
		t.Errorf("GetRockTypesByLocation() returned %d rock types, want 0", len(result))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetPrimaryRockType(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"id", "name", "base_drying_hours", "porosity_percent",
		"is_wet_sensitive", "description", "rock_type_group_id", "group_name",
	}).AddRow(
		1, "Granite", 4.0, 1.5,
		false, "Hard igneous rock", 1, "Igneous",
	)

	mock.ExpectQuery("SELECT (.+) FROM woulder.rock_types rt(.+)is_primary = TRUE").
		WithArgs(10).
		WillReturnRows(rows)

	repo := rocks.NewPostgresRepository(db)
	result, err := repo.GetPrimaryRockType(context.Background(), 10)

	if err != nil {
		t.Errorf("GetPrimaryRockType() error = %v", err)
	}

	if result.ID != 1 {
		t.Errorf("GetPrimaryRockType() ID = %v, want 1", result.ID)
	}

	if result.Name != "Granite" {
		t.Errorf("GetPrimaryRockType() name = %v, want Granite", result.Name)
	}

	if result.GroupName != "Igneous" {
		t.Errorf("GetPrimaryRockType() group = %v, want Igneous", result.GroupName)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetPrimaryRockType_Fallback(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	// Primary query returns no rows (sql.ErrNoRows triggers fallback)
	rows := sqlmock.NewRows([]string{
		"id", "name", "base_drying_hours", "porosity_percent",
		"is_wet_sensitive", "description", "rock_type_group_id", "group_name",
	})

	mock.ExpectQuery("SELECT (.+) FROM woulder.rock_types rt(.+)is_primary = TRUE").
		WithArgs(10).
		WillReturnRows(rows) // Empty result set

	// Fallback query returns all rock types
	fallbackRows := sqlmock.NewRows([]string{
		"id", "name", "base_drying_hours", "porosity_percent",
		"is_wet_sensitive", "description", "rock_type_group_id", "group_name",
	}).AddRow(
		2, "Sandstone", 12.0, 15.0,
		true, "Sedimentary rock", 2, "Sedimentary",
	)

	mock.ExpectQuery("SELECT (.+) FROM woulder.rock_types rt").
		WithArgs(10).
		WillReturnRows(fallbackRows)

	repo := rocks.NewPostgresRepository(db)
	result, err := repo.GetPrimaryRockType(context.Background(), 10)

	if err != nil {
		t.Errorf("GetPrimaryRockType() error = %v", err)
	}

	if result == nil {
		t.Fatal("GetPrimaryRockType() returned nil, expected fallback result")
	}

	if result.Name != "Sandstone" {
		t.Errorf("GetPrimaryRockType() fallback name = %v, want Sandstone", result.Name)
	}

	if !result.IsWetSensitive {
		t.Error("GetPrimaryRockType() sandstone should be wet sensitive")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetPrimaryRockType_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	// Primary query returns no rows
	rows := sqlmock.NewRows([]string{
		"id", "name", "base_drying_hours", "porosity_percent",
		"is_wet_sensitive", "description", "rock_type_group_id", "group_name",
	})

	mock.ExpectQuery("SELECT (.+) FROM woulder.rock_types rt(.+)is_primary = TRUE").
		WithArgs(999).
		WillReturnRows(rows) // Empty result set

	// Fallback query also returns empty
	fallbackRows := sqlmock.NewRows([]string{
		"id", "name", "base_drying_hours", "porosity_percent",
		"is_wet_sensitive", "description", "rock_type_group_id", "group_name",
	})

	mock.ExpectQuery("SELECT (.+) FROM woulder.rock_types rt").
		WithArgs(999).
		WillReturnRows(fallbackRows)

	repo := rocks.NewPostgresRepository(db)
	_, err = repo.GetPrimaryRockType(context.Background(), 999)

	if err == nil {
		t.Error("GetPrimaryRockType() expected error for location with no rock types, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetSunExposureByLocation(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"id", "location_id", "south_facing_percent", "west_facing_percent",
		"east_facing_percent", "north_facing_percent", "slab_percent",
		"overhang_percent", "tree_coverage_percent", "description",
	}).AddRow(
		1, 10, 60.0, 20.0,
		15.0, 5.0, 40.0,
		30.0, 25.0, "South-facing crag with moderate tree cover",
	)

	mock.ExpectQuery("SELECT (.+) FROM woulder.location_sun_exposure").
		WithArgs(10).
		WillReturnRows(rows)

	repo := rocks.NewPostgresRepository(db)
	result, err := repo.GetSunExposureByLocation(context.Background(), 10)

	if err != nil {
		t.Errorf("GetSunExposureByLocation() error = %v", err)
	}

	if result == nil {
		t.Fatal("GetSunExposureByLocation() returned nil, expected data")
	}

	if result.SouthFacingPercent != 60.0 {
		t.Errorf("GetSunExposureByLocation() south facing = %v, want 60.0", result.SouthFacingPercent)
	}

	if result.TreeCoveragePercent != 25.0 {
		t.Errorf("GetSunExposureByLocation() tree coverage = %v, want 25.0", result.TreeCoveragePercent)
	}

	if result.OverhangPercent != 30.0 {
		t.Errorf("GetSunExposureByLocation() overhang = %v, want 30.0", result.OverhangPercent)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetSunExposureByLocation_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	// Return empty rows (sql.ErrNoRows)
	rows := sqlmock.NewRows([]string{
		"id", "location_id", "south_facing_percent", "west_facing_percent",
		"east_facing_percent", "north_facing_percent", "slab_percent",
		"overhang_percent", "tree_coverage_percent", "description",
	})

	mock.ExpectQuery("SELECT (.+) FROM woulder.location_sun_exposure").
		WithArgs(999).
		WillReturnRows(rows)

	repo := rocks.NewPostgresRepository(db)
	result, err := repo.GetSunExposureByLocation(context.Background(), 999)

	// No sun exposure data is NOT an error - returns nil
	if err != nil {
		t.Errorf("GetSunExposureByLocation() error = %v, want nil", err)
	}

	if result != nil {
		t.Errorf("GetSunExposureByLocation() returned data, want nil for missing exposure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
