package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// SaveMPArea saves or updates a Mountain Project area
func (db *Database) SaveMPArea(ctx context.Context, area *models.MPArea) error {
	query := `
		INSERT INTO woulder.mp_areas (
			mp_area_id, name, parent_mp_area_id, area_type, location_id,
			latitude, longitude, last_synced_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (mp_area_id) DO UPDATE SET
			name = EXCLUDED.name,
			parent_mp_area_id = EXCLUDED.parent_mp_area_id,
			area_type = EXCLUDED.area_type,
			location_id = EXCLUDED.location_id,
			latitude = EXCLUDED.latitude,
			longitude = EXCLUDED.longitude,
			last_synced_at = EXCLUDED.last_synced_at
	`

	_, err := db.conn.ExecContext(ctx, query,
		area.MPAreaID,
		area.Name,
		area.ParentMPAreaID,
		area.AreaType,
		area.LocationID,
		area.Latitude,
		area.Longitude,
		time.Now(),
	)

	return err
}

// SaveMPRoute saves or updates a Mountain Project route
func (db *Database) SaveMPRoute(ctx context.Context, route *models.MPRoute) error {
	query := `
		INSERT INTO woulder.mp_routes (
			mp_route_id, mp_area_id, name, route_type, rating, location_id,
			latitude, longitude, aspect
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (mp_route_id) DO UPDATE SET
			mp_area_id = EXCLUDED.mp_area_id,
			name = EXCLUDED.name,
			route_type = EXCLUDED.route_type,
			rating = EXCLUDED.rating,
			location_id = EXCLUDED.location_id,
			latitude = EXCLUDED.latitude,
			longitude = EXCLUDED.longitude,
			aspect = EXCLUDED.aspect
	`

	_, err := db.conn.ExecContext(ctx, query,
		route.MPRouteID,
		route.MPAreaID,
		route.Name,
		route.RouteType,
		route.Rating,
		route.LocationID,
		route.Latitude,
		route.Longitude,
		route.Aspect,
	)

	return err
}

// SaveMPTick saves a Mountain Project tick
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

// GetMPAreaByID retrieves a Mountain Project area by ID
func (db *Database) GetMPAreaByID(ctx context.Context, mpAreaID string) (*models.MPArea, error) {
	query := `
		SELECT id, mp_area_id, name, parent_mp_area_id, area_type,
		       location_id, latitude, longitude, last_synced_at, created_at, updated_at
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
		&area.Latitude,
		&area.Longitude,
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

// UpdateRouteGPS updates only the GPS coordinates and aspect for a route
func (db *Database) UpdateRouteGPS(ctx context.Context, routeID string, latitude, longitude float64, aspect string) error {
	query := `
		UPDATE woulder.mp_routes
		SET latitude = $1, longitude = $2, aspect = $3, updated_at = NOW()
		WHERE mp_route_id = $4
	`

	_, err := db.conn.ExecContext(ctx, query, latitude, longitude, aspect, routeID)
	return err
}
