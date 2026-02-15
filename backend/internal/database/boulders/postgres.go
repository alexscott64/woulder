package boulders

import (
	"context"
	"database/sql"

	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/lib/pq"
)

// PostgresRepository implements Repository using PostgreSQL.
type PostgresRepository struct {
	db DBConn
}

// NewPostgresRepository creates a new PostgreSQL boulders repository.
func NewPostgresRepository(db DBConn) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// GetProfile retrieves the drying profile for a specific boulder.
func (r *PostgresRepository) GetProfile(ctx context.Context, mpRouteID int64) (*models.BoulderDryingProfile, error) {
	var profile models.BoulderDryingProfile
	err := r.db.QueryRowContext(ctx, queryGetProfile, mpRouteID).Scan(
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
		return nil, nil // No profile exists - not an error
	}

	if err != nil {
		return nil, err
	}

	return &profile, nil
}

// GetProfilesByIDs retrieves multiple boulder drying profiles in a single query.
func (r *PostgresRepository) GetProfilesByIDs(ctx context.Context, mpRouteIDs []int64) (map[int64]*models.BoulderDryingProfile, error) {
	if len(mpRouteIDs) == 0 {
		return make(map[int64]*models.BoulderDryingProfile), nil
	}

	rows, err := r.db.QueryContext(ctx, queryGetProfilesByIDs, pq.Array(mpRouteIDs))
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

// SaveProfile creates or updates a boulder drying profile.
func (r *PostgresRepository) SaveProfile(ctx context.Context, profile *models.BoulderDryingProfile) error {
	_, err := r.db.ExecContext(ctx, querySaveProfile,
		profile.MPRouteID,
		profile.TreeCoveragePercent,
		profile.RockTypeOverride,
		profile.LastSunCalcAt,
		profile.SunExposureHoursCache,
	)

	return err
}
