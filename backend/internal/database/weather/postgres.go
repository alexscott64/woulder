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
			&d.Icon, &d.CreatedAt,
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
			&d.Icon, &d.CreatedAt,
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
		&d.Icon, &d.CreatedAt,
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
