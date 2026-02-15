package climbing_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alexscott64/woulder/backend/internal/database/climbing"
)

// ====================
// History Tests
// ====================

func TestPostgresRepository_GetLastClimbedForLocation(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"route_name", "route_rating", "climbed_at", "climbed_by",
		"style", "comment", "days_since_climb",
	}).AddRow(
		"Monkey Face", "5.13a", time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC),
		"john_doe", "Lead", "Great climb!", 10,
	)

	mock.ExpectQuery(`SELECT\s+r\.name AS route_name`).
		WithArgs(10).
		WillReturnRows(rows)

	repo := climbing.NewPostgresRepository(db)
	result, err := repo.History().GetLastClimbedForLocation(context.Background(), 10)

	if err != nil {
		t.Errorf("GetLastClimbedForLocation() error = %v", err)
	}

	if result == nil {
		t.Fatal("GetLastClimbedForLocation() returned nil")
	}

	if result.RouteName != "Monkey Face" {
		t.Errorf("GetLastClimbedForLocation() route name = %v, want Monkey Face", result.RouteName)
	}

	if result.DaysSinceClimb != 10 {
		t.Errorf("GetLastClimbedForLocation() days since = %v, want 10", result.DaysSinceClimb)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetLastClimbedForLocation_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"route_name", "route_rating", "climbed_at", "climbed_by",
		"style", "comment", "days_since_climb",
	})

	mock.ExpectQuery(`SELECT\s+r\.name AS route_name`).
		WithArgs(999).
		WillReturnRows(rows)

	repo := climbing.NewPostgresRepository(db)
	result, err := repo.History().GetLastClimbedForLocation(context.Background(), 999)

	if err != nil {
		t.Errorf("GetLastClimbedForLocation() error = %v, want nil", err)
	}

	if result != nil {
		t.Errorf("GetLastClimbedForLocation() returned %v, want nil for no climbs", result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetClimbHistoryForLocation(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"mp_route_id", "route_name", "route_rating", "mp_area_id", "area_name",
		"climbed_at", "climbed_by", "style", "comment", "days_since_climb",
	}).AddRow(
		int64(1001), "Monkey Face", "5.13a", int64(123), "Smith Rock",
		time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC),
		sql.NullString{String: "john_doe", Valid: true},
		sql.NullString{String: "Lead", Valid: true},
		sql.NullString{String: "Great!", Valid: true},
		5,
	).AddRow(
		int64(1002), "Morning Glory", "5.11b", int64(123), "Smith Rock",
		time.Date(2024, 6, 10, 14, 0, 0, 0, time.UTC),
		sql.NullString{String: "jane_doe", Valid: true},
		sql.NullString{String: "TR", Valid: true},
		sql.NullString{},
		10,
	)

	mock.ExpectQuery(`WITH adjusted_ticks AS`).
		WithArgs(10, 20).
		WillReturnRows(rows)

	repo := climbing.NewPostgresRepository(db)
	result, err := repo.History().GetClimbHistoryForLocation(context.Background(), 10, 20)

	if err != nil {
		t.Errorf("GetClimbHistoryForLocation() error = %v", err)
	}

	if len(result) != 2 {
		t.Errorf("GetClimbHistoryForLocation() returned %d entries, want 2", len(result))
	}

	if result[0].RouteName != "Monkey Face" {
		t.Errorf("GetClimbHistoryForLocation() first route = %v, want Monkey Face", result[0].RouteName)
	}

	if result[0].Comment == nil || *result[0].Comment != "Great!" {
		t.Errorf("GetClimbHistoryForLocation() first comment incorrect")
	}

	if result[1].Comment != nil {
		t.Errorf("GetClimbHistoryForLocation() second comment should be nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// ====================
// Activity Tests
// ====================

func TestPostgresRepository_GetAreasOrderedByActivity(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"mp_area_id", "name", "parent_mp_area_id", "last_climb_at",
		"unique_routes", "total_ticks", "days_since_climb", "has_subareas", "subarea_count",
	}).AddRow(
		int64(200), "Dihedrals", sql.NullInt64{Int64: 100, Valid: true},
		time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC),
		25, 150, 5, true, 3,
	).AddRow(
		int64(201), "Christian Brothers", sql.NullInt64{Int64: 100, Valid: true},
		time.Date(2024, 6, 10, 14, 0, 0, 0, time.UTC),
		18, 95, 10, false, 0,
	)

	mock.ExpectQuery(`WITH RECURSIVE adjusted_ticks AS`).
		WithArgs(10).
		WillReturnRows(rows)

	repo := climbing.NewPostgresRepository(db)
	result, err := repo.Activity().GetAreasOrderedByActivity(context.Background(), 10)

	if err != nil {
		t.Errorf("GetAreasOrderedByActivity() error = %v", err)
	}

	if len(result) != 2 {
		t.Errorf("GetAreasOrderedByActivity() returned %d areas, want 2", len(result))
	}

	if result[0].Name != "Dihedrals" {
		t.Errorf("GetAreasOrderedByActivity() first area = %v, want Dihedrals", result[0].Name)
	}

	if result[0].TotalTicks != 150 {
		t.Errorf("GetAreasOrderedByActivity() total ticks = %v, want 150", result[0].TotalTicks)
	}

	if !result[0].HasSubareas {
		t.Error("GetAreasOrderedByActivity() first area should have subareas")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetSubareasOrderedByActivity(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"mp_area_id", "name", "parent_mp_area_id", "last_climb_at",
		"unique_routes", "total_ticks", "days_since_climb", "has_subareas", "subarea_count",
	}).AddRow(
		int64(300), "Morning Glory Wall", sql.NullInt64{Int64: 200, Valid: true},
		time.Date(2024, 6, 12, 9, 0, 0, 0, time.UTC),
		12, 48, 8, false, 0,
	)

	mock.ExpectQuery(`WITH RECURSIVE adjusted_ticks AS`).
		WithArgs(int64(200), 10).
		WillReturnRows(rows)

	repo := climbing.NewPostgresRepository(db)
	result, err := repo.Activity().GetSubareasOrderedByActivity(context.Background(), int64(200), 10)

	if err != nil {
		t.Errorf("GetSubareasOrderedByActivity() error = %v", err)
	}

	if len(result) != 1 {
		t.Errorf("GetSubareasOrderedByActivity() returned %d subareas, want 1", len(result))
	}

	if result[0].Name != "Morning Glory Wall" {
		t.Errorf("GetSubareasOrderedByActivity() subarea name = %v, want Morning Glory Wall", result[0].Name)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetRoutesOrderedByActivity(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"mp_route_id", "name", "rating", "mp_area_id", "last_climb_at",
		"days_since_climb", "user_name", "adjusted_climbed_at", "style", "comment", "area_name", "no_ticks",
	}).AddRow(
		int64(1001), "Monkey Face", "5.13a", int64(200),
		time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC), 5,
		sql.NullString{String: "john_doe", Valid: true},
		sql.NullTime{Time: time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC), Valid: true},
		sql.NullString{String: "Lead", Valid: true},
		sql.NullString{String: "Amazing!", Valid: true},
		sql.NullString{String: "Dihedrals", Valid: true},
		0,
	)

	mock.ExpectQuery(`WITH area_routes AS`).
		WithArgs(int64(200), 10, 50).
		WillReturnRows(rows)

	repo := climbing.NewPostgresRepository(db)
	result, err := repo.Activity().GetRoutesOrderedByActivity(context.Background(), int64(200), 10, 50)

	if err != nil {
		t.Errorf("GetRoutesOrderedByActivity() error = %v", err)
	}

	if len(result) != 1 {
		t.Errorf("GetRoutesOrderedByActivity() returned %d routes, want 1", len(result))
	}

	if result[0].Name != "Monkey Face" {
		t.Errorf("GetRoutesOrderedByActivity() route name = %v, want Monkey Face", result[0].Name)
	}

	if result[0].MostRecentTick == nil {
		t.Fatal("GetRoutesOrderedByActivity() most recent tick should not be nil")
	}

	if result[0].MostRecentTick.ClimbedBy != "john_doe" {
		t.Errorf("GetRoutesOrderedByActivity() climbed by = %v, want john_doe", result[0].MostRecentTick.ClimbedBy)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetRecentTicksForRoute(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"mp_route_id", "route_name", "route_rating", "mp_area_id", "area_name",
		"climbed_at", "climbed_by", "style", "comment", "days_since_climb",
	}).AddRow(
		int64(1001), "Monkey Face", "5.13a", int64(200), "Dihedrals",
		time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC),
		sql.NullString{String: "john_doe", Valid: true},
		sql.NullString{String: "Lead", Valid: true},
		sql.NullString{String: "Great!", Valid: true},
		5,
	).AddRow(
		int64(1001), "Monkey Face", "5.13a", int64(200), "Dihedrals",
		time.Date(2024, 6, 10, 14, 0, 0, 0, time.UTC),
		sql.NullString{String: "jane_doe", Valid: true},
		sql.NullString{String: "Follow", Valid: true},
		sql.NullString{},
		10,
	)

	mock.ExpectQuery(`WITH adjusted_ticks AS`).
		WithArgs(int64(1001), 10).
		WillReturnRows(rows)

	repo := climbing.NewPostgresRepository(db)
	result, err := repo.Activity().GetRecentTicksForRoute(context.Background(), int64(1001), 10)

	if err != nil {
		t.Errorf("GetRecentTicksForRoute() error = %v", err)
	}

	if len(result) != 2 {
		t.Errorf("GetRecentTicksForRoute() returned %d ticks, want 2", len(result))
	}

	if result[0].ClimbedBy != "john_doe" {
		t.Errorf("GetRecentTicksForRoute() first climber = %v, want john_doe", result[0].ClimbedBy)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// ====================
// Search Tests
// ====================

func TestPostgresRepository_SearchInLocation(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"result_type", "id", "name", "rating", "mp_area_id", "area_name",
		"last_climb_at", "days_since_climb", "total_ticks", "unique_routes",
		"user_name", "tick_climbed_at", "style", "comment", "no_ticks",
	}).AddRow(
		"area", int64(200), "Dihedrals", sql.NullString{},
		int64(200), sql.NullString{},
		time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC), 5,
		sql.NullInt64{Int64: 150, Valid: true}, sql.NullInt64{Int64: 25, Valid: true},
		sql.NullString{}, sql.NullTime{},
		sql.NullString{}, sql.NullString{},
		0,
	).AddRow(
		"route", int64(1001), "Dihedrals Route", sql.NullString{String: "5.10a", Valid: true},
		int64(200), sql.NullString{String: "Dihedrals", Valid: true},
		time.Date(2024, 6, 12, 9, 0, 0, 0, time.UTC), 8,
		sql.NullInt64{}, sql.NullInt64{},
		sql.NullString{String: "bob", Valid: true},
		sql.NullTime{Time: time.Date(2024, 6, 12, 9, 0, 0, 0, time.UTC), Valid: true},
		sql.NullString{String: "Lead", Valid: true},
		sql.NullString{},
		0,
	)

	mock.ExpectQuery(`WITH location_routes AS`).
		WithArgs(10, "%dihedral%", 20).
		WillReturnRows(rows)

	repo := climbing.NewPostgresRepository(db)
	result, err := repo.Search().SearchInLocation(context.Background(), 10, "dihedral", 20)

	if err != nil {
		t.Errorf("SearchInLocation() error = %v", err)
	}

	if len(result) != 2 {
		t.Errorf("SearchInLocation() returned %d results, want 2", len(result))
	}

	if result[0].ResultType != "area" {
		t.Errorf("SearchInLocation() first type = %v, want area", result[0].ResultType)
	}

	if result[1].ResultType != "route" {
		t.Errorf("SearchInLocation() second type = %v, want route", result[1].ResultType)
	}

	if result[1].MostRecentTick == nil {
		t.Error("SearchInLocation() route should have most recent tick")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_SearchRoutesInLocation(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"mp_route_id", "name", "rating", "mp_area_id", "last_climb_at",
		"days_since_climb", "user_name", "adjusted_climbed_at", "style", "comment", "area_name", "no_ticks",
	}).AddRow(
		int64(1001), "Morning Glory", "5.11b", int64(200),
		time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC), 5,
		sql.NullString{String: "alice", Valid: true},
		sql.NullTime{Time: time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC), Valid: true},
		sql.NullString{String: "Lead", Valid: true},
		sql.NullString{},
		sql.NullString{String: "Dihedrals", Valid: true},
		0,
	)

	mock.ExpectQuery(`WITH location_routes AS`).
		WithArgs(10, "%glory%", 25).
		WillReturnRows(rows)

	repo := climbing.NewPostgresRepository(db)
	result, err := repo.Search().SearchRoutesInLocation(context.Background(), 10, "glory", 25)

	if err != nil {
		t.Errorf("SearchRoutesInLocation() error = %v", err)
	}

	if len(result) != 1 {
		t.Errorf("SearchRoutesInLocation() returned %d routes, want 1", len(result))
	}

	if result[0].Name != "Morning Glory" {
		t.Errorf("SearchRoutesInLocation() route name = %v, want Morning Glory", result[0].Name)
	}

	if result[0].MostRecentTick == nil {
		t.Fatal("SearchRoutesInLocation() most recent tick should not be nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_SearchRoutesInLocation_NoResults(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"mp_route_id", "name", "rating", "mp_area_id", "last_climb_at",
		"days_since_climb", "user_name", "adjusted_climbed_at", "style", "comment", "area_name", "no_ticks",
	})

	mock.ExpectQuery(`WITH location_routes AS`).
		WithArgs(10, "%nonexistent%", 25).
		WillReturnRows(rows)

	repo := climbing.NewPostgresRepository(db)
	result, err := repo.Search().SearchRoutesInLocation(context.Background(), 10, "nonexistent", 25)

	if err != nil {
		t.Errorf("SearchRoutesInLocation() error = %v, want nil", err)
	}

	if len(result) != 0 {
		t.Errorf("SearchRoutesInLocation() returned %d routes, want 0", len(result))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
