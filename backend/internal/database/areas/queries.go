package areas

const (
	// queryGetAll retrieves all active areas.
	// Ordered by display_order (manual sorting) then by name.
	// Index: is_active for filtering
	queryGetAll = `
		SELECT id, name, description, region,
		       display_order, is_active, created_at, updated_at
		FROM woulder.areas
		WHERE is_active = TRUE
		ORDER BY display_order, name
	`

	// queryGetAllWithLocationCounts retrieves areas with location counts.
	// LEFT JOIN ensures areas without locations are still returned.
	// Indexes: areas.is_active, locations.area_id for joins
	queryGetAllWithLocationCounts = `
		SELECT a.id, a.name, a.description, a.region,
		       a.display_order, a.is_active, a.created_at, a.updated_at,
		       COUNT(l.id) AS location_count
		FROM woulder.areas a
		LEFT JOIN woulder.locations l ON l.area_id = a.id
		WHERE a.is_active = TRUE
		GROUP BY a.id
		ORDER BY a.display_order, a.name
	`

	// queryGetByID retrieves a single active area by ID.
	// Primary key lookup with is_active filter - very fast.
	queryGetByID = `
		SELECT id, name, description, region,
		       display_order, is_active, created_at, updated_at
		FROM woulder.areas
		WHERE id = $1 AND is_active = TRUE
	`
)
