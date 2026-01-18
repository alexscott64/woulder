package database

import (
	"context"
	"database/sql"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// GetAreasOrderedByActivity retrieves top-level areas ordered by most recent climb activity
// Shows children of the root area with aggregated activity from all descendant areas
// Uses the same smart date filtering as GetClimbHistoryForLocation
func (db *Database) GetAreasOrderedByActivity(ctx context.Context, locationID int) ([]models.AreaActivitySummary, error) {
	query := `
		WITH RECURSIVE adjusted_ticks AS (
			SELECT
				t.mp_route_id,
				CASE
					WHEN t.climbed_at > NOW() + INTERVAL '350 days'
					     AND t.climbed_at < NOW() + INTERVAL '380 days'
					THEN t.climbed_at - INTERVAL '1 year'
					ELSE t.climbed_at
				END AS adjusted_climbed_at
			FROM woulder.mp_ticks t
			WHERE
				t.climbed_at <= NOW() + INTERVAL '30 days'
				AND t.climbed_at >= NOW() - INTERVAL '2 years'
		),
		root_area AS (
			SELECT mp_area_id
			FROM woulder.mp_areas
			WHERE location_id = $1 AND parent_mp_area_id IS NULL
			LIMIT 1
		),
		top_level_areas AS (
			SELECT mp_area_id, name, parent_mp_area_id
			FROM woulder.mp_areas
			WHERE location_id = $1
			  AND parent_mp_area_id IN (SELECT mp_area_id FROM root_area)
		),
		area_tree AS (
			-- Start with top-level areas
			SELECT mp_area_id, mp_area_id as top_level_id
			FROM top_level_areas

			UNION ALL

			-- Recursively include all descendant areas
			SELECT a.mp_area_id, at.top_level_id
			FROM woulder.mp_areas a
			INNER JOIN area_tree at ON a.parent_mp_area_id = at.mp_area_id
			WHERE a.location_id = $1
		)
		SELECT
			tla.mp_area_id,
			tla.name,
			tla.parent_mp_area_id,
			MAX(adj.adjusted_climbed_at) AS last_climb_at,
			COUNT(DISTINCT r.mp_route_id) AS unique_routes,
			COUNT(*)::int AS total_ticks,
			EXTRACT(DAY FROM (NOW() - MAX(adj.adjusted_climbed_at)))::int AS days_since_climb,
			EXISTS(SELECT 1 FROM woulder.mp_areas sub WHERE sub.parent_mp_area_id = tla.mp_area_id) AS has_subareas,
			(SELECT COUNT(*)::int FROM woulder.mp_areas sub WHERE sub.parent_mp_area_id = tla.mp_area_id) AS subarea_count
		FROM top_level_areas tla
		INNER JOIN area_tree atree ON tla.mp_area_id = atree.top_level_id
		INNER JOIN woulder.mp_routes r ON atree.mp_area_id = r.mp_area_id
		INNER JOIN adjusted_ticks adj ON r.mp_route_id = adj.mp_route_id
		GROUP BY tla.mp_area_id, tla.name, tla.parent_mp_area_id
		HAVING MAX(adj.adjusted_climbed_at) IS NOT NULL
		ORDER BY MAX(adj.adjusted_climbed_at) DESC
	`

	rows, err := db.conn.QueryContext(ctx, query, locationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var areas []models.AreaActivitySummary
	for rows.Next() {
		var area models.AreaActivitySummary
		err := rows.Scan(
			&area.MPAreaID,
			&area.Name,
			&area.ParentMPAreaID,
			&area.LastClimbAt,
			&area.UniqueRoutes,
			&area.TotalTicks,
			&area.DaysSinceClimb,
			&area.HasSubareas,
			&area.SubareaCount,
		)
		if err != nil {
			return nil, err
		}
		areas = append(areas, area)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return areas, nil
}

// GetSubareasOrderedByActivity retrieves subareas of a parent area ordered by most recent climb activity
// Recursively aggregates activity from all descendant areas
// Uses the same smart date filtering as GetClimbHistoryForLocation
func (db *Database) GetSubareasOrderedByActivity(ctx context.Context, parentAreaID string, locationID int) ([]models.AreaActivitySummary, error) {
	query := `
		WITH RECURSIVE adjusted_ticks AS (
			SELECT
				t.mp_route_id,
				CASE
					WHEN t.climbed_at > NOW() + INTERVAL '350 days'
					     AND t.climbed_at < NOW() + INTERVAL '380 days'
					THEN t.climbed_at - INTERVAL '1 year'
					ELSE t.climbed_at
				END AS adjusted_climbed_at
			FROM woulder.mp_ticks t
			WHERE
				t.climbed_at <= NOW() + INTERVAL '30 days'
				AND t.climbed_at >= NOW() - INTERVAL '2 years'
		),
		direct_subareas AS (
			SELECT mp_area_id, name, parent_mp_area_id
			FROM woulder.mp_areas
			WHERE parent_mp_area_id = $1
			  AND location_id = $2
		),
		area_tree AS (
			-- Start with direct subareas
			SELECT mp_area_id, mp_area_id as subarea_id
			FROM direct_subareas

			UNION ALL

			-- Recursively include all descendant areas
			SELECT a.mp_area_id, at.subarea_id
			FROM woulder.mp_areas a
			INNER JOIN area_tree at ON a.parent_mp_area_id = at.mp_area_id
			WHERE a.location_id = $2
		)
		SELECT
			sa.mp_area_id,
			sa.name,
			sa.parent_mp_area_id,
			COALESCE(MAX(adj.adjusted_climbed_at), NOW() - INTERVAL '10 years') AS last_climb_at,
			COUNT(DISTINCT r.mp_route_id) AS unique_routes,
			COALESCE(COUNT(adj.mp_route_id), 0)::int AS total_ticks,
			COALESCE(EXTRACT(DAY FROM (NOW() - MAX(adj.adjusted_climbed_at)))::int, 3650) AS days_since_climb,
			EXISTS(SELECT 1 FROM woulder.mp_areas sub WHERE sub.parent_mp_area_id = sa.mp_area_id) AS has_subareas,
			(SELECT COUNT(*)::int FROM woulder.mp_areas sub WHERE sub.parent_mp_area_id = sa.mp_area_id) AS subarea_count
		FROM direct_subareas sa
		LEFT JOIN area_tree atree ON sa.mp_area_id = atree.subarea_id
		LEFT JOIN woulder.mp_routes r ON atree.mp_area_id = r.mp_area_id
		LEFT JOIN adjusted_ticks adj ON r.mp_route_id = adj.mp_route_id
		GROUP BY sa.mp_area_id, sa.name, sa.parent_mp_area_id
		ORDER BY MAX(adj.adjusted_climbed_at) DESC NULLS LAST
	`

	rows, err := db.conn.QueryContext(ctx, query, parentAreaID, locationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var areas []models.AreaActivitySummary
	for rows.Next() {
		var area models.AreaActivitySummary
		err := rows.Scan(
			&area.MPAreaID,
			&area.Name,
			&area.ParentMPAreaID,
			&area.LastClimbAt,
			&area.UniqueRoutes,
			&area.TotalTicks,
			&area.DaysSinceClimb,
			&area.HasSubareas,
			&area.SubareaCount,
		)
		if err != nil {
			return nil, err
		}
		areas = append(areas, area)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return areas, nil
}

// GetRoutesOrderedByActivity retrieves ALL routes in an area ordered by most recent climb activity
// Shows routes with ticks first (ordered by recency), then routes without ticks (alphabetically)
// Includes the most recent tick for each route if it has any
// Uses the same smart date filtering as GetClimbHistoryForLocation
func (db *Database) GetRoutesOrderedByActivity(ctx context.Context, areaID string, locationID int, limit int) ([]models.RouteActivitySummary, error) {
	query := `
		WITH adjusted_ticks AS (
			SELECT
				t.mp_route_id,
				t.user_name,
				CASE
					WHEN t.climbed_at > NOW() + INTERVAL '350 days'
					     AND t.climbed_at < NOW() + INTERVAL '380 days'
					THEN t.climbed_at - INTERVAL '1 year'
					ELSE t.climbed_at
				END AS adjusted_climbed_at,
				t.style,
				t.comment,
				ROW_NUMBER() OVER (PARTITION BY t.mp_route_id ORDER BY
					CASE
						WHEN t.climbed_at > NOW() + INTERVAL '350 days'
						     AND t.climbed_at < NOW() + INTERVAL '380 days'
						THEN t.climbed_at - INTERVAL '1 year'
						ELSE t.climbed_at
					END DESC) AS tick_rank
			FROM woulder.mp_ticks t
			WHERE
				t.climbed_at <= NOW() + INTERVAL '30 days'
				AND t.climbed_at >= NOW() - INTERVAL '2 years'
		)
		SELECT
			r.mp_route_id,
			r.name,
			r.rating,
			r.mp_area_id,
			COALESCE(MAX(at.adjusted_climbed_at), NOW() - INTERVAL '100 years') AS last_climb_at,
			COALESCE(EXTRACT(DAY FROM (NOW() - MAX(at.adjusted_climbed_at)))::int, 36500) AS days_since_climb,
			at.user_name,
			at.adjusted_climbed_at,
			at.style,
			at.comment,
			a.name AS area_name,
			CASE WHEN MAX(at.adjusted_climbed_at) IS NULL THEN 1 ELSE 0 END AS no_ticks
		FROM woulder.mp_routes r
		LEFT JOIN adjusted_ticks at ON r.mp_route_id = at.mp_route_id AND at.tick_rank = 1
		INNER JOIN woulder.mp_areas a ON r.mp_area_id = a.mp_area_id
		WHERE r.mp_area_id = $1
		  AND r.location_id = $2
		GROUP BY r.mp_route_id, r.name, r.rating, r.mp_area_id, a.name, at.user_name, at.adjusted_climbed_at, at.style, at.comment
		ORDER BY no_ticks ASC, MAX(at.adjusted_climbed_at) DESC NULLS LAST, r.name ASC
		LIMIT $3
	`

	rows, err := db.conn.QueryContext(ctx, query, areaID, locationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routes []models.RouteActivitySummary
	for rows.Next() {
		var route models.RouteActivitySummary
		var climbedBy, style, comment, areaName sql.NullString
		var climbedAt sql.NullTime
		var noTicks int

		err := rows.Scan(
			&route.MPRouteID,
			&route.Name,
			&route.Rating,
			&route.MPAreaID,
			&route.LastClimbAt,
			&route.DaysSinceClimb,
			&climbedBy,
			&climbedAt,
			&style,
			&comment,
			&areaName,
			&noTicks,
		)
		if err != nil {
			return nil, err
		}

		// Only populate most recent tick if the route has been climbed
		if climbedAt.Valid {
			var mostRecentTick models.ClimbHistoryEntry
			mostRecentTick.MPRouteID = route.MPRouteID
			mostRecentTick.RouteName = route.Name
			mostRecentTick.RouteRating = route.Rating
			mostRecentTick.MPAreaID = route.MPAreaID
			mostRecentTick.ClimbedBy = climbedBy.String
			mostRecentTick.ClimbedAt = climbedAt.Time
			mostRecentTick.Style = style.String
			if comment.Valid {
				mostRecentTick.Comment = &comment.String
			}
			mostRecentTick.AreaName = areaName.String
			mostRecentTick.DaysSinceClimb = route.DaysSinceClimb

			route.MostRecentTick = mostRecentTick
		}

		routes = append(routes, route)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return routes, nil
}

// GetRecentTicksForRoute retrieves the most recent ticks for a specific route
func (db *Database) GetRecentTicksForRoute(ctx context.Context, routeID string, limit int) ([]models.ClimbHistoryEntry, error) {
	query := `
		WITH adjusted_ticks AS (
			SELECT
				t.mp_route_id,
				t.user_name,
				CASE
					WHEN t.climbed_at > NOW() + INTERVAL '350 days'
					     AND t.climbed_at < NOW() + INTERVAL '380 days'
					THEN t.climbed_at - INTERVAL '1 year'
					ELSE t.climbed_at
				END AS adjusted_climbed_at,
				t.style,
				t.comment
			FROM woulder.mp_ticks t
			WHERE t.mp_route_id = $1
			  AND t.climbed_at <= NOW() + INTERVAL '30 days'
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
		INNER JOIN woulder.mp_routes r ON at.mp_route_id = r.mp_route_id
		INNER JOIN woulder.mp_areas a ON r.mp_area_id = a.mp_area_id
		ORDER BY at.adjusted_climbed_at DESC
		LIMIT $2
	`

	rows, err := db.conn.QueryContext(ctx, query, routeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ticks []models.ClimbHistoryEntry
	for rows.Next() {
		var tick models.ClimbHistoryEntry
		var comment sql.NullString

		err := rows.Scan(
			&tick.MPRouteID,
			&tick.RouteName,
			&tick.RouteRating,
			&tick.MPAreaID,
			&tick.AreaName,
			&tick.ClimbedAt,
			&tick.ClimbedBy,
			&tick.Style,
			&comment,
			&tick.DaysSinceClimb,
		)
		if err != nil {
			return nil, err
		}

		if comment.Valid {
			tick.Comment = &comment.String
		}

		ticks = append(ticks, tick)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ticks, nil
}

