package database

import (
	"context"
	"database/sql"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// GetLastClimbedForLocation retrieves the most recent climb for a location
func (db *Database) GetLastClimbedForLocation(ctx context.Context, locationID int) (*models.LastClimbedInfo, error) {
	query := `
		SELECT
			r.name AS route_name,
			r.rating AS route_rating,
			t.climbed_at,
			t.user_name AS climbed_by,
			t.style,
			t.comment,
			EXTRACT(DAY FROM (NOW() - t.climbed_at))::int AS days_since_climb
		FROM woulder.mp_ticks t
		JOIN woulder.mp_routes r ON t.mp_route_id = r.mp_route_id
		WHERE r.location_id = $1
		ORDER BY t.climbed_at DESC
		LIMIT 1
	`

	var info models.LastClimbedInfo
	err := db.conn.QueryRowContext(ctx, query, locationID).Scan(
		&info.RouteName,
		&info.RouteRating,
		&info.ClimbedAt,
		&info.ClimbedBy,
		&info.Style,
		&info.Comment,
		&info.DaysSinceClimb,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No climb data for this location
	}

	if err != nil {
		return nil, err
	}

	return &info, nil
}

// GetClimbHistoryForLocation retrieves recent climb history for a location with area information
// Includes smart data quality filtering:
// - Adjusts dates that are 1 year in the future (likely typo: 2026 -> 2025)
// - Filters out dates more than 30 days in the future (intentionally bad data)
// - Only shows climbs from the past 2 years (excludes very old or very future dates)
func (db *Database) GetClimbHistoryForLocation(ctx context.Context, locationID int, limit int) ([]models.ClimbHistoryEntry, error) {
	query := `
		WITH adjusted_ticks AS (
			SELECT
				t.mp_route_id,
				t.user_name,
				-- Smart date adjustment: if date is 350-380 days in future, subtract 1 year
				-- This catches typos like "2026" when they meant "2025"
				CASE
					WHEN t.climbed_at > NOW() + INTERVAL '350 days'
					     AND t.climbed_at < NOW() + INTERVAL '380 days'
					THEN t.climbed_at - INTERVAL '1 year'
					ELSE t.climbed_at
				END AS adjusted_climbed_at,
				t.style,
				t.comment
			FROM woulder.mp_ticks t
			WHERE
				-- Filter out dates more than 30 days in the future (bad data)
				t.climbed_at <= NOW() + INTERVAL '30 days'
				-- Filter out climbs older than 2 years (keep it recent)
				AND t.climbed_at >= NOW() - INTERVAL '2 years'
		)
		SELECT
			r.mp_route_id,
			r.name AS route_name,
			r.rating AS route_rating,
			r.mp_area_id,
			a.name AS area_name,
			at.adjusted_climbed_at AS climbed_at,
			at.user_name AS climbed_by,
			at.style,
			at.comment,
			EXTRACT(DAY FROM (NOW() - at.adjusted_climbed_at))::int AS days_since_climb
		FROM adjusted_ticks at
		JOIN woulder.mp_routes r ON at.mp_route_id = r.mp_route_id
		JOIN woulder.mp_areas a ON r.mp_area_id = a.mp_area_id
		WHERE r.location_id = $1
		ORDER BY at.adjusted_climbed_at DESC
		LIMIT $2
	`

	rows, err := db.conn.QueryContext(ctx, query, locationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []models.ClimbHistoryEntry
	for rows.Next() {
		var entry models.ClimbHistoryEntry
		var climbedBy, style sql.NullString
		var comment sql.NullString

		err := rows.Scan(
			&entry.MPRouteID,
			&entry.RouteName,
			&entry.RouteRating,
			&entry.MPAreaID,
			&entry.AreaName,
			&entry.ClimbedAt,
			&climbedBy,
			&style,
			&comment,
			&entry.DaysSinceClimb,
		)
		if err != nil {
			return nil, err
		}

		// Handle nullable fields
		entry.ClimbedBy = climbedBy.String // Will be empty string if NULL
		entry.Style = style.String         // Will be empty string if NULL
		if comment.Valid {
			entry.Comment = &comment.String
		}

		history = append(history, entry)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return history, nil
}
