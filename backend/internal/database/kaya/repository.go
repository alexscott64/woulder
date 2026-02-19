// Package kaya handles all Kaya data operations including
// locations, climbs, ascents, users, posts, and sync progress management.
package kaya

import (
	"context"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// KayaAscentWithDetails represents an ascent with denormalized climb and user data
// This is returned by the optimized GetAscentsWithDetailsForWoulderLocation query
type KayaAscentWithDetails struct {
	KayaAscentID  string
	KayaClimbSlug string
	Date          time.Time
	Comment       *string
	ClimbName     string
	ClimbGrade    *string
	AreaName      string
	Username      string
}

// Repository is a composite of all Kaya sub-repositories.
// It provides a unified interface for accessing Kaya data operations.
type Repository interface {
	Users() UsersRepository
	Locations() LocationsRepository
	Climbs() ClimbsRepository
	Ascents() AscentsRepository
	Posts() PostsRepository
	Sync() SyncRepository
}

// UsersRepository handles Kaya user operations.
type UsersRepository interface {
	// SaveUser saves or updates a Kaya user.
	SaveUser(ctx context.Context, user *models.KayaUser) error

	// GetUserByID retrieves a Kaya user by Kaya user ID.
	// Returns nil if not found.
	GetUserByID(ctx context.Context, kayaUserID string) (*models.KayaUser, error)

	// GetUserByUsername retrieves a Kaya user by username.
	// Returns nil if not found.
	GetUserByUsername(ctx context.Context, username string) (*models.KayaUser, error)
}

// LocationsRepository handles Kaya location operations.
type LocationsRepository interface {
	// SaveLocation saves or updates a Kaya location.
	SaveLocation(ctx context.Context, location *models.KayaLocation) error

	// GetLocationByID retrieves a Kaya location by Kaya location ID.
	// Returns nil if not found.
	GetLocationByID(ctx context.Context, kayaLocationID string) (*models.KayaLocation, error)

	// GetLocationBySlug retrieves a Kaya location by slug.
	// Returns nil if not found.
	GetLocationBySlug(ctx context.Context, slug string) (*models.KayaLocation, error)

	// GetSubLocations retrieves all direct children of a location.
	GetSubLocations(ctx context.Context, parentKayaLocationID string) ([]*models.KayaLocation, error)

	// GetAllLocations retrieves all locations ordered by name.
	GetAllLocations(ctx context.Context) ([]*models.KayaLocation, error)
}

// ClimbsRepository handles Kaya climb operations.
type ClimbsRepository interface {
	// SaveClimb saves or updates a Kaya climb.
	SaveClimb(ctx context.Context, climb *models.KayaClimb) error

	// GetClimbBySlug retrieves a Kaya climb by slug.
	// Returns nil if not found.
	GetClimbBySlug(ctx context.Context, slug string) (*models.KayaClimb, error)

	// GetClimbsByLocation retrieves all climbs for a location.
	GetClimbsByLocation(ctx context.Context, kayaLocationID string) ([]*models.KayaClimb, error)

	// GetClimbsByDestination retrieves all climbs for a destination.
	GetClimbsByDestination(ctx context.Context, kayaDestinationID string) ([]*models.KayaClimb, error)

	// GetClimbsOrderedByActivityForWoulderLocation retrieves climbs by recent activity for a Woulder location.
	GetClimbsOrderedByActivityForWoulderLocation(ctx context.Context, woulderLocationID int, limit int) ([]models.UnifiedRouteActivitySummary, error)

	// GetMatchedClimbsForArea retrieves Kaya climbs that have been matched to MP routes in a specific area
	GetMatchedClimbsForArea(ctx context.Context, mpAreaID int64, limit int) ([]models.UnifiedRouteActivitySummary, error)
}

// AscentsRepository handles Kaya ascent operations.
type AscentsRepository interface {
	// SaveAscent saves or updates a Kaya ascent.
	SaveAscent(ctx context.Context, ascent *models.KayaAscent) error

	// GetAscentByID retrieves a Kaya ascent by Kaya ascent ID.
	// Returns nil if not found.
	GetAscentByID(ctx context.Context, kayaAscentID string) (*models.KayaAscent, error)

	// GetAscentsByClimb retrieves all ascents for a climb.
	GetAscentsByClimb(ctx context.Context, kayaClimbSlug string, limit int) ([]*models.KayaAscent, error)

	// GetAscentsByUser retrieves all ascents for a user.
	GetAscentsByUser(ctx context.Context, kayaUserID string, limit int) ([]*models.KayaAscent, error)

	// GetRecentAscents retrieves recent ascents across all climbs.
	GetRecentAscents(ctx context.Context, limit int) ([]*models.KayaAscent, error)

	// GetAscentsByWoulderLocation retrieves ascents for climbs at a Woulder location.
	GetAscentsByWoulderLocation(ctx context.Context, woulderLocationID int, limit int) ([]*models.KayaAscent, error)

	// GetAscentsWithDetailsForWoulderLocation retrieves ascents with climb and user details in a single query.
	// This is optimized to avoid N+1 queries by joining climbs and users.
	GetAscentsWithDetailsForWoulderLocation(ctx context.Context, woulderLocationID int, limit int) ([]KayaAscentWithDetails, error)

	// GetAscentsForMatchedRoute retrieves Kaya ascents for routes matched to a specific MP route
	GetAscentsForMatchedRoute(ctx context.Context, mpRouteID int64, limit int) ([]KayaAscentWithDetails, error)
}

// PostsRepository handles Kaya post operations.
type PostsRepository interface {
	// SavePost saves or updates a Kaya post.
	SavePost(ctx context.Context, post *models.KayaPost) error

	// SavePostItem saves or updates a Kaya post item.
	SavePostItem(ctx context.Context, item *models.KayaPostItem) error

	// GetPostByID retrieves a Kaya post by Kaya post ID.
	// Returns nil if not found.
	GetPostByID(ctx context.Context, kayaPostID string) (*models.KayaPost, error)

	// GetPostItemsByPost retrieves all items for a post.
	GetPostItemsByPost(ctx context.Context, kayaPostID string) ([]*models.KayaPostItem, error)

	// GetRecentPosts retrieves recent posts.
	GetRecentPosts(ctx context.Context, limit int) ([]*models.KayaPost, error)
}

// SyncRepository handles sync progress and scheduling operations.
type SyncRepository interface {
	// SaveSyncProgress saves or updates sync progress for a location.
	SaveSyncProgress(ctx context.Context, progress *models.KayaSyncProgress) error

	// GetSyncProgress retrieves sync progress for a location.
	// Returns nil if not found.
	GetSyncProgress(ctx context.Context, kayaLocationID string) (*models.KayaSyncProgress, error)

	// GetLocationsDueForSync retrieves locations that need syncing.
	GetLocationsDueForSync(ctx context.Context, limit int) ([]*models.KayaSyncProgress, error)

	// UpdateSyncStatus updates the sync status for a location.
	UpdateSyncStatus(ctx context.Context, kayaLocationID string, status string, syncError *string) error

	// IncrementSyncCounters increments the sync counters for a location.
	IncrementSyncCounters(ctx context.Context, kayaLocationID string, climbs, ascents, subLocations int) error
}
