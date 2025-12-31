package models

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
}
