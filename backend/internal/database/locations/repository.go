// Package locations provides repository operations for climbing location management.
// Locations are specific climbing areas (crags, boulders) within geographic areas.
package locations

import (
	"context"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// Repository defines operations for location data management.
// All methods are safe for concurrent use.
type Repository interface {
	// GetAll retrieves all locations ordered by name.
	// Returns an empty slice if no locations are found.
	GetAll(ctx context.Context) ([]models.Location, error)

	// GetByID retrieves a specific location by its ID.
	// Returns database.ErrNotFound if the location does not exist.
	GetByID(ctx context.Context, id int) (*models.Location, error)

	// GetByArea retrieves all locations in a specific area.
	// Results are ordered by name.
	// Returns an empty slice if no locations are found in the area.
	GetByArea(ctx context.Context, areaID int) ([]models.Location, error)
}
