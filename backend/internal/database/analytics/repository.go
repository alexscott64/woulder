// Package analytics provides repository operations for site analytics tracking.
// Tracks visitor sessions, page views, and user interactions for the /iglooghost CMS dashboard.
package analytics

import (
	"context"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// Repository defines operations for analytics data.
// All methods are safe for concurrent use.
type Repository interface {
	// --- Session operations ---

	// CreateSession creates a new visitor session.
	CreateSession(ctx context.Context, session *models.AnalyticsSession) error

	// UpdateSessionActivity updates last_active_at, page_count, duration, and bounce status.
	UpdateSessionActivity(ctx context.Context, sessionID string) error

	// GetSessionByID retrieves a session by its UUID.
	GetSessionByID(ctx context.Context, sessionID string) (*models.AnalyticsSession, error)

	// --- Event operations ---

	// InsertEvent records a single analytics event.
	InsertEvent(ctx context.Context, event *models.AnalyticsEvent) error

	// InsertEvents records multiple analytics events in a batch.
	InsertEvents(ctx context.Context, events []models.AnalyticsEvent) error

	// --- Admin user operations ---

	// GetAdminByUsername retrieves an admin user by username.
	GetAdminByUsername(ctx context.Context, username string) (*models.AnalyticsAdminUser, error)

	// UpsertAdmin creates or updates an admin user.
	UpsertAdmin(ctx context.Context, username, passwordHash string) error

	// UpdateLastLogin updates the last login timestamp for an admin.
	UpdateLastLogin(ctx context.Context, username string) error

	// --- Metrics queries ---

	// GetOverviewMetrics returns high-level dashboard metrics for a time period.
	GetOverviewMetrics(ctx context.Context, since time.Time) (*models.OverviewMetrics, error)

	// GetVisitorsOverTime returns daily visitor/session/pageview counts.
	GetVisitorsOverTime(ctx context.Context, since time.Time) ([]models.VisitorDataPoint, error)

	// GetTopPages returns the most viewed pages.
	GetTopPages(ctx context.Context, since time.Time, limit int) ([]models.TopPage, error)

	// GetTopLocations returns the most viewed climbing locations.
	GetTopLocations(ctx context.Context, since time.Time, limit int) ([]models.TopLocation, error)

	// GetTopAreas returns the most viewed climbing areas.
	GetTopAreas(ctx context.Context, since time.Time, limit int) ([]models.TopArea, error)

	// GetTopRoutes returns the most viewed routes/boulders.
	GetTopRoutes(ctx context.Context, since time.Time, limit int) ([]models.TopRoute, error)

	// GetFeatureUsage returns feature usage breakdown.
	GetFeatureUsage(ctx context.Context, since time.Time) ([]models.FeatureUsage, error)

	// GetGeography returns visitor geographic distribution.
	GetGeography(ctx context.Context, since time.Time, limit int) ([]models.GeoLocation, error)

	// GetDeviceBreakdown returns device type distribution.
	GetDeviceBreakdown(ctx context.Context, since time.Time) ([]models.DeviceBreakdown, error)

	// GetBrowserBreakdown returns browser distribution.
	GetBrowserBreakdown(ctx context.Context, since time.Time) ([]models.DeviceBreakdown, error)

	// GetOSBreakdown returns OS distribution.
	GetOSBreakdown(ctx context.Context, since time.Time) ([]models.DeviceBreakdown, error)

	// GetReferrers returns top referrer sources.
	GetReferrers(ctx context.Context, since time.Time, limit int) ([]models.ReferrerInfo, error)

	// GetRecentSessions returns recent sessions with details.
	GetRecentSessions(ctx context.Context, limit int) ([]models.SessionDetail, error)

	// GetSessionEvents returns events for a specific session.
	GetSessionEvents(ctx context.Context, sessionID string) ([]models.AnalyticsEvent, error)

	// --- Cleanup ---

	// CleanOldData removes analytics data older than the given duration.
	CleanOldData(ctx context.Context, olderThan time.Duration) error
}
