package database

import (
	"context"
	"database/sql"

	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/lib/pq"
)

// GetBoulderDryingProfile retrieves the drying profile for a specific boulder
func (db *Database) GetBoulderDryingProfile(ctx context.Context, mpRouteID int64) (*models.BoulderDryingProfile, error) {
	query := `
		SELECT id, mp_route_id, tree_coverage_percent, rock_type_override,
		       last_sun_calc_at, sun_exposure_hours_cache, created_at, updated_at
		FROM woulder.boulder_drying_profiles
		WHERE mp_route_id = $1
	`

	var profile models.BoulderDryingProfile
	err := db.conn.QueryRowContext(ctx, query, mpRouteID).Scan(
		&profile.ID,
		&profile.MPRouteID,
		&profile.TreeCoveragePercent,
		&profile.RockTypeOverride,
		&profile.LastSunCalcAt,
		&profile.SunExposureHoursCache,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No profile exists yet
	}

	if err != nil {
		return nil, err
	}

	return &profile, nil
}

// GetBoulderDryingProfilesByRouteIDs retrieves multiple boulder drying profiles by route IDs in a single query
// This is significantly faster than calling GetBoulderDryingProfile in a loop (eliminates N+1 problem)
func (db *Database) GetBoulderDryingProfilesByRouteIDs(ctx context.Context, mpRouteIDs []int64) (map[int64]*models.BoulderDryingProfile, error) {
	if len(mpRouteIDs) == 0 {
		return make(map[int64]*models.BoulderDryingProfile), nil
	}

	query := `
		SELECT id, mp_route_id, tree_coverage_percent, rock_type_override,
		       last_sun_calc_at, sun_exposure_hours_cache, created_at, updated_at
		FROM woulder.boulder_drying_profiles
		WHERE mp_route_id = ANY($1)
	`

	rows, err := db.conn.QueryContext(ctx, query, pq.Array(mpRouteIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	profiles := make(map[int64]*models.BoulderDryingProfile)
	for rows.Next() {
		var profile models.BoulderDryingProfile
		err := rows.Scan(
			&profile.ID,
			&profile.MPRouteID,
			&profile.TreeCoveragePercent,
			&profile.RockTypeOverride,
			&profile.LastSunCalcAt,
			&profile.SunExposureHoursCache,
			&profile.CreatedAt,
			&profile.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		profiles[profile.MPRouteID] = &profile
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return profiles, nil
}

// SaveBoulderDryingProfile creates or updates a boulder drying profile
func (db *Database) SaveBoulderDryingProfile(ctx context.Context, profile *models.BoulderDryingProfile) error {
	query := `
		INSERT INTO woulder.boulder_drying_profiles (
			mp_route_id, tree_coverage_percent, rock_type_override,
			last_sun_calc_at, sun_exposure_hours_cache
		)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (mp_route_id) DO UPDATE SET
			tree_coverage_percent = EXCLUDED.tree_coverage_percent,
			rock_type_override = EXCLUDED.rock_type_override,
			last_sun_calc_at = EXCLUDED.last_sun_calc_at,
			sun_exposure_hours_cache = EXCLUDED.sun_exposure_hours_cache,
			updated_at = NOW()
	`

	_, err := db.conn.ExecContext(ctx, query,
		profile.MPRouteID,
		profile.TreeCoveragePercent,
		profile.RockTypeOverride,
		profile.LastSunCalcAt,
		profile.SunExposureHoursCache,
	)

	return err
}

// GetMPRouteByID retrieves a Mountain Project route by ID
func (db *Database) GetMPRouteByID(ctx context.Context, mpRouteID int64) (*models.MPRoute, error) {
	query := `
		SELECT id, mp_route_id, mp_area_id, name, route_type, rating,
		       location_id, latitude, longitude, aspect, created_at, updated_at
		FROM woulder.mp_routes
		WHERE mp_route_id = $1
	`

	var route models.MPRoute
	err := db.conn.QueryRowContext(ctx, query, mpRouteID).Scan(
		&route.ID,
		&route.MPRouteID,
		&route.MPAreaID,
		&route.Name,
		&route.RouteType,
		&route.Rating,
		&route.LocationID,
		&route.Latitude,
		&route.Longitude,
		&route.Aspect,
		&route.CreatedAt,
		&route.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &route, nil
}

// GetMPRoutesByIDs retrieves multiple Mountain Project routes by IDs in a single query
// This is significantly faster than calling GetMPRouteByID in a loop (eliminates N+1 problem)
func (db *Database) GetMPRoutesByIDs(ctx context.Context, mpRouteIDs []int64) (map[int64]*models.MPRoute, error) {
	if len(mpRouteIDs) == 0 {
		return make(map[int64]*models.MPRoute), nil
	}

	query := `
		SELECT id, mp_route_id, mp_area_id, name, route_type, rating,
		       location_id, latitude, longitude, aspect, created_at, updated_at
		FROM woulder.mp_routes
		WHERE mp_route_id = ANY($1)
	`

	rows, err := db.conn.QueryContext(ctx, query, pq.Array(mpRouteIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	routes := make(map[int64]*models.MPRoute)
	for rows.Next() {
		var route models.MPRoute
		err := rows.Scan(
			&route.ID,
			&route.MPRouteID,
			&route.MPAreaID,
			&route.Name,
			&route.RouteType,
			&route.Rating,
			&route.LocationID,
			&route.Latitude,
			&route.Longitude,
			&route.Aspect,
			&route.CreatedAt,
			&route.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		routes[route.MPRouteID] = &route
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return routes, nil
}

// GetLocationByID retrieves a location by its ID
func (db *Database) GetLocationByID(ctx context.Context, locationID int) (*models.Location, error) {
	query := `
		SELECT id, name, latitude, longitude, area_id, created_at, updated_at
		FROM woulder.locations
		WHERE id = $1
	`

	var location models.Location
	err := db.conn.QueryRowContext(ctx, query, locationID).Scan(
		&location.ID,
		&location.Name,
		&location.Latitude,
		&location.Longitude,
		&location.AreaID,
		&location.CreatedAt,
		&location.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &location, nil
}

// GetRoutesWithGPSByArea retrieves all routes in an area that have GPS coordinates
// This is used for calculating area-level drying statistics
func (db *Database) GetRoutesWithGPSByArea(ctx context.Context, mpAreaID int64) ([]*models.MPRoute, error) {
	query := `
		WITH RECURSIVE area_tree AS (
			-- Start with the given area
			SELECT mp_area_id
			FROM woulder.mp_areas
			WHERE mp_area_id = $1

			UNION ALL

			-- Recursively include all descendant areas
			SELECT a.mp_area_id
			FROM woulder.mp_areas a
			INNER JOIN area_tree at ON a.parent_mp_area_id = at.mp_area_id
		)
		SELECT id, mp_route_id, mp_area_id, name, route_type, rating,
		       location_id, latitude, longitude, aspect, created_at, updated_at
		FROM woulder.mp_routes r
		WHERE r.mp_area_id IN (SELECT mp_area_id FROM area_tree)
		  AND r.latitude IS NOT NULL
		  AND r.longitude IS NOT NULL
		ORDER BY r.name
	`

	rows, err := db.conn.QueryContext(ctx, query, mpAreaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routes []*models.MPRoute
	for rows.Next() {
		var route models.MPRoute
		err := rows.Scan(
			&route.ID,
			&route.MPRouteID,
			&route.MPAreaID,
			&route.Name,
			&route.RouteType,
			&route.Rating,
			&route.LocationID,
			&route.Latitude,
			&route.Longitude,
			&route.Aspect,
			&route.CreatedAt,
			&route.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		routes = append(routes, &route)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return routes, nil
}
