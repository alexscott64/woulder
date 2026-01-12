package models

import "time"

// Location represents a saved weather location
type Location struct {
	ID              int       `json:"id" db:"id"`
	Name            string    `json:"name" db:"name"`
	Latitude        float64   `json:"latitude" db:"latitude"`
	Longitude       float64   `json:"longitude" db:"longitude"`
	ElevationFt     int       `json:"elevation_ft" db:"elevation_ft"`         // Elevation in feet above sea level
	AreaID          int       `json:"area_id" db:"area_id"`                   // Foreign key to areas table
	HasSeepageRisk  bool      `json:"has_seepage_risk" db:"has_seepage_risk"` // Location has seepage/snowmelt issues
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// River represents a river crossing associated with a location
type River struct {
	ID                     int       `json:"id" db:"id"`
	LocationID             int       `json:"location_id" db:"location_id"`
	GaugeID                string    `json:"gauge_id" db:"gauge_id"`                               // USGS river gauge station ID (may be nearby if no direct gauge)
	RiverName              string    `json:"river_name" db:"river_name"`                           // Name of the river/creek for crossing
	SafeCrossingCFS        int       `json:"safe_crossing_cfs" db:"safe_crossing_cfs"`             // Safe crossing threshold in CFS
	CautionCrossingCFS     int       `json:"caution_crossing_cfs" db:"caution_crossing_cfs"`       // Caution threshold in CFS
	DrainageAreaSqMi       *float64  `json:"drainage_area_sq_mi" db:"drainage_area_sq_mi"`         // Drainage area for flow estimation
	GaugeDrainageAreaSqMi  *float64  `json:"gauge_drainage_area_sq_mi" db:"gauge_drainage_area_sq_mi"` // Reference gauge drainage area
	FlowDivisor            *float64  `json:"flow_divisor" db:"flow_divisor"`                       // Simple divisor for gauge value (e.g., 2.0 means gauge/2)
	IsEstimated            bool      `json:"is_estimated" db:"is_estimated"`                       // TRUE if flow is estimated
	Description            *string   `json:"description" db:"description"`                         // Additional notes about the crossing
	CreatedAt              time.Time `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time `json:"updated_at" db:"updated_at"`
}

// WeatherData represents weather information for a location
type WeatherData struct {
	ID            int       `json:"id" db:"id"`
	LocationID    int       `json:"location_id" db:"location_id"`
	Timestamp     time.Time `json:"timestamp" db:"timestamp"`
	Temperature   float64   `json:"temperature" db:"temperature"`       // Fahrenheit
	FeelsLike     float64   `json:"feels_like" db:"feels_like"`         // Fahrenheit
	Precipitation float64   `json:"precipitation" db:"precipitation"`   // inches
	Humidity      int       `json:"humidity" db:"humidity"`             // percentage
	WindSpeed     float64   `json:"wind_speed" db:"wind_speed"`         // mph
	WindDirection int       `json:"wind_direction" db:"wind_direction"` // degrees
	CloudCover    int       `json:"cloud_cover" db:"cloud_cover"`       // percentage
	Pressure      int       `json:"pressure" db:"pressure"`             // hPa
	Description   string    `json:"description" db:"description"`       // e.g. "light rain"
	Icon          string    `json:"icon" db:"icon"`                     // OpenWeatherMap icon code
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// DailySunTimes represents sunrise/sunset for a single day
type DailySunTimes struct {
	Date    string `json:"date"`    // Date in YYYY-MM-DD format
	Sunrise string `json:"sunrise"` // Sunrise time (ISO 8601)
	Sunset  string `json:"sunset"`  // Sunset time (ISO 8601)
}

// ClimbingCondition represents overall climbing conditions
type ClimbingCondition struct {
	Level   string   `json:"level"`   // "good", "marginal", "bad"
	Reasons []string `json:"reasons"` // Contributing factors
}

// PestConditions represents pest activity levels
type PestConditions struct {
	MosquitoLevel    string   `json:"mosquito_level"`     // "low", "moderate", "high", "very_high", "extreme"
	MosquitoScore    int      `json:"mosquito_score"`     // 0-100
	OutdoorPestLevel string   `json:"outdoor_pest_level"` // "low", "moderate", "high", "very_high", "extreme"
	OutdoorPestScore int      `json:"outdoor_pest_score"` // 0-100
	Factors          []string `json:"factors"`            // Contributing factors
}

// WeatherForecast represents forecast data
type WeatherForecast struct {
	LocationID       int                `json:"location_id"`
	Location         Location           `json:"location"`
	Current          WeatherData        `json:"current"`
	Hourly           []WeatherData      `json:"hourly"`
	Historical       []WeatherData      `json:"historical"`
	Sunrise          string             `json:"sunrise,omitempty"`              // Today's sunrise time (ISO 8601)
	Sunset           string             `json:"sunset,omitempty"`               // Today's sunset time (ISO 8601)
	DailySunTimes    []DailySunTimes    `json:"daily_sun_times,omitempty"`      // Sunrise/sunset for each forecast day
	RockDryingStatus *RockDryingStatus  `json:"rock_drying_status,omitempty"`   // Rock drying status
	SnowDepthInches  *float64           `json:"snow_depth_inches,omitempty"`    // Current snow depth on ground in inches
	DailySnowDepth   map[string]float64 `json:"daily_snow_depth,omitempty"`     // Snow depth forecast by date (YYYY-MM-DD)
	TodayCondition   *ClimbingCondition `json:"today_condition,omitempty"`      // Today's overall climbing condition
	RainLast48h      *float64           `json:"rain_last_48h,omitempty"`        // Total rain in last 48 hours (inches)
	RainNext48h      *float64           `json:"rain_next_48h,omitempty"`        // Forecast rain in next 48 hours (inches)
	PestConditions   *PestConditions      `json:"pest_conditions,omitempty"`      // Pest activity levels (mosquitoes, outdoor pests)
	LastClimbedInfo  *LastClimbedInfo     `json:"last_climbed_info,omitempty"`    // DEPRECATED: Most recent climb (use climb_history instead)
	ClimbHistory     []ClimbHistoryEntry  `json:"climb_history,omitempty"`        // Recent climb history at this location (from Mountain Project)
}

// RiverData represents river gauge information with current conditions
type RiverData struct {
	River           River   `json:"river"`             // River crossing info from database
	FlowCFS         float64 `json:"flow_cfs"`          // Current flow in cubic feet per second
	GaugeHeightFt   float64 `json:"gauge_height_ft"`   // Current gauge height in feet
	IsSafe          bool    `json:"is_safe"`           // Whether it's safe to cross
	Status          string  `json:"status"`            // "safe", "caution", "unsafe"
	StatusMessage   string  `json:"status_message"`    // Human-readable status message
	Timestamp       string  `json:"timestamp"`         // When the data was recorded
	PercentOfSafe   float64 `json:"percent_of_safe"`   // Current flow as percentage of safe threshold
}
