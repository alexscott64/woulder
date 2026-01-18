package database

import (
	"context"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// GetRiversByLocation retrieves all rivers associated with a location
func (db *Database) GetRiversByLocation(ctx context.Context, locationID int) ([]models.River, error) {
	query := `
		SELECT id, location_id, gauge_id, river_name,
		       safe_crossing_cfs, caution_crossing_cfs,
		       drainage_area_sq_mi, gauge_drainage_area_sq_mi,
		       flow_divisor, is_estimated, description,
		       created_at, updated_at
		FROM woulder.rivers
		WHERE location_id = $1
		ORDER BY river_name
	`

	rows, err := db.conn.QueryContext(ctx, query, locationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rivers []models.River
	for rows.Next() {
		var r models.River
		if err := rows.Scan(
			&r.ID, &r.LocationID, &r.GaugeID, &r.RiverName,
			&r.SafeCrossingCFS, &r.CautionCrossingCFS,
			&r.DrainageAreaSqMi, &r.GaugeDrainageAreaSqMi,
			&r.FlowDivisor, &r.IsEstimated, &r.Description,
			&r.CreatedAt, &r.UpdatedAt,
		); err != nil {
			return nil, err
		}
		rivers = append(rivers, r)
	}

	return rivers, nil
}

// GetRiverByID retrieves a specific river by ID
func (db *Database) GetRiverByID(ctx context.Context, id int) (*models.River, error) {
	query := `
		SELECT id, location_id, gauge_id, river_name,
		       safe_crossing_cfs, caution_crossing_cfs,
		       drainage_area_sq_mi, gauge_drainage_area_sq_mi,
		       flow_divisor, is_estimated, description,
		       created_at, updated_at
		FROM woulder.rivers
		WHERE id = $1
	`

	var r models.River
	err := db.conn.QueryRowContext(ctx, query, id).Scan(
		&r.ID, &r.LocationID, &r.GaugeID, &r.RiverName,
		&r.SafeCrossingCFS, &r.CautionCrossingCFS,
		&r.DrainageAreaSqMi, &r.GaugeDrainageAreaSqMi,
		&r.FlowDivisor, &r.IsEstimated, &r.Description,
		&r.CreatedAt, &r.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &r, nil
}
