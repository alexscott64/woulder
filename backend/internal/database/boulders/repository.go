// Package boulders provides repository operations for boulder drying profile management.
// Boulder drying profiles store specific drying characteristics for individual boulders,
// including tree coverage, rock type overrides, and sun exposure calculations.
package boulders

import (
	"context"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// Repository defines operations for boulder drying profile management.
// All methods are safe for concurrent use.
type Repository interface {
	// GetProfile retrieves the drying profile for a specific boulder (route).
	// Returns nil if no profile exists (not an error - profile is optional).
	GetProfile(ctx context.Context, mpRouteID int64) (*models.BoulderDryingProfile, error)

	// GetProfilesByIDs retrieves multiple boulder drying profiles in a single query.
	// This eliminates N+1 query problems when loading profiles for many routes.
	// Returns a map of route_id -> profile. Missing profiles are simply not in the map.
	GetProfilesByIDs(ctx context.Context, mpRouteIDs []int64) (map[int64]*models.BoulderDryingProfile, error)

	// SaveProfile creates or updates a boulder drying profile.
	// Uses upsert logic based on mp_route_id.
	SaveProfile(ctx context.Context, profile *models.BoulderDryingProfile) error
}
