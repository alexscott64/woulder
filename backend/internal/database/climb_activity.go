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
		root_areas AS (
			-- Find "virtual root" areas for this location:
			-- Areas that have this location_id but their parent doesn't (or parent is NULL)
			SELECT a.mp_area_id
			FROM woulder.mp_areas a
			LEFT JOIN woulder.mp_areas parent ON a.parent_mp_area_id = parent.mp_area_id
			WHERE a.location_id = $1
			  AND (a.parent_mp_area_id IS NULL OR parent.location_id IS NULL OR parent.location_id != $1)
		),
		top_level_areas AS (
			-- If there's a single root area, show its children
			-- If there are multiple root areas, show them directly
			SELECT mp_area_id, name, parent_mp_area_id
			FROM woulder.mp_areas
			WHERE location_id = $1
			  AND (
				(parent_mp_area_id IN (SELECT mp_area_id FROM root_areas) AND (SELECT COUNT(*) FROM root_areas) = 1)
				OR (mp_area_id IN (SELECT mp_area_id FROM root_areas) AND (SELECT COUNT(*) FROM root_areas) > 1)
			  )
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
func (db *Database) GetSubareasOrderedByActivity(ctx context.Context, parentAreaID int64, locationID int) ([]models.AreaActivitySummary, error) {
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
func (db *Database) GetRoutesOrderedByActivity(ctx context.Context, areaID int64, locationID int, limit int) ([]models.RouteActivitySummary, error) {
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
			mostRecentTick := &models.ClimbHistoryEntry{
				MPRouteID:      route.MPRouteID,
				RouteName:      route.Name,
				RouteRating:    route.Rating,
				MPAreaID:       route.MPAreaID,
				ClimbedBy:      climbedBy.String,
				ClimbedAt:      climbedAt.Time,
				Style:          style.String,
				AreaName:       areaName.String,
				DaysSinceClimb: route.DaysSinceClimb,
			}
			if comment.Valid {
				mostRecentTick.Comment = &comment.String
			}
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
func (db *Database) GetRecentTicksForRoute(ctx context.Context, routeID int64, limit int) ([]models.ClimbHistoryEntry, error) {
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

// SearchInLocation searches all areas and routes in a location by name
// Returns unified search results (both areas and routes) ordered by most recent climb activity
func (db *Database) SearchInLocation(ctx context.Context, locationID int, searchQuery string, limit int) ([]models.SearchResult, error) {
	query := `
		WITH location_routes AS (
			-- Filter routes by location first
			SELECT r.mp_route_id, r.name, r.rating, r.mp_area_id
			FROM woulder.mp_routes r
			WHERE r.location_id = $1
		),
		adjusted_ticks AS (
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
			-- Only process ticks for routes in this location
			INNER JOIN location_routes lr ON t.mp_route_id = lr.mp_route_id
			WHERE
				t.climbed_at <= NOW() + INTERVAL '30 days'
				AND t.climbed_at >= NOW() - INTERVAL '2 years'
		),
		-- Areas with activity
		area_results AS (
			SELECT
				'area' AS result_type,
				a.mp_area_id AS id,
				a.name,
				NULL::text AS rating,
				a.mp_area_id,
				NULL::text AS area_name,
				COALESCE(MAX(at.adjusted_climbed_at), NOW() - INTERVAL '100 years') AS last_climb_at,
				COALESCE(EXTRACT(DAY FROM (NOW() - MAX(at.adjusted_climbed_at)))::int, 36500) AS days_since_climb,
				COUNT(DISTINCT lr.mp_route_id)::int AS total_ticks,
				COUNT(DISTINCT lr.mp_route_id)::int AS unique_routes,
				NULL::text AS user_name,
				NULL::timestamp AS tick_climbed_at,
				NULL::text AS style,
				NULL::text AS comment,
				CASE WHEN MAX(at.adjusted_climbed_at) IS NULL THEN 1 ELSE 0 END AS no_ticks
			FROM woulder.mp_areas a
			INNER JOIN location_routes lr ON a.mp_area_id = lr.mp_area_id
			LEFT JOIN adjusted_ticks at ON lr.mp_route_id = at.mp_route_id
			WHERE LOWER(a.name) LIKE LOWER($2)
			GROUP BY a.mp_area_id, a.name
		),
		-- Routes with activity
		route_results AS (
			SELECT
				'route' AS result_type,
				lr.mp_route_id AS id,
				lr.name,
				lr.rating,
				lr.mp_area_id,
				a.name AS area_name,
				COALESCE(MAX(at.adjusted_climbed_at), NOW() - INTERVAL '100 years') AS last_climb_at,
				COALESCE(EXTRACT(DAY FROM (NOW() - MAX(at.adjusted_climbed_at)))::int, 36500) AS days_since_climb,
				NULL::int AS total_ticks,
				NULL::int AS unique_routes,
				at.user_name,
				at.adjusted_climbed_at AS tick_climbed_at,
				at.style,
				at.comment,
				CASE WHEN MAX(at.adjusted_climbed_at) IS NULL THEN 1 ELSE 0 END AS no_ticks
			FROM location_routes lr
			LEFT JOIN adjusted_ticks at ON lr.mp_route_id = at.mp_route_id AND at.tick_rank = 1
			INNER JOIN woulder.mp_areas a ON lr.mp_area_id = a.mp_area_id
			WHERE (LOWER(lr.name) LIKE LOWER($2) OR LOWER(lr.rating) LIKE LOWER($2) OR LOWER(a.name) LIKE LOWER($2))
			GROUP BY lr.mp_route_id, lr.name, lr.rating, lr.mp_area_id, a.name, at.user_name, at.adjusted_climbed_at, at.style, at.comment
		)
		-- Combine results
		SELECT * FROM area_results
		UNION ALL
		SELECT * FROM route_results
		ORDER BY no_ticks ASC, last_climb_at DESC NULLS LAST, name ASC
		LIMIT $3
	`

	searchPattern := "%" + searchQuery + "%"
	rows, err := db.conn.QueryContext(ctx, query, locationID, searchPattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.SearchResult
	for rows.Next() {
		var result models.SearchResult
		var rating, areaName, climbedBy, style, comment sql.NullString
		var totalTicks, uniqueRoutes sql.NullInt64
		var climbedAt sql.NullTime
		var noTicks int

		err := rows.Scan(
			&result.ResultType,
			&result.ID,
			&result.Name,
			&rating,
			&result.MPAreaID,
			&areaName,
			&result.LastClimbAt,
			&result.DaysSinceClimb,
			&totalTicks,
			&uniqueRoutes,
			&climbedBy,
			&climbedAt,
			&style,
			&comment,
			&noTicks,
		)
		if err != nil {
			return nil, err
		}

		// Populate optional fields based on result type
		if rating.Valid {
			result.Rating = &rating.String
		}
		if areaName.Valid {
			result.AreaName = &areaName.String
		}
		if totalTicks.Valid {
			ticks := int(totalTicks.Int64)
			result.TotalTicks = &ticks
		}
		if uniqueRoutes.Valid {
			routes := int(uniqueRoutes.Int64)
			result.UniqueRoutes = &routes
		}

		// Only populate most recent tick for routes with activity
		if result.ResultType == "route" && climbedAt.Valid {
			mostRecentTick := &models.ClimbHistoryEntry{
				MPRouteID:      result.ID,
				RouteName:      result.Name,
				RouteRating:    *result.Rating,
				MPAreaID:       result.MPAreaID,
				ClimbedBy:      climbedBy.String,
				ClimbedAt:      climbedAt.Time,
				Style:          style.String,
				AreaName:       *result.AreaName,
				DaysSinceClimb: result.DaysSinceClimb,
			}
			if comment.Valid {
				mostRecentTick.Comment = &comment.String
			}
			result.MostRecentTick = mostRecentTick
		}

		results = append(results, result)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// SearchRoutesInLocation searches all routes in a location by name or grade
// Returns routes ordered by most recent climb activity
func (db *Database) SearchRoutesInLocation(ctx context.Context, locationID int, searchQuery string, limit int) ([]models.RouteActivitySummary, error) {
	query := `
		WITH location_routes AS (
			-- Filter routes by location and search query first
			SELECT r.mp_route_id, r.name, r.rating, r.mp_area_id, a.name AS area_name
			FROM woulder.mp_routes r
			INNER JOIN woulder.mp_areas a ON r.mp_area_id = a.mp_area_id
			WHERE r.location_id = $1
			  AND (LOWER(r.name) LIKE LOWER($2) OR LOWER(r.rating) LIKE LOWER($2) OR LOWER(a.name) LIKE LOWER($2))
		),
		adjusted_ticks AS (
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
			-- Only process ticks for routes in this location
			INNER JOIN location_routes lr ON t.mp_route_id = lr.mp_route_id
			WHERE
				t.climbed_at <= NOW() + INTERVAL '30 days'
				AND t.climbed_at >= NOW() - INTERVAL '2 years'
		)
		SELECT
			lr.mp_route_id,
			lr.name,
			lr.rating,
			lr.mp_area_id,
			COALESCE(MAX(at.adjusted_climbed_at), NOW() - INTERVAL '100 years') AS last_climb_at,
			COALESCE(EXTRACT(DAY FROM (NOW() - MAX(at.adjusted_climbed_at)))::int, 36500) AS days_since_climb,
			at.user_name,
			at.adjusted_climbed_at,
			at.style,
			at.comment,
			lr.area_name,
			CASE WHEN MAX(at.adjusted_climbed_at) IS NULL THEN 1 ELSE 0 END AS no_ticks
		FROM location_routes lr
		LEFT JOIN adjusted_ticks at ON lr.mp_route_id = at.mp_route_id AND at.tick_rank = 1
		GROUP BY lr.mp_route_id, lr.name, lr.rating, lr.mp_area_id, lr.area_name, at.user_name, at.adjusted_climbed_at, at.style, at.comment
		ORDER BY no_ticks ASC, MAX(at.adjusted_climbed_at) DESC NULLS LAST, lr.name ASC
		LIMIT $3
	`

	searchPattern := "%" + searchQuery + "%"
	rows, err := db.conn.QueryContext(ctx, query, locationID, searchPattern, limit)
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
			mostRecentTick := &models.ClimbHistoryEntry{
				MPRouteID:      route.MPRouteID,
				RouteName:      route.Name,
				RouteRating:    route.Rating,
				MPAreaID:       route.MPAreaID,
				ClimbedBy:      climbedBy.String,
				ClimbedAt:      climbedAt.Time,
				Style:          style.String,
				AreaName:       areaName.String,
				DaysSinceClimb: route.DaysSinceClimb,
			}
			if comment.Valid {
				mostRecentTick.Comment = &comment.String
			}
			route.MostRecentTick = mostRecentTick
		}

		routes = append(routes, route)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return routes, nil
}

