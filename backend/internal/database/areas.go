package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// GetAllAreas retrieves all active areas ordered by display order and name
func (db *Database) GetAllAreas(ctx context.Context) ([]models.Area, error) {
	query := `
		SELECT id, name, description, region,
		       display_order, is_active, created_at, updated_at
		FROM woulder.areas
		WHERE is_active = TRUE
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
			&a.Region, &a.DisplayOrder, &a.IsActive,
			&a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		areas = append(areas, a)
	}

	return areas, nil
}

// GetAreasWithLocationCounts retrieves all active areas with their location counts
func (db *Database) GetAreasWithLocationCounts(ctx context.Context) ([]models.AreaWithLocationCount, error) {
	query := `
		SELECT a.id, a.name, a.description, a.region,
		       a.display_order, a.is_active, a.created_at, a.updated_at,
		       COUNT(l.id) AS location_count
		FROM woulder.areas a
		LEFT JOIN woulder.locations l ON l.area_id = a.id
		WHERE a.is_active = TRUE
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
			&a.Region, &a.DisplayOrder, &a.IsActive,
			&a.CreatedAt, &a.UpdatedAt,
			&a.LocationCount,
		); err != nil {
			return nil, err
		}
		areas = append(areas, a)
	}

	return areas, nil
}

// GetAreaByID retrieves a specific active area by ID
func (db *Database) GetAreaByID(ctx context.Context, id int) (*models.Area, error) {
	query := `
		SELECT id, name, description, region,
		       display_order, is_active, created_at, updated_at
		FROM woulder.areas
		WHERE id = $1 AND is_active = TRUE
	`

	var a models.Area
	err := db.conn.QueryRowContext(ctx, query, id).Scan(
		&a.ID, &a.Name, &a.Description,
		&a.Region, &a.DisplayOrder, &a.IsActive,
		&a.CreatedAt, &a.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &a, nil
}

// GetRockTypesByLocation retrieves rock types for a location
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

// GetPrimaryRockType retrieves the primary rock type for a location
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

// GetSunExposureByLocation retrieves sun exposure data for a location
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
