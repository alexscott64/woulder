package weather_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alexscott64/woulder/backend/internal/database"
	"github.com/alexscott64/woulder/backend/internal/database/weather"
	"github.com/alexscott64/woulder/backend/internal/models"
)

func TestPostgresRepository_Save(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	data := &models.WeatherData{
		LocationID:    1,
		Timestamp:     now,
		Temperature:   72.5,
		FeelsLike:     70.0,
		Precipitation: 0.1,
		Humidity:      65,
		WindSpeed:     10.5,
		WindDirection: 180,
		CloudCover:    50,
		Pressure:      1013,
		Description:   "Partly cloudy",
		Icon:          "02d",
	}

	mock.ExpectExec("INSERT INTO woulder.weather_data").
		WithArgs(
			data.LocationID, data.Timestamp, data.Temperature, data.FeelsLike,
			data.Precipitation, data.Humidity, data.WindSpeed, data.WindDirection,
			data.CloudCover, data.Pressure, data.Description, data.Icon,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := weather.NewPostgresRepository(db)
	err = repo.Save(context.Background(), data)

	if err != nil {
		t.Errorf("Save() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetHistorical(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "location_id", "timestamp", "temperature", "feels_like",
		"precipitation", "humidity", "wind_speed", "wind_direction",
		"cloud_cover", "pressure", "description", "icon", "created_at",
	}).AddRow(
		1, 10, now.Add(-24*time.Hour), 65.0, 63.0,
		0.0, 50, 5.0, 90,
		25, 1015, "Clear", "01d", now.Add(-25*time.Hour),
	).AddRow(
		2, 10, now.Add(-12*time.Hour), 70.0, 68.0,
		0.0, 55, 7.0, 120,
		30, 1014, "Few clouds", "02d", now.Add(-13*time.Hour),
	)

	mock.ExpectQuery("SELECT (.+) FROM woulder.weather_data").
		WithArgs(10, 3).
		WillReturnRows(rows)

	repo := weather.NewPostgresRepository(db)
	result, err := repo.GetHistorical(context.Background(), 10, 3)

	if err != nil {
		t.Errorf("GetHistorical() error = %v", err)
	}

	if len(result) != 2 {
		t.Errorf("GetHistorical() returned %d records, want 2", len(result))
	}

	if result[0].Temperature != 65.0 {
		t.Errorf("GetHistorical() first record temperature = %v, want 65.0", result[0].Temperature)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetHistorical_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"id", "location_id", "timestamp", "temperature", "feels_like",
		"precipitation", "humidity", "wind_speed", "wind_direction",
		"cloud_cover", "pressure", "description", "icon", "created_at",
	})

	mock.ExpectQuery("SELECT (.+) FROM woulder.weather_data").
		WithArgs(999, 7).
		WillReturnRows(rows)

	repo := weather.NewPostgresRepository(db)
	result, err := repo.GetHistorical(context.Background(), 999, 7)

	if err != nil {
		t.Errorf("GetHistorical() error = %v, want nil", err)
	}

	if len(result) != 0 {
		t.Errorf("GetHistorical() returned %d records, want 0", len(result))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetForecast(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "location_id", "timestamp", "temperature", "feels_like",
		"precipitation", "humidity", "wind_speed", "wind_direction",
		"cloud_cover", "pressure", "description", "icon", "created_at",
	}).AddRow(
		3, 10, now.Add(6*time.Hour), 75.0, 73.0,
		0.1, 60, 12.0, 180,
		70, 1012, "Partly cloudy", "03d", now,
	).AddRow(
		4, 10, now.Add(12*time.Hour), 80.0, 78.0,
		0.2, 65, 15.0, 200,
		80, 1011, "Cloudy", "04d", now,
	)

	mock.ExpectQuery("SELECT (.+) FROM woulder.weather_data").
		WithArgs(10, 24).
		WillReturnRows(rows)

	repo := weather.NewPostgresRepository(db)
	result, err := repo.GetForecast(context.Background(), 10, 24)

	if err != nil {
		t.Errorf("GetForecast() error = %v", err)
	}

	if len(result) != 2 {
		t.Errorf("GetForecast() returned %d records, want 2", len(result))
	}

	if result[0].Temperature != 75.0 {
		t.Errorf("GetForecast() first record temperature = %v, want 75.0", result[0].Temperature)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetCurrent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "location_id", "timestamp", "temperature", "feels_like",
		"precipitation", "humidity", "wind_speed", "wind_direction",
		"cloud_cover", "pressure", "description", "icon", "created_at",
	}).AddRow(
		5, 10, now, 72.0, 70.0,
		0.0, 55, 8.0, 150,
		40, 1013, "Clear", "01d", now.Add(-1*time.Hour),
	)

	mock.ExpectQuery("SELECT (.+) FROM woulder.weather_data").
		WithArgs(10).
		WillReturnRows(rows)

	repo := weather.NewPostgresRepository(db)
	result, err := repo.GetCurrent(context.Background(), 10)

	if err != nil {
		t.Errorf("GetCurrent() error = %v", err)
	}

	if result.ID != 5 {
		t.Errorf("GetCurrent() ID = %v, want 5", result.ID)
	}

	if result.Temperature != 72.0 {
		t.Errorf("GetCurrent() temperature = %v, want 72.0", result.Temperature)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetCurrent_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT (.+) FROM woulder.weather_data").
		WithArgs(999).
		WillReturnError(database.ErrNotFound)

	repo := weather.NewPostgresRepository(db)
	_, err = repo.GetCurrent(context.Background(), 999)

	if err == nil {
		t.Error("GetCurrent() expected error for non-existent location, got nil")
	}

	if !database.IsNotFound(err) {
		t.Errorf("GetCurrent() expected ErrNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_CleanOld(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("DELETE FROM woulder.weather_data").
		WithArgs(7).
		WillReturnResult(sqlmock.NewResult(0, 42))

	repo := weather.NewPostgresRepository(db)
	err = repo.CleanOld(context.Background(), 7)

	if err != nil {
		t.Errorf("CleanOld() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_DeleteOldForLocation(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("DELETE FROM woulder.weather_data").
		WithArgs(10, 7).
		WillReturnResult(sqlmock.NewResult(0, 15))

	repo := weather.NewPostgresRepository(db)
	err = repo.DeleteOldForLocation(context.Background(), 10, 7)

	if err != nil {
		t.Errorf("DeleteOldForLocation() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
