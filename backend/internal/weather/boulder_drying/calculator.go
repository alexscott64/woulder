package boulder_drying

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// BoulderDryingStatus represents the drying status for a specific boulder
type BoulderDryingStatus struct {
	MPRouteID           string    `json:"mp_route_id"`
	IsWet               bool      `json:"is_wet"`
	IsSafe              bool      `json:"is_safe"`
	HoursUntilDry       float64   `json:"hours_until_dry"`
	Status              string    `json:"status"` // "critical", "poor", "fair", "good"
	Message             string    `json:"message"`
	ConfidenceScore     int       `json:"confidence_score"` // 0-100
	LastRainTimestamp   time.Time `json:"last_rain_timestamp"`
	SunExposureHours    float64   `json:"sun_exposure_hours"`     // Hours of direct sun next 6 days
	TreeCoveragePercent float64   `json:"tree_coverage_percent"` // 0-100
	RockType            string    `json:"rock_type"`
	Aspect              string    `json:"aspect"`      // N, NE, E, SE, S, SW, W, NW
	Latitude            float64   `json:"latitude"`    // Boulder GPS
	Longitude           float64   `json:"longitude"`   // Boulder GPS
}

// Calculator computes boulder-specific drying times
type Calculator struct {
	sunClient  *SunPositionClient
	treeClient *TreeCoverClient
}

// NewCalculator creates a new boulder drying calculator
func NewCalculator(ipGeoAPIKey string) *Calculator {
	return &Calculator{
		sunClient:  NewSunPositionClient(ipGeoAPIKey),
		treeClient: NewTreeCoverClient(),
	}
}

// CalculateBoulderDryingStatus computes the drying status for a specific boulder
// Extends location-level rock drying with boulder-specific factors:
// - Real-time sun exposure based on boulder GPS + aspect
// - Boulder-specific tree coverage from satellite data
// - Confidence scoring based on data availability
func (c *Calculator) CalculateBoulderDryingStatus(
	ctx context.Context,
	route *models.MPRoute,
	locationDrying *models.RockDryingStatus,
	profile *models.BoulderDryingProfile,
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

	// Get tree coverage (from profile cache or fetch new)
	if profile != nil && profile.TreeCoveragePercent != nil {
		status.TreeCoveragePercent = *profile.TreeCoveragePercent
	} else if status.Latitude != 0 && status.Longitude != 0 {
		// Fetch tree coverage from satellite data
		treeCoverage, err := c.treeClient.GetTreeCoverage(ctx, status.Latitude, status.Longitude)
		if err != nil {
			log.Printf("Warning: Failed to fetch tree coverage for %s: %v", route.MPRouteID, err)
			// Fall back to location-level tree coverage (if available)
			status.TreeCoveragePercent = 0 // Default to no tree coverage
			status.ConfidenceScore -= 15
		} else {
			status.TreeCoveragePercent = treeCoverage
		}
	} else {
		status.TreeCoveragePercent = 0
		status.ConfidenceScore -= 15
	}

	// Get sun exposure hours (cached or calculate)
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

	// Calculate boulder-specific drying time
	status.HoursUntilDry = c.calculateBoulderDryingTime(locationDrying, status)

	// Parse LastRainTimestamp string to time.Time
	if locationDrying.LastRainTimestamp != "" {
		lastRain, err := time.Parse(time.RFC3339, locationDrying.LastRainTimestamp)
		if err == nil {
			status.LastRainTimestamp = lastRain
		}
	}

	// Determine wet/safe/status
	status.IsWet = status.HoursUntilDry > 0
	status.IsSafe = !status.IsWet || !locationDrying.IsWetSensitive
	status.Status = c.determineDryingStatus(status)
	status.Message = c.generateStatusMessage(status, locationDrying)

	return status, nil
}

// SunExposureCache stores cached sun exposure calculation
type SunExposureCache struct {
	Aspect              string    `json:"aspect"`
	TreeCoverage        float64   `json:"tree_coverage"`
	SunExposureHours    float64   `json:"sun_exposure_hours"`
	CalculatedAt        time.Time `json:"calculated_at"`
	ForecastStartDate   string    `json:"forecast_start_date"`
}

// calculateSunExposure computes hours of direct sun hitting the boulder over next 6 days
// Uses cached sun position data if available, otherwise fetches from API
func (c *Calculator) calculateSunExposure(
	ctx context.Context,
	lat, lon float64,
	aspect string,
	treeCoverage float64,
	profile *models.BoulderDryingProfile,
) (float64, error) {
	// Check cache first (6-hour TTL)
	if profile != nil && profile.LastSunCalcAt != nil && profile.SunExposureHoursCache != nil {
		cacheAge := time.Since(*profile.LastSunCalcAt)
		if cacheAge < 6*time.Hour {
			// Parse cached sun exposure from JSONB (stored as string)
			var cache SunExposureCache
			if err := json.Unmarshal([]byte(*profile.SunExposureHoursCache), &cache); err == nil {
				// Verify cache is still valid (same aspect and similar tree coverage)
				if cache.Aspect == aspect && absFloat(cache.TreeCoverage-treeCoverage) < 5.0 {
					log.Printf("Using cached sun exposure for boulder (age: %v): %.1f hours",
						cacheAge.Round(time.Minute), cache.SunExposureHours)
					return cache.SunExposureHours, nil
				}
			}
		}
	}

	// Fetch sun position data for next 6 days
	sunData, err := c.sunClient.GetSunPositionForecast(ctx, lat, lon, 6)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch sun position data: %w", err)
	}

	// Calculate total sun exposure hours
	totalSunHours := 0.0
	aspectDegrees := AspectToDegrees(aspect)

	for _, hourData := range sunData {
		// Only count hours when sun is above horizon
		if hourData.Elevation <= 0 {
			continue
		}

		// Check if sun hits boulder face based on aspect
		// Boulder receives sun when sun azimuth is within ±90° of aspect direction
		azimuthDiff := angleDifference(hourData.Azimuth, aspectDegrees)
		if azimuthDiff <= 90 {
			// Apply tree coverage reduction
			sunFactor := 1.0
			if treeCoverage > 75 {
				sunFactor = 0.3 // Heavy tree cover blocks 70% of sun
			} else if treeCoverage > 50 {
				sunFactor = 0.6 // Moderate tree cover blocks 40% of sun
			} else if treeCoverage > 25 {
				sunFactor = 0.8 // Light tree cover blocks 20% of sun
			}

			totalSunHours += sunFactor
		}
	}

	log.Printf("Calculated sun exposure for boulder: %.1f hours over 6 days (aspect: %s, tree: %.0f%%)",
		totalSunHours, aspect, treeCoverage)

	return totalSunHours, nil
}

// GetSunExposureCacheData returns the cache data to be saved (used by service layer)
func (c *Calculator) GetSunExposureCacheData(aspect string, treeCoverage, sunExposureHours float64) ([]byte, error) {
	cache := SunExposureCache{
		Aspect:            aspect,
		TreeCoverage:      treeCoverage,
		SunExposureHours:  sunExposureHours,
		CalculatedAt:      time.Now(),
		ForecastStartDate: time.Now().Format("2006-01-02"),
	}
	return json.Marshal(cache)
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
