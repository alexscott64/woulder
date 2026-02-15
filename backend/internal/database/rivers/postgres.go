package rivers

import (
	"context"

	"github.com/alexscott64/woulder/backend/internal/database/dberrors"
	"github.com/alexscott64/woulder/backend/internal/models"
)

// PostgresRepository implements Repository using PostgreSQL.
type PostgresRepository struct {
	db DBConn
}

// NewPostgresRepository creates a new PostgreSQL river repository.
// The db parameter can be either *sql.DB or *sql.Tx, allowing this
// repository to work both standalone and within transactions.
func NewPostgresRepository(db DBConn) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// GetByLocation retrieves all rivers associated with a location.
func (r *PostgresRepository) GetByLocation(ctx context.Context, locationID int) ([]models.River, error) {
	rows, err := r.db.QueryContext(ctx, queryGetByLocation, locationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rivers []models.River
	for rows.Next() {
		var river models.River
		err := rows.Scan(
			&river.ID,
			&river.LocationID,
			&river.GaugeID,
			&river.RiverName,
			&river.SafeCrossingCFS,
			&river.CautionCrossingCFS,
			&river.DrainageAreaSqMi,
			&river.GaugeDrainageAreaSqMi,
			&river.FlowDivisor,
			&river.IsEstimated,
			&river.Description,
			&river.CreatedAt,
			&river.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		rivers = append(rivers, river)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return rivers, nil
}

// GetByID retrieves a specific river by its ID.
func (r *PostgresRepository) GetByID(ctx context.Context, id int) (*models.River, error) {
	var river models.River
	err := r.db.QueryRowContext(ctx, queryGetByID, id).Scan(
		&river.ID,
		&river.LocationID,
		&river.GaugeID,
		&river.RiverName,
		&river.SafeCrossingCFS,
		&river.CautionCrossingCFS,
		&river.DrainageAreaSqMi,
		&river.GaugeDrainageAreaSqMi,
		&river.FlowDivisor,
		&river.IsEstimated,
		&river.Description,
		&river.CreatedAt,
		&river.UpdatedAt,
	)

	if err != nil {
		return nil, dberrors.WrapNotFound(err)
	}

	return &river, nil
}
