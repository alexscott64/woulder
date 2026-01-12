package database

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
	_ "github.com/lib/pq"
)

//go:embed setup_postgres.sql
var setupSQL string

type Database struct {
	conn *sql.DB
}

func New() (*Database, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	if port == "" {
		port = "5432"
	}
	if sslmode == "" {
		sslmode = "require"
	}

	if host == "" || user == "" || password == "" || dbname == "" {
		return nil, fmt.Errorf("missing required database configuration")
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	database := &Database{conn: db}

	needsInit, err := database.needsInitialization()
	if err != nil {
		return nil, err
	}

	if needsInit {
		log.Println("Database schema not found, running setup...")
		if err := database.runSetup(); err != nil {
			return nil, err
		}
	}

	return database, nil
}

func (db *Database) needsInitialization() (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM information_schema.schemata WHERE schema_name = 'woulder')`
	err := db.conn.QueryRow(query).Scan(&exists)
	return !exists, err
}

func (db *Database) runSetup() error {
	_, err := db.conn.Exec(setupSQL)
	return err
}

func (db *Database) Close() error {
	return db.conn.Close()
}

func (db *Database) Ping(ctx context.Context) error {
	return db.conn.PingContext(ctx)
}

// -------------------- Locations --------------------

func (db *Database) GetAllLocations(ctx context.Context) ([]models.Location, error) {
	query := `
		SELECT id, name, latitude, longitude, elevation_ft, area_id,
		       has_seepage_risk, created_at, updated_at
		FROM woulder.locations
		ORDER BY name
	`

	rows, err := db.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []models.Location
	for rows.Next() {
		var loc models.Location
		if err := rows.Scan(
			&loc.ID, &loc.Name, &loc.Latitude, &loc.Longitude,
			&loc.ElevationFt, &loc.AreaID, &loc.HasSeepageRisk,
			&loc.CreatedAt, &loc.UpdatedAt,
		); err != nil {
			return nil, err
		}
		locations = append(locations, loc)
	}

	return locations, rows.Err()
}

func (db *Database) GetLocation(ctx context.Context, id int) (*models.Location, error) {
	query := `
		SELECT id, name, latitude, longitude, elevation_ft, area_id,
		       created_at, updated_at
		FROM woulder.locations
		WHERE id = $1
	`

	var loc models.Location
	err := db.conn.QueryRowContext(ctx, query, id).Scan(
		&loc.ID, &loc.Name, &loc.Latitude, &loc.Longitude,
		&loc.ElevationFt, &loc.AreaID,
		&loc.CreatedAt, &loc.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &loc, nil
}

func (db *Database) GetLocationsByArea(ctx context.Context, areaID int) ([]models.Location, error) {
	query := `
		SELECT id, name, latitude, longitude, elevation_ft, area_id,
		       created_at, updated_at
		FROM woulder.locations
		WHERE area_id = $1
		ORDER BY name
	`

	rows, err := db.conn.QueryContext(ctx, query, areaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []models.Location
	for rows.Next() {
		var loc models.Location
		if err := rows.Scan(
			&loc.ID, &loc.Name, &loc.Latitude, &loc.Longitude,
			&loc.ElevationFt, &loc.AreaID,
			&loc.CreatedAt, &loc.UpdatedAt,
		); err != nil {
			return nil, err
		}
		locations = append(locations, loc)
	}

	return locations, nil
}

// -------------------- Weather --------------------

func (db *Database) SaveWeatherData(ctx context.Context, data *models.WeatherData) error {
	query := `
		INSERT INTO woulder.weather_data (
			location_id, timestamp, temperature, feels_like, precipitation,
			humidity, wind_speed, wind_direction, cloud_cover, pressure,
			description, icon
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT(location_id, timestamp) DO UPDATE SET
			temperature = EXCLUDED.temperature,
			feels_like = EXCLUDED.feels_like,
			precipitation = EXCLUDED.precipitation,
			humidity = EXCLUDED.humidity,
			wind_speed = EXCLUDED.wind_speed,
			wind_direction = EXCLUDED.wind_direction,
			cloud_cover = EXCLUDED.cloud_cover,
			pressure = EXCLUDED.pressure,
			description = EXCLUDED.description,
			icon = EXCLUDED.icon
	`

	_, err := db.conn.ExecContext(ctx, query,
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

func (db *Database) GetHistoricalWeather(ctx context.Context, locationID, days int) ([]models.WeatherData, error) {
	query := `
		SELECT id, location_id, timestamp, temperature, feels_like,
		       precipitation, humidity, wind_speed, wind_direction,
		       cloud_cover, pressure, description, icon, created_at
		FROM woulder.weather_data
		WHERE location_id = $1
		  AND timestamp >= NOW() - INTERVAL '1 day' * $2
		  AND timestamp <= NOW()
		ORDER BY timestamp ASC
	`

	rows, err := db.conn.QueryContext(ctx, query, locationID, days)
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

	return result, nil
}

func (db *Database) GetForecastWeather(
	ctx context.Context,
	locationID int,
	hours int,
) ([]models.WeatherData, error) {

	query := `
		SELECT id, location_id, timestamp, temperature, feels_like,
		       precipitation, humidity, wind_speed, wind_direction,
		       cloud_cover, pressure, description, icon, created_at
		FROM woulder.weather_data
		WHERE location_id = $1
		  AND timestamp >= NOW() - ($2 * INTERVAL '1 hour')
		ORDER BY timestamp ASC
	`

	rows, err := db.conn.QueryContext(ctx, query, locationID, hours)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.WeatherData
	for rows.Next() {
		var d models.WeatherData
		if err := rows.Scan(
			&d.ID,
			&d.LocationID,
			&d.Timestamp,
			&d.Temperature,
			&d.FeelsLike,
			&d.Precipitation,
			&d.Humidity,
			&d.WindSpeed,
			&d.WindDirection,
			&d.CloudCover,
			&d.Pressure,
			&d.Description,
			&d.Icon,
			&d.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, d)
	}

	return result, nil
}

func (db *Database) GetCurrentWeather(ctx context.Context, locationID int) (*models.WeatherData, error) {
	query := `
		SELECT id, location_id, timestamp, temperature, feels_like,
		       precipitation, humidity, wind_speed, wind_direction,
		       cloud_cover, pressure, description, icon, created_at
		FROM woulder.weather_data
		WHERE location_id = $1
		ORDER BY ABS(EXTRACT(EPOCH FROM (timestamp - NOW())))
		LIMIT 1
	`

	var d models.WeatherData
	err := db.conn.QueryRowContext(ctx, query, locationID).Scan(
		&d.ID, &d.LocationID, &d.Timestamp,
		&d.Temperature, &d.FeelsLike, &d.Precipitation,
		&d.Humidity, &d.WindSpeed, &d.WindDirection,
		&d.CloudCover, &d.Pressure, &d.Description,
		&d.Icon, &d.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &d, nil
}

func (db *Database) CleanOldWeatherData(ctx context.Context, days int) error {
	query := `DELETE FROM woulder.weather_data WHERE timestamp < NOW() - INTERVAL '1 day' * $1`
	_, err := db.conn.ExecContext(ctx, query, days)
	return err
}

// -------------------- Rivers --------------------

func (db *Database) GetRiversByLocation(ctx context.Context, locationID int) ([]models.River, error) {
	query := `
		SELECT id, location_id, gauge_id, river_name,
		       safe_crossing_cfs, caution_crossing_cfs,
		       drainage_area_sq_mi, gauge_drainage_area_sq_mi,
		       flow_divisor, is_estimated, description,
		       created_at, updated_at
		FROM woulder.rivers
		WHERE location_id = $1
		ORDER BY river_name
	`

	rows, err := db.conn.QueryContext(ctx, query, locationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rivers []models.River
	for rows.Next() {
		var r models.River
		if err := rows.Scan(
			&r.ID, &r.LocationID, &r.GaugeID, &r.RiverName,
			&r.SafeCrossingCFS, &r.CautionCrossingCFS,
			&r.DrainageAreaSqMi, &r.GaugeDrainageAreaSqMi,
			&r.FlowDivisor, &r.IsEstimated, &r.Description,
			&r.CreatedAt, &r.UpdatedAt,
		); err != nil {
			return nil, err
		}
		rivers = append(rivers, r)
	}

	return rivers, nil
}

func (db *Database) GetRiverByID(ctx context.Context, id int) (*models.River, error) {
	query := `
		SELECT id, location_id, gauge_id, river_name,
		       safe_crossing_cfs, caution_crossing_cfs,
		       drainage_area_sq_mi, gauge_drainage_area_sq_mi,
		       flow_divisor, is_estimated, description,
		       created_at, updated_at
		FROM woulder.rivers
		WHERE id = $1
	`

	var r models.River
	err := db.conn.QueryRowContext(ctx, query, id).Scan(
		&r.ID, &r.LocationID, &r.GaugeID, &r.RiverName,
		&r.SafeCrossingCFS, &r.CautionCrossingCFS,
		&r.DrainageAreaSqMi, &r.GaugeDrainageAreaSqMi,
		&r.FlowDivisor, &r.IsEstimated, &r.Description,
		&r.CreatedAt, &r.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &r, nil
}

// -------------------- Areas --------------------

func (db *Database) GetAllAreas(ctx context.Context) ([]models.Area, error) {
	query := `
		SELECT id, name, description, region,
		       display_order, created_at, updated_at
		FROM woulder.areas
		ORDER BY display_order, name
	`

	rows, err := db.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var areas []models.Area
	for rows.Next() {
		var a models.Area
		if err := rows.Scan(
			&a.ID, &a.Name, &a.Description,
			&a.Region, &a.DisplayOrder,
			&a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		areas = append(areas, a)
	}

	return areas, nil
}

func (db *Database) GetAreasWithLocationCounts(ctx context.Context) ([]models.AreaWithLocationCount, error) {
	query := `
		SELECT a.id, a.name, a.description, a.region,
		       a.display_order, a.created_at, a.updated_at,
		       COUNT(l.id) AS location_count
		FROM woulder.areas a
		LEFT JOIN woulder.locations l ON l.area_id = a.id
		GROUP BY a.id
		ORDER BY a.display_order, a.name
	`

	rows, err := db.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var areas []models.AreaWithLocationCount
	for rows.Next() {
		var a models.AreaWithLocationCount
		if err := rows.Scan(
			&a.ID, &a.Name, &a.Description,
			&a.Region, &a.DisplayOrder,
			&a.CreatedAt, &a.UpdatedAt,
			&a.LocationCount,
		); err != nil {
			return nil, err
		}
		areas = append(areas, a)
	}

	return areas, nil
}

func (db *Database) GetAreaByID(ctx context.Context, id int) (*models.Area, error) {
	query := `
		SELECT id, name, description, region,
		       display_order, created_at, updated_at
		FROM woulder.areas
		WHERE id = $1
	`

	var a models.Area
	err := db.conn.QueryRowContext(ctx, query, id).Scan(
		&a.ID, &a.Name, &a.Description,
		&a.Region, &a.DisplayOrder,
		&a.CreatedAt, &a.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &a, nil
}

// -------------------- Rock + Sun Exposure --------------------

func (db *Database) GetRockTypesByLocation(ctx context.Context, locationID int) ([]models.RockType, error) {
	query := `
		SELECT rt.id, rt.name, rt.base_drying_hours,
		       rt.porosity_percent, rt.is_wet_sensitive,
		       rt.description, rt.rock_type_group_id,
		       rtg.group_name
		FROM woulder.rock_types rt
		INNER JOIN woulder.location_rock_types lrt ON rt.id = lrt.rock_type_id
		INNER JOIN woulder.rock_type_groups rtg ON rt.rock_type_group_id = rtg.id
		WHERE lrt.location_id = $1
		ORDER BY lrt.is_primary DESC, rt.name ASC
	`

	rows, err := db.conn.QueryContext(ctx, query, locationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rockTypes []models.RockType
	for rows.Next() {
		var rt models.RockType
		if err := rows.Scan(
			&rt.ID, &rt.Name, &rt.BaseDryingHours,
			&rt.PorosityPercent, &rt.IsWetSensitive,
			&rt.Description, &rt.RockTypeGroupID,
			&rt.GroupName,
		); err != nil {
			return nil, err
		}
		rockTypes = append(rockTypes, rt)
	}

	return rockTypes, nil
}

func (db *Database) GetPrimaryRockType(ctx context.Context, locationID int) (*models.RockType, error) {
	query := `
		SELECT rt.id, rt.name, rt.base_drying_hours,
		       rt.porosity_percent, rt.is_wet_sensitive,
		       rt.description, rt.rock_type_group_id,
		       rtg.group_name
		FROM woulder.rock_types rt
		INNER JOIN woulder.location_rock_types lrt ON rt.id = lrt.rock_type_id
		INNER JOIN woulder.rock_type_groups rtg ON rt.rock_type_group_id = rtg.id
		WHERE lrt.location_id = $1 AND lrt.is_primary = TRUE
		LIMIT 1
	`

	var rt models.RockType
	err := db.conn.QueryRowContext(ctx, query, locationID).Scan(
		&rt.ID, &rt.Name, &rt.BaseDryingHours,
		&rt.PorosityPercent, &rt.IsWetSensitive,
		&rt.Description, &rt.RockTypeGroupID,
		&rt.GroupName,
	)

	if err == sql.ErrNoRows {
		rocks, err := db.GetRockTypesByLocation(ctx, locationID)
		if err != nil || len(rocks) == 0 {
			return nil, fmt.Errorf("no rock types found for location %d", locationID)
		}
		return &rocks[0], nil
	}

	if err != nil {
		return nil, err
	}

	return &rt, nil
}

func (db *Database) GetSunExposureByLocation(ctx context.Context, locationID int) (*models.LocationSunExposure, error) {
	query := `
		SELECT id, location_id,
		       south_facing_percent, west_facing_percent,
		       east_facing_percent, north_facing_percent,
		       slab_percent, overhang_percent,
		       tree_coverage_percent, description
		FROM woulder.location_sun_exposure
		WHERE location_id = $1
	`

	var se models.LocationSunExposure
	err := db.conn.QueryRowContext(ctx, query, locationID).Scan(
		&se.ID, &se.LocationID,
		&se.SouthFacingPercent, &se.WestFacingPercent,
		&se.EastFacingPercent, &se.NorthFacingPercent,
		&se.SlabPercent, &se.OverhangPercent,
		&se.TreeCoveragePercent, &se.Description,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &se, nil
}

// -------------------- Mountain Project --------------------

func (db *Database) SaveMPArea(ctx context.Context, area *models.MPArea) error {
	query := `
		INSERT INTO woulder.mp_areas (
			mp_area_id, name, parent_mp_area_id, area_type, location_id, last_synced_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (mp_area_id) DO UPDATE SET
			name = EXCLUDED.name,
			parent_mp_area_id = EXCLUDED.parent_mp_area_id,
			area_type = EXCLUDED.area_type,
			location_id = EXCLUDED.location_id,
			last_synced_at = EXCLUDED.last_synced_at
	`

	_, err := db.conn.ExecContext(ctx, query,
		area.MPAreaID,
		area.Name,
		area.ParentMPAreaID,
		area.AreaType,
		area.LocationID,
		time.Now(),
	)

	return err
}

func (db *Database) SaveMPRoute(ctx context.Context, route *models.MPRoute) error {
	query := `
		INSERT INTO woulder.mp_routes (
			mp_route_id, mp_area_id, name, route_type, rating, location_id
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (mp_route_id) DO UPDATE SET
			mp_area_id = EXCLUDED.mp_area_id,
			name = EXCLUDED.name,
			route_type = EXCLUDED.route_type,
			rating = EXCLUDED.rating,
			location_id = EXCLUDED.location_id
	`

	_, err := db.conn.ExecContext(ctx, query,
		route.MPRouteID,
		route.MPAreaID,
		route.Name,
		route.RouteType,
		route.Rating,
		route.LocationID,
	)

	return err
}

func (db *Database) SaveMPTick(ctx context.Context, tick *models.MPTick) error {
	query := `
		INSERT INTO woulder.mp_ticks (
			mp_route_id, user_name, climbed_at, style, comment
		)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (mp_route_id, user_name, climbed_at) DO NOTHING
	`

	_, err := db.conn.ExecContext(ctx, query,
		tick.MPRouteID,
		tick.UserName,
		tick.ClimbedAt,
		tick.Style,
		tick.Comment,
	)

	return err
}

func (db *Database) GetLastClimbedForLocation(ctx context.Context, locationID int) (*models.LastClimbedInfo, error) {
	query := `
		SELECT
			r.name AS route_name,
			r.rating AS route_rating,
			t.climbed_at,
			t.user_name AS climbed_by,
			t.style,
			t.comment,
			EXTRACT(DAY FROM (NOW() - t.climbed_at))::int AS days_since_climb
		FROM woulder.mp_ticks t
		JOIN woulder.mp_routes r ON t.mp_route_id = r.mp_route_id
		WHERE r.location_id = $1
		ORDER BY t.climbed_at DESC
		LIMIT 1
	`

	var info models.LastClimbedInfo
	err := db.conn.QueryRowContext(ctx, query, locationID).Scan(
		&info.RouteName,
		&info.RouteRating,
		&info.ClimbedAt,
		&info.ClimbedBy,
		&info.Style,
		&info.Comment,
		&info.DaysSinceClimb,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No climb data for this location
	}

	if err != nil {
		return nil, err
	}

	return &info, nil
}

// GetClimbHistoryForLocation retrieves recent climb history for a location with area information
// Includes smart data quality filtering:
// - Adjusts dates that are 1 year in the future (likely typo: 2026 -> 2025)
// - Filters out dates more than 30 days in the future (intentionally bad data)
// - Only shows climbs from the past 2 years (excludes very old or very future dates)
func (db *Database) GetClimbHistoryForLocation(ctx context.Context, locationID int, limit int) ([]models.ClimbHistoryEntry, error) {
	query := `
		WITH adjusted_ticks AS (
			SELECT
				t.mp_route_id,
				t.user_name,
				-- Smart date adjustment: if date is 350-380 days in future, subtract 1 year
				-- This catches typos like "2026" when they meant "2025"
				CASE
					WHEN t.climbed_at > NOW() + INTERVAL '350 days'
					     AND t.climbed_at < NOW() + INTERVAL '380 days'
					THEN t.climbed_at - INTERVAL '1 year'
					ELSE t.climbed_at
				END AS adjusted_climbed_at,
				t.style,
				t.comment
			FROM woulder.mp_ticks t
			WHERE
				-- Filter out dates more than 30 days in the future (bad data)
				t.climbed_at <= NOW() + INTERVAL '30 days'
				-- Filter out climbs older than 2 years (keep it recent)
				AND t.climbed_at >= NOW() - INTERVAL '2 years'
		)
		SELECT
			r.mp_route_id,
			r.name AS route_name,
			r.rating AS route_rating,
			r.mp_area_id,
			a.name AS area_name,
			at.adjusted_climbed_at AS climbed_at,
			at.user_name AS climbed_by,
			at.style,
			at.comment,
			EXTRACT(DAY FROM (NOW() - at.adjusted_climbed_at))::int AS days_since_climb
		FROM adjusted_ticks at
		JOIN woulder.mp_routes r ON at.mp_route_id = r.mp_route_id
		JOIN woulder.mp_areas a ON r.mp_area_id = a.mp_area_id
		WHERE r.location_id = $1
		ORDER BY at.adjusted_climbed_at DESC
		LIMIT $2
	`

	rows, err := db.conn.QueryContext(ctx, query, locationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []models.ClimbHistoryEntry
	for rows.Next() {
		var entry models.ClimbHistoryEntry
		err := rows.Scan(
			&entry.MPRouteID,
			&entry.RouteName,
			&entry.RouteRating,
			&entry.MPAreaID,
			&entry.AreaName,
			&entry.ClimbedAt,
			&entry.ClimbedBy,
			&entry.Style,
			&entry.Comment,
			&entry.DaysSinceClimb,
		)
		if err != nil {
			return nil, err
		}
		history = append(history, entry)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return history, nil
}

func (db *Database) GetMPAreaByID(ctx context.Context, mpAreaID string) (*models.MPArea, error) {
	query := `
		SELECT id, mp_area_id, name, parent_mp_area_id, area_type,
		       location_id, last_synced_at, created_at, updated_at
		FROM woulder.mp_areas
		WHERE mp_area_id = $1
	`

	var area models.MPArea
	err := db.conn.QueryRowContext(ctx, query, mpAreaID).Scan(
		&area.ID,
		&area.MPAreaID,
		&area.Name,
		&area.ParentMPAreaID,
		&area.AreaType,
		&area.LocationID,
		&area.LastSyncedAt,
		&area.CreatedAt,
		&area.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &area, nil
}

// GetLastTickTimestampForRoute returns the timestamp of the most recent tick for a route
func (db *Database) GetLastTickTimestampForRoute(ctx context.Context, routeID string) (*time.Time, error) {
	query := `
		SELECT MAX(climbed_at) AS last_tick
		FROM woulder.mp_ticks
		WHERE mp_route_id = $1
	`

	var lastTick *time.Time
	err := db.conn.QueryRowContext(ctx, query, routeID).Scan(&lastTick)

	if err == sql.ErrNoRows {
		return nil, nil // No ticks for this route yet
	}

	if err != nil {
		return nil, err
	}

	return lastTick, nil
}

// GetAllRouteIDsForLocation returns all route IDs associated with a location
func (db *Database) GetAllRouteIDsForLocation(ctx context.Context, locationID int) ([]string, error) {
	query := `
		SELECT mp_route_id
		FROM woulder.mp_routes
		WHERE location_id = $1
	`

	rows, err := db.conn.QueryContext(ctx, query, locationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routeIDs []string
	for rows.Next() {
		var routeID string
		if err := rows.Scan(&routeID); err != nil {
			return nil, err
		}
		routeIDs = append(routeIDs, routeID)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return routeIDs, nil
}
