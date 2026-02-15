package boulders

const (
	// queryGetProfile retrieves a single boulder drying profile.
	// Primary key lookup via mp_route_id - very fast.
	// Returns NULL-aware fields for optional data.
	queryGetProfile = `
		SELECT id, mp_route_id, tree_coverage_percent, rock_type_override,
		       last_sun_calc_at, sun_exposure_hours_cache, created_at, updated_at
		FROM woulder.boulder_drying_profiles
		WHERE mp_route_id = $1
	`

	// queryGetProfilesByIDs retrieves multiple profiles in one query.
	// Uses ANY($1) for efficient IN-list querying with array parameter.
	// Index: mp_route_id for fast lookups
	queryGetProfilesByIDs = `
		SELECT id, mp_route_id, tree_coverage_percent, rock_type_override,
		       last_sun_calc_at, sun_exposure_hours_cache, created_at, updated_at
		FROM woulder.boulder_drying_profiles
		WHERE mp_route_id = ANY($1)
	`

	// querySaveProfile upserts a boulder drying profile.
	// ON CONFLICT handles both insert and update cases.
	// Unique constraint: mp_route_id
	querySaveProfile = `
		INSERT INTO woulder.boulder_drying_profiles (
			mp_route_id, tree_coverage_percent, rock_type_override,
			last_sun_calc_at, sun_exposure_hours_cache
		)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (mp_route_id) DO UPDATE SET
			tree_coverage_percent = EXCLUDED.tree_coverage_percent,
			rock_type_override = EXCLUDED.rock_type_override,
			last_sun_calc_at = EXCLUDED.last_sun_calc_at,
			sun_exposure_hours_cache = EXCLUDED.sun_exposure_hours_cache,
			updated_at = NOW()
	`
)
