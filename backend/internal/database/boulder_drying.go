package database

import (
	"context"
	"database/sql"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// GetBoulderDryingProfile retrieves the drying profile for a specific boulder
func (db *Database) GetBoulderDryingProfile(ctx context.Context, mpRouteID string) (*models.BoulderDryingProfile, error) {
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
func (db *Database) GetMPRouteByID(ctx context.Context, mpRouteID string) (*models.MPRoute, error) {
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
