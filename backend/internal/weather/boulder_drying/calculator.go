package boulder_drying

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/alexscott64/woulder/backend/internal/weather/sun"
)

// DryingForecastPeriod represents dry/wet status for a time period
type DryingForecastPeriod struct {
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	IsDry         bool      `json:"is_dry"`
	Status        string    `json:"status"` // "dry", "drying", "wet"
	HoursUntilDry float64   `json:"hours_until_dry,omitempty"` // Only present if wet
	RainAmount    float64   `json:"rain_amount,omitempty"`     // Inches of rain in this period
}

// BoulderDryingStatus represents the drying status for a specific boulder
type BoulderDryingStatus struct {
	MPRouteID           int64                   `json:"mp_route_id"`
	IsWet               bool                    `json:"is_wet"`
	IsSafe              bool                    `json:"is_safe"`
	HoursUntilDry       float64                 `json:"hours_until_dry"`
	Status              string                  `json:"status"` // "critical", "poor", "fair", "good"
	Message             string                  `json:"message"`
	ConfidenceScore     int                     `json:"confidence_score"` // 0-100
	LastRainTimestamp   *time.Time              `json:"last_rain_timestamp,omitempty"` // Pointer to allow null
	SunExposureHours    float64                 `json:"sun_exposure_hours"`     // Hours of direct sun next 6 days
	TreeCoveragePercent float64                 `json:"tree_coverage_percent"` // 0-100
	RockType            string                  `json:"rock_type"`
	Aspect              string                  `json:"aspect"`      // N, NE, E, SE, S, SW, W, NW
	Latitude            float64                 `json:"latitude"`    // Boulder GPS
	Longitude           float64                 `json:"longitude"`   // Boulder GPS
	Forecast            []DryingForecastPeriod  `json:"forecast,omitempty"` // 6-day dry/wet forecast
}

// Calculator computes boulder-specific drying times
type Calculator struct {
	treeClient *TreeCoverClient
}

// NewCalculator creates a new boulder drying calculator
// Note: apiKey parameter is deprecated and ignored (kept for backwards compatibility)
func NewCalculator(apiKey string) *Calculator {
	return &Calculator{
		treeClient: NewTreeCoverClient(),
	}
}

// CalculateBoulderDryingStatus computes the drying status for a specific boulder
// Extends location-level rock drying with boulder-specific factors:
// - Real-time sun exposure based on boulder GPS + aspect
// - Boulder-specific tree coverage from satellite data
// - Confidence scoring based on data availability
// - 6-day dry/wet forecast based on precipitation forecast
// locationTreeCoverage: optional location-level tree coverage percentage (0 to use GPS-based estimates)
// hourlyForecast: optional hourly weather forecast for 6-day forecast (pass nil to skip forecast)
func (c *Calculator) CalculateBoulderDryingStatus(
	ctx context.Context,
	route *models.MPRoute,
	locationDrying *models.RockDryingStatus,
	profile *models.BoulderDryingProfile,
	locationTreeCoverage float64,
	hourlyForecast []models.WeatherData,
) (*BoulderDryingStatus, error) {
	status := &BoulderDryingStatus{
		MPRouteID:         route.MPRouteID,
		RockType:          locationDrying.PrimaryRockType,
		ConfidenceScore:   100, // Start at full confidence, reduce for missing data
	}

	// Extract GPS coordinates
	if route.Latitude != nil && route.Longitude != nil {
		status.Latitude = *route.Latitude
		status.Longitude = *route.Longitude
	} else {
		// Missing GPS coordinates - reduce confidence
		status.ConfidenceScore -= 30
		log.Printf("Warning: Route %s missing GPS coordinates", route.MPRouteID)
	}

	// Extract aspect
	if route.Aspect != nil {
		status.Aspect = *route.Aspect
	} else {
		// Missing aspect - reduce confidence, default to South (most sun)
		status.Aspect = "S"
		status.ConfidenceScore -= 20
		log.Printf("Warning: Route %s missing aspect, defaulting to South", route.MPRouteID)
	}

	// Get tree coverage (from profile cache ONLY - never fetch during request)
	// Tree coverage should be pre-populated by background job
	if profile != nil && profile.TreeCoveragePercent != nil {
		status.TreeCoveragePercent = *profile.TreeCoveragePercent
	} else {
		// Use location-level tree coverage as fallback (never call external API during request)
		if locationTreeCoverage > 0 {
			status.TreeCoveragePercent = locationTreeCoverage
		} else {
			status.TreeCoveragePercent = 30.0 // Default if no location data
			status.ConfidenceScore -= 15
		}
	}

	// Get sun exposure hours (always calculate fresh - takes milliseconds)
	sunStart := time.Now()
	if status.Latitude != 0 && status.Longitude != 0 {
		sunHours, err := c.calculateSunExposure(ctx, status.Latitude, status.Longitude, status.Aspect, status.TreeCoveragePercent, profile)
		if err != nil {
			log.Printf("Warning: Failed to calculate sun exposure for %s: %v", route.MPRouteID, err)
			// Fall back to location-level sun exposure estimate
			status.SunExposureHours = c.estimateSunExposureFromAspect(status.Aspect)
			status.ConfidenceScore -= 25
		} else {
			status.SunExposureHours = sunHours
		}
	} else {
		// No GPS - use aspect-based estimate
		status.SunExposureHours = c.estimateSunExposureFromAspect(status.Aspect)
		status.ConfidenceScore -= 25
	}
	sunTime := time.Since(sunStart)
	if sunTime > 10*time.Millisecond {
		log.Printf("[PERF]     Sun exposure calculation took %v (should be <10ms)", sunTime)
	}

	// Calculate boulder-specific drying time
	status.HoursUntilDry = c.calculateBoulderDryingTime(locationDrying, status)

	// Parse LastRainTimestamp string to time.Time (avoid zero time values)
	if locationDrying.LastRainTimestamp != "" {
		lastRain, err := time.Parse(time.RFC3339, locationDrying.LastRainTimestamp)
		if err == nil {
			status.LastRainTimestamp = &lastRain
		}
	} else {
		// If no last rain timestamp, set to nil (will be omitted from JSON)
		status.LastRainTimestamp = nil
	}

	// Determine wet/safe/status
	status.IsWet = status.HoursUntilDry > 0
	status.IsSafe = !status.IsWet || !locationDrying.IsWetSensitive
	status.Status = c.determineDryingStatus(status)
	status.Message = c.generateStatusMessage(status, locationDrying)

	// Calculate 6-day forecast if hourly forecast provided
	if len(hourlyForecast) > 0 {
		forecastStart := time.Now()
		// Use minimum drying time of 4 hours for forecast to avoid excessive wet/dry transitions
		baseDryingHours := status.HoursUntilDry
		if baseDryingHours < 4.0 {
			baseDryingHours = 4.0
		}
		status.Forecast = c.Calculate6DayForecast(status, hourlyForecast, baseDryingHours)
		forecastTime := time.Since(forecastStart)
		if forecastTime > 50*time.Millisecond {
			log.Printf("[PERF]     6-day forecast calculation took %v (should be <50ms)", forecastTime)
		}
	}

	return status, nil
}

// calculateSunExposure computes hours of direct sun hitting the boulder over next 6 days
// Always calculates fresh since sun exposure is time-dependent (next 6 days from NOW)
// This is fast because it uses offline astronomical calculations (no API calls)
func (c *Calculator) calculateSunExposure(
	ctx context.Context,
	lat, lon float64,
	aspect string,
	treeCoverage float64,
	profile *models.BoulderDryingProfile,
) (float64, error) {
	// Calculate sun exposure using offline algorithm (fast - takes milliseconds)
	// Calculate for next 6 days (144 hours) starting from NOW
	startTime := time.Now()
	totalSunHours := sun.CalculateSunExposure(lat, lon, aspect, treeCoverage, startTime, 144)

	return totalSunHours, nil
}

// calculateBoulderDryingTime applies boulder-specific modifiers to location drying time
func (c *Calculator) calculateBoulderDryingTime(
	locationDrying *models.RockDryingStatus,
	boulderStatus *BoulderDryingStatus,
) float64 {
	baseDryingTime := locationDrying.HoursUntilDry

	// If location is already dry, boulder is also dry
	if baseDryingTime <= 0 {
		return 0
	}

	// Apply sun exposure modifier
	// More sun hours = faster drying
	sunModifier := 1.0
	avgSunPerDay := boulderStatus.SunExposureHours / 6.0
	if avgSunPerDay >= 8 {
		sunModifier = 0.7 // Exceptional sun exposure (-30% drying time)
	} else if avgSunPerDay >= 6 {
		sunModifier = 0.85 // Good sun exposure (-15% drying time)
	} else if avgSunPerDay >= 4 {
		sunModifier = 1.0 // Average sun exposure (no change)
	} else if avgSunPerDay >= 2 {
		sunModifier = 1.15 // Poor sun exposure (+15% drying time)
	} else {
		sunModifier = 1.3 // Minimal sun exposure (+30% drying time)
	}

	// Apply tree coverage penalty (in addition to sun exposure calculation)
	// Tree cover also affects airflow and humidity
	treeModifier := 1.0
	if boulderStatus.TreeCoveragePercent > 75 {
		treeModifier = 1.3 // Heavy tree cover (+30% drying time)
	} else if boulderStatus.TreeCoveragePercent > 50 {
		treeModifier = 1.15 // Moderate tree cover (+15% drying time)
	} else if boulderStatus.TreeCoveragePercent > 25 {
		treeModifier = 1.05 // Light tree cover (+5% drying time)
	}

	return baseDryingTime * sunModifier * treeModifier
}

// determineDryingStatus maps hours until dry to status level
func (c *Calculator) determineDryingStatus(status *BoulderDryingStatus) string {
	if !status.IsWet {
		return "good"
	}

	// Wet-sensitive rock gets "critical" status when wet
	if status.RockType == "sandstone" || status.RockType == "arkose" || status.RockType == "graywacke" {
		return "critical"
	}

	if status.HoursUntilDry >= 48 {
		return "poor"
	} else if status.HoursUntilDry >= 24 {
		return "fair"
	} else {
		return "fair"
	}
}

// generateStatusMessage creates human-readable status message
func (c *Calculator) generateStatusMessage(status *BoulderDryingStatus, locationDrying *models.RockDryingStatus) string {
	if !status.IsWet {
		return "Boulder is dry and ready to climb"
	}

	if status.Status == "critical" {
		return fmt.Sprintf("DO NOT CLIMB - %s is wet-sensitive and currently wet (%.0fh until dry)",
			status.RockType, status.HoursUntilDry)
	}

	return fmt.Sprintf("Boulder is wet (%.0fh until dry) - %s",
		status.HoursUntilDry, locationDrying.Message)
}

// estimateSunExposureFromAspect provides fallback sun exposure estimate based on aspect alone
// Used when GPS or sun position API is unavailable
func (c *Calculator) estimateSunExposureFromAspect(aspect string) float64 {
	// Rough estimate of daily sun hours based on aspect (winter PNW)
	// Multiply by 6 days for total forecast period
	dailySunHours := map[string]float64{
		"N":  2.0,  // Minimal sun
		"NE": 3.0,  // Morning sun only
		"E":  4.0,  // Morning sun
		"SE": 6.0,  // Morning + midday sun
		"S":  8.0,  // Maximum sun exposure
		"SW": 6.0,  // Afternoon sun
		"W":  4.0,  // Afternoon sun only
		"NW": 3.0,  // Late afternoon only
	}

	if hours, ok := dailySunHours[aspect]; ok {
		return hours * 6.0 // 6 days forecast
	}
	return 24.0 // Default fallback (4 hours/day)
}

// angleDifference calculates the minimum angular difference between two directions
// Returns value in range [0, 180]
func angleDifference(angle1, angle2 float64) float64 {
	diff := angle1 - angle2
	if diff < 0 {
		diff = -diff
	}
	if diff > 180 {
		diff = 360 - diff
	}
	return diff
}

// absFloat returns the absolute value of a float64
func absFloat(x float64) float64 {
	return math.Abs(x)
}

// Calculate6DayForecast generates a 6-day forecast showing when the boulder will be dry/wet
// based on future precipitation and drying times
func (c *Calculator) Calculate6DayForecast(
	status *BoulderDryingStatus,
	hourlyForecast []models.WeatherData,
	baseDryingHours float64, // Base drying time for location
) []DryingForecastPeriod {
	if len(hourlyForecast) == 0 {
		log.Printf("Warning: Calculate6DayForecast called with empty hourlyForecast")
		return nil
	}

	var forecast []DryingForecastPeriod
	now := time.Now()

	// Track current wet/dry state
	currentlyWet := status.IsWet
	wetSince := now
	if currentlyWet && status.LastRainTimestamp != nil && !status.LastRainTimestamp.IsZero() {
		wetSince = *status.LastRainTimestamp
	}

	// Add initial period
	if currentlyWet {
		// Start with wet period
		forecast = append(forecast, DryingForecastPeriod{
			StartTime:     now,
			IsDry:         false,
			Status:        "wet",
			RainAmount:    0.0, // Will accumulate as we process
			HoursUntilDry: baseDryingHours,
		})
	} else {
		// Start with dry period
		forecast = append(forecast, DryingForecastPeriod{
			StartTime: now,
			IsDry:     true,
			Status:    "dry",
		})
	}

	// Minimum rain threshold (inches) to consider it "wet"
	const rainThreshold = 0.01

	// Process hourly forecast
	for i, hour := range hourlyForecast {
		// Stop after 6 days (144 hours)
		if hour.Timestamp.Sub(now).Hours() > 144 {
			break
		}

		// Check if this hour has significant rain
		hasRain := hour.Precipitation >= rainThreshold

		// State transitions
		if !currentlyWet && hasRain {
			// Transition: dry -> wet
			// Close previous dry period
			if len(forecast) > 0 {
				forecast[len(forecast)-1].EndTime = hour.Timestamp
			}

			// Start new wet period
			currentlyWet = true
			wetSince = hour.Timestamp
			forecast = append(forecast, DryingForecastPeriod{
				StartTime:  hour.Timestamp,
				IsDry:      false,
				Status:     "wet",
				RainAmount: hour.Precipitation,
			})
		} else if currentlyWet {
			// Currently wet - check if more rain or drying
			if len(forecast) == 0 {
				// Safety check - forecast should never be empty here, but if it is, skip this iteration
				continue
			}

			if hasRain {
				// More rain - accumulate and reset drying clock
				forecast[len(forecast)-1].RainAmount += hour.Precipitation
				wetSince = hour.Timestamp // Reset drying clock
			}

			// Calculate hours since last rain
			hoursSinceRain := hour.Timestamp.Sub(wetSince).Hours()

			// Calculate drying time including extra time for heavy rain
			currentRain := forecast[len(forecast)-1].RainAmount
			dryingTime := baseDryingHours
			if currentRain > 0.5 {
				// Significant rain - add extra drying time: 12h per inch over 0.5"
				dryingTime += (currentRain - 0.5) * 12
			}

			if hoursSinceRain >= dryingTime {
				// Transition: wet -> dry
				dryTime := wetSince.Add(time.Duration(dryingTime) * time.Hour)

				// Only create transition if wet period had meaningful duration (at least 1 hour)
				wetDuration := dryTime.Sub(forecast[len(forecast)-1].StartTime).Hours()
				if wetDuration >= 1.0 {
					forecast[len(forecast)-1].EndTime = dryTime
					forecast[len(forecast)-1].HoursUntilDry = 0

					// Start new dry period
					currentlyWet = false
					forecast = append(forecast, DryingForecastPeriod{
						StartTime: dryTime,
						IsDry:     true,
						Status:    "dry",
					})
				} else {
					// Wet period too short - just mark it as dry and remove it
					currentlyWet = false
					if len(forecast) > 0 {
						forecast = forecast[:len(forecast)-1]
					}
				}
			} else {
				// Still wet - update hours until dry and status
				forecast[len(forecast)-1].HoursUntilDry = dryingTime - hoursSinceRain

				// Determine status based on progress
				progress := hoursSinceRain / dryingTime
				if progress > 0.5 {
					forecast[len(forecast)-1].Status = "drying"
				} else {
					forecast[len(forecast)-1].Status = "wet"
				}
			}
		}

		// If last hour, close final period
		if i == len(hourlyForecast)-1 {
			if forecast[len(forecast)-1].EndTime.IsZero() {
				forecast[len(forecast)-1].EndTime = hour.Timestamp.Add(6 * 24 * time.Hour)
			}
		}
	}

	return forecast
}
