package calculator

import (
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// SnowAccumulationModel implements a SWE-based temperature-indexed snow model
// Tracks snow accumulation and melt using Snow Water Equivalent (SWE) and snow density
// - Snow accumulation with temperature-dependent density and freezing level transition
// - Rain-on-snow compaction and melt
// - Temperature-based melt above 34°F
// - Wind-enhanced melt and sublimation
// - Humidity-based sublimation
// - Elevation-adjusted temperatures
// - Natural compaction / settling

// SnowAccumulationModel tracks the state of snow accumulation over time
type SnowAccumulationModel struct {
	snowSWE     float64 // inches of water equivalent
	snowDensity float64 // fraction (0.08–0.35 typical)
}

// NewSnowAccumulationModel creates a new snow accumulation model
func NewSnowAccumulationModel() *SnowAccumulationModel {
	return &SnowAccumulationModel{
		snowSWE:     0.0,
		snowDensity: 0.12, // default density
	}
}

// CalculateSnowAccumulation calculates snow depth for each day given historical and forecast data
// Returns a map of date strings (YYYY-MM-DD) to snow depth in inches
func CalculateSnowAccumulation(historicalData, futureData []models.WeatherData, elevationFt float64) map[string]float64 {
	model := NewSnowAccumulationModel()
	snowDepthByDay := make(map[string]float64)

	// Load Pacific timezone once (for date keys)
	pacificTZ, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		// Fallback to UTC if timezone load fails
		pacificTZ = time.UTC
	}

	// Combine all data chronologically (make a copy to avoid modifying input slices)
	allData := make([]models.WeatherData, 0, len(historicalData)+len(futureData))
	allData = append(allData, historicalData...)
	allData = append(allData, futureData...)
	// Sort by timestamp (should already be sorted, but be safe)
	sortWeatherDataByTime(allData)

	for _, hour := range allData {
		// Use temperature as-is (Open-Meteo API already provides temps at location elevation)
		// Note: elevationFt parameter is kept for potential future use but not currently applied
		temp := hour.Temperature
		precip := hour.Precipitation // This is HOURLY precipitation (not 3-hour)
		windSpeed := hour.WindSpeed
		humidity := float64(hour.Humidity)

		// --- Freezing Level Transition (30-34°F mix zone) ---
		snowFraction := getSnowFraction(temp)

		if precip > 0 {
			if snowFraction > 0 {
				// Snowfall portion (precip is 3-hour total)
				snowPrecip := precip * snowFraction
				model.snowSWE += snowPrecip
				newSnowDensity := getNewSnowDensity(temp)
				model.snowDensity = blendDensity(model.snowDensity, model.snowSWE, snowPrecip, newSnowDensity)
			}

			if snowFraction < 1 && model.snowSWE > 0 {
				// Rain-on-snow portion
				rainPrecip := precip * (1 - snowFraction)
				model.snowSWE += rainPrecip * 0.7 // most rain infiltrates the pack
				model.snowDensity = min(0.35, model.snowDensity+0.03) // pack compacts

				// Rain energy melt (warmer rain melts more snow)
				rainTemp := max(temp, 32.0)
				rainEnergyMelt := rainPrecip * (rainTemp - 32) * 0.01
				model.snowSWE = max(0, model.snowSWE-rainEnergyMelt)
			}
		}

		// --- Temperature-based melt ---
		if temp > 34 && model.snowSWE > 0 {
			melt := calculateSWEMelt(temp) // Hourly period
			model.snowSWE = max(0, model.snowSWE-melt)
		} else if temp > 30 && temp <= 34 && model.snowSWE > 0 {
			// Very slow melt in transition zone (30-34°F) from solar radiation
			// Reduced significantly to prevent excessive melt at freezing temps
			baseMelt := (temp - 30) * 0.001
			model.snowSWE = max(0, model.snowSWE-baseMelt)
		}

		// --- Wind-enhanced melt and sublimation (increased for realism) ---
		if windSpeed > 10 && model.snowSWE > 0 {
			windMelt := (windSpeed - 10) * 0.002 // Tuned for PNW maritime conditions with strong winds
			model.snowSWE = max(0, model.snowSWE-windMelt)
		}

		// --- Humidity-based sublimation (increased for dry conditions) ---
		if humidity < 60 && model.snowSWE > 0 {
			sublimation := (60 - humidity) * 0.0005 // Increased from 0.0001 for faster sublimation
			model.snowSWE = max(0, model.snowSWE-sublimation)
		}

		// --- Compaction / settling ---
		if model.snowSWE > 0 {
			model.snowDensity = min(0.4, model.snowDensity+getCompactionRate(temp))
		}

		// --- Derive depth ---
		snowDepth := 0.0
		if model.snowSWE > 0 {
			snowDepth = model.snowSWE / model.snowDensity
		}

		// --- Store daily snow depth ---
		// Use Pacific timezone for date key to match frontend expectations
		// Store the MAXIMUM snow depth for each day (not the last value)
		// This prevents end-of-day melt from showing 0" when there was snow earlier
		dateKey := hour.Timestamp.In(pacificTZ).Format("2006-01-02")
		if existingDepth, exists := snowDepthByDay[dateKey]; !exists || snowDepth > existingDepth {
			snowDepthByDay[dateKey] = snowDepth
		}
	}

	return snowDepthByDay
}

// GetCurrentSnowDepth returns the current snow depth at the current moment
func GetCurrentSnowDepth(historicalData, currentData []models.WeatherData, elevationFt float64) float64 {
	// Just use CalculateSnowAccumulation and get today's value
	snowDepthByDay := CalculateSnowAccumulation(historicalData, currentData, elevationFt)

	if len(currentData) == 0 {
		return 0
	}

	// Use Pacific timezone for date key to match CalculateSnowAccumulation
	pacificTZ, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		pacificTZ = time.UTC
	}

	// Get today's date (use actual current time, not forecast time)
	nowPacific := time.Now().In(pacificTZ)
	todayKey := nowPacific.Format("2006-01-02")

	// Return snow depth for today
	if depth, exists := snowDepthByDay[todayKey]; exists {
		return depth
	}

	// Fallback: try yesterday (in case today hasn't been processed yet)
	yesterdayKey := nowPacific.AddDate(0, 0, -1).Format("2006-01-02")
	if depth, exists := snowDepthByDay[yesterdayKey]; exists {
		return depth
	}

	// Fallback: return the latest available date
	var latestDate string
	var latestDepth float64
	for date, depth := range snowDepthByDay {
		if date > latestDate {
			latestDate = date
			latestDepth = depth
		}
	}
	return latestDepth
}

// getSnowFraction calculates snow fraction based on temperature (freezing level transition)
// Returns 1.0 for all snow (temp <= 30°F)
// Returns 0.0 for all rain (temp >= 34°F)
// Returns gradient between 30-34°F
func getSnowFraction(temp float64) float64 {
	if temp <= 30 {
		return 1.0 // All snow
	}
	if temp >= 34 {
		return 0.0 // All rain
	}
	// Linear interpolation in transition zone
	return (34 - temp) / 4
}

// getNewSnowDensity estimates density of new snow based on temperature
func getNewSnowDensity(temp float64) float64 {
	if temp <= 20 {
		return 0.08 // very fluffy
	}
	if temp <= 28 {
		return 0.12 // cold
	}
	if temp <= 32 {
		return 0.18 // near freezing, wet
	}
	return 0.2 // default slightly wet
}

// blendDensity blends new snow density with existing pack
func blendDensity(currentDensity, currentSWE, newSWE, newDensity float64) float64 {
	if currentSWE == 0 {
		return newDensity
	}
	return (currentDensity*(currentSWE-newSWE) + newDensity*newSWE) / currentSWE
}

// calculateSWEMelt calculates temperature-driven SWE melt (PNW degree-day approximation)
// Returns inches SWE per hour
// Pacific Northwest maritime climate has moderate melt rates
// Uses non-linear melt that accelerates at warmer temperatures
func calculateSWEMelt(temp float64) float64 {
	if temp <= 34 {
		return 0
	}

	degreesDiff := temp - 34

	// Non-linear melt rate: moderate near freezing, rapid at warm temps
	// Base: 0.02 per degree, plus quadratic term for acceleration
	// At 36°F (2° above): 0.02 * 2 + 0.0003 * 4 = 0.041 SWE/hour
	// At 40°F (6° above): 0.02 * 6 + 0.0003 * 36 = 0.131 SWE/hour
	// At 50°F (16° above): 0.02 * 16 + 0.0003 * 256 = 0.397 SWE/hour
	baseMelt := 0.02*degreesDiff + 0.0003*degreesDiff*degreesDiff

	return baseMelt
}

// getCompactionRate returns compaction rate based on temperature
func getCompactionRate(temp float64) float64 {
	if temp < 20 {
		return 0.0003
	}
	if temp < 28 {
		return 0.0006
	}
	if temp < 32 {
		return 0.0012
	}
	return 0.0025 // warm, wet snow compacts fast
}

// sortWeatherDataByTime sorts weather data by timestamp (in place)
func sortWeatherDataByTime(data []models.WeatherData) {
	// Simple bubble sort (fine for small datasets)
	n := len(data)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if data[j].Timestamp.After(data[j+1].Timestamp) {
				data[j], data[j+1] = data[j+1], data[j]
			}
		}
	}
}

// min returns the minimum of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two float64 values
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
