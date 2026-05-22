package mountainproject

// AreasRepository queries

// querySaveArea inserts or updates a Mountain Project area.
// Uses ON CONFLICT to handle upserts efficiently.
// Indexes: mp_area_id UNIQUE
//
// PERFORMANCE: The DO UPDATE branch is guarded by IS DISTINCT FROM so that
// re-scraping an area whose content has not changed produces zero heap
// writes, zero WAL records, and zero index churn. The catalog is largely
// static (areas only change when MP editors edit them), so the
// overwhelming majority of upserts during background sync are no-ops; this
// guard collapses ~millions of "rewrite an identical row" UPDATEs that
// would otherwise dirty heap pages and inflate RDS checkpoint write/sync
// time. last_synced_at is intentionally excluded from the comparison
// because we want it to update only when the row actually changes; if
// you need a "last seen" timestamp regardless, query the sync job log.
const querySaveArea = `
	INSERT INTO woulder.mp_areas (
		mp_area_id, name, parent_mp_area_id, area_type, location_id,
		latitude, longitude, last_synced_at
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	ON CONFLICT (mp_area_id) DO UPDATE SET
		name = EXCLUDED.name,
		parent_mp_area_id = EXCLUDED.parent_mp_area_id,
		area_type = EXCLUDED.area_type,
		location_id = EXCLUDED.location_id,
		latitude = EXCLUDED.latitude,
		longitude = EXCLUDED.longitude,
		last_synced_at = EXCLUDED.last_synced_at
	WHERE mp_areas.name              IS DISTINCT FROM EXCLUDED.name
	   OR mp_areas.parent_mp_area_id IS DISTINCT FROM EXCLUDED.parent_mp_area_id
	   OR mp_areas.area_type         IS DISTINCT FROM EXCLUDED.area_type
	   OR mp_areas.location_id       IS DISTINCT FROM EXCLUDED.location_id
	   OR mp_areas.latitude          IS DISTINCT FROM EXCLUDED.latitude
	   OR mp_areas.longitude         IS DISTINCT FROM EXCLUDED.longitude
`

// queryGetAreaByID retrieves a Mountain Project area by its MP area ID.
const queryGetAreaByID = `
	SELECT id, mp_area_id, name, parent_mp_area_id, area_type,
	       location_id, latitude, longitude, last_synced_at, created_at, updated_at
	FROM woulder.mp_areas
	WHERE mp_area_id = $1
`

// queryUpdateAreaRouteCount updates the cached route count for an area.
const queryUpdateAreaRouteCount = `
	UPDATE woulder.mp_areas
	SET route_count_total = $1,
	    route_count_last_checked = NOW(),
	    updated_at = NOW()
	WHERE mp_area_id = $2
`

// queryGetAreaRouteCount retrieves the cached route count for an area.
const queryGetAreaRouteCount = `
	SELECT route_count_total
	FROM woulder.mp_areas
	WHERE mp_area_id = $1
`

// queryGetChildAreas retrieves all direct children of a parent area.
const queryGetChildAreas = `
	SELECT mp_area_id, name
	FROM woulder.mp_areas
	WHERE parent_mp_area_id = $1
	ORDER BY name
`

// queryGetAllStateConfigs retrieves all state configurations for syncing.
const queryGetAllStateConfigs = `
	SELECT state_name, mp_area_id, is_active
	FROM woulder.mp_state_configs
	ORDER BY display_order, state_name
`

// RoutesRepository queries

// querySaveRoute inserts or updates a Mountain Project route.
// Uses ON CONFLICT to handle upserts efficiently.
// Indexes: mp_route_id UNIQUE
//
// PERFORMANCE: As with querySaveArea, the DO UPDATE branch is guarded so
// that re-scraping an unchanged route produces zero writes. mp_routes was
// previously the single largest source of UPDATEs in the database: 287k
// live rows accumulated 84M+ UPDATEs (~293 updates per row), and only
// ~25% of those were HOT, meaning ~63M index entries were being rewritten
// for no functional benefit. The IS DISTINCT FROM guard short-circuits
// the no-op case at the Postgres level (no tuple version, no WAL, no
// index update).
const querySaveRoute = `
	INSERT INTO woulder.mp_routes (
		mp_route_id, mp_area_id, name, route_type, rating, location_id,
		latitude, longitude, aspect
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	ON CONFLICT (mp_route_id) DO UPDATE SET
		mp_area_id = EXCLUDED.mp_area_id,
		name = EXCLUDED.name,
		route_type = EXCLUDED.route_type,
		rating = EXCLUDED.rating,
		location_id = EXCLUDED.location_id,
		latitude = EXCLUDED.latitude,
		longitude = EXCLUDED.longitude,
		aspect = EXCLUDED.aspect
	WHERE mp_routes.mp_area_id  IS DISTINCT FROM EXCLUDED.mp_area_id
	   OR mp_routes.name        IS DISTINCT FROM EXCLUDED.name
	   OR mp_routes.route_type  IS DISTINCT FROM EXCLUDED.route_type
	   OR mp_routes.rating      IS DISTINCT FROM EXCLUDED.rating
	   OR mp_routes.location_id IS DISTINCT FROM EXCLUDED.location_id
	   OR mp_routes.latitude    IS DISTINCT FROM EXCLUDED.latitude
	   OR mp_routes.longitude   IS DISTINCT FROM EXCLUDED.longitude
	   OR mp_routes.aspect      IS DISTINCT FROM EXCLUDED.aspect
`

// queryGetRouteByID retrieves a Mountain Project route by its MP route ID.
const queryGetRouteByID = `
	SELECT id, mp_route_id, mp_area_id, name, route_type, rating,
		   location_id, latitude, longitude, aspect,
		   difficulty, pitches, height_feet, mp_rating, popularity,
		   description_text, location_text, protection_text, safety_text,
		   created_at, updated_at
	FROM woulder.mp_routes
	WHERE mp_route_id = $1
`

// queryGetRoutesByIDs retrieves multiple Mountain Project routes by IDs.
const queryGetRoutesByIDs = `
	SELECT id, mp_route_id, mp_area_id, name, route_type, rating,
		   location_id, latitude, longitude, aspect,
		   difficulty, pitches, height_feet, mp_rating, popularity,
		   description_text, location_text, protection_text, safety_text,
		   created_at, updated_at
	FROM woulder.mp_routes
	WHERE mp_route_id = ANY($1)
`

// queryGetRoutesWithGPSByArea retrieves all routes with GPS in an area and subareas.
const queryGetRoutesWithGPSByArea = `
	WITH RECURSIVE area_tree AS (
		-- Start with the given area
		SELECT mp_area_id
		FROM woulder.mp_areas
		WHERE mp_area_id = $1

		UNION ALL

		-- Recursively include all descendant areas
		SELECT a.mp_area_id
		FROM woulder.mp_areas a
		INNER JOIN area_tree at ON a.parent_mp_area_id = at.mp_area_id
	)
	SELECT id, mp_route_id, mp_area_id, name, route_type, rating,
	       location_id, latitude, longitude, aspect,
	       difficulty, pitches, height_feet, mp_rating, popularity,
	       description_text, location_text, protection_text, safety_text,
	       created_at, updated_at
	FROM woulder.mp_routes r
	WHERE r.mp_area_id IN (SELECT mp_area_id FROM area_tree)
	  AND r.latitude IS NOT NULL
	  AND r.longitude IS NOT NULL
	ORDER BY r.name
`

// queryGetAllRouteIDsForLocation retrieves all route IDs for a location.
const queryGetAllRouteIDsForLocation = `
	SELECT mp_route_id
	FROM woulder.mp_routes
	WHERE location_id = $1
`

// queryUpdateRouteGPS updates GPS coordinates and aspect for a route.
const queryUpdateRouteGPS = `
	UPDATE woulder.mp_routes
	SET latitude = $1, longitude = $2, aspect = $3, updated_at = NOW()
	WHERE mp_route_id = $4
`

// queryGetRouteIDsForArea retrieves all route IDs in an area.
const queryGetRouteIDsForArea = `
	SELECT mp_route_id::text
	FROM woulder.mp_routes
	WHERE mp_area_id = $1
	ORDER BY mp_route_id
`

// queryUpsertRoute inserts or updates a route with full update semantics.
// Used by mountainprojectsync for compatibility.
//
// PERFORMANCE: See querySaveRoute. Same IS DISTINCT FROM guard, with the
// added wrinkle that updated_at = NOW() in the SET list is fine because
// the WHERE clause skips the entire UPDATE branch when nothing changed,
// so updated_at only advances when there is a real content change.
const queryUpsertRoute = `
	INSERT INTO woulder.mp_routes (
		mp_route_id, mp_area_id, location_id, name, route_type, rating,
		latitude, longitude, aspect
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	ON CONFLICT (mp_route_id) DO UPDATE SET
		mp_area_id = EXCLUDED.mp_area_id,
		name = EXCLUDED.name,
		route_type = EXCLUDED.route_type,
		rating = EXCLUDED.rating,
		latitude = EXCLUDED.latitude,
		longitude = EXCLUDED.longitude,
		aspect = EXCLUDED.aspect,
		updated_at = NOW()
	WHERE mp_routes.mp_area_id IS DISTINCT FROM EXCLUDED.mp_area_id
	   OR mp_routes.name       IS DISTINCT FROM EXCLUDED.name
	   OR mp_routes.route_type IS DISTINCT FROM EXCLUDED.route_type
	   OR mp_routes.rating     IS DISTINCT FROM EXCLUDED.rating
	   OR mp_routes.latitude   IS DISTINCT FROM EXCLUDED.latitude
	   OR mp_routes.longitude  IS DISTINCT FROM EXCLUDED.longitude
	   OR mp_routes.aspect     IS DISTINCT FROM EXCLUDED.aspect
`

// queryUpdateRouteDetails updates detailed route information fields.
const queryUpdateRouteDetails = `
	UPDATE woulder.mp_routes
	SET difficulty = $2,
		pitches = $3,
		height_feet = $4,
		mp_rating = $5,
		popularity = $6,
		description_text = $7,
		location_text = $8,
		protection_text = $9,
		safety_text = $10,
		updated_at = NOW()
	WHERE mp_route_id = $1
`

// TicksRepository queries

// querySaveTick inserts a Mountain Project tick.
// Uses ON CONFLICT DO NOTHING to handle duplicates.
// Indexes: (mp_route_id, user_name, climbed_at) UNIQUE
const querySaveTick = `
	INSERT INTO woulder.mp_ticks (
		mp_route_id, user_name, climbed_at, style, comment
	)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (mp_route_id, user_name, climbed_at) DO NOTHING
`

// queryUpdateRouteTickSyncTimestamp updates the last_tick_sync_at for a route.
const queryUpdateRouteTickSyncTimestamp = `
	UPDATE woulder.mp_routes
	SET last_tick_sync_at = NOW()
	WHERE mp_route_id = $1
`

// queryGetLastTickTimestamp retrieves the most recent tick timestamp for a route.
const queryGetLastTickTimestamp = `
	SELECT MAX(climbed_at) AS last_tick
	FROM woulder.mp_ticks
	WHERE mp_route_id = $1
`

// CommentsRepository queries

// querySaveAreaComment inserts or updates an area comment.
// Uses ON CONFLICT to handle upserts efficiently.
// Indexes: mp_comment_id UNIQUE
//
// PERFORMANCE: IS DISTINCT FROM guard, same rationale as querySaveRoute.
// Comments are immutable from MP's perspective in practice, so almost
// every upsert during a re-scrape is a no-op.
const querySaveAreaComment = `
	INSERT INTO woulder.mp_comments (
		mp_comment_id, comment_type, mp_area_id, mp_route_id,
		user_name, user_id, comment_text, commented_at
	)
	VALUES ($1, 'area', $2, NULL, $3, NULL, $4, $5)
	ON CONFLICT (mp_comment_id)
	DO UPDATE SET
		user_name = EXCLUDED.user_name,
		comment_text = EXCLUDED.comment_text,
		commented_at = EXCLUDED.commented_at,
		updated_at = CURRENT_TIMESTAMP
	WHERE mp_comments.user_name    IS DISTINCT FROM EXCLUDED.user_name
	   OR mp_comments.comment_text IS DISTINCT FROM EXCLUDED.comment_text
	   OR mp_comments.commented_at IS DISTINCT FROM EXCLUDED.commented_at
`

// querySaveRouteComment inserts or updates a route comment.
// Uses ON CONFLICT to handle upserts efficiently.
// Indexes: mp_comment_id UNIQUE
//
// PERFORMANCE: See querySaveAreaComment.
const querySaveRouteComment = `
	INSERT INTO woulder.mp_comments (
		mp_comment_id, comment_type, mp_area_id, mp_route_id,
		user_name, user_id, comment_text, commented_at
	)
	VALUES ($1, 'route', NULL, $2, $3, NULL, $4, $5)
	ON CONFLICT (mp_comment_id)
	DO UPDATE SET
		user_name = EXCLUDED.user_name,
		comment_text = EXCLUDED.comment_text,
		commented_at = EXCLUDED.commented_at,
		updated_at = CURRENT_TIMESTAMP
	WHERE mp_comments.user_name    IS DISTINCT FROM EXCLUDED.user_name
	   OR mp_comments.comment_text IS DISTINCT FROM EXCLUDED.comment_text
	   OR mp_comments.commented_at IS DISTINCT FROM EXCLUDED.commented_at
`

// queryUpdateRouteCommentSyncTimestamp updates last_comment_sync_at for a route.
const queryUpdateRouteCommentSyncTimestamp = `
	UPDATE woulder.mp_routes
	SET last_comment_sync_at = NOW()
	WHERE mp_route_id = $1
`

// queryUpsertAreaComment inserts or updates an area comment with user_id support.
// Used by mountainprojectsync for compatibility.
//
// PERFORMANCE: See querySaveAreaComment. user_id is included in the
// comparison because it's part of the payload here (unlike the
// querySave* variants which always pass NULL for user_id).
const queryUpsertAreaComment = `
	INSERT INTO woulder.mp_comments (
		mp_comment_id, comment_type, mp_area_id, mp_route_id,
		user_name, user_id, comment_text, commented_at
	)
	VALUES ($1, 'area', $2, NULL, $3, $4, $5, $6)
	ON CONFLICT (mp_comment_id) DO UPDATE SET
		user_name = EXCLUDED.user_name,
		user_id = EXCLUDED.user_id,
		comment_text = EXCLUDED.comment_text,
		commented_at = EXCLUDED.commented_at,
		updated_at = NOW()
	WHERE mp_comments.user_name    IS DISTINCT FROM EXCLUDED.user_name
	   OR mp_comments.user_id      IS DISTINCT FROM EXCLUDED.user_id
	   OR mp_comments.comment_text IS DISTINCT FROM EXCLUDED.comment_text
	   OR mp_comments.commented_at IS DISTINCT FROM EXCLUDED.commented_at
`

// queryUpsertRouteComment inserts or updates a route comment with user_id support.
// Used by mountainprojectsync for compatibility.
//
// PERFORMANCE: See queryUpsertAreaComment.
const queryUpsertRouteComment = `
	INSERT INTO woulder.mp_comments (
		mp_comment_id, comment_type, mp_area_id, mp_route_id,
		user_name, user_id, comment_text, commented_at
	)
	VALUES ($1, 'route', NULL, $2, $3, $4, $5, $6)
	ON CONFLICT (mp_comment_id) DO UPDATE SET
		user_name = EXCLUDED.user_name,
		user_id = EXCLUDED.user_id,
		comment_text = EXCLUDED.comment_text,
		commented_at = EXCLUDED.commented_at,
		updated_at = NOW()
	WHERE mp_comments.user_name    IS DISTINCT FROM EXCLUDED.user_name
	   OR mp_comments.user_id      IS DISTINCT FROM EXCLUDED.user_id
	   OR mp_comments.comment_text IS DISTINCT FROM EXCLUDED.comment_text
	   OR mp_comments.commented_at IS DISTINCT FROM EXCLUDED.commented_at
`

// SyncRepository queries

// queryUpdateRouteSyncPriorities recalculates route priorities using a hybrid multi-signal system
// that adapts to seasonal patterns, activity surges, and per-area population differences.
//
// Priority Signals (evaluated in order):
// 1. Seasonal routes (Ice/Alpine/Snow/Mixed): ALWAYS HIGH (need constant monitoring for conditions)
// 2. Activity surge: HIGH if recent tick (14d) after 90+ day gap (catches season starts)
// 3. Per-area ranking: HIGH if top 15% in area (popular routes in each area)
// 4. Absolute threshold: HIGH if 5+ ticks in 90 days (busy routes)
// 5. MEDIUM if any activity (1-4 ticks in 90d) or above-average for area (60th+ percentile)
// 6. LOW otherwise (no activity and below 60th percentile)
//
// Target distribution: ~50K high (daily sync), ~125K medium (weekly sync), ~100K low (monthly sync)
// This system ensures seasonal routes are always monitored while optimizing API calls for most routes.
const queryUpdateRouteSyncPriorities = `
	WITH route_metrics AS (
		SELECT
			r.mp_route_id,
			r.mp_area_id,
			r.route_type,
			COUNT(CASE WHEN t.climbed_at >= NOW() - INTERVAL '14 days' THEN 1 END) AS tick_count_14d,
			COUNT(CASE WHEN t.climbed_at >= NOW() - INTERVAL '90 days' THEN 1 END) AS tick_count_90d,
			COUNT(t.climbed_at) AS total_tick_count,
			(NOW()::date - MAX(t.climbed_at)::date) AS days_since_last_tick,
			(NOW()::date - r.created_at::date) AS route_age_days
		FROM woulder.mp_routes r
		LEFT JOIN woulder.mp_ticks t ON r.mp_route_id = t.mp_route_id
		WHERE r.location_id IS NULL  -- Only non-location routes
		GROUP BY r.mp_route_id, r.mp_area_id, r.route_type, r.created_at
	),
	area_percentiles AS (
		SELECT
			mp_route_id,
			PERCENT_RANK() OVER (
				PARTITION BY mp_area_id
				ORDER BY tick_count_90d
			) as area_percentile
		FROM route_metrics
	)
	UPDATE woulder.mp_routes r
	SET
		tick_count_14d = m.tick_count_14d,
		tick_count_90d = m.tick_count_90d,
		total_tick_count = m.total_tick_count,
		days_since_last_tick = m.days_since_last_tick,
		area_percentile = p.area_percentile,
		sync_priority = CASE
			-- Seasonal routes: ALWAYS HIGH (ice climbing, alpine, etc need constant monitoring)
			WHEN m.route_type IN ('Ice', 'Alpine', 'Snow', 'Mixed') THEN 'high'

			-- Activity surge: Recent ticks after long gap = HIGH (catches season starts)
			WHEN m.tick_count_14d >= 1 AND m.days_since_last_tick > 90 THEN 'high'

			-- Per-area top performers: Top 15% = HIGH (popular routes in each area)
			WHEN p.area_percentile >= 0.85 THEN 'high'

			-- Absolute threshold: Busy routes = HIGH (5+ ticks in 90 days)
			WHEN m.tick_count_90d >= 5 THEN 'high'

			-- Medium: Any activity (1-4 ticks) OR above-average for area (60th+ percentile)
			WHEN m.tick_count_90d >= 1 OR p.area_percentile >= 0.60 THEN 'medium'

			-- Low: Everything else (no activity and below 60th percentile)
			ELSE 'low'
		END,
		updated_at = NOW()
	FROM route_metrics m
	JOIN area_percentiles p ON m.mp_route_id = p.mp_route_id
	WHERE r.mp_route_id = m.mp_route_id
`

// queryGetLocationRoutesDueForTickSync retrieves location routes needing tick sync.
// Location routes always sync daily regardless of activity.
const queryGetLocationRoutesDueForTickSync = `
	SELECT mp_route_id
	FROM woulder.mp_routes
	WHERE
		location_id IS NOT NULL
		AND (
			last_tick_sync_at IS NULL
			OR last_tick_sync_at < NOW() - INTERVAL '24 hours'
		)
	ORDER BY last_tick_sync_at ASC NULLS FIRST
`

// queryGetLocationRoutesDueForCommentSync retrieves location routes needing comment sync.
// Location routes always sync daily regardless of activity.
const queryGetLocationRoutesDueForCommentSync = `
	SELECT mp_route_id
	FROM woulder.mp_routes
	WHERE
		location_id IS NOT NULL
		AND (
			last_comment_sync_at IS NULL
			OR last_comment_sync_at < NOW() - INTERVAL '24 hours'
		)
	ORDER BY last_comment_sync_at ASC NULLS FIRST
`

// queryGetRoutesDueForTickSyncTemplate is a template for priority-based tick sync queries.
// The %s placeholder is replaced with the appropriate interval (24 hours, 7 days, 30 days).
const queryGetRoutesDueForTickSyncTemplate = `
	SELECT mp_route_id
	FROM woulder.mp_routes
	WHERE
		location_id IS NULL
		AND sync_priority = $1
		AND (
			last_tick_sync_at IS NULL
			OR last_tick_sync_at < NOW() - INTERVAL '%s'
		)
	ORDER BY last_tick_sync_at ASC NULLS FIRST
`

// queryGetRoutesDueForCommentSyncTemplate is a template for priority-based comment sync queries.
// The %s placeholder is replaced with the appropriate interval (24 hours, 7 days, 30 days).
const queryGetRoutesDueForCommentSyncTemplate = `
	SELECT mp_route_id
	FROM woulder.mp_routes
	WHERE
		location_id IS NULL
		AND sync_priority = $1
		AND (
			last_comment_sync_at IS NULL
			OR last_comment_sync_at < NOW() - INTERVAL '%s'
		)
	ORDER BY last_comment_sync_at ASC NULLS FIRST
`

// queryGetPriorityDistribution retrieves count of routes in each priority tier.
const queryGetPriorityDistribution = `
	SELECT sync_priority, COUNT(*) as count
	FROM woulder.mp_routes
	WHERE location_id IS NULL
	GROUP BY sync_priority
`
