package heatmap

// SQL queries for heat map and activity data operations.
// Heat map queries are performance-critical for clustering and must be optimized.

const (
	// queryHeatMapDataLightweight retrieves minimal data for clustering performance.
	// Active routes and has_subareas are not calculated to reduce query complexity.
	// Used for initial map loads and low zoom levels where speed is critical.
	queryHeatMapDataLightweight = `
		SELECT
			a.mp_area_id,
			a.name,
			a.latitude,
			a.longitude,
			0 as active_routes,  -- Not calculated in lightweight mode
			COUNT(t.id) as total_ticks,
			MAX(t.climbed_at) as last_activity,
			COUNT(DISTINCT CASE WHEN t.user_name IS NOT NULL THEN t.user_name END) as unique_climbers,
			false as has_subareas  -- Not calculated in lightweight mode
		FROM woulder.mp_areas a
		JOIN woulder.mp_routes r ON r.mp_area_id = a.mp_area_id
		JOIN woulder.mp_ticks t ON t.mp_route_id = r.mp_route_id
		WHERE t.climbed_at >= $1
			AND t.climbed_at <= $2
			AND a.latitude IS NOT NULL
			AND a.longitude IS NOT NULL
			AND ($3::float IS NULL OR (
				a.latitude BETWEEN $3 AND $4
				AND a.longitude BETWEEN $5 AND $6
			))
			AND ($7::text[] IS NULL OR r.route_type = ANY($7))
		GROUP BY a.mp_area_id, a.name, a.latitude, a.longitude
		HAVING COUNT(t.id) >= $8
		ORDER BY COUNT(t.id) DESC
		LIMIT $9;
	`

	// queryHeatMapDataFull retrieves complete data with all aggregations.
	// Includes active route counts and subarea detection for detailed clustering.
	// Used for high zoom levels and detailed area views.
	queryHeatMapDataFull = `
		SELECT
			a.mp_area_id,
			a.name,
			a.latitude,
			a.longitude,
			COUNT(DISTINCT t.mp_route_id) as active_routes,
			COUNT(t.id) as total_ticks,
			MAX(t.climbed_at) as last_activity,
			COUNT(DISTINCT t.user_name) as unique_climbers,
			EXISTS(
				SELECT 1 FROM woulder.mp_areas sub
				WHERE sub.parent_mp_area_id = a.mp_area_id
				LIMIT 1
			) as has_subareas
		FROM woulder.mp_areas a
		JOIN woulder.mp_routes r ON r.mp_area_id = a.mp_area_id
		JOIN woulder.mp_ticks t ON t.mp_route_id = r.mp_route_id
		WHERE t.climbed_at >= $1
			AND t.climbed_at <= $2
			AND a.latitude IS NOT NULL
			AND a.longitude IS NOT NULL
			AND ($3::float IS NULL OR (
				a.latitude BETWEEN $3 AND $4
				AND a.longitude BETWEEN $5 AND $6
			))
			AND ($7::text[] IS NULL OR r.route_type = ANY($7))
		GROUP BY a.mp_area_id, a.name, a.latitude, a.longitude
		HAVING COUNT(t.id) >= $8
		ORDER BY COUNT(t.id) DESC
		LIMIT $9;
	`

	// queryAreaInfo retrieves base area information.
	queryAreaInfo = `
		SELECT
			a.mp_area_id,
			a.name,
			a.parent_mp_area_id,
			a.latitude,
			a.longitude
		FROM woulder.mp_areas a
		WHERE a.mp_area_id = $1
	`

	// queryAreaActivityStats retrieves aggregated activity statistics for an area.
	queryAreaActivityStats = `
		SELECT
			COUNT(t.id) as total_ticks,
			COUNT(DISTINCT t.mp_route_id) as active_routes,
			COUNT(DISTINCT CASE WHEN t.user_name IS NOT NULL THEN t.user_name END) as unique_climbers,
			MAX(t.climbed_at) as last_activity
		FROM woulder.mp_ticks t
		JOIN woulder.mp_routes r ON t.mp_route_id = r.mp_route_id
		WHERE r.mp_area_id = $1
			AND t.climbed_at >= $2
			AND t.climbed_at <= $3
	`

	// queryRecentTicks retrieves the most recent ticks for an area.
	queryRecentTicks = `
		SELECT
			t.mp_route_id,
			r.name as route_name,
			COALESCE(r.rating, '') as rating,
			COALESCE(t.user_name, '') as user_name,
			t.climbed_at,
			COALESCE(t.style, '') as style,
			COALESCE(t.comment, '') as comment
		FROM woulder.mp_ticks t
		JOIN woulder.mp_routes r ON t.mp_route_id = r.mp_route_id
		WHERE r.mp_area_id = $1
			AND t.climbed_at >= $2
			AND t.climbed_at <= $3
		ORDER BY t.climbed_at DESC
		LIMIT 20
	`

	// queryRecentComments retrieves recent comments for routes in an area.
	// Strips HTML tags from comment text using regex for clean display.
	queryRecentComments = `
		SELECT
			c.id,
			c.user_name,
			REGEXP_REPLACE(
				REGEXP_REPLACE(c.comment_text, '<[^>]+>', '', 'g'),
				'\s+', ' ', 'g'
			) as comment_text,
			c.commented_at,
			c.mp_route_id,
			r.name as route_name
		FROM woulder.mp_comments c
		JOIN woulder.mp_routes r ON c.mp_route_id = r.mp_route_id
		WHERE r.mp_area_id = $1
			AND c.commented_at >= $2
			AND c.commented_at <= $3
		ORDER BY c.commented_at DESC
		LIMIT 10
	`

	// queryActivityTimeline retrieves daily activity aggregation for an area.
	queryActivityTimeline = `
		SELECT
			DATE(t.climbed_at) as date,
			COUNT(t.id) as tick_count,
			COUNT(DISTINCT t.mp_route_id) as route_count
		FROM woulder.mp_ticks t
		JOIN woulder.mp_routes r ON t.mp_route_id = r.mp_route_id
		WHERE r.mp_area_id = $1
			AND t.climbed_at >= $2
			AND t.climbed_at <= $3
		GROUP BY DATE(t.climbed_at)
		ORDER BY DATE(t.climbed_at) ASC
	`

	// queryTopRoutes retrieves the most active routes in an area.
	queryTopRoutes = `
		SELECT
			r.mp_route_id,
			r.name,
			r.rating,
			COUNT(t.id) as tick_count,
			MAX(t.climbed_at) as last_activity
		FROM woulder.mp_routes r
		JOIN woulder.mp_ticks t ON r.mp_route_id = t.mp_route_id
		WHERE r.mp_area_id = $1
			AND t.climbed_at >= $2
			AND t.climbed_at <= $3
		GROUP BY r.mp_route_id, r.name, r.rating
		ORDER BY COUNT(t.id) DESC
		LIMIT 10
	`

	// queryRoutesByBounds retrieves routes within geographic bounds with activity.
	// Used for precision route-level clustering at high zoom levels.
	queryRoutesByBounds = `
		SELECT
			r.mp_route_id,
			r.name,
			r.rating,
			r.latitude,
			r.longitude,
			COUNT(t.id) as tick_count,
			MAX(t.climbed_at) as last_activity,
			r.mp_area_id,
			a.name as area_name
		FROM woulder.mp_routes r
		JOIN woulder.mp_ticks t ON r.mp_route_id = t.mp_route_id
		JOIN woulder.mp_areas a ON r.mp_area_id = a.mp_area_id
		WHERE r.latitude IS NOT NULL
			AND r.longitude IS NOT NULL
			AND r.latitude BETWEEN $1 AND $2
			AND r.longitude BETWEEN $3 AND $4
			AND t.climbed_at >= $5
			AND t.climbed_at <= $6
		GROUP BY r.mp_route_id, r.name, r.rating, r.latitude, r.longitude, r.mp_area_id, a.name
		ORDER BY COUNT(t.id) DESC
		LIMIT $7
	`

	// queryRouteTicksInDateRange retrieves all ticks for a specific route.
	queryRouteTicksInDateRange = `
		SELECT
			t.mp_route_id,
			r.name as route_name,
			COALESCE(r.rating, '') as rating,
			COALESCE(t.user_name, '') as user_name,
			t.climbed_at,
			COALESCE(t.style, '') as style,
			COALESCE(t.comment, '') as comment
		FROM woulder.mp_ticks t
		JOIN woulder.mp_routes r ON t.mp_route_id = r.mp_route_id
		WHERE t.mp_route_id = $1
			AND t.climbed_at >= $2
			AND t.climbed_at <= $3
		ORDER BY t.climbed_at DESC
		LIMIT $4
	`

	// querySearchRoutesInAreas searches for routes within specified areas by name.
	// Case-insensitive partial match using LIKE with LOWER().
	// Only returns routes with activity (HAVING COUNT(t.id) > 0).
	querySearchRoutesInAreas = `
		SELECT
			r.mp_route_id,
			r.name,
			r.rating,
			r.latitude,
			r.longitude,
			COUNT(t.id) as tick_count,
			MAX(t.climbed_at) as last_activity,
			r.mp_area_id,
			a.name as area_name
		FROM woulder.mp_routes r
		LEFT JOIN woulder.mp_ticks t ON r.mp_route_id = t.mp_route_id
			AND t.climbed_at >= $2
			AND t.climbed_at <= $3
		JOIN woulder.mp_areas a ON r.mp_area_id = a.mp_area_id
		WHERE r.mp_area_id = ANY($1)
			AND LOWER(r.name) LIKE LOWER($4)
		GROUP BY r.mp_route_id, r.name, r.rating, r.latitude, r.longitude, r.mp_area_id, a.name
		HAVING COUNT(t.id) > 0
		ORDER BY COUNT(t.id) DESC, r.name ASC
		LIMIT $5
	`
)
