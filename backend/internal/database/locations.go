package database

import (
	"context"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// GetAllLocations retrieves all locations ordered by name
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

// GetLocation retrieves a specific location by ID
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

// GetLocationsByArea retrieves all locations in a specific area
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
