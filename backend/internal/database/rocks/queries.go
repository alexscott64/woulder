package rocks

const (
	// queryGetRockTypesByLocation retrieves all rock types for a location.
	// Joins through location_rock_types to get location-specific associations.
	// Includes rock type group information for categorization.
	// Indexes: location_rock_types(location_id), rock_types(id), rock_type_groups(id)
	queryGetRockTypesByLocation = `
		SELECT rt.id, rt.name, rt.base_drying_hours,
		       rt.porosity_percent, rt.is_wet_sensitive,
		       rt.description, rt.rock_type_group_id,
		       rtg.group_name
		FROM woulder.rock_types rt
		INNER JOIN woulder.location_rock_types lrt ON rt.id = lrt.rock_type_id
		INNER JOIN woulder.rock_type_groups rtg ON rt.rock_type_group_id = rtg.id
		WHERE lrt.location_id = $1
		ORDER BY lrt.is_primary DESC, rt.name ASC
	`

	// queryGetPrimaryRockType retrieves the primary rock type for a location.
	// Uses is_primary flag to identify the main rock type.
	// Index: location_rock_types(location_id, is_primary)
	queryGetPrimaryRockType = `
		SELECT rt.id, rt.name, rt.base_drying_hours,
		       rt.porosity_percent, rt.is_wet_sensitive,
		       rt.description, rt.rock_type_group_id,
		       rtg.group_name
		FROM woulder.rock_types rt
		INNER JOIN woulder.location_rock_types lrt ON rt.id = lrt.rock_type_id
		INNER JOIN woulder.rock_type_groups rtg ON rt.rock_type_group_id = rtg.id
		WHERE lrt.location_id = $1 AND lrt.is_primary = TRUE
		LIMIT 1
	`

	// queryGetSunExposureByLocation retrieves sun exposure data for a location.
	// Contains directional exposure percentages and features like tree coverage.
	// Primary key lookup via location_id - very fast.
	queryGetSunExposureByLocation = `
		SELECT id, location_id,
		       south_facing_percent, west_facing_percent,
		       east_facing_percent, north_facing_percent,
		       slab_percent, overhang_percent,
		       tree_coverage_percent, description
		FROM woulder.location_sun_exposure
		WHERE location_id = $1
	`
)
