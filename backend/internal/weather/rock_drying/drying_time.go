package rock_drying

import (
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// estimateDryingTime calculates hours needed for rock to dry
func estimateDryingTime(
	rockType models.RockType,
	currentWeather *models.WeatherData,
	historicalWeather []models.WeatherData,
	sunExposure *models.LocationSunExposure,
	hasSeepageRisk bool,
	rainAmount float64,
) float64 {
	// Start with base drying time (hours to dry after 0.1" rain in ideal conditions)
	baseDrying := rockType.BaseDryingHours

	// Adjust for rain amount (0.1" is baseline, scale proportionally)
	rainFactor := rainAmount / 0.1
	if rainFactor < 0.5 {
		rainFactor = 0.5 // Minimum for any measurable rain
	}
	if rainFactor > 3.0 {
		rainFactor = 3.0 // Cap at 3x for heavy rain
	}

	dryingTime := baseDrying * rainFactor

	// POROSITY FACTOR (previously missing!)
	// Higher porosity = slower drying (rock absorbs more water)
	// Baseline porosity: 5%
	// Scale: 1% porosity = 0.95x time, 20% porosity = 1.3x time
	porosityFactor := 1.0 + ((rockType.PorosityPercent - 5.0) / 100.0)
	if porosityFactor < 0.7 {
		porosityFactor = 0.7 // Min 70% time for non-porous rocks
	}
	if porosityFactor > 1.5 {
		porosityFactor = 1.5 // Max 150% time for porous rocks
	}
	dryingTime *= porosityFactor

	// Current weather modifiers (for baseline estimate)
	// Temperature modifier
	if currentWeather.Temperature > 70 {
		dryingTime *= 0.75 // -25% for hot weather
	} else if currentWeather.Temperature > 65 {
		dryingTime *= 0.85 // -15% for warm weather
	} else if currentWeather.Temperature < 50 {
		dryingTime *= 1.4 // +40% for cold weather
	} else if currentWeather.Temperature < 55 {
		dryingTime *= 1.2 // +20% for cool weather
	}

	// Cloud cover modifier (affects sun drying)
	if currentWeather.CloudCover < 30 {
		dryingTime *= 0.8 // -20% for sunny conditions
	} else if currentWeather.CloudCover < 50 {
		dryingTime *= 0.9 // -10% for partly cloudy
	}

	// Wind modifier (5-15 mph is ideal)
	if currentWeather.WindSpeed >= 5 && currentWeather.WindSpeed <= 15 {
		dryingTime *= 0.8 // -20% for good wind
	} else if currentWeather.WindSpeed < 3 {
		dryingTime *= 1.1 // +10% for calm conditions
	}

	// Humidity modifier
	if currentWeather.Humidity < 40 {
		dryingTime *= 0.75 // -25% for very low humidity
	} else if currentWeather.Humidity < 50 {
		dryingTime *= 0.85 // -15% for low humidity
	} else if currentWeather.Humidity > 80 {
		dryingTime *= 1.35 // +35% for very high humidity
	} else if currentWeather.Humidity > 70 {
		dryingTime *= 1.2 // +20% for high humidity
	}

	// Sun exposure modifier
	if sunExposure != nil {
		sunFactor := calculateSunExposureFactor(*sunExposure, *currentWeather)
		dryingTime /= sunFactor // Divide because higher sunFactor = faster drying
	}

	// Seepage risk modifier (groundwater, snowmelt, etc.)
	if hasSeepageRisk {
		dryingTime *= 1.4 // +40% for seepage areas
	}

	// Wet-sensitive rocks need extra drying time for safety
	if rockType.IsWetSensitive {
		dryingTime *= 1.5 // +50% safety margin for sandstone
	}

	return dryingTime
}

// calculateTimeWeightedDrying calculates how much drying has occurred since rain ended
// Returns a value from 0.0 (no drying) to 1.0 (complete drying)
func calculateTimeWeightedDrying(
	rainEndTime time.Time,
	currentWeather *models.WeatherData,
	historicalWeather []models.WeatherData,
	sunExposure *models.LocationSunExposure,
) float64 {
	totalDryingPower := 0.0

	// Analyze each hour since rain stopped
	for _, h := range historicalWeather {
		if h.Timestamp.After(rainEndTime) {
			// Calculate drying power for this hour
			hourlyPower := calculateHourlyDryingPower(h, sunExposure)
			totalDryingPower += hourlyPower
		}
	}

	// Include current conditions if applicable
	if currentWeather.Timestamp.After(rainEndTime) {
		hourlyPower := calculateHourlyDryingPower(*currentWeather, sunExposure)
		hoursSinceLast := time.Since(currentWeather.Timestamp).Hours()
		if hoursSinceLast < 1.0 {
			totalDryingPower += hourlyPower * hoursSinceLast
		}
	}

	// Normalize: totalDryingPower represents "effective drying hours"
	// We'll return this as a fraction (capped at 1.0)
	// The caller will use this to adjust required drying time
	hoursSinceRain := time.Since(rainEndTime).Hours()
	if hoursSinceRain <= 0 {
		return 0.0
	}

	// Progress = (effective drying hours) / (actual elapsed hours)
	// This gives us a multiplier for how fast drying is occurring
	progress := totalDryingPower / hoursSinceRain

	// Cap at 1.0 (fully dry)
	if progress > 1.0 {
		return 1.0
	}

	return progress
}

// calculateHourlyDryingPower calculates drying effectiveness for a single hour
// Returns a value where 1.0 = baseline drying, >1.0 = faster, <1.0 = slower
func calculateHourlyDryingPower(
	weather models.WeatherData,
	sunExposure *models.LocationSunExposure,
) float64 {
	power := 1.0

	// Temperature effect (warm = faster drying)
	if weather.Temperature > 70 {
		power *= 1.3 // +30% for hot weather
	} else if weather.Temperature > 65 {
		power *= 1.15 // +15% for warm weather
	} else if weather.Temperature < 50 {
		power *= 0.6 // -40% for cold weather
	} else if weather.Temperature < 55 {
		power *= 0.8 // -20% for cool weather
	}

	// Wind effect (5-15 mph is ideal)
	if weather.WindSpeed >= 5 && weather.WindSpeed <= 15 {
		power *= 1.25 // +25% for ideal wind
	} else if weather.WindSpeed > 15 && weather.WindSpeed <= 25 {
		power *= 1.1 // +10% for moderate wind
	} else if weather.WindSpeed < 3 {
		power *= 0.85 // -15% for calm conditions
	}

	// Humidity effect
	if weather.Humidity < 40 {
		power *= 1.3 // +30% for very dry air
	} else if weather.Humidity < 50 {
		power *= 1.15 // +15% for dry air
	} else if weather.Humidity > 80 {
		power *= 0.6 // -40% for very humid
	} else if weather.Humidity > 70 {
		power *= 0.75 // -25% for humid
	}

	// Sun exposure effect
	if sunExposure != nil {
		sunFactor := calculateSunExposureFactor(*sunExposure, weather)
		power *= sunFactor
	}

	// Cloud cover effect (affects sun drying)
	if weather.CloudCover < 30 {
		power *= 1.2 // +20% for sunny
	} else if weather.CloudCover < 50 {
		power *= 1.1 // +10% for partly cloudy
	} else if weather.CloudCover > 80 {
		power *= 0.85 // -15% for overcast
	}

	return power
}

// calculateSunExposureFactor calculates sun exposure multiplier based on location profile
func calculateSunExposureFactor(
	sunExposure models.LocationSunExposure,
	weather models.WeatherData,
) float64 {
	factor := 1.0

	// Aspect weighting (south is best, north is worst)
	// South: 1.3x, West: 1.15x, East: 1.05x, North: 0.85x
	aspectBonus := 0.0
	aspectBonus += (sunExposure.SouthFacingPercent / 100.0) * 0.3  // South: +30%
	aspectBonus += (sunExposure.WestFacingPercent / 100.0) * 0.15  // West: +15%
	aspectBonus += (sunExposure.EastFacingPercent / 100.0) * 0.05  // East: +5%
	aspectBonus += (sunExposure.NorthFacingPercent / 100.0) * -0.15 // North: -15%
	factor += aspectBonus

	// Rock angle weighting
	// Slabs: +20% (water runs off, more sun exposure)
	// Overhangs: -10% (less sun, stays wet longer)
	angleBonus := 0.0
	angleBonus += (sunExposure.SlabPercent / 100.0) * 0.2    // Slabs: +20%
	angleBonus += (sunExposure.OverhangPercent / 100.0) * -0.1 // Overhangs: -10%
	factor += angleBonus

	// Tree coverage penalty (shade reduces drying)
	// 0-25% trees: no penalty
	// 25-50% trees: -10%
	// 50-75% trees: -20%
	// 75-100% trees: -30%
	if sunExposure.TreeCoveragePercent > 75 {
		factor *= 0.7 // -30%
	} else if sunExposure.TreeCoveragePercent > 50 {
		factor *= 0.8 // -20%
	} else if sunExposure.TreeCoveragePercent > 25 {
		factor *= 0.9 // -10%
	}

	// Ensure factor stays within reasonable bounds
	if factor < 0.5 {
		factor = 0.5 // Minimum 50% drying rate
	}
	if factor > 1.5 {
		factor = 1.5 // Maximum 150% drying rate
	}

	return factor
}
