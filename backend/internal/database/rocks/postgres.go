package rocks

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/alexscott64/woulder/backend/internal/database"
	"github.com/alexscott64/woulder/backend/internal/models"
)

// PostgresRepository implements Repository using PostgreSQL.
type PostgresRepository struct {
	db database.DBConn
}

// NewPostgresRepository creates a new PostgreSQL rocks repository.
func NewPostgresRepository(db database.DBConn) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// GetRockTypesByLocation retrieves all rock types for a location.
func (r *PostgresRepository) GetRockTypesByLocation(ctx context.Context, locationID int) ([]models.RockType, error) {
	rows, err := r.db.QueryContext(ctx, queryGetRockTypesByLocation, locationID)
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

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return rockTypes, nil
}

// GetPrimaryRockType retrieves the primary rock type for a location.
func (r *PostgresRepository) GetPrimaryRockType(ctx context.Context, locationID int) (*models.RockType, error) {
	var rt models.RockType
	err := r.db.QueryRowContext(ctx, queryGetPrimaryRockType, locationID).Scan(
		&rt.ID, &rt.Name, &rt.BaseDryingHours,
		&rt.PorosityPercent, &rt.IsWetSensitive,
		&rt.Description, &rt.RockTypeGroupID,
		&rt.GroupName,
	)

	if err == sql.ErrNoRows {
		// Fallback: if no primary is set, get the first rock type
		rocks, err := r.GetRockTypesByLocation(ctx, locationID)
		if err != nil || len(rocks) == 0 {
			return nil, fmt.Errorf("no rock types found for location %d: %w", locationID, database.ErrNotFound)
		}
		return &rocks[0], nil
	}

	if err != nil {
		return nil, database.WrapNotFound(err)
	}

	return &rt, nil
}

// GetSunExposureByLocation retrieves sun exposure data for a location.
func (r *PostgresRepository) GetSunExposureByLocation(ctx context.Context, locationID int) (*models.LocationSunExposure, error) {
	var se models.LocationSunExposure
	err := r.db.QueryRowContext(ctx, queryGetSunExposureByLocation, locationID).Scan(
		&se.ID, &se.LocationID,
		&se.SouthFacingPercent, &se.WestFacingPercent,
		&se.EastFacingPercent, &se.NorthFacingPercent,
		&se.SlabPercent, &se.OverhangPercent,
		&se.TreeCoveragePercent, &se.Description,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No sun exposure data is not an error
	}

	if err != nil {
		return nil, err
	}

	return &se, nil
}
