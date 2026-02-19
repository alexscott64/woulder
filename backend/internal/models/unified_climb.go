package models

import "time"

// UnifiedRouteActivitySummary represents a route/climb from either MP or Kaya with activity data
// This is used to show mixed results in the "By Area" view
type UnifiedRouteActivitySummary struct {
	// Common fields (present for both MP and Kaya)
	ID             string    `json:"id"`               // Composite ID: "mp-{routeID}" or "kaya-{slug}"
	Name           string    `json:"name"`             // Route/climb name
	Rating         string    `json:"rating"`           // Grade (V4, 5.10a, etc.)
	AreaName       string    `json:"area_name"`        // Parent area/location name
	LastClimbAt    time.Time `json:"last_climb_at"`    // Most recent ascent timestamp
	DaysSinceClimb int       `json:"days_since_climb"` // Days since last ascent
	Source         string    `json:"source"`           // "mp" or "kaya"

	// MP-specific fields (null for Kaya)
	MPRouteID      *int64             `json:"mp_route_id,omitempty"`      // Mountain Project route ID
	MPAreaID       *int64             `json:"mp_area_id,omitempty"`       // Parent MP area ID
	MostRecentTick *ClimbHistoryEntry `json:"most_recent_tick,omitempty"` // Latest MP tick details

	// Kaya-specific fields (null for MP)
	KayaClimbSlug    *string            `json:"kaya_climb_slug,omitempty"`    // Kaya climb slug
	MostRecentAscent *KayaAscentSummary `json:"most_recent_ascent,omitempty"` // Latest Kaya ascent details
}

// KayaAscentSummary represents a Kaya ascent in a simplified format for route lists
type KayaAscentSummary struct {
	KayaAscentID string    `json:"kaya_ascent_id"`
	ClimbedAt    time.Time `json:"climbed_at"`
	ClimbedBy    string    `json:"climbed_by"` // Username
	Comment      *string   `json:"comment,omitempty"`
	GradeName    *string   `json:"grade_name,omitempty"` // User's perceived grade
}
