package models

import "time"

// RockTypeGroup represents a category of rock types with similar drying characteristics
type RockTypeGroup struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// RockType represents a type of climbing rock with drying characteristics
type RockType struct {
	ID              int     `json:"id"`
	Name            string  `json:"name"`
	BaseDryingHours float64 `json:"base_drying_hours"` // Hours to dry after 0.1" rain in ideal conditions
	PorosityPercent float64 `json:"porosity_percent"`  // Average porosity percentage
	IsWetSensitive  bool    `json:"is_wet_sensitive"`  // True for sandstone/soft rocks
	Description     string  `json:"description"`
	RockTypeGroupID int     `json:"rock_type_group_id"`
	GroupName       string  `json:"group_name,omitempty"` // Populated via JOIN
}

// LocationRockType represents the association between a location and its rock types
type LocationRockType struct {
	LocationID int  `json:"location_id"`
	RockTypeID int  `json:"rock_type_id"`
	IsPrimary  bool `json:"is_primary"` // Primary rock type for the location
}

// LocationSunExposure represents sun exposure profile for a climbing location
type LocationSunExposure struct {
	ID                  int     `json:"id" db:"id"`
	LocationID          int     `json:"location_id" db:"location_id"`
	SouthFacingPercent  float64 `json:"south_facing_percent" db:"south_facing_percent"`   // 0-100
	WestFacingPercent   float64 `json:"west_facing_percent" db:"west_facing_percent"`     // 0-100
	EastFacingPercent   float64 `json:"east_facing_percent" db:"east_facing_percent"`     // 0-100
	NorthFacingPercent  float64 `json:"north_facing_percent" db:"north_facing_percent"`   // 0-100
	SlabPercent         float64 `json:"slab_percent" db:"slab_percent"`                   // 0-100
	OverhangPercent     float64 `json:"overhang_percent" db:"overhang_percent"`           // 0-100
	TreeCoveragePercent float64 `json:"tree_coverage_percent" db:"tree_coverage_percent"` // 0-100
	Description         string  `json:"description,omitempty" db:"description"`
}

// RainEvent represents a contiguous rain event derived from weather data
type RainEvent struct {
	StartTime     time.Time // First precipitation reading
	EndTime       time.Time // Last precipitation reading
	TotalRain     float64   // Total rainfall in inches
	Duration      float64   // Duration in hours
	MaxHourlyRate float64   // Maximum hourly rate (inches/hour)
	AvgHourlyRate float64   // Average hourly rate (inches/hour)
}

// RockDryingStatus represents the current drying state of rock at a location
type RockDryingStatus struct {
	IsWet             bool     `json:"is_wet"`              // Currently wet
	IsSafe            bool     `json:"is_safe"`             // Safe to climb
	IsWetSensitive    bool     `json:"is_wet_sensitive"`    // Contains wet-sensitive rock
	HoursUntilDry     float64  `json:"hours_until_dry"`     // Estimated hours until dry
	LastRainTimestamp string   `json:"last_rain_timestamp"` // When it last rained
	Status            string   `json:"status"`              // "critical", "poor", "fair", "good"
	Message           string   `json:"message"`             // Human-readable message
	RockTypes         []string `json:"rock_types"`          // List of rock type names
	PrimaryRockType   string   `json:"primary_rock_type"`   // Primary rock type name (specific type)
	PrimaryGroupName  string   `json:"primary_group_name"`  // Primary rock type group name (display name)
	ConfidenceScore   int      `json:"confidence_score"`    // 0-100 confidence in this prediction
}

// RockTemperatureStatus represents the current rock surface temperature conditions
// produced by the rock_temp calculator (heat + condensation friction analysis).
type RockTemperatureStatus struct {
	EstimatedSurfaceTempF float64           `json:"estimated_surface_temp_f"`
	AirTempF              float64           `json:"air_temp_f"`
	TempDifferentialF     float64           `json:"temp_differential_f"`
	Condition             string            `json:"condition"`        // "prime","good","marginal","poor","very_poor","too_cold"
	FrictionQuality       string            `json:"friction_quality"` // "excellent","good","reduced","poor"
	NextTransition        *Transition       `json:"next_transition,omitempty"`
	Message               string            `json:"message"`
	SendWindows           []SendWindow      `json:"send_windows,omitempty"`
	HourlyForecast        []RockTempHour    `json:"hourly_forecast,omitempty"`
	DailyForecast         []DailyRockTemp   `json:"daily_forecast,omitempty"`
	Condensation          *CondensationInfo `json:"condensation,omitempty"`
	ConfidenceScore       int               `json:"confidence_score"`
	ConfidenceFactors     []string          `json:"confidence_factors,omitempty"`
	RockType              string            `json:"rock_type"`
}

// Transition is the next condition-tier change in the forecast.
type Transition struct {
	Time        time.Time `json:"time"`
	ToCondition string    `json:"to_condition"`
}

// CondensationInfo describes wet-from-dewpoint conditions
// (separate from rock_drying's wet-from-rain logic).
type CondensationInfo struct {
	Active            bool       `json:"active"` // T_rock < T_dewpoint right now
	DewpointF         float64    `json:"dewpoint_f"`
	SurfaceVsDewpoint float64    `json:"surface_vs_dewpoint"` // negative = condensing, positive = dry
	ClearsAt          *time.Time `json:"clears_at,omitempty"`
	Severity          string     `json:"severity"` // "none","light","heavy"
	Reason            string     `json:"reason"`
}

// SendWindow is a contiguous period of good climbing conditions.
type SendWindow struct {
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	DurationH     float64   `json:"duration_h"`
	Condition     string    `json:"condition"` // "prime" or "good"
	AvgTempF      float64   `json:"avg_temp_f"`
	PeakTempF     float64   `json:"peak_temp_f"`
	DryThroughout bool      `json:"dry_throughout"` // false if any hour has light condensation
}

// RockTempHour is the per-hour rock temperature output.
type RockTempHour struct {
	Time       time.Time `json:"time"`
	SurfaceF   float64   `json:"surface_f"`
	AirF       float64   `json:"air_f"`
	DewpointF  float64   `json:"dewpoint_f"`
	Condensing bool      `json:"condensing"`
	Condition  string    `json:"condition"`
}

// DailyRockTemp summarizes one calendar day of rock surface temperature
// conditions for a location. Times are in the location's local timezone
// at the boundary; the LocalDate field is YYYY-MM-DD in that timezone.
type DailyRockTemp struct {
	LocalDate        string      `json:"local_date"`                 // YYYY-MM-DD, location-local
	PeakSurfaceTempF float64     `json:"peak_surface_temp_f"`        // hottest hour of the day
	MinSurfaceTempF  float64     `json:"min_surface_temp_f"`         // coldest hour of the day
	PeakCondition    string      `json:"peak_condition"`             // tier of peak hour
	OverallCondition string      `json:"overall_condition"`          // worst tier observed (best->worst: prime,good,marginal,poor,very_poor; too_cold treated separately)
	HasCondensation  bool        `json:"has_condensation"`           // any hour was condensing
	BestSendWindow   *SendWindow `json:"best_send_window,omitempty"` // single best window of the day, prime preferred over good
	WindowCount      int         `json:"window_count"`               // count of all windows on this day
}
