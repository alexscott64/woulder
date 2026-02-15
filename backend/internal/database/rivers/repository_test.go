package rivers_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alexscott64/woulder/backend/internal/database"
	"github.com/alexscott64/woulder/backend/internal/database/rivers"
	"github.com/alexscott64/woulder/backend/internal/models"
)

func TestPostgresRepository_GetByLocation(t *testing.T) {
	// Create mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	// Expected river data
	now := time.Now()
	drainageArea := 25.0
	gaugeDrainageArea := 30.0
	flowDivisor := 1.2
	description := "Test river description"
	expectedRivers := []models.River{
		{
			ID:                    1,
			LocationID:            10,
			GaugeID:               "12345678",
			RiverName:             "Test River",
			SafeCrossingCFS:       50,
			CautionCrossingCFS:    100,
			DrainageAreaSqMi:      &drainageArea,
			GaugeDrainageAreaSqMi: &gaugeDrainageArea,
			FlowDivisor:           &flowDivisor,
			IsEstimated:           false,
			Description:           &description,
			CreatedAt:             now,
			UpdatedAt:             now,
		},
	}

	// Setup mock expectations
	rows := sqlmock.NewRows([]string{
		"id", "location_id", "gauge_id", "river_name",
		"safe_crossing_cfs", "caution_crossing_cfs",
		"drainage_area_sq_mi", "gauge_drainage_area_sq_mi",
		"flow_divisor", "is_estimated", "description",
		"created_at", "updated_at",
	}).AddRow(
		expectedRivers[0].ID,
		expectedRivers[0].LocationID,
		expectedRivers[0].GaugeID,
		expectedRivers[0].RiverName,
		expectedRivers[0].SafeCrossingCFS,
		expectedRivers[0].CautionCrossingCFS,
		expectedRivers[0].DrainageAreaSqMi,
		expectedRivers[0].GaugeDrainageAreaSqMi,
		expectedRivers[0].FlowDivisor,
		expectedRivers[0].IsEstimated,
		expectedRivers[0].Description,
		expectedRivers[0].CreatedAt,
		expectedRivers[0].UpdatedAt,
	)

	mock.ExpectQuery("SELECT (.+) FROM woulder.rivers WHERE location_id").
		WithArgs(10).
		WillReturnRows(rows)

	// Create repository and execute
	repo := rivers.NewPostgresRepository(db)
	result, err := repo.GetByLocation(context.Background(), 10)

	// Assert
	if err != nil {
		t.Errorf("GetByLocation() error = %v", err)
	}

	if len(result) != 1 {
		t.Errorf("GetByLocation() returned %d rivers, want 1", len(result))
	}

	if result[0].RiverName != expectedRivers[0].RiverName {
		t.Errorf("GetByLocation() river name = %v, want %v", result[0].RiverName, expectedRivers[0].RiverName)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetByLocation_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	// Setup mock to return no rows
	rows := sqlmock.NewRows([]string{
		"id", "location_id", "gauge_id", "river_name",
		"safe_crossing_cfs", "caution_crossing_cfs",
		"drainage_area_sq_mi", "gauge_drainage_area_sq_mi",
		"flow_divisor", "is_estimated", "description",
		"created_at", "updated_at",
	})

	mock.ExpectQuery("SELECT (.+) FROM woulder.rivers WHERE location_id").
		WithArgs(999).
		WillReturnRows(rows)

	repo := rivers.NewPostgresRepository(db)
	result, err := repo.GetByLocation(context.Background(), 999)

	if err != nil {
		t.Errorf("GetByLocation() error = %v, want nil", err)
	}

	if len(result) != 0 {
		t.Errorf("GetByLocation() returned %d rivers, want 0", len(result))
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
	drainageArea := 25.0
	gaugeDrainageArea := 30.0
	flowDivisor := 1.2
	description := "Test river description"
	expectedRiver := &models.River{
		ID:                    1,
		LocationID:            10,
		GaugeID:               "12345678",
		RiverName:             "Test River",
		SafeCrossingCFS:       50,
		CautionCrossingCFS:    100,
		DrainageAreaSqMi:      &drainageArea,
		GaugeDrainageAreaSqMi: &gaugeDrainageArea,
		FlowDivisor:           &flowDivisor,
		IsEstimated:           false,
		Description:           &description,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	rows := sqlmock.NewRows([]string{
		"id", "location_id", "gauge_id", "river_name",
		"safe_crossing_cfs", "caution_crossing_cfs",
		"drainage_area_sq_mi", "gauge_drainage_area_sq_mi",
		"flow_divisor", "is_estimated", "description",
		"created_at", "updated_at",
	}).AddRow(
		expectedRiver.ID,
		expectedRiver.LocationID,
		expectedRiver.GaugeID,
		expectedRiver.RiverName,
		expectedRiver.SafeCrossingCFS,
		expectedRiver.CautionCrossingCFS,
		expectedRiver.DrainageAreaSqMi,
		expectedRiver.GaugeDrainageAreaSqMi,
		expectedRiver.FlowDivisor,
		expectedRiver.IsEstimated,
		expectedRiver.Description,
		expectedRiver.CreatedAt,
		expectedRiver.UpdatedAt,
	)

	mock.ExpectQuery("SELECT (.+) FROM woulder.rivers WHERE id").
		WithArgs(1).
		WillReturnRows(rows)

	repo := rivers.NewPostgresRepository(db)
	result, err := repo.GetByID(context.Background(), 1)

	if err != nil {
		t.Errorf("GetByID() error = %v", err)
	}

	if result.ID != expectedRiver.ID {
		t.Errorf("GetByID() ID = %v, want %v", result.ID, expectedRiver.ID)
	}

	if result.RiverName != expectedRiver.RiverName {
		t.Errorf("GetByID() river name = %v, want %v", result.RiverName, expectedRiver.RiverName)
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

	mock.ExpectQuery("SELECT (.+) FROM woulder.rivers WHERE id").
		WithArgs(999).
		WillReturnError(database.ErrNotFound)

	repo := rivers.NewPostgresRepository(db)
	_, err = repo.GetByID(context.Background(), 999)

	if err == nil {
		t.Error("GetByID() expected error for non-existent river, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
