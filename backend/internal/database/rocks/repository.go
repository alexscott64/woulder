// Package rocks provides repository operations for rock type management.
// Rock types describe the climbing surface characteristics (granite, sandstone, etc.)
// which affect drying times and climbing conditions.
package rocks

import (
	"context"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// Repository defines operations for rock type data management.
// All methods are safe for concurrent use.
type Repository interface {
	// GetRockTypesByLocation retrieves all rock types for a location.
	// Results are ordered by primary rock type first, then alphabetically.
	// Returns an empty slice if no rock types are found.
	GetRockTypesByLocation(ctx context.Context, locationID int) ([]models.RockType, error)

	// GetPrimaryRockType retrieves the primary rock type for a location.
	// Falls back to the first rock type if no primary is explicitly set.
	// Returns database.ErrNotFound if no rock types exist for the location.
	GetPrimaryRockType(ctx context.Context, locationID int) (*models.RockType, error)

	// GetSunExposureByLocation retrieves sun exposure profile for a location.
	// Contains directional exposure percentages and tree coverage data.
	// Returns nil if no sun exposure data exists for the location.
	GetSunExposureByLocation(ctx context.Context, locationID int) (*models.LocationSunExposure, error)
}
