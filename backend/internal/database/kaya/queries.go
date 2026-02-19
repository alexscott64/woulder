package kaya

// SQL queries for Kaya data operations

// Users queries
const (
	querySaveUser = `
		INSERT INTO woulder.kaya_users (
			kaya_user_id, username, fname, lname, photo_url, bio,
			height, ape_index, limit_grade_bouldering_id, limit_grade_bouldering_name,
			limit_grade_routes_id, limit_grade_routes_name, is_private, is_premium
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (kaya_user_id) DO UPDATE SET
			username = EXCLUDED.username,
			fname = EXCLUDED.fname,
			lname = EXCLUDED.lname,
			photo_url = EXCLUDED.photo_url,
			bio = EXCLUDED.bio,
			height = EXCLUDED.height,
			ape_index = EXCLUDED.ape_index,
			limit_grade_bouldering_id = EXCLUDED.limit_grade_bouldering_id,
			limit_grade_bouldering_name = EXCLUDED.limit_grade_bouldering_name,
			limit_grade_routes_id = EXCLUDED.limit_grade_routes_id,
			limit_grade_routes_name = EXCLUDED.limit_grade_routes_name,
			is_private = EXCLUDED.is_private,
			is_premium = EXCLUDED.is_premium,
			updated_at = NOW()
	`

	queryGetUserByID = `
		SELECT id, kaya_user_id, username, fname, lname, photo_url, bio,
			height, ape_index, limit_grade_bouldering_id, limit_grade_bouldering_name,
			limit_grade_routes_id, limit_grade_routes_name, is_private, is_premium,
			created_at, updated_at
		FROM woulder.kaya_users
		WHERE kaya_user_id = $1
	`

	queryGetUserByUsername = `
		SELECT id, kaya_user_id, username, fname, lname, photo_url, bio,
			height, ape_index, limit_grade_bouldering_id, limit_grade_bouldering_name,
			limit_grade_routes_id, limit_grade_routes_name, is_private, is_premium,
			created_at, updated_at
		FROM woulder.kaya_users
		WHERE username = $1
	`
)

// Locations queries
const (
	querySaveLocation = `
		INSERT INTO woulder.kaya_locations (
			kaya_location_id, slug, name, latitude, longitude, photo_url, description,
			location_type_id, location_type_name, parent_location_id, parent_location_slug,
			parent_location_name, climb_count, boulder_count, route_count, ascent_count,
			is_gb_moderated_bouldering, is_gb_moderated_routes, is_access_sensitive,
			is_closed, has_maps_disabled, closed_date, description_bouldering,
			description_routes, description_short_bouldering, description_short_routes,
			access_description_bouldering, access_description_routes,
			access_issues_description_bouldering, access_issues_description_routes,
			climb_type_id, woulder_location_id, last_synced_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30,
			$31, $32, $33
		)
		ON CONFLICT (kaya_location_id) DO UPDATE SET
			slug = EXCLUDED.slug,
			name = EXCLUDED.name,
			latitude = EXCLUDED.latitude,
			longitude = EXCLUDED.longitude,
			photo_url = EXCLUDED.photo_url,
			description = EXCLUDED.description,
			location_type_id = EXCLUDED.location_type_id,
			location_type_name = EXCLUDED.location_type_name,
			parent_location_id = EXCLUDED.parent_location_id,
			parent_location_slug = EXCLUDED.parent_location_slug,
			parent_location_name = EXCLUDED.parent_location_name,
			climb_count = EXCLUDED.climb_count,
			boulder_count = EXCLUDED.boulder_count,
			route_count = EXCLUDED.route_count,
			ascent_count = EXCLUDED.ascent_count,
			is_gb_moderated_bouldering = EXCLUDED.is_gb_moderated_bouldering,
			is_gb_moderated_routes = EXCLUDED.is_gb_moderated_routes,
			is_access_sensitive = EXCLUDED.is_access_sensitive,
			is_closed = EXCLUDED.is_closed,
			has_maps_disabled = EXCLUDED.has_maps_disabled,
			closed_date = EXCLUDED.closed_date,
			description_bouldering = EXCLUDED.description_bouldering,
			description_routes = EXCLUDED.description_routes,
			description_short_bouldering = EXCLUDED.description_short_bouldering,
			description_short_routes = EXCLUDED.description_short_routes,
			access_description_bouldering = EXCLUDED.access_description_bouldering,
			access_description_routes = EXCLUDED.access_description_routes,
			access_issues_description_bouldering = EXCLUDED.access_issues_description_bouldering,
			access_issues_description_routes = EXCLUDED.access_issues_description_routes,
			climb_type_id = EXCLUDED.climb_type_id,
			woulder_location_id = EXCLUDED.woulder_location_id,
			last_synced_at = EXCLUDED.last_synced_at,
			updated_at = NOW()
	`

	queryGetLocationByID = `
		SELECT id, kaya_location_id, slug, name, latitude, longitude, photo_url, description,
			location_type_id, location_type_name, parent_location_id, parent_location_slug,
			parent_location_name, climb_count, boulder_count, route_count, ascent_count,
			is_gb_moderated_bouldering, is_gb_moderated_routes, is_access_sensitive,
			is_closed, has_maps_disabled, closed_date, description_bouldering,
			description_routes, description_short_bouldering, description_short_routes,
			access_description_bouldering, access_description_routes,
			access_issues_description_bouldering, access_issues_description_routes,
			climb_type_id, woulder_location_id, last_synced_at, created_at, updated_at
		FROM woulder.kaya_locations
		WHERE kaya_location_id = $1
	`

	queryGetLocationBySlug = `
		SELECT id, kaya_location_id, slug, name, latitude, longitude, photo_url, description,
			location_type_id, location_type_name, parent_location_id, parent_location_slug,
			parent_location_name, climb_count, boulder_count, route_count, ascent_count,
			is_gb_moderated_bouldering, is_gb_moderated_routes, is_access_sensitive,
			is_closed, has_maps_disabled, closed_date, description_bouldering,
			description_routes, description_short_bouldering, description_short_routes,
			access_description_bouldering, access_description_routes,
			access_issues_description_bouldering, access_issues_description_routes,
			climb_type_id, woulder_location_id, last_synced_at, created_at, updated_at
		FROM woulder.kaya_locations
		WHERE slug = $1
	`

	queryGetSubLocations = `
		SELECT id, kaya_location_id, slug, name, latitude, longitude, photo_url, description,
			location_type_id, location_type_name, parent_location_id, parent_location_slug,
			parent_location_name, climb_count, boulder_count, route_count, ascent_count,
			is_gb_moderated_bouldering, is_gb_moderated_routes, is_access_sensitive,
			is_closed, has_maps_disabled, closed_date, description_bouldering,
			description_routes, description_short_bouldering, description_short_routes,
			access_description_bouldering, access_description_routes,
			access_issues_description_bouldering, access_issues_description_routes,
			climb_type_id, woulder_location_id, last_synced_at, created_at, updated_at
		FROM woulder.kaya_locations
		WHERE parent_location_id = $1
		ORDER BY name
	`

	queryGetAllLocations = `
		SELECT id, kaya_location_id, slug, name, latitude, longitude, photo_url, description,
			location_type_id, location_type_name, parent_location_id, parent_location_slug,
			parent_location_name, climb_count, boulder_count, route_count, ascent_count,
			is_gb_moderated_bouldering, is_gb_moderated_routes, is_access_sensitive,
			is_closed, has_maps_disabled, closed_date, description_bouldering,
			description_routes, description_short_bouldering, description_short_routes,
			access_description_bouldering, access_description_routes,
			access_issues_description_bouldering, access_issues_description_routes,
			climb_type_id, woulder_location_id, last_synced_at, created_at, updated_at
		FROM woulder.kaya_locations
		ORDER BY name
	`
)

// Climbs queries
const (
	querySaveClimb = `
		INSERT INTO woulder.kaya_climbs (
			kaya_climb_id, slug, name, grade_id, grade_name, grade_ordering,
			grade_climb_type_id, climb_type_id, climb_type_name, rating, ascent_count,
			kaya_destination_id, kaya_destination_name, kaya_area_id, kaya_area_name,
			color_name, gym_name, board_name, is_gb_moderated, is_access_sensitive,
			is_closed, is_offensive, woulder_location_id, last_synced_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22, $23, $24
		)
		ON CONFLICT (slug) DO UPDATE SET
			kaya_climb_id = EXCLUDED.kaya_climb_id,
			name = EXCLUDED.name,
			grade_id = EXCLUDED.grade_id,
			grade_name = EXCLUDED.grade_name,
			grade_ordering = EXCLUDED.grade_ordering,
			grade_climb_type_id = EXCLUDED.grade_climb_type_id,
			climb_type_id = EXCLUDED.climb_type_id,
			climb_type_name = EXCLUDED.climb_type_name,
			rating = EXCLUDED.rating,
			ascent_count = EXCLUDED.ascent_count,
			kaya_destination_id = EXCLUDED.kaya_destination_id,
			kaya_destination_name = EXCLUDED.kaya_destination_name,
			kaya_area_id = EXCLUDED.kaya_area_id,
			kaya_area_name = EXCLUDED.kaya_area_name,
			color_name = EXCLUDED.color_name,
			gym_name = EXCLUDED.gym_name,
			board_name = EXCLUDED.board_name,
			is_gb_moderated = EXCLUDED.is_gb_moderated,
			is_access_sensitive = EXCLUDED.is_access_sensitive,
			is_closed = EXCLUDED.is_closed,
			is_offensive = EXCLUDED.is_offensive,
			woulder_location_id = EXCLUDED.woulder_location_id,
			last_synced_at = EXCLUDED.last_synced_at,
			updated_at = NOW()
	`

	queryGetClimbBySlug = `
		SELECT id, kaya_climb_id, slug, name, grade_id, grade_name, grade_ordering,
			grade_climb_type_id, climb_type_id, climb_type_name, rating, ascent_count,
			kaya_destination_id, kaya_destination_name, kaya_area_id, kaya_area_name,
			color_name, gym_name, board_name, is_gb_moderated, is_access_sensitive,
			is_closed, is_offensive, woulder_location_id, last_synced_at,
			created_at, updated_at
		FROM woulder.kaya_climbs
		WHERE slug = $1
	`

	queryGetClimbsByLocation = `
		SELECT id, kaya_climb_id, slug, name, grade_id, grade_name, grade_ordering,
			grade_climb_type_id, climb_type_id, climb_type_name, rating, ascent_count,
			kaya_destination_id, kaya_destination_name, kaya_area_id, kaya_area_name,
			color_name, gym_name, board_name, is_gb_moderated, is_access_sensitive,
			is_closed, is_offensive, woulder_location_id, last_synced_at,
			created_at, updated_at
		FROM woulder.kaya_climbs
		WHERE kaya_area_id = $1 OR kaya_destination_id = $1
		ORDER BY grade_ordering, name
	`

	queryGetClimbsByDestination = `
		SELECT id, kaya_climb_id, slug, name, grade_id, grade_name, grade_ordering,
			grade_climb_type_id, climb_type_id, climb_type_name, rating, ascent_count,
			kaya_destination_id, kaya_destination_name, kaya_area_id, kaya_area_name,
			color_name, gym_name, board_name, is_gb_moderated, is_access_sensitive,
			is_closed, is_offensive, woulder_location_id, last_synced_at,
			created_at, updated_at
		FROM woulder.kaya_climbs
		WHERE kaya_destination_id = $1
		ORDER BY grade_ordering, name
	`
)

// Ascents queries
const (
	querySaveAscent = `
		INSERT INTO woulder.kaya_ascents (
			kaya_ascent_id, kaya_climb_slug, kaya_user_id, date, comment, rating,
			stiffness, grade_id, grade_name, photo_url, photo_thumb_url,
			video_url, video_thumb_url
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (kaya_ascent_id) DO UPDATE SET
			comment = EXCLUDED.comment,
			rating = EXCLUDED.rating,
			stiffness = EXCLUDED.stiffness,
			grade_id = EXCLUDED.grade_id,
			grade_name = EXCLUDED.grade_name,
			photo_url = EXCLUDED.photo_url,
			photo_thumb_url = EXCLUDED.photo_thumb_url,
			video_url = EXCLUDED.video_url,
			video_thumb_url = EXCLUDED.video_thumb_url,
			updated_at = NOW()
	`

	queryGetAscentByID = `
		SELECT id, kaya_ascent_id, kaya_climb_slug, kaya_user_id, date, comment, rating,
			stiffness, grade_id, grade_name, photo_url, photo_thumb_url,
			video_url, video_thumb_url, created_at, updated_at
		FROM woulder.kaya_ascents
		WHERE kaya_ascent_id = $1
	`

	queryGetAscentsByClimb = `
		SELECT id, kaya_ascent_id, kaya_climb_slug, kaya_user_id, date, comment, rating,
			stiffness, grade_id, grade_name, photo_url, photo_thumb_url,
			video_url, video_thumb_url, created_at, updated_at
		FROM woulder.kaya_ascents
		WHERE kaya_climb_slug = $1
		ORDER BY date DESC
		LIMIT $2
	`

	queryGetAscentsByUser = `
		SELECT id, kaya_ascent_id, kaya_climb_slug, kaya_user_id, date, comment, rating,
			stiffness, grade_id, grade_name, photo_url, photo_thumb_url,
			video_url, video_thumb_url, created_at, updated_at
		FROM woulder.kaya_ascents
		WHERE kaya_user_id = $1
		ORDER BY date DESC
		LIMIT $2
	`

	queryGetRecentAscents = `
		SELECT id, kaya_ascent_id, kaya_climb_slug, kaya_user_id, date, comment, rating,
			stiffness, grade_id, grade_name, photo_url, photo_thumb_url,
			video_url, video_thumb_url, created_at, updated_at
		FROM woulder.kaya_ascents
		ORDER BY date DESC
		LIMIT $1
	`

	queryGetAscentsByWoulderLocation = `
		SELECT a.id, a.kaya_ascent_id, a.kaya_climb_slug, a.kaya_user_id, a.date, a.comment, a.rating,
			a.stiffness, a.grade_id, a.grade_name, a.photo_url, a.photo_thumb_url,
			a.video_url, a.video_thumb_url, a.created_at, a.updated_at
		FROM woulder.kaya_ascents a
		JOIN woulder.kaya_climbs c ON a.kaya_climb_slug = c.slug
		WHERE c.woulder_location_id = $1
		ORDER BY a.date DESC
		LIMIT $2
	`
)

// Posts queries
const (
	querySavePost = `
		INSERT INTO woulder.kaya_posts (kaya_post_id, kaya_user_id, date_created)
		VALUES ($1, $2, $3)
		ON CONFLICT (kaya_post_id) DO UPDATE SET
			date_created = EXCLUDED.date_created,
			updated_at = NOW()
	`

	querySavePostItem = `
		INSERT INTO woulder.kaya_post_items (
			kaya_post_item_id, kaya_post_id, kaya_climb_slug, kaya_ascent_id,
			photo_url, video_url, video_thumbnail_url, caption
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (kaya_post_item_id) DO UPDATE SET
			kaya_climb_slug = EXCLUDED.kaya_climb_slug,
			kaya_ascent_id = EXCLUDED.kaya_ascent_id,
			photo_url = EXCLUDED.photo_url,
			video_url = EXCLUDED.video_url,
			video_thumbnail_url = EXCLUDED.video_thumbnail_url,
			caption = EXCLUDED.caption,
			updated_at = NOW()
	`

	queryGetPostByID = `
		SELECT id, kaya_post_id, kaya_user_id, date_created, created_at, updated_at
		FROM woulder.kaya_posts
		WHERE kaya_post_id = $1
	`

	queryGetPostItemsByPost = `
		SELECT id, kaya_post_item_id, kaya_post_id, kaya_climb_slug, kaya_ascent_id,
			photo_url, video_url, video_thumbnail_url, caption, created_at, updated_at
		FROM woulder.kaya_post_items
		WHERE kaya_post_id = $1
		ORDER BY id
	`

	queryGetRecentPosts = `
		SELECT id, kaya_post_id, kaya_user_id, date_created, created_at, updated_at
		FROM woulder.kaya_posts
		ORDER BY date_created DESC
		LIMIT $1
	`
)

// Sync queries
const (
	querySaveSyncProgress = `
		INSERT INTO woulder.kaya_sync_progress (
			kaya_location_id, location_name, status, last_sync_at, next_sync_at,
			sync_error, climbs_synced, ascents_synced, sub_locations_synced
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (kaya_location_id) DO UPDATE SET
			location_name = EXCLUDED.location_name,
			status = EXCLUDED.status,
			last_sync_at = EXCLUDED.last_sync_at,
			next_sync_at = EXCLUDED.next_sync_at,
			sync_error = EXCLUDED.sync_error,
			climbs_synced = EXCLUDED.climbs_synced,
			ascents_synced = EXCLUDED.ascents_synced,
			sub_locations_synced = EXCLUDED.sub_locations_synced,
			updated_at = NOW()
	`

	queryGetSyncProgress = `
		SELECT id, kaya_location_id, location_name, status, last_sync_at, next_sync_at,
			sync_error, climbs_synced, ascents_synced, sub_locations_synced,
			created_at, updated_at
		FROM woulder.kaya_sync_progress
		WHERE kaya_location_id = $1
	`

	queryGetLocationsDueForSync = `
		SELECT id, kaya_location_id, location_name, status, last_sync_at, next_sync_at,
			sync_error, climbs_synced, ascents_synced, sub_locations_synced,
			created_at, updated_at
		FROM woulder.kaya_sync_progress
		WHERE status IN ('pending', 'failed')
			OR (status = 'completed' AND next_sync_at <= NOW())
		ORDER BY
			CASE status
				WHEN 'pending' THEN 1
				WHEN 'failed' THEN 2
				WHEN 'completed' THEN 3
			END,
			next_sync_at ASC NULLS FIRST
		LIMIT $1
	`

	queryUpdateSyncStatus = `
		UPDATE woulder.kaya_sync_progress
		SET status = $2, sync_error = $3, updated_at = NOW()
		WHERE kaya_location_id = $1
	`

	queryIncrementSyncCounters = `
		UPDATE woulder.kaya_sync_progress
		SET
			climbs_synced = climbs_synced + $2,
			ascents_synced = ascents_synced + $3,
			sub_locations_synced = sub_locations_synced + $4,
			updated_at = NOW()
		WHERE kaya_location_id = $1
	`
)
