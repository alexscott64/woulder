package locations

const (
	// queryGetAll retrieves all locations.
	// Ordered by name for consistent display.
	// No filtering - returns all active locations.
	queryGetAll = `
		SELECT id, name, latitude, longitude, elevation_ft, area_id,
		       has_seepage_risk, created_at, updated_at
		FROM woulder.locations
		ORDER BY name
	`

	// queryGetByID retrieves a single location by ID.
	// Primary key lookup - very fast.
	queryGetByID = `
		SELECT id, name, latitude, longitude, elevation_ft, area_id,
		       has_seepage_risk, created_at, updated_at
		FROM woulder.locations
		WHERE id = $1
	`

	// queryGetByArea retrieves all locations in a specific area.
	// Ordered by name for consistent display.
	// Index: area_id for efficient filtering
	queryGetByArea = `
		SELECT id, name, latitude, longitude, elevation_ft, area_id,
		       has_seepage_risk, created_at, updated_at
		FROM woulder.locations
		WHERE area_id = $1
		ORDER BY name
	`
)
