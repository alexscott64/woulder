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

// SaveAreaComment saves a Mountain Project area comment
func (db *Database) SaveAreaComment(ctx context.Context, mpCommentID, mpAreaID int64, userName, commentText string, commentedAt time.Time) error {
	query := `
		INSERT INTO woulder.mp_comments (
			mp_comment_id, comment_type, mp_area_id, mp_route_id,
			user_name, user_id, comment_text, commented_at
		)
		VALUES ($1, 'area', $2, NULL, $3, NULL, $4, $5)
		ON CONFLICT (mp_comment_id)
		DO UPDATE SET
			user_name = EXCLUDED.user_name,
			comment_text = EXCLUDED.comment_text,
			commented_at = EXCLUDED.commented_at,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := db.conn.ExecContext(ctx, query,
		mpCommentID,
		mpAreaID,
		userName,
		commentText,
		commentedAt,
	)

	return err
}

// SaveRouteComment saves a Mountain Project route comment
func (db *Database) SaveRouteComment(ctx context.Context, mpCommentID, mpRouteID int64, userName, commentText string, commentedAt time.Time) error {
	query := `
		INSERT INTO woulder.mp_comments (
			mp_comment_id, comment_type, mp_area_id, mp_route_id,
			user_name, user_id, comment_text, commented_at
		)
		VALUES ($1, 'route', NULL, $2, $3, NULL, $4, $5)
		ON CONFLICT (mp_comment_id)
		DO UPDATE SET
			user_name = EXCLUDED.user_name,
			comment_text = EXCLUDED.comment_text,
			commented_at = EXCLUDED.commented_at,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := db.conn.ExecContext(ctx, query,
		mpCommentID,
		mpRouteID,
		userName,
		commentText,
		commentedAt,
	)

	return err
}

// GetMPAreaByID retrieves a Mountain Project area by ID
func (db *Database) GetMPAreaByID(ctx context.Context, mpAreaID int64) (*models.MPArea, error) {
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
func (db *Database) GetLastTickTimestampForRoute(ctx context.Context, routeID int64) (*time.Time, error) {
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
func (db *Database) GetAllRouteIDsForLocation(ctx context.Context, locationID int) ([]int64, error) {
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

	var routeIDs []int64
	for rows.Next() {
		var routeID int64
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
func (db *Database) UpdateRouteGPS(ctx context.Context, routeID int64, latitude, longitude float64, aspect string) error {
	query := `
		UPDATE woulder.mp_routes
		SET latitude = $1, longitude = $2, aspect = $3, updated_at = NOW()
		WHERE mp_route_id = $4
	`

	_, err := db.conn.ExecContext(ctx, query, latitude, longitude, aspect, routeID)
	return err
}

// UpdateAreaRouteCount updates the cached route count for an area
func (db *Database) UpdateAreaRouteCount(ctx context.Context, mpAreaID string, total int) error {
	query := `
		UPDATE woulder.mp_areas
		SET route_count_total = $1,
		    route_count_last_checked = NOW(),
		    updated_at = NOW()
		WHERE mp_area_id = $2
	`
	_, err := db.conn.ExecContext(ctx, query, total, mpAreaID)
	return err
}

// GetAreaRouteCount retrieves the cached route count for an area
// Returns -1 if the area doesn't exist or hasn't been checked yet
func (db *Database) GetAreaRouteCount(ctx context.Context, mpAreaID string) (int, error) {
	var count sql.NullInt64
	query := `SELECT route_count_total FROM woulder.mp_areas WHERE mp_area_id = $1`
	err := db.conn.QueryRowContext(ctx, query, mpAreaID).Scan(&count)
	if err == sql.ErrNoRows {
		return -1, nil
	}
	if err != nil {
		return -1, err
	}
	if !count.Valid {
		return -1, nil // Count not yet set
	}
	return int(count.Int64), nil
}

// GetChildAreas retrieves all direct children of an area
func (db *Database) GetChildAreas(ctx context.Context, parentMPAreaID string) ([]struct {
	MPAreaID string
	Name     string
}, error) {
	query := `
		SELECT mp_area_id, name
		FROM woulder.mp_areas
		WHERE parent_mp_area_id = $1
		ORDER BY name
	`
	rows, err := db.conn.QueryContext(ctx, query, parentMPAreaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var areas []struct {
		MPAreaID string
		Name     string
	}
	for rows.Next() {
		var area struct {
			MPAreaID string
			Name     string
		}
		if err := rows.Scan(&area.MPAreaID, &area.Name); err != nil {
			return nil, err
		}
		areas = append(areas, area)
	}
	return areas, rows.Err()
}

// GetRouteIDsForArea retrieves all route IDs currently in an area
func (db *Database) GetRouteIDsForArea(ctx context.Context, mpAreaID string) ([]string, error) {
	query := `
		SELECT mp_route_id::text
		FROM woulder.mp_routes
		WHERE mp_area_id = $1
		ORDER BY mp_route_id
	`
	rows, err := db.conn.QueryContext(ctx, query, mpAreaID)
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
	return routeIDs, rows.Err()
}

// GetAllStateConfigs retrieves all state configurations
func (db *Database) GetAllStateConfigs(ctx context.Context) ([]struct {
	StateName string
	MPAreaID  string
	IsActive  bool
}, error) {
	query := `
		SELECT state_name, mp_area_id, is_active
		FROM woulder.mp_state_configs
		ORDER BY display_order, state_name
	`
	rows, err := db.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []struct {
		StateName string
		MPAreaID  string
		IsActive  bool
	}
	for rows.Next() {
		var config struct {
			StateName string
			MPAreaID  string
			IsActive  bool
		}
		if err := rows.Scan(&config.StateName, &config.MPAreaID, &config.IsActive); err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}
	return configs, rows.Err()
}

// UpsertRoute inserts or updates a route (compatibility with mountainprojectsync)
func (db *Database) UpsertRoute(ctx context.Context, mpRouteID, mpAreaID int64, locationID *int, name, routeType, rating string, lat, lon *float64, aspect *string) error {
	query := `
		INSERT INTO woulder.mp_routes (
			mp_route_id, mp_area_id, location_id, name, route_type, rating,
			latitude, longitude, aspect
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (mp_route_id) DO UPDATE SET
			mp_area_id = EXCLUDED.mp_area_id,
			name = EXCLUDED.name,
			route_type = EXCLUDED.route_type,
			rating = EXCLUDED.rating,
			latitude = EXCLUDED.latitude,
			longitude = EXCLUDED.longitude,
			aspect = EXCLUDED.aspect,
			updated_at = NOW()
	`

	_, err := db.conn.ExecContext(ctx, query, mpRouteID, mpAreaID, locationID, name, routeType, rating, lat, lon, aspect)
	return err
}

// UpsertTick inserts or updates a tick (compatibility with mountainprojectsync)
func (db *Database) UpsertTick(ctx context.Context, mpRouteID int64, userName string, climbedAt time.Time, style string, comment *string) error {
	query := `
		INSERT INTO woulder.mp_ticks (mp_route_id, user_name, climbed_at, style, comment)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (mp_route_id, user_name, climbed_at) DO NOTHING
	`

	_, err := db.conn.ExecContext(ctx, query, mpRouteID, userName, climbedAt, style, comment)
	return err
}

// UpsertAreaComment inserts or updates an area comment (compatibility with mountainprojectsync)
func (db *Database) UpsertAreaComment(ctx context.Context, mpCommentID, mpAreaID int64, userName string, userID *string, commentText string, commentedAt time.Time) error {
	query := `
		INSERT INTO woulder.mp_comments (
			mp_comment_id, comment_type, mp_area_id, mp_route_id,
			user_name, user_id, comment_text, commented_at
		)
		VALUES ($1, 'area', $2, NULL, $3, $4, $5, $6)
		ON CONFLICT (mp_comment_id) DO UPDATE SET
			user_name = EXCLUDED.user_name,
			user_id = EXCLUDED.user_id,
			comment_text = EXCLUDED.comment_text,
			commented_at = EXCLUDED.commented_at,
			updated_at = NOW()
	`

	_, err := db.conn.ExecContext(ctx, query, mpCommentID, mpAreaID, userName, userID, commentText, commentedAt)
	return err
}

// UpsertRouteComment inserts or updates a route comment (compatibility with mountainprojectsync)
func (db *Database) UpsertRouteComment(ctx context.Context, mpCommentID, mpRouteID int64, userName string, userID *string, commentText string, commentedAt time.Time) error {
	query := `
		INSERT INTO woulder.mp_comments (
			mp_comment_id, comment_type, mp_area_id, mp_route_id,
			user_name, user_id, comment_text, commented_at
		)
		VALUES ($1, 'route', NULL, $2, $3, $4, $5, $6)
		ON CONFLICT (mp_comment_id) DO UPDATE SET
			user_name = EXCLUDED.user_name,
			user_id = EXCLUDED.user_id,
			comment_text = EXCLUDED.comment_text,
			commented_at = EXCLUDED.commented_at,
			updated_at = NOW()
	`

	_, err := db.conn.ExecContext(ctx, query, mpCommentID, mpRouteID, userName, userID, commentText, commentedAt)
	return err
}
