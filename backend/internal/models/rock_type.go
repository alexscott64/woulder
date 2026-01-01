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
	PorosityPercent float64 `json:"porosity_percent"`   // Average porosity percentage
	IsWetSensitive  bool    `json:"is_wet_sensitive"`   // True for sandstone/soft rocks
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
	ID                 int     `json:"id" db:"id"`
	LocationID         int     `json:"location_id" db:"location_id"`
	SouthFacingPercent float64 `json:"south_facing_percent" db:"south_facing_percent"` // 0-100
	WestFacingPercent  float64 `json:"west_facing_percent" db:"west_facing_percent"`   // 0-100
	EastFacingPercent  float64 `json:"east_facing_percent" db:"east_facing_percent"`   // 0-100
	NorthFacingPercent float64 `json:"north_facing_percent" db:"north_facing_percent"` // 0-100
	SlabPercent        float64 `json:"slab_percent" db:"slab_percent"`                 // 0-100
	OverhangPercent    float64 `json:"overhang_percent" db:"overhang_percent"`         // 0-100
	TreeCoveragePercent float64 `json:"tree_coverage_percent" db:"tree_coverage_percent"` // 0-100
	Description        string  `json:"description,omitempty" db:"description"`
}

// RainEvent represents a contiguous rain event derived from weather data
type RainEvent struct {
	StartTime      time.Time // First precipitation reading
	EndTime        time.Time // Last precipitation reading
	TotalRain      float64   // Total rainfall in inches
	Duration       float64   // Duration in hours
	MaxHourlyRate  float64   // Maximum hourly rate (inches/hour)
	AvgHourlyRate  float64   // Average hourly rate (inches/hour)
}

// RockDryingStatus represents the current drying state of rock at a location
type RockDryingStatus struct {
	IsWet              bool     `json:"is_wet"`                // Currently wet
	IsSafe             bool     `json:"is_safe"`               // Safe to climb
	IsWetSensitive     bool     `json:"is_wet_sensitive"`      // Contains wet-sensitive rock
	HoursUntilDry      float64  `json:"hours_until_dry"`       // Estimated hours until dry
	LastRainTimestamp  string   `json:"last_rain_timestamp"`   // When it last rained
	Status             string   `json:"status"`                // "critical", "poor", "fair", "good"
	Message            string   `json:"message"`               // Human-readable message
	RockTypes          []string `json:"rock_types"`            // List of rock type names
	PrimaryRockType    string   `json:"primary_rock_type"`     // Primary rock type name (specific type)
	PrimaryGroupName   string   `json:"primary_group_name"`    // Primary rock type group name (display name)
	ConfidenceScore    int      `json:"confidence_score"`      // 0-100 confidence in this prediction
}
