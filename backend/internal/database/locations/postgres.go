package locations

import (
	"context"

	"github.com/alexscott64/woulder/backend/internal/database/dberrors"
	"github.com/alexscott64/woulder/backend/internal/models"
)

// PostgresRepository implements Repository using PostgreSQL.
type PostgresRepository struct {
	db DBConn
}

// NewPostgresRepository creates a new PostgreSQL locations repository.
func NewPostgresRepository(db DBConn) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// GetAll retrieves all locations.
func (r *PostgresRepository) GetAll(ctx context.Context) ([]models.Location, error) {
	rows, err := r.db.QueryContext(ctx, queryGetAll)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []models.Location
	for rows.Next() {
		var loc models.Location
		if err := rows.Scan(
			&loc.ID,
			&loc.Name,
			&loc.Latitude,
			&loc.Longitude,
			&loc.ElevationFt,
			&loc.AreaID,
			&loc.HasSeepageRisk,
			&loc.CreatedAt,
			&loc.UpdatedAt,
		); err != nil {
			return nil, err
		}
		locations = append(locations, loc)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return locations, nil
}

// GetByID retrieves a specific location by ID.
func (r *PostgresRepository) GetByID(ctx context.Context, id int) (*models.Location, error) {
	var loc models.Location
	err := r.db.QueryRowContext(ctx, queryGetByID, id).Scan(
		&loc.ID,
		&loc.Name,
		&loc.Latitude,
		&loc.Longitude,
		&loc.ElevationFt,
		&loc.AreaID,
		&loc.HasSeepageRisk,
		&loc.CreatedAt,
		&loc.UpdatedAt,
	)

	if err != nil {
		return nil, dberrors.WrapNotFound(err)
	}

	return &loc, nil
}

// GetByArea retrieves all locations in a specific area.
func (r *PostgresRepository) GetByArea(ctx context.Context, areaID int) ([]models.Location, error) {
	rows, err := r.db.QueryContext(ctx, queryGetByArea, areaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []models.Location
	for rows.Next() {
		var loc models.Location
		if err := rows.Scan(
			&loc.ID,
			&loc.Name,
			&loc.Latitude,
			&loc.Longitude,
			&loc.ElevationFt,
			&loc.AreaID,
			&loc.HasSeepageRisk,
			&loc.CreatedAt,
			&loc.UpdatedAt,
		); err != nil {
			return nil, err
		}
		locations = append(locations, loc)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return locations, nil
}
