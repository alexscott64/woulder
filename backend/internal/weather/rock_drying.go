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
) models.RockDryingStatus {
	if len(rockTypes) == 0 {
		return models.RockDryingStatus{
			IsWet:         false,
			IsSafe:        true,
			Status:        "good",
			Message:       "No rock type data available",
			RockTypes:     []string{},
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

	// Find last significant rain event
	lastRainTime, totalRain := c.findLastRainEvent(historicalWeather, currentWeather)

	// If currently raining
	if currentWeather.Precipitation > 0.01 {
		status := "poor"
		message := "Currently raining - rock is wet"

		if hasWetSensitive {
			status = "critical"
			message = "DO NOT CLIMB - " + primaryRock.GroupName + " is wet-sensitive and currently raining"
		}

		return models.RockDryingStatus{
			IsWet:             true,
			IsSafe:            false,
			IsWetSensitive:    hasWetSensitive,
			HoursUntilDry:     c.estimateDryingTime(primaryRock, currentWeather, totalRain),
			LastRainTimestamp: time.Now().Format(time.RFC3339),
			Status:            status,
			Message:           message,
			RockTypes:         rockTypeNames,
			PrimaryRockType:   primaryRock.Name,
			PrimaryGroupName:  primaryRock.GroupName,
		}
	}

	// Calculate time since last rain
	if lastRainTime.IsZero() {
		// No recent rain
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
		}
	}

	hoursSinceRain := time.Since(lastRainTime).Hours()
	requiredDryingTime := c.estimateDryingTime(primaryRock, currentWeather, totalRain)

	// Check if rock has dried
	if hoursSinceRain >= requiredDryingTime {
		return models.RockDryingStatus{
			IsWet:             false,
			IsSafe:            true,
			IsWetSensitive:    hasWetSensitive,
			HoursUntilDry:     0,
			LastRainTimestamp: lastRainTime.Format(time.RFC3339),
			Status:            "good",
			Message:           "Rock is dry and safe to climb",
			RockTypes:         rockTypeNames,
			PrimaryRockType:   primaryRock.Name,
			PrimaryGroupName:  primaryRock.GroupName,
		}
	}

	// Rock is still drying
	hoursRemaining := requiredDryingTime - hoursSinceRain
	status := "fair"
	message := "Rock is drying"

	if hoursRemaining > requiredDryingTime*0.5 {
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
		HoursUntilDry:     math.Ceil(hoursRemaining),
		LastRainTimestamp: lastRainTime.Format(time.RFC3339),
		Status:            status,
		Message:           message,
		RockTypes:         rockTypeNames,
		PrimaryRockType:   primaryRock.Name,
		PrimaryGroupName:  primaryRock.GroupName,
	}
}

// findLastRainEvent finds the most recent significant rain event
func (c *RockDryingCalculator) findLastRainEvent(historical []models.WeatherData, current *models.WeatherData) (time.Time, float64) {
	var lastRainTime time.Time
	var totalRain float64

	// Check historical data (reversed to go from most recent to oldest)
	for i := len(historical) - 1; i >= 0; i-- {
		h := historical[i]
		if h.Precipitation > 0.01 {
			if lastRainTime.IsZero() {
				lastRainTime = h.Timestamp
			}
			totalRain += h.Precipitation
		} else if !lastRainTime.IsZero() {
			// Found end of rain event
			break
		}
	}

	return lastRainTime, totalRain
}

// estimateDryingTime calculates hours needed for rock to dry
func (c *RockDryingCalculator) estimateDryingTime(rockType models.RockType, weather *models.WeatherData, rainAmount float64) float64 {
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

	// Temperature modifier
	if weather.Temperature > 65 {
		dryingTime *= 0.8 // -20% for warm weather
	} else if weather.Temperature < 55 {
		dryingTime *= 1.3 // +30% for cold weather
	}

	// Cloud cover modifier (affects sun drying)
	if weather.CloudCover < 50 {
		dryingTime *= 0.75 // -25% for sunny conditions
	}

	// Wind modifier (5-15 mph is ideal)
	if weather.WindSpeed >= 5 && weather.WindSpeed <= 15 {
		dryingTime *= 0.8 // -20% for good wind
	} else if weather.WindSpeed < 3 {
		dryingTime *= 1.1 // +10% for calm conditions
	}

	// Humidity modifier
	if weather.Humidity < 50 {
		dryingTime *= 0.85 // -15% for low humidity
	} else if weather.Humidity > 70 {
		dryingTime *= 1.25 // +25% for high humidity
	}

	// Wet-sensitive rocks need extra drying time for safety
	if rockType.IsWetSensitive {
		dryingTime *= 1.5 // +50% safety margin for sandstone
	}

	return dryingTime
}
