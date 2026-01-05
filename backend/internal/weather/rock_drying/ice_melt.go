package rock_drying

import (
	"math"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// estimateIceMeltTime estimates hours until ice melts from rock
// Ice from frozen precipitation is thinner than snow, melts faster
func estimateIceMeltTime(
	recentPrecipInches float64,
	currentWeather *models.WeatherData,
	historicalWeather []models.WeatherData,
	rockType models.RockType,
	sunExposure *models.LocationSunExposure,
) float64 {
	temp := currentWeather.Temperature

	// Check for warming trend (same logic as snow melt)
	avgRecentTemp := temp
	if len(historicalWeather) > 0 {
		recentCount := int(math.Min(12, float64(len(historicalWeather))))
		tempSum := 0.0
		for i := 0; i < recentCount; i++ {
			tempSum += historicalWeather[i].Temperature
		}
		avgRecentTemp = tempSum / float64(recentCount)
	}

	// If at or below freezing AND no warming trend, return reasonable estimate
	if temp <= 32 {
		// If warming trend exists, use avg temp
		if avgRecentTemp > 34 {
			temp = avgRecentTemp
		} else {
			// Ice melts faster than snow, so use shorter estimates
			currentMonth := currentWeather.Timestamp.Month()

			// Estimate based on ice thickness (precip amount)
			iceThicknessInches := recentPrecipInches * 10.0

			if currentMonth >= 6 && currentMonth <= 8 {
				// Summer - ice melts within 1-2 days
				return 24.0 + (iceThicknessInches * 8.0)
			} else if currentMonth == 3 || currentMonth == 4 || currentMonth == 5 || currentMonth == 9 || currentMonth == 10 {
				// Spring/Fall - ice melts within 2-4 days
				return 48.0 + (iceThicknessInches * 12.0)
			} else {
				// Winter - ice can persist for a week
				return 84.0 + (iceThicknessInches * 18.0)
			}
		}
	}

	// Ice thickness estimate (conservative): recent precip forms thin ice layer
	// Assume 10:1 ratio (0.1" rain = ~1" of ice coating)
	iceThicknessInches := recentPrecipInches * 10.0

	// Ice melts faster than snow at same temp (denser, better heat conduction)
	// Base melt rate is ~50% faster than snow
	tempAboveFreezing := temp - 32.0
	baseMeltRate := 0.03 * math.Pow(1.15, tempAboveFreezing) // Faster than snow

	// Apply similar modifiers as snow melt
	sunMultiplier := 1.0
	if sunExposure != nil {
		southFactor := 1.0 + (sunExposure.SouthFacingPercent / 100.0) * 0.6 // Ice benefits more from sun
		westFactor := 1.0 + (sunExposure.WestFacingPercent / 100.0) * 0.4
		treeFactor := 1.0 - (sunExposure.TreeCoveragePercent / 100.0) * 0.5

		sunMultiplier = (southFactor + westFactor) / 2.0 * treeFactor
	}

	// Rock thermal properties (same logic as snow)
	rockMultiplier := 1.0
	if rockType.PorosityPercent > 0 {
		rockMultiplier = 1.3 - (rockType.PorosityPercent / 100.0)
		if rockMultiplier < 0.7 {
			rockMultiplier = 0.7
		}
	}

	// Wind helps sublimate ice
	windMultiplier := 1.0
	if currentWeather.WindSpeed > 5 {
		if currentWeather.WindSpeed < 15 {
			windMultiplier = 1.3 // Wind more effective on ice
		} else {
			windMultiplier = 1.6
		}
	}

	effectiveMeltRate := baseMeltRate * sunMultiplier * rockMultiplier * windMultiplier

	if effectiveMeltRate < 0.01 {
		effectiveMeltRate = 0.01
	}

	hoursToMelt := iceThicknessInches / effectiveMeltRate

	// Adjust for temperature trend
	avgForecastTemp := currentWeather.Temperature
	if len(historicalWeather) > 0 {
		recentCount := int(math.Min(6, float64(len(historicalWeather))))
		tempSum := 0.0
		for i := 0; i < recentCount; i++ {
			tempSum += historicalWeather[i].Temperature
		}
		avgForecastTemp = tempSum / float64(recentCount)
	}

	if avgForecastTemp > temp {
		hoursToMelt *= 0.8 // Warming = faster melt
	} else if avgForecastTemp < temp-5 {
		hoursToMelt *= 1.4 // Cooling = slower melt
	}

	// Cap at reasonable maximum
	if hoursToMelt > 168 { // 1 week max for ice
		hoursToMelt = 168
	}

	// Add rock drying time after ice melts
	rockDryTime := rockType.BaseDryingHours * 0.4 // Ice melt water dries quickly

	return hoursToMelt + rockDryTime
}
