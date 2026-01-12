package models

import "time"

// MPArea represents a Mountain Project area (hierarchical region containing routes/subareas)
type MPArea struct {
	ID             int        `json:"id" db:"id"`
	MPAreaID       string     `json:"mp_area_id" db:"mp_area_id"`
	Name           string     `json:"name" db:"name"`
	ParentMPAreaID *string    `json:"parent_mp_area_id,omitempty" db:"parent_mp_area_id"`
	AreaType       string     `json:"area_type" db:"area_type"`
	LocationID     *int       `json:"location_id,omitempty" db:"location_id"`
	LastSyncedAt   *time.Time `json:"last_synced_at,omitempty" db:"last_synced_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// MPRoute represents a Mountain Project route (boulder problem or climbing route)
type MPRoute struct {
	ID         int       `json:"id" db:"id"`
	MPRouteID  string    `json:"mp_route_id" db:"mp_route_id"`
	MPAreaID   string    `json:"mp_area_id" db:"mp_area_id"`
	Name       string    `json:"name" db:"name"`
	RouteType  string    `json:"route_type" db:"route_type"`
	Rating     string    `json:"rating" db:"rating"`
	LocationID *int      `json:"location_id,omitempty" db:"location_id"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// MPTick represents a single climb log (tick) from a Mountain Project user
type MPTick struct {
	ID        int       `json:"id" db:"id"`
	MPRouteID string    `json:"mp_route_id" db:"mp_route_id"`
	UserName  string    `json:"user_name" db:"user_name"`
	ClimbedAt time.Time `json:"climbed_at" db:"climbed_at"`
	Style     string    `json:"style" db:"style"`
	Comment   *string   `json:"comment,omitempty" db:"comment"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ClimbHistoryEntry represents a single climb from the location's history
// Used for API responses to show recent climbs at a location
type ClimbHistoryEntry struct {
	MPRouteID      string    `json:"mp_route_id"`       // Mountain Project route ID for linking
	RouteName      string    `json:"route_name"`
	RouteRating    string    `json:"route_rating"`
	MPAreaID       string    `json:"mp_area_id"`        // Mountain Project area ID for linking
	AreaName       string    `json:"area_name"`         // e.g., "Xyz Boulders"
	ClimbedAt      time.Time `json:"climbed_at"`
	ClimbedBy      string    `json:"climbed_by"`
	Style          string    `json:"style"`
	Comment        *string   `json:"comment,omitempty"`
	DaysSinceClimb int       `json:"days_since_climb"`
}

// LastClimbedInfo is an aggregated view of the most recent climb at a location
// Used for API responses to show when a route was last climbed
// DEPRECATED: Use ClimbHistoryEntry instead
type LastClimbedInfo struct {
	RouteName      string    `json:"route_name"`
	RouteRating    string    `json:"route_rating"`
	ClimbedAt      time.Time `json:"climbed_at"`
	ClimbedBy      string    `json:"climbed_by"`
	Style          string    `json:"style"`
	Comment        *string   `json:"comment,omitempty"`
	DaysSinceClimb int       `json:"days_since_climb"`
}
