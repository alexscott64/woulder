package heatmap_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alexscott64/woulder/backend/internal/database/heatmap"
)

func TestPostgresRepository_GetHeatMapData_Lightweight(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	rows := sqlmock.NewRows([]string{
		"mp_area_id", "name", "latitude", "longitude",
		"active_routes", "total_ticks", "last_activity",
		"unique_climbers", "has_subareas",
	}).AddRow(
		int64(123), "Smith Rock", 44.3672, -121.1408,
		0, 150, time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC),
		45, false,
	).AddRow(
		int64(456), "Red Rock Canyon", 36.1357, -115.4274,
		0, 200, time.Date(2024, 7, 20, 14, 15, 0, 0, time.UTC),
		60, false,
	)

	// Use regex that matches the query structure regardless of comments/whitespace
	mock.ExpectQuery(`SELECT\s+a\.mp_area_id`).
		WithArgs(startDate, endDate, nil, nil, nil, nil, nil, 10, 100).
		WillReturnRows(rows)

	repo := heatmap.NewPostgresRepository(db)
	result, err := repo.GetHeatMapData(context.Background(), startDate, endDate, nil, 10, 100, nil, true)

	if err != nil {
		t.Errorf("GetHeatMapData() error = %v", err)
	}

	if len(result) != 2 {
		t.Errorf("GetHeatMapData() returned %d points, want 2", len(result))
	}

	if result[0].MPAreaID != 123 {
		t.Errorf("GetHeatMapData() first area ID = %v, want 123", result[0].MPAreaID)
	}

	if result[0].Name != "Smith Rock" {
		t.Errorf("GetHeatMapData() first area name = %v, want Smith Rock", result[0].Name)
	}

	if result[0].TotalTicks != 150 {
		t.Errorf("GetHeatMapData() first area ticks = %v, want 150", result[0].TotalTicks)
	}

	if result[1].UniqueClimbers != 60 {
		t.Errorf("GetHeatMapData() second area climbers = %v, want 60", result[1].UniqueClimbers)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetHeatMapData_Full(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	rows := sqlmock.NewRows([]string{
		"mp_area_id", "name", "latitude", "longitude",
		"active_routes", "total_ticks", "last_activity",
		"unique_climbers", "has_subareas",
	}).AddRow(
		int64(789), "Joshua Tree", 33.8734, -115.9010,
		85, 300, time.Date(2024, 8, 10, 9, 0, 0, 0, time.UTC),
		120, true,
	)

	// Use regex that matches the query structure
	mock.ExpectQuery(`SELECT\s+a\.mp_area_id`).
		WithArgs(startDate, endDate, nil, nil, nil, nil, nil, 5, 50).
		WillReturnRows(rows)

	repo := heatmap.NewPostgresRepository(db)
	result, err := repo.GetHeatMapData(context.Background(), startDate, endDate, nil, 5, 50, nil, false)

	if err != nil {
		t.Errorf("GetHeatMapData() error = %v", err)
	}

	if len(result) != 1 {
		t.Errorf("GetHeatMapData() returned %d points, want 1", len(result))
	}

	if result[0].ActiveRoutes != 85 {
		t.Errorf("GetHeatMapData() active routes = %v, want 85", result[0].ActiveRoutes)
	}

	if !result[0].HasSubareas {
		t.Error("GetHeatMapData() has_subareas should be true")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetHeatMapData_WithBounds(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	bounds := &heatmap.GeoBounds{
		MinLat: 36.0,
		MaxLat: 37.0,
		MinLon: -116.0,
		MaxLon: -115.0,
	}

	rows := sqlmock.NewRows([]string{
		"mp_area_id", "name", "latitude", "longitude",
		"active_routes", "total_ticks", "last_activity",
		"unique_climbers", "has_subareas",
	}).AddRow(
		int64(999), "Test Area", 36.5, -115.5,
		0, 50, time.Date(2024, 5, 1, 12, 0, 0, 0, time.UTC),
		20, false,
	)

	mock.ExpectQuery(`SELECT\s+a\.mp_area_id`).
		WithArgs(startDate, endDate, 36.0, 37.0, -116.0, -115.0, nil, 1, 100).
		WillReturnRows(rows)

	repo := heatmap.NewPostgresRepository(db)
	result, err := repo.GetHeatMapData(context.Background(), startDate, endDate, bounds, 1, 100, nil, true)

	if err != nil {
		t.Errorf("GetHeatMapData() error = %v", err)
	}

	if len(result) != 1 {
		t.Errorf("GetHeatMapData() returned %d points, want 1", len(result))
	}

	if result[0].Latitude != 36.5 {
		t.Errorf("GetHeatMapData() latitude = %v, want 36.5", result[0].Latitude)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetHeatMapData_InvalidBounds(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	// Invalid bounds: min > max
	bounds := &heatmap.GeoBounds{
		MinLat: 37.0,
		MaxLat: 36.0, // < MinLat (invalid)
		MinLon: -115.0,
		MaxLon: -116.0, // < MinLon (invalid)
	}

	repo := heatmap.NewPostgresRepository(db)
	_, err = repo.GetHeatMapData(context.Background(), startDate, endDate, bounds, 1, 100, nil, true)

	if err == nil {
		t.Error("GetHeatMapData() expected error for invalid bounds, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetAreaActivityDetail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	areaID := int64(123)
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	// Mock area info query
	areaRows := sqlmock.NewRows([]string{
		"mp_area_id", "name", "parent_mp_area_id", "latitude", "longitude",
	}).AddRow(
		areaID, "Smith Rock", sql.NullInt64{Int64: 100, Valid: true},
		sql.NullFloat64{Float64: 44.3672, Valid: true},
		sql.NullFloat64{Float64: -121.1408, Valid: true},
	)

	mock.ExpectQuery(`SELECT\s+a\.mp_area_id(.+)FROM woulder\.mp_areas a\s+WHERE`).
		WithArgs(areaID).
		WillReturnRows(areaRows)

	// Mock activity stats query
	statsRows := sqlmock.NewRows([]string{
		"total_ticks", "active_routes", "unique_climbers", "last_activity",
	}).AddRow(
		250, 45, 80, time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC),
	)

	mock.ExpectQuery(`SELECT\s+COUNT\(t\.id\)`).
		WithArgs(areaID, startDate, endDate).
		WillReturnRows(statsRows)

	// Mock recent ticks query
	ticksRows := sqlmock.NewRows([]string{
		"mp_route_id", "route_name", "rating", "user_name",
		"climbed_at", "style", "comment",
	}).AddRow(
		int64(1001), "Monkey Face", "5.13a", "john_doe",
		time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC), "Lead", "Great climb!",
	)

	mock.ExpectQuery(`SELECT\s+t\.mp_route_id(.+)ORDER BY t\.climbed_at DESC`).
		WithArgs(areaID, startDate, endDate).
		WillReturnRows(ticksRows)

	// Mock recent comments query
	commentsRows := sqlmock.NewRows([]string{
		"id", "user_name", "comment_text", "commented_at", "mp_route_id", "route_name",
	}).AddRow(
		int64(5001), "jane_doe", "Watch out for loose rock",
		time.Date(2024, 6, 10, 15, 0, 0, 0, time.UTC),
		sql.NullInt64{Int64: 1001, Valid: true}, sql.NullString{String: "Monkey Face", Valid: true},
	)

	mock.ExpectQuery(`SELECT\s+c\.id(.+)REGEXP_REPLACE`).
		WithArgs(areaID, startDate, endDate).
		WillReturnRows(commentsRows)

	// Mock activity timeline query
	timelineRows := sqlmock.NewRows([]string{
		"date", "tick_count", "route_count",
	}).AddRow(
		time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC), 10, 5,
	)

	mock.ExpectQuery(`SELECT\s+DATE\(t\.climbed_at\)(.+)GROUP BY DATE`).
		WithArgs(areaID, startDate, endDate).
		WillReturnRows(timelineRows)

	// Mock top routes query
	topRoutesRows := sqlmock.NewRows([]string{
		"mp_route_id", "name", "rating", "tick_count", "last_activity",
	}).AddRow(
		int64(1001), "Monkey Face", "5.13a", 35,
		time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC),
	)

	mock.ExpectQuery(`SELECT\s+r\.mp_route_id(.+)ORDER BY COUNT`).
		WithArgs(areaID, startDate, endDate).
		WillReturnRows(topRoutesRows)

	repo := heatmap.NewPostgresRepository(db)
	result, err := repo.GetAreaActivityDetail(context.Background(), areaID, startDate, endDate)

	if err != nil {
		t.Errorf("GetAreaActivityDetail() error = %v", err)
	}

	if result.MPAreaID != areaID {
		t.Errorf("GetAreaActivityDetail() area ID = %v, want %v", result.MPAreaID, areaID)
	}

	if result.Name != "Smith Rock" {
		t.Errorf("GetAreaActivityDetail() name = %v, want Smith Rock", result.Name)
	}

	if result.TotalTicks != 250 {
		t.Errorf("GetAreaActivityDetail() total ticks = %v, want 250", result.TotalTicks)
	}

	if result.ActiveRoutes != 45 {
		t.Errorf("GetAreaActivityDetail() active routes = %v, want 45", result.ActiveRoutes)
	}

	if len(result.RecentTicks) != 1 {
		t.Errorf("GetAreaActivityDetail() recent ticks = %d, want 1", len(result.RecentTicks))
	}

	if len(result.RecentComments) != 1 {
		t.Errorf("GetAreaActivityDetail() recent comments = %d, want 1", len(result.RecentComments))
	}

	if len(result.ActivityTimeline) != 1 {
		t.Errorf("GetAreaActivityDetail() timeline = %d, want 1", len(result.ActivityTimeline))
	}

	if len(result.TopRoutes) != 1 {
		t.Errorf("GetAreaActivityDetail() top routes = %d, want 1", len(result.TopRoutes))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetAreaActivityDetail_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	areaID := int64(999)
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	// Area not found
	areaRows := sqlmock.NewRows([]string{
		"mp_area_id", "name", "parent_mp_area_id", "latitude", "longitude",
	})

	mock.ExpectQuery(`SELECT\s+a\.mp_area_id(.+)FROM woulder\.mp_areas a\s+WHERE`).
		WithArgs(areaID).
		WillReturnRows(areaRows)

	repo := heatmap.NewPostgresRepository(db)
	_, err = repo.GetAreaActivityDetail(context.Background(), areaID, startDate, endDate)

	if err == nil {
		t.Error("GetAreaActivityDetail() expected error for non-existent area, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetRoutesByBounds(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	bounds := heatmap.GeoBounds{
		MinLat: 44.0,
		MaxLat: 45.0,
		MinLon: -122.0,
		MaxLon: -121.0,
	}
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	rows := sqlmock.NewRows([]string{
		"mp_route_id", "name", "rating", "latitude", "longitude",
		"tick_count", "last_activity", "mp_area_id", "area_name",
	}).AddRow(
		int64(2001), "Morning Glory", "5.11b",
		sql.NullFloat64{Float64: 44.5, Valid: true},
		sql.NullFloat64{Float64: -121.5, Valid: true},
		25, time.Date(2024, 6, 1, 8, 0, 0, 0, time.UTC),
		int64(123), "Smith Rock",
	)

	mock.ExpectQuery(`SELECT\s+r\.mp_route_id(.+)r\.latitude BETWEEN`).
		WithArgs(44.0, 45.0, -122.0, -121.0, startDate, endDate, 50).
		WillReturnRows(rows)

	repo := heatmap.NewPostgresRepository(db)
	result, err := repo.GetRoutesByBounds(context.Background(), bounds, startDate, endDate, 50)

	if err != nil {
		t.Errorf("GetRoutesByBounds() error = %v", err)
	}

	if len(result) != 1 {
		t.Errorf("GetRoutesByBounds() returned %d routes, want 1", len(result))
	}

	if result[0].Name != "Morning Glory" {
		t.Errorf("GetRoutesByBounds() route name = %v, want Morning Glory", result[0].Name)
	}

	if result[0].TickCount != 25 {
		t.Errorf("GetRoutesByBounds() tick count = %v, want 25", result[0].TickCount)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetRouteTicksInDateRange(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	routeID := int64(2001)
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	rows := sqlmock.NewRows([]string{
		"mp_route_id", "route_name", "rating", "user_name",
		"climbed_at", "style", "comment",
	}).AddRow(
		routeID, "Morning Glory", "5.11b", "alice",
		time.Date(2024, 6, 1, 8, 0, 0, 0, time.UTC), "Lead", "Awesome!",
	).AddRow(
		routeID, "Morning Glory", "5.11b", "bob",
		time.Date(2024, 5, 15, 10, 30, 0, 0, time.UTC), "TR", "",
	)

	mock.ExpectQuery(`SELECT\s+t\.mp_route_id(.+)WHERE t\.mp_route_id`).
		WithArgs(routeID, startDate, endDate, 20).
		WillReturnRows(rows)

	repo := heatmap.NewPostgresRepository(db)
	result, err := repo.GetRouteTicksInDateRange(context.Background(), routeID, startDate, endDate, 20)

	if err != nil {
		t.Errorf("GetRouteTicksInDateRange() error = %v", err)
	}

	if len(result) != 2 {
		t.Errorf("GetRouteTicksInDateRange() returned %d ticks, want 2", len(result))
	}

	if result[0].UserName != "alice" {
		t.Errorf("GetRouteTicksInDateRange() first user = %v, want alice", result[0].UserName)
	}

	if result[1].Style != "TR" {
		t.Errorf("GetRouteTicksInDateRange() second style = %v, want TR", result[1].Style)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_SearchRoutesInAreas(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	areaIDs := []int64{123, 456}
	searchQuery := "glory"
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	rows := sqlmock.NewRows([]string{
		"mp_route_id", "name", "rating", "latitude", "longitude",
		"tick_count", "last_activity", "mp_area_id", "area_name",
	}).AddRow(
		int64(2001), "Morning Glory", "5.11b",
		sql.NullFloat64{Float64: 44.5, Valid: true},
		sql.NullFloat64{Float64: -121.5, Valid: true},
		30, time.Date(2024, 6, 1, 8, 0, 0, 0, time.UTC),
		int64(123), "Smith Rock",
	).AddRow(
		int64(2002), "Glory Days", "5.10c",
		sql.NullFloat64{Float64: 36.2, Valid: true},
		sql.NullFloat64{Float64: -115.4, Valid: true},
		15, time.Date(2024, 5, 20, 12, 0, 0, 0, time.UTC),
		int64(456), "Red Rock",
	)

	mock.ExpectQuery(`SELECT\s+r\.mp_route_id(.+)LOWER\(r\.name\) LIKE LOWER`).
		WithArgs(sqlmock.AnyArg(), startDate, endDate, "%glory%", 25).
		WillReturnRows(rows)

	repo := heatmap.NewPostgresRepository(db)
	result, err := repo.SearchRoutesInAreas(context.Background(), areaIDs, searchQuery, startDate, endDate, 25)

	if err != nil {
		t.Errorf("SearchRoutesInAreas() error = %v", err)
	}

	if len(result) != 2 {
		t.Errorf("SearchRoutesInAreas() returned %d routes, want 2", len(result))
	}

	if result[0].Name != "Morning Glory" {
		t.Errorf("SearchRoutesInAreas() first route = %v, want Morning Glory", result[0].Name)
	}

	if result[1].Name != "Glory Days" {
		t.Errorf("SearchRoutesInAreas() second route = %v, want Glory Days", result[1].Name)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_SearchRoutesInAreas_NoResults(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	areaIDs := []int64{123}
	searchQuery := "nonexistent"
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	rows := sqlmock.NewRows([]string{
		"mp_route_id", "name", "rating", "latitude", "longitude",
		"tick_count", "last_activity", "mp_area_id", "area_name",
	})

	mock.ExpectQuery(`SELECT\s+r\.mp_route_id`).
		WithArgs(sqlmock.AnyArg(), startDate, endDate, "%nonexistent%", 25).
		WillReturnRows(rows)

	repo := heatmap.NewPostgresRepository(db)
	result, err := repo.SearchRoutesInAreas(context.Background(), areaIDs, searchQuery, startDate, endDate, 25)

	if err != nil {
		t.Errorf("SearchRoutesInAreas() error = %v, want nil", err)
	}

	if len(result) != 0 {
		t.Errorf("SearchRoutesInAreas() returned %d routes, want 0", len(result))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
