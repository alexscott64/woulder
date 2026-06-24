package weather

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/alexscott64/woulder/backend/internal/database/dberrors"
	"github.com/alexscott64/woulder/backend/internal/models"
)

// weatherDataColumnCount is the number of columns inserted per row by
// bulkInsertForecast. Must stay in sync with the column list in
// buildBulkInsertQuery and with querySave.
const weatherDataColumnCount = 16

// maxBulkInsertRows caps the number of rows in a single bulk INSERT.
// PostgreSQL allows up to 65,535 bind parameters per statement (uint16);
// at 16 params per row that's 4096 rows. We pick a comfortable margin
// below that to leave room for query-planner overhead and to keep any one
// transaction's WAL footprint bounded. The expected payload from Open-Meteo
// is ~396 rows, so this only matters as a safety valve.
const maxBulkInsertRows = 2000

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
//
// PERFORMANCE: Previous implementation issued 1 DELETE + N (~384) single-row
// INSERTs per call inside the transaction. Across ~34 locations on the
// background refresh loop that produced ~13k statements per cycle, each one
// a network round-trip to RDS, generating enough WAL churn to spike RDS
// checkpoint lag into the multi-minute range. The current implementation
// collapses the inserts into a single multi-row INSERT (chunked at
// maxBulkInsertRows for safety against PostgreSQL's 65535-parameter limit),
// reducing the per-location write to 1 DELETE + 1 INSERT (or a small handful
// of INSERTs in the unlikely event of a >2000-row payload).
func (r *PostgresRepository) ReplaceFutureForLocation(ctx context.Context, locationID int, rows []models.WeatherData) error {
	// Stamp the location ID on every row up-front so the bulk-insert path
	// can read the field uniformly. This also matches the previous
	// behaviour where callers could pass rows fresh from the API client
	// without pre-setting LocationID.
	for i := range rows {
		rows[i].LocationID = locationID
	}

	exec := func(conn DBConn) error {
		if _, err := conn.ExecContext(ctx, queryDeleteFutureForLocation, locationID); err != nil {
			return err
		}
		return bulkInsertForecast(ctx, conn, rows)
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

// bulkInsertForecast inserts the supplied weather rows in chunks of
// maxBulkInsertRows, each chunk as a single multi-row INSERT ... VALUES
// statement. Uses the same ON CONFLICT(location_id, timestamp) DO UPDATE
// behaviour as the single-row querySave, so re-inserting an existing
// (location_id, timestamp) refreshes the row in place.
//
// No-op if rows is empty.
func bulkInsertForecast(ctx context.Context, conn DBConn, rows []models.WeatherData) error {
	if len(rows) == 0 {
		return nil
	}
	for start := 0; start < len(rows); start += maxBulkInsertRows {
		end := start + maxBulkInsertRows
		if end > len(rows) {
			end = len(rows)
		}
		chunk := rows[start:end]
		query, args := buildBulkInsertQuery(chunk)
		if _, err := conn.ExecContext(ctx, query, args...); err != nil {
			return err
		}
	}
	return nil
}

// buildBulkInsertQuery constructs a single multi-row INSERT for the given
// chunk of weather rows. Column list and ON CONFLICT clause mirror querySave
// in queries.go so the cache-replacement path has identical upsert semantics
// to the per-row Save path. The chunk MUST be non-empty.
func buildBulkInsertQuery(chunk []models.WeatherData) (string, []interface{}) {
	var b strings.Builder
	b.WriteString(`INSERT INTO woulder.weather_data (
		location_id, timestamp, temperature, feels_like, precipitation,
		humidity, wind_speed, wind_direction, cloud_cover, pressure,
		description, icon,
		shortwave_radiation, direct_radiation, diffuse_radiation, dewpoint_f
	) VALUES `)

	args := make([]interface{}, 0, len(chunk)*weatherDataColumnCount)
	for i, d := range chunk {
		if i > 0 {
			b.WriteString(",")
		}
		base := i * weatherDataColumnCount
		// $1..$16 for the first row, $17..$32 for the second, etc.
		fmt.Fprintf(&b,
			"($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
			base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8,
			base+9, base+10, base+11, base+12, base+13, base+14, base+15, base+16,
		)
		args = append(args,
			d.LocationID, d.Timestamp, d.Temperature, d.FeelsLike,
			d.Precipitation, d.Humidity, d.WindSpeed, d.WindDirection,
			d.CloudCover, d.Pressure, d.Description, d.Icon,
			d.ShortwaveRadiation, d.DirectRadiation, d.DiffuseRadiation, d.DewpointF,
		)
	}

	b.WriteString(` ON CONFLICT(location_id, timestamp) DO UPDATE SET
		temperature = EXCLUDED.temperature,
		feels_like = EXCLUDED.feels_like,
		precipitation = EXCLUDED.precipitation,
		humidity = EXCLUDED.humidity,
		wind_speed = EXCLUDED.wind_speed,
		wind_direction = EXCLUDED.wind_direction,
		cloud_cover = EXCLUDED.cloud_cover,
		pressure = EXCLUDED.pressure,
		description = EXCLUDED.description,
		icon = EXCLUDED.icon,
		shortwave_radiation = EXCLUDED.shortwave_radiation,
		direct_radiation = EXCLUDED.direct_radiation,
		diffuse_radiation = EXCLUDED.diffuse_radiation,
		dewpoint_f = EXCLUDED.dewpoint_f,
		created_at = CURRENT_TIMESTAMP
	WHERE weather_data.temperature         IS DISTINCT FROM EXCLUDED.temperature
	   OR weather_data.feels_like          IS DISTINCT FROM EXCLUDED.feels_like
	   OR weather_data.precipitation       IS DISTINCT FROM EXCLUDED.precipitation
	   OR weather_data.humidity            IS DISTINCT FROM EXCLUDED.humidity
	   OR weather_data.wind_speed          IS DISTINCT FROM EXCLUDED.wind_speed
	   OR weather_data.wind_direction      IS DISTINCT FROM EXCLUDED.wind_direction
	   OR weather_data.cloud_cover         IS DISTINCT FROM EXCLUDED.cloud_cover
	   OR weather_data.pressure            IS DISTINCT FROM EXCLUDED.pressure
	   OR weather_data.description         IS DISTINCT FROM EXCLUDED.description
	   OR weather_data.icon                IS DISTINCT FROM EXCLUDED.icon
	   OR weather_data.shortwave_radiation IS DISTINCT FROM EXCLUDED.shortwave_radiation
	   OR weather_data.direct_radiation    IS DISTINCT FROM EXCLUDED.direct_radiation
	   OR weather_data.diffuse_radiation   IS DISTINCT FROM EXCLUDED.diffuse_radiation
	   OR weather_data.dewpoint_f          IS DISTINCT FROM EXCLUDED.dewpoint_f`)

	return b.String(), args
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
