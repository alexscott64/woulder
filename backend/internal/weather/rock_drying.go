package weather

import (
	"math"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// RockDryingCalculator calculates rock drying status based on weather and rock type
type RockDryingCalculator struct{}

// CalculateDryingStatus determines if rock is dry and safe to climb
func (c *RockDryingCalculator) CalculateDryingStatus(
	rockTypes []models.RockType,
	currentWeather *models.WeatherData,
	historicalWeather []models.WeatherData,
	sunExposure *models.LocationSunExposure,
	hasSeepageRisk bool,
	snowDepthInches *float64,
) models.RockDryingStatus {
	if len(rockTypes) == 0 {
		return models.RockDryingStatus{
			IsWet:           false,
			IsSafe:          true,
			Status:          "good",
			Message:         "No rock type data available",
			RockTypes:       []string{},
			ConfidenceScore: 30, // Low confidence due to missing data
		}
	}

	// Determine primary rock type (first in list, should be marked as primary in DB)
	primaryRock := rockTypes[0]
	rockTypeNames := make([]string, len(rockTypes))
	hasWetSensitive := false

	for i, rt := range rockTypes {
		rockTypeNames[i] = rt.Name
		if rt.IsWetSensitive {
			hasWetSensitive = true
		}
	}

	// Check for snow on ground - this is CRITICAL only for wet-sensitive rocks
	if snowDepthInches != nil && *snowDepthInches > 0.5 {
		// Only mark as "critical" for wet-sensitive rocks (sandstone, arkose, graywacke)
		status := "poor"
		message := "Snow on ground - rock may be wet"

		if hasWetSensitive {
			status = "critical"
			message = "DO NOT CLIMB - " + primaryRock.GroupName + " is wet-sensitive and there is snow on ground"
		} else if *snowDepthInches > 2.0 {
			message = "Significant snow accumulation on ground"
		}

		// Estimate snow melt time based on conditions
		estimatedDryTime := c.estimateSnowMeltTime(
			*snowDepthInches,
			currentWeather,
			historicalWeather,
			primaryRock,
			sunExposure,
		)

		return models.RockDryingStatus{
			IsWet:             true,
			IsSafe:            false,
			IsWetSensitive:    hasWetSensitive,
			HoursUntilDry:     estimatedDryTime,
			LastRainTimestamp: time.Now().Format(time.RFC3339),
			Status:            status,
			Message:           message,
			RockTypes:         rockTypeNames,
			PrimaryRockType:   primaryRock.Name,
			PrimaryGroupName:  primaryRock.GroupName,
			ConfidenceScore:   95, // High confidence - snow is visible
		}
	}

	// Check for freezing temperatures - water frozen on rock
	if currentWeather.Temperature <= 32 {
		// Check if there was recent precipitation that's now frozen
		recentPrecip := 0.0
		cutoffTime := time.Now().Add(-48 * time.Hour)

		for _, h := range historicalWeather {
			if h.Timestamp.After(cutoffTime) && h.Precipitation > 0.01 {
				recentPrecip += h.Precipitation
			}
		}

		if recentPrecip > 0.1 {
			// Only mark as "critical" for wet-sensitive rocks (sandstone, arkose, graywacke)
			status := "poor"
			message := "Freezing temps with recent precipitation - ice on rock"

			if hasWetSensitive {
				status = "critical"
				message = "DO NOT CLIMB - " + primaryRock.GroupName + " is wet-sensitive and may have ice"
			}

			// Estimate ice melt time based on conditions
			estimatedDryTime := c.estimateIceMeltTime(
				recentPrecip,
				currentWeather,
				historicalWeather,
				primaryRock,
				sunExposure,
			)

			return models.RockDryingStatus{
				IsWet:             true,
				IsSafe:            false,
				IsWetSensitive:    hasWetSensitive,
				HoursUntilDry:     estimatedDryTime,
				LastRainTimestamp: time.Now().Format(time.RFC3339),
				Status:            status,
				Message:           message,
				RockTypes:         rockTypeNames,
				PrimaryRockType:   primaryRock.Name,
				PrimaryGroupName:  primaryRock.GroupName,
				ConfidenceScore:   90,
			}
		}
	}

	// If currently raining
	if currentWeather.Precipitation > 0.01 {
		status := "poor"
		message := "Currently raining - rock is wet"

		if hasWetSensitive {
			status = "critical"
			message = "DO NOT CLIMB - " + primaryRock.GroupName + " is wet-sensitive and currently raining"
		}

		// Estimate drying time from current rain
		dryingTime := c.estimateDryingTime(primaryRock, currentWeather, historicalWeather,
			sunExposure, hasSeepageRisk, currentWeather.Precipitation)

		return models.RockDryingStatus{
			IsWet:             true,
			IsSafe:            false,
			IsWetSensitive:    hasWetSensitive,
			HoursUntilDry:     dryingTime,
			LastRainTimestamp: time.Now().Format(time.RFC3339),
			Status:            status,
			Message:           message,
			RockTypes:         rockTypeNames,
			PrimaryRockType:   primaryRock.Name,
			PrimaryGroupName:  primaryRock.GroupName,
			ConfidenceScore:   c.calculateConfidence(false, time.Now(), historicalWeather, hasWetSensitive, sunExposure),
		}
	}

	// Find last rain event
	lastRainEvent := c.findLastRainEvent(historicalWeather, currentWeather)

	// If no recent rain
	if lastRainEvent == nil {
		return models.RockDryingStatus{
			IsWet:            false,
			IsSafe:           true,
			IsWetSensitive:   hasWetSensitive,
			HoursUntilDry:    0,
			Status:           "good",
			Message:          "Rock is dry - no recent rain",
			RockTypes:        rockTypeNames,
			PrimaryRockType:  primaryRock.Name,
			PrimaryGroupName: primaryRock.GroupName,
			ConfidenceScore:  c.calculateConfidence(true, time.Time{}, historicalWeather, hasWetSensitive, sunExposure),
		}
	}

	// Calculate time-weighted drying progress
	requiredDryingTime := c.estimateDryingTime(primaryRock, currentWeather, historicalWeather,
		sunExposure, hasSeepageRisk, lastRainEvent.TotalRain)

	// Calculate how much drying has occurred (time-weighted)
	dryingProgress := c.calculateTimeWeightedDrying(lastRainEvent.EndTime, currentWeather, historicalWeather, sunExposure)

	// Adjust required time based on actual drying progress
	effectiveDryingTime := requiredDryingTime * (1.0 - dryingProgress)

	// Check if rock has dried
	if effectiveDryingTime <= 0 {
		return models.RockDryingStatus{
			IsWet:             false,
			IsSafe:            true,
			IsWetSensitive:    hasWetSensitive,
			HoursUntilDry:     0,
			LastRainTimestamp: lastRainEvent.EndTime.Format(time.RFC3339),
			Status:            "good",
			Message:           "Rock is dry and safe to climb",
			RockTypes:         rockTypeNames,
			PrimaryRockType:   primaryRock.Name,
			PrimaryGroupName:  primaryRock.GroupName,
			ConfidenceScore:   c.calculateConfidence(true, lastRainEvent.EndTime, historicalWeather, hasWetSensitive, sunExposure),
		}
	}

	// Rock is still drying
	status := "fair"
	message := "Rock is drying"

	if effectiveDryingTime > requiredDryingTime*0.5 {
		// More than 50% of drying time remaining
		status = "poor"
		message = "Rock is still wet"
	}

	// Wet-sensitive rocks (sandstone, arkose, graywacke) are ALWAYS critical when wet
	if hasWetSensitive {
		status = "critical"
		message = "DO NOT CLIMB - " + primaryRock.GroupName + " is wet-sensitive and still wet"
	}

	return models.RockDryingStatus{
		IsWet:             true,
		IsSafe:            status != "critical",
		IsWetSensitive:    hasWetSensitive,
		HoursUntilDry:     math.Ceil(effectiveDryingTime),
		LastRainTimestamp: lastRainEvent.EndTime.Format(time.RFC3339),
		Status:            status,
		Message:           message,
		RockTypes:         rockTypeNames,
		PrimaryRockType:   primaryRock.Name,
		PrimaryGroupName:  primaryRock.GroupName,
		ConfidenceScore:   c.calculateConfidence(false, lastRainEvent.EndTime, historicalWeather, hasWetSensitive, sunExposure),
	}
}

// findLastRainEvent finds the most recent rain event and calculates its characteristics
func (c *RockDryingCalculator) findLastRainEvent(historical []models.WeatherData, current *models.WeatherData) *models.RainEvent {
	if len(historical) == 0 {
		return nil
	}

	var event *models.RainEvent
	var totalRain float64
	var maxRate float64
	var rainHours int
	var startTime, endTime time.Time

	// Check historical data (reversed to go from most recent to oldest)
	for i := len(historical) - 1; i >= 0; i-- {
		h := historical[i]
		if h.Precipitation > 0.01 {
			// Part of rain event
			if event == nil {
				// Start of new rain event
				event = &models.RainEvent{
					StartTime: h.Timestamp,
					EndTime:   h.Timestamp,
				}
				endTime = h.Timestamp
			}
			startTime = h.Timestamp
			totalRain += h.Precipitation
			rainHours++

			// Track max hourly rate
			if h.Precipitation > maxRate {
				maxRate = h.Precipitation
			}
		} else if event != nil {
			// End of rain event (found dry period)
			break
		}
	}

	if event == nil {
		return nil
	}

	// Finalize rain event
	event.StartTime = startTime
	event.EndTime = endTime
	event.TotalRain = totalRain
	event.Duration = endTime.Sub(startTime).Hours()
	event.MaxHourlyRate = maxRate

	if rainHours > 0 {
		event.AvgHourlyRate = totalRain / float64(rainHours)
	}

	return event
}

// calculateTimeWeightedDrying calculates how much drying has occurred since rain ended
// Returns a value from 0.0 (no drying) to 1.0 (complete drying)
func (c *RockDryingCalculator) calculateTimeWeightedDrying(
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
			hourlyPower := c.calculateHourlyDryingPower(h, sunExposure)
			totalDryingPower += hourlyPower
		}
	}

	// Include current conditions if applicable
	if currentWeather.Timestamp.After(rainEndTime) {
		hourlyPower := c.calculateHourlyDryingPower(*currentWeather, sunExposure)
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
func (c *RockDryingCalculator) calculateHourlyDryingPower(
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
		sunFactor := c.calculateSunExposureFactor(*sunExposure, weather)
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
func (c *RockDryingCalculator) calculateSunExposureFactor(
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

// estimateDryingTime calculates hours needed for rock to dry
func (c *RockDryingCalculator) estimateDryingTime(
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
		sunFactor := c.calculateSunExposureFactor(*sunExposure, *currentWeather)
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

// calculateConfidence calculates confidence score (0-100) for the prediction
func (c *RockDryingCalculator) calculateConfidence(
	isDry bool,
	lastRainTime time.Time,
	historicalWeather []models.WeatherData,
	isWetSensitive bool,
	sunExposure *models.LocationSunExposure,
) int {
	confidence := 75.0 // Start with baseline

	// Data completeness factor
	if len(historicalWeather) < 24 {
		confidence -= 15 // -15 for insufficient historical data
	} else if len(historicalWeather) < 48 {
		confidence -= 8 // -8 for limited historical data
	}

	// Sun exposure profile availability
	if sunExposure == nil {
		confidence -= 10 // -10 for missing sun exposure data
	}

	// Time since rain factor (affects uncertainty)
	if !isDry && !lastRainTime.IsZero() {
		hoursSinceRain := time.Since(lastRainTime).Hours()

		if hoursSinceRain < 6 {
			confidence -= 5 // Recent rain = slightly more uncertain
		} else if hoursSinceRain > 72 {
			confidence += 8 // Long time = more confident in dryness
		}
	}

	// Weather stability (check for variable conditions)
	if len(historicalWeather) >= 12 {
		tempVariance := c.calculateTemperatureVariance(historicalWeather)
		if tempVariance > 15 {
			confidence -= 8 // High variance = less confident
		} else if tempVariance > 10 {
			confidence -= 4
		}
	}

	// Wet-sensitive rock factor (err on side of caution)
	if isWetSensitive && !isDry {
		confidence -= 5 // Lower confidence when wet-sensitive rock is wet
	}

	// Dry conditions boost
	if isDry && (lastRainTime.IsZero() || time.Since(lastRainTime).Hours() > 96) {
		confidence += 10 // High confidence in long-term dry conditions
	}

	// Clamp to realistic range (20-95)
	if confidence < 20 {
		confidence = 20
	}
	if confidence > 95 {
		confidence = 95
	}

	return int(confidence)
}

// calculateTemperatureVariance calculates temperature variance in recent weather
func (c *RockDryingCalculator) calculateTemperatureVariance(weather []models.WeatherData) float64 {
	if len(weather) == 0 {
		return 0
	}

	// Calculate mean
	sum := 0.0
	for _, w := range weather {
		sum += w.Temperature
	}
	mean := sum / float64(len(weather))

	// Calculate variance
	varianceSum := 0.0
	for _, w := range weather {
		diff := w.Temperature - mean
		varianceSum += diff * diff
	}
	variance := varianceSum / float64(len(weather))

	return math.Sqrt(variance) // Return standard deviation
}

// estimateSnowMeltTime estimates hours until snow melts off rock
// Based on temperature, sun exposure, and rock thermal properties
func (c *RockDryingCalculator) estimateSnowMeltTime(
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

	// If below freezing, snow won't melt - return very high estimate
	if temp <= 32 {
		return 999.0
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

// estimateIceMeltTime estimates hours until ice melts from rock
// Ice from frozen precipitation is thinner than snow, melts faster
func (c *RockDryingCalculator) estimateIceMeltTime(
	recentPrecipInches float64,
	currentWeather *models.WeatherData,
	historicalWeather []models.WeatherData,
	rockType models.RockType,
	sunExposure *models.LocationSunExposure,
) float64 {
	temp := currentWeather.Temperature

	// If at or below freezing, ice won't melt soon
	if temp <= 32 {
		return 999.0
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
