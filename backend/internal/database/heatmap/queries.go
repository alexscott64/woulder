package heatmap

// SQL queries for heat map and activity data operations.
// Heat map queries are performance-critical for clustering and must be optimized.

const (
	// queryHeatMapDataLightweight retrieves minimal data for clustering performance.
	// Active routes and has_subareas are not calculated to reduce query complexity.
	// Used for initial map loads and low zoom levels where speed is critical.
	// UPDATED: Now includes both MP ticks and Kaya ascents for comprehensive activity data.
	queryHeatMapDataLightweight = `
		WITH combined_activity AS (
			-- MP ticks
			SELECT
				a.mp_area_id,
				t.id::text as activity_id,
				t.climbed_at,
				t.user_name,
				t.mp_route_id
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
			
			UNION ALL
			
			-- Kaya ascents matched to MP routes
			SELECT
				a.mp_area_id,
				ka.kaya_ascent_id as activity_id,
				ka.date as climbed_at,
				ku.username as user_name,
				mr.mp_route_id
			FROM woulder.mp_areas a
			JOIN woulder.mp_routes r ON r.mp_area_id = a.mp_area_id
			JOIN woulder.kaya_mp_route_matches mr ON mr.mp_route_id = r.mp_route_id
			JOIN woulder.kaya_climbs kc ON kc.slug = mr.kaya_climb_id
			JOIN woulder.kaya_ascents ka ON ka.kaya_climb_slug = kc.slug
			JOIN woulder.kaya_users ku ON ku.kaya_user_id = ka.kaya_user_id
			WHERE ka.date >= $1
				AND ka.date <= $2
				AND a.latitude IS NOT NULL
				AND a.longitude IS NOT NULL
				AND ($3::float IS NULL OR (
					a.latitude BETWEEN $3 AND $4
					AND a.longitude BETWEEN $5 AND $6
				))
				AND mr.match_confidence >= 0.60
				AND ($7::text[] IS NULL OR r.route_type = ANY($7))
		)
		SELECT
			a.mp_area_id,
			a.name,
			a.latitude,
			a.longitude,
			0 as active_routes,
			COUNT(ca.activity_id) as total_ticks,
			MAX(ca.climbed_at) as last_activity,
			COUNT(DISTINCT CASE WHEN ca.user_name IS NOT NULL THEN ca.user_name END) as unique_climbers,
			false as has_subareas
		FROM woulder.mp_areas a
		JOIN combined_activity ca ON ca.mp_area_id = a.mp_area_id
		GROUP BY a.mp_area_id, a.name, a.latitude, a.longitude
		HAVING COUNT(ca.activity_id) >= $8
		ORDER BY COUNT(ca.activity_id) DESC
		LIMIT $9;
	`

	// queryHeatMapDataFull retrieves complete data with all aggregations.
	// Includes active route counts and subarea detection for detailed clustering.
	// Used for high zoom levels and detailed area views.
	// UPDATED: Now includes both MP ticks and Kaya ascents for comprehensive activity data.
	queryHeatMapDataFull = `
		WITH combined_activity AS (
			-- MP ticks
			SELECT
				a.mp_area_id,
				t.id::text as activity_id,
				t.climbed_at,
				t.user_name,
				t.mp_route_id
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
			
			UNION ALL
			
			-- Kaya ascents matched to MP routes
			SELECT
				a.mp_area_id,
				ka.kaya_ascent_id as activity_id,
				ka.date as climbed_at,
				ku.username as user_name,
				mr.mp_route_id
			FROM woulder.mp_areas a
			JOIN woulder.mp_routes r ON r.mp_area_id = a.mp_area_id
			JOIN woulder.kaya_mp_route_matches mr ON mr.mp_route_id = r.mp_route_id
			JOIN woulder.kaya_climbs kc ON kc.slug = mr.kaya_climb_id
			JOIN woulder.kaya_ascents ka ON ka.kaya_climb_slug = kc.slug
			JOIN woulder.kaya_users ku ON ku.kaya_user_id = ka.kaya_user_id
			WHERE ka.date >= $1
				AND ka.date <= $2
				AND a.latitude IS NOT NULL
				AND a.longitude IS NOT NULL
				AND ($3::float IS NULL OR (
					a.latitude BETWEEN $3 AND $4
					AND a.longitude BETWEEN $5 AND $6
				))
				AND mr.match_confidence >= 0.60
				AND ($7::text[] IS NULL OR r.route_type = ANY($7))
		)
		SELECT
			a.mp_area_id,
			a.name,
			a.latitude,
			a.longitude,
			COUNT(DISTINCT ca.mp_route_id) as active_routes,
			COUNT(ca.activity_id) as total_ticks,
			MAX(ca.climbed_at) as last_activity,
			COUNT(DISTINCT ca.user_name) as unique_climbers,
			EXISTS(
				SELECT 1 FROM woulder.mp_areas sub
				WHERE sub.parent_mp_area_id = a.mp_area_id
				LIMIT 1
			) as has_subareas
		FROM woulder.mp_areas a
		JOIN combined_activity ca ON ca.mp_area_id = a.mp_area_id
		GROUP BY a.mp_area_id, a.name, a.latitude, a.longitude
		HAVING COUNT(ca.activity_id) >= $8
		ORDER BY COUNT(ca.activity_id) DESC
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
	// Includes both MP ticks and Kaya ascents.
	queryAreaActivityStats = `
		WITH combined_activity AS (
			-- MP ticks
			SELECT
				t.id::text as activity_id,
				t.climbed_at,
				t.user_name,
				t.mp_route_id
			FROM woulder.mp_ticks t
			JOIN woulder.mp_routes r ON t.mp_route_id = r.mp_route_id
			WHERE r.mp_area_id = $1
				AND t.climbed_at >= $2
				AND t.climbed_at <= $3
			
			UNION ALL
			
			-- Kaya ascents matched to MP routes in this area
			SELECT
				ka.kaya_ascent_id as activity_id,
				ka.date as climbed_at,
				ku.username as user_name,
				mr.mp_route_id
			FROM woulder.kaya_ascents ka
			JOIN woulder.kaya_climbs kc ON ka.kaya_climb_slug = kc.slug
			JOIN woulder.kaya_mp_route_matches mr ON kc.slug = mr.kaya_climb_id
			JOIN woulder.mp_routes r ON mr.mp_route_id = r.mp_route_id
			JOIN woulder.kaya_users ku ON ku.kaya_user_id = ka.kaya_user_id
			WHERE r.mp_area_id = $1
				AND ka.date >= $2
				AND ka.date <= $3
				AND mr.match_confidence >= 0.60
		)
		SELECT
			COUNT(activity_id) as total_ticks,
			COUNT(DISTINCT mp_route_id) as active_routes,
			COUNT(DISTINCT CASE WHEN user_name IS NOT NULL THEN user_name END) as unique_climbers,
			MAX(climbed_at) as last_activity
		FROM combined_activity
	`

	// queryRecentTicks retrieves the most recent ticks for an area.
	// Includes both MP ticks and Kaya ascents.
	queryRecentTicks = `
		WITH combined_ticks AS (
			-- MP ticks
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
			
			UNION ALL
			
			-- Kaya ascents
			SELECT
				mr.mp_route_id,
				kc.name as route_name,
				COALESCE(kc.grade_name, '') as rating,
				ku.username as user_name,
				ka.date as climbed_at,
				'' as style,
				COALESCE(ka.comment, '') as comment
			FROM woulder.kaya_ascents ka
			JOIN woulder.kaya_climbs kc ON ka.kaya_climb_slug = kc.slug
			JOIN woulder.kaya_mp_route_matches mr ON kc.slug = mr.kaya_climb_id
			JOIN woulder.mp_routes r ON mr.mp_route_id = r.mp_route_id
			JOIN woulder.kaya_users ku ON ku.kaya_user_id = ka.kaya_user_id
			WHERE r.mp_area_id = $1
				AND ka.date >= $2
				AND ka.date <= $3
				AND mr.match_confidence >= 0.60
		)
		SELECT * FROM combined_ticks
		ORDER BY climbed_at DESC
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
	// Includes both MP ticks and Kaya ascents.
	queryActivityTimeline = `
		WITH combined_activity AS (
			-- MP ticks
			SELECT
				DATE(t.climbed_at) as date,
				t.id::text as activity_id,
				t.mp_route_id
			FROM woulder.mp_ticks t
			JOIN woulder.mp_routes r ON t.mp_route_id = r.mp_route_id
			WHERE r.mp_area_id = $1
				AND t.climbed_at >= $2
				AND t.climbed_at <= $3
			
			UNION ALL
			
			-- Kaya ascents
			SELECT
				DATE(ka.date) as date,
				ka.kaya_ascent_id as activity_id,
				mr.mp_route_id
			FROM woulder.kaya_ascents ka
			JOIN woulder.kaya_climbs kc ON ka.kaya_climb_slug = kc.slug
			JOIN woulder.kaya_mp_route_matches mr ON kc.slug = mr.kaya_climb_id
			JOIN woulder.mp_routes r ON mr.mp_route_id = r.mp_route_id
			WHERE r.mp_area_id = $1
				AND ka.date >= $2
				AND ka.date <= $3
				AND mr.match_confidence >= 0.60
		)
		SELECT
			date,
			COUNT(activity_id) as tick_count,
			COUNT(DISTINCT mp_route_id) as route_count
		FROM combined_activity
		GROUP BY date
		ORDER BY date ASC
	`

	// queryTopRoutes retrieves the most active routes in an area.
	// Includes both MP ticks and Kaya ascents.
	queryTopRoutes = `
		WITH combined_activity AS (
			-- MP ticks
			SELECT
				r.mp_route_id,
				r.name,
				r.rating,
				t.id::text as activity_id,
				t.climbed_at
			FROM woulder.mp_routes r
			JOIN woulder.mp_ticks t ON r.mp_route_id = t.mp_route_id
			WHERE r.mp_area_id = $1
				AND t.climbed_at >= $2
				AND t.climbed_at <= $3
			
			UNION ALL
			
			-- Kaya ascents
			SELECT
				r.mp_route_id,
				r.name,
				r.rating,
				ka.kaya_ascent_id as activity_id,
				ka.date as climbed_at
			FROM woulder.mp_routes r
			JOIN woulder.kaya_mp_route_matches mr ON mr.mp_route_id = r.mp_route_id
			JOIN woulder.kaya_climbs kc ON kc.slug = mr.kaya_climb_id
			JOIN woulder.kaya_ascents ka ON ka.kaya_climb_slug = kc.slug
			WHERE r.mp_area_id = $1
				AND ka.date >= $2
				AND ka.date <= $3
				AND mr.match_confidence >= 0.60
		)
		SELECT
			mp_route_id,
			name,
			rating,
			COUNT(activity_id) as tick_count,
			MAX(climbed_at) as last_activity
		FROM combined_activity
		GROUP BY mp_route_id, name, rating
		ORDER BY COUNT(activity_id) DESC
		LIMIT 10
	`

	// queryRoutesByBounds retrieves routes within geographic bounds with activity.
	// Used for precision route-level clustering at high zoom levels.
	// Includes both MP ticks and Kaya ascents.
	queryRoutesByBounds = `
		WITH combined_activity AS (
			-- MP ticks
			SELECT
				r.mp_route_id,
				r.name,
				r.rating,
				r.latitude,
				r.longitude,
				r.mp_area_id,
				a.name as area_name,
				t.id::text as activity_id,
				t.climbed_at
			FROM woulder.mp_routes r
			JOIN woulder.mp_ticks t ON r.mp_route_id = t.mp_route_id
			JOIN woulder.mp_areas a ON r.mp_area_id = a.mp_area_id
			WHERE r.latitude IS NOT NULL
				AND r.longitude IS NOT NULL
				AND r.latitude BETWEEN $1 AND $2
				AND r.longitude BETWEEN $3 AND $4
				AND t.climbed_at >= $5
				AND t.climbed_at <= $6
			
			UNION ALL
			
			-- Kaya ascents
			SELECT
				r.mp_route_id,
				r.name,
				r.rating,
				r.latitude,
				r.longitude,
				r.mp_area_id,
				a.name as area_name,
				ka.kaya_ascent_id as activity_id,
				ka.date as climbed_at
			FROM woulder.mp_routes r
			JOIN woulder.kaya_mp_route_matches mr ON mr.mp_route_id = r.mp_route_id
			JOIN woulder.kaya_climbs kc ON kc.slug = mr.kaya_climb_id
			JOIN woulder.kaya_ascents ka ON ka.kaya_climb_slug = kc.slug
			JOIN woulder.mp_areas a ON r.mp_area_id = a.mp_area_id
			WHERE r.latitude IS NOT NULL
				AND r.longitude IS NOT NULL
				AND r.latitude BETWEEN $1 AND $2
				AND r.longitude BETWEEN $3 AND $4
				AND ka.date >= $5
				AND ka.date <= $6
				AND mr.match_confidence >= 0.60
		)
		SELECT
			mp_route_id,
			name,
			rating,
			latitude,
			longitude,
			COUNT(activity_id) as tick_count,
			MAX(climbed_at) as last_activity,
			mp_area_id,
			area_name
		FROM combined_activity
		GROUP BY mp_route_id, name, rating, latitude, longitude, mp_area_id, area_name
		ORDER BY COUNT(activity_id) DESC
		LIMIT $7
	`

	// queryRouteTicksInDateRange retrieves all ticks for a specific route.
	// Includes both MP ticks and Kaya ascents.
	queryRouteTicksInDateRange = `
		WITH combined_ticks AS (
			-- MP ticks
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
			
			UNION ALL
			
			-- Kaya ascents
			SELECT
				mr.mp_route_id,
				kc.name as route_name,
				COALESCE(kc.grade_name, '') as rating,
				ku.username as user_name,
				ka.date as climbed_at,
				'' as style,
				COALESCE(ka.comment, '') as comment
			FROM woulder.kaya_ascents ka
			JOIN woulder.kaya_climbs kc ON ka.kaya_climb_slug = kc.slug
			JOIN woulder.kaya_mp_route_matches mr ON kc.slug = mr.kaya_climb_id
			JOIN woulder.kaya_users ku ON ku.kaya_user_id = ka.kaya_user_id
			WHERE mr.mp_route_id = $1
				AND ka.date >= $2
				AND ka.date <= $3
				AND mr.match_confidence >= 0.60
		)
		SELECT * FROM combined_ticks
		ORDER BY climbed_at DESC
		LIMIT $4
	`

	// querySearchRoutesInAreas searches for routes within specified areas by name.
	// Case-insensitive partial match using LIKE with LOWER().
	// Only returns routes with activity (HAVING COUNT(activity_id) > 0).
	// Includes both MP ticks and Kaya ascents.
	querySearchRoutesInAreas = `
		WITH combined_activity AS (
			-- MP ticks
			SELECT
				r.mp_route_id,
				t.id::text as activity_id,
				t.climbed_at
			FROM woulder.mp_routes r
			LEFT JOIN woulder.mp_ticks t ON r.mp_route_id = t.mp_route_id
				AND t.climbed_at >= $2
				AND t.climbed_at <= $3
			WHERE r.mp_area_id = ANY($1)
				AND t.id IS NOT NULL
			
			UNION ALL
			
			-- Kaya ascents
			SELECT
				r.mp_route_id,
				ka.kaya_ascent_id as activity_id,
				ka.date as climbed_at
			FROM woulder.mp_routes r
			JOIN woulder.kaya_mp_route_matches mr ON mr.mp_route_id = r.mp_route_id
			JOIN woulder.kaya_climbs kc ON kc.slug = mr.kaya_climb_id
			JOIN woulder.kaya_ascents ka ON ka.kaya_climb_slug = kc.slug
			WHERE r.mp_area_id = ANY($1)
				AND ka.date >= $2
				AND ka.date <= $3
				AND mr.match_confidence >= 0.60
		)
		SELECT
			r.mp_route_id,
			r.name,
			r.rating,
			r.latitude,
			r.longitude,
			COUNT(ca.activity_id) as tick_count,
			MAX(ca.climbed_at) as last_activity,
			r.mp_area_id,
			a.name as area_name
		FROM woulder.mp_routes r
		LEFT JOIN combined_activity ca ON r.mp_route_id = ca.mp_route_id
		JOIN woulder.mp_areas a ON r.mp_area_id = a.mp_area_id
		WHERE r.mp_area_id = ANY($1)
			AND LOWER(r.name) LIKE LOWER($4)
		GROUP BY r.mp_route_id, r.name, r.rating, r.latitude, r.longitude, r.mp_area_id, a.name
		HAVING COUNT(ca.activity_id) > 0
		ORDER BY COUNT(ca.activity_id) DESC, r.name ASC
		LIMIT $5
	`
)
