package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// PostgresRepository implements Repository using PostgreSQL.
type PostgresRepository struct {
	db DBConn
}

// NewPostgresRepository creates a new PostgreSQL analytics repository.
func NewPostgresRepository(db DBConn) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// --- Session operations ---

// CreateSession creates a new visitor session.
func (r *PostgresRepository) CreateSession(ctx context.Context, session *models.AnalyticsSession) error {
	_, err := r.db.ExecContext(ctx, queryCreateSession,
		session.SessionID, session.VisitorID, session.IPAddress, session.UserAgent,
		session.Referrer, session.DeviceType, session.Browser, session.OS,
		session.ScreenWidth, session.ScreenHeight,
	)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

// UpdateSessionActivity updates session activity metrics.
func (r *PostgresRepository) UpdateSessionActivity(ctx context.Context, sessionID string) error {
	_, err := r.db.ExecContext(ctx, queryUpdateSessionActivity, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update session activity: %w", err)
	}
	return nil
}

// GetSessionByID retrieves a session by its UUID.
func (r *PostgresRepository) GetSessionByID(ctx context.Context, sessionID string) (*models.AnalyticsSession, error) {
	var s models.AnalyticsSession
	err := r.db.QueryRowContext(ctx, queryGetSessionByID, sessionID).Scan(
		&s.ID, &s.SessionID, &s.VisitorID, &s.IPAddress, &s.UserAgent, &s.Referrer,
		&s.Country, &s.Region, &s.City, &s.DeviceType, &s.Browser, &s.OS,
		&s.ScreenWidth, &s.ScreenHeight, &s.StartedAt, &s.LastActiveAt,
		&s.PageCount, &s.DurationSeconds, &s.IsBounce, &s.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return &s, nil
}

// --- Event operations ---

// InsertEvent records a single analytics event.
func (r *PostgresRepository) InsertEvent(ctx context.Context, event *models.AnalyticsEvent) error {
	metadata := event.Metadata
	if metadata == nil {
		metadata = json.RawMessage("{}")
	}
	_, err := r.db.ExecContext(ctx, queryInsertEvent,
		event.SessionID, event.EventType, event.EventName, event.PagePath, metadata,
	)
	if err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}
	return nil
}

// InsertEvents records multiple analytics events in a batch.
func (r *PostgresRepository) InsertEvents(ctx context.Context, events []models.AnalyticsEvent) error {
	for _, event := range events {
		if err := r.InsertEvent(ctx, &event); err != nil {
			return err
		}
	}
	return nil
}

// --- Admin user operations ---

// GetAdminByUsername retrieves an admin user by username.
func (r *PostgresRepository) GetAdminByUsername(ctx context.Context, username string) (*models.AnalyticsAdminUser, error) {
	var user models.AnalyticsAdminUser
	err := r.db.QueryRowContext(ctx, queryGetAdminByUsername, username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.LastLoginAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin user: %w", err)
	}
	return &user, nil
}

// UpsertAdmin creates or updates an admin user.
func (r *PostgresRepository) UpsertAdmin(ctx context.Context, username, passwordHash string) error {
	_, err := r.db.ExecContext(ctx, queryUpsertAdmin, username, passwordHash)
	if err != nil {
		return fmt.Errorf("failed to upsert admin: %w", err)
	}
	return nil
}

// UpdateLastLogin updates the last login timestamp for an admin.
func (r *PostgresRepository) UpdateLastLogin(ctx context.Context, username string) error {
	_, err := r.db.ExecContext(ctx, queryUpdateLastLogin, username)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	return nil
}

// --- Metrics queries ---

// GetOverviewMetrics returns high-level dashboard metrics for a time period.
func (r *PostgresRepository) GetOverviewMetrics(ctx context.Context, since time.Time) (*models.OverviewMetrics, error) {
	var m models.OverviewMetrics
	err := r.db.QueryRowContext(ctx, queryOverviewMetrics, since).Scan(
		&m.UniqueVisitors, &m.TotalSessions, &m.TotalPageViews,
		&m.AvgSessionDuration, &m.BounceRate, &m.TotalEvents,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get overview metrics: %w", err)
	}
	return &m, nil
}

// GetVisitorsOverTime returns daily visitor/session/pageview counts.
func (r *PostgresRepository) GetVisitorsOverTime(ctx context.Context, since time.Time) ([]models.VisitorDataPoint, error) {
	rows, err := r.db.QueryContext(ctx, queryVisitorsOverTime, since)
	if err != nil {
		return nil, fmt.Errorf("failed to query visitors over time: %w", err)
	}
	defer rows.Close()

	var result []models.VisitorDataPoint
	for rows.Next() {
		var dp models.VisitorDataPoint
		if err := rows.Scan(&dp.Date, &dp.UniqueVisitors, &dp.Sessions, &dp.PageViews); err != nil {
			return nil, fmt.Errorf("failed to scan visitor data point: %w", err)
		}
		result = append(result, dp)
	}
	return result, rows.Err()
}

// GetTopPages returns the most viewed pages.
func (r *PostgresRepository) GetTopPages(ctx context.Context, since time.Time, limit int) ([]models.TopPage, error) {
	rows, err := r.db.QueryContext(ctx, queryTopPages, since, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top pages: %w", err)
	}
	defer rows.Close()

	var result []models.TopPage
	for rows.Next() {
		var p models.TopPage
		if err := rows.Scan(&p.PagePath, &p.ViewCount, &p.UniqueVisitors); err != nil {
			return nil, fmt.Errorf("failed to scan top page: %w", err)
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

// GetTopLocations returns the most viewed climbing locations.
func (r *PostgresRepository) GetTopLocations(ctx context.Context, since time.Time, limit int) ([]models.TopLocation, error) {
	rows, err := r.db.QueryContext(ctx, queryTopLocations, since, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top locations: %w", err)
	}
	defer rows.Close()

	var result []models.TopLocation
	for rows.Next() {
		var loc models.TopLocation
		var locationIDStr string
		if err := rows.Scan(&locationIDStr, &loc.LocationName, &loc.ViewCount, &loc.UniqueVisitors); err != nil {
			return nil, fmt.Errorf("failed to scan top location: %w", err)
		}
		// Parse location ID from string (comes from JSONB)
		fmt.Sscanf(locationIDStr, "%d", &loc.LocationID)
		result = append(result, loc)
	}
	return result, rows.Err()
}

// GetTopAreas returns the most viewed climbing areas.
func (r *PostgresRepository) GetTopAreas(ctx context.Context, since time.Time, limit int) ([]models.TopArea, error) {
	rows, err := r.db.QueryContext(ctx, queryTopAreas, since, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top areas: %w", err)
	}
	defer rows.Close()

	var result []models.TopArea
	for rows.Next() {
		var a models.TopArea
		if err := rows.Scan(&a.AreaID, &a.AreaName, &a.LocationID, &a.ViewCount, &a.UniqueVisitors); err != nil {
			return nil, fmt.Errorf("failed to scan top area: %w", err)
		}
		result = append(result, a)
	}
	return result, rows.Err()
}

// GetTopRoutes returns the most viewed routes/boulders.
func (r *PostgresRepository) GetTopRoutes(ctx context.Context, since time.Time, limit int) ([]models.TopRoute, error) {
	rows, err := r.db.QueryContext(ctx, queryTopRoutes, since, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top routes: %w", err)
	}
	defer rows.Close()

	var result []models.TopRoute
	for rows.Next() {
		var route models.TopRoute
		if err := rows.Scan(&route.RouteID, &route.RouteName, &route.RouteType, &route.ViewCount, &route.UniqueVisitors); err != nil {
			return nil, fmt.Errorf("failed to scan top route: %w", err)
		}
		result = append(result, route)
	}
	return result, rows.Err()
}

// GetFeatureUsage returns feature usage breakdown.
func (r *PostgresRepository) GetFeatureUsage(ctx context.Context, since time.Time) ([]models.FeatureUsage, error) {
	rows, err := r.db.QueryContext(ctx, queryFeatureUsage, since)
	if err != nil {
		return nil, fmt.Errorf("failed to query feature usage: %w", err)
	}
	defer rows.Close()

	var result []models.FeatureUsage
	for rows.Next() {
		var f models.FeatureUsage
		if err := rows.Scan(&f.FeatureName, &f.UsageCount, &f.UniqueUsers); err != nil {
			return nil, fmt.Errorf("failed to scan feature usage: %w", err)
		}
		result = append(result, f)
	}
	return result, rows.Err()
}

// GetGeography returns visitor geographic distribution.
func (r *PostgresRepository) GetGeography(ctx context.Context, since time.Time, limit int) ([]models.GeoLocation, error) {
	rows, err := r.db.QueryContext(ctx, queryGeography, since, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query geography: %w", err)
	}
	defer rows.Close()

	var result []models.GeoLocation
	for rows.Next() {
		var g models.GeoLocation
		if err := rows.Scan(&g.Country, &g.Region, &g.City, &g.VisitCount, &g.UniqueVisitors); err != nil {
			return nil, fmt.Errorf("failed to scan geography: %w", err)
		}
		result = append(result, g)
	}
	return result, rows.Err()
}

// GetDeviceBreakdown returns device type distribution.
func (r *PostgresRepository) GetDeviceBreakdown(ctx context.Context, since time.Time) ([]models.DeviceBreakdown, error) {
	return r.getBreakdown(ctx, queryDeviceBreakdown, since)
}

// GetBrowserBreakdown returns browser distribution.
func (r *PostgresRepository) GetBrowserBreakdown(ctx context.Context, since time.Time) ([]models.DeviceBreakdown, error) {
	return r.getBreakdown(ctx, queryBrowserBreakdown, since)
}

// GetOSBreakdown returns OS distribution.
func (r *PostgresRepository) GetOSBreakdown(ctx context.Context, since time.Time) ([]models.DeviceBreakdown, error) {
	return r.getBreakdown(ctx, queryOSBreakdown, since)
}

// getBreakdown is a helper for aggregation queries that return name + count.
func (r *PostgresRepository) getBreakdown(ctx context.Context, query string, since time.Time) ([]models.DeviceBreakdown, error) {
	rows, err := r.db.QueryContext(ctx, query, since)
	if err != nil {
		return nil, fmt.Errorf("failed to query breakdown: %w", err)
	}
	defer rows.Close()

	var result []models.DeviceBreakdown
	var total int
	for rows.Next() {
		var d models.DeviceBreakdown
		if err := rows.Scan(&d.Name, &d.Count); err != nil {
			return nil, fmt.Errorf("failed to scan breakdown: %w", err)
		}
		total += d.Count
		result = append(result, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Calculate percentages
	for i := range result {
		if total > 0 {
			result[i].Percentage = float64(result[i].Count) / float64(total) * 100
		}
	}
	return result, nil
}

// GetReferrers returns top referrer sources.
func (r *PostgresRepository) GetReferrers(ctx context.Context, since time.Time, limit int) ([]models.ReferrerInfo, error) {
	rows, err := r.db.QueryContext(ctx, queryReferrers, since, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query referrers: %w", err)
	}
	defer rows.Close()

	var result []models.ReferrerInfo
	for rows.Next() {
		var ref models.ReferrerInfo
		if err := rows.Scan(&ref.Referrer, &ref.VisitCount, &ref.UniqueVisitors); err != nil {
			return nil, fmt.Errorf("failed to scan referrer: %w", err)
		}
		result = append(result, ref)
	}
	return result, rows.Err()
}

// GetRecentSessions returns recent sessions with details.
func (r *PostgresRepository) GetRecentSessions(ctx context.Context, limit int) ([]models.SessionDetail, error) {
	rows, err := r.db.QueryContext(ctx, queryRecentSessions, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent sessions: %w", err)
	}
	defer rows.Close()

	var result []models.SessionDetail
	for rows.Next() {
		var sd models.SessionDetail
		s := &sd.AnalyticsSession
		if err := rows.Scan(
			&s.ID, &s.SessionID, &s.VisitorID, &s.IPAddress, &s.UserAgent, &s.Referrer,
			&s.Country, &s.Region, &s.City, &s.DeviceType, &s.Browser, &s.OS,
			&s.ScreenWidth, &s.ScreenHeight, &s.StartedAt, &s.LastActiveAt,
			&s.PageCount, &s.DurationSeconds, &s.IsBounce, &s.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		result = append(result, sd)
	}
	return result, rows.Err()
}

// GetSessionEvents returns events for a specific session.
func (r *PostgresRepository) GetSessionEvents(ctx context.Context, sessionID string) ([]models.AnalyticsEvent, error) {
	rows, err := r.db.QueryContext(ctx, querySessionEvents, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query session events: %w", err)
	}
	defer rows.Close()

	var result []models.AnalyticsEvent
	for rows.Next() {
		var e models.AnalyticsEvent
		if err := rows.Scan(&e.ID, &e.SessionID, &e.EventType, &e.EventName, &e.PagePath, &e.Metadata, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		result = append(result, e)
	}
	return result, rows.Err()
}

// --- Cleanup ---

// CleanOldData removes analytics data older than the given duration.
func (r *PostgresRepository) CleanOldData(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)

	// Delete events first (FK constraint)
	if _, err := r.db.ExecContext(ctx, queryCleanOldEvents, cutoff); err != nil {
		return fmt.Errorf("failed to clean old events: %w", err)
	}

	// Then delete sessions
	if _, err := r.db.ExecContext(ctx, queryCleanOldSessions, cutoff); err != nil {
		return fmt.Errorf("failed to clean old sessions: %w", err)
	}

	return nil
}
