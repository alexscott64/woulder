package rock_drying

import (
	"math"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// estimateSnowMeltTime estimates hours until snow melts off rock
// Based on temperature, sun exposure, and rock thermal properties
func estimateSnowMeltTime(
	snowDepthInches float64,
	currentWeather *models.WeatherData,
	historicalWeather []models.WeatherData,
	rockType models.RockType,
	sunExposure *models.LocationSunExposure,
) float64 {
	// Base melt rate: inches per hour at different temps (empirical values)
	// Snow melts at ~0.1-0.2 inches per hour at 40°F
	// Faster at higher temps, slower at lower temps

	temp := currentWeather.Temperature

	// Check if it will warm up above freezing soon by looking at historical/forecast trends
	// Look at recent temps to see if there's a warming trend
	avgRecentTemp := temp
	if len(historicalWeather) > 0 {
		recentCount := int(math.Min(12, float64(len(historicalWeather)))) // Look at last 12 hours
		tempSum := 0.0
		for i := 0; i < recentCount; i++ {
			tempSum += historicalWeather[i].Temperature
		}
		avgRecentTemp = tempSum / float64(recentCount)
	}

	// If current temp is at/below freezing AND no warming trend, return high estimate
	// But not 999 - give a reasonable estimate based on seasonal expectations
	if temp <= 32 {
		// If warming trend exists (avg recent temps are warmer), use avg temp for estimate
		if avgRecentTemp > 34 {
			// Use average temp which is above freezing
			temp = avgRecentTemp
		} else {
			// No warming trend - estimate based on typical spring/winter patterns
			// In winter (Dec-Feb), snow can persist for weeks
			// In spring/fall (Mar-May, Sep-Nov), typically melts within a few days
			// In summer (Jun-Aug), unlikely to have snow, but if we do, melts fast
			currentMonth := currentWeather.Timestamp.Month()

			if currentMonth >= 6 && currentMonth <= 8 {
				// Summer - assume will melt within 2-3 days even if cold now
				return 48.0 + (snowDepthInches * 12.0) // 48h base + 12h per inch
			} else if currentMonth == 3 || currentMonth == 4 || currentMonth == 5 || currentMonth == 9 || currentMonth == 10 {
				// Spring/Fall - typically melts within a week
				return 96.0 + (snowDepthInches * 24.0) // 4 days base + 24h per inch
			} else {
				// Winter - can persist for 1-2 weeks
				return 168.0 + (snowDepthInches * 36.0) // 1 week base + 36h per inch
			}
		}
	}

	// Calculate base melt rate (inches per hour)
	// Formula: exponential increase with temperature above freezing
	// At 35°F: ~0.05 in/hr, at 40°F: ~0.15 in/hr, at 50°F: ~0.4 in/hr, at 60°F: ~0.8 in/hr
	tempAboveFreezing := temp - 32.0
	baseMeltRate := 0.02 * math.Pow(1.12, tempAboveFreezing)

	// Sun exposure multiplier (dark rocks absorb more heat, accelerating melt)
	// South-facing and sun-exposed rocks melt snow faster
	sunMultiplier := 1.0
	if sunExposure != nil {
		// South-facing rocks get most sun in winter (Northern Hemisphere)
		southFactor := 1.0 + (sunExposure.SouthFacingPercent / 100.0) * 0.5 // Up to +50%
		// West-facing gets afternoon sun
		westFactor := 1.0 + (sunExposure.WestFacingPercent / 100.0) * 0.3 // Up to +30%
		// Tree coverage blocks sun
		treeFactor := 1.0 - (sunExposure.TreeCoveragePercent / 100.0) * 0.4 // Up to -40%

		sunMultiplier = (southFactor + westFactor) / 2.0 * treeFactor
	}

	// Rock thermal properties (darker/denser rocks absorb more heat)
	// Granite (dark, dense) melts snow faster than limestone (light, porous)
	// Use porosity as proxy: higher porosity = lighter color = slower melt
	rockMultiplier := 1.0
	if rockType.PorosityPercent > 0 {
		// Lower porosity (denser, darker rock) = faster melt
		// Granite (1% porosity): 1.2x, Sandstone (15% porosity): 0.85x
		rockMultiplier = 1.3 - (rockType.PorosityPercent / 100.0)
		if rockMultiplier < 0.7 {
			rockMultiplier = 0.7 // Cap minimum at 0.7
		}
	}

	// Wind factor (wind accelerates sublimation and melt)
	windMultiplier := 1.0
	if currentWeather.WindSpeed > 5 {
		// Moderate wind (5-15 mph) helps: +20%
		// Strong wind (15+ mph) helps more: +40%
		if currentWeather.WindSpeed < 15 {
			windMultiplier = 1.2
		} else {
			windMultiplier = 1.4
		}
	}

	// Calculate effective melt rate
	effectiveMeltRate := baseMeltRate * sunMultiplier * rockMultiplier * windMultiplier

	// Ensure minimum melt rate to avoid division by zero or unrealistic estimates
	if effectiveMeltRate < 0.01 {
		effectiveMeltRate = 0.01
	}

	// Calculate hours to melt
	hoursToMelt := snowDepthInches / effectiveMeltRate

	// Account for forecast: look at upcoming temps to refine estimate
	// If temps are expected to warm, reduce estimate; if cooling, increase
	avgForecastTemp := currentWeather.Temperature
	if len(historicalWeather) > 0 {
		// Use last few hours as proxy for near-term forecast trend
		recentCount := int(math.Min(6, float64(len(historicalWeather))))
		tempSum := 0.0
		for i := 0; i < recentCount; i++ {
			tempSum += historicalWeather[i].Temperature
		}
		avgForecastTemp = tempSum / float64(recentCount)
	}

	// Adjust based on temperature trend
	if avgForecastTemp > temp {
		// Warming trend: faster melt
		hoursToMelt *= 0.85
	} else if avgForecastTemp < temp-5 {
		// Cooling trend: slower melt
		hoursToMelt *= 1.3
	}

	// Cap at reasonable maximum (2 weeks = 336 hours)
	if hoursToMelt > 336 {
		hoursToMelt = 336
	}

	// Add base time for rock to dry after snow melts
	// Use rock's base drying time as proxy (rock needs to dry after snow melts)
	rockDryTime := rockType.BaseDryingHours * 0.5 // Snow melt water dries faster than rain

	return hoursToMelt + rockDryTime
}
