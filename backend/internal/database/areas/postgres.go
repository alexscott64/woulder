package areas

import (
	"context"

	"github.com/alexscott64/woulder/backend/internal/database/dberrors"
	"github.com/alexscott64/woulder/backend/internal/models"
)

// PostgresRepository implements Repository using PostgreSQL.
type PostgresRepository struct {
	db DBConn
}

// NewPostgresRepository creates a new PostgreSQL areas repository.
func NewPostgresRepository(db DBConn) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// GetAll retrieves all active areas.
func (r *PostgresRepository) GetAll(ctx context.Context) ([]models.Area, error) {
	rows, err := r.db.QueryContext(ctx, queryGetAll)
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

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return areas, nil
}

// GetAllWithLocationCounts retrieves all areas with their location counts.
func (r *PostgresRepository) GetAllWithLocationCounts(ctx context.Context) ([]models.AreaWithLocationCount, error) {
	rows, err := r.db.QueryContext(ctx, queryGetAllWithLocationCounts)
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

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return areas, nil
}

// GetByID retrieves a specific active area by ID.
func (r *PostgresRepository) GetByID(ctx context.Context, id int) (*models.Area, error) {
	var a models.Area
	err := r.db.QueryRowContext(ctx, queryGetByID, id).Scan(
		&a.ID, &a.Name, &a.Description,
		&a.Region, &a.DisplayOrder, &a.IsActive,
		&a.CreatedAt, &a.UpdatedAt,
	)

	if err != nil {
		return nil, dberrors.WrapNotFound(err)
	}

	return &a, nil
}
