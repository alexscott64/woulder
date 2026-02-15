// Package rivers provides repository operations for river data management.
// Rivers are associated with climbing locations and include gauge information
// for assessing water flow and crossing safety.
package rivers

import (
	"context"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// Repository defines operations for river data management.
// All methods are safe for concurrent use.
type Repository interface {
	// GetByLocation retrieves all rivers associated with a location.
	// Returns an empty slice if no rivers are found for the location.
	GetByLocation(ctx context.Context, locationID int) ([]models.River, error)

	// GetByID retrieves a specific river by its ID.
	// Returns database.ErrNotFound if the river does not exist.
	GetByID(ctx context.Context, id int) (*models.River, error)
}
