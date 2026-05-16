package weather

import (
	"context"
	"database/sql"

	"github.com/alexscott64/woulder/backend/internal/database/dberrors"
	"github.com/alexscott64/woulder/backend/internal/models"
)

// PostgresRepository implements Repository using PostgreSQL.
type PostgresRepository struct {
	db DBConn
}

// NewPostgresRepository creates a new PostgreSQL weather repository.
// The db parameter can be either *sql.DB or *sql.Tx, allowing this
// repository to work both standalone and within transactions.
func NewPostgresRepository(db DBConn) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Save saves or updates weather data for a location.
func (r *PostgresRepository) Save(ctx context.Context, data *models.WeatherData) error {
	_, err := r.db.ExecContext(ctx, querySave,
		data.LocationID,
		data.Timestamp,
		data.Temperature,
		data.FeelsLike,
		data.Precipitation,
		data.Humidity,
		data.WindSpeed,
		data.WindDirection,
		data.CloudCover,
		data.Pressure,
		data.Description,
		data.Icon,
		data.ShortwaveRadiation,
		data.DirectRadiation,
		data.DiffuseRadiation,
		data.DewpointF,
	)
	return err
}

// GetHistorical retrieves weather data from the past N days.
func (r *PostgresRepository) GetHistorical(ctx context.Context, locationID int, days int) ([]models.WeatherData, error) {
	rows, err := r.db.QueryContext(ctx, queryGetHistorical, locationID, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.WeatherData
	for rows.Next() {
		var d models.WeatherData
		if err := rows.Scan(
			&d.ID, &d.LocationID, &d.Timestamp,
			&d.Temperature, &d.FeelsLike, &d.Precipitation,
			&d.Humidity, &d.WindSpeed, &d.WindDirection,
			&d.CloudCover, &d.Pressure, &d.Description,
			&d.Icon,
			&d.ShortwaveRadiation, &d.DirectRadiation, &d.DiffuseRadiation, &d.DewpointF,
			&d.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, d)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// GetForecast retrieves forecast weather data for the next N hours.
func (r *PostgresRepository) GetForecast(ctx context.Context, locationID int, hours int) ([]models.WeatherData, error) {
	rows, err := r.db.QueryContext(ctx, queryGetForecast, locationID, hours)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.WeatherData
	for rows.Next() {
		var d models.WeatherData
		if err := rows.Scan(
			&d.ID, &d.LocationID, &d.Timestamp,
			&d.Temperature, &d.FeelsLike, &d.Precipitation,
			&d.Humidity, &d.WindSpeed, &d.WindDirection,
			&d.CloudCover, &d.Pressure, &d.Description,
			&d.Icon,
			&d.ShortwaveRadiation, &d.DirectRadiation, &d.DiffuseRadiation, &d.DewpointF,
			&d.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, d)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// GetCurrent retrieves the most recent weather data for a location.
func (r *PostgresRepository) GetCurrent(ctx context.Context, locationID int) (*models.WeatherData, error) {
	var d models.WeatherData
	err := r.db.QueryRowContext(ctx, queryGetCurrent, locationID).Scan(
		&d.ID, &d.LocationID, &d.Timestamp,
		&d.Temperature, &d.FeelsLike, &d.Precipitation,
		&d.Humidity, &d.WindSpeed, &d.WindDirection,
		&d.CloudCover, &d.Pressure, &d.Description,
		&d.Icon,
		&d.ShortwaveRadiation, &d.DirectRadiation, &d.DiffuseRadiation, &d.DewpointF,
		&d.CreatedAt,
	)

	if err != nil {
		return nil, dberrors.WrapNotFound(err)
	}

	return &d, nil
}

// CleanOld deletes weather data older than the specified number of days.
func (r *PostgresRepository) CleanOld(ctx context.Context, daysToKeep int) error {
	_, err := r.db.ExecContext(ctx, queryCleanOld, daysToKeep)
	return err
}

// DeleteOldForLocation deletes stale weather data for a specific location.
func (r *PostgresRepository) DeleteOldForLocation(ctx context.Context, locationID int, daysToKeep int) error {
	_, err := r.db.ExecContext(ctx, queryDeleteOldForLocation, locationID, daysToKeep)
	return err
}

// DeleteFutureForLocation deletes ALL future forecast data for a specific location.
func (r *PostgresRepository) DeleteFutureForLocation(ctx context.Context, locationID int) error {
	_, err := r.db.ExecContext(ctx, queryDeleteFutureForLocation, locationID)
	return err
}

// ReplaceFutureForLocation atomically replaces future forecast rows for a
// location: it deletes ALL current future rows and inserts the supplied rows
// in a single transaction. Either all changes commit, or none do — eliminating
// the destructive intermediate state where the cache was purged but a fresh
// (potentially truncated) save did not complete.
//
// When the underlying DBConn is a *sql.DB this opens a real transaction.
// When it is already a *sql.Tx (i.e. caller is composing into a larger
// transaction) we run the operations inline on that tx. As a safety fallback
// for any other DBConn implementation (mocks, etc.), we run the operations
// sequentially without a transaction — callers in the hot path always use
// *sql.DB in production.
//
// Note: locationID is also stamped onto each row before save, so callers can
// pass rows fresh from the API client without pre-setting LocationID.
func (r *PostgresRepository) ReplaceFutureForLocation(ctx context.Context, locationID int, rows []models.WeatherData) error {
	exec := func(conn DBConn) error {
		if _, err := conn.ExecContext(ctx, queryDeleteFutureForLocation, locationID); err != nil {
			return err
		}
		for i := range rows {
			rows[i].LocationID = locationID
			d := &rows[i]
			if _, err := conn.ExecContext(ctx, querySave,
				d.LocationID, d.Timestamp, d.Temperature, d.FeelsLike,
				d.Precipitation, d.Humidity, d.WindSpeed, d.WindDirection,
				d.CloudCover, d.Pressure, d.Description, d.Icon,
				d.ShortwaveRadiation, d.DirectRadiation, d.DiffuseRadiation, d.DewpointF,
			); err != nil {
				return err
			}
		}
		return nil
	}

	switch conn := r.db.(type) {
	case *sql.DB:
		tx, err := conn.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		if err := exec(tx); err != nil {
			_ = tx.Rollback()
			return err
		}
		return tx.Commit()
	case *sql.Tx:
		// Already in a transaction — run inline.
		return exec(conn)
	default:
		// Fallback (e.g. test mocks): no transaction available.
		return exec(r.db)
	}
}

// UpsertDailyAggregates upserts daily weather rollups for a location/date range.
func (r *PostgresRepository) UpsertDailyAggregates(ctx context.Context, locationID int, startDate, endDate string) error {
	_, err := r.db.ExecContext(ctx, queryUpsertDailyAggregates, locationID, startDate, endDate)
	return err
}

// GetDailyAggregates returns daily weather rollups for a location/date range.
func (r *PostgresRepository) GetDailyAggregates(ctx context.Context, locationID int, startDate, endDate string) ([]models.WeatherDailyAggregate, error) {
	rows, err := r.db.QueryContext(ctx, queryGetDailyAggregates, locationID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]models.WeatherDailyAggregate, 0)
	for rows.Next() {
		var d models.WeatherDailyAggregate
		var sunrise sql.NullTime
		var sunset sql.NullTime
		if err := rows.Scan(
			&d.ID,
			&d.LocationID,
			&d.LocalDate,
			&d.MinTemperature,
			&d.MaxTemperature,
			&d.AvgTemperature,
			&d.TotalPrecipitation,
			&d.AvgHumidity,
			&d.AvgWindSpeed,
			&d.SnowEstimateInches,
			&sunrise,
			&sunset,
			&d.SourceHourCount,
			&d.CreatedAt,
			&d.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if sunrise.Valid {
			t := sunrise.Time
			d.SunriseAt = &t
		}
		if sunset.Valid {
			t := sunset.Time
			d.SunsetAt = &t
		}
		result = append(result, d)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
