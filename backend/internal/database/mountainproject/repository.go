// Package mountainproject handles all Mountain Project data operations including
// areas, routes, ticks, comments, and sync priority management.
package mountainproject

import (
	"context"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// Repository is a composite of all Mountain Project sub-repositories.
// It provides a unified interface for accessing MP data operations.
type Repository interface {
	Areas() AreasRepository
	Routes() RoutesRepository
	Ticks() TicksRepository
	Comments() CommentsRepository
	Sync() SyncRepository
}

// AreasRepository handles Mountain Project area operations.
type AreasRepository interface {
	// SaveArea saves or updates a Mountain Project area.
	SaveArea(ctx context.Context, area *models.MPArea) error

	// GetAreaByID retrieves a Mountain Project area by MP area ID.
	// Returns nil if not found.
	GetAreaByID(ctx context.Context, mpAreaID int64) (*models.MPArea, error)

	// UpdateRouteCount updates the cached route count for an area.
	UpdateRouteCount(ctx context.Context, mpAreaID string, total int) error

	// GetRouteCount retrieves the cached route count for an area.
	// Returns -1 if the area doesn't exist or hasn't been checked yet.
	GetRouteCount(ctx context.Context, mpAreaID string) (int, error)

	// GetChildAreas retrieves all direct children of an area.
	GetChildAreas(ctx context.Context, parentMPAreaID string) ([]ChildArea, error)

	// GetAllStateConfigs retrieves all state configurations ordered by display_order.
	GetAllStateConfigs(ctx context.Context) ([]StateConfig, error)
}

// RoutesRepository handles Mountain Project route operations.
type RoutesRepository interface {
	// SaveRoute saves or updates a Mountain Project route.
	SaveRoute(ctx context.Context, route *models.MPRoute) error

	// GetByID retrieves a Mountain Project route by MP route ID.
	// Returns nil if not found.
	GetByID(ctx context.Context, mpRouteID int64) (*models.MPRoute, error)

	// GetByIDs retrieves multiple Mountain Project routes by IDs in a single query.
	// This eliminates N+1 query problems when loading multiple routes.
	// Returns a map of route_id -> route. Missing routes are simply not in the map.
	GetByIDs(ctx context.Context, mpRouteIDs []int64) (map[int64]*models.MPRoute, error)

	// GetAllIDsForLocation returns all route IDs associated with a location.
	GetAllIDsForLocation(ctx context.Context, locationID int) ([]int64, error)

	// UpdateGPS updates only the GPS coordinates and aspect for a route.
	UpdateGPS(ctx context.Context, routeID int64, latitude, longitude float64, aspect string) error

	// GetIDsForArea retrieves all route IDs currently in an area.
	GetIDsForArea(ctx context.Context, mpAreaID string) ([]string, error)

	// GetWithGPSByArea retrieves all routes in an area (including subareas) that have GPS coordinates.
	// Uses recursive CTE to traverse area hierarchy.
	GetWithGPSByArea(ctx context.Context, mpAreaID int64) ([]*models.MPRoute, error)

	// UpsertRoute inserts or updates a route (compatibility with mountainprojectsync).
	UpsertRoute(ctx context.Context, mpRouteID, mpAreaID int64, locationID *int, name, routeType, rating string, lat, lon *float64, aspect *string) error

	// UpdateRouteDetails updates the detailed route information fields.
	UpdateRouteDetails(ctx context.Context, mpRouteID int64, difficulty *string, pitches *int, heightFeet *int, mpRating, popularity *float64, descriptionText, locationText, protectionText, safetyText *string) error
}

// TicksRepository handles Mountain Project tick operations.
type TicksRepository interface {
	// SaveTick saves a Mountain Project tick and updates the last_tick_sync_at timestamp.
	SaveTick(ctx context.Context, tick *models.MPTick) error

	// GetLastTimestampForRoute returns the timestamp of the most recent tick for a route.
	// Returns nil if no ticks exist.
	GetLastTimestampForRoute(ctx context.Context, routeID int64) (*time.Time, error)

	// UpsertTick inserts or updates a tick (compatibility with mountainprojectsync).
	UpsertTick(ctx context.Context, mpRouteID int64, userName string, climbedAt time.Time, style string, comment *string) error
}

// CommentsRepository handles Mountain Project comment operations.
type CommentsRepository interface {
	// SaveAreaComment saves a Mountain Project area comment.
	SaveAreaComment(ctx context.Context, mpCommentID, mpAreaID int64, userName, commentText string, commentedAt time.Time) error

	// SaveRouteComment saves a Mountain Project route comment and updates last_comment_sync_at.
	SaveRouteComment(ctx context.Context, mpCommentID, mpRouteID int64, userName, commentText string, commentedAt time.Time) error

	// UpsertAreaComment inserts or updates an area comment (compatibility with mountainprojectsync).
	UpsertAreaComment(ctx context.Context, mpCommentID, mpAreaID int64, userName string, userID *string, commentText string, commentedAt time.Time) error

	// UpsertRouteComment inserts or updates a route comment (compatibility with mountainprojectsync).
	UpsertRouteComment(ctx context.Context, mpCommentID, mpRouteID int64, userName string, userID *string, commentText string, commentedAt time.Time) error
}

// SyncRepository handles sync priority and scheduling operations.
type SyncRepository interface {
	// UpdateRoutePriorities recalculates sync priority for all NON-LOCATION routes.
	// Only updates routes WHERE location_id IS NULL (location routes always sync daily).
	UpdateRoutePriorities(ctx context.Context) error

	// GetLocationRoutesDueForSync returns ALL routes with location_id that need syncing.
	// These routes always sync daily regardless of activity.
	GetLocationRoutesDueForSync(ctx context.Context, syncType string) ([]int64, error)

	// GetRoutesDueForTickSync returns NON-LOCATION routes due for tick syncing based on priority.
	// Only returns routes WHERE location_id IS NULL.
	GetRoutesDueForTickSync(ctx context.Context, priority string) ([]int64, error)

	// GetRoutesDueForCommentSync returns NON-LOCATION routes due for comment syncing based on priority.
	// Only returns routes WHERE location_id IS NULL.
	GetRoutesDueForCommentSync(ctx context.Context, priority string) ([]int64, error)

	// GetPriorityDistribution returns count of routes in each priority tier (for monitoring).
	GetPriorityDistribution(ctx context.Context) (map[string]int, error)
}

// ChildArea represents a child area in the hierarchy.
type ChildArea struct {
	MPAreaID string
	Name     string
}

// StateConfig represents a state configuration for Mountain Project syncing.
type StateConfig struct {
	StateName string
	MPAreaID  string
	IsActive  bool
}
