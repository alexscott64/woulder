package mountainproject_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alexscott64/woulder/backend/internal/database/mountainproject"
	"github.com/alexscott64/woulder/backend/internal/models"
)

// AreasRepository Tests

func TestPostgresRepository_SaveArea(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	parentID := int64(100)
	locID := 1
	lat := 44.3672
	lon := -121.1408
	area := &models.MPArea{
		MPAreaID:       123,
		Name:           "Smith Rock",
		ParentMPAreaID: &parentID,
		AreaType:       "Crag",
		LocationID:     &locID,
		Latitude:       &lat,
		Longitude:      &lon,
	}

	mock.ExpectExec(`INSERT INTO woulder\.mp_areas`).
		WithArgs(area.MPAreaID, area.Name, area.ParentMPAreaID, area.AreaType,
			area.LocationID, area.Latitude, area.Longitude, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := mountainproject.NewPostgresRepository(db)
	err = repo.Areas().SaveArea(context.Background(), area)

	if err != nil {
		t.Errorf("SaveArea() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetAreaByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	mpAreaID := int64(123)
	parentID := int64(100)
	locID := 1
	lat := 44.3672
	lon := -121.1408
	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "mp_area_id", "name", "parent_mp_area_id", "area_type",
		"location_id", "latitude", "longitude", "last_synced_at", "created_at", "updated_at",
	}).AddRow(
		1, mpAreaID, "Smith Rock", &parentID, "Crag",
		&locID, &lat, &lon, &now, now, now,
	)

	mock.ExpectQuery(`SELECT id, mp_area_id, name`).
		WithArgs(mpAreaID).
		WillReturnRows(rows)

	repo := mountainproject.NewPostgresRepository(db)
	result, err := repo.Areas().GetByID(context.Background(), mpAreaID)

	if err != nil {
		t.Errorf("GetByID() error = %v", err)
	}

	if result == nil {
		t.Fatal("GetByID() returned nil")
	}

	if result.MPAreaID != mpAreaID {
		t.Errorf("GetByID() area ID = %v, want %v", result.MPAreaID, mpAreaID)
	}

	if result.Name != "Smith Rock" {
		t.Errorf("GetByID() name = %v, want Smith Rock", result.Name)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetAreaByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`SELECT id, mp_area_id, name`).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	repo := mountainproject.NewPostgresRepository(db)
	result, err := repo.Areas().GetByID(context.Background(), 999)

	if err != nil {
		t.Errorf("GetByID() error = %v, want nil", err)
	}

	if result != nil {
		t.Errorf("GetByID() = %v, want nil for not found", result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_UpdateRouteCount(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(`UPDATE woulder\.mp_areas`).
		WithArgs(150, "123").
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := mountainproject.NewPostgresRepository(db)
	err = repo.Areas().UpdateRouteCount(context.Background(), "123", 150)

	if err != nil {
		t.Errorf("UpdateRouteCount() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetRouteCount(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"route_count_total"}).AddRow(150)

	mock.ExpectQuery(`SELECT route_count_total`).
		WithArgs("123").
		WillReturnRows(rows)

	repo := mountainproject.NewPostgresRepository(db)
	count, err := repo.Areas().GetRouteCount(context.Background(), "123")

	if err != nil {
		t.Errorf("GetRouteCount() error = %v", err)
	}

	if count != 150 {
		t.Errorf("GetRouteCount() = %v, want 150", count)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetChildAreas(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"mp_area_id", "name"}).
		AddRow("124", "Morning Glory Wall").
		AddRow("125", "Monkey Face")

	mock.ExpectQuery(`SELECT mp_area_id, name`).
		WithArgs("123").
		WillReturnRows(rows)

	repo := mountainproject.NewPostgresRepository(db)
	areas, err := repo.Areas().GetChildAreas(context.Background(), "123")

	if err != nil {
		t.Errorf("GetChildAreas() error = %v", err)
	}

	if len(areas) != 2 {
		t.Errorf("GetChildAreas() returned %d areas, want 2", len(areas))
	}

	if areas[0].Name != "Morning Glory Wall" {
		t.Errorf("GetChildAreas() first area name = %v, want Morning Glory Wall", areas[0].Name)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetAllStateConfigs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"state_name", "mp_area_id", "is_active"}).
		AddRow("Oregon", "105708959", true).
		AddRow("Washington", "105708964", true)

	mock.ExpectQuery(`SELECT state_name, mp_area_id, is_active`).
		WillReturnRows(rows)

	repo := mountainproject.NewPostgresRepository(db)
	configs, err := repo.Areas().GetAllStateConfigs(context.Background())

	if err != nil {
		t.Errorf("GetAllStateConfigs() error = %v", err)
	}

	if len(configs) != 2 {
		t.Errorf("GetAllStateConfigs() returned %d configs, want 2", len(configs))
	}

	if configs[0].StateName != "Oregon" {
		t.Errorf("GetAllStateConfigs() first state = %v, want Oregon", configs[0].StateName)
	}

	if !configs[0].IsActive {
		t.Errorf("GetAllStateConfigs() first state IsActive = false, want true")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// RoutesRepository Tests

func TestPostgresRepository_SaveRoute(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	locID := 1
	lat := 44.3672
	lon := -121.1408
	aspect := "SE"
	route := &models.MPRoute{
		MPRouteID:  456,
		MPAreaID:   123,
		Name:       "Just Do It",
		RouteType:  "Sport",
		Rating:     "5.14c",
		LocationID: &locID,
		Latitude:   &lat,
		Longitude:  &lon,
		Aspect:     &aspect,
	}

	mock.ExpectExec(`INSERT INTO woulder\.mp_routes`).
		WithArgs(route.MPRouteID, route.MPAreaID, route.Name, route.RouteType,
			route.Rating, route.LocationID, route.Latitude, route.Longitude, route.Aspect).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := mountainproject.NewPostgresRepository(db)
	err = repo.Routes().SaveRoute(context.Background(), route)

	if err != nil {
		t.Errorf("SaveRoute() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetAllIDsForLocation(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"mp_route_id"}).
		AddRow(int64(456)).
		AddRow(int64(789))

	mock.ExpectQuery(`SELECT mp_route_id`).
		WithArgs(1).
		WillReturnRows(rows)

	repo := mountainproject.NewPostgresRepository(db)
	routeIDs, err := repo.Routes().GetAllIDsForLocation(context.Background(), 1)

	if err != nil {
		t.Errorf("GetAllIDsForLocation() error = %v", err)
	}

	if len(routeIDs) != 2 {
		t.Errorf("GetAllIDsForLocation() returned %d IDs, want 2", len(routeIDs))
	}

	if routeIDs[0] != 456 {
		t.Errorf("GetAllIDsForLocation() first ID = %v, want 456", routeIDs[0])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_UpdateGPS(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(`UPDATE woulder\.mp_routes`).
		WithArgs(44.3672, -121.1408, "SE", int64(456)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := mountainproject.NewPostgresRepository(db)
	err = repo.Routes().UpdateGPS(context.Background(), 456, 44.3672, -121.1408, "SE")

	if err != nil {
		t.Errorf("UpdateGPS() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetIDsForArea(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"mp_route_id"}).
		AddRow("456").
		AddRow("789")

	mock.ExpectQuery(`SELECT mp_route_id::text`).
		WithArgs("123").
		WillReturnRows(rows)

	repo := mountainproject.NewPostgresRepository(db)
	routeIDs, err := repo.Routes().GetIDsForArea(context.Background(), "123")

	if err != nil {
		t.Errorf("GetIDsForArea() error = %v", err)
	}

	if len(routeIDs) != 2 {
		t.Errorf("GetIDsForArea() returned %d IDs, want 2", len(routeIDs))
	}

	if routeIDs[0] != "456" {
		t.Errorf("GetIDsForArea() first ID = %v, want 456", routeIDs[0])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_UpsertRoute(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	locationID := 1
	lat := 44.3672
	lon := -121.1408
	aspect := "SE"

	mock.ExpectExec(`INSERT INTO woulder\.mp_routes`).
		WithArgs(int64(456), int64(123), &locationID, "Just Do It", "Sport", "5.14c", &lat, &lon, &aspect).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := mountainproject.NewPostgresRepository(db)
	err = repo.Routes().UpsertRoute(context.Background(), 456, 123, &locationID, "Just Do It", "Sport", "5.14c", &lat, &lon, &aspect)

	if err != nil {
		t.Errorf("UpsertRoute() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TicksRepository Tests

func TestPostgresRepository_SaveTick(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	climbedAt := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	comment := "Great climb!"
	tick := &models.MPTick{
		MPRouteID: 456,
		UserName:  "climber123",
		ClimbedAt: climbedAt,
		Style:     "Redpoint",
		Comment:   &comment,
	}

	mock.ExpectExec(`INSERT INTO woulder\.mp_ticks`).
		WithArgs(tick.MPRouteID, tick.UserName, tick.ClimbedAt, tick.Style, tick.Comment).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(`UPDATE woulder\.mp_routes`).
		WithArgs(tick.MPRouteID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := mountainproject.NewPostgresRepository(db)
	err = repo.Ticks().SaveTick(context.Background(), tick)

	if err != nil {
		t.Errorf("SaveTick() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetLastTimestampForRoute(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	lastTick := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{"last_tick"}).AddRow(lastTick)

	mock.ExpectQuery(`SELECT MAX\(climbed_at\)`).
		WithArgs(int64(456)).
		WillReturnRows(rows)

	repo := mountainproject.NewPostgresRepository(db)
	result, err := repo.Ticks().GetLastTimestampForRoute(context.Background(), 456)

	if err != nil {
		t.Errorf("GetLastTimestampForRoute() error = %v", err)
	}

	if result == nil {
		t.Fatal("GetLastTimestampForRoute() returned nil")
	}

	if !result.Equal(lastTick) {
		t.Errorf("GetLastTimestampForRoute() = %v, want %v", result, lastTick)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetLastTimestampForRoute_NoTicks(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`SELECT MAX\(climbed_at\)`).
		WithArgs(int64(456)).
		WillReturnError(sql.ErrNoRows)

	repo := mountainproject.NewPostgresRepository(db)
	result, err := repo.Ticks().GetLastTimestampForRoute(context.Background(), 456)

	if err != nil {
		t.Errorf("GetLastTimestampForRoute() error = %v, want nil", err)
	}

	if result != nil {
		t.Errorf("GetLastTimestampForRoute() = %v, want nil for no ticks", result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// CommentsRepository Tests

func TestPostgresRepository_SaveAreaComment(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	commentedAt := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	mock.ExpectExec(`INSERT INTO woulder\.mp_comments`).
		WithArgs(int64(789), int64(123), "commenter123", "Great area!", commentedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := mountainproject.NewPostgresRepository(db)
	err = repo.Comments().SaveAreaComment(context.Background(), 789, 123, "commenter123", "Great area!", commentedAt)

	if err != nil {
		t.Errorf("SaveAreaComment() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_SaveRouteComment(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	commentedAt := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	mock.ExpectExec(`INSERT INTO woulder\.mp_comments`).
		WithArgs(int64(789), int64(456), "commenter123", "Great route!", commentedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(`UPDATE woulder\.mp_routes`).
		WithArgs(int64(456)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := mountainproject.NewPostgresRepository(db)
	err = repo.Comments().SaveRouteComment(context.Background(), 789, 456, "commenter123", "Great route!", commentedAt)

	if err != nil {
		t.Errorf("SaveRouteComment() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_UpsertAreaComment(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	commentedAt := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	userID := "user123"

	mock.ExpectExec(`INSERT INTO woulder\.mp_comments`).
		WithArgs(int64(789), int64(123), "commenter123", &userID, "Great area!", commentedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := mountainproject.NewPostgresRepository(db)
	err = repo.Comments().UpsertAreaComment(context.Background(), 789, 123, "commenter123", &userID, "Great area!", commentedAt)

	if err != nil {
		t.Errorf("UpsertAreaComment() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_UpsertRouteComment(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	commentedAt := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	userID := "user123"

	mock.ExpectExec(`INSERT INTO woulder\.mp_comments`).
		WithArgs(int64(789), int64(456), "commenter123", &userID, "Great route!", commentedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := mountainproject.NewPostgresRepository(db)
	err = repo.Comments().UpsertRouteComment(context.Background(), 789, 456, "commenter123", &userID, "Great route!", commentedAt)

	if err != nil {
		t.Errorf("UpsertRouteComment() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// SyncRepository Tests

func TestPostgresRepository_UpdateRoutePriorities(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(`WITH route_metrics AS`).
		WillReturnResult(sqlmock.NewResult(0, 100))

	repo := mountainproject.NewPostgresRepository(db)
	err = repo.Sync().UpdateRoutePriorities(context.Background())

	if err != nil {
		t.Errorf("UpdateRoutePriorities() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetLocationRoutesDueForSync_Ticks(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"mp_route_id"}).
		AddRow(int64(456)).
		AddRow(int64(789))

	mock.ExpectQuery(`SELECT mp_route_id`).
		WillReturnRows(rows)

	repo := mountainproject.NewPostgresRepository(db)
	routeIDs, err := repo.Sync().GetLocationRoutesDueForSync(context.Background(), "ticks")

	if err != nil {
		t.Errorf("GetLocationRoutesDueForSync() error = %v", err)
	}

	if len(routeIDs) != 2 {
		t.Errorf("GetLocationRoutesDueForSync() returned %d IDs, want 2", len(routeIDs))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetLocationRoutesDueForSync_Comments(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"mp_route_id"}).
		AddRow(int64(456))

	mock.ExpectQuery(`SELECT mp_route_id`).
		WillReturnRows(rows)

	repo := mountainproject.NewPostgresRepository(db)
	routeIDs, err := repo.Sync().GetLocationRoutesDueForSync(context.Background(), "comments")

	if err != nil {
		t.Errorf("GetLocationRoutesDueForSync() error = %v", err)
	}

	if len(routeIDs) != 1 {
		t.Errorf("GetLocationRoutesDueForSync() returned %d IDs, want 1", len(routeIDs))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetLocationRoutesDueForSync_InvalidType(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := mountainproject.NewPostgresRepository(db)
	_, err = repo.Sync().GetLocationRoutesDueForSync(context.Background(), "invalid")

	if err == nil {
		t.Errorf("GetLocationRoutesDueForSync() expected error for invalid type, got nil")
	}
}

func TestPostgresRepository_GetRoutesDueForTickSync_HighPriority(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"mp_route_id"}).
		AddRow(int64(456)).
		AddRow(int64(789))

	mock.ExpectQuery(`SELECT mp_route_id`).
		WithArgs("high").
		WillReturnRows(rows)

	repo := mountainproject.NewPostgresRepository(db)
	routeIDs, err := repo.Sync().GetRoutesDueForTickSync(context.Background(), "high")

	if err != nil {
		t.Errorf("GetRoutesDueForTickSync() error = %v", err)
	}

	if len(routeIDs) != 2 {
		t.Errorf("GetRoutesDueForTickSync() returned %d IDs, want 2", len(routeIDs))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetRoutesDueForTickSync_InvalidPriority(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := mountainproject.NewPostgresRepository(db)
	_, err = repo.Sync().GetRoutesDueForTickSync(context.Background(), "invalid")

	if err == nil {
		t.Errorf("GetRoutesDueForTickSync() expected error for invalid priority, got nil")
	}
}

func TestPostgresRepository_GetRoutesDueForCommentSync_MediumPriority(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"mp_route_id"}).
		AddRow(int64(456))

	mock.ExpectQuery(`SELECT mp_route_id`).
		WithArgs("medium").
		WillReturnRows(rows)

	repo := mountainproject.NewPostgresRepository(db)
	routeIDs, err := repo.Sync().GetRoutesDueForCommentSync(context.Background(), "medium")

	if err != nil {
		t.Errorf("GetRoutesDueForCommentSync() error = %v", err)
	}

	if len(routeIDs) != 1 {
		t.Errorf("GetRoutesDueForCommentSync() returned %d IDs, want 1", len(routeIDs))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetPriorityDistribution(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"sync_priority", "count"}).
		AddRow("high", 50000).
		AddRow("medium", 125000).
		AddRow("low", 100000)

	mock.ExpectQuery(`SELECT sync_priority, COUNT`).
		WillReturnRows(rows)

	repo := mountainproject.NewPostgresRepository(db)
	distribution, err := repo.Sync().GetPriorityDistribution(context.Background())

	if err != nil {
		t.Errorf("GetPriorityDistribution() error = %v", err)
	}

	if len(distribution) != 3 {
		t.Errorf("GetPriorityDistribution() returned %d priorities, want 3", len(distribution))
	}

	if distribution["high"] != 50000 {
		t.Errorf("GetPriorityDistribution() high = %v, want 50000", distribution["high"])
	}

	if distribution["medium"] != 125000 {
		t.Errorf("GetPriorityDistribution() medium = %v, want 125000", distribution["medium"])
	}

	if distribution["low"] != 100000 {
		t.Errorf("GetPriorityDistribution() low = %v, want 100000", distribution["low"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
