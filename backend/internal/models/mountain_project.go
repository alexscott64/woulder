package models

import "time"

// MPArea represents a Mountain Project area (hierarchical region containing routes/subareas)
type MPArea struct {
	ID             int        `json:"id" db:"id"`
	MPAreaID       int64      `json:"mp_area_id" db:"mp_area_id"`
	Name           string     `json:"name" db:"name"`
	ParentMPAreaID *int64     `json:"parent_mp_area_id,omitempty" db:"parent_mp_area_id"`
	AreaType       string     `json:"area_type" db:"area_type"`
	LocationID     *int       `json:"location_id,omitempty" db:"location_id"`
	Latitude       *float64   `json:"latitude,omitempty" db:"latitude"`
	Longitude      *float64   `json:"longitude,omitempty" db:"longitude"`
	LastSyncedAt   *time.Time `json:"last_synced_at,omitempty" db:"last_synced_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// MPRoute represents a Mountain Project route (boulder problem or climbing route)
type MPRoute struct {
	ID         int       `json:"id" db:"id"`
	MPRouteID  int64     `json:"mp_route_id" db:"mp_route_id"`
	MPAreaID   int64     `json:"mp_area_id" db:"mp_area_id"`
	Name       string    `json:"name" db:"name"`
	RouteType  string    `json:"route_type" db:"route_type"`
	Rating     string    `json:"rating" db:"rating"`
	LocationID *int      `json:"location_id,omitempty" db:"location_id"`
	Latitude   *float64  `json:"latitude,omitempty" db:"latitude"`
	Longitude  *float64  `json:"longitude,omitempty" db:"longitude"`
	Aspect     *string   `json:"aspect,omitempty" db:"aspect"` // N, NE, E, SE, S, SW, W, NW
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// MPTick represents a single climb log (tick) from a Mountain Project user
type MPTick struct {
	ID        int       `json:"id" db:"id"`
	MPRouteID int64     `json:"mp_route_id" db:"mp_route_id"`
	UserName  string    `json:"user_name" db:"user_name"`
	ClimbedAt time.Time `json:"climbed_at" db:"climbed_at"`
	Style     string    `json:"style" db:"style"`
	Comment   *string   `json:"comment,omitempty" db:"comment"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// BoulderDryingProfile stores boulder-specific drying metadata
type BoulderDryingProfile struct {
	ID                    int        `json:"id" db:"id"`
	MPRouteID             int64      `json:"mp_route_id" db:"mp_route_id"`
	TreeCoveragePercent   *float64   `json:"tree_coverage_percent,omitempty" db:"tree_coverage_percent"`
	RockTypeOverride      *string    `json:"rock_type_override,omitempty" db:"rock_type_override"`
	LastSunCalcAt         *time.Time `json:"last_sun_calc_at,omitempty" db:"last_sun_calc_at"`
	SunExposureHoursCache *string    `json:"sun_exposure_hours_cache,omitempty" db:"sun_exposure_hours_cache"` // JSONB stored as string
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at" db:"updated_at"`
}

// ClimbHistoryEntry represents a single climb from the location's history
// Used for API responses to show recent climbs at a location
type ClimbHistoryEntry struct {
	MPRouteID      int64     `json:"mp_route_id"` // Mountain Project route ID for linking
	RouteName      string    `json:"route_name"`
	RouteRating    string    `json:"route_rating"`
	MPAreaID       int64     `json:"mp_area_id"` // Mountain Project area ID for linking
	AreaName       string    `json:"area_name"`  // e.g., "Xyz Boulders"
	ClimbedAt      time.Time `json:"climbed_at"`
	ClimbedBy      string    `json:"climbed_by"`
	Style          string    `json:"style"`
	Comment        *string   `json:"comment,omitempty"`
	DaysSinceClimb int       `json:"days_since_climb"`
	Source         string    `json:"source"` // "mp" or "kaya"
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

// AreaDryingStats represents aggregated drying conditions for an area
type AreaDryingStats struct {
	TotalRoutes       int        `json:"total_routes"`                  // Total routes in area with GPS data
	DryCount          int        `json:"dry_count"`                     // Routes currently dry
	DryingCount       int        `json:"drying_count"`                  // Routes drying (<24h until dry)
	WetCount          int        `json:"wet_count"`                     // Routes wet (>24h until dry)
	PercentDry        float64    `json:"percent_dry"`                   // Percentage of routes dry (0-100)
	AvgHoursUntilDry  float64    `json:"avg_hours_until_dry"`           // Average hours until dry (wet routes only)
	AvgTreeCoverage   float64    `json:"avg_tree_coverage"`             // Average tree coverage % (0-100)
	ConfidenceScore   int        `json:"confidence_score"`              // Overall confidence (0-100)
	LastRainTimestamp *time.Time `json:"last_rain_timestamp,omitempty"` // Most recent rain timestamp from all routes
}

// AreaActivitySummary represents an area with recent activity metadata
// Used for API responses to show areas ordered by recent climbing activity
type AreaActivitySummary struct {
	MPAreaID       int64            `json:"mp_area_id"`                  // Mountain Project area ID
	Name           string           `json:"name"`                        // Area name
	ParentMPAreaID *int64           `json:"parent_mp_area_id,omitempty"` // Parent area ID (null for root)
	LastClimbAt    time.Time        `json:"last_climb_at"`               // Most recent climb timestamp
	TotalTicks     int              `json:"total_ticks"`                 // Total number of climbs
	UniqueRoutes   int              `json:"unique_routes"`               // Number of distinct routes with activity
	DaysSinceClimb int              `json:"days_since_climb"`            // Days since last climb
	HasSubareas    bool             `json:"has_subareas"`                // Whether this area has child subareas
	SubareaCount   int              `json:"subarea_count"`               // Number of direct child subareas
	DryingStats    *AreaDryingStats `json:"drying_stats,omitempty"`      // Aggregated drying conditions (optional)
}

// RouteActivitySummary represents a boulder with recent activity
// Used for API responses to show routes ordered by recent climbing activity
type RouteActivitySummary struct {
	MPRouteID      int64               `json:"mp_route_id"`                // Mountain Project route ID
	Name           string              `json:"name"`                       // Route name
	Rating         string              `json:"rating"`                     // Grade (V4, 5.10a, etc.)
	MPAreaID       int64               `json:"mp_area_id"`                 // Parent area ID
	LastClimbAt    time.Time           `json:"last_climb_at"`              // Most recent climb timestamp
	MostRecentTick *ClimbHistoryEntry  `json:"most_recent_tick,omitempty"` // Latest tick details (null if no ticks)
	RecentTicks    []ClimbHistoryEntry `json:"recent_ticks,omitempty"`     // Additional recent ticks (optional)
	DaysSinceClimb int                 `json:"days_since_climb"`           // Days since last climb
}

// SearchResult represents a unified search result that can be either an area or a route
// Used for API responses when searching across both areas and routes
type SearchResult struct {
	ResultType     string             `json:"result_type"`                // "area" or "route"
	ID             int64              `json:"id"`                         // Area or route ID
	Name           string             `json:"name"`                       // Area or route name
	Rating         *string            `json:"rating,omitempty"`           // Grade (only for routes)
	MPAreaID       int64              `json:"mp_area_id"`                 // Area ID (self for areas, parent for routes)
	AreaName       *string            `json:"area_name,omitempty"`        // Parent area name (only for routes)
	LastClimbAt    time.Time          `json:"last_climb_at"`              // Most recent climb timestamp
	DaysSinceClimb int                `json:"days_since_climb"`           // Days since last climb
	TotalTicks     *int               `json:"total_ticks,omitempty"`      // Total ticks (only for areas)
	UniqueRoutes   *int               `json:"unique_routes,omitempty"`    // Unique routes (only for areas)
	MostRecentTick *ClimbHistoryEntry `json:"most_recent_tick,omitempty"` // Latest tick (only for routes)
}
