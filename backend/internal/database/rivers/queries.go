package rivers

// SQL queries for river operations.
// These queries are separated from implementation for clarity and reusability.

const (
	// queryGetByLocation retrieves all rivers for a location.
	// Ordered by river name for consistent results.
	// Index: location_id for efficient filtering
	queryGetByLocation = `
		SELECT id, location_id, gauge_id, river_name,
		       safe_crossing_cfs, caution_crossing_cfs,
		       drainage_area_sq_mi, gauge_drainage_area_sq_mi,
		       flow_divisor, is_estimated, description,
		       created_at, updated_at
		FROM woulder.rivers
		WHERE location_id = $1
		ORDER BY river_name
	`

	// queryGetByID retrieves a specific river by ID.
	// Primary key lookup - very fast.
	queryGetByID = `
		SELECT id, location_id, gauge_id, river_name,
		       safe_crossing_cfs, caution_crossing_cfs,
		       drainage_area_sq_mi, gauge_drainage_area_sq_mi,
		       flow_divisor, is_estimated, description,
		       created_at, updated_at
		FROM woulder.rivers
		WHERE id = $1
	`
)
