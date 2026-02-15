package boulders_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alexscott64/woulder/backend/internal/database/boulders"
	"github.com/alexscott64/woulder/backend/internal/models"
)

func TestPostgresRepository_GetProfile(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	treeCoverage := 35.5
	rockType := "Granite"
	sunCache := `{"morning": 3.5, "afternoon": 4.2}`

	rows := sqlmock.NewRows([]string{
		"id", "mp_route_id", "tree_coverage_percent", "rock_type_override",
		"last_sun_calc_at", "sun_exposure_hours_cache", "created_at", "updated_at",
	}).AddRow(
		1, int64(12345), &treeCoverage, &rockType,
		&now, &sunCache, now, now,
	)

	mock.ExpectQuery("SELECT (.+) FROM woulder.boulder_drying_profiles WHERE mp_route_id").
		WithArgs(int64(12345)).
		WillReturnRows(rows)

	repo := boulders.NewPostgresRepository(db)
	result, err := repo.GetProfile(context.Background(), 12345)

	if err != nil {
		t.Errorf("GetProfile() error = %v", err)
	}

	if result == nil {
		t.Fatal("GetProfile() returned nil, expected profile")
	}

	if result.MPRouteID != 12345 {
		t.Errorf("GetProfile() route_id = %v, want 12345", result.MPRouteID)
	}

	if result.TreeCoveragePercent == nil || *result.TreeCoveragePercent != 35.5 {
		t.Errorf("GetProfile() tree coverage = %v, want 35.5", result.TreeCoveragePercent)
	}

	if result.RockTypeOverride == nil || *result.RockTypeOverride != "Granite" {
		t.Errorf("GetProfile() rock type = %v, want Granite", result.RockTypeOverride)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetProfile_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"id", "mp_route_id", "tree_coverage_percent", "rock_type_override",
		"last_sun_calc_at", "sun_exposure_hours_cache", "created_at", "updated_at",
	})

	mock.ExpectQuery("SELECT (.+) FROM woulder.boulder_drying_profiles WHERE mp_route_id").
		WithArgs(int64(999)).
		WillReturnRows(rows)

	repo := boulders.NewPostgresRepository(db)
	result, err := repo.GetProfile(context.Background(), 999)

	if err != nil {
		t.Errorf("GetProfile() error = %v, want nil", err)
	}

	// No profile is not an error - returns nil
	if result != nil {
		t.Error("GetProfile() expected nil for missing profile, got data")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetProfilesByIDs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	treeCoverage1 := 35.5
	treeCoverage2 := 60.0
	rockType1 := "Granite"
	sunCache1 := `{"morning": 3.5}`

	rows := sqlmock.NewRows([]string{
		"id", "mp_route_id", "tree_coverage_percent", "rock_type_override",
		"last_sun_calc_at", "sun_exposure_hours_cache", "created_at", "updated_at",
	}).AddRow(
		1, int64(12345), &treeCoverage1, &rockType1,
		&now, &sunCache1, now, now,
	).AddRow(
		2, int64(67890), &treeCoverage2, nil,
		nil, nil, now, now,
	)

	mock.ExpectQuery("SELECT (.+) FROM woulder.boulder_drying_profiles WHERE mp_route_id = ANY").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(rows)

	repo := boulders.NewPostgresRepository(db)
	routeIDs := []int64{12345, 67890}
	result, err := repo.GetProfilesByIDs(context.Background(), routeIDs)

	if err != nil {
		t.Errorf("GetProfilesByIDs() error = %v", err)
	}

	if len(result) != 2 {
		t.Errorf("GetProfilesByIDs() returned %d profiles, want 2", len(result))
	}

	profile1, ok := result[12345]
	if !ok {
		t.Error("GetProfilesByIDs() missing profile for route 12345")
	} else {
		if profile1.TreeCoveragePercent == nil || *profile1.TreeCoveragePercent != 35.5 {
			t.Errorf("GetProfilesByIDs() route 12345 tree coverage = %v, want 35.5", profile1.TreeCoveragePercent)
		}
	}

	profile2, ok := result[67890]
	if !ok {
		t.Error("GetProfilesByIDs() missing profile for route 67890")
	} else {
		if profile2.TreeCoveragePercent == nil || *profile2.TreeCoveragePercent != 60.0 {
			t.Errorf("GetProfilesByIDs() route 67890 tree coverage = %v, want 60.0", profile2.TreeCoveragePercent)
		}
		if profile2.RockTypeOverride != nil {
			t.Error("GetProfilesByIDs() route 67890 should have nil rock type override")
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetProfilesByIDs_Empty(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := boulders.NewPostgresRepository(db)
	result, err := repo.GetProfilesByIDs(context.Background(), []int64{})

	if err != nil {
		t.Errorf("GetProfilesByIDs() error = %v, want nil", err)
	}

	if len(result) != 0 {
		t.Errorf("GetProfilesByIDs() returned %d profiles, want 0", len(result))
	}
}

func TestPostgresRepository_SaveProfile(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	treeCoverage := 45.0
	rockType := "Sandstone"
	sunCache := `{"morning": 2.0, "afternoon": 3.5}`

	profile := &models.BoulderDryingProfile{
		MPRouteID:             12345,
		TreeCoveragePercent:   &treeCoverage,
		RockTypeOverride:      &rockType,
		LastSunCalcAt:         &now,
		SunExposureHoursCache: &sunCache,
	}

	mock.ExpectExec("INSERT INTO woulder.boulder_drying_profiles").
		WithArgs(profile.MPRouteID, profile.TreeCoveragePercent, profile.RockTypeOverride,
			profile.LastSunCalcAt, profile.SunExposureHoursCache).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := boulders.NewPostgresRepository(db)
	err = repo.SaveProfile(context.Background(), profile)

	if err != nil {
		t.Errorf("SaveProfile() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_SaveProfile_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	treeCoverage := 55.0

	profile := &models.BoulderDryingProfile{
		MPRouteID:           12345,
		TreeCoveragePercent: &treeCoverage,
		LastSunCalcAt:       &now,
	}

	// ON CONFLICT DO UPDATE
	mock.ExpectExec("INSERT INTO woulder.boulder_drying_profiles").
		WithArgs(profile.MPRouteID, profile.TreeCoveragePercent, profile.RockTypeOverride,
								profile.LastSunCalcAt, profile.SunExposureHoursCache).
		WillReturnResult(sqlmock.NewResult(0, 1)) // 0 insert, 1 update

	repo := boulders.NewPostgresRepository(db)
	err = repo.SaveProfile(context.Background(), profile)

	if err != nil {
		t.Errorf("SaveProfile() update error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
