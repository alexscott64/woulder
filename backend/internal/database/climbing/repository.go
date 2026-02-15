// Package climbing provides repository operations for climbing activity and history.
// This domain handles activity queries, climb history, and search functionality
// for routes and areas. All activity queries use smart date filtering to handle
// data quality issues (future-dated ticks, typos in years).
package climbing

import (
	"context"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// Repository defines operations for climbing activity and history.
// All methods are safe for concurrent use.
type Repository interface {
	// History operations
	History() HistoryRepository

	// Activity operations
	Activity() ActivityRepository

	// Search operations
	Search() SearchRepository
}

// HistoryRepository handles climb history operations.
type HistoryRepository interface {
	// GetLastClimbedForLocation retrieves the most recent climb for a location.
	// DEPRECATED: Use GetClimbHistoryForLocation with limit=1 instead.
	// Returns nil if no climbs exist for the location.
	GetLastClimbedForLocation(ctx context.Context, locationID int) (*models.LastClimbedInfo, error)

	// GetClimbHistoryForLocation retrieves recent climb history for a location.
	// Includes smart data quality filtering:
	// - Adjusts dates ~1 year in the future (likely typo: 2026 -> 2025)
	// - Filters out dates >30 days in the future (bad data)
	// - Only shows climbs from the past 2 years
	// Results ordered by climbed_at descending (most recent first).
	GetClimbHistoryForLocation(ctx context.Context, locationID int, limit int) ([]models.ClimbHistoryEntry, error)

	// GetClimbHistoryForLocations retrieves recent climb history for multiple locations in a single query.
	// More efficient than calling GetClimbHistoryForLocation multiple times.
	// Returns a map of locationID -> climb history entries.
	GetClimbHistoryForLocations(ctx context.Context, locationIDs []int, limit int) (map[int][]models.ClimbHistoryEntry, error)
}

// ActivityRepository handles area and route activity queries.
type ActivityRepository interface {
	// GetAreasOrderedByActivity retrieves top-level areas ordered by recent activity.
	// Shows children of the root area with aggregated activity from all descendants.
	// Uses smart date filtering. Results ordered by most recent activity.
	GetAreasOrderedByActivity(ctx context.Context, locationID int) ([]models.AreaActivitySummary, error)

	// GetSubareasOrderedByActivity retrieves subareas of a parent ordered by activity.
	// Recursively aggregates activity from all descendant areas.
	// Uses smart date filtering. Results ordered by most recent activity.
	GetSubareasOrderedByActivity(ctx context.Context, parentAreaID int64, locationID int) ([]models.AreaActivitySummary, error)

	// GetRoutesOrderedByActivity retrieves ALL routes in an area ordered by activity.
	// Shows routes with ticks first (by recency), then routes without ticks (alphabetically).
	// Includes the most recent tick for each route if it has any.
	// Uses smart date filtering.
	GetRoutesOrderedByActivity(ctx context.Context, areaID int64, locationID int, limit int) ([]models.RouteActivitySummary, error)

	// GetRecentTicksForRoute retrieves the most recent ticks for a specific route.
	// Uses smart date filtering. Results ordered by climbed_at descending.
	GetRecentTicksForRoute(ctx context.Context, routeID int64, limit int) ([]models.ClimbHistoryEntry, error)
}

// SearchRepository handles search operations for routes and areas.
type SearchRepository interface {
	// SearchInLocation searches all areas and routes in a location by name.
	// Returns unified search results (both areas and routes) ordered by recent activity.
	// Uses smart date filtering. Case-insensitive partial match.
	SearchInLocation(ctx context.Context, locationID int, searchQuery string, limit int) ([]models.SearchResult, error)

	// SearchRoutesInLocation searches routes in a location by name, rating, or area name.
	// Returns routes ordered by most recent climb activity.
	// Uses smart date filtering. Case-insensitive partial match.
	SearchRoutesInLocation(ctx context.Context, locationID int, searchQuery string, limit int) ([]models.RouteActivitySummary, error)
}
