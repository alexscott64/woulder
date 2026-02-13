package models

import "time"

// HeatMapPoint represents a geographic point with climbing activity
type HeatMapPoint struct {
	MPAreaID       int64     `json:"mp_area_id"`
	Name           string    `json:"name"`
	Latitude       float64   `json:"latitude"`
	Longitude      float64   `json:"longitude"`
	ActivityScore  int       `json:"activity_score"` // Weighted: recent=higher
	TotalTicks     int       `json:"total_ticks"`
	ActiveRoutes   int       `json:"active_routes"`
	LastActivity   time.Time `json:"last_activity"`
	UniqueClimbers int       `json:"unique_climbers"`
	HasSubareas    bool      `json:"has_subareas"`
}

// AreaActivityDetail provides comprehensive activity data for an area
// Reuses existing ClimbHistoryEntry, CommentSummary, and RouteActivitySummary types
type AreaActivityDetail struct {
	MPAreaID         int64             `json:"mp_area_id"`
	Name             string            `json:"name"`
	ParentMPAreaID   *int64            `json:"parent_mp_area_id,omitempty"`
	Latitude         *float64          `json:"latitude,omitempty"`
	Longitude        *float64          `json:"longitude,omitempty"`
	TotalTicks       int               `json:"total_ticks"`
	ActiveRoutes     int               `json:"active_routes"`
	UniqueClimbers   int               `json:"unique_climbers"`
	LastActivity     time.Time         `json:"last_activity"`
	RecentTicks      []TickDetail      `json:"recent_ticks"`
	RecentComments   []CommentSummary  `json:"recent_comments"`
	ActivityTimeline []DailyActivity   `json:"activity_timeline"`
	TopRoutes        []TopRouteSummary `json:"top_routes"`
}

// TickDetail represents a single tick for area detail views
type TickDetail struct {
	MPRouteID int64     `json:"mp_route_id"`
	RouteName string    `json:"route_name"`
	Rating    string    `json:"rating"`
	UserName  string    `json:"user_name"`
	ClimbedAt time.Time `json:"climbed_at"`
	Style     string    `json:"style"`
	Comment   string    `json:"comment"`
}

// CommentSummary represents a comment with basic info
type CommentSummary struct {
	ID          int64     `json:"id"`
	UserName    string    `json:"user_name"`
	CommentText string    `json:"comment_text"`
	CommentedAt time.Time `json:"commented_at"`
	MPRouteID   *int64    `json:"mp_route_id,omitempty"`
	RouteName   *string   `json:"route_name,omitempty"`
}

// DailyActivity represents activity count for a single day
type DailyActivity struct {
	Date       time.Time `json:"date"`
	TickCount  int       `json:"tick_count"`
	RouteCount int       `json:"route_count"`
}

// RouteActivity represents a route with location and activity
type RouteActivity struct {
	MPRouteID    int64     `json:"mp_route_id"`
	Name         string    `json:"name"`
	Rating       string    `json:"rating"`
	Latitude     *float64  `json:"latitude"`
	Longitude    *float64  `json:"longitude"`
	TickCount    int       `json:"tick_count"`
	LastActivity time.Time `json:"last_activity"`
	MPAreaID     int64     `json:"mp_area_id"`
	AreaName     string    `json:"area_name"`
}

// TopRouteSummary is a simplified route with activity count for top routes lists
type TopRouteSummary struct {
	MPRouteID    int64     `json:"mp_route_id"`
	Name         string    `json:"name"`
	Rating       string    `json:"rating"`
	TickCount    int       `json:"tick_count"`
	LastActivity time.Time `json:"last_activity"`
}
