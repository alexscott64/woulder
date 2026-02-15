// Package areas provides repository operations for geographic area management.
// Areas represent major climbing regions (e.g., Pacific Northwest, Southwest)
// and contain multiple locations.
package areas

import (
	"context"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// Repository defines operations for area data management.
// All methods are safe for concurrent use.
type Repository interface {
	// GetAll retrieves all active areas ordered by display order and name.
	// Returns an empty slice if no areas are found.
	GetAll(ctx context.Context) ([]models.Area, error)

	// GetAllWithLocationCounts retrieves all active areas with their associated location counts.
	// Useful for displaying area statistics in UI.
	// Returns an empty slice if no areas are found.
	GetAllWithLocationCounts(ctx context.Context) ([]models.AreaWithLocationCount, error)

	// GetByID retrieves a specific active area by its ID.
	// Returns database.ErrNotFound if the area does not exist or is inactive.
	GetByID(ctx context.Context, id int) (*models.Area, error)
}
