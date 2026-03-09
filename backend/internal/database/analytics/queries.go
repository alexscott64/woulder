package analytics

// SQL queries for analytics data operations.

const (
	// --- Session queries ---

	queryCreateSession = `
		INSERT INTO woulder.analytics_sessions (
			session_id, visitor_id, ip_address, user_agent, referrer,
			device_type, browser, os, screen_width, screen_height,
			started_at, last_active_at, page_count, duration_seconds, is_bounce
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW(), 0, 0, TRUE)
		ON CONFLICT (session_id) DO NOTHING
	`

	queryUpdateSessionActivity = `
		UPDATE woulder.analytics_sessions
		SET last_active_at = NOW(),
			page_count = (SELECT COUNT(*) FROM woulder.analytics_events WHERE session_id = $1 AND event_type = 'page_view'),
			duration_seconds = EXTRACT(EPOCH FROM (NOW() - started_at))::INTEGER,
			is_bounce = (SELECT COUNT(*) FROM woulder.analytics_events WHERE session_id = $1 AND event_type = 'page_view') <= 1
		WHERE session_id = $1
	`

	queryGetSessionByID = `
		SELECT id, session_id, visitor_id, ip_address, user_agent, referrer,
			country, region, city, device_type, browser, os,
			screen_width, screen_height, started_at, last_active_at,
			page_count, duration_seconds, is_bounce, created_at
		FROM woulder.analytics_sessions
		WHERE session_id = $1
	`

	// --- Event queries ---

	queryInsertEvent = `
		INSERT INTO woulder.analytics_events (session_id, event_type, event_name, page_path, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`

	// --- Admin queries ---

	queryGetAdminByUsername = `
		SELECT id, username, password_hash, created_at, last_login_at
		FROM woulder.analytics_admin_users
		WHERE username = $1
	`

	queryUpsertAdmin = `
		INSERT INTO woulder.analytics_admin_users (username, password_hash, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (username) DO UPDATE SET password_hash = $2
	`

	queryUpdateLastLogin = `
		UPDATE woulder.analytics_admin_users
		SET last_login_at = NOW()
		WHERE username = $1
	`

	// --- Metrics queries ---

	queryOverviewMetrics = `
		SELECT
			(SELECT COUNT(DISTINCT visitor_id) FROM woulder.analytics_sessions WHERE started_at >= $1) AS unique_visitors,
			(SELECT COUNT(*) FROM woulder.analytics_sessions WHERE started_at >= $1) AS total_sessions,
			(SELECT COUNT(*) FROM woulder.analytics_events WHERE event_type = 'page_view' AND created_at >= $1) AS total_page_views,
			COALESCE((SELECT AVG(duration_seconds)::FLOAT FROM woulder.analytics_sessions WHERE started_at >= $1 AND duration_seconds > 0), 0) AS avg_session_duration,
			COALESCE(
				(SELECT COUNT(*)::FLOAT / NULLIF(COUNT(*), 0) FROM woulder.analytics_sessions WHERE started_at >= $1 AND is_bounce = TRUE)
				/ NULLIF((SELECT COUNT(*) FROM woulder.analytics_sessions WHERE started_at >= $1), 0),
				0
			) AS bounce_rate,
			(SELECT COUNT(*) FROM woulder.analytics_events WHERE created_at >= $1) AS total_events
	`

	queryVisitorsOverTime = `
		SELECT
			TO_CHAR(started_at AT TIME ZONE 'America/Los_Angeles', 'YYYY-MM-DD') AS date,
			COUNT(DISTINCT visitor_id) AS unique_visitors,
			COUNT(*) AS sessions,
			COALESCE((
				SELECT COUNT(*) FROM woulder.analytics_events e
				WHERE e.event_type = 'page_view'
				AND e.session_id IN (
					SELECT s2.session_id FROM woulder.analytics_sessions s2
					WHERE TO_CHAR(s2.started_at AT TIME ZONE 'America/Los_Angeles', 'YYYY-MM-DD') = TO_CHAR(s.started_at AT TIME ZONE 'America/Los_Angeles', 'YYYY-MM-DD')
				)
			), 0) AS page_views
		FROM woulder.analytics_sessions s
		WHERE started_at >= $1
		GROUP BY TO_CHAR(started_at AT TIME ZONE 'America/Los_Angeles', 'YYYY-MM-DD')
		ORDER BY date ASC
	`

	queryTopPages = `
		SELECT
			COALESCE(page_path, '/') AS page_path,
			COUNT(*) AS view_count,
			COUNT(DISTINCT e.session_id) AS unique_visitors
		FROM woulder.analytics_events e
		WHERE e.event_type = 'page_view' AND e.created_at >= $1
		GROUP BY page_path
		ORDER BY view_count DESC
		LIMIT $2
	`

	queryTopLocations = `
		SELECT
			COALESCE((metadata->>'location_id')::TEXT, 'unknown') AS location_id,
			COALESCE(metadata->>'location_name', 'Unknown') AS location_name,
			COUNT(*) AS view_count,
			COUNT(DISTINCT e.session_id) AS unique_visitors
		FROM woulder.analytics_events e
		WHERE e.event_type = 'location_view' AND e.created_at >= $1
		GROUP BY metadata->>'location_id', metadata->>'location_name'
		ORDER BY view_count DESC
		LIMIT $2
	`

	queryTopAreas = `
		SELECT
			COALESCE(metadata->>'area_id', 'unknown') AS area_id,
			COALESCE(metadata->>'area_name', 'Unknown') AS area_name,
			COALESCE((metadata->>'location_id')::INT, 0) AS location_id,
			COUNT(*) AS view_count,
			COUNT(DISTINCT e.session_id) AS unique_visitors
		FROM woulder.analytics_events e
		WHERE e.event_type = 'area_view' AND e.created_at >= $1
		GROUP BY metadata->>'area_id', metadata->>'area_name', metadata->>'location_id'
		ORDER BY view_count DESC
		LIMIT $2
	`

	queryTopRoutes = `
		SELECT
			COALESCE(metadata->>'route_id', 'unknown') AS route_id,
			COALESCE(metadata->>'route_name', 'Unknown') AS route_name,
			COALESCE(metadata->>'route_type', '') AS route_type,
			COUNT(*) AS view_count,
			COUNT(DISTINCT e.session_id) AS unique_visitors
		FROM woulder.analytics_events e
		WHERE e.event_type = 'route_view' AND e.created_at >= $1
		GROUP BY metadata->>'route_id', metadata->>'route_name', metadata->>'route_type'
		ORDER BY view_count DESC
		LIMIT $2
	`

	queryFeatureUsage = `
		SELECT
			event_name AS feature_name,
			COUNT(*) AS usage_count,
			COUNT(DISTINCT e.session_id) AS unique_users
		FROM woulder.analytics_events e
		WHERE e.event_type IN ('page_view', 'modal_open', 'heatmap', 'search', 'settings')
		AND e.created_at >= $1
		GROUP BY event_name
		ORDER BY usage_count DESC
	`

	queryGeography = `
		SELECT
			COALESCE(country, 'Unknown') AS country,
			COALESCE(region, '') AS region,
			COALESCE(city, '') AS city,
			COUNT(*) AS visit_count,
			COUNT(DISTINCT visitor_id) AS unique_visitors
		FROM woulder.analytics_sessions
		WHERE started_at >= $1
		GROUP BY country, region, city
		ORDER BY visit_count DESC
		LIMIT $2
	`

	queryDeviceBreakdown = `
		SELECT
			COALESCE(device_type, 'unknown') AS name,
			COUNT(*) AS count
		FROM woulder.analytics_sessions
		WHERE started_at >= $1
		GROUP BY device_type
		ORDER BY count DESC
	`

	queryBrowserBreakdown = `
		SELECT
			COALESCE(browser, 'Unknown') AS name,
			COUNT(*) AS count
		FROM woulder.analytics_sessions
		WHERE started_at >= $1
		GROUP BY browser
		ORDER BY count DESC
	`

	queryOSBreakdown = `
		SELECT
			COALESCE(os, 'Unknown') AS name,
			COUNT(*) AS count
		FROM woulder.analytics_sessions
		WHERE started_at >= $1
		GROUP BY os
		ORDER BY count DESC
	`

	queryReferrers = `
		SELECT
			COALESCE(referrer, 'Direct') AS referrer,
			COUNT(*) AS visit_count,
			COUNT(DISTINCT visitor_id) AS unique_visitors
		FROM woulder.analytics_sessions
		WHERE started_at >= $1
		GROUP BY referrer
		ORDER BY visit_count DESC
		LIMIT $2
	`

	queryRecentSessions = `
		SELECT id, session_id, visitor_id, ip_address, user_agent, referrer,
			country, region, city, device_type, browser, os,
			screen_width, screen_height, started_at, last_active_at,
			page_count, duration_seconds, is_bounce, created_at
		FROM woulder.analytics_sessions
		ORDER BY last_active_at DESC
		LIMIT $1
	`

	querySessionEvents = `
		SELECT id, session_id, event_type, event_name, page_path, metadata, created_at
		FROM woulder.analytics_events
		WHERE session_id = $1
		ORDER BY created_at ASC
	`

	queryCleanOldSessions = `
		DELETE FROM woulder.analytics_sessions
		WHERE started_at < $1
	`

	queryCleanOldEvents = `
		DELETE FROM woulder.analytics_events
		WHERE created_at < $1
	`
)
