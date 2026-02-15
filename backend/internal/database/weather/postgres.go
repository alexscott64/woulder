package weather

import (
	"context"

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
