package models

import (
	"encoding/json"
	"time"
)

// AnalyticsSession represents a visitor session on the site.
type AnalyticsSession struct {
	ID              int       `json:"id"`
	SessionID       string    `json:"session_id"`
	VisitorID       string    `json:"visitor_id"`
	IPAddress       *string   `json:"ip_address,omitempty"`
	UserAgent       *string   `json:"user_agent,omitempty"`
	Referrer        *string   `json:"referrer,omitempty"`
	Country         *string   `json:"country,omitempty"`
	Region          *string   `json:"region,omitempty"`
	City            *string   `json:"city,omitempty"`
	DeviceType      string    `json:"device_type"`
	Browser         *string   `json:"browser,omitempty"`
	OS              *string   `json:"os,omitempty"`
	ScreenWidth     *int      `json:"screen_width,omitempty"`
	ScreenHeight    *int      `json:"screen_height,omitempty"`
	StartedAt       time.Time `json:"started_at"`
	LastActiveAt    time.Time `json:"last_active_at"`
	PageCount       int       `json:"page_count"`
	DurationSeconds int       `json:"duration_seconds"`
	IsBounce        bool      `json:"is_bounce"`
	CreatedAt       time.Time `json:"created_at"`
}

// AnalyticsEvent represents a user interaction event.
type AnalyticsEvent struct {
	ID        int64           `json:"id"`
	SessionID string          `json:"session_id"`
	EventType string          `json:"event_type"`
	EventName string          `json:"event_name"`
	PagePath  *string         `json:"page_path,omitempty"`
	Metadata  json.RawMessage `json:"metadata,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

// AnalyticsAdminUser represents an admin user for the dashboard.
type AnalyticsAdminUser struct {
	ID           int        `json:"id"`
	Username     string     `json:"username"`
	PasswordHash string     `json:"-"`
	CreatedAt    time.Time  `json:"created_at"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
}

// --- Request/Response types ---

// CreateSessionRequest is sent from the frontend to create/update a session.
type CreateSessionRequest struct {
	SessionID    string  `json:"session_id"`
	VisitorID    string  `json:"visitor_id"`
	Referrer     *string `json:"referrer,omitempty"`
	DeviceType   string  `json:"device_type"`
	Browser      *string `json:"browser,omitempty"`
	OS           *string `json:"os,omitempty"`
	ScreenWidth  *int    `json:"screen_width,omitempty"`
	ScreenHeight *int    `json:"screen_height,omitempty"`
	UserAgent    *string `json:"user_agent,omitempty"`
}

// TrackEventRequest is sent from the frontend to record events.
type TrackEventRequest struct {
	SessionID string          `json:"session_id"`
	EventType string          `json:"event_type"`
	EventName string          `json:"event_name"`
	PagePath  *string         `json:"page_path,omitempty"`
	Metadata  json.RawMessage `json:"metadata,omitempty"`
}

// BatchEventsRequest sends multiple events in one request.
type BatchEventsRequest struct {
	SessionID string              `json:"session_id"`
	Events    []TrackEventRequest `json:"events"`
}

// HeartbeatRequest keeps a session alive.
type HeartbeatRequest struct {
	SessionID string `json:"session_id"`
}

// LoginRequest for admin authentication.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse with JWT token.
type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

// --- Metrics response types ---

// OverviewMetrics is the top-level dashboard summary.
type OverviewMetrics struct {
	UniqueVisitors     int     `json:"unique_visitors"`
	TotalSessions      int     `json:"total_sessions"`
	TotalPageViews     int     `json:"total_page_views"`
	AvgSessionDuration float64 `json:"avg_session_duration_seconds"`
	BounceRate         float64 `json:"bounce_rate"`
	TotalEvents        int     `json:"total_events"`
}

// VisitorDataPoint represents a single time-series data point.
type VisitorDataPoint struct {
	Date           string `json:"date"`
	UniqueVisitors int    `json:"unique_visitors"`
	Sessions       int    `json:"sessions"`
	PageViews      int    `json:"page_views"`
}

// TopPage represents a most-viewed page.
type TopPage struct {
	PagePath       string `json:"page_path"`
	ViewCount      int    `json:"view_count"`
	UniqueVisitors int    `json:"unique_visitors"`
}

// TopLocation represents a most-viewed climbing location.
type TopLocation struct {
	LocationID     int    `json:"location_id"`
	LocationName   string `json:"location_name"`
	ViewCount      int    `json:"view_count"`
	UniqueVisitors int    `json:"unique_visitors"`
}

// TopArea represents a most-viewed climbing area.
type TopArea struct {
	AreaID         string `json:"area_id"`
	AreaName       string `json:"area_name"`
	LocationID     int    `json:"location_id,omitempty"`
	ViewCount      int    `json:"view_count"`
	UniqueVisitors int    `json:"unique_visitors"`
}

// TopRoute represents a most-viewed route/boulder.
type TopRoute struct {
	RouteID        string `json:"route_id"`
	RouteName      string `json:"route_name"`
	RouteType      string `json:"route_type,omitempty"`
	ViewCount      int    `json:"view_count"`
	UniqueVisitors int    `json:"unique_visitors"`
}

// FeatureUsage represents usage count for a feature.
type FeatureUsage struct {
	FeatureName string `json:"feature_name"`
	UsageCount  int    `json:"usage_count"`
	UniqueUsers int    `json:"unique_users"`
}

// GeoLocation represents visitor geographic data.
type GeoLocation struct {
	Country        string `json:"country"`
	Region         string `json:"region,omitempty"`
	City           string `json:"city,omitempty"`
	VisitCount     int    `json:"visit_count"`
	UniqueVisitors int    `json:"unique_visitors"`
}

// DeviceBreakdown represents device/browser/OS stats.
type DeviceBreakdown struct {
	Name       string  `json:"name"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

// ReferrerInfo represents a traffic source.
type ReferrerInfo struct {
	Referrer       string `json:"referrer"`
	VisitCount     int    `json:"visit_count"`
	UniqueVisitors int    `json:"unique_visitors"`
}

// SessionDetail is an enriched session for the admin dashboard.
type SessionDetail struct {
	AnalyticsSession
	Events []AnalyticsEvent `json:"events,omitempty"`
}
